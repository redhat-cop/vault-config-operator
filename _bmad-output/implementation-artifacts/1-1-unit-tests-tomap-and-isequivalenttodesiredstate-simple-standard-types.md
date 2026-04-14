# Story 1.1: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Simple Standard Types

Status: done

## Story

As an operator developer,
I want unit tests verifying `toMap()` produces correct snake_case Vault API payloads and `IsEquivalentToDesiredState` correctly compares desired vs actual state,
So that I can confidently modify CRD fields without breaking the declarative reconciliation logic.

## Acceptance Criteria

1. **Given** a CRD instance with all fields populated **When** `toMap()` is called on the inline config struct **Then** the returned map keys are snake_case and match the Vault API field names exactly

2. **Given** a CRD instance and a Vault read response payload that matches the desired state **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true`

3. **Given** a CRD instance and a Vault read response payload with a different value for one managed field **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `false`

4. **Given** a CRD instance and a Vault read response payload containing extra fields not managed by the operator **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true` (extra fields are ignored)

## Types Covered

| # | Type | File | Has `toMap()` | Tests Added by Story 1.1 | Test File |
|---|------|------|---------------|--------------------------|-----------|
| 1 | IdentityOIDCScope | `api/v1alpha1/identityoidcscope_types.go` | Yes (on `*IdentityOIDCScopeSpec`) | Extra-field behavior | `api/v1alpha1/identityoidc_test.go` |
| 2 | IdentityOIDCProvider | `api/v1alpha1/identityoidcprovider_types.go` | Yes (on `*IdentityOIDCProviderSpec`) | Extra-field behavior | `api/v1alpha1/identityoidc_test.go` |
| 3 | IdentityOIDCClient | `api/v1alpha1/identityoidcclient_types.go` | Yes (on `*IdentityOIDCClientSpec`) | Extra-field behavior | `api/v1alpha1/identityoidc_test.go` |
| 4 | IdentityOIDCAssignment | `api/v1alpha1/identityoidcassignment_types.go` | Yes (on `*IdentityOIDCAssignmentSpec`) | Extra-field behavior | `api/v1alpha1/identityoidc_test.go` |
| 5 | IdentityTokenConfig | `api/v1alpha1/identitytokenconfig_types.go` | Yes (on `*IdentityTokenConfigSpec`) | Extra-field behavior | `api/v1alpha1/identitytoken_test.go` |
| 6 | IdentityTokenKey | `api/v1alpha1/identitytokenkey_types.go` | Yes (on `*IdentityTokenKeySpec`) | Extra-field behavior | `api/v1alpha1/identitytoken_test.go` |
| 7 | IdentityTokenRole | `api/v1alpha1/identitytokenrole_types.go` | Yes (on `*IdentityTokenRoleSpec`) | Extra-field behavior | `api/v1alpha1/identitytoken_test.go` |
| 8 | PasswordPolicy | `api/v1alpha1/passwordpolicy_types.go` | **No** — inline in `GetPayload()` | GetPath, GetPayload, IsEquivalent, extra-field, IsDeletable, Conditions | `api/v1alpha1/passwordpolicy_test.go` |
| 9 | AuditRequestHeader | `api/v1alpha1/auditrequestheader_types.go` | **No** — inline in `GetPayload()` | GetPath, GetPayload, IsEquivalent, missing-key, non-bool, extra-field-ignored, IsDeletable, Conditions | `api/v1alpha1/auditrequestheader_test.go` |

## Tasks / Subtasks

- [x] Task 1: Extend IdentityOIDC tests with AC #4 (extra fields ignored) (AC: 4)
  - [x] 1.1: Add `TestIdentityOIDCScopeIsEquivalentExtraFieldsReturnsFalse` — documents that reflect.DeepEqual returns false with extra fields; comment explains reconciler framework pre-filters.
  - [x] 1.2: Add equivalent extra-field tests for IdentityOIDCProvider, IdentityOIDCClient, IdentityOIDCAssignment
- [x] Task 2: Extend IdentityToken tests with AC #4 (extra fields ignored) (AC: 4)
  - [x] 2.1: Add extra-field tests for IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole
- [x] Task 3: Add PasswordPolicy unit tests (AC: 1, 2, 3, 4)
  - [x] 3.1: Create `api/v1alpha1/passwordpolicy_test.go`
  - [x] 3.2: Add `TestPasswordPolicyGetPath` — with and without `spec.name`
  - [x] 3.3: Add `TestPasswordPolicyGetPayload` — verify `{"policy": <HCL>}` output
  - [x] 3.4: Add `TestPasswordPolicyIsEquivalentToDesiredState` — matching and non-matching
  - [x] 3.5: Add extra-fields test (same note as Task 1.1 applies)
  - [x] 3.6: Add `TestPasswordPolicyIsDeletable` and `TestPasswordPolicyConditions`
- [x] Task 4: Add AuditRequestHeader unit tests (AC: 1, 2, 3, 4)
  - [x] 4.1: Create `api/v1alpha1/auditrequestheader_test.go`
  - [x] 4.2: Add `TestAuditRequestHeaderGetPath` — verify `sys/config/auditing/request-headers/{name}`
  - [x] 4.3: Add `TestAuditRequestHeaderGetPayload` — verify `{"hmac": bool}` output
  - [x] 4.4: Add `TestAuditRequestHeaderIsEquivalentToDesiredState` — matching and non-matching
  - [x] 4.5: Add edge case: `IsEquivalentToDesiredState` with missing `hmac` key returns `false`
  - [x] 4.6: Add edge case: `IsEquivalentToDesiredState` with non-bool `hmac` value returns `false`
  - [x] 4.7: Add `TestAuditRequestHeaderIsDeletable` and `TestAuditRequestHeaderConditions`
- [x] Task 5: Verify all tests pass (AC: all)
  - [x] 5.1: Run `go test ./api/v1alpha1/ -v -count=1` to confirm all new and existing tests pass
  - [x] 5.2: Run `make test` to verify no regressions in full unit test suite

### Review Findings

- [x] [Review][Defer] AC #4 still says extra unmanaged fields should return `true`, but the implemented tests intentionally assert `false` for the 7 `reflect.DeepEqual`-based types — deferred pending product decision on whether to amend AC #4 or change the implementations (tracked in Story 7-4)
- [x] [Review][Patch] Remove the incorrect claim that the reconciler pre-filters Vault read payloads before `IsEquivalentToDesiredState` [`api/v1alpha1/identityoidc_test.go:426`, `api/v1alpha1/identitytoken_test.go:275`, `api/v1alpha1/passwordpolicy_test.go:98`, `_bmad-output/implementation-artifacts/1-1-unit-tests-tomap-and-isequivalenttodesiredstate-simple-standard-types.md`]
- [x] [Review][Patch] Refresh the story's `Types Covered` table so it matches the implemented test files and completed tasks [`_bmad-output/implementation-artifacts/1-1-unit-tests-tomap-and-isequivalenttodesiredstate-simple-standard-types.md`]
- [x] [Review][Patch] Strengthen `AuditRequestHeader` path coverage to prove `GetPath()` uses `spec.name` only and does not silently fall back to `metadata.name` [`api/v1alpha1/auditrequestheader_test.go:9`]
- [x] [Review][Patch] Fix the Dev Agent Record counts: `AuditRequestHeader` has 8 top-level test functions, and the diff adds 21 new test functions total, not 9 / 20 [`_bmad-output/implementation-artifacts/1-1-unit-tests-tomap-and-isequivalenttodesiredstate-simple-standard-types.md`]

## Dev Notes

### Critical: AC #4 (Extra Fields) vs Current Implementation

**7 of 9 types use `reflect.DeepEqual(desiredState, payload)` for `IsEquivalentToDesiredState`.** This means if the `payload` map (Vault read response) contains extra keys not in `desiredState`, the comparison returns `false` — the opposite of what AC #4 requires.

**In the real reconciliation flow** (`VaultEndpoint.CreateOrUpdate` in `api/v1alpha1/utils/vaultobject.go`), the framework passes the raw `secret.Data` map from Vault directly into `IsEquivalentToDesiredState` — there is **no pre-filtering**. This means extra Vault-returned keys cause an unnecessary (but harmless) write on every reconcile for these types. Story 7-4 tracks hardening this behavior.

**For AC #4 testing, the implemented approach is:**

1. **Document behavior as-is**: Tests prove that extra fields cause `IsEquivalentToDesiredState` to return `false` for these types. This documents actual behavior without changing production code.

**AuditRequestHeader is the exception**: Its `IsEquivalentToDesiredState` only checks `payload["hmac"].(bool)` — it inherently ignores extra fields. Test AC #4 directly for this type.

### Implementation Pattern — Standard Go `testing` Package

All tests in `api/v1alpha1/` use the standard Go `testing` package (NOT Ginkgo). Follow the exact pattern from existing tests.

**Build tag**: These files do NOT need a build tag — they are in `api/v1alpha1/` which runs with default `go test`. Only `controllers/` test files use `//go:build !integration` or `//go:build integration`.

**Import pattern**:
```go
package v1alpha1

import (
    "reflect"
    "testing"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
```

### Type-Specific Implementation Details

**PasswordPolicy** — No `toMap()` method. Payload is built directly in `GetPayload()`:
```go
func (d *PasswordPolicy) GetPayload() map[string]interface{} {
    return map[string]interface{}{
        "policy": d.Spec.PasswordPolicy,
    }
}
func (d *PasswordPolicy) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    return reflect.DeepEqual(d.GetPayload(), payload)
}
```
- Path: `sys/policies/password/{name}` (or `sys/policies/password/{spec.name}`)
- Single Vault key: `policy`
- Spec field: `Spec.PasswordPolicy` (json: `passwordPolicy`)

**AuditRequestHeader** — No `toMap()`. Custom equivalence logic:
```go
func (d *AuditRequestHeader) GetPayload() map[string]interface{} {
    return map[string]interface{}{
        "hmac": d.Spec.HMAC,
    }
}
func (d *AuditRequestHeader) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    if hmac, ok := payload["hmac"].(bool); ok {
        return hmac == d.Spec.HMAC
    }
    return false
}
```
- Path: `sys/config/auditing/request-headers/{spec.name}` (uses `d.Spec.Name` not `d.Name`)
- Single Vault key: `hmac` (bool)
- Note: `GetPath()` always uses `d.Spec.Name`, not `d.Name` — no `metadata.name` fallback
- `IsEquivalentToDesiredState` type-asserts `hmac` as `bool` — returns `false` for missing key or non-bool type

**IdentityOIDCScope** — `toMap()` omits empty fields:
```go
func (i *IdentityOIDCScopeSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    if i.Template != "" { payload["template"] = i.Template }
    if i.Description != "" { payload["description"] = i.Description }
    return payload
}
```
- Fields omitted when empty string → tests should cover both populated and empty cases

**IdentityOIDCProvider** — `toMap()` omits `issuer` when empty, slices always present:
```go
func (i *IdentityOIDCProviderSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    if i.Issuer != "" { payload["issuer"] = i.Issuer }
    payload["allowed_client_ids"] = i.AllowedClientIDs
    payload["scopes_supported"] = i.ScopesSupported
    return payload
}
```
- `allowed_client_ids` and `scopes_supported` are always present (may be `nil`)

**IdentityOIDCClient** — All 6 keys always present:
- `key`, `redirect_uris`, `assignments`, `client_type`, `id_token_ttl`, `access_token_ttl`

**IdentityOIDCAssignment** — Both keys always present:
- `entity_ids`, `group_ids`

**IdentityTokenConfig** — Single key `issuer` always present (even if empty string)

**IdentityTokenKey** — 4 keys always present:
- `rotation_period`, `verification_ttl`, `allowed_client_ids`, `algorithm`

**IdentityTokenRole** — `template` and `client_id` omitted when empty; `key` and `ttl` always present

### Project Structure Notes

- New test files go in `api/v1alpha1/` alongside existing `identityoidc_test.go` and `identitytoken_test.go`
- Create `api/v1alpha1/passwordpolicy_test.go` (new file)
- Create `api/v1alpha1/auditrequestheader_test.go` (new file)
- Extensions to existing OIDC/Token tests go in the existing test files
- No changes to `controllers/` directory in this story
- No decoder methods needed (those are for integration tests, not unit tests)

### Testing Standards Summary

- Use `testing.T` with table-driven subtests (`t.Run`)
- Use `reflect.DeepEqual` for map comparisons
- Verify both positive (matching) and negative (non-matching) cases
- Test omit-empty behavior where applicable
- Test `GetPath` with and without `spec.name` override
- Test `IsDeletable` and `GetConditions`/`SetConditions` for new types
- Run `go test ./api/v1alpha1/ -v -count=1` to validate

### References

- [Source: api/v1alpha1/identityoidc_test.go] — Existing OIDC unit test patterns
- [Source: api/v1alpha1/identitytoken_test.go] — Existing Token unit test patterns
- [Source: api/v1alpha1/passwordpolicy_types.go#L38-L51] — PasswordPolicy GetPath/GetPayload/IsEquivalentToDesiredState
- [Source: api/v1alpha1/auditrequestheader_types.go#L88-L111] — AuditRequestHeader GetPath/GetPayload/IsEquivalentToDesiredState
- [Source: _bmad-output/project-context.md] — Testing rules and conventions

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6 (via Cursor)

### Debug Log References
No debug issues encountered. All tests passed on first run.

### Completion Notes List
- **Task 1 & 2 (AC #4 — extra-field behavior):** Followed the "document behavior as-is" approach from Dev Notes. All 7 types that use `reflect.DeepEqual` return `false` when extra fields are present; the reconciler passes raw Vault data with no filtering, so extra keys trigger an unnecessary write (tracked in Story 7-4). AuditRequestHeader is the exception — its field-specific `payload["hmac"].(bool)` check inherently ignores extra fields, confirmed with a test asserting `true`.
- **Task 3 (PasswordPolicy):** Created new test file with 6 tests covering GetPath (with/without spec.name), GetPayload (single "policy" key with HCL content), IsEquivalentToDesiredState (matching/non-matching), extra-fields behavior, IsDeletable, and Conditions.
- **Task 4 (AuditRequestHeader):** Created new test file with 8 test functions covering GetPath (with metadata.name regression guard), GetPayload (hmac true/false), IsEquivalentToDesiredState (4 matching/non-matching cases), missing hmac key edge case, non-bool hmac value edge cases (string, int, nil), extra-fields-ignored behavior, IsDeletable, and Conditions.
- **Task 5:** All 49 tests pass in `api/v1alpha1/`. `make test` passes with no regressions.

### Change Log
- 2026-04-12: Implemented story 1.1 — added 21 new test functions across 4 files

### File List
- `api/v1alpha1/identityoidc_test.go` (modified — added 4 extra-field tests)
- `api/v1alpha1/identitytoken_test.go` (modified — added 3 extra-field tests)
- `api/v1alpha1/passwordpolicy_test.go` (new — 6 test functions)
- `api/v1alpha1/auditrequestheader_test.go` (new — 8 test functions)
