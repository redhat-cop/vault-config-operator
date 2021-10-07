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
var vaultrolelog = logf.Log.WithName("vaultrole-resource")

func (r *VaultRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-vaultrole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=vaultroles,verbs=create;update,versions=v1alpha1,name=vvaultrole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &VaultRole{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *VaultRole) ValidateCreate() error {
	vaultrolelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.ValidateEitherTargetNamespaceSelectorOrTargetNamespace()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *VaultRole) ValidateUpdate(old runtime.Object) error {
	vaultrolelog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*SecretEngineMount).Spec.Path {
		return errors.New("spec.path cannot be updated")
	}
	return r.ValidateEitherTargetNamespaceSelectorOrTargetNamespace()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *VaultRole) ValidateDelete() error {
	vaultrolelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *VaultRole) ValidateEitherTargetNamespaceSelectorOrTargetNamespace() error {
	count := 0
	if r.Spec.TargetNamespaceSelector != nil {
		count++
	}
	if r.Spec.TargetNamespaces != nil {
		count++
	}
	if count > 1 {
		return errors.New("Only one of TargetNamespaceSelector or TargetNamespaces can be specified.")
	}
	return nil
}
