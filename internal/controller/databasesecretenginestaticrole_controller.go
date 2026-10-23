package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretenginestaticroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretenginestaticroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretenginestaticroles/finalizers,verbs=update

// SetupDatabaseSecretEngineStaticRoleController registers the DatabaseSecretEngineStaticRole controller with the Manager.
func SetupDatabaseSecretEngineStaticRoleController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.DatabaseSecretEngineStaticRole](mgr, "DatabaseSecretEngineStaticRole", vaultresourcecontroller.StandardVaultResourceFactory)
}
