//go:build integration
// +build integration

package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("SSHSecretEngineRole controller", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var mountInstance *redhatcopv1alpha1.SecretEngineMount
	var configInstance *redhatcopv1alpha1.SSHSecretEngineConfig
	var roleCaInstance *redhatcopv1alpha1.SSHSecretEngineRole
	var roleOtpInstance *redhatcopv1alpha1.SSHSecretEngineRole

	AfterAll(func() {
		if roleCaInstance != nil {
			k8sIntegrationClient.Delete(ctx, roleCaInstance) //nolint:errcheck
		}
		if roleOtpInstance != nil {
			k8sIntegrationClient.Delete(ctx, roleOtpInstance) //nolint:errcheck
		}
		if configInstance != nil {
			k8sIntegrationClient.Delete(ctx, configInstance) //nolint:errcheck
		}
		if mountInstance != nil {
			k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
		}
	})

	Context("When creating prerequisite SSH engine mount and config", func() {
		It("Should create the SSH mount and configure the CA", func() {

			By("Loading and creating the SecretEngineMount fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ssh/ssh-secret-engine-mount.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			mountInstance = &redhatcopv1alpha1.SecretEngineMount{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, mountInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			created := &redhatcopv1alpha1.SecretEngineMount{}

			By("Waiting for ReconcileSuccessful=True on mount")
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

			By("Loading and creating the SSHSecretEngineConfig fixture")
			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ssh/ssh-secret-engine-config.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			configInstance = &redhatcopv1alpha1.SSHSecretEngineConfig{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, configInstance)).Should(Succeed())

			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			configCreated := &redhatcopv1alpha1.SSHSecretEngineConfig{}

			By("Waiting for ReconcileSuccessful=True on config")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, configCreated)
				if err != nil {
					return false
				}
				for _, condition := range configCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When creating an SSHSecretEngineRole with key_type=ca", func() {
		It("Should create the CA role in Vault", func() {

			By("Loading and creating the CA role fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ssh/ssh-secret-engine-role-ca.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			roleCaInstance = &redhatcopv1alpha1.SSHSecretEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, roleCaInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleCaInstance.Name, Namespace: roleCaInstance.Namespace}
			created := &redhatcopv1alpha1.SSHSecretEngineRole{}

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

			By("Verifying the CA role in Vault")
			secret, err := vaultClient.Logical().Read("test-ssh/test-ssh-mount/roles/test-ssh-role-ca")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			keyType, ok := secret.Data["key_type"].(string)
			Expect(ok).To(BeTrue(), "expected key_type to be a string")
			Expect(keyType).To(Equal("ca"))

			allowUserCerts, ok := secret.Data["allow_user_certificates"].(bool)
			Expect(ok).To(BeTrue(), "expected allow_user_certificates to be a bool")
			Expect(allowUserCerts).To(BeTrue())
		})
	})

	Context("When creating an SSHSecretEngineRole with key_type=otp", func() {
		It("Should create the OTP role in Vault", func() {

			By("Loading and creating the OTP role fixture")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/ssh/ssh-secret-engine-role-otp.yaml", vaultAdminNamespaceName)
			Expect(err).To(BeNil())
			roleOtpInstance = &redhatcopv1alpha1.SSHSecretEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, roleOtpInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleOtpInstance.Name, Namespace: roleOtpInstance.Namespace}
			created := &redhatcopv1alpha1.SSHSecretEngineRole{}

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

			By("Verifying the OTP role in Vault")
			secret, err := vaultClient.Logical().Read("test-ssh/test-ssh-mount/roles/test-ssh-role-otp")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			keyType, ok := secret.Data["key_type"].(string)
			Expect(ok).To(BeTrue(), "expected key_type to be a string")
			Expect(keyType).To(Equal("otp"))

			defaultUser, ok := secret.Data["default_user"].(string)
			Expect(ok).To(BeTrue(), "expected default_user to be a string")
			Expect(defaultUser).To(Equal("ubuntu"))
		})
	})

	Context("When deleting SSHSecretEngineRole resources", func() {
		It("Should clean up roles from Vault and remove all K8s resources", func() {

			Expect(roleCaInstance).NotTo(BeNil(), "expected CA role to be created before delete phase")
			Expect(roleOtpInstance).NotTo(BeNil(), "expected OTP role to be created before delete phase")

			By("Deleting the CA role CR")
			Expect(k8sIntegrationClient.Delete(ctx, roleCaInstance)).Should(Succeed())
			caLookupKey := types.NamespacedName{Name: roleCaInstance.Name, Namespace: roleCaInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, caLookupKey, &redhatcopv1alpha1.SSHSecretEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the CA role is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-ssh/test-ssh-mount/roles/test-ssh-role-ca")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the OTP role CR")
			Expect(k8sIntegrationClient.Delete(ctx, roleOtpInstance)).Should(Succeed())
			otpLookupKey := types.NamespacedName{Name: roleOtpInstance.Name, Namespace: roleOtpInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, otpLookupKey, &redhatcopv1alpha1.SSHSecretEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the OTP role is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-ssh/test-ssh-mount/roles/test-ssh-role-otp")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the config CR")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.SSHSecretEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Deleting the SecretEngineMount")
			Expect(k8sIntegrationClient.Delete(ctx, mountInstance)).Should(Succeed())
			mountLookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, mountLookupKey, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
