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
	"context"
	"errors"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var authenginemountlog = logf.Log.WithName("authenginemount-resource")

func (r *AuthEngineMount) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-authenginemount,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=authenginemounts,verbs=create,versions=v1alpha1,name=mauthenginemount.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomDefaulter = &AuthEngineMount{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AuthEngineMount) Default(_ context.Context, obj runtime.Object) error {
	cr, ok := obj.(*AuthEngineMount)
	if !ok {
		return nil
	}
	authenginemountlog.Info("default", "name", cr.Name)
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-authenginemount,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=authenginemounts,verbs=update,versions=v1alpha1,name=vauthenginemount.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomValidator = &AuthEngineMount{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AuthEngineMount) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr, ok := obj.(*AuthEngineMount)
	if !ok {
		return nil, nil
	}
	authenginemountlog.Info("validate create", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AuthEngineMount) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newCR, ok := newObj.(*AuthEngineMount)
	if !ok {
		return nil, nil
	}
	oldCR, _ := oldObj.(*AuthEngineMount)
	authenginemountlog.Info("validate update", "name", newCR.Name)

	// the path cannot be updated
	if newCR.Spec.Path != oldCR.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	// only mount config can be modified
	oldMount := oldCR.Spec.AuthMount
	newMount := newCR.Spec.AuthMount
	oldMount.Config = AuthMountConfig{}
	newMount.Config = AuthMountConfig{}
	if !reflect.DeepEqual(oldMount, newMount) {
		return nil, errors.New("only .spec.config can be modified")
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AuthEngineMount) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr, ok := obj.(*AuthEngineMount)
	if !ok {
		return nil, nil
	}
	authenginemountlog.Info("validate delete", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
