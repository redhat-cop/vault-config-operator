# AWS Secret Engine

[AWS engine documentation](https://developer.hashicorp.com/vault/docs/secrets/aws)

## Overview

The AWS secret engine generates IAM credentials (access keys, STS tokens, or assumed-role credentials) on demand, with automatic revocation after the configured TTL expires. It enables applications to access AWS services without managing long-lived static credentials. The operator manages both the root configuration (AWS account connection) and the roles that define what type of credentials are generated.

The vault-config-operator supports the following CRDs for the AWS engine:

- [AWSSecretEngineConfig](#awssecretengineconfig)
- [AWSSecretEngineRole](#awssecretenginerole)

## AWSSecretEngineConfig

The `AWSSecretEngineConfig` CRD allows you to configure an [AWS secret engine connection](https://developer.hashicorp.com/vault/api-docs/secret/aws#configure-root-iam-credentials).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSSecretEngineConfig
metadata:
  name: aws-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: aws
  accessKey: AKIAIOSFODNN7EXAMPLE
  region: us-east-1
  rootCredentials:
    secret:
      name: aws-admin-credentials
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the secret engine being configured. They may point to different mounts.

> **Note:** Deleting the `AWSSecretEngineConfig` CR from Kubernetes will **not** delete the configuration from Vault (`IsDeletable=false`). The Vault configuration persists until the engine mount is removed.

### Vault CLI Equivalent

```shell
vault write [namespace/]aws/config/root \
    access_key=AKIAIOSFODNN7EXAMPLE \
    secret_key=<retrieved from credentials> \
    region=us-east-1
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/config/root` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| accessKey | string | No | AWS access key ID. If not set, retrieved from the credential source |
| region | string | No | AWS region. Defaults to `us-east-1` if not set |
| iamEndpoint | string | No | Custom HTTP endpoint for IAM API calls |
| stsEndpoint | string | No | Custom HTTP endpoint for STS API calls |
| maxRetries | int | No | Maximum number of retries for recoverable errors. `-1` uses the AWS SDK default |
| usernameTemplate | string | No | Go template for dynamic IAM username generation |
| rootCredentials | object | Yes | Credential source for the AWS connection. See [Credential Resolution](#credential-resolution) |

## AWSSecretEngineRole

The `AWSSecretEngineRole` CRD allows you to create an [AWS secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/aws#create-update-role) for generating dynamic AWS credentials.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSSecretEngineRole
metadata:
  name: s3-read-only
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: aws
  credentialType: iam_user
  policyArns:
    - "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]aws/roles/s3-read-only \
    credential_type=iam_user \
    policy_arns="arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| credentialType | string | Yes | Type of credential to generate: `iam_user`, `assumed_role`, `federation_token`, or `session_token` |
| roleArns | []string | Conditional | ARNs of AWS roles to assume. Required when `credentialType` is `assumed_role` |
| policyArns | []string | No | AWS managed policy ARNs to attach. Not valid for `session_token` |
| policyDocument | string | No | Inline IAM policy document (JSON). Not valid for `session_token` |
| iamGroups | []string | No | IAM group names. For `iam_user`, generated users are added to these groups. For `assumed_role` and `federation_token`, group policies are combined with `policyArns`/`policyDocument`. Not valid for `session_token` |
| iamTags | []string | No | Key=value tags for `iam_user` credential type only |
| defaultSTSTTL | string | No | Default TTL for STS credentials. Only valid for `assumed_role` and `federation_token` |
| maxSTSTTL | string | No | Maximum TTL for STS credentials. Only valid for `assumed_role` and `federation_token` |
| userPath | string | No | IAM user path. Only valid for `iam_user` |
| permissionsBoundaryARN | string | No | Permissions boundary ARN. Only valid for `iam_user` |
| externalID | string | No | External ID for assume-role operations. Only valid for `assumed_role` |
| sessionTags | []string | No | STS session tags. Only valid for `assumed_role` |
| mfaSerialNumber | string | No | MFA device serial number or ARN |

## Credential Resolution

The root credentials (AWS access key and secret key) for the engine connection can be retrieved in three different ways via the `rootCredentials` field. Exactly one of `secret`, `vaultSecret`, or `randomSecret` must be specified.

> **Note:** When using `randomSecret`, `spec.accessKey` must be specified in the CRD because `randomSecret` only provides the secret key.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  rootCredentials:
    secret:
      name: aws-admin-credentials
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve credentials from another Vault path.

```yaml
spec:
  rootCredentials:
    vaultSecret:
      path: secret/data/aws-credentials
    usernameKey: username
    passwordKey: password
```

### Using a RandomSecret

Specify the `randomSecret` field. When the [RandomSecret](../secret-management.md#randomsecret) generates a new secret, this configuration will also be updated. A `spec.accessKey` must be specified when using RandomSecret.

```yaml
spec:
  accessKey: AKIAIOSFODNN7EXAMPLE
  rootCredentials:
    randomSecret:
      name: aws-random-secret-key
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault AWS Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/aws) — Vault documentation
- [Vault AWS Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/aws) — Vault API reference
