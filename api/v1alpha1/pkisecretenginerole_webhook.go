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
var pkisecretenginerolelog = logf.Log.WithName("pkisecretenginerole-resource")

func (r *PKISecretEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-pkisecretenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=pkisecretengineroles,verbs=create,versions=v1alpha1,name=mpkisecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomDefaulter = &PKISecretEngineRole{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *PKISecretEngineRole) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*PKISecretEngineRole)
	pkisecretenginerolelog.Info("default", "name", cr.Name)
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-pkisecretenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=pkisecretengineroles,verbs=update,versions=v1alpha1,name=vpkisecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomValidator = &PKISecretEngineRole{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *PKISecretEngineRole) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*PKISecretEngineRole)
	pkisecretenginerolelog.Info("validate create", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *PKISecretEngineRole) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	old := oldObj.(*PKISecretEngineRole)
	new := newObj.(*PKISecretEngineRole)

	pkisecretenginerolelog.Info("validate update", "name", new.Name)
	pkisecretenginerolelog.Info("validate update", "name", new.Name)

	// the path cannot be updated
	if new.Spec.Path != old.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *PKISecretEngineRole) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*PKISecretEngineRole)
	pkisecretenginerolelog.Info("validate delete", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
