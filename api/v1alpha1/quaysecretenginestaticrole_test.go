package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestQuaySecretEngineStaticRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *QuaySecretEngineStaticRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &QuaySecretEngineStaticRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: QuaySecretEngineStaticRoleSpec{
					Path: "quay",
					Name: "custom-name",
					QuayBaseRole: QuayBaseRole{
						NamespaceName: "n",
					},
				},
			},
			expectedPath: "quay/static-roles/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &QuaySecretEngineStaticRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: QuaySecretEngineStaticRoleSpec{
					Path: "quay",
					QuayBaseRole: QuayBaseRole{
						NamespaceName: "n",
					},
				},
			},
			expectedPath: "quay/static-roles/meta-name",
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

func TestQuayBaseRoleToMapAllKeys(t *testing.T) {
	createRepos := true
	role := QuayBaseRole{
		NamespaceType:      NamespaceTypeOrganization,
		NamespaceName:      "myorg",
		CreateRepositories: &createRepos,
		DefaultPermission:  permPtr(PermissionRead),
		Teams:              &map[string]TeamRole{"team1": TeamRoleAdmin},
		Repositories:       &map[string]Permission{"repo1": PermissionWrite},
	}

	result := role.toMap()

	expectedKeys := []string{
		"namespace_type", "namespace_name", "create_repositories",
		"default_permission", "teams", "repositories",
	}
	if len(result) != 6 {
		t.Fatalf("expected 6 keys in toMap() output, got %d", len(result))
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

func TestQuayBaseRoleToMapMinimalKeys(t *testing.T) {
	role := QuayBaseRole{
		NamespaceType: NamespaceTypeOrganization,
		NamespaceName: "myorg",
	}

	result := role.toMap()

	expectedKeys := []string{"namespace_type", "namespace_name", "create_repositories"}
	if len(result) != 3 {
		t.Fatalf("expected 3 keys in toMap() output, got %d", len(result))
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

func TestQuaySecretEngineStaticRoleIsEquivalentMatching(t *testing.T) {
	createRepos := true
	role := &QuaySecretEngineStaticRole{
		Spec: QuaySecretEngineStaticRoleSpec{
			Path: "quay",
			QuayBaseRole: QuayBaseRole{
				NamespaceType:      NamespaceTypeOrganization,
				NamespaceName:      "myorg",
				CreateRepositories: &createRepos,
				DefaultPermission:  permPtr(PermissionRead),
				Teams:              &map[string]TeamRole{"team1": TeamRoleAdmin},
				Repositories:       &map[string]Permission{"repo1": PermissionWrite},
			},
		},
	}

	payload := role.Spec.QuayBaseRole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestQuaySecretEngineStaticRoleIsEquivalentNonMatching(t *testing.T) {
	role := &QuaySecretEngineStaticRole{
		Spec: QuaySecretEngineStaticRoleSpec{
			Path: "quay",
			QuayBaseRole: QuayBaseRole{
				NamespaceType: NamespaceTypeOrganization,
				NamespaceName: "myorg",
			},
		},
	}

	payload := role.Spec.QuayBaseRole.toMap()
	payload["namespace_name"] = "otherorg"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload to NOT be equivalent")
	}
}

func TestQuaySecretEngineStaticRoleIsEquivalentExtraFields(t *testing.T) {
	role := &QuaySecretEngineStaticRole{
		Spec: QuaySecretEngineStaticRoleSpec{
			Path: "quay",
			QuayBaseRole: QuayBaseRole{
				NamespaceType: NamespaceTypeOrganization,
				NamespaceName: "myorg",
			},
		},
	}

	payload := role.Spec.QuayBaseRole.toMap()
	payload["extra_vault_field"] = "some-value"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent")
	}
}

func TestQuaySecretEngineStaticRoleIsDeletable(t *testing.T) {
	role := &QuaySecretEngineStaticRole{}
	if !role.IsDeletable() {
		t.Error("expected QuaySecretEngineStaticRole to be deletable")
	}
}

func TestQuaySecretEngineStaticRoleConditions(t *testing.T) {
	role := &QuaySecretEngineStaticRole{}

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
