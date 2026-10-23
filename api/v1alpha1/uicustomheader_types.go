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
	"reflect"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ vaultutils.VaultObject = &UICustomHeader{}
var _ vaultutils.ConditionsAware = &UICustomHeader{}

func (d *UICustomHeader) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *UICustomHeader) GetPath() string {
	return vaultutils.CleansePath("sys/config/ui/headers/" + vaultutils.ResourceName(d.Spec.Name, d.Name))
}

func (d *UICustomHeader) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *UICustomHeader) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredValues := d.Spec.Values

	// Case 1: Vault returned multivalue format (if API evolves or proxy adds param)
	if vaultValues, ok := payload["values"]; ok {
		return reflect.DeepEqual(vaultutils.ToInterfaceArray(desiredValues), vaultValues)
	}

	// Case 2: Vault returned single-value format (standard GET behavior)
	if vaultValue, ok := payload["value"]; ok {
		if len(desiredValues) == 1 {
			return desiredValues[0] == vaultValue
		}
		// Multi-value desired but only single-value readable — always re-write.
		// Vault POST is idempotent so this is safe.
		return false
	}

	// No value key at all — empty response means not configured
	return len(desiredValues) == 0
}

func (d *UICustomHeader) IsDeletable() bool {
	return true
}

// UICustomHeaderSpec defines the desired state of UICustomHeader
type UICustomHeaderSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	UICustomHeaderConfig `json:",inline"`

	// The name of the custom header in Vault. If specified, takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?`
	Name string `json:"name,omitempty"`
}

// UICustomHeaderConfig contains the Vault UI custom header configuration fields.
type UICustomHeaderConfig struct {
	// Values is the list of values to return for this custom header.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Values []string `json:"values" vault:"values"`
}

func (c *UICustomHeaderConfig) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(c)
}

// UICustomHeaderStatus defines the observed state of UICustomHeader
type UICustomHeaderStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *UICustomHeader) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *UICustomHeader) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UICustomHeader is the Schema for the uicustomheaders API
type UICustomHeader struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UICustomHeaderSpec   `json:"spec,omitempty"`
	Status UICustomHeaderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UICustomHeaderList contains a list of UICustomHeader
type UICustomHeaderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UICustomHeader `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UICustomHeader{}, &UICustomHeaderList{})
}

func (d *UICustomHeader) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
