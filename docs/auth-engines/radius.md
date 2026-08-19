# RADIUS Auth Engine

[RADIUS auth engine documentation](https://developer.hashicorp.com/vault/docs/auth/radius)

## Overview

RADIUS provides authentication by forwarding login requests to an external RADIUS server. It enables organizations to leverage existing RADIUS infrastructure for Vault authentication while managing user-to-policy mappings declaratively.

The vault-config-operator supports the following CRDs for the RADIUS engine:

- [RADIUSAuthEngineConfig](#radiusauthengineconfig)
- [RADIUSAuthEngineUser](#radiusauthengineuser)

## RADIUSAuthEngineConfig

The `RADIUSAuthEngineConfig` CRD allows you to configure an authentication engine mount of [type RADIUS](https://developer.hashicorp.com/vault/api-docs/auth/radius).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RADIUSAuthEngineConfig
metadata:
  name: radius-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: radius
  host: "radius.example.com"
  port: 1812
  dialTimeout: 10
  readTimeout: 10
  nasPort: 10
  radiusCredentials:
    secret:
      name: radius-shared-secret
    passwordKey: secret
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/radius/config \
    host=radius.example.com \
    port=1812 \
    secret=<retrieved dynamically> \
    dial_timeout=10 \
    read_timeout=10 \
    nas_port=10
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the RADIUS auth engine is mounted. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| host | string | Yes | RADIUS server hostname or IP address |
| port | int | No | UDP port the RADIUS server listens on (default: 1812) |
| unregisteredUserPolicies | string | No | Comma-separated policies for users without explicit mappings |
| dialTimeout | int | No | Seconds to wait for backend connection (default: 10) |
| readTimeout | int | No | Seconds before response times out (default: 10) |
| nasPort | int | No | NAS-Port attribute of the RADIUS request (default: 10) |
| tokenTTL | string | No | Incremental lifetime for generated tokens |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | Policies to encode onto generated tokens |
| tokenBoundCIDRs | []string | No | CIDR blocks restricting authentication |
| tokenExplicitMaxTTL | string | No | Hard cap max TTL for tokens |
| tokenNoDefaultPolicy | bool | No | Exclude default policy from tokens |
| tokenNumUses | int | No | Max token uses (0 = unlimited) |
| tokenPeriod | string | No | Maximum allowed period for periodic tokens |
| tokenType | string | No | Token type: service, batch, default, default-service, default-batch |

## RADIUSAuthEngineUser

The `RADIUSAuthEngineUser` CRD allows you to create a [RADIUS user-to-policy mapping](https://developer.hashicorp.com/vault/api-docs/auth/radius#register-user).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RADIUSAuthEngineUser
metadata:
  name: radius-user-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: radius
  name: "testuser"
  policies: "default,dev"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/radius/users/testuser \
    policies="default,dev"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the RADIUS auth engine mount |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Username for the RADIUS user. Defaults to metadata.name if not specified |
| policies | string | No | Comma-separated list of policies associated with this user |

## Credential Resolution

The RADIUS shared secret can be retrieved in three different ways:

### Using a Kubernetes Secret

```yaml
spec:
  radiusCredentials:
    secret:
      name: radius-shared-secret
    passwordKey: secret
```

### Using a Vault Secret

```yaml
spec:
  radiusCredentials:
    vaultSecret:
      path: secret/data/radius-credentials
    passwordKey: secret
```

### Using a RandomSecret

```yaml
spec:
  radiusCredentials:
    randomSecret:
      name: radius-random-secret
    passwordKey: secret
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault RADIUS Auth Method API](https://developer.hashicorp.com/vault/api-docs/auth/radius) — Vault API reference
