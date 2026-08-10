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

var gcpsecretengineconfiglog = logf.Log.WithName("gcpsecretengineconfig-resource")

func (r *GCPSecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-gcpsecretengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=gcpsecretengineconfigs,verbs=create,versions=v1alpha1,name=mgcpsecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*GCPSecretEngineConfig] = &GCPSecretEngineConfig{}

func (r *GCPSecretEngineConfig) Default(ctx context.Context, obj *GCPSecretEngineConfig) error {
	gcpsecretengineconfiglog.Info("default", "name", obj.Name)
	if obj.Spec.GCPCredentials.PasswordKey == "" || obj.Spec.GCPCredentials.PasswordKey == "password" {
		obj.Spec.GCPCredentials.PasswordKey = "credentials"
	}
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-gcpsecretengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=gcpsecretengineconfigs,verbs=create;update,versions=v1alpha1,name=vgcpsecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*GCPSecretEngineConfig] = &GCPSecretEngineConfig{}

func (r *GCPSecretEngineConfig) ValidateCreate(ctx context.Context, obj *GCPSecretEngineConfig) (admission.Warnings, error) {
	gcpsecretengineconfiglog.Info("validate create", "name", obj.Name)
	return nil, obj.isValid()
}

func (r *GCPSecretEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *GCPSecretEngineConfig) (admission.Warnings, error) {
	gcpsecretengineconfiglog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	return nil, newObj.isValid()
}

func (r *GCPSecretEngineConfig) ValidateDelete(ctx context.Context, obj *GCPSecretEngineConfig) (admission.Warnings, error) {
	gcpsecretengineconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
