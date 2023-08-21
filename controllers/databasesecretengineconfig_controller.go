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
	"time"

	"github.com/go-logr/logr"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
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
)

// DatabaseSecretEngineConfigReconciler reconciles a DatabaseSecretEngineConfig object
type DatabaseSecretEngineConfigReconciler struct {
	vaultresourcecontroller.ReconcilerBase
	Log            logr.Logger
	ControllerName string
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretengineconfigs;randomsecrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatabaseSecretEngineConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *DatabaseSecretEngineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the instance
	instance := &redhatcopv1alpha1.DatabaseSecretEngineConfig{}
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

	vaultResource := vaultresourcecontroller.NewVaultResource(&r.ReconcilerBase, instance)

	_, err = vaultResource.Reconcile(ctx1, instance)

	if err != nil {
		return reconcile.Result{}, err
	}

	// if we get here the database secret engine is successfully reconciled, we can think about the root password rotation
	// if rotation is requested and no rotation timestamp exist, rotate and update the status
	//    if rotation period is define rescheduled for the period.
	// if rotation is requested and rotation period is defined and we are at more than 95% of the rotation period being passed, rotate and update status
	//    and reschedule for the period
	// if rotation is requested and rotation period is defined and we are at more than 95% reschedule for the remainder of the period

	if instance.Spec.RootPasswordRotation != nil && instance.Spec.RootPasswordRotation.Enable {
		log.V(1).Info("we need to rotate the password")
		if instance.Status.LastRootPasswordRotation.IsZero() {
			log.V(1).Info("first password rotation")
			err = r.rotateRootPassword(ctx1, instance)
			if err != nil {
				return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
			}
			if instance.Spec.RootPasswordRotation.RotationPeriod.Duration != time.Duration(0) {
				return reconcile.Result{RequeueAfter: instance.Spec.RootPasswordRotation.RotationPeriod.Duration}, nil
			}
			return reconcile.Result{}, nil
		} else {

			if instance.Spec.RootPasswordRotation.RotationPeriod.Duration != time.Duration(0) {
				log.V(1).Info("recurring password rotation")
				//(now-lastRotation)/duration > .95
				if (float64(time.Since(instance.Status.LastRootPasswordRotation.Time)) / float64(instance.Spec.RootPasswordRotation.RotationPeriod.Duration)) > 0.95 {
					log.V(1).Info("time to rotate")
					err = r.rotateRootPassword(ctx1, instance)
					if err != nil {
						return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
					}
					return reconcile.Result{RequeueAfter: instance.Spec.RootPasswordRotation.RotationPeriod.Duration}, nil
				} else {
					log.V(1).Info("not yet time to rotate")
					return reconcile.Result{RequeueAfter: time.Until(instance.Status.LastRootPasswordRotation.Time.Add(instance.Spec.RootPasswordRotation.RotationPeriod.Duration))}, nil
				}
			}
			log.V(1).Info("no need to rotate anymore")
		}
	}
	log.V(1).Info("password rotation not requested")
	return reconcile.Result{}, nil
}

func (r *DatabaseSecretEngineConfigReconciler) rotateRootPassword(ctx context.Context, instance *redhatcopv1alpha1.DatabaseSecretEngineConfig) error {
	err := instance.RotateRootPassword(ctx)
	if err != nil {
		return err
	}
	instance.Status.LastRootPasswordRotation = metav1.Now()
	err = r.GetClient().Status().Update(ctx, instance)
	if err != nil {
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseSecretEngineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// this will filter routes that have the annotation and on update only if the annotation is changed.
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
		For(&redhatcopv1alpha1.DatabaseSecretEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.ResourceGenerationChangedPredicate{})).
		Watches(&corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			dbsecs, err := r.findApplicableBDSCForSecret(ctx, s)
			if err != nil {
				r.Log.Error(err, "unable to find applicable databaseSecretEngines for namespace", "namespace", s.Name)
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
			dbsecs, err := r.findApplicableDBSCForRandomSecret(ctx, rs)
			if err != nil {
				r.Log.Error(err, "unable to find applicable databaseSecretEngines for namespace", "namespace", rs.Name)
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

func (r *DatabaseSecretEngineConfigReconciler) findApplicableBDSCForSecret(ctx context.Context, secret *corev1.Secret) ([]redhatcopv1alpha1.DatabaseSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.DatabaseSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.DatabaseSecretEngineConfigList{}
	err := r.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of DatabaseSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.RootCredentials.Secret != nil && vr.Spec.RootCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
func (r *DatabaseSecretEngineConfigReconciler) findApplicableDBSCForRandomSecret(ctx context.Context, randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.DatabaseSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.DatabaseSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.DatabaseSecretEngineConfigList{}
	err := r.GetClient().List(ctx, vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of DatabaseSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.RootCredentials.RandomSecret != nil && vr.Spec.RootCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
