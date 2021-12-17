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
	"reflect"

	"time"

	"github.com/go-logr/logr"
	"github.com/redhat-cop/operator-utils/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	template "github.com/redhat-cop/vault-config-operator/pkg/template"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultSecretReconciler reconciles a VaultSecret object
type VaultSecretReconciler struct {
	util.ReconcilerBase
	Log            logr.Logger
	ControllerName string
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultsecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VaultSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *VaultSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the instance
	instance := &redhatcopv1alpha1.VaultSecret{}
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

	if util.IsBeingDeleted(instance) {
		if !util.HasFinalizer(instance, redhatcopv1alpha1.GetFinalizer(instance)) {
			return reconcile.Result{}, nil
		}
		util.RemoveFinalizer(instance, redhatcopv1alpha1.GetFinalizer(instance))
		err = r.GetClient().Update(ctx, instance)
		if err != nil {
			r.Log.Error(err, "unable to update instance", "instance", instance)
			return r.ManageError(ctx, instance, err)
		}
		return reconcile.Result{}, nil
	}
	// how to read this if: if the secret has been initialized once and there is no refresh period or time to refresh has not arrived yet, return.
	if instance.Status.LastVaultSecretUpdate != nil && (instance.Spec.RefreshPeriod != nil || (instance.Spec.RefreshPeriod != nil && !instance.Status.LastVaultSecretUpdate.Add(instance.Spec.RefreshPeriod.Duration).Before(time.Now()))) {
		return reconcile.Result{}, nil
	}

	ctx = context.WithValue(ctx, "kubeClient", r.GetClient())

	err = r.manageReconcileLogic(ctx, instance)
	if err != nil {
		r.Log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return r.ManageError(ctx, instance, err)
	}

	//we reschedule the next reconcile at the time in the future corresponding to
	nextSchedule := time.Until(instance.Status.LastVaultSecretUpdate.Add(instance.Spec.RefreshPeriod.Duration))
	if nextSchedule > 0 {
		return r.ManageSuccessWithRequeue(ctx, instance, nextSchedule)
	} else {
		return r.ManageSuccessWithRequeue(ctx, instance, time.Second)
	}

}

func (r *VaultSecretReconciler) formatK8sSecret(instance *redhatcopv1alpha1.VaultSecret, data interface{}) (*corev1.Secret, error) {

	stringData := make(map[string]string)
	for k, v := range instance.Spec.TemplatizedK8sSecret.StringData {
		tpl, err := template.NewTemplate(&template.NewTemplateInput{
			Contents: v,
			Data:     data,
		})

		if err != nil {
			r.Log.Error(err, "unable to create template", "instance", instance)
			return nil, err
		}

		rslt, err := tpl.Execute()

		if err != nil {
			r.Log.Error(err, "unable to execute template", "instance", instance)
			return nil, err
		}
		stringData[k] = string(rslt.Output)
	}

	k8sSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Spec.TemplatizedK8sSecret.Name,
			Namespace:   instance.Namespace,
			Annotations: instance.Spec.TemplatizedK8sSecret.Annotations,
		},
		StringData: stringData,
		Type:       corev1.SecretType(instance.Spec.TemplatizedK8sSecret.Type),
	}

	return k8sSecret, nil
}

func (r *VaultSecretReconciler) manageReconcileLogic(ctx context.Context, instance *redhatcopv1alpha1.VaultSecret) error {

	mergedMap := make(map[string]interface{})
	for _, kvSecret := range instance.Spec.KVSecrets {
		vaultClient, err := kvSecret.Authentication.GetVaultClient(ctx, instance.Namespace)
		if err != nil {
			r.Log.Error(err, "unable to create vault client", "instance", instance)
			return err
		}

		ctx = context.WithValue(ctx, "vaultClient", vaultClient)
		vaultEndpoint := vaultutils.NewVaultEndpointObj(kvSecret)

		data, _, err := vaultEndpoint.Read(ctx)
		if err != nil {
			r.Log.Error(err, "unable to read Vault Secret", "instance", instance)
			return err
		}

		mergedMap[kvSecret.Name] = data
	}

	k8sSecret, err := r.formatK8sSecret(instance, mergedMap)

	if err != nil {
		r.Log.Error(err, "unable to format k8s secret", "instance", instance)
		return err
	}

	err = r.CreateOrUpdateResource(ctx, instance, instance.Namespace, k8sSecret)

	if err != nil {
		r.Log.Error(err, "unable to create or update k8s secret", "secret", k8sSecret)
		return err
	}

	now := metav1.NewTime(time.Now())
	instance.Status.LastVaultSecretUpdate = &now

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VaultSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {

	needsCreation := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*redhatcopv1alpha1.VaultSecret)
			if !ok {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*redhatcopv1alpha1.VaultSecret)
			if !ok {
				return false
			}
			//TODO this is probably wrong
			return !reflect.DeepEqual(oldSecret.Spec, newSecret.Spec)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.VaultSecret{}, builder.WithPredicates(needsCreation)).
		Owns(&corev1.Secret{}).
		Complete(r)
}
