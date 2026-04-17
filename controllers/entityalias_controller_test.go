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

var _ = Describe("EntityAlias controller", func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	Context("When creating an EntityAlias", func() {
		It("Should create an EntityAlias in Vault", func() {

			By("Creating a new Entity first")
			entityInstance, err := decoder.GetEntityInstance("../test/identity/01-entity-sample.yaml")
			Expect(err).To(BeNil())
			entityInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, entityInstance)).Should(Succeed())

			entityLookupKey := types.NamespacedName{Name: entityInstance.Name, Namespace: entityInstance.Namespace}
			entityCreated := &redhatcopv1alpha1.Entity{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, entityLookupKey, entityCreated)
				if err != nil {
					return false
				}

				for _, condition := range entityCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("Creating a new EntityAlias")
			entityAliasInstance, err := decoder.GetEntityAliasInstance("../test/identity/02-entityalias-sample.yaml")
			Expect(err).To(BeNil())
			entityAliasInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, entityAliasInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: entityAliasInstance.Name, Namespace: entityAliasInstance.Namespace}
			created := &redhatcopv1alpha1.EntityAlias{}

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

			By("Verifying that the EntityAlias has an ID")
			Expect(created.Status.ID).NotTo(BeEmpty())

			By("Deleting the EntityAlias")
			Expect(k8sIntegrationClient.Delete(ctx, created)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the Entity")
			Expect(k8sIntegrationClient.Delete(ctx, entityCreated)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, entityLookupKey, entityCreated)
				return err != nil
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("When updating an EntityAlias", func() {
		It("Should update the EntityAlias in Vault and reflect updated ObservedGeneration", func() {

			By("Creating a new Entity first")
			entityInstance, err := decoder.GetEntityInstance("../test/identity/01-entity-sample.yaml")
			Expect(err).To(BeNil())
			entityInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, entityInstance)).Should(Succeed())

			entityLookupKey := types.NamespacedName{Name: entityInstance.Name, Namespace: entityInstance.Namespace}
			entityCreated := &redhatcopv1alpha1.Entity{}

			By("Waiting for Entity to be reconciled successfully")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, entityLookupKey, entityCreated)
				if err != nil {
					return false
				}
				for _, condition := range entityCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Creating a new EntityAlias")
			entityAliasInstance, err := decoder.GetEntityAliasInstance("../test/identity/02-entityalias-sample.yaml")
			Expect(err).To(BeNil())
			entityAliasInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, entityAliasInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: entityAliasInstance.Name, Namespace: entityAliasInstance.Namespace}
			created := &redhatcopv1alpha1.EntityAlias{}

			By("Waiting for EntityAlias to be reconciled successfully")
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

			By("Verifying EntityAlias has a Status.ID")
			Expect(created.Status.ID).NotTo(BeEmpty())
			capturedAliasID := created.Status.ID

			By("Verifying initial custom_metadata in Vault")
			initialSecret, err := vaultClient.Logical().Read("identity/entity-alias/id/" + capturedAliasID)
			Expect(err).To(BeNil())
			Expect(initialSecret).NotTo(BeNil())
			initialCustomMetadata, ok := initialSecret.Data["custom_metadata"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected custom_metadata field to be map[string]interface{}")
			Expect(initialCustomMetadata["contact"]).To(Equal("admin@example.com"))

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

			By("Getting the latest EntityAlias before update")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())

			By("Updating the EntityAlias customMetadata")
			created.Spec.CustomMetadata = map[string]string{
				"contact": "admin@example.com",
				"purpose": "integration-test",
			}
			Expect(k8sIntegrationClient.Update(ctx, created)).Should(Succeed())

			By("Waiting for Vault to reflect the updated custom_metadata")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/entity-alias/id/" + capturedAliasID)
				if err != nil || secret == nil {
					return false
				}
				customMetadata, ok := secret.Data["custom_metadata"].(map[string]interface{})
				if !ok {
					return false
				}
				purpose, ok := customMetadata["purpose"]
				if !ok {
					return false
				}
				return purpose == "integration-test"
			}, timeout, interval).Should(BeTrue())

			By("Verifying full custom_metadata in Vault after update")
			finalSecret, err := vaultClient.Logical().Read("identity/entity-alias/id/" + capturedAliasID)
			Expect(err).To(BeNil())
			Expect(finalSecret).NotTo(BeNil())
			finalCustomMetadata, ok := finalSecret.Data["custom_metadata"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected custom_metadata field to be map[string]interface{}")
			Expect(finalCustomMetadata["contact"]).To(Equal("admin@example.com"))
			Expect(finalCustomMetadata["purpose"]).To(Equal("integration-test"))

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

			By("Deleting the EntityAlias")
			Expect(k8sIntegrationClient.Delete(ctx, created)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the Entity")
			Expect(k8sIntegrationClient.Delete(ctx, entityCreated)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, entityLookupKey, entityCreated)
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying EntityAlias is deleted from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/entity-alias/id/" + capturedAliasID)
				if err != nil {
					return false
				}
				return secret == nil
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
