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
	"reflect"
	"strconv"

	vault "github.com/hashicorp/vault/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecretEngineMountSpec defines the desired state of SecretEngineMount
type SecretEngineMountSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Authentication is the kube aoth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

	Mount `json:",inline"`

	// Path at which this secret engine will be available
	// The final path will be {[spec.authentication.namespace]}/{spec.path}/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path /sys/mounts/{[spec.authentication.namespace]}/{spec.path}/{metadata.name}.
	// +kubebuilder:validation:Required
	Path Path `json:"path,omitempty"`
}

func (d *SecretEngineMount) GetPath() string {
	return string(d.Spec.Path) + "/" + d.Name
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
	// Options undocumented
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Options map[string]string `json:"options,omitempty"`

	// DefaultLeaseTTL  The default lease duration, specified as a string duration like "5s" or "30m".
	// +kubebuilder:validation:Optional
	DefaultLeaseTTL string `json:"defaultLeaseTTL"`

	// Description another description...
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

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
	// +kubebuilder:validation:UniqueItems=true
	AuditNonHMACRequestKeys []string `json:"auditNonHMACRequestKeys,omitempty"`

	// AuditNonHMACResponseKeys list of keys that will not be HMAC'd by audit devices in the response data object.
	// +kubebuilder:validation:Optional
	// +listType=set
	// +kubebuilder:validation:UniqueItems=true
	AuditNonHMACResponseKeys []string `json:"auditNonHMACResponseKeys,omitempty"`

	// ListingVisibility Specifies whether to show this mount in the UI-specific listing endpoint. Valid values are "unauth" or "hidden". If not set, behaves like "hidden"
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"unauth","hidden"}
	// +kubebuilder:default:="hidden"
	ListingVisibility string `json:"listingVisibility,omitempty"`

	// PassthroughRequestHeaders list of headers to whitelist and pass from the request to the plugin.
	// +kubebuilder:validation:Optional
	// +listType=set
	// +kubebuilder:validation:UniqueItems=true
	PassthroughRequestHeaders []string `json:"passthroughRequestHeaders,omitempty"`

	// AllowedResponseHeaders list of headers to whitelist, allowing a plugin to include them in the response.
	// +kubebuilder:validation:Optional
	// +listType=set
	// +kubebuilder:validation:UniqueItems=true
	AllowedResponseHeaders []string `json:"allowedResponseHeaders,omitempty"`

	// TokenType undocumented
	// +kubebuilder:validation:Optional
	TokenType string `json:"tokenType,omitempty"`
}

// SecretEngineMountStatus defines the observed state of SecretEngineMount
type SecretEngineMountStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
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

func fromMountConfigOutput(mountConfigOutput *vault.MountConfigOutput) *MountConfig {
	return &MountConfig{
		DefaultLeaseTTL:           strconv.Itoa(mountConfigOutput.DefaultLeaseTTL),
		MaxLeaseTTL:               strconv.Itoa(mountConfigOutput.MaxLeaseTTL),
		ForceNoCache:              mountConfigOutput.ForceNoCache,
		AuditNonHMACRequestKeys:   mountConfigOutput.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:  mountConfigOutput.AuditNonHMACResponseKeys,
		ListingVisibility:         mountConfigOutput.ListingVisibility,
		PassthroughRequestHeaders: mountConfigOutput.PassthroughRequestHeaders,
		AllowedResponseHeaders:    mountConfigOutput.AllowedResponseHeaders,
		TokenType:                 mountConfigOutput.TokenType,
	}
}

func (mountConfig *MountConfig) getMountConfigInputFromMountConfig() *vault.MountConfigInput {
	return &vault.MountConfigInput{
		Options:                   mountConfig.Options,
		DefaultLeaseTTL:           mountConfig.DefaultLeaseTTL,
		Description:               mountConfig.Description,
		MaxLeaseTTL:               mountConfig.MaxLeaseTTL,
		ForceNoCache:              mountConfig.ForceNoCache,
		AuditNonHMACRequestKeys:   mountConfig.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:  mountConfig.AuditNonHMACResponseKeys,
		ListingVisibility:         mountConfig.ListingVisibility,
		PassthroughRequestHeaders: mountConfig.PassthroughRequestHeaders,
		AllowedResponseHeaders:    mountConfig.AllowedResponseHeaders,
		TokenType:                 mountConfig.TokenType,
	}
}

func (mount *Mount) GetMountInputFromMount() *vault.MountInput {
	return &vault.MountInput{
		Type:                  mount.Type,
		Description:           mount.Description,
		Config:                *mount.Config.getMountConfigInputFromMountConfig(),
		Local:                 mount.Local,
		SealWrap:              mount.SealWrap,
		ExternalEntropyAccess: mount.ExternalEntropyAccess,
		Options:               mount.Options,
	}
}

func (mountConfig *MountConfig) IsEquivalentTo(secretEngineMount *vault.MountConfigOutput) bool {
	currentMountConfig := fromMountConfigOutput(secretEngineMount)
	return reflect.DeepEqual(currentMountConfig, mountConfig)
}
