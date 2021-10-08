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

	"github.com/go-logr/logr"
	"github.com/redhat-cop/operator-utils/pkg/util"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// DatabaseSecretEngineConfigReconciler reconciles a DatabaseSecretEngineConfig object
type DatabaseSecretEngineConfigReconciler struct {
	util.ReconcilerBase
	Log            logr.Logger
	ControllerName string
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=serviceaccounts;secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=databasesecretengineconfigs;randomsecrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatabaseSecretEngineConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *DatabaseSecretEngineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the instance
	instance := &redhatcopv1alpha1.DatabaseSecretEngineConfig{}
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
func (r *DatabaseSecretEngineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// this will filter routes that have the annotation and on update only if the annotation is changed.
	isBasicAuthSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/basic-auth" {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return bytes.Compare(oldSecret.Data["username"], newSecret.Data["username"]) != 0 || bytes.Compare(oldSecret.Data["password"], newSecret.Data["password"]) != 0
		},
		CreateFunc: func(e event.CreateEvent) bool {
			newSecret, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/basic-auth" {
				return false
			}
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},

		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	isUpdatedRandomSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*redhatcopv1alpha1.RandomSecret)
			if !ok {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*redhatcopv1alpha1.RandomSecret)
			if !ok {
				return true
			}
			return !newSecret.Status.LastVaultSecretUpdate.Time.Equal(oldSecret.Status.LastVaultSecretUpdate.Time)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},

		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.DatabaseSecretEngineConfig{}).
		Watches(&source.Kind{Type: &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Namespace",
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			dbsecs, err := r.findApplicableBDSCForSecret(s)
			if err != nil {
				r.Log.Error(err, "unable to find applicable vaultRoles for namespace", "namespace", s.Name)
				return []reconcile.Request{}
			}
			for _, dbsec := range dbsecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dbsec.GetName(),
						Namespace: dbsec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isBasicAuthSecret)).
		Watches(&source.Kind{Type: &redhatcopv1alpha1.RandomSecret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Namespace",
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			rs := a.(*redhatcopv1alpha1.RandomSecret)
			dbsecs, err := r.findApplicableDBSCForRandomSecret(rs)
			if err != nil {
				r.Log.Error(err, "unable to find applicable vaultRoles for namespace", "namespace", rs.Name)
				return []reconcile.Request{}
			}
			for _, dbsec := range dbsecs {
				res = append(res, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dbsec.GetName(),
						Namespace: dbsec.GetNamespace(),
					},
				})
			}
			return res
		}), builder.WithPredicates(isUpdatedRandomSecret)).
		Complete(r)
}

func (r *DatabaseSecretEngineConfigReconciler) findApplicableBDSCForSecret(secret *corev1.Secret) ([]redhatcopv1alpha1.DatabaseSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.DatabaseSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.DatabaseSecretEngineConfigList{}
	err := r.GetClient().List(context.TODO(), vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of DatabaseSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.RootCredentials.Secret != nil && vr.Spec.RootCredentials.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
func (r *DatabaseSecretEngineConfigReconciler) findApplicableDBSCForRandomSecret(randomSecret *redhatcopv1alpha1.RandomSecret) ([]redhatcopv1alpha1.DatabaseSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.DatabaseSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.DatabaseSecretEngineConfigList{}
	err := r.GetClient().List(context.TODO(), vrl, &client.ListOptions{
		Namespace: randomSecret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of DatabaseSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.RootCredentials.RandomSecret != nil && vr.Spec.RootCredentials.RandomSecret.Name == randomSecret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}

func (r *DatabaseSecretEngineConfigReconciler) IsValid(obj metav1.Object) (bool, error) {
	instance, ok := obj.(*redhatcopv1alpha1.DatabaseSecretEngineConfig)
	if !ok {
		return false, errors.New("unable to conver metav1.Object to *VaultRoleReconciler")
	}
	err := instance.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
	return err != nil, err
}

func (r *DatabaseSecretEngineConfigReconciler) IsInitialized(obj metav1.Object) bool {
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
	return isInitialized
}

func (r *DatabaseSecretEngineConfigReconciler) manageCleanUpLogic(context context.Context, instance *redhatcopv1alpha1.DatabaseSecretEngineConfig) error {
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

func (r *DatabaseSecretEngineConfigReconciler) setInternalCredentials(context context.Context, instance *redhatcopv1alpha1.DatabaseSecretEngineConfig, vaultEndpoint *vaultutils.VaultEndpoint) error {
	if instance.Spec.RootCredentials.RandomSecret != nil {
		randomSecret := &redhatcopv1alpha1.RandomSecret{}
		err := r.GetClient().Get(context, types.NamespacedName{
			Namespace: instance.Namespace,
			Name:      instance.Spec.RootCredentials.RandomSecret.Name,
		}, randomSecret)
		if err != nil {
			r.Log.Error(err, "unable to retrieve RandomSecret", "instance", instance)
			return err
		}
		secret, err := vaultEndpoint.GetVaultClient().Logical().Read(randomSecret.GetPath())
		if err != nil {
			r.Log.Error(err, "unable to retrieve vault secret", "instance", instance)
			return err
		}
		instance.SetUsernameAndPassword(instance.Spec.Username, secret.Data[randomSecret.Spec.SecretKey].(string))
		return nil
	}
	if instance.Spec.RootCredentials.Secret != nil {
		secret := &corev1.Secret{}
		err := r.GetClient().Get(context, types.NamespacedName{
			Namespace: instance.Namespace,
			Name:      instance.Spec.RootCredentials.RandomSecret.Name,
		}, secret)
		if err != nil {
			r.Log.Error(err, "unable to retrieve Secret", "instance", instance)
			return err
		}
		if username, ok := secret.Data["username"]; ok {
			instance.SetUsernameAndPassword(string(username), string(secret.Data["password"]))
		} else {
			instance.SetUsernameAndPassword(instance.Spec.Username, string(secret.Data["password"]))
		}
		return nil
	}
	if instance.Spec.RootCredentials.VaultSecret != nil {
		secret, err := vaultEndpoint.GetVaultClient().Logical().Read(string(instance.Spec.Path))
		if err != nil {
			r.Log.Error(err, "unable to retrieve vault secret", "instance", instance)
			return err
		}
		if username, ok := secret.Data["username"]; ok {
			instance.SetUsernameAndPassword(username.(string), secret.Data["password"].(string))
		} else {
			instance.SetUsernameAndPassword(instance.Spec.Username, secret.Data["password"].(string))
		}
		return nil
	}
	return errors.New("no means of retrieving a secret was specified")
}

func (r *DatabaseSecretEngineConfigReconciler) manageReconcileLogic(context context.Context, instance *redhatcopv1alpha1.DatabaseSecretEngineConfig) error {
	vaultEndpoint, err := vaultutils.NewVaultEndpoint(context, &instance.Spec.Authentication, instance, instance.Namespace, r.GetClient(), r.Log.WithName("vaultutils"))
	if err != nil {
		r.Log.Error(err, "unable to initialize vaultEndpoint with", "instance", instance)
		return err
	}
	err = r.setInternalCredentials(context, instance, vaultEndpoint)
	if err != nil {
		r.Log.Error(err, "unable to retrieve needed secret", "instance", instance)
		return err
	}
	err = vaultEndpoint.CreateOrUpdate()
	if err != nil {
		r.Log.Error(err, "unable to create/update VaultRole", "instance", instance)
		return err
	}
	return nil
}
