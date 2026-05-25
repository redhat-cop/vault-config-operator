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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var awssecretenginerolelog = logf.Log.WithName("awssecretenginerole-resource")

func (r *AWSSecretEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-awssecretenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=awssecretengineroles,verbs=create;update,versions=v1alpha1,name=mawssecretenginerole.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &AWSSecretEngineRole{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AWSSecretEngineRole) Default() {
	awssecretenginerolelog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-awssecretenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=awssecretengineroles,verbs=create;update,versions=v1alpha1,name=vawssecretenginerole.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AWSSecretEngineRole{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecretEngineRole) ValidateCreate() (admission.Warnings, error) {
	awssecretenginerolelog.Info("validate create", "name", r.Name)

	return nil, r.validateAWSSecretEngineRole()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecretEngineRole) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	awssecretenginerolelog.Info("validate update", "name", r.Name)

	if r.Spec.Path != old.(*AWSSecretEngineRole).Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}

	return nil, r.validateAWSSecretEngineRole()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AWSSecretEngineRole) ValidateDelete() (admission.Warnings, error) {
	awssecretenginerolelog.Info("validate delete", "name", r.Name)

	return nil, nil
}

// validateAWSSecretEngineRole validates the AWS Secret Engine Role spec
func (r *AWSSecretEngineRole) validateAWSSecretEngineRole() error {
	spec := &r.Spec.AWSSERole
	credType := spec.CredentialType

	// Validate credential_type is provided
	if credType == "" {
		return errors.New("spec.credentialType is required")
	}

	// Validate role_arns
	if credType == "assumed_role" {
		if len(spec.RoleARNs) == 0 {
			return errors.New("spec.roleArns is required when credentialType is 'assumed_role'")
		}
	} else {
		if len(spec.RoleARNs) > 0 {
			return fmt.Errorf("spec.roleArns is not allowed when credentialType is '%s'", credType)
		}
	}

	// Validate policy_arns and policy_document for iam_user and federation_token
	if credType == "iam_user" || credType == "federation_token" {
		if len(spec.PolicyARNs) == 0 && spec.PolicyDocument == "" {
			return fmt.Errorf("at least one of spec.policyArns or spec.policyDocument must be specified when credentialType is '%s'", credType)
		}
	}

	// Validate policy_arns, policy_document, and iam_groups are not used with session_token
	if credType == "session_token" {
		if len(spec.PolicyARNs) > 0 {
			return errors.New("spec.policyArns is not allowed when credentialType is 'session_token'")
		}
		if spec.PolicyDocument != "" {
			return errors.New("spec.policyDocument is not allowed when credentialType is 'session_token'")
		}
		if len(spec.IAMGroups) > 0 {
			return errors.New("spec.iamGroups is not allowed when credentialType is 'session_token'")
		}
	}

	// Validate session_tags
	if len(spec.SessionTags) > 0 && credType != "assumed_role" {
		return fmt.Errorf("spec.sessionTags is only valid when credentialType is 'assumed_role', got '%s'", credType)
	}

	// Validate default_sts_ttl and max_sts_ttl
	if credType != "assumed_role" && credType != "federation_token" {
		if spec.DefaultSTSTTL != "" {
			return fmt.Errorf("spec.defaultStsTtl is only valid when credentialType is 'assumed_role' or 'federation_token', got '%s'", credType)
		}
		if spec.MaxSTSTTL != "" {
			return fmt.Errorf("spec.maxStsTtl is only valid when credentialType is 'assumed_role' or 'federation_token', got '%s'", credType)
		}
	}

	// Validate external_id
	if spec.ExternalID != "" && credType != "assumed_role" {
		return fmt.Errorf("spec.externalId is only valid when credentialType is 'assumed_role', got '%s'", credType)
	}

	// Validate user_path, permissions_boundary_arn, iam_tags, and mfa_serial_number for non-iam_user credential types
	if credType != "iam_user" {
		// Validate user_path
		if spec.UserPath != "" {
			return fmt.Errorf("spec.userPath is only valid when credentialType is 'iam_user', got '%s'", credType)
		}

		// Validate permissions_boundary_arn
		if spec.PermissionsBoundaryARN != "" {
			return fmt.Errorf("spec.permissionsBoundaryArn is only valid when credentialType is 'iam_user', got '%s'", credType)
		}

		// Validate iam_tags
		if len(spec.IAMTags) > 0 {
			return fmt.Errorf("spec.iamTags is only valid when credentialType is 'iam_user', got '%s'", credType)
		}

		// Validate mfa_serial_number
		if spec.MFASerialNumber != "" {
			return fmt.Errorf("spec.mfaSerialNumber is only valid when credentialType is 'iam_user', got '%s'", credType)
		}
	}

	return nil
}
