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

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/redhat-cop/vault-config-operator/controllers/k8sevt"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// RabbitMQSecretEngineConfigReconciler reconciles a RabbitMQSecretEngineConfig object
type RabbitMQSecretEngineConfigReconciler struct {
	vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=rabbitmqsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=rabbitmqsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=randomsecrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RabbitMQSecretEngineConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *RabbitMQSecretEngineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the instance
	instance := &redhatcopv1alpha1.RabbitMQSecretEngineConfig{}
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
	if !instance.DeletionTimestamp.IsZero() {
		// No resources supported for deletion.
		return reconcile.Result{}, nil
	}

	err = r.manageReconcileLogic(ctx1, instance)
	if err != nil {
		r.Log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
	}

	return vaultresourcecontroller.ManageOutcome(ctx1, r.ReconcilerBase, instance, nil)
}

func (r *RabbitMQSecretEngineConfigReconciler) manageReconcileLogic(context context.Context, instance client.Object) error {
	log := log.FromContext(context)
	rabbitMQVaultEndpoint := vaultutils.NewRabbitMQEngineConfigVaultEndpoint(instance)
	// prepare internal values
	if err := instance.(vaultutils.VaultObject).PrepareInternalValues(context, instance); err != nil {
		log.Error(err, "unable to prepare internal values", "instance", instance)
		return err
	}
	if err := rabbitMQVaultEndpoint.Create(context); err != nil {
		log.Error(err, "unable to create/update vault resource", "instance", instance)
		return err
	}
	if err := rabbitMQVaultEndpoint.CreateOrUpdateLease(context); err != nil {
		log.Error(err, "unable to create/update lease resource", "instance", instance)
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RabbitMQSecretEngineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	filter := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return true
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
		For(&redhatcopv1alpha1.RabbitMQSecretEngineConfig{},
			builder.WithPredicates(filter, vaultresourcecontroller.ResourceGenerationChangedPredicate{}, k8sevt.Log{})).
		Complete(r)
}
