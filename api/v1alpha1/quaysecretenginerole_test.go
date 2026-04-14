package v1alpha1

import (
	"encoding/json"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func permPtr(p Permission) *Permission { return &p }

func TestQuaySecretEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *QuaySecretEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &QuaySecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: QuaySecretEngineRoleSpec{
					Path: "quay",
					Name: "custom-name",
					QuayRole: QuayRole{
						QuayBaseRole: QuayBaseRole{
							NamespaceName: "n",
						},
					},
				},
			},
			expectedPath: "quay/roles/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &QuaySecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: QuaySecretEngineRoleSpec{
					Path: "quay",
					QuayRole: QuayRole{
						QuayBaseRole: QuayBaseRole{
							NamespaceName: "n",
						},
					},
				},
			},
			expectedPath: "quay/roles/meta-name",
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

func TestQuayRoleToMapAllKeys(t *testing.T) {
	createRepos := true
	role := QuayRole{
		QuayBaseRole: QuayBaseRole{
			NamespaceType:      NamespaceTypeOrganization,
			NamespaceName:      "myorg",
			CreateRepositories: &createRepos,
			DefaultPermission:  permPtr(PermissionRead),
			Teams:              &map[string]TeamRole{"team1": TeamRoleAdmin},
			Repositories:       &map[string]Permission{"repo1": PermissionWrite},
		},
		TTL:    &metav1.Duration{Duration: time.Hour},
		MaxTTL: &metav1.Duration{Duration: 24 * time.Hour},
	}

	result := role.toMap()

	expectedKeys := []string{
		"namespace_type", "namespace_name", "create_repositories",
		"default_permission", "teams", "repositories", "ttl", "max_ttl",
	}
	if len(result) != 8 {
		t.Fatalf("expected 8 keys in toMap() output, got %d", len(result))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	teamsStr, ok := result["teams"].(string)
	if !ok {
		t.Fatalf("teams should be a JSON string, got %T", result["teams"])
	}
	var teamsDecoded map[string]string
	if err := json.Unmarshal([]byte(teamsStr), &teamsDecoded); err != nil {
		t.Fatalf("teams JSON: %v", err)
	}
	if teamsDecoded["team1"] != string(TeamRoleAdmin) {
		t.Errorf("teams decoded = %v", teamsDecoded)
	}

	reposStr, ok := result["repositories"].(string)
	if !ok {
		t.Fatalf("repositories should be a JSON string, got %T", result["repositories"])
	}
	var reposDecoded map[string]string
	if err := json.Unmarshal([]byte(reposStr), &reposDecoded); err != nil {
		t.Fatalf("repositories JSON: %v", err)
	}
	if reposDecoded["repo1"] != string(PermissionWrite) {
		t.Errorf("repositories decoded = %v", reposDecoded)
	}
}

func TestQuayRoleToMapMinimalKeys(t *testing.T) {
	role := QuayRole{
		QuayBaseRole: QuayBaseRole{
			NamespaceType: NamespaceTypeOrganization,
			NamespaceName: "myorg",
		},
	}

	result := role.toMap()

	expectedKeys := []string{
		"namespace_type", "namespace_name", "create_repositories", "ttl", "max_ttl",
	}
	if len(result) != 5 {
		t.Fatalf("expected 5 keys in toMap() output, got %d", len(result))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}
	for _, absent := range []string{"default_permission", "teams", "repositories"} {
		if _, ok := result[absent]; ok {
			t.Errorf("expected key %q to be absent", absent)
		}
	}
}

func TestQuaySecretEngineRoleIsEquivalentMatching(t *testing.T) {
	createRepos := true
	role := &QuaySecretEngineRole{
		Spec: QuaySecretEngineRoleSpec{
			Path: "quay",
			QuayRole: QuayRole{
				QuayBaseRole: QuayBaseRole{
					NamespaceType:      NamespaceTypeOrganization,
					NamespaceName:      "myorg",
					CreateRepositories: &createRepos,
					DefaultPermission:  permPtr(PermissionRead),
					Teams:              &map[string]TeamRole{"team1": TeamRoleAdmin},
					Repositories:       &map[string]Permission{"repo1": PermissionWrite},
				},
				TTL:    &metav1.Duration{Duration: time.Hour},
				MaxTTL: &metav1.Duration{Duration: 24 * time.Hour},
			},
		},
	}

	payload := role.Spec.QuayRole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestQuaySecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &QuaySecretEngineRole{
		Spec: QuaySecretEngineRoleSpec{
			Path: "quay",
			QuayRole: QuayRole{
				QuayBaseRole: QuayBaseRole{
					NamespaceType: NamespaceTypeOrganization,
					NamespaceName: "myorg",
				},
			},
		},
	}

	payload := role.Spec.QuayRole.toMap()
	payload["namespace_name"] = "otherorg"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload to NOT be equivalent")
	}
}

func TestQuaySecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &QuaySecretEngineRole{
		Spec: QuaySecretEngineRoleSpec{
			Path: "quay",
			QuayRole: QuayRole{
				QuayBaseRole: QuayBaseRole{
					NamespaceType: NamespaceTypeOrganization,
					NamespaceName: "myorg",
				},
			},
		},
	}

	payload := role.Spec.QuayRole.toMap()
	payload["extra_vault_field"] = "some-value"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent")
	}
}

func TestQuaySecretEngineRoleIsDeletable(t *testing.T) {
	role := &QuaySecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected QuaySecretEngineRole to be deletable")
	}
}

func TestQuaySecretEngineRoleConditions(t *testing.T) {
	role := &QuaySecretEngineRole{}

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
