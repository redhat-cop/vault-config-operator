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
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LDAPSecretEngineStaticRoleSpec defines the desired state of LDAPSecretEngineStaticRole
type LDAPSecretEngineStaticRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/static-role/{spec.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	LDAPSEStaticRole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type LDAPSEStaticRole struct {
	// Username is the existing LDAP username to manage password rotation for. Cannot be modified after creation.
	// +kubebuilder:validation:Required
	Username string `json:"username"`

	// DN is the Distinguished Name of the existing LDAP entry. Takes precedence over username. Cannot be modified after creation.
	// +kubebuilder:validation:Optional
	DN string `json:"dn,omitempty"`

	// RotationPeriod is the time in seconds before credential rotation (minimum 10s)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=10
	RotationPeriod int `json:"rotationPeriod"`

	// SkipImportRotation when true skips the initial password rotation on role creation
	// +kubebuilder:validation:Optional
	SkipImportRotation bool `json:"skipImportRotation,omitempty"`
}

var _ vaultutils.VaultObject = &LDAPSecretEngineStaticRole{}
var _ vaultutils.ConditionsAware = &LDAPSecretEngineStaticRole{}

func (d *LDAPSecretEngineStaticRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *LDAPSecretEngineStaticRole) IsDeletable() bool {
	return true
}

func (d *LDAPSecretEngineStaticRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/static-role/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/static-role/" + d.Name)
}

func (d *LDAPSecretEngineStaticRole) GetPayload() map[string]any {
	return d.Spec.LDAPSEStaticRole.toMap()
}

func (d *LDAPSecretEngineStaticRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.LDAPSEStaticRole.toMap()
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *LDAPSecretEngineStaticRole) IsInitialized() bool {
	return true
}

func (d *LDAPSecretEngineStaticRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *LDAPSecretEngineStaticRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *LDAPSecretEngineStaticRole) IsValid() (bool, error) {
	return true, nil
}

func (i *LDAPSEStaticRole) toMap() map[string]any {
	payload := map[string]any{}
	payload["username"] = i.Username
	payload["dn"] = i.DN
	payload["rotation_period"] = json.Number(strconv.Itoa(i.RotationPeriod))
	payload["skip_import_rotation"] = i.SkipImportRotation
	return payload
}

// LDAPSecretEngineStaticRoleStatus defines the observed state of LDAPSecretEngineStaticRole
type LDAPSecretEngineStaticRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LDAPSecretEngineStaticRole is the Schema for the ldapsecretenginestaticroles API
type LDAPSecretEngineStaticRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LDAPSecretEngineStaticRoleSpec   `json:"spec,omitempty"`
	Status LDAPSecretEngineStaticRoleStatus `json:"status,omitempty"`
}

func (m *LDAPSecretEngineStaticRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *LDAPSecretEngineStaticRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// LDAPSecretEngineStaticRoleList contains a list of LDAPSecretEngineStaticRole
type LDAPSecretEngineStaticRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LDAPSecretEngineStaticRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LDAPSecretEngineStaticRole{}, &LDAPSecretEngineStaticRoleList{})
}

func (d *LDAPSecretEngineStaticRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
