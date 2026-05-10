package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
)

type pathUpdateTestCase struct {
	name         string
	newObj       webhook.Validator
	oldObj       runtime.Object
	expectErr    bool
	errSubstring string
}

func TestValidateUpdateRejectsPathChange(t *testing.T) {
	cases := []pathUpdateTestCase{
		// --- Rejection tests (path changed) ---

		{
			name:         "AuthEngineMount rejects path change",
			newObj:       &AuthEngineMount{Spec: AuthEngineMountSpec{Path: "new/path"}},
			oldObj:       &AuthEngineMount{Spec: AuthEngineMountSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "AzureAuthEngineConfig rejects path change",
			newObj:       &AzureAuthEngineConfig{Spec: AzureAuthEngineConfigSpec{Path: "new/path"}},
			oldObj:       &AzureAuthEngineConfig{Spec: AzureAuthEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "AzureSecretEngineConfig rejects path change",
			newObj:       &AzureSecretEngineConfig{Spec: AzureSecretEngineConfigSpec{Path: "new/path"}},
			oldObj:       &AzureSecretEngineConfig{Spec: AzureSecretEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "AzureSecretEngineRole rejects path change",
			newObj:       &AzureSecretEngineRole{Spec: AzureSecretEngineRoleSpec{Path: "new/path"}},
			oldObj:       &AzureSecretEngineRole{Spec: AzureSecretEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "CertAuthEngineConfig rejects path change",
			newObj:       &CertAuthEngineConfig{Spec: CertAuthEngineConfigSpec{Path: "new/path"}},
			oldObj:       &CertAuthEngineConfig{Spec: CertAuthEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "CertAuthEngineRole rejects path change",
			newObj:       &CertAuthEngineRole{Spec: CertAuthEngineRoleSpec{Path: "new/path"}},
			oldObj:       &CertAuthEngineRole{Spec: CertAuthEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "DatabaseSecretEngineConfig rejects path change",
			newObj:       &DatabaseSecretEngineConfig{Spec: DatabaseSecretEngineConfigSpec{Path: "new/path"}},
			oldObj:       &DatabaseSecretEngineConfig{Spec: DatabaseSecretEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "DatabaseSecretEngineRole rejects path change",
			newObj:       &DatabaseSecretEngineRole{Spec: DatabaseSecretEngineRoleSpec{Path: "new/path"}},
			oldObj:       &DatabaseSecretEngineRole{Spec: DatabaseSecretEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "DatabaseSecretEngineStaticRole rejects path change",
			newObj:       &DatabaseSecretEngineStaticRole{Spec: DatabaseSecretEngineStaticRoleSpec{Path: "new/path"}},
			oldObj:       &DatabaseSecretEngineStaticRole{Spec: DatabaseSecretEngineStaticRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "GCPAuthEngineConfig rejects path change",
			newObj:       &GCPAuthEngineConfig{Spec: GCPAuthEngineConfigSpec{Path: "new/path"}},
			oldObj:       &GCPAuthEngineConfig{Spec: GCPAuthEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "GitHubSecretEngineConfig rejects path change",
			newObj:       &GitHubSecretEngineConfig{Spec: GitHubSecretEngineConfigSpec{Path: "new/path"}},
			oldObj:       &GitHubSecretEngineConfig{Spec: GitHubSecretEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "GitHubSecretEngineRole rejects path change",
			newObj:       &GitHubSecretEngineRole{Spec: GitHubSecretEngineRoleSpec{Path: "new/path"}},
			oldObj:       &GitHubSecretEngineRole{Spec: GitHubSecretEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "JWTOIDCAuthEngineConfig rejects path change",
			newObj:       &JWTOIDCAuthEngineConfig{Spec: JWTOIDCAuthEngineConfigSpec{Path: "new/path"}},
			oldObj:       &JWTOIDCAuthEngineConfig{Spec: JWTOIDCAuthEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "KubernetesAuthEngineConfig rejects path change",
			newObj:       &KubernetesAuthEngineConfig{Spec: KubernetesAuthEngineConfigSpec{Path: "new/path"}},
			oldObj:       &KubernetesAuthEngineConfig{Spec: KubernetesAuthEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "KubernetesAuthEngineRole rejects path change",
			newObj:       &KubernetesAuthEngineRole{Spec: KubernetesAuthEngineRoleSpec{Path: "new/path"}},
			oldObj:       &KubernetesAuthEngineRole{Spec: KubernetesAuthEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "KubernetesSecretEngineConfig rejects path change",
			newObj:       &KubernetesSecretEngineConfig{Spec: KubernetesSecretEngineConfigSpec{Path: "new/path"}},
			oldObj:       &KubernetesSecretEngineConfig{Spec: KubernetesSecretEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "KubernetesSecretEngineRole rejects path change",
			newObj:       &KubernetesSecretEngineRole{Spec: KubernetesSecretEngineRoleSpec{Path: "new/path"}},
			oldObj:       &KubernetesSecretEngineRole{Spec: KubernetesSecretEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "LDAPAuthEngineConfig rejects path change",
			newObj:       &LDAPAuthEngineConfig{Spec: LDAPAuthEngineConfigSpec{Path: "new/path"}},
			oldObj:       &LDAPAuthEngineConfig{Spec: LDAPAuthEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "PKISecretEngineConfig rejects path change",
			newObj:       &PKISecretEngineConfig{Spec: PKISecretEngineConfigSpec{Path: "new/path"}},
			oldObj:       &PKISecretEngineConfig{Spec: PKISecretEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "PKISecretEngineRole rejects path change",
			newObj:       &PKISecretEngineRole{Spec: PKISecretEngineRoleSpec{Path: "new/path"}},
			oldObj:       &PKISecretEngineRole{Spec: PKISecretEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "QuaySecretEngineConfig rejects path change",
			newObj:       &QuaySecretEngineConfig{Spec: QuaySecretEngineConfigSpec{Path: "new/path"}},
			oldObj:       &QuaySecretEngineConfig{Spec: QuaySecretEngineConfigSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "QuaySecretEngineRole rejects path change",
			newObj:       &QuaySecretEngineRole{Spec: QuaySecretEngineRoleSpec{Path: "new/path"}},
			oldObj:       &QuaySecretEngineRole{Spec: QuaySecretEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "QuaySecretEngineStaticRole rejects path change",
			newObj:       &QuaySecretEngineStaticRole{Spec: QuaySecretEngineStaticRoleSpec{Path: "new/path"}},
			oldObj:       &QuaySecretEngineStaticRole{Spec: QuaySecretEngineStaticRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "RabbitMQSecretEngineRole rejects path change",
			newObj:       &RabbitMQSecretEngineRole{Spec: RabbitMQSecretEngineRoleSpec{Path: "new/path"}},
			oldObj:       &RabbitMQSecretEngineRole{Spec: RabbitMQSecretEngineRoleSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "RandomSecret rejects path change",
			newObj:       &RandomSecret{Spec: RandomSecretSpec{Path: "new/path"}},
			oldObj:       &RandomSecret{Spec: RandomSecretSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "SecretEngineMount rejects path change",
			newObj:       &SecretEngineMount{Spec: SecretEngineMountSpec{Path: "new/path"}},
			oldObj:       &SecretEngineMount{Spec: SecretEngineMountSpec{Path: "old/path"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},

		// --- Allowance tests (path unchanged, other field changed) ---

		{
			name: "AuthEngineMount allows config-only update",
			newObj: &AuthEngineMount{Spec: AuthEngineMountSpec{
				Path:      "same/path",
				AuthMount: AuthMount{Config: AuthMountConfig{DefaultLeaseTTL: "2h"}},
			}},
			oldObj: &AuthEngineMount{Spec: AuthEngineMountSpec{
				Path:      "same/path",
				AuthMount: AuthMount{Config: AuthMountConfig{DefaultLeaseTTL: "1h"}},
			}},
			expectErr: false,
		},
		{
			name:      "AzureAuthEngineConfig allows non-path update",
			newObj:    &AzureAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: AzureAuthEngineConfigSpec{Path: "same/path"}},
			oldObj:    &AzureAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: AzureAuthEngineConfigSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name: "AzureSecretEngineConfig allows non-path update",
			newObj: &AzureSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: AzureSecretEngineConfigSpec{
				Path:             "same/path",
				AzureCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
			}},
			oldObj: &AzureSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: AzureSecretEngineConfigSpec{
				Path:             "same/path",
				AzureCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
			}},
			expectErr: false,
		},
		{
			name:      "AzureSecretEngineRole allows non-path update",
			newObj:    &AzureSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: AzureSecretEngineRoleSpec{Path: "same/path"}},
			oldObj:    &AzureSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: AzureSecretEngineRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name:      "CertAuthEngineConfig allows non-path update",
			newObj:    &CertAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: CertAuthEngineConfigSpec{Path: "same/path", Name: "same-name"}},
			oldObj:    &CertAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: CertAuthEngineConfigSpec{Path: "same/path", Name: "same-name"}},
			expectErr: false,
		},
		{
			name:      "CertAuthEngineRole allows non-path update",
			newObj:    &CertAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: CertAuthEngineRoleSpec{Path: "same/path"}},
			oldObj:    &CertAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: CertAuthEngineRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name: "DatabaseSecretEngineConfig allows non-path update",
			newObj: &DatabaseSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: DatabaseSecretEngineConfigSpec{
				Path:            "same/path",
				RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
			}},
			oldObj: &DatabaseSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: DatabaseSecretEngineConfigSpec{
				Path:            "same/path",
				RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
			}},
			expectErr: false,
		},
		{
			name:      "DatabaseSecretEngineRole allows non-path update",
			newObj:    &DatabaseSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: DatabaseSecretEngineRoleSpec{Path: "same/path"}},
			oldObj:    &DatabaseSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: DatabaseSecretEngineRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name: "DatabaseSecretEngineStaticRole allows non-path update",
			newObj: &DatabaseSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: DatabaseSecretEngineStaticRoleSpec{
				Path:           "same/path",
				DBSEStaticRole: DBSEStaticRole{PasswordCredentialConfig: &PasswordCredentialConfig{}},
			}},
			oldObj: &DatabaseSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: DatabaseSecretEngineStaticRoleSpec{
				Path:           "same/path",
				DBSEStaticRole: DBSEStaticRole{PasswordCredentialConfig: &PasswordCredentialConfig{}},
			}},
			expectErr: false,
		},
		{
			name:      "GCPAuthEngineConfig allows non-path update",
			newObj:    &GCPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: GCPAuthEngineConfigSpec{Path: "same/path"}},
			oldObj:    &GCPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: GCPAuthEngineConfigSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name: "GitHubSecretEngineConfig allows non-path update",
			newObj: &GitHubSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: GitHubSecretEngineConfigSpec{
				Path:            "same/path",
				SSHKeyReference: SSHKeyConfig{Secret: &corev1.LocalObjectReference{Name: "ssh"}},
			}},
			oldObj: &GitHubSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: GitHubSecretEngineConfigSpec{
				Path:            "same/path",
				SSHKeyReference: SSHKeyConfig{Secret: &corev1.LocalObjectReference{Name: "ssh"}},
			}},
			expectErr: false,
		},
		{
			name:      "GitHubSecretEngineRole allows non-path update",
			newObj:    &GitHubSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: GitHubSecretEngineRoleSpec{Path: "same/path"}},
			oldObj:    &GitHubSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: GitHubSecretEngineRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name:      "JWTOIDCAuthEngineConfig allows non-path update",
			newObj:    &JWTOIDCAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: JWTOIDCAuthEngineConfigSpec{Path: "same/path"}},
			oldObj:    &JWTOIDCAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: JWTOIDCAuthEngineConfigSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name:      "KubernetesAuthEngineConfig allows non-path update",
			newObj:    &KubernetesAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: KubernetesAuthEngineConfigSpec{Path: "same/path"}},
			oldObj:    &KubernetesAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: KubernetesAuthEngineConfigSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name: "KubernetesAuthEngineRole allows non-path update",
			newObj: &KubernetesAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: KubernetesAuthEngineRoleSpec{
				Path:             "same/path",
				TargetNamespaces: vaultutils.TargetNamespaceConfig{TargetNamespaces: []string{"ns1"}},
			}},
			oldObj: &KubernetesAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: KubernetesAuthEngineRoleSpec{
				Path:             "same/path",
				TargetNamespaces: vaultutils.TargetNamespaceConfig{TargetNamespaces: []string{"ns1"}},
			}},
			expectErr: false,
		},
		{
			name: "KubernetesSecretEngineConfig allows non-path update",
			newObj: &KubernetesSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: KubernetesSecretEngineConfigSpec{
				Path:         "same/path",
				JWTReference: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "jwt"}},
			}},
			oldObj: &KubernetesSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: KubernetesSecretEngineConfigSpec{
				Path:         "same/path",
				JWTReference: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "jwt"}},
			}},
			expectErr: false,
		},
		{
			name:      "KubernetesSecretEngineRole allows non-path update",
			newObj:    &KubernetesSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: KubernetesSecretEngineRoleSpec{Path: "same/path"}},
			oldObj:    &KubernetesSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: KubernetesSecretEngineRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name:      "LDAPAuthEngineConfig allows non-path update",
			newObj:    &LDAPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: LDAPAuthEngineConfigSpec{Path: "same/path"}},
			oldObj:    &LDAPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: LDAPAuthEngineConfigSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name: "PKISecretEngineConfig allows non-path update",
			newObj: &PKISecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: PKISecretEngineConfigSpec{
				Path:    "same/path",
				PKIType: PKIType{Type: "root", PrivateKeyType: "internal"},
			}},
			oldObj: &PKISecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: PKISecretEngineConfigSpec{
				Path:    "same/path",
				PKIType: PKIType{Type: "root", PrivateKeyType: "internal"},
			}},
			expectErr: false,
		},
		{
			name:      "PKISecretEngineRole allows non-path update",
			newObj:    &PKISecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: PKISecretEngineRoleSpec{Path: "same/path"}},
			oldObj:    &PKISecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: PKISecretEngineRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name: "QuaySecretEngineConfig allows non-path update",
			newObj: &QuaySecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: QuaySecretEngineConfigSpec{
				Path:            "same/path",
				RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
			}},
			oldObj: &QuaySecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: QuaySecretEngineConfigSpec{
				Path:            "same/path",
				RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
			}},
			expectErr: false,
		},
		{
			name:      "QuaySecretEngineRole allows non-path update",
			newObj:    &QuaySecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: QuaySecretEngineRoleSpec{Path: "same/path"}},
			oldObj:    &QuaySecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: QuaySecretEngineRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name:      "QuaySecretEngineStaticRole allows non-path update",
			newObj:    &QuaySecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: QuaySecretEngineStaticRoleSpec{Path: "same/path"}},
			oldObj:    &QuaySecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: QuaySecretEngineStaticRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name:      "RabbitMQSecretEngineRole allows non-path update",
			newObj:    &RabbitMQSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: RabbitMQSecretEngineRoleSpec{Path: "same/path"}},
			oldObj:    &RabbitMQSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: RabbitMQSecretEngineRoleSpec{Path: "same/path"}},
			expectErr: false,
		},
		{
			name: "RandomSecret allows non-path update",
			newObj: &RandomSecret{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: RandomSecretSpec{
				Path:         "same/path",
				SecretKey:    "key",
				SecretFormat: VaultPasswordPolicy{PasswordPolicyName: "default"},
			}},
			oldObj: &RandomSecret{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: RandomSecretSpec{
				Path:         "same/path",
				SecretKey:    "key",
				SecretFormat: VaultPasswordPolicy{PasswordPolicyName: "default"},
			}},
			expectErr: false,
		},
		{
			name: "SecretEngineMount allows config-only update",
			newObj: &SecretEngineMount{Spec: SecretEngineMountSpec{
				Path:  "same/path",
				Mount: Mount{Config: MountConfig{DefaultLeaseTTL: "2h"}},
			}},
			oldObj: &SecretEngineMount{Spec: SecretEngineMountSpec{
				Path:  "same/path",
				Mount: Mount{Config: MountConfig{DefaultLeaseTTL: "1h"}},
			}},
			expectErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.newObj.ValidateUpdate(tc.oldObj)

			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errSubstring)
				} else if !strings.Contains(err.Error(), tc.errSubstring) {
					t.Errorf("expected error containing %q, got: %v", tc.errSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestRabbitMQSecretEngineConfigHandleRejectsPathChange(t *testing.T) {
	handler := &RabbitMQSecretEngineConfigValidation{Client: nil}

	t.Run("rejects path change", func(t *testing.T) {
		newObj := &RabbitMQSecretEngineConfig{Spec: RabbitMQSecretEngineConfigSpec{Path: "new/path"}}
		oldObj := &RabbitMQSecretEngineConfig{Spec: RabbitMQSecretEngineConfigSpec{Path: "old/path"}}

		newRaw, err := json.Marshal(newObj)
		if err != nil {
			t.Fatalf("failed to marshal new object: %v", err)
		}
		oldRaw, err := json.Marshal(oldObj)
		if err != nil {
			t.Fatalf("failed to marshal old object: %v", err)
		}

		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				Object:    runtime.RawExtension{Raw: newRaw},
				OldObject: runtime.RawExtension{Raw: oldRaw},
			},
		}

		resp := handler.Handle(context.Background(), req)
		if resp.Allowed {
			t.Error("expected request to be rejected when spec.path changes")
		}
		if resp.Result == nil || resp.Result.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %v", http.StatusBadRequest, resp.Result)
		}
		if resp.Result != nil && !strings.Contains(resp.Result.Message, "spec.path cannot be updated") {
			t.Errorf("expected error message containing %q, got: %s", "spec.path cannot be updated", resp.Result.Message)
		}
	})

	t.Run("allows non-path update", func(t *testing.T) {
		newObj := &RabbitMQSecretEngineConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "new"},
			Spec:       RabbitMQSecretEngineConfigSpec{Path: "same/path"},
		}
		oldObj := &RabbitMQSecretEngineConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "old"},
			Spec:       RabbitMQSecretEngineConfigSpec{Path: "same/path"},
		}

		newRaw, err := json.Marshal(newObj)
		if err != nil {
			t.Fatalf("failed to marshal new object: %v", err)
		}
		oldRaw, err := json.Marshal(oldObj)
		if err != nil {
			t.Fatalf("failed to marshal old object: %v", err)
		}

		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				Object:    runtime.RawExtension{Raw: newRaw},
				OldObject: runtime.RawExtension{Raw: oldRaw},
			},
		}

		resp := handler.Handle(context.Background(), req)
		if !resp.Allowed {
			t.Errorf("expected request to be allowed when spec.path unchanged, got rejected: %v", resp.Result)
		}
	})
}
