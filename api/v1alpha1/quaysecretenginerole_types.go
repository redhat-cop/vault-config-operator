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
	"encoding/json"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/roles/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	QuayRole `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

var _ vaultutils.VaultObject = &QuaySecretEngineRole{}

func (d *QuaySecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *QuaySecretEngineRole) IsDeletable() bool {
	return true
}

func (d *QuaySecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}
func (q *QuaySecretEngineRole) GetPayload() map[string]any {
	return q.Spec.toMap()
}

func (r *QuayRole) toMap() map[string]any {
	payload := r.QuayBaseRole.toMap()
	payload["ttl"] = r.TTL
	payload["max_ttl"] = r.MaxTTL
	return payload
}

func setMapJson(payload map[string]any, fieldName string, field any) {
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

var _ vaultutils.ConditionsAware = &QuaySecretEngineRole{}

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
	NamespaceType NamespaceType `json:"namespaceType,omitempty" vault:"namespace_type"`

	// NamespaceName Name of the Quay account.
	// +kubebuilder:validation:Optional
	NamespaceName string `json:"namespaceName,omitempty" vault:"namespace_name"`

	// CreateRepositories Access to create Quay repositories.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	CreateRepositories *bool `json:"createRepositories" vault:"-"`

	// Teams Permissions granted to the Robot Account to Teams.
	// +kubebuilder:validation:Optional
	Teams *map[string]TeamRole `json:"teams,omitempty" vault:"-"`

	// Repositories Permissions granted to the Robot Account to Repositories.
	// +kubebuilder:validation:Optional
	Repositories *map[string]Permission `json:"repositories,omitempty" vault:"-"`

	// DefaultPermission Default permission applied to existing and newly created repositories in the namespace.
	// When set on an organization namespace, the robot account receives this permission on all current
	// repositories and an organization prototype is created so future repositories inherit the same permission.
	// Per-repository overrides in the Repositories field take precedence over this default.
	// +kubebuilder:validation:Optional
	DefaultPermission *Permission `json:"defaultPermission,omitempty" vault:"-"`
}

type QuayRole struct {
	QuayBaseRole `json:",inline"`

	// TTL Time-to-Live for the credential
	// +kubebuilder:validation:Optional
	TTL *metav1.Duration `json:"TTL,omitempty" vault:"-"`

	// MaxTTL Maximum Time-to-Live for the credential
	// +kubebuilder:validation:Optional
	MaxTTL *metav1.Duration `json:"maxTTL,omitempty" vault:"-"`
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
