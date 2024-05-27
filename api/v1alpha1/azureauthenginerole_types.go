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

// AzureAuthEngineRoleSpec defines the desired state of AzureAuthEngineRole
type AzureAuthEngineRoleSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/groups/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	AzureRole `json:",inline"`
}

// AzureAuthEngineRoleStatus defines the observed state of AzureAuthEngineRole
type AzureAuthEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AzureAuthEngineRole is the Schema for the azureauthengineroles API
type AzureAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status AzureAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AzureAuthEngineRoleList contains a list of AzureAuthEngineRole
type AzureAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureAuthEngineRole `json:"items"`
}

type AzureRole struct {

	// Name of the role.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// The list of Service Principal IDs that login is restricted to.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundServicePrincipalIDs []string `json:"boundServicePrincipalIDs,omitempty"`

	// The list of group ids that login is restricted to.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundGroupIDs []string `json:"boundGroupIDs,omitempty"`

	// The list of locations that login is restricted to.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundLocations []string `json:"boundLocations,omitempty"`

	// The list of subscription IDs that login is restricted to.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundSubscriptionIDs []string `json:"boundSubscriptionIDs,omitempty"`

	// The list of resource groups that login is restricted to.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundResourceGroups []string `json:"boundResourceGroups,omitempty"`

	// The list of scale set names that the login is restricted to.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundScaleSets []string `json:"boundScaleSets,omitempty"`

	// The incremental lifetime for generated tokens.
	//This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenTTL string `json:"tokenTTL,omitempty"`

	// The maximum lifetime for generated tokens.
	// This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

	// List of token policies to encode onto generated tokens.
	// Depending on the auth method, this list may be supplemented by user/group/other values.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	TokenPolicies []string `json:"tokenPolicies,omitempty"`

	// DEPRECATED: Please use the token_policies parameter instead.
	// List of token policies to encode onto generated tokens.
	// Depending on the auth method, this list may be supplemented by user/group/other values.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	Policies []string `json:"policies,omitempty"`

	// List of CIDR blocks.
	// If set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

	// If set, will encode an explicit max TTL onto the token.
	// This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

	// If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy"`

	// The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited.
	// If you require the token to have the ability to create child tokens, you will need to set this value to 0.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	TokenNumUses int64 `json:"tokenNumUses"`

	// The maximum allowed period value when a periodic token is requested from this role.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	TokenPeriod int64 `json:"tokenPeriod"`

	// The type of token that should be generated.
	// Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens).
	// For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time.
	// For machine based authentication cases, you should use batch type tokens.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenType string `json:"tokenType,omitempty"`
}

var _ vaultutils.VaultObject = &AzureAuthEngineRole{}
var _ vaultutils.ConditionsAware = &AzureAuthEngineRole{}

func init() {
	SchemeBuilder.Register(&AzureAuthEngineRole{}, &AzureAuthEngineRoleList{})
}

func (r *AzureAuthEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AzureAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (d *AzureAuthEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *AzureAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *AzureAuthEngineRole) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/role/" + string(r.Spec.Name))
}

func (r *AzureAuthEngineRole) GetPayload() map[string]interface{} {
	return r.Spec.AzureRole.toMap()
}

func (r *AzureAuthEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.AzureRole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (r *AzureAuthEngineRole) IsInitialized() bool {
	return true
}

func (r *AzureAuthEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (d *AzureAuthEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *AzureAuthEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AzureRole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["name"] = r.Name
	payload["bound_service_principal_ids"] = r.BoundServicePrincipalIDs
	payload["bound_group_ids"] = r.BoundGroupIDs
	payload["bound_locations"] = r.BoundLocations
	payload["bound_subscription_ids"] = r.BoundSubscriptionIDs
	payload["bound_resource_groups"] = r.BoundResourceGroups
	payload["bound_scale_sets"] = r.BoundScaleSets
	payload["token_ttl"] = r.TokenTTL
	payload["token_max_ttl"] = r.TokenMaxTTL
	payload["token_policies"] = r.TokenPolicies
	payload["policies"] = r.Policies
	payload["token_bound_cidrs"] = r.TokenBoundCIDRs
	payload["token_explicit_max_ttl"] = r.TokenExplicitMaxTTL
	payload["token_no_default_policy"] = r.TokenNoDefaultPolicy
	payload["token_num_uses"] = r.TokenNumUses
	payload["token_period"] = r.TokenPeriod
	payload["token_type"] = r.TokenType

	return payload
}
