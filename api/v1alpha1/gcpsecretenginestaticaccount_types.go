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

// GCPSecretEngineStaticAccountSpec defines the desired state of GCPSecretEngineStaticAccount
type GCPSecretEngineStaticAccountSpec struct {
	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/static-account/{metadata.name}.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	GCPSEStaticAccount `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type GCPSEStaticAccount struct {
	// SecretType specifies the type of secret generated. Accepted values: access_token, service_account_key.
	// Cannot be updated after creation.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"access_token","service_account_key"}
	// +kubebuilder:default="access_token"
	SecretType string `json:"secretType"`

	// ServiceAccountEmail is the email of the GCP service account to manage. Cannot be updated.
	// +kubebuilder:validation:Required
	ServiceAccountEmail string `json:"serviceAccountEmail"`

	// Bindings is the bindings configuration string. Both JSON and HCL formats are supported.
	// When comparing with Vault state, JSON is parsed first; if that fails, HCL resource blocks
	// are parsed (extracting resource names and roles). If neither format can be parsed,
	// bindings are excluded from drift detection as a graceful fallback.
	// +kubebuilder:validation:Optional
	Bindings string `json:"bindings,omitempty"`

	// TokenScopes is a list of OAuth scopes for access_token type static accounts only.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenScopes []string `json:"tokenScopes,omitempty"`
}

var _ vaultutils.VaultObject = &GCPSecretEngineStaticAccount{}
var _ vaultutils.ConditionsAware = &GCPSecretEngineStaticAccount{}

func (d *GCPSecretEngineStaticAccount) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *GCPSecretEngineStaticAccount) IsDeletable() bool {
	return true
}

func (d *GCPSecretEngineStaticAccount) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "static-account" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "static-account" + "/" + d.Name)
}

func (d *GCPSecretEngineStaticAccount) GetPayload() map[string]any {
	return d.Spec.GCPSEStaticAccount.toMap()
}

func (d *GCPSecretEngineStaticAccount) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.GCPSEStaticAccount.toMap()
	sortAnyStringSlice(desiredState, "token_scopes")
	removeUnsetFields(desiredState, payload)
	normalizeBindingsForComparison(desiredState, payload)
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	sortAnyStringSlice(filteredPayload, "token_scopes")
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (d *GCPSecretEngineStaticAccount) IsInitialized() bool {
	return true
}

func (d *GCPSecretEngineStaticAccount) IsValid() (bool, error) {
	return true, nil
}

func (d *GCPSecretEngineStaticAccount) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *GCPSecretEngineStaticAccount) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *GCPSecretEngineStaticAccount) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (i *GCPSEStaticAccount) toMap() map[string]any {
	payload := map[string]any{}
	payload["secret_type"] = i.SecretType
	payload["service_account_email"] = i.ServiceAccountEmail
	payload["bindings"] = i.Bindings
	payload["token_scopes"] = toInterfaceArray(i.TokenScopes)
	return payload
}

// GCPSecretEngineStaticAccountStatus defines the observed state of GCPSecretEngineStaticAccount
type GCPSecretEngineStaticAccountStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *GCPSecretEngineStaticAccount) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *GCPSecretEngineStaticAccount) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GCPSecretEngineStaticAccount is the Schema for the gcpsecretenginestaticaccounts API
type GCPSecretEngineStaticAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPSecretEngineStaticAccountSpec   `json:"spec,omitempty"`
	Status GCPSecretEngineStaticAccountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GCPSecretEngineStaticAccountList contains a list of GCPSecretEngineStaticAccount
type GCPSecretEngineStaticAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPSecretEngineStaticAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPSecretEngineStaticAccount{}, &GCPSecretEngineStaticAccountList{})
}
