# Story 9.4: Evaluate and plan pkg/errors migration

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want to assess the effort of migrating from the archived `github.com/pkg/errors` to Go standard error wrapping (`fmt.Errorf` with `%w`),
So that we can decide whether to include this in the upgrade or defer it.

## Pre-Analysis: Corrected Scope

**The epic premise is outdated.** Story R1.3 (completed 2026-06-01) already removed `github.com/pkg/errors` from the codebase. The `pkg/errors` package is no longer in `go.mod` — not even as an indirect dependency. All former `errors.Errorf` call sites were migrated to `fmt.Errorf` in R1.3 Task 1.

**Actual remaining scope:** The one external error-handling dependency left is `github.com/hashicorp/go-multierror v1.1.1` (direct dependency, with its transitive dep `github.com/hashicorp/errwrap v1.1.0`). This package is effectively unmaintained — HashiCorp's own README now states: "For new projects, we recommend using `errors.Join` from the standard library." The Go community consensus (as of 2025-2026) is to migrate away from `go-multierror` to `errors.Join` (available since Go 1.20; project uses Go 1.26).

This story therefore evaluates:
1. The `go-multierror` → `errors.Join` migration (the only remaining external error dependency)
2. Whether to adopt broader `fmt.Errorf("%w", ...)` error wrapping across the codebase
3. A migrate-now vs. defer decision with cost/benefit analysis

## Acceptance Criteria

1. **Given** `pkg/errors` was already removed in Story R1.3, **When** the current error dependency landscape is inventoried, **Then** the report confirms `pkg/errors` is gone and identifies `go-multierror` as the sole remaining external error package.

2. **Given** `go-multierror` is used in 1 file (`api/v1alpha1/randomsecret_types.go`) with 1 call site (`isValid()`), **When** all usages are inventoried and the migration path to `errors.Join` is documented, **Then** a concrete migration plan is produced with exact code changes.

3. **Given** `go-multierror` formats errors differently than `errors.Join`, **When** the formatting difference is analyzed, **Then** the impact on webhook error messages is documented and a decision is recorded.

4. **Given** the migration plan, **When** a migrate-now vs. defer decision is made, **Then** the decision and rationale are documented in this story file.

5. **Given** the decision is "migrate now", **When** the migration is executed, **Then** `go-multierror` is removed as a direct dependency (no source code imports it; it may remain as an indirect transitive dep of `vault/api`), `go build ./...` succeeds, and all tests pass.

6. **Given** the decision is "defer", **When** the rationale is documented, **Then** a follow-up story reference is added to the backlog.

## Tasks / Subtasks

- [x] Task 1: Inventory current error-handling dependencies (AC: #1)
  - [x] 1.1 Confirm `github.com/pkg/errors` is absent from `go.mod` (removed in R1.3)
  - [x] 1.2 Confirm `github.com/hashicorp/go-multierror v1.1.1` is the sole remaining external error package (direct dep)
  - [x] 1.3 Confirm `github.com/hashicorp/errwrap v1.1.0` is its sole transitive dependency (indirect dep)
  - [x] 1.4 Document that all `errors.New()` calls (51+ files) use standard library — no migration needed
  - [x] 1.5 Document that `fmt.Errorf` with `%w` is used in 1 file (`vaultauditobject.go`) — already idiomatic
- [x] Task 2: Analyze `go-multierror` usage and migration path (AC: #2)
  - [x] 2.1 Confirm single direct usage: `randomsecret_types.go` `isValid()` function (lines 313-319)
  - [x] 2.2 Document the 4 validation checks aggregated: `validateEitherPasswordPolicyReferenceOrInline`, `validateInlinePasswordPolicyFormat`, `validateSecretKey`, `validateKVv2DataInPath`
  - [x] 2.3 Document how `isValid()` is called: from `IsValid()` (line 308), which is called from `randomsecret_webhook.go` `ValidateCreate` (line 60) and `ValidateUpdate` (line 78)
  - [x] 2.4 Write concrete migration code: replace `multierror.Error{}` + `multierror.Append` pattern with `[]error` slice + `errors.Join`
- [x] Task 3: Analyze formatting impact (AC: #3)
  - [x] 3.1 Document current `go-multierror` format: numbered bullet list (`2 errors occurred:\n    * error 1\n    * error 2`)
  - [x] 3.2 Document `errors.Join` format: newline-separated (`error 1\nerror 2`)
  - [x] 3.3 Assess impact: error messages appear in webhook admission responses — format change is cosmetic but user-visible
  - [x] 3.4 Check if any tests assert on exact error message format (expected: none do — they check nil/non-nil)
- [x] Task 4: Make and document the migrate-now vs. defer decision (AC: #4)
  - [x] 4.1 Evaluate: 1 call site, ~7 lines changed, removes 2 dependencies, zero risk of behavioral regression
  - [x] 4.2 Record decision in this story file
- [x] Task 5: Execute migration if decision is "migrate now" (AC: #5)
  - [x] 5.1 Replace `isValid()` implementation in `randomsecret_types.go`
  - [x] 5.2 Remove `go-multierror` import, add `errors` import if not present
  - [x] 5.3 Run `go mod tidy` — confirm `go-multierror` removed as direct dep from `go.mod` (remains indirect via vault/api)
  - [x] 5.4 Run `go build ./...` — confirm compilation succeeds
  - [x] 5.5 Run `make test` — confirm all unit tests pass
  - [x] 5.6 Run `make integration` — confirm all integration tests pass

### Review Findings

- [x] [Review][Decision] Resolve AC #5 vs. actual dependency outcome — Updated AC #5 to require removal as a direct dependency only; indirect presence via `vault/api` is acceptable.
- [x] [Review][Patch] Remove stale `go.sum` change claim — corrected below

## Dev Notes

### Current Error Dependency Landscape (Post-R1.3)

| Package | Status | Usage | In go.mod |
|---------|--------|-------|-----------|
| `github.com/pkg/errors` | **GONE** — removed in R1.3 | 0 call sites | Not present |
| `github.com/hashicorp/go-multierror` v1.1.1 | Unmaintained, recommends `errors.Join` | 1 file, 1 call site | Direct |
| `github.com/hashicorp/errwrap` v1.1.0 | Transitive dep of go-multierror | 0 direct usage | Indirect |
| Standard `errors` package | Active stdlib | 51+ files, `errors.New()` only | N/A |
| `fmt.Errorf` with `%w` | Active stdlib | 1 file (`vaultauditobject.go`) | N/A |

### go-multierror: Single Call Site Analysis

**File:** `api/v1alpha1/randomsecret_types.go` (lines 313-319)

```go
func (r *RandomSecret) isValid() error {
    result := &multierror.Error{}
    result = multierror.Append(result, r.validateEitherPasswordPolicyReferenceOrInline())
    result = multierror.Append(result, r.validateInlinePasswordPolicyFormat())
    result = multierror.Append(result, r.validateSecretKey())
    result = multierror.Append(result, r.validateKVv2DataInPath())
    return result.ErrorOrNil()
}
```

**Call chain:** `isValid()` ← `IsValid()` (line 308) ← `randomsecret_webhook.go:ValidateCreate` (line 60) and `ValidateUpdate` (line 78). The returned error becomes the webhook admission response message shown to users who `kubectl apply` an invalid RandomSecret CR.

### Migration: Exact Code Change

**Before (go-multierror):**
```go
import (
    "github.com/hashicorp/go-multierror"
)

func (r *RandomSecret) isValid() error {
    result := &multierror.Error{}
    result = multierror.Append(result, r.validateEitherPasswordPolicyReferenceOrInline())
    result = multierror.Append(result, r.validateInlinePasswordPolicyFormat())
    result = multierror.Append(result, r.validateSecretKey())
    result = multierror.Append(result, r.validateKVv2DataInPath())
    return result.ErrorOrNil()
}
```

**After (errors.Join):**
```go
func (r *RandomSecret) isValid() error {
    return errors.Join(
        r.validateEitherPasswordPolicyReferenceOrInline(),
        r.validateInlinePasswordPolicyFormat(),
        r.validateSecretKey(),
        r.validateKVv2DataInPath(),
    )
}
```

**Why this works:** `errors.Join` accepts any number of `error` arguments, ignores `nil` values, and returns `nil` if all arguments are `nil`. This matches the exact semantics of `multierror.Append` + `ErrorOrNil()`. The `errors` package is already imported in this file (line 21).

### Formatting Difference (User-Visible)

**go-multierror output** (current — when 2 validations fail):
```
2 errors occurred:
    * only one of InlinePasswordPolicy or passwordPolicyName can be defined
    * KVv2 secrets must have /data defined in the path
```

**errors.Join output** (after migration — same 2 failures):
```
only one of InlinePasswordPolicy or passwordPolicyName can be defined
KVv2 secrets must have /data defined in the path
```

**Impact:** The format change is cosmetic. The error content is identical. Webhook admission errors appear in `kubectl apply` output. No tests assert on the exact multi-error format — they check `err != nil` or `err == nil`. The `errors.Join` format is arguably cleaner (no "N errors occurred" prefix, no bullet markers).

### Behavioral Difference: Unwrap Semantics

- `go-multierror` implements `Unwrap() error` (Go 1.13 — chained, returns one error at a time)
- `errors.Join` implements `Unwrap() []error` (Go 1.20 — returns all errors at once)

**Impact on this codebase:** Zero. No code calls `errors.Is`, `errors.As`, or `errors.Unwrap` on the return value of `isValid()`. The webhook simply passes the error to the admission response.

### Decision Framework

| Factor | Migrate Now | Defer |
|--------|-------------|-------|
| **Effort** | ~7 lines changed in 1 file | N/A |
| **Risk** | Zero — identical nil/non-nil semantics | N/A |
| **Dependencies removed** | 2 (`go-multierror`, `errwrap`) | 0 |
| **Formatting change** | Cosmetic — cleaner output | No change |
| **Test changes needed** | None | N/A |
| **Alignment with Go ecosystem** | Uses stdlib, follows HashiCorp's own recommendation | Keeps unmaintained dep |

**Recommendation: Migrate now.** The migration is trivial (1 call site, 7 lines), removes 2 dependencies, follows the library maintainer's own recommendation, and carries zero behavioral risk. Deferring provides no benefit — the migration will never be easier than it is now.

### Error Wrapping Adoption Assessment

**Current state:** The codebase uses `errors.New("message")` almost exclusively (100+ call sites across 51 files). Only 1 call site uses `fmt.Errorf` with `%w` wrapping (`vaultauditobject.go`).

**Should broader `%w` wrapping be adopted?** No — not in this story's scope. The existing pattern is intentional:
- Webhook validators return simple `errors.New("spec.path cannot be updated")` — no wrapping needed because these are terminal validation messages, not errors to be unwrapped by callers.
- Type methods (`PrepareInternalValues`, credential resolution) return `errors.New("secret not found")` — callers log the error and return it up the chain. Wrapping would add no value since the error is always handled by `ManageOutcome` which logs and sets a condition.
- Controllers use `apierrors.IsNotFound(err)` pattern — these are already K8s API error types with built-in type checking.

**Conclusion:** Broader `%w` wrapping would be a refactoring exercise with no practical benefit for this operator's error handling patterns. The codebase's error strategy is correct for its use case.

### Files to Modify

| File | Change |
|------|--------|
| `api/v1alpha1/randomsecret_types.go` | Replace `isValid()` body: `multierror` pattern → `errors.Join` pattern; remove `go-multierror` import |
| `go.mod` | `go mod tidy` removes `go-multierror` (direct) and `errwrap` (indirect) |
| `go.sum` | Automatically regenerated by `go mod tidy` |

**Files NOT modified:**
- No other `*.go` source files — `go-multierror` has zero other usages
- No `*_test.go` files — no test asserts on multi-error format
- No webhook files — `isValid()` is called, not the multierror import
- No CRD regeneration needed (`make manifests generate` not required)
- No Makefile or CI workflow changes
- No `project-context.md` update needed (go-multierror is not listed in Key Dependencies)

### Anti-Pattern Prevention

- **DO NOT** add `errors.Join` wrapping throughout the codebase — this story is scoped to replacing `go-multierror`, not refactoring all error handling.
- **DO NOT** change any `errors.New(...)` call sites to `fmt.Errorf("%w", ...)` — the current unwrapped pattern is intentional (see Error Wrapping Adoption Assessment).
- **DO NOT** change the 4 validation functions (`validateEitherPasswordPolicyReferenceOrInline`, etc.) — only the aggregation in `isValid()` changes.
- **DO NOT** change `randomsecret_webhook.go` — it calls `isValid()` and receives an `error`; the change is transparent.
- **DO NOT** run `go get -u` or upgrade any other dependencies — this is a removal, not an upgrade.
- **DO NOT** add any test assertions on the specific multi-error format — tests should continue checking nil/non-nil.

### Scope Boundary

**IN scope:** Inventory error dependencies, document `go-multierror` → `errors.Join` migration path, make migrate/defer decision, execute migration if decided.

**OUT of scope:**
- Upgrading vault/api (Story 9.1)
- Upgrading ginkgo/gomega (Story 9.2)
- Upgrading peripheral deps (Story 9.3)
- Upgrading Vault test infrastructure (Story 9.5)
- Adopting broader `fmt.Errorf("%w", ...)` wrapping patterns across the codebase
- Any changes to error handling logic (only the aggregation mechanism changes)

### Previous Story Intelligence

**Story R1.3 (Dependency Modernization, done 2026-06-01):** Already removed `pkg/errors` from the codebase (Task 1). Also simplified `VaultSecret.isValid()` which was dead `multierror` code (Task 3). Explicitly noted: "Do NOT switch `RandomSecret.isValid()` from `multierror` to `errors.Join` — the actual migration is not required for that story." This story is the follow-up that R1.3 deferred.

**Story 9.3 (Upgrade peripheral deps):** Pure `go.mod`/`go.sum` dependency upgrade with no source code changes. Same low-risk pattern.

**Story 9.2 (Upgrade ginkgo/gomega):** Pure dependency upgrade. Confirmed `go get` + `go mod tidy` + `go build ./...` pattern works reliably.

**Story 9.1 (Upgrade vault/api):** Pure dependency upgrade. Key lesson: verify `go build ./...` first before running tests.

**Story 9.0 (Fix RabbitMQ serialization bug):** Pure bugfix. No dependency interaction.

**Epic 8 (Go + K8s Stack Upgrade, done 2026-07-19):** Established current Go 1.26 + controller-runtime v0.24 baseline. All tests pass on this stack.

### Project Structure Notes

- Only `api/v1alpha1/randomsecret_types.go`, `go.mod`, and `go.sum` are modified
- No new files created
- Removes 2 dependencies from `go.mod` (net reduction)
- The `errors` package is already imported in `randomsecret_types.go` (line 21) — no new import needed

### References

- [Source: go.mod:11 — current go-multierror v1.1.1 direct dependency]
- [Source: go.mod:51 — current errwrap v1.1.0 indirect dependency (transitive via go-multierror)]
- [Source: api/v1alpha1/randomsecret_types.go:26 — go-multierror import]
- [Source: api/v1alpha1/randomsecret_types.go:313-319 — isValid() with multierror.Append pattern]
- [Source: api/v1alpha1/randomsecret_types.go:308-310 — IsValid() wrapper calling isValid()]
- [Source: api/v1alpha1/randomsecret_webhook.go:60 — ValidateCreate calls isValid()]
- [Source: api/v1alpha1/randomsecret_webhook.go:78 — ValidateUpdate calls isValid()]
- [Source: api/v1alpha1/randomsecret_types.go:21 — standard errors package already imported]
- [Source: _bmad-output/implementation-artifacts/R1-3-dependency-modernization-drop-deprecated-replace-handrolled.md — R1.3 removed pkg/errors, deferred go-multierror migration]
- [Source: _bmad-output/planning-artifacts/epics.md:1962-1976 — Story 9.4 epic definition]
- [Source: https://github.com/hashicorp/go-multierror — README recommends errors.Join for new projects]
- [Source: https://github.com/hashicorp/go-multierror/issues/98 — community confirms unmaintained status]
- [Source: https://pkg.go.dev/errors#Join — Go stdlib errors.Join documentation]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (via Cursor)

### Debug Log References

None — clean execution, no debugging required.

### Completion Notes List

- Inventory confirmed: `pkg/errors` absent (removed in R1.3), `go-multierror v1.1.1` was the sole remaining external error package with `errwrap v1.1.0` as its transitive dep
- Single call site confirmed: `randomsecret_types.go:isValid()` — 4 validation checks aggregated via `multierror.Append` pattern
- Call chain confirmed: `isValid()` ← `IsValid()` ← `ValidateCreate`/`ValidateUpdate` in `randomsecret_webhook.go`
- Formatting impact: cosmetic only — `go-multierror` numbered bullet list → `errors.Join` newline-separated. No tests assert on format
- Decision: **Migrate now** — 1 call site, ~7 lines changed, removes direct dependency, zero behavioral risk
- Migration executed: replaced `multierror.Error{}`/`multierror.Append`/`ErrorOrNil()` with `errors.Join(...)`, removed `go-multierror` import
- `go mod tidy` result: `go-multierror` moved from direct to indirect (still required transitively by `hashicorp/vault/api`); `errwrap` remains indirect for same reason
- `go build ./...` succeeds, `make test` passes (all unit tests), `make integration` passes (all integration tests, 579s)

### Change Log

- Migrated `RandomSecret.isValid()` from `go-multierror` pattern to stdlib `errors.Join` (Date: 2026-07-29)
- Removed `go-multierror` as direct dependency; now indirect-only via `vault/api` (Date: 2026-07-29)

### File List

- `api/v1alpha1/randomsecret_types.go` — replaced `isValid()` body, removed `go-multierror` import
- `go.mod` — `go-multierror` moved from direct to indirect deps block (via `go mod tidy`)
