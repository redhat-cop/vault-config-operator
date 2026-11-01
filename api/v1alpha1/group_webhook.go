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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var grouplog = logf.Log.WithName("group-resource")

func (r *Group) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &Group{}).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-group,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=groups,verbs=create;update,versions=v1alpha1,name=mgroup.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*Group] = &Group{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *Group) Default(ctx context.Context, obj *Group) error {
	grouplog.Info("default", "name", obj.Name)

	// TODO(user): fill in your defaulting logic.
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-group,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=groups,verbs=create;update,versions=v1alpha1,name=vgroup.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*Group] = &Group{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *Group) ValidateCreate(ctx context.Context, obj *Group) (admission.Warnings, error) {
	grouplog.Info("validate create", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *Group) ValidateUpdate(ctx context.Context, oldObj, newObj *Group) (admission.Warnings, error) {
	grouplog.Info("validate update", "name", newObj.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *Group) ValidateDelete(ctx context.Context, obj *Group) (admission.Warnings, error) {
	grouplog.Info("validate delete", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
