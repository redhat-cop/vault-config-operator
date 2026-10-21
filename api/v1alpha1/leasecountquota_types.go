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
	"fmt"
	"reflect"
	"strconv"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ vaultutils.VaultObject = &LeaseCountQuota{}
var _ vaultutils.ConditionsAware = &LeaseCountQuota{}

func (d *LeaseCountQuota) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *LeaseCountQuota) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("sys/quotas/lease-count/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("sys/quotas/lease-count/" + d.Name)
}

func (d *LeaseCountQuota) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *LeaseCountQuota) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.LeaseCountQuotaConfig.toMap()
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *LeaseCountQuota) IsInitialized() bool {
	return true
}

func (d *LeaseCountQuota) IsDeletable() bool {
	return true
}

func (d *LeaseCountQuota) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *LeaseCountQuota) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *LeaseCountQuota) IsValid() (bool, error) {
	if r.Spec.MaxLeases < 1 {
		return false, fmt.Errorf("maxLeases must be at least 1, got %d", r.Spec.MaxLeases)
	}
	return true, nil
}

// LeaseCountQuotaSpec defines the desired state of LeaseCountQuota
type LeaseCountQuotaSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// The name of the quota in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`

	LeaseCountQuotaConfig `json:",inline"`
}

// LeaseCountQuotaConfig contains the lease count quota fields written to sys/quotas/lease-count/{name}.
// Note: Lease count quotas are an Enterprise-only feature.
type LeaseCountQuotaConfig struct {
	// MaxLeases is the maximum number of leases allowed by this quota rule. Must be >= 1.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MaxLeases int `json:"maxLeases"`

	// QuotaPath is an optional Vault API path prefix to scope this quota to. Named "quotaPath" to avoid collision with the VaultObject Path concept. Mapped to "path" in the Vault API payload.
	// +kubebuilder:validation:Optional
	QuotaPath string `json:"quotaPath,omitempty"`

	// Role is an optional auth role to scope this quota to. Only leases from this role will count.
	// +kubebuilder:validation:Optional
	Role string `json:"role,omitempty"`

	// Inheritable if true, sub-namespaces inherit this quota.
	// +kubebuilder:validation:Optional
	Inheritable bool `json:"inheritable,omitempty"`
}

func (i *LeaseCountQuotaConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["max_leases"] = json.Number(strconv.Itoa(i.MaxLeases))
	payload["path"] = i.QuotaPath
	payload["role"] = i.Role
	payload["inheritable"] = i.Inheritable
	return payload
}

// LeaseCountQuotaStatus defines the observed state of LeaseCountQuota
type LeaseCountQuotaStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *LeaseCountQuota) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *LeaseCountQuota) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LeaseCountQuota is the Schema for the leasecountquotas API
type LeaseCountQuota struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LeaseCountQuotaSpec   `json:"spec,omitempty"`
	Status LeaseCountQuotaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LeaseCountQuotaList contains a list of LeaseCountQuota
type LeaseCountQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LeaseCountQuota `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LeaseCountQuota{}, &LeaseCountQuotaList{})
}

func (d *LeaseCountQuota) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
