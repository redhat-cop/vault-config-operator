//go:build integration
// +build integration

package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("NomadSecretEngine controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var mountInstance *redhatcopv1alpha1.SecretEngineMount
	var configInstance *redhatcopv1alpha1.NomadSecretEngineConfig
	var roleInstance *redhatcopv1alpha1.NomadSecretEngineRole

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
	})

	Context("When creating prerequisite resources", func() {
		It("Should create the Nomad secret engine mount", func() {

			By("Loading and creating the SecretEngineMount fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/nomadsecretengine/nomad-secret-engine-mount.yaml", vaultAdminNamespaceName)
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
			_, exists := secret.Data["test-nomadse/test-nomadse-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-nomadse/test-nomadse-mount/' in sys/mounts")
		})
	})

	Context("When creating a NomadSecretEngineConfig", func() {
		It("Should write the Nomad secrets engine config to Vault", func() {

			By("Loading and creating the NomadSecretEngineConfig fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/nomadsecretengine/nomad-secret-engine-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			configInstance = &redhatcopv1alpha1.NomadSecretEngineConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.NomadSecretEngineConfig{}

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
			secret, err := vaultClient.Logical().Read("test-nomadse/test-nomadse-mount/config/access")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["address"]).To(Equal("http://nomad.nomad.svc.cluster.local:4646"))
		})
	})

	Context("When creating a NomadSecretEngineRole", func() {
		It("Should create the role in Vault", func() {

			By("Loading and creating the NomadSecretEngineRole fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/nomadsecretengine/nomad-secret-engine-role.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			roleInstance = &redhatcopv1alpha1.NomadSecretEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, roleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			created := &redhatcopv1alpha1.NomadSecretEngineRole{}

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
			secret, err := vaultClient.Logical().Read("test-nomadse/test-nomadse-mount/role/nomad-role-test")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["token_type"]).To(Equal("client"))
		})
	})

	Context("When generating dynamic Nomad credentials", func() {
		It("Should generate a Nomad ACL token via creds/{name}", func() {

			By("Requesting dynamic credentials from the role")
			secret, err := vaultClient.Logical().Read("test-nomadse/test-nomadse-mount/creds/nomad-role-test")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil(), "expected creds endpoint to return a secret")
			Expect(secret.Data["secret_id"]).NotTo(BeEmpty(), "expected a non-empty secret_id (Nomad ACL token)")
		})
	})

	Context("When deleting NomadSecretEngine resources", func() {
		It("Should clean up role from Vault and preserve config", func() {

			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")

			By("Deleting the role CR (IsDeletable=true)")
			if roleInstance != nil {
				Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
				roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
				Eventually(func() bool {
					err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.NomadSecretEngineRole{})
					return apierrors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())

				By("Verifying the role is removed from Vault")
				Eventually(func() bool {
					secret, err := vaultClient.Logical().Read("test-nomadse/test-nomadse-mount/role/nomad-role-test")
					return err == nil && secret == nil
				}, timeout, interval).Should(BeTrue())
			}

			By("Deleting the config CR (IsDeletable=false)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.NomadSecretEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the config persists in Vault (IsDeletable=false)")
			secret, err := vaultClient.Logical().Read("test-nomadse/test-nomadse-mount/config/access")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["address"]).To(Equal("http://nomad.nomad.svc.cluster.local:4646"))

			By("Deleting the SecretEngineMount")
			Expect(k8sIntegrationClient.Delete(ctx, mountInstance)).Should(Succeed())
			mountLookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, mountLookupKey, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
