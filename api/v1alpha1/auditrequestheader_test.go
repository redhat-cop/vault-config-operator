package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAuditRequestHeaderGetPath(t *testing.T) {
	tests := []struct {
		name         string
		header       *AuditRequestHeader
		expectedPath string
	}{
		{
			name: "uses spec.name for path",
			header: &AuditRequestHeader{
				Spec: AuditRequestHeaderSpec{
					Name: "X-Custom-Header",
				},
			},
			expectedPath: "sys/config/auditing/request-headers/X-Custom-Header",
		},
		{
			name: "uses spec.name even when metadata.name differs",
			header: &AuditRequestHeader{
				ObjectMeta: metav1.ObjectMeta{
					Name: "metadata-name-should-be-ignored",
				},
				Spec: AuditRequestHeaderSpec{
					Name: "X-Vault-Token",
				},
			},
			expectedPath: "sys/config/auditing/request-headers/X-Vault-Token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.header.GetPath(); result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestAuditRequestHeaderGetPayload(t *testing.T) {
	tests := []struct {
		name     string
		hmac     bool
		expected map[string]interface{}
	}{
		{
			name:     "hmac true",
			hmac:     true,
			expected: map[string]interface{}{"hmac": true},
		},
		{
			name:     "hmac false",
			hmac:     false,
			expected: map[string]interface{}{"hmac": false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &AuditRequestHeader{
				Spec: AuditRequestHeaderSpec{
					HMAC: tt.hmac,
				},
			}
			payload := header.GetPayload()
			if payload["hmac"] != tt.expected["hmac"] {
				t.Errorf("expected hmac=%v, got %v", tt.expected["hmac"], payload["hmac"])
			}
			if len(payload) != 1 {
				t.Errorf("expected exactly 1 key in payload, got %d", len(payload))
			}
		})
	}
}

func TestAuditRequestHeaderIsEquivalentToDesiredState(t *testing.T) {
	tests := []struct {
		name       string
		specHMAC   bool
		payload    map[string]interface{}
		equivalent bool
	}{
		{
			name:       "matching hmac=true",
			specHMAC:   true,
			payload:    map[string]interface{}{"hmac": true},
			equivalent: true,
		},
		{
			name:       "matching hmac=false",
			specHMAC:   false,
			payload:    map[string]interface{}{"hmac": false},
			equivalent: true,
		},
		{
			name:       "non-matching hmac values",
			specHMAC:   true,
			payload:    map[string]interface{}{"hmac": false},
			equivalent: false,
		},
		{
			name:       "non-matching hmac values reversed",
			specHMAC:   false,
			payload:    map[string]interface{}{"hmac": true},
			equivalent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := &AuditRequestHeader{
				Spec: AuditRequestHeaderSpec{
					HMAC: tt.specHMAC,
				},
			}
			result := header.IsEquivalentToDesiredState(tt.payload)
			if result != tt.equivalent {
				t.Errorf("IsEquivalentToDesiredState() = %v, expected %v", result, tt.equivalent)
			}
		})
	}
}

func TestAuditRequestHeaderIsEquivalentMissingHMACKey(t *testing.T) {
	header := &AuditRequestHeader{
		Spec: AuditRequestHeaderSpec{
			HMAC: true,
		},
	}

	payload := map[string]interface{}{}
	if header.IsEquivalentToDesiredState(payload) {
		t.Error("expected missing hmac key to return false")
	}
}

func TestAuditRequestHeaderIsEquivalentNonBoolHMAC(t *testing.T) {
	header := &AuditRequestHeader{
		Spec: AuditRequestHeaderSpec{
			HMAC: true,
		},
	}

	tests := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name:    "string value",
			payload: map[string]interface{}{"hmac": "true"},
		},
		{
			name:    "integer value",
			payload: map[string]interface{}{"hmac": 1},
		},
		{
			name:    "nil value",
			payload: map[string]interface{}{"hmac": nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if header.IsEquivalentToDesiredState(tt.payload) {
				t.Errorf("expected non-bool hmac value to return false, payload: %v", tt.payload)
			}
		})
	}
}

// AuditRequestHeader uses a field-specific check (payload["hmac"].(bool)),
// not reflect.DeepEqual, so extra fields are inherently ignored.
func TestAuditRequestHeaderIsEquivalentExtraFieldsIgnored(t *testing.T) {
	header := &AuditRequestHeader{
		Spec: AuditRequestHeaderSpec{
			HMAC: true,
		},
	}

	payloadWithExtra := map[string]interface{}{
		"hmac":        true,
		"extra_field": "vault-returned-value",
		"another":     42,
	}
	if !header.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected AuditRequestHeader to ignore extra fields (uses field-specific check, not reflect.DeepEqual)")
	}
}

func TestAuditRequestHeaderIsDeletable(t *testing.T) {
	header := &AuditRequestHeader{}
	if !header.IsDeletable() {
		t.Error("expected AuditRequestHeader to be deletable")
	}
}

func TestAuditRequestHeaderConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	header := &AuditRequestHeader{}
	header.SetConditions([]metav1.Condition{condition})
	if len(header.GetConditions()) != 1 || header.GetConditions()[0].Type != "Ready" {
		t.Error("expected AuditRequestHeader conditions to be set and retrieved")
	}
}
