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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RandomSecretSpec defines the desired state of RandomSecret
type RandomSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Authentication is the kube aoth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the secret.
	// The final path will be {[spec.authentication.namespace]}/{spec.path}/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path Path `json:"path,omitempty"`

	// SecretFormat specifies a map of key and password policies used to generate random values
	// +kubebuilder:validation:Required
	// +mapType=granular
	SecretFormat map[string]PasswordPolicy

	// RefreshPeriod if specified, the operator will refresh the secret with the given frequency
	// +kubebuilder:validation:Optional
	RefreshPeriod metav1.Duration `json:"refreshPeriod,omitempty"`
}

type PasswordPolicy struct {
	// PasswordPolicyName a ref to a password policy defined in Vault. Notice that in order to use this, the Vault role you use needs the following capabilities = ["read"] on /sys/policy/password.
	// Only one of PasswordPolicyName or InlinePasswordPolicy can be specified
	// +kubebuilder:validation:Optional
	PasswordPolicyName string `json:"passwordPolicyName,omitempty"`

	// InlinePasswordPolicy is an inline password policy specified using Vault password policy syntax (https://www.vaultproject.io/docs/concepts/password-policies#password-policy-syntax)
	// Only one of PasswordPolicyName or InlinePasswordPolicy can be specified
	// +kubebuilder:validation:Optional
	InlinePasswordPolicy string `json:"inlinePasswordPolicy,omitempty"`
}

// RandomSecretStatus defines the observed state of RandomSecret
type RandomSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RandomSecret is the Schema for the randomsecrets API
type RandomSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RandomSecretSpec   `json:"spec,omitempty"`
	Status RandomSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RandomSecretList contains a list of RandomSecret
type RandomSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RandomSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RandomSecret{}, &RandomSecretList{})
}

type PasswordPolicyFormat struct {
	Length int                  `hcl:"length"`
	Rules  []PasswordPolicyRule `hcl:"rule,block"`
}

type PasswordPolicyRule struct {
	RuleType string `hcl:"type,label"`
	Charset  string `hcl:"charset"`
	MinChars string `hcl:"min-chars"`
}
