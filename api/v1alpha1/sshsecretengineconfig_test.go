package v1alpha1

import (
	"context"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSSHSecretEngineConfigGetPath(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
		},
	}
	if got := config.GetPath(); got != "ssh/config/ca" {
		t.Errorf("GetPath() = %q, expected %q", got, "ssh/config/ca")
	}
}

func TestSSHSEConfigToMap(t *testing.T) {
	config := SSHSEConfig{
		GenerateSigningKey:  true,
		KeyType:             "ssh-rsa",
		KeyBits:             4096,
		retrievedPrivateKey: "private-key-data",
		retrievedPublicKey:  "public-key-data",
	}

	result := config.toMap()

	expectedKeys := []string{"generate_signing_key", "key_type", "key_bits", "private_key", "public_key"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys in toMap() output, got %d", len(expectedKeys), len(result))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}
	if result["generate_signing_key"] != true {
		t.Errorf("generate_signing_key = %v, want true", result["generate_signing_key"])
	}
	if result["key_type"] != "ssh-rsa" {
		t.Errorf("key_type = %v, want ssh-rsa", result["key_type"])
	}
	if result["key_bits"] != 4096 {
		t.Errorf("key_bits = %v, want 4096", result["key_bits"])
	}
	if result["private_key"] != "private-key-data" {
		t.Errorf("private_key = %v", result["private_key"])
	}
	if result["public_key"] != "public-key-data" {
		t.Errorf("public_key = %v", result["public_key"])
	}
}

func TestSSHSecretEngineConfigIsEquivalentPrivateKeyDeleted(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey:  true,
				KeyType:             "ssh-rsa",
				KeyBits:             4096,
				retrievedPrivateKey: "secret-private-key",
				retrievedPublicKey:  "ssh-rsa AAAA...",
			},
		},
	}

	payload := map[string]any{
		"generate_signing_key": true,
		"key_type":             "ssh-rsa",
		"key_bits":             4096,
		"public_key":           "ssh-rsa AAAA...",
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true when payload omits private_key but matches other fields after private_key delete")
	}
}

func TestSSHSecretEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey:  true,
				KeyType:             "ssh-rsa",
				KeyBits:             0,
				retrievedPrivateKey: "",
				retrievedPublicKey:  "ssh-rsa AAAA...",
			},
		},
	}

	payload := map[string]any{
		"generate_signing_key": true,
		"key_type":             "ssh-rsa",
		"key_bits":             0,
		"public_key":           "ssh-rsa AAAA...",
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for payload matching desired state without private_key")
	}
}

func TestSSHSecretEngineConfigIsEquivalentNonMatching_ExternalKey(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey:  false,
				KeyType:             "ssh-rsa",
				KeyBits:             4096,
				retrievedPrivateKey: "key-data",
				retrievedPublicKey:  "ssh-rsa AAAA...",
			},
		},
	}

	payload := map[string]any{
		"public_key": "ecdsa-sha2-nistp256 BBBB... different-key",
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when public_key differs in external key mode")
	}
}

func TestSSHSecretEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: true,
				KeyType:            "ssh-rsa",
				KeyBits:            0,
				retrievedPublicKey: "ssh-rsa AAAA...",
			},
		},
	}

	payload := map[string]any{
		"generate_signing_key": true,
		"key_type":             "ssh-rsa",
		"key_bits":             0,
		"public_key":           "ssh-rsa AAAA...",
		"extra_vault_field":    "should-be-ignored",
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestSSHSecretEngineConfigIsEquivalentPayloadWithPrivateKey(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey:  true,
				KeyType:             "ssh-rsa",
				KeyBits:             4096,
				retrievedPrivateKey: "private-key",
				retrievedPublicKey:  "ssh-rsa AAAA...",
			},
		},
	}

	payload := map[string]any{
		"generate_signing_key": true,
		"key_type":             "ssh-rsa",
		"key_bits":             4096,
		"public_key":           "ssh-rsa AAAA...",
		"private_key":          "private-key",
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: private_key is deleted from desiredState and filtered from payload")
	}
}

func TestSSHSecretEngineConfigIsEquivalent_GenerateSigningKey_OnlyPublicKeyCompared(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: true,
				KeyType:            "ssh-rsa",
				KeyBits:            4096,
			},
		},
	}

	payload := map[string]any{
		"public_key": "ssh-rsa AAAA... vault-generated",
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: with generateSigningKey=true, only public_key should be compared; create-time fields should be excluded")
	}
}

func TestSSHSecretEngineConfigIsEquivalent_GenerateSigningKey_ConvergesWithVaultPayload(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: true,
				KeyType:            "ssh-rsa",
				KeyBits:            4096,
			},
		},
	}

	payload := map[string]any{
		"public_key":    "ssh-rsa AAAA... vault-generated-key",
		"extra_field_1": "ignored",
		"extra_field_2": 42,
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: with generateSigningKey=true, all fields are create-time-only, so any existing Vault CA should converge")
	}
}

func TestSSHSecretEngineConfigIsEquivalent_ExternalKey_OnlyPublicKeyCompared(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey:  false,
				KeyType:             "ssh-rsa",
				KeyBits:             4096,
				retrievedPrivateKey: "secret-key",
				retrievedPublicKey:  "ssh-rsa AAAA...",
			},
		},
	}

	payload := map[string]any{
		"public_key": "ssh-rsa AAAA...",
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: external key mode should only compare public_key (Vault GET config/ca returns only public_key)")
	}
}

func TestSSHSecretEngineConfigIsEquivalent_ExternalKey_DetectsPublicKeyDrift(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey:  false,
				KeyType:             "ssh-rsa",
				KeyBits:             4096,
				retrievedPrivateKey: "secret-key",
				retrievedPublicKey:  "ssh-rsa AAAA... expected-key",
			},
		},
	}

	payload := map[string]any{
		"public_key": "ssh-rsa BBBB... different-key-in-vault",
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false: public_key drift should be detected in external key mode")
	}
}

func TestSSHSecretEngineConfigIsDeletable(t *testing.T) {
	config := &SSHSecretEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected SSHSecretEngineConfig to be deletable")
	}
}

func TestSSHSecretEngineConfigConditions(t *testing.T) {
	config := &SSHSecretEngineConfig{}

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

func TestSSHSecretEngineConfigIsValid_GenerateSigningKey(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: true,
			},
		},
	}
	valid, err := config.IsValid()
	if !valid || err != nil {
		t.Errorf("expected valid when generateSigningKey=true, got valid=%v, err=%v", valid, err)
	}
}

func TestSSHSecretEngineConfigIsValid_NoGenerateNoRef(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: false,
			},
		},
	}
	valid, err := config.IsValid()
	if valid || err == nil {
		t.Error("expected invalid when generateSigningKey=false and caKeyReference is nil")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_GenerateSigningKey(t *testing.T) {
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: true,
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPrivateKey != "" {
		t.Errorf("expected empty retrievedPrivateKey, got %q", config.Spec.retrievedPrivateKey)
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromSecret(t *testing.T) {
	ns := "ns-ssh-se"
	sec := newK8sSecret(ns, "ssh-ca-key", map[string][]byte{
		"private_key": []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIE..."),
		"public_key":  []byte("ssh-rsa AAAA..."),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: false,
			},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "ssh-ca-key"},
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPrivateKey != "-----BEGIN RSA PRIVATE KEY-----\nMIIE..." {
		t.Errorf("retrievedPrivateKey = %q", config.Spec.retrievedPrivateKey)
	}
	if config.Spec.retrievedPublicKey != "ssh-rsa AAAA..." {
		t.Errorf("retrievedPublicKey = %q", config.Spec.retrievedPublicKey)
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromSecret_CustomKeyFields(t *testing.T) {
	ns := "ns-ssh-custom"
	sec := newK8sSecret(ns, "ssh-ca-custom", map[string][]byte{
		"my_priv": []byte("custom-private-key"),
		"my_pub":  []byte("custom-public-key"),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: false,
			},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "ssh-ca-custom"},
				PasswordKey: "my_priv",
				UsernameKey: "my_pub",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPrivateKey != "custom-private-key" {
		t.Errorf("retrievedPrivateKey = %q, want %q", config.Spec.retrievedPrivateKey, "custom-private-key")
	}
	if config.Spec.retrievedPublicKey != "custom-public-key" {
		t.Errorf("retrievedPublicKey = %q, want %q", config.Spec.retrievedPublicKey, "custom-public-key")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromVaultSecret_CustomKeyFields(t *testing.T) {
	ns := "ns-ssh-vault-custom"
	vaultPath := "secret/data/ssh-ca-custom"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"my_priv": "vault-custom-private",
		"my_pub":  "vault-custom-public",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: false,
			},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				PasswordKey: "my_priv",
				UsernameKey: "my_pub",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPrivateKey != "vault-custom-private" {
		t.Errorf("retrievedPrivateKey = %q, want %q", config.Spec.retrievedPrivateKey, "vault-custom-private")
	}
	if config.Spec.retrievedPublicKey != "vault-custom-public" {
		t.Errorf("retrievedPublicKey = %q, want %q", config.Spec.retrievedPublicKey, "vault-custom-public")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromVaultSecret(t *testing.T) {
	ns := "ns-ssh-se"
	vaultPath := "secret/data/ssh-ca"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"private_key": "vault-private-key",
		"public_key":  "vault-public-key",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: false,
			},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPrivateKey != "vault-private-key" {
		t.Errorf("retrievedPrivateKey = %q", config.Spec.retrievedPrivateKey)
	}
	if config.Spec.retrievedPublicKey != "vault-public-key" {
		t.Errorf("retrievedPublicKey = %q", config.Spec.retrievedPublicKey)
	}
}

func TestSSHSecretEngineConfig_Default_OverridesGenericDefaults(t *testing.T) {
	r := &SSHSecretEngineConfig{}
	obj := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ssh"},
		Spec: SSHSecretEngineConfigSpec{
			CAKeyReference: &vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "ssh-ca"},
				PasswordKey: "password",
				UsernameKey: "username",
			},
		},
	}
	if err := r.Default(context.Background(), obj); err != nil {
		t.Fatalf("Default: %v", err)
	}
	if obj.Spec.CAKeyReference.PasswordKey != "private_key" {
		t.Errorf("PasswordKey = %q, want %q", obj.Spec.CAKeyReference.PasswordKey, "private_key")
	}
	if obj.Spec.CAKeyReference.UsernameKey != "public_key" {
		t.Errorf("UsernameKey = %q, want %q", obj.Spec.CAKeyReference.UsernameKey, "public_key")
	}
}

func TestSSHSecretEngineConfig_Default_PreservesCustomKeys(t *testing.T) {
	r := &SSHSecretEngineConfig{}
	obj := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ssh-custom"},
		Spec: SSHSecretEngineConfigSpec{
			CAKeyReference: &vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "ssh-ca"},
				PasswordKey: "my_priv",
				UsernameKey: "my_pub",
			},
		},
	}
	if err := r.Default(context.Background(), obj); err != nil {
		t.Fatalf("Default: %v", err)
	}
	if obj.Spec.CAKeyReference.PasswordKey != "my_priv" {
		t.Errorf("PasswordKey = %q, want %q", obj.Spec.CAKeyReference.PasswordKey, "my_priv")
	}
	if obj.Spec.CAKeyReference.UsernameKey != "my_pub" {
		t.Errorf("UsernameKey = %q, want %q", obj.Spec.CAKeyReference.UsernameKey, "my_pub")
	}
}

func TestSSHSecretEngineConfig_Default_NilCAKeyReference(t *testing.T) {
	r := &SSHSecretEngineConfig{}
	obj := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "test-ssh-noref"},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: true},
		},
	}
	if err := r.Default(context.Background(), obj); err != nil {
		t.Fatalf("Default: %v", err)
	}
	if obj.Spec.CAKeyReference != nil {
		t.Errorf("CAKeyReference should remain nil")
	}
}

func TestSSHSecretEngineConfig_caPrivateKeyField_KubebuilderDefault(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			CAKeyReference: &vaultutils.RootCredentialConfig{
				PasswordKey: "password",
				UsernameKey: "username",
			},
		},
	}
	if got := config.caPrivateKeyField(); got != "private_key" {
		t.Errorf("caPrivateKeyField() = %q, want %q (should ignore kubebuilder default 'password')", got, "private_key")
	}
	if got := config.caPublicKeyField(); got != "public_key" {
		t.Errorf("caPublicKeyField() = %q, want %q (should ignore kubebuilder default 'username')", got, "public_key")
	}
}

func TestSSHSecretEngineConfig_caPrivateKeyField_CustomFieldNames(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			CAKeyReference: &vaultutils.RootCredentialConfig{
				PasswordKey: "my_custom_priv",
				UsernameKey: "my_custom_pub",
			},
		},
	}
	if got := config.caPrivateKeyField(); got != "my_custom_priv" {
		t.Errorf("caPrivateKeyField() = %q, want %q", got, "my_custom_priv")
	}
	if got := config.caPublicKeyField(); got != "my_custom_pub" {
		t.Errorf("caPublicKeyField() = %q, want %q", got, "my_custom_pub")
	}
}

func TestSSHSecretEngineConfig_caPrivateKeyField_Empty(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			CAKeyReference: &vaultutils.RootCredentialConfig{
				PasswordKey: "",
				UsernameKey: "",
			},
		},
	}
	if got := config.caPrivateKeyField(); got != "private_key" {
		t.Errorf("caPrivateKeyField() = %q, want %q", got, "private_key")
	}
	if got := config.caPublicKeyField(); got != "public_key" {
		t.Errorf("caPublicKeyField() = %q, want %q", got, "public_key")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromVaultSecret_MissingPrivateKey(t *testing.T) {
	vaultPath := "secret/data/ssh-ca-no-priv"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"public_key": "ssh-rsa AAAA...",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-test"},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when vault secret is missing private_key field")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromVaultSecret_NilSecret(t *testing.T) {
	vaultPath := "secret/data/nonexistent"
	handler := newFakeVaultHandler()
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-test"},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when vault secret path returns nil")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromSecret_KubebuilderDefaults(t *testing.T) {
	ns := "ns-ssh-defaults"
	sec := newK8sSecret(ns, "ssh-ca-key", map[string][]byte{
		"private_key": []byte("priv-from-default-field"),
		"public_key":  []byte("pub-from-default-field"),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "ssh-ca-key"},
				PasswordKey: "password",
				UsernameKey: "username",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPrivateKey != "priv-from-default-field" {
		t.Errorf("retrievedPrivateKey = %q, want %q (should resolve using 'private_key' not 'password')",
			config.Spec.retrievedPrivateKey, "priv-from-default-field")
	}
	if config.Spec.retrievedPublicKey != "pub-from-default-field" {
		t.Errorf("retrievedPublicKey = %q, want %q (should resolve using 'public_key' not 'username')",
			config.Spec.retrievedPublicKey, "pub-from-default-field")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromVaultSecret_KubebuilderDefaults(t *testing.T) {
	vaultPath := "secret/data/ssh-ca-defaults"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"private_key": "vault-priv-from-default-field",
		"public_key":  "vault-pub-from-default-field",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-test"},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				PasswordKey: "password",
				UsernameKey: "username",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPrivateKey != "vault-priv-from-default-field" {
		t.Errorf("retrievedPrivateKey = %q, want %q (should resolve using 'private_key' not 'password')",
			config.Spec.retrievedPrivateKey, "vault-priv-from-default-field")
	}
	if config.Spec.retrievedPublicKey != "vault-pub-from-default-field" {
		t.Errorf("retrievedPublicKey = %q, want %q (should resolve using 'public_key' not 'username')",
			config.Spec.retrievedPublicKey, "vault-pub-from-default-field")
	}
}

func TestSSHSecretEngineConfigIsEquivalent_ExternalKey_ConvergesWithVaultPayload(t *testing.T) {
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey:  false,
				KeyType:             "ssh-rsa",
				KeyBits:             4096,
				retrievedPrivateKey: "secret-key",
				retrievedPublicKey:  "ssh-rsa AAAA...",
			},
		},
	}

	payload := map[string]any{
		"public_key":    "ssh-rsa AAAA...",
		"extra_field_1": "ignored",
		"extra_field_2": 42,
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: extra fields in Vault response should be ignored, only public_key is compared")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromVaultSecret_KVv2(t *testing.T) {
	vaultPath := "secret/data/ssh-ca-v2"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"data": map[string]any{
			"private_key": "kv2-private-key",
			"public_key":  "kv2-public-key",
		},
		"metadata": map[string]any{
			"version": 1,
		},
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-test"},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues with KV v2: %v", err)
	}
	if config.Spec.retrievedPrivateKey != "kv2-private-key" {
		t.Errorf("retrievedPrivateKey = %q, want %q", config.Spec.retrievedPrivateKey, "kv2-private-key")
	}
	if config.Spec.retrievedPublicKey != "kv2-public-key" {
		t.Errorf("retrievedPublicKey = %q, want %q", config.Spec.retrievedPublicKey, "kv2-public-key")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromSecret_MissingPrivateKey(t *testing.T) {
	ns := "ns-ssh-missing-priv"
	sec := newK8sSecret(ns, "ssh-ca-no-priv", map[string][]byte{
		"public_key": []byte("ssh-rsa AAAA..."),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "ssh-ca-no-priv"},
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing private_key field")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromSecret_EmptyPrivateKey(t *testing.T) {
	ns := "ns-ssh-empty-priv"
	sec := newK8sSecret(ns, "ssh-ca-empty-priv", map[string][]byte{
		"private_key": []byte(""),
		"public_key":  []byte("ssh-rsa AAAA..."),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "ssh-ca-empty-priv"},
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret has empty private_key field")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromSecret_MissingPublicKey(t *testing.T) {
	ns := "ns-ssh-missing-pub"
	sec := newK8sSecret(ns, "ssh-ca-no-pub", map[string][]byte{
		"private_key": []byte("-----BEGIN RSA PRIVATE KEY-----\nMIIE..."),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "ssh-ca-no-pub"},
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing public_key field")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromVaultSecret_MissingPublicKey(t *testing.T) {
	vaultPath := "secret/data/ssh-ca-no-pub"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"private_key": "vault-private-key",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-test"},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when vault secret is missing public_key field")
	}
}

func TestSSHSecretEngineConfig_PrepareInternalValues_FromVaultSecret_EmptyPublicKey(t *testing.T) {
	vaultPath := "secret/data/ssh-ca-empty-pub"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"private_key": "vault-private-key",
		"public_key":  "",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &SSHSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-test"},
		Spec: SSHSecretEngineConfigSpec{
			SSHSEConfig: SSHSEConfig{GenerateSigningKey: false},
			CAKeyReference: &vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when vault secret has empty public_key field")
	}
}

func TestSSHSecretEngineConfig_IsEquivalent_GeneratedCA_ExistingCAIsConverged(t *testing.T) {
	// Design decision test: When generateSigningKey=true and a CA already exists
	// in Vault, we treat it as converged. Vault's SSH CA is immutable once created
	// (/config/ca is write-once), so any existing CA satisfies the desired state.
	// Re-generating would destroy the CA and invalidate all signed certificates.
	config := &SSHSecretEngineConfig{
		Spec: SSHSecretEngineConfigSpec{
			Path: "ssh",
			SSHSEConfig: SSHSEConfig{
				GenerateSigningKey: true,
				KeyType:            "ssh-rsa",
				KeyBits:            4096,
			},
		},
	}

	tests := []struct {
		name    string
		payload map[string]any
	}{
		{
			name:    "vault returns only public_key",
			payload: map[string]any{"public_key": "ssh-rsa AAAA...vault-generated"},
		},
		{
			name:    "vault returns public_key with different key_type",
			payload: map[string]any{"public_key": "ecdsa-sha2-nistp256 AAAA...", "key_type": "ecdsa-sha2-nistp256"},
		},
		{
			name:    "vault returns empty payload",
			payload: map[string]any{},
		},
		{
			name:    "vault returns extra metadata fields",
			payload: map[string]any{"public_key": "ssh-rsa AAAA...", "nonce": "abc123", "serial": 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !config.IsEquivalentToDesiredState(tt.payload) {
				t.Errorf("expected true: with generateSigningKey=true, any existing CA in Vault should be treated as converged (immutable write-once resource)")
			}
		})
	}
}
