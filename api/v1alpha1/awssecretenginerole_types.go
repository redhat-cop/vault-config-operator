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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AWSSecretEngineRoleSpec defines the desired state of AWSSecretEngineRole
type AWSSecretEngineRoleSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/groups/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	AWSSERole `json:",inline"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// AWSSecretEngineRoleStatus defines the observed state of AWSSecretEngineRole
type AWSSecretEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
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

type AWSSERole struct {
	// Specifies the type of credential to be used when retrieving credentials from the role.
	// Must be one of iam_user, assumed_role, federation_token, or session_token.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=iam_user;assumed_role;federation_token;session_token
	CredentialType string `json:"credentialType,omitempty"`

	// Specifies the ARNs of the AWS roles this Vault role is allowed to assume.
	// Required when credential_type is assumed_role and prohibited otherwise.
	// +kubebuilder:validation:Optional
	RoleARNs []string `json:"roleArns,omitempty"`

	// Specifies a list of AWS managed policy ARNs. The behavior depends on the credential type.
	// With iam_user, the policies will be attached to IAM users when they are requested.
	// With assumed_role and federation_token, the policy ARNs will act as a filter on what the credentials can do.
	// +kubebuilder:validation:Optional
	PolicyARNs []string `json:"policyArns,omitempty"`

	// The IAM policy document for the role. The behavior depends on the credential type.
	// With iam_user, the policy document will be attached to the IAM user generated.
	// With assumed_role and federation_token, the policy document will act as a filter on what the credentials can do.
	// +kubebuilder:validation:Optional
	PolicyDocument string `json:"policyDocument,omitempty"`

	// A list of IAM group names. IAM users generated against this vault role will be added to these IAM Groups.
	// For a credential type of assumed_role or federation_token, the policies sent to the corresponding AWS call
	// will be the policies from each group in iam_groups combined with the policy_document and policy_arns parameters.
	// +kubebuilder:validation:Optional
	IAMGroups []string `json:"iamGroups,omitempty"`

	// A list of strings representing a key/value pair to be used as a tag for any iam_user user that is created by this role.
	// Format is a key and value separated by an = (e.g. test_key=value).
	// +kubebuilder:validation:Optional
	IAMTags []string `json:"iamTags,omitempty"`

	// The default TTL for STS credentials. When a TTL is not specified when STS credentials are requested,
	// and a default TTL is specified on the role, then this default TTL will be used.
	// Valid only when credential_type is one of assumed_role or federation_token.
	// +kubebuilder:validation:Optional
	DefaultSTSTTL string `json:"defaultStsTtl,omitempty"`

	// The max allowed TTL for STS credentials (credentials TTL are capped to max_sts_ttl).
	// Valid only when credential_type is one of assumed_role or federation_token.
	// +kubebuilder:validation:Optional
	MaxSTSTTL string `json:"maxStsTtl,omitempty"`

	// The set of key-value pairs to be included as tags for the STS session.
	// Format is key=value.
	// Valid only when credential_type is set to assumed_role.
	// +kubebuilder:validation:Optional
	SessionTags []string `json:"sessionTags,omitempty"`

	// The external ID to use when assuming the role.
	// Valid only when credential_type is set to assumed_role.
	// +kubebuilder:validation:Optional
	ExternalID string `json:"externalId,omitempty"`

	// The path for the user name. Valid only when credential_type is iam_user. Default is /
	// +kubebuilder:validation:Optional
	UserPath string `json:"userPath,omitempty"`

	// The ARN of the AWS Permissions Boundary to attach to IAM users created in the role.
	// Valid only when credential_type is iam_user. If not specified, then no permissions boundary policy will be attached.
	// +kubebuilder:validation:Optional
	PermissionsBoundaryARN string `json:"permissionsBoundaryArn,omitempty"`

	// The ARN or hardware device number of the device configured to the IAM user for multi-factor authentication.
	// Only required if the IAM user has an MFA device set up in AWS.
	// +kubebuilder:validation:Optional
	MFASerialNumber string `json:"mfaSerialNumber,omitempty"`
}

var _ vaultutils.VaultObject = &AWSSecretEngineRole{}
var _ vaultutils.ConditionsAware = &AWSSecretEngineRole{}

func init() {
	SchemeBuilder.Register(&AWSSecretEngineRole{}, &AWSSecretEngineRoleList{})
}

func (r *AWSSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (d *AWSSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}

func (d *AWSSecretEngineRole) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (d *AWSSecretEngineRole) IsDeletable() bool {
	return true
}

func (d *AWSSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AWSSecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.AWSSERole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (d *AWSSecretEngineRole) IsInitialized() bool {
	return true
}

func (r *AWSSecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (d *AWSSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *AWSSecretEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AWSSecretEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AWSSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (i *AWSSERole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["credential_type"] = i.CredentialType

	if len(i.RoleARNs) > 0 {
		payload["role_arns"] = i.RoleARNs
	}
	if len(i.PolicyARNs) > 0 {
		payload["policy_arns"] = i.PolicyARNs
	}
	if i.PolicyDocument != "" {
		payload["policy_document"] = i.PolicyDocument
	}
	if len(i.IAMGroups) > 0 {
		payload["iam_groups"] = i.IAMGroups
	}
	if len(i.IAMTags) > 0 {
		payload["iam_tags"] = i.IAMTags
	}
	if i.DefaultSTSTTL != "" {
		payload["default_sts_ttl"] = i.DefaultSTSTTL
	}
	if i.MaxSTSTTL != "" {
		payload["max_sts_ttl"] = i.MaxSTSTTL
	}
	if len(i.SessionTags) > 0 {
		payload["session_tags"] = i.SessionTags
	}
	if i.ExternalID != "" {
		payload["external_id"] = i.ExternalID
	}
	if i.UserPath != "" {
		payload["user_path"] = i.UserPath
	}
	if i.PermissionsBoundaryARN != "" {
		payload["permissions_boundary_arn"] = i.PermissionsBoundaryARN
	}
	if i.MFASerialNumber != "" {
		payload["mfa_serial_number"] = i.MFASerialNumber
	}

	return payload
}
