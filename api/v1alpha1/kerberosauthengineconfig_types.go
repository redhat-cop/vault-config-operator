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
	"fmt"
	"reflect"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// KerberosAuthEngineConfigSpec defines the desired state of KerberosAuthEngineConfig
type KerberosAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which the Kerberos auth engine is mounted.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	KerberosAuthConfig `json:",inline"`

	// KeytabSecret references a Kubernetes Secret containing the base64-encoded keytab file.
	// The Secret must contain a key (default: "keytab") with the base64 keytab content.
	// +kubebuilder:validation:Required
	KeytabSecret corev1.LocalObjectReference `json:"keytabSecret"`

	// KeytabKey is the key within the KeytabSecret that contains the base64-encoded keytab.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="keytab"
	KeytabKey string `json:"keytabKey"`
}

type KerberosAuthConfig struct {
	// ServiceAccount is the service account associated with both the keytab entry and an LDAP service account.
	// +kubebuilder:validation:Required
	ServiceAccount string `json:"serviceAccount"`

	// RemoveInstanceName strips instance names from a Kerberos service principal name when parsing the keytab.
	// +kubebuilder:validation:Optional
	RemoveInstanceName bool `json:"removeInstanceName,omitempty"`

	// AddGroupAliases adds any LDAP groups found for the user as group aliases.
	// +kubebuilder:validation:Optional
	AddGroupAliases bool `json:"addGroupAliases,omitempty"`

	retrievedKeytab string `json:"-"`
}

// KerberosAuthEngineConfigStatus defines the observed state of KerberosAuthEngineConfig
type KerberosAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KerberosAuthEngineConfig is the Schema for the kerberosauthengineconfigs API
type KerberosAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KerberosAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status KerberosAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KerberosAuthEngineConfigList contains a list of KerberosAuthEngineConfig
type KerberosAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KerberosAuthEngineConfig `json:"items"`
}

var _ vaultutils.VaultObject = &KerberosAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &KerberosAuthEngineConfig{}

func (d *KerberosAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *KerberosAuthEngineConfig) IsDeletable() bool {
	return false
}

func (d *KerberosAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/config")
}

func (d *KerberosAuthEngineConfig) GetPayload() map[string]any {
	return d.Spec.KerberosAuthConfig.toMap()
}

func (d *KerberosAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.KerberosAuthConfig.toMap()
	delete(desiredState, "keytab")
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

func (d *KerberosAuthEngineConfig) IsInitialized() bool {
	return true
}

func (d *KerberosAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.setKeytabFromSecret(context)
}

func (d *KerberosAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *KerberosAuthEngineConfig) IsValid() (bool, error) {
	return true, nil
}

func (r *KerberosAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *KerberosAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *KerberosAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (d *KerberosAuthEngineConfig) setKeytabFromSecret(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := vaultutils.KubeClientFromContext(context)

	secret := &corev1.Secret{}
	err := kubeClient.Get(context, types.NamespacedName{
		Namespace: d.Namespace,
		Name:      d.Spec.KeytabSecret.Name,
	}, secret)
	if err != nil {
		log.Error(err, "unable to retrieve keytab Secret", "instance", d)
		return err
	}

	keytabKey := d.Spec.KeytabKey
	if keytabKey == "" {
		keytabKey = "keytab"
	}

	keytabBytes, ok := secret.Data[keytabKey]
	if !ok {
		return fmt.Errorf("keytab Secret %q missing key %q", d.Spec.KeytabSecret.Name, keytabKey)
	}
	d.Spec.KerberosAuthConfig.retrievedKeytab = string(keytabBytes)
	return nil
}

func (c *KerberosAuthConfig) toMap() map[string]any {
	payload := map[string]any{}
	payload["keytab"] = c.retrievedKeytab
	payload["service_account"] = c.ServiceAccount
	payload["remove_instance_name"] = c.RemoveInstanceName
	payload["add_group_aliases"] = c.AddGroupAliases
	return payload
}

func init() {
	SchemeBuilder.Register(&KerberosAuthEngineConfig{}, &KerberosAuthEngineConfigList{})
}
