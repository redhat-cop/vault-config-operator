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
	"encoding/json"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GCPSecretEngineRolesetSpec defines the desired state of GCPSecretEngineRoleset
type GCPSecretEngineRolesetSpec struct {
	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/roleset/{metadata.name}.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	GCPSERoleset `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type GCPSERoleset struct {
	// SecretType specifies the type of secret generated. Accepted values: access_token, service_account_key.
	// Cannot be updated after creation.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"access_token","service_account_key"}
	// +kubebuilder:default="access_token"
	SecretType string `json:"secretType"`

	// Project is the GCP project ID that this roleset's service account belongs to. Cannot be updated.
	// +kubebuilder:validation:Required
	Project string `json:"project"`

	// Bindings is the bindings configuration string. Both JSON and HCL formats are supported.
	// When comparing with Vault state, JSON is parsed first; if that fails, HCL resource blocks
	// are parsed (extracting resource names and roles). If neither format can be parsed,
	// bindings are excluded from drift detection as a graceful fallback.
	// +kubebuilder:validation:Optional
	Bindings string `json:"bindings,omitempty"`

	// TokenScopes is a list of OAuth scopes for access_token type rolesets only.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenScopes []string `json:"tokenScopes,omitempty"`
}

var _ vaultutils.VaultObject = &GCPSecretEngineRoleset{}
var _ vaultutils.ConditionsAware = &GCPSecretEngineRoleset{}

func (d *GCPSecretEngineRoleset) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *GCPSecretEngineRoleset) IsDeletable() bool {
	return true
}

func (d *GCPSecretEngineRoleset) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roleset" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roleset" + "/" + d.Name)
}

func (d *GCPSecretEngineRoleset) GetPayload() map[string]any {
	return d.Spec.GCPSERoleset.toMap()
}

func (d *GCPSecretEngineRoleset) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.GCPSERoleset.toMap()
	delete(desiredState, "project")
	sortAnyStringSlice(desiredState, "token_scopes")
	removeUnsetFields(desiredState, payload)
	normalizeBindingsForComparison(desiredState, payload)
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	sortAnyStringSlice(filteredPayload, "token_scopes")
	return reflect.DeepEqual(desiredState, filteredPayload)
}

// normalizeBindingsForComparison handles the type mismatch where the operator
// sends bindings as an HCL/JSON string but Vault returns a parsed map. It first
// attempts JSON parsing; if that fails it tries HCL parsing (extracting resource
// blocks into a map keyed by resource name with role slices as values). If both
// fail, bindings is excluded from comparison as a graceful fallback.
//
// After parsing, nested role arrays within each binding entry are sorted so that
// equivalent bindings with different role order are detected as equal.
func normalizeBindingsForComparison(desiredState, payload map[string]any) {
	desiredBindings, isString := desiredState["bindings"].(string)
	if !isString {
		return
	}
	payloadBindings, payloadIsMap := payload["bindings"].(map[string]any)
	if !payloadIsMap {
		return
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(desiredBindings), &parsed); err == nil {
		sortBindingsRoles(parsed)
		sortBindingsRoles(payloadBindings)
		desiredState["bindings"] = parsed
		return
	}

	if hclParsed, err := parseHCLBindings(desiredBindings); err == nil && len(hclParsed) > 0 {
		sortBindingsRoles(hclParsed)
		sortBindingsRoles(payloadBindings)
		desiredState["bindings"] = hclParsed
		return
	}

	delete(desiredState, "bindings")
}

// sortBindingsRoles sorts all role string slices within a bindings map so that
// order differences don't cause false drift.
func sortBindingsRoles(bindings map[string]any) {
	for key := range bindings {
		sortAnyStringSlice(bindings, key)
	}
}

// parseHCLBindings parses GCP-style HCL bindings into a map suitable for
// comparison with Vault's API response. The expected HCL structure is:
//
//	resource "RESOURCE_NAME" { roles = ["ROLE1", "ROLE2"] }
//
// The returned map keys are the resource names and values are []any of role strings.
func parseHCLBindings(hclStr string) (map[string]any, error) {
	file, diags := hclsyntax.ParseConfig([]byte(hclStr), "bindings.hcl", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, diags
	}
	result := map[string]any{}
	for _, block := range body.Blocks {
		if block.Type != "resource" || len(block.Labels) != 1 {
			continue
		}
		attrs, attrDiags := block.Body.JustAttributes()
		if attrDiags.HasErrors() {
			continue
		}
		rolesAttr, exists := attrs["roles"]
		if !exists {
			continue
		}
		val, valDiags := rolesAttr.Expr.Value(nil)
		if valDiags.HasErrors() {
			continue
		}
		roles := []any{}
		for _, v := range val.AsValueSlice() {
			roles = append(roles, v.AsString())
		}
		result[block.Labels[0]] = roles
	}
	return result, nil
}

func (d *GCPSecretEngineRoleset) IsInitialized() bool {
	return true
}

func (d *GCPSecretEngineRoleset) IsValid() (bool, error) {
	return true, nil
}

func (d *GCPSecretEngineRoleset) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *GCPSecretEngineRoleset) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *GCPSecretEngineRoleset) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (i *GCPSERoleset) toMap() map[string]any {
	payload := map[string]any{}
	payload["secret_type"] = i.SecretType
	payload["project"] = i.Project
	payload["bindings"] = i.Bindings
	payload["token_scopes"] = toInterfaceArray(i.TokenScopes)
	return payload
}

// GCPSecretEngineRolesetStatus defines the observed state of GCPSecretEngineRoleset
type GCPSecretEngineRolesetStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *GCPSecretEngineRoleset) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *GCPSecretEngineRoleset) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GCPSecretEngineRoleset is the Schema for the gcpsecretenginerolesets API
type GCPSecretEngineRoleset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPSecretEngineRolesetSpec   `json:"spec,omitempty"`
	Status GCPSecretEngineRolesetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GCPSecretEngineRolesetList contains a list of GCPSecretEngineRoleset
type GCPSecretEngineRolesetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPSecretEngineRoleset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPSecretEngineRoleset{}, &GCPSecretEngineRolesetList{})
}
