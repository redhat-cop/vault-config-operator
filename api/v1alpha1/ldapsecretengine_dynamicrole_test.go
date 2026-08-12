package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLDAPSecretEngineDynamicRole_GetPath(t *testing.T) {
	role := &LDAPSecretEngineDynamicRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: LDAPSecretEngineDynamicRoleSpec{
			Path: "ldap/test",
		},
	}
	expected := "ldap/test/role/my-role"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestLDAPSecretEngineDynamicRole_GetPathWithNameOverride(t *testing.T) {
	role := &LDAPSecretEngineDynamicRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: LDAPSecretEngineDynamicRoleSpec{
			Path: "ldap/test",
			Name: "custom-name",
		},
	}
	expected := "ldap/test/role/custom-name"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestLDAPSecretEngineDynamicRole_IsDeletable(t *testing.T) {
	role := &LDAPSecretEngineDynamicRole{}
	if !role.IsDeletable() {
		t.Error("expected IsDeletable() = true")
	}
}

func TestLDAPSecretEngineDynamicRole_toMap(t *testing.T) {
	role := LDAPSEDynamicRole{
		CreationLDIF:     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nobjectClass: person\ncn: {{.Username}}\nsn: {{.Username}}\nuserPassword: {{.Password}}",
		DeletionLDIF:     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
		RollbackLDIF:     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
		UsernameTemplate: "v_{{.DisplayName}}_{{.RoleName}}_{{random 10}}_{{unix_time}}",
		DefaultTTL:       "1h",
		MaxTTL:           "24h",
	}

	result := role.toMap()

	if result["creation_ldif"] != role.CreationLDIF {
		t.Errorf("creation_ldif = %v", result["creation_ldif"])
	}
	if result["deletion_ldif"] != role.DeletionLDIF {
		t.Errorf("deletion_ldif = %v", result["deletion_ldif"])
	}
	if result["rollback_ldif"] != role.RollbackLDIF {
		t.Errorf("rollback_ldif = %v", result["rollback_ldif"])
	}
	if result["username_template"] != "v_{{.DisplayName}}_{{.RoleName}}_{{random 10}}_{{unix_time}}" {
		t.Errorf("username_template = %v", result["username_template"])
	}
	if result["default_ttl"] != json.Number("3600") {
		t.Errorf("default_ttl = %v, want json.Number(3600)", result["default_ttl"])
	}
	if result["max_ttl"] != json.Number("86400") {
		t.Errorf("max_ttl = %v, want json.Number(86400)", result["max_ttl"])
	}
}

func TestLDAPSecretEngineDynamicRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	role := &LDAPSecretEngineDynamicRole{
		Spec: LDAPSecretEngineDynamicRoleSpec{
			Path: "ldap/test",
			LDAPSEDynamicRole: LDAPSEDynamicRole{
				CreationLDIF:     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nobjectClass: person",
				DeletionLDIF:     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
				RollbackLDIF:     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
				UsernameTemplate: "v_{{.DisplayName}}_{{.RoleName}}_{{random 10}}_{{unix_time}}",
				DefaultTTL:       "1h",
				MaxTTL:           "24h",
			},
		},
	}

	payload := map[string]any{
		"creation_ldif":     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nobjectClass: person",
		"deletion_ldif":     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
		"rollback_ldif":     "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
		"username_template": "v_{{.DisplayName}}_{{.RoleName}}_{{random 10}}_{{unix_time}}",
		"default_ttl":       json.Number("3600"),
		"max_ttl":           json.Number("86400"),
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for matching dynamic role payload")
	}
}

func TestLDAPSecretEngineDynamicRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &LDAPSecretEngineDynamicRole{
		Spec: LDAPSecretEngineDynamicRoleSpec{
			Path: "ldap/test",
			LDAPSEDynamicRole: LDAPSEDynamicRole{
				CreationLDIF: "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nobjectClass: person",
				DeletionLDIF: "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
				DefaultTTL:   "1h",
				MaxTTL:       "24h",
			},
		},
	}

	payload := map[string]any{
		"creation_ldif": "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nobjectClass: person",
		"deletion_ldif": "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
		"default_ttl":   json.Number("7200"),
		"max_ttl":       json.Number("86400"),
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when default_ttl differs (3600 vs 7200)")
	}
}

func TestLDAPSecretEngineDynamicRole_IsEquivalentToDesiredState_MutableFieldChanges(t *testing.T) {
	baseCreation := "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nobjectClass: person"
	baseDeletion := "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete"

	tests := []struct {
		name      string
		role      LDAPSEDynamicRole
		payload   map[string]any
		wantEqual bool
	}{
		{
			name: "creationLDIF changed returns false",
			role: LDAPSEDynamicRole{
				CreationLDIF: baseCreation,
				DeletionLDIF: baseDeletion,
				DefaultTTL:   "1h",
				MaxTTL:       "24h",
			},
			payload: map[string]any{
				"creation_ldif": "dn: cn={{.Username}},ou=People,dc=example,dc=com\nobjectClass: person",
				"deletion_ldif": baseDeletion,
				"default_ttl":   json.Number("3600"),
				"max_ttl":       json.Number("86400"),
			},
			wantEqual: false,
		},
		{
			name: "maxTTL changed returns false",
			role: LDAPSEDynamicRole{
				CreationLDIF: baseCreation,
				DeletionLDIF: baseDeletion,
				DefaultTTL:   "1h",
				MaxTTL:       "24h",
			},
			payload: map[string]any{
				"creation_ldif": baseCreation,
				"deletion_ldif": baseDeletion,
				"default_ttl":   json.Number("3600"),
				"max_ttl":       json.Number("172800"),
			},
			wantEqual: false,
		},
		{
			name: "identical spec returns true",
			role: LDAPSEDynamicRole{
				CreationLDIF: baseCreation,
				DeletionLDIF: baseDeletion,
				DefaultTTL:   "1h",
				MaxTTL:       "24h",
			},
			payload: map[string]any{
				"creation_ldif": baseCreation,
				"deletion_ldif": baseDeletion,
				"default_ttl":   json.Number("3600"),
				"max_ttl":       json.Number("86400"),
			},
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &LDAPSecretEngineDynamicRole{
				Spec: LDAPSecretEngineDynamicRoleSpec{
					Path:              "ldap/test",
					LDAPSEDynamicRole: tt.role,
				},
			}
			got := role.IsEquivalentToDesiredState(tt.payload)
			if got != tt.wantEqual {
				t.Errorf("IsEquivalentToDesiredState() = %v, want %v", got, tt.wantEqual)
			}
		})
	}
}

func TestLDAPSecretEngineDynamicRole_IsEquivalentToDesiredState_VaultTTLReadback(t *testing.T) {
	tests := []struct {
		name       string
		defaultTTL string
		maxTTL     string
		vaultDef   json.Number
		vaultMax   json.Number
		wantEqual  bool
	}{
		{
			name:       "1h matches Vault 3600",
			defaultTTL: "1h",
			maxTTL:     "24h",
			vaultDef:   json.Number("3600"),
			vaultMax:   json.Number("86400"),
			wantEqual:  true,
		},
		{
			name:       "30m matches Vault 1800",
			defaultTTL: "30m",
			maxTTL:     "2h",
			vaultDef:   json.Number("1800"),
			vaultMax:   json.Number("7200"),
			wantEqual:  true,
		},
		{
			name:       "1h30m matches Vault 5400",
			defaultTTL: "1h30m",
			maxTTL:     "48h",
			vaultDef:   json.Number("5400"),
			vaultMax:   json.Number("172800"),
			wantEqual:  true,
		},
		{
			name:       "empty TTL matches Vault 0",
			defaultTTL: "",
			maxTTL:     "",
			vaultDef:   json.Number("0"),
			vaultMax:   json.Number("0"),
			wantEqual:  true,
		},
		{
			name:       "1h does not match Vault 7200",
			defaultTTL: "1h",
			maxTTL:     "24h",
			vaultDef:   json.Number("7200"),
			vaultMax:   json.Number("86400"),
			wantEqual:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &LDAPSecretEngineDynamicRole{
				Spec: LDAPSecretEngineDynamicRoleSpec{
					Path: "ldap/test",
					LDAPSEDynamicRole: LDAPSEDynamicRole{
						CreationLDIF: "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nobjectClass: person",
						DeletionLDIF: "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
						DefaultTTL:   tt.defaultTTL,
						MaxTTL:       tt.maxTTL,
					},
				},
			}

			payload := map[string]any{
				"creation_ldif": "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nobjectClass: person",
				"deletion_ldif": "dn: cn={{.Username}},ou=Users,dc=example,dc=com\nchangetype: delete",
				"default_ttl":   tt.vaultDef,
				"max_ttl":       tt.vaultMax,
			}

			got := role.IsEquivalentToDesiredState(payload)
			if got != tt.wantEqual {
				t.Errorf("IsEquivalentToDesiredState() = %v, want %v", got, tt.wantEqual)
			}
		})
	}
}
