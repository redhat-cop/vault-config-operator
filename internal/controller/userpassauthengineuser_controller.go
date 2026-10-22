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

// userpassAuthEngineUserHelper provides Watch handler methods for UserpassAuthEngineUser.
type userpassAuthEngineUserHelper struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=userpassauthengineusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=userpassauthengineusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=userpassauthengineusers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// SetupUserpassAuthEngineUserController registers the UserpassAuthEngineUser controller with the Manager.
func SetupUserpassAuthEngineUserController(mgr ctrl.Manager) error {
	reconciler := &vaultresourcecontroller.GenericVaultReconciler[
		redhatcopv1alpha1.UserpassAuthEngineUser,
		*redhatcopv1alpha1.UserpassAuthEngineUser,
	]{
		ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "UserpassAuthEngineUser"),
		Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
	}
	h := &userpassAuthEngineUserHelper{ReconcilerBase: reconciler.ReconcilerBase}

	isPasswordSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/basic-auth" {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return !bytes.Equal(oldSecret.Data["password"], newSecret.Data["password"])
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
		For(&redhatcopv1alpha1.UserpassAuthEngineUser{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			users, err := h.findApplicableUsersForSecret(ctx, s)
			if err != nil {
				h.Log.Error(err, "unable to find applicable UserpassAuthEngineUser for secret", "secret", s.Name)
				return []reconcile.Request{}
			}
			for _, user := range users {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      user.GetName(),
						Namespace: user.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isPasswordSecret)).
		Watches(&redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "RandomSecret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			users, err := h.findApplicableUsersForRandomSecret(ctx, rs)
			if err != nil {
				h.Log.Error(err, "unable to find applicable UserpassAuthEngineUser for RandomSecret", "randomSecret", rs.Name)
				return []reconcile.Request{}
			}
			for _, user := range users {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      user.GetName(),
						Namespace: user.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(reconciler)
}

func (h *userpassAuthEngineUserHelper) findApplicableUsersForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.UserpassAuthEngineUser, error) {
	result := []redhatcopv1alpha1.UserpassAuthEngineUser{}
	vrl := &redhatcopv1alpha1.UserpassAuthEngineUserList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of UserpassAuthEngineUser")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.PasswordCredentials.Secret != nil && vr.Spec.PasswordCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (h *userpassAuthEngineUserHelper) findApplicableUsersForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.UserpassAuthEngineUser, error) {
	result := []redhatcopv1alpha1.UserpassAuthEngineUser{}
	vrl := &redhatcopv1alpha1.UserpassAuthEngineUserList{}
	err := h.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		h.Log.Error(err, "unable to retrieve the list of UserpassAuthEngineUser")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.PasswordCredentials.RandomSecret != nil && vr.Spec.PasswordCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
