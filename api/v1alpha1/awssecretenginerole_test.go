package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAWSSecretEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *AWSSecretEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &AWSSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AWSSecretEngineRoleSpec{
					Path: "aws",
					Name: "custom-name",
				},
			},
			expectedPath: vaultutils.CleansePath("aws/roles/custom-name"),
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &AWSSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AWSSecretEngineRoleSpec{
					Path: "aws",
				},
			},
			expectedPath: vaultutils.CleansePath("aws/roles/meta-name"),
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

func TestAWSSERoleToMap(t *testing.T) {
	role := AWSSERole{
		CredentialType:         "iam_user",
		PolicyARNs:             []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"},
		PolicyDocument:         `{"Version": "2012-10-17"}`,
		IAMGroups:              []string{"group1", "group2"},
		IAMTags:                []string{"env=prod", "team=platform"},
		UserPath:               "/custom/path/",
		PermissionsBoundaryARN: "arn:aws:iam::123456789012:policy/boundary",
		MFASerialNumber:        "arn:aws:iam::123456789012:mfa/user",
	}

	result := role.toMap()

	if result["credential_type"] != "iam_user" {
		t.Errorf("credential_type = %v, expected 'iam_user'", result["credential_type"])
	}

	policyArns, ok := result["policy_arns"].([]string)
	if !ok {
		t.Fatalf("policy_arns should be []string, got %T", result["policy_arns"])
	}
	if len(policyArns) != 1 || policyArns[0] != "arn:aws:iam::aws:policy/ReadOnlyAccess" {
		t.Errorf("policy_arns = %v", policyArns)
	}

	if result["policy_document"] != `{"Version": "2012-10-17"}` {
		t.Errorf("policy_document = %v", result["policy_document"])
	}

	iamGroups, ok := result["iam_groups"].([]string)
	if !ok {
		t.Fatalf("iam_groups should be []string, got %T", result["iam_groups"])
	}
	if len(iamGroups) != 2 {
		t.Errorf("expected 2 iam_groups, got %d", len(iamGroups))
	}

	iamTags, ok := result["iam_tags"].([]string)
	if !ok {
		t.Fatalf("iam_tags should be []string, got %T", result["iam_tags"])
	}
	if len(iamTags) != 2 {
		t.Errorf("expected 2 iam_tags, got %d", len(iamTags))
	}

	if result["user_path"] != "/custom/path/" {
		t.Errorf("user_path = %v", result["user_path"])
	}

	if result["permissions_boundary_arn"] != "arn:aws:iam::123456789012:policy/boundary" {
		t.Errorf("permissions_boundary_arn = %v", result["permissions_boundary_arn"])
	}

	if result["mfa_serial_number"] != "arn:aws:iam::123456789012:mfa/user" {
		t.Errorf("mfa_serial_number = %v", result["mfa_serial_number"])
	}

	// Fields that should not be in the map when not set
	if _, exists := result["role_arns"]; exists {
		t.Error("role_arns should not be in map when empty")
	}
	if _, exists := result["default_sts_ttl"]; exists {
		t.Error("default_sts_ttl should not be in map when empty")
	}
	if _, exists := result["max_sts_ttl"]; exists {
		t.Error("max_sts_ttl should not be in map when empty")
	}
	if _, exists := result["session_tags"]; exists {
		t.Error("session_tags should not be in map when empty")
	}
	if _, exists := result["external_id"]; exists {
		t.Error("external_id should not be in map when empty")
	}
}

func TestAWSSERoleToMapAssumedRole(t *testing.T) {
	role := AWSSERole{
		CredentialType: "assumed_role",
		RoleARNs:       []string{"arn:aws:iam::123456789012:role/MyRole"},
		PolicyARNs:     []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"},
		SessionTags:    []string{"project=foo", "dept=engineering"},
		DefaultSTSTTL:  "1h",
		MaxSTSTTL:      "12h",
		ExternalID:     "external-id-123",
	}

	result := role.toMap()

	if result["credential_type"] != "assumed_role" {
		t.Errorf("credential_type = %v", result["credential_type"])
	}

	roleArns, ok := result["role_arns"].([]string)
	if !ok {
		t.Fatalf("role_arns should be []string, got %T", result["role_arns"])
	}
	if len(roleArns) != 1 || roleArns[0] != "arn:aws:iam::123456789012:role/MyRole" {
		t.Errorf("role_arns = %v", roleArns)
	}

	sessionTags, ok := result["session_tags"].([]string)
	if !ok {
		t.Fatalf("session_tags should be []string, got %T", result["session_tags"])
	}
	if len(sessionTags) != 2 {
		t.Errorf("expected 2 session_tags, got %d", len(sessionTags))
	}

	if result["default_sts_ttl"] != "1h" {
		t.Errorf("default_sts_ttl = %v", result["default_sts_ttl"])
	}

	if result["max_sts_ttl"] != "12h" {
		t.Errorf("max_sts_ttl = %v", result["max_sts_ttl"])
	}

	if result["external_id"] != "external-id-123" {
		t.Errorf("external_id = %v", result["external_id"])
	}

	// Fields that should not be in the map for assumed_role
	if _, exists := result["user_path"]; exists {
		t.Error("user_path should not be in map for assumed_role")
	}
	if _, exists := result["permissions_boundary_arn"]; exists {
		t.Error("permissions_boundary_arn should not be in map for assumed_role")
	}
	if _, exists := result["iam_tags"]; exists {
		t.Error("iam_tags should not be in map for assumed_role")
	}
	if _, exists := result["mfa_serial_number"]; exists {
		t.Error("mfa_serial_number should not be in map for assumed_role")
	}
}

func TestAWSSecretEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSSERole: AWSSERole{
				CredentialType: "iam_user",
				PolicyARNs:     []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"},
				PolicyDocument: `{"Version": "2012-10-17"}`,
				IAMGroups:      []string{"group1"},
				UserPath:       "/users/",
			},
		},
	}

	payload := role.Spec.AWSSERole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAWSSecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSSERole: AWSSERole{
				CredentialType: "assumed_role",
				RoleARNs:       []string{"arn:aws:iam::123456789012:role/MyRole"},
				DefaultSTSTTL:  "1h",
			},
		},
	}

	payload := role.Spec.AWSSERole.toMap()
	payload["default_sts_ttl"] = "2h"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different default_sts_ttl) to NOT be equivalent")
	}
}

func TestAWSSecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSSERole: AWSSERole{
				CredentialType: "iam_user",
				PolicyDocument: `{"Version": "2012-10-17"}`,
			},
		},
	}

	payload := role.Spec.AWSSERole.toMap()
	payload["extra_vault_field"] = "some-value"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent (bare reflect.DeepEqual)")
	}
}

func TestAWSSecretEngineRoleIsDeletable(t *testing.T) {
	role := &AWSSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected AWSSecretEngineRole to be deletable")
	}
}

func TestAWSSecretEngineRoleConditions(t *testing.T) {
	role := &AWSSecretEngineRole{}

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
