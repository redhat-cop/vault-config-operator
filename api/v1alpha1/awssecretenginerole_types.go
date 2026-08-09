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
	"sort"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AWSSecretEngineRoleSpec defines the desired state of AWSSecretEngineRole
type AWSSecretEngineRoleSpec struct {

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the role.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/roles/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	AWSRole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type AWSRole struct {
	// CredentialType specifies the type of credential to be used when this role is requested (iam_user, assumed_role, federation_token, session_token)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:={"iam_user","assumed_role","federation_token","session_token"}
	CredentialType string `json:"credentialType"`

	// RoleArns specifies ARNs of AWS roles to be assumed (required for assumed_role)
	// +kubebuilder:validation:Optional
	// +listType=set
	RoleArns []string `json:"roleArns,omitempty"`

	// PolicyArns specifies AWS managed policy ARNs
	// +kubebuilder:validation:Optional
	// +listType=set
	PolicyArns []string `json:"policyArns,omitempty"`

	// PolicyDocument is the IAM policy document JSON
	// +kubebuilder:validation:Optional
	PolicyDocument string `json:"policyDocument,omitempty"`

	// IAMGroups specifies IAM group names. For iam_user, generated users are added
	// to these groups. For assumed_role and federation_token, the group policies are
	// combined with policyArns/policyDocument for the STS call. Not valid for session_token.
	// +kubebuilder:validation:Optional
	// +listType=set
	IAMGroups []string `json:"iamGroups,omitempty"`

	// IAMTags specifies key=value tags for iam_user
	// +kubebuilder:validation:Optional
	// +listType=set
	IAMTags []string `json:"iamTags,omitempty"`

	// DefaultSTSTTL is the default TTL for STS credentials
	// +kubebuilder:validation:Optional
	DefaultSTSTTL string `json:"defaultSTSTTL,omitempty"`

	// MaxSTSTTL is the max TTL for STS credentials
	// +kubebuilder:validation:Optional
	MaxSTSTTL string `json:"maxSTSTTL,omitempty"`

	// UserPath is the IAM user path (iam_user only)
	// +kubebuilder:validation:Optional
	UserPath string `json:"userPath,omitempty"`

	// PermissionsBoundaryARN is the permissions boundary (iam_user only)
	// +kubebuilder:validation:Optional
	PermissionsBoundaryARN string `json:"permissionsBoundaryARN,omitempty"`

	// ExternalID is the external ID for assume role
	// +kubebuilder:validation:Optional
	ExternalID string `json:"externalID,omitempty"`

	// SessionTags are STS session tags (assumed_role only)
	// +kubebuilder:validation:Optional
	// +listType=set
	SessionTags []string `json:"sessionTags,omitempty"`

	// MFASerialNumber is the MFA device ARN
	// +kubebuilder:validation:Optional
	MFASerialNumber string `json:"mfaSerialNumber,omitempty"`
}

var _ vaultutils.VaultObject = &AWSSecretEngineRole{}

var _ vaultutils.ConditionsAware = &AWSSecretEngineRole{}

func (d *AWSSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *AWSSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AWSSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}

func (d *AWSSecretEngineRole) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (d *AWSSecretEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.AWSRole.toMap()
	if ct, ok := desiredState["credential_type"]; ok {
		desiredState["credential_types"] = []any{ct}
		delete(desiredState, "credential_type")
	}
	removeUnsetFields(desiredState, payload)
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	setFields := []string{"role_arns", "policy_arns", "iam_groups", "iam_tags", "session_tags"}
	for _, key := range setFields {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func sortAnyStringSlice(m map[string]any, key string) {
	v, ok := m[key]
	if !ok {
		return
	}
	s, ok := v.([]any)
	if !ok || len(s) < 2 {
		return
	}
	sort.Slice(s, func(i, j int) bool {
		si, _ := s[i].(string)
		sj, _ := s[j].(string)
		return si < sj
	})
}

func (d *AWSSecretEngineRole) IsInitialized() bool {
	return true
}

func (d *AWSSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *AWSSecretEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AWSSecretEngineRole) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *AWSSecretEngineRole) isValid() error {
	if r.Spec.CredentialType != "assumed_role" {
		if r.Spec.ExternalID != "" {
			return errors.New("spec.externalID is only allowed when credentialType is assumed_role")
		}
		if len(r.Spec.SessionTags) > 0 {
			return errors.New("spec.sessionTags is only allowed when credentialType is assumed_role")
		}
	}
	if r.Spec.CredentialType != "iam_user" {
		if len(r.Spec.IAMTags) > 0 {
			return errors.New("spec.iamTags is only allowed when credentialType is iam_user")
		}
		if r.Spec.UserPath != "" {
			return errors.New("spec.userPath is only allowed when credentialType is iam_user")
		}
		if r.Spec.PermissionsBoundaryARN != "" {
			return errors.New("spec.permissionsBoundaryARN is only allowed when credentialType is iam_user")
		}
	}
	if r.Spec.CredentialType != "assumed_role" && r.Spec.CredentialType != "federation_token" {
		if r.Spec.DefaultSTSTTL != "" {
			return errors.New("spec.defaultSTSTTL is only allowed when credentialType is assumed_role or federation_token")
		}
		if r.Spec.MaxSTSTTL != "" {
			return errors.New("spec.maxSTSTTL is only allowed when credentialType is assumed_role or federation_token")
		}
	}
	if r.Spec.CredentialType == "session_token" {
		if len(r.Spec.IAMGroups) > 0 {
			return errors.New("spec.iamGroups is not allowed when credentialType is session_token")
		}
	}

	switch r.Spec.CredentialType {
	case "assumed_role":
		if len(r.Spec.RoleArns) == 0 {
			return errors.New("spec.roleArns is required when credentialType is assumed_role")
		}
	case "session_token":
		if len(r.Spec.PolicyArns) > 0 {
			return errors.New("spec.policyArns is not allowed when credentialType is session_token")
		}
		if r.Spec.PolicyDocument != "" {
			return errors.New("spec.policyDocument is not allowed when credentialType is session_token")
		}
		if len(r.Spec.RoleArns) > 0 {
			return errors.New("spec.roleArns is only allowed when credentialType is assumed_role")
		}
	case "iam_user", "federation_token":
		if len(r.Spec.RoleArns) > 0 {
			return errors.New("spec.roleArns is only allowed when credentialType is assumed_role")
		}
	}
	return nil
}

func (i *AWSRole) toMap() map[string]any {
	payload := map[string]any{}
	payload["credential_type"] = i.CredentialType
	payload["role_arns"] = toInterfaceArray(i.RoleArns)
	payload["policy_arns"] = toInterfaceArray(i.PolicyArns)
	payload["policy_document"] = i.PolicyDocument
	payload["iam_groups"] = toInterfaceArray(i.IAMGroups)
	payload["iam_tags"] = toInterfaceArray(i.IAMTags)
	payload["default_sts_ttl"] = i.DefaultSTSTTL
	payload["max_sts_ttl"] = i.MaxSTSTTL
	payload["user_path"] = i.UserPath
	payload["permissions_boundary_arn"] = i.PermissionsBoundaryARN
	payload["external_id"] = i.ExternalID
	payload["session_tags"] = toInterfaceArray(i.SessionTags)
	payload["mfa_serial_number"] = i.MFASerialNumber
	return payload
}

// AWSSecretEngineRoleStatus defines the observed state of AWSSecretEngineRole
type AWSSecretEngineRoleStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *AWSSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *AWSSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSSecretEngineRole is the Schema for the awssecretengineroles API
type AWSSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status AWSSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSSecretEngineRoleList contains a list of AWSSecretEngineRole
type AWSSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AWSSecretEngineRole{}, &AWSSecretEngineRoleList{})
}

func (d *AWSSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
