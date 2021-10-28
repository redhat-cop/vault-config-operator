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
	"net/url"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VaultSecretSpec defines the desired state of VaultSecret
type VaultSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Url of the Vault instance.
	// +kubebuilder:validation:Required
	Url string `json:"url,omitempty"`
	// RefreshPeriod if specified, the operator will refresh the secret with the given frequency.
	// Defaults to five minutes, and must be at least one minute.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="5m"
	RefreshPeriod *metav1.Duration `json:"refreshPeriod,omitempty"`
	// KVSecrets are the Key/Value secrets in Vault.
	// +kubebuilder:validation:Required
	KVSecrets []KVSecret `json:"kvSecrets,omitempty"`
	// TemplatizedK8sSecret is the formatted K8s Secret created by templating from the Vault KV secrets.
	// +kubebuilder:validation:Required
	TemplatizedK8sSecret TemplatizedK8sSecret `json:"output,omitempty"`
}

// VaultSecretStatus defines the observed state of VaultSecret
type VaultSecretStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	//LastVaultSecretUpdate last time when this secret was updated from Vault
	LastVaultSecretUpdate *metav1.Time `json:"lastVaultSecretUpdate,omitempty"`
}

func (vs *VaultSecret) GetConditions() []metav1.Condition {
	return vs.Status.Conditions
}

func (vs *VaultSecret) SetConditions(conditions []metav1.Condition) {
	vs.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VaultSecret is the Schema for the vaultsecrets API
type VaultSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultSecretSpec   `json:"spec,omitempty"`
	Status VaultSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VaultSecretList contains a list of VaultSecret
type VaultSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultSecret{}, &VaultSecretList{})
}

type KVSecret struct {
	// Name is an arbitrary, but unique, name for this KV Vault secret and referenced when templating.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// Authentication is the kube aoth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`
	// Keys is a list of keys to use for templating. If none are listed all keys are referenceable for templating.
	// +kubebuilder:validation:Optional
	Keys []string `json:"keys,omitempty"`
}

type TemplatizedK8sSecret struct {
	// Name is the K8s Secret name to output to.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// Type is the K8s Secret type to output to.
	// +kubebuilder:validation:Required
	Type string `json:"type,omitempty"`
	// StringData is the K8s Secret stringData and allows specifying non-binary secret data in string form with go templating support
	// to transform the Vault KV secrets into a formatted K8s Secret.
	// The Sprig template library and Helm functions (like toYaml) are supported.
	// +kubebuilder:validation:Required
	StringData map[string]string `json:"stringData,omitempty"`
	// Annotations are annotations to add to the final K8s Secret.
	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (vs *VaultSecret) IsValid() (bool, error) {
	err := vs.isValid()
	return err == nil, err
}

func (vs *VaultSecret) isValid() error {
	result := &multierror.Error{}
	result = multierror.Append(result, vs.validUrl())
	result = multierror.Append(result, vs.validResyncInterval())
	return result.ErrorOrNil()
}

func (vs *VaultSecret) validUrl() error {

	u, err := url.Parse(vs.Spec.Url)

	errCount := 0

	errs := errors.New("invalid url")
	if err != nil {
		errs = errors.Wrap(errs, err.Error())
		errCount++
	}

	if u.Scheme == "" {
		errs = errors.Wrap(errs, "no valid scheme in url")
		errCount++
	}

	if u.Host == "" {
		errs = errors.Wrap(errs, "no valid host in url")
		errCount++
	}

	if errCount > 0 {
		return errs
	}

	return nil
}

func (vs *VaultSecret) validResyncInterval() error {

	if vs.Spec.RefreshPeriod.Minutes() < 1 {
		return errors.New("ResyncInterval must be at least 1 minute")
	}

	return nil
}
