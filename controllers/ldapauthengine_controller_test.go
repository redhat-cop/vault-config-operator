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

var _ = Describe("LDAPAuthEngine controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var bindSecret *corev1.Secret
	var mountInstance *redhatcopv1alpha1.AuthEngineMount
	var configInstance *redhatcopv1alpha1.LDAPAuthEngineConfig
	var groupInstance *redhatcopv1alpha1.LDAPAuthEngineGroup

	AfterAll(func() {
		if groupInstance != nil {
			k8sIntegrationClient.Delete(ctx, groupInstance) //nolint:errcheck
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
		It("Should create the bind credentials secret and LDAP auth mount", func() {

			By("Creating the bind credentials K8s Secret")
			bindSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ldap-bind-creds",
					Namespace: vaultAdminNamespaceName,
				},
				Data: map[string][]byte{
					"username": []byte("cn=admin,dc=example,dc=com"),
					"password": []byte("admin"),
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, bindSecret)).Should(Succeed())

			By("Loading and creating the AuthEngineMount fixture")
			var err error
			mountInstance, err = decoder.GetAuthEngineMountInstance("../test/ldapauthengine/test-ldap-auth-mount.yaml")
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
			_, exists := secret.Data["test-ldap-auth/test-laec-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-ldap-auth/test-laec-mount/' in sys/auth")
		})
	})

	Context("When creating a LDAPAuthEngineConfig", func() {
		It("Should write the LDAP config to Vault", func() {

			By("Loading and creating the LDAPAuthEngineConfig fixture")
			var err error
			configInstance, err = decoder.GetLDAPAuthEngineConfigInstance("../test/ldapauthengine/test-ldap-auth-config.yaml")
			Expect(err).To(BeNil())
			configInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.LDAPAuthEngineConfig{}

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
			secret, err := vaultClient.Logical().Read("auth/test-ldap-auth/test-laec-mount/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["url"]).To(Equal("ldap://ldap.ldap.svc.cluster.local"))
			Expect(secret.Data["binddn"]).To(Equal("cn=admin,dc=example,dc=com"))
			Expect(secret.Data["insecure_tls"]).To(BeTrue())
		})
	})

	Context("When creating a LDAPAuthEngineGroup", func() {
		It("Should create the group mapping in Vault", func() {

			By("Loading and creating the LDAPAuthEngineGroup fixture")
			var err error
			groupInstance, err = decoder.GetLDAPAuthEngineGroupInstance("../test/ldapauthengine/test-ldap-auth-group.yaml")
			Expect(err).To(BeNil())
			groupInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, groupInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: groupInstance.Name, Namespace: groupInstance.Namespace}
			created := &redhatcopv1alpha1.LDAPAuthEngineGroup{}

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

			By("Verifying the group mapping in Vault")
			secret, err := vaultClient.Logical().Read("auth/test-ldap-auth/test-laec-mount/groups/admins-group")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			policies, ok := secret.Data["policies"].([]interface{})
			Expect(ok).To(BeTrue(), "expected policies to be []interface{}")
			Expect(policies).To(ContainElement("vault-admin"))
		})
	})

	Context("When deleting LDAPAuthEngine resources", func() {
		It("Should clean up group from Vault and remove all resources", func() {

			Expect(mountInstance).NotTo(BeNil(), "expected auth mount to be created before delete phase")
			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")
			Expect(groupInstance).NotTo(BeNil(), "expected group to be created before delete phase")

			By("Deleting the group CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, groupInstance)).Should(Succeed())
			groupLookupKey := types.NamespacedName{Name: groupInstance.Name, Namespace: groupInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, groupLookupKey, &redhatcopv1alpha1.LDAPAuthEngineGroup{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the group is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("auth/test-ldap-auth/test-laec-mount/groups/admins-group")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the config CR (IsDeletable=false, no Vault cleanup)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.LDAPAuthEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the LDAP config still exists in Vault (IsDeletable=false means no Vault cleanup)")
			configSecret, err := vaultClient.Logical().Read("auth/test-ldap-auth/test-laec-mount/config")
			Expect(err).To(BeNil())
			Expect(configSecret).NotTo(BeNil(), "expected LDAP config to persist in Vault after CR deletion")
			Expect(configSecret.Data["url"]).To(Equal("ldap://ldap.ldap.svc.cluster.local"))

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
				_, exists := secret.Data["test-ldap-auth/test-laec-mount/"]
				return !exists
			}, timeout, interval).Should(BeTrue())

			By("Deleting the bind credentials secret")
			Expect(k8sIntegrationClient.Delete(ctx, bindSecret)).Should(Succeed())
		})
	})
})
