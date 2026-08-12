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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("LDAPSecretEngine controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var bindSecret *corev1.Secret
	var mountInstance *redhatcopv1alpha1.SecretEngineMount
	var configInstance *redhatcopv1alpha1.LDAPSecretEngineConfig
	var staticRoleInstance *redhatcopv1alpha1.LDAPSecretEngineStaticRole
	var dynamicRoleInstance *redhatcopv1alpha1.LDAPSecretEngineDynamicRole

	AfterAll(func() {
		if dynamicRoleInstance != nil {
			k8sIntegrationClient.Delete(ctx, dynamicRoleInstance) //nolint:errcheck
		}
		if staticRoleInstance != nil {
			k8sIntegrationClient.Delete(ctx, staticRoleInstance) //nolint:errcheck
		}
		if configInstance != nil {
			k8sIntegrationClient.Delete(ctx, configInstance) //nolint:errcheck
		}
		if mountInstance != nil {
			k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
		}
		if bindSecret != nil {
			k8sIntegrationClient.Delete(ctx, bindSecret) //nolint:errcheck
		}
	})

	Context("When creating prerequisite resources", func() {
		It("Should create the bind credentials secret and LDAP secret engine mount", func() {

			By("Creating the bind credentials K8s Secret")
			bindSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ldapse-bind-creds",
					Namespace: vaultAdminNamespaceName,
				},
				Data: map[string][]byte{
					"username": []byte("cn=admin,dc=example,dc=com"),
					"password": []byte("admin"),
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, bindSecret)).Should(Succeed())

			By("Loading and creating the SecretEngineMount fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ldapsecretengine/ldap-secret-engine-mount.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			mountInstance = &redhatcopv1alpha1.SecretEngineMount{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, mountInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			created := &redhatcopv1alpha1.SecretEngineMount{}

			By("Waiting for ReconcileSuccessful=True on mount")
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

			By("Verifying the mount exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/mounts")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			_, exists := secret.Data["test-ldapse/test-ldapse-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-ldapse/test-ldapse-mount/' in sys/mounts")
		})
	})

	Context("When creating a LDAPSecretEngineConfig", func() {
		It("Should write the LDAP secrets engine config to Vault", func() {

			By("Loading and creating the LDAPSecretEngineConfig fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ldapsecretengine/ldap-secret-engine-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			configInstance = &redhatcopv1alpha1.LDAPSecretEngineConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.LDAPSecretEngineConfig{}

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
			secret, err := vaultClient.Logical().Read("test-ldapse/test-ldapse-mount/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["url"]).To(Equal("ldap://ldap.ldap.svc.cluster.local"))
			Expect(secret.Data["binddn"]).To(Equal("cn=admin,dc=example,dc=com"))
			Expect(secret.Data["schema"]).To(Equal("openldap"))
		})
	})

	Context("When creating a LDAPSecretEngineStaticRole", func() {
		It("Should create the static role in Vault", func() {

			By("Loading and creating the LDAPSecretEngineStaticRole fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ldapsecretengine/ldap-secret-engine-static-role.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			staticRoleInstance = &redhatcopv1alpha1.LDAPSecretEngineStaticRole{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, staticRoleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: staticRoleInstance.Name, Namespace: staticRoleInstance.Namespace}
			created := &redhatcopv1alpha1.LDAPSecretEngineStaticRole{}

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

			By("Verifying the static role in Vault")
			secret, err := vaultClient.Logical().Read("test-ldapse/test-ldapse-mount/static-role/test-ldapse-static-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["username"]).To(Equal("trevor"))
			Expect(secret.Data["dn"]).To(Equal("uid=trevor,ou=Users,dc=example,dc=com"))
		})
	})

	Context("When creating a LDAPSecretEngineDynamicRole", func() {
		It("Should create the dynamic role in Vault", func() {

			By("Loading and creating the LDAPSecretEngineDynamicRole fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ldapsecretengine/ldap-secret-engine-dynamic-role.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			dynamicRoleInstance = &redhatcopv1alpha1.LDAPSecretEngineDynamicRole{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, dynamicRoleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: dynamicRoleInstance.Name, Namespace: dynamicRoleInstance.Namespace}
			created := &redhatcopv1alpha1.LDAPSecretEngineDynamicRole{}

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

			By("Verifying the dynamic role in Vault")
			secret, err := vaultClient.Logical().Read("test-ldapse/test-ldapse-mount/role/test-ldapse-dynamic-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["creation_ldif"]).NotTo(BeEmpty())
			Expect(secret.Data["deletion_ldif"]).NotTo(BeEmpty())
		})
	})

	Context("When updating a LDAPSecretEngineConfig mutable field", func() {
		It("Should write the updated config URL to Vault", func() {

			Expect(configInstance).NotTo(BeNil(), "expected config to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			current := &redhatcopv1alpha1.LDAPSecretEngineConfig{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating the config URL")
			current.Spec.LDAPSEConfig.URL = "ldap://ldap-updated.ldap.svc.cluster.local"
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated url")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-ldapse/test-ldapse-mount/config")
				if err != nil || secret == nil {
					return false
				}
				return secret.Data["url"] == "ldap://ldap-updated.ldap.svc.cluster.local"
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.LDAPSecretEngineConfig{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				for _, condition := range updated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful {
						return condition.ObservedGeneration > initialGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When updating a LDAPSecretEngineDynamicRole mutable field", func() {
		It("Should write the updated default_ttl to Vault", func() {

			Expect(dynamicRoleInstance).NotTo(BeNil(), "expected dynamic role to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: dynamicRoleInstance.Name, Namespace: dynamicRoleInstance.Namespace}
			current := &redhatcopv1alpha1.LDAPSecretEngineDynamicRole{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating defaultTTL to 2h")
			current.Spec.LDAPSEDynamicRole.DefaultTTL = "2h"
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated default_ttl")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-ldapse/test-ldapse-mount/role/test-ldapse-dynamic-role")
				if err != nil || secret == nil {
					return false
				}
				ttl, ok := secret.Data["default_ttl"].(json.Number)
				if !ok {
					return false
				}
				val, err := ttl.Int64()
				if err != nil {
					return false
				}
				return val == int64(7200) // 2h in seconds
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.LDAPSecretEngineDynamicRole{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				for _, condition := range updated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful {
						return condition.ObservedGeneration > initialGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting LDAPSecretEngine resources", func() {
		It("Should clean up roles and config from Vault", func() {

			Expect(mountInstance).NotTo(BeNil(), "expected mount to be created before delete phase")
			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")

			By("Deleting the dynamic role CR (IsDeletable=true)")
			if dynamicRoleInstance != nil {
				Expect(k8sIntegrationClient.Delete(ctx, dynamicRoleInstance)).Should(Succeed())
				dynamicLookupKey := types.NamespacedName{Name: dynamicRoleInstance.Name, Namespace: dynamicRoleInstance.Namespace}
				Eventually(func() bool {
					err := k8sIntegrationClient.Get(ctx, dynamicLookupKey, &redhatcopv1alpha1.LDAPSecretEngineDynamicRole{})
					return apierrors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())

				By("Verifying the dynamic role is removed from Vault")
				Eventually(func() bool {
					secret, err := vaultClient.Logical().Read("test-ldapse/test-ldapse-mount/role/test-ldapse-dynamic-role")
					return err == nil && secret == nil
				}, timeout, interval).Should(BeTrue())
			}

			By("Deleting the static role CR (IsDeletable=true)")
			if staticRoleInstance != nil {
				Expect(k8sIntegrationClient.Delete(ctx, staticRoleInstance)).Should(Succeed())
				staticLookupKey := types.NamespacedName{Name: staticRoleInstance.Name, Namespace: staticRoleInstance.Namespace}
				Eventually(func() bool {
					err := k8sIntegrationClient.Get(ctx, staticLookupKey, &redhatcopv1alpha1.LDAPSecretEngineStaticRole{})
					return apierrors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())

				By("Verifying the static role is removed from Vault")
				Eventually(func() bool {
					secret, err := vaultClient.Logical().Read("test-ldapse/test-ldapse-mount/static-role/test-ldapse-static-role")
					return err == nil && secret == nil
				}, timeout, interval).Should(BeTrue())
			}

			By("Deleting the config CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.LDAPSecretEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the config is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-ldapse/test-ldapse-mount/config")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the SecretEngineMount")
			Expect(k8sIntegrationClient.Delete(ctx, mountInstance)).Should(Succeed())
			mountLookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, mountLookupKey, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the bind credentials secret")
			Expect(k8sIntegrationClient.Delete(ctx, bindSecret)).Should(Succeed())
		})
	})
})
