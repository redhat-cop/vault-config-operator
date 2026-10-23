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
	"fmt"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubernetesAuthEngineConfigSpec defines the desired state of KubernetesAuthEngineConfig
type KubernetesAuthEngineConfigSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/auth/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	KAECConfig `json:",inline"`

	// TokenReviewerServiceAccount A service account JWT used to access the TokenReview API to validate other JWTs during login. If not set, the JWT submitted in the login payload will be used to access the Kubernetes TokenReview API.
	// Mutually exclusive with TokenReviewerJWTSecret.
	// +kubebuilder:validation:Optional
	TokenReviewerServiceAccount *corev1.LocalObjectReference `json:"tokenReviewerServiceAccount,omitempty" vault:"-"`

	// TokenReviewerJWTSecret sources the token_reviewer_jwt from a Kubernetes Secret.
	// This is useful when the Kubernetes auth backend's kubernetesHost points at a remote cluster,
	// an authenticating proxy, or any endpoint where a locally-minted ServiceAccount JWT is not valid.
	// The Secret is read from the CR's namespace.
	// Mutually exclusive with TokenReviewerServiceAccount.
	// +kubebuilder:validation:Optional
	TokenReviewerJWTSecret *corev1.SecretKeySelector `json:"tokenReviewerJWTSecret,omitempty" vault:"-"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

func (d *KubernetesAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *KubernetesAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/" + vaultutils.ResourceName(d.Spec.Name, d.Name) + "/config")
}

func (d *KubernetesAuthEngineConfig) GetPayload() map[string]any {
	return d.Spec.KAECConfig.toMap()
}

var _ vaultutils.VaultObject = &KubernetesAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &KubernetesAuthEngineConfig{}

func (d *KubernetesAuthEngineConfig) IsDeletable() bool {
	return false
}

func (d *KubernetesAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	log := log.FromContext(context)

	// TokenReviewerJWTSecret takes precedence over TokenReviewerServiceAccount
	if d.Spec.TokenReviewerJWTSecret != nil {
		kubeClient := vaultutils.KubeClientFromContext(context)
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: d.Namespace,
			Name:      d.Spec.TokenReviewerJWTSecret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve token reviewer JWT secret", "secret", d.Namespace+"/"+d.Spec.TokenReviewerJWTSecret.Name)
			return err
		}
		key := d.Spec.TokenReviewerJWTSecret.Key
		if key == "" {
			key = "token"
		}
		jwt, ok := secret.Data[key]
		if !ok {
			err := fmt.Errorf("key %q not found in secret %s/%s", key, d.Namespace, d.Spec.TokenReviewerJWTSecret.Name)
			log.Error(err, "token reviewer JWT secret key missing")
			return err
		}
		d.Spec.retrievedTokenReviewerJWT = string(jwt)
	} else if d.Spec.TokenReviewerServiceAccount != nil {
		jwt, err := d.getJWTToken(context)
		if err != nil {
			log.Error(err, "unable to retrieve jwt token for service account", "service account", d.Namespace+"/"+d.Spec.TokenReviewerServiceAccount.Name)
			return err
		}
		d.Spec.retrievedTokenReviewerJWT = jwt
	}

	return nil
}

func (kc *KubernetesAuthEngineConfig) getJWTToken(context context.Context) (string, error) {
	expiration := int64(60 * 60 * 24 * 365)
	return vaultutils.GetJWTTokenWithDuration(context, kc.Spec.TokenReviewerServiceAccount.Name, kc.Namespace, expiration)
}

func (r *KubernetesAuthEngineConfig) IsValid() (bool, error) {
	if r.Spec.TokenReviewerServiceAccount != nil && r.Spec.TokenReviewerJWTSecret != nil {
		return false, fmt.Errorf("spec.tokenReviewerServiceAccount and spec.tokenReviewerJWTSecret are mutually exclusive")
	}
	return true, nil
}

type KAECConfig struct {

	// KubernetesHost Host must be a host string, a host:port pair, or a URL to the base of the Kubernetes API server.
	// +kubebuilder:validation:Required
	// +kubebuilder:default="https://kubernetes.default.svc:443"
	KubernetesHost string `json:"kubernetesHost" vault:"kubernetes_host"`

	// kubernetesCACert PEM encoded CA cert for use by the TLS client used to talk with the Kubernetes API. NOTE: Every line must end with a newline: \n
	// if omitted will default to the content of the file "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt" in the operator pod
	// +kubebuilder:validation:Optional
	KubernetesCACert string `json:"kubernetesCACert,omitempty" vault:"kubernetes_ca_cert"`

	// PEMKeys Optional list of PEM-formatted public keys or certificates used to verify the signatures of Kubernetes service account JWTs. If a certificate is given, its public key will be extracted. Not every installation of Kubernetes exposes these keys.
	// +kubebuilder:validation:Optional
	PEMKeys []string `json:"PEMKeys,omitempty" vault:"pem_keys"`

	// Issuer Optional JWT issuer. If no issuer is specified, then this plugin will use kubernetes/serviceaccount as the default issuer. See these instructions for looking up the issuer for a given Kubernetes cluster.
	// +kubebuilder:validation:Optional
	Issuer string `json:"issuer,omitempty" vault:"issuer"`

	// DisableISSValidation Disable JWT issuer validation. Allows to skip ISS validation.
	// +kubebuilder:validation:Optional
	DisableISSValidation bool `json:"disableISSValidation,omitempty" vault:"disable_iss_validation"`

	// DisableLocalCAJWT Disable defaulting to the local CA cert and service account JWT when running in a Kubernetes pod.
	// +kubebuilder:validation:Optional
	DisableLocalCAJWT bool `json:"disableLocalCAJWT,omitempty" vault:"disable_local_ca_jwt"`

	// UseOperatorPodCA . This field is considered only if `kubernetesCACert` is not set and `disableLocalCAJWT` is set to true.
	// In this case if this field is set to true the operator pod's CA is injected. This is the original behavior before the introduction of this field
	// If tis field is set to false, the os ca bundle of where vault is running will be used.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	UseOperatorPodCA bool `json:"useOperatorPodCA" vault:"-"`

	// UseAnnotationsAsAliasMetadata  Use annotations from the client token's associated service account as alias metadata for the Vault entity. Only annotations with the vault.hashicorp.com/alias-metadata- key prefix are targeted as alias metadata and your annotations must be 512 characters or less due to the Vault alias metadata value limit. For example, if you configure the annotation vault.hashicorp.com/alias-metadata-foo, Vault saves the string "foo" along with the annotation value to the alias metadata. To save alias metadata, Vault must have permission to read service accounts from the Kubernetes API.
	// +kubebuilder:validation:Optional
	UseAnnotationsAsAliasMetadata bool `json:"useAnnotationsAsAliasMetadata,omitempty" vault:"use_annotations_as_alias_metadata"`

	retrievedTokenReviewerJWT string `json:"-"`
}

// KubernetesAuthEngineConfigStatus defines the observed state of KubernetesAuthEngineConfig
type KubernetesAuthEngineConfigStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *KubernetesAuthEngineConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *KubernetesAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubernetesAuthEngineConfig is the Schema for the kubernetesauthengineconfigs API
type KubernetesAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status KubernetesAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubernetesAuthEngineConfigList contains a list of KubernetesAuthEngineConfig
type KubernetesAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesAuthEngineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubernetesAuthEngineConfig{}, &KubernetesAuthEngineConfigList{})
}

func (i *KAECConfig) toMap() map[string]any {
	payload := vaultutils.StructToVaultPayload(i)
	if i.retrievedTokenReviewerJWT != "" {
		payload["token_reviewer_jwt"] = i.retrievedTokenReviewerJWT
	}
	return payload
}

func (d *KubernetesAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
