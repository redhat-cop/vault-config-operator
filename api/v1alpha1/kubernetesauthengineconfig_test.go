package v1alpha1

import (
	"context"
	"reflect"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// --- TokenReviewerJWTSecret tests ---

func TestKubernetesAuthEngineConfig_PrepareInternalValues_JWTSecretHappyPath(t *testing.T) {
	secret := newK8sSecret("vault-ns", "reviewer-cred", map[string][]byte{
		"token": []byte("my-remote-jwt"),
	})
	kubeClient := newFakeKubeClient(secret)
	ctx := pivContext(kubeClient, nil)

	config := &KubernetesAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "k8scfg", Namespace: "vault-ns"},
		Spec: KubernetesAuthEngineConfigSpec{
			Path: "kubernetes",
			TokenReviewerJWTSecret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "reviewer-cred"},
				Key:                  "token",
			},
			KAECConfig: KAECConfig{
				KubernetesHost: "https://remote-cluster:443",
			},
		},
	}

	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedTokenReviewerJWT != "my-remote-jwt" {
		t.Fatalf("retrievedTokenReviewerJWT: got %q want %q", config.Spec.retrievedTokenReviewerJWT, "my-remote-jwt")
	}

	payload := config.GetPayload()
	if payload["token_reviewer_jwt"] != "my-remote-jwt" {
		t.Fatalf("payload token_reviewer_jwt: got %q want %q", payload["token_reviewer_jwt"], "my-remote-jwt")
	}
}

func TestKubernetesAuthEngineConfig_PrepareInternalValues_JWTSecretCustomKey(t *testing.T) {
	secret := newK8sSecret("vault-ns", "reviewer-cred", map[string][]byte{
		"jwt-data": []byte("custom-key-jwt"),
	})
	kubeClient := newFakeKubeClient(secret)
	ctx := pivContext(kubeClient, nil)

	config := &KubernetesAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "k8scfg", Namespace: "vault-ns"},
		Spec: KubernetesAuthEngineConfigSpec{
			Path: "kubernetes",
			TokenReviewerJWTSecret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "reviewer-cred"},
				Key:                  "jwt-data",
			},
			KAECConfig: KAECConfig{
				KubernetesHost: "https://remote-cluster:443",
			},
		},
	}

	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedTokenReviewerJWT != "custom-key-jwt" {
		t.Fatalf("retrievedTokenReviewerJWT: got %q want %q", config.Spec.retrievedTokenReviewerJWT, "custom-key-jwt")
	}
}

func TestKubernetesAuthEngineConfig_PrepareInternalValues_JWTSecretDefaultKey(t *testing.T) {
	secret := newK8sSecret("vault-ns", "reviewer-cred", map[string][]byte{
		"token": []byte("default-key-jwt"),
	})
	kubeClient := newFakeKubeClient(secret)
	ctx := pivContext(kubeClient, nil)

	config := &KubernetesAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "k8scfg", Namespace: "vault-ns"},
		Spec: KubernetesAuthEngineConfigSpec{
			Path: "kubernetes",
			TokenReviewerJWTSecret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "reviewer-cred"},
				Key:                  "", // empty key defaults to "token"
			},
			KAECConfig: KAECConfig{
				KubernetesHost: "https://remote-cluster:443",
			},
		},
	}

	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedTokenReviewerJWT != "default-key-jwt" {
		t.Fatalf("retrievedTokenReviewerJWT: got %q want %q", config.Spec.retrievedTokenReviewerJWT, "default-key-jwt")
	}
}

func TestKubernetesAuthEngineConfig_PrepareInternalValues_JWTSecretNotFound(t *testing.T) {
	kubeClient := newFakeKubeClient() // no secrets
	ctx := pivContext(kubeClient, nil)

	config := &KubernetesAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "k8scfg", Namespace: "vault-ns"},
		Spec: KubernetesAuthEngineConfigSpec{
			Path: "kubernetes",
			TokenReviewerJWTSecret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "nonexistent"},
				Key:                  "token",
			},
			KAECConfig: KAECConfig{
				KubernetesHost: "https://remote-cluster:443",
			},
		},
	}

	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when secret does not exist")
	}
}

func TestKubernetesAuthEngineConfig_PrepareInternalValues_JWTSecretKeyMissing(t *testing.T) {
	secret := newK8sSecret("vault-ns", "reviewer-cred", map[string][]byte{
		"other-key": []byte("some-value"),
	})
	kubeClient := newFakeKubeClient(secret)
	ctx := pivContext(kubeClient, nil)

	config := &KubernetesAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "k8scfg", Namespace: "vault-ns"},
		Spec: KubernetesAuthEngineConfigSpec{
			Path: "kubernetes",
			TokenReviewerJWTSecret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "reviewer-cred"},
				Key:                  "token",
			},
			KAECConfig: KAECConfig{
				KubernetesHost: "https://remote-cluster:443",
			},
		},
	}

	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when key is not found in secret")
	}
}

func TestKubernetesAuthEngineConfig_PrepareInternalValues_JWTSecretTakesPrecedence(t *testing.T) {
	secret := newK8sSecret("vault-ns", "reviewer-cred", map[string][]byte{
		"token": []byte("secret-jwt"),
	})
	kubeClient := newFakeKubeClient(secret)
	ctx := pivContext(kubeClient, nil)

	config := &KubernetesAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "k8scfg", Namespace: "vault-ns"},
		Spec: KubernetesAuthEngineConfigSpec{
			Path: "kubernetes",
			TokenReviewerJWTSecret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "reviewer-cred"},
				Key:                  "token",
			},
			TokenReviewerServiceAccount: &corev1.LocalObjectReference{
				Name: "my-service-account",
			},
			KAECConfig: KAECConfig{
				KubernetesHost: "https://remote-cluster:443",
			},
		},
	}

	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedTokenReviewerJWT != "secret-jwt" {
		t.Fatalf("expected Secret to take precedence, got %q", config.Spec.retrievedTokenReviewerJWT)
	}
}

// --- Mutual exclusivity validation ---

func TestKubernetesAuthEngineConfig_IsValid_BothSourcesSet(t *testing.T) {
	config := &KubernetesAuthEngineConfig{
		Spec: KubernetesAuthEngineConfigSpec{
			TokenReviewerServiceAccount: &corev1.LocalObjectReference{Name: "sa"},
			TokenReviewerJWTSecret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
				Key:                  "token",
			},
		},
	}
	valid, err := config.IsValid()
	if valid || err == nil {
		t.Fatal("expected validation error when both sources are set")
	}
}

func TestKubernetesAuthEngineConfig_IsValid_OnlySecretSet(t *testing.T) {
	config := &KubernetesAuthEngineConfig{
		Spec: KubernetesAuthEngineConfigSpec{
			TokenReviewerJWTSecret: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "secret"},
				Key:                  "token",
			},
		},
	}
	valid, err := config.IsValid()
	if !valid || err != nil {
		t.Fatalf("expected valid, got valid=%v err=%v", valid, err)
	}
}

func TestKubernetesAuthEngineConfig_IsValid_OnlySASet(t *testing.T) {
	config := &KubernetesAuthEngineConfig{
		Spec: KubernetesAuthEngineConfigSpec{
			TokenReviewerServiceAccount: &corev1.LocalObjectReference{Name: "sa"},
		},
	}
	valid, err := config.IsValid()
	if !valid || err != nil {
		t.Fatalf("expected valid, got valid=%v err=%v", valid, err)
	}
}

func TestKubernetesAuthEngineConfig_IsValid_NeitherSet(t *testing.T) {
	config := &KubernetesAuthEngineConfig{}
	valid, err := config.IsValid()
	if !valid || err != nil {
		t.Fatalf("expected valid, got valid=%v err=%v", valid, err)
	}
}

// --- toMap empty-string fix ---

func TestKAECConfigToMap_OmitsTokenReviewerJWTWhenEmpty(t *testing.T) {
	config := KAECConfig{
		KubernetesHost: "https://kubernetes.default.svc:443",
	}
	result := config.toMap()
	if _, exists := result["token_reviewer_jwt"]; exists {
		t.Error("toMap() should omit token_reviewer_jwt when retrievedTokenReviewerJWT is empty")
	}
}

func TestKAECConfigToMap_IncludesTokenReviewerJWTWhenSet(t *testing.T) {
	config := KAECConfig{
		KubernetesHost: "https://kubernetes.default.svc:443",
	}
	config.retrievedTokenReviewerJWT = "a-jwt"
	result := config.toMap()
	if result["token_reviewer_jwt"] != "a-jwt" {
		t.Errorf("toMap() should include token_reviewer_jwt when set, got %v", result["token_reviewer_jwt"])
	}
}

func TestKubernetesAuthEngineConfigGetPath(t *testing.T) {
	tests := []struct {
		name         string
		config       *KubernetesAuthEngineConfig
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			config: &KubernetesAuthEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: KubernetesAuthEngineConfigSpec{
					Path: "kubernetes",
					Name: "custom-name",
				},
			},
			expectedPath: "auth/kubernetes/custom-name/config",
		},
		{
			name: "without spec.name falls back to metadata.name",
			config: &KubernetesAuthEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: KubernetesAuthEngineConfigSpec{
					Path: "kubernetes",
				},
			},
			expectedPath: "auth/kubernetes/meta-name/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestKAECConfigToMap(t *testing.T) {
	config := KAECConfig{
		KubernetesHost:                "https://kubernetes.default.svc:443",
		KubernetesCACert:              "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
		PEMKeys:                       []string{"pem-key-1", "pem-key-2"},
		Issuer:                        "https://kubernetes.default.svc",
		DisableISSValidation:          true,
		DisableLocalCAJWT:             true,
		UseAnnotationsAsAliasMetadata: false,
	}

	result := config.toMap()

	// token_reviewer_jwt is omitted when empty, so expect 7 keys
	if len(result) != 7 {
		t.Errorf("expected 7 keys in map, got %d: %v", len(result), result)
	}

	expected := map[string]any{
		"kubernetes_host":                   "https://kubernetes.default.svc:443",
		"kubernetes_ca_cert":                "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
		"pem_keys":                          []any{"pem-key-1", "pem-key-2"},
		"issuer":                            "https://kubernetes.default.svc",
		"disable_iss_validation":            true,
		"disable_local_ca_jwt":              true,
		"use_annotations_as_alias_metadata": false,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestKAECConfigToMapUnexportedTokenReviewerJWT(t *testing.T) {
	config := KAECConfig{
		KubernetesHost: "https://k8s.example.com",
	}
	config.retrievedTokenReviewerJWT = "my-jwt-token"

	result := config.toMap()

	if result["token_reviewer_jwt"] != "my-jwt-token" {
		t.Errorf("expected token_reviewer_jwt = 'my-jwt-token', got %v", result["token_reviewer_jwt"])
	}
}

func TestKubernetesAuthEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &KubernetesAuthEngineConfig{
		Spec: KubernetesAuthEngineConfigSpec{
			KAECConfig: KAECConfig{
				KubernetesHost:       "https://k8s.example.com",
				KubernetesCACert:     "cert-data",
				PEMKeys:              []string{"key1"},
				Issuer:               "issuer",
				DisableISSValidation: false,
				DisableLocalCAJWT:    false,
			},
		},
	}

	payload := config.Spec.KAECConfig.toMap()

	if !vaultutils.IsEquivalentToDesiredState(config, payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestKubernetesAuthEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &KubernetesAuthEngineConfig{
		Spec: KubernetesAuthEngineConfigSpec{
			KAECConfig: KAECConfig{
				KubernetesHost:   "https://k8s.example.com",
				KubernetesCACert: "cert-data",
			},
		},
	}

	payload := config.Spec.KAECConfig.toMap()
	payload["kubernetes_host"] = "https://different-host.com"

	if vaultutils.IsEquivalentToDesiredState(config, payload) {
		t.Error("expected non-matching payload (different kubernetes_host) to NOT be equivalent")
	}
}

// Responses may include keys beyond KAECConfig.toMap(); extra keys are ignored
// via filterPayloadToDesiredKeys before comparison.
func TestKubernetesAuthEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &KubernetesAuthEngineConfig{
		Spec: KubernetesAuthEngineConfigSpec{
			KAECConfig: KAECConfig{
				KubernetesHost: "https://k8s.example.com",
			},
		},
	}

	payload := config.Spec.KAECConfig.toMap()
	payload["extra_field"] = "unexpected"

	if !vaultutils.IsEquivalentToDesiredState(config, payload) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestKubernetesAuthEngineConfigIsDeletable(t *testing.T) {
	config := &KubernetesAuthEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected KubernetesAuthEngineConfig to NOT be deletable")
	}
}

func TestKubernetesAuthEngineConfigConditions(t *testing.T) {
	config := &KubernetesAuthEngineConfig{}

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

func TestKubernetesAuthEngineConfig_PrepareInternalValues_NilTokenReviewerNoOp(t *testing.T) {
	config := &KubernetesAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "k8scfg", Namespace: "vault-ns"},
		Spec: KubernetesAuthEngineConfigSpec{
			Path:                        "kubernetes",
			TokenReviewerServiceAccount: nil,
			KAECConfig: KAECConfig{
				KubernetesHost: "https://kubernetes.default.svc:443",
			},
		},
	}
	ctx := context.Background()
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedTokenReviewerJWT != "" {
		t.Fatalf("retrievedTokenReviewerJWT: got %q want empty", config.Spec.retrievedTokenReviewerJWT)
	}
}

func TestKubernetesAuthEngineConfig_PrepareInternalValues_NonNilSAErrorPropagated(t *testing.T) {
	config := &KubernetesAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "k8scfg", Namespace: "vault-ns"},
		Spec: KubernetesAuthEngineConfigSpec{
			Path: "kubernetes",
			TokenReviewerServiceAccount: &corev1.LocalObjectReference{
				Name: "my-service-account",
			},
			KAECConfig: KAECConfig{
				KubernetesHost: "https://kubernetes.default.svc:443",
			},
		},
	}
	unreachable := &rest.Config{Host: "https://127.0.0.1:1"}
	ctx := pivContextWithRestConfig(newFakeKubeClient(), nil, unreachable)

	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when TokenReviewerServiceAccount is set but token request fails")
	}
	if config.Spec.retrievedTokenReviewerJWT != "" {
		t.Fatalf("retrievedTokenReviewerJWT should be empty on error, got %q", config.Spec.retrievedTokenReviewerJWT)
	}
}
