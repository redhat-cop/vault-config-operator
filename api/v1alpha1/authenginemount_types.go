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

// AuthEngineMountSpec defines the desired state of AuthEngineMount
type AuthEngineMountSpec struct {

	// Authentication is the kube aoth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

	AuthMount `json:",inline"`

	// Path at which this auth engine will be mounted
	// The final path will be {[spec.authentication.namespace]}/auth/{spec.path}/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path /sys/auth/{[spec.authentication.namespace]}/{spec.path}/{metadata.name}.
	// +kubebuilder:validation:Required
	Path Path `json:"path,omitempty"`
}

type AuthMount struct {

	// Description Specifies a human-friendly description of the auth method.
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`

	// Type Specifies the name of the authentication method type, such as "github" or "token".
	// +kubebuilder:validation:Required
	Type string `json:"type,omitempty"`

	// Config Specifies configuration options for this auth method.
	// +kubebuilder:validation:Optional
	Config AuthMountConfig `json:"config,omitempty"`

	// Local Specifies if the auth method is local only. Local auth methods are not replicated nor (if a secondary) removed by replication. Logins via local auth methods do not make use of identity, i.e. no entity or groups will be attached to the token.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Local bool `json:"local,omitempty"`

	// SealWrap Enable seal wrapping for the mount, causing values stored by the mount to be wrapped by the seal's encryption capability.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	SealWrap bool `json:"sealwrap,omitempty"`
}

type AuthMountConfig struct {
	// DefaultLeaseTTL  The default lease duration, specified as a string duration like "5s" or "30m".
	// +kubebuilder:validation:Optional
	DefaultLeaseTTL string `json:"defaultLeaseTTL"`

	// MaxLeaseTTL The maximum lease duration, specified as a string duration like "5s" or "30m".
	// +kubebuilder:validation:Optional
	MaxLeaseTTL string `json:"maxLeaseTTL"`

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

func (mount *AuthMount) GetMountInputFromMount() *vault.MountInput {
	return &vault.MountInput{
		Type:        mount.Type,
		Description: mount.Description,
		Config:      *mount.Config.getMountConfigInputFromMountConfig(),
		Local:       mount.Local,
		SealWrap:    mount.SealWrap,
	}
}

func (mountConfig *AuthMountConfig) getMountConfigInputFromMountConfig() *vault.MountConfigInput {
	return &vault.MountConfigInput{
		DefaultLeaseTTL:           mountConfig.DefaultLeaseTTL,
		MaxLeaseTTL:               mountConfig.MaxLeaseTTL,
		AuditNonHMACRequestKeys:   mountConfig.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:  mountConfig.AuditNonHMACResponseKeys,
		ListingVisibility:         mountConfig.ListingVisibility,
		PassthroughRequestHeaders: mountConfig.PassthroughRequestHeaders,
		AllowedResponseHeaders:    mountConfig.AllowedResponseHeaders,
	}
}

func (d *AuthEngineMount) GetPath() string {
	return cleansePath(string("auth/"+d.Spec.Path) + "/" + d.Name)
}

func (mountConfig *AuthMountConfig) IsEquivalentTo(secretEngineMount *vault.MountConfigOutput) bool {
	currentMountConfig := authMountConfigFromMountConfigOutput(secretEngineMount)
	return reflect.DeepEqual(currentMountConfig, mountConfig)
}

func authMountConfigFromMountConfigOutput(mountConfigOutput *vault.MountConfigOutput) *AuthMountConfig {
	return &AuthMountConfig{
		DefaultLeaseTTL:           strconv.Itoa(mountConfigOutput.DefaultLeaseTTL),
		MaxLeaseTTL:               strconv.Itoa(mountConfigOutput.MaxLeaseTTL),
		AuditNonHMACRequestKeys:   mountConfigOutput.AuditNonHMACRequestKeys,
		AuditNonHMACResponseKeys:  mountConfigOutput.AuditNonHMACResponseKeys,
		ListingVisibility:         mountConfigOutput.ListingVisibility,
		PassthroughRequestHeaders: mountConfigOutput.PassthroughRequestHeaders,
		AllowedResponseHeaders:    mountConfigOutput.AllowedResponseHeaders,
	}
}

// AuthEngineMountStatus defines the observed state of AuthEngineMount
type AuthEngineMountStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *AuthEngineMount) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *AuthEngineMount) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AuthEngineMount is the Schema for the authenginemounts API
type AuthEngineMount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuthEngineMountSpec   `json:"spec,omitempty"`
	Status AuthEngineMountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AuthEngineMountList contains a list of AuthEngineMount
type AuthEngineMountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuthEngineMount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AuthEngineMount{}, &AuthEngineMountList{})
}
