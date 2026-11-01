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

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var ldapauthenginegrouplog = logf.Log.WithName("ldapauthenginegroup-resource")

func (r *LDAPAuthEngineGroup) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-ldapauthenginegroup,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=ldapauthenginegroups,verbs=create,versions=v1alpha1,name=mldapauthenginegroup.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*LDAPAuthEngineGroup] = &LDAPAuthEngineGroup{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *LDAPAuthEngineGroup) Default(ctx context.Context, obj *LDAPAuthEngineGroup) error {
	ldapauthenginegrouplog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-ldapauthenginegroup,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=ldapauthenginegroups,verbs=update,versions=v1alpha1,name=vldapauthenginegroup.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*LDAPAuthEngineGroup] = &LDAPAuthEngineGroup{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *LDAPAuthEngineGroup) ValidateCreate(ctx context.Context, obj *LDAPAuthEngineGroup) (admission.Warnings, error) {
	ldapauthenginegrouplog.Info("validate create", "name", obj.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *LDAPAuthEngineGroup) ValidateUpdate(ctx context.Context, oldObj, newObj *LDAPAuthEngineGroup) (admission.Warnings, error) {
	ldapauthenginegrouplog.Info("validate update", "name", newObj.Name)

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *LDAPAuthEngineGroup) ValidateDelete(ctx context.Context, obj *LDAPAuthEngineGroup) (admission.Warnings, error) {
	ldapauthenginegrouplog.Info("validate delete", "name", obj.Name)

	return nil, nil
}
