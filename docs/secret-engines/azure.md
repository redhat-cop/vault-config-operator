# Azure Secret Engine

[Azure engine documentation](https://developer.hashicorp.com/vault/docs/secrets/azure)

## Overview

The Azure secret engine dynamically generates Azure service principals and their credentials (client secrets). It creates service principals with configurable Azure role assignments and group memberships, enabling applications to obtain short-lived Azure credentials without managing static service principal secrets.

The vault-config-operator supports the following CRDs for the Azure engine:

- [AzureSecretEngineConfig](#azuresecretengineconfig)
- [AzureSecretEngineRole](#azuresecretenginerole)

## AzureSecretEngineConfig

The `AzureSecretEngineConfig` CRD allows you to configure an [Azure secret engine](https://developer.hashicorp.com/vault/api-docs/secret/azure#configure).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AzureSecretEngineConfig
metadata:
  name: azure-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: azure
  subscriptionID: 00000000-0000-0000-0000-000000000000
  tenantID: 11111111-1111-1111-1111-111111111111
  environment: AzurePublicCloud
  rootPasswordTTL: "182d"
  azureCredentials:
    secret:
      name: azure-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the Azure secret engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/config \
    subscription_id="00000000-0000-0000-0000-000000000000" \
    tenant_id="11111111-1111-1111-1111-111111111111" \
    client_id="<retrieved from azureCredentials>" \
    client_secret="<retrieved from azureCredentials>" \
    environment="AzurePublicCloud" \
    root_password_ttl="182d"
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the Azure secret engine. Full path: `[namespace/]{path}/config` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| subscriptionID | string | Yes | — | Azure subscription ID |
| tenantID | string | Yes | — | Azure Active Directory tenant ID |
| clientID | string | No | — | Client ID for credentials to query Azure APIs. If set directly, takes precedence over the client ID retrieved from `azureCredentials` |
| environment | string | No | `"AzurePublicCloud"` | Azure cloud environment. Allowed values: `AzurePublicCloud`, `AzureUSGovernmentCloud`, `AzureChinaCloud`, `AzureGermanCloud` |
| passwordPolicy | string | No | — | Name of a Vault password policy for generating passwords |
| rootPasswordTTL | string | No | `"182d"` | How long the root password is valid for in Azure when `rotate-root` generates a new client secret |
| azureCredentials | object | No | — | Credential resolution for Azure Client ID and Client Secret. See [Credential Resolution](#credential-resolution). If omitted, Azure environment variables are used |

## AzureSecretEngineRole

The `AzureSecretEngineRole` CRD allows you to create an [Azure secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/azure#create-update-role) for generating dynamic Azure service principal credentials.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AzureSecretEngineRole
metadata:
  name: azure-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: azure
  azureRoles: '[{"role_name":"Contributor","scope":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg"}]'
  TTL: 1h
  maxTTL: 24h
  persistApp: false
```

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/roles/<role-name> \
    azure_roles='[{"role_name":"Contributor","scope":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg"}]' \
    ttl=1h \
    max_ttl=24h \
    persist_app=false
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the Azure secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | — | Override the Vault object name. Defaults to `metadata.name` |
| azureRoles | string (JSON) | No | — | List of Azure roles to assign to the generated service principal. Must be a JSON-encoded string. See [Vault Azure roles docs](https://developer.hashicorp.com/vault/api-docs/secret/azure#azure_roles) |
| azureGroups | string (JSON) | No | — | List of Azure groups to assign the generated service principal to. Must be a JSON-encoded string. See [Vault Azure groups docs](https://developer.hashicorp.com/vault/api-docs/secret/azure#azure_groups) |
| applicationObjectID | string | No | — | Application Object ID for an existing service principal to use instead of creating dynamic ones. If present, `azureRoles` will be ignored |
| persistApp | bool | No | `false` | Persist the created service principal and application for the lifetime of the role |
| TTL | string | No | — | Default TTL for generated service principals |
| maxTTL | string | No | — | Maximum TTL for generated service principals |
| permanentlyDelete | string | No | — | Whether to permanently delete dynamically created Applications and Service Principals (`"true"` or `"false"` as a string). Must be `"false"` when `applicationObjectID` is present |
| signInAudience | string | No | — | Security principal types allowed to sign in. Allowed values: `AzureADMyOrg`, `AzureADMultipleOrgs`, `AzureADandPersonalMicrosoftAccount`, `PersonalMicrosoftAccount` |
| tags | string | No | — | Comma-separated Azure tags to attach to the application |

## Credential Resolution

The Azure Client ID and Client Secret (used to manage Azure service principals) can be retrieved in three different ways via the `azureCredentials` field. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified. If `azureCredentials` is omitted or left empty, the operator will use Azure environment variables (e.g., `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`).

> **Note:** If `clientID` is set directly in the spec, it takes precedence over the username (client ID) retrieved from the referenced secret. The password (client secret) is always retrieved from the credential source.

> **Key defaults:** `usernameKey` defaults to `"username"` and `passwordKey` defaults to `"password"`. Override these when the referenced secret uses different key names (e.g., `usernameKey: clientid`, `passwordKey: clientsecret`).

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
- [Vault Azure Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/azure) — Vault documentation
- [Vault Azure Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/azure) — Vault API reference
