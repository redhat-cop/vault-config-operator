package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/redhat-cop/operator-utils/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"
)

type VaultResource struct {
	VaultEndpoint
	util.ReconcilerBase
	Log logr.Logger
}

func (r *VaultResource) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

}
