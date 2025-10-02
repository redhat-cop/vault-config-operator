//go:build integration
// +build integration

package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	"time"

	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
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
