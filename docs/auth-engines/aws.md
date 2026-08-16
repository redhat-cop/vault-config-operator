# AWS Auth Engine

[AWS auth method documentation](https://developer.hashicorp.com/vault/docs/auth/aws)

## Overview

AWS auth provides automated authentication to Vault for AWS IAM principals and EC2 instances. It enables workloads running on AWS to authenticate to Vault without managing static credentials, integrating with AWS STS and EC2 metadata services.

The vault-config-operator supports the following CRDs for the AWS auth engine:

- [AWSAuthEngineClientConfig](#awsauthengineclientconfig)
- [AWSAuthEngineIdentityConfig](#awsauthengineidentityconfig)
- [AWSAuthEngineRole](#awsauthenginerole)

## AWSAuthEngineClientConfig

The `AWSAuthEngineClientConfig` CRD allows you to configure the [AWS auth method client settings](https://developer.hashicorp.com/vault/api-docs/auth/aws#configure-client).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSAuthEngineClientConfig
metadata:
  name: aws-auth-client-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: aws
  stsEndpoint: "https://sts.us-east-1.amazonaws.com"
  stsRegion: "us-east-1"
  iamServerIDHeaderValue: "vault.example.com"
  maxRetries: 3
  AWSCredentials:
    secret:
      name: aws-credentials
    usernameKey: access_key
    passwordKey: secret_key
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/aws/config/client \
    access_key=<retrieved dynamically> \
    secret_key=<retrieved dynamically> \
    sts_endpoint="https://sts.us-east-1.amazonaws.com" \
    sts_region="us-east-1" \
    iam_server_id_header_value="vault.example.com" \
    max_retries=3
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the AWS auth engine is mounted. Full path: `[namespace/]auth/{path}/config/client` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| accessKey | string | No | AWS access key ID (can also be resolved from AWSCredentials) |
| endpoint | string | No | Custom URL for EC2 API calls |
| iamEndpoint | string | No | Custom URL for IAM API calls |
| stsEndpoint | string | No | Custom URL for STS API calls |
| stsRegion | string | No | Region for STS API calls (should be set with stsEndpoint) |
| useSTSRegionFromClient | bool | No | Use client request region for STS |
| iamServerIDHeaderValue | string | No | Value to require in the X-Vault-AWS-IAM-Server-ID header |
| allowedSTSHeaderValues | string | No | Comma-separated list of additional permitted STS headers |
| maxRetries | int | No | Max retries for recoverable errors (-1 = AWS SDK default) |
| AWSCredentials | object | Yes | Credential source configuration (secret, vaultSecret, or randomSecret) |

## AWSAuthEngineIdentityConfig

The `AWSAuthEngineIdentityConfig` CRD allows you to configure the [AWS auth method identity settings](https://developer.hashicorp.com/vault/api-docs/auth/aws#configure-identity-access-list-tidy-operation).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSAuthEngineIdentityConfig
metadata:
  name: aws-auth-identity-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: aws
  iamAlias: full_arn
  iamMetadata: default
  ec2Alias: instance_id
  ec2Metadata: default
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/aws/config/identity \
    iam_alias=full_arn \
    iam_metadata=default \
    ec2_alias=instance_id \
    ec2_metadata=default
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the AWS auth engine is mounted. Full path: `[namespace/]auth/{path}/config/identity` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| iamAlias | string | No | Identity alias for IAM auth (role_id, unique_id, canonical_arn, full_arn). Default: role_id |
| iamMetadata | string | No | Metadata on login token for IAM auth. Default: default |
| ec2Alias | string | No | Identity alias for EC2 auth (role_id, instance_id, image_id). Default: role_id |
| ec2Metadata | string | No | Metadata on login token for EC2 auth. Default: default |

## AWSAuthEngineRole

The `AWSAuthEngineRole` CRD allows you to create an [AWS auth method role](https://developer.hashicorp.com/vault/api-docs/auth/aws#create-role).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSAuthEngineRole
metadata:
  name: aws-auth-iam-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: aws
  name: my-iam-role
  authType: iam
  boundIAMPrincipalARN:
    - "arn:aws:iam::123456789012:role/MyRole"
  resolveAWSUniqueIDs: true
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
  tokenPolicies:
    - dev
    - prod
  tokenType: service
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/aws/role/my-iam-role \
    auth_type=iam \
    bound_iam_principal_arn="arn:aws:iam::123456789012:role/MyRole" \
    resolve_aws_unique_ids=true \
    token_ttl=1h \
    token_max_ttl=24h \
    token_policies="dev,prod" \
    token_type=service
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the AWS auth engine is mounted |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | Yes | Name of the Vault role |
| authType | string | No | Auth type: iam or ec2. Default: iam |
| boundAmiID | []string | No | EC2/IAM+infer: constrain to specific AMI IDs |
| boundAccountID | []string | No | EC2/IAM+infer: constrain to specific account IDs |
| boundRegion | []string | No | EC2/IAM+infer: constrain to specific regions |
| boundVpcID | []string | No | EC2/IAM+infer: constrain to specific VPC IDs |
| boundSubnetID | []string | No | EC2/IAM+infer: constrain to specific subnet IDs |
| boundIAMRoleARN | []string | No | EC2/IAM+infer: constrain to specific IAM role ARNs |
| boundIAMInstanceProfileARN | []string | No | EC2/IAM+infer: constrain to specific instance profile ARNs |
| boundEC2InstanceID | []string | No | EC2/IAM+infer: constrain to specific instance IDs |
| roleTag | string | No | EC2 only: tag key for role tag on instances |
| boundIAMPrincipalARN | []string | No | IAM only: constrain to specific principal ARNs (wildcards supported) |
| inferredEntityType | string | No | IAM only: enables EC2 inferencing (value: ec2_instance) |
| inferredAWSRegion | string | No | IAM only: region for inferred entities (required with inferredEntityType) |
| resolveAWSUniqueIDs | bool | No | IAM only: resolve principal ARNs to unique IDs. Default: true |
| allowInstanceMigration | bool | No | EC2 only: allow migration of underlying instance |
| disallowReauthentication | bool | No | EC2 only: only one token per instance ID |
| tokenTTL | string | No | Incremental lifetime for generated tokens |
| tokenMaxTTL | string | No | Maximum lifetime for generated tokens |
| tokenPolicies | []string | No | Policies to encode onto generated tokens |
| tokenBoundCIDRs | []string | No | CIDR blocks that restrict authentication |
| tokenExplicitMaxTTL | string | No | Hard cap max TTL for tokens |
| tokenNoDefaultPolicy | bool | No | Omit default policy from generated tokens |
| tokenNumUses | int | No | Max number of times a token may be used (0 = unlimited) |
| tokenPeriod | int | No | Maximum allowed period for periodic tokens |
| tokenType | string | No | Token type: service, batch, default, default-service, default-batch |

## Credential Resolution

The AWS access key and secret key can be retrieved in three different ways:

### Using a Kubernetes Secret

```yaml
spec:
  AWSCredentials:
    secret:
      name: aws-credentials
    usernameKey: access_key
    passwordKey: secret_key
```

### Using a Vault Secret

```yaml
spec:
  AWSCredentials:
    vaultSecret:
      path: secret/data/aws-credentials
    usernameKey: access_key
    passwordKey: secret_key
```

### Using a RandomSecret

```yaml
spec:
  accessKey: "AKIAIOSFODNN7EXAMPLE"
  AWSCredentials:
    randomSecret:
      name: aws-random-secret-key
    usernameKey: access_key
    passwordKey: secret_key
```

Note: When using `randomSecret`, `spec.accessKey` must be set explicitly because the RandomSecret only provides the secret key.

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault AWS Auth Method API](https://developer.hashicorp.com/vault/api-docs/auth/aws) — Vault API reference
