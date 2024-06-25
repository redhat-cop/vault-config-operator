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

// AzureSecretEngineConfigSpec defines the desired state of AzureSecretEngineConfig
type AzureSecretEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// AzureCredentials consists in ClientID and ClientSecret, which can be created as Kubernetes Secret, VaultSecret or RandomSecret
	// +kubebuilder:validation:Optional
	AzureCredentials vaultutils.RootCredentialConfig `json:"azureCredentials,omitempty"`

	// +kubebuilder:validation:Required
	AzureSEConfig `json:",inline"`
}

// AzureSecretEngineConfigStatus defines the observed state of AzureSecretEngineConfig
type AzureSecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AzureSecretEngineConfig is the Schema for the azuresecretengineconfigs API
type AzureSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status AzureSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AzureSecretEngineConfigList contains a list of AzureSecretEngineConfig
type AzureSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureSecretEngineConfig `json:"items"`
}

type AzureSEConfig struct {

	// The subscription id for the Azure Active Directory. This value can also be provided with the AZURE_SUBSCRIPTION_ID environment variable.
	// +kubebuilder:validation:Required
	SubscriptionID string `json:"subscriptionID"`

	// The tenant id for the Azure Active Directory organization. This value can also be provided with the AZURE_TENANT_ID environment variable.
	// +kubebuilder:validation:Required
	TenantID string `json:"tenantID"`

	// The client id for credentials to query the Azure APIs.
	// Currently read permissions to query compute resources are required.
	// This value can also be provided with the AZURE_CLIENT_ID environment variable.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	ClientID string `json:"clientID,omitempty"`

	// The Azure cloud environment. Valid values: AzurePublicCloud, AzureUSGovernmentCloud, AzureChinaCloud, AzureGermanCloud.
	// This value can also be provided with the AZURE_ENVIRONMENT environment variable
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="AzurePublicCloud"
	Environment string `json:"environment,omitempty"`

	// Specifies a password policy to use when creating dynamic credentials. Defaults to generating an alphanumeric password if not set.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	PasswordPolicy string `json:"passwordPolicy,omitempty"`

	// Specifies how long the root password is valid for in Azure when rotate-root generates a new client secret. Uses duration format strings.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="182d"
	RootPasswordTTL string `json:"rootPasswordTTL,omitempty"`

	retrievedClientID string `json:"-"`

	retrievedClientPassword string `json:"-"`
}

var _ vaultutils.VaultObject = &AzureSecretEngineConfig{}
var _ vaultutils.ConditionsAware = &AzureSecretEngineConfig{}

func init() {
	SchemeBuilder.Register(&AzureSecretEngineConfig{}, &AzureSecretEngineConfigList{})
}

func (r *AzureSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (d *AzureSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *AzureSecretEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AzureSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (d *AzureSecretEngineConfig) GetPath() string {
	return string(d.Spec.Path) + "/" + "config"
}

func (d *AzureSecretEngineConfig) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (r *AzureSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.AzureSEConfig.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (r *AzureSecretEngineConfig) IsInitialized() bool {
	return true
}

func (r *AzureSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *AzureSecretEngineConfig) isValid() error {
	return r.Spec.AzureCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}

func (r *AzureSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {

	if reflect.DeepEqual(r.Spec.AzureCredentials, vaultutils.RootCredentialConfig{PasswordKey: "clientsecret", UsernameKey: "clientid"}) {
		return nil
	}

	return r.setInternalCredentials(context)
}

func (d *AzureSecretEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AzureSecretEngineConfig) setInternalCredentials(context context.Context) error {
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
			r.SetClientIDAndClientSecret(r.Spec.AzureSEConfig.ClientID, string(secret.Data[r.Spec.AzureCredentials.PasswordKey]))
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
			r.SetClientIDAndClientSecret(r.Spec.AzureSEConfig.ClientID, secret.Data[r.Spec.AzureCredentials.PasswordKey].(string))
			log.V(1).Info("", "clientid", r.Spec.AzureSEConfig.ClientID, "clientsecret", secret.Data[r.Spec.AzureCredentials.PasswordKey].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *AzureSecretEngineConfig) SetClientIDAndClientSecret(ClientID string, ClientSecret string) {
	r.Spec.AzureSEConfig.retrievedClientID = ClientID
	r.Spec.AzureSEConfig.retrievedClientPassword = ClientSecret
}

func (i *AzureSEConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["subscription_id"] = i.SubscriptionID
	payload["tenant_id"] = i.TenantID
	payload["client_id"] = i.retrievedClientID
	payload["client_secret"] = i.retrievedClientPassword
	payload["environment"] = i.Environment
	payload["password_policy"] = i.PasswordPolicy
	payload["root_password_ttl"] = i.RootPasswordTTL

	return payload
}
