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
	"errors"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type VaultEngineObject interface {
	GetEngineListPah() string
	GetEngineTunePath() string
	GetTunePayload() map[string]interface{}
}

type VaultEngineEndpoint struct {
	*VaultEndpoint
	vaultEngineObject VaultEngineObject
}

func NewVaultEngineEndpoint(obj client.Object) *VaultEngineEndpoint {
	return &VaultEngineEndpoint{
		vaultEngineObject: obj.(VaultEngineObject),
		VaultEndpoint:     NewVaultEndpoint(obj),
	}
}

func (ve *VaultEngineEndpoint) Exists(context context.Context) (bool, error) {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)
	secret, err := vaultClient.Logical().List(ve.vaultEngineObject.GetEngineListPah())
	if err != nil || secret == nil {
		log.Error(err, "unable to list engines at", "path", ve.vaultEngineObject.GetEngineListPah())
		return false, err
	}
	found := false
	for key := range secret.Data {
		if strings.Trim(key, "/") == strings.Split(ve.vaultObject.GetPath(), "/")[len(strings.Split(ve.vaultObject.GetPath(), "/"))] {
			found = true
			break
		}
	}
	return found, nil
}

func (ve *VaultEngineEndpoint) CreateOrUpdateTuneConfig(context context.Context) error {
	log := log.FromContext(context)
	currentTunePayload, err := ve.readTuneConfig(context)
	if err != nil {
		log.Error(err, "unable to read object at", "path", ve.vaultEngineObject.GetEngineTunePath())
		return err
	}

	if !ve.vaultObject.IsEquivalentToDesiredState(currentTunePayload) {
		return write(context, ve.vaultEngineObject.GetEngineTunePath(), ve.vaultEngineObject.GetTunePayload())
	}

	return nil
}

func (ve *VaultEngineEndpoint) readTuneConfig(context context.Context) (map[string]interface{}, error) {
	log := log.FromContext(context)
	secret, _, err := read(context, ve.vaultEngineObject.GetEngineTunePath())
	if err != nil {
		log.Error(err, "unable to read object at", "path", ve.vaultEngineObject.GetEngineTunePath())
		return nil, err
	}
	if secret == nil {
		return nil, errors.New("nul tune config found at path " + ve.vaultEngineObject.GetEngineTunePath())
	}
	return secret, nil
}
