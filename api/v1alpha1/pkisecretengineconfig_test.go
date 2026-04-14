package v1alpha1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPKISecretEngineConfigGetPath(t *testing.T) {
	config := &PKISecretEngineConfig{
		Spec: PKISecretEngineConfigSpec{
			Path: "pki",
		},
	}
	if got := config.GetPath(); got != "pki" {
		t.Errorf("GetPath() = %q, want %q", got, "pki")
	}
}

func TestPKICommonToMap(t *testing.T) {
	c := PKICommon{
		CommonName:          "example-ca",
		AltNames:            "a.example.com,b.example.com",
		IPSans:              "10.0.0.1",
		URISans:             "spiffe://cluster/ns",
		OtherSans:           "1.2.3.4;UTF8:other",
		TTL:                 metav1.Duration{Duration: 24 * time.Hour},
		Format:              "pem",
		PrivateKeyFormat:    "der",
		KeyType:             "rsa",
		KeyBits:             4096,
		MaxPathLength:       2,
		ExcludeCnFromSans:   true,
		PermittedDnsDomains: []string{"corp.example.com", "*.svc"},
		OU:                  "eng",
		Organization:        "Acme",
		Country:             "US",
		Locality:            "Boston",
		Province:            "MA",
		StreetAddress:       "1 Main St",
		PostalCode:          "02101",
		SerialNumber:        "SN-100",
	}

	result := c.toMap()

	expectedKeys := []string{
		"common_name", "alt_names", "ip_sans", "uri_sans", "other_sans", "ttl",
		"format", "private_key_format", "key_type", "key_bits", "max_path_length",
		"exclude_cn_from_sans", "permitted_dns_domains", "ou", "organization",
		"country", "locality", "province", "street_address", "postal_code", "serial_number",
	}
	if len(result) != 21 {
		t.Errorf("expected 21 keys in toMap() output, got %d", len(result))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}
	if _, ok := result["type"]; ok {
		t.Error("did not expect key \"type\" in toMap() output")
	}

	if _, ok := result["ttl"].(metav1.Duration); !ok {
		t.Errorf("ttl type = %T, want metav1.Duration", result["ttl"])
	}
	if result["key_bits"] != 4096 {
		t.Errorf("key_bits = %v, want 4096", result["key_bits"])
	}
	if result["max_path_length"] != 2 {
		t.Errorf("max_path_length = %v, want 2", result["max_path_length"])
	}
	if result["exclude_cn_from_sans"] != true {
		t.Errorf("exclude_cn_from_sans = %v, want true", result["exclude_cn_from_sans"])
	}
	domains, ok := result["permitted_dns_domains"].([]string)
	if !ok || len(domains) != 2 {
		t.Errorf("permitted_dns_domains = %v, want []string of len 2", result["permitted_dns_domains"])
	}
}

func TestPKIConfigUrlsToMap(t *testing.T) {
	u := PKIConfigUrls{
		IssuingCertificates:   []string{"http://ca.example/issuing.pem"},
		CRLDistributionPoints: []string{"http://ca.example/crl.pem"},
		OcspServers:           []string{"http://ocsp.example"},
	}
	result := u.toMap()
	if len(result) != 3 {
		t.Errorf("expected 3 keys, got %d", len(result))
	}
	for _, key := range []string{"issuing_certificates", "crl_distribution_points", "ocsp_servers"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q", key)
		}
	}
	issuing, ok := result["issuing_certificates"].([]string)
	if !ok || len(issuing) != 1 {
		t.Errorf("issuing_certificates = %v", result["issuing_certificates"])
	}
}

func TestPKIConfigCRLToMap(t *testing.T) {
	c := PKIConfigCRL{
		CRLExpiry:  metav1.Duration{Duration: 72 * time.Hour},
		CRLDisable: true,
	}
	result := c.toMap()
	if len(result) != 2 {
		t.Errorf("expected 2 keys, got %d", len(result))
	}
	if _, ok := result["expiry"].(metav1.Duration); !ok {
		t.Errorf("expiry type = %T, want metav1.Duration", result["expiry"])
	}
	if result["disable"] != true {
		t.Errorf("disable = %v, want true", result["disable"])
	}
}

func TestPKISecretEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &PKISecretEngineConfig{
		Spec: PKISecretEngineConfigSpec{
			PKICommon: PKICommon{
				CommonName: "my-ca",
				KeyBits:    2048,
				Format:     "pem",
			},
		},
	}
	payload := config.Spec.PKICommon.toMap()
	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for matching PKICommon map")
	}
}

func TestPKISecretEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &PKISecretEngineConfig{
		Spec: PKISecretEngineConfigSpec{
			PKICommon: PKICommon{
				CommonName: "my-ca",
				KeyBits:    2048,
			},
		},
	}
	payload := config.Spec.PKICommon.toMap()
	payload["common_name"] = "other-ca"
	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when common_name differs")
	}
}

func TestPKISecretEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &PKISecretEngineConfig{
		Spec: PKISecretEngineConfigSpec{
			PKICommon: PKICommon{
				CommonName: "my-ca",
			},
		},
	}
	payload := config.Spec.PKICommon.toMap()
	payload["extra_from_vault"] = "x"
	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when payload has extra keys")
	}
}

func TestPKISecretEngineConfigIsDeletable(t *testing.T) {
	config := &PKISecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected PKISecretEngineConfig not to be deletable")
	}
}

func TestPKISecretEngineConfigConditions(t *testing.T) {
	config := &PKISecretEngineConfig{}
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
