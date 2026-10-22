package controller

import (
	"context"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// kerberosAuthEngineConfigHelper provides Watch handler methods for KerberosAuthEngineConfig.
type kerberosAuthEngineConfigHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupKerberosAuthEngineConfigController registers the KerberosAuthEngineConfig controller with the Manager.
func SetupKerberosAuthEngineConfigController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.KerberosAuthEngineConfig,
		*redhatcopv1alpha1.KerberosAuthEngineConfig,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KerberosAuthEngineConfig"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &kerberosAuthEngineConfigHelper{ReconcilerBase: reconciler.ReconcilerBase}

	isKeytabSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return !reflect.DeepEqual(oldSecret.Data, newSecret.Data)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.KerberosAuthEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			configs, err := h.findApplicableKAECForSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable KerberosAuthEngineConfig for namespace", "namespace", s.Namespace)
				return []reconcile.Request{}
			}
			for _, cfg := range configs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      cfg.GetName(),
						Namespace: cfg.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isKeytabSecret)).
		Complete(reconciler)
}

func (h *kerberosAuthEngineConfigHelper) findApplicableKAECForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.KerberosAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.KerberosAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.KerberosAuthEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of KerberosAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.KeytabSecret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
