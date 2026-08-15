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

// AppRoleAuthEngineRoleSpec defines the desired state of AppRoleAuthEngineRole
type AppRoleAuthEngineRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/role/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	AppRoleRole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type AppRoleRole struct {
	// BindSecretID requires the secret_id to be presented during login
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	BindSecretID bool `json:"bindSecretID"`

	// SecretIDBoundCIDRs is the list of CIDR blocks for login operations
	// +kubebuilder:validation:Optional
	// +listType=set
	SecretIDBoundCIDRs []string `json:"secretIDBoundCIDRs,omitempty"`

	// SecretIDNumUses is the number of times a SecretID can be used (0=unlimited)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	SecretIDNumUses int `json:"secretIDNumUses,omitempty"`

	// SecretIDTTL is the duration after which a SecretID expires (e.g. "10m", "1h")
	// +kubebuilder:validation:Optional
	SecretIDTTL string `json:"secretIDTTL,omitempty"`

	// LocalSecretIDs makes SecretIDs cluster-local. Immutable after creation.
	// +kubebuilder:validation:Optional
	LocalSecretIDs bool `json:"localSecretIDs,omitempty"`

	// TokenTTL is the incremental lifetime for generated tokens (e.g. "1h")
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty"`

	// TokenMaxTTL is the maximum lifetime for generated tokens (e.g. "24h")
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

	// TokenPolicies is the list of token policies
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenPolicies []string `json:"tokenPolicies,omitempty"`

	// TokenBoundCIDRs is the list of CIDR blocks for token authentication
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

	// TokenExplicitMaxTTL is the hard cap max TTL (e.g. "24h")
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

	// TokenNoDefaultPolicy excludes the default policy from generated tokens
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`

	// TokenNumUses is the max number of token uses (0=unlimited)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int `json:"tokenNumUses,omitempty"`

	// TokenPeriod is the renewal period for periodic tokens (e.g. "24h")
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty"`

	// TokenType is the type of token to generate: service, batch, or default
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"service","batch","default","default-service","default-batch",""}
	TokenType string `json:"tokenType,omitempty"`
}

var _ vaultutils.VaultObject = &AppRoleAuthEngineRole{}
var _ vaultutils.ConditionsAware = &AppRoleAuthEngineRole{}
var _ vaultutils.CreateOnlyFieldsProvider = &AppRoleAuthEngineRole{}

func (d *AppRoleAuthEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AppRoleAuthEngineRole) IsDeletable() bool {
	return true
}

func (d *AppRoleAuthEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Name)
}

func (d *AppRoleAuthEngineRole) GetPayload() map[string]any {
	return d.Spec.AppRoleRole.toMap()
}

func (d *AppRoleAuthEngineRole) GetCreateOnlyFields() []string {
	return []string{"local_secret_ids"}
}

// appRoleVaultReadAliases maps Vault read-response keys to the canonical write keys.
// Vault returns "period" instead of "token_period" on AppRole role reads.
var appRoleVaultReadAliases = map[string]string{
	"period": "token_period",
}

func (d *AppRoleAuthEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.AppRoleRole.toMap()
	normalizeVaultReadAliases(payload, appRoleVaultReadAliases)
	removeUnsetFields(desiredState, payload)

	// local_secret_ids is create-only in Vault and immutable via webhook;
	// exclude it from drift comparison so updates only trigger on mutable fields.
	delete(desiredState, "local_secret_ids")
	delete(payload, "local_secret_ids")

	for _, boolKey := range []string{"token_no_default_policy"} {
		if boolVal, ok := desiredState[boolKey].(bool); ok && !boolVal {
			if _, inPayload := payload[boolKey]; !inPayload {
				delete(desiredState, boolKey)
			}
		}
	}
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	setFields := []string{"token_policies", "token_bound_cidrs", "secret_id_bound_cidrs"}
	for _, key := range setFields {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (d *AppRoleAuthEngineRole) IsInitialized() bool {
	return true
}

func (d *AppRoleAuthEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *AppRoleAuthEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AppRoleAuthEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (d *AppRoleAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *AppRoleRole) toMap() map[string]any {
	payload := map[string]any{
		"bind_secret_id":          d.BindSecretID,
		"secret_id_bound_cidrs":   toInterfaceArray(d.SecretIDBoundCIDRs),
		"secret_id_num_uses":      json.Number(strconv.Itoa(d.SecretIDNumUses)),
		"secret_id_ttl":           durationToSeconds(d.SecretIDTTL),
		"local_secret_ids":        d.LocalSecretIDs,
		"token_ttl":               durationToSeconds(d.TokenTTL),
		"token_max_ttl":           durationToSeconds(d.TokenMaxTTL),
		"token_policies":          toInterfaceArray(d.TokenPolicies),
		"token_bound_cidrs":       toInterfaceArray(d.TokenBoundCIDRs),
		"token_explicit_max_ttl":  durationToSeconds(d.TokenExplicitMaxTTL),
		"token_no_default_policy": d.TokenNoDefaultPolicy,
		"token_num_uses":          json.Number(strconv.Itoa(d.TokenNumUses)),
		"token_period":            durationToSeconds(d.TokenPeriod),
		"token_type":              d.TokenType,
	}
	return payload
}

// AppRoleAuthEngineRoleStatus defines the observed state of AppRoleAuthEngineRole
type AppRoleAuthEngineRoleStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *AppRoleAuthEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *AppRoleAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AppRoleAuthEngineRole is the Schema for the approleauthengineroles API
type AppRoleAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppRoleAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status AppRoleAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AppRoleAuthEngineRoleList contains a list of AppRoleAuthEngineRole
type AppRoleAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppRoleAuthEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppRoleAuthEngineRole{}, &AppRoleAuthEngineRoleList{})
}
