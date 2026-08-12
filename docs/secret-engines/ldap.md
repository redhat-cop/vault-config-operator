# LDAP Secret Engine

[LDAP engine documentation](https://developer.hashicorp.com/vault/docs/secrets/ldap)

## Overview

The LDAP secret engine manages LDAP credentials through two mechanisms: static password rotation for pre-existing LDAP accounts, and dynamic credential generation using LDIF templates for on-demand LDAP user provisioning. It enables applications and administrators to manage LDAP passwords centrally through Vault, eliminating manual password rotation and reducing the risk of credential exposure.

The vault-config-operator supports the following CRDs for the LDAP engine:

- [LDAPSecretEngineConfig](#ldapsecretengineconfig)
- [LDAPSecretEngineStaticRole](#ldapsecretenginestaticrole)
- [LDAPSecretEngineDynamicRole](#ldapsecretenginedynamicrole)

## LDAPSecretEngineConfig

The `LDAPSecretEngineConfig` CRD allows you to configure an [LDAP secret engine connection](https://developer.hashicorp.com/vault/api-docs/secret/ldap#configure-connection).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPSecretEngineConfig
metadata:
  name: ldap-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: ldap
  url: "ldaps://ldap.myorg.com:636"
  schema: openldap
  bindDN: "cn=admin,dc=myorg,dc=com"
  userDN: "ou=users,dc=myorg,dc=com"
  userAttr: cn
  bindCredentials:
    secret:
      name: ldap-bind-credentials
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the secret engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]ldap/config \
    url="ldaps://ldap.myorg.com:636" \
    schema=openldap \
    binddn="cn=admin,dc=myorg,dc=com" \
    bindpass=<retrieved from credentials> \
    userdn="ou=users,dc=myorg,dc=com" \
    userattr=cn
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| url | string | No | LDAP server URL. Default: `ldap://127.0.0.1` |
| schema | string | No | LDAP schema to use when storing entry passwords: `openldap`, `ad`, or `racf`. Default: `openldap` |
| bindDN | string | No | Distinguished name of object to bind for managing user entries. If not set, retrieved from the credential source |
| passwordPolicy | string | No | Name of a Vault password policy to use for generating passwords |
| userDN | string | No | Base DN under which to perform user search |
| userAttr | string | No | Attribute field name used to perform user search |
| upnDomain | string | No | Domain used to construct a UPN string for Active Directory |
| requestTimeout | string | No | Timeout in seconds for the LDAP connection |
| startTLS | bool | No | Issue a StartTLS command after establishing an unencrypted connection |
| insecureTLS | bool | No | Skip LDAP server SSL certificate verification |
| certificate | string | No | CA certificate for verifying the LDAP server certificate (PEM encoded) |
| clientTLSCert | string | No | Client certificate for LDAP mutual TLS (PEM encoded) |
| clientTLSKey | string | No | Client key for LDAP mutual TLS (PEM encoded) |
| skipStaticRoleImportRotation | bool | No | Default value for `skipImportRotation` on static roles created under this config |
| credentialType | string | No | Type of password to generate: `password` or `phrase` (for RACF schema) |
| length | int | No | Generated password string length (deprecated: use `passwordPolicy` instead) |
| connectionTimeout | string | No | Timeout before trying the next URL (deprecated: use `requestTimeout`) |
| bindCredentials | object | Yes | Credential source for the LDAP bind credentials. See [Credential Resolution](#credential-resolution) |

## LDAPSecretEngineStaticRole

The `LDAPSecretEngineStaticRole` CRD allows you to create an [LDAP secret engine static role](https://developer.hashicorp.com/vault/api-docs/secret/ldap#create-static-role) for managing automatic password rotation of pre-existing LDAP accounts.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPSecretEngineStaticRole
metadata:
  name: app-service-account
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: ldap
  username: app-svc
  rotationPeriod: 86400
```

### Vault CLI Equivalent

```shell
vault write [namespace/]ldap/static-role/app-service-account \
    username=app-svc \
    rotation_period=86400
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/static-role/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| username | string | Yes | Existing LDAP username to manage password rotation for. Cannot be modified after creation |
| dn | string | No | Distinguished Name of the existing LDAP entry. Takes precedence over `username`. Cannot be modified after creation |
| rotationPeriod | int | Yes | Time in seconds between automatic password rotations. Minimum: `10` |
| skipImportRotation | bool | No | When `true`, skips the initial password rotation on role creation |

## LDAPSecretEngineDynamicRole

The `LDAPSecretEngineDynamicRole` CRD allows you to create an [LDAP secret engine dynamic role](https://developer.hashicorp.com/vault/api-docs/secret/ldap#create-update-role) for generating on-demand LDAP credentials using LDIF templates.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPSecretEngineDynamicRole
metadata:
  name: dynamic-app-user
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: ldap
  creationLDIF: |
    dn: cn={{.Username}},ou=users,dc=myorg,dc=com
    objectClass: inetOrgPerson
    cn: {{.Username}}
    sn: {{.Username}}
    userPassword: {{.Password}}
  deletionLDIF: |
    dn: cn={{.Username}},ou=users,dc=myorg,dc=com
    changetype: delete
  rollbackLDIF: |
    dn: cn={{.Username}},ou=users,dc=myorg,dc=com
    changetype: delete
  defaultTTL: "1h"
  maxTTL: "24h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]ldap/role/dynamic-app-user \
    creation_ldif=@creation.ldif \
    deletion_ldif=@deletion.ldif \
    rollback_ldif=@rollback.ldif \
    default_ttl=1h \
    max_ttl=24h
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/role/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| creationLDIF | string | Yes | Templatized LDIF string for creating LDAP user accounts. Supports Go template variables: `{{.Username}}`, `{{.Password}}` |
| deletionLDIF | string | Yes | Templatized LDIF string for deleting LDAP user accounts |
| rollbackLDIF | string | No | Templatized LDIF string for rollback on creation failure (recommended) |
| usernameTemplate | string | No | Go template for dynamic username generation |
| defaultTTL | string | No | Default TTL for credential leases (duration format, e.g., `1h`) |
| maxTTL | string | No | Maximum TTL for credential leases (duration format, e.g., `24h`) |

## Credential Resolution

The LDAP bind credentials (bindDN and bindPass) can be retrieved in three different ways via the `bindCredentials` field. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified.

> **Note:** When using `randomSecret`, `spec.bindDN` must be specified in the CRD because `randomSecret` only provides the bind password.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  bindCredentials:
    secret:
      name: ldap-bind-credentials
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve credentials from another Vault path.

```yaml
spec:
  bindCredentials:
    vaultSecret:
      path: secret/ldap-bind
    usernameKey: username
    passwordKey: password
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated. A `spec.bindDN` must be specified when using RandomSecret.

```yaml
spec:
  bindDN: "cn=admin,dc=myorg,dc=com"
  bindCredentials:
    randomSecret:
      name: ldap-random-password
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault LDAP Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/ldap) — Vault documentation
- [Vault LDAP Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/ldap) — Vault API reference
