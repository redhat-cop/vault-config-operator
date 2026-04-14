package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDatabaseSecretEngineStaticRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *DatabaseSecretEngineStaticRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &DatabaseSecretEngineStaticRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: DatabaseSecretEngineStaticRoleSpec{
					Path: "database",
					Name: "custom-name",
				},
			},
			expectedPath: "database/static-roles/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &DatabaseSecretEngineStaticRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: DatabaseSecretEngineStaticRoleSpec{
					Path: "database",
				},
			},
			expectedPath: "database/static-roles/meta-name",
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

func TestDBSEStaticRoleToMap(t *testing.T) {
	role := DBSEStaticRole{
		DBName:             "my-db",
		Username:           "static-user",
		RotationPeriod:     86400,
		RotationStatements: []string{"ALTER USER"},
		CredentialType:     "password",
	}

	result := role.toMap()

	expectedKeys := []string{
		"db_name", "username", "rotation_period",
		"rotation_statements", "credential_type",
	}

	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if result["db_name"] != "my-db" {
		t.Errorf("db_name = %v, expected my-db", result["db_name"])
	}
	if result["username"] != "static-user" {
		t.Errorf("username = %v, expected static-user", result["username"])
	}
	if result["rotation_period"] != 86400 {
		t.Errorf("rotation_period = %v, expected 86400", result["rotation_period"])
	}
	if result["credential_type"] != "password" {
		t.Errorf("credential_type = %v, expected password", result["credential_type"])
	}

	if _, ok := result["credential_config"]; ok {
		t.Error("expected no 'credential_config' key when both config pointers are nil")
	}
}

func TestDBSEStaticRoleToMapPasswordCredentialConfig(t *testing.T) {
	role := DBSEStaticRole{
		DBName:         "my-db",
		Username:       "static-user",
		RotationPeriod: 86400,
		CredentialType: "password",
		PasswordCredentialConfig: &PasswordCredentialConfig{
			PasswordPolicy: "my-password-policy",
		},
	}

	result := role.toMap()

	credConfig, ok := result["credential_config"]
	if !ok {
		t.Fatal("expected 'credential_config' key in toMap() output")
	}

	expected := map[string]string{"password_policy": "my-password-policy"}
	if !reflect.DeepEqual(credConfig, expected) {
		t.Errorf("credential_config = %v, expected %v", credConfig, expected)
	}
}

func TestDBSEStaticRoleToMapRSACredentialConfig(t *testing.T) {
	role := DBSEStaticRole{
		DBName:         "my-db",
		Username:       "static-user",
		RotationPeriod: 86400,
		CredentialType: "rsa_private_key",
		RSAPrivateKeyCredentialConfig: &RSAPrivateKeyCredentialConfig{
			KeyBits: 2048,
			Format:  "pkcs8",
		},
	}

	result := role.toMap()

	credConfig, ok := result["credential_config"]
	if !ok {
		t.Fatal("expected 'credential_config' key in toMap() output")
	}

	expected := map[string]string{"key_bits": "2048", "format": "pkcs8"}
	if !reflect.DeepEqual(credConfig, expected) {
		t.Errorf("credential_config = %v, expected %v", credConfig, expected)
	}
}

func TestDBSEStaticRoleToMapNoCredentialConfig(t *testing.T) {
	role := DBSEStaticRole{
		DBName:         "my-db",
		Username:       "static-user",
		RotationPeriod: 86400,
		CredentialType: "password",
	}

	result := role.toMap()

	if _, ok := result["credential_config"]; ok {
		t.Error("expected no 'credential_config' key when both configs are nil")
	}
}

func TestDatabaseSecretEngineStaticRoleIsEquivalentMatching(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			Path: "database",
			DBSEStaticRole: DBSEStaticRole{
				DBName:             "my-db",
				Username:           "static-user",
				RotationPeriod:     86400,
				RotationStatements: []string{"ALTER USER"},
				CredentialType:     "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}

	payload := role.Spec.DBSEStaticRole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestDatabaseSecretEngineStaticRoleIsEquivalentNonMatching(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			Path: "database",
			DBSEStaticRole: DBSEStaticRole{
				DBName:         "my-db",
				Username:       "static-user",
				RotationPeriod: 86400,
				CredentialType: "password",
			},
		},
	}

	payload := role.Spec.DBSEStaticRole.toMap()
	payload["username"] = "different-user" // changed

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different username) to NOT be equivalent")
	}
}

// DatabaseSecretEngineStaticRole uses bare reflect.DeepEqual without filtering,
// so extra keys in the payload cause a mismatch.
func TestDatabaseSecretEngineStaticRoleIsEquivalentExtraFields(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			Path: "database",
			DBSEStaticRole: DBSEStaticRole{
				DBName:         "my-db",
				Username:       "static-user",
				RotationPeriod: 86400,
				CredentialType: "password",
			},
		},
	}

	payload := role.Spec.DBSEStaticRole.toMap()
	payload["extra_vault_field"] = "some-value"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare reflect.DeepEqual, no filtering)")
	}
}

// Documents that credential_config as map[string]string vs map[string]interface{}
// causes a DeepEqual mismatch due to Go type differences.
func TestDatabaseSecretEngineStaticRoleIsEquivalentCredentialConfigTypeMismatch(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			Path: "database",
			DBSEStaticRole: DBSEStaticRole{
				DBName:         "my-db",
				Username:       "static-user",
				RotationPeriod: 86400,
				CredentialType: "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}

	// Build a payload that matches in values but uses map[string]interface{}
	// instead of map[string]string for credential_config
	payload := map[string]interface{}{
		"db_name":             "my-db",
		"username":            "static-user",
		"rotation_period":     86400,
		"rotation_statements": []string(nil),
		"credential_type":     "password",
		"credential_config":   map[string]interface{}{"password_policy": "my-policy"},
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected false: map[string]string != map[string]interface{} in reflect.DeepEqual")
	}
}

func TestDatabaseSecretEngineStaticRoleIsDeletable(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{}
	if !role.IsDeletable() {
		t.Error("expected DatabaseSecretEngineStaticRole to be deletable")
	}
}

func TestDatabaseSecretEngineStaticRoleConditions(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{}

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

func TestPasswordCredentialConfigToMap(t *testing.T) {
	config := &PasswordCredentialConfig{
		PasswordPolicy: "my-policy",
	}

	result := config.toMap()
	expected := map[string]string{"password_policy": "my-policy"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() = %v, expected %v", result, expected)
	}
}

func TestRSAPrivateKeyCredentialConfigToMap(t *testing.T) {
	config := &RSAPrivateKeyCredentialConfig{
		KeyBits: 4096,
		Format:  "pkcs8",
	}

	result := config.toMap()
	expected := map[string]string{"key_bits": "4096", "format": "pkcs8"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() = %v, expected %v", result, expected)
	}

	// Verify key_bits is string (from strconv.Itoa), not int
	if result["key_bits"] != "4096" {
		t.Errorf("key_bits should be string '4096', got %v", result["key_bits"])
	}
}
