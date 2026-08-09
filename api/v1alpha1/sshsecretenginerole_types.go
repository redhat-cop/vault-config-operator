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
	"time"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ vaultutils.VaultObject = &SSHSecretEngineRole{}
var _ vaultutils.ConditionsAware = &SSHSecretEngineRole{}

// SSHSecretEngineRoleSpec defines the desired state of SSHSecretEngineRole
type SSHSecretEngineRoleSpec struct {

	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/roles/{metadata.name}.
	// +kubebuilder:validation:Required
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
	KeyType string `json:"keyType"`

	// DefaultUser specifies the default username. Required for OTP type.
	// +kubebuilder:validation:Optional
	DefaultUser string `json:"defaultUser,omitempty"`

	// DefaultUserTemplate if set, default_user can contain identity templates.
	// +kubebuilder:validation:Optional
	DefaultUserTemplate bool `json:"defaultUserTemplate,omitempty"`

	// AllowedUsers comma-separated list or "*" for any user.
	// +kubebuilder:validation:Optional
	AllowedUsers string `json:"allowedUsers,omitempty"`

	// AllowedUsersTemplate if set, allowed_users can contain identity templates.
	// +kubebuilder:validation:Optional
	AllowedUsersTemplate bool `json:"allowedUsersTemplate,omitempty"`

	// TTL for credentials. Uses duration format strings.
	// +kubebuilder:validation:Optional
	TTL string `json:"ttl,omitempty"`

	// MaxTTL maximum TTL for credentials.
	// +kubebuilder:validation:Optional
	MaxTTL string `json:"maxTTL,omitempty"`

	// Port SSH port number.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=22
	Port int `json:"port"`

	// --- OTP-only fields ---

	// CIDRList comma-separated CIDR blocks. Required for OTP unless zero-address.
	// +kubebuilder:validation:Optional
	CIDRList string `json:"cidrList,omitempty"`

	// ExcludeCIDRList comma-separated CIDR blocks to exclude.
	// +kubebuilder:validation:Optional
	ExcludeCIDRList string `json:"excludeCidrList,omitempty"`

	// --- CA-only fields ---

	// AllowedDomains comma-separated domains for host certificates.
	// +kubebuilder:validation:Optional
	AllowedDomains string `json:"allowedDomains,omitempty"`

	// AllowedDomainsTemplate if set, allowed_domains can contain identity templates.
	// +kubebuilder:validation:Optional
	AllowedDomainsTemplate bool `json:"allowedDomainsTemplate,omitempty"`

	// AllowUserCertificates if true, certificates can be signed for user use.
	// +kubebuilder:validation:Optional
	AllowUserCertificates bool `json:"allowUserCertificates,omitempty"`

	// AllowHostCertificates if true, certificates can be signed for host use.
	// +kubebuilder:validation:Optional
	AllowHostCertificates bool `json:"allowHostCertificates,omitempty"`

	// AllowBareDomains if true, host certs can use base domains from allowed_domains.
	// +kubebuilder:validation:Optional
	AllowBareDomains bool `json:"allowBareDomains,omitempty"`

	// AllowSubdomains if true, host certs can use subdomains of allowed_domains.
	// +kubebuilder:validation:Optional
	AllowSubdomains bool `json:"allowSubdomains,omitempty"`

	// AllowUserKeyIDs if true, users can override the key ID.
	// +kubebuilder:validation:Optional
	AllowUserKeyIDs bool `json:"allowUserKeyIDs,omitempty"`

	// KeyIDFormat custom format for key id of signed certificate.
	// +kubebuilder:validation:Optional
	KeyIDFormat string `json:"keyIDFormat,omitempty"`

	// AllowedUserKeyLengths map of ssh key types to allowed lengths.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	AllowedUserKeyLengths map[string]int `json:"allowedUserKeyLengths,omitempty"`

	// AllowedCriticalOptions comma-separated list of critical options.
	// +kubebuilder:validation:Optional
	AllowedCriticalOptions string `json:"allowedCriticalOptions,omitempty"`

	// AllowedExtensions comma-separated list of extensions or "*".
	// +kubebuilder:validation:Optional
	AllowedExtensions string `json:"allowedExtensions,omitempty"`

	// DefaultCriticalOptions map of default critical options.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	DefaultCriticalOptions map[string]string `json:"defaultCriticalOptions,omitempty"`

	// DefaultExtensions map of default extensions.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	DefaultExtensions map[string]string `json:"defaultExtensions,omitempty"`

	// DefaultExtensionsTemplate if set, default_extensions can contain identity templates.
	// +kubebuilder:validation:Optional
	DefaultExtensionsTemplate bool `json:"defaultExtensionsTemplate,omitempty"`

	// AllowEmptyPrincipals allow signing certs with no valid principals.
	// +kubebuilder:validation:Optional
	AllowEmptyPrincipals bool `json:"allowEmptyPrincipals,omitempty"`

	// AlgorithmSigner algorithm to sign keys with (ssh-rsa, rsa-sha2-256, rsa-sha2-512, default).
	// +kubebuilder:validation:Optional
	AlgorithmSigner string `json:"algorithmSigner,omitempty"`

	// NotBeforeDuration duration to backdate ValidAfter property.
	// +kubebuilder:validation:Optional
	NotBeforeDuration string `json:"notBeforeDuration,omitempty"`
}

func (i *SSHSERole) toMap() map[string]any {
	payload := map[string]any{}
	payload["key_type"] = i.KeyType
	payload["default_user"] = i.DefaultUser
	payload["default_user_template"] = i.DefaultUserTemplate
	payload["allowed_users"] = i.AllowedUsers
	payload["allowed_users_template"] = i.AllowedUsersTemplate
	payload["ttl"] = durationToSeconds(i.TTL)
	payload["max_ttl"] = durationToSeconds(i.MaxTTL)
	payload["port"] = json.Number(strconv.Itoa(i.Port))

	if i.KeyType == "otp" {
		payload["cidr_list"] = i.CIDRList
		payload["exclude_cidr_list"] = i.ExcludeCIDRList
	}

	if i.KeyType == "ca" {
		payload["allowed_domains"] = i.AllowedDomains
		payload["allowed_domains_template"] = i.AllowedDomainsTemplate
		payload["allow_user_certificates"] = i.AllowUserCertificates
		payload["allow_host_certificates"] = i.AllowHostCertificates
		payload["allow_bare_domains"] = i.AllowBareDomains
		payload["allow_subdomains"] = i.AllowSubdomains
		payload["allow_user_key_ids"] = i.AllowUserKeyIDs
		payload["key_id_format"] = i.KeyIDFormat
		payload["allowed_user_key_lengths"] = toAnyMapJsonNumber(i.AllowedUserKeyLengths)
		payload["allowed_critical_options"] = i.AllowedCriticalOptions
		payload["allowed_extensions"] = i.AllowedExtensions
		payload["default_critical_options"] = toAnyMapString(i.DefaultCriticalOptions)
		payload["default_extensions"] = toAnyMapString(i.DefaultExtensions)
		payload["default_extensions_template"] = i.DefaultExtensionsTemplate
		payload["allow_empty_principals"] = i.AllowEmptyPrincipals
		payload["algorithm_signer"] = i.AlgorithmSigner
		payload["not_before_duration"] = durationToSeconds(i.NotBeforeDuration)
	}

	return payload
}

// durationToSeconds converts a Go duration string (e.g. "4h", "30s") to a
// json.Number representing integer seconds, matching the Vault API response
// format where durations are returned as numeric seconds.
func durationToSeconds(s string) json.Number {
	if s == "" {
		return json.Number("0")
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return json.Number(s)
	}
	return json.Number(strconv.FormatInt(int64(d.Seconds()), 10))
}

// toAnyMapJsonNumber converts map[string]int to map[string]any with json.Number
// values, matching the Vault API response where the Go client decodes JSON
// numbers via UseNumber(). Returns an empty map (not nil) for nil input,
// because Vault returns empty JSON objects {} for unset map fields.
func toAnyMapJsonNumber(m map[string]int) map[string]any {
	if len(m) == 0 {
		return map[string]any{}
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = json.Number(strconv.Itoa(v))
	}
	return result
}

// toAnyMapString converts map[string]string to map[string]any,
// matching the representation Vault returns after JSON deserialization.
// Returns an empty map (not nil) for nil input, because Vault returns
// empty JSON objects {} for unset map fields.
func toAnyMapString(m map[string]string) map[string]any {
	if len(m) == 0 {
		return map[string]any{}
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
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
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}

func (d *SSHSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *SSHSecretEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.SSHSERole.toMap()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *SSHSecretEngineRole) IsInitialized() bool {
	return true
}

func (d *SSHSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *SSHSecretEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *SSHSecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (d *SSHSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
