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

var kerberosauthenginegroupplog = logf.Log.WithName("kerberosauthenginegroup-resource")

func (r *KerberosAuthEngineGroup) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-kerberosauthenginegroup,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kerberosauthenginegroups,verbs=create,versions=v1alpha1,name=mkerberosauthenginegroup.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*KerberosAuthEngineGroup] = &KerberosAuthEngineGroup{}

func (r *KerberosAuthEngineGroup) Default(ctx context.Context, obj *KerberosAuthEngineGroup) error {
	kerberosauthenginegroupplog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-kerberosauthenginegroup,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kerberosauthenginegroups,verbs=create;update,versions=v1alpha1,name=vkerberosauthenginegroup.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*KerberosAuthEngineGroup] = &KerberosAuthEngineGroup{}

func (r *KerberosAuthEngineGroup) ValidateCreate(ctx context.Context, obj *KerberosAuthEngineGroup) (admission.Warnings, error) {
	kerberosauthenginegroupplog.Info("validate create", "name", obj.Name)
	return nil, nil
}

func (r *KerberosAuthEngineGroup) ValidateUpdate(ctx context.Context, oldObj, newObj *KerberosAuthEngineGroup) (admission.Warnings, error) {
	kerberosauthenginegroupplog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	return nil, nil
}

func (r *KerberosAuthEngineGroup) ValidateDelete(ctx context.Context, obj *KerberosAuthEngineGroup) (admission.Warnings, error) {
	kerberosauthenginegroupplog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
