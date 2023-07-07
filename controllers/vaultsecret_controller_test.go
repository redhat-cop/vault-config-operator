//go:build integration
// +build integration

package controllers

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("VaultSecret controller", func() {

	timeout := time.Second * 120
	interval := time.Second * 2
	Context("When creating a VaultSecret from multiple secrets", func() {
		It("Should create a Secret when created", func() {

			By("By creating a new PasswordPolicy")
			ppInstance, err := decoder.GetPasswordPolicyInstance("../test/password-policy.yaml")
			Expect(err).To(BeNil())
			ppInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, ppInstance)).Should(Succeed())

			pplookupKey := types.NamespacedName{Name: ppInstance.Name, Namespace: ppInstance.Namespace}
			ppCreated := &redhatcopv1alpha1.PasswordPolicy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pplookupKey, ppCreated)
				if err != nil {
					return false
				}

				for _, condition := range ppCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating new Policies")
			pInstance, err := decoder.GetPolicyInstance("../test/kv-engine-admin-policy.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, pInstance)).Should(Succeed())

			pLookupKey := types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}
			pCreated := &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pLookupKey, pCreated)
				if err != nil {
					return false
				}

				for _, condition := range pCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			pInstance, err = decoder.GetPolicyInstance("../test/secret-writer-policy.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, pInstance)).Should(Succeed())

			pLookupKey = types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}
			pCreated = &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pLookupKey, pCreated)
				if err != nil {
					return false
				}

				for _, condition := range pCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			pInstance, err = decoder.GetPolicyInstance("../test/vaultsecret/policy-secret-reader.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, pInstance)).Should(Succeed())

			pLookupKey = types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}
			pCreated = &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pLookupKey, pCreated)
				if err != nil {
					return false
				}

				for _, condition := range pCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating new KubernetesAuthEngineRoles")

			kaerInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/kv-engine-admin-role.yaml")
			Expect(err).To(BeNil())
			kaerInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerInstance)).Should(Succeed())

			kaerLookupKey := types.NamespacedName{Name: kaerInstance.Name, Namespace: kaerInstance.Namespace}
			kaerCreated := &redhatcopv1alpha1.KubernetesAuthEngineRole{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, kaerLookupKey, kaerCreated)
				if err != nil {
					return false
				}

				for _, condition := range kaerCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			kaerInstance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/secret-writer-role.yaml")
			Expect(err).To(BeNil())
			kaerInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerInstance)).Should(Succeed())

			kaerLookupKey = types.NamespacedName{Name: kaerInstance.Name, Namespace: kaerInstance.Namespace}
			kaerCreated = &redhatcopv1alpha1.KubernetesAuthEngineRole{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, kaerLookupKey, kaerCreated)
				if err != nil {
					return false
				}

				for _, condition := range kaerCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			kaerInstance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/vaultsecret/kubernetesauthenginerole-secret-reader.yaml")
			Expect(err).To(BeNil())
			kaerInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerInstance)).Should(Succeed())

			kaerLookupKey = types.NamespacedName{Name: kaerInstance.Name, Namespace: kaerInstance.Namespace}
			kaerCreated = &redhatcopv1alpha1.KubernetesAuthEngineRole{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, kaerLookupKey, kaerCreated)
				if err != nil {
					return false
				}

				for _, condition := range kaerCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new SecretEngineMount")

			semInstance, err := decoder.GetSecretEngineMountInstance("../test/kv-secret-engine.yaml")
			Expect(err).To(BeNil())
			semInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, semInstance)).Should(Succeed())

			semLookupKey := types.NamespacedName{Name: semInstance.Name, Namespace: semInstance.Namespace}
			semCreated := &redhatcopv1alpha1.SecretEngineMount{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, semLookupKey, semCreated)
				if err != nil {
					return false
				}

				for _, condition := range semCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating new RandomSecrets")

			rsInstance, err := decoder.GetRandomSecretInstance("../test/random-secret.yaml")
			Expect(err).To(BeNil())
			rsInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstance)).Should(Succeed())

			rslookupKey := types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated := &redhatcopv1alpha1.RandomSecret{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rslookupKey, rsCreated)
				if err != nil {
					return false
				}

				for _, condition := range rsCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			rsInstanceAnotherPassword, err := decoder.GetRandomSecretInstance("../test/vaultsecret/randomsecret-another-password.yaml")
			Expect(err).To(BeNil())
			rsInstanceAnotherPassword.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstanceAnotherPassword)).Should(Succeed())

			rslookupKey = types.NamespacedName{Name: rsInstanceAnotherPassword.Name, Namespace: rsInstanceAnotherPassword.Namespace}
			rsCreated = &redhatcopv1alpha1.RandomSecret{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rslookupKey, rsCreated)
				if err != nil {
					return false
				}

				for _, condition := range rsCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new VaultSecret")

			instance, err := decoder.GetVaultSecretInstance("../test/vaultsecret/vaultsecret-randomsecret.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.VaultSecret{}

			// We'll need to retry getting this newly created VaultSecret, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}

				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By checking the Secret Exists with proper Owner Reference")

			lookupKey = types.NamespacedName{Name: instance.Spec.TemplatizedK8sSecret.Name, Namespace: instance.Namespace}
			secret := &corev1.Secret{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, secret)
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

			By("Deleting the VaultSecret")

			Expect(k8sIntegrationClient.Delete(ctx, instance)).Should(Succeed())

			By("Checking the K8s Secret was deleted")

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, &corev1.Secret{})
				if err != nil {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Deleting RandomSecrets")

			Expect(k8sIntegrationClient.Delete(ctx, rsInstance)).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, rsInstanceAnotherPassword)).Should(Succeed())

			By("Checking the KVv1 secrets in vault were deleted")

			Eventually(func() bool {
				_, err := vaultClient.KVv1(string(rsInstance.Spec.Path)).Get(ctx, rsInstance.Name)
				if err != nil {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				_, err := vaultClient.KVv1(string(rsInstanceAnotherPassword.Spec.Path)).Get(ctx, rsInstanceAnotherPassword.Name)
				if err != nil {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())

			Eventually(func() error {
				secret, _ := vaultClient.Logical().Read(string(rsInstanceAnotherPassword.Spec.Path))
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("By Deleting the SecretEngineMount")

			Expect(k8sIntegrationClient.Delete(ctx, semInstance)).Should(Succeed())

			By("Checking the secret engine mount path in vault is deleted")

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(semInstance.GetPath())
				fmt.Fprintf(GinkgoWriter, "Some log text: %v\n", secret)
				if err != nil {
					return nil
				}
				return fmt.Errorf("secret is not nil %s", secret.Data)
			}, timeout, interval).Should(Succeed())

		})
	})
})
