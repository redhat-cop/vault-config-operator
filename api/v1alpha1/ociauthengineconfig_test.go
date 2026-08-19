package v1alpha1

import (
	"encoding/json"
	"testing"
)

func TestOCIAuthEngineConfig_toMap(t *testing.T) {
	config := &OCIAuthConfig{
		HomeTenancyID: "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq",
	}

	result := config.toMap()

	if result["home_tenancy_id"] != "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq" {
		t.Errorf("expected home_tenancy_id to match, got %v", result["home_tenancy_id"])
	}

	if len(result) != 1 {
		t.Errorf("expected exactly 1 key in toMap output, got %d", len(result))
	}
}

func TestOCIAuthEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &OCIAuthEngineConfig{}
	instance.Spec.OCIAuthConfig = OCIAuthConfig{
		HomeTenancyID: "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq",
	}

	vaultPayload := map[string]any{
		"home_tenancy_id": "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state")
	}
}

func TestOCIAuthEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &OCIAuthEngineConfig{}
	instance.Spec.OCIAuthConfig = OCIAuthConfig{
		HomeTenancyID: "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq",
	}

	vaultPayload := map[string]any{
		"home_tenancy_id": "ocid1.tenancy.oc1..different-tenancy-id",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched home_tenancy_id")
	}
}

func TestOCIAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &OCIAuthEngineConfig{}
	instance.Spec.OCIAuthConfig = OCIAuthConfig{
		HomeTenancyID: "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq",
	}

	vaultPayload := map[string]any{
		"home_tenancy_id": "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq",
		"request_id":      "extra-vault-field",
		"lease_duration":  json.Number("0"),
		"renewable":       false,
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestOCIAuthEngineConfig_GetPath(t *testing.T) {
	instance := &OCIAuthEngineConfig{}
	instance.Spec.Path = "oci"

	expected := "auth/oci/config"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestOCIAuthEngineConfig_IsDeletable(t *testing.T) {
	instance := &OCIAuthEngineConfig{}
	if instance.IsDeletable() {
		t.Error("expected IsDeletable() to return false for OCI auth engine config")
	}
}
