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
	"github.com/go-logr/logr"
	vault "github.com/hashicorp/vault/api"
)

type VaultConnection interface {
	GetNamespace() string
	GetRole() string
	GetKubeAuthPath() string
}

type VaultObject interface {
	GetPath() string
	GetPayload() map[string]interface{}
	IsEquivalentToDesiredState(payload map[string]interface{}) bool
}

type VaultEndpoint struct {
	VaultConnection
	VaultObject
	log    logr.Logger
	client *vault.Client
}

func NewVaultEndpoint(vaultConnection VaultConnection, vaultObject VaultObject, jwt string, log logr.Logger) (*VaultEndpoint, error) {
	vaultEndpoint := &VaultEndpoint{
		log:             log,
		VaultConnection: vaultConnection,
		VaultObject:     vaultObject,
	}
	vaultClient, err := vaultEndpoint.createVaultClient(jwt)
	if err != nil {
		vaultEndpoint.log.Error(err, "unable to create vault client")
		return nil, err
	}
	vaultEndpoint.client = vaultClient
	return vaultEndpoint, nil
}

func (ve *VaultEndpoint) GetVaultClient() *vault.Client {
	return ve.client
}

func (ve *VaultEndpoint) createVaultClient(jwt string) (*vault.Client, error) {
	config := vault.DefaultConfig()
	client, err := vault.NewClient(config)
	if err != nil {
		ve.log.Error(err, "unable initialize vault client")
		return nil, err
	}
	if ve.GetNamespace() != "" {
		client.SetNamespace(ve.GetNamespace())
	}
	secret, err := client.Logical().Write(ve.GetKubeAuthPath()+"/login", map[string]interface{}{
		"jwt":  jwt,
		"role": ve.GetRole(),
	})

	if err != nil {
		ve.log.Error(err, "unable to login to vault")
		return nil, err
	}

	client.SetToken(secret.Auth.ClientToken)

	return client, nil
}

func (ve *VaultEndpoint) Read() (map[string]interface{}, bool, error) {
	secret, err := ve.GetVaultClient().Logical().Read(ve.GetPath())
	if err != nil {
		if respErr, ok := err.(*vault.ResponseError); ok {
			if respErr.StatusCode == 404 {
				return nil, false, nil
			}
		}
		ve.log.Error(err, "unable to read object at", "path", ve.GetPath())
		return nil, false, err
	}
	return secret.Data, true, nil
}

func (ve *VaultEndpoint) Write(payload map[string]interface{}) error {
	_, err := ve.GetVaultClient().Logical().Write(ve.GetPath(), payload)
	if err != nil {
		ve.log.Error(err, "unable to write object at", "path", ve.GetPath())
		return err
	}
	return nil
}
func (ve *VaultEndpoint) Delete() error {
	_, err := ve.GetVaultClient().Logical().Delete(ve.GetPath())
	if err != nil {
		ve.log.Error(err, "unable to delete object at", "path", ve.GetPath())
		return err
	}
	return nil
}
func (ve *VaultEndpoint) CreateOrUpdate(payload map[string]interface{}) error {
	currentPayload, found, err := ve.Read()
	if err != nil {
		ve.log.Error(err, "unable to read object at", "path", ve.GetPath())
		return err
	}
	if !found {
		return ve.Write(payload)
	} else {
		if !ve.IsEquivalentToDesiredState(currentPayload) {
			return ve.Write(payload)
		}
	}
	return nil
}
