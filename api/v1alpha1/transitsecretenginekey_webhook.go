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
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var transitsecretenginekeylog = logf.Log.WithName("transitsecretenginekey-resource")

func (r *TransitSecretEngineKey) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-transitsecretenginekey,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=transitsecretenginekeys,verbs=create,versions=v1alpha1,name=mtransitsecretenginekey.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*TransitSecretEngineKey] = &TransitSecretEngineKey{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *TransitSecretEngineKey) Default(ctx context.Context, obj *TransitSecretEngineKey) error {
	transitsecretenginekeylog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-transitsecretenginekey,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=transitsecretenginekeys,verbs=create;update,versions=v1alpha1,name=vtransitsecretenginekey.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*TransitSecretEngineKey] = &TransitSecretEngineKey{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *TransitSecretEngineKey) ValidateCreate(ctx context.Context, obj *TransitSecretEngineKey) (admission.Warnings, error) {
	transitsecretenginekeylog.Info("validate create", "name", obj.Name)

	if obj.Spec.ConvergentEncryption && !obj.Spec.Derived {
		return nil, errors.New("spec.convergentEncryption requires spec.derived to be true")
	}
	return nil, validateTransitKeyConstraints(&obj.Spec.TransitKeyConfig)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *TransitSecretEngineKey) ValidateUpdate(ctx context.Context, oldObj, newObj *TransitSecretEngineKey) (admission.Warnings, error) {
	transitsecretenginekeylog.Info("validate update", "name", newObj.Name)

	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	if newObj.Spec.Type != oldObj.Spec.Type {
		return nil, errors.New("spec.type cannot be updated after key creation")
	}
	if newObj.Spec.Derived != oldObj.Spec.Derived {
		return nil, errors.New("spec.derived cannot be updated after key creation")
	}
	if newObj.Spec.ConvergentEncryption != oldObj.Spec.ConvergentEncryption {
		return nil, errors.New("spec.convergentEncryption cannot be updated after key creation")
	}
	if newObj.Spec.KeySize != oldObj.Spec.KeySize {
		return nil, errors.New("spec.keySize cannot be updated after key creation")
	}
	if oldObj.Spec.Exportable && !newObj.Spec.Exportable {
		return nil, errors.New("spec.exportable cannot be changed from true to false (one-way flag)")
	}
	if oldObj.Spec.AllowPlaintextBackup && !newObj.Spec.AllowPlaintextBackup {
		return nil, errors.New("spec.allowPlaintextBackup cannot be changed from true to false (one-way flag)")
	}
	if newObj.Spec.ConvergentEncryption && !newObj.Spec.Derived {
		return nil, errors.New("spec.convergentEncryption requires spec.derived to be true")
	}
	return nil, validateTransitKeyConstraints(&newObj.Spec.TransitKeyConfig)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *TransitSecretEngineKey) ValidateDelete(ctx context.Context, obj *TransitSecretEngineKey) (admission.Warnings, error) {
	transitsecretenginekeylog.Info("validate delete", "name", obj.Name)
	return nil, nil
}

const minAutoRotatePeriod = time.Hour

func validateTransitKeyConstraints(cfg *TransitKeyConfig) error {
	if cfg.KeySize > 0 && cfg.Type != "hmac" {
		return fmt.Errorf("spec.keySize is only applicable to hmac key type, got type %q", cfg.Type)
	}
	if cfg.KeySize > 0 && cfg.Type == "hmac" && cfg.KeySize < 32 {
		return fmt.Errorf("spec.keySize must be at least 32 bytes for hmac keys, got %d", cfg.KeySize)
	}
	if cfg.AutoRotatePeriod != "" && cfg.AutoRotatePeriod != "0" {
		d, err := time.ParseDuration(cfg.AutoRotatePeriod)
		if err != nil {
			return fmt.Errorf("spec.autoRotatePeriod is not a valid duration: %w", err)
		}
		if d < minAutoRotatePeriod {
			return fmt.Errorf("spec.autoRotatePeriod must be at least %s when enabled, got %s", minAutoRotatePeriod, d)
		}
	}
	return nil
}
