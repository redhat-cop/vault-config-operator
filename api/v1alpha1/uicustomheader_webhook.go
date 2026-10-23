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

var uicustomheaderlog = logf.Log.WithName("uicustomheader-resource")

func (r *UICustomHeader) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-uicustomheader,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=uicustomheaders,verbs=create,versions=v1alpha1,name=muicustomheader.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*UICustomHeader] = &UICustomHeader{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *UICustomHeader) Default(ctx context.Context, obj *UICustomHeader) error {
	uicustomheaderlog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-uicustomheader,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=uicustomheaders,verbs=create;update,versions=v1alpha1,name=vuicustomheader.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*UICustomHeader] = &UICustomHeader{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *UICustomHeader) ValidateCreate(ctx context.Context, obj *UICustomHeader) (admission.Warnings, error) {
	uicustomheaderlog.Info("validate create", "name", obj.Name)
	if len(obj.Spec.Values) == 0 {
		return nil, errors.New("spec.values must contain at least one value")
	}
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *UICustomHeader) ValidateUpdate(ctx context.Context, oldObj, newObj *UICustomHeader) (admission.Warnings, error) {
	uicustomheaderlog.Info("validate update", "name", newObj.Name)
	if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
		return nil, err
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	if len(newObj.Spec.Values) == 0 {
		return nil, errors.New("spec.values must contain at least one value")
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *UICustomHeader) ValidateDelete(ctx context.Context, obj *UICustomHeader) (admission.Warnings, error) {
	uicustomheaderlog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
