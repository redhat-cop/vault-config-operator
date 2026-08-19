# Kerberos Auth Engine

[Kerberos auth method documentation](https://developer.hashicorp.com/vault/docs/auth/kerberos)

## Overview

Kerberos provides SPNEGO-based authentication to Vault using Kerberos tickets. It enables enterprise environments with Active Directory or MIT Kerberos infrastructure to authenticate users to Vault using existing Kerberos credentials, with LDAP-based group resolution for policy assignment.

The vault-config-operator supports the following CRDs for the Kerberos engine:

- [KerberosAuthEngineConfig](#kerberosauthengineconfig)
- [KerberosAuthEngineLDAPConfig](#kerberosauthengineldapconfig)
- [KerberosAuthEngineGroup](#kerberosauthenginegroup)

## KerberosAuthEngineConfig

The `KerberosAuthEngineConfig` CRD allows you to configure the [Kerberos auth engine](https://developer.hashicorp.com/vault/api-docs/auth/kerberos#configure-kerberos) base settings including the service keytab.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KerberosAuthEngineConfig
metadata:
  name: kerberos-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: kerberos
  serviceAccount: vault_svc
  removeInstanceName: false
  addGroupAliases: true
  keytabSecret:
    name: kerberos-keytab
  keytabKey: keytab
```

The keytab content is sourced from a Kubernetes Secret, not provided inline in the CR spec. The Secret must contain the base64-encoded keytab file.

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/kerberos/config \
    keytab=<base64-encoded-keytab> \
    service_account=vault_svc \
    remove_instance_name=false \
    add_group_aliases=true
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the Kerberos auth engine is mounted. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| serviceAccount | string | Yes | Service account associated with the keytab entry and LDAP service account |
| removeInstanceName | bool | No | Strip instance names from Kerberos service principal names in the keytab |
| addGroupAliases | bool | No | Add LDAP groups found for the user as group aliases |
| keytabSecret | object | Yes | Reference to a Kubernetes Secret containing the base64-encoded keytab |
| keytabKey | string | No | Key within the Secret containing the keytab (default: `keytab`) |

## KerberosAuthEngineLDAPConfig

The `KerberosAuthEngineLDAPConfig` CRD allows you to configure the [LDAP backend](https://developer.hashicorp.com/vault/api-docs/auth/kerberos#configure-ldap) for the Kerberos auth engine. This is required for group resolution and user attribute lookup.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KerberosAuthEngineLDAPConfig
metadata:
  name: kerberos-ldap-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: kerberos
  url: "ldaps://ldap.example.com:636"
  tlsMinVersion: "tls12"
  tlsMaxVersion: "tls12"
  denyNullBind: true
  userDN: "ou=Users,dc=example,dc=com"
  userAttr: "samaccountname"
  groupDN: "ou=Groups,dc=example,dc=com"
  groupAttr: "cn"
  groupFilter: "(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))"
  ldapCredentials:
    secret:
      name: ldap-bind-credentials
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/kerberos/config/ldap \
    url="ldaps://ldap.example.com:636" \
    binddn="cn=vault,ou=Users,dc=example,dc=com" \
    bindpass="<password>" \
    userdn="ou=Users,dc=example,dc=com" \
    userattr="samaccountname" \
    groupdn="ou=Groups,dc=example,dc=com" \
    groupattr="cn" \
    groupfilter="(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the Kerberos auth engine is mounted. Full path: `[namespace/]auth/{path}/config/ldap` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| url | string | Yes | LDAP server URL (e.g., `ldaps://ldap.example.com:636`). Multiple URLs can be comma-separated |
| caseSensitiveNames | bool | No | Case-sensitive user/group names for policy matching |
| startTLS | bool | No | Issue StartTLS after establishing unencrypted connection |
| tlsMinVersion | string | No | Minimum TLS version (default: `tls12`) |
| tlsMaxVersion | string | No | Maximum TLS version (default: `tls12`) |
| insecureTLS | bool | No | Skip LDAP server SSL certificate verification |
| bindDN | string | No | Distinguished name for bind (overrides credential source username) |
| userDN | string | No | Base DN for user search |
| userAttr | string | No | Attribute on user objects matching the authenticating username |
| discoverDN | bool | No | Use anonymous bind to discover bind DN |
| denyNullBind | bool | No | Prevent bypassing auth with empty password (default: `true`) |
| upnDomain | string | No | userPrincipalDomain for UPN string construction |
| groupFilter | string | No | Go template for group membership query |
| groupDN | string | No | LDAP search base for group membership |
| groupAttr | string | No | LDAP attribute for group enumeration |
| tokenTTL | string | No | Incremental lifetime for generated tokens |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens |
| tokenNumUses | int | No | Max token uses (0 = unlimited) |
| tokenType | string | No | Token type: service, batch, default, default-service, default-batch |
| ldapCredentials | object | Yes | LDAP bind credentials. See [Credential Resolution](#credential-resolution) |
| tLSConfig | object | No | TLS certificate configuration for LDAP connection |

## KerberosAuthEngineGroup

The `KerberosAuthEngineGroup` CRD allows you to create a [Kerberos group-to-policy mapping](https://developer.hashicorp.com/vault/api-docs/auth/kerberos#create-update-kerberos-group).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KerberosAuthEngineGroup
metadata:
  name: kerberos-engineering-group
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: kerberos
  name: engineering
  policies: "admin,default"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/kerberos/groups/engineering \
    policies="admin,default"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the Kerberos auth engine is mounted. Full path: `[namespace/]auth/{path}/groups/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | Yes | Name of the Kerberos LDAP group |
| policies | string | No | Comma-separated list of policies associated with the group |

## Credential Resolution

The LDAP bind credentials (bindDN and bindPass) for `KerberosAuthEngineLDAPConfig` can be retrieved in three different ways:

### Using a Kubernetes Secret

```yaml
spec:
  ldapCredentials:
    secret:
      name: ldap-bind-credentials
```

### Using a Vault Secret

```yaml
spec:
  ldapCredentials:
    vaultSecret:
      path: secret/data/ldap-credentials
    usernameKey: username
    passwordKey: password
```

### Using a RandomSecret

```yaml
spec:
  ldapCredentials:
    randomSecret:
      name: ldap-random-password
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [LDAP Auth Engine](ldap.md) — Standalone LDAP auth engine (uses same LDAP field set)
