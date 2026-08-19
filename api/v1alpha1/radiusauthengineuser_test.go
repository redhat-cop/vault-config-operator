package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRADIUSAuthEngineUser_toMap(t *testing.T) {
	instance := &RADIUSAuthEngineUser{}
	instance.Spec.Policies = "default,dev"

	result := instance.toMap()

	if result["policies"] != "default,dev" {
		t.Errorf("expected policies=default,dev, got %v", result["policies"])
	}
	if len(result) != 1 {
		t.Errorf("expected exactly 1 field in toMap output, got %d", len(result))
	}
}

func TestRADIUSAuthEngineUser_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &RADIUSAuthEngineUser{}
	instance.Spec.Policies = "default,dev"

	vaultPayload := map[string]any{
		"policies": "default,dev",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state")
	}
}

func TestRADIUSAuthEngineUser_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &RADIUSAuthEngineUser{}
	instance.Spec.Policies = "default,dev"

	vaultPayload := map[string]any{
		"policies": "admin,superuser",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched policies")
	}
}

func TestRADIUSAuthEngineUser_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &RADIUSAuthEngineUser{}
	instance.Spec.Policies = "default,dev"

	vaultPayload := map[string]any{
		"policies":       "default,dev",
		"request_id":     "extra-field",
		"lease_duration": json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestRADIUSAuthEngineUser_GetPath_WithNameOverride(t *testing.T) {
	instance := &RADIUSAuthEngineUser{}
	instance.Spec.Path = "radius"
	instance.Spec.Name = "custom-user"
	instance.ObjectMeta.Name = "metadata-name"

	expected := "auth/radius/users/custom-user"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestRADIUSAuthEngineUser_GetPath_WithoutNameOverride(t *testing.T) {
	instance := &RADIUSAuthEngineUser{}
	instance.Spec.Path = "radius"
	instance.Spec.Name = ""
	instance.ObjectMeta.Name = "metadata-name"

	expected := "auth/radius/users/metadata-name"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestRADIUSAuthEngineUser_IsDeletable(t *testing.T) {
	instance := &RADIUSAuthEngineUser{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true for auth engine user")
	}
}

func TestRADIUSAuthEngineUser_Conditions(t *testing.T) {
	instance := &RADIUSAuthEngineUser{}
	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}
	instance.SetConditions(conditions)
	got := instance.GetConditions()
	if len(got) != 1 || got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected conditions round-trip, got %v", got)
	}
}
