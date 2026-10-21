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

// AzureSecretEngineRoleSpec defines the desired state of AzureSecretEngineRole
type AzureSecretEngineRoleSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuraiton to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/auth/{spec.path}/groups/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	AzureSERole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// AzureSecretEngineRoleStatus defines the observed state of AzureSecretEngineRole
type AzureSecretEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AzureSecretEngineRole is the Schema for the azuresecretengineroles API
type AzureSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status AzureSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AzureSecretEngineRoleList contains a list of AzureSecretEngineRole
type AzureSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureSecretEngineRole `json:"items"`
}

type AzureSERole struct {
	// List of Azure roles to be assigned to the generated service principal.
	// The array must be in JSON format, properly escaped as a string. See roles docs for details on role definition.
	// +kubebuilder:validation:Optional
	AzureRoles string `json:"azureRoles,omitempty" vault:"azure_roles"`

	// List of Azure groups that the generated service principal will be assigned to.
	// The array must be in JSON format, properly escaped as a string. See groups docs for more details.
	// +kubebuilder:validation:Optional
	AzureGroups string `json:"azureGroups,omitempty" vault:"azure_groups"`

	// Application Object ID for an existing service principal that will be used instead of creating dynamic service principals.
	// If present, azure_roles will be ignored. See roles docs for details on role definition.
	// +kubebuilder:validation:Optional
	ApplicationObjectID string `json:"applicationObjectID,omitempty" vault:"application_object_id"`

	// If set to true, persists the created service principal and application for the lifetime of the role.
	// Useful for when the Service Principal needs to maintain ownership of objects it creates
	// +kubebuilder:validation:Optional
	PersistApp bool `json:"persistApp,omitempty" vault:"persist_app"`

	// Specifies the default TTL for service principals generated using this role.
	// Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to the system/engine default TTL time.
	// +kubebuilder:validation:Optional
	TTL string `json:"TTL,omitempty" vault:"ttl"`

	// Specifies the maximum TTL for service principals generated using this role.
	// Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to the system/engine max TTL time.
	// +kubebuilder:validation:Optional
	MaxTTL string `json:"maxTTL,omitempty" vault:"max_ttl"`

	// Specifies whether to permanently delete Applications and Service Principals that are dynamically created by Vault.
	// If application_object_id is present, permanently_delete must be false.
	// +kubebuilder:validation:Optional
	PermanentlyDelete string `json:"permanentlyDelete,omitempty" vault:"permanently_delete"`

	// Specifies the security principal types that are allowed to sign in to the application.
	// Valid values are: AzureADMyOrg, AzureADMultipleOrgs, AzureADandPersonalMicrosoftAccount, PersonalMicrosoftAccount.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"AzureADMyOrg","AzureADMultipleOrgs","AzureADandPersonalMicrosoftAccount","PersonalMicrosoftAccount"}
	SignInAudience string `json:"signInAudience,omitempty" vault:"sign_in_audience"`

	// A comma-separated string of Azure tags to attach to an application.
	// +kubebuilder:validation:Optional
	Tags string `json:"tags,omitempty" vault:"tags"`
}

var _ vaultutils.VaultObject = &AzureSecretEngineRole{}
var _ vaultutils.ConditionsAware = &AzureSecretEngineRole{}

func init() {
	SchemeBuilder.Register(&AzureSecretEngineRole{}, &AzureSecretEngineRoleList{})
}

func (r *AzureSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (d *AzureSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}

func (d *AzureSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *AzureSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *AzureSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *AzureSecretEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AzureSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (i *AzureSERole) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(i)
}
