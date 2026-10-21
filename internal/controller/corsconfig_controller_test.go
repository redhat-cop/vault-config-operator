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

var _ = Describe("CORSConfig controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var corsConfigInstance *redhatcopv1alpha1.CORSConfig

	AfterAll(func() {
		if corsConfigInstance != nil {
			k8sIntegrationClient.Delete(ctx, corsConfigInstance) //nolint:errcheck
		}
	})

	Context("When creating a CORSConfig", func() {
		It("Should configure CORS in Vault and reconcile successfully", func() {

			By("Loading and creating the CORSConfig fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/corsconfig/cors-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			corsConfigInstance = &redhatcopv1alpha1.CORSConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, corsConfigInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: corsConfigInstance.Name, Namespace: corsConfigInstance.Namespace}
			created := &redhatcopv1alpha1.CORSConfig{}

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

			By("Verifying CORS is enabled in Vault")
			secret, err := vaultClient.Logical().Read("sys/config/cors")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["enabled"]).To(BeTrue())
		})
	})

	Context("When updating a CORSConfig", func() {
		It("Should update CORS config in Vault", func() {
			Expect(corsConfigInstance).NotTo(BeNil())

			lookupKey := types.NamespacedName{Name: corsConfigInstance.Name, Namespace: corsConfigInstance.Namespace}
			current := &redhatcopv1alpha1.CORSConfig{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())

			By("Updating the CORSConfig CR")
			current.Spec.AllowedOrigins = []string{"https://example.com", "https://updated.com"}
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for ReconcileSuccessful=True after update")
			Eventually(func() bool {
				updated := &redhatcopv1alpha1.CORSConfig{}
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

	Context("When deleting a CORSConfig", func() {
		It("Should disable CORS in Vault (IsDeletable=true)", func() {
			Expect(corsConfigInstance).NotTo(BeNil())

			By("Deleting the CORSConfig CR")
			Expect(k8sIntegrationClient.Delete(ctx, corsConfigInstance)).Should(Succeed())

			By("Verifying CORS is disabled in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/config/cors")
				if err != nil {
					return false
				}
				if secret == nil {
					return true
				}
				enabled, ok := secret.Data["enabled"]
				if !ok {
					return true
				}
				return enabled == false
			}, timeout, interval).Should(BeTrue())
		})
	})
})
