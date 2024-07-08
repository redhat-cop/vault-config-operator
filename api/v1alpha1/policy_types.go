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
	"regexp"
	"strings"

	vault "github.com/hashicorp/vault/api"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var _ vaultutils.VaultObject = &Policy{}
var _ vaultutils.ConditionsAware = &PKISecretEngineRole{}

func (d *Policy) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *Policy) GetPath() string {
	if d.Spec.Name != "" {
		if d.Spec.Type != "" {
			return vaultutils.CleansePath("sys/policies/" + d.Spec.Type + "/" + d.Spec.Name)
		}
		return vaultutils.CleansePath("sys/policy/" + d.Spec.Name)
	}
	if d.Spec.Type != "" {
		return vaultutils.CleansePath("sys/policies/" + d.Spec.Type + "/" + d.Name)
	}
	return vaultutils.CleansePath("sys/policy/" + d.Name)
}
func (d *Policy) GetPayload() map[string]interface{} {
	return map[string]interface{}{
		"policy": d.Spec.Policy,
	}
}
func (d *Policy) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.GetPayload()
	desiredState["name"] = map[bool]string{true: d.Spec.Name, false: d.Name}[d.Spec.Name != ""]
	if d.Spec.Type == "" {
		desiredState["rules"] = desiredState["policy"]
		delete(desiredState, "policy")
	}
	return reflect.DeepEqual(desiredState, payload)
}

func (d *Policy) IsInitialized() bool {
	return true
}

func (d *Policy) IsDeletable() bool {
	return true
}

func (d *Policy) PrepareInternalValues(context context.Context, object client.Object) error {
	// Fast path escape if no "${..}" placeholder is detected
	match, err := regexp.MatchString("\\${[^}]+}", d.Spec.Policy)
	if err != nil || !match {
		return nil
	}

	log := log.FromContext(context)

	// Retrieves the list of auth engines to get their accessors
	// Kinda duplicates logic found in VaultEngineObject.retrieveAccessor
	vaultClient := context.Value("vaultClient").(*vault.Client)
	secret, err := vaultClient.Logical().Read("sys/auth")
	if err != nil {
		// Log but ignore the error: do not resolve placeholders
		log.Error(err, "could not resolve auth engine accessor(s) in policy rule - unable to retrieve auth engines at", "path", "sys/auth")
		return nil
	}
	if secret == nil {
		return errors.New("could not resolve auth engine accessor(s) in policy rule - listing auth engines at sys/auth unexpectedly returned null")
	}

	for key, data := range secret.Data {
		authenginepath := strings.Trim(key, "/")
		placeholder := "${auth/" + authenginepath + "/@accessor}"
		accessor := data.(map[string]interface{})["accessor"].(string)
		d.Spec.Policy = strings.ReplaceAll(d.Spec.Policy, placeholder, accessor)
	}

	log.V(1).Info("Auth engine accessor(s) resolved", "policy", d.Spec.Policy)
	return nil
}

func (d *Policy) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *Policy) IsValid() (bool, error) {
	return true, nil
}

// PolicySpec defines the desired state of Policy
type PolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Policy is a Vault policy expressed in HCL language.
	// +kubebuilder:validation:Required
	Policy string `json:"policy,omitempty"`

	// Type represents the policy type, currently the only supported policy type is "acl", but in the future rgp and egp  might be supported. If not specified a policy will be created at /sys/policies/<name>, if specified (the recommended approach) a policy will be created at /sys/policies/acl/<name>
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum={"acl"}
	Type string `json:"type,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

func (m *Policy) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *Policy) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Policy is the Schema for the policies API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}

func (d *Policy) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}
