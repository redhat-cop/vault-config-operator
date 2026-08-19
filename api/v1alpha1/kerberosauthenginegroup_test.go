package v1alpha1

import (
	"testing"
)

func TestKerberosAuthEngineGroup_toMap(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	instance.Spec.Policies = "admin,default"

	result := instance.toMap()

	if result["policies"] != "admin,default" {
		t.Errorf("expected policies=admin,default, got %v", result["policies"])
	}
	if _, exists := result["name"]; exists {
		t.Error("expected name NOT to be in toMap() output for Kerberos groups")
	}
}

func TestKerberosAuthEngineGroup_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	instance.Spec.Policies = "admin,default"

	vaultPayload := map[string]any{
		"policies": "admin,default",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching policies")
	}
}

func TestKerberosAuthEngineGroup_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	instance.Spec.Policies = "admin,default"

	vaultPayload := map[string]any{
		"policies": "readonly",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched policies")
	}
}

func TestKerberosAuthEngineGroup_GetPath(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	instance.Spec.Path = "kerberos"
	instance.Spec.Name = "engineering"

	expected := "auth/kerberos/groups/engineering"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestKerberosAuthEngineGroup_IsDeletable(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true for auth engine group")
	}
}

func TestKerberosAuthEngineGroup_IsEquivalentToDesiredState_VaultPoliciesArray(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	instance.Spec.Policies = "admin,default"

	vaultPayload := map[string]any{
		"policies": []interface{}{"admin", "default"},
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns policies as array matching CR comma-separated string")
	}
}

func TestKerberosAuthEngineGroup_IsEquivalentToDesiredState_VaultPoliciesArrayReordered(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	instance.Spec.Policies = "default,admin"

	vaultPayload := map[string]any{
		"policies": []interface{}{"admin", "default"},
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns policies in different order")
	}
}

func TestKerberosAuthEngineGroup_IsEquivalentToDesiredState_VaultPoliciesArrayMismatch(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	instance.Spec.Policies = "admin,default"

	vaultPayload := map[string]any{
		"policies": []interface{}{"readonly"},
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false when Vault policies array does not match CR")
	}
}

func TestKerberosAuthEngineGroup_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &KerberosAuthEngineGroup{}
	instance.Spec.Policies = "admin"

	vaultPayload := map[string]any{
		"policies":   "admin",
		"request_id": "extra-vault-field",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}
