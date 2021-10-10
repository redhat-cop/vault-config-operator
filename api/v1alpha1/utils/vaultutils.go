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

	"github.com/go-logr/logr"
	vault "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type VaultConnection interface {
	GetNamespace() string
	GetRole() string
	GetKubeAuthPath() string
	GetServiceAccountName() string
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

func NewVaultEndpoint(context context.Context, vaultConnection VaultConnection, vaultObject VaultObject, kubeNamespace string, kubeClient client.Client, log logr.Logger) (*VaultEndpoint, error) {
	vaultEndpoint := &VaultEndpoint{
		log:             log,
		VaultConnection: vaultConnection,
		VaultObject:     vaultObject,
	}
	jwt, err := vaultEndpoint.getJWTToken(context, kubeNamespace, kubeClient)
	if err != nil {
		vaultEndpoint.log.Error(err, "unable to retrieve jwt token for", "namespace", kubeNamespace, "serviceaccount", vaultEndpoint.GetServiceAccountName())
		return nil, err
	}
	vaultClient, err := vaultEndpoint.createVaultClient(jwt)
	if err != nil {
		vaultEndpoint.log.Error(err, "unable to create vault client")
		return nil, err
	}
	vaultEndpoint.client = vaultClient
	return vaultEndpoint, nil
}

func (ve *VaultEndpoint) getJWTToken(context context.Context, kubeNamespace string, kubeClient client.Client) (string, error) {
	serviceAccount := &corev1.ServiceAccount{}
	err := kubeClient.Get(context, client.ObjectKey{
		Namespace: kubeNamespace,
		Name:      ve.GetServiceAccountName(),
	}, serviceAccount)
	if err != nil {
		ve.log.Error(err, "unable to retrieve", "service account", client.ObjectKey{
			Namespace: kubeNamespace,
			Name:      ve.GetServiceAccountName(),
		})
		return "", err
	}
	var tokenSecretName string
	for _, secretName := range serviceAccount.Secrets {
		if strings.Contains(secretName.Name, "token") {
			tokenSecretName = secretName.Name
			break
		}
	}
	if tokenSecretName == "" {
		return "", errors.New("unable to find token secret name for service account" + kubeNamespace + "/" + ve.GetServiceAccountName())
	}
	secret := &corev1.Secret{}
	err = kubeClient.Get(context, client.ObjectKey{
		Namespace: kubeNamespace,
		Name:      tokenSecretName,
	}, secret)
	if err != nil {
		ve.log.Error(err, "unable to retrieve", "secret", client.ObjectKey{
			Namespace: kubeNamespace,
			Name:      tokenSecretName,
		})
		return "", err
	}
	if jwt, ok := secret.Data["token"]; ok {
		return string(jwt), nil
	} else {
		return "", errors.New("unable to find \"token\" key in secret" + kubeNamespace + "/" + tokenSecretName)
	}
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
	secret, err := client.Logical().Write(ve.GetKubeAuthPath(), map[string]interface{}{
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
	if secret == nil {
		return nil, false, nil
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
func (ve *VaultEndpoint) DeleteIfExists() error {
	_, err := ve.GetVaultClient().Logical().Delete(ve.GetPath())
	if err != nil {
		if respErr, ok := err.(*vault.ResponseError); ok {
			if respErr.StatusCode == 404 {
				return nil
			}
		}
		ve.log.Error(err, "unable to delete object at", "path", ve.GetPath())
		return err
	}
	return nil
}
func (ve *VaultEndpoint) CreateOrUpdate() error {
	currentPayload, found, err := ve.Read()
	if err != nil {
		ve.log.Error(err, "unable to read object at", "path", ve.GetPath())
		return err
	}
	if !found {
		return ve.Write(ve.GetPayload())
	} else {
		if !ve.IsEquivalentToDesiredState(currentPayload) {
			return ve.Write(ve.GetPayload())
		}
	}
	return nil
}
