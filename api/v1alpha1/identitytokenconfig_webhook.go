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

// log is for logging in this package.
var identitytokenconfiglog = logf.Log.WithName("identitytokenconfig-resource")

func (r *IdentityTokenConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &IdentityTokenConfig{}).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-identitytokenconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=identitytokenconfigs,verbs=create;update,versions=v1alpha1,name=midentitytokenconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*IdentityTokenConfig] = &IdentityTokenConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *IdentityTokenConfig) Default(ctx context.Context, obj *IdentityTokenConfig) error {
	identitytokenconfiglog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-identitytokenconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=identitytokenconfigs,verbs=create;update;delete,versions=v1alpha1,name=videntitytokenconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*IdentityTokenConfig] = &IdentityTokenConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityTokenConfig) ValidateCreate(ctx context.Context, obj *IdentityTokenConfig) (admission.Warnings, error) {
	identitytokenconfiglog.Info("validate create", "name", obj.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityTokenConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *IdentityTokenConfig) (admission.Warnings, error) {
	identitytokenconfiglog.Info("validate update", "name", newObj.Name)
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityTokenConfig) ValidateDelete(ctx context.Context, obj *IdentityTokenConfig) (admission.Warnings, error) {
	identitytokenconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
