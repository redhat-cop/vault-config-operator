# Logger Configuration

[Vault loggers documentation](https://developer.hashicorp.com/vault/api-docs/system/loggers)

## Overview

Vault allows operators to modify the log verbosity level of individual loggers at runtime. This is useful for debugging specific subsystems (e.g., `core`, `audit`, `expiration`) without restarting Vault or changing its configuration file.

The vault-config-operator supports managing logger levels declaratively via the `LoggerConfig` CRD. Each CR represents a single logger's desired verbosity level.

## LoggerConfig

The `LoggerConfig` CRD allows you to configure a [Vault logger level](https://developer.hashicorp.com/vault/api-docs/system/loggers).

**Important — Transient Nature:** Logger level changes made via this API are **not persisted across Vault restarts**. When Vault is restarted or reloaded, logger levels revert to the configured default (from `vault.hcl` or `VAULT_LOG_LEVEL`). This is correct declarative behavior — the operator's periodic reconcile (drift detection) will automatically re-apply the desired level after a Vault restart.

**Important:** `LoggerConfig` is deletable. Deleting the CR will revert the logger to Vault's configured default level by issuing a DELETE to `sys/loggers/{name}`.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LoggerConfig
metadata:
  name: core
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  level: debug
```

### Named Logger Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LoggerConfig
metadata:
  name: logger-metadata-name
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: audit
  level: trace
```

### Vault CLI Equivalent

```shell
vault write sys/loggers/core level=debug
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | The name of the Vault logger to configure (e.g., `audit`, `core`, `expiration`). If specified, takes precedence over `metadata.name`. Pattern: `[a-z0-9]([-a-z0-9]*[a-z0-9])?` |
| level | string | Yes | The log verbosity level. Supported values in order of increasing detail: `error`, `warn`, `info`, `debug`, `trace`. |

## See Also

- [Vault API: /sys/loggers](https://developer.hashicorp.com/vault/api-docs/system/loggers)
