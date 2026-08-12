package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTerraformCloudSecretEngineRole_toMap(t *testing.T) {
	role := TFCSERole{
		Organization:   "my-org",
		TeamID:         "",
		UserID:         "",
		CredentialType: "organization",
		Description:    "org token",
		TTL:            "1h",
		MaxTTL:         "24h",
	}

	result := role.toMap()

	if result["organization"] != "my-org" {
		t.Errorf("organization = %v, expected my-org", result["organization"])
	}
	if _, ok := result["team_id"]; ok {
		t.Error("team_id should not be in map when empty")
	}
	if _, ok := result["user_id"]; ok {
		t.Error("user_id should not be in map when empty")
	}
	if result["credential_type"] != "organization" {
		t.Errorf("credential_type = %v, expected organization", result["credential_type"])
	}
	if result["description"] != "org token" {
		t.Errorf("description = %v, expected 'org token'", result["description"])
	}

	ttl, ok := result["ttl"].(json.Number)
	if !ok {
		t.Fatalf("ttl type = %T, expected json.Number", result["ttl"])
	}
	if ttl.String() != "3600" {
		t.Errorf("ttl = %v, expected 3600", ttl)
	}

	maxTTL, ok := result["max_ttl"].(json.Number)
	if !ok {
		t.Fatalf("max_ttl type = %T, expected json.Number", result["max_ttl"])
	}
	if maxTTL.String() != "86400" {
		t.Errorf("max_ttl = %v, expected 86400", maxTTL)
	}
}

func TestTerraformCloudSecretEngineRole_toMap_UserID(t *testing.T) {
	role := TFCSERole{
		UserID:         "user-glhf1234",
		CredentialType: "user",
		TTL:            "30m",
	}

	result := role.toMap()

	if result["user_id"] != "user-glhf1234" {
		t.Errorf("user_id = %v, expected user-glhf1234", result["user_id"])
	}
	if result["credential_type"] != "user" {
		t.Errorf("credential_type = %v, expected user", result["credential_type"])
	}
	if _, ok := result["organization"]; ok {
		t.Error("organization should not be in map when empty")
	}
	if _, ok := result["team_id"]; ok {
		t.Error("team_id should not be in map when empty")
	}

	ttl, ok := result["ttl"].(json.Number)
	if !ok {
		t.Fatalf("ttl type = %T, expected json.Number", result["ttl"])
	}
	if ttl.String() != "1800" {
		t.Errorf("ttl = %v, expected 1800", ttl)
	}
}

func TestTerraformCloudSecretEngineRole_toMap_EmptyOptionals(t *testing.T) {
	role := TFCSERole{
		CredentialType: "team",
	}

	result := role.toMap()

	if len(result) != 1 {
		t.Fatalf("expected 1 key (credential_type only), got %d: %v", len(result), result)
	}
	if result["credential_type"] != "team" {
		t.Errorf("credential_type = %v, expected team", result["credential_type"])
	}
}

func TestTerraformCloudSecretEngineRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	role := &TerraformCloudSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "tfuser"},
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform",
			TFCSERole: TFCSERole{
				UserID:         "user-glhf1234",
				CredentialType: "user",
				Description:    "description",
				TTL:            "1h",
				MaxTTL:         "24h",
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_type": "user",
		"description":     "description",
		"max_ttl":         json.Number("86400"),
		"name":            "tfuser",
		"organization":    "",
		"team_id":         "",
		"ttl":             json.Number("3600"),
		"user_id":         "user-glhf1234",
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when all managed fields match")
	}
}

func TestTerraformCloudSecretEngineRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &TerraformCloudSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "tfuser"},
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform",
			TFCSERole: TFCSERole{
				UserID:         "user-glhf1234",
				CredentialType: "user",
				TTL:            "1h",
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_type": "organization",
		"ttl":             json.Number("3600"),
		"user_id":         "user-glhf1234",
	}

	if role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when credential_type differs")
	}
}

func TestTerraformCloudSecretEngineRole_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	role := &TerraformCloudSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "tfuser"},
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform",
			TFCSERole: TFCSERole{
				UserID:         "user-glhf1234",
				CredentialType: "user",
				TTL:            "1h",
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_type": "user",
		"ttl":             json.Number("3600"),
		"user_id":         "user-glhf1234",
		"name":            "tfuser",
		"organization":    "",
		"team_id":         "",
		"description":     "",
		"max_ttl":         json.Number("0"),
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: extra Vault fields (name, empty strings) should be filtered out")
	}
}

func TestTerraformCloudSecretEngineRole_GetPath(t *testing.T) {
	role := &TerraformCloudSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform",
		},
	}
	expected := "terraform/role/my-role"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestTerraformCloudSecretEngineRole_GetPathWithNameOverride(t *testing.T) {
	role := &TerraformCloudSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform",
			Name: "custom-name",
		},
	}
	expected := "terraform/role/custom-name"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestTerraformCloudSecretEngineRole_IsDeletable(t *testing.T) {
	role := &TerraformCloudSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected TerraformCloudSecretEngineRole to be deletable")
	}
}

func TestTerraformCloudSecretEngineRole_Conditions(t *testing.T) {
	role := &TerraformCloudSecretEngineRole{}

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

func TestTerraformCloudSecretEngineRole_ValidateUpdate_RejectsPathChange(t *testing.T) {
	r := &TerraformCloudSecretEngineRole{}
	oldObj := &TerraformCloudSecretEngineRole{
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform",
		},
	}
	newObj := &TerraformCloudSecretEngineRole{
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform-new",
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.path is changed")
	}
}

func TestTerraformCloudSecretEngineRole_ValidateUpdate_RejectsNameChange(t *testing.T) {
	r := &TerraformCloudSecretEngineRole{}
	oldObj := &TerraformCloudSecretEngineRole{
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform",
			Name: "old-name",
		},
	}
	newObj := &TerraformCloudSecretEngineRole{
		Spec: TerraformCloudSecretEngineRoleSpec{
			Path: "terraform",
			Name: "new-name",
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.name is changed")
	}
}
