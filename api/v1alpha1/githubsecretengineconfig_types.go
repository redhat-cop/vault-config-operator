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
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitHubSecretEngineConfigSpec defines the desired state of GitHubSecretEngineConfig
type GitHubSecretEngineConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

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

	GHConfig `json:",inline"`

	// SSHKeyReference allows ofr options to retrieve the ssh key. For security reasons it is never displayed.
	// +kubebuilder:validation:Required
	SSHKeyReference SSHKeyConfig `json:"sSHKeyReference,omitempty"`
}

type GHConfig struct {
	// ApplicationID the Application ID of the GitHub App.
	// +kubebuilder:validation:Required
	ApplicationID int64 `json:"applicationID,omitempty"`

	// GitHubAPIBaseURL the base URL for API requests (defaults to the public GitHub API).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="https://api.github.com"
	GitHubAPIBaseURL string `json:"gitHubAPIBaseURL,omitempty"`

	retrievedSSHKey string `json:"-"`
}

func (i *GHConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["app_id"] = i.ApplicationID
	payload["prv_key"] = i.retrievedSSHKey
	payload["base_url"] = i.GitHubAPIBaseURL
	return payload
}

type SSHKeyConfig struct {
	// VaultSecret retrieves the sshkey from a Vault secret. The sshkey will be retrieve at the key "key" (pun intented).
	// +kubebuilder:validation:Optional
	VaultSecret *vaultutils.VaultSecretReference `json:"vaultSecret,omitempty"`

	// Secret retrieves the ssh key from a Kubernetes secret. The secret must be of ssh type (https://kubernetes.io/docs/concepts/configuration/secret/#ssh-authentication-secrets).
	// +kubebuilder:validation:Optional
	Secret *corev1.LocalObjectReference `json:"secret,omitempty"`
}

var _ vaultutils.VaultObject = &GitHubSecretEngineConfig{}

func (d *GitHubSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *GitHubSecretEngineConfig) GetPath() string {
	return string(d.Spec.Path) + "/" + "config"
}
func (d *GitHubSecretEngineConfig) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *GitHubSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.GHConfig.toMap()
	delete(desiredState, "prv_key")
	return reflect.DeepEqual(desiredState, payload)
}

func (d *GitHubSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *GitHubSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (r *GitHubSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *GitHubSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
	vaultClient := context.Value("vaultClient").(*vault.Client)
	if r.Spec.SSHKeyReference.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.SSHKeyReference.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		if secret.Type != corev1.SecretTypeSSHAuth {
			err := errors.New("secret must be of type: " + string(corev1.SecretTypeSSHAuth))
			log.Error(err, "wrong ", "secret type", secret.Type)
			return err
		}
		r.Spec.retrievedSSHKey = string(secret.Data[corev1.SSHAuthPrivateKey])
		return nil
	}
	if r.Spec.SSHKeyReference.VaultSecret != nil {
		secret, err := vaultClient.Logical().Read(string(r.Spec.SSHKeyReference.VaultSecret.Path))
		if err != nil {
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		r.Spec.retrievedSSHKey = secret.Data["key"].(string)
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

// GitHubSecretEngineConfigStatus defines the observed state of GitHubSecretEngineConfig
type GitHubSecretEngineConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

var _ vaultresourcecontroller.ConditionsAware = &GitHubSecretEngineConfig{}

func (m *GitHubSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *GitHubSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GitHubSecretEngineConfig is the Schema for the githubsecretengineconfigs API
type GitHubSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitHubSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status GitHubSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitHubSecretEngineConfigList contains a list of GitHubSecretEngineConfig
type GitHubSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitHubSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitHubSecretEngineConfig{}, &GitHubSecretEngineConfigList{})
}

func (r *GitHubSecretEngineConfig) isValid() error {
	return r.validateEitherFromVaultSecretOrFromSecret()
}

func (r *GitHubSecretEngineConfig) validateEitherFromVaultSecretOrFromSecret() error {
	count := 0
	if r.Spec.SSHKeyReference.Secret != nil {
		count++
	}
	if r.Spec.SSHKeyReference.VaultSecret != nil {
		count++
	}
	if count != 1 {
		return errors.New("only one of spec.sSHKeyReference.vaultSecret or spec.sSHKeyReference.secret can be specified")
	}
	return nil
}

func (d *GitHubSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
