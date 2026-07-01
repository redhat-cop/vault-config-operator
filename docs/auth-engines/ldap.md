# LDAP Auth Engine

[LDAP engine documentation](https://developer.hashicorp.com/vault/docs/auth/ldap)

## Overview

The LDAP auth method allows users to authenticate with Vault using credentials stored in an existing LDAP directory. Vault verifies login credentials against the LDAP server and can map LDAP groups to Vault policies, enabling centralized identity management for Vault access.

The vault-config-operator supports the following CRDs for the LDAP engine:

- [LDAPAuthEngineConfig](#ldapauthengineconfig)
- [LDAPAuthEngineGroup](#ldapauthenginegroup)

## LDAPAuthEngineConfig

The `LDAPAuthEngineConfig` CRD allows you to configure an [LDAP auth engine](https://developer.hashicorp.com/vault/api-docs/auth/ldap#configure-ldap).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPAuthEngineConfig
metadata:
  name: ldap-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: ldap
  url: ldaps://ldap.example.com:636
  bindDN: cn=vault-svc,ou=Services,dc=example,dc=com
  bindCredentials:
    secret:
      name: ldap-bind-credentials
  userDN: ou=Users,dc=example,dc=com
  userAttr: sAMAccountName
  groupDN: ou=Groups,dc=example,dc=com
  groupAttr: cn
  groupFilter: (|(memberUid={{.Username}})(member={{.UserDN}})(uniqueMember={{.UserDN}}))
  insecureTLS: false
  TLSMinVersion: tls12
  TLSMaxVersion: tls13
  tLSConfig:
    tlsSecret:
      name: ldap-tls-certificate
  tokenPolicies: default
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the LDAP auth engine being configured. They may point to different mounts.

### Vault CLI Equivalent

> **Note:** The example YAML above uses `tLSConfig` to supply TLS material via a Kubernetes Secret. The operator resolves the Secret's `ca.crt`, `tls.crt`, and `tls.key` entries into the Vault API fields shown below (`certificate`, `client_tls_cert`, `client_tls_key`). Similarly, `bindCredentials` is resolved into `binddn` and `bindpass`.

```shell
vault write [namespace/]auth/<path>/config \
    url="ldaps://ldap.example.com:636" \
    binddn="cn=vault-svc,ou=Services,dc=example,dc=com" \
    bindpass="<retrieved from bindCredentials>" \
    userdn="ou=Users,dc=example,dc=com" \
    userattr="sAMAccountName" \
    groupdn="ou=Groups,dc=example,dc=com" \
    groupattr="cn" \
    groupfilter="(|(memberUid={{.Username}})(member={{.UserDN}})(uniqueMember={{.UserDN}}))" \
    insecure_tls=false \
    tls_min_version="tls12" \
    tls_max_version="tls13" \
    certificate=@ca.pem \
    client_tls_cert=@client.pem \
    client_tls_key=@client-key.pem \
    token_policies="default" \
    token_ttl="1h" \
    token_max_ttl="24h"
```

### Field Descriptions

**Common Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path for the LDAP auth engine. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| bindCredentials | object | Yes | Credential resolution configuration for the LDAP bind account. See [Credential Resolution](#credential-resolution) |
| tLSConfig | object | No | TLS certificate via Kubernetes Secret. See [TLS Configuration](#tls-configuration) |

**Connection & TLS Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| url | string | Yes | `ldap://127.0.0.1` | LDAP server URL. Examples: `ldap://ldap.myorg.com`, `ldaps://ldap.myorg.com:636`. Multiple URLs can be comma-separated |
| startTLS | bool | No | `false` | Issue a StartTLS command after establishing an unencrypted connection |
| TLSMinVersion | string | No | `tls12` | Minimum TLS version. Values: `tls10`, `tls11`, `tls12`, `tls13` |
| TLSMaxVersion | string | No | `tls12` | Maximum TLS version. Values: `tls10`, `tls11`, `tls12`, `tls13` |
| insecureTLS | bool | No | `false` | Skip LDAP server SSL certificate verification |
| certificate | string | No | — | CA certificate (PEM) to verify the LDAP server certificate |
| clientTLSCert | string | No | — | Client certificate (PEM) to present to the LDAP server |
| clientTLSKey | string | No | — | Client private key (PEM) for mutual TLS |
| requestTimeout | string | No | `90s` | Connection timeout for LDAP requests |

**User Search Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| bindDN | string | No | — | Username for LDAP bind. If in `account@domain.com` format, transformed to proper LDAP bind DN. Takes precedence over username from `bindCredentials` |
| userDN | string | No | — | Base DN for user search. Example: `ou=Users,dc=example,dc=com` |
| userAttr | string | No | `cn` | Attribute on user objects to match against the login username. Examples: `sAMAccountName`, `cn`, `uid` |
| userFilter | string | No | — | Optional LDAP user search filter. Template variables: `UserAttr`, `Username`. Default: `({{.UserAttr}}={{.Username}})` |
| discoverDN | bool | No | `false` | Use anonymous bind to discover the bind DN of a user |
| denyNullBind | bool | No | `true` | Prevent users from bypassing authentication with an empty password |
| UPNDomain | string | No | — | Domain for UPN construction. Login will bind as `username@UPNDomain` |
| caseSensitiveNames | bool | No | `false` | If set, user and group names in policies are case sensitive. Login usernames are always sent as-is to LDAP |
| usernameAsAlias | bool | No | `false` | Force the auth method to use the login username as the identity alias name |

**Group Search Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| groupDN | string | No | — | Base DN for group membership search. Example: `ou=Groups,dc=example,dc=com` |
| groupFilter | string | No | — | Go template for group membership queries. Template variables: `UserDN`, `Username`. Default: `(|(memberUid={{.Username}})(member={{.UserDN}})(uniqueMember={{.UserDN}}))` |
| groupAttr | string | No | `cn` | Attribute on group objects to enumerate membership. Use `cn` for group objects, `memberOf` for user objects |
| anonymousGroupSearch | bool | No | `false` | Use anonymous binds for LDAP group searches. Initial credentials are still used for the connection test |

**Token Parameters:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| tokenTTL | string | No | — | Incremental lifetime for generated tokens (e.g., `"1h"`) |
| tokenMaxTTL | string | No | — | Maximum lifetime for generated tokens |
| tokenPolicies | string | No | — | Comma-separated list of policies to attach to generated tokens |
| tokenBoundCIDRs | string | No | — | Comma-separated CIDR blocks restricting token usage |
| tokenExplicitMaxTTL | string | No | — | Hard cap TTL overriding `tokenTTL` and `tokenMaxTTL` |
| tokenNoDefaultPolicy | bool | No | `false` | Exclude the `default` policy from generated tokens |
| tokenNumUses | int64 | No | `0` | Maximum token usage count. `0` means unlimited |
| tokenPeriod | int64 | No | `0` | The period, if any, to set on the token |
| tokenType | string | No | — | Token type: `service`, `batch`, `default`, `default-service`, or `default-batch` |

### TLS Configuration

The LDAP engine supports two methods for configuring TLS certificates:

**Method 1: Inline fields** — Set PEM-encoded certificates directly in the CR spec:

```yaml
spec:
  certificate: |
    -----BEGIN CERTIFICATE-----
    MIIDQjCCAiqgAwIBAgI...
    -----END CERTIFICATE-----
  clientTLSCert: |
    -----BEGIN CERTIFICATE-----
    MIIDQjCCAiqgAwIBAgI...
    -----END CERTIFICATE-----
  clientTLSKey: |
    -----BEGIN RSA PRIVATE KEY-----
    MIIEowIBAAKCAQEA...
    -----END RSA PRIVATE KEY-----
```

**Method 2: `tLSConfig` field** — Reference a Kubernetes TLS Secret containing the certificates:

```yaml
spec:
  tLSConfig:
    tlsSecret:
      name: ldap-tls-certificate
```

When `tLSConfig.tlsSecret` is set, the operator reads the following keys from the referenced Secret:

| Secret Key | Maps To | Description |
|------------|---------|-------------|
| `ca.crt` | `certificate` | CA certificate for LDAP server verification |
| `tls.crt` | `clientTLSCert` | Client certificate for mutual TLS |
| `tls.key` | `clientTLSKey` | Client private key for mutual TLS |

The `tLSConfig` approach is recommended as the Kubernetes-native way to manage TLS material.

## LDAPAuthEngineGroup

The `LDAPAuthEngineGroup` CRD allows you to create an [LDAP auth group-to-policy mapping](https://developer.hashicorp.com/vault/api-docs/auth/ldap#create-update-ldap-group).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPAuthEngineGroup
metadata:
  name: ldap-admins
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: ldap
  name: vault-admins
  policies: "admin, audit, users"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/groups/<name> \
    policies="admin, audit, users"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | LDAP auth mount path. Full path: `[namespace/]auth/{path}/groups/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | Yes | The name of the LDAP group |
| policies | string | No | Comma-separated list of Vault policies to associate with the group |

## Credential Resolution

The LDAP bind password (and optionally the bind DN) can be retrieved in three different ways via the `bindCredentials` field. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified — the webhook rejects manifests that set none or more than one.

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
      path: secret/data/ldap-credentials
    usernameKey: username
    passwordKey: password
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated. A `bindDN` must be specified in the spec when using RandomSecret.

```yaml
spec:
  bindDN: cn=vault-svc,ou=Services,dc=example,dc=com
  bindCredentials:
    randomSecret:
      name: ldap-random-password
```

> **Note:** If `bindDN` is set in the spec, it takes precedence over the username retrieved from the referenced secret. The password is always retrieved from the credential source.

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault LDAP Auth Method](https://developer.hashicorp.com/vault/docs/auth/ldap) — Vault documentation
- [Vault LDAP Auth API](https://developer.hashicorp.com/vault/api-docs/auth/ldap) — Vault API reference
