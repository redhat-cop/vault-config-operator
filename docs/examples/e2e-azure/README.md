# Azure Auth + Azure Dynamic Credentials — End-to-End Example

## Scenario

This example demonstrates a complete Vault configuration where Azure VMs authenticate using their managed identity and then request dynamic Azure service principal credentials for cross-subscription resource access. The Azure auth engine validates the VM's identity, and a Policy connects the authenticated identity to the Azure secret engine credential paths.

## Prerequisites

- A running Vault instance with the vault-config-operator deployed
- The operator's Kubernetes auth configured at `kubernetes` with a `policy-admin` role
- A Kubernetes Secret `azure-auth-credentials` (type `kubernetes.io/basic-auth`) with `clientid` and `clientsecret` keys for the Azure auth engine
- A Kubernetes Secret `azure-se-credentials` (type `kubernetes.io/basic-auth`) with `clientid` and `clientsecret` keys for the Azure secret engine (must have permissions to create service principals)

## Resources Created

| # | Kind | Name | Purpose |
|---|------|------|---------|
| 1 | AuthEngineMount | azure-auth | Mounts Azure auth at `infra/azure-auth` |
| 2 | AzureAuthEngineConfig | azure-auth-config | Configures Azure auth with tenant/credentials |
| 3 | Policy | azure-sp-reader | Grants access to dynamic Azure credentials |
| 4 | SecretEngineMount | azure-se | Mounts Azure secret engine at `infra/azure-se` |
| 5 | AzureSecretEngineConfig | azure-se-config | Configures Azure secret engine subscription |
| 6 | AzureSecretEngineRole | contributor-role | Defines dynamic SP credential parameters |
| 7 | AzureAuthEngineRole | azure-vm-role | Auth role linking VMs to the policy |

> **Path composition:** `AuthEngineMount` and `SecretEngineMount` compose their Vault mount path as `{spec.path}/{metadata.name}`. This example uses `path: infra` as a grouping prefix, producing auth mount `infra/azure-auth` and secret mount `infra/azure-se`.

## Apply

```shell
kubectl apply -f e2e-azure.yaml
```

## Verify

```shell
# Verify auth engine is mounted
vault auth list | grep infra/azure-auth

# Verify secret engine is configured
vault read infra/azure-se/config

# Verify credential role exists
vault read infra/azure-se/roles/contributor-role

# Generate dynamic credentials (after authenticating via Azure auth)
vault read infra/azure-se/creds/contributor-role
```

## How It Works

1. **Auth mount** — The Azure auth engine at `infra/azure-auth` validates JWTs from Azure IMDS against Azure AD.
2. **Policy bridge** — The `azure-sp-reader` policy grants `read` access to `infra/azure-se/creds/contributor-role`.
3. **Secret engine** — The Azure secret engine at `infra/azure-se` creates dynamic service principals with Contributor role assignments.
4. **Auth role** — The `azure-vm-role` binds Azure subscription and resource group to the `azure-sp-reader` policy, completing the auth-to-secrets flow.

> **Note:** The auth mount (`infra/azure-auth`) and secret engine mount (`infra/azure-se`) use distinct resource names under the shared `infra` prefix — both are Azure type engines, so different names prevent path collision.

## Cleanup

```shell
kubectl delete -f e2e-azure.yaml
```

> **Note:** Deleting the CRs removes the Kubernetes objects. The Azure auth and secret engine configuration persist in Vault until their mounts are manually deleted via `vault secrets disable infra/azure-se` and `vault auth disable infra/azure-auth`.
