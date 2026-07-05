package v1alpha1

import (
	"reflect"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespacePath(t *testing.T) {
	namespace := &Namespace{
		Spec: NamespaceSpec{
			Name: "myNamespace",
			Path: "myParentNamespace",
		},
	}

	result := namespace.GetPath()
	expected := "sys/namespaces/myNamespace"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestNamespaceToMap(t *testing.T) {
	namespace := &Namespace{
		Spec: NamespaceSpec{
			Name: "myNamespace",
			Path: "myParentNamespace",
		},
	}

	result := namespace.toMap()

	if len(result) != 2 {
		t.Errorf("expected 2 keys in map, got %d", len(result))
	}

	expected := map[string]interface{}{
		"name": "myNamespace",
		"path": vaultutils.Path("myParentNamespace"),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestNamespaceIsEquivalentMatching(t *testing.T) {
	namespace := &Namespace{
		Spec: NamespaceSpec{
			Name: "myNamespace",
			Path: "myParentNamespace",
		},
	}

	payload := namespace.toMap()

	if !namespace.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestNamespaceIsEquivalentNonMatching(t *testing.T) {
	namespace := &Namespace{
		Spec: NamespaceSpec{
			Name: "myNamespace",
			Path: "myParentNamespace",
		},
	}

	payload := namespace.toMap()
	payload["path"] = "different-path"

	if namespace.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different policies) to NOT be equivalent")
	}
}

func TestNamespaceIsEquivalentExtraFields(t *testing.T) {
	namespace := &Namespace{
		Spec: NamespaceSpec{
			Name: "admins",
			Path: "myParentNamespace",
		},
	}

	payload := namespace.toMap()
	payload["extra_field"] = "unexpected"

	if !namespace.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestNamespaceIsDeletable(t *testing.T) {
	namespace := &Namespace{}
	if !namespace.IsDeletable() {
		t.Error("expected Namespace to be deletable")
	}
}

func TestNamespaceConditions(t *testing.T) {
	namespace := &Namespace{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	namespace.SetConditions(conditions)
	got := namespace.GetConditions()

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
