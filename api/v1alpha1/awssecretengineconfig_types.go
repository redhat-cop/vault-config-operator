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

// AWSSecretEngineConfigSpec defines the desired state of AWSSecretEngineConfig
type AWSSecretEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config/root.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	AWSRootConfig `json:",inline"`

	// RootCredentials specifies how to retrieve the AWS access key and secret key credentials.
	// +kubebuilder:validation:Required
	RootCredentials vaultutils.RootCredentialConfig `json:"rootCredentials,omitempty"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type AWSRootConfig struct {
	// AccessKey is the AWS access key ID (set via spec or resolved from credentials)
	// +kubebuilder:validation:Optional
	AccessKey string `json:"accessKey,omitempty"`

	// Region is the AWS region (defaults to us-east-1 if not set)
	// +kubebuilder:validation:Optional
	Region string `json:"region,omitempty"`

	// IAMEndpoint is a custom HTTP IAM endpoint
	// +kubebuilder:validation:Optional
	IAMEndpoint string `json:"iamEndpoint,omitempty"`

	// STSEndpoint is a custom HTTP STS endpoint
	// +kubebuilder:validation:Optional
	STSEndpoint string `json:"stsEndpoint,omitempty"`

	// MaxRetries is the max retries for recoverable errors (-1 = SDK default)
	// +kubebuilder:validation:Optional
	MaxRetries *int `json:"maxRetries,omitempty"`

	// UsernameTemplate is a Go template for dynamic username generation
	// +kubebuilder:validation:Optional
	UsernameTemplate string `json:"usernameTemplate,omitempty"`

	retrievedAccessKey string `json:"-"`
	retrievedSecretKey string `json:"-"`
}

var _ vaultutils.VaultObject = &AWSSecretEngineConfig{}

func (d *AWSSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AWSSecretEngineConfig) IsDeletable() bool {
	return false
}

func (d *AWSSecretEngineConfig) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "config" + "/" + "root")
}

func (d *AWSSecretEngineConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *AWSSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.AWSRootConfig.toMap()
	delete(desiredState, "secret_key")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *AWSSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *AWSSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (d *AWSSecretEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AWSSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *AWSSecretEngineConfig) setInternalCredentials(context context.Context) error {
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
			secretKey, ok := actualData[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.SetAccessKeyAndSecretKey(r.Spec.AccessKey, secretKey)
		} else {
			secretKey, ok := secret.Data[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.SetAccessKeyAndSecretKey(r.Spec.AccessKey, secretKey)
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
		passwordBytes, ok := secret.Data[r.Spec.RootCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.RootCredentials.Secret.Name, r.Spec.RootCredentials.PasswordKey)
		}
		if r.Spec.AccessKey == "" {
			usernameBytes, ok := secret.Data[r.Spec.RootCredentials.UsernameKey]
			if !ok {
				return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.RootCredentials.Secret.Name, r.Spec.RootCredentials.UsernameKey)
			}
			r.SetAccessKeyAndSecretKey(string(usernameBytes), string(passwordBytes))
		} else {
			r.SetAccessKeyAndSecretKey(r.Spec.AccessKey, string(passwordBytes))
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
		passwordVal, ok := secret.Data[r.Spec.RootCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.RootCredentials.PasswordKey)
		}
		if r.Spec.AccessKey == "" {
			usernameVal, ok := secret.Data[r.Spec.RootCredentials.UsernameKey].(string)
			if !ok {
				return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.RootCredentials.UsernameKey)
			}
			r.SetAccessKeyAndSecretKey(usernameVal, passwordVal)
		} else {
			r.SetAccessKeyAndSecretKey(r.Spec.AccessKey, passwordVal)
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *AWSSecretEngineConfig) SetAccessKeyAndSecretKey(accessKey string, secretKey string) {
	r.Spec.AWSRootConfig.retrievedAccessKey = accessKey
	r.Spec.AWSRootConfig.retrievedSecretKey = secretKey
}

func (i *AWSRootConfig) toMap() map[string]any {
	payload := map[string]any{}
	if i.retrievedAccessKey != "" {
		payload["access_key"] = i.retrievedAccessKey
	} else if i.AccessKey != "" {
		payload["access_key"] = i.AccessKey
	}
	if i.retrievedSecretKey != "" {
		payload["secret_key"] = i.retrievedSecretKey
	}
	payload["region"] = i.Region
	payload["iam_endpoint"] = i.IAMEndpoint
	payload["sts_endpoint"] = i.STSEndpoint
	payload["username_template"] = i.UsernameTemplate
	if i.MaxRetries != nil {
		payload["max_retries"] = *i.MaxRetries
	} else {
		payload["max_retries"] = -1
	}
	return payload
}

func (r *AWSSecretEngineConfig) isValid() error {
	if err := r.Spec.RootCredentials.ValidateCredentialSource(); err != nil {
		return err
	}
	if r.Spec.RootCredentials.RandomSecret != nil && r.Spec.AccessKey == "" {
		return errors.New("spec.accessKey must be set when using randomSecret credentials (randomSecret only provides the secret key)")
	}
	return nil
}

// AWSSecretEngineConfigStatus defines the observed state of AWSSecretEngineConfig
type AWSSecretEngineConfigStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ vaultutils.ConditionsAware = &AWSSecretEngineConfig{}

func (m *AWSSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *AWSSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSSecretEngineConfig is the Schema for the awssecretengineconfigs API
type AWSSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status AWSSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSSecretEngineConfigList contains a list of AWSSecretEngineConfig
type AWSSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSSecretEngineConfig{}, &AWSSecretEngineConfigList{})
}

func (d *AWSSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
