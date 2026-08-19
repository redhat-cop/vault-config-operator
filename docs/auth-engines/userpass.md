# Userpass Auth Engine

[Userpass engine documentation](https://developer.hashicorp.com/vault/docs/auth/userpass)

## Overview

Userpass provides username/password-based authentication to Vault. It enables human users or service accounts to authenticate using a traditional username and password pair. Unlike most auth engines, Userpass has no separate configuration endpoint — the mount itself (via `AuthEngineMount`) is the configuration. Users are the only API resource managed by the operator.

The vault-config-operator supports the following CRD for the Userpass engine:

- [UserpassAuthEngineUser](#userpassauthengineuser)

## UserpassAuthEngineUser

The `UserpassAuthEngineUser` CRD allows you to create a [Userpass user](https://developer.hashicorp.com/vault/api-docs/auth/userpass#create-update-user).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: UserpassAuthEngineUser
metadata:
  name: my-app-user
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: userpass
  passwordCredentials:
    secret:
      name: my-user-password
    passwordKey: password
  tokenPolicies:
    - default
    - app-policy
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/{path}/users/{name} \
    password=<retrieved from K8s Secret> \
    token_policies="default,app-policy" \
    token_ttl=1h \
    token_max_ttl=24h
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the userpass engine is mounted. Full Vault path: `[namespace/]auth/{path}/users/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault username. Defaults to `metadata.name`. Immutable after creation. |
| passwordCredentials | object | Yes | Credential source for the user's password. See [Credential Resolution](#credential-resolution) below. |
| tokenTTL | string | No | Incremental lifetime for generated tokens (e.g. "1h") |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens (e.g. "24h") |
| tokenPolicies | []string | No | Token policies to assign to the user |
| tokenBoundCIDRs | []string | No | CIDR blocks restricting token use |
| tokenExplicitMaxTTL | string | No | Hard cap max TTL for tokens |
| tokenNoDefaultPolicy | bool | No | Exclude the default policy from generated tokens |
| tokenNumUses | int | No | Maximum number of token uses (0 = unlimited) |
| tokenPeriod | string | No | Renewal period for periodic tokens |
| tokenType | string | No | Token type: `service`, `batch`, `default`, `default-service`, `default-batch` |

## Credential Resolution

The password for a Userpass user must be provided via a credential source — it is never specified inline in the CR spec. The operator resolves the password at reconcile time from one of three sources.

### Using a Kubernetes Secret

```yaml
spec:
  passwordCredentials:
    secret:
      name: my-user-password
    passwordKey: password
```

### Using a Vault Secret

```yaml
spec:
  passwordCredentials:
    vaultSecret:
      path: secret/data/user-credentials
    passwordKey: password
```

### Using a RandomSecret

```yaml
spec:
  passwordCredentials:
    randomSecret:
      name: my-random-password
    passwordKey: password
```

When using a Kubernetes Secret or RandomSecret, the controller watches for changes and re-reconciles the user when the password source is updated (password rotation).

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
