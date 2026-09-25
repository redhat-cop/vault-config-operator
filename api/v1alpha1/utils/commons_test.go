package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil returns empty", nil, ""},
		{"string passes through", "hello", "hello"},
		{"empty string passes through", "", ""},
		{"[]byte converts to string", []byte("PEM-DATA"), "PEM-DATA"},
		{"int formats via Sprintf", 42, "42"},
		{"bool formats via Sprintf", true, "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToString(tt.input)
			if got != tt.expected {
				t.Errorf("ToString(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetTargetNamespace(t *testing.T) {
	tests := []struct {
		name            string
		namespace       string
		targetNamespace string
		expected        string
	}{
		{"no namespaces is root", "", "", ""},
		{"namespace only", "tenant-a", "", "tenant-a"},
		{"target from root", "", "tenant-a", "tenant-a"},
		{"target is relative to namespace", "org", "tenant-a", "org/tenant-a"},
		{"nested target from root", "", "org/tenant-a", "org/tenant-a"},
		{"surrounding slashes are removed", "org/", "/tenant-a/", "org/tenant-a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := &KubeAuthConfiguration{Namespace: tt.namespace, TargetNamespace: tt.targetNamespace}
			got := kc.GetTargetNamespace()
			if got != tt.expected {
				t.Errorf("GetTargetNamespace() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestVaultClientTargetNamespace checks the namespace header of the login and of a write.
func TestVaultClientTargetNamespace(t *testing.T) {
	tests := []struct {
		name            string
		namespace       string
		targetNamespace string
		wantLoginNs     string
		wantWriteNs     string
	}{
		{"no target keeps login and writes together", "tenant-a", "", "tenant-a", "tenant-a"},
		{"login at root, write into child", "", "tenant-a", "", "tenant-a"},
		{"login in parent, write into nested child", "org", "tenant-a", "org", "org/tenant-a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loginNs, writeNs string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/v1/auth/kubernetes/login":
					loginNs = r.Header.Get(vault.NamespaceHeaderName)
					fmt.Fprint(w, `{"auth":{"client_token":"test-token"}}`) //nolint:errcheck // test helper
				case "/v1/sys/policies/acl/test":
					writeNs = r.Header.Get(vault.NamespaceHeaderName)
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()
			t.Setenv("VAULT_ADDR", srv.URL)
			t.Setenv("VAULT_NAMESPACE", "")

			kc := &KubeAuthConfiguration{
				Path:            "kubernetes",
				Role:            "test",
				Namespace:       tt.namespace,
				TargetNamespace: tt.targetNamespace,
			}
			ctx := ContextWithVaultConnection(context.Background(), nil)
			client, err := kc.createVaultClient(ctx, "test-jwt", "default")
			if err != nil {
				t.Fatalf("createVaultClient: %v", err)
			}
			if _, err := kc.withTargetNamespace(client).Logical().Write("sys/policies/acl/test", map[string]any{"policy": ""}); err != nil {
				t.Fatalf("write: %v", err)
			}

			if loginNs != tt.wantLoginNs {
				t.Errorf("login namespace = %q, want %q", loginNs, tt.wantLoginNs)
			}
			if writeNs != tt.wantWriteNs {
				t.Errorf("write namespace = %q, want %q", writeNs, tt.wantWriteNs)
			}
			if client.Namespace() != tt.wantLoginNs {
				t.Errorf("cached client namespace = %q, want %q", client.Namespace(), tt.wantLoginNs)
			}
		})
	}
}
