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

var _ vaultutils.VaultObject = &KubernetesSecretEngineConfig{}
var _ vaultresourcecontroller.ConditionsAware = &KubernetesSecretEngineConfig{}

// KubernetesSecretEngineConfigSpec defines the desired state of KubernetesSecretEngineConfig
type KubernetesSecretEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// JWTReference specifies how to retrieve the JWT token for this Kubernetes Engine connection. Only VaultSecretReference or LocalObjectRefence can be used, random secret is not allowed.
	// +kubebuilder:validation:Required
	JWTReference vaultutils.RootCredentialConfig `json:"jwtReference,omitempty"`

	KubeSEConfig `json:",inline"`
}

// KubernetesSecretEngineConfigStatus defines the observed state of KubernetesSecretEngineConfig
type KubernetesSecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *KubernetesSecretEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *KubernetesSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubernetesSecretEngineConfig is the Schema for the kubernetessecretengineconfigs API
type KubernetesSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status KubernetesSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubernetesSecretEngineConfigList contains a list of KubernetesSecretEngineConfig
type KubernetesSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesSecretEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubernetesSecretEngineConfig{}, &KubernetesSecretEngineConfigList{})
}

func (d *KubernetesSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *KubernetesSecretEngineConfig) GetPath() string {
	return string(d.Spec.Path) + "/" + "config"
}
func (d *KubernetesSecretEngineConfig) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *KubernetesSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.KubeSEConfig.toMap()
	delete(desiredState, "service_account_jwt")
	return reflect.DeepEqual(desiredState, payload)
}

func (d *KubernetesSecretEngineConfig) IsInitialized() bool {
	return true
}

func (d *KubernetesSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setInternalCredentials(context)
}

func (r *KubernetesSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *KubernetesSecretEngineConfig) isValid() error {
	return r.Spec.JWTReference.ValidateEitherFromVaultSecretOrFromSecret()
}

func (r *KubernetesSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
	vaultClient := context.Value("vaultClient").(*vault.Client)
	if r.Spec.JWTReference.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.JWTReference.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		if secret.Type != corev1.SecretTypeServiceAccountToken {
			err := errors.New("secret must be of type: " + string(corev1.SecretTypeServiceAccountToken))
			log.Error(err, "wrong ", "secret type", secret.Type)
			return err
		}
		r.Spec.retrievedServiceAccountJWT = string(secret.Data[corev1.ServiceAccountTokenKey])
		return nil
	}
	if r.Spec.JWTReference.VaultSecret != nil {
		secret, err := vaultClient.Logical().Read(string(r.Spec.JWTReference.VaultSecret.Path))
		if err != nil {
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		r.Spec.retrievedServiceAccountJWT = secret.Data["key"].(string)
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

type KubeSEConfig struct {
	// KubernetesHost Kubernetes API URL to connect to.
	// +kubebuilder:validation:Required
	KubernetesHost string `json:"kubernetesHost,omitempty"`

	// KubernetesCACert PEM encoded CA certificate to verify the Kubernetes API server certificate.
	// +kubebuilder:validation:Optional
	KubernetesCACert string `json:"kubernetesCACert,omitempty"`

	// DisableLocalCAJWT Disable defaulting to the local CA certificate and service account JWT when running in a Kubernetes pod.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	DisableLocalCAJWT bool `json:"disableLocalCAJWT,omitempty"`

	retrievedServiceAccountJWT string `json:"-"`
}

func (i *KubeSEConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["kubernetes_host"] = i.KubernetesHost
	payload["kubernetes_ca_cert"] = i.KubernetesCACert
	payload["service_account_jwt"] = i.retrievedServiceAccountJWT
	payload["disable_local_ca_jwt"] = i.DisableLocalCAJWT
	return payload
}

func (d *KubernetesSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
