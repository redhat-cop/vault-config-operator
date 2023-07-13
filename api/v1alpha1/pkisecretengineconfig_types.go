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
	"errors"
	"reflect"

	vault "github.com/hashicorp/vault/api"
	"github.com/redhat-cop/operator-utils/pkg/util/apis"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PKISecretEngineConfigSpec defines the desired state of PKISecretEngineConfig
type PKISecretEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	PKIType `json:",inline"`

	PKICommon `json:",inline"`

	PKIConfig `json:",inline"`

	PKIIntermediate `json:",inline"`
}

type PKIType struct {
	// Specifies the type of certificate authority. Root CA or Intermediate CA. This is part of the request URL.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"root","intermediate"}
	// +kubebuilder:default="root"
	Type string `json:"type,omitempty"`

	// Specifies the type of the root to create. If exported, the private key will be returned in the response; if internal the private key will not be returned and cannot be retrieved later. This is part of the request URL.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"internal","exported"}
	// +kubebuilder:default="internal"
	PrivateKeyType string `json:"privateKeyType,omitempty"`
}

type PKICommon struct {

	// Specifies the requested CN for the certificate.
	// +kubebuilder:validation:Required
	CommonName string `json:"commonName,omitempty"`

	// Specifies the requested Subject Alternative Names, in a comma-delimited list. These can be host names or email addresses; they will be parsed into their respective fields.
	// +kubebuilder:validation:Optional
	AltNames string `json:"altNames,omitempty"`

	// Specifies the requested IP Subject Alternative Names, in a comma-delimited list.
	// +kubebuilder:validation:Optional
	IPSans string `json:"IPSans,omitempty"`

	// Specifies the requested URI Subject Alternative Names, in a comma-delimited list.
	// +kubebuilder:validation:Optional
	URISans string `json:"URISans,omitempty"`

	// Specifies custom OID/UTF8-string SANs. These must match values specified on the role in allowed_other_sans (see role creation for allowed_other_sans globbing rules). The format is the same as OpenSSL: <oid>;<type>:<value> where the only current valid type is UTF8. This can be a comma-delimited list or a JSON string slice.
	// +kubebuilder:validation:Optional
	OtherSans string `json:"otherSans,omitempty"`

	// Specifies the requested Time To Live (after which the certificate will be expired). This cannot be larger than the engine's max (or, if not set, the system max).
	// +kubebuilder:validation:Optional
	TTL metav1.Duration `json:"TTL,omitempty"`

	// Specifies the format for returned data. Can be pem, der, or pem_bundle. If der, the output is base64 encoded. If pem_bundle, the certificate field will contain the private key (if exported) and certificate, concatenated; if the issuing CA is not a Vault-derived self-signed root, this will be included as well.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"pem","pem_bundle", "der"}
	// +kubebuilder:default="pem"
	Format string `json:"format,omitempty"`

	// Specifies the format for marshaling the private key. Defaults to der which will return either base64-encoded DER or PEM-encoded DER, depending on the value of format. The other option is pkcs8 which will return the key marshalled as PEM-encoded PKCS8.
	// +kubebuilder:validation:Optional
	PrivateKeyFormat string `json:"privateKeyFormat,omitempty"`

	// Specifies the desired key type; must be rsa or ec.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"rsa","ec"}
	// +kubebuilder:default="rsa"
	KeyType string `json:"keyType,omitempty"`

	// Specifies the number of bits to use. This must be changed to a valid value if the key_type is ec, e.g., 224, 256, 384 or 521.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=2048
	KeyBits int `json:"keyBits,omitempty"`

	// Specifies the maximum path length to encode in the generated certificate. -1 means no limit. Unless the signing certificate has a maximum path length set, in which case the path length is set to one less than that of the signing certificate. A limit of 0 means a literal path length of zero.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=-1
	MaxPathLength int `json:"maxPathLength,omitempty"`

	// If set, the given common_name will not be included in DNS or Email Subject Alternate Names (as appropriate). Useful if the CN is not a hostname or email address, but is instead some human-readable identifier.
	// +kubebuilder:validation:Optional
	ExcludeCnFromSans bool `json:"excludeCnFromSans,omitempty"`

	// A comma separated string (or, string array) containing DNS domains for which certificates are allowed to be issued or signed by this CA certificate. Note that subdomains are allowed, as per RFC.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	PermittedDnsDomains []string `json:"permittedDnsDomains,omitempty"`

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
}

type PKIConfig struct {
	// +kubebuilder:validation:Optional
	PKIConfigUrls `json:",inline"`
	// +kubebuilder:validation:Optional
	PKIConfigCRL `json:",inline"`
}

type PKIConfigUrls struct {
	// Specifies the URL values for the Issuing Certificate field. This can be an array or a comma-separated string list.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	IssuingCertificates []string `json:"issuingCertificates,omitempty"`

	// Specifies the URL values for the CRL Distribution Points field. This can be an array or a comma-separated string list.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	CRLDistributionPoints []string `json:"CRLDistributionPoints,omitempty"`

	// Specifies the URL values for the OCSP Servers field. This can be an array or a comma-separated string list.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	OcspServers []string `json:"ocspServers,omitempty"`
}

type PKIConfigCRL struct {
	// Specifies the time until expiration.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="72h"
	CRLExpiry metav1.Duration `json:"CRLExpiry,omitempty"`

	// Disables or enables CRL building.
	// +kubebuilder:validation:Optional
	CRLDisable bool `json:"CRLDisable,omitempty"`
}

type PKIIntermediate struct {
	// ExternalSignSecret retrieves the signed intermediate certificate from a Kubernetes secret. Allows submitting the signed CA certificate corresponding to a private key generated.
	// +kubebuilder:validation:Optional
	ExternalSignSecret *corev1.LocalObjectReference `json:"externalSignSecret,omitempty"`

	// CertificateKey key to be used when retrieving the signed certificate
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="tls.crt"
	CertificateKey string `json:"certificateKey,omitempty"`

	// Use the configured refered Vault PKISecretEngineConfig to issue a certificate with appropriate values for acting as an intermediate CA.
	// +kubebuilder:validation:Optional
	InternalSign *corev1.LocalObjectReference `json:"internalSign,omitempty"`

	cSR string `json:"-"`

	signedIntermediate string `json:"-"`
}

var _ vaultutils.VaultObject = &PKISecretEngineConfig{}
var _ vaultutils.VaultPKIEngineObject = &PKISecretEngineConfig{}

func (d *PKISecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (p *PKISecretEngineConfig) GetPath() string {
	return string(p.Spec.Path)
}

func (p *PKISecretEngineConfig) GetDeletePath() string {
	return string(p.Spec.Path) + "/root"
}

func (p *PKISecretEngineConfig) GetGeneratePath() string {
	return string(p.Spec.Path) + "/" + p.Spec.Type + "/generate/" + p.Spec.PrivateKeyType
}

func (p *PKISecretEngineConfig) GetConfigUrlsPath() string {
	return string(p.Spec.Path) + "/config/urls"
}

func (p *PKISecretEngineConfig) GetConfigCrlPath() string {
	return string(p.Spec.Path) + "/config/crl"
}

func (p *PKISecretEngineConfig) GetSignIntermediatePath() string {
	return string(p.Spec.InternalSign.Name) + "/root/sign-intermediate"
}

func (p *PKISecretEngineConfig) GetIntermediateSetSignedPath() string {
	return string(p.Spec.Path) + "/intermediate/set-signed"
}

func (p *PKISecretEngineConfig) GetGeneratedStatus() bool {
	return p.Status.Generated
}

func (p *PKISecretEngineConfig) SetGeneratedStatus(status bool) {
	p.Status.Generated = status
}

func (p *PKISecretEngineConfig) CreateExported(context context.Context, secret *vault.Secret) (bool, error) {
	log := log.FromContext(context)
	payload := p.GetExportedPayload(secret.Data)
	exported := p.Spec.PrivateKeyType == "exported"

	if exported {
		kubeClient := context.Value("kubeClient").(client.Client)
		kubeSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.Name,
				Namespace: p.Namespace,
				Labels: map[string]string{
					"redhatcop.redhat.io/pkisecretengineconfigs": p.Name,
				},
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(p, p.GroupVersionKind()),
				},
			},
			StringData: payload,
		}

		err := kubeClient.Create(context, kubeSecret)
		if err != nil {
			log.Error(err, "unable to create exported secret", "path", kubeSecret)
			return false, err
		}
	}

	if p.Spec.Type == "intermediate" {
		p.Spec.PKIIntermediate.cSR = payload["csr"]
	}

	return exported, nil
}

func (p *PKISecretEngineConfig) SetExportedStatus(status bool) {
	p.Status.Exported = status
}

func (p *PKISecretEngineConfig) SetIntermediate(context context.Context) error {

	if p.Spec.Type == "intermediate" {

		log := log.FromContext(context)
		vaultClient := context.Value("vaultClient").(*vault.Client)

		if p.Spec.InternalSign != nil && p.Spec.InternalSign.Name != "" {

			if p.Spec.PKIIntermediate.cSR == "" {
				kubeClient := context.Value("kubeClient").(client.Client)
				secret := &corev1.Secret{}

				err := kubeClient.Get(context, types.NamespacedName{
					Name:      p.Name,
					Namespace: p.Namespace,
				}, secret)

				if err != nil {
					log.Error(err, "unable to retrieve Exported CSR Secret", "instance", p)
					return err
				}
				p.Spec.PKIIntermediate.cSR = (string(secret.Data["csr"]))
			}

			secret, err := vaultClient.Logical().Write(p.GetSignIntermediatePath(), p.GetSignIntermediatePayload())
			if err != nil {
				log.Error(err, "unable to write object at", "path", p.GetIntermediateSetSignedPayload())
				return err
			}

			p.setSignedIntermediate(secret.Data["certificate"].(string))

		} else {

			if p.Spec.ExternalSignSecret == nil {
				err := errors.New("waiting spec.externalSignSecret with signed intermediate certificate")
				log.Error(err, "missing spec.externalSignSecret", "instance", p)
				return err
			}

			kubeClient := context.Value("kubeClient").(client.Client)
			secret := &corev1.Secret{}

			err := kubeClient.Get(context, types.NamespacedName{
				Namespace: p.Namespace,
				Name:      p.Spec.ExternalSignSecret.Name,
			}, secret)

			if err != nil {
				log.Error(err, "unable to retrieve Intermediate Secret for sign in", "instance", p)
				return err
			}

			p.setSignedIntermediate(string(secret.Data[p.Spec.CertificateKey]))

		}

		_, err := vaultClient.Logical().Write(p.GetIntermediateSetSignedPath(), p.GetIntermediateSetSignedPayload())
		if err != nil {
			log.Error(err, "unable to write object at", "path", p.GetIntermediateSetSignedPayload())
			return err
		}

	}
	return nil
}

func (p *PKISecretEngineConfig) GetSignedStatus() bool {
	if p.Spec.Type == "root" {
		return true
	}
	return p.Status.Signed
}

func (p *PKISecretEngineConfig) SetSignedStatus(status bool) {
	p.Status.Signed = status
}

func (p *PKISecretEngineConfig) GetExportedPayload(data map[string]interface{}) map[string]string {

	payload := map[string]string{}

	if p.Spec.Type == "root" {
		payload["issuing_ca"] = vaultutils.ToString(data["issuing_ca"])
		payload["expiration"] = data["expiration"].(json.Number).String()
		payload["certificate"] = vaultutils.ToString(data["certificate"])
		payload["serial_number"] = vaultutils.ToString(data["serial_number"])
	} else {
		payload["csr"] = vaultutils.ToString(data["csr"])
	}
	payload["private_key"] = vaultutils.ToString(data["private_key"])
	payload["private_key_type"] = vaultutils.ToString(data["private_key_type"])

	return payload
}

func (p *PKISecretEngineConfig) GetSignIntermediatePayload() map[string]interface{} {
	payload := p.GetPayload()

	payload["csr"] = p.Spec.PKIIntermediate.cSR
	return payload
}

func (p *PKISecretEngineConfig) GetIntermediateSetSignedPayload() map[string]interface{} {
	payload := map[string]interface{}{}

	payload["certificate"] = p.Spec.signedIntermediate

	return payload
}

func (p *PKISecretEngineConfig) GetPayload() map[string]interface{} {
	return p.Spec.PKICommon.toMap()
}

func (p *PKISecretEngineConfig) GetConfigUrlsPayload() map[string]interface{} {
	return p.Spec.PKIConfig.PKIConfigUrls.toMap()
}

func (p *PKISecretEngineConfig) GetConfigCrlPayload() map[string]interface{} {
	return p.Spec.PKIConfig.PKIConfigCRL.toMap()
}

func (p *PKISecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := p.Spec.PKICommon.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (p *PKISecretEngineConfig) IsInitialized() bool {
	return true
}

func (p *PKISecretEngineConfig) IsValid() (bool, error) {
	err := p.isValid()
	return err == nil, err
}

func (p *PKISecretEngineConfig) isValid() error {
	return nil
}
func (p *PKISecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}
func (p *PKISecretEngineConfig) setSignedIntermediate(signed string) {
	p.Spec.signedIntermediate = signed
}

// PKISecretEngineConfigStatus defines the observed state of PKISecretEngineConfig
type PKISecretEngineConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// +kubebuilder:validation:Optional
	Generated bool `json:"generated,omitempty"`

	// +kubebuilder:validation:Optional
	Exported bool `json:"exported,omitempty"`

	// +kubebuilder:validation:Optional
	Signed bool `json:"signed,omitempty"`
}

var _ apis.ConditionsAware = &PKISecretEngineConfig{}

func (m *PKISecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *PKISecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PKISecretEngineConfig is the Schema for the pkisecretengineconfigs API
type PKISecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PKISecretEngineConfigSpec   `json:"spec,omitempty"`
	Status PKISecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PKISecretEngineConfigList contains a list of PKISecretEngineConfig
type PKISecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PKISecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PKISecretEngineConfig{}, &PKISecretEngineConfigList{})
}

func (i *PKICommon) toMap() map[string]interface{} {
	payload := map[string]interface{}{}

	// payload["type"] = i.Type
	payload["common_name"] = i.CommonName
	payload["alt_names"] = i.AltNames
	payload["ip_sans"] = i.IPSans
	payload["uri_sans"] = i.URISans
	payload["other_sans"] = i.OtherSans
	payload["ttl"] = i.TTL
	payload["format"] = i.Format
	payload["private_key_format"] = i.PrivateKeyFormat
	payload["key_type"] = i.KeyType
	payload["key_bits"] = i.KeyBits
	payload["max_path_length"] = i.MaxPathLength
	payload["exclude_cn_from_sans"] = i.ExcludeCnFromSans
	payload["permitted_dns_domains"] = i.PermittedDnsDomains
	payload["ou"] = i.OU
	payload["organization"] = i.Organization
	payload["country"] = i.Country
	payload["locality"] = i.Locality
	payload["province"] = i.Province
	payload["street_address"] = i.StreetAddress
	payload["postal_code"] = i.PostalCode
	payload["serial_number"] = i.SerialNumber

	return payload
}

func (i *PKIConfigUrls) toMap() map[string]interface{} {
	payload := map[string]interface{}{}

	payload["issuing_certificates"] = i.IssuingCertificates
	payload["crl_distribution_points"] = i.CRLDistributionPoints
	payload["ocsp_servers"] = i.OcspServers

	return payload
}

func (i *PKIConfigCRL) toMap() map[string]interface{} {
	payload := map[string]interface{}{}

	payload["expiry"] = i.CRLExpiry
	payload["disable"] = i.CRLDisable

	return payload
}

func (d *PKISecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
