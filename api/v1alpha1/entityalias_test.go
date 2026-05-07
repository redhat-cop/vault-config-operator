package v1alpha1

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEntityAliasGetPath(t *testing.T) {
	ea := &EntityAlias{
		Status: EntityAliasStatus{
			ID: "abc-123",
		},
	}

	result := ea.GetPath()
	expected := "identity/entity-alias/id/abc-123"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestEntityAliasGetPathEmptyID(t *testing.T) {
	ea := &EntityAlias{}
	result := ea.GetPath()
	expected := "identity/entity-alias/id"
	if result != expected {
		t.Errorf("GetPath() with empty ID = %v, expected %v", result, expected)
	}
}

func TestEntityAliasToMapWithCustomMetadata(t *testing.T) {
	spec := &EntityAliasSpec{
		retrievedName:          "my-alias",
		retrievedAliasID:       "alias-id-1",
		retrievedMountAccessor: "auth_kubernetes_abc",
		retrievedCanonicalID:   "canonical-id-1",
		EntityAliasConfig: EntityAliasConfig{
			CustomMetadata: map[string]string{"env": "prod", "team": "platform"},
		},
	}

	m := spec.toMap()

	if len(m) != 5 {
		t.Errorf("expected 5 keys (with custom_metadata), got %d", len(m))
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
	cm, ok := m["custom_metadata"].(map[string]string)
	if !ok {
		t.Fatal("expected custom_metadata to be map[string]string")
	}
	if cm["env"] != "prod" {
		t.Errorf("expected custom_metadata[env] = 'prod', got %v", cm["env"])
	}
}

func TestEntityAliasToMapWithoutCustomMetadata(t *testing.T) {
	spec := &EntityAliasSpec{
		retrievedName:          "my-alias",
		retrievedAliasID:       "alias-id-1",
		retrievedMountAccessor: "auth_kubernetes_abc",
		retrievedCanonicalID:   "canonical-id-1",
	}

	m := spec.toMap()

	if len(m) != 4 {
		t.Errorf("expected 4 keys (without custom_metadata), got %d", len(m))
	}
	if _, ok := m["custom_metadata"]; ok {
		t.Error("expected custom_metadata to be absent when CustomMetadata is empty")
	}
}

func TestEntityAliasIsEquivalentMatching(t *testing.T) {
	ea := &EntityAlias{
		Spec: EntityAliasSpec{
			retrievedName:          "my-alias",
			retrievedAliasID:       "alias-id-1",
			retrievedMountAccessor: "auth_kubernetes_abc",
			retrievedCanonicalID:   "canonical-id-1",
		},
	}

	// 8 Vault-only keys are deleted before comparison
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
		"local":                     false,
		"namespace_id":              "root",
	}
	if !ea.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (with 8 Vault-only keys) to be equivalent after deletion")
	}
}

func TestEntityAliasIsEquivalentMatchingWithCustomMetadata(t *testing.T) {
	ea := &EntityAlias{
		Spec: EntityAliasSpec{
			retrievedName:          "my-alias",
			retrievedAliasID:       "alias-id-1",
			retrievedMountAccessor: "auth_kubernetes_abc",
			retrievedCanonicalID:   "canonical-id-1",
			EntityAliasConfig: EntityAliasConfig{
				CustomMetadata: map[string]string{"env": "prod"},
			},
		},
	}

	payload := map[string]interface{}{
		"name":            "my-alias",
		"id":              "alias-id-1",
		"mount_accessor":  "auth_kubernetes_abc",
		"canonical_id":    "canonical-id-1",
		"custom_metadata": map[string]string{"env": "prod"},
		"creation_time":   "2024-01-01T00:00:00Z",
		"local":           false,
	}
	if !ea.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload with custom_metadata to be equivalent")
	}
}

func TestEntityAliasIsEquivalentNonMatching(t *testing.T) {
	ea := &EntityAlias{
		Spec: EntityAliasSpec{
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
	if ea.IsEquivalentToDesiredState(payload) {
		t.Error("expected different mount_accessor to not be equivalent")
	}
}

// EntityAlias only deletes 8 specific Vault keys. Any other extra keys
// cause reflect.DeepEqual to return false.
func TestEntityAliasIsEquivalentExtraFields(t *testing.T) {
	ea := &EntityAlias{
		Spec: EntityAliasSpec{
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
	if ea.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra keys beyond the 8 known Vault keys to cause DeepEqual to return false")
	}
}

func TestEntityAliasGetPayload(t *testing.T) {
	ea := &EntityAlias{
		Spec: EntityAliasSpec{
			retrievedName:          "my-alias",
			retrievedAliasID:       "alias-id-1",
			retrievedMountAccessor: "auth_kubernetes_abc",
			retrievedCanonicalID:   "canonical-id-1",
		},
	}

	payload := ea.GetPayload()
	if payload["name"] != "my-alias" {
		t.Errorf("GetPayload() should delegate to Spec.toMap(), got name = %v", payload["name"])
	}
}

func TestEntityAliasIsDeletable(t *testing.T) {
	ea := &EntityAlias{}
	if !ea.IsDeletable() {
		t.Error("expected EntityAlias to be deletable")
	}
}

func TestEntityAliasConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	ea := &EntityAlias{}
	ea.SetConditions([]metav1.Condition{condition})
	got := ea.GetConditions()
	if len(got) != 1 || got[0].Type != "Ready" {
		t.Error("expected EntityAlias conditions to be set and retrieved")
	}
}

func TestEntityAlias_PrepareInternalValues_SuccessWithPrepopulatedStatusID(t *testing.T) {
	ea := &EntityAlias{
		ObjectMeta: metav1.ObjectMeta{Name: "alias-res", Namespace: "default"},
		Spec: EntityAliasSpec{
			EntityAliasConfig: EntityAliasConfig{
				AuthEngineMountPath: "kubernetes",
				EntityName:          "my-entity",
			},
		},
		Status: EntityAliasStatus{ID: "existing-alias-id"},
	}
	handler := newFakeVaultHandler()
	handler.setGet("sys/auth/kubernetes", map[string]interface{}{
		"accessor": "auth_k8s_1234",
		"type":     "kubernetes",
	})
	handler.setGet("identity/entity/name/my-entity", map[string]interface{}{
		"id": "entity-uuid-567",
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	if err := ea.PrepareInternalValues(ctx, ea); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if ea.Spec.retrievedMountAccessor != "auth_k8s_1234" {
		t.Errorf("retrievedMountAccessor: got %q", ea.Spec.retrievedMountAccessor)
	}
	if ea.Spec.retrievedCanonicalID != "entity-uuid-567" {
		t.Errorf("retrievedCanonicalID: got %q", ea.Spec.retrievedCanonicalID)
	}
	if ea.Spec.retrievedAliasID != "existing-alias-id" {
		t.Errorf("retrievedAliasID: got %q", ea.Spec.retrievedAliasID)
	}
	if ea.Spec.retrievedName != ea.Name {
		t.Errorf("retrievedName: got %q want %q", ea.Spec.retrievedName, ea.Name)
	}
}

func TestEntityAlias_PrepareInternalValues_SpecNameOverride(t *testing.T) {
	ea := &EntityAlias{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-alias", Namespace: "default"},
		Spec: EntityAliasSpec{
			Name: "custom-alias",
			EntityAliasConfig: EntityAliasConfig{
				AuthEngineMountPath: "kubernetes",
				EntityName:          "my-entity",
			},
		},
		Status: EntityAliasStatus{ID: "existing-alias-id"},
	}
	handler := newFakeVaultHandler()
	handler.setGet("sys/auth/kubernetes", map[string]interface{}{
		"accessor": "auth_k8s_1234",
		"type":     "kubernetes",
	})
	handler.setGet("identity/entity/name/my-entity", map[string]interface{}{
		"id": "entity-uuid-567",
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	if err := ea.PrepareInternalValues(ctx, ea); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if ea.Spec.retrievedName != "custom-alias" {
		t.Errorf("retrievedName: got %q want custom-alias", ea.Spec.retrievedName)
	}
}

func TestEntityAlias_PrepareInternalValues_AuthEngineNotFound(t *testing.T) {
	ea := &EntityAlias{
		ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: "default"},
		Spec: EntityAliasSpec{
			EntityAliasConfig: EntityAliasConfig{
				AuthEngineMountPath: "kubernetes",
				EntityName:          "my-entity",
			},
		},
		Status: EntityAliasStatus{ID: "id"},
	}
	handler := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	err := ea.PrepareInternalValues(ctx, ea)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "auth engine not found") {
		t.Fatalf("err: %v", err)
	}
}

func TestEntityAlias_PrepareInternalValues_EntityNotFound(t *testing.T) {
	ea := &EntityAlias{
		ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: "default"},
		Spec: EntityAliasSpec{
			EntityAliasConfig: EntityAliasConfig{
				AuthEngineMountPath: "kubernetes",
				EntityName:          "my-entity",
			},
		},
		Status: EntityAliasStatus{ID: "id"},
	}
	handler := newFakeVaultHandler()
	handler.setGet("sys/auth/kubernetes", map[string]interface{}{
		"accessor": "auth_k8s_1234",
		"type":     "kubernetes",
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	err := ea.PrepareInternalValues(ctx, ea)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "entity not found") {
		t.Fatalf("err: %v", err)
	}
}
