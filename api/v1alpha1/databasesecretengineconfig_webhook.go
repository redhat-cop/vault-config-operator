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

// log is for logging in this package.
var databasesecretengineconfiglog = logf.Log.WithName("databasesecretengineconfig-resource")

func (r *DatabaseSecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-databasesecretengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=databasesecretengineconfigs,verbs=create,versions=v1alpha1,name=mdatabasesecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*DatabaseSecretEngineConfig] = &DatabaseSecretEngineConfig{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) Default(ctx context.Context, obj *DatabaseSecretEngineConfig) error {
	databasesecretengineconfiglog.Info("default", "name", obj.Name)
	return nil
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-databasesecretengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=databasesecretengineconfigs,verbs=create;update,versions=v1alpha1,name=vdatabasesecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Validator[*DatabaseSecretEngineConfig] = &DatabaseSecretEngineConfig{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateCreate(ctx context.Context, obj *DatabaseSecretEngineConfig) (admission.Warnings, error) {
	databasesecretengineconfiglog.Info("validate create", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, obj.isValid()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *DatabaseSecretEngineConfig) (admission.Warnings, error) {
	databasesecretengineconfiglog.Info("validate update", "name", newObj.Name)

	// the path cannot be updated
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	//connection_url, username and verify_connection cannot be changed because they cannot be compare with the actual.
	// if newObj.Spec.ConnectionURL != oldObj.Spec.ConnectionURL {
	// 	return errors.New("spec.connectionURL cannot be updated")
	// }
	// if newObj.Spec.Username != oldObj.Spec.Username {
	// 	return errors.New("spec.username cannot be updated")
	// }
	// if newObj.Spec.VerifyConnection != oldObj.Spec.VerifyConnection {
	// 	return errors.New("spec.verifyConnection cannot be updated")
	// }
	return nil, newObj.isValid()
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateDelete(ctx context.Context, obj *DatabaseSecretEngineConfig) (admission.Warnings, error) {
	databasesecretengineconfiglog.Info("validate delete", "name", obj.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
