# RabbitMQ Secret Engine

[RabbitMQ engine documentation](https://developer.hashicorp.com/vault/docs/secrets/rabbitmq)

## Overview

The RabbitMQ secret engine generates dynamic RabbitMQ credentials (username/password pairs) on demand, with automatic deletion after the configured lease expires. It connects to a RabbitMQ cluster's management API and creates users with specified vhost permissions and topic-level access controls.

The vault-config-operator supports the following CRDs for the RabbitMQ engine:

- [RabbitMQSecretEngineConfig](#rabbitmqsecretengineconfig)
- [RabbitMQSecretEngineRole](#rabbitmqsecretenginerole)

## RabbitMQSecretEngineConfig

The `RabbitMQSecretEngineConfig` CRD allows you to configure a [RabbitMQ secret engine connection](https://developer.hashicorp.com/vault/api-docs/secret/rabbitmq#configure-connection).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RabbitMQSecretEngineConfig
metadata:
  name: my-rabbitmq-config
spec:
  authentication:
    path: kubernetes
    role: rabbitmq-engine-admin
  path: vault-config-operator/rabbitmq
  connectionURI: https://my-rabbitmq.example.com:15672
  username: admin
  verifyConnection: true
  passwordPolicy: my-password-policy
  usernameTemplate: "v-{{.RoleName}}-{{unix_time}}"
  leaseTTL: 86400
  leaseMaxTTL: 172800
  rootCredentials:
    secret:
      name: rabbitmq-admin-password
    passwordKey: password
    usernameKey: username
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the RabbitMQ secret engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/config/connection \
    connection_uri="https://my-rabbitmq.example.com:15672" \
    username=admin \
    password=<retrieved from credentials> \
    verify_connection=true \
    password_policy=my-password-policy \
    username_template="v-{{.RoleName}}-{{unix_time}}"

vault write [namespace/]<path>/config/lease \
    ttl=86400 \
    max_ttl=172800
```

> **Note:** The operator writes to two separate Vault paths: connection config (`{path}/config/connection`) and lease config (`{path}/config/lease`). The lease write is only issued if `leaseTTL` or `leaseMaxTTL` is non-zero.

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the RabbitMQ secret engine. Connection config path: `[namespace/]{path}/config/connection`. Lease config path: `{path}/config/lease` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| connectionURI | string | Yes | RabbitMQ management API URL (must start with `http://` or `https://`) |
| username | string | No | Administrator username for the RabbitMQ cluster. If provided, takes precedence over the username from `rootCredentials`. Required when using `rootCredentials.randomSecret` |
| verifyConnection | bool | No | Verify the connection during configuration. Defaults to `false` |
| passwordPolicy | string | No | Name of a Vault password policy for generating passwords. Defaults to alphanumeric if not set |
| usernameTemplate | string | No | Go template for generating dynamic usernames |
| leaseTTL | int | No | Lease TTL for generated credentials, in seconds |
| leaseMaxTTL | int | No | Maximum lease TTL for generated credentials, in seconds |
| rootCredentials | object | Yes | Credential source for the RabbitMQ admin connection. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified. See [Credential Resolution](#credential-resolution) |

> **Note:** Deleting the `RabbitMQSecretEngineConfig` CR does **not** remove the RabbitMQ connection config from Vault. The configuration can only be removed by deleting the entire engine mount.

## RabbitMQSecretEngineRole

The `RabbitMQSecretEngineRole` CRD allows you to create a [RabbitMQ secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/rabbitmq#create-role) for generating dynamic credentials with specific vhost permissions.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RabbitMQSecretEngineRole
metadata:
  name: my-rabbitmq-role
spec:
  authentication:
    path: kubernetes
    role: rabbitmq-engine-admin
  path: vault-config-operator/rabbitmq
  tags: "administrator"
  vhosts:
    - vhostName: "/"
      permissions:
        read: ".*"
        write: ".*"
        configure: ".*"
    - vhostName: "my-vhost"
      permissions:
        read: "my-queue"
        write: "my-exchange"
        configure: ""
  vhostTopics:
    - vhostName: "/"
      topics:
        - topicName: "my-topic"
          permissions:
            read: ".*"
            write: ".*"
        - topicName: "audit-topic"
          permissions:
            read: ".*"
            write: ""
```

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/roles/<role-name> \
    tags="administrator" \
    vhosts='{"/":{"configure":".*","write":".*","read":".*"},"my-vhost":{"configure":"","write":"my-exchange","read":"my-queue"}}' \
    vhost_topics='{"/":{"my-topic":{"write":".*","read":".*"},"audit-topic":{"write":"","read":".*"}}}'
```

> **Note:** The Vault API expects `vhosts` and `vhost_topics` as JSON-encoded strings, not nested objects. The operator handles this serialization automatically.

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the RabbitMQ secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| tags | string | No | Comma-separated RabbitMQ management tags (e.g., `administrator`, `monitoring`). Determines management UI access level |
| vhosts | []object | No | Vhost permissions for the generated user. Each entry specifies a `vhostName` and `permissions` object |
| vhosts[].vhostName | string | Yes | Name of an existing RabbitMQ vhost |
| vhosts[].permissions | object | Yes | Permissions for the vhost |
| vhosts[].permissions.configure | string | No | Regex pattern for configure permission |
| vhosts[].permissions.write | string | No | Regex pattern for write permission |
| vhosts[].permissions.read | string | No | Regex pattern for read permission |
| vhostTopics | []object | No | Topic-level permissions (requires RabbitMQ 3.7.0+). Each entry specifies a `vhostName` and `topics` list |
| vhostTopics[].vhostName | string | Yes | Name of an existing RabbitMQ vhost |
| vhostTopics[].topics | []object | Yes | List of topic permissions within the vhost |
| vhostTopics[].topics[].topicName | string | Yes | Name of an existing topic/exchange |
| vhostTopics[].topics[].permissions | object | Yes | Permissions for the topic |
| vhostTopics[].topics[].permissions.configure | string | No | Regex pattern for configure permission |
| vhostTopics[].topics[].permissions.write | string | No | Regex pattern for write permission |
| vhostTopics[].topics[].permissions.read | string | No | Regex pattern for read permission |

## Credential Resolution

The root credentials (password and optionally the username) for the RabbitMQ connection can be retrieved in three different ways via the `rootCredentials` field. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified — the webhook rejects manifests that set none or more than one.

### Using a Kubernetes Secret

Specify the `secret` field. The secret uses `usernameKey` and `passwordKey` to identify which keys hold the credentials. If the secret is updated, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    secret:
      name: rabbitmq-admin-password
    usernameKey: username
    passwordKey: password
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve credentials from another Vault path.

```yaml
spec:
  rootCredentials:
    vaultSecret:
      path: secret/data/rabbitmq-credentials
    usernameKey: username
    passwordKey: password
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated. A `username` **must** be specified in `spec.username` when using RandomSecret, because RandomSecret only generates a single value (the password).

```yaml
spec:
  username: my-rabbitmq-admin
  rootCredentials:
    randomSecret:
      name: rabbitmq-random-password
```

> **Note:** If `spec.username` is provided in the CRD, it takes precedence over the username from the credential source. If not provided, the username is retrieved from the credential source along with the password. The `usernameKey` (default: `"username"`) and `passwordKey` (default: `"password"`) fields control which keys are read from the referenced secret.

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault RabbitMQ Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/rabbitmq) — Vault documentation
- [Vault RabbitMQ Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/rabbitmq) — Vault API reference
