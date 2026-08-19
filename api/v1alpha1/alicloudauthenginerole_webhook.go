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

var alicloudauthenginerolelog = logf.Log.WithName("alicloudauthenginerole-resource")

func (r *AliCloudAuthEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-alicloudauthenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=alicloudauthengineroles,verbs=create,versions=v1alpha1,name=malicloudauthenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*AliCloudAuthEngineRole] = &AliCloudAuthEngineRole{}

func (r *AliCloudAuthEngineRole) Default(ctx context.Context, obj *AliCloudAuthEngineRole) error {
	alicloudauthenginerolelog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-alicloudauthenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=alicloudauthengineroles,verbs=create;update,versions=v1alpha1,name=validalicloudauthenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*AliCloudAuthEngineRole] = &AliCloudAuthEngineRole{}

func (r *AliCloudAuthEngineRole) ValidateCreate(ctx context.Context, obj *AliCloudAuthEngineRole) (admission.Warnings, error) {
	alicloudauthenginerolelog.Info("validate create", "name", obj.Name)
	if _, err := obj.IsValid(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *AliCloudAuthEngineRole) ValidateUpdate(ctx context.Context, oldObj, newObj *AliCloudAuthEngineRole) (admission.Warnings, error) {
	alicloudauthenginerolelog.Info("validate update", "name", newObj.Name)

	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	if _, err := newObj.IsValid(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *AliCloudAuthEngineRole) ValidateDelete(ctx context.Context, obj *AliCloudAuthEngineRole) (admission.Warnings, error) {
	alicloudauthenginerolelog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
