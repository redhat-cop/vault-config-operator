//go:build integration
// +build integration

package controllers

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("SecretEngineMount controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var simpleInstance *redhatcopv1alpha1.SecretEngineMount
	var tunedInstance *redhatcopv1alpha1.SecretEngineMount
	var namedInstance *redhatcopv1alpha1.SecretEngineMount

	AfterAll(func() {
		if simpleInstance != nil {
			k8sIntegrationClient.Delete(ctx, simpleInstance) //nolint:errcheck
		}
		if tunedInstance != nil {
			k8sIntegrationClient.Delete(ctx, tunedInstance) //nolint:errcheck
		}
		if namedInstance != nil {
			k8sIntegrationClient.Delete(ctx, namedInstance) //nolint:errcheck
		}
	})

	Context("When creating a simple SecretEngineMount", func() {
		It("Should enable the engine in Vault and populate the accessor", func() {

			By("Loading and creating the simple SecretEngineMount fixture")
			var err error
			simpleInstance, err = decoder.GetSecretEngineMountInstance("../test/secretenginemount/simple-kv-mount.yaml")
			Expect(err).To(BeNil())
			simpleInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, simpleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: simpleInstance.Name, Namespace: simpleInstance.Namespace}
			created := &redhatcopv1alpha1.SecretEngineMount{}

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

			By("Verifying the mount exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/mounts")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			mountData, exists := secret.Data["test-kv-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-kv-mount/' in sys/mounts")

			mountMap, ok := mountData.(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected mount data to be a map")
			Expect(mountMap["type"]).To(Equal("kv"))
			mountOptions, ok := mountMap["options"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected mount options to be a map")
			Expect(mountOptions["version"]).To(Equal("2"))

			By("Verifying status.accessor matches the Vault mount accessor")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())
			Expect(created.Status.Accessor).NotTo(BeEmpty(), "expected status.accessor to be populated")
			vaultAccessor, ok := mountMap["accessor"].(string)
			Expect(ok).To(BeTrue(), "expected accessor field in Vault mount data to be a string")
			Expect(created.Status.Accessor).To(Equal(vaultAccessor), "status.accessor should match Vault mount accessor")
		})
	})

	Context("When creating a SecretEngineMount with tune config", func() {
		It("Should apply the tune config in Vault", func() {

			By("Loading and creating the tuned SecretEngineMount fixture")
			var err error
			tunedInstance, err = decoder.GetSecretEngineMountInstance("../test/secretenginemount/tuned-kv-mount.yaml")
			Expect(err).To(BeNil())
			tunedInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, tunedInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: tunedInstance.Name, Namespace: tunedInstance.Namespace}
			created := &redhatcopv1alpha1.SecretEngineMount{}

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

			By("Verifying the tune config in Vault")
			tuneSecret, err := vaultClient.Logical().Read("sys/mounts/test-tuned-kv-mount/tune")
			Expect(err).To(BeNil())
			Expect(tuneSecret).NotTo(BeNil())

			maxLeaseTTLRaw := tuneSecret.Data["max_lease_ttl"]
			Expect(maxLeaseTTLRaw).NotTo(BeNil(), "expected max_lease_ttl in tune response")
			Expect(fmt.Sprintf("%v", maxLeaseTTLRaw)).To(Equal("31536000"), "expected 8760h = 31536000 seconds")

			Expect(tuneSecret.Data["listing_visibility"]).To(Equal("unauth"))
		})
	})

	Context("When creating a SecretEngineMount with spec.name override", func() {
		It("Should mount at the spec.name path", func() {

			By("Loading and creating the named SecretEngineMount fixture")
			var err error
			namedInstance, err = decoder.GetSecretEngineMountInstance("../test/secretenginemount/named-kv-mount.yaml")
			Expect(err).To(BeNil())
			namedInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, namedInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: namedInstance.Name, Namespace: namedInstance.Namespace}
			created := &redhatcopv1alpha1.SecretEngineMount{}

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

			By("Verifying the mount exists at the spec.name path")
			secret, err := vaultClient.Logical().Read("sys/mounts")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			_, exists := secret.Data["test-named-kv-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-named-kv-mount/' in sys/mounts")

			By("Verifying the metadata.name path was NOT created")
			_, metadataExists := secret.Data["test-named-sem-metadata/"]
			Expect(metadataExists).To(BeFalse(), "mount should NOT exist at metadata.name path 'test-named-sem-metadata/'")
		})
	})

	Context("When deleting SecretEngineMounts", func() {
		It("Should disable the engines in Vault", func() {

			Expect(simpleInstance).NotTo(BeNil(), "expected simple SecretEngineMount to be created before delete phase")
			Expect(tunedInstance).NotTo(BeNil(), "expected tuned SecretEngineMount to be created before delete phase")
			Expect(namedInstance).NotTo(BeNil(), "expected named SecretEngineMount to be created before delete phase")

			By("Deleting the simple SecretEngineMount CR")
			Expect(k8sIntegrationClient.Delete(ctx, simpleInstance)).Should(Succeed())
			simpleLookupKey := types.NamespacedName{Name: simpleInstance.Name, Namespace: simpleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, simpleLookupKey, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the tuned SecretEngineMount CR")
			Expect(k8sIntegrationClient.Delete(ctx, tunedInstance)).Should(Succeed())
			tunedLookupKey := types.NamespacedName{Name: tunedInstance.Name, Namespace: tunedInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, tunedLookupKey, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the named SecretEngineMount CR")
			Expect(k8sIntegrationClient.Delete(ctx, namedInstance)).Should(Succeed())
			namedLookupKey := types.NamespacedName{Name: namedInstance.Name, Namespace: namedInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, namedLookupKey, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying all mounts are removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/mounts")
				if err != nil || secret == nil {
					return false
				}
				_, simpleExists := secret.Data["test-kv-mount/"]
				_, tunedExists := secret.Data["test-tuned-kv-mount/"]
				_, namedExists := secret.Data["test-named-kv-mount/"]
				return !simpleExists && !tunedExists && !namedExists
			}, timeout, interval).Should(BeTrue())
		})
	})
})
