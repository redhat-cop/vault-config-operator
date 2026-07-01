package v1alpha1

import (
	"reflect"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGCPAuthEngineConfigGetPath(t *testing.T) {
	config := &GCPAuthEngineConfig{
		Spec: GCPAuthEngineConfigSpec{
			Path: "gcp",
		},
	}

	result := config.GetPath()
	expected := "auth/gcp/config"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestGCPConfigToMap(t *testing.T) {
	customEndpoint := &apiextensionsv1.JSON{Raw: []byte(`{"api":"https://private.googleapis.com"}`)}

	config := GCPConfig{
		IAMalias:       "unique_id",
		IAMmetadata:    "default",
		GCEalias:       "role_id",
		GCEmetadata:    "default",
		CustomEndpoint: customEndpoint,
	}
	config.retrievedCredentials = `{"type":"service_account","project_id":"my-project"}`

	result := config.toMap()

	if len(result) != 6 {
		t.Errorf("expected 6 keys in map, got %d", len(result))
	}

	expected := map[string]any{
		"credentials":     `{"type":"service_account","project_id":"my-project"}`,
		"iam_alias":       "unique_id",
		"iam_metadata":    "default",
		"gce_alias":       "role_id",
		"gce_metadata":    "default",
		"custom_endpoint": customEndpoint,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestGCPConfigToMapUnexportedCredentials(t *testing.T) {
	config := GCPConfig{}
	config.retrievedCredentials = "my-gcp-credentials-json"

	result := config.toMap()

	if result["credentials"] != "my-gcp-credentials-json" {
		t.Errorf("expected credentials from retrievedCredentials, got %v", result["credentials"])
	}
}

func TestGCPConfigToMapCustomEndpointJSON(t *testing.T) {
	customEndpoint := &apiextensionsv1.JSON{Raw: []byte(`{"api":"https://api.example.com"}`)}
	config := GCPConfig{
		CustomEndpoint: customEndpoint,
	}

	result := config.toMap()

	val, ok := result["custom_endpoint"].(*apiextensionsv1.JSON)
	if !ok {
		t.Fatalf("expected custom_endpoint to be *apiextensionsv1.JSON, got %T", result["custom_endpoint"])
	}
	if !reflect.DeepEqual(val, customEndpoint) {
		t.Errorf("expected custom_endpoint to be stored directly")
	}
}

func TestGCPAuthEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &GCPAuthEngineConfig{
		Spec: GCPAuthEngineConfigSpec{
			GCPConfig: GCPConfig{
				IAMalias:    "unique_id",
				IAMmetadata: "default",
				GCEalias:    "role_id",
				GCEmetadata: "default",
			},
		},
	}

	payload := config.Spec.GCPConfig.toMap()

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestGCPAuthEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &GCPAuthEngineConfig{
		Spec: GCPAuthEngineConfigSpec{
			GCPConfig: GCPConfig{
				IAMalias: "unique_id",
			},
		},
	}

	payload := config.Spec.GCPConfig.toMap()
	payload["iam_alias"] = "role_id"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different iam_alias) to NOT be equivalent")
	}
}

func TestGCPAuthEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &GCPAuthEngineConfig{
		Spec: GCPAuthEngineConfigSpec{
			GCPConfig: GCPConfig{
				IAMalias: "unique_id",
			},
		},
	}

	payload := config.Spec.GCPConfig.toMap()
	payload["extra_field"] = "unexpected"

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestGCPAuthEngineConfigIsDeletable(t *testing.T) {
	config := &GCPAuthEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected GCPAuthEngineConfig to NOT be deletable")
	}
}

func TestGCPAuthEngineConfigConditions(t *testing.T) {
	config := &GCPAuthEngineConfig{}

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

func TestGCPAuthEngineConfig_PrepareInternalValues_DefaultNoOp(t *testing.T) {
	defaultCreds := vaultutils.RootCredentialConfig{
		UsernameKey: "serviceaccount",
		PasswordKey: "credentials",
	}
	config := &GCPAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: "ns-gcp-auth"},
		Spec: GCPAuthEngineConfigSpec{
			GCPCredentials: defaultCreds,
			GCPConfig: GCPConfig{
				IAMalias:    "unique_id",
				IAMmetadata: "default",
			},
		},
	}
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedServiceAccount != "" || config.Spec.retrievedCredentials != "" {
		t.Errorf("expected empty retrieved GCP credentials, got sa=%q creds=%q",
			config.Spec.retrievedServiceAccount, config.Spec.retrievedCredentials)
	}
}

func TestGCPAuthEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-gcp-auth"
	sec := newK8sSecret(ns, "gcp-creds", map[string][]byte{
		"serviceaccount": []byte("sa@proj.iam.gserviceaccount.com"),
		"credentials":    []byte(`{"type":"service_account"}`),
	})
	kube := newFakeKubeClient(sec)
	vc, ts := newFakeVaultClient(t, newFakeVaultHandler())
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &GCPAuthEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: GCPAuthEngineConfigSpec{
			GCPCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "gcp-creds"},
				UsernameKey: "serviceaccount",
				PasswordKey: "credentials",
			},
			GCPConfig: GCPConfig{
				ServiceAccount: "",
				IAMalias:       "unique_id",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedServiceAccount != "sa@proj.iam.gserviceaccount.com" {
		t.Errorf("retrievedServiceAccount = %q", config.Spec.retrievedServiceAccount)
	}
	if config.Spec.retrievedCredentials != `{"type":"service_account"}` {
		t.Errorf("retrievedCredentials = %q", config.Spec.retrievedCredentials)
	}
}
