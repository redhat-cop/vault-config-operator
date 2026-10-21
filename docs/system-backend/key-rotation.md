# Key Rotation Configuration

Vault's [automatic key rotation](https://developer.hashicorp.com/vault/api-docs/system/rotate-config) controls the schedule for automatic barrier key rotation. The vault-config-operator supports managing this configuration declaratively via the `KeyRotationConfig` CRD.

## KeyRotationConfig

The `KeyRotationConfig` CRD manages the key rotation configuration at `sys/rotate/config`. This is a singleton resource — only one CR should exist per Vault instance. If multiple CRs target the same Vault, the last-reconciled wins.

**Important:** `KeyRotationConfig` is non-deletable. Deleting the CR will remove it from Kubernetes, but the rotation configuration will persist in Vault.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KeyRotationConfig
metadata:
  name: key-rotation-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  maxOperations: 3888000000
  interval: "4320h"
  enabled: true
```

### Disabled Rotation Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KeyRotationConfig
metadata:
  name: key-rotation-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  enabled: false
```

### Field Description

- `maxOperations`: The number of operations performed before the barrier key is automatically rotated. Optional; omitted from the payload when zero.
- `interval`: The duration after which the barrier key is automatically rotated (e.g., `"4320h"`). Must be at least `24h` if specified. Optional.
- `enabled`: Controls whether automatic rotation is active. Defaults to `true`. When set to `false`, the boolean value `false` is explicitly sent to Vault.
