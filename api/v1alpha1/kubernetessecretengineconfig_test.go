package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesSecretEngineConfigGetPath(t *testing.T) {
	config := &KubernetesSecretEngineConfig{
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
		},
	}
	if got := config.GetPath(); got != "kubernetes/config" {
		t.Errorf("GetPath() = %q, expected %q", got, "kubernetes/config")
	}
}

func TestKubeSEConfigToMap(t *testing.T) {
	config := KubeSEConfig{
		KubernetesHost:             "https://k8s.example.com",
		KubernetesCACert:           "pem-ca-data",
		retrievedServiceAccountJWT: "jwt-token-123",
		DisableLocalCAJWT:          true,
	}

	result := config.toMap()

	if len(result) != 4 {
		t.Fatalf("expected 4 keys in toMap() output, got %d", len(result))
	}
	for _, key := range []string{"kubernetes_host", "kubernetes_ca_cert", "service_account_jwt", "disable_local_ca_jwt"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}
	if result["kubernetes_host"] != "https://k8s.example.com" {
		t.Errorf("kubernetes_host = %v", result["kubernetes_host"])
	}
	if result["kubernetes_ca_cert"] != "pem-ca-data" {
		t.Errorf("kubernetes_ca_cert = %v", result["kubernetes_ca_cert"])
	}
	if result["service_account_jwt"] != "jwt-token-123" {
		t.Errorf("service_account_jwt = %v", result["service_account_jwt"])
	}
	if result["disable_local_ca_jwt"] != true {
		t.Errorf("disable_local_ca_jwt = %v", result["disable_local_ca_jwt"])
	}
}

func TestKubernetesSecretEngineConfigIsEquivalentJWTDeleted(t *testing.T) {
	config := &KubernetesSecretEngineConfig{
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
			KubeSEConfig: KubeSEConfig{
				KubernetesHost:             "https://k8s.example.com",
				KubernetesCACert:           "pem-ca-data",
				retrievedServiceAccountJWT: "jwt-token-123",
				DisableLocalCAJWT:          true,
			},
		},
	}

	payload := map[string]interface{}{
		"kubernetes_host":      "https://k8s.example.com",
		"kubernetes_ca_cert":   "pem-ca-data",
		"disable_local_ca_jwt": true,
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true when payload omits service_account_jwt but matches other fields after JWT delete")
	}
}

func TestKubernetesSecretEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &KubernetesSecretEngineConfig{
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
			KubeSEConfig: KubeSEConfig{
				KubernetesHost:             "https://k8s.example.com",
				KubernetesCACert:           "pem-ca-data",
				retrievedServiceAccountJWT: "ignored-for-equivalence",
				DisableLocalCAJWT:          false,
			},
		},
	}

	desiredState := config.Spec.KubeSEConfig.toMap()
	delete(desiredState, "service_account_jwt")

	if !config.IsEquivalentToDesiredState(desiredState) {
		t.Error("expected true for payload matching desired state without service_account_jwt")
	}
}

func TestKubernetesSecretEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &KubernetesSecretEngineConfig{
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
			KubeSEConfig: KubeSEConfig{
				KubernetesHost:             "https://k8s.example.com",
				KubernetesCACert:           "pem-ca-data",
				retrievedServiceAccountJWT: "jwt-token-123",
				DisableLocalCAJWT:          true,
			},
		},
	}

	payload := map[string]interface{}{
		"kubernetes_host":      "https://k8s.other.example.com",
		"kubernetes_ca_cert":   "pem-ca-data",
		"disable_local_ca_jwt": true,
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when kubernetes_host differs")
	}
}

func TestKubernetesSecretEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &KubernetesSecretEngineConfig{
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
			KubeSEConfig: KubeSEConfig{
				KubernetesHost:             "https://k8s.example.com",
				KubernetesCACert:           "pem-ca-data",
				retrievedServiceAccountJWT: "jwt-token-123",
				DisableLocalCAJWT:          true,
			},
		},
	}

	payload := map[string]interface{}{
		"kubernetes_host":      "https://k8s.example.com",
		"kubernetes_ca_cert":   "pem-ca-data",
		"disable_local_ca_jwt": true,
		"extra_field":          "from-vault",
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when payload has extra keys")
	}
}

func TestKubernetesSecretEngineConfigIsEquivalentPayloadWithJWT(t *testing.T) {
	config := &KubernetesSecretEngineConfig{
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
			KubeSEConfig: KubeSEConfig{
				KubernetesHost:             "https://k8s.example.com",
				KubernetesCACert:           "pem-ca-data",
				retrievedServiceAccountJWT: "jwt-token-123",
				DisableLocalCAJWT:          true,
			},
		},
	}

	payload := map[string]interface{}{
		"kubernetes_host":      "https://k8s.example.com",
		"kubernetes_ca_cert":   "pem-ca-data",
		"disable_local_ca_jwt": true,
		"service_account_jwt":  "jwt-token-123",
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false: desired state deletes service_account_jwt, so payload still containing it has an extra key")
	}
}

func TestKubernetesSecretEngineConfigIsDeletable(t *testing.T) {
	config := &KubernetesSecretEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected KubernetesSecretEngineConfig to be deletable")
	}
}

func TestKubernetesSecretEngineConfigConditions(t *testing.T) {
	config := &KubernetesSecretEngineConfig{}

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
