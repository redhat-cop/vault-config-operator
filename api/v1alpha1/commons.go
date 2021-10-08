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
	"time"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Pattern:=`^(.*/)([^/]*)$`
type Path string

type KubeAuthConfiguration struct {
	// ServiceAccount is the service account used for the kube auth authentication
	// +kubebuilder:validation:Required
	// +kubebuilder:default="{Name: &#34;default&#34;}"
	ServiceAccount corev1.LocalObjectReference `json:"serviceAccount,omitempty"`

	// Path is the path of the role used for this kube auth authentication
	// +kubebuilder:validation:Required
	// +kubebuilder:default=kubernetes
	Path string `json:"path,omitempty"`

	// Role the role to be used during authentication
	// +kubebuilder:validation:Required
	Role string `json:"role,omitempty"`

	//Namespace is the Vault namespace to be used in all the operations withing this connection/authentication. Only available in Vault Enterprise.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

var _ vaultutils.VaultConnection = &KubeAuthConfiguration{}

func (kc *KubeAuthConfiguration) GetNamespace() string {
	return kc.Namespace
}
func (kc *KubeAuthConfiguration) GetRole() string {
	return kc.Path
}
func (kc *KubeAuthConfiguration) GetKubeAuthPath() string {
	return kc.Role
}

func (kc *KubeAuthConfiguration) GetServiceAccountName() string {
	return kc.ServiceAccount.Name
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
