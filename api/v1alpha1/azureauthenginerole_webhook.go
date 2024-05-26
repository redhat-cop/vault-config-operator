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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
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

var _ webhook.Defaulter = &AzureAuthEngineRole{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AzureAuthEngineRole) Default() {
	azureauthenginerolelog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-azureauthenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=azureauthengineroles,verbs=update,versions=v1alpha1,name=vazureauthenginerole.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AzureAuthEngineRole{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AzureAuthEngineRole) ValidateCreate() (admission.Warnings, error) {
	azureauthenginerolelog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AzureAuthEngineRole) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	azureauthenginerolelog.Info("validate update", "name", r.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AzureAuthEngineRole)  ValidateDelete() (admission.Warnings, error) {
	azureauthenginerolelog.Info("validate delete", "name", r.Name)

	return nil, nil
}
