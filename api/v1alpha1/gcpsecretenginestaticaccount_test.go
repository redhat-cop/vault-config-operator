package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGCPSecretEngineStaticAccount_toMap(t *testing.T) {
	sa := GCPSEStaticAccount{
		SecretType:          "access_token",
		ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
		Bindings:            `resource "//cloudresourcemanager.googleapis.com/projects/my-project" { roles = ["roles/viewer"] }`,
		TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
	}

	result := sa.toMap()

	expectedKeys := []string{"secret_type", "service_account_email", "bindings", "token_scopes"}
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
	if email, ok := result["service_account_email"].(string); !ok || email != "example@my-project.iam.gserviceaccount.com" {
		t.Errorf("service_account_email = %v, expected example@my-project.iam.gserviceaccount.com", result["service_account_email"])
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

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_Match(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				Bindings:            `resource "//cloudresourcemanager.googleapis.com/projects/my-project" { roles = ["roles/viewer"] }`,
				TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":             "access_token",
		"service_account_email":   "example@my-project.iam.gserviceaccount.com",
		"service_account_project": "my-project",
		"bindings":                `resource "//cloudresourcemanager.googleapis.com/projects/my-project" { roles = ["roles/viewer"] }`,
		"token_scopes":            []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match (bindings match)")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_BindingsHCLParsed(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				Bindings:            `resource "//cloudresourcemanager.googleapis.com/projects/my-project" { roles = ["roles/viewer"] }`,
				TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "example@my-project.iam.gserviceaccount.com",
		"bindings": map[string]any{
			"//cloudresourcemanager.googleapis.com/projects/my-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: HCL bindings should parse to equivalent map")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_BindingsHCLDiffers(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				Bindings:            `resource "//cloudresourcemanager.googleapis.com/projects/my-project" { roles = ["roles/editor"] }`,
				TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "example@my-project.iam.gserviceaccount.com",
		"bindings": map[string]any{
			"//cloudresourcemanager.googleapis.com/projects/my-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: parsed HCL bindings differ from Vault map")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_BindingsUnparseableFallback(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				Bindings:            `not valid json { and not valid hcl either !!!`,
				TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "example@my-project.iam.gserviceaccount.com",
		"bindings": map[string]any{
			"project/my-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: unparseable bindings should be excluded from comparison (graceful fallback)")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_BindingsJSONMatchesMap(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				Bindings:            `{"project/my-project":["roles/viewer"]}`,
				TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "example@my-project.iam.gserviceaccount.com",
		"bindings": map[string]any{
			"project/my-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: JSON bindings string should parse to equivalent map")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_BindingsJSONDiffers(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				Bindings:            `{"project/my-project":["roles/editor"]}`,
				TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "example@my-project.iam.gserviceaccount.com",
		"bindings": map[string]any{
			"project/my-project": []any{"roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: parsed JSON bindings differ from Vault map")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_BindingsRoleOrderInsensitive(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				Bindings:            `{"//cloudresourcemanager.googleapis.com/projects/my-project":["roles/viewer","roles/editor"]}`,
				TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "example@my-project.iam.gserviceaccount.com",
		"bindings": map[string]any{
			"//cloudresourcemanager.googleapis.com/projects/my-project": []any{"roles/editor", "roles/viewer"},
		},
		"token_scopes": []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if !sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: bindings with roles in different order should be equivalent")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_BindingsChanged(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				Bindings:            "new-bindings",
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "example@my-project.iam.gserviceaccount.com",
		"bindings":              "old-bindings",
	}

	if sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when bindings differ")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_TokenScopesReordered(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				TokenScopes: []string{
					"https://www.googleapis.com/auth/cloud-platform",
					"https://www.googleapis.com/auth/compute",
				},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "example@my-project.iam.gserviceaccount.com",
		"token_scopes": []any{
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/cloud-platform",
		},
	}

	if !sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: token_scopes order should not matter (set semantics)")
	}
}

func TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "example@my-project.iam.gserviceaccount.com",
				TokenScopes:         []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
	}

	vaultPayload := map[string]any{
		"secret_type":           "access_token",
		"service_account_email": "different@my-project.iam.gserviceaccount.com",
		"token_scopes":          []any{"https://www.googleapis.com/auth/cloud-platform"},
	}

	if sa.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when service_account_email differs")
	}
}

func TestGCPSecretEngineStaticAccount_GetPath(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "my-sa"},
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
		},
	}
	expected := "gcp/static-account/my-sa"
	if got := sa.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestGCPSecretEngineStaticAccount_GetPath_WithNameOverride(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{
		ObjectMeta: metav1.ObjectMeta{Name: "k8s-name"},
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			Name: "vault-name",
		},
	}
	expected := "gcp/static-account/vault-name"
	if got := sa.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestGCPSecretEngineStaticAccount_IsDeletable(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{}
	if !sa.IsDeletable() {
		t.Error("expected GCPSecretEngineStaticAccount to be deletable")
	}
}

func TestGCPSecretEngineStaticAccount_Conditions(t *testing.T) {
	sa := &GCPSecretEngineStaticAccount{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	sa.SetConditions(conditions)
	got := sa.GetConditions()

	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected condition type 'ReconcileSuccessful', got %v", got[0].Type)
	}
}

func TestGCPSecretEngineStaticAccount_ValidateUpdate_RejectsPathChange(t *testing.T) {
	r := &GCPSecretEngineStaticAccount{}
	oldObj := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "sa@project.iam.gserviceaccount.com",
			},
		},
	}
	newObj := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp-new",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "sa@project.iam.gserviceaccount.com",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.path is changed")
	}
}

func TestGCPSecretEngineStaticAccount_ValidateUpdate_RejectsSecretTypeChange(t *testing.T) {
	r := &GCPSecretEngineStaticAccount{}
	oldObj := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "sa@project.iam.gserviceaccount.com",
			},
		},
	}
	newObj := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "service_account_key",
				ServiceAccountEmail: "sa@project.iam.gserviceaccount.com",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.secretType is changed")
	}
}

func TestGCPSecretEngineStaticAccount_ValidateUpdate_RejectsEmailChange(t *testing.T) {
	r := &GCPSecretEngineStaticAccount{}
	oldObj := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "old-sa@project.iam.gserviceaccount.com",
			},
		},
	}
	newObj := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "new-sa@project.iam.gserviceaccount.com",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.serviceAccountEmail is changed")
	}
}

func TestGCPSecretEngineStaticAccount_ValidateUpdate_AllowsBindingsChange(t *testing.T) {
	r := &GCPSecretEngineStaticAccount{}
	oldObj := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			Name: "my-sa",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "sa@project.iam.gserviceaccount.com",
				Bindings:            "old-bindings",
			},
		},
	}
	newObj := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Path: "gcp",
			Name: "my-sa",
			GCPSEStaticAccount: GCPSEStaticAccount{
				SecretType:          "access_token",
				ServiceAccountEmail: "sa@project.iam.gserviceaccount.com",
				Bindings:            "new-bindings",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err != nil {
		t.Errorf("expected no error when only bindings change, got: %v", err)
	}
}

func TestGCPSecretEngineStaticAccount_GetVaultConnection(t *testing.T) {
	conn := &vaultutils.VaultConnection{Address: "http://vault:8200"}
	sa := &GCPSecretEngineStaticAccount{
		Spec: GCPSecretEngineStaticAccountSpec{
			Connection: conn,
		},
	}
	if sa.GetVaultConnection() != conn {
		t.Error("expected GetVaultConnection to return the spec connection")
	}
}
