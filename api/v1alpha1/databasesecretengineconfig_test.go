package v1alpha1

import (
	"context"
	"reflect"
	"testing"
	"time"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDatabaseSecretEngineConfigGetPath(t *testing.T) {
	tests := []struct {
		name         string
		config       *DatabaseSecretEngineConfig
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			config: &DatabaseSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: DatabaseSecretEngineConfigSpec{
					Path: "database",
					Name: "custom-name",
				},
			},
			expectedPath: "database/config/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			config: &DatabaseSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: DatabaseSecretEngineConfigSpec{
					Path: "database",
				},
			},
			expectedPath: "database/config/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func TestDBSEConfigToMap(t *testing.T) {
	verifyConn := true
	config := DBSEConfig{
		PluginName:             "postgresql-database-plugin",
		PluginVersion:          "v1.0.0",
		VerifyConnection:       &verifyConn,
		AllowedRoles:           []string{"readonly", "readwrite"},
		RootRotationStatements: []string{"ALTER USER \"{{name}}\" WITH PASSWORD '{{password}}'"},
		PasswordPolicy:         "my-policy",
		ConnectionURL:          "postgresql://{{username}}:{{password}}@localhost:5432/mydb",
		DisableEscaping:        true,
	}

	result := config.toMap()

	expectedKeys := []string{
		"plugin_name", "plugin_version", "verify_connection",
		"allowed_roles", "root_credentials_rotate_statements",
		"password_policy", "connection_url", "disable_escaping",
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if result["plugin_name"] != "postgresql-database-plugin" {
		t.Errorf("plugin_name = %v, expected postgresql-database-plugin", result["plugin_name"])
	}
	if result["plugin_version"] != "v1.0.0" {
		t.Errorf("plugin_version = %v, expected v1.0.0", result["plugin_version"])
	}
	if result["verify_connection"] != true {
		t.Errorf("verify_connection = %v, expected true", result["verify_connection"])
	}
	if result["password_policy"] != "my-policy" {
		t.Errorf("password_policy = %v, expected my-policy", result["password_policy"])
	}
	if result["connection_url"] != "postgresql://{{username}}:{{password}}@localhost:5432/mydb" {
		t.Errorf("connection_url = %v, expected the connection URL", result["connection_url"])
	}
	if result["disable_escaping"] != true {
		t.Errorf("disable_escaping = %v, expected true", result["disable_escaping"])
	}
}

func TestDBSEConfigToMapDatabaseSpecificConfig(t *testing.T) {
	tests := []struct {
		name        string
		specific    map[string]string
		checkKeys   []string
		checkAbsent bool
	}{
		{
			name:      "non-nil map with entries merges into top level",
			specific:  map[string]string{"tls_ca": "/ca.pem", "tls_certificate_key": "/cert.pem"},
			checkKeys: []string{"tls_ca", "tls_certificate_key"},
		},
		{
			name:        "empty map adds no extra keys",
			specific:    map[string]string{},
			checkKeys:   []string{"tls_ca"},
			checkAbsent: true,
		},
		{
			name:        "nil map adds no extra keys and does not panic",
			specific:    nil,
			checkKeys:   []string{"tls_ca"},
			checkAbsent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DBSEConfig{
				PluginName:             "mongodb-database-plugin",
				DatabaseSpecificConfig: tt.specific,
			}
			result := config.toMap()

			for _, key := range tt.checkKeys {
				_, exists := result[key]
				if tt.checkAbsent && exists {
					t.Errorf("expected key %q to be absent, but it was present", key)
				}
				if !tt.checkAbsent && !exists {
					t.Errorf("expected key %q to be present, but it was absent", key)
				}
			}

			if !tt.checkAbsent {
				if result["tls_ca"] != "/ca.pem" {
					t.Errorf("tls_ca = %v, expected /ca.pem", result["tls_ca"])
				}
				if result["tls_certificate_key"] != "/cert.pem" {
					t.Errorf("tls_certificate_key = %v, expected /cert.pem", result["tls_certificate_key"])
				}
			}
		})
	}
}

func TestDBSEConfigToMapUsernameField(t *testing.T) {
	tests := []struct {
		name              string
		username          string
		retrievedUsername string
		expectKey         bool
		expectedValue     string
	}{
		{
			name:          "Username set uses it",
			username:      "admin",
			expectKey:     true,
			expectedValue: "admin",
		},
		{
			name:              "Username empty + retrievedUsername set uses retrieved",
			username:          "",
			retrievedUsername: "retrieved-user",
			expectKey:         true,
			expectedValue:     "retrieved-user",
		},
		{
			name:      "both empty produces no username key",
			username:  "",
			expectKey: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DBSEConfig{
				PluginName:        "postgresql-database-plugin",
				Username:          tt.username,
				retrievedUsername: tt.retrievedUsername,
			}
			result := config.toMap()
			val, exists := result["username"]
			if tt.expectKey && !exists {
				t.Error("expected 'username' key in map, but it was absent")
			}
			if !tt.expectKey && exists {
				t.Errorf("expected no 'username' key in map, but got %v", val)
			}
			if tt.expectKey && val != tt.expectedValue {
				t.Errorf("username = %v, expected %v", val, tt.expectedValue)
			}
		})
	}
}

func TestDBSEConfigToMapPasswordField(t *testing.T) {
	tests := []struct {
		name              string
		retrievedPassword string
		expectKey         bool
	}{
		{
			name:              "retrievedPassword set produces password key",
			retrievedPassword: "s3cret",
			expectKey:         true,
		},
		{
			name:              "retrievedPassword empty produces no password key",
			retrievedPassword: "",
			expectKey:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DBSEConfig{
				PluginName:        "postgresql-database-plugin",
				retrievedPassword: tt.retrievedPassword,
			}
			result := config.toMap()
			val, exists := result["password"]
			if tt.expectKey && !exists {
				t.Error("expected 'password' key in map, but it was absent")
			}
			if !tt.expectKey && exists {
				t.Errorf("expected no 'password' key in map, but got %v", val)
			}
			if tt.expectKey && val != "s3cret" {
				t.Errorf("password = %v, expected s3cret", val)
			}
		})
	}
}

func TestDBSEConfigToMapAllowedRolesTypeConversion(t *testing.T) {
	config := DBSEConfig{
		PluginName:   "postgresql-database-plugin",
		AllowedRoles: []string{"role1", "role2"},
	}
	result := config.toMap()

	roles, ok := result["allowed_roles"].([]any)
	if !ok {
		t.Fatalf("allowed_roles should be []any, got %T", result["allowed_roles"])
	}
	expected := []any{"role1", "role2"}
	if !reflect.DeepEqual(roles, expected) {
		t.Errorf("allowed_roles = %v, expected %v", roles, expected)
	}
}

func TestDatabaseSecretEngineConfigIsEquivalentRootPasswordRotationGate(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:    "postgresql-database-plugin",
				ConnectionURL: "postgresql://localhost/mydb",
				AllowedRoles:  []string{"*"},
				RootPasswordRotation: &RootPasswordRotation{
					Enable: true,
				},
			},
		},
		Status: DatabaseSecretEngineConfigStatus{
			LastRootPasswordRotation: metav1.Time{}, // zero value
		},
	}

	payload := map[string]any{"anything": "value"}
	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when RootPasswordRotation.Enable=true and LastRootPasswordRotation is zero")
	}
}

func TestDatabaseSecretEngineConfigIsEquivalentRootPasswordRotationWithTimestamp(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:    "postgresql-database-plugin",
				ConnectionURL: "postgresql://localhost/mydb",
				AllowedRoles:  []string{"*"},
				RootPasswordRotation: &RootPasswordRotation{
					Enable: true,
				},
			},
		},
		Status: DatabaseSecretEngineConfigStatus{
			LastRootPasswordRotation: metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	// Realistic Vault read response — Vault does NOT return verify_connection
	// or disable_escaping, and connection_details only contains connection_url
	// and username.
	connectionDetails := map[string]any{
		"connection_url": "postgresql://localhost/mydb",
		"username":       nil,
	}

	payload := map[string]any{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "",
		"allowed_roles":                      []any{"*"},
		"root_credentials_rotate_statements": []any{},
		"password_policy":                    "",
		"connection_details":                 connectionDetails,
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true when RootPasswordRotation.Enable=true with non-zero timestamp and matching payload")
	}
}

func TestDatabaseSecretEngineConfigIsEquivalentConnectionDetailsRemapping(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:             "postgresql-database-plugin",
				ConnectionURL:          "postgresql://localhost/mydb",
				Username:               "admin",
				DisableEscaping:        true,
				AllowedRoles:           []string{"*"},
				RootRotationStatements: []string{"ALTER USER"},
			},
		},
	}

	// Realistic Vault read response — connection_details only contains
	// connection_url and username. verify_connection and disable_escaping
	// are NOT returned by Vault.
	connectionDetails := map[string]any{
		"connection_url": "postgresql://localhost/mydb",
		"username":       "admin",
	}

	payload := map[string]any{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "",
		"allowed_roles":                      []any{"*"},
		"root_credentials_rotate_statements": []any{"ALTER USER"},
		"password_policy":                    "",
		"connection_details":                 connectionDetails,
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true when payload matches Vault's connection_details remapping structure")
	}
}

func TestDatabaseSecretEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:             "postgresql-database-plugin",
				PluginVersion:          "v1.0.0",
				VerifyConnection:       boolPtr(true),
				AllowedRoles:           []string{"*"},
				RootRotationStatements: []string{"ALTER USER"},
				PasswordPolicy:         "my-policy",
				ConnectionURL:          "postgresql://localhost/mydb",
				Username:               "admin",
				DisableEscaping:        false,
			},
		},
	}

	// Realistic Vault read response — verify_connection is NOT returned
	// (write-only), connection_details only has connection_url and username.
	connectionDetails := map[string]any{
		"connection_url": "postgresql://localhost/mydb",
		"username":       "admin",
	}

	payload := map[string]any{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "v1.0.0",
		"allowed_roles":                      []any{"*"},
		"root_credentials_rotate_statements": []any{"ALTER USER"},
		"password_policy":                    "my-policy",
		"connection_details":                 connectionDetails,
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for fully matching payload with connection_details structure")
	}
}

func TestDatabaseSecretEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:    "postgresql-database-plugin",
				AllowedRoles:  []string{"*"},
				ConnectionURL: "postgresql://localhost/mydb",
			},
		},
	}

	// Realistic Vault read response with a different plugin_name
	connectionDetails := map[string]any{
		"connection_url": "postgresql://localhost/mydb",
	}

	payload := map[string]any{
		"plugin_name":                        "mysql-database-plugin", // changed
		"plugin_version":                     "",
		"allowed_roles":                      []any{"*"},
		"root_credentials_rotate_statements": []any{},
		"password_policy":                    "",
		"connection_details":                 connectionDetails,
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when plugin_name differs")
	}
}

func TestDatabaseSecretEngineConfigIsEquivalentExtraFieldsFiltered(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:    "postgresql-database-plugin",
				AllowedRoles:  []string{"*"},
				ConnectionURL: "postgresql://localhost/mydb",
			},
		},
	}

	// Realistic Vault response with extra Vault-internal fields
	connectionDetails := map[string]any{
		"connection_url": "postgresql://localhost/mydb",
	}

	payload := map[string]any{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "",
		"allowed_roles":                      []any{"*"},
		"root_credentials_rotate_statements": []any{},
		"password_policy":                    "",
		"connection_details":                 connectionDetails,
		"backend":                            "extra-field-from-vault",
		"some_other_key":                     123,
	}

	// DatabaseSecretEngineConfig DOES filter extra fields from payload
	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: DatabaseSecretEngineConfig filters extra fields from payload before comparison")
	}
}

func TestDatabaseSecretEngineConfigIsEquivalentRootRotateStatementsRemainAtTopLevel(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:             "postgresql-database-plugin",
				AllowedRoles:           []string{"*"},
				ConnectionURL:          "postgresql://localhost/mydb",
				RootRotationStatements: []string{"ALTER USER \"{{name}}\""},
			},
		},
	}

	// Realistic Vault response: root_credentials_rotate_statements is at top
	// level, NOT inside connection_details.
	connectionDetails := map[string]any{
		"connection_url": "postgresql://localhost/mydb",
	}

	payload := map[string]any{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "",
		"allowed_roles":                      []any{"*"},
		"root_credentials_rotate_statements": []any{"ALTER USER \"{{name}}\""},
		"password_policy":                    "",
		"connection_details":                 connectionDetails,
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true when root_credentials_rotate_statements exists at top-level")
	}

	// Verify that removing it from top-level causes failure
	payloadMissingTopLevel := map[string]any{
		"plugin_name":        "postgresql-database-plugin",
		"plugin_version":     "",
		"allowed_roles":      []any{"*"},
		"password_policy":    "",
		"connection_details": connectionDetails,
	}

	// With symmetric comparison, when root_credentials_rotate_statements is absent
	// from the Vault payload, it is simply not compared (Vault doesn't have it).
	// The operator treats this as equivalent since it can't know Vault's state for
	// fields Vault doesn't return.
	if !config.IsEquivalentToDesiredState(payloadMissingTopLevel) {
		t.Error("expected true: symmetric comparison skips fields absent from Vault response")
	}
}

func TestDatabaseSecretEngineConfigIsDeletable(t *testing.T) {
	config := &DatabaseSecretEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected DatabaseSecretEngineConfig to be deletable")
	}
}

// TestDatabaseSecretEngineConfigIsEquivalent_Issue339_RealisticVaultResponse validates the
// fix for https://github.com/redhat-cop/vault-config-operator/issues/339.
// After root password rotation, a realistic Vault GET response (which omits write-only
// fields like verify_connection and has minimal connection_details) must be detected as
// equivalent so that the operator does NOT rewrite the config with the stale password.
func TestDatabaseSecretEngineConfigIsEquivalent_Issue339_RealisticVaultResponse(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:    "postgresql-database-plugin",
				ConnectionURL: "postgresql://{{username}}:{{password}}@my-postgresql.svc:5432",
				AllowedRoles:  []string{"read-only"},
				Username:      "postgres",
				RootPasswordRotation: &RootPasswordRotation{
					Enable: true,
				},
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "pg-creds"},
			},
		},
		Status: DatabaseSecretEngineConfigStatus{
			LastRootPasswordRotation: metav1.Time{Time: time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)},
			CredentialsHash:          "abc123",
		},
	}

	// Vault GET /database/config/{name} response — no verify_connection,
	// no disable_escaping, connection_details has only connection_url + username.
	vaultPayload := map[string]any{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "",
		"allowed_roles":                      []any{"read-only"},
		"root_credentials_rotate_statements": []any{},
		"password_policy":                    "",
		"connection_details": map[string]any{
			"connection_url": "postgresql://{{username}}:{{password}}@my-postgresql.svc:5432",
			"username":       "postgres",
		},
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: after root rotation the Vault response (without write-only fields) must be detected as equivalent to prevent stale password overwrite")
	}
}

// TestDatabaseSecretEngineConfigIsEquivalent_VerifyConnectionNilDefaultsToTrue ensures
// that when VerifyConnection is nil (user didn't set it), the operator sends true to Vault.
func TestDatabaseSecretEngineConfigIsEquivalent_VerifyConnectionNilDefaultsToTrue(t *testing.T) {
	config := DBSEConfig{
		PluginName:    "postgresql-database-plugin",
		ConnectionURL: "postgresql://localhost/mydb",
	}
	result := config.toMap()
	if result["verify_connection"] != true {
		t.Errorf("verify_connection should default to true when nil, got %v", result["verify_connection"])
	}
}

// TestDatabaseSecretEngineConfigIsEquivalent_VerifyConnectionExplicitFalse ensures
// that when VerifyConnection is explicitly set to false, the operator sends false.
func TestDatabaseSecretEngineConfigIsEquivalent_VerifyConnectionExplicitFalse(t *testing.T) {
	config := DBSEConfig{
		PluginName:       "postgresql-database-plugin",
		ConnectionURL:    "postgresql://localhost/mydb",
		VerifyConnection: boolPtr(false),
	}
	result := config.toMap()
	if result["verify_connection"] != false {
		t.Errorf("verify_connection should be false when explicitly set, got %v", result["verify_connection"])
	}
}

func TestDatabaseSecretEngineConfigConditions(t *testing.T) {
	config := &DatabaseSecretEngineConfig{}

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

func TestDatabaseSecretEngineConfigGetRootPasswordRotationPath(t *testing.T) {
	tests := []struct {
		name         string
		config       *DatabaseSecretEngineConfig
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			config: &DatabaseSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: DatabaseSecretEngineConfigSpec{
					Path: "database",
					Name: "custom-name",
				},
			},
			expectedPath: "database/rotate-root/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			config: &DatabaseSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: DatabaseSecretEngineConfigSpec{
					Path: "database",
				},
			},
			expectedPath: "database/rotate-root/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetRootPasswordRotationPath()
			if result != tt.expectedPath {
				t.Errorf("GetRootPasswordRotationPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestToInterfaceArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []any
	}{
		{
			name:     "non-empty slice",
			input:    []string{"a", "b"},
			expected: []any{"a", "b"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []any{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toInterfaceArray(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("toInterfaceArray(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
			if result == nil {
				t.Error("toInterfaceArray should return empty slice, not nil")
			}
		})
	}
}

func TestDatabaseSecretEngineConfig_PrepareInternalValues(t *testing.T) {
	const ns = "test-ns"

	tests := []struct {
		name     string
		setup    func(t *testing.T) (*DatabaseSecretEngineConfig, context.Context)
		wantUser string
		wantPass string
	}{
		{
			name: "secret branch username and password from K8s secret",
			setup: func(t *testing.T) (*DatabaseSecretEngineConfig, context.Context) {
				t.Helper()
				sec := newK8sSecret(ns, "db-cred", map[string][]byte{
					"username": []byte("db-user"),
					"password": []byte("db-pass"),
				})
				kube := newFakeKubeClient(sec)
				handler := newFakeVaultHandler()
				vc, ts := newFakeVaultClient(t, handler)
				t.Cleanup(ts.Close)
				instance := &DatabaseSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "db-config"},
					Spec: DatabaseSecretEngineConfigSpec{
						Path: "database",
						DBSEConfig: DBSEConfig{
							PluginName: "postgresql-database-plugin",
						},
						RootCredentials: vaultutils.RootCredentialConfig{
							Secret:      &corev1.LocalObjectReference{Name: "db-cred"},
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
				}
				return instance, pivContext(kube, vc)
			},
			wantUser: "db-user",
			wantPass: "db-pass",
		},
		{
			name: "secret branch spec username takes precedence",
			setup: func(t *testing.T) (*DatabaseSecretEngineConfig, context.Context) {
				t.Helper()
				sec := newK8sSecret(ns, "db-cred", map[string][]byte{
					"username": []byte("from-secret"),
					"password": []byte("db-pass"),
				})
				kube := newFakeKubeClient(sec)
				handler := newFakeVaultHandler()
				vc, ts := newFakeVaultClient(t, handler)
				t.Cleanup(ts.Close)
				instance := &DatabaseSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "db-config"},
					Spec: DatabaseSecretEngineConfigSpec{
						Path: "database",
						DBSEConfig: DBSEConfig{
							PluginName: "postgresql-database-plugin",
							Username:   "spec-user",
						},
						RootCredentials: vaultutils.RootCredentialConfig{
							Secret:      &corev1.LocalObjectReference{Name: "db-cred"},
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
				}
				return instance, pivContext(kube, vc)
			},
			wantUser: "spec-user",
			wantPass: "db-pass",
		},
		{
			name: "vault secret KV v1 username and password from vault",
			setup: func(t *testing.T) (*DatabaseSecretEngineConfig, context.Context) {
				t.Helper()
				kube := newFakeKubeClient()
				handler := newFakeVaultHandler()
				handler.setGet("secret/creds", map[string]any{
					"username": "vault-user",
					"password": "vault-pass",
				})
				vc, ts := newFakeVaultClient(t, handler)
				t.Cleanup(ts.Close)
				instance := &DatabaseSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "db-config"},
					Spec: DatabaseSecretEngineConfigSpec{
						Path: "database",
						DBSEConfig: DBSEConfig{
							PluginName: "postgresql-database-plugin",
						},
						RootCredentials: vaultutils.RootCredentialConfig{
							VaultSecret: &vaultutils.VaultSecretReference{Path: "secret/creds"},
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
				}
				return instance, pivContext(kube, vc)
			},
			wantUser: "vault-user",
			wantPass: "vault-pass",
		},
		{
			name: "vault secret spec username takes precedence",
			setup: func(t *testing.T) (*DatabaseSecretEngineConfig, context.Context) {
				t.Helper()
				kube := newFakeKubeClient()
				handler := newFakeVaultHandler()
				handler.setGet("secret/creds", map[string]any{
					"username": "vault-user",
					"password": "vault-pass",
				})
				vc, ts := newFakeVaultClient(t, handler)
				t.Cleanup(ts.Close)
				instance := &DatabaseSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "db-config"},
					Spec: DatabaseSecretEngineConfigSpec{
						Path: "database",
						DBSEConfig: DBSEConfig{
							PluginName: "postgresql-database-plugin",
							Username:   "override-user",
						},
						RootCredentials: vaultutils.RootCredentialConfig{
							VaultSecret: &vaultutils.VaultSecretReference{Path: "secret/creds"},
							UsernameKey: "username",
							PasswordKey: "password",
						},
					},
				}
				return instance, pivContext(kube, vc)
			},
			wantUser: "override-user",
			wantPass: "vault-pass",
		},
		{
			name: "random secret KV v1 password from secret key",
			setup: func(t *testing.T) (*DatabaseSecretEngineConfig, context.Context) {
				t.Helper()
				randomSecret := &RandomSecret{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "my-random"},
					Spec: RandomSecretSpec{
						Path:                vaultutils.Path("secret/random"),
						SecretKey:           "api_key",
						IsKVSecretsEngineV2: false,
					},
				}
				kube := newFakeKubeClient(randomSecret)
				handler := newFakeVaultHandler()
				handler.setGet("secret/random/my-random", map[string]any{
					"api_key": "rand-secret-value",
				})
				vc, ts := newFakeVaultClient(t, handler)
				t.Cleanup(ts.Close)
				instance := &DatabaseSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "db-config"},
					Spec: DatabaseSecretEngineConfigSpec{
						Path: "database",
						DBSEConfig: DBSEConfig{
							PluginName: "postgresql-database-plugin",
							Username:   "db-admin",
						},
						RootCredentials: vaultutils.RootCredentialConfig{
							RandomSecret: &corev1.LocalObjectReference{Name: "my-random"},
							UsernameKey:  "username",
							PasswordKey:  "password",
						},
					},
				}
				return instance, pivContext(kube, vc)
			},
			wantUser: "db-admin",
			wantPass: "rand-secret-value",
		},
		{
			name: "random secret KV v2 nested data password",
			setup: func(t *testing.T) (*DatabaseSecretEngineConfig, context.Context) {
				t.Helper()
				randomSecret := &RandomSecret{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "my-random"},
					Spec: RandomSecretSpec{
						Path:                vaultutils.Path("secret/random"),
						SecretKey:           "password",
						IsKVSecretsEngineV2: true,
					},
				}
				kube := newFakeKubeClient(randomSecret)
				handler := newFakeVaultHandler()
				handler.setGet("secret/random/my-random", map[string]any{
					"data": map[string]any{
						"password": "kv2-pw",
					},
				})
				vc, ts := newFakeVaultClient(t, handler)
				t.Cleanup(ts.Close)
				instance := &DatabaseSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: "db-config"},
					Spec: DatabaseSecretEngineConfigSpec{
						Path: "database",
						DBSEConfig: DBSEConfig{
							PluginName: "postgresql-database-plugin",
							Username:   "kv2-user",
						},
						RootCredentials: vaultutils.RootCredentialConfig{
							RandomSecret: &corev1.LocalObjectReference{Name: "my-random"},
							UsernameKey:  "username",
							PasswordKey:  "password",
						},
					},
				}
				return instance, pivContext(kube, vc)
			},
			wantUser: "kv2-user",
			wantPass: "kv2-pw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance, ctx := tt.setup(t)
			if err := instance.PrepareInternalValues(ctx, instance); err != nil {
				t.Fatalf("PrepareInternalValues: %v", err)
			}
			if instance.Spec.DBSEConfig.retrievedUsername != tt.wantUser {
				t.Errorf("retrievedUsername = %q, want %q", instance.Spec.DBSEConfig.retrievedUsername, tt.wantUser)
			}
			if instance.Spec.DBSEConfig.retrievedPassword != tt.wantPass {
				t.Errorf("retrievedPassword = %q, want %q", instance.Spec.DBSEConfig.retrievedPassword, tt.wantPass)
			}
		})
	}
}

func TestSetUsernameAndPassword_CredentialsChangedFlag(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Status: DatabaseSecretEngineConfigStatus{
			CredentialsHash: computeCredentialsHash("user", "old-pass"),
		},
	}

	config.SetUsernameAndPassword("user", "new-pass")

	if !config.Spec.DBSEConfig.credentialsChanged {
		t.Error("expected credentialsChanged=true when password hash changes")
	}
	if !config.Status.LastRootPasswordRotation.IsZero() {
		t.Error("expected LastRootPasswordRotation to be reset to zero")
	}
}

func TestSetUsernameAndPassword_SameCredentialsNotMarkedChanged(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Status: DatabaseSecretEngineConfigStatus{
			CredentialsHash: computeCredentialsHash("user", "same-pass"),
		},
	}

	config.SetUsernameAndPassword("user", "same-pass")

	if config.Spec.DBSEConfig.credentialsChanged {
		t.Error("expected credentialsChanged=false when credentials are unchanged")
	}
}

func TestSetUsernameAndPassword_FirstTimeNoFlag(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Status: DatabaseSecretEngineConfigStatus{
			CredentialsHash: "", // first reconcile, no hash yet
		},
	}

	config.SetUsernameAndPassword("user", "pass")

	if config.Spec.DBSEConfig.credentialsChanged {
		t.Error("expected credentialsChanged=false on first reconcile (empty hash)")
	}
}

func TestIsEquivalentToDesiredState_CredentialsChangedForcesRewrite(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:         "postgresql-database-plugin",
				ConnectionURL:      "postgresql://localhost/mydb",
				AllowedRoles:       []string{"*"},
				credentialsChanged: true,
			},
		},
	}

	payload := map[string]any{"anything": "value"}
	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when credentialsChanged=true, regardless of rotation settings")
	}
}

func TestIsEquivalentToDesiredState_CredentialsChangedNoRotation(t *testing.T) {
	config := &DatabaseSecretEngineConfig{
		Spec: DatabaseSecretEngineConfigSpec{
			Path: "database",
			DBSEConfig: DBSEConfig{
				PluginName:    "postgresql-database-plugin",
				ConnectionURL: "postgresql://localhost/mydb",
				AllowedRoles:  []string{"*"},
				RootPasswordRotation: &RootPasswordRotation{
					Enable: false,
				},
				credentialsChanged: true,
			},
		},
	}

	payload := map[string]any{"anything": "value"}
	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when credentialsChanged=true even with rotation disabled")
	}
}
