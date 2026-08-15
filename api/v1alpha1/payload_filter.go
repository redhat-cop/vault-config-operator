package v1alpha1

import "encoding/json"

// filterPayloadToDesiredKeys returns a new map containing only the keys present in desiredState,
// with values taken from payload. This allows reflect.DeepEqual to compare only the fields the
// operator manages, ignoring extra fields Vault adds to its read responses (timestamps, IDs,
// computed metadata, etc.).
func filterPayloadToDesiredKeys(desiredState, payload map[string]any) map[string]any {
	filtered := make(map[string]any, len(desiredState))
	for key := range desiredState {
		if val, exists := payload[key]; exists {
			filtered[key] = val
		}
	}
	return filtered
}

// normalizeVaultReadAliases copies values from known Vault read-response aliases into the
// canonical write-key when the write-key is absent. For example, Vault returns "period"
// on read but expects "token_period" on write — this prevents false drift detection.
func normalizeVaultReadAliases(payload map[string]any, aliases map[string]string) {
	for readKey, writeKey := range aliases {
		if _, hasWriteKey := payload[writeKey]; hasWriteKey {
			continue
		}
		if val, hasReadKey := payload[readKey]; hasReadKey {
			payload[writeKey] = val
		}
	}
}

// removeUnsetFields removes keys from desiredState whose values are zero (empty string,
// empty slice, or default numeric sentinel) and that are also absent from the Vault payload.
// This prevents false drift when the operator includes all managed fields with zero defaults
// but Vault omits fields that were never set.
func removeUnsetFields(desiredState, payload map[string]any) {
	for key, val := range desiredState {
		if _, inPayload := payload[key]; inPayload {
			continue
		}
		switch v := val.(type) {
		case string:
			if v == "" {
				delete(desiredState, key)
			}
		case json.Number:
			if v == "0" {
				delete(desiredState, key)
			}
		case []any:
			if len(v) == 0 {
				delete(desiredState, key)
			}
		case int:
			if v == -1 {
				delete(desiredState, key)
			}
		}
	}
}
