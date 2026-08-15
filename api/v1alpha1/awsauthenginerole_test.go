package v1alpha1

import (
	"reflect"
	"testing"
)

func TestAWSAuthEngineRole_toMap_IAM(t *testing.T) {
	role := &AWSAuthRole{
		Name:                 "test-iam-role",
		AuthType:             "iam",
		BoundIAMPrincipalARN: []string{"arn:aws:iam::123456789012:role/MyRole"},
		ResolveAWSUniqueIDs:  true,
		TokenTTL:             "1h",
		TokenMaxTTL:          "24h",
		TokenPolicies:        []string{"dev", "prod"},
		TokenType:            "service",
		TokenNumUses:         5,
	}

	result := role.toMap()

	if result["auth_type"] != "iam" {
		t.Errorf("expected auth_type=iam, got %v", result["auth_type"])
	}
	expectedPrincipalARN := []any{"arn:aws:iam::123456789012:role/MyRole"}
	if !reflect.DeepEqual(result["bound_iam_principal_arn"], expectedPrincipalARN) {
		t.Errorf("expected bound_iam_principal_arn=%v, got %v", expectedPrincipalARN, result["bound_iam_principal_arn"])
	}
	if result["resolve_aws_unique_ids"] != true {
		t.Errorf("expected resolve_aws_unique_ids=true, got %v", result["resolve_aws_unique_ids"])
	}
	if result["token_ttl"] != "1h" {
		t.Errorf("expected token_ttl=1h, got %v", result["token_ttl"])
	}
	if result["token_max_ttl"] != "24h" {
		t.Errorf("expected token_max_ttl=24h, got %v", result["token_max_ttl"])
	}
	expectedPolicies := []any{"dev", "prod"}
	if !reflect.DeepEqual(result["token_policies"], expectedPolicies) {
		t.Errorf("expected token_policies=%v, got %v", expectedPolicies, result["token_policies"])
	}
	if result["token_type"] != "service" {
		t.Errorf("expected token_type=service, got %v", result["token_type"])
	}
	if result["token_num_uses"] != int64(5) {
		t.Errorf("expected token_num_uses=5, got %v", result["token_num_uses"])
	}
	emptySlice := []any{}
	if !reflect.DeepEqual(result["bound_ami_id"], emptySlice) {
		t.Errorf("expected bound_ami_id to be empty slice, got %v", result["bound_ami_id"])
	}
}

func TestAWSAuthEngineRole_toMap_EC2(t *testing.T) {
	role := &AWSAuthRole{
		Name:                     "test-ec2-role",
		AuthType:                 "ec2",
		BoundAmiID:               []string{"ami-fce36987"},
		BoundAccountID:           []string{"123456789012"},
		BoundRegion:              []string{"us-east-1", "us-west-2"},
		RoleTag:                  "VaultRole",
		AllowInstanceMigration:   false,
		DisallowReauthentication: true,
		TokenPolicies:            []string{"default", "dev"},
	}

	result := role.toMap()

	if result["auth_type"] != "ec2" {
		t.Errorf("expected auth_type=ec2, got %v", result["auth_type"])
	}
	expectedAmiID := []any{"ami-fce36987"}
	if !reflect.DeepEqual(result["bound_ami_id"], expectedAmiID) {
		t.Errorf("expected bound_ami_id=%v, got %v", expectedAmiID, result["bound_ami_id"])
	}
	expectedAccountID := []any{"123456789012"}
	if !reflect.DeepEqual(result["bound_account_id"], expectedAccountID) {
		t.Errorf("expected bound_account_id=%v, got %v", expectedAccountID, result["bound_account_id"])
	}
	expectedRegion := []any{"us-east-1", "us-west-2"}
	if !reflect.DeepEqual(result["bound_region"], expectedRegion) {
		t.Errorf("expected bound_region=%v, got %v", expectedRegion, result["bound_region"])
	}
	if result["role_tag"] != "VaultRole" {
		t.Errorf("expected role_tag=VaultRole, got %v", result["role_tag"])
	}
	if result["allow_instance_migration"] != false {
		t.Errorf("expected allow_instance_migration=false, got %v", result["allow_instance_migration"])
	}
	if result["disallow_reauthentication"] != true {
		t.Errorf("expected disallow_reauthentication=true, got %v", result["disallow_reauthentication"])
	}
}

func TestAWSAuthEngineRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	instance := &AWSAuthEngineRole{}
	instance.Spec.AWSAuthRole = AWSAuthRole{
		Name:                "my-role",
		AuthType:            "ec2",
		BoundAmiID:          []string{"ami-fce36987"},
		ResolveAWSUniqueIDs: true,
		TokenPolicies:       []string{"default", "dev"},
	}

	vaultPayload := map[string]any{
		"auth_type":                      "ec2",
		"bound_ami_id":                   []any{"ami-fce36987"},
		"bound_account_id":               []any{},
		"bound_region":                   []any{},
		"bound_vpc_id":                   []any{},
		"bound_subnet_id":                []any{},
		"bound_iam_role_arn":             []any{},
		"bound_iam_instance_profile_arn": []any{},
		"bound_ec2_instance_id":          []any{},
		"role_tag":                       "",
		"bound_iam_principal_arn":        []any{},
		"inferred_entity_type":           "",
		"inferred_aws_region":            "",
		"resolve_aws_unique_ids":         true,
		"allow_instance_migration":       false,
		"disallow_reauthentication":      false,
		"token_ttl":                      "",
		"token_max_ttl":                  "",
		"token_policies":                 []any{"default", "dev"},
		"policies":                       []any{},
		"token_bound_cidrs":              []any{},
		"token_explicit_max_ttl":         "",
		"token_no_default_policy":        false,
		"token_num_uses":                 int64(0),
		"token_period":                   int64(0),
		"token_type":                     "",
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true for matching payload")
	}
}

func TestAWSAuthEngineRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	instance := &AWSAuthEngineRole{}
	instance.Spec.AWSAuthRole = AWSAuthRole{
		Name:                "my-role",
		AuthType:            "iam",
		ResolveAWSUniqueIDs: true,
	}

	vaultPayload := map[string]any{
		"auth_type":              "ec2",
		"resolve_aws_unique_ids": true,
	}

	if instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return false for mismatched auth_type")
	}
}

func TestAWSAuthEngineRole_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	instance := &AWSAuthEngineRole{}
	instance.Spec.AWSAuthRole = AWSAuthRole{
		Name:                 "my-role",
		AuthType:             "iam",
		BoundIAMPrincipalARN: []string{"arn:aws:iam::123456789012:role/MyRole"},
		ResolveAWSUniqueIDs:  true,
	}

	vaultPayload := map[string]any{
		"auth_type":                 "iam",
		"bound_iam_principal_arn":   []any{"arn:aws:iam::123456789012:role/MyRole"},
		"resolve_aws_unique_ids":    true,
		"allow_instance_migration":  false,
		"disallow_reauthentication": false,
		"token_no_default_policy":   false,
		"token_num_uses":            int64(0),
		"token_period":              int64(0),
		"request_id":                "extra-field",
		"lease_id":                  "",
		"renewable":                 false,
	}

	if !instance.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected IsEquivalentToDesiredState to return true when Vault returns extra fields")
	}
}

func TestAWSAuthEngineRole_Webhook_IAMOnlyFields_EC2Rejected(t *testing.T) {
	role := &AWSAuthRole{
		Name:                 "test-role",
		AuthType:             "ec2",
		BoundIAMPrincipalARN: []string{"arn:aws:iam::123456789012:role/MyRole"},
	}

	err := validateAWSAuthRoleSpec(role)
	if err == nil {
		t.Error("expected validation error for IAM-only field on EC2 role")
	}

	role2 := &AWSAuthRole{
		Name:               "test-role",
		AuthType:           "ec2",
		InferredEntityType: "ec2_instance",
	}

	err = validateAWSAuthRoleSpec(role2)
	if err == nil {
		t.Error("expected validation error for inferredEntityType on EC2 role")
	}

	role3 := &AWSAuthRole{
		Name:              "test-role",
		AuthType:          "ec2",
		InferredAWSRegion: "us-east-1",
	}

	err = validateAWSAuthRoleSpec(role3)
	if err == nil {
		t.Error("expected validation error for inferredAWSRegion on EC2 role")
	}
}

func TestAWSAuthEngineRole_Webhook_EC2OnlyFields_IAMRejected(t *testing.T) {
	role := &AWSAuthRole{
		Name:     "test-role",
		AuthType: "iam",
		RoleTag:  "VaultRole",
	}

	err := validateAWSAuthRoleSpec(role)
	if err == nil {
		t.Error("expected validation error for EC2-only field roleTag on IAM role")
	}

	role2 := &AWSAuthRole{
		Name:                   "test-role",
		AuthType:               "iam",
		AllowInstanceMigration: true,
	}

	err = validateAWSAuthRoleSpec(role2)
	if err == nil {
		t.Error("expected validation error for EC2-only field allowInstanceMigration on IAM role")
	}

	role3 := &AWSAuthRole{
		Name:                     "test-role",
		AuthType:                 "iam",
		DisallowReauthentication: true,
	}

	err = validateAWSAuthRoleSpec(role3)
	if err == nil {
		t.Error("expected validation error for EC2-only field disallowReauthentication on IAM role")
	}
}

func TestAWSAuthEngineRole_Webhook_MutuallyExclusive(t *testing.T) {
	role := &AWSAuthRole{
		Name:                     "test-role",
		AuthType:                 "ec2",
		AllowInstanceMigration:   true,
		DisallowReauthentication: true,
	}

	err := validateAWSAuthRoleSpec(role)
	if err == nil {
		t.Error("expected validation error for mutually exclusive allowInstanceMigration and disallowReauthentication")
	}
}

func TestAWSAuthEngineRole_Webhook_ValidIAMRole(t *testing.T) {
	role := &AWSAuthRole{
		Name:                 "test-role",
		AuthType:             "iam",
		BoundIAMPrincipalARN: []string{"arn:aws:iam::123456789012:role/MyRole"},
		ResolveAWSUniqueIDs:  true,
		TokenPolicies:        []string{"dev"},
	}

	err := validateAWSAuthRoleSpec(role)
	if err != nil {
		t.Errorf("expected no validation error for valid IAM role, got: %v", err)
	}
}

func TestAWSAuthEngineRole_Webhook_ValidEC2Role(t *testing.T) {
	role := &AWSAuthRole{
		Name:           "test-role",
		AuthType:       "ec2",
		BoundAmiID:     []string{"ami-12345"},
		BoundAccountID: []string{"123456789012"},
		RoleTag:        "VaultRole",
		TokenPolicies:  []string{"default"},
	}

	err := validateAWSAuthRoleSpec(role)
	if err != nil {
		t.Errorf("expected no validation error for valid EC2 role, got: %v", err)
	}
}

func TestAWSAuthEngineRole_GetPath(t *testing.T) {
	instance := &AWSAuthEngineRole{}
	instance.Spec.Path = "aws"
	instance.Spec.Name = "my-role"

	expected := "auth/aws/role/my-role"
	if instance.GetPath() != expected {
		t.Errorf("expected path=%s, got %s", expected, instance.GetPath())
	}
}

func TestAWSAuthEngineRole_IsDeletable(t *testing.T) {
	instance := &AWSAuthEngineRole{}
	if !instance.IsDeletable() {
		t.Error("expected IsDeletable() to return true")
	}
}
