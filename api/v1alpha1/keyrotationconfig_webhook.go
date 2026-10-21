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
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var keyrotationconfiglog = logf.Log.WithName("keyrotationconfig-resource")

func (r *KeyRotationConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-keyrotationconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=keyrotationconfigs,verbs=create,versions=v1alpha1,name=mkeyrotationconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*KeyRotationConfig] = &KeyRotationConfig{}

func (r *KeyRotationConfig) Default(ctx context.Context, obj *KeyRotationConfig) error {
	keyrotationconfiglog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-keyrotationconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=keyrotationconfigs,verbs=create;update,versions=v1alpha1,name=vkeyrotationconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*KeyRotationConfig] = &KeyRotationConfig{}

func (r *KeyRotationConfig) ValidateCreate(ctx context.Context, obj *KeyRotationConfig) (admission.Warnings, error) {
	keyrotationconfiglog.Info("validate create", "name", obj.Name)
	return nil, validateInterval(obj.Spec.Interval)
}

func (r *KeyRotationConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *KeyRotationConfig) (admission.Warnings, error) {
	keyrotationconfiglog.Info("validate update", "name", newObj.Name)

	if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
		return nil, err
	}

	return nil, validateInterval(newObj.Spec.Interval)
}

func (r *KeyRotationConfig) ValidateDelete(ctx context.Context, obj *KeyRotationConfig) (admission.Warnings, error) {
	keyrotationconfiglog.Info("validate delete", "name", obj.Name)
	return nil, nil
}

func validateInterval(interval string) error {
	if interval == "" {
		return nil
	}
	d, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("interval %q is not a valid Go duration: %w", interval, err)
	}
	if d < 24*time.Hour {
		return fmt.Errorf("interval must be at least 24h, got %s", interval)
	}
	return nil
}
