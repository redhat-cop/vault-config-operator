package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcproviders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcproviders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcproviders/finalizers,verbs=update

// SetupIdentityOIDCProviderController registers the IdentityOIDCProvider controller with the Manager.
func SetupIdentityOIDCProviderController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.IdentityOIDCProvider](mgr, "IdentityOIDCProvider", vaultresourcecontroller.StandardVaultResourceFactory)
}
