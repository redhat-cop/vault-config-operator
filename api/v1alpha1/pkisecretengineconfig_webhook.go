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
var pkisecretengineconfiglog = logf.Log.WithName("pkisecretengineconfig-resource")

func (r *PKISecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-pkisecretengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=pkisecretengineconfigs,verbs=create,versions=v1alpha1,name=mpkisecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomDefaulter = &PKISecretEngineConfig{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *PKISecretEngineConfig) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*PKISecretEngineConfig)
	pkisecretengineconfiglog.Info("default", "name", cr.Name)
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-pkisecretengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=pkisecretengineconfigs,verbs=update,versions=v1alpha1,name=vpkisecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomValidator = &PKISecretEngineConfig{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *PKISecretEngineConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*PKISecretEngineConfig)
	pkisecretengineconfiglog.Info("validate create", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *PKISecretEngineConfig) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	old := oldObj.(*PKISecretEngineConfig)
	new := newObj.(*PKISecretEngineConfig)

	pkisecretengineconfiglog.Info("validate update", "name", new.Name)

	// the path cannot be updated
	if new.Spec.Path != old.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}

	if new.Spec.Type != old.Spec.Type {
		return nil, errors.New("spec.type cannot be updated")
	}

	if new.Spec.PrivateKeyType != old.Spec.PrivateKeyType {
		return nil, errors.New("spec.privateKeyType cannot be updated")
	}

	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *PKISecretEngineConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*PKISecretEngineConfig)
	pkisecretengineconfiglog.Info("validate delete", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
