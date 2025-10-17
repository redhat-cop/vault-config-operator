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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var certauthenginerolelog = logf.Log.WithName("certauthenginerole-resource")

func (r *CertAuthEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-certauthenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=certauthengineroles,verbs=create;update,versions=v1alpha1,name=mcertauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &CertAuthEngineRole{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *CertAuthEngineRole) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*CertAuthEngineRole)
	certauthenginerolelog.Info("default", "name", cr.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-certauthenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=certauthengineroles,verbs=create;update,versions=v1alpha1,name=vcertauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &CertAuthEngineRole{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *CertAuthEngineRole) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*CertAuthEngineRole)
	certauthenginerolelog.Info("validate create", "name", cr.Name)
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *CertAuthEngineRole) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	old := oldObj.(*CertAuthEngineRole)
	new := newObj.(*CertAuthEngineRole)

	certauthenginerolelog.Info("validate update", "name", new.Name)

	if new.Spec.Path != old.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}

	if new.Spec.Name != old.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}

	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *CertAuthEngineRole) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*CertAuthEngineRole)
	certauthenginerolelog.Info("validate delete", "name", cr.Name)

	return nil, nil
}
