//go:build integration
// +build integration

package controller

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("AppRoleAuthEngineRole controller", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var mountInstance *redhatcopv1alpha1.AuthEngineMount
	var roleInstance *redhatcopv1alpha1.AppRoleAuthEngineRole

	AfterAll(func() {
		if roleInstance != nil {
			k8sIntegrationClient.Delete(ctx, roleInstance) //nolint:errcheck
		}
		if mountInstance != nil {
			k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
		}
	})

	Context("When creating an AppRole auth mount", func() {
		It("Should create the mount in Vault", func() {

			By("Loading and creating the AuthEngineMount fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/approleauthengine/test-approle-auth-mount.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			mountInstance = &redhatcopv1alpha1.AuthEngineMount{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, mountInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			created := &redhatcopv1alpha1.AuthEngineMount{}

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

			By("Verifying the auth mount exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/auth")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			_, exists := secret.Data["test-approle-auth/test-approle-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-approle-auth/test-approle-mount/' in sys/auth")
		})
	})

	Context("When creating an AppRoleAuthEngineRole", func() {
		It("Should create the role in Vault with correct settings", func() {

			By("Loading and creating the AppRoleAuthEngineRole fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/approleauthengine/test-approle-auth-role.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			roleInstance = &redhatcopv1alpha1.AppRoleAuthEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, roleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			created := &redhatcopv1alpha1.AppRoleAuthEngineRole{}

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

			By("Verifying the role in Vault")
			secret, err := vaultClient.Logical().Read("auth/test-approle-auth/test-approle-mount/role/test-approle-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			bindSecretID, ok := secret.Data["bind_secret_id"].(bool)
			Expect(ok).To(BeTrue(), "expected bind_secret_id to be a bool")
			Expect(bindSecretID).To(BeTrue())

			tokenPolicies, ok := secret.Data["token_policies"].([]any)
			Expect(ok).To(BeTrue(), "expected token_policies to be []any")
			Expect(tokenPolicies).To(ContainElement("default"))
			Expect(tokenPolicies).To(ContainElement("app-policy"))
		})
	})

	Context("When updating an AppRoleAuthEngineRole", func() {
		It("Should reflect updated values in Vault", func() {

			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before update phase")

			By("Updating the role using the updated fixture")
			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, roleInstance)).Should(Succeed())
			roleInstance.Spec.TokenPolicies = []string{"default", "updated-policy"}
			roleInstance.Spec.TokenTTL = "1h"
			roleInstance.Spec.TokenMaxTTL = "2h"
			roleInstance.Spec.SecretIDTTL = "30m"
			roleInstance.Spec.TokenPeriod = "4h"
			Expect(k8sIntegrationClient.Update(ctx, roleInstance)).Should(Succeed())

			By("Waiting for ReconcileSuccessful=True after update")
			updated := &redhatcopv1alpha1.AppRoleAuthEngineRole{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				cond := apimeta.FindStatusCondition(updated.Status.Conditions, vaultresourcecontroller.ReconcileSuccessful)
				return cond != nil && cond.Status == metav1.ConditionTrue && cond.ObservedGeneration == updated.Generation
			}, timeout, interval).Should(BeTrue())

			By("Verifying the exact updated policy list in Vault")
			secret, err := vaultClient.Logical().Read("auth/test-approle-auth/test-approle-mount/role/test-approle-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			tokenPolicies, ok := secret.Data["token_policies"].([]any)
			Expect(ok).To(BeTrue(), "expected token_policies to be []any")
			Expect(tokenPolicies).To(ConsistOf("default", "updated-policy"))

			By("Verifying the updated duration fields in Vault")
			tokenTTL, ok := secret.Data["token_ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected token_ttl to be json.Number")
			Expect(tokenTTL.String()).To(Equal("3600"), "expected token_ttl=3600 (1h)")

			tokenMaxTTL, ok := secret.Data["token_max_ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected token_max_ttl to be json.Number")
			Expect(tokenMaxTTL.String()).To(Equal("7200"), "expected token_max_ttl=7200 (2h)")

			secretIDTTL, ok := secret.Data["secret_id_ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected secret_id_ttl to be json.Number")
			Expect(secretIDTTL.String()).To(Equal("1800"), "expected secret_id_ttl=1800 (30m)")

			By("Verifying period was set correctly")
			tokenPeriod, ok := secret.Data["period"].(json.Number)
			Expect(ok).To(BeTrue(), "expected period to be json.Number")
			Expect(tokenPeriod.String()).To(Equal("14400"), "expected period=14400 (4h)")
		})
	})

	Context("When deleting AppRoleAuthEngineRole resources", func() {
		It("Should clean up the role from Vault and remove all resources", func() {

			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before delete phase")
			Expect(mountInstance).NotTo(BeNil(), "expected mount to be created before delete phase")

			By("Deleting the role CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
			roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.AppRoleAuthEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the role is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("auth/test-approle-auth/test-approle-mount/role/test-approle-role")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the AuthEngineMount")
			Expect(k8sIntegrationClient.Delete(ctx, mountInstance)).Should(Succeed())
			mountLookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, mountLookupKey, &redhatcopv1alpha1.AuthEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the auth mount is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/auth")
				if err != nil || secret == nil {
					return false
				}
				_, exists := secret.Data["test-approle-auth/test-approle-mount/"]
				return !exists
			}, timeout, interval).Should(BeTrue())
		})
	})
})
