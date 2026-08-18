package v1alpha1

import (
	"testing"
)

func TestGitHubAuthEngineTeamMap_toMap(t *testing.T) {
	instance := &GitHubAuthEngineTeamMap{}
	instance.Spec.Policies = "policy1,policy2"

	result := instance.toMap()

	if result["value"] != "policy1,policy2" {
		t.Errorf("expected value=policy1,policy2, got %v", result["value"])
	}
	if len(result) != 1 {
		t.Errorf("expected exactly 1 key in map, got %d keys: %v", len(result), result)
	}
}

func TestGitHubAuthEngineTeamMap_toMap_EmptyPolicies(t *testing.T) {
	instance := &GitHubAuthEngineTeamMap{}
	instance.Spec.Policies = ""

	result := instance.toMap()

	if result["value"] != "" {
		t.Errorf("expected value to be empty string, got %v", result["value"])
	}
}

func TestGitHubAuthEngineTeamMap_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &GitHubAuthEngineTeamMap{}
	instance.Spec.Policies = "dev-policy"

	vaultPayload := map[string]any{
		"key":   "dev",
		"value": "dev-policy",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true (key field should be filtered out)")
	}
}

func TestGitHubAuthEngineTeamMap_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &GitHubAuthEngineTeamMap{}
	instance.Spec.Policies = "dev-policy"

	vaultPayload := map[string]any{
		"key":   "dev",
		"value": "different-policy",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched value")
	}
}

func TestGitHubAuthEngineTeamMap_GetPath(t *testing.T) {
	instance := &GitHubAuthEngineTeamMap{}
	instance.Spec.Path = "github"
	instance.Spec.Name = "dev-team"

	expected := "auth/github/map/teams/dev-team"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestGitHubAuthEngineTeamMap_IsDeletable(t *testing.T) {
	instance := &GitHubAuthEngineTeamMap{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true for team map type")
	}
}
