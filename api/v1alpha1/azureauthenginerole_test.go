package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAzureAuthEngineRoleGetPath(t *testing.T) {
	role := &AzureAuthEngineRole{
		Spec: AzureAuthEngineRoleSpec{
			Path: "azure",
			AzureRole: AzureRole{
				Name: "my-role",
			},
		},
	}

	result := role.GetPath()
	expected := "auth/azure/role/my-role"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestAzureRoleToMap(t *testing.T) {
	role := AzureRole{
		Name:                     "test-role",
		BoundServicePrincipalIDs: []string{"sp-1", "sp-2"},
		BoundGroupIDs:            []string{"group-1"},
		BoundLocations:           []string{"eastus"},
		BoundSubscriptionIDs:     []string{"sub-1"},
		BoundResourceGroups:      []string{"rg-1"},
		BoundScaleSets:           []string{"ss-1"},
		TokenTTL:                 "1h",
		TokenMaxTTL:              "24h",
		TokenPolicies:            []string{"reader"},
		Policies:                 []string{"legacy-policy"},
		TokenBoundCIDRs:          []string{"10.0.0.0/8"},
		TokenExplicitMaxTTL:      "",
		TokenNoDefaultPolicy:     false,
		TokenNumUses:             0,
		TokenPeriod:              0,
		TokenType:                "service",
	}

	result := role.toMap()

	if len(result) != 17 {
		t.Errorf("expected 17 keys in map, got %d", len(result))
	}

	expected := map[string]interface{}{
		"name":                        "test-role",
		"bound_service_principal_ids": []string{"sp-1", "sp-2"},
		"bound_group_ids":             []string{"group-1"},
		"bound_locations":             []string{"eastus"},
		"bound_subscription_ids":      []string{"sub-1"},
		"bound_resource_groups":       []string{"rg-1"},
		"bound_scale_sets":            []string{"ss-1"},
		"token_ttl":                   "1h",
		"token_max_ttl":               "24h",
		"token_policies":              []string{"reader"},
		"policies":                    []string{"legacy-policy"},
		"token_bound_cidrs":           []string{"10.0.0.0/8"},
		"token_explicit_max_ttl":      "",
		"token_no_default_policy":     false,
		"token_num_uses":              int64(0),
		"token_period":                int64(0),
		"token_type":                  "service",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestAzureRoleToMapDualPoliciesField(t *testing.T) {
	role := AzureRole{
		Name:          "test-role",
		TokenPolicies: []string{"token-policy"},
		Policies:      []string{"legacy-policy"},
	}

	result := role.toMap()

	if _, exists := result["token_policies"]; !exists {
		t.Error("expected 'token_policies' key to exist")
	}
	if _, exists := result["policies"]; !exists {
		t.Error("expected 'policies' key to exist (legacy/backward-compatibility field)")
	}

	if !reflect.DeepEqual(result["token_policies"], []string{"token-policy"}) {
		t.Errorf("expected token_policies = [token-policy], got %v", result["token_policies"])
	}
	if !reflect.DeepEqual(result["policies"], []string{"legacy-policy"}) {
		t.Errorf("expected policies = [legacy-policy], got %v", result["policies"])
	}
}

func TestAzureAuthEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &AzureAuthEngineRole{
		Spec: AzureAuthEngineRoleSpec{
			AzureRole: AzureRole{
				Name:          "test-role",
				TokenPolicies: []string{"reader"},
				Policies:      []string{"legacy"},
				TokenType:     "service",
			},
		},
	}

	payload := role.Spec.AzureRole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAzureAuthEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &AzureAuthEngineRole{
		Spec: AzureAuthEngineRoleSpec{
			AzureRole: AzureRole{
				Name:      "test-role",
				TokenType: "service",
			},
		},
	}

	payload := role.Spec.AzureRole.toMap()
	payload["name"] = "different-role"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different name) to NOT be equivalent")
	}
}

func TestAzureAuthEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &AzureAuthEngineRole{
		Spec: AzureAuthEngineRoleSpec{
			AzureRole: AzureRole{
				Name: "test-role",
			},
		},
	}

	payload := role.Spec.AzureRole.toMap()
	payload["extra_field"] = "unexpected"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
	}
}

func TestAzureAuthEngineRoleIsDeletable(t *testing.T) {
	role := &AzureAuthEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected AzureAuthEngineRole to be deletable")
	}
}

func TestAzureAuthEngineRoleConditions(t *testing.T) {
	role := &AzureAuthEngineRole{}

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
