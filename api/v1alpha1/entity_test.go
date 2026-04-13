package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEntityGetPath(t *testing.T) {
	tests := []struct {
		name         string
		entity       *Entity
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			entity: &Entity{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       EntitySpec{Name: "spec-name"},
			},
			expectedPath: "identity/entity/name/spec-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			entity: &Entity{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       EntitySpec{},
			},
			expectedPath: "identity/entity/name/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.entity.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestEntityToMap(t *testing.T) {
	spec := &EntitySpec{
		EntityConfig: EntityConfig{
			Metadata: map[string]string{"team": "platform"},
			Policies: []string{"default", "admin"},
			Disabled: false,
		},
	}

	m := spec.toMap()

	if len(m) != 3 {
		t.Errorf("expected 3 keys, got %d", len(m))
	}
	if _, ok := m["metadata"]; !ok {
		t.Error("expected metadata key to be present")
	}
	if _, ok := m["policies"]; !ok {
		t.Error("expected policies key to be present")
	}
	if m["disabled"] != false {
		t.Errorf("expected disabled = false, got %v", m["disabled"])
	}
}

func TestEntityIsEquivalentMatching(t *testing.T) {
	entity := &Entity{
		Spec: EntitySpec{
			EntityConfig: EntityConfig{
				Metadata: map[string]string{"team": "platform"},
				Policies: []string{"default"},
				Disabled: false,
			},
		},
	}

	// All 11 Vault-only keys are deleted before comparison
	payload := map[string]interface{}{
		"metadata":            map[string]string{"team": "platform"},
		"policies":            []string{"default"},
		"disabled":            false,
		"name":                "my-entity",
		"id":                  "entity-id-1",
		"aliases":             []interface{}{},
		"creation_time":       "2024-01-01T00:00:00Z",
		"last_update_time":    "2024-01-02T00:00:00Z",
		"merged_entity_ids":   nil,
		"direct_group_ids":    []interface{}{"g1"},
		"group_ids":           []interface{}{"g1"},
		"inherited_group_ids": nil,
		"namespace_id":        "root",
		"bucket_key_hash":     "abc123",
	}
	if !entity.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (with 11 Vault-only keys) to be equivalent after deletion")
	}
}

func TestEntityIsEquivalentNonMatching(t *testing.T) {
	entity := &Entity{
		Spec: EntitySpec{
			EntityConfig: EntityConfig{
				Metadata: map[string]string{"team": "platform"},
				Policies: []string{"default"},
				Disabled: false,
			},
		},
	}

	payload := map[string]interface{}{
		"metadata": map[string]string{"team": "platform"},
		"policies": []string{"default"},
		"disabled": true,
	}
	if entity.IsEquivalentToDesiredState(payload) {
		t.Error("expected different disabled value to not be equivalent")
	}
}

// Entity only deletes 11 specific Vault keys. Any other extra keys
// cause reflect.DeepEqual to return false.
func TestEntityIsEquivalentExtraFields(t *testing.T) {
	entity := &Entity{
		Spec: EntitySpec{
			EntityConfig: EntityConfig{
				Metadata: map[string]string{"team": "platform"},
				Policies: []string{"default"},
				Disabled: false,
			},
		},
	}

	payload := map[string]interface{}{
		"metadata":      map[string]string{"team": "platform"},
		"policies":      []string{"default"},
		"disabled":      false,
		"unknown_field": "unexpected",
	}
	if entity.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra keys beyond the 11 known Vault keys to cause DeepEqual to return false")
	}
}

func TestEntityGetPayload(t *testing.T) {
	entity := &Entity{
		Spec: EntitySpec{
			EntityConfig: EntityConfig{
				Metadata: map[string]string{"team": "platform"},
				Policies: []string{"default"},
				Disabled: true,
			},
		},
	}

	payload := entity.GetPayload()
	if payload["disabled"] != true {
		t.Errorf("GetPayload() should delegate to Spec.toMap(), got disabled = %v", payload["disabled"])
	}
}

func TestEntityIsDeletable(t *testing.T) {
	entity := &Entity{}
	if !entity.IsDeletable() {
		t.Error("expected Entity to be deletable")
	}
}

func TestEntityConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	entity := &Entity{}
	entity.SetConditions([]metav1.Condition{condition})
	got := entity.GetConditions()
	if len(got) != 1 || got[0].Type != "Ready" {
		t.Error("expected Entity conditions to be set and retrieved")
	}
}
