package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTerraformCloudSecretEngineConfig_toMap(t *testing.T) {
	config := TFCSEConfig{
		Address:        "https://app.terraform.io",
		retrievedToken: "my-tfc-token",
	}

	result := config.toMap()

	expectedKeys := []string{"address", "token"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys in toMap() output, got %d: %v", len(expectedKeys), len(result), result)
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if addr, ok := result["address"].(string); !ok || addr != "https://app.terraform.io" {
		t.Errorf("address = %v, expected https://app.terraform.io", result["address"])
	}
	if token, ok := result["token"].(string); !ok || token != "my-tfc-token" {
		t.Errorf("token = %v, expected my-tfc-token", result["token"])
	}
}

func TestTerraformCloudSecretEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCSEConfig: TFCSEConfig{
				Address:        "https://app.terraform.io",
				retrievedToken: "my-tfc-token",
			},
		},
	}

	vaultPayload := map[string]any{
		"address":   "https://app.terraform.io",
		"base_path": "/api/v2/",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match (token excluded, base_path filtered)")
	}
}

func TestTerraformCloudSecretEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCSEConfig: TFCSEConfig{
				Address: "https://app.terraform.io",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "https://tfe.example.com",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when address differs")
	}
}

func TestTerraformCloudSecretEngineConfig_IsEquivalentToDesiredState_TokenInPayload(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCSEConfig: TFCSEConfig{
				Address:        "https://app.terraform.io",
				retrievedToken: "my-tfc-token",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "https://app.terraform.io",
		"token":   "some-unexpected-value",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: token in payload is filtered out since it is excluded from desiredState")
	}
}

func TestTerraformCloudSecretEngineConfig_GetPath(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
		},
	}
	expected := "terraform/config"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestTerraformCloudSecretEngineConfig_IsDeletable(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected TerraformCloudSecretEngineConfig not to be deletable")
	}
}

func TestTerraformCloudSecretEngineConfig_Conditions(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{}

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
}

func TestTerraformCloudSecretEngineConfig_IsValid(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			TFCCredentials: TFCCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "tfc-token"},
				PasswordKey: "token",
			},
		},
	}
	ok, err := config.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid with credential source, got ok=%v, err=%v", ok, err)
	}
}

func TestTerraformCloudSecretEngineConfig_IsValid_NoCredentialSource(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			TFCCredentials: TFCCredentialConfig{},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid with no credential source")
	}
}

func TestTerraformCloudSecretEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-tfc"
	sec := newK8sSecret(ns, "tfc-token", map[string][]byte{
		"token": []byte("my-tfc-api-token"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &TerraformCloudSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "tfc-token"},
				PasswordKey: "token",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.TFCSEConfig.retrievedToken != "my-tfc-api-token" {
		t.Errorf("retrievedToken = %q, want my-tfc-api-token", config.Spec.TFCSEConfig.retrievedToken)
	}
}

func TestTerraformCloudSecretEngineConfig_PrepareInternalValues_FromVaultSecret(t *testing.T) {
	ns := "ns-tfc"
	vaultPath := "secret/data/tfc-creds"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"token": "vault-tfc-token",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &TerraformCloudSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				PasswordKey: "token",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.TFCSEConfig.retrievedToken != "vault-tfc-token" {
		t.Errorf("retrievedToken = %q, want vault-tfc-token", config.Spec.TFCSEConfig.retrievedToken)
	}
}

func TestTerraformCloudSecretEngineConfig_PrepareInternalValues_K8sSecretMissingKey(t *testing.T) {
	ns := "ns-tfc"
	sec := newK8sSecret(ns, "tfc-token", map[string][]byte{
		"other-key": []byte("something"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &TerraformCloudSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "tfc-token"},
				PasswordKey: "token",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing token key")
	}
}

func TestTerraformCloudSecretEngineConfig_Default_SetsTokenPasswordKey(t *testing.T) {
	r := &TerraformCloudSecretEngineConfig{}
	obj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "tfc-token"},
			},
		},
	}
	if err := r.Default(nil, obj); err != nil {
		t.Fatalf("Default() error: %v", err)
	}
	if obj.Spec.TFCCredentials.PasswordKey != "token" {
		t.Errorf("PasswordKey = %q, want %q", obj.Spec.TFCCredentials.PasswordKey, "token")
	}
}

func TestTerraformCloudSecretEngineConfig_Default_OverridesPasswordDefault(t *testing.T) {
	r := &TerraformCloudSecretEngineConfig{}
	obj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "tfc-token"},
				PasswordKey: "password",
			},
		},
	}
	if err := r.Default(nil, obj); err != nil {
		t.Fatalf("Default() error: %v", err)
	}
	if obj.Spec.TFCCredentials.PasswordKey != "token" {
		t.Errorf("PasswordKey = %q, want %q", obj.Spec.TFCCredentials.PasswordKey, "token")
	}
}

func TestTerraformCloudSecretEngineConfig_Default_PreservesCustomPasswordKey(t *testing.T) {
	r := &TerraformCloudSecretEngineConfig{}
	obj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "tfc-token"},
				PasswordKey: "my_custom_key",
			},
		},
	}
	if err := r.Default(nil, obj); err != nil {
		t.Fatalf("Default() error: %v", err)
	}
	if obj.Spec.TFCCredentials.PasswordKey != "my_custom_key" {
		t.Errorf("PasswordKey = %q, want %q", obj.Spec.TFCCredentials.PasswordKey, "my_custom_key")
	}
}

func TestTerraformCloudSecretEngineConfig_ValidateUpdate_RejectsPathChange(t *testing.T) {
	r := &TerraformCloudSecretEngineConfig{}
	oldObj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	newObj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform-new",
			TFCCredentials: TFCCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.path is changed")
	}
}

func TestTerraformCloudSecretEngineConfig_ValidateUpdate_RejectsNameChange(t *testing.T) {
	r := &TerraformCloudSecretEngineConfig{}
	oldObj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			Name: "old-name",
			TFCCredentials: TFCCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	newObj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			Name: "new-name",
			TFCCredentials: TFCCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.name is changed")
	}
}

func TestTFCCredentialConfig_ValidateCredentialSource_EmptySecretName(t *testing.T) {
	cred := TFCCredentialConfig{
		Secret: &corev1.LocalObjectReference{Name: ""},
	}
	err := cred.ValidateCredentialSource()
	if err == nil {
		t.Error("expected error when secret.name is empty")
	}
}

func TestTFCCredentialConfig_ValidateCredentialSource_EmptyRandomSecretName(t *testing.T) {
	cred := TFCCredentialConfig{
		RandomSecret: &corev1.LocalObjectReference{Name: ""},
	}
	err := cred.ValidateCredentialSource()
	if err == nil {
		t.Error("expected error when randomSecret.name is empty")
	}
}

func TestTFCCredentialConfig_ValidateCredentialSource_EmptyVaultSecretPath(t *testing.T) {
	cred := TFCCredentialConfig{
		VaultSecret: &vaultutils.VaultSecretReference{Path: ""},
	}
	err := cred.ValidateCredentialSource()
	if err == nil {
		t.Error("expected error when vaultSecret.path is empty")
	}
}

func TestTFCCredentialConfig_ValidateCredentialSource_ValidSecretName(t *testing.T) {
	cred := TFCCredentialConfig{
		Secret: &corev1.LocalObjectReference{Name: "my-secret"},
	}
	err := cred.ValidateCredentialSource()
	if err != nil {
		t.Errorf("unexpected error for valid secret name: %v", err)
	}
}

func TestTerraformCloudSecretEngineConfig_IsValid_EmptyPasswordKeyWithSecret(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			TFCCredentials: TFCCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "tfc-token"},
				PasswordKey: "",
			},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid when passwordKey is empty with a secret credential source")
	}
}

func TestTerraformCloudSecretEngineConfig_IsValid_EmptyPasswordKeyWithVaultSecret(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			TFCCredentials: TFCCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: "secret/tfc"},
				PasswordKey: "",
			},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid when passwordKey is empty with a vaultSecret credential source")
	}
}

func TestTerraformCloudSecretEngineConfig_IsValid_EmptyPasswordKeyWithRandomSecret(t *testing.T) {
	config := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			TFCCredentials: TFCCredentialConfig{
				RandomSecret: &corev1.LocalObjectReference{Name: "random-cred"},
				PasswordKey:  "",
			},
		},
	}
	ok, err := config.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid when passwordKey is empty with randomSecret (does not use passwordKey), got ok=%v err=%v", ok, err)
	}
}

func TestTerraformCloudSecretEngineConfig_ValidateUpdate_RejectsEmptyPasswordKey(t *testing.T) {
	r := &TerraformCloudSecretEngineConfig{}
	oldObj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "cred"},
				PasswordKey: "token",
			},
		},
	}
	newObj := &TerraformCloudSecretEngineConfig{
		Spec: TerraformCloudSecretEngineConfigSpec{
			Path: "terraform",
			TFCCredentials: TFCCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "cred"},
				PasswordKey: "",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when passwordKey is blanked on update")
	}
}
