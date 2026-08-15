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

// AWSAuthEngineClientConfigSpec defines the desired state of AWSAuthEngineClientConfig
type AWSAuthEngineClientConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/config/client.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	AWSAuthClientConfig `json:",inline"`

	// AWSCredentials specifies how to retrieve the AWS access key and secret key credentials.
	// +kubebuilder:validation:Required
	AWSCredentials vaultutils.RootCredentialConfig `json:"AWSCredentials,omitempty"`
}

type AWSAuthClientConfig struct {
	// AccessKey is the AWS access key ID for API calls.
	// +kubebuilder:validation:Optional
	AccessKey string `json:"accessKey,omitempty"`

	// Endpoint is a custom URL for EC2 API calls.
	// +kubebuilder:validation:Optional
	Endpoint string `json:"endpoint,omitempty"`

	// IAMEndpoint is a custom URL for IAM API calls.
	// +kubebuilder:validation:Optional
	IAMEndpoint string `json:"iamEndpoint,omitempty"`

	// STSEndpoint is a custom URL for STS API calls.
	// +kubebuilder:validation:Optional
	STSEndpoint string `json:"stsEndpoint,omitempty"`

	// STSRegion is the region for STS API calls (set with stsEndpoint).
	// +kubebuilder:validation:Optional
	STSRegion string `json:"stsRegion,omitempty"`

	// UseSTSRegionFromClient overrides stsEndpoint/stsRegion to use client request region.
	// +kubebuilder:validation:Optional
	UseSTSRegionFromClient bool `json:"useSTSRegionFromClient,omitempty"`

	// IAMServerIDHeaderValue is the value to require in the X-Vault-AWS-IAM-Server-ID header.
	// +kubebuilder:validation:Optional
	IAMServerIDHeaderValue string `json:"iamServerIDHeaderValue,omitempty"`

	// AllowedSTSHeaderValues is a comma-separated list of additional permitted STS request headers.
	// +kubebuilder:validation:Optional
	AllowedSTSHeaderValues string `json:"allowedSTSHeaderValues,omitempty"`

	// MaxRetries is the max retries for recoverable errors (-1 = AWS SDK default).
	// +kubebuilder:validation:Optional
	MaxRetries *int `json:"maxRetries,omitempty"`

	retrievedAccessKey string `json:"-"`
	retrievedSecretKey string `json:"-"`
}

// AWSAuthEngineClientConfigStatus defines the observed state of AWSAuthEngineClientConfig
type AWSAuthEngineClientConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSAuthEngineClientConfig is the Schema for the awsauthengineclientconfigs API
type AWSAuthEngineClientConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSAuthEngineClientConfigSpec   `json:"spec,omitempty"`
	Status AWSAuthEngineClientConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSAuthEngineClientConfigList contains a list of AWSAuthEngineClientConfig
type AWSAuthEngineClientConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSAuthEngineClientConfig `json:"items"`
}

var _ vaultutils.VaultObject = &AWSAuthEngineClientConfig{}
var _ vaultutils.ConditionsAware = &AWSAuthEngineClientConfig{}

func (d *AWSAuthEngineClientConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AWSAuthEngineClientConfig) IsDeletable() bool {
	return true
}

func (r *AWSAuthEngineClientConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AWSAuthEngineClientConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *AWSAuthEngineClientConfig) SetAccessKeyAndSecretKey(accessKey string, secretKey string) {
	r.Spec.AWSAuthClientConfig.retrievedAccessKey = accessKey
	r.Spec.AWSAuthClientConfig.retrievedSecretKey = secretKey
}

func (r *AWSAuthEngineClientConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config/client")
}

func (r *AWSAuthEngineClientConfig) GetPayload() map[string]any {
	return r.Spec.AWSAuthClientConfig.toMap()
}

func (r *AWSAuthEngineClientConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.AWSAuthClientConfig.toMap()
	delete(desiredState, "secret_key")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (r *AWSAuthEngineClientConfig) IsInitialized() bool {
	return true
}

func (r *AWSAuthEngineClientConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *AWSAuthEngineClientConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *AWSAuthEngineClientConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return r.setInternalCredentials(context)
}

func (r *AWSAuthEngineClientConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AWSAuthEngineClientConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if r.Spec.AWSCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.AWSCredentials.RandomSecret.Name,
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
	if r.Spec.AWSCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.AWSCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		passwordBytes, ok := secret.Data[r.Spec.AWSCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.AWSCredentials.Secret.Name, r.Spec.AWSCredentials.PasswordKey)
		}
		if r.Spec.AccessKey == "" {
			usernameBytes, ok := secret.Data[r.Spec.AWSCredentials.UsernameKey]
			if !ok {
				return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.AWSCredentials.Secret.Name, r.Spec.AWSCredentials.UsernameKey)
			}
			r.SetAccessKeyAndSecretKey(string(usernameBytes), string(passwordBytes))
		} else {
			r.SetAccessKeyAndSecretKey(r.Spec.AccessKey, string(passwordBytes))
		}
		return nil
	}
	if r.Spec.AWSCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.AWSCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		passwordVal, ok := secret.Data[r.Spec.AWSCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.AWSCredentials.PasswordKey)
		}
		if r.Spec.AccessKey == "" {
			usernameVal, ok := secret.Data[r.Spec.AWSCredentials.UsernameKey].(string)
			if !ok {
				return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.AWSCredentials.UsernameKey)
			}
			r.SetAccessKeyAndSecretKey(usernameVal, passwordVal)
		} else {
			r.SetAccessKeyAndSecretKey(r.Spec.AccessKey, passwordVal)
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *AWSAuthEngineClientConfig) isValid() error {
	if err := r.Spec.AWSCredentials.ValidateCredentialSource(); err != nil {
		return err
	}
	if r.Spec.AWSCredentials.RandomSecret != nil && r.Spec.AccessKey == "" {
		return errors.New("spec.accessKey must be set when using randomSecret credentials (randomSecret only provides the secret key)")
	}
	return nil
}

func (i *AWSAuthClientConfig) toMap() map[string]any {
	payload := map[string]any{}
	if i.retrievedAccessKey != "" {
		payload["access_key"] = i.retrievedAccessKey
	} else if i.AccessKey != "" {
		payload["access_key"] = i.AccessKey
	}
	if i.retrievedSecretKey != "" {
		payload["secret_key"] = i.retrievedSecretKey
	}
	payload["endpoint"] = i.Endpoint
	payload["iam_endpoint"] = i.IAMEndpoint
	payload["sts_endpoint"] = i.STSEndpoint
	payload["sts_region"] = i.STSRegion
	payload["use_sts_region_from_client"] = i.UseSTSRegionFromClient
	payload["iam_server_id_header_value"] = i.IAMServerIDHeaderValue
	payload["allowed_sts_header_values"] = i.AllowedSTSHeaderValues
	if i.MaxRetries != nil {
		payload["max_retries"] = *i.MaxRetries
	} else {
		payload["max_retries"] = -1
	}
	return payload
}

func init() {
	SchemeBuilder.Register(&AWSAuthEngineClientConfig{}, &AWSAuthEngineClientConfigList{})
}
