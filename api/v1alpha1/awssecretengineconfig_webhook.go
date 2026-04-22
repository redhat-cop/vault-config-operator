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
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var awssecretengineconfiglog = logf.Log.WithName("awssecretengineconfig-resource")

func (r *AWSSecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-AWSSecretEngineConfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=AWSSecretEngineConfigs,verbs=create;update,versions=v1alpha1,name=mawssecretengineconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &AWSSecretEngineConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AWSSecretEngineConfig) Default() {
	awssecretengineconfiglog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-AWSSecretEngineConfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=AWSSecretEngineConfigs,verbs=create;update,versions=v1alpha1,name=vawssecretengineconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AWSSecretEngineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecretEngineConfig) ValidateCreate() (admission.Warnings, error) {
	awssecretengineconfiglog.Info("validate create", "name", r.Name)

	return nil, r.validateAWSSecretEngineConfig()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecretEngineConfig) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	awssecretengineconfiglog.Info("validate update", "name", r.Name)

	if r.Spec.Path != old.(*AWSSecretEngineConfig).Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}

	return nil, r.validateAWSSecretEngineConfig()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecretEngineConfig) ValidateDelete() (admission.Warnings, error) {
	awssecretengineconfiglog.Info("validate delete", "name", r.Name)

	return nil, nil
}

// validateAWSSecretEngineConfig validates the AWS Secret Engine Config spec
func (r *AWSSecretEngineConfig) validateAWSSecretEngineConfig() error {
	spec := &r.Spec.AWSSEConfig

	// Validate mutually exclusive credential options
	hasAccessKey := r.Spec.AWSCredentials.Secret != nil || r.Spec.AWSCredentials.RandomSecret != nil || r.Spec.AWSCredentials.VaultSecret != nil
	hasIdentityTokenAudience := spec.IdentityTokenAudience != ""

	if hasAccessKey && hasIdentityTokenAudience {
		return errors.New("spec.awsCredentials and spec.identityTokenAudience are mutually exclusive")
	}

	// If using identity token federation, role_arn is required
	if hasIdentityTokenAudience && spec.RoleARN == "" {
		return errors.New("spec.roleArn is required when spec.identityTokenAudience is specified")
	}

	// Validate rotation settings
	hasRotationPeriod := spec.RotationPeriod != 0
	hasRotationSchedule := spec.RotationSchedule != ""

	if hasRotationPeriod && hasRotationSchedule {
		return errors.New("cannot set both spec.rotationPeriod and spec.rotationSchedule")
	}

	if hasRotationPeriod && spec.RotationPeriod < 10 {
		return errors.New("spec.rotationPeriod must be at least 10 seconds")
	}

	if spec.RotationWindow != 0 && spec.RotationWindow < 3600 {
		return errors.New("spec.rotationWindow must be at least 1 hour (3600 seconds)")
	}

	if spec.RotationWindow != 0 && hasRotationPeriod {
		return errors.New("cannot set spec.rotationWindow when using spec.rotationPeriod")
	}

	// Validate STS fallback endpoints and regions match in length
	if len(spec.STSFallbackEndpoints) != len(spec.STSFallbackRegions) {
		return errors.New("spec.stsFallbackEndpoints and spec.stsFallbackRegions must have the same length")
	}

	return nil
}
