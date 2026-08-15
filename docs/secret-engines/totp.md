# TOTP Secret Engine

[TOTP engine documentation](https://developer.hashicorp.com/vault/api-docs/secret/totp)

## Overview

The TOTP (Time-based One-Time Password) secret engine generates time-based credentials per the TOTP standard (RFC 6238). It can act as both a TOTP code generator (for services that require TOTP codes) and a TOTP code provider (generating keys for users). The engine is self-contained within Vault and requires no external service dependencies.

The vault-config-operator supports the following CRD for the TOTP engine:

- [TOTPSecretEngineKey](#totpsecretenginekey)

## TOTPSecretEngineKey

The `TOTPSecretEngineKey` CRD allows you to manage [TOTP keys](https://developer.hashicorp.com/vault/api-docs/secret/totp#create-key) in a TOTP secret engine mount. Keys can be created in two modes: **generate mode** (Vault generates the secret internally) or **import mode** (you provide an existing TOTP secret).

### Example — Generate Mode

In generate mode, Vault creates the TOTP secret key internally:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: TOTPSecretEngineKey
metadata:
  name: my-totp-key
spec:
  authentication:
    path: kubernetes
    role: totp-engine-admin
  path: my-totp-mount
  generate: true
  issuer: MyOrganization
  accountName: user@example.com
  algorithm: SHA1
  digits: 6
  period: 30
```

### Example — Import Mode

In import mode, you provide an existing Base32-encoded key or an `otpauth://` URL:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: TOTPSecretEngineKey
metadata:
  name: my-imported-totp-key
spec:
  authentication:
    path: kubernetes
    role: totp-engine-admin
  path: my-totp-mount
  generate: false
  key: JBSWY3DPEHPK3PXP
  issuer: ExternalService
  accountName: user@example.com
  algorithm: SHA1
  digits: 6
  period: 30
```

### Vault CLI Equivalent

```shell
vault write [namespace/]my-totp-mount/keys/my-totp-key \
    generate=true \
    issuer=MyOrganization \
    account_name=user@example.com \
    algorithm=SHA1 \
    digits=6 \
    period=30
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the TOTP engine is mounted. Full path: `[namespace/]{path}/keys/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name (defaults to `metadata.name`) |
| generate | bool | No | `true` = Vault generates the key; `false` = import an existing key (default: `false`) |
| exported | bool | No | Return QR code and URL on generate. Only used when `generate=true` |
| keySize | int | No | Key size in bytes for generated keys. Only used when `generate=true` (default: 20) |
| url | string | No | TOTP key URL (`otpauth://` format). Only used when `generate=false` |
| key | string | No | Base32-encoded root key for importing. Only used when `generate=false` |
| issuer | string | Yes | Name of the key's issuing organization |
| accountName | string | Yes | Name of the account associated with the key |
| period | int | No | Counter period in seconds (default: 30) |
| algorithm | string | No | Hash algorithm: `SHA1`, `SHA256`, or `SHA512` (default: `SHA1`) |
| digits | int | No | Number of digits in the TOTP code: `6` or `8` (default: 6) |
| skew | int | No | Allowed delay periods for validation: `0` or `1`. Only used when `generate=true` |
| qrSize | int | No | QR code pixel size. Only used when `generate=true` and `exported=true`. `0` = no QR |

### Immutable Fields

The following fields cannot be changed after creation:

- `spec.path` — The engine mount path
- `spec.name` — The Vault object name override
- `spec.generate` — Generation mode (generate vs import)
- `spec.key` — Base32-encoded root key (import mode)
- `spec.url` — TOTP key URL (import mode)
- `spec.keySize` — Generated key size
- `spec.exported` — QR code export flag
- `spec.skew` — Allowed delay periods (write-only, not verifiable via drift detection)
- `spec.qrSize` — QR code pixel size (write-only, not verifiable via drift detection)

Fields `spec.algorithm`, `spec.digits`, `spec.period`, `spec.issuer`, and `spec.accountName` may be updated after creation.

### Required Fields by Mode

- **Generate mode** (`generate: true`): `issuer` and `accountName` are required.
- **Import mode** (`generate: false`): `key` or `url` is required, plus `issuer` and `accountName` (required even when `url` is provided, to ensure drift detection works correctly).

> **Note:** When using `spec.url` for import, the explicit `spec.issuer` and `spec.accountName` values should match the values encoded in the `otpauth://` URL. The operator uses the spec fields (not parsed URL values) for drift detection. A mismatch between spec fields and URL-encoded values will cause reconciliation drift on every sync cycle.

### Deletion Behavior

When a `TOTPSecretEngineKey` CR is deleted from Kubernetes, the corresponding key is also deleted from Vault (`IsDeletable=true`).

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault TOTP Secrets Engine API](https://developer.hashicorp.com/vault/api-docs/secret/totp) — Vault API reference
