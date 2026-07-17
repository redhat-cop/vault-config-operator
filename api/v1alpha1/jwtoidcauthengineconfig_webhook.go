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

// log is for logging in this package.
var jwtoidcauthengineconfiglog = logf.Log.WithName("jwtoidcauthengineconfig-resource")

func (r *JWTOIDCAuthEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-jwtoidcauthengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=jwtoidcauthengineconfigs,verbs=create,versions=v1alpha1,name=mjwtoidcauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*JWTOIDCAuthEngineConfig] = &JWTOIDCAuthEngineConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *JWTOIDCAuthEngineConfig) Default(ctx context.Context, obj *JWTOIDCAuthEngineConfig) error {
	jwtoidcauthengineconfiglog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-jwtoidcauthengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=jwtoidcauthengineconfigs,verbs=update,versions=v1alpha1,name=vjwtoidcauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*JWTOIDCAuthEngineConfig] = &JWTOIDCAuthEngineConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *JWTOIDCAuthEngineConfig) ValidateCreate(ctx context.Context, obj *JWTOIDCAuthEngineConfig) (admission.Warnings, error) {
	jwtoidcauthengineconfiglog.Info("validate create", "name", obj.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *JWTOIDCAuthEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *JWTOIDCAuthEngineConfig) (admission.Warnings, error) {
	jwtoidcauthengineconfiglog.Info("validate update", "name", newObj.Name)

	// the path cannot be updated
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *JWTOIDCAuthEngineConfig) ValidateDelete(ctx context.Context, obj *JWTOIDCAuthEngineConfig) (admission.Warnings, error) {
	jwtoidcauthengineconfiglog.Info("validate delete", "name", obj.Name)

	return nil, nil
}
