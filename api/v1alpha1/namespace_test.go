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

// Vault's response to GET sys/namespaces/<name>.
func vaultNamespaceRead(path string) map[string]interface{} {
	return map[string]interface{}{"custom_metadata": map[string]interface{}{}, "id": "lDdTO", "path": path + "/"}
}

func TestNamespaceIsEquivalentToVaultRead(t *testing.T) {
	for _, tc := range []struct {
		name string
		spec NamespaceSpec
		read string
	}{
		{"top-level namespace", NamespaceSpec{Name: "team-a"}, "team-a"},
		{"nested namespace", NamespaceSpec{Name: "team-a", Path: "org"}, "org/team-a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			namespace := &Namespace{Spec: tc.spec}
			if !namespace.IsEquivalentToDesiredState(vaultNamespaceRead(tc.read)) {
				t.Error("expected an existing namespace to be equivalent to its desired state")
			}
		})
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
