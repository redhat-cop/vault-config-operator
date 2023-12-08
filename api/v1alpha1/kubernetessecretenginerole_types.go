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

// KubernetesSecretEngineRoleSpec defines the desired state of KubernetesSecretEngineRole
type KubernetesSecretEngineRoleSpec struct {

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

	// TargetNamespaces specifies how to retrieve the list of Kubernetes namespaces this role can generate credentials for.
	// +kubebuilder:validation:Required
	TargetNamespaces vaultutils.TargetNamespaceConfig `json:"targetNamespaces,omitempty"`

	KubeSERole `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

var _ vaultutils.VaultObject = &KubernetesSecretEngineRole{}

var _ vaultutils.ConditionsAware = &KubernetesSecretEngineRole{}

func (d *KubernetesSecretEngineRole) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}
func (d *KubernetesSecretEngineRole) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}
func (d *KubernetesSecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.KubeSERole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (d *KubernetesSecretEngineRole) IsInitialized() bool {
	return true
}

func (d *KubernetesSecretEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *KubernetesSecretEngineRole) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *KubernetesSecretEngineRole) IsValid() (bool, error) {
	return true, nil
}

type KubeSERole struct {

	// AllowedKubernetesNamespaces The list of Kubernetes namespaces this role can generate credentials for. If set to "*" all namespaces are allowed.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	AllowedKubernetesNamespaces []string `json:"allowedKubernetesNamespaces,omitempty"`

	// A label selector for Kubernetes namespaces in which credentials can be generated.
	// Accepts either a JSON or YAML object. The value should be of type LabelSelector as illustrated: "'{'matchLabels':{'stage':'prod','sa-generator':'vault'}}".
	// If set with allowed_kubernetes_namespaces, the conditions are ORed.
	// +kubebuilder:validation:Optional
	AllowedKubernetesNamespaceSelector string `json:"allowedKubernetesNamespaceSelector,omitempty"`

	// DeafulTTL Specifies the TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/engine default TTL time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="0s"
	DefaultTTL metav1.Duration `json:"defaultTTL,omitempty"`

	// MaxTTL Specifies the maximum TTL for the leases associated with this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to system/mount default TTL time; this value is allowed to be less than the mount max TTL (or, if not set, the system max TTL), but it is not allowed to be longer. See also The TTL General Case.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="0s"
	MaxTTL metav1.Duration `json:"maxTTL,omitempty"`

	// DefaultAudiences The default intended audiences for generated Kubernetes tokens, specified by a comma separated string. e.g "custom-audience-0,custom-audience-1".
	// If not set or set to "", the Kubernetes cluster default for audiences of service account tokens will be used.
	// +kubebuilder:validation:Optional
	DefaultAudiences string `json:"defaultAudiences,omitempty"`

	// ServiceAccountName The pre-existing service account to generate tokens for. Mutually exclusive with all role parameters. If set, only a Kubernetes token will be created when credentials are requested. See the Kubernetes service account documentation for more details on service accounts.
	// +kubebuilder:validation:Optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// KubernetesRoleName The pre-existing Role or ClusterRole to bind a generated service account to. If set, Kubernetes token, service account, and role binding objects will be created when credentials are requested. See the Kubernetes roles documentation for more details on Kubernetes roles.
	// +kubebuilder:validation:Optional
	KubernetesRoleName string `json:"kubernetesRoleName,omitempty"`

	// KubernetesRoleType Specifies whether the Kubernetes role is a Role or ClusterRole
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="Role"
	// +kubebuilder:validation:Enum={"Role","ClusterRole"}
	KubernetesRoleType string `json:"kubernetesRoleType,omitempty"`

	// GenerateRoleRules The Role or ClusterRole rules to use when generating a role. Accepts either JSON or YAML formatted rules. If set, the entire chain of Kubernetes objects will be generated when credentials are requested. The value should be a rules key with an array of PolicyRule objects, as illustrated in the Kubernetes RBAC documentation and Sample Payload 3 below.
	// +kubebuilder:validation:Optional
	GenerateRoleRules string `json:"generateRoleRules,omitempty"`

	// NameTemplate The name template to use when generating service accounts, roles and role bindings. If unset, a default template is used. See username templating for details on how to write a custom template.
	// +kubebuilder:validation:Optional
	NameTemplate string `json:"nameTemplate,omitempty"`

	// ExtraAnnotations Additional annotations to apply to all generated Kubernetes objects. See the Kubernetes annotations documentation for more details on annotations.
	// +kubebuilder:validation:Optional
	ExtraAnnotations map[string]string `json:"extraAnnotations,omitempty"`

	// ExtraLabels Additional labels to apply to all generated Kubernetes objects. See the Kubernetes labels documentation for more details on labels.
	// +kubebuilder:validation:Optional
	ExtraLabels map[string]string `json:"extraLabels,omitempty"`
}

func (i *KubeSERole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["allowed_kubernetes_namespaces"] = i.AllowedKubernetesNamespaces
	payload["allowed_kubernetes_namespace_selector"] = i.AllowedKubernetesNamespaceSelector
	payload["token_max_ttl"] = i.DefaultTTL
	payload["token_default_ttl"] = i.MaxTTL
	payload["token_default_audiences"] = i.DefaultAudiences
	payload["service_account_name"] = i.ServiceAccountName
	payload["kubernetes_role_name"] = i.KubernetesRoleName
	payload["kubernetes_role_type"] = i.KubernetesRoleType
	payload["generated_role_rules"] = i.GenerateRoleRules
	payload["name_template"] = i.NameTemplate
	payload["extra_annotations"] = i.ExtraAnnotations
	payload["extra_labels"] = i.ExtraLabels
	return payload
}

// KubernetesSecretEngineRoleStatus defines the observed state of KubernetesSecretEngineRole
type KubernetesSecretEngineRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *KubernetesSecretEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *KubernetesSecretEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubernetesSecretEngineRole is the Schema for the kubernetessecretengineroles API
type KubernetesSecretEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesSecretEngineRoleSpec   `json:"spec,omitempty"`
	Status KubernetesSecretEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubernetesSecretEngineRoleList contains a list of KubernetesSecretEngineRole
type KubernetesSecretEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesSecretEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubernetesSecretEngineRole{}, &KubernetesSecretEngineRoleList{})
}

func (d *KubernetesSecretEngineRole) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *KubernetesSecretEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
