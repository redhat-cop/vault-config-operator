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
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("KubernetesSecretEngine controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var saTokenSecret *corev1.Secret
	var mountInstance *redhatcopv1alpha1.SecretEngineMount
	var configInstance *redhatcopv1alpha1.KubernetesSecretEngineConfig
	var roleInstance *redhatcopv1alpha1.KubernetesSecretEngineRole

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
		if saTokenSecret != nil {
			k8sIntegrationClient.Delete(ctx, saTokenSecret) //nolint:errcheck
		}
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "test-kubese-sa", Namespace: vaultAdminNamespaceName},
		}
		k8sIntegrationClient.Delete(ctx, sa) //nolint:errcheck
		crb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "test-kubese-sa-cluster-admin"},
		}
		k8sIntegrationClient.Delete(ctx, crb) //nolint:errcheck
	})

	Context("When creating prerequisite resources", func() {
		It("Should create the SA, RBAC, token secret, and kubernetes engine mount", func() {

			By("Creating the ServiceAccount")
			sa := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-kubese-sa",
					Namespace: vaultAdminNamespaceName,
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, sa)).Should(Succeed())

			By("Creating the ClusterRoleBinding")
			crb := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-kubese-sa-cluster-admin",
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "cluster-admin",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "test-kubese-sa",
						Namespace: vaultAdminNamespaceName,
					},
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, crb)).Should(Succeed())

			By("Creating the SA token K8s Secret")
			saTokenSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-kubese-sa-token",
					Namespace: vaultAdminNamespaceName,
					Annotations: map[string]string{
						corev1.ServiceAccountNameKey: "test-kubese-sa",
					},
				},
				Type: corev1.SecretTypeServiceAccountToken,
			}
			Expect(k8sIntegrationClient.Create(ctx, saTokenSecret)).Should(Succeed())

			By("Waiting for the token controller to populate the token")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, types.NamespacedName{
					Name: "test-kubese-sa-token", Namespace: vaultAdminNamespaceName,
				}, saTokenSecret)
				if err != nil {
					return false
				}
				_, hasToken := saTokenSecret.Data[corev1.ServiceAccountTokenKey]
				return hasToken
			}, timeout, interval).Should(BeTrue())

			By("Loading and creating the SecretEngineMount fixture")
			var err error
			mountInstance, err = decoder.GetSecretEngineMountInstance("../test/kubernetessecretengine/test-kubese-mount.yaml")
			Expect(err).To(BeNil())
			mountInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, mountInstance)).Should(Succeed())

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
			_, exists := secret.Data["test-kubese/test-kubese-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-kubese/test-kubese-mount/' in sys/mounts")
		})
	})

	Context("When creating a KubernetesSecretEngineConfig", func() {
		It("Should write the Kubernetes config to Vault", func() {

			By("Loading and creating the KubernetesSecretEngineConfig fixture")
			var err error
			configInstance, err = decoder.GetKubernetesSecretEngineConfigInstance("../test/kubernetessecretengine/test-kubese-config.yaml")
			Expect(err).To(BeNil())
			configInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.KubernetesSecretEngineConfig{}

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
			secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			kubeHost, ok := secret.Data["kubernetes_host"].(string)
			Expect(ok).To(BeTrue(), "expected kubernetes_host to be a string")
			Expect(kubeHost).To(Equal("https://kubernetes.default.svc:443"))

			disableLocalCA, ok := secret.Data["disable_local_ca_jwt"].(bool)
			Expect(ok).To(BeTrue(), "expected disable_local_ca_jwt to be a bool")
			Expect(disableLocalCA).To(BeFalse())

			_, jwtPresent := secret.Data["service_account_jwt"]
			Expect(jwtPresent).To(BeFalse(), "service_account_jwt must not be returned by Vault")
		})
	})

	Context("When creating a KubernetesSecretEngineRole", func() {
		It("Should create the role in Vault with correct settings", func() {

			By("Loading and creating the KubernetesSecretEngineRole fixture")
			var err error
			roleInstance, err = decoder.GetKubernetesSecretEngineRoleInstance("../test/kubernetessecretengine/test-kubese-role.yaml")
			Expect(err).To(BeNil())
			roleInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, roleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			created := &redhatcopv1alpha1.KubernetesSecretEngineRole{}

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
			secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/roles/test-kubese-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			roleName, ok := secret.Data["kubernetes_role_name"].(string)
			Expect(ok).To(BeTrue(), "expected kubernetes_role_name to be a string")
			Expect(roleName).To(Equal("edit"))

			roleType, ok := secret.Data["kubernetes_role_type"].(string)
			Expect(ok).To(BeTrue(), "expected kubernetes_role_type to be a string")
			Expect(roleType).To(Equal("ClusterRole"))

			allowedNs, ok := secret.Data["allowed_kubernetes_namespaces"].([]interface{})
			Expect(ok).To(BeTrue(), "expected allowed_kubernetes_namespaces to be []interface{}")
			Expect(allowedNs).To(ContainElement("default"))
		})
	})

	Context("When updating a KubernetesSecretEngineRole", func() {
		It("Should update the role in Vault and reflect updated ObservedGeneration", func() {

			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			current := &redhatcopv1alpha1.KubernetesSecretEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating kubernetesRoleName to 'view'")
			current.Spec.KubeSERole.KubernetesRoleName = "view"
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated kubernetes_role_name")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/roles/test-kubese-role")
				if err != nil || secret == nil {
					return false
				}
				roleName, ok := secret.Data["kubernetes_role_name"].(string)
				if !ok {
					return false
				}
				return roleName == "view"
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.KubernetesSecretEngineRole{}
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

	Context("When deleting KubernetesSecretEngine resources", func() {
		It("Should clean up role and config from Vault and remove all K8s resources", func() {

			Expect(mountInstance).NotTo(BeNil(), "expected mount to be created before delete phase")
			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")
			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before delete phase")

			By("Deleting the role CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
			roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.KubernetesSecretEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the role is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/roles/test-kubese-role")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the config CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.KubernetesSecretEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the config is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/config")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the SecretEngineMount")
			Expect(k8sIntegrationClient.Delete(ctx, mountInstance)).Should(Succeed())
			mountLookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, mountLookupKey, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the mount is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/mounts")
				if err != nil || secret == nil {
					return false
				}
				_, exists := secret.Data["test-kubese/test-kubese-mount/"]
				return !exists
			}, timeout, interval).Should(BeTrue())

			By("Deleting the SA token secret")
			Expect(k8sIntegrationClient.Delete(ctx, saTokenSecret)).Should(Succeed())

			By("Deleting the ServiceAccount and ClusterRoleBinding")
			sa := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: "test-kubese-sa", Namespace: vaultAdminNamespaceName},
			}
			Expect(k8sIntegrationClient.Delete(ctx, sa)).Should(Succeed())
			crb := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "test-kubese-sa-cluster-admin"},
			}
			Expect(k8sIntegrationClient.Delete(ctx, crb)).Should(Succeed())
		})
	})
})
