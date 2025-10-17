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
var databasesecretengineconfiglog = logf.Log.WithName("databasesecretengineconfig-resource")

func (r *DatabaseSecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-databasesecretengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=databasesecretengineconfigs,verbs=create,versions=v1alpha1,name=mdatabasesecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomDefaulter = &DatabaseSecretEngineConfig{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) Default(_ context.Context, obj runtime.Object) error {
	o := obj.(*DatabaseSecretEngineConfig)
	authenginemountlog.Info("default", "name", o.Name)
	return nil
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-databasesecretengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=databasesecretengineconfigs,verbs=create;update,versions=v1alpha1,name=vdatabasesecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomValidator = &DatabaseSecretEngineConfig{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	o := obj.(*DatabaseSecretEngineConfig)
	databasesecretengineconfiglog.Info("validate create", "name", o.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, o.isValid()
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	old := oldObj.(*DatabaseSecretEngineConfig)
	new := newObj.(*DatabaseSecretEngineConfig)

	databasesecretengineconfiglog.Info("validate update", "name", new.Name)

	// the path cannot be updated
	if new.Spec.Path != old.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	//connection_url, username and verify_connection cannot be changed because they cannot be compare with the actual.
	// if new.Spec.ConnectionURL != old.Spec.ConnectionURL {
	// 	return nil, errors.New("spec.connectionURL cannot be updated")
	// }
	// if new.Spec.Username != old.Spec.Username {
	// 	return nil, errors.New("spec.username cannot be updated")
	// }
	// if new.Spec.VerifyConnection != old.Spec.VerifyConnection {
	// 	return nil, errors.New("spec.verifyConnection cannot be updated")
	// }
	return nil, new.isValid()
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *DatabaseSecretEngineConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	o := obj.(*DatabaseSecretEngineConfig)
	databasesecretengineconfiglog.Info("validate delete", "name", o.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
