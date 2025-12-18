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
var azureauthenginerolelog = logf.Log.WithName("azureauthenginerole-resource")

func (r *AzureAuthEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-azureauthenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=azureauthengineroles,verbs=create,versions=v1alpha1,name=mazureauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &AzureAuthEngineRole{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AzureAuthEngineRole) Default(_ context.Context, obj runtime.Object) error {
	cr, ok := obj.(*AzureAuthEngineRole)
	if !ok {
		return nil
	}
	azureauthenginerolelog.Info("default", "name", cr.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-azureauthenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=azureauthengineroles,verbs=update,versions=v1alpha1,name=vazureauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &AzureAuthEngineRole{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AzureAuthEngineRole) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr, ok := obj.(*AzureAuthEngineRole)
	if !ok {
		return nil, nil
	}
	azureauthenginerolelog.Info("validate create", "name", cr.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AzureAuthEngineRole) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newCR, ok := newObj.(*AzureAuthEngineRole)
	if !ok {
		return nil, nil
	}
	_ = oldObj // currently unused; keep signature for interface compliance
	azureauthenginerolelog.Info("validate update", "name", newCR.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AzureAuthEngineRole) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr, ok := obj.(*AzureAuthEngineRole)
	if !ok {
		return nil, nil
	}
	azureauthenginerolelog.Info("validate delete", "name", cr.Name)

	return nil, nil
}
