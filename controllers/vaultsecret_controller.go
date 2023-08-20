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
	//utilstemplates "github.com/redhat-cop/operator-utils/pkg/util/templates"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	vaultsecretutils "github.com/redhat-cop/vault-config-operator/controllers/vaultsecretutils"
)

const (
	hashAnnotationName = "vaultsecret.redhatcop.redhat.io/secret-hash"
	vaultSecretKind    = "VaultSecret"
	secretKind         = "Secret"
	secretAPIVersion   = "v1"
)

// VaultSecretReconciler reconciles a VaultSecret object
type VaultSecretReconciler struct {
	vaultresourcecontroller.ReconcilerBase
	Log            logr.Logger
	ControllerName string
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=vaultsecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

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
	ctx = context.WithValue(ctx, "restConfig", r.GetRestConfig())

	if !instance.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(instance, vaultutils.GetFinalizer(instance)) {
			return reconcile.Result{}, nil
		}
		err := r.manageCleanUpLogic(ctx, instance)
		if err != nil {
			r.Log.Error(err, "unable to delete instance", "instance", instance)
			return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
		}
		controllerutil.RemoveFinalizer(instance, vaultutils.GetFinalizer(instance))
		err = r.GetClient().Update(ctx, instance)
		if err != nil {
			r.Log.Error(err, "unable to update instance", "instance", instance)
			return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
		}
		return reconcile.Result{}, nil
	}

	shouldSync, err := r.shouldSync(ctx, instance)
	if err != nil {
		// There was a problem determining if the event should cause a sync.
		return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
	}

	if shouldSync {
		err = r.manageSyncLogic(ctx, instance)
		if err != nil {
			r.Log.Error(err, "unable to complete sync logic", "instance", instance)
			return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
		}
	}

	duration, ok := r.calculateDuration(instance)

	// If a duration incalculable, simply don't requeue
	if !ok {
		instance.Status.NextVaultSecretUpdate = nil
		return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, nil)
	}

	nextUpdateTime := instance.Status.LastVaultSecretUpdate.Add(duration)

	nextTimestamp := metav1.NewTime(nextUpdateTime)
	instance.Status.NextVaultSecretUpdate = &nextTimestamp

	//we reschedule the next reconcile at the time in the future corresponding to
	nextSchedule := time.Until(nextUpdateTime)
	if nextSchedule > 0 {
		return vaultresourcecontroller.ManageOutcomeWithRequeue(ctx, r.ReconcilerBase, instance, err, nextSchedule)
	} else {
		return vaultresourcecontroller.ManageOutcomeWithRequeue(ctx, r.ReconcilerBase, instance, err, time.Second)
	}

}

func (r *VaultSecretReconciler) manageCleanUpLogic(context context.Context, instance *redhatcopv1alpha1.VaultSecret) error {

	k8sSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       secretKind,
			APIVersion: secretAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.TemplatizedK8sSecret.Name,
			Namespace: instance.Namespace,
		},
	}
	err := r.DeleteResourceIfExists(context, k8sSecret)
	if err != nil {
		r.Log.Error(err, "unable to delete k8s secret", "instance", instance, "k8s secret", k8sSecret)
		return err
	}
	return nil
}

func (r *VaultSecretReconciler) formatK8sSecret(instance *redhatcopv1alpha1.VaultSecret, data interface{}) (*corev1.Secret, error) {

	bytesData := make(map[string][]byte)
	for k, v := range instance.Spec.TemplatizedK8sSecret.StringData {

		tpl, err := template.New("").Funcs(vaultresourcecontroller.AdvancedTemplateFuncMap(r.GetRestConfig(), r.Log)).Parse(v)
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

	annotations := make(map[string]string)
	annotations[hashAnnotationName] = vaultsecretutils.HashData(bytesData)

	for k, v := range instance.Spec.TemplatizedK8sSecret.Annotations {
		annotations[k] = v
	}

	k8sSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       secretKind,
			APIVersion: secretAPIVersion,
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

func toNamespacedName(obj metav1.Object) string {
	if obj == nil {
		return ""
	}

	namespacedName := &types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
	return namespacedName.String()
}

func (r *VaultSecretReconciler) shouldSync(ctx context.Context, instance *redhatcopv1alpha1.VaultSecret) (bool, error) {

	secretNamespacedName := &types.NamespacedName{
		Name:      instance.Spec.TemplatizedK8sSecret.Name,
		Namespace: instance.Namespace,
	}
	secret := &corev1.Secret{}
	err := r.GetClient().Get(ctx, *secretNamespacedName, secret)
	if err != nil {
		//if k8s secret does not exist (it was deleted), it should sync
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		// else there was an Error reading the object. It should not sync.
		return false, err
	} else {

		//if the secret exists and isn't owned by this VaultSecret then the name needs to be different
		if !vaultresourcecontroller.IsOwner(instance, secret) {
			return false, fmt.Errorf("the k8s Secret %v is not owned by VaultSecret %v", secretNamespacedName.String(), toNamespacedName(instance))
		}

		if secret.Annotations != nil {
			hash, ok := secret.Annotations[hashAnnotationName]
			if !ok {
				return true, nil
			}
			// if the hash value in the k8s secret doesnt match the final data section, sync
			if hash != vaultsecretutils.HashData(secret.Data) {
				return true, nil
			}
			// else the hash matches. continue with logic.
		} else {
			// if annotation is nil, sync
			return true, nil
		}
	}

	// if the vaultsecret has synced before
	if instance.Status.LastVaultSecretUpdate != nil {
		duration, ok := r.calculateDuration(instance)
		// if the next duration is incalculable (no refreshperiod or lease duration), do not sync
		if !ok {
			return false, nil
		}
		// if the resync period has not elapsed, do not sync
		if !instance.Status.LastVaultSecretUpdate.Add(duration).Before(time.Now()) {
			return false, nil
		}
	}

	return true, nil
}

func (r *VaultSecretReconciler) manageSyncLogic(ctx context.Context, instance *redhatcopv1alpha1.VaultSecret) error {

	r.Log.V(1).Info("Sync VaultSecret", "namespacedName", toNamespacedName(instance))

	mergedMap := make(map[string]interface{})

	definitionsStatus := make([]redhatcopv1alpha1.VaultSecretDefinitionStatus, len(instance.Spec.VaultSecretDefinitions))

	for idx, vaultSecretDefinition := range instance.Spec.VaultSecretDefinitions {
		ctx = context.WithValue(ctx, "vaultConnection", vaultSecretDefinition.GetVaultConnection())
		vaultClient, err := vaultSecretDefinition.Authentication.GetVaultClient(ctx, instance.Namespace)
		if err != nil {
			r.Log.Error(err, "unable to create vault client", "instance", instance)
			return err
		}

		ctx = context.WithValue(ctx, "vaultClient", vaultClient)
		vaultSecretEndpoint := vaultutils.NewVaultSecretEndpoint(&vaultSecretDefinition)
		vaultSecret, ok, err := vaultSecretEndpoint.GetSecret(ctx)
		if err != nil {
			r.Log.Error(err, "unable to read vault secret for ", "path", vaultSecretDefinition.GetPath())
			return err
		}
		if !ok {
			return errors.New("secret not found at path: " + vaultSecretDefinition.GetPath())
		}

		r.Log.V(1).Info("", "", vaultSecret.LeaseDuration)

		definitionsStatus[idx] = redhatcopv1alpha1.VaultSecretDefinitionStatus{
			Name:          vaultSecretDefinition.Name,
			LeaseID:       vaultSecret.LeaseID,
			LeaseDuration: vaultSecret.LeaseDuration,
			Renewable:     vaultSecret.Renewable,
		}

		if vaultSecret.Data == nil {
			return errors.New("no data returned from vault secret for " + vaultSecretDefinition.GetPath())
		}

		// if its a kv v2, then the structure returned is different
		kv2DataMap, ok := vaultSecret.Data["data"]
		if ok && reflect.ValueOf(kv2DataMap).Kind() == reflect.Map {
			mergedMap[vaultSecretDefinition.Name] = kv2DataMap
		} else {
			mergedMap[vaultSecretDefinition.Name] = vaultSecret.Data
		}

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

			if !newVaultSecret.GetDeletionTimestamp().IsZero() {
				r.Log.V(1).Info("Update Event - Marked for deletion", "kind", vaultSecretKind, "namespacedName", toNamespacedName(e.ObjectNew))
				return true
			}

			if !reflect.DeepEqual(oldVaultSecret.Spec, newVaultSecret.Spec) {
				r.Log.V(1).Info("Update Event - Spec changed", "kind", vaultSecretKind, "namespacedName", toNamespacedName(e.ObjectNew))
				return true
			}

			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			r.Log.V(1).Info("Create Event", "kind", vaultSecretKind, "namespacedName", toNamespacedName(e.Object))
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			r.Log.V(1).Info("Delete Event", "kind", vaultSecretKind, "namespacedName", toNamespacedName(e.Object))
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	k8sSecretPredicate := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return false
			}

			if r.isOwnedSecretByController(newSecret) {
				hash, ok := newSecret.Annotations[hashAnnotationName]
				if !ok || hash != vaultsecretutils.HashData(newSecret.Data) {
					r.Log.V(1).Info("Update Event - hash mismatch", "kind", secretKind, "namespacedName", toNamespacedName(e.ObjectNew))
					return true
				}
			}

			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			//Useful for debugging
			secret, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return false
			}
			if r.isOwnedSecretByController(secret) {
				r.Log.V(1).Info("Delete Event", "kind", secretKind, "namespacedName", toNamespacedName(e.Object))
				return true
			}
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.VaultSecret{}, builder.WithPredicates(vaultSecretPredicate, vaultresourcecontroller.ResourceGenerationChangedPredicate{})).
		Owns(&corev1.Secret{}, builder.WithPredicates(k8sSecretPredicate)).
		Complete(r)
}

func (r *VaultSecretReconciler) isOwnedSecretByController(secret *corev1.Secret) bool {
	for _, ownerRef := range secret.ObjectMeta.OwnerReferences {
		if ownerRef.Kind == vaultSecretKind {
			return true
		}
	}
	return false
}
