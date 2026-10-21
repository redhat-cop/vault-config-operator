package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRotationPolicyGetPath(t *testing.T) {
	tests := []struct {
		name         string
		policy       *RotationPolicy
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			policy: &RotationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: RotationPolicySpec{
					Name: "custom-name",
				},
			},
			expectedPath: "sys/policies/rotation/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			policy: &RotationPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: RotationPolicySpec{},
			},
			expectedPath: "sys/policies/rotation/test-policy",
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

func TestRotationPolicyGetPayload(t *testing.T) {
	policy := &RotationPolicy{
		Spec: RotationPolicySpec{
			RotationPolicyConfig: RotationPolicyConfig{
				MaxRetriesPerCycle: 3,
				MaxRetryCycles:     5,
			},
		},
	}

	payload := policy.GetPayload()

	if payload["max_retries_per_cycle"] != json.Number("3") {
		t.Errorf("expected max_retries_per_cycle = json.Number(\"3\"), got %v (%T)", payload["max_retries_per_cycle"], payload["max_retries_per_cycle"])
	}
	if payload["max_retry_cycles"] != json.Number("5") {
		t.Errorf("expected max_retry_cycles = json.Number(\"5\"), got %v (%T)", payload["max_retry_cycles"], payload["max_retry_cycles"])
	}
	if len(payload) != 2 {
		t.Errorf("expected exactly 2 keys in payload, got %d", len(payload))
	}
}

func TestRotationPolicyIsEquivalentToDesiredState(t *testing.T) {
	policy := &RotationPolicy{
		Spec: RotationPolicySpec{
			RotationPolicyConfig: RotationPolicyConfig{
				MaxRetriesPerCycle: 3,
				MaxRetryCycles:     5,
			},
		},
	}

	vaultPayload := map[string]any{
		"max_retries_per_cycle": json.Number("3"),
		"max_retry_cycles":      json.Number("5"),
	}
	if !policy.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestRotationPolicyIsEquivalentNegative(t *testing.T) {
	policy := &RotationPolicy{
		Spec: RotationPolicySpec{
			RotationPolicyConfig: RotationPolicyConfig{
				MaxRetriesPerCycle: 3,
				MaxRetryCycles:     5,
			},
		},
	}

	nonMatching := map[string]any{
		"max_retries_per_cycle": json.Number("3"),
		"max_retry_cycles":      json.Number("10"),
	}
	if policy.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestRotationPolicyIsEquivalentExtraFields(t *testing.T) {
	policy := &RotationPolicy{
		Spec: RotationPolicySpec{
			RotationPolicyConfig: RotationPolicyConfig{
				MaxRetriesPerCycle: 3,
				MaxRetryCycles:     5,
			},
		},
	}

	payloadWithExtra := map[string]any{
		"max_retries_per_cycle": json.Number("3"),
		"max_retry_cycles":      json.Number("5"),
		"extra_field":           "vault-returned-value",
	}
	if !policy.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestRotationPolicyIsDeletable(t *testing.T) {
	policy := &RotationPolicy{}
	if !policy.IsDeletable() {
		t.Error("expected RotationPolicy to be deletable")
	}
}

func TestRotationPolicyConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	policy := &RotationPolicy{}
	policy.SetConditions([]metav1.Condition{condition})
	if len(policy.GetConditions()) != 1 || policy.GetConditions()[0].Type != "Ready" {
		t.Error("expected RotationPolicy conditions to be set and retrieved")
	}
}
