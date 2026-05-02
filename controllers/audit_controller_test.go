//go:build integration
// +build integration

package controllers

import (
	"errors"
	"time"

	vault "github.com/hashicorp/vault/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Audit controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var auditInstance *redhatcopv1alpha1.Audit
	var headerInstance *redhatcopv1alpha1.AuditRequestHeader

	AfterAll(func() {
		if headerInstance != nil {
			k8sIntegrationClient.Delete(ctx, headerInstance) //nolint:errcheck
		}
		if auditInstance != nil {
			k8sIntegrationClient.Delete(ctx, auditInstance) //nolint:errcheck
		}
	})

	Context("When creating an Audit device", func() {
		It("Should enable a file audit device in Vault", func() {

			By("Loading and creating the Audit fixture")
			var err error
			auditInstance, err = decoder.GetAuditInstance("../test/audit/audit.yaml")
			Expect(err).To(BeNil())
			auditInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, auditInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: auditInstance.Name, Namespace: auditInstance.Namespace}
			created := &redhatcopv1alpha1.Audit{}

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

			By("Verifying the audit device exists in Vault via Sys().ListAudit()")
			audits, err := vaultClient.Sys().ListAudit()
			Expect(err).To(BeNil())

			auditDevice, exists := audits["test-audit/"]
			Expect(exists).To(BeTrue(), "expected audit device 'test-audit' to exist (with trailing slash key)")
			Expect(auditDevice.Type).To(Equal("file"))
			Expect(auditDevice.Options["file_path"]).To(Equal("stdout"))
		})
	})

	Context("When creating an AuditRequestHeader", func() {
		It("Should configure the request header in Vault", func() {

			By("Loading and creating the AuditRequestHeader fixture")
			var err error
			headerInstance, err = decoder.GetAuditRequestHeaderInstance("../test/audit/auditrequestheader.yaml")
			Expect(err).To(BeNil())
			headerInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, headerInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: headerInstance.Name, Namespace: headerInstance.Namespace}
			created := &redhatcopv1alpha1.AuditRequestHeader{}

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

			By("Verifying the header config exists in Vault via Logical().Read()")
			secret, err := vaultClient.Logical().Read("sys/config/auditing/request-headers/X-Custom-Test-Header")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data).NotTo(BeNil())

			headerData, ok := secret.Data["X-Custom-Test-Header"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected header data to be a map")
			hmac, ok := headerData["hmac"].(bool)
			Expect(ok).To(BeTrue(), "expected hmac to be a bool")
			Expect(hmac).To(BeTrue())
		})
	})

	Context("When updating an Audit device", func() {
		It("Should update the audit device via disable/re-enable and reflect updated ObservedGeneration", func() {

			Expect(auditInstance).NotTo(BeNil(), "expected audit to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: auditInstance.Name, Namespace: auditInstance.Namespace}
			current := &redhatcopv1alpha1.Audit{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating the audit device description")
			current.Spec.Description = "updated-description"
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated description")
			Eventually(func() bool {
				audits, err := vaultClient.Sys().ListAudit()
				if err != nil {
					return false
				}
				auditDevice, exists := audits["test-audit/"]
				if !exists {
					return false
				}
				return auditDevice.Description == "updated-description"
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.Audit{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				for _, condition := range updated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return condition.ObservedGeneration > initialGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When updating an AuditRequestHeader", func() {
		It("Should update the header config in Vault and reflect updated ObservedGeneration", func() {

			Expect(headerInstance).NotTo(BeNil(), "expected header to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: headerInstance.Name, Namespace: headerInstance.Namespace}
			current := &redhatcopv1alpha1.AuditRequestHeader{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating hmac from true to false")
			current.Spec.HMAC = false
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated hmac value")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/config/auditing/request-headers/X-Custom-Test-Header")
				if err != nil || secret == nil {
					return false
				}
				headerData, ok := secret.Data["X-Custom-Test-Header"].(map[string]interface{})
				if !ok {
					return false
				}
				hmac, ok := headerData["hmac"].(bool)
				if !ok {
					return false
				}
				return !hmac
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.AuditRequestHeader{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				for _, condition := range updated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return condition.ObservedGeneration > initialGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting Audit resources", func() {
		It("Should clean up both resources from Vault", func() {

			Expect(headerInstance).NotTo(BeNil(), "expected header to be created before delete phase")
			Expect(auditInstance).NotTo(BeNil(), "expected audit to be created before delete phase")

			By("Deleting the AuditRequestHeader CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, headerInstance)).Should(Succeed())
			headerLookupKey := types.NamespacedName{Name: headerInstance.Name, Namespace: headerInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, headerLookupKey, &redhatcopv1alpha1.AuditRequestHeader{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the header is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/config/auditing/request-headers/X-Custom-Test-Header")
				if err != nil {
					var responseErr *vault.ResponseError
					return errors.As(err, &responseErr) && (responseErr.StatusCode == 400 || responseErr.StatusCode == 404)
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the Audit CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, auditInstance)).Should(Succeed())
			auditLookupKey := types.NamespacedName{Name: auditInstance.Name, Namespace: auditInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, auditLookupKey, &redhatcopv1alpha1.Audit{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the audit device is removed from Vault")
			Eventually(func() bool {
				audits, err := vaultClient.Sys().ListAudit()
				if err != nil {
					return false
				}
				_, exists := audits["test-audit/"]
				return !exists
			}, timeout, interval).Should(BeTrue())
		})
	})
})
