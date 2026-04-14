package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCertAuthEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *CertAuthEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &CertAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: CertAuthEngineRoleSpec{
					Path: "cert",
					Name: "custom-name",
				},
			},
			expectedPath: "auth/cert/certs/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &CertAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: CertAuthEngineRoleSpec{
					Path: "cert",
				},
			},
			expectedPath: "auth/cert/certs/meta-name",
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

func TestCertAuthEngineRoleInternalToMap(t *testing.T) {
	role := CertAuthEngineRoleInternal{
		Certificate:                "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
		AllowedCommonNames:         []string{"*.example.com"},
		AllowedDNSSANs:             []string{"*.example.com"},
		AllowedEmailSANs:           []string{"admin@example.com"},
		AllowedURISANs:             []string{"spiffe://cluster.local/*"},
		AllowedOrganizationalUnits: []string{"Engineering"},
		RequiredExtensions:         []string{"1.2.3.4.5:value"},
		AllowedMetadataExtensions:  []string{"1.2.3.4.5"},
		OCSPEnabled:                true,
		OCSPCACertificates:         "ocsp-ca-cert",
		OCSPServersOverride:        []string{"https://ocsp.example.com"},
		OCSPFailOpen:               false,
		OCSPThisUpdateMaxAge:       "12h",
		OCSPMaxRetries:             4,
		OCSPQueryAllServers:        false,
		DisplayName:                "my-cert-role",
		TokenTTL:                   "1h",
		TokenMaxTTL:                "24h",
		TokenPolicies:              []string{"reader"},
		TokenBoundCIDRs:            []string{"10.0.0.0/8"},
		TokenExplicitMaxTTL:        "",
		TokenNoDefaultPolicy:       false,
		TokenNumUses:               0,
		TokenPeriod:                "0",
		TokenType:                  "service",
	}

	result := role.toMap()

	if len(result) != 25 {
		t.Errorf("expected 25 keys in map, got %d", len(result))
	}

	expected := map[string]any{
		"certificate":                  "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
		"allowed_common_names":         []string{"*.example.com"},
		"allowed_dns_sans":             []string{"*.example.com"},
		"allowed_email_sans":           []string{"admin@example.com"},
		"allowed_uri_sans":             []string{"spiffe://cluster.local/*"},
		"allowed_organizational_units": []string{"Engineering"},
		"required_extensions":          []string{"1.2.3.4.5:value"},
		"allowed_metadata_extensions":  []string{"1.2.3.4.5"},
		"ocsp_enabled":                 true,
		"ocsp_ca_certificates":         "ocsp-ca-cert",
		"ocsp_servers_override":        []string{"https://ocsp.example.com"},
		"ocsp_fail_open":               false,
		"ocsp_this_update_max_age":     "12h",
		"ocsp_max_retries":             int64(4),
		"ocsp_query_all_servers":       false,
		"display_name":                 "my-cert-role",
		"token_ttl":                    "1h",
		"token_max_ttl":                "24h",
		"token_policies":               []string{"reader"},
		"token_bound_cidrs":            []string{"10.0.0.0/8"},
		"token_explicit_max_ttl":       "",
		"token_no_default_policy":      false,
		"token_num_uses":               int64(0),
		"token_period":                 "0",
		"token_type":                   "service",
	}

	if !reflect.DeepEqual(result, expected) {
		for k, v := range expected {
			if !reflect.DeepEqual(result[k], v) {
				t.Errorf("key %q: got %v (%T), want %v (%T)", k, result[k], result[k], v, v)
			}
		}
	}
}

func TestCertAuthEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &CertAuthEngineRole{
		Spec: CertAuthEngineRoleSpec{
			CertAuthEngineRoleInternal: CertAuthEngineRoleInternal{
				Certificate: "cert-data",
				TokenTTL:    "1h",
				TokenType:   "service",
			},
		},
	}

	payload := role.Spec.CertAuthEngineRoleInternal.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestCertAuthEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &CertAuthEngineRole{
		Spec: CertAuthEngineRoleSpec{
			CertAuthEngineRoleInternal: CertAuthEngineRoleInternal{
				Certificate: "cert-data",
			},
		},
	}

	payload := role.Spec.CertAuthEngineRoleInternal.toMap()
	payload["certificate"] = "different-cert"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different certificate) to NOT be equivalent")
	}
}

func TestCertAuthEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &CertAuthEngineRole{
		Spec: CertAuthEngineRoleSpec{
			CertAuthEngineRoleInternal: CertAuthEngineRoleInternal{
				Certificate: "cert-data",
			},
		},
	}

	payload := role.Spec.CertAuthEngineRoleInternal.toMap()
	payload["extra_field"] = "unexpected"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual)")
	}
}

func TestCertAuthEngineRoleIsDeletable(t *testing.T) {
	role := &CertAuthEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected CertAuthEngineRole to be deletable")
	}
}

func TestCertAuthEngineRoleConditions(t *testing.T) {
	role := &CertAuthEngineRole{}

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
