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
	"strings"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ vaultutils.VaultObject = &RateLimitQuota{}
var _ vaultutils.ConditionsAware = &RateLimitQuota{}

func (d *RateLimitQuota) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *RateLimitQuota) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("sys/quotas/rate-limit/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("sys/quotas/rate-limit/" + d.Name)
}

func (d *RateLimitQuota) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *RateLimitQuota) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.RateLimitQuotaConfig.toMap()
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *RateLimitQuota) IsInitialized() bool {
	return true
}

func (d *RateLimitQuota) IsDeletable() bool {
	return true
}

func (d *RateLimitQuota) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *RateLimitQuota) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *RateLimitQuota) IsValid() (bool, error) {
	rate, err := strconv.ParseFloat(r.Spec.Rate, 64)
	if err != nil {
		return false, fmt.Errorf("rate must be a valid number, got %q: %w", r.Spec.Rate, err)
	}
	if rate <= 0 {
		return false, fmt.Errorf("rate must be greater than 0, got %v", r.Spec.Rate)
	}
	return true, nil
}

// RateLimitQuotaSpec defines the desired state of RateLimitQuota
type RateLimitQuotaSpec struct {
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

	RateLimitQuotaConfig `json:",inline"`
}

// RateLimitQuotaConfig contains the rate limit quota fields written to sys/quotas/rate-limit/{name}.
type RateLimitQuotaConfig struct {
	// Rate is the maximum number of requests per interval allowed by this quota rule, expressed as a numeric string (e.g. "100", "897.3").
	// +kubebuilder:validation:Required
	Rate string `json:"rate"`

	// Interval is the duration over which the rate is evaluated (e.g. "1m", "30s"). Default is "1s" if omitted.
	// +kubebuilder:validation:Optional
	Interval string `json:"interval,omitempty"`

	// BlockInterval is how long a client is blocked when the rate is exceeded (e.g. "5m"). Default is "0s" (no block).
	// +kubebuilder:validation:Optional
	BlockInterval string `json:"blockInterval,omitempty"`

	// QuotaPath is an optional Vault API path prefix to scope this quota to. Named "quotaPath" to avoid collision with the VaultObject Path concept. Mapped to "path" in the Vault API payload.
	// +kubebuilder:validation:Optional
	QuotaPath string `json:"quotaPath,omitempty"`

	// Role is an optional auth role to scope this quota to. Only requests authenticated via this role will count.
	// +kubebuilder:validation:Optional
	Role string `json:"role,omitempty"`

	// Inheritable if true, sub-namespaces inherit this quota.
	// +kubebuilder:validation:Optional
	Inheritable bool `json:"inheritable,omitempty"`
}

func (i *RateLimitQuotaConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["rate"] = rateToJsonNumber(i.Rate)
	payload["interval"] = durationToSeconds(i.Interval)
	payload["block_interval"] = durationToSeconds(i.BlockInterval)
	payload["path"] = i.QuotaPath
	payload["role"] = i.Role
	payload["inheritable"] = i.Inheritable
	return payload
}

func rateToJsonNumber(rate string) json.Number {
	rate = strings.TrimSpace(rate)
	if rate == "" {
		return json.Number("0")
	}
	return json.Number(rate)
}

// RateLimitQuotaStatus defines the observed state of RateLimitQuota
type RateLimitQuotaStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *RateLimitQuota) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *RateLimitQuota) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RateLimitQuota is the Schema for the ratelimitquotas API
type RateLimitQuota struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RateLimitQuotaSpec   `json:"spec,omitempty"`
	Status RateLimitQuotaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RateLimitQuotaList contains a list of RateLimitQuota
type RateLimitQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimitQuota `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RateLimitQuota{}, &RateLimitQuotaList{})
}

func (d *RateLimitQuota) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
