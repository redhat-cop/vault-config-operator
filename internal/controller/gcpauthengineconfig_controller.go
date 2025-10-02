/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	vaultresourcecontroller2 "github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
)

// GCPAuthEngineConfigReconciler reconciles a GCPAuthEngineConfig object
type GCPAuthEngineConfigReconciler struct {
	vaultresourcecontroller2.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *GCPAuthEngineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	instance := &redhatcopv1alpha1.GCPAuthEngineConfig{}
	err := r.GetClient().Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ctx1, err := prepareContext(ctx, r.ReconcilerBase, instance)
	if err != nil {
		r.Log.Error(err, "unable to prepare context", "instance", instance)
		return vaultresourcecontroller2.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
	}
	vaultResource := vaultresourcecontroller2.NewVaultResource(&r.ReconcilerBase, instance)

	return vaultResource.Reconcile(ctx1, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPAuthEngineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
			return bytes.Equal(oldSecret.Data["serviceaccount"], newSecret.Data["credentials"]) || bytes.Equal(oldSecret.Data["credentials"], newSecret.Data["credentials"])
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
		For(&redhatcopv1alpha1.GCPAuthEngineConfig{}, builder.WithPredicates(vaultresourcecontroller2.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			dbsecs, err := r.findApplicableGAEForSecret(ctx, s)
			if err != nil {
				r.Log.Error(err, "unable to find applicable GCPAuthEngine for namespace", "namespace", s.Name)
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
		Watches(&redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "RandomSecret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			dbsecs, err := r.findApplicableGAEForRandomSecret(ctx, rs)
			if err != nil {
				r.Log.Error(err, "unable to find applicable GCPAuthEngine for namespace", "namespace", rs.Name)
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
		Complete(r)

}

func (r *GCPAuthEngineConfigReconciler) findApplicableGAEForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.GCPAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.GCPAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.GCPAuthEngineConfigList{}
	err := r.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of GCPAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.GCPCredentials.Secret != nil && vr.Spec.GCPCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (r *GCPAuthEngineConfigReconciler) findApplicableGAEForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.GCPAuthEngineConfig, error) {
	result := []redhatcopv1alpha1.GCPAuthEngineConfig{}
	vrl := &redhatcopv1alpha1.GCPAuthEngineConfigList{}
	err := r.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of GCPAuthEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.GCPCredentials.RandomSecret != nil && vr.Spec.GCPCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
