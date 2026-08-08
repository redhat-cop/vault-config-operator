package v1alpha1

import (
	"context"
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSSHSecretEngineRoleGetPath(t *testing.T) {
	role := &SSHSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: SSHSecretEngineRoleSpec{
			Path: "ssh",
		},
	}
	expected := "ssh/roles/my-role"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestSSHSecretEngineRoleGetPathWithNameOverride(t *testing.T) {
	role := &SSHSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: SSHSecretEngineRoleSpec{
			Path: "ssh",
			Name: "custom-name",
		},
	}
	expected := "ssh/roles/custom-name"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestSSHSERoleToMapCA(t *testing.T) {
	role := SSHSERole{
		KeyType:               "ca",
		DefaultUser:           "admin",
		AllowedUsers:          "*",
		TTL:                   "4h",
		MaxTTL:                "768h",
		Port:                  22,
		AllowUserCertificates: true,
		AllowHostCertificates: true,
		AllowBareDomains:      true,
		AllowSubdomains:       true,
		AllowedDomains:        "example.com,test.com",
		AllowedExtensions:     "permit-pty",
		AlgorithmSigner:       "rsa-sha2-256",
		NotBeforeDuration:     "30s",
	}

	result := role.toMap()

	if result["key_type"] != "ca" {
		t.Errorf("key_type = %v, want ca", result["key_type"])
	}
	if result["default_user"] != "admin" {
		t.Errorf("default_user = %v", result["default_user"])
	}
	if result["allowed_users"] != "*" {
		t.Errorf("allowed_users = %v", result["allowed_users"])
	}
	if result["ttl"] != json.Number("14400") {
		t.Errorf("ttl = %v, want json.Number(14400) for 4h", result["ttl"])
	}
	if result["max_ttl"] != json.Number("2764800") {
		t.Errorf("max_ttl = %v, want json.Number(2764800) for 768h", result["max_ttl"])
	}
	if result["port"] != json.Number("22") {
		t.Errorf("port = %v, want json.Number(22)", result["port"])
	}
	if result["allow_user_certificates"] != true {
		t.Errorf("allow_user_certificates = %v", result["allow_user_certificates"])
	}
	if result["allow_host_certificates"] != true {
		t.Errorf("allow_host_certificates = %v", result["allow_host_certificates"])
	}
	if result["allow_bare_domains"] != true {
		t.Errorf("allow_bare_domains = %v", result["allow_bare_domains"])
	}
	if result["allow_subdomains"] != true {
		t.Errorf("allow_subdomains = %v", result["allow_subdomains"])
	}
	if result["allowed_domains"] != "example.com,test.com" {
		t.Errorf("allowed_domains = %v", result["allowed_domains"])
	}
	if result["allowed_extensions"] != "permit-pty" {
		t.Errorf("allowed_extensions = %v", result["allowed_extensions"])
	}
	if result["algorithm_signer"] != "rsa-sha2-256" {
		t.Errorf("algorithm_signer = %v", result["algorithm_signer"])
	}
	if result["not_before_duration"] != json.Number("30") {
		t.Errorf("not_before_duration = %v, want json.Number(30) for 30s", result["not_before_duration"])
	}

	if _, ok := result["cidr_list"]; ok {
		t.Error("cidr_list should not be present for CA role")
	}
	if _, ok := result["exclude_cidr_list"]; ok {
		t.Error("exclude_cidr_list should not be present for CA role")
	}
}

func TestSSHSERoleToMapOTP(t *testing.T) {
	role := SSHSERole{
		KeyType:     "otp",
		DefaultUser: "ubuntu",
		CIDRList:    "0.0.0.0/0",
		Port:        22,
	}

	result := role.toMap()

	if result["key_type"] != "otp" {
		t.Errorf("key_type = %v, want otp", result["key_type"])
	}
	if result["default_user"] != "ubuntu" {
		t.Errorf("default_user = %v", result["default_user"])
	}
	if result["cidr_list"] != "0.0.0.0/0" {
		t.Errorf("cidr_list = %v", result["cidr_list"])
	}
	if result["port"] != json.Number("22") {
		t.Errorf("port = %v, want json.Number(22)", result["port"])
	}

	if _, ok := result["allowed_domains"]; ok {
		t.Error("allowed_domains should not be present for OTP role")
	}
	if _, ok := result["allow_user_certificates"]; ok {
		t.Error("allow_user_certificates should not be present for OTP role")
	}
}

func TestSSHSecretEngineRoleIsEquivalentCA(t *testing.T) {
	role := &SSHSecretEngineRole{
		Spec: SSHSecretEngineRoleSpec{
			Path: "ssh",
			SSHSERole: SSHSERole{
				KeyType:               "ca",
				DefaultUser:           "admin",
				AllowedUsers:          "*",
				TTL:                   "4h",
				MaxTTL:                "768h",
				Port:                  22,
				AllowUserCertificates: true,
				AllowHostCertificates: false,
				AllowedExtensions:     "permit-pty",
			},
		},
	}

	payload := map[string]any{
		"key_type":                    "ca",
		"default_user":                "admin",
		"default_user_template":       false,
		"allowed_users":               "*",
		"allowed_users_template":      false,
		"ttl":                         json.Number("14400"),
		"max_ttl":                     json.Number("2764800"),
		"port":                        json.Number("22"),
		"allowed_domains":             "",
		"allowed_domains_template":    false,
		"allow_user_certificates":     true,
		"allow_host_certificates":     false,
		"allow_bare_domains":          false,
		"allow_subdomains":            false,
		"allow_user_key_ids":          false,
		"key_id_format":               "",
		"allowed_user_key_lengths":    map[string]any{},
		"allowed_critical_options":    "",
		"allowed_extensions":          "permit-pty",
		"default_critical_options":    map[string]any{},
		"default_extensions":          map[string]any{},
		"default_extensions_template": false,
		"allow_empty_principals":      false,
		"algorithm_signer":            "",
		"not_before_duration":         json.Number("0"),
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for matching CA role payload")
	}
}

func TestSSHSecretEngineRoleIsEquivalentOTP(t *testing.T) {
	role := &SSHSecretEngineRole{
		Spec: SSHSecretEngineRoleSpec{
			Path: "ssh",
			SSHSERole: SSHSERole{
				KeyType:     "otp",
				DefaultUser: "ubuntu",
				CIDRList:    "0.0.0.0/0",
				Port:        22,
			},
		},
	}

	payload := map[string]any{
		"key_type":               "otp",
		"default_user":           "ubuntu",
		"default_user_template":  false,
		"allowed_users":          "",
		"allowed_users_template": false,
		"ttl":                    json.Number("0"),
		"max_ttl":                json.Number("0"),
		"port":                   json.Number("22"),
		"cidr_list":              "0.0.0.0/0",
		"exclude_cidr_list":      "",
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for matching OTP role payload")
	}
}

func TestSSHSecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &SSHSecretEngineRole{
		Spec: SSHSecretEngineRoleSpec{
			Path: "ssh",
			SSHSERole: SSHSERole{
				KeyType:     "otp",
				DefaultUser: "ubuntu",
				CIDRList:    "0.0.0.0/0",
				Port:        22,
			},
		},
	}

	payload := map[string]any{
		"key_type":               "otp",
		"default_user":           "centos",
		"default_user_template":  false,
		"allowed_users":          "",
		"allowed_users_template": false,
		"ttl":                    json.Number("0"),
		"max_ttl":                json.Number("0"),
		"port":                   json.Number("22"),
		"cidr_list":              "0.0.0.0/0",
		"exclude_cidr_list":      "",
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when default_user differs")
	}
}

func TestSSHSecretEngineRoleIsEquivalentMapFieldsNoDrift(t *testing.T) {
	role := &SSHSecretEngineRole{
		Spec: SSHSecretEngineRoleSpec{
			Path: "ssh",
			SSHSERole: SSHSERole{
				KeyType:               "ca",
				DefaultUser:           "admin",
				AllowedUsers:          "*",
				Port:                  22,
				AllowUserCertificates: true,
				AllowedUserKeyLengths: map[string]int{
					"rsa":     2048,
					"ecdsa":   256,
					"ed25519": 0,
				},
				DefaultCriticalOptions: map[string]string{
					"force-command": "/usr/bin/id",
				},
				DefaultExtensions: map[string]string{
					"permit-pty":       "",
					"permit-port-forwarding": "",
				},
			},
		},
	}

	// Simulate what Vault returns after JSON deserialization:
	// the Vault Go client uses UseNumber(), so all JSON numbers become json.Number
	vaultPayload := map[string]any{
		"key_type":                 "ca",
		"default_user":            "admin",
		"default_user_template":   false,
		"allowed_users":           "*",
		"allowed_users_template":  false,
		"ttl":                     json.Number("0"),
		"max_ttl":                 json.Number("0"),
		"port":                    json.Number("22"),
		"allowed_domains":         "",
		"allowed_domains_template": false,
		"allow_user_certificates": true,
		"allow_host_certificates": false,
		"allow_bare_domains":      false,
		"allow_subdomains":        false,
		"allow_user_key_ids":      false,
		"key_id_format":           "",
		"allowed_user_key_lengths": map[string]any{
			"rsa":     json.Number("2048"),
			"ecdsa":   json.Number("256"),
			"ed25519": json.Number("0"),
		},
		"allowed_critical_options": "",
		"allowed_extensions":       "",
		"default_critical_options": map[string]any{
			"force-command": "/usr/bin/id",
		},
		"default_extensions": map[string]any{
			"permit-pty":             "",
			"permit-port-forwarding": "",
		},
		"default_extensions_template": false,
		"allow_empty_principals":      false,
		"algorithm_signer":            "",
		"not_before_duration":         json.Number("0"),
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected no drift: map fields should match Vault's JSON-deserialized types (json.Number for numbers)")
	}
}

func TestSSHSecretEngineRoleIsEquivalentNilMapsVsEmptyMaps(t *testing.T) {
	role := &SSHSecretEngineRole{
		Spec: SSHSecretEngineRoleSpec{
			Path: "ssh",
			SSHSERole: SSHSERole{
				KeyType:               "ca",
				DefaultUser:           "admin",
				Port:                  22,
				AllowUserCertificates: true,
			},
		},
	}

	// Vault returns empty JSON objects {} for unset map fields, not null.
	// toMap() must produce empty maps (not nil) to match.
	vaultPayload := map[string]any{
		"key_type":                    "ca",
		"default_user":               "admin",
		"default_user_template":       false,
		"allowed_users":              "",
		"allowed_users_template":      false,
		"ttl":                         json.Number("0"),
		"max_ttl":                     json.Number("0"),
		"port":                        json.Number("22"),
		"allowed_domains":             "",
		"allowed_domains_template":    false,
		"allow_user_certificates":     true,
		"allow_host_certificates":     false,
		"allow_bare_domains":          false,
		"allow_subdomains":            false,
		"allow_user_key_ids":          false,
		"key_id_format":               "",
		"allowed_user_key_lengths":    map[string]any{},
		"allowed_critical_options":    "",
		"allowed_extensions":          "",
		"default_critical_options":    map[string]any{},
		"default_extensions":          map[string]any{},
		"default_extensions_template": false,
		"allow_empty_principals":      false,
		"algorithm_signer":            "",
		"not_before_duration":         json.Number("0"),
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected no drift: nil CRD maps should match Vault's empty map {} response")
	}
}

func TestSSHSecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &SSHSecretEngineRole{
		Spec: SSHSecretEngineRoleSpec{
			Path: "ssh",
			SSHSERole: SSHSERole{
				KeyType:     "otp",
				DefaultUser: "ubuntu",
				CIDRList:    "0.0.0.0/0",
				Port:        22,
			},
		},
	}

	payload := map[string]any{
		"key_type":               "otp",
		"default_user":           "ubuntu",
		"default_user_template":  false,
		"allowed_users":          "",
		"allowed_users_template": false,
		"ttl":                    json.Number("0"),
		"max_ttl":                json.Number("0"),
		"port":                   json.Number("22"),
		"cidr_list":              "0.0.0.0/0",
		"exclude_cidr_list":      "",
		"extra_vault_field":      "should-be-ignored",
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestSSHSecretEngineRoleValidateUpdate_RejectsNameChange(t *testing.T) {
	r := &SSHSecretEngineRole{}
	_, err := r.ValidateUpdate(context.Background(),
		&SSHSecretEngineRole{Spec: SSHSecretEngineRoleSpec{Path: "ssh", Name: "old-name"}},
		&SSHSecretEngineRole{Spec: SSHSecretEngineRoleSpec{Path: "ssh", Name: "new-name"}},
	)
	if err == nil {
		t.Error("expected error when spec.name changes")
	}
	if err != nil && err.Error() != "spec.name cannot be updated" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSSHSecretEngineRoleValidateUpdate_AllowsSameNameAndPath(t *testing.T) {
	r := &SSHSecretEngineRole{}
	_, err := r.ValidateUpdate(context.Background(),
		&SSHSecretEngineRole{Spec: SSHSecretEngineRoleSpec{Path: "ssh", Name: "same-name"}},
		&SSHSecretEngineRole{Spec: SSHSecretEngineRoleSpec{Path: "ssh", Name: "same-name"}},
	)
	if err != nil {
		t.Errorf("expected no error when name and path unchanged, got: %v", err)
	}
}

func TestSSHSecretEngineRoleValidateUpdate_RejectsPathChange(t *testing.T) {
	r := &SSHSecretEngineRole{}
	_, err := r.ValidateUpdate(context.Background(),
		&SSHSecretEngineRole{Spec: SSHSecretEngineRoleSpec{Path: "old-ssh", Name: "role1"}},
		&SSHSecretEngineRole{Spec: SSHSecretEngineRoleSpec{Path: "new-ssh", Name: "role1"}},
	)
	if err == nil {
		t.Error("expected error when spec.path changes")
	}
	if err != nil && err.Error() != "spec.path cannot be updated" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSSHSecretEngineRoleIsDeletable(t *testing.T) {
	role := &SSHSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected SSHSecretEngineRole to be deletable")
	}
}

func TestSSHSecretEngineRoleConditions(t *testing.T) {
	role := &SSHSecretEngineRole{}

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
}
