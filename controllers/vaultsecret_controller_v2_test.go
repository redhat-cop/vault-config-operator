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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("VaultSecret controller for v2 secrets", func() {

	timeout := time.Second * 120
	interval := time.Second * 2
	Context("When creating a VaultSecret from multiple secrets", func() {
		It("Should create a Secret when created", func() {

			By("By creating a new PasswordPolicy")
			ppInstance, err := decoder.GetPasswordPolicyInstance("../test/randomsecret/v2/00-passwordpolicy-simple-password-policy-v2.yaml")
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
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating new Policies")
			pInstance, err := decoder.GetPolicyInstance("../test/randomsecret/v2/01-policy-kv-engine-admin-v2.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName
			//SUBSTITUE
			pInstance.Spec.Policy = strings.Replace(pInstance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)
			Expect(k8sIntegrationClient.Create(ctx, pInstance)).Should(Succeed())

			pLookupKey := types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}
			pCreated := &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pLookupKey, pCreated)
				if err != nil {
					return false
				}

				for _, condition := range pCreated.Status.Conditions {
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			pInstance, err = decoder.GetPolicyInstance("../test/randomsecret/v2/04-policy-secret-writer-v2.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName
			//SUBSTITUE
			pInstance.Spec.Policy = strings.Replace(pInstance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)
			Expect(k8sIntegrationClient.Create(ctx, pInstance)).Should(Succeed())

			pLookupKey = types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}
			pCreated = &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pLookupKey, pCreated)
				if err != nil {
					return false
				}

				for _, condition := range pCreated.Status.Conditions {
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			pInstance, err = decoder.GetPolicyInstance("../test/vaultsecret/v2/00-policy-secret-reader-v2.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName
			//SUBSTITUE
			pInstance.Spec.Policy = strings.Replace(pInstance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)
			Expect(k8sIntegrationClient.Create(ctx, pInstance)).Should(Succeed())

			pLookupKey = types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}
			pCreated = &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pLookupKey, pCreated)
				if err != nil {
					return false
				}

				for _, condition := range pCreated.Status.Conditions {
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating new KubernetesAuthEngineRoles")

			kaerInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml")
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
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			kaerInstance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml")
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
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			kaerInstance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml")
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
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new SecretEngineMount")

			semInstance, err := decoder.GetSecretEngineMountInstance("../test/randomsecret/v2/03-secretenginemount-kv-v2.yaml")
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
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating new RandomSecrets")

			rsInstance, err := decoder.GetRandomSecretInstance("../test/randomsecret/v2/06-randomsecret-randomsecret-password-v2.yaml")
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
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			rsInstance, err = decoder.GetRandomSecretInstance("../test/randomsecret/v2/07-randomsecret-randomsecret-another-password-v2.yaml")
			Expect(err).To(BeNil())
			rsInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstance)).Should(Succeed())

			rslookupKey = types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated = &redhatcopv1alpha1.RandomSecret{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rslookupKey, rsCreated)
				if err != nil {
					return false
				}

				for _, condition := range rsCreated.Status.Conditions {
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new VaultSecret")

			ctx := context.Background()

			instance, err := decoder.GetVaultSecretInstance("../test/vaultsecret/v2/07-vaultsecret-randomsecret-v2.yaml")
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
					if condition.Type == "ReconcileSuccessful" && condition.Status == metav1.ConditionTrue {
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
		})
	})
})
