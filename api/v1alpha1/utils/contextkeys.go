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
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type contextKey int

const (
	KubeClientKey contextKey = iota
	RestConfigKey
	VaultConnectionKey
	VaultClientKey
)

func ContextWithKubeClient(ctx context.Context, c client.Client) context.Context {
	return context.WithValue(ctx, KubeClientKey, c)
}

func KubeClientFromContext(ctx context.Context) client.Client {
	return ctx.Value(KubeClientKey).(client.Client)
}

func ContextWithRestConfig(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, RestConfigKey, cfg)
}

func RestConfigFromContext(ctx context.Context) *rest.Config {
	return ctx.Value(RestConfigKey).(*rest.Config)
}

func ContextWithVaultConnection(ctx context.Context, vc *VaultConnection) context.Context {
	return context.WithValue(ctx, VaultConnectionKey, vc)
}

func VaultConnectionFromContext(ctx context.Context) *VaultConnection {
	return ctx.Value(VaultConnectionKey).(*VaultConnection)
}

func ContextWithVaultClient(ctx context.Context, vc *vault.Client) context.Context {
	return context.WithValue(ctx, VaultClientKey, vc)
}

func VaultClientFromContext(ctx context.Context) *vault.Client {
	return ctx.Value(VaultClientKey).(*vault.Client)
}
