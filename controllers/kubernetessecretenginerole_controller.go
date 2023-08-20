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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
)

// KubernetesSecretEngineRoleReconciler reconciles a KubernetesSecretEngineRole object
type KubernetesSecretEngineRoleReconciler struct {
	vaultresourcecontroller.ReconcilerBase
	Log            logr.Logger
	ControllerName string
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetessecretengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetessecretengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetessecretengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KubernetesSecretEngineRole object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *KubernetesSecretEngineRoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	instance := &redhatcopv1alpha1.KubernetesSecretEngineRole{}
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
	vaultResource := vaultresourcecontroller.NewVaultResource(&r.ReconcilerBase, instance)

	return vaultResource.Reconcile(ctx1, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubernetesSecretEngineRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.KubernetesSecretEngineRole{}, builder.WithPredicates(vaultresourcecontroller.ResourceGenerationChangedPredicate{})).
		Complete(r)
}
