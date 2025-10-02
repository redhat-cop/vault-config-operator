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

var _ vaultutils.VaultObject = &SecretEngineMount{}
var _ vaultutils.VaultEngineObject = &SecretEngineMount{}
var _ vaultutils.ConditionsAware = &SecretEngineMount{}

func (d *SecretEngineMount) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *SecretEngineMount) IsDeletable() bool {
	return true
}

func (d *SecretEngineMount) GetPath() string {
	var pathComponent string
	if d.Spec.Path != "" {
		pathComponent = string(d.Spec.Path)
	} else {
		// When Path is empty, use the name directly as the mount path
		if d.Spec.Name != "" {
			return vaultutils.CleansePath(d.GetEngineListPath() + "/" + d.Spec.Name)
		}
		return vaultutils.CleansePath(d.GetEngineListPath() + "/" + d.Name)
	}

	if d.Spec.Name != "" {
		return vaultutils.CleansePath(d.GetEngineListPath() + "/" + pathComponent + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(d.GetEngineListPath() + "/" + pathComponent + "/" + d.Name)
}
func (d *SecretEngineMount) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *SecretEngineMount) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	configMap := d.Spec.Config.toMap()
	delete(configMap, "options")
	delete(configMap, "description")
	return reflect.DeepEqual(configMap, payload)
}

func (d *SecretEngineMount) IsInitialized() bool {
	return true
}

func (d *SecretEngineMount) IsValid() (bool, error) {
	return true, nil
}

func (d *SecretEngineMount) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *SecretEngineMount) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *SecretEngineMount) GetEngineListPath() string {
	return "sys/mounts"
}
func (d *SecretEngineMount) GetEngineTunePath() string {
	return d.GetPath() + "/tune"
}
func (d *SecretEngineMount) GetTunePayload() map[string]interface{} {
	return d.Spec.Config.toMap()
}

func (d *SecretEngineMount) SetAccessor(accessor string) {
	d.Status.Accessor = accessor
}

// SecretEngineMountSpec defines the desired state of SecretEngineMount
type SecretEngineMountSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	Mount `json:",inline"`

	// Path at which this secret engine will be available. If not specified, defaults to the resource name (/sys/mounts/{[spec.authentication.namespace]}/{metadata.name}).
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.path]}/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on computed path /sys/mounts/{[spec.authentication.namespace]}/{[spec.path]}/{metadata.name} or /sys/mounts/{[spec.authentication.namespace]}/{metadata.name} if path is empty.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// +k8s:openapi-gen=true
type Mount struct {
	// Type Specifies the type of the backend, such as "aws".
	// +kubebuilder:validation:Required
	Type string `json:"type,omitempty"`

	// Description Specifies the human-friendly description of the mount.
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`

	// Specifies configuration options for this mount; if set on a specific mount, values will override any global defaults (e.g. the system TTL/Max TTL)
	// +kubebuilder:validation:Optional
	Config MountConfig `json:"config,omitempty"`

	// Local Specifies if the secrets engine is a local mount only. Local mounts are not replicated nor (if a secondary) removed by replication.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	Local bool `json:"local,omitempty"`

	// SealWrap Enable seal wrapping for the mount, causing values stored by the mount to be wrapped by the seal's encryption capability.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	SealWrap bool `json:"sealWrap,omitempty"`

	// ExternalEntropyAccess Enable the secrets engine to access Vault's external entropy source.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	ExternalEntropyAccess bool `json:"externalEntropyAccess,omitempty"`

	// Options Specifies mount type specific options that are passed to the backend.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Options map[string]string `json:"options,omitempty"`
}

// +k8s:openapi-gen=true
type MountConfig struct {

	// DefaultLeaseTTL  The default lease duration, specified as a string duration like "5s" or "30m".
	// +kubebuilder:validation:Optional
	DefaultLeaseTTL string `json:"defaultLeaseTTL"`

	// MaxLeaseTTL The maximum lease duration, specified as a string duration like "5s" or "30m".
	// +kubebuilder:validation:Optional
	MaxLeaseTTL string `json:"maxLeaseTTL"`

	// ForceNoCache Disable caching.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	ForceNoCache bool `json:"forceNoCache"`

	// AuditNonHMACRequestKeys list of keys that will not be HMAC'd by audit devices in the request data object.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AuditNonHMACRequestKeys []string `json:"auditNonHMACRequestKeys,omitempty"`

	// AuditNonHMACResponseKeys list of keys that will not be HMAC'd by audit devices in the response data object.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AuditNonHMACResponseKeys []string `json:"auditNonHMACResponseKeys,omitempty"`

	// ListingVisibility Specifies whether to show this mount in the UI-specific listing endpoint. Valid values are "unauth" or "hidden". If not set, behaves like "hidden"
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"unauth","hidden"}
	// +kubebuilder:default:="hidden"
	ListingVisibility string `json:"listingVisibility,omitempty"`

	// PassthroughRequestHeaders list of headers to whitelist and pass from the request to the plugin.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	PassthroughRequestHeaders []string `json:"passthroughRequestHeaders,omitempty"`

	// AllowedResponseHeaders list of headers to whitelist, allowing a plugin to include them in the response.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AllowedResponseHeaders []string `json:"allowedResponseHeaders,omitempty"`
}

// SecretEngineMountStatus defines the observed state of SecretEngineMount
type SecretEngineMountStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// +kubebuilder:validation:Optional
	Accessor string `json:"accessor,omitempty"`
}

func (m *SecretEngineMount) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *SecretEngineMount) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SecretEngineMount is the Schema for the secretenginemounts API
type SecretEngineMount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretEngineMountSpec   `json:"spec,omitempty"`
	Status SecretEngineMountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretEngineMountList contains a list of SecretEngineMount
type SecretEngineMountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretEngineMount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretEngineMount{}, &SecretEngineMountList{})
}

func (mc *MountConfig) toMap() map[string]interface{} {
	return map[string]interface{}{
		"default_lease_ttl":            mc.DefaultLeaseTTL,
		"max_lease_ttl":                mc.MaxLeaseTTL,
		"force_no_cache":               mc.ForceNoCache,
		"audit_non_hmac_request_keys":  mc.AuditNonHMACRequestKeys,
		"audit_non_hmac_response_keys": mc.AuditNonHMACResponseKeys,
		"listing_visibility":           mc.ListingVisibility,
		"passthrough_request_headers":  mc.PassthroughRequestHeaders,
		"allowed_response_headers":     mc.AllowedResponseHeaders,
	}
}

func (m *Mount) toMap() map[string]interface{} {
	return map[string]interface{}{
		"type":                    m.Type,
		"description":             m.Description,
		"config":                  m.Config.toMap(),
		"local":                   m.Local,
		"seal_wrap":               m.SealWrap,
		"external_entropy_access": m.ExternalEntropyAccess,
		"options":                 m.Options,
	}
}

func (d *SecretEngineMount) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
