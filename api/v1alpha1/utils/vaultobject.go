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
	"strings"

	vault "github.com/hashicorp/vault/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type VaultObject interface {
	GetPath() string
	GetPayload() map[string]interface{}
	// IsEquivalentToDesiredState returns wether the passed payload is equivalent to the payload that the current object would generate. When this is a engine object the tune payload will be compared
	IsEquivalentToDesiredState(payload map[string]interface{}) bool
	IsInitialized() bool
	IsValid() (bool, error)
	PrepareInternalValues(context context.Context, object client.Object) error
	GetKubeAuthConfiguration() *KubeAuthConfiguration
	GetVaultConnection() *VaultConnection
}

type VaultEndpoint struct {
	vaultObject VaultObject
}

func NewVaultEndpoint(obj client.Object) *VaultEndpoint {
	return &VaultEndpoint{
		vaultObject: obj.(VaultObject),
	}
}

// Deletes all versions and metadata of the KVv2 secret
// This is similar to vaultClient.KVv2(mountPath string).DeleteMetadata(ctx context.Context, secretPath string) but works better with existing interface
func (ve *VaultEndpoint) DeleteKVv2IfExists(context context.Context) error {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)

	// should match pathToDelete := fmt.Sprintf("%s/metadata/%s", kv.mountPath, secretPath)
	pathToDelete := strings.Replace(ve.vaultObject.GetPath(), "/data/", "/metadata/", 1)

	_, err := vaultClient.Logical().Delete(pathToDelete)
	if err != nil {
		if respErr, ok := err.(*vault.ResponseError); ok {
			if respErr.StatusCode == 404 {
				return nil
			}
		}
		log.Error(err, "unable to delete object at", "path", pathToDelete)
		return err
	}
	return nil
}

func (ve *VaultEndpoint) DeleteIfExists(context context.Context) error {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)
	_, err := vaultClient.Logical().Delete(ve.vaultObject.GetPath())
	if err != nil {
		if respErr, ok := err.(*vault.ResponseError); ok {
			if respErr.StatusCode == 404 {
				return nil
			}
		}
		log.Error(err, "unable to delete object at", "path", ve.vaultObject.GetPath())
		return err
	}
	return nil
}

func (ve *VaultEndpoint) Create(context context.Context) error {
	return write(context, ve.vaultObject.GetPath(), ve.vaultObject.GetPayload())
}

func (ve *VaultEndpoint) CreateOrUpdate(context context.Context) error {
	log := log.FromContext(context)
	currentPayload, found, err := read(context, ve.vaultObject.GetPath())
	if err != nil {
		log.Error(err, "unable to read object at", "path", ve.vaultObject.GetPath())
		return err
	}
	if !found {
		return write(context, ve.vaultObject.GetPath(), ve.vaultObject.GetPayload())
	} else {
		if !ve.vaultObject.IsEquivalentToDesiredState(currentPayload) {
			return write(context, ve.vaultObject.GetPath(), ve.vaultObject.GetPayload())
		}
	}
	return nil
}

type RabbitMQEngineConfigVaultObject interface {
	VaultObject
	GetLeasePath() string
	GetLeasePayload() map[string]interface{}
	CheckTTLValuesProvided() bool
}

type RabbitMQEngineConfigVaultEndpoint struct {
	rabbitMQEngineConfigVaultEndpoint RabbitMQEngineConfigVaultObject
}

func (ve *RabbitMQEngineConfigVaultEndpoint) CreateOrUpdateLease(context context.Context) error {
	log := log.FromContext(context)
	// Skip lease configuration if no values provided
	if ve.rabbitMQEngineConfigVaultEndpoint.CheckTTLValuesProvided() {
		return nil
	}
	currentPayload, found, err := read(context, ve.rabbitMQEngineConfigVaultEndpoint.GetLeasePath())
	if err != nil {
		log.Error(err, "unable to read object at", "path", ve.rabbitMQEngineConfigVaultEndpoint.GetLeasePath())
		return err
	}
	if !found {
		return write(context, ve.rabbitMQEngineConfigVaultEndpoint.GetLeasePath(), ve.rabbitMQEngineConfigVaultEndpoint.GetLeasePayload())
	} else {
		if !ve.rabbitMQEngineConfigVaultEndpoint.IsEquivalentToDesiredState(currentPayload) {
			return write(context, ve.rabbitMQEngineConfigVaultEndpoint.GetLeasePath(), ve.rabbitMQEngineConfigVaultEndpoint.GetLeasePayload())
		}
	}
	return nil
}

func (ve *RabbitMQEngineConfigVaultEndpoint) Create(context context.Context) error {
	return write(context, ve.rabbitMQEngineConfigVaultEndpoint.GetPath(), ve.rabbitMQEngineConfigVaultEndpoint.GetPayload())
}

func NewRabbitMQEngineConfigVaultEndpoint(obj client.Object) *RabbitMQEngineConfigVaultEndpoint {
	return &RabbitMQEngineConfigVaultEndpoint{
		rabbitMQEngineConfigVaultEndpoint: obj.(RabbitMQEngineConfigVaultObject),
	}
}
