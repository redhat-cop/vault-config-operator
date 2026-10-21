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
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseSecretEngineRoleSpec defines the desired state of DatabaseSecretEngineRole
type DatabaseSecretEngineRoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/roles/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	DBSERole `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

var _ vaultutils.VaultObject = &DatabaseSecretEngineRole{}

var _ vaultutils.ConditionsAware = &DatabaseSecretEngineRole{}

func (d *DatabaseSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *DatabaseSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *DatabaseSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}
func (d *DatabaseSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

type DBSERole struct {
	// DBName The name of the database connection to use for this role.
	// +kubebuilder:validation:Optional
	DBName string `json:"dBName,omitempty" vault:"db_name"`

	// DeafulTTL Specifies the TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.
	// +kubebuilder:validation:Optional
	DefaultTTL metav1.Duration `json:"defaultTTL,omitempty" vault:"default_ttl"`

	// MaxTTL Specifies the maximum TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/mount default TTL time; this value is allowed to be less than the mount max TTL (or, if not set, the system max TTL), but it is not allowed to be longer. See also The TTL General Case.
	// +kubebuilder:validation:Optional
	MaxTTL metav1.Duration `json:"maxTTL,omitempty" vault:"max_ttl"`

	// CreationStatements Specifies the database statements executed to create and configure a user. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	CreationStatements []string `json:"creationStatements,omitempty" vault:"creation_statements"`

	// RevocationStatements Specifies the database statements to be executed to revoke a user. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RevocationStatements []string `json:"revocationStatements,omitempty" vault:"revocation_statements"`

	// RollbackStatements Specifies the database statements to be executed to rollback a create operation in the event of an error. Not every plugin type will support this functionality. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RollbackStatements []string `json:"rollbackStatements,omitempty" vault:"rollback_statements"`

	// RenewStatements Specifies the database statements to be executed to renew a user. Not every plugin type will support this functionality. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RenewStatements []string `json:"renewStatements,omitempty" vault:"renew_statements"`
}

// DatabaseSecretEngineRoleStatus defines the observed state of DatabaseSecretEngineRole
type DatabaseSecretEngineRoleStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *DatabaseSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *DatabaseSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DatabaseSecretEngineRole is the Schema for the databasesecretengineroles API
type DatabaseSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status DatabaseSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseSecretEngineRoleList contains a list of DatabaseSecretEngineRole
type DatabaseSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseSecretEngineRole{}, &DatabaseSecretEngineRoleList{})
}

func (i *DBSERole) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(i)
}

func (d *DatabaseSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
