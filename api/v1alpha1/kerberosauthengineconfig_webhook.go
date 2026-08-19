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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var kerberosauthengineconflog = logf.Log.WithName("kerberosauthengineconfig-resource")

func (r *KerberosAuthEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-kerberosauthengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kerberosauthengineconfigs,verbs=create,versions=v1alpha1,name=mkerberosauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*KerberosAuthEngineConfig] = &KerberosAuthEngineConfig{}

func (r *KerberosAuthEngineConfig) Default(ctx context.Context, obj *KerberosAuthEngineConfig) error {
	kerberosauthengineconflog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-kerberosauthengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kerberosauthengineconfigs,verbs=create;update,versions=v1alpha1,name=vkerberosauthengineconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*KerberosAuthEngineConfig] = &KerberosAuthEngineConfig{}

func (r *KerberosAuthEngineConfig) ValidateCreate(ctx context.Context, obj *KerberosAuthEngineConfig) (admission.Warnings, error) {
	kerberosauthengineconflog.Info("validate create", "name", obj.Name)
	return nil, nil
}

func (r *KerberosAuthEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *KerberosAuthEngineConfig) (admission.Warnings, error) {
	kerberosauthengineconflog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, nil
}

func (r *KerberosAuthEngineConfig) ValidateDelete(ctx context.Context, obj *KerberosAuthEngineConfig) (admission.Warnings, error) {
	kerberosauthengineconflog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
