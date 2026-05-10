package v1alpha1

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGroupAliasGetPath(t *testing.T) {
	ga := &GroupAlias{
		Status: GroupAliasStatus{
			ID: "abc-123",
		},
	}

	result := ga.GetPath()
	expected := "identity/group-alias/id/abc-123"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestGroupAliasGetPathEmptyID(t *testing.T) {
	ga := &GroupAlias{}
	result := ga.GetPath()
	expected := "identity/group-alias/id"
	if result != expected {
		t.Errorf("GetPath() with empty ID = %v, expected %v", result, expected)
	}
}

func TestGroupAliasToMap(t *testing.T) {
	spec := &GroupAliasSpec{
		retrievedName:          "my-alias",
		retrievedAliasID:       "alias-id-1",
		retrievedMountAccessor: "auth_kubernetes_abc",
		retrievedCanonicalID:   "canonical-id-1",
	}

	m := spec.toMap()

	if len(m) != 4 {
		t.Errorf("expected 4 keys, got %d", len(m))
	}
	if m["name"] != "my-alias" {
		t.Errorf("expected name 'my-alias', got %v", m["name"])
	}
	if m["id"] != "alias-id-1" {
		t.Errorf("expected id 'alias-id-1', got %v", m["id"])
	}
	if m["mount_accessor"] != "auth_kubernetes_abc" {
		t.Errorf("expected mount_accessor 'auth_kubernetes_abc', got %v", m["mount_accessor"])
	}
	if m["canonical_id"] != "canonical-id-1" {
		t.Errorf("expected canonical_id 'canonical-id-1', got %v", m["canonical_id"])
	}
}

func TestGroupAliasIsEquivalentMatching(t *testing.T) {
	ga := &GroupAlias{
		Spec: GroupAliasSpec{
			retrievedName:          "my-alias",
			retrievedAliasID:       "alias-id-1",
			retrievedMountAccessor: "auth_kubernetes_abc",
			retrievedCanonicalID:   "canonical-id-1",
		},
	}

	// 6 Vault-only keys are deleted before comparison
	payload := map[string]interface{}{
		"name":                      "my-alias",
		"id":                        "alias-id-1",
		"mount_accessor":            "auth_kubernetes_abc",
		"canonical_id":              "canonical-id-1",
		"creation_time":             "2024-01-01T00:00:00Z",
		"last_update_time":          "2024-01-02T00:00:00Z",
		"merged_from_canonical_ids": nil,
		"metadata":                  nil,
		"mount_path":                "auth/kubernetes/",
		"mount_type":                "kubernetes",
	}
	if !ga.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (with 6 Vault-only keys) to be equivalent after deletion")
	}
}

func TestGroupAliasIsEquivalentNonMatching(t *testing.T) {
	ga := &GroupAlias{
		Spec: GroupAliasSpec{
			retrievedName:          "my-alias",
			retrievedAliasID:       "alias-id-1",
			retrievedMountAccessor: "auth_kubernetes_abc",
			retrievedCanonicalID:   "canonical-id-1",
		},
	}

	payload := map[string]interface{}{
		"name":           "my-alias",
		"id":             "alias-id-1",
		"mount_accessor": "auth_different_accessor",
		"canonical_id":   "canonical-id-1",
	}
	if ga.IsEquivalentToDesiredState(payload) {
		t.Error("expected different mount_accessor to not be equivalent")
	}
}

// Payload keys not present in desiredState are filtered before comparison.
func TestGroupAliasIsEquivalentExtraFields(t *testing.T) {
	ga := &GroupAlias{
		Spec: GroupAliasSpec{
			retrievedName:          "my-alias",
			retrievedAliasID:       "alias-id-1",
			retrievedMountAccessor: "auth_kubernetes_abc",
			retrievedCanonicalID:   "canonical-id-1",
		},
	}

	payload := map[string]interface{}{
		"name":           "my-alias",
		"id":             "alias-id-1",
		"mount_accessor": "auth_kubernetes_abc",
		"canonical_id":   "canonical-id-1",
		"unknown_field":  "unexpected",
	}
	if !ga.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestGroupAliasGetPayload(t *testing.T) {
	ga := &GroupAlias{
		Spec: GroupAliasSpec{
			retrievedName:          "my-alias",
			retrievedAliasID:       "alias-id-1",
			retrievedMountAccessor: "auth_kubernetes_abc",
			retrievedCanonicalID:   "canonical-id-1",
		},
	}

	payload := ga.GetPayload()
	if payload["name"] != "my-alias" {
		t.Errorf("GetPayload() should delegate to Spec.toMap(), got name = %v", payload["name"])
	}
}

func TestGroupAliasIsDeletable(t *testing.T) {
	ga := &GroupAlias{}
	if !ga.IsDeletable() {
		t.Error("expected GroupAlias to be deletable")
	}
}

func TestGroupAliasConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	ga := &GroupAlias{}
	ga.SetConditions([]metav1.Condition{condition})
	got := ga.GetConditions()
	if len(got) != 1 || got[0].Type != "Ready" {
		t.Error("expected GroupAlias conditions to be set and retrieved")
	}
}

func TestGroupAlias_PrepareInternalValues_SuccessWithPrepopulatedStatusID(t *testing.T) {
	ga := &GroupAlias{
		ObjectMeta: metav1.ObjectMeta{Name: "group-alias-res", Namespace: "default"},
		Spec: GroupAliasSpec{
			GroupAliasConfig: GroupAliasConfig{
				AuthEngineMountPath: "kubernetes",
				GroupName:           "my-group",
			},
		},
		Status: GroupAliasStatus{ID: "existing-group-alias-id"},
	}
	handler := newFakeVaultHandler()
	handler.setGet("sys/auth/kubernetes", map[string]interface{}{
		"accessor": "auth_k8s_1234",
		"type":     "kubernetes",
	})
	handler.setGet("identity/group/name/my-group", map[string]interface{}{
		"id": "group-uuid-567",
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	if err := ga.PrepareInternalValues(ctx, ga); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if ga.Spec.retrievedMountAccessor != "auth_k8s_1234" {
		t.Errorf("retrievedMountAccessor: got %q", ga.Spec.retrievedMountAccessor)
	}
	if ga.Spec.retrievedCanonicalID != "group-uuid-567" {
		t.Errorf("retrievedCanonicalID: got %q", ga.Spec.retrievedCanonicalID)
	}
	if ga.Spec.retrievedAliasID != "existing-group-alias-id" {
		t.Errorf("retrievedAliasID: got %q", ga.Spec.retrievedAliasID)
	}
	if ga.Spec.retrievedName != ga.Name {
		t.Errorf("retrievedName: got %q want %q", ga.Spec.retrievedName, ga.Name)
	}
}

func TestGroupAlias_PrepareInternalValues_SpecNameOverride(t *testing.T) {
	ga := &GroupAlias{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name", Namespace: "default"},
		Spec: GroupAliasSpec{
			Name: "custom-group-alias",
			GroupAliasConfig: GroupAliasConfig{
				AuthEngineMountPath: "kubernetes",
				GroupName:           "my-group",
			},
		},
		Status: GroupAliasStatus{ID: "existing-group-alias-id"},
	}
	handler := newFakeVaultHandler()
	handler.setGet("sys/auth/kubernetes", map[string]interface{}{
		"accessor": "auth_k8s_1234",
		"type":     "kubernetes",
	})
	handler.setGet("identity/group/name/my-group", map[string]interface{}{
		"id": "group-uuid-567",
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	if err := ga.PrepareInternalValues(ctx, ga); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if ga.Spec.retrievedName != "custom-group-alias" {
		t.Errorf("retrievedName: got %q", ga.Spec.retrievedName)
	}
}

func TestGroupAlias_PrepareInternalValues_GroupNotFound(t *testing.T) {
	ga := &GroupAlias{
		ObjectMeta: metav1.ObjectMeta{Name: "g", Namespace: "default"},
		Spec: GroupAliasSpec{
			GroupAliasConfig: GroupAliasConfig{
				AuthEngineMountPath: "kubernetes",
				GroupName:           "missing-group",
			},
		},
		Status: GroupAliasStatus{ID: "id"},
	}
	handler := newFakeVaultHandler()
	handler.setGet("sys/auth/kubernetes", map[string]interface{}{
		"accessor": "auth_k8s_1234",
		"type":     "kubernetes",
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	err := ga.PrepareInternalValues(ctx, ga)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "group not found") {
		t.Fatalf("err: %v", err)
	}
}
