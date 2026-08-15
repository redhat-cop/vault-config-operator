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

var terraformcloudsecretenginerolelog = logf.Log.WithName("terraformcloudsecretenginerole-resource")

func (r *TerraformCloudSecretEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-terraformcloudsecretenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=terraformcloudsecretengineroles,verbs=create,versions=v1alpha1,name=mterraformcloudsecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*TerraformCloudSecretEngineRole] = &TerraformCloudSecretEngineRole{}

func (r *TerraformCloudSecretEngineRole) Default(ctx context.Context, obj *TerraformCloudSecretEngineRole) error {
	terraformcloudsecretenginerolelog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-terraformcloudsecretenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=terraformcloudsecretengineroles,verbs=create;update,versions=v1alpha1,name=vterraformcloudsecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*TerraformCloudSecretEngineRole] = &TerraformCloudSecretEngineRole{}

func (r *TerraformCloudSecretEngineRole) ValidateCreate(ctx context.Context, obj *TerraformCloudSecretEngineRole) (admission.Warnings, error) {
	terraformcloudsecretenginerolelog.Info("validate create", "name", obj.Name)
	return nil, obj.isValid()
}

func (r *TerraformCloudSecretEngineRole) ValidateUpdate(ctx context.Context, oldObj, newObj *TerraformCloudSecretEngineRole) (admission.Warnings, error) {
	terraformcloudsecretenginerolelog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	return nil, newObj.isValid()
}

func (r *TerraformCloudSecretEngineRole) ValidateDelete(ctx context.Context, obj *TerraformCloudSecretEngineRole) (admission.Warnings, error) {
	terraformcloudsecretenginerolelog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
