# Azure Auth Engine

[Azure engine documentation](https://developer.hashicorp.com/vault/docs/auth/azure)

## Overview

The Azure auth method allows authentication for Azure-hosted entities using Azure Active Directory credentials. VMs, VM Scale Sets, and other Azure resources present a signed JWT from the Azure Instance Metadata Service (IMDS), which Vault verifies against Azure AD. This enables Azure workloads to authenticate to Vault without managing static credentials.

The vault-config-operator supports the following CRDs for the Azure engine:

- [AzureAuthEngineConfig](#azureauthengineconfig)
- [AzureAuthEngineRole](#azureauthenginerole)

## AzureAuthEngineConfig

The `AzureAuthEngineConfig` CRD allows you to configure an [Azure auth engine](https://developer.hashicorp.com/vault/api-docs/auth/azure#configure).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AzureAuthEngineConfig
metadata:
  name: azure-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: azure
  tenantID: 00000000-0000-0000-0000-000000000000
  resource: https://management.azure.com/
  environment: AzurePublicCloud
  maxRetries: 3
  maxRetryDelay: 60
  retryDelay: 4
  azureCredentials:
    secret:
      name: azure-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/config \
    tenant_id="00000000-0000-0000-0000-000000000000" \
    resource="https://management.azure.com/" \
    environment="AzurePublicCloud" \
    client_id="<retrieved from azureCredentials>" \
    client_secret="<retrieved from azureCredentials>" \
    max_retries=3 \
    max_retry_delay=60 \
    retry_delay=4
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path for the Azure auth engine. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| azureCredentials | object | No | — | Credential resolution for Azure Client ID and Client Secret. See [Credential Resolution](#credential-resolution) |
| tenantID | string | Yes | — | Tenant ID for the Azure Active Directory organization. Can also be provided via the `AZURE_TENANT_ID` environment variable |
| resource | string | Yes | — | Resource URL for the application registered in Azure AD. Must match the audience (`aud` claim) of the JWT provided to the login API. Can also be provided via the `AZURE_AD_RESOURCE` environment variable |
| environment | string | No | `"AzurePublicCloud"` | Azure cloud environment. Allowed values: `AzurePublicCloud`, `AzureUSGovernmentCloud`, `AzureChinaCloud`, `AzureGermanCloud`. Can also be provided via the `AZURE_ENVIRONMENT` environment variable |
| clientID | string | No | — | Client ID for credentials to query Azure APIs. If set directly, takes precedence over the client ID retrieved from `azureCredentials`. Can also be provided via the `AZURE_CLIENT_ID` environment variable |
| maxRetries | int64 | No | `3` | Maximum number of attempts a failed operation will be retried before producing an error |
| maxRetryDelay | int64 | No | `60` | Maximum delay (seconds) allowed before retrying an operation |
| retryDelay | int64 | No | `4` | Initial delay (seconds) before retrying an operation. Increases exponentially |

## AzureAuthEngineRole

The `AzureAuthEngineRole` CRD allows you to create an [Azure auth role](https://developer.hashicorp.com/vault/api-docs/auth/azure#create-update-role).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AzureAuthEngineRole
metadata:
  name: azure-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: azure
  name: my-azure-role
  boundSubscriptionIDs:
    - 00000000-0000-0000-0000-000000000000
  boundResourceGroups:
    - my-resource-group
  boundServicePrincipalIDs:
    - 11111111-1111-1111-1111-111111111111
  tokenPolicies:
    - app-policy
    - default
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/role/<name> \
    bound_subscription_ids="00000000-0000-0000-0000-000000000000" \
    bound_resource_groups="my-resource-group" \
    bound_service_principal_ids="11111111-1111-1111-1111-111111111111" \
    token_policies="app-policy,default" \
    token_ttl="1h" \
    token_max_ttl="24h"
```

### Field Descriptions

**Binding Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Azure auth mount path. Full path: `[namespace/]auth/{path}/role/{name}` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | Yes | — | Name of the role in Vault |
| boundServicePrincipalIDs | []string | No | — | Service Principal IDs that login is restricted to |
| boundGroupIDs | []string | No | — | Group IDs that login is restricted to |
| boundLocations | []string | No | — | Locations that login is restricted to |
| boundSubscriptionIDs | []string | No | — | Subscription IDs that login is restricted to |
| boundResourceGroups | []string | No | — | Resource groups that login is restricted to |
| boundScaleSets | []string | No | — | Scale set names that login is restricted to |

**Token Parameters:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| tokenTTL | string | No | — | Incremental lifetime for generated tokens. Referenced at renewal time |
| tokenMaxTTL | string | No | — | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | — | Policies to attach to generated tokens |
| policies | []string | No | — | **DEPRECATED:** Use `tokenPolicies` instead |
| tokenBoundCIDRs | []string | No | — | CIDR blocks restricting token usage |
| tokenExplicitMaxTTL | string | No | — | Hard cap TTL overriding `tokenTTL` and `tokenMaxTTL` |
| tokenNoDefaultPolicy | bool | No | `false` | Exclude the `default` policy from generated tokens |
| tokenNumUses | int64 | No | `0` | Maximum token usage count. `0` means unlimited |
| tokenPeriod | int64 | No | `0` | Maximum allowed period value when a periodic token is requested |
| tokenType | string | No | — | Token type. Allowed values: `service`, `batch`, `default`, `default-service`, `default-batch` |

## Credential Resolution

The Azure Client ID and Client Secret (used to query Azure APIs for verifying instance metadata) can be retrieved in three different ways via the `azureCredentials` field. If `azureCredentials` is omitted or left empty, the operator will use Azure environment variables (e.g., `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`).

> **Note:** If `clientID` is set directly in the spec, it takes precedence over the username (client ID) retrieved from the referenced secret. The password (client secret) is always retrieved from the credential source.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  azureCredentials:
    secret:
      name: azure-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve credentials from another Vault path.

```yaml
spec:
  azureCredentials:
    vaultSecret:
      path: secret/data/azure-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated. A `clientID` must be specified directly in the spec when using RandomSecret — the client ID is not read from the generated secret.

```yaml
spec:
  clientID: 00000000-0000-0000-0000-000000000000
  azureCredentials:
    randomSecret:
      name: azure-random-secret
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Azure Auth Method](https://developer.hashicorp.com/vault/docs/auth/azure) — Vault documentation
- [Vault Azure Auth API](https://developer.hashicorp.com/vault/api-docs/auth/azure) — Vault API reference
