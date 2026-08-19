package v1alpha1

import (
	"context"
	"encoding/json"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestCFAuthEngineConfig_toMap(t *testing.T) {
	config := &CFAuthConfig{
		IdentityCACertificates:    []string{"-----BEGIN CERTIFICATE-----\ncert1\n-----END CERTIFICATE-----"},
		CFAPIAddr:                 "https://api.sys.example.cf-app.com",
		CFAPITrustedCertificates:  []string{"-----BEGIN CERTIFICATE-----\ntrusted1\n-----END CERTIFICATE-----"},
		LoginMaxSecondsNotBefore:  300,
		LoginMaxSecondsNotAfter:   60,
		CFAPIMutualTLSCertificate: "mutual-tls-cert",
		CFAPIMutualTLSKey:         "mutual-tls-key",
		retrievedCFUsername:       "vault-cf-user",
		retrievedCFPassword:       "vault-cf-pass",
	}

	result := config.toMap()

	if result["cf_api_addr"] != "https://api.sys.example.cf-app.com" {
		t.Errorf("expected cf_api_addr=https://api.sys.example.cf-app.com, got %v", result["cf_api_addr"])
	}
	if result["cf_username"] != "vault-cf-user" {
		t.Errorf("expected cf_username=vault-cf-user, got %v", result["cf_username"])
	}
	if result["cf_password"] != "vault-cf-pass" {
		t.Errorf("expected cf_password=vault-cf-pass, got %v", result["cf_password"])
	}

	caCerts, ok := result["identity_ca_certificates"].([]any)
	if !ok {
		t.Fatalf("expected identity_ca_certificates to be []any, got %T", result["identity_ca_certificates"])
	}
	if len(caCerts) != 1 || caCerts[0] != "-----BEGIN CERTIFICATE-----\ncert1\n-----END CERTIFICATE-----" {
		t.Errorf("unexpected identity_ca_certificates: %v", caCerts)
	}

	trustedCerts, ok := result["cf_api_trusted_certificates"].([]any)
	if !ok {
		t.Fatalf("expected cf_api_trusted_certificates to be []any, got %T", result["cf_api_trusted_certificates"])
	}
	if len(trustedCerts) != 1 {
		t.Errorf("expected 1 trusted certificate, got %d", len(trustedCerts))
	}

	if result["login_max_seconds_not_before"] != json.Number("300") {
		t.Errorf("expected login_max_seconds_not_before=json.Number(300), got %v (type %T)", result["login_max_seconds_not_before"], result["login_max_seconds_not_before"])
	}
	if result["login_max_seconds_not_after"] != json.Number("60") {
		t.Errorf("expected login_max_seconds_not_after=json.Number(60), got %v (type %T)", result["login_max_seconds_not_after"], result["login_max_seconds_not_after"])
	}
	if result["cf_api_mutual_tls_certificate"] != "mutual-tls-cert" {
		t.Errorf("expected cf_api_mutual_tls_certificate=mutual-tls-cert, got %v", result["cf_api_mutual_tls_certificate"])
	}
	if result["cf_api_mutual_tls_key"] != "mutual-tls-key" {
		t.Errorf("expected cf_api_mutual_tls_key=mutual-tls-key, got %v", result["cf_api_mutual_tls_key"])
	}
}

func TestCFAuthEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &CFAuthEngineConfig{}
	instance.Spec.CFAuthConfig = CFAuthConfig{
		IdentityCACertificates:   []string{"-----BEGIN CERTIFICATE-----\ncert1\n-----END CERTIFICATE-----"},
		CFAPIAddr:                "https://api.sys.example.cf-app.com",
		CFAPITrustedCertificates: nil,
		LoginMaxSecondsNotBefore: 5,
		LoginMaxSecondsNotAfter:  1,
		retrievedCFUsername:      "vault",
		retrievedCFPassword:      "super-secret",
	}

	vaultPayload := map[string]any{
		"identity_ca_certificates":     []any{"-----BEGIN CERTIFICATE-----\ncert1\n-----END CERTIFICATE-----"},
		"cf_api_addr":                  "https://api.sys.example.cf-app.com",
		"cf_username":                  "vault",
		"login_max_seconds_not_before": json.Number("5"),
		"login_max_seconds_not_after":  json.Number("1"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state (cf_password and cf_api_mutual_tls_key stripped, unset optionals removed)")
	}
}

func TestCFAuthEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &CFAuthEngineConfig{}
	instance.Spec.CFAuthConfig = CFAuthConfig{
		IdentityCACertificates:   []string{"cert1"},
		CFAPIAddr:                "https://api.sys.example.cf-app.com",
		LoginMaxSecondsNotBefore: 300,
		LoginMaxSecondsNotAfter:  60,
		retrievedCFUsername:      "vault",
		retrievedCFPassword:      "secret",
	}

	vaultPayload := map[string]any{
		"identity_ca_certificates":     []any{"cert1"},
		"cf_api_addr":                  "https://api.sys.DIFFERENT.cf-app.com",
		"cf_username":                  "vault",
		"login_max_seconds_not_before": json.Number("300"),
		"login_max_seconds_not_after":  json.Number("60"),
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched cf_api_addr")
	}
}

func TestCFAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &CFAuthEngineConfig{}
	instance.Spec.CFAuthConfig = CFAuthConfig{
		IdentityCACertificates:   []string{"cert1"},
		CFAPIAddr:                "https://api.sys.example.cf-app.com",
		LoginMaxSecondsNotBefore: 300,
		LoginMaxSecondsNotAfter:  60,
		retrievedCFUsername:      "vault",
		retrievedCFPassword:      "secret",
	}

	vaultPayload := map[string]any{
		"identity_ca_certificates":     []any{"cert1"},
		"cf_api_addr":                  "https://api.sys.example.cf-app.com",
		"cf_username":                  "vault",
		"login_max_seconds_not_before": json.Number("300"),
		"login_max_seconds_not_after":  json.Number("60"),
		"request_id":                   "extra-vault-field",
		"lease_duration":               json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestCFAuthEngineConfig_IsEquivalentToDesiredState_PasswordStripping(t *testing.T) {
	instance := &CFAuthEngineConfig{}
	instance.Spec.CFAuthConfig = CFAuthConfig{
		IdentityCACertificates:    []string{"cert1"},
		CFAPIAddr:                 "https://api.sys.example.cf-app.com",
		LoginMaxSecondsNotBefore:  300,
		LoginMaxSecondsNotAfter:   60,
		CFAPIMutualTLSCertificate: "",
		CFAPIMutualTLSKey:         "write-only-key",
		retrievedCFUsername:       "vault",
		retrievedCFPassword:       "super-secret-password",
	}

	vaultPayload := map[string]any{
		"identity_ca_certificates":     []any{"cert1"},
		"cf_api_addr":                  "https://api.sys.example.cf-app.com",
		"cf_username":                  "vault",
		"login_max_seconds_not_before": json.Number("300"),
		"login_max_seconds_not_after":  json.Number("60"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: cf_password and cf_api_mutual_tls_key are write-only and must be excluded from drift comparison")
	}
}

func TestCFAuthEngineConfig_IsEquivalentToDesiredState_UnsetOptionals(t *testing.T) {
	instance := &CFAuthEngineConfig{}
	instance.Spec.CFAuthConfig = CFAuthConfig{
		IdentityCACertificates:    []string{"cert1"},
		CFAPIAddr:                 "https://api.sys.example.cf-app.com",
		LoginMaxSecondsNotBefore:  300,
		LoginMaxSecondsNotAfter:   60,
		CFAPITrustedCertificates:  nil,
		CFAPIMutualTLSCertificate: "",
		retrievedCFUsername:       "vault",
		retrievedCFPassword:       "secret",
	}

	vaultPayload := map[string]any{
		"identity_ca_certificates":     []any{"cert1"},
		"cf_api_addr":                  "https://api.sys.example.cf-app.com",
		"cf_username":                  "vault",
		"login_max_seconds_not_before": json.Number("300"),
		"login_max_seconds_not_after":  json.Number("60"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: empty optional fields (cf_api_trusted_certificates, cf_api_mutual_tls_certificate) should be removed by removeUnsetFields")
	}
}

func TestCFAuthEngineConfig_IsEquivalentToDesiredState_ReorderedCertsEquivalent(t *testing.T) {
	instance := &CFAuthEngineConfig{}
	instance.Spec.CFAuthConfig = CFAuthConfig{
		IdentityCACertificates: []string{
			"-----BEGIN CERTIFICATE-----\ncertA\n-----END CERTIFICATE-----",
			"-----BEGIN CERTIFICATE-----\ncertB\n-----END CERTIFICATE-----",
		},
		CFAPIAddr: "https://api.sys.example.cf-app.com",
		CFAPITrustedCertificates: []string{
			"-----BEGIN CERTIFICATE-----\ntrustedX\n-----END CERTIFICATE-----",
			"-----BEGIN CERTIFICATE-----\ntrustedY\n-----END CERTIFICATE-----",
		},
		LoginMaxSecondsNotBefore: 300,
		LoginMaxSecondsNotAfter:  60,
		retrievedCFUsername:      "vault",
		retrievedCFPassword:      "secret",
	}

	vaultPayload := map[string]any{
		"identity_ca_certificates": []any{
			"-----BEGIN CERTIFICATE-----\ncertB\n-----END CERTIFICATE-----",
			"-----BEGIN CERTIFICATE-----\ncertA\n-----END CERTIFICATE-----",
		},
		"cf_api_addr": "https://api.sys.example.cf-app.com",
		"cf_api_trusted_certificates": []any{
			"-----BEGIN CERTIFICATE-----\ntrustedY\n-----END CERTIFICATE-----",
			"-----BEGIN CERTIFICATE-----\ntrustedX\n-----END CERTIFICATE-----",
		},
		"cf_username":                  "vault",
		"login_max_seconds_not_before": json.Number("300"),
		"login_max_seconds_not_after":  json.Number("60"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when cert fields are reordered (set semantics)")
	}
}

func TestCFAuthEngineConfig_GetPath(t *testing.T) {
	instance := &CFAuthEngineConfig{}
	instance.Spec.Path = "cf"

	expected := "auth/cf/config"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestCFAuthEngineConfig_IsDeletable(t *testing.T) {
	instance := &CFAuthEngineConfig{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true for CF auth engine config (has DELETE endpoint)")
	}
}

func TestCFAuthEngineConfig_Default_EmptyKeys(t *testing.T) {
	webhook := &CFAuthEngineConfig{}
	obj := &CFAuthEngineConfig{}
	obj.Spec.CFCredentials.UsernameKey = ""
	obj.Spec.CFCredentials.PasswordKey = ""

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.CFCredentials.UsernameKey != "cf_username" {
		t.Errorf("expected usernameKey=cf_username, got %s", obj.Spec.CFCredentials.UsernameKey)
	}
	if obj.Spec.CFCredentials.PasswordKey != "cf_password" {
		t.Errorf("expected passwordKey=cf_password, got %s", obj.Spec.CFCredentials.PasswordKey)
	}
}

func TestCFAuthEngineConfig_Default_InheritedSchemaDefaults(t *testing.T) {
	webhook := &CFAuthEngineConfig{}
	obj := &CFAuthEngineConfig{}
	obj.Spec.CFCredentials.UsernameKey = "username"
	obj.Spec.CFCredentials.PasswordKey = "password"

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.CFCredentials.UsernameKey != "cf_username" {
		t.Errorf("expected usernameKey=cf_username when inherited default 'username' is present, got %s", obj.Spec.CFCredentials.UsernameKey)
	}
	if obj.Spec.CFCredentials.PasswordKey != "cf_password" {
		t.Errorf("expected passwordKey=cf_password when inherited default 'password' is present, got %s", obj.Spec.CFCredentials.PasswordKey)
	}
}

func TestCFAuthEngineConfig_Default_ExplicitEmptyStringGetsRemapped(t *testing.T) {
	webhook := &CFAuthEngineConfig{}
	obj := &CFAuthEngineConfig{}
	obj.Spec.CFCredentials.UsernameKey = "username"
	obj.Spec.CFCredentials.PasswordKey = "password"

	rawObj := []byte(`{"spec":{"cfCredentials":{"usernameKey":"","passwordKey":""}}}`)
	ctx := contextWithAdmissionRequest(t, rawObj)

	if err := webhook.Default(ctx, obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.CFCredentials.UsernameKey != "cf_username" {
		t.Errorf("expected explicit empty usernameKey to be remapped to cf_username, got %s", obj.Spec.CFCredentials.UsernameKey)
	}
	if obj.Spec.CFCredentials.PasswordKey != "cf_password" {
		t.Errorf("expected explicit empty passwordKey to be remapped to cf_password, got %s", obj.Spec.CFCredentials.PasswordKey)
	}
}

func TestCFAuthEngineConfig_Default_PreservesCustomKeys(t *testing.T) {
	webhook := &CFAuthEngineConfig{}
	obj := &CFAuthEngineConfig{}
	obj.Spec.CFCredentials.UsernameKey = "my_custom_user"
	obj.Spec.CFCredentials.PasswordKey = "my_custom_pass"

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.CFCredentials.UsernameKey != "my_custom_user" {
		t.Errorf("expected custom usernameKey to be preserved, got %s", obj.Spec.CFCredentials.UsernameKey)
	}
	if obj.Spec.CFCredentials.PasswordKey != "my_custom_pass" {
		t.Errorf("expected custom passwordKey to be preserved, got %s", obj.Spec.CFCredentials.PasswordKey)
	}
}

func TestCFAuthEngineConfig_Default_UpdateRemapsEmptyKeys(t *testing.T) {
	webhook := &CFAuthEngineConfig{}
	obj := &CFAuthEngineConfig{}
	obj.Spec.CFCredentials.UsernameKey = ""
	obj.Spec.CFCredentials.PasswordKey = ""

	rawObj := []byte(`{"spec":{"cfCredentials":{"usernameKey":"","passwordKey":""}}}`)
	ctx := contextWithAdmissionRequest(t, rawObj)

	if err := webhook.Default(ctx, obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.CFCredentials.UsernameKey != "cf_username" {
		t.Errorf("expected update with empty usernameKey to be remapped to cf_username, got %s", obj.Spec.CFCredentials.UsernameKey)
	}
	if obj.Spec.CFCredentials.PasswordKey != "cf_password" {
		t.Errorf("expected update with empty passwordKey to be remapped to cf_password, got %s", obj.Spec.CFCredentials.PasswordKey)
	}
}

func contextWithAdmissionRequest(t *testing.T, rawObj []byte) context.Context {
	t.Helper()
	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{Raw: rawObj},
		},
	}
	return admission.NewContextWithRequest(context.Background(), req)
}
