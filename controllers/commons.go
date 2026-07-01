package controllers

import (
	"context"

	//"github.com/redhat-cop/operator-utils/pkg/util"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type VaultAuthenticableResource interface {
	client.Object
	vaultutils.VaultObject
}

func prepareContext(ctx context.Context, r vaultresourcecontroller.ReconcilerBase, VAR VaultAuthenticableResource) (context.Context, error) {
	rlog := log.FromContext(ctx)
	ctx = vaultutils.ContextWithKubeClient(ctx, r.GetClient())
	ctx = vaultutils.ContextWithRestConfig(ctx, r.GetRestConfig())
	ctx = vaultutils.ContextWithVaultConnection(ctx, VAR.GetVaultConnection())
	vaultClient, err := VAR.GetKubeAuthConfiguration().GetVaultClient(ctx, VAR.GetNamespace())
	if err != nil {
		rlog.Error(err, "unable to create vault client", "KubeAuthConfiguration", VAR.GetKubeAuthConfiguration(), "namespace", VAR.GetNamespace())
		return nil, err
	}
	ctx = vaultutils.ContextWithVaultClient(ctx, vaultClient)
	return ctx, nil
}
