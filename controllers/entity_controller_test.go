//go:build integration
// +build integration

package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Entity controller", func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	Context("When creating an Entity", func() {
		It("Should create an Entity in Vault", func() {

			By("Creating a new Entity")
			entityInstance, err := decoder.GetEntityInstance("../test/identity/01-entity-sample.yaml")
			Expect(err).To(BeNil())
			entityInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, entityInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: entityInstance.Name, Namespace: entityInstance.Namespace}
			created := &redhatcopv1alpha1.Entity{}

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

			By("Deleting the Entity")
			Expect(k8sIntegrationClient.Delete(ctx, created)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				return err != nil
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("When updating an Entity", func() {
		It("Should update the Entity in Vault and reflect updated ObservedGeneration", func() {

			By("Creating a new Entity")
			entityInstance, err := decoder.GetEntityInstance("../test/identity/01-entity-sample.yaml")
			Expect(err).To(BeNil())
			entityInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, entityInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: entityInstance.Name, Namespace: entityInstance.Namespace}
			created := &redhatcopv1alpha1.Entity{}

			By("Waiting for Entity to be reconciled successfully")
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

			By("Verifying initial Vault state")
			initialSecret, err := vaultClient.Logical().Read("identity/entity/name/test-entity")
			Expect(err).To(BeNil())
			Expect(initialSecret).NotTo(BeNil())
			initialMetadata, ok := initialSecret.Data["metadata"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected metadata field to be map[string]interface{}")
			Expect(initialMetadata["team"]).To(Equal("engineering"))
			Expect(initialMetadata["environment"]).To(Equal("test"))
			initialPolicies, ok := initialSecret.Data["policies"].([]interface{})
			Expect(ok).To(BeTrue(), "expected policies field to be []interface{}")
			Expect(initialPolicies).To(ContainElement("default"))
			Expect(initialSecret.Data["disabled"]).To(Equal(false))

			By("Recording initial ObservedGeneration")
			var initialObservedGeneration int64
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())
			for _, condition := range created.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialObservedGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialObservedGeneration).To(BeNumerically(">", 0))

			By("Getting the latest Entity before update")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())

			By("Updating the Entity spec")
			created.Spec.Policies = []string{"default", "kv-reader"}
			if created.Spec.Metadata == nil {
				created.Spec.Metadata = map[string]string{}
			}
			created.Spec.Metadata["owner"] = "integration-test"
			Expect(k8sIntegrationClient.Update(ctx, created)).Should(Succeed())

			By("Waiting for Vault to reflect both updated policies")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/entity/name/test-entity")
				if err != nil || secret == nil {
					return false
				}
				policies, ok := secret.Data["policies"].([]interface{})
				if !ok {
					return false
				}
				hasDefault := false
				hasKvReader := false
				for _, p := range policies {
					if p == "default" {
						hasDefault = true
					}
					if p == "kv-reader" {
						hasKvReader = true
					}
				}
				return hasDefault && hasKvReader
			}, timeout, interval).Should(BeTrue())

			By("Verifying updated metadata in Vault")
			updatedSecret, err := vaultClient.Logical().Read("identity/entity/name/test-entity")
			Expect(err).To(BeNil())
			Expect(updatedSecret).NotTo(BeNil())
			updatedMetadata, ok := updatedSecret.Data["metadata"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected metadata field to be map[string]interface{}")
			Expect(updatedMetadata["owner"]).To(Equal("integration-test"))
			Expect(updatedMetadata["team"]).To(Equal("engineering"))
			Expect(updatedMetadata["environment"]).To(Equal("test"))

			By("Verifying ObservedGeneration increased")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}
				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return condition.ObservedGeneration > initialObservedGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Deleting the Entity")
			Expect(k8sIntegrationClient.Delete(ctx, created)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying Entity is deleted from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/entity/name/test-entity")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

		})
	})
})
