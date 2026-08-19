# Cloud Foundry Auth Engine

[Cloud Foundry engine documentation](https://developer.hashicorp.com/vault/docs/auth/cf)

## Overview

Cloud Foundry (CF) provides an authentication method for Vault that allows CF application instances to authenticate using their instance identity certificates. It enables secure, identity-based access to Vault secrets for applications running on Cloud Foundry platforms.

The vault-config-operator supports the following CRDs for the Cloud Foundry engine:

- [CFAuthEngineConfig](#cfauthengineconfig)
- [CFAuthEngineRole](#cfauthenginerole)

## CFAuthEngineConfig

The `CFAuthEngineConfig` CRD allows you to configure an authentication engine mount of [type CF](https://developer.hashicorp.com/vault/api-docs/auth/cf).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: CFAuthEngineConfig
metadata:
  name: cf-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: cf
  identityCACertificates:
    - "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
  cfAPIAddr: "https://api.sys.example.cf-app.com"
  cfCredentials:
    secret:
      name: cf-api-credentials
    usernameKey: cf_username
    passwordKey: cf_password
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/cf/config \
    identity_ca_certificates="-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----" \
    cf_api_addr="https://api.sys.example.cf-app.com" \
    cf_username=<retrieved dynamically> \
    cf_password=<retrieved dynamically>
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the CF auth engine is mounted. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| identityCACertificates | []string | Yes | Root CA certificate(s) for verifying CF_INSTANCE_CERT |
| cfAPIAddr | string | Yes | Full API address of the CF deployment |
| cfAPITrustedCertificates | []string | No | Certificate(s) presented by CF API for trust |
| loginMaxSecondsNotBefore | int | No | Max seconds in the past for signature creation (default: 300) |
| loginMaxSecondsNotAfter | int | No | Max seconds in the future for signature creation (default: 60) |
| cfAPIMutualTLSCertificate | string | No | Client certificate for mutual TLS with CF API |
| cfAPIMutualTLSKey | string | No | Client key for mutual TLS with CF API (write-only) |
| cfUsername | string | Conditional | CF API username. Required when using randomSecret (use `spec.cfUsername`, not `spec.username`) |
| cfCredentials | object | No | Credential source for CF API username and password |

## CFAuthEngineRole

The `CFAuthEngineRole` CRD allows you to create a [CF authentication role](https://developer.hashicorp.com/vault/api-docs/auth/cf).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: CFAuthEngineRole
metadata:
  name: cf-role-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: cf
  name: "my-cf-role"
  boundApplicationIDs:
    - "09d7eb6a-afc2-49a0-bb32-858c22f2b346"
  boundSpaceIDs:
    - "21005ebb-8943-433e-84e6-d9d9d7338853"
  boundOrganizationIDs:
    - "9785a884-5e93-49bd-97ee-57bf7c2b20e0"
  tokenPolicies:
    - "default"
    - "cf-app-policy"
  tokenTTL: "1h"
  tokenMaxTTL: "4h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/cf/roles/my-cf-role \
    bound_application_ids="09d7eb6a-afc2-49a0-bb32-858c22f2b346" \
    bound_space_ids="21005ebb-8943-433e-84e6-d9d9d7338853" \
    bound_organization_ids="9785a884-5e93-49bd-97ee-57bf7c2b20e0" \
    token_policies="default,cf-app-policy" \
    token_ttl="1h" \
    token_max_ttl="4h"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the CF auth engine mount where the role is created |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Role name override. Defaults to metadata.name |
| boundApplicationIDs | []string | No | Application IDs to constrain membership |
| boundSpaceIDs | []string | No | Space IDs to constrain membership |
| boundOrganizationIDs | []string | No | Organization IDs to constrain membership |
| boundInstanceIDs | []string | No | Instance IDs to constrain membership (changes on `cf push`) |
| disableIPMatching | bool | No | Disable IP-to-cert matching for proxied logins |
| tokenTTL | string | No | Incremental lifetime for generated tokens |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | Policies to encode onto generated tokens |
| policies | []string | No | Deprecated — use tokenPolicies instead |
| tokenBoundCIDRs | []string | No | CIDR blocks for IP restriction |
| tokenExplicitMaxTTL | string | No | Hard cap max TTL for tokens |
| tokenNoDefaultPolicy | bool | No | Exclude default policy from generated tokens |
| tokenNumUses | int | No | Max token uses (0 = unlimited) |
| tokenPeriod | string | No | Maximum allowed period for periodic tokens |
| tokenType | string | No | Token type: service, batch, default, default-service, default-batch |

## Credential Resolution

The CF API username and password can be retrieved in three different ways:

### Using a Kubernetes Secret

```yaml
spec:
  cfCredentials:
    secret:
      name: cf-api-credentials
    usernameKey: cf_username
    passwordKey: cf_password
```

### Using a Vault Secret

```yaml
spec:
  cfCredentials:
    vaultSecret:
      path: secret/data/cf-credentials
    usernameKey: cf_username
    passwordKey: cf_password
```

### Using a RandomSecret

When using a RandomSecret, the secret only provides the password. You **must** set
`spec.cfUsername` (not `spec.username`) to supply the CF API username explicitly:

```yaml
spec:
  cfUsername: "cf-api-user"
  cfCredentials:
    randomSecret:
      name: cf-random-password
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault CF Auth Method API](https://developer.hashicorp.com/vault/api-docs/auth/cf) — Vault API documentation
