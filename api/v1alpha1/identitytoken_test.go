package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIdentityTokenConfigGetPath(t *testing.T) {
	config := &IdentityTokenConfig{}
	expected := "identity/oidc/config"
	if result := config.GetPath(); result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestIdentityTokenConfigGetPayload(t *testing.T) {
	config := &IdentityTokenConfig{
		Spec: IdentityTokenConfigSpec{
			IdentityTokenConfigConfig: IdentityTokenConfigConfig{
				Issuer: "https://example.com:1234",
			},
		},
	}

	payload := config.GetPayload()
	if payload["issuer"] != "https://example.com:1234" {
		t.Errorf("expected issuer 'https://example.com:1234', got %v", payload["issuer"])
	}
}

func TestIdentityTokenConfigGetPayloadEmptyIssuer(t *testing.T) {
	config := &IdentityTokenConfig{
		Spec: IdentityTokenConfigSpec{
			IdentityTokenConfigConfig: IdentityTokenConfigConfig{},
		},
	}

	payload := config.GetPayload()
	if payload["issuer"] != "" {
		t.Errorf("expected empty issuer, got %v", payload["issuer"])
	}
}

func TestIdentityTokenConfigIsNotDeletable(t *testing.T) {
	config := &IdentityTokenConfig{}
	if config.IsDeletable() {
		t.Error("expected IdentityTokenConfig to not be deletable")
	}
}

func TestIdentityTokenConfigIsEquivalentToDesiredState(t *testing.T) {
	config := &IdentityTokenConfig{
		Spec: IdentityTokenConfigSpec{
			IdentityTokenConfigConfig: IdentityTokenConfigConfig{
				Issuer: "https://example.com",
			},
		},
	}

	matching := map[string]interface{}{"issuer": "https://example.com"}
	if !config.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]interface{}{"issuer": "https://other.com"}
	if config.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestIdentityTokenKeyGetPath(t *testing.T) {
	tests := []struct {
		name         string
		key          *IdentityTokenKey
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			key: &IdentityTokenKey{
				ObjectMeta: metav1.ObjectMeta{Name: "test-key"},
				Spec:       IdentityTokenKeySpec{Name: "custom-name"},
			},
			expectedPath: "identity/oidc/key/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			key: &IdentityTokenKey{
				ObjectMeta: metav1.ObjectMeta{Name: "test-key"},
				Spec:       IdentityTokenKeySpec{},
			},
			expectedPath: "identity/oidc/key/test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.key.GetPath(); result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestIdentityTokenKeyGetPayload(t *testing.T) {
	key := &IdentityTokenKey{
		Spec: IdentityTokenKeySpec{
			IdentityTokenKeyConfig: IdentityTokenKeyConfig{
				RotationPeriod:   "12h",
				VerificationTTL:  "24h",
				AllowedClientIDs: []string{"*"},
				Algorithm:        "RS256",
			},
		},
	}

	payload := key.GetPayload()

	if payload["rotation_period"] != "12h" {
		t.Errorf("expected rotation_period '12h', got %v", payload["rotation_period"])
	}
	if payload["verification_ttl"] != "24h" {
		t.Errorf("expected verification_ttl '24h', got %v", payload["verification_ttl"])
	}
	if !reflect.DeepEqual(payload["allowed_client_ids"], []string{"*"}) {
		t.Errorf("expected allowed_client_ids [*], got %v", payload["allowed_client_ids"])
	}
	if payload["algorithm"] != "RS256" {
		t.Errorf("expected algorithm 'RS256', got %v", payload["algorithm"])
	}
}

func TestIdentityTokenKeyIsEquivalentToDesiredState(t *testing.T) {
	key := &IdentityTokenKey{
		Spec: IdentityTokenKeySpec{
			IdentityTokenKeyConfig: IdentityTokenKeyConfig{
				RotationPeriod:   "24h",
				VerificationTTL:  "24h",
				AllowedClientIDs: []string{"*"},
				Algorithm:        "RS256",
			},
		},
	}

	matching := map[string]interface{}{
		"rotation_period":    "24h",
		"verification_ttl":   "24h",
		"allowed_client_ids": []string{"*"},
		"algorithm":          "RS256",
	}
	if !key.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]interface{}{
		"rotation_period":    "12h",
		"verification_ttl":   "24h",
		"allowed_client_ids": []string{"*"},
		"algorithm":          "RS256",
	}
	if key.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestIdentityTokenRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *IdentityTokenRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &IdentityTokenRole{
				ObjectMeta: metav1.ObjectMeta{Name: "test-role"},
				Spec:       IdentityTokenRoleSpec{Name: "custom-name"},
			},
			expectedPath: "identity/oidc/role/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			role: &IdentityTokenRole{
				ObjectMeta: metav1.ObjectMeta{Name: "test-role"},
				Spec:       IdentityTokenRoleSpec{},
			},
			expectedPath: "identity/oidc/role/test-role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.role.GetPath(); result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestIdentityTokenRoleGetPayload(t *testing.T) {
	role := &IdentityTokenRole{
		Spec: IdentityTokenRoleSpec{
			IdentityTokenRoleConfig: IdentityTokenRoleConfig{
				Key:      "named-key-001",
				Template: `{"groups": {{identity.entity.groups.names}}}`,
				ClientID: "my-client-id",
				TTL:      "12h",
			},
		},
	}

	payload := role.GetPayload()

	if payload["key"] != "named-key-001" {
		t.Errorf("expected key 'named-key-001', got %v", payload["key"])
	}
	if payload["template"] != `{"groups": {{identity.entity.groups.names}}}` {
		t.Errorf("expected template value, got %v", payload["template"])
	}
	if payload["client_id"] != "my-client-id" {
		t.Errorf("expected client_id 'my-client-id', got %v", payload["client_id"])
	}
	if payload["ttl"] != "12h" {
		t.Errorf("expected ttl '12h', got %v", payload["ttl"])
	}
}

func TestIdentityTokenRoleGetPayloadOmitsOptional(t *testing.T) {
	role := &IdentityTokenRole{
		Spec: IdentityTokenRoleSpec{
			IdentityTokenRoleConfig: IdentityTokenRoleConfig{
				Key: "named-key-001",
				TTL: "24h",
			},
		},
	}

	payload := role.GetPayload()

	if _, ok := payload["template"]; ok {
		t.Errorf("expected template to be absent from payload")
	}
	if _, ok := payload["client_id"]; ok {
		t.Errorf("expected client_id to be absent from payload")
	}
}

func TestIdentityTokenRoleIsEquivalentToDesiredState(t *testing.T) {
	role := &IdentityTokenRole{
		Spec: IdentityTokenRoleSpec{
			IdentityTokenRoleConfig: IdentityTokenRoleConfig{
				Key: "key-001",
				TTL: "24h",
			},
		},
	}

	matching := map[string]interface{}{
		"key": "key-001",
		"ttl": "24h",
	}
	if !role.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]interface{}{
		"key": "key-002",
		"ttl": "24h",
	}
	if role.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestIdentityTokenIsDeletable(t *testing.T) {
	key := &IdentityTokenKey{}
	if !key.IsDeletable() {
		t.Error("expected IdentityTokenKey to be deletable")
	}

	role := &IdentityTokenRole{}
	if !role.IsDeletable() {
		t.Error("expected IdentityTokenRole to be deletable")
	}
}

func TestIdentityTokenConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	config := &IdentityTokenConfig{}
	config.SetConditions([]metav1.Condition{condition})
	if len(config.GetConditions()) != 1 || config.GetConditions()[0].Type != "Ready" {
		t.Error("expected IdentityTokenConfig conditions to be set and retrieved")
	}

	key := &IdentityTokenKey{}
	key.SetConditions([]metav1.Condition{condition})
	if len(key.GetConditions()) != 1 || key.GetConditions()[0].Type != "Ready" {
		t.Error("expected IdentityTokenKey conditions to be set and retrieved")
	}

	role := &IdentityTokenRole{}
	role.SetConditions([]metav1.Condition{condition})
	if len(role.GetConditions()) != 1 || role.GetConditions()[0].Type != "Ready" {
		t.Error("expected IdentityTokenRole conditions to be set and retrieved")
	}
}
