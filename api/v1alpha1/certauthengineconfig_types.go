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

	"github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CertAuthEngineConfigSpec defines the desired state of CertAuthEngineConfig
type CertAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/{metadata.name}/config.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	CertAuthEngineConfigInternal `json:",inline"`
}

// CertAuthEngineConfigStatus defines the observed state of CertAuthEngineConfig
type CertAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CertAuthEngineConfig is the Schema for the certauthengineconfigs API
type CertAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status CertAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CertAuthEngineConfigList contains a list of CertAuthEngineConfig
type CertAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertAuthEngineConfig `json:"items"`
}

type CertAuthEngineConfigInternal struct {
	// If set, during renewal, skips the matching of presented client identity with the client identity used during login.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	DisableBinding bool `json:"disableBinding,omitempty"`

	// If set, metadata of the certificate including the metadata corresponding to allowedMetadataExtensions will be stored in the alias.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	EnableIdentityAliasMetadata bool `json:"enableIdentityAliasMetadata,omitempty"`

	// The size of the OCSP response LRU cache. Note that this cache is used for all configured certificates.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=100
	OCSPCacheSize int `json:"ocspCacheSize,omitempty"`

	// The size of the role cache. Use -1 to disable role caching.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=200
	RoleCacheSize int `json:"roleCacheSize,omitempty"`
}

func (c *CertAuthEngineConfigInternal) toMap() map[string]any {
	payload := make(map[string]any)
	payload["disable_binding"] = c.DisableBinding
	payload["enable_identity_alias_metadata"] = c.EnableIdentityAliasMetadata
	payload["ocsp_cache_size"] = c.OCSPCacheSize
	payload["role_cache_size"] = c.RoleCacheSize

	return payload
}

var _ vaultutils.VaultObject = &CertAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &CertAuthEngineConfig{}

func (r *CertAuthEngineConfig) GetPath() string {
	if r.Spec.Name != "" {
		return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/" + r.Spec.Name + "/config")
	}

	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/" + r.Name + "/config")
}

func (r *CertAuthEngineConfig) GetPayload() map[string]interface{} {
	return r.Spec.CertAuthEngineConfigInternal.toMap()
}

// IsEquivalentToDesiredState returns wether the passed payload is equivalent to the payload that the current object would generate. When this is a engine object the tune payload will be compared
func (r *CertAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.CertAuthEngineConfigInternal.toMap()

	return reflect.DeepEqual(desiredState, payload)
}

func (r *CertAuthEngineConfig) IsInitialized() bool {
	return true
}

func (r *CertAuthEngineConfig) IsValid() (bool, error) {
	return true, nil
}

func (r *CertAuthEngineConfig) IsDeletable() bool {
	return true
}

func (r *CertAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *CertAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *CertAuthEngineConfig) GetKubeAuthConfiguration() *utils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *CertAuthEngineConfig) GetVaultConnection() *utils.VaultConnection {
	return r.Spec.Connection
}

func (r *CertAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *CertAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&CertAuthEngineConfig{}, &CertAuthEngineConfigList{})
}
