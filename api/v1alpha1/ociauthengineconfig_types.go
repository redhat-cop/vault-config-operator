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

// OCIAuthEngineConfigSpec defines the desired state of OCIAuthEngineConfig
type OCIAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the OCI auth engine is mounted.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/config.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	OCIAuthConfig `json:",inline"`
}

// OCIAuthEngineConfigStatus defines the observed state of OCIAuthEngineConfig
type OCIAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OCIAuthEngineConfig is the Schema for the ociauthengineconfigs API
type OCIAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OCIAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status OCIAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OCIAuthEngineConfigList contains a list of OCIAuthEngineConfig
type OCIAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OCIAuthEngineConfig `json:"items"`
}

type OCIAuthConfig struct {
	// HomeTenancyID is the Tenancy OCID of the OCI account.
	// +kubebuilder:validation:Required
	HomeTenancyID string `json:"homeTenancyID"`
}

var _ vaultutils.VaultObject = &OCIAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &OCIAuthEngineConfig{}

func (d *OCIAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *OCIAuthEngineConfig) IsDeletable() bool {
	return false
}

func (r *OCIAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *OCIAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *OCIAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config")
}

func (r *OCIAuthEngineConfig) GetPayload() map[string]any {
	return r.Spec.OCIAuthConfig.toMap()
}

func (r *OCIAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.OCIAuthConfig.toMap()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (r *OCIAuthEngineConfig) IsInitialized() bool {
	return true
}

func (r *OCIAuthEngineConfig) IsValid() (bool, error) {
	return true, nil
}

func (r *OCIAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *OCIAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *OCIAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func init() {
	SchemeBuilder.Register(&OCIAuthEngineConfig{}, &OCIAuthEngineConfigList{})
}

func (c *OCIAuthConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["home_tenancy_id"] = c.HomeTenancyID
	return payload
}
