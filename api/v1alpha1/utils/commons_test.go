package utils

import (
	"testing"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil returns empty", nil, ""},
		{"string passes through", "hello", "hello"},
		{"empty string passes through", "", ""},
		{"[]byte converts to string", []byte("PEM-DATA"), "PEM-DATA"},
		{"int formats via Sprintf", 42, "42"},
		{"bool formats via Sprintf", true, "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToString(tt.input)
			if got != tt.expected {
				t.Errorf("ToString(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
