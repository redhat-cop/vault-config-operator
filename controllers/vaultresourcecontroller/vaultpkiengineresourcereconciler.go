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

	"github.com/redhat-cop/operator-utils/pkg/util"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type VaultPKIEngineResource struct {
	vaultPKIEngineEndpoint *vaultutils.VaultPKIEngineEndpoint
	reconcilerBase         *util.ReconcilerBase
}

func NewVaultPKIEngineResource(reconcilerBase *util.ReconcilerBase, obj client.Object) *VaultPKIEngineResource {
	return &VaultPKIEngineResource{
		reconcilerBase:         reconcilerBase,
		vaultPKIEngineEndpoint: vaultutils.NewVaultPKIEngineEndpoint(obj),
	}
}

func (r *VaultPKIEngineResource) manageCleanUpLogic(context context.Context, instance client.Object) error {
	log := log.FromContext(context)
	log.Info("DeleteIfExists", "Try to: ", instance)
	err := r.vaultPKIEngineEndpoint.DeleteIfExists(context)
	if err != nil {
		log.Error(err, "unable to delete vault resource", "instance", instance)
		return err
	}
	return nil
}

func (r *VaultPKIEngineResource) Reconcile(ctx context.Context, instance client.Object) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if util.IsBeingDeleted(instance) {
		log.Info("Delete", "Try to: ", instance)

		if !util.HasFinalizer(instance, redhatcopv1alpha1.GetFinalizer(instance)) {
			log.Info("Finaliter?", "Try to: ", instance)
			return reconcile.Result{}, nil
		}
		err := r.manageCleanUpLogic(ctx, instance)
		if err != nil {
			log.Error(err, "unable to delete instance", "instance", instance)
			return r.reconcilerBase.ManageError(ctx, instance, err)
		}
		log.Info("RemoveFinalizer", "Try to: ", instance)
		util.RemoveFinalizer(instance, redhatcopv1alpha1.GetFinalizer(instance))
		err = r.reconcilerBase.GetClient().Update(ctx, instance)
		if err != nil {
			log.Error(err, "unable to update instance", "instance", instance)
			return r.reconcilerBase.ManageError(ctx, instance, err)
		}
		return reconcile.Result{}, nil
	}
	err := r.manageReconcileLogic(ctx, instance)
	if err != nil {
		log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return r.reconcilerBase.ManageError(ctx, instance, err)
	}
	return r.reconcilerBase.ManageSuccess(ctx, instance)
}

func (r *VaultPKIEngineResource) manageReconcileLogic(context context.Context, instance client.Object) error {
	log := log.FromContext(context)
	// prepare internal values
	err := instance.(vaultutils.VaultObject).PrepareInternalValues(context, instance)
	if err != nil {
		log.Error(err, "unable to prepare internal values", "instance", instance)
		return err
	}
	generated := instance.(vaultutils.VaultPKIEngineObject).GetGeneratedStatus()
	if !generated {
		vaultSecret, err := r.vaultPKIEngineEndpoint.Generate(context)
		if err != nil {
			log.Error(err, "unable to generate CA", "instance", instance)
			return err
		}
		instance.(vaultutils.VaultPKIEngineObject).SetGeneratedStatus(true)

		// Exported
		exported, err := r.vaultPKIEngineEndpoint.CreateExported(context, vaultSecret)
		if err != nil {
			log.Error(err, "unable to create exported configuration", "instance", instance)
			return err
		}
		instance.(vaultutils.VaultPKIEngineObject).SetExportedStatus(exported)

		// Intermediate
		err = r.vaultPKIEngineEndpoint.CreateIntermediate(context)
		if err != nil {
			log.Error(err, "unable to create intermediate configuration", "instance", instance)
			return err
		}
	}

	// Config
	err = r.vaultPKIEngineEndpoint.CreateOrUpdateConfigUrls(context)
	if err != nil {
		log.Error(err, "unable to create or update url config", "instance", instance)
		return err
	}
	err = r.vaultPKIEngineEndpoint.CreateOrUpdateConfigCrl(context)
	if err != nil {
		log.Error(err, "unable to create or update crl config", "instance", instance)
		return err
	}

	return nil
}
