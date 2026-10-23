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
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ vaultutils.VaultObject = &LoggerConfig{}
var _ vaultutils.ConditionsAware = &LoggerConfig{}

func (d *LoggerConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *LoggerConfig) GetPath() string {
	return vaultutils.CleansePath("sys/loggers/" + vaultutils.ResourceName(d.Spec.Name, d.Name))
}

func (d *LoggerConfig) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *LoggerConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	loggerName := d.Spec.Name
	if loggerName == "" {
		loggerName = d.Name
	}

	// Vault read format: {loggerName: level}
	if vaultLevel, ok := payload[loggerName]; ok {
		return d.Spec.Level == vaultLevel
	}

	// Fallback: check write format (defensive)
	if level, ok := payload["level"]; ok {
		return d.Spec.Level == level
	}

	return false
}

func (d *LoggerConfig) IsDeletable() bool {
	return true
}

// LoggerConfigSpec defines the desired state of LoggerConfig
type LoggerConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	LoggerConfigConfig `json:",inline"`

	// The name of the Vault logger to configure (e.g., "audit", "core", "expiration").
	// If specified, takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// LoggerConfigConfig contains the Vault logger configuration fields.
type LoggerConfigConfig struct {
	// Level is the log verbosity level. Supported values in order of detail:
	// trace, debug, info, warn, error.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum={"trace","debug","info","warn","error"}
	Level string `json:"level" vault:"level"`
}

func (c *LoggerConfigConfig) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(c)
}

// LoggerConfigStatus defines the observed state of LoggerConfig
type LoggerConfigStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *LoggerConfig) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *LoggerConfig) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LoggerConfig is the Schema for the loggerconfigs API
type LoggerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoggerConfigSpec   `json:"spec,omitempty"`
	Status LoggerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LoggerConfigList contains a list of LoggerConfig
type LoggerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoggerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoggerConfig{}, &LoggerConfigList{})
}

func (d *LoggerConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
