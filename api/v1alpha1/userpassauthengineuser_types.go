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
	"fmt"
	"reflect"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// UserpassAuthEngineUserSpec defines the desired state of UserpassAuthEngineUser
type UserpassAuthEngineUserSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Optional
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{[spec.authentication.targetNamespace]}/auth/{spec.path}/users/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Optional
	Path vaultutils.Path `json:"path,omitempty"`

	UserpassUser `json:",inline"`

	// PasswordCredentials specifies where to retrieve the password for this user.
	// The password is resolved from a K8s Secret, VaultSecret, or RandomSecret.
	// Only the passwordKey is used (usernameKey is ignored — the username comes from metadata.name or spec.name).
	// +kubebuilder:validation:Required
	PasswordCredentials vaultutils.RootCredentialConfig `json:"passwordCredentials"`

	// The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9._]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

type UserpassUser struct {
	// TokenTTL is the incremental lifetime for generated tokens (e.g. "1h")
	// +kubebuilder:validation:Optional
	TokenTTL string `json:"tokenTTL,omitempty" vault:"token_ttl,duration,omitempty"`

	// TokenMaxTTL is the maximum lifetime for generated tokens (e.g. "24h")
	// +kubebuilder:validation:Optional
	TokenMaxTTL string `json:"tokenMaxTTL,omitempty" vault:"token_max_ttl,duration,omitempty"`

	// TokenPolicies is the list of token policies
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenPolicies []string `json:"tokenPolicies,omitempty" vault:"token_policies,omitempty"`

	// TokenBoundCIDRs is the list of CIDR blocks for token authentication
	// +kubebuilder:validation:Optional
	// +listType=set
	TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty" vault:"token_bound_cidrs,omitempty"`

	// TokenExplicitMaxTTL is the hard cap max TTL (e.g. "24h")
	// +kubebuilder:validation:Optional
	TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty" vault:"token_explicit_max_ttl,duration,omitempty"`

	// TokenNoDefaultPolicy excludes the default policy from generated tokens
	// +kubebuilder:validation:Optional
	TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty" vault:"token_no_default_policy,omitempty"`

	// TokenNumUses is the max number of token uses (0=unlimited)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	TokenNumUses int `json:"tokenNumUses,omitempty" vault:"token_num_uses,number,omitempty"`

	// TokenPeriod is the renewal period for periodic tokens (e.g. "24h")
	// +kubebuilder:validation:Optional
	TokenPeriod string `json:"tokenPeriod,omitempty" vault:"token_period,duration,omitempty"`

	// TokenType is the type of token to generate: service, batch, or default
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:={"service","batch","default","default-service","default-batch",""}
	TokenType string `json:"tokenType,omitempty" vault:"token_type,omitempty"`

	retrievedPassword string `json:"-"`
}

// UserpassAuthEngineUserStatus defines the observed state of UserpassAuthEngineUser
type UserpassAuthEngineUserStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// UserpassAuthEngineUser is the Schema for the userpassauthengineusers API
type UserpassAuthEngineUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserpassAuthEngineUserSpec   `json:"spec,omitempty"`
	Status UserpassAuthEngineUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// UserpassAuthEngineUserList contains a list of UserpassAuthEngineUser
type UserpassAuthEngineUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserpassAuthEngineUser `json:"items"`
}

var _ vaultutils.VaultObject = &UserpassAuthEngineUser{}
var _ vaultutils.ConditionsAware = &UserpassAuthEngineUser{}

func (d *UserpassAuthEngineUser) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *UserpassAuthEngineUser) IsDeletable() bool {
	return true
}

func (d *UserpassAuthEngineUser) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/users/" + vaultutils.ResourceName(d.Spec.Name, d.Name))
}

func (d *UserpassAuthEngineUser) GetPayload() map[string]any {
	return d.Spec.UserpassUser.toMap()
}

func (d *UserpassAuthEngineUser) IsEquivalentToDesiredState(payload map[string]any) bool {
	desiredState := d.Spec.UserpassUser.toMap()
	delete(desiredState, "password")
	removeUnsetFields(desiredState, payload)

	for _, boolKey := range []string{"token_no_default_policy"} {
		if boolVal, ok := desiredState[boolKey].(bool); ok && !boolVal {
			if _, inPayload := payload[boolKey]; !inPayload {
				delete(desiredState, boolKey)
			}
		}
	}
	filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
	setFields := []string{"token_policies", "token_bound_cidrs"}
	for _, key := range setFields {
		sortAnyStringSlice(desiredState, key)
		sortAnyStringSlice(filteredPayload, key)
	}
	return reflect.DeepEqual(desiredState, filteredPayload)
}

func (d *UserpassAuthEngineUser) PrepareInternalValues(ctx context.Context, object client.Object) error {
	return d.setInternalCredentials(ctx)
}

func (d *UserpassAuthEngineUser) IsValid() (bool, error) {
	err := d.isValid()
	return err == nil, err
}

func (d *UserpassAuthEngineUser) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (m *UserpassAuthEngineUser) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *UserpassAuthEngineUser) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

func (d *UserpassAuthEngineUser) setInternalCredentials(ctx context.Context) error {
	log := log.FromContext(ctx)
	kubeClient := vaultutils.KubeClientFromContext(ctx)
	if d.Spec.PasswordCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(ctx, types.NamespacedName{
			Namespace: d.Namespace,
			Name:      d.Spec.PasswordCredentials.RandomSecret.Name,
		}, randomSecret)
		if err != nil {
			log.Error(err, "unable to retrieve RandomSecret", "instance", d)
			return err
		}
		secret, exists, err := vaultutils.ReadSecret(ctx, randomSecret.GetPath())
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", d)
			return err
		}

		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		password, ok := secretData[randomSecret.Spec.SecretKey].(string)
		if !ok {
			return fmt.Errorf("secret key %q not found or not a string in vault secret", randomSecret.Spec.SecretKey)
		}
		d.Spec.UserpassUser.retrievedPassword = password
		return nil
	}
	if d.Spec.PasswordCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(ctx, types.NamespacedName{
			Namespace: d.Namespace,
			Name:      d.Spec.PasswordCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", d)
			return err
		}
		passwordBytes, ok := secret.Data[d.Spec.PasswordCredentials.PasswordKey]
		if !ok {
			return fmt.Errorf("K8s Secret %q missing key %q", d.Spec.PasswordCredentials.Secret.Name, d.Spec.PasswordCredentials.PasswordKey)
		}
		d.Spec.UserpassUser.retrievedPassword = string(passwordBytes)
		return nil
	}
	if d.Spec.PasswordCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(ctx, string(d.Spec.PasswordCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", d)
			return err
		}
		secretData := vaultutils.UnwrapKVv2Data(secret.Data)
		passwordVal, ok := secretData[d.Spec.PasswordCredentials.PasswordKey].(string)
		if !ok {
			return fmt.Errorf("VaultSecret key %q not found or not a string", d.Spec.PasswordCredentials.PasswordKey)
		}
		d.Spec.UserpassUser.retrievedPassword = passwordVal
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (d *UserpassAuthEngineUser) isValid() error {
	return d.Spec.PasswordCredentials.ValidateCredentialSource()
}

func (d *UserpassUser) toMap() map[string]any {
	payload := vaultutils.StructToVaultPayload(d)
	if d.retrievedPassword != "" {
		payload["password"] = d.retrievedPassword
	}
	return payload
}

func init() {
	SchemeBuilder.Register(&UserpassAuthEngineUser{}, &UserpassAuthEngineUserList{})
}
