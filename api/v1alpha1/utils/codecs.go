/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"encoding/json"
	"strconv"
	"time"
)

// DurationToSeconds converts a Go duration string (e.g. "4h", "30s") to a
// json.Number representing integer seconds, matching the Vault API response
// format where durations are returned as numeric seconds.
func DurationToSeconds(s string) json.Number {
	if s == "" {
		return json.Number("0")
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return json.Number(s)
	}
	return json.Number(strconv.FormatInt(int64(d.Seconds()), 10))
}

// ToAnyMapJsonNumber converts map[string]int to map[string]any with json.Number
// values, matching the Vault API response where the Go client decodes JSON
// numbers via UseNumber(). Returns an empty map (not nil) for nil input,
// because Vault returns empty JSON objects {} for unset map fields.
func ToAnyMapJsonNumber(m map[string]int) map[string]any {
	if len(m) == 0 {
		return map[string]any{}
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = json.Number(strconv.Itoa(v))
	}
	return result
}

// ToAnyMapString converts map[string]string to map[string]any,
// matching the representation Vault returns after JSON deserialization.
// Returns an empty map (not nil) for nil input, because Vault returns
// empty JSON objects {} for unset map fields.
func ToAnyMapString(m map[string]string) map[string]any {
	if len(m) == 0 {
		return map[string]any{}
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// ToInterfaceArray converts []string to []any, matching the Vault API
// response format where JSON arrays are deserialized as []any.
func ToInterfaceArray(values []string) []any {
	result := []any{}
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

// ResourceName returns specName if non-empty, otherwise metaName.
// This consolidates the common pattern in GetPath() methods:
//
//	if d.Spec.Name != "" { return d.Spec.Name } else { return d.Name }
func ResourceName(specName, metaName string) string {
	if specName != "" {
		return specName
	}
	return metaName
}
