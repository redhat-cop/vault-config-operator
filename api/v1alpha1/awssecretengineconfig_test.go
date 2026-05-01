package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAWSSecretEngineConfigGetPath(t *testing.T) {
	tests := []struct {
		name         string
		config       *AWSSecretEngineConfig
		expectedPath string
	}{
		{
			name: "basic path",
			config: &AWSSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "aws-config"},
				Spec: AWSSecretEngineConfigSpec{
					Path: "aws",
				},
			},
			expectedPath: vaultutils.CleansePath("aws/config/root"),
		},
		{
			name: "custom path",
			config: &AWSSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "custom-config"},
				Spec: AWSSecretEngineConfigSpec{
					Path: "custom/aws/path",
				},
			},
			expectedPath: vaultutils.CleansePath("custom/aws/path/config/root"),
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

func TestAWSSEConfigToMap(t *testing.T) {
	config := AWSSEConfig{
		MaxRetries:           3,
		Region:               "us-west-2",
		IAMEndpoint:          "https://iam.example.com",
		STSEndpoint:          "https://sts.example.com",
		STSRegion:            "us-west-2",
		STSFallbackEndpoints: []string{"https://sts1.example.com", "https://sts2.example.com"},
		STSFallbackRegions:   []string{"us-east-1", "eu-west-1"},
		UsernameTemplate:     "vault-{{.RoleName}}-{{unix_time}}",
		retrievedAccessKey:   "AKIAIOSFODNN7EXAMPLE",
		retrievedSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	result := config.toMap()

	if result["max_retries"] != 3 {
		t.Errorf("max_retries = %v, expected 3", result["max_retries"])
	}

	if result["region"] != "us-west-2" {
		t.Errorf("region = %v, expected 'us-west-2'", result["region"])
	}

	if result["iam_endpoint"] != "https://iam.example.com" {
		t.Errorf("iam_endpoint = %v", result["iam_endpoint"])
	}

	if result["sts_endpoint"] != "https://sts.example.com" {
		t.Errorf("sts_endpoint = %v", result["sts_endpoint"])
	}

	if result["sts_region"] != "us-west-2" {
		t.Errorf("sts_region = %v", result["sts_region"])
	}

	stsFallbackEndpoints, ok := result["sts_fallback_endpoints"].([]string)
	if !ok {
		t.Fatalf("sts_fallback_endpoints should be []string, got %T", result["sts_fallback_endpoints"])
	}
	if len(stsFallbackEndpoints) != 2 {
		t.Errorf("expected 2 sts_fallback_endpoints, got %d", len(stsFallbackEndpoints))
	}

	stsFallbackRegions, ok := result["sts_fallback_regions"].([]string)
	if !ok {
		t.Fatalf("sts_fallback_regions should be []string, got %T", result["sts_fallback_regions"])
	}
	if len(stsFallbackRegions) != 2 {
		t.Errorf("expected 2 sts_fallback_regions, got %d", len(stsFallbackRegions))
	}

	if result["username_template"] != "vault-{{.RoleName}}-{{unix_time}}" {
		t.Errorf("username_template = %v", result["username_template"])
	}

	if result["access_key"] != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("access_key = %v", result["access_key"])
	}

	if result["secret_key"] != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("secret_key = %v", result["secret_key"])
	}

	// Fields that should not be in the map when not set
	if _, exists := result["identity_token_audience"]; exists {
		t.Error("identity_token_audience should not be in map when empty")
	}
	if _, exists := result["role_arn"]; exists {
		t.Error("role_arn should not be in map when empty")
	}
	if _, exists := result["rotation_period"]; exists {
		t.Error("rotation_period should not be in map when 0")
	}
	if _, exists := result["rotation_schedule"]; exists {
		t.Error("rotation_schedule should not be in map when empty")
	}
}

func TestAWSSEConfigToMapWithRotation(t *testing.T) {
	config := AWSSEConfig{
		MaxRetries:               -1,
		RotationPeriod:           86400,
		RotationWindow:           3600,
		DisableAutomatedRotation: true,
		retrievedAccessKey:       "AKIAIOSFODNN7EXAMPLE",
		retrievedSecretKey:       "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	result := config.toMap()

	if result["max_retries"] != -1 {
		t.Errorf("max_retries = %v, expected -1", result["max_retries"])
	}

	if result["rotation_period"] != 86400 {
		t.Errorf("rotation_period = %v, expected 86400", result["rotation_period"])
	}

	if result["rotation_window"] != 3600 {
		t.Errorf("rotation_window = %v, expected 3600", result["rotation_window"])
	}

	disableRotation, ok := result["disable_automated_rotation"].(bool)
	if !ok {
		t.Fatalf("disable_automated_rotation should be bool, got %T", result["disable_automated_rotation"])
	}
	if !disableRotation {
		t.Errorf("disable_automated_rotation = %v, expected true", disableRotation)
	}
}

func TestAWSSEConfigToMapWithIdentityToken(t *testing.T) {
	config := AWSSEConfig{
		MaxRetries:            -1,
		RoleARN:               "arn:aws:iam::123456789012:role/VaultRole",
		IdentityTokenAudience: "vault.example.com",
		IdentityTokenTTL:      "7200",
	}

	result := config.toMap()

	if result["role_arn"] != "arn:aws:iam::123456789012:role/VaultRole" {
		t.Errorf("role_arn = %v", result["role_arn"])
	}

	if result["identity_token_audience"] != "vault.example.com" {
		t.Errorf("identity_token_audience = %v", result["identity_token_audience"])
	}

	if result["identity_token_ttl"] != "7200" {
		t.Errorf("identity_token_ttl = %v", result["identity_token_ttl"])
	}

	// Should not include access/secret keys when using identity token
	if _, exists := result["access_key"]; exists {
		t.Error("access_key should not be in map when using identity token")
	}
	if _, exists := result["secret_key"]; exists {
		t.Error("secret_key should not be in map when using identity token")
	}
}

func TestAWSSecretEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSSEConfig: AWSSEConfig{
				MaxRetries: 3,
				Region:     "us-west-2",
			},
		},
	}

	payload := config.Spec.AWSSEConfig.toMap()

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAWSSecretEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSSEConfig: AWSSEConfig{
				MaxRetries:     3,
				RotationPeriod: 86400,
			},
		},
	}

	payload := config.Spec.AWSSEConfig.toMap()
	payload["rotation_period"] = 172800

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different rotation_period) to NOT be equivalent")
	}
}

func TestAWSSecretEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSSEConfig: AWSSEConfig{
				MaxRetries: -1,
				Region:     "us-east-1",
			},
		},
	}

	payload := config.Spec.AWSSEConfig.toMap()
	payload["extra_vault_field"] = "some-value"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare reflect.DeepEqual)")
	}
}

func TestAWSSecretEngineConfigIsDeletable(t *testing.T) {
	config := &AWSSecretEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected AWSSecretEngineConfig to be deletable")
	}
}

func TestAWSSecretEngineConfigConditions(t *testing.T) {
	config := &AWSSecretEngineConfig{}

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
