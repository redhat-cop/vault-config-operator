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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var secretenginemountlog = logf.Log.WithName("secretenginemount-resource")

func (r *SecretEngineMount) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-secretenginemount,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=secretenginemounts,verbs=create,versions=v1alpha1,name=msecretenginemount.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*SecretEngineMount] = &SecretEngineMount{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *SecretEngineMount) Default(ctx context.Context, obj *SecretEngineMount) error {
	secretenginemountlog.Info("default", "name", obj.Name)
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-secretenginemount,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=secretenginemounts,verbs=update,versions=v1alpha1,name=vsecretenginemount.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*SecretEngineMount] = &SecretEngineMount{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *SecretEngineMount) ValidateCreate(ctx context.Context, obj *SecretEngineMount) (admission.Warnings, error) {
	secretenginemountlog.Info("validate create", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *SecretEngineMount) ValidateUpdate(ctx context.Context, oldObj, newObj *SecretEngineMount) (admission.Warnings, error) {
	secretenginemountlog.Info("validate update", "name", newObj.Name)

	// the path cannot be updated
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	// only mount config can be modified
	oldMount := oldObj.Spec.Mount
	newMount := newObj.Spec.Mount
	oldMount.Config = MountConfig{}
	newMount.Config = MountConfig{}
	if !reflect.DeepEqual(oldMount, newMount) {
		return nil, errors.New("only .spec.config can be modified")
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *SecretEngineMount) ValidateDelete(ctx context.Context, obj *SecretEngineMount) (admission.Warnings, error) {
	secretenginemountlog.Info("validate delete", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
