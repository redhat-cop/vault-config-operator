package v1alpha1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDatabaseSecretEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *DatabaseSecretEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &DatabaseSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: DatabaseSecretEngineRoleSpec{
					Path: "database",
					Name: "custom-name",
				},
			},
			expectedPath: "database/roles/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &DatabaseSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: DatabaseSecretEngineRoleSpec{
					Path: "database",
				},
			},
			expectedPath: "database/roles/meta-name",
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

func TestDBSERoleToMap(t *testing.T) {
	role := DBSERole{
		DBName:               "my-db",
		DefaultTTL:           metav1.Duration{Duration: 1 * time.Hour},
		MaxTTL:               metav1.Duration{Duration: 24 * time.Hour},
		CreationStatements:   []string{"CREATE ROLE \"{{name}}\""},
		RevocationStatements: []string{"DROP ROLE \"{{name}}\""},
		RollbackStatements:   []string{"DROP ROLE \"{{name}}\""},
		RenewStatements:      []string{"ALTER ROLE \"{{name}}\""},
	}

	result := role.toMap()

	expectedKeys := []string{
		"db_name", "default_ttl", "max_ttl",
		"creation_statements", "revocation_statements",
		"rollback_statements", "renew_statements",
	}

	if len(result) != 7 {
		t.Errorf("expected 7 keys in toMap() output, got %d", len(result))
	}

	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if result["db_name"] != "my-db" {
		t.Errorf("db_name = %v, expected my-db", result["db_name"])
	}
}

func TestDBSERoleToMapDurationTypes(t *testing.T) {
	role := DBSERole{
		DefaultTTL: metav1.Duration{Duration: 30 * time.Minute},
		MaxTTL:     metav1.Duration{Duration: 2 * time.Hour},
	}

	result := role.toMap()

	defaultTTL, ok := result["default_ttl"].(metav1.Duration)
	if !ok {
		t.Fatalf("default_ttl should be metav1.Duration, got %T", result["default_ttl"])
	}
	if defaultTTL.Duration != 30*time.Minute {
		t.Errorf("default_ttl = %v, expected 30m", defaultTTL.Duration)
	}

	maxTTL, ok := result["max_ttl"].(metav1.Duration)
	if !ok {
		t.Fatalf("max_ttl should be metav1.Duration, got %T", result["max_ttl"])
	}
	if maxTTL.Duration != 2*time.Hour {
		t.Errorf("max_ttl = %v, expected 2h", maxTTL.Duration)
	}
}

func TestDatabaseSecretEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &DatabaseSecretEngineRole{
		Spec: DatabaseSecretEngineRoleSpec{
			Path: "database",
			DBSERole: DBSERole{
				DBName:               "my-db",
				DefaultTTL:           metav1.Duration{Duration: 1 * time.Hour},
				MaxTTL:               metav1.Duration{Duration: 24 * time.Hour},
				CreationStatements:   []string{"CREATE ROLE"},
				RevocationStatements: []string{"DROP ROLE"},
				RollbackStatements:   []string{"ROLLBACK"},
				RenewStatements:      []string{"RENEW"},
			},
		},
	}

	payload := role.Spec.DBSERole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestDatabaseSecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &DatabaseSecretEngineRole{
		Spec: DatabaseSecretEngineRoleSpec{
			Path: "database",
			DBSERole: DBSERole{
				DBName:             "my-db",
				DefaultTTL:         metav1.Duration{Duration: 1 * time.Hour},
				MaxTTL:             metav1.Duration{Duration: 24 * time.Hour},
				CreationStatements: []string{"CREATE ROLE"},
			},
		},
	}

	payload := role.Spec.DBSERole.toMap()
	payload["db_name"] = "other-db" // changed

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different db_name) to NOT be equivalent")
	}
}

// DatabaseSecretEngineRole uses bare reflect.DeepEqual without filtering, so
// extra keys in the payload cause a mismatch.
func TestDatabaseSecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &DatabaseSecretEngineRole{
		Spec: DatabaseSecretEngineRoleSpec{
			Path: "database",
			DBSERole: DBSERole{
				DBName:             "my-db",
				DefaultTTL:         metav1.Duration{Duration: 1 * time.Hour},
				MaxTTL:             metav1.Duration{Duration: 24 * time.Hour},
				CreationStatements: []string{"CREATE ROLE"},
			},
		},
	}

	payload := role.Spec.DBSERole.toMap()
	payload["extra_vault_field"] = "some-value"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare reflect.DeepEqual, no filtering)")
	}
}

// DBSERole.toMap() stores statement fields as []string. Vault may return them
// as []interface{}. reflect.DeepEqual treats these as different types, so the
// comparison fails. This documents the known type-skew behavior.
func TestDatabaseSecretEngineRoleIsEquivalentStatementTypeSkew(t *testing.T) {
	role := &DatabaseSecretEngineRole{
		Spec: DatabaseSecretEngineRoleSpec{
			Path: "database",
			DBSERole: DBSERole{
				DBName:             "my-db",
				DefaultTTL:         metav1.Duration{Duration: 1 * time.Hour},
				MaxTTL:             metav1.Duration{Duration: 24 * time.Hour},
				CreationStatements: []string{"CREATE ROLE"},
			},
		},
	}

	payload := map[string]interface{}{
		"db_name":               "my-db",
		"default_ttl":          metav1.Duration{Duration: 1 * time.Hour},
		"max_ttl":              metav1.Duration{Duration: 24 * time.Hour},
		"creation_statements":   []interface{}{"CREATE ROLE"},
		"revocation_statements": []interface{}{},
		"rollback_statements":   []interface{}{},
		"renew_statements":      []interface{}{},
	}

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected false: []string from toMap() != []interface{} from Vault JSON (bare DeepEqual, no type coercion)")
	}
}

func TestDatabaseSecretEngineRoleIsDeletable(t *testing.T) {
	role := &DatabaseSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected DatabaseSecretEngineRole to be deletable")
	}
}

func TestDatabaseSecretEngineRoleConditions(t *testing.T) {
	role := &DatabaseSecretEngineRole{}

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
