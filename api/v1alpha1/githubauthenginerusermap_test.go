package v1alpha1

import (
	"testing"
)

func TestGitHubAuthEngineUserMap_toMap(t *testing.T) {
	instance := &GitHubAuthEngineUserMap{}
	instance.Spec.Policies = "sethvargo-policy"

	result := instance.toMap()

	if result["value"] != "sethvargo-policy" {
		t.Errorf("expected value=sethvargo-policy, got %v", result["value"])
	}
	if len(result) != 1 {
		t.Errorf("expected exactly 1 key in map, got %d keys: %v", len(result), result)
	}
}

func TestGitHubAuthEngineUserMap_toMap_EmptyPolicies(t *testing.T) {
	instance := &GitHubAuthEngineUserMap{}
	instance.Spec.Policies = ""

	result := instance.toMap()

	if result["value"] != "" {
		t.Errorf("expected value to be empty string, got %v", result["value"])
	}
}

func TestGitHubAuthEngineUserMap_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &GitHubAuthEngineUserMap{}
	instance.Spec.Policies = "sethvargo-policy"

	vaultPayload := map[string]any{
		"key":   "sethvargo",
		"value": "sethvargo-policy",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true (key field should be filtered out)")
	}
}

func TestGitHubAuthEngineUserMap_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &GitHubAuthEngineUserMap{}
	instance.Spec.Policies = "sethvargo-policy"

	vaultPayload := map[string]any{
		"key":   "sethvargo",
		"value": "different-policy",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched value")
	}
}

func TestGitHubAuthEngineUserMap_GetPath(t *testing.T) {
	instance := &GitHubAuthEngineUserMap{}
	instance.Spec.Path = "github"
	instance.Spec.Name = "sethvargo"

	expected := "auth/github/map/users/sethvargo"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestGitHubAuthEngineUserMap_IsDeletable(t *testing.T) {
	instance := &GitHubAuthEngineUserMap{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true for user map type")
	}
}
