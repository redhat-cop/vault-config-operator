package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecretEngineMountGetPath(t *testing.T) {
	tests := []struct {
		name         string
		mount        *SecretEngineMount
		expectedPath string
	}{
		{
			name: "with path and name specified",
			mount: &SecretEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: SecretEngineMountSpec{
					Path: "custom-path",
					Name: "custom-name",
				},
			},
			expectedPath: "sys/mounts/custom-path/custom-name",
		},
		{
			name: "with path but no name specified",
			mount: &SecretEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: SecretEngineMountSpec{
					Path: "custom-path",
				},
			},
			expectedPath: "sys/mounts/custom-path/test-mount",
		},
		{
			name: "with name but no path specified",
			mount: &SecretEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: SecretEngineMountSpec{
					Name: "custom-name",
				},
			},
			expectedPath: "sys/mounts/custom-name",
		},
		{
			name: "with neither path nor name specified",
			mount: &SecretEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: SecretEngineMountSpec{},
			},
			expectedPath: "sys/mounts/test-mount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mount.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestMountConfigToMap(t *testing.T) {
	config := MountConfig{
		DefaultLeaseTTL:           "1h",
		MaxLeaseTTL:               "24h",
		ForceNoCache:              true,
		AuditNonHMACRequestKeys:   []string{"key1", "key2"},
		AuditNonHMACResponseKeys:  []string{"resp1"},
		ListingVisibility:         "unauth",
		PassthroughRequestHeaders: []string{"X-Custom"},
		AllowedResponseHeaders:    []string{"X-Response"},
	}

	result := config.toMap()

	expected := map[string]interface{}{
		"default_lease_ttl":            "1h",
		"max_lease_ttl":                "24h",
		"force_no_cache":               true,
		"audit_non_hmac_request_keys":  []string{"key1", "key2"},
		"audit_non_hmac_response_keys": []string{"resp1"},
		"listing_visibility":           "unauth",
		"passthrough_request_headers":  []string{"X-Custom"},
		"allowed_response_headers":     []string{"X-Response"},
	}

	if len(result) != 8 {
		t.Errorf("expected 8 keys in config map, got %d", len(result))
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestMountToMap(t *testing.T) {
	config := MountConfig{
		DefaultLeaseTTL:   "1h",
		MaxLeaseTTL:       "24h",
		ListingVisibility: "hidden",
	}

	mount := Mount{
		Type:                  "kv",
		Description:           "KV secrets",
		Config:                config,
		Local:                 true,
		SealWrap:              false,
		ExternalEntropyAccess: true,
		Options:               map[string]string{"version": "2"},
	}

	result := mount.toMap()

	if len(result) != 7 {
		t.Errorf("expected 7 keys in mount map, got %d", len(result))
	}

	if result["type"] != "kv" {
		t.Errorf("expected type 'kv', got %v", result["type"])
	}
	if result["description"] != "KV secrets" {
		t.Errorf("expected description 'KV secrets', got %v", result["description"])
	}
	if result["local"] != true {
		t.Errorf("expected local true, got %v", result["local"])
	}
	if result["seal_wrap"] != false {
		t.Errorf("expected seal_wrap false, got %v", result["seal_wrap"])
	}
	if result["external_entropy_access"] != true {
		t.Errorf("expected external_entropy_access true, got %v", result["external_entropy_access"])
	}
	if !reflect.DeepEqual(result["options"], map[string]string{"version": "2"}) {
		t.Errorf("expected options map, got %v", result["options"])
	}

	configMap, ok := result["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected config to be map[string]interface{}")
	}
	if configMap["default_lease_ttl"] != "1h" {
		t.Errorf("expected nested config default_lease_ttl '1h', got %v", configMap["default_lease_ttl"])
	}
}

func TestSecretEngineMountGetPayload(t *testing.T) {
	mount := &SecretEngineMount{
		Spec: SecretEngineMountSpec{
			Mount: Mount{
				Type:        "kv",
				Description: "KV store",
				Config: MountConfig{
					DefaultLeaseTTL:   "30m",
					MaxLeaseTTL:       "1h",
					ListingVisibility: "hidden",
				},
				Local:                 false,
				SealWrap:              true,
				ExternalEntropyAccess: false,
				Options:               map[string]string{"version": "2"},
			},
		},
	}

	payload := mount.GetPayload()

	// GetPayload returns the full mount spec via Mount.toMap()
	if payload["type"] != "kv" {
		t.Errorf("expected type 'kv', got %v", payload["type"])
	}
	if payload["description"] != "KV store" {
		t.Errorf("expected description 'KV store', got %v", payload["description"])
	}
	if payload["seal_wrap"] != true {
		t.Errorf("expected seal_wrap true, got %v", payload["seal_wrap"])
	}
	if !reflect.DeepEqual(payload["options"], map[string]string{"version": "2"}) {
		t.Errorf("expected options in payload, got %v", payload["options"])
	}

	configMap, ok := payload["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected payload config to be map[string]interface{}")
	}
	if configMap["default_lease_ttl"] != "30m" {
		t.Errorf("expected nested default_lease_ttl '30m', got %v", configMap["default_lease_ttl"])
	}
}

func TestSecretEngineMountGetTunePayload(t *testing.T) {
	mount := &SecretEngineMount{
		Spec: SecretEngineMountSpec{
			Mount: Mount{
				Type:        "kv",
				Description: "KV store",
				Config: MountConfig{
					DefaultLeaseTTL:   "30m",
					MaxLeaseTTL:       "1h",
					ForceNoCache:      true,
					ListingVisibility: "hidden",
				},
				Local:    false,
				SealWrap: true,
			},
		},
	}

	tunePayload := mount.GetTunePayload()

	// GetTunePayload returns Config.toMap() (no delete of options/description)
	if _, ok := tunePayload["type"]; ok {
		t.Error("expected tune payload to NOT contain 'type' (mount-level field)")
	}
	if _, ok := tunePayload["seal_wrap"]; ok {
		t.Error("expected tune payload to NOT contain 'seal_wrap' (mount-level field)")
	}

	if tunePayload["default_lease_ttl"] != "30m" {
		t.Errorf("expected default_lease_ttl '30m', got %v", tunePayload["default_lease_ttl"])
	}
	if tunePayload["force_no_cache"] != true {
		t.Errorf("expected force_no_cache true, got %v", tunePayload["force_no_cache"])
	}

	// GetTunePayload returns Config.toMap() without the delete calls that
	// IsEquivalentToDesiredState performs. Since MountConfig.toMap() doesn't
	// include "options" or "description", the result should be identical.
	configMap := mount.Spec.Config.toMap()
	if !reflect.DeepEqual(tunePayload, configMap) {
		t.Errorf("GetTunePayload() should equal Config.toMap()")
	}
}

func TestSecretEngineMountIsEquivalentToDesiredStateMatching(t *testing.T) {
	mount := &SecretEngineMount{
		Spec: SecretEngineMountSpec{
			Mount: Mount{
				Config: MountConfig{
					DefaultLeaseTTL:           "1h",
					MaxLeaseTTL:               "24h",
					ForceNoCache:              false,
					AuditNonHMACRequestKeys:   []string{"key1"},
					AuditNonHMACResponseKeys:  []string{"resp1"},
					ListingVisibility:         "hidden",
					PassthroughRequestHeaders: []string{"X-Custom"},
					AllowedResponseHeaders:    []string{"X-Resp"},
				},
			},
		},
	}

	// IsEquivalentToDesiredState deletes "options" and "description" from the
	// config map before comparison. Since MountConfig.toMap() doesn't include
	// those keys, the deletes are no-ops — the payload matches after deletion.
	configMap := mount.Spec.Config.toMap()
	delete(configMap, "options")
	delete(configMap, "description")

	if !mount.IsEquivalentToDesiredState(configMap) {
		t.Error("expected matching tune config payload to be equivalent")
	}
}

// IsEquivalentToDesiredState compares only Config.toMap() (tune config), NOT
// the full mount spec from GetPayload(). This test proves AC #3: passing the
// full mount spec (which includes type, local, seal_wrap, etc.) must return
// false.
func TestSecretEngineMountIsEquivalentRejectsFullMountPayload(t *testing.T) {
	mount := &SecretEngineMount{
		Spec: SecretEngineMountSpec{
			Mount: Mount{
				Type:        "kv",
				Description: "KV store",
				Config: MountConfig{
					DefaultLeaseTTL:   "1h",
					MaxLeaseTTL:       "24h",
					ListingVisibility: "hidden",
				},
				Local:                 false,
				SealWrap:              true,
				ExternalEntropyAccess: false,
				Options:               map[string]string{"version": "2"},
			},
		},
	}

	fullPayload := mount.GetPayload()

	if mount.IsEquivalentToDesiredState(fullPayload) {
		t.Error("expected full mount spec payload (from GetPayload) to NOT be equivalent — comparison should be tune-only")
	}
}

func TestSecretEngineMountIsEquivalentToDesiredStateNonMatching(t *testing.T) {
	mount := &SecretEngineMount{
		Spec: SecretEngineMountSpec{
			Mount: Mount{
				Config: MountConfig{
					DefaultLeaseTTL:   "1h",
					MaxLeaseTTL:       "24h",
					ListingVisibility: "hidden",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"default_lease_ttl":            "2h", // changed
		"max_lease_ttl":                "24h",
		"force_no_cache":               false,
		"audit_non_hmac_request_keys":  []string(nil),
		"audit_non_hmac_response_keys": []string(nil),
		"listing_visibility":           "hidden",
		"passthrough_request_headers":  []string(nil),
		"allowed_response_headers":     []string(nil),
	}

	if mount.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different default_lease_ttl) to NOT be equivalent")
	}
}

// Extra fields in the Vault tune response cause IsEquivalentToDesiredState to
// return false because reflect.DeepEqual compares full maps. The reconciler
// does NOT pre-filter the tune response for engine mounts. Story 7-4 tracks
// hardening this behavior.
func TestSecretEngineMountIsEquivalentToDesiredStateExtraFields(t *testing.T) {
	mount := &SecretEngineMount{
		Spec: SecretEngineMountSpec{
			Mount: Mount{
				Config: MountConfig{
					DefaultLeaseTTL:   "1h",
					MaxLeaseTTL:       "24h",
					ListingVisibility: "hidden",
				},
			},
		},
	}

	payload := map[string]interface{}{
		"default_lease_ttl":            "1h",
		"max_lease_ttl":                "24h",
		"force_no_cache":               false,
		"audit_non_hmac_request_keys":  []string(nil),
		"audit_non_hmac_response_keys": []string(nil),
		"listing_visibility":           "hidden",
		"passthrough_request_headers":  []string(nil),
		"allowed_response_headers":     []string(nil),
		"plugin_version":               "",
		"user_lockout_config":          map[string]interface{}{},
	}

	if mount.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (reflect.DeepEqual compares full maps)")
	}
}

func TestSecretEngineMountIsDeletable(t *testing.T) {
	mount := &SecretEngineMount{}
	if !mount.IsDeletable() {
		t.Error("expected SecretEngineMount to be deletable")
	}
}

func TestSecretEngineMountConditions(t *testing.T) {
	mount := &SecretEngineMount{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	mount.SetConditions(conditions)
	got := mount.GetConditions()

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
