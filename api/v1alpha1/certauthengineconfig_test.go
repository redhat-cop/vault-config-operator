package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCertAuthEngineConfigGetPath(t *testing.T) {
	tests := []struct {
		name         string
		config       *CertAuthEngineConfig
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			config: &CertAuthEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: CertAuthEngineConfigSpec{
					Path: "cert",
					Name: "custom-name",
				},
			},
			expectedPath: "auth/cert/custom-name/config",
		},
		{
			name: "without spec.name falls back to metadata.name",
			config: &CertAuthEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: CertAuthEngineConfigSpec{
					Path: "cert",
				},
			},
			expectedPath: "auth/cert/meta-name/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestCertAuthEngineConfigInternalToMap(t *testing.T) {
	config := CertAuthEngineConfigInternal{
		DisableBinding:              true,
		EnableIdentityAliasMetadata: true,
		OCSPCacheSize:               200,
		RoleCacheSize:               500,
	}

	result := config.toMap()

	if len(result) != 4 {
		t.Errorf("expected 4 keys in map, got %d", len(result))
	}

	expected := map[string]any{
		"disable_binding":                true,
		"enable_identity_alias_metadata": true,
		"ocsp_cache_size":                200,
		"role_cache_size":                500,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestCertAuthEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &CertAuthEngineConfig{
		Spec: CertAuthEngineConfigSpec{
			CertAuthEngineConfigInternal: CertAuthEngineConfigInternal{
				DisableBinding: false,
				OCSPCacheSize:  100,
				RoleCacheSize:  200,
			},
		},
	}

	payload := config.Spec.CertAuthEngineConfigInternal.toMap()

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestCertAuthEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &CertAuthEngineConfig{
		Spec: CertAuthEngineConfigSpec{
			CertAuthEngineConfigInternal: CertAuthEngineConfigInternal{
				OCSPCacheSize: 100,
			},
		},
	}

	payload := config.Spec.CertAuthEngineConfigInternal.toMap()
	payload["ocsp_cache_size"] = 999

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different ocsp_cache_size) to NOT be equivalent")
	}
}

func TestCertAuthEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &CertAuthEngineConfig{
		Spec: CertAuthEngineConfigSpec{
			CertAuthEngineConfigInternal: CertAuthEngineConfigInternal{
				OCSPCacheSize: 100,
			},
		},
	}

	payload := config.Spec.CertAuthEngineConfigInternal.toMap()
	payload["extra_field"] = "unexpected"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
	}
}

func TestCertAuthEngineConfigIsDeletable(t *testing.T) {
	config := &CertAuthEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected CertAuthEngineConfig to be deletable")
	}
}

func TestCertAuthEngineConfigConditions(t *testing.T) {
	config := &CertAuthEngineConfig{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	config.SetConditions(conditions)
	got := config.GetConditions()

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
