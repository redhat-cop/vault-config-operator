package v1alpha1

import (
	"reflect"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJWTOIDCAuthEngineConfigGetPath(t *testing.T) {
	config := &JWTOIDCAuthEngineConfig{
		Spec: JWTOIDCAuthEngineConfigSpec{
			Path: "jwt",
		},
	}

	result := config.GetPath()
	expected := "auth/jwt/config"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestJWTOIDCConfigToMap(t *testing.T) {
	providerConfig := &apiextensionsv1.JSON{Raw: []byte(`{"provider":"azure"}`)}

	config := JWTOIDCConfig{
		OIDCDiscoveryURL:     "https://accounts.google.com",
		OIDCDiscoveryCAPEM:   "ca-pem-data",
		OIDCResponseMode:     "query",
		OIDCResponseTypes:    []string{"code"},
		JWKSURL:              "https://jwks.example.com",
		JWKSCAPEM:            "jwks-ca-pem",
		JWTValidationPubKeys: []string{"pubkey1"},
		BoundIssuer:          "https://issuer.example.com",
		JWTSupportedAlgs:     []string{"RS256"},
		DefaultRole:          "default-role",
		ProviderConfig:       providerConfig,
		NamespaceInState:     true,
	}
	config.retrievedClientID = "my-client-id"
	config.retrievedClientPassword = "my-client-secret"

	result := config.toMap()

	if len(result) != 14 {
		t.Errorf("expected 14 keys in map, got %d", len(result))
	}

	expected := map[string]interface{}{
		"oidc_discovery_url":     "https://accounts.google.com",
		"oidc_discovery_ca_pem":  "ca-pem-data",
		"oidc_client_id":         "my-client-id",
		"oidc_client_secret":     "my-client-secret",
		"oidc_response_mode":     "query",
		"oidc_response_types":    []string{"code"},
		"jwks_url":               "https://jwks.example.com",
		"jwks_ca_pem":            "jwks-ca-pem",
		"jwt_validation_pubkeys": []string{"pubkey1"},
		"bound_issuer":           "https://issuer.example.com",
		"jwt_supported_algs":     []string{"RS256"},
		"default_role":           "default-role",
		"provider_config":        providerConfig,
		"namespace_in_state":     true,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestJWTOIDCConfigToMapUnexportedCredentials(t *testing.T) {
	config := JWTOIDCConfig{
		OIDCDiscoveryURL: "https://accounts.google.com",
	}
	config.retrievedClientID = "my-client-id"
	config.retrievedClientPassword = "my-client-secret"

	result := config.toMap()

	if result["oidc_client_id"] != "my-client-id" {
		t.Errorf("expected oidc_client_id 'my-client-id', got %v", result["oidc_client_id"])
	}
	if result["oidc_client_secret"] != "my-client-secret" {
		t.Errorf("expected oidc_client_secret 'my-client-secret', got %v", result["oidc_client_secret"])
	}
}

func TestJWTOIDCConfigToMapProviderConfigJSON(t *testing.T) {
	providerConfig := &apiextensionsv1.JSON{Raw: []byte(`{"provider":"google"}`)}
	config := JWTOIDCConfig{
		ProviderConfig: providerConfig,
	}

	result := config.toMap()

	val, ok := result["provider_config"].(*apiextensionsv1.JSON)
	if !ok {
		t.Fatalf("expected provider_config to be *apiextensionsv1.JSON, got %T", result["provider_config"])
	}
	if !reflect.DeepEqual(val, providerConfig) {
		t.Errorf("expected provider_config to be stored directly, got %v", val)
	}
}

func TestJWTOIDCAuthEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &JWTOIDCAuthEngineConfig{
		Spec: JWTOIDCAuthEngineConfigSpec{
			JWTOIDCConfig: JWTOIDCConfig{
				OIDCDiscoveryURL: "https://accounts.google.com",
				BoundIssuer:      "https://issuer.example.com",
				NamespaceInState: true,
			},
		},
	}

	payload := config.Spec.JWTOIDCConfig.toMap()

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestJWTOIDCAuthEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &JWTOIDCAuthEngineConfig{
		Spec: JWTOIDCAuthEngineConfigSpec{
			JWTOIDCConfig: JWTOIDCConfig{
				OIDCDiscoveryURL: "https://accounts.google.com",
			},
		},
	}

	payload := config.Spec.JWTOIDCConfig.toMap()
	payload["oidc_discovery_url"] = "https://different.example.com"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different oidc_discovery_url) to NOT be equivalent")
	}
}

func TestJWTOIDCAuthEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &JWTOIDCAuthEngineConfig{
		Spec: JWTOIDCAuthEngineConfigSpec{
			JWTOIDCConfig: JWTOIDCConfig{
				OIDCDiscoveryURL: "https://accounts.google.com",
			},
		},
	}

	payload := config.Spec.JWTOIDCConfig.toMap()
	payload["extra_field"] = "unexpected"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
	}
}

func TestJWTOIDCAuthEngineConfigIsDeletable(t *testing.T) {
	config := &JWTOIDCAuthEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected JWTOIDCAuthEngineConfig to NOT be deletable")
	}
}

func TestJWTOIDCAuthEngineConfigConditions(t *testing.T) {
	config := &JWTOIDCAuthEngineConfig{}

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
