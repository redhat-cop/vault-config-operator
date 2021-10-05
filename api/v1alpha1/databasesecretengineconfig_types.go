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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseSecretEngineConfigSpec defines the desired state of DatabaseSecretEngineConfig
type DatabaseSecretEngineConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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
	// +kubebuilder:validation:UniqueItems=true
	AllowedRoles []string `json:"allowedRoles,omitempty"`

	// RootRotationStatements Specifies the database statements to be executed to rotate the root user's credentials. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// +kubebuilder:validation:UniqueItems=true
	RootRotationStatements []string `json:"rootRotationStatements,omitempty"`

	// PasswordPolicy The name of the password policy to use when generating passwords for this database. If not specified, this will use a default policy defined as: 20 characters with at least 1 uppercase, 1 lowercase, 1 number, and 1 dash character.
	// +kubebuilder:validation:Optional
	PasswordPolicy string `json:"passwordPolicy,omitempty"`

	// ConnecectionURL Specifies the connection string used to connect to the database. Some plugins use url rather than connection_url. This allows for simple templating of the username and password of the root user. Typically, this is done by including a {{username}}, {{name}}, and/or {{password}} field within the string. These fields are typically be replaced with the values in the username and password fields.
	// +kubebuilder:validation:Required
	ConnecectionURL string `json:"connecectionURL,omitempty"`

	//ConfiguationFromSecret retrieves the credentials from a Vault secret. This will map the "username" and "password" keys of the secret to the username and password of this config. All other keys will be ignored. Only one of RootCredentialsFromVaultSecret or RootCredentialsFromSecret can be specified.
	// username: Specifies the name of the user to use as the "root" user when connecting to the database. This "root" user is used to create/update/delete users managed by these plugins, so you will need to ensure that this user has permissions to manipulate users appropriate to the database. This is typically used in the connection_url field via the templating directive {{username}} or {{name}}.
	// password: Specifies the password to use when connecting with the username. This value will not be returned by Vault when performing a read upon the configuration. This is typically used in the connection_url field via the templating directive {{password}}.
	// +kubebuilder:validation:Optional
	RootCredentialsFromVaultSecret VaultSecretReference `json:"rootCredentialsFromVaultSecret,omitempty"`

	// ConfiguationFromSecret retrieves the credentials from a Kubernetes secret. The secret must be of basicauth type (https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). This will map the "username" and "password" keys of the secret to the username and password of this config. All other keys will be ignored. Only one of RootCredentialsFromVaultSecret or RootCredentialsFromSecret can be specified.
	// username: Specifies the name of the user to use as the "root" user when connecting to the database. This "root" user is used to create/update/delete users managed by these plugins, so you will need to ensure that this user has permissions to manipulate users appropriate to the database. This is typically used in the connection_url field via the templating directive {{username}} or {{name}}.
	// password: Specifies the password to use when connecting with the username. This value will not be returned by Vault when performing a read upon the configuration. This is typically used in the connection_url field via the templating directive {{password}}.
	// +kubebuilder:validation:Optional
	RootCredentialsFromSecret corev1.LocalObjectReference `json:"rootCredentialsFromSecret,omitempty"`

	// DatabaseSpecificConfig this are the configuraiton specific to each database type
	// +kubebuilder:validation:Optional
	// +mapType=granular
	DatabaseSpecificConfig map[string]string `json:"databaseSpecificConfig,omitempty"`

	//internal field on the username
	Username string `json:"-"`
}

// DatabaseSecretEngineConfigStatus defines the observed state of DatabaseSecretEngineConfig
type DatabaseSecretEngineConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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

func DatabaseSecretEngineConfigSpecFromMap(payload map[string]interface{}) *DatabaseSecretEngineConfigSpec {
	o := &DatabaseSecretEngineConfigSpec{}
	o.PluginName = payload["plugin_name"].(string)
	o.VerifyConnection = payload["verify_connection"].(bool)
	o.AllowedRoles = payload["allowed_roles"].([]string)
	o.RootRotationStatements = payload["root_rotation_statements"].([]string)
	o.PasswordPolicy = payload["password_policy"].(string)
	o.ConnecectionURL = payload["connection_url"].(string)
	o.Username = payload["username"].(string)
	return o
}

func (i *DatabaseSecretEngineConfigSpec) ToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["plugin_name"] = i.PluginName
	payload["verify_connection"] = i.VerifyConnection
	payload["allowed_roles"] = i.AllowedRoles
	payload["root_rotation_statements"] = i.RootRotationStatements
	payload["password_policy"] = i.PasswordPolicy
	payload["connection_url"] = i.ConnecectionURL
	payload["username"] = i.Username
	for key, value := range i.DatabaseSpecificConfig {
		payload[key] = value
	}
	return payload
}
