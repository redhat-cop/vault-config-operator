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
	"reflect"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LDAPSecretEngineDynamicRoleSpec defines the desired state of LDAPSecretEngineDynamicRole
type LDAPSecretEngineDynamicRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/role/{spec.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	LDAPSEDynamicRole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type LDAPSEDynamicRole struct {
	// CreationLDIF is a templatized LDIF string for creating LDAP user accounts (may be base64 encoded)
	// +kubebuilder:validation:Required
	CreationLDIF string `json:"creationLDIF" vault:"creation_ldif"`

	// DeletionLDIF is a templatized LDIF string for deleting LDAP user accounts (may be base64 encoded)
	// +kubebuilder:validation:Required
	DeletionLDIF string `json:"deletionLDIF" vault:"deletion_ldif"`

	// RollbackLDIF is a templatized LDIF string for rollback on creation failure (recommended)
	// +kubebuilder:validation:Optional
	RollbackLDIF string `json:"rollbackLDIF,omitempty" vault:"rollback_ldif"`

	// UsernameTemplate is a Go template for dynamic username generation
	// +kubebuilder:validation:Optional
	UsernameTemplate string `json:"usernameTemplate,omitempty" vault:"username_template"`

	// DefaultTTL specifies the default TTL for leases (duration format string, e.g. "1h")
	// +kubebuilder:validation:Optional
	DefaultTTL string `json:"defaultTTL,omitempty" vault:"default_ttl,duration"`

	// MaxTTL specifies the maximum TTL for leases (duration format string, e.g. "24h")
	// +kubebuilder:validation:Optional
	MaxTTL string `json:"maxTTL,omitempty" vault:"max_ttl,duration"`
}

var _ vaultutils.VaultObject = &LDAPSecretEngineDynamicRole{}
var _ vaultutils.ConditionsAware = &LDAPSecretEngineDynamicRole{}

func (d *LDAPSecretEngineDynamicRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *LDAPSecretEngineDynamicRole) IsDeletable() bool {
	return true
}

func (d *LDAPSecretEngineDynamicRole) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/role/" + vaultutils.ResourceName(d.Spec.Name, d.Name))
}

func (d *LDAPSecretEngineDynamicRole) GetPayload() map[string]any {
	return d.Spec.LDAPSEDynamicRole.toMap()
}

func (d *LDAPSecretEngineDynamicRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.LDAPSEDynamicRole.toMap()
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (i *LDAPSEDynamicRole) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(i)
}

// LDAPSecretEngineDynamicRoleStatus defines the observed state of LDAPSecretEngineDynamicRole
type LDAPSecretEngineDynamicRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LDAPSecretEngineDynamicRole is the Schema for the ldapsecretengine_dynamicroles API
type LDAPSecretEngineDynamicRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LDAPSecretEngineDynamicRoleSpec   `json:"spec,omitempty"`
	Status LDAPSecretEngineDynamicRoleStatus `json:"status,omitempty"`
}

func (m *LDAPSecretEngineDynamicRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *LDAPSecretEngineDynamicRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// LDAPSecretEngineDynamicRoleList contains a list of LDAPSecretEngineDynamicRole
type LDAPSecretEngineDynamicRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LDAPSecretEngineDynamicRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LDAPSecretEngineDynamicRole{}, &LDAPSecretEngineDynamicRoleList{})
}

func (d *LDAPSecretEngineDynamicRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
