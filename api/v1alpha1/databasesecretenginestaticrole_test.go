package v1alpha1

import (
	"reflect"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func staticRoleIntPtr(v int) *int { return &v }

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

func TestDBSEStaticRoleToMapWithRotationPeriod(t *testing.T) {
	role := DBSEStaticRole{
		DBName:             "my-db",
		Username:           "static-user",
		RotationPeriod:     staticRoleIntPtr(86400),
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
	if _, ok := result["rotation_schedule"]; ok {
		t.Error("expected no 'rotation_schedule' key when not set")
	}
	if _, ok := result["rotation_window"]; ok {
		t.Error("expected no 'rotation_window' key when not set")
	}
}

func TestDBSEStaticRoleToMapWithRotationSchedule(t *testing.T) {
	role := DBSEStaticRole{
		DBName:           "my-db",
		Username:         "static-user",
		RotationSchedule: "0 0 * * SAT",
		RotationWindow:   staticRoleIntPtr(3600),
		CredentialType:   "password",
	}

	result := role.toMap()

	if result["rotation_schedule"] != "0 0 * * SAT" {
		t.Errorf("rotation_schedule = %v, expected '0 0 * * SAT'", result["rotation_schedule"])
	}
	if result["rotation_window"] != 3600 {
		t.Errorf("rotation_window = %v, expected 3600", result["rotation_window"])
	}
	if _, ok := result["rotation_period"]; ok {
		t.Error("expected no 'rotation_period' key when rotation_schedule is set")
	}
}

func TestDBSEStaticRoleToMapWithRotationScheduleNoWindow(t *testing.T) {
	role := DBSEStaticRole{
		DBName:           "my-db",
		Username:         "static-user",
		RotationSchedule: "0 0 * * SAT",
		CredentialType:   "password",
	}

	result := role.toMap()

	if result["rotation_schedule"] != "0 0 * * SAT" {
		t.Errorf("rotation_schedule = %v, expected '0 0 * * SAT'", result["rotation_schedule"])
	}
	if _, ok := result["rotation_window"]; ok {
		t.Error("expected no 'rotation_window' key when not set")
	}
	if _, ok := result["rotation_period"]; ok {
		t.Error("expected no 'rotation_period' key when rotation_schedule is set")
	}
}

func TestDBSEStaticRoleToMapPasswordCredentialConfig(t *testing.T) {
	role := DBSEStaticRole{
		DBName:         "my-db",
		Username:       "static-user",
		RotationPeriod: staticRoleIntPtr(86400),
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
		RotationPeriod: staticRoleIntPtr(86400),
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
		RotationPeriod: staticRoleIntPtr(86400),
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
				RotationPeriod:     staticRoleIntPtr(86400),
				RotationStatements: []string{"ALTER USER"},
				CredentialType:     "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}

	payload := role.Spec.DBSEStaticRole.toMap()

	if !vaultutils.IsEquivalentToDesiredState(role, payload) {
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
				RotationPeriod: staticRoleIntPtr(86400),
				CredentialType: "password",
			},
		},
	}

	payload := role.Spec.DBSEStaticRole.toMap()
	payload["username"] = "different-user" // changed

	if vaultutils.IsEquivalentToDesiredState(role, payload) {
		t.Error("expected non-matching payload (different username) to NOT be equivalent")
	}
}

func TestDatabaseSecretEngineStaticRoleIsEquivalentExtraFields(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			Path: "database",
			DBSEStaticRole: DBSEStaticRole{
				DBName:         "my-db",
				Username:       "static-user",
				RotationPeriod: staticRoleIntPtr(86400),
				CredentialType: "password",
			},
		},
	}

	payload := role.Spec.DBSEStaticRole.toMap()
	payload["extra_vault_field"] = "some-value"

	if !vaultutils.IsEquivalentToDesiredState(role, payload) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestDatabaseSecretEngineStaticRoleIsEquivalentCredentialConfigTypeMismatch(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			Path: "database",
			DBSEStaticRole: DBSEStaticRole{
				DBName:         "my-db",
				Username:       "static-user",
				RotationPeriod: staticRoleIntPtr(86400),
				CredentialType: "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}

	payload := map[string]any{
		"db_name":             "my-db",
		"username":            "static-user",
		"rotation_period":     86400,
		"rotation_statements": []string(nil),
		"credential_type":     "password",
		"credential_config":   map[string]any{"password_policy": "my-policy"},
	}

	if vaultutils.IsEquivalentToDesiredState(role, payload) {
		t.Error("expected false: map[string]string != map[string]any in reflect.DeepEqual")
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

	if result["key_bits"] != "4096" {
		t.Errorf("key_bits should be string '4096', got %v", result["key_bits"])
	}
}

// Validation tests

func TestValidateRotationConfig_PeriodOnly(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			DBSEStaticRole: DBSEStaticRole{
				DBName:         "my-db",
				Username:       "user",
				RotationPeriod: staticRoleIntPtr(86400),
				CredentialType: "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}
	valid, err := role.IsValid()
	if !valid || err != nil {
		t.Errorf("expected valid, got err: %v", err)
	}
}

func TestValidateRotationConfig_ScheduleOnly(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			DBSEStaticRole: DBSEStaticRole{
				DBName:           "my-db",
				Username:         "user",
				RotationSchedule: "0 0 * * SAT",
				CredentialType:   "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}
	valid, err := role.IsValid()
	if !valid || err != nil {
		t.Errorf("expected valid, got err: %v", err)
	}
}

func TestValidateRotationConfig_ScheduleWithWindow(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			DBSEStaticRole: DBSEStaticRole{
				DBName:           "my-db",
				Username:         "user",
				RotationSchedule: "0 0 * * SAT",
				RotationWindow:   staticRoleIntPtr(3600),
				CredentialType:   "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}
	valid, err := role.IsValid()
	if !valid || err != nil {
		t.Errorf("expected valid, got err: %v", err)
	}
}

func TestValidateRotationConfig_BothPeriodAndSchedule(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			DBSEStaticRole: DBSEStaticRole{
				DBName:           "my-db",
				Username:         "user",
				RotationPeriod:   staticRoleIntPtr(86400),
				RotationSchedule: "0 0 * * SAT",
				CredentialType:   "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}
	valid, err := role.IsValid()
	if valid {
		t.Error("expected invalid when both rotationPeriod and rotationSchedule are set")
	}
	if err == nil {
		t.Error("expected error when both rotationPeriod and rotationSchedule are set")
	}
}

func TestValidateRotationConfig_NeitherPeriodNorSchedule(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			DBSEStaticRole: DBSEStaticRole{
				DBName:         "my-db",
				Username:       "user",
				CredentialType: "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}
	valid, err := role.IsValid()
	if valid {
		t.Error("expected invalid when neither rotationPeriod nor rotationSchedule is set")
	}
	if err == nil {
		t.Error("expected error when neither rotationPeriod nor rotationSchedule is set")
	}
}

func TestValidateRotationConfig_WindowWithoutSchedule(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			DBSEStaticRole: DBSEStaticRole{
				DBName:         "my-db",
				Username:       "user",
				RotationPeriod: staticRoleIntPtr(86400),
				RotationWindow: staticRoleIntPtr(3600),
				CredentialType: "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}
	valid, err := role.IsValid()
	if valid {
		t.Error("expected invalid when rotationWindow is set without rotationSchedule")
	}
	if err == nil {
		t.Error("expected error when rotationWindow is set without rotationSchedule")
	}
}

func TestDatabaseSecretEngineStaticRoleIsEquivalentWithSchedule(t *testing.T) {
	role := &DatabaseSecretEngineStaticRole{
		Spec: DatabaseSecretEngineStaticRoleSpec{
			Path: "database",
			DBSEStaticRole: DBSEStaticRole{
				DBName:           "my-db",
				Username:         "static-user",
				RotationSchedule: "0 0 * * SAT",
				RotationWindow:   staticRoleIntPtr(3600),
				CredentialType:   "password",
				PasswordCredentialConfig: &PasswordCredentialConfig{
					PasswordPolicy: "my-policy",
				},
			},
		},
	}

	payload := role.Spec.DBSEStaticRole.toMap()

	if !vaultutils.IsEquivalentToDesiredState(role, payload) {
		t.Error("expected matching payload with rotation_schedule to be equivalent")
	}
}
