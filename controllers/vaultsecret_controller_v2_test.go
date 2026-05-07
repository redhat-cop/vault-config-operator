//go:build integration
// +build integration

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("VaultSecret controller for v2 secrets", func() {

	timeout := time.Second * 120
	interval := time.Second * 2
	Context("When creating a VaultSecret from multiple secrets", func() {
		It("Should create a Secret when created, and be removed from Vault when deleted", func() {

			By("Setting up KV v2 stack with reader")
			stack := SetupKVv2StackWithReader(ctx, timeout, interval)

			By("Creating new RandomSecrets")

			rsInstance, err := decoder.GetRandomSecretInstance("../test/randomsecret/v2/06-randomsecret-randomsecret-password-v2.yaml")
			Expect(err).To(BeNil())
			rsInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstance)).Should(Succeed())

			rslookupKey := types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated := &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rslookupKey, rsCreated, timeout, interval)

			rsInstanceAnotherPassword, err := decoder.GetRandomSecretInstance("../test/randomsecret/v2/07-randomsecret-randomsecret-another-password-v2.yaml")
			Expect(err).To(BeNil())
			rsInstanceAnotherPassword.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstanceAnotherPassword)).Should(Succeed())

			rslookupKey = types.NamespacedName{Name: rsInstanceAnotherPassword.Name, Namespace: rsInstanceAnotherPassword.Namespace}
			rsCreated = &redhatcopv1alpha1.RandomSecret{}

			waitForReconcileSuccess(ctx, rslookupKey, rsCreated, timeout, interval)

			By("Creating a new VaultSecret")

			ctx := context.Background()

			instance, err := decoder.GetVaultSecretInstance("../test/vaultsecret/v2/07-vaultsecret-randomsecret-v2.yaml")
			Expect(err).To(BeNil())
			instance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, instance)).Should(Succeed())

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

			By("Tearing down KV v2 stack")
			TeardownKVv2Stack(ctx, stack, timeout, interval)

		})
	})
})
