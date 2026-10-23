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

// oktaAuthEngineConfigHelper provides Watch handler methods for OktaAuthEngineConfig.
type oktaAuthEngineConfigHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupOktaAuthEngineConfigController registers the OktaAuthEngineConfig controller with the Manager.
func SetupOktaAuthEngineConfigController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.OktaAuthEngineConfig,
		*redhatcopv1alpha1.OktaAuthEngineConfig,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "OktaAuthEngineConfig"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &oktaAuthEngineConfigHelper{ReconcilerBase: reconciler.ReconcilerBase}

	isAPITokenSecret := predicate.Funcs{
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
		For(&redhatcopv1alpha1.OktaAuthEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			oaecs, err := h.findApplicableOAECForSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable OktaAuthEngineConfigs for namespace", "namespace", s.Namespace)
				return []reconcile.Request{}
			}
			for _, oaec := range oaecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      oaec.GetName(),
						Namespace: oaec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isAPITokenSecret)).
		Watches(&redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "RandomSecret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			oaecs, err := h.findApplicableOAECForRandomSecret(ctx, rs)
			if err != nil {
				h.Log.Error(err, "unable to find applicable OktaAuthEngineConfigs for namespace", "namespace", rs.Namespace)
				return []reconcile.Request{}
			}
			for _, oaec := range oaecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      oaec.GetName(),
						Namespace: oaec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(reconciler)
}

func equalSecretData(oldSecret, newSecret *corev1.Secret) bool {
	if len(oldSecret.Data) != len(newSecret.Data) {
		return false
	}
	for k, v := range oldSecret.Data {
		if nv, ok := newSecret.Data[k]; !ok || string(v) != string(nv) {
			return false
		}
	}
	return true
}

func (h *oktaAuthEngineConfigHelper) findApplicableOAECForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.OktaAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.OktaAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.OktaAuthEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of OktaAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.OktaCredentials.Secret != nil && vr.Spec.OktaCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (h *oktaAuthEngineConfigHelper) findApplicableOAECForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.OktaAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.OktaAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.OktaAuthEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of OktaAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.OktaCredentials.RandomSecret != nil && vr.Spec.OktaCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
