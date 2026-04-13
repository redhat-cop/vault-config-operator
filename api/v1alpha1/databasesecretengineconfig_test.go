package v1alpha1

import (
	"reflect"
	"testing"
	"time"

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

func TestDBSEConfigToMap(t *testing.T) {
	config := DBSEConfig{
		PluginName:             "postgresql-database-plugin",
		PluginVersion:          "v1.0.0",
		VerifyConnection:       true,
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

	roles, ok := result["allowed_roles"].([]interface{})
	if !ok {
		t.Fatalf("allowed_roles should be []interface{}, got %T", result["allowed_roles"])
	}
	expected := []interface{}{"role1", "role2"}
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

	payload := map[string]interface{}{"anything": "value"}
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

	// Independent Vault read fixture — no reuse of toMap()
	connectionDetails := map[string]interface{}{
		"connection_url":                     "postgresql://localhost/mydb",
		"disable_escaping":                   false,
		"root_credentials_rotate_statements": []interface{}{},
		"username":                           nil,
	}

	payload := map[string]interface{}{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "",
		"verify_connection":                  false,
		"allowed_roles":                      []interface{}{"*"},
		"root_credentials_rotate_statements": []interface{}{},
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

	// Build the payload as an independent Vault read response — no reuse of toMap()
	connectionDetails := map[string]interface{}{
		"connection_url":                     "postgresql://localhost/mydb",
		"disable_escaping":                   true,
		"root_credentials_rotate_statements": []interface{}{"ALTER USER"},
		"username":                           "admin",
	}

	payload := map[string]interface{}{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "",
		"verify_connection":                  false,
		"allowed_roles":                      []interface{}{"*"},
		"root_credentials_rotate_statements": []interface{}{"ALTER USER"},
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
				VerifyConnection:       true,
				AllowedRoles:           []string{"*"},
				RootRotationStatements: []string{"ALTER USER"},
				PasswordPolicy:         "my-policy",
				ConnectionURL:          "postgresql://localhost/mydb",
				Username:               "admin",
				DisableEscaping:        false,
			},
		},
	}

	// Independent Vault read fixture — no reuse of toMap() output.
	// allowed_roles built as []interface{} to mirror Vault JSON deserialization (AC2).
	connectionDetails := map[string]interface{}{
		"connection_url":                     "postgresql://localhost/mydb",
		"disable_escaping":                   false,
		"root_credentials_rotate_statements": []interface{}{"ALTER USER"},
		"username":                           "admin",
	}

	payload := map[string]interface{}{
		"plugin_name":                        "postgresql-database-plugin",
		"plugin_version":                     "v1.0.0",
		"verify_connection":                  true,
		"allowed_roles":                      []interface{}{"*"},
		"root_credentials_rotate_statements": []interface{}{"ALTER USER"},
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

	desiredState := config.Spec.DBSEConfig.toMap()
	connectionDetails := map[string]interface{}{
		"connection_url":                     desiredState["connection_url"],
		"disable_escaping":                   desiredState["disable_escaping"],
		"root_credentials_rotate_statements": desiredState["root_credentials_rotate_statements"],
		"username":                           desiredState["username"],
	}

	payload := map[string]interface{}{
		"plugin_name":                        "mysql-database-plugin", // changed
		"plugin_version":                     desiredState["plugin_version"],
		"verify_connection":                  desiredState["verify_connection"],
		"allowed_roles":                      desiredState["allowed_roles"],
		"root_credentials_rotate_statements": desiredState["root_credentials_rotate_statements"],
		"password_policy":                    desiredState["password_policy"],
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

	desiredState := config.Spec.DBSEConfig.toMap()
	connectionDetails := map[string]interface{}{
		"connection_url":                     desiredState["connection_url"],
		"disable_escaping":                   desiredState["disable_escaping"],
		"root_credentials_rotate_statements": desiredState["root_credentials_rotate_statements"],
		"username":                           desiredState["username"],
	}

	payload := map[string]interface{}{
		"plugin_name":                        desiredState["plugin_name"],
		"plugin_version":                     desiredState["plugin_version"],
		"verify_connection":                  desiredState["verify_connection"],
		"allowed_roles":                      desiredState["allowed_roles"],
		"root_credentials_rotate_statements": desiredState["root_credentials_rotate_statements"],
		"password_policy":                    desiredState["password_policy"],
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

	desiredState := config.Spec.DBSEConfig.toMap()
	connectionDetails := map[string]interface{}{
		"connection_url":                     desiredState["connection_url"],
		"disable_escaping":                   desiredState["disable_escaping"],
		"root_credentials_rotate_statements": desiredState["root_credentials_rotate_statements"],
		"username":                           desiredState["username"],
	}

	// root_credentials_rotate_statements must exist BOTH at top-level AND inside connection_details
	payload := map[string]interface{}{
		"plugin_name":                        desiredState["plugin_name"],
		"plugin_version":                     desiredState["plugin_version"],
		"verify_connection":                  desiredState["verify_connection"],
		"allowed_roles":                      desiredState["allowed_roles"],
		"root_credentials_rotate_statements": desiredState["root_credentials_rotate_statements"],
		"password_policy":                    desiredState["password_policy"],
		"connection_details":                 connectionDetails,
	}

	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true when root_credentials_rotate_statements exists both at top-level and inside connection_details")
	}

	// Verify that removing it from top-level causes failure
	payloadMissingTopLevel := map[string]interface{}{
		"plugin_name":        desiredState["plugin_name"],
		"plugin_version":     desiredState["plugin_version"],
		"verify_connection":  desiredState["verify_connection"],
		"allowed_roles":      desiredState["allowed_roles"],
		"password_policy":    desiredState["password_policy"],
		"connection_details": connectionDetails,
	}

	if config.IsEquivalentToDesiredState(payloadMissingTopLevel) {
		t.Error("expected false when root_credentials_rotate_statements is missing from top-level")
	}
}

func TestDatabaseSecretEngineConfigIsDeletable(t *testing.T) {
	config := &DatabaseSecretEngineConfig{}
	if !config.IsDeletable() {
		t.Error("expected DatabaseSecretEngineConfig to be deletable")
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
		expected []interface{}
	}{
		{
			name:     "non-empty slice",
			input:    []string{"a", "b"},
			expected: []interface{}{"a", "b"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []interface{}{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: []interface{}{},
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
