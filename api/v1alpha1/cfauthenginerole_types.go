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

// CFAuthEngineRoleSpec defines the desired state of CFAuthEngineRole
type CFAuthEngineRoleSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the CF auth engine is mounted.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/roles/{name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// Name of the CF role. Overrides metadata.name for the Vault object name.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`

	CFAuthRole `json:",inline"`
}

type CFAuthRole struct {
	// BoundApplicationIDs constrains instances to specific application IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundApplicationIDs []string `json:"boundApplicationIDs,omitempty"`

	// BoundSpaceIDs constrains instances to specific space IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundSpaceIDs []string `json:"boundSpaceIDs,omitempty"`

	// BoundOrganizationIDs constrains instances to specific organization IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundOrganizationIDs []string `json:"boundOrganizationIDs,omitempty"`

	// BoundInstanceIDs constrains to specific instance IDs. Changes on cf push.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundInstanceIDs []string `json:"boundInstanceIDs,omitempty"`

	// DisableIPMatching if true, disables IP-to-cert matching for proxied logins.
	// +kubebuilder:validation:Optional
	DisableIPMatching bool `json:"disableIPMatching,omitempty"`

	// TokenTTL is the incremental lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty"`

	// TokenMaxTTL is the maximum lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

	// TokenPolicies are policies to encode onto generated tokens.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenPolicies []string `json:"tokenPolicies,omitempty"`

	// Policies is deprecated — use tokenPolicies instead.
	// +kubebuilder:validation:Optional
	// +listType=set
	Policies []string `json:"policies,omitempty"`

	// TokenBoundCIDRs are CIDR blocks restricting authentication and tying the token.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

	// TokenExplicitMaxTTL is the hard cap max TTL for tokens.
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

	// TokenNoDefaultPolicy if true, omits the default policy from generated tokens.
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`

	// TokenNumUses is the max number of times a generated token may be used (0 = unlimited).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int64 `json:"tokenNumUses,omitempty"`

	// TokenPeriod is the maximum allowed period for periodic tokens.
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty"`

	// TokenType is the type of token to generate.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
	TokenType string `json:"tokenType,omitempty"`
}

// CFAuthEngineRoleStatus defines the observed state of CFAuthEngineRole
type CFAuthEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CFAuthEngineRole is the Schema for the cfauthengineroles API
type CFAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CFAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status CFAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CFAuthEngineRoleList contains a list of CFAuthEngineRole
type CFAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CFAuthEngineRole `json:"items"`
}

var _ vaultutils.VaultObject = &CFAuthEngineRole{}
var _ vaultutils.ConditionsAware = &CFAuthEngineRole{}

func init() {
	SchemeBuilder.Register(&CFAuthEngineRole{}, &CFAuthEngineRoleList{})
}

func (d *CFAuthEngineRole) IsDeletable() bool {
	return true
}

func (r *CFAuthEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *CFAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (d *CFAuthEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *CFAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *CFAuthEngineRole) GetPath() string {
	if r.Spec.Name != "" {
		return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/roles/" + r.Spec.Name)
	}
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/roles/" + r.Name)
}

func (r *CFAuthEngineRole) GetPayload() map[string]any {
	return r.Spec.CFAuthRole.toMap()
}

// cfRoleVaultReadAliases maps Vault read-response keys to the canonical write keys.
// Vault returns "ttl" instead of "token_ttl", "max_ttl" instead of "token_max_ttl", etc.
var cfRoleVaultReadAliases = map[string]string{
	"ttl":         "token_ttl",
	"max_ttl":     "token_max_ttl",
	"period":      "token_period",
	"bound_cidrs": "token_bound_cidrs",
}

func (r *CFAuthEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.CFAuthRole.toMap()
	normalizeVaultReadAliases(payload, cfRoleVaultReadAliases)
	removeUnsetFields(desiredState, payload)
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	setFields := []string{
		"bound_application_ids", "bound_space_ids", "bound_organization_ids",
		"bound_instance_ids", "token_policies", "policies", "token_bound_cidrs",
	}
	for _, key := range setFields {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (r *CFAuthEngineRole) IsInitialized() bool {
	return true
}

func (r *CFAuthEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (d *CFAuthEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *CFAuthEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *CFAuthRole) toMap() map[string]any {
	payload := map[string]any{}
	payload["bound_application_ids"] = toInterfaceArray(r.BoundApplicationIDs)
	payload["bound_space_ids"] = toInterfaceArray(r.BoundSpaceIDs)
	payload["bound_organization_ids"] = toInterfaceArray(r.BoundOrganizationIDs)
	payload["bound_instance_ids"] = toInterfaceArray(r.BoundInstanceIDs)
	payload["disable_ip_matching"] = r.DisableIPMatching
	payload["token_ttl"] = durationToSeconds(r.TokenTTL)
	payload["token_max_ttl"] = durationToSeconds(r.TokenMaxTTL)
	payload["token_policies"] = toInterfaceArray(r.TokenPolicies)
	payload["policies"] = toInterfaceArray(r.Policies)
	payload["token_bound_cidrs"] = toInterfaceArray(r.TokenBoundCIDRs)
	payload["token_explicit_max_ttl"] = durationToSeconds(r.TokenExplicitMaxTTL)
	payload["token_no_default_policy"] = r.TokenNoDefaultPolicy
	payload["token_num_uses"] = json.Number(strconv.FormatInt(r.TokenNumUses, 10))
	payload["token_period"] = durationToSeconds(r.TokenPeriod)
	payload["token_type"] = r.TokenType
	return payload
}
