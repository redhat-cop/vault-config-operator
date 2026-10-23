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
	"errors"
	"reflect"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// LDAPAuthEngineConfigSpec defines the desired state of LDAPAuthEngineConfig
type LDAPAuthEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/auth/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	LDAPConfig `json:",inline"`

	// BindCredentials is used to connect to the LDAP service on the specified LDAP Server.
	// BindCredentials consists in bindDN and bindPass, which can be created as Kubernetes Secret, VaultSecret or RandomSecret.
	// +kubebuilder:validation:Optional
	BindCredentials vaultutils.RootCredentialConfig `json:"bindCredentials,omitempty" vault:"-"`

	// CertificateConfig represents the LDAP service certificate configuration.
	// CertificateConfig consists in certificate, clientTLSCert and clientTLSKey which can be consumed from an Kubernetes Secret.
	// +kubebuilder:validation:Optional
	TLSConfig vaultutils.TLSConfig `json:"tLSConfig,omitempty" vault:"-"`
}

func (d *LDAPAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *LDAPAuthEngineConfig) IsDeletable() bool {
	return false
}

func (d *LDAPAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/config")
}

func (d *LDAPAuthEngineConfig) GetPayload() map[string]any {
	return d.Spec.LDAPConfig.toMap()
}
func (d *LDAPAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.LDAPConfig.toMap()
	delete(desiredState, "bindpass")
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

var _ vaultutils.VaultObject = &LDAPAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &LDAPAuthEngineConfig{}

func (d *LDAPAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *LDAPAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {

	if reflect.DeepEqual(d.Spec.TLSConfig, vaultutils.TLSConfig{TLSSecret: &corev1.LocalObjectReference{}}) {
		return nil
	}

	return d.setTLSConfig(context)
}

func (r *LDAPAuthEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *LDAPAuthEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if r.Spec.BindCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.BindCredentials.RandomSecret.Name,
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
		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		secretVal, ok := secretData[randomSecret.Spec.SecretKey].(string)
		if !ok {
			return errors.New("secret key not found or not a string in vault secret")
		}
		r.SetUsernameAndPassword(r.Spec.BindDN, secretVal)
		return nil
	}
	if r.Spec.BindCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.BindCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		if r.Spec.BindDN == "" {
			r.SetUsernameAndPassword(string(secret.Data[r.Spec.BindCredentials.UsernameKey]), string(secret.Data[r.Spec.BindCredentials.PasswordKey]))
		} else {
			r.SetUsernameAndPassword(r.Spec.LDAPConfig.BindDN, string(secret.Data[r.Spec.BindCredentials.PasswordKey]))
		}
		return nil
	}
	if r.Spec.BindCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.BindCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		passwordVal, ok := secretData[r.Spec.BindCredentials.PasswordKey].(string)
		if !ok {
			return errors.New("password key not found or not a string in vault secret")
		}
		if r.Spec.BindDN == "" {
			usernameVal, ok := secretData[r.Spec.BindCredentials.UsernameKey].(string)
			if !ok {
				return errors.New("username key not found or not a string in vault secret")
			}
			r.SetUsernameAndPassword(usernameVal, passwordVal)
		} else {
			r.SetUsernameAndPassword(r.Spec.LDAPConfig.BindDN, passwordVal)
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *LDAPAuthEngineConfig) setTLSConfig(context context.Context) error {
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

type LDAPConfig struct {

	// URL The LDAP server to connect to. Examples: ldap://ldap.myorg.com, ldaps://ldap.myorg.com:636.
	// Multiple URLs can be specified with commas, e.g. ldap://ldap.myorg.com,ldap://ldap2.myorg.com; these will be tried in-order.
	// +kubebuilder:validation:Required
	// +kubebuilder:default="ldap://127.0.0.1"
	URL string `json:"url" vault:"url"`

	// CaseSensitiveNames If set, user and group names assigned to policies within the backend will be case sensitive.
	// Otherwise, names will be normalized to lower case. Case will still be preserved when sending the username to the LDAP server at login time; this is only for matching local user/group definitions.
	// +kubebuilder:validation:Optional
	CaseSensitiveNames bool `json:"caseSensitiveNames,omitempty" vault:"case_sensitive_names"`

	// RequestTimeout Timeout, in seconds, for the connection when making requests against the server before returning back an error.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="90s"
	RequestTimeout string `json:"requestTimeout" vault:"request_timeout"`

	// StartTLS If true, issues a StartTLS command after establishing an unencrypted connection.
	// +kubebuilder:validation:Optional
	StartTLS bool `json:"startTLS,omitempty" vault:"starttls"`

	// TLSMinVersion Minimum TLS version to use. Accepted values are tls10, tls11, tls12 or tls13
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="tls12"
	// +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}
	TLSMinVersion string `json:"TLSMinVersion,omitempty" vault:"tls_min_version"`

	// TLSMaxVersion Maximum TLS version to use. Accepted values are tls10, tls11, tls12 or tls13
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="tls12"
	// +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}
	TLSMaxVersion string `json:"TLSMaxVersion,omitempty" vault:"tls_max_version"`

	// InsecureTLS If true, skips LDAP server SSL certificate verification - insecure, use with caution!
	// +kubebuilder:validation:Optional
	InsecureTLS bool `json:"insecureTLS,omitempty" vault:"insecure_tls"`

	// Certificate CA certificate to use when verifying LDAP server certificate, must be x509 PEM encoded.
	// +kubebuilder:validation:Optional
	Certificate string `json:"certificate,omitempty" vault:"certificate"`

	// ClientTLSCert Client certificate to provide to the LDAP server, must be x509 PEM encoded
	// +kubebuilder:validation:Optional
	ClientTLSCert string `json:"clientTLSCert,omitempty" vault:"client_tls_cert"`

	// ClientTLSKey Client certificate key to provide to the LDAP server, must be x509 PEM encoded
	// +kubebuilder:validation:Optional
	ClientTLSKey string `json:"clientTLSKey,omitempty" vault:"client_tls_key"`

	// BindDN - Username used to connect to the LDAP service on the specified LDAP Server.
	// If in the form accountname@domain.com, the username is transformed into a proper LDAP bind DN, for example, CN=accountname,CN=users,DC=domain,DC=com, when accessing the LDAP server.
	// If username is provided it takes precedence over the username retrieved from the referenced secrets
	// +kubebuilder:validation:Optional
	BindDN string `json:"bindDN,omitempty" vault:"binddn,omitempty"`

	// UserDN Base DN under which to perform user search. Example: ou=Users,dc=example,dc=com
	// +kubebuilder:validation:Optional
	UserDN string `json:"userDN,omitempty" vault:"userdn"`

	// UserAttr Attribute on user attribute object matching the username passed when authenticating. Examples: sAMAccountName, cn, uid
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="cn"
	UserAttr string `json:"userAttr" vault:"userattr"`

	// DiscoverDN Use anonymous bind to discover the bind DN of a user.
	// +kubebuilder:validation:Optional
	DiscoverDN bool `json:"discoverDN,omitempty" vault:"discoverdn"`

	// DenyNullBind This option prevents users from bypassing authentication when providing an empty password
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	DenyNullBind bool `json:"denyNullBind" vault:"deny_null_bind"`

	// UPNDomain  The userPrincipalDomain used to construct the UPN string for the authenticating user.
	// The constructed UPN will appear as [username]@UPNDomain. Example: example.com, which will cause vault to bind as username@example.com
	// +kubebuilder:validation:Optional
	UPNDomain string `json:"UPNDomain,omitempty" vault:"upndomain"`

	// UserFilter An optional LDAP user search filter. The template can access the following context variables: UserAttr, Username.
	// The default is ({{.UserAttr}}={{.Username}}), or ({{.UserAttr}}={{.Username@.upndomain}}) if upndomain is set.
	// +kubebuilder:validation:Optional
	UserFilter string `json:"userFilter,omitempty" vault:"userfilter"`

	// AnonymousGroupSearch Use anonymous binds when performing LDAP group searches (note: even when true, the initial credentials will still be used for the initial connection test).
	// +kubebuilder:validation:Optional
	AnonymousGroupSearch bool `json:"anonymousGroupSearch,omitempty" vault:"anonymous_group_search"`

	// GroupFilter Go template used when constructing the group membership query. The template can access the following context variables: [UserDN, Username].
	// The default is (|(memberUid={{.Username}})(member={{.UserDN}})(uniqueMember={{.UserDN}})), which is compatible with several common directory schemas.
	// To support nested group resolution for Active Directory, instead use the following query: (&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))
	// +kubebuilder:validation:Optional
	GroupFilter string `json:"groupFilter,omitempty" vault:"groupfilter"`

	// GroupDN LDAP search base to use for group membership search. This can be the root containing either groups or users. Example: ou=Groups,dc=example,dc=com
	// +kubebuilder:validation:Optional
	GroupDN string `json:"groupDN,omitempty" vault:"groupdn"`

	// GroupAttr LDAP attribute to follow on objects returned by groupfilter in order to enumerate user group membership.
	// Examples: for groupfilter queries returning group objects, use: cn. For queries returning user objects, use: memberOf. The default is cn.
	// +kubebuilder:validation:Optional
	GroupAttr string `json:"groupAttr,omitempty" vault:"groupattr"`

	// UsernameAsAlias If set to true, forces the auth method to use the username passed by the user as the alias name.
	// +kubebuilder:validation:Optional
	UsernameAsAlias bool `json:"usernameAsAlias,omitempty" vault:"username_as_alias"`

	// TokenTTL The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty" vault:"token_ttl,duration"`

	// TokenMaxTTL The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty" vault:"token_max_ttl,duration"`

	// TokenPolicies List of policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values.
	// +kubebuilder:validation:Optional
	TokenPolicies string `json:"tokenPolicies,omitempty" vault:"token_policies"`

	// TokenBoundCIDRs List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
	// +kubebuilder:validation:Optional
	TokenBoundCIDRs string `json:"tokenBoundCIDRs,omitempty" vault:"token_bound_cidrs"`

	// TonenExplicitMaxTTL If set, will encode an explicit max TTL onto the token. This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty" vault:"token_explicit_max_ttl,duration"`

	// TokenNoDefaultPolicy If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies.
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty" vault:"token_no_default_policy"`

	// TokenNumUses The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited.
	// If you require the token to have the ability to create child tokens, you will need to set this value to 0.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int64 `json:"tokenNumUses,omitempty" vault:"token_num_uses"`

	// TokenPeriod The period, if any, to set on the token
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenPeriod int64 `json:"tokenPeriod,omitempty" vault:"token_period,duration"`

	// The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens).
	// For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
	TokenType string `json:"tokenType,omitempty" vault:"token_type"`

	retrievedPassword string `json:"-"`

	retrievedUsername string `json:"-"`

	retrievedCertificate string `json:"-"`

	retrievedClientTLSCert string `json:"-"`

	retrievedClientTLSKey string `json:"-"`
}

// LDAPAuthEngineConfigStatus defines the observed state of LDAPAuthEngineConfig
type LDAPAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LDAPAuthEngineConfig is the Schema for the ldapauthengineconfigs API
type LDAPAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LDAPAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status LDAPAuthEngineConfigStatus `json:"status,omitempty"`
}

func (m *LDAPAuthEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *LDAPAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (m *LDAPAuthEngineConfig) SetUsernameAndPassword(bindDN string, bindPass string) {
	m.Spec.LDAPConfig.retrievedUsername = bindDN
	m.Spec.LDAPConfig.retrievedPassword = bindPass
}

//+kubebuilder:object:root=true

// LDAPAuthEngineConfigList contains a list of LDAPAuthEngineConfig
type LDAPAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LDAPAuthEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LDAPAuthEngineConfig{}, &LDAPAuthEngineConfigList{})
}

func (i *LDAPConfig) toMap() map[string]any {
	if i.Certificate != "" || i.ClientTLSCert != "" || i.ClientTLSKey != "" {
		i.retrievedCertificate = i.Certificate
		i.retrievedClientTLSCert = i.ClientTLSCert
		i.retrievedClientTLSKey = i.ClientTLSKey
	}
	payload := vaultutils.StructToVaultPayload(i)
	payload["certificate"] = i.retrievedCertificate
	payload["client_tls_cert"] = i.retrievedClientTLSCert
	payload["client_tls_key"] = i.retrievedClientTLSKey
	payload["binddn"] = i.retrievedUsername
	payload["bindpass"] = i.retrievedPassword
	return payload
}

func (r *LDAPAuthEngineConfig) isValid() error {
	return r.Spec.BindCredentials.ValidateCredentialSource()
}

func (d *LDAPAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
