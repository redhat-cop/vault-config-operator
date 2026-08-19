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
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// KerberosAuthEngineLDAPConfigSpec defines the desired state of KerberosAuthEngineLDAPConfig
type KerberosAuthEngineLDAPConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the Kerberos auth engine is mounted.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	KerberosLDAPConfig `json:",inline"`

	// LDAPCredentials is used to connect to the LDAP service.
	// Consists in bindDN and bindPass, sourced from a K8s Secret, VaultSecret, or RandomSecret.
	// +kubebuilder:validation:Required
	LDAPCredentials vaultutils.RootCredentialConfig `json:"ldapCredentials,omitempty"`

	// TLSConfig represents the LDAP service certificate configuration.
	// +kubebuilder:validation:Optional
	TLSConfig vaultutils.TLSConfig `json:"tLSConfig,omitempty"`
}

type KerberosLDAPConfig struct {
	// URL is the LDAP server to connect to. Multiple URLs can be comma-separated.
	// +kubebuilder:validation:Required
	URL string `json:"url"`

	// CaseSensitiveNames if set, user and group names are case sensitive for policy matching.
	// +kubebuilder:validation:Optional
	CaseSensitiveNames bool `json:"caseSensitiveNames,omitempty"`

	// StartTLS issues a StartTLS command after establishing an unencrypted connection.
	// +kubebuilder:validation:Optional
	StartTLS bool `json:"startTLS,omitempty"`

	// TLSMinVersion is the minimum TLS version to use.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="tls12"
	// +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}
	TLSMinVersion string `json:"tlsMinVersion"`

	// TLSMaxVersion is the maximum TLS version to use.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="tls12"
	// +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}
	TLSMaxVersion string `json:"tlsMaxVersion"`

	// InsecureTLS skips LDAP server SSL certificate verification.
	// +kubebuilder:validation:Optional
	InsecureTLS bool `json:"insecureTLS,omitempty"`

	// Certificate is the CA certificate for verifying the LDAP server certificate (x509 PEM).
	// +kubebuilder:validation:Optional
	Certificate string `json:"certificate,omitempty"`

	// ClientTLSCert is the client certificate to provide to the LDAP server (x509 PEM).
	// +kubebuilder:validation:Optional
	ClientTLSCert string `json:"clientTLSCert,omitempty"`

	// ClientTLSKey is the client certificate key to provide to the LDAP server (x509 PEM).
	// +kubebuilder:validation:Optional
	ClientTLSKey string `json:"clientTLSKey,omitempty"`

	// AliasMetadata configures alias metadata for the Kerberos auth mount.
	// +kubebuilder:validation:Optional
	AliasMetadata map[string]string `json:"aliasMetadata,omitempty"`

	// BindDN is the distinguished name used to bind when performing user search.
	// +kubebuilder:validation:Optional
	BindDN string `json:"bindDN,omitempty"`

	// UserDN is the base DN for user search.
	// +kubebuilder:validation:Optional
	UserDN string `json:"userDN,omitempty"`

	// UserAttr is the attribute on user objects matching the authenticating username.
	// +kubebuilder:validation:Optional
	UserAttr string `json:"userAttr,omitempty"`

	// DiscoverDN uses anonymous bind to discover the bind DN of a user.
	// +kubebuilder:validation:Optional
	DiscoverDN bool `json:"discoverDN,omitempty"`

	// DenyNullBind prevents users from bypassing authentication with an empty password.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	DenyNullBind bool `json:"denyNullBind"`

	// UPNDomain is the userPrincipalDomain for constructing UPN strings.
	// +kubebuilder:validation:Optional
	UPNDomain string `json:"upnDomain,omitempty"`

	// GroupFilter is a Go template for constructing the group membership query.
	// +kubebuilder:validation:Optional
	GroupFilter string `json:"groupFilter,omitempty"`

	// GroupDN is the LDAP search base for group membership search.
	// +kubebuilder:validation:Optional
	GroupDN string `json:"groupDN,omitempty"`

	// GroupAttr is the LDAP attribute for enumerating user group membership.
	// +kubebuilder:validation:Optional
	GroupAttr string `json:"groupAttr,omitempty"`

	// TokenTTL is the incremental lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty"`

	// TokenMaxTTL is the maximum lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

	// TokenPolicies are policies to encode onto generated tokens.
	// +kubebuilder:validation:Optional
	TokenPolicies string `json:"tokenPolicies,omitempty"`

	// TokenBoundCIDRs are CIDR blocks restricting authentication.
	// +kubebuilder:validation:Optional
	TokenBoundCIDRs string `json:"tokenBoundCIDRs,omitempty"`

	// TokenExplicitMaxTTL is the hard cap max TTL for tokens.
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

	// TokenNoDefaultPolicy if true, omits the default policy from generated tokens.
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`

	// TokenNumUses is the max number of times a generated token may be used (0 = unlimited).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int64 `json:"tokenNumUses,omitempty"`

	// TokenPeriod is the maximum allowed period for periodic tokens.
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty"`

	// TokenType is the type of token to generate.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
	TokenType string `json:"tokenType,omitempty"`

	retrievedPassword      string `json:"-"`
	retrievedUsername      string `json:"-"`
	retrievedCertificate   string `json:"-"`
	retrievedClientTLSCert string `json:"-"`
	retrievedClientTLSKey  string `json:"-"`
}

// KerberosAuthEngineLDAPConfigStatus defines the observed state of KerberosAuthEngineLDAPConfig
type KerberosAuthEngineLDAPConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KerberosAuthEngineLDAPConfig is the Schema for the kerberosauthengineldapconfigs API
type KerberosAuthEngineLDAPConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KerberosAuthEngineLDAPConfigSpec   `json:"spec,omitempty"`
	Status KerberosAuthEngineLDAPConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KerberosAuthEngineLDAPConfigList contains a list of KerberosAuthEngineLDAPConfig
type KerberosAuthEngineLDAPConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KerberosAuthEngineLDAPConfig `json:"items"`
}

var _ vaultutils.VaultObject = &KerberosAuthEngineLDAPConfig{}
var _ vaultutils.ConditionsAware = &KerberosAuthEngineLDAPConfig{}

func (d *KerberosAuthEngineLDAPConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *KerberosAuthEngineLDAPConfig) IsDeletable() bool {
	return false
}

func (d *KerberosAuthEngineLDAPConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/config/ldap")
}

func (d *KerberosAuthEngineLDAPConfig) GetPayload() map[string]any {
	return d.Spec.KerberosLDAPConfig.toMap()
}

func (d *KerberosAuthEngineLDAPConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.KerberosLDAPConfig.toMap()
	delete(desiredState, "bindpass")
	delete(desiredState, "client_tls_key")
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *KerberosAuthEngineLDAPConfig) IsInitialized() bool {
	return true
}

func (d *KerberosAuthEngineLDAPConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *KerberosAuthEngineLDAPConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	if reflect.DeepEqual(d.Spec.TLSConfig, vaultutils.TLSConfig{TLSSecret: &corev1.LocalObjectReference{}}) {
		return nil
	}
	return d.setTLSConfig(context)
}

func (r *KerberosAuthEngineLDAPConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *KerberosAuthEngineLDAPConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *KerberosAuthEngineLDAPConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *KerberosAuthEngineLDAPConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *KerberosAuthEngineLDAPConfig) SetUsernameAndPassword(bindDN string, bindPass string) {
	r.Spec.KerberosLDAPConfig.retrievedUsername = bindDN
	r.Spec.KerberosLDAPConfig.retrievedPassword = bindPass
}

func (r *KerberosAuthEngineLDAPConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if r.Spec.LDAPCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.LDAPCredentials.RandomSecret.Name,
		}, randomSecret)
		if err != nil {
			log.Error(err, "unable to retrieve RandomSecret", "instance", r)
			return err
		}
		secret, exists, err := vaultutils.ReadSecret(context, randomSecret.GetPath())
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		r.SetUsernameAndPassword(r.Spec.BindDN, secret.Data[randomSecret.Spec.SecretKey].(string))
		return nil
	}
	if r.Spec.LDAPCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.LDAPCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		if r.Spec.BindDN == "" {
			r.SetUsernameAndPassword(string(secret.Data[r.Spec.LDAPCredentials.UsernameKey]), string(secret.Data[r.Spec.LDAPCredentials.PasswordKey]))
		} else {
			r.SetUsernameAndPassword(r.Spec.KerberosLDAPConfig.BindDN, string(secret.Data[r.Spec.LDAPCredentials.PasswordKey]))
		}
		return nil
	}
	if r.Spec.LDAPCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.LDAPCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		if r.Spec.BindDN == "" {
			r.SetUsernameAndPassword(secret.Data[r.Spec.LDAPCredentials.UsernameKey].(string), secret.Data[r.Spec.LDAPCredentials.PasswordKey].(string))
		} else {
			r.SetUsernameAndPassword(r.Spec.KerberosLDAPConfig.BindDN, secret.Data[r.Spec.LDAPCredentials.PasswordKey].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *KerberosAuthEngineLDAPConfig) setTLSConfig(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)

	if r.Spec.TLSConfig.TLSSecret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.TLSConfig.TLSSecret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve TLS Secret", "instance", r)
			return err
		}
		r.Spec.retrievedCertificate = string(secret.Data["ca.crt"])
		r.Spec.retrievedClientTLSCert = string(secret.Data["tls.crt"])
		r.Spec.retrievedClientTLSKey = string(secret.Data["tls.key"])
		return nil
	} else if r.Spec.Certificate != "" || r.Spec.ClientTLSCert != "" || r.Spec.ClientTLSKey != "" {
		r.Spec.retrievedCertificate = r.Spec.Certificate
		r.Spec.retrievedClientTLSCert = r.Spec.ClientTLSCert
		r.Spec.retrievedClientTLSKey = r.Spec.ClientTLSKey
		return nil
	}

	return nil
}

func (r *KerberosAuthEngineLDAPConfig) isValid() error {
	return r.Spec.LDAPCredentials.ValidateCredentialSource()
}

func (c *KerberosLDAPConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["url"] = c.URL
	payload["case_sensitive_names"] = c.CaseSensitiveNames
	payload["starttls"] = c.StartTLS
	payload["tls_min_version"] = c.TLSMinVersion
	payload["tls_max_version"] = c.TLSMaxVersion
	payload["insecure_tls"] = c.InsecureTLS
	payload["certificate"] = c.retrievedCertificate
	payload["client_tls_cert"] = c.retrievedClientTLSCert
	payload["client_tls_key"] = c.retrievedClientTLSKey
	if len(c.AliasMetadata) > 0 {
		payload["alias_metadata"] = toAnyMapString(c.AliasMetadata)
	}
	payload["binddn"] = c.retrievedUsername
	payload["bindpass"] = c.retrievedPassword
	payload["userdn"] = c.UserDN
	payload["userattr"] = c.UserAttr
	payload["discoverdn"] = c.DiscoverDN
	payload["deny_null_bind"] = c.DenyNullBind
	payload["upndomain"] = c.UPNDomain
	payload["groupfilter"] = c.GroupFilter
	payload["groupdn"] = c.GroupDN
	payload["groupattr"] = c.GroupAttr
	payload["token_ttl"] = durationToSeconds(c.TokenTTL)
	payload["token_max_ttl"] = durationToSeconds(c.TokenMaxTTL)
	payload["token_policies"] = c.TokenPolicies
	payload["token_bound_cidrs"] = c.TokenBoundCIDRs
	payload["token_explicit_max_ttl"] = durationToSeconds(c.TokenExplicitMaxTTL)
	payload["token_no_default_policy"] = c.TokenNoDefaultPolicy
	payload["token_num_uses"] = json.Number(strconv.FormatInt(c.TokenNumUses, 10))
	payload["token_period"] = durationToSeconds(c.TokenPeriod)
	payload["token_type"] = c.TokenType
	return payload
}

func init() {
	SchemeBuilder.Register(&KerberosAuthEngineLDAPConfig{}, &KerberosAuthEngineLDAPConfigList{})
}
