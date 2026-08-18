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

var awsauthengineclientconfiglog = logf.Log.WithName("awsauthengineclientconfig-resource")

func (r *AWSAuthEngineClientConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-awsauthengineclientconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=awsauthengineclientconfigs,verbs=create,versions=v1alpha1,name=mawsauthengineclientconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*AWSAuthEngineClientConfig] = &AWSAuthEngineClientConfig{}

func (r *AWSAuthEngineClientConfig) Default(ctx context.Context, obj *AWSAuthEngineClientConfig) error {
	awsauthengineclientconfiglog.Info("default", "name", obj.Name)

	usernameOmitted, hasRequest := awsCredentialKeyOmitted(ctx, "usernameKey")
	passwordOmitted, _ := awsCredentialKeyOmitted(ctx, "passwordKey")
	if hasRequest {
		if usernameOmitted {
			obj.Spec.AWSCredentials.UsernameKey = "access_key"
		}
		if passwordOmitted {
			obj.Spec.AWSCredentials.PasswordKey = "secret_key"
		}
		return nil
	}

	// No admission request (unit tests / unexpected): apply AWS defaults only for
	// empty values or the inherited RootCredentialConfig kubebuilder defaults.
	if obj.Spec.AWSCredentials.UsernameKey == "" || obj.Spec.AWSCredentials.UsernameKey == "username" {
		obj.Spec.AWSCredentials.UsernameKey = "access_key"
	}
	if obj.Spec.AWSCredentials.PasswordKey == "" || obj.Spec.AWSCredentials.PasswordKey == "password" {
		obj.Spec.AWSCredentials.PasswordKey = "secret_key"
	}
	return nil
}

// awsCredentialKeyOmitted reports whether spec.AWSCredentials.<key> was absent from
// the admission request body. When the user omitted the key, the CRD may still
// populate the inherited RootCredentialConfig default ("username"/"password");
// those must be remapped to access_key/secret_key. When the user explicitly sent
// the key, even if the value is "username" or "password", it must be preserved.
func awsCredentialKeyOmitted(ctx context.Context, key string) (omitted bool, hasRequest bool) {
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
	creds, _ := spec["AWSCredentials"].(map[string]any)
	if creds == nil {
		return true, true
	}
	_, present := creds[key]
	return !present, true
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-awsauthengineclientconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=awsauthengineclientconfigs,verbs=create;update,versions=v1alpha1,name=vawsauthengineclientconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*AWSAuthEngineClientConfig] = &AWSAuthEngineClientConfig{}

func (r *AWSAuthEngineClientConfig) ValidateCreate(ctx context.Context, obj *AWSAuthEngineClientConfig) (admission.Warnings, error) {
	awsauthengineclientconfiglog.Info("validate create", "name", obj.Name)
	return nil, obj.isValid()
}

func (r *AWSAuthEngineClientConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *AWSAuthEngineClientConfig) (admission.Warnings, error) {
	awsauthengineclientconfiglog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, newObj.isValid()
}

func (r *AWSAuthEngineClientConfig) ValidateDelete(ctx context.Context, obj *AWSAuthEngineClientConfig) (admission.Warnings, error) {
	awsauthengineclientconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
