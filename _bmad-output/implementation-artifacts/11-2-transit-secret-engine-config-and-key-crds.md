# Story 11.2: Transit Secret Engine — Config and Key CRDs

baseline_commit: 042adfc0e7e5e819d85e82fc5ec8bd41513cdd54
Status: review

## Story

As an operator developer,
I want CRDs for TransitSecretEngineKey (encryption key lifecycle),
so that Vault's Transit encryption-as-a-service can be managed declaratively.

## Acceptance Criteria

1. **Given** a TransitSecretEngineKey CR is created with key type and configuration
   **When** the reconciler processes it
   **Then** the key exists in Vault at `{path}/keys/{name}` and the CR has condition `ReconcileSuccessful=True`

2. **Given** the key spec is updated (e.g., `minDecryptionVersion`, `deletionAllowed`)
   **When** the reconciler processes the update
   **Then** the key configuration is updated in Vault via `{path}/keys/{name}/config`

3. **Given** the CR is deleted
   **When** the reconciler processes deletion
   **Then** the key is deleted from Vault (only if `deletion_allowed=true` was set in config)

## Tasks / Subtasks

- [x] Task 1: Define CRD type (AC: #1, #2, #3)
  - [x] Create `api/v1alpha1/transitsecretenginekey_types.go`
  - [x] Define `TransitSecretEngineKeySpec` with `Connection`, `Authentication`, `Path`, inline `TransitKeyConfig`, `Name`
  - [x] Define `TransitKeyConfig` struct with create-time and config-time fields
  - [x] Implement `VaultObject` interface: `GetPath()`, `GetPayload()`, `IsEquivalentToDesiredState()`, etc.
  - [x] Implement a `TransitKeyVaultObject` interface with `GetConfigPath()` and `GetConfigPayload()`
  - [x] Implement `ConditionsAware` interface on status
  - [x] Register type in `init()`

- [x] Task 2: Implement reconciliation endpoint (AC: #1, #2)
  - [x] Create `api/v1alpha1/utils/vaulttransitkeyobject.go` with `VaultTransitKeyEndpoint`
  - [x] Implement custom `CreateOrUpdate`: read from `GetPath()`, if not found create at `GetPath()`, if found and config differs write to `GetConfigPath()`
  - [x] Implement `DeleteIfExists` using standard `VaultEndpoint.DeleteIfExists` pattern

- [x] Task 3: Implement controller (AC: #1, #2, #3)
  - [x] Create `internal/controller/transitsecretenginekey_controller.go`
  - [x] Embed `ReconcilerBase`, standard reconcile flow using `ReconcileWithFunctions`
  - [x] Use custom `VaultTransitKeyEndpoint` instead of standard `VaultResource`
  - [x] Add RBAC markers
  - [x] Implement `SetupWithManager`

- [x] Task 4: Implement webhook (AC: #1, #2, #3)
  - [x] Create `api/v1alpha1/transitsecretenginekey_webhook.go`
  - [x] Implement `admission.Defaulter[*TransitSecretEngineKey]` and `admission.Validator[*TransitSecretEngineKey]`
  - [x] `ValidateUpdate`: reject `spec.path` changes (immutable) and reject changes to create-time-only fields (`type`, `derived`, `convergentEncryption`)
  - [x] Kubebuilder webhook markers for mutate and validate paths

- [x] Task 5: Register in main.go (AC: #1)
  - [x] Add controller registration in `main.go`
  - [x] Add webhook registration in `ENABLE_WEBHOOKS` block

- [x] Task 6: Unit tests for toMap/IsEquivalentToDesiredState (AC: #1, #2)
  - [x] Create `api/v1alpha1/transitsecretenginekey_types_test.go`
  - [x] Test `toMap()` produces correct Vault payload (snake_case keys)
  - [x] Test `IsEquivalentToDesiredState()` with Vault-read-shaped payload (only config fields)
  - [x] Negative test: full key metadata (including `keys`, `name`, `supports_*`) returns true (extra fields filtered)
  - [x] Negative test: mismatched config field returns false

- [x] Task 7: Integration tests (AC: #1, #2, #3)
  - [x] Create `internal/controller/transitsecretenginekey_controller_test.go` with `//go:build integration`
  - [x] Create test YAML fixtures in `test/transit/`
  - [x] Test create: CR → key exists in Vault → ReconcileSuccessful=True
  - [x] Test update: change `minDecryptionVersion` → config updated in Vault
  - [x] Test delete: set `deletionAllowed=true`, delete CR → key removed from Vault
  - [x] Register controller in integration test suite

- [x] Task 8: Run code generation and verify (AC: all)
  - [x] `make manifests generate fmt vet test`
  - [x] Verify CRD generated in `config/crd/bases/`
  - [x] Verify RBAC generated
  - [x] Verify deepcopy generated

## Dev Notes

### Vault Transit Key API — Dual-Path Architecture

Transit keys use **two distinct API paths** for lifecycle management:

| Operation | Method | Path | Payload |
|-----------|--------|------|---------|
| Create | POST | `{mount}/keys/{name}` | create-time fields only |
| Read | GET | `{mount}/keys/{name}` | — (returns full metadata) |
| Update Config | POST | `{mount}/keys/{name}/config` | config fields only |
| Delete | DELETE | `{mount}/keys/{name}` | — (requires deletion_allowed) |

**Create-time fields** (immutable after creation):
- `type` (string, default: "aes256-gcm96") — key algorithm
- `derived` (bool, default: false) — enable key derivation
- `convergent_encryption` (bool, default: false) — requires derived=true
- `exportable` (bool, default: false) — can be set true later via config, never unset
- `allow_plaintext_backup` (bool, default: false) — can be set true later via config, never unset
- `key_size` (int, default: 0) — only for HMAC keys (32-512 bytes)
- `auto_rotate_period` (duration, default: "0") — mutable via config

**Config-time fields** (mutable via `/keys/{name}/config`):
- `min_decryption_version` (int, default: 0)
- `min_encryption_version` (int, default: 0)
- `deletion_allowed` (bool, default: false)
- `exportable` (bool, default: false) — one-way: can set true, never unset
- `allow_plaintext_backup` (bool, default: false) — one-way: can set true, never unset
- `auto_rotate_period` (duration, default: "") — "0" disables, minimum 1h

**Read response** returns ALL metadata including computed fields not managed by operator:
```json
{
  "type": "aes256-gcm96",
  "deletion_allowed": false,
  "derived": false,
  "exportable": false,
  "allow_plaintext_backup": false,
  "keys": {"1": 1442851412},
  "min_decryption_version": 1,
  "min_encryption_version": 0,
  "name": "foo",
  "supports_encryption": true,
  "supports_decryption": true,
  "supports_derivation": true,
  "supports_signing": false,
  "imported": false
}
```

### Reconciliation Strategy — Custom VaultTransitKeyEndpoint

The standard `VaultEndpoint.CreateOrUpdate()` always writes to `GetPath()`. For Transit, **updates must go to a different path** (`/config`). Follow the pattern established by `RabbitMQEngineConfigVaultEndpoint` and `VaultPKIEngineEndpoint`:

1. Create `api/v1alpha1/utils/vaulttransitkeyobject.go`:
   - Define `VaultTransitKeyObject` interface extending `VaultObject` with `GetConfigPath() string` and `GetConfigPayload() map[string]any`
   - Define `VaultTransitKeyEndpoint` struct holding the interface
   - `CreateOrUpdate(ctx)`:
     - Read from `GetPath()` (i.e., `{path}/keys/{name}`)
     - If not found → `write(ctx, GetPath(), GetPayload())` (creates key)
     - If found → compare `IsEquivalentToDesiredState(currentPayload)`; if not equivalent → `write(ctx, GetConfigPath(), GetConfigPayload())` (updates config only)
   - `DeleteIfExists(ctx)` — standard: `vaultClient.Logical().Delete(GetPath())`

2. In the controller, create the custom endpoint and pass to `ReconcileWithFunctions`:
   ```go
   endpoint := vaultutils.NewVaultTransitKeyEndpoint(instance)
   return vaultresourcecontroller.ReconcileWithFunctions(ctx1, r.ReconcilerBase, instance,
       endpoint.DeleteIfExists,
       func(ctx context.Context, obj client.Object) error {
           if err := obj.(vaultutils.VaultObject).PrepareInternalValues(ctx, obj); err != nil { return err }
           if err := obj.(vaultutils.VaultObject).PrepareTLSConfig(ctx, obj); err != nil { return err }
           return endpoint.CreateOrUpdate(ctx)
       },
   )
   ```

### IsEquivalentToDesiredState — Config Fields Only

`IsEquivalentToDesiredState` must compare ONLY the config-level mutable fields from the Vault read response. The read response includes many non-managed fields (`keys`, `name`, `supports_*`, `imported`, `type`, `derived`). Use `filterPayloadToDesiredKeys` with a desired state map containing only config fields:

```go
func (d *TransitSecretEngineKey) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.TransitKeyConfig.configToMap()
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

The `configToMap()` method returns only the mutable config fields: `min_decryption_version`, `min_encryption_version`, `deletion_allowed`, `exportable`, `allow_plaintext_backup`, `auto_rotate_period`.

### CRD Type Design

```go
type TransitSecretEngineKeySpec struct {
    Connection     *vaultutils.VaultConnection        `json:"connection,omitempty"`
    Authentication vaultutils.KubeAuthConfiguration   `json:"authentication"`
    Path           vaultutils.Path                    `json:"path"`
    TransitKeyConfig `json:",inline"`
    Name           string                             `json:"name,omitempty"`
}

type TransitKeyConfig struct {
    // Create-time fields (immutable after key creation)
    Type                  string `json:"type" kubebuilder:"default=aes256-gcm96"`
    Derived               bool   `json:"derived,omitempty"`
    ConvergentEncryption  bool   `json:"convergentEncryption,omitempty"`
    KeySize               int    `json:"keySize,omitempty"`

    // Config-time fields (mutable via keys/{name}/config)
    MinDecryptionVersion  int    `json:"minDecryptionVersion,omitempty"`
    MinEncryptionVersion  int    `json:"minEncryptionVersion,omitempty"`
    DeletionAllowed       bool   `json:"deletionAllowed,omitempty"`
    Exportable            bool   `json:"exportable,omitempty"`
    AllowPlaintextBackup  bool   `json:"allowPlaintextBackup,omitempty"`
    AutoRotatePeriod      string `json:"autoRotatePeriod,omitempty"`
}
```

Two `toMap` methods needed:
- `toMap()` → create payload: `type`, `derived`, `convergent_encryption`, `exportable`, `allow_plaintext_backup`, `key_size`, `auto_rotate_period`
- `configToMap()` → config payload: `min_decryption_version`, `min_encryption_version`, `deletion_allowed`, `exportable`, `allow_plaintext_backup`, `auto_rotate_period`

### GetPath and GetConfigPath

```go
func (d *TransitSecretEngineKey) GetPath() string {
    name := d.Name
    if d.Spec.Name != "" { name = d.Spec.Name }
    return vaultutils.CleansePath(string(d.Spec.Path) + "/keys/" + name)
}

// GetConfigPath returns the config endpoint for mutable field updates
func (d *TransitSecretEngineKey) GetConfigPath() string {
    return d.GetPath() + "/config"
}
```

### Webhook Validation

`ValidateUpdate` must reject:
1. `spec.path` changes (standard immutable path rule)
2. `spec.type` changes (immutable after key creation)
3. `spec.derived` changes (immutable)
4. `spec.convergentEncryption` changes (immutable)
5. `spec.keySize` changes (immutable)

### IsDeletable Behavior

`IsDeletable()` should return `true` — the operator should always attempt deletion. If `deletion_allowed` is `false` in Vault, the DELETE call will fail with a Vault error, which is the correct behavior (Vault enforces this constraint, not the operator). The operator logs the error and retries.

**Alternative considered:** conditionally return `d.Spec.DeletionAllowed`. Rejected because:
- The spec field could be stale (user might have set it true in Vault directly)
- The standard pattern is to let Vault enforce its own constraints
- The finalizer should still be added so cleanup is attempted

### Integration Test Strategy

Transit keys can run entirely within Vault — no external service dependency. Classification: **Install in Kind** (Vault is already available in the integration test infrastructure).

Test lifecycle:
1. Create: Deploy TransitSecretEngineKey CR → verify `ReconcileSuccessful=True` → verify key exists via `vault read transit/keys/{name}`
2. Update: Patch `minDecryptionVersion` → verify config updated via Vault read
3. Delete: Set `deletionAllowed=true` first, then delete CR → verify key gone from Vault

Fixture path: `test/transit/transit-secret-engine-key.yaml`

### Kubebuilder Markers for Validation

```go
// Type field: non-zero default, no omitempty
// +kubebuilder:default="aes256-gcm96"
// +kubebuilder:validation:Enum={"aes128-gcm96","aes256-gcm96","chacha20-poly1305","ed25519","ecdsa-p256","ecdsa-p384","ecdsa-p521","rsa-2048","rsa-3072","rsa-4096","hmac"}

// MinDecryptionVersion: zero-value default, omitempty
// +kubebuilder:validation:Minimum=0

// MinEncryptionVersion: zero-value default, omitempty
// +kubebuilder:validation:Minimum=0

// KeySize: zero-value default, omitempty
// +kubebuilder:validation:Minimum=0
// +kubebuilder:validation:Maximum=512
```

Note: Enterprise-only key types (`aes128-cmac`, `aes256-cmac`, `ml-dsa`, `hybrid`, `slh-dsa`, `aes128-cbc`, `aes256-cbc`) are excluded from the enum. They can be added later if needed.

### Project Structure Notes

| Artifact | Path |
|----------|------|
| Types | `api/v1alpha1/transitsecretenginekey_types.go` |
| Webhook | `api/v1alpha1/transitsecretenginekey_webhook.go` |
| Controller | `internal/controller/transitsecretenginekey_controller.go` |
| Endpoint helper | `api/v1alpha1/utils/vaulttransitkeyobject.go` |
| Unit tests | `api/v1alpha1/transitsecretenginekey_types_test.go` |
| Integration tests | `internal/controller/transitsecretenginekey_controller_test.go` |
| Test fixtures | `test/transit/transit-secret-engine-key.yaml` |
| CRD output | `config/crd/bases/redhatcop.redhat.io_transitsecretenginekeys.yaml` |
| Main registration | `cmd/main.go` |

### References

- [Vault Transit API](https://developer.hashicorp.com/vault/api-docs/secret/transit) — Create Key, Update Key Configuration, Read Key, Delete Key endpoints
- [Source: project-context.md] — VaultObject interface, ConditionsAware, CRD type structure, controller patterns, webhook rules, testing standards
- [Source: api/v1alpha1/utils/vaultobject.go] — VaultObject interface, VaultEndpoint.CreateOrUpdate pattern, filterPayloadToDesiredKeys, RabbitMQEngineConfigVaultEndpoint dual-path pattern
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go] — Secret-stripping in IsEquivalentToDesiredState, credential resolution pattern
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go] — Standard VaultObject implementation, GetPath with Name override
- [Source: internal/controller/vaultresourcecontroller/vaultresourcereconciler.go] — VaultResource reconciler, manageReconcileLogic, how to use ReconcileWithFunctions
- [Source: internal/controller/vaultresourcecontroller/reconcile_skeleton.go] — ReconcileWithFunctions signature, deletion/finalizer management
- [Source: api/v1alpha1/kubernetessecretengineconfig_webhook.go] — Webhook pattern with immutable path, validator/defaulter interfaces

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

### Review Findings

- [ ] [Review][Patch] `spec.name` remains mutable and can orphan Vault keys [api/v1alpha1/transitsecretenginekey_webhook.go:58]
- [ ] [Review][Patch] Transit key equivalence compares `int` desired values against Vault numeric payload types and can re-write `/config` forever under drift detection [api/v1alpha1/transitsecretenginekey_types.go:177]
- [ ] [Review][Patch] One-way Transit flags can be changed from `true` back to `false` even though the story constrains them as irreversible [api/v1alpha1/transitsecretenginekey_webhook.go:58]
- [ ] [Review][Patch] Transit-specific create/update invariants are documented but not enforced by schema or validating webhook [api/v1alpha1/transitsecretenginekey_webhook.go:47]
- [ ] [Review][Patch] Shared Kind defaults were changed to worktree-local ports instead of using per-run overrides [integration/cluster-kind.yaml:13]
