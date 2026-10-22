package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=entities,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=entities/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=entities/finalizers,verbs=update

// SetupEntityController registers the Entity controller with the Manager.
func SetupEntityController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.Entity](mgr, "Entity", vaultresourcecontroller.StandardVaultResourceFactory)
}
