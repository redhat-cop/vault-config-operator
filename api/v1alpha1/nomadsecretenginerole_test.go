package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNomadSecretEngineRoleGetPath(t *testing.T) {
	role := &NomadSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: NomadSecretEngineRoleSpec{
			Path: "nomad",
		},
	}
	expected := "nomad/role/my-role"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestNomadSecretEngineRoleGetPathWithNameOverride(t *testing.T) {
	role := &NomadSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: NomadSecretEngineRoleSpec{
			Path: "nomad",
			Name: "custom-role",
		},
	}
	expected := "nomad/role/custom-role"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestNomadSecretEngineRoleIsDeletable(t *testing.T) {
	role := &NomadSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected NomadSecretEngineRole to be deletable")
	}
}

func TestNomadSERole_toMap(t *testing.T) {
	role := NomadSERole{
		Policies: []string{"readonly", "devops"},
		Global:   true,
		Type:     "client",
	}

	result := role.toMap()

	if result["type"] != "client" {
		t.Errorf("type = %v, expected client", result["type"])
	}
	if result["global"] != true {
		t.Errorf("global = %v, expected true", result["global"])
	}
	policies, ok := result["policies"].([]any)
	if !ok {
		t.Fatalf("policies is not []any, got %T", result["policies"])
	}
	if len(policies) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(policies))
	}
	if policies[0] != "readonly" || policies[1] != "devops" {
		t.Errorf("policies = %v, expected [readonly, devops]", policies)
	}
}

func TestNomadSERole_toMap_EmptyPolicies(t *testing.T) {
	role := NomadSERole{
		Type: "management",
	}

	result := role.toMap()

	policies, ok := result["policies"].([]any)
	if !ok {
		t.Fatalf("policies is not []any, got %T", result["policies"])
	}
	if len(policies) != 0 {
		t.Errorf("expected empty policies slice, got %v", policies)
	}
	if result["type"] != "management" {
		t.Errorf("type = %v, expected management", result["type"])
	}
}

func TestNomadSecretEngineRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	role := &NomadSecretEngineRole{
		Spec: NomadSecretEngineRoleSpec{
			Path: "nomad",
			NomadSERole: NomadSERole{
				Policies: []string{"readonly"},
				Type:     "client",
			},
		},
	}

	vaultPayload := map[string]any{
		"token_type": "client",
		"policies":   []any{"readonly"},
		"lease":      "0s",
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match (type→token_type mapped, lease filtered)")
	}
}

func TestNomadSecretEngineRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &NomadSecretEngineRole{
		Spec: NomadSecretEngineRoleSpec{
			Path: "nomad",
			NomadSERole: NomadSERole{
				Policies: []string{"readonly"},
				Type:     "client",
			},
		},
	}

	vaultPayload := map[string]any{
		"token_type": "client",
		"policies":   []any{"admin"},
		"lease":      "0s",
	}

	if role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when policies differ")
	}
}

func TestNomadSecretEngineRole_IsEquivalentToDesiredState_TypeMapping(t *testing.T) {
	role := &NomadSecretEngineRole{
		Spec: NomadSecretEngineRoleSpec{
			Path: "nomad",
			NomadSERole: NomadSERole{
				Policies: []string{"devops"},
				Type:     "client",
			},
		},
	}

	vaultPayload := map[string]any{
		"token_type": "client",
		"policies":   []any{"devops"},
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: type→token_type mapping should match")
	}
}

func TestNomadSecretEngineRole_IsEquivalentToDesiredState_ManagementType(t *testing.T) {
	role := &NomadSecretEngineRole{
		Spec: NomadSecretEngineRoleSpec{
			Path: "nomad",
			NomadSERole: NomadSERole{
				Type: "management",
			},
		},
	}

	vaultPayload := map[string]any{
		"token_type": "management",
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: management type should map to token_type=management")
	}
}

func TestNomadSecretEngineRole_IsEquivalentToDesiredState_TypeMismatch(t *testing.T) {
	role := &NomadSecretEngineRole{
		Spec: NomadSecretEngineRoleSpec{
			Path: "nomad",
			NomadSERole: NomadSERole{
				Policies: []string{"readonly"},
				Type:     "client",
			},
		},
	}

	vaultPayload := map[string]any{
		"token_type": "management",
		"policies":   []any{"readonly"},
	}

	if role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: token_type mismatch (client vs management)")
	}
}

func TestNomadSecretEngineRoleConditions(t *testing.T) {
	role := &NomadSecretEngineRole{}

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
}
