//go:build integration
// +build integration

package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Identity Token controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var configInstance *redhatcopv1alpha1.IdentityTokenConfig
	var keyInstance *redhatcopv1alpha1.IdentityTokenKey
	var roleInstance *redhatcopv1alpha1.IdentityTokenRole

	AfterAll(func() {
		if roleInstance != nil {
			k8sIntegrationClient.Delete(ctx, roleInstance) //nolint:errcheck
		}
		if keyInstance != nil {
			k8sIntegrationClient.Delete(ctx, keyInstance) //nolint:errcheck
		}
		if configInstance != nil {
			k8sIntegrationClient.Delete(ctx, configInstance) //nolint:errcheck
		}
	})

	Context("When creating an IdentityTokenConfig", func() {
		It("Should configure the OIDC issuer in Vault", func() {

			By("Loading and creating the IdentityTokenConfig fixture")
			var err error
			configInstance, err = decoder.GetIdentityTokenConfigInstance("../test/identitytoken/identitytokenconfig.yaml")
			Expect(err).To(BeNil())
			configInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.IdentityTokenConfig{}

			By("Waiting for ReconcileSuccessful=True")
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

			By("Verifying the config exists in Vault")
			secret, err := vaultClient.Logical().Read("identity/oidc/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data).NotTo(BeNil())
		})
	})

	Context("When creating an IdentityTokenKey", func() {
		It("Should create the key in Vault with correct settings", func() {

			By("Loading and creating the IdentityTokenKey fixture")
			var err error
			keyInstance, err = decoder.GetIdentityTokenKeyInstance("../test/identitytoken/identitytokenkey.yaml")
			Expect(err).To(BeNil())
			keyInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, keyInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: keyInstance.Name, Namespace: keyInstance.Namespace}
			created := &redhatcopv1alpha1.IdentityTokenKey{}

			By("Waiting for ReconcileSuccessful=True")
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

			By("Verifying the key exists in Vault with correct settings")
			secret, err := vaultClient.Logical().Read("identity/oidc/key/test-key")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			algorithm, ok := secret.Data["algorithm"].(string)
			Expect(ok).To(BeTrue(), "expected algorithm to be a string")
			Expect(algorithm).To(Equal("RS256"))

			allowedClientIDs, ok := secret.Data["allowed_client_ids"].([]interface{})
			Expect(ok).To(BeTrue(), "expected allowed_client_ids to be []interface{}")
			Expect(allowedClientIDs).To(ContainElement("*"))
		})
	})

	Context("When creating an IdentityTokenRole", func() {
		It("Should create the role in Vault referencing the key", func() {

			By("Loading and creating the IdentityTokenRole fixture")
			var err error
			roleInstance, err = decoder.GetIdentityTokenRoleInstance("../test/identitytoken/identitytokenrole.yaml")
			Expect(err).To(BeNil())
			roleInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, roleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			created := &redhatcopv1alpha1.IdentityTokenRole{}

			By("Waiting for ReconcileSuccessful=True")
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

			By("Verifying the role exists in Vault with correct key reference")
			secret, err := vaultClient.Logical().Read("identity/oidc/role/test-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			key, ok := secret.Data["key"].(string)
			Expect(ok).To(BeTrue(), "expected key to be a string")
			Expect(key).To(Equal("test-key"))
		})
	})

	Context("When updating an IdentityTokenKey", func() {
		It("Should update the key in Vault and reflect updated ObservedGeneration", func() {

			Expect(keyInstance).NotTo(BeNil(), "expected key to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: keyInstance.Name, Namespace: keyInstance.Namespace}
			current := &redhatcopv1alpha1.IdentityTokenKey{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating the key algorithm to ES256")
			current.Spec.Algorithm = "ES256"
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated algorithm")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/oidc/key/test-key")
				if err != nil || secret == nil {
					return false
				}
				algorithm, ok := secret.Data["algorithm"].(string)
				if !ok {
					return false
				}
				return algorithm == "ES256"
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.IdentityTokenKey{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				for _, condition := range updated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return condition.ObservedGeneration > initialGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting Identity Token resources", func() {
		It("Should clean up deletable resources from Vault in reverse dependency order", func() {

			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before delete phase")
			Expect(keyInstance).NotTo(BeNil(), "expected key to be created before delete phase")
			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")

			By("Deleting the IdentityTokenRole CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
			roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.IdentityTokenRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the role is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/oidc/role/test-role")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the IdentityTokenKey CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, keyInstance)).Should(Succeed())
			keyLookupKey := types.NamespacedName{Name: keyInstance.Name, Namespace: keyInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, keyLookupKey, &redhatcopv1alpha1.IdentityTokenKey{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the key is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/oidc/key/test-key")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the IdentityTokenConfig CR (IsDeletable=false)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.IdentityTokenConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the Vault config STILL exists (IsDeletable=false means no Vault cleanup)")
			secret, err := vaultClient.Logical().Read("identity/oidc/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data).NotTo(BeNil())
		})
	})
})
