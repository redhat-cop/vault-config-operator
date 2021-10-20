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

package utils

import (
	"context"

	vault "github.com/hashicorp/vault/api"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func write(context context.Context, path string, payload map[string]interface{}) error {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)
	_, err := vaultClient.Logical().Write(path, payload)
	if err != nil {
		log.Error(err, "unable to write object at", "path", path)
		return err
	}
	return nil
}

func read(context context.Context, path string) (map[string]interface{}, bool, error) {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)
	secret, err := vaultClient.Logical().Read(path)
	if err != nil {
		if respErr, ok := err.(*vault.ResponseError); ok {
			if respErr.StatusCode == 404 {
				return nil, false, nil
			}
		}
		log.Error(err, "unable to read object at", "path", path)
		return nil, false, err
	}
	if secret == nil {
		return nil, false, nil
	}
	return secret.Data, true, nil
}
