package v1alpha1

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPolicyGetPath(t *testing.T) {
	tests := []struct {
		name         string
		policy       *Policy
		expectedPath string
	}{
		{
			name: "with spec.name and spec.type set",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Name: "spec-name", Type: "acl"},
			},
			expectedPath: "sys/policies/acl/spec-name",
		},
		{
			name: "with spec.name set but no type",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Name: "spec-name"},
			},
			expectedPath: "sys/policy/spec-name",
		},
		{
			name: "without spec.name but with type",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Type: "acl"},
			},
			expectedPath: "sys/policies/acl/meta-name",
		},
		{
			name: "without spec.name and without type",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{},
			},
			expectedPath: "sys/policy/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.policy.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestPolicyGetPayload(t *testing.T) {
	policy := &Policy{
		Spec: PolicySpec{
			Policy: `path "secret/*" { capabilities = ["read"] }`,
		},
	}

	payload := policy.GetPayload()

	if len(payload) != 1 {
		t.Errorf("expected exactly 1 key in payload, got %d", len(payload))
	}
	if payload["policy"] != policy.Spec.Policy {
		t.Errorf("expected policy text %q, got %v", policy.Spec.Policy, payload["policy"])
	}
}

func TestPolicyIsEquivalentNoType(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
		},
	}

	// When Type == "", the method renames "policy" → "rules" and adds "name"
	payload := map[string]any{
		"name":  "meta-name",
		"rules": policyText,
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (no type, rules key) to be equivalent")
	}
}

func TestPolicyIsEquivalentWithType(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
			Type:   "acl",
		},
	}

	// When Type == "acl", the method keeps "policy" key and adds "name"
	payload := map[string]any{
		"name":   "meta-name",
		"policy": policyText,
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload (with type, policy key) to be equivalent")
	}
}

func TestPolicyIsEquivalentNameFromSpec(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
			Name:   "spec-name",
		},
	}

	payload := map[string]any{
		"name":  "spec-name",
		"rules": policyText,
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected name to come from Spec.Name when set")
	}
}

func TestPolicyIsEquivalentNameFromMetadata(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
		},
	}

	payload := map[string]any{
		"name":  "meta-name",
		"rules": policyText,
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected name to fall back to metadata name when Spec.Name is empty")
	}
}

func TestPolicyIsEquivalentNonMatching(t *testing.T) {
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: `path "secret/*" { capabilities = ["read"] }`,
		},
	}

	payload := map[string]any{
		"name":  "meta-name",
		"rules": `path "secret/*" { capabilities = ["list"] }`,
	}
	if policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected different policy text to not be equivalent")
	}
}

// Payload keys not present in desiredState are filtered before comparison.
func TestPolicyIsEquivalentExtraFields(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
		Spec: PolicySpec{
			Policy: policyText,
		},
	}

	payload := map[string]any{
		"name":        "meta-name",
		"rules":       policyText,
		"extra_field": "unexpected",
	}
	if !policy.IsEquivalentToDesiredState(payload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestPolicyIsDeletable(t *testing.T) {
	policy := &Policy{}
	if !policy.IsDeletable() {
		t.Error("expected Policy to be deletable")
	}
}

func TestPolicyConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	policy := &Policy{}
	policy.SetConditions([]metav1.Condition{condition})
	got := policy.GetConditions()
	if len(got) != 1 || got[0].Type != "Ready" {
		t.Error("expected Policy conditions to be set and retrieved")
	}
}

// Verify the internal structure of the desiredState map produced by
// IsEquivalentToDesiredState across all 4 Type/Name combinations.
func TestPolicyIsEquivalentAllVariants(t *testing.T) {
	policyText := `path "secret/*" { capabilities = ["read"] }`

	tests := []struct {
		name    string
		policy  *Policy
		payload map[string]any
		want    bool
	}{
		{
			name: "no type, spec.name empty → rules key, metadata name",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Policy: policyText},
			},
			payload: map[string]any{"name": "meta-name", "rules": policyText},
			want:    true,
		},
		{
			name: "no type, spec.name set → rules key, spec name",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Policy: policyText, Name: "spec-name"},
			},
			payload: map[string]any{"name": "spec-name", "rules": policyText},
			want:    true,
		},
		{
			name: "type acl, spec.name empty → policy key, metadata name",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Policy: policyText, Type: "acl"},
			},
			payload: map[string]any{"name": "meta-name", "policy": policyText},
			want:    true,
		},
		{
			name: "type acl, spec.name set → policy key, spec name",
			policy: &Policy{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec:       PolicySpec{Policy: policyText, Type: "acl", Name: "spec-name"},
			},
			payload: map[string]any{"name": "spec-name", "policy": policyText},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policy.IsEquivalentToDesiredState(tt.payload)
			if got != tt.want {
				desiredState := tt.policy.GetPayload()
				desiredState["name"] = map[bool]string{true: tt.policy.Spec.Name, false: tt.policy.Name}[tt.policy.Spec.Name != ""]
				if tt.policy.Spec.Type == "" {
					desiredState["rules"] = desiredState["policy"]
					delete(desiredState, "policy")
				}
				t.Errorf("IsEquivalentToDesiredState() = %v, want %v\ndesiredState: %v\npayload:      %v\nmatch: %v",
					got, tt.want, desiredState, tt.payload, reflect.DeepEqual(desiredState, tt.payload))
			}
		})
	}
}

func TestPolicy_PrepareInternalValues_NoPlaceholderFastPath(t *testing.T) {
	want := `path "secret/*" { capabilities = ["read"] }`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "default"},
		Spec:       PolicySpec{Policy: want},
	}
	kube := newFakeKubeClient()
	handler := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)

	if err := policy.PrepareInternalValues(ctx, policy); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if policy.Spec.Policy != want {
		t.Fatalf("policy text changed: got %q want %q", policy.Spec.Policy, want)
	}
}

func TestPolicy_PrepareInternalValues_ReplaceKubernetesAccessor(t *testing.T) {
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "default"},
		Spec:       PolicySpec{Policy: `grant "${auth/kubernetes/@accessor}"`},
	}
	handler := newFakeVaultHandler()
	handler.setGet("sys/auth", map[string]any{
		"kubernetes/": map[string]any{
			"accessor": "auth_kubernetes_abc123",
			"type":     "kubernetes",
		},
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	if err := policy.PrepareInternalValues(ctx, policy); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	want := `grant "auth_kubernetes_abc123"`
	if policy.Spec.Policy != want {
		t.Fatalf("policy: got %q want %q", policy.Spec.Policy, want)
	}
}

func TestPolicy_PrepareInternalValues_MultipleAuthEngines(t *testing.T) {
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "default"},
		Spec:       PolicySpec{Policy: `a:${auth/kubernetes/@accessor} b:${auth/ldap/@accessor}`},
	}
	handler := newFakeVaultHandler()
	handler.setGet("sys/auth", map[string]any{
		"kubernetes/": map[string]any{"accessor": "acc-k8s", "type": "kubernetes"},
		"ldap/":       map[string]any{"accessor": "acc-ldap", "type": "ldap"},
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	if err := policy.PrepareInternalValues(ctx, policy); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	want := `a:acc-k8s b:acc-ldap`
	if policy.Spec.Policy != want {
		t.Fatalf("policy: got %q want %q", policy.Spec.Policy, want)
	}
}

func TestPolicy_PrepareInternalValues_ReadErrorSwallowedPlaceholdersUnchanged(t *testing.T) {
	original := `grant "${auth/kubernetes/@accessor}"`
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "default"},
		Spec:       PolicySpec{Policy: original},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	err := policy.PrepareInternalValues(ctx, policy)
	if err != nil {
		t.Fatalf("expected nil error on sys/auth read failure, got: %v", err)
	}
	if policy.Spec.Policy != original {
		t.Fatalf("policy text should be unchanged on read error: got %q want %q", policy.Spec.Policy, original)
	}
}

func TestPolicy_PrepareInternalValues_NilSecretReturnsError(t *testing.T) {
	policy := &Policy{
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "default"},
		Spec:       PolicySpec{Policy: `use ${auth/kubernetes/@accessor}`},
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/v1/"):]
		if r.Method == http.MethodGet && path == "sys/auth" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(newFakeKubeClient(), vc)

	err := policy.PrepareInternalValues(ctx, policy)
	if err == nil {
		t.Fatal("expected error when sys/auth returns 204 (nil secret)")
	}
	if !strings.Contains(err.Error(), "unexpectedly returned null") {
		t.Fatalf("unexpected error: %v", err)
	}
}
