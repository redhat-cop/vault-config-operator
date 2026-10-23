package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcscopes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcscopes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identityoidcscopes/finalizers,verbs=update

// SetupIdentityOIDCScopeController registers the IdentityOIDCScope controller with the Manager.
func SetupIdentityOIDCScopeController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.IdentityOIDCScope](mgr, "IdentityOIDCScope", vaultresourcecontroller.StandardVaultResourceFactory)
}
