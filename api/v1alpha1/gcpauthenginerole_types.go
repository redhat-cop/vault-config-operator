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

// GCPAuthEngineRoleSpec defines the desired state of GCPAuthEngineRole
type GCPAuthEngineRoleSpec struct {
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

	GCPRole `json:",inline"`
}

// GCPAuthEngineRoleStatus defines the observed state of GCPAuthEngineRole
type GCPAuthEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GCPAuthEngineRole is the Schema for the gcpauthengineroles API
type GCPAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status GCPAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GCPAuthEngineRoleList contains a list of GCPAuthEngineRole
type GCPAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPAuthEngineRole `json:"items"`
}

type GCPRole struct {
	// Name of the role.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// The type of this role. Certain fields correspond to specific roles and will be rejected otherwise. Please see below for more information.
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// An array of service account emails or IDs that login is restricted to, either directly or through an associated instance.
	// If set to *, all service accounts are allowed (you can bind this further using bound_projects.)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	BoundServiceAccounts []string `json:"boundServiceAccounts,omitempty"`

	// An array of GCP project IDs. Only entities belonging to this project can authenticate under the role.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	BoundProjects []string `json:"boundProjects,omitempty"`

	// If true, any auth token generated under this token will have associated group aliases, namely project-$PROJECT_ID, folder-$PROJECT_ID, and organization-$ORG_ID for the entities project and all its folder or organization ancestors.
	// This requires Vault to have IAM permission resourcemanager.projects.get.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	AddGroupAliases bool `json:"addGroupAliases"`

	// The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenTTL string `json:"tokenTTL,omitempty"`

	// The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

	// List of token policies to encode onto generated tokens.
	// Depending on the auth method, this list may be supplemented by user/group/other values.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	TokenPolicies []string `json:"tokenPolicies,omitempty"`

	// DEPRECATED: Please use the token_policies parameter instead. List of token policies to encode onto generated tokens.
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

	// The following parameters are only valid when the role is of type "iam".

	// The number of seconds past the time of authentication that the login param JWT must expire within.
	// For example, if a user attempts to login with a token that expires within an hour and this is set to 15 minutes, Vault will return an error prompting the user to create a new signed JWT with a shorter exp.
	// The GCE metadata tokens currently do not allow the exp claim to be customized.
	// The following parameter is only valid when the role is of type "iam".
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	MaxJWTExp string `json:"maxJWTExp,omitempty"`

	// A flag to determine if this role should allow GCE instances to authenticate by inferring service accounts from the GCE identity metadata token.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	AllowGCEInference bool `json:"allowGCEInference"`

	// The following parameters are only valid when the role is of type "gce"

	// The list of zones that a GCE instance must belong to in order to be authenticated.
	// If bound_instance_groups is provided, it is assumed to be a zonal group and the group must belong to this zone.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundZones []string `json:"boundZones,omitempty"`

	// The list of regions that a GCE instance must belong to in order to be authenticated.
	// If bound_instance_groups is provided, it is assumed to be a regional group and the group must belong to this region.
	// If bound_zones are provided, this attribute is ignored.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundRegions []string `json:"boundRegions,omitempty"`

	// The instance groups that an authorized instance must belong to in order to be authenticated.
	// If specified, either bound_zones or bound_regions must be set too.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundInstanceGroups []string `json:"boundInstanceGroups,omitempty"`

	// A comma-separated list of GCP labels formatted as "key:value" strings that must be set on authorized GCE instances.
	// Because GCP labels are not currently ACL'd, we recommend that this be used in conjunction with other restrictions.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundLabels []string `json:"boundLabels,omitempty"`
}

var _ vaultutils.VaultObject = &GCPAuthEngineRole{}
var _ vaultutils.ConditionsAware = &GCPAuthEngineRole{}

func init() {
	SchemeBuilder.Register(&GCPAuthEngineRole{}, &GCPAuthEngineRoleList{})
}

func (d *GCPAuthEngineRole) IsDeletable() bool {
	return true
}

func (r *GCPAuthEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *GCPAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (d *GCPAuthEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *GCPAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *GCPAuthEngineRole) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/role/" + string(r.Spec.Name))
}

func (r *GCPAuthEngineRole) GetPayload() map[string]interface{} {
	return r.Spec.GCPRole.toMap()
}

func (r *GCPAuthEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.GCPRole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (r *GCPAuthEngineRole) IsInitialized() bool {
	return true
}

func (r *GCPAuthEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (d *GCPAuthEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *GCPAuthEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *GCPRole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["name"] = r.Name
	payload["type"] = r.Type
	payload["bound_service_accounts"] = r.BoundServiceAccounts
	payload["bound_projects"] = r.BoundProjects
	payload["add_group_aliases"] = r.AddGroupAliases
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
	payload["max_jwt_exp"] = r.MaxJWTExp
	payload["allow_gce_inference"] = r.AllowGCEInference
	payload["bound_zones"] = r.BoundZones
	payload["bound_regions"] = r.BoundRegions
	payload["bound_instance_groups"] = r.BoundInstanceGroups
	payload["bound_labels"] = r.BoundLabels

	return payload
}
