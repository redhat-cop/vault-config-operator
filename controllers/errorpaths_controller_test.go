//go:build integration
// +build integration

package controllers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Error path handling", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	hasReconcileFailedCondition := func(conditions []metav1.Condition) (bool, string) {
		for _, condition := range conditions {
			if condition.Type == vaultresourcecontroller.ReconcileFailed &&
				condition.Status == metav1.ConditionFalse {
				return true, condition.Message
			}
		}
		return false, ""
	}

	hasReconcileSuccessful := func(conditions []metav1.Condition) bool {
		for _, condition := range conditions {
			if condition.Type == vaultresourcecontroller.ReconcileSuccessful &&
				condition.Status == metav1.ConditionTrue {
				return true
			}
		}
		return false
	}

	Context("Invalid ServiceAccount", func() {
		It("Should set ReconcileFailed condition and allow clean deletion", func() {

			By("Loading the Policy fixture with non-existent ServiceAccount")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/error-paths/policy-invalid-serviceaccount.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			instance := &redhatcopv1alpha1.Policy{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, instance)).Should(Succeed())

			By("Waiting for ReconcileFailed condition")
			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			fetched := &redhatcopv1alpha1.Policy{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, fetched)
				if err != nil {
					return false
				}
				hasFailed, _ := hasReconcileFailedCondition(fetched.GetConditions())
				return hasFailed
			}, timeout, interval).Should(BeTrue())

			By("Verifying the error message references the ServiceAccount or token failure")
			_, msg := hasReconcileFailedCondition(fetched.GetConditions())
			Expect(msg).To(ContainSubstring("nonexistent-sa-xyz"))

			By("Verifying a ProcessingError warning event was emitted")
			Eventually(func() bool {
				eventList := &corev1.EventList{}
				err := k8sIntegrationClient.List(ctx, eventList, client.InNamespace(vaultAdminNamespaceName))
				if err != nil {
					return false
				}
				for _, event := range eventList.Items {
					if event.InvolvedObject.Name == instance.Name &&
						event.Reason == "ProcessingError" &&
						event.Type == "Warning" {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Verifying the CR never reached ReconcileSuccessful")
			Expect(hasReconcileSuccessful(fetched.GetConditions())).To(BeFalse())

			By("Deleting the CR and verifying no finalizer deadlock")
			Expect(k8sIntegrationClient.Delete(ctx, instance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.Policy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Invalid Vault auth role", func() {
		It("Should set ReconcileFailed condition and allow clean deletion", func() {

			By("Loading the Policy fixture with non-existent Vault role")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/error-paths/policy-invalid-role.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			instance := &redhatcopv1alpha1.Policy{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			fetched := &redhatcopv1alpha1.Policy{}

			By("Waiting for ReconcileFailed condition")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, fetched)
				if err != nil {
					return false
				}
				hasFailed, _ := hasReconcileFailedCondition(fetched.GetConditions())
				return hasFailed
			}, timeout, interval).Should(BeTrue())

			By("Verifying the error message references a login failure")
			_, msg := hasReconcileFailedCondition(fetched.GetConditions())
			Expect(msg).To(SatisfyAny(
				ContainSubstring("nonexistent-role-xyz"),
				ContainSubstring("login"),
				ContainSubstring("role"),
			))

			By("Verifying the CR never reached ReconcileSuccessful")
			Expect(hasReconcileSuccessful(fetched.GetConditions())).To(BeFalse())

			By("Deleting the CR and verifying no finalizer deadlock")
			Expect(k8sIntegrationClient.Delete(ctx, instance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.Policy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("Invalid Vault write path", func() {
		var dummySecret *corev1.Secret

		It("Should set ReconcileFailed condition and allow clean deletion", func() {

			By("Creating a dummy K8s secret for rootCredentials")
			dummySecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dummy-error-path-creds",
					Namespace: vaultAdminNamespaceName,
				},
				Data: map[string][]byte{
					"username": []byte("fakeuser"),
					"password": []byte("fakepass"),
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, dummySecret)).Should(Succeed())

			By("Loading the DatabaseSecretEngineConfig fixture with non-existent mount")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/error-paths/databasesecretengineconfig-invalid-mount.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			instance := &redhatcopv1alpha1.DatabaseSecretEngineConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			fetched := &redhatcopv1alpha1.DatabaseSecretEngineConfig{}

			By("Waiting for ReconcileFailed condition")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, fetched)
				if err != nil {
					return false
				}
				hasFailed, _ := hasReconcileFailedCondition(fetched.GetConditions())
				return hasFailed
			}, timeout, interval).Should(BeTrue())

			By("Verifying the error message references a Vault write failure")
			_, msg := hasReconcileFailedCondition(fetched.GetConditions())
			Expect(msg).To(SatisfyAny(
				ContainSubstring("nonexistent-db-mount-xyz"),
				ContainSubstring("no handler"),
				ContainSubstring("unsupported path"),
			))

			By("Verifying the CR never reached ReconcileSuccessful")
			Expect(hasReconcileSuccessful(fetched.GetConditions())).To(BeFalse())

			By("Deleting the CR and verifying no finalizer deadlock")
			Expect(k8sIntegrationClient.Delete(ctx, instance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.DatabaseSecretEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up the dummy secret")
			Expect(k8sIntegrationClient.Delete(ctx, dummySecret)).Should(Succeed())
		})
	})
})
