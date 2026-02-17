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

// IdentityTokenKeySpec defines the desired state of IdentityTokenKey
type IdentityTokenKeySpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	IdentityTokenKeyConfig `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type IdentityTokenKeyConfig struct {

	// RotationPeriod controls how often to generate a new signing key.
	// Uses duration format strings.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="24h"
	RotationPeriod string `json:"rotationPeriod,omitempty"`

	// VerificationTTL controls how long the public portion of a signing key will be
	// available for verification after being rotated. Uses duration format strings.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="24h"
	VerificationTTL string `json:"verificationTTL,omitempty"`

	// AllowedClientIDs is a list of role client IDs allowed to use this key for signing.
	// If empty, no roles are allowed. If "*", all roles are allowed.
	// +kubebuilder:validation:Optional
	// +listType=set
	AllowedClientIDs []string `json:"allowedClientIDs,omitempty"`

	// Algorithm is the signing algorithm to use.
	// Allowed values are: RS256 (default), RS384, RS512, ES256, ES384, ES512, EdDSA.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"RS256","RS384","RS512","ES256","ES384","ES512","EdDSA"}
	// +kubebuilder:default:="RS256"
	Algorithm string `json:"algorithm,omitempty"`
}

// IdentityTokenKeyStatus defines the observed state of IdentityTokenKey
type IdentityTokenKeyStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IdentityTokenKey is the Schema for the identitytokenkeys API
type IdentityTokenKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IdentityTokenKeySpec   `json:"spec,omitempty"`
	Status IdentityTokenKeyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IdentityTokenKeyList contains a list of IdentityTokenKey
type IdentityTokenKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IdentityTokenKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IdentityTokenKey{}, &IdentityTokenKeyList{})
}

var _ vaultutils.VaultObject = &IdentityTokenKey{}
var _ vaultutils.ConditionsAware = &IdentityTokenKey{}

func (m *IdentityTokenKey) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *IdentityTokenKey) IsDeletable() bool {
	return true
}

func (m *IdentityTokenKey) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *IdentityTokenKey) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *IdentityTokenKey) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath("identity/oidc/key/" + d.Spec.Name)
	}
	return vaultutils.CleansePath("identity/oidc/key/" + d.Name)
}

func (d *IdentityTokenKey) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *IdentityTokenKeySpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["rotation_period"] = i.RotationPeriod
	payload["verification_ttl"] = i.VerificationTTL
	payload["allowed_client_ids"] = i.AllowedClientIDs
	payload["algorithm"] = i.Algorithm
	return payload
}

func (d *IdentityTokenKey) IsInitialized() bool {
	return true
}

func (d *IdentityTokenKey) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *IdentityTokenKey) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *IdentityTokenKey) IsValid() (bool, error) {
	return true, nil
}

func (d *IdentityTokenKey) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *IdentityTokenKey) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	return reflect.DeepEqual(desiredState, payload)
}
