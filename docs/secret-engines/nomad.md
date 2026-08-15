# Nomad Secret Engine

[Nomad secret engine documentation](https://developer.hashicorp.com/vault/docs/secrets/nomad)

## Overview

The Nomad secret engine generates [Nomad](https://www.nomadproject.io/) ACL tokens dynamically based on pre-existing Nomad ACL policies. This allows applications to obtain short-lived Nomad tokens on demand, eliminating the need for long-lived management tokens.

The vault-config-operator supports the following CRDs for the Nomad engine:

- [NomadSecretEngineConfig](#nomadsecretengineconfig)
- [NomadSecretEngineRole](#nomadsecretenginerole)

## NomadSecretEngineConfig

The `NomadSecretEngineConfig` CRD allows you to configure a [Nomad secret engine](https://developer.hashicorp.com/vault/api-docs/secret/nomad#configure-access).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: NomadSecretEngineConfig
metadata:
  name: nomad-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: nomad
  address: "http://nomad.example.com:4646"
  rootCredentials:
    secret:
      name: nomad-mgmt-token
    passwordKey: token
```

### Vault CLI Equivalent

```shell
vault write [namespace/]{path}/config/access \
    address="http://nomad.example.com:4646" \
    token=<retrieved from credentials>
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the secret engine being configured. They may point to different mounts.

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/config/access` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| address | string | Yes | Nomad instance address as `"protocol://host:port"` (e.g., `"http://127.0.0.1:4646"`) |
| maxTokenNameLength | int | No | Max length for generated Nomad token names. 0 uses Nomad's default |
| caCert | string | No | CA certificate for verifying the Nomad server certificate (x509 PEM) |
| clientCert | string | No | Client certificate for Nomad TLS communication (x509 PEM). Must set clientKey as well |
| clientKey | string | No | Client key for Nomad TLS communication (x509 PEM). Must set clientCert as well |
| rootCredentials | object | Yes | How to retrieve the Nomad management token. See [Credential Resolution](#credential-resolution) |

## NomadSecretEngineRole

The `NomadSecretEngineRole` CRD allows you to create a [Nomad secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/nomad#create-update-role).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: NomadSecretEngineRole
metadata:
  name: nomad-readonly
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: nomad
  policies:
    - readonly
  type: client
```

### Vault CLI Equivalent

```shell
vault write [namespace/]{path}/role/nomad-readonly \
    policies="readonly" \
    type="client"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/role/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name (defaults to metadata.name) |
| policies | []string | No | Nomad ACL policies to assign to the generated token |
| global | bool | No | Whether the token should be global (replicated across Nomad regions). Default: false |
| type | string | No | Token type: `"client"` or `"management"`. Default: `"client"` |

## Credential Resolution

The Nomad management token can be retrieved in three different ways:

### Using a Kubernetes Secret

Specify the `rootCredentials.secret` field. The secret must contain a key matching `passwordKey` (default: `"password"`). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    secret:
      name: nomad-mgmt-token
    passwordKey: token
```

### Using a Vault Secret

Specify the `rootCredentials.vaultSecret` field to retrieve the token from another Vault path.

```yaml
spec:
  rootCredentials:
    vaultSecret:
      path: secret/data/nomad-token
    passwordKey: token
```

### Using a RandomSecret

Specify the `rootCredentials.randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    randomSecret:
      name: nomad-random-token
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Nomad Secrets API](https://developer.hashicorp.com/vault/api-docs/secret/nomad) — Vault API reference
