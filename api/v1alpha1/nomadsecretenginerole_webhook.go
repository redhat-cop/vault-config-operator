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

var nomadsecretenginerolelog = logf.Log.WithName("nomadsecretenginerole-resource")

func (r *NomadSecretEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-nomadsecretenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=nomadsecretengineroles,verbs=create,versions=v1alpha1,name=mnomadsecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*NomadSecretEngineRole] = &NomadSecretEngineRole{}

func (r *NomadSecretEngineRole) Default(ctx context.Context, obj *NomadSecretEngineRole) error {
	nomadsecretenginerolelog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-nomadsecretenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=nomadsecretengineroles,verbs=create;update,versions=v1alpha1,name=vnomadsecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*NomadSecretEngineRole] = &NomadSecretEngineRole{}

func (r *NomadSecretEngineRole) ValidateCreate(ctx context.Context, obj *NomadSecretEngineRole) (admission.Warnings, error) {
	nomadsecretenginerolelog.Info("validate create", "name", obj.Name)
	return nil, nil
}

func (r *NomadSecretEngineRole) ValidateUpdate(ctx context.Context, oldObj, newObj *NomadSecretEngineRole) (admission.Warnings, error) {
	nomadsecretenginerolelog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	return nil, nil
}

func (r *NomadSecretEngineRole) ValidateDelete(ctx context.Context, obj *NomadSecretEngineRole) (admission.Warnings, error) {
	nomadsecretenginerolelog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
