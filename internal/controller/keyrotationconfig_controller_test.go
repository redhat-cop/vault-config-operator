//go:build integration
// +build integration

package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("KeyRotationConfig controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var keyRotationConfigInstance *redhatcopv1alpha1.KeyRotationConfig

	AfterAll(func() {
		if keyRotationConfigInstance != nil {
			k8sIntegrationClient.Delete(ctx, keyRotationConfigInstance) //nolint:errcheck
		}
	})

	Context("When creating a KeyRotationConfig", func() {
		It("Should configure key rotation in Vault and reconcile successfully", func() {

			By("Loading and creating the KeyRotationConfig fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/keyrotationconfig/key-rotation-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			keyRotationConfigInstance = &redhatcopv1alpha1.KeyRotationConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, keyRotationConfigInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: keyRotationConfigInstance.Name, Namespace: keyRotationConfigInstance.Namespace}
			created := &redhatcopv1alpha1.KeyRotationConfig{}

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

			By("Verifying the rotation config exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/rotate/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
		})
	})

	Context("When updating a KeyRotationConfig", func() {
		It("Should update the rotation config in Vault", func() {
			Expect(keyRotationConfigInstance).NotTo(BeNil())

			lookupKey := types.NamespacedName{Name: keyRotationConfigInstance.Name, Namespace: keyRotationConfigInstance.Namespace}
			current := &redhatcopv1alpha1.KeyRotationConfig{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())

			By("Updating the KeyRotationConfig CR")
			current.Spec.Interval = "8640h"
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for ReconcileSuccessful=True after update")
			Eventually(func() bool {
				updated := &redhatcopv1alpha1.KeyRotationConfig{}
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				for _, condition := range updated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting a KeyRotationConfig", func() {
		It("Should not delete the rotation config from Vault (IsDeletable=false)", func() {
			Expect(keyRotationConfigInstance).NotTo(BeNil())

			By("Deleting the KeyRotationConfig CR")
			Expect(k8sIntegrationClient.Delete(ctx, keyRotationConfigInstance)).Should(Succeed())

			By("Verifying the rotation config still exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/rotate/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
		})
	})
})
