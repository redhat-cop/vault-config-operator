package v1alpha1

// filterPayloadToDesiredKeys returns a new map containing only the keys present in desiredState,
// with values taken from payload. This allows reflect.DeepEqual to compare only the fields the
// operator manages, ignoring extra fields Vault adds to its read responses (timestamps, IDs,
// computed metadata, etc.).
func filterPayloadToDesiredKeys(desiredState, payload map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{}, len(desiredState))
	for key := range desiredState {
		if val, exists := payload[key]; exists {
			filtered[key] = val
		}
	}
	return filtered
}
