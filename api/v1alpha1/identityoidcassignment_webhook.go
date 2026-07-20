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
var identityoidcassignmentlog = logf.Log.WithName("identityoidcassignment-resource")

func (r *IdentityOIDCAssignment) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &IdentityOIDCAssignment{}).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-identityoidcassignment,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=identityoidcassignments,verbs=create;update,versions=v1alpha1,name=midentityoidcassignment.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*IdentityOIDCAssignment] = &IdentityOIDCAssignment{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *IdentityOIDCAssignment) Default(ctx context.Context, obj *IdentityOIDCAssignment) error {
	identityoidcassignmentlog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-identityoidcassignment,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=identityoidcassignments,verbs=create;update,versions=v1alpha1,name=videntityoidcassignment.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*IdentityOIDCAssignment] = &IdentityOIDCAssignment{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityOIDCAssignment) ValidateCreate(ctx context.Context, obj *IdentityOIDCAssignment) (admission.Warnings, error) {
	identityoidcassignmentlog.Info("validate create", "name", obj.Name)
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityOIDCAssignment) ValidateUpdate(ctx context.Context, oldObj, newObj *IdentityOIDCAssignment) (admission.Warnings, error) {
	identityoidcassignmentlog.Info("validate update", "name", newObj.Name)
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *IdentityOIDCAssignment) ValidateDelete(ctx context.Context, obj *IdentityOIDCAssignment) (admission.Warnings, error) {
	identityoidcassignmentlog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
