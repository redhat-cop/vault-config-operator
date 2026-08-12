package v1alpha1

import (
	"encoding/json"
	"strings"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLDAPSecretEngineConfig_GetPath(t *testing.T) {
	config := &LDAPSecretEngineConfig{
		Spec: LDAPSecretEngineConfigSpec{
			Path: "ldap/test",
		},
	}
	expected := "ldap/test/config"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestLDAPSecretEngineConfig_IsDeletable(t *testing.T) {
	config := &LDAPSecretEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected IsDeletable() = true")
	}
}

func TestLDAPSecretEngineConfig_toMap(t *testing.T) {
	length := 64
	config := LDAPSEConfig{
		URL:                          "ldaps://ldap.example.com:636",
		Schema:                       "openldap",
		BindDN:                       "cn=admin,dc=example,dc=com",
		PasswordPolicy:               "my-policy",
		UserDN:                       "ou=Users,dc=example,dc=com",
		UserAttr:                     "cn",
		UPNDomain:                    "example.com",
		RequestTimeout:               "90s",
		StartTLS:                     true,
		InsecureTLS:                  false,
		Certificate:                  "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
		ClientTLSCert:                "-----BEGIN CERTIFICATE-----\nMIID...\n-----END CERTIFICATE-----",
		ClientTLSKey:                 "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----",
		SkipStaticRoleImportRotation: true,
		CredentialType:               "password",
		Length:                       &length,
		ConnectionTimeout:            "30s",
		retrievedBindDN:              "cn=admin,dc=example,dc=com",
		retrievedBindPass:            "s3cret",
	}

	result := config.toMap()

	if result["url"] != "ldaps://ldap.example.com:636" {
		t.Errorf("url = %v", result["url"])
	}
	if result["schema"] != "openldap" {
		t.Errorf("schema = %v", result["schema"])
	}
	if result["binddn"] != "cn=admin,dc=example,dc=com" {
		t.Errorf("binddn = %v", result["binddn"])
	}
	if result["bindpass"] != "s3cret" {
		t.Errorf("bindpass = %v", result["bindpass"])
	}
	if result["password_policy"] != "my-policy" {
		t.Errorf("password_policy = %v", result["password_policy"])
	}
	if result["userdn"] != "ou=Users,dc=example,dc=com" {
		t.Errorf("userdn = %v", result["userdn"])
	}
	if result["userattr"] != "cn" {
		t.Errorf("userattr = %v", result["userattr"])
	}
	if result["upndomain"] != "example.com" {
		t.Errorf("upndomain = %v", result["upndomain"])
	}
	if result["request_timeout"] != "90s" {
		t.Errorf("request_timeout = %v", result["request_timeout"])
	}
	if result["starttls"] != true {
		t.Errorf("starttls = %v", result["starttls"])
	}
	if result["insecure_tls"] != false {
		t.Errorf("insecure_tls = %v", result["insecure_tls"])
	}
	if result["certificate"] != "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----" {
		t.Errorf("certificate = %v", result["certificate"])
	}
	if result["client_tls_cert"] != "-----BEGIN CERTIFICATE-----\nMIID...\n-----END CERTIFICATE-----" {
		t.Errorf("client_tls_cert = %v", result["client_tls_cert"])
	}
	if result["client_tls_key"] != "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----" {
		t.Errorf("client_tls_key = %v", result["client_tls_key"])
	}
	if result["skip_static_role_import_rotation"] != true {
		t.Errorf("skip_static_role_import_rotation = %v", result["skip_static_role_import_rotation"])
	}
	if result["credential_type"] != "password" {
		t.Errorf("credential_type = %v, want 'password'", result["credential_type"])
	}
	if result["length"] != json.Number("64") {
		t.Errorf("length = %v, want json.Number(64)", result["length"])
	}
	if result["connection_timeout"] != "30s" {
		t.Errorf("connection_timeout = %v", result["connection_timeout"])
	}
}

func TestLDAPSecretEngineConfig_toMap_BindDNFromSpec(t *testing.T) {
	config := LDAPSEConfig{
		URL:    "ldap://127.0.0.1",
		Schema: "openldap",
		BindDN: "cn=admin,dc=example,dc=com",
	}

	result := config.toMap()
	if result["binddn"] != "cn=admin,dc=example,dc=com" {
		t.Errorf("binddn = %v, expected spec.bindDN when no retrieved credentials", result["binddn"])
	}
}

func TestLDAPSecretEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	config := &LDAPSecretEngineConfig{
		Spec: LDAPSecretEngineConfigSpec{
			Path:            "ldap/test",
			BindCredentials: vaultutils.RootCredentialConfig{},
			LDAPSEConfig: LDAPSEConfig{
				URL:               "ldaps://ldap.example.com:636",
				Schema:            "openldap",
				BindDN:            "cn=admin,dc=example,dc=com",
				StartTLS:          false,
				InsecureTLS:       true,
				RequestTimeout:    "90s",
				retrievedBindDN:   "cn=admin,dc=example,dc=com",
				retrievedBindPass: "s3cret",
			},
		},
	}

	payload := map[string]any{
		"url":                              "ldaps://ldap.example.com:636",
		"schema":                           "openldap",
		"binddn":                           "cn=admin,dc=example,dc=com",
		"starttls":                         false,
		"insecure_tls":                     true,
		"request_timeout":                  "90s",
		"skip_static_role_import_rotation": false,
		"case_sensitive_names":             false,
		"tls_max_version":                  "tls12",
		"tls_min_version":                  "tls12",
		"length":                           json.Number("64"),
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for matching payload (bindpass stripped from desired)")
	}
}

func TestLDAPSecretEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	config := &LDAPSecretEngineConfig{
		Spec: LDAPSecretEngineConfigSpec{
			Path: "ldap/test",
			LDAPSEConfig: LDAPSEConfig{
				URL:               "ldaps://ldap.example.com:636",
				Schema:            "openldap",
				BindDN:            "cn=admin,dc=example,dc=com",
				retrievedBindDN:   "cn=admin,dc=example,dc=com",
				retrievedBindPass: "s3cret",
			},
		},
	}

	payload := map[string]any{
		"url":                              "ldaps://ldap.example.com:636",
		"schema":                           "ad",
		"binddn":                           "cn=admin,dc=example,dc=com",
		"starttls":                         false,
		"insecure_tls":                     false,
		"skip_static_role_import_rotation": false,
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when schema differs (openldap vs ad)")
	}
}

func TestLDAPSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	config := &LDAPSecretEngineConfig{
		Spec: LDAPSecretEngineConfigSpec{
			Path: "ldap/test",
			LDAPSEConfig: LDAPSEConfig{
				URL:               "ldap://127.0.0.1",
				Schema:            "openldap",
				retrievedBindDN:   "cn=admin,dc=example,dc=com",
				retrievedBindPass: "s3cret",
			},
		},
	}

	payload := map[string]any{
		"url":                              "ldap://127.0.0.1",
		"schema":                           "openldap",
		"binddn":                           "cn=admin,dc=example,dc=com",
		"starttls":                         false,
		"insecure_tls":                     false,
		"skip_static_role_import_rotation": false,
		"case_sensitive_names":             false,
		"tls_max_version":                  "tls12",
		"tls_min_version":                  "tls12",
		"length":                           json.Number("64"),
		"last_rotation_tolerance":          json.Number("0"),
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: extra Vault fields should be ignored via filterPayloadToDesiredKeys")
	}
}

func TestLDAPSecretEngineConfig_IsEquivalentToDesiredState_MutableFieldChanges(t *testing.T) {
	tests := []struct {
		name      string
		config    LDAPSEConfig
		payload   map[string]any
		wantEqual bool
	}{
		{
			name: "url changed returns false",
			config: LDAPSEConfig{
				URL:    "ldap://old.example.com",
				Schema: "openldap",
			},
			payload: map[string]any{
				"url":                              "ldap://new.example.com",
				"schema":                           "openldap",
				"starttls":                         false,
				"insecure_tls":                     false,
				"skip_static_role_import_rotation": false,
			},
			wantEqual: false,
		},
		{
			name: "starttls changed returns false",
			config: LDAPSEConfig{
				URL:      "ldap://127.0.0.1",
				Schema:   "openldap",
				StartTLS: false,
			},
			payload: map[string]any{
				"url":                              "ldap://127.0.0.1",
				"schema":                           "openldap",
				"starttls":                         true,
				"insecure_tls":                     false,
				"skip_static_role_import_rotation": false,
			},
			wantEqual: false,
		},
		{
			name: "userDN changed returns false",
			config: LDAPSEConfig{
				URL:    "ldap://127.0.0.1",
				Schema: "openldap",
				UserDN: "ou=Users,dc=example,dc=com",
			},
			payload: map[string]any{
				"url":                              "ldap://127.0.0.1",
				"schema":                           "openldap",
				"starttls":                         false,
				"insecure_tls":                     false,
				"skip_static_role_import_rotation": false,
				"userdn":                           "ou=People,dc=example,dc=com",
			},
			wantEqual: false,
		},
		{
			name: "identical spec returns true",
			config: LDAPSEConfig{
				URL:    "ldap://127.0.0.1",
				Schema: "openldap",
			},
			payload: map[string]any{
				"url":                              "ldap://127.0.0.1",
				"schema":                           "openldap",
				"starttls":                         false,
				"insecure_tls":                     false,
				"skip_static_role_import_rotation": false,
			},
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LDAPSecretEngineConfig{
				Spec: LDAPSecretEngineConfigSpec{
					Path:         "ldap/test",
					LDAPSEConfig: tt.config,
				},
			}
			got := config.IsEquivalentToDesiredState(tt.payload)
			if got != tt.wantEqual {
				t.Errorf("IsEquivalentToDesiredState() = %v, want %v", got, tt.wantEqual)
			}
		})
	}
}

func TestLDAPSecretEngineConfig_GetKubeAuthConfiguration(t *testing.T) {
	config := &LDAPSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: LDAPSecretEngineConfigSpec{
			Authentication: vaultutils.KubeAuthConfiguration{
				ServiceAccount: &corev1.LocalObjectReference{Name: "vault-sa"},
			},
		},
	}
	auth := config.GetKubeAuthConfiguration()
	if auth.ServiceAccount.Name != "vault-sa" {
		t.Errorf("GetKubeAuthConfiguration() service account = %v", auth.ServiceAccount.Name)
	}
}

func TestLDAPSecretEngineConfig_isValid_RandomSecretRequiresBindDN(t *testing.T) {
	tests := []struct {
		name    string
		config  *LDAPSecretEngineConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "RandomSecret without bindDN is rejected",
			config: &LDAPSecretEngineConfig{
				Spec: LDAPSecretEngineConfigSpec{
					BindCredentials: vaultutils.RootCredentialConfig{
						RandomSecret: &corev1.LocalObjectReference{Name: "my-random"},
					},
					LDAPSEConfig: LDAPSEConfig{BindDN: ""},
				},
			},
			wantErr: true,
			errMsg:  "spec.bindDN must be set when using randomSecret credentials",
		},
		{
			name: "RandomSecret with bindDN is accepted",
			config: &LDAPSecretEngineConfig{
				Spec: LDAPSecretEngineConfigSpec{
					BindCredentials: vaultutils.RootCredentialConfig{
						RandomSecret: &corev1.LocalObjectReference{Name: "my-random"},
					},
					LDAPSEConfig: LDAPSEConfig{BindDN: "cn=admin,dc=example,dc=com"},
				},
			},
			wantErr: false,
		},
		{
			name: "Secret without bindDN is accepted",
			config: &LDAPSecretEngineConfig{
				Spec: LDAPSecretEngineConfigSpec{
					BindCredentials: vaultutils.RootCredentialConfig{
						Secret: &corev1.LocalObjectReference{Name: "my-secret"},
					},
					LDAPSEConfig: LDAPSEConfig{BindDN: ""},
				},
			},
			wantErr: false,
		},
		{
			name: "VaultSecret without bindDN is accepted",
			config: &LDAPSecretEngineConfig{
				Spec: LDAPSecretEngineConfigSpec{
					BindCredentials: vaultutils.RootCredentialConfig{
						VaultSecret: &vaultutils.VaultSecretReference{Path: "secret/data/ldap"},
					},
					LDAPSEConfig: LDAPSEConfig{BindDN: ""},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.isValid()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}
