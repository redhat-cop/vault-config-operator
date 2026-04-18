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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AWSSecretEngineConfigSpec defines the desired state of AWSSecretEngineConfig
type AWSSecretEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	// Authentication is the kube auth configuraiton to be used to execute this request
	// +kubebuilder:validation:Required
	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config/root.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// AWSCredentials consists of access_key and secret_key, which can be created as Kubernetes Secret, VaultSecret or RandomSecret
	// +kubebuilder:validation:Optional
	AWSCredentials vaultutils.RootCredentialConfig `json:"awsCredentials,omitempty"`

	// +kubebuilder:validation:Required
	AWSSEConfig `json:",inline"`
}

// AWSSecretEngineConfigStatus defines the observed state of AWSSecretEngineConfig
type AWSSecretEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AWSSecretEngineConfig is the Schema for the awssecretengineconfigs API
type AWSSecretEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSSecretEngineConfigSpec   `json:"spec,omitempty"`
	Status AWSSecretEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AWSSecretEngineConfigList contains a list of AWSSecretEngineConfig
type AWSSecretEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSSecretEngineConfig `json:"items"`
}

type AWSSEConfig struct {
	// Number of max retries the client should use for recoverable errors. The default (-1) falls back to the AWS SDK's default behavior.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=-1
	MaxRetries int `json:"maxRetries,omitempty"`

	// Role ARN to assume for plugin workload identity federation (Enterprise). Required with identity_token_audience.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	RoleARN string `json:"roleArn,omitempty"`

	// The audience claim value for plugin identity tokens (Enterprise). Must match an allowed audience configured for the target IAM OIDC identity provider.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	IdentityTokenAudience string `json:"identityTokenAudience,omitempty"`

	// The TTL of generated tokens (Enterprise). Defaults to 1 hour. Uses duration format strings.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="3600"
	IdentityTokenTTL string `json:"identityTokenTtl,omitempty"`

	// Specifies the AWS region. If not set it will use the AWS_REGION env var, AWS_DEFAULT_REGION env var, or us-east-1 in that order.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	Region string `json:"region,omitempty"`

	// Specifies a custom HTTP IAM endpoint to use.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	IAMEndpoint string `json:"iamEndpoint,omitempty"`

	// Specifies a custom HTTP STS endpoint to use.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	STSEndpoint string `json:"stsEndpoint,omitempty"`

	// Specifies a custom STS region to use (should match sts_endpoint).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	STSRegion string `json:"stsRegion,omitempty"`

	// Specifies an ordered list of fallback STS endpoints to use.
	// +kubebuilder:validation:Optional
	STSFallbackEndpoints []string `json:"stsFallbackEndpoints,omitempty"`

	// Specifies an ordered list of fallback STS regions to use (should match fallback endpoints).
	// +kubebuilder:validation:Optional
	STSFallbackRegions []string `json:"stsFallbackRegions,omitempty"`

	// Template describing how dynamic usernames are generated.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	UsernameTemplate string `json:"usernameTemplate,omitempty"`

	// The amount of time, in seconds, Vault should wait before rotating the root credential (Enterprise). A zero value tells Vault not to rotate the root credential.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	RotationPeriod int `json:"rotationPeriod,omitempty"`

	// The schedule, in cron-style time format, defining the schedule on which Vault should rotate the root token (Enterprise).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	RotationSchedule string `json:"rotationSchedule,omitempty"`

	// The maximum amount of time, in seconds, allowed to complete a rotation when a scheduled token rotation occurs (Enterprise).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=0
	RotationWindow int `json:"rotationWindow,omitempty"`

	// Cancels all upcoming rotations of the root credential until unset (Enterprise).
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	DisableAutomatedRotation bool `json:"disableAutomatedRotation,omitempty"`

	retrievedAccessKey string `json:"-"`

	retrievedSecretKey string `json:"-"`
}

var _ vaultutils.VaultObject = &AWSSecretEngineConfig{}
var _ vaultutils.ConditionsAware = &AWSSecretEngineConfig{}

func init() {
	SchemeBuilder.Register(&AWSSecretEngineConfig{}, &AWSSecretEngineConfigList{})
}

func (d *AWSSecretEngineConfig) IsDeletable() bool {
	return true
}

func (r *AWSSecretEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (d *AWSSecretEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *AWSSecretEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *AWSSecretEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (d *AWSSecretEngineConfig) GetPath() string {
	return string(d.Spec.Path) + "/" + "config/root"
}

func (d *AWSSecretEngineConfig) GetPayload() map[string]interface{} {
	return d.Spec.toMap()
}

func (r *AWSSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.AWSSEConfig.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (r *AWSSecretEngineConfig) IsInitialized() bool {
	return true
}

func (r *AWSSecretEngineConfig) IsValid() (bool, error) {
	err := r.isValid()
	return err == nil, err
}

func (r *AWSSecretEngineConfig) isValid() error {
	return r.Spec.AWSCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}

func (r *AWSSecretEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
	if reflect.DeepEqual(r.Spec.AWSCredentials, vaultutils.RootCredentialConfig{PasswordKey: "secret_key", UsernameKey: "access_key"}) {
		return nil
	}

	return r.setInternalCredentials(context)
}

func (d *AWSSecretEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *AWSSecretEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
	if r.Spec.AWSCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.AWSCredentials.RandomSecret.Name,
		}, randomSecret)
		if err != nil {
			log.Error(err, "unable to retrieve RandomSecret", "instance", r)
			return err
		}
		secret, exists, err := vaultutils.ReadSecret(context, randomSecret.GetPath())
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		accessKey := secret.Data[r.Spec.AWSCredentials.UsernameKey].(string)
		secretKey := secret.Data[r.Spec.AWSCredentials.PasswordKey].(string)
		r.setAccessKeyAndSecretKey(accessKey+":"+secretKey, accessKey, secretKey)
		return nil
	}
	if r.Spec.AWSCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.AWSCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		accessKey, ok := secret.Data[r.Spec.AWSCredentials.UsernameKey]
		if !ok {
			err := errors.New("unable to find access_key key in secret")
			log.Error(err, "unable to retrieve field for Secret", "instance", r)
			return err
		}
		secretKey, ok := secret.Data[r.Spec.AWSCredentials.PasswordKey]
		if !ok {
			err := errors.New("unable to find secret_key key in secret")
			log.Error(err, "unable to retrieve field for Secret", "instance", r)
			return err
		}
		r.setAccessKeyAndSecretKey(string(accessKey)+":"+string(secretKey), string(accessKey), string(secretKey))
		return nil
	}
	if r.Spec.AWSCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.AWSCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		accessKey := secret.Data[r.Spec.AWSCredentials.UsernameKey].(string)
		secretKey := secret.Data[r.Spec.AWSCredentials.PasswordKey].(string)
		r.setAccessKeyAndSecretKey(accessKey+":"+secretKey, accessKey, secretKey)
		log.V(1).Info("", "accessKey", accessKey, "secretKey", secretKey)
		return nil
	}
	return errors.New("no aws credentials source specified")
}

func (r *AWSSecretEngineConfig) setAccessKeyAndSecretKey(secretName string, accessKey string, secretKey string) {
	r.Spec.AWSSEConfig.retrievedAccessKey = accessKey
	r.Spec.AWSSEConfig.retrievedSecretKey = secretKey
}

func (i *AWSSEConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}

	if i.retrievedAccessKey != "" {
		payload["access_key"] = i.retrievedAccessKey
	}
	if i.retrievedSecretKey != "" {
		payload["secret_key"] = i.retrievedSecretKey
	}

	payload["max_retries"] = i.MaxRetries

	if i.Region != "" {
		payload["region"] = i.Region
	}
	if i.IAMEndpoint != "" {
		payload["iam_endpoint"] = i.IAMEndpoint
	}
	if i.STSEndpoint != "" {
		payload["sts_endpoint"] = i.STSEndpoint
	}
	if i.STSRegion != "" {
		payload["sts_region"] = i.STSRegion
	}
	if len(i.STSFallbackEndpoints) > 0 {
		payload["sts_fallback_endpoints"] = i.STSFallbackEndpoints
	}
	if len(i.STSFallbackRegions) > 0 {
		payload["sts_fallback_regions"] = i.STSFallbackRegions
	}
	if i.UsernameTemplate != "" {
		payload["username_template"] = i.UsernameTemplate
	}
	if i.RoleARN != "" {
		payload["role_arn"] = i.RoleARN
	}
	if i.IdentityTokenAudience != "" {
		payload["identity_token_audience"] = i.IdentityTokenAudience
	}
	if i.IdentityTokenTTL != "" && i.IdentityTokenTTL != "3600" {
		payload["identity_token_ttl"] = i.IdentityTokenTTL
	}
	if i.RotationPeriod != 0 {
		payload["rotation_period"] = i.RotationPeriod
	}
	if i.RotationSchedule != "" {
		payload["rotation_schedule"] = i.RotationSchedule
	}
	if i.RotationWindow != 0 {
		payload["rotation_window"] = i.RotationWindow
	}
	if i.DisableAutomatedRotation {
		payload["disable_automated_rotation"] = i.DisableAutomatedRotation
	}

	return payload
}

func (r *AWSSecretEngineConfigSpec) toMap() map[string]interface{} {
	return r.AWSSEConfig.toMap()
}
