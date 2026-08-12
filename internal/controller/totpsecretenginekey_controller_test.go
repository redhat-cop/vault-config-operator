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

var _ = Describe("TOTPSecretEngineKey controller", Ordered, func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	Context("When preparing a TOTP Secret Engine", func() {
		It("Should set up policy, auth role, and secret engine mount", func() {
			By("By creating a new Policy")
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/totpsecretengine/totp-engine-admin-policy.yaml", vaultAdminNamespaceName)
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
			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/totpsecretengine/totp-engine-kube-auth-role.yaml", vaultAdminNamespaceName)
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
			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/totpsecretengine/totp-secret-engine.yaml", vaultTestNamespaceName)
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

	Context("When creating a TOTPSecretEngineKey in generate mode", func() {
		It("Should create the key in Vault with ReconcileSuccessful=True", func() {
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/totpsecretengine/totp-secret-engine-key-generate.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			instance := &redhatcopv1alpha1.TOTPSecretEngineKey{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.TOTPSecretEngineKey{}

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
			secret, err := vaultClient.Logical().Read("test-vault-config-operator/totp/keys/my-totp-key")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["issuer"]).To(Equal("TestOrg"))
			Expect(secret.Data["account_name"]).To(Equal("testuser@example.com"))
			Expect(secret.Data["algorithm"]).To(Equal("SHA1"))
		})
	})

	Context("When verifying TOTP code generation", func() {
		It("Should be able to generate a TOTP code from the key", func() {
			secret, err := vaultClient.Logical().Read("test-vault-config-operator/totp/code/my-totp-key")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["code"]).NotTo(BeEmpty())
		})
	})

	Context("When creating a TOTPSecretEngineKey in import mode", func() {
		It("Should create the imported key in Vault with ReconcileSuccessful=True", func() {
			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../../test/totpsecretengine/totp-secret-engine-key-import.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			instance := &redhatcopv1alpha1.TOTPSecretEngineKey{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.TOTPSecretEngineKey{}

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

			By("Verifying the imported key exists in Vault")
			secret, err := vaultClient.Logical().Read("test-vault-config-operator/totp/keys/my-totp-import-key")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			Expect(secret.Data["issuer"]).To(Equal("ImportOrg"))
			Expect(secret.Data["account_name"]).To(Equal("importuser@example.com"))
		})
	})

	Context("When updating a TOTPSecretEngineKey", func() {
		It("Should reconcile changes to Vault", func() {
			By("Updating the issuer field on the import-mode key")
			instance := &redhatcopv1alpha1.TOTPSecretEngineKey{}
			lookupKey := types.NamespacedName{Name: "my-totp-import-key", Namespace: vaultTestNamespaceName}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, instance)).Should(Succeed())

			instance.Spec.Issuer = "UpdatedOrg"
			Expect(k8sIntegrationClient.Update(ctx, instance)).Should(Succeed())

			By("Verifying Vault reflects the updated issuer")
			Eventually(func() string {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/totp/keys/my-totp-import-key")
				if err != nil || secret == nil {
					return ""
				}
				issuer, _ := secret.Data["issuer"].(string)
				return issuer
			}, timeout, interval).Should(Equal("UpdatedOrg"))
		})
	})

	Context("When deleting a TOTPSecretEngineKey", func() {
		It("Should remove the key from Vault", func() {
			By("Deleting the generate-mode TOTPSecretEngineKey")
			keyInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.TOTPSecretEngineKey]("../../test/totpsecretengine/totp-secret-engine-key-generate.yaml")
			Expect(err).To(BeNil())
			keyInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, keyInstance)).Should(Succeed())

			By("Verifying the generate-mode key is removed from Vault")
			Eventually(func() error {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/totp/keys/my-totp-key")
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

			By("Deleting the import-mode TOTPSecretEngineKey")
			importKeyInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.TOTPSecretEngineKey]("../../test/totpsecretengine/totp-secret-engine-key-import.yaml")
			Expect(err).To(BeNil())
			importKeyInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, importKeyInstance)).Should(Succeed())

			By("Verifying the import-mode key is removed from Vault")
			Eventually(func() error {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/totp/keys/my-totp-import-key")
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
			semInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.SecretEngineMount]("../../test/totpsecretengine/totp-secret-engine.yaml")
			Expect(err).To(BeNil())
			semInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Delete(ctx, semInstance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: semInstance.Name, Namespace: semInstance.Namespace}, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up KubernetesAuthEngineRole")
			kaerInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.KubernetesAuthEngineRole]("../../test/totpsecretengine/totp-engine-kube-auth-role.yaml")
			Expect(err).To(BeNil())
			kaerInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Delete(ctx, kaerInstance)).Should(Succeed())
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: kaerInstance.Name, Namespace: kaerInstance.Namespace}, &redhatcopv1alpha1.KubernetesAuthEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Cleaning up Policy")
			pInstance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.Policy]("../../test/totpsecretengine/totp-engine-admin-policy.yaml")
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
