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

// LDAPAuthEngineGroupSpec defines the desired state of LDAPAuthEngineGroup
type LDAPAuthEngineGroupSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path will be {[spec.authentication.namespace]}/auth/{spec.path}/groups/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// The name of the LDAP group
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Comma-separated list of policies associated to the group
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	Policies string `json:"policies,omitempty"`
}

var _ vaultutils.VaultObject = &LDAPAuthEngineGroup{}

func (d *LDAPAuthEngineGroup) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *LDAPAuthEngineGroup) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/groups/" + string(d.Spec.Name))
}

func (d *LDAPAuthEngineGroup) GetPayload() map[string]interface{} {
	return d.toMap()
}

func (d *LDAPAuthEngineGroup) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	return reflect.DeepEqual(d.GetPayload(), payload)
}

func (d *LDAPAuthEngineGroup) IsInitialized() bool {
	return true
}

func (d *LDAPAuthEngineGroup) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *LDAPAuthEngineGroup) IsValid() (bool, error) {
	return true, nil
}

// LDAPAuthEngineGroupStatus defines the observed state of LDAPAuthEngineGroup
type LDAPAuthEngineGroupStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LDAPAuthEngineGroup is the Schema for the ldapauthenginegroups API
type LDAPAuthEngineGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LDAPAuthEngineGroupSpec   `json:"spec,omitempty"`
	Status LDAPAuthEngineGroupStatus `json:"status,omitempty"`
}

func (m *LDAPAuthEngineGroup) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *LDAPAuthEngineGroup) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// LDAPAuthEngineGroupList contains a list of LDAPAuthEngineGroup
type LDAPAuthEngineGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LDAPAuthEngineGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LDAPAuthEngineGroup{}, &LDAPAuthEngineGroupList{})
}

func (i *LDAPAuthEngineGroup) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["name"] = i.Spec.Name
	payload["policies"] = i.Spec.Policies

	return payload
}

func (d *LDAPAuthEngineGroup) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
