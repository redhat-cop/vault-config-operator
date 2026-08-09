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

// VaultTransitKeyObject extends VaultObject with the dual-path methods
// required by Transit keys: config updates go to a separate path.
type VaultTransitKeyObject interface {
	VaultObject
	GetConfigPath() string
	GetConfigPayload() map[string]any
}

// VaultTransitKeyEndpoint handles Transit key lifecycle with dual-path
// create/update semantics: creates at GetPath(), updates config at GetConfigPath().
type VaultTransitKeyEndpoint struct {
	transitKeyObject VaultTransitKeyObject
}

func NewVaultTransitKeyEndpoint(obj client.Object) *VaultTransitKeyEndpoint {
	return &VaultTransitKeyEndpoint{
		transitKeyObject: obj.(VaultTransitKeyObject),
	}
}

// CreateOrUpdate reads the key from Vault. If not found, creates it at GetPath().
// If found and config differs, writes only the config-time fields to GetConfigPath().
func (ve *VaultTransitKeyEndpoint) CreateOrUpdate(ctx context.Context) error {
	log := log.FromContext(ctx)
	currentPayload, found, err := read(ctx, ve.transitKeyObject.GetPath())
	if err != nil {
		log.Error(err, "unable to read object at", "path", ve.transitKeyObject.GetPath())
		return err
	}
	if !found {
		err = write(ctx, ve.transitKeyObject.GetPath(), ve.transitKeyObject.GetPayload())
		if err != nil {
			return err
		}
		err = write(ctx, ve.transitKeyObject.GetConfigPath(), ve.transitKeyObject.GetConfigPayload())
		if err != nil {
			// Best-effort rollback: Vault creates keys with deletion_allowed=false,
			// so this DELETE may fail if the config write (which would set
			// deletion_allowed=true) is exactly what failed. In that case the key
			// remains in Vault, the error is returned, and the reconciler will
			// retry the config write on the next reconcile loop.
			log.Error(err, "config write failed after key creation, attempting best-effort rollback", "path", ve.transitKeyObject.GetPath())
			if delErr := deleteIfExists(ctx, ve.transitKeyObject.GetPath()); delErr != nil {
				log.Error(delErr, "best-effort rollback failed (key may be orphaned until next successful reconcile)", "path", ve.transitKeyObject.GetPath())
			}
			return err
		}
		return nil
	}
	if !ve.transitKeyObject.IsEquivalentToDesiredState(currentPayload) {
		return write(ctx, ve.transitKeyObject.GetConfigPath(), ve.transitKeyObject.GetConfigPayload())
	}
	return nil
}

// DeleteIfExists deletes the Transit key from Vault.
func (ve *VaultTransitKeyEndpoint) DeleteIfExists(ctx context.Context) error {
	log := log.FromContext(ctx)
	vaultClient := VaultClientFromContext(ctx)
	_, err := vaultClient.Logical().Delete(ve.transitKeyObject.GetPath())
	if err != nil {
		if respErr, ok := err.(*vault.ResponseError); ok {
			if respErr.StatusCode == 404 {
				return nil
			}
		}
		log.Error(err, "unable to delete object at", "path", ve.transitKeyObject.GetPath())
		return err
	}
	return nil
}
