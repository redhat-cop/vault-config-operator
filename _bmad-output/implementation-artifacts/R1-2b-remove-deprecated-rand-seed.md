# Story R1.2b: Remove Deprecated `rand.Seed` (SA1019)

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the deprecated `rand.Seed` call removed from the random secret password generator,
So that the code uses the modern Go 1.20+ automatic seeding and the SA1019 lint finding is resolved.

## Acceptance Criteria

1. **Given** `rand.Seed(time.Now().UnixNano())` on line 303 of `randomsecret_types.go` **When** the entire second `init()` function (lines 302-304) is removed **Then** Go 1.20+ automatic seeding provides equivalent behavior
2. **Given** the `"time"` import was only used for `rand.Seed` **When** removed **Then** the unused import is also cleaned up
3. **Given** `randStringBytes` still uses `rand.Intn` from `"math/rand"` **When** the seed call is removed **Then** password generation behavior is unchanged (Go auto-seeds from crypto-quality entropy since 1.20)
4. **Given** all fixes **When** `golangci-lint run --disable-all --enable=staticcheck ./...` shows zero SA1019 findings for `rand.Seed` **Then** lint is clean for this category (note: the `ioutil` SA1019 finding in `decoder.go` is R1.3 scope — it will still appear)
5. **Given** all fixes **When** `make test` and existing `RandomSecret` unit/integration tests pass **Then** no regressions

## Tasks / Subtasks

- [ ] Task 1: Remove the second `init()` function (AC: 1)
  - [ ] 1.1: In `api/v1alpha1/randomsecret_types.go`, delete the entire `init()` block at lines 302-304:
    ```go
    func init() {
        rand.Seed(time.Now().UnixNano())
    }
    ```
    The first `init()` at line 209 (`SchemeBuilder.Register(...)`) is unrelated and must be preserved.
- [ ] Task 2: Clean up `"time"` import (AC: 2)
  - [ ] 2.1: Remove `"time"` from the import block (line 25). It has zero other usages in this file.
  - [ ] 2.2: The `"math/rand"` import (line 23) must stay — `randStringBytes` uses `rand.Intn`.
- [ ] Task 3: Verify lint (AC: 4)
  - [ ] 3.1: Run `golangci-lint run --disable-all --enable=staticcheck ./...`
  - [ ] 3.2: Confirm zero SA1019 findings for `rand.Seed`. The only remaining SA1019 should be the `ioutil` usage in `controllers/controllertestutils/decoder.go` (R1.3 scope — do NOT fix it in this story).
- [ ] Task 4: Verify no regressions (AC: 5)
  - [ ] 4.1: Run `make manifests generate fmt vet test`
  - [ ] 4.2: Run `make integration`

## Dev Notes

### Scope: 1 File, 2 Deletions

This is the smallest possible story. Total changes:

| # | File | Line(s) | Change | Risk |
|---|------|---------|--------|------|
| 1 | `api/v1alpha1/randomsecret_types.go` | 302-304 | Delete second `init()` function containing `rand.Seed(time.Now().UnixNano())` | Zero — Go 1.20+ auto-seeds `math/rand` from crypto entropy |
| 2 | `api/v1alpha1/randomsecret_types.go` | 25 | Remove `"time"` import (only usage was the deleted `rand.Seed` call) | Zero — compile-time guaranteed by `go vet` |

### Why This Is Safe

Since Go 1.20, `math/rand` is automatically seeded with a random value. Calling `rand.Seed()` is not only deprecated but actually counterproductive — manually seeding with `time.Now().UnixNano()` provides *worse* entropy than the default automatic seeding. The `go.mod` declares Go 1.22.0, so automatic seeding is guaranteed.

The `randStringBytes` function (line 306) continues to use `rand.Intn` — this is correct for password character selection. The `math/rand` package is appropriate here (not `crypto/rand`) because the passwords are for Vault random secrets where the randomness quality of `math/rand` with auto-seeding is sufficient.

### File Anatomy: Two `init()` Functions

The file has two separate `init()` functions:

1. **Line 209** — `SchemeBuilder.Register(&RandomSecret{}, &RandomSecretList{})` — required CRD registration, DO NOT TOUCH
2. **Line 302** — `rand.Seed(time.Now().UnixNano())` — the deprecated call to delete

Go allows multiple `init()` functions per file. Deleting the second one has zero effect on the first.

### golangci-lint Version

The epic's lint baseline was captured with **golangci-lint v1.64.8**. The `project-context.md` references v1.59.1 (older). If the dev agent doesn't have v1.64.8, install it:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

Or use the Makefile target: `make golangci-lint`.

### What NOT to Touch

- Do NOT modify `randStringBytes` or any other function in this file
- Do NOT switch from `math/rand` to `crypto/rand` — that's a behavioral change outside scope
- Do NOT fix the `ioutil` SA1019 finding in `controllers/controllertestutils/decoder.go` — that's Story R1.3
- Do NOT modify the first `init()` function (SchemeBuilder registration)

### Project Structure Notes

- Single file change — no new files needed
- No `*_types.go` Spec/Status struct changes — `make manifests generate` should produce zero diffs but must still be run as the verification gate
- `make fmt` will clean up any import formatting automatically

### Testing Strategy

- No new tests needed — this is a lint compliance fix
- Existing unit tests (`make test`) exercise `RandomSecret` password generation via `generatePassword` and `IsValid`
- Existing integration tests (`make integration`) exercise the full `RandomSecret` controller reconcile flow
- The behavioral contract is unchanged: `rand.Intn` was always doing the work; the seed call was just setup that Go now handles automatically

### Previous Story Intelligence (R1.2a)

R1.2a is the immediately preceding story in the lint-fix sequence. Key learnings to apply:
- **`make integration` takes ~576-579s** — budget accordingly for Task 4.2
- **Kind cluster can degrade** (terminating namespaces) — if integration tests fail unexpectedly, try a fresh cluster
- **Run `make manifests generate` even for non-type changes** — catches unexpected diffs
- **`commons.go` was modified by R1.1** for context key migration — unrelated to this story but confirms the epic ordering dependency
- **R1.1 must be completed first** (context key migration), then R1.2a (errcheck), then this story. If R1.1 and R1.2a are not yet merged, coordinate with the sprint status.

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.2b] — acceptance criteria, lint finding, task list
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, lint baseline, story ordering (R1.1 → R1.2a → R1.2b → R1.3 → R1.2c)
- [Source: _bmad-output/project-context.md#Code Quality Gates] — golangci-lint availability
- [Source: api/v1alpha1/randomsecret_types.go:302-304] — `init()` function with `rand.Seed`
- [Source: api/v1alpha1/randomsecret_types.go:306-311] — `randStringBytes` function using `rand.Intn`
- [Source: api/v1alpha1/randomsecret_types.go:209-211] — first `init()` with SchemeBuilder registration (preserve)
- [Source: api/v1alpha1/randomsecret_types.go:23-25] — `math/rand` and `time` imports
- [Source: _bmad-output/implementation-artifacts/R1-2a-fix-unchecked-error-returns-errcheck.md] — previous story context

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### Change Log

### File List
