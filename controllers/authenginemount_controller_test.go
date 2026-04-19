//go:build integration
// +build integration

package controllers

import (
	"encoding/json"
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

var _ = Describe("AuthEngineMount controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	var simpleInstance *redhatcopv1alpha1.AuthEngineMount
	var tunedInstance *redhatcopv1alpha1.AuthEngineMount
	var namedInstance *redhatcopv1alpha1.AuthEngineMount

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

	Context("When creating a simple AuthEngineMount", func() {
		It("Should enable the auth method in Vault and populate the accessor", func() {

			By("Loading and creating the simple AuthEngineMount fixture")
			var err error
			simpleInstance, err = decoder.GetAuthEngineMountInstance("../test/authenginemount/simple-approle-mount.yaml")
			Expect(err).To(BeNil())
			simpleInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, simpleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: simpleInstance.Name, Namespace: simpleInstance.Namespace}
			created := &redhatcopv1alpha1.AuthEngineMount{}

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

			By("Verifying the auth mount exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/auth")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			mountData, exists := secret.Data["test-auth-mount/test-aem-simple/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-auth-mount/test-aem-simple/' in sys/auth")

			mountMap, ok := mountData.(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected mount data to be a map")
			Expect(mountMap["type"]).To(Equal("approle"))

			By("Verifying status.accessor is populated and matches Vault")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())
			Expect(created.Status.Accessor).NotTo(BeEmpty(), "expected status.accessor to be populated")
			vaultAccessor, ok := mountMap["accessor"].(string)
			Expect(ok).To(BeTrue(), "expected accessor field in Vault mount data to be a string")
			Expect(created.Status.Accessor).To(Equal(vaultAccessor), "status.accessor should match Vault mount accessor")
		})
	})

	Context("When creating an AuthEngineMount with tune config", func() {
		It("Should apply the tune config in Vault", func() {

			By("Loading and creating the tuned AuthEngineMount fixture")
			var err error
			tunedInstance, err = decoder.GetAuthEngineMountInstance("../test/authenginemount/tuned-approle-mount.yaml")
			Expect(err).To(BeNil())
			tunedInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, tunedInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: tunedInstance.Name, Namespace: tunedInstance.Namespace}
			created := &redhatcopv1alpha1.AuthEngineMount{}

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
			tuneSecret, err := vaultClient.Logical().Read("sys/auth/test-auth-mount/test-aem-tuned/tune")
			Expect(err).To(BeNil())
			Expect(tuneSecret).NotTo(BeNil())

			maxLeaseTTLRaw := tuneSecret.Data["max_lease_ttl"]
			Expect(maxLeaseTTLRaw).NotTo(BeNil(), "expected max_lease_ttl in tune response")
			maxLeaseTTL, ok := maxLeaseTTLRaw.(json.Number)
			if ok {
				maxLeaseTTLInt, err := maxLeaseTTL.Int64()
				Expect(err).To(BeNil(), "expected max_lease_ttl to parse as int64")
				Expect(maxLeaseTTLInt).To(Equal(int64(31536000)), "expected 8760h = 31536000 seconds")
			} else {
				Expect(fmt.Sprintf("%v", maxLeaseTTLRaw)).To(Equal("31536000"), "expected 8760h = 31536000 seconds (fallback)")
			}

			Expect(tuneSecret.Data["listing_visibility"]).To(Equal("unauth"))
		})
	})

	Context("When creating an AuthEngineMount with spec.name override", func() {
		It("Should mount at the spec.name path", func() {

			By("Loading and creating the named AuthEngineMount fixture")
			var err error
			namedInstance, err = decoder.GetAuthEngineMountInstance("../test/authenginemount/named-approle-mount.yaml")
			Expect(err).To(BeNil())
			namedInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, namedInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: namedInstance.Name, Namespace: namedInstance.Namespace}
			created := &redhatcopv1alpha1.AuthEngineMount{}

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
			secret, err := vaultClient.Logical().Read("sys/auth")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			_, exists := secret.Data["test-auth-mount/test-aem-named/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-auth-mount/test-aem-named/' in sys/auth")

			By("Verifying the metadata.name path was NOT created")
			_, metadataExists := secret.Data["test-auth-mount/test-aem-metadata/"]
			Expect(metadataExists).To(BeFalse(), "mount should NOT exist at metadata.name path 'test-auth-mount/test-aem-metadata/'")
		})
	})

	Context("When deleting AuthEngineMounts", func() {
		It("Should disable the auth methods in Vault", func() {

			Expect(simpleInstance).NotTo(BeNil(), "expected simple AuthEngineMount to be created before delete phase")
			Expect(tunedInstance).NotTo(BeNil(), "expected tuned AuthEngineMount to be created before delete phase")
			Expect(namedInstance).NotTo(BeNil(), "expected named AuthEngineMount to be created before delete phase")

			By("Deleting the simple AuthEngineMount CR")
			Expect(k8sIntegrationClient.Delete(ctx, simpleInstance)).Should(Succeed())
			simpleLookupKey := types.NamespacedName{Name: simpleInstance.Name, Namespace: simpleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, simpleLookupKey, &redhatcopv1alpha1.AuthEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the tuned AuthEngineMount CR")
			Expect(k8sIntegrationClient.Delete(ctx, tunedInstance)).Should(Succeed())
			tunedLookupKey := types.NamespacedName{Name: tunedInstance.Name, Namespace: tunedInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, tunedLookupKey, &redhatcopv1alpha1.AuthEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the named AuthEngineMount CR")
			Expect(k8sIntegrationClient.Delete(ctx, namedInstance)).Should(Succeed())
			namedLookupKey := types.NamespacedName{Name: namedInstance.Name, Namespace: namedInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, namedLookupKey, &redhatcopv1alpha1.AuthEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying all auth mounts are removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/auth")
				if err != nil || secret == nil {
					return false
				}
				_, simpleExists := secret.Data["test-auth-mount/test-aem-simple/"]
				_, tunedExists := secret.Data["test-auth-mount/test-aem-tuned/"]
				_, namedExists := secret.Data["test-auth-mount/test-aem-named/"]
				return !simpleExists && !tunedExists && !namedExists
			}, timeout, interval).Should(BeTrue())
		})
	})
})
