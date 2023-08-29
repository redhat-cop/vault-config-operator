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
var githubsecretengineconfiglog = logf.Log.WithName("githubsecretengineconfig-resource")

func (r *GitHubSecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-githubsecretengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=githubsecretengineconfigs,verbs=create,versions=v1alpha1,name=mgithubsecretengineconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &GitHubSecretEngineConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *GitHubSecretEngineConfig) Default() {
	githubsecretengineconfiglog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-githubsecretengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=githubsecretengineconfigs,verbs=create;update,versions=v1alpha1,name=vgithubsecretengineconfig.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &GitHubSecretEngineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GitHubSecretEngineConfig) ValidateCreate() (admission.Warnings, error) {
	githubsecretengineconfiglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, r.isValid()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GitHubSecretEngineConfig) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	githubsecretengineconfiglog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*GitHubSecretEngineConfig).Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, r.isValid()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GitHubSecretEngineConfig) ValidateDelete() (admission.Warnings, error) {
	githubsecretengineconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
