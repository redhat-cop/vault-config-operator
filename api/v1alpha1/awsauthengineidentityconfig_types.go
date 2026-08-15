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

// AWSAuthEngineIdentityConfigSpec defines the desired state of AWSAuthEngineIdentityConfig
type AWSAuthEngineIdentityConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/config/identity.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	AWSAuthIdentityConfig `json:",inline"`
}

type AWSAuthIdentityConfig struct {
	// IAMAlias controls identity alias generation for IAM auth.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="role_id"
	// +kubebuilder:validation:Enum={"role_id","unique_id","canonical_arn","full_arn"}
	IAMAlias string `json:"iamAlias"`

	// IAMMetadata is the metadata to include on the login token for IAM auth.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="default"
	IAMMetadata string `json:"iamMetadata"`

	// EC2Alias controls identity alias generation for EC2 auth.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="role_id"
	// +kubebuilder:validation:Enum={"role_id","instance_id","image_id"}
	EC2Alias string `json:"ec2Alias"`

	// EC2Metadata is the metadata to include on the login token for EC2 auth.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="default"
	EC2Metadata string `json:"ec2Metadata"`
}

// AWSAuthEngineIdentityConfigStatus defines the observed state of AWSAuthEngineIdentityConfig
type AWSAuthEngineIdentityConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSAuthEngineIdentityConfig is the Schema for the awsauthengineidentityconfigs API
type AWSAuthEngineIdentityConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSAuthEngineIdentityConfigSpec   `json:"spec,omitempty"`
	Status AWSAuthEngineIdentityConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSAuthEngineIdentityConfigList contains a list of AWSAuthEngineIdentityConfig
type AWSAuthEngineIdentityConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSAuthEngineIdentityConfig `json:"items"`
}

var _ vaultutils.VaultObject = &AWSAuthEngineIdentityConfig{}
var _ vaultutils.ConditionsAware = &AWSAuthEngineIdentityConfig{}

func (d *AWSAuthEngineIdentityConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *AWSAuthEngineIdentityConfig) IsDeletable() bool {
	return false
}

func (r *AWSAuthEngineIdentityConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AWSAuthEngineIdentityConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *AWSAuthEngineIdentityConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config/identity")
}

func (r *AWSAuthEngineIdentityConfig) GetPayload() map[string]any {
	return r.Spec.AWSAuthIdentityConfig.toMap()
}

func (r *AWSAuthEngineIdentityConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := r.Spec.AWSAuthIdentityConfig.toMap()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (r *AWSAuthEngineIdentityConfig) IsInitialized() bool {
	return true
}

func (r *AWSAuthEngineIdentityConfig) IsValid() (bool, error) {
	return true, nil
}

func (r *AWSAuthEngineIdentityConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *AWSAuthEngineIdentityConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (r *AWSAuthEngineIdentityConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (i *AWSAuthIdentityConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["iam_alias"] = i.IAMAlias
	payload["iam_metadata"] = i.IAMMetadata
	payload["ec2_alias"] = i.EC2Alias
	payload["ec2_metadata"] = i.EC2Metadata
	return payload
}

func init() {
	SchemeBuilder.Register(&AWSAuthEngineIdentityConfig{}, &AWSAuthEngineIdentityConfigList{})
}
