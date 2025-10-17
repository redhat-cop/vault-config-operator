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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var grouplog = logf.Log.WithName("group-resource")

func (r *Group) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-group,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=groups,verbs=create;update,versions=v1alpha1,name=mgroup.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &Group{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *Group) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*Group)
	grouplog.Info("default", "name", cr.Name)

	// TODO(user): fill in your defaulting logic.
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-group,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=groups,verbs=create;update,versions=v1alpha1,name=vgroup.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &Group{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *Group) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*Group)
	grouplog.Info("validate create", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *Group) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	_ = oldObj.(*Group)
	new := newObj.(*Group)

	grouplog.Info("validate update", "name", new.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *Group) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*Group)
	grouplog.Info("validate delete", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
