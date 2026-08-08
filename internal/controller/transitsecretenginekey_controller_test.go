//go:build integration
// +build integration

package controller

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/controllertestutils"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("TransitSecretEngineKey controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	Context("When preparing a Transit Secret Engine", func() {
		It("Should set up policy, auth role, and secret engine mount", func() {
			By("By creating a new Policy")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/transit/transit-engine-admin-policy.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			pInstance := &redhatcopv1alpha1.Policy{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, pInstance)).Should(Succeed())

			pLookupKey := types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}
			pCreated := &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pLookupKey, pCreated)
				if err != nil {
					return false
				}
				for _, condition := range pCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating KubernetesAuthEngineRole")
			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/transit/transit-engine-kube-auth-role.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			kaerInstance := &redhatcopv1alpha1.KubernetesAuthEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, kaerInstance)).Should(Succeed())

			kaerLookupKey := types.NamespacedName{Name: kaerInstance.Name, Namespace: kaerInstance.Namespace}
			kaerCreated := &redhatcopv1alpha1.KubernetesAuthEngineRole{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, kaerLookupKey, kaerCreated)
				if err != nil {
					return false
				}
				for _, condition := range kaerCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new SecretEngineMount")
			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/transit/transit-secret-engine.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			semInstance := &redhatcopv1alpha1.SecretEngineMount{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, semInstance)).Should(Succeed())

			semLookupKey := types.NamespacedName{Name: semInstance.Name, Namespace: semInstance.Namespace}
			semCreated := &redhatcopv1alpha1.SecretEngineMount{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, semLookupKey, semCreated)
				if err != nil {
					return false
				}
				for _, condition := range semCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When creating a TransitSecretEngineKey", func() {
		It("Should create the key in Vault with ReconcileSuccessful=True", func() {
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/transit/transit-secret-engine-key.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			instance := &redhatcopv1alpha1.TransitSecretEngineKey{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.TransitSecretEngineKey{}

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

			By("Verifying the key exists in Vault")
			secret, err := vaultClient.Logical().Read("test-vault-config-operator/transit/keys/my-transit-key")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["type"]).To(Equal("aes256-gcm96"))
		})
	})

	Context("When updating a TransitSecretEngineKey config", func() {
		It("Should update the config in Vault", func() {
			By("Getting the latest TransitSecretEngineKey before update")
			lookupKey := types.NamespacedName{Name: "my-transit-key", Namespace: vaultTestNamespaceName}
			created := &redhatcopv1alpha1.TransitSecretEngineKey{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())

			var initialObservedGeneration int64
			for _, condition := range created.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialObservedGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialObservedGeneration).To(BeNumerically(">", 0))

			By("Updating minDecryptionVersion and deletionAllowed")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())
			created.Spec.MinDecryptionVersion = 1
			created.Spec.DeletionAllowed = true
			Expect(k8sIntegrationClient.Update(ctx, created)).Should(Succeed())

			By("Waiting for Vault to reflect the config update")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/transit/keys/my-transit-key")
				if err != nil || secret == nil {
					return false
				}
				deletionAllowed, ok := secret.Data["deletion_allowed"].(bool)
				if !ok {
					return false
				}
				return deletionAllowed
			}, timeout, interval).Should(BeTrue())

			By("Verifying min_decryption_version reflects the updated value")
			updatedSecret, err := vaultClient.Logical().Read("test-vault-config-operator/transit/keys/my-transit-key")
			Expect(err).To(BeNil())
			Expect(updatedSecret).NotTo(BeNil())
			minDecVersion := updatedSecret.Data["min_decryption_version"]
			Expect(minDecVersion).NotTo(BeNil(), "min_decryption_version should be present in Vault response")
			switch v := minDecVersion.(type) {
			case json.Number:
				minDecVersionInt, err := v.Int64()
				Expect(err).To(BeNil())
				Expect(minDecVersionInt).To(Equal(int64(1)))
			case float64:
				Expect(int64(v)).To(Equal(int64(1)))
			default:
				Fail(fmt.Sprintf("unexpected type for min_decryption_version: %T (%v)", minDecVersion, minDecVersion))
			}

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
		})
	})

	Context("When deleting a TransitSecretEngineKey", func() {
		It("Should remove the key from Vault", func() {
			By("Verifying deletion_allowed is true before deletion")
			secret, err := vaultClient.Logical().Read("test-vault-config-operator/transit/keys/my-transit-key")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["deletion_allowed"]).To(Equal(true))

			By("Deleting TransitSecretEngineKey")
			keyInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.TransitSecretEngineKey]("../../test/transit/transit-secret-engine-key.yaml")
			Expect(err).To(BeNil())
			keyInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, keyInstance)).Should(Succeed())

			By("Verifying the key is removed from Vault")
			Eventually(func() error {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/transit/keys/my-transit-key")
				if err != nil {
					return err
				}
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Cleaning up SecretEngineMount")
			semInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.SecretEngineMount]("../../test/transit/transit-secret-engine.yaml")
			Expect(err).To(BeNil())
			semInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Delete(ctx, semInstance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: semInstance.Name, Namespace: semInstance.Namespace}, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up KubernetesAuthEngineRole")
			kaerInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.KubernetesAuthEngineRole]("../../test/transit/transit-engine-kube-auth-role.yaml")
			Expect(err).To(BeNil())
			kaerInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Delete(ctx, kaerInstance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: kaerInstance.Name, Namespace: kaerInstance.Namespace}, &redhatcopv1alpha1.KubernetesAuthEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up Policy")
			pInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.Policy]("../../test/transit/transit-engine-admin-policy.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Delete(ctx, pInstance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}, &redhatcopv1alpha1.Policy{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
