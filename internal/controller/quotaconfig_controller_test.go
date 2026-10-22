//go:build integration_quota
// +build integration_quota

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

var _ = Describe("QuotaConfig controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var quotaConfigInstance *redhatcopv1alpha1.QuotaConfig

	AfterAll(func() {
		if quotaConfigInstance != nil {
			k8sIntegrationClient.Delete(ctx, quotaConfigInstance) //nolint:errcheck
		}
	})

	Context("When creating a QuotaConfig", func() {
		It("Should configure the global quota config in Vault and reconcile successfully", func() {

			By("Loading and creating the QuotaConfig fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/quotaconfig/quota-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			quotaConfigInstance = &redhatcopv1alpha1.QuotaConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, quotaConfigInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: quotaConfigInstance.Name, Namespace: quotaConfigInstance.Namespace}
			created := &redhatcopv1alpha1.QuotaConfig{}

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

			By("Verifying the quota config exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/quotas/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
		})
	})

	Context("When updating a QuotaConfig", func() {
		It("Should update the global quota config in Vault", func() {
			Expect(quotaConfigInstance).NotTo(BeNil())

			lookupKey := types.NamespacedName{Name: quotaConfigInstance.Name, Namespace: quotaConfigInstance.Namespace}
			current := &redhatcopv1alpha1.QuotaConfig{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())

			By("Updating the QuotaConfig CR")
			current.Spec.EnableRateLimitAuditLogging = true
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for ReconcileSuccessful=True after update")
			Eventually(func() bool {
				updated := &redhatcopv1alpha1.QuotaConfig{}
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

	Context("When deleting a QuotaConfig", func() {
		It("Should not delete the global quota config from Vault (IsDeletable=false)", func() {
			Expect(quotaConfigInstance).NotTo(BeNil())

			By("Deleting the QuotaConfig CR")
			Expect(k8sIntegrationClient.Delete(ctx, quotaConfigInstance)).Should(Succeed())

			By("Verifying the global quota config still exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/quotas/config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
		})
	})
})
