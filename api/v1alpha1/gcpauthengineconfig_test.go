package v1alpha1

import (
	"reflect"
	"testing"

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

	expected := map[string]interface{}{
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

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
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
