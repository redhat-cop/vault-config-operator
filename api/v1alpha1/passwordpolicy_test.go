package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPasswordPolicyGetPath(t *testing.T) {
	tests := []struct {
		name         string
		policy       *PasswordPolicy
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			policy: &PasswordPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: PasswordPolicySpec{
					Name: "custom-name",
				},
			},
			expectedPath: "sys/policies/password/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			policy: &PasswordPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: PasswordPolicySpec{},
			},
			expectedPath: "sys/policies/password/test-policy",
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

func TestPasswordPolicyGetPayload(t *testing.T) {
	hcl := `length = 20
rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 1
}`
	policy := &PasswordPolicy{
		Spec: PasswordPolicySpec{
			PasswordPolicy: hcl,
		},
	}

	payload := policy.GetPayload()

	if payload["policy"] != hcl {
		t.Errorf("expected policy HCL content, got %v", payload["policy"])
	}
	if len(payload) != 1 {
		t.Errorf("expected exactly 1 key in payload, got %d", len(payload))
	}
}

func TestPasswordPolicyIsEquivalentToDesiredState(t *testing.T) {
	hcl := `length = 20
rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 1
}`
	policy := &PasswordPolicy{
		Spec: PasswordPolicySpec{
			PasswordPolicy: hcl,
		},
	}

	matching := map[string]interface{}{
		"policy": hcl,
	}
	if !policy.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]interface{}{
		"policy": "length = 10",
	}
	if policy.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

// PasswordPolicy uses reflect.DeepEqual(GetPayload(), payload), so extra keys
// in the payload cause the comparison to return false. The reconciler passes
// the raw Vault read response with no key filtering. Story 7-4 tracks
// hardening this behavior.
func TestPasswordPolicyIsEquivalentExtraFieldsReturnsFalse(t *testing.T) {
	policy := &PasswordPolicy{
		Spec: PasswordPolicySpec{
			PasswordPolicy: "length = 20",
		},
	}

	payloadWithExtra := map[string]interface{}{
		"policy":      "length = 20",
		"extra_field": "vault-returned-value",
	}
	if policy.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected payload with extra fields to NOT be equivalent (reflect.DeepEqual compares full maps)")
	}
}

func TestPasswordPolicyIsDeletable(t *testing.T) {
	policy := &PasswordPolicy{}
	if !policy.IsDeletable() {
		t.Error("expected PasswordPolicy to be deletable")
	}
}

func TestPasswordPolicyConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	policy := &PasswordPolicy{}
	policy.SetConditions([]metav1.Condition{condition})
	if len(policy.GetConditions()) != 1 || policy.GetConditions()[0].Type != "Ready" {
		t.Error("expected PasswordPolicy conditions to be set and retrieved")
	}
}
