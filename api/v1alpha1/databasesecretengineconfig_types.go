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

	vault "github.com/hashicorp/vault/api"
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

	// +kubebuilder:validation:Required
	DBSEConfig `json:",inline"`

	// RootCredentials specifies how to retrieve the credentials for this DatabaseEngine connection.
	// +kubebuilder:validation:Required
	RootCredentials vaultutils.RootCredentialConfig `json:"rootCredentials,omitempty"`
}

var _ vaultutils.VaultObject = &DatabaseSecretEngineConfig{}

func (d *DatabaseSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *DatabaseSecretEngineConfig) GetPath() string {
	return string(d.Spec.Path) + "/" + "config" + "/" + d.Name
}
func (d *DatabaseSecretEngineConfig) GetRootPasswordRotationPath() string {
	return string(d.Spec.Path) + "/" + "rotate-root" + "/" + d.Name
}
func (d *DatabaseSecretEngineConfig) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *DatabaseSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.DBSEConfig.toMap()
	connectionDetails := map[string]interface{}{}
	connectionDetails["connection_url"] = desiredState["connection_url"]
	connectionDetails["disable_escaping"] = desiredState["disable_escaping"]
	connectionDetails["root_credentials_rotate_statements"] = desiredState["root_credentials_rotate_statements"]
	connectionDetails["username"] = desiredState["username"]
	if desiredState["verify_connection"] == true {
		connectionDetails["verify_connection"] = desiredState["verify_connection"]
	}
	desiredState["connection_details"] = connectionDetails
	//delete fields that have been moved to connection_details
	delete(desiredState, "password")
	delete(desiredState, "connection_url")
	delete(desiredState, "username")
	delete(desiredState, "verify_connection")
	delete(desiredState, "disable_escaping")

	return reflect.DeepEqual(desiredState, payload)
}

func toInterfaceArray(values []string) []interface{} {
	result := []interface{}{}
	for _, value := range values {
		result = append(result, value)
	}
	return result
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
	if !r.Status.LastRootPasswordRotation.IsZero() {
		log.V(1).Info("root credentials rotation already occurred - credentials retrieval skipped")
		return nil
	}

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

		if randomSecret.Spec.IsKVSecretsEngineV2 {
			var actualData map[string]interface{} = secret.Data["data"].(map[string]interface{})
			r.SetUsernameAndPassword(r.Spec.Username, (actualData[randomSecret.Spec.SecretKey]).(string))
		} else {
			r.SetUsernameAndPassword(r.Spec.Username, secret.Data[randomSecret.Spec.SecretKey].(string))
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

	// PluginVersion Specifies the semantic version of the plugin to use for this connection.
	// +kubebuilder:validation:Optional
	PluginVersion string `json:"pluginVersion,omitempty"`

	// VerifyConnection Specifies if the connection is verified during initial configuration. Defaults to true.
	// +kubebuilder:validation:Optional
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

	// DisableEscaping Determines whether special characters in the username and password fields will be escaped. Useful for alternate connection string formats like ADO. More information regarding this parameter can be found on the databases secrets engine docs. Defaults to false
	// +kubebuilder:validation:Optional
	DisableEscaping bool `json:"disableEscaping,omitempty"`

	// DatabaseSpecificConfig this are the configuration specific to each database type
	// +kubebuilder:validation:Optional
	// +mapType=granular
	DatabaseSpecificConfig map[string]string `json:"databaseSpecificConfig,omitempty"`

	retrievedPassword string `json:"-"`

	retrievedUsername string `json:"-"`

	// +kubebuilder:validation:Optional
	RootPasswordRotation *RootPasswordRotation `json:"rootPasswordRotation,omitempty"`
}

type RootPasswordRotation struct {
	// Enabled whether the toot password should be rotated with the rotation statement. If set to true the root password will be rotated immediately.
	// +kubebuilder:validation:Optional
	Enable bool `json:"enable,omitempty"`
	// RotationPeriod if this value is set, the root password will be rotated approximately with teh requested frequency.
	// +kubebuilder:validation:Optional
	RotationPeriod metav1.Duration `json:"rotationPeriod,omitempty"`
}

// DatabaseSecretEngineConfigStatus defines the observed state of DatabaseSecretEngineConfig
type DatabaseSecretEngineConfigStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// +kubebuilder:validation:Optional
	LastRootPasswordRotation metav1.Time `json:"lastRootPasswordRotation,omitempty"`
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
	payload["plugin_version"] = i.PluginVersion
	payload["verify_connection"] = i.VerifyConnection

	payload["allowed_roles"] = toInterfaceArray(i.AllowedRoles)
	payload["root_credentials_rotate_statements"] = toInterfaceArray(i.RootRotationStatements)
	payload["password_policy"] = i.PasswordPolicy
	payload["connection_url"] = i.ConnectionURL
	for key, value := range i.DatabaseSpecificConfig {
		payload[key] = value
	}
	if i.Username != "" {
		payload["username"] = i.Username
	} else if i.retrievedUsername != "" { // Only set the username in payload if retrieved - see setInternalCredentials()
		payload["username"] = i.retrievedUsername
	}
	payload["disable_escaping"] = i.DisableEscaping
	if i.retrievedPassword != "" { // Only set the password in payload if retrieved - see setInternalCredentials()
		payload["password"] = i.retrievedPassword
	}

	return payload
}

func (r *DatabaseSecretEngineConfig) isValid() error {
	return r.Spec.RootCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}

func (d *DatabaseSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *DatabaseSecretEngineConfig) RotateRootPassword(ctx context.Context) error {
	log := log.FromContext(ctx)
	vaultClient := ctx.Value("vaultClient").(*vault.Client)
	_, err := vaultClient.Logical().WriteWithContext(ctx, d.GetRootPasswordRotationPath(), nil)
	if err != nil {
		log.Error(err, "unable to rotate root password", "instance", d)
		return err
	}
	return nil
}
