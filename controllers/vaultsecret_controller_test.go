//go:build integration
// +build integration

package controllers

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("VaultSecret controller", func() {

	timeout := time.Second * 120
	interval := time.Second * 2
	Context("When creating a VaultSecret from multiple secrets", func() {
		It("Should create a Secret when created, and be removed from Vault when deleted", func() {

			By("Creating a new PasswordPolicy")

			passwordPolicySimplePasswordInstance, err := decoder.GetPasswordPolicyInstance("../test/password-policy.yaml")
			Expect(err).To(BeNil())
			passwordPolicySimplePasswordInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, passwordPolicySimplePasswordInstance)).Should(Succeed())

			pplookupKey := types.NamespacedName{Name: passwordPolicySimplePasswordInstance.Name, Namespace: passwordPolicySimplePasswordInstance.Namespace}
			ppCreated := &redhatcopv1alpha1.PasswordPolicy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pplookupKey, ppCreated)
				if err != nil {
					return false
				}

				for _, condition := range ppCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("Creating new Policies")

			policyKVEngineAdminInstance, err := decoder.GetPolicyInstance("../test/kv-engine-admin-policy.yaml")
			Expect(err).To(BeNil())
			policyKVEngineAdminInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, policyKVEngineAdminInstance)).Should(Succeed())

			pLookupKey := types.NamespacedName{Name: policyKVEngineAdminInstance.Name, Namespace: policyKVEngineAdminInstance.Namespace}
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

			policySecretWriterInstance, err := decoder.GetPolicyInstance("../test/secret-writer-policy.yaml")
			Expect(err).To(BeNil())
			policySecretWriterInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, policySecretWriterInstance)).Should(Succeed())

			pLookupKey = types.NamespacedName{Name: policySecretWriterInstance.Name, Namespace: policySecretWriterInstance.Namespace}
			pCreated = &redhatcopv1alpha1.Policy{}

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

			policySecretReaderInstance, err := decoder.GetPolicyInstance("../test/vaultsecret/policy-secret-reader.yaml")
			Expect(err).To(BeNil())
			policySecretReaderInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, policySecretReaderInstance)).Should(Succeed())

			pLookupKey = types.NamespacedName{Name: policySecretReaderInstance.Name, Namespace: policySecretReaderInstance.Namespace}
			pCreated = &redhatcopv1alpha1.Policy{}

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

			By("Creating new KubernetesAuthEngineRoles")

			kaerKVEngineAdminInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/kv-engine-admin-role.yaml")
			Expect(err).To(BeNil())
			kaerKVEngineAdminInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerKVEngineAdminInstance)).Should(Succeed())

			kaerLookupKey := types.NamespacedName{Name: kaerKVEngineAdminInstance.Name, Namespace: kaerKVEngineAdminInstance.Namespace}
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

			kaerSecretWriterInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/secret-writer-role.yaml")
			Expect(err).To(BeNil())
			kaerSecretWriterInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerSecretWriterInstance)).Should(Succeed())

			kaerLookupKey = types.NamespacedName{Name: kaerSecretWriterInstance.Name, Namespace: kaerSecretWriterInstance.Namespace}
			kaerCreated = &redhatcopv1alpha1.KubernetesAuthEngineRole{}

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

			kaerSecretReaderInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/vaultsecret/kubernetesauthenginerole-secret-reader.yaml")
			Expect(err).To(BeNil())
			kaerSecretReaderInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerSecretReaderInstance)).Should(Succeed())

			kaerLookupKey = types.NamespacedName{Name: kaerSecretReaderInstance.Name, Namespace: kaerSecretReaderInstance.Namespace}
			kaerCreated = &redhatcopv1alpha1.KubernetesAuthEngineRole{}

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

			By("Creating a new SecretEngineMount")

			semInstance, err := decoder.GetSecretEngineMountInstance("../test/kv-secret-engine.yaml")
			Expect(err).To(BeNil())
			semInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, semInstance)).Should(Succeed())

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

			By("Creating new RandomSecrets")

			rsInstance, err := decoder.GetRandomSecretInstance("../test/random-secret.yaml")
			Expect(err).To(BeNil())
			rsInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstance)).Should(Succeed())

			rslookupKey := types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated := &redhatcopv1alpha1.RandomSecret{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rslookupKey, rsCreated)
				if err != nil {
					return false
				}

				for _, condition := range rsCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			rsInstanceAnotherPassword, err := decoder.GetRandomSecretInstance("../test/vaultsecret/randomsecret-another-password.yaml")
			Expect(err).To(BeNil())
			rsInstanceAnotherPassword.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstanceAnotherPassword)).Should(Succeed())

			rslookupKey = types.NamespacedName{Name: rsInstanceAnotherPassword.Name, Namespace: rsInstanceAnotherPassword.Namespace}
			rsCreated = &redhatcopv1alpha1.RandomSecret{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rslookupKey, rsCreated)
				if err != nil {
					return false
				}

				for _, condition := range rsCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("Creating a new VaultSecret")

			instance, err := decoder.GetVaultSecretInstance("../test/vaultsecret/vaultsecret-randomsecret.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.VaultSecret{}

			// We'll need to retry getting this newly created VaultSecret, given that creation may not immediately happen.
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

			By("Checking the Secret Exists with proper Owner Reference")

			lookupKey = types.NamespacedName{Name: instance.Spec.TemplatizedK8sSecret.Name, Namespace: instance.Namespace}
			secret := &corev1.Secret{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, secret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			kind := reflect.TypeOf(redhatcopv1alpha1.VaultSecret{}).Name()
			Expect(secret.GetObjectMeta().GetOwnerReferences()[0].Kind).Should(Equal(kind))

			By("Checking the Secret Data matches expected pattern")

			var isLowerCaseLetter = regexp.MustCompile(`^[a-z]+$`).MatchString
			for k := range instance.Spec.TemplatizedK8sSecret.StringData {
				val, ok := secret.Data[k]
				Expect(ok).To(BeTrue())
				s := string(val)
				Expect(isLowerCaseLetter(s)).To(BeTrue())
				Expect(len(s)).To(Equal(20))
			}

			By("Recording the initial ObservedGeneration from VaultSecret conditions")
			vsLookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			Expect(k8sIntegrationClient.Get(ctx, vsLookupKey, created)).Should(Succeed())
			var initialObservedGeneration int64
			for _, condition := range created.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialObservedGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialObservedGeneration).Should(BeNumerically(">", 0))

			By("Verifying initial K8s Secret has both keys with expected pattern")
			initialSecret := &corev1.Secret{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, initialSecret)).Should(Succeed())
			Expect(initialSecret.Data).Should(HaveLen(2))
			Expect(initialSecret.Data).Should(HaveKey("password"))
			Expect(initialSecret.Data).Should(HaveKey("anotherpassword"))

			// --- Update Test A: spec change WITHOUT syncOnResourceChange (time-based sync) ---

			By("Updating VaultSecret spec without syncOnResourceChange (time-based sync)")
			Eventually(func() error {
				if err := k8sIntegrationClient.Get(ctx, vsLookupKey, created); err != nil {
					return err
				}
				shortRefresh := metav1.Duration{Duration: time.Second}
				created.Spec.RefreshPeriod = &shortRefresh
				created.Spec.TemplatizedK8sSecret.StringData = map[string]string{
					"password": "{{ .randomsecret.password }}",
				}
				return k8sIntegrationClient.Update(ctx, created)
			}, timeout, interval).Should(Succeed())

			By("Waiting for K8s Secret to have exactly 1 data key via time-based sync")
			Eventually(func() int {
				s := &corev1.Secret{}
				if err := k8sIntegrationClient.Get(ctx, lookupKey, s); err != nil {
					return -1
				}
				return len(s.Data)
			}, timeout, interval).Should(Equal(1))

			By("Verifying the remaining key is password with expected pattern")
			updatedSecret := &corev1.Secret{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, updatedSecret)).Should(Succeed())
			Expect(updatedSecret.Data).Should(HaveKey("password"))
			Expect(updatedSecret.Data).ShouldNot(HaveKey("anotherpassword"))
			updatedVal := string(updatedSecret.Data["password"])
			Expect(isLowerCaseLetter(updatedVal)).To(BeTrue())
			Expect(len(updatedVal)).To(Equal(20))

			By("Verifying ObservedGeneration incremented after time-based sync")
			Eventually(func() bool {
				if err := k8sIntegrationClient.Get(ctx, vsLookupKey, created); err != nil {
					return false
				}
				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful &&
						condition.Status == metav1.ConditionTrue &&
						condition.ObservedGeneration > initialObservedGeneration {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			// --- Update Test B: spec change WITH syncOnResourceChange (immediate sync) ---

			By("Recording ObservedGeneration baseline for syncOnResourceChange test")
			Expect(k8sIntegrationClient.Get(ctx, vsLookupKey, created)).Should(Succeed())
			var genAfterTimeSync int64
			for _, condition := range created.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					genAfterTimeSync = condition.ObservedGeneration
					break
				}
			}

			By("Updating VaultSecret spec with syncOnResourceChange enabled (immediate sync)")
			Eventually(func() error {
				if err := k8sIntegrationClient.Get(ctx, vsLookupKey, created); err != nil {
					return err
				}
				longRefresh := metav1.Duration{Duration: 10 * time.Minute}
				created.Spec.RefreshPeriod = &longRefresh
				created.Spec.TemplatizedK8sSecret.StringData = map[string]string{
					"password":        "{{ .randomsecret.password }}",
					"anotherpassword": "{{ .anotherrandomsecret.password }}",
				}
				created.Spec.SyncOnResourceChange = true
				return k8sIntegrationClient.Update(ctx, created)
			}, timeout, interval).Should(Succeed())

			By("Waiting for K8s Secret to have 2 data keys via syncOnResourceChange")
			Eventually(func() int {
				s := &corev1.Secret{}
				if err := k8sIntegrationClient.Get(ctx, lookupKey, s); err != nil {
					return -1
				}
				return len(s.Data)
			}, timeout, interval).Should(Equal(2))

			By("Verifying both keys are present with expected pattern after syncOnResourceChange update")
			syncSecret := &corev1.Secret{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, syncSecret)).Should(Succeed())
			for _, key := range []string{"password", "anotherpassword"} {
				Expect(syncSecret.Data).Should(HaveKey(key))
				s := string(syncSecret.Data[key])
				Expect(isLowerCaseLetter(s)).To(BeTrue())
				Expect(len(s)).To(Equal(20))
			}

			By("Verifying ObservedGeneration incremented after syncOnResourceChange update")
			Eventually(func() bool {
				if err := k8sIntegrationClient.Get(ctx, vsLookupKey, created); err != nil {
					return false
				}
				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful &&
						condition.Status == metav1.ConditionTrue &&
						condition.ObservedGeneration > genAfterTimeSync {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			// --- Update Test C: output label change ---

			By("Updating VaultSecret output labels to add a new label")
			Eventually(func() error {
				if err := k8sIntegrationClient.Get(ctx, vsLookupKey, created); err != nil {
					return err
				}
				if created.Spec.TemplatizedK8sSecret.Labels == nil {
					created.Spec.TemplatizedK8sSecret.Labels = make(map[string]string)
				}
				created.Spec.TemplatizedK8sSecret.Labels["updated"] = "true"
				return k8sIntegrationClient.Update(ctx, created)
			}, timeout, interval).Should(Succeed())

			By("Waiting for K8s Secret to reflect the new label")
			Eventually(func() bool {
				s := &corev1.Secret{}
				if err := k8sIntegrationClient.Get(ctx, lookupKey, s); err != nil {
					return false
				}
				v, ok := s.Labels["updated"]
				return ok && v == "true"
			}, timeout, interval).Should(BeTrue())

			By("Verifying existing labels are preserved alongside the new one")
			labelSecret := &corev1.Secret{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, labelSecret)).Should(Succeed())
			Expect(labelSecret.Labels).Should(HaveKeyWithValue("app", "test-vault-config-operator"))
			Expect(labelSecret.Labels).Should(HaveKeyWithValue("updated", "true"))

			// --- Update Test D: output annotation change ---

			By("Updating VaultSecret output annotations to add a new annotation")
			Eventually(func() error {
				if err := k8sIntegrationClient.Get(ctx, vsLookupKey, created); err != nil {
					return err
				}
				if created.Spec.TemplatizedK8sSecret.Annotations == nil {
					created.Spec.TemplatizedK8sSecret.Annotations = make(map[string]string)
				}
				created.Spec.TemplatizedK8sSecret.Annotations["reviewed"] = "true"
				return k8sIntegrationClient.Update(ctx, created)
			}, timeout, interval).Should(Succeed())

			By("Waiting for K8s Secret to reflect the new annotation")
			Eventually(func() bool {
				s := &corev1.Secret{}
				if err := k8sIntegrationClient.Get(ctx, lookupKey, s); err != nil {
					return false
				}
				v, ok := s.Annotations["reviewed"]
				return ok && v == "true"
			}, timeout, interval).Should(BeTrue())

			By("Verifying existing annotations are preserved alongside the new one")
			annSecret := &corev1.Secret{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, annSecret)).Should(Succeed())
			Expect(annSecret.Annotations).Should(HaveKeyWithValue("refresh", "every-minute"))
			Expect(annSecret.Annotations).Should(HaveKeyWithValue("reviewed", "true"))

			By("Deleting the VaultSecret")

			Expect(k8sIntegrationClient.Delete(ctx, instance)).Should(Succeed())

			By("Checking the K8s Secret was deleted")

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, &corev1.Secret{})
				if err != nil {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Deleting RandomSecrets")

			Expect(k8sIntegrationClient.Delete(ctx, rsInstance)).Should(Succeed())

			Eventually(func() error {
				secret, _ := vaultClient.Logical().Read(rsInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, rsInstanceAnotherPassword)).Should(Succeed())

			Eventually(func() error {
				secret, _ := vaultClient.Logical().Read(rsInstanceAnotherPassword.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting the SecretEngineMount")

			Expect(k8sIntegrationClient.Delete(ctx, semInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(semInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting KubernetesAuthEngineRoles")

			Expect(k8sIntegrationClient.Delete(ctx, kaerSecretReaderInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(kaerSecretReaderInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, kaerSecretWriterInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(kaerSecretWriterInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, kaerKVEngineAdminInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(kaerKVEngineAdminInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting Policies")

			Expect(k8sIntegrationClient.Delete(ctx, policySecretReaderInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(policySecretReaderInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, policySecretWriterInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(policySecretWriterInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			Expect(k8sIntegrationClient.Delete(ctx, policyKVEngineAdminInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(policyKVEngineAdminInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting PasswordPolicy")

			Expect(k8sIntegrationClient.Delete(ctx, passwordPolicySimplePasswordInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(passwordPolicySimplePasswordInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

		})
	})
})
