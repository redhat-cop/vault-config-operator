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

// AWSSecretEngineConfigReconciler reconciles a AWSSecretEngineConfig object
type AWSSecretEngineConfigReconciler struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

func (r *AWSSecretEngineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	instance := &redhatcopv1alpha1.AWSSecretEngineConfig{}
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

func (r *AWSSecretEngineConfigReconciler) manageReconcileLogic(ctx context.Context, instance client.Object) error {
	log := log.FromContext(ctx)
	if err := instance.(vaultutils.VaultObject).PrepareInternalValues(ctx, instance); err != nil {
		log.Error(err, "unable to prepare internal values", "instance", instance)
		return err
	}
	// Always write root config: secret_key is write-only (Vault never returns it
	// on read), so drift detection cannot observe credential rotations. This follows
	// the RabbitMQ config pattern (always-write for connection credentials).
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
func (r *AWSSecretEngineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {

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
		For(&redhatcopv1alpha1.AWSSecretEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			awsconfigs, err := r.findApplicableAWSConfigForSecret(ctx, s)
			if err != nil {
				r.Log.Error(err, "unable to find applicable AWSSecretEngineConfigs for secret", "secret", s.Name)
				return []reconcile.Request{}
			}
			for _, awsconfig := range awsconfigs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      awsconfig.GetName(),
						Namespace: awsconfig.GetNamespace(),
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
			awsconfigs, err := r.findApplicableAWSConfigForRandomSecret(ctx, rs)
			if err != nil {
				r.Log.Error(err, "unable to find applicable AWSSecretEngineConfigs for RandomSecret", "randomSecret", rs.Name)
				return []reconcile.Request{}
			}
			for _, awsconfig := range awsconfigs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      awsconfig.GetName(),
						Namespace: awsconfig.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(r)
}

func (r *AWSSecretEngineConfigReconciler) findApplicableAWSConfigForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.AWSSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.AWSSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.AWSSecretEngineConfigList{}
	err := r.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of AWSSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.RootCredentials.Secret != nil && vr.Spec.RootCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (r *AWSSecretEngineConfigReconciler) findApplicableAWSConfigForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.AWSSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.AWSSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.AWSSecretEngineConfigList{}
	err := r.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of AWSSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.RootCredentials.RandomSecret != nil && vr.Spec.RootCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
