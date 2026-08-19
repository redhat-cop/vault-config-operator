# AliCloud Auth Engine

[AliCloud auth method documentation](https://developer.hashicorp.com/vault/docs/auth/alicloud)

## Overview

AliCloud is a cloud-provider authentication method in Vault that allows AliCloud RAM roles to authenticate using signed `GetCallerIdentity` requests validated against AliCloud STS. It enables workloads running in AliCloud to authenticate to Vault without managing static credentials.

The vault-config-operator supports the following CRD for the AliCloud engine:

- [AliCloudAuthEngineRole](#alicloudauthenginerole)

**Note:** AliCloud auth has no separate configuration endpoint — the mount itself (via `AuthEngineMount`) is the configuration. Only role definitions are managed declaratively.

## AliCloudAuthEngineRole

The `AliCloudAuthEngineRole` CRD allows you to create an [AliCloud auth role](https://developer.hashicorp.com/vault/api-docs/auth/alicloud#create-role).

### Example

First, create an AliCloud auth engine mount using `AuthEngineMount`:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: my-alicloud-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: alicloud
  path: my-alicloud
```

Then create a role referencing the composed mount path (`{spec.path}/{metadata.name}`):

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AliCloudAuthEngineRole
metadata:
  name: my-alicloud-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: my-alicloud/my-alicloud-mount
  arn: "acs:ram::5138828231865461:role/my-alicloud-role"
  tokenPolicies:
    - app-policy
    - default
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
  tokenType: service
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/my-alicloud/my-alicloud-mount/role/my-alicloud-role \
    arn="acs:ram::5138828231865461:role/my-alicloud-role" \
    token_policies="app-policy,default" \
    token_ttl=1h \
    token_max_ttl=24h \
    token_type=service
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the AliCloud auth engine where the role is created. Full path: `[namespace/]auth/{path}/role/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override role name in Vault (defaults to `metadata.name`) |
| arn | string | Yes | The AliCloud RAM role ARN |
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

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault AliCloud API](https://developer.hashicorp.com/vault/api-docs/auth/alicloud) — Complete Vault API reference
