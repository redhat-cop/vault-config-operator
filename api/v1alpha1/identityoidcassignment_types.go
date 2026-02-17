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

// IdentityOIDCAssignmentSpec defines the desired state of IdentityOIDCAssignment
type IdentityOIDCAssignmentSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	IdentityOIDCAssignmentConfig `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type IdentityOIDCAssignmentConfig struct {

	// EntityIDs is a list of Vault entity IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	EntityIDs []string `json:"entityIDs,omitempty"`

	// GroupIDs is a list of Vault group IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	GroupIDs []string `json:"groupIDs,omitempty"`
}

// IdentityOIDCAssignmentStatus defines the observed state of IdentityOIDCAssignment
type IdentityOIDCAssignmentStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IdentityOIDCAssignment is the Schema for the identityoidcassignments API
type IdentityOIDCAssignment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityOIDCAssignmentSpec   `json:"spec,omitempty"`
	Status IdentityOIDCAssignmentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IdentityOIDCAssignmentList contains a list of IdentityOIDCAssignment
type IdentityOIDCAssignmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityOIDCAssignment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IdentityOIDCAssignment{}, &IdentityOIDCAssignmentList{})
}

var _ vaultutils.VaultObject = &IdentityOIDCAssignment{}
var _ vaultutils.ConditionsAware = &IdentityOIDCAssignment{}

func (m *IdentityOIDCAssignment) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *IdentityOIDCAssignment) IsDeletable() bool {
	return true
}

func (m *IdentityOIDCAssignment) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *IdentityOIDCAssignment) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *IdentityOIDCAssignment) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("identity/oidc/assignment/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("identity/oidc/assignment/" + d.Name)
}

func (d *IdentityOIDCAssignment) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *IdentityOIDCAssignmentSpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["entity_ids"] = i.EntityIDs
	payload["group_ids"] = i.GroupIDs
	return payload
}

func (d *IdentityOIDCAssignment) IsInitialized() bool {
	return true
}

func (d *IdentityOIDCAssignment) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *IdentityOIDCAssignment) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *IdentityOIDCAssignment) IsValid() (bool, error) {
	return true, nil
}

func (d *IdentityOIDCAssignment) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *IdentityOIDCAssignment) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	return reflect.DeepEqual(desiredState, payload)
}
