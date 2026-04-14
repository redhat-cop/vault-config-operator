package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIdentityOIDCProviderGetPath(t *testing.T) {
	tests := []struct {
		name         string
		provider     *IdentityOIDCProvider
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			provider: &IdentityOIDCProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-provider",
				},
				Spec: IdentityOIDCProviderSpec{
					Name: "custom-name",
				},
			},
			expectedPath: "identity/oidc/provider/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			provider: &IdentityOIDCProvider{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-provider",
				},
				Spec: IdentityOIDCProviderSpec{},
			},
			expectedPath: "identity/oidc/provider/test-provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.provider.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestIdentityOIDCProviderGetPayload(t *testing.T) {
	provider := &IdentityOIDCProvider{
		Spec: IdentityOIDCProviderSpec{
			IdentityOIDCProviderConfig: IdentityOIDCProviderConfig{
				Issuer:           "https://example.com",
				AllowedClientIDs: []string{"client1", "client2"},
				ScopesSupported:  []string{"openid", "profile"},
			},
		},
	}

	payload := provider.GetPayload()

	if payload["issuer"] != "https://example.com" {
		t.Errorf("expected issuer 'https://example.com', got %v", payload["issuer"])
	}
	if !reflect.DeepEqual(payload["allowed_client_ids"], []string{"client1", "client2"}) {
		t.Errorf("expected allowed_client_ids [client1 client2], got %v", payload["allowed_client_ids"])
	}
	if !reflect.DeepEqual(payload["scopes_supported"], []string{"openid", "profile"}) {
		t.Errorf("expected scopes_supported [openid profile], got %v", payload["scopes_supported"])
	}
}

func TestIdentityOIDCProviderGetPayloadWithoutIssuer(t *testing.T) {
	provider := &IdentityOIDCProvider{
		Spec: IdentityOIDCProviderSpec{
			IdentityOIDCProviderConfig: IdentityOIDCProviderConfig{
				AllowedClientIDs: []string{"*"},
			},
		},
	}

	payload := provider.GetPayload()

	if _, ok := payload["issuer"]; ok {
		t.Errorf("expected issuer to be absent from payload, got %v", payload["issuer"])
	}
}

func TestIdentityOIDCProviderIsEquivalentToDesiredState(t *testing.T) {
	provider := &IdentityOIDCProvider{
		Spec: IdentityOIDCProviderSpec{
			IdentityOIDCProviderConfig: IdentityOIDCProviderConfig{
				Issuer:           "https://example.com",
				AllowedClientIDs: []string{"client1"},
				ScopesSupported:  []string{"openid"},
			},
		},
	}

	matching := map[string]interface{}{
		"issuer":             "https://example.com",
		"allowed_client_ids": []string{"client1"},
		"scopes_supported":   []string{"openid"},
	}
	if !provider.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]interface{}{
		"issuer":             "https://other.com",
		"allowed_client_ids": []string{"client1"},
		"scopes_supported":   []string{"openid"},
	}
	if provider.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestIdentityOIDCScopeGetPath(t *testing.T) {
	tests := []struct {
		name         string
		scope        *IdentityOIDCScope
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			scope: &IdentityOIDCScope{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-scope",
				},
				Spec: IdentityOIDCScopeSpec{
					Name: "custom-name",
				},
			},
			expectedPath: "identity/oidc/scope/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			scope: &IdentityOIDCScope{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-scope",
				},
				Spec: IdentityOIDCScopeSpec{},
			},
			expectedPath: "identity/oidc/scope/test-scope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.scope.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestIdentityOIDCScopeGetPayload(t *testing.T) {
	scope := &IdentityOIDCScope{
		Spec: IdentityOIDCScopeSpec{
			IdentityOIDCScopeConfig: IdentityOIDCScopeConfig{
				Template:    `{ "groups": {{identity.entity.groups.names}} }`,
				Description: "A test scope",
			},
		},
	}

	payload := scope.GetPayload()

	if payload["template"] != `{ "groups": {{identity.entity.groups.names}} }` {
		t.Errorf("expected template value, got %v", payload["template"])
	}
	if payload["description"] != "A test scope" {
		t.Errorf("expected description 'A test scope', got %v", payload["description"])
	}
}

func TestIdentityOIDCScopeGetPayloadOmitsEmpty(t *testing.T) {
	scope := &IdentityOIDCScope{
		Spec: IdentityOIDCScopeSpec{
			IdentityOIDCScopeConfig: IdentityOIDCScopeConfig{},
		},
	}

	payload := scope.GetPayload()

	if _, ok := payload["template"]; ok {
		t.Errorf("expected template to be absent from payload")
	}
	if _, ok := payload["description"]; ok {
		t.Errorf("expected description to be absent from payload")
	}
}

func TestIdentityOIDCScopeIsEquivalentToDesiredState(t *testing.T) {
	scope := &IdentityOIDCScope{
		Spec: IdentityOIDCScopeSpec{
			IdentityOIDCScopeConfig: IdentityOIDCScopeConfig{
				Template:    "tmpl",
				Description: "desc",
			},
		},
	}

	matching := map[string]interface{}{
		"template":    "tmpl",
		"description": "desc",
	}
	if !scope.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]interface{}{
		"template":    "other",
		"description": "desc",
	}
	if scope.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestIdentityOIDCClientGetPath(t *testing.T) {
	tests := []struct {
		name         string
		client       *IdentityOIDCClient
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			client: &IdentityOIDCClient{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-client",
				},
				Spec: IdentityOIDCClientSpec{
					Name: "custom-name",
				},
			},
			expectedPath: "identity/oidc/client/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			client: &IdentityOIDCClient{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-client",
				},
				Spec: IdentityOIDCClientSpec{},
			},
			expectedPath: "identity/oidc/client/test-client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestIdentityOIDCClientGetPayload(t *testing.T) {
	c := &IdentityOIDCClient{
		Spec: IdentityOIDCClientSpec{
			IdentityOIDCClientConfig: IdentityOIDCClientConfig{
				Key:            "default",
				RedirectURIs:   []string{"https://example.com/callback"},
				Assignments:    []string{"allow_all"},
				ClientType:     "confidential",
				IDTokenTTL:     "24h",
				AccessTokenTTL: "24h",
			},
		},
	}

	payload := c.GetPayload()

	if payload["key"] != "default" {
		t.Errorf("expected key 'default', got %v", payload["key"])
	}
	if !reflect.DeepEqual(payload["redirect_uris"], []string{"https://example.com/callback"}) {
		t.Errorf("expected redirect_uris, got %v", payload["redirect_uris"])
	}
	if !reflect.DeepEqual(payload["assignments"], []string{"allow_all"}) {
		t.Errorf("expected assignments [allow_all], got %v", payload["assignments"])
	}
	if payload["client_type"] != "confidential" {
		t.Errorf("expected client_type 'confidential', got %v", payload["client_type"])
	}
	if payload["id_token_ttl"] != "24h" {
		t.Errorf("expected id_token_ttl '24h', got %v", payload["id_token_ttl"])
	}
	if payload["access_token_ttl"] != "24h" {
		t.Errorf("expected access_token_ttl '24h', got %v", payload["access_token_ttl"])
	}
}

func TestIdentityOIDCClientIsEquivalentToDesiredState(t *testing.T) {
	c := &IdentityOIDCClient{
		Spec: IdentityOIDCClientSpec{
			IdentityOIDCClientConfig: IdentityOIDCClientConfig{
				Key:            "default",
				RedirectURIs:   []string{"https://example.com/callback"},
				Assignments:    []string{"allow_all"},
				ClientType:     "confidential",
				IDTokenTTL:     "24h",
				AccessTokenTTL: "24h",
			},
		},
	}

	matching := map[string]interface{}{
		"key":              "default",
		"redirect_uris":    []string{"https://example.com/callback"},
		"assignments":      []string{"allow_all"},
		"client_type":      "confidential",
		"id_token_ttl":     "24h",
		"access_token_ttl": "24h",
	}
	if !c.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]interface{}{
		"key":              "other-key",
		"redirect_uris":    []string{"https://example.com/callback"},
		"assignments":      []string{"allow_all"},
		"client_type":      "confidential",
		"id_token_ttl":     "24h",
		"access_token_ttl": "24h",
	}
	if c.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestIdentityOIDCAssignmentGetPath(t *testing.T) {
	tests := []struct {
		name         string
		assignment   *IdentityOIDCAssignment
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			assignment: &IdentityOIDCAssignment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-assignment",
				},
				Spec: IdentityOIDCAssignmentSpec{
					Name: "custom-name",
				},
			},
			expectedPath: "identity/oidc/assignment/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			assignment: &IdentityOIDCAssignment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-assignment",
				},
				Spec: IdentityOIDCAssignmentSpec{},
			},
			expectedPath: "identity/oidc/assignment/test-assignment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.assignment.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestIdentityOIDCAssignmentGetPayload(t *testing.T) {
	a := &IdentityOIDCAssignment{
		Spec: IdentityOIDCAssignmentSpec{
			IdentityOIDCAssignmentConfig: IdentityOIDCAssignmentConfig{
				EntityIDs: []string{"entity-1", "entity-2"},
				GroupIDs:  []string{"group-1"},
			},
		},
	}

	payload := a.GetPayload()

	if !reflect.DeepEqual(payload["entity_ids"], []string{"entity-1", "entity-2"}) {
		t.Errorf("expected entity_ids [entity-1 entity-2], got %v", payload["entity_ids"])
	}
	if !reflect.DeepEqual(payload["group_ids"], []string{"group-1"}) {
		t.Errorf("expected group_ids [group-1], got %v", payload["group_ids"])
	}
}

func TestIdentityOIDCAssignmentIsEquivalentToDesiredState(t *testing.T) {
	a := &IdentityOIDCAssignment{
		Spec: IdentityOIDCAssignmentSpec{
			IdentityOIDCAssignmentConfig: IdentityOIDCAssignmentConfig{
				EntityIDs: []string{"entity-1"},
				GroupIDs:  []string{"group-1"},
			},
		},
	}

	matching := map[string]interface{}{
		"entity_ids": []string{"entity-1"},
		"group_ids":  []string{"group-1"},
	}
	if !a.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]interface{}{
		"entity_ids": []string{"entity-2"},
		"group_ids":  []string{"group-1"},
	}
	if a.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

// Tests for AC #4: extra-field handling in IsEquivalentToDesiredState.
// These types use reflect.DeepEqual(desiredState, payload), so extra keys in
// the payload cause the comparison to return false. In production the
// reconciler calls IsEquivalentToDesiredState with the raw Vault read
// response (no key filtering), meaning extra Vault-returned fields trigger
// an unnecessary write. Story 7-4 tracks hardening this behavior.

func TestIdentityOIDCScopeIsEquivalentExtraFieldsReturnsFalse(t *testing.T) {
	scope := &IdentityOIDCScope{
		Spec: IdentityOIDCScopeSpec{
			IdentityOIDCScopeConfig: IdentityOIDCScopeConfig{
				Template:    "tmpl",
				Description: "desc",
			},
		},
	}

	payloadWithExtra := map[string]interface{}{
		"template":    "tmpl",
		"description": "desc",
		"extra_field": "vault-returned-value",
	}
	if scope.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected payload with extra fields to NOT be equivalent (reflect.DeepEqual compares full maps)")
	}
}

func TestIdentityOIDCProviderIsEquivalentExtraFieldsReturnsFalse(t *testing.T) {
	provider := &IdentityOIDCProvider{
		Spec: IdentityOIDCProviderSpec{
			IdentityOIDCProviderConfig: IdentityOIDCProviderConfig{
				Issuer:           "https://example.com",
				AllowedClientIDs: []string{"client1"},
				ScopesSupported:  []string{"openid"},
			},
		},
	}

	payloadWithExtra := map[string]interface{}{
		"issuer":             "https://example.com",
		"allowed_client_ids": []string{"client1"},
		"scopes_supported":   []string{"openid"},
		"extra_field":        "vault-returned-value",
	}
	if provider.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected payload with extra fields to NOT be equivalent (reflect.DeepEqual compares full maps)")
	}
}

func TestIdentityOIDCClientIsEquivalentExtraFieldsReturnsFalse(t *testing.T) {
	c := &IdentityOIDCClient{
		Spec: IdentityOIDCClientSpec{
			IdentityOIDCClientConfig: IdentityOIDCClientConfig{
				Key:            "default",
				RedirectURIs:   []string{"https://example.com/callback"},
				Assignments:    []string{"allow_all"},
				ClientType:     "confidential",
				IDTokenTTL:     "24h",
				AccessTokenTTL: "24h",
			},
		},
	}

	payloadWithExtra := map[string]interface{}{
		"key":              "default",
		"redirect_uris":    []string{"https://example.com/callback"},
		"assignments":      []string{"allow_all"},
		"client_type":      "confidential",
		"id_token_ttl":     "24h",
		"access_token_ttl": "24h",
		"client_id":        "generated-id-from-vault",
		"client_secret":    "generated-secret-from-vault",
	}
	if c.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected payload with extra fields to NOT be equivalent (reflect.DeepEqual compares full maps)")
	}
}

func TestIdentityOIDCAssignmentIsEquivalentExtraFieldsReturnsFalse(t *testing.T) {
	a := &IdentityOIDCAssignment{
		Spec: IdentityOIDCAssignmentSpec{
			IdentityOIDCAssignmentConfig: IdentityOIDCAssignmentConfig{
				EntityIDs: []string{"entity-1"},
				GroupIDs:  []string{"group-1"},
			},
		},
	}

	payloadWithExtra := map[string]interface{}{
		"entity_ids":  []string{"entity-1"},
		"group_ids":   []string{"group-1"},
		"extra_field": "vault-returned-value",
	}
	if a.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected payload with extra fields to NOT be equivalent (reflect.DeepEqual compares full maps)")
	}
}

func TestIdentityOIDCIsDeletable(t *testing.T) {
	provider := &IdentityOIDCProvider{}
	if !provider.IsDeletable() {
		t.Error("expected IdentityOIDCProvider to be deletable")
	}

	scope := &IdentityOIDCScope{}
	if !scope.IsDeletable() {
		t.Error("expected IdentityOIDCScope to be deletable")
	}

	client := &IdentityOIDCClient{}
	if !client.IsDeletable() {
		t.Error("expected IdentityOIDCClient to be deletable")
	}

	assignment := &IdentityOIDCAssignment{}
	if !assignment.IsDeletable() {
		t.Error("expected IdentityOIDCAssignment to be deletable")
	}
}

func TestIdentityOIDCConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	provider := &IdentityOIDCProvider{}
	provider.SetConditions([]metav1.Condition{condition})
	if len(provider.GetConditions()) != 1 || provider.GetConditions()[0].Type != "Ready" {
		t.Error("expected IdentityOIDCProvider conditions to be set and retrieved")
	}

	scope := &IdentityOIDCScope{}
	scope.SetConditions([]metav1.Condition{condition})
	if len(scope.GetConditions()) != 1 || scope.GetConditions()[0].Type != "Ready" {
		t.Error("expected IdentityOIDCScope conditions to be set and retrieved")
	}

	client := &IdentityOIDCClient{}
	client.SetConditions([]metav1.Condition{condition})
	if len(client.GetConditions()) != 1 || client.GetConditions()[0].Type != "Ready" {
		t.Error("expected IdentityOIDCClient conditions to be set and retrieved")
	}

	assignment := &IdentityOIDCAssignment{}
	assignment.SetConditions([]metav1.Condition{condition})
	if len(assignment.GetConditions()) != 1 || assignment.GetConditions()[0].Type != "Ready" {
		t.Error("expected IdentityOIDCAssignment conditions to be set and retrieved")
	}
}
