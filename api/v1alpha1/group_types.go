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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GroupSpec defines the desired state of Group
type GroupSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	GroupConfig `json:",inline"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type GroupConfig struct {

	// Type Type of the group, internal or external. Defaults to internal
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"internal","external"}
	// +kubebuilder:default:="internal"
	Type string `json:"type,omitempty"`

	// Metadata Metadata to be associated with the group.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Metadata map[string]string `json:"metadata,omitempty"`

	// Policies Policies to be tied to the group.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	Policies []string `json:"policies,omitempty"`

	// MemberGroupIDs Group IDs to be assigned as group members.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	MemberGroupIDs []string `json:"memberGroupIDs,omitempty"`

	// MemberGroupNames Group names to be resolved to IDs and assigned as group members.
	// Each name is looked up via Vault's identity API at reconciliation time.
	// The resolved IDs are merged with any explicitly provided MemberGroupIDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	MemberGroupNames []string `json:"memberGroupNames,omitempty"`

	// MemberEntityIDs Entity IDs to be assigned as group members.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	MemberEntityIDs []string `json:"memberEntityIDs,omitempty"`

	// MemberEntityNames Entity names to be resolved to IDs and assigned as group members.
	// Each name is looked up via Vault's identity API at reconciliation time.
	// The resolved IDs are merged with any explicitly provided MemberEntityIDs.
	// +kubebuilder:validation:Optional
	// +listType=set
	// kubebuilder:validation:UniqueItems=true
	MemberEntityNames []string `json:"memberEntityNames,omitempty"`

	// resolvedMemberGroupIDs holds the IDs resolved from MemberGroupNames during PrepareInternalValues.
	resolvedMemberGroupIDs []string `json:"-"`

	// resolvedMemberEntityIDs holds the IDs resolved from MemberEntityNames during PrepareInternalValues.
	resolvedMemberEntityIDs []string `json:"-"`
}

// GroupStatus defines the observed state of Group
type GroupStatus struct {
	// ID is the Vault-assigned unique identifier for this identity group.
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

// Group is the Schema for the groups API
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupSpec   `json:"spec,omitempty"`
	Status GroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GroupList contains a list of Group
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Group{}, &GroupList{})
}

var _ vaultutils.VaultObject = &Group{}
var _ vaultutils.ConditionsAware = &Group{}
var _ vaultutils.VaultStatusEnricher = &Group{}

func (m *Group) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (d *Group) IsDeletable() bool {
	return true
}

func (m *Group) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *Group) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *Group) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string("/identity/group/name/" + d.Spec.Name))
	}
	return vaultutils.CleansePath(string("/identity/group/name/" + d.Name))
}

func (d *Group) GetPayload() map[string]any {
	return d.Spec.toMap()
}

func (i *GroupSpec) toMap() map[string]any {
	payload := map[string]any{}
	payload["type"] = i.Type
	payload["metadata"] = i.Metadata
	payload["policies"] = i.Policies
	if i.Type == "internal" {
		payload["member_group_ids"] = mergeUniqueStrings(i.MemberGroupIDs, i.resolvedMemberGroupIDs)
		payload["member_entity_ids"] = mergeUniqueStrings(i.MemberEntityIDs, i.resolvedMemberEntityIDs)
	}
	return payload
}

// mergeUniqueStrings returns a deduplicated union of two string slices.
// Returns nil when both inputs are empty/nil to preserve clean payloads.
func mergeUniqueStrings(a, b []string) []string {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(a)+len(b))
	result := make([]string, 0, len(a)+len(b))
	for _, s := range a {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	for _, s := range b {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

func (d *Group) IsInitialized() bool {
	return true
}

func (d *Group) PrepareInternalValues(context context.Context, object client.Object) error {
	log := log.FromContext(context)

	if d.Spec.Type != "internal" {
		return nil
	}

	resolvedGroupIDs, err := resolveIdentityNames(context, "group", d.Spec.MemberGroupNames)
	if err != nil {
		log.Error(err, "unable to resolve member group names")
		return err
	}
	d.Spec.resolvedMemberGroupIDs = resolvedGroupIDs

	resolvedEntityIDs, err := resolveIdentityNames(context, "entity", d.Spec.MemberEntityNames)
	if err != nil {
		log.Error(err, "unable to resolve member entity names")
		return err
	}
	d.Spec.resolvedMemberEntityIDs = resolvedEntityIDs

	return nil
}

// resolveIdentityNames looks up each name via Vault's identity API and returns
// the corresponding IDs. identityType should be "group" or "entity".
func resolveIdentityNames(ctx context.Context, identityType string, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	log := log.FromContext(ctx)
	ids := make([]string, 0, len(names))
	for _, name := range names {
		path := vaultutils.CleansePath("/identity/" + identityType + "/name/" + name)
		secret, found, err := vaultutils.ReadSecret(ctx, path)
		if err != nil {
			log.Error(err, "unable to retrieve identity object", "type", identityType, "name", name)
			return nil, err
		}
		if !found {
			err = errors.New(identityType + " not found: " + name)
			log.Error(err, identityType+" not found", "name", name)
			return nil, err
		}
		id, ok := secret.Data["id"].(string)
		if !ok {
			err = errors.New(identityType + " has no id field: " + name)
			log.Error(err, identityType+" missing id", "name", name)
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (d *Group) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *Group) IsValid() (bool, error) {
	return true, nil
}

func (d *Group) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (d *Group) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.toMap()
	return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}

// EnrichStatus reads the group back from Vault and persists the Vault-assigned ID in status.
func (d *Group) EnrichStatus(ctx context.Context) error {
	vaultClient := vaultutils.VaultClientFromContext(ctx)
	secret, err := vaultClient.Logical().ReadWithContext(ctx, d.GetPath())
	if err != nil {
		return err
	}
	if secret != nil && secret.Data != nil {
		if id, ok := secret.Data["id"].(string); ok {
			d.Status.ID = id
		}
	}
	return nil
}
