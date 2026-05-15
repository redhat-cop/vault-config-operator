//go:build integration
// +build integration

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fixturePasswordPolicyV2      = "../test/randomsecret/v2/00-passwordpolicy-simple-password-policy-v2.yaml"
	fixturePolicyKVEngineAdminV2 = "../test/randomsecret/v2/01-policy-kv-engine-admin-v2.yaml"
	fixtureRoleKVEngineAdminV2   = "../test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml"
	fixtureSEMKVv2               = "../test/randomsecret/v2/03-secretenginemount-kv-v2.yaml"
	fixturePolicySecretWriterV2  = "../test/randomsecret/v2/04-policy-secret-writer-v2.yaml"
	fixtureRoleSecretWriterV2    = "../test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml"
	fixturePolicySecretReaderV2  = "../test/vaultsecret/v2/00-policy-secret-reader-v2.yaml"
	fixtureRoleSecretReaderV2    = "../test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml"
)

type KVv2Stack struct {
	PasswordPolicy      *redhatcopv1alpha1.PasswordPolicy
	PolicyKVEngineAdmin *redhatcopv1alpha1.Policy
	PolicySecretWriter  *redhatcopv1alpha1.Policy
	PolicySecretReader  *redhatcopv1alpha1.Policy
	RoleKVEngineAdmin   *redhatcopv1alpha1.KubernetesAuthEngineRole
	RoleSecretWriter    *redhatcopv1alpha1.KubernetesAuthEngineRole
	RoleSecretReader    *redhatcopv1alpha1.KubernetesAuthEngineRole
	SecretEngineMount   *redhatcopv1alpha1.SecretEngineMount
}

type conditionsGetter interface {
	GetConditions() []metav1.Condition
}

func waitForReconcileSuccess(ctx context.Context, key types.NamespacedName, obj client.Object, timeout, interval time.Duration) {
	Eventually(func() bool {
		if err := k8sIntegrationClient.Get(ctx, key, obj); err != nil {
			return false
		}
		ca, ok := obj.(conditionsGetter)
		if !ok {
			return false
		}
		for _, c := range ca.GetConditions() {
			if c.Type == vaultresourcecontroller.ReconcileSuccessful && c.Status == metav1.ConditionTrue {
				return true
			}
		}
		return false
	}, timeout, interval).Should(BeTrue())
}

func waitForVaultCleanup(vaultPath string, timeout, interval time.Duration) {
	Eventually(func() error {
		secret, _ := vaultClient.Logical().Read(vaultPath)
		if secret == nil {
			return nil
		}
		out, _ := json.Marshal(secret)
		return fmt.Errorf("secret is not nil %s", string(out))
	}, timeout, interval).Should(Succeed())
}

func SetupKVv2Stack(ctx context.Context, timeout, interval time.Duration) *KVv2Stack {
	stack := &KVv2Stack{}

	name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixturePasswordPolicyV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.PasswordPolicy = &redhatcopv1alpha1.PasswordPolicy{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.PasswordPolicy)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.PasswordPolicy{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixturePolicyKVEngineAdminV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.PolicyKVEngineAdmin = &redhatcopv1alpha1.Policy{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.PolicyKVEngineAdmin)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixturePolicySecretWriterV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.PolicySecretWriter = &redhatcopv1alpha1.Policy{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.PolicySecretWriter)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixtureRoleKVEngineAdminV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.RoleKVEngineAdmin = &redhatcopv1alpha1.KubernetesAuthEngineRole{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.RoleKVEngineAdmin)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixtureRoleSecretWriterV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.RoleSecretWriter = &redhatcopv1alpha1.KubernetesAuthEngineRole{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.RoleSecretWriter)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixtureSEMKVv2, vaultTestNamespaceName)
	Expect(err).To(BeNil())
	stack.SecretEngineMount = &redhatcopv1alpha1.SecretEngineMount{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, stack.SecretEngineMount)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, &redhatcopv1alpha1.SecretEngineMount{}, timeout, interval)

	return stack
}

func SetupKVv2StackWithReader(ctx context.Context, timeout, interval time.Duration) *KVv2Stack {
	stack := &KVv2Stack{}

	name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixturePasswordPolicyV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.PasswordPolicy = &redhatcopv1alpha1.PasswordPolicy{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.PasswordPolicy)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.PasswordPolicy{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixturePolicyKVEngineAdminV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.PolicyKVEngineAdmin = &redhatcopv1alpha1.Policy{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.PolicyKVEngineAdmin)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixturePolicySecretWriterV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.PolicySecretWriter = &redhatcopv1alpha1.Policy{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.PolicySecretWriter)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixturePolicySecretReaderV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.PolicySecretReader = &redhatcopv1alpha1.Policy{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.PolicySecretReader)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixtureRoleKVEngineAdminV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.RoleKVEngineAdmin = &redhatcopv1alpha1.KubernetesAuthEngineRole{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.RoleKVEngineAdmin)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixtureRoleSecretWriterV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.RoleSecretWriter = &redhatcopv1alpha1.KubernetesAuthEngineRole{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.RoleSecretWriter)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixtureRoleSecretReaderV2, vaultAdminNamespaceName)
	Expect(err).To(BeNil())
	stack.RoleSecretReader = &redhatcopv1alpha1.KubernetesAuthEngineRole{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, stack.RoleSecretReader)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	name, err = decoder.CreateFromYAML(ctx, k8sIntegrationClient, fixtureSEMKVv2, vaultTestNamespaceName)
	Expect(err).To(BeNil())
	stack.SecretEngineMount = &redhatcopv1alpha1.SecretEngineMount{}
	Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, stack.SecretEngineMount)).Should(Succeed())
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: name, Namespace: vaultTestNamespaceName}, &redhatcopv1alpha1.SecretEngineMount{}, timeout, interval)

	return stack
}

func TeardownKVv2Stack(ctx context.Context, stack *KVv2Stack, timeout, interval time.Duration) {
	Expect(k8sIntegrationClient.Delete(ctx, stack.SecretEngineMount)).Should(Succeed())
	waitForVaultCleanup(stack.SecretEngineMount.GetPath(), timeout, interval)

	if stack.RoleSecretReader != nil {
		Expect(k8sIntegrationClient.Delete(ctx, stack.RoleSecretReader)).Should(Succeed())
		waitForVaultCleanup(stack.RoleSecretReader.GetPath(), timeout, interval)
	}

	Expect(k8sIntegrationClient.Delete(ctx, stack.RoleSecretWriter)).Should(Succeed())
	waitForVaultCleanup(stack.RoleSecretWriter.GetPath(), timeout, interval)

	Expect(k8sIntegrationClient.Delete(ctx, stack.RoleKVEngineAdmin)).Should(Succeed())
	waitForVaultCleanup(stack.RoleKVEngineAdmin.GetPath(), timeout, interval)

	if stack.PolicySecretReader != nil {
		Expect(k8sIntegrationClient.Delete(ctx, stack.PolicySecretReader)).Should(Succeed())
		waitForVaultCleanup(stack.PolicySecretReader.GetPath(), timeout, interval)
	}

	Expect(k8sIntegrationClient.Delete(ctx, stack.PolicySecretWriter)).Should(Succeed())
	waitForVaultCleanup(stack.PolicySecretWriter.GetPath(), timeout, interval)

	Expect(k8sIntegrationClient.Delete(ctx, stack.PolicyKVEngineAdmin)).Should(Succeed())
	waitForVaultCleanup(stack.PolicyKVEngineAdmin.GetPath(), timeout, interval)

	Expect(k8sIntegrationClient.Delete(ctx, stack.PasswordPolicy)).Should(Succeed())
	waitForVaultCleanup(stack.PasswordPolicy.GetPath(), timeout, interval)
}
