package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppRoleAuthEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *AppRoleAuthEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &AppRoleAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AppRoleAuthEngineRoleSpec{
					Path: "approle",
					Name: "custom-name",
				},
			},
			expectedPath: "auth/approle/role/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &AppRoleAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AppRoleAuthEngineRoleSpec{
					Path: "approle",
				},
			},
			expectedPath: "auth/approle/role/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestAppRoleRoleToMap(t *testing.T) {
	role := AppRoleRole{
		BindSecretID:         true,
		SecretIDBoundCIDRs:   []string{"10.0.0.0/8", "192.168.1.0/24"},
		SecretIDNumUses:      40,
		SecretIDTTL:          "10m",
		LocalSecretIDs:       true,
		TokenTTL:             "20m",
		TokenMaxTTL:          "30m",
		TokenPolicies:        []string{"default", "app-policy"},
		TokenBoundCIDRs:      []string{"172.16.0.0/12"},
		TokenExplicitMaxTTL:  "24h",
		TokenNoDefaultPolicy: true,
		TokenNumUses:         5,
		TokenPeriod:          "1h",
		TokenType:            "service",
	}

	result := role.toMap()

	expected := map[string]any{
		"bind_secret_id":          true,
		"secret_id_bound_cidrs":   []any{"10.0.0.0/8", "192.168.1.0/24"},
		"secret_id_num_uses":      json.Number("40"),
		"secret_id_ttl":           json.Number("600"),
		"local_secret_ids":        true,
		"token_ttl":               json.Number("1200"),
		"token_max_ttl":           json.Number("1800"),
		"token_policies":          []any{"default", "app-policy"},
		"token_bound_cidrs":       []any{"172.16.0.0/12"},
		"token_explicit_max_ttl":  json.Number("86400"),
		"token_no_default_policy": true,
		"token_num_uses":          json.Number("5"),
		"token_period":            json.Number("3600"),
		"token_type":              "service",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}

	if len(result) != 14 {
		t.Errorf("expected 14 keys in map (all fields set), got %d", len(result))
	}
}

func TestAppRoleRoleToMap_MinimalFields(t *testing.T) {
	role := AppRoleRole{
		BindSecretID: true,
	}

	result := role.toMap()

	// Unconditional emission: all 14 fields are always present (matching Nomad/Consul pattern).
	// removeUnsetFields handles drift detection for zero-valued fields Vault omits.
	if len(result) != 14 {
		t.Errorf("expected 14 keys in map (unconditional emission), got %d: %v", len(result), result)
	}

	expected := map[string]any{
		"bind_secret_id":          true,
		"secret_id_bound_cidrs":   []any{},
		"secret_id_num_uses":      json.Number("0"),
		"secret_id_ttl":           json.Number("0"),
		"local_secret_ids":        false,
		"token_ttl":               json.Number("0"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestAppRoleRoleToMap_ZeroDurationFields(t *testing.T) {
	role := AppRoleRole{
		BindSecretID:        true,
		SecretIDTTL:         "0s",
		TokenTTL:            "0",
		TokenMaxTTL:         "0s",
		TokenExplicitMaxTTL: "0s",
		TokenPeriod:         "0s",
	}

	result := role.toMap()

	// All 14 fields are emitted unconditionally
	if len(result) != 14 {
		t.Errorf("expected 14 keys in map (unconditional emission), got %d: %v", len(result), result)
	}

	for _, key := range []string{"secret_id_ttl", "token_ttl", "token_max_ttl", "token_explicit_max_ttl", "token_period"} {
		val, exists := result[key]
		if !exists {
			t.Errorf("zero-duration field %q should be emitted", key)
			continue
		}
		if val != json.Number("0") {
			t.Errorf("expected %q = json.Number(\"0\"), got %v", key, val)
		}
	}
}

// TestAppRoleRoleToMap_ClearedFields documents the project-standard behavior when fields
// that were previously set are "cleared" back to zero/empty values. With unconditional
// emission (matching Nomad/Consul pattern), cleared fields produce zero values in the map.
// The removeUnsetFields helper in IsEquivalentToDesiredState prevents false drift when
// Vault omits these zero-valued fields from its read response.
func TestAppRoleRoleToMap_ClearedFields(t *testing.T) {
	populated := AppRoleRole{
		BindSecretID:         true,
		SecretIDBoundCIDRs:   []string{"10.0.0.0/8"},
		SecretIDNumUses:      10,
		SecretIDTTL:          "30m",
		LocalSecretIDs:       true,
		TokenTTL:             "1h",
		TokenMaxTTL:          "4h",
		TokenPolicies:        []string{"default", "admin"},
		TokenBoundCIDRs:      []string{"172.16.0.0/12"},
		TokenExplicitMaxTTL:  "24h",
		TokenNoDefaultPolicy: true,
		TokenNumUses:         5,
		TokenPeriod:          "12h",
		TokenType:            "service",
	}
	populatedMap := populated.toMap()

	if len(populatedMap) != 14 {
		t.Fatalf("populated role should have 14 keys, got %d", len(populatedMap))
	}
	if _, ok := populatedMap["secret_id_bound_cidrs"]; !ok {
		t.Fatal("populated role should have secret_id_bound_cidrs")
	}

	cleared := AppRoleRole{
		BindSecretID: true,
	}
	clearedMap := cleared.toMap()

	if len(clearedMap) != 14 {
		t.Fatalf("cleared role should still have 14 keys (unconditional emission), got %d", len(clearedMap))
	}

	// Cleared fields produce zero values, not absent keys
	zeroChecks := map[string]any{
		"secret_id_bound_cidrs":   []any{},
		"secret_id_num_uses":      json.Number("0"),
		"secret_id_ttl":           json.Number("0"),
		"local_secret_ids":        false,
		"token_ttl":               json.Number("0"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
	}
	for key, expectedVal := range zeroChecks {
		got, exists := clearedMap[key]
		if !exists {
			t.Errorf("cleared field %q should still be present in map", key)
			continue
		}
		if !reflect.DeepEqual(got, expectedVal) {
			t.Errorf("cleared field %q = %v (%T), want %v (%T)", key, got, got, expectedVal, expectedVal)
		}
	}
}

// TestAppRoleRoleToMap_ClearedFields_NoDrift verifies that clearing fields back to
// zero/empty values does not cause false drift detection when Vault omits those fields.
func TestAppRoleRoleToMap_ClearedFields_NoDrift(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID: true,
			},
		},
	}

	// Vault response for a role where policies/CIDRs/TTLs were previously set
	// but are now at defaults — Vault may omit these fields entirely.
	payload := map[string]any{
		"bind_secret_id": true,
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("cleared fields should not cause drift when Vault omits zero-valued fields")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_ZeroPeriodNoDrift(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:  true,
				TokenPolicies: []string{"default"},
				TokenPeriod:   "0s",
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id": true,
		"token_policies": []any{"default"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected tokenPeriod='0s' to NOT cause drift when Vault omits token_period")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_Match(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:    true,
				SecretIDNumUses: 40,
				SecretIDTTL:     "10m",
				TokenTTL:        "20m",
				TokenMaxTTL:     "30m",
				TokenPolicies:   []string{"default"},
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id":     true,
		"secret_id_num_uses": json.Number("40"),
		"secret_id_ttl":      json.Number("600"),
		"token_ttl":          json.Number("1200"),
		"token_max_ttl":      json.Number("1800"),
		"token_policies":     []any{"default"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:  true,
				TokenPolicies: []string{"default", "app-policy"},
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id": true,
		"token_policies": []any{"default", "other-policy"},
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different token_policies) to NOT be equivalent")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:  true,
				TokenPolicies: []string{"default"},
				TokenTTL:      "1h",
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id":        true,
		"token_policies":        []any{"default"},
		"token_ttl":             json.Number("3600"),
		"role_id":               "some-uuid-role-id",
		"secret_id_bound_cidrs": []any{},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra Vault fields (role_id, etc.) to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_VaultPeriodAlias(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:  true,
				TokenPolicies: []string{"default"},
				TokenPeriod:   "1h",
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id": true,
		"token_policies": []any{"default"},
		"period":         json.Number("3600"),
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected Vault 'period' alias to be treated as equivalent to 'token_period'")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_VaultPeriodAlias_Mismatch(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:  true,
				TokenPolicies: []string{"default"},
				TokenPeriod:   "1h",
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id": true,
		"token_policies": []any{"default"},
		"period":         json.Number("7200"),
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected different period value (7200 vs 3600) to NOT be equivalent")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_TokenPeriodPreferred(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:  true,
				TokenPolicies: []string{"default"},
				TokenPeriod:   "1h",
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id": true,
		"token_policies": []any{"default"},
		"token_period":   json.Number("3600"),
		"period":         json.Number("9999"),
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected token_period to take precedence over period alias when both present")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_UnorderedPolicies(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:  true,
				TokenPolicies: []string{"default", "app-policy"},
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id": true,
		"token_policies": []any{"app-policy", "default"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected order-independent comparison for token_policies")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_UnorderedCIDRs(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:       true,
				SecretIDBoundCIDRs: []string{"192.168.1.0/24", "10.0.0.0/8"},
				TokenBoundCIDRs:    []string{"172.16.0.0/12", "10.0.0.0/8"},
			},
		},
	}

	payload := map[string]any{
		"bind_secret_id":        true,
		"secret_id_bound_cidrs": []any{"10.0.0.0/8", "192.168.1.0/24"},
		"token_bound_cidrs":     []any{"10.0.0.0/8", "172.16.0.0/12"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected order-independent comparison for CIDR fields")
	}
}

func TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_IgnoresLocalSecretIDs(t *testing.T) {
	role := &AppRoleAuthEngineRole{
		Spec: AppRoleAuthEngineRoleSpec{
			AppRoleRole: AppRoleRole{
				BindSecretID:   true,
				LocalSecretIDs: true,
				TokenPolicies:  []string{"default"},
			},
		},
	}

	// Vault responds with local_secret_ids=false (or absent) — should NOT trigger drift
	// because local_secret_ids is create-only and immutable via webhook.
	payload := map[string]any{
		"bind_secret_id":   true,
		"local_secret_ids": false,
		"token_policies":   []any{"default"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected local_secret_ids difference to be ignored (field is create-only/immutable)")
	}

	// Also verify when Vault omits local_secret_ids entirely
	payloadOmitted := map[string]any{
		"bind_secret_id": true,
		"token_policies": []any{"default"},
	}

	if !role.IsEquivalentToDesiredState(payloadOmitted) {
		t.Error("expected missing local_secret_ids in Vault response to be ignored")
	}
}

func TestAppRoleAuthEngineRoleIsDeletable(t *testing.T) {
	role := &AppRoleAuthEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected AppRoleAuthEngineRole to be deletable")
	}
}

func TestAppRoleAuthEngineRoleConditions(t *testing.T) {
	role := &AppRoleAuthEngineRole{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	role.SetConditions(conditions)
	got := role.GetConditions()

	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected condition type 'ReconcileSuccessful', got %v", got[0].Type)
	}
	if got[0].Status != metav1.ConditionTrue {
		t.Errorf("expected condition status True, got %v", got[0].Status)
	}
}
