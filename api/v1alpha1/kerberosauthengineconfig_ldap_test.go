package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestKerberosAuthEngineLDAPConfig_toMap(t *testing.T) {
	config := &KerberosLDAPConfig{
		URL:                    "ldaps://ldap.myorg.com:636",
		CaseSensitiveNames:     false,
		StartTLS:               false,
		TLSMinVersion:          "tls12",
		TLSMaxVersion:          "tls12",
		InsecureTLS:            false,
		Certificate:            "",
		ClientTLSCert:          "",
		ClientTLSKey:           "",
		AliasMetadata:          map[string]string{"spn": "spnego"},
		BindDN:                 "",
		UserDN:                 "ou=Users,dc=example,dc=com",
		UserAttr:               "samaccountname",
		DiscoverDN:             false,
		DenyNullBind:           true,
		UPNDomain:              "",
		GroupFilter:            "(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))",
		GroupDN:                "ou=Groups,dc=example,dc=com",
		GroupAttr:              "cn",
		TokenTTL:               "1h",
		TokenMaxTTL:            "4h",
		TokenPolicies:          "default,admin",
		TokenBoundCIDRs:        "10.0.0.0/8",
		TokenExplicitMaxTTL:    "8h",
		TokenNoDefaultPolicy:   true,
		TokenNumUses:           5,
		TokenPeriod:            "24h",
		TokenType:              "service",
		retrievedPassword:      "bind-password",
		retrievedUsername:      "cn=vault,ou=Users,dc=example,dc=com",
		retrievedCertificate:   "PEM-cert-data",
		retrievedClientTLSCert: "PEM-client-cert",
		retrievedClientTLSKey:  "PEM-client-key",
	}

	result := config.toMap()

	if result["url"] != "ldaps://ldap.myorg.com:636" {
		t.Errorf("expected url=ldaps://ldap.myorg.com:636, got %v", result["url"])
	}
	if result["case_sensitive_names"] != false {
		t.Errorf("expected case_sensitive_names=false, got %v", result["case_sensitive_names"])
	}
	if result["starttls"] != false {
		t.Errorf("expected starttls=false, got %v", result["starttls"])
	}
	if result["tls_min_version"] != "tls12" {
		t.Errorf("expected tls_min_version=tls12, got %v", result["tls_min_version"])
	}
	if result["tls_max_version"] != "tls12" {
		t.Errorf("expected tls_max_version=tls12, got %v", result["tls_max_version"])
	}
	if result["insecure_tls"] != false {
		t.Errorf("expected insecure_tls=false, got %v", result["insecure_tls"])
	}
	if result["certificate"] != "PEM-cert-data" {
		t.Errorf("expected certificate=PEM-cert-data, got %v", result["certificate"])
	}
	if result["client_tls_cert"] != "PEM-client-cert" {
		t.Errorf("expected client_tls_cert=PEM-client-cert, got %v", result["client_tls_cert"])
	}
	if result["client_tls_key"] != "PEM-client-key" {
		t.Errorf("expected client_tls_key=PEM-client-key, got %v", result["client_tls_key"])
	}
	expectedAliasMetadata := map[string]any{"spn": "spnego"}
	if gotAM, ok := result["alias_metadata"].(map[string]any); !ok {
		t.Errorf("expected alias_metadata to be map[string]any, got %T", result["alias_metadata"])
	} else if len(gotAM) != len(expectedAliasMetadata) || gotAM["spn"] != "spnego" {
		t.Errorf("expected alias_metadata=%v, got %v", expectedAliasMetadata, gotAM)
	}
	if result["binddn"] != "cn=vault,ou=Users,dc=example,dc=com" {
		t.Errorf("expected binddn from resolved credentials, got %v", result["binddn"])
	}
	if result["bindpass"] != "bind-password" {
		t.Errorf("expected bindpass from resolved credentials, got %v", result["bindpass"])
	}
	if result["userdn"] != "ou=Users,dc=example,dc=com" {
		t.Errorf("expected userdn=ou=Users,dc=example,dc=com, got %v", result["userdn"])
	}
	if result["userattr"] != "samaccountname" {
		t.Errorf("expected userattr=samaccountname, got %v", result["userattr"])
	}
	if result["discoverdn"] != false {
		t.Errorf("expected discoverdn=false, got %v", result["discoverdn"])
	}
	if result["deny_null_bind"] != true {
		t.Errorf("expected deny_null_bind=true, got %v", result["deny_null_bind"])
	}
	if result["upndomain"] != "" {
		t.Errorf("expected upndomain empty, got %v", result["upndomain"])
	}
	if result["groupfilter"] != "(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))" {
		t.Errorf("expected groupfilter value, got %v", result["groupfilter"])
	}
	if result["groupdn"] != "ou=Groups,dc=example,dc=com" {
		t.Errorf("expected groupdn=ou=Groups,dc=example,dc=com, got %v", result["groupdn"])
	}
	if result["groupattr"] != "cn" {
		t.Errorf("expected groupattr=cn, got %v", result["groupattr"])
	}
	if result["token_ttl"] != json.Number("3600") {
		t.Errorf("expected token_ttl=3600, got %v (type %T)", result["token_ttl"], result["token_ttl"])
	}
	if result["token_max_ttl"] != json.Number("14400") {
		t.Errorf("expected token_max_ttl=14400, got %v", result["token_max_ttl"])
	}
	if result["token_policies"] != "default,admin" {
		t.Errorf("expected token_policies=default,admin, got %v", result["token_policies"])
	}
	if result["token_bound_cidrs"] != "10.0.0.0/8" {
		t.Errorf("expected token_bound_cidrs=10.0.0.0/8, got %v", result["token_bound_cidrs"])
	}
	if result["token_explicit_max_ttl"] != json.Number("28800") {
		t.Errorf("expected token_explicit_max_ttl=28800, got %v", result["token_explicit_max_ttl"])
	}
	if result["token_no_default_policy"] != true {
		t.Errorf("expected token_no_default_policy=true, got %v", result["token_no_default_policy"])
	}
	if result["token_num_uses"] != json.Number("5") {
		t.Errorf("expected token_num_uses=json.Number(5), got %v (type %T)", result["token_num_uses"], result["token_num_uses"])
	}
	if result["token_period"] != json.Number("86400") {
		t.Errorf("expected token_period=86400, got %v", result["token_period"])
	}
	if result["token_type"] != "service" {
		t.Errorf("expected token_type=service, got %v", result["token_type"])
	}
}

func TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &KerberosAuthEngineLDAPConfig{}
	instance.Spec.KerberosLDAPConfig = KerberosLDAPConfig{
		URL:                "ldaps://ldap.myorg.com:636",
		CaseSensitiveNames: false,
		StartTLS:           false,
		TLSMinVersion:      "tls12",
		TLSMaxVersion:      "tls12",
		InsecureTLS:        false,
		UserDN:             "ou=Users,dc=example,dc=com",
		UserAttr:           "samaccountname",
		DiscoverDN:         false,
		DenyNullBind:       true,
		GroupFilter:        "(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))",
		GroupDN:            "ou=Groups,dc=example,dc=com",
		GroupAttr:          "cn",
		TokenNumUses:       0,
		retrievedUsername:  "cn=vault,ou=Users,dc=example,dc=com",
		retrievedPassword:  "secret-pass",
	}

	vaultPayload := map[string]any{
		"url":                     "ldaps://ldap.myorg.com:636",
		"case_sensitive_names":    false,
		"starttls":                false,
		"tls_min_version":         "tls12",
		"tls_max_version":         "tls12",
		"insecure_tls":            false,
		"certificate":             "",
		"client_tls_cert":         "",
		"alias_metadata":          "",
		"binddn":                  "cn=vault,ou=Users,dc=example,dc=com",
		"bindpass":                "",
		"userdn":                  "ou=Users,dc=example,dc=com",
		"userattr":                "samaccountname",
		"discoverdn":              false,
		"deny_null_bind":          true,
		"upndomain":               "",
		"groupfilter":             "(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))",
		"groupdn":                 "ou=Groups,dc=example,dc=com",
		"groupattr":               "cn",
		"token_ttl":               json.Number("0"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          "",
		"token_bound_cidrs":       "",
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching state (bindpass and client_tls_key stripped)")
	}
}

func TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &KerberosAuthEngineLDAPConfig{}
	instance.Spec.KerberosLDAPConfig = KerberosLDAPConfig{
		URL:               "ldaps://ldap.myorg.com:636",
		TLSMinVersion:     "tls12",
		TLSMaxVersion:     "tls12",
		DenyNullBind:      true,
		retrievedUsername: "cn=vault,ou=Users,dc=example,dc=com",
		retrievedPassword: "pass",
	}

	vaultPayload := map[string]any{
		"url":                     "ldaps://different-ldap.myorg.com:636",
		"case_sensitive_names":    false,
		"starttls":                false,
		"tls_min_version":         "tls12",
		"tls_max_version":         "tls12",
		"insecure_tls":            false,
		"certificate":             "",
		"client_tls_cert":         "",
		"alias_metadata":          "",
		"binddn":                  "cn=vault,ou=Users,dc=example,dc=com",
		"bindpass":                "",
		"userdn":                  "",
		"userattr":                "",
		"discoverdn":              false,
		"deny_null_bind":          true,
		"upndomain":               "",
		"groupfilter":             "",
		"groupdn":                 "",
		"groupattr":               "",
		"token_ttl":               json.Number("0"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          "",
		"token_bound_cidrs":       "",
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched url")
	}
}

func TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_BindpassStripping(t *testing.T) {
	instance := &KerberosAuthEngineLDAPConfig{}
	instance.Spec.KerberosLDAPConfig = KerberosLDAPConfig{
		URL:                   "ldaps://ldap.myorg.com:636",
		TLSMinVersion:         "tls12",
		TLSMaxVersion:         "tls12",
		DenyNullBind:          true,
		retrievedUsername:     "cn=vault,ou=Users,dc=example,dc=com",
		retrievedPassword:     "super-secret-password",
		retrievedClientTLSKey: "secret-key-data",
	}

	vaultPayload := map[string]any{
		"url":                     "ldaps://ldap.myorg.com:636",
		"case_sensitive_names":    false,
		"starttls":                false,
		"tls_min_version":         "tls12",
		"tls_max_version":         "tls12",
		"insecure_tls":            false,
		"certificate":             "",
		"client_tls_cert":         "",
		"alias_metadata":          "",
		"binddn":                  "cn=vault,ou=Users,dc=example,dc=com",
		"bindpass":                "",
		"userdn":                  "",
		"userattr":                "",
		"discoverdn":              false,
		"deny_null_bind":          true,
		"upndomain":               "",
		"groupfilter":             "",
		"groupdn":                 "",
		"groupattr":               "",
		"token_ttl":               json.Number("0"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          "",
		"token_bound_cidrs":       "",
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: bindpass and client_tls_key are write-only and must be excluded from drift comparison")
	}
}

func TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &KerberosAuthEngineLDAPConfig{}
	instance.Spec.KerberosLDAPConfig = KerberosLDAPConfig{
		URL:               "ldaps://ldap.myorg.com:636",
		TLSMinVersion:     "tls12",
		TLSMaxVersion:     "tls12",
		DenyNullBind:      true,
		retrievedUsername: "cn=vault,ou=Users,dc=example,dc=com",
		retrievedPassword: "pass",
	}

	vaultPayload := map[string]any{
		"url":                     "ldaps://ldap.myorg.com:636",
		"case_sensitive_names":    false,
		"starttls":                false,
		"tls_min_version":         "tls12",
		"tls_max_version":         "tls12",
		"insecure_tls":            false,
		"certificate":             "",
		"client_tls_cert":         "",
		"alias_metadata":          "",
		"binddn":                  "cn=vault,ou=Users,dc=example,dc=com",
		"bindpass":                "",
		"userdn":                  "",
		"userattr":                "",
		"discoverdn":              false,
		"deny_null_bind":          true,
		"upndomain":               "",
		"groupfilter":             "",
		"groupdn":                 "",
		"groupattr":               "",
		"token_ttl":               json.Number("0"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          "",
		"token_bound_cidrs":       "",
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
		"request_id":              "extra-vault-field",
		"lease_duration":          json.Number("0"),
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestKerberosAuthEngineLDAPConfig_AliasMetadata_DeepEqualWithVaultPayload(t *testing.T) {
	config := &KerberosLDAPConfig{
		AliasMetadata: map[string]string{"spn": "spnego", "realm": "EXAMPLE.COM"},
	}

	toMapResult := config.toMap()

	vaultPayload := map[string]any{
		"spn":   "spnego",
		"realm": "EXAMPLE.COM",
	}

	desiredAM, ok := toMapResult["alias_metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected alias_metadata to be map[string]any, got %T", toMapResult["alias_metadata"])
	}
	if !reflect.DeepEqual(desiredAM, vaultPayload) {
		t.Errorf("toMap() alias_metadata should DeepEqual a Vault-shaped map[string]any payload\n  got:  %v\n  want: %v", desiredAM, vaultPayload)
	}
}

func TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_WithAliasMetadata(t *testing.T) {
	instance := &KerberosAuthEngineLDAPConfig{}
	instance.Spec.KerberosLDAPConfig = KerberosLDAPConfig{
		URL:               "ldaps://ldap.myorg.com:636",
		TLSMinVersion:     "tls12",
		TLSMaxVersion:     "tls12",
		DenyNullBind:      true,
		AliasMetadata:     map[string]string{"spn": "spnego"},
		retrievedUsername: "cn=vault,ou=Users,dc=example,dc=com",
		retrievedPassword: "pass",
	}

	vaultPayload := map[string]any{
		"url":                     "ldaps://ldap.myorg.com:636",
		"case_sensitive_names":    false,
		"starttls":                false,
		"tls_min_version":         "tls12",
		"tls_max_version":         "tls12",
		"insecure_tls":            false,
		"certificate":             "",
		"client_tls_cert":         "",
		"alias_metadata":          map[string]any{"spn": "spnego"},
		"binddn":                  "cn=vault,ou=Users,dc=example,dc=com",
		"bindpass":                "",
		"userdn":                  "",
		"userattr":                "",
		"discoverdn":              false,
		"deny_null_bind":          true,
		"upndomain":               "",
		"groupfilter":             "",
		"groupdn":                 "",
		"groupattr":               "",
		"token_ttl":               json.Number("0"),
		"token_max_ttl":           json.Number("0"),
		"token_policies":          "",
		"token_bound_cidrs":       "",
		"token_explicit_max_ttl":  json.Number("0"),
		"token_no_default_policy": false,
		"token_num_uses":          json.Number("0"),
		"token_period":            json.Number("0"),
		"token_type":              "",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true: aliasMetadata map[string]any from Vault must match toMap() output")
	}
}

func TestKerberosAuthEngineLDAPConfig_GetPath(t *testing.T) {
	instance := &KerberosAuthEngineLDAPConfig{}
	instance.Spec.Path = "kerberos"

	expected := "auth/kerberos/config/ldap"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestKerberosAuthEngineLDAPConfig_IsDeletable(t *testing.T) {
	instance := &KerberosAuthEngineLDAPConfig{}
	if instance.IsDeletable() {
		t.Error("expected IsDeletable() to return false for auth engine LDAP config")
	}
}
