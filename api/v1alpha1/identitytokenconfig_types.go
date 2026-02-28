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

// IdentityTokenConfigSpec defines the desired state of IdentityTokenConfig
type IdentityTokenConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	IdentityTokenConfigConfig `json:",inline"`
}

type IdentityTokenConfigConfig struct {

	// Issuer is the issuer URL to be used in the iss claim of the token.
	// If not set, Vault's api_addr will be used. The issuer is a case sensitive URL
	// using the https scheme that contains scheme, host, and an optional port number.
	// +kubebuilder:validation:Optional
	Issuer string `json:"issuer,omitempty"`
}

// IdentityTokenConfigStatus defines the observed state of IdentityTokenConfig
type IdentityTokenConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IdentityTokenConfig is the Schema for the identitytokenconfigs API
type IdentityTokenConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityTokenConfigSpec   `json:"spec,omitempty"`
	Status IdentityTokenConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IdentityTokenConfigList contains a list of IdentityTokenConfig
type IdentityTokenConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityTokenConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IdentityTokenConfig{}, &IdentityTokenConfigList{})
}

var _ vaultutils.VaultObject = &IdentityTokenConfig{}
var _ vaultutils.ConditionsAware = &IdentityTokenConfig{}

func (m *IdentityTokenConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *IdentityTokenConfig) IsDeletable() bool {
	return false
}

func (m *IdentityTokenConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *IdentityTokenConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *IdentityTokenConfig) GetPath() string {
	return vaultutils.CleansePath("identity/oidc/config")
}

func (d *IdentityTokenConfig) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *IdentityTokenConfigSpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["issuer"] = i.Issuer
	return payload
}

func (d *IdentityTokenConfig) IsInitialized() bool {
	return true
}

func (d *IdentityTokenConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *IdentityTokenConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *IdentityTokenConfig) IsValid() (bool, error) {
	return true, nil
}

func (d *IdentityTokenConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *IdentityTokenConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	return reflect.DeepEqual(desiredState, payload)
}
