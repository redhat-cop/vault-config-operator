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

// GitHubAuthEngineConfigSpec defines the desired state of GitHubAuthEngineConfig
type GitHubAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the GitHub auth engine is mounted.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/config.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	GitHubAuthConfig `json:",inline"`
}

// GitHubAuthConfig contains the configuration for Vault's GitHub auth method
type GitHubAuthConfig struct {
	// Organization is the GitHub organization users must be part of.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Organization string `json:"organization"`

	// OrganizationID is the numeric ID of the organization. Vault auto-fetches if not provided.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	OrganizationID int64 `json:"organizationID,omitempty"`

	// BaseURL is the API endpoint for GitHub Enterprise. Leave empty for public GitHub.
	// +kubebuilder:validation:Optional
	BaseURL string `json:"baseURL,omitempty"`

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

	// TokenBoundCIDRs are CIDR blocks that restrict authentication.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

	// TokenExplicitMaxTTL is the hard cap max TTL for tokens.
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

	// TokenNoDefaultPolicy if true, omits the default policy from generated tokens.
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`

	// TokenNumUses is the max number of times a token may be used (0 = unlimited).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int64 `json:"tokenNumUses,omitempty"`

	// TokenPeriod is the maximum allowed period for periodic tokens.
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty"`

	// TokenType is the type of token to generate (service, batch, default).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch",""}
	TokenType string `json:"tokenType,omitempty"`
}

// GitHubAuthEngineConfigStatus defines the observed state of GitHubAuthEngineConfig
type GitHubAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GitHubAuthEngineConfig is the Schema for the githubauthengineconfigs API
type GitHubAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitHubAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status GitHubAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitHubAuthEngineConfigList contains a list of GitHubAuthEngineConfig
type GitHubAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitHubAuthEngineConfig `json:"items"`
}

var _ vaultutils.VaultObject = &GitHubAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &GitHubAuthEngineConfig{}

func (d *GitHubAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *GitHubAuthEngineConfig) IsDeletable() bool {
	return false
}

func (r *GitHubAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *GitHubAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *GitHubAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config")
}

func (r *GitHubAuthEngineConfig) GetPayload() map[string]any {
	return r.Spec.GitHubAuthConfig.toMap()
}

func (r *GitHubAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.GitHubAuthConfig.toMap()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (r *GitHubAuthEngineConfig) IsInitialized() bool {
	return true
}

func (r *GitHubAuthEngineConfig) IsValid() (bool, error) {
	return true, nil
}

func (r *GitHubAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *GitHubAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *GitHubAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func init() {
	SchemeBuilder.Register(&GitHubAuthEngineConfig{}, &GitHubAuthEngineConfigList{})
}

func (c *GitHubAuthConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["organization"] = c.Organization
	if c.OrganizationID != 0 {
		payload["organization_id"] = json.Number(strconv.FormatInt(c.OrganizationID, 10))
	}
	if c.BaseURL != "" {
		payload["base_url"] = c.BaseURL
	}
	if c.TokenTTL != "" {
		payload["token_ttl"] = durationToSeconds(c.TokenTTL)
	}
	if c.TokenMaxTTL != "" {
		payload["token_max_ttl"] = durationToSeconds(c.TokenMaxTTL)
	}
	if len(c.TokenPolicies) > 0 {
		payload["token_policies"] = toInterfaceArray(c.TokenPolicies)
	}
	if len(c.TokenBoundCIDRs) > 0 {
		payload["token_bound_cidrs"] = toInterfaceArray(c.TokenBoundCIDRs)
	}
	if c.TokenExplicitMaxTTL != "" {
		payload["token_explicit_max_ttl"] = durationToSeconds(c.TokenExplicitMaxTTL)
	}
	if c.TokenNoDefaultPolicy {
		payload["token_no_default_policy"] = c.TokenNoDefaultPolicy
	}
	if c.TokenNumUses != 0 {
		payload["token_num_uses"] = json.Number(strconv.FormatInt(c.TokenNumUses, 10))
	}
	if c.TokenPeriod != "" {
		payload["token_period"] = durationToSeconds(c.TokenPeriod)
	}
	if c.TokenType != "" {
		payload["token_type"] = c.TokenType
	}
	return payload
}
