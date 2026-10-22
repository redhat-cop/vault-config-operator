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

// gitHubSecretEngineConfigHelper provides Watch handler methods for GitHubSecretEngineConfig.
type gitHubSecretEngineConfigHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubsecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupGitHubSecretEngineConfigController registers the GitHubSecretEngineConfig controller with the Manager.
func SetupGitHubSecretEngineConfigController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.GitHubSecretEngineConfig,
		*redhatcopv1alpha1.GitHubSecretEngineConfig,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GitHubSecretEngineConfig"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &gitHubSecretEngineConfigHelper{ReconcilerBase: reconciler.ReconcilerBase}

	// this will filter routes that have the annotation and on update only if the annotation is changed.
	isSSHSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/ssh-auth" {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return bytes.Equal(oldSecret.Data["ssh-privatekey"], newSecret.Data["ssh-privatekey"])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			newSecret, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/ssh-auth" {
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
		For(&redhatcopv1alpha1.GitHubSecretEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			dbsecs, err := h.findApplicableGHSCForSecret(ctx, s)
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
		}), builder.WithPredicates(isSSHSecret)).
		Complete(reconciler)
}

func (h *gitHubSecretEngineConfigHelper) findApplicableGHSCForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.GitHubSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.GitHubSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.GitHubSecretEngineConfigList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of DatabaseSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.SSHKeyReference.Secret != nil && vr.Spec.SSHKeyReference.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
