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
)

// log is for logging in this package.
var databasesecretengineconfiglog = logf.Log.WithName("databasesecretengineconfig-resource")

func (r *DatabaseSecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-databasesecretengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=databasesecretengineconfigs,verbs=create;update,versions=v1alpha1,name=vdatabasesecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &DatabaseSecretEngineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateCreate() error {
	databasesecretengineconfiglog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateUpdate(old runtime.Object) error {
	databasesecretengineconfiglog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*RandomSecret).Spec.Path {
		return errors.New("spec.path cannot be updated")
	}
	return r.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateDelete() error {
	databasesecretengineconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *DatabaseSecretEngineConfig) ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret() error {
	count := 0
	if r.Spec.RootCredentials.RandomSecret != nil {
		count++
	}
	if r.Spec.RootCredentials.Secret != nil {
		count++
	}
	if r.Spec.RootCredentials.VaultSecret != nil {
		count++
	}
	if count > 1 {
		return errors.New("Only one of RootCredentialsFromVaultSecret or RootCredentialsFromSecret or RootCredentialsFromRandomSecret can be specified.")
	}
	return nil
}
