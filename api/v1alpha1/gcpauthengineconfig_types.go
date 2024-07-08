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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GCPAuthEngineConfigSpec defines the desired state of GCPAuthEngineConfig
type GCPAuthEngineConfigSpec struct {
	// Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
	// +kubebuilder:validation:Optional
	Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

	Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

	// Path at which to make the configuration.
	// The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/config/{metadata.name}.
	// The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
	// +kubebuilder:validation:Required
	Path vaultutils.Path `json:"path,omitempty"`

	// +kubebuilder:validation:Required
	GCPConfig `json:",inline"`

	// GCPCredentials in JSON string containing the contents of a GCP service account credentials file.
	// +kubebuilder:validation:Optional
	GCPCredentials vaultutils.RootCredentialConfig `json:"GCPCredentials,omitempty"`
}

// GCPAuthEngineConfigStatus defines the observed state of GCPAuthEngineConfig
type GCPAuthEngineConfigStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GCPAuthEngineConfig is the Schema for the gcpauthengineconfigs API
type GCPAuthEngineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPAuthEngineConfigSpec   `json:"spec,omitempty"`
	Status GCPAuthEngineConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GCPAuthEngineConfigList contains a list of GCPAuthEngineConfig
type GCPAuthEngineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPAuthEngineConfig `json:"items"`
}

type GCPConfig struct {


	// Service Account Name. A service account is a special kind of account typically used by an application or compute workload, such as a Compute Engine instance, rather than a person. 
	// A service account is identified by its email address, which is unique to the account.
	// Applications use service accounts to make authorized API calls by authenticating as either the service account itself, or as Google Workspace or Cloud Identity users through domain-wide delegation. 
	// When an application authenticates as a service account, it has access to all resources that the service account has permission to access.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=""
	ServiceAccount string `json:"serviceAccount,omitempty"`	

	// Must be either unique_id or role_id.
	// If unique_id is specified, the service account's unique ID will be used for alias names during login.
	// If role_id is specified, the ID of the Vault role will be used. Only used if role type is iam.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="default"
	IAMalias string `json:"IAMalias,omitempty"`

	// The metadata to include on the token returned by the login endpoint. This metadata will be added to both audit logs, and on the iam_alias.
	// By default, it includes project_id, role, service_account_id, and service_account_email.
	// To include no metadata, set to "" via the CLI or [] via the API. To use only particular fields, select the explicit fields.
	// To restore to defaults, send only a field of default.
	// Only select fields that will have a low rate of change for your iam_alias because each change triggers a storage write and can have a performance impact at scale.
	// Only used if role type is iam.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="default"
	IAMmetadata string `json:"IAMmetadata,omitempty"`

	// Must be either instance_id or role_id. If instance_id is specified, the GCE instance ID will be used for alias names during login.
	// If role_id is specified, the ID of the Vault role will be used. Only used if role type is gce.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="role_id"
	GCEalias string `json:"GCEalias,omitempty"`

	// The metadata to include on the token returned by the login endpoint. This metadata will be added to both audit logs, and on the gce_alias.
	// By default, it includes instance_creation_timestamp, instance_id, instance_name, project_id, project_number, role, service_account_id, service_account_email, and zone.
	// To include no metadata, set to "" via the CLI or [] via the API. To use only particular fields, select the explicit fields. To restore to defaults, send only a field of default.
	// Only select fields that will have a low rate of change for your gce_alias because each change triggers a storage write and can have a performance impact at scale.
	// Only used if role type is gce.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="default"
	GCEmetadata string `json:"GCEmetadata,omitempty"`

	// Specifies overrides to service endpoints used when making API requests.
	// This allows specific requests made during authentication to target alternative service endpoints for use in Private Google Access environments.
	// Overrides are set at the subdomain level using the following keys:
	// api - Replaces the service endpoint used in API requests to https://www.googleapis.com.
	// iam - Replaces the service endpoint used in API requests to https://iam.googleapis.com.
	// crm - Replaces the service endpoint used in API requests to https://cloudresourcemanager.googleapis.com.
	// compute - Replaces the service endpoint used in API requests to https://compute.googleapis.com.
	// The endpoint value provided for a given key has the form of scheme://host:port. The scheme:// and :port portions of the endpoint value are optional.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default={}
	CustomEndpoint *apiextensionsv1.JSON `json:"customEndpoint,omitempty"`

	retrievedServiceAccount string `json:"-"`
	retrievedCredentials string `json:"-"`
}

var _ vaultutils.VaultObject = &GCPAuthEngineConfig{}
var _ vaultutils.ConditionsAware = &GCPAuthEngineConfig{}

func (d *GCPAuthEngineConfig) GetVaultConnection() *vaultutils.VaultConnection {
	return d.Spec.Connection
}

func (r *GCPAuthEngineConfig) GetConditions() []metav1.Condition {
	return r.Status.Conditions
}

func (r *GCPAuthEngineConfig) SetConditions(conditions []metav1.Condition) {
	r.Status.Conditions = conditions
}

func (r *GCPAuthEngineConfig) SetServiceAccountAndCredentials(ServiceAccount string, Credentials string) {
	r.Spec.GCPConfig.retrievedServiceAccount = ServiceAccount
	r.Spec.GCPConfig.retrievedCredentials = Credentials
}

func (r *GCPAuthEngineConfig) GetPath() string {
	return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config")
}

func (r *GCPAuthEngineConfig) GetPayload() map[string]interface{} {
	return r.Spec.GCPConfig.toMap()
}

func (r *GCPAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
	desiredState := r.Spec.GCPConfig.toMap()
	return reflect.DeepEqual(desiredState, payload)
}

func (r *GCPAuthEngineConfig) IsInitialized() bool {
	return true
}

func (r *GCPAuthEngineConfig) IsValid() (bool, error) {
	return true, nil
}

func (r *GCPAuthEngineConfig) GetKubeAuthConfiguration() *vaultutils.KubeAuthConfiguration {
	return &r.Spec.Authentication
}

func (r *GCPAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {

	if reflect.DeepEqual(r.Spec.GCPCredentials, vaultutils.RootCredentialConfig{UsernameKey: "serviceaccount", PasswordKey: "credentials"}) {
		return nil
	}

	return r.setInternalCredentials(context)
}

func (r *GCPAuthEngineConfig) PrepareTLSConfig(context context.Context, object client.Object) error {
	return nil
}

func (r *GCPAuthEngineConfig) setInternalCredentials(context context.Context) error {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
	if r.Spec.GCPCredentials.RandomSecret != nil {
		randomSecret := &RandomSecret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.GCPCredentials.RandomSecret.Name,
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
		r.SetServiceAccountAndCredentials(r.Spec.ServiceAccount, secret.Data[randomSecret.Spec.SecretKey].(string))
		return nil
	}
	if r.Spec.GCPCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := kubeClient.Get(context, types.NamespacedName{
			Namespace: r.Namespace,
			Name:      r.Spec.GCPCredentials.Secret.Name,
		}, secret)
		if err != nil {
			log.Error(err, "unable to retrieve Secret", "instance", r)
			return err
		}
		if r.Spec.ServiceAccount == "" {
			r.SetServiceAccountAndCredentials(string(secret.Data[r.Spec.GCPCredentials.UsernameKey]), string(secret.Data[r.Spec.GCPCredentials.PasswordKey]))
		} else {
			r.SetServiceAccountAndCredentials(r.Spec.GCPCredentials.UsernameKey, string(secret.Data[r.Spec.GCPCredentials.PasswordKey]))
		}
		return nil
	}
	if r.Spec.GCPCredentials.VaultSecret != nil {
		secret, exists, err := vaultutils.ReadSecret(context, string(r.Spec.GCPCredentials.VaultSecret.Path))
		if err != nil {
			return err
		}
		if !exists {
			err = errors.New("secret not found")
			log.Error(err, "unable to retrieve vault secret", "instance", r)
			return err
		}
		if r.Spec.ServiceAccount == "" {
			r.SetServiceAccountAndCredentials(secret.Data[r.Spec.GCPCredentials.UsernameKey].(string), secret.Data[r.Spec.GCPCredentials.PasswordKey].(string))
			log.V(1).Info("", "serviceaccount", secret.Data[r.Spec.GCPCredentials.UsernameKey].(string), "credentials", secret.Data[r.Spec.GCPCredentials.PasswordKey].(string))
		} else {
			r.SetServiceAccountAndCredentials(r.Spec.GCPConfig.ServiceAccount, secret.Data[r.Spec.GCPCredentials.PasswordKey].(string))
			log.V(1).Info("", "serviceaccount", r.Spec.GCPConfig.ServiceAccount, "credentials", secret.Data[r.Spec.GCPCredentials.PasswordKey].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func init() {
	SchemeBuilder.Register(&GCPAuthEngineConfig{}, &GCPAuthEngineConfigList{})
}

func (i *GCPConfig) toMap() map[string]interface{} {
	payload := map[string]interface{}{}
	payload["credentials"] = i.retrievedCredentials
	payload["iam_alias"] = i.IAMalias
	payload["iam_metadata"] = i.IAMmetadata
	payload["gce_alias"] = i.GCEalias
	payload["gce_metadata"] = i.GCEmetadata
	payload["custom_endpoint"] = i.CustomEndpoint

	return payload
}
