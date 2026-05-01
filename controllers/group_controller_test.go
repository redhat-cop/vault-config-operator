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

var _ = Describe("Group and GroupAlias controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var groupInstance *redhatcopv1alpha1.Group
	var aliasInstance *redhatcopv1alpha1.GroupAlias
	var capturedAliasID string

	AfterAll(func() {
		if aliasInstance != nil {
			k8sIntegrationClient.Delete(ctx, aliasInstance) //nolint:errcheck
		}
		if groupInstance != nil {
			k8sIntegrationClient.Delete(ctx, groupInstance) //nolint:errcheck
		}
	})

	Context("When creating a Group", func() {
		It("Should create the group in Vault with correct settings", func() {

			By("Loading and creating the Group fixture")
			var err error
			groupInstance, err = decoder.GetGroupInstance("../test/groups/test-group.yaml")
			Expect(err).To(BeNil())
			groupInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, groupInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: groupInstance.Name, Namespace: groupInstance.Namespace}
			created := &redhatcopv1alpha1.Group{}

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

			By("Verifying the group exists in Vault with correct settings")
			secret, err := vaultClient.Logical().Read("identity/group/name/test-group")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			groupType, ok := secret.Data["type"].(string)
			Expect(ok).To(BeTrue(), "expected type to be a string")
			Expect(groupType).To(Equal("external"))

			metadata, ok := secret.Data["metadata"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected metadata to be map[string]interface{}")
			Expect(metadata["team"]).To(Equal("team-abc"))

			policies, ok := secret.Data["policies"].([]interface{})
			Expect(ok).To(BeTrue(), "expected policies to be []interface{}")
			Expect(policies).To(ConsistOf("team-abc-access"))
		})
	})

	Context("When updating a Group", func() {
		It("Should update the group in Vault and reflect updated ObservedGeneration", func() {

			Expect(groupInstance).NotTo(BeNil(), "expected group to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: groupInstance.Name, Namespace: groupInstance.Namespace}
			current := &redhatcopv1alpha1.Group{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating the Group policies")
			current.Spec.Policies = append(current.Spec.Policies, "kv-reader")
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated policies")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/group/name/test-group")
				if err != nil || secret == nil {
					return false
				}
				policies, ok := secret.Data["policies"].([]interface{})
				if !ok {
					return false
				}
				for _, p := range policies {
					ps, ok := p.(string)
					if ok && ps == "kv-reader" {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Verifying the full policy set after update")
			updatedSecret, err := vaultClient.Logical().Read("identity/group/name/test-group")
			Expect(err).To(BeNil())
			Expect(updatedSecret).NotTo(BeNil())
			updatedPolicies, ok := updatedSecret.Data["policies"].([]interface{})
			Expect(ok).To(BeTrue(), "expected policies to be []interface{}")
			Expect(updatedPolicies).To(ConsistOf("team-abc-access", "kv-reader"))

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.Group{}
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

	Context("When creating a GroupAlias", func() {
		It("Should create the alias in Vault with Status.ID populated", func() {

			Expect(groupInstance).NotTo(BeNil(), "expected group to be created before alias phase")

			By("Loading and creating the GroupAlias fixture")
			var err error
			aliasInstance, err = decoder.GetGroupAliasInstance("../test/groups/test-groupalias.yaml")
			Expect(err).To(BeNil())
			aliasInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, aliasInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: aliasInstance.Name, Namespace: aliasInstance.Namespace}
			created := &redhatcopv1alpha1.GroupAlias{}

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

			By("Verifying Status.ID is populated")
			Expect(created.Status.ID).NotTo(BeEmpty())
			capturedAliasID = created.Status.ID

			By("Verifying the alias exists in Vault")
			secret, err := vaultClient.Logical().Read("identity/group-alias/id/" + capturedAliasID)
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			aliasName, ok := secret.Data["name"].(string)
			Expect(ok).To(BeTrue(), "expected name to be a string")
			Expect(aliasName).To(Equal("test-groupalias"))

			mountAccessor, ok := secret.Data["mount_accessor"].(string)
			Expect(ok).To(BeTrue(), "expected mount_accessor to be a string")
			Expect(mountAccessor).NotTo(BeEmpty())

			By("Verifying mount_accessor matches the kubernetes auth mount")
			authMounts, err := vaultClient.Logical().Read("sys/auth")
			Expect(err).To(BeNil())
			Expect(authMounts).NotTo(BeNil())
			kubeMount, ok := authMounts.Data["kubernetes/"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected kubernetes/ auth mount to exist")
			expectedAccessor, ok := kubeMount["accessor"].(string)
			Expect(ok).To(BeTrue(), "expected accessor to be a string")
			Expect(mountAccessor).To(Equal(expectedAccessor))

			By("Verifying canonical_id matches the group's ID")
			groupSecret, err := vaultClient.Logical().Read("identity/group/name/test-group")
			Expect(err).To(BeNil())
			Expect(groupSecret).NotTo(BeNil())
			groupID, ok := groupSecret.Data["id"].(string)
			Expect(ok).To(BeTrue(), "expected group id to be a string")

			canonicalID, ok := secret.Data["canonical_id"].(string)
			Expect(ok).To(BeTrue(), "expected canonical_id to be a string")
			Expect(canonicalID).To(Equal(groupID))
		})
	})

	Context("When deleting GroupAlias and Group resources", func() {
		It("Should clean up alias and group from Vault and remove all K8s resources", func() {

			Expect(aliasInstance).NotTo(BeNil(), "expected alias to be created before delete phase")
			Expect(groupInstance).NotTo(BeNil(), "expected group to be created before delete phase")
			Expect(capturedAliasID).NotTo(BeEmpty(), "expected alias ID to be captured before delete phase")

			By("Deleting the GroupAlias CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, aliasInstance)).Should(Succeed())
			aliasLookupKey := types.NamespacedName{Name: aliasInstance.Name, Namespace: aliasInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, aliasLookupKey, &redhatcopv1alpha1.GroupAlias{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the alias is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/group-alias/id/" + capturedAliasID)
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the Group CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, groupInstance)).Should(Succeed())
			groupLookupKey := types.NamespacedName{Name: groupInstance.Name, Namespace: groupInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, groupLookupKey, &redhatcopv1alpha1.Group{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the group is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("identity/group/name/test-group")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
