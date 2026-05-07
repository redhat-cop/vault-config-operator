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

	passwordPolicyInstance, err := decoder.GetPasswordPolicyInstance(fixturePasswordPolicyV2)
	Expect(err).To(BeNil())
	passwordPolicyInstance.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, passwordPolicyInstance)).Should(Succeed())
	stack.PasswordPolicy = passwordPolicyInstance
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: passwordPolicyInstance.Name, Namespace: passwordPolicyInstance.Namespace}, &redhatcopv1alpha1.PasswordPolicy{}, timeout, interval)

	policyKVEngineAdmin, err := decoder.GetPolicyInstance(fixturePolicyKVEngineAdminV2)
	Expect(err).To(BeNil())
	policyKVEngineAdmin.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, policyKVEngineAdmin)).Should(Succeed())
	stack.PolicyKVEngineAdmin = policyKVEngineAdmin
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: policyKVEngineAdmin.Name, Namespace: policyKVEngineAdmin.Namespace}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	policySecretWriter, err := decoder.GetPolicyInstance(fixturePolicySecretWriterV2)
	Expect(err).To(BeNil())
	policySecretWriter.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, policySecretWriter)).Should(Succeed())
	stack.PolicySecretWriter = policySecretWriter
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: policySecretWriter.Name, Namespace: policySecretWriter.Namespace}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	roleKVEngineAdmin, err := decoder.GetKubernetesAuthEngineRoleInstance(fixtureRoleKVEngineAdminV2)
	Expect(err).To(BeNil())
	roleKVEngineAdmin.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, roleKVEngineAdmin)).Should(Succeed())
	stack.RoleKVEngineAdmin = roleKVEngineAdmin
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: roleKVEngineAdmin.Name, Namespace: roleKVEngineAdmin.Namespace}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	roleSecretWriter, err := decoder.GetKubernetesAuthEngineRoleInstance(fixtureRoleSecretWriterV2)
	Expect(err).To(BeNil())
	roleSecretWriter.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, roleSecretWriter)).Should(Succeed())
	stack.RoleSecretWriter = roleSecretWriter
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: roleSecretWriter.Name, Namespace: roleSecretWriter.Namespace}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	semInstance, err := decoder.GetSecretEngineMountInstance(fixtureSEMKVv2)
	Expect(err).To(BeNil())
	semInstance.Namespace = vaultTestNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, semInstance)).Should(Succeed())
	stack.SecretEngineMount = semInstance
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: semInstance.Name, Namespace: semInstance.Namespace}, &redhatcopv1alpha1.SecretEngineMount{}, timeout, interval)

	return stack
}

func SetupKVv2StackWithReader(ctx context.Context, timeout, interval time.Duration) *KVv2Stack {
	stack := &KVv2Stack{}

	passwordPolicyInstance, err := decoder.GetPasswordPolicyInstance(fixturePasswordPolicyV2)
	Expect(err).To(BeNil())
	passwordPolicyInstance.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, passwordPolicyInstance)).Should(Succeed())
	stack.PasswordPolicy = passwordPolicyInstance
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: passwordPolicyInstance.Name, Namespace: passwordPolicyInstance.Namespace}, &redhatcopv1alpha1.PasswordPolicy{}, timeout, interval)

	policyKVEngineAdmin, err := decoder.GetPolicyInstance(fixturePolicyKVEngineAdminV2)
	Expect(err).To(BeNil())
	policyKVEngineAdmin.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, policyKVEngineAdmin)).Should(Succeed())
	stack.PolicyKVEngineAdmin = policyKVEngineAdmin
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: policyKVEngineAdmin.Name, Namespace: policyKVEngineAdmin.Namespace}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	policySecretWriter, err := decoder.GetPolicyInstance(fixturePolicySecretWriterV2)
	Expect(err).To(BeNil())
	policySecretWriter.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, policySecretWriter)).Should(Succeed())
	stack.PolicySecretWriter = policySecretWriter
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: policySecretWriter.Name, Namespace: policySecretWriter.Namespace}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	policySecretReader, err := decoder.GetPolicyInstance(fixturePolicySecretReaderV2)
	Expect(err).To(BeNil())
	policySecretReader.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, policySecretReader)).Should(Succeed())
	stack.PolicySecretReader = policySecretReader
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: policySecretReader.Name, Namespace: policySecretReader.Namespace}, &redhatcopv1alpha1.Policy{}, timeout, interval)

	roleKVEngineAdmin, err := decoder.GetKubernetesAuthEngineRoleInstance(fixtureRoleKVEngineAdminV2)
	Expect(err).To(BeNil())
	roleKVEngineAdmin.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, roleKVEngineAdmin)).Should(Succeed())
	stack.RoleKVEngineAdmin = roleKVEngineAdmin
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: roleKVEngineAdmin.Name, Namespace: roleKVEngineAdmin.Namespace}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	roleSecretWriter, err := decoder.GetKubernetesAuthEngineRoleInstance(fixtureRoleSecretWriterV2)
	Expect(err).To(BeNil())
	roleSecretWriter.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, roleSecretWriter)).Should(Succeed())
	stack.RoleSecretWriter = roleSecretWriter
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: roleSecretWriter.Name, Namespace: roleSecretWriter.Namespace}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	roleSecretReader, err := decoder.GetKubernetesAuthEngineRoleInstance(fixtureRoleSecretReaderV2)
	Expect(err).To(BeNil())
	roleSecretReader.Namespace = vaultAdminNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, roleSecretReader)).Should(Succeed())
	stack.RoleSecretReader = roleSecretReader
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: roleSecretReader.Name, Namespace: roleSecretReader.Namespace}, &redhatcopv1alpha1.KubernetesAuthEngineRole{}, timeout, interval)

	semInstance, err := decoder.GetSecretEngineMountInstance(fixtureSEMKVv2)
	Expect(err).To(BeNil())
	semInstance.Namespace = vaultTestNamespaceName
	Expect(k8sIntegrationClient.Create(ctx, semInstance)).Should(Succeed())
	stack.SecretEngineMount = semInstance
	waitForReconcileSuccess(ctx, types.NamespacedName{Name: semInstance.Name, Namespace: semInstance.Namespace}, &redhatcopv1alpha1.SecretEngineMount{}, timeout, interval)

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
