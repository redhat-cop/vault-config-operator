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

// IdentityOIDCScopeSpec defines the desired state of IdentityOIDCScope
type IdentityOIDCScopeSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	IdentityOIDCScopeConfig `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type IdentityOIDCScopeConfig struct {

	// Template is the JSON template string for the scope. This may be provided as escaped JSON or base64 encoded JSON.
	// +kubebuilder:validation:Optional
	Template string `json:"template,omitempty"`

	// Description is a description of the scope.
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`
}

// IdentityOIDCScopeStatus defines the observed state of IdentityOIDCScope
type IdentityOIDCScopeStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IdentityOIDCScope is the Schema for the identityoidcscopes API
type IdentityOIDCScope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityOIDCScopeSpec   `json:"spec,omitempty"`
	Status IdentityOIDCScopeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IdentityOIDCScopeList contains a list of IdentityOIDCScope
type IdentityOIDCScopeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityOIDCScope `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IdentityOIDCScope{}, &IdentityOIDCScopeList{})
}

var _ vaultutils.VaultObject = &IdentityOIDCScope{}
var _ vaultutils.ConditionsAware = &IdentityOIDCScope{}

func (m *IdentityOIDCScope) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *IdentityOIDCScope) IsDeletable() bool {
	return true
}

func (m *IdentityOIDCScope) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *IdentityOIDCScope) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *IdentityOIDCScope) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("identity/oidc/scope/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("identity/oidc/scope/" + d.Name)
}

func (d *IdentityOIDCScope) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *IdentityOIDCScopeSpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	if i.Template != "" {
		payload["template"] = i.Template
	}
	if i.Description != "" {
		payload["description"] = i.Description
	}
	return payload
}

func (d *IdentityOIDCScope) IsInitialized() bool {
	return true
}

func (d *IdentityOIDCScope) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *IdentityOIDCScope) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *IdentityOIDCScope) IsValid() (bool, error) {
	return true, nil
}

func (d *IdentityOIDCScope) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *IdentityOIDCScope) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	return reflect.DeepEqual(desiredState, payload)
}
