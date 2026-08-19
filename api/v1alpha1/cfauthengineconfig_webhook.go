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

var cfauthengineconfiglog = logf.Log.WithName("cfauthengineconfig-resource")

func (r *CFAuthEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-cfauthengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=cfauthengineconfigs,verbs=create;update,versions=v1alpha1,name=mcfauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*CFAuthEngineConfig] = &CFAuthEngineConfig{}

func (r *CFAuthEngineConfig) Default(ctx context.Context, obj *CFAuthEngineConfig) error {
	cfauthengineconfiglog.Info("default", "name", obj.Name)

	usernameOmitted, hasRequest := cfCredentialKeyOmitted(ctx, "usernameKey")
	passwordOmitted, _ := cfCredentialKeyOmitted(ctx, "passwordKey")
	if hasRequest {
		if usernameOmitted {
			obj.Spec.CFCredentials.UsernameKey = "cf_username"
		}
		if passwordOmitted {
			obj.Spec.CFCredentials.PasswordKey = "cf_password"
		}
		return nil
	}

	if obj.Spec.CFCredentials.UsernameKey == "" || obj.Spec.CFCredentials.UsernameKey == "username" {
		obj.Spec.CFCredentials.UsernameKey = "cf_username"
	}
	if obj.Spec.CFCredentials.PasswordKey == "" || obj.Spec.CFCredentials.PasswordKey == "password" {
		obj.Spec.CFCredentials.PasswordKey = "cf_password"
	}
	return nil
}

// cfCredentialKeyOmitted reports whether spec.cfCredentials.<key> should be treated
// as omitted from the admission request body for remapping purposes. A key is
// considered omitted when it is absent OR when it is present with an empty-string
// value (explicit empty means "use the default"). When the user omitted the key, the
// CRD may still populate the inherited RootCredentialConfig default ("username"/
// "password"); those must be remapped to cf_username/cf_password.
func cfCredentialKeyOmitted(ctx context.Context, key string) (omitted bool, hasRequest bool) {
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
	creds, _ := spec["cfCredentials"].(map[string]any)
	if creds == nil {
		return true, true
	}
	val, present := creds[key]
	if !present {
		return true, true
	}
	strVal, isStr := val.(string)
	if isStr && strVal == "" {
		return true, true
	}
	return false, true
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-cfauthengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=cfauthengineconfigs,verbs=create;update,versions=v1alpha1,name=vcfauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*CFAuthEngineConfig] = &CFAuthEngineConfig{}

func (r *CFAuthEngineConfig) ValidateCreate(ctx context.Context, obj *CFAuthEngineConfig) (admission.Warnings, error) {
	cfauthengineconfiglog.Info("validate create", "name", obj.Name)
	return nil, obj.isValid()
}

func (r *CFAuthEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *CFAuthEngineConfig) (admission.Warnings, error) {
	cfauthengineconfiglog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, newObj.isValid()
}

func (r *CFAuthEngineConfig) ValidateDelete(ctx context.Context, obj *CFAuthEngineConfig) (admission.Warnings, error) {
	cfauthengineconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
