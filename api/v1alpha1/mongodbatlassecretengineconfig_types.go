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

// MongoDBAtlasSecretEngineConfigSpec defines the desired state of MongoDBAtlasSecretEngineConfig
type MongoDBAtlasSecretEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/config.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	MongoDBAtlasSEConfig `json:",inline"`

	// RootCredentials specifies how to retrieve the MongoDB Atlas Programmatic API Key pair.
	// UsernameKey maps to public_key, PasswordKey maps to private_key.
	// +kubebuilder:validation:Optional
	RootCredentials vaultutils.RootCredentialConfig `json:"rootCredentials,omitempty"`
}

type MongoDBAtlasSEConfig struct {
	// PublicKey is the MongoDB Atlas Public Programmatic API Key.
	// Can be set directly here or resolved from credentials via RootCredentials.
	// When using RandomSecret, this field must be set (RandomSecret only provides the private key).
	// +kubebuilder:validation:Optional
	PublicKey string `json:"publicKey,omitempty" vault:"public_key,omitempty"`

	retrievedPublicKey  string `json:"-"`
	retrievedPrivateKey string `json:"-"`
}

var _ vaultutils.VaultObject = &MongoDBAtlasSecretEngineConfig{}

func (d *MongoDBAtlasSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *MongoDBAtlasSecretEngineConfig) IsDeletable() bool {
	return false
}

func (d *MongoDBAtlasSecretEngineConfig) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "config")
}

func (d *MongoDBAtlasSecretEngineConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *MongoDBAtlasSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.MongoDBAtlasSEConfig.toMap()
	delete(desiredState, "private_key")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *MongoDBAtlasSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (r *MongoDBAtlasSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *MongoDBAtlasSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if r.Spec.RootCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.RootCredentials.RandomSecret.Name,
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
			return fmt.Errorf("secret key %q not found or not a string in vault secret", randomSecret.Spec.SecretKey)
		}
		r.SetPublicKeyAndPrivateKey(r.Spec.PublicKey, secretVal)
		return nil
	}
	if r.Spec.RootCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.RootCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		passwordBytes, ok := secret.Data[r.Spec.RootCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.RootCredentials.Secret.Name, r.Spec.RootCredentials.PasswordKey)
		}
		if r.Spec.PublicKey == "" {
			usernameBytes, ok := secret.Data[r.Spec.RootCredentials.UsernameKey]
			if !ok {
				return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.RootCredentials.Secret.Name, r.Spec.RootCredentials.UsernameKey)
			}
			r.SetPublicKeyAndPrivateKey(string(usernameBytes), string(passwordBytes))
		} else {
			r.SetPublicKeyAndPrivateKey(r.Spec.PublicKey, string(passwordBytes))
		}
		return nil
	}
	if r.Spec.RootCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.RootCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		passwordVal, ok := secretData[r.Spec.RootCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.RootCredentials.PasswordKey)
		}
		if r.Spec.PublicKey == "" {
			usernameVal, ok := secretData[r.Spec.RootCredentials.UsernameKey].(string)
			if !ok {
				return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.RootCredentials.UsernameKey)
			}
			r.SetPublicKeyAndPrivateKey(usernameVal, passwordVal)
		} else {
			r.SetPublicKeyAndPrivateKey(r.Spec.PublicKey, passwordVal)
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *MongoDBAtlasSecretEngineConfig) SetPublicKeyAndPrivateKey(publicKey string, privateKey string) {
	r.Spec.MongoDBAtlasSEConfig.retrievedPublicKey = publicKey
	r.Spec.MongoDBAtlasSEConfig.retrievedPrivateKey = privateKey
}

func (i *MongoDBAtlasSEConfig) toMap() map[string]any {
	payload := vaultutils.StructToVaultPayload(i)
	if i.retrievedPublicKey != "" {
		payload["public_key"] = i.retrievedPublicKey
	}
	if i.retrievedPrivateKey != "" {
		payload["private_key"] = i.retrievedPrivateKey
	}
	return payload
}

func (r *MongoDBAtlasSecretEngineConfig) isValid() error {
	if err := r.Spec.RootCredentials.ValidateCredentialSource(); err != nil {
		return err
	}
	if r.Spec.RootCredentials.RandomSecret != nil && r.Spec.PublicKey == "" {
		return errors.New("spec.publicKey must be set when using randomSecret credentials (randomSecret only provides the private key)")
	}
	return nil
}

// MongoDBAtlasSecretEngineConfigStatus defines the observed state of MongoDBAtlasSecretEngineConfig
type MongoDBAtlasSecretEngineConfigStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ vaultutils.ConditionsAware = &MongoDBAtlasSecretEngineConfig{}

func (m *MongoDBAtlasSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *MongoDBAtlasSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MongoDBAtlasSecretEngineConfig is the Schema for the mongodbatlassecretengineconfigs API
type MongoDBAtlasSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBAtlasSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status MongoDBAtlasSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongoDBAtlasSecretEngineConfigList contains a list of MongoDBAtlasSecretEngineConfig
type MongoDBAtlasSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDBAtlasSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDBAtlasSecretEngineConfig{}, &MongoDBAtlasSecretEngineConfigList{})
}

func (d *MongoDBAtlasSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
