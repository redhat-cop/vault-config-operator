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

// kubernetesSecretEngineConfigHelper provides Watch handler methods for KubernetesSecretEngineConfig.
type kubernetesSecretEngineConfigHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetessecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetessecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetessecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupKubernetesSecretEngineConfigController registers the KubernetesSecretEngineConfig controller with the Manager.
func SetupKubernetesSecretEngineConfigController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.KubernetesSecretEngineConfig,
		*redhatcopv1alpha1.KubernetesSecretEngineConfig,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesSecretEngineConfig"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &kubernetesSecretEngineConfigHelper{ReconcilerBase: reconciler.ReconcilerBase}

	// this will filter routes that have the annotation and on update only if the annotation is changed.
	isSATokenSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == corev1.SecretTypeServiceAccountToken {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return bytes.Equal(oldSecret.Data[corev1.ServiceAccountTokenKey], newSecret.Data[corev1.ServiceAccountTokenKey])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			newSecret, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == corev1.SecretTypeServiceAccountToken {
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

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.KubernetesSecretEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			dbsecs, err := h.findApplicableKSECForSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable github SecretEngines for namespace", "namespace", s.Name)
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
		}), builder.WithPredicates(isSATokenSecret)).
		Complete(reconciler)
}

func (h *kubernetesSecretEngineConfigHelper) findApplicableKSECForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.KubernetesSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.KubernetesSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.KubernetesSecretEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of DatabaseSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.JWTReference.Secret != nil && vr.Spec.JWTReference.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
