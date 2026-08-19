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
	"encoding/json"
	"errors"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var radiusauthengineconfiglog = logf.Log.WithName("radiusauthengineconfig-resource")

func (r *RADIUSAuthEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-radiusauthengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=radiusauthengineconfigs,verbs=create;update,versions=v1alpha1,name=mradiusauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*RADIUSAuthEngineConfig] = &RADIUSAuthEngineConfig{}

func (r *RADIUSAuthEngineConfig) Default(ctx context.Context, obj *RADIUSAuthEngineConfig) error {
	radiusauthengineconfiglog.Info("default", "name", obj.Name)

	passwordOmitted, hasRequest := radiusCredentialKeyOmitted(ctx, "passwordKey")
	if hasRequest {
		if passwordOmitted {
			obj.Spec.RADIUSCredentials.PasswordKey = "secret"
		}
		return nil
	}

	// No admission request (unit tests / unexpected): apply RADIUS defaults only for
	// empty values or the inherited RootCredentialConfig kubebuilder default.
	if obj.Spec.RADIUSCredentials.PasswordKey == "" || obj.Spec.RADIUSCredentials.PasswordKey == "password" {
		obj.Spec.RADIUSCredentials.PasswordKey = "secret"
	}
	return nil
}

// radiusCredentialKeyOmitted reports whether spec.radiusCredentials.<key> was absent from
// the admission request body. When the user omitted the key, the CRD schema default
// ("password") may still be populated; it must be remapped to "secret" for RADIUS.
// When the user explicitly sent the key, even if the value equals the schema default,
// it must be preserved.
func radiusCredentialKeyOmitted(ctx context.Context, key string) (omitted bool, hasRequest bool) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return false, false
	}
	var raw map[string]any
	if err := json.Unmarshal(req.Object.Raw, &raw); err != nil {
		return false, false
	}
	spec, _ := raw["spec"].(map[string]any)
	if spec == nil {
		return true, true
	}
	creds, _ := spec["radiusCredentials"].(map[string]any)
	if creds == nil {
		return true, true
	}
	_, present := creds[key]
	return !present, true
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-radiusauthengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=radiusauthengineconfigs,verbs=create;update,versions=v1alpha1,name=vradiusauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*RADIUSAuthEngineConfig] = &RADIUSAuthEngineConfig{}

func (r *RADIUSAuthEngineConfig) ValidateCreate(ctx context.Context, obj *RADIUSAuthEngineConfig) (admission.Warnings, error) {
	radiusauthengineconfiglog.Info("validate create", "name", obj.Name)
	return nil, obj.isValid()
}

func (r *RADIUSAuthEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *RADIUSAuthEngineConfig) (admission.Warnings, error) {
	radiusauthengineconfiglog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, newObj.isValid()
}

func (r *RADIUSAuthEngineConfig) ValidateDelete(ctx context.Context, obj *RADIUSAuthEngineConfig) (admission.Warnings, error) {
	radiusauthengineconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
