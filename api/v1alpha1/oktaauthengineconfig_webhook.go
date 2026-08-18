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

var oktaauthengineconfiglog = logf.Log.WithName("oktaauthengineconfig-resource")

func (r *OktaAuthEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-oktaauthengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=oktaauthengineconfigs,verbs=create,versions=v1alpha1,name=moktaauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*OktaAuthEngineConfig] = &OktaAuthEngineConfig{}

func (r *OktaAuthEngineConfig) Default(ctx context.Context, obj *OktaAuthEngineConfig) error {
	oktaauthengineconfiglog.Info("default", "name", obj.Name)
	if obj.Spec.OktaCredentials.PasswordKey == "" {
		obj.Spec.OktaCredentials.PasswordKey = "api_token"
	}
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-oktaauthengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=oktaauthengineconfigs,verbs=create;update,versions=v1alpha1,name=voktaauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*OktaAuthEngineConfig] = &OktaAuthEngineConfig{}

func (r *OktaAuthEngineConfig) ValidateCreate(ctx context.Context, obj *OktaAuthEngineConfig) (admission.Warnings, error) {
	oktaauthengineconfiglog.Info("validate create", "name", obj.Name)
	return nil, obj.isValid()
}

func (r *OktaAuthEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *OktaAuthEngineConfig) (admission.Warnings, error) {
	oktaauthengineconfiglog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, newObj.isValid()
}

func (r *OktaAuthEngineConfig) ValidateDelete(ctx context.Context, obj *OktaAuthEngineConfig) (admission.Warnings, error) {
	oktaauthengineconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
