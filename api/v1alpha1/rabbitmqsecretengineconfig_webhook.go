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
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var rabbitmqsecretengineconfiglog = logf.Log.WithName("rabbitmqsecretengineconfig-resource")

func (r *RabbitMQSecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-rabbitmqsecretengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=rabbitmqsecretengineconfigs,verbs=update,versions=v1alpha1,name=vrabbitmqsecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &RabbitMQSecretEngineConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *RabbitMQSecretEngineConfig) ValidateCreate() error {
	rabbitmqsecretengineconfiglog.Info("validate create", "name", r.Name)

	var ctx context.Context

	vaultClient, err := r.Spec.Authentication.GetVaultClient(ctx, r.Namespace)
	if err != nil {
		rabbitmqsecretengineconfiglog.Error(err, "unable to create vault client required for instance validation", "instance", r)
		return err
	}
	mounts, err := vaultClient.Sys().ListMounts()
	if err != nil {
		// Operator requires permission to list mounts for validation
		rabbitmqsecretengineconfiglog.Error(err, "failed to list mounts", "instance", r)
		return err
	}
	if mountExists := checkIfMountExists(mounts, r.Spec.Path); mountExists {
		err := fmt.Errorf("RabbitMQ Engine with path '%s' already exists", r.Spec.Path)
		rabbitmqsecretengineconfiglog.Error(err, "Error! Engine already mounted")
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *RabbitMQSecretEngineConfig) ValidateUpdate(old runtime.Object) error {
	rabbitmqsecretengineconfiglog.Info("validate update", "name", r.Name)

	// the path cannot be updated
	if r.Spec.Path != old.(*RabbitMQSecretEngineConfig).Spec.Path {
		return errors.New("spec.path cannot be updated")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *RabbitMQSecretEngineConfig) ValidateDelete() error {
	rabbitmqsecretengineconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// Iterate over mounts list and return true if provided path is found 
func checkIfMountExists(mounts map[string]*vault.MountOutput, path Path) (bool) {
	// Make sure path has / at the end as mounts always have it

	for mount, _ := range mounts {
		if mount == string(path) { 
			return true
		}
	}
	return false
}
