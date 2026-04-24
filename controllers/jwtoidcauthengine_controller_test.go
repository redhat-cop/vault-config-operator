//go:build integration
// +build integration

package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("JWTOIDCAuthEngine controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var oidcSecret *corev1.Secret
	var mountInstance *redhatcopv1alpha1.AuthEngineMount
	var configInstance *redhatcopv1alpha1.JWTOIDCAuthEngineConfig
	var roleInstance *redhatcopv1alpha1.JWTOIDCAuthEngineRole

	AfterAll(func() {
		if roleInstance != nil {
			k8sIntegrationClient.Delete(ctx, roleInstance) //nolint:errcheck
		}
		if configInstance != nil {
			k8sIntegrationClient.Delete(ctx, configInstance) //nolint:errcheck
		}
		if mountInstance != nil {
			k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
		}
		if oidcSecret != nil {
			k8sIntegrationClient.Delete(ctx, oidcSecret) //nolint:errcheck
		}
	})

	Context("When creating prerequisite resources", func() {
		It("Should create the OIDC credentials secret and JWT/OIDC auth mount", func() {

			By("Creating the OIDC credentials K8s Secret")
			oidcSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-oidc-creds",
					Namespace: vaultAdminNamespaceName,
				},
				Data: map[string][]byte{
					"oidc_client_id":     []byte("vault-oidc"),
					"oidc_client_secret": []byte("test-client-secret"),
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, oidcSecret)).Should(Succeed())

			By("Loading and creating the AuthEngineMount fixture")
			var err error
			mountInstance, err = decoder.GetAuthEngineMountInstance("../test/jwtoidcauthengine/test-jwtoidc-auth-mount.yaml")
			Expect(err).To(BeNil())
			mountInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, mountInstance)).Should(Succeed())

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
			_, exists := secret.Data["test-jwt-oidc-auth/test-joaec-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-jwt-oidc-auth/test-joaec-mount/' in sys/auth")
		})
	})

	Context("When creating a JWTOIDCAuthEngineConfig", func() {
		It("Should write the OIDC config to Vault", func() {

			By("Loading and creating the JWTOIDCAuthEngineConfig fixture")
			var err error
			configInstance, err = decoder.GetJWTOIDCAuthEngineConfigInstance("../test/jwtoidcauthengine/test-jwtoidc-auth-config.yaml")
			Expect(err).To(BeNil())
			configInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.JWTOIDCAuthEngineConfig{}

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

			By("Verifying the config in Vault")
			secret, err := vaultClient.Logical().Read("auth/test-jwt-oidc-auth/test-joaec-mount/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["oidc_discovery_url"]).To(Equal("http://keycloak.keycloak.svc.cluster.local:8080/realms/test-realm"))
			Expect(secret.Data["oidc_client_id"]).To(Equal("vault-oidc"))
		})
	})

	Context("When creating a JWTOIDCAuthEngineRole", func() {
		It("Should create the role in Vault with correct OIDC settings", func() {

			By("Loading and creating the JWTOIDCAuthEngineRole fixture")
			var err error
			roleInstance, err = decoder.GetJWTOIDCAuthEngineRoleInstance("../test/jwtoidcauthengine/test-jwtoidc-auth-role.yaml")
			Expect(err).To(BeNil())
			roleInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, roleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			created := &redhatcopv1alpha1.JWTOIDCAuthEngineRole{}

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
			secret, err := vaultClient.Logical().Read("auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			roleType, ok := secret.Data["role_type"].(string)
			Expect(ok).To(BeTrue(), "expected role_type to be a string")
			Expect(roleType).To(Equal("oidc"))

			userClaim, ok := secret.Data["user_claim"].(string)
			Expect(ok).To(BeTrue(), "expected user_claim to be a string")
			Expect(userClaim).To(Equal("email"))

			tokenPolicies, ok := secret.Data["token_policies"].([]interface{})
			Expect(ok).To(BeTrue(), "expected token_policies to be []interface{}")
			Expect(tokenPolicies).To(ContainElement("default"))

			oidcScopes, ok := secret.Data["oidc_scopes"].([]interface{})
			Expect(ok).To(BeTrue(), "expected oidc_scopes to be []interface{}")
			Expect(oidcScopes).To(ContainElement("openid"))
			Expect(oidcScopes).To(ContainElement("email"))

			allowedRedirects, ok := secret.Data["allowed_redirect_uris"].([]interface{})
			Expect(ok).To(BeTrue(), "expected allowed_redirect_uris to be []interface{}")
			Expect(allowedRedirects).To(ContainElement("http://localhost:8250/oidc/callback"))
		})
	})

	Context("When deleting JWTOIDCAuthEngine resources", func() {
		It("Should clean up role from Vault and remove all resources", func() {

			Expect(mountInstance).NotTo(BeNil(), "expected auth mount to be created before delete phase")
			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")
			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before delete phase")

			By("Deleting the role CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
			roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.JWTOIDCAuthEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the role is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the config CR (IsDeletable=false, no Vault cleanup)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.JWTOIDCAuthEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the OIDC config still exists in Vault (IsDeletable=false means no Vault cleanup)")
			configSecret, err := vaultClient.Logical().Read("auth/test-jwt-oidc-auth/test-joaec-mount/config")
			Expect(err).To(BeNil())
			Expect(configSecret).NotTo(BeNil(), "expected OIDC config to persist in Vault after CR deletion")
			Expect(configSecret.Data["oidc_discovery_url"]).To(Equal("http://keycloak.keycloak.svc.cluster.local:8080/realms/test-realm"))

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
				_, exists := secret.Data["test-jwt-oidc-auth/test-joaec-mount/"]
				return !exists
			}, timeout, interval).Should(BeTrue())

			By("Deleting the OIDC credentials secret")
			Expect(k8sIntegrationClient.Delete(ctx, oidcSecret)).Should(Succeed())
		})
	})
})
