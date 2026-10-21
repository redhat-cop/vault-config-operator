# Rotation Policy

> **⚠️ Vault Enterprise Required:** The `/sys/policies/rotation` API requires a Vault Enterprise license. This CRD will produce a 4xx error when used against OSS Vault.

[Vault Rotation Policy API documentation](https://developer.hashicorp.com/vault/api-docs/system/policies-rotation)

## Overview

Rotation policies control Vault's automatic credential rotation retry behavior. They define how many times Vault should retry a failed rotation and how many cycles of retries should be attempted before giving up.

The vault-config-operator supports the following CRD for rotation policies:

- [RotationPolicy](#rotationpolicy)

## RotationPolicy

The `RotationPolicy` CRD allows you to configure a [Vault Rotation Policy](https://developer.hashicorp.com/vault/api-docs/system/policies-rotation).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RotationPolicy
metadata:
  name: my-rotation-policy
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  # Enterprise Vault required
  maxRetriesPerCycle: 3
  maxRetryCycles: 5
```

With `spec.name` override (Vault object name differs from K8s metadata.name):

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RotationPolicy
metadata:
  name: k8s-resource-name
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: vault-rotation-policy-name
  maxRetriesPerCycle: 3
  maxRetryCycles: 5
```

### Vault CLI Equivalent

```shell
vault write sys/policies/rotation/my-rotation-policy max_retries_per_cycle=3 max_retry_cycles=5
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Name of the rotation policy in Vault. If specified, takes precedence over `metadata.name`. Immutable after creation. |
| maxRetriesPerCycle | int | No | Maximum number of times Vault applies backoff for rotation entries after a rotation failure |
| maxRetryCycles | int | No | Maximum number of times Vault requeues rotation entries after reaching the `maxRetriesPerCycle` limit |

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Rotation Policy API](https://developer.hashicorp.com/vault/api-docs/system/policies-rotation) — Vault API reference
