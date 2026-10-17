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
		name         string
		group        *Group
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			group: &Group{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       GroupSpec{Name: "spec-name"},
			},
			expectedPath: "identity/group/name/spec-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			group: &Group{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       GroupSpec{},
			},
			expectedPath: "identity/group/name/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.group.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestGroupToMapInternal(t *testing.T) {
	spec := &GroupSpec{
		GroupConfig: GroupConfig{
			Type:            "internal",
			Metadata:        map[string]string{"team": "platform"},
			Policies:        []string{"default", "admin"},
			MemberGroupIDs:  []string{"group-1"},
			MemberEntityIDs: []string{"entity-1", "entity-2"},
		},
	}

	m := spec.toMap()

	if len(m) != 5 {
		t.Errorf("expected 5 keys for internal group, got %d", len(m))
	}
	if m["type"] != "internal" {
		t.Errorf("expected type 'internal', got %v", m["type"])
	}
	if _, ok := m["metadata"]; !ok {
		t.Error("expected metadata key to be present")
	}
	if _, ok := m["policies"]; !ok {
		t.Error("expected policies key to be present")
	}
	if _, ok := m["member_group_ids"]; !ok {
		t.Error("expected member_group_ids key for internal group")
	}
	if _, ok := m["member_entity_ids"]; !ok {
		t.Error("expected member_entity_ids key for internal group")
	}
}

func TestGroupToMapExternal(t *testing.T) {
	spec := &GroupSpec{
		GroupConfig: GroupConfig{
			Type:            "external",
			Metadata:        map[string]string{"team": "platform"},
			Policies:        []string{"default"},
			MemberGroupIDs:  []string{"should-be-excluded"},
			MemberEntityIDs: []string{"should-be-excluded"},
		},
	}

	m := spec.toMap()

	if len(m) != 3 {
		t.Errorf("expected 3 keys for external group, got %d", len(m))
	}
	if m["type"] != "external" {
		t.Errorf("expected type 'external', got %v", m["type"])
	}
	if _, ok := m["member_group_ids"]; ok {
		t.Error("expected member_group_ids to be absent for external group")
	}
	if _, ok := m["member_entity_ids"]; ok {
		t.Error("expected member_entity_ids to be absent for external group")
	}
}

func TestGroupIsEquivalentMatching(t *testing.T) {
	group := &Group{
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:            "internal",
				Metadata:        map[string]string{"team": "platform"},
				Policies:        []string{"default"},
				MemberGroupIDs:  []string{"group-1"},
				MemberEntityIDs: []string{"entity-1"},
			},
		},
	}

	// "name" is deleted from payload before comparison
	payload := map[string]any{
		"name":              "some-group",
		"type":              "internal",
		"metadata":          map[string]string{"team": "platform"},
		"policies":          []string{"default"},
		"member_group_ids":  []string{"group-1"},
		"member_entity_ids": []string{"entity-1"},
	}
	if !group.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (with extra 'name' key) to be equivalent")
	}
}

func TestGroupIsEquivalentNonMatching(t *testing.T) {
	group := &Group{
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:            "internal",
				Metadata:        map[string]string{"team": "platform"},
				Policies:        []string{"default"},
				MemberGroupIDs:  []string{"group-1"},
				MemberEntityIDs: []string{"entity-1"},
			},
		},
	}

	payload := map[string]any{
		"type":              "internal",
		"metadata":          map[string]string{"team": "other-team"},
		"policies":          []string{"default"},
		"member_group_ids":  []string{"group-1"},
		"member_entity_ids": []string{"entity-1"},
	}
	if group.IsEquivalentToDesiredState(payload) {
		t.Error("expected different metadata to not be equivalent")
	}
}

// Payload keys not present in desiredState are filtered before comparison.
func TestGroupIsEquivalentExtraFields(t *testing.T) {
	group := &Group{
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:            "internal",
				Metadata:        map[string]string{"team": "platform"},
				Policies:        []string{"default"},
				MemberGroupIDs:  []string{"group-1"},
				MemberEntityIDs: []string{"entity-1"},
			},
		},
	}

	payload := map[string]any{
		"name":              "some-group",
		"type":              "internal",
		"metadata":          map[string]string{"team": "platform"},
		"policies":          []string{"default"},
		"member_group_ids":  []string{"group-1"},
		"member_entity_ids": []string{"entity-1"},
		"extra_field":       "unexpected",
	}
	if !group.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestGroupGetPayload(t *testing.T) {
	group := &Group{
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:     "internal",
				Metadata: map[string]string{"team": "platform"},
				Policies: []string{"default"},
			},
		},
	}

	payload := group.GetPayload()
	if payload["type"] != "internal" {
		t.Errorf("GetPayload() should delegate to Spec.toMap(), got type = %v", payload["type"])
	}
}

func TestGroupIsDeletable(t *testing.T) {
	group := &Group{}
	if !group.IsDeletable() {
		t.Error("expected Group to be deletable")
	}
}

func TestGroupConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	group := &Group{}
	group.SetConditions([]metav1.Condition{condition})
	got := group.GetConditions()
	if len(got) != 1 || got[0].Type != "Ready" {
		t.Error("expected Group conditions to be set and retrieved")
	}
}

// newGroupVaultTestClient creates a vault.Client pointed at the given httptest.Server.
func newGroupVaultTestClient(t *testing.T, srv *httptest.Server) *vault.Client {
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
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck // test helper
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
	ctx := vaultutils.ContextWithVaultClient(context.Background(), newGroupVaultTestClient(t, srv))

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
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck // test helper
			"data": map[string]any{
				"name": "my-group", // no "id" key
			},
		})
	}))
	defer srv.Close()

	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "my-group"},
	}
	ctx := vaultutils.ContextWithVaultClient(context.Background(), newGroupVaultTestClient(t, srv))

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
	ctx := vaultutils.ContextWithVaultClient(context.Background(), newGroupVaultTestClient(t, srv))

	if err := group.EnrichStatus(ctx); err == nil {
		t.Error("expected error from Vault 500 response, got nil")
	}
}

func TestMergeUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []string
		expected []string
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: nil,
		},
		{
			name:     "a only",
			a:        []string{"x", "y"},
			b:        nil,
			expected: []string{"x", "y"},
		},
		{
			name:     "b only",
			a:        nil,
			b:        []string{"x", "y"},
			expected: []string{"x", "y"},
		},
		{
			name:     "no overlap",
			a:        []string{"a", "b"},
			b:        []string{"c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "with duplicates",
			a:        []string{"a", "b"},
			b:        []string{"b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "both empty slices",
			a:        []string{},
			b:        []string{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeUniqueStrings(tt.a, tt.b)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected %v at index %d, got %v", tt.expected[i], i, v)
				}
			}
		})
	}
}

func TestGroupToMapWithResolvedNames(t *testing.T) {
	spec := &GroupSpec{
		GroupConfig: GroupConfig{
			Type:                    "internal",
			Metadata:                map[string]string{"team": "platform"},
			Policies:                []string{"default"},
			MemberGroupIDs:          []string{"explicit-group-id"},
			MemberEntityIDs:         []string{"explicit-entity-id"},
			resolvedMemberGroupIDs:  []string{"resolved-group-id"},
			resolvedMemberEntityIDs: []string{"resolved-entity-id"},
		},
	}

	m := spec.toMap()

	groupIDs := m["member_group_ids"].([]string)
	if len(groupIDs) != 2 {
		t.Fatalf("expected 2 member_group_ids, got %d: %v", len(groupIDs), groupIDs)
	}
	if groupIDs[0] != "explicit-group-id" || groupIDs[1] != "resolved-group-id" {
		t.Errorf("unexpected member_group_ids: %v", groupIDs)
	}

	entityIDs := m["member_entity_ids"].([]string)
	if len(entityIDs) != 2 {
		t.Fatalf("expected 2 member_entity_ids, got %d: %v", len(entityIDs), entityIDs)
	}
	if entityIDs[0] != "explicit-entity-id" || entityIDs[1] != "resolved-entity-id" {
		t.Errorf("unexpected member_entity_ids: %v", entityIDs)
	}
}

func TestGroupToMapWithResolvedNamesDeduplicated(t *testing.T) {
	spec := &GroupSpec{
		GroupConfig: GroupConfig{
			Type:                   "internal",
			MemberGroupIDs:         []string{"same-id"},
			resolvedMemberGroupIDs: []string{"same-id"},
		},
	}

	m := spec.toMap()

	groupIDs := m["member_group_ids"].([]string)
	if len(groupIDs) != 1 {
		t.Fatalf("expected deduplication to 1 ID, got %d: %v", len(groupIDs), groupIDs)
	}
	if groupIDs[0] != "same-id" {
		t.Errorf("unexpected member_group_ids: %v", groupIDs)
	}
}

func TestGroupPrepareInternalValuesResolvesNames(t *testing.T) {
	const groupID = "resolved-group-uuid"
	const entityID = "resolved-entity-uuid"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/identity/group/name/child-group":
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck // test helper
				"data": map[string]any{"id": groupID, "name": "child-group"},
			})
		case "/v1/identity/entity/name/my-entity":
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck // test helper
				"data": map[string]any{"id": entityID, "name": "my-entity"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "parent-group"},
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:              "internal",
				MemberGroupNames:  []string{"child-group"},
				MemberEntityNames: []string{"my-entity"},
			},
		},
	}

	ctx := vaultutils.ContextWithVaultClient(context.Background(), newGroupVaultTestClient(t, srv))

	if err := group.PrepareInternalValues(ctx, group); err != nil {
		t.Fatalf("PrepareInternalValues() unexpected error: %v", err)
	}

	payload := group.GetPayload()
	groupIDs := payload["member_group_ids"].([]string)
	if len(groupIDs) != 1 || groupIDs[0] != groupID {
		t.Errorf("expected resolved group ID %q, got %v", groupID, groupIDs)
	}
	entityIDs := payload["member_entity_ids"].([]string)
	if len(entityIDs) != 1 || entityIDs[0] != entityID {
		t.Errorf("expected resolved entity ID %q, got %v", entityID, entityIDs)
	}
}

func TestGroupPrepareInternalValuesSkipsExternal(t *testing.T) {
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "ext-group"},
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:             "external",
				MemberGroupNames: []string{"should-not-resolve"},
			},
		},
	}

	ctx := context.Background()
	if err := group.PrepareInternalValues(ctx, group); err != nil {
		t.Fatalf("PrepareInternalValues() for external group should succeed, got: %v", err)
	}
}

func TestGroupPrepareInternalValuesGroupNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "parent-group"},
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:             "internal",
				MemberGroupNames: []string{"nonexistent-group"},
			},
		},
	}

	ctx := vaultutils.ContextWithVaultClient(context.Background(), newGroupVaultTestClient(t, srv))

	if err := group.PrepareInternalValues(ctx, group); err == nil {
		t.Error("expected error when referenced group does not exist, got nil")
	}
}

func TestGroupPrepareInternalValuesNoNames(t *testing.T) {
	group := &Group{
		ObjectMeta: metav1.ObjectMeta{Name: "simple-group"},
		Spec: GroupSpec{
			GroupConfig: GroupConfig{
				Type:           "internal",
				MemberGroupIDs: []string{"explicit-id"},
			},
		},
	}

	ctx := context.Background()
	if err := group.PrepareInternalValues(ctx, group); err != nil {
		t.Fatalf("PrepareInternalValues() with no names should succeed, got: %v", err)
	}

	payload := group.GetPayload()
	groupIDs := payload["member_group_ids"].([]string)
	if len(groupIDs) != 1 || groupIDs[0] != "explicit-id" {
		t.Errorf("expected only explicit ID, got %v", groupIDs)
	}
}
