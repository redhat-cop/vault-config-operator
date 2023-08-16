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

	"github.com/go-logr/logr"
	"github.com/redhat-cop/operator-utils/pkg/util"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
)

// GitHubSecretEngineConfigReconciler reconciles a GitHubSecretEngineConfig object
type GitHubSecretEngineConfigReconciler struct {
	util.ReconcilerBase
	Log            logr.Logger
	ControllerName string
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubsecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GitHubSecretEngineConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *GitHubSecretEngineConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the instance
	instance := &redhatcopv1alpha1.GitHubSecretEngineConfig{}
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
func (r *GitHubSecretEngineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// this will filter routes that have the annotation and on update only if the annotation is changed.
	isSSHSecret := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newSecret, ok := e.ObjectNew.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/ssh-auth" {
				return false
			}
			oldSecret, ok := e.ObjectOld.DeepCopyObject().(*corev1.Secret)
			if !ok {
				return true
			}
			return bytes.Equal(oldSecret.Data["ssh-privatekey"], newSecret.Data["ssh-privatekey"])
		},
		CreateFunc: func(e event.CreateEvent) bool {
			newSecret, ok := e.Object.DeepCopyObject().(*corev1.Secret)
			if !ok || newSecret.Type == "kubernetes.io/ssh-auth" {
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

	return ctrl.NewControllerManagedBy(mgr).
		For(&redhatcopv1alpha1.GitHubSecretEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.ResourceGenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
		}}, handler.EnqueueRequestsFromMapFunc(func(a client.Object) []reconcile.Request {
			res := []reconcile.Request{}
			s := a.(*corev1.Secret)
			dbsecs, err := r.findApplicableGHSCForSecret(s)
			if err != nil {
				r.Log.Error(err, "unable to find applicable github SecretEngines for namespace", "namespace", s.Name)
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
		}), builder.WithPredicates(isSSHSecret)).
		Complete(r)
}

func (r *GitHubSecretEngineConfigReconciler) findApplicableGHSCForSecret(secret *corev1.Secret) ([]redhatcopv1alpha1.GitHubSecretEngineConfig, error) {
	result := []redhatcopv1alpha1.GitHubSecretEngineConfig{}
	vrl := &redhatcopv1alpha1.GitHubSecretEngineConfigList{}
	err := r.GetClient().List(context.TODO(), vrl, &client.ListOptions{
		Namespace: secret.Namespace,
	})
	if err != nil {
		r.Log.Error(err, "unable to retrieve the list of DatabaseSecretEngineConfig")
		return nil, err
	}
	for _, vr := range vrl.Items {
		if vr.Spec.SSHKeyReference.Secret != nil && vr.Spec.SSHKeyReference.Secret.Name == secret.Name {
			result = append(result, vr)
		}
	}
	return result, nil
}
