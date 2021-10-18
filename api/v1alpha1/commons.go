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
	"time"

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

func parseOrDie(val string) metav1.Duration {
	d, err := time.ParseDuration(val)
	if err != nil {
		panic(err)
	}
	return metav1.Duration{
		Duration: d,
	}
}

type VaultSecretReference struct {
	// Path is the path to the secret
	// +kubebuilder:validation:Required
	Path string `json:"path,omitempty"`
}

func GetFinalizer(instance client.Object) string {
	return "controller-" + strings.ToLower(instance.GetObjectKind().GroupVersionKind().Kind)
}
