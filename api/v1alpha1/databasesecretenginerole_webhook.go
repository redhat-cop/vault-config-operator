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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var databasesecretenginerolelog = logf.Log.WithName("databasesecretenginerole-resource")

func (r *DatabaseSecretEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-databasesecretenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=databasesecretengineroles,verbs=create,versions=v1alpha1,name=mdatabasesecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &DatabaseSecretEngineRole{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *DatabaseSecretEngineRole) Default() {
	authenginemountlog.Info("default", "name", r.Name)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-databasesecretenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=databasesecretengineroles,verbs=update,versions=v1alpha1,name=vdatabasesecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &DatabaseSecretEngineRole{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DatabaseSecretEngineRole) ValidateCreate() (admission.Warnings, error) {
	databasesecretenginerolelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DatabaseSecretEngineRole) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	databasesecretenginerolelog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*DatabaseSecretEngineRole).Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DatabaseSecretEngineRole) ValidateDelete() (admission.Warnings, error) {
	databasesecretenginerolelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
