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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var quotaconfiglog = logf.Log.WithName("quotaconfig-resource")

func (r *QuotaConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-quotaconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=quotaconfigs,verbs=create,versions=v1alpha1,name=mquotaconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*QuotaConfig] = &QuotaConfig{}

func (r *QuotaConfig) Default(ctx context.Context, obj *QuotaConfig) error {
	quotaconfiglog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-quotaconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=quotaconfigs,verbs=create;update,versions=v1alpha1,name=vquotaconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*QuotaConfig] = &QuotaConfig{}

func (r *QuotaConfig) ValidateCreate(ctx context.Context, obj *QuotaConfig) (admission.Warnings, error) {
	quotaconfiglog.Info("validate create", "name", obj.Name)
	return nil, nil
}

func (r *QuotaConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *QuotaConfig) (admission.Warnings, error) {
	quotaconfiglog.Info("validate update", "name", newObj.Name)

	if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *QuotaConfig) ValidateDelete(ctx context.Context, obj *QuotaConfig) (admission.Warnings, error) {
	quotaconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
