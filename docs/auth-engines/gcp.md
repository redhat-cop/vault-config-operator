# GCP Auth Engine

[GCP engine documentation](https://developer.hashicorp.com/vault/docs/auth/gcp)

## Overview

The GCP auth method allows authentication for Google Cloud entities using either IAM service account credentials or GCE instance identity metadata. IAM roles verify signed JWTs from service accounts, while GCE roles verify instance identity tokens from Compute Engine VMs. This enables workloads running on Google Cloud to authenticate to Vault without managing static credentials.

The vault-config-operator supports the following CRDs for the GCP engine:

- [GCPAuthEngineConfig](#gcpauthengineconfig)
- [GCPAuthEngineRole](#gcpauthenginerole)

## GCPAuthEngineConfig

The `GCPAuthEngineConfig` CRD allows you to configure a [GCP auth engine](https://developer.hashicorp.com/vault/api-docs/auth/gcp#configure).

> **Note:** Deleting a `GCPAuthEngineConfig` CR does **not** remove the GCP config from Vault. The auth mount must be disabled separately.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GCPAuthEngineConfig
metadata:
  name: gcp-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: gcp
  IAMalias: default
  IAMmetadata: default
  GCEalias: role_id
  GCEmetadata: default
  GCPCredentials:
    secret:
      name: gcp-credentials
    usernameKey: serviceaccount
    passwordKey: credentials
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/config \
    credentials="<retrieved from GCPCredentials>" \
    iam_alias="default" \
    iam_metadata="default" \
    gce_alias="role_id" \
    gce_metadata="default" \
    custom_endpoint='{}'
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path for the GCP auth engine. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| GCPCredentials | object | No | — | Credential resolution for GCP service account JSON. See [Credential Resolution](#credential-resolution) |
| serviceAccount | string | No | — | GCP service account email. If set, takes precedence over the username from the referenced credential secret. Used internally for credential resolution — not sent to the Vault config endpoint as a separate key |
| IAMalias | string | No | `"default"` | Identity alias for IAM roles. Allowed values: `unique_id` (service account unique ID), `role_id` (Vault role ID) |
| IAMmetadata | string | No | `"default"` | Metadata to include on tokens for IAM roles. Defaults include `project_id`, `role`, `service_account_id`, `service_account_email`. Set to `""` to include no metadata |
| GCEalias | string | No | `"role_id"` | Identity alias for GCE roles. Allowed values: `instance_id`, `role_id` |
| GCEmetadata | string | No | `"default"` | Metadata to include on tokens for GCE roles. Defaults include `instance_creation_timestamp`, `instance_id`, `instance_name`, `project_id`, `project_number`, `role`, `service_account_id`, `service_account_email`, `zone`. Set to `""` to include no metadata |
| customEndpoint | JSON | No | `{}` | Overrides to service endpoints for Private Google Access environments. Supported keys: `api`, `iam`, `crm`, `compute`. Format: `scheme://host:port` |

## GCPAuthEngineRole

The `GCPAuthEngineRole` CRD allows you to create a [GCP auth role](https://developer.hashicorp.com/vault/api-docs/auth/gcp#create-update-role).

The GCP auth engine supports two role types: **IAM** roles verify signed JWTs from service accounts, while **GCE** roles verify instance identity tokens from Compute Engine VMs. The `type` field determines which role-specific fields apply.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GCPAuthEngineRole
metadata:
  name: gcp-iam-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: gcp
  name: my-iam-role
  type: iam
  boundServiceAccounts:
    - my-service-account@my-project.iam.gserviceaccount.com
  boundProjects:
    - my-project
  maxJWTExp: "900"
  allowGCEInference: true
  tokenPolicies:
    - app-policy
    - default
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/role/<name> \
    type="iam" \
    bound_service_accounts="my-service-account@my-project.iam.gserviceaccount.com" \
    bound_projects="my-project" \
    max_jwt_exp="900" \
    allow_gce_inference=true \
    token_policies="app-policy,default" \
    token_ttl="1h" \
    token_max_ttl="24h"
```

### Field Descriptions

**Role Identity Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | GCP auth mount path. Full path: `[namespace/]auth/{path}/role/{name}` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | Yes | — | Name of the role in Vault |
| type | string | Yes | — | Type of role. Allowed values: `iam`, `gce` |
| boundServiceAccounts | []string | No | `[]` | Service account emails or IDs that login is restricted to. Set to `*` to allow all service accounts (can be further restricted with `boundProjects`) |
| boundProjects | []string | No | `[]` | GCP project IDs. Only entities belonging to these projects can authenticate |
| addGroupAliases | bool | No | `false` | If `true`, tokens will have group aliases for the entity's project and all ancestor folders/organizations (`project-$PROJECT_ID`, `folder-$FOLDER_ID`, `organization-$ORG_ID`). Requires `resourcemanager.projects.get` IAM permission |

**Token Parameters:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| tokenTTL | string | No | — | Incremental lifetime for generated tokens. Referenced at renewal time |
| tokenMaxTTL | string | No | — | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | — | Policies to attach to generated tokens |
| policies | []string | No | — | **DEPRECATED:** Use `tokenPolicies` instead |
| tokenBoundCIDRs | []string | No | — | CIDR blocks restricting token usage |
| tokenExplicitMaxTTL | string | No | — | Hard cap TTL overriding `tokenTTL` and `tokenMaxTTL` |
| tokenNoDefaultPolicy | bool | No | `false` | Exclude the `default` policy from generated tokens |
| tokenNumUses | int64 | No | `0` | Maximum token usage count. `0` means unlimited |
| tokenPeriod | int64 | No | `0` | Maximum allowed period value when a periodic token is requested |
| tokenType | string | No | — | Token type. Allowed values: `service`, `batch`, `default`, `default-service`, `default-batch` |

**IAM-Only Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| maxJWTExp | string | No | — | Maximum allowed expiration (seconds) for the login JWT past the time of authentication. For example, setting this to `"900"` requires the JWT to expire within 15 minutes of login |
| allowGCEInference | bool | No | `false` | Allow GCE instances to authenticate by inferring service accounts from the GCE identity metadata token |

**GCE-Only Fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| boundZones | []string | No | — | Zones that a GCE instance must belong to. If `boundInstanceGroups` is set, the group must belong to this zone |
| boundRegions | []string | No | — | Regions that a GCE instance must belong to. If `boundInstanceGroups` is set, the group must belong to this region. Ignored if `boundZones` is also set |
| boundInstanceGroups | []string | No | — | Instance groups that the instance must belong to. Requires `boundZones` or `boundRegions` to also be set |
| boundLabels | []string | No | — | GCP labels formatted as `"key:value"` strings that must be set on authorized instances. Recommended to use in conjunction with other restrictions since GCP labels are not ACL'd |

## Credential Resolution

The GCP service account JSON credentials (used to authenticate with Google Cloud APIs) can be retrieved in three different ways via the `GCPCredentials` field. If `GCPCredentials` is omitted or left empty, the operator will use GCP environment credentials (e.g., workload identity or instance metadata).

> **Note:** If `serviceAccount` is set directly in the spec, it takes precedence over the username (service account email) retrieved from the referenced secret. The password (credentials JSON) is always retrieved from the credential source.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  GCPCredentials:
    secret:
      name: gcp-credentials
    usernameKey: serviceaccount
    passwordKey: credentials
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve credentials from another Vault path.

```yaml
spec:
  GCPCredentials:
    vaultSecret:
      path: secret/data/gcp-credentials
    usernameKey: serviceaccount
    passwordKey: credentials
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated. A `serviceAccount` must be specified directly in the spec when using RandomSecret — the service account email is not read from the generated secret.

```yaml
spec:
  serviceAccount: my-sa@my-project.iam.gserviceaccount.com
  GCPCredentials:
    randomSecret:
      name: gcp-random-credentials
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault GCP Auth Method](https://developer.hashicorp.com/vault/docs/auth/gcp) — Vault documentation
- [Vault GCP Auth API](https://developer.hashicorp.com/vault/api-docs/auth/gcp) — Vault API reference
