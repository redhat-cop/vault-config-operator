//go:build integration_quota
// +build integration_quota

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

var _ = Describe("RateLimitQuota controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var rateLimitInstance *redhatcopv1alpha1.RateLimitQuota

	AfterAll(func() {
		if rateLimitInstance != nil {
			k8sIntegrationClient.Delete(ctx, rateLimitInstance) //nolint:errcheck
		}
	})

	Context("When creating a RateLimitQuota", func() {
		It("Should create the rate limit quota in Vault and reconcile successfully", func() {

			By("Loading and creating the RateLimitQuota fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/quotaconfig/rate-limit-quota.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			rateLimitInstance = &redhatcopv1alpha1.RateLimitQuota{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, rateLimitInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: rateLimitInstance.Name, Namespace: rateLimitInstance.Namespace}
			created := &redhatcopv1alpha1.RateLimitQuota{}

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

			By("Verifying the rate limit quota exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/quotas/rate-limit/test-rate-limit-quota")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
		})
	})

	Context("When updating a RateLimitQuota", func() {
		It("Should update the rate limit quota in Vault", func() {
			Expect(rateLimitInstance).NotTo(BeNil())

			lookupKey := types.NamespacedName{Name: rateLimitInstance.Name, Namespace: rateLimitInstance.Namespace}
			current := &redhatcopv1alpha1.RateLimitQuota{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())

			By("Updating the rate from 100 to 200")
			current.Spec.Rate = "200"
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for ReconcileSuccessful=True after update")
			Eventually(func() bool {
				updated := &redhatcopv1alpha1.RateLimitQuota{}
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

	Context("When deleting a RateLimitQuota", func() {
		It("Should remove the rate limit quota from Vault", func() {
			Expect(rateLimitInstance).NotTo(BeNil())

			By("Deleting the RateLimitQuota CR")
			Expect(k8sIntegrationClient.Delete(ctx, rateLimitInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: rateLimitInstance.Name, Namespace: rateLimitInstance.Namespace}
			By("Waiting for the RateLimitQuota to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.RateLimitQuota{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the rate limit quota no longer exists in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/quotas/rate-limit/test-rate-limit-quota")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
