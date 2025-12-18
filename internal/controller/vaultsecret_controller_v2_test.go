//go:build integration
// +build integration

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	"reflect"
	"regexp"
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
		It("Should create a Secret when created, and be removed from Vault when deleted", func() {

			By("Creating a new PasswordPolicy")
			passwordPolicySimplePasswordInstance, err := decoder.GetPasswordPolicyInstance("../test/randomsecret/v2/00-passwordpolicy-simple-password-policy-v2.yaml")
			Expect(err).To(BeNil())
			passwordPolicySimplePasswordInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, passwordPolicySimplePasswordInstance)).Should(Succeed())

			pplookupKey := types.NamespacedName{Name: passwordPolicySimplePasswordInstance.Name, Namespace: passwordPolicySimplePasswordInstance.Namespace}
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

			By("Creating new Policies")
			policyKVEngineAdminInstance, err := decoder.GetPolicyInstance("../test/randomsecret/v2/01-policy-kv-engine-admin-v2.yaml")
			Expect(err).To(BeNil())
			policyKVEngineAdminInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, policyKVEngineAdminInstance)).Should(Succeed())

			pLookupKey := types.NamespacedName{Name: policyKVEngineAdminInstance.Name, Namespace: policyKVEngineAdminInstance.Namespace}
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

			policySecretWriterInstance, err := decoder.GetPolicyInstance("../test/randomsecret/v2/04-policy-secret-writer-v2.yaml")
			Expect(err).To(BeNil())
			policySecretWriterInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, policySecretWriterInstance)).Should(Succeed())

			pLookupKey = types.NamespacedName{Name: policySecretWriterInstance.Name, Namespace: policySecretWriterInstance.Namespace}
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

			policySecretReaderInstance, err := decoder.GetPolicyInstance("../test/vaultsecret/v2/00-policy-secret-reader-v2.yaml")
			Expect(err).To(BeNil())
			policySecretReaderInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, policySecretReaderInstance)).Should(Succeed())

			pLookupKey = types.NamespacedName{Name: policySecretReaderInstance.Name, Namespace: policySecretReaderInstance.Namespace}
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

			By("Creating new KubernetesAuthEngineRoles")

			kaerKVEngineAdminInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml")
			Expect(err).To(BeNil())
			kaerKVEngineAdminInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerKVEngineAdminInstance)).Should(Succeed())

			kaerLookupKey := types.NamespacedName{Name: kaerKVEngineAdminInstance.Name, Namespace: kaerKVEngineAdminInstance.Namespace}
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

			kaerSecretWriterInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml")
			Expect(err).To(BeNil())
			kaerSecretWriterInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerSecretWriterInstance)).Should(Succeed())

			kaerLookupKey = types.NamespacedName{Name: kaerSecretWriterInstance.Name, Namespace: kaerSecretWriterInstance.Namespace}
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

			kaerSecretReaderInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml")
			Expect(err).To(BeNil())
			kaerSecretReaderInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerSecretReaderInstance)).Should(Succeed())

			kaerLookupKey = types.NamespacedName{Name: kaerSecretReaderInstance.Name, Namespace: kaerSecretReaderInstance.Namespace}
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

			By("Creating a new SecretEngineMount")

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
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("Creating new RandomSecrets")

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
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			rsInstanceAnotherPassword, err := decoder.GetRandomSecretInstance("../test/randomsecret/v2/07-randomsecret-randomsecret-another-password-v2.yaml")
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

			By("Creating a new VaultSecret")

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
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("Checking the Secret Exists with proper Owner Reference")

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

			By("Checking the Secret Data matches expected pattern")

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

			Eventually(func() error {
				secret, _ := vaultClient.Logical().Read(rsInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, rsInstanceAnotherPassword)).Should(Succeed())

			Eventually(func() error {
				secret, _ := vaultClient.Logical().Read(rsInstanceAnotherPassword.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting the SecretEngineMount")

			Expect(k8sIntegrationClient.Delete(ctx, semInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(semInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting KubernetesAuthEngineRoles")

			Expect(k8sIntegrationClient.Delete(ctx, kaerSecretReaderInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(kaerSecretReaderInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, kaerSecretWriterInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(kaerSecretWriterInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, kaerKVEngineAdminInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(kaerKVEngineAdminInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting Policies")

			Expect(k8sIntegrationClient.Delete(ctx, policySecretReaderInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(policySecretReaderInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, policySecretWriterInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(policySecretWriterInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, policyKVEngineAdminInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(policyKVEngineAdminInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting PasswordPolicy")

			Expect(k8sIntegrationClient.Delete(ctx, passwordPolicySimplePasswordInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(passwordPolicySimplePasswordInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

		})
	})
})
