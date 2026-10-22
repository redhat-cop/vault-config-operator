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

// kerberosAuthEngineLDAPConfigHelper provides Watch handler methods for KerberosAuthEngineLDAPConfig.
type kerberosAuthEngineLDAPConfigHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineldapconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineldapconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineldapconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupKerberosAuthEngineLDAPConfigController registers the KerberosAuthEngineLDAPConfig controller with the Manager.
func SetupKerberosAuthEngineLDAPConfigController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.KerberosAuthEngineLDAPConfig,
		*redhatcopv1alpha1.KerberosAuthEngineLDAPConfig,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KerberosAuthEngineLDAPConfig"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &kerberosAuthEngineLDAPConfigHelper{ReconcilerBase: reconciler.ReconcilerBase}

	isSecretDataChanged := predicate.Funcs{
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
		For(&redhatcopv1alpha1.KerberosAuthEngineLDAPConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			configs, err := h.findApplicableKLDAPForSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable KerberosAuthEngineLDAPConfig for namespace", "namespace", s.Namespace)
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
		}), builder.WithPredicates(isSecretDataChanged)).
		Watches(&redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "RandomSecret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			configs, err := h.findApplicableKLDAPForRandomSecret(ctx, rs)
			if err != nil {
				h.Log.Error(err, "unable to find applicable KerberosAuthEngineLDAPConfig for namespace", "namespace", rs.Namespace)
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
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(reconciler)
}

func (h *kerberosAuthEngineLDAPConfigHelper) findApplicableKLDAPForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.KerberosAuthEngineLDAPConfig, error) {
	result := []redhatcopv1alpha1.KerberosAuthEngineLDAPConfig{}
	vrl := &redhatcopv1alpha1.KerberosAuthEngineLDAPConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of KerberosAuthEngineLDAPConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.LDAPCredentials.Secret != nil && vr.Spec.LDAPCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
		if vr.Spec.TLSConfig.TLSSecret != nil && vr.Spec.TLSConfig.TLSSecret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (h *kerberosAuthEngineLDAPConfigHelper) findApplicableKLDAPForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.KerberosAuthEngineLDAPConfig, error) {
	result := []redhatcopv1alpha1.KerberosAuthEngineLDAPConfig{}
	vrl := &redhatcopv1alpha1.KerberosAuthEngineLDAPConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of KerberosAuthEngineLDAPConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.LDAPCredentials.RandomSecret != nil && vr.Spec.LDAPCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
