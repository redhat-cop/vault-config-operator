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
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var awsauthenginerolelog = logf.Log.WithName("awsauthenginerole-resource")

func (r *AWSAuthEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-awsauthenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=awsauthengineroles,verbs=create,versions=v1alpha1,name=mawsauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*AWSAuthEngineRole] = &AWSAuthEngineRole{}

func (r *AWSAuthEngineRole) Default(ctx context.Context, obj *AWSAuthEngineRole) error {
	awsauthenginerolelog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-awsauthenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=awsauthengineroles,verbs=create;update,versions=v1alpha1,name=vawsauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*AWSAuthEngineRole] = &AWSAuthEngineRole{}

func (r *AWSAuthEngineRole) ValidateCreate(ctx context.Context, obj *AWSAuthEngineRole) (admission.Warnings, error) {
	awsauthenginerolelog.Info("validate create", "name", obj.Name)
	return nil, validateAWSAuthRoleSpec(&obj.Spec.AWSAuthRole)
}

func (r *AWSAuthEngineRole) ValidateUpdate(ctx context.Context, oldObj, newObj *AWSAuthEngineRole) (admission.Warnings, error) {
	awsauthenginerolelog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	return nil, validateAWSAuthRoleSpec(&newObj.Spec.AWSAuthRole)
}

func (r *AWSAuthEngineRole) ValidateDelete(ctx context.Context, obj *AWSAuthEngineRole) (admission.Warnings, error) {
	awsauthenginerolelog.Info("validate delete", "name", obj.Name)
	return nil, nil
}

func validateAWSAuthRoleSpec(role *AWSAuthRole) error {
	if role.AuthType == "ec2" {
		if len(role.BoundIAMPrincipalARN) > 0 {
			return fmt.Errorf("boundIAMPrincipalARN is not valid for auth_type ec2")
		}
		if role.InferredEntityType != "" {
			return fmt.Errorf("inferredEntityType is not valid for auth_type ec2")
		}
		if role.InferredAWSRegion != "" {
			return fmt.Errorf("inferredAWSRegion is not valid for auth_type ec2")
		}
	}
	if role.AuthType == "iam" {
		if role.RoleTag != "" {
			return fmt.Errorf("roleTag is not valid for auth_type iam")
		}
		if role.AllowInstanceMigration {
			return fmt.Errorf("allowInstanceMigration is not valid for auth_type iam")
		}
		if role.DisallowReauthentication {
			return fmt.Errorf("disallowReauthentication is not valid for auth_type iam")
		}
		if role.InferredEntityType != "" {
			if role.InferredEntityType != "ec2_instance" {
				return fmt.Errorf("inferredEntityType must be \"ec2_instance\" when set")
			}
			if role.InferredAWSRegion == "" {
				return fmt.Errorf("inferredAWSRegion is required when inferredEntityType is set")
			}
		} else {
			if role.InferredAWSRegion != "" {
				return fmt.Errorf("inferredAWSRegion is only valid when inferredEntityType is set")
			}
			// boundAccountID is intentionally omitted: Vault accepts bound_account_id
			// for all role types (ec2, iam, and iam+inferred). The Vault API does not
			// reject it for pure IAM roles, so the operator should not be stricter than
			// Vault itself. See https://developer.hashicorp.com/vault/api-docs/auth/aws
			ec2OnlyBounds := []struct {
				name string
				val  []string
			}{
				{"boundAmiID", role.BoundAmiID},
				{"boundRegion", role.BoundRegion},
				{"boundVpcID", role.BoundVpcID},
				{"boundSubnetID", role.BoundSubnetID},
				{"boundIAMRoleARN", role.BoundIAMRoleARN},
				{"boundIAMInstanceProfileARN", role.BoundIAMInstanceProfileARN},
				{"boundEC2InstanceID", role.BoundEC2InstanceID},
			}
			for _, b := range ec2OnlyBounds {
				if len(b.val) > 0 {
					return fmt.Errorf("%s is not valid for auth_type iam without inferred_entity_type", b.name)
				}
			}
		}
	}
	if role.AllowInstanceMigration && role.DisallowReauthentication {
		return fmt.Errorf("allowInstanceMigration and disallowReauthentication are mutually exclusive")
	}
	return nil
}
