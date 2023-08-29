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
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var quaysecretengineconfiglog = logf.Log.WithName("quaysecretengineconfig-resource")

func (r *QuaySecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-quaysecretengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=quaysecretengineconfigs,verbs=create,versions=v1alpha1,name=mquaysecretengineconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &QuaySecretEngineConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *QuaySecretEngineConfig) Default() {
	quaysecretengineconfiglog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-quaysecretengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=quaysecretengineconfigs,verbs=create;update,versions=v1alpha1,name=vquaysecretengineconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &QuaySecretEngineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *QuaySecretEngineConfig) ValidateCreate() (admission.Warnings, error) {
	quaysecretengineconfiglog.Info("validate create", "name", r.Name)

	return nil, r.isValid()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *QuaySecretEngineConfig) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	quaysecretengineconfiglog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*QuaySecretEngineConfig).Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, r.isValid()

}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *QuaySecretEngineConfig) ValidateDelete() (admission.Warnings, error) {
	quaysecretengineconfiglog.Info("validate delete", "name", r.Name)

	return nil, nil
}
