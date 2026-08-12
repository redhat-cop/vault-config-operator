package v1alpha1

import (
	"reflect"
	"testing"
)

func TestFilterPayloadToDesiredKeysBasic(t *testing.T) {
	desiredState := map[string]any{
		"key1": "val1",
		"key2": 42,
	}
	payload := map[string]any{
		"key1":  "val1",
		"key2":  42,
		"extra": "should-be-filtered",
	}

	filtered := filterPayloadToDesiredKeys(desiredState, payload)

	expected := map[string]any{
		"key1": "val1",
		"key2": 42,
	}
	if !reflect.DeepEqual(filtered, expected) {
		t.Errorf("filterPayloadToDesiredKeys() = %v, want %v", filtered, expected)
	}
}

func TestFilterPayloadToDesiredKeysPreservesPayloadValues(t *testing.T) {
	desiredState := map[string]any{
		"name": "original",
	}
	payload := map[string]any{
		"name":  "vault-returned",
		"extra": "ignored",
	}

	filtered := filterPayloadToDesiredKeys(desiredState, payload)

	if filtered["name"] != "vault-returned" {
		t.Errorf("expected payload value 'vault-returned', got %v", filtered["name"])
	}
	if len(filtered) != 1 {
		t.Errorf("expected 1 key in filtered map, got %d", len(filtered))
	}
}

func TestFilterPayloadToDesiredKeysMissingKeyInPayload(t *testing.T) {
	desiredState := map[string]any{
		"present": "yes",
		"missing": "not-in-payload",
	}
	payload := map[string]any{
		"present": "yes",
		"extra":   "ignored",
	}

	filtered := filterPayloadToDesiredKeys(desiredState, payload)

	if len(filtered) != 1 {
		t.Errorf("expected 1 key (only 'present'), got %d keys", len(filtered))
	}
	if filtered["present"] != "yes" {
		t.Errorf("expected present='yes', got %v", filtered["present"])
	}
}

func TestFilterPayloadToDesiredKeysEmptyDesiredState(t *testing.T) {
	desiredState := map[string]any{}
	payload := map[string]any{
		"key1": "val1",
	}

	filtered := filterPayloadToDesiredKeys(desiredState, payload)

	if len(filtered) != 0 {
		t.Errorf("expected empty filtered map, got %d keys", len(filtered))
	}
}

func TestFilterPayloadToDesiredKeysDoesNotMutateInputs(t *testing.T) {
	desiredState := map[string]any{
		"key1": "val1",
	}
	payload := map[string]any{
		"key1":  "val1",
		"extra": "should-remain",
	}

	filterPayloadToDesiredKeys(desiredState, payload)

	if len(payload) != 2 {
		t.Errorf("expected payload to still have 2 keys, got %d", len(payload))
	}
	if _, ok := payload["extra"]; !ok {
		t.Error("expected 'extra' key to remain in original payload")
	}
}

func TestRemoveUnsetFields_ZeroStringRemovedWhenAbsentFromPayload(t *testing.T) {
	desiredState := map[string]any{
		"region":       "us-west-2",
		"iam_endpoint": "",
	}
	payload := map[string]any{
		"region": "us-west-2",
	}

	removeUnsetFields(desiredState, payload)

	if _, ok := desiredState["iam_endpoint"]; ok {
		t.Error("expected iam_endpoint to be removed (zero-valued and absent from payload)")
	}
	if _, ok := desiredState["region"]; !ok {
		t.Error("expected region to remain (present in payload)")
	}
}

func TestRemoveUnsetFields_ZeroStringKeptWhenPresentInPayload(t *testing.T) {
	desiredState := map[string]any{
		"region": "",
	}
	payload := map[string]any{
		"region": "us-west-2",
	}

	removeUnsetFields(desiredState, payload)

	if _, ok := desiredState["region"]; !ok {
		t.Error("expected region to remain (present in payload, needed for drift detection)")
	}
}

func TestRemoveUnsetFields_EmptySliceRemovedWhenAbsentFromPayload(t *testing.T) {
	desiredState := map[string]any{
		"credential_types": []any{"iam_user"},
		"policy_arns":      []any{},
	}
	payload := map[string]any{
		"credential_types": []any{"iam_user"},
	}

	removeUnsetFields(desiredState, payload)

	if _, ok := desiredState["policy_arns"]; ok {
		t.Error("expected policy_arns to be removed (empty slice and absent from payload)")
	}
}

func TestRemoveUnsetFields_EmptySliceKeptWhenPresentInPayload(t *testing.T) {
	desiredState := map[string]any{
		"policy_arns": []any{},
	}
	payload := map[string]any{
		"policy_arns": []any{"arn:aws:iam::123456789012:policy/OldPolicy"},
	}

	removeUnsetFields(desiredState, payload)

	if _, ok := desiredState["policy_arns"]; !ok {
		t.Error("expected policy_arns to remain (present in payload with stale value)")
	}
}

func TestRemoveUnsetFields_NonZeroValuesNeverRemoved(t *testing.T) {
	desiredState := map[string]any{
		"region":      "us-west-2",
		"policy_arns": []any{"arn:aws:iam::123456789012:policy/Pol"},
	}
	payload := map[string]any{}

	removeUnsetFields(desiredState, payload)

	if _, ok := desiredState["region"]; !ok {
		t.Error("expected non-zero region to remain even when absent from payload")
	}
	if _, ok := desiredState["policy_arns"]; !ok {
		t.Error("expected non-empty policy_arns to remain even when absent from payload")
	}
}

func TestRemoveUnsetFields_DefaultIntRemovedWhenAbsentFromPayload(t *testing.T) {
	desiredState := map[string]any{
		"region":      "us-west-2",
		"max_retries": -1,
	}
	payload := map[string]any{
		"region": "us-west-2",
	}

	removeUnsetFields(desiredState, payload)

	if _, ok := desiredState["max_retries"]; ok {
		t.Error("expected max_retries (-1 default) to be removed when absent from payload")
	}
	if _, ok := desiredState["region"]; !ok {
		t.Error("expected region to remain")
	}
}

func TestRemoveUnsetFields_IntKeptWhenPresentInPayload(t *testing.T) {
	desiredState := map[string]any{
		"max_retries": -1,
	}
	payload := map[string]any{
		"max_retries": 5,
	}

	removeUnsetFields(desiredState, payload)

	if _, ok := desiredState["max_retries"]; !ok {
		t.Error("expected max_retries to remain (present in payload, needed for drift detection)")
	}
}

func TestRemoveUnsetFields_NonDefaultIntNeverRemoved(t *testing.T) {
	desiredState := map[string]any{
		"max_retries": 3,
	}
	payload := map[string]any{}

	removeUnsetFields(desiredState, payload)

	if _, ok := desiredState["max_retries"]; !ok {
		t.Error("expected non-default max_retries (3) to remain even when absent from payload")
	}
}

func TestSortAnyStringSlice_SortsInPlace(t *testing.T) {
	m := map[string]any{
		"scopes": []any{"z-scope", "a-scope", "m-scope"},
	}

	sortAnyStringSlice(m, "scopes")

	got := m["scopes"].([]any)
	want := []any{"a-scope", "m-scope", "z-scope"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sortAnyStringSlice() = %v, want %v", got, want)
	}
}

func TestSortAnyStringSlice_MissingKey(t *testing.T) {
	m := map[string]any{
		"other": "val",
	}

	sortAnyStringSlice(m, "scopes")

	if _, ok := m["scopes"]; ok {
		t.Error("expected key to remain absent")
	}
}

func TestSortAnyStringSlice_SingleElement(t *testing.T) {
	m := map[string]any{
		"scopes": []any{"only-one"},
	}

	sortAnyStringSlice(m, "scopes")

	got := m["scopes"].([]any)
	if len(got) != 1 || got[0] != "only-one" {
		t.Errorf("sortAnyStringSlice() single element = %v", got)
	}
}

func TestSortAnyStringSlice_NonSliceType(t *testing.T) {
	m := map[string]any{
		"scopes": "not-a-slice",
	}

	sortAnyStringSlice(m, "scopes")

	if m["scopes"] != "not-a-slice" {
		t.Error("expected non-slice value to remain unchanged")
	}
}
