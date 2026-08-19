# GitHub Auth Engine

[GitHub auth method documentation](https://developer.hashicorp.com/vault/docs/auth/github)

## Overview

The GitHub auth method allows users to authenticate to Vault using a GitHub personal access token. Users must be part of a specified GitHub organization and can receive Vault policies based on their team memberships or individual user mappings.

The vault-config-operator supports the following CRDs for the GitHub auth engine:

- [GitHubAuthEngineConfig](#githubauthengineconfig)
- [GitHubAuthEngineTeamMap](#githubauthengineteammap)
- [GitHubAuthEngineUserMap](#githubauthenginerusermap)

## GitHubAuthEngineConfig

The `GitHubAuthEngineConfig` CRD allows you to configure an authentication engine mount of [type GitHub](https://developer.hashicorp.com/vault/api-docs/auth/github#configure-method).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GitHubAuthEngineConfig
metadata:
  name: github-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: github
  organization: "acme-org"
  organizationID: 12345
  baseURL: "https://github.example.com/api/v3"
  tokenTTL: "1h"
  tokenMaxTTL: "4h"
  tokenPolicies:
    - default
    - dev-policy
  tokenBoundCIDRs:
    - "10.0.0.0/8"
  tokenNoDefaultPolicy: true
  tokenNumUses: 5
  tokenType: "service"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/github/config \
    organization=acme-org \
    organization_id=12345 \
    base_url="https://github.example.com/api/v3" \
    token_ttl=1h \
    token_max_ttl=4h \
    token_policies="default,dev-policy" \
    token_bound_cidrs="10.0.0.0/8" \
    token_no_default_policy=true \
    token_num_uses=5 \
    token_type=service
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the GitHub auth engine is mounted. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| organization | string | Yes | GitHub organization users must be part of |
| organizationID | int64 | No | Numeric ID of the organization. Vault auto-fetches if not provided |
| baseURL | string | No | API endpoint for GitHub Enterprise. Leave empty for public GitHub |
| tokenTTL | string | No | Incremental lifetime for generated tokens (e.g., "1h", "30m") |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | Policies to encode onto generated tokens |
| tokenBoundCIDRs | []string | No | CIDR blocks that restrict authentication |
| tokenExplicitMaxTTL | string | No | Hard cap max TTL for tokens |
| tokenNoDefaultPolicy | bool | No | If true, omits the default policy from generated tokens |
| tokenNumUses | int64 | No | Max number of times a token may be used (0 = unlimited) |
| tokenPeriod | string | No | Maximum allowed period for periodic tokens |
| tokenType | string | No | Token type: "service", "batch", "default", "default-service", "default-batch" |

### Deletion Behavior

The `GitHubAuthEngineConfig` is **not deletable** (`IsDeletable=false`). Deleting the Kubernetes CR removes the K8s object but leaves the Vault configuration intact. The config is tied to the mount lifecycle — to fully remove the GitHub auth configuration, delete the associated `AuthEngineMount`.

## GitHubAuthEngineTeamMap

The `GitHubAuthEngineTeamMap` CRD allows you to create a [GitHub team-to-policy mapping](https://developer.hashicorp.com/vault/api-docs/auth/github#map-github-teams).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GitHubAuthEngineTeamMap
metadata:
  name: dev-team-mapping
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: github
  name: dev-team
  policies: "dev-policy,readonly"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/github/map/teams/dev-team \
    value="dev-policy,readonly"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the GitHub auth engine mount |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | Yes | GitHub team name in slugified format (lowercase, hyphens only) |
| policies | string | No | Comma-separated list of policies to assign to this team |

## GitHubAuthEngineUserMap

The `GitHubAuthEngineUserMap` CRD allows you to create a [GitHub user-to-policy mapping](https://developer.hashicorp.com/vault/api-docs/auth/github#map-github-users).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GitHubAuthEngineUserMap
metadata:
  name: sethvargo-mapping
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: github
  name: sethvargo
  policies: "sethvargo-policy,admin"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/github/map/users/sethvargo \
    value="sethvargo-policy,admin"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the GitHub auth engine mount |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | Yes | GitHub username |
| policies | string | No | Comma-separated list of policies to assign to this user |

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault GitHub Auth Method API](https://developer.hashicorp.com/vault/api-docs/auth/github)
