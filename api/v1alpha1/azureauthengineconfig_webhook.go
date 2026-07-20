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

// log is for logging in this package.
var azureauthengineconfiglog = logf.Log.WithName("azureauthengineconfig-resource")

func (r *AzureAuthEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-azureauthengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=azureauthengineconfigs,verbs=create,versions=v1alpha1,name=mazureauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*AzureAuthEngineConfig] = &AzureAuthEngineConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *AzureAuthEngineConfig) Default(ctx context.Context, obj *AzureAuthEngineConfig) error {
	azureauthengineconfiglog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-azureauthengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=azureauthengineconfigs,verbs=update,versions=v1alpha1,name=vazureauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*AzureAuthEngineConfig] = &AzureAuthEngineConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *AzureAuthEngineConfig) ValidateCreate(ctx context.Context, obj *AzureAuthEngineConfig) (admission.Warnings, error) {
	azureauthengineconfiglog.Info("validate create", "name", obj.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *AzureAuthEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *AzureAuthEngineConfig) (admission.Warnings, error) {
	azureauthengineconfiglog.Info("validate update", "name", newObj.Name)

	// the path cannot be updated
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *AzureAuthEngineConfig) ValidateDelete(ctx context.Context, obj *AzureAuthEngineConfig) (admission.Warnings, error) {
	azureauthengineconfiglog.Info("validate delete", "name", obj.Name)

	return nil, nil
}
