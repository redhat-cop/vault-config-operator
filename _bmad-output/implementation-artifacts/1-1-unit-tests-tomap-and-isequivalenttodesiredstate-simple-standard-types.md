# Story 1.1: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Simple Standard Types

Status: ready-for-dev

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

| # | Type | File | Has `toMap()` | Has Existing Tests | Test File |
|---|------|------|---------------|--------------------|-----------|
| 1 | IdentityOIDCScope | `api/v1alpha1/identityoidcscope_types.go` | Yes (on `*IdentityOIDCScopeSpec`) | Yes — GetPath, GetPayload, omit-empty, IsEquivalentToDesiredState | `api/v1alpha1/identityoidc_test.go` |
| 2 | IdentityOIDCProvider | `api/v1alpha1/identityoidcprovider_types.go` | Yes (on `*IdentityOIDCProviderSpec`) | Yes — GetPath, GetPayload, without-issuer, IsEquivalentToDesiredState | `api/v1alpha1/identityoidc_test.go` |
| 3 | IdentityOIDCClient | `api/v1alpha1/identityoidcclient_types.go` | Yes (on `*IdentityOIDCClientSpec`) | Yes — GetPath, GetPayload, IsEquivalentToDesiredState | `api/v1alpha1/identityoidc_test.go` |
| 4 | IdentityOIDCAssignment | `api/v1alpha1/identityoidcassignment_types.go` | Yes (on `*IdentityOIDCAssignmentSpec`) | Yes — GetPath, GetPayload, IsEquivalentToDesiredState | `api/v1alpha1/identityoidc_test.go` |
| 5 | IdentityTokenConfig | `api/v1alpha1/identitytokenconfig_types.go` | Yes (on `*IdentityTokenConfigSpec`) | Yes — GetPath, GetPayload, empty-issuer, IsNotDeletable, IsEquivalentToDesiredState | `api/v1alpha1/identitytoken_test.go` |
| 6 | IdentityTokenKey | `api/v1alpha1/identitytokenkey_types.go` | Yes (on `*IdentityTokenKeySpec`) | Yes — GetPath, GetPayload, IsEquivalentToDesiredState | `api/v1alpha1/identitytoken_test.go` |
| 7 | IdentityTokenRole | `api/v1alpha1/identitytokenrole_types.go` | Yes (on `*IdentityTokenRoleSpec`) | Yes — GetPath, GetPayload, omits-optional, IsEquivalentToDesiredState | `api/v1alpha1/identitytoken_test.go` |
| 8 | PasswordPolicy | `api/v1alpha1/passwordpolicy_types.go` | **No** — inline in `GetPayload()` | **No** GetPayload/IsEquivalent tests | **None** — create new |
| 9 | AuditRequestHeader | `api/v1alpha1/auditrequestheader_types.go` | **No** — inline in `GetPayload()` | **No tests at all** | **None** — create new |

## Tasks / Subtasks

- [ ] Task 1: Extend IdentityOIDC tests with AC #4 (extra fields ignored) (AC: 4)
  - [ ] 1.1: Add `TestIdentityOIDCScopeIsEquivalentExtraFieldsIgnored` — but **NOTE**: current implementation uses `reflect.DeepEqual(desiredState, payload)` which will **fail** with extra fields. See Dev Notes for details on how to handle this.
  - [ ] 1.2: Add equivalent extra-field tests for IdentityOIDCProvider, IdentityOIDCClient, IdentityOIDCAssignment
- [ ] Task 2: Extend IdentityToken tests with AC #4 (extra fields ignored) (AC: 4)
  - [ ] 2.1: Add extra-field tests for IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole
- [ ] Task 3: Add PasswordPolicy unit tests (AC: 1, 2, 3, 4)
  - [ ] 3.1: Create `api/v1alpha1/passwordpolicy_test.go`
  - [ ] 3.2: Add `TestPasswordPolicyGetPath` — with and without `spec.name`
  - [ ] 3.3: Add `TestPasswordPolicyGetPayload` — verify `{"policy": <HCL>}` output
  - [ ] 3.4: Add `TestPasswordPolicyIsEquivalentToDesiredState` — matching and non-matching
  - [ ] 3.5: Add extra-fields test (same note as Task 1.1 applies)
  - [ ] 3.6: Add `TestPasswordPolicyIsDeletable` and `TestPasswordPolicyConditions`
- [ ] Task 4: Add AuditRequestHeader unit tests (AC: 1, 2, 3, 4)
  - [ ] 4.1: Create `api/v1alpha1/auditrequestheader_test.go`
  - [ ] 4.2: Add `TestAuditRequestHeaderGetPath` — verify `sys/config/auditing/request-headers/{name}`
  - [ ] 4.3: Add `TestAuditRequestHeaderGetPayload` — verify `{"hmac": bool}` output
  - [ ] 4.4: Add `TestAuditRequestHeaderIsEquivalentToDesiredState` — matching and non-matching
  - [ ] 4.5: Add edge case: `IsEquivalentToDesiredState` with missing `hmac` key returns `false`
  - [ ] 4.6: Add edge case: `IsEquivalentToDesiredState` with non-bool `hmac` value returns `false`
  - [ ] 4.7: Add `TestAuditRequestHeaderIsDeletable` and `TestAuditRequestHeaderConditions`
- [ ] Task 5: Verify all tests pass (AC: all)
  - [ ] 5.1: Run `go test ./api/v1alpha1/ -v -count=1` to confirm all new and existing tests pass
  - [ ] 5.2: Run `make test` to verify no regressions in full unit test suite

## Dev Notes

### Critical: AC #4 (Extra Fields) vs Current Implementation

**7 of 9 types use `reflect.DeepEqual(desiredState, payload)` for `IsEquivalentToDesiredState`.** This means if the `payload` map (Vault read response) contains extra keys not in `desiredState`, the comparison returns `false` — the opposite of what AC #4 requires.

**However**, in the real reconciliation flow (`VaultResource.CreateOrUpdate`), the framework **pre-filters** the Vault read response to only include keys present in the desired state before calling `IsEquivalentToDesiredState`. The filtering happens at the `VaultResource` level, not in the type method itself.

**Therefore, for AC #4 testing, you have two valid approaches:**

1. **Document behavior as-is**: Write tests proving that extra fields cause `IsEquivalentToDesiredState` to return `false` for these types. Add a comment explaining that the reconciler framework handles filtering before calling this method. This is the **recommended** approach — it documents actual behavior without changing production code.

2. **Modify implementations**: Change `IsEquivalentToDesiredState` to filter extra keys before comparison. This is **not recommended** for this story — it would change production behavior and requires its own story + broader testing.

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

### Debug Log References

### Completion Notes List

### File List
