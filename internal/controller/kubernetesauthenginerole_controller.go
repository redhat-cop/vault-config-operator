package controller

import (
	"context"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// kubernetesAuthEngineRoleHelper provides Watch handler methods for KubernetesAuthEngineRole.
type kubernetesAuthEngineRoleHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetesauthengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetesauthengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetesauthengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupKubernetesAuthEngineRoleController registers the KubernetesAuthEngineRole controller with the Manager.
func SetupKubernetesAuthEngineRoleController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.KubernetesAuthEngineRole,
		*redhatcopv1alpha1.KubernetesAuthEngineRole,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesAuthEngineRole"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &kubernetesAuthEngineRoleHelper{ReconcilerBase: reconciler.ReconcilerBase}

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.KubernetesAuthEngineRole{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind: "Namespace",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			ns := a.(*corev1.Namespace)
			ncl, err := h.findApplicableKubernetesAuthEngineRoles(ctx, ns)
			if err != nil {
				h.Log.Error(err, "unable to find applicable kubernetesAuthEngineRoles for namespace", "namespace", ns.Name)
				return []reconcile.Request{}
			}
			for _, kubernetesAuthEngineRole := range ncl {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      kubernetesAuthEngineRole.GetName(),
						Namespace: kubernetesAuthEngineRole.GetNamespace(),
					},
				})
			}
			return res
		})).
		Complete(reconciler)
}

func (h *kubernetesAuthEngineRoleHelper) findApplicableKubernetesAuthEngineRoles(ctx context.Context, namespace *corev1.Namespace) ([]redhatcopv1alpha1.KubernetesAuthEngineRole, error) {
	result := []redhatcopv1alpha1.KubernetesAuthEngineRole{}
	vrl := &redhatcopv1alpha1.KubernetesAuthEngineRoleList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of KubernetesAuthEngineRoles")
		return []redhatcopv1alpha1.KubernetesAuthEngineRole{}, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.TargetNamespaces.TargetNamespaceSelector != nil {
			labelSelector, err := metav1.LabelSelectorAsSelector(vr.Spec.TargetNamespaces.TargetNamespaceSelector)
			if err != nil {
				h.Log.Error(err, "unable to create selector from label selector", "selector", vr.Spec.TargetNamespaces.TargetNamespaceSelector)
				return []redhatcopv1alpha1.KubernetesAuthEngineRole{}, err
			}
			labelsAslabels := labels.Set(namespace.GetLabels())
			if labelSelector.Matches(labelsAslabels) {
				result = append(result, vr)
			}
		}
	}
	return result, nil
}
