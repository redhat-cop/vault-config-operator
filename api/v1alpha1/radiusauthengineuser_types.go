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

// RADIUSAuthEngineUserSpec defines the desired state of RADIUSAuthEngineUser
type RADIUSAuthEngineUserSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the RADIUS auth engine is mounted.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// Name of the RADIUS user. If specified, takes precedence over metadata.name.
	// +kubebuilder:validation:Optional
	Name string `json:"name,omitempty"`

	// Policies is a comma-separated list of policies associated with this user.
	// +kubebuilder:validation:Optional
	Policies string `json:"policies,omitempty"`
}

var _ vaultutils.VaultObject = &RADIUSAuthEngineUser{}
var _ vaultutils.ConditionsAware = &RADIUSAuthEngineUser{}

func (d *RADIUSAuthEngineUser) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *RADIUSAuthEngineUser) GetPath() string {
	name := d.Spec.Name
	if name == "" {
		name = d.Name
	}
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/users/" + name)
}

func (d *RADIUSAuthEngineUser) IsDeletable() bool {
	return true
}

func (d *RADIUSAuthEngineUser) GetPayload() map[string]any {
	return d.toMap()
}

func (d *RADIUSAuthEngineUser) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.GetPayload()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *RADIUSAuthEngineUser) IsInitialized() bool {
	return true
}

func (d *RADIUSAuthEngineUser) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *RADIUSAuthEngineUser) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *RADIUSAuthEngineUser) IsValid() (bool, error) {
	return true, nil
}

// RADIUSAuthEngineUserStatus defines the observed state of RADIUSAuthEngineUser
type RADIUSAuthEngineUserStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RADIUSAuthEngineUser is the Schema for the radiusauthengineusers API
type RADIUSAuthEngineUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RADIUSAuthEngineUserSpec   `json:"spec,omitempty"`
	Status RADIUSAuthEngineUserStatus `json:"status,omitempty"`
}

func (m *RADIUSAuthEngineUser) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *RADIUSAuthEngineUser) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// RADIUSAuthEngineUserList contains a list of RADIUSAuthEngineUser
type RADIUSAuthEngineUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RADIUSAuthEngineUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RADIUSAuthEngineUser{}, &RADIUSAuthEngineUserList{})
}

func (d *RADIUSAuthEngineUser) toMap() map[string]any {
	payload := map[string]any{}
	payload["policies"] = d.Spec.Policies
	return payload
}

func (d *RADIUSAuthEngineUser) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
