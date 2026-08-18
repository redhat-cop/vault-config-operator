package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestOktaAuthEngineGroup_toMap(t *testing.T) {
	group := &OktaAuthEngineGroup{
		Spec: OktaAuthEngineGroupSpec{
			Policies: "admin,reader",
		},
	}

	result := group.toMap()

	if len(result) != 1 {
		t.Errorf("expected 1 key in map, got %d", len(result))
	}

	if result["policies"] != "admin,reader" {
		t.Errorf("expected policies=admin,reader, got %v", result["policies"])
	}
}

func TestOktaAuthEngineGroup_IsEquivalentToDesiredState_Match(t *testing.T) {
	group := &OktaAuthEngineGroup{
		Spec: OktaAuthEngineGroupSpec{
			Policies: "admin,reader",
		},
	}

	vaultPayload := map[string]any{
		"policies": "admin,reader",
	}

	if !group.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestOktaAuthEngineGroup_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	group := &OktaAuthEngineGroup{
		Spec: OktaAuthEngineGroupSpec{
			Policies: "admin,reader",
		},
	}

	vaultPayload := map[string]any{
		"policies": "different-policy",
	}

	if group.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected non-matching payload (different policies) to NOT be equivalent")
	}
}

func TestOktaAuthEngineGroup_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	group := &OktaAuthEngineGroup{
		Spec: OktaAuthEngineGroupSpec{
			Policies: "admin",
		},
	}

	vaultPayload := map[string]any{
		"policies":    "admin",
		"extra_field": "unexpected",
	}

	if !group.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestOktaAuthEngineGroup_GetPath(t *testing.T) {
	group := &OktaAuthEngineGroup{
		Spec: OktaAuthEngineGroupSpec{
			Path: "okta",
			Name: "admins",
		},
	}

	expected := "auth/okta/groups/admins"
	result := group.GetPath()
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestOktaAuthEngineGroup_IsDeletable(t *testing.T) {
	group := &OktaAuthEngineGroup{}
	if !group.IsDeletable() {
		t.Error("expected OktaAuthEngineGroup to be deletable")
	}
}

func TestOktaAuthEngineGroup_Conditions(t *testing.T) {
	group := &OktaAuthEngineGroup{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	group.SetConditions(conditions)
	got := group.GetConditions()

	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected condition type 'ReconcileSuccessful', got %v", got[0].Type)
	}
	if got[0].Status != metav1.ConditionTrue {
		t.Errorf("expected condition status True, got %v", got[0].Status)
	}
}
