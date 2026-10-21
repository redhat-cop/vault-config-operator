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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var leasecountquotalog = logf.Log.WithName("leasecountquota-resource")

func (r *LeaseCountQuota) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-leasecountquota,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=leasecountquotas,verbs=create,versions=v1alpha1,name=mleasecountquota.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*LeaseCountQuota] = &LeaseCountQuota{}

func (r *LeaseCountQuota) Default(ctx context.Context, obj *LeaseCountQuota) error {
	leasecountquotalog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-leasecountquota,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=leasecountquotas,verbs=create;update,versions=v1alpha1,name=vleasecountquota.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*LeaseCountQuota] = &LeaseCountQuota{}

func (r *LeaseCountQuota) ValidateCreate(ctx context.Context, obj *LeaseCountQuota) (admission.Warnings, error) {
	leasecountquotalog.Info("validate create", "name", obj.Name)

	if obj.Spec.MaxLeases < 1 {
		return nil, fmt.Errorf("spec.maxLeases must be at least 1")
	}

	return nil, nil
}

func (r *LeaseCountQuota) ValidateUpdate(ctx context.Context, oldObj, newObj *LeaseCountQuota) (admission.Warnings, error) {
	leasecountquotalog.Info("validate update", "name", newObj.Name)

	if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
		return nil, err
	}

	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, fmt.Errorf("spec.name is immutable after creation, was %q, got %q", oldObj.Spec.Name, newObj.Spec.Name)
	}

	if newObj.Spec.MaxLeases < 1 {
		return nil, fmt.Errorf("spec.maxLeases must be at least 1")
	}

	return nil, nil
}

func (r *LeaseCountQuota) ValidateDelete(ctx context.Context, obj *LeaseCountQuota) (admission.Warnings, error) {
	leasecountquotalog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
