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

var approleauthenginerolelog = logf.Log.WithName("approleauthenginerole-resource")

func (r *AppRoleAuthEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-approleauthenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=approleauthengineroles,verbs=create,versions=v1alpha1,name=mapproleauthenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*AppRoleAuthEngineRole] = &AppRoleAuthEngineRole{}

func (r *AppRoleAuthEngineRole) Default(ctx context.Context, obj *AppRoleAuthEngineRole) error {
	approleauthenginerolelog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-approleauthenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=approleauthengineroles,verbs=create;update,versions=v1alpha1,name=vapproleauthenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*AppRoleAuthEngineRole] = &AppRoleAuthEngineRole{}

func (r *AppRoleAuthEngineRole) ValidateCreate(ctx context.Context, obj *AppRoleAuthEngineRole) (admission.Warnings, error) {
	approleauthenginerolelog.Info("validate create", "name", obj.Name)
	if _, err := obj.IsValid(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *AppRoleAuthEngineRole) ValidateUpdate(ctx context.Context, oldObj, newObj *AppRoleAuthEngineRole) (admission.Warnings, error) {
	approleauthenginerolelog.Info("validate update", "name", newObj.Name)

	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	if newObj.Spec.LocalSecretIDs != oldObj.Spec.LocalSecretIDs {
		return nil, errors.New("spec.localSecretIDs cannot be updated")
	}
	if _, err := newObj.IsValid(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *AppRoleAuthEngineRole) ValidateDelete(ctx context.Context, obj *AppRoleAuthEngineRole) (admission.Warnings, error) {
	approleauthenginerolelog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
