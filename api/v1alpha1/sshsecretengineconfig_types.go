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

var _ vaultutils.VaultObject = &SSHSecretEngineConfig{}
var _ vaultutils.ConditionsAware = &SSHSecretEngineConfig{}

// SSHSecretEngineConfigSpec defines the desired state of SSHSecretEngineConfig
type SSHSecretEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config/ca.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// CAKeyReference specifies how to retrieve the SSH CA private key.
	// Only needed when generateSigningKey is false.
	// Only VaultSecretReference or LocalObjectReference can be used.
	// +kubebuilder:validation:Optional
	CAKeyReference *vaultutils.RootCredentialConfig `json:"caKeyReference,omitempty"`

	SSHSEConfig `json:",inline"`
}

type SSHSEConfig struct {
	// GenerateSigningKey if true, Vault generates the SSH CA key pair internally.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	GenerateSigningKey bool `json:"generateSigningKey"`

	// KeyType desired key type for generated SSH CA key (ssh-rsa, ecdsa-sha2-nistp256, etc.)
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="ssh-rsa"
	KeyType string `json:"keyType"`

	// KeyBits desired key bits for generated SSH CA key (0 = default)
	// +kubebuilder:validation:Optional
	KeyBits int `json:"keyBits,omitempty"`

	retrievedPrivateKey string `json:"-"`
	retrievedPublicKey  string `json:"-"`
}

func (i *SSHSEConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["generate_signing_key"] = i.GenerateSigningKey
	payload["key_type"] = i.KeyType
	payload["key_bits"] = i.KeyBits
	payload["private_key"] = i.retrievedPrivateKey
	payload["public_key"] = i.retrievedPublicKey
	return payload
}

// SSHSecretEngineConfigStatus defines the observed state of SSHSecretEngineConfig
type SSHSecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *SSHSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *SSHSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SSHSecretEngineConfig is the Schema for the sshsecretengineconfigs API
type SSHSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSHSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status SSHSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SSHSecretEngineConfigList contains a list of SSHSecretEngineConfig
type SSHSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSHSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SSHSecretEngineConfig{}, &SSHSecretEngineConfigList{})
}

func (d *SSHSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *SSHSecretEngineConfig) IsDeletable() bool {
	return true
}

func (d *SSHSecretEngineConfig) GetPath() string {
	return string(d.Spec.Path) + "/" + "config/ca"
}

func (d *SSHSecretEngineConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *SSHSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.SSHSEConfig.toMap()
	// Vault's GET config/ca only returns public_key. private_key is write-only,
	// and generate_signing_key/key_type/key_bits are not returned on read.
	delete(desiredState, "private_key")
	delete(desiredState, "generate_signing_key")
	delete(desiredState, "key_type")
	delete(desiredState, "key_bits")
	if d.Spec.GenerateSigningKey {
		// Vault's SSH CA is immutable once created — /config/ca is write-once.
		// Any existing CA satisfies convergence; re-generating would invalidate
		// all previously signed certificates.
		delete(desiredState, "public_key")
	}
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *SSHSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *SSHSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *SSHSecretEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *SSHSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *SSHSecretEngineConfig) isValid() error {
	if r.Spec.GenerateSigningKey {
		return nil
	}
	if r.Spec.CAKeyReference == nil {
		return errors.New("spec.caKeyReference is required when generateSigningKey is false")
	}
	if r.Spec.CAKeyReference.RandomSecret != nil {
		return errors.New("spec.caKeyReference.randomSecret is not allowed; only vaultSecret or secret can be specified")
	}
	return r.Spec.CAKeyReference.ValidateCredentialSource()
}

func (r *SSHSecretEngineConfig) caPrivateKeyField() string {
	if r.Spec.CAKeyReference != nil &&
		r.Spec.CAKeyReference.PasswordKey != "" &&
		r.Spec.CAKeyReference.PasswordKey != "password" {
		return r.Spec.CAKeyReference.PasswordKey
	}
	return "private_key"
}

func (r *SSHSecretEngineConfig) caPublicKeyField() string {
	if r.Spec.CAKeyReference != nil &&
		r.Spec.CAKeyReference.UsernameKey != "" &&
		r.Spec.CAKeyReference.UsernameKey != "username" {
		return r.Spec.CAKeyReference.UsernameKey
	}
	return "public_key"
}

func (r *SSHSecretEngineConfig) setInternalCredentials(context context.Context) error {
	if r.Spec.GenerateSigningKey {
		return nil
	}
	if r.Spec.CAKeyReference == nil {
		return nil
	}
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	privateKeyField := r.caPrivateKeyField()
	publicKeyField := r.caPublicKeyField()
	if r.Spec.CAKeyReference.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.CAKeyReference.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		privBytes, privOK := secret.Data[privateKeyField]
		if !privOK || len(privBytes) == 0 {
			err := errors.New("referenced K8s Secret is missing required field: " + privateKeyField)
			log.Error(err, "unable to resolve private key from K8s Secret", "secret", r.Spec.CAKeyReference.Secret.Name)
			return err
		}
		pubBytes, pubOK := secret.Data[publicKeyField]
		if !pubOK || len(pubBytes) == 0 {
			err := errors.New("referenced K8s Secret is missing required field: " + publicKeyField)
			log.Error(err, "unable to resolve public key from K8s Secret", "secret", r.Spec.CAKeyReference.Secret.Name)
			return err
		}
		r.Spec.retrievedPrivateKey = string(privBytes)
		r.Spec.retrievedPublicKey = string(pubBytes)
		return nil
	}
	if r.Spec.CAKeyReference.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.CAKeyReference.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err := errors.New("vault secret not found at path: " + string(r.Spec.CAKeyReference.VaultSecret.Path))
			log.Error(err, "unable to resolve CA key material", "instance", r)
			return err
		}
		var secretData map[string]any
		if dataInterface, ok := secret.Data["data"]; ok {
			secretData, ok = dataInterface.(map[string]any)
			if !ok {
				err := errors.New("vault secret data is not in expected format for KV v2")
				log.Error(err, "unable to parse vault secret data", "path", r.Spec.CAKeyReference.VaultSecret.Path)
				return err
			}
		} else {
			secretData = secret.Data
		}
		if pk, ok := secretData[privateKeyField].(string); ok && pk != "" {
			r.Spec.retrievedPrivateKey = pk
		} else {
			err := errors.New("vault secret does not contain expected key: " + privateKeyField)
			log.Error(err, "unable to resolve private key from vault secret", "path", r.Spec.CAKeyReference.VaultSecret.Path)
			return err
		}
		if pub, ok := secretData[publicKeyField].(string); ok && pub != "" {
			r.Spec.retrievedPublicKey = pub
		} else {
			err := errors.New("vault secret does not contain expected key: " + publicKeyField)
			log.Error(err, "unable to resolve public key from vault secret", "path", r.Spec.CAKeyReference.VaultSecret.Path)
			return err
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (d *SSHSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
