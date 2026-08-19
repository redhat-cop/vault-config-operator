package v1alpha1

import (
	"encoding/json"
	"testing"
)

func TestKerberosAuthEngineConfig_toMap(t *testing.T) {
	config := &KerberosAuthConfig{
		ServiceAccount:     "vault_svc",
		RemoveInstanceName: true,
		AddGroupAliases:    true,
		retrievedKeytab:    "base64-encoded-keytab-data",
	}

	result := config.toMap()

	if result["keytab"] != "base64-encoded-keytab-data" {
		t.Errorf("expected keytab=base64-encoded-keytab-data, got %v", result["keytab"])
	}
	if result["service_account"] != "vault_svc" {
		t.Errorf("expected service_account=vault_svc, got %v", result["service_account"])
	}
	if result["remove_instance_name"] != true {
		t.Errorf("expected remove_instance_name=true, got %v", result["remove_instance_name"])
	}
	if result["add_group_aliases"] != true {
		t.Errorf("expected add_group_aliases=true, got %v", result["add_group_aliases"])
	}
}

func TestKerberosAuthEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &KerberosAuthEngineConfig{}
	instance.Spec.KerberosAuthConfig = KerberosAuthConfig{
		ServiceAccount:     "vault_svc",
		RemoveInstanceName: false,
		AddGroupAliases:    true,
		retrievedKeytab:    "some-keytab-content",
	}

	vaultPayload := map[string]any{
		"service_account":      "vault_svc",
		"remove_instance_name": false,
		"add_group_aliases":    true,
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state (keytab stripped)")
	}
}

func TestKerberosAuthEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &KerberosAuthEngineConfig{}
	instance.Spec.KerberosAuthConfig = KerberosAuthConfig{
		ServiceAccount:     "vault_svc",
		RemoveInstanceName: false,
		AddGroupAliases:    false,
		retrievedKeytab:    "keytab",
	}

	vaultPayload := map[string]any{
		"service_account":      "different_svc",
		"remove_instance_name": false,
		"add_group_aliases":    false,
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched service_account")
	}
}

func TestKerberosAuthEngineConfig_IsEquivalentToDesiredState_KeytabStripping(t *testing.T) {
	instance := &KerberosAuthEngineConfig{}
	instance.Spec.KerberosAuthConfig = KerberosAuthConfig{
		ServiceAccount:     "vault_svc",
		RemoveInstanceName: false,
		AddGroupAliases:    false,
		retrievedKeytab:    "super-secret-keytab-content",
	}

	vaultPayload := map[string]any{
		"service_account":      "vault_svc",
		"remove_instance_name": false,
		"add_group_aliases":    false,
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: keytab is write-only and must be excluded from drift comparison")
	}
}

func TestKerberosAuthEngineConfig_GetPath(t *testing.T) {
	instance := &KerberosAuthEngineConfig{}
	instance.Spec.Path = "kerberos"

	expected := "auth/kerberos/config"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestKerberosAuthEngineConfig_IsDeletable(t *testing.T) {
	instance := &KerberosAuthEngineConfig{}
	if instance.IsDeletable() {
		t.Error("expected IsDeletable() to return false for auth engine config")
	}
}

func TestKerberosAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &KerberosAuthEngineConfig{}
	instance.Spec.KerberosAuthConfig = KerberosAuthConfig{
		ServiceAccount:     "vault_svc",
		RemoveInstanceName: false,
		AddGroupAliases:    false,
		retrievedKeytab:    "keytab",
	}

	vaultPayload := map[string]any{
		"service_account":      "vault_svc",
		"remove_instance_name": false,
		"add_group_aliases":    false,
		"request_id":           "extra-vault-field",
		"lease_duration":       json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}
