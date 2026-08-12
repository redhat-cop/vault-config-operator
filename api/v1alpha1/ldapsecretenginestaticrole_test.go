package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLDAPSecretEngineStaticRole_GetPath(t *testing.T) {
	role := &LDAPSecretEngineStaticRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: LDAPSecretEngineStaticRoleSpec{
			Path: "ldap/test",
		},
	}
	expected := "ldap/test/static-role/my-role"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestLDAPSecretEngineStaticRole_GetPathWithNameOverride(t *testing.T) {
	role := &LDAPSecretEngineStaticRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: LDAPSecretEngineStaticRoleSpec{
			Path: "ldap/test",
			Name: "custom-name",
		},
	}
	expected := "ldap/test/static-role/custom-name"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestLDAPSecretEngineStaticRole_IsDeletable(t *testing.T) {
	role := &LDAPSecretEngineStaticRole{}
	if !role.IsDeletable() {
		t.Error("expected IsDeletable() = true")
	}
}

func TestLDAPSecretEngineStaticRole_toMap(t *testing.T) {
	role := LDAPSEStaticRole{
		Username:           "hashicorp",
		DN:                 "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
		RotationPeriod:     86400,
		SkipImportRotation: true,
	}

	result := role.toMap()

	if result["username"] != "hashicorp" {
		t.Errorf("username = %v", result["username"])
	}
	if result["dn"] != "uid=hashicorp,ou=Users,dc=hashicorp,dc=com" {
		t.Errorf("dn = %v", result["dn"])
	}
	if result["rotation_period"] != json.Number("86400") {
		t.Errorf("rotation_period = %v, want json.Number(86400)", result["rotation_period"])
	}
	if result["skip_import_rotation"] != true {
		t.Errorf("skip_import_rotation = %v", result["skip_import_rotation"])
	}
}

func TestLDAPSecretEngineStaticRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	role := &LDAPSecretEngineStaticRole{
		Spec: LDAPSecretEngineStaticRoleSpec{
			Path: "ldap/test",
			LDAPSEStaticRole: LDAPSEStaticRole{
				Username:           "hashicorp",
				DN:                 "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
				RotationPeriod:     86400,
				SkipImportRotation: false,
			},
		},
	}

	payload := map[string]any{
		"username":             "hashicorp",
		"dn":                   "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
		"rotation_period":      json.Number("86400"),
		"skip_import_rotation": false,
		"last_vault_rotation":  "2026-03-30T16:10:00Z",
		"next_vault_rotation":  "2026-03-31T16:10:00Z",
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: last_vault_rotation/next_vault_rotation should be filtered out")
	}
}

func TestLDAPSecretEngineStaticRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &LDAPSecretEngineStaticRole{
		Spec: LDAPSecretEngineStaticRoleSpec{
			Path: "ldap/test",
			LDAPSEStaticRole: LDAPSEStaticRole{
				Username:       "hashicorp",
				DN:             "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
				RotationPeriod: 86400,
			},
		},
	}

	payload := map[string]any{
		"username":             "hashicorp",
		"dn":                   "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
		"rotation_period":      json.Number("3600"),
		"skip_import_rotation": false,
		"last_vault_rotation":  "2026-03-30T16:10:00Z",
		"next_vault_rotation":  "2026-03-31T16:10:00Z",
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when rotation_period differs (86400 vs 3600)")
	}
}

func TestLDAPSecretEngineStaticRole_IsEquivalentToDesiredState_MutableFieldChanges(t *testing.T) {
	tests := []struct {
		name      string
		role      LDAPSEStaticRole
		payload   map[string]any
		wantEqual bool
	}{
		{
			name: "rotationPeriod changed returns false",
			role: LDAPSEStaticRole{
				Username:       "hashicorp",
				DN:             "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
				RotationPeriod: 86400,
			},
			payload: map[string]any{
				"username":             "hashicorp",
				"dn":                   "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
				"rotation_period":      json.Number("43200"),
				"skip_import_rotation": false,
			},
			wantEqual: false,
		},
		{
			name: "skipImportRotation changed returns false",
			role: LDAPSEStaticRole{
				Username:           "hashicorp",
				DN:                 "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
				RotationPeriod:     86400,
				SkipImportRotation: false,
			},
			payload: map[string]any{
				"username":             "hashicorp",
				"dn":                   "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
				"rotation_period":      json.Number("86400"),
				"skip_import_rotation": true,
			},
			wantEqual: false,
		},
		{
			name: "identical spec returns true",
			role: LDAPSEStaticRole{
				Username:       "hashicorp",
				DN:             "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
				RotationPeriod: 86400,
			},
			payload: map[string]any{
				"username":             "hashicorp",
				"dn":                   "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
				"rotation_period":      json.Number("86400"),
				"skip_import_rotation": false,
			},
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &LDAPSecretEngineStaticRole{
				Spec: LDAPSecretEngineStaticRoleSpec{
					Path:             "ldap/test",
					LDAPSEStaticRole: tt.role,
				},
			}
			got := role.IsEquivalentToDesiredState(tt.payload)
			if got != tt.wantEqual {
				t.Errorf("IsEquivalentToDesiredState() = %v, want %v", got, tt.wantEqual)
			}
		})
	}
}
