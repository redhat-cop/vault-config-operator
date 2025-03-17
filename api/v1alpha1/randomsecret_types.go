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
	"math/rand"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hclsimple"
	vault "github.com/hashicorp/vault/api"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/scylladb/go-set/u8set"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RandomSecretSpec defines the desired state of RandomSecret
type RandomSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuration to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the secret.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/{metadata.name}.
	// If IsKVSecretsEngineV2 is false, the authentication role must have the following capabilities = [ "create", "update", "delete"] on the {[spec.authentication.namespace]}/{spec.path}/{metadata.name} path.
	// If IsKVSecretsEngineV2 is true, the authentication role must have the following capabilities = [ "create", "update"] on the {[spec.authentication.namespace]}/{spec.path}/data/{metadata.name} path and capabilities = [ "delete"] on the {[spec.authentication.namespace]}/{spec.path}/metadata/{metadata.name} path.
	// Additionally, if IsKVSecretsEngineV2 is true, it is acceptable for this value to have a suffix of "/data" or not. This suffix is no longer needed but still supported for backwards compatibility.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// SecretFormat specifies a map of key and password policies used to generate random values
	// +kubebuilder:validation:Required
	SecretFormat VaultPasswordPolicy `json:"secretFormat,omitempty"`

	// RefreshPeriod if specified, the operator will refresh the secret with the given frequency. This will also set the ttl of the secret which provides a hint for how often consumers should check back for a new value when reading the secret's lease_duration.
	// +kubebuilder:validation:Optional
	RefreshPeriod *metav1.Duration `json:"refreshPeriod,omitempty"`

	// SecretKey is the key to be used for this secret when stored in Vault kv
	// +kubebuilder:validation:Required
	SecretKey string `json:"secretKey,omitempty"`

	calculatedSecret string `json:"-"`

	// IsKVSecretsEngineV2 indicates if the KV Secrets engine is V2 or not. Default is false to indicate the payload to send is for KV Secret Engine V1.
	// +kubebuilder:validation:Required
	// +kubebuilder:default=false
	IsKVSecretsEngineV2 bool `json:"isKVSecretsEngineV2,omitempty"`

	// The name of the obejct created in Vault. If this is specified it takes precedence over {metatada.name}
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
	Name string `json:"name,omitempty"`
}

const ttlKey string = "ttl"

var _ vaultutils.VaultObject = &RandomSecret{}
var _ vaultutils.ConditionsAware = &RandomSecret{}

func (d *RandomSecret) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (d *RandomSecret) GetPath() string {
	if d.Spec.Name != "" {
		return vaultutils.CleansePath(string(d.Spec.Path) + "/" + d.Spec.Name)
	}
	return vaultutils.CleansePath(string(d.Spec.Path) + "/" + d.Name)
}

func (d *RandomSecret) IsDeletable() bool {
	return true
}

func (d *RandomSecret) getV1Payload() map[string]interface{} {

	payload := map[string]interface{}{
		d.Spec.SecretKey: d.Spec.calculatedSecret,
	}

	if d.Spec.RefreshPeriod != nil && d.Spec.RefreshPeriod.Duration > 0 {
		payload[ttlKey] = d.Spec.RefreshPeriod.Duration.String()
	}

	return payload
}

func (d *RandomSecret) IsKVSecretsEngineV2() bool {
	return d.Spec.IsKVSecretsEngineV2
}

func (d *RandomSecret) GetPayload() map[string]interface{} {
	if d.IsKVSecretsEngineV2() {
		return map[string]interface{}{
			"data": d.getV1Payload(),
		}
	}
	return d.getV1Payload()
}

func (d *RandomSecret) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	return false
}

func (d *RandomSecret) IsInitialized() bool {
	return true
}

func (d *RandomSecret) PrepareInternalValues(context context.Context, object client.Object) error {
	return d.GenerateNewPassword(context)
}

func (d *RandomSecret) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

type VaultPasswordPolicy struct {
	// PasswordPolicyName a ref to a password policy defined in Vault. Notice that in order to use this, the Vault role you use needs the following capabilities = ["read"] on /sys/policy/password.
	// Only one of PasswordPolicyName or InlinePasswordPolicy can be specified
	// +kubebuilder:validation:Optional
	PasswordPolicyName string `json:"passwordPolicyName,omitempty"`

	// InlinePasswordPolicy is an inline password policy specified using Vault password policy syntax (https://www.vaultproject.io/docs/concepts/password-policies#password-policy-syntax)
	// Only one of PasswordPolicyName or InlinePasswordPolicy can be specified
	// +kubebuilder:validation:Optional
	InlinePasswordPolicy string `json:"inlinePasswordPolicy,omitempty"`
}

// RandomSecretStatus defines the observed state of RandomSecret
type RandomSecretStatus struct {

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	//LastVaultSecretUpdate last time when this secret was updated in Vault
	LastVaultSecretUpdate *metav1.Time `json:"lastVaultSecretUpdate,omitempty"`
}

func (m *RandomSecret) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *RandomSecret) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RandomSecret is the Schema for the randomsecrets API
type RandomSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RandomSecretSpec   `json:"spec,omitempty"`
	Status RandomSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RandomSecretList contains a list of RandomSecret
type RandomSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RandomSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RandomSecret{}, &RandomSecretList{})
}

type PasswordPolicyFormat struct {
	Length int                  `hcl:"length"`
	Rules  []PasswordPolicyRule `hcl:"rule,block"`
}

type PasswordPolicyRule struct {
	RuleType string `hcl:"type,label"`
	Charset  string `hcl:"charset"`
	MinChars int    `hcl:"min-chars"`
}

func (d *RandomSecret) GenerateNewPassword(context context.Context) error {
	if d.Spec.SecretFormat.InlinePasswordPolicy != "" {
		policy := &PasswordPolicyFormat{}
		err := hclsimple.Decode(d.Spec.SecretKey, []byte(d.Spec.SecretFormat.InlinePasswordPolicy), nil, policy)
		if err != nil {
			return err
		}
		found := d.calculateSecret(policy, 10000)
		if !found {
			return errors.New("password could not be generated, will retry")
		} else {
			return nil
		}
	}
	if d.Spec.SecretFormat.PasswordPolicyName != "" {
		vaultClient := context.Value("vaultClient").(*vault.Client)
		response, err := vaultClient.Logical().Read("/sys/policies/password/" + d.Spec.SecretFormat.PasswordPolicyName + "/generate")
		if err != nil {
			return err
		} else {
			if response == nil || response.Data == nil {
				return errors.New("no data returned by password policy")
			}
			if password, ok := response.Data["password"]; ok {
				d.Spec.calculatedSecret = password.(string)
				return nil
			} else {
				return errors.New("password policy did not generate a password")
			}
		}
	}
	return errors.New("no password policy method specified")
}

func (d *RandomSecret) calculateSecret(policy *PasswordPolicyFormat, attempts int) bool {

	filteredPasswordPolicyRules := []PasswordPolicyRule{}
	for i := range policy.Rules {
		if policy.Rules[i].RuleType == "charset" {
			filteredPasswordPolicyRules = append(filteredPasswordPolicyRules, policy.Rules[i])
		}
	}
	// let's build the array of runes needed by this random password
	intSet := u8set.New()
	charSetToRule := map[*u8set.Set]PasswordPolicyRule{}
	for i := range filteredPasswordPolicyRules {
		charset := u8set.New([]byte(filteredPasswordPolicyRules[i].Charset)...)
		intSet.Merge(charset)
		charSetToRule[charset] = filteredPasswordPolicyRules[i]
	}

	var valid bool = true
	var randomString string
	for attempt := 0; attempt < attempts; attempt++ {
		randomString = randStringBytes(policy.Length, intSet.List())
		// now we need to check if the new string complies with the requirements
		for charset, rule := range charSetToRule {
			counter := 0
			for i := range randomString {
				if charset.Has(randomString[i]) {
					counter++
				}
			}
			if counter < rule.MinChars {
				valid = false
				break
			}
		}
		if valid {
			break
		}
	}
	if valid {
		d.Spec.calculatedSecret = randomString
	}
	return valid
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randStringBytes(n int, letterUints []uint8) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterUints[rand.Intn(len(letterUints))]
	}
	return string(b)
}

func (r *RandomSecret) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *RandomSecret) isValid() error {
	result := &multierror.Error{}
	result = multierror.Append(result, r.validateEitherPasswordPolicyReferenceOrInline())
	result = multierror.Append(result, r.validateInlinePasswordPolicyFormat())
	result = multierror.Append(result, r.validateSecretKey())
	result = multierror.Append(result, r.validateKVv2DataInPath())
	return result.ErrorOrNil()
}

func (r *RandomSecret) validateEitherPasswordPolicyReferenceOrInline() error {
	count := 0
	if r.Spec.SecretFormat.InlinePasswordPolicy != "" {
		count++
	}
	if r.Spec.SecretFormat.PasswordPolicyName != "" {
		count++
	}
	if count != 1 {
		return errors.New("only one of InlinePasswordPolicy or passwordPolicyName can be defined")
	}
	return nil
}

func (r *RandomSecret) validateInlinePasswordPolicyFormat() error {
	if r.Spec.SecretFormat.InlinePasswordPolicy != "" {
		passwordPolicyFormat := &PasswordPolicyFormat{}
		if strings.HasSuffix(r.Spec.SecretKey, ".hcl") {
			return hclsimple.Decode(r.Spec.SecretKey, []byte(r.Spec.SecretFormat.InlinePasswordPolicy), nil, passwordPolicyFormat)
		} else {
			return hclsimple.Decode(r.Spec.SecretKey+".hcl", []byte(r.Spec.SecretFormat.InlinePasswordPolicy), nil, passwordPolicyFormat)
		}
	}
	return nil
}

func (r *RandomSecret) validateSecretKey() error {
	if r.Spec.RefreshPeriod != nil && r.Spec.RefreshPeriod.Duration > 0 && r.Spec.SecretKey == ttlKey {
		return fmt.Errorf("secretKey must not be %v since this is a protected key when RefreshPeriod is set", ttlKey)
	}
	return nil
}

func (d *RandomSecret) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &d.Spec.Authentication
}

func (r *RandomSecret) validateKVv2DataInPath() error {
	if r.IsKVSecretsEngineV2() && !strings.Contains(r.GetPath(), "/data/") {
		return errors.New("KVv2 secrets must have /data defined in the path, for example /secret-mount-path/data/path")
	}
	return nil
}
