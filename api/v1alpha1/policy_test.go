package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPolicyGetPath(t *testing.T) {
	tests := []struct {
		name         string
		policy       *Policy
		expectedPath string
	}{
		{
			name: "with spec.name and spec.type set",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Name: "spec-name", Type: "acl"},
			},
			expectedPath: "sys/policies/acl/spec-name",
		},
		{
			name: "with spec.name set but no type",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Name: "spec-name"},
			},
			expectedPath: "sys/policy/spec-name",
		},
		{
			name: "without spec.name but with type",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Type: "acl"},
			},
			expectedPath: "sys/policies/acl/meta-name",
		},
		{
			name: "without spec.name and without type",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{},
			},
			expectedPath: "sys/policy/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.policy.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestPolicyGetPayload(t *testing.T) {
	policy := &Policy{
		Spec: PolicySpec{
			Policy: `path "secret/*" { capabilities = ["read"] }`,
		},
	}

	payload := policy.GetPayload()

	if len(payload) != 1 {
		t.Errorf("expected exactly 1 key in payload, got %d", len(payload))
	}
	if payload["policy"] != policy.Spec.Policy {
		t.Errorf("expected policy text %q, got %v", policy.Spec.Policy, payload["policy"])
	}
}

func TestPolicyIsEquivalentNoType(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
		},
	}

	// When Type == "", the method renames "policy" → "rules" and adds "name"
	payload := map[string]interface{}{
		"name":  "meta-name",
		"rules": policyText,
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (no type, rules key) to be equivalent")
	}
}

func TestPolicyIsEquivalentWithType(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
			Type:   "acl",
		},
	}

	// When Type == "acl", the method keeps "policy" key and adds "name"
	payload := map[string]interface{}{
		"name":   "meta-name",
		"policy": policyText,
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (with type, policy key) to be equivalent")
	}
}

func TestPolicyIsEquivalentNameFromSpec(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
			Name:   "spec-name",
		},
	}

	payload := map[string]interface{}{
		"name":  "spec-name",
		"rules": policyText,
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected name to come from Spec.Name when set")
	}
}

func TestPolicyIsEquivalentNameFromMetadata(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
		},
	}

	payload := map[string]interface{}{
		"name":  "meta-name",
		"rules": policyText,
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected name to fall back to metadata name when Spec.Name is empty")
	}
}

func TestPolicyIsEquivalentNonMatching(t *testing.T) {
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: `path "secret/*" { capabilities = ["read"] }`,
		},
	}

	payload := map[string]interface{}{
		"name":  "meta-name",
		"rules": `path "secret/*" { capabilities = ["list"] }`,
	}
	if policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected different policy text to not be equivalent")
	}
}

// Policy uses reflect.DeepEqual after mutations, so any extra keys in the
// payload beyond the expected set cause the comparison to return false.
func TestPolicyIsEquivalentExtraFields(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
		},
	}

	payload := map[string]interface{}{
		"name":        "meta-name",
		"rules":       policyText,
		"extra_field": "unexpected",
	}
	if policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (reflect.DeepEqual compares full maps)")
	}
}

func TestPolicyIsDeletable(t *testing.T) {
	policy := &Policy{}
	if !policy.IsDeletable() {
		t.Error("expected Policy to be deletable")
	}
}

func TestPolicyConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	policy := &Policy{}
	policy.SetConditions([]metav1.Condition{condition})
	got := policy.GetConditions()
	if len(got) != 1 || got[0].Type != "Ready" {
		t.Error("expected Policy conditions to be set and retrieved")
	}
}

// Verify the internal structure of the desiredState map produced by
// IsEquivalentToDesiredState across all 4 Type/Name combinations.
func TestPolicyIsEquivalentAllVariants(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`

	tests := []struct {
		name    string
		policy  *Policy
		payload map[string]interface{}
		want    bool
	}{
		{
			name: "no type, spec.name empty → rules key, metadata name",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Policy: policyText},
			},
			payload: map[string]interface{}{"name": "meta-name", "rules": policyText},
			want:    true,
		},
		{
			name: "no type, spec.name set → rules key, spec name",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Policy: policyText, Name: "spec-name"},
			},
			payload: map[string]interface{}{"name": "spec-name", "rules": policyText},
			want:    true,
		},
		{
			name: "type acl, spec.name empty → policy key, metadata name",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Policy: policyText, Type: "acl"},
			},
			payload: map[string]interface{}{"name": "meta-name", "policy": policyText},
			want:    true,
		},
		{
			name: "type acl, spec.name set → policy key, spec name",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Policy: policyText, Type: "acl", Name: "spec-name"},
			},
			payload: map[string]interface{}{"name": "spec-name", "policy": policyText},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policy.IsEquivalentToDesiredState(tt.payload)
			if got != tt.want {
				desiredState := tt.policy.GetPayload()
				desiredState["name"] = map[bool]string{true: tt.policy.Spec.Name, false: tt.policy.Name}[tt.policy.Spec.Name != ""]
				if tt.policy.Spec.Type == "" {
					desiredState["rules"] = desiredState["policy"]
					delete(desiredState, "policy")
				}
				t.Errorf("IsEquivalentToDesiredState() = %v, want %v\ndesiredState: %v\npayload:      %v\nmatch: %v",
					got, tt.want, desiredState, tt.payload, reflect.DeepEqual(desiredState, tt.payload))
			}
		})
	}
}
