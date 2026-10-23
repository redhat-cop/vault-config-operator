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

// KeyRotationConfigSpec defines the desired state of KeyRotationConfig
type KeyRotationConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	KeyRotationConfigConfig `json:",inline"`
}

// KeyRotationConfigConfig contains the key rotation configuration fields written to sys/rotate/config.
type KeyRotationConfigConfig struct {

	// MaxOperations is the number of operations performed before the barrier key is automatically rotated.
	// +kubebuilder:validation:Optional
	MaxOperations int64 `json:"maxOperations,omitempty" vault:"max_operations,number,omitempty"`

	// Interval is the duration after which the barrier key is automatically rotated (e.g., "4320h").
	// Must be at least 24h if specified.
	// +kubebuilder:validation:Optional
	Interval string `json:"interval,omitempty" vault:"interval,omitempty"`

	// Enabled controls whether automatic rotation is active. Defaults to true.
	// +kubebuilder:default=true
	Enabled bool `json:"enabled" vault:"enabled"`
}

func (i *KeyRotationConfigConfig) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(i)
}

// KeyRotationConfigStatus defines the observed state of KeyRotationConfig
type KeyRotationConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KeyRotationConfig is the Schema for the keyrotationconfigs API
type KeyRotationConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeyRotationConfigSpec   `json:"spec,omitempty"`
	Status KeyRotationConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KeyRotationConfigList contains a list of KeyRotationConfig
type KeyRotationConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeyRotationConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeyRotationConfig{}, &KeyRotationConfigList{})
}

var _ vaultutils.VaultObject = &KeyRotationConfig{}
var _ vaultutils.ConditionsAware = &KeyRotationConfig{}

func (m *KeyRotationConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *KeyRotationConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *KeyRotationConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *KeyRotationConfig) GetPath() string {
	return "sys/rotate/config"
}

func (d *KeyRotationConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *KeyRotationConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.KeyRotationConfigConfig.toMap()
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *KeyRotationConfig) IsDeletable() bool {
	return false
}

func (d *KeyRotationConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
