//go:build integration
// +build integration

package controllers

import (
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("PasswordPolicy controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var simplePPInstance *redhatcopv1alpha1.PasswordPolicy
	var namedPPInstance *redhatcopv1alpha1.PasswordPolicy

	AfterAll(func() {
		if simplePPInstance != nil {
			k8sIntegrationClient.Delete(ctx, simplePPInstance) //nolint:errcheck
		}
		if namedPPInstance != nil {
			k8sIntegrationClient.Delete(ctx, namedPPInstance) //nolint:errcheck
		}
	})

	Context("When creating a simple PasswordPolicy", func() {
		It("Should create the password policy in Vault and generate valid passwords", func() {

			By("Loading and creating the simple PasswordPolicy fixture")
			var err error
			simplePPInstance, err = decoder.GetPasswordPolicyInstance("../test/passwordpolicy/simple-password-policy.yaml")
			Expect(err).To(BeNil())
			simplePPInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, simplePPInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: simplePPInstance.Name, Namespace: simplePPInstance.Namespace}
			created := &redhatcopv1alpha1.PasswordPolicy{}

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

			By("Verifying the password policy exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/policies/password/test-simple-password-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["policy"]).To(ContainSubstring("abcdefghijklmnopqrstuvwxyz"))

			By("Verifying password generation matches the policy constraints")
			genSecret, err := vaultClient.Logical().Read("sys/policies/password/test-simple-password-policy/generate")
			Expect(err).To(BeNil())
			Expect(genSecret).NotTo(BeNil())
			passwordValue, ok := genSecret.Data["password"]
			Expect(ok).To(BeTrue(), "expected generated password response to include password")
			password, ok := passwordValue.(string)
			Expect(ok).To(BeTrue(), "expected generated password to be a string")
			Expect(password).To(MatchRegexp("^[a-z]{20}$"))
		})
	})

	Context("When creating a PasswordPolicy with spec.name override", func() {
		It("Should create the policy at the spec.name path in Vault", func() {

			By("Loading and creating the named PasswordPolicy fixture")
			var err error
			namedPPInstance, err = decoder.GetPasswordPolicyInstance("../test/passwordpolicy/named-password-policy.yaml")
			Expect(err).To(BeNil())
			namedPPInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, namedPPInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: namedPPInstance.Name, Namespace: namedPPInstance.Namespace}
			created := &redhatcopv1alpha1.PasswordPolicy{}

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

			By("Verifying the policy exists at the spec.name path in Vault")
			secret, err := vaultClient.Logical().Read("sys/policies/password/test-named-password-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["policy"]).To(ContainSubstring("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))

			By("Verifying the metadata.name path was not created in Vault")
			metadataNameSecret, err := vaultClient.Logical().Read("sys/policies/password/test-named-pp-metadata")
			Expect(err).To(BeNil())
			Expect(metadataNameSecret).To(BeNil())

			By("Verifying password generation matches the named policy constraints")
			genSecret, err := vaultClient.Logical().Read("sys/policies/password/test-named-password-policy/generate")
			Expect(err).To(BeNil())
			Expect(genSecret).NotTo(BeNil())
			passwordValue, ok := genSecret.Data["password"]
			Expect(ok).To(BeTrue(), "expected generated password response to include password")
			password, ok := passwordValue.(string)
			Expect(ok).To(BeTrue(), "expected generated password to be a string")
			matched, err := regexp.MatchString("^[A-Z0-9]{10}$", password)
			Expect(err).To(BeNil())
			Expect(matched).To(BeTrue(), "expected password to match ^[A-Z0-9]{10}$, got: %s", password)
		})
	})

	Context("When deleting PasswordPolicies", func() {
		It("Should remove password policies from Vault", func() {

			Expect(simplePPInstance).NotTo(BeNil(), "expected simple PasswordPolicy to be created before delete phase")
			Expect(namedPPInstance).NotTo(BeNil(), "expected named PasswordPolicy to be created before delete phase")

			By("Deleting the simple PasswordPolicy CR")
			Expect(k8sIntegrationClient.Delete(ctx, simplePPInstance)).Should(Succeed())

			simpleLookupKey := types.NamespacedName{Name: simplePPInstance.Name, Namespace: simplePPInstance.Namespace}
			By("Waiting for the simple PasswordPolicy to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, simpleLookupKey, &redhatcopv1alpha1.PasswordPolicy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the named PasswordPolicy CR")
			Expect(k8sIntegrationClient.Delete(ctx, namedPPInstance)).Should(Succeed())

			namedLookupKey := types.NamespacedName{Name: namedPPInstance.Name, Namespace: namedPPInstance.Namespace}
			By("Waiting for the named PasswordPolicy to be removed from K8s")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, namedLookupKey, &redhatcopv1alpha1.PasswordPolicy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the simple password policy no longer exists in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/policies/password/test-simple-password-policy")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying the named password policy no longer exists in Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/policies/password/test-named-password-policy")
				if err != nil {
					return false
				}
				return secret == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
})
