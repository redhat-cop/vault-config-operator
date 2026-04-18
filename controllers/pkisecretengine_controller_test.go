//go:build integration
// +build integration

package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("PKISecretEngineConfig controller", func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	Context("When preparing a PKI Secren Engine", func() {
		It("Should create a PKI Secret Engine when created", func() {
			By("By creating new Policies")
			pInstance, err := decoder.GetPolicyInstance("../test/pkisecretengine/pki-secret-engine-admin-policy.yaml")
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

			kaerInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/pkisecretengine/pki-secret-engine-kube-auth-role.yaml")
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

			By("By creating a new SecretEngineMount")

			semInstance, err := decoder.GetSecretEngineMountInstance("../test/pkisecretengine/pki-secret-engine.yaml")
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
		})
	})

	Context("When creating a PKISecretEngineConfig", func() {
		It("Should configure the PKI for the specific pki path when created", func() {

			rsInstance, err := decoder.GetPKISecretEngineConfigInstance("../test/pkisecretengine/pki-secret-engine-config.yaml")
			Expect(err).To(BeNil())
			rsInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstance)).Should(Succeed())

			rslookupKey := types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated := &redhatcopv1alpha1.PKISecretEngineConfig{}

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
		})
	})

	Context("When creating a PKISecretEngineRole", func() {
		It("Should configure the PKI role for the specific pki path when created", func() {

			rsInstance, err := decoder.GetPKISecretEngineRoleInstance("../test/pkisecretengine/pki-secret-engine-role.yaml")
			Expect(err).To(BeNil())
			rsInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstance)).Should(Succeed())

			rslookupKey := types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated := &redhatcopv1alpha1.PKISecretEngineRole{}

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
		})
	})

	Context("When updating a PKISecretEngineRole", func() {
		It("Should update the role in Vault and reflect updated ObservedGeneration", func() {

			By("Verifying initial Vault state for PKISecretEngineRole")
			initialSecret, err := vaultClient.Logical().Read("test-vault-config-operator/pki/roles/pki-example")
			Expect(err).To(BeNil())
			Expect(initialSecret).NotTo(BeNil())
			initialDomains, ok := initialSecret.Data["allowed_domains"].([]interface{})
			Expect(ok).To(BeTrue(), "expected allowed_domains to be []interface{}")
			Expect(initialDomains).To(ContainElement("internal.io"))
			Expect(initialDomains).To(ContainElement("pki-vault-demo.svc"))
			Expect(initialDomains).To(ContainElement("example.com"))

			By("Recording initial ObservedGeneration")
			lookupKey := types.NamespacedName{Name: "pki-example", Namespace: vaultTestNamespaceName}
			created := &redhatcopv1alpha1.PKISecretEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())
			var initialObservedGeneration int64
			for _, condition := range created.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialObservedGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialObservedGeneration).To(BeNumerically(">", 0))

			By("Getting the latest PKISecretEngineRole before update")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())

			By("Updating the PKISecretEngineRole spec")
			created.Spec.AllowedDomains = append(created.Spec.AllowedDomains, "test.io")
			created.Spec.MaxTTL = metav1.Duration{Duration: 4380 * time.Hour}
			Expect(k8sIntegrationClient.Update(ctx, created)).Should(Succeed())

			By("Waiting for Vault to reflect the updated allowed_domains")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/pki/roles/pki-example")
				if err != nil || secret == nil {
					return false
				}
				domains, ok := secret.Data["allowed_domains"].([]interface{})
				if !ok {
					return false
				}
				for _, d := range domains {
					if d == "test.io" {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Verifying max_ttl reflects the updated value")
			updatedSecret, err := vaultClient.Logical().Read("test-vault-config-operator/pki/roles/pki-example")
			Expect(err).To(BeNil())
			Expect(updatedSecret).NotTo(BeNil())
			maxTTLRaw := updatedSecret.Data["max_ttl"]
			maxTTLJson, ok := maxTTLRaw.(json.Number)
			if ok {
				maxTTLInt, err := maxTTLJson.Int64()
				Expect(err).To(BeNil())
				Expect(maxTTLInt).To(Equal(int64(4380 * 3600)))
			} else {
				Expect(fmt.Sprintf("%v", maxTTLRaw)).To(ContainSubstring("4380"))
			}

			By("Verifying ObservedGeneration increased")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}
				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return condition.ObservedGeneration > initialObservedGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When updating a PKISecretEngineConfig CRL config", func() {
		It("Should update the CRL config in Vault and reflect updated ObservedGeneration", func() {

			By("Reading initial CRL config from Vault")
			initialCRL, err := vaultClient.Logical().Read("test-vault-config-operator/pki/config/crl")
			Expect(err).To(BeNil())
			Expect(initialCRL).NotTo(BeNil())
			initialDisable := initialCRL.Data["disable"]

			By("Recording initial ObservedGeneration for PKISecretEngineConfig")
			configLookupKey := types.NamespacedName{Name: "pki", Namespace: vaultTestNamespaceName}
			configInstance := &redhatcopv1alpha1.PKISecretEngineConfig{}
			Expect(k8sIntegrationClient.Get(ctx, configLookupKey, configInstance)).Should(Succeed())
			var initialObservedGeneration int64
			for _, condition := range configInstance.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialObservedGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialObservedGeneration).To(BeNumerically(">", 0))

			By("Getting the latest PKISecretEngineConfig before update")
			Expect(k8sIntegrationClient.Get(ctx, configLookupKey, configInstance)).Should(Succeed())

			By("Updating the PKISecretEngineConfig CRL fields")
			configInstance.Spec.CRLExpiry.Duration = 48 * time.Hour
			configInstance.Spec.CRLDisable = true
			Expect(k8sIntegrationClient.Update(ctx, configInstance)).Should(Succeed())

			By("Waiting for Vault CRL config to reflect the disable=true change")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/pki/config/crl")
				if err != nil || secret == nil {
					return false
				}
				disable, ok := secret.Data["disable"].(bool)
				if !ok {
					return false
				}
				return disable
			}, timeout, interval).Should(BeTrue())

			By("Verifying disable changed from initial value")
			Expect(initialDisable).To(Equal(false))

			By("Verifying CRL expiry reflects the updated value")
			updatedCRL, err := vaultClient.Logical().Read("test-vault-config-operator/pki/config/crl")
			Expect(err).To(BeNil())
			Expect(updatedCRL).NotTo(BeNil())
			Expect(fmt.Sprintf("%v", updatedCRL.Data["expiry"])).To(ContainSubstring("48h"))

			By("Verifying ObservedGeneration increased")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, configInstance)
				if err != nil {
					return false
				}
				for _, condition := range configInstance.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return condition.ObservedGeneration > initialObservedGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting a PKISecretEngineRole", func() {
		It("It should be deleted from Vault", func() {

			By("Deleting PKISecretEngineRoleInstance(")

			pkiRoleInstance, err := decoder.GetPKISecretEngineRoleInstance("../test/pkisecretengine/pki-secret-engine-role.yaml")
			Expect(err).To(BeNil())
			pkiRoleInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, pkiRoleInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(pkiRoleInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting PKISecretEngineConfigInstance")
			pkiConfigInstance, err := decoder.GetPKISecretEngineConfigInstance("../test/pkisecretengine/pki-secret-engine-config.yaml")
			Expect(err).To(BeNil())
			pkiConfigInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, pkiConfigInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(pkiConfigInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting SecretEngineMount")

			semInstance, err := decoder.GetSecretEngineMountInstance("../test/pkisecretengine/pki-secret-engine.yaml")
			Expect(err).To(BeNil())
			semInstance.Namespace = vaultTestNamespaceName

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

			By("Deleting KubernetesAuthEngineRoleInstance")

			kaerInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/pkisecretengine/pki-secret-engine-kube-auth-role.yaml")
			Expect(err).To(BeNil())
			kaerInstance.Namespace = vaultAdminNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, kaerInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(kaerInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting Policy")
			pInstance, err := decoder.GetPolicyInstance("../test/pkisecretengine/pki-secret-engine-admin-policy.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, pInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(pInstance.GetPath())
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
