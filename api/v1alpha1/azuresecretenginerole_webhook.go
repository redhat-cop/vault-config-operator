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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var azuresecretenginerolelog = logf.Log.WithName("azuresecretenginerole-resource")

func (r *AzureSecretEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-azuresecretenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=azuresecretengineroles,verbs=create,versions=v1alpha1,name=mazuresecretenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &AzureSecretEngineRole{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *AzureSecretEngineRole) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*AzureSecretEngineRole)
	azuresecretenginerolelog.Info("default", "name", cr.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-azuresecretenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=azuresecretengineroles,verbs=update,versions=v1alpha1,name=vazuresecretenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &AzureSecretEngineRole{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *AzureSecretEngineRole) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*AzureSecretEngineRole)
	azuresecretenginerolelog.Info("validate create", "name", cr.Name)
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *AzureSecretEngineRole) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldCR := oldObj.(*AzureSecretEngineRole)
	newCR := newObj.(*AzureSecretEngineRole)

	azuresecretenginerolelog.Info("validate update", "name", newCR.Name)
	if newCR.Spec.Path != oldCR.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}

	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *AzureSecretEngineRole) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*AzureSecretEngineRole)
	azuresecretenginerolelog.Info("validate delete", "name", cr.Name)
	return nil, nil
}
