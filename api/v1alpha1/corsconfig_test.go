package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCORSConfigGetPath(t *testing.T) {
	cc := &CORSConfig{}
	expected := "sys/config/cors"
	if cc.GetPath() != expected {
		t.Errorf("GetPath() = %v, expected %v", cc.GetPath(), expected)
	}
}

func TestCORSConfigIsDeletable(t *testing.T) {
	cc := &CORSConfig{}
	if !cc.IsDeletable() {
		t.Error("expected CORSConfig to be deletable")
	}
}

func TestCORSConfigToMap(t *testing.T) {
	tests := []struct {
		name     string
		data     CORSConfigConfig
		expected map[string]any
	}{
		{
			name: "origins and headers",
			data: CORSConfigConfig{
				AllowedOrigins: []string{"https://example.com", "https://other.com"},
				AllowedHeaders: []string{"X-Custom-Header"},
			},
			expected: map[string]any{
				"allowed_origins": []any{"https://example.com", "https://other.com"},
				"allowed_headers": []any{"X-Custom-Header"},
			},
		},
		{
			name: "origins only — empty headers omitted",
			data: CORSConfigConfig{
				AllowedOrigins: []string{"*"},
			},
			expected: map[string]any{
				"allowed_origins": []any{"*"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.toMap()

			// Verify correct number of keys
			if len(result) != len(tt.expected) {
				t.Errorf("toMap() returned %d keys, expected %d", len(result), len(tt.expected))
			}

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

			// Verify no "enabled" key exists
			if _, ok := result["enabled"]; ok {
				t.Error("toMap() must not include 'enabled' for CORSConfig")
			}
		})
	}
}

func TestCORSConfigToMapEmptyHeadersOmitted(t *testing.T) {
	data := CORSConfigConfig{
		AllowedOrigins: []string{"https://example.com"},
		AllowedHeaders: []string{},
	}
	result := data.toMap()
	if _, ok := result["allowed_headers"]; ok {
		t.Error("toMap() should not include allowed_headers when empty")
	}
}

func TestCORSConfigConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	cc := &CORSConfig{}
	cc.SetConditions([]metav1.Condition{condition})
	if len(cc.GetConditions()) != 1 || cc.GetConditions()[0].Type != "Ready" {
		t.Error("expected CORSConfig conditions to be set and retrieved")
	}
}

func TestCORSConfigIsEquivalentToDesiredStatePositive(t *testing.T) {
	cc := &CORSConfig{
		Spec: CORSConfigSpec{
			CORSConfigConfig: CORSConfigConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedHeaders: []string{"X-Custom"},
			},
		},
	}

	payload := map[string]any{
		"allowed_origins": []any{"https://example.com"},
		"allowed_headers": []any{"X-Custom", "Content-Type", "Authorization"},
		"enabled":         true,
	}
	if !cc.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent (subset check on headers)")
	}
}

func TestCORSConfigIsEquivalentToDesiredStateNegative(t *testing.T) {
	cc := &CORSConfig{
		Spec: CORSConfigSpec{
			CORSConfigConfig: CORSConfigConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedHeaders: []string{"X-Custom"},
			},
		},
	}

	payload := map[string]any{
		"allowed_origins": []any{"https://different.com"},
		"allowed_headers": []any{"X-Custom", "Content-Type"},
	}
	if cc.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching origins to be not equivalent")
	}
}

func TestCORSConfigIsEquivalentExtraFields(t *testing.T) {
	cc := &CORSConfig{
		Spec: CORSConfigSpec{
			CORSConfigConfig: CORSConfigConfig{
				AllowedOrigins: []string{"https://example.com"},
			},
		},
	}

	payload := map[string]any{
		"allowed_origins": []any{"https://example.com"},
		"enabled":         true,
		"request_id":      "vault-returned-id",
	}
	if !cc.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra fields to be ignored")
	}
}

func TestCORSConfigIsEquivalentVaultAppendedHeaders(t *testing.T) {
	cc := &CORSConfig{
		Spec: CORSConfigSpec{
			CORSConfigConfig: CORSConfigConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedHeaders: []string{"X-Custom"},
			},
		},
	}

	// Vault returns desired headers plus its own defaults
	payload := map[string]any{
		"allowed_origins": []any{"https://example.com"},
		"allowed_headers": []any{"Content-Type", "X-Requested-With", "X-Custom", "Authorization"},
	}
	if !cc.IsEquivalentToDesiredState(payload) {
		t.Error("expected Vault-appended default headers to not cause false drift")
	}
}

func TestCORSConfigIsEquivalentMissingDesiredHeader(t *testing.T) {
	cc := &CORSConfig{
		Spec: CORSConfigSpec{
			CORSConfigConfig: CORSConfigConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedHeaders: []string{"X-Custom", "X-Missing"},
			},
		},
	}

	payload := map[string]any{
		"allowed_origins": []any{"https://example.com"},
		"allowed_headers": []any{"Content-Type", "X-Custom"},
	}
	if cc.IsEquivalentToDesiredState(payload) {
		t.Error("expected missing desired header to not be equivalent")
	}
}

func TestCORSConfigIsEquivalentOriginsOrderInsensitive(t *testing.T) {
	cc := &CORSConfig{
		Spec: CORSConfigSpec{
			CORSConfigConfig: CORSConfigConfig{
				AllowedOrigins: []string{"https://b.com", "https://a.com"},
			},
		},
	}

	payload := map[string]any{
		"allowed_origins": []any{"https://a.com", "https://b.com"},
	}
	if !cc.IsEquivalentToDesiredState(payload) {
		t.Error("expected origins to be order-insensitive")
	}
}

func TestCORSConfigIsEquivalentEmptyHeaders(t *testing.T) {
	cc := &CORSConfig{
		Spec: CORSConfigSpec{
			CORSConfigConfig: CORSConfigConfig{
				AllowedOrigins: []string{"https://example.com"},
			},
		},
	}

	// No headers in desired, Vault may return some defaults
	payload := map[string]any{
		"allowed_origins": []any{"https://example.com"},
		"allowed_headers": []any{"Content-Type", "Authorization"},
	}
	if !cc.IsEquivalentToDesiredState(payload) {
		t.Error("expected empty desired headers to be equivalent regardless of Vault headers")
	}
}
