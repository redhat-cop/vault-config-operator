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

package v1alpha1

import (
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CORSConfigSpec defines the desired state of CORSConfig
type CORSConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	CORSConfigConfig `json:",inline"`
}

// CORSConfigConfig contains the CORS configuration fields written to sys/config/cors.
type CORSConfigConfig struct {

	// AllowedOrigins is the list of origins permitted to make cross-origin requests.
	// Use ["*"] to allow all origins.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	AllowedOrigins []string `json:"allowedOrigins" vault:"allowed_origins"`

	// AllowedHeaders is the list of additional headers permitted in cross-origin requests.
	// Vault always includes its own default headers; specifying extras here is additive.
	// +kubebuilder:validation:Optional
	AllowedHeaders []string `json:"allowedHeaders,omitempty" vault:"allowed_headers,omitempty"`
}

func (i *CORSConfigConfig) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(i)
}

// CORSConfigStatus defines the observed state of CORSConfig
type CORSConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CORSConfig is the Schema for the corsconfigs API
type CORSConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CORSConfigSpec   `json:"spec,omitempty"`
	Status CORSConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CORSConfigList contains a list of CORSConfig
type CORSConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CORSConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CORSConfig{}, &CORSConfigList{})
}

var _ vaultutils.VaultObject = &CORSConfig{}
var _ vaultutils.ConditionsAware = &CORSConfig{}

func (m *CORSConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *CORSConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *CORSConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *CORSConfig) GetPath() string {
	return "sys/config/cors"
}

func (d *CORSConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *CORSConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.CORSConfigConfig.toMap()

	// For allowed_origins: order-insensitive set comparison
	desiredOrigins := toStringSet(desiredState["allowed_origins"])
	payloadOrigins := toStringSet(payload["allowed_origins"])
	if !setsEqual(desiredOrigins, payloadOrigins) {
		return false
	}

	// For allowed_headers: subset check — Vault appends its own default headers,
	// so we only require that all desired headers appear in the payload.
	if desiredHeaders, ok := desiredState["allowed_headers"]; ok {
		desiredSet := toStringSet(desiredHeaders)
		payloadSet := toStringSet(payload["allowed_headers"])
		if !isSubset(desiredSet, payloadSet) {
			return false
		}
	}

	return true
}

func (d *CORSConfig) IsDeletable() bool {
	return true
}

func (d *CORSConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

// toStringSet converts an interface value (expected to be []any of strings) to a map[string]bool set.
func toStringSet(v any) map[string]bool {
	set := map[string]bool{}
	if v == nil {
		return set
	}
	switch s := v.(type) {
	case []any:
		for _, item := range s {
			if str, ok := item.(string); ok {
				set[str] = true
			}
		}
	case []string:
		for _, str := range s {
			set[str] = true
		}
	}
	return set
}

// setsEqual returns true if a and b contain the same elements.
func setsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

// isSubset returns true if every element in sub is also in super.
func isSubset(sub, super map[string]bool) bool {
	for k := range sub {
		if !super[k] {
			return false
		}
	}
	return true
}
