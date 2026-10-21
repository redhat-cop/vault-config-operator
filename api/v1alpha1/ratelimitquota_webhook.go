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
	"strconv"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var ratelimitquotalog = logf.Log.WithName("ratelimitquota-resource")

func (r *RateLimitQuota) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-ratelimitquota,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=ratelimitquotas,verbs=create,versions=v1alpha1,name=mratelimitquota.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*RateLimitQuota] = &RateLimitQuota{}

func (r *RateLimitQuota) Default(ctx context.Context, obj *RateLimitQuota) error {
	ratelimitquotalog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-ratelimitquota,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=ratelimitquotas,verbs=create;update,versions=v1alpha1,name=vratelimitquota.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*RateLimitQuota] = &RateLimitQuota{}

func (r *RateLimitQuota) ValidateCreate(ctx context.Context, obj *RateLimitQuota) (admission.Warnings, error) {
	ratelimitquotalog.Info("validate create", "name", obj.Name)

	if err := validateRate(obj.Spec.Rate); err != nil {
		return nil, err
	}

	return nil, nil
}

func validateRate(rate string) error {
	r, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return fmt.Errorf("spec.rate must be a valid number, got %q", rate)
	}
	if r <= 0 {
		return fmt.Errorf("spec.rate must be greater than 0, got %q", rate)
	}
	return nil
}

func (r *RateLimitQuota) ValidateUpdate(ctx context.Context, oldObj, newObj *RateLimitQuota) (admission.Warnings, error) {
	ratelimitquotalog.Info("validate update", "name", newObj.Name)

	if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
		return nil, err
	}

	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, fmt.Errorf("spec.name is immutable after creation, was %q, got %q", oldObj.Spec.Name, newObj.Spec.Name)
	}

	if err := validateRate(newObj.Spec.Rate); err != nil {
		return nil, err
	}

	return nil, nil
}

func (r *RateLimitQuota) ValidateDelete(ctx context.Context, obj *RateLimitQuota) (admission.Warnings, error) {
	ratelimitquotalog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
