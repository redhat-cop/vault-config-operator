//go:build integration
// +build integration

package controllers

import (
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
			_, err := vaultClient.Logical().Write("sys/policy/test-drift-policy", map[string]any{
				"policy": driftedRules,
			})
			Expect(err).To(BeNil())

			By("Verifying the drift is present in Vault")
			secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["rules"].(string)).To(ContainSubstring("drifted"))

			By("Waiting for RequeueAfter to fire and correct the drift automatically")
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
			}, 30*time.Second, 2*time.Second).Should(Equal(originalRules))
		})

		It("Should not write when no drift exists (no false positive)", func() {
			By("Verifying Vault currently has the correct rules (baseline)")
			secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["rules"].(string)).To(Equal(originalRules))

			By("Waiting for at least one RequeueAfter cycle to fire")
			time.Sleep(7 * time.Second)

			By("Verifying Vault policy remains unchanged over an observation window (no false positive writes)")
			Consistently(func() string {
				secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy")
				if err != nil || secret == nil {
					return ""
				}
				rules, ok := secret.Data["rules"].(string)
				if !ok {
					return ""
				}
				return rules
			}, 10*time.Second, 2*time.Second).Should(Equal(originalRules))
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
			_, err = vaultClient.Logical().Write("sys/mounts/drift-test-kv/tune", map[string]any{
				"default_lease_ttl": "7200",
			})
			Expect(err).To(BeNil())

			By("Verifying the drift is present")
			driftedTune, err := vaultClient.Logical().Read("sys/mounts/drift-test-kv/tune")
			Expect(err).To(BeNil())
			Expect(driftedTune).NotTo(BeNil())

			By("Waiting for RequeueAfter to fire and correct the tune config drift automatically")
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
			}, 30*time.Second, 2*time.Second).Should(BeTrue())
		})
	})

	Context("Drift detection disabled — drift persists", Ordered, func() {
		var policyInstance *redhatcopv1alpha1.Policy
		var origSyncPeriod time.Duration
		var origDriftEnv string
		var origDriftSet bool

		BeforeAll(func() {
			origSyncPeriod = vaultresourcecontroller.SyncPeriod
			origDriftEnv, origDriftSet = os.LookupEnv("ENABLE_DRIFT_DETECTION")

			os.Unsetenv("ENABLE_DRIFT_DETECTION")
			vaultresourcecontroller.SetSyncPeriod(5 * time.Second)

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
			vaultresourcecontroller.SetSyncPeriod(origSyncPeriod)
			if origDriftSet {
				os.Setenv("ENABLE_DRIFT_DETECTION", origDriftEnv)
			} else {
				os.Unsetenv("ENABLE_DRIFT_DETECTION")
			}
		})

		It("Should NOT correct drift when drift detection is disabled", func() {
			By("Overwriting the policy in Vault directly")
			driftedRules := `path "drifted/*" { capabilities = ["read"] }`
			_, err := vaultClient.Logical().Write("sys/policy/test-drift-policy-disabled", map[string]any{
				"policy": driftedRules,
			})
			Expect(err).To(BeNil())

			By("Verifying the drift persists — no RequeueAfter fires when drift detection is disabled")
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
			}, 15*time.Second, 2*time.Second).Should(ContainSubstring("drifted"))
		})
	})
})
