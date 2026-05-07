package v1alpha1

import (
	"context"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestQuaySecretEngineConfigGetPath(t *testing.T) {
	config := &QuaySecretEngineConfig{
		Spec: QuaySecretEngineConfigSpec{
			Path: "quay",
		},
	}
	if got := config.GetPath(); got != "quay/config" {
		t.Errorf("GetPath() = %q, expected %q", got, "quay/config")
	}
}

func TestQuayConfigToMap(t *testing.T) {
	config := QuayConfig{
		URL:                    "https://quay.example.com",
		retrievedToken:         "my-token",
		CACertertificate:       "pem-data",
		DisableSslVerification: true,
	}

	result := config.toMap()

	if len(result) != 4 {
		t.Fatalf("expected 4 keys in toMap() output, got %d", len(result))
	}
	for _, key := range []string{"url", "token", "ca_certificate", "disable_ssl_verification"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}
	if result["url"] != "https://quay.example.com" {
		t.Errorf("url = %v, expected https://quay.example.com", result["url"])
	}
	if result["token"] != "my-token" {
		t.Errorf("token = %v, expected my-token", result["token"])
	}
	if result["ca_certificate"] != "pem-data" {
		t.Errorf("ca_certificate = %v, expected pem-data", result["ca_certificate"])
	}
	if result["disable_ssl_verification"] != true {
		t.Errorf("disable_ssl_verification = %v, expected true", result["disable_ssl_verification"])
	}
}

func TestQuaySecretEngineConfigIsEquivalentPasswordDeleted(t *testing.T) {
	config := &QuaySecretEngineConfig{
		Spec: QuaySecretEngineConfigSpec{
			Path: "quay",
			QuayConfig: QuayConfig{
				URL:                    "https://quay.example.com",
				retrievedToken:         "my-token",
				CACertertificate:       "pem-data",
				DisableSslVerification: true,
			},
		},
	}

	payload := map[string]interface{}{
		"url":                      "https://quay.example.com",
		"token":                    "my-token",
		"ca_certificate":           "pem-data",
		"disable_ssl_verification": true,
	}
	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true when payload matches all keys from toMap() (password delete is a no-op)")
	}

	payloadWithPassword := map[string]interface{}{
		"url":                      "https://quay.example.com",
		"token":                    "my-token",
		"ca_certificate":           "pem-data",
		"disable_ssl_verification": true,
		"password":                 "extra",
	}
	if config.IsEquivalentToDesiredState(payloadWithPassword) {
		t.Error("expected false when payload contains password key")
	}
}

func TestQuaySecretEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &QuaySecretEngineConfig{
		Spec: QuaySecretEngineConfigSpec{
			Path: "quay",
			QuayConfig: QuayConfig{
				URL:                    "https://quay.example.com",
				retrievedToken:         "ignored-for-equivalence",
				CACertertificate:       "pem-data",
				DisableSslVerification: false,
			},
		},
	}

	desiredState := config.Spec.QuayConfig.toMap()
	delete(desiredState, "password")

	if !config.IsEquivalentToDesiredState(desiredState) {
		t.Error("expected true for payload matching desired state after password delete")
	}
}

func TestQuaySecretEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &QuaySecretEngineConfig{
		Spec: QuaySecretEngineConfigSpec{
			Path: "quay",
			QuayConfig: QuayConfig{
				URL:                    "https://quay.example.com",
				retrievedToken:         "my-token",
				CACertertificate:       "pem-data",
				DisableSslVerification: true,
			},
		},
	}

	payload := map[string]interface{}{
		"url":                      "https://quay.example.com",
		"token":                    "my-token",
		"ca_certificate":           "other-pem",
		"disable_ssl_verification": true,
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when ca_certificate differs")
	}
}

func TestQuaySecretEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &QuaySecretEngineConfig{
		Spec: QuaySecretEngineConfigSpec{
			Path: "quay",
			QuayConfig: QuayConfig{
				URL:                    "https://quay.example.com",
				retrievedToken:         "my-token",
				CACertertificate:       "pem-data",
				DisableSslVerification: true,
			},
		},
	}

	payload := map[string]interface{}{
		"url":                      "https://quay.example.com",
		"token":                    "my-token",
		"ca_certificate":           "pem-data",
		"disable_ssl_verification": true,
		"extra_field":              "from-vault",
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when payload has extra keys")
	}
}

func TestQuaySecretEngineConfigIsDeletable(t *testing.T) {
	config := &QuaySecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected QuaySecretEngineConfig not to be deletable")
	}
}

func TestQuaySecretEngineConfigConditions(t *testing.T) {
	config := &QuaySecretEngineConfig{}

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

func TestQuaySecretEngineConfig_PrepareInternalValues(t *testing.T) {
	const ns = "test-ns"

	tests := []struct {
		name      string
		setup     func(t *testing.T) (*QuaySecretEngineConfig, context.Context)
		wantToken string
	}{
		{
			name: "secret branch token from K8s secret password key",
			setup: func(t *testing.T) (*QuaySecretEngineConfig, context.Context) {
				t.Helper()
				sec := newK8sSecret(ns, "quay-token", map[string][]byte{
					"password": []byte("k8s-quay-token"),
				})
				kube := newFakeKubeClient(sec)
				handler := newFakeVaultHandler()
				vc, ts := newFakeVaultClient(t, handler)
				t.Cleanup(ts.Close)
				instance := &QuaySecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "quay-config"},
					Spec: QuaySecretEngineConfigSpec{
						Path: "quay",
						QuayConfig: QuayConfig{
							URL: "https://quay.example.com",
						},
						RootCredentials: vaultutils.RootCredentialConfig{
							Secret:      &corev1.LocalObjectReference{Name: "quay-token"},
							PasswordKey: "password",
							UsernameKey: "username",
						},
					},
				}
				return instance, pivContext(kube, vc)
			},
			wantToken: "k8s-quay-token",
		},
		{
			name: "vault secret token from password key",
			setup: func(t *testing.T) (*QuaySecretEngineConfig, context.Context) {
				t.Helper()
				kube := newFakeKubeClient()
				handler := newFakeVaultHandler()
				handler.setGet("secret/quay-creds", map[string]interface{}{
					"password": "vault-quay-token",
				})
				vc, ts := newFakeVaultClient(t, handler)
				t.Cleanup(ts.Close)
				instance := &QuaySecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "quay-config"},
					Spec: QuaySecretEngineConfigSpec{
						Path: "quay",
						QuayConfig: QuayConfig{
							URL: "https://quay.example.com",
						},
						RootCredentials: vaultutils.RootCredentialConfig{
							VaultSecret: &vaultutils.VaultSecretReference{Path: "secret/quay-creds"},
							PasswordKey: "password",
							UsernameKey: "username",
						},
					},
				}
				return instance, pivContext(kube, vc)
			},
			wantToken: "vault-quay-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, ctx := tt.setup(t)
			if err := instance.PrepareInternalValues(ctx, instance); err != nil {
				t.Fatalf("PrepareInternalValues: %v", err)
			}
			if instance.Spec.QuayConfig.retrievedToken != tt.wantToken {
				t.Errorf("retrievedToken = %q, want %q", instance.Spec.QuayConfig.retrievedToken, tt.wantToken)
			}
		})
	}
}
