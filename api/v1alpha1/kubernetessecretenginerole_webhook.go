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
var kubernetessecretenginerolelog = logf.Log.WithName("kubernetessecretenginerole-resource")

func (r *KubernetesSecretEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-kubernetessecretenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kubernetessecretengineroles,verbs=create,versions=v1alpha1,name=mkubernetessecretenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &KubernetesSecretEngineRole{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *KubernetesSecretEngineRole) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*KubernetesSecretEngineRole)
	kubernetessecretenginerolelog.Info("default", "name", cr.Name)
	// TODO(user): fill in your defaulting logic.
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-kubernetessecretenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kubernetessecretengineroles,verbs=update,versions=v1alpha1,name=vkubernetessecretenginerole.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &KubernetesSecretEngineRole{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *KubernetesSecretEngineRole) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*KubernetesSecretEngineRole)
	kubernetessecretenginerolelog.Info("validate create", "name", cr.Name)
	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *KubernetesSecretEngineRole) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	old := oldObj.(*KubernetesSecretEngineRole)
	new := newObj.(*KubernetesSecretEngineRole)

	kubernetessecretenginerolelog.Info("validate update", "name", new.Name)
	// the path cannot be updated
	if new.Spec.Path != old.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *KubernetesSecretEngineRole) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*KubernetesSecretEngineRole)
	kubernetessecretenginerolelog.Info("validate delete", "name", cr.Name)
	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
