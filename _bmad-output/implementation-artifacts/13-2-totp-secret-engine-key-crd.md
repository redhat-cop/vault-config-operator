---
baseline_commit: 2ccad3a67685ec8bdd0c585247ee148143415e6c
---

# Story 13.2: TOTP Secret Engine — Key CRD

Status: in-progress

## Story

As an operator developer,
I want a CRD for TOTPSecretEngineKey,
So that TOTP key generation and management can be managed declaratively.

## Acceptance Criteria

1. **Given** a TOTPSecretEngineKey CR is created with `generate=true`, issuer, and account name **When** the reconciler processes it **Then** the TOTP key exists in Vault at `{path}/keys/{name}` and ReconcileSuccessful=True

2. **Given** a TOTPSecretEngineKey CR is created with `generate=false` and a `key` or `url` value **When** the reconciler processes it **Then** the TOTP key exists in Vault at `{path}/keys/{name}` using the provided secret and ReconcileSuccessful=True

3. **Given** the TOTPSecretEngineKey CR spec is updated (e.g., issuer, account_name, algorithm, period, digits changed) **When** the reconciler processes the update **Then** the Vault key reflects the updated values

4. **Given** the TOTPSecretEngineKey CR is deleted **When** the reconciler processes deletion **Then** the key is removed from Vault via `DELETE {path}/keys/{name}` and the CR is removed from K8s (`IsDeletable=true`)

5. **Given** a TOTPSecretEngineKey CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates, and `spec.digits` is validated as 6 or 8

6. **Given** the CRD type is implemented **When** the story is marked done **Then** a documentation file exists at `docs/secret-engines/totp.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [x] Task 1: Create `TOTPSecretEngineKey` type (AC: 1, 2, 3, 4)
  - [x] 1.1: Create `api/v1alpha1/totpsecretenginekey_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `TOTPKeyConfig` struct, `Name`
  - [x] 1.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/keys/{name}`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()` (no-op), `IsDeletable()=true`
  - [x] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [x] 1.4: Implement `toMap()` on `TOTPKeyConfig` — convert to Vault API snake_case fields, emit `digits` and `period` as `json.Number`, emit `key_size` and `skew` and `qr_size` as `json.Number`
  - [x] 1.5: Implement `IsEquivalentToDesiredState()` — compare only Vault-read-visible fields (`account_name`, `algorithm`, `digits`, `issuer`, `period`); write-only fields (`key`, `url`, `generate`, `exported`, `key_size`, `qr_size`, `skew`) are never returned by Vault

- [x] Task 2: Create webhook (AC: 5)
  - [x] 2.1: Create `api/v1alpha1/totpsecretenginekey_webhook.go` — `admission.Defaulter[*TOTPSecretEngineKey]`, `admission.Validator[*TOTPSecretEngineKey]`, immutable `spec.path`/`spec.name`, validate `digits` is 6 or 8, validate `algorithm` is SHA1/SHA256/SHA512, validate `skew` is 0 or 1

- [x] Task 3: Create controller (AC: 1, 2, 3, 4)
  - [x] 3.1: Create `internal/controller/totpsecretenginekey_controller.go` — embed `ReconcilerBase`, standard `VaultResource` reconcile flow (no always-write needed), `For()` with default periodic reconcile predicate

- [x] Task 4: Register in main.go (AC: 1, 2)
  - [x] 4.1: Add controller registration for `TOTPSecretEngineKeyReconciler`
  - [x] 4.2: Add webhook registration inside `ENABLE_WEBHOOKS` guard

- [x] Task 5: Unit tests (AC: 1, 2, 3, 5)
  - [x] 5.1: Create `api/v1alpha1/totpsecretenginekey_test.go` — test `toMap()` output with snake_case keys and `json.Number` for numeric fields; test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures; negative tests proving mismatched managed fields return false
  - [x] 5.2: Test `readVisibleMap()` output separately to verify only read-visible fields are compared

- [x] Task 6: Test fixtures and integration tests (AC: 1, 2, 4)
  - [x] 6.1: Create test YAML fixtures in `test/totpsecretengine/` — generate-mode and import-mode key CRs
  - [x] 6.2: Create integration tests in `internal/controller/totpsecretenginekey_controller_test.go` — key create (generate mode)/verify/delete lifecycle

- [x] Task 7: CRD registration and code generation (AC: all)
  - [x] 7.1: Run `make manifests generate fmt vet test`
  - [x] 7.2: Add new CRD YAML file to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [x] 7.3: Verify all existing tests still pass

- [x] Task 8: Documentation (AC: 6)
  - [x] 8.1: Create `docs/secret-engines/totp.md` following `docs/engine-doc-template.md`
  - [x] 8.2: Update `docs/secret-engines/index.md` with link to new doc

### Review Findings

- [ ] [Review][Patch] Generate-mode `skew` default is forced to `0` instead of Vault's documented `1` [`api/v1alpha1/totpsecretenginekey_types.go`]
- [ ] [Review][Patch] `qrSize: 0` cannot be represented because zero is treated as the "use default 200" sentinel [`api/v1alpha1/totpsecretenginekey_types.go`]
- [ ] [Review][Patch] Generate mode does not validate required `issuer` and `accountName` inputs before reconcile [`api/v1alpha1/totpsecretenginekey_webhook.go`]
- [ ] [Review][Patch] AC2 and AC3 lack integration coverage for import mode and update reconciliation paths [`internal/controller/totpsecretenginekey_controller_test.go`]

#### Iteration 2

- [ ] [Review][Patch] Write-only spec fields are accepted on update but never reconciled [`api/v1alpha1/totpsecretenginekey_types.go:192`]
- [ ] [Review][Patch] URL-only import mode can never reach a stable desired state [`api/v1alpha1/totpsecretenginekey_types.go:247`]
- [ ] [Review][Patch] Documentation omits mode-specific required fields for generate vs import flows [`docs/secret-engines/totp.md:76`]

#### Iteration 3

- [ ] [Review][Patch] AC3 update fields are incorrectly immutable [`api/v1alpha1/totpsecretenginekey_webhook.go:93`]
- [ ] [Review][Patch] `skew` and `qrSize` updates are accepted but never reconciled [`api/v1alpha1/totpsecretenginekey_webhook.go:89`]
- [ ] [Review][Patch] Field table still marks `issuer` and `accountName` as not required [`docs/secret-engines/totp.md:76`]

#### Iteration 4

- [ ] [Review][Decision] Define authoritative URL-import behavior for read-visible fields — `spec.url` can embed `issuer`, `account_name`, `algorithm`, `digits`, and `period`, but `readVisibleMap()` compares only the top-level spec fields. Decide whether the operator should parse and normalize those values from `spec.url` or reject URLs whose embedded values conflict with the explicit spec, because the current behavior can leave URL-based imports permanently out of sync.
- [ ] [Review][Patch] Reject incompatible cross-mode fields during validation [`api/v1alpha1/totpsecretenginekey_webhook.go:58`]
- [ ] [Review][Patch] Add coverage for URL import and TOTP webhook invariants [`internal/controller/totpsecretenginekey_controller_test.go:138`]

#### Iteration 5

- [ ] [Review][Patch] Import-mode validation still accepts generate-only `spec.skew` and `spec.qrSize` [`api/v1alpha1/totpsecretenginekey_webhook.go:72`]
- [ ] [Review][Patch] URL import coverage is still missing from the new import-mode controller test path [`internal/controller/totpsecretenginekey_controller_test.go:138`]
- [ ] [Review][Patch] URL import drift note still omits `algorithm`, `digits`, and `period` convergence requirements [`docs/secret-engines/totp.md:116`]
- [ ] [Review][Patch] TOTP update-time webhook invariants are still uncovered in webhook validation tests [`api/v1alpha1/webhook_validate_update_test.go`]

## Dev Notes

### Integration Test Classification: Vault-Only (Self-Contained)

Per the project's Integration Test Infrastructure Philosophy, the TOTP engine is self-contained within Vault — it requires no external service. Unlike LDAP (needs OpenLDAP), Database (needs PostgreSQL), or RabbitMQ (needs broker), TOTP key management is purely internal to Vault.

**Test approach:**
- Enable a TOTP secrets engine mount (e.g., `vault secrets enable -path=totp/test totp`) — use a `SecretEngineMount` CR to create the mount in the integration test
- Create a TOTPSecretEngineKey CR with `generate=true` (Vault generates the key internally)
- Verify ReconcileSuccessful condition
- Optionally verify Vault can generate codes from the key: `GET {path}/code/{name}`
- Delete the CR and verify the key is removed from Vault

No external service deployment, no mocking needed. Similar in simplicity to the Transit engine integration tests.

### Story Intelligence Chain — Previous Story Context

**Epic 12 stories (12.1 Consul, 12.2 GCP, 12.3 LDAP)** are the most recent completed predecessors:
- **Pattern:** Types file with inline config struct, `toMap()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, webhook, controller, unit tests, integration tests
- **`removeUnsetFields` helper** (from Epic 11): prevents false drift when the operator includes all managed fields with zero defaults but Vault omits fields that were never set
- **`json.Number` requirement**: All numeric fields in `toMap()` must emit `json.Number` (e.g., `json.Number(strconv.Itoa(v))`) because Vault's Go client uses `json.Decoder.UseNumber()`
- **Epic 12 retrospective action items** (all applied): `toMap()` normalization rule (emit Vault-read format), 5-iteration review cap escalation, sprint-status/story-file atomicity guard, final consistency check as blocking gate

**TransitSecretEngineKey** (Epic 11, Story 11.2) is the closest structural analog — also a single "key" CRD with no separate config type:
- Uses `GetPath()` = `{path}/keys/{name}` — same path pattern as TOTP
- Uses `GetConfigPath()` and split create/config writes — TOTP does NOT need this (no separate config endpoint)
- Uses custom `VaultTransitKeyEndpoint` — TOTP does NOT need this; standard `VaultEndpoint` is sufficient
- Transit has immutable create-time fields (type, derived, convergent_encryption, key_size) — TOTP has NO immutable create-time fields beyond path/name
- Transit compares via `configToMap()` — TOTP should compare via a `readVisibleMap()` method (only fields Vault returns on GET)

**No story from Epic 13 has been completed yet** — 13.0 (retroactive docs) and 13.1 (Nomad) are both still in backlog.

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Create/update key | POST | `{path}/keys/{name}` |
| Read key | GET | `{path}/keys/{name}` |
| Delete key | DELETE | `{path}/keys/{name}` |
| List keys | LIST | `{path}/keys` |
| Generate code | GET | `{path}/code/{name}` |
| Validate code | POST | `{path}/code/{name}` |

The CRD manages the key lifecycle (create/read/update/delete). Code generation and validation are runtime operations — NOT managed by the CRD.

### TOTP Key — Two Create Modes

**Generate mode** (`generate=true`): Vault generates the key internally.
- Required fields: `issuer`, `account_name`
- Optional fields: `key_size` (default 20), `algorithm` (default SHA1), `digits` (default 6), `period` (default 30), `skew` (default 1), `exported` (default true), `qr_size` (default 200)
- Response includes `barcode` (base64 PNG) and `url` (otpauth:// URI) when `exported=true`

**Import mode** (`generate=false`, the default): User provides a key from an external service.
- Required: `key` (Base32-encoded root secret) OR `url` (otpauth:// URI with embedded secret)
- Optional fields: `issuer`, `account_name`, `algorithm`, `digits`, `period`

### Vault API Field Reference — Write (`POST {path}/keys/{name}`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `generate` | bool | false | Generate key (true) vs import (false) |
| `exported` | bool | true | Return QR code/url on generate. Only when `generate=true` |
| `key_size` | int | 20 | Key size in bytes. Only when `generate=true` |
| `url` | string | "" | TOTP key URL (otpauth://). Only when `generate=false` |
| `key` | string | "" | Base32 root key. Only when `generate=false` |
| `issuer` | string | "" | Key issuing organization name |
| `account_name` | string | "" | Account name associated with the key |
| `period` | int/duration | 30 | Counter period in seconds |
| `algorithm` | string | "SHA1" | Hash algorithm: SHA1, SHA256, SHA512 |
| `digits` | int | 6 | TOTP code digit count: 6 or 8 |
| `skew` | int | 1 | Allowed delay periods for validation: 0 or 1. Only when `generate=true` |
| `qr_size` | int | 200 | QR code pixel size. Only when `generate=true` and `exported=true`. 0 = no QR |

### Vault API Field Reference — Read (`GET {path}/keys/{name}`)

```json
{
  "data": {
    "account_name": "test@gmail.com",
    "algorithm": "SHA1",
    "digits": 6,
    "issuer": "Google",
    "period": 30
  }
}
```

**Critical observation:** Vault returns ONLY 5 fields on read: `account_name`, `algorithm`, `digits`, `issuer`, `period`. ALL other fields (`key`, `url`, `generate`, `exported`, `key_size`, `qr_size`, `skew`) are write-only — never returned on read.

`digits` and `period` are returned as JSON numbers → `json.Number` in Go.

### Critical: `IsEquivalentToDesiredState` — Compare Only Read-Visible Fields

Since Vault only returns 5 fields on read, `IsEquivalentToDesiredState` must compare ONLY those fields. A dedicated `readVisibleMap()` method should produce the comparison payload:

```go
func (c *TOTPKeyConfig) readVisibleMap() map[string]any {
    payload := map[string]any{}
    payload["issuer"] = c.Issuer
    payload["account_name"] = c.AccountName
    payload["algorithm"] = c.algorithmOrDefault()
    payload["digits"] = json.Number(strconv.Itoa(c.digitsOrDefault()))
    payload["period"] = json.Number(strconv.Itoa(c.periodOrDefault()))
    return payload
}
```

Then `IsEquivalentToDesiredState`:
```go
func (d *TOTPSecretEngineKey) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.TOTPKeyConfig.readVisibleMap()
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

This is different from Transit's approach (which uses `configToMap()` for a subset). TOTP has no config endpoint — the read-visible fields are simply a subset of the create payload.

### Default Value Handling

Vault applies defaults on write. The CRD must emit matching defaults in `readVisibleMap()` to prevent false drift:
- `algorithm`: default `"SHA1"` — use `+kubebuilder:default="SHA1"`
- `digits`: default `6` — use `+kubebuilder:default=6`
- `period`: default `30` — use `+kubebuilder:default=30`

Use helper methods (e.g., `algorithmOrDefault()`) that return the Vault default when the field is empty/zero, similar to Transit's `autoRotatePeriodOrDefault()`.

### `toMap()` — Full Write Payload

`toMap()` produces the full write payload for `POST {path}/keys/{name}`. Include ALL fields, including write-only ones. The `CreateOrUpdate` flow calls `GetPayload()` → `toMap()` when it needs to write.

```go
func (c *TOTPKeyConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["generate"] = c.Generate
    payload["issuer"] = c.Issuer
    payload["account_name"] = c.AccountName
    payload["algorithm"] = c.algorithmOrDefault()
    payload["digits"] = json.Number(strconv.Itoa(c.digitsOrDefault()))
    payload["period"] = json.Number(strconv.Itoa(c.periodOrDefault()))
    if c.Generate {
        payload["exported"] = c.Exported
        payload["key_size"] = json.Number(strconv.Itoa(c.keySizeOrDefault()))
        payload["skew"] = json.Number(strconv.Itoa(c.Skew))
        payload["qr_size"] = json.Number(strconv.Itoa(c.qrSizeOrDefault()))
    } else {
        if c.Key != "" {
            payload["key"] = c.Key
        }
        if c.URL != "" {
            payload["url"] = c.URL
        }
    }
    return payload
}
```

### No Always-Write Pattern Needed

Unlike config types with write-only credentials (AWS secret_key, LDAP bindpass), TOTP keys do NOT need the always-write pattern. The standard `VaultEndpoint.CreateOrUpdate` flow works correctly:
1. Read key from Vault
2. Compare read response against `readVisibleMap()` via `IsEquivalentToDesiredState`
3. Write only if state differs

The write-only fields (`key`, `url`, `generate`, etc.) are provided on every write attempt, but writes only happen when read-visible fields differ. This is correct because:
- If `issuer`/`account_name`/etc. haven't changed → no write → existing key preserved
- If they changed → write → Vault replaces the key (expected behavior)

### No Custom Endpoint Needed

Transit uses `VaultTransitKeyEndpoint` because it has a split create/config model (two separate Vault paths). TOTP has a single endpoint (`{path}/keys/{name}`) for both create and config. Use the standard `VaultEndpoint` via `vaultutils.NewVaultEndpoint(instance)`.

The controller follows the simple pattern from `AWSSecretEngineRoleReconciler` or `NomadSecretEngineRole` (if it existed), not the Transit pattern.

### CRD Field Spec — TOTPSecretEngineKey

```go
type TOTPKeyConfig struct {
    // Generate specifies if a key should be generated by Vault (true) or imported (false).
    // When true, Vault generates the secret key internally.
    // When false, provide key or url to import an existing TOTP secret.
    // +kubebuilder:validation:Optional
    Generate bool `json:"generate,omitempty"`

    // Exported controls whether a QR code and url are returned when generating a key.
    // Only used when generate is true.
    // +kubebuilder:validation:Optional
    Exported *bool `json:"exported,omitempty"`

    // KeySize specifies the size in bytes of the Vault-generated key.
    // Only used when generate is true.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    KeySize int `json:"keySize,omitempty"`

    // URL specifies a TOTP key URL (otpauth:// format) for importing.
    // Only used when generate is false.
    // +kubebuilder:validation:Optional
    URL string `json:"url,omitempty"`

    // Key specifies the Base32-encoded root key for importing.
    // Only used when generate is false.
    // +kubebuilder:validation:Optional
    Key string `json:"key,omitempty"`

    // Issuer specifies the name of the key's issuing organization.
    // +kubebuilder:validation:Optional
    Issuer string `json:"issuer,omitempty"`

    // AccountName specifies the name of the account associated with the key.
    // +kubebuilder:validation:Optional
    AccountName string `json:"accountName,omitempty"`

    // Period specifies the length of time in seconds used to generate a counter for the TOTP code.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=30
    // +kubebuilder:validation:Minimum=1
    Period int `json:"period"`

    // Algorithm specifies the hashing algorithm used to generate the TOTP code.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="SHA1"
    // +kubebuilder:validation:Enum={"SHA1","SHA256","SHA512"}
    Algorithm string `json:"algorithm"`

    // Digits specifies the number of digits in the generated TOTP code (6 or 8).
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=6
    // +kubebuilder:validation:Enum={6,8}
    Digits int `json:"digits"`

    // Skew specifies the number of delay periods allowed when validating a TOTP code (0 or 1).
    // Only used when generate is true.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    // +kubebuilder:validation:Maximum=1
    Skew int `json:"skew,omitempty"`

    // QRSize specifies the pixel size of the square QR code.
    // Only used when generate is true and exported is true. 0 means no QR code.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    QRSize int `json:"qrSize,omitempty"`
}
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- `Algorithm`: `+kubebuilder:default="SHA1"`, `+kubebuilder:validation:Enum={"SHA1","SHA256","SHA512"}`
- `Digits`: `+kubebuilder:default=6`, `+kubebuilder:validation:Enum={6,8}`
- `Period`: `+kubebuilder:default=30`, `+kubebuilder:validation:Minimum=1`
- `Skew`: `+kubebuilder:validation:Minimum=0`, `+kubebuilder:validation:Maximum=1`
- `KeySize`: `+kubebuilder:validation:Minimum=0`
- `QRSize`: `+kubebuilder:validation:Minimum=0`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`

### RBAC Markers

```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=totpsecretenginekeys,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=totpsecretenginekeys/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=totpsecretenginekeys/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### Webhook Validation Rules

**ValidateCreate:**
- If `generate=false` and both `key` and `url` are empty → reject ("one of spec.key or spec.url is required when spec.generate is false")
- Standard validations (digits, algorithm, skew) are handled by kubebuilder markers

**ValidateUpdate:**
- `spec.path` cannot be updated (immutable)
- `spec.name` cannot be updated (immutable)

**ValidateDelete:** no-op

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/totpsecretenginekey_types.go` | NEW | CRD type, VaultObject, ConditionsAware, toMap, readVisibleMap |
| `api/v1alpha1/totpsecretenginekey_webhook.go` | NEW | Webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/totpsecretenginekey_test.go` | NEW | Unit tests for toMap, readVisibleMap, IsEquivalentToDesiredState |
| `internal/controller/totpsecretenginekey_controller.go` | NEW | Reconciler — standard VaultResource pattern |
| `internal/controller/totpsecretenginekey_controller_test.go` | NEW | Integration tests for key lifecycle |
| `cmd/main.go` | UPDATE | Register controller + webhook |
| `config/crd/kustomization.yaml` | UPDATE | Add new CRD YAML file to resources list |
| `test/totpsecretengine/` | NEW | Test YAML fixtures for generate and import mode keys |
| `docs/secret-engines/totp.md` | NEW | User documentation following engine-doc-template |
| `docs/secret-engines/index.md` | UPDATE | Add link to TOTP doc |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~40+ controllers and ~40+ webhooks. New registrations follow the exact same pattern:
- Controller: `(&controller.TOTPSecretEngineKeyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "TOTPSecretEngineKey")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.TOTPSecretEngineKey{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — purely additive.

**`config/crd/kustomization.yaml`**: Add the new CRD YAML file (`bases/redhatcop.redhat.io_totpsecretenginekeys.yaml`) to the `resources` list. Required for Helm chart build.

**`docs/secret-engines/index.md`**: Add a link to the new `totp.md` doc. No existing content changed.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Single key CRD (closest analog) | `api/v1alpha1/transitsecretenginekey_types.go` |
| Key webhook with immutable fields | `api/v1alpha1/transitsecretenginekey_webhook.go` |
| Simple controller (no credential watches) | `internal/controller/transitsecretenginekey_controller.go` |
| Standard VaultEndpoint | `api/v1alpha1/utils/vaultobject.go` (`NewVaultEndpoint`) |
| filterPayloadToDesiredKeys helper | `api/v1alpha1/payload_filter.go` |
| removeUnsetFields helper | `api/v1alpha1/payload_filter.go` |
| Unit test payload construction | Project context: never derive expected from code under test |
| Integration test with Vault-only engine | `internal/controller/transitsecretenginekey_controller_test.go` |
| Test fixture creation (unstructured) | `decoder.CreateFromYAML(ctx, client, path, namespace)` |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**`totpsecretenginekey_test.go`:**

1. `TestTOTPSecretEngineKey_toMap_GenerateMode` — verify: `generate=true`, `issuer`, `account_name`, `algorithm`, `digits` as `json.Number("6")`, `period` as `json.Number("30")`, `exported`, `key_size` as `json.Number`, `skew` as `json.Number`, `qr_size` as `json.Number`; verify `key` and `url` are NOT in output
2. `TestTOTPSecretEngineKey_toMap_ImportMode` — verify: `generate=false`, `key` or `url` included, `issuer`, `account_name`, etc.; verify `exported`/`key_size`/`skew`/`qr_size` are NOT in output
3. `TestTOTPSecretEngineKey_readVisibleMap` — verify: only `issuer`, `account_name`, `algorithm`, `digits`, `period` present; all others absent
4. `TestTOTPSecretEngineKey_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (`{"account_name": "test@gmail.com", "algorithm": "SHA1", "digits": json.Number("6"), "issuer": "Google", "period": json.Number("30")}`), verify returns `true`
5. `TestTOTPSecretEngineKey_IsEquivalentToDesiredState_Mismatch` — change a managed field (e.g., `issuer`), verify returns `false`
6. `TestTOTPSecretEngineKey_IsEquivalentToDesiredState_ExtraVaultFields` — add hypothetical extra Vault fields, verify still returns `true` (filtered out)

### Anti-Patterns / DO NOT

- **DO NOT** use the `VaultTransitKeyEndpoint` — TOTP does not have a split create/config model; use standard `VaultEndpoint`
- **DO NOT** implement code generation or validation endpoints — those are runtime operations, not declarative state; the CRD manages the key definition only
- **DO NOT** use the always-write controller pattern — TOTP has no write-only credentials; standard `CreateOrUpdate` is correct
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` or `readVisibleMap()` in unit tests — construct independent Vault-read-shaped fixtures with hardcoded values
- **DO NOT** forget to add the new CRD YAML file to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** compare write-only fields in `IsEquivalentToDesiredState` — Vault never returns `key`, `url`, `generate`, `exported`, `key_size`, `qr_size`, `skew` on read; comparing them would cause perpetual drift
- **DO NOT** implement PrepareInternalValues for the `key` field — the TOTP key is configuration data (like TLS certificates in other types), not runtime credentials requiring K8s Secret resolution

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`)
- Test fixture directory `test/totpsecretengine/` follows the existing pattern (`test/awssecretengine/`, `test/ssh/`, `test/transit/`)
- No conflicts with existing code — purely additive
- Single CRD type (like TransitSecretEngineKey) — no config/role split

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-13, Story 13.2 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/transitsecretenginekey_types.go — closest structural analog (single key CRD)]
- [Source: api/v1alpha1/transitsecretenginekey_webhook.go — webhook pattern with immutable create-time fields]
- [Source: internal/controller/transitsecretenginekey_controller.go — simple controller with standard VaultEndpoint]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys and removeUnsetFields helpers]
- [Source: _bmad-output/implementation-artifacts/12-3-ldap-ad-secret-engine-config-and-role-crds.md — most recent predecessor story]
- [Source: Vault TOTP Secrets API — https://developer.hashicorp.com/vault/api-docs/secret/totp]
- [Source: docs/engine-doc-template.md — documentation template for DNFR5]

## Code Review Record

### Review Model Used

(to be filled during review — must differ from dev model)

### Review Findings

(to be filled during review)

### Decisions Needed / Decisions Taken

(to be filled during review)

### Fixes Applied

(to be filled during review)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (via Cursor)

### Debug Log References

- Initial integration test run failed: TOTP controller was not registered in `suite_integration_test.go`. Fixed by adding controller registration. Second run passed 112/112 specs (exit code 0).

### Completion Notes List

- Implemented TOTPSecretEngineKey CRD with full VaultObject interface (GetPath, GetPayload, IsEquivalentToDesiredState, IsDeletable=true)
- Implemented `toMap()` with generate/import mode split — generate-mode emits exported, key_size, skew, qr_size; import-mode emits key/url
- Implemented `readVisibleMap()` comparing only the 5 Vault-read-visible fields (issuer, account_name, algorithm, digits, period)
- All numeric fields emit `json.Number` per project convention
- Default helpers (algorithmOrDefault, digitsOrDefault, periodOrDefault, keySizeOrDefault, qrSizeOrDefault) prevent false drift
- Webhook validates import-mode requires key or url, enforces immutable path/name on update
- Controller uses standard `VaultEndpoint` (no custom endpoint, no always-write pattern)
- Unit tests cover toMap generate/import modes, readVisibleMap, IsEquivalentToDesiredState match/mismatch/extra-fields/defaults
- Integration tests cover create (generate mode) → verify Vault state → verify TOTP code generation → delete → verify removal
- All 112 integration specs pass, all unit tests pass

### Change Log

- 2026-08-12: Implemented all 8 tasks for Story 13.2 — TOTP Secret Engine Key CRD

### File List

- api/v1alpha1/totpsecretenginekey_types.go (NEW)
- api/v1alpha1/totpsecretenginekey_webhook.go (NEW)
- api/v1alpha1/totpsecretenginekey_test.go (NEW)
- api/v1alpha1/zz_generated.deepcopy.go (MODIFIED — auto-generated)
- internal/controller/totpsecretenginekey_controller.go (NEW)
- internal/controller/totpsecretenginekey_controller_test.go (NEW)
- internal/controller/suite_integration_test.go (MODIFIED — added TOTP controller registration)
- cmd/main.go (MODIFIED — added controller + webhook registration)
- config/crd/bases/redhatcop.redhat.io_totpsecretenginekeys.yaml (NEW — auto-generated)
- config/crd/kustomization.yaml (MODIFIED — added CRD to resources list)
- config/rbac/role.yaml (MODIFIED — auto-generated RBAC rules)
- config/webhook/manifests.yaml (MODIFIED — auto-generated webhook configs)
- test/totpsecretengine/totp-engine-admin-policy.yaml (NEW)
- test/totpsecretengine/totp-engine-kube-auth-role.yaml (NEW)
- test/totpsecretengine/totp-secret-engine.yaml (NEW)
- test/totpsecretengine/totp-secret-engine-key-generate.yaml (NEW)
- test/totpsecretengine/totp-secret-engine-key-import.yaml (NEW)
- docs/secret-engines/totp.md (NEW)
- docs/secret-engines/index.md (MODIFIED — added TOTP link)
