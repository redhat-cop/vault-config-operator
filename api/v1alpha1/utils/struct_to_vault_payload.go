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
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// StructToVaultPayload converts a struct to a map[string]any suitable for
// writing to Vault, driven by `vault:` struct tags.
//
// Tag grammar:
//
//	vault:"<key>[,<codec>][,omitempty][,when=Field:value][,default=X][,spreadmap]"
//	vault:"-"          // exclude field
//	vault:",inline"    // merge embedded struct fields into parent
//
// Codecs: identity (default), duration, number, intmap, stringmap.
func StructToVaultPayload(v any) map[string]any {
	payload := map[string]any{}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return payload
	}
	walkStruct(val, payload)
	return payload
}

func walkStruct(structVal reflect.Value, payload map[string]any) {
	structType := structVal.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := structVal.Field(i)

		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("vault")
		if tag == "" {
			continue
		}
		if tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		key := parts[0]
		opts := parseTagOptions(parts[1:])

		if opts.inline {
			if fieldVal.Kind() == reflect.Struct {
				walkStruct(fieldVal, payload)
			} else if fieldVal.Kind() == reflect.Pointer && !fieldVal.IsNil() && fieldVal.Elem().Kind() == reflect.Struct {
				walkStruct(fieldVal.Elem(), payload)
			}
			continue
		}

		if opts.when != "" {
			whenField, whenValue := parseWhen(opts.when)
			siblingVal := structVal.FieldByName(whenField)
			if !siblingVal.IsValid() || fmt.Sprintf("%v", siblingVal.Interface()) != whenValue {
				continue
			}
		}

		if opts.spreadmap {
			spreadMap(fieldVal, payload)
			continue
		}

		// Handle pointer types with default
		if fieldVal.Kind() == reflect.Pointer {
			if fieldVal.IsNil() {
				if opts.defaultVal != "" {
					payload[key] = parseDefault(opts.defaultVal, field.Type.Elem())
				} else if opts.omitempty {
					continue
				}
				continue
			}
			fieldVal = fieldVal.Elem()
		} else if opts.omitempty && isZero(fieldVal) {
			continue
		}

		payload[key] = encodeValue(fieldVal, opts.codec)
	}
}

type tagOptions struct {
	codec      string
	omitempty  bool
	when       string
	defaultVal string
	spreadmap  bool
	inline     bool
}

func parseTagOptions(parts []string) tagOptions {
	opts := tagOptions{}
	for _, p := range parts {
		switch {
		case p == "omitempty":
			opts.omitempty = true
		case p == "spreadmap":
			opts.spreadmap = true
		case p == "inline":
			opts.inline = true
		case strings.HasPrefix(p, "when="):
			opts.when = strings.TrimPrefix(p, "when=")
		case strings.HasPrefix(p, "default="):
			opts.defaultVal = strings.TrimPrefix(p, "default=")
		case p == "duration" || p == "number" || p == "intmap" || p == "stringmap":
			opts.codec = p
		}
	}
	return opts
}

func parseWhen(s string) (field, value string) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return s, ""
}

func parseDefault(val string, elemType reflect.Type) any {
	switch elemType.Kind() {
	case reflect.Bool:
		b, _ := strconv.ParseBool(val)
		return b
	case reflect.Int, reflect.Int64:
		n, _ := strconv.ParseInt(val, 10, 64)
		return json.Number(strconv.FormatInt(n, 10))
	default:
		return val
	}
}

func spreadMap(fieldVal reflect.Value, payload map[string]any) {
	if fieldVal.Kind() == reflect.Map && !fieldVal.IsNil() {
		for _, k := range fieldVal.MapKeys() {
			payload[fmt.Sprintf("%v", k.Interface())] = fieldVal.MapIndex(k).Interface()
		}
	}
}

func encodeValue(fieldVal reflect.Value, codec string) any {
	switch codec {
	case "duration":
		return DurationToSeconds(fmt.Sprintf("%v", fieldVal.Interface()))
	case "number":
		return toJsonNumber(fieldVal)
	case "intmap":
		return encodeIntMap(fieldVal)
	case "stringmap":
		return encodeStringMap(fieldVal)
	default:
		return encodeIdentity(fieldVal)
	}
}

func toJsonNumber(v reflect.Value) json.Number {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return json.Number(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return json.Number(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		return json.Number(strconv.FormatFloat(v.Float(), 'f', -1, 64))
	default:
		return json.Number(fmt.Sprintf("%v", v.Interface()))
	}
}

func encodeIntMap(v reflect.Value) map[string]any {
	if v.Kind() != reflect.Map || v.IsNil() {
		return map[string]any{}
	}
	result := make(map[string]any, v.Len())
	for _, k := range v.MapKeys() {
		result[fmt.Sprintf("%v", k.Interface())] = json.Number(strconv.FormatInt(v.MapIndex(k).Int(), 10))
	}
	if len(result) == 0 {
		return map[string]any{}
	}
	return result
}

func encodeStringMap(v reflect.Value) map[string]any {
	if v.Kind() != reflect.Map || v.IsNil() {
		return map[string]any{}
	}
	result := make(map[string]any, v.Len())
	for _, k := range v.MapKeys() {
		result[fmt.Sprintf("%v", k.Interface())] = fmt.Sprintf("%v", v.MapIndex(k).Interface())
	}
	if len(result) == 0 {
		return map[string]any{}
	}
	return result
}

func encodeIdentity(v reflect.Value) any {
	if v.Kind() == reflect.Slice {
		if v.Type().Elem().Kind() == reflect.String {
			result := make([]any, v.Len())
			for i := 0; i < v.Len(); i++ {
				result[i] = v.Index(i).Interface()
			}
			return result
		}
	}
	return v.Interface()
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	default:
		return false
	}
}
