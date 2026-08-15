package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMongoDBAtlasSecretEngineRoleGetPath(t *testing.T) {
	role := &MongoDBAtlasSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
		},
	}
	expected := "mongodbatlas/roles/my-role"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestMongoDBAtlasSecretEngineRoleGetPathWithNameOverride(t *testing.T) {
	role := &MongoDBAtlasSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "k8s-name"},
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			Name: "vault-name",
		},
	}
	expected := "mongodbatlas/roles/vault-name"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestMongoDBAtlasSecretEngineRoleIsDeletable(t *testing.T) {
	role := &MongoDBAtlasSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected MongoDBAtlasSecretEngineRole to be deletable")
	}
}

func TestMongoDBAtlasSERole_toMap_OrgLevel(t *testing.T) {
	role := MongoDBAtlasSERole{
		OrganizationID: "org-123",
		Roles:          []string{"ORG_READ_ONLY", "ORG_MEMBER"},
		IPAddresses:    []string{"192.168.1.3", "192.168.1.4"},
		TTL:            "30m",
		MaxTTL:         "1h",
	}

	result := role.toMap()

	if result["organization_id"] != "org-123" {
		t.Errorf("organization_id = %v, expected org-123", result["organization_id"])
	}
	if result["project_id"] != "" {
		t.Errorf("project_id = %v, expected empty string", result["project_id"])
	}
	roles, ok := result["roles"].([]any)
	if !ok {
		t.Fatalf("roles should be []any, got %T", result["roles"])
	}
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
	ipAddrs, ok := result["ip_addresses"].([]any)
	if !ok {
		t.Fatalf("ip_addresses should be []any, got %T", result["ip_addresses"])
	}
	if len(ipAddrs) != 2 {
		t.Errorf("expected 2 ip_addresses, got %d", len(ipAddrs))
	}
	if result["ttl"] != "30m" {
		t.Errorf("ttl = %v, expected 30m", result["ttl"])
	}
	if result["max_ttl"] != "1h" {
		t.Errorf("max_ttl = %v, expected 1h", result["max_ttl"])
	}
}

func TestMongoDBAtlasSERole_toMap_ProjectLevel(t *testing.T) {
	role := MongoDBAtlasSERole{
		ProjectID:  "proj-456",
		Roles:      []string{"GROUP_CLUSTER_MANAGER"},
		CIDRBlocks: []string{"192.168.1.0/24"},
	}

	result := role.toMap()

	if result["project_id"] != "proj-456" {
		t.Errorf("project_id = %v, expected proj-456", result["project_id"])
	}
	if result["organization_id"] != "" {
		t.Errorf("organization_id = %v, expected empty string", result["organization_id"])
	}
	roles, ok := result["roles"].([]any)
	if !ok {
		t.Fatalf("roles should be []any, got %T", result["roles"])
	}
	if len(roles) != 1 || roles[0] != "GROUP_CLUSTER_MANAGER" {
		t.Errorf("roles = %v, expected [GROUP_CLUSTER_MANAGER]", roles)
	}
	cidrBlocks, ok := result["cidr_blocks"].([]any)
	if !ok {
		t.Fatalf("cidr_blocks should be []any, got %T", result["cidr_blocks"])
	}
	if len(cidrBlocks) != 1 || cidrBlocks[0] != "192.168.1.0/24" {
		t.Errorf("cidr_blocks = %v, expected [192.168.1.0/24]", cidrBlocks)
	}
}

func TestMongoDBAtlasSecretEngineRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	role := &MongoDBAtlasSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSERole: MongoDBAtlasSERole{
				ProjectID: "proj-456",
				Roles:     []string{"GROUP_CLUSTER_MANAGER"},
				TTL:       "30m",
				MaxTTL:    "1h",
			},
		},
	}

	vaultPayload := map[string]any{
		"project_id":      "proj-456",
		"organization_id": "",
		"roles":           []any{"GROUP_CLUSTER_MANAGER"},
		"ip_addresses":    []any{},
		"cidr_blocks":     []any{},
		"project_roles":   []any{},
		"ttl":             "30m",
		"max_ttl":         "1h",
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when all managed fields match")
	}
}

func TestMongoDBAtlasSecretEngineRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &MongoDBAtlasSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSERole: MongoDBAtlasSERole{
				ProjectID: "proj-456",
				Roles:     []string{"GROUP_CLUSTER_MANAGER"},
				TTL:       "30m",
			},
		},
	}

	vaultPayload := map[string]any{
		"project_id":      "proj-456",
		"organization_id": "",
		"roles":           []any{"GROUP_READ_ONLY"},
		"ip_addresses":    []any{},
		"cidr_blocks":     []any{},
		"project_roles":   []any{},
		"ttl":             "30m",
		"max_ttl":         "",
	}

	if role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when roles differ")
	}
}

func TestMongoDBAtlasSecretEngineRole_IsEquivalentToDesiredState_UnsetFields(t *testing.T) {
	role := &MongoDBAtlasSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSERole: MongoDBAtlasSERole{
				ProjectID: "proj-456",
				Roles:     []string{"GROUP_CLUSTER_MANAGER"},
			},
		},
	}

	vaultPayload := map[string]any{
		"project_id": "proj-456",
		"roles":      []any{"GROUP_CLUSTER_MANAGER"},
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: unset fields absent from Vault should not cause false drift")
	}
}

func TestMongoDBAtlasSecretEngineRole_IsEquivalentToDesiredState_UnsortedRoles(t *testing.T) {
	role := &MongoDBAtlasSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			MongoDBAtlasSERole: MongoDBAtlasSERole{
				OrganizationID: "org-123",
				Roles:          []string{"ORG_READ_ONLY", "ORG_MEMBER", "ORG_OWNER"},
				IPAddresses:    []string{"10.0.0.2", "10.0.0.1"},
			},
		},
	}

	vaultPayload := map[string]any{
		"organization_id": "org-123",
		"project_id":      "",
		"roles":           []any{"ORG_OWNER", "ORG_READ_ONLY", "ORG_MEMBER"},
		"ip_addresses":    []any{"10.0.0.1", "10.0.0.2"},
		"cidr_blocks":     []any{},
		"project_roles":   []any{},
		"ttl":             "",
		"max_ttl":         "",
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: set fields should match regardless of order")
	}
}

func TestMongoDBAtlasSecretEngineRoleConditions(t *testing.T) {
	role := &MongoDBAtlasSecretEngineRole{}

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
}

func TestMongoDBAtlasSecretEngineRole_ValidateUpdate_RejectsPathChange(t *testing.T) {
	r := &MongoDBAtlasSecretEngineRole{}
	oldObj := &MongoDBAtlasSecretEngineRole{
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			Name: "role-1",
		},
	}
	newObj := &MongoDBAtlasSecretEngineRole{
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "different-path",
			Name: "role-1",
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.path is changed")
	}
}

func TestMongoDBAtlasSecretEngineRole_ValidateUpdate_RejectsNameChange(t *testing.T) {
	r := &MongoDBAtlasSecretEngineRole{}
	oldObj := &MongoDBAtlasSecretEngineRole{
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			Name: "old-name",
		},
	}
	newObj := &MongoDBAtlasSecretEngineRole{
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			Name: "new-name",
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.name is changed")
	}
}

func TestMongoDBAtlasSecretEngineRole_ValidateUpdate_AllowsFieldUpdate(t *testing.T) {
	r := &MongoDBAtlasSecretEngineRole{}
	oldObj := &MongoDBAtlasSecretEngineRole{
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			Name: "role-1",
			MongoDBAtlasSERole: MongoDBAtlasSERole{
				Roles: []string{"ORG_READ_ONLY"},
			},
		},
	}
	newObj := &MongoDBAtlasSecretEngineRole{
		Spec: MongoDBAtlasSecretEngineRoleSpec{
			Path: "mongodbatlas",
			Name: "role-1",
			MongoDBAtlasSERole: MongoDBAtlasSERole{
				Roles: []string{"ORG_OWNER"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err != nil {
		t.Errorf("expected no error when only role fields changed, got: %v", err)
	}
}

func TestMongoDBAtlasSERole_toMap_AllFields(t *testing.T) {
	role := MongoDBAtlasSERole{
		OrganizationID: "org-123",
		ProjectID:      "proj-456",
		Roles:          []string{"ORG_READ_ONLY"},
		IPAddresses:    []string{"192.168.1.3"},
		CIDRBlocks:     []string{"10.0.0.0/8"},
		ProjectRoles:   []string{"GROUP_READ_ONLY"},
		TTL:            "2h",
		MaxTTL:         "4h",
	}

	result := role.toMap()

	expectedKeys := []string{"organization_id", "project_id", "roles", "ip_addresses", "cidr_blocks", "project_roles", "ttl", "max_ttl"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys, got %d: %v", len(expectedKeys), len(result), result)
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}
}
