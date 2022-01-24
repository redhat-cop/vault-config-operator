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
	
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:object:generate:=false
type RabbitMQSecretEngineConfigValidation struct {
	Client  client.Client
}

func (r *RabbitMQSecretEngineConfigValidation) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Operation {
	case "CREATE":
		rabbitMQSecretEngineConfig := &RabbitMQSecretEngineConfig{}
		// Using json Unmarshal as Decoder has issues to decode specific type 
		if err := json.Unmarshal(req.Object.Raw, rabbitMQSecretEngineConfig); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		vaultNamespace := rabbitMQSecretEngineConfig.Spec.Authentication.Namespace
		rabbitMQSecretEngineConfigList := &RabbitMQSecretEngineConfigList{}
		if err := r.Client.List(ctx, rabbitMQSecretEngineConfigList); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		for _, config := range rabbitMQSecretEngineConfigList.Items {
			if vaultNamespace != "" {
				// Check Vault namespace with the path
				if vaultNamespace == config.Spec.Authentication.Namespace && config.Spec.Path == rabbitMQSecretEngineConfig.Spec.Path {
					return admission.Errored(http.StatusBadRequest, errors.New("rabbitMQ engine already configured at spec.path in Vault Namespace " + vaultNamespace))
				}
			} else {
				if config.Spec.Path == rabbitMQSecretEngineConfig.Spec.Path {
					return admission.Errored(http.StatusBadRequest, errors.New("rabbitMQ engine already configured at spec.path"))
				}
			}
		}
		return admission.Allowed("")
	case "UPDATE":
		rabbitMQSecretEngineConfig := &RabbitMQSecretEngineConfig{}
		// Using json Unmarshal as Decoder has issues to decode specific type 
		if err := json.Unmarshal(req.Object.Raw, rabbitMQSecretEngineConfig); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		oldRabbitMQSecretEngineConfig := &RabbitMQSecretEngineConfig{}
		if err := json.Unmarshal(req.OldObject.Raw, oldRabbitMQSecretEngineConfig); err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		if rabbitMQSecretEngineConfig.Spec.Path != oldRabbitMQSecretEngineConfig.Spec.Path {
			return admission.Errored(http.StatusBadRequest, errors.New("spec.path cannot be updated"))
		}
		return admission.Allowed("")
	default:
		return admission.Allowed("")
	}
}
