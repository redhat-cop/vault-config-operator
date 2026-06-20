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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type VaultPKIEngineResource struct {
	vaultPKIEngineEndpoint *vaultutils.VaultPKIEngineEndpoint
	reconcilerBase         *ReconcilerBase
}

func NewVaultPKIEngineResource(reconcilerBase *ReconcilerBase, obj client.Object) *VaultPKIEngineResource {
	return &VaultPKIEngineResource{
		reconcilerBase:         reconcilerBase,
		vaultPKIEngineEndpoint: vaultutils.NewVaultPKIEngineEndpoint(obj),
	}
}

func (r *VaultPKIEngineResource) Reconcile(ctx context.Context, instance client.Object) (ctrl.Result, error) {
	return ReconcileWithFunctions(ctx, r.reconcilerBase, instance,
		r.vaultPKIEngineEndpoint.DeleteIfExists,
		r.manageReconcileLogic,
	)
}

func (r *VaultPKIEngineResource) manageReconcileLogic(context context.Context, instance client.Object) error {
	log := log.FromContext(context)
	// prepare internal values
	err := instance.(vaultutils.VaultObject).PrepareInternalValues(context, instance)
	if err != nil {
		log.Error(err, "unable to prepare internal values", "instance", instance)
		return err
	}

	//Generate
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

	}

	// Sign Intermediate
	signed := instance.(vaultutils.VaultPKIEngineObject).GetSignedStatus()
	if !signed {
		err = r.vaultPKIEngineEndpoint.CreateIntermediate(context)
		if err != nil {
			log.Error(err, "unable to create intermediate configuration", "instance", instance)
			return err
		}
		instance.(vaultutils.VaultPKIEngineObject).SetSignedStatus(true)
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
