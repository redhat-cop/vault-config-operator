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

type VaultAuditResource struct {
	vaultAuditEndpoint *vaultutils.VaultAuditEndpoint
	reconcilerBase     *ReconcilerBase
}

func NewVaultAuditResource(reconcilerBase *ReconcilerBase, obj client.Object) *VaultAuditResource {
	return &VaultAuditResource{
		reconcilerBase:     reconcilerBase,
		vaultAuditEndpoint: vaultutils.NewVaultAuditEndpoint(obj),
	}
}

func (r *VaultAuditResource) Reconcile(ctx context.Context, instance client.Object) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("starting audit reconcile cycle")
	log.V(1).Info("reconcile", "instance", instance)

	if !instance.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(instance, vaultutils.GetFinalizer(instance)) {
			return reconcile.Result{}, nil
		}
		err := r.manageCleanUpLogic(ctx, instance)
		if err != nil {
			log.Error(err, "unable to delete instance", "instance", instance)
			return ManageOutcome(ctx, *r.reconcilerBase, instance, err)
		}
		controllerutil.RemoveFinalizer(instance, vaultutils.GetFinalizer(instance))
		err = r.reconcilerBase.GetClient().Update(ctx, instance)
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return ManageOutcome(ctx, *r.reconcilerBase, instance, err)
		}
		return reconcile.Result{}, nil
	}

	err := r.manageReconcileLogic(ctx, instance)
	if err != nil {
		log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return ManageOutcome(ctx, *r.reconcilerBase, instance, err)
	}

	return ManageOutcome(ctx, *r.reconcilerBase, instance, err)
}

func (r *VaultAuditResource) manageCleanUpLogic(context context.Context, instance client.Object) error {
	log := log.FromContext(context)
	if vaultObject, ok := instance.(vaultutils.VaultObject); ok {
		if !vaultObject.IsDeletable() {
			return nil
		}
	}
	if conditionAware, ok := instance.(vaultutils.ConditionsAware); ok {
		for _, condition := range conditionAware.GetConditions() {
			if condition.Status == metav1.ConditionTrue && condition.Type == ReconcileSuccessful {
				err := r.vaultAuditEndpoint.DeleteIfExists(context)
				if err != nil {
					log.Error(err, "unable to disable audit device", "instance", instance)
					return err
				}
			}
		}
	}
	return nil
}

func (r *VaultAuditResource) manageReconcileLogic(context context.Context, instance client.Object) error {
	log := log.FromContext(context)

	// prepare internal values
	err := instance.(vaultutils.VaultObject).PrepareInternalValues(context, instance)
	if err != nil {
		log.Error(err, "unable to prepare internal values", "instance", instance)
		return err
	}

	err = instance.(vaultutils.VaultObject).PrepareTLSConfig(context, instance)
	if err != nil {
		log.Error(err, "unable to prepare TLS Config values", "instance", instance)
		return err
	}

	// Enable or update the audit device
	err = r.vaultAuditEndpoint.CreateOrUpdate(context)
	if err != nil {
		log.Error(err, "unable to enable/update audit device", "instance", instance)
		return err
	}

	return nil
}
