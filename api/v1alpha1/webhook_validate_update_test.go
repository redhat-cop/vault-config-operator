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
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
)

type pathUpdateTestCase struct {
	name         string
	validateFn   func() (admission.Warnings, error)
	expectErr    bool
	errSubstring string
}

func TestValidateUpdateRejectsPathChange(t *testing.T) {
	cases := []pathUpdateTestCase{
		// --- Rejection tests (path changed) ---

		{
			name: "AuthEngineMount rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &AuthEngineMount{}
				return r.ValidateUpdate(context.Background(),
					&AuthEngineMount{Spec: AuthEngineMountSpec{Path: "old/path"}},
					&AuthEngineMount{Spec: AuthEngineMountSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "AzureAuthEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &AzureAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&AzureAuthEngineConfig{Spec: AzureAuthEngineConfigSpec{Path: "old/path"}},
					&AzureAuthEngineConfig{Spec: AzureAuthEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "AzureSecretEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &AzureSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&AzureSecretEngineConfig{Spec: AzureSecretEngineConfigSpec{Path: "old/path"}},
					&AzureSecretEngineConfig{Spec: AzureSecretEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "AzureSecretEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &AzureSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&AzureSecretEngineRole{Spec: AzureSecretEngineRoleSpec{Path: "old/path"}},
					&AzureSecretEngineRole{Spec: AzureSecretEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "CertAuthEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &CertAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&CertAuthEngineConfig{Spec: CertAuthEngineConfigSpec{Path: "old/path"}},
					&CertAuthEngineConfig{Spec: CertAuthEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "CertAuthEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &CertAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&CertAuthEngineRole{Spec: CertAuthEngineRoleSpec{Path: "old/path"}},
					&CertAuthEngineRole{Spec: CertAuthEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "DatabaseSecretEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &DatabaseSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&DatabaseSecretEngineConfig{Spec: DatabaseSecretEngineConfigSpec{Path: "old/path"}},
					&DatabaseSecretEngineConfig{Spec: DatabaseSecretEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "DatabaseSecretEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &DatabaseSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&DatabaseSecretEngineRole{Spec: DatabaseSecretEngineRoleSpec{Path: "old/path"}},
					&DatabaseSecretEngineRole{Spec: DatabaseSecretEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "DatabaseSecretEngineStaticRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &DatabaseSecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&DatabaseSecretEngineStaticRole{Spec: DatabaseSecretEngineStaticRoleSpec{Path: "old/path"}},
					&DatabaseSecretEngineStaticRole{Spec: DatabaseSecretEngineStaticRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "GCPAuthEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &GCPAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&GCPAuthEngineConfig{Spec: GCPAuthEngineConfigSpec{Path: "old/path"}},
					&GCPAuthEngineConfig{Spec: GCPAuthEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "GitHubSecretEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &GitHubSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&GitHubSecretEngineConfig{Spec: GitHubSecretEngineConfigSpec{Path: "old/path"}},
					&GitHubSecretEngineConfig{Spec: GitHubSecretEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "GitHubSecretEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &GitHubSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&GitHubSecretEngineRole{Spec: GitHubSecretEngineRoleSpec{Path: "old/path"}},
					&GitHubSecretEngineRole{Spec: GitHubSecretEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "JWTOIDCAuthEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &JWTOIDCAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&JWTOIDCAuthEngineConfig{Spec: JWTOIDCAuthEngineConfigSpec{Path: "old/path"}},
					&JWTOIDCAuthEngineConfig{Spec: JWTOIDCAuthEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "KubernetesAuthEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesAuthEngineConfig{Spec: KubernetesAuthEngineConfigSpec{Path: "old/path"}},
					&KubernetesAuthEngineConfig{Spec: KubernetesAuthEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "KubernetesAuthEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesAuthEngineRole{Spec: KubernetesAuthEngineRoleSpec{Path: "old/path"}},
					&KubernetesAuthEngineRole{Spec: KubernetesAuthEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "KubernetesSecretEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesSecretEngineConfig{Spec: KubernetesSecretEngineConfigSpec{Path: "old/path"}},
					&KubernetesSecretEngineConfig{Spec: KubernetesSecretEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "KubernetesSecretEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesSecretEngineRole{Spec: KubernetesSecretEngineRoleSpec{Path: "old/path"}},
					&KubernetesSecretEngineRole{Spec: KubernetesSecretEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "LDAPAuthEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&LDAPAuthEngineConfig{Spec: LDAPAuthEngineConfigSpec{Path: "old/path"}},
					&LDAPAuthEngineConfig{Spec: LDAPAuthEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "PKISecretEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &PKISecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&PKISecretEngineConfig{Spec: PKISecretEngineConfigSpec{Path: "old/path"}},
					&PKISecretEngineConfig{Spec: PKISecretEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "PKISecretEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &PKISecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&PKISecretEngineRole{Spec: PKISecretEngineRoleSpec{Path: "old/path"}},
					&PKISecretEngineRole{Spec: PKISecretEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "QuaySecretEngineConfig rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &QuaySecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&QuaySecretEngineConfig{Spec: QuaySecretEngineConfigSpec{Path: "old/path"}},
					&QuaySecretEngineConfig{Spec: QuaySecretEngineConfigSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "QuaySecretEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &QuaySecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&QuaySecretEngineRole{Spec: QuaySecretEngineRoleSpec{Path: "old/path"}},
					&QuaySecretEngineRole{Spec: QuaySecretEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "QuaySecretEngineStaticRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &QuaySecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&QuaySecretEngineStaticRole{Spec: QuaySecretEngineStaticRoleSpec{Path: "old/path"}},
					&QuaySecretEngineStaticRole{Spec: QuaySecretEngineStaticRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "RabbitMQSecretEngineRole rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &RabbitMQSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&RabbitMQSecretEngineRole{Spec: RabbitMQSecretEngineRoleSpec{Path: "old/path"}},
					&RabbitMQSecretEngineRole{Spec: RabbitMQSecretEngineRoleSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "RandomSecret rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &RandomSecret{}
				return r.ValidateUpdate(context.Background(),
					&RandomSecret{Spec: RandomSecretSpec{Path: "old/path"}},
					&RandomSecret{Spec: RandomSecretSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name: "SecretEngineMount rejects path change",
			validateFn: func() (admission.Warnings, error) {
				r := &SecretEngineMount{}
				return r.ValidateUpdate(context.Background(),
					&SecretEngineMount{Spec: SecretEngineMountSpec{Path: "old/path"}},
					&SecretEngineMount{Spec: SecretEngineMountSpec{Path: "new/path"}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},

		// --- Allowance tests (path unchanged, other field changed) ---

		{
			name: "AuthEngineMount allows config-only update",
			validateFn: func() (admission.Warnings, error) {
				r := &AuthEngineMount{}
				return r.ValidateUpdate(context.Background(),
					&AuthEngineMount{Spec: AuthEngineMountSpec{
						Path:      "same/path",
						AuthMount: AuthMount{Config: AuthMountConfig{DefaultLeaseTTL: "1h"}},
					}},
					&AuthEngineMount{Spec: AuthEngineMountSpec{
						Path:      "same/path",
						AuthMount: AuthMount{Config: AuthMountConfig{DefaultLeaseTTL: "2h"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "AzureAuthEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &AzureAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&AzureAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: AzureAuthEngineConfigSpec{Path: "same/path"}},
					&AzureAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: AzureAuthEngineConfigSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "AzureSecretEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &AzureSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&AzureSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: AzureSecretEngineConfigSpec{
						Path:             "same/path",
						AzureCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
					&AzureSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: AzureSecretEngineConfigSpec{
						Path:             "same/path",
						AzureCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "AzureSecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &AzureSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&AzureSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: AzureSecretEngineRoleSpec{Path: "same/path"}},
					&AzureSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: AzureSecretEngineRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "CertAuthEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &CertAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&CertAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: CertAuthEngineConfigSpec{Path: "same/path", Name: "same-name"}},
					&CertAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: CertAuthEngineConfigSpec{Path: "same/path", Name: "same-name"}},
				)
			},
			expectErr: false,
		},
		{
			name: "CertAuthEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &CertAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&CertAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: CertAuthEngineRoleSpec{Path: "same/path"}},
					&CertAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: CertAuthEngineRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "DatabaseSecretEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &DatabaseSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&DatabaseSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: DatabaseSecretEngineConfigSpec{
						Path:            "same/path",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
					&DatabaseSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: DatabaseSecretEngineConfigSpec{
						Path:            "same/path",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "DatabaseSecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &DatabaseSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&DatabaseSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: DatabaseSecretEngineRoleSpec{Path: "same/path"}},
					&DatabaseSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: DatabaseSecretEngineRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "DatabaseSecretEngineStaticRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &DatabaseSecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&DatabaseSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: DatabaseSecretEngineStaticRoleSpec{
						Path:           "same/path",
						DBSEStaticRole: DBSEStaticRole{PasswordCredentialConfig: &PasswordCredentialConfig{}},
					}},
					&DatabaseSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: DatabaseSecretEngineStaticRoleSpec{
						Path:           "same/path",
						DBSEStaticRole: DBSEStaticRole{PasswordCredentialConfig: &PasswordCredentialConfig{}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "GCPAuthEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &GCPAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&GCPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: GCPAuthEngineConfigSpec{Path: "same/path"}},
					&GCPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: GCPAuthEngineConfigSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "GitHubSecretEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &GitHubSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&GitHubSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: GitHubSecretEngineConfigSpec{
						Path:            "same/path",
						SSHKeyReference: SSHKeyConfig{Secret: &corev1.LocalObjectReference{Name: "ssh"}},
					}},
					&GitHubSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: GitHubSecretEngineConfigSpec{
						Path:            "same/path",
						SSHKeyReference: SSHKeyConfig{Secret: &corev1.LocalObjectReference{Name: "ssh"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "GitHubSecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &GitHubSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&GitHubSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: GitHubSecretEngineRoleSpec{Path: "same/path"}},
					&GitHubSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: GitHubSecretEngineRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "JWTOIDCAuthEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &JWTOIDCAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&JWTOIDCAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: JWTOIDCAuthEngineConfigSpec{Path: "same/path"}},
					&JWTOIDCAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: JWTOIDCAuthEngineConfigSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "KubernetesAuthEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: KubernetesAuthEngineConfigSpec{Path: "same/path"}},
					&KubernetesAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: KubernetesAuthEngineConfigSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "KubernetesAuthEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: KubernetesAuthEngineRoleSpec{
						Path:             "same/path",
						TargetNamespaces: vaultutils.TargetNamespaceConfig{TargetNamespaces: []string{"ns1"}},
					}},
					&KubernetesAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: KubernetesAuthEngineRoleSpec{
						Path:             "same/path",
						TargetNamespaces: vaultutils.TargetNamespaceConfig{TargetNamespaces: []string{"ns1"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "KubernetesSecretEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: KubernetesSecretEngineConfigSpec{
						Path:         "same/path",
						JWTReference: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "jwt"}},
					}},
					&KubernetesSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: KubernetesSecretEngineConfigSpec{
						Path:         "same/path",
						JWTReference: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "jwt"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "KubernetesSecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: KubernetesSecretEngineRoleSpec{Path: "same/path"}},
					&KubernetesSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: KubernetesSecretEngineRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "LDAPAuthEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&LDAPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: LDAPAuthEngineConfigSpec{Path: "same/path"}},
					&LDAPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: LDAPAuthEngineConfigSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "PKISecretEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &PKISecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&PKISecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: PKISecretEngineConfigSpec{
						Path:    "same/path",
						PKIType: PKIType{Type: "root", PrivateKeyType: "internal"},
					}},
					&PKISecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: PKISecretEngineConfigSpec{
						Path:    "same/path",
						PKIType: PKIType{Type: "root", PrivateKeyType: "internal"},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "PKISecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &PKISecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&PKISecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: PKISecretEngineRoleSpec{Path: "same/path"}},
					&PKISecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: PKISecretEngineRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "QuaySecretEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &QuaySecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&QuaySecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: QuaySecretEngineConfigSpec{
						Path:            "same/path",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
					&QuaySecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: QuaySecretEngineConfigSpec{
						Path:            "same/path",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "QuaySecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &QuaySecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&QuaySecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: QuaySecretEngineRoleSpec{Path: "same/path"}},
					&QuaySecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: QuaySecretEngineRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "QuaySecretEngineStaticRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &QuaySecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&QuaySecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: QuaySecretEngineStaticRoleSpec{Path: "same/path"}},
					&QuaySecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: QuaySecretEngineStaticRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "RabbitMQSecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &RabbitMQSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&RabbitMQSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: RabbitMQSecretEngineRoleSpec{Path: "same/path"}},
					&RabbitMQSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: RabbitMQSecretEngineRoleSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "RandomSecret allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &RandomSecret{}
				return r.ValidateUpdate(context.Background(),
					&RandomSecret{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: RandomSecretSpec{
						Path:         "same/path",
						SecretKey:    "key",
						SecretFormat: VaultPasswordPolicy{PasswordPolicyName: "default"},
					}},
					&RandomSecret{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: RandomSecretSpec{
						Path:         "same/path",
						SecretKey:    "key",
						SecretFormat: VaultPasswordPolicy{PasswordPolicyName: "default"},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "SecretEngineMount allows config-only update",
			validateFn: func() (admission.Warnings, error) {
				r := &SecretEngineMount{}
				return r.ValidateUpdate(context.Background(),
					&SecretEngineMount{Spec: SecretEngineMountSpec{
						Path:  "same/path",
						Mount: Mount{Config: MountConfig{DefaultLeaseTTL: "1h"}},
					}},
					&SecretEngineMount{Spec: SecretEngineMountSpec{
						Path:  "same/path",
						Mount: Mount{Config: MountConfig{DefaultLeaseTTL: "2h"}},
					}},
				)
			},
			expectErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.validateFn()

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
