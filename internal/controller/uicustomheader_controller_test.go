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

var _ = Describe("UICustomHeader controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var simpleHeaderInstance *redhatcopv1alpha1.UICustomHeader
	var namedHeaderInstance *redhatcopv1alpha1.UICustomHeader

	AfterAll(func() {
		if simpleHeaderInstance != nil {
			k8sIntegrationClient.Delete(ctx, simpleHeaderInstance) //nolint:errcheck
		}
		if namedHeaderInstance != nil {
			k8sIntegrationClient.Delete(ctx, namedHeaderInstance) //nolint:errcheck
		}
	})

	Context("When creating a simple UICustomHeader", func() {
		It("Should configure the header in Vault", func() {

			By("Loading and creating the simple UICustomHeader fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/uicustomheader/simple-ui-custom-header.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			simpleHeaderInstance = &redhatcopv1alpha1.UICustomHeader{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, simpleHeaderInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: simpleHeaderInstance.Name, Namespace: simpleHeaderInstance.Namespace}
			created := &redhatcopv1alpha1.UICustomHeader{}

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

			By("Verifying the header exists in Vault with correct value")
			secret, err := vaultClient.Logical().Read("sys/config/ui/headers/x-custom-test-header")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data).NotTo(BeNil())
			Expect(secret.Data["value"]).To(Equal("test-value-1"))
		})
	})

	Context("When creating a UICustomHeader with spec.name override", func() {
		It("Should configure the header at the spec.name path in Vault", func() {

			By("Loading and creating the named UICustomHeader fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/uicustomheader/named-ui-custom-header.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			namedHeaderInstance = &redhatcopv1alpha1.UICustomHeader{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, namedHeaderInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: namedHeaderInstance.Name, Namespace: namedHeaderInstance.Namespace}
			created := &redhatcopv1alpha1.UICustomHeader{}

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

			By("Verifying the header exists at the spec.name path in Vault")
			secret, err := vaultClient.Logical().Read("sys/config/ui/headers/X-Named-Test-Header")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data).NotTo(BeNil())

			By("Verifying the metadata.name path was not created in Vault")
			metadataNameSecret, err := vaultClient.Logical().Read("sys/config/ui/headers/test-named-header-metadata")
			Expect(err).To(BeNil())
			Expect(metadataNameSecret).To(BeNil())
		})
	})

	Context("When deleting UICustomHeaders", func() {
		It("Should remove the headers from Vault", func() {

			Expect(simpleHeaderInstance).NotTo(BeNil(), "expected simple UICustomHeader to be created before delete phase")
			Expect(namedHeaderInstance).NotTo(BeNil(), "expected named UICustomHeader to be created before delete phase")

			By("Deleting the simple UICustomHeader CR")
			Expect(k8sIntegrationClient.Delete(ctx, simpleHeaderInstance)).Should(Succeed())

			simpleLookupKey := types.NamespacedName{Name: simpleHeaderInstance.Name, Namespace: simpleHeaderInstance.Namespace}
			By("Waiting for the simple UICustomHeader to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, simpleLookupKey, &redhatcopv1alpha1.UICustomHeader{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the named UICustomHeader CR")
			Expect(k8sIntegrationClient.Delete(ctx, namedHeaderInstance)).Should(Succeed())

			namedLookupKey := types.NamespacedName{Name: namedHeaderInstance.Name, Namespace: namedHeaderInstance.Namespace}
			By("Waiting for the named UICustomHeader to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, namedLookupKey, &redhatcopv1alpha1.UICustomHeader{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the simple header no longer exists in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/config/ui/headers/x-custom-test-header")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying the named header no longer exists in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/config/ui/headers/X-Named-Test-Header")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
