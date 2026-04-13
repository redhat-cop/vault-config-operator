package v1alpha1

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPKISecretEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *PKISecretEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki",
					Name: "custom-role",
				},
			},
			expectedPath: "pki/roles/custom-role",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki",
				},
			},
			expectedPath: "pki/roles/meta-name",
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

func TestPKIRoleToMap(t *testing.T) {
	r := PKIRole{
		TTL:                           metav1.Duration{Duration: 1 * time.Hour},
		MaxTTL:                        metav1.Duration{Duration: 24 * time.Hour},
		AllowLocalhost:                true,
		AllowedDomains:                []string{"example.com", "*.example.com"},
		AllowedDomainsTemplate:        true,
		AllowBareDomains:              false,
		AllowSubdomains:               true,
		AllowGlobDomains:              false,
		AllowAnyName:                  false,
		EnforceHostnames:              true,
		AllowIPSans:                   true,
		AllowedURISans:                []string{"https://idp.example/*"},
		AllowedOtherSans:              "1.3.6.1.4.1.311.20.2.3;UTF8:abc",
		ServerFlag:                    true,
		ClientFlag:                    false,
		CodeSigningFlag:               false,
		EmailProtectionFlag:           false,
		KeyType:                       "rsa",
		KeyBits:                       2048,
		KeyUsage:                      []KeyUsage{"DigitalSignature", "KeyEncipherment"},
		ExtKeyUsage:                   []ExtKeyUsage{"ServerAuth", "ClientAuth"},
		ExtKeyUsageOids:               []string{"1.3.6.1.5.5.7.3.1"},
		UseCSRCommonName:              true,
		UseCSRSans:                    false,
		OU:                            "unit",
		Organization:                  "Org",
		Country:                       "US",
		Locality:                      "City",
		Province:                      "ST",
		StreetAddress:                 "Road 1",
		PostalCode:                    "12345",
		SerialNumber:                  "role-serial",
		GenerateLease:                 true,
		NoStore:                       false,
		RequireCn:                     true,
		PolicyIdentifiers:             []string{"1.2.3.4"},
		BasicConstraintsValidForNonCa: true,
		NotBeforeDuration:             metav1.Duration{Duration: 45 * time.Second},
	}

	result := r.toMap()

	expectedKeys := []string{
		"ttl", "max_ttl", "allow_localhost", "allowed_domains", "allowed_domains_template",
		"allow_bare_domains", "allow_subdomains", "allow_glob_domains", "allow_any_name",
		"enforce_hostnames", "allow_ip_sans", "allowed_uri_sans", "allowed_other_sans",
		"server_flag", "client_flag", "code_signing_flag", "email_protection_flag",
		"key_type", "key_bits", "key_usage", "ext_key_usage", "ext_key_usage_oids",
		"use_csr_common_name", "use_csr_sans", "ou", "organization", "country",
		"locality", "province", "street_address", "postal_code", "serial_number",
		"generate_lease", "no_store", "require_cn", "policy_identifiers",
		"basic_constraints_valid_for_non_ca", "not_before_duration",
	}
	if len(result) != 38 {
		t.Errorf("expected 38 keys in toMap() output, got %d", len(result))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if _, ok := result["ttl"].(metav1.Duration); !ok {
		t.Errorf("ttl type = %T, want metav1.Duration", result["ttl"])
	}
	if _, ok := result["max_ttl"].(metav1.Duration); !ok {
		t.Errorf("max_ttl type = %T, want metav1.Duration", result["max_ttl"])
	}
	if result["key_type"] != "rsa" {
		t.Errorf("key_type = %v, want rsa", result["key_type"])
	}
	if result["key_bits"] != 2048 {
		t.Errorf("key_bits = %v, want 2048", result["key_bits"])
	}
	ku, ok := result["key_usage"].([]KeyUsage)
	if !ok || len(ku) != 2 {
		t.Errorf("key_usage = %v, want []KeyUsage len 2", result["key_usage"])
	}
	if !reflect.DeepEqual(result["not_before_duration"], metav1.Duration{Duration: 45 * time.Second}) {
		t.Errorf("not_before_duration = %v", result["not_before_duration"])
	}
}

func TestPKISecretEngineRoleIsEquivalentMatching(t *testing.T) {
	config := &PKISecretEngineRole{
		Spec: PKISecretEngineRoleSpec{
			PKIRole: PKIRole{
				KeyType: "ec",
				KeyBits: 256,
			},
		},
	}
	payload := config.Spec.PKIRole.toMap()
	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for matching payload")
	}
}

func TestPKISecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	config := &PKISecretEngineRole{
		Spec: PKISecretEngineRoleSpec{
			PKIRole: PKIRole{
				KeyType: "rsa",
			},
		},
	}
	payload := config.Spec.PKIRole.toMap()
	payload["key_type"] = "ec"
	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when key_type differs")
	}
}

func TestPKISecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	config := &PKISecretEngineRole{
		Spec: PKISecretEngineRoleSpec{
			PKIRole: PKIRole{
				AllowLocalhost: true,
			},
		},
	}
	payload := config.Spec.PKIRole.toMap()
	payload["vault_extra"] = true
	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when payload has extra keys")
	}
}

func TestPKISecretEngineRoleIsDeletable(t *testing.T) {
	config := &PKISecretEngineRole{}
	if !config.IsDeletable() {
		t.Error("expected PKISecretEngineRole to be deletable")
	}
}

func TestPKISecretEngineRoleConditions(t *testing.T) {
	config := &PKISecretEngineRole{}
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
		t.Errorf("condition type = %v, want ReconcileSuccessful", got[0].Type)
	}
	if got[0].Status != metav1.ConditionTrue {
		t.Errorf("condition status = %v, want True", got[0].Status)
	}
}
