# UI Custom Headers

[Vault UI custom headers documentation](https://developer.hashicorp.com/vault/api-docs/system/config-ui-headers)

## Overview

Vault allows operators to configure custom response headers for the UI. These headers are returned with every response from Vault's HTTP API when serving UI requests. This is useful for setting security headers (e.g., `X-Frame-Options`, `Content-Security-Policy`) or custom application headers.

The vault-config-operator supports managing UI custom headers declaratively via the `UICustomHeader` CRD. Each CR represents a single header name and its associated value(s).

## UICustomHeader

The `UICustomHeader` CRD allows you to configure a [Vault UI custom header](https://developer.hashicorp.com/vault/api-docs/system/config-ui-headers).

**Important:** The Vault token used by the operator must have the `sudo` capability on `sys/config/ui/headers/*`. The root token has all capabilities including `sudo`.

**Important:** `UICustomHeader` is deletable. Deleting the CR will remove the header from Vault by issuing a DELETE to `sys/config/ui/headers/{name}`.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: UICustomHeader
metadata:
  name: x-custom-header
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  values:
    - "custom-value-1"
```

### Multi-Value Header Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: UICustomHeader
metadata:
  name: test-header-metadata
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: X-Custom-Multi-Header
  values:
    - "value-one"
    - "value-two"
```

### Vault CLI Equivalent

```shell
vault write sys/config/ui/headers/X-Custom-Header values=custom-value-1
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | The name of the custom header in Vault. If specified, takes precedence over `metadata.name`. Allows uppercase letters (e.g., `X-Custom-Header`). Pattern: `[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?` |
| values | []string | Yes | The list of values to return for this custom header. Must contain at least one entry. |

### Known Limitations

For `UICustomHeader` CRs with multiple values, the operator will re-write the header on every reconcile cycle. This is because Vault's standard `GET sys/config/ui/headers/{name}` returns only a single value (the `value` key) without the `?multivalue=true` query parameter, and the operator's framework does not support query parameters on reads. Since the POST is idempotent, this is safe — the reconcile frequency is controlled by `SYNC_PERIOD_SECONDS` (default 10 hours).

## See Also

- [Vault API: /sys/config/ui/headers](https://developer.hashicorp.com/vault/api-docs/system/config-ui-headers)
