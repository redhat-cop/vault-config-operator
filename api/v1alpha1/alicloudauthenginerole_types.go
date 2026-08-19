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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AliCloudAuthEngineRoleSpec defines the desired state of AliCloudAuthEngineRole
type AliCloudAuthEngineRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/role/{spec.name || metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	AliCloudAuthRole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type AliCloudAuthRole struct {
	// ARN is the AliCloud RAM role ARN. Must correspond with the name of the role reflected in the arn.
	// +kubebuilder:validation:Required
	ARN string `json:"arn"`

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

var _ vaultutils.VaultObject = &AliCloudAuthEngineRole{}
var _ vaultutils.ConditionsAware = &AliCloudAuthEngineRole{}

func (d *AliCloudAuthEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AliCloudAuthEngineRole) IsDeletable() bool {
	return true
}

func (d *AliCloudAuthEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Name)
}

func (d *AliCloudAuthEngineRole) GetPayload() map[string]any {
	return d.Spec.AliCloudAuthRole.toMap()
}

// aliCloudVaultReadAliases maps Vault read-response keys to the canonical write keys.
// Vault returns deprecated field names (without the token_ prefix) alongside the canonical versions.
var aliCloudVaultReadAliases = map[string]string{
	"ttl":         "token_ttl",
	"max_ttl":     "token_max_ttl",
	"period":      "token_period",
	"policies":    "token_policies",
	"bound_cidrs": "token_bound_cidrs",
}

func (d *AliCloudAuthEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.AliCloudAuthRole.toMap()
	normalizeVaultReadAliases(payload, aliCloudVaultReadAliases)
	removeUnsetFields(desiredState, payload)

	for _, boolKey := range []string{"token_no_default_policy"} {
		if boolVal, ok := desiredState[boolKey].(bool); ok && !boolVal {
			if _, inPayload := payload[boolKey]; !inPayload {
				delete(desiredState, boolKey)
			}
		}
	}

	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	setFields := []string{"token_policies", "token_bound_cidrs"}
	for _, key := range setFields {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (d *AliCloudAuthEngineRole) IsInitialized() bool {
	return true
}

func (d *AliCloudAuthEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *AliCloudAuthEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AliCloudAuthEngineRole) IsValid() (bool, error) {
	effectiveName := r.Spec.Name
	if effectiveName == "" {
		effectiveName = r.Name
	}

	arnRoleName, ok := extractAliCloudARNRoleName(r.Spec.ARN)
	if !ok {
		return false, fmt.Errorf(
			"spec.arn %q is not a valid AliCloud RAM role ARN (expected a role/ segment with a non-empty role name)",
			r.Spec.ARN,
		)
	}
	if !strings.EqualFold(arnRoleName, effectiveName) {
		return false, fmt.Errorf(
			"spec.arn role name %q does not match the effective Vault role name %q; "+
				"the Vault AliCloud auth API requires these to correspond",
			arnRoleName, effectiveName,
		)
	}
	return true, nil
}

// extractAliCloudARNRoleName extracts the role name from an AliCloud RAM role ARN.
// Typical form: acs:ram::<account>:role/<RoleName>
// The resource-type delimiter is a colon immediately before "role/", so we match
// ":role/" to avoid false positives on policy ARNs that contain "role/" in a path
// segment (e.g. acs:ram::...:policy/team/role/dev-role).
func extractAliCloudARNRoleName(arn string) (string, bool) {
	const delimiter = ":role/"
	idx := strings.LastIndex(arn, delimiter)
	if idx < 0 {
		return "", false
	}
	name := arn[idx+len(delimiter):]
	if name == "" {
		return "", false
	}
	return name, true
}

func (d *AliCloudAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *AliCloudAuthRole) toMap() map[string]any {
	payload := map[string]any{
		"arn":                     d.ARN,
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

// AliCloudAuthEngineRoleStatus defines the observed state of AliCloudAuthEngineRole
type AliCloudAuthEngineRoleStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *AliCloudAuthEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *AliCloudAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AliCloudAuthEngineRole is the Schema for the alicloudauthengineroles API
type AliCloudAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AliCloudAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status AliCloudAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AliCloudAuthEngineRoleList contains a list of AliCloudAuthEngineRole
type AliCloudAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AliCloudAuthEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AliCloudAuthEngineRole{}, &AliCloudAuthEngineRoleList{})
}
