package v1alpha1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesSecretEngineRoleGetPath(t *testing.T) {
	tests := []struct {
		name         string
		role         *KubernetesSecretEngineRole
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			role: &KubernetesSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: KubernetesSecretEngineRoleSpec{
					Path: "kubernetes",
					Name: "custom-name",
				},
			},
			expectedPath: "kubernetes/roles/custom-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			role: &KubernetesSecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: KubernetesSecretEngineRoleSpec{
					Path: "kubernetes",
				},
			},
			expectedPath: "kubernetes/roles/meta-name",
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

func TestKubeSERoleToMap(t *testing.T) {
	role := KubeSERole{
		AllowedKubernetesNamespaces:        []string{"ns1", "ns2"},
		AllowedKubernetesNamespaceSelector: `{"matchLabels":{"app":"vault"}}`,
		DefaultTTL:                         metav1.Duration{Duration: time.Hour},
		MaxTTL:                             metav1.Duration{Duration: 24 * time.Hour},
		DefaultAudiences:                   "aud1,aud2",
		ServiceAccountName:                 "sa-name",
		KubernetesRoleName:                 "role-name",
		KubernetesRoleType:                 "ClusterRole",
		GenerateRoleRules:                  `{"rules":[]}`,
		NameTemplate:                       "vault-{{.Name}}",
		ExtraAnnotations:                   map[string]string{"ann": "v1"},
		ExtraLabels:                        map[string]string{"lbl": "v2"},
	}

	result := role.toMap()

	expectedKeys := []string{
		"allowed_kubernetes_namespaces",
		"allowed_kubernetes_namespace_selector",
		"token_max_ttl",
		"token_default_ttl",
		"token_default_audiences",
		"service_account_name",
		"kubernetes_role_name",
		"kubernetes_role_type",
		"generated_role_rules",
		"name_template",
		"extra_annotations",
		"extra_labels",
	}
	if len(result) != 12 {
		t.Fatalf("expected 12 keys in toMap() output, got %d", len(result))
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	ns, ok := result["allowed_kubernetes_namespaces"].([]string)
	if !ok {
		t.Fatalf("allowed_kubernetes_namespaces = %T", result["allowed_kubernetes_namespaces"])
	}
	if len(ns) != 2 || ns[0] != "ns1" || ns[1] != "ns2" {
		t.Errorf("allowed_kubernetes_namespaces = %v", ns)
	}

	if result["allowed_kubernetes_namespace_selector"] != `{"matchLabels":{"app":"vault"}}` {
		t.Errorf("allowed_kubernetes_namespace_selector = %v", result["allowed_kubernetes_namespace_selector"])
	}

	tokenMax, ok := result["token_max_ttl"].(metav1.Duration)
	if !ok {
		t.Fatalf("token_max_ttl should be metav1.Duration, got %T", result["token_max_ttl"])
	}
	if tokenMax.Duration != 24*time.Hour {
		t.Errorf("token_max_ttl = %v, expected 24h (from MaxTTL)", tokenMax.Duration)
	}

	tokenDef, ok := result["token_default_ttl"].(metav1.Duration)
	if !ok {
		t.Fatalf("token_default_ttl should be metav1.Duration, got %T", result["token_default_ttl"])
	}
	if tokenDef.Duration != time.Hour {
		t.Errorf("token_default_ttl = %v, expected 1h (from DefaultTTL)", tokenDef.Duration)
	}

	ann, ok := result["extra_annotations"].(map[string]string)
	if !ok {
		t.Fatalf("extra_annotations = %T", result["extra_annotations"])
	}
	if ann["ann"] != "v1" {
		t.Errorf("extra_annotations = %v", ann)
	}

	lbl, ok := result["extra_labels"].(map[string]string)
	if !ok {
		t.Fatalf("extra_labels = %T", result["extra_labels"])
	}
	if lbl["lbl"] != "v2" {
		t.Errorf("extra_labels = %v", lbl)
	}
}

func TestKubernetesSecretEngineRoleIsEquivalentMatching(t *testing.T) {
	role := &KubernetesSecretEngineRole{
		Spec: KubernetesSecretEngineRoleSpec{
			Path: "kubernetes",
			KubeSERole: KubeSERole{
				AllowedKubernetesNamespaces: []string{"default"},
				DefaultTTL:                  metav1.Duration{Duration: time.Hour},
				MaxTTL:                      metav1.Duration{Duration: 24 * time.Hour},
			},
		},
	}

	payload := role.Spec.KubeSERole.toMap()

	if !role.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestKubernetesSecretEngineRoleIsEquivalentNonMatching(t *testing.T) {
	role := &KubernetesSecretEngineRole{
		Spec: KubernetesSecretEngineRoleSpec{
			Path: "kubernetes",
			KubeSERole: KubeSERole{
				ServiceAccountName: "sa1",
			},
		},
	}

	payload := role.Spec.KubeSERole.toMap()
	payload["service_account_name"] = "sa2"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected non-matching payload to NOT be equivalent")
	}
}

func TestKubernetesSecretEngineRoleIsEquivalentExtraFields(t *testing.T) {
	role := &KubernetesSecretEngineRole{
		Spec: KubernetesSecretEngineRoleSpec{
			Path: "kubernetes",
			KubeSERole: KubeSERole{
				KubernetesRoleName: "r1",
			},
		},
	}

	payload := role.Spec.KubeSERole.toMap()
	payload["extra_vault_field"] = "some-value"

	if role.IsEquivalentToDesiredState(payload) {
		t.Error("expected payload with extra fields to NOT be equivalent")
	}
}

func TestKubernetesSecretEngineRoleIsDeletable(t *testing.T) {
	role := &KubernetesSecretEngineRole{}
	if !role.IsDeletable() {
		t.Error("expected KubernetesSecretEngineRole to be deletable")
	}
}

func TestKubernetesSecretEngineRoleConditions(t *testing.T) {
	role := &KubernetesSecretEngineRole{}

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
