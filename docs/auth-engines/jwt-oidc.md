# JWT/OIDC Auth Engine

[JWT/OIDC engine documentation](https://developer.hashicorp.com/vault/docs/auth/jwt)

## Overview

The JWT/OIDC auth method allows authentication using JWTs (including OIDC tokens) that are cryptographically verified using locally-provided keys, or by fetching a set of keys from a remote JWKS endpoint or OIDC Discovery URL. The engine supports two mutually-exclusive modes: **OIDC mode** (interactive browser-based login via an OIDC provider) and **JWT mode** (direct JWT validation without a provider round-trip).

The vault-config-operator supports the following CRDs for the JWT/OIDC engine:

- [JWTOIDCAuthEngineConfig](#jwtoidcauthengineconfig)
- [JWTOIDCAuthEngineRole](#jwtoidcauthenginerole)

## JWTOIDCAuthEngineConfig

The `JWTOIDCAuthEngineConfig` CRD allows you to configure a [JWT/OIDC auth engine](https://developer.hashicorp.com/vault/api-docs/auth/jwt#configure).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: JWTOIDCAuthEngineConfig
metadata:
  name: oidc-azure-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: oidc
  OIDCDiscoveryURL: https://login.microsoftonline.com/{tenant-id}/v2.0
  OIDCClientID: 00000000-0000-0000-0000-000000000000
  OIDCResponseMode: form_post
  OIDCResponseTypes:
    - code
  JWTSupportedAlgs:
    - RS256
  defaultRole: azure-user
  namespaceInState: true
  providerConfig:
    provider: azure
  OIDCCredentials:
    secret:
      name: oidc-client-credentials
    usernameKey: client_id
    passwordKey: client_secret
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/config \
    oidc_discovery_url="https://login.microsoftonline.com/{tenant-id}/v2.0" \
    oidc_client_id="00000000-0000-0000-0000-000000000000" \
    oidc_client_secret="<retrieved from OIDCCredentials>" \
    oidc_response_mode="form_post" \
    oidc_response_types="code" \
    jwt_supported_algs="RS256" \
    default_role="azure-user" \
    namespace_in_state=true \
    provider_config='{"provider":"azure"}'
```

### Field Descriptions

**Common Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path for the JWT/OIDC auth engine. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| OIDCCredentials | object | No | Credential resolution for OIDC mode (client ID + client secret). See [Credential Resolution](#credential-resolution) |

**OIDC Discovery Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| OIDCDiscoveryURL | string | No* | — | The OIDC Discovery URL, without any `.well-known` component (base path). Cannot be used with `JWKSURL` or `JWTValidationPubKeys` |
| OIDCDiscoveryCAPEM | string | No | — | CA certificate or chain (PEM) to validate connections to the OIDC Discovery URL. If not set, system certificates are used |
| OIDCClientID | string | No | — | The OAuth Client ID from the provider for OIDC roles. If set, takes precedence over the client ID retrieved from `OIDCCredentials`. If omitted, the client ID is read from the credential source instead |
| OIDCResponseMode | string | No | — | Response mode for the OAuth2 request. Allowed values: `query`, `form_post`. If using Vault namespaces and set to `form_post`, set `namespaceInState` to `false` |
| OIDCResponseTypes | []string | No | — | Response types to request. Allowed values: `code`, `id_token`. Note: `id_token` may only be used if `OIDCResponseMode` is `form_post` |

\*Required when using OIDC mode (`OIDCDiscoveryURL` must be set)

**JWKS/JWT Validation Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| JWKSURL | string | No* | — | JWKS URL for authenticating signatures. Cannot be used with `OIDCDiscoveryURL` or `JWTValidationPubKeys` |
| JWKSCAPEM | string | No | — | CA certificate or chain (PEM) to validate connections to the JWKS URL. If not set, system certificates are used |
| JWTValidationPubKeys | []string | No* | — | PEM-encoded public keys for local signature authentication. Cannot be used with `JWKSURL` or `OIDCDiscoveryURL` |

\*One of `OIDCDiscoveryURL`, `JWKSURL`, or `JWTValidationPubKeys` must be set

**General Settings:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| boundIssuer | string | No | — | Value to match against the `iss` claim in a JWT |
| JWTSupportedAlgs | []string | No | `[RS256]` for OIDC | Supported signing algorithms. Defaults to all available algorithms for JWT roles |
| defaultRole | string | No | — | Default role to use if none is provided during login |
| providerConfig | JSON | No | — | Provider-specific configuration. Providers with specific handling: Azure, Google. See [OIDC Provider Setup](https://developer.hashicorp.com/vault/docs/auth/jwt/oidc-providers) |
| namespaceInState | bool | No | `true` | Pass namespace in the OIDC state parameter instead of as a separate query parameter. When `true`, the redirect URL(s) need not contain a namespace query parameter |

### JWT vs OIDC Modes

The JWT/OIDC engine supports two mutually-exclusive validation source configurations. Exactly **one** of the three validation source fields must be set:

| Mode | Validation Source | Required Config Fields | Credentials Needed? |
|------|------------------|------------------------|---------------------|
| OIDC | OIDC Discovery URL | `OIDCDiscoveryURL`, optionally `OIDCDiscoveryCAPEM` | Yes — `OIDCCredentials` |
| JWT (JWKS) | Remote JWKS URL | `JWKSURL`, optionally `JWKSCAPEM` | No |
| JWT (Public Keys) | Local public keys | `JWTValidationPubKeys` | No |

**Key differences:**
- **OIDC mode** enables interactive browser-based login. It requires `OIDCCredentials` (client ID + client secret) to communicate with the OIDC provider. Roles using this mode have `roleType: "oidc"` (the default).
- **JWT mode** validates tokens directly using either a remote JWKS endpoint or locally-configured public keys. No provider credentials are needed. Roles using this mode **must** explicitly set `roleType: "jwt"` — the default is `"oidc"`, so omitting `roleType` on a JWT-backed engine will produce a role with incorrect semantics.

Setting more than one validation source field is invalid and will be rejected by Vault.

## JWTOIDCAuthEngineRole

The `JWTOIDCAuthEngineRole` CRD allows you to create a [JWT/OIDC auth role](https://developer.hashicorp.com/vault/api-docs/auth/jwt#create-update-role).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: JWTOIDCAuthEngineRole
metadata:
  name: azure-user-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: oidc
  name: azure-user
  roleType: oidc
  userClaim: email
  boundAudiences:
    - 00000000-0000-0000-0000-000000000000
  OIDCScopes:
    - openid
    - profile
    - email
  allowedRedirectURIs:
    - https://vault.example.com:8250/oidc/callback
    - http://localhost:8250/oidc/callback
  groupsClaim: groups
  claimMappings:
    preferred_username: display_name
  tokenPolicies:
    - app-reader
    - default
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/role/<name> \
    role_type="oidc" \
    user_claim="email" \
    bound_audiences="00000000-0000-0000-0000-000000000000" \
    oidc_scopes="openid,profile,email" \
    allowed_redirect_uris="https://vault.example.com:8250/oidc/callback,http://localhost:8250/oidc/callback" \
    groups_claim="groups" \
    claim_mappings='{"preferred_username":"display_name"}' \
    token_policies="app-reader,default" \
    token_ttl="1h" \
    token_max_ttl="24h"
```

### Field Descriptions

**Role Identity Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | JWT/OIDC auth mount path. Full path: `[namespace/]auth/{path}/role/{name}` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | Yes | — | Name of the role in Vault. If set to a different value than `metadata.name`, the Vault role name will differ from the CR name |
| roleType | string | No | `oidc` | Type of role. Allowed values: `oidc`, `jwt` |
| userClaim | string | Yes | — | Claim to uniquely identify the user; used as the name for the Identity entity alias. Must be a string claim |
| userClaimJSONPointer | bool | No | `false` | Use JSON pointer syntax for referencing claims in `userClaim` |

**Claims & Binding Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| boundAudiences | []string | No* | — | List of `aud` claims to match against. Any match is sufficient |
| boundSubject | string | No | — | If set, requires the `sub` claim to match this value |
| boundClaims | JSON | No | — | Map of claims (keys) to match against respective claim values. Keys support JSON pointer syntax |
| boundClaimsType | string | No | `string` | Interpretation of `boundClaims` values. Allowed values: `string` (exact match), `glob` (wildcard with `*`) |
| groupsClaim | string | No | — | Claim to identify the user's groups; used as names for Identity group aliases. Must be a list of strings. Supports JSON pointer syntax |
| claimMappings | map[string]string | No | — | Map of claims (keys) to be copied to specified metadata fields (values). Keys support JSON pointer syntax |

\*Required for `jwt` roles, optional for `oidc` roles

**OIDC-Specific Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| OIDCScopes | []string | No | — | OIDC scopes to request. The standard `openid` scope is automatically included |
| allowedRedirectURIs | []string | Yes* | — | Allowed values for `redirect_uri` during OIDC logins |
| verboseOIDCLogging | bool | No | `false` | Log received OIDC tokens and claims at debug level. Not recommended in production |
| maxage | int64 | No | `0` | Allowable elapsed time (seconds) since the user's last active authentication with the OIDC provider. If set, the `max_age` request parameter is included in the authentication request |

\*Required for `oidc` roles

**JWT-Specific Fields (Leeway):**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| clockSkewLeeway | int64 | No | `60` | Leeway (seconds) added to all claims to account for clock skew. Set to `-1` to disable |
| expirationLeeway | int64 | No | `150` | Leeway (seconds) added to expiration (`exp`) claims. Set to `-1` to disable |
| notBeforeLeeway | int64 | No | `150` | Leeway (seconds) added to not-before (`nbf`) claims. Set to `-1` to disable |

**Token Parameters:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| tokenTTL | string | No | — | Incremental lifetime for generated tokens. Referenced at renewal time |
| tokenMaxTTL | string | No | — | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | — | Policies to attach to generated tokens |
| tokenBoundCIDRs | []string | No | — | CIDR blocks restricting token usage |
| tokenExplicitMaxTTL | string | No | — | Hard cap TTL overriding `tokenTTL` and `tokenMaxTTL` |
| tokenNoDefaultPolicy | bool | No | `false` | Exclude the `default` policy from generated tokens |
| tokenNumUses | int64 | No | `0` | Maximum token usage count. `0` means unlimited |
| tokenPeriod | int64 | No | `0` | The period, if any, to set on the token |
| tokenType | string | No | — | Token type. Allowed values: `service`, `batch`, `default`, `default-service`, `default-batch` |

## Credential Resolution

The OIDC client ID and client secret (used to communicate with the OIDC provider) can be retrieved in three different ways via the `OIDCCredentials` field. This field is **only needed for OIDC mode** — JWT mode does not require provider credentials.

> **Note:** If `OIDCClientID` is set directly in the spec, it takes precedence over the username (client ID) retrieved from the referenced secret. The password (client secret) is always retrieved from the credential source.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  OIDCCredentials:
    secret:
      name: oidc-client-credentials
    usernameKey: client_id
    passwordKey: client_secret
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve credentials from another Vault path.

```yaml
spec:
  OIDCCredentials:
    vaultSecret:
      path: secret/data/oidc-credentials
    usernameKey: client_id
    passwordKey: client_secret
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated. An `OIDCClientID` must be specified directly in the spec when using RandomSecret — the client ID is not read from the generated secret. The `usernameKey` and `passwordKey` fields are ignored for this method.

```yaml
spec:
  OIDCClientID: 00000000-0000-0000-0000-000000000000
  OIDCCredentials:
    randomSecret:
      name: oidc-random-secret
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault JWT/OIDC Auth Method](https://developer.hashicorp.com/vault/docs/auth/jwt) — Vault documentation
- [Vault JWT/OIDC Auth API](https://developer.hashicorp.com/vault/api-docs/auth/jwt) — Vault API reference
- [OIDC Provider Setup](https://developer.hashicorp.com/vault/docs/auth/jwt/oidc-providers) — Provider-specific configuration guides
