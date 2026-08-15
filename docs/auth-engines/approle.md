# AppRole Auth Engine

[AppRole auth method documentation](https://developer.hashicorp.com/vault/docs/auth/approle)

## Overview

AppRole is a machine-oriented authentication method in Vault. It allows machines and applications to authenticate with Vault using a role ID and secret ID pair, making it the most common auth method for automated workloads and CI/CD pipelines.

The vault-config-operator supports the following CRD for the AppRole engine:

- [AppRoleAuthEngineRole](#approleauthenginerole)

**Note:** AppRole has no separate configuration endpoint — the mount itself (via `AuthEngineMount`) is the configuration. Only role definitions are managed declaratively. Secret-ID lifecycle operations (generate, list, destroy) are operational and not covered by this CRD.

## AppRoleAuthEngineRole

The `AppRoleAuthEngineRole` CRD allows you to create an [AppRole role](https://developer.hashicorp.com/vault/api-docs/auth/approle#create-update-approle-role).

### Example

First, create an AppRole auth engine mount using `AuthEngineMount`:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: my-approle-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: approle
  path: my-approle
```

Then create a role referencing the composed mount path (`{spec.path}/{metadata.name}`):

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AppRoleAuthEngineRole
metadata:
  name: my-app-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: my-approle/my-approle-mount
  bindSecretID: true
  tokenPolicies:
    - app-policy
    - default
  secretIDTTL: "10m"
  secretIDNumUses: 40
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
  tokenType: service
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/my-approle/my-approle-mount/role/my-app-role \
    bind_secret_id=true \
    token_policies="app-policy,default" \
    secret_id_ttl=10m \
    secret_id_num_uses=40 \
    token_ttl=1h \
    token_max_ttl=24h \
    token_type=service
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the AppRole auth engine where the role is created. Full path: `[namespace/]auth/{path}/role/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override role name in Vault (defaults to `metadata.name`) |
| bindSecretID | bool | No | Require secret_id for login (default: `true`) |
| secretIDBoundCIDRs | []string | No | CIDR blocks allowed for login operations |
| secretIDNumUses | int | No | Times a SecretID can be used (0=unlimited) |
| secretIDTTL | string | No | Duration after which a SecretID expires (e.g. `"10m"`, `"1h"`) |
| localSecretIDs | bool | No | Make SecretIDs cluster-local. **Immutable after creation** |
| tokenTTL | string | No | Incremental lifetime for generated tokens (e.g. `"1h"`) |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens (e.g. `"24h"`) |
| tokenPolicies | []string | No | Policies to associate with generated tokens |
| tokenBoundCIDRs | []string | No | CIDR blocks for token authentication |
| tokenExplicitMaxTTL | string | No | Hard cap max TTL for tokens |
| tokenNoDefaultPolicy | bool | No | Exclude default policy from generated tokens |
| tokenNumUses | int | No | Max number of token uses (0=unlimited) |
| tokenPeriod | string | No | Renewal period for periodic tokens |
| tokenType | string | No | Token type: `service`, `batch`, `default`, `default-service`, `default-batch` |

### Immutability Rules

The following fields cannot be changed after creation (enforced by the admission webhook):

- `spec.path` — The auth engine mount path
- `spec.name` — The Vault role name override
- `spec.localSecretIDs` — Cluster-local SecretID setting (Vault enforces this constraint)

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault AppRole API](https://developer.hashicorp.com/vault/api-docs/auth/approle) — Complete Vault API reference
