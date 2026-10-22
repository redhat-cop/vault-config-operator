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

// aWSAuthEngineClientConfigHelper provides Watch handler methods for AWSAuthEngineClientConfig.
type aWSAuthEngineClientConfigHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineclientconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineclientconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineclientconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupAWSAuthEngineClientConfigController registers the AWSAuthEngineClientConfig controller with the Manager.
func SetupAWSAuthEngineClientConfigController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.AWSAuthEngineClientConfig,
		*redhatcopv1alpha1.AWSAuthEngineClientConfig,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AWSAuthEngineClientConfig"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &aWSAuthEngineClientConfigHelper{ReconcilerBase: reconciler.ReconcilerBase}

	isAWSCredentialSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok || (newSecret.Type != corev1.SecretTypeOpaque && newSecret.Type != corev1.SecretTypeBasicAuth) {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return !reflect.DeepEqual(oldSecret.Data, newSecret.Data)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			newSecret, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			if !ok || (newSecret.Type != corev1.SecretTypeOpaque && newSecret.Type != corev1.SecretTypeBasicAuth) {
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
		For(&redhatcopv1alpha1.AWSAuthEngineClientConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			configs, err := h.findApplicableAWSAuthConfigsForSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable AWSAuthEngineClientConfig for namespace", "namespace", s.Name)
				return []reconcile.Request{}
			}
			for _, config := range configs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      config.GetName(),
						Namespace: config.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isAWSCredentialSecret)).
		Watches(&redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "RandomSecret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			configs, err := h.findApplicableAWSAuthConfigsForRandomSecret(ctx, rs)
			if err != nil {
				h.Log.Error(err, "unable to find applicable AWSAuthEngineClientConfig for namespace", "namespace", rs.Name)
				return []reconcile.Request{}
			}
			for _, config := range configs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      config.GetName(),
						Namespace: config.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(reconciler)
}

func (h *aWSAuthEngineClientConfigHelper) findApplicableAWSAuthConfigsForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.AWSAuthEngineClientConfig, error) {
	result := []redhatcopv1alpha1.AWSAuthEngineClientConfig{}
	vrl := &redhatcopv1alpha1.AWSAuthEngineClientConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of AWSAuthEngineClientConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.AWSCredentials.Secret != nil && vr.Spec.AWSCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (h *aWSAuthEngineClientConfigHelper) findApplicableAWSAuthConfigsForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.AWSAuthEngineClientConfig, error) {
	result := []redhatcopv1alpha1.AWSAuthEngineClientConfig{}
	vrl := &redhatcopv1alpha1.AWSAuthEngineClientConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of AWSAuthEngineClientConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.AWSCredentials.RandomSecret != nil && vr.Spec.AWSCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
