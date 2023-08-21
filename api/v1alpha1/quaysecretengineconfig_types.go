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

// QuaySecretEngineConfigSpec defines the desired state of QuaySecretEngineConfig
type QuaySecretEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	QuayConfig `json:",inline"`

	// RootCredentials specifies how to retrieve the credentials for this Quay connection.
	// +kubebuilder:validation:Required
	RootCredentials vaultutils.RootCredentialConfig `json:"rootCredentials,omitempty"`
}

var _ vaultutils.VaultObject = &QuaySecretEngineConfig{}

func (d *QuaySecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (q *QuaySecretEngineConfig) GetPath() string {
	return string(q.Spec.Path) + "/" + "config"
}

func (q *QuaySecretEngineConfig) GetPayload() map[string]interface{} {
	return q.Spec.toMap()
}

func (q *QuaySecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := q.Spec.QuayConfig.toMap()
	delete(desiredState, "password")
	return reflect.DeepEqual(desiredState, payload)
}

func (q *QuaySecretEngineConfig) IsInitialized() bool {
	return true
}

func (q *QuaySecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return q.setInternalCredentials(context)
}

func (q *QuaySecretEngineConfig) IsValid() (bool, error) {
	err := q.isValid()
	return err == nil, err
}

func (q *QuaySecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
	if q.Spec.RootCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: q.Namespace,
			Name:      q.Spec.RootCredentials.RandomSecret.Name,
		}, randomSecret)
		if err != nil {
			log.Error(err, "unable to retrieve RandomSecret", "instance", q)
			return err
		}
		secret, exists, err := vaultutils.ReadSecret(context, randomSecret.GetPath())
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", q)
			return err
		}
		q.SetToken(secret.Data[randomSecret.Spec.SecretKey].(string))
		return nil
	}
	if q.Spec.RootCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: q.Namespace,
			Name:      q.Spec.RootCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", q)
			return err
		}
		q.SetToken(string(secret.Data[q.Spec.RootCredentials.PasswordKey]))
		return nil
	}
	if q.Spec.RootCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(q.Spec.RootCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", q)
			return err
		}
		q.SetToken(secret.Data[q.Spec.RootCredentials.PasswordKey].(string))
		log.V(1).Info("", "token", secret.Data[q.Spec.RootCredentials.PasswordKey].(string))
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

type QuayConfig struct {

	// url Specifies the location of the Quay instance
	// +kubebuilder:validation:Required
	URL string `json:"url,omitempty"`

	// CACertertificate PEM encoded CA cert for use by the TLS client used to communicate with Quay.
	// +kubebuilder:validation:Optional
	CACertertificate string `json:"caCertificate,omitempty"`

	// DisableSslVerification Disable SSL verification when communicating with Quay.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	DisableSslVerification bool `json:"disableSslVerification,omitempty"`

	retrievedToken string `json:"-"`
}

// QuaySecretEngineConfigStatus defines the observed state of QuaySecretEngineConfig
type QuaySecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ vaultutils.ConditionsAware = &QuaySecretEngineConfig{}

func (q *QuaySecretEngineConfig) GetConditions() []metav1.Condition {
	return q.Status.Conditions
}

func (q *QuaySecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	q.Status.Conditions = conditions
}

func (q *QuaySecretEngineConfig) SetToken(token string) {
	q.Spec.retrievedToken = token
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// QuaySecretEngineConfig is the Schema for the quaysecretengineconfigs API
type QuaySecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuaySecretEngineConfigSpec   `json:"spec,omitempty"`
	Status QuaySecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuaySecretEngineConfigList contains a list of QuaySecretEngineConfig
type QuaySecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuaySecretEngineConfig `json:"items"`
}

func (i *QuayConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["url"] = i.URL
	payload["token"] = i.retrievedToken
	payload["ca_certificate"] = i.CACertertificate
	payload["disable_ssl_verification"] = i.DisableSslVerification
	return payload
}

func init() {
	SchemeBuilder.Register(&QuaySecretEngineConfig{}, &QuaySecretEngineConfigList{})
}

func (r *QuaySecretEngineConfig) isValid() error {
	return r.Spec.RootCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}

func (d *QuaySecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
