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
	"errors"
	"reflect"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TerraformCloudSecretEngineRoleSpec defines the desired state of TerraformCloudSecretEngineRole
type TerraformCloudSecretEngineRoleSpec struct {
	// Connection represents the information needed to connect to Vault.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/role/{metadata.name}.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	TFCSERole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type TFCSERole struct {
	// Organization is the name of the Terraform Cloud organization. Conflicts with UserID.
	// +kubebuilder:validation:Optional
	Organization string `json:"organization,omitempty"`

	// TeamID is the Terraform Cloud team ID. Conflicts with UserID.
	// +kubebuilder:validation:Optional
	TeamID string `json:"teamID,omitempty"`

	// UserID is the Terraform Cloud user ID. Conflicts with Organization and TeamID.
	// +kubebuilder:validation:Optional
	UserID string `json:"userID,omitempty"`

	// CredentialType specifies the type of credential to generate.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"team","team_legacy","user","organization"}
	CredentialType string `json:"credentialType,omitempty"`

	// Description is a human-readable description used as a prefix in HCP Terraform UI. Applies to User and Team tokens.
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`

	// TTL specifies the TTL for generated tokens (duration format, e.g. "1h")
	// +kubebuilder:validation:Optional
	TTL string `json:"ttl,omitempty"`

	// MaxTTL specifies the maximum TTL for generated tokens (duration format, e.g. "24h")
	// +kubebuilder:validation:Optional
	MaxTTL string `json:"maxTTL,omitempty"`
}

var _ vaultutils.VaultObject = &TerraformCloudSecretEngineRole{}
var _ vaultutils.ConditionsAware = &TerraformCloudSecretEngineRole{}

func (d *TerraformCloudSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *TerraformCloudSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *TerraformCloudSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "role" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "role" + "/" + d.Name)
}

func (d *TerraformCloudSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.TFCSERole.toMap()
}

func (d *TerraformCloudSecretEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.TFCSERole.toMap()
	removeUnsetFields(desiredState, payload)
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *TerraformCloudSecretEngineRole) IsInitialized() bool {
	return true
}

func (d *TerraformCloudSecretEngineRole) IsValid() (bool, error) {
	err := d.isValid()
	return err == nil, err
}

func (d *TerraformCloudSecretEngineRole) isValid() error {
	if d.Spec.CredentialType == "" {
		return errors.New("spec.credentialType is required and must be one of: team, team_legacy, user, organization")
	}

	validCredentialTypes := map[string]bool{
		"team":         true,
		"team_legacy":  true,
		"user":         true,
		"organization": true,
	}

	if !validCredentialTypes[d.Spec.CredentialType] {
		return errors.New("spec.credentialType must be one of: team, team_legacy, user, organization")
	}

	if d.Spec.UserID != "" && (d.Spec.Organization != "" || d.Spec.TeamID != "") {
		return errors.New("spec.userID conflicts with spec.organization and spec.teamID")
	}

	switch d.Spec.CredentialType {
	case "user":
		if d.Spec.UserID == "" {
			return errors.New("spec.userID is required when credentialType is \"user\"")
		}
	case "team", "team_legacy":
		if d.Spec.Organization == "" {
			return errors.New("spec.organization is required when credentialType is \"team\" or \"team_legacy\"")
		}
		if d.Spec.TeamID == "" {
			return errors.New("spec.teamID is required when credentialType is \"team\" or \"team_legacy\"")
		}
	case "organization":
		if d.Spec.Organization == "" {
			return errors.New("spec.organization is required when credentialType is \"organization\"")
		}
	}

	return nil
}

func (d *TerraformCloudSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *TerraformCloudSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *TerraformCloudSecretEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (d *TFCSERole) toMap() map[string]any {
	payload := map[string]any{}
	if d.Organization != "" {
		payload["organization"] = d.Organization
	}
	if d.TeamID != "" {
		payload["team_id"] = d.TeamID
	}
	if d.UserID != "" {
		payload["user_id"] = d.UserID
	}
	if d.CredentialType != "" {
		payload["credential_type"] = d.CredentialType
	}
	if d.Description != "" {
		payload["description"] = d.Description
	}
	if d.TTL != "" {
		payload["ttl"] = durationToSeconds(d.TTL)
	}
	if d.MaxTTL != "" {
		payload["max_ttl"] = durationToSeconds(d.MaxTTL)
	}
	return payload
}

// TerraformCloudSecretEngineRoleStatus defines the observed state of TerraformCloudSecretEngineRole
type TerraformCloudSecretEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *TerraformCloudSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *TerraformCloudSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TerraformCloudSecretEngineRole is the Schema for the terraformcloudsecretengineroles API
type TerraformCloudSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformCloudSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status TerraformCloudSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TerraformCloudSecretEngineRoleList contains a list of TerraformCloudSecretEngineRole
type TerraformCloudSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TerraformCloudSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TerraformCloudSecretEngineRole{}, &TerraformCloudSecretEngineRoleList{})
}
