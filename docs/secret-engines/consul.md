# Consul Secret Engine

[Consul engine documentation](https://developer.hashicorp.com/vault/docs/secrets/consul)

## Overview

The Consul secret engine generates Consul ACL tokens on demand, with automatic revocation after the configured TTL expires. It enables applications to access Consul services with dynamically generated, short-lived tokens instead of static management tokens, improving security posture through credential rotation.

The vault-config-operator supports the following CRDs for the Consul engine:

- [ConsulSecretEngineConfig](#consulsecretengineconfig)
- [ConsulSecretEngineRole](#consulsecretenginerole)

## ConsulSecretEngineConfig

The `ConsulSecretEngineConfig` CRD allows you to configure a [Consul secret engine access](https://developer.hashicorp.com/vault/api-docs/secret/consul#configure-access).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: ConsulSecretEngineConfig
metadata:
  name: consul-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: consul
  address: "consul.service.consul:8500"
  scheme: http
  rootCredentials:
    secret:
      name: consul-management-token
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the secret engine being configured. They may point to different mounts.

> **Note:** Deleting the `ConsulSecretEngineConfig` CR from Kubernetes will **not** delete the configuration from Vault (`IsDeletable=false`). The Vault configuration persists until the engine mount is removed.

### Vault CLI Equivalent

```shell
vault write [namespace/]consul/config/access \
    address="consul.service.consul:8500" \
    scheme=http \
    token=<retrieved from credentials>
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/config/access` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| address | string | Yes | Consul instance address as `host:port` |
| scheme | string | No | URL scheme: `http` or `https`. Default: `http` |
| caCert | string | No | CA certificate for verifying the Consul server certificate (PEM encoded) |
| clientCert | string | No | Client certificate for Consul mutual TLS (PEM encoded). If set, `clientKey` must also be set |
| clientKey | string | No | Client key for Consul mutual TLS (PEM encoded). If set, `clientCert` must also be set |
| rootCredentials | object | Yes | Credential source for the Consul ACL management token. See [Credential Resolution](#credential-resolution) |

## ConsulSecretEngineRole

The `ConsulSecretEngineRole` CRD allows you to create a [Consul secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/consul#create-update-role) for generating dynamic Consul ACL tokens.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: ConsulSecretEngineRole
metadata:
  name: consul-read-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: consul
  consulPolicies:
    - readonly
  ttl: "1h"
  maxTTL: "24h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]consul/roles/consul-read-role \
    consul_policies="readonly" \
    ttl=1h \
    max_ttl=24h
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| consulPolicies | []string | No | List of Consul ACL policies to assign to the generated token |
| consulRoles | []string | No | List of Consul roles to attach to the generated token (Consul 1.5+) |
| serviceIdentities | []string | No | List of Consul service identities to assign (Consul 1.5+) |
| nodeIdentities | []string | No | List of Consul node identities to assign (Consul 1.8+) |
| consulNamespace | string | No | Consul Enterprise namespace for the token (Consul 1.7+) |
| partition | string | No | Consul admin partition for the token (Consul 1.11+) |
| local | bool | No | If `true`, creates a token that is not replicated globally (Consul 1.4+) |
| ttl | string | No | TTL for the generated Consul token (duration format) |
| maxTTL | string | No | Maximum TTL for the generated Consul token (duration format) |

## Credential Resolution

The Consul ACL management token can be retrieved in three different ways via the `rootCredentials` field. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    secret:
      name: consul-management-token
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve the management token from another Vault path.

```yaml
spec:
  rootCredentials:
    vaultSecret:
      path: secret/data/consul-token
    passwordKey: password
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    randomSecret:
      name: consul-random-token
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Consul Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/consul) — Vault documentation
- [Vault Consul Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/consul) — Vault API reference
