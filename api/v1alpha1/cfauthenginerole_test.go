package v1alpha1

import (
	"encoding/json"
	"testing"
)

func TestCFAuthEngineRole_toMap(t *testing.T) {
	role := &CFAuthRole{
		BoundApplicationIDs:  []string{"09d7eb6a-afc2-49a0-bb32-858c22f2b346"},
		BoundSpaceIDs:        []string{"21005ebb-8943-433e-84e6-d9d9d7338853"},
		BoundOrganizationIDs: []string{"9785a884-5e93-49bd-97ee-57bf7c2b20e0"},
		BoundInstanceIDs:     []string{"f3e0f176-3f83-4efb-5842-2ff4"},
		DisableIPMatching:    true,
		TokenTTL:             "1h",
		TokenMaxTTL:          "4h",
		TokenPolicies:        []string{"default", "cf-policy"},
		Policies:             []string{"legacy-policy"},
		TokenBoundCIDRs:      []string{"127.0.0.1/32"},
		TokenExplicitMaxTTL:  "8h",
		TokenNoDefaultPolicy: true,
		TokenNumUses:         5,
		TokenPeriod:          "24h",
		TokenType:            "service",
	}

	result := role.toMap()

	appIDs, ok := result["bound_application_ids"].([]any)
	if !ok {
		t.Fatalf("expected bound_application_ids to be []any, got %T", result["bound_application_ids"])
	}
	if len(appIDs) != 1 || appIDs[0] != "09d7eb6a-afc2-49a0-bb32-858c22f2b346" {
		t.Errorf("unexpected bound_application_ids: %v", appIDs)
	}

	spaceIDs, ok := result["bound_space_ids"].([]any)
	if !ok {
		t.Fatalf("expected bound_space_ids to be []any, got %T", result["bound_space_ids"])
	}
	if len(spaceIDs) != 1 || spaceIDs[0] != "21005ebb-8943-433e-84e6-d9d9d7338853" {
		t.Errorf("unexpected bound_space_ids: %v", spaceIDs)
	}

	orgIDs, ok := result["bound_organization_ids"].([]any)
	if !ok {
		t.Fatalf("expected bound_organization_ids to be []any, got %T", result["bound_organization_ids"])
	}
	if len(orgIDs) != 1 {
		t.Errorf("expected 1 org ID, got %d", len(orgIDs))
	}

	instanceIDs, ok := result["bound_instance_ids"].([]any)
	if !ok {
		t.Fatalf("expected bound_instance_ids to be []any, got %T", result["bound_instance_ids"])
	}
	if len(instanceIDs) != 1 {
		t.Errorf("expected 1 instance ID, got %d", len(instanceIDs))
	}

	if result["disable_ip_matching"] != true {
		t.Errorf("expected disable_ip_matching=true, got %v", result["disable_ip_matching"])
	}
	if result["token_ttl"] != json.Number("3600") {
		t.Errorf("expected token_ttl=3600, got %v", result["token_ttl"])
	}
	if result["token_max_ttl"] != json.Number("14400") {
		t.Errorf("expected token_max_ttl=14400, got %v", result["token_max_ttl"])
	}

	policies, ok := result["token_policies"].([]any)
	if !ok {
		t.Fatalf("expected token_policies to be []any, got %T", result["token_policies"])
	}
	if len(policies) != 2 || policies[0] != "default" || policies[1] != "cf-policy" {
		t.Errorf("expected token_policies=[default cf-policy], got %v", policies)
	}

	legacyPolicies, ok := result["policies"].([]any)
	if !ok {
		t.Fatalf("expected policies to be []any, got %T", result["policies"])
	}
	if len(legacyPolicies) != 1 || legacyPolicies[0] != "legacy-policy" {
		t.Errorf("expected policies=[legacy-policy], got %v", legacyPolicies)
	}

	cidrs, ok := result["token_bound_cidrs"].([]any)
	if !ok {
		t.Fatalf("expected token_bound_cidrs to be []any, got %T", result["token_bound_cidrs"])
	}
	if len(cidrs) != 1 || cidrs[0] != "127.0.0.1/32" {
		t.Errorf("expected token_bound_cidrs=[127.0.0.1/32], got %v", cidrs)
	}

	if result["token_explicit_max_ttl"] != json.Number("28800") {
		t.Errorf("expected token_explicit_max_ttl=28800, got %v", result["token_explicit_max_ttl"])
	}
	if result["token_no_default_policy"] != true {
		t.Errorf("expected token_no_default_policy=true, got %v", result["token_no_default_policy"])
	}
	if result["token_num_uses"] != json.Number("5") {
		t.Errorf("expected token_num_uses=json.Number(5), got %v (type %T)", result["token_num_uses"], result["token_num_uses"])
	}
	if result["token_period"] != json.Number("86400") {
		t.Errorf("expected token_period=86400, got %v", result["token_period"])
	}
	if result["token_type"] != "service" {
		t.Errorf("expected token_type=service, got %v", result["token_type"])
	}
}

func TestCFAuthEngineRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &CFAuthEngineRole{}
	instance.Spec.CFAuthRole = CFAuthRole{
		BoundApplicationIDs:  []string{"09d7eb6a-afc2-49a0-bb32-858c22f2b346"},
		BoundSpaceIDs:        []string{"21005ebb-8943-433e-84e6-d9d9d7338853"},
		BoundOrganizationIDs: []string{"9785a884-5e93-49bd-97ee-57bf7c2b20e0"},
		BoundInstanceIDs:     []string{"f3e0f176-3f83-4efb-5842-2ff4"},
		DisableIPMatching:    false,
		TokenTTL:             "1h",
		TokenMaxTTL:          "1h",
		TokenPolicies:        []string{"default"},
		TokenNumUses:         0,
		TokenType:            "service",
	}

	// Vault-read-shaped payload: Vault returns "ttl"/"max_ttl"/"period"/"bound_cidrs"
	// instead of "token_ttl"/"token_max_ttl"/"token_period"/"token_bound_cidrs"
	vaultPayload := map[string]any{
		"bound_application_ids":   []any{"09d7eb6a-afc2-49a0-bb32-858c22f2b346"},
		"bound_space_ids":         []any{"21005ebb-8943-433e-84e6-d9d9d7338853"},
		"bound_organization_ids":  []any{"9785a884-5e93-49bd-97ee-57bf7c2b20e0"},
		"bound_instance_ids":      []any{"f3e0f176-3f83-4efb-5842-2ff4"},
		"disable_ip_matching":     false,
		"ttl":                     json.Number("3600"),
		"max_ttl":                 json.Number("3600"),
		"token_policies":          []any{"default"},
		"policies":                []any{},
		"bound_cidrs":             []any{},
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"period":                  json.Number("0"),
		"token_type":              "service",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state (Vault read aliases normalized)")
	}
}

func TestCFAuthEngineRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &CFAuthEngineRole{}
	instance.Spec.CFAuthRole = CFAuthRole{
		BoundApplicationIDs: []string{"09d7eb6a-afc2-49a0-bb32-858c22f2b346"},
		BoundSpaceIDs:       []string{"21005ebb-8943-433e-84e6-d9d9d7338853"},
		TokenTTL:            "1h",
		TokenType:           "service",
	}

	// Vault-read-shaped payload with read aliases
	vaultPayload := map[string]any{
		"bound_application_ids": []any{"DIFFERENT-APP-ID"},
		"bound_space_ids":       []any{"21005ebb-8943-433e-84e6-d9d9d7338853"},
		"ttl":                   json.Number("3600"),
		"max_ttl":               json.Number("0"),
		"token_type":            "service",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched bound_application_ids")
	}
}

func TestCFAuthEngineRole_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &CFAuthEngineRole{}
	instance.Spec.CFAuthRole = CFAuthRole{
		BoundApplicationIDs: []string{"app-id-1"},
		TokenTTL:            "1h",
		TokenType:           "service",
	}

	// Vault-read-shaped payload with read aliases and extra Vault metadata
	vaultPayload := map[string]any{
		"bound_application_ids":   []any{"app-id-1"},
		"bound_space_ids":         []any{},
		"bound_organization_ids":  []any{},
		"bound_instance_ids":      []any{},
		"disable_ip_matching":     false,
		"ttl":                     json.Number("3600"),
		"max_ttl":                 json.Number("0"),
		"token_policies":          []any{},
		"policies":                []any{},
		"bound_cidrs":             []any{},
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"period":                  json.Number("0"),
		"token_type":              "service",
		"request_id":              "extra-vault-field",
		"lease_duration":          json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestCFAuthEngineRole_GetPath_WithName(t *testing.T) {
	instance := &CFAuthEngineRole{}
	instance.Spec.Path = "cf"
	instance.Spec.Name = "my-role"

	expected := "auth/cf/roles/my-role"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestCFAuthEngineRole_GetPath_WithMetadataName(t *testing.T) {
	instance := &CFAuthEngineRole{}
	instance.Spec.Path = "cf"
	instance.Name = "metadata-role"

	expected := "auth/cf/roles/metadata-role"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestCFAuthEngineRole_IsEquivalentToDesiredState_DifferentOrderEquivalent(t *testing.T) {
	instance := &CFAuthEngineRole{}
	instance.Spec.CFAuthRole = CFAuthRole{
		BoundApplicationIDs:  []string{"app-z", "app-a", "app-m"},
		BoundSpaceIDs:        []string{"space-2", "space-1"},
		BoundOrganizationIDs: []string{"org-b", "org-a"},
		TokenPolicies:        []string{"policy-z", "policy-a"},
		TokenBoundCIDRs:      []string{"10.0.0.0/8", "192.168.0.0/16"},
		TokenTTL:             "1h",
		TokenType:            "service",
	}

	vaultPayload := map[string]any{
		"bound_application_ids":   []any{"app-a", "app-m", "app-z"},
		"bound_space_ids":         []any{"space-1", "space-2"},
		"bound_organization_ids":  []any{"org-a", "org-b"},
		"bound_instance_ids":      []any{},
		"disable_ip_matching":     false,
		"ttl":                     json.Number("3600"),
		"max_ttl":                 json.Number("0"),
		"token_policies":          []any{"policy-a", "policy-z"},
		"policies":                []any{},
		"bound_cidrs":             []any{"192.168.0.0/16", "10.0.0.0/8"},
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"period":                  json.Number("0"),
		"token_type":              "service",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when set-like fields differ only in order")
	}
}

func TestCFAuthEngineRole_IsDeletable(t *testing.T) {
	instance := &CFAuthEngineRole{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true for CF auth engine role")
	}
}
