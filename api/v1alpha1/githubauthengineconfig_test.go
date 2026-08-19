package v1alpha1

import (
	"encoding/json"
	"testing"
)

func TestGitHubAuthEngineConfig_toMap_AllFields(t *testing.T) {
	config := &GitHubAuthConfig{
		Organization:         "acme-org",
		OrganizationID:       12345,
		BaseURL:              "https://github.example.com/api/v3",
		TokenTTL:             "1h",
		TokenMaxTTL:          "4h",
		TokenPolicies:        []string{"default", "dev-policy"},
		TokenBoundCIDRs:      []string{"10.0.0.0/8", "192.168.1.0/24"},
		TokenExplicitMaxTTL:  "8h",
		TokenNoDefaultPolicy: true,
		TokenNumUses:         5,
		TokenPeriod:          "2h",
		TokenType:            "service",
	}

	result := config.toMap()

	if result["organization"] != "acme-org" {
		t.Errorf("expected organization=acme-org, got %v", result["organization"])
	}
	if result["organization_id"] != json.Number("12345") {
		t.Errorf("expected organization_id=json.Number(12345), got %v (type %T)", result["organization_id"], result["organization_id"])
	}
	if result["base_url"] != "https://github.example.com/api/v3" {
		t.Errorf("expected base_url=https://github.example.com/api/v3, got %v", result["base_url"])
	}
	if result["token_ttl"] != json.Number("3600") {
		t.Errorf("expected token_ttl=json.Number(3600), got %v (type %T)", result["token_ttl"], result["token_ttl"])
	}
	if result["token_max_ttl"] != json.Number("14400") {
		t.Errorf("expected token_max_ttl=json.Number(14400), got %v (type %T)", result["token_max_ttl"], result["token_max_ttl"])
	}

	policies, ok := result["token_policies"].([]any)
	if !ok {
		t.Fatalf("expected token_policies to be []any, got %T", result["token_policies"])
	}
	if len(policies) != 2 || policies[0] != "default" || policies[1] != "dev-policy" {
		t.Errorf("expected token_policies=[default, dev-policy], got %v", policies)
	}

	cidrs, ok := result["token_bound_cidrs"].([]any)
	if !ok {
		t.Fatalf("expected token_bound_cidrs to be []any, got %T", result["token_bound_cidrs"])
	}
	if len(cidrs) != 2 || cidrs[0] != "10.0.0.0/8" || cidrs[1] != "192.168.1.0/24" {
		t.Errorf("expected token_bound_cidrs=[10.0.0.0/8, 192.168.1.0/24], got %v", cidrs)
	}

	if result["token_explicit_max_ttl"] != json.Number("28800") {
		t.Errorf("expected token_explicit_max_ttl=json.Number(28800), got %v (type %T)", result["token_explicit_max_ttl"], result["token_explicit_max_ttl"])
	}
	if result["token_no_default_policy"] != true {
		t.Errorf("expected token_no_default_policy=true, got %v", result["token_no_default_policy"])
	}
	if result["token_num_uses"] != json.Number("5") {
		t.Errorf("expected token_num_uses=json.Number(5), got %v (type %T)", result["token_num_uses"], result["token_num_uses"])
	}
	if result["token_period"] != json.Number("7200") {
		t.Errorf("expected token_period=json.Number(7200), got %v (type %T)", result["token_period"], result["token_period"])
	}
	if result["token_type"] != "service" {
		t.Errorf("expected token_type=service, got %v", result["token_type"])
	}
}

func TestGitHubAuthEngineConfig_toMap_MinimalFields(t *testing.T) {
	config := &GitHubAuthConfig{
		Organization: "minimal-org",
	}

	result := config.toMap()

	if result["organization"] != "minimal-org" {
		t.Errorf("expected organization=minimal-org, got %v", result["organization"])
	}

	if _, exists := result["organization_id"]; exists {
		t.Error("expected organization_id to not be present when zero")
	}
	if _, exists := result["base_url"]; exists {
		t.Error("expected base_url to not be present when empty")
	}
	if _, exists := result["token_ttl"]; exists {
		t.Error("expected token_ttl to not be present when empty")
	}
	if _, exists := result["token_max_ttl"]; exists {
		t.Error("expected token_max_ttl to not be present when empty")
	}
	if _, exists := result["token_policies"]; exists {
		t.Error("expected token_policies to not be present when empty")
	}
	if _, exists := result["token_bound_cidrs"]; exists {
		t.Error("expected token_bound_cidrs to not be present when empty")
	}
	if _, exists := result["token_explicit_max_ttl"]; exists {
		t.Error("expected token_explicit_max_ttl to not be present when empty")
	}
	if _, exists := result["token_no_default_policy"]; exists {
		t.Error("expected token_no_default_policy to not be present when false")
	}
	if _, exists := result["token_num_uses"]; exists {
		t.Error("expected token_num_uses to not be present when zero")
	}
	if _, exists := result["token_period"]; exists {
		t.Error("expected token_period to not be present when empty")
	}
	if _, exists := result["token_type"]; exists {
		t.Error("expected token_type to not be present when empty")
	}

	if len(result) != 1 {
		t.Errorf("expected exactly 1 key in minimal map, got %d keys: %v", len(result), result)
	}
}

func TestGitHubAuthEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &GitHubAuthEngineConfig{}
	instance.Spec.GitHubAuthConfig = GitHubAuthConfig{
		Organization:   "acme-org",
		OrganizationID: 12345,
		TokenTTL:       "1h",
		TokenPolicies:  []string{"default", "dev-policy"},
	}

	vaultPayload := map[string]any{
		"organization":    "acme-org",
		"organization_id": json.Number("12345"),
		"token_ttl":       json.Number("3600"),
		"token_policies":  []any{"default", "dev-policy"},
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching payload")
	}
}

func TestGitHubAuthEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &GitHubAuthEngineConfig{}
	instance.Spec.GitHubAuthConfig = GitHubAuthConfig{
		Organization: "acme-org",
		TokenTTL:     "1h",
	}

	vaultPayload := map[string]any{
		"organization": "different-org",
		"token_ttl":    json.Number("3600"),
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched organization")
	}
}

func TestGitHubAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &GitHubAuthEngineConfig{}
	instance.Spec.GitHubAuthConfig = GitHubAuthConfig{
		Organization: "acme-org",
	}

	vaultPayload := map[string]any{
		"organization":    "acme-org",
		"organization_id": json.Number("99999"),
		"token_ttl":       json.Number("0"),
		"token_max_ttl":   json.Number("0"),
		"token_policies":  []any{},
		"request_id":      "some-vault-id",
		"lease_duration":  json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields not in desired state")
	}
}

func TestGitHubAuthEngineConfig_GetPath(t *testing.T) {
	instance := &GitHubAuthEngineConfig{}
	instance.Spec.Path = "github"

	expected := "auth/github/config"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestGitHubAuthEngineConfig_IsDeletable(t *testing.T) {
	instance := &GitHubAuthEngineConfig{}
	if instance.IsDeletable() {
		t.Error("expected IsDeletable() to return false for config type")
	}
}
