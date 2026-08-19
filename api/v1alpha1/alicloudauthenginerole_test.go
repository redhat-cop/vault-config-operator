package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAliCloudAuthEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *AliCloudAuthEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &AliCloudAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AliCloudAuthEngineRoleSpec{
					Path: "alicloud",
					Name: "custom-name",
				},
			},
			expectedPath: "auth/alicloud/role/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &AliCloudAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AliCloudAuthEngineRoleSpec{
					Path: "alicloud",
				},
			},
			expectedPath: "auth/alicloud/role/meta-name",
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

func TestAliCloudAuthRoleToMap(t *testing.T) {
	role := AliCloudAuthRole{
		ARN:                  "acs:ram::5138828231865461:role/dev-role",
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
		"arn":                     "acs:ram::5138828231865461:role/dev-role",
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

	if len(result) != 10 {
		t.Errorf("expected 10 keys in map (all fields set), got %d", len(result))
	}
}

func TestAliCloudAuthRoleToMap_MinimalFields(t *testing.T) {
	role := AliCloudAuthRole{
		ARN: "acs:ram::5138828231865461:role/minimal-role",
	}

	result := role.toMap()

	if len(result) != 10 {
		t.Errorf("expected 10 keys in map (unconditional emission), got %d: %v", len(result), result)
	}

	expected := map[string]any{
		"arn":                     "acs:ram::5138828231865461:role/minimal-role",
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

func TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_Match(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN:           "acs:ram::5138828231865461:role/dev-role",
				TokenTTL:      "20m",
				TokenMaxTTL:   "30m",
				TokenPolicies: []string{"default"},
			},
		},
	}

	payload := map[string]any{
		"arn":            "acs:ram::5138828231865461:role/dev-role",
		"token_ttl":      json.Number("1200"),
		"token_max_ttl":  json.Number("1800"),
		"token_policies": []any{"default"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN:           "acs:ram::5138828231865461:role/dev-role",
				TokenPolicies: []string{"default", "app-policy"},
			},
		},
	}

	payload := map[string]any{
		"arn":            "acs:ram::9999999999999999:role/other-role",
		"token_policies": []any{"default", "app-policy"},
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different arn) to NOT be equivalent")
	}
}

func TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN:           "acs:ram::5138828231865461:role/dev-role",
				TokenPolicies: []string{"default"},
				TokenTTL:      "1h",
			},
		},
	}

	payload := map[string]any{
		"arn":            "acs:ram::5138828231865461:role/dev-role",
		"token_policies": []any{"default"},
		"token_ttl":      json.Number("3600"),
		"policies":       []any{"default"},
		"ttl":            json.Number("3600"),
		"max_ttl":        json.Number("0"),
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra Vault fields (policies, ttl, max_ttl) to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_DeprecatedAliases(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN:         "acs:ram::5138828231865461:role/dev-role",
				TokenTTL:    "30m",
				TokenMaxTTL: "1h",
				TokenPeriod: "1h",
			},
		},
	}

	payload := map[string]any{
		"arn":     "acs:ram::5138828231865461:role/dev-role",
		"ttl":     json.Number("1800"),
		"max_ttl": json.Number("3600"),
		"period":  json.Number("3600"),
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected Vault deprecated aliases (ttl, max_ttl, period) to be mapped to token_* versions")
	}
}

func TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_DeprecatedAliases_Mismatch(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN:         "acs:ram::5138828231865461:role/dev-role",
				TokenTTL:    "30m",
				TokenMaxTTL: "1h",
				TokenPeriod: "1h",
			},
		},
	}

	payload := map[string]any{
		"arn":     "acs:ram::5138828231865461:role/dev-role",
		"ttl":     json.Number("1800"),
		"max_ttl": json.Number("3600"),
		"period":  json.Number("7200"),
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected different period value (7200 vs 3600) to NOT be equivalent")
	}
}

func TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_UnorderedPolicies(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN:           "acs:ram::5138828231865461:role/dev-role",
				TokenPolicies: []string{"default", "app-policy"},
			},
		},
	}

	payload := map[string]any{
		"arn":            "acs:ram::5138828231865461:role/dev-role",
		"token_policies": []any{"app-policy", "default"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected order-independent comparison for token_policies")
	}
}

func TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_DeprecatedPoliciesAlias(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN:           "acs:ram::5138828231865461:role/dev-role",
				TokenPolicies: []string{"default", "app-policy"},
			},
		},
	}

	payload := map[string]any{
		"arn":      "acs:ram::5138828231865461:role/dev-role",
		"policies": []any{"app-policy", "default"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected Vault deprecated 'policies' alias to be equivalent to desired 'token_policies'")
	}
}

func TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_DeprecatedBoundCIDRsAlias(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN:             "acs:ram::5138828231865461:role/dev-role",
				TokenBoundCIDRs: []string{"10.0.0.0/8"},
			},
		},
	}

	payload := map[string]any{
		"arn":         "acs:ram::5138828231865461:role/dev-role",
		"bound_cidrs": []any{"10.0.0.0/8"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected Vault deprecated 'bound_cidrs' alias to be equivalent to desired 'token_bound_cidrs'")
	}
}

func TestAliCloudAuthEngineRoleIsValid_ARNMatchesSpecName(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: AliCloudAuthEngineRoleSpec{
			Name: "dev-role",
			AliCloudAuthRole: AliCloudAuthRole{
				ARN: "acs:ram::5138828231865461:role/dev-role",
			},
		},
	}

	valid, err := role.IsValid()
	if !valid || err != nil {
		t.Errorf("expected valid when ARN role name matches spec.name, got valid=%v err=%v", valid, err)
	}
}

func TestAliCloudAuthEngineRoleIsValid_ARNMatchesMetadataName(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-role"},
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN: "acs:ram::5138828231865461:role/dev-role",
			},
		},
	}

	valid, err := role.IsValid()
	if !valid || err != nil {
		t.Errorf("expected valid when ARN role name matches metadata.name, got valid=%v err=%v", valid, err)
	}
}

func TestAliCloudAuthEngineRoleIsValid_ARNMismatchRejects(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN: "acs:ram::5138828231865461:role/other-role",
			},
		},
	}

	valid, err := role.IsValid()
	if valid || err == nil {
		t.Error("expected invalid when ARN role name does not match effective Vault role name")
	}
}

func TestAliCloudAuthEngineRoleIsValid_ARNMismatchWithSpecName(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: AliCloudAuthEngineRoleSpec{
			Name: "custom-name",
			AliCloudAuthRole: AliCloudAuthRole{
				ARN: "acs:ram::5138828231865461:role/wrong-name",
			},
		},
	}

	valid, err := role.IsValid()
	if valid || err == nil {
		t.Error("expected invalid when ARN role name does not match spec.name")
	}
}

func TestAliCloudAuthEngineRoleIsValid_CaseInsensitiveMatch(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-role"},
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN: "acs:ram::5138828231865461:role/Dev-Role",
			},
		},
	}

	valid, err := role.IsValid()
	if !valid || err != nil {
		t.Errorf("expected case-insensitive match to be valid, got valid=%v err=%v", valid, err)
	}
}

func TestAliCloudAuthEngineRoleIsValid_NoRoleSegmentInARN(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-role"},
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN: "acs:ram::5138828231865461:policy/dev-role",
			},
		},
	}

	valid, err := role.IsValid()
	if valid || err == nil {
		t.Error("expected invalid when ARN has no role/ segment (policy ARN is not a role ARN)")
	}
}

func TestAliCloudAuthEngineRoleIsValid_EmptyRoleNameAfterPrefix(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-role"},
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN: "acs:ram::5138828231865461:role/",
			},
		},
	}

	valid, err := role.IsValid()
	if valid || err == nil {
		t.Error("expected invalid when ARN has role/ prefix but empty role name")
	}
}

func TestExtractAliCloudARNRoleName(t *testing.T) {
	tests := []struct {
		arn      string
		wantName string
		wantOK   bool
	}{
		{"acs:ram::5138828231865461:role/dev-role", "dev-role", true},
		{"acs:ram::5138828231865461:role/MyRole", "MyRole", true},
		{"acs:ram::5138828231865461:policy/dev-role", "", false},
		{"acs:ram::5138828231865461:role/", "", false},
		{"", "", false},
		// Policy ARN that contains "role/" in a path segment must be rejected
		{"acs:ram::5138828231865461:policy/team/role/dev-role", "", false},
	}

	for _, tt := range tests {
		name, ok := extractAliCloudARNRoleName(tt.arn)
		if name != tt.wantName || ok != tt.wantOK {
			t.Errorf("extractAliCloudARNRoleName(%q) = (%q, %v), want (%q, %v)",
				tt.arn, name, ok, tt.wantName, tt.wantOK)
		}
	}
}

func TestAliCloudAuthEngineRoleIsValid_PolicyARNWithRoleInPath(t *testing.T) {
	role := &AliCloudAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-role"},
		Spec: AliCloudAuthEngineRoleSpec{
			AliCloudAuthRole: AliCloudAuthRole{
				ARN: "acs:ram::5138828231865461:policy/team/role/dev-role",
			},
		},
	}

	valid, err := role.IsValid()
	if valid || err == nil {
		t.Error("expected invalid when policy ARN contains 'role/' in a path segment (false positive)")
	}
}

func TestAliCloudAuthEngineRoleIsDeletable(t *testing.T) {
	role := &AliCloudAuthEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected AliCloudAuthEngineRole to be deletable")
	}
}

func TestAliCloudAuthEngineRoleConditions(t *testing.T) {
	role := &AliCloudAuthEngineRole{}

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
