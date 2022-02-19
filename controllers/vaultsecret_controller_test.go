//go:build integration
// +build integration

package controllers

import (
	"context"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//TODO: Example: https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/cronjob-tutorial/testdata/project/controllers/cronjob_controller_test.go
// Define utility constants for object names and testing timeouts/durations and intervals.

var vaultTestNamespace *corev1.Namespace
var vaultAdminNamespace *corev1.Namespace

const (
	vaultTestNamespaceName  = "test-vault-config-operator"
	vaultAdminNamespaceName = "vault-admin"
)

var _ = Describe("VaultSecret controller", func() {

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, vaultTestNamespace)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, vaultAdminNamespace)).Should(Succeed())
	})

	BeforeEach(func() {
		ctx := context.Background()

		// rightNow := time.Now()
		// namespace := fmt.Sprintf("vaultsecret-controller-test-%v%v%v", rightNow.Hour(), rightNow.Minute(), rightNow.Second())

		vaultAdminNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: vaultAdminNamespaceName,
			},
		}

		Expect(k8sClient.Create(ctx, vaultAdminNamespace)).Should(Succeed())

		func() {
			By("By creating a new PasswordPolicy")
			instance, err := decoder.GetPasswordPolicyInstance("../test/password-policy.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		}()

		func() {
			By("By creating new Policies")
			instance, err := decoder.GetPolicyInstance("../test/kv-engine-admin-policy.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName
			//SUBSTITUE
			instance.Spec.Policy = strings.Replace(instance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			instance, err = decoder.GetPolicyInstance("../test/secret-writer-policy.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName
			//SUBSTITUE
			instance.Spec.Policy = strings.Replace(instance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			instance, err = decoder.GetPolicyInstance("../test/vaultsecret/policy-secret-reader.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName
			//SUBSTITUE
			instance.Spec.Policy = strings.Replace(instance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		}()

		func() {
			By("By creating new KubernetesAuthEngineRoles")
			instance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/kv-engine-admin-role.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			instance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/secret-writer-role.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			instance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/vaultsecret/kubernetesauthenginerole-secret-reader.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

		}()

		vaultTestNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: vaultTestNamespaceName,
				Labels: map[string]string{
					"database-engine-admin": "true",
				},
			},
		}

		Expect(k8sClient.Create(ctx, vaultTestNamespace)).Should(Succeed())

		func() {
			By("By creating a new SecretEngineMount")
			instance, err := decoder.GetSecretEngineMountInstance("../test/kv-secret-engine.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultTestNamespaceName
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		}()

		func() {

			By("By creating new RandomSecrets")
			instance, err := decoder.GetRandomSecretInstance("../test/random-secret.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultTestNamespaceName
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			instance, err = decoder.GetRandomSecretInstance("../test/vaultsecret/randomsecret-another-password.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultTestNamespaceName
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())
		}()

	})

	timeout := time.Second * 120
	interval := time.Second * 2
	Context("When creating a VaultSecret from multiple secrets", func() {
		It("Should create a Secret when created", func() {
			By("By creating a new VaultSecret")
			ctx := context.Background()

			instance, err := decoder.GetVaultSecretInstance("../test/vaultsecret/vaultsecret-randomsecret.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultTestNamespaceName
			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.VaultSecret{}

			// We'll need to retry getting this newly created VaultSecret, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}

				for _, condition := range created.Status.Conditions {
					if condition.Type == "ReconcileSuccess" {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By checking the Secret Exists with proper Owner Reference")

			lookupKey = types.NamespacedName{Name: instance.Spec.TemplatizedK8sSecret.Name, Namespace: instance.Namespace}
			secret := &corev1.Secret{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, secret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			kind := reflect.TypeOf(redhatcopv1alpha1.VaultSecret{}).Name()
			Expect(secret.GetObjectMeta().GetOwnerReferences()[0].Kind).Should(Equal(kind))

			By("By checking the Secret Data matches expected pattern")

			var isLowerCaseLetter = regexp.MustCompile(`^[a-z]+$`).MatchString
			for k := range instance.Spec.TemplatizedK8sSecret.StringData {
				val, ok := secret.Data[k]
				Expect(ok).To(BeTrue())
				s := string(val)
				Expect(isLowerCaseLetter(s)).To(BeTrue())
				Expect(len(s)).To(Equal(20))
			}

		})
	})
})
