package v1alpha1

import (
	"reflect"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestJWTOIDCAuthEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *JWTOIDCAuthEngineRole
		expectedPath string
	}{
		{
			name: "with spec.Name (embedded JWTOIDCRole.Name)",
			role: &JWTOIDCAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: JWTOIDCAuthEngineRoleSpec{
					Path: "jwt",
					JWTOIDCRole: JWTOIDCRole{
						Name: "custom-role",
					},
				},
			},
			expectedPath: "auth/jwt/role/custom-role",
		},
		{
			name: "without spec.Name falls back to metadata.name",
			role: &JWTOIDCAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: JWTOIDCAuthEngineRoleSpec{
					Path: "jwt",
				},
			},
			expectedPath: "auth/jwt/role/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestJWTOIDCRoleToMap(t *testing.T) {
	boundClaims := &apiextensionsv1.JSON{Raw: []byte(`{"groups":"admin"}`)}
	claimMappings := map[string]string{"email": "email", "name": "name"}

	role := JWTOIDCRole{
		Name:                 "test-role",
		RoleType:             "oidc",
		BoundAudiences:       []string{"aud1", "aud2"},
		UserClaim:            "sub",
		UserClaimJSONPointer: false,
		ClockSkewLeeway:      60,
		ExpirationLeeway:     150,
		NotBeforeLeeway:      150,
		BoundSubject:         "subject",
		BoundClaims:          boundClaims,
		BoundClaimsType:      "string",
		GroupsClaim:          "groups",
		ClaimMappings:        claimMappings,
		OIDCScopes:           []string{"profile", "email"},
		AllowedRedirectURIs:  []string{"https://vault.example.com/callback"},
		VerboseOIDCLogging:   false,
		MaxAge:               0,
		TokenTTL:             "1h",
		TokenMaxTTL:          "24h",
		TokenPolicies:        []string{"reader"},
		TokenBoundCIDRs:      []string{"10.0.0.0/8"},
		TokenExplicitMaxTTL:  "",
		TokenNoDefaultPolicy: false,
		TokenNumUses:         0,
		TokenPeriod:          0,
		TokenType:            "service",
	}

	result := role.toMap()

	if len(result) != 26 {
		t.Errorf("expected 26 keys in map, got %d", len(result))
	}

	expected := map[string]interface{}{
		"name":                    "test-role",
		"role_type":               "oidc",
		"bound_audiences":         []string{"aud1", "aud2"},
		"user_claim":              "sub",
		"user_claim_json_pointer": false,
		"clock_skew_leeway":       int64(60),
		"expiration_leeway":       int64(150),
		"not_before_leeway":       int64(150),
		"bound_subject":           "subject",
		"bound_claims":            boundClaims,
		"bound_claims_type":       "string",
		"groups_claim":            "groups",
		"claim_mappings":          claimMappings,
		"oidc_scopes":             []string{"profile", "email"},
		"allowed_redirect_uris":   []string{"https://vault.example.com/callback"},
		"verbose_oidc_logging":    false,
		"max_age":                 int64(0),
		"token_ttl":               "1h",
		"token_max_ttl":           "24h",
		"token_policies":          []string{"reader"},
		"token_bound_cidrs":       []string{"10.0.0.0/8"},
		"token_explicit_max_ttl":  "",
		"token_no_default_policy": false,
		"token_num_uses":          int64(0),
		"token_period":            int64(0),
		"token_type":              "service",
	}

	if !reflect.DeepEqual(result, expected) {
		for k, v := range expected {
			if !reflect.DeepEqual(result[k], v) {
				t.Errorf("key %q: got %v (%T), want %v (%T)", k, result[k], result[k], v, v)
			}
		}
	}
}

func TestJWTOIDCRoleToMapBoundClaimsJSON(t *testing.T) {
	boundClaims := &apiextensionsv1.JSON{Raw: []byte(`{"groups":"admin"}`)}
	role := JWTOIDCRole{
		Name:        "test-role",
		UserClaim:   "sub",
		BoundClaims: boundClaims,
	}

	result := role.toMap()

	val, ok := result["bound_claims"].(*apiextensionsv1.JSON)
	if !ok {
		t.Fatalf("expected bound_claims to be *apiextensionsv1.JSON, got %T", result["bound_claims"])
	}
	if !reflect.DeepEqual(val, boundClaims) {
		t.Errorf("expected bound_claims to be stored directly")
	}
}

func TestJWTOIDCRoleToMapClaimMappings(t *testing.T) {
	claimMappings := map[string]string{"email": "email", "name": "name"}
	role := JWTOIDCRole{
		Name:          "test-role",
		UserClaim:     "sub",
		ClaimMappings: claimMappings,
	}

	result := role.toMap()

	val, ok := result["claim_mappings"].(map[string]string)
	if !ok {
		t.Fatalf("expected claim_mappings to be map[string]string, got %T", result["claim_mappings"])
	}
	if !reflect.DeepEqual(val, claimMappings) {
		t.Errorf("expected claim_mappings to be stored directly, got %v", val)
	}
}

func TestJWTOIDCAuthEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &JWTOIDCAuthEngineRole{
		Spec: JWTOIDCAuthEngineRoleSpec{
			JWTOIDCRole: JWTOIDCRole{
				Name:      "test-role",
				UserClaim: "sub",
				TokenType: "service",
			},
		},
	}

	payload := role.Spec.JWTOIDCRole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestJWTOIDCAuthEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &JWTOIDCAuthEngineRole{
		Spec: JWTOIDCAuthEngineRoleSpec{
			JWTOIDCRole: JWTOIDCRole{
				Name:      "test-role",
				UserClaim: "sub",
			},
		},
	}

	payload := role.Spec.JWTOIDCRole.toMap()
	payload["user_claim"] = "email"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different user_claim) to NOT be equivalent")
	}
}

func TestJWTOIDCAuthEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &JWTOIDCAuthEngineRole{
		Spec: JWTOIDCAuthEngineRoleSpec{
			JWTOIDCRole: JWTOIDCRole{
				Name:      "test-role",
				UserClaim: "sub",
			},
		},
	}

	payload := role.Spec.JWTOIDCRole.toMap()
	payload["extra_field"] = "unexpected"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
	}
}

func TestJWTOIDCAuthEngineRoleIsDeletable(t *testing.T) {
	role := &JWTOIDCAuthEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected JWTOIDCAuthEngineRole to be deletable")
	}
}

func TestJWTOIDCAuthEngineRoleConditions(t *testing.T) {
	role := &JWTOIDCAuthEngineRole{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	role.SetConditions(conditions)
	got := role.GetConditions()

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
