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

var _ = Describe("KubernetesAuthEngine controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var mountInstance *redhatcopv1alpha1.AuthEngineMount
	var configInstance *redhatcopv1alpha1.KubernetesAuthEngineConfig
	var roleInstance *redhatcopv1alpha1.KubernetesAuthEngineRole
	var roleSelectorInstance *redhatcopv1alpha1.KubernetesAuthEngineRole

	AfterAll(func() {
		if roleSelectorInstance != nil {
			k8sIntegrationClient.Delete(ctx, roleSelectorInstance) //nolint:errcheck
		}
		if roleInstance != nil {
			k8sIntegrationClient.Delete(ctx, roleInstance) //nolint:errcheck
		}
		if configInstance != nil {
			k8sIntegrationClient.Delete(ctx, configInstance) //nolint:errcheck
		}
		if mountInstance != nil {
			k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
		}
	})

	Context("When creating the prerequisite auth mount", func() {
		It("Should enable the kubernetes auth method in Vault", func() {

			By("Loading and creating the AuthEngineMount fixture")
			var err error
			mountInstance, err = decoder.GetAuthEngineMountInstance("../test/kubernetesauthengine/test-kube-auth-mount.yaml")
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
			_, exists := secret.Data["test-k8s-auth/test-kaec-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-k8s-auth/test-kaec-mount/' in sys/auth")
		})
	})

	Context("When creating a KubernetesAuthEngineConfig", func() {
		It("Should write the config to Vault", func() {

			By("Loading and creating the KubernetesAuthEngineConfig fixture")
			var err error
			configInstance, err = decoder.GetKubernetesAuthEngineConfigInstance("../test/kubernetesauthengine/test-kube-auth-config.yaml")
			Expect(err).To(BeNil())
			configInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.KubernetesAuthEngineConfig{}

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
			secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["kubernetes_host"]).To(Equal("https://kubernetes.default.svc:443"))
		})
	})

	Context("When creating a KubernetesAuthEngineRole with explicit namespaces", func() {
		It("Should create the role in Vault with correct bindings", func() {

			By("Loading and creating the KubernetesAuthEngineRole fixture")
			var err error
			roleInstance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/kubernetesauthengine/test-kube-auth-role.yaml")
			Expect(err).To(BeNil())
			roleInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, roleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			created := &redhatcopv1alpha1.KubernetesAuthEngineRole{}

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
			secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			boundSANames, ok := secret.Data["bound_service_account_names"].([]interface{})
			Expect(ok).To(BeTrue(), "expected bound_service_account_names to be []interface{}")
			Expect(boundSANames).To(ContainElement("default"))

			boundSANamespaces, ok := secret.Data["bound_service_account_namespaces"].([]interface{})
			Expect(ok).To(BeTrue(), "expected bound_service_account_namespaces to be []interface{}")
			Expect(boundSANamespaces).To(ContainElement("vault-admin"))

			tokenPolicies, ok := secret.Data["token_policies"].([]interface{})
			Expect(ok).To(BeTrue(), "expected token_policies to be []interface{}")
			Expect(tokenPolicies).To(ContainElement("vault-admin"))
		})
	})

	Context("When creating a KubernetesAuthEngineRole with namespace selector", func() {
		It("Should resolve the selector and set bound namespaces", func() {

			By("Loading and creating the KubernetesAuthEngineRole selector fixture")
			var err error
			roleSelectorInstance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/kubernetesauthengine/test-kube-auth-role-selector.yaml")
			Expect(err).To(BeNil())
			roleSelectorInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, roleSelectorInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleSelectorInstance.Name, Namespace: roleSelectorInstance.Namespace}
			created := &redhatcopv1alpha1.KubernetesAuthEngineRole{}

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

			By("Verifying the role in Vault has resolved namespace selector")
			secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role-selector")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			boundSANamespaces, ok := secret.Data["bound_service_account_namespaces"].([]interface{})
			Expect(ok).To(BeTrue(), "expected bound_service_account_namespaces to be []interface{}")
			Expect(boundSANamespaces).To(ContainElement("test-vault-config-operator"))
		})
	})

	Context("When deleting KubernetesAuthEngine resources", func() {
		It("Should clean up roles from Vault and remove all resources", func() {

			Expect(mountInstance).NotTo(BeNil(), "expected auth mount to be created before delete phase")
			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")
			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before delete phase")
			Expect(roleSelectorInstance).NotTo(BeNil(), "expected role-selector to be created before delete phase")

			By("Deleting the explicit-namespace role CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
			roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.KubernetesAuthEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the selector-based role CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, roleSelectorInstance)).Should(Succeed())
			roleSelectorLookupKey := types.NamespacedName{Name: roleSelectorInstance.Name, Namespace: roleSelectorInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, roleSelectorLookupKey, &redhatcopv1alpha1.KubernetesAuthEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying both roles are removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role-selector")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the config CR (IsDeletable=false, no Vault cleanup)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.KubernetesAuthEngineConfig{})
				return apierrors.IsNotFound(err)
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
				_, exists := secret.Data["test-k8s-auth/test-kaec-mount/"]
				return !exists
			}, timeout, interval).Should(BeTrue())
		})
	})
})
