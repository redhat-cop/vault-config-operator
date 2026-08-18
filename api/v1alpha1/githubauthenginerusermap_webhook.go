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

var githubauthengineusermaplog = logf.Log.WithName("githubauthenginerusermap-resource")

func (r *GitHubAuthEngineUserMap) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, r).
		WithDefaulter(r).
		WithValidator(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-githubauthenginerusermap,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=githubauthengineusermaps,verbs=create,versions=v1alpha1,name=mgithubauthenginerusermap.kb.io,admissionReviewVersions=v1

var _ admission.Defaulter[*GitHubAuthEngineUserMap] = &GitHubAuthEngineUserMap{}

func (r *GitHubAuthEngineUserMap) Default(ctx context.Context, obj *GitHubAuthEngineUserMap) error {
	githubauthengineusermaplog.Info("default", "name", obj.Name)
	return nil
}

//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-githubauthenginerusermap,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=githubauthengineusermaps,verbs=update,versions=v1alpha1,name=vgithubauthenginerusermap.kb.io,admissionReviewVersions=v1

var _ admission.Validator[*GitHubAuthEngineUserMap] = &GitHubAuthEngineUserMap{}

func (r *GitHubAuthEngineUserMap) ValidateCreate(ctx context.Context, obj *GitHubAuthEngineUserMap) (admission.Warnings, error) {
	githubauthengineusermaplog.Info("validate create", "name", obj.Name)
	return nil, nil
}

func (r *GitHubAuthEngineUserMap) ValidateUpdate(ctx context.Context, oldObj, newObj *GitHubAuthEngineUserMap) (admission.Warnings, error) {
	githubauthengineusermaplog.Info("validate update", "name", newObj.Name)
	if newObj.Spec.Path != oldObj.Spec.Path {
		return nil, errors.New("spec.path cannot be updated")
	}
	if newObj.Spec.Name != oldObj.Spec.Name {
		return nil, errors.New("spec.name cannot be updated")
	}
	return nil, nil
}

func (r *GitHubAuthEngineUserMap) ValidateDelete(ctx context.Context, obj *GitHubAuthEngineUserMap) (admission.Warnings, error) {
	githubauthengineusermaplog.Info("validate delete", "name", obj.Name)
	return nil, nil
}
