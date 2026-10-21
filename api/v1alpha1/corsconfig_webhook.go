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
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var corsconfiglog = logf.Log.WithName("corsconfig-resource")

func (r *CORSConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-corsconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=corsconfigs,verbs=create,versions=v1alpha1,name=mcorsconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*CORSConfig] = &CORSConfig{}

func (r *CORSConfig) Default(ctx context.Context, obj *CORSConfig) error {
	corsconfiglog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-corsconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=corsconfigs,verbs=create;update,versions=v1alpha1,name=vcorsconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*CORSConfig] = &CORSConfig{}

func (r *CORSConfig) ValidateCreate(ctx context.Context, obj *CORSConfig) (admission.Warnings, error) {
	corsconfiglog.Info("validate create", "name", obj.Name)

	if len(obj.Spec.AllowedOrigins) == 0 {
		return nil, fmt.Errorf("allowedOrigins must not be empty")
	}

	return nil, nil
}

func (r *CORSConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *CORSConfig) (admission.Warnings, error) {
	corsconfiglog.Info("validate update", "name", newObj.Name)

	if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
		return nil, err
	}

	if len(newObj.Spec.AllowedOrigins) == 0 {
		return nil, fmt.Errorf("allowedOrigins must not be empty")
	}

	return nil, nil
}

func (r *CORSConfig) ValidateDelete(ctx context.Context, obj *CORSConfig) (admission.Warnings, error) {
	corsconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
