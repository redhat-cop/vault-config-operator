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
	"context"
	"reflect"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TransitSecretEngineKeySpec defines the desired state of TransitSecretEngineKey
type TransitSecretEngineKeySpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the k8s auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/keys/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path"`

	TransitKeyConfig `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// TransitKeyConfig defines the configuration fields for a Transit encryption key.
// Fields are split into create-time (immutable after key creation) and config-time (mutable via /keys/{name}/config).
type TransitKeyConfig struct {

	// Type specifies the key algorithm. Immutable after key creation.
	// +kubebuilder:default="aes256-gcm96"
	// +kubebuilder:validation:Enum={"aes128-gcm96","aes256-gcm96","chacha20-poly1305","ed25519","ecdsa-p256","ecdsa-p384","ecdsa-p521","rsa-2048","rsa-3072","rsa-4096","hmac"}
	Type string `json:"type"`

	// Derived enables key derivation. Immutable after key creation.
	// +kubebuilder:validation:Optional
	Derived bool `json:"derived,omitempty"`

	// ConvergentEncryption enables convergent encryption mode. Requires derived=true. Immutable after key creation.
	// +kubebuilder:validation:Optional
	ConvergentEncryption bool `json:"convergentEncryption,omitempty"`

	// KeySize is the key size in bytes. Only applicable to HMAC keys (32-512 bytes). Immutable after key creation.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=512
	KeySize int `json:"keySize,omitempty"`

	// MinDecryptionVersion specifies the minimum version of ciphertext allowed to be decrypted.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	MinDecryptionVersion int `json:"minDecryptionVersion,omitempty"`

	// MinEncryptionVersion specifies the minimum version of the key that can be used to encrypt plaintext.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	MinEncryptionVersion int `json:"minEncryptionVersion,omitempty"`

	// DeletionAllowed controls whether the key is allowed to be deleted.
	// +kubebuilder:validation:Optional
	DeletionAllowed bool `json:"deletionAllowed,omitempty"`

	// Exportable controls whether the key is exportable. One-way: can be set to true but never unset.
	// +kubebuilder:validation:Optional
	Exportable bool `json:"exportable,omitempty"`

	// AllowPlaintextBackup controls whether plaintext backup of the key is allowed. One-way: can be set to true but never unset.
	// +kubebuilder:validation:Optional
	AllowPlaintextBackup bool `json:"allowPlaintextBackup,omitempty"`

	// AutoRotatePeriod specifies the period at which the key is automatically rotated. "0" disables auto-rotation. Minimum 1h when enabled.
	// +kubebuilder:validation:Optional
	AutoRotatePeriod string `json:"autoRotatePeriod,omitempty"`
}

// TransitSecretEngineKeyStatus defines the observed state of TransitSecretEngineKey
type TransitSecretEngineKeyStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TransitSecretEngineKey is the Schema for the transitsecretenginekeys API
type TransitSecretEngineKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitSecretEngineKeySpec   `json:"spec,omitempty"`
	Status TransitSecretEngineKeyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TransitSecretEngineKeyList contains a list of TransitSecretEngineKey
type TransitSecretEngineKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitSecretEngineKey `json:"items"`
}

var _ vaultutils.ConditionsAware = &TransitSecretEngineKey{}

func (d *TransitSecretEngineKey) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (m *TransitSecretEngineKey) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *TransitSecretEngineKey) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *TransitSecretEngineKey) IsDeletable() bool {
	return true
}

func init() {
	SchemeBuilder.Register(&TransitSecretEngineKey{}, &TransitSecretEngineKeyList{})
}

var _ vaultutils.VaultObject = &TransitSecretEngineKey{}

func (d *TransitSecretEngineKey) GetPath() string {
	name := d.Name
	if d.Spec.Name != "" {
		name = d.Spec.Name
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "keys" + "/" + name)
}

func (d *TransitSecretEngineKey) GetConfigPath() string {
	return d.GetPath() + "/config"
}

func (d *TransitSecretEngineKey) GetPayload() map[string]any {
	return d.Spec.TransitKeyConfig.toMap()
}

func (d *TransitSecretEngineKey) GetConfigPayload() map[string]any {
	return d.Spec.TransitKeyConfig.configToMap()
}

func (d *TransitSecretEngineKey) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.TransitKeyConfig.configToMap()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *TransitSecretEngineKey) IsInitialized() bool {
	return true
}

func (d *TransitSecretEngineKey) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *TransitSecretEngineKey) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *TransitSecretEngineKey) IsValid() (bool, error) {
	return true, nil
}

func (d *TransitSecretEngineKey) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

// toMap returns the create-time payload for POST to {path}/keys/{name}
func (c *TransitKeyConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["type"] = c.Type
	payload["derived"] = c.Derived
	payload["convergent_encryption"] = c.ConvergentEncryption
	payload["exportable"] = c.Exportable
	payload["allow_plaintext_backup"] = c.AllowPlaintextBackup
	payload["key_size"] = c.KeySize
	payload["auto_rotate_period"] = c.autoRotatePeriodOrDefault()
	return payload
}

// configToMap returns the config-time payload for POST to {path}/keys/{name}/config.
// Numeric fields use float64 to match what Vault's JSON API returns (encoding/json
// unmarshals numbers as float64 into any), avoiding false drift detection.
func (c *TransitKeyConfig) configToMap() map[string]any {
	payload := map[string]any{}
	payload["min_decryption_version"] = float64(c.MinDecryptionVersion)
	payload["min_encryption_version"] = float64(c.MinEncryptionVersion)
	payload["deletion_allowed"] = c.DeletionAllowed
	payload["exportable"] = c.Exportable
	payload["allow_plaintext_backup"] = c.AllowPlaintextBackup
	payload["auto_rotate_period"] = c.autoRotatePeriodOrDefault()
	return payload
}

// autoRotatePeriodOrDefault returns "0" (Vault's disabled value) when the field
// is empty/unset, preventing false drift between our "" and Vault's "0".
func (c *TransitKeyConfig) autoRotatePeriodOrDefault() string {
	if c.AutoRotatePeriod == "" {
		return "0"
	}
	return c.AutoRotatePeriod
}
