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
	"os"
	"reflect"
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RADIUSAuthEngineConfigSpec defines the desired state of RADIUSAuthEngineConfig
type RADIUSAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the RADIUS auth engine is mounted.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	RADIUSAuthConfig `json:",inline"`

	// RADIUSCredentials is used to provide the RADIUS shared secret.
	// The shared secret can be sourced from a Kubernetes Secret, VaultSecret, or RandomSecret.
	// +kubebuilder:validation:Optional
	RADIUSCredentials vaultutils.RootCredentialConfig `json:"radiusCredentials,omitempty"`
}

type RADIUSAuthConfig struct {
	// Host is the RADIUS server to connect to (e.g. "radius.myorg.com", "127.0.0.1").
	// +kubebuilder:validation:Required
	Host string `json:"host"`

	// Port is the UDP port where the RADIUS server is listening. Defaults to 1812.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=1812
	Port int `json:"port"`

	// UnregisteredUserPolicies is a comma-separated list of policies to grant to
	// users who authenticate via RADIUS but have no explicit user mapping.
	// +kubebuilder:validation:Optional
	UnregisteredUserPolicies string `json:"unregisteredUserPolicies,omitempty"`

	// DialTimeout is the number of seconds to wait for a backend connection before timing out.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=10
	DialTimeout int `json:"dialTimeout"`

	// ReadTimeout is the number of seconds before a response times out.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=10
	ReadTimeout int `json:"readTimeout"`

	// NASPort is the NAS-Port attribute of the RADIUS request.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=10
	NASPort int `json:"nasPort"`

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
	TokenNumUses int `json:"tokenNumUses,omitempty"`

	// TokenPeriod is the maximum allowed period for periodic tokens.
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty"`

	// TokenType is the type of token to generate.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch",""}
	TokenType string `json:"tokenType,omitempty"`

	retrievedSecret string `json:"-"`
}

// RADIUSAuthEngineConfigStatus defines the observed state of RADIUSAuthEngineConfig
type RADIUSAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RADIUSAuthEngineConfig is the Schema for the radiusauthengineconfigs API
type RADIUSAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RADIUSAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status RADIUSAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RADIUSAuthEngineConfigList contains a list of RADIUSAuthEngineConfig
type RADIUSAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RADIUSAuthEngineConfig `json:"items"`
}

var _ vaultutils.VaultObject = &RADIUSAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &RADIUSAuthEngineConfig{}

func (d *RADIUSAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *RADIUSAuthEngineConfig) IsDeletable() bool {
	return false
}

func (d *RADIUSAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/config")
}

func (d *RADIUSAuthEngineConfig) GetPayload() map[string]any {
	return d.Spec.RADIUSAuthConfig.toMap()
}

func (d *RADIUSAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.RADIUSAuthConfig.toMap()
	delete(desiredState, "secret")
	removeUnsetFields(desiredState, payload)
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	for _, key := range []string{"token_policies", "token_bound_cidrs"} {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (d *RADIUSAuthEngineConfig) IsInitialized() bool {
	return true
}

func (d *RADIUSAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *RADIUSAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *RADIUSAuthEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *RADIUSAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *RADIUSAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *RADIUSAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

// resolveRADIUSPasswordKey returns the effective password key for RADIUS credential
// lookups. RADIUS secrets use the key "secret" by convention, but the shared
// RootCredentialConfig CRD schema defaults passwordKey to "password". When the
// mutating webhook is active it remaps omitted values; this function covers the
// cases where the webhook did not run (ENABLE_WEBHOOKS=false) or the field was
// left completely empty.
func resolveRADIUSPasswordKey(key string) string {
	if key == "" {
		return "secret"
	}
	if key == "password" {
		if webhooks, ok := os.LookupEnv("ENABLE_WEBHOOKS"); ok && webhooks == "false" {
			return "secret"
		}
	}
	return key
}

func (r *RADIUSAuthEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	r.Spec.RADIUSCredentials.PasswordKey = resolveRADIUSPasswordKey(r.Spec.RADIUSCredentials.PasswordKey)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if r.Spec.RADIUSCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.RADIUSCredentials.RandomSecret.Name,
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
			sharedSecret, ok := actualData[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.Spec.RADIUSAuthConfig.retrievedSecret = sharedSecret
		} else {
			sharedSecret, ok := secret.Data[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.Spec.RADIUSAuthConfig.retrievedSecret = sharedSecret
		}

		return nil
	}
	if r.Spec.RADIUSCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.RADIUSCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		passwordBytes, ok := secret.Data[r.Spec.RADIUSCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.RADIUSCredentials.Secret.Name, r.Spec.RADIUSCredentials.PasswordKey)
		}
		r.Spec.RADIUSAuthConfig.retrievedSecret = string(passwordBytes)
		return nil
	}
	if r.Spec.RADIUSCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.RADIUSCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		passwordVal, ok := secret.Data[r.Spec.RADIUSCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.RADIUSCredentials.PasswordKey)
		}
		r.Spec.RADIUSAuthConfig.retrievedSecret = passwordVal
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *RADIUSAuthEngineConfig) isValid() error {
	return r.Spec.RADIUSCredentials.ValidateCredentialSource()
}

func (c *RADIUSAuthConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["host"] = c.Host
	payload["secret"] = c.retrievedSecret
	payload["port"] = json.Number(strconv.Itoa(c.Port))
	payload["dial_timeout"] = json.Number(strconv.Itoa(c.DialTimeout))
	payload["read_timeout"] = json.Number(strconv.Itoa(c.ReadTimeout))
	payload["nas_port"] = json.Number(strconv.Itoa(c.NASPort))
	payload["unregistered_user_policies"] = c.UnregisteredUserPolicies
	payload["token_ttl"] = durationToSeconds(c.TokenTTL)
	payload["token_max_ttl"] = durationToSeconds(c.TokenMaxTTL)
	payload["token_policies"] = toInterfaceArray(c.TokenPolicies)
	payload["token_bound_cidrs"] = toInterfaceArray(c.TokenBoundCIDRs)
	payload["token_explicit_max_ttl"] = durationToSeconds(c.TokenExplicitMaxTTL)
	payload["token_no_default_policy"] = c.TokenNoDefaultPolicy
	payload["token_num_uses"] = json.Number(strconv.Itoa(c.TokenNumUses))
	payload["token_period"] = durationToSeconds(c.TokenPeriod)
	payload["token_type"] = c.TokenType
	return payload
}

func init() {
	SchemeBuilder.Register(&RADIUSAuthEngineConfig{}, &RADIUSAuthEngineConfigList{})
}
