package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConsulSecretEngineRoleGetPath(t *testing.T) {
	role := &ConsulSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: ConsulSecretEngineRoleSpec{
			Path: "consul",
		},
	}
	expected := "consul/roles/my-role"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestConsulSecretEngineRoleGetPathWithNameOverride(t *testing.T) {
	role := &ConsulSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: ConsulSecretEngineRoleSpec{
			Path: "consul",
			Name: "custom-name",
		},
	}
	expected := "consul/roles/custom-name"
	if got := role.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestConsulSecretEngineRoleIsDeletable(t *testing.T) {
	role := &ConsulSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected ConsulSecretEngineRole to be deletable")
	}
}

func TestConsulSERole_toMap_Policies(t *testing.T) {
	role := ConsulSERole{
		ConsulPolicies: []string{"readonly", "admin"},
		ConsulRoles:    []string{"web-role"},
		TTL:            "1h",
		MaxTTL:         "24h",
	}

	result := role.toMap()

	if result["ttl"] != "1h" {
		t.Errorf("ttl = %v, expected 1h", result["ttl"])
	}
	if result["max_ttl"] != "24h" {
		t.Errorf("max_ttl = %v, expected 24h", result["max_ttl"])
	}

	consulPolicies, ok := result["consul_policies"].([]any)
	if !ok {
		t.Fatalf("consul_policies type = %T, expected []any", result["consul_policies"])
	}
	if len(consulPolicies) != 2 {
		t.Fatalf("consul_policies length = %d, expected 2", len(consulPolicies))
	}
	if consulPolicies[0] != "readonly" || consulPolicies[1] != "admin" {
		t.Errorf("consul_policies = %v, expected [readonly admin]", consulPolicies)
	}

	consulRoles, ok := result["consul_roles"].([]any)
	if !ok {
		t.Fatalf("consul_roles type = %T, expected []any", result["consul_roles"])
	}
	if len(consulRoles) != 1 || consulRoles[0] != "web-role" {
		t.Errorf("consul_roles = %v, expected [web-role]", consulRoles)
	}
}

func TestConsulSERole_toMap_ServiceIdentities(t *testing.T) {
	role := ConsulSERole{
		ServiceIdentities: []string{"web:dc1,dc2", "api:dc1"},
		NodeIdentities:    []string{"node1:dc1"},
	}

	result := role.toMap()

	si, ok := result["service_identities"].([]any)
	if !ok {
		t.Fatalf("service_identities type = %T, expected []any", result["service_identities"])
	}
	if len(si) != 2 {
		t.Fatalf("service_identities length = %d, expected 2", len(si))
	}
	if si[0] != "web:dc1,dc2" || si[1] != "api:dc1" {
		t.Errorf("service_identities = %v, expected [web:dc1,dc2 api:dc1]", si)
	}

	ni, ok := result["node_identities"].([]any)
	if !ok {
		t.Fatalf("node_identities type = %T, expected []any", result["node_identities"])
	}
	if len(ni) != 1 || ni[0] != "node1:dc1" {
		t.Errorf("node_identities = %v, expected [node1:dc1]", ni)
	}
}

func TestConsulSERole_toMap_AllFields(t *testing.T) {
	role := ConsulSERole{
		ConsulPolicies:    []string{"policy1"},
		ConsulRoles:       []string{"role1"},
		ServiceIdentities: []string{"svc1:dc1"},
		NodeIdentities:    []string{"node1:dc1"},
		ConsulNamespace:   "my-ns",
		Partition:         "my-partition",
		Local:             true,
		TTL:               "2h",
		MaxTTL:            "48h",
	}

	result := role.toMap()

	expectedKeys := []string{"consul_policies", "consul_roles", "service_identities", "node_identities", "consul_namespace", "partition", "local", "ttl", "max_ttl"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys in toMap() output, got %d: %v", len(expectedKeys), len(result), result)
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if result["consul_namespace"] != "my-ns" {
		t.Errorf("consul_namespace = %v, expected my-ns", result["consul_namespace"])
	}
	if result["partition"] != "my-partition" {
		t.Errorf("partition = %v, expected my-partition", result["partition"])
	}
	if result["local"] != true {
		t.Errorf("local = %v, expected true", result["local"])
	}
}

func TestConsulSecretEngineRole_IsEquivalentToDesiredState_Match(t *testing.T) {
	role := &ConsulSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: ConsulSecretEngineRoleSpec{
			Path: "consul",
			ConsulSERole: ConsulSERole{
				ConsulPolicies: []string{"readonly"},
				TTL:            "1h",
				MaxTTL:         "24h",
				Local:          false,
			},
		},
	}

	vaultPayload := map[string]any{
		"consul_policies":    []any{"readonly"},
		"consul_roles":       []any{},
		"service_identities": []any{},
		"node_identities":    []any{},
		"consul_namespace":   "",
		"partition":          "",
		"local":              false,
		"ttl":                "1h",
		"max_ttl":            "24h",
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when all managed fields match")
	}
}

func TestConsulSecretEngineRole_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	role := &ConsulSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: ConsulSecretEngineRoleSpec{
			Path: "consul",
			ConsulSERole: ConsulSERole{
				ConsulPolicies: []string{"readonly"},
				TTL:            "1h",
			},
		},
	}

	vaultPayload := map[string]any{
		"consul_policies":    []any{"readonly"},
		"consul_roles":       []any{},
		"service_identities": []any{},
		"node_identities":    []any{},
		"consul_namespace":   "",
		"partition":          "",
		"local":              false,
		"ttl":                "2h",
		"max_ttl":            "",
	}

	if role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when ttl differs")
	}
}

func TestConsulSecretEngineRole_IsEquivalentToDesiredState_UnsetFields(t *testing.T) {
	role := &ConsulSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: ConsulSecretEngineRoleSpec{
			Path: "consul",
			ConsulSERole: ConsulSERole{
				ConsulPolicies: []string{"readonly"},
				TTL:            "1h",
			},
		},
	}

	vaultPayload := map[string]any{
		"consul_policies": []any{"readonly"},
		"ttl":             "1h",
		"local":           false,
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: unset fields (empty strings/slices) absent from Vault should not cause false drift")
	}
}

func TestConsulSecretEngineRole_IsEquivalentToDesiredState_UnsortedPolicies(t *testing.T) {
	role := &ConsulSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: ConsulSecretEngineRoleSpec{
			Path: "consul",
			ConsulSERole: ConsulSERole{
				ConsulPolicies: []string{"beta", "alpha", "gamma"},
			},
		},
	}

	vaultPayload := map[string]any{
		"consul_policies": []any{"gamma", "alpha", "beta"},
		"local":           false,
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: set fields should be compared order-independently")
	}
}

func TestConsulSecretEngineRole_IsEquivalentToDesiredState_PolicyContentMismatch(t *testing.T) {
	role := &ConsulSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: ConsulSecretEngineRoleSpec{
			Path: "consul",
			ConsulSERole: ConsulSERole{
				ConsulPolicies: []string{"alpha", "beta"},
			},
		},
	}

	vaultPayload := map[string]any{
		"consul_policies": []any{"alpha", "gamma"},
		"local":           false,
	}

	if role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: different policy content must be detected")
	}
}

func TestConsulSecretEngineRole_IsEquivalentToDesiredState_LocalFalseOmittedByVault(t *testing.T) {
	role := &ConsulSecretEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "my-role"},
		Spec: ConsulSecretEngineRoleSpec{
			Path: "consul",
			ConsulSERole: ConsulSERole{
				ConsulPolicies: []string{"readonly"},
				TTL:            "1h",
				Local:          false,
			},
		},
	}

	vaultPayload := map[string]any{
		"consul_policies": []any{"readonly"},
		"ttl":             "1h",
	}

	if !role.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: local=false omitted by Vault should not cause drift")
	}
}

func TestConsulSecretEngineRoleConditions(t *testing.T) {
	role := &ConsulSecretEngineRole{}

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
