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

package controller

import (
	"bytes"
	"context"

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
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
)

// GCPSecretEngineConfigReconciler reconciles a GCPSecretEngineConfig object
type GCPSecretEngineConfigReconciler struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

func (r *GCPSecretEngineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	instance := &redhatcopv1alpha1.GCPSecretEngineConfig{}
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
		return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
	}

	vaultEndpoint := vaultutils.NewVaultEndpoint(instance)
	return vaultresourcecontroller.ReconcileWithFunctions(ctx1, &r.ReconcilerBase, instance,
		vaultEndpoint.DeleteIfExists,
		r.manageReconcileLogic,
	)
}

func (r *GCPSecretEngineConfigReconciler) manageReconcileLogic(ctx context.Context, instance client.Object) error {
	log := log.FromContext(ctx)
	if err := instance.(vaultutils.VaultObject).PrepareInternalValues(ctx, instance); err != nil {
		log.Error(err, "unable to prepare internal values", "instance", instance)
		return err
	}
	vaultEndpoint := vaultutils.NewVaultEndpoint(instance)
	if err := vaultEndpoint.Create(ctx); err != nil {
		log.Error(err, "unable to create/update vault resource", "instance", instance)
		return err
	}
	if enricher, ok := instance.(vaultutils.VaultStatusEnricher); ok {
		if enrichErr := enricher.EnrichStatus(ctx); enrichErr != nil {
			log.Error(enrichErr, "unable to enrich status from Vault", "instance", instance)
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPSecretEngineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {

	isCredentialSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			if len(oldSecret.Data) != len(newSecret.Data) {
				return true
			}
			for key, oldVal := range oldSecret.Data {
				if newVal, exists := newSecret.Data[key]; !exists || !bytes.Equal(oldVal, newVal) {
					return true
				}
			}
			return false
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
		For(&redhatcopv1alpha1.GCPSecretEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			gcpconfigs, err := r.findApplicableGCPConfigForSecret(ctx, s)
			if err != nil {
				r.Log.Error(err, "unable to find applicable GCPSecretEngineConfigs for secret", "secret", s.Name)
				return []reconcile.Request{}
			}
			for _, gcpconfig := range gcpconfigs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      gcpconfig.GetName(),
						Namespace: gcpconfig.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isCredentialSecret)).
		Watches(&redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "RandomSecret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			gcpconfigs, err := r.findApplicableGCPConfigForRandomSecret(ctx, rs)
			if err != nil {
				r.Log.Error(err, "unable to find applicable GCPSecretEngineConfigs for RandomSecret", "randomSecret", rs.Name)
				return []reconcile.Request{}
			}
			for _, gcpconfig := range gcpconfigs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      gcpconfig.GetName(),
						Namespace: gcpconfig.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(r)
}

func (r *GCPSecretEngineConfigReconciler) findApplicableGCPConfigForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.GCPSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.GCPSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.GCPSecretEngineConfigList{}
	err := r.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of GCPSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.GCPCredentials.Secret != nil && vr.Spec.GCPCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (r *GCPSecretEngineConfigReconciler) findApplicableGCPConfigForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.GCPSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.GCPSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.GCPSecretEngineConfigList{}
	err := r.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of GCPSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.GCPCredentials.RandomSecret != nil && vr.Spec.GCPCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
