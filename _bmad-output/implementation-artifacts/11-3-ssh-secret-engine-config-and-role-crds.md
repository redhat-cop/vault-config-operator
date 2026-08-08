# Story 11.3: SSH Secret Engine — Config and Role CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for SSHSecretEngineConfig and SSHSecretEngineRole,
So that Vault's SSH secret engine (signed keys and OTP) can be managed declaratively.

## Acceptance Criteria

1. **Given** an SSHSecretEngineConfig CR is created with CA key configuration
   **When** the reconciler processes it
   **Then** the SSH CA is configured in Vault and ReconcileSuccessful=True

2. **Given** an SSHSecretEngineRole CR is created with key_type=ca
   **When** the reconciler processes it
   **Then** the role exists in Vault with CA-specific fields and ReconcileSuccessful=True

3. **Given** an SSHSecretEngineRole CR is created with key_type=otp
   **When** the reconciler processes it
   **Then** the role exists in Vault with OTP-specific fields and ReconcileSuccessful=True

4. **Given** both CRs are deleted
   **When** the reconcilers process deletions
   **Then** Vault resources are cleaned up (config/ca deleted, role deleted)

5. **Given** an SSHSecretEngineConfig CR specifies private_key via a K8s Secret reference
   **When** the reconciler processes it
   **Then** PrepareInternalValues resolves the private key from the K8s Secret

6. **Given** spec.path is updated on either CR
   **When** the webhook validates the update
   **Then** the update is rejected with "spec.path cannot be updated"

## Tasks / Subtasks

- [ ] Task 1: Define SSHSecretEngineConfig CRD type (AC: #1, #5)
  - [ ] Create `api/v1alpha1/sshsecretengineconfig_types.go`
  - [ ] Define `SSHSEConfig` inline struct with fields: `generateSigningKey`, `keyType`, `keyBits`
  - [ ] Add `RootCredentialConfig` field for private_key/public_key credential resolution
  - [ ] Implement `VaultObject` interface: `GetPath()` → `{path}/config/ca`
  - [ ] Implement `toMap()` converting CRD fields to Vault API snake_case keys
  - [ ] Implement `IsEquivalentToDesiredState()` — must delete `private_key` from desiredState (Vault never returns it on read)
  - [ ] Implement `PrepareInternalValues()` — resolve private_key + public_key from K8s Secret or VaultSecret
  - [ ] Implement `ConditionsAware` interface
  - [ ] Implement `IsValid()` with credential source validation
  - [ ] `IsDeletable()` returns `true`
  - [ ] Register type in `init()`

- [ ] Task 2: Define SSHSecretEngineRole CRD type (AC: #2, #3)
  - [ ] Create `api/v1alpha1/sshsecretenginerole_types.go`
  - [ ] Define `SSHSERole` inline struct with all role fields (shared + CA-specific + OTP-specific)
  - [ ] Implement `VaultObject` interface: `GetPath()` → `{path}/roles/{name}`
  - [ ] Implement `toMap()` — must conditionally include fields based on `keyType` value
  - [ ] Implement `IsEquivalentToDesiredState()` using `filterPayloadToDesiredKeys`
  - [ ] Implement `ConditionsAware` interface
  - [ ] `IsDeletable()` returns `true`
  - [ ] Register type in `init()`
  - [ ] Add `Name` field with pattern validation for Vault object name override

- [ ] Task 3: Create webhooks (AC: #6)
  - [ ] Create `api/v1alpha1/sshsecretengineconfig_webhook.go`
  - [ ] Create `api/v1alpha1/sshsecretenginerole_webhook.go`
  - [ ] Implement `admission.Defaulter[*T]` and `admission.Validator[*T]` for both
  - [ ] `ValidateUpdate` must reject `spec.path` changes
  - [ ] Config webhook: `ValidateCreate`/`ValidateUpdate` call credential validation
  - [ ] Role webhook: validate `keyType` is "ca" or "otp"

- [ ] Task 4: Create controllers (AC: #1, #2, #3, #4)
  - [ ] Create `internal/controller/sshsecretengineconfig_controller.go`
  - [ ] Create `internal/controller/sshsecretenginerole_controller.go`
  - [ ] Both embed `vaultresourcecontroller.ReconcilerBase`
  - [ ] Config controller: Watch K8s Secrets for credential changes (same pattern as KubernetesSecretEngineConfig)
  - [ ] Role controller: simple `For()` with default periodic reconcile predicate
  - [ ] Add RBAC markers

- [ ] Task 5: Register in main.go (AC: all)
  - [ ] Register both controllers (outside webhook guard)
  - [ ] Register both webhooks (inside `ENABLE_WEBHOOKS` guard)

- [ ] Task 6: Unit tests for toMap and IsEquivalentToDesiredState (AC: #1, #2, #3)
  - [ ] Create `api/v1alpha1/sshsecretengineconfig_test.go`
  - [ ] Create `api/v1alpha1/sshsecretenginerole_test.go`
  - [ ] Config: verify toMap output, verify IsEquivalentToDesiredState ignores private_key
  - [ ] Role CA: verify toMap includes CA-specific fields, verify IsEquivalentToDesiredState
  - [ ] Role OTP: verify toMap includes OTP-specific fields (cidr_list, etc.)
  - [ ] Negative tests: verify field mismatch returns false

- [ ] Task 7: Integration tests (AC: #1, #2, #3, #4)
  - [ ] Create test fixtures in `test/ssh/`
  - [ ] Create `internal/controller/sshsecretengineconfig_controller_test.go` (integration)
  - [ ] Create `internal/controller/sshsecretenginerole_controller_test.go` (integration)
  - [ ] Config test: create with `generate_signing_key=true`, verify reconcile success, delete
  - [ ] Role CA test: create role with key_type=ca, verify reconcile success, delete
  - [ ] Role OTP test: create role with key_type=otp, verify reconcile success, delete
  - [ ] Register controllers in `suite_integration_test.go`

- [ ] Task 8: Run code generation and verify (AC: all)
  - [ ] `make manifests generate fmt vet`
  - [ ] `make test` — unit tests pass
  - [ ] Verify generated CRDs in `config/crd/bases/`

## Dev Notes

### Vault API Reference

**Config CA endpoint:** `POST {path}/config/ca`
- `private_key` (string) — private key part of SSH CA key pair; required if generate_signing_key=false
- `public_key` (string) — public key part of SSH CA key pair; required if generate_signing_key=false
- `generate_signing_key` (bool: true) — if true, Vault generates the key pair internally
- `key_type` (string: "ssh-rsa") — desired key type when generating (ssh-rsa, ecdsa-sha2-nistp256, etc.)
- `key_bits` (int: 0) — key bits when generating (0 = default: 4096 for RSA, P-256 for EC)

**Important:** Vault returns the `public_key` on read but NEVER returns `private_key`. `IsEquivalentToDesiredState` must delete `private_key` from desiredState before comparison (same pattern as KubernetesSecretEngineConfig deletes `service_account_jwt`).

**Config CA DELETE:** `DELETE {path}/config/ca` — deletes the CA information. `IsDeletable()` = true.

**Roles endpoint:** `POST/GET/DELETE {path}/roles/{name}`

**Role fields — shared (both CA and OTP):**
- `key_type` (string) — "ca" or "otp" (REQUIRED)
- `default_user` (string) — default username (required for OTP)
- `default_user_template` (bool: false)
- `allowed_users` (string) — comma-separated list or "*"
- `allowed_users_template` (bool: false)
- `ttl` (string) — TTL duration
- `max_ttl` (string) — max TTL duration
- `port` (int: 22) — SSH port number

**Role fields — OTP only (not applicable for CA):**
- `cidr_list` (string) — REQUIRED for OTP unless zero-address configured
- `exclude_cidr_list` (string)

**Role fields — CA only (not applicable for OTP):**
- `allowed_domains` (string) — comma-separated
- `allowed_domains_template` (bool: false)
- `allow_user_certificates` (bool: false)
- `allow_host_certificates` (bool: false)
- `allow_bare_domains` (bool: false)
- `allow_subdomains` (bool: false)
- `allow_user_key_ids` (bool: false)
- `key_id_format` (string)
- `allowed_user_key_lengths` (map) — map of key type to allowed lengths
- `allowed_critical_options` (string) — comma-separated
- `allowed_extensions` (string) — comma-separated, or "*"
- `default_critical_options` (map<string|string>)
- `default_extensions` (map<string|string>)
- `default_extensions_template` (bool: false)
- `allow_empty_principals` (bool: false)
- `algorithm_signer` (string: "default") — ssh-rsa, rsa-sha2-256, rsa-sha2-512, or default
- `not_before_duration` (duration: "30s")

**Vault GET response for roles:** Returns different field sets based on key_type. OTP response example: `{"cidr_list":"x.x.x.x/y","default_user":"username","key_type":"otp","port":22}`. CA response example: `{"allow_bare_domains":false,"allow_host_certificates":true,...,"max_ttl":"768h","ttl":"4h"}`.

### Implementation Patterns

**Reference implementations (follow these exactly):**
- `api/v1alpha1/kubernetessecretengineconfig_types.go` — Config with PrepareInternalValues credential resolution pattern
- `api/v1alpha1/kubernetessecretenginerole_types.go` — Role with Name override and toMap
- `api/v1alpha1/kubernetessecretengineconfig_webhook.go` — Webhook with credential validation
- `internal/controller/kubernetessecretengineconfig_controller.go` — Controller with Secret watching
- `internal/controller/kubernetessecretenginerole_controller.go` — Simple role controller

**SSHSecretEngineConfig specific patterns:**
- `GetPath()` → `string(d.Spec.Path) + "/" + "config/ca"`
- `GetPayload()` → `d.Spec.toMap()` (includes resolved credentials)
- `IsEquivalentToDesiredState()` → delete `private_key` from desiredState (Vault never returns it), then use `filterPayloadToDesiredKeys`
- `PrepareInternalValues()` → resolve private_key and public_key from K8s Secret (`RootCredentialConfig`). The Secret should contain keys like `private_key` and `public_key`. Only `private_key` resolution is critical; `public_key` can also be provided inline in spec.
- `IsDeletable()` → `true` (config/ca supports DELETE)

**SSHSecretEngineRole specific patterns:**
- `GetPath()` → `CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + name)` (use Spec.Name override or metadata.name)
- `toMap()` must include ALL fields regardless of key_type — Vault accepts extra fields gracefully and only returns relevant ones on read. The `filterPayloadToDesiredKeys` comparison handles the difference.
- `IsEquivalentToDesiredState()` → standard `filterPayloadToDesiredKeys` pattern

**Conditional toMap approach:** Include all fields in toMap() output. Use zero-value omission where appropriate. The `filterPayloadToDesiredKeys` mechanism ensures comparison only checks keys present in desiredState against what Vault returns. This avoids complex conditional logic in toMap while still working correctly because:
1. Vault ignores irrelevant fields on write (OTP fields ignored for CA roles)
2. Vault only returns relevant fields on read
3. `filterPayloadToDesiredKeys` filters read response to only keys in desiredState

**Alternative (safer) approach for toMap:** Only emit fields relevant to the key_type. This prevents sending noise to Vault and makes the CR spec clearer. To do this, check `i.KeyType` in toMap and conditionally include CA-only or OTP-only fields. The downside is more complex code and potential drift if Vault adds shared fields later.

**Recommended:** Use the conditional approach (only emit relevant fields based on key_type) since the Vault API docs explicitly mark fields as "[Not applicable for CA type]" — this keeps the Vault write payload clean.

### SSHSEConfig Struct Design

```go
type SSHSEConfig struct {
    // GenerateSigningKey if true, Vault generates the SSH CA key pair internally.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=true
    GenerateSigningKey bool `json:"generateSigningKey"`

    // KeyType desired key type for generated SSH CA key (ssh-rsa, ecdsa-sha2-nistp256, etc.)
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="ssh-rsa"
    KeyType string `json:"keyType"`

    // KeyBits desired key bits for generated SSH CA key (0 = default)
    // +kubebuilder:validation:Optional
    KeyBits int `json:"keyBits,omitempty"`

    retrievedPrivateKey string `json:"-"`
    retrievedPublicKey  string `json:"-"`
}
```

### SSHSecretEngineConfig Spec Design

```go
type SSHSecretEngineConfigSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which to make the configuration.
    // The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config/ca.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    // CAKeyReference specifies how to retrieve the SSH CA private key.
    // Only needed when generateSigningKey is false.
    // Only VaultSecretReference or LocalObjectReference can be used.
    // +kubebuilder:validation:Optional
    CAKeyReference *vaultutils.RootCredentialConfig `json:"caKeyReference,omitempty"`

    SSHSEConfig `json:",inline"`
}
```

**Key design decision:** `CAKeyReference` is a pointer (optional) because when `generateSigningKey=true`, no external credential is needed. The `IsValid()` method must enforce: if `generateSigningKey=false`, then `CAKeyReference` must be non-nil and valid.

### SSHSERole Struct Design

```go
type SSHSERole struct {
    // KeyType specifies the type of credentials generated. Must be "otp" or "ca".
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Enum={"otp","ca"}
    KeyType string `json:"keyType"`

    // DefaultUser specifies the default username. Required for OTP type.
    // +kubebuilder:validation:Optional
    DefaultUser string `json:"defaultUser,omitempty"`

    // DefaultUserTemplate if set, default_user can contain identity templates.
    // +kubebuilder:validation:Optional
    DefaultUserTemplate bool `json:"defaultUserTemplate,omitempty"`

    // AllowedUsers comma-separated list or "*" for any user.
    // +kubebuilder:validation:Optional
    AllowedUsers string `json:"allowedUsers,omitempty"`

    // AllowedUsersTemplate if set, allowed_users can contain identity templates.
    // +kubebuilder:validation:Optional
    AllowedUsersTemplate bool `json:"allowedUsersTemplate,omitempty"`

    // TTL for credentials. Uses duration format strings.
    // +kubebuilder:validation:Optional
    TTL string `json:"ttl,omitempty"`

    // MaxTTL maximum TTL for credentials.
    // +kubebuilder:validation:Optional
    MaxTTL string `json:"maxTTL,omitempty"`

    // Port SSH port number.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=22
    Port int `json:"port"`

    // --- OTP-only fields ---

    // CIDRList comma-separated CIDR blocks. Required for OTP unless zero-address.
    // +kubebuilder:validation:Optional
    CIDRList string `json:"cidrList,omitempty"`

    // ExcludeCIDRList comma-separated CIDR blocks to exclude.
    // +kubebuilder:validation:Optional
    ExcludeCIDRList string `json:"excludeCidrList,omitempty"`

    // --- CA-only fields ---

    // AllowedDomains comma-separated domains for host certificates.
    // +kubebuilder:validation:Optional
    AllowedDomains string `json:"allowedDomains,omitempty"`

    // AllowedDomainsTemplate if set, allowed_domains can contain identity templates.
    // +kubebuilder:validation:Optional
    AllowedDomainsTemplate bool `json:"allowedDomainsTemplate,omitempty"`

    // AllowUserCertificates if true, certificates can be signed for user use.
    // +kubebuilder:validation:Optional
    AllowUserCertificates bool `json:"allowUserCertificates,omitempty"`

    // AllowHostCertificates if true, certificates can be signed for host use.
    // +kubebuilder:validation:Optional
    AllowHostCertificates bool `json:"allowHostCertificates,omitempty"`

    // AllowBareDomains if true, host certs can use base domains from allowed_domains.
    // +kubebuilder:validation:Optional
    AllowBareDomains bool `json:"allowBareDomains,omitempty"`

    // AllowSubdomains if true, host certs can use subdomains of allowed_domains.
    // +kubebuilder:validation:Optional
    AllowSubdomains bool `json:"allowSubdomains,omitempty"`

    // AllowUserKeyIDs if true, users can override the key ID.
    // +kubebuilder:validation:Optional
    AllowUserKeyIDs bool `json:"allowUserKeyIDs,omitempty"`

    // KeyIDFormat custom format for key id of signed certificate.
    // +kubebuilder:validation:Optional
    KeyIDFormat string `json:"keyIDFormat,omitempty"`

    // AllowedUserKeyLengths map of ssh key types to allowed lengths.
    // +kubebuilder:validation:Optional
    AllowedUserKeyLengths map[string]int `json:"allowedUserKeyLengths,omitempty"`

    // AllowedCriticalOptions comma-separated list of critical options.
    // +kubebuilder:validation:Optional
    AllowedCriticalOptions string `json:"allowedCriticalOptions,omitempty"`

    // AllowedExtensions comma-separated list of extensions or "*".
    // +kubebuilder:validation:Optional
    AllowedExtensions string `json:"allowedExtensions,omitempty"`

    // DefaultCriticalOptions map of default critical options.
    // +kubebuilder:validation:Optional
    DefaultCriticalOptions map[string]string `json:"defaultCriticalOptions,omitempty"`

    // DefaultExtensions map of default extensions.
    // +kubebuilder:validation:Optional
    DefaultExtensions map[string]string `json:"defaultExtensions,omitempty"`

    // DefaultExtensionsTemplate if set, default_extensions can contain identity templates.
    // +kubebuilder:validation:Optional
    DefaultExtensionsTemplate bool `json:"defaultExtensionsTemplate,omitempty"`

    // AllowEmptyPrincipals allow signing certs with no valid principals.
    // +kubebuilder:validation:Optional
    AllowEmptyPrincipals bool `json:"allowEmptyPrincipals,omitempty"`

    // AlgorithmSigner algorithm to sign keys with (ssh-rsa, rsa-sha2-256, rsa-sha2-512, default).
    // +kubebuilder:validation:Optional
    AlgorithmSigner string `json:"algorithmSigner,omitempty"`

    // NotBeforeDuration duration to backdate ValidAfter property.
    // +kubebuilder:validation:Optional
    NotBeforeDuration string `json:"notBeforeDuration,omitempty"`
}
```

### toMap Implementation for SSHSERole (Conditional Approach)

```go
func (i *SSHSERole) toMap() map[string]any {
    payload := map[string]any{}
    payload["key_type"] = i.KeyType
    payload["default_user"] = i.DefaultUser
    payload["default_user_template"] = i.DefaultUserTemplate
    payload["allowed_users"] = i.AllowedUsers
    payload["allowed_users_template"] = i.AllowedUsersTemplate
    payload["ttl"] = i.TTL
    payload["max_ttl"] = i.MaxTTL
    payload["port"] = i.Port

    if i.KeyType == "otp" {
        payload["cidr_list"] = i.CIDRList
        payload["exclude_cidr_list"] = i.ExcludeCIDRList
    }

    if i.KeyType == "ca" {
        payload["allowed_domains"] = i.AllowedDomains
        payload["allowed_domains_template"] = i.AllowedDomainsTemplate
        payload["allow_user_certificates"] = i.AllowUserCertificates
        payload["allow_host_certificates"] = i.AllowHostCertificates
        payload["allow_bare_domains"] = i.AllowBareDomains
        payload["allow_subdomains"] = i.AllowSubdomains
        payload["allow_user_key_ids"] = i.AllowUserKeyIDs
        payload["key_id_format"] = i.KeyIDFormat
        payload["allowed_user_key_lengths"] = i.AllowedUserKeyLengths
        payload["allowed_critical_options"] = i.AllowedCriticalOptions
        payload["allowed_extensions"] = i.AllowedExtensions
        payload["default_critical_options"] = i.DefaultCriticalOptions
        payload["default_extensions"] = i.DefaultExtensions
        payload["default_extensions_template"] = i.DefaultExtensionsTemplate
        payload["allow_empty_principals"] = i.AllowEmptyPrincipals
        payload["algorithm_signer"] = i.AlgorithmSigner
        payload["not_before_duration"] = i.NotBeforeDuration
    }

    return payload
}
```

### Integration Test Strategy

**External dependency classification:** Category 1 — Vault handles SSH CA internally (no external SSH service needed).

**Config test:** Use `generate_signing_key=true` — Vault generates its own CA key pair, no K8s Secret needed for the test.

**Role CA test:** Create a role with `key_type=ca`, `allow_user_certificates=true`, `allowed_users="*"`. Verify reconcile success.

**Role OTP test:** Create a role with `key_type=otp`, `default_user="ubuntu"`, `cidr_list="0.0.0.0/0"`. Verify reconcile success.

**Test fixtures location:** `test/ssh/`
- `ssh-secret-engine-config.yaml`
- `ssh-secret-engine-role-ca.yaml`
- `ssh-secret-engine-role-otp.yaml`

**Prerequisite:** The SSH secret engine must be mounted before tests run. Either:
- Add a `SecretEngineMount` fixture for the SSH engine
- Or mount it in a `BeforeAll` block via direct Vault client call

Use the pattern from existing integration tests — mount via a `SecretEngineMount` CR fixture.

### File Structure (New Files)

```
api/v1alpha1/sshsecretengineconfig_types.go      (NEW)
api/v1alpha1/sshsecretengineconfig_webhook.go     (NEW)
api/v1alpha1/sshsecretengineconfig_test.go        (NEW)
api/v1alpha1/sshsecretenginerole_types.go         (NEW)
api/v1alpha1/sshsecretenginerole_webhook.go       (NEW)
api/v1alpha1/sshsecretenginerole_test.go          (NEW)
internal/controller/sshsecretengineconfig_controller.go       (NEW)
internal/controller/sshsecretengineconfig_controller_test.go  (NEW - integration)
internal/controller/sshsecretenginerole_controller.go         (NEW)
internal/controller/sshsecretenginerole_controller_test.go    (NEW - integration)
test/ssh/ssh-secret-engine-config.yaml            (NEW)
test/ssh/ssh-secret-engine-role-ca.yaml           (NEW)
test/ssh/ssh-secret-engine-role-otp.yaml          (NEW)
test/ssh/ssh-secret-engine-mount.yaml             (NEW - prerequisite mount)
```

### Files to Update

```
cmd/main.go                                       (UPDATE - register controllers + webhooks)
internal/controller/suite_integration_test.go     (UPDATE - register SSH controllers)
```

### Critical Gotchas

1. **private_key is never returned by Vault on read** — `IsEquivalentToDesiredState` MUST delete `private_key` from desiredState before comparison. Same pattern as KubernetesSecretEngineConfig deletes `service_account_jwt`.

2. **generate_signing_key=true means no credential resolution needed** — `PrepareInternalValues` should be a no-op when `generateSigningKey=true`. Only resolve credentials when `generateSigningKey=false` AND `CAKeyReference` is set.

3. **Config CA is idempotent on first write but errors on subsequent writes** — Vault returns an error if you try to POST to config/ca when a CA already exists. The operator's `IsEquivalentToDesiredState` check must return true if the CA exists with the expected public_key, preventing re-writes. If the read returns data, the config exists.

4. **Vault SSH config/ca read response:** GET to `{path}/config/ca` returns `{"public_key": "ssh-rsa AAAA..."}` — only the public_key. This means `IsEquivalentToDesiredState` can only compare `public_key` (and `key_type`, `key_bits` if returned). The desiredState must only contain fields that Vault returns on read.

5. **AllowedUserKeyLengths map type:** The Vault API accepts `map<string|(int|[]int|string)>` but for CRD simplicity use `map[string]int` (single length per key type). This covers the common case. If multiple lengths are needed per type, consider `map[string]string` with comma-separated values.

6. **kubebuilder markers for maps:** Use `// +mapType=granular` on map fields (`DefaultCriticalOptions`, `DefaultExtensions`, `AllowedUserKeyLengths`).

7. **Build tags:** Unit tests use `//go:build !integration`, integration tests use `//go:build integration`.

8. **Webhook registration uses lowercase type name:** Path must be `/mutate-redhatcop-redhat-io-v1alpha1-sshsecretengineconfig` and `/validate-redhatcop-redhat-io-v1alpha1-sshsecretengineconfig` (all lowercase, no hyphens in type name).

### Project Structure Notes

- All files follow existing naming conventions exactly
- Controllers in `internal/controller/` (go/v4 layout since Epic 10)
- CRD types and webhooks in `api/v1alpha1/`
- Test fixtures in `test/ssh/` (new directory)
- No conflicts with existing code — this is entirely additive

### References

- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go] — PrepareInternalValues pattern with RootCredentialConfig
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go] — Role type with Name override and toMap
- [Source: api/v1alpha1/kubernetessecretengineconfig_webhook.go] — Webhook with credential validation
- [Source: internal/controller/kubernetessecretengineconfig_controller.go] — Controller with Secret watching
- [Source: internal/controller/kubernetessecretenginerole_controller.go] — Simple role controller
- [Source: api/v1alpha1/payload_filter.go] — filterPayloadToDesiredKeys helper
- [Source: _bmad-output/project-context.md] — Full project rules and patterns
- [Vault API: https://developer.hashicorp.com/vault/api-docs/secret/ssh] — SSH secrets engine API reference

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
