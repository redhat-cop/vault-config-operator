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

var gcpsecretenginerolesetlog = logf.Log.WithName("gcpsecretengineroleset-resource")

func (r *GCPSecretEngineRoleset) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-gcpsecretengineroleset,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=gcpsecretenginerolesets,verbs=create,versions=v1alpha1,name=mgcpsecretengineroleset.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*GCPSecretEngineRoleset] = &GCPSecretEngineRoleset{}

func (r *GCPSecretEngineRoleset) Default(ctx context.Context, obj *GCPSecretEngineRoleset) error {
	gcpsecretenginerolesetlog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-gcpsecretengineroleset,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=gcpsecretenginerolesets,verbs=create;update,versions=v1alpha1,name=vgcpsecretengineroleset.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*GCPSecretEngineRoleset] = &GCPSecretEngineRoleset{}

func (r *GCPSecretEngineRoleset) ValidateCreate(ctx context.Context, obj *GCPSecretEngineRoleset) (admission.Warnings, error) {
	gcpsecretenginerolesetlog.Info("validate create", "name", obj.Name)
	return nil, nil
}

func (r *GCPSecretEngineRoleset) ValidateUpdate(ctx context.Context, oldObj, newObj *GCPSecretEngineRoleset) (admission.Warnings, error) {
	gcpsecretenginerolesetlog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	if newObj.Spec.SecretType != oldObj.Spec.SecretType {
		return nil, errors.New("spec.secretType cannot be updated")
	}
	if newObj.Spec.Project != oldObj.Spec.Project {
		return nil, errors.New("spec.project cannot be updated")
	}
	return nil, nil
}

func (r *GCPSecretEngineRoleset) ValidateDelete(ctx context.Context, obj *GCPSecretEngineRoleset) (admission.Warnings, error) {
	gcpsecretenginerolesetlog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
