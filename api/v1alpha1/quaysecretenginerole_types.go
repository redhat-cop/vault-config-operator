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
	"encoding/json"
	"reflect"

	"github.com/redhat-cop/operator-utils/pkg/util/apis"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:validation:Enum:={"admin","read","write"}
type Permission string

// +kubebuilder:validation:Enum:={"admin","creator","member"}
type TeamRole string

type NamespaceType string

const (
	TeamRoleAdmin             TeamRole      = "admin"
	TeamRoleCreator           TeamRole      = "creator"
	TeamRoleMember            TeamRole      = "member"
	NamespaceTypeUser         NamespaceType = "user"
	NamespaceTypeOrganization NamespaceType = "organization"
	PermissionAdmin           Permission    = "admin"
	PermissionRead            Permission    = "read"
	PermissionWrite           Permission    = "write"
)

// QuaySecretEngineRoleSpec defines the desired state of QuaySecretEngineRole
type QuaySecretEngineRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path will be {[spec.authentication.namespace]}/{spec.path}/roles/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	QuayRole `json:",inline"`
}

var _ vaultutils.VaultObject = &QuaySecretEngineRole{}

func (d *QuaySecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (q *QuaySecretEngineRole) GetPath() string {
	return string(q.Spec.Path) + "/" + "roles" + "/" + q.Name
}
func (q *QuaySecretEngineRole) GetPayload() map[string]interface{} {
	return q.Spec.toMap()
}
func (q *QuaySecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := q.Spec.QuayRole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (q *QuaySecretEngineRole) IsInitialized() bool {
	return true
}

func (q *QuaySecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (q *QuaySecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (r *QuayRole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["namespace_type"] = r.NamespaceType
	payload["namespace_name"] = r.NamespaceName
	payload["create_repositories"] = r.CreateRepositories
	if r.DefaultPermission != nil {
		payload["default_permission"] = r.DefaultPermission
	}
	if r.Teams != nil {
		setMapJson(payload, "teams", r.Teams)

	}
	if r.Repositories != nil {
		setMapJson(payload, "repositories", r.Repositories)
	}
	payload["ttl"] = r.TTL
	payload["max_ttl"] = r.MaxTTL
	return payload
}

func setMapJson(payload map[string]interface{}, fieldName string, field interface{}) {
	j, err := json.Marshal(field)
	if err == nil {
		payload[fieldName] = string(j)
	}

}

// QuaySecretEngineRoleStatus defines the observed state of QuaySecretEngineRole
type QuaySecretEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ apis.ConditionsAware = &QuaySecretEngineRole{}

func (q *QuaySecretEngineRole) GetConditions() []metav1.Condition {
	return q.Status.Conditions
}

func (q *QuaySecretEngineRole) SetConditions(conditions []metav1.Condition) {
	q.Status.Conditions = conditions
}

type QuayBaseRole struct {
	// NamespaceType Type of account namespace to manage.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"organization","user"}
	// +kubebuilder:default="organization"
	NamespaceType NamespaceType `json:"namespaceType,omitempty"`

	// NamespaceName Name of the Quay account.
	// +kubebuilder:validation:Required
	NamespaceName string `json:"namespaceName,omitempty"`

	// CreateRepositories Access to create Quay repositories.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	CreateRepositories *bool `json:"createRepositories,omitempty"`

	// Teams Permissions granted to the Robot Account to Teams.
	// +kubebuilder:validation:Optional
	Teams *map[string]TeamRole `json:"teams,omitempty"`

	// Teams Permissions granted to the Robot Account to Repositories.
	// +kubebuilder:validation:Optional
	Repositories *map[string]Permission `json:"repositories,omitempty"`

	// DefaultPermission Permissions granted to the Robot Account in newly created repositories
	// +kubebuilder:validation:Optional
	DefaultPermission *Permission `json:"defaultPermission,omitempty"`
}

type QuayRole struct {
	QuayBaseRole `json:",inline"`

	// TTL Time-to-Live for the credential
	// +kubebuilder:validation:Optional
	TTL *metav1.Duration `json:"TTL,omitempty"`

	// MaxTTL Maximum Time-to-Live for the credential
	// +kubebuilder:validation:Optional
	MaxTTL *metav1.Duration `json:"maxTTL,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// QuaySecretEngineRole is the Schema for the quaysecretengineroles API
type QuaySecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuaySecretEngineRoleSpec   `json:"spec,omitempty"`
	Status QuaySecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuaySecretEngineRoleList contains a list of QuaySecretEngineRole
type QuaySecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuaySecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuaySecretEngineRole{}, &QuaySecretEngineRoleList{})
}

func (d *QuaySecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
