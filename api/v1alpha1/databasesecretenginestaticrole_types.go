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
	"errors"
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseSecretEngineStaticRoleSpec defines the desired state of DatabaseSecretEngineStaticRole
type DatabaseSecretEngineStaticRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/roles/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	DBSEStaticRole `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type DBSEStaticRole struct {
	// DBName The name of the database connection to use for this role.
	// +kubebuilder:validation:Optional
	DBName string `json:"dBName,omitempty" vault:"db_name"`

	// Username Specifies the database username that this Vault role corresponds to.
	// +kubebuilder:validation:Optional
	Username string `json:"username,omitempty" vault:"username"`

	// RotationPeriod Specifies the amount of time Vault should wait before rotating the password.
	// The minimum is 5 seconds. Mutually exclusive with RotationSchedule.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=5
	RotationPeriod *int `json:"rotationPeriod,omitempty" vault:"rotation_period,omitempty"`

	// RotationSchedule Specifies a cron-style rotation schedule for the static role.
	// Mutually exclusive with RotationPeriod. Uses standard 5-field cron format (minute, hour,
	// day of month, month, day of week). For example, "0 0 * * SAT" rotates every Saturday at midnight UTC.
	// +kubebuilder:validation:Optional
	RotationSchedule string `json:"rotationSchedule,omitempty" vault:"rotation_schedule,omitempty"`

	// RotationWindow Specifies the maximum duration (in seconds) in which a scheduled rotation
	// is allowed to occur. Only valid when RotationSchedule is set. The minimum is 3600 (1 hour).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=3600
	RotationWindow *int `json:"rotationWindow,omitempty" vault:"rotation_window,omitempty"`

	// RotationStatements Specifies the database statements to be executed to rotate the password for the configured database user. Not every plugin type will support this functionality. See the plugin's API page for more information on support and formatting for this parameter.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	RotationStatements []string `json:"rotationStatements,omitempty" vault:"rotation_statements"`

	// CredentialType Specifies the type of credential that will be generated for the role. Options include: password, rsa_private_key. See the plugin's API page for credential types supported by individual databases.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"password","rsa_private_key"}
	CredentialType string `json:"credentialType,omitempty" vault:"credential_type"`

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
	PasswordPolicy string `json:"passwordPolicy,omitempty" vault:"password_policy"`
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

func (i *DBSEStaticRole) toMap() map[string]any {
	payload := vaultutils.StructToVaultPayload(i)
	if i.PasswordCredentialConfig != nil {
		payload["credential_config"] = i.PasswordCredentialConfig.toMap()
	}
	if i.RSAPrivateKeyCredentialConfig != nil {
		payload["credential_config"] = i.RSAPrivateKeyCredentialConfig.toMap()
	}
	return payload
}

var _ vaultutils.VaultObject = &DatabaseSecretEngineStaticRole{}

var _ vaultutils.ConditionsAware = &DatabaseSecretEngineStaticRole{}

func (d *DatabaseSecretEngineStaticRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *DatabaseSecretEngineStaticRole) IsDeletable() bool {
	return true
}

func (d *DatabaseSecretEngineStaticRole) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "static-roles" + "/" + vaultutils.ResourceName(d.Spec.Name, d.Name))
}
func (d *DatabaseSecretEngineStaticRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (r *DatabaseSecretEngineStaticRole) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *DatabaseSecretEngineStaticRole) isValid() error {
	if err := r.validateRotationConfig(); err != nil {
		return err
	}
	return r.validateEitherPasswordOrKey()
}

func (r *DatabaseSecretEngineStaticRole) validateRotationConfig() error {
	hasPeriod := r.Spec.RotationPeriod != nil
	hasSchedule := r.Spec.RotationSchedule != ""

	if hasPeriod && hasSchedule {
		return errors.New("spec.rotationPeriod and spec.rotationSchedule are mutually exclusive; set one but not both")
	}
	if !hasPeriod && !hasSchedule {
		return errors.New("one of spec.rotationPeriod or spec.rotationSchedule must be set")
	}
	if r.Spec.RotationWindow != nil && !hasSchedule {
		return errors.New("spec.rotationWindow is only valid when spec.rotationSchedule is set")
	}
	return nil
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
