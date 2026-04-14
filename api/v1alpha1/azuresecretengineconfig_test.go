package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAzureSecretEngineConfigGetPath(t *testing.T) {
	config := &AzureSecretEngineConfig{
		Spec: AzureSecretEngineConfigSpec{
			Path: "azure",
		},
	}

	result := config.GetPath()
	expected := "azure/config"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestAzureSEConfigToMap(t *testing.T) {
	config := AzureSEConfig{
		SubscriptionID:  "sub-123",
		TenantID:        "tenant-456",
		ClientID:        "exported-should-not-map-to-client_id",
		Environment:     "AzurePublicCloud",
		PasswordPolicy:  "my-policy",
		RootPasswordTTL: "182d",
	}
	config.retrievedClientID = "client-789"
	config.retrievedClientPassword = "secret-abc"

	result := config.toMap()

	if len(result) != 7 {
		t.Errorf("expected 7 keys in toMap() output, got %d", len(result))
	}

	expected := map[string]interface{}{
		"subscription_id":   "sub-123",
		"tenant_id":         "tenant-456",
		"client_id":         "client-789",
		"client_secret":     "secret-abc",
		"environment":       "AzurePublicCloud",
		"password_policy":   "my-policy",
		"root_password_ttl": "182d",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestAzureSecretEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &AzureSecretEngineConfig{
		Spec: AzureSecretEngineConfigSpec{
			Path: "azure",
			AzureSEConfig: AzureSEConfig{
				SubscriptionID:  "sub-123",
				TenantID:        "tenant-456",
				Environment:     "AzurePublicCloud",
				PasswordPolicy:  "my-policy",
				RootPasswordTTL: "182d",
			},
		},
	}
	config.Spec.retrievedClientID = "client-789"
	config.Spec.retrievedClientPassword = "secret-abc"

	payload := config.Spec.AzureSEConfig.toMap()

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAzureSecretEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &AzureSecretEngineConfig{
		Spec: AzureSecretEngineConfigSpec{
			Path: "azure",
			AzureSEConfig: AzureSEConfig{
				SubscriptionID:  "sub-123",
				TenantID:        "tenant-456",
				Environment:     "AzurePublicCloud",
				PasswordPolicy:  "my-policy",
				RootPasswordTTL: "182d",
			},
		},
	}
	config.Spec.retrievedClientID = "client-789"
	config.Spec.retrievedClientPassword = "secret-abc"

	payload := config.Spec.AzureSEConfig.toMap()
	payload["tenant_id"] = "other-tenant"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different tenant_id) to NOT be equivalent")
	}
}

func TestAzureSecretEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &AzureSecretEngineConfig{
		Spec: AzureSecretEngineConfigSpec{
			Path: "azure",
			AzureSEConfig: AzureSEConfig{
				SubscriptionID:  "sub-123",
				TenantID:        "tenant-456",
				Environment:     "AzurePublicCloud",
				PasswordPolicy:  "my-policy",
				RootPasswordTTL: "182d",
			},
		},
	}
	config.Spec.retrievedClientID = "client-789"
	config.Spec.retrievedClientPassword = "secret-abc"

	payload := config.Spec.AzureSEConfig.toMap()
	payload["extra_vault_field"] = "some-value"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare reflect.DeepEqual)")
	}
}

func TestAzureSecretEngineConfigIsDeletable(t *testing.T) {
	config := &AzureSecretEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected AzureSecretEngineConfig to be deletable")
	}
}

func TestAzureSecretEngineConfigConditions(t *testing.T) {
	config := &AzureSecretEngineConfig{}

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
