# Story R1.3: Dependency Modernization — Drop Deprecated Packages & Replace Hand-Rolled Stdlib Equivalents

Status: done

## Story

As an operator developer,
I want deprecated and unnecessary dependencies removed and hand-rolled utilities replaced with stdlib equivalents,
So that the codebase uses maintained, idiomatic Go patterns.

## Acceptance Criteria

1. **Given** `github.com/pkg/errors` is used only in `advanced-funcmap.go` for `errors.Errorf(warn)` **When** replaced with `fmt.Errorf("%s", warn)` and the import removed **Then** the dependency can be dropped from `go.mod`
2. **Given** `io/ioutil.ReadFile` in `controllertestutils/decoder.go` **When** replaced with `os.ReadFile` and the `ioutil` import removed **Then** the deprecated API is eliminated and the SA1019 lint finding is resolved
3. **Given** `VaultSecret.isValid()` creates an empty `multierror.Error{}` and returns `nil` **When** simplified to `return nil` **Then** dead code is removed; `RandomSecret.isValid()` retains real `multierror.Append` aggregation so the `go-multierror` dependency stays
4. **Given** `AddOrReplaceCondition` in `api/v1alpha1/utils/commons.go` duplicates `apimeta.SetStatusCondition` from `k8s.io/apimachinery/pkg/api/meta` **When** all call sites are migrated to the stdlib function **Then** the hand-rolled version can be removed
5. **Given** `ValidateEitherFromVaultSecretOrFromSecret` and `ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret` are near-duplicates **When** consolidated into a single function **Then** code duplication is eliminated
6. **Given** filename `vautlpkiengineobject.go` contains a typo **When** renamed to `vaultpkiengineobject.go` **Then** naming is consistent
7. **Given** all changes **When** `make manifests generate fmt vet test` and `make integration` pass **Then** no regressions

## Tasks / Subtasks

- [x] Task 1: Replace `pkg/errors` with stdlib `fmt` (AC: 1)
  - [x] 1.1: In `controllers/vaultresourcecontroller/advanced-funcmap.go`, replace the two `errors.Errorf(warn)` calls (lines 71, 74) with `fmt.Errorf("%s", warn)`
  - [x] 1.2: Remove the `"github.com/pkg/errors"` import (line 30)
  - [x] 1.3: Add `"fmt"` to the import block if not already present
  - [x] 1.4: Run `go mod tidy` — `github.com/pkg/errors` should be removed from `go.mod` (it has no other direct consumers in Go source)
- [x] Task 2: Replace `ioutil.ReadFile` with `os.ReadFile` (AC: 2)
  - [x] 2.1: In `controllers/controllertestutils/decoder.go` line 60 (`decodeFile`), change `ioutil.ReadFile(filename)` to `os.ReadFile(filename)`
  - [x] 2.2: Remove `"io/ioutil"` from the import block (line 7). `"os"` is already imported (line 8).
- [x] Task 3: Simplify `VaultSecret.isValid()` (AC: 3)
  - [x] 3.1: In `api/v1alpha1/vaultsecret_types.go` lines 182-185, replace the body of `isValid()` with `return nil`
  - [x] 3.2: Remove the `"github.com/hashicorp/go-multierror"` import (line 20)
  - [x] 3.3: Do NOT remove `go-multierror` from `go.mod` — `randomsecret_types.go` has real usage with `multierror.Append` chains in its `isValid()`
- [x] Task 4: Migrate `AddOrReplaceCondition` to `apimeta.SetStatusCondition` (AC: 4)
  - [x] 4.1: In `controllers/vaultresourcecontroller/utils.go` line 154, replace:
    ```go
    conditionsAware.SetConditions(vaultutils.AddOrReplaceCondition(condition, conditionsAware.GetConditions()))
    ```
    with:
    ```go
    conditions := conditionsAware.GetConditions()
    apimeta.SetStatusCondition(&conditions, condition)
    conditionsAware.SetConditions(conditions)
    ```
  - [x] 4.2: Add import `apimeta "k8s.io/apimachinery/pkg/api/meta"` to `utils.go`
  - [x] 4.3: Delete the `AddOrReplaceCondition` function from `api/v1alpha1/utils/commons.go` (lines 44-54)
  - [x] 4.4: Verify no other call sites exist (analysis confirms only one: `utils.go:154`)
- [x] Task 5: Consolidate `ValidateEither` functions (AC: 5)
  - [x] 5.1: Replace both functions in `api/v1alpha1/utils/commons.go` (lines 401-430) with a single `ValidateCredentialSource` function that counts all three non-nil fields (`VaultSecret`, `Secret`, `RandomSecret`)
  - [x] 5.2: Update the 1 call site for `ValidateEitherFromVaultSecretOrFromSecret`:
    - `api/v1alpha1/kubernetessecretengineconfig_types.go:143`
  - [x] 5.3: Update the 5 call sites for `ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret`:
    - `api/v1alpha1/azuresecretengineconfig_types.go:173`
    - `api/v1alpha1/databasesecretengineconfig_types.go:391`
    - `api/v1alpha1/ldapauthengineconfig_types.go:457`
    - `api/v1alpha1/quaysecretengineconfig_types.go:227`
    - `api/v1alpha1/rabbitmqsecretengineconfig_types.go:166`
- [x] Task 6: Rename typo filename (AC: 6)
  - [x] 6.1: `git mv api/v1alpha1/utils/vautlpkiengineobject.go api/v1alpha1/utils/vaultpkiengineobject.go`
  - [x] 6.2: No import updates needed — file is in `package utils`, same package. All references are to functions/types within the package, not to the filename.
- [x] Task 7: Clean up PKI debug logging
  - [x] 7.1: In `controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go`, fix 4 log calls with informal key names:
    - Line 53: `log.Info("DeleteIfExists", "Try to: ", instance)` → `log.Info("deleting vault resource if exists")`
    - Line 70: `log.Info("Delete", "Try to: ", instance)` → `log.Info("processing deletion")`
    - Line 73: `log.Info("Finaliter?", "Try to: ", instance)` → `log.Info("no finalizer found, skipping cleanup")`
    - Line 81: `log.Info("RemoveFinalizer", "Try to: ", instance)` → `log.Info("removing finalizer")`
- [x] Task 8: Verify no regressions (AC: 7)
  - [x] 8.1: Run `make manifests generate fmt vet test`
  - [x] 8.2: Run `golangci-lint run --disable-all --enable=staticcheck ./...` — confirm zero SA1019 findings for `ioutil`
  - [x] 8.3: Run `make integration`

### Review Findings

- [x] [Review][Patch] Preserve `JWTReference` rejection of `randomSecret` [`api/v1alpha1/kubernetessecretengineconfig_types.go:142`] — `ValidateCredentialSource()` now accepts `RandomSecret`, but `JWTReference` is documented to allow only `Secret` or `VaultSecret`, and `setInternalCredentials()` still has no `RandomSecret` handling.

## Dev Notes

### Prerequisites

Stories R1.1, R1.2a, and R1.2b must be completed and merged first. This story is 4th in the lint-fix sequence: R1.1 → R1.2a → R1.2b → R1.3 → R1.2c (lint gate).

R1.1 touches `commons.go` (context key migration), `vautlpkiengineobject.go` (PKI write path fix), and webhook files. R1.2a touches `commons.go` (`ConfigureTLS` error handling) and `decoder.go` (`AddToScheme` panic). R1.2b touches `randomsecret_types.go` (`rand.Seed` removal). Verify line numbers below still match after those merges.

### Task 1: `pkg/errors` Removal — Exact Code

**File:** `controllers/vaultresourcecontroller/advanced-funcmap.go`

The `required` template function (lines 68-78) uses `errors.Errorf` from `github.com/pkg/errors`:

```go
f["required"] = func(warn string, val interface{}) (interface{}, error) {
    if val == nil {
        return val, errors.Errorf(warn)     // line 71 — change to fmt.Errorf("%s", warn)
    } else if _, ok := val.(string); ok {
        if val == "" {
            return val, errors.Errorf(warn) // line 74 — change to fmt.Errorf("%s", warn)
        }
    }
    return val, nil
}
```

Use `fmt.Errorf("%s", warn)` (not `fmt.Errorf(warn)`) to avoid the `warn` string being interpreted as a format string. The `%s` verb is explicit and safe.

After removing the `"github.com/pkg/errors"` import, run `go mod tidy`. Verify `pkg/errors` is removed from `go.mod` — it has zero other Go source consumers (confirmed by repo-wide search). It may remain as a transitive dependency of other modules, which is fine — `go mod tidy` handles that.

### Task 2: `ioutil.ReadFile` Replacement — Exact Code

**File:** `controllers/controllertestutils/decoder.go`

The `decodeFile` function (line 60) still uses `ioutil.ReadFile`:

```go
func (d *decoder) decodeFile(filename string) (runtime.Object, *schema.GroupVersionKind, error) {
    stream, err := ioutil.ReadFile(filename)  // line 60 — change to os.ReadFile(filename)
    // ...
}
```

`os.ReadFile` is the stdlib replacement since Go 1.16. The `"os"` import already exists at line 8. The `CreateFromYAML` method at line 40 already uses `os.ReadFile` — this is the last holdout.

After removing the `"io/ioutil"` import, this resolves the SA1019 finding that R1.2c depends on.

### Task 3: `VaultSecret.isValid()` Simplification — Exact Code

**File:** `api/v1alpha1/vaultsecret_types.go`

Current dead code (lines 182-185):

```go
func (vs *VaultSecret) isValid() error {
    result := &multierror.Error{}
    return result.ErrorOrNil()
}
```

This always returns `nil`. No validation errors are ever appended. Replace with:

```go
func (vs *VaultSecret) isValid() error {
    return nil
}
```

Remove `"github.com/hashicorp/go-multierror"` from the import block (line 20).

**Do NOT touch `RandomSecret.isValid()`** — it has genuine `multierror.Append` usage aggregating 4 validation checks (lines 320-325). The `go-multierror` dependency stays in `go.mod`.

### Task 4: `AddOrReplaceCondition` Migration — Exact Code

**Definition to remove** (`api/v1alpha1/utils/commons.go` lines 44-54):

```go
func AddOrReplaceCondition(c metav1.Condition, conditions []metav1.Condition) []metav1.Condition {
    for i, condition := range conditions {
        if c.Type == condition.Type {
            conditions[i] = c
            return conditions
        }
    }
    conditions = append(conditions, c)
    return conditions
}
```

**Single call site** (`controllers/vaultresourcecontroller/utils.go` line 154):

```go
conditionsAware.SetConditions(vaultutils.AddOrReplaceCondition(condition, conditionsAware.GetConditions()))
```

Replace with `apimeta.SetStatusCondition` from `k8s.io/apimachinery/pkg/api/meta`:

```go
conditions := conditionsAware.GetConditions()
apimeta.SetStatusCondition(&conditions, condition)
conditionsAware.SetConditions(conditions)
```

**Key API difference:** `SetStatusCondition` takes `*[]metav1.Condition` (mutates in place), whereas the hand-rolled function takes and returns a slice. The three-line pattern above handles this correctly.

**Behavioral difference:** `apimeta.SetStatusCondition` manages `LastTransitionTime` automatically — it only updates the timestamp when `Status` actually changes. Review the two condition construction sites in `ManageOutcomeWithRequeue` (lines 136-152): the success condition sets `Status: ConditionTrue` and the failure condition sets `Status: ConditionFalse`. Both set `LastTransitionTime: metav1.Now()`. With `SetStatusCondition`, the explicit `LastTransitionTime` in the condition struct is overridden — `SetStatusCondition` preserves the old timestamp if `Status` hasn't changed. This is actually **better** behavior (fewer spurious condition updates), but verify that no test asserts on `LastTransitionTime` being updated on every reconcile.

Add import: `apimeta "k8s.io/apimachinery/pkg/api/meta"` to `utils.go`. The `k8s.io/apimachinery` module (v0.29.2) is already in `go.mod` — no new dependency needed.

### Task 5: `ValidateEither` Consolidation — Exact Code

**Two current functions** (`api/v1alpha1/utils/commons.go` lines 401-430):

`ValidateEitherFromVaultSecretOrFromSecret` checks only `Secret` + `VaultSecret` (2 fields) but its error message incorrectly mentions `randomSecret`.

`ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret` checks all 3 fields.

**Consolidation:** Replace both with a single function that always checks all 3 fields:

```go
func (credentials *RootCredentialConfig) ValidateCredentialSource() error {
    count := 0
    if credentials.VaultSecret != nil {
        count++
    }
    if credentials.Secret != nil {
        count++
    }
    if credentials.RandomSecret != nil {
        count++
    }
    if count != 1 {
        return errors.New("exactly one of spec.rootCredentials.vaultSecret, spec.rootCredentials.secret, or spec.rootCredentials.randomSecret must be specified")
    }
    return nil
}
```

**Why checking all 3 is safe for the 2-field caller:** `KubernetesSecretEngineConfig` (the sole `ValidateEitherFromVaultSecretOrFromSecret` caller) uses `JWTReference` which is also a `RootCredentialConfig`. If someone ever sets `RandomSecret` on a `JWTReference`, the consolidated function will correctly reject it — which is the right behavior. The original 2-field function had a bug: its error message mentioned `randomSecret` but didn't check it.

**Call site updates** (6 total, all mechanical rename):

| File | Line | Old call | New call |
|------|------|----------|----------|
| `kubernetessecretengineconfig_types.go` | 143 | `r.Spec.JWTReference.ValidateEitherFromVaultSecretOrFromSecret()` | `r.Spec.JWTReference.ValidateCredentialSource()` |
| `azuresecretengineconfig_types.go` | 173 | `r.Spec.AzureCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()` | `r.Spec.AzureCredentials.ValidateCredentialSource()` |
| `databasesecretengineconfig_types.go` | 391 | `r.Spec.RootCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()` | `r.Spec.RootCredentials.ValidateCredentialSource()` |
| `ldapauthengineconfig_types.go` | 457 | `r.Spec.BindCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()` | `r.Spec.BindCredentials.ValidateCredentialSource()` |
| `quaysecretengineconfig_types.go` | 227 | `r.Spec.RootCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()` | `r.Spec.RootCredentials.ValidateCredentialSource()` |
| `rabbitmqsecretengineconfig_types.go` | 166 | `rabbitMQ.Spec.RootCredentials.ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret()` | `rabbitMQ.Spec.RootCredentials.ValidateCredentialSource()` |

### Task 6: Filename Rename

`git mv` the file — no import path changes needed because the file is in `package utils` and all external references use the package import path, not the filename.

Verify after rename: `go build ./...` should compile cleanly.

### Task 7: PKI Debug Logging Cleanup — Exact Lines

**File:** `controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go`

4 log calls use informal/typo'd message strings and the non-standard `"Try to: "` key pattern (note the trailing space and colon — violates structured logging conventions):

| Line | Current | Replacement |
|------|---------|-------------|
| 53 | `log.Info("DeleteIfExists", "Try to: ", instance)` | `log.Info("deleting vault resource if exists")` |
| 70 | `log.Info("Delete", "Try to: ", instance)` | `log.Info("processing deletion")` |
| 73 | `log.Info("Finaliter?", "Try to: ", instance)` | `log.Info("no finalizer found, skipping cleanup")` |
| 81 | `log.Info("RemoveFinalizer", "Try to: ", instance)` | `log.Info("removing finalizer")` |

The `instance` object is already available in structured logging context via controller-runtime's enriched logger. Passing it as a value with a malformed key (`"Try to: "` with trailing space/colon) produces ugly structured log output. The replacement messages are descriptive enough without the object dump.

### Summary of All Changes

| # | File | Change | Risk |
|---|------|--------|------|
| 1 | `controllers/vaultresourcecontroller/advanced-funcmap.go` | Replace 2x `errors.Errorf` → `fmt.Errorf`, drop `pkg/errors` import | Zero — identical runtime behavior |
| 2 | `controllers/controllertestutils/decoder.go` | `ioutil.ReadFile` → `os.ReadFile`, drop `ioutil` import | Zero — `os.ReadFile` is the same function |
| 3 | `api/v1alpha1/vaultsecret_types.go` | Simplify `isValid()` to `return nil`, drop `multierror` import | Zero — function already returned nil |
| 4 | `controllers/vaultresourcecontroller/utils.go` | Migrate to `apimeta.SetStatusCondition` | Low — slightly different `LastTransitionTime` semantics (better) |
| 5 | `api/v1alpha1/utils/commons.go` | Delete `AddOrReplaceCondition`, consolidate `ValidateEither` functions | Low — functional behavior preserved |
| 6 | `api/v1alpha1/kubernetessecretengineconfig_types.go` | Rename `ValidateEither` call | Zero — mechanical rename |
| 7 | `api/v1alpha1/azuresecretengineconfig_types.go` | Rename `ValidateEither` call | Zero — mechanical rename |
| 8 | `api/v1alpha1/databasesecretengineconfig_types.go` | Rename `ValidateEither` call | Zero — mechanical rename |
| 9 | `api/v1alpha1/ldapauthengineconfig_types.go` | Rename `ValidateEither` call | Zero — mechanical rename |
| 10 | `api/v1alpha1/quaysecretengineconfig_types.go` | Rename `ValidateEither` call | Zero — mechanical rename |
| 11 | `api/v1alpha1/rabbitmqsecretengineconfig_types.go` | Rename `ValidateEither` call | Zero — mechanical rename |
| 12 | `api/v1alpha1/utils/vautlpkiengineobject.go` | Rename file → `vaultpkiengineobject.go` | Zero — same package, no import paths change |
| 13 | `controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go` | Fix 4 informal log messages | Zero — log format only |
| 14 | `go.mod` / `go.sum` | `go mod tidy` removes `pkg/errors` direct dependency | Zero — automatic |

Total: **14 file touches, ~13 files modified** (the PKI file is rename, not edit).

### What NOT to Touch

- Do NOT switch `RandomSecret.isValid()` from `multierror` to `errors.Join` — the epic mentions "evaluate" but the actual migration is not required for this story. `errors.Join` has different formatting behavior (`\n`-separated vs `multierror`'s numbered list), which could change webhook error messages. Leave as-is.
- Do NOT modify any `*_types.go` Spec or Status struct fields — there are no CRD schema changes in this story
- Do NOT modify `toMap()`, `IsEquivalentToDesiredState()`, or webhook validation logic
- Do NOT modify `main.go` or controller registration
- Do NOT create a `.golangci.yml` config file

### `go-multierror` Retention Analysis

After removing `multierror` from `vaultsecret_types.go`, the package is still used directly in `randomsecret_types.go`:

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

This is genuine error aggregation — 4 validation checks that can each independently fail. `go mod tidy` will keep `go-multierror` as a direct dependency. Do NOT attempt to migrate this to `errors.Join`.

### `apimeta.SetStatusCondition` — `LastTransitionTime` Semantics

The stdlib `SetStatusCondition` (from `k8s.io/apimachinery/pkg/api/meta`) has smarter `LastTransitionTime` handling than the hand-rolled version:

- If the condition `Type` already exists with the **same `Status`**, it preserves the existing `LastTransitionTime` (avoids spurious updates)
- If the `Status` changed (or the condition is new), it sets `LastTransitionTime` to the current time

The hand-rolled `AddOrReplaceCondition` always overwrites the entire condition struct, including `LastTransitionTime`, on every reconcile. This means switching to `SetStatusCondition` may result in **fewer** status updates (timestamps don't change when status is stable), which is better for etcd write pressure.

Verify that no integration test asserts on `LastTransitionTime` being refreshed on every reconcile. The standard tests check `condition.Status == ConditionTrue` via `Eventually(func() bool {...})` — they don't check timestamps.

### golangci-lint Version

The epic's lint baseline was captured with **golangci-lint v1.64.8**. Install it:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

Verification command for the SA1019 fix:
```bash
golangci-lint run --disable-all --enable=staticcheck ./...
```

This must show zero SA1019 findings after Task 2 (the `ioutil` finding was the last one after R1.2b resolved `rand.Seed`).

### Testing Strategy

- No new tests needed — this is a dependency modernization story
- Existing unit tests (`make test`) exercise `VaultSecret.isValid()`, credential validation, and the `required` template function
- Existing integration tests (`make integration`) exercise the full condition-setting flow via `ManageOutcome`
- The `apimeta.SetStatusCondition` behavioral difference (stable timestamps) is a correctness improvement, not a regression
- `make manifests generate` should produce zero diffs — no `*_types.go` Spec/Status struct changes or RBAC marker changes

### Kind Cluster Considerations

- `make integration` takes ~576-579s (from R1.2a/R1.2b experience)
- Kind cluster can degrade (terminating namespaces) — if tests fail unexpectedly, try a fresh cluster with `make integration` (which recreates the cluster)
- Run `make manifests generate` even for non-type changes — catches unexpected diffs

### Project Structure Notes

- No new files created (except the renamed PKI file)
- All changes are within existing code organization boundaries
- The `ValidateCredentialSource` consolidated function stays in `api/v1alpha1/utils/commons.go` where the originals live
- The `apimeta` import in `utils.go` adds a reference to `k8s.io/apimachinery/pkg/api/meta` which is already an available module

### Previous Story Intelligence

**From R1.2b (immediately preceding):**
- `make integration` takes ~576-579s — budget accordingly
- Kind cluster can degrade — fresh cluster if tests fail unexpectedly
- Run `make manifests generate` even for non-type changes — catches unexpected diffs

**From R1.2a:**
- `decoder.go` was modified (R1.2a added `panic(err)` for `AddToScheme`) — the `ioutil` import on line 7 should still be present (R1.2a was not scoped to fix it)
- `commons.go` was modified by both R1.1 (context keys) and R1.2a (`ConfigureTLS` error handling) — verify line numbers for `AddOrReplaceCondition` (should still be ~44-54) and `ValidateEither` functions (~401-430)

**From R1.1:**
- Largest story — 57 context key migration sites across 18+ files
- `vautlpkiengineobject.go` was modified for the PKI write-path bug fix — the file still has the typo filename (R1.1 scope excluded the rename)
- `commons.go` was modified for context key constants — the `AddOrReplaceCondition` function was untouched

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.3] — acceptance criteria, task list, scope
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, lint baseline, story ordering (R1.1 → R1.2a → R1.2b → R1.3 → R1.2c)
- [Source: _bmad-output/project-context.md#Technology Stack & Versions] — Go 1.22.0, k8s.io/apimachinery v0.29.2
- [Source: _bmad-output/project-context.md#Code Quality Gates] — golangci-lint availability, no committed config
- [Source: _bmad-output/project-context.md#Logging Conventions] — structured key-value logging rules
- [Source: controllers/vaultresourcecontroller/advanced-funcmap.go:30,67-78] — `pkg/errors` import and `errors.Errorf` usage
- [Source: controllers/controllertestutils/decoder.go:7,59-65] — `ioutil` import and `ioutil.ReadFile` usage
- [Source: api/v1alpha1/vaultsecret_types.go:20,182-185] — `multierror` import and dead `isValid()`
- [Source: api/v1alpha1/randomsecret_types.go:27,320-326] — real `multierror.Append` usage (do not touch)
- [Source: api/v1alpha1/utils/commons.go:44-54] — `AddOrReplaceCondition` definition
- [Source: controllers/vaultresourcecontroller/utils.go:154] — sole `AddOrReplaceCondition` call site
- [Source: api/v1alpha1/utils/commons.go:401-430] — two `ValidateEither` functions
- [Source: api/v1alpha1/utils/vautlpkiengineobject.go:17] — typo filename, package utils
- [Source: controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go:53,70,73,81] — informal debug log messages
- [Source: _bmad-output/implementation-artifacts/R1-2b-remove-deprecated-rand-seed.md] — previous story context
- [Source: _bmad-output/implementation-artifacts/R1-2a-fix-unchecked-error-returns-errcheck.md] — R1.2a story context
- [Source: _bmad-output/implementation-artifacts/R1-1-correctness-fixes-context-keys-pki-bug-webhook-loggers-tostring.md] — R1.1 story context

## Dev Agent Record

### Agent Model Used

Opus 4.6 (Cursor Agent)

### Debug Log References

- Integration tests have a pre-existing infrastructure issue: Kind cluster namespace termination race condition causes "unable to create new content in namespace because it is being terminated" errors. This affects ~14 tests across both baseline and post-change runs. Not related to code changes.
- `pkg/errors` remains in `go.mod` as an indirect dependency (transitive via other modules). `go mod tidy` correctly removed it as a direct dependency.
- `go-multierror` correctly retained as direct dependency — `randomsecret_types.go` has genuine `multierror.Append` usage.

### Completion Notes List

- Task 1: Replaced 2x `errors.Errorf(warn)` with `fmt.Errorf("%s", warn)` in `advanced-funcmap.go`. Added `"fmt"` import, removed `"github.com/pkg/errors"` import. `go mod tidy` demoted `pkg/errors` to indirect.
- Task 2: Replaced `ioutil.ReadFile(filename)` with `os.ReadFile(filename)` in `decoder.go`. Removed `"io/ioutil"` import. `"os"` was already imported. Resolves SA1019 lint finding.
- Task 3: Simplified `VaultSecret.isValid()` to `return nil` (was creating empty `multierror.Error{}` and returning `ErrorOrNil()` which always returned nil). Removed `"github.com/hashicorp/go-multierror"` import from `vaultsecret_types.go`.
- Task 4: Migrated sole `AddOrReplaceCondition` call site in `ManageOutcomeWithRequeue` to `apimeta.SetStatusCondition` (3-line pattern: get conditions, set status condition, set conditions). Added `apimeta "k8s.io/apimachinery/pkg/api/meta"` import. Deleted `AddOrReplaceCondition` function from `commons.go`. `SetStatusCondition` has smarter `LastTransitionTime` handling (preserves timestamp when status unchanged).
- Task 5: Consolidated `ValidateEitherFromVaultSecretOrFromSecret` and `ValidateEitherFromVaultSecretOrFromSecretOrFromRandomSecret` into single `ValidateCredentialSource` that checks all 3 fields. Updated all 6 call sites. Error message improved to "exactly one of ... must be specified".
- Task 6: Renamed `vautlpkiengineobject.go` → `vaultpkiengineobject.go` via `git mv`. No import changes needed (same package).
- Task 7: Fixed 4 informal log messages in `vaultpkiengineresourcereconciler.go` — replaced malformed `"Try to: "` key-value pairs with clean descriptive message strings.
- Task 8: All verification passed: `make manifests generate fmt vet test` clean, zero SA1019 staticcheck findings, no manifest diffs. Integration test failures are pre-existing infrastructure issues (identical to baseline).

### Change Log

- 2026-06-01: Implemented all 8 tasks for dependency modernization story R1.3. Replaced deprecated `pkg/errors` and `ioutil` APIs with stdlib equivalents, simplified dead `VaultSecret.isValid()` code, migrated hand-rolled `AddOrReplaceCondition` to stdlib `apimeta.SetStatusCondition`, consolidated duplicate `ValidateEither` functions, renamed typo filename, and cleaned up informal PKI debug logging.

### File List

- `controllers/vaultresourcecontroller/advanced-funcmap.go` — replaced `errors.Errorf` with `fmt.Errorf`, swapped imports
- `controllers/controllertestutils/decoder.go` — replaced `ioutil.ReadFile` with `os.ReadFile`, removed `ioutil` import
- `api/v1alpha1/vaultsecret_types.go` — simplified `isValid()` to `return nil`, removed `multierror` import
- `controllers/vaultresourcecontroller/utils.go` — migrated to `apimeta.SetStatusCondition`, added `apimeta` import
- `api/v1alpha1/utils/commons.go` — deleted `AddOrReplaceCondition`, replaced 2 `ValidateEither` functions with `ValidateCredentialSource`
- `api/v1alpha1/kubernetessecretengineconfig_types.go` — updated call to `ValidateCredentialSource`
- `api/v1alpha1/azuresecretengineconfig_types.go` — updated call to `ValidateCredentialSource`
- `api/v1alpha1/databasesecretengineconfig_types.go` — updated call to `ValidateCredentialSource`
- `api/v1alpha1/ldapauthengineconfig_types.go` — updated call to `ValidateCredentialSource`
- `api/v1alpha1/quaysecretengineconfig_types.go` — updated call to `ValidateCredentialSource`
- `api/v1alpha1/rabbitmqsecretengineconfig_types.go` — updated call to `ValidateCredentialSource`
- `api/v1alpha1/utils/vautlpkiengineobject.go` → `api/v1alpha1/utils/vaultpkiengineobject.go` — renamed (typo fix)
- `controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go` — fixed 4 informal log messages
- `go.mod` / `go.sum` — `go mod tidy` demoted `pkg/errors` to indirect
