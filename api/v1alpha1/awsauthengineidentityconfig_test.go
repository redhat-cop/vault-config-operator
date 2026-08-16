package v1alpha1

import (
	"testing"
)

func TestAWSAuthEngineIdentityConfig_toMap(t *testing.T) {
	config := &AWSAuthIdentityConfig{
		IAMAlias:    "full_arn",
		IAMMetadata: "default",
		EC2Alias:    "instance_id",
		EC2Metadata: "default",
	}

	result := config.toMap()

	if result["iam_alias"] != "full_arn" {
		t.Errorf("expected iam_alias=full_arn, got %v", result["iam_alias"])
	}
	if result["iam_metadata"] != "default" {
		t.Errorf("expected iam_metadata=default, got %v", result["iam_metadata"])
	}
	if result["ec2_alias"] != "instance_id" {
		t.Errorf("expected ec2_alias=instance_id, got %v", result["ec2_alias"])
	}
	if result["ec2_metadata"] != "default" {
		t.Errorf("expected ec2_metadata=default, got %v", result["ec2_metadata"])
	}
}

func TestAWSAuthEngineIdentityConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &AWSAuthEngineIdentityConfig{}
	instance.Spec.AWSAuthIdentityConfig = AWSAuthIdentityConfig{
		IAMAlias:    "full_arn",
		IAMMetadata: "default",
		EC2Alias:    "role_id",
		EC2Metadata: "default",
	}

	vaultPayload := map[string]any{
		"iam_alias":    "full_arn",
		"iam_metadata": "default",
		"ec2_alias":    "role_id",
		"ec2_metadata": "default",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching payload")
	}
}

func TestAWSAuthEngineIdentityConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &AWSAuthEngineIdentityConfig{}
	instance.Spec.AWSAuthIdentityConfig = AWSAuthIdentityConfig{
		IAMAlias:    "full_arn",
		IAMMetadata: "default",
		EC2Alias:    "role_id",
		EC2Metadata: "default",
	}

	vaultPayload := map[string]any{
		"iam_alias":    "unique_id",
		"iam_metadata": "default",
		"ec2_alias":    "role_id",
		"ec2_metadata": "default",
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched iam_alias")
	}
}

func TestAWSAuthEngineIdentityConfig_GetPath(t *testing.T) {
	instance := &AWSAuthEngineIdentityConfig{}
	instance.Spec.Path = "aws"

	expected := "auth/aws/config/identity"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestAWSAuthEngineIdentityConfig_IsDeletable(t *testing.T) {
	instance := &AWSAuthEngineIdentityConfig{}
	if instance.IsDeletable() {
		t.Error("expected IsDeletable() to return false")
	}
}
