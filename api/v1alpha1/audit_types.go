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

// AuditSpec defines the desired state of Audit
type AuditSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path is the path where the audit device will be mounted (e.g., "file", "file2", "syslog")
	// +kubebuilder:validation:Required
	Path string `json:"path"`

	// Type specifies the type of audit device (e.g., "file", "socket", "syslog")
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// Description is a human-friendly description of the audit device
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`

	// Options contains the configuration options for the audit device
	// For file type, this would include "file_path"
	// +kubebuilder:validation:Required
	Options map[string]string `json:"options"`

	// Local specifies if the audit device is a local mount only. Local mounts are not replicated or removed upon replication
	// +kubebuilder:validation:Optional
	Local bool `json:"local,omitempty"`
}

// AuditStatus defines the observed state of Audit
type AuditStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Audit is the Schema for the audits API
type Audit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuditSpec   `json:"spec,omitempty"`
	Status AuditStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuditList contains a list of Audit
type AuditList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Audit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Audit{}, &AuditList{})
}

var _ vaultutils.VaultObject = &Audit{}
var _ vaultutils.VaultAuditObject = &Audit{}
var _ vaultutils.ConditionsAware = &Audit{}

func (d *Audit) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *Audit) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *Audit) GetPath() string {
	return vaultutils.CleansePath("sys/audit/" + d.Spec.Path)
}

func (d *Audit) GetPayload() map[string]interface{} {
	payload := map[string]interface{}{
		"type":        d.Spec.Type,
		"description": d.Spec.Description,
		"local":       d.Spec.Local,
		"options":     d.Spec.Options,
	}
	return payload
}

func (d *Audit) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredPayload := d.GetPayload()
	if payload["type"] != desiredPayload["type"] {
		return false
	}
	if payload["description"] != desiredPayload["description"] {
		return false
	}
	if payload["local"] != desiredPayload["local"] {
		return false
	}
	return true
}

func (d *Audit) IsInitialized() bool {
	return true
}

func (d *Audit) IsValid() (bool, error) {
	return true, nil
}

func (d *Audit) IsDeletable() bool {
	return true
}

func (d *Audit) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *Audit) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (i *Audit) GetConditions() []metav1.Condition {
	return i.Status.Conditions
}

func (i *Audit) SetConditions(conditions []metav1.Condition) {
	i.Status.Conditions = conditions
}
