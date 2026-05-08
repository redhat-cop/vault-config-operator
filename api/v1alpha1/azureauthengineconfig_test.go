package v1alpha1

import (
	"reflect"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
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

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
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

func TestAzureAuthEngineConfig_PrepareInternalValues_DefaultNoOp(t *testing.T) {
	defaultCreds := vaultutils.RootCredentialConfig{
		PasswordKey: "clientsecret",
		UsernameKey: "clientid",
	}
	config := &AzureAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-azure-auth"},
		Spec: AzureAuthEngineConfigSpec{
			AzureCredentials: defaultCreds,
			AzureConfig: AzureConfig{
				TenantID: "tenant-1",
				Resource: "https://management.azure.com/",
			},
		},
	}
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedClientID != "" || config.Spec.retrievedClientPassword != "" {
		t.Errorf("expected empty retrieved credentials, got clientID=%q clientPassword=%q",
			config.Spec.retrievedClientID, config.Spec.retrievedClientPassword)
	}
}

func TestAzureAuthEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-azure-auth"
	sec := newK8sSecret(ns, "az-creds", map[string][]byte{
		"clientid":     []byte("cid-from-secret"),
		"clientsecret": []byte("pwd-from-k8s"),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &AzureAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: AzureAuthEngineConfigSpec{
			AzureCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "az-creds"},
				UsernameKey: "clientid",
				PasswordKey: "clientsecret",
			},
			AzureConfig: AzureConfig{
				TenantID: "tenant-1",
				Resource: "https://management.azure.com/",
				ClientID: "",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedClientID != "cid-from-secret" {
		t.Errorf("retrievedClientID = %q, want cid-from-secret", config.Spec.retrievedClientID)
	}
	if config.Spec.retrievedClientPassword != "pwd-from-k8s" {
		t.Errorf("retrievedClientPassword = %q, want pwd-from-k8s", config.Spec.retrievedClientPassword)
	}
}
