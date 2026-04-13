package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAzureAuthEngineConfigGetPath(t *testing.T) {
	config := &AzureAuthEngineConfig{
		Spec: AzureAuthEngineConfigSpec{
			Path: "azure",
		},
	}

	result := config.GetPath()
	expected := "auth/azure/config"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestAzureConfigToMap(t *testing.T) {
	config := AzureConfig{
		TenantID:      "tenant-123",
		Resource:      "https://management.azure.com/",
		Environment:   "AzurePublicCloud",
		ClientID:      "exported-not-used-in-tomap",
		MaxRetries:    3,
		MaxRetryDelay: 60,
		RetryDelay:    4,
	}
	config.retrievedClientID = "actual-client-id"
	config.retrievedClientPassword = "actual-client-secret"

	result := config.toMap()

	if len(result) != 8 {
		t.Errorf("expected 8 keys in map, got %d", len(result))
	}

	expected := map[string]interface{}{
		"tenant_id":       "tenant-123",
		"resource":        "https://management.azure.com/",
		"environment":     "AzurePublicCloud",
		"client_id":       "actual-client-id",
		"client_secret":   "actual-client-secret",
		"max_retries":     int64(3),
		"max_retry_delay": int64(60),
		"retry_delay":     int64(4),
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestAzureConfigToMapUnexportedCredentials(t *testing.T) {
	config := AzureConfig{
		TenantID: "tenant-123",
		Resource: "https://management.azure.com/",
	}
	config.retrievedClientID = "my-client-id"
	config.retrievedClientPassword = "my-client-secret"

	result := config.toMap()

	if result["client_id"] != "my-client-id" {
		t.Errorf("expected client_id 'my-client-id', got %v", result["client_id"])
	}
	if result["client_secret"] != "my-client-secret" {
		t.Errorf("expected client_secret 'my-client-secret', got %v", result["client_secret"])
	}
}

func TestAzureAuthEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &AzureAuthEngineConfig{
		Spec: AzureAuthEngineConfigSpec{
			AzureConfig: AzureConfig{
				TenantID:    "tenant-123",
				Resource:    "https://management.azure.com/",
				Environment: "AzurePublicCloud",
				MaxRetries:  3,
			},
		},
	}

	payload := config.Spec.AzureConfig.toMap()

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAzureAuthEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &AzureAuthEngineConfig{
		Spec: AzureAuthEngineConfigSpec{
			AzureConfig: AzureConfig{
				TenantID: "tenant-123",
				Resource: "https://management.azure.com/",
			},
		},
	}

	payload := config.Spec.AzureConfig.toMap()
	payload["tenant_id"] = "different-tenant"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different tenant_id) to NOT be equivalent")
	}
}

func TestAzureAuthEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &AzureAuthEngineConfig{
		Spec: AzureAuthEngineConfigSpec{
			AzureConfig: AzureConfig{
				TenantID: "tenant-123",
				Resource: "https://management.azure.com/",
			},
		},
	}

	payload := config.Spec.AzureConfig.toMap()
	payload["extra_field"] = "unexpected"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
	}
}

func TestAzureAuthEngineConfigIsDeletable(t *testing.T) {
	config := &AzureAuthEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected AzureAuthEngineConfig to be deletable")
	}
}

func TestAzureAuthEngineConfigConditions(t *testing.T) {
	config := &AzureAuthEngineConfig{}

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
