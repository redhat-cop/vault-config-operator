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

// AzureAuthEngineConfigSpec defines the desired state of AzureAuthEngineConfig
type AzureAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// +kubebuilder:validation:Required
	AzureConfig `json:",inline"`

	// AzureCredentials consists in ClientID and ClientSecret, which can be created as Kubernetes Secret, VaultSecret or RandomSecret
	// +kubebuilder:validation:Optional
	AzureCredentials vaultutils.RootCredentialConfig `json:"azureCredentials,omitempty"`
}

// AzureAuthEngineConfigStatus defines the observed state of AzureAuthEngineConfig
type AzureAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AzureAuthEngineConfig is the Schema for the azureauthengineconfigs API
type AzureAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status AzureAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AzureAuthEngineConfigList contains a list of AzureAuthEngineConfig
type AzureAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureAuthEngineConfig `json:"items"`
}

type AzureConfig struct {
	//The tenant id for the Azure Active Directory organization. This value can also be provided with the AZURE_TENANT_ID environment variable.
	// +kubebuilder:validation:Required
	TenantID string `json:"tenantID"`

	//The resource URL for the application registered in Azure Active Directory.
	//The value is expected to match the audience (aud claim) of the JWT provided to the login API.
	//See the resource parameter for how the audience is set when requesting a JWT access token from the Azure Instance Metadata Service (IMDS) endpoint.
	//This value can also be provided with the AZURE_AD_RESOURCE environment variable.
	// +kubebuilder:validation:Required
	Resource string `json:"resource"`

	//The Azure cloud environment. Valid values: AzurePublicCloud, AzureUSGovernmentCloud, AzureChinaCloud, AzureGermanCloud.
	//This value can also be provided with the AZURE_ENVIRONMENT environment variable
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="AzurePublicCloud"
	Environment string `json:"environment,omitempty"`

	//The client id for credentials to query the Azure APIs.
	//Currently read permissions to query compute resources are required.
	//This value can also be provided with the AZURE_CLIENT_ID environment variable.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	ClientID string `json:"clientID,omitempty"`

	//The maximum number of attempts a failed operation will be retried before producing an error.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=3
	MaxRetries int64 `json:"maxRetries"`

	//The maximum delay, in seconds, allowed before retrying an operation
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=60
	MaxRetryDelay int64 `json:"maxRetryDelay"`

	//The initial amount of delay, in seconds, to use before retrying an operation.
	//Increases exponentially
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=4
	RetryDelay int64 `json:"retryDelay"`

	retrievedClientID string `json:"-"`

	retrievedClientPassword string `json:"-"`
}

var _ vaultutils.VaultObject = &AzureAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &AzureAuthEngineConfig{}

func (d *AzureAuthEngineConfig) IsDeletable() bool {
	return true
}

func (d *AzureAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *AzureAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AzureAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *AzureAuthEngineConfig) SetClientIDAndClientSecret(ClientID string, ClientSecret string) {
	r.Spec.AzureConfig.retrievedClientID = ClientID
	r.Spec.AzureConfig.retrievedClientPassword = ClientSecret
}

func (r *AzureAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config")
}

func (r *AzureAuthEngineConfig) GetPayload() map[string]interface{} {
	return r.Spec.AzureConfig.toMap()
}

func (r *AzureAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.AzureConfig.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (r *AzureAuthEngineConfig) IsInitialized() bool {
	return true
}

func (r *AzureAuthEngineConfig) IsValid() (bool, error) {
	return true, nil
}

func (r *AzureAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *AzureAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {

	if reflect.DeepEqual(r.Spec.AzureCredentials, vaultutils.RootCredentialConfig{PasswordKey: "clientsecret", UsernameKey: "clientid"}) {
		return nil
	}

	return r.setInternalCredentials(context)
}

func (r *AzureAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AzureAuthEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
	if r.Spec.AzureCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.AzureCredentials.RandomSecret.Name,
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
		r.SetClientIDAndClientSecret(r.Spec.ClientID, secret.Data[randomSecret.Spec.SecretKey].(string))
		return nil
	}
	if r.Spec.AzureCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.AzureCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		if r.Spec.ClientID == "" {
			r.SetClientIDAndClientSecret(string(secret.Data[r.Spec.AzureCredentials.UsernameKey]), string(secret.Data[r.Spec.AzureCredentials.PasswordKey]))
		} else {
			r.SetClientIDAndClientSecret(r.Spec.AzureConfig.ClientID, string(secret.Data[r.Spec.AzureCredentials.PasswordKey]))
		}
		return nil
	}
	if r.Spec.AzureCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.AzureCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		if r.Spec.ClientID == "" {
			r.SetClientIDAndClientSecret(secret.Data[r.Spec.AzureCredentials.UsernameKey].(string), secret.Data[r.Spec.AzureCredentials.PasswordKey].(string))
			log.V(1).Info("", "clientid", secret.Data[r.Spec.AzureCredentials.UsernameKey].(string), "clientsecret", secret.Data[r.Spec.AzureCredentials.PasswordKey].(string))
		} else {
			r.SetClientIDAndClientSecret(r.Spec.AzureConfig.ClientID, secret.Data[r.Spec.AzureCredentials.PasswordKey].(string))
			log.V(1).Info("", "clientid", r.Spec.AzureConfig.ClientID, "clientsecret", secret.Data[r.Spec.AzureCredentials.PasswordKey].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func init() {
	SchemeBuilder.Register(&AzureAuthEngineConfig{}, &AzureAuthEngineConfigList{})
}

func (i *AzureConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["tenant_id"] = i.TenantID
	payload["resource"] = i.Resource
	payload["environment"] = i.Environment
	payload["client_id"] = i.retrievedClientID
	payload["client_secret"] = i.retrievedClientPassword
	payload["max_retries"] = i.MaxRetries
	payload["max_retry_delay"] = i.MaxRetryDelay
	payload["retry_delay"] = i.RetryDelay

	return payload
}
