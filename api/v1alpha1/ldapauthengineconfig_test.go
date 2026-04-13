package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLDAPAuthEngineConfigGetPath(t *testing.T) {
	config := &LDAPAuthEngineConfig{
		Spec: LDAPAuthEngineConfigSpec{
			Path: "ldap",
		},
	}

	result := config.GetPath()
	expected := "auth/ldap/config"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestLDAPConfigToMap(t *testing.T) {
	config := LDAPConfig{
		URL:                  "ldap://ldap.example.com",
		CaseSensitiveNames:   true,
		RequestTimeout:       "90s",
		StartTLS:             false,
		TLSMinVersion:        "tls12",
		TLSMaxVersion:        "tls13",
		InsecureTLS:          false,
		Certificate:          "",
		ClientTLSCert:        "",
		ClientTLSKey:         "",
		BindDN:               "",
		UserDN:               "ou=Users,dc=example,dc=com",
		UserAttr:             "cn",
		DiscoverDN:           false,
		DenyNullBind:         true,
		UPNDomain:            "example.com",
		UserFilter:           "",
		AnonymousGroupSearch: false,
		GroupFilter:          "",
		GroupDN:              "ou=Groups,dc=example,dc=com",
		GroupAttr:            "cn",
		UsernameAsAlias:      false,
		TokenTTL:             "1h",
		TokenMaxTTL:          "24h",
		TokenPolicies:        "default",
		TokenBoundCIDRs:      "",
		TokenExplicitMaxTTL:  "",
		TokenNoDefaultPolicy: false,
		TokenNumUses:         0,
		TokenPeriod:          0,
		TokenType:            "service",
	}
	config.retrievedUsername = "cn=admin,dc=example,dc=com"
	config.retrievedPassword = "s3cret"

	result := config.toMap()
	expected := map[string]interface{}{
		"url":                     "ldap://ldap.example.com",
		"case_sensitive_names":    true,
		"request_timeout":         "90s",
		"starttls":                false,
		"tls_min_version":         "tls12",
		"tls_max_version":         "tls13",
		"insecure_tls":            false,
		"certificate":             "",
		"client_tls_cert":         "",
		"client_tls_key":          "",
		"binddn":                  "cn=admin,dc=example,dc=com",
		"bindpass":                "s3cret",
		"userdn":                  "ou=Users,dc=example,dc=com",
		"userattr":                "cn",
		"discoverdn":              false,
		"deny_null_bind":          true,
		"upndomain":               "example.com",
		"userfilter":              "",
		"anonymous_group_search":  false,
		"groupfilter":             "",
		"groupdn":                 "ou=Groups,dc=example,dc=com",
		"groupattr":               "cn",
		"username_as_alias":       false,
		"token_ttl":               "1h",
		"token_max_ttl":           "24h",
		"token_policies":          "default",
		"token_bound_cidrs":       "",
		"token_explicit_max_ttl":  "",
		"token_no_default_policy": false,
		"token_num_uses":          int64(0),
		"token_period":            int64(0),
		"token_type":              "service",
	}

	if len(result) != 32 {
		t.Errorf("expected 32 keys in map, got %d", len(result))
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestLDAPConfigToMapInlineCertsCopiedToRetrieved(t *testing.T) {
	config := LDAPConfig{
		URL:           "ldap://ldap.example.com",
		Certificate:   "inline-ca-cert",
		ClientTLSCert: "inline-client-cert",
		ClientTLSKey:  "inline-client-key",
	}

	result := config.toMap()

	if result["certificate"] != "inline-ca-cert" {
		t.Errorf("expected certificate 'inline-ca-cert', got %v", result["certificate"])
	}
	if result["client_tls_cert"] != "inline-client-cert" {
		t.Errorf("expected client_tls_cert 'inline-client-cert', got %v", result["client_tls_cert"])
	}
	if result["client_tls_key"] != "inline-client-key" {
		t.Errorf("expected client_tls_key 'inline-client-key', got %v", result["client_tls_key"])
	}

	if config.retrievedCertificate != "inline-ca-cert" {
		t.Errorf("expected retrievedCertificate mutated to 'inline-ca-cert', got %v", config.retrievedCertificate)
	}
	if config.retrievedClientTLSCert != "inline-client-cert" {
		t.Errorf("expected retrievedClientTLSCert mutated to 'inline-client-cert', got %v", config.retrievedClientTLSCert)
	}
	if config.retrievedClientTLSKey != "inline-client-key" {
		t.Errorf("expected retrievedClientTLSKey mutated to 'inline-client-key', got %v", config.retrievedClientTLSKey)
	}
}

func TestLDAPConfigToMapRetrievedCertsUsedWhenInlineEmpty(t *testing.T) {
	config := LDAPConfig{
		URL: "ldap://ldap.example.com",
	}
	config.retrievedCertificate = "retrieved-ca-cert"
	config.retrievedClientTLSCert = "retrieved-client-cert"
	config.retrievedClientTLSKey = "retrieved-client-key"

	result := config.toMap()

	if result["certificate"] != "retrieved-ca-cert" {
		t.Errorf("expected certificate 'retrieved-ca-cert', got %v", result["certificate"])
	}
	if result["client_tls_cert"] != "retrieved-client-cert" {
		t.Errorf("expected client_tls_cert 'retrieved-client-cert', got %v", result["client_tls_cert"])
	}
	if result["client_tls_key"] != "retrieved-client-key" {
		t.Errorf("expected client_tls_key 'retrieved-client-key', got %v", result["client_tls_key"])
	}
}

func TestLDAPConfigToMapEmptyCertsRemainEmpty(t *testing.T) {
	config := LDAPConfig{
		URL: "ldap://ldap.example.com",
	}

	result := config.toMap()

	if result["certificate"] != "" {
		t.Errorf("expected empty certificate, got %v", result["certificate"])
	}
	if result["client_tls_cert"] != "" {
		t.Errorf("expected empty client_tls_cert, got %v", result["client_tls_cert"])
	}
	if result["client_tls_key"] != "" {
		t.Errorf("expected empty client_tls_key, got %v", result["client_tls_key"])
	}
}

func TestLDAPConfigToMapBindCredentialsFromRetrieved(t *testing.T) {
	config := LDAPConfig{
		URL: "ldap://ldap.example.com",
	}
	config.retrievedUsername = "bind-user"
	config.retrievedPassword = "bind-pass"

	result := config.toMap()

	if result["binddn"] != "bind-user" {
		t.Errorf("expected binddn from retrievedUsername 'bind-user', got %v", result["binddn"])
	}
	if result["bindpass"] != "bind-pass" {
		t.Errorf("expected bindpass from retrievedPassword 'bind-pass', got %v", result["bindpass"])
	}
}

func TestLDAPAuthEngineConfigIsEquivalentBindpassDeleted(t *testing.T) {
	config := &LDAPAuthEngineConfig{
		Spec: LDAPAuthEngineConfigSpec{
			LDAPConfig: LDAPConfig{
				URL:      "ldap://ldap.example.com",
				UserAttr: "cn",
			},
		},
	}
	config.Spec.LDAPConfig.retrievedUsername = "admin"
	config.Spec.LDAPConfig.retrievedPassword = "secret"

	// Build a payload that matches all fields except bindpass is missing
	// (simulates Vault read response which does not return bindpass)
	payload := config.Spec.LDAPConfig.toMap()
	delete(payload, "bindpass")

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload without bindpass to be equivalent (bindpass is deleted from desiredState before comparison)")
	}
}

func TestLDAPAuthEngineConfigIsEquivalentBindpassPresentReturnsFalse(t *testing.T) {
	config := &LDAPAuthEngineConfig{
		Spec: LDAPAuthEngineConfigSpec{
			LDAPConfig: LDAPConfig{
				URL:      "ldap://ldap.example.com",
				UserAttr: "cn",
			},
		},
	}
	config.Spec.LDAPConfig.retrievedUsername = "admin"
	config.Spec.LDAPConfig.retrievedPassword = "secret"

	payload := config.Spec.LDAPConfig.toMap()

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with bindpass present to NOT be equivalent after desiredState deletes bindpass")
	}
}

func TestLDAPAuthEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &LDAPAuthEngineConfig{
		Spec: LDAPAuthEngineConfigSpec{
			LDAPConfig: LDAPConfig{
				URL:      "ldap://ldap.example.com",
				UserAttr: "cn",
			},
		},
	}
	config.Spec.LDAPConfig.retrievedUsername = "admin"
	config.Spec.LDAPConfig.retrievedPassword = "secret"

	// IsEquivalentToDesiredState deletes bindpass from desiredState,
	// so a matching payload should also not have bindpass.
	payload := config.Spec.LDAPConfig.toMap()
	delete(payload, "bindpass")

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (without bindpass) to be equivalent")
	}
}

func TestLDAPAuthEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &LDAPAuthEngineConfig{
		Spec: LDAPAuthEngineConfigSpec{
			LDAPConfig: LDAPConfig{
				URL:      "ldap://ldap.example.com",
				UserAttr: "cn",
			},
		},
	}

	payload := config.Spec.LDAPConfig.toMap()
	delete(payload, "bindpass")
	payload["url"] = "ldap://different-host.com"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different url) to NOT be equivalent")
	}
}

func TestLDAPAuthEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &LDAPAuthEngineConfig{
		Spec: LDAPAuthEngineConfigSpec{
			LDAPConfig: LDAPConfig{
				URL:      "ldap://ldap.example.com",
				UserAttr: "cn",
			},
		},
	}

	payload := config.Spec.LDAPConfig.toMap()
	delete(payload, "bindpass")
	payload["extra_field"] = "unexpected"

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare DeepEqual after bindpass delete)")
	}
}

func TestLDAPAuthEngineConfigIsDeletable(t *testing.T) {
	config := &LDAPAuthEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected LDAPAuthEngineConfig to NOT be deletable")
	}
}

func TestLDAPAuthEngineConfigConditions(t *testing.T) {
	config := &LDAPAuthEngineConfig{}

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
