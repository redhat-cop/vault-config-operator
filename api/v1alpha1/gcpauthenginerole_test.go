package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGCPAuthEngineRoleGetPath(t *testing.T) {
	role := &GCPAuthEngineRole{
		Spec: GCPAuthEngineRoleSpec{
			Path: "gcp",
			GCPRole: GCPRole{
				Name: "my-role",
			},
		},
	}

	result := role.GetPath()
	expected := "auth/gcp/role/my-role"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestGCPRoleToMap(t *testing.T) {
	role := GCPRole{
		Name:                 "test-role",
		Type:                 "iam",
		BoundServiceAccounts: []string{"sa@project.iam.gserviceaccount.com"},
		BoundProjects:        []string{"project-1"},
		AddGroupAliases:      true,
		TokenTTL:             "1h",
		TokenMaxTTL:          "24h",
		TokenPolicies:        []string{"reader"},
		Policies:             []string{"legacy-policy"},
		TokenBoundCIDRs:      []string{"10.0.0.0/8"},
		TokenExplicitMaxTTL:  "",
		TokenNoDefaultPolicy: false,
		TokenNumUses:         0,
		TokenPeriod:          0,
		TokenType:            "service",
		MaxJWTExp:            "15m",
		AllowGCEInference:    true,
		BoundZones:           []string{"us-central1-a"},
		BoundRegions:         []string{"us-central1"},
		BoundInstanceGroups:  []string{"ig-1"},
		BoundLabels:          []string{"env:production"},
	}

	result := role.toMap()

	if len(result) != 21 {
		t.Errorf("expected 21 keys in map, got %d", len(result))
	}

	expected := map[string]interface{}{
		"name":                    "test-role",
		"type":                    "iam",
		"bound_service_accounts":  []string{"sa@project.iam.gserviceaccount.com"},
		"bound_projects":          []string{"project-1"},
		"add_group_aliases":       true,
		"token_ttl":               "1h",
		"token_max_ttl":           "24h",
		"token_policies":          []string{"reader"},
		"policies":                []string{"legacy-policy"},
		"token_bound_cidrs":       []string{"10.0.0.0/8"},
		"token_explicit_max_ttl":  "",
		"token_no_default_policy": false,
		"token_num_uses":          int64(0),
		"token_period":            int64(0),
		"token_type":              "service",
		"max_jwt_exp":             "15m",
		"allow_gce_inference":     true,
		"bound_zones":             []string{"us-central1-a"},
		"bound_regions":           []string{"us-central1"},
		"bound_instance_groups":   []string{"ig-1"},
		"bound_labels":            []string{"env:production"},
	}

	if !reflect.DeepEqual(result, expected) {
		for k, v := range expected {
			if !reflect.DeepEqual(result[k], v) {
				t.Errorf("key %q: got %v (%T), want %v (%T)", k, result[k], result[k], v, v)
			}
		}
	}
}

func TestGCPRoleToMapDualPoliciesField(t *testing.T) {
	role := GCPRole{
		Name:          "test-role",
		Type:          "iam",
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

func TestGCPAuthEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &GCPAuthEngineRole{
		Spec: GCPAuthEngineRoleSpec{
			GCPRole: GCPRole{
				Name:      "test-role",
				Type:      "iam",
				TokenType: "service",
			},
		},
	}

	payload := role.Spec.GCPRole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestGCPAuthEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &GCPAuthEngineRole{
		Spec: GCPAuthEngineRoleSpec{
			GCPRole: GCPRole{
				Name: "test-role",
				Type: "iam",
			},
		},
	}

	payload := role.Spec.GCPRole.toMap()
	payload["type"] = "gce"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different type) to NOT be equivalent")
	}
}

func TestGCPAuthEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &GCPAuthEngineRole{
		Spec: GCPAuthEngineRoleSpec{
			GCPRole: GCPRole{
				Name: "test-role",
				Type: "iam",
			},
		},
	}

	payload := role.Spec.GCPRole.toMap()
	payload["extra_field"] = "unexpected"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
	}
}

func TestGCPAuthEngineRoleIsDeletable(t *testing.T) {
	role := &GCPAuthEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected GCPAuthEngineRole to be deletable")
	}
}

func TestGCPAuthEngineRoleConditions(t *testing.T) {
	role := &GCPAuthEngineRole{}

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
