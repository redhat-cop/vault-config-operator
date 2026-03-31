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

	vault "github.com/hashicorp/vault/api"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EntityAliasSpec defines the desired state of EntityAlias
type EntityAliasSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	EntityAliasConfig `json:",inline"`

	retrievedMountAccessor string `json:"-"`

	retrievedCanonicalID string `json:"-"`

	retrievedAliasID string `json:"-"`

	retrievedName string `json:"-"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type EntityAliasConfig struct {
	// AuthEngineMountPath is the path where the auth engine is mounted
	// +kubebuilder:validation:Required
	AuthEngineMountPath string `json:"authEngineMountPath,omitempty"`

	// EntityName is the name of the entity to which this alias belongs
	// +kubebuilder:validation:Required
	EntityName string `json:"entityName,omitempty"`

	// CustomMetadata is a map of arbitrary string to string valued user-provided metadata meant to describe the alias.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	CustomMetadata map[string]string `json:"customMetadata,omitempty"`
}

// EntityAliasStatus defines the observed state of EntityAlias
type EntityAliasStatus struct {
	// +kubebuilder:validation:Optional
	ID string `json:"id,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EntityAlias is the Schema for the entityaliases API
type EntityAlias struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EntityAliasSpec   `json:"spec,omitempty"`
	Status EntityAliasStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EntityAliasList contains a list of EntityAlias
type EntityAliasList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EntityAlias `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EntityAlias{}, &EntityAliasList{})
}

var _ vaultutils.VaultObject = &EntityAlias{}
var _ vaultutils.ConditionsAware = &EntityAlias{}

func (m *EntityAlias) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *EntityAlias) IsDeletable() bool {
	return true
}

func (m *EntityAlias) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *EntityAlias) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *EntityAlias) GetPath() string {
	return vaultutils.CleansePath("/identity/entity-alias/id/" + d.Status.ID)
}

func (d *EntityAlias) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (i *EntityAliasSpec) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["name"] = i.retrievedName
	payload["id"] = i.retrievedAliasID
	payload["mount_accessor"] = i.retrievedMountAccessor
	payload["canonical_id"] = i.retrievedCanonicalID
	if len(i.CustomMetadata) > 0 {
		payload["custom_metadata"] = i.CustomMetadata
	}
	return payload
}

func (d *EntityAlias) IsInitialized() bool {
	return true
}

func (d *EntityAlias) PrepareInternalValues(context context.Context, object client.Object) error {
	log := log.FromContext(context)
	// let find the auth engine mount accessor
	secret, found, err := vaultutils.ReadSecret(context, vaultutils.CleansePath("sys/auth/"+d.Spec.AuthEngineMountPath))
	if err != nil {
		log.Error(err, "unable to retrieve authEngineMount", "path", d.Spec.AuthEngineMountPath)
		return err
	}
	if !found {
		err = errors.New("auth engine not found")
		log.Error(err, "authEngineMount not found at path", "path", d.Spec.AuthEngineMountPath)
		return err
	}
	d.Spec.retrievedMountAccessor = secret.Data["accessor"].(string)

	secret, found, err = vaultutils.ReadSecret(context, vaultutils.CleansePath("/identity/entity/name/"+d.Spec.EntityName))
	if err != nil {
		log.Error(err, "unable to retrieve entity", "name", d.Spec.EntityName)
		return err
	}
	if !found {
		err = errors.New("entity not found")
		log.Error(err, "entity not found", "name", d.Spec.EntityName)
		return err
	}
	d.Spec.retrievedCanonicalID = secret.Data["id"].(string)
	if d.Spec.Name != "" {
		d.Spec.retrievedName = d.Spec.Name
	} else {
		d.Spec.retrievedName = d.Name
	}

	if d.Status.ID == "" {
		//we have to create the entity alias as unfortunately this api is asymmetric
		payload := map[string]interface{}{
			"name":           map[bool]string{true: d.Spec.Name, false: d.Name}[d.Spec.Name != ""],
			"mount_accessor": d.Spec.retrievedMountAccessor,
			"canonical_id":   d.Spec.retrievedCanonicalID,
		}
		if len(d.Spec.CustomMetadata) > 0 {
			payload["custom_metadata"] = d.Spec.CustomMetadata
		}
		log.V(1).Info("create entity alias", "payload", payload)
		vaultClient := context.Value("vaultClient").(*vault.Client)
		result, err := vaultClient.Logical().Write("/identity/entity-alias", payload)
		if err != nil {
			log.Error(err, "unable to create entity alias", "entity alias", d.Spec)
			return err
		}
		d.Status.ID = result.Data["id"].(string)
		kubeClient := context.Value("kubeClient").(client.Client)
		err = kubeClient.Status().Update(context, d, &client.SubResourceUpdateOptions{})
		if err != nil {
			log.Error(err, "unable to update entity alias status, your kube and vault systems may now be inconsistent", "instance", d)
			return err
		}
	}

	d.Spec.retrievedAliasID = d.Status.ID
	return nil
}

func (d *EntityAlias) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *EntityAlias) IsValid() (bool, error) {
	return true, nil
}

func (d *EntityAlias) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *EntityAlias) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := d.Spec.toMap()
	delete(payload, "creation_time")
	delete(payload, "last_update_time")
	delete(payload, "merged_from_canonical_ids")
	delete(payload, "metadata")
	delete(payload, "mount_path")
	delete(payload, "mount_type")
	delete(payload, "local")
	delete(payload, "namespace_id")
	return reflect.DeepEqual(desiredState, payload)
}
