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

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var rabbitmqsecretenginerolelog = logf.Log.WithName("rabbitmqsecretenginerole-resource")

func (r *RabbitMQSecretEngineRole) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-rabbitmqsecretenginerole,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=rabbitmqsecretengineroles,verbs=create,versions=v1alpha1,name=mrabbitmqsecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &RabbitMQSecretEngineRole{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *RabbitMQSecretEngineRole) Default() {
	rabbitmqsecretenginerolelog.Info("default", "name", r.Name)
	if !controllerutil.ContainsFinalizer(r, vaultutils.GetFinalizer(r)) {
		controllerutil.AddFinalizer(r, vaultutils.GetFinalizer(r))
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-rabbitmqsecretenginerole,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=rabbitmqsecretengineroles,verbs=update,versions=v1alpha1,name=vrabbitmqsecretenginerole.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &RabbitMQSecretEngineRole{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *RabbitMQSecretEngineRole) ValidateCreate() error {
	rabbitmqsecretenginerolelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *RabbitMQSecretEngineRole) ValidateUpdate(old runtime.Object) error {
	rabbitmqsecretenginerolelog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*RabbitMQSecretEngineRole).Spec.Path {
		return errors.New("spec.path cannot be updated")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *RabbitMQSecretEngineRole) ValidateDelete() error {
	rabbitmqsecretenginerolelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
