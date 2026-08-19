package v1alpha1

import (
	"encoding/json"
	"testing"
)

func TestOCIAuthEngineRole_toMap(t *testing.T) {
	role := &OCIAuthRole{
		Name: "my-role",
		OCIDList: []string{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
			"ocid1.dynamicgroup.oc1..aaaaaaaa5hmfyrdaxvmt52ekju5n7ffamn2pdvxaq6esb2vzzoduexamplea",
		},
		TokenTTL:             "30m",
		TokenMaxTTL:          "1h",
		TokenPolicies:        []string{"dev", "prod"},
		Policies:             []string{"legacy-policy"},
		TokenBoundCIDRs:      []string{"10.0.0.0/8"},
		TokenExplicitMaxTTL:  "2h",
		TokenNoDefaultPolicy: true,
		TokenNumUses:         5,
		TokenPeriod:          3600,
		TokenType:            "service",
	}

	result := role.toMap()

	ocidList, ok := result["ocid_list"].([]any)
	if !ok {
		t.Fatalf("expected ocid_list to be []any, got %T", result["ocid_list"])
	}
	if len(ocidList) != 2 {
		t.Fatalf("expected ocid_list length 2, got %d", len(ocidList))
	}
	if ocidList[0] != "ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq" {
		t.Errorf("unexpected ocid_list[0]: %v", ocidList[0])
	}
	if ocidList[1] != "ocid1.dynamicgroup.oc1..aaaaaaaa5hmfyrdaxvmt52ekju5n7ffamn2pdvxaq6esb2vzzoduexamplea" {
		t.Errorf("unexpected ocid_list[1]: %v", ocidList[1])
	}

	if result["token_ttl"] != json.Number("1800") {
		t.Errorf("expected token_ttl=json.Number(1800), got %v (type %T)", result["token_ttl"], result["token_ttl"])
	}
	if result["token_max_ttl"] != json.Number("3600") {
		t.Errorf("expected token_max_ttl=json.Number(3600), got %v (type %T)", result["token_max_ttl"], result["token_max_ttl"])
	}

	tokenPolicies, ok := result["token_policies"].([]any)
	if !ok {
		t.Fatalf("expected token_policies to be []any, got %T", result["token_policies"])
	}
	if len(tokenPolicies) != 2 || tokenPolicies[0] != "dev" || tokenPolicies[1] != "prod" {
		t.Errorf("expected token_policies=[dev prod], got %v", tokenPolicies)
	}

	policies, ok := result["policies"].([]any)
	if !ok {
		t.Fatalf("expected policies to be []any, got %T", result["policies"])
	}
	if len(policies) != 1 || policies[0] != "legacy-policy" {
		t.Errorf("expected policies=[legacy-policy], got %v", policies)
	}

	cidrs, ok := result["token_bound_cidrs"].([]any)
	if !ok {
		t.Fatalf("expected token_bound_cidrs to be []any, got %T", result["token_bound_cidrs"])
	}
	if len(cidrs) != 1 || cidrs[0] != "10.0.0.0/8" {
		t.Errorf("expected token_bound_cidrs=[10.0.0.0/8], got %v", cidrs)
	}

	if result["token_explicit_max_ttl"] != json.Number("7200") {
		t.Errorf("expected token_explicit_max_ttl=json.Number(7200), got %v (type %T)", result["token_explicit_max_ttl"], result["token_explicit_max_ttl"])
	}
	if result["token_no_default_policy"] != true {
		t.Errorf("expected token_no_default_policy=true, got %v", result["token_no_default_policy"])
	}
	if result["token_num_uses"] != json.Number("5") {
		t.Errorf("expected token_num_uses=json.Number(5), got %v (type %T)", result["token_num_uses"], result["token_num_uses"])
	}
	if result["token_period"] != json.Number("3600") {
		t.Errorf("expected token_period=json.Number(3600), got %v (type %T)", result["token_period"], result["token_period"])
	}
	if result["token_type"] != "service" {
		t.Errorf("expected token_type=service, got %v", result["token_type"])
	}

	if _, exists := result["name"]; exists {
		t.Error("name should NOT be in toMap() output — Vault uses URL path for role name")
	}
}

func TestOCIAuthEngineRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &OCIAuthEngineRole{}
	instance.Spec.OCIAuthRole = OCIAuthRole{
		Name: "my-role",
		OCIDList: []string{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
			"ocid1.dynamicgroup.oc1..aaaaaaaa5hmfyrdaxvmt52ekju5n7ffamn2pdvxaq6esb2vzzoduexamplea",
		},
		TokenTTL:    "30m",
		TokenMaxTTL: "1h",
		TokenPolicies: []string{
			"dev",
			"prod",
		},
		TokenType: "service",
	}

	vaultPayload := map[string]any{
		"ocid_list": []any{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
			"ocid1.dynamicgroup.oc1..aaaaaaaa5hmfyrdaxvmt52ekju5n7ffamn2pdvxaq6esb2vzzoduexamplea",
		},
		"token_ttl":               json.Number("1800"),
		"token_max_ttl":           json.Number("3600"),
		"token_policies":          []any{"dev", "prod"},
		"policies":                []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "service",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state")
	}
}

func TestOCIAuthEngineRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &OCIAuthEngineRole{}
	instance.Spec.OCIAuthRole = OCIAuthRole{
		Name: "my-role",
		OCIDList: []string{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
		},
		TokenTTL: "30m",
	}

	vaultPayload := map[string]any{
		"ocid_list": []any{
			"ocid1.group.oc1..DIFFERENT-OCID",
		},
		"token_ttl":               json.Number("1800"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          []any{},
		"policies":                []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched ocid_list")
	}
}

func TestOCIAuthEngineRole_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &OCIAuthEngineRole{}
	instance.Spec.OCIAuthRole = OCIAuthRole{
		Name: "my-role",
		OCIDList: []string{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
		},
	}

	vaultPayload := map[string]any{
		"ocid_list": []any{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
		},
		"token_ttl":               json.Number("0"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          []any{},
		"policies":                []any{},
		"token_bound_cidrs":       []any{},
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
		"request_id":              "some-vault-internal-field",
		"lease_id":                "",
		"renewable":               false,
		"lease_duration":          json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestOCIAuthEngineRole_GetPath_WithNameOverride(t *testing.T) {
	instance := &OCIAuthEngineRole{}
	instance.Spec.Path = "oci"
	instance.Spec.Name = "custom-role-name"

	expected := "auth/oci/role/custom-role-name"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestOCIAuthEngineRole_GetPath_FallbackToMetadataName(t *testing.T) {
	instance := &OCIAuthEngineRole{}
	instance.Spec.Path = "oci"
	instance.Spec.Name = ""
	instance.ObjectMeta.Name = "metadata-role-name"

	expected := "auth/oci/role/metadata-role-name"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestOCIAuthEngineRole_IsDeletable(t *testing.T) {
	instance := &OCIAuthEngineRole{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true for OCI auth engine role")
	}
}

func TestOCIAuthEngineRole_IsEquivalentToDesiredState_SetFieldOrderInsensitive(t *testing.T) {
	instance := &OCIAuthEngineRole{}
	instance.Spec.OCIAuthRole = OCIAuthRole{
		OCIDList: []string{
			"ocid1.dynamicgroup.oc1..aaaaaaaa5hmfyrdaxvmt52ekju5n7ffamn2pdvxaq6esb2vzzoduexamplea",
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
		},
		TokenPolicies:   []string{"prod", "dev"},
		Policies:        []string{"z-legacy", "a-legacy"},
		TokenBoundCIDRs: []string{"192.168.0.0/16", "10.0.0.0/8"},
	}

	vaultPayload := map[string]any{
		"ocid_list": []any{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
			"ocid1.dynamicgroup.oc1..aaaaaaaa5hmfyrdaxvmt52ekju5n7ffamn2pdvxaq6esb2vzzoduexamplea",
		},
		"token_policies":    []any{"dev", "prod"},
		"policies":          []any{"a-legacy", "z-legacy"},
		"token_bound_cidrs": []any{"10.0.0.0/8", "192.168.0.0/16"},
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when set fields have same elements in different order")
	}
}

func TestOCIAuthEngineRole_IsEquivalentToDesiredState_SparseRole(t *testing.T) {
	instance := &OCIAuthEngineRole{}
	instance.Spec.OCIAuthRole = OCIAuthRole{
		OCIDList: []string{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
		},
	}

	// Vault only returns the fields that were explicitly set; unset token fields are omitted.
	vaultPayload := map[string]any{
		"ocid_list": []any{
			"ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
		},
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for sparse role with only ocid_list set")
	}
}
