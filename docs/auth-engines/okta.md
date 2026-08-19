# Okta Auth Engine

[Okta auth engine documentation](https://developer.hashicorp.com/vault/docs/auth/okta)

## Overview

Okta provides authentication using the Okta identity provider. It enables users to log in to Vault using their Okta credentials and have Vault policies mapped based on their Okta group membership.

The vault-config-operator supports the following CRDs for the Okta engine:

- [OktaAuthEngineConfig](#oktaauthengineconfig)
- [OktaAuthEngineGroup](#oktaauthenginegroup)

## OktaAuthEngineConfig

The `OktaAuthEngineConfig` CRD allows you to configure an authentication engine mount of [type Okta](https://developer.hashicorp.com/vault/api-docs/auth/okta).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: OktaAuthEngineConfig
metadata:
  name: okta-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: okta
  orgName: my-org
  baseURL: okta.com
  oktaCredentials:
    secret:
      name: okta-api-token
    passwordKey: api_token
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/okta/config \
    org_name=my-org \
    base_url=okta.com \
    api_token=<retrieved dynamically>
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the Okta auth engine is mounted. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](contributing-vault-apis.md) |
| orgName | string | Yes | Name of the organization to be used in the Okta API |
| baseURL | string | No | Base domain for API requests. Defaults to "okta.com". Other values: "oktapreview.com", "okta-emea.com" |
| bypassOktaMFA | bool | No | Bypass Okta MFA request. Useful when using Vault's built-in MFA |
| tokenTTL | string | No | Incremental lifetime for generated tokens |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | Policies to encode onto generated tokens |
| tokenBoundCIDRs | []string | No | CIDR blocks restricting authentication |
| tokenExplicitMaxTTL | string | No | Hard cap max TTL for tokens |
| tokenNoDefaultPolicy | bool | No | Exclude default policy from generated tokens |
| tokenNumUses | int | No | Max token uses (0 = unlimited) |
| tokenPeriod | string | No | Maximum allowed period for periodic tokens |
| tokenType | string | No | Token type: service, batch, default, default-service, default-batch |
| oktaCredentials | object | No | Credential source for the Okta API token |

## OktaAuthEngineGroup

The `OktaAuthEngineGroup` CRD allows you to create an [Okta group-to-policy mapping](https://developer.hashicorp.com/vault/api-docs/auth/okta#register-group).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: OktaAuthEngineGroup
metadata:
  name: okta-admins
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: okta
  name: admins
  policies: "admin,reader"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/okta/groups/admins \
    policies="admin,reader"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the Okta auth mount where the group will be created |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](contributing-vault-apis.md) |
| name | string | Yes | Name of the Okta group |
| policies | string | No | Comma-separated list of policies associated with the group |

## Credential Resolution

The Okta API token can be retrieved in three different ways:

### Using a Kubernetes Secret

```yaml
spec:
  oktaCredentials:
    secret:
      name: okta-api-token
    passwordKey: api_token
```

### Using a Vault Secret

```yaml
spec:
  oktaCredentials:
    vaultSecret:
      path: secret/data/okta-credentials
    passwordKey: api_token
```

### Using a RandomSecret

```yaml
spec:
  oktaCredentials:
    randomSecret:
      name: okta-random-token
    passwordKey: api_token
```

## See Also

- [Authentication](auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Okta Auth Method API](https://developer.hashicorp.com/vault/api-docs/auth/okta) — Vault API reference
