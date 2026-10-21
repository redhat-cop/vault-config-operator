# Quota Management

Vault's [resource quotas](https://developer.hashicorp.com/vault/docs/concepts/resource-quotas) allow operators to control the rate of requests and the number of leases across the cluster. The vault-config-operator supports the following quota-related CRDs:

- [QuotaConfig](#quotaconfig) — Global quota configuration
- [RateLimitQuota](#ratelimitquota) — Per-path/role request rate limits
- [LeaseCountQuota](#leasecountquota) — Per-path/role lease count limits (Enterprise only)

## QuotaConfig

The `QuotaConfig` CRD manages the global quota configuration at `sys/quotas/config`. This is a singleton resource — only one CR should exist per Vault instance. If multiple CRs target the same Vault, the last-reconciled wins.

**Important:** `QuotaConfig` is non-deletable. Deleting the CR will remove it from Kubernetes, but the global configuration will persist in Vault.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: QuotaConfig
metadata:
  name: global-quota-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  enableRateLimitAuditLogging: true
  enableRateLimitResponseHeaders: true
  rateLimitExemptPaths:
    - sys/health
    - sys/init
```

### Field Description

- `enableRateLimitAuditLogging`: When `true`, rate-limit quota rejections are included in audit log entries.
- `enableRateLimitResponseHeaders`: When `true`, Vault returns rate-limit quota information in HTTP response headers.
- `rateLimitExemptPaths`: A list of path suffixes exempted from rate limiting. Vault may add its own built-in paths.
- `absoluteRateLimitExemptPaths`: A list of exact paths exempted from rate limiting, overriding any built-in behaviour.

## RateLimitQuota

The `RateLimitQuota` CRD manages named rate-limit quotas at `sys/quotas/rate-limit/{name}`. Rate-limit quotas restrict the number of requests per second to Vault, optionally scoped to a specific path and/or role.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RateLimitQuota
metadata:
  name: global-rate-limit
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  rate: 100
  interval: "1m"
  blockInterval: "5m"
```

### Path-scoped Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RateLimitQuota
metadata:
  name: secret-rate-limit
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  rate: 50
  interval: "30s"
  quotaPath: "secret/"
```

### Field Description

- `name`: Optional override for the quota name in Vault. If omitted, `metadata.name` is used. Immutable after creation.
- `rate`: The maximum number of requests allowed per interval. Must be greater than 0.
- `interval`: The evaluation window (e.g., `"1m"`, `"30s"`). Defaults to `"1s"` if omitted.
- `blockInterval`: How long a client is blocked when the rate is exceeded (e.g., `"5m"`). Defaults to `"0s"` (no block).
- `quotaPath`: Optional Vault API path prefix to scope this quota to.
- `role`: Optional auth role to scope this quota to.
- `inheritable`: If `true`, sub-namespaces inherit this quota.

## LeaseCountQuota

> **Enterprise Only:** Lease count quotas require Vault Enterprise. Applying this CRD against an OSS Vault instance will result in a `ReconcileFailed` condition.

The `LeaseCountQuota` CRD manages named lease-count quotas at `sys/quotas/lease-count/{name}`. Lease-count quotas restrict the total number of leases that can be generated, optionally scoped to a specific path and/or role.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LeaseCountQuota
metadata:
  name: global-lease-limit
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  maxLeases: 10000
```

### Path-scoped Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LeaseCountQuota
metadata:
  name: database-lease-limit
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  maxLeases: 1000
  quotaPath: "database/"
```

### Field Description

- `name`: Optional override for the quota name in Vault. If omitted, `metadata.name` is used. Immutable after creation.
- `maxLeases`: The maximum number of leases allowed. Must be at least 1.
- `quotaPath`: Optional Vault API path prefix to scope this quota to.
- `role`: Optional auth role to scope this quota to.
- `inheritable`: If `true`, sub-namespaces inherit this quota.
