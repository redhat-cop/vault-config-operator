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

package vaultresourcecontroller

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const ReconcileSuccessful = "ReconcileSuccessful"
const ReconcileSuccessfulReason = "LastReconcileCycleSucceded"
const ReconcileFailed = "ReconcileFailed"
const ReconcileFailedReason = "LastReconcileCycleFailed"

// SyncPeriod stores the manager's sync period for use in predicates
var SyncPeriod time.Duration = 36000 * time.Second // Default to 10 hours

// SetSyncPeriod sets the sync period for use in predicates
func SetSyncPeriod(period time.Duration) {
	SyncPeriod = period
}

// IsDriftDetectionEnabled returns whether drift detection is enabled
// Controlled via ENABLE_DRIFT_DETECTION environment variable (default: false)
func IsDriftDetectionEnabled() bool {
	if enableDrift, ok := os.LookupEnv("ENABLE_DRIFT_DETECTION"); ok {
		return enableDrift == "true"
	}
	return false
}

func IsOwner(owner, owned metav1.Object) bool {
	runtimeObj, ok := (owner).(runtime.Object)
	if !ok {
		return false
	}
	for _, ownerRef := range owned.GetOwnerReferences() {
		if ownerRef.Name == owner.GetName() && ownerRef.UID == owner.GetUID() && ownerRef.Kind == runtimeObj.GetObjectKind().GroupVersionKind().Kind {
			return true
		}
	}
	return false
}

type ReconcilerBase struct {
	apireader      client.Reader
	client         client.Client
	scheme         *runtime.Scheme
	restConfig     *rest.Config
	recorder       record.EventRecorder
	Log            logr.Logger
	ControllerName string
}

// GetClient returns the underlying client
func (r *ReconcilerBase) GetClient() client.Client {
	return r.client
}

// GetRestConfig returns the undelying rest config
func (r *ReconcilerBase) GetRestConfig() *rest.Config {
	return r.restConfig
}

// GetRecorder returns the underlying recorder
func (r *ReconcilerBase) GetRecorder() record.EventRecorder {
	return r.recorder
}

// GetScheme returns the scheme
func (r *ReconcilerBase) GetScheme() *runtime.Scheme {
	return r.scheme
}

// GetDiscoveryClient returns a discovery client for the current reconciler
func (r *ReconcilerBase) GetDiscoveryClient() (*discovery.DiscoveryClient, error) {
	return discovery.NewDiscoveryClientForConfig(r.GetRestConfig())
}

func NewReconcilerBase(client client.Client, scheme *runtime.Scheme, restConfig *rest.Config, recorder record.EventRecorder, apireader client.Reader, log logr.Logger, controllerName string) ReconcilerBase {
	return ReconcilerBase{
		apireader:      apireader,
		client:         client,
		scheme:         scheme,
		restConfig:     restConfig,
		recorder:       recorder,
		Log:            log,
		ControllerName: controllerName,
	}
}

// NewReconcilerBase is a contruction function to create a new ReconcilerBase.
func NewFromManager(mgr manager.Manager, controllerName string) ReconcilerBase {
	return NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor(controllerName), mgr.GetAPIReader(), mgr.GetLogger().WithName("controllers").WithName(controllerName), controllerName)
}

func ManageOutcomeWithRequeue(context context.Context, r ReconcilerBase, obj client.Object, issue error, requeueAfter time.Duration) (reconcile.Result, error) {
	log := log.FromContext(context)
	conditionsAware := (obj).(vaultutils.ConditionsAware)

	var condition metav1.Condition
	if issue == nil {
		condition = metav1.Condition{
			Type:               ReconcileSuccessful,
			LastTransitionTime: metav1.Now(),
			ObservedGeneration: obj.GetGeneration(),
			Reason:             ReconcileSuccessfulReason,
			Status:             metav1.ConditionTrue,
		}
	} else {
		r.GetRecorder().Event(obj, "Warning", "ProcessingError", issue.Error())
		condition = metav1.Condition{
			Type:               ReconcileFailed,
			LastTransitionTime: metav1.Now(),
			ObservedGeneration: obj.GetGeneration(),
			Message:            issue.Error(),
			Reason:             ReconcileFailedReason,
			Status:             metav1.ConditionFalse,
		}
	}
	conditionsAware.SetConditions(vaultutils.AddOrReplaceCondition(condition, conditionsAware.GetConditions()))
	err := r.GetClient().Status().Update(context, obj)
	if err != nil {
		log.Error(err, "unable to update status")
		return reconcile.Result{}, err
	}
	if vaultObject, ok := obj.(vaultutils.VaultObject); ok {
		if vaultObject.IsDeletable() {
			if issue == nil && !controllerutil.ContainsFinalizer(obj, vaultutils.GetFinalizer(obj)) {
				controllerutil.AddFinalizer(obj, vaultutils.GetFinalizer(obj))
				// BEWARE: this call *mutates* the object in memory with Kube's response, there *must be invoked last*
				err := r.GetClient().Update(context, obj)
				if err != nil {
					log.Error(err, "unable to add reconciler")
					return reconcile.Result{}, err
				}
			}
		}
	} else {
		if issue == nil && !controllerutil.ContainsFinalizer(obj, vaultutils.GetFinalizer(obj)) {
			controllerutil.AddFinalizer(obj, vaultutils.GetFinalizer(obj))
			// BEWARE: this call *mutates* the object in memory with Kube's response, there *must be invoked last*
			err := r.GetClient().Update(context, obj)
			if err != nil {
				log.Error(err, "unable to add reconciler")
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{RequeueAfter: requeueAfter}, issue
}

func ManageOutcome(context context.Context, r ReconcilerBase, obj client.Object, issue error) (reconcile.Result, error) {
	return ManageOutcomeWithRequeue(context, r, obj, issue, 0)
}

// PeriodicReconcilePredicate combines generation-based filtering with optional time-based reconciliation
// to allow periodic drift detection while avoiding excessive API calls.
// This predicate replaces ResourceGenerationChangedPredicate and provides both:
// 1. Immediate reconciliation on spec changes (generation change)
// 2. Optional periodic reconciliation for drift detection when enabled via ENABLE_DRIFT_DETECTION
type PeriodicReconcilePredicate struct {
	predicate.Funcs
	// ReconcileInterval defines how often to allow reconciliation even without spec changes
	ReconcileInterval time.Duration
}

// NewPeriodicReconcilePredicate creates a new predicate with the specified reconcile interval
func NewPeriodicReconcilePredicate(reconcileInterval time.Duration) PeriodicReconcilePredicate {
	return PeriodicReconcilePredicate{
		ReconcileInterval: reconcileInterval,
	}
}

// NewPeriodicReconcilePredicateWithSyncPeriod creates a new predicate using the manager's sync period as the reconcile interval
func NewPeriodicReconcilePredicateWithSyncPeriod(syncPeriod time.Duration) PeriodicReconcilePredicate {
	return NewPeriodicReconcilePredicate(syncPeriod)
}

// NewDefaultPeriodicReconcilePredicate creates a new predicate using the configured sync period
func NewDefaultPeriodicReconcilePredicate() PeriodicReconcilePredicate {
	return NewPeriodicReconcilePredicate(SyncPeriod)
}

// Update implements UpdateEvent filter that provides unified reconciliation logic:
// - Always reconciles on spec changes (generation change) for immediate response
// - Optionally reconciles periodically after time interval for drift detection when enabled
func (p PeriodicReconcilePredicate) Update(e event.UpdateEvent) bool {
	// Check for nil objects at the interface level and underlying object level
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	// Always reconcile if the generation changed (spec change)
	if e.ObjectNew.GetGeneration() != e.ObjectOld.GetGeneration() {
		return true
	}

	// Only do periodic drift detection if enabled
	if !IsDriftDetectionEnabled() {
		return false
	}

	// For periodic reconciliation, check if enough time has passed since last successful reconcile
	if conditionsAware, ok := e.ObjectNew.(vaultutils.ConditionsAware); ok {
		conditions := conditionsAware.GetConditions()
		for _, condition := range conditions {
			if condition.Type == ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
				// If we have a successful reconcile condition, check if the interval has elapsed
				timeSinceLastReconcile := time.Since(condition.LastTransitionTime.Time)
				if timeSinceLastReconcile >= p.ReconcileInterval {
					return true
				}
				break
			}
		}
	}

	// Don't reconcile if generation hasn't changed and interval hasn't elapsed
	return false
}

// CreateOrUpdateResource creates a resource if it doesn't exist, and updates (overwrites it), if it exist
// if owner is not nil, the owner field os set
// if namespace is not "", the namespace field of the object is overwritten with the passed value
func (r *ReconcilerBase) CreateOrUpdateResource(context context.Context, owner client.Object, namespace string, obj client.Object) error {
	log := log.FromContext(context)
	if owner != nil {
		_ = controllerutil.SetControllerReference(owner, obj, r.GetScheme())
	}
	if namespace != "" {
		obj.SetNamespace(namespace)
	}

	obj2 := &unstructured.Unstructured{}
	obj2.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	err := r.GetClient().Get(context, types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, obj2)

	if apierrors.IsNotFound(err) {
		err = r.GetClient().Create(context, obj)
		if err != nil {
			log.Error(err, "unable to create object", "object", obj)
			return err
		}
		return nil
	}
	if err == nil {
		obj.SetResourceVersion(obj2.GetResourceVersion())
		err = r.GetClient().Update(context, obj)
		if err != nil {
			log.Error(err, "unable to update object", "object", obj)
			return err
		}
		return nil

	}
	log.Error(err, "unable to lookup object", "object", obj)
	return err
}

// DeleteResourceIfExists deletes an existing resource. It doesn't fail if the resource does not exist
func (r *ReconcilerBase) DeleteResourceIfExists(context context.Context, obj client.Object) error {
	log := log.FromContext(context)
	err := r.GetClient().Delete(context, obj)
	if err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "unable to delete object ", "object", obj)
		return err
	}
	return nil
}
