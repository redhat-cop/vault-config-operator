# Database Secret Engine

[Database engine documentation](https://developer.hashicorp.com/vault/docs/secrets/databases)

## Overview

The Database secret engine generates dynamic database credentials (username/password pairs) on demand, with automatic revocation after the configured TTL expires. It supports a wide range of database plugins (PostgreSQL, MySQL, MongoDB, etc.) and enables applications to access databases without managing static credentials. For pre-existing database users, it can also manage password rotation via static roles.

The vault-config-operator supports the following CRDs for the Database engine:

- [DatabaseSecretEngineConfig](#databasesecretengineconfig)
- [DatabaseSecretEngineRole](#databasesecretenginerole)
- [DatabaseSecretEngineStaticRole](#databasesecretenginestaticrole)

## DatabaseSecretEngineConfig

The `DatabaseSecretEngineConfig` CRD allows you to configure a [Database secret engine connection](https://developer.hashicorp.com/vault/api-docs/secret/databases#configure-connection).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: DatabaseSecretEngineConfig
metadata:
  name: my-postgresql-database
spec:
  authentication:
    path: kubernetes
    role: database-engine-admin
  path: postgresql-vault-demo/database
  pluginName: postgresql-database-plugin
  allowedRoles:
    - read-only
    - read-write
  connectionURL: "postgresql://{{username}}:{{password}}@my-postgresql.db.svc:5432"
  username: vault-root
  passwordAuthentication: scram-sha-256
  rootCredentials:
    secret:
      name: postgresql-admin-password
  rootPasswordRotation:
    enable: true
    rotationPeriod: 720h
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the secret engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]postgresql-vault-demo/database/config/my-postgresql-database \
    plugin_name=postgresql-database-plugin \
    allowed_roles="read-only,read-write" \
    connection_url="postgresql://{{username}}:{{password}}@my-postgresql.db.svc:5432" \
    username=vault-root \
    password=<retrieved from credentials> \
    password_authentication=scram-sha-256
```

> **Note:** `rootPasswordRotation` is an operator-side feature. When enabled, the operator calls `vault write [namespace/]{path}/rotate-root/{name}` after creating the connection. There is no way to recover the root password after rotation.

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/config/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| pluginName | string | Yes | Database plugin to use (e.g., `postgresql-database-plugin`, `mysql-database-plugin`, `mongodb-database-plugin`) |
| pluginVersion | string | No | Specific version of the database plugin to use |
| verifyConnection | bool | No | If `true`, Vault verifies the connection is usable during configuration. Defaults to `true` |
| allowedRoles | []string | No | List of roles allowed to use this connection. Use `["*"]` to allow all roles. Defaults to `["*"]` |
| rootRotationStatements | []string | No | SQL statements to execute when rotating the root password. Uses the plugin default if not set |
| passwordAuthentication | string | No | Password authentication method: `password` or `scram-sha-256`. Defaults to `password` |
| passwordPolicy | string | No | Name of a Vault password policy to use when generating passwords for this connection |
| connectionURL | string | Yes | Connection URL template. Use `{{username}}` and `{{password}}` as placeholders |
| username | string | No | Username for the database connection. If not set, retrieved from the credential source |
| disableEscaping | bool | No | Disable special character escaping in `username` and `password` fields of `connectionURL`. Defaults to `false` |
| databaseSpecificConfig | map[string]string | No | Plugin-specific key-value pairs added directly to the Vault write payload (e.g., `tls_ca`, `tls_certificate_key` for MongoDB) |
| rootCredentials | object | Yes | Credential source for the database connection. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified. See [Credential Resolution](#credential-resolution) |
| rootPasswordRotation | object | No | Operator-side feature to trigger Vault's `rotate-root` endpoint |
| rootPasswordRotation.enable | bool | No | When `true`, the root password is rotated immediately on first reconcile. Defaults to `false` |
| rootPasswordRotation.rotationPeriod | duration | No | If set, schedules periodic root password rotation at this interval |

## DatabaseSecretEngineRole

The `DatabaseSecretEngineRole` CRD allows you to create a [Database secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/databases#create-role) for generating dynamic credentials.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: DatabaseSecretEngineRole
metadata:
  name: read-only
spec:
  authentication:
    path: kubernetes
    role: database-engine-admin
  path: postgresql-vault-demo/database
  dBName: my-postgresql-database
  defaultTTL: 1h
  maxTTL: 24h
  creationStatements:
    - CREATE ROLE "{{name}}" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';
    - GRANT SELECT ON ALL TABLES IN SCHEMA public TO "{{name}}";
  revocationStatements:
    - REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM "{{name}}";
    - DROP ROLE IF EXISTS "{{name}}";
```

### Vault CLI Equivalent

```shell
vault write [namespace/]postgresql-vault-demo/database/roles/read-only \
    db_name=my-postgresql-database \
    default_ttl=1h \
    max_ttl=24h \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
    revocation_statements="REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM \"{{name}}\"; DROP ROLE IF EXISTS \"{{name}}\";"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| dBName | string | Yes | Name of the database connection (the `DatabaseSecretEngineConfig` resource name) this role uses |
| defaultTTL | duration | No | Default TTL for generated credentials. Uses system/engine default if not set |
| maxTTL | duration | No | Maximum TTL for generated credentials. Uses system/mount default if not set |
| creationStatements | []string | No | SQL statements executed to create a new credential. Supports Vault template variables: `{{name}}`, `{{password}}`, `{{expiration}}` |
| revocationStatements | []string | No | SQL statements executed to revoke a credential |
| rollbackStatements | []string | No | SQL statements executed to roll back a failed credential creation |
| renewStatements | []string | No | SQL statements executed when a credential lease is renewed |

## DatabaseSecretEngineStaticRole

The `DatabaseSecretEngineStaticRole` CRD allows you to create a [Database secret engine static role](https://developer.hashicorp.com/vault/api-docs/secret/databases#create-static-role) for managing password rotation of pre-existing database users.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: DatabaseSecretEngineStaticRole
metadata:
  name: read-only-static
spec:
  authentication:
    path: kubernetes
    role: database-engine-admin
  path: postgresql-vault-demo/database
  dBName: my-postgresql-database
  username: my-static-user
  rotationPeriod: 86400
  rotationStatements:
    - ALTER ROLE "{{name}}" WITH PASSWORD '{{password}}';
  credentialType: password
  passwordCredentialConfig:
    passwordPolicy: my-custom-policy
```

### Vault CLI Equivalent

```shell
vault write [namespace/]postgresql-vault-demo/database/static-roles/read-only-static \
    db_name=my-postgresql-database \
    username=my-static-user \
    rotation_period=86400 \
    rotation_statements="ALTER ROLE \"{{name}}\" WITH PASSWORD '{{password}}';" \
    credential_type=password \
    credential_config='{"password_policy": "my-custom-policy"}'
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/static-roles/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| dBName | string | Yes | Name of the database connection (the `DatabaseSecretEngineConfig` resource name) this static role uses |
| username | string | Yes | The existing database username that Vault will manage and rotate credentials for |
| rotationPeriod | int | Yes | Number of seconds between each automatic password rotation. Minimum: `5` |
| rotationStatements | []string | No | SQL statements executed to rotate the user's password. Uses the plugin default if not set |
| credentialType | string | Yes | Type of credential to generate: `password` or `rsa_private_key` |
| passwordCredentialConfig | object | Conditional | Required when `credentialType` is `password`. Mutually exclusive with `rsaPrivateKeyCredentialConfig` — the webhook rejects CRs with zero or both |
| passwordCredentialConfig.passwordPolicy | string | No | Name of a Vault password policy to use for generating rotated passwords |
| rsaPrivateKeyCredentialConfig | object | Conditional | Required when `credentialType` is `rsa_private_key`. Mutually exclusive with `passwordCredentialConfig` — the webhook rejects CRs with zero or both |
| rsaPrivateKeyCredentialConfig.keyBits | int | No | RSA key size in bits: `2048`, `3072`, or `4096` |
| rsaPrivateKeyCredentialConfig.format | string | No | Key output format: `pkcs8` |

## Credential Resolution

The root credentials (password and optionally the username) for the database connection can be retrieved in three different ways via the `rootCredentials` field. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified — the webhook rejects manifests that set none or more than one.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    secret:
      name: postgresql-admin-password
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve credentials from another Vault path.

```yaml
spec:
  rootCredentials:
    vaultSecret:
      path: secret/data/db-credentials
    usernameKey: username
    passwordKey: password
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated. A `username` must be specified in `spec.username` when using RandomSecret.

```yaml
spec:
  username: my-database-user
  rootCredentials:
    randomSecret:
      name: db-random-password
```

> **Note:** If `spec.username` is provided in the CRD, it takes precedence over the username from the credential source. If not provided, the username is retrieved from the credential source along with the password. The `usernameKey` (default: `"username"`) and `passwordKey` (default: `"password"`) fields control which keys are read from the referenced secret.

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Database Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/databases) — Vault documentation
- [Vault Database Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/databases) — Vault API reference
