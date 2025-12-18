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
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var kubernetesauthengineconfiglog = logf.Log.WithName("kubernetesauthengineconfig-resource")

func (r *KubernetesAuthEngineConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-kubernetesauthengineconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kubernetesauthengineconfigs,verbs=create,versions=v1alpha1,name=mkubernetesauthengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomDefaulter = &KubernetesAuthEngineConfig{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (r *KubernetesAuthEngineConfig) Default(_ context.Context, obj runtime.Object) error {
	cr := obj.(*KubernetesAuthEngineConfig)
	kubernetesauthengineconfiglog.Info("default", "name", cr.Name)
	if cr.Spec.UseOperatorPodCA && cr.Spec.KubernetesCACert == "" {
		b, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
		if err != nil {
			kubernetesauthengineconfiglog.Error(err, "unable to read file /var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
			return nil
		}
		cr.Spec.KubernetesCACert = string(b)
	}
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-kubernetesauthengineconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=kubernetesauthengineconfigs,verbs=update,versions=v1alpha1,name=vkubernetesauthengineconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ admission.CustomValidator = &KubernetesAuthEngineConfig{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (r *KubernetesAuthEngineConfig) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*KubernetesAuthEngineConfig)
	kubernetesauthengineconfiglog.Info("validate create", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (r *KubernetesAuthEngineConfig) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	old := oldObj.(*KubernetesAuthEngineConfig)
	new := newObj.(*KubernetesAuthEngineConfig)

	kubernetesauthengineconfiglog.Info("validate update", "name", new.Name)

	// the path cannot be updated
	if new.Spec.Path != old.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (r *KubernetesAuthEngineConfig) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	cr := obj.(*KubernetesAuthEngineConfig)
	kubernetesauthengineconfiglog.Info("validate delete", "name", cr.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
