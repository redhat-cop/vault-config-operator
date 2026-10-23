package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenconfigs/finalizers,verbs=update

// SetupIdentityTokenConfigController registers the IdentityTokenConfig controller with the Manager.
func SetupIdentityTokenConfigController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.IdentityTokenConfig](mgr, "IdentityTokenConfig", vaultresourcecontroller.StandardVaultResourceFactory)
}
