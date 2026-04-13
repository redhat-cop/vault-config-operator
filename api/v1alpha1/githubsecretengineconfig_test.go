package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGitHubSecretEngineConfigGetPath(t *testing.T) {
	config := &GitHubSecretEngineConfig{
		Spec: GitHubSecretEngineConfigSpec{
			Path: "github",
		},
	}
	if got := config.GetPath(); got != "github/config" {
		t.Errorf("GetPath() = %q, expected %q", got, "github/config")
	}
}

func TestGHConfigToMap(t *testing.T) {
	config := GHConfig{
		ApplicationID:    12345,
		GitHubAPIBaseURL: "https://api.github.com",
		retrievedSSHKey:  "test-ssh-key",
	}

	result := config.toMap()

	if len(result) != 3 {
		t.Fatalf("expected 3 keys in toMap() output, got %d", len(result))
	}
	for _, key := range []string{"app_id", "prv_key", "base_url"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if aid, ok := result["app_id"].(int64); !ok || aid != 12345 {
		t.Errorf("app_id = %v (%T), expected int64 12345", result["app_id"], result["app_id"])
	}
	if pk, ok := result["prv_key"].(string); !ok || pk != "test-ssh-key" {
		t.Errorf("prv_key = %v (%T), expected string test-ssh-key", result["prv_key"], result["prv_key"])
	}
	if bu, ok := result["base_url"].(string); !ok || bu != "https://api.github.com" {
		t.Errorf("base_url = %v (%T), expected string", result["base_url"], result["base_url"])
	}
}

func TestGitHubSecretEngineConfigIsEquivalentPrvKeyDeleted(t *testing.T) {
	config := &GitHubSecretEngineConfig{
		Spec: GitHubSecretEngineConfigSpec{
			Path: "github",
			GHConfig: GHConfig{
				ApplicationID:    12345,
				GitHubAPIBaseURL: "https://api.github.com",
				retrievedSSHKey:  "test-ssh-key",
			},
		},
	}

	payload := map[string]interface{}{
		"app_id":   int64(12345),
		"base_url": "https://api.github.com",
	}

	if _, has := payload["prv_key"]; has {
		t.Fatal("payload must not contain prv_key")
	}
	if !config.IsEquivalentToDesiredState(payload) {
		t.Error("expected true when app_id and base_url match after prv_key is omitted from comparison")
	}
}

func TestGitHubSecretEngineConfigIsEquivalentMatching(t *testing.T) {
	config := &GitHubSecretEngineConfig{
		Spec: GitHubSecretEngineConfigSpec{
			Path: "github",
			GHConfig: GHConfig{
				ApplicationID:    12345,
				GitHubAPIBaseURL: "https://api.github.com",
				retrievedSSHKey:  "ignored-for-equivalence",
			},
		},
	}

	desiredState := config.Spec.GHConfig.toMap()
	delete(desiredState, "prv_key")

	if !config.IsEquivalentToDesiredState(desiredState) {
		t.Error("expected true for payload matching desired state without prv_key")
	}
}

func TestGitHubSecretEngineConfigIsEquivalentNonMatching(t *testing.T) {
	config := &GitHubSecretEngineConfig{
		Spec: GitHubSecretEngineConfigSpec{
			Path: "github",
			GHConfig: GHConfig{
				ApplicationID:    12345,
				GitHubAPIBaseURL: "https://api.github.com",
				retrievedSSHKey:  "key",
			},
		},
	}

	payload := map[string]interface{}{
		"app_id":   int64(12345),
		"base_url": "https://api.github.example.com",
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when base_url differs")
	}
}

func TestGitHubSecretEngineConfigIsEquivalentExtraFields(t *testing.T) {
	config := &GitHubSecretEngineConfig{
		Spec: GitHubSecretEngineConfigSpec{
			Path: "github",
			GHConfig: GHConfig{
				ApplicationID:    12345,
				GitHubAPIBaseURL: "https://api.github.com",
				retrievedSSHKey:  "key",
			},
		},
	}

	payload := map[string]interface{}{
		"app_id":      int64(12345),
		"base_url":    "https://api.github.com",
		"extra_field": "from-vault",
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when payload has extra keys (bare reflect.DeepEqual after prv_key deletion)")
	}
}

func TestGitHubSecretEngineConfigIsEquivalentPayloadWithPrvKey(t *testing.T) {
	config := &GitHubSecretEngineConfig{
		Spec: GitHubSecretEngineConfigSpec{
			Path: "github",
			GHConfig: GHConfig{
				ApplicationID:    12345,
				GitHubAPIBaseURL: "https://api.github.com",
				retrievedSSHKey:  "test-ssh-key",
			},
		},
	}

	payload := map[string]interface{}{
		"app_id":   int64(12345),
		"base_url": "https://api.github.com",
		"prv_key":  "test-ssh-key",
	}

	if config.IsEquivalentToDesiredState(payload) {
		t.Error("expected false: desired state deletes prv_key, so payload still containing prv_key has an extra key")
	}
}

func TestGitHubSecretEngineConfigIsDeletable(t *testing.T) {
	config := &GitHubSecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected GitHubSecretEngineConfig not to be deletable")
	}
}

func TestGitHubSecretEngineConfigConditions(t *testing.T) {
	config := &GitHubSecretEngineConfig{}

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
