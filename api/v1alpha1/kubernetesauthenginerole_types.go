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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubernetesAuthEngineRoleSpec defines the desired state of KubernetesAuthEngineRole
type KubernetesAuthEngineRoleSpec struct {

	// Authentication is the kube aoth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path will be {[spec.authentication.namespace]}/auth/{spec.path}/role/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	VRole `json:",inline"`

	// TargetNamespaces specifies how to retrieve the namespaces bound to this Vault role.
	// +kubebuilder:validation:Required
	TargetNamespaces vaultutils.TargetNamespaceConfig `json:"targetNamespaces,omitempty"`
}

var _ vaultutils.VaultObject = &KubernetesAuthEngineRole{}

func (d *KubernetesAuthEngineRole) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Name)
}
func (d *KubernetesAuthEngineRole) GetPayload() map[string]interface{} {
	return d.Spec.VRole.toMap()
}
func (d *KubernetesAuthEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.VRole.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (d *KubernetesAuthEngineRole) IsInitialized() bool {
	return true
}

func (d *KubernetesAuthEngineRole) PrepareInternalValues(context context.Context, object client.Object) error {
	log := log.FromContext(context)
	if d.Spec.TargetNamespaces.TargetNamespaceSelector != nil {
		namespaces, err := d.findSelectedNamespaceNames(context)
		if err != nil {
			log.Error(err, "unable to retrieve selected namespaces", "instance", object)
			return err
		}
		if len(namespaces) < 1 {
			// workaround for when there are no namespace, as we don't want this call to error out
			d.SetInternalNamespaces([]string{"__no_namespace__"})
		} else {
			d.SetInternalNamespaces(namespaces)
		}
	} else {
		d.SetInternalNamespaces(d.Spec.TargetNamespaces.TargetNamespaces)
	}
	return nil
}

func (r *KubernetesAuthEngineRole) findSelectedNamespaceNames(context context.Context) ([]string, error) {
	log := log.FromContext(context)
	result := []string{}
	namespaceList := &corev1.NamespaceList{}
	labelSelector, err := metav1.LabelSelectorAsSelector(r.Spec.TargetNamespaces.TargetNamespaceSelector)
	if err != nil {
		log.Error(err, "unable to create selector from label selector", "selector", r.Spec.TargetNamespaces.TargetNamespaceSelector)
		return nil, err
	}
	kubeClient := context.Value("kubeClient").(client.Client)
	err = kubeClient.List(context, namespaceList, &client.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Error(err, "unable to retrieve the list of namespaces")
		return nil, err
	}
	for i := range namespaceList.Items {
		result = append(result, namespaceList.Items[i].Name)
	}
	return result, nil
}

func (r *KubernetesAuthEngineRole) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

type VRole struct {

	// TargetServiceAccounts is a list of service account names that will receive this role
	// +kubebuilder:validation:MinItems=1
	// kubebuilder:validation:UniqueItems=true
	// +kubebuilder:default={"default"}
	TargetServiceAccounts []string `json:"targetServiceAccounts"`

	// Audience Audience claim to verify in the JWT.
	// +kubebuilder:validation:Optional
	Audience *string `json:"audience,omitempty"`

	// AliasNameSource Configures how identity aliases are generated. Valid choices are: serviceaccount_uid, serviceaccount_name When serviceaccount_uid is specified, the machine generated UID from the service account will be used as the identity alias name. When serviceaccount_name is specified, the service account's namespace and name will be used as the identity alias name e.g vault/vault-auth. While it is strongly advised that you use serviceaccount_uid, you may also use serviceaccount_name in cases where you want to set the alias ahead of time, and the risks are mitigated or otherwise acceptable given your use case. It is very important to limit who is able to delete/create service accounts within a given cluster. See the Create an Entity Alias document which further expands on the potential security implications mentioned above.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"serviceaccount_uid", "serviceaccount_name"}
	// +kubebuilder:default={"serviceaccount_uid"}
	AliasNameSource string `json:"aliasNameSource,omitempty"`

	// TokenTTL The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=0
	TokenTTL int `json:"tokenTTL,omitempty"`

	// Policies is a list of policy names to be bound to this role.
	// +kubebuilder:validation:MinItems=1
	// kubebuilder:validation:UniqueItems=true
	// +kubebuilder:validation:Required
	Policies []string `json:"policies"`

	// TokenMaxTTL The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=0
	TokenMaxTTL int `json:"tokenMaxTTL,omitempty"`

	// TokenBoundCIDRs List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

	// TokenExplicitMaxTTL If set, will encode an explicit max TTL onto the token. This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=0
	TokenExplicitMaxTTL int `json:"tokenExplicitMaxTTL,omitempty"`

	// TokenNoDefaultPolicy If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`

	// TokenNumUses The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited. If you require the token to have the ability to create child tokens, you will need to set this value to 0.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=0
	TokenNumUses int `json:"tokenNumUses,omitempty"`

	// TokenPeriod The period, if any, to set on the token.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=0
	TokenPeriod int `json:"tokenPeriod,omitempty"`

	// TokenType The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens). For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"service","batch","default","default-service","default-batch"}
	// +kubebuilder:default={"default"}
	TokenType string `json:"tokenType,omitempty"`

	// this field is for internal use and will not be serialized
	namespaces []string `json:"-"`
}

// KubernetesAuthEngineRoleStatus defines the observed state of KubernetesAuthEngineRole
type KubernetesAuthEngineRoleStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *KubernetesAuthEngineRole) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *KubernetesAuthEngineRole) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (m *KubernetesAuthEngineRole) SetInternalNamespaces(namespaces []string) {
	m.Spec.namespaces = namespaces
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubernetesAuthEngineRole can be used to define a KubernetesAuthEngineRole for the kube-auth authentication method
type KubernetesAuthEngineRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesAuthEngineRoleSpec   `json:"spec,omitempty"`
	Status KubernetesAuthEngineRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubernetesAuthEngineRoleList contains a list of KubernetesAuthEngineRole
type KubernetesAuthEngineRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesAuthEngineRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubernetesAuthEngineRole{}, &KubernetesAuthEngineRoleList{})
}

func (i *VRole) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["bound_service_account_names"] = i.TargetServiceAccounts
	payload["bound_service_account_namespaces"] = i.namespaces
	payload["alias_name_source"] = i.AliasNameSource
	if i.Audience != nil {
		payload["audience"] = i.Audience
	}
	payload["token_ttl"] = i.TokenTTL
	payload["token_max_ttl"] = i.TokenMaxTTL
	payload["token_policies"] = i.Policies
	payload["token_bound_cidrs"] = i.TokenBoundCIDRs
	payload["token_explicit_max_ttl"] = i.TokenExplicitMaxTTL
	payload["token_no_default_policy"] = i.TokenNoDefaultPolicy
	payload["token_num_uses"] = i.TokenNumUses
	payload["token_period"] = i.TokenPeriod
	payload["token_type"] = i.TokenType
	return payload
}

func (r *KubernetesAuthEngineRole) isValid() error {
	return r.validateEitherTargetNamespaceSelectorOrTargetNamespace()
}

func (r *KubernetesAuthEngineRole) validateEitherTargetNamespaceSelectorOrTargetNamespace() error {
	count := 0
	if r.Spec.TargetNamespaces.TargetNamespaceSelector != nil {
		count++
	}
	if r.Spec.TargetNamespaces.TargetNamespaces != nil {
		count++
	}
	if count != 1 {
		return errors.New("only one of TargetNamespaceSelector or TargetNamespaces can be specified")
	}
	return nil
}

func (d *KubernetesAuthEngineRole) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
