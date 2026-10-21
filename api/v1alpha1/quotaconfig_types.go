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

var _ vaultutils.VaultObject = &QuotaConfig{}
var _ vaultutils.ConditionsAware = &QuotaConfig{}

func (d *QuotaConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *QuotaConfig) GetPath() string {
	return "sys/quotas/config"
}

func (d *QuotaConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *QuotaConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.QuotaConfigData.toMap()
	sortAnyStringSlice(payload, "rate_limit_exempt_paths")
	sortAnyStringSlice(payload, "absolute_rate_limit_exempt_paths")
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *QuotaConfig) IsInitialized() bool {
	return true
}

func (d *QuotaConfig) IsDeletable() bool {
	return false
}

func (d *QuotaConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *QuotaConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *QuotaConfig) IsValid() (bool, error) {
	return true, nil
}

// QuotaConfigSpec defines the desired state of QuotaConfig
type QuotaConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	QuotaConfigData `json:",inline"`
}

// QuotaConfigData contains the quota configuration fields written to sys/quotas/config.
type QuotaConfigData struct {
	// EnableRateLimitAuditLogging toggles whether audit log entries also include rate-limit quota rejections.
	// +kubebuilder:validation:Optional
	EnableRateLimitAuditLogging bool `json:"enableRateLimitAuditLogging,omitempty"`

	// EnableRateLimitResponseHeaders controls whether Vault returns rate-limit quota information in HTTP response headers.
	// +kubebuilder:validation:Optional
	EnableRateLimitResponseHeaders bool `json:"enableRateLimitResponseHeaders,omitempty"`

	// RateLimitExemptPaths is a list of path suffixes exempted from rate limiting. Vault may add its own built-in paths.
	// +kubebuilder:validation:Optional
	RateLimitExemptPaths []string `json:"rateLimitExemptPaths,omitempty"`

	// AbsoluteRateLimitExemptPaths is a list of exact paths exempted from rate limiting. Overrides any built-in behaviour.
	// +kubebuilder:validation:Optional
	AbsoluteRateLimitExemptPaths []string `json:"absoluteRateLimitExemptPaths,omitempty"`
}

func (i *QuotaConfigData) toMap() map[string]any {
	payload := map[string]any{}
	payload["enable_rate_limit_audit_logging"] = i.EnableRateLimitAuditLogging
	payload["enable_rate_limit_response_headers"] = i.EnableRateLimitResponseHeaders
	payload["rate_limit_exempt_paths"] = toInterfaceArray(i.RateLimitExemptPaths)
	payload["absolute_rate_limit_exempt_paths"] = toInterfaceArray(i.AbsoluteRateLimitExemptPaths)
	sortAnyStringSlice(payload, "rate_limit_exempt_paths")
	sortAnyStringSlice(payload, "absolute_rate_limit_exempt_paths")
	return payload
}

// QuotaConfigStatus defines the observed state of QuotaConfig
type QuotaConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *QuotaConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *QuotaConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// QuotaConfig is the Schema for the quotaconfigs API
type QuotaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QuotaConfigSpec   `json:"spec,omitempty"`
	Status QuotaConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// QuotaConfigList contains a list of QuotaConfig
type QuotaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuotaConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QuotaConfig{}, &QuotaConfigList{})
}

func (d *QuotaConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
