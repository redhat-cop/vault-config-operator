package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLoggerConfigGetPath(t *testing.T) {
	tests := []struct {
		name         string
		logger       *LoggerConfig
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			logger: &LoggerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-logger-metadata",
				},
				Spec: LoggerConfigSpec{
					Name: "core",
				},
			},
			expectedPath: "sys/loggers/core",
		},
		{
			name: "without spec.name uses metadata.name",
			logger: &LoggerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "audit",
				},
				Spec: LoggerConfigSpec{},
			},
			expectedPath: "sys/loggers/audit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.logger.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestLoggerConfigGetPayload(t *testing.T) {
	logger := &LoggerConfig{
		Spec: LoggerConfigSpec{
			LoggerConfigConfig: LoggerConfigConfig{
				Level: "debug",
			},
		},
	}

	payload := logger.GetPayload()

	level, ok := payload["level"]
	if !ok {
		t.Fatal("expected payload to contain 'level' key")
	}
	if level != "debug" {
		t.Errorf("expected level 'debug', got %v", level)
	}
	if len(payload) != 1 {
		t.Errorf("expected exactly 1 key in payload, got %d", len(payload))
	}
}

func TestLoggerConfigIsEquivalentToDesiredState(t *testing.T) {
	logger := &LoggerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "core",
		},
		Spec: LoggerConfigSpec{
			LoggerConfigConfig: LoggerConfigConfig{
				Level: "debug",
			},
		},
	}

	// Simulated Vault read response — key is logger name, value is level
	vaultPayload := map[string]any{
		"core": "debug",
	}
	if !logger.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected Vault read response {core: debug} to match desired state")
	}
}

func TestLoggerConfigIsEquivalentNegative(t *testing.T) {
	logger := &LoggerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "core",
		},
		Spec: LoggerConfigSpec{
			LoggerConfigConfig: LoggerConfigConfig{
				Level: "debug",
			},
		},
	}

	// Mismatched level
	mismatchPayload := map[string]any{
		"core": "info",
	}
	if logger.IsEquivalentToDesiredState(mismatchPayload) {
		t.Error("expected mismatched level to return false")
	}
}

func TestLoggerConfigIsEquivalentExtraFields(t *testing.T) {
	logger := &LoggerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "core",
		},
		Spec: LoggerConfigSpec{
			LoggerConfigConfig: LoggerConfigConfig{
				Level: "debug",
			},
		},
	}

	// Extra Vault-returned keys should not affect the comparison
	payloadWithExtra := map[string]any{
		"core":  "debug",
		"extra": "some-value",
	}
	if !logger.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected extra fields to be ignored in logger name-keyed comparison")
	}
}

func TestLoggerConfigIsEquivalentWithSpecName(t *testing.T) {
	logger := &LoggerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-logger-metadata",
		},
		Spec: LoggerConfigSpec{
			Name: "audit",
			LoggerConfigConfig: LoggerConfigConfig{
				Level: "trace",
			},
		},
	}

	// Vault response uses spec.name as key
	vaultPayload := map[string]any{
		"audit": "trace",
	}
	if !logger.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected spec.name 'audit' to be used as payload key")
	}

	// Vault response uses metadata.name — should NOT match since spec.name takes precedence
	wrongKeyPayload := map[string]any{
		"test-logger-metadata": "trace",
	}
	if logger.IsEquivalentToDesiredState(wrongKeyPayload) {
		t.Error("expected metadata name key to not match when spec.name is set")
	}
}

func TestLoggerConfigIsDeletable(t *testing.T) {
	logger := &LoggerConfig{}
	if !logger.IsDeletable() {
		t.Error("expected LoggerConfig to be deletable")
	}
}

func TestLoggerConfigConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	logger := &LoggerConfig{}
	logger.SetConditions([]metav1.Condition{condition})
	if len(logger.GetConditions()) != 1 || logger.GetConditions()[0].Type != "Ready" {
		t.Error("expected LoggerConfig conditions to be set and retrieved")
	}
}
