//go:build integration
// +build integration

package controllers

import (
	"encoding/json"
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

var _ = Describe("RabbitMQSecretEngine controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var rmqSecret *corev1.Secret
	var mountInstance *redhatcopv1alpha1.SecretEngineMount
	var configInstance *redhatcopv1alpha1.RabbitMQSecretEngineConfig
	var roleInstance *redhatcopv1alpha1.RabbitMQSecretEngineRole

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
		if rmqSecret != nil {
			k8sIntegrationClient.Delete(ctx, rmqSecret) //nolint:errcheck
		}
	})

	Context("When creating prerequisite resources", func() {
		It("Should create the RabbitMQ credentials secret and rabbitmq engine mount", func() {

			By("Creating the RabbitMQ credentials K8s Secret")
			rmqSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rmq-creds",
					Namespace: vaultAdminNamespaceName,
				},
				StringData: map[string]string{
					"password": "testpassword123",
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, rmqSecret)).Should(Succeed())

			By("Loading and creating the SecretEngineMount fixture")
			var err error
			mountInstance, err = decoder.GetSecretEngineMountInstance("../test/rabbitmqsecretengine/test-rmq-mount.yaml")
			Expect(err).To(BeNil())
			mountInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, mountInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			created := &redhatcopv1alpha1.SecretEngineMount{}

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

			By("Verifying the mount exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/mounts")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			_, exists := secret.Data["test-rmqse/test-rmq-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-rmqse/test-rmq-mount/' in sys/mounts")
		})
	})

	Context("When creating a RabbitMQSecretEngineConfig", func() {
		It("Should write the RabbitMQ connection and lease config to Vault", func() {

			By("Loading and creating the RabbitMQSecretEngineConfig fixture")
			var err error
			configInstance, err = decoder.GetRabbitMQSecretEngineConfigInstance("../test/rabbitmqsecretengine/test-rmq-config.yaml")
			Expect(err).To(BeNil())
			configInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.RabbitMQSecretEngineConfig{}

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

			By("Verifying the lease config in Vault (connection config is write-only in Vault's RabbitMQ engine)")
			leaseSecret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/config/lease")
			Expect(err).To(BeNil())
			Expect(leaseSecret).NotTo(BeNil())

			ttl, ok := leaseSecret.Data["ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected ttl to be json.Number")
			ttlVal, err := ttl.Int64()
			Expect(err).To(BeNil())
			Expect(ttlVal).To(Equal(int64(3600)))

			maxTTL, ok := leaseSecret.Data["max_ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected max_ttl to be json.Number")
			maxTTLVal, err := maxTTL.Int64()
			Expect(err).To(BeNil())
			Expect(maxTTLVal).To(Equal(int64(86400)))
		})
	})

	Context("When creating a RabbitMQSecretEngineRole", func() {
		It("Should create the role in Vault with correct settings", func() {

			By("Loading and creating the RabbitMQSecretEngineRole fixture")
			var err error
			roleInstance, err = decoder.GetRabbitMQSecretEngineRoleInstance("../test/rabbitmqsecretengine/test-rmq-role.yaml")
			Expect(err).To(BeNil())
			roleInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, roleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			created := &redhatcopv1alpha1.RabbitMQSecretEngineRole{}

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
			secret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/roles/test-rmq-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			tags, ok := secret.Data["tags"].(string)
			Expect(ok).To(BeTrue(), "expected tags to be a string")
			Expect(tags).To(Equal("administrator"))

			vhosts, ok := secret.Data["vhosts"].(string)
			Expect(ok).To(BeTrue(), "expected vhosts to be a JSON string")
			Expect(vhosts).To(ContainSubstring("configure"))
			Expect(vhosts).To(ContainSubstring("read"))
			Expect(vhosts).To(ContainSubstring("write"))
		})
	})

	Context("When updating a RabbitMQSecretEngineRole", func() {
		It("Should update the role in Vault and reflect updated ObservedGeneration", func() {

			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			current := &redhatcopv1alpha1.RabbitMQSecretEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating tags to management")
			current.Spec.RMQSERole.Tags = "management"
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated tags")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/roles/test-rmq-role")
				if err != nil || secret == nil {
					return false
				}
				tags, ok := secret.Data["tags"].(string)
				if !ok {
					return false
				}
				return tags == "management"
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.RabbitMQSecretEngineRole{}
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

	Context("When deleting RabbitMQSecretEngine resources", func() {
		It("Should clean up role from Vault, preserve config in Vault, and remove all K8s resources", func() {

			Expect(mountInstance).NotTo(BeNil(), "expected mount to be created before delete phase")
			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")
			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before delete phase")

			By("Deleting the role CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
			roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.RabbitMQSecretEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the role is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/roles/test-rmq-role")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the config CR (IsDeletable=false, no Vault cleanup)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.RabbitMQSecretEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the lease config still exists in Vault (IsDeletable=false means no Vault cleanup; connection endpoint is write-only)")
			leaseSecret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/config/lease")
			Expect(err).To(BeNil())
			Expect(leaseSecret).NotTo(BeNil(), "expected lease config to persist in Vault after CR deletion")

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
				_, exists := secret.Data["test-rmqse/test-rmq-mount/"]
				return !exists
			}, timeout, interval).Should(BeTrue())

			By("Deleting the RabbitMQ credentials secret")
			Expect(k8sIntegrationClient.Delete(ctx, rmqSecret)).Should(Succeed())
		})
	})
})
