# TLS Certificate Auth Engine

[TLS Certificate engine documentation](https://developer.hashicorp.com/vault/docs/auth/cert)

## Overview

The TLS Certificate auth method allows authentication using SSL/TLS client certificates that are either signed by a CA or self-signed. Clients provide a TLS certificate during the login handshake, and Vault verifies it against configured CA certificates and optional constraints such as allowed Common Names, DNS SANs, or Organizational Units.

The vault-config-operator supports the following CRDs for the TLS Certificate engine:

- [CertAuthEngineConfig](#certauthengineconfig)
- [CertAuthEngineRole](#certauthenginerole)

## CertAuthEngineConfig

The `CertAuthEngineConfig` CRD allows you to configure a [TLS Certificate auth engine](https://developer.hashicorp.com/vault/api-docs/auth/cert#configure-tls-certificate-method).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: CertAuthEngineConfig
metadata:
  name: cert-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: cert
  ocspCacheSize: 100
  roleCacheSize: 200
  disableBinding: false
  enableIdentityAliasMetadata: false
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/<name>/config \
    disable_binding=false \
    enable_identity_alias_metadata=false \
    ocsp_cache_size=100 \
    role_cache_size=200
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path for the cert auth engine. Full path: `[namespace/]auth/{path}/{name}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| disableBinding | bool | No | Skip client identity matching during renewal. Defaults to `false` |
| enableIdentityAliasMetadata | bool | No | Store certificate metadata in the identity alias. Defaults to `false` |
| ocspCacheSize | int | No | Size of the OCSP response LRU cache. Minimum: 0. Defaults to `100` |
| roleCacheSize | int | No | Size of the role cache. Set to `-1` to disable caching. Minimum: -1. Defaults to `200` |

## CertAuthEngineRole

The `CertAuthEngineRole` CRD allows you to create a [TLS Certificate auth role](https://developer.hashicorp.com/vault/api-docs/auth/cert#create-ca-certificate-role).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: CertAuthEngineRole
metadata:
  name: my-app-cert-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: cert
  certificate: |
    -----BEGIN CERTIFICATE-----
    MIIDQjCCAiqgAwIBAgI...
    -----END CERTIFICATE-----
  allowedCommonNames:
    - "*.example.com"
  allowedDNSSANs:
    - "*.internal.example.com"
  tokenPolicies:
    - app-policy
    - default
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/certs/<role-name> \
    certificate=@ca.pem \
    allowed_common_names="*.example.com" \
    allowed_dns_sans="*.internal.example.com" \
    token_policies="app-policy,default" \
    token_ttl="1h" \
    token_max_ttl="24h"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path for the cert auth engine. Full path: `[namespace/]auth/{path}/certs/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| certificate | string | Yes | PEM-format CA certificate used to verify client certificates |
| allowedCommonNames | []string | No | Glob patterns for allowed Common Names. Empty allows all |
| allowedDNSSANs | []string | No | Glob patterns for allowed DNS Subject Alternative Names. Empty allows all |
| allowedEmailSANs | []string | No | Glob patterns for allowed Email Subject Alternative Names. Empty allows all |
| allowedURISANs | []string | No | Glob patterns for allowed URI Subject Alternative Names. Empty allows all |
| allowedOrganizationalUnits | []string | No | Glob patterns for allowed Organizational Units. Empty allows all |
| requiredExtensions | []string | No | Required Custom Extension OIDs. Format: `oid:value` or `hex:oid:value` |
| allowedMetadataExtensions | []string | No | OID extensions to add as metadata on successful authentication |
| ocspEnabled | bool | No | Validate certificate revocation via OCSP. Defaults to `false` |
| ocspCACertificates | string | No | Additional OCSP responder certificates in base64-encoded PEM format |
| ocspServersOverride | []string | No | Override OCSP server addresses |
| ocspFailOpen | bool | No | Allow login if the OCSP response is unavailable or unknown. Defaults to `false` |
| ocspThisUpdateMaxAge | string | No | Maximum age of the OCSP `thisUpdate` field (duration string) |
| ocspMaxRetries | int64 | No | OCSP request retry count. Set to `0` to disable retries. Minimum: 0. Defaults to `4` |
| ocspQueryAllServers | bool | No | Query all OCSP servers and require unanimous agreement. Defaults to `false` |
| displayName | string | No | Display name on tokens issued via this role. Defaults to the role name |
| tokenTTL | string | No | Incremental token lifetime (e.g., `"1h"`) |
| tokenMaxTTL | string | No | Maximum token lifetime |
| tokenPolicies | []string | No | Policies to attach to generated tokens |
| tokenBoundCIDRs | []string | No | CIDR blocks restricting token usage |
| tokenExplicitMaxTTL | string | No | Hard cap TTL overriding `tokenTTL` and `tokenMaxTTL` |
| tokenNoDefaultPolicy | bool | No | Exclude the `default` policy from generated tokens. Defaults to `false` |
| tokenNumUses | int64 | No | Maximum token usage count. `0` means unlimited |
| tokenPeriod | string | No | Maximum allowed period for periodic token requests |
| tokenType | string | No | Token type: `service`, `batch`, `default`, `default-service`, or `default-batch` |

## Credential Resolution

Unlike auth engines that resolve credentials (passwords, client secrets) from external sources before writing to the Vault API, the TLS Certificate auth engine's sensitive field is the `certificate` PEM itself — a string field set directly in the CR spec.

The `certificate` field is specified inline in the `CertAuthEngineRole` spec as a PEM-encoded string:

```yaml
spec:
  certificate: |
    -----BEGIN CERTIFICATE-----
    MIIDQjCCAiqgAwIBAgI...
    -----END CERTIFICATE-----
```

For production usage, the CA certificate PEM is often maintained in a Kubernetes Secret and then copied into the CR manifest. There is no automatic `spec.credentialSecret` or `spec.vaultSecretRef` reference mechanism for this engine type — CertAuth is one of the simpler credential patterns in the operator.

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault TLS Certificate Auth Method](https://developer.hashicorp.com/vault/docs/auth/cert) — Vault documentation
- [Vault TLS Certificate Auth API](https://developer.hashicorp.com/vault/api-docs/auth/cert) — Vault API reference
