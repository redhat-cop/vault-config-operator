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
	"fmt"
	"reflect"

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
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the Okta auth engine is mounted.
	// +kubebuilder:validation:Optional
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
	OrgName string `json:"orgName" vault:"org_name"`

	// BaseURL is the base domain for Okta API requests.
	// If unset, "okta.com" is used. Other valid values: "oktapreview.com", "okta-emea.com".
	// +kubebuilder:validation:Optional
	BaseURL string `json:"baseURL,omitempty" vault:"base_url"`

	// BypassOktaMFA bypasses an Okta MFA request. Useful when using Vault's built-in MFA.
	// +kubebuilder:validation:Optional
	BypassOktaMFA bool `json:"bypassOktaMFA,omitempty" vault:"bypass_okta_mfa"`

	// TokenTTL is the incremental lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty" vault:"token_ttl,duration"`

	// TokenMaxTTL is the maximum lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty" vault:"token_max_ttl,duration"`

	// TokenPolicies are policies to encode onto generated tokens.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenPolicies []string `json:"tokenPolicies,omitempty" vault:"token_policies"`

	// TokenBoundCIDRs are CIDR blocks restricting authentication and tying the token.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty" vault:"token_bound_cidrs"`

	// TokenExplicitMaxTTL is the hard cap max TTL for tokens.
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty" vault:"token_explicit_max_ttl,duration"`

	// TokenNoDefaultPolicy if true, omits the default policy from generated tokens.
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty" vault:"token_no_default_policy"`

	// TokenNumUses is the max number of times a generated token may be used (0 = unlimited).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int64 `json:"tokenNumUses,omitempty" vault:"token_num_uses,number"`

	// TokenPeriod is the maximum allowed period for periodic tokens.
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty" vault:"token_period,duration"`

	// TokenType is the type of token to generate.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
	TokenType string `json:"tokenType,omitempty" vault:"token_type"`

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

func (d *OktaAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
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

		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		apiToken, ok := secretData[randomSecret.Spec.SecretKey].(string)
		if !ok {
			return fmt.Errorf("secret key %q not found or not a string in vault secret", randomSecret.Spec.SecretKey)
		}
		r.Spec.OktaAuthConfig.retrievedAPIToken = apiToken
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
		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		passwordVal, ok := secretData[r.Spec.OktaCredentials.PasswordKey].(string)
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
	payload := vaultutils.StructToVaultPayload(c)
	payload["api_token"] = c.retrievedAPIToken
	return payload
}

func init() {
	SchemeBuilder.Register(&OktaAuthEngineConfig{}, &OktaAuthEngineConfigList{})
}
