package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMongoDBAtlasSecretEngineConfigGetPath(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
		},
	}
	expected := "mongodbatlas/config"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestMongoDBAtlasSecretEngineConfigGetPathWithDifferentMount(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "my-atlas-engine",
		},
	}
	expected := "my-atlas-engine/config"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestMongoDBAtlasSecretEngineConfigIsDeletable(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected MongoDBAtlasSecretEngineConfig not to be deletable")
	}
}

func TestMongoDBAtlasSEConfig_toMap(t *testing.T) {
	config := MongoDBAtlasSEConfig{
		retrievedPublicKey:  "test-public-key",
		retrievedPrivateKey: "test-private-key",
	}

	result := config.toMap()

	if len(result) != 2 {
		t.Fatalf("expected 2 keys in toMap() output, got %d: %v", len(result), result)
	}
	if pk, ok := result["public_key"].(string); !ok || pk != "test-public-key" {
		t.Errorf("public_key = %v (%T), expected string test-public-key", result["public_key"], result["public_key"])
	}
	if sk, ok := result["private_key"].(string); !ok || sk != "test-private-key" {
		t.Errorf("private_key = %v (%T), expected string test-private-key", result["private_key"], result["private_key"])
	}
}

func TestMongoDBAtlasSEConfig_toMap_PublicKeyFromSpec(t *testing.T) {
	config := MongoDBAtlasSEConfig{
		PublicKey: "spec-public-key",
	}

	result := config.toMap()

	if pk, ok := result["public_key"].(string); !ok || pk != "spec-public-key" {
		t.Errorf("public_key = %v, expected spec-public-key", result["public_key"])
	}
	if _, ok := result["private_key"]; ok {
		t.Error("private_key should not be present when no retrieved private key")
	}
}

func TestMongoDBAtlasSEConfig_toMap_RetrievedPublicKeyOverridesSpec(t *testing.T) {
	config := MongoDBAtlasSEConfig{
		PublicKey:           "spec-public-key",
		retrievedPublicKey:  "retrieved-public-key",
		retrievedPrivateKey: "test-private-key",
	}

	result := config.toMap()

	if pk, ok := result["public_key"].(string); !ok || pk != "retrieved-public-key" {
		t.Errorf("public_key = %v, expected retrieved-public-key (retrieved should override spec)", result["public_key"])
	}
}

func TestMongoDBAtlasSecretEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSEConfig: MongoDBAtlasSEConfig{
				PublicKey: "test-public-key",
			},
		},
	}

	vaultPayload := map[string]any{
		"public_key": "test-public-key",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match")
	}
}

func TestMongoDBAtlasSecretEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSEConfig: MongoDBAtlasSEConfig{
				PublicKey: "new-public-key",
			},
		},
	}

	vaultPayload := map[string]any{
		"public_key": "old-public-key",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when public_key differs")
	}
}

func TestMongoDBAtlasSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSEConfig: MongoDBAtlasSEConfig{
				PublicKey: "test-public-key",
			},
		},
	}

	vaultPayload := map[string]any{
		"public_key":  "test-public-key",
		"extra_field": "from-vault",
		"another":     123,
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestMongoDBAtlasSecretEngineConfig_IsEquivalentToDesiredState_PrivateKeyExcluded(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSEConfig: MongoDBAtlasSEConfig{
				PublicKey:           "test-public-key",
				retrievedPublicKey:  "test-public-key",
				retrievedPrivateKey: "test-private-key",
			},
		},
	}

	vaultPayload := map[string]any{
		"public_key": "test-public-key",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: private_key is write-only and must be excluded from drift comparison")
	}
}

func TestMongoDBAtlasSecretEngineConfig_IsEquivalentToDesiredState_PublicKeyChangeDetected(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSEConfig: MongoDBAtlasSEConfig{
				retrievedPublicKey:  "new-key",
				retrievedPrivateKey: "test-private-key",
			},
		},
	}

	vaultPayload := map[string]any{
		"public_key": "old-key",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: public_key change must be detected")
	}
}

func TestMongoDBAtlasSecretEngineConfigConditions(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{}

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

func TestMongoDBAtlasSecretEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-atlas"
	sec := newK8sSecret(ns, "atlas-creds", map[string][]byte{
		"username": []byte("test-public-key"),
		"password": []byte("test-private-key"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &MongoDBAtlasSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "atlas-creds"},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPublicKey != "test-public-key" {
		t.Errorf("retrievedPublicKey = %q, want %q", config.Spec.retrievedPublicKey, "test-public-key")
	}
	if config.Spec.retrievedPrivateKey != "test-private-key" {
		t.Errorf("retrievedPrivateKey = %q, want %q", config.Spec.retrievedPrivateKey, "test-private-key")
	}
}

func TestMongoDBAtlasSecretEngineConfig_PrepareInternalValues_FromVaultSecret(t *testing.T) {
	ns := "ns-atlas"
	vaultPath := "secret/data/atlas-creds"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"username": "vault-public-key",
		"password": "vault-private-key",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &MongoDBAtlasSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			RootCredentials: vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPublicKey != "vault-public-key" {
		t.Errorf("retrievedPublicKey = %q, want %q", config.Spec.retrievedPublicKey, "vault-public-key")
	}
	if config.Spec.retrievedPrivateKey != "vault-private-key" {
		t.Errorf("retrievedPrivateKey = %q, want %q", config.Spec.retrievedPrivateKey, "vault-private-key")
	}
}

func TestMongoDBAtlasSecretEngineConfig_PrepareInternalValues_SpecPublicKeyOverride(t *testing.T) {
	ns := "ns-atlas"
	sec := newK8sSecret(ns, "atlas-creds", map[string][]byte{
		"username": []byte("secret-public-key"),
		"password": []byte("test-private-key"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &MongoDBAtlasSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSEConfig: MongoDBAtlasSEConfig{
				PublicKey: "spec-public-key",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "atlas-creds"},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedPublicKey != "spec-public-key" {
		t.Errorf("retrievedPublicKey = %q, want %q (spec.publicKey should take precedence)", config.Spec.retrievedPublicKey, "spec-public-key")
	}
}

func TestMongoDBAtlasSecretEngineConfig_IsValid_ExactlyOneCredentialSource(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
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

func TestMongoDBAtlasSecretEngineConfig_IsValid_NoCredentialSource(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			RootCredentials: vaultutils.RootCredentialConfig{},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid with no credential source")
	}
}

func TestMongoDBAtlasSecretEngineConfig_IsValid_RandomSecretWithoutPublicKey(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			RootCredentials: vaultutils.RootCredentialConfig{
				RandomSecret: &corev1.LocalObjectReference{Name: "random"},
			},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid when randomSecret is used without spec.publicKey")
	}
}

func TestMongoDBAtlasSecretEngineConfig_IsValid_RandomSecretWithPublicKey(t *testing.T) {
	config := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			MongoDBAtlasSEConfig: MongoDBAtlasSEConfig{
				PublicKey: "my-public-key",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				RandomSecret: &corev1.LocalObjectReference{Name: "random"},
			},
		},
	}
	ok, err := config.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid with randomSecret and spec.publicKey set, got ok=%v, err=%v", ok, err)
	}
}

func TestMongoDBAtlasSecretEngineConfig_ValidateUpdate_RejectsPathChange(t *testing.T) {
	r := &MongoDBAtlasSecretEngineConfig{}
	oldObj := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	newObj := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "different-path",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.path is changed")
	}
}

func TestMongoDBAtlasSecretEngineConfig_ValidateUpdate_AllowsSamePathUpdate(t *testing.T) {
	r := &MongoDBAtlasSecretEngineConfig{}
	oldObj := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	newObj := &MongoDBAtlasSecretEngineConfig{
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSEConfig: MongoDBAtlasSEConfig{
				PublicKey: "new-key",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err != nil {
		t.Errorf("expected no error when spec.path unchanged, got: %v", err)
	}
}

func TestMongoDBAtlasSecretEngineConfig_SetInternalCredentials_K8sSecretMissingPasswordKey(t *testing.T) {
	ns := "ns-atlas"
	sec := newK8sSecret(ns, "atlas-creds", map[string][]byte{
		"username": []byte("test-public-key"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &MongoDBAtlasSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "atlas-creds"},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing password key")
	}
}

func TestMongoDBAtlasSecretEngineConfig_SetInternalCredentials_K8sSecretMissingUsernameKey(t *testing.T) {
	ns := "ns-atlas"
	sec := newK8sSecret(ns, "atlas-creds", map[string][]byte{
		"password": []byte("test-private-key"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &MongoDBAtlasSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: MongoDBAtlasSecretEngineConfigSpec{
			Path: "mongodbatlas",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "atlas-creds"},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing username key")
	}
}
