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

var ociauthengineconflog = logf.Log.WithName("ociauthengineconfig-resource")

func (r *OCIAuthEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-ociauthengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=ociauthengineconfigs,verbs=create,versions=v1alpha1,name=mociauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*OCIAuthEngineConfig] = &OCIAuthEngineConfig{}

func (r *OCIAuthEngineConfig) Default(ctx context.Context, obj *OCIAuthEngineConfig) error {
	ociauthengineconflog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-ociauthengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=ociauthengineconfigs,verbs=create;update,versions=v1alpha1,name=vociauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*OCIAuthEngineConfig] = &OCIAuthEngineConfig{}

func (r *OCIAuthEngineConfig) ValidateCreate(ctx context.Context, obj *OCIAuthEngineConfig) (admission.Warnings, error) {
	ociauthengineconflog.Info("validate create", "name", obj.Name)
	return nil, nil
}

func (r *OCIAuthEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *OCIAuthEngineConfig) (admission.Warnings, error) {
	ociauthengineconflog.Info("validate update", "name", newObj.Name)

	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, nil
}

func (r *OCIAuthEngineConfig) ValidateDelete(ctx context.Context, obj *OCIAuthEngineConfig) (admission.Warnings, error) {
	ociauthengineconflog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
