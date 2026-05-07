package v1alpha1

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

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

	if len(result) != 8 {
		t.Errorf("expected 8 keys in map, got %d", len(result))
	}

	expected := map[string]interface{}{
		"kubernetes_host":                   "https://kubernetes.default.svc:443",
		"kubernetes_ca_cert":                "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
		"token_reviewer_jwt":                "",
		"pem_keys":                          []string{"pem-key-1", "pem-key-2"},
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

	if !config.IsEquivalentToDesiredState(payload) {
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

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different kubernetes_host) to NOT be equivalent")
	}
}

// Extra fields cause IsEquivalentToDesiredState to return false because
// reflect.DeepEqual compares full maps (bare DeepEqual, no filtering).
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

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
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
