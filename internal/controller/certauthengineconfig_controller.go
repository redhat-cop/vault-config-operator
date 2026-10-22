package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=certauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=certauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=certauthengineconfigs/finalizers,verbs=update

// SetupCertAuthEngineConfigController registers the CertAuthEngineConfig controller with the Manager.
func SetupCertAuthEngineConfigController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.CertAuthEngineConfig](mgr, "CertAuthEngineConfig", vaultresourcecontroller.StandardVaultResourceFactory)
}
