package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAuthEngineMountGetPath(t *testing.T) {
	tests := []struct {
		name         string
		mount        *AuthEngineMount
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			mount: &AuthEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: AuthEngineMountSpec{
					Path: "kubernetes",
					Name: "custom-name",
				},
			},
			expectedPath: "sys/auth/kubernetes/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			mount: &AuthEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: AuthEngineMountSpec{
					Path: "kubernetes",
				},
			},
			expectedPath: "sys/auth/kubernetes/test-mount",
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

func TestAuthMountConfigToMap(t *testing.T) {
	desc := "my auth engine"
	config := AuthMountConfig{
		DefaultLeaseTTL:           "1h",
		MaxLeaseTTL:               "24h",
		AuditNonHMACRequestKeys:   []string{"key1", "key2"},
		AuditNonHMACResponseKeys:  []string{"resp1"},
		ListingVisibility:         "unauth",
		PassthroughRequestHeaders: []string{"X-Custom"},
		AllowedResponseHeaders:    []string{"X-Response"},
		TokenType:                 "default-service",
		Description:               &desc,
		Options:                   map[string]string{"version": "2"},
	}

	result := config.toMap()

	expected := map[string]interface{}{
		"default_lease_ttl":            "1h",
		"max_lease_ttl":                "24h",
		"audit_non_hmac_request_keys":  []string{"key1", "key2"},
		"audit_non_hmac_response_keys": []string{"resp1"},
		"listing_visibility":           "unauth",
		"passthrough_request_headers":  []string{"X-Custom"},
		"allowed_response_headers":     []string{"X-Response"},
		"token_type":                   "default-service",
		"description":                  &desc,
		"options":                      map[string]string{"version": "2"},
	}

	if len(result) != 10 {
		t.Errorf("expected 10 keys in config map, got %d", len(result))
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestAuthMountConfigToMapNilDescription(t *testing.T) {
	config := AuthMountConfig{
		ListingVisibility: "hidden",
	}

	result := config.toMap()

	// The map value is a *string(nil) — not a bare nil interface.
	// reflect.ValueOf correctly detects the nil pointer inside the interface.
	descVal := result["description"]
	if descVal != nil && !reflect.ValueOf(descVal).IsNil() {
		t.Errorf("expected nil *string description when pointer is nil, got %v", descVal)
	}
}

func TestAuthMountToMap(t *testing.T) {
	desc := "my auth"
	config := AuthMountConfig{
		DefaultLeaseTTL:   "1h",
		MaxLeaseTTL:       "24h",
		ListingVisibility: "hidden",
		Description:       &desc,
	}

	mount := AuthMount{
		Type:        "kubernetes",
		Description: "K8s auth",
		Config:      config,
		Local:       true,
		SealWrap:    false,
	}

	result := mount.toMap()

	if len(result) != 5 {
		t.Errorf("expected 5 keys in mount map, got %d", len(result))
	}

	if result["type"] != "kubernetes" {
		t.Errorf("expected type 'kubernetes', got %v", result["type"])
	}
	if result["description"] != "K8s auth" {
		t.Errorf("expected description 'K8s auth', got %v", result["description"])
	}
	if result["local"] != true {
		t.Errorf("expected local true, got %v", result["local"])
	}
	if result["seal_wrap"] != false {
		t.Errorf("expected seal_wrap false, got %v", result["seal_wrap"])
	}

	configMap, ok := result["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected config to be map[string]interface{}")
	}
	if configMap["default_lease_ttl"] != "1h" {
		t.Errorf("expected nested config default_lease_ttl '1h', got %v", configMap["default_lease_ttl"])
	}
}

func TestAuthEngineMountGetPayload(t *testing.T) {
	desc := "tune desc"
	mount := &AuthEngineMount{
		Spec: AuthEngineMountSpec{
			AuthMount: AuthMount{
				Type:        "ldap",
				Description: "LDAP auth",
				Config: AuthMountConfig{
					DefaultLeaseTTL:   "30m",
					MaxLeaseTTL:       "1h",
					ListingVisibility: "hidden",
					Description:       &desc,
				},
				Local:    false,
				SealWrap: true,
			},
		},
	}

	payload := mount.GetPayload()

	// GetPayload returns the full mount spec via AuthMount.toMap()
	if payload["type"] != "ldap" {
		t.Errorf("expected type 'ldap', got %v", payload["type"])
	}
	if payload["description"] != "LDAP auth" {
		t.Errorf("expected description 'LDAP auth', got %v", payload["description"])
	}
	if payload["seal_wrap"] != true {
		t.Errorf("expected seal_wrap true, got %v", payload["seal_wrap"])
	}

	configMap, ok := payload["config"].(map[string]interface{})
	if !ok {
		t.Fatal("expected payload config to be map[string]interface{}")
	}
	if configMap["default_lease_ttl"] != "30m" {
		t.Errorf("expected nested default_lease_ttl '30m', got %v", configMap["default_lease_ttl"])
	}
}

func TestAuthEngineMountGetTunePayload(t *testing.T) {
	desc := "tune desc"
	mount := &AuthEngineMount{
		Spec: AuthEngineMountSpec{
			AuthMount: AuthMount{
				Type:        "ldap",
				Description: "LDAP auth",
				Config: AuthMountConfig{
					DefaultLeaseTTL:   "30m",
					MaxLeaseTTL:       "1h",
					ListingVisibility: "hidden",
					TokenType:         "default-service",
					Description:       &desc,
					Options:           map[string]string{"version": "1"},
				},
				Local:    false,
				SealWrap: true,
			},
		},
	}

	tunePayload := mount.GetTunePayload()

	// GetTunePayload returns only Config.toMap(), not the full mount spec
	if _, ok := tunePayload["type"]; ok {
		t.Error("expected tune payload to NOT contain 'type' (mount-level field)")
	}
	if _, ok := tunePayload["seal_wrap"]; ok {
		t.Error("expected tune payload to NOT contain 'seal_wrap' (mount-level field)")
	}

	if tunePayload["default_lease_ttl"] != "30m" {
		t.Errorf("expected default_lease_ttl '30m', got %v", tunePayload["default_lease_ttl"])
	}
	if tunePayload["token_type"] != "default-service" {
		t.Errorf("expected token_type 'default-service', got %v", tunePayload["token_type"])
	}

	// AuthEngineMount: GetTunePayload and IsEquivalentToDesiredState use the
	// same map (Config.toMap()), so they should be identical.
	configMap := mount.Spec.Config.toMap()
	if !reflect.DeepEqual(tunePayload, configMap) {
		t.Errorf("GetTunePayload() should equal Config.toMap()")
	}
}

func TestAuthEngineMountIsEquivalentToDesiredStateMatching(t *testing.T) {
	desc := "my auth"
	mount := &AuthEngineMount{
		Spec: AuthEngineMountSpec{
			AuthMount: AuthMount{
				Config: AuthMountConfig{
					DefaultLeaseTTL:           "1h",
					MaxLeaseTTL:               "24h",
					AuditNonHMACRequestKeys:   []string{"key1"},
					AuditNonHMACResponseKeys:  []string{"resp1"},
					ListingVisibility:         "hidden",
					PassthroughRequestHeaders: []string{"X-Custom"},
					AllowedResponseHeaders:    []string{"X-Resp"},
					TokenType:                 "default-service",
					Description:               &desc,
					Options:                   map[string]string{"version": "2"},
				},
			},
		},
	}

	// Payload constructed from Config.toMap() — proves the comparison map
	// matches its own output (the implementation-level contract).
	payload := mount.Spec.Config.toMap()

	if !mount.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching tune config payload to be equivalent")
	}
}

// IsEquivalentToDesiredState compares only Config.toMap() (tune config), NOT
// the full mount spec from GetPayload(). This test proves AC #2: passing the
// full mount spec (which includes type, local, seal_wrap, etc.) must return
// false.
func TestAuthEngineMountIsEquivalentRejectsFullMountPayload(t *testing.T) {
	desc := "my auth"
	mount := &AuthEngineMount{
		Spec: AuthEngineMountSpec{
			AuthMount: AuthMount{
				Type:        "kubernetes",
				Description: "K8s auth",
				Config: AuthMountConfig{
					DefaultLeaseTTL:   "1h",
					MaxLeaseTTL:       "24h",
					ListingVisibility: "hidden",
					Description:       &desc,
				},
				Local:    false,
				SealWrap: true,
			},
		},
	}

	fullPayload := mount.GetPayload()

	if mount.IsEquivalentToDesiredState(fullPayload) {
		t.Error("expected full mount spec payload (from GetPayload) to NOT be equivalent — comparison should be tune-only")
	}
}

func TestAuthEngineMountIsEquivalentToDesiredStateNonMatching(t *testing.T) {
	desc := "my auth"
	mount := &AuthEngineMount{
		Spec: AuthEngineMountSpec{
			AuthMount: AuthMount{
				Config: AuthMountConfig{
					DefaultLeaseTTL:   "1h",
					MaxLeaseTTL:       "24h",
					ListingVisibility: "hidden",
					Description:       &desc,
				},
			},
		},
	}

	payload := map[string]interface{}{
		"default_lease_ttl":            "2h", // changed
		"max_lease_ttl":                "24h",
		"audit_non_hmac_request_keys":  []string(nil),
		"audit_non_hmac_response_keys": []string(nil),
		"listing_visibility":           "hidden",
		"passthrough_request_headers":  []string(nil),
		"allowed_response_headers":     []string(nil),
		"token_type":                   "",
		"description":                  &desc,
		"options":                      map[string]string(nil),
	}

	if mount.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different default_lease_ttl) to NOT be equivalent")
	}
}

// Extra fields in the Vault tune response cause IsEquivalentToDesiredState to
// return false because reflect.DeepEqual compares full maps. The reconciler
// does NOT pre-filter the tune response for engine mounts. Story 7-4 tracks
// hardening this behavior.
func TestAuthEngineMountIsEquivalentToDesiredStateExtraFields(t *testing.T) {
	mount := &AuthEngineMount{
		Spec: AuthEngineMountSpec{
			AuthMount: AuthMount{
				Config: AuthMountConfig{
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
		"audit_non_hmac_request_keys":  []string(nil),
		"audit_non_hmac_response_keys": []string(nil),
		"listing_visibility":           "hidden",
		"passthrough_request_headers":  []string(nil),
		"allowed_response_headers":     []string(nil),
		"token_type":                   "",
		"description":                  (*string)(nil),
		"options":                      map[string]string(nil),
		"force_no_cache":               false,
		"plugin_version":               "",
	}

	if mount.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (reflect.DeepEqual compares full maps)")
	}
}

func TestAuthEngineMountIsDeletable(t *testing.T) {
	mount := &AuthEngineMount{}
	if !mount.IsDeletable() {
		t.Error("expected AuthEngineMount to be deletable")
	}
}

func TestAuthEngineMountConditions(t *testing.T) {
	mount := &AuthEngineMount{}

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
