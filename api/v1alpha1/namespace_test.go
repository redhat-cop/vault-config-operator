package v1alpha1

import (
	"encoding/json"
	"net/http"
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

	if result := namespace.toMap(); len(result) != 0 {
		t.Errorf("expected an empty payload, got %v", result)
	}
}

// Vault's response to GET sys/namespaces/<name>.
func vaultNamespaceRead(path string) map[string]interface{} {
	return map[string]interface{}{"custom_metadata": map[string]interface{}{}, "id": "lDdTO", "path": path + "/"}
}

func TestNamespaceCreateOrUpdate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		exists     bool
		wantWrites int
	}{
		{"missing namespace is created", false, 1},
		{"existing namespace is not written again", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			vaultClient, ts := newFakeVaultClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/sys/namespaces/team-a" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				switch r.Method {
				case http.MethodGet:
					if !tc.exists {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{"data": vaultNamespaceRead("org/team-a")}) // test handler; encode error is not actionable
				case http.MethodPut, http.MethodPost:
					writes++
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			}))
			defer ts.Close()

			namespace := &Namespace{Spec: NamespaceSpec{Name: "team-a", Path: "org"}}
			if err := vaultutils.NewVaultEndpoint(namespace).CreateOrUpdate(pivContext(nil, vaultClient)); err != nil {
				t.Fatalf("CreateOrUpdate: %v", err)
			}
			if writes != tc.wantWrites {
				t.Errorf("writes = %d, want %d", writes, tc.wantWrites)
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
