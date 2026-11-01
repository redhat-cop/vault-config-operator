package v1alpha1

import (
	"context"
	"encoding/json"
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
			name: "AppRoleAuthEngineRole rejects localSecretIDs change",
			validateFn: func() (admission.Warnings, error) {
				r := &AppRoleAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&AppRoleAuthEngineRole{Spec: AppRoleAuthEngineRoleSpec{
						Path:        "same/path",
						AppRoleRole: AppRoleRole{LocalSecretIDs: false},
					}},
					&AppRoleAuthEngineRole{Spec: AppRoleAuthEngineRoleSpec{
						Path:        "same/path",
						AppRoleRole: AppRoleRole{LocalSecretIDs: true},
					}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.localSecretIDs cannot be updated",
		},
		{
			name: "LDAPSecretEngineStaticRole rejects username change",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPSecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&LDAPSecretEngineStaticRole{Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						LDAPSEStaticRole: LDAPSEStaticRole{Username: "old-user"},
					}},
					&LDAPSecretEngineStaticRole{Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						LDAPSEStaticRole: LDAPSEStaticRole{Username: "new-user"},
					}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.username cannot be updated",
		},
		{
			name: "LDAPSecretEngineStaticRole rejects dn change",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPSecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&LDAPSecretEngineStaticRole{Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						LDAPSEStaticRole: LDAPSEStaticRole{DN: "uid=old,dc=example,dc=com"},
					}},
					&LDAPSecretEngineStaticRole{Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						LDAPSEStaticRole: LDAPSEStaticRole{DN: "uid=new,dc=example,dc=com"},
					}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.dn cannot be updated",
		},

		// --- Allowance tests (path unchanged, other field changed) ---

		{
			name: "AppRoleAuthEngineRole allows mutable field update",
			validateFn: func() (admission.Warnings, error) {
				r := &AppRoleAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&AppRoleAuthEngineRole{Spec: AppRoleAuthEngineRoleSpec{
						Path: "same/path",
						Name: "same-name",
						AppRoleRole: AppRoleRole{
							BindSecretID:   true,
							LocalSecretIDs: false,
							TokenPolicies:  []string{"default"},
							TokenTTL:       "1h",
						},
					}},
					&AppRoleAuthEngineRole{Spec: AppRoleAuthEngineRoleSpec{
						Path: "same/path",
						Name: "same-name",
						AppRoleRole: AppRoleRole{
							BindSecretID:   true,
							LocalSecretIDs: false,
							TokenPolicies:  []string{"default", "new-policy"},
							TokenTTL:       "2h",
						},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "AWSSecretEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &AWSSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&AWSSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: AWSSecretEngineConfigSpec{
						Path:            "same/path",
						Name:            "same-name",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
					&AWSSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: AWSSecretEngineConfigSpec{
						Path:            "same/path",
						Name:            "same-name",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "AWSSecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &AWSSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&AWSSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: AWSSecretEngineRoleSpec{
						Path:    "same/path",
						Name:    "same-name",
						AWSRole: AWSRole{CredentialType: "iam_user"},
					}},
					&AWSSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: AWSSecretEngineRoleSpec{
						Path:    "same/path",
						Name:    "same-name",
						AWSRole: AWSRole{CredentialType: "iam_user"},
					}},
				)
			},
			expectErr: false,
		},
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
				rp := 3600
				return r.ValidateUpdate(context.Background(),
					&DatabaseSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: DatabaseSecretEngineStaticRoleSpec{
						Path:           "same/path",
						DBSEStaticRole: DBSEStaticRole{RotationPeriod: &rp, PasswordCredentialConfig: &PasswordCredentialConfig{}},
					}},
					&DatabaseSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: DatabaseSecretEngineStaticRoleSpec{
						Path:           "same/path",
						DBSEStaticRole: DBSEStaticRole{RotationPeriod: &rp, PasswordCredentialConfig: &PasswordCredentialConfig{}},
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
			name: "KubernetesAuthEngineConfig rejects both tokenReviewerServiceAccount and tokenReviewerJWTSecret on update",
			validateFn: func() (admission.Warnings, error) {
				r := &KubernetesAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&KubernetesAuthEngineConfig{Spec: KubernetesAuthEngineConfigSpec{Path: "same/path"}},
					&KubernetesAuthEngineConfig{Spec: KubernetesAuthEngineConfigSpec{
						Path:                        "same/path",
						TokenReviewerServiceAccount: &corev1.LocalObjectReference{Name: "sa"},
						TokenReviewerJWTSecret: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
							Key:                  "token",
						},
					}},
				)
			},
			expectErr:    true,
			errSubstring: "mutually exclusive",
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
			name: "LDAPAuthEngineConfig allows mutable field update",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&LDAPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "cfg"}, Spec: LDAPAuthEngineConfigSpec{
						Path:       "same/path",
						LDAPConfig: LDAPConfig{URL: "ldap://old.example.com"},
					}},
					&LDAPAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "cfg"}, Spec: LDAPAuthEngineConfigSpec{
						Path:       "same/path",
						LDAPConfig: LDAPConfig{URL: "ldap://new.example.com"},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "LDAPSecretEngineConfig allows mutable field update",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&LDAPSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "cfg"}, Spec: LDAPSecretEngineConfigSpec{
						Path:            "same/path",
						Name:            "same-name",
						LDAPSEConfig:    LDAPSEConfig{URL: "ldap://old.example.com", Schema: "openldap"},
						BindCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
					&LDAPSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "cfg"}, Spec: LDAPSecretEngineConfigSpec{
						Path:            "same/path",
						Name:            "same-name",
						LDAPSEConfig:    LDAPSEConfig{URL: "ldaps://new.example.com:636", Schema: "openldap", StartTLS: true},
						BindCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "LDAPSecretEngineDynamicRole allows mutable field update",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPSecretEngineDynamicRole{}
				return r.ValidateUpdate(context.Background(),
					&LDAPSecretEngineDynamicRole{ObjectMeta: metav1.ObjectMeta{Name: "role"}, Spec: LDAPSecretEngineDynamicRoleSpec{
						Path: "same/path",
						Name: "same-name",
						LDAPSEDynamicRole: LDAPSEDynamicRole{
							CreationLDIF: "dn: cn={{.Username}},ou=Users,dc=example,dc=com",
							DeletionLDIF: "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
							DefaultTTL:   "1h",
						},
					}},
					&LDAPSecretEngineDynamicRole{ObjectMeta: metav1.ObjectMeta{Name: "role"}, Spec: LDAPSecretEngineDynamicRoleSpec{
						Path: "same/path",
						Name: "same-name",
						LDAPSEDynamicRole: LDAPSEDynamicRole{
							CreationLDIF: "dn: cn={{.Username}},ou=Users,dc=example,dc=com",
							DeletionLDIF: "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
							DefaultTTL:   "2h",
						},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "LDAPSecretEngineStaticRole allows mutable field update",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPSecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&LDAPSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "role"}, Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						Name:             "same-name",
						LDAPSEStaticRole: LDAPSEStaticRole{Username: "same-user", DN: "uid=same,dc=example,dc=com", RotationPeriod: 86400},
					}},
					&LDAPSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "role"}, Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						Name:             "same-name",
						LDAPSEStaticRole: LDAPSEStaticRole{Username: "same-user", DN: "uid=same,dc=example,dc=com", RotationPeriod: 43200},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "OCIAuthEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &OCIAuthEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&OCIAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: OCIAuthEngineConfigSpec{Path: "same/path"}},
					&OCIAuthEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: OCIAuthEngineConfigSpec{Path: "same/path"}},
				)
			},
			expectErr: false,
		},
		{
			name: "OCIAuthEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &OCIAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&OCIAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: OCIAuthEngineRoleSpec{
						Path:        "same/path",
						OCIAuthRole: OCIAuthRole{Name: "same-name", OCIDList: []string{"ocid1.group.oc1..aaa"}},
					}},
					&OCIAuthEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: OCIAuthEngineRoleSpec{
						Path:        "same/path",
						OCIAuthRole: OCIAuthRole{Name: "same-name", OCIDList: []string{"ocid1.group.oc1..aaa", "ocid1.group.oc1..bbb"}},
					}},
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
			name: "ConsulSecretEngineConfig allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &ConsulSecretEngineConfig{}
				return r.ValidateUpdate(context.Background(),
					&ConsulSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: ConsulSecretEngineConfigSpec{
						Path:            "same/path",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
					&ConsulSecretEngineConfig{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: ConsulSecretEngineConfigSpec{
						Path:            "same/path",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "cred"}},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "ConsulSecretEngineRole allows non-path update",
			validateFn: func() (admission.Warnings, error) {
				r := &ConsulSecretEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&ConsulSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: ConsulSecretEngineRoleSpec{
						Path: "same/path",
						Name: "same-name",
					}},
					&ConsulSecretEngineRole{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: ConsulSecretEngineRoleSpec{
						Path: "same/path",
						Name: "same-name",
					}},
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

func TestRabbitMQSecretEngineConfigHandleAllowsNonPathUpdate(t *testing.T) {
	handler := &RabbitMQSecretEngineConfigValidation{Client: nil}

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
