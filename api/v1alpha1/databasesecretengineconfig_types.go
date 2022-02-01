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
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseSecretEngineConfigSpec defines the desired state of DatabaseSecretEngineConfig
type DatabaseSecretEngineConfigSpec struct {
	// Authentication is the kube aoth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path will be {[spec.authentication.namespace]}/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path Path `json:"path,omitempty"`

	DBSEConfig `json:",inline"`

	// RootCredentials specifies how to retrieve the credentials for this DatabaseEngine connection.
	// +kubebuilder:validation:Required
	RootCredentials RootCredentialConfig `json:"rootCredentials,omitempty"`
}

var _ vaultutils.VaultObject = &DatabaseSecretEngineConfig{}

func (d *DatabaseSecretEngineConfig) GetPath() string {
	return string(d.Spec.Path) + "/" + "config" + "/" + d.Name
}
func (d *DatabaseSecretEngineConfig) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *DatabaseSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.DBSEConfig.toMap()
	delete(desiredState, "password")
	return reflect.DeepEqual(desiredState, payload)
}

func (d *DatabaseSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *DatabaseSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (r *DatabaseSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *DatabaseSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
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
		r.SetUsernameAndPassword(r.Spec.Username, secret.Data[randomSecret.Spec.SecretKey].(string))
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
		if r.Spec.Username == "" {
			r.SetUsernameAndPassword(string(secret.Data[r.Spec.RootCredentials.UsernameKey]), string(secret.Data[r.Spec.RootCredentials.PasswordKey]))
		} else {
			r.SetUsernameAndPassword(r.Spec.Username, string(secret.Data[r.Spec.RootCredentials.PasswordKey]))
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
		if r.Spec.Username == "" {
			r.SetUsernameAndPassword(secret.Data[r.Spec.RootCredentials.UsernameKey].(string), secret.Data[r.Spec.RootCredentials.PasswordKey].(string))
			log.V(1).Info("", "username", secret.Data[r.Spec.RootCredentials.UsernameKey].(string), "password", secret.Data[r.Spec.RootCredentials.PasswordKey].(string))
		} else {
			r.SetUsernameAndPassword(r.Spec.Username, secret.Data[r.Spec.RootCredentials.PasswordKey].(string))
			log.V(1).Info("", "username", r.Spec.Username, "password", secret.Data[r.Spec.RootCredentials.PasswordKey].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

type DBSEConfig struct {

	// PluginName Specifies the name of the plugin to use for this connection.
	// +kubebuilder:validation:Required
	PluginName string `json:"pluginName,omitempty"`

	// VerifyConnection Specifies if the connection is verified during initial configuration. Defaults to true.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	VerifyConnection bool `json:"verifyConnection,omitempty"`

	// AllowedRoles List of the roles allowed to use this connection. Defaults to empty (no roles), if contains a "*" any role can use this connection.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={"*"}
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AllowedRoles []string `json:"allowedRoles,omitempty"`

	// RootRotationStatements Specifies the database statements to be executed to rotate the root user's credentials. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RootRotationStatements []string `json:"rootRotationStatements,omitempty"`

	// PasswordPolicy The name of the password policy to use when generating passwords for this database. If not specified, this will use a default policy defined as: 20 characters with at least 1 uppercase, 1 lowercase, 1 number, and 1 dash character.
	// +kubebuilder:validation:Optional
	PasswordPolicy string `json:"passwordPolicy,omitempty"`

	// ConnectionURL Specifies the connection string used to connect to the database. Some plugins use url rather than connection_url. This allows for simple templating of the username and password of the root user. Typically, this is done by including a "{{"username"}}", "{{"name"}}", and/or "{{"password"}}" field within the string. These fields are typically be replaced with the values in the username and password fields.
	// +kubebuilder:validation:Required
	ConnectionURL string `json:"connectionURL,omitempty"`

	// Username Specifies the name of the user to use as the "root" user when connecting to the database. This "root" user is used to create/update/delete users managed by these plugins, so you will need to ensure that this user has permissions to manipulate users appropriate to the database. This is typically used in the connection_url field via the templating directive "{{"username"}}" or "{{"name"}}"
	// If username is provided it takes precedence over the username retrieved from the referenced secrets
	// +kubebuilder:validation:Optional
	Username string `json:"username,omitempty"`

	// DatabaseSpecificConfig this are the configuraiton specific to each database type
	// +kubebuilder:validation:Optional
	// +mapType=granular
	DatabaseSpecificConfig map[string]string `json:"databaseSpecificConfig,omitempty"`

	retrievedPassword string `json:"-"`

	retrievedUsername string `json:"-"`
}

// DatabaseSecretEngineConfigStatus defines the observed state of DatabaseSecretEngineConfig
type DatabaseSecretEngineConfigStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ apis.ConditionsAware = &DatabaseSecretEngineConfig{}

func (m *DatabaseSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *DatabaseSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (m *DatabaseSecretEngineConfig) SetUsernameAndPassword(username string, password string) {
	m.Spec.DBSEConfig.retrievedUsername = username
	m.Spec.DBSEConfig.retrievedPassword = password
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DatabaseSecretEngineConfig is the Schema for the databasesecretengineconfigs API
type DatabaseSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status DatabaseSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseSecretEngineConfigList contains a list of DatabaseSecretEngineConfig
type DatabaseSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseSecretEngineConfig{}, &DatabaseSecretEngineConfigList{})
}

func (i *DBSEConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["plugin_name"] = i.PluginName
	payload["verify_connection"] = i.VerifyConnection
	payload["allowed_roles"] = i.AllowedRoles
	payload["root_rotation_statements"] = i.RootRotationStatements
	payload["password_policy"] = i.PasswordPolicy
	payload["connection_url"] = i.ConnectionURL
	for key, value := range i.DatabaseSpecificConfig {
		payload[key] = value
	}
	payload["username"] = i.retrievedUsername
	payload["password"] = i.retrievedPassword
	return payload
}

func (r *DatabaseSecretEngineConfig) isValid() error {
	return r.Spec.RootCredentials.validateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}
