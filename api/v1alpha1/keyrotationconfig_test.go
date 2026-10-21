package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKeyRotationConfigGetPath(t *testing.T) {
	kr := &KeyRotationConfig{}
	expected := "sys/rotate/config"
	if kr.GetPath() != expected {
		t.Errorf("GetPath() = %v, expected %v", kr.GetPath(), expected)
	}
}

func TestKeyRotationConfigIsDeletable(t *testing.T) {
	kr := &KeyRotationConfig{}
	if kr.IsDeletable() {
		t.Error("expected KeyRotationConfig to not be deletable")
	}
}

func TestKeyRotationConfigToMap(t *testing.T) {
	tests := []struct {
		name     string
		data     KeyRotationConfigConfig
		expected map[string]any
	}{
		{
			name: "all fields set",
			data: KeyRotationConfigConfig{
				MaxOperations: 3888000000,
				Interval:      "4320h",
				Enabled:       true,
			},
			expected: map[string]any{
				"max_operations": json.Number("3888000000"),
				"interval":       "4320h",
				"enabled":        true,
			},
		},
		{
			name: "enabled false",
			data: KeyRotationConfigConfig{
				Enabled: false,
			},
			expected: map[string]any{
				"enabled": false,
			},
		},
		{
			name: "only interval",
			data: KeyRotationConfigConfig{
				Interval: "720h",
				Enabled:  true,
			},
			expected: map[string]any{
				"interval": "720h",
				"enabled":  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.toMap()

			for key, expectedVal := range tt.expected {
				gotVal, ok := result[key]
				if !ok {
					t.Errorf("toMap() missing key %q", key)
					continue
				}
				gotJSON, _ := json.Marshal(gotVal)
				expectedJSON, _ := json.Marshal(expectedVal)
				if string(gotJSON) != string(expectedJSON) {
					t.Errorf("toMap()[%q] = %s, expected %s", key, gotJSON, expectedJSON)
				}
			}

			// Verify empty MaxOperations not included
			if tt.data.MaxOperations == 0 {
				if _, ok := result["max_operations"]; ok {
					t.Error("toMap() should not include max_operations when zero")
				}
			}

			// Verify empty Interval not included
			if tt.data.Interval == "" {
				if _, ok := result["interval"]; ok {
					t.Error("toMap() should not include interval when empty")
				}
			}
		})
	}
}

func TestKeyRotationConfigToMapJsonNumber(t *testing.T) {
	data := KeyRotationConfigConfig{
		MaxOperations: 100,
		Enabled:       true,
	}
	result := data.toMap()
	val, ok := result["max_operations"]
	if !ok {
		t.Fatal("expected max_operations in toMap()")
	}
	num, ok := val.(json.Number)
	if !ok {
		t.Fatalf("expected max_operations to be json.Number, got %T", val)
	}
	if num.String() != "100" {
		t.Errorf("expected json.Number '100', got %q", num.String())
	}
}

func TestKeyRotationConfigToMapEnabledFalse(t *testing.T) {
	data := KeyRotationConfigConfig{
		Enabled: false,
	}
	result := data.toMap()
	val, ok := result["enabled"]
	if !ok {
		t.Fatal("expected 'enabled' in toMap()")
	}
	b, ok := val.(bool)
	if !ok {
		t.Fatalf("expected enabled to be bool, got %T", val)
	}
	if b != false {
		t.Error("expected enabled to be false")
	}
}

func TestKeyRotationConfigConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	kr := &KeyRotationConfig{}
	kr.SetConditions([]metav1.Condition{condition})
	if len(kr.GetConditions()) != 1 || kr.GetConditions()[0].Type != "Ready" {
		t.Error("expected KeyRotationConfig conditions to be set and retrieved")
	}
}

func TestKeyRotationConfigIsEquivalentToDesiredStatePositive(t *testing.T) {
	kr := &KeyRotationConfig{
		Spec: KeyRotationConfigSpec{
			KeyRotationConfigConfig: KeyRotationConfigConfig{
				MaxOperations: 100,
				Interval:      "4320h",
				Enabled:       true,
			},
		},
	}

	payload := map[string]any{
		"max_operations": json.Number("100"),
		"interval":       "4320h",
		"enabled":        true,
	}
	if !kr.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestKeyRotationConfigIsEquivalentToDesiredStateNegative(t *testing.T) {
	kr := &KeyRotationConfig{
		Spec: KeyRotationConfigSpec{
			KeyRotationConfigConfig: KeyRotationConfigConfig{
				MaxOperations: 100,
				Enabled:       true,
			},
		},
	}

	payload := map[string]any{
		"max_operations": json.Number("200"),
		"enabled":        true,
	}
	if kr.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching max_operations to not be equivalent")
	}
}

func TestKeyRotationConfigIsEquivalentExtraFields(t *testing.T) {
	kr := &KeyRotationConfig{
		Spec: KeyRotationConfigSpec{
			KeyRotationConfigConfig: KeyRotationConfigConfig{
				Enabled: true,
			},
		},
	}

	payload := map[string]any{
		"enabled":            true,
		"request_id":         "vault-returned-id",
		"last_rotation_time": "2024-01-01T00:00:00Z",
	}
	if !kr.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra fields to be ignored")
	}
}

func TestKeyRotationConfigIsEquivalentEmptyIntervalOmitted(t *testing.T) {
	kr := &KeyRotationConfig{
		Spec: KeyRotationConfigSpec{
			KeyRotationConfigConfig: KeyRotationConfigConfig{
				Enabled: true,
			},
		},
	}

	payload := map[string]any{
		"enabled":  true,
		"interval": "4320h",
	}
	if !kr.IsEquivalentToDesiredState(payload) {
		t.Error("expected empty interval to be omitted from desired state comparison")
	}
}
