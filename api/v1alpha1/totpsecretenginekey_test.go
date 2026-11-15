package v1alpha1

import (
	"encoding/json"
	"testing"
)

func intPtr(i int) *int { return &i }

func TestTOTPSecretEngineKey_toMap_GenerateMode(t *testing.T) {
	exported := true
	cfg := TOTPKeyConfig{
		Generate:    true,
		Exported:    &exported,
		KeySize:     32,
		Issuer:      "Google",
		AccountName: "test@gmail.com",
		Period:      30,
		Algorithm:   "SHA1",
		Digits:      6,
		Skew:        intPtr(1),
		QRSize:      intPtr(200),
	}

	result := cfg.toMap()

	if result["generate"] != true {
		t.Errorf("expected generate=true, got %v", result["generate"])
	}
	if result["issuer"] != "Google" {
		t.Errorf("expected issuer=Google, got %v", result["issuer"])
	}
	if result["account_name"] != "test@gmail.com" {
		t.Errorf("expected account_name=test@gmail.com, got %v", result["account_name"])
	}
	if result["algorithm"] != "SHA1" {
		t.Errorf("expected algorithm=SHA1, got %v", result["algorithm"])
	}
	if result["digits"] != json.Number("6") {
		t.Errorf("expected digits=json.Number(\"6\"), got %v (type %T)", result["digits"], result["digits"])
	}
	if result["period"] != json.Number("30") {
		t.Errorf("expected period=json.Number(\"30\"), got %v (type %T)", result["period"], result["period"])
	}
	if result["exported"] != true {
		t.Errorf("expected exported=true, got %v", result["exported"])
	}
	if result["key_size"] != json.Number("32") {
		t.Errorf("expected key_size=json.Number(\"32\"), got %v (type %T)", result["key_size"], result["key_size"])
	}
	if result["skew"] != json.Number("1") {
		t.Errorf("expected skew=json.Number(\"1\"), got %v (type %T)", result["skew"], result["skew"])
	}
	if result["qr_size"] != json.Number("200") {
		t.Errorf("expected qr_size=json.Number(\"200\"), got %v (type %T)", result["qr_size"], result["qr_size"])
	}

	if _, exists := result["key"]; exists {
		t.Error("key should NOT be in generate-mode output")
	}
	if _, exists := result["url"]; exists {
		t.Error("url should NOT be in generate-mode output")
	}
}

func TestTOTPSecretEngineKey_toMap_ImportModeWithKey(t *testing.T) {
	cfg := TOTPKeyConfig{
		Generate:    false,
		Key:         "JBSWY3DPEHPK3PXP",
		Issuer:      "MyApp",
		AccountName: "user@example.com",
		Period:      30,
		Algorithm:   "SHA256",
		Digits:      8,
	}

	result := cfg.toMap()

	if result["generate"] != false {
		t.Errorf("expected generate=false, got %v", result["generate"])
	}
	if result["key"] != "JBSWY3DPEHPK3PXP" {
		t.Errorf("expected key=JBSWY3DPEHPK3PXP, got %v", result["key"])
	}
	if result["issuer"] != "MyApp" {
		t.Errorf("expected issuer=MyApp, got %v", result["issuer"])
	}
	if result["account_name"] != "user@example.com" {
		t.Errorf("expected account_name=user@example.com, got %v", result["account_name"])
	}
	if result["algorithm"] != "SHA256" {
		t.Errorf("expected algorithm=SHA256, got %v", result["algorithm"])
	}
	if result["digits"] != json.Number("8") {
		t.Errorf("expected digits=json.Number(\"8\"), got %v (type %T)", result["digits"], result["digits"])
	}
	if result["period"] != json.Number("30") {
		t.Errorf("expected period=json.Number(\"30\"), got %v (type %T)", result["period"], result["period"])
	}

	if _, exists := result["exported"]; exists {
		t.Error("exported should NOT be in import-mode output")
	}
	if _, exists := result["key_size"]; exists {
		t.Error("key_size should NOT be in import-mode output")
	}
	if _, exists := result["skew"]; exists {
		t.Error("skew should NOT be in import-mode output")
	}
	if _, exists := result["qr_size"]; exists {
		t.Error("qr_size should NOT be in import-mode output")
	}
	if _, exists := result["url"]; exists {
		t.Error("url should NOT be in import-mode output when key is set")
	}
}

func TestTOTPSecretEngineKey_toMap_ImportModeWithURL(t *testing.T) {
	cfg := TOTPKeyConfig{
		Generate:    false,
		URL:         "otpauth://totp/Example:alice@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Example",
		Issuer:      "Example",
		AccountName: "alice@example.com",
		Period:      30,
		Algorithm:   "SHA1",
		Digits:      6,
	}

	result := cfg.toMap()

	if result["url"] != "otpauth://totp/Example:alice@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Example" {
		t.Errorf("expected url field, got %v", result["url"])
	}
	if _, exists := result["key"]; exists {
		t.Error("key should NOT be in output when url is set")
	}
}

func TestTOTPSecretEngineKey_toMap_Defaults(t *testing.T) {
	exported := false
	cfg := TOTPKeyConfig{
		Generate: true,
		Exported: &exported,
		Issuer:   "DefaultTest",
	}

	result := cfg.toMap()

	if result["algorithm"] != "SHA1" {
		t.Errorf("expected default algorithm=SHA1, got %v", result["algorithm"])
	}
	if result["digits"] != json.Number("6") {
		t.Errorf("expected default digits=json.Number(\"6\"), got %v", result["digits"])
	}
	if result["period"] != json.Number("30") {
		t.Errorf("expected default period=json.Number(\"30\"), got %v", result["period"])
	}
	if result["key_size"] != json.Number("20") {
		t.Errorf("expected default key_size=json.Number(\"20\"), got %v", result["key_size"])
	}
	if _, exists := result["qr_size"]; exists {
		t.Error("qr_size should NOT be in output when QRSize is nil (let Vault default)")
	}
	if _, exists := result["skew"]; exists {
		t.Error("skew should NOT be in output when Skew is nil (let Vault default)")
	}
	if result["exported"] != false {
		t.Errorf("expected exported=false, got %v", result["exported"])
	}
}

func TestTOTPSecretEngineKey_toMap_QRSizeZero(t *testing.T) {
	exported := true
	cfg := TOTPKeyConfig{
		Generate:    true,
		Exported:    &exported,
		Issuer:      "TestOrg",
		AccountName: "user@test.com",
		QRSize:      intPtr(0),
	}

	result := cfg.toMap()

	if result["qr_size"] != json.Number("0") {
		t.Errorf("expected qr_size=json.Number(\"0\") for explicit zero, got %v (type %T)", result["qr_size"], result["qr_size"])
	}
}

func TestTOTPSecretEngineKey_toMap_SkewZero(t *testing.T) {
	exported := true
	cfg := TOTPKeyConfig{
		Generate:    true,
		Exported:    &exported,
		Issuer:      "TestOrg",
		AccountName: "user@test.com",
		Skew:        intPtr(0),
	}

	result := cfg.toMap()

	if result["skew"] != json.Number("0") {
		t.Errorf("expected skew=json.Number(\"0\") for explicit zero, got %v (type %T)", result["skew"], result["skew"])
	}
}

func TestTOTPSecretEngineKey_readVisibleMap(t *testing.T) {
	cfg := TOTPKeyConfig{
		Generate:    true,
		Exported:    nil,
		KeySize:     32,
		Key:         "should-not-appear",
		URL:         "should-not-appear",
		Issuer:      "Google",
		AccountName: "test@gmail.com",
		Period:      30,
		Algorithm:   "SHA1",
		Digits:      6,
		Skew:        intPtr(1),
		QRSize:      intPtr(200),
	}

	result := cfg.readVisibleMap()

	expectedKeys := map[string]bool{
		"issuer":       true,
		"account_name": true,
		"algorithm":    true,
		"digits":       true,
		"period":       true,
	}

	for key := range result {
		if !expectedKeys[key] {
			t.Errorf("unexpected key %q in readVisibleMap output", key)
		}
	}
	if len(result) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d", len(expectedKeys), len(result))
	}

	if result["issuer"] != "Google" {
		t.Errorf("expected issuer=Google, got %v", result["issuer"])
	}
	if result["account_name"] != "test@gmail.com" {
		t.Errorf("expected account_name=test@gmail.com, got %v", result["account_name"])
	}
	if result["algorithm"] != "SHA1" {
		t.Errorf("expected algorithm=SHA1, got %v", result["algorithm"])
	}
	if result["digits"] != json.Number("6") {
		t.Errorf("expected digits=json.Number(\"6\"), got %v (type %T)", result["digits"], result["digits"])
	}
	if result["period"] != json.Number("30") {
		t.Errorf("expected period=json.Number(\"30\"), got %v (type %T)", result["period"], result["period"])
	}
}

func TestTOTPSecretEngineKey_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &TOTPSecretEngineKey{}
	instance.Spec.Issuer = "Google"
	instance.Spec.AccountName = "test@gmail.com"
	instance.Spec.Algorithm = "SHA1"
	instance.Spec.Digits = 6
	instance.Spec.Period = 30

	vaultPayload := map[string]any{
		"account_name": "test@gmail.com",
		"algorithm":    "SHA1",
		"digits":       json.Number("6"),
		"issuer":       "Google",
		"period":       json.Number("30"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching payload")
	}
}

func TestTOTPSecretEngineKey_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &TOTPSecretEngineKey{}
	instance.Spec.Issuer = "Google"
	instance.Spec.AccountName = "test@gmail.com"
	instance.Spec.Algorithm = "SHA1"
	instance.Spec.Digits = 6
	instance.Spec.Period = 30

	vaultPayload := map[string]any{
		"account_name": "test@gmail.com",
		"algorithm":    "SHA1",
		"digits":       json.Number("6"),
		"issuer":       "DifferentIssuer",
		"period":       json.Number("30"),
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false when issuer differs")
	}
}

func TestTOTPSecretEngineKey_IsEquivalentToDesiredState_MismatchDigits(t *testing.T) {
	instance := &TOTPSecretEngineKey{}
	instance.Spec.Issuer = "Google"
	instance.Spec.AccountName = "test@gmail.com"
	instance.Spec.Algorithm = "SHA1"
	instance.Spec.Digits = 6
	instance.Spec.Period = 30

	vaultPayload := map[string]any{
		"account_name": "test@gmail.com",
		"algorithm":    "SHA1",
		"digits":       json.Number("8"),
		"issuer":       "Google",
		"period":       json.Number("30"),
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false when digits differ")
	}
}

func TestTOTPSecretEngineKey_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &TOTPSecretEngineKey{}
	instance.Spec.Issuer = "Google"
	instance.Spec.AccountName = "test@gmail.com"
	instance.Spec.Algorithm = "SHA1"
	instance.Spec.Digits = 6
	instance.Spec.Period = 30

	vaultPayload := map[string]any{
		"account_name":   "test@gmail.com",
		"algorithm":      "SHA1",
		"digits":         json.Number("6"),
		"issuer":         "Google",
		"period":         json.Number("30"),
		"extra_field":    "should-be-ignored",
		"another_field":  42,
		"request_id":     "abc-123",
		"lease_duration": 0,
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true even with extra Vault fields")
	}
}

func TestTOTPSecretEngineKey_IsEquivalentToDesiredState_DefaultValues(t *testing.T) {
	instance := &TOTPSecretEngineKey{}
	instance.Spec.Issuer = "TestOrg"
	instance.Spec.AccountName = "user@test.com"

	vaultPayload := map[string]any{
		"account_name": "user@test.com",
		"algorithm":    "SHA1",
		"digits":       json.Number("6"),
		"issuer":       "TestOrg",
		"period":       json.Number("30"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to match with default values for algorithm/digits/period")
	}
}
