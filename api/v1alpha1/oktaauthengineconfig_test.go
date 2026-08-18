package v1alpha1

import (
	"context"
	"encoding/json"
	"testing"
)

func TestOktaAuthEngineConfig_toMap(t *testing.T) {
	config := &OktaAuthConfig{
		OrgName:              "my-org",
		BaseURL:              "okta.com",
		BypassOktaMFA:        true,
		TokenTTL:             "1h",
		TokenMaxTTL:          "4h",
		TokenPolicies:        []string{"default", "admin"},
		TokenBoundCIDRs:      []string{"10.0.0.0/8"},
		TokenExplicitMaxTTL:  "8h",
		TokenNoDefaultPolicy: true,
		TokenNumUses:         5,
		TokenPeriod:          "24h",
		TokenType:            "service",
		retrievedAPIToken:    "okta-api-token-value",
	}

	result := config.toMap()

	if result["org_name"] != "my-org" {
		t.Errorf("expected org_name=my-org, got %v", result["org_name"])
	}
	if result["api_token"] != "okta-api-token-value" {
		t.Errorf("expected api_token=okta-api-token-value, got %v", result["api_token"])
	}
	if result["base_url"] != "okta.com" {
		t.Errorf("expected base_url=okta.com, got %v", result["base_url"])
	}
	if result["bypass_okta_mfa"] != true {
		t.Errorf("expected bypass_okta_mfa=true, got %v", result["bypass_okta_mfa"])
	}
	if result["token_ttl"] != "1h" {
		t.Errorf("expected token_ttl=1h, got %v", result["token_ttl"])
	}
	if result["token_max_ttl"] != "4h" {
		t.Errorf("expected token_max_ttl=4h, got %v", result["token_max_ttl"])
	}

	policies, ok := result["token_policies"].([]any)
	if !ok {
		t.Fatalf("expected token_policies to be []any, got %T", result["token_policies"])
	}
	if len(policies) != 2 || policies[0] != "default" || policies[1] != "admin" {
		t.Errorf("expected token_policies=[default admin], got %v", policies)
	}

	cidrs, ok := result["token_bound_cidrs"].([]any)
	if !ok {
		t.Fatalf("expected token_bound_cidrs to be []any, got %T", result["token_bound_cidrs"])
	}
	if len(cidrs) != 1 || cidrs[0] != "10.0.0.0/8" {
		t.Errorf("expected token_bound_cidrs=[10.0.0.0/8], got %v", cidrs)
	}

	if result["token_explicit_max_ttl"] != "8h" {
		t.Errorf("expected token_explicit_max_ttl=8h, got %v", result["token_explicit_max_ttl"])
	}
	if result["token_no_default_policy"] != true {
		t.Errorf("expected token_no_default_policy=true, got %v", result["token_no_default_policy"])
	}
	if result["token_num_uses"] != json.Number("5") {
		t.Errorf("expected token_num_uses=json.Number(5), got %v (type %T)", result["token_num_uses"], result["token_num_uses"])
	}
	if result["token_period"] != "24h" {
		t.Errorf("expected token_period=24h, got %v", result["token_period"])
	}
	if result["token_type"] != "service" {
		t.Errorf("expected token_type=service, got %v", result["token_type"])
	}
}

func TestOktaAuthEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &OktaAuthEngineConfig{}
	instance.Spec.OktaAuthConfig = OktaAuthConfig{
		OrgName:              "my-org",
		BaseURL:              "okta.com",
		BypassOktaMFA:        false,
		TokenTTL:             "",
		TokenMaxTTL:          "",
		TokenPolicies:        nil,
		TokenBoundCIDRs:      nil,
		TokenExplicitMaxTTL:  "",
		TokenNoDefaultPolicy: false,
		TokenNumUses:         0,
		TokenPeriod:          "",
		TokenType:            "",
		retrievedAPIToken:    "secret-token-value",
	}

	vaultPayload := map[string]any{
		"org_name":                "my-org",
		"base_url":                "okta.com",
		"bypass_okta_mfa":         false,
		"token_ttl":               "",
		"token_max_ttl":           "",
		"token_policies":          []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  "",
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            "",
		"token_type":              "",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state (api_token stripped)")
	}
}

func TestOktaAuthEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &OktaAuthEngineConfig{}
	instance.Spec.OktaAuthConfig = OktaAuthConfig{
		OrgName:           "my-org",
		BaseURL:           "okta.com",
		retrievedAPIToken: "token",
	}

	vaultPayload := map[string]any{
		"org_name":                "different-org",
		"base_url":                "okta.com",
		"bypass_okta_mfa":         false,
		"token_ttl":               "",
		"token_max_ttl":           "",
		"token_policies":          []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  "",
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            "",
		"token_type":              "",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched org_name")
	}
}

func TestOktaAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &OktaAuthEngineConfig{}
	instance.Spec.OktaAuthConfig = OktaAuthConfig{
		OrgName:           "my-org",
		BaseURL:           "okta.com",
		retrievedAPIToken: "token",
	}

	vaultPayload := map[string]any{
		"org_name":                "my-org",
		"base_url":                "okta.com",
		"bypass_okta_mfa":         false,
		"token_ttl":               "",
		"token_max_ttl":           "",
		"token_policies":          []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  "",
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            "",
		"token_type":              "",
		"request_id":              "extra-vault-field",
		"lease_duration":          json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestOktaAuthEngineConfig_IsEquivalentToDesiredState_APITokenStripping(t *testing.T) {
	instance := &OktaAuthEngineConfig{}
	instance.Spec.OktaAuthConfig = OktaAuthConfig{
		OrgName:           "my-org",
		BaseURL:           "okta.com",
		retrievedAPIToken: "super-secret-token",
	}

	vaultPayload := map[string]any{
		"org_name":                "my-org",
		"base_url":                "okta.com",
		"bypass_okta_mfa":         false,
		"token_ttl":               "",
		"token_max_ttl":           "",
		"token_policies":          []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  "",
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            "",
		"token_type":              "",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: api_token is write-only and must be excluded from drift comparison")
	}
}

func TestOktaAuthEngineConfig_GetPath(t *testing.T) {
	instance := &OktaAuthEngineConfig{}
	instance.Spec.Path = "okta"

	expected := "auth/okta/config"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestOktaAuthEngineConfig_IsDeletable(t *testing.T) {
	instance := &OktaAuthEngineConfig{}
	if instance.IsDeletable() {
		t.Error("expected IsDeletable() to return false for auth engine config")
	}
}

func TestOktaAuthEngineConfig_Default_EmptyPasswordKey(t *testing.T) {
	webhook := &OktaAuthEngineConfig{}
	obj := &OktaAuthEngineConfig{}
	obj.Spec.OktaCredentials.PasswordKey = ""

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.OktaCredentials.PasswordKey != "api_token" {
		t.Errorf("expected passwordKey=api_token, got %s", obj.Spec.OktaCredentials.PasswordKey)
	}
}

func TestOktaAuthEngineConfig_Default_PreservesCustomPasswordKey(t *testing.T) {
	webhook := &OktaAuthEngineConfig{}
	obj := &OktaAuthEngineConfig{}
	obj.Spec.OktaCredentials.PasswordKey = "my_custom_key"

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.OktaCredentials.PasswordKey != "my_custom_key" {
		t.Errorf("expected custom passwordKey to be preserved, got %s", obj.Spec.OktaCredentials.PasswordKey)
	}
}
