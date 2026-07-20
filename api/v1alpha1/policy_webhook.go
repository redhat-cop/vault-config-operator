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
var policylog = logf.Log.WithName("policy-resource")

func (r *Policy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-policy,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=policies,verbs=create,versions=v1alpha1,name=mpolicy.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*Policy] = &Policy{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *Policy) Default(ctx context.Context, obj *Policy) error {
	policylog.Info("default", "name", obj.Name)
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-policy,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=policies,verbs=create;update,versions=v1alpha1,name=vpolicy.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*Policy] = &Policy{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *Policy) ValidateCreate(ctx context.Context, obj *Policy) (admission.Warnings, error) {
	policylog.Info("validate create", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *Policy) ValidateUpdate(ctx context.Context, oldObj, newObj *Policy) (admission.Warnings, error) {
	policylog.Info("validate update", "name", newObj.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *Policy) ValidateDelete(ctx context.Context, obj *Policy) (admission.Warnings, error) {
	policylog.Info("validate delete", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
