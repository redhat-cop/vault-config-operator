# Story 1.2: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Engine Mount Types

Status: ready-for-dev

## Story

As an operator developer,
I want unit tests for the engine mount types where `IsEquivalentToDesiredState` compares only the tune config (not the full mount spec),
So that the unique comparison semantics of mount types are verified.

## Acceptance Criteria

1. **Given** an AuthEngineMount instance with Config fields populated **When** `Config.toMap()` is called **Then** the returned map contains `default_lease_ttl`, `max_lease_ttl`, `listing_visibility`, and all other tune fields

2. **Given** an AuthEngineMount instance and a Vault tune response payload **When** `IsEquivalentToDesiredState(payload)` is called **Then** it compares only `Config.toMap()` against the payload (not the full mount spec)

3. **Given** a SecretEngineMount instance with the same pattern **When** `IsEquivalentToDesiredState(payload)` is called **Then** same tune-only comparison behavior is verified

4. **Given** a Vault tune response payload with extra fields not in the config map **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `false` (current behavior — extra fields cause `reflect.DeepEqual` to fail; the reconciler does NOT pre-filter for engine mounts)

## Types Covered

| # | Type | File | Config Struct | Has Existing Tests | Test File |
|---|------|------|---------------|--------------------|-----------|
| 1 | AuthEngineMount | `api/v1alpha1/authenginemount_types.go` | `AuthMountConfig` (10 keys) + `AuthMount` (5 keys) | **No** toMap/IsEquivalent tests | **None** — create new |
| 2 | SecretEngineMount | `api/v1alpha1/secretenginemount_types.go` | `MountConfig` (8 keys) + `Mount` (7 keys) | **Only** `GetPath` tests | `api/v1alpha1/secretenginemount_test.go` |

## Tasks / Subtasks

- [ ] Task 1: Add AuthEngineMount unit tests (AC: 1, 2, 4)
  - [ ] 1.1: Create `api/v1alpha1/authenginemount_test.go`
  - [ ] 1.2: Add `TestAuthEngineMountGetPath` — with `spec.name`, without (fallback to `metadata.name`)
  - [ ] 1.3: Add `TestAuthMountConfigToMap` — verify all 10 keys: `default_lease_ttl`, `max_lease_ttl`, `audit_non_hmac_request_keys`, `audit_non_hmac_response_keys`, `listing_visibility`, `passthrough_request_headers`, `allowed_response_headers`, `token_type`, `description`, `options`
  - [ ] 1.4: Add `TestAuthMountToMap` — verify all 5 keys: `type`, `description`, `config` (nested), `local`, `seal_wrap`
  - [ ] 1.5: Add `TestAuthEngineMountGetPayload` — verify it returns full mount spec (via `AuthMount.toMap()`), not just config
  - [ ] 1.6: Add `TestAuthEngineMountGetTunePayload` — verify it returns only `Config.toMap()`
  - [ ] 1.7: Add `TestAuthEngineMountIsEquivalentToDesiredState` — matching tune config → `true`
  - [ ] 1.8: Add `TestAuthEngineMountIsEquivalentToDesiredState` — non-matching tune config (one field changed) → `false`
  - [ ] 1.9: Add `TestAuthEngineMountIsEquivalentToDesiredState` — extra fields in payload → `false` (document this behavior; see Dev Notes)
  - [ ] 1.10: Add `TestAuthEngineMountIsDeletable` — returns `true`
  - [ ] 1.11: Add `TestAuthEngineMountConditions` — GetConditions/SetConditions round-trip
- [ ] Task 2: Extend SecretEngineMount unit tests (AC: 3, 4)
  - [ ] 2.1: Add `TestMountConfigToMap` to `api/v1alpha1/secretenginemount_test.go` — verify all 8 keys: `default_lease_ttl`, `max_lease_ttl`, `force_no_cache`, `audit_non_hmac_request_keys`, `audit_non_hmac_response_keys`, `listing_visibility`, `passthrough_request_headers`, `allowed_response_headers`
  - [ ] 2.2: Add `TestMountToMap` — verify all 7 keys: `type`, `description`, `config` (nested), `local`, `seal_wrap`, `external_entropy_access`, `options`
  - [ ] 2.3: Add `TestSecretEngineMountGetPayload` — verify it returns full mount spec (via `Mount.toMap()`)
  - [ ] 2.4: Add `TestSecretEngineMountGetTunePayload` — verify it returns `Config.toMap()` (includes `options` and `description` unlike `IsEquivalentToDesiredState`)
  - [ ] 2.5: Add `TestSecretEngineMountIsEquivalentToDesiredState` — matching tune config (after options/description delete) → `true`
  - [ ] 2.6: Add `TestSecretEngineMountIsEquivalentToDesiredState` — non-matching tune config → `false`
  - [ ] 2.7: Add `TestSecretEngineMountIsEquivalentToDesiredState` — extra fields in payload → `false`
  - [ ] 2.8: Add `TestSecretEngineMountIsDeletable` — returns `true`
  - [ ] 2.9: Add `TestSecretEngineMountConditions` — GetConditions/SetConditions round-trip
- [ ] Task 3: Verify all tests pass (AC: all)
  - [ ] 3.1: Run `go test ./api/v1alpha1/ -v -count=1` to confirm all new and existing tests pass
  - [ ] 3.2: Run `make test` to verify no regressions in full unit test suite

## Dev Notes

### Critical: Engine Mount Types Use Tune-Only Comparison

Unlike standard `VaultResource` types that compare the full write payload against the read response, engine mount types use `VaultEngineResource` which has a **completely different reconciliation flow**:

1. `VaultEngineResource.manageReconcileLogic()` checks if the engine exists via `retrieveAccessor()`
2. If the engine **does not exist** → calls `Create()` with `GetPayload()` (full mount spec)
3. If the engine **already exists** → calls `CreateOrUpdateTuneConfig()` which:
   - Reads the current tune config from `GetEngineTunePath()` (e.g., `sys/auth/{path}/tune`)
   - Passes the **raw Vault response** directly to `IsEquivalentToDesiredState(currentTunePayload)` — NO pre-filtering
   - If not equivalent → writes `GetTunePayload()` to `GetEngineTunePath()`

This means `IsEquivalentToDesiredState` receives the **Vault tune API response** and must compare it against only the config-level fields. Both types implement this by comparing `Config.toMap()` (not `GetPayload()`).

[Source: api/v1alpha1/utils/vaultengineobject.go#L84-L97 — `CreateOrUpdateTuneConfig` flow]

### Critical Behavioral Differences Between AuthEngineMount and SecretEngineMount

**AuthEngineMount** — `AuthMountConfig.toMap()` produces **10 keys** including `options`, `token_type`, and `description`:
```go
func (d *AuthEngineMount) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    configMap := d.Spec.Config.toMap()  // 10 keys
    return reflect.DeepEqual(configMap, payload)
}
```

**SecretEngineMount** — `MountConfig.toMap()` produces **8 keys** (no `options`, no `description`, no `token_type`), BUT the method explicitly deletes `options` and `description` before comparison:
```go
func (d *SecretEngineMount) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    configMap := d.Spec.Config.toMap()  // 8 keys
    delete(configMap, "options")        // No-op: MountConfig.toMap() doesn't include these
    delete(configMap, "description")    // No-op: same reason
    return reflect.DeepEqual(configMap, payload)
}
```

The `delete()` calls are currently **no-ops** because `MountConfig.toMap()` doesn't include those keys. They appear to be defensive code. Test this explicitly to document the behavior.

### AC #4 (Extra Fields) — Same Issue as Story 1.1

Both types use `reflect.DeepEqual(configMap, payload)` with **no pre-filtering** of the Vault tune response. The Vault `/sys/auth/{path}/tune` and `/sys/mounts/{path}/tune` endpoints return extra fields beyond what the operator manages (e.g., `force_no_cache` in the auth tune response, `plugin_version`, `user_lockout_config` in newer Vault versions).

**Recommended approach (same as story 1.1):** Write tests proving that extra fields cause `IsEquivalentToDesiredState` to return `false`, with a comment explaining this is the documented behavior. Do NOT modify production code in this story.

### GetTunePayload vs IsEquivalentToDesiredState Discrepancy (SecretEngineMount)

`GetTunePayload()` returns `d.Spec.Config.toMap()` **without** the delete of `options`/`description`. Since `MountConfig.toMap()` doesn't include those keys anyway, the result is identical. Test that both produce the same output.

For `AuthEngineMount`, `GetTunePayload()` also returns `d.Spec.Config.toMap()` — which DOES include `options` and `description`. And `IsEquivalentToDesiredState` uses the same map without deletion. So `GetTunePayload()` and the comparison map are always identical for AuthEngineMount.

### GetPath Differences

**AuthEngineMount** — Path always requires `spec.path`:
```
sys/auth/{spec.path}/{spec.name or metadata.name}
```
`GetEngineListPath()` returns `"sys/auth"`.

**SecretEngineMount** — Path is optional. When `spec.path` is empty, the name is used directly:
```
spec.path set:   sys/mounts/{spec.path}/{spec.name or metadata.name}
spec.path empty: sys/mounts/{spec.name or metadata.name}
```
`GetEngineListPath()` returns `"sys/mounts"`. Existing `GetPath` tests in `secretenginemount_test.go` cover all 4 path combinations.

### Implementation Pattern — Standard Go `testing` Package

All tests in `api/v1alpha1/` use the standard Go `testing` package (NOT Ginkgo). Follow the exact pattern from existing tests in `secretenginemount_test.go` and `identityoidc_test.go`.

**Build tag**: These files do NOT need a build tag — they are in `api/v1alpha1/` which runs with default `go test`.

**Import pattern**:
```go
package v1alpha1

import (
    "reflect"
    "testing"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
```

### Type-Specific Source Code Details

**AuthMountConfig fields** (10 in toMap):
| CRD Field (camelCase) | Vault Key (snake_case) | Go Type |
|---|---|---|
| `defaultLeaseTTL` | `default_lease_ttl` | `string` |
| `maxLeaseTTL` | `max_lease_ttl` | `string` |
| `auditNonHMACRequestKeys` | `audit_non_hmac_request_keys` | `[]string` |
| `auditNonHMACResponseKeys` | `audit_non_hmac_response_keys` | `[]string` |
| `listingVisibility` | `listing_visibility` | `string` (default: `"hidden"`) |
| `passthroughRequestHeaders` | `passthrough_request_headers` | `[]string` |
| `allowedResponseHeaders` | `allowed_response_headers` | `[]string` |
| `tokenType` | `token_type` | `string` |
| `description` | `description` | `*string` (pointer!) |
| `options` | `options` | `map[string]string` |

**MountConfig fields** (8 in toMap):
| CRD Field (camelCase) | Vault Key (snake_case) | Go Type |
|---|---|---|
| `defaultLeaseTTL` | `default_lease_ttl` | `string` |
| `maxLeaseTTL` | `max_lease_ttl` | `string` |
| `forceNoCache` | `force_no_cache` | `bool` (default: `false`) |
| `auditNonHMACRequestKeys` | `audit_non_hmac_request_keys` | `[]string` |
| `auditNonHMACResponseKeys` | `audit_non_hmac_response_keys` | `[]string` |
| `listingVisibility` | `listing_visibility` | `string` (default: `"hidden"`) |
| `passthroughRequestHeaders` | `passthrough_request_headers` | `[]string` |
| `allowedResponseHeaders` | `allowed_response_headers` | `[]string` |

**AuthMountConfig.description is a `*string` pointer** — test both `nil` and non-nil cases. When nil, the map value will be `nil` (not an empty string).

### Previous Story Intelligence (Story 1.1)

Story 1.1 established these patterns:
- Table-driven tests with `t.Run` subtests
- `reflect.DeepEqual` for map comparisons
- Testing both positive (matching) and negative (non-matching) cases
- AC #4 extra-fields: documented behavior as-is (tests proving `reflect.DeepEqual` returns `false` with extra keys)
- Tests for `GetPath` with and without `spec.name` override
- Tests for `IsDeletable` and `GetConditions`/`SetConditions`
- No build tags needed for `api/v1alpha1/` test files

### Project Structure Notes

- Create `api/v1alpha1/authenginemount_test.go` (new file)
- Extend `api/v1alpha1/secretenginemount_test.go` (existing file — already has `TestSecretEngineMountGetPath`)
- No changes to `controllers/` directory
- No decoder methods needed (unit tests only)

### References

- [Source: api/v1alpha1/authenginemount_types.go#L147-L160] — AuthMountConfig.toMap()
- [Source: api/v1alpha1/authenginemount_types.go#L162-L170] — AuthMount.toMap()
- [Source: api/v1alpha1/authenginemount_types.go#L183-L186] — AuthEngineMount.IsEquivalentToDesiredState
- [Source: api/v1alpha1/authenginemount_types.go#L210-L212] — AuthEngineMount.GetTunePayload
- [Source: api/v1alpha1/secretenginemount_types.go#L255-L266] — MountConfig.toMap()
- [Source: api/v1alpha1/secretenginemount_types.go#L268-L278] — Mount.toMap()
- [Source: api/v1alpha1/secretenginemount_types.go#L63-L68] — SecretEngineMount.IsEquivalentToDesiredState (with delete)
- [Source: api/v1alpha1/secretenginemount_types.go#L92-L94] — SecretEngineMount.GetTunePayload
- [Source: api/v1alpha1/utils/vaultengineobject.go#L84-L97] — VaultEngineEndpoint.CreateOrUpdateTuneConfig
- [Source: api/v1alpha1/secretenginemount_test.go] — Existing SecretEngineMount GetPath tests
- [Source: _bmad-output/implementation-artifacts/1-1-unit-tests-tomap-and-isequivalenttodesiredstate-simple-standard-types.md] — Previous story patterns
- [Source: _bmad-output/project-context.md] — Testing rules and conventions

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
