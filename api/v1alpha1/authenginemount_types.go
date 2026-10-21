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

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AuthEngineMountSpec defines the desired state of AuthEngineMount
type AuthEngineMountSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	AuthMount `json:",inline"`

	// Path at which this auth engine will be mounted
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/auth/{spec.path}/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path /sys/auth/{[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/{spec.path}/{metadata.name}.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

func (d *AuthEngineMount) IsDeletable() bool {
	return true
}

func (d *AuthEngineMount) GetPath() string {
	return vaultutils.CleansePath(d.GetEngineListPath() + "/" + string(d.Spec.Path) + "/" + vaultutils.ResourceName(d.Spec.Name, d.Name))
}

type AuthMount struct {

	// Description Specifies a human-friendly description of the auth method.
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty" vault:"description"`

	// Type Specifies the name of the authentication method type, such as "github" or "token".
	// +kubebuilder:validation:Optional
	Type string `json:"type,omitempty" vault:"type"`

	// Config Specifies configuration options for this auth method.
	// +kubebuilder:validation:Optional
	Config AuthMountConfig `json:"config,omitempty" vault:"-"`

	// Local Specifies if the auth method is local only. Local auth methods are not replicated nor (if a secondary) removed by replication. Logins via local auth methods do not make use of identity, i.e. no entity or groups will be attached to the token.
	// +kubebuilder:validation:Optional
	Local bool `json:"local,omitempty" vault:"local"`

	// SealWrap Enable seal wrapping for the mount, causing values stored by the mount to be wrapped by the seal's encryption capability.
	// +kubebuilder:validation:Optional
	SealWrap bool `json:"sealwrap,omitempty" vault:"seal_wrap"`
}

type AuthMountConfig struct {
	// DefaultLeaseTTL  The default lease duration, specified as a string duration like "5s" or "30m".
	// +kubebuilder:validation:Optional
	DefaultLeaseTTL string `json:"defaultLeaseTTL" vault:"default_lease_ttl"`

	// MaxLeaseTTL The maximum lease duration, specified as a string duration like "5s" or "30m".
	// +kubebuilder:validation:Optional
	MaxLeaseTTL string `json:"maxLeaseTTL" vault:"max_lease_ttl"`

	// AuditNonHMACRequestKeys list of keys that will not be HMAC'd by audit devices in the request data object.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems:=true
	AuditNonHMACRequestKeys []string `json:"auditNonHMACRequestKeys,omitempty" vault:"audit_non_hmac_request_keys"`

	// AuditNonHMACResponseKeys list of keys that will not be HMAC'd by audit devices in the response data object.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AuditNonHMACResponseKeys []string `json:"auditNonHMACResponseKeys,omitempty" vault:"audit_non_hmac_response_keys"`

	// ListingVisibility Specifies whether to show this mount in the UI-specific listing endpoint. Valid values are "unauth" or "hidden". If not set, behaves like "hidden"
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"unauth","hidden"}
	// +kubebuilder:default:="hidden"
	ListingVisibility string `json:"listingVisibility,omitempty" vault:"listing_visibility"`

	// PassthroughRequestHeaders list of headers to whitelist and pass from the request to the plugin.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems:=true
	PassthroughRequestHeaders []string `json:"passthroughRequestHeaders,omitempty" vault:"passthrough_request_headers"`

	// AllowedResponseHeaders list of headers to whitelist, allowing a plugin to include them in the response.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AllowedResponseHeaders []string `json:"allowedResponseHeaders,omitempty" vault:"allowed_response_headers"`

	// Options undocumented
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Options map[string]string `json:"options,omitempty" vault:"options,stringmap"`

	// TokenType undocumented
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
	TokenType string `json:"tokenType,omitempty" vault:"token_type"`

	// Description another description...
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty" vault:"description"`
}

var _ vaultutils.VaultObject = &AuthEngineMount{}
var _ vaultutils.VaultEngineObject = &AuthEngineMount{}
var _ vaultutils.ConditionsAware = &AuthEngineMount{}

func (mc *AuthMountConfig) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(mc)
}

func (m *AuthMount) toMap() map[string]any {
	payload := vaultutils.StructToVaultPayload(m)
	payload["description"] = m.Description
	payload["config"] = m.Config.toMap()
	return payload
}

func (d *AuthEngineMount) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AuthEngineMount) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *AuthEngineMount) GetPayload() map[string]any {
	return d.Spec.toMap()
}
func (d *AuthEngineMount) IsEquivalentToDesiredState(payload map[string]any) bool {
	configMap := d.Spec.Config.toMap()
	return reflect.DeepEqual(configMap, filterPayloadToDesiredKeys(configMap, payload))
}

func (d *AuthEngineMount) GetEngineListPath() string {
	return "sys/auth"
}
func (d *AuthEngineMount) GetEngineTunePath() string {
	return d.GetPath() + "/tune"
}
func (d *AuthEngineMount) GetTunePayload() map[string]any {
	return d.Spec.Config.toMap()
}

func (d *AuthEngineMount) SetAccessor(accessor string) {
	d.Status.Accessor = accessor
}

// AuthEngineMountStatus defines the observed state of AuthEngineMount
type AuthEngineMountStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// +kubebuilder:validation:Optional
	Accessor string `json:"accessor,omitempty"`
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
