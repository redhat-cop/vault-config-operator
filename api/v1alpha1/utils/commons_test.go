package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	vault "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
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
		vaultNamespace  string
		namespace       string
		targetNamespace string
		expected        string
	}{
		{"no namespaces is root", "", "", "", ""},
		{"namespace only", "", "tenant-a", "", "tenant-a"},
		{"target from root", "", "", "tenant-a", "tenant-a"},
		{"target is relative to namespace", "", "org", "tenant-a", "org/tenant-a"},
		{"nested target from root", "", "", "org/tenant-a", "org/tenant-a"},
		{"surrounding slashes are removed", "", "org/", "/tenant-a/", "org/tenant-a"},
		{"namespace without target is cleansed", "", "org/tenant-a/", "", "org/tenant-a"},
		{"VAULT_NAMESPACE is the login namespace", "admin", "", "", "admin"},
		{"target is relative to VAULT_NAMESPACE", "admin", "", "tenant-a", "admin/tenant-a"},
		{"namespace overrides VAULT_NAMESPACE", "admin", "org", "tenant-a", "org/tenant-a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(vault.EnvVaultNamespace, tt.vaultNamespace)
			kc := &KubeAuthConfiguration{Namespace: tt.namespace, TargetNamespace: tt.targetNamespace}
			got := kc.GetTargetNamespace()
			if got != tt.expected {
				t.Errorf("GetTargetNamespace() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValidateTargetNamespaceUpdate(t *testing.T) {
	tests := []struct {
		name           string
		vaultNamespace string
		old            KubeAuthConfiguration
		new            KubeAuthConfiguration
		wantErr        bool
	}{
		{"unchanged empty", "", KubeAuthConfiguration{}, KubeAuthConfiguration{}, false},
		{"unchanged", "", KubeAuthConfiguration{TargetNamespace: "tenant-a"}, KubeAuthConfiguration{TargetNamespace: "tenant-a"}, false},
		{"changed", "", KubeAuthConfiguration{TargetNamespace: "tenant-a"}, KubeAuthConfiguration{TargetNamespace: "tenant-b"}, true},
		{"added", "", KubeAuthConfiguration{}, KubeAuthConfiguration{TargetNamespace: "tenant-a"}, true},
		{"removed", "", KubeAuthConfiguration{TargetNamespace: "tenant-a"}, KubeAuthConfiguration{}, true},
		{"only slashes changed", "", KubeAuthConfiguration{TargetNamespace: "tenant-a"}, KubeAuthConfiguration{TargetNamespace: "/tenant-a/"}, false},
		{"child moved from namespace to targetNamespace", "", KubeAuthConfiguration{Namespace: "org/tenant-a"}, KubeAuthConfiguration{Namespace: "org", TargetNamespace: "tenant-a"}, false},
		{"child moved from targetNamespace to namespace", "", KubeAuthConfiguration{Namespace: "org", TargetNamespace: "tenant-a"}, KubeAuthConfiguration{Namespace: "org/tenant-a"}, false},
		{"child moved from VAULT_NAMESPACE to targetNamespace", "admin/tenant-a", KubeAuthConfiguration{}, KubeAuthConfiguration{Namespace: "admin", TargetNamespace: "tenant-a"}, false},
		{"namespace change without a targetNamespace change is not checked", "", KubeAuthConfiguration{Namespace: "org", TargetNamespace: "tenant-a"}, KubeAuthConfiguration{Namespace: "org2", TargetNamespace: "tenant-a"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(vault.EnvVaultNamespace, tt.vaultNamespace)
			err := tt.new.ValidateTargetNamespaceUpdate(&tt.old)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTargetNamespaceUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithChildNamespace(t *testing.T) {
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetNamespace("org/tenant-a")

	if got := WithChildNamespace(client, ""); got != client {
		t.Error("expected the same client when child is empty")
	}
	if got := WithChildNamespace(client, "/team/").Namespace(); got != "org/tenant-a/team" {
		t.Errorf("child namespace = %q, want %q", got, "org/tenant-a/team")
	}
	if client.Namespace() != "org/tenant-a" {
		t.Errorf("original client namespace = %q, want %q", client.Namespace(), "org/tenant-a")
	}
}

func TestGetVaultClientTargetNamespace(t *testing.T) {
	tests := []struct {
		name            string
		vaultNamespace  string
		namespace       string
		targetNamespace string
		wantLoginNs     string
		wantWriteNs     string
	}{
		{"no target keeps login and writes together", "", "tenant-a", "", "tenant-a", "tenant-a"},
		{"login at root, write into child", "", "", "tenant-a", "", "tenant-a"},
		{"login in parent, write into nested child", "", "org", "tenant-a", "org", "org/tenant-a"},
		{"login in VAULT_NAMESPACE, write into child", "admin", "", "tenant-a", "admin", "admin/tenant-a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loginNs, writeNs string
			logins := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/api/v1/namespaces/default/serviceaccounts/default/token":
					fmt.Fprint(w, `{"kind":"TokenRequest","apiVersion":"authentication.k8s.io/v1","status":{"token":"test-jwt"}}`) //nolint:errcheck // test helper
				case "/v1/auth/kubernetes/login":
					logins++
					loginNs = r.Header.Get(vault.NamespaceHeaderName)
					fmt.Fprint(w, `{"auth":{"client_token":"test-token"}}`) //nolint:errcheck // test helper
				case "/v1/auth/token/lookup-self":
					fmt.Fprint(w, `{"data":{"id":"test-token"}}`) //nolint:errcheck // test helper
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
			t.Setenv(vault.EnvVaultNamespace, tt.vaultNamespace)
			// "false" turns off the cache and the lifetime watcher
			t.Setenv("CACHE_VAULT_TOKEN", "false")

			kc := &KubeAuthConfiguration{
				ServiceAccount:  &corev1.LocalObjectReference{Name: "default"},
				Path:            "kubernetes",
				Role:            "test",
				Namespace:       tt.namespace,
				TargetNamespace: tt.targetNamespace,
			}
			ctx := ContextWithVaultConnection(context.Background(), nil)
			ctx = ContextWithRestConfig(ctx, &rest.Config{Host: srv.URL})

			checkWrite := func(t *testing.T, client *vault.Client) {
				t.Helper()
				writeNs = ""
				if _, err := client.Logical().Write("sys/policies/acl/test", map[string]any{"policy": ""}); err != nil {
					t.Fatalf("write: %v", err)
				}
				if writeNs != tt.wantWriteNs {
					t.Errorf("write namespace = %q, want %q", writeNs, tt.wantWriteNs)
				}
			}

			t.Run("new client", func(t *testing.T) {
				client, err := kc.GetVaultClient(ctx, "default")
				if err != nil {
					t.Fatalf("GetVaultClient: %v", err)
				}
				if loginNs != tt.wantLoginNs {
					t.Errorf("login namespace = %q, want %q", loginNs, tt.wantLoginNs)
				}
				checkWrite(t, client)
			})

			t.Run("cached client", func(t *testing.T) {
				loginClient, err := kc.createVaultClient(ctx, "test-jwt", "default")
				if err != nil {
					t.Fatalf("createVaultClient: %v", err)
				}
				vaultClientCache.Put(kc, "default", loginClient)
				defer vaultClientCache.Delete(kc, "default")
				t.Setenv("CACHE_VAULT_TOKEN", "true")

				logins = 0
				client, err := kc.GetVaultClient(ctx, "default")
				if err != nil {
					t.Fatalf("GetVaultClient: %v", err)
				}
				if logins != 0 {
					t.Fatalf("expected the cached client, got %d new logins", logins)
				}
				checkWrite(t, client)
				if loginClient.Namespace() != tt.wantLoginNs {
					t.Errorf("cached client namespace = %q, want %q", loginClient.Namespace(), tt.wantLoginNs)
				}
			})
		})
	}
}
