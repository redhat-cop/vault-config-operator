package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vault "github.com/hashicorp/vault/api"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGroupGetPath(t *testing.T) {
	tests := []struct {
		name     string
		group    *Group
		expected string
	}{
		{
			name: "uses metadata name when spec name is empty",
			group: &Group{
				ObjectMeta: metav1.ObjectMeta{Name: "my-group"},
			},
			expected: "identity/group/name/my-group",
		},
		{
			name: "spec name takes precedence over metadata name",
			group: &Group{
				ObjectMeta: metav1.ObjectMeta{Name: "my-group"},
				Spec:       GroupSpec{Name: "override-name"},
			},
			expected: "identity/group/name/override-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.group.GetPath(); got != tt.expected {
				t.Errorf("GetPath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGroupGetPayloadInternalIncludesMemberFields(t *testing.T) {
	group := &Group{
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:            "internal",
				Policies:        []string{"default"},
				MemberGroupIDs:  []string{"group-1"},
				MemberEntityIDs: []string{"entity-1"},
			},
		},
	}
	payload := group.GetPayload()
	if payload["type"] != "internal" {
		t.Errorf("expected type 'internal', got %v", payload["type"])
	}
	if _, ok := payload["member_group_ids"]; !ok {
		t.Error("expected member_group_ids to be present for internal group")
	}
	if _, ok := payload["member_entity_ids"]; !ok {
		t.Error("expected member_entity_ids to be present for internal group")
	}
}

func TestGroupGetPayloadExternalOmitsMemberFields(t *testing.T) {
	group := &Group{
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type: "external",
			},
		},
	}
	payload := group.GetPayload()
	if _, ok := payload["member_group_ids"]; ok {
		t.Error("expected member_group_ids to be absent for external group")
	}
	if _, ok := payload["member_entity_ids"]; ok {
		t.Error("expected member_entity_ids to be absent for external group")
	}
}

func TestGroupIsEquivalentToDesiredState(t *testing.T) {
	group := &Group{
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:     "internal",
				Policies: []string{"default"},
			},
		},
	}
	if !group.IsEquivalentToDesiredState(group.GetPayload()) {
		t.Error("expected group to be equivalent to its own payload")
	}
}

func TestGroupIsEquivalentToDesiredStateMismatch(t *testing.T) {
	group := &Group{
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:     "internal",
				Policies: []string{"default"},
			},
		},
	}
	payload := group.GetPayload()
	payload["type"] = "external"
	if group.IsEquivalentToDesiredState(payload) {
		t.Error("expected group to not be equivalent to payload with different type")
	}
}

// newVaultTestClient creates a vault.Client pointed at the given httptest.Server.
func newVaultTestClient(t *testing.T, srv *httptest.Server) *vault.Client {
	t.Helper()
	cfg := vault.DefaultConfig()
	cfg.Address = srv.URL
	client, err := vault.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault.NewClient: %v", err)
	}
	client.SetToken("test-token")
	return client
}

func TestGroupEnrichStatus(t *testing.T) {
	const wantID = "abc-123-def-456"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"id":   wantID,
				"name": "my-group",
			},
		})
	}))
	defer srv.Close()

	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "my-group"},
	}
	ctx := vaultutils.ContextWithVaultClient(context.Background(), newVaultTestClient(t, srv))

	if err := group.EnrichStatus(ctx); err != nil {
		t.Fatalf("EnrichStatus() unexpected error: %v", err)
	}
	if group.Status.ID != wantID {
		t.Errorf("Status.ID = %q, want %q", group.Status.ID, wantID)
	}
}

func TestGroupEnrichStatusNoIDField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"name": "my-group", // no "id" key
			},
		})
	}))
	defer srv.Close()

	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "my-group"},
	}
	ctx := vaultutils.ContextWithVaultClient(context.Background(), newVaultTestClient(t, srv))

	if err := group.EnrichStatus(ctx); err != nil {
		t.Fatalf("EnrichStatus() unexpected error: %v", err)
	}
	if group.Status.ID != "" {
		t.Errorf("expected Status.ID to remain empty, got %q", group.Status.ID)
	}
}

func TestGroupEnrichStatusVaultError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errors":["internal server error"]}`, http.StatusInternalServerError)
	}))
	defer srv.Close()

	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "my-group"},
	}
	ctx := vaultutils.ContextWithVaultClient(context.Background(), newVaultTestClient(t, srv))

	if err := group.EnrichStatus(ctx); err == nil {
		t.Error("expected error from Vault 500 response, got nil")
	}
}
