# GCP Secret Engine

[GCP engine documentation](https://developer.hashicorp.com/vault/docs/secrets/gcp)

## Overview

The GCP secret engine generates Google Cloud IAM service account credentials (OAuth2 access tokens or service account keys) on demand, with automatic revocation after the configured TTL expires. It supports two credential management models: rolesets (Vault-managed service accounts) and static accounts (pre-existing service accounts managed by Vault). This enables applications to access GCP services without managing long-lived service account keys.

The vault-config-operator supports the following CRDs for the GCP engine:

- [GCPSecretEngineConfig](#gcpsecretengineconfig)
- [GCPSecretEngineRoleset](#gcpsecretengineroleset)
- [GCPSecretEngineStaticAccount](#gcpsecretenginestaticaccount)

## GCPSecretEngineConfig

The `GCPSecretEngineConfig` CRD allows you to configure a [GCP secret engine](https://developer.hashicorp.com/vault/api-docs/secret/gcp#write-config).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GCPSecretEngineConfig
metadata:
  name: gcp-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: gcp
  gcpCredentials:
    secret:
      name: gcp-service-account-key
  ttl: "1h"
  maxTTL: "24h"
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the secret engine being configured. They may point to different mounts.

> **Note:** Deleting the `GCPSecretEngineConfig` CR from Kubernetes will **not** delete the configuration from Vault (`IsDeletable=false`). The Vault configuration persists until the engine mount is removed.

### Vault CLI Equivalent

```shell
vault write [namespace/]gcp/config \
    credentials=@/path/to/service-account-key.json \
    ttl=1h \
    max_ttl=24h
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| ttl | string | No | Default TTL for long-lived credentials such as service account keys (duration format) |
| maxTTL | string | No | Maximum TTL for long-lived credentials (duration format) |
| gcpCredentials | object | Yes | Credential source for the GCP service account key. See [Credential Resolution](#credential-resolution) |

## GCPSecretEngineRoleset

The `GCPSecretEngineRoleset` CRD allows you to create a [GCP secret engine roleset](https://developer.hashicorp.com/vault/api-docs/secret/gcp#create-update-roleset). Rolesets create and manage a dedicated GCP service account per roleset, with IAM bindings applied according to the specified configuration.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GCPSecretEngineRoleset
metadata:
  name: gcs-reader
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: gcp
  secretType: access_token
  project: my-gcp-project
  bindings: |
    resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" {
      roles = ["roles/storage.objectViewer"]
    }
  tokenScopes:
    - "https://www.googleapis.com/auth/cloud-platform"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]gcp/roleset/gcs-reader \
    secret_type=access_token \
    project=my-gcp-project \
    bindings='resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" { roles = ["roles/storage.objectViewer"] }' \
    token_scopes="https://www.googleapis.com/auth/cloud-platform"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/roleset/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| secretType | string | Yes | Type of secret generated: `access_token` or `service_account_key`. Cannot be updated after creation. Default: `access_token` |
| project | string | Yes | GCP project ID that this roleset's service account belongs to. Cannot be updated after creation |
| bindings | string | No | IAM bindings configuration string. Both JSON and HCL formats are supported |
| tokenScopes | []string | No | List of OAuth scopes for `access_token` type rolesets only |

## GCPSecretEngineStaticAccount

The `GCPSecretEngineStaticAccount` CRD allows you to create a [GCP secret engine static account](https://developer.hashicorp.com/vault/api-docs/secret/gcp#create-update-static-account). Static accounts manage credentials for a pre-existing GCP service account, without creating a new one.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GCPSecretEngineStaticAccount
metadata:
  name: my-static-sa
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: gcp
  secretType: access_token
  serviceAccountEmail: my-sa@my-gcp-project.iam.gserviceaccount.com
  bindings: |
    resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" {
      roles = ["roles/viewer"]
    }
  tokenScopes:
    - "https://www.googleapis.com/auth/cloud-platform"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]gcp/static-account/my-static-sa \
    secret_type=access_token \
    service_account_email=my-sa@my-gcp-project.iam.gserviceaccount.com \
    bindings='resource "//cloudresourcemanager.googleapis.com/projects/my-gcp-project" { roles = ["roles/viewer"] }' \
    token_scopes="https://www.googleapis.com/auth/cloud-platform"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/static-account/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| secretType | string | Yes | Type of secret generated: `access_token` or `service_account_key`. Cannot be updated after creation. Default: `access_token` |
| serviceAccountEmail | string | Yes | Email of the pre-existing GCP service account to manage. Cannot be updated after creation |
| bindings | string | No | IAM bindings configuration string. Both JSON and HCL formats are supported |
| tokenScopes | []string | No | List of OAuth scopes for `access_token` type static accounts only |

## Credential Resolution

The GCP service account credentials JSON can be retrieved in two different ways via the `gcpCredentials` field. Exactly one of `secret` or `vaultSecret` must be specified.

> **Note:** The `passwordKey` defaults to `"credentials"` (the standard GCP service account key field name) instead of the usual `"password"`.

### Using a Kubernetes Secret

Specify the `secret` field. The secret should contain the GCP service account key JSON.

```yaml
spec:
  gcpCredentials:
    secret:
      name: gcp-service-account-key
    passwordKey: credentials
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve the credentials from another Vault path.

```yaml
spec:
  gcpCredentials:
    vaultSecret:
      path: secret/gcp-credentials
    passwordKey: credentials
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault GCP Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/gcp) — Vault documentation
- [Vault GCP Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/gcp) — Vault API reference
