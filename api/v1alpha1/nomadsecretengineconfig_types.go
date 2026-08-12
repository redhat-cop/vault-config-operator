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

// NomadSecretEngineConfigSpec defines the desired state of NomadSecretEngineConfig
type NomadSecretEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config/access.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	NomadSEConfig `json:",inline"`

	// RootCredentials specifies how to retrieve the Nomad ACL management token.
	// +kubebuilder:validation:Required
	RootCredentials vaultutils.RootCredentialConfig `json:"rootCredentials,omitempty"`
}

type NomadSEConfig struct {
	// Address specifies the Nomad instance address as "protocol://host:port" (e.g., "http://127.0.0.1:4646")
	// +kubebuilder:validation:Required
	Address string `json:"address"`

	// MaxTokenNameLength specifies the maximum length for generated Nomad token names.
	// 0 uses Nomad's default (64 for ≤0.8.3, 256 for ≥0.8.4).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	MaxTokenNameLength int `json:"maxTokenNameLength,omitempty"`

	// CACert is the CA certificate for verifying the Nomad server certificate (x509 PEM)
	// +kubebuilder:validation:Optional
	CACert string `json:"caCert,omitempty"`

	// ClientCert is the client certificate for Nomad TLS communication (x509 PEM).
	// If set, clientKey must also be set.
	// +kubebuilder:validation:Optional
	ClientCert string `json:"clientCert,omitempty"`

	// ClientKey is the client key for Nomad TLS communication (x509 PEM).
	// If set, clientCert must also be set.
	// +kubebuilder:validation:Optional
	ClientKey string `json:"clientKey,omitempty"`

	retrievedToken string `json:"-"`
}

var _ vaultutils.VaultObject = &NomadSecretEngineConfig{}

func (d *NomadSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *NomadSecretEngineConfig) IsDeletable() bool {
	return false
}

func (d *NomadSecretEngineConfig) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "config" + "/" + "access")
}

func (d *NomadSecretEngineConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *NomadSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.NomadSEConfig.toMap()
	delete(desiredState, "token")
	delete(desiredState, "ca_cert")
	delete(desiredState, "client_cert")
	delete(desiredState, "client_key")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *NomadSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *NomadSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *NomadSecretEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *NomadSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *NomadSecretEngineConfig) setInternalCredentials(context context.Context) error {
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

		if randomSecret.Spec.IsKVSecretsEngineV2 {
			actualData, ok := secret.Data["data"].(map[string]any)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 response missing nested data map")
			}
			token, ok := actualData[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.Spec.NomadSEConfig.retrievedToken = token
		} else {
			token, ok := secret.Data[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.Spec.NomadSEConfig.retrievedToken = token
		}

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
		tokenBytes, ok := secret.Data[r.Spec.RootCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.RootCredentials.Secret.Name, r.Spec.RootCredentials.PasswordKey)
		}
		r.Spec.NomadSEConfig.retrievedToken = string(tokenBytes)
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
		var tokenVal string
		if dataInterface, exists := secret.Data["data"]; exists {
			if nestedData, ok := dataInterface.(map[string]any); ok {
				if val, ok := nestedData[r.Spec.RootCredentials.PasswordKey].(string); ok {
					tokenVal = val
				}
			}
		}
		if tokenVal == "" {
			if val, ok := secret.Data[r.Spec.RootCredentials.PasswordKey].(string); ok {
				tokenVal = val
			}
		}
		if tokenVal == "" {
			return fmt.Errorf("VaultSecret key %q not found or not a string in secret at %q", r.Spec.RootCredentials.PasswordKey, r.Spec.RootCredentials.VaultSecret.Path)
		}
		r.Spec.NomadSEConfig.retrievedToken = tokenVal
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (i *NomadSEConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["address"] = i.Address
	if i.MaxTokenNameLength != 0 {
		payload["max_token_name_length"] = json.Number(strconv.Itoa(i.MaxTokenNameLength))
	}
	if i.retrievedToken != "" {
		payload["token"] = i.retrievedToken
	}
	payload["ca_cert"] = i.CACert
	payload["client_cert"] = i.ClientCert
	payload["client_key"] = i.ClientKey
	return payload
}

func (r *NomadSecretEngineConfig) isValid() error {
	if err := r.Spec.RootCredentials.ValidateCredentialSource(); err != nil {
		return err
	}
	if (r.Spec.ClientCert != "") != (r.Spec.ClientKey != "") {
		return errors.New("spec.clientCert and spec.clientKey must both be set or both be empty")
	}
	return nil
}

// NomadSecretEngineConfigStatus defines the observed state of NomadSecretEngineConfig
type NomadSecretEngineConfigStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ vaultutils.ConditionsAware = &NomadSecretEngineConfig{}

func (m *NomadSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *NomadSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NomadSecretEngineConfig is the Schema for the nomadsecretengineconfigs API
type NomadSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NomadSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status NomadSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NomadSecretEngineConfigList contains a list of NomadSecretEngineConfig
type NomadSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NomadSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NomadSecretEngineConfig{}, &NomadSecretEngineConfigList{})
}

func (d *NomadSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
