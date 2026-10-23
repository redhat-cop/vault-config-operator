# Story 17.2: System Configuration Singleton CRDs (CORSConfig + KeyRotationConfig)

Status: done

## Story

As an operator developer,
I want CRDs for CORSConfig and KeyRotationConfig,
So that Vault's CORS settings and automatic barrier key rotation schedule can be managed declaratively.

## Acceptance Criteria

### CORSConfig

1. **Given** a CORSConfig CR is created with `allowedOrigins` and `allowedHeaders` **When** the reconciler processes it **Then** CORS is enabled in Vault at `sys/config/cors` with the specified origins and headers, and `ReconcileSuccessful=True`
2. **Given** a CORSConfig CR is updated (e.g., origins list changed) **When** the reconciler processes the update **Then** Vault CORS configuration is updated to match
3. **Given** a CORSConfig CR is deleted **When** the reconciler processes deletion **Then** CORS is disabled in Vault (DELETE to `sys/config/cors`)
4. **Given** `allowedHeaders` is set in the spec **When** `IsEquivalentToDesiredState` compares desired vs. Vault state **Then** comparison accounts for Vault appending its default headers to the response (desired headers must be a subset of Vault-returned headers)
5. **Given** `allowedOrigins` is set to `["*"]` (wildcard) **When** the reconciler processes it **Then** Vault CORS allows all origins

### KeyRotationConfig

6. **Given** a KeyRotationConfig CR is created with `maxOperations`, `interval`, and/or `enabled` **When** the reconciler processes it **Then** the auto-rotation config is written to Vault at `sys/rotate/config` and `ReconcileSuccessful=True`
7. **Given** a KeyRotationConfig CR is updated **When** the reconciler processes the update **Then** Vault rotation configuration is updated to match
8. **Given** a KeyRotationConfig CR is deleted **When** the reconciler processes deletion **Then** the Vault config is NOT deleted (`IsDeletable()` returns false — rotation config always exists, only its values change)
9. **Given** `interval` is specified as a Go duration string (e.g., `"4320h"`) **When** `toMap()` emits the value **Then** it emits the duration in whichever format Vault reads back (see Dev Notes on interval format verification)

### Cross-Cutting

10. **Given** the CRD types are implemented **When** the story is marked done **Then** documentation exists at `docs/system-backend/cors.md` and `docs/system-backend/key-rotation.md`
11. **Given** the CRD types are implemented **When** `make manifests generate fmt vet test` and `make lint` are run **Then** all pass with zero errors

## Tasks / Subtasks

- [ ] Task 1 — CORSConfig CRD type (AC: #1, #2, #4, #5)
  - [ ] 1.1 Create `api/v1alpha1/corsconfig_types.go`: CORSConfigSpec with `Connection`, `Authentication`, inline `CORSConfigConfig` struct; `GetPath()` → `"sys/config/cors"`; `IsDeletable()` → `true`; `toMap()` emitting `allowed_origins` and `allowed_headers`
  - [ ] 1.2 Implement custom `IsEquivalentToDesiredState()` for CORSConfig: subset check on `allowed_headers` (Vault appends defaults), exact match on `allowed_origins`
  - [ ] 1.3 Register `&CORSConfig{}` and `&CORSConfigList{}` in `init()`
- [ ] Task 2 — CORSConfig webhook (AC: #1, #5)
  - [ ] 2.1 Create `api/v1alpha1/corsconfig_webhook.go`: `admission.Defaulter[*CORSConfig]`, `admission.Validator[*CORSConfig]`
  - [ ] 2.2 `ValidateCreate`: reject if `allowedOrigins` is empty
  - [ ] 2.3 `ValidateUpdate`: call `Authentication.ValidateTargetNamespaceUpdate()`, reject if `allowedOrigins` is empty
  - [ ] 2.4 Add kubebuilder webhook markers for mutating and validating paths
- [ ] Task 3 — CORSConfig controller (AC: #1, #2, #3)
  - [ ] 3.1 Create `internal/controller/corsconfig_controller.go` following PasswordPolicy controller pattern
  - [ ] 3.2 Add RBAC markers for `corsconfigs` resource
  - [ ] 3.3 Register controller in `cmd/main.go`
  - [ ] 3.4 Register webhook in `cmd/main.go` under `ENABLE_WEBHOOKS` guard
- [ ] Task 4 — KeyRotationConfig CRD type (AC: #6, #7, #8, #9)
  - [ ] 4.1 Create `api/v1alpha1/keyrotationconfig_types.go`: KeyRotationConfigSpec with `Connection`, `Authentication`, inline `KeyRotationConfigConfig` struct; `GetPath()` → `"sys/rotate/config"`; `IsDeletable()` → `false`
  - [ ] 4.2 Implement `toMap()` with `json.Number` for `maxOperations`, string for `interval` (verify format — see Dev Notes), bool for `enabled`
  - [ ] 4.3 Standard `IsEquivalentToDesiredState()` with `filterPayloadToDesiredKeys` + `reflect.DeepEqual`
  - [ ] 4.4 Register `&KeyRotationConfig{}` and `&KeyRotationConfigList{}` in `init()`
- [ ] Task 5 — KeyRotationConfig webhook (AC: #6)
  - [ ] 5.1 Create `api/v1alpha1/keyrotationconfig_webhook.go`: `admission.Defaulter[*KeyRotationConfig]`, `admission.Validator[*KeyRotationConfig]`
  - [ ] 5.2 `ValidateCreate`/`ValidateUpdate`: validate `interval` is a valid Go duration ≥ 24h when specified, validate `maxOperations` is in range [1000000, 3865470566] when specified
  - [ ] 5.3 `ValidateUpdate`: call `Authentication.ValidateTargetNamespaceUpdate()`
  - [ ] 5.4 Add kubebuilder webhook markers
- [ ] Task 6 — KeyRotationConfig controller (AC: #6, #7, #8)
  - [ ] 6.1 Create `internal/controller/keyrotationconfig_controller.go` following PasswordPolicy controller pattern
  - [ ] 6.2 Add RBAC markers for `keyrotationconfigs` resource
  - [ ] 6.3 Register controller in `cmd/main.go`
  - [ ] 6.4 Register webhook in `cmd/main.go` under `ENABLE_WEBHOOKS` guard
- [ ] Task 7 — Unit tests (AC: #1, #4, #5, #6, #9)
  - [ ] 7.1 Create `api/v1alpha1/corsconfig_test.go`: `GetPath()`, `GetPayload()`/`toMap()`, `IsEquivalentToDesiredState` positive + negative + extra fields + Vault-appended-headers, `IsDeletable()` true, `ConditionsAware`
  - [ ] 7.2 Create `api/v1alpha1/keyrotationconfig_test.go`: `GetPath()`, `GetPayload()`/`toMap()` with `json.Number` verification, `IsEquivalentToDesiredState` positive + negative + extra fields, `IsDeletable()` false, `ConditionsAware`
- [ ] Task 8 — Integration tests (AC: #1, #2, #3, #6, #7, #8)
  - [ ] 8.1 Create test fixtures: `test/corsconfig/cors-config.yaml`, `test/keyrotationconfig/key-rotation-config.yaml`
  - [ ] 8.2 Create `internal/controller/corsconfig_controller_test.go` with `//go:build integration`: create → reconcile success → update → reconcile success → delete → verify CORS disabled
  - [ ] 8.3 Create `internal/controller/keyrotationconfig_controller_test.go` with `//go:build integration`: create → reconcile success → update → reconcile success → delete CR → verify Vault config persists (IsDeletable=false)
  - [ ] 8.4 Add controller + webhook registration in `internal/controller/suite_integration_test.go`
- [ ] Task 9 — Code generation + CRD registration (AC: #11)
  - [ ] 9.1 Run `make manifests generate fmt vet`
  - [ ] 9.2 Add new CRD YAML files to `config/crd/kustomization.yaml` resources list
  - [ ] 9.3 Run `make test` — all unit tests pass
  - [ ] 9.4 Run `make lint` — zero violations
- [ ] Task 10 — Documentation (AC: #10)
  - [ ] 10.1 Create `docs/system-backend/` directory (first system-backend docs)
  - [ ] 10.2 Create `docs/system-backend/cors.md` following `docs/engine-doc-template.md`
  - [ ] 10.3 Create `docs/system-backend/key-rotation.md` following `docs/engine-doc-template.md`
  - [ ] 10.4 Add CORSConfig and KeyRotationConfig entries to the System Backend section of `readme.md` with links to the Vault API docs

## Dev Notes

### Singleton Pattern

Both CRDs are **singletons** — they manage a single global Vault configuration with a fixed path and no `{name}` segment. Key differences from standard named-resource types:

1. **No `spec.path` field** — `GetPath()` returns a hardcoded string (`"sys/config/cors"` or `"sys/rotate/config"`)
2. **No `spec.name` field** — the Vault path has no named component
3. **No immutable-path webhook rule** — there's no path to mutate
4. **One-per-cluster convention** — document that only one CR of each type should exist per Vault instance. If multiple CRs exist targeting the same Vault, the last to reconcile wins. A singleton-guard webhook is out of scope for this story but could be added later.

Use `IdentityTokenConfig` (`api/v1alpha1/identitytokenconfig_types.go`) as the **primary pattern** — it's a singleton with a fixed path, no Path/Name fields, and `IsDeletable() false`. CORSConfig follows the same structure but with `IsDeletable() true`.

### CORSConfig — Vault API Details

**Vault API:** `sys/config/cors`
- **POST** enables CORS with specified origins and headers
- **GET** reads current CORS config (returns `enabled`, `allowed_origins`, `allowed_headers`)
- **DELETE** disables CORS entirely
- **Requires `sudo` capability** on the auth token

**CRD Spec Fields:**

```go
type CORSConfigConfig struct {
    // AllowedOrigins is the list of origins permitted for cross-origin requests.
    // Use ["*"] to allow all origins.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinItems=1
    AllowedOrigins []string `json:"allowedOrigins"`

    // AllowedHeaders is the list of additional headers permitted on cross-origin requests.
    // Vault always includes its own default headers; these are appended to that list.
    // +kubebuilder:validation:Optional
    AllowedHeaders []string `json:"allowedHeaders,omitempty"`
}
```

**`toMap()` output:**
```go
func (c *CORSConfigConfig) toMap() map[string]any {
    payload := map[string]any{
        "allowed_origins": toInterfaceArray(c.AllowedOrigins),
    }
    if len(c.AllowedHeaders) > 0 {
        payload["allowed_headers"] = toInterfaceArray(c.AllowedHeaders)
    }
    return payload
}
```

**Critical — Custom `IsEquivalentToDesiredState` for `allowed_headers`:**

The Vault CORS read response appends Vault's default headers to the `allowed_headers` array. Per the [Vault docs](https://developer.hashicorp.com/vault/api-docs/system/config-cors): *"Headers set via this parameter will be appended to the list of headers that Vault allows by default."*

If the operator sets `allowedHeaders: ["X-Custom-Header"]`, Vault returns:
```json
"allowed_headers": ["Content-Type", "X-Requested-With", "X-Vault-Token", ..., "X-Custom-Header"]
```

A naive `reflect.DeepEqual` would report false drift every reconcile. Implement a custom comparison:

```go
func (d *CORSConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.toMap()
    filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)

    // allowed_origins: exact set comparison (order-insensitive)
    if !sameStringSet(desiredState["allowed_origins"], filteredPayload["allowed_origins"]) {
        return false
    }

    // allowed_headers: subset check — our desired headers must all be present
    // in Vault's response (which includes Vault's defaults)
    if headers, ok := desiredState["allowed_headers"]; ok {
        vaultHeaders, _ := payload["allowed_headers"]  // use full payload, not filtered
        if !isSubsetOf(headers, vaultHeaders) {
            return false
        }
    }
    return true
}
```

Use `go-set` (`scylladb/go-set`) for set operations — it's already a project dependency. Or implement simple helper functions: `sameStringSet(a, b interface{}) bool` and `isSubsetOf(subset, superset interface{}) bool` that convert `[]interface{}` to string sets.

**Note on `enabled` field:** The CORS read response includes `"enabled": true`. Do NOT include `enabled` in `toMap()` — the act of writing via POST enables CORS, and DELETE disables it. The `IsDeletable()=true` contract handles disable-on-delete. `filterPayloadToDesiredKeys` excludes `enabled` from comparison since it's not in `toMap()`.

### KeyRotationConfig — Vault API Details

**Vault API:** `sys/rotate/config`
- **POST** configures automatic barrier key rotation
- **GET** reads current config (returns `max_operations`, `interval`, `enabled` in a `data` wrapper)
- **No DELETE** — the config always exists; you can only change values
- **Requires `sudo` capability**

**CRD Spec Fields:**

```go
type KeyRotationConfigConfig struct {
    // MaxOperations is the number of encryptions after which the barrier key is
    // automatically rotated. Must be between 1,000,000 and 3,865,470,566 (default).
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=1000000
    // +kubebuilder:validation:Maximum=3865470566
    MaxOperations int64 `json:"maxOperations,omitempty"`

    // Interval is the duration after which the active key triggers automatic rotation.
    // Specified as a Go duration string (e.g., "4320h"). Must be at least 24h.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`^[0-9]+(h|m|s)+$`
    Interval string `json:"interval,omitempty"`

    // Enabled controls whether automatic key rotation is active. Defaults to true.
    // Setting to false disables rotation but continues tracking encryption counts.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=true
    Enabled bool `json:"enabled"`
}
```

**`toMap()` output:**
```go
func (c *KeyRotationConfigConfig) toMap() map[string]any {
    payload := map[string]any{}
    if c.MaxOperations > 0 {
        payload["max_operations"] = json.Number(strconv.FormatInt(c.MaxOperations, 10))
    }
    if c.Interval != "" {
        payload["interval"] = c.Interval  // See interval format note below
    }
    payload["enabled"] = c.Enabled
    return payload
}
```

**Critical — `interval` Format Verification:**

The Vault API accepts `interval` as a Go duration string (e.g., `"4320h"`). The read response in the [Vault docs](https://developer.hashicorp.com/vault/api-docs/system/rotate-config) shows `"interval": "4320h"` (a string, NOT integer seconds). However, Vault may normalize the format internally (e.g., `"4320h"` → `"4320h0m0s"`).

**The dev MUST verify empirically** what Vault reads back by:
1. Running the integration test infrastructure (`make integration` scaffolding)
2. Writing `"4320h"` via POST
3. Reading back via GET and inspecting the raw `secret.Data["interval"]` value
4. If Vault normalizes (e.g., returns `"4320h0m0s"`), add a normalization step in `toMap()` using `time.ParseDuration()` then `.String()` to emit the canonical Go format

If Vault returns the same string as written, emit the raw string. If Vault returns integer seconds (unlikely for this endpoint but possible), use `durationToSeconds()` per project convention.

**`max_operations` uses `json.Number`** per the mandatory rule — Vault client's `UseNumber()` decoding returns `json.Number` for the integer in the read response.

**`enabled` is a bool** — Vault returns a JSON boolean, which the Go client decodes as `bool` (booleans are not affected by `UseNumber()`). Use `+kubebuilder:default=true` and **no `omitempty`** on the JSON tag per CRD Field Default & Validation Rule #2 (non-zero default).

### Webhook Implementation

**CORSConfig webhook:**
- `ValidateCreate`: reject if `AllowedOrigins` is empty (len == 0)
- `ValidateUpdate`: reject if `AllowedOrigins` is empty; call `newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication)`
- No `spec.path`/`spec.name` immutability check needed (singleton — no path or name fields)

**KeyRotationConfig webhook:**
- `ValidateCreate`/`ValidateUpdate`: validate `Interval` ≥ 24h using `time.ParseDuration()` when non-empty; `MaxOperations` range validated via kubebuilder markers
- `ValidateUpdate`: call `newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication)`

Webhook markers use the standard pattern:
```
//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-corsconfig,...
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-corsconfig,...
```

### Controller Implementation

Both controllers follow the exact PasswordPolicy controller pattern:
1. Embed `vaultresourcecontroller.ReconcilerBase`
2. Fetch instance, check `IsNotFound`, call `prepareContext`
3. Create `vaultresourcecontroller.NewVaultResource(&r.ReconcilerBase, instance)` (standard `VaultResource` — NOT engine or audit variant)
4. Call `vaultResource.Reconcile(ctx1, instance)`

**Registration in `cmd/main.go`:**
```go
// Controller registration (alphabetical order section)
if err = (&controller.CORSConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "CORSConfig")}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "CORSConfig")
    os.Exit(1)
}
// ...
if err = (&controller.KeyRotationConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KeyRotationConfig")}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "KeyRotationConfig")
    os.Exit(1)
}

// Webhook registration (inside ENABLE_WEBHOOKS guard)
if err = (&redhatcopv1alpha1.CORSConfig{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create webhook", "webhook", "CORSConfig")
    os.Exit(1)
}
if err = (&redhatcopv1alpha1.KeyRotationConfig{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create webhook", "webhook", "KeyRotationConfig")
    os.Exit(1)
}
```

### Integration Test Strategy

**Classification:** Both endpoints are `/sys/` endpoints on OSS Vault — **Install in Kind** category. The test Vault instance (root token) has `sudo` capability required by both endpoints.

**CORSConfig integration test (`corsconfig_controller_test.go`):**
- Create → assert `ReconcileSuccessful=True` → verify Vault CORS enabled with correct origins
- Update (change origins) → assert reconcile success → verify updated origins
- Delete → assert CR deleted from K8s → verify CORS disabled in Vault (Vault read should show `enabled: false` or empty config)

**KeyRotationConfig integration test (`keyrotationconfig_controller_test.go`):**
- Create → assert `ReconcileSuccessful=True` → verify Vault rotation config matches
- Update (change max_operations) → assert reconcile success → verify updated config
- Delete → assert CR deleted from K8s → verify Vault config **still exists** with values intact (IsDeletable=false, no finalizer cleanup)
- Use `Ordered` decorator since delete context depends on create context state

**Test fixtures use `decoder.CreateFromYAML` + `DecodeInstance[T]` generic method — no new decoder methods needed.**

### File Structure

| File | Action | Purpose |
|---|---|---|
| `api/v1alpha1/corsconfig_types.go` | **CREATE** | CRD type, VaultObject implementation, toMap(), custom IsEquivalentToDesiredState |
| `api/v1alpha1/corsconfig_webhook.go` | **CREATE** | Defaulter + Validator webhook |
| `api/v1alpha1/corsconfig_test.go` | **CREATE** | Unit tests for GetPath, toMap, IsEquivalentToDesiredState, IsDeletable |
| `api/v1alpha1/keyrotationconfig_types.go` | **CREATE** | CRD type, VaultObject implementation, toMap() with json.Number |
| `api/v1alpha1/keyrotationconfig_webhook.go` | **CREATE** | Defaulter + Validator webhook with duration validation |
| `api/v1alpha1/keyrotationconfig_test.go` | **CREATE** | Unit tests for GetPath, toMap, IsEquivalentToDesiredState, IsDeletable |
| `internal/controller/corsconfig_controller.go` | **CREATE** | Standard VaultResource controller |
| `internal/controller/keyrotationconfig_controller.go` | **CREATE** | Standard VaultResource controller |
| `internal/controller/corsconfig_controller_test.go` | **CREATE** | Integration test (//go:build integration) |
| `internal/controller/keyrotationconfig_controller_test.go` | **CREATE** | Integration test (//go:build integration) |
| `test/corsconfig/cors-config.yaml` | **CREATE** | Test fixture YAML |
| `test/keyrotationconfig/key-rotation-config.yaml` | **CREATE** | Test fixture YAML |
| `docs/system-backend/cors.md` | **CREATE** | CORSConfig documentation |
| `docs/system-backend/key-rotation.md` | **CREATE** | KeyRotationConfig documentation |
| `cmd/main.go` | **MODIFY** | Register controllers + webhooks |
| `config/crd/kustomization.yaml` | **MODIFY** | Add new CRD YAML resources |
| `internal/controller/suite_integration_test.go` | **MODIFY** | Register controllers + webhooks for integration test manager |

### Testing Requirements

**Unit tests must follow the established anti-pattern rules:**
- Never derive expected payloads from `toMap()` — construct independent Vault-read-shaped fixtures
- `json.Number` in simulated Vault payloads (not Go `int` or `float64`)
- Positive + negative + extra-fields tests for `IsEquivalentToDesiredState`

**CORSConfig unit tests must additionally cover:**
- `allowed_headers` subset comparison: Vault payload with extra default headers → returns `true`
- `allowed_headers` missing user header: → returns `false`
- `allowed_origins` order-insensitive comparison
- Empty `allowedHeaders` in spec: `toMap()` omits the key

**KeyRotationConfig unit tests must additionally cover:**
- `max_operations` emits `json.Number` (not int)
- `enabled=false` emits bool `false`
- Empty `interval` omitted from payload

- **Novelty Risk:** LOW — straightforward singleton pattern following IdentityTokenConfig. CORSConfig's custom IsEquivalentToDesiredState for header subset checking is the only non-trivial element.
- Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings. _(KeyRotationConfig `interval` is a special case — Vault returns a duration string, not integer seconds. Emit whichever format Vault reads back; verify empirically.)_

### Project Structure Notes

- **No `spec.path` or `spec.name` fields** — singletons with hardcoded paths. This is consistent with `IdentityTokenConfig` and planned `QuotaConfig` (Story 17.1).
- **`docs/system-backend/` is a new directory** — first system backend docs. No index page required yet (Epic 17 scope), but could be added later.
- Alignment with unified project structure: types in `api/v1alpha1/`, controllers in `internal/controller/`, fixtures in `test/<typename>/`.

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 17, Story 17.2, lines 2937–2981]
- [Source: api/v1alpha1/identitytokenconfig_types.go — singleton pattern reference]
- [Source: api/v1alpha1/passwordpolicy_types.go — simple VaultObject pattern]
- [Source: api/v1alpha1/passwordpolicy_webhook.go — webhook pattern]
- [Source: internal/controller/passwordpolicy_controller.go — standard VaultResource controller pattern]
- [Source: _bmad-output/project-context.md — CRD Field Default & Validation Rules, json.Number rule, duration rule]
- [Vault API: /sys/config/cors](https://developer.hashicorp.com/vault/api-docs/system/config-cors)
- [Vault API: /sys/rotate/config](https://developer.hashicorp.com/vault/api-docs/system/rotate-config)

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

### Review Model Used

- Quick lens: GPT 5.4 Medium (adversarial review)

### Review Findings

Pass 1 — 4 findings: 0 high, 0 medium, 3 low, 1 false
- `[low]` `[reject]` KeyRotationConfig literal string comparison on interval may cause false drift — deferred to empirical verification in integration tests by design
- `[low]` `[reject]` CORSConfig integration test only checks enabled=true, doesn't verify field values — follows established project integration test pattern
- `[low]` `[reject]` KeyRotationConfig integration test only checks existence — same pattern as above
- `[false]` `[reject]` Non-ASCII em dashes violate "ASCII-only rule" — no such rule exists; reviewer fabricated constraint

### Decisions Needed / Decisions Taken

None — all findings rejected.

### Fixes Applied

None — no patches required.
