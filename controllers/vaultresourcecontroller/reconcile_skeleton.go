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

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type deleteFunc func(ctx context.Context) error

type reconcileFunc func(ctx context.Context, instance client.Object) error

func manageCleanUpLogic(ctx context.Context, instance client.Object, deleteFn deleteFunc) error {
	log := log.FromContext(ctx)
	if vaultObject, ok := instance.(vaultutils.VaultObject); ok {
		if !vaultObject.IsDeletable() {
			return nil
		}
	}
	if conditionAware, ok := instance.(vaultutils.ConditionsAware); ok {
		for _, condition := range conditionAware.GetConditions() {
			if condition.Status == metav1.ConditionTrue && condition.Type == ReconcileSuccessful {
				err := deleteFn(ctx)
				if err != nil {
					log.Error(err, "unable to delete vault resource", "instance", instance)
					return err
				}
			}
		}
	}
	return nil
}

func ReconcileWithFunctions(ctx context.Context, reconcilerBase *ReconcilerBase, instance client.Object, cleanupFn deleteFunc, reconcileFn reconcileFunc) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("starting reconcile cycle")
	log.V(1).Info("reconcile", "instance", instance)
	if !instance.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(instance, vaultutils.GetFinalizer(instance)) {
			return reconcile.Result{}, nil
		}
		err := manageCleanUpLogic(ctx, instance, cleanupFn)
		if err != nil {
			log.Error(err, "unable to delete instance", "instance", instance)
			return ManageOutcome(ctx, *reconcilerBase, instance, err)
		}
		controllerutil.RemoveFinalizer(instance, vaultutils.GetFinalizer(instance))
		err = reconcilerBase.GetClient().Update(ctx, instance)
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return ManageOutcome(ctx, *reconcilerBase, instance, err)
		}
		return reconcile.Result{}, nil
	}
	err := reconcileFn(ctx, instance)
	if err != nil {
		log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return ManageOutcome(ctx, *reconcilerBase, instance, err)
	}
	return ManageOutcome(ctx, *reconcilerBase, instance, err)
}
