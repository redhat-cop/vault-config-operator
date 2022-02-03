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
	"bytes"
	"context"

	"errors"
	"fmt"
	"math"
	"reflect"
	"text/template"
	"time"

	"github.com/go-logr/logr"
	"github.com/redhat-cop/operator-utils/pkg/util"
	utilstemplates "github.com/redhat-cop/operator-utils/pkg/util/templates"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	vaultsecretutils "github.com/redhat-cop/vault-config-operator/controllers/vaultsecretutils"
)

const hashAnnotationName = "vaultsecret.redhatcop.redhat.io/secret-hash"

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
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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

	ctx = context.WithValue(ctx, "kubeClient", r.GetClient())

	shouldReconcile, err := r.shouldReconcile(ctx, instance)
	if err != nil {
		// There was a problem determining if the event should reconcile. Requeue the request.
		return r.ManageError(ctx, instance, err)
	}
	if !shouldReconcile {
		return r.ManageSuccess(ctx, instance)
	}

	err = r.manageReconcileLogic(ctx, instance)
	if err != nil {
		r.Log.Error(err, "unable to complete reconcile logic", "instance", instance)
		return r.ManageError(ctx, instance, err)
	}

	duration, ok := r.calculateDuration(instance)

	// If a duration incalculable, simply don't requeue
	if !ok {
		instance.Status.NextVaultSecretUpdate = nil
		return r.ManageSuccess(ctx, instance)
	}

	nextUpdateTime := instance.Status.LastVaultSecretUpdate.Add(duration)

	nextTimestamp := metav1.NewTime(nextUpdateTime)
	instance.Status.NextVaultSecretUpdate = &nextTimestamp

	//we reschedule the next reconcile at the time in the future corresponding to
	nextSchedule := time.Until(nextUpdateTime)
	if nextSchedule > 0 {
		return r.ManageSuccessWithRequeue(ctx, instance, nextSchedule)
	} else {
		return r.ManageSuccessWithRequeue(ctx, instance, time.Second)
	}

}

func (r *VaultSecretReconciler) formatK8sSecret(instance *redhatcopv1alpha1.VaultSecret, data interface{}) (*corev1.Secret, error) {

	bytesData := make(map[string][]byte)
	for k, v := range instance.Spec.TemplatizedK8sSecret.StringData {

		tpl, err := template.New("").Funcs(utilstemplates.AdvancedTemplateFuncMap(r.GetRestConfig(), r.Log)).Parse(v)
		if err != nil {
			r.Log.Error(err, "unable to create template", "instance", instance)
			return nil, err
		}

		var b bytes.Buffer
		err = tpl.Execute(&b, data)
		if err != nil {
			r.Log.Error(err, "unable to execute template", "instance", instance)
			return nil, err
		}
		bytesData[k] = b.Bytes()
	}

	annotations := instance.Spec.TemplatizedK8sSecret.Annotations
	annotations[hashAnnotationName] = vaultsecretutils.HashData(bytesData)

	k8sSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Spec.TemplatizedK8sSecret.Name,
			Namespace:   instance.Namespace,
			Annotations: annotations,
			Labels:      instance.Spec.TemplatizedK8sSecret.Labels,
		},
		Data: bytesData,
		Type: corev1.SecretType(instance.Spec.TemplatizedK8sSecret.Type),
	}

	return k8sSecret, nil
}

// Calculates the resync period based on the RefreshPeriod, and LeaseDurations returned from Vault for each secret defined (the smallest duration will be returned).
// If no RefreshPeriod or Leasedurations are found return -1 and bool of false indicating that its was incalculable.
func (r *VaultSecretReconciler) calculateDuration(instance *redhatcopv1alpha1.VaultSecret) (time.Duration, bool) {

	// if set, always use refresh period if set
	if instance.Spec.RefreshPeriod != nil {
		return instance.Spec.RefreshPeriod.Duration, true
	}

	if instance.Status.VaultSecretDefinitionsStatus != nil {
		// use the smallest LeaseDuration in the VaultDefinitionsStatus array
		var smallestLeaseDurationSeconds int = math.MaxInt64

		for _, defstat := range instance.Status.VaultSecretDefinitionsStatus {
			if defstat.LeaseDuration < smallestLeaseDurationSeconds {
				smallestLeaseDurationSeconds = defstat.LeaseDuration
			}
		}

		//No lease durations found
		if smallestLeaseDurationSeconds == math.MaxInt64 {
			return -1, false
		}

		percentage := float64(instance.Spec.RefreshThreshold) / float64(100)
		scaledSeconds := float64(smallestLeaseDurationSeconds) * percentage
		duration := time.Duration(scaledSeconds) * time.Second
		return duration, true
	}

	//No refresh period or definitions status known
	return -1, false

}

func (r *VaultSecretReconciler) shouldReconcile(ctx context.Context, instance *redhatcopv1alpha1.VaultSecret) (bool, error) {

	secretNamespacedName := &types.NamespacedName{
		Name:      instance.Spec.TemplatizedK8sSecret.Name,
		Namespace: instance.Namespace,
	}
	secret := &corev1.Secret{}
	err := r.GetClient().Get(ctx, *secretNamespacedName, secret)
	if err != nil {
		//if k8s secret does not exist, reconcile
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		// Else there was an Error reading the object. It should Requeue.
		return false, err
	} else {
		// verify secret ownership
		for _, e := range secret.ObjectMeta.OwnerReferences {
			instanceNamespacedName := &types.NamespacedName{
				Name:      instance.Name,
				Namespace: instance.Namespace,
			}
			//if the secret exists and isn't owned by this VaultSecret then the name needs to be different
			if e.Kind == instance.Kind && e.UID != instance.UID {
				return false, fmt.Errorf("the k8s Secret %v is not owned by VaultSecret %v", secretNamespacedName.String(), instanceNamespacedName.String())
			}
		}

		if secret.Annotations != nil {
			hash, ok := secret.Annotations[hashAnnotationName]
			if !ok {
				return true, nil
			}
			// if the hash value in the k8s secret doesnt matches the final data section, reconcile
			if hash != vaultsecretutils.HashData(secret.Data) {
				return true, nil
			}
			// else the hash matches. continue with logic.
		} else {
			// if annotation is nil, reconcile
			return true, nil
		}
	}

	// if the vaultsecret has reconciled before
	if instance.Status.LastVaultSecretUpdate != nil {
		duration, ok := r.calculateDuration(instance)
		// if the next duration is incalculable (no refreshperiod or lease duration), do not reconcile
		if !ok {
			return false, nil
		}
		// if the resync period has not elapsed, do not reconcile
		if !instance.Status.LastVaultSecretUpdate.Add(duration).Before(time.Now()) {
			return false, nil
		}
	}

	return true, nil
}

func (r *VaultSecretReconciler) manageReconcileLogic(ctx context.Context, instance *redhatcopv1alpha1.VaultSecret) error {

	r.Log.V(1).Info("Reconcile VaultSecret", "namespacedName", fmt.Sprintf("%v/%v", instance.Namespace, instance.Name))

	mergedMap := make(map[string]interface{})

	definitionsStatus := make([]redhatcopv1alpha1.VaultSecretDefinitionStatus, len(instance.Spec.VaultSecretDefinitions))

	for idx, vaultSecretDefinition := range instance.Spec.VaultSecretDefinitions {
		vaultClient, err := vaultSecretDefinition.Authentication.GetVaultClient(ctx, instance.Namespace)
		if err != nil {
			r.Log.Error(err, "unable to create vault client", "instance", instance)
			return err
		}

		ctx = context.WithValue(ctx, "vaultClient", vaultClient)
		vaultEndpoint := vaultutils.NewVaultEndpointObj(&vaultSecretDefinition)
		vaultSecret, ok, _ := vaultEndpoint.GetSecret(ctx)
		if !ok {
			return errors.New("unable to read Vault Secret for " + vaultSecretDefinition.GetPath())
		}

		definitionsStatus[idx] = redhatcopv1alpha1.VaultSecretDefinitionStatus{
			Name:          vaultSecretDefinition.Name,
			LeaseID:       vaultSecret.LeaseID,
			LeaseDuration: vaultSecret.LeaseDuration,
			Renewable:     vaultSecret.Renewable,
		}

		mergedMap[vaultSecretDefinition.Name] = vaultSecret.Data
	}

	k8sSecret, err := r.formatK8sSecret(instance, mergedMap)
	if err != nil {
		r.Log.Error(err, "unable to format k8s secret", "instance", instance)
		return err
	}

	err = r.CreateOrUpdateResource(ctx, instance, instance.GetNamespace(), k8sSecret)
	if err != nil {
		return err
	}

	now := metav1.NewTime(time.Now())
	instance.Status.LastVaultSecretUpdate = &now
	instance.Status.VaultSecretDefinitionsStatus = definitionsStatus

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VaultSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {

	vaultSecretPredicate := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newVaultSecret, ok := e.ObjectNew.DeepCopyObject().(*redhatcopv1alpha1.VaultSecret)
			if !ok {
				return false
			}
			oldVaultSecret, ok := e.ObjectOld.DeepCopyObject().(*redhatcopv1alpha1.VaultSecret)
			if !ok {
				return false
			}

			if !reflect.DeepEqual(oldVaultSecret.Spec, newVaultSecret.Spec) {
				r.Log.V(1).Info("Update Event - Spec changed", "object", e.ObjectNew)
				return true
			}
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			r.Log.V(1).Info("Create Event", "object", e.Object)
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			r.Log.V(1).Info("Delete Event", "object", e.Object)
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.VaultSecret{}, builder.WithPredicates(vaultSecretPredicate)).
		Owns(&corev1.Secret{}).
		Complete(r)
}
