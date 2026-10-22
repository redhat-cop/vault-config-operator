package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=entityaliases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=entityaliases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=entityaliases/finalizers,verbs=update

// SetupEntityAliasController registers the EntityAlias controller with the Manager.
func SetupEntityAliasController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.EntityAlias](mgr, "EntityAlias", vaultresourcecontroller.StandardVaultResourceFactory)
}
