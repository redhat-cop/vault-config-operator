package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGCPSecretEngineConfig_toMap(t *testing.T) {
	config := GCPSEConfig{
		TTL:                  "1h",
		MaxTTL:               "4h",
		retrievedCredentials: `{"type":"service_account","project_id":"my-project"}`,
	}

	result := config.toMap()

	expectedKeys := []string{"credentials", "ttl", "max_ttl"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys in toMap() output, got %d: %v", len(expectedKeys), len(result), result)
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if creds, ok := result["credentials"].(string); !ok || creds != `{"type":"service_account","project_id":"my-project"}` {
		t.Errorf("credentials = %v, expected JSON string", result["credentials"])
	}
	if ttl, ok := result["ttl"].(string); !ok || ttl != "1h" {
		t.Errorf("ttl = %v, expected 1h", result["ttl"])
	}
	if maxTTL, ok := result["max_ttl"].(string); !ok || maxTTL != "4h" {
		t.Errorf("max_ttl = %v, expected 4h", result["max_ttl"])
	}
}

func TestGCPSecretEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	config := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPSEConfig: GCPSEConfig{
				TTL:                  "1h",
				MaxTTL:               "4h",
				retrievedCredentials: `{"type":"service_account"}`,
			},
		},
	}

	vaultPayload := map[string]any{
		"ttl":     "1h",
		"max_ttl": "4h",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match (credentials excluded from comparison)")
	}
}

func TestGCPSecretEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	config := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPSEConfig: GCPSEConfig{
				TTL:    "1h",
				MaxTTL: "4h",
			},
		},
	}

	vaultPayload := map[string]any{
		"ttl":     "2h",
		"max_ttl": "4h",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when ttl differs")
	}
}

func TestGCPSecretEngineConfig_IsEquivalentToDesiredState_CredentialsInPayload(t *testing.T) {
	config := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPSEConfig: GCPSEConfig{
				TTL:                  "1h",
				MaxTTL:               "4h",
				retrievedCredentials: `{"type":"service_account"}`,
			},
		},
	}

	vaultPayload := map[string]any{
		"ttl":         "1h",
		"max_ttl":     "4h",
		"credentials": "some-unexpected-value",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: credentials in payload are filtered out since they are excluded from desiredState")
	}
}

func TestGCPSecretEngineConfig_IsEquivalentToDesiredState_UnsetFieldsIgnored(t *testing.T) {
	config := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPSEConfig: GCPSEConfig{
				TTL: "1h",
			},
		},
	}

	vaultPayload := map[string]any{
		"ttl": "1h",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: unset fields absent from Vault should not cause false drift")
	}
}

func TestGCPSecretEngineConfig_GetPath(t *testing.T) {
	config := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
		},
	}
	expected := "gcp/config"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestGCPSecretEngineConfig_IsDeletable(t *testing.T) {
	config := &GCPSecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected GCPSecretEngineConfig not to be deletable")
	}
}

func TestGCPSecretEngineConfig_Conditions(t *testing.T) {
	config := &GCPSecretEngineConfig{}

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

func TestGCPSecretEngineConfig_IsValid(t *testing.T) {
	config := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			GCPCredentials: GCPCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "gcp-creds"},
			},
		},
	}
	ok, err := config.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid with credential source, got ok=%v, err=%v", ok, err)
	}
}

func TestGCPSecretEngineConfig_IsValid_NoCredentialSource(t *testing.T) {
	config := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			GCPCredentials: GCPCredentialConfig{},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid with no credential source")
	}
}

func TestGCPSecretEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-gcp"
	sec := newK8sSecret(ns, "gcp-creds", map[string][]byte{
		"credentials": []byte(`{"type":"service_account","project_id":"my-project"}`),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &GCPSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPCredentials: GCPCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "gcp-creds"},
				PasswordKey: "credentials",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.GCPSEConfig.retrievedCredentials != `{"type":"service_account","project_id":"my-project"}` {
		t.Errorf("retrievedCredentials = %q, want GCP JSON", config.Spec.GCPSEConfig.retrievedCredentials)
	}
}

func TestGCPSecretEngineConfig_PrepareInternalValues_FromVaultSecret(t *testing.T) {
	ns := "ns-gcp"
	vaultPath := "secret/data/gcp-creds"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"credentials": `{"type":"service_account","project_id":"vault-project"}`,
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &GCPSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPCredentials: GCPCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				PasswordKey: "credentials",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.GCPSEConfig.retrievedCredentials != `{"type":"service_account","project_id":"vault-project"}` {
		t.Errorf("retrievedCredentials = %q, want vault JSON", config.Spec.GCPSEConfig.retrievedCredentials)
	}
}

func TestGCPSecretEngineConfig_PrepareInternalValues_K8sSecretMissingKey(t *testing.T) {
	ns := "ns-gcp"
	sec := newK8sSecret(ns, "gcp-creds", map[string][]byte{
		"other-key": []byte("something"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &GCPSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPCredentials: GCPCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "gcp-creds"},
				PasswordKey: "credentials",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing credentials key")
	}
}

func TestGCPSecretEngineConfig_Default_SetsCredentialsPasswordKey(t *testing.T) {
	r := &GCPSecretEngineConfig{}
	obj := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPCredentials: GCPCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "gcp-creds"},
			},
		},
	}
	if err := r.Default(nil, obj); err != nil {
		t.Fatalf("Default() error: %v", err)
	}
	if obj.Spec.GCPCredentials.PasswordKey != "credentials" {
		t.Errorf("PasswordKey = %q, want %q", obj.Spec.GCPCredentials.PasswordKey, "credentials")
	}
}

func TestGCPSecretEngineConfig_Default_OverridesPasswordDefault(t *testing.T) {
	r := &GCPSecretEngineConfig{}
	obj := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPCredentials: GCPCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "gcp-creds"},
				PasswordKey: "password",
			},
		},
	}
	if err := r.Default(nil, obj); err != nil {
		t.Fatalf("Default() error: %v", err)
	}
	if obj.Spec.GCPCredentials.PasswordKey != "credentials" {
		t.Errorf("PasswordKey = %q, want %q", obj.Spec.GCPCredentials.PasswordKey, "credentials")
	}
}

func TestGCPSecretEngineConfig_Default_PreservesCustomPasswordKey(t *testing.T) {
	r := &GCPSecretEngineConfig{}
	obj := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPCredentials: GCPCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "gcp-creds"},
				PasswordKey: "my_custom_key",
			},
		},
	}
	if err := r.Default(nil, obj); err != nil {
		t.Fatalf("Default() error: %v", err)
	}
	if obj.Spec.GCPCredentials.PasswordKey != "my_custom_key" {
		t.Errorf("PasswordKey = %q, want %q", obj.Spec.GCPCredentials.PasswordKey, "my_custom_key")
	}
}

func TestGCPSecretEngineConfig_ValidateUpdate_RejectsPathChange(t *testing.T) {
	r := &GCPSecretEngineConfig{}
	oldObj := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			GCPCredentials: GCPCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	newObj := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp-new",
			GCPCredentials: GCPCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.path is changed")
	}
}

func TestGCPSecretEngineConfig_ValidateUpdate_RejectsNameChange(t *testing.T) {
	r := &GCPSecretEngineConfig{}
	oldObj := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			Name: "old-name",
			GCPCredentials: GCPCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	newObj := &GCPSecretEngineConfig{
		Spec: GCPSecretEngineConfigSpec{
			Path: "gcp",
			Name: "new-name",
			GCPCredentials: GCPCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.name is changed")
	}
}
