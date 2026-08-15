# Terraform Cloud Secret Engine

[Terraform Cloud engine documentation](https://developer.hashicorp.com/vault/docs/secrets/terraform)

## Overview

The Terraform Cloud secret engine generates dynamic Terraform Cloud API tokens for organizations, teams, and users. It enables automated credential management for Terraform Cloud/Enterprise workflows and integrates with HCP Terraform's token lifecycle.

The vault-config-operator supports the following CRDs for the Terraform Cloud engine:

- [TerraformCloudSecretEngineConfig](#terraformcloudsecretengineconfig)
- [TerraformCloudSecretEngineRole](#terraformcloudsecretenginerole)

## TerraformCloudSecretEngineConfig

The `TerraformCloudSecretEngineConfig` CRD allows you to configure a [Terraform Cloud secret engine](https://developer.hashicorp.com/vault/api-docs/secret/terraform#write-configuration).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: TerraformCloudSecretEngineConfig
metadata:
  name: tfc-config
spec:
  authentication:
    path: kubernetes
    role: tfc-engine-admin
  path: terraform
  address: "https://app.terraform.io"
  tfcCredentials:
    secret:
      name: tfc-api-token
    passwordKey: token
```

### Vault CLI Equivalent

```shell
vault write [namespace/]terraform/config \
    address="https://app.terraform.io" \
    token=<retrieved dynamically>
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the Terraform Cloud engine is mounted. Full path: `[namespace/]{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| address | string | No | URL of the Terraform Cloud/Enterprise instance. Defaults to `https://app.terraform.io` |
| tfcCredentials | object | Yes | Credential source for the Terraform Cloud API token. See [Credential Resolution](#credential-resolution) |

## TerraformCloudSecretEngineRole

The `TerraformCloudSecretEngineRole` CRD allows you to create a [Terraform Cloud role](https://developer.hashicorp.com/vault/api-docs/secret/terraform#create-update-role) for generating dynamic API tokens.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: TerraformCloudSecretEngineRole
metadata:
  name: tfc-user-role
spec:
  authentication:
    path: kubernetes
    role: tfc-engine-admin
  path: terraform
  credentialType: user
  userID: user-glhf1234
  description: "vault-generated user token"
  ttl: "3600"
  maxTTL: "86400"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]terraform/role/tfc-user-role \
    credential_type=user \
    user_id=user-glhf1234 \
    description="vault-generated user token" \
    ttl=3600 \
    max_ttl=86400
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the Terraform Cloud engine mount where the role will be created |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| organization | string | No | Terraform Cloud organization name. Conflicts with `userID` |
| teamID | string | No | Terraform Cloud team ID. Conflicts with `userID` |
| userID | string | No | Terraform Cloud user ID. Conflicts with `organization` and `teamID` |
| credentialType | string | No | Type of credential: `team`, `team_legacy`, `user`, or `organization` |
| description | string | No | Description prefix for the token in HCP Terraform UI |
| ttl | string | No | TTL for generated tokens. Written as numeric seconds to Vault (e.g. `"3600"`) matching Vault's read format for consistent drift detection. Duration strings (e.g. `"1h"`) are also accepted and converted to seconds. |
| maxTTL | string | No | Maximum TTL for generated tokens. Written as numeric seconds to Vault (e.g. `"86400"`) matching Vault's read format for consistent drift detection. Duration strings (e.g. `"24h"`) are also accepted and converted to seconds. |

## Credential Resolution

The Terraform Cloud API token can be retrieved in three different ways:

### Using a Kubernetes Secret

Specify the `tfcCredentials.secret` field. The secret must contain a key matching `passwordKey` (defaults to `token`). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  tfcCredentials:
    secret:
      name: tfc-api-token
    passwordKey: token
```

### Using a Vault Secret

Specify the `tfcCredentials.vaultSecret` field to retrieve the token from another Vault path.

```yaml
spec:
  tfcCredentials:
    vaultSecret:
      path: secret/data/tfc-credentials
    passwordKey: token
```

### Using a RandomSecret

Specify the `tfcCredentials.randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated.

```yaml
spec:
  tfcCredentials:
    randomSecret:
      name: tfc-random-token
    passwordKey: token
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Terraform Cloud Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/terraform) — Vault API reference
