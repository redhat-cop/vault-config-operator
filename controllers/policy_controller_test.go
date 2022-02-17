//go:build integrationz
// +build integrationz

package controllers

import (
	"context"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"

	"k8s.io/apimachinery/pkg/types"
)

//TODO: Example: https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/cronjob-tutorial/testdata/project/controllers/cronjob_controller_test.go
// Define utility constants for object names and testing timeouts/durations and intervals.

var _ = Describe("Policy controller", func() {

	timeout := time.Second * 10
	interval := time.Millisecond * 250

	Context("When creating the kv-engine-admin Policy", func() {
		It("Should be Successful when created", func() {
			By("By creating a new Policy")
			ctx := context.Background()

			instance, err := decoder.GetPolicyInstance("../test/kv-engine-admin-policy.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName

			//SUBSTITUE
			instance.Spec.Policy = strings.Replace(instance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}

				for _, condition := range created.Status.Conditions {
					if condition.Type == "ReconcileSuccess" {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("When creating the secret-writer Policy", func() {
		It("Should be Successful when created", func() {
			By("By creating a new Policy")
			ctx := context.Background()

			instance, err := decoder.GetPolicyInstance("../test/secret-writer-policy.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName

			//SUBSTITUE
			instance.Spec.Policy = strings.Replace(instance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}

				for _, condition := range created.Status.Conditions {
					if condition.Type == "ReconcileSuccess" {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("When creating the secret-reader Policy", func() {
		It("Should be Successful when created", func() {
			By("By creating a new Policy")
			ctx := context.Background()

			instance, err := decoder.GetPolicyInstance("../test/vaultsecret/policy-secret-reader.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultAdminNamespaceName

			//SUBSTITUE
			instance.Spec.Policy = strings.Replace(instance.Spec.Policy, "${accessor}", os.Getenv("ACCESSOR"), -1)

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}

				for _, condition := range created.Status.Conditions {
					if condition.Type == "ReconcileSuccess" {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

		})
	})

})
