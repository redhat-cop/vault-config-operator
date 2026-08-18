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

var _ = Describe("UserpassAuthEngineUser controller", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var mountInstance *redhatcopv1alpha1.AuthEngineMount
	var userInstance *redhatcopv1alpha1.UserpassAuthEngineUser

	AfterAll(func() {
		if userInstance != nil {
			k8sIntegrationClient.Delete(ctx, userInstance) //nolint:errcheck
		}
		if mountInstance != nil {
			k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
		}
	})

	Context("When creating a Userpass auth mount", func() {
		It("Should create the mount in Vault", func() {

			By("Loading and creating the AuthEngineMount fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/userpassauthengine/test-userpass-auth-mount.yaml", vaultAdminNamespaceName)
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
			_, exists := secret.Data["test-userpass-auth/test-userpass-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-userpass-auth/test-userpass-mount/' in sys/auth")
		})
	})

	Context("When creating a password Secret", func() {
		It("Should create the K8s Secret", func() {
			By("Loading and creating the password Secret fixture")
			_, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/userpassauthengine/test-userpass-password-secret.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
		})
	})

	Context("When creating a UserpassAuthEngineUser", func() {
		It("Should create the user in Vault with correct settings", func() {

			By("Loading and creating the UserpassAuthEngineUser fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/userpassauthengine/test-userpass-auth-user.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			userInstance = &redhatcopv1alpha1.UserpassAuthEngineUser{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, userInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: userInstance.Name, Namespace: userInstance.Namespace}
			created := &redhatcopv1alpha1.UserpassAuthEngineUser{}

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

			By("Verifying the user exists in Vault")
			secret, err := vaultClient.Logical().Read("auth/test-userpass-auth/test-userpass-mount/users/test-userpass-user")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			tokenPolicies, ok := secret.Data["token_policies"].([]any)
			Expect(ok).To(BeTrue(), "expected token_policies to be []any")
			Expect(tokenPolicies).To(ContainElement("default"))
			Expect(tokenPolicies).To(ContainElement("app-policy"))

			tokenTTL, ok := secret.Data["token_ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected token_ttl to be json.Number")
			Expect(tokenTTL.String()).To(Equal("1200"), "expected token_ttl=1200 (20m)")
		})
	})

	Context("When updating a UserpassAuthEngineUser", func() {
		It("Should reflect updated values in Vault", func() {

			Expect(userInstance).NotTo(BeNil(), "expected user to be created before update phase")

			By("Updating the user spec fields")
			lookupKey := types.NamespacedName{Name: userInstance.Name, Namespace: userInstance.Namespace}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, userInstance)).Should(Succeed())
			userInstance.Spec.TokenPolicies = []string{"default", "admin-policy"}
			userInstance.Spec.TokenTTL = "1h"
			userInstance.Spec.TokenMaxTTL = "2h"
			Expect(k8sIntegrationClient.Update(ctx, userInstance)).Should(Succeed())

			By("Waiting for ReconcileSuccessful=True after update")
			updated := &redhatcopv1alpha1.UserpassAuthEngineUser{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				cond := apimeta.FindStatusCondition(updated.Status.Conditions, vaultresourcecontroller.ReconcileSuccessful)
				return cond != nil && cond.Status == metav1.ConditionTrue && cond.ObservedGeneration == updated.Generation
			}, timeout, interval).Should(BeTrue())

			By("Verifying the updated policy list in Vault")
			secret, err := vaultClient.Logical().Read("auth/test-userpass-auth/test-userpass-mount/users/test-userpass-user")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			tokenPolicies, ok := secret.Data["token_policies"].([]any)
			Expect(ok).To(BeTrue(), "expected token_policies to be []any")
			Expect(tokenPolicies).To(ConsistOf("default", "admin-policy"))

			By("Verifying the updated TTL in Vault")
			tokenTTL, ok := secret.Data["token_ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected token_ttl to be json.Number")
			Expect(tokenTTL.String()).To(Equal("3600"), "expected token_ttl=3600 (1h)")
		})
	})

	Context("When deleting UserpassAuthEngineUser resources", func() {
		It("Should clean up the user from Vault and remove all resources", func() {

			Expect(userInstance).NotTo(BeNil(), "expected user to be created before delete phase")
			Expect(mountInstance).NotTo(BeNil(), "expected mount to be created before delete phase")

			By("Deleting the user CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, userInstance)).Should(Succeed())
			userLookupKey := types.NamespacedName{Name: userInstance.Name, Namespace: userInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, userLookupKey, &redhatcopv1alpha1.UserpassAuthEngineUser{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the user is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("auth/test-userpass-auth/test-userpass-mount/users/test-userpass-user")
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
				_, exists := secret.Data["test-userpass-auth/test-userpass-mount/"]
				return !exists
			}, timeout, interval).Should(BeTrue())
		})
	})
})
