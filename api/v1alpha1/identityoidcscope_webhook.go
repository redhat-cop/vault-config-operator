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
var identityoidcscopelog = logf.Log.WithName("identityoidcscope-resource")

func (r *IdentityOIDCScope) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &IdentityOIDCScope{}).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-identityoidcscope,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=identityoidcscopes,verbs=create;update,versions=v1alpha1,name=midentityoidcscope.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*IdentityOIDCScope] = &IdentityOIDCScope{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *IdentityOIDCScope) Default(ctx context.Context, obj *IdentityOIDCScope) error {
	identityoidcscopelog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-identityoidcscope,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=identityoidcscopes,verbs=create;update,versions=v1alpha1,name=videntityoidcscope.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*IdentityOIDCScope] = &IdentityOIDCScope{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityOIDCScope) ValidateCreate(ctx context.Context, obj *IdentityOIDCScope) (admission.Warnings, error) {
	identityoidcscopelog.Info("validate create", "name", obj.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityOIDCScope) ValidateUpdate(ctx context.Context, oldObj, newObj *IdentityOIDCScope) (admission.Warnings, error) {
	identityoidcscopelog.Info("validate update", "name", newObj.Name)
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityOIDCScope) ValidateDelete(ctx context.Context, obj *IdentityOIDCScope) (admission.Warnings, error) {
	identityoidcscopelog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
