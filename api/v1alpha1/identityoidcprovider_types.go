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

// IdentityOIDCProviderSpec defines the desired state of IdentityOIDCProvider
type IdentityOIDCProviderSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	IdentityOIDCProviderConfig `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type IdentityOIDCProviderConfig struct {

	// Issuer specifies what will be used as the scheme://host:port component for the iss claim of ID tokens.
	// This defaults to a URL with Vault's api_addr as the scheme://host:port component and
	// /v1/:namespace/identity/oidc/provider/:name as the path component.
	// If provided explicitly, it must point to a Vault instance that is network reachable by clients for ID token validation.
	// +kubebuilder:validation:Optional
	Issuer string `json:"issuer,omitempty"`

	// AllowedClientIDs is the list of client IDs that are permitted to use the provider.
	// If empty, no clients are allowed. If "*" is provided, all clients are allowed.
	// +kubebuilder:validation:Optional
	// +listType=set
	AllowedClientIDs []string `json:"allowedClientIDs,omitempty"`

	// ScopesSupported is the list of scopes available for requesting on the provider.
	// +kubebuilder:validation:Optional
	// +listType=set
	ScopesSupported []string `json:"scopesSupported,omitempty"`
}

// IdentityOIDCProviderStatus defines the observed state of IdentityOIDCProvider
type IdentityOIDCProviderStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IdentityOIDCProvider is the Schema for the identityoidcproviders API
type IdentityOIDCProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityOIDCProviderSpec   `json:"spec,omitempty"`
	Status IdentityOIDCProviderStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IdentityOIDCProviderList contains a list of IdentityOIDCProvider
type IdentityOIDCProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityOIDCProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IdentityOIDCProvider{}, &IdentityOIDCProviderList{})
}

var _ vaultutils.VaultObject = &IdentityOIDCProvider{}
var _ vaultutils.ConditionsAware = &IdentityOIDCProvider{}

func (m *IdentityOIDCProvider) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *IdentityOIDCProvider) IsDeletable() bool {
	return true
}

func (m *IdentityOIDCProvider) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *IdentityOIDCProvider) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *IdentityOIDCProvider) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("identity/oidc/provider/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("identity/oidc/provider/" + d.Name)
}

func (d *IdentityOIDCProvider) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *IdentityOIDCProviderSpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	if i.Issuer != "" {
		payload["issuer"] = i.Issuer
	}
	payload["allowed_client_ids"] = i.AllowedClientIDs
	payload["scopes_supported"] = i.ScopesSupported
	return payload
}

func (d *IdentityOIDCProvider) IsInitialized() bool {
	return true
}

func (d *IdentityOIDCProvider) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *IdentityOIDCProvider) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *IdentityOIDCProvider) IsValid() (bool, error) {
	return true, nil
}

func (d *IdentityOIDCProvider) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *IdentityOIDCProvider) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	return reflect.DeepEqual(desiredState, payload)
}
