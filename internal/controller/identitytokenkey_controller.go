package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenkeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenkeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=identitytokenkeys/finalizers,verbs=update

// SetupIdentityTokenKeyController registers the IdentityTokenKey controller with the Manager.
func SetupIdentityTokenKeyController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.IdentityTokenKey](mgr, "IdentityTokenKey", vaultresourcecontroller.StandardVaultResourceFactory)
}
