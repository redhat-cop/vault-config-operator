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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	"github.com/redhat-cop/operator-utils/pkg/util"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
)

// KubernetesAuthEngineRoleReconciler reconciles a KubernetesAuthEngineRole object
type KubernetesAuthEngineRoleReconciler struct {
	util.ReconcilerBase
	Log            logr.Logger
	ControllerName string
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetesauthengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetesauthengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kubernetesauthengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KubernetesAuthEngineRole object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *KubernetesAuthEngineRoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the instance
	instance := &redhatcopv1alpha1.KubernetesAuthEngineRole{}
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

	return vaultResource.Reconcile(ctx1, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubernetesAuthEngineRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.KubernetesAuthEngineRole{}, builder.WithPredicates(vaultresourcecontroller.ResourceGenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind: "Namespace",
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			ns := a.(*corev1.Namespace)
			ncl, err := r.findApplicableKubernetesAuthEngineRoles(ns)
			if err != nil {
				r.Log.Error(err, "unable to find applicable kubernetesAuthEngineRoles for namespace", "namespace", ns.Name)
				return []reconcile.Request{}
			}
			for _, kubernetesAuthEngineRole := range ncl {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      kubernetesAuthEngineRole.GetName(),
						Namespace: kubernetesAuthEngineRole.GetNamespace(),
					},
				})
			}
			return res
		})).
		Complete(r)
}

func (r *KubernetesAuthEngineRoleReconciler) findApplicableKubernetesAuthEngineRoles(namespace *corev1.Namespace) ([]redhatcopv1alpha1.KubernetesAuthEngineRole, error) {
	result := []redhatcopv1alpha1.KubernetesAuthEngineRole{}
	vrl := &redhatcopv1alpha1.KubernetesAuthEngineRoleList{}
	err := r.GetClient().List(context.TODO(), vrl, &client.ListOptions{})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of KubernetesAuthEngineRoles")
		return []redhatcopv1alpha1.KubernetesAuthEngineRole{}, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.TargetNamespaces.TargetNamespaceSelector != nil {
			labelSelector, err := metav1.LabelSelectorAsSelector(vr.Spec.TargetNamespaces.TargetNamespaceSelector)
			if err != nil {
				r.Log.Error(err, "unable to create selector from label selector", "selector", vr.Spec.TargetNamespaces.TargetNamespaceSelector)
				return []redhatcopv1alpha1.KubernetesAuthEngineRole{}, err
			}
			labelsAslabels := labels.Set(namespace.GetLabels())
			if labelSelector.Matches(labelsAslabels) {
				result = append(result, vr)
			}
		}
	}
	return result, nil
}
