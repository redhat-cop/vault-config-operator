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

// TerraformCloudSecretEngineConfigSpec defines the desired state of TerraformCloudSecretEngineConfig
type TerraformCloudSecretEngineConfigSpec struct {
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

	// TFCCredentials specifies how to retrieve the Terraform Cloud API token.
	// +kubebuilder:validation:Required
	TFCCredentials TFCCredentialConfig `json:"tfcCredentials,omitempty"`

	TFCSEConfig `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// TFCCredentialConfig specifies how to retrieve the Terraform Cloud API token.
// It mirrors RootCredentialConfig but defaults passwordKey to "token".
// +kubebuilder:object:generate=true
type TFCCredentialConfig struct {
	// VaultSecret retrieves the credentials from a Vault secret. Only one of VaultSecret, Secret, or RandomSecret can be specified.
	// +kubebuilder:validation:Optional
	VaultSecret *vaultutils.VaultSecretReference `json:"vaultSecret,omitempty"`

	// Secret retrieves the credentials from a Kubernetes secret. Only one of VaultSecret, Secret, or RandomSecret can be specified.
	// +kubebuilder:validation:Optional
	Secret *corev1.LocalObjectReference `json:"secret,omitempty"`

	// RandomSecret retrieves the credentials from the Vault secret corresponding to this RandomSecret. Only one of VaultSecret, Secret, or RandomSecret can be specified.
	// +kubebuilder:validation:Optional
	RandomSecret *corev1.LocalObjectReference `json:"randomSecret,omitempty"`

	// PasswordKey key to be used when retrieving the token
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="token"
	PasswordKey string `json:"passwordKey"`
}

func (c *TFCCredentialConfig) ValidateCredentialSource() error {
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
		return errors.New("exactly one of spec.tfcCredentials.vaultSecret, spec.tfcCredentials.secret, or spec.tfcCredentials.randomSecret must be specified")
	}
	if c.VaultSecret != nil && c.VaultSecret.Path == "" {
		return errors.New("spec.tfcCredentials.vaultSecret.path must not be empty")
	}
	if c.Secret != nil && c.Secret.Name == "" {
		return errors.New("spec.tfcCredentials.secret.name must not be empty")
	}
	if c.RandomSecret != nil && c.RandomSecret.Name == "" {
		return errors.New("spec.tfcCredentials.randomSecret.name must not be empty")
	}
	return nil
}

type TFCSEConfig struct {
	// Address is the URL of the Terraform Cloud/Enterprise instance.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="https://app.terraform.io"
	Address string `json:"address"`

	retrievedToken string `json:"-"`
}

var _ vaultutils.VaultObject = &TerraformCloudSecretEngineConfig{}
var _ vaultutils.ConditionsAware = &TerraformCloudSecretEngineConfig{}

func (d *TerraformCloudSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *TerraformCloudSecretEngineConfig) IsDeletable() bool {
	return false
}

func (d *TerraformCloudSecretEngineConfig) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "config")
}

func (d *TerraformCloudSecretEngineConfig) GetPayload() map[string]any {
	return d.Spec.TFCSEConfig.toMap()
}

func (d *TerraformCloudSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.TFCSEConfig.toMap()
	delete(desiredState, "token")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *TerraformCloudSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *TerraformCloudSecretEngineConfig) IsValid() (bool, error) {
	err := d.isValid()
	return err == nil, err
}

func (d *TerraformCloudSecretEngineConfig) isValid() error {
	if err := d.Spec.TFCCredentials.ValidateCredentialSource(); err != nil {
		return err
	}
	if (d.Spec.TFCCredentials.Secret != nil || d.Spec.TFCCredentials.VaultSecret != nil) && d.Spec.TFCCredentials.PasswordKey == "" {
		return errors.New("spec.tfcCredentials.passwordKey must not be empty when a credential source is specified")
	}
	return nil
}

func (d *TerraformCloudSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *TerraformCloudSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *TerraformCloudSecretEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *TerraformCloudSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if d.Spec.TFCCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: d.Namespace,
			Name:      d.Spec.TFCCredentials.RandomSecret.Name,
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
			d.Spec.TFCSEConfig.retrievedToken = secretKey
		} else {
			secretKey, ok := secret.Data[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			d.Spec.TFCSEConfig.retrievedToken = secretKey
		}
		return nil
	}
	if d.Spec.TFCCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: d.Namespace,
			Name:      d.Spec.TFCCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", d)
			return err
		}
		passwordBytes, ok := secret.Data[d.Spec.TFCCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", d.Spec.TFCCredentials.Secret.Name, d.Spec.TFCCredentials.PasswordKey)
		}
		d.Spec.TFCSEConfig.retrievedToken = string(passwordBytes)
		return nil
	}
	if d.Spec.TFCCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(d.Spec.TFCCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", d)
			return err
		}
		passwordVal, ok := secret.Data[d.Spec.TFCCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", d.Spec.TFCCredentials.PasswordKey)
		}
		d.Spec.TFCSEConfig.retrievedToken = passwordVal
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (i *TFCSEConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["address"] = i.Address
	payload["token"] = i.retrievedToken
	return payload
}

// TerraformCloudSecretEngineConfigStatus defines the observed state of TerraformCloudSecretEngineConfig
type TerraformCloudSecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *TerraformCloudSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *TerraformCloudSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TerraformCloudSecretEngineConfig is the Schema for the terraformcloudsecretengineconfigs API
type TerraformCloudSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformCloudSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status TerraformCloudSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TerraformCloudSecretEngineConfigList contains a list of TerraformCloudSecretEngineConfig
type TerraformCloudSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TerraformCloudSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TerraformCloudSecretEngineConfig{}, &TerraformCloudSecretEngineConfigList{})
}
