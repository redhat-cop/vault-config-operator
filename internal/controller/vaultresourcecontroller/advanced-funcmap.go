/*
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

/*
Concept pulled from Helm to match a known templating pattern
https://github.com/helm/helm/blob/master/pkg/engine/funcs.go
*/

package vaultresourcecontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/sprig/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

// AdvancedTemplateFuncMap to add Sprig and additional templating functions
func AdvancedTemplateFuncMap(config *rest.Config, logger logr.Logger) template.FuncMap {
	f := sprig.HermeticTxtFuncMap()
	// Removed these functions from the core Sprig package for security concerns
	delete(f, "env")
	delete(f, "expandenv")

	extra := template.FuncMap{
		"toToml":        toTOML,
		"toYaml":        toYAML,
		"fromYaml":      fromYAML,
		"fromYamlArray": fromYAMLArray,
		"toJson":        toJSON,
		"fromJson":      fromJSON,
		"fromJsonArray": fromJSONArray,

		// A variety of known templating functions that have not been implemented yet
		"include":  func(string, interface{}) string { return "not implemented" },
		"tpl":      func(string, interface{}) interface{} { return "not implemented" },
		"required": func(string, interface{}) (interface{}, error) { return "not implemented", nil },
	}

	for k, v := range extra {
		f[k] = v
	}

	// Adding additional functionality found in Helm
	f["lookup"] = NewLookupFunction(config, logger)

	// Add the `required` function here so we can use lintMode
	f["required"] = func(warn string, val interface{}) (interface{}, error) {
		if val == nil {
			return val, errors.Errorf(warn)
		} else if _, ok := val.(string); ok {
			if val == "" {
				return val, errors.Errorf(warn)
			}
		}
		return val, nil
	}

	return f
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// fromYAML converts a YAML document into a map[string]interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromYAML(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// fromYAMLArray converts a YAML array into a []interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string as
// the first and only item in the returned array.
func fromYAMLArray(str string) []interface{} {
	a := []interface{}{}

	if err := yaml.Unmarshal([]byte(str), &a); err != nil {
		a = []interface{}{err.Error()}
	}
	return a
}

// toTOML takes an interface, marshals it to toml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toTOML(v interface{}) string {
	b := bytes.NewBuffer(nil)
	e := toml.NewEncoder(b)
	err := e.Encode(v)
	if err != nil {
		return err.Error()
	}
	return b.String()
}

// toJSON takes an interface, marshals it to json, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

// fromJSON converts a JSON document into a map[string]interface{}.
//
// This is not a general-purpose JSON parser, and will not parse all valid
// JSON documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromJSON(str string) map[string]interface{} {
	m := make(map[string]interface{})

	if err := json.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// fromJSONArray converts a JSON array into a []interface{}.
//
// This is not a general-purpose JSON parser, and will not parse all valid
// JSON documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string as
// the first and only item in the returned array.
func fromJSONArray(str string) []interface{} {
	a := []interface{}{}

	if err := json.Unmarshal([]byte(str), &a); err != nil {
		a = []interface{}{err.Error()}
	}
	return a
}

type lookupFunc = func(apiversion string, resource string, namespace string, name string) (map[string]interface{}, error)

// NewLookupFunction get information at runtime from cluster
func NewLookupFunction(config *rest.Config, logger logr.Logger) lookupFunc {
	return func(apiversion string, resource string, namespace string, name string) (map[string]interface{}, error) {
		var client dynamic.ResourceInterface
		ctx := context.TODO()
		ctx = context.WithValue(ctx, "restConfig", config)
		ctx = log.IntoContext(ctx, logger.WithName("lookup function"))
		c, namespaced, err := GetDynamicClientForGVK(ctx, schema.FromAPIVersionAndKind(apiversion, resource))
		if err != nil {
			return map[string]interface{}{}, err
		}
		if namespaced && namespace != "" {
			client = c.Namespace(namespace)
		} else {
			client = c
		}
		if name != "" {
			// this will return a single object
			obj, err := client.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					// Just return an empty interface when the object was not found.
					// That way, users can use `if not (lookup ...)` in their templates.
					return map[string]interface{}{}, nil
				}
				return map[string]interface{}{}, err
			}
			return obj.UnstructuredContent(), nil
		}
		//this will return a list
		obj, err := client.List(ctx, metav1.ListOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Just return an empty interface when the object was not found.
				// That way, users can use `if not (lookup ...)` in their templates.
				return map[string]interface{}{}, nil
			}
			return map[string]interface{}{}, err
		}
		return obj.UnstructuredContent(), nil
	}
}
