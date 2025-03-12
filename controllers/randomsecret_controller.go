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
	"context"
	"reflect"
	"time"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RandomSecretReconciler reconciles a RandomSecret object
type RandomSecretReconciler struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=randomsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=randomsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=randomsecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RandomSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *RandomSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the instance
	instance := &redhatcopv1alpha1.RandomSecret{}
	err := r.GetClient().Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	ctx1, err := prepareContext(ctx, r.ReconcilerBase, instance)
	if err != nil {
		r.Log.Error(err, "unable to prepare context", "instance", instance)
		return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
	}

	if !instance.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(instance, vaultutils.GetFinalizer(instance)) {
			return reconcile.Result{}, nil
		}
		err := r.manageCleanUpLogic(ctx1, instance)
		if err != nil {
			r.Log.Error(err, "unable to delete instance", "instance", instance)
			return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
		}
		controllerutil.RemoveFinalizer(instance, vaultutils.GetFinalizer(instance))
		err = r.GetClient().Update(ctx1, instance)
		if err != nil {
			r.Log.Error(err, "unable to update instance", "instance", instance)
			return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
		}
		return reconcile.Result{}, nil
	}

	// how to read this if: if the secret has been initialized once and there is no refresh period or time to refresh has not arrived yet, return.
	if instance.Status.LastVaultSecretUpdate != nil && (instance.Spec.RefreshPeriod == nil || (instance.Spec.RefreshPeriod != nil && !instance.Status.LastVaultSecretUpdate.Add(instance.Spec.RefreshPeriod.Duration).Before(time.Now()))) {
		return reconcile.Result{}, nil
	}

	err = r.manageReconcileLogic(ctx1, instance)
	if err != nil {
		r.Log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
	}

	if instance.Spec.RefreshPeriod.Size() > 0 {
		//we reschedule the next reconcile at the time in the future corresponding to
		nextSchedule := time.Until(instance.Status.LastVaultSecretUpdate.Add(instance.Spec.RefreshPeriod.Duration))
		if nextSchedule > 0 {
			vaultresourcecontroller.ManageOutcomeWithRequeue(ctx, r.ReconcilerBase, instance, err, nextSchedule)
			return vaultresourcecontroller.ManageOutcomeWithRequeue(ctx, r.ReconcilerBase, instance, err, nextSchedule)
		} else {
			return vaultresourcecontroller.ManageOutcomeWithRequeue(ctx, r.ReconcilerBase, instance, err, time.Second)
		}

	}
	return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
}

func (r *RandomSecretReconciler) manageCleanUpLogic(context context.Context, instance *redhatcopv1alpha1.RandomSecret) error {
	if instance.Spec.KvSecretRetainPolicy == redhatcopv1alpha1.RetainKvSecretRetainPolicy {
		return nil
	}

	vaultEndpoint := vaultutils.NewVaultEndpoint(instance)

	if instance.IsKVSecretsEngineV2() {
		err := vaultEndpoint.DeleteKVv2IfExists(context)
		if err != nil {
			r.Log.Error(err, "unable to delete KVv2 Vault Secret", "instance", instance)
			return err
		}
	} else {
		err := vaultEndpoint.DeleteIfExists(context)
		if err != nil {
			r.Log.Error(err, "unable to delete Vault Secret", "instance", instance)
			return err
		}
	}
	return nil
}

func (r *RandomSecretReconciler) manageReconcileLogic(context context.Context, instance *redhatcopv1alpha1.RandomSecret) error {
	// how to read this if: if the secret has been initialized once and there no refresh period or time ro refresh has not arrived yet, return.
	if instance.Status.LastVaultSecretUpdate != nil && (instance.Spec.RefreshPeriod == nil || (instance.Spec.RefreshPeriod != nil && !instance.Status.LastVaultSecretUpdate.Add(instance.Spec.RefreshPeriod.Duration).Before(time.Now()))) {
		return nil
	}
	vaultEndpoint := vaultutils.NewVaultEndpoint(instance)
	// When this is a newly created RandomSecret and no refresh period is defined (= one-off random password)
	// check that the Vault KV secret does not exist already to avoid overwriting its existing value.
	if instance.Status.LastVaultSecretUpdate == nil && instance.Spec.RefreshPeriod == nil {
		found, err := vaultEndpoint.Exists(context)
		if err != nil {
			r.Log.Error(err, "unable to verify secret existence", "instance", instance)
			return err
		}
		if found {
			r.Log.Info("no refresh period is defined and Vault secret already exists - nothing to do", "name", instance.Name)
			return nil
		}
	}
	err := instance.PrepareInternalValues(context, instance)
	if err != nil {
		r.Log.Error(err, "unable to generate new secret", "instance", instance)
		return err
	}
	err = vaultEndpoint.Create(context)
	if err != nil {
		r.Log.Error(err, "unable to create/update Vault Secret", "instance", instance)
		return err
	}
	now := metav1.NewTime(time.Now())
	instance.Status.LastVaultSecretUpdate = &now
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RandomSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {

	needsCreation := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*redhatcopv1alpha1.RandomSecret)
			if !ok {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*redhatcopv1alpha1.RandomSecret)
			if !ok {
				return false
			}
			return !newSecret.GetDeletionTimestamp().IsZero() || newSecret.Spec.RefreshPeriod != oldSecret.Spec.RefreshPeriod || !reflect.DeepEqual(newSecret.Spec.SecretFormat, oldSecret.Spec.SecretFormat)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},

		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.RandomSecret{}, builder.WithPredicates(needsCreation, vaultresourcecontroller.ResourceGenerationChangedPredicate{})).
		Complete(r)
}
