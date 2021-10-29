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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RabbitMQSecretEngineRoleSpec defines the desired state of RabbitMQSecretEngineRole
type RabbitMQSecretEngineRoleSpec struct {
	// Authentication is the k8s auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path will be {[spec.authentication.namespace]}/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path Path `json:"path,omitempty"`

	// +kubebuilder:validation:Required
	RMQSERole `json:",inline"`
}

type RMQSERole struct {
	// Comma-separated RabbitMQ permissions tags to associate with the user. This determines the level of
	// access to the RabbitMQ management UI granted to the user. Omitting this field will
	// lead to a user than can still connect to the cluster through messaging protocols,
	// but cannot perform any management actions.
	// +kubebuilder:validation:Optional
	Tags []string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	Vhosts []Vhost `json:",inline"`

	// This option requires RabbitMQ 3.7.0 or later.
	// +kubebuilder:validation:Optional
	VhostTopics []VhostTopic `json:",inline"`
}

type Vhost struct {
	// Name of an existing vhost; required property
	// +kubebuilder:validation:Required
	VhostName string `json:"vhostName"`
	// Permissions to grant to the user in the specific vhost; required property.
	// +kubebuilder:validation:Required
	Permissions VhostPermissions `json:"permissions"`
}

type VhostTopic struct {
	// Name of an existing vhost; required property
	// +kubebuilder:validation:Required
	VhostName string `json:"vhostName"`

	// List of topics to provide
	// +kubebuilder:validation:Required
	Topics []Topic `json:",inline"`
}

type Topic struct {
	// Name of an existing topic; required property
	// +kubebuilder:validation:Required
	TopicName string `json:"vhostName"`
	
	// Permissions to grant to the user in the specific vhost
	// +kubebuilder:validation:Required
	Permissions VhostPermissions `json:"permissions"`
}

// Set of RabbitMQ permissions: configure, read and write.
// By not setting a property (configure/write/read), it result in an empty string which does not match any permission.
type VhostPermissions struct {
	// +kubebuilder:validation:Optional
	Configure string `json:"configure,omitempty"`
	// +kubebuilder:validation:Optional
	Write string `json:"write,omitempty"`
	// +kubebuilder:validation:Optional
	Read string `json:"read,omitempty"`
}

// RabbitMQSecretEngineRoleStatus defines the observed state of RabbitMQSecretEngineRole
type RabbitMQSecretEngineRoleStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RabbitMQSecretEngineRole is the Schema for the rabbitmqsecretengineroles API
type RabbitMQSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status RabbitMQSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RabbitMQSecretEngineRoleList contains a list of RabbitMQSecretEngineRole
type RabbitMQSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RabbitMQSecretEngineRole{}, &RabbitMQSecretEngineRoleList{})
}

var _ vaultutils.VaultObject = &RabbitMQSecretEngineRole{}

func (rabbitMQ *RabbitMQSecretEngineRole) GetPath() string {
	return string(rabbitMQ.Spec.Path) + "/" + "roles" + "/" + rabbitMQ.Name
}
func (rabbitMQ *RabbitMQSecretEngineRole) GetPayload() map[string]interface{} {
	return rabbitMQ.Spec.rabbitMQToMap()
}
func (rabbitMQ *RabbitMQSecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := rabbitMQ.Spec.RMQSERole.rabbitMQToMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (rabbitMQ *RabbitMQSecretEngineRole) IsInitialized() bool {
	return true
}

func (rabbitMQ *RabbitMQSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (rabbitMQ *RabbitMQSecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

func (fields *RMQSERole) rabbitMQToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["tags"] = fields.Tags
	payload["vhosts"] = fields.Vhosts
	payload["vhost_topics"] = fields.VhostTopics
	return payload
}
