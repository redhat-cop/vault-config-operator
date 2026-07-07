# Quay Secret Engine

[vault-plugin-secrets-quay](https://github.com/redhat-cop/vault-plugin-secrets-quay)

## Overview

The Quay secret engine is a third-party plugin (vault-plugin-secrets-quay) that manages Quay robot accounts and their credentials. It can dynamically create robot accounts with scoped repository and team permissions, or manage credentials for pre-existing static robot accounts.

The vault-config-operator supports the following CRDs for the Quay engine:

- [QuaySecretEngineConfig](#quaysecretengineconfig)
- [QuaySecretEngineRole](#quaysecretenginerole)
- [QuaySecretEngineStaticRole](#quaysecretenginestaticrole)

## QuaySecretEngineConfig

The `QuaySecretEngineConfig` CRD allows you to configure a [Quay secret engine](https://github.com/redhat-cop/vault-plugin-secrets-quay#config) connection.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: QuaySecretEngineConfig
metadata:
  name: quay-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: quay
  url: https://quay.example.com
  disableSslVerification: false
  rootCredentials:
    secret:
      name: quay-admin-token
    passwordKey: token
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the Quay secret engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/config \
    url="https://quay.example.com" \
    token=<retrieved from credentials> \
    disable_ssl_verification=false
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the Quay secret engine. Full path: `[namespace/]{path}/config` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| url | string | Yes | — | URL of the Quay instance |
| caCertificate | string | No | — | PEM-encoded CA cert for TLS communication with Quay |
| disableSslVerification | bool | No | `false` | Disable SSL verification when communicating with Quay |
| rootCredentials | object | Yes | — | Credential source for the Quay admin token. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified. See [Credential Resolution](#credential-resolution) |

> **Note:** Deleting the `QuaySecretEngineConfig` CR does **not** remove the config from Vault. The configuration can only be removed by deleting the entire engine mount.

## QuaySecretEngineRole

The `QuaySecretEngineRole` CRD allows you to create a [Quay secret engine role](https://github.com/redhat-cop/vault-plugin-secrets-quay#roles) for generating dynamic robot account credentials with TTL-based expiration.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: QuaySecretEngineRole
metadata:
  name: my-quay-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: quay
  namespaceName: my-org
  namespaceType: organization
  createRepositories: false
  defaultPermission: read
  TTL: 1h
  maxTTL: 24h
  teams:
    dev-team: member
    ops-team: admin
  repositories:
    my-repo: write
    shared-repo: read
```

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/roles/<role-name> \
    namespace_name=my-org \
    namespace_type=organization \
    create_repositories=false \
    default_permission=read \
    ttl=1h \
    max_ttl=24h \
    teams='{"dev-team":"member","ops-team":"admin"}' \
    repositories='{"my-repo":"write","shared-repo":"read"}'
```

> **Note:** The `teams` and `repositories` fields are serialized as JSON strings in the Vault API. The operator handles this conversion automatically.

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the Quay secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | — | Override the Vault object name. Defaults to `metadata.name` |
| namespaceName | string | Yes | — | Name of the Quay organization or user |
| namespaceType | string | No | `"organization"` | Type of account namespace to manage. Allowed values: `organization`, `user` |
| createRepositories | *bool | No | `false` | Allow the robot account to create new repositories |
| defaultPermission | string | No | — | Permission granted to the robot account in newly created repositories. Allowed values: `admin`, `read`, `write` |
| teams | map[string]TeamRole | No | — | Team permissions for the robot account. Maps team names to roles. Allowed values: `admin`, `creator`, `member` |
| repositories | map[string]Permission | No | — | Repository permissions for the robot account. Maps repository names to permissions. Allowed values: `admin`, `read`, `write` |
| TTL | duration | No | — | Time-to-live for the generated credential |
| maxTTL | duration | No | — | Maximum time-to-live for the generated credential |

## QuaySecretEngineStaticRole

The `QuaySecretEngineStaticRole` CRD allows you to create a [Quay secret engine static role](https://github.com/redhat-cop/vault-plugin-secrets-quay#static-roles) for managing credentials of a fixed robot account (no TTL-based expiration).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: QuaySecretEngineStaticRole
metadata:
  name: my-quay-static-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: quay
  namespaceName: my-org
  namespaceType: organization
  createRepositories: false
  repositories:
    my-repo: write
```

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/static-roles/<role-name> \
    namespace_name=my-org \
    namespace_type=organization \
    create_repositories=false \
    repositories='{"my-repo":"write"}'
```

> **Note:** To read static credentials, use `vault read {path}/static-creds/{name}`.

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the Quay secret engine. Full Vault path: `[namespace/]{path}/static-roles/{name}` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | — | Override the Vault object name. Defaults to `metadata.name` |
| namespaceName | string | Yes | — | Name of the Quay organization or user |
| namespaceType | string | No | `"organization"` | Type of account namespace to manage. Allowed values: `organization`, `user` |
| createRepositories | *bool | No | `false` | Allow the robot account to create new repositories |
| defaultPermission | string | No | — | Permission granted to the robot account in newly created repositories. Allowed values: `admin`, `read`, `write` |
| teams | map[string]TeamRole | No | — | Team permissions for the robot account. Maps team names to roles. Allowed values: `admin`, `creator`, `member` |
| repositories | map[string]Permission | No | — | Repository permissions for the robot account. Maps repository names to permissions. Allowed values: `admin`, `read`, `write` |

## Credential Resolution

The admin token for the Quay connection can be retrieved in three different ways via the `rootCredentials` field. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified — the webhook rejects manifests that set none or more than one.

> **Note:** Quay authentication uses a token (not username/password). The `passwordKey` field controls which key the operator reads the token from. It defaults to `"password"`, so you must set `passwordKey: token` (or the appropriate key name) when using a Kubernetes secret or Vault secret whose token is stored under a different key.

### Using a Kubernetes Secret

Specify the `secret` field. If the secret is updated, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    secret:
      name: quay-admin-token
    passwordKey: token
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve the token from another Vault path.

```yaml
spec:
  rootCredentials:
    vaultSecret:
      path: secret/data/quay-credentials
    passwordKey: token
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    randomSecret:
      name: quay-random-token
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [vault-plugin-secrets-quay](https://github.com/redhat-cop/vault-plugin-secrets-quay) — Plugin documentation and source
