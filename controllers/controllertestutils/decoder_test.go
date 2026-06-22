package controllertestutils

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateFromYAML_PreservesOnlyYAMLFields(t *testing.T) {
	yamlContent := `apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineRole
metadata:
  name: test-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-mount
  policies:
    - vault-admin
  targetServiceAccounts:
    - default
  targetNamespaces:
    targetNamespaces:
      - vault-admin
`
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test-role.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp YAML: %v", err)
	}

	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "redhatcop.redhat.io", Version: "v1alpha1", Kind: "KubernetesAuthEngineRole"},
		&unstructured.Unstructured{},
	)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	d := NewDecoder()

	name, err := d.CreateFromYAML(context.Background(), fakeClient, yamlPath, "test-ns")
	if err != nil {
		t.Fatalf("CreateFromYAML returned error: %v", err)
	}

	if name != "test-role" {
		t.Errorf("expected name 'test-role', got '%s'", name)
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "redhatcop.redhat.io", Version: "v1alpha1", Kind: "KubernetesAuthEngineRole",
	})
	err = fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-role", Namespace: "test-ns"}, obj)
	if err != nil {
		t.Fatalf("failed to Get created object: %v", err)
	}

	spec, ok := obj.Object["spec"].(map[string]any)
	if !ok {
		t.Fatal("expected spec to be a map")
	}

	if _, exists := spec["tokenType"]; exists {
		t.Error("tokenType should NOT be present — it was absent in YAML and must not appear via unstructured creation")
	}
	if _, exists := spec["aliasNameSource"]; exists {
		t.Error("aliasNameSource should NOT be present — it was absent in YAML and must not appear via unstructured creation")
	}

	if _, exists := spec["policies"]; !exists {
		t.Error("policies should be present — it was specified in YAML")
	}
	if _, exists := spec["path"]; !exists {
		t.Error("path should be present — it was specified in YAML")
	}
	if _, exists := spec["authentication"]; !exists {
		t.Error("authentication should be present — it was specified in YAML")
	}

	if obj.GetNamespace() != "test-ns" {
		t.Errorf("expected namespace 'test-ns', got '%s'", obj.GetNamespace())
	}
}

func TestCreateFromYAML_FileNotFound(t *testing.T) {
	fakeClient := fake.NewClientBuilder().Build()
	d := NewDecoder()

	_, err := d.CreateFromYAML(context.Background(), fakeClient, "/nonexistent/file.yaml", "test-ns")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestCreateFromYAML_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(yamlPath, []byte("{{invalid yaml content"), 0644); err != nil {
		t.Fatalf("failed to write temp YAML: %v", err)
	}

	fakeClient := fake.NewClientBuilder().Build()
	d := NewDecoder()

	_, err := d.CreateFromYAML(context.Background(), fakeClient, yamlPath, "test-ns")
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
