package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGCPSecretEngineRoleset_toMap(t *testing.T) {
	roleset := GCPSERoleset{
		SecretType:  "access_token",
		Project:     "my-gcp-project",
		Bindings:    `resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" { roles = ["roles/viewer"] }`,
		TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
	}

	result := roleset.toMap()

	expectedKeys := []string{"secret_type", "project", "bindings", "token_scopes"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys in toMap() output, got %d: %v", len(expectedKeys), len(result), result)
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if st, ok := result["secret_type"].(string); !ok || st != "access_token" {
		t.Errorf("secret_type = %v, expected access_token", result["secret_type"])
	}
	if proj, ok := result["project"].(string); !ok || proj != "my-gcp-project" {
		t.Errorf("project = %v, expected my-gcp-project", result["project"])
	}
	if b, ok := result["bindings"].(string); !ok || b == "" {
		t.Errorf("bindings = %v, expected non-empty string", result["bindings"])
	}
	scopes, ok := result["token_scopes"].([]any)
	if !ok || len(scopes) != 1 {
		t.Fatalf("token_scopes = %v (%T), expected []any with 1 element", result["token_scopes"], result["token_scopes"])
	}
	if scopes[0] != "https://www.googleapis.com/auth/cloud-platform" {
		t.Errorf("token_scopes[0] = %v, expected cloud-platform scope", scopes[0])
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_Match(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				Bindings:    `resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" { roles = ["roles/viewer"] }`,
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":             "access_token",
		"service_account_email":   "vault-myroleset-123@my-gcp-project.iam.gserviceaccount.com",
		"service_account_project": "my-gcp-project",
		"bindings":                `resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" { roles = ["roles/viewer"] }`,
		"token_scopes":            []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match (project excluded, bindings match)")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_BindingsHCLParsed(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				Bindings:    `resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" { roles = ["roles/viewer"] }`,
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"bindings": map[string]any{
			"//cloudresourcemanager.googleapis.com/projects/my-gcp-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: HCL bindings should parse to equivalent map")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_BindingsHCLDiffers(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				Bindings:    `resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" { roles = ["roles/editor"] }`,
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"bindings": map[string]any{
			"//cloudresourcemanager.googleapis.com/projects/my-gcp-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: parsed HCL bindings differ from Vault map")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_BindingsUnparseableFallback(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				Bindings:    `not valid json { and not valid hcl either !!!`,
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"bindings": map[string]any{
			"project/my-gcp-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: unparseable bindings should be excluded from comparison (graceful fallback)")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_BindingsJSONMatchesMap(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				Bindings:    `{"project/my-gcp-project":["roles/viewer"]}`,
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"bindings": map[string]any{
			"project/my-gcp-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: JSON bindings string should parse to equivalent map")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_BindingsJSONDiffers(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				Bindings:    `{"project/my-gcp-project":["roles/editor"]}`,
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"bindings": map[string]any{
			"project/my-gcp-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: parsed JSON bindings differ from Vault map")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_BindingsRoleOrderInsensitive(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				Bindings:    `{"//cloudresourcemanager.googleapis.com/projects/my-gcp-project":["roles/viewer","roles/editor"]}`,
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"bindings": map[string]any{
			"//cloudresourcemanager.googleapis.com/projects/my-gcp-project": []any{"roles/editor", "roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: bindings with roles in different order should be equivalent")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_BindingsHCLRoleOrderInsensitive(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				Bindings:    `resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" { roles = ["roles/viewer", "roles/editor"] }`,
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"bindings": map[string]any{
			"//cloudresourcemanager.googleapis.com/projects/my-gcp-project": []any{"roles/editor", "roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: HCL bindings with roles in different order should be equivalent")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_BindingsChanged(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "my-gcp-project",
				Bindings:   "new-bindings",
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"bindings":    "old-bindings",
	}

	if roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when bindings differ")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_TokenScopesReordered(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "my-gcp-project",
				TokenScopes: []string{
					"https://www.googleapis.com/auth/cloud-platform",
					"https://www.googleapis.com/auth/compute",
				},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type": "access_token",
		"token_scopes": []any{
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/cloud-platform",
		},
	}

	if !roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: token_scopes order should not matter (set semantics)")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":  "service_account_key",
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when secret_type differs")
	}
}

func TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType:  "access_token",
				Project:     "my-gcp-project",
				TokenScopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":             "access_token",
		"service_account_email":   "vault-myroleset-123@my-gcp-project.iam.gserviceaccount.com",
		"service_account_project": "my-gcp-project",
		"token_scopes":            []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !roleset.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: extra Vault-added fields are filtered from payload")
	}
}

func TestGCPSecretEngineRoleset_GetPath(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		ObjectMeta: metav1.ObjectMeta{Name: "my-roleset"},
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
		},
	}
	expected := "gcp/roleset/my-roleset"
	if got := roleset.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestGCPSecretEngineRoleset_GetPath_WithNameOverride(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{
		ObjectMeta: metav1.ObjectMeta{Name: "k8s-name"},
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			Name: "vault-name",
		},
	}
	expected := "gcp/roleset/vault-name"
	if got := roleset.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestGCPSecretEngineRoleset_IsDeletable(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{}
	if !roleset.IsDeletable() {
		t.Error("expected GCPSecretEngineRoleset to be deletable")
	}
}

func TestGCPSecretEngineRoleset_Conditions(t *testing.T) {
	roleset := &GCPSecretEngineRoleset{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	roleset.SetConditions(conditions)
	got := roleset.GetConditions()

	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected condition type 'ReconcileSuccessful', got %v", got[0].Type)
	}
}

func TestGCPSecretEngineRoleset_ValidateUpdate_RejectsPathChange(t *testing.T) {
	r := &GCPSecretEngineRoleset{}
	oldObj := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "my-project",
			},
		},
	}
	newObj := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp-new",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "my-project",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.path is changed")
	}
}

func TestGCPSecretEngineRoleset_ValidateUpdate_RejectsSecretTypeChange(t *testing.T) {
	r := &GCPSecretEngineRoleset{}
	oldObj := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "my-project",
			},
		},
	}
	newObj := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType: "service_account_key",
				Project:    "my-project",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.secretType is changed")
	}
}

func TestGCPSecretEngineRoleset_ValidateUpdate_RejectsProjectChange(t *testing.T) {
	r := &GCPSecretEngineRoleset{}
	oldObj := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "old-project",
			},
		},
	}
	newObj := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "new-project",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.project is changed")
	}
}

func TestGCPSecretEngineRoleset_ValidateUpdate_AllowsBindingsChange(t *testing.T) {
	r := &GCPSecretEngineRoleset{}
	oldObj := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			Name: "my-roleset",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "my-project",
				Bindings:   "old-bindings",
			},
		},
	}
	newObj := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Path: "gcp",
			Name: "my-roleset",
			GCPSERoleset: GCPSERoleset{
				SecretType: "access_token",
				Project:    "my-project",
				Bindings:   "new-bindings",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err != nil {
		t.Errorf("expected no error when only bindings change, got: %v", err)
	}
}

func TestGCPSecretEngineRoleset_GetVaultConnection(t *testing.T) {
	conn := &vaultutils.VaultConnection{Address: "http://vault:8200"}
	roleset := &GCPSecretEngineRoleset{
		Spec: GCPSecretEngineRolesetSpec{
			Connection: conn,
		},
	}
	if roleset.GetVaultConnection() != conn {
		t.Error("expected GetVaultConnection to return the spec connection")
	}
}
