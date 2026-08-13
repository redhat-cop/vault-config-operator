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

// NomadSecretEngineRoleSpec defines the desired state of NomadSecretEngineRole
type NomadSecretEngineRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/role/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	NomadSERole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type NomadSERole struct {
	// Policies is the list of Nomad ACL policies to assign to the generated token.
	// These must be created beforehand in Nomad.
	// +kubebuilder:validation:Optional
	// +listType=set
	Policies []string `json:"policies,omitempty"`

	// Global specifies if the token should be global (replicated across Nomad regions).
	// +kubebuilder:validation:Optional
	Global bool `json:"global,omitempty"`

	// Type specifies the type of Nomad token to create: "client" or "management".
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="client"
	// +kubebuilder:validation:Enum:={"client","management"}
	Type string `json:"type"`
}

var _ vaultutils.VaultObject = &NomadSecretEngineRole{}

var _ vaultutils.ConditionsAware = &NomadSecretEngineRole{}

func (d *NomadSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *NomadSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *NomadSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "role" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "role" + "/" + d.Name)
}

func (d *NomadSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *NomadSecretEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.NomadSERole.toMap()
	if typeVal, ok := desiredState["type"]; ok {
		desiredState["token_type"] = typeVal
		delete(desiredState, "type")
	}
	removeUnsetFields(desiredState, payload)
	if globalVal, ok := desiredState["global"].(bool); ok && !globalVal {
		if _, inPayload := payload["global"]; !inPayload {
			delete(desiredState, "global")
		}
	}
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	sortAnyStringSlice(desiredState, "policies")
	sortAnyStringSlice(filteredPayload, "policies")
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (d *NomadSecretEngineRole) IsInitialized() bool {
	return true
}

func (d *NomadSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *NomadSecretEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *NomadSecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (i *NomadSERole) toMap() map[string]any {
	payload := map[string]any{}
	payload["policies"] = toInterfaceArray(i.Policies)
	payload["global"] = i.Global
	payload["type"] = i.Type
	return payload
}

// NomadSecretEngineRoleStatus defines the observed state of NomadSecretEngineRole
type NomadSecretEngineRoleStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *NomadSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *NomadSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NomadSecretEngineRole is the Schema for the nomadsecretengineroles API
type NomadSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NomadSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status NomadSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NomadSecretEngineRoleList contains a list of NomadSecretEngineRole
type NomadSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NomadSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NomadSecretEngineRole{}, &NomadSecretEngineRoleList{})
}

func (d *NomadSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
