package v1alpha1

import (
	"testing"

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
