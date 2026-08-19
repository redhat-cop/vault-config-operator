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

// OktaAuthEngineGroupSpec defines the desired state of OktaAuthEngineGroup
type OktaAuthEngineGroupSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the Okta auth engine is mounted.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// Name of the Okta group.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Policies is a comma-separated list of policies associated with the group.
	// +kubebuilder:validation:Optional
	Policies string `json:"policies,omitempty"`
}

var _ vaultutils.VaultObject = &OktaAuthEngineGroup{}
var _ vaultutils.ConditionsAware = &OktaAuthEngineGroup{}

func (d *OktaAuthEngineGroup) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *OktaAuthEngineGroup) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/groups/" + d.Spec.Name)
}

func (d *OktaAuthEngineGroup) IsDeletable() bool {
	return true
}

func (d *OktaAuthEngineGroup) GetPayload() map[string]any {
	return d.toMap()
}

func (d *OktaAuthEngineGroup) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.GetPayload()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *OktaAuthEngineGroup) IsInitialized() bool {
	return true
}

func (d *OktaAuthEngineGroup) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *OktaAuthEngineGroup) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *OktaAuthEngineGroup) IsValid() (bool, error) {
	return true, nil
}

// OktaAuthEngineGroupStatus defines the observed state of OktaAuthEngineGroup
type OktaAuthEngineGroupStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OktaAuthEngineGroup is the Schema for the oktaauthenginegroups API
type OktaAuthEngineGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OktaAuthEngineGroupSpec   `json:"spec,omitempty"`
	Status OktaAuthEngineGroupStatus `json:"status,omitempty"`
}

func (m *OktaAuthEngineGroup) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *OktaAuthEngineGroup) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// OktaAuthEngineGroupList contains a list of OktaAuthEngineGroup
type OktaAuthEngineGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OktaAuthEngineGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OktaAuthEngineGroup{}, &OktaAuthEngineGroupList{})
}

func (i *OktaAuthEngineGroup) toMap() map[string]any {
	payload := map[string]any{}
	payload["policies"] = i.Spec.Policies
	return payload
}

func (d *OktaAuthEngineGroup) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
