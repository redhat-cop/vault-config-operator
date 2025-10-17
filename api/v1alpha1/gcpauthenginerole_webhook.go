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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var gcpauthenginerolelog = logf.Log.WithName("gcpauthenginerole-resource")

func (r *GCPAuthEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-gcpauthenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=gcpauthengineroles,verbs=create,versions=v1alpha1,name=mgcpauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &GCPAuthEngineRole{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *GCPAuthEngineRole) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*GCPAuthEngineRole)
	gcpauthenginerolelog.Info("default", "name", cr.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-gcpauthenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=gcpauthengineroles,verbs=update,versions=v1alpha1,name=vgcpauthenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &GCPAuthEngineRole{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *GCPAuthEngineRole) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*GCPAuthEngineRole)
	gcpauthenginerolelog.Info("validate create", "name", cr.Name)

	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *GCPAuthEngineRole) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	_ = oldObj.(*GCPAuthEngineRole)
	new := newObj.(*GCPAuthEngineRole)

	gcpauthenginerolelog.Info("validate update", "name", new.Name)

	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *GCPAuthEngineRole) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*GCPAuthEngineRole)
	gcpauthenginerolelog.Info("validate delete", "name", cr.Name)

	return nil, nil
}
