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

// IdentityTokenRoleSpec defines the desired state of IdentityTokenRole
type IdentityTokenRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	IdentityTokenRoleConfig `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type IdentityTokenRoleConfig struct {

	// Key is a configured named key, the key must already exist.
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// Template is the template string to use for generating tokens.
	// This may be in string-ified JSON or base64 format.
	// +kubebuilder:validation:Optional
	Template string `json:"template,omitempty"`

	// ClientID is an optional client ID. A random ID will be generated if left unset.
	// +kubebuilder:validation:Optional
	ClientID string `json:"clientID,omitempty"`

	// TTL is the TTL of the tokens generated against the role.
	// Uses duration format strings.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="24h"
	TTL string `json:"ttl,omitempty"`
}

// IdentityTokenRoleStatus defines the observed state of IdentityTokenRole
type IdentityTokenRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IdentityTokenRole is the Schema for the identitytokenroles API
type IdentityTokenRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityTokenRoleSpec   `json:"spec,omitempty"`
	Status IdentityTokenRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IdentityTokenRoleList contains a list of IdentityTokenRole
type IdentityTokenRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityTokenRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IdentityTokenRole{}, &IdentityTokenRoleList{})
}

var _ vaultutils.VaultObject = &IdentityTokenRole{}
var _ vaultutils.ConditionsAware = &IdentityTokenRole{}

func (m *IdentityTokenRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *IdentityTokenRole) IsDeletable() bool {
	return true
}

func (m *IdentityTokenRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *IdentityTokenRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *IdentityTokenRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("identity/oidc/role/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("identity/oidc/role/" + d.Name)
}

func (d *IdentityTokenRole) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *IdentityTokenRoleSpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["key"] = i.Key
	if i.Template != "" {
		payload["template"] = i.Template
	}
	if i.ClientID != "" {
		payload["client_id"] = i.ClientID
	}
	payload["ttl"] = i.TTL
	return payload
}

func (d *IdentityTokenRole) IsInitialized() bool {
	return true
}

func (d *IdentityTokenRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *IdentityTokenRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *IdentityTokenRole) IsValid() (bool, error) {
	return true, nil
}

func (d *IdentityTokenRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *IdentityTokenRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	return reflect.DeepEqual(desiredState, payload)
}
