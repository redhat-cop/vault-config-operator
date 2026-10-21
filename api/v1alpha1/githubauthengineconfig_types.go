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

// GitHubAuthEngineConfigSpec defines the desired state of GitHubAuthEngineConfig
type GitHubAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the GitHub auth engine is mounted.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/auth/{spec.path}/config.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	GitHubAuthConfig `json:",inline"`
}

// GitHubAuthConfig contains the configuration for Vault's GitHub auth method
type GitHubAuthConfig struct {
	// Organization is the GitHub organization users must be part of.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Organization string `json:"organization" vault:"organization"`

	// OrganizationID is the numeric ID of the organization. Vault auto-fetches if not provided.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	OrganizationID int64 `json:"organizationID,omitempty" vault:"organization_id,number,omitempty"`

	// BaseURL is the API endpoint for GitHub Enterprise. Leave empty for public GitHub.
	// +kubebuilder:validation:Optional
	BaseURL string `json:"baseURL,omitempty" vault:"base_url,omitempty"`

	// TokenTTL is the incremental lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty" vault:"token_ttl,duration,omitempty"`

	// TokenMaxTTL is the maximum lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty" vault:"token_max_ttl,duration,omitempty"`

	// TokenPolicies are policies to encode onto generated tokens.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenPolicies []string `json:"tokenPolicies,omitempty" vault:"token_policies,omitempty"`

	// TokenBoundCIDRs are CIDR blocks that restrict authentication.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty" vault:"token_bound_cidrs,omitempty"`

	// TokenExplicitMaxTTL is the hard cap max TTL for tokens.
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty" vault:"token_explicit_max_ttl,duration,omitempty"`

	// TokenNoDefaultPolicy if true, omits the default policy from generated tokens.
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty" vault:"token_no_default_policy,omitempty"`

	// TokenNumUses is the max number of times a token may be used (0 = unlimited).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int64 `json:"tokenNumUses,omitempty" vault:"token_num_uses,number,omitempty"`

	// TokenPeriod is the maximum allowed period for periodic tokens.
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty" vault:"token_period,duration,omitempty"`

	// TokenType is the type of token to generate (service, batch, default).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch",""}
	TokenType string `json:"tokenType,omitempty" vault:"token_type,omitempty"`
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

func (r *GitHubAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func init() {
	SchemeBuilder.Register(&GitHubAuthEngineConfig{}, &GitHubAuthEngineConfigList{})
}

func (c *GitHubAuthConfig) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(c)
}
