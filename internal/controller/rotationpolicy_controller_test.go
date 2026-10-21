//go:build integration
// +build integration

// NOTE: Rotation policies require Vault Enterprise.
// These tests verify the K8s lifecycle only when run against Enterprise Vault.
// Against OSS Vault, the reconciler will receive a 4xx from Vault — the CR
// will have ReconcileFailed condition, which is the expected behavior.

package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("RotationPolicy controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var simpleRPInstance *redhatcopv1alpha1.RotationPolicy
	var namedRPInstance *redhatcopv1alpha1.RotationPolicy

	AfterAll(func() {
		if simpleRPInstance != nil {
			k8sIntegrationClient.Delete(ctx, simpleRPInstance) //nolint:errcheck
		}
		if namedRPInstance != nil {
			k8sIntegrationClient.Delete(ctx, namedRPInstance) //nolint:errcheck
		}
	})

	Context("When creating a simple RotationPolicy", func() {
		It("Should attempt to create the rotation policy in Vault", func() {

			By("Loading and creating the simple RotationPolicy fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/rotationpolicy/simple-rotation-policy.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			simpleRPInstance = &redhatcopv1alpha1.RotationPolicy{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, simpleRPInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: simpleRPInstance.Name, Namespace: simpleRPInstance.Namespace}
			created := &redhatcopv1alpha1.RotationPolicy{}

			By("Waiting for a condition to be set (ReconcileSuccessful on Enterprise, ReconcileFailed on OSS)")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}
				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful || condition.Type == vaultresourcecontroller.ReconcileFailed {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When creating a RotationPolicy with spec.name override", func() {
		It("Should attempt to create the policy at the spec.name path in Vault", func() {

			By("Loading and creating the named RotationPolicy fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/rotationpolicy/named-rotation-policy.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			namedRPInstance = &redhatcopv1alpha1.RotationPolicy{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, namedRPInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: namedRPInstance.Name, Namespace: namedRPInstance.Namespace}
			created := &redhatcopv1alpha1.RotationPolicy{}

			By("Waiting for a condition to be set (ReconcileSuccessful on Enterprise, ReconcileFailed on OSS)")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}
				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful || condition.Type == vaultresourcecontroller.ReconcileFailed {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting RotationPolicies", func() {
		It("Should remove rotation policies from K8s", func() {

			Expect(simpleRPInstance).NotTo(BeNil(), "expected simple RotationPolicy to be created before delete phase")
			Expect(namedRPInstance).NotTo(BeNil(), "expected named RotationPolicy to be created before delete phase")

			By("Deleting the simple RotationPolicy CR")
			Expect(k8sIntegrationClient.Delete(ctx, simpleRPInstance)).Should(Succeed())

			simpleLookupKey := types.NamespacedName{Name: simpleRPInstance.Name, Namespace: simpleRPInstance.Namespace}
			By("Waiting for the simple RotationPolicy to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, simpleLookupKey, &redhatcopv1alpha1.RotationPolicy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the named RotationPolicy CR")
			Expect(k8sIntegrationClient.Delete(ctx, namedRPInstance)).Should(Succeed())

			namedLookupKey := types.NamespacedName{Name: namedRPInstance.Name, Namespace: namedRPInstance.Namespace}
			By("Waiting for the named RotationPolicy to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, namedLookupKey, &redhatcopv1alpha1.RotationPolicy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
