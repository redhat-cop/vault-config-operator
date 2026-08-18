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

// GitHubAuthEngineTeamMapSpec defines the desired state of GitHubAuthEngineTeamMap
type GitHubAuthEngineTeamMapSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the GitHub auth engine is mounted.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/map/teams/{spec.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// Name is the GitHub team name in slugified format.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name"`

	// Policies is a comma-separated list of policies to assign to this team.
	// +kubebuilder:validation:Optional
	Policies string `json:"policies,omitempty"`
}

// GitHubAuthEngineTeamMapStatus defines the observed state of GitHubAuthEngineTeamMap
type GitHubAuthEngineTeamMapStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GitHubAuthEngineTeamMap is the Schema for the githubauthengineteammaps API
type GitHubAuthEngineTeamMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitHubAuthEngineTeamMapSpec   `json:"spec,omitempty"`
	Status GitHubAuthEngineTeamMapStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitHubAuthEngineTeamMapList contains a list of GitHubAuthEngineTeamMap
type GitHubAuthEngineTeamMapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitHubAuthEngineTeamMap `json:"items"`
}

var _ vaultutils.VaultObject = &GitHubAuthEngineTeamMap{}
var _ vaultutils.ConditionsAware = &GitHubAuthEngineTeamMap{}

func (d *GitHubAuthEngineTeamMap) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *GitHubAuthEngineTeamMap) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/map/teams/" + d.Spec.Name)
}

func (d *GitHubAuthEngineTeamMap) IsDeletable() bool {
	return true
}

func (d *GitHubAuthEngineTeamMap) GetPayload() map[string]any {
	return d.toMap()
}

func (d *GitHubAuthEngineTeamMap) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.GetPayload()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *GitHubAuthEngineTeamMap) IsInitialized() bool {
	return true
}

func (d *GitHubAuthEngineTeamMap) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *GitHubAuthEngineTeamMap) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *GitHubAuthEngineTeamMap) IsValid() (bool, error) {
	return true, nil
}

func (d *GitHubAuthEngineTeamMap) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (m *GitHubAuthEngineTeamMap) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *GitHubAuthEngineTeamMap) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&GitHubAuthEngineTeamMap{}, &GitHubAuthEngineTeamMapList{})
}

func (d *GitHubAuthEngineTeamMap) toMap() map[string]any {
	payload := map[string]any{}
	payload["value"] = d.Spec.Policies
	return payload
}
