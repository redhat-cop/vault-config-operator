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
	"reflect"
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ vaultutils.VaultObject = &RotationPolicy{}
var _ vaultutils.ConditionsAware = &RotationPolicy{}

func (d *RotationPolicy) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *RotationPolicy) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("sys/policies/rotation/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("sys/policies/rotation/" + d.Name)
}

func (d *RotationPolicy) GetPayload() map[string]any {
	return d.Spec.RotationPolicyConfig.toMap()
}

func (d *RotationPolicy) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.GetPayload()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *RotationPolicy) IsInitialized() bool {
	return true
}

func (d *RotationPolicy) IsDeletable() bool {
	return true
}

func (d *RotationPolicy) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *RotationPolicy) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *RotationPolicy) IsValid() (bool, error) {
	return true, nil
}

// RotationPolicySpec defines the desired state of RotationPolicy
type RotationPolicySpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	RotationPolicyConfig `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// RotationPolicyConfig contains the Vault rotation policy configuration fields.
type RotationPolicyConfig struct {

	// MaxRetriesPerCycle is the maximum number of times Vault should apply
	// backoff for rotation entries after a rotation failure.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	MaxRetriesPerCycle int `json:"maxRetriesPerCycle,omitempty"`

	// MaxRetryCycles is the maximum number of times Vault should requeue
	// rotation entries after reaching the maxRetriesPerCycle limit.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	MaxRetryCycles int `json:"maxRetryCycles,omitempty"`
}

func (d *RotationPolicyConfig) toMap() map[string]any {
	result := map[string]any{}
	if d.MaxRetriesPerCycle != 0 {
		result["max_retries_per_cycle"] = json.Number(strconv.Itoa(d.MaxRetriesPerCycle))
	}
	if d.MaxRetryCycles != 0 {
		result["max_retry_cycles"] = json.Number(strconv.Itoa(d.MaxRetryCycles))
	}
	return result
}

// RotationPolicyStatus defines the observed state of RotationPolicy
type RotationPolicyStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *RotationPolicy) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *RotationPolicy) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RotationPolicy is the Schema for the rotationpolicies API
type RotationPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RotationPolicySpec   `json:"spec,omitempty"`
	Status RotationPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RotationPolicyList contains a list of RotationPolicy
type RotationPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RotationPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RotationPolicy{}, &RotationPolicyList{})
}

func (d *RotationPolicy) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
