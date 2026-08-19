# OCI Auth Engine

[OCI auth method documentation](https://developer.hashicorp.com/vault/docs/auth/oci)

## Overview

OCI (Oracle Cloud Infrastructure) authentication allows OCI instances and users to authenticate with Vault using OCI identity principals. It enables workloads running on OCI compute instances to obtain Vault tokens without managing static credentials, using instance principal authentication provided by the OCI platform.

The vault-config-operator supports the following CRDs for the OCI engine:

- [OCIAuthEngineConfig](#ociauthengineconfig)
- [OCIAuthEngineRole](#ociauthenginerole)

## OCIAuthEngineConfig

The `OCIAuthEngineConfig` CRD allows you to configure an authentication engine mount of [type OCI](https://developer.hashicorp.com/vault/api-docs/auth/oci).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: OCIAuthEngineConfig
metadata:
  name: oci-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: oci
  homeTenancyID: "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/oci/config \
    home_tenancy_id="ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the OCI auth engine is mounted. Full path: `[namespace/]auth/{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| homeTenancyID | string | Yes | The Tenancy OCID of the OCI account |

## OCIAuthEngineRole

The `OCIAuthEngineRole` CRD allows you to create an [OCI authentication role](https://developer.hashicorp.com/vault/api-docs/auth/oci#create-role).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: OCIAuthEngineRole
metadata:
  name: oci-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: oci
  name: oci-role
  ocidList:
    - "ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq"
    - "ocid1.dynamicgroup.oc1..aaaaaaaa5hmfyrdaxvmt52ekju5n7ffamn2pdvxaq6esb2vzzoduexamplea"
  tokenTTL: "30m"
  tokenMaxTTL: "1h"
  tokenPolicies:
    - dev
    - prod
  tokenType: "service"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/oci/role/oci-role \
    ocid_list="ocid1.group.oc1..example,ocid1.dynamicgroup.oc1..example" \
    token_ttl=1800 \
    token_max_ttl=3600 \
    token_policies="dev,prod" \
    token_type="service"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the OCI auth engine mount where the role will be created |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Optional Vault role name override. If omitted, `metadata.name` is used as the role name in the Vault path |
| ocidList | []string | Yes | List of Group or Dynamic Group OCIDs that can take this role |
| tokenTTL | string | No | Incremental lifetime for generated tokens (e.g., "30m", "1h") |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | Policies to encode onto generated tokens |
| policies | []string | No | Deprecated — use tokenPolicies instead |
| tokenBoundCIDRs | []string | No | CIDR blocks restricting authentication |
| tokenExplicitMaxTTL | string | No | Hard cap max TTL for tokens |
| tokenNoDefaultPolicy | bool | No | If true, omits the default policy from generated tokens |
| tokenNumUses | int | No | Max number of times a generated token may be used (0 = unlimited) |
| tokenPeriod | int | No | Maximum allowed period for periodic tokens |
| tokenType | string | No | Token type: service, batch, default, default-service, default-batch |

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [OCI Auth Method API](https://developer.hashicorp.com/vault/api-docs/auth/oci) — Vault API documentation
