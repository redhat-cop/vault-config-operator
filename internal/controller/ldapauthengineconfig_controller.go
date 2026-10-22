package controller

import (
	"bytes"
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

// lDAPAuthEngineConfigHelper provides Watch handler methods for LDAPAuthEngineConfig.
type lDAPAuthEngineConfigHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupLDAPAuthEngineConfigController registers the LDAPAuthEngineConfig controller with the Manager.
func SetupLDAPAuthEngineConfigController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.LDAPAuthEngineConfig,
		*redhatcopv1alpha1.LDAPAuthEngineConfig,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPAuthEngineConfig"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &lDAPAuthEngineConfigHelper{ReconcilerBase: reconciler.ReconcilerBase}

	isBasicAuthSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/basic-auth" {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return bytes.Equal(oldSecret.Data["username"], newSecret.Data["username"]) || bytes.Equal(oldSecret.Data["password"], newSecret.Data["password"])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			newSecret, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/basic-auth" {
				return false
			}
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},

		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	isTLSSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/tls" {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			// return bytes.Equal(oldSecret.Data["username"], newSecret.Data["username"]) || bytes.Equal(oldSecret.Data["password"], newSecret.Data["password"])
			return bytes.Equal(oldSecret.Data["ca.crt"], newSecret.Data["ca.crt"]) || bytes.Equal(oldSecret.Data["tls.crt"], newSecret.Data["tls.crt"]) || bytes.Equal(oldSecret.Data["tls.key"], newSecret.Data["tls.key"])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			newSecret, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/tls" {
				return false
			}
			return true
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
		For(&redhatcopv1alpha1.LDAPAuthEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			dbsecs, err := h.findApplicableLAECForSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable LDAPAuthEngines for namespace", "namespace", s.Name)
				return []reconcile.Request{}
			}
			for _, dbsec := range dbsecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dbsec.GetName(),
						Namespace: dbsec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isBasicAuthSecret)).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			dbsecs, err := h.findApplicableLAECForTLSSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable LDAPAuthEngines for namespace", "namespace", s.Name)
				return []reconcile.Request{}
			}
			for _, dbsec := range dbsecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dbsec.GetName(),
						Namespace: dbsec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isTLSSecret)).
		Watches(&redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "RandomSecret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			dbsecs, err := h.findApplicableLAECForRandomSecret(ctx, rs)
			if err != nil {
				h.Log.Error(err, "unable to find applicable LDAPAuthEngines for namespace", "namespace", rs.Name)
				return []reconcile.Request{}
			}
			for _, dbsec := range dbsecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dbsec.GetName(),
						Namespace: dbsec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(reconciler)
}

func (h *lDAPAuthEngineConfigHelper) findApplicableLAECForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.LDAPAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.LDAPAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.LDAPAuthEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of LDAPAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.BindCredentials.Secret != nil && vr.Spec.BindCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (h *lDAPAuthEngineConfigHelper) findApplicableLAECForTLSSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.LDAPAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.LDAPAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.LDAPAuthEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of LDAPAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.TLSConfig.TLSSecret != nil && vr.Spec.TLSConfig.TLSSecret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (h *lDAPAuthEngineConfigHelper) findApplicableLAECForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.LDAPAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.LDAPAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.LDAPAuthEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of LDAPAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.BindCredentials.RandomSecret != nil && vr.Spec.BindCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
