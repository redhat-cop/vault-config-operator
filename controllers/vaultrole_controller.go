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
	"errors"

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
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
)

// VaultRoleReconciler reconciles a VaultRole object
type VaultRoleReconciler struct {
	util.ReconcilerBase
	Log            logr.Logger
	ControllerName string
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultroles/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=serviceaccounts;secrets;namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VaultRole object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *VaultRoleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the instance
	instance := &redhatcopv1alpha1.VaultRole{}
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

	if ok, err := r.IsValid(instance); !ok {
		return r.ManageError(ctx, instance, err)
	}

	if ok := r.IsInitialized(instance); !ok {
		err := r.GetClient().Update(ctx, instance)
		if err != nil {
			r.Log.Error(err, "unable to update instance", "instance", instance)
			return r.ManageError(ctx, instance, err)
		}
		return reconcile.Result{}, nil
	}

	if util.IsBeingDeleted(instance) {
		if !util.HasFinalizer(instance, r.ControllerName) {
			return reconcile.Result{}, nil
		}
		err := r.manageCleanUpLogic(ctx, instance)
		if err != nil {
			r.Log.Error(err, "unable to delete instance", "instance", instance)
			return r.ManageError(ctx, instance, err)
		}
		util.RemoveFinalizer(instance, r.ControllerName)
		err = r.GetClient().Update(ctx, instance)
		if err != nil {
			r.Log.Error(err, "unable to update instance", "instance", instance)
			return r.ManageError(ctx, instance, err)
		}
		return reconcile.Result{}, nil
	}

	err = r.manageReconcileLogic(ctx, instance)
	if err != nil {
		r.Log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return r.ManageError(ctx, instance, err)
	}

	return r.ManageSuccess(ctx, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VaultRoleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.VaultRole{}, builder.WithPredicates(util.ResourceGenerationOrFinalizerChangedPredicate{})).
		Watches(&source.Kind{Type: &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				Kind: "Namespace",
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			ns := a.(*corev1.Namespace)
			ncl, err := r.findApplicableVaultRoles(ns)
			if err != nil {
				r.Log.Error(err, "unable to find applicable vaultRoles for namespace", "namespace", ns.Name)
				return []reconcile.Request{}
			}
			for _, vaultRole := range ncl {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      vaultRole.GetName(),
						Namespace: vaultRole.GetNamespace(),
					},
				})
			}
			return res
		})).
		Complete(r)
}

func (r *VaultRoleReconciler) findApplicableVaultRoles(namespace *corev1.Namespace) ([]redhatcopv1alpha1.VaultRole, error) {
	result := []redhatcopv1alpha1.VaultRole{}
	vrl := &redhatcopv1alpha1.VaultRoleList{}
	err := r.GetClient().List(context.TODO(), vrl, &client.ListOptions{})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of VaultRoles")
		return []redhatcopv1alpha1.VaultRole{}, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.TargetNamespaces.TargetNamespaceSelector != nil {
			labelSelector, err := metav1.LabelSelectorAsSelector(vr.Spec.TargetNamespaces.TargetNamespaceSelector)
			if err != nil {
				r.Log.Error(err, "unable to create selector from label selector", "selector", vr.Spec.TargetNamespaces.TargetNamespaceSelector)
				return []redhatcopv1alpha1.VaultRole{}, err
			}
			labelsAslabels := labels.Set(namespace.GetLabels())
			if labelSelector.Matches(labelsAslabels) {
				result = append(result, vr)
			}
		}
	}
	return result, nil
}

func (r *VaultRoleReconciler) findSelectedNamespaceNames(context context.Context, instance *redhatcopv1alpha1.VaultRole) ([]string, error) {
	result := []string{}
	namespaceList := &corev1.NamespaceList{}
	labelSelector, err := metav1.LabelSelectorAsSelector(instance.Spec.TargetNamespaces.TargetNamespaceSelector)
	if err != nil {
		r.Log.Error(err, "unable to create selector from label selector", "selector", instance.Spec.TargetNamespaces.TargetNamespaceSelector)
		return nil, err
	}
	err = r.GetClient().List(context, namespaceList, &client.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of namespaces")
		return nil, err
	}
	for i := range namespaceList.Items {
		result = append(result, namespaceList.Items[i].Name)
	}
	return result, nil
}

func (r *VaultRoleReconciler) IsValid(obj metav1.Object) (bool, error) {
	instance, ok := obj.(*redhatcopv1alpha1.VaultRole)
	if !ok {
		return false, errors.New("unable to conver metav1.Object to *VaultRoleReconciler")
	}
	err := instance.ValidateEitherTargetNamespaceSelectorOrTargetNamespace()
	return err == nil, err
}

func (r *VaultRoleReconciler) IsInitialized(obj metav1.Object) bool {
	isInitialized := true
	cobj, ok := obj.(client.Object)
	if !ok {
		r.Log.Error(errors.New("unable to convert to client.Object"), "unable to convert to client.Object")
		return false
	}
	if !util.HasFinalizer(cobj, r.ControllerName) {
		util.AddFinalizer(cobj, r.ControllerName)
		isInitialized = false
	}
	instance, ok := obj.(*redhatcopv1alpha1.VaultRole)
	if !ok {
		r.Log.Error(errors.New("unable to convert to redhatcopv1alpha1.VaultRole"), "unable to convert to redhatcopv1alpha1.VaultRole")
		return false
	}
	if instance.Spec.Authentication.ServiceAccount == nil {
		instance.Spec.Authentication.ServiceAccount = &corev1.LocalObjectReference{
			Name: "default",
		}
		isInitialized = false
	}
	return isInitialized
}

func (r *VaultRoleReconciler) manageCleanUpLogic(context context.Context, instance *redhatcopv1alpha1.VaultRole) error {
	vaultEndpoint, err := vaultutils.NewVaultEndpoint(context, &instance.Spec.Authentication, instance, instance.Namespace, r.GetClient(), r.Log.WithName("vaultutils"))
	if err != nil {
		r.Log.Error(err, "unable to initialize vaultEndpoint with", "instance", instance)
		return err
	}
	err = vaultEndpoint.DeleteIfExists()
	if err != nil {
		r.Log.Error(err, "unable to delete VaultRole", "instance", instance)
		return err
	}
	return nil
}

func (r *VaultRoleReconciler) manageReconcileLogic(context context.Context, instance *redhatcopv1alpha1.VaultRole) error {
	if instance.Spec.TargetNamespaces.TargetNamespaceSelector != nil {
		namespaces, err := r.findSelectedNamespaceNames(context, instance)
		if err != nil {
			r.Log.Error(err, "unable to retrieve selected namespaces", "instance", instance)
			return err
		}
		instance.SetInternalNamespaces(namespaces)
	} else {
		instance.SetInternalNamespaces(instance.Spec.TargetNamespaces.TargetNamespaces)
	}
	vaultEndpoint, err := vaultutils.NewVaultEndpoint(context, &instance.Spec.Authentication, instance, instance.Namespace, r.GetClient(), r.Log.WithName("vaultutils"))
	if err != nil {
		r.Log.Error(err, "unable to initialize vaultEndpoint with", "instance", instance)
		return err
	}
	err = vaultEndpoint.CreateOrUpdate()
	if err != nil {
		r.Log.Error(err, "unable to create/update VaultRole", "instance", instance)
		return err
	}
	return nil
}
