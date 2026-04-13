package v1alpha1

import (
	"testing"

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
	payload := map[string]interface{}{
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

	payload := map[string]interface{}{
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

// Group only deletes "name" from the payload. Any other extra keys cause
// reflect.DeepEqual to return false.
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

	payload := map[string]interface{}{
		"name":              "some-group",
		"type":              "internal",
		"metadata":          map[string]string{"team": "platform"},
		"policies":          []string{"default"},
		"member_group_ids":  []string{"group-1"},
		"member_entity_ids": []string{"entity-1"},
		"extra_field":       "unexpected",
	}
	if group.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra keys beyond 'name' to cause DeepEqual to return false")
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
