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
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/roles/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	DBSERole `json:",inline"`
}

var _ vaultutils.VaultObject = &DatabaseSecretEngineRole{}

var _ vaultresourcecontroller.ConditionsAware = &DatabaseSecretEngineRole{}

func (d *DatabaseSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *DatabaseSecretEngineRole) GetPath() string {
	return string(d.Spec.Path) + "/" + "roles" + "/" + d.Name
}
func (d *DatabaseSecretEngineRole) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *DatabaseSecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.DBSERole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (d *DatabaseSecretEngineRole) IsInitialized() bool {
	return true
}

func (d *DatabaseSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *DatabaseSecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

type DBSERole struct {
	// DBName The name of the database connection to use for this role.
	// +kubebuilder:validation:Required
	DBName string `json:"dBName,omitempty"`

	// DeafulTTL Specifies the TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="0s"
	DefaultTTL metav1.Duration `json:"defaultTTL,omitempty"`

	// MaxTTL Specifies the maximum TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/mount default TTL time; this value is allowed to be less than the mount max TTL (or, if not set, the system max TTL), but it is not allowed to be longer. See also The TTL General Case.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="0s"
	MaxTTL metav1.Duration `json:"maxTTL,omitempty"`

	// CreationStatements Specifies the database statements executed to create and configure a user. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	CreationStatements []string `json:"creationStatements,omitempty"`

	// RevocationStatements Specifies the database statements to be executed to revoke a user. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RevocationStatements []string `json:"revocationStatements,omitempty"`

	// RollbackStatements Specifies the database statements to be executed to rollback a create operation in the event of an error. Not every plugin type will support this functionality. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RollbackStatements []string `json:"rollbackStatements,omitempty"`

	// RenewStatements Specifies the database statements to be executed to renew a user. Not every plugin type will support this functionality. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RenewStatements []string `json:"renewStatements,omitempty"`
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

func (i *DBSERole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["db_name"] = i.DBName
	payload["default_ttl"] = i.DefaultTTL
	payload["max_ttl"] = i.MaxTTL
	payload["creation_statements"] = i.CreationStatements
	payload["revocation_statements"] = i.RevocationStatements
	payload["rollback_statements"] = i.RollbackStatements
	payload["renew_statements"] = i.RenewStatements
	return payload
}

func (d *DatabaseSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
