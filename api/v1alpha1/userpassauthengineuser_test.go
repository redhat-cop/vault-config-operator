package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUserpassAuthEngineUserGetPath(t *testing.T) {
	tests := []struct {
		name         string
		user         *UserpassAuthEngineUser
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			user: &UserpassAuthEngineUser{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: UserpassAuthEngineUserSpec{
					Path: "userpass",
					Name: "custom-user",
				},
			},
			expectedPath: "auth/userpass/users/custom-user",
		},
		{
			name: "without spec.name falls back to metadata.name",
			user: &UserpassAuthEngineUser{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: UserpassAuthEngineUserSpec{
					Path: "userpass",
				},
			},
			expectedPath: "auth/userpass/users/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestUserpassUserToMap(t *testing.T) {
	user := UserpassUser{
		TokenTTL:             "20m",
		TokenMaxTTL:          "30m",
		TokenPolicies:        []string{"default", "app-policy"},
		TokenBoundCIDRs:      []string{"172.16.0.0/12"},
		TokenExplicitMaxTTL:  "24h",
		TokenNoDefaultPolicy: true,
		TokenNumUses:         5,
		TokenPeriod:          "1h",
		TokenType:            "service",
		retrievedPassword:    "s3cret",
	}

	result := user.toMap()

	expected := map[string]any{
		"password":                "s3cret",
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
		t.Errorf("expected 10 keys in map (all fields set including password), got %d", len(result))
	}
}

func TestUserpassUserToMap_MinimalFields(t *testing.T) {
	user := UserpassUser{
		retrievedPassword: "onlypassword",
	}

	result := user.toMap()

	expected := map[string]any{
		"password": "onlypassword",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 key in map (only password), got %d: %v", len(result), result)
	}
}

func TestUserpassAuthEngineUserIsEquivalentToDesiredState_Match(t *testing.T) {
	user := &UserpassAuthEngineUser{
		Spec: UserpassAuthEngineUserSpec{
			UserpassUser: UserpassUser{
				TokenTTL:      "20m",
				TokenMaxTTL:   "30m",
				TokenPolicies: []string{"default"},
				TokenNumUses:  5,
			},
		},
	}

	payload := map[string]any{
		"token_ttl":      json.Number("1200"),
		"token_max_ttl":  json.Number("1800"),
		"token_policies": []any{"default"},
		"token_num_uses": json.Number("5"),
	}

	if !user.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestUserpassAuthEngineUserIsEquivalentToDesiredState_Mismatch(t *testing.T) {
	user := &UserpassAuthEngineUser{
		Spec: UserpassAuthEngineUserSpec{
			UserpassUser: UserpassUser{
				TokenPolicies: []string{"default", "app-policy"},
			},
		},
	}

	payload := map[string]any{
		"token_policies": []any{"default", "other-policy"},
	}

	if user.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different token_policies) to NOT be equivalent")
	}
}

func TestUserpassAuthEngineUserIsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	user := &UserpassAuthEngineUser{
		Spec: UserpassAuthEngineUserSpec{
			UserpassUser: UserpassUser{
				TokenPolicies: []string{"default"},
				TokenTTL:      "1h",
			},
		},
	}

	payload := map[string]any{
		"token_policies":       []any{"default"},
		"token_ttl":            json.Number("3600"),
		"some_extra_field":     "should-be-ignored",
		"another_vault_field":  json.Number("12345"),
		"token_bound_cidrs":    []any{},
		"token_explicit_extra": true,
	}

	if !user.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra Vault fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestUserpassAuthEngineUserIsEquivalentToDesiredState_PasswordStripped(t *testing.T) {
	user := &UserpassAuthEngineUser{
		Spec: UserpassAuthEngineUserSpec{
			UserpassUser: UserpassUser{
				TokenPolicies:     []string{"default"},
				TokenTTL:          "1h",
				retrievedPassword: "s3cret",
			},
		},
	}

	// Vault never returns password on read
	payload := map[string]any{
		"token_policies": []any{"default"},
		"token_ttl":      json.Number("3600"),
	}

	if !user.IsEquivalentToDesiredState(payload) {
		t.Error("expected IsEquivalentToDesiredState to strip password from desiredState and return true")
	}
}

func TestUserpassAuthEngineUserIsEquivalentToDesiredState_UnorderedPolicies(t *testing.T) {
	user := &UserpassAuthEngineUser{
		Spec: UserpassAuthEngineUserSpec{
			UserpassUser: UserpassUser{
				TokenPolicies: []string{"default", "app-policy"},
			},
		},
	}

	payload := map[string]any{
		"token_policies": []any{"app-policy", "default"},
	}

	if !user.IsEquivalentToDesiredState(payload) {
		t.Error("expected order-independent comparison for token_policies")
	}
}

func TestUserpassAuthEngineUserIsDeletable(t *testing.T) {
	user := &UserpassAuthEngineUser{}
	if !user.IsDeletable() {
		t.Error("expected UserpassAuthEngineUser to be deletable")
	}
}

func TestUserpassAuthEngineUserConditions(t *testing.T) {
	user := &UserpassAuthEngineUser{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	user.SetConditions(conditions)
	got := user.GetConditions()

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
