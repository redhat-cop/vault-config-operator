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

var _ = Describe("Identity OIDC controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var scopeInstance *redhatcopv1alpha1.IdentityOIDCScope
	var assignmentInstance *redhatcopv1alpha1.IdentityOIDCAssignment
	var clientInstance *redhatcopv1alpha1.IdentityOIDCClient
	var providerInstance *redhatcopv1alpha1.IdentityOIDCProvider

	AfterAll(func() {
		if providerInstance != nil {
			k8sIntegrationClient.Delete(ctx, providerInstance) //nolint:errcheck
		}
		if clientInstance != nil {
			k8sIntegrationClient.Delete(ctx, clientInstance) //nolint:errcheck
		}
		if assignmentInstance != nil {
			k8sIntegrationClient.Delete(ctx, assignmentInstance) //nolint:errcheck
		}
		if scopeInstance != nil {
			k8sIntegrationClient.Delete(ctx, scopeInstance) //nolint:errcheck
		}
	})

	Context("When creating an IdentityOIDCScope", func() {
		It("Should create the scope in Vault with correct settings", func() {

			By("Loading and creating the IdentityOIDCScope fixture")
			var err error
			scopeInstance, err = decoder.GetIdentityOIDCScopeInstance("../test/identityoidc/identityoidcscope.yaml")
			Expect(err).To(BeNil())
			scopeInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, scopeInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: scopeInstance.Name, Namespace: scopeInstance.Namespace}
			created := &redhatcopv1alpha1.IdentityOIDCScope{}

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

			By("Verifying the scope exists in Vault with correct settings")
			secret, err := vaultClient.Logical().Read("identity/oidc/scope/test-scope")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			description, ok := secret.Data["description"].(string)
			Expect(ok).To(BeTrue(), "expected description to be a string")
			Expect(description).To(Equal("A test scope for groups claim."))

			template, ok := secret.Data["template"].(string)
			Expect(ok).To(BeTrue(), "expected template to be a string")
			Expect(template).To(ContainSubstring("identity.entity.groups.names"))
		})
	})

	Context("When creating an IdentityOIDCAssignment", func() {
		It("Should create the assignment in Vault", func() {

			By("Loading and creating the IdentityOIDCAssignment fixture")
			var err error
			assignmentInstance, err = decoder.GetIdentityOIDCAssignmentInstance("../test/identityoidc/identityoidcassignment.yaml")
			Expect(err).To(BeNil())
			assignmentInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, assignmentInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: assignmentInstance.Name, Namespace: assignmentInstance.Namespace}
			created := &redhatcopv1alpha1.IdentityOIDCAssignment{}

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

			By("Verifying the assignment exists in Vault")
			secret, err := vaultClient.Logical().Read("identity/oidc/assignment/test-assignment")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
		})
	})

	Context("When creating an IdentityOIDCClient", func() {
		It("Should create the client in Vault with correct settings", func() {

			By("Loading and creating the IdentityOIDCClient fixture")
			var err error
			clientInstance, err = decoder.GetIdentityOIDCClientInstance("../test/identityoidc/identityoidcclient.yaml")
			Expect(err).To(BeNil())
			clientInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, clientInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: clientInstance.Name, Namespace: clientInstance.Namespace}
			created := &redhatcopv1alpha1.IdentityOIDCClient{}

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

			By("Verifying the client exists in Vault with correct settings")
			secret, err := vaultClient.Logical().Read("identity/oidc/client/test-client")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			clientType, ok := secret.Data["client_type"].(string)
			Expect(ok).To(BeTrue(), "expected client_type to be a string")
			Expect(clientType).To(Equal("confidential"))

			key, ok := secret.Data["key"].(string)
			Expect(ok).To(BeTrue(), "expected key to be a string")
			Expect(key).To(Equal("default"))

			clientID, ok := secret.Data["client_id"].(string)
			Expect(ok).To(BeTrue(), "expected client_id to be a string")
			Expect(clientID).NotTo(BeEmpty())

			redirectURIs, ok := secret.Data["redirect_uris"].([]interface{})
			Expect(ok).To(BeTrue(), "expected redirect_uris to be []interface{}")
			Expect(redirectURIs).To(ContainElement("https://example.com/callback"))

			assignments, ok := secret.Data["assignments"].([]interface{})
			Expect(ok).To(BeTrue(), "expected assignments to be []interface{}")
			Expect(assignments).To(ContainElement("test-assignment"))
		})
	})

	Context("When creating an IdentityOIDCProvider", func() {
		It("Should create the provider in Vault with correct settings", func() {

			By("Loading and creating the IdentityOIDCProvider fixture")
			var err error
			providerInstance, err = decoder.GetIdentityOIDCProviderInstance("../test/identityoidc/identityoidcprovider.yaml")
			Expect(err).To(BeNil())
			providerInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, providerInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: providerInstance.Name, Namespace: providerInstance.Namespace}
			created := &redhatcopv1alpha1.IdentityOIDCProvider{}

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

			By("Verifying the provider exists in Vault with correct settings")
			secret, err := vaultClient.Logical().Read("identity/oidc/provider/test-provider")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			scopesSupported, ok := secret.Data["scopes_supported"].([]interface{})
			Expect(ok).To(BeTrue(), "expected scopes_supported to be []interface{}")
			Expect(scopesSupported).To(ContainElement("test-scope"))

			allowedClientIDs, ok := secret.Data["allowed_client_ids"].([]interface{})
			Expect(ok).To(BeTrue(), "expected allowed_client_ids to be []interface{}")
			Expect(allowedClientIDs).To(ContainElement("*"))
		})
	})

	Context("When updating an IdentityOIDCScope", func() {
		It("Should update the scope in Vault and reflect updated ObservedGeneration", func() {

			Expect(scopeInstance).NotTo(BeNil(), "expected scope to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: scopeInstance.Name, Namespace: scopeInstance.Namespace}
			current := &redhatcopv1alpha1.IdentityOIDCScope{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating the scope description")
			current.Spec.Description = "Updated scope description."
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated description")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/oidc/scope/test-scope")
				if err != nil || secret == nil {
					return false
				}
				description, ok := secret.Data["description"].(string)
				if !ok {
					return false
				}
				return description == "Updated scope description."
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.IdentityOIDCScope{}
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

	Context("When deleting all OIDC resources", func() {
		It("Should clean up all resources from Vault in reverse dependency order", func() {

			Expect(providerInstance).NotTo(BeNil(), "expected provider to be created before delete phase")
			Expect(clientInstance).NotTo(BeNil(), "expected client to be created before delete phase")
			Expect(assignmentInstance).NotTo(BeNil(), "expected assignment to be created before delete phase")
			Expect(scopeInstance).NotTo(BeNil(), "expected scope to be created before delete phase")

			By("Deleting the IdentityOIDCProvider CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, providerInstance)).Should(Succeed())
			providerLookupKey := types.NamespacedName{Name: providerInstance.Name, Namespace: providerInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, providerLookupKey, &redhatcopv1alpha1.IdentityOIDCProvider{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the provider is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/oidc/provider/test-provider")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the IdentityOIDCClient CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, clientInstance)).Should(Succeed())
			clientLookupKey := types.NamespacedName{Name: clientInstance.Name, Namespace: clientInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, clientLookupKey, &redhatcopv1alpha1.IdentityOIDCClient{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the client is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/oidc/client/test-client")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the IdentityOIDCAssignment CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, assignmentInstance)).Should(Succeed())
			assignmentLookupKey := types.NamespacedName{Name: assignmentInstance.Name, Namespace: assignmentInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, assignmentLookupKey, &redhatcopv1alpha1.IdentityOIDCAssignment{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the assignment is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/oidc/assignment/test-assignment")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the IdentityOIDCScope CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, scopeInstance)).Should(Succeed())
			scopeLookupKey := types.NamespacedName{Name: scopeInstance.Name, Namespace: scopeInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, scopeLookupKey, &redhatcopv1alpha1.IdentityOIDCScope{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the scope is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/oidc/scope/test-scope")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
