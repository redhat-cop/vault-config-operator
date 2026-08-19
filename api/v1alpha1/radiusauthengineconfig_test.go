package v1alpha1

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestRADIUSAuthEngineConfig_toMap(t *testing.T) {
	config := &RADIUSAuthConfig{
		Host:                     "radius.example.com",
		Port:                     1812,
		UnregisteredUserPolicies: "default",
		DialTimeout:              10,
		ReadTimeout:              10,
		NASPort:                  10,
		TokenTTL:                 "1h",
		TokenMaxTTL:              "4h",
		TokenPolicies:            []string{"default", "admin"},
		TokenBoundCIDRs:          []string{"10.0.0.0/8"},
		TokenExplicitMaxTTL:      "8h",
		TokenNoDefaultPolicy:     true,
		TokenNumUses:             5,
		TokenPeriod:              "24h",
		TokenType:                "service",
		retrievedSecret:          "my-shared-secret",
	}

	result := config.toMap()

	if result["host"] != "radius.example.com" {
		t.Errorf("expected host=radius.example.com, got %v", result["host"])
	}
	if result["secret"] != "my-shared-secret" {
		t.Errorf("expected secret=my-shared-secret, got %v", result["secret"])
	}
	if result["port"] != json.Number("1812") {
		t.Errorf("expected port=json.Number(1812), got %v (type %T)", result["port"], result["port"])
	}
	if result["unregistered_user_policies"] != "default" {
		t.Errorf("expected unregistered_user_policies=default, got %v", result["unregistered_user_policies"])
	}
	if result["dial_timeout"] != json.Number("10") {
		t.Errorf("expected dial_timeout=json.Number(10), got %v (type %T)", result["dial_timeout"], result["dial_timeout"])
	}
	if result["read_timeout"] != json.Number("10") {
		t.Errorf("expected read_timeout=json.Number(10), got %v (type %T)", result["read_timeout"], result["read_timeout"])
	}
	if result["nas_port"] != json.Number("10") {
		t.Errorf("expected nas_port=json.Number(10), got %v (type %T)", result["nas_port"], result["nas_port"])
	}
	if result["token_ttl"] != json.Number("3600") {
		t.Errorf("expected token_ttl=3600, got %v", result["token_ttl"])
	}
	if result["token_max_ttl"] != json.Number("14400") {
		t.Errorf("expected token_max_ttl=14400, got %v", result["token_max_ttl"])
	}

	policies, ok := result["token_policies"].([]any)
	if !ok {
		t.Fatalf("expected token_policies to be []any, got %T", result["token_policies"])
	}
	if len(policies) != 2 || policies[0] != "default" || policies[1] != "admin" {
		t.Errorf("expected token_policies=[default admin], got %v", policies)
	}

	cidrs, ok := result["token_bound_cidrs"].([]any)
	if !ok {
		t.Fatalf("expected token_bound_cidrs to be []any, got %T", result["token_bound_cidrs"])
	}
	if len(cidrs) != 1 || cidrs[0] != "10.0.0.0/8" {
		t.Errorf("expected token_bound_cidrs=[10.0.0.0/8], got %v", cidrs)
	}

	if result["token_explicit_max_ttl"] != json.Number("28800") {
		t.Errorf("expected token_explicit_max_ttl=28800, got %v", result["token_explicit_max_ttl"])
	}
	if result["token_no_default_policy"] != true {
		t.Errorf("expected token_no_default_policy=true, got %v", result["token_no_default_policy"])
	}
	if result["token_num_uses"] != json.Number("5") {
		t.Errorf("expected token_num_uses=json.Number(5), got %v (type %T)", result["token_num_uses"], result["token_num_uses"])
	}
	if result["token_period"] != json.Number("86400") {
		t.Errorf("expected token_period=86400, got %v", result["token_period"])
	}
	if result["token_type"] != "service" {
		t.Errorf("expected token_type=service, got %v", result["token_type"])
	}
}

func TestRADIUSAuthEngineConfig_toMap_SparseConfig(t *testing.T) {
	config := &RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		retrievedSecret: "my-shared-secret",
	}

	result := config.toMap()

	if result["host"] != "radius.example.com" {
		t.Errorf("expected host=radius.example.com, got %v", result["host"])
	}
	if result["secret"] != "my-shared-secret" {
		t.Errorf("expected secret=my-shared-secret, got %v", result["secret"])
	}
	if result["port"] != json.Number("1812") {
		t.Errorf("expected port=1812, got %v", result["port"])
	}
	if result["dial_timeout"] != json.Number("10") {
		t.Errorf("expected dial_timeout=10, got %v", result["dial_timeout"])
	}
	if result["read_timeout"] != json.Number("10") {
		t.Errorf("expected read_timeout=10, got %v", result["read_timeout"])
	}
	if result["nas_port"] != json.Number("10") {
		t.Errorf("expected nas_port=10, got %v", result["nas_port"])
	}

	// toMap() always emits all optional fields with zero values for convergence correctness.
	// Verify zero/empty defaults are present.
	zeroFields := map[string]any{
		"unregistered_user_policies": "",
		"token_ttl":                  json.Number("0"),
		"token_max_ttl":              json.Number("0"),
		"token_explicit_max_ttl":     json.Number("0"),
		"token_no_default_policy":    false,
		"token_num_uses":             json.Number("0"),
		"token_period":               json.Number("0"),
		"token_type":                 "default",
	}
	for key, expected := range zeroFields {
		val, exists := result[key]
		if !exists {
			t.Errorf("sparse toMap() must always include %q, but key is missing", key)
		} else if val != expected {
			t.Errorf("sparse toMap()[%q] = %v (%T), want %v (%T)", key, val, val, expected, expected)
		}
	}

	// Slice fields must be present as empty slices
	for _, key := range []string{"token_policies", "token_bound_cidrs"} {
		val, exists := result[key]
		if !exists {
			t.Errorf("sparse toMap() must always include %q, but key is missing", key)
			continue
		}
		s, ok := val.([]any)
		if !ok {
			t.Errorf("sparse toMap()[%q] type = %T, want []any", key, val)
		} else if len(s) != 0 {
			t.Errorf("sparse toMap()[%q] = %v, want empty slice", key, s)
		}
	}
}

func TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_SparseConfig(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.RADIUSAuthConfig = RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		retrievedSecret: "my-shared-secret",
	}

	// Vault omits most unset optionals but returns token_type as "default".
	vaultPayload := map[string]any{
		"host":         "radius.example.com",
		"port":         json.Number("1812"),
		"dial_timeout": json.Number("10"),
		"read_timeout": json.Number("10"),
		"nas_port":     json.Number("10"),
		"token_type":   "default",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: sparse config with Vault defaults must match")
	}
}

func TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_OmittedTokenTypeMatchesVaultDefault(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.RADIUSAuthConfig = RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		retrievedSecret: "my-shared-secret",
		// TokenType omitted — must match Vault's read-default of "default".
	}

	vaultPayload := map[string]any{
		"host":         "radius.example.com",
		"port":         json.Number("1812"),
		"dial_timeout": json.Number("10"),
		"read_timeout": json.Number("10"),
		"nas_port":     json.Number("10"),
		"token_type":   "default",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected omitted tokenType to be equivalent to Vault token_type=default")
	}

	vaultPayload["token_type"] = "service"
	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected omitted tokenType NOT to match an explicit Vault token_type=service")
	}
}

func TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_StaleTokenTTL(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.RADIUSAuthConfig = RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		retrievedSecret: "secret",
	}

	// Vault has a stale token_ttl from a previous configuration; the CR no longer
	// specifies TokenTTL (zero value). The operator must detect the drift and rewrite.
	vaultPayload := map[string]any{
		"host":         "radius.example.com",
		"port":         json.Number("1812"),
		"dial_timeout": json.Number("10"),
		"read_timeout": json.Number("10"),
		"nas_port":     json.Number("10"),
		"token_ttl":    json.Number("3600"),
		"token_type":   "default",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false: stale token_ttl in Vault must trigger rewrite when CR has empty TokenTTL")
	}
}

func TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.RADIUSAuthConfig = RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		TokenType:       "default",
		retrievedSecret: "super-secret",
	}

	vaultPayload := map[string]any{
		"host":         "radius.example.com",
		"port":         json.Number("1812"),
		"dial_timeout": json.Number("10"),
		"read_timeout": json.Number("10"),
		"nas_port":     json.Number("10"),
		"token_type":   "default",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state (secret stripped)")
	}
}

func TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.RADIUSAuthConfig = RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		TokenType:       "default",
		retrievedSecret: "secret",
	}

	vaultPayload := map[string]any{
		"host":         "different-radius.example.com",
		"port":         json.Number("1812"),
		"dial_timeout": json.Number("10"),
		"read_timeout": json.Number("10"),
		"nas_port":     json.Number("10"),
		"token_type":   "default",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched host")
	}
}

func TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.RADIUSAuthConfig = RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		TokenType:       "default",
		retrievedSecret: "secret",
	}

	vaultPayload := map[string]any{
		"host":           "radius.example.com",
		"port":           json.Number("1812"),
		"dial_timeout":   json.Number("10"),
		"read_timeout":   json.Number("10"),
		"nas_port":       json.Number("10"),
		"token_type":     "default",
		"request_id":     "extra-vault-field",
		"lease_duration": json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_SecretStripping(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.RADIUSAuthConfig = RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		TokenType:       "default",
		retrievedSecret: "super-secret-value",
	}

	vaultPayload := map[string]any{
		"host":         "radius.example.com",
		"port":         json.Number("1812"),
		"dial_timeout": json.Number("10"),
		"read_timeout": json.Number("10"),
		"nas_port":     json.Number("10"),
		"token_type":   "default",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: secret is write-only and must be excluded from drift comparison")
	}
}

func TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_SetOrderIndependent(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.RADIUSAuthConfig = RADIUSAuthConfig{
		Host:            "radius.example.com",
		Port:            1812,
		DialTimeout:     10,
		ReadTimeout:     10,
		NASPort:         10,
		TokenPolicies:   []string{"admin", "default", "ops"},
		TokenBoundCIDRs: []string{"192.168.1.0/24", "10.0.0.0/8"},
		retrievedSecret: "secret",
	}

	vaultPayload := map[string]any{
		"host":              "radius.example.com",
		"port":              json.Number("1812"),
		"dial_timeout":      json.Number("10"),
		"read_timeout":      json.Number("10"),
		"nas_port":          json.Number("10"),
		"token_type":        "default",
		"token_policies":    []any{"ops", "default", "admin"},
		"token_bound_cidrs": []any{"10.0.0.0/8", "192.168.1.0/24"},
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: set-valued fields with different order must be treated as equivalent")
	}
}

func TestRADIUSAuthEngineConfig_GetPath(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	instance.Spec.Path = "radius"

	expected := "auth/radius/config"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestRADIUSAuthEngineConfig_IsDeletable(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	if instance.IsDeletable() {
		t.Error("expected IsDeletable() to return false for auth engine config")
	}
}

func TestRADIUSAuthEngineConfig_Conditions(t *testing.T) {
	instance := &RADIUSAuthEngineConfig{}
	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}
	instance.SetConditions(conditions)
	got := instance.GetConditions()
	if len(got) != 1 || got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected conditions round-trip, got %v", got)
	}
}

func TestRADIUSAuthEngineConfig_Default_EmptyPasswordKey(t *testing.T) {
	webhook := &RADIUSAuthEngineConfig{}
	obj := &RADIUSAuthEngineConfig{}
	obj.Spec.RADIUSCredentials.PasswordKey = ""

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.RADIUSCredentials.PasswordKey != "secret" {
		t.Errorf("expected passwordKey=secret, got %s", obj.Spec.RADIUSCredentials.PasswordKey)
	}
}

func TestRADIUSAuthEngineConfig_Default_InheritedSchemaDefault(t *testing.T) {
	webhook := &RADIUSAuthEngineConfig{}
	obj := &RADIUSAuthEngineConfig{}
	obj.Spec.RADIUSCredentials.PasswordKey = "password"

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.RADIUSCredentials.PasswordKey != "secret" {
		t.Errorf("expected inherited default 'password' remapped to 'secret', got %s", obj.Spec.RADIUSCredentials.PasswordKey)
	}
}

func TestRADIUSAuthEngineConfig_Default_PreservesCustomPasswordKey(t *testing.T) {
	webhook := &RADIUSAuthEngineConfig{}
	obj := &RADIUSAuthEngineConfig{}
	obj.Spec.RADIUSCredentials.PasswordKey = "my_custom_key"

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.RADIUSCredentials.PasswordKey != "my_custom_key" {
		t.Errorf("expected custom passwordKey to be preserved, got %s", obj.Spec.RADIUSCredentials.PasswordKey)
	}
}

func TestRADIUSAuthEngineConfig_Default_PreservesExplicitPasswordKey(t *testing.T) {
	webhook := &RADIUSAuthEngineConfig{}
	obj := &RADIUSAuthEngineConfig{}
	obj.Spec.RADIUSCredentials.PasswordKey = "password"

	ctx := admission.NewContextWithRequest(context.Background(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: []byte(`{"spec":{"radiusCredentials":{"passwordKey":"password","secret":{"name":"radius-creds"}}}}`),
			},
		},
	})

	if err := webhook.Default(ctx, obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.RADIUSCredentials.PasswordKey != "password" {
		t.Errorf("expected explicit passwordKey=password to be preserved, got %s", obj.Spec.RADIUSCredentials.PasswordKey)
	}
}

func TestRADIUSAuthEngineConfig_Default_RemapsOmittedPasswordKey(t *testing.T) {
	webhook := &RADIUSAuthEngineConfig{}
	obj := &RADIUSAuthEngineConfig{}
	obj.Spec.RADIUSCredentials.PasswordKey = "password"

	ctx := admission.NewContextWithRequest(context.Background(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{
				Raw: []byte(`{"spec":{"radiusCredentials":{"secret":{"name":"radius-creds"}}}}`),
			},
		},
	})

	if err := webhook.Default(ctx, obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.RADIUSCredentials.PasswordKey != "secret" {
		t.Errorf("expected omitted passwordKey remapped to secret, got %s", obj.Spec.RADIUSCredentials.PasswordKey)
	}
}

func TestRADIUSAuthEngineConfig_Default_RemapsOmittedPasswordKeyOnUpdate(t *testing.T) {
	webhook := &RADIUSAuthEngineConfig{}
	obj := &RADIUSAuthEngineConfig{}
	obj.Spec.RADIUSCredentials.PasswordKey = "password"

	ctx := admission.NewContextWithRequest(context.Background(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Update,
			Object: runtime.RawExtension{
				Raw: []byte(`{"spec":{"radiusCredentials":{"secret":{"name":"radius-creds"}}}}`),
			},
		},
	})

	if err := webhook.Default(ctx, obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.RADIUSCredentials.PasswordKey != "secret" {
		t.Errorf("expected omitted passwordKey remapped to secret on UPDATE, got %s", obj.Spec.RADIUSCredentials.PasswordKey)
	}
}

func TestRADIUSAuthEngineConfig_Default_PreservesExplicitPasswordKeyOnUpdate(t *testing.T) {
	webhook := &RADIUSAuthEngineConfig{}
	obj := &RADIUSAuthEngineConfig{}
	obj.Spec.RADIUSCredentials.PasswordKey = "password"

	ctx := admission.NewContextWithRequest(context.Background(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Update,
			Object: runtime.RawExtension{
				Raw: []byte(`{"spec":{"radiusCredentials":{"passwordKey":"password","secret":{"name":"radius-creds"}}}}`),
			},
		},
	})

	if err := webhook.Default(ctx, obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.RADIUSCredentials.PasswordKey != "password" {
		t.Errorf("expected explicit passwordKey=password preserved on UPDATE, got %s", obj.Spec.RADIUSCredentials.PasswordKey)
	}
}

func TestResolveRADIUSPasswordKey_EmptyDefaultsToSecret(t *testing.T) {
	got := resolveRADIUSPasswordKey("")
	if got != "secret" {
		t.Errorf("resolveRADIUSPasswordKey(\"\") = %q, want \"secret\"", got)
	}
}

func TestResolveRADIUSPasswordKey_PasswordRemappedWhenWebhooksDisabled(t *testing.T) {
	t.Setenv("ENABLE_WEBHOOKS", "false")
	got := resolveRADIUSPasswordKey("password")
	if got != "secret" {
		t.Errorf("resolveRADIUSPasswordKey(\"password\") with ENABLE_WEBHOOKS=false = %q, want \"secret\"", got)
	}
}

func TestResolveRADIUSPasswordKey_PasswordPreservedWhenWebhooksEnabled(t *testing.T) {
	os.Unsetenv("ENABLE_WEBHOOKS")
	got := resolveRADIUSPasswordKey("password")
	if got != "password" {
		t.Errorf("resolveRADIUSPasswordKey(\"password\") with webhooks enabled = %q, want \"password\"", got)
	}
}

func TestResolveRADIUSPasswordKey_CustomKeyPreserved(t *testing.T) {
	got := resolveRADIUSPasswordKey("my_custom_key")
	if got != "my_custom_key" {
		t.Errorf("resolveRADIUSPasswordKey(\"my_custom_key\") = %q, want \"my_custom_key\"", got)
	}
}

func TestResolveRADIUSPasswordKey_PasswordPreservedWhenWebhooksExplicitlyTrue(t *testing.T) {
	t.Setenv("ENABLE_WEBHOOKS", "true")
	got := resolveRADIUSPasswordKey("password")
	if got != "password" {
		t.Errorf("resolveRADIUSPasswordKey(\"password\") with ENABLE_WEBHOOKS=true = %q, want \"password\"", got)
	}
}
