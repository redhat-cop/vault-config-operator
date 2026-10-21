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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var rotationpolicylog = logf.Log.WithName("rotationpolicy-resource")

func (r *RotationPolicy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-rotationpolicy,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=rotationpolicies,verbs=create,versions=v1alpha1,name=mrotationpolicy.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*RotationPolicy] = &RotationPolicy{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *RotationPolicy) Default(ctx context.Context, obj *RotationPolicy) error {
	rotationpolicylog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-rotationpolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=rotationpolicies,verbs=create;update,versions=v1alpha1,name=vrotationpolicy.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*RotationPolicy] = &RotationPolicy{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *RotationPolicy) ValidateCreate(ctx context.Context, obj *RotationPolicy) (admission.Warnings, error) {
	rotationpolicylog.Info("validate create", "name", obj.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *RotationPolicy) ValidateUpdate(ctx context.Context, oldObj, newObj *RotationPolicy) (admission.Warnings, error) {
	rotationpolicylog.Info("validate update", "name", newObj.Name)

	if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
		return nil, err
	}

	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *RotationPolicy) ValidateDelete(ctx context.Context, obj *RotationPolicy) (admission.Warnings, error) {
	rotationpolicylog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
