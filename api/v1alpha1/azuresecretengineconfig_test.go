package v1alpha1

import (
	"reflect"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
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

func TestAzureSecretEngineConfig_PrepareInternalValues_DefaultNoOp(t *testing.T) {
	defaultCreds := vaultutils.RootCredentialConfig{
		PasswordKey: "clientsecret",
		UsernameKey: "clientid",
	}
	config := &AzureSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-azure-se"},
		Spec: AzureSecretEngineConfigSpec{
			AzureCredentials: defaultCreds,
			AzureSEConfig: AzureSEConfig{
				SubscriptionID: "sub-1",
				TenantID:       "tenant-1",
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

func TestAzureSecretEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-azure-se"
	sec := newK8sSecret(ns, "az-se-creds", map[string][]byte{
		"clientid":     []byte("se-cid"),
		"clientsecret": []byte("se-secret"),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &AzureSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: AzureSecretEngineConfigSpec{
			AzureCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "az-se-creds"},
				UsernameKey: "clientid",
				PasswordKey: "clientsecret",
			},
			AzureSEConfig: AzureSEConfig{
				SubscriptionID: "sub-1",
				TenantID:       "tenant-1",
				ClientID:       "",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedClientID != "se-cid" {
		t.Errorf("retrievedClientID = %q, want se-cid", config.Spec.retrievedClientID)
	}
	if config.Spec.retrievedClientPassword != "se-secret" {
		t.Errorf("retrievedClientPassword = %q, want se-secret", config.Spec.retrievedClientPassword)
	}
}
