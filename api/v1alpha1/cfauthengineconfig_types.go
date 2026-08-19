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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CFAuthEngineConfigSpec defines the desired state of CFAuthEngineConfig
type CFAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the CF auth engine is mounted.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/config.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	CFAuthConfig `json:",inline"`

	// CFCredentials is used to provide the CF API username and password.
	// The credentials can be sourced from a Kubernetes Secret, VaultSecret, or RandomSecret.
	// +kubebuilder:validation:Optional
	CFCredentials vaultutils.RootCredentialConfig `json:"cfCredentials,omitempty"`
}

type CFAuthConfig struct {
	// CFUsername is the CF API username. Required when using randomSecret credentials
	// (randomSecret only provides the password). Use spec.cfUsername (not spec.username)
	// to supply the username for Cloud Foundry. When using Secret or VaultSecret
	// credentials, the username is retrieved from the secret using usernameKey instead.
	// +kubebuilder:validation:Optional
	CFUsername string `json:"cfUsername,omitempty"`

	// IdentityCACertificates is the root CA certificate(s) for verifying CF_INSTANCE_CERT.
	// +kubebuilder:validation:Required
	// +listType=set
	IdentityCACertificates []string `json:"identityCACertificates"`

	// CFAPIAddr is the full API address of the CF deployment.
	// +kubebuilder:validation:Required
	CFAPIAddr string `json:"cfAPIAddr"`

	// CFAPITrustedCertificates is the certificate(s) presented by the CF API.
	// +kubebuilder:validation:Optional
	// +listType=set
	CFAPITrustedCertificates []string `json:"cfAPITrustedCertificates,omitempty"`

	// LoginMaxSecondsNotBefore is the max seconds in the past for signature creation.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=300
	LoginMaxSecondsNotBefore int `json:"loginMaxSecondsNotBefore"`

	// LoginMaxSecondsNotAfter is the max seconds in the future for signature creation.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=60
	LoginMaxSecondsNotAfter int `json:"loginMaxSecondsNotAfter"`

	// CFAPIMutualTLSCertificate is the client certificate for mutual TLS with the CF API.
	// +kubebuilder:validation:Optional
	CFAPIMutualTLSCertificate string `json:"cfAPIMutualTLSCertificate,omitempty"`

	// CFAPIMutualTLSKey is the client key for mutual TLS with the CF API. Write-only.
	// +kubebuilder:validation:Optional
	CFAPIMutualTLSKey string `json:"cfAPIMutualTLSKey,omitempty"`

	retrievedCFUsername string `json:"-"`
	retrievedCFPassword string `json:"-"`
}

// CFAuthEngineConfigStatus defines the observed state of CFAuthEngineConfig
type CFAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CFAuthEngineConfig is the Schema for the cfauthengineconfigs API
type CFAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CFAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status CFAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CFAuthEngineConfigList contains a list of CFAuthEngineConfig
type CFAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CFAuthEngineConfig `json:"items"`
}

var _ vaultutils.VaultObject = &CFAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &CFAuthEngineConfig{}

func (d *CFAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *CFAuthEngineConfig) IsDeletable() bool {
	return true
}

func (r *CFAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *CFAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *CFAuthEngineConfig) SetCFUsernameAndPassword(username string, password string) {
	r.Spec.CFAuthConfig.retrievedCFUsername = username
	r.Spec.CFAuthConfig.retrievedCFPassword = password
}

func (r *CFAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config")
}

func (r *CFAuthEngineConfig) GetPayload() map[string]any {
	return r.Spec.CFAuthConfig.toMap()
}

func (r *CFAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.CFAuthConfig.toMap()
	delete(desiredState, "cf_password")
	delete(desiredState, "cf_api_mutual_tls_key")
	removeUnsetFields(desiredState, payload)
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	for _, key := range []string{"identity_ca_certificates", "cf_api_trusted_certificates"} {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (r *CFAuthEngineConfig) IsInitialized() bool {
	return true
}

func (r *CFAuthEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *CFAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *CFAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return r.setInternalCredentials(context)
}

func (r *CFAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *CFAuthEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)
	if r.Spec.CFCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.CFCredentials.RandomSecret.Name,
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
			passwordVal, ok := actualData[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret KV v2 key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.SetCFUsernameAndPassword(r.Spec.CFUsername, passwordVal)
		} else {
			passwordVal, ok := secret.Data[randomSecret.Spec.SecretKey].(string)
			if !ok {
				return fmt.Errorf("RandomSecret key %q not found or not a string", randomSecret.Spec.SecretKey)
			}
			r.SetCFUsernameAndPassword(r.Spec.CFUsername, passwordVal)
		}

		return nil
	}
	if r.Spec.CFCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.CFCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		passwordBytes, ok := secret.Data[r.Spec.CFCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.CFCredentials.Secret.Name, r.Spec.CFCredentials.PasswordKey)
		}
		usernameBytes, ok := secret.Data[r.Spec.CFCredentials.UsernameKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", r.Spec.CFCredentials.Secret.Name, r.Spec.CFCredentials.UsernameKey)
		}
		r.SetCFUsernameAndPassword(string(usernameBytes), string(passwordBytes))
		return nil
	}
	if r.Spec.CFCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.CFCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		passwordVal, ok := secret.Data[r.Spec.CFCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.CFCredentials.PasswordKey)
		}
		usernameVal, ok := secret.Data[r.Spec.CFCredentials.UsernameKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", r.Spec.CFCredentials.UsernameKey)
		}
		r.SetCFUsernameAndPassword(usernameVal, passwordVal)
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *CFAuthEngineConfig) isValid() error {
	if err := r.Spec.CFCredentials.ValidateCredentialSource(); err != nil {
		return err
	}
	if r.Spec.CFCredentials.RandomSecret != nil && r.Spec.CFUsername == "" {
		return errors.New("spec.cfUsername must be set when using randomSecret credentials (randomSecret only provides the password)")
	}
	return nil
}

func (c *CFAuthConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["identity_ca_certificates"] = toInterfaceArray(c.IdentityCACertificates)
	payload["cf_api_addr"] = c.CFAPIAddr
	if c.retrievedCFUsername != "" {
		payload["cf_username"] = c.retrievedCFUsername
	} else if c.CFUsername != "" {
		payload["cf_username"] = c.CFUsername
	}
	payload["cf_password"] = c.retrievedCFPassword
	payload["cf_api_trusted_certificates"] = toInterfaceArray(c.CFAPITrustedCertificates)
	payload["login_max_seconds_not_before"] = json.Number(strconv.Itoa(c.LoginMaxSecondsNotBefore))
	payload["login_max_seconds_not_after"] = json.Number(strconv.Itoa(c.LoginMaxSecondsNotAfter))
	payload["cf_api_mutual_tls_certificate"] = c.CFAPIMutualTLSCertificate
	payload["cf_api_mutual_tls_key"] = c.CFAPIMutualTLSKey
	return payload
}

func init() {
	SchemeBuilder.Register(&CFAuthEngineConfig{}, &CFAuthEngineConfigList{})
}
