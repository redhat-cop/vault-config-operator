package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLDAPAuthEngineGroupGetPath(t *testing.T) {
	group := &LDAPAuthEngineGroup{
		Spec: LDAPAuthEngineGroupSpec{
			Path: "ldap",
			Name: "admins",
		},
	}

	result := group.GetPath()
	expected := "auth/ldap/groups/admins"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestLDAPAuthEngineGroupToMap(t *testing.T) {
	group := &LDAPAuthEngineGroup{
		Spec: LDAPAuthEngineGroupSpec{
			Name:     "admins",
			Policies: "admin,reader",
		},
	}

	result := group.toMap()

	if len(result) != 2 {
		t.Errorf("expected 2 keys in map, got %d", len(result))
	}

	expected := map[string]interface{}{
		"name":     "admins",
		"policies": "admin,reader",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestLDAPAuthEngineGroupIsEquivalentMatching(t *testing.T) {
	group := &LDAPAuthEngineGroup{
		Spec: LDAPAuthEngineGroupSpec{
			Name:     "admins",
			Policies: "admin,reader",
		},
	}

	payload := group.toMap()

	if !group.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestLDAPAuthEngineGroupIsEquivalentNonMatching(t *testing.T) {
	group := &LDAPAuthEngineGroup{
		Spec: LDAPAuthEngineGroupSpec{
			Name:     "admins",
			Policies: "admin,reader",
		},
	}

	payload := group.toMap()
	payload["policies"] = "different-policy"

	if group.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different policies) to NOT be equivalent")
	}
}

func TestLDAPAuthEngineGroupIsEquivalentExtraFields(t *testing.T) {
	group := &LDAPAuthEngineGroup{
		Spec: LDAPAuthEngineGroupSpec{
			Name:     "admins",
			Policies: "admin",
		},
	}

	payload := group.toMap()
	payload["extra_field"] = "unexpected"

	if group.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
	}
}

func TestLDAPAuthEngineGroupIsDeletable(t *testing.T) {
	group := &LDAPAuthEngineGroup{}
	if !group.IsDeletable() {
		t.Error("expected LDAPAuthEngineGroup to be deletable")
	}
}

func TestLDAPAuthEngineGroupConditions(t *testing.T) {
	group := &LDAPAuthEngineGroup{}

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
