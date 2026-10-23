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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconcilable is implemented by VaultResource, VaultEngineResource,
// VaultAuditResource, and VaultPKIEngineResource.
type Reconcilable interface {
	Reconcile(ctx context.Context, instance client.Object) (ctrl.Result, error)
}

// Compile-time interface checks
var _ Reconcilable = &VaultResource{}
var _ Reconcilable = &VaultEngineResource{}
var _ Reconcilable = &VaultAuditResource{}
var _ Reconcilable = &VaultPKIEngineResource{}

// VaultResourceFactory creates the appropriate Reconcilable for a given instance.
type VaultResourceFactory func(base *ReconcilerBase, obj client.Object) Reconcilable

// StandardVaultResourceFactory wraps NewVaultResource to return a Reconcilable.
func StandardVaultResourceFactory(base *ReconcilerBase, obj client.Object) Reconcilable {
	return NewVaultResource(base, obj)
}

// VaultEngineResourceFactory wraps NewVaultEngineResource to return a Reconcilable.
func VaultEngineResourceFactory(base *ReconcilerBase, obj client.Object) Reconcilable {
	return NewVaultEngineResource(base, obj)
}

// VaultAuditResourceFactory wraps NewVaultAuditResource to return a Reconcilable.
func VaultAuditResourceFactory(base *ReconcilerBase, obj client.Object) Reconcilable {
	return NewVaultAuditResource(base, obj)
}

// VaultPKIEngineResourceFactory wraps NewVaultPKIEngineResource to return a Reconcilable.
func VaultPKIEngineResourceFactory(base *ReconcilerBase, obj client.Object) Reconcilable {
	return NewVaultPKIEngineResource(base, obj)
}

// GenericVaultReconciler is a generic reconciler for standard VaultObject types.
// T is the value type (e.g. Policy), PT is the pointer type constraint.
type GenericVaultReconciler[T any, PT interface {
	*T
	client.Object
	vaultutils.VaultObject
	vaultutils.ConditionsAware
}] struct {
	ReconcilerBase
	Factory VaultResourceFactory
}

// Reconcile implements the standard reconcile loop: Get → PrepareContext → Factory → Reconcile.
func (r *GenericVaultReconciler[T, PT]) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := PT(new(T))
	err := r.GetClient().Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	ctx1, err := PrepareContext(ctx, r.ReconcilerBase, instance)
	if err != nil {
		r.Log.Error(err, "unable to prepare context", "instance", instance)
		return ManageOutcome(ctx, r.ReconcilerBase, instance, err)
	}

	resource := r.Factory(&r.ReconcilerBase, instance)
	return resource.Reconcile(ctx1, instance)
}

// SetupGenericController registers a generic controller for a standard VaultObject type.
func SetupGenericController[T any, PT interface {
	*T
	client.Object
	vaultutils.VaultObject
	vaultutils.ConditionsAware
}](mgr manager.Manager, name string, factory VaultResourceFactory) error {
	reconciler := &GenericVaultReconciler[T, PT]{
		ReconcilerBase: NewFromManager(mgr, name),
		Factory:        factory,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(PT(new(T)), builder.WithPredicates(NewDefaultPeriodicReconcilePredicate())).
		Complete(reconciler)
}

// SetupGenericControllerWithBuilder registers a generic controller and passes
// the controller builder to the customize function for adding custom Watches.
func SetupGenericControllerWithBuilder[T any, PT interface {
	*T
	client.Object
	vaultutils.VaultObject
	vaultutils.ConditionsAware
}](mgr manager.Manager, name string, factory VaultResourceFactory,
	customize func(*GenericVaultReconciler[T, PT], *builder.Builder) *builder.Builder,
) error {
	reconciler := &GenericVaultReconciler[T, PT]{
		ReconcilerBase: NewFromManager(mgr, name),
		Factory:        factory,
	}
	b := ctrl.NewControllerManagedBy(mgr).
		For(PT(new(T)), builder.WithPredicates(NewDefaultPeriodicReconcilePredicate()))
	b = customize(reconciler, b)
	return b.Complete(reconciler)
}
