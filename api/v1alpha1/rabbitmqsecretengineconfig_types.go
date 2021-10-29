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

	"github.com/redhat-cop/operator-utils/pkg/util/apis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RabbitMQSecretEngineConfigSpec defines the desired state of RabbitMQSecretEngineConfig
type RabbitMQSecretEngineConfigSpec struct {
	// Authentication is the k8s auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path will be {[spec.authentication.namespace]}/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path Path `json:"path,omitempty"`

	// +kubebuilder:validation:Required
	RMQSEConfig `json:",inline"`

	// RootCredentials specifies how to retrieve the credentials for this RabbitMQEngine connection.
	// +kubebuilder:validation:Required
	RootCredentials RootCredentialConfig `json:"rootCredentials,omitempty"`
}

type RMQSEConfig struct {
	// ConnectionURL Specifies the connection string used to connect to the RabbitMQ cluster. 
	// +kubebuilder:validation:Required
	ConnectionURI string `json:"connectionURI,omitempty"`

	// Username Specifies the name of the user to use as the "administrator" user when connecting to the RabbitMQ cluster. This "administrator" user is used to create/update/delete users, so you will need to ensure that this user has permissions to manipulate users. If management plugin is used, this user need to have "administrator" tag, no additional permissions necessary.  
	// If username is provided it takes precedence over the username retrieved from the referenced secrets
	// +kubebuilder:validation:Optional
	Username string `json:"username,omitempty"`
	
	// VerifyConnection Specifies if the connection is verified during initial configuration. Defaults to true.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	VerifyConnection bool `json:"verifyConnection,omitempty"`

	// PasswordPolicy The name of the password policy to use when generating passwords for this database. Defaults to generating an alphanumeric password if not set.
	// +kubebuilder:validation:Optional
	PasswordPolicy string `json:"passwordPolicy,omitempty"`

	// UsernameTemplate Vault username template describing how dynamic usernames are generated.
	UsernameTemplate string `json:"usernameTemplate,omitempty"`

	// Lease TTL for generated credentials in seconds.
	// +kubebuilder:validation:Optional
	LeaseTTL int `json:"leaseTTL,omitempty"`

	// Lease maximum TTL for generated credentials in seconds.
	// +kubebuilder:validation:Optional
	LeaseMaxTTL int `json:"leaseMaxTTL,omitempty"`

	retrievedPassword string `json:"-"`

	retrievedUsername string `json:"-"`
}

// RabbitMQSecretEngineConfigStatus defines the observed state of RabbitMQSecretEngineConfig
type RabbitMQSecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RabbitMQSecretEngineConfig is the Schema for the rabbitmqsecretengineconfigs API
type RabbitMQSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status RabbitMQSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RabbitMQSecretEngineConfigList contains a list of RabbitMQSecretEngineConfig
type RabbitMQSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQSecretEngineConfig{}, &RabbitMQSecretEngineConfigList{})
}

func (fields *RMQSEConfig) rabbitMQToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["connection_uri"] = fields.ConnectionURI
	payload["verify_connection"] = fields.VerifyConnection
	payload["username"] = fields.retrievedUsername
	payload["password"] = fields.retrievedPassword

	payload["username_template"] = fields.UsernameTemplate
	payload["password_policy"] = fields.PasswordPolicy

	return payload
}

var _ apis.ConditionsAware = &RabbitMQSecretEngineConfig{}

func (m *RabbitMQSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *RabbitMQSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (m *RabbitMQSecretEngineConfig) SetUsernameAndPassword(username string, password string) {
	m.Spec.RMQSEConfig.retrievedUsername = username
	m.Spec.RMQSEConfig.retrievedPassword = password
}

func (rabbitMQ *RabbitMQSecretEngineConfig) isValid() error {
	return rabbitMQ.Spec.RootCredentials.validateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}

var _ vaultutils.VaultObject = &RabbitMQSecretEngineConfig{}

func (rabbitMQ *RabbitMQSecretEngineConfig) GetPath() string {
	return string(rabbitMQ.Spec.Path) + "/config/connection"
}

func (rabbitMQ *RabbitMQSecretEngineConfig) GetPayload() map[string]interface{} {
	return rabbitMQ.Spec.rabbitMQToMap()
}

func (rabbitMQ *RabbitMQSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := rabbitMQ.Spec.RMQSEConfig.rabbitMQToMap()
	delete(desiredState, "password")
	return reflect.DeepEqual(desiredState, payload)
}

func (rabbitMQ *RabbitMQSecretEngineConfig) IsInitialized() bool {
	return true
}

func (rabbitMQ *RabbitMQSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return rabbitMQ.setInternalCredentials(context)
}

func (rabbitMQ *RabbitMQSecretEngineConfig) IsValid() (bool, error) {
	err := rabbitMQ.isValid()
	return err == nil, err
}

func (rabbitMQ *RabbitMQSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	k8sClient := context.Value("k8sClient").(client.Client)
	if rabbitMQ.Spec.RootCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := k8sClient.Get(context, types.NamespacedName{
			Namespace: rabbitMQ.Namespace,
			Name:      rabbitMQ.Spec.RootCredentials.RandomSecret.Name,
		}, randomSecret)
		if err != nil {
			log.Error(err, "unable to retrieve RandomSecret", "instance", rabbitMQ)
			return err
		}
		secret, err := GetVaultSecret(randomSecret.GetPath(), context)
		if err != nil {
			return err
		}
		rabbitMQ.SetUsernameAndPassword(rabbitMQ.Spec.Username, secret.Data[randomSecret.Spec.SecretKey].(string))
		return nil
	}
	if rabbitMQ.Spec.RootCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := k8sClient.Get(context, types.NamespacedName{
			Namespace: rabbitMQ.Namespace,
			Name:      rabbitMQ.Spec.RootCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", rabbitMQ)
			return err
		}
		if rabbitMQ.Spec.Username == "" {
			rabbitMQ.SetUsernameAndPassword(string(secret.Data[rabbitMQ.Spec.RootCredentials.UsernameKey]), string(secret.Data[rabbitMQ.Spec.RootCredentials.PasswordKey]))
		} else {
			rabbitMQ.SetUsernameAndPassword(rabbitMQ.Spec.Username, string(secret.Data[rabbitMQ.Spec.RootCredentials.PasswordKey]))
		}
		return nil
	}
	if rabbitMQ.Spec.RootCredentials.VaultSecret != nil {
		secret, err := GetVaultSecret(string(rabbitMQ.Spec.RootCredentials.VaultSecret.Path), context)
		if err != nil {
			return err
		}
		if rabbitMQ.Spec.Username == "" {
			rabbitMQ.SetUsernameAndPassword(secret.Data[rabbitMQ.Spec.RootCredentials.UsernameKey].(string), secret.Data[rabbitMQ.Spec.RootCredentials.PasswordKey].(string))
			log.V(1).Info("", "username", secret.Data[rabbitMQ.Spec.RootCredentials.UsernameKey].(string), "password", secret.Data[rabbitMQ.Spec.RootCredentials.PasswordKey].(string))
		} else {
			rabbitMQ.SetUsernameAndPassword(rabbitMQ.Spec.Username, secret.Data[rabbitMQ.Spec.RootCredentials.PasswordKey].(string))
			log.V(1).Info("", "username", rabbitMQ.Spec.Username, "password", secret.Data[rabbitMQ.Spec.RootCredentials.PasswordKey].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (rabbitMQ *RabbitMQSecretEngineConfig) GetLeasePath() string {
	return string(rabbitMQ.Spec.Path) + "/config/lease"
}
