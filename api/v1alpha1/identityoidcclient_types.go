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

// IdentityOIDCClientSpec defines the desired state of IdentityOIDCClient
type IdentityOIDCClientSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	IdentityOIDCClientConfig `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type IdentityOIDCClientConfig struct {

	// Key is a reference to a named key resource. This key will be used to sign ID tokens for the client.
	// This cannot be modified after creation. If not supplied, defaults to the built-in default key.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="default"
	Key string `json:"key,omitempty"`

	// RedirectURIs is the list of redirection URI values used by the client.
	// One of these values must exactly match the redirect_uri parameter value used in each authentication request.
	// +kubebuilder:validation:Optional
	// +listType=set
	RedirectURIs []string `json:"redirectURIs,omitempty"`

	// Assignments is a list of assignment resources associated with the client.
	// Client assignments limit the Vault entities and groups that are allowed to authenticate through the client.
	// By default, no Vault entities are allowed. To allow all Vault entities to authenticate through the client,
	// supply the built-in allow_all assignment.
	// +kubebuilder:validation:Optional
	// +listType=set
	Assignments []string `json:"assignments,omitempty"`

	// ClientType is the client type based on its ability to maintain confidentiality of credentials.
	// This cannot be modified after creation. Must be "confidential" or "public".
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"confidential","public"}
	// +kubebuilder:default:="confidential"
	ClientType string `json:"clientType,omitempty"`

	// IDTokenTTL is the time-to-live for ID tokens obtained by the client.
	// Accepts duration format strings. The value should be less than the verification_ttl on the key.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="24h"
	IDTokenTTL string `json:"idTokenTTL,omitempty"`

	// AccessTokenTTL is the time-to-live for access tokens obtained by the client.
	// Accepts duration format strings.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="24h"
	AccessTokenTTL string `json:"accessTokenTTL,omitempty"`
}

// IdentityOIDCClientStatus defines the observed state of IdentityOIDCClient
type IdentityOIDCClientStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IdentityOIDCClient is the Schema for the identityoidcclients API
type IdentityOIDCClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityOIDCClientSpec   `json:"spec,omitempty"`
	Status IdentityOIDCClientStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IdentityOIDCClientList contains a list of IdentityOIDCClient
type IdentityOIDCClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityOIDCClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IdentityOIDCClient{}, &IdentityOIDCClientList{})
}

var _ vaultutils.VaultObject = &IdentityOIDCClient{}
var _ vaultutils.ConditionsAware = &IdentityOIDCClient{}

func (m *IdentityOIDCClient) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *IdentityOIDCClient) IsDeletable() bool {
	return true
}

func (m *IdentityOIDCClient) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *IdentityOIDCClient) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *IdentityOIDCClient) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("identity/oidc/client/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("identity/oidc/client/" + d.Name)
}

func (d *IdentityOIDCClient) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *IdentityOIDCClientSpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["key"] = i.Key
	payload["redirect_uris"] = i.RedirectURIs
	payload["assignments"] = i.Assignments
	payload["client_type"] = i.ClientType
	payload["id_token_ttl"] = i.IDTokenTTL
	payload["access_token_ttl"] = i.AccessTokenTTL
	return payload
}

func (d *IdentityOIDCClient) IsInitialized() bool {
	return true
}

func (d *IdentityOIDCClient) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *IdentityOIDCClient) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *IdentityOIDCClient) IsValid() (bool, error) {
	return true, nil
}

func (d *IdentityOIDCClient) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *IdentityOIDCClient) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	return reflect.DeepEqual(desiredState, payload)
}
