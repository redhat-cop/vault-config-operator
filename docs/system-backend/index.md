# System Backend

The vault-config-operator supports managing Vault's system backend configuration declaratively via the following CRDs:

## Quota Management

- [QuotaConfig](quotas.md#quotaconfig) — Global quota configuration (rate-limit audit logging, exempt paths)
- [RateLimitQuota](quotas.md#ratelimitquota) — Rate limit quota configuration
- [LeaseCountQuota](quotas.md#leasecountquota) — Lease count quota configuration (Enterprise only)

## System Configuration

- [CORSConfig](cors.md#corsconfig) — CORS settings (allowed origins and headers)
- [KeyRotationConfig](key-rotation.md#keyrotationconfig) — Automatic key rotation schedule
- [RotationPolicy](rotation-policy.md#rotationpolicy) — Credential rotation retry behavior (Enterprise only)

## UI and Observability

- [UICustomHeader](ui-headers.md#uicustomheader) — Custom UI response headers
- [LoggerConfig](loggers.md#loggerconfig) — Runtime logger level configuration
