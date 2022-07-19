package utils

import (
	"context"
	"errors"

	vault "github.com/hashicorp/vault/api"
)

type VaultSecretObject interface {
	GetPath() string
	GetRequestMethod() string
	GetPostRequestPayload() map[string]string
}

func NewVaultSecretEndpoint(obj VaultSecretObject) *VaultSecretEndpoint {
	return &VaultSecretEndpoint{
		vaultSecretObject: obj,
	}
}

type VaultSecretEndpoint struct {
	vaultSecretObject VaultSecretObject
}

func (ve *VaultSecretEndpoint) GetSecret(context context.Context) (*vault.Secret, bool, error) {
	if ve.vaultSecretObject.GetRequestMethod() == "GET" {
		return ReadSecret(context, ve.vaultSecretObject.GetPath())
	}
	if ve.vaultSecretObject.GetRequestMethod() == "POST" {
		return ReadSecretWithPayload(context, ve.vaultSecretObject.GetPath(), ve.vaultSecretObject.GetPostRequestPayload())
	}
	return nil, false, errors.New("unknown request method:" + ve.vaultSecretObject.GetRequestMethod())
}
