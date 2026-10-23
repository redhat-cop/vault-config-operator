package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenroles/finalizers,verbs=update

// SetupIdentityTokenRoleController registers the IdentityTokenRole controller with the Manager.
func SetupIdentityTokenRoleController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.IdentityTokenRole](mgr, "IdentityTokenRole", vaultresourcecontroller.StandardVaultResourceFactory)
}
