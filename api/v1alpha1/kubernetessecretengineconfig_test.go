package v1alpha1

import (
	"strings"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
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

func TestKubernetesSecretEngineConfig_PrepareInternalValues_FromServiceAccountTokenSecret(t *testing.T) {
	ns := "ns-kube-se"
	sec := newTypedK8sSecret(ns, "sa-jwt", corev1.SecretTypeServiceAccountToken, map[string][]byte{
		corev1.ServiceAccountTokenKey: []byte("eyJhbGciOiJSUzI1NiJ9.substance"),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &KubernetesSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
			JWTReference: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "sa-jwt"},
			},
			KubeSEConfig: KubeSEConfig{
				KubernetesHost: "https://kubernetes.default",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	want := "eyJhbGciOiJSUzI1NiJ9.substance"
	if config.Spec.retrievedServiceAccountJWT != want {
		t.Errorf("retrievedServiceAccountJWT = %q, want %q", config.Spec.retrievedServiceAccountJWT, want)
	}
}

func TestKubernetesSecretEngineConfig_PrepareInternalValues_WrongSecretType(t *testing.T) {
	ns := "ns-kube-se"
	sec := newTypedK8sSecret(ns, "bad-jwt-secret", corev1.SecretTypeOpaque, map[string][]byte{
		corev1.ServiceAccountTokenKey: []byte("x"),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &KubernetesSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
			JWTReference: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "bad-jwt-secret"},
			},
			KubeSEConfig: KubeSEConfig{
				KubernetesHost: "https://kubernetes.default",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error for wrong secret type")
	}
	if !strings.Contains(err.Error(), "secret must be of type") {
		t.Errorf("error = %q, want substring 'secret must be of type'", err.Error())
	}
}

func TestKubernetesSecretEngineConfig_PrepareInternalValues_FromVaultSecret(t *testing.T) {
	ns := "ns-kube-se"
	vaultPath := "secret/data/k8s-jwt"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]interface{}{
		"key": "jwt-from-vault-path",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &KubernetesSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: KubernetesSecretEngineConfigSpec{
			Path: "kubernetes",
			JWTReference: vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
			},
			KubeSEConfig: KubeSEConfig{
				KubernetesHost: "https://kubernetes.default",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedServiceAccountJWT != "jwt-from-vault-path" {
		t.Errorf("retrievedServiceAccountJWT = %q", config.Spec.retrievedServiceAccountJWT)
	}
}
