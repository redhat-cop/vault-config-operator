package controller

import (
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=passwordpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=passwordpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=passwordpolicies/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupPasswordPolicyController registers the PasswordPolicy controller with the Manager.
func SetupPasswordPolicyController(mgr ctrl.Manager) error {
	return vaultresourcecontroller.SetupGenericController[redhatcopv1alpha1.PasswordPolicy](mgr, "PasswordPolicy", vaultresourcecontroller.StandardVaultResourceFactory)
}
