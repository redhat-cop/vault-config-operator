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
	"reflect"

	"github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertAuthEngineRoleSpec defines the desired state of CertAuthEngineRole
type CertAuthEngineRoleSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/auth/{spec.path}/certs/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`

	CertAuthEngineRoleInternal `json:",inline"`
}

// CertAuthEngineRoleStatus defines the observed state of CertAuthEngineRole
type CertAuthEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CertAuthEngineRole is the Schema for the certauthengineroles API
type CertAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status CertAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CertAuthEngineRoleList contains a list of CertAuthEngineRole
type CertAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertAuthEngineRole `json:"items"`
}

type CertAuthEngineRoleInternal struct {
	// The PEM-format CA certificate.
	// +kubebuilder:validation:Optional
	Certificate string `json:"certificate,omitempty" vault:"certificate"`

	// Constrain the Common Names in the client certificate with a globbed pattern.
	// Value is a comma-separated list of patterns.
	// Authentication requires at least one Name matching at least one pattern. If not set, defaults to allowing all names.
	// +kubebuilder:validation:Optional
	AllowedCommonNames []string `json:"allowedCommonNames,omitempty" vault:"allowed_common_names"`

	// Constrain the Alternative Names in the client certificate with a globbed pattern.
	// Value is a comma-separated list of patterns.
	// Authentication requires at least one DNS matching at least one pattern. If not set, defaults to allowing all dns.
	// +kubebuilder:validation:Optional
	AllowedDNSSANs []string `json:"allowedDNSSANs,omitempty" vault:"allowed_dns_sans"`

	// Constrain the Alternative Names in the client certificate with a globbed pattern.
	// Value is a comma-separated list of patterns.
	// Authentication requires at least one Email matching at least one pattern. If not set, defaults to allowing all emails.
	// +kubebuilder:validation:Optional
	AllowedEmailSANs []string `json:"allowedEmailSANs,omitempty" vault:"allowed_email_sans"`

	// Constrain the Alternative Names in the client certificate with a globbed pattern.
	// Value is a comma-separated list of URI patterns.
	// Authentication requires at least one URI matching at least one pattern. If not set, defaults to allowing all URIs.
	// +kubebuilder:validation:Optional
	AllowedURISANs []string `json:"allowedURISANs,omitempty" vault:"allowed_uri_sans"`

	// Constrain the Organizational Units (OU) in the client certificate with a globbed pattern.
	// Value is a comma-separated list of OU patterns.
	// Authentication requires at least one OU matching at least one pattern. If not set, defaults to allowing all OUs.
	// +kubebuilder:validation:Optional
	AllowedOrganizationalUnits []string `json:"allowedOrganizationalUnits,omitempty" vault:"allowed_organizational_units"`

	// Require specific Custom Extension OIDs to exist and match the pattern.
	// Value is a comma separated string or array of oid:value.
	// Expects the extension value to be some type of ASN1 encoded string. All conditions must be met.
	// To match on the hex-encoded value of the extension, including non-string extensions, use the format hex:<oid>:<value>.
	// Supports globbing on value.
	// +kubebuilder:validation:Optional
	RequiredExtensions []string `json:"requiredExtensions,omitempty" vault:"required_extensions"`

	// A comma separated string or array of oid extensions.
	// Upon successful authentication, these extensions will be added as metadata if they are present in the certificate.
	// The metadata key will be the string consisting of the oid numbers separated by a dash (-) instead of a dot (.) to allow usage in ACL templates.
	// +kubebuilder:validation:Optional
	AllowedMetadataExtensions []string `json:"allowedMetadataExtensions,omitempty" vault:"allowed_metadata_extensions"`

	// If enabled, validate certificates' revocation status using OCSP.
	// +kubebuilder:validation:Optional
	OCSPEnabled bool `json:"ocspEnabled,omitempty" vault:"ocsp_enabled"`

	// Any additional OCSP responder certificates needed to verify OCSP responses.
	// Provided as base64 encoded PEM data.
	// +kubebuilder:validation:Optional
	OCSPCACertificates string `json:"ocspCACertificates,omitempty" vault:"ocsp_ca_certificates"`

	// A comma-separated list of OCSP server addresses.
	// If unset, the OCSP server is determined from the AuthorityInformationAccess extension on the certificate being inspected.
	// +kubebuilder:validation:Optional
	OCSPServersOverride []string `json:"ocspServersOverride,omitempty" vault:"ocsp_servers_override"`

	// If true and an OCSP response cannot be fetched or is of an unknown status, the login will proceed as if the certificate has not been revoked.
	// +kubebuilder:validation:Optional
	OCSPFailOpen bool `json:"ocspFailOpen,omitempty" vault:"ocsp_fail_open"`

	// If greater than 0, specifies the maximum age of an OCSP thisUpdate field.
	// This avoids accepting old responses without a nextUpdate field.
	// +kubebuilder:validation:Optional
	OCSPThisUpdateMaxAge string `json:"ocspThisUpdateMaxAge,omitempty" vault:"ocsp_this_update_max_age"`

	// The number of retries attempted before giving up on an OCSP request. 0 will disable retries.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=4
	// +kubebuilder:validation:Minimum=0
	OCSPMaxRetries int64 `json:"ocspMaxRetries" vault:"ocsp_max_retries"`

	// If set to true, rather than accepting the first successful OCSP response, query all servers and consider the certificate valid only if all servers agree.
	// +kubebuilder:validation:Optional
	OCSPQueryAllServers bool `json:"ocspQueryAllServers,omitempty" vault:"ocsp_query_all_servers"`

	// The display_name to set on tokens issued when authenticating against this CA certificate.
	// If not set, defaults to the name of the role.
	// +kubebuilder:validation:Optional
	DisplayName string `json:"displayName,omitempty" vault:"display_name"`

	// The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty" vault:"token_ttl,duration"`

	// The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty" vault:"token_max_ttl,duration"`

	// List of token policies to encode onto generated tokens.
	// Depending on the auth method, this list may be supplemented by user/group/other values.
	// +kubebuilder:validation:Optional
	TokenPolicies []string `json:"tokenPolicies,omitempty" vault:"token_policies"`

	// List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
	// +kubebuilder:validation:Optional
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty" vault:"token_bound_cidrs"`

	// If set, will encode an explicit max TTL onto the token.
	// This is a hard cap even if tokenTTL and tokenMaxTTL would otherwise allow a renewal.
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty" vault:"token_explicit_max_ttl,duration"`

	// If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in tokenPolicies.
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty" vault:"token_no_default_policy"`

	// The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited.
	// If you require the token to have the ability to create child tokens, you will need to set this value to 0.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int64 `json:"tokenNumUses,omitempty" vault:"token_num_uses"`

	// The maximum allowed period value when a periodic token is requested from this role.
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty" vault:"token_period,duration"`

	// The type of token that should be generated.
	// Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens).
	// For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time.
	// For machine based authentication cases, you should use batch type tokens.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
	TokenType string `json:"tokenType,omitempty" vault:"token_type"`
}

func (r *CertAuthEngineRoleInternal) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(r)
}

var _ vaultutils.VaultObject = &CertAuthEngineRole{}
var _ vaultutils.ConditionsAware = &CertAuthEngineRole{}

func (r *CertAuthEngineRole) GetPath() string {
	if r.Spec.Name != "" {
		return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/certs/" + r.Spec.Name)
	}

	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/certs/" + r.Name)
}

func (r *CertAuthEngineRole) GetPayload() map[string]any {
	return r.Spec.CertAuthEngineRoleInternal.toMap()
}

// IsEquivalentToDesiredState returns wether the passed payload is equivalent to the payload that the current object would generate. When this is a engine object the tune payload will be compared
func (r *CertAuthEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.CertAuthEngineRoleInternal.toMap()

	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (r *CertAuthEngineRole) IsDeletable() bool {
	return true
}

func (r *CertAuthEngineRole) GetKubeAuthConfiguration() *utils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *CertAuthEngineRole) GetVaultConnection() *utils.VaultConnection {
	return r.Spec.Connection
}

func (r *CertAuthEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *CertAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&CertAuthEngineRole{}, &CertAuthEngineRoleList{})
}
