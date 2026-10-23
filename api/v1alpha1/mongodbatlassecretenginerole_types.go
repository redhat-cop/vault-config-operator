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

// MongoDBAtlasSecretEngineRoleSpec defines the desired state of MongoDBAtlasSecretEngineRole
type MongoDBAtlasSecretEngineRoleSpec struct {

	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/roles/{metadata.name}.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	MongoDBAtlasSERole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type MongoDBAtlasSERole struct {
	// OrganizationID is the unique identifier for the Atlas organization.
	// Required if projectID is not set.
	// +kubebuilder:validation:Optional
	OrganizationID string `json:"organizationID,omitempty" vault:"organization_id"`

	// ProjectID is the unique identifier for the Atlas project.
	// Required if organizationID is not set.
	// +kubebuilder:validation:Optional
	ProjectID string `json:"projectID,omitempty" vault:"project_id"`

	// Roles is the list of Atlas roles that the API Key needs to have.
	// At least one role is required.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +listType=set
	Roles []string `json:"roles" vault:"roles"`

	// IPAddresses is the list of IP addresses to add to the API key whitelist.
	// +kubebuilder:validation:Optional
	// +listType=set
	IPAddresses []string `json:"ipAddresses,omitempty" vault:"ip_addresses"`

	// CIDRBlocks is the list of CIDR notation entries to add to the API key whitelist.
	// +kubebuilder:validation:Optional
	// +listType=set
	CIDRBlocks []string `json:"cidrBlocks,omitempty" vault:"cidr_blocks"`

	// ProjectRoles is the list of roles assigned when an Organization API key is
	// assigned to a Project API key.
	// +kubebuilder:validation:Optional
	// +listType=set
	ProjectRoles []string `json:"projectRoles,omitempty" vault:"project_roles"`

	// TTL specifies the duration after which the issued credential should expire.
	// Uses duration format strings (e.g., "2h", "30m").
	// +kubebuilder:validation:Optional
	TTL string `json:"ttl,omitempty" vault:"ttl"`

	// MaxTTL specifies the maximum allowed lifetime of credentials issued using this role.
	// +kubebuilder:validation:Optional
	MaxTTL string `json:"maxTTL,omitempty" vault:"max_ttl"`
}

var _ vaultutils.VaultObject = &MongoDBAtlasSecretEngineRole{}

var _ vaultutils.ConditionsAware = &MongoDBAtlasSecretEngineRole{}

func (d *MongoDBAtlasSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *MongoDBAtlasSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *MongoDBAtlasSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}

func (d *MongoDBAtlasSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *MongoDBAtlasSecretEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.MongoDBAtlasSERole.toMap()
	removeUnsetFields(desiredState, payload)
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	setFields := []string{"roles", "ip_addresses", "cidr_blocks", "project_roles"}
	for _, key := range setFields {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (i *MongoDBAtlasSERole) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(i)
}

// MongoDBAtlasSecretEngineRoleStatus defines the observed state of MongoDBAtlasSecretEngineRole
type MongoDBAtlasSecretEngineRoleStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *MongoDBAtlasSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *MongoDBAtlasSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MongoDBAtlasSecretEngineRole is the Schema for the mongodbatlassecretengineroles API
type MongoDBAtlasSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBAtlasSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status MongoDBAtlasSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongoDBAtlasSecretEngineRoleList contains a list of MongoDBAtlasSecretEngineRole
type MongoDBAtlasSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDBAtlasSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDBAtlasSecretEngineRole{}, &MongoDBAtlasSecretEngineRoleList{})
}

func (d *MongoDBAtlasSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
