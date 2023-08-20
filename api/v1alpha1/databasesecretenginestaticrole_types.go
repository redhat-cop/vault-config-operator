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
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseSecretEngineStaticRoleSpec defines the desired state of DatabaseSecretEngineStaticRole
type DatabaseSecretEngineStaticRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/roles/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	DBSEStaticRole `json:",inline"`
}

type DBSEStaticRole struct {
	// DBName The name of the database connection to use for this role.
	// +kubebuilder:validation:Required
	DBName string `json:"dBName,omitempty"`

	// Username Specifies the database username that this Vault role corresponds to.
	// +kubebuilder:validation:Required
	Username string `json:"username,omitempty"`

	// RotationPeriod Specifies the amount of time Vault should wait before rotating the password. The minimum is 5 seconds.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=5
	RotationPeriod int `json:"rotationPeriod,omitempty"`

	// RotationStatements Specifies the database statements to be executed to rotate the password for the configured database user. Not every plugin type will support this functionality. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RotationStatements []string `json:"rotationStatements,omitempty"`

	// CredentialType Specifies the type of credential that will be generated for the role. Options include: password, rsa_private_key. See the plugin's API page for credential types supported by individual databases.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"password","rsa_private_key"}
	CredentialType string `json:"credentialType,omitempty"`

	// PasswordCredentialConfig specifies the configuraiton when the password credential type is chosen.
	// +kubebuilder:validation:Optional
	PasswordCredentialConfig *PasswordCredentialConfig `json:"passwordCredentialConfig,omitempty"`

	// RSAPrivateKeyCredentialConfig specifies the configuraiton when the rsa_private_key credential type is chosen.
	// +kubebuilder:validation:Optional

	RSAPrivateKeyCredentialConfig *RSAPrivateKeyCredentialConfig `json:"rsaPrivateKeyCredentialConfig,omitempty"`
}

type PasswordCredentialConfig struct {
	// PasswordPolicy The policy used for password generation. If not provided, defaults to the password policy of the database configuration
	// +kubebuilder:validation:Optional
	PasswordPolicy string `json:"passwordPolicy,omitempty"`
}

func (i *PasswordCredentialConfig) toMap() map[string]string {
	payload := map[string]string{}
	payload["password_policy"] = i.PasswordPolicy
	return payload
}

type RSAPrivateKeyCredentialConfig struct {
	// KeyBits The bit size of the RSA key to generate. Options include: 2048, 3072, 4096.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={2048,3072,4096}
	KeyBits int `json:"keyBits,omitempty"`
	// Format The output format of the generated private key credential. The private key will be returned from the API in PEM encoding. Options include: pkcs8
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"pkcs8"}
	Format string `json:"format,omitempty"`
}

func (i *RSAPrivateKeyCredentialConfig) toMap() map[string]string {
	payload := map[string]string{}
	payload["key_bits"] = strconv.Itoa(i.KeyBits)
	payload["format"] = i.Format
	return payload
}

func (i *DBSEStaticRole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["db_name"] = i.DBName
	payload["username"] = i.Username
	payload["rotation_period"] = i.RotationPeriod
	payload["rotation_statements"] = i.RotationStatements
	payload["credential_type"] = i.CredentialType
	if i.PasswordCredentialConfig != nil {
		payload["credential_config"] = i.PasswordCredentialConfig.toMap()
	}
	if i.RSAPrivateKeyCredentialConfig != nil {
		payload["credential_config"] = i.RSAPrivateKeyCredentialConfig.toMap()
	}
	return payload
}

var _ vaultutils.VaultObject = &DatabaseSecretEngineStaticRole{}

var _ vaultresourcecontroller.ConditionsAware = &DatabaseSecretEngineStaticRole{}

func (d *DatabaseSecretEngineStaticRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *DatabaseSecretEngineStaticRole) GetPath() string {
	return string(d.Spec.Path) + "/" + "static-roles" + "/" + d.Name
}
func (d *DatabaseSecretEngineStaticRole) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *DatabaseSecretEngineStaticRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.DBSEStaticRole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (d *DatabaseSecretEngineStaticRole) IsInitialized() bool {
	return true
}

func (d *DatabaseSecretEngineStaticRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *DatabaseSecretEngineStaticRole) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *DatabaseSecretEngineStaticRole) isValid() error {
	return r.validateEitherPasswordOrKey()
}

func (r *DatabaseSecretEngineStaticRole) validateEitherPasswordOrKey() error {
	count := 0
	if r.Spec.PasswordCredentialConfig != nil {
		count++
	}
	if r.Spec.RSAPrivateKeyCredentialConfig != nil {
		count++
	}
	if count != 1 {
		return errors.New("only one of spec.passwordCredentialConfig or spec.rsaPrivateKeyCredentialConfig can be specified")
	}
	return nil
}

func (d *DatabaseSecretEngineStaticRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

// DatabaseSecretEngineStaticRoleStatus defines the observed state of DatabaseSecretEngineStaticRole
type DatabaseSecretEngineStaticRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *DatabaseSecretEngineStaticRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *DatabaseSecretEngineStaticRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DatabaseSecretEngineStaticRole is the Schema for the databasesecretenginestaticroles API
type DatabaseSecretEngineStaticRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSecretEngineStaticRoleSpec   `json:"spec,omitempty"`
	Status DatabaseSecretEngineStaticRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseSecretEngineStaticRoleList contains a list of DatabaseSecretEngineStaticRole
type DatabaseSecretEngineStaticRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseSecretEngineStaticRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseSecretEngineStaticRole{}, &DatabaseSecretEngineStaticRoleList{})
}
