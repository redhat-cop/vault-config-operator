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

// JWTOIDCAuthEngineRoleSpec defines the desired state of JWTOIDCAuthEngineRole
type JWTOIDCAuthEngineRoleSpec struct {

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

	JWTOIDCRole `json:",inline"`
}

type JWTOIDCRole struct {

	// Name of the role
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Type of role, either "oidc" (default) or "jwt"
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	RoleType string `json:"roleType,omitempty"`

	// List of aud claims to match against. Any match is sufficient. Required for "jwt" roles, optional for "oidc" roles
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	BoundAudiences []string `json:"boundAudiences,omitempty"`

	// The claim to use to uniquely identify the user; this will be used as the name for the Identity entity alias created due to a successful login.
	// The claim value must be a string
	// +kubebuilder:validation:Required
	UserClaim string `json:"userClaim"`

	// Specifies if the user_claim value uses JSON pointer syntax for referencing claims.
	// By default, the user_claim value will not use JSON pointer.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	UserClaimJSONPointer bool `json:"userClaimJSONPointer"`

	// The amount of leeway to add to all claims to account for clock skew, in seconds.
	// Defaults to 60 seconds if set to 0 and can be disabled if set to -1.
	// Accepts an integer number of seconds, or a Go duration format string. Only applicable with "jwt" roles
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	ClockSkewLeeway int64 `json:"clockSkewLeeway"`

	// The amount of leeway to add to expiration (exp) claims to account for clock skew, in seconds.
	// Defaults to 150 seconds if set to 0 and can be disabled if set to -1.
	// Accepts an integer number of seconds, or a Go duration format string. Only applicable with "jwt" roles.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	ExpirationLeeway int64 `json:"expirationLeeway"`

	// he amount of leeway to add to not before (nbf) claims to account for clock skew, in seconds
	// Defaults to 150 seconds if set to 0 and can be disabled if set to -1.
	// Accepts an integer number of seconds, or a Go duration format string. Only applicable with "jwt" roles
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	NotBeforeLeeway int64 `json:"notBeforeLeeway"`

	// If set, requires that the sub claim matches this value.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	BoundSubject string `json:"boundSubject,omitempty"`

	// If set, a map of claims (keys) to match against respective claim values (values)
	// The expected value may be a single string or a list of strings
	// The interpretation of the bound claim values is configured with bound_claims_type
	// Keys support JSON pointer syntax for referencing claims
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	BoundClaims map[string]string `json:"boundClaims,omitempty"`

	// Configures the interpretation of the bound_claims values.
	// If "string" (the default), the values will treated as string literals and must match exactly.
	// If set to "glob", the values will be interpreted as globs, with * matching any number of characters
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="string"
	BoundClaimsType string `json:"boundClaimsType,omitempty"`

	// The claim to use to uniquely identify the set of groups to which the user belongs; this will be used as the names for the Identity group aliases created due to a successful login.
	// The claim value must be a list of strings. Supports JSON pointer syntax for referencing claims
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	GroupsClaim string `json:"groupsClaim,omitempty"`

	// If set, a map of claims (keys) to be copied to specified metadata fields (values)
	// Keys support JSON pointer syntax for referencing claims
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	ClaimMappings map[string]string `json:"claimMappings,omitempty"`

	// If set, a list of OIDC scopes to be used with an OIDC role
	// The standard scope "openid" is automatically included and need not be specified
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	OIDCScopes []string `json:"OIDCScopes,omitempty"`

	// The list of allowed values for redirect_uri during OIDC logins
	// +kubebuilder:validation:Required
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AllowedRedirectURIs []string `json:"allowedRedirectURIs,omitempty"`

	// Log received OIDC tokens and claims when debug-level logging is active
	// Not recommended in production since sensitive information may be present in OIDC responses
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	VerboseOIDCLogging bool `json:"verboseOIDCLogging"`

	// Specifies the allowable elapsed time in seconds since the last time the user was actively authenticated with the OIDC provider
	// If set, the max_age request parameter will be included in the authentication request
	// See AuthRequest for additional details
	// Accepts an integer number of seconds, or a Go duration format string
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	MaxAge int64 `json:"maxage"`

	// The incremental lifetime for generated tokens
	// This current value of this will be referenced at renewal time
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenTTL string `json:"tokenTTL,omitempty"`

	// The maximum lifetime for generated tokens.
	// This current value of this will be referenced at renewal time
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

	// List of policies to encode onto generated tokens
	// Depending on the auth method, this list may be supplemented by user/group/other values
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	TokenPolicies []string `json:"tokenPolicies,omitempty"`

	// List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

	// If set, will encode an explicit max TTL onto the token.
	// This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

	// If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy"`

	// The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited.
	// If you require the token to have the ability to create child tokens, you will need to set this value to 0
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	TokenNumUses int64 `json:"tokenNumUses"`

	// The period, if any, to set on the token
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	TokenPeriod int64 `json:"tokenPeriod"`

	// The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens).
	// For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenType string `json:"tokenType,omitempty"`
}

// JWTOIDCAuthEngineRoleStatus defines the observed state of JWTOIDCAuthEngineRole
type JWTOIDCAuthEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (r *JWTOIDCAuthEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *JWTOIDCAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// JWTOIDCAuthEngineRole is the Schema for the jwtoidcauthengineroles API
type JWTOIDCAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JWTOIDCAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status JWTOIDCAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JWTOIDCAuthEngineRoleList contains a list of JWTOIDCAuthEngineRole
type JWTOIDCAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JWTOIDCAuthEngineRole `json:"items"`
}

var _ vaultutils.VaultObject = &JWTOIDCAuthEngineRole{}

func init() {
	SchemeBuilder.Register(&JWTOIDCAuthEngineRole{}, &JWTOIDCAuthEngineRoleList{})
}

func (d *JWTOIDCAuthEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *JWTOIDCAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *JWTOIDCAuthEngineRole) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/role/" + string(r.Spec.Name))
}

func (r *JWTOIDCAuthEngineRole) GetPayload() map[string]interface{} {
	return r.Spec.JWTOIDCRole.toMap()
}

func (r *JWTOIDCAuthEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.JWTOIDCRole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (r *JWTOIDCAuthEngineRole) IsInitialized() bool {
	return true
}

func (r *JWTOIDCAuthEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (d *JWTOIDCAuthEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *JWTOIDCRole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["name"] = r.Name
	payload["role_type"] = r.RoleType
	payload["bound_audiences"] = r.BoundAudiences
	payload["user_claim"] = r.UserClaim
	payload["user_claim_json_pointer"] = r.UserClaimJSONPointer
	payload["clock_skew_leeway"] = r.ClockSkewLeeway
	payload["expiration_leeway"] = r.ExpirationLeeway
	payload["not_before_leeway"] = r.NotBeforeLeeway
	payload["bound_subject"] = r.BoundSubject
	payload["bound_claims"] = r.BoundClaims
	payload["bound_claims_type"] = r.BoundClaimsType
	payload["groups_claim"] = r.GroupsClaim
	payload["claim_mappings"] = r.ClaimMappings
	payload["oidc_scopes"] = r.OIDCScopes
	payload["allowed_redirect_uris"] = r.AllowedRedirectURIs
	payload["verbose_oidc_logging"] = r.VerboseOIDCLogging
	payload["max_age"] = r.MaxAge
	payload["token_ttl"] = r.TokenTTL
	payload["token_max_ttl"] = r.TokenMaxTTL
	payload["token_policies"] = r.TokenPolicies
	payload["token_bound_cidrs"] = r.TokenBoundCIDRs
	payload["token_explicit_max_ttl"] = r.TokenExplicitMaxTTL
	payload["token_no_default_policy"] = r.TokenNoDefaultPolicy
	payload["token_num_uses"] = r.TokenNumUses
	payload["token_period"] = r.TokenPeriod
	payload["token_type"] = r.TokenType

	return payload
}
