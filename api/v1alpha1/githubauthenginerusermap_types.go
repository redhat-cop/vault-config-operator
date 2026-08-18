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

// GitHubAuthEngineUserMapSpec defines the desired state of GitHubAuthEngineUserMap
type GitHubAuthEngineUserMapSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the GitHub auth engine is mounted.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/map/users/{spec.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// Name is the GitHub username.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Policies is a comma-separated list of policies to assign to this user.
	// +kubebuilder:validation:Optional
	Policies string `json:"policies,omitempty"`
}

// GitHubAuthEngineUserMapStatus defines the observed state of GitHubAuthEngineUserMap
type GitHubAuthEngineUserMapStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GitHubAuthEngineUserMap is the Schema for the githubauthengineusermaps API
type GitHubAuthEngineUserMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitHubAuthEngineUserMapSpec   `json:"spec,omitempty"`
	Status GitHubAuthEngineUserMapStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitHubAuthEngineUserMapList contains a list of GitHubAuthEngineUserMap
type GitHubAuthEngineUserMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitHubAuthEngineUserMap `json:"items"`
}

var _ vaultutils.VaultObject = &GitHubAuthEngineUserMap{}
var _ vaultutils.ConditionsAware = &GitHubAuthEngineUserMap{}

func (d *GitHubAuthEngineUserMap) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *GitHubAuthEngineUserMap) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/map/users/" + d.Spec.Name)
}

func (d *GitHubAuthEngineUserMap) IsDeletable() bool {
	return true
}

func (d *GitHubAuthEngineUserMap) GetPayload() map[string]any {
	return d.toMap()
}

func (d *GitHubAuthEngineUserMap) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.GetPayload()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *GitHubAuthEngineUserMap) IsInitialized() bool {
	return true
}

func (d *GitHubAuthEngineUserMap) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *GitHubAuthEngineUserMap) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *GitHubAuthEngineUserMap) IsValid() (bool, error) {
	return true, nil
}

func (d *GitHubAuthEngineUserMap) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (m *GitHubAuthEngineUserMap) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *GitHubAuthEngineUserMap) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&GitHubAuthEngineUserMap{}, &GitHubAuthEngineUserMapList{})
}

func (d *GitHubAuthEngineUserMap) toMap() map[string]any {
	payload := map[string]any{}
	payload["value"] = d.Spec.Policies
	return payload
}
