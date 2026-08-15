package v1alpha1

import (
	"context"
	"encoding/json"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
)

func TestAWSAuthEngineClientConfig_toMap(t *testing.T) {
	maxRetries := 3
	config := &AWSAuthClientConfig{
		AccessKey:              "AKIAIOSFODNN7EXAMPLE",
		Endpoint:               "https://ec2.custom.endpoint.com",
		IAMEndpoint:            "https://iam.custom.endpoint.com",
		STSEndpoint:            "https://sts.custom.endpoint.com",
		STSRegion:              "us-west-2",
		UseSTSRegionFromClient: true,
		IAMServerIDHeaderValue: "vault.example.com",
		AllowedSTSHeaderValues: "X-Custom-Header",
		MaxRetries:             &maxRetries,
		retrievedAccessKey:     "RESOLVED_ACCESS_KEY",
		retrievedSecretKey:     "RESOLVED_SECRET_KEY",
	}

	result := config.toMap()

	if result["access_key"] != "RESOLVED_ACCESS_KEY" {
		t.Errorf("expected access_key=RESOLVED_ACCESS_KEY, got %v", result["access_key"])
	}
	if result["secret_key"] != "RESOLVED_SECRET_KEY" {
		t.Errorf("expected secret_key=RESOLVED_SECRET_KEY, got %v", result["secret_key"])
	}
	if result["endpoint"] != "https://ec2.custom.endpoint.com" {
		t.Errorf("expected endpoint=https://ec2.custom.endpoint.com, got %v", result["endpoint"])
	}
	if result["iam_endpoint"] != "https://iam.custom.endpoint.com" {
		t.Errorf("expected iam_endpoint=https://iam.custom.endpoint.com, got %v", result["iam_endpoint"])
	}
	if result["sts_endpoint"] != "https://sts.custom.endpoint.com" {
		t.Errorf("expected sts_endpoint=https://sts.custom.endpoint.com, got %v", result["sts_endpoint"])
	}
	if result["sts_region"] != "us-west-2" {
		t.Errorf("expected sts_region=us-west-2, got %v", result["sts_region"])
	}
	if result["use_sts_region_from_client"] != true {
		t.Errorf("expected use_sts_region_from_client=true, got %v", result["use_sts_region_from_client"])
	}
	if result["iam_server_id_header_value"] != "vault.example.com" {
		t.Errorf("expected iam_server_id_header_value=vault.example.com, got %v", result["iam_server_id_header_value"])
	}
	if result["allowed_sts_header_values"] != "X-Custom-Header" {
		t.Errorf("expected allowed_sts_header_values=X-Custom-Header, got %v", result["allowed_sts_header_values"])
	}
	if result["max_retries"] != json.Number("3") {
		t.Errorf("expected max_retries=json.Number(3), got %v", result["max_retries"])
	}
}

func TestAWSAuthEngineClientConfig_toMap_DefaultMaxRetries(t *testing.T) {
	config := &AWSAuthClientConfig{
		retrievedAccessKey: "KEY",
		retrievedSecretKey: "SECRET",
	}

	result := config.toMap()

	if result["max_retries"] != json.Number("-1") {
		t.Errorf("expected max_retries=json.Number(-1) when nil, got %v", result["max_retries"])
	}
}

func TestAWSAuthEngineClientConfig_toMap_FallbackToSpecAccessKey(t *testing.T) {
	config := &AWSAuthClientConfig{
		AccessKey: "SPEC_ACCESS_KEY",
	}

	result := config.toMap()

	if result["access_key"] != "SPEC_ACCESS_KEY" {
		t.Errorf("expected access_key=SPEC_ACCESS_KEY, got %v", result["access_key"])
	}
	if _, exists := result["secret_key"]; exists {
		t.Errorf("expected secret_key to not be present when not resolved, got %v", result["secret_key"])
	}
}

func TestAWSAuthEngineClientConfig_IsEquivalentToDesiredState_MatchNoCredentials(t *testing.T) {
	maxRetries := 3
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.AWSAuthClientConfig = AWSAuthClientConfig{
		Endpoint:               "https://ec2.example.com",
		IAMEndpoint:            "https://iam.example.com",
		STSEndpoint:            "https://sts.example.com",
		STSRegion:              "us-east-1",
		UseSTSRegionFromClient: false,
		IAMServerIDHeaderValue: "vault.example.com",
		AllowedSTSHeaderValues: "",
		MaxRetries:             &maxRetries,
		retrievedAccessKey:     "AKIAIOSFODNN7EXAMPLE",
	}

	vaultPayload := map[string]any{
		"access_key":                 "AKIAIOSFODNN7EXAMPLE",
		"endpoint":                   "https://ec2.example.com",
		"iam_endpoint":               "https://iam.example.com",
		"sts_endpoint":               "https://sts.example.com",
		"sts_region":                 "us-east-1",
		"use_sts_region_from_client": false,
		"iam_server_id_header_value": "vault.example.com",
		"allowed_sts_header_values":  "",
		"max_retries":                json.Number("3"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when no secret_key is set")
	}
}

func TestAWSAuthEngineClientConfig_IsEquivalentToDesiredState_SecretKeyExcludedFromComparison(t *testing.T) {
	maxRetries := 3
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.AWSAuthClientConfig = AWSAuthClientConfig{
		Endpoint:               "https://ec2.example.com",
		IAMEndpoint:            "https://iam.example.com",
		STSEndpoint:            "https://sts.example.com",
		STSRegion:              "us-east-1",
		UseSTSRegionFromClient: false,
		IAMServerIDHeaderValue: "vault.example.com",
		AllowedSTSHeaderValues: "",
		MaxRetries:             &maxRetries,
		retrievedAccessKey:     "AKIAIOSFODNN7EXAMPLE",
		retrievedSecretKey:     "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	vaultPayload := map[string]any{
		"access_key":                 "AKIAIOSFODNN7EXAMPLE",
		"endpoint":                   "https://ec2.example.com",
		"iam_endpoint":               "https://iam.example.com",
		"sts_endpoint":               "https://sts.example.com",
		"sts_region":                 "us-east-1",
		"use_sts_region_from_client": false,
		"iam_server_id_header_value": "vault.example.com",
		"allowed_sts_header_values":  "",
		"max_retries":                json.Number("3"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: secret_key is write-only and must be excluded from drift comparison (credential rotation is handled by Secret watch predicate)")
	}
}

func TestAWSAuthEngineClientConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	maxRetries := 3
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.AWSAuthClientConfig = AWSAuthClientConfig{
		STSEndpoint:        "https://sts.example.com",
		MaxRetries:         &maxRetries,
		retrievedAccessKey: "AKIAIOSFODNN7EXAMPLE",
		retrievedSecretKey: "SECRET",
	}

	vaultPayload := map[string]any{
		"access_key":                 "AKIAIOSFODNN7EXAMPLE",
		"endpoint":                   "",
		"iam_endpoint":               "",
		"sts_endpoint":               "https://sts.DIFFERENT.com",
		"sts_region":                 "",
		"use_sts_region_from_client": false,
		"iam_server_id_header_value": "",
		"allowed_sts_header_values":  "",
		"max_retries":                json.Number("3"),
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched sts_endpoint")
	}
}

func TestAWSAuthEngineClientConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	maxRetries := -1
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.AWSAuthClientConfig = AWSAuthClientConfig{
		MaxRetries:         &maxRetries,
		retrievedAccessKey: "AKIAIOSFODNN7EXAMPLE",
	}

	vaultPayload := map[string]any{
		"access_key":                 "AKIAIOSFODNN7EXAMPLE",
		"endpoint":                   "",
		"iam_endpoint":               "",
		"sts_endpoint":               "",
		"sts_region":                 "",
		"use_sts_region_from_client": false,
		"iam_server_id_header_value": "",
		"allowed_sts_header_values":  "",
		"max_retries":                json.Number("-1"),
		"request_id":                 "extra-vault-field",
		"lease_duration":             0,
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestAWSAuthEngineClientConfig_GetPath(t *testing.T) {
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.Path = "aws"

	expected := "auth/aws/config/client"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestAWSAuthEngineClientConfig_IsDeletable(t *testing.T) {
	instance := &AWSAuthEngineClientConfig{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true")
	}
}

func TestAWSAuthEngineClientConfig_IsValid_EmptyUsernameKey(t *testing.T) {
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.AWSCredentials.Secret = &corev1.LocalObjectReference{Name: "my-secret"}
	instance.Spec.AWSCredentials.UsernameKey = ""
	instance.Spec.AWSCredentials.PasswordKey = "secret_key"

	err := instance.isValid()
	if err == nil {
		t.Error("expected validation error for empty usernameKey with Secret credential source")
	}
}

func TestAWSAuthEngineClientConfig_IsValid_EmptyPasswordKey(t *testing.T) {
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.AWSCredentials.Secret = &corev1.LocalObjectReference{Name: "my-secret"}
	instance.Spec.AWSCredentials.UsernameKey = "access_key"
	instance.Spec.AWSCredentials.PasswordKey = ""

	err := instance.isValid()
	if err == nil {
		t.Error("expected validation error for empty passwordKey with Secret credential source")
	}
}

func TestAWSAuthEngineClientConfig_IsValid_EmptyKeysWithVaultSecret(t *testing.T) {
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.AWSCredentials.VaultSecret = &vaultutils.VaultSecretReference{Path: "secret/data/aws"}
	instance.Spec.AWSCredentials.UsernameKey = ""
	instance.Spec.AWSCredentials.PasswordKey = "secret_key"

	err := instance.isValid()
	if err == nil {
		t.Error("expected validation error for empty usernameKey with VaultSecret credential source")
	}
}

func TestAWSAuthEngineClientConfig_IsValid_NonEmptyKeysPass(t *testing.T) {
	instance := &AWSAuthEngineClientConfig{}
	instance.Spec.AWSCredentials.Secret = &corev1.LocalObjectReference{Name: "my-secret"}
	instance.Spec.AWSCredentials.UsernameKey = "my_access_key"
	instance.Spec.AWSCredentials.PasswordKey = "my_secret_key"

	err := instance.isValid()
	if err != nil {
		t.Errorf("expected no validation error for non-empty keys, got: %v", err)
	}
}

func TestAWSAuthEngineClientConfig_Default_EmptyKeys(t *testing.T) {
	webhook := &AWSAuthEngineClientConfig{}
	obj := &AWSAuthEngineClientConfig{}
	obj.Spec.AWSCredentials.UsernameKey = ""
	obj.Spec.AWSCredentials.PasswordKey = ""

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.AWSCredentials.UsernameKey != "access_key" {
		t.Errorf("expected usernameKey=access_key, got %s", obj.Spec.AWSCredentials.UsernameKey)
	}
	if obj.Spec.AWSCredentials.PasswordKey != "secret_key" {
		t.Errorf("expected passwordKey=secret_key, got %s", obj.Spec.AWSCredentials.PasswordKey)
	}
}

func TestAWSAuthEngineClientConfig_Default_InheritedDefaults(t *testing.T) {
	webhook := &AWSAuthEngineClientConfig{}
	obj := &AWSAuthEngineClientConfig{}
	obj.Spec.AWSCredentials.UsernameKey = "username"
	obj.Spec.AWSCredentials.PasswordKey = "password"

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.AWSCredentials.UsernameKey != "access_key" {
		t.Errorf("expected usernameKey=access_key when inherited default, got %s", obj.Spec.AWSCredentials.UsernameKey)
	}
	if obj.Spec.AWSCredentials.PasswordKey != "secret_key" {
		t.Errorf("expected passwordKey=secret_key when inherited default, got %s", obj.Spec.AWSCredentials.PasswordKey)
	}
}

func TestAWSAuthEngineClientConfig_Default_PreservesCustomKeys(t *testing.T) {
	webhook := &AWSAuthEngineClientConfig{}
	obj := &AWSAuthEngineClientConfig{}
	obj.Spec.AWSCredentials.UsernameKey = "my_custom_user"
	obj.Spec.AWSCredentials.PasswordKey = "my_custom_pass"

	if err := webhook.Default(context.Background(), obj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Spec.AWSCredentials.UsernameKey != "my_custom_user" {
		t.Errorf("expected custom usernameKey to be preserved, got %s", obj.Spec.AWSCredentials.UsernameKey)
	}
	if obj.Spec.AWSCredentials.PasswordKey != "my_custom_pass" {
		t.Errorf("expected custom passwordKey to be preserved, got %s", obj.Spec.AWSCredentials.PasswordKey)
	}
}
