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

// OCIAuthEngineRoleSpec defines the desired state of OCIAuthEngineRole
type OCIAuthEngineRoleSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the OCI auth engine is mounted.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/role/{name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	OCIAuthRole `json:",inline"`
}

// OCIAuthEngineRoleStatus defines the observed state of OCIAuthEngineRole
type OCIAuthEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OCIAuthEngineRole is the Schema for the ociauthengineroles API
type OCIAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OCIAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status OCIAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OCIAuthEngineRoleList contains a list of OCIAuthEngineRole
type OCIAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OCIAuthEngineRole `json:"items"`
}

type OCIAuthRole struct {
	// Name is an optional Vault role-name override. If omitted, metadata.name is used.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`

	// OCIDList is a list of Group or Dynamic Group OCIDs that can take this role.
	// +kubebuilder:validation:Required
	// +listType=set
	OCIDList []string `json:"ocidList"`

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
	// +kubebuilder:validation:Minimum=0
	TokenPeriod int64 `json:"tokenPeriod,omitempty"`

	// TokenType is the type of token to generate.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
	TokenType string `json:"tokenType,omitempty"`
}

var _ vaultutils.VaultObject = &OCIAuthEngineRole{}
var _ vaultutils.ConditionsAware = &OCIAuthEngineRole{}

func init() {
	SchemeBuilder.Register(&OCIAuthEngineRole{}, &OCIAuthEngineRoleList{})
}

func (d *OCIAuthEngineRole) IsDeletable() bool {
	return true
}

func (r *OCIAuthEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *OCIAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (d *OCIAuthEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *OCIAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *OCIAuthEngineRole) GetPath() string {
	if r.Spec.Name != "" {
		return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/role/" + r.Spec.Name)
	}
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/role/" + r.Name)
}

func (r *OCIAuthEngineRole) GetPayload() map[string]any {
	return r.Spec.OCIAuthRole.toMap()
}

func (r *OCIAuthEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.OCIAuthRole.toMap()
	removeUnsetFields(desiredState, payload)
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	setFields := []string{"ocid_list", "token_policies", "policies", "token_bound_cidrs"}
	for _, key := range setFields {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (r *OCIAuthEngineRole) IsInitialized() bool {
	return true
}

func (r *OCIAuthEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (d *OCIAuthEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *OCIAuthEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *OCIAuthRole) toMap() map[string]any {
	payload := map[string]any{}
	payload["ocid_list"] = toInterfaceArray(r.OCIDList)
	payload["token_ttl"] = durationToSeconds(r.TokenTTL)
	payload["token_max_ttl"] = durationToSeconds(r.TokenMaxTTL)
	payload["token_policies"] = toInterfaceArray(r.TokenPolicies)
	payload["policies"] = toInterfaceArray(r.Policies)
	payload["token_bound_cidrs"] = toInterfaceArray(r.TokenBoundCIDRs)
	payload["token_explicit_max_ttl"] = durationToSeconds(r.TokenExplicitMaxTTL)
	payload["token_no_default_policy"] = r.TokenNoDefaultPolicy
	payload["token_num_uses"] = json.Number(strconv.FormatInt(r.TokenNumUses, 10))
	payload["token_period"] = json.Number(strconv.FormatInt(r.TokenPeriod, 10))
	payload["token_type"] = r.TokenType
	return payload
}
