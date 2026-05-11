//go:build integration
// +build integration

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Random Secret controller for v2 secrets", func() {

	timeout := time.Second * 120
	interval := time.Second * 2
	Context("When recreating a random secret with retain policy", func() {
		It("Should create a Vault Secret when created, and be retained in Vault when the random secret is deleted", func() {

			By("Setting up KV v2 stack with reader")
			stack := SetupKVv2StackWithReader(ctx, timeout, interval)

			By("Creating new RandomSecret")

			var rsInstance *redhatcopv1alpha1.RandomSecret
			var instance *redhatcopv1alpha1.VaultSecret

			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/randomsecret/v2/08-randomsecret-retain-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			rsInstance = &redhatcopv1alpha1.RandomSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, rsInstance)).Should(Succeed())

			rslookupKey := types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated := &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rslookupKey, rsCreated, timeout, interval)

			By("Creating a new VaultSecret")

			ctx := context.Background()

			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/vaultsecret/v2/08-vaultsecret-randomsecret-retain-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			instance = &redhatcopv1alpha1.VaultSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.VaultSecret{}

			waitForReconcileSuccess(ctx, lookupKey, created, timeout, interval)

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

			By("Reference the password")

			var password_to_retain = secret.StringData["password"]

			By("Checking the Secret Data matches expected pattern")

			var isLowerCaseLetter = regexp.MustCompile(`^[a-z]+$`).MatchString
			for k := range instance.Spec.TemplatizedK8sSecret.StringData {
				val, ok := secret.Data[k]
				Expect(ok).To(BeTrue())
				s := string(val)
				Expect(isLowerCaseLetter(s)).To(BeTrue())
				Expect(len(s)).To(Equal(20))
			}

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

			By("Deleting RandomSecret")

			Expect(k8sIntegrationClient.Delete(ctx, rsInstance)).Should(Succeed())

			// Random secret should be deleted from k8s

			rslookupKey = types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			Eventually(func() bool {

				err := k8sIntegrationClient.Get(ctx, rslookupKey, rsCreated)
				if err != nil {
					return true
				}

				return false
			}, timeout, interval).Should(BeTrue())

			// Should not have removed the Vault secret

			Eventually(func() error {
				secret, _ := vaultClient.Logical().Read(rsInstance.GetPath())
				if secret == nil {
					return fmt.Errorf("secret is nil %s", rsInstance.GetPath())
				}
				return nil
			}, timeout, interval).Should(Succeed())

			// BEGIN
			By("Creating new RandomSecret (again)")

			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/randomsecret/v2/08-randomsecret-retain-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			rsInstance = &redhatcopv1alpha1.RandomSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, rsInstance)).Should(Succeed())

			rsCreated = &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rslookupKey, rsCreated, timeout, interval)

			By("Creating a new VaultSecret (again)")

			ctx = context.Background()

			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/vaultsecret/v2/08-vaultsecret-randomsecret-retain-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			instance = &redhatcopv1alpha1.VaultSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, instance)).Should(Succeed())

			lookupKey = types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created = &redhatcopv1alpha1.VaultSecret{}

			waitForReconcileSuccess(ctx, lookupKey, created, timeout, interval)

			By("Checking the Secret Exists with proper Owner Reference (again)")

			lookupKey = types.NamespacedName{Name: instance.Spec.TemplatizedK8sSecret.Name, Namespace: instance.Namespace}
			secret = &corev1.Secret{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, secret)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			kind = reflect.TypeOf(redhatcopv1alpha1.VaultSecret{}).Name()
			Expect(secret.GetObjectMeta().GetOwnerReferences()[0].Kind).Should(Equal(kind))

			By("Checking the password did not change and matches the previous password")

			Expect(secret.StringData["password"]).To(Equal(password_to_retain))

			By("Checking the Secret Data matches expected pattern (again)")

			isLowerCaseLetter = regexp.MustCompile(`^[a-z]+$`).MatchString
			for k := range instance.Spec.TemplatizedK8sSecret.StringData {
				val, ok := secret.Data[k]
				Expect(ok).To(BeTrue())
				s := string(val)
				Expect(isLowerCaseLetter(s)).To(BeTrue())
				Expect(len(s)).To(Equal(20))
			}

			By("Deleting the VaultSecret (again)")

			Expect(k8sIntegrationClient.Delete(ctx, instance)).Should(Succeed())

			By("Checking the K8s Secret was deleted (again)")

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, &corev1.Secret{})
				if err != nil {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())

			By("Deleting RandomSecret (again)")

			Expect(k8sIntegrationClient.Delete(ctx, rsInstance)).Should(Succeed())

			Eventually(func() error {
				secret, _ := vaultClient.Logical().Read(rsInstance.GetPath())
				if secret == nil {
					return fmt.Errorf("secret is nil %s", rsInstance.GetPath())
				}
				return nil
			}, timeout, interval).Should(Succeed())

			//END

			By("Tearing down KV v2 stack")
			TeardownKVv2Stack(ctx, stack, timeout, interval)

		})
	})

	Context("When creating multiple RandomSecrets contributing to the same Vault path", func() {
		It("Should merge different keys into the same Vault secret", func() {

			By("Setting up KV v2 stack")
			stack := SetupKVv2Stack(ctx, timeout, interval)

			By("Creating first RandomSecret with password key")

			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/randomsecret/v2/09-randomsecret-multikey-password-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			rsPasswordInstance := &redhatcopv1alpha1.RandomSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, rsPasswordInstance)).Should(Succeed())

			rsPasswordLookupKey := types.NamespacedName{Name: rsPasswordInstance.Name, Namespace: rsPasswordInstance.Namespace}
			rsPasswordCreated := &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rsPasswordLookupKey, rsPasswordCreated, timeout, interval)

			By("Verifying Vault secret has password key")

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(rsPasswordInstance.GetPath())
				if err != nil {
					return err
				}
				if secret == nil {
					return fmt.Errorf("secret is nil")
				}
				data, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("data field is not a map")
				}
				if _, ok := data["password"]; !ok {
					return fmt.Errorf("password key not found")
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("Creating second RandomSecret with username key at same path")

			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/randomsecret/v2/10-randomsecret-multikey-username-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			rsUsernameInstance := &redhatcopv1alpha1.RandomSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, rsUsernameInstance)).Should(Succeed())

			rsUsernameLookupKey := types.NamespacedName{Name: rsUsernameInstance.Name, Namespace: rsUsernameInstance.Namespace}
			rsUsernameCreated := &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rsUsernameLookupKey, rsUsernameCreated, timeout, interval)

			By("Verifying Vault secret has both password and username keys")

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(rsPasswordInstance.GetPath())
				if err != nil {
					return err
				}
				if secret == nil {
					return fmt.Errorf("secret is nil")
				}
				data, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("data field is not a map")
				}
				if _, ok := data["password"]; !ok {
					return fmt.Errorf("password key not found")
				}
				if _, ok := data["username"]; !ok {
					return fmt.Errorf("username key not found")
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("Deleting first RandomSecret (password)")

			Expect(k8sIntegrationClient.Delete(ctx, rsPasswordInstance)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rsPasswordLookupKey, rsPasswordCreated)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Verifying username key still exists in Vault after deleting password RandomSecret")

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(rsPasswordInstance.GetPath())
				if err != nil {
					return err
				}
				if secret == nil {
					return fmt.Errorf("secret is nil")
				}
				data, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("data field is not a map")
				}
				if _, ok := data["username"]; !ok {
					return fmt.Errorf("username key not found")
				}
				// Password should still exist since kvSecretRetainPolicy is Retain
				if _, ok := data["password"]; !ok {
					return fmt.Errorf("password key should still exist with Retain policy")
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("Deleting second RandomSecret (username)")

			Expect(k8sIntegrationClient.Delete(ctx, rsUsernameInstance)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rsUsernameLookupKey, rsUsernameCreated)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Tearing down KV v2 stack")
			TeardownKVv2Stack(ctx, stack, timeout, interval)

		})
	})

	Context("When recreating a multi-key RandomSecret with retain policy", func() {
		It("Should preserve existing key values and not overwrite them on recreate", func() {

			By("Setting up KV v2 stack")
			stack := SetupKVv2Stack(ctx, timeout, interval)

			By("Creating first RandomSecret with password key")

			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/randomsecret/v2/09-randomsecret-multikey-password-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			rsPasswordInstance := &redhatcopv1alpha1.RandomSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, rsPasswordInstance)).Should(Succeed())

			rsPasswordLookupKey := types.NamespacedName{Name: rsPasswordInstance.Name, Namespace: rsPasswordInstance.Namespace}
			rsPasswordCreated := &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rsPasswordLookupKey, rsPasswordCreated, timeout, interval)

			By("Creating second RandomSecret with username key at same path")

			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/randomsecret/v2/10-randomsecret-multikey-username-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			rsUsernameInstance := &redhatcopv1alpha1.RandomSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, rsUsernameInstance)).Should(Succeed())

			rsUsernameLookupKey := types.NamespacedName{Name: rsUsernameInstance.Name, Namespace: rsUsernameInstance.Namespace}
			rsUsernameCreated := &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rsUsernameLookupKey, rsUsernameCreated, timeout, interval)

			By("Capturing original password and username values from Vault")

			var originalPassword, originalUsername string
			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(rsPasswordInstance.GetPath())
				if err != nil {
					return err
				}
				if secret == nil {
					return fmt.Errorf("secret is nil")
				}
				data, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("data field is not a map")
				}
				pw, ok := data["password"]
				if !ok {
					return fmt.Errorf("password key not found")
				}
				un, ok := data["username"]
				if !ok {
					return fmt.Errorf("username key not found")
				}
				originalPassword = pw.(string)
				originalUsername = un.(string)
				return nil
			}, timeout, interval).Should(Succeed())

			Expect(originalPassword).NotTo(BeEmpty())
			Expect(originalUsername).NotTo(BeEmpty())

			By("Deleting the password RandomSecret K8s resource (Vault secret retained)")

			Expect(k8sIntegrationClient.Delete(ctx, rsPasswordInstance)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rsPasswordLookupKey, rsPasswordCreated)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Recreating the password RandomSecret K8s resource")

			name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/randomsecret/v2/09-randomsecret-multikey-password-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			rsPasswordInstance = &redhatcopv1alpha1.RandomSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, rsPasswordInstance)).Should(Succeed())

			rsPasswordCreated = &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rsPasswordLookupKey, rsPasswordCreated, timeout, interval)

			By("Verifying both keys are preserved with their original values")

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(rsPasswordInstance.GetPath())
				if err != nil {
					return err
				}
				if secret == nil {
					return fmt.Errorf("secret is nil")
				}
				data, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("data field is not a map")
				}
				pw, ok := data["password"]
				if !ok {
					return fmt.Errorf("password key not found")
				}
				if pw.(string) != originalPassword {
					return fmt.Errorf("password was overwritten: expected %q, got %q", originalPassword, pw.(string))
				}
				un, ok := data["username"]
				if !ok {
					return fmt.Errorf("username key not found")
				}
				if un.(string) != originalUsername {
					return fmt.Errorf("username was overwritten: expected %q, got %q", originalUsername, un.(string))
				}
				return nil
			}, timeout, interval).Should(Succeed())

			By("Cleaning up - Deleting RandomSecrets")

			Expect(k8sIntegrationClient.Delete(ctx, rsPasswordInstance)).Should(Succeed())
			Expect(k8sIntegrationClient.Delete(ctx, rsUsernameInstance)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rsPasswordLookupKey, rsPasswordCreated)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rsUsernameLookupKey, rsUsernameCreated)
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Tearing down KV v2 stack")
			TeardownKVv2Stack(ctx, stack, timeout, interval)

		})
	})

	Context("When updating a RandomSecret with refreshPeriod", func() {
		It("Should regenerate the password with the updated format after the refresh period elapses", func() {

			By("Setting up KV v2 stack")
			stack := SetupKVv2Stack(ctx, timeout, interval)

			By("Creating new RandomSecret with refreshPeriod")

			name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/randomsecret/v2/11-randomsecret-refresh-v2.yaml", vaultTestNamespaceName)
			Expect(err).To(BeNil())
			rsInstance := &redhatcopv1alpha1.RandomSecret{}
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, rsInstance)).Should(Succeed())

			rsLookupKey := types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated := &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rsLookupKey, rsCreated, timeout, interval)

			By("Reading the initial password from Vault")

			var initialPassword string
			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(rsInstance.GetPath())
				if err != nil {
					return err
				}
				if secret == nil {
					return fmt.Errorf("secret is nil at %s", rsInstance.GetPath())
				}
				data, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("data field is not a map")
				}
				pw, ok := data["password"].(string)
				if !ok {
					return fmt.Errorf("password key not found or not a string")
				}
				initialPassword = pw
				return nil
			}, timeout, interval).Should(Succeed())

			By("Verifying the initial password matches the simple-password-policy-v2 pattern (20-char lowercase)")

			Expect(regexp.MustCompile(`^[a-z]{20}$`).MatchString(initialPassword)).To(BeTrue(),
				"initial password should be 20 lowercase chars, got: %s", initialPassword)

			By("Recording the initial ObservedGeneration")

			var initialObservedGeneration int64
			Expect(k8sIntegrationClient.Get(ctx, rsLookupKey, rsCreated)).Should(Succeed())
			for _, condition := range rsCreated.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialObservedGeneration = condition.ObservedGeneration
				}
			}
			Expect(initialObservedGeneration).To(BeNumerically(">", 0))

			By("Getting the latest RandomSecret for update")

			Expect(k8sIntegrationClient.Get(ctx, rsLookupKey, rsCreated)).Should(Succeed())

			By("Updating the RandomSecret spec to use an inline password policy (10-char uppercase)")

			rsCreated.Spec.SecretFormat = redhatcopv1alpha1.VaultPasswordPolicy{
				InlinePasswordPolicy: "length = 10\nrule \"charset\" {\n  charset = \"ABCDEFGHIJKLMNOPQRSTUVWXYZ\"\n  min-chars = 10\n}",
			}
			Expect(k8sIntegrationClient.Update(ctx, rsCreated)).Should(Succeed())

			By("Verifying password does not change immediately (refresh guard blocks early reconcile)")

			Consistently(func() bool {
				secret, err := vaultClient.Logical().Read(rsInstance.GetPath())
				if err != nil || secret == nil {
					return false
				}
				data, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					return false
				}
				pw, ok := data["password"].(string)
				if !ok {
					return false
				}
				return pw == initialPassword
			}, 5*time.Second, time.Second).Should(BeTrue(),
				"password should remain unchanged until the refresh period elapses, got changed value")

			By("Waiting for Vault secret to reflect the updated password format")

			refreshTimeout := time.Second * 60
			refreshInterval := time.Second * 2
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read(rsInstance.GetPath())
				if err != nil || secret == nil {
					return false
				}
				data, ok := secret.Data["data"].(map[string]interface{})
				if !ok {
					return false
				}
				pw, ok := data["password"].(string)
				if !ok {
					return false
				}
				return pw != initialPassword
			}, refreshTimeout, refreshInterval).Should(BeTrue())

			By("Verifying the new password matches the inline policy pattern (10-char uppercase)")

			secret, err := vaultClient.Logical().Read(rsInstance.GetPath())
			Expect(err).To(BeNil())
			Expect(secret).ToNot(BeNil())
			data := secret.Data["data"].(map[string]interface{})
			newPassword := data["password"].(string)
			Expect(regexp.MustCompile(`^[A-Z]{10}$`).MatchString(newPassword)).To(BeTrue(),
				"updated password should be 10 uppercase chars, got: %s", newPassword)

			By("Verifying ObservedGeneration reflects the current object generation")

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rsLookupKey, rsCreated)
				if err != nil {
					return false
				}
				for _, condition := range rsCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return condition.ObservedGeneration == rsCreated.Generation
					}
				}
				return false
			}, refreshTimeout, refreshInterval).Should(BeTrue())

			By("Deleting RandomSecret")

			Expect(k8sIntegrationClient.Delete(ctx, rsInstance)).Should(Succeed())

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rsLookupKey, &redhatcopv1alpha1.RandomSecret{})
				return err != nil
			}, timeout, interval).Should(BeTrue())

			By("Tearing down KV v2 stack")
			TeardownKVv2Stack(ctx, stack, timeout, interval)

		})
	})
})
