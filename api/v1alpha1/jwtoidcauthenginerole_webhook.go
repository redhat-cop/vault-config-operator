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
var jwtoidcauthenginerolelog = logf.Log.WithName("jwtoidcauthenginerole-resource")

func (r *JWTOIDCAuthEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-jwtoidcauthenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=jwtoidcauthengineroles,verbs=create,versions=v1alpha1,name=mjwtoidcauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*JWTOIDCAuthEngineRole] = &JWTOIDCAuthEngineRole{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *JWTOIDCAuthEngineRole) Default(ctx context.Context, obj *JWTOIDCAuthEngineRole) error {
	jwtoidcauthenginerolelog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-jwtoidcauthenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=jwtoidcauthengineroles,verbs=update,versions=v1alpha1,name=vjwtoidcauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*JWTOIDCAuthEngineRole] = &JWTOIDCAuthEngineRole{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *JWTOIDCAuthEngineRole) ValidateCreate(ctx context.Context, obj *JWTOIDCAuthEngineRole) (admission.Warnings, error) {
	jwtoidcauthenginerolelog.Info("validate create", "name", obj.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *JWTOIDCAuthEngineRole) ValidateUpdate(ctx context.Context, oldObj, newObj *JWTOIDCAuthEngineRole) (admission.Warnings, error) {
	jwtoidcauthenginerolelog.Info("validate update", "name", newObj.Name)

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *JWTOIDCAuthEngineRole) ValidateDelete(ctx context.Context, obj *JWTOIDCAuthEngineRole) (admission.Warnings, error) {
	jwtoidcauthenginerolelog.Info("validate delete", "name", obj.Name)

	return nil, nil
}
