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

var _ = Describe("Policy controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var simplePolicyInstance *redhatcopv1alpha1.Policy
	var aclPolicyInstance *redhatcopv1alpha1.Policy

	AfterAll(func() {
		if simplePolicyInstance != nil {
			k8sIntegrationClient.Delete(ctx, simplePolicyInstance) //nolint:errcheck
		}
		if aclPolicyInstance != nil {
			k8sIntegrationClient.Delete(ctx, aclPolicyInstance) //nolint:errcheck
		}
	})

	Context("When creating a simple Policy", func() {
		It("Should create the policy in Vault at sys/policy/<name>", func() {

			By("Loading and creating the simple Policy fixture")
			var err error
			simplePolicyInstance, err = decoder.GetPolicyInstance("../test/policy/simple-policy.yaml")
			Expect(err).To(BeNil())
			simplePolicyInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, simplePolicyInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: simplePolicyInstance.Name, Namespace: simplePolicyInstance.Namespace}
			created := &redhatcopv1alpha1.Policy{}

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

			By("Verifying the policy exists in Vault via sys/policy/<name>")
			secret, err := vaultClient.Logical().Read("sys/policy/test-simple-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["rules"]).To(ContainSubstring("secret/data/test/*"))
			Expect(secret.Data["rules"]).To(ContainSubstring(`capabilities = ["create", "read", "update", "delete", "list"]`))
		})
	})

	Context("When creating a Policy with accessor placeholder and type: acl", func() {
		It("Should resolve accessor and create at sys/policies/acl/<name>", func() {

			By("Loading and creating the ACL Policy fixture with accessor placeholder")
			var err error
			aclPolicyInstance, err = decoder.GetPolicyInstance("../test/policy/acl-policy-with-accessor.yaml")
			Expect(err).To(BeNil())
			aclPolicyInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, aclPolicyInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: aclPolicyInstance.Name, Namespace: aclPolicyInstance.Namespace}
			created := &redhatcopv1alpha1.Policy{}

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

			By("Verifying the policy exists in Vault via sys/policies/acl/<name>")
			secret, err := vaultClient.Logical().Read("sys/policies/acl/test-acl-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			By("Verifying accessor placeholder was resolved")
			policyText, ok := secret.Data["policy"].(string)
			Expect(ok).To(BeTrue(), "expected secret.Data[\"policy\"] to be a string")
			Expect(policyText).NotTo(ContainSubstring("${auth/kubernetes/@accessor}"))
			Expect(policyText).To(ContainSubstring("auth_kubernetes_"))
		})
	})

	Context("When deleting Policies", func() {
		It("Should remove policies from Vault", func() {

			By("Deleting the simple Policy CR")
			Expect(k8sIntegrationClient.Delete(ctx, simplePolicyInstance)).Should(Succeed())

			simpleLookupKey := types.NamespacedName{Name: simplePolicyInstance.Name, Namespace: simplePolicyInstance.Namespace}
			By("Waiting for the simple Policy to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, simpleLookupKey, &redhatcopv1alpha1.Policy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the ACL Policy CR")
			Expect(k8sIntegrationClient.Delete(ctx, aclPolicyInstance)).Should(Succeed())

			aclLookupKey := types.NamespacedName{Name: aclPolicyInstance.Name, Namespace: aclPolicyInstance.Namespace}
			By("Waiting for the ACL Policy to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, aclLookupKey, &redhatcopv1alpha1.Policy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the simple policy no longer exists in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/policy/test-simple-policy")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying the ACL policy no longer exists in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/policies/acl/test-acl-policy")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
