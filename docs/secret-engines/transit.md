# Transit Secret Engine

[Transit engine documentation](https://developer.hashicorp.com/vault/docs/secrets/transit)

## Overview

The Transit secret engine provides encryption-as-a-service, allowing applications to encrypt and decrypt data without managing encryption keys directly. It handles key management, key rotation, and cryptographic operations (encrypt, decrypt, sign, verify, hash, HMAC) server-side, enabling applications to offload all cryptographic concerns to Vault.

The vault-config-operator supports the following CRD for the Transit engine:

- [TransitSecretEngineKey](#transitsecretenginekey)

## TransitSecretEngineKey

The `TransitSecretEngineKey` CRD allows you to create a [Transit encryption key](https://developer.hashicorp.com/vault/api-docs/secret/transit#create-key).

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the secret engine being configured. They may point to different mounts.

> **Create-time vs config-time fields:** Some fields (`type`, `derived`, `convergentEncryption`, `keySize`) are immutable after key creation. Other fields (`minDecryptionVersion`, `minEncryptionVersion`, `deletionAllowed`, `exportable`, `allowPlaintextBackup`, `autoRotatePeriod`) can be updated at any time via the key's config endpoint.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: TransitSecretEngineKey
metadata:
  name: my-encryption-key
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: transit
  type: aes256-gcm96
  deletionAllowed: true
  autoRotatePeriod: "720h"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]transit/keys/my-encryption-key \
    type=aes256-gcm96

vault write [namespace/]transit/keys/my-encryption-key/config \
    deletion_allowed=true \
    auto_rotate_period=720h
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/keys/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| type | string | No | Key algorithm. Immutable after creation. Default: `aes256-gcm96`. Allowed values: `aes128-gcm96`, `aes256-gcm96`, `chacha20-poly1305`, `ed25519`, `ecdsa-p256`, `ecdsa-p384`, `ecdsa-p521`, `rsa-2048`, `rsa-3072`, `rsa-4096`, `hmac` |
| derived | bool | No | Enable key derivation mode. Immutable after creation |
| convergentEncryption | bool | No | Enable convergent encryption mode. Requires `derived=true`. Immutable after creation |
| keySize | int | No | Key size in bytes. Only applicable to HMAC keys (32-512 bytes). Immutable after creation |
| minDecryptionVersion | int | No | Minimum version of ciphertext allowed to be decrypted. Mutable |
| minEncryptionVersion | int | No | Minimum version of the key that can be used to encrypt plaintext. Mutable |
| deletionAllowed | bool | No | Whether the key is allowed to be deleted. Mutable |
| exportable | bool | No | Whether the key is exportable. One-way: can be set to `true` but never unset. Mutable |
| allowPlaintextBackup | bool | No | Whether plaintext backup of the key is allowed. One-way: can be set to `true` but never unset. Mutable |
| autoRotatePeriod | string | No | Period for automatic key rotation (duration format, e.g., `720h`). `"0"` disables auto-rotation. Minimum `1h` when enabled. Mutable |

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Transit Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/transit) — Vault documentation
- [Vault Transit Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/transit) — Vault API reference
