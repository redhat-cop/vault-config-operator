package v1alpha1

import (
	"context"
	"reflect"
	"sort"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesAuthEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *KubernetesAuthEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &KubernetesAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: KubernetesAuthEngineRoleSpec{
					Path: "kubernetes",
					Name: "custom-name",
				},
			},
			expectedPath: "auth/kubernetes/role/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &KubernetesAuthEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: KubernetesAuthEngineRoleSpec{
					Path: "kubernetes",
				},
			},
			expectedPath: "auth/kubernetes/role/meta-name",
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

func TestVRoleToMap(t *testing.T) {
	aud := "my-audience"
	role := VRole{
		TargetServiceAccounts: []string{"sa1", "sa2"},
		Audience:              &aud,
		AliasNameSource:       "serviceaccount_uid",
		TokenTTL:              3600,
		TokenMaxTTL:           86400,
		Policies:              []string{"policy1", "policy2"},
		TokenBoundCIDRs:       []string{"10.0.0.0/8"},
		TokenExplicitMaxTTL:   0,
		TokenNoDefaultPolicy:  false,
		TokenNumUses:          0,
		TokenPeriod:           0,
		TokenType:             "default",
		namespaces:            []string{"ns1", "ns2"},
	}

	result := role.toMap()

	if len(result) != 13 {
		t.Errorf("expected 13 keys in map (with audience), got %d", len(result))
	}

	expected := map[string]interface{}{
		"bound_service_account_names":      []string{"sa1", "sa2"},
		"bound_service_account_namespaces": []string{"ns1", "ns2"},
		"alias_name_source":                "serviceaccount_uid",
		"audience":                         &aud,
		"token_ttl":                        3600,
		"token_max_ttl":                    86400,
		"token_policies":                   []string{"policy1", "policy2"},
		"token_bound_cidrs":                []string{"10.0.0.0/8"},
		"token_explicit_max_ttl":           0,
		"token_no_default_policy":          false,
		"token_num_uses":                   0,
		"token_period":                     0,
		"token_type":                       "default",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("toMap() mismatch:\n  got  %v\n  want %v", result, expected)
	}
}

func TestVRoleToMapAudienceNil(t *testing.T) {
	role := VRole{
		TargetServiceAccounts: []string{"default"},
		Policies:              []string{"reader"},
		namespaces:            []string{"default"},
	}

	result := role.toMap()

	if _, exists := result["audience"]; exists {
		t.Error("expected no 'audience' key when Audience is nil")
	}

	if len(result) != 12 {
		t.Errorf("expected 12 keys in map (without audience), got %d", len(result))
	}
}

func TestVRoleToMapAudienceSet(t *testing.T) {
	aud := "my-aud"
	role := VRole{
		TargetServiceAccounts: []string{"default"},
		Audience:              &aud,
		Policies:              []string{"reader"},
		namespaces:            []string{"default"},
	}

	result := role.toMap()

	audVal, exists := result["audience"]
	if !exists {
		t.Fatal("expected 'audience' key when Audience is non-nil")
	}

	audPtr, ok := audVal.(*string)
	if !ok {
		t.Fatalf("expected audience value to be *string, got %T", audVal)
	}
	if *audPtr != "my-aud" {
		t.Errorf("expected audience value 'my-aud', got '%s'", *audPtr)
	}

	if len(result) != 13 {
		t.Errorf("expected 13 keys in map (with audience), got %d", len(result))
	}
}

func TestVRoleToMapUnexportedNamespaces(t *testing.T) {
	role := VRole{
		TargetServiceAccounts: []string{"default"},
		Policies:              []string{"reader"},
	}
	role.namespaces = []string{"ns-a", "ns-b", "ns-c"}

	result := role.toMap()

	ns, ok := result["bound_service_account_namespaces"].([]string)
	if !ok {
		t.Fatal("expected bound_service_account_namespaces to be []string")
	}
	if !reflect.DeepEqual(ns, []string{"ns-a", "ns-b", "ns-c"}) {
		t.Errorf("expected namespaces [ns-a ns-b ns-c], got %v", ns)
	}
}

func TestKubernetesAuthEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &KubernetesAuthEngineRole{
		Spec: KubernetesAuthEngineRoleSpec{
			VRole: VRole{
				TargetServiceAccounts: []string{"sa1"},
				AliasNameSource:       "serviceaccount_uid",
				Policies:              []string{"policy1"},
				TokenType:             "default",
				namespaces:            []string{"default"},
			},
		},
	}

	payload := role.Spec.VRole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestKubernetesAuthEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &KubernetesAuthEngineRole{
		Spec: KubernetesAuthEngineRoleSpec{
			VRole: VRole{
				TargetServiceAccounts: []string{"sa1"},
				Policies:              []string{"policy1"},
				namespaces:            []string{"default"},
			},
		},
	}

	payload := role.Spec.VRole.toMap()
	payload["alias_name_source"] = "serviceaccount_name"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload (different alias_name_source) to NOT be equivalent")
	}
}

func TestKubernetesAuthEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &KubernetesAuthEngineRole{
		Spec: KubernetesAuthEngineRoleSpec{
			VRole: VRole{
				TargetServiceAccounts: []string{"sa1"},
				Policies:              []string{"policy1"},
				namespaces:            []string{"default"},
			},
		},
	}

	payload := role.Spec.VRole.toMap()
	payload["extra_vault_field"] = "unexpected"

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected extra fields to be ignored by filterPayloadToDesiredKeys")
	}
}

func TestKubernetesAuthEngineRoleIsDeletable(t *testing.T) {
	role := &KubernetesAuthEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected KubernetesAuthEngineRole to be deletable")
	}
}

func TestKubernetesAuthEngineRoleConditions(t *testing.T) {
	role := &KubernetesAuthEngineRole{}

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

func baseKubernetesAuthEngineRole() *KubernetesAuthEngineRole {
	return &KubernetesAuthEngineRole{
		ObjectMeta: metav1.ObjectMeta{Name: "kar", Namespace: "default"},
		Spec: KubernetesAuthEngineRoleSpec{
			Path: "kubernetes",
			VRole: VRole{
				TargetServiceAccounts: []string{"default"},
				Policies:              []string{"p1"},
			},
		},
	}
}

func TestKubernetesAuthEngineRole_PrepareInternalValues_ExplicitNamespaces(t *testing.T) {
	role := baseKubernetesAuthEngineRole()
	role.Spec.TargetNamespaces = vaultutils.TargetNamespaceConfig{
		TargetNamespaces: []string{"ns1", "ns2"},
	}
	ctx := pivContext(newFakeKubeClient(), nil)

	if err := role.PrepareInternalValues(ctx, role); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if !reflect.DeepEqual(role.Spec.namespaces, []string{"ns1", "ns2"}) {
		t.Fatalf("namespaces: got %v", role.Spec.namespaces)
	}
}

func TestKubernetesAuthEngineRole_PrepareInternalValues_LabelSelector(t *testing.T) {
	nsA := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "labeled-a", Labels: map[string]string{"env": "test"}},
	}
	nsB := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "labeled-b", Labels: map[string]string{"env": "test"}},
	}
	kube := newFakeKubeClient(nsA, nsB)
	role := baseKubernetesAuthEngineRole()
	role.Spec.TargetNamespaces = vaultutils.TargetNamespaceConfig{
		TargetNamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"env": "test"},
		},
	}
	handler := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)

	if err := role.PrepareInternalValues(ctx, role); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	got := append([]string(nil), role.Spec.namespaces...)
	sort.Strings(got)
	want := []string{"labeled-a", "labeled-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("namespaces: got %v want %v", got, want)
	}
}

func TestKubernetesAuthEngineRole_PrepareInternalValues_LabelSelectorNoMatch(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "other-ns", Labels: map[string]string{"env": "prod"}},
	}
	kube := newFakeKubeClient(ns)
	role := baseKubernetesAuthEngineRole()
	role.Spec.TargetNamespaces = vaultutils.TargetNamespaceConfig{
		TargetNamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"env": "staging"},
		},
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "kubeClient", kube)

	if err := role.PrepareInternalValues(ctx, role); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if !reflect.DeepEqual(role.Spec.namespaces, []string{"__no_namespace__"}) {
		t.Fatalf("namespaces: got %v", role.Spec.namespaces)
	}
}
