package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConsulSecretEngineConfigGetPath(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
		},
	}
	expected := "consul/config/access"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestConsulSecretEngineConfigGetPathWithDifferentMount(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			Path: "my-consul-engine",
		},
	}
	expected := "my-consul-engine/config/access"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestConsulSecretEngineConfigIsDeletable(t *testing.T) {
	config := &ConsulSecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected ConsulSecretEngineConfig not to be deletable")
	}
}

func TestConsulSEConfig_toMap(t *testing.T) {
	config := ConsulSEConfig{
		Address:        "127.0.0.1:8500",
		Scheme:         "https",
		CACert:         "-----BEGIN CERTIFICATE-----\nMIIB...",
		ClientCert:     "-----BEGIN CERTIFICATE-----\nMIIC...",
		ClientKey:      "-----BEGIN RSA PRIVATE KEY-----\nMIIE...",
		retrievedToken: "my-consul-token",
	}

	result := config.toMap()

	expectedKeys := []string{"address", "scheme", "token", "ca_cert", "client_cert", "client_key"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys in toMap() output, got %d: %v", len(expectedKeys), len(result), result)
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if addr, ok := result["address"].(string); !ok || addr != "127.0.0.1:8500" {
		t.Errorf("address = %v, expected 127.0.0.1:8500", result["address"])
	}
	if scheme, ok := result["scheme"].(string); !ok || scheme != "https" {
		t.Errorf("scheme = %v, expected https", result["scheme"])
	}
	if token, ok := result["token"].(string); !ok || token != "my-consul-token" {
		t.Errorf("token = %v, expected my-consul-token", result["token"])
	}
	if ca, ok := result["ca_cert"].(string); !ok || ca != "-----BEGIN CERTIFICATE-----\nMIIB..." {
		t.Errorf("ca_cert = %v, expected PEM cert", result["ca_cert"])
	}
}

func TestConsulSEConfig_toMap_NoToken(t *testing.T) {
	config := ConsulSEConfig{
		Address: "127.0.0.1:8500",
		Scheme:  "http",
	}

	result := config.toMap()

	if _, ok := result["token"]; ok {
		t.Error("expected token to be absent when retrievedToken is empty")
	}
	if result["address"] != "127.0.0.1:8500" {
		t.Errorf("address = %v, expected 127.0.0.1:8500", result["address"])
	}
	if result["scheme"] != "http" {
		t.Errorf("scheme = %v, expected http", result["scheme"])
	}
}

func TestConsulSecretEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address:        "127.0.0.1:8500",
				Scheme:         "https",
				retrievedToken: "my-token",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "127.0.0.1:8500",
		"scheme":  "https",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match (token/certs stripped from comparison)")
	}
}

func TestConsulSecretEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "https",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "10.0.0.1:8500",
		"scheme":  "https",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when address differs")
	}
}

func TestConsulSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "http",
			},
		},
	}

	vaultPayload := map[string]any{
		"address":     "127.0.0.1:8500",
		"scheme":      "http",
		"extra_field": "from-vault",
		"another":     123,
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestConsulSecretEngineConfig_IsEquivalentToDesiredState_TokenInPayload(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address:        "127.0.0.1:8500",
				Scheme:         "http",
				retrievedToken: "my-token",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "127.0.0.1:8500",
		"scheme":  "http",
		"token":   "some-unexpected-token",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: token is deleted from desiredState before comparison, so it's filtered as extra field")
	}
}

func TestConsulSecretEngineConfig_IsEquivalentToDesiredState_SchemeChange(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "https",
			},
		},
	}

	vaultPayload := map[string]any{
		"address": "127.0.0.1:8500",
		"scheme":  "http",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: scheme change must be detected")
	}
}

func TestConsulSecretEngineConfigConditions(t *testing.T) {
	config := &ConsulSecretEngineConfig{}

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

func TestConsulSecretEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-consul"
	sec := newK8sSecret(ns, "consul-creds", map[string][]byte{
		"token": []byte("my-consul-mgmt-token"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &ConsulSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "http",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "consul-creds"},
				PasswordKey: "token",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedToken != "my-consul-mgmt-token" {
		t.Errorf("retrievedToken = %q, want %q", config.Spec.retrievedToken, "my-consul-mgmt-token")
	}
}

func TestConsulSecretEngineConfig_PrepareInternalValues_FromVaultSecret(t *testing.T) {
	ns := "ns-consul"
	vaultPath := "secret/data/consul-token"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"token": "vault-stored-consul-token",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &ConsulSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "http",
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
	if config.Spec.retrievedToken != "vault-stored-consul-token" {
		t.Errorf("retrievedToken = %q, want %q", config.Spec.retrievedToken, "vault-stored-consul-token")
	}
}

func TestConsulSecretEngineConfig_IsValid_ExactlyOneCredentialSource(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
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

func TestConsulSecretEngineConfig_IsValid_NoCredentialSource(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			RootCredentials: vaultutils.RootCredentialConfig{},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid with no credential source")
	}
}

func TestConsulSecretEngineConfig_SetInternalCredentials_K8sSecretMissingKey(t *testing.T) {
	ns := "ns-consul"
	sec := newK8sSecret(ns, "consul-creds", map[string][]byte{
		"wrong-key": []byte("value"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &ConsulSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "consul-creds"},
				PasswordKey: "token",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing password key")
	}
}

func TestConsulSecretEngineConfig_SetInternalCredentials_VaultSecretMissingKey(t *testing.T) {
	ns := "ns-consul"
	vaultPath := "secret/data/consul-token"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"wrong-key": "value",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &ConsulSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
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

func TestConsulSecretEngineConfig_PrepareInternalValues_FromVaultSecretKVv2(t *testing.T) {
	ns := "ns-consul"
	vaultPath := "secret/data/consul-token"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"data": map[string]any{
			"token": "kv2-consul-token",
		},
		"metadata": map[string]any{
			"version": 1,
		},
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &ConsulSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "http",
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
	if config.Spec.retrievedToken != "kv2-consul-token" {
		t.Errorf("retrievedToken = %q, want %q", config.Spec.retrievedToken, "kv2-consul-token")
	}
}

func TestConsulSecretEngineConfig_PrepareInternalValues_VaultSecretKVv1WithDataStringKey(t *testing.T) {
	ns := "ns-consul"
	vaultPath := "secret/consul-token"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"token": "kv1-consul-token",
		"data":  "extra-metadata-string",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &ConsulSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "http",
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
	if config.Spec.retrievedToken != "kv1-consul-token" {
		t.Errorf("retrievedToken = %q, want %q", config.Spec.retrievedToken, "kv1-consul-token")
	}
}

func TestConsulSecretEngineConfig_PrepareInternalValues_VaultSecretKVv2FallbackToTopLevel(t *testing.T) {
	ns := "ns-consul"
	vaultPath := "secret/data/consul-token"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"data": map[string]any{
			"other-key": "other-value",
		},
		"token": "top-level-token",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &ConsulSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "http",
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
	if config.Spec.retrievedToken != "top-level-token" {
		t.Errorf("retrievedToken = %q, want %q", config.Spec.retrievedToken, "top-level-token")
	}
}

func TestConsulSecretEngineConfig_PrepareInternalValues_FromRandomSecret(t *testing.T) {
	ns := "ns-consul"
	randomSecretPath := "secret/data/random-consul"
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
	config := &ConsulSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: ConsulSecretEngineConfigSpec{
			Path: "consul",
			ConsulSEConfig: ConsulSEConfig{
				Address: "127.0.0.1:8500",
				Scheme:  "http",
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

func TestConsulSecretEngineConfig_IsValid_ClientCertWithoutKey(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			ConsulSEConfig: ConsulSEConfig{
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

func TestConsulSecretEngineConfig_IsValid_ClientKeyWithoutCert(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			ConsulSEConfig: ConsulSEConfig{
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

func TestConsulSecretEngineConfig_IsValid_BothClientCertAndKey(t *testing.T) {
	config := &ConsulSecretEngineConfig{
		Spec: ConsulSecretEngineConfigSpec{
			ConsulSEConfig: ConsulSEConfig{
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
