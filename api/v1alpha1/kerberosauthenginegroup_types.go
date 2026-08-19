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
	"sort"
	"strings"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KerberosAuthEngineGroupSpec defines the desired state of KerberosAuthEngineGroup
type KerberosAuthEngineGroupSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the Kerberos auth engine is mounted.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// Name of the Kerberos LDAP group.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Policies is a comma-separated list of policies associated with the group.
	// +kubebuilder:validation:Optional
	Policies string `json:"policies,omitempty"`
}

var _ vaultutils.VaultObject = &KerberosAuthEngineGroup{}
var _ vaultutils.ConditionsAware = &KerberosAuthEngineGroup{}

func (d *KerberosAuthEngineGroup) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *KerberosAuthEngineGroup) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/groups/" + d.Spec.Name)
}

func (d *KerberosAuthEngineGroup) IsDeletable() bool {
	return true
}

func (d *KerberosAuthEngineGroup) GetPayload() map[string]any {
	return d.toMap()
}

func (d *KerberosAuthEngineGroup) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.GetPayload()
	filtered := filterPayloadToDesiredKeys(desiredState, payload)
	normalizePoliciesField(desiredState)
	normalizePoliciesField(filtered)
	return reflect.DeepEqual(desiredState, filtered)
}

// normalizePoliciesField converts the "policies" value in a map to a sorted
// comma-separated string so that a CR's comma-delimited string and Vault's
// JSON array representation compare as equal when they contain the same set.
func normalizePoliciesField(m map[string]any) {
	val, ok := m["policies"]
	if !ok {
		return
	}
	var policies []string
	switch v := val.(type) {
	case string:
		if v == "" {
			m["policies"] = ""
			return
		}
		for _, p := range strings.Split(v, ",") {
			policies = append(policies, strings.TrimSpace(p))
		}
	case []interface{}:
		for _, p := range v {
			if s, ok := p.(string); ok {
				policies = append(policies, s)
			}
		}
	case []string:
		policies = v
	default:
		return
	}
	sort.Strings(policies)
	m["policies"] = strings.Join(policies, ",")
}

func (d *KerberosAuthEngineGroup) IsInitialized() bool {
	return true
}

func (d *KerberosAuthEngineGroup) PrepareInternalValues(context context.Context, object client.Object) error {
	return nil
}

func (d *KerberosAuthEngineGroup) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *KerberosAuthEngineGroup) IsValid() (bool, error) {
	return true, nil
}

// KerberosAuthEngineGroupStatus defines the observed state of KerberosAuthEngineGroup
type KerberosAuthEngineGroupStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KerberosAuthEngineGroup is the Schema for the kerberosauthenginegroups API
type KerberosAuthEngineGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KerberosAuthEngineGroupSpec   `json:"spec,omitempty"`
	Status KerberosAuthEngineGroupStatus `json:"status,omitempty"`
}

func (m *KerberosAuthEngineGroup) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *KerberosAuthEngineGroup) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// KerberosAuthEngineGroupList contains a list of KerberosAuthEngineGroup
type KerberosAuthEngineGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KerberosAuthEngineGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KerberosAuthEngineGroup{}, &KerberosAuthEngineGroupList{})
}

func (i *KerberosAuthEngineGroup) toMap() map[string]any {
	payload := map[string]any{}
	payload["policies"] = i.Spec.Policies
	return payload
}

func (d *KerberosAuthEngineGroup) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
