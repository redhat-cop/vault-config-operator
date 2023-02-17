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
	"reflect"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var _ vaultutils.VaultObject = &PasswordPolicy{}

func (d *PasswordPolicy) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *PasswordPolicy) GetPath() string {
	return "sys/policies/password/" + d.Name
}
func (d *PasswordPolicy) GetPayload() map[string]interface{} {
	return map[string]interface{}{
		"policy": d.Spec.PasswordPolicy,
	}
}
func (d *PasswordPolicy) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	return reflect.DeepEqual(d.GetPayload(), payload)
}

func (d *PasswordPolicy) IsInitialized() bool {
	return true
}

func (d *PasswordPolicy) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *PasswordPolicy) IsValid() (bool, error) {
	return true, nil
}

// PasswordPolicySpec defines the desired state of PasswordPolicy
type PasswordPolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// PasswordPolicy  is a Vault password policy (https://www.vaultproject.io/docs/concepts/password-policies) expressed in HCL language.
	// +kubebuilder:validation:Required
	PasswordPolicy string `json:"passwordPolicy,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`
}

// PolicyStatus defines the observed state of Policy
type PasswordPolicyStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *PasswordPolicy) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *PasswordPolicy) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PasswordPolicy is the Schema for the passowordpolicies API
type PasswordPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PasswordPolicySpec   `json:"spec,omitempty"`
	Status PasswordPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PasswordPolicyList contains a list of PasswordPolicy
type PasswordPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PasswordPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PasswordPolicy{}, &PasswordPolicyList{})
}

func (d *PasswordPolicy) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
