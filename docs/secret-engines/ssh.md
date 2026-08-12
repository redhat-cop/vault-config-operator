# SSH Secret Engine

[SSH engine documentation](https://developer.hashicorp.com/vault/docs/secrets/ssh)

## Overview

The SSH secret engine provides signed SSH certificates and one-time passwords for secure SSH access. It can act as a Certificate Authority (CA) for signing SSH host and user certificates, or generate one-time passwords for SSH sessions. This enables centralized SSH access management without distributing static SSH keys.

The vault-config-operator supports the following CRDs for the SSH engine:

- [SSHSecretEngineConfig](#sshsecretengineconfig)
- [SSHSecretEngineRole](#sshsecretenginerole)

## SSHSecretEngineConfig

The `SSHSecretEngineConfig` CRD allows you to configure an [SSH secret engine CA](https://developer.hashicorp.com/vault/api-docs/secret/ssh#submit-ca-information).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SSHSecretEngineConfig
metadata:
  name: ssh-ca-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: ssh
  generateSigningKey: true
  keyType: ssh-rsa
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the secret engine being configured. They may point to different mounts.

> **Write-once CA key pair:** The SSH CA key pair is generated once on first write to `config/ca` and cannot be rotated without remounting the engine. This is a Vault limitation, not an operator limitation.

### Vault CLI Equivalent

```shell
vault write [namespace/]ssh/config/ca \
    generate_signing_key=true \
    key_type=ssh-rsa
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/config/ca` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| generateSigningKey | bool | No | If `true`, Vault generates the SSH CA key pair internally. Default: `true` |
| keyType | string | No | Desired key type for the generated SSH CA key (e.g., `ssh-rsa`, `ecdsa-sha2-nistp256`). Default: `ssh-rsa` |
| keyBits | int | No | Desired key bits for the generated SSH CA key. `0` uses the default for the key type |
| caKeyReference | object | No | Specifies how to retrieve an externally-managed SSH CA private key. Only needed when `generateSigningKey` is `false`. See [Credential Resolution](#credential-resolution) |

## SSHSecretEngineRole

The `SSHSecretEngineRole` CRD allows you to create an [SSH secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/ssh#create-role).

### Example (CA type)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SSHSecretEngineRole
metadata:
  name: ssh-user-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: ssh
  keyType: ca
  defaultUser: ubuntu
  allowedUsers: "*"
  allowUserCertificates: true
  defaultExtensions:
    permit-pty: ""
  ttl: "30m"
  maxTTL: "4h"
```

### Example (OTP type)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SSHSecretEngineRole
metadata:
  name: ssh-otp-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: ssh
  keyType: otp
  defaultUser: centos
  cidrList: "10.0.0.0/8"
```

### Vault CLI Equivalent (CA type)

```shell
vault write [namespace/]ssh/roles/ssh-user-role \
    key_type=ca \
    default_user=ubuntu \
    allowed_users="*" \
    allow_user_certificates=true \
    default_extensions='{"permit-pty": ""}' \
    ttl=30m \
    max_ttl=4h
```

### Vault CLI Equivalent (OTP type)

```shell
vault write [namespace/]ssh/roles/ssh-otp-role \
    key_type=otp \
    default_user=centos \
    cidr_list="10.0.0.0/8"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| keyType | string | Yes | Type of credentials generated: `otp` or `ca` |
| defaultUser | string | No | Default username for SSH sessions. Required for OTP type |
| defaultUserTemplate | bool | No | If `true`, `defaultUser` can contain identity templates |
| allowedUsers | string | No | Comma-separated list of allowed users, or `"*"` for any user |
| allowedUsersTemplate | bool | No | If `true`, `allowedUsers` can contain identity templates |
| ttl | string | No | TTL for credentials (duration format, e.g., `30m`) |
| maxTTL | string | No | Maximum TTL for credentials (duration format) |
| port | int | No | SSH port number. Default: `22` |

**OTP-only fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| cidrList | string | Conditional | Comma-separated CIDR blocks for OTP access. Required for OTP unless zero-address roles are configured |
| excludeCidrList | string | No | Comma-separated CIDR blocks to exclude from OTP access |

**CA-only fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| allowedDomains | string | No | Comma-separated domains for host certificates |
| allowedDomainsTemplate | bool | No | If `true`, `allowedDomains` can contain identity templates |
| allowUserCertificates | bool | No | If `true`, certificates can be signed for user use |
| allowHostCertificates | bool | No | If `true`, certificates can be signed for host use |
| allowBareDomains | bool | No | If `true`, host certs can use base domains from `allowedDomains` |
| allowSubdomains | bool | No | If `true`, host certs can use subdomains of `allowedDomains` |
| allowUserKeyIDs | bool | No | If `true`, users can override the key ID |
| keyIDFormat | string | No | Custom format string for the key ID of signed certificates |
| allowedUserKeyLengths | map[string]int | No | Map of SSH key types to allowed key lengths |
| allowedCriticalOptions | string | No | Comma-separated list of allowed critical options |
| allowedExtensions | string | No | Comma-separated list of allowed extensions, or `"*"` |
| defaultCriticalOptions | map[string]string | No | Map of default critical options |
| defaultExtensions | map[string]string | No | Map of default extensions |
| defaultExtensionsTemplate | bool | No | If `true`, `defaultExtensions` can contain identity templates |
| allowEmptyPrincipals | bool | No | If `true`, allow signing certs with no valid principals |
| algorithmSigner | string | No | Algorithm to sign keys with: `ssh-rsa`, `rsa-sha2-256`, `rsa-sha2-512`, or `default` |
| notBeforeDuration | string | No | Duration to backdate the `ValidAfter` property (duration format) |

## Credential Resolution

When `generateSigningKey` is `false`, an externally-managed CA key pair must be provided via the `caKeyReference` field. The key material can be retrieved from a Kubernetes Secret or a Vault Secret (RandomSecret is not allowed for SSH CA keys).

### Using a Kubernetes Secret

Specify the `secret` field under `caKeyReference`. The secret must contain the SSH CA private key and public key.

```yaml
spec:
  generateSigningKey: false
  caKeyReference:
    secret:
      name: ssh-ca-keypair
    passwordKey: private_key
    usernameKey: public_key
```

### Using a Vault Secret

Specify the `vaultSecret` field under `caKeyReference` to retrieve the CA key pair from a Vault path.

```yaml
spec:
  generateSigningKey: false
  caKeyReference:
    vaultSecret:
      path: secret/ssh-ca-keys
    passwordKey: private_key
    usernameKey: public_key
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault SSH Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/ssh) — Vault documentation
- [Vault SSH Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/ssh) — Vault API reference
