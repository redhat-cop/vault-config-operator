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
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path will be {[spec.authentication.namespace]}/auth/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path Path `json:"path,omitempty"`

	LDAPConfig `json:",inline"`

	// BindCredentials used to connect to the LDAP service on the specified LDAP Server
	// BindCredentials consists in bindDN and bindPass, which can be created as Kubernetes Secret, VaultSecret or RandomSecret
	// +kubebuilder:validation:Required
	BindCredentials RootCredentialConfig `json:"bindCredentials,omitempty"`
}

func (d *LDAPAuthEngineConfig) GetPath() string {
	return cleansePath("auth/" + string(d.Spec.Path) + "/config")
}

func (d *LDAPAuthEngineConfig) GetPayload() map[string]interface{} {
	return d.Spec.LDAPConfig.toMap()
}
func (d *LDAPAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.LDAPConfig.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

var _ vaultutils.VaultObject = &LDAPAuthEngineConfig{}

func (d *LDAPAuthEngineConfig) IsInitialized() bool {
	return true
}

func (d *LDAPAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (r *LDAPAuthEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *LDAPAuthEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
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
		r.SetUsernameAndPassword(r.Spec.BindDN, secret.Data[randomSecret.Spec.SecretKey].(string))
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
			r.SetUsernameAndPassword(r.Spec.BindDN, string(secret.Data[r.Spec.BindCredentials.PasswordKey]))
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
		if r.Spec.BindDN == "" {
			r.SetUsernameAndPassword(secret.Data[r.Spec.BindCredentials.UsernameKey].(string), secret.Data[r.Spec.BindCredentials.PasswordKey].(string))
			log.V(1).Info("", "username", secret.Data[r.Spec.BindCredentials.UsernameKey].(string), "password", secret.Data[r.Spec.BindCredentials.PasswordKey].(string))
		} else {
			r.SetUsernameAndPassword(r.Spec.BindDN, secret.Data[r.Spec.BindCredentials.PasswordKey].(string))
			log.V(1).Info("", "username", r.Spec.BindDN, "password", secret.Data[r.Spec.BindCredentials.PasswordKey].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

type LDAPConfig struct {

	// URL The LDAP server to connect to. Examples: ldap://ldap.myorg.com, ldaps://ldap.myorg.com:636.
	// Multiple URLs can be specified with commas, e.g. ldap://ldap.myorg.com,ldap://ldap2.myorg.com; these will be tried in-order.
	// +kubebuilder:validation:Required
	// +kubebuilder:default="ldap://127.0.0.1"
	URL string `json:"url"`

	// CaseSensitiveNames If set, user and group names assigned to policies within the backend will be case sensitive.
	// Otherwise, names will be normalized to lower case. Case will still be preserved when sending the username to the LDAP server at login time; this is only for matching local user/group definitions.
	// +kubebuilder:validation:Required
	// +kubebuilder:default=false
	CaseSensitiveNames bool `json:"caseSensitiveNames"`

	// RequestTimeout Timeout, in seconds, for the connection when making requests against the server before returning back an error.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="90s"
	RequestTimeout string `json:"requestTimeout"`

	// StartTls If true, issues a StartTLS command after establishing an unencrypted connection.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	StartTls bool `json:"startTls"`

	// TlsMinVersion Minimum TLS version to use. Accepted values are tls10, tls11, tls12 or tls13
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="tls12"
	TlsMinVersion string `json:"tlsMinVersion"`

	// TlsMaxVersion Maximum TLS version to use. Accepted values are tls10, tls11, tls12 or tls13
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="tls12"
	TlsMaxVersion string `json:"tlsMaxVersion"`

	// InsecureTls If true, skips LDAP server SSL certificate verification - insecure, use with caution!
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	InsecureTls bool `json:"insecureTls"`

	// Certificate CA certificate to use when verifying LDAP server certificate, must be x509 PEM encoded.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	Certificate string `json:"certificate,omitempty"`

	// ClientTlsCert Client certificate to provide to the LDAP server, must be x509 PEM encoded
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	ClientTlsCert string `json:"clientTlsCert,omitempty"`

	// ClientTlsKey Client certificate key to provide to the LDAP server, must be x509 PEM encoded
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	ClientTlsKey string `json:"clientTlsKey,omitempty"`

	// BindDN - Username used to connect to the LDAP service on the specified LDAP Server.
	// If in the form accountname@domain.com, the username is transformed into a proper LDAP bind DN, for example, CN=accountname,CN=users,DC=domain,DC=com, when accessing the LDAP server.
	// If username is provided it takes precedence over the username retrieved from the referenced secrets
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	BindDN string `json:"bindDN,omitempty"`

	// BindPass Password to use along with binddn when performing user search.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	BindPass string `json:"bindPass,omitempty"`

	// UserDN Base DN under which to perform user search. Example: ou=Users,dc=example,dc=com
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	UserDN string `json:"userDN,omitempty"`

	// UserAttr Attribute on user attribute object matching the username passed when authenticating. Examples: sAMAccountName, cn, uid
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="cn"
	UserAttr string `json:"userAttr"`

	// DiscoverDN Use anonymous bind to discover the bind DN of a user.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	DiscoverDN bool `json:"discoverDN"`

	// DenyNullBind This option prevents users from bypassing authentication when providing an empty password
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	DenyNullBind bool `json:"denyNullBind"`

	// UPNDomain  The userPrincipalDomain used to construct the UPN string for the authenticating user.
	// The constructed UPN will appear as [username]@UPNDomain. Example: example.com, which will cause vault to bind as username@example.com
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	UPNDomain string `json:"UPNDomain,omitempty"`

	// UserFilter An optional LDAP user search filter. The template can access the following context variables: UserAttr, Username.
	// The default is ({{.UserAttr}}={{.Username}}), or ({{.UserAttr}}={{.Username@.upndomain}}) if upndomain is set.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	UserFilter string `json:"userFilter,omitempty"`

	// AnonymousGroupSearch Use anonymous binds when performing LDAP group searches (note: even when true, the initial credentials will still be used for the initial connection test).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	AnonymousGroupSearch bool `json:"anonymousGroupSearch"`

	// GroupFilter Go template used when constructing the group membership query. The template can access the following context variables: [UserDN, Username].
	// The default is (|(memberUid={{.Username}})(member={{.UserDN}})(uniqueMember={{.UserDN}})), which is compatible with several common directory schemas.
	// To support nested group resolution for Active Directory, instead use the following query: (&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	GroupFilter string `json:"groupFilter,omitempty"`

	// GroupDN LDAP search base to use for group membership search. This can be the root containing either groups or users. Example: ou=Groups,dc=example,dc=com
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	GroupDN string `json:"groupDN,omitempty"`

	// GroupAttr LDAP attribute to follow on objects returned by groupfilter in order to enumerate user group membership.
	// Examples: for groupfilter queries returning group objects, use: cn. For queries returning user objects, use: memberOf. The default is cn.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	GroupAttr string `json:"groupAttr,omitempty"`

	// UsernameAsAlias If set to true, forces the auth method to use the username passed by the user as the alias name.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	UsernameAsAlias bool `json:"usernameAsAlias"`

	// TokenTtl The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenTtl string `json:"tokenTtl,omitempty"`

	// TokenMaxTtl The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenMaxTtl string `json:"tokenMaxTtl,omitempty"`

	// TokenPolicies List of policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenPolicies string `json:"tokenPolicies,omitempty"`

	// TokenBoundCIDRs List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenBoundCIDRs string `json:"tokenBoundCIDRs,omitempty"`

	// TokenExplicitMaxTtl If set, will encode an explicit max TTL onto the token. This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenExplicitMaxTtl string `json:"tokenExplicitMaxTtl,omitempty"`

	// TokenNoDefaultPolicy If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy"`

	// TokenNumUses The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited.
	// If you require the token to have the ability to create child tokens, you will need to set this value to 0.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	TokenNumUses int64 `json:"tokenNumUses"`

	// TokenPeriod The period, if any, to set on the token
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	TokenPeriod int64 `json:"tokenPeriod"`

	// The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens).
	// For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	TokenType string `json:"tokenType,omitempty"`

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
	m.Spec.LDAPConfig.BindDN = bindDN
	m.Spec.LDAPConfig.BindPass = bindPass
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

func (i *LDAPConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["url"] = i.URL
	payload["case_sensitive_names"] = i.CaseSensitiveNames
	payload["request_timeout"] = i.RequestTimeout
	payload["starttls"] = i.StartTls
	payload["tls_min_version"] = i.TlsMinVersion
	payload["tls_max_version"] = i.TlsMaxVersion
	payload["insecure_tls"] = i.InsecureTls
	payload["certificate"] = i.Certificate
	payload["client_tls_cert"] = i.ClientTlsCert
	payload["client_tls_key"] = i.ClientTlsKey
	payload["binddn"] = i.BindDN
	payload["bindpass"] = i.BindPass
	payload["userdn"] = i.UserDN
	payload["userattr"] = i.UserAttr
	payload["discoverdn"] = i.DiscoverDN
	payload["deny_null_bind"] = i.DenyNullBind
	payload["upndomain"] = i.UPNDomain
	payload["userfilter"] = i.UserFilter
	payload["anonymous_group_search"] = i.AnonymousGroupSearch
	payload["groupfilter"] = i.GroupFilter
	payload["groupdn"] = i.GroupDN
	payload["groupattr"] = i.GroupAttr
	payload["username_as_alias"] = i.UsernameAsAlias
	payload["token_ttl"] = i.TokenTtl
	payload["token_max_ttl"] = i.TokenMaxTtl
	payload["token_policies"] = i.TokenPolicies
	payload["token_bound_cidrs"] = i.TokenBoundCIDRs
	payload["token_explicit_max_ttl"] = i.TokenExplicitMaxTtl
	payload["token_no_default_policy"] = i.TokenNoDefaultPolicy
	payload["token_num_uses"] = i.TokenNumUses
	payload["token_period"] = i.TokenPeriod
	payload["token_type"] = i.TokenType

	return payload
}

func (r *LDAPAuthEngineConfig) isValid() error {
	return r.Spec.BindCredentials.validateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}
