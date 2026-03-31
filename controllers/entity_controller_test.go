//go:build integration
// +build integration

package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

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
})
