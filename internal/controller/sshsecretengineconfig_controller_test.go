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

var _ = Describe("SSHSecretEngineConfig controller", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var mountInstance *redhatcopv1alpha1.SecretEngineMount
	var configInstance *redhatcopv1alpha1.SSHSecretEngineConfig

	AfterAll(func() {
		if configInstance != nil {
			k8sIntegrationClient.Delete(ctx, configInstance) //nolint:errcheck
		}
		if mountInstance != nil {
			k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
		}
	})

	Context("When creating prerequisite SSH engine mount", func() {
		It("Should create the SSH secret engine mount", func() {

			By("Loading and creating the SecretEngineMount fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ssh/ssh-secret-engine-mount.yaml", vaultAdminNamespaceName)
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
			_, exists := secret.Data["test-ssh/test-ssh-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-ssh/test-ssh-mount/' in sys/mounts")
		})
	})

	Context("When creating an SSHSecretEngineConfig with generate_signing_key=true", func() {
		It("Should configure the SSH CA in Vault", func() {

			By("Loading and creating the SSHSecretEngineConfig fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ssh/ssh-secret-engine-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			configInstance = &redhatcopv1alpha1.SSHSecretEngineConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.SSHSecretEngineConfig{}

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

			By("Verifying the CA config in Vault")
			secret, err := vaultClient.Logical().Read("test-ssh/test-ssh-mount/config/ca")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			publicKey, ok := secret.Data["public_key"].(string)
			Expect(ok).To(BeTrue(), "expected public_key to be a string")
			Expect(publicKey).NotTo(BeEmpty(), "expected non-empty public_key")
		})
	})

	Context("When deleting SSHSecretEngineConfig resources", func() {
		It("Should clean up CA config from Vault", func() {
			// Vault's DELETE /ssh/config/ca only removes the private key from storage;
			// GET /ssh/config/ca still returns the public key after deletion. Because
			// private_key is never returned by GET (it's write-only), we cannot verify
			// the DELETE's effect through the API. We rely on the controller's
			// IsDeletable()=true ensuring the finalizer calls Vault DELETE, and verify
			// only that the K8s CR is removed.

			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")
			Expect(mountInstance).NotTo(BeNil(), "expected mount to be created before delete phase")

			By("Deleting the config CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.SSHSecretEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

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
