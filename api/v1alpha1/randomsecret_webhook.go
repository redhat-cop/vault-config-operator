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

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var randomsecretlog = logf.Log.WithName("randomsecret-resource")

func (r *RandomSecret) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-randomsecret,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=randomsecrets,verbs=create;update,versions=v1alpha1,name=vrandomsecret.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &RandomSecret{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *RandomSecret) ValidateCreate() error {
	randomsecretlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *RandomSecret) ValidateUpdate(old runtime.Object) error {
	randomsecretlog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*RandomSecret).Spec.Path {
		return errors.New("spec.path cannot be updated")
	}

	// TODO(user): fill in your validation logic upon object update.
	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *RandomSecret) ValidateDelete() error {
	randomsecretlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *RandomSecret) validate() error {
	result := &multierror.Error{}
	result = multierror.Append(result, r.validateEitherPasswordPolicyReferenceOrInline())
	result = multierror.Append(result, r.validateInlinePasswordPolicyFormat())
	return result.ErrorOrNil()
}

func (r *RandomSecret) validateEitherPasswordPolicyReferenceOrInline() error {
	result := &multierror.Error{}
	for key, value := range r.Spec.SecretFormat {
		count := 0
		if value.InlinePasswordPolicy != "" {
			count++
		}
		if value.PasswordPolicyName != "" {
			count++
		}
		if count > 1 {
			result = multierror.Append(result, errors.New("only one of InlinePasswordPolicy or PasswordPolicyName can be defined in key: "+key))
		}
	}
	return result.ErrorOrNil()
}

func (r *RandomSecret) validateInlinePasswordPolicyFormat() error {
	result := &multierror.Error{}
	for key, value := range r.Spec.SecretFormat {
		if value.InlinePasswordPolicy != "" {
			passwordPolicyFormat := &PasswordPolicyFormat{}
			result = multierror.Append(result, hclsimple.Decode(key, []byte(value.InlinePasswordPolicy), nil, passwordPolicyFormat))
		}
	}
	return result.ErrorOrNil()
}
