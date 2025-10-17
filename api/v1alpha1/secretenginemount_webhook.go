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
var secretenginemountlog = logf.Log.WithName("secretenginemount-resource")

func (r *SecretEngineMount) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-secretenginemount,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=secretenginemounts,verbs=create,versions=v1alpha1,name=msecretenginemount.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomDefaulter = &SecretEngineMount{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SecretEngineMount) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*SecretEngineMount)
	authenginemountlog.Info("default", "name", cr.Name)
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-secretenginemount,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=secretenginemounts,verbs=update,versions=v1alpha1,name=vsecretenginemount.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomValidator = &SecretEngineMount{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SecretEngineMount) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*SecretEngineMount)
	secretenginemountlog.Info("validate create", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SecretEngineMount) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldSEM := oldObj.(*SecretEngineMount)
	cr := newObj.(*SecretEngineMount)
	secretenginemountlog.Info("validate update", "name", cr.Name)

	// the path cannot be updated
	if cr.Spec.Path != oldSEM.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	// only mount config can be modified
	oldMount := oldSEM.Spec.Mount
	newMount := cr.Spec.Mount
	oldMount.Config = MountConfig{}
	newMount.Config = MountConfig{}
	if !reflect.DeepEqual(oldMount, newMount) {
		return nil, errors.New("only .spec.config can be modified")
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SecretEngineMount) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*SecretEngineMount)
	secretenginemountlog.Info("validate delete", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
