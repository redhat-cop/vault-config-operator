# Story 7.1: Webhook Validation Tests for Immutable `spec.path` Rule

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want unit tests verifying that `ValidateUpdate` rejects changes to `spec.path` on all types that have this rule,
So that the most critical webhook validation is tested.

## Acceptance Criteria

1. **Given** an existing CR instance and a modified copy with a different `spec.path` **When** `ValidateUpdate(old)` is called **Then** it returns an error containing "spec.path cannot be updated"

2. **Given** an existing CR instance and a modified copy with only non-path fields changed **When** `ValidateUpdate(old)` is called **Then** it returns nil (update is allowed)

3. **Given** all 26 standard types that enforce the immutable `spec.path` rule via `ValidateUpdate` **When** the test suite runs **Then** every type has both a rejection test (path changed) and an allowance test (path unchanged)

4. **Given** `RabbitMQSecretEngineConfig` uses a custom `Handle` method instead of `ValidateUpdate` **When** the test suite runs **Then** it is tested separately via `admission.Request` construction

5. **Given** the tests are added **When** `make test` runs **Then** all existing tests pass with zero regressions

## Tasks / Subtasks

- [x] Task 1: Create table-driven `ValidateUpdate` test for 26 standard types (AC: 1, 2, 3)
  - [x] 1.1: Create `api/v1alpha1/webhook_validate_update_test.go` (no build tag — matches existing pattern in api/v1alpha1/)
  - [x] 1.2: Define test table struct: `{ name, newObj webhook.Validator, oldObj runtime.Object, expectErr bool, errSubstring string }`
  - [x] 1.3: Add rejection entries (path changed) for all 26 types
  - [x] 1.4: Add allowance entries (path unchanged, other field changed) for all 26 types
  - [x] 1.5: Implement test loop calling `ValidateUpdate(old)` and asserting error/nil

- [x] Task 2: Test `RabbitMQSecretEngineConfig` custom Handle (AC: 4)
  - [x] 2.1: In the same file, add a dedicated test for `RabbitMQSecretEngineConfigValidation.Handle`
  - [x] 2.2: Construct `admission.Request` with `Operation: "UPDATE"`, old and new objects marshaled to `req.Object.Raw` / `req.OldObject.Raw`
  - [x] 2.3: Test rejection (path changed → `admission.Errored`)
  - [x] 2.4: Test allowance (path unchanged → `admission.Allowed`)

- [x] Task 3: Verify no regressions (AC: 5)
  - [x] 3.1: Run `make test` — all unit tests pass
  - [x] 3.2: Run `make fmt && make vet` — no formatting or static analysis issues

## Dev Notes

### The 26 Standard Types (ValidateUpdate Pattern)

All use identical logic — compare `r.Spec.Path != old.(*TypeName).Spec.Path` and return `errors.New("spec.path cannot be updated")`:

| # | Type | Webhook File | Additional Immutable Checks |
|---|------|--------------|-----------------------------|
| 1 | `AuthEngineMount` | `authenginemount_webhook.go` | Also blocks non-config spec changes via `reflect.DeepEqual` |
| 2 | `AzureAuthEngineConfig` | `azureauthengineconfig_webhook.go` | — |
| 3 | `AzureSecretEngineConfig` | `azuresecretengineconfig_webhook.go` | — |
| 4 | `AzureSecretEngineRole` | `azuresecretenginerole_webhook.go` | — |
| 5 | `CertAuthEngineConfig` | `certauthengineconfig_webhook.go` | Also blocks `spec.type` changes |
| 6 | `CertAuthEngineRole` | `certauthenginerole_webhook.go` | — |
| 7 | `DatabaseSecretEngineConfig` | `databasesecretengineconfig_webhook.go` | — |
| 8 | `DatabaseSecretEngineRole` | `databasesecretenginerole_webhook.go` | — |
| 9 | `DatabaseSecretEngineStaticRole` | `databasesecretenginestaticrole_webhook.go` | — |
| 10 | `GCPAuthEngineConfig` | `gcpauthengineconfig_webhook.go` | — |
| 11 | `GitHubSecretEngineConfig` | `githubsecretengineconfig_webhook.go` | — |
| 12 | `GitHubSecretEngineRole` | `githubsecretenginerole_webhook.go` | — |
| 13 | `JWTOIDCAuthEngineConfig` | `jwtoidcauthengineconfig_webhook.go` | — |
| 14 | `KubernetesAuthEngineConfig` | `kubernetesauthengineconfig_webhook.go` | — |
| 15 | `KubernetesAuthEngineRole` | `kubernetesauthenginerole_webhook.go` | — |
| 16 | `KubernetesSecretEngineConfig` | `kubernetessecretengineconfig_webhook.go` | — |
| 17 | `KubernetesSecretEngineRole` | `kubernetessecretenginerole_webhook.go` | — |
| 18 | `LDAPAuthEngineConfig` | `ldapauthengineconfig_webhook.go` | — |
| 19 | `PKISecretEngineConfig` | `pkisecretengineconfig_webhook.go` | Also blocks `spec.type` changes |
| 20 | `PKISecretEngineRole` | `pkisecretenginerole_webhook.go` | — |
| 21 | `QuaySecretEngineConfig` | `quaysecretengineconfig_webhook.go` | — |
| 22 | `QuaySecretEngineRole` | `quaysecretenginerole_webhook.go` | — |
| 23 | `QuaySecretEngineStaticRole` | `quaysecretenginestaticrole_webhook.go` | — |
| 24 | `RabbitMQSecretEngineRole` | `rabbitmqsecretenginerole_webhook.go` | — |
| 25 | `RandomSecret` | `randomsecret_webhook.go` | Also blocks `spec.secretKey` changes |
| 26 | `SecretEngineMount` | `secretenginemount_webhook.go` | Also blocks non-config spec changes via `reflect.DeepEqual` |

### The 1 Special Case (Handle Pattern)

`RabbitMQSecretEngineConfig` (`rabbitmqsecretengineconfig_webhook.go`) does NOT implement `webhook.Validator`. It uses a raw `admission.Handler` with a `Handle(ctx, req)` method that JSON-unmarshals `req.Object.Raw` and `req.OldObject.Raw`, then compares `Spec.Path`. On mismatch it returns `admission.Errored(http.StatusBadRequest, errors.New("spec.path cannot be updated"))`.

Testing this type requires constructing an `admission.Request` with raw JSON payloads. The `Handle` method also requires a `client.Client` for CREATE validation (checks for duplicate paths), but for UPDATE the client is unused — only the raw objects matter.

### Test File Design

**File:** `api/v1alpha1/webhook_validate_update_test.go`

**Package:** `v1alpha1` (same package — direct access to unexported types and methods)

**Build tag:** `//go:build !integration` (NOT needed actually — files in `api/v1alpha1/` run with `make test` by default; only `controllers/` uses build tags. Existing tests like `randomsecret_test.go` have no build tag. Omit the build tag to match existing pattern.)

**Test function:** `TestValidateUpdateRejectsPathChange(t *testing.T)` — table-driven.

Each table entry needs:
1. **name**: descriptive test case name (e.g., `"RandomSecret rejects path change"`, `"RandomSecret allows non-path update"`)
2. **newObj**: a minimal instance of the type with a `Spec.Path` value
3. **oldObj**: same type with different (for rejection) or same (for allowance) `Spec.Path`
4. **expectErr**: `true` for rejection, `false` for allowance
5. **errSubstring**: `"spec.path cannot be updated"` for rejection cases

**Minimal instance construction:** Each type needs only `Spec.Path` set for the rejection test. For the allowance test, set the same `Spec.Path` on both old and new, and change a different field (any arbitrary field — e.g., set a `Name` or `Spec.Name` field). Use the simplest possible instance — no `ObjectMeta`, no full spec construction.

**CRITICAL:** The `Spec.Path` field is of type `vaultutils.Path` (a string alias) on most types. Use `vaultutils.Path("some/path")` for assignment. Import `vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"`.

**CRITICAL:** `ValidateUpdate` signature is `ValidateUpdate(old runtime.Object) (admission.Warnings, error)`. The `old` parameter is `runtime.Object`, but inside each webhook it does `old.(*TypeName)` type assertion. The test must pass a pointer to the concrete type, not a bare `runtime.Object`.

**Interface for table-driven test:** Since `webhook.Validator` is the interface, the test table can use:

```go
type pathUpdateTestCase struct {
    name         string
    newObj       webhook.Validator
    oldObj       runtime.Object
    expectErr    bool
    errSubstring string
}
```

Call `newObj.ValidateUpdate(oldObj)` for each entry.

### RabbitMQSecretEngineConfig Test Design

**Separate test function:** `TestRabbitMQSecretEngineConfigHandleRejectsPathChange(t *testing.T)`

Construction:

```go
handler := &RabbitMQSecretEngineConfigValidation{Client: nil} // nil is fine for UPDATE

newObj := &RabbitMQSecretEngineConfig{Spec: RabbitMQSecretEngineConfigSpec{Path: "new/path"}}
oldObj := &RabbitMQSecretEngineConfig{Spec: RabbitMQSecretEngineConfigSpec{Path: "old/path"}}

newRaw, _ := json.Marshal(newObj)
oldRaw, _ := json.Marshal(oldObj)

req := admission.Request{
    AdmissionRequest: admissionv1.AdmissionRequest{
        Operation: admissionv1.Update,
        Object:    runtime.RawExtension{Raw: newRaw},
        OldObject: runtime.RawExtension{Raw: oldRaw},
    },
}

resp := handler.Handle(context.Background(), req)
```

Assert `resp.Allowed == false` for rejection, `resp.Allowed == true` for allowance.

**Import:** `admissionv1 "k8s.io/api/admission/v1"` (already used in `webhook_suite_test.go`).

### Types with `spec.path` but NO Immutable Rule (Gap Discovery)

These 5 types have `Spec.Path` but their webhooks do NOT reject path changes — **do NOT test them** for path rejection (they would pass, giving a false positive). This is an intentional design gap to document, not a bug:

| Type | Webhook Status |
|------|----------------|
| `VaultSecret` | `ValidateUpdate` only calls `r.isValid()` — no path check |
| `GCPAuthEngineRole` | `ValidateUpdate` returns `nil, nil` (stub) |
| `JWTOIDCAuthEngineRole` | `ValidateUpdate` returns `nil, nil` (stub) |
| `AzureAuthEngineRole` | `ValidateUpdate` returns `nil, nil` (stub) |
| `LDAPAuthEngineGroup` | `ValidateUpdate` returns `nil, nil` (stub) |

**Note:** The Audit type has `Path` in `AuditSpec` but has **no webhook file at all** (`audit_webhook.go` does not exist).

### Allowance Test — "Other Field Changed" Approach

For each type's allowance test case, the new and old instances must have:
- **Same** `Spec.Path`
- **Different** value in at least one other field

Use `ObjectMeta.Name` as the simplest differentiator — it's always available and changes won't trigger any validation beyond `spec.path`. Example:

```go
{
    name:      "RandomSecret allows non-path update",
    newObj:    &RandomSecret{ObjectMeta: metav1.ObjectMeta{Name: "new"}, Spec: RandomSecretSpec{Path: "same/path", SecretKey: "key"}},
    oldObj:    &RandomSecret{ObjectMeta: metav1.ObjectMeta{Name: "old"}, Spec: RandomSecretSpec{Path: "same/path", SecretKey: "key"}},
    expectErr: false,
}
```

**CAUTION with AuthEngineMount and SecretEngineMount:** These types have **additional** immutable checks beyond `spec.path`. The allowance test must ensure the "other field changed" doesn't trigger the secondary rule. For `AuthEngineMount`, the only mutable part is `spec.config` — so the allowance test should change a `Config` field. For `SecretEngineMount`, same approach.

**CAUTION with RandomSecret:** Also blocks `spec.secretKey` changes. The allowance test must keep `SecretKey` the same.

**CAUTION with CertAuthEngineConfig and PKISecretEngineConfig:** Also block `spec.type` changes. Keep `Type` the same in allowance test.

### What NOT to Test

- **Additional immutable checks** (e.g., `spec.secretKey` on RandomSecret, `spec.config` restrictions on mount types): Out of scope for this story. The epic's acceptance criteria only mention `spec.path`. Additional immutable field tests would be a separate story.
- **`ValidateCreate` behavior**: Not in scope — this story is specifically about `ValidateUpdate`.
- **`ValidateDelete` behavior**: Not in scope.
- **Envtest / admission webhook integration**: Tests should be pure unit tests calling methods directly, NOT going through the API server. The `webhook_suite_test.go` envtest harness exists but is overkill for this story.

### `isValid()` Side Effects in Allowance Tests

Some types' `ValidateUpdate` calls `r.isValid()` after the path check (e.g., `RandomSecret`, `DatabaseSecretEngineConfig`). For the allowance test, `isValid()` may return an error if required fields are missing. To avoid false failures:
- Set the minimum required fields so `isValid()` passes, OR
- Accept that `isValid()` may return an error and only check that it is NOT the "spec.path cannot be updated" error

The simpler approach: construct allowance test instances with enough fields to pass `isValid()`. Inspect each type's `isValid()` method to determine minimum required fields, or use a generous instance with common fields populated.

**Practical approach:** Many `isValid()` methods only check specific format constraints. If `isValid()` fails, the error message will be distinguishable from "spec.path cannot be updated". The test can verify `err == nil || !strings.Contains(err.Error(), "spec.path cannot be updated")` for allowance cases — but cleaner is to just populate the minimal fields.

**Recommended:** For allowance cases, split into two assertions:
1. If err != nil, assert it does NOT contain "spec.path cannot be updated"
2. This tolerates `isValid()` errors while still proving the path check passed

### No `make manifests generate` Needed

This story only adds a test file. No CRD types, controllers, or webhooks are changed.

### Project Structure Notes

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/webhook_validate_update_test.go` | New | Table-driven ValidateUpdate tests for 26 standard types + RabbitMQSecretEngineConfig Handle test |

### Previous Story Intelligence

**From Story 7.0 (immediate predecessor — test helper refactoring):**
- Story 7.0 is `ready-for-dev` but not yet implemented — no learnings available
- It targets `controllers/` test files, not `api/v1alpha1/` — no overlap with this story

**From Epic 6 Retrospective:**
- "Continue detailed dev notes in story specs" — applied above
- "First stories focused on pure unit tests" — this story is the first pure unit test story
- Suggested ordering: 7.0 → 7.1 → 7.2 → 7.3 → 7.4 → 7.5
- No technical dependencies from Epic 6
- All 83+ integration specs passing, coverage at 53.7%
- Codebase stable on main at `9fc8b3c`

### Git Intelligence (Recent Commits)

```
9fc8b3c Bmad epic 6 (#321)
7ce3e42 Merge pull request #320 from raffaelespazzoli/bmad-epic-5
d64b2b1 Complete Epic 5 retrospective and close epic
```

No recent changes to `api/v1alpha1/*_webhook.go` files outside Epic 6 scope. Codebase clean on main.

### References

- [Source: api/v1alpha1/*_webhook.go] — 43 webhook files, 27 enforce immutable spec.path (26 via ValidateUpdate, 1 via Handle)
- [Source: api/v1alpha1/randomsecret_webhook.go] — Representative ValidateUpdate pattern
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_webhook.go] — Custom Handle pattern for RabbitMQSecretEngineConfig
- [Source: api/v1alpha1/authenginemount_webhook.go] — Additional immutable checks beyond spec.path
- [Source: api/v1alpha1/randomsecret_test.go] — Existing unit test pattern (standard Go testing.T, table-driven)
- [Source: api/v1alpha1/webhook_suite_test.go] — Envtest webhook suite (NOT used for this story)
- [Source: _bmad-output/project-context.md] — Testing rules, webhook patterns
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.1] — Epic requirements and acceptance criteria
- [Source: _bmad-output/implementation-artifacts/epic-6-retro-2026-05-02.md] — Latest retrospective, suggested ordering

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (Cursor Agent)

### Debug Log References

None — all tests passed on first attempt.

### Completion Notes List

- Created `api/v1alpha1/webhook_validate_update_test.go` with 54 test cases total
- Task 1: Table-driven `TestValidateUpdateRejectsPathChange` with 52 cases covering all 26 standard types (26 rejection + 26 allowance)
  - Rejection tests verify `ValidateUpdate` returns error containing "spec.path cannot be updated" when `Spec.Path` changes
  - Allowance tests verify `ValidateUpdate` does NOT return a path-immutability error when only non-path fields change
  - Special handling for AuthEngineMount/SecretEngineMount (config-only changes allowed), CertAuthEngineConfig (keeps Spec.Name same), PKISecretEngineConfig (keeps Type/PrivateKeyType same), RandomSecret (keeps SecretKey same)
  - For types that call `isValid()` after path check, allowance tests tolerate non-path-related validation errors
- Task 2: `TestRabbitMQSecretEngineConfigHandleRejectsPathChange` with 2 cases testing the custom `Handle` method via `admission.Request` construction
  - Rejection: verifies `Allowed=false` and `StatusBadRequest` when path changes
  - Allowance: verifies `Allowed=true` when path unchanged
- Task 3: `make test` passes (0 regressions), `make fmt && make vet` clean
- Coverage in `api/v1alpha1` improved from 16.9% to 19.7%
- No build tag used — matches existing test pattern in `api/v1alpha1/` directory

### File List

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/webhook_validate_update_test.go` | New | Table-driven ValidateUpdate tests for 26 standard types + RabbitMQSecretEngineConfig Handle test |
