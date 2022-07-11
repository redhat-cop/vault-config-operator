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
	"strings"

	vault "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:validation:Pattern:=`^(?:/?[\w;:@&=\$-\.\+]*)+/?`
type Path string

type KubeAuthConfiguration struct {
	// ServiceAccount is the service account used for the kube auth authentication
	// +kubebuilder:validation:Required
	// +kubebuilder:default={"name": "default"}
	ServiceAccount *corev1.LocalObjectReference `json:"serviceAccount,omitempty"`

	// Path is the path of the role used for this kube auth authentication. The operator will try to authenticate at {[namespace/]}auth/{spec.path}
	// +kubebuilder:validation:Required
	// +kubebuilder:default=kubernetes
	Path Path `json:"path,omitempty"`

	// Role the role to be used during authentication
	// +kubebuilder:validation:Required
	Role string `json:"role,omitempty"`

	//Namespace is the Vault namespace to be used in all the operations withing this connection/authentication. Only available in Vault Enterprise.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

func (kc *KubeAuthConfiguration) GetNamespace() string {
	return kc.Namespace
}
func (kc *KubeAuthConfiguration) GetRole() string {
	return kc.Role
}
func (kc *KubeAuthConfiguration) GetKubeAuthPath() string {
	return cleansePath("auth/" + string(kc.Path) + "/login")
}

func (kc *KubeAuthConfiguration) GetServiceAccountName() string {
	return kc.ServiceAccount.Name
}

func (kc *KubeAuthConfiguration) GetVaultClient(context context.Context, kubeNamespace string) (*vault.Client, error) {
	log := log.FromContext(context)
	jwt, err := kc.getJWTToken(context, kubeNamespace)
	if err != nil {
		log.Error(err, "unable to retrieve jwt token for", "namespace", kubeNamespace, "serviceaccount", kc.GetServiceAccountName())
		return nil, err
	}
	vaultClient, err := kc.createVaultClient(context, jwt)
	if err != nil {
		log.Error(err, "unable to create vault client")
		return nil, err
	}
	return vaultClient, nil
}

func getJWTToken(context context.Context, serviceAccountName string, kubeNamespace string) (string, error) {
	log := log.FromContext(context)
	kubeClient := context.Value("kubeClient").(client.Client)
	serviceAccount := &corev1.ServiceAccount{}
	err := kubeClient.Get(context, client.ObjectKey{
		Namespace: kubeNamespace,
		Name:      serviceAccountName,
	}, serviceAccount)
	if err != nil {
		log.Error(err, "unable to retrieve", "service account", client.ObjectKey{
			Namespace: kubeNamespace,
			Name:      serviceAccountName,
		})
		return "", err
	}
	var tokenSecretName string
	for _, secretName := range serviceAccount.Secrets {
		if strings.Contains(secretName.Name, "token") {
			tokenSecretName = secretName.Name
			break
		}
	}
	if tokenSecretName == "" {
		return "", errors.New("unable to find token secret name for service account" + kubeNamespace + "/" + serviceAccountName)
	}
	secret := &corev1.Secret{}
	err = kubeClient.Get(context, client.ObjectKey{
		Namespace: kubeNamespace,
		Name:      tokenSecretName,
	}, secret)
	if err != nil {
		log.Error(err, "unable to retrieve", "secret", client.ObjectKey{
			Namespace: kubeNamespace,
			Name:      tokenSecretName,
		})
		return "", err
	}
	if jwt, ok := secret.Data["token"]; ok {
		return string(jwt), nil
	} else {
		return "", errors.New("unable to find \"token\" key in secret" + kubeNamespace + "/" + tokenSecretName)
	}
}

func (kc *KubeAuthConfiguration) getJWTToken(context context.Context, kubeNamespace string) (string, error) {
	return getJWTToken(context, kc.GetServiceAccountName(), kubeNamespace)
}

func (kc *KubeAuthConfiguration) createVaultClient(context context.Context, jwt string) (*vault.Client, error) {
	log := log.FromContext(context)
	config := vault.DefaultConfig()
	client, err := vault.NewClient(config)
	if err != nil {
		log.Error(err, "unable initialize vault client")
		return nil, err
	}
	if kc.GetNamespace() != "" {
		client.SetNamespace(kc.GetNamespace())
	}
	secret, err := client.Logical().Write(kc.GetKubeAuthPath(), map[string]interface{}{
		"jwt":  jwt,
		"role": kc.GetRole(),
	})

	if err != nil {
		log.Error(err, "unable to login to vault")
		return nil, err
	}

	client.SetToken(secret.Auth.ClientToken)

	return client, nil
}

func cleansePath(path string) string {
	return strings.Trim(strings.ReplaceAll(path, "//", "/"), "/")
}

func toString(name interface{}) string {
	if name != nil {
		return name.(string)
	}
	return ""
}

// func parseOrDie(val string) metav1.Duration {
// 	d, err := time.ParseDuration(val)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return metav1.Duration{
// 		Duration: d,
// 	}
// }

type VaultSecretReference struct {
	// Path is the path to the secret
	// +kubebuilder:validation:Required
	Path string `json:"path,omitempty"`
}

type RootCredentialConfig struct {
	// VaultSecret retrieves the credentials from a Vault secret. This will map the "username" and "password" keys of the secret to the username and password of this config. All other keys will be ignored. Only one of RootCredentialsFromVaultSecret or RootCredentialsFromSecret or RootCredentialsFromRandomSecret can be specified.
	// username: Specifies the name of the user to use as the "root" user when connecting to the database. This "root" user is used to create/update/delete users managed by these plugins, so you will need to ensure that this user has permissions to manipulate users appropriate to the database. This is typically used in the connection_url field via the templating directive "{{"username"}}" or "{{"name"}}".
	// password: Specifies the password to use when connecting with the username. This value will not be returned by Vault when performing a read upon the configuration. This is typically used in the connection_url field via the templating directive "{{"password"}}".
	// If username is provided as spec.username, it takes precedence over the username retrieved from the referenced secret
	// +kubebuilder:validation:Optional
	VaultSecret *VaultSecretReference `json:"vaultSecret,omitempty"`

	// Secret retrieves the credentials from a Kubernetes secret. The secret must be of basicauth type (https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). This will map the "username" and "password" keys of the secret to the username and password of this config. If the kubernetes secret is updated, this configuration will also be updated. All other keys will be ignored. Only one of RootCredentialsFromVaultSecret or RootCredentialsFromSecret or RootCredentialsFromRandomSecret can be specified.
	// username: Specifies the name of the user to use as the "root" user when connecting to the database. This "root" user is used to create/update/delete users managed by these plugins, so you will need to ensure that this user has permissions to manipulate users appropriate to the database. This is typically used in the connection_url field via the templating directive "{{"username"}}" or "{{"name"}}".
	// password: Specifies the password to use when connecting with the username. This value will not be returned by Vault when performing a read upon the configuration. This is typically used in the connection_url field via the templating directive "{{"password"}}".
	// If username is provided as spec.username, it takes precedence over the username retrieved from the referenced secret
	// +kubebuilder:validation:Optional
	Secret *corev1.LocalObjectReference `json:"secret,omitempty"`

	// RandomSecret retrieves the credentials from the Vault secret corresponding to this RandomSecret. This will map the "username" and "password" keys of the secret to the username and password of this config. All other keys will be ignored. If the RandomSecret is refreshed the operator retrieves the new secret from Vault and updates this configuration. Only one of RootCredentialsFromVaultSecret or RootCredentialsFromSecret or RootCredentialsFromRandomSecret can be specified.
	// When using randomSecret a username must be specified in the spec.username
	// password: Specifies the password to use when connecting with the username. This value will not be returned by Vault when performing a read upon the configuration. This is typically used in the connection_url field via the templating directive "{{"password"}}"".
	// +kubebuilder:validation:Optional
	RandomSecret *corev1.LocalObjectReference `json:"randomSecret,omitempty"`

	// PasswordKey key to be used when retrieving the password, required with VaultSecrets and Kubernetes secrets, ignored with RandomSecret
	// +kubebuilder:validation:Optional
	PasswordKey string `json:"passwordKey,omitempty"`

	// UsernameKey key to be used when retrieving the username, optional with VaultSecrets and Kubernetes secrets, ignored with RandomSecret
	// +kubebuilder:validation:Optional
	UsernameKey string `json:"usernameKey,omitempty"`
}

func (credentials *RootCredentialConfig) validateEitherFromVaultSecretOrFromSecret() error {
	count := 0
	if credentials.Secret != nil {
		count++
	}
	if credentials.VaultSecret != nil {
		count++
	}
	if count != 1 {
		return errors.New("only one of spec.rootCredentials.vaultSecret or spec.rootCredentials.secret or spec.rootCredentials.randomSecret can be specified")
	}
	return nil
}

func (credentials *RootCredentialConfig) validateEitherFromVaultSecretOrFromSecretOrFromRandomSecret() error {
	count := 0
	if credentials.RandomSecret != nil {
		count++
	}
	if credentials.Secret != nil {
		count++
	}
	if credentials.VaultSecret != nil {
		count++
	}
	if count != 1 {
		return errors.New("only one of spec.rootCredentials.vaultSecret or spec.rootCredentials.secret or spec.rootCredentials.randomSecret can be specified")
	}
	return nil
}

func GetFinalizer(instance client.Object) string {
	return "controller-" + strings.ToLower(instance.GetObjectKind().GroupVersionKind().Kind)
}

type TargetNamespaceConfig struct {
	// TargetNamespaceSelector is a selector of namespaces from which service accounts will receove this role. Either TargetNamespaceSelector or TargetNamespaces can be specified
	// +kubebuilder:validation:Optional
	TargetNamespaceSelector *metav1.LabelSelector `json:"targetNamespaceSelector,omitempty"`

	// TargetNamespaces is a list of namespace from which service accounts will receive this role. Either TargetNamespaceSelector or TargetNamespaces can be specified.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MinItems=1
	// kubebuilder:validation:UniqueItems=true
	// +listType=set
	TargetNamespaces []string `json:"targetNamespaces,omitempty"`
}
