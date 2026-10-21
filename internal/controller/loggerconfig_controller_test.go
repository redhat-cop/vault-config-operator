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

var _ = Describe("LoggerConfig controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var simpleLoggerInstance *redhatcopv1alpha1.LoggerConfig
	var namedLoggerInstance *redhatcopv1alpha1.LoggerConfig

	AfterAll(func() {
		if simpleLoggerInstance != nil {
			k8sIntegrationClient.Delete(ctx, simpleLoggerInstance) //nolint:errcheck
		}
		if namedLoggerInstance != nil {
			k8sIntegrationClient.Delete(ctx, namedLoggerInstance) //nolint:errcheck
		}
	})

	Context("When creating a simple LoggerConfig", func() {
		It("Should configure the logger level in Vault", func() {

			By("Loading and creating the simple LoggerConfig fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/loggerconfig/simple-logger-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			simpleLoggerInstance = &redhatcopv1alpha1.LoggerConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, simpleLoggerInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: simpleLoggerInstance.Name, Namespace: simpleLoggerInstance.Namespace}
			created := &redhatcopv1alpha1.LoggerConfig{}

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

			By("Verifying the logger level is set in Vault")
			secret, err := vaultClient.Logical().Read("sys/loggers/core")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["core"]).To(Equal("debug"))
		})
	})

	Context("When creating a LoggerConfig with spec.name override", func() {
		It("Should configure the logger at the spec.name path in Vault", func() {

			By("Loading and creating the named LoggerConfig fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/loggerconfig/named-logger-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			namedLoggerInstance = &redhatcopv1alpha1.LoggerConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, namedLoggerInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: namedLoggerInstance.Name, Namespace: namedLoggerInstance.Namespace}
			created := &redhatcopv1alpha1.LoggerConfig{}

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

			By("Verifying the logger level is set at the spec.name path in Vault")
			secret, err := vaultClient.Logical().Read("sys/loggers/audit")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["audit"]).To(Equal("trace"))
		})
	})

	Context("When deleting LoggerConfigs", func() {
		It("Should revert the loggers in Vault", func() {

			Expect(simpleLoggerInstance).NotTo(BeNil(), "expected simple LoggerConfig to be created before delete phase")
			Expect(namedLoggerInstance).NotTo(BeNil(), "expected named LoggerConfig to be created before delete phase")

			By("Deleting the simple LoggerConfig CR")
			Expect(k8sIntegrationClient.Delete(ctx, simpleLoggerInstance)).Should(Succeed())

			simpleLookupKey := types.NamespacedName{Name: simpleLoggerInstance.Name, Namespace: simpleLoggerInstance.Namespace}
			By("Waiting for the simple LoggerConfig to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, simpleLookupKey, &redhatcopv1alpha1.LoggerConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the named LoggerConfig CR")
			Expect(k8sIntegrationClient.Delete(ctx, namedLoggerInstance)).Should(Succeed())

			namedLookupKey := types.NamespacedName{Name: namedLoggerInstance.Name, Namespace: namedLoggerInstance.Namespace}
			By("Waiting for the named LoggerConfig to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, namedLookupKey, &redhatcopv1alpha1.LoggerConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the simple logger reverted in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/loggers/core")
				if err != nil {
					return false
				}
				// After deletion, logger should revert to default or entry disappears
				return secret == nil || secret.Data["core"] != "debug"
			}, timeout, interval).Should(BeTrue())

			By("Verifying the named logger reverted in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/loggers/audit")
				if err != nil {
					return false
				}
				return secret == nil || secret.Data["audit"] != "trace"
			}, timeout, interval).Should(BeTrue())
		})
	})
})
