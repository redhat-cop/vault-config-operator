package v1alpha1

import (
	"encoding/json"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNomadSecretEngineConfigGetPath(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
		},
	}
	expected := "nomad/config/access"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestNomadSecretEngineConfigGetPathWithDifferentMount(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			Path: "my-nomad-engine",
		},
	}
	expected := "my-nomad-engine/config/access"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestNomadSecretEngineConfigIsDeletable(t *testing.T) {
	config := &NomadSecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected NomadSecretEngineConfig not to be deletable")
	}
}

func TestNomadSEConfig_toMap(t *testing.T) {
	config := NomadSEConfig{
		Address:            "http://127.0.0.1:4646",
		MaxTokenNameLength: 128,
		CACert:             "-----BEGIN CERTIFICATE-----\nMIIB...",
		ClientCert:         "-----BEGIN CERTIFICATE-----\nMIIC...",
		ClientKey:          "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
		retrievedToken:     "my-nomad-token",
	}

	result := config.toMap()

	expectedKeys := []string{"address", "max_token_name_length", "token", "ca_cert", "client_cert", "client_key"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys in toMap() output, got %d: %v", len(expectedKeys), len(result), result)
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if addr, ok := result["address"].(string); !ok || addr != "http://127.0.0.1:4646" {
		t.Errorf("address = %v, expected http://127.0.0.1:4646", result["address"])
	}
	if token, ok := result["token"].(string); !ok || token != "my-nomad-token" {
		t.Errorf("token = %v, expected my-nomad-token", result["token"])
	}
	if maxLen, ok := result["max_token_name_length"].(json.Number); !ok || maxLen != json.Number("128") {
		t.Errorf("max_token_name_length = %v (%T), expected json.Number(128)", result["max_token_name_length"], result["max_token_name_length"])
	}
	if ca, ok := result["ca_cert"].(string); !ok || ca != "-----BEGIN CERTIFICATE-----\nMIIB..." {
		t.Errorf("ca_cert = %v, expected PEM cert", result["ca_cert"])
	}
}

func TestNomadSEConfig_toMap_NoToken(t *testing.T) {
	config := NomadSEConfig{
		Address: "http://127.0.0.1:4646",
	}

	result := config.toMap()

	if _, ok := result["token"]; ok {
		t.Error("expected token to be absent when retrievedToken is empty")
	}
	if result["address"] != "http://127.0.0.1:4646" {
		t.Errorf("address = %v, expected http://127.0.0.1:4646", result["address"])
	}
}

func TestNomadSEConfig_toMap_NoMaxTokenNameLength(t *testing.T) {
	config := NomadSEConfig{
		Address:        "http://127.0.0.1:4646",
		retrievedToken: "tok",
	}

	result := config.toMap()

	if _, ok := result["max_token_name_length"]; ok {
		t.Error("expected max_token_name_length to be absent when zero")
	}
}

func TestNomadSecretEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			NomadSEConfig: NomadSEConfig{
				Address:        "http://127.0.0.1:4646",
				retrievedToken: "my-token",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "http://127.0.0.1:4646",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match (token/certs stripped from comparison)")
	}
}

func TestNomadSecretEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			NomadSEConfig: NomadSEConfig{
				Address: "http://127.0.0.1:4646",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "http://10.0.0.1:4646",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when address differs")
	}
}

func TestNomadSecretEngineConfig_IsEquivalentToDesiredState_TokenInPayload(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			NomadSEConfig: NomadSEConfig{
				Address:        "http://127.0.0.1:4646",
				retrievedToken: "my-token",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "http://127.0.0.1:4646",
		"token":   "some-unexpected-token",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: token is deleted from desiredState before comparison, so it's filtered as extra field")
	}
}

func TestNomadSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			NomadSEConfig: NomadSEConfig{
				Address: "http://127.0.0.1:4646",
			},
		},
	}

	vaultPayload := map[string]any{
		"address":     "http://127.0.0.1:4646",
		"extra_field": "from-vault",
		"another":     123,
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestNomadSecretEngineConfigConditions(t *testing.T) {
	config := &NomadSecretEngineConfig{}

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

func TestNomadSecretEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-nomad"
	sec := newK8sSecret(ns, "nomad-creds", map[string][]byte{
		"token": []byte("my-nomad-mgmt-token"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &NomadSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			NomadSEConfig: NomadSEConfig{
				Address: "http://127.0.0.1:4646",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "nomad-creds"},
				PasswordKey: "token",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedToken != "my-nomad-mgmt-token" {
		t.Errorf("retrievedToken = %q, want %q", config.Spec.retrievedToken, "my-nomad-mgmt-token")
	}
}

func TestNomadSecretEngineConfig_PrepareInternalValues_FromVaultSecret(t *testing.T) {
	ns := "ns-nomad"
	vaultPath := "secret/data/nomad-token"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"token": "vault-stored-nomad-token",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &NomadSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			NomadSEConfig: NomadSEConfig{
				Address: "http://127.0.0.1:4646",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				PasswordKey: "token",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedToken != "vault-stored-nomad-token" {
		t.Errorf("retrievedToken = %q, want %q", config.Spec.retrievedToken, "vault-stored-nomad-token")
	}
}

func TestNomadSecretEngineConfig_PrepareInternalValues_FromRandomSecret(t *testing.T) {
	ns := "ns-nomad"
	randomSecretPath := "secret/data/random-nomad"
	handler := newFakeVaultHandler()
	handler.setGet(randomSecretPath+"/my-random-secret", map[string]any{
		"data": map[string]any{
			"password": "random-generated-token",
		},
		"metadata": map[string]any{
			"version": 1,
		},
	})
	randomSecret := &RandomSecret{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "my-random-secret"},
		Spec: RandomSecretSpec{
			Path:                vaultutils.Path(randomSecretPath),
			SecretKey:           "password",
			IsKVSecretsEngineV2: true,
		},
	}
	kube := newFakeKubeClient(randomSecret)
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &NomadSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			NomadSEConfig: NomadSEConfig{
				Address: "http://127.0.0.1:4646",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				RandomSecret: &corev1.LocalObjectReference{Name: "my-random-secret"},
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedToken != "random-generated-token" {
		t.Errorf("retrievedToken = %q, want %q", config.Spec.retrievedToken, "random-generated-token")
	}
}

func TestNomadSecretEngineConfig_IsValid_ExactlyOneCredentialSource(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "test"},
			},
		},
	}
	ok, err := config.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid with exactly one credential source, got ok=%v, err=%v", ok, err)
	}
}

func TestNomadSecretEngineConfig_IsValid_NoCredentialSource(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			RootCredentials: vaultutils.RootCredentialConfig{},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid with no credential source")
	}
}

func TestNomadSecretEngineConfig_IsValid_ClientCertWithoutKey(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			NomadSEConfig: NomadSEConfig{
				ClientCert: "-----BEGIN CERTIFICATE-----\nMIIC...",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "test"},
			},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid when clientCert is set without clientKey")
	}
}

func TestNomadSecretEngineConfig_IsValid_ClientKeyWithoutCert(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			NomadSEConfig: NomadSEConfig{
				ClientKey: "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "test"},
			},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid when clientKey is set without clientCert")
	}
}

func TestNomadSecretEngineConfig_IsValid_BothClientCertAndKey(t *testing.T) {
	config := &NomadSecretEngineConfig{
		Spec: NomadSecretEngineConfigSpec{
			NomadSEConfig: NomadSEConfig{
				ClientCert: "-----BEGIN CERTIFICATE-----\nMIIC...",
				ClientKey:  "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "test"},
			},
		},
	}
	ok, err := config.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid with both clientCert and clientKey, got ok=%v, err=%v", ok, err)
	}
}

func TestNomadSecretEngineConfig_SetInternalCredentials_K8sSecretMissingKey(t *testing.T) {
	ns := "ns-nomad"
	sec := newK8sSecret(ns, "nomad-creds", map[string][]byte{
		"wrong-key": []byte("value"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &NomadSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "nomad-creds"},
				PasswordKey: "token",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing password key")
	}
}

func TestNomadSecretEngineConfig_SetInternalCredentials_VaultSecretMissingKey(t *testing.T) {
	ns := "ns-nomad"
	vaultPath := "secret/data/nomad-token"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"wrong-key": "value",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &NomadSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: NomadSecretEngineConfigSpec{
			Path: "nomad",
			RootCredentials: vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				PasswordKey: "token",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when VaultSecret is missing password key")
	}
}
