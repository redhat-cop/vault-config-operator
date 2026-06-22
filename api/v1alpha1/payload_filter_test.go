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
