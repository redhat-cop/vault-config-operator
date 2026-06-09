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

var _ vaultutils.VaultObject = &Namespace{}
var _ vaultutils.ConditionsAware = &Namespace{}

func (d *Namespace) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *Namespace) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("sys/namespaces/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("sys/namespaces/" + d.Name)
}

func (d *Namespace) GetPayload() map[string]interface{} {
	return d.toMap()
}

func (d *Namespace) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.GetPayload()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *Namespace) IsInitialized() bool {
	return true
}

func (d *Namespace) IsDeletable() bool {
	return true
}

func (d *Namespace) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *Namespace) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *Namespace) IsValid() (bool, error) {
	return true, nil
}

// NamespaceSpec defines the desired state of Namespace
type NamespaceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`

	// Path at which to create the namespace (in case of nested namespaces).
	// The final path in Vault will be {spec.path}/{spec.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Path vaultutils.Path `json:"path,omitempty"`
}

// NamespaceStatus defines the observed state of Namespace
type NamespaceStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *Namespace) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *Namespace) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Namespace is the Schema for the policies API
type Namespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceSpec   `json:"spec,omitempty"`
	Status NamespaceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NamespaceList contains a list of Namespace
type NamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Namespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Namespace{}, &NamespaceList{})
}

func (d *Namespace) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (i *Namespace) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["name"] = i.Spec.Name
	payload["path"] = i.Spec.Path

	return payload
}
