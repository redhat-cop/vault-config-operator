package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTransitSecretEngineKeyGetPath(t *testing.T) {
	tests := []struct {
		name         string
		key          *TransitSecretEngineKey
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			key: &TransitSecretEngineKey{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: TransitSecretEngineKeySpec{
					Path: "transit",
					Name: "custom-key",
				},
			},
			expectedPath: vaultutils.CleansePath("transit/keys/custom-key"),
		},
		{
			name: "without spec.name falls back to metadata.name",
			key: &TransitSecretEngineKey{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: TransitSecretEngineKeySpec{
					Path: "transit",
				},
			},
			expectedPath: vaultutils.CleansePath("transit/keys/meta-name"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.key.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestTransitSecretEngineKeyGetConfigPath(t *testing.T) {
	key := &TransitSecretEngineKey{
		ObjectMeta: metav1.ObjectMeta{Name: "my-key"},
		Spec: TransitSecretEngineKeySpec{
			Path: "transit",
		},
	}
	expected := vaultutils.CleansePath("transit/keys/my-key") + "/config"
	result := key.GetConfigPath()
	if result != expected {
		t.Errorf("GetConfigPath() = %v, expected %v", result, expected)
	}
}

func TestTransitKeyConfigToMap(t *testing.T) {
	config := TransitKeyConfig{
		Type:                 "aes256-gcm96",
		Derived:              true,
		ConvergentEncryption: true,
		KeySize:              0,
		Exportable:           true,
		AllowPlaintextBackup: false,
		AutoRotatePeriod:     "24h",
	}

	result := config.toMap()

	expectedKeys := []string{"type", "derived", "convergent_encryption", "exportable", "allow_plaintext_backup", "key_size", "auto_rotate_period"}
	if len(result) != len(expectedKeys) {
		t.Errorf("toMap() len = %d, want %d keys", len(result), len(expectedKeys))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if result["type"] != "aes256-gcm96" {
		t.Errorf("type = %v, want aes256-gcm96", result["type"])
	}
	if result["derived"] != true {
		t.Errorf("derived = %v, want true", result["derived"])
	}
	if result["convergent_encryption"] != true {
		t.Errorf("convergent_encryption = %v, want true", result["convergent_encryption"])
	}
	if result["exportable"] != true {
		t.Errorf("exportable = %v, want true", result["exportable"])
	}
	if result["allow_plaintext_backup"] != false {
		t.Errorf("allow_plaintext_backup = %v, want false", result["allow_plaintext_backup"])
	}
	if result["key_size"] != 0 {
		t.Errorf("key_size = %v, want 0", result["key_size"])
	}
	if result["auto_rotate_period"] != "24h" {
		t.Errorf("auto_rotate_period = %v, want 24h", result["auto_rotate_period"])
	}
}

func TestTransitKeyConfigConfigToMap(t *testing.T) {
	config := TransitKeyConfig{
		Type:                 "aes256-gcm96",
		Derived:              true,
		MinDecryptionVersion: 2,
		MinEncryptionVersion: 1,
		DeletionAllowed:      true,
		Exportable:           true,
		AllowPlaintextBackup: false,
		AutoRotatePeriod:     "48h",
	}

	result := config.configToMap()

	expectedKeys := []string{"min_decryption_version", "min_encryption_version", "deletion_allowed", "exportable", "allow_plaintext_backup", "auto_rotate_period"}
	if len(result) != len(expectedKeys) {
		t.Errorf("configToMap() len = %d, want %d keys", len(result), len(expectedKeys))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in configToMap() output", key)
		}
	}

	if _, ok := result["type"]; ok {
		t.Error("configToMap() should not include create-time field 'type'")
	}
	if _, ok := result["derived"]; ok {
		t.Error("configToMap() should not include create-time field 'derived'")
	}

	if result["min_decryption_version"] != float64(2) {
		t.Errorf("min_decryption_version = %v (%T), want float64(2)", result["min_decryption_version"], result["min_decryption_version"])
	}
	if result["min_encryption_version"] != float64(1) {
		t.Errorf("min_encryption_version = %v (%T), want float64(1)", result["min_encryption_version"], result["min_encryption_version"])
	}
	if result["deletion_allowed"] != true {
		t.Errorf("deletion_allowed = %v, want true", result["deletion_allowed"])
	}
	if result["exportable"] != true {
		t.Errorf("exportable = %v, want true", result["exportable"])
	}
	if result["auto_rotate_period"] != "48h" {
		t.Errorf("auto_rotate_period = %v, want 48h", result["auto_rotate_period"])
	}
}

func TestTransitKeyConfigAutoRotatePeriodZeroValue(t *testing.T) {
	config := TransitKeyConfig{
		Type: "aes256-gcm96",
	}

	createPayload := config.toMap()
	if createPayload["auto_rotate_period"] != "0" {
		t.Errorf("toMap() auto_rotate_period = %q, want \"0\" for unset field", createPayload["auto_rotate_period"])
	}

	configPayload := config.configToMap()
	if configPayload["auto_rotate_period"] != "0" {
		t.Errorf("configToMap() auto_rotate_period = %q, want \"0\" for unset field", configPayload["auto_rotate_period"])
	}

	configWithValue := TransitKeyConfig{
		Type:             "aes256-gcm96",
		AutoRotatePeriod: "72h",
	}
	if result := configWithValue.toMap(); result["auto_rotate_period"] != "72h" {
		t.Errorf("toMap() auto_rotate_period = %q, want \"72h\" for explicitly set field", result["auto_rotate_period"])
	}
	if result := configWithValue.configToMap(); result["auto_rotate_period"] != "72h" {
		t.Errorf("configToMap() auto_rotate_period = %q, want \"72h\" for explicitly set field", result["auto_rotate_period"])
	}
}

func TestTransitSecretEngineKeyIsEquivalentMatching(t *testing.T) {
	key := &TransitSecretEngineKey{
		Spec: TransitSecretEngineKeySpec{
			Path: "transit",
			TransitKeyConfig: TransitKeyConfig{
				Type:                 "aes256-gcm96",
				MinDecryptionVersion: 1,
				MinEncryptionVersion: 0,
				DeletionAllowed:      false,
				Exportable:           false,
				AllowPlaintextBackup: false,
				AutoRotatePeriod:     "",
			},
		},
	}

	// Vault returns "0" for disabled auto-rotation and numeric values as float64
	vaultPayload := map[string]any{
		"min_decryption_version": float64(1),
		"min_encryption_version": float64(0),
		"deletion_allowed":       false,
		"exportable":             false,
		"allow_plaintext_backup": false,
		"auto_rotate_period":     "0",
	}

	if !key.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestTransitSecretEngineKeyIsEquivalentExtraFields(t *testing.T) {
	key := &TransitSecretEngineKey{
		Spec: TransitSecretEngineKeySpec{
			Path: "transit",
			TransitKeyConfig: TransitKeyConfig{
				MinDecryptionVersion: 1,
				MinEncryptionVersion: 0,
				DeletionAllowed:      false,
				Exportable:           false,
				AllowPlaintextBackup: false,
				AutoRotatePeriod:     "",
			},
		},
	}

	// Vault returns "0" for disabled auto-rotation and numeric values as float64
	vaultPayload := map[string]any{
		"type":                   "aes256-gcm96",
		"deletion_allowed":       false,
		"derived":                false,
		"exportable":             false,
		"allow_plaintext_backup": false,
		"keys":                   map[string]any{"1": float64(1442851412)},
		"min_decryption_version": float64(1),
		"min_encryption_version": float64(0),
		"name":                   "foo",
		"supports_encryption":    true,
		"supports_decryption":    true,
		"supports_derivation":    true,
		"supports_signing":       false,
		"imported":               false,
		"auto_rotate_period":     "0",
	}

	if !key.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected extra Vault metadata fields (keys, name, supports_*, imported, type, derived) to be filtered out")
	}
}

func TestTransitSecretEngineKeyIsEquivalentMismatch(t *testing.T) {
	key := &TransitSecretEngineKey{
		Spec: TransitSecretEngineKeySpec{
			Path: "transit",
			TransitKeyConfig: TransitKeyConfig{
				MinDecryptionVersion: 2,
				MinEncryptionVersion: 0,
				DeletionAllowed:      false,
				Exportable:           false,
				AllowPlaintextBackup: false,
				AutoRotatePeriod:     "",
			},
		},
	}

	// Vault returns "0" for disabled auto-rotation and numeric values as float64
	vaultPayload := map[string]any{
		"min_decryption_version": float64(1),
		"min_encryption_version": float64(0),
		"deletion_allowed":       false,
		"exportable":             false,
		"allow_plaintext_backup": false,
		"auto_rotate_period":     "0",
	}

	if key.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected mismatched min_decryption_version to return false")
	}
}

func TestTransitSecretEngineKeyIsEquivalentFloat64Match(t *testing.T) {
	key := &TransitSecretEngineKey{
		Spec: TransitSecretEngineKeySpec{
			Path: "transit",
			TransitKeyConfig: TransitKeyConfig{
				MinDecryptionVersion: 5,
				MinEncryptionVersion: 3,
				DeletionAllowed:      true,
				Exportable:           true,
				AllowPlaintextBackup: true,
				AutoRotatePeriod:     "24h",
			},
		},
	}

	// Simulate Vault's actual JSON response (numbers as float64)
	vaultPayload := map[string]any{
		"min_decryption_version": float64(5),
		"min_encryption_version": float64(3),
		"deletion_allowed":       true,
		"exportable":             true,
		"allow_plaintext_backup": true,
		"auto_rotate_period":     "24h",
	}

	if !key.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected float64 numeric values from Vault to match without false drift detection")
	}
}

func TestTransitSecretEngineKeyIsDeletable(t *testing.T) {
	key := &TransitSecretEngineKey{}
	if !key.IsDeletable() {
		t.Error("expected TransitSecretEngineKey to be deletable")
	}
}

func TestTransitSecretEngineKeyConditions(t *testing.T) {
	key := &TransitSecretEngineKey{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	key.SetConditions(conditions)
	got := key.GetConditions()

	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected condition type 'ReconcileSuccessful', got %v", got[0].Type)
	}
	if got[0].Status != metav1.ConditionTrue {
		t.Errorf("expected condition status True, got %v", got[0].Status)
	}
}
