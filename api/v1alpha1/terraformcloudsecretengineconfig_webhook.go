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

var terraformcloudsecretengineconfiglog = logf.Log.WithName("terraformcloudsecretengineconfig-resource")

func (r *TerraformCloudSecretEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-terraformcloudsecretengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=terraformcloudsecretengineconfigs,verbs=create,versions=v1alpha1,name=mterraformcloudsecretengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.Defaulter[*TerraformCloudSecretEngineConfig] = &TerraformCloudSecretEngineConfig{}

func (r *TerraformCloudSecretEngineConfig) Default(ctx context.Context, obj *TerraformCloudSecretEngineConfig) error {
	terraformcloudsecretengineconfiglog.Info("default", "name", obj.Name)
	if obj.Spec.TFCCredentials.PasswordKey == "" || obj.Spec.TFCCredentials.PasswordKey == "password" {
		obj.Spec.TFCCredentials.PasswordKey = "token"
	}
	return nil
}
