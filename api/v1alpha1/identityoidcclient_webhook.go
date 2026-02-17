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
var identityoidcclientlog = logf.Log.WithName("identityoidcclient-resource")

func (r *IdentityOIDCClient) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-identityoidcclient,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=identityoidcclients,verbs=create;update,versions=v1alpha1,name=midentityoidcclient.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &IdentityOIDCClient{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *IdentityOIDCClient) Default() {
	identityoidcclientlog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-identityoidcclient,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=identityoidcclients,verbs=create;update,versions=v1alpha1,name=videntityoidcclient.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &IdentityOIDCClient{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *IdentityOIDCClient) ValidateCreate() (admission.Warnings, error) {
	identityoidcclientlog.Info("validate create", "name", r.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *IdentityOIDCClient) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	identityoidcclientlog.Info("validate update", "name", r.Name)

	// key and client_type cannot be updated after creation
	oldClient := old.(*IdentityOIDCClient)
	if r.Spec.Key != oldClient.Spec.Key {
		return nil, errors.New("spec.key cannot be updated")
	}
	if r.Spec.ClientType != oldClient.Spec.ClientType {
		return nil, errors.New("spec.clientType cannot be updated")
	}
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *IdentityOIDCClient) ValidateDelete() (admission.Warnings, error) {
	identityoidcclientlog.Info("validate delete", "name", r.Name)
	return nil, nil
}
