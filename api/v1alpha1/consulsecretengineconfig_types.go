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

// ConsulSecretEngineConfigSpec defines the desired state of ConsulSecretEngineConfig
type ConsulSecretEngineConfigSpec struct {

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

	ConsulSEConfig `json:",inline"`

	// RootCredentials specifies how to retrieve the Consul ACL management token.
	// +kubebuilder:validation:Required
	RootCredentials vaultutils.RootCredentialConfig `json:"rootCredentials,omitempty"`
}

type ConsulSEConfig struct {
	// Address specifies the Consul instance address as "host:port".
	// +kubebuilder:validation:Required
	Address string `json:"address"`

	// Scheme specifies the URL scheme (http or https).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="http"
	// +kubebuilder:validation:Enum:={"http","https"}
	Scheme string `json:"scheme"`

	// CACert is the CA certificate for verifying the Consul server certificate (x509 PEM).
	// +kubebuilder:validation:Optional
	CACert string `json:"caCert,omitempty"`

	// ClientCert is the client certificate for Consul TLS communication (x509 PEM).
	// If set, clientKey must also be set.
	// +kubebuilder:validation:Optional
	ClientCert string `json:"clientCert,omitempty"`

	// ClientKey is the client key for Consul TLS communication (x509 PEM).
	// If set, clientCert must also be set.
	// +kubebuilder:validation:Optional
	ClientKey string `json:"clientKey,omitempty"`

	retrievedToken string `json:"-"`
}

var _ vaultutils.VaultObject = &ConsulSecretEngineConfig{}

func (d *ConsulSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *ConsulSecretEngineConfig) IsDeletable() bool {
	return false
}

func (d *ConsulSecretEngineConfig) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "config" + "/" + "access")
}

func (d *ConsulSecretEngineConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *ConsulSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.ConsulSEConfig.toMap()
	delete(desiredState, "token")
	delete(desiredState, "ca_cert")
	delete(desiredState, "client_cert")
	delete(desiredState, "client_key")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *ConsulSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *ConsulSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *ConsulSecretEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *ConsulSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *ConsulSecretEngineConfig) setInternalCredentials(context context.Context) error {
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
			r.Spec.ConsulSEConfig.retrievedToken = token
		} else {
			token, ok := secret.Data[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.Spec.ConsulSEConfig.retrievedToken = token
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
		r.Spec.ConsulSEConfig.retrievedToken = string(tokenBytes)
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
		r.Spec.ConsulSEConfig.retrievedToken = tokenVal
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (i *ConsulSEConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["address"] = i.Address
	payload["scheme"] = i.Scheme
	if i.retrievedToken != "" {
		payload["token"] = i.retrievedToken
	}
	payload["ca_cert"] = i.CACert
	payload["client_cert"] = i.ClientCert
	payload["client_key"] = i.ClientKey
	return payload
}

func (r *ConsulSecretEngineConfig) isValid() error {
	if err := r.Spec.RootCredentials.ValidateCredentialSource(); err != nil {
		return err
	}
	if (r.Spec.ClientCert != "") != (r.Spec.ClientKey != "") {
		return errors.New("spec.clientCert and spec.clientKey must both be set or both be empty")
	}
	return nil
}

// ConsulSecretEngineConfigStatus defines the observed state of ConsulSecretEngineConfig
type ConsulSecretEngineConfigStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ vaultutils.ConditionsAware = &ConsulSecretEngineConfig{}

func (m *ConsulSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *ConsulSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConsulSecretEngineConfig is the Schema for the consulsecretengineconfigs API
type ConsulSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsulSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status ConsulSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConsulSecretEngineConfigList contains a list of ConsulSecretEngineConfig
type ConsulSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConsulSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConsulSecretEngineConfig{}, &ConsulSecretEngineConfigList{})
}

func (d *ConsulSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
