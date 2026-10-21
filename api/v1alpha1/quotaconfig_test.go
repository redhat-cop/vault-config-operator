package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestQuotaConfigGetPath(t *testing.T) {
	qc := &QuotaConfig{}
	expected := "sys/quotas/config"
	if qc.GetPath() != expected {
		t.Errorf("GetPath() = %v, expected %v", qc.GetPath(), expected)
	}
}

func TestQuotaConfigIsDeletable(t *testing.T) {
	qc := &QuotaConfig{}
	if qc.IsDeletable() {
		t.Error("expected QuotaConfig to not be deletable")
	}
}

func TestQuotaConfigToMap(t *testing.T) {
	tests := []struct {
		name     string
		data     QuotaConfigData
		expected map[string]any
	}{
		{
			name: "all fields set",
			data: QuotaConfigData{
				EnableRateLimitAuditLogging:    true,
				EnableRateLimitResponseHeaders: true,
				RateLimitExemptPaths:           []string{"sys/health", "sys/init"},
				AbsoluteRateLimitExemptPaths:   []string{"/v1/sys/health"},
			},
			expected: map[string]any{
				"enable_rate_limit_audit_logging":    true,
				"enable_rate_limit_response_headers": true,
				"rate_limit_exempt_paths":            []any{"sys/health", "sys/init"},
				"absolute_rate_limit_exempt_paths":   []any{"/v1/sys/health"},
			},
		},
		{
			name: "empty fields",
			data: QuotaConfigData{},
			expected: map[string]any{
				"enable_rate_limit_audit_logging":    false,
				"enable_rate_limit_response_headers": false,
				"rate_limit_exempt_paths":            []any{},
				"absolute_rate_limit_exempt_paths":   []any{},
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
		})
	}
}

func TestQuotaConfigToMapPreservesSliceOrder(t *testing.T) {
	data := QuotaConfigData{
		RateLimitExemptPaths: []string{"z-path", "a-path", "m-path"},
	}
	result := data.toMap()
	paths := result["rate_limit_exempt_paths"].([]any)
	if paths[0].(string) != "z-path" || paths[1].(string) != "a-path" || paths[2].(string) != "m-path" {
		t.Errorf("expected original order preserved, got %v", paths)
	}
}

func TestQuotaConfigIsEquivalentSortsBeforeCompare(t *testing.T) {
	qc := &QuotaConfig{
		Spec: QuotaConfigSpec{
			QuotaConfigData: QuotaConfigData{
				RateLimitExemptPaths: []string{"z-path", "a-path", "m-path"},
			},
		},
	}
	payload := map[string]any{
		"rate_limit_exempt_paths": []any{"a-path", "m-path", "z-path"},
	}
	if !qc.IsEquivalentToDesiredState(payload) {
		t.Error("expected equivalent after sorting")
	}
}

func TestQuotaConfigIsEquivalentToDesiredState(t *testing.T) {
	qc := &QuotaConfig{
		Spec: QuotaConfigSpec{
			QuotaConfigData: QuotaConfigData{
				EnableRateLimitAuditLogging: true,
			},
		},
	}

	matching := map[string]any{
		"enable_rate_limit_audit_logging": true,
	}
	if !qc.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]any{
		"enable_rate_limit_audit_logging": false,
	}
	if qc.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestQuotaConfigIsEquivalentExtraFields(t *testing.T) {
	qc := &QuotaConfig{
		Spec: QuotaConfigSpec{
			QuotaConfigData: QuotaConfigData{
				EnableRateLimitAuditLogging: true,
			},
		},
	}

	payloadWithExtra := map[string]any{
		"enable_rate_limit_audit_logging": true,
		"request_id":                      "vault-returned-id",
	}
	if !qc.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestQuotaConfigConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	qc := &QuotaConfig{}
	qc.SetConditions([]metav1.Condition{condition})
	if len(qc.GetConditions()) != 1 || qc.GetConditions()[0].Type != "Ready" {
		t.Error("expected QuotaConfig conditions to be set and retrieved")
	}
}
