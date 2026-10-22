package controller

import (
	"context"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// rADIUSAuthEngineConfigHelper provides Watch handler methods for RADIUSAuthEngineConfig.
type rADIUSAuthEngineConfigHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupRADIUSAuthEngineConfigController registers the RADIUSAuthEngineConfig controller with the Manager.
func SetupRADIUSAuthEngineConfigController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.RADIUSAuthEngineConfig,
		*redhatcopv1alpha1.RADIUSAuthEngineConfig,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RADIUSAuthEngineConfig"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &rADIUSAuthEngineConfigHelper{ReconcilerBase: reconciler.ReconcilerBase}

	isSharedSecretSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return !equalSecretData(oldSecret, newSecret)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			_, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			return ok
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	isUpdatedRandomSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*redhatcopv1alpha1.RandomSecret)
			if !ok {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*redhatcopv1alpha1.RandomSecret)
			if !ok {
				return true
			}
			if newSecret.Status.LastVaultSecretUpdate != nil {
				if oldSecret.Status.LastVaultSecretUpdate != nil {
					return !newSecret.Status.LastVaultSecretUpdate.Time.Equal(oldSecret.Status.LastVaultSecretUpdate.Time)
				}
				return true
			}
			return false
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
		For(&redhatcopv1alpha1.RADIUSAuthEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			raecs, err := h.findApplicableRAECForSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable RADIUSAuthEngineConfigs for namespace", "namespace", s.Namespace)
				return []reconcile.Request{}
			}
			for _, raec := range raecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      raec.GetName(),
						Namespace: raec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isSharedSecretSecret)).
		Watches(&redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "RandomSecret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			raecs, err := h.findApplicableRAECForRandomSecret(ctx, rs)
			if err != nil {
				h.Log.Error(err, "unable to find applicable RADIUSAuthEngineConfigs for namespace", "namespace", rs.Namespace)
				return []reconcile.Request{}
			}
			for _, raec := range raecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      raec.GetName(),
						Namespace: raec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(reconciler)
}

func (h *rADIUSAuthEngineConfigHelper) findApplicableRAECForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.RADIUSAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.RADIUSAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.RADIUSAuthEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of RADIUSAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.RADIUSCredentials.Secret != nil && vr.Spec.RADIUSCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (h *rADIUSAuthEngineConfigHelper) findApplicableRAECForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.RADIUSAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.RADIUSAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.RADIUSAuthEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of RADIUSAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.RADIUSCredentials.RandomSecret != nil && vr.Spec.RADIUSCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
