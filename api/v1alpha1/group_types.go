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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GroupSpec defines the desired state of Group
type GroupSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	GroupConfig `json:",inline"`
}

type GroupConfig struct {

	// Type Type of the group, internal or external. Defaults to internal
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"internal","external"}
	// +kubebuilder:default:="internal"
	Type string `json:"type,omitempty"`

	// Metadata Metadata to be associated with the group.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Metadata map[string]string `json:"metadata,omitempty"`

	// Policies Policies to be tied to the group.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	Policies []string `json:"policies,omitempty"`

	// MemberGroupIDs Group IDs to be assigned as group members.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	MemberGroupIDs []string `json:"memberGroupIDs,omitempty"`

	// MemberEntityIDs Entity IDs to be assigned as group members.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	MemberEntityIDs []string `json:"memberEntityIDs,omitempty"`
}

// GroupStatus defines the observed state of Group
type GroupStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Group is the Schema for the groups API
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec,omitempty"`
	Status GroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GroupList contains a list of Group
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
}

var _ vaultutils.VaultObject = &Group{}
var _ vaultutils.ConditionsAware = &Group{}

func (m *Group) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *Group) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *Group) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *Group) GetPath() string {
	return string("/identity/group/name/" + d.Name)
}

func (d *Group) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *GroupSpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["type"] = i.Type
	payload["metadata"] = i.Metadata
	payload["policies"] = i.Policies
	if i.Type == "internal" {
		payload["member_group_ids"] = i.MemberEntityIDs
		payload["member_entity_ids"] = i.MemberEntityIDs
	}
	return payload
}

func (d *Group) IsInitialized() bool {
	return true
}

func (d *Group) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *Group) IsValid() (bool, error) {
	return true, nil
}

func (d *Group) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *Group) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	delete(payload, "name")
	return reflect.DeepEqual(desiredState, payload)
}
