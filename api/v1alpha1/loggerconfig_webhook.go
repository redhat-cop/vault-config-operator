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

var loggerconfiglog = logf.Log.WithName("loggerconfig-resource")

func (r *LoggerConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-loggerconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=loggerconfigs,verbs=create,versions=v1alpha1,name=mloggerconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*LoggerConfig] = &LoggerConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *LoggerConfig) Default(ctx context.Context, obj *LoggerConfig) error {
	loggerconfiglog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-loggerconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=loggerconfigs,verbs=create;update,versions=v1alpha1,name=vloggerconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*LoggerConfig] = &LoggerConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *LoggerConfig) ValidateCreate(ctx context.Context, obj *LoggerConfig) (admission.Warnings, error) {
	loggerconfiglog.Info("validate create", "name", obj.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *LoggerConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *LoggerConfig) (admission.Warnings, error) {
	loggerconfiglog.Info("validate update", "name", newObj.Name)
	if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
		return nil, err
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *LoggerConfig) ValidateDelete(ctx context.Context, obj *LoggerConfig) (admission.Warnings, error) {
	loggerconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
