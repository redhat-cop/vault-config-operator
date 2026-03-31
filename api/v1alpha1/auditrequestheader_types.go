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

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AuditRequestHeaderSpec defines the desired state of AuditRequestHeader
type AuditRequestHeaderSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Name is the name of the request header to configure
	// The final path in Vault will be sys/config/auditing/request-headers/{metadata.name}
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// HMAC specifies if this header's value should be HMAC'd in the audit logs
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	HMAC bool `json:"hmac,omitempty"`
}

// AuditRequestHeaderStatus defines the observed state of AuditRequestHeader
type AuditRequestHeaderStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AuditRequestHeader is the Schema for the auditrequestheaders API
type AuditRequestHeader struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuditRequestHeaderSpec   `json:"spec,omitempty"`
	Status AuditRequestHeaderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuditRequestHeaderList contains a list of AuditRequestHeader
type AuditRequestHeaderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuditRequestHeader `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuditRequestHeader{}, &AuditRequestHeaderList{})
}

var _ vaultutils.VaultObject = &AuditRequestHeader{}
var _ vaultutils.ConditionsAware = &AuditRequestHeader{}

func (d *AuditRequestHeader) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AuditRequestHeader) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *AuditRequestHeader) GetPath() string {
	return vaultutils.CleansePath("sys/config/auditing/request-headers/" + d.Spec.Name)
}

func (d *AuditRequestHeader) GetPayload() map[string]interface{} {
	return map[string]interface{}{
		"hmac": d.Spec.HMAC,
	}
}

func (d *AuditRequestHeader) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	if hmac, ok := payload["hmac"].(bool); ok {
		return hmac == d.Spec.HMAC
	}
	return false
}

func (d *AuditRequestHeader) IsInitialized() bool {
	return true
}

func (d *AuditRequestHeader) IsValid() (bool, error) {
	return true, nil
}

func (d *AuditRequestHeader) IsDeletable() bool {
	return true
}

func (d *AuditRequestHeader) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *AuditRequestHeader) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (i *AuditRequestHeader) GetConditions() []metav1.Condition {
	return i.Status.Conditions
}

func (i *AuditRequestHeader) SetConditions(conditions []metav1.Condition) {
	i.Status.Conditions = conditions
}
