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

// GCPSecretEngineConfigSpec defines the desired state of GCPSecretEngineConfig
type GCPSecretEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// GCPCredentials specifies how to retrieve the GCP credentials JSON.
	// +kubebuilder:validation:Required
	GCPCredentials GCPCredentialConfig `json:"gcpCredentials,omitempty"`

	GCPSEConfig `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// GCPCredentialConfig specifies how to retrieve GCP credentials. It mirrors
// RootCredentialConfig but defaults passwordKey to "credentials" (the standard
// GCP service account key field) instead of "password".
// +kubebuilder:object:generate=true
type GCPCredentialConfig struct {
	// VaultSecret retrieves the credentials from a Vault secret. Only one of VaultSecret, Secret, or RandomSecret can be specified.
	// +kubebuilder:validation:Optional
	VaultSecret *vaultutils.VaultSecretReference `json:"vaultSecret,omitempty"`

	// Secret retrieves the credentials from a Kubernetes secret. Only one of VaultSecret, Secret, or RandomSecret can be specified.
	// +kubebuilder:validation:Optional
	Secret *corev1.LocalObjectReference `json:"secret,omitempty"`

	// RandomSecret retrieves the credentials from the Vault secret corresponding to this RandomSecret. Only one of VaultSecret, Secret, or RandomSecret can be specified.
	// +kubebuilder:validation:Optional
	RandomSecret *corev1.LocalObjectReference `json:"randomSecret,omitempty"`

	// PasswordKey key to be used when retrieving the credentials, required with VaultSecrets and Kubernetes secrets, ignored with RandomSecret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="credentials"
	PasswordKey string `json:"passwordKey,omitempty"`

	// UsernameKey key to be used when retrieving the username, optional with VaultSecrets and Kubernetes secrets, ignored with RandomSecret
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="username"
	UsernameKey string `json:"usernameKey,omitempty"`
}

func (c *GCPCredentialConfig) ValidateCredentialSource() error {
	count := 0
	if c.VaultSecret != nil {
		count++
	}
	if c.Secret != nil {
		count++
	}
	if c.RandomSecret != nil {
		count++
	}
	if count != 1 {
		return errors.New("exactly one of spec.gcpCredentials.vaultSecret, spec.gcpCredentials.secret, or spec.gcpCredentials.randomSecret must be specified")
	}
	return nil
}

type GCPSEConfig struct {
	// TTL specifies the default TTL for long-lived credentials (service account keys). Duration format.
	// +kubebuilder:validation:Optional
	TTL string `json:"ttl,omitempty"`

	// MaxTTL specifies the maximum TTL for long-lived credentials. Duration format.
	// +kubebuilder:validation:Optional
	MaxTTL string `json:"maxTTL,omitempty"`

	retrievedCredentials string `json:"-"`
}

var _ vaultutils.VaultObject = &GCPSecretEngineConfig{}
var _ vaultutils.ConditionsAware = &GCPSecretEngineConfig{}

func (d *GCPSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *GCPSecretEngineConfig) IsDeletable() bool {
	return false
}

func (d *GCPSecretEngineConfig) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "config")
}

func (d *GCPSecretEngineConfig) GetPayload() map[string]any {
	return d.Spec.GCPSEConfig.toMap()
}

func (d *GCPSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.GCPSEConfig.toMap()
	delete(desiredState, "credentials")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *GCPSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *GCPSecretEngineConfig) IsValid() (bool, error) {
	err := d.isValid()
	return err == nil, err
}

func (d *GCPSecretEngineConfig) isValid() error {
	return d.Spec.GCPCredentials.ValidateCredentialSource()
}

func (d *GCPSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *GCPSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *GCPSecretEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *GCPSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if d.Spec.GCPCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: d.Namespace,
			Name:      d.Spec.GCPCredentials.RandomSecret.Name,
		}, randomSecret)
		if err != nil {
			log.Error(err, "unable to retrieve RandomSecret", "instance", d)
			return err
		}
		secret, exists, err := vaultutils.ReadSecret(context, randomSecret.GetPath())
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", d)
			return err
		}
		if randomSecret.Spec.IsKVSecretsEngineV2 {
			actualData, ok := secret.Data["data"].(map[string]any)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 response missing nested data map")
			}
			secretKey, ok := actualData[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			d.Spec.GCPSEConfig.retrievedCredentials = secretKey
		} else {
			secretKey, ok := secret.Data[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			d.Spec.GCPSEConfig.retrievedCredentials = secretKey
		}
		return nil
	}
	if d.Spec.GCPCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: d.Namespace,
			Name:      d.Spec.GCPCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", d)
			return err
		}
		passwordBytes, ok := secret.Data[d.Spec.GCPCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", d.Spec.GCPCredentials.Secret.Name, d.Spec.GCPCredentials.PasswordKey)
		}
		d.Spec.GCPSEConfig.retrievedCredentials = string(passwordBytes)
		return nil
	}
	if d.Spec.GCPCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(d.Spec.GCPCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", d)
			return err
		}
		passwordVal, ok := secret.Data[d.Spec.GCPCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", d.Spec.GCPCredentials.PasswordKey)
		}
		d.Spec.GCPSEConfig.retrievedCredentials = passwordVal
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (i *GCPSEConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["credentials"] = i.retrievedCredentials
	payload["ttl"] = i.TTL
	payload["max_ttl"] = i.MaxTTL
	return payload
}

// GCPSecretEngineConfigStatus defines the observed state of GCPSecretEngineConfig
type GCPSecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *GCPSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *GCPSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GCPSecretEngineConfig is the Schema for the gcpsecretengineconfigs API
type GCPSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status GCPSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GCPSecretEngineConfigList contains a list of GCPSecretEngineConfig
type GCPSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPSecretEngineConfig{}, &GCPSecretEngineConfigList{})
}
