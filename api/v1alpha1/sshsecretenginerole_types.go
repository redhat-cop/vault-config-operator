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

var _ vaultutils.VaultObject = &SSHSecretEngineRole{}
var _ vaultutils.ConditionsAware = &SSHSecretEngineRole{}

// SSHSecretEngineRoleSpec defines the desired state of SSHSecretEngineRole
type SSHSecretEngineRoleSpec struct {

	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/roles/{metadata.name}.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	SSHSERole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type SSHSERole struct {
	// KeyType specifies the type of credentials generated. Must be "otp" or "ca".
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"otp","ca"}
	KeyType string `json:"keyType" vault:"key_type"`

	// DefaultUser specifies the default username. Required for OTP type.
	// +kubebuilder:validation:Optional
	DefaultUser string `json:"defaultUser,omitempty" vault:"default_user"`

	// DefaultUserTemplate if set, default_user can contain identity templates.
	// +kubebuilder:validation:Optional
	DefaultUserTemplate bool `json:"defaultUserTemplate,omitempty" vault:"default_user_template"`

	// AllowedUsers comma-separated list or "*" for any user.
	// +kubebuilder:validation:Optional
	AllowedUsers string `json:"allowedUsers,omitempty" vault:"allowed_users"`

	// AllowedUsersTemplate if set, allowed_users can contain identity templates.
	// +kubebuilder:validation:Optional
	AllowedUsersTemplate bool `json:"allowedUsersTemplate,omitempty" vault:"allowed_users_template"`

	// TTL for credentials. Uses duration format strings.
	// +kubebuilder:validation:Optional
	TTL string `json:"ttl,omitempty" vault:"ttl,duration"`

	// MaxTTL maximum TTL for credentials.
	// +kubebuilder:validation:Optional
	MaxTTL string `json:"maxTTL,omitempty" vault:"max_ttl,duration"`

	// Port SSH port number.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=22
	Port int `json:"port" vault:"port,number"`

	// --- OTP-only fields ---

	// CIDRList comma-separated CIDR blocks. Required for OTP unless zero-address.
	// +kubebuilder:validation:Optional
	CIDRList string `json:"cidrList,omitempty" vault:"cidr_list,when=KeyType:otp"`

	// ExcludeCIDRList comma-separated CIDR blocks to exclude.
	// +kubebuilder:validation:Optional
	ExcludeCIDRList string `json:"excludeCidrList,omitempty" vault:"exclude_cidr_list,when=KeyType:otp"`

	// --- CA-only fields ---

	// AllowedDomains comma-separated domains for host certificates.
	// +kubebuilder:validation:Optional
	AllowedDomains string `json:"allowedDomains,omitempty" vault:"allowed_domains,when=KeyType:ca"`

	// AllowedDomainsTemplate if set, allowed_domains can contain identity templates.
	// +kubebuilder:validation:Optional
	AllowedDomainsTemplate bool `json:"allowedDomainsTemplate,omitempty" vault:"allowed_domains_template,when=KeyType:ca"`

	// AllowUserCertificates if true, certificates can be signed for user use.
	// +kubebuilder:validation:Optional
	AllowUserCertificates bool `json:"allowUserCertificates,omitempty" vault:"allow_user_certificates,when=KeyType:ca"`

	// AllowHostCertificates if true, certificates can be signed for host use.
	// +kubebuilder:validation:Optional
	AllowHostCertificates bool `json:"allowHostCertificates,omitempty" vault:"allow_host_certificates,when=KeyType:ca"`

	// AllowBareDomains if true, host certs can use base domains from allowed_domains.
	// +kubebuilder:validation:Optional
	AllowBareDomains bool `json:"allowBareDomains,omitempty" vault:"allow_bare_domains,when=KeyType:ca"`

	// AllowSubdomains if true, host certs can use subdomains of allowed_domains.
	// +kubebuilder:validation:Optional
	AllowSubdomains bool `json:"allowSubdomains,omitempty" vault:"allow_subdomains,when=KeyType:ca"`

	// AllowUserKeyIDs if true, users can override the key ID.
	// +kubebuilder:validation:Optional
	AllowUserKeyIDs bool `json:"allowUserKeyIDs,omitempty" vault:"allow_user_key_ids,when=KeyType:ca"`

	// KeyIDFormat custom format for key id of signed certificate.
	// +kubebuilder:validation:Optional
	KeyIDFormat string `json:"keyIDFormat,omitempty" vault:"key_id_format,when=KeyType:ca"`

	// AllowedUserKeyLengths map of ssh key types to allowed lengths.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	AllowedUserKeyLengths map[string]int `json:"allowedUserKeyLengths,omitempty" vault:"allowed_user_key_lengths,intmap,when=KeyType:ca"`

	// AllowedCriticalOptions comma-separated list of critical options.
	// +kubebuilder:validation:Optional
	AllowedCriticalOptions string `json:"allowedCriticalOptions,omitempty" vault:"allowed_critical_options,when=KeyType:ca"`

	// AllowedExtensions comma-separated list of extensions or "*".
	// +kubebuilder:validation:Optional
	AllowedExtensions string `json:"allowedExtensions,omitempty" vault:"allowed_extensions,when=KeyType:ca"`

	// DefaultCriticalOptions map of default critical options.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	DefaultCriticalOptions map[string]string `json:"defaultCriticalOptions,omitempty" vault:"default_critical_options,stringmap,when=KeyType:ca"`

	// DefaultExtensions map of default extensions.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	DefaultExtensions map[string]string `json:"defaultExtensions,omitempty" vault:"default_extensions,stringmap,when=KeyType:ca"`

	// DefaultExtensionsTemplate if set, default_extensions can contain identity templates.
	// +kubebuilder:validation:Optional
	DefaultExtensionsTemplate bool `json:"defaultExtensionsTemplate,omitempty" vault:"default_extensions_template,when=KeyType:ca"`

	// AllowEmptyPrincipals allow signing certs with no valid principals.
	// +kubebuilder:validation:Optional
	AllowEmptyPrincipals bool `json:"allowEmptyPrincipals,omitempty" vault:"allow_empty_principals,when=KeyType:ca"`

	// AlgorithmSigner algorithm to sign keys with (ssh-rsa, rsa-sha2-256, rsa-sha2-512, default).
	// +kubebuilder:validation:Optional
	AlgorithmSigner string `json:"algorithmSigner,omitempty" vault:"algorithm_signer,when=KeyType:ca"`

	// NotBeforeDuration duration to backdate ValidAfter property.
	// +kubebuilder:validation:Optional
	NotBeforeDuration string `json:"notBeforeDuration,omitempty" vault:"not_before_duration,duration,when=KeyType:ca"`
}

func (i *SSHSERole) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(i)
}

// SSHSecretEngineRoleStatus defines the observed state of SSHSecretEngineRole
type SSHSecretEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *SSHSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *SSHSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SSHSecretEngineRole is the Schema for the sshsecretengineroles API
type SSHSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSHSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status SSHSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SSHSecretEngineRoleList contains a list of SSHSecretEngineRole
type SSHSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSHSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SSHSecretEngineRole{}, &SSHSecretEngineRoleList{})
}

func (d *SSHSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *SSHSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *SSHSecretEngineRole) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + vaultutils.ResourceName(d.Spec.Name, d.Name))
}

func (d *SSHSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *SSHSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
