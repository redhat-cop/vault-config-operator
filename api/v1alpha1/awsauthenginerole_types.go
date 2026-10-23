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

// AWSAuthEngineRoleSpec defines the desired state of AWSAuthEngineRole
type AWSAuthEngineRoleSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/auth/{spec.path}/role/{spec.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	AWSAuthRole `json:",inline"`
}

type AWSAuthRole struct {
	// Name of the role.
	// +kubebuilder:validation:Required
	Name string `json:"name" vault:"-"`

	// AuthType specifies the auth type for this role (iam or ec2).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="iam"
	// +kubebuilder:validation:Enum={"iam","ec2"}
	AuthType string `json:"authType,omitempty" vault:"auth_type"`

	// BoundAmiID constrains EC2 instances to specific AMI IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundAmiID []string `json:"boundAmiID,omitempty" vault:"bound_ami_id"`

	// BoundAccountID constrains EC2 instances to specific account IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundAccountID []string `json:"boundAccountID,omitempty" vault:"bound_account_id"`

	// BoundRegion constrains EC2 instances to specific regions.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundRegion []string `json:"boundRegion,omitempty" vault:"bound_region"`

	// BoundVpcID constrains EC2 instances to specific VPC IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundVpcID []string `json:"boundVpcID,omitempty" vault:"bound_vpc_id"`

	// BoundSubnetID constrains EC2 instances to specific subnet IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundSubnetID []string `json:"boundSubnetID,omitempty" vault:"bound_subnet_id"`

	// BoundIAMRoleARN constrains EC2 instances to specific IAM role ARNs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundIAMRoleARN []string `json:"boundIAMRoleARN,omitempty" vault:"bound_iam_role_arn"`

	// BoundIAMInstanceProfileARN constrains EC2 instances to specific instance profile ARNs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundIAMInstanceProfileARN []string `json:"boundIAMInstanceProfileARN,omitempty" vault:"bound_iam_instance_profile_arn"`

	// BoundEC2InstanceID constrains to specific EC2 instance IDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundEC2InstanceID []string `json:"boundEC2InstanceID,omitempty" vault:"bound_ec2_instance_id"`

	// RoleTag enables role tags for EC2 auth. Value is the tag key on the instance.
	// +kubebuilder:validation:Optional
	RoleTag string `json:"roleTag,omitempty" vault:"role_tag"`

	// BoundIAMPrincipalARN constrains IAM auth to specific principal ARNs. Wildcards supported.
	// +kubebuilder:validation:Optional
	// +listType=set
	BoundIAMPrincipalARN []string `json:"boundIAMPrincipalARN,omitempty" vault:"bound_iam_principal_arn"`

	// InferredEntityType enables IAM role inferencing. Only valid value: "ec2_instance".
	// +kubebuilder:validation:Optional
	InferredEntityType string `json:"inferredEntityType,omitempty" vault:"inferred_entity_type"`

	// InferredAWSRegion is the region to search for inferred entities. Required with inferredEntityType.
	// +kubebuilder:validation:Optional
	InferredAWSRegion string `json:"inferredAWSRegion,omitempty" vault:"inferred_aws_region"`

	// ResolveAWSUniqueIDs resolves bound_iam_principal_arn to AWS unique IDs.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	ResolveAWSUniqueIDs bool `json:"resolveAWSUniqueIDs" vault:"resolve_aws_unique_ids,when=AuthType:iam"`

	// AllowInstanceMigration allows migration of the underlying EC2 instance. EC2 auth only.
	// +kubebuilder:validation:Optional
	AllowInstanceMigration bool `json:"allowInstanceMigration,omitempty" vault:"allow_instance_migration"`

	// DisallowReauthentication if true, only allows a single token per instance ID. EC2 auth only.
	// +kubebuilder:validation:Optional
	DisallowReauthentication bool `json:"disallowReauthentication,omitempty" vault:"disallow_reauthentication"`

	// TokenTTL is the incremental lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty" vault:"token_ttl,duration"`

	// TokenMaxTTL is the maximum lifetime for generated tokens.
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty" vault:"token_max_ttl,duration"`

	// TokenPolicies are policies to encode onto generated tokens.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenPolicies []string `json:"tokenPolicies,omitempty" vault:"token_policies"`

	// Policies is deprecated — use tokenPolicies instead.
	// +kubebuilder:validation:Optional
	// +listType=set
	Policies []string `json:"policies,omitempty" vault:"policies"`

	// TokenBoundCIDRs are CIDR blocks that restrict authentication and tie the token.
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty" vault:"token_bound_cidrs"`

	// TokenExplicitMaxTTL is the hard cap max TTL for tokens.
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty" vault:"token_explicit_max_ttl,duration"`

	// TokenNoDefaultPolicy if true, omits the default policy from generated tokens.
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty" vault:"token_no_default_policy"`

	// TokenNumUses is the max number of times a token may be used (0 = unlimited).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int64 `json:"tokenNumUses,omitempty" vault:"token_num_uses,number"`

	// TokenPeriod is the maximum allowed period for periodic tokens.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenPeriod int64 `json:"tokenPeriod,omitempty" vault:"token_period,number"`

	// TokenType is the type of token to generate (service, batch, default).
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
	TokenType string `json:"tokenType,omitempty" vault:"token_type"`
}

// AWSAuthEngineRoleStatus defines the observed state of AWSAuthEngineRole
type AWSAuthEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSAuthEngineRole is the Schema for the awsauthengineroles API
type AWSAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status AWSAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSAuthEngineRoleList contains a list of AWSAuthEngineRole
type AWSAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSAuthEngineRole `json:"items"`
}

var _ vaultutils.VaultObject = &AWSAuthEngineRole{}
var _ vaultutils.ConditionsAware = &AWSAuthEngineRole{}

func init() {
	SchemeBuilder.Register(&AWSAuthEngineRole{}, &AWSAuthEngineRoleList{})
}

func (d *AWSAuthEngineRole) IsDeletable() bool {
	return true
}

func (r *AWSAuthEngineRole) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AWSAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (d *AWSAuthEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *AWSAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *AWSAuthEngineRole) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/role/" + r.Spec.Name)
}

func (r *AWSAuthEngineRole) GetPayload() map[string]any {
	return r.Spec.AWSAuthRole.toMap()
}

func (r *AWSAuthEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.AWSAuthRole.toMap()
	removeUnsetFields(desiredState, payload)
	boolFields := []string{"resolve_aws_unique_ids", "allow_instance_migration", "disallow_reauthentication", "token_no_default_policy"}
	for _, key := range boolFields {
		if boolVal, ok := desiredState[key].(bool); ok && !boolVal {
			if _, inPayload := payload[key]; !inPayload {
				delete(desiredState, key)
			}
		}
	}
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (r *AWSAuthRole) toMap() map[string]any {
	return vaultutils.StructToVaultPayload(r)
}
