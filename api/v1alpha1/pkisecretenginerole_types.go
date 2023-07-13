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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PKISecretEngineRoleSpec defines the desired state of PKISecretEngineRole
type PKISecretEngineRoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/roles/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	PKIRole `json:",inline"`
}

var _ vaultutils.VaultObject = &PKISecretEngineRole{}

func (d *PKISecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *PKISecretEngineRole) GetPath() string {
	return string(d.Spec.Path) + "/" + "roles" + "/" + d.Name
}
func (d *PKISecretEngineRole) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *PKISecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.PKIRole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (d *PKISecretEngineRole) IsInitialized() bool {
	return true
}

func (d *PKISecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *PKISecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

type PKIRole struct {

	// Specifies the Time To Live value provided as a string duration with time suffix. Hour is the largest suffix. If not set, uses the system default value or the value of max_ttl, whichever is shorter.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="0s"
	TTL metav1.Duration `json:"TTL,omitempty"`

	// Specifies the maximum Time To Live provided as a string duration with time suffix. Hour is the largest suffix. If not set, defaults to the system maximum lease TTL.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="0s"
	MaxTTL metav1.Duration `json:"maxTTL,omitempty"`

	// +kubebuilder:validation:Optional
	AllowLocalhost bool `json:"allowLocalhost,omitempty"`

	// Specifies the domains of the role. This is used with the allow_bare_domains and allow_subdomains options.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AllowedDomains []string `json:"allowedDomains,omitempty"`

	// When set, allowed_domains may contain templates, as with ACL Path Templating.
	// +kubebuilder:validation:Optional
	AllowedDomainsTemplate bool `json:"allowedDomainsTemplate,omitempty"`

	// Specifies if clients can request certificates matching the value of the actual domains themselves; e.g. if a configured domain set with allowed_domains is example.com, this allows clients to actually request a certificate containing the name example.com as one of the DNS values on the final certificate. In some scenarios, this can be considered a security risk.
	// +kubebuilder:validation:Optional
	AllowBareDomains bool `json:"allowBareDomains,omitempty"`

	// Specifies if clients can request certificates with CNs that are subdomains of the CNs allowed by the other role options. This includes wildcard subdomains. For example, an allowed_domains value of example.com with this option set to true will allow foo.example.com and bar.example.com as well as *.example.com. This is redundant when using the allow_any_name option.
	// +kubebuilder:validation:Optional
	AllowSubdomains bool `json:"allowSubdomains,omitempty"`

	// Allows names specified in allowed_domains to contain glob patterns (e.g. ftp*.example.com). Clients will be allowed to request certificates with names matching the glob patterns.
	// +kubebuilder:validation:Optional
	AllowGlobDomains bool `json:"allowGlobDomains,omitempty"`

	// Specifies if clients can request any CN. Useful in some circumstances, but make sure you understand whether it is appropriate for your installation before enabling it.
	// +kubebuilder:validation:Optional
	AllowAnyName bool `json:"allowAnyName,omitempty"`

	// Specifies if only valid host names are allowed for CNs, DNS SANs, and the host part of email addresses.
	// +kubebuilder:validation:Optional
	EnforceHostnames bool `json:"enforceHostnames,omitempty"`

	// Specifies if clients can request IP Subject Alternative Names. No authorization checking is performed except to verify that the given values are valid IP addresses.
	// +kubebuilder:validation:Optional
	AllowIPSans bool `json:"allowIPSans,omitempty"`

	// Defines allowed URI Subject Alternative Names. No authorization checking is performed except to verify that the given values are valid URIs. This can be a comma-delimited list or a JSON string slice. Values can contain glob patterns (e.g. spiffe://hostname/*).
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AllowedURISans []string `json:"allowedURISans,omitempty"`

	// Defines allowed custom OID/UTF8-string SANs. This can be a comma-delimited list or a JSON string slice, where each element has the same format as OpenSSL: <oid>;<type>:<value>, but the only valid type is UTF8 or UTF-8. The value part of an element may be a * to allow any value with that OID. Alternatively, specifying a single * will allow any other_sans input.
	// +kubebuilder:validation:Optional
	AllowedOtherSans string `json:"allowedOtherSans,omitempty"`

	// Specifies if certificates are flagged for server use.
	// +kubebuilder:validation:Optional
	ServerFlag bool `json:"serverFlag,omitempty"`

	// Specifies if certificates are flagged for client use.
	// +kubebuilder:validation:Optional
	ClientFlag bool `json:"clientFlag,omitempty"`

	// Specifies if certificates are flagged for code signing use.
	// +kubebuilder:validation:Optional
	CodeSigningFlag bool `json:"codeSigningFlag,omitempty"`

	// Specifies if certificates are flagged for email protection use.
	// +kubebuilder:validation:Optional
	EmailProtectionFlag bool `json:"emailProtectionFlag,omitempty"`

	// Specifies the type of key to generate for generated private keys and the type of key expected for submitted CSRs. Currently, rsa and ec are supported, or when signing CSRs any can be specified to allow keys of either type and with any bit size (subject to > 1024 bits for RSA keys).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"rsa","ec"}
	// +kubebuilder:default="rsa"
	KeyType string `json:"keyType,omitempty"`

	// Specifies the number of bits to use for the generated keys. This will need to be changed for ec keys, e.g., 224, 256, 384 or 521.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=2048
	KeyBits int `json:"keyBits,omitempty"`

	// Specifies the allowed key usage constraint on issued certificates. Valid values can be found at https://golang.org/pkg/crypto/x509/#KeyUsage - simply drop the KeyUsage part of the value. Values are not case-sensitive. To specify no key usage constraints, set this to an empty list.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:=DigitalSignature;KeyAgreement;KeyEncipherment;ContentCommitment;DataEncipherment;CertSign;CRLSign;EncipherOnly;DecipherOnly
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	KeyUsage []string `json:"keyUsage,omitempty"`

	// Specifies the allowed extended key usage constraint on issued certificates. Valid values can be found at https://golang.org/pkg/crypto/x509/#ExtKeyUsage - simply drop the ExtKeyUsage part of the value. Values are not case-sensitive. To specify no key usage constraints, set this to an empty list.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:=ServerAuth;ClientAuth;CodeSigning;EmailProtection;IPSECEndSystem;IPSECTunnel;IPSECUser;TimeStamping;OCSPSigning;MicrosoftServerGatedCrypto;NetscapeServerGatedCrypto;MicrosoftCommercialCodeSigning;MicrosoftKernelCodeSigning
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	ExtKeyUsage []string `json:"extKeyUsage,omitempty"`

	// A comma-separated string or list of extended key usage oids.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	ExtKeyUsageOids []string `json:"extKeyUsageOids,omitempty"`

	// When used with the CSR signing endpoint, the common name in the CSR will be used instead of taken from the JSON data. This does not include any requested SANs in the CSR; use use_csr_sans for that.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	UseCSRCommonName bool `json:"useCSRCommonName,omitempty"`

	// When used with the CSR signing endpoint, the subject alternate names in the CSR will be used instead of taken from the JSON data. This does not include the common name in the CSR; use use_csr_common_name for that.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	UseCSRSans bool `json:"useCSRSans,omitempty"`

	// Specifies the OU (OrganizationalUnit) values in the subject field of issued certificates. This is a comma-separated string or JSON array.
	// +kubebuilder:validation:Optional
	OU string `json:"ou,omitempty"`

	// Specifies the O (Organization) values in the subject field of issued certificates. This is a comma-separated string or JSON array.
	// +kubebuilder:validation:Optional
	Organization string `json:"organization,omitempty"`

	// Specifies the C (Country) values in the subject field of issued certificates. This is a comma-separated string or JSON array.
	// +kubebuilder:validation:Optional
	Country string `json:"country,omitempty"`

	// Specifies the L (Locality) values in the subject field of issued certificates. This is a comma-separated string or JSON array.
	// +kubebuilder:validation:Optional
	Locality string `json:"locality,omitempty"`

	// Specifies the ST (Province) values in the subject field of issued certificates. This is a comma-separated string or JSON array.
	// +kubebuilder:validation:Optional
	Province string `json:"province,omitempty"`

	// Specifies the Street Address values in the subject field of issued certificates. This is a comma-separated string or JSON array.
	// +kubebuilder:validation:Optional
	StreetAddress string `json:"streetAddress,omitempty"`

	// Specifies the Postal Code values in the subject field of issued certificates. This is a comma-separated string or JSON array.
	// +kubebuilder:validation:Optional
	PostalCode string `json:"postalCode,omitempty"`

	// Specifies the Serial Number, if any. Otherwise Vault will generate a random serial for you. If you want more than one, specify alternative names in the alt_names map using OID 2.5.4.5.
	// +kubebuilder:validation:Optional
	SerialNumber string `json:"serialNumber,omitempty"`

	// Specifies if certificates issued/signed against this role will have Vault leases attached to them. Certificates can be added to the CRL by vault revoke <lease_id> when certificates are associated with leases. It can also be done using the pki/revoke endpoint. However, when lease generation is disabled, invoking pki/revoke would be the only way to add the certificates to the CRL.
	// +kubebuilder:validation:Optional
	GenerateLease bool `json:"generateLease,omitempty"`

	// If set, certificates issued/signed against this role will not be stored in the storage backend. This can improve performance when issuing large numbers of certificates. However, certificates issued in this way cannot be enumerated or revoked, so this option is recommended only for certificates that are non-sensitive, or extremely short-lived. This option implies a value of false for generate_lease.
	// +kubebuilder:validation:Optional
	NoStore bool `json:"noStore,omitempty"`

	// If set to false, makes the common_name field optional while generating a certificate.
	// +kubebuilder:validation:Optional
	RequireCn bool `json:"requireCn,omitempty"`

	// A comma-separated string or list of policy OIDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	PolicyIdentifiers []string `json:"policyIdentifiers,omitempty"`

	// Mark Basic Constraints valid when issuing non-CA certificates.
	// +kubebuilder:validation:Optional
	BasicConstraintsValidForNonCa bool `json:"basicConstraintsValidForNonCa,omitempty"`

	// Specifies the duration by which to backdate the NotBefore property.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="30s"
	NotBeforeDuration metav1.Duration `json:"notBeforeDuration,omitempty"`
}

// PKISecretEngineRoleStatus defines the observed state of PKISecretEngineRole
type PKISecretEngineRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *PKISecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *PKISecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PKISecretEngineRole is the Schema for the pkisecretengineroles API
type PKISecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PKISecretEngineRoleSpec   `json:"spec,omitempty"`
	Status PKISecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PKISecretEngineRoleList contains a list of PKISecretEngineRole
type PKISecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PKISecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PKISecretEngineRole{}, &PKISecretEngineRoleList{})
}

func (i *PKIRole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["ttl"] = i.TTL
	payload["max_ttl"] = i.MaxTTL
	payload["allow_localhost"] = i.AllowLocalhost
	payload["allowed_domains"] = i.AllowedDomains
	payload["allowed_domains_template"] = i.AllowedDomainsTemplate
	payload["allow_bare_domains"] = i.AllowBareDomains
	payload["allow_subdomains"] = i.AllowSubdomains
	payload["allow_glob_domains"] = i.AllowGlobDomains
	payload["allow_any_name"] = i.AllowAnyName
	payload["enforce_hostnames"] = i.EnforceHostnames
	payload["allow_ip_sans"] = i.AllowIPSans
	payload["allowed_uri_sans"] = i.AllowedURISans
	payload["allowed_other_sans"] = i.AllowedOtherSans
	payload["server_flag"] = i.ServerFlag
	payload["client_flag"] = i.ClientFlag
	payload["code_signing_flag"] = i.CodeSigningFlag
	payload["email_protection_flag"] = i.EmailProtectionFlag
	payload["key_type"] = i.KeyType
	payload["key_bits"] = i.KeyBits
	payload["key_usage"] = i.KeyUsage
	payload["ext_key_usage"] = i.ExtKeyUsage
	payload["ext_key_usage_oids"] = i.ExtKeyUsageOids
	payload["use_csr_common_name"] = i.UseCSRCommonName
	payload["use_csr_sans"] = i.UseCSRSans
	payload["ou"] = i.OU
	payload["organization"] = i.Organization
	payload["country"] = i.Country
	payload["locality"] = i.Locality
	payload["province"] = i.Province
	payload["street_address"] = i.StreetAddress
	payload["postal_code"] = i.PostalCode
	payload["serial_number"] = i.SerialNumber
	payload["generate_lease"] = i.GenerateLease
	payload["no_store"] = i.NoStore
	payload["require_cn"] = i.RequireCn
	payload["policy_identifiers"] = i.PolicyIdentifiers
	payload["basic_constraints_valid_for_non_ca"] = i.BasicConstraintsValidForNonCa
	payload["not_before_duration"] = i.NotBeforeDuration
	return payload
}

func (d *PKISecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
