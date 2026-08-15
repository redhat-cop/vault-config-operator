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

var totpsecretenginekeylog = logf.Log.WithName("totpsecretenginekey-resource")

func (r *TOTPSecretEngineKey) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-totpsecretenginekey,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=totpsecretenginekeys,verbs=create,versions=v1alpha1,name=mtotpsecretenginekey.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*TOTPSecretEngineKey] = &TOTPSecretEngineKey{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *TOTPSecretEngineKey) Default(ctx context.Context, obj *TOTPSecretEngineKey) error {
	totpsecretenginekeylog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-totpsecretenginekey,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=totpsecretenginekeys,verbs=create;update,versions=v1alpha1,name=vtotpsecretenginekey.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*TOTPSecretEngineKey] = &TOTPSecretEngineKey{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *TOTPSecretEngineKey) ValidateCreate(ctx context.Context, obj *TOTPSecretEngineKey) (admission.Warnings, error) {
	totpsecretenginekeylog.Info("validate create", "name", obj.Name)
	return nil, obj.isValid()
}

func (r *TOTPSecretEngineKey) isValid() error {
	if r.Spec.Issuer == "" {
		return errors.New("spec.issuer is required to ensure drift detection works correctly")
	}
	if r.Spec.AccountName == "" {
		return errors.New("spec.accountName is required to ensure drift detection works correctly")
	}
	if r.Spec.Generate {
		if r.Spec.Key != "" {
			return errors.New("spec.key cannot be set when spec.generate is true")
		}
		if r.Spec.URL != "" {
			return errors.New("spec.url cannot be set when spec.generate is true")
		}
	} else {
		if r.Spec.Key == "" && r.Spec.URL == "" {
			return errors.New("one of spec.key or spec.url is required when spec.generate is false")
		}
		if r.Spec.Exported != nil {
			return errors.New("spec.exported cannot be set when spec.generate is false")
		}
		if r.Spec.KeySize != 0 {
			return errors.New("spec.keySize cannot be set when spec.generate is false")
		}
	}
	return nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *TOTPSecretEngineKey) ValidateUpdate(ctx context.Context, oldObj, newObj *TOTPSecretEngineKey) (admission.Warnings, error) {
	totpsecretenginekeylog.Info("validate update", "name", newObj.Name)

	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	if err := validateTOTPImmutableFields(oldObj, newObj); err != nil {
		return nil, err
	}
	return nil, newObj.isValid()
}

func validateTOTPImmutableFields(oldObj, newObj *TOTPSecretEngineKey) error {
	if newObj.Spec.Generate != oldObj.Spec.Generate {
		return fmt.Errorf("spec.generate cannot be updated after key creation")
	}
	if newObj.Spec.Key != oldObj.Spec.Key {
		return fmt.Errorf("spec.key cannot be updated after key creation")
	}
	if newObj.Spec.URL != oldObj.Spec.URL {
		return fmt.Errorf("spec.url cannot be updated after key creation")
	}
	if newObj.Spec.KeySize != oldObj.Spec.KeySize {
		return fmt.Errorf("spec.keySize cannot be updated after key creation")
	}
	if !boolPtrEqual(newObj.Spec.Exported, oldObj.Spec.Exported) {
		return fmt.Errorf("spec.exported cannot be updated after key creation")
	}
	if !intPtrEqual(newObj.Spec.Skew, oldObj.Spec.Skew) {
		return fmt.Errorf("spec.skew cannot be updated after key creation")
	}
	if !intPtrEqual(newObj.Spec.QRSize, oldObj.Spec.QRSize) {
		return fmt.Errorf("spec.qrSize cannot be updated after key creation")
	}
	return nil
}

func boolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *TOTPSecretEngineKey) ValidateDelete(ctx context.Context, obj *TOTPSecretEngineKey) (admission.Warnings, error) {
	totpsecretenginekeylog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
