package v1alpha1

import (
	"testing"

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
			expectedPath: "aws/roles/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &AWSSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: AWSSecretEngineRoleSpec{
					Path: "aws",
				},
			},
			expectedPath: "aws/roles/meta-name",
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

func TestAWSSecretEngineRoleIsDeletable(t *testing.T) {
	role := &AWSSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected AWSSecretEngineRole to be deletable")
	}
}

func TestAWSRole_toMap_IAMUser(t *testing.T) {
	role := AWSRole{
		CredentialType:         "iam_user",
		PolicyArns:             []string{"arn:aws:iam::123456789012:policy/MyPolicy"},
		IAMGroups:              []string{"developers"},
		IAMTags:                []string{"env=prod"},
		UserPath:               "/engineering/",
		PermissionsBoundaryARN: "arn:aws:iam::123456789012:policy/Boundary",
	}

	result := role.toMap()

	if ct, ok := result["credential_type"].(string); !ok || ct != "iam_user" {
		t.Errorf("credential_type = %v (%T), expected string iam_user", result["credential_type"], result["credential_type"])
	}
	if pa, ok := result["policy_arns"].([]any); !ok || len(pa) != 1 || pa[0] != "arn:aws:iam::123456789012:policy/MyPolicy" {
		t.Errorf("policy_arns = %v, unexpected", result["policy_arns"])
	}
	if ig, ok := result["iam_groups"].([]any); !ok || len(ig) != 1 || ig[0] != "developers" {
		t.Errorf("iam_groups = %v, unexpected", result["iam_groups"])
	}
	if it, ok := result["iam_tags"].([]any); !ok || len(it) != 1 || it[0] != "env=prod" {
		t.Errorf("iam_tags = %v, unexpected", result["iam_tags"])
	}
	if up, ok := result["user_path"].(string); !ok || up != "/engineering/" {
		t.Errorf("user_path = %v, expected /engineering/", result["user_path"])
	}
	if pb, ok := result["permissions_boundary_arn"].(string); !ok || pb != "arn:aws:iam::123456789012:policy/Boundary" {
		t.Errorf("permissions_boundary_arn = %v, unexpected", result["permissions_boundary_arn"])
	}
	if ra, ok := result["role_arns"].([]any); !ok || len(ra) != 0 {
		t.Errorf("role_arns should be empty for iam_user when not set, got %v", result["role_arns"])
	}
}

func TestAWSRole_toMap_AssumedRole(t *testing.T) {
	role := AWSRole{
		CredentialType: "assumed_role",
		RoleArns:       []string{"arn:aws:iam::123456789012:role/DeveloperRole"},
		PolicyArns:     []string{"arn:aws:iam::123456789012:policy/MyPolicy"},
		DefaultSTSTTL:  "1h",
		MaxSTSTTL:      "4h",
		ExternalID:     "external-123",
		SessionTags:    []string{"project=myproject"},
	}

	result := role.toMap()

	if ct, ok := result["credential_type"].(string); !ok || ct != "assumed_role" {
		t.Errorf("credential_type = %v, expected assumed_role", result["credential_type"])
	}
	if ra, ok := result["role_arns"].([]any); !ok || len(ra) != 1 || ra[0] != "arn:aws:iam::123456789012:role/DeveloperRole" {
		t.Errorf("role_arns = %v, unexpected", result["role_arns"])
	}
	if dttl, ok := result["default_sts_ttl"].(string); !ok || dttl != "1h" {
		t.Errorf("default_sts_ttl = %v, expected 1h", result["default_sts_ttl"])
	}
	if mttl, ok := result["max_sts_ttl"].(string); !ok || mttl != "4h" {
		t.Errorf("max_sts_ttl = %v, expected 4h", result["max_sts_ttl"])
	}
	if eid, ok := result["external_id"].(string); !ok || eid != "external-123" {
		t.Errorf("external_id = %v, expected external-123", result["external_id"])
	}
	if st, ok := result["session_tags"].([]any); !ok || len(st) != 1 || st[0] != "project=myproject" {
		t.Errorf("session_tags = %v, unexpected", result["session_tags"])
	}
}

func TestAWSRole_toMap_FederationToken(t *testing.T) {
	role := AWSRole{
		CredentialType: "federation_token",
		PolicyArns:     []string{"arn:aws:iam::123456789012:policy/FedPolicy"},
		PolicyDocument: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:*","Resource":"*"}]}`,
		DefaultSTSTTL:  "30m",
		MaxSTSTTL:      "2h",
	}

	result := role.toMap()

	if ct, ok := result["credential_type"].(string); !ok || ct != "federation_token" {
		t.Errorf("credential_type = %v, expected federation_token", result["credential_type"])
	}
	if pd, ok := result["policy_document"].(string); !ok || pd == "" {
		t.Errorf("policy_document = %v, expected non-empty JSON", result["policy_document"])
	}
	if ra, ok := result["role_arns"].([]any); !ok || len(ra) != 0 {
		t.Errorf("role_arns should be empty for federation_token when not set, got %v", result["role_arns"])
	}
}

func TestAWSRole_toMap_SessionToken(t *testing.T) {
	role := AWSRole{
		CredentialType: "session_token",
	}

	result := role.toMap()

	if ct, ok := result["credential_type"].(string); !ok || ct != "session_token" {
		t.Errorf("credential_type = %v, expected session_token", result["credential_type"])
	}
	if _, ok := result["credential_type"]; !ok {
		t.Error("expected credential_type key")
	}
}

func TestAWSRole_toMap_MFASerialNumber(t *testing.T) {
	role := AWSRole{
		CredentialType:  "assumed_role",
		RoleArns:        []string{"arn:aws:iam::123456789012:role/MFARole"},
		MFASerialNumber: "arn:aws:iam::123456789012:mfa/my-device",
	}

	result := role.toMap()

	if msn, ok := result["mfa_serial_number"].(string); !ok || msn != "arn:aws:iam::123456789012:mfa/my-device" {
		t.Errorf("mfa_serial_number = %v, unexpected", result["mfa_serial_number"])
	}
}

func TestAWSRole_toMap_AllManagedFieldsPresent(t *testing.T) {
	role := AWSRole{
		CredentialType: "iam_user",
	}

	result := role.toMap()

	managedKeys := []string{
		"credential_type", "role_arns", "policy_arns", "policy_document",
		"iam_groups", "iam_tags", "default_sts_ttl", "max_sts_ttl",
		"user_path", "permissions_boundary_arn", "external_id",
		"session_tags", "mfa_serial_number",
	}
	for _, key := range managedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected managed key %q in toMap() output even when zero-valued", key)
		}
	}
}

func TestAWSSecretEngineRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSRole: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/DeveloperRole"},
				PolicyArns:     []string{"arn:aws:iam::123456789012:policy/MyPolicy"},
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_types": []any{"assumed_role"},
		"role_arns":        []any{"arn:aws:iam::123456789012:role/DeveloperRole"},
		"policy_arns":      []any{"arn:aws:iam::123456789012:policy/MyPolicy"},
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected matching payload with credential_types (plural array) to be equivalent")
	}
}

func TestAWSSecretEngineRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSRole: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/DeveloperRole"},
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_types": []any{"assumed_role"},
		"role_arns":        []any{"arn:aws:iam::123456789012:role/OtherRole"},
	}

	if role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when role_arns differ")
	}
}

func TestAWSSecretEngineRole_IsEquivalentToDesiredState_CredentialTypesMapping(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSRole: AWSRole{
				CredentialType: "iam_user",
				PolicyArns:     []string{"arn:aws:iam::123456789012:policy/AdminPolicy"},
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_types": []any{"iam_user"},
		"policy_arns":      []any{"arn:aws:iam::123456789012:policy/AdminPolicy"},
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: credential_type (singular) should map to credential_types (plural array)")
	}

	wrongTypePayload := map[string]any{
		"credential_types": []any{"assumed_role"},
		"policy_arns":      []any{"arn:aws:iam::123456789012:policy/AdminPolicy"},
	}

	if role.IsEquivalentToDesiredState(wrongTypePayload) {
		t.Error("expected false: credential_types value differs (iam_user vs assumed_role)")
	}
}

func TestAWSSecretEngineRole_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSRole: AWSRole{
				CredentialType: "federation_token",
				PolicyArns:     []string{"arn:aws:iam::123456789012:policy/FedPolicy"},
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_types": []any{"federation_token"},
		"policy_arns":      []any{"arn:aws:iam::123456789012:policy/FedPolicy"},
		"extra_vault_key":  "should-be-ignored",
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: extra fields from Vault read should be ignored")
	}
}

func TestAWSSecretEngineRole_IsEquivalentToDesiredState_ClearedFieldDetected(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSRole: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/Role"},
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_types": []any{"assumed_role"},
		"role_arns":        []any{"arn:aws:iam::123456789012:role/Role"},
		"policy_arns":      []any{"arn:aws:iam::123456789012:policy/OldPolicy"},
		"external_id":      "old-external-id",
		"session_tags":     []any{"old=tag"},
	}

	if role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: policyArns, externalID, sessionTags were cleared but Vault still has them")
	}
}

func TestAWSSecretEngineRole_IsEquivalentToDesiredState_UnsetFieldsIgnored(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSRole: AWSRole{
				CredentialType: "iam_user",
				PolicyArns:     []string{"arn:aws:iam::123456789012:policy/MyPolicy"},
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_types": []any{"iam_user"},
		"policy_arns":      []any{"arn:aws:iam::123456789012:policy/MyPolicy"},
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: fields never set by user and absent from Vault should not cause false drift")
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

func TestAWSSecretEngineRole_IsValid_AssumedRoleRequiresRoleArns(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "assumed_role",
			},
		},
	}
	ok, err := role.IsValid()
	if ok || err == nil {
		t.Error("expected invalid: assumed_role requires roleArns")
	}
}

func TestAWSSecretEngineRole_IsValid_AssumedRoleWithRoleArns(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/Role"},
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: assumed_role with roleArns, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_SessionTokenRejectsPolicyArns(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "session_token",
				PolicyArns:     []string{"arn:aws:iam::123456789012:policy/Pol"},
			},
		},
	}
	ok, err := role.IsValid()
	if ok || err == nil {
		t.Error("expected invalid: session_token prohibits policyArns")
	}
}

func TestAWSSecretEngineRole_IsValid_SessionTokenRejectsPolicyDocument(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "session_token",
				PolicyDocument: `{"Statement":[]}`,
			},
		},
	}
	ok, err := role.IsValid()
	if ok || err == nil {
		t.Error("expected invalid: session_token prohibits policyDocument")
	}
}

func TestAWSSecretEngineRole_IsValid_SessionTokenValid(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "session_token",
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: session_token with no prohibited fields, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_IAMUserValid(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "iam_user",
				PolicyArns:     []string{"arn:aws:iam::123456789012:policy/Pol"},
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: iam_user with policyArns, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_FederationTokenValid(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "federation_token",
				PolicyArns:     []string{"arn:aws:iam::123456789012:policy/Pol"},
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: federation_token with policyArns, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_IAMUserRejectsRoleArns(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "iam_user",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/SomeRole"},
			},
		},
	}
	ok, err := role.IsValid()
	if ok || err == nil {
		t.Error("expected invalid: iam_user should not allow roleArns")
	}
}

func TestAWSSecretEngineRole_IsValid_FederationTokenRejectsRoleArns(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "federation_token",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/SomeRole"},
			},
		},
	}
	ok, err := role.IsValid()
	if ok || err == nil {
		t.Error("expected invalid: federation_token should not allow roleArns")
	}
}

func TestAWSSecretEngineRole_IsValid_SessionTokenRejectsRoleArns(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "session_token",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/SomeRole"},
			},
		},
	}
	ok, err := role.IsValid()
	if ok || err == nil {
		t.Error("expected invalid: session_token should not allow roleArns")
	}
}

func TestAWSSecretEngineRole_IsValid_AssumedRoleRejectsIAMUserFields(t *testing.T) {
	tests := []struct {
		name string
		role AWSRole
	}{
		{
			name: "rejects iamTags",
			role: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/Role"},
				IAMTags:        []string{"env=prod"},
			},
		},
		{
			name: "rejects userPath",
			role: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/Role"},
				UserPath:       "/engineering/",
			},
		},
		{
			name: "rejects permissionsBoundaryARN",
			role: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/Role"},
				PermissionsBoundaryARN: "arn:aws:iam::123456789012:policy/Boundary",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &AWSSecretEngineRole{Spec: AWSSecretEngineRoleSpec{AWSRole: tt.role}}
			ok, err := role.IsValid()
			if ok || err == nil {
				t.Errorf("expected invalid: assumed_role %s", tt.name)
			}
		})
	}
}

func TestAWSSecretEngineRole_IsValid_AssumedRoleAllowsOwnFields(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/Role"},
				ExternalID:     "ext-123",
				SessionTags:    []string{"project=myproject"},
				DefaultSTSTTL:  "1h",
				MaxSTSTTL:      "4h",
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: assumed_role with externalID, sessionTags, and STS TTLs, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_IAMUserRejectsAssumedRoleFields(t *testing.T) {
	tests := []struct {
		name string
		role AWSRole
	}{
		{
			name: "rejects externalID",
			role: AWSRole{
				CredentialType: "iam_user",
				ExternalID:     "ext-123",
			},
		},
		{
			name: "rejects sessionTags",
			role: AWSRole{
				CredentialType: "iam_user",
				SessionTags:    []string{"tag=val"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &AWSSecretEngineRole{Spec: AWSSecretEngineRoleSpec{AWSRole: tt.role}}
			ok, err := role.IsValid()
			if ok || err == nil {
				t.Errorf("expected invalid: iam_user %s", tt.name)
			}
		})
	}
}

func TestAWSSecretEngineRole_IsValid_IAMUserRejectsSTSTTLs(t *testing.T) {
	tests := []struct {
		name string
		role AWSRole
	}{
		{
			name: "rejects defaultSTSTTL",
			role: AWSRole{
				CredentialType: "iam_user",
				DefaultSTSTTL:  "1h",
			},
		},
		{
			name: "rejects maxSTSTTL",
			role: AWSRole{
				CredentialType: "iam_user",
				MaxSTSTTL:      "4h",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &AWSSecretEngineRole{Spec: AWSSecretEngineRoleSpec{AWSRole: tt.role}}
			ok, err := role.IsValid()
			if ok || err == nil {
				t.Errorf("expected invalid: iam_user %s", tt.name)
			}
		})
	}
}

func TestAWSSecretEngineRole_IsValid_IAMUserAllowsOwnFields(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType:         "iam_user",
				PolicyArns:             []string{"arn:aws:iam::123456789012:policy/Pol"},
				IAMTags:                []string{"env=prod"},
				UserPath:               "/engineering/",
				PermissionsBoundaryARN: "arn:aws:iam::123456789012:policy/Boundary",
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: iam_user with iamTags, userPath, permissionsBoundaryARN, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_FederationTokenRejectsInapplicableFields(t *testing.T) {
	tests := []struct {
		name string
		role AWSRole
	}{
		{
			name: "rejects externalID",
			role: AWSRole{
				CredentialType: "federation_token",
				ExternalID:     "ext-123",
			},
		},
		{
			name: "rejects sessionTags",
			role: AWSRole{
				CredentialType: "federation_token",
				SessionTags:    []string{"tag=val"},
			},
		},
		{
			name: "rejects iamTags",
			role: AWSRole{
				CredentialType: "federation_token",
				IAMTags:        []string{"env=prod"},
			},
		},
		{
			name: "rejects userPath",
			role: AWSRole{
				CredentialType: "federation_token",
				UserPath:       "/path/",
			},
		},
		{
			name: "rejects permissionsBoundaryARN",
			role: AWSRole{
				CredentialType: "federation_token",
				PermissionsBoundaryARN: "arn:aws:iam::123456789012:policy/Boundary",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &AWSSecretEngineRole{Spec: AWSSecretEngineRoleSpec{AWSRole: tt.role}}
			ok, err := role.IsValid()
			if ok || err == nil {
				t.Errorf("expected invalid: federation_token %s", tt.name)
			}
		})
	}
}

func TestAWSSecretEngineRole_IsValid_FederationTokenAllowsSTSTTLs(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "federation_token",
				PolicyArns:     []string{"arn:aws:iam::123456789012:policy/Pol"},
				DefaultSTSTTL:  "30m",
				MaxSTSTTL:      "2h",
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: federation_token with STS TTLs, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_SessionTokenRejectsInapplicableFields(t *testing.T) {
	tests := []struct {
		name string
		role AWSRole
	}{
		{
			name: "rejects externalID",
			role: AWSRole{
				CredentialType: "session_token",
				ExternalID:     "ext-123",
			},
		},
		{
			name: "rejects sessionTags",
			role: AWSRole{
				CredentialType: "session_token",
				SessionTags:    []string{"tag=val"},
			},
		},
		{
			name: "rejects iamTags",
			role: AWSRole{
				CredentialType: "session_token",
				IAMTags:        []string{"env=prod"},
			},
		},
		{
			name: "rejects userPath",
			role: AWSRole{
				CredentialType: "session_token",
				UserPath:       "/path/",
			},
		},
		{
			name: "rejects permissionsBoundaryARN",
			role: AWSRole{
				CredentialType: "session_token",
				PermissionsBoundaryARN: "arn:aws:iam::123456789012:policy/Boundary",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &AWSSecretEngineRole{Spec: AWSSecretEngineRoleSpec{AWSRole: tt.role}}
			ok, err := role.IsValid()
			if ok || err == nil {
				t.Errorf("expected invalid: session_token %s", tt.name)
			}
		})
	}
}

func TestAWSSecretEngineRole_IsValid_SessionTokenRejectsSTSTTLs(t *testing.T) {
	tests := []struct {
		name string
		role AWSRole
	}{
		{
			name: "rejects defaultSTSTTL",
			role: AWSRole{
				CredentialType: "session_token",
				DefaultSTSTTL:  "30m",
			},
		},
		{
			name: "rejects maxSTSTTL",
			role: AWSRole{
				CredentialType: "session_token",
				MaxSTSTTL:      "2h",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &AWSSecretEngineRole{Spec: AWSSecretEngineRoleSpec{AWSRole: tt.role}}
			ok, err := role.IsValid()
			if ok || err == nil {
				t.Errorf("expected invalid: session_token %s", tt.name)
			}
		})
	}
}

func TestAWSSecretEngineRole_IsValid_SessionTokenRejectsIAMGroups(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "session_token",
				IAMGroups:      []string{"developers"},
			},
		},
	}
	ok, err := role.IsValid()
	if ok || err == nil {
		t.Error("expected invalid: session_token should not allow iamGroups")
	}
}

func TestAWSSecretEngineRole_IsValid_FederationTokenAllowsIAMGroups(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "federation_token",
				IAMGroups:      []string{"developers"},
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: federation_token supports iamGroups (group policies used in STS call), got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_AssumedRoleAllowsIAMGroups(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::123456789012:role/Role"},
				IAMGroups:      []string{"developers"},
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: assumed_role supports iamGroups (group policies used in STS call), got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsValid_IAMUserAllowsIAMGroups(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			AWSRole: AWSRole{
				CredentialType: "iam_user",
				IAMGroups:      []string{"developers"},
			},
		},
	}
	ok, err := role.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid: iam_user should allow iamGroups, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineRole_IsEquivalentToDesiredState_ReorderedSlicesMatch(t *testing.T) {
	role := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			AWSRole: AWSRole{
				CredentialType: "assumed_role",
				RoleArns:       []string{"arn:aws:iam::111:role/B", "arn:aws:iam::111:role/A"},
				PolicyArns:     []string{"arn:aws:iam::111:policy/Z", "arn:aws:iam::111:policy/A"},
				SessionTags:    []string{"env=prod", "app=web"},
			},
		},
	}

	vaultPayload := map[string]any{
		"credential_types": []any{"assumed_role"},
		"role_arns":        []any{"arn:aws:iam::111:role/A", "arn:aws:iam::111:role/B"},
		"policy_arns":      []any{"arn:aws:iam::111:policy/A", "arn:aws:iam::111:policy/Z"},
		"session_tags":     []any{"app=web", "env=prod"},
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: reordered set-like fields should not cause false drift")
	}
}

func TestAWSSecretEngineRole_ValidateUpdate_RejectsNameChange(t *testing.T) {
	r := &AWSSecretEngineRole{}
	oldObj := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			Name: "old-name",
			AWSRole: AWSRole{
				CredentialType: "iam_user",
			},
		},
	}
	newObj := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			Name: "new-name",
			AWSRole: AWSRole{
				CredentialType: "iam_user",
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.name is changed")
	}
}

func TestAWSSecretEngineRole_ValidateUpdate_AllowsSameNameUpdate(t *testing.T) {
	r := &AWSSecretEngineRole{}
	oldObj := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			Name: "same-name",
			AWSRole: AWSRole{
				CredentialType: "iam_user",
			},
		},
	}
	newObj := &AWSSecretEngineRole{
		Spec: AWSSecretEngineRoleSpec{
			Path: "aws",
			Name: "same-name",
			AWSRole: AWSRole{
				CredentialType: "iam_user",
				PolicyArns:     []string{"arn:aws:iam::111:policy/New"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err != nil {
		t.Errorf("expected no error when spec.name unchanged, got: %v", err)
	}
}
