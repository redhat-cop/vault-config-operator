package controller

import (
	"context"

	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
)

// VaultAuthenticableResource is an alias for backward compatibility.
type VaultAuthenticableResource = vaultresourcecontroller.VaultAuthenticableResource

// prepareContext delegates to the canonical implementation in vaultresourcecontroller.
func prepareContext(ctx context.Context, r vaultresourcecontroller.ReconcilerBase, VAR VaultAuthenticableResource) (context.Context, error) {
	return vaultresourcecontroller.PrepareContext(ctx, r, VAR)
}
