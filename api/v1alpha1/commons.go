package v1alpha1

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubeAuthConfiguration struct {
	// ServiceAccount is the service account used for the kube auth authentication
	// +kubebuilder:validation:Required
	// +kubebuilder:default={Name:default}
	ServiceAccount corev1.LocalObjectReference `json:"serviceAccount,omitempty"`

	// Path is the path of the role used for this kube auth authentication
	// +kubebuilder:validation:Required
	Path string `json:"path,omitempty"`

	//Namespace is the Vault namespace to be used in all the operations withing this connection/authentication. Only available in Vault Enterprise.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
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
