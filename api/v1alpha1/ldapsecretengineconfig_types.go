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

// LDAPSecretEngineConfigSpec defines the desired state of LDAPSecretEngineConfig
type LDAPSecretEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/config.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	LDAPSEConfig `json:",inline"`

	// BindCredentials is used to connect to the LDAP service on the specified LDAP Server.
	// BindCredentials consists in bindDN and bindPass, which can be created as Kubernetes Secret, VaultSecret or RandomSecret.
	// +kubebuilder:validation:Optional
	BindCredentials vaultutils.RootCredentialConfig `json:"bindCredentials,omitempty"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type LDAPSEConfig struct {
	// URL is the LDAP server to connect to. Examples: ldaps://ldap.myorg.com, ldaps://ldap.myorg.com:636
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="ldap://127.0.0.1"
	URL string `json:"url" vault:"url"`

	// Schema is the LDAP schema to use when storing entry passwords
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="openldap"
	// +kubebuilder:validation:Enum:={"openldap","ad","racf"}
	Schema string `json:"schema,omitempty" vault:"schema"`

	// BindDN is the Distinguished name of object to bind for managing user entries
	// +kubebuilder:validation:Optional
	BindDN string `json:"bindDN,omitempty" vault:"binddn,omitempty"`

	// PasswordPolicy is the name of the password policy to use for generating passwords
	// +kubebuilder:validation:Optional
	PasswordPolicy string `json:"passwordPolicy,omitempty" vault:"password_policy"`

	// UserDN is the base DN under which to perform user search
	// +kubebuilder:validation:Optional
	UserDN string `json:"userDN,omitempty" vault:"userdn"`

	// UserAttr is the attribute field name used to perform user search
	// +kubebuilder:validation:Optional
	UserAttr string `json:"userAttr,omitempty" vault:"userattr"`

	// UPNDomain is used to construct a UPN string for Active Directory
	// +kubebuilder:validation:Optional
	UPNDomain string `json:"upnDomain,omitempty" vault:"upndomain"`

	// RequestTimeout is timeout in seconds for the connection
	// +kubebuilder:validation:Optional
	RequestTimeout string `json:"requestTimeout,omitempty" vault:"request_timeout"`

	// StartTLS issues a StartTLS command after establishing an unencrypted connection
	// +kubebuilder:validation:Optional
	StartTLS bool `json:"startTLS,omitempty" vault:"starttls"`

	// InsecureTLS skips LDAP server SSL certificate verification
	// +kubebuilder:validation:Optional
	InsecureTLS bool `json:"insecureTLS,omitempty" vault:"insecure_tls"`

	// Certificate is the CA certificate for verifying LDAP server certificate (PEM)
	// +kubebuilder:validation:Optional
	Certificate string `json:"certificate,omitempty" vault:"certificate"`

	// ClientTLSCert is the client certificate for LDAP (PEM)
	// +kubebuilder:validation:Optional
	ClientTLSCert string `json:"clientTLSCert,omitempty" vault:"client_tls_cert"`

	// ClientTLSKey is the client key for LDAP (PEM)
	// +kubebuilder:validation:Optional
	ClientTLSKey string `json:"clientTLSKey,omitempty" vault:"client_tls_key"`

	// SkipStaticRoleImportRotation is the default value for skip_import_rotation on static roles
	// +kubebuilder:validation:Optional
	SkipStaticRoleImportRotation bool `json:"skipStaticRoleImportRotation,omitempty" vault:"skip_static_role_import_rotation"`

	// CredentialType is the type of password to generate (password or phrase for RACF)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"password","phrase"}
	CredentialType string `json:"credentialType,omitempty" vault:"credential_type,omitempty"`

	// Length is the generated password string length (deprecated: use passwordPolicy)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	Length *int `json:"length,omitempty" vault:"length,number,omitempty"`

	// ConnectionTimeout is timeout before trying the next URL (deprecated: use requestTimeout)
	// +kubebuilder:validation:Optional
	ConnectionTimeout string `json:"connectionTimeout,omitempty" vault:"connection_timeout"`

	retrievedBindDN   string `json:"-"`
	retrievedBindPass string `json:"-"`
}

var _ vaultutils.VaultObject = &LDAPSecretEngineConfig{}
var _ vaultutils.ConditionsAware = &LDAPSecretEngineConfig{}

func (d *LDAPSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *LDAPSecretEngineConfig) IsDeletable() bool {
	return true
}

func (d *LDAPSecretEngineConfig) GetPath() string {
	return vaultutils.CleansePath(string(d.Spec.Path) + "/config")
}

func (d *LDAPSecretEngineConfig) GetPayload() map[string]any {
	return d.Spec.LDAPSEConfig.toMap()
}

func (d *LDAPSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.LDAPSEConfig.toMap()
	delete(desiredState, "bindpass")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *LDAPSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (r *LDAPSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *LDAPSecretEngineConfig) isValid() error {
	if err := r.Spec.BindCredentials.ValidateCredentialSource(); err != nil {
		return err
	}
	if r.Spec.BindCredentials.RandomSecret != nil && r.Spec.LDAPSEConfig.BindDN == "" {
		return errors.New("spec.bindDN must be set when using randomSecret credentials (randomSecret only provides the bind password)")
	}
	return nil
}

func (r *LDAPSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if r.Spec.BindCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.BindCredentials.RandomSecret.Name,
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

		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		secretVal, ok := secretData[randomSecret.Spec.SecretKey].(string)
		if !ok {
			return fmt.Errorf("secret key %q not found or not a string in vault secret", randomSecret.Spec.SecretKey)
		}
		r.SetBindDNAndBindPass(r.Spec.LDAPSEConfig.BindDN, secretVal)
		return nil
	}
	if r.Spec.BindCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.BindCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		passwordBytes, ok := secret.Data[r.Spec.BindCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.BindCredentials.Secret.Name, r.Spec.BindCredentials.PasswordKey)
		}
		if r.Spec.LDAPSEConfig.BindDN == "" {
			usernameBytes, ok := secret.Data[r.Spec.BindCredentials.UsernameKey]
			if !ok {
				return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.BindCredentials.Secret.Name, r.Spec.BindCredentials.UsernameKey)
			}
			r.SetBindDNAndBindPass(string(usernameBytes), string(passwordBytes))
		} else {
			r.SetBindDNAndBindPass(r.Spec.LDAPSEConfig.BindDN, string(passwordBytes))
		}
		return nil
	}
	if r.Spec.BindCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.BindCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		passwordVal, ok := secretData[r.Spec.BindCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.BindCredentials.PasswordKey)
		}
		if r.Spec.LDAPSEConfig.BindDN == "" {
			usernameVal, ok := secretData[r.Spec.BindCredentials.UsernameKey].(string)
			if !ok {
				return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.BindCredentials.UsernameKey)
			}
			r.SetBindDNAndBindPass(usernameVal, passwordVal)
		} else {
			r.SetBindDNAndBindPass(r.Spec.LDAPSEConfig.BindDN, passwordVal)
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *LDAPSecretEngineConfig) SetBindDNAndBindPass(bindDN string, bindPass string) {
	r.Spec.LDAPSEConfig.retrievedBindDN = bindDN
	r.Spec.LDAPSEConfig.retrievedBindPass = bindPass
}

func (i *LDAPSEConfig) toMap() map[string]any {
	payload := vaultutils.StructToVaultPayload(i)
	if i.retrievedBindDN != "" {
		payload["binddn"] = i.retrievedBindDN
	}
	if i.retrievedBindPass != "" {
		payload["bindpass"] = i.retrievedBindPass
	}
	return payload
}

// LDAPSecretEngineConfigStatus defines the observed state of LDAPSecretEngineConfig
type LDAPSecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LDAPSecretEngineConfig is the Schema for the ldapsecretengineconfigs API
type LDAPSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LDAPSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status LDAPSecretEngineConfigStatus `json:"status,omitempty"`
}

func (m *LDAPSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *LDAPSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// LDAPSecretEngineConfigList contains a list of LDAPSecretEngineConfig
type LDAPSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LDAPSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LDAPSecretEngineConfig{}, &LDAPSecretEngineConfigList{})
}

func (d *LDAPSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
