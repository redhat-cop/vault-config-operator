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

// EntitySpec defines the desired state of Entity
type EntitySpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	EntityConfig `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type EntityConfig struct {

	// Metadata Metadata to be associated with the entity.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Metadata map[string]string `json:"metadata,omitempty"`

	// Policies Policies to be tied to the entity.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	Policies []string `json:"policies,omitempty"`

	// Disabled Whether the entity is disabled. Disabled entities' associated tokens cannot be used, but are not revoked.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	Disabled bool `json:"disabled,omitempty"`
}

// EntityStatus defines the observed state of Entity
type EntityStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Entity is the Schema for the entities API
type Entity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EntitySpec   `json:"spec,omitempty"`
	Status EntityStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EntityList contains a list of Entity
type EntityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Entity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Entity{}, &EntityList{})
}

var _ vaultutils.VaultObject = &Entity{}
var _ vaultutils.ConditionsAware = &Entity{}

func (m *Entity) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *Entity) IsDeletable() bool {
	return true
}

func (m *Entity) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *Entity) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *Entity) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string("/identity/entity/name/" + d.Spec.Name))
	}
	return vaultutils.CleansePath(string("/identity/entity/name/" + d.Name))
}

func (d *Entity) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *EntitySpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["metadata"] = i.Metadata
	payload["policies"] = i.Policies
	payload["disabled"] = i.Disabled
	return payload
}

func (d *Entity) IsInitialized() bool {
	return true
}

func (d *Entity) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *Entity) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *Entity) IsValid() (bool, error) {
	return true, nil
}

func (d *Entity) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *Entity) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	delete(payload, "name")
	delete(payload, "id")
	delete(payload, "aliases")
	delete(payload, "creation_time")
	delete(payload, "last_update_time")
	delete(payload, "merged_entity_ids")
	delete(payload, "direct_group_ids")
	delete(payload, "group_ids")
	delete(payload, "inherited_group_ids")
	delete(payload, "namespace_id")
	delete(payload, "bucket_key_hash")
	return reflect.DeepEqual(desiredState, payload)
}
