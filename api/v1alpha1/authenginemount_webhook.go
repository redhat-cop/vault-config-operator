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

package v1alpha1

import (
	"errors"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var authenginemountlog = logf.Log.WithName("authenginemount-resource")

func (r *AuthEngineMount) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-authenginemount,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=authenginemounts,verbs=create;update,versions=v1alpha1,name=mauthenginemount.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &AuthEngineMount{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AuthEngineMount) Default() {
	authenginemountlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-authenginemount,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=authenginemounts,verbs=create;update,versions=v1alpha1,name=vauthenginemount.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &AuthEngineMount{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AuthEngineMount) ValidateCreate() error {
	authenginemountlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AuthEngineMount) ValidateUpdate(old runtime.Object) error {
	authenginemountlog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*AuthEngineMount).Spec.Path {
		return errors.New("spec.path cannot be updated")
	}
	// only mount config can be modified
	oldMount := old.(*AuthEngineMount).Spec.AuthMount
	newMount := r.Spec.AuthMount
	oldMount.Config = AuthMountConfig{}
	newMount.Config = AuthMountConfig{}
	if !reflect.DeepEqual(oldMount, newMount) {
		return errors.New("only .spec.config can be modified")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AuthEngineMount) ValidateDelete() error {
	authenginemountlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
