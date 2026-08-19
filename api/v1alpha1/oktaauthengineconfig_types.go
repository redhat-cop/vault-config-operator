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
	"fmt"
	"reflect"
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OktaAuthEngineConfigSpec defines the desired state of OktaAuthEngineConfig
type OktaAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the Okta auth engine is mounted.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	OktaAuthConfig `json:",inline"`

	// OktaCredentials is used to provide the Okta API token.
	// The API token can be sourced from a Kubernetes Secret, VaultSecret, or RandomSecret.
	// +kubebuilder:validation:Optional
	OktaCredentials vaultutils.RootCredentialConfig `json:"oktaCredentials,omitempty"`
}

type OktaAuthConfig struct {
	// OrgName is the name of the organization to be used in the Okta API.
	// +kubebuilder:validation:Required
	OrgName string `json:"orgName"`

	// BaseURL is the base domain for Okta API requests.
	// If unset, "okta.com" is used. Other valid values: "oktapreview.com", "okta-emea.com".
	// +kubebuilder:validation:Optional
	BaseURL string `json:"baseURL,omitempty"`

	// BypassOktaMFA bypasses an Okta MFA request. Useful when using Vault's built-in MFA.
	// +kubebuilder:validation:Optional
	BypassOktaMFA bool `json:"bypassOktaMFA,omitempty"`

	// TokenTTL is the incremental lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty"`

	// TokenMaxTTL is the maximum lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

	// TokenPolicies are policies to encode onto generated tokens.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenPolicies []string `json:"tokenPolicies,omitempty"`

	// TokenBoundCIDRs are CIDR blocks restricting authentication and tying the token.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

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

	retrievedAPIToken string `json:"-"`
}

// OktaAuthEngineConfigStatus defines the observed state of OktaAuthEngineConfig
type OktaAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OktaAuthEngineConfig is the Schema for the oktaauthengineconfigs API
type OktaAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OktaAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status OktaAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OktaAuthEngineConfigList contains a list of OktaAuthEngineConfig
type OktaAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OktaAuthEngineConfig `json:"items"`
}

var _ vaultutils.VaultObject = &OktaAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &OktaAuthEngineConfig{}

func (d *OktaAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *OktaAuthEngineConfig) IsDeletable() bool {
	return false
}

func (d *OktaAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/config")
}

func (d *OktaAuthEngineConfig) GetPayload() map[string]any {
	return d.Spec.OktaAuthConfig.toMap()
}

func (d *OktaAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.OktaAuthConfig.toMap()
	delete(desiredState, "api_token")
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *OktaAuthEngineConfig) IsInitialized() bool {
	return true
}

func (d *OktaAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *OktaAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *OktaAuthEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *OktaAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *OktaAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *OktaAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *OktaAuthEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if r.Spec.OktaCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.OktaCredentials.RandomSecret.Name,
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

		if randomSecret.Spec.IsKVSecretsEngineV2 {
			actualData, ok := secret.Data["data"].(map[string]any)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 response missing nested data map")
			}
			apiToken, ok := actualData[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.Spec.OktaAuthConfig.retrievedAPIToken = apiToken
		} else {
			apiToken, ok := secret.Data[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.Spec.OktaAuthConfig.retrievedAPIToken = apiToken
		}

		return nil
	}
	if r.Spec.OktaCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.OktaCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		passwordBytes, ok := secret.Data[r.Spec.OktaCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.OktaCredentials.Secret.Name, r.Spec.OktaCredentials.PasswordKey)
		}
		r.Spec.OktaAuthConfig.retrievedAPIToken = string(passwordBytes)
		return nil
	}
	if r.Spec.OktaCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.OktaCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		passwordVal, ok := secret.Data[r.Spec.OktaCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.OktaCredentials.PasswordKey)
		}
		r.Spec.OktaAuthConfig.retrievedAPIToken = passwordVal
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *OktaAuthEngineConfig) isValid() error {
	return r.Spec.OktaCredentials.ValidateCredentialSource()
}

func (c *OktaAuthConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["org_name"] = c.OrgName
	payload["api_token"] = c.retrievedAPIToken
	payload["base_url"] = c.BaseURL
	payload["bypass_okta_mfa"] = c.BypassOktaMFA
	payload["token_ttl"] = durationToSeconds(c.TokenTTL)
	payload["token_max_ttl"] = durationToSeconds(c.TokenMaxTTL)
	payload["token_policies"] = toInterfaceArray(c.TokenPolicies)
	payload["token_bound_cidrs"] = toInterfaceArray(c.TokenBoundCIDRs)
	payload["token_explicit_max_ttl"] = durationToSeconds(c.TokenExplicitMaxTTL)
	payload["token_no_default_policy"] = c.TokenNoDefaultPolicy
	payload["token_num_uses"] = json.Number(strconv.FormatInt(c.TokenNumUses, 10))
	payload["token_period"] = durationToSeconds(c.TokenPeriod)
	payload["token_type"] = c.TokenType
	return payload
}

func init() {
	SchemeBuilder.Register(&OktaAuthEngineConfig{}, &OktaAuthEngineConfigList{})
}
