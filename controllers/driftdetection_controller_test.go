//go:build integration
// +build integration

package controllers

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/controllertestutils"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func enableDriftDetection(syncPeriod time.Duration) func() {
	origSyncPeriod := vaultresourcecontroller.SyncPeriod
	origDriftEnv, origDriftSet := os.LookupEnv("ENABLE_DRIFT_DETECTION")

	os.Setenv("ENABLE_DRIFT_DETECTION", "true")
	vaultresourcecontroller.SetSyncPeriod(syncPeriod)

	return func() {
		vaultresourcecontroller.SetSyncPeriod(origSyncPeriod)
		if origDriftSet {
			os.Setenv("ENABLE_DRIFT_DETECTION", origDriftEnv)
		} else {
			os.Unsetenv("ENABLE_DRIFT_DETECTION")
		}
	}
}

func triggerNonSpecUpdate(triggerCtx context.Context, c client.Client, obj client.Object) error {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["drift-detection-trigger"] = time.Now().Format(time.RFC3339Nano)
	obj.SetAnnotations(annotations)
	return c.Update(triggerCtx, obj)
}

func getReconcileSuccessfulTime(obj *redhatcopv1alpha1.Policy) *metav1.Time {
	for _, c := range obj.Status.Conditions {
		if c.Type == vaultresourcecontroller.ReconcileSuccessful && c.Status == metav1.ConditionTrue {
			return &c.LastTransitionTime
		}
	}
	return nil
}

var _ = Describe("Drift detection", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	Context("Policy drift detection with drift detection enabled", Ordered, func() {
		var policyInstance *redhatcopv1alpha1.Policy
		var cleanupDrift func()
		var originalRules string

		BeforeAll(func() {
			cleanupDrift = enableDriftDetection(5 * time.Second)

			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/drift-detection/policy-drift-test.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			policyInstance = &redhatcopv1alpha1.Policy{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, policyInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: policyInstance.Name, Namespace: policyInstance.Namespace}
			created := &redhatcopv1alpha1.Policy{}

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

			secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			originalRules = secret.Data["rules"].(string)
			Expect(originalRules).To(ContainSubstring("secret/data/drift-test/*"))
		})

		AfterAll(func() {
			if policyInstance != nil {
				k8sIntegrationClient.Delete(ctx, policyInstance) //nolint:errcheck
				Eventually(func() bool {
					err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: policyInstance.Name, Namespace: policyInstance.Namespace}, &redhatcopv1alpha1.Policy{})
					return apierrors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())
			}
			if cleanupDrift != nil {
				cleanupDrift()
			}
		})

		It("Should correct drift when Vault policy is manually modified", func() {
			By("Overwriting the policy in Vault directly")
			driftedRules := `path "drifted/*" { capabilities = ["read"] }`
			_, err := vaultClient.Logical().Write("sys/policy/test-drift-policy", map[string]interface{}{
				"policy": driftedRules,
			})
			Expect(err).To(BeNil())

			By("Verifying the drift is present in Vault")
			secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["rules"].(string)).To(ContainSubstring("drifted"))

			By("Waiting for the SyncPeriod to elapse")
			time.Sleep(6 * time.Second)

			By("Triggering a non-spec annotation update on the CR")
			lookupKey := types.NamespacedName{Name: policyInstance.Name, Namespace: policyInstance.Namespace}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, policyInstance)).Should(Succeed())
			Expect(triggerNonSpecUpdate(ctx, k8sIntegrationClient, policyInstance)).Should(Succeed())

			By("Verifying the operator corrected the drift back to the original rules")
			Eventually(func() string {
				secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy")
				if err != nil || secret == nil {
					return ""
				}
				rules, ok := secret.Data["rules"].(string)
				if !ok {
					return ""
				}
				return rules
			}, 60*time.Second, 2*time.Second).Should(Equal(originalRules))
		})

		It("Should not write when no drift exists (no false positive)", func() {
			By("Getting the current ReconcileSuccessful LastTransitionTime")
			lookupKey := types.NamespacedName{Name: policyInstance.Name, Namespace: policyInstance.Namespace}
			current := &redhatcopv1alpha1.Policy{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			initialTime := getReconcileSuccessfulTime(current)
			Expect(initialTime).NotTo(BeNil(), "expected ReconcileSuccessful condition to exist")

			By("Waiting for the SyncPeriod to elapse")
			time.Sleep(6 * time.Second)

			By("Triggering another non-spec annotation update")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, policyInstance)).Should(Succeed())
			Expect(triggerNonSpecUpdate(ctx, k8sIntegrationClient, policyInstance)).Should(Succeed())

			By("Waiting for the reconcile to complete (LastTransitionTime advances)")
			Eventually(func() bool {
				updated := &redhatcopv1alpha1.Policy{}
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				newTime := getReconcileSuccessfulTime(updated)
				if newTime == nil {
					return false
				}
				return newTime.After(initialTime.Time)
			}, 60*time.Second, 2*time.Second).Should(BeTrue())

			By("Verifying Vault policy still has the exact original rules (no unnecessary write)")
			secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["rules"].(string)).To(Equal(originalRules))
		})
	})

	Context("SecretEngineMount tune config drift detection", Ordered, func() {
		var mountInstance *redhatcopv1alpha1.SecretEngineMount
		var cleanupDrift func()

		BeforeAll(func() {
			cleanupDrift = enableDriftDetection(5 * time.Second)

			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/drift-detection/secretenginemount-drift-test.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			mountInstance = &redhatcopv1alpha1.SecretEngineMount{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, mountInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			created := &redhatcopv1alpha1.SecretEngineMount{}

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
		})

		AfterAll(func() {
			if mountInstance != nil {
				k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
				Eventually(func() bool {
					err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}, &redhatcopv1alpha1.SecretEngineMount{})
					return apierrors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())
			}
			if cleanupDrift != nil {
				cleanupDrift()
			}
		})

		It("Should correct drift when tune config is manually modified", func() {
			By("Reading current tune config from Vault")
			tuneSecret, err := vaultClient.Logical().Read("sys/mounts/drift-test-kv/tune")
			Expect(err).To(BeNil())
			Expect(tuneSecret).NotTo(BeNil())

			By("Modifying the tune config directly in Vault")
			_, err = vaultClient.Logical().Write("sys/mounts/drift-test-kv/tune", map[string]interface{}{
				"default_lease_ttl": "7200",
			})
			Expect(err).To(BeNil())

			By("Verifying the drift is present")
			driftedTune, err := vaultClient.Logical().Read("sys/mounts/drift-test-kv/tune")
			Expect(err).To(BeNil())
			Expect(driftedTune).NotTo(BeNil())

			By("Waiting for the SyncPeriod to elapse")
			time.Sleep(6 * time.Second)

			By("Triggering a non-spec annotation update on the CR")
			lookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, mountInstance)).Should(Succeed())
			Expect(triggerNonSpecUpdate(ctx, k8sIntegrationClient, mountInstance)).Should(Succeed())

			By("Verifying the operator corrected the tune config drift")
			Eventually(func() bool {
				tuneSecret, err := vaultClient.Logical().Read("sys/mounts/drift-test-kv/tune")
				if err != nil || tuneSecret == nil {
					return false
				}
				defaultLeaseTTL := tuneSecret.Data["default_lease_ttl"]
				if defaultLeaseTTL == nil {
					return false
				}
				// CR specifies "1h" = 3600 seconds; Vault returns it as a json.Number
				ttlStr, ok := defaultLeaseTTL.(interface{ String() string })
				if ok {
					return ttlStr.String() == "3600"
				}
				// Fallback: check numeric comparison
				switch v := defaultLeaseTTL.(type) {
				case float64:
					return v == 3600
				default:
					return false
				}
			}, 60*time.Second, 2*time.Second).Should(BeTrue())
		})
	})

	Context("Drift detection disabled — drift persists", Ordered, func() {
		var policyInstance *redhatcopv1alpha1.Policy

		BeforeAll(func() {
			os.Unsetenv("ENABLE_DRIFT_DETECTION")
			vaultresourcecontroller.SetSyncPeriod(36000 * time.Second)

			var err error
			policyInstance, err = controllertestutils.DecodeInstance[*redhatcopv1alpha1.Policy]("../test/drift-detection/policy-drift-test.yaml")
			Expect(err).To(BeNil())
			policyInstance.Name = "test-drift-policy-disabled"
			policyInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, policyInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: policyInstance.Name, Namespace: policyInstance.Namespace}
			created := &redhatcopv1alpha1.Policy{}

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
		})

		AfterAll(func() {
			if policyInstance != nil {
				k8sIntegrationClient.Delete(ctx, policyInstance) //nolint:errcheck
				Eventually(func() bool {
					err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: policyInstance.Name, Namespace: policyInstance.Namespace}, &redhatcopv1alpha1.Policy{})
					return apierrors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())
			}
		})

		It("Should NOT correct drift when drift detection is disabled", func() {
			By("Recording the current ReconcileSuccessful LastTransitionTime")
			lookupKey := types.NamespacedName{Name: policyInstance.Name, Namespace: policyInstance.Namespace}
			baseline := &redhatcopv1alpha1.Policy{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, baseline)).Should(Succeed())
			baselineTime := getReconcileSuccessfulTime(baseline)
			Expect(baselineTime).NotTo(BeNil(), "expected ReconcileSuccessful condition to exist")

			By("Overwriting the policy in Vault directly")
			driftedRules := `path "drifted/*" { capabilities = ["read"] }`
			_, err := vaultClient.Logical().Write("sys/policy/test-drift-policy-disabled", map[string]interface{}{
				"policy": driftedRules,
			})
			Expect(err).To(BeNil())

			By("Triggering an annotation update on the CR")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, policyInstance)).Should(Succeed())
			Expect(triggerNonSpecUpdate(ctx, k8sIntegrationClient, policyInstance)).Should(Succeed())

			By("Verifying the drift persists and no reconcile ran")
			Consistently(func() string {
				secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy-disabled")
				if err != nil || secret == nil {
					return ""
				}
				rules, ok := secret.Data["rules"].(string)
				if !ok {
					return ""
				}
				return rules
			}, 10*time.Second, 2*time.Second).Should(ContainSubstring("drifted"))

			By("Verifying ReconcileSuccessful LastTransitionTime did NOT advance (predicate rejected the event)")
			afterUpdate := &redhatcopv1alpha1.Policy{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, afterUpdate)).Should(Succeed())
			afterTime := getReconcileSuccessfulTime(afterUpdate)
			Expect(afterTime).NotTo(BeNil())
			Expect(afterTime.Time.Equal(baselineTime.Time)).To(BeTrue(),
				"ReconcileSuccessful.LastTransitionTime should not have advanced — predicate should have rejected the update event")
		})
	})
})
