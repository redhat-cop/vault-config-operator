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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type VaultPKIEngineObject interface {
	GetGeneratePath() string
	GetDeletePath() string
	GetGeneratedStatus() bool
	SetGeneratedStatus(status bool)
	GetConfigUrlsPath() string
	GetConfigCrlPath() string
	GetConfigUrlsPayload() map[string]interface{}
	GetConfigCrlPayload() map[string]interface{}
	CreateExported(context context.Context, secret *vault.Secret) (bool, error)
	SetExportedStatus(status bool)
	SetIntermediate(context context.Context) error
	GetSignedStatus() bool
	SetSignedStatus(status bool)
}

type VaultPKIEngineEndpoint struct {
	*VaultEndpoint
	vaultPKIEngineObject VaultPKIEngineObject
}

func NewVaultPKIEngineEndpoint(obj client.Object) *VaultPKIEngineEndpoint {
	return &VaultPKIEngineEndpoint{
		vaultPKIEngineObject: obj.(VaultPKIEngineObject),
		VaultEndpoint:        NewVaultEndpoint(obj),
	}
}

func (ve *VaultPKIEngineEndpoint) Exists(context context.Context) (bool, error) {
	return ve.vaultPKIEngineObject.GetGeneratedStatus(), nil
}

func (ve *VaultPKIEngineEndpoint) Generate(context context.Context) (*vault.Secret, error) {
	return writeWithResponse(context, ve.vaultPKIEngineObject.GetGeneratePath(), ve.vaultObject.GetPayload())
}

func (ve *VaultPKIEngineEndpoint) DeleteIfExists(context context.Context) error {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)
	_, err := vaultClient.Logical().Delete(ve.vaultPKIEngineObject.GetDeletePath())
	if err != nil {
		if respErr, ok := err.(*vault.ResponseError); ok {
			if respErr.StatusCode == 404 {
				return nil
			}
		}
		log.Error(err, "unable to delete object at", "path", ve.vaultPKIEngineObject.GetDeletePath())
		return err
	}
	return nil
}

func (ve *VaultPKIEngineEndpoint) CreateOrUpdateConfigUrls(context context.Context) error {
	return ve.CreateOrUpdateConfig(context, ve.vaultPKIEngineObject.GetConfigUrlsPath(), ve.vaultPKIEngineObject.GetConfigUrlsPayload())
}

// func (ve *VaultPKIEngineEndpoint) readConfigUrls(context context.Context) (map[string]interface{}, error) {
// 	return ve.readConfig(context, ve.vaultPKIEngineObject.GetConfigUrlsPath())
// }

func (ve *VaultPKIEngineEndpoint) CreateOrUpdateConfigCrl(context context.Context) error {
	return ve.CreateOrUpdateConfig(context, ve.vaultPKIEngineObject.GetConfigCrlPath(), ve.vaultPKIEngineObject.GetConfigCrlPayload())
}

// func (ve *VaultPKIEngineEndpoint) readConfigCrl(context context.Context) (map[string]interface{}, error) {
// 	return ve.readConfig(context, ve.vaultPKIEngineObject.GetConfigCrlPath())
// }

func (ve *VaultPKIEngineEndpoint) readConfig(context context.Context, configPath string) (map[string]interface{}, error) {
	log := log.FromContext(context)
	config, _, err := read(context, configPath)
	if err != nil {
		log.Error(err, "unable to read object at", "path", configPath)
		return nil, err
	}
	return config, nil
}

func (ve *VaultPKIEngineEndpoint) CreateOrUpdateConfig(context context.Context, configPath string, payload map[string]interface{}) error {
	log := log.FromContext(context)
	currentConfigPayload, err := ve.readConfig(context, configPath)
	if err != nil {
		log.Error(err, "unable to read object at", "path", configPath)
		return err
	}

	if !ve.vaultObject.IsEquivalentToDesiredState(currentConfigPayload) {
		return write(context, ve.vaultPKIEngineObject.GetConfigCrlPath(), payload)
	}

	return nil
}

func (ve *VaultPKIEngineEndpoint) CreateExported(context context.Context, secret *vault.Secret) (bool, error) {
	return ve.vaultPKIEngineObject.CreateExported(context, secret)
}

func (ve *VaultPKIEngineEndpoint) CreateIntermediate(context context.Context) error {

	return ve.vaultPKIEngineObject.SetIntermediate(context)
}
