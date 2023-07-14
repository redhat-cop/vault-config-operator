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

	"github.com/redhat-cop/operator-utils/pkg/util/apis"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitHubSecretEngineRoleSpec defines the desired state of GitHubSecretEngineRole
type GitHubSecretEngineRoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/permissionset/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// PermissionsSet All parameters are optional. Omitting them results in a token that has access to all of the repositories and permissions that the GitHub App has.
	// When crafting Vault policy, hyper security sensitive organisations may wish to favour repository_ids (GitHub repository IDs are immutable) instead of repositories (GitHub repository names are mutable).
	// +kubebuilder:validation:Optional
	PermissionSet `json:",inline"`
}

type PermissionSet struct {

	//  InstallationID the ID of the app installation. Note the Installation ID from the URL of this page (usually: https://github.com/settings/installations/<installation id>) if you wish to configure using the installation ID directly. Only one of installationID or organizationName is required. If both are provided, installationID takes precedence.
	// +kubebuilder:validation:Optional
	InstallationID int64 `json:"installationID,omitempty"`

	// OrganizationName the name of the organization with the GitHub App installation. Only one of installationID or organizationName is required. If both are provided, installationID takes precedence.
	// +kubebuilder:validation:Optional
	OrganizationName string `json:"organizationName,omitempty"`

	// Repositories a list of the names of the repositories within the organisation that the installation token can access
	// +kubebuilder:validation:Optional
	Repositories []string `json:"repositories,omitempty"`

	// Repositories a list of the IDs of the repositories that the installation token can access. See [this StackOverflow](https://stackoverflow.com/a/47223479) post for the quickest way to find a repository ID
	// +kubebuilder:validation:Optional
	RepositoriesIDs []string `json:"repositoriesIDs,omitempty"`

	// Permissions a key value map of permission names to their access type (read or write). See [GitHub’s documentation](https://developer.github.com/v3/apps/permissions) on permission names and access types.
	// +kubebuilder:validation:Optional
	Permissions map[string]string `json:"permissions,omitempty"`
}

func (i *PermissionSet) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["installation_id"] = i.InstallationID
	payload["org_name"] = i.OrganizationName
	payload["repositories"] = i.Repositories
	payload["repository_ids"] = i.RepositoriesIDs
	payload["permissions"] = i.Permissions
	return payload
}

var _ vaultutils.VaultObject = &GitHubSecretEngineRole{}

func (d *GitHubSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *GitHubSecretEngineRole) GetPath() string {
	return string(d.Spec.Path) + "/" + "permissionset" + "/" + d.Name
}
func (d *GitHubSecretEngineRole) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *GitHubSecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.PermissionSet.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (d *GitHubSecretEngineRole) IsInitialized() bool {
	return true
}

func (d *GitHubSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *GitHubSecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

// GitHubSecretEngineRoleStatus defines the observed state of GitHubSecretEngineRole
type GitHubSecretEngineRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster

	// Important: Run "make" to regenerate code after modifying this file
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ apis.ConditionsAware = &GitHubSecretEngineRole{}

func (m *GitHubSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *GitHubSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GitHubSecretEngineRole is the Schema for the githubsecretengineroles API
type GitHubSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitHubSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status GitHubSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitHubSecretEngineRoleList contains a list of GitHubSecretEngineRole
type GitHubSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitHubSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitHubSecretEngineRole{}, &GitHubSecretEngineRoleList{})
}

func (d *GitHubSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
