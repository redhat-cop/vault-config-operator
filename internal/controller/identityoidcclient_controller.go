package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcclients,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcclients/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcclients/finalizers,verbs=update

// SetupIdentityOIDCClientController registers the IdentityOIDCClient controller with the Manager.
func SetupIdentityOIDCClientController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.IdentityOIDCClient](mgr, "IdentityOIDCClient", vaultresourcecontroller.StandardVaultResourceFactory)
}
