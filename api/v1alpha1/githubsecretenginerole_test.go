package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGitHubSecretEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *GitHubSecretEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &GitHubSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: GitHubSecretEngineRoleSpec{
					Path: "github",
					Name: "custom-name",
				},
			},
			expectedPath: "github/permissionset/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &GitHubSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: GitHubSecretEngineRoleSpec{
					Path: "github",
				},
			},
			expectedPath: "github/permissionset/meta-name",
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

func TestPermissionSetToMap(t *testing.T) {
	ps := PermissionSet{
		InstallationID:   123,
		OrganizationName: "my-org",
		Repositories:     []string{"repo1", "repo2"},
		RepositoriesIDs:  []string{"111", "222"},
		Permissions:      map[string]string{"contents": "read", "pull_requests": "write"},
	}

	result := ps.toMap()

	expectedKeys := []string{
		"installation_id", "org_name", "repositories", "repository_ids", "permissions",
	}
	if len(result) != 5 {
		t.Errorf("expected 5 keys in toMap() output, got %d", len(result))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	iid, ok := result["installation_id"].(int64)
	if !ok || iid != 123 {
		t.Errorf("installation_id = %v (%T), expected int64 123", result["installation_id"], result["installation_id"])
	}
	if on, ok := result["org_name"].(string); !ok || on != "my-org" {
		t.Errorf("org_name = %v (%T), expected string my-org", result["org_name"], result["org_name"])
	}
	repos, ok := result["repositories"].([]string)
	if !ok {
		t.Fatalf("repositories should be []string, got %T", result["repositories"])
	}
	if len(repos) != 2 || repos[0] != "repo1" || repos[1] != "repo2" {
		t.Errorf("repositories = %v, expected [repo1 repo2]", repos)
	}
	ids, ok := result["repository_ids"].([]string)
	if !ok {
		t.Fatalf("repository_ids should be []string, got %T", result["repository_ids"])
	}
	if len(ids) != 2 || ids[0] != "111" || ids[1] != "222" {
		t.Errorf("repository_ids = %v, expected [111 222]", ids)
	}
	perms, ok := result["permissions"].(map[string]string)
	if !ok {
		t.Fatalf("permissions should be map[string]string, got %T", result["permissions"])
	}
	if perms["contents"] != "read" || perms["pull_requests"] != "write" {
		t.Errorf("permissions = %v, unexpected map contents", perms)
	}
}

func TestGitHubSecretEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &GitHubSecretEngineRole{
		Spec: GitHubSecretEngineRoleSpec{
			Path: "github",
			PermissionSet: PermissionSet{
				InstallationID:   123,
				OrganizationName: "my-org",
				Repositories:     []string{"repo1", "repo2"},
				RepositoriesIDs:  []string{"111", "222"},
				Permissions:      map[string]string{"contents": "read", "pull_requests": "write"},
			},
		},
	}

	payload := map[string]interface{}{
		"installation_id": int64(123),
		"org_name":        "my-org",
		"repositories":    []string{"repo1", "repo2"},
		"repository_ids":  []string{"111", "222"},
		"permissions":     map[string]string{"contents": "read", "pull_requests": "write"},
	}

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestGitHubSecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &GitHubSecretEngineRole{
		Spec: GitHubSecretEngineRoleSpec{
			Path: "github",
			PermissionSet: PermissionSet{
				InstallationID:   123,
				OrganizationName: "my-org",
				Repositories:     []string{"repo1", "repo2"},
				RepositoriesIDs:  []string{"111", "222"},
				Permissions:      map[string]string{"contents": "read", "pull_requests": "write"},
			},
		},
	}

	payload := role.Spec.PermissionSet.toMap()
	payload["org_name"] = "other-org"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different org_name) to NOT be equivalent")
	}
}

func TestGitHubSecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &GitHubSecretEngineRole{
		Spec: GitHubSecretEngineRoleSpec{
			Path: "github",
			PermissionSet: PermissionSet{
				InstallationID: 123,
			},
		},
	}

	payload := role.Spec.PermissionSet.toMap()
	payload["extra_vault_field"] = "some-value"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare reflect.DeepEqual, no filtering)")
	}
}

func TestGitHubSecretEngineRoleIsDeletable(t *testing.T) {
	role := &GitHubSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected GitHubSecretEngineRole to be deletable")
	}
}

func TestGitHubSecretEngineRoleConditions(t *testing.T) {
	role := &GitHubSecretEngineRole{}

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
