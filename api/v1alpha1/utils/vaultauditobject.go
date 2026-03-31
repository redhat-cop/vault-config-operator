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
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type VaultAuditObject interface {
	VaultObject
}

type VaultAuditEndpoint struct {
	*VaultEndpoint
	vaultAuditObject VaultAuditObject
}

func NewVaultAuditEndpoint(obj client.Object) *VaultAuditEndpoint {
	return &VaultAuditEndpoint{
		vaultAuditObject: obj.(VaultAuditObject),
		VaultEndpoint:    NewVaultEndpoint(obj),
	}
}

// Exists checks if the audit device is currently enabled
func (ve *VaultAuditEndpoint) Exists(context context.Context) (bool, error) {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)

	audits, err := vaultClient.Sys().ListAudit()
	if err != nil {
		log.Error(err, "unable to list audit devices")
		return false, err
	}

	// The path in the list has a trailing slash, so we need to check both formats
	path := ve.vaultObject.GetPath()
	// Extract just the audit device name from sys/audit/<name>
	auditName := path[len("sys/audit/"):]

	_, exists := audits[auditName+"/"]
	return exists, nil
}

// Enable enables the audit device
func (ve *VaultAuditEndpoint) Enable(context context.Context) error {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)

	path := ve.vaultObject.GetPath()
	// Extract just the audit device name from sys/audit/<name>
	auditName := path[len("sys/audit/"):]

	payload := ve.vaultObject.GetPayload()

	options := &vault.EnableAuditOptions{
		Type:        payload["type"].(string),
		Description: payload["description"].(string),
		Local:       payload["local"].(bool),
		Options:     payload["options"].(map[string]string),
	}

	err := vaultClient.Sys().EnableAuditWithOptions(auditName, options)
	if err != nil {
		log.Error(err, "unable to enable audit device", "path", auditName)
		return err
	}

	log.Info("audit device enabled successfully", "path", auditName)
	return nil
}

// Disable disables the audit device
func (ve *VaultAuditEndpoint) Disable(context context.Context) error {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)

	path := ve.vaultObject.GetPath()
	// Extract just the audit device name from sys/audit/<name>
	auditName := path[len("sys/audit/"):]

	err := vaultClient.Sys().DisableAudit(auditName)
	if err != nil {
		if respErr, ok := err.(*vault.ResponseError); ok {
			if respErr.StatusCode == 404 {
				// Already disabled
				return nil
			}
		}
		log.Error(err, "unable to disable audit device", "path", auditName)
		return err
	}

	log.Info("audit device disabled successfully", "path", auditName)
	return nil
}

// IsEquivalentToDesired checks if the current audit device configuration matches the desired state
func (ve *VaultAuditEndpoint) IsEquivalentToDesired(context context.Context) (bool, error) {
	log := log.FromContext(context)
	vaultClient := context.Value("vaultClient").(*vault.Client)

	audits, err := vaultClient.Sys().ListAudit()
	if err != nil {
		log.Error(err, "unable to list audit devices")
		return false, err
	}

	path := ve.vaultObject.GetPath()
	// Extract just the audit device name from sys/audit/<name>
	auditName := path[len("sys/audit/"):]

	currentAudit, exists := audits[auditName+"/"]
	if !exists {
		return false, nil
	}

	// Convert current audit to a comparable format
	currentPayload := map[string]interface{}{
		"type":        currentAudit.Type,
		"description": currentAudit.Description,
		"local":       currentAudit.Local,
		"options":     currentAudit.Options,
	}

	return ve.vaultObject.IsEquivalentToDesiredState(currentPayload), nil
}

// CreateOrUpdate enables or updates the audit device configuration
func (ve *VaultAuditEndpoint) CreateOrUpdate(context context.Context) error {
	log := log.FromContext(context)

	exists, err := ve.Exists(context)
	if err != nil {
		log.Error(err, "unable to check if audit device exists")
		return err
	}

	if !exists {
		// Enable the audit device
		return ve.Enable(context)
	}

	// Check if the configuration matches
	equivalent, err := ve.IsEquivalentToDesired(context)
	if err != nil {
		log.Error(err, "unable to check if audit device is equivalent")
		return err
	}

	if !equivalent {
		// Audit devices cannot be updated in place, we need to disable and re-enable
		log.Info("audit device configuration changed, disabling and re-enabling")
		err = ve.Disable(context)
		if err != nil {
			return fmt.Errorf("unable to disable audit device for update: %w", err)
		}

		return ve.Enable(context)
	}

	log.Info("audit device already configured correctly")
	return nil
}

// DeleteIfExists disables the audit device if it exists
func (ve *VaultAuditEndpoint) DeleteIfExists(context context.Context) error {
	exists, err := ve.Exists(context)
	if err != nil {
		return err
	}

	if exists {
		return ve.Disable(context)
	}

	return nil
}
