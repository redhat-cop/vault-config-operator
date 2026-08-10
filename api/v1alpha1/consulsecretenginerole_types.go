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

// ConsulSecretEngineRoleSpec defines the desired state of ConsulSecretEngineRole
type ConsulSecretEngineRoleSpec struct {

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

	ConsulSERole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type ConsulSERole struct {
	// ConsulPolicies is the list of Consul ACL policies to assign to the generated token.
	// +kubebuilder:validation:Optional
	// +listType=set
	ConsulPolicies []string `json:"consulPolicies,omitempty"`

	// ConsulRoles is the list of Consul roles to attach to the generated token (Consul 1.5+).
	// +kubebuilder:validation:Optional
	// +listType=set
	ConsulRoles []string `json:"consulRoles,omitempty"`

	// ServiceIdentities is the list of Consul service identities to assign (Consul 1.5+).
	// +kubebuilder:validation:Optional
	// +listType=set
	ServiceIdentities []string `json:"serviceIdentities,omitempty"`

	// NodeIdentities is the list of Consul node identities to assign (Consul 1.8+).
	// +kubebuilder:validation:Optional
	// +listType=set
	NodeIdentities []string `json:"nodeIdentities,omitempty"`

	// ConsulNamespace specifies the Consul Enterprise namespace for the token (Consul 1.7+).
	// +kubebuilder:validation:Optional
	ConsulNamespace string `json:"consulNamespace,omitempty"`

	// Partition specifies the Consul admin partition for the token (Consul 1.11+).
	// +kubebuilder:validation:Optional
	Partition string `json:"partition,omitempty"`

	// Local if true creates a token that is not replicated globally (Consul 1.4+).
	// +kubebuilder:validation:Optional
	Local bool `json:"local,omitempty"`

	// TTL specifies the TTL for the generated Consul token. Uses duration format strings.
	// +kubebuilder:validation:Optional
	TTL string `json:"ttl,omitempty"`

	// MaxTTL specifies the max TTL for the generated Consul token. Uses duration format strings.
	// +kubebuilder:validation:Optional
	MaxTTL string `json:"maxTTL,omitempty"`
}

var _ vaultutils.VaultObject = &ConsulSecretEngineRole{}

var _ vaultutils.ConditionsAware = &ConsulSecretEngineRole{}

func (d *ConsulSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *ConsulSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *ConsulSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}

func (d *ConsulSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *ConsulSecretEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.ConsulSERole.toMap()
	removeUnsetFields(desiredState, payload)
	if localVal, ok := desiredState["local"].(bool); ok && !localVal {
		if _, inPayload := payload["local"]; !inPayload {
			delete(desiredState, "local")
		}
	}
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	setFields := []string{"consul_policies", "consul_roles", "service_identities", "node_identities"}
	for _, key := range setFields {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (d *ConsulSecretEngineRole) IsInitialized() bool {
	return true
}

func (d *ConsulSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *ConsulSecretEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *ConsulSecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (i *ConsulSERole) toMap() map[string]any {
	payload := map[string]any{}
	payload["consul_policies"] = toInterfaceArray(i.ConsulPolicies)
	payload["consul_roles"] = toInterfaceArray(i.ConsulRoles)
	payload["service_identities"] = toInterfaceArray(i.ServiceIdentities)
	payload["node_identities"] = toInterfaceArray(i.NodeIdentities)
	payload["consul_namespace"] = i.ConsulNamespace
	payload["partition"] = i.Partition
	payload["local"] = i.Local
	payload["ttl"] = i.TTL
	payload["max_ttl"] = i.MaxTTL
	return payload
}

// ConsulSecretEngineRoleStatus defines the observed state of ConsulSecretEngineRole
type ConsulSecretEngineRoleStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *ConsulSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *ConsulSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConsulSecretEngineRole is the Schema for the consulsecretengineroles API
type ConsulSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsulSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status ConsulSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConsulSecretEngineRoleList contains a list of ConsulSecretEngineRole
type ConsulSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConsulSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConsulSecretEngineRole{}, &ConsulSecretEngineRoleList{})
}

func (d *ConsulSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
