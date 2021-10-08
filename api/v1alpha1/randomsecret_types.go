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
	"errors"
	"math/rand"
	"time"

	"github.com/hashicorp/hcl/v2/hclsimple"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/scylladb/go-set/u8set"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RandomSecretSpec defines the desired state of RandomSecret
type RandomSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Authentication is the kube aoth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to create the secret.
	// The final path will be {[spec.authentication.namespace]}/{spec.path}/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path Path `json:"path,omitempty"`

	// SecretFormat specifies a map of key and password policies used to generate random values
	// +kubebuilder:validation:Required
	SecretFormat PasswordPolicy `json:"secretFormat,omitempty"`

	// RefreshPeriod if specified, the operator will refresh the secret with the given frequency
	// +kubebuilder:validation:Optional
	RefreshPeriod metav1.Duration `json:"refreshPeriod,omitempty"`

	// SecretKey is the key to be used for this secret when stored in Vault kv
	// +kubebuilder:validation:Required
	SecretKey string `json:"secretKey,omitempty"`

	calculatedSecret string `json:"-"`
}

var _ vaultutils.VaultObject = &RandomSecret{}

func (d *RandomSecret) GetPath() string {
	return string(d.Spec.Path) + "/" + d.Name
}
func (d *RandomSecret) GetPayload() map[string]interface{} {
	return map[string]interface{}{
		d.Spec.SecretKey: d.Spec.calculatedSecret,
	}
}
func (d *RandomSecret) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	return false
}

type PasswordPolicy struct {
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
	LastVaultSecretUpdate metav1.Time `json:"lastVaultSecretUpdate,omitempty"`
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

func (d *RandomSecret) GenerateNewPassword(vaultEdnpoint *vaultutils.VaultEndpoint) error {
	if d.Spec.SecretFormat.InlinePasswordPolicy != "" {
		policy := &PasswordPolicyFormat{}
		err := hclsimple.Decode(d.Spec.SecretKey, []byte(d.Spec.SecretFormat.InlinePasswordPolicy), nil, policy)
		if err != nil {
			return err
		}
		found := d.calculateSecret(policy, 10000)
		if !found {
			return errors.New("passowrd could not be generated, will retry...")
		} else {
			return nil
		}
	}
	if d.Spec.SecretFormat.PasswordPolicyName != "" {
		response, err := vaultEdnpoint.GetVaultClient().Logical().Read("/sys/policies/password/" + d.Spec.SecretFormat.PasswordPolicyName + "/generate")
		if err != nil {
			return err
		} else {
			d.Spec.calculatedSecret = response.Data["secret"].(string)
			return nil
		}
	}
	return errors.New("no passowrd polic methos specified")
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
