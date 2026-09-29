package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	vault "github.com/hashicorp/vault/api"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestNamespaceVaultNamespaceHeader(t *testing.T) {
	tests := []struct {
		name            string
		namespace       string
		targetNamespace string
		path            vaultutils.Path
		wantLoginNs     string
		wantNs          string
	}{
		{"login at root", "", "", "", "", ""},
		{"path under the login namespace", "org", "", "parent", "org", "org/parent"},
		{"path under the target namespace", "org", "tenant-a", "parent", "org", "org/tenant-a/parent"},
		{"target namespace without path", "org", "tenant-a", "", "org", "org/tenant-a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var loginNs string
			namespaceHeaders := map[string]string{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/api/v1/namespaces/default/serviceaccounts/default/token":
					fmt.Fprint(w, `{"kind":"TokenRequest","apiVersion":"authentication.k8s.io/v1","status":{"token":"test-jwt"}}`) //nolint:errcheck // test helper
				case "/v1/auth/kubernetes/login":
					loginNs = r.Header.Get(vault.NamespaceHeaderName)
					fmt.Fprint(w, `{"auth":{"client_token":"test-token"}}`) //nolint:errcheck // test helper
				case "/v1/sys/namespaces/team-a":
					namespaceHeaders[r.Method] = r.Header.Get(vault.NamespaceHeaderName)
					if r.Method == http.MethodGet {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()
			t.Setenv("VAULT_ADDR", srv.URL)
			t.Setenv(vault.EnvVaultNamespace, "")
			t.Setenv("CACHE_VAULT_TOKEN", "false")

			instance := &redhatcopv1alpha1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: "team-a", Namespace: "default"},
				Spec: redhatcopv1alpha1.NamespaceSpec{
					Path: tt.path,
					Authentication: vaultutils.KubeAuthConfiguration{
						ServiceAccount:  &corev1.LocalObjectReference{Name: "default"},
						Path:            "kubernetes",
						Role:            "test",
						Namespace:       tt.namespace,
						TargetNamespace: tt.targetNamespace,
					},
				},
			}
			base := vaultresourcecontroller.NewReconcilerBase(nil, nil, &rest.Config{Host: srv.URL}, nil, nil, logr.Discard(), "Namespace")
			ctx, err := prepareContext(context.Background(), base, instance)
			if err != nil {
				t.Fatalf("prepareContext: %v", err)
			}

			endpoint := vaultutils.NewVaultEndpoint(instance)
			if err := endpoint.CreateOrUpdate(ctx); err != nil {
				t.Fatalf("CreateOrUpdate: %v", err)
			}
			if err := endpoint.DeleteIfExists(ctx); err != nil {
				t.Fatalf("DeleteIfExists: %v", err)
			}

			if loginNs != tt.wantLoginNs {
				t.Errorf("login namespace = %q, want %q", loginNs, tt.wantLoginNs)
			}
			for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
				got, ok := namespaceHeaders[method]
				if !ok {
					t.Errorf("no %s request for the namespace", method)
				} else if got != tt.wantNs {
					t.Errorf("%s namespace = %q, want %q", method, got, tt.wantNs)
				}
			}
			if instance.Spec.Authentication.TargetNamespace != tt.targetNamespace {
				t.Errorf("spec.authentication.targetNamespace changed to %q", instance.Spec.Authentication.TargetNamespace)
			}
		})
	}
}
