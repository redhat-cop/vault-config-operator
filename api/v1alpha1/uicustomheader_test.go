package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUICustomHeaderGetPath(t *testing.T) {
	tests := []struct {
		name         string
		header       *UICustomHeader
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			header: &UICustomHeader{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-header-metadata",
				},
				Spec: UICustomHeaderSpec{
					Name: "X-Custom-Header",
				},
			},
			expectedPath: "sys/config/ui/headers/X-Custom-Header",
		},
		{
			name: "without spec.name uses metadata.name",
			header: &UICustomHeader{
				ObjectMeta: metav1.ObjectMeta{
					Name: "x-default-header",
				},
				Spec: UICustomHeaderSpec{},
			},
			expectedPath: "sys/config/ui/headers/x-default-header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.header.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestUICustomHeaderGetPayload(t *testing.T) {
	header := &UICustomHeader{
		Spec: UICustomHeaderSpec{
			UICustomHeaderConfig: UICustomHeaderConfig{
				Values: []string{"val1", "val2"},
			},
		},
	}

	payload := header.GetPayload()

	values, ok := payload["values"]
	if !ok {
		t.Fatal("expected payload to contain 'values' key")
	}
	valuesSlice, ok := values.([]any)
	if !ok {
		t.Fatal("expected 'values' to be []any")
	}
	if len(valuesSlice) != 2 {
		t.Errorf("expected 2 values, got %d", len(valuesSlice))
	}
	if valuesSlice[0] != "val1" || valuesSlice[1] != "val2" {
		t.Errorf("expected values [val1, val2], got %v", valuesSlice)
	}
	if len(payload) != 1 {
		t.Errorf("expected exactly 1 key in payload, got %d", len(payload))
	}
}

func TestUICustomHeaderIsEquivalentToDesiredState(t *testing.T) {
	header := &UICustomHeader{
		Spec: UICustomHeaderSpec{
			UICustomHeaderConfig: UICustomHeaderConfig{
				Values: []string{"my-header-value"},
			},
		},
	}

	// Simulated Vault single-value read response
	singleValuePayload := map[string]any{
		"value": "my-header-value",
	}
	if !header.IsEquivalentToDesiredState(singleValuePayload) {
		t.Error("expected single-value Vault response to match single desired value")
	}
}

func TestUICustomHeaderIsEquivalentMultiValue(t *testing.T) {
	header := &UICustomHeader{
		Spec: UICustomHeaderSpec{
			UICustomHeaderConfig: UICustomHeaderConfig{
				Values: []string{"val1", "val2"},
			},
		},
	}

	// Simulated Vault multi-value read response
	multiValuePayload := map[string]any{
		"values": []any{"val1", "val2"},
	}
	if !header.IsEquivalentToDesiredState(multiValuePayload) {
		t.Error("expected multi-value Vault response to match desired values")
	}
}

func TestUICustomHeaderIsEquivalentNegative(t *testing.T) {
	header := &UICustomHeader{
		Spec: UICustomHeaderSpec{
			UICustomHeaderConfig: UICustomHeaderConfig{
				Values: []string{"expected-value"},
			},
		},
	}

	// Mismatched single-value
	mismatchPayload := map[string]any{
		"value": "different-value",
	}
	if header.IsEquivalentToDesiredState(mismatchPayload) {
		t.Error("expected mismatched single value to return false")
	}

	// Multi-value desired but single-value returned — should always re-write
	multiDesiredHeader := &UICustomHeader{
		Spec: UICustomHeaderSpec{
			UICustomHeaderConfig: UICustomHeaderConfig{
				Values: []string{"val1", "val2"},
			},
		},
	}
	singlePayload := map[string]any{
		"value": "val1",
	}
	if multiDesiredHeader.IsEquivalentToDesiredState(singlePayload) {
		t.Error("expected multi-value desired with single-value read to return false")
	}
}

func TestUICustomHeaderIsEquivalentExtraFields(t *testing.T) {
	header := &UICustomHeader{
		Spec: UICustomHeaderSpec{
			UICustomHeaderConfig: UICustomHeaderConfig{
				Values: []string{"my-value"},
			},
		},
	}

	payloadWithExtra := map[string]any{
		"value":       "my-value",
		"extra_field": "vault-returned-value",
	}
	if !header.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected extra fields to be ignored in single-value comparison")
	}
}

func TestUICustomHeaderIsDeletable(t *testing.T) {
	header := &UICustomHeader{}
	if !header.IsDeletable() {
		t.Error("expected UICustomHeader to be deletable")
	}
}

func TestUICustomHeaderConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	header := &UICustomHeader{}
	header.SetConditions([]metav1.Condition{condition})
	if len(header.GetConditions()) != 1 || header.GetConditions()[0].Type != "Ready" {
		t.Error("expected UICustomHeader conditions to be set and retrieved")
	}
}
