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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// JWTOIDCAuthEngineConfigSpec defines the desired state of JWTOIDCAuthEngineConfig
type JWTOIDCAuthEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	JWTOIDCConfig `json:",inline"`

	// OIDCCredentials from the provider for OIDC roles
	// OIDCCredentials consists in OIDCClientID and OIDCClientSecret, which can be created as Kubernetes Secret, VaultSecret or RandomSecret
	// +kubebuilder:validation:Optional
	OIDCCredentials vaultutils.RootCredentialConfig `json:"OIDCCredentials,omitempty"`
}

// JWTOIDCAuthEngineConfigStatus defines the observed state of JWTOIDCAuthEngineConfig
type JWTOIDCAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// JWTOIDCAuthEngineConfig is the Schema for the jwtoidcauthengineconfigs API
type JWTOIDCAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JWTOIDCAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status JWTOIDCAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JWTOIDCAuthEngineConfigList contains a list of JWTOIDCAuthEngineConfig
type JWTOIDCAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JWTOIDCAuthEngineConfig `json:"items"`
}

type JWTOIDCConfig struct {
	// The OIDC Discovery URL, without any .well-known component (base path). Cannot be used with "jwks_url" or "jwt_validation_pubkeys"
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	OIDCDiscoveryURL string `json:"OIDCDiscoveryURL"`

	// The CA certificate or chain of certificates, in PEM format, to use to validate connections to the OIDC Discovery URL.
	// If not set, system certificates are used
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	OIDCDiscoveryCAPEM string `json:"OIDCDiscoveryCAPEM,omitempty"`

	// The OAuth Client ID from the provider for OIDC roles.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	OIDCClientID string `json:"OIDCClientID,omitempty"`

	// The response mode to be used in the OAuth2 request.
	// Allowed values are "query" and "form_post". Defaults to "query".
	// If using Vault namespaces, and oidc_response_mode is "form_post", then "namespace_in_state" should be set to false
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	OIDCResponseMode string `json:"OIDCResponseMode,omitempty"`

	// The response types to request. Allowed values are "code" and "id_token". Defaults to "code".
	// Note: "id_token" may only be used if "oidc_response_mode" is set to "form_post"
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	OIDCResponseTypes []string `json:"OIDCResponseTypes,omitempty"`

	// JWKS URL to use to authenticate signatures.
	// Cannot be used with "oidc_discovery_url" or "jwt_validation_pubkeys"
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	JWKSURL string `json:"JWKSURL,omitempty"`

	// The CA certificate or chain of certificates, in PEM format, to use to validate connections to the JWKS URL.
	// If not set, system certificates are used.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	JWKSCAPEM string `json:"JWKSCAPEM,omitempty"`

	// A list of PEM-encoded public keys to use to authenticate signatures locally. Cannot be used with "jwks_url" or "oidc_discovery_url"
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	JWTValidationPubKeys []string `json:"JWTValidationPubKeys,omitempty"`

	// The value against which to match the iss claim in a JWT
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	BoundIssuer string `json:"boundIssuer,omitempty"`

	// A list of supported signing algorithms. Defaults to [RS256] for OIDC roles. Defaults to all available algorithms for JWT roles
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	JWTSupportedAlgs []string `json:"JWTSupportedAlgs,omitempty"`

	// The default role to use if none is provided during login
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	DefaultRole string `json:"defaultRole,omitempty"`

	// Configuration options for provider-specific handling. Providers with specific handling include: Azure, Google.
	// The options are described in each provider's section in OIDC Provider Setup
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	ProviderConfig *apiextensionsv1.JSON `json:"providerConfig,omitempty"`

	// Pass namespace in the OIDC state parameter instead of as a separate query parameter.
	// With this setting, the allowed redirect URL(s) in Vault and on the provider side should not contain a namespace query parameter.
	// This means only one redirect URL entry needs to be maintained on the provider side for all vault namespaces that will be authenticating against it.
	// Defaults to true for new configs
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	NamespaceInState bool `json:"namespaceInState,omitempty"`

	retrievedClientID string `json:"-"`

	retrievedClientPassword string `json:"-"`
}

var _ vaultutils.VaultObject = &JWTOIDCAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &JWTOIDCAuthEngineConfig{}

func (d *JWTOIDCAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *JWTOIDCAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *JWTOIDCAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *JWTOIDCAuthEngineConfig) SetUsernameAndPassword(OIDCClientID string, OIDCClientSecret string) {
	r.Spec.JWTOIDCConfig.retrievedClientID = OIDCClientID
	r.Spec.JWTOIDCConfig.retrievedClientPassword = OIDCClientSecret
}

func (r *JWTOIDCAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config")
}

func (r *JWTOIDCAuthEngineConfig) GetPayload() map[string]interface{} {
	return r.Spec.JWTOIDCConfig.toMap()
}

func (r *JWTOIDCAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.JWTOIDCConfig.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (r *JWTOIDCAuthEngineConfig) IsInitialized() bool {
	return true
}

func (r *JWTOIDCAuthEngineConfig) IsValid() (bool, error) {
	return true, nil
}

func (r *JWTOIDCAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *JWTOIDCAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {

	if reflect.DeepEqual(r.Spec.OIDCCredentials, vaultutils.RootCredentialConfig{PasswordKey: "password", UsernameKey: "username"}) {
		return nil
	}

	return r.setInternalCredentials(context)
}

func (r *JWTOIDCAuthEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
	if r.Spec.OIDCCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.OIDCCredentials.RandomSecret.Name,
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
		r.SetUsernameAndPassword(r.Spec.OIDCClientID, secret.Data[randomSecret.Spec.SecretKey].(string))
		return nil
	}
	if r.Spec.OIDCCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.OIDCCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		if r.Spec.OIDCClientID == "" {
			r.SetUsernameAndPassword(string(secret.Data[r.Spec.OIDCCredentials.UsernameKey]), string(secret.Data[r.Spec.OIDCCredentials.PasswordKey]))
		} else {
			r.SetUsernameAndPassword(r.Spec.JWTOIDCConfig.OIDCClientID, string(secret.Data[r.Spec.OIDCCredentials.PasswordKey]))
		}
		return nil
	}
	if r.Spec.OIDCCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.OIDCCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		if r.Spec.OIDCClientID == "" {
			r.SetUsernameAndPassword(secret.Data[r.Spec.OIDCCredentials.UsernameKey].(string), secret.Data[r.Spec.OIDCCredentials.PasswordKey].(string))
			log.V(1).Info("", "username", secret.Data[r.Spec.OIDCCredentials.UsernameKey].(string), "password", secret.Data[r.Spec.OIDCCredentials.PasswordKey].(string))
		} else {
			r.SetUsernameAndPassword(r.Spec.JWTOIDCConfig.OIDCClientID, secret.Data[r.Spec.OIDCCredentials.PasswordKey].(string))
			log.V(1).Info("", "username", r.Spec.JWTOIDCConfig.OIDCClientID, "password", secret.Data[r.Spec.OIDCCredentials.PasswordKey].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func init() {
	SchemeBuilder.Register(&JWTOIDCAuthEngineConfig{}, &JWTOIDCAuthEngineConfigList{})
}

func (i *JWTOIDCConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["oidc_discovery_url"] = i.OIDCDiscoveryURL
	payload["oidc_discovery_ca_pem"] = i.OIDCDiscoveryCAPEM
	payload["oidc_client_id"] = i.retrievedClientID
	payload["oidc_client_secret"] = i.retrievedClientPassword
	payload["oidc_response_mode"] = i.OIDCResponseMode
	payload["oidc_response_types"] = i.OIDCResponseTypes
	payload["jwks_url"] = i.JWKSURL
	payload["jwks_ca_pem"] = i.JWKSCAPEM
	payload["jwt_validation_pubkeys"] = i.JWTValidationPubKeys
	payload["bound_issuer"] = i.BoundIssuer
	payload["jwt_supported_algs"] = i.JWTSupportedAlgs
	payload["default_role"] = i.DefaultRole
	payload["provider_config"] = i.ProviderConfig
	payload["namespace_in_state"] = i.NamespaceInState

	return payload
}
