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

var kerberosauthengineldapconflog = logf.Log.WithName("kerberosauthengineldapconfig-resource")

func (r *KerberosAuthEngineLDAPConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-kerberosauthengineldapconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kerberosauthengineldapconfigs,verbs=create,versions=v1alpha1,name=mkerberosauthengineldapconfig.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*KerberosAuthEngineLDAPConfig] = &KerberosAuthEngineLDAPConfig{}

func (r *KerberosAuthEngineLDAPConfig) Default(ctx context.Context, obj *KerberosAuthEngineLDAPConfig) error {
	kerberosauthengineldapconflog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-kerberosauthengineldapconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kerberosauthengineldapconfigs,verbs=create;update,versions=v1alpha1,name=vkerberosauthengineldapconfig.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*KerberosAuthEngineLDAPConfig] = &KerberosAuthEngineLDAPConfig{}

func (r *KerberosAuthEngineLDAPConfig) ValidateCreate(ctx context.Context, obj *KerberosAuthEngineLDAPConfig) (admission.Warnings, error) {
	kerberosauthengineldapconflog.Info("validate create", "name", obj.Name)
	return nil, obj.isValid()
}

func (r *KerberosAuthEngineLDAPConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *KerberosAuthEngineLDAPConfig) (admission.Warnings, error) {
	kerberosauthengineldapconflog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, newObj.isValid()
}

func (r *KerberosAuthEngineLDAPConfig) ValidateDelete(ctx context.Context, obj *KerberosAuthEngineLDAPConfig) (admission.Warnings, error) {
	kerberosauthengineldapconflog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
