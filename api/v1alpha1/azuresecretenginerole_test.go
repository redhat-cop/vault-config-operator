package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAzureSecretEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *AzureSecretEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &AzureSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AzureSecretEngineRoleSpec{
					Path: "azure",
					Name: "custom-name",
				},
			},
			expectedPath: vaultutils.CleansePath("azure/roles/custom-name"),
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &AzureSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AzureSecretEngineRoleSpec{
					Path: "azure",
				},
			},
			expectedPath: vaultutils.CleansePath("azure/roles/meta-name"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestAzureSERoleToMap(t *testing.T) {
	role := AzureSERole{
		AzureRoles:          `[{"role_name": "Contributor"}]`,
		AzureGroups:         `[{"group_name": "mygroup"}]`,
		ApplicationObjectID: "app-obj-123",
		PersistApp:          true,
		TTL:                 "1h",
		MaxTTL:              "24h",
		PermanentlyDelete:   "true",
		SignInAudience:      "AzureADMyOrg",
		Tags:                "team:dev",
	}

	result := role.toMap()

	expectedKeys := []string{
		"azure_roles", "azure_groups", "application_object_id", "persist_app",
		"ttl", "max_ttl", "permanently_delete", "sign_in_audience", "tags",
	}
	if len(result) != 9 {
		t.Errorf("expected 9 keys in toMap() output, got %d", len(result))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if result["azure_roles"] != `[{"role_name": "Contributor"}]` {
		t.Errorf("azure_roles = %v", result["azure_roles"])
	}
	if result["azure_groups"] != `[{"group_name": "mygroup"}]` {
		t.Errorf("azure_groups = %v", result["azure_groups"])
	}
	if result["application_object_id"] != "app-obj-123" {
		t.Errorf("application_object_id = %v", result["application_object_id"])
	}
	persist, ok := result["persist_app"].(bool)
	if !ok {
		t.Fatalf("persist_app should be bool, got %T", result["persist_app"])
	}
	if !persist {
		t.Errorf("persist_app = %v, expected true", persist)
	}
	if result["ttl"] != "1h" {
		t.Errorf("ttl = %v", result["ttl"])
	}
	if result["max_ttl"] != "24h" {
		t.Errorf("max_ttl = %v", result["max_ttl"])
	}
	if result["permanently_delete"] != "true" {
		t.Errorf("permanently_delete = %v", result["permanently_delete"])
	}
	if result["sign_in_audience"] != "AzureADMyOrg" {
		t.Errorf("sign_in_audience = %v", result["sign_in_audience"])
	}
	if result["tags"] != "team:dev" {
		t.Errorf("tags = %v", result["tags"])
	}
}

func TestAzureSecretEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &AzureSecretEngineRole{
		Spec: AzureSecretEngineRoleSpec{
			Path: "azure",
			AzureSERole: AzureSERole{
				AzureRoles:          `[{"role_name": "Contributor"}]`,
				AzureGroups:         `[{"group_name": "mygroup"}]`,
				ApplicationObjectID: "app-obj-123",
				PersistApp:          true,
				TTL:                 "1h",
				MaxTTL:              "24h",
				PermanentlyDelete:   "true",
				SignInAudience:      "AzureADMyOrg",
				Tags:                "team:dev",
			},
		},
	}

	payload := role.Spec.AzureSERole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAzureSecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &AzureSecretEngineRole{
		Spec: AzureSecretEngineRoleSpec{
			Path: "azure",
			AzureSERole: AzureSERole{
				AzureRoles:          `[{"role_name": "Contributor"}]`,
				ApplicationObjectID: "app-obj-123",
				PersistApp:          true,
			},
		},
	}

	payload := role.Spec.AzureSERole.toMap()
	payload["ttl"] = "2h"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different ttl) to NOT be equivalent")
	}
}

func TestAzureSecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &AzureSecretEngineRole{
		Spec: AzureSecretEngineRoleSpec{
			Path: "azure",
			AzureSERole: AzureSERole{
				AzureRoles:          `[{"role_name": "Contributor"}]`,
				ApplicationObjectID: "app-obj-123",
				PersistApp:          true,
			},
		},
	}

	payload := role.Spec.AzureSERole.toMap()
	payload["extra_vault_field"] = "some-value"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare reflect.DeepEqual)")
	}
}

func TestAzureSecretEngineRoleIsDeletable(t *testing.T) {
	role := &AzureSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected AzureSecretEngineRole to be deletable")
	}
}

func TestAzureSecretEngineRoleConditions(t *testing.T) {
	role := &AzureSecretEngineRole{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	role.SetConditions(conditions)
	got := role.GetConditions()

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
