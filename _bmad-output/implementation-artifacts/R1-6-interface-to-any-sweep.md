# Story R1.6: `interface{}` to `any` Sweep

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want all occurrences of `interface{}` replaced with the `any` alias (Go 1.18+),
So that the codebase uses idiomatic modern Go.

## Acceptance Criteria

1. **Given** the codebase uses Go 1.22 **When** all `interface{}` occurrences are replaced with `any` **Then** the code compiles identically
2. **Given** the rename is purely mechanical **When** `make test` is run **Then** all tests pass unchanged
3. **Given** no `interface{}` remains in non-generated files **When** searched **Then** zero matches (excluding `zz_generated.deepcopy.go` which is auto-generated)

## Tasks / Subtasks

- [ ] Task 1: Replace `interface{}` → `any` across `api/v1alpha1/` (AC: 1, 3)
  - [ ] 1.1: Replace in all `*_types.go` files (~40 files, ~4 occurrences each)
  - [ ] 1.2: Replace in all `*_test.go` files (~30 files, varies per file)
  - [ ] 1.3: Replace in `payload_filter.go` (3 occurrences)
  - [ ] 1.4: Replace in `api/v1alpha1/utils/` — `commons.go` (2), `vaultutils.go` (4), `vaultobject.go` (7), `vaultengineobject.go` (4), `vaultauditobject.go` (1), `vautlpkiengineobject.go` (6), `vaultobject_test.go` (45)
  - [ ] 1.5: Verify `zz_generated.deepcopy.go` is NOT touched — currently contains zero `interface{}` so no exclusion gymnastics needed
  - [ ] 1.6: Run `go build ./api/...` to confirm compilation
- [ ] Task 2: Replace `interface{}` → `any` across `controllers/` (AC: 1, 3)
  - [ ] 2.1: Replace in `controllers/vaultresourcecontroller/advanced-funcmap.go` (31 occurrences — highest count in controllers)
  - [ ] 2.2: Replace in `controllers/vaultsecret_controller.go` (2 occurrences)
  - [ ] 2.3: Replace in all `controllers/*_controller_test.go` files (~15 files, varies per file)
  - [ ] 2.4: Replace in `controllers/controllertestutils/decoder_test.go` (1 occurrence)
  - [ ] 2.5: Run `go build ./controllers/...` to confirm compilation
- [ ] Task 3: Verify zero remaining occurrences (AC: 3)
  - [ ] 3.1: Run `grep -r 'interface{}' api/ controllers/ --include='*.go' | grep -v zz_generated` and confirm zero matches
- [ ] Task 4: Run `make manifests generate fmt vet test` (AC: 1, 2)
  - [ ] 4.1: Run `make manifests generate` — should produce zero diffs (no `*_types.go` structural changes, only alias rename)
  - [ ] 4.2: Run `make fmt` — should be a no-op (gofmt treats `any` and `interface{}` identically in formatted output)
  - [ ] 4.3: Run `make vet` — should pass
  - [ ] 4.4: Run `make test` — all envtest-based unit tests pass
- [ ] Task 5: Run `make integration` (AC: 2)
  - [ ] 5.1: Run `make integration` — full Kind+Vault integration suite passes

## Dev Notes

### Prerequisites

Stories R1.1 through R1.5 must be completed and merged first. Per the epic ordering: R1.1 → R1.2a → R1.2b → R1.3 → R1.2c → R1.7 → R1.8 → R1.9 → R1.4 → R1.5 → **R1.6**. This story must be the **last** merged — it touches nearly every file and will conflict with anything else in flight.

### Scope

This is a **purely mechanical, zero-behavioral-change** replacement. Go's `any` is a built-in alias for `interface{}` introduced in Go 1.18. The project uses Go 1.22 so this is safe.

**Approximate scale:**
- **~90 non-generated `.go` files** affected across `api/v1alpha1/` and `controllers/`
- **~570 total `interface{}` occurrences** to replace
- No files outside `api/` and `controllers/` contain `interface{}` (`main.go` is clean)
- `zz_generated.deepcopy.go` has zero occurrences — no exclusion needed

### Replacement Strategy

Use `sed` or IDE find-and-replace. The replacement is literal string substitution — no regex complexity needed:

```bash
# Replace in all .go files under api/ (excluding zz_generated for safety)
find api/ -name '*.go' ! -name 'zz_generated*' -exec sed -i 's/interface{}/any/g' {} +

# Replace in all .go files under controllers/
find controllers/ -name '*.go' -exec sed -i 's/interface{}/any/g' {} +
```

**Critical:** The `sed` pattern `interface{}` has no regex special characters, but the braces `{}` are literal in the replacement context. This is a safe global replace because `interface{}` is syntactically unambiguous in Go — it always means the empty interface type.

### Comments Containing `interface{}`

Some files have comments that reference `interface{}` (e.g., `// fromYAML converts a YAML document into a map[string]interface{}.`). These comments should ALSO be updated to `any` for consistency. The `sed` approach handles this automatically.

### Occurrence Patterns (What the Dev Will See)

The `interface{}` occurrences fall into these categories:

1. **`map[string]interface{}`** — Most common. Used in `toMap()`, `GetPayload()`, `IsEquivalentToDesiredState()`, Vault API payloads, and test fixtures. Becomes `map[string]any`.

2. **`[]interface{}`** — Array types used by `toInterfaceArray()` helper and in test fixtures for Vault responses. Becomes `[]any`.

3. **Function signatures** — `func(..., payload map[string]interface{})`, `func ToString(name interface{})`. Parameter and return types. Becomes `any`.

4. **Type assertions** — `data.(map[string]interface{})`, `name.(string)`. The assertion target becomes `data.(map[string]any)`.

5. **Map/slice literals** — `map[string]interface{}{...}`, `[]interface{}{...}`. Becomes `map[string]any{...}`, `[]any{...}`.

6. **Interface definitions** — `GetPayload() map[string]interface{}` in the `VaultObject` interface. Becomes `GetPayload() map[string]any`.

7. **Template function signatures** — In `advanced-funcmap.go`, function literals with `interface{}` params. Becomes `any`.

### Files with Highest Occurrence Counts (Sanity-Check These)

| File | Count | Notes |
|------|-------|-------|
| `api/v1alpha1/isequivalent_audit_test.go` | 100 | Mostly `map[string]any` literals in test fixtures |
| `api/v1alpha1/utils/vaultobject_test.go` | 45 | Test fixtures for VaultObject tests |
| `api/v1alpha1/databasesecretengineconfig_test.go` | 36 | Test fixtures for DB config tests |
| `controllers/vaultresourcecontroller/advanced-funcmap.go` | 31 | Template function signatures and serialization helpers |
| `api/v1alpha1/policy_test.go` | 16 | Policy test fixtures |
| `api/v1alpha1/pkisecretengineconfig_types.go` | 14 | PKI config has many map fields |
| `api/v1alpha1/auditrequestheader_test.go` | 14 | Audit request header test fixtures |
| `api/v1alpha1/identityoidc_test.go` | 12 | OIDC identity test fixtures |
| `api/v1alpha1/rabbitmqsecretengineconfig_types.go` | 11 | RabbitMQ config with multiple map methods |
| `api/v1alpha1/payload_filter_test.go` | 11 | Payload filter test fixtures |
| `api/v1alpha1/rabbitmqsecretenginerole_types.go` | 10 | RabbitMQ role with vhost maps |
| `api/v1alpha1/databasesecretengineconfig_types.go` | 10 | DB config type with connection details |

### What NOT to Touch

- **`zz_generated.deepcopy.go`** — Auto-generated; currently has zero `interface{}` so there's nothing to exclude, but never edit this file manually in any case. `make generate` regenerates it.
- **`main.go`** — No `interface{}` present.
- **Files outside `api/` and `controllers/`** — No `interface{}` found elsewhere.
- **`go.mod` / `go.sum`** — No dependency changes.
- **No `.golangci.yml` config file** — Do NOT create one.
- **`Makefile`, `Dockerfile`, CI configs** — Not Go source, no changes.

### `make manifests generate` Expectations

Running `make manifests generate` after the replacement should produce **zero diffs**:
- `make manifests` runs `controller-gen` which reads kubebuilder markers, not type syntax — `any` vs `interface{}` is irrelevant to CRD generation.
- `make generate` regenerates `zz_generated.deepcopy.go` from the types — `any` and `interface{}` are identical to the Go compiler, so deepcopy output is unchanged.

If any diffs appear, they are likely unrelated to this story (e.g., stale generated files from a prior change). Investigate before committing.

### Edge Case: `interface{}` Inside String Literals

There should be zero occurrences of `interface{}` inside Go string literals (e.g., `fmt.Sprintf("type: %T is interface{}")`) in this codebase. The `sed` replacement is safe. However, after replacement, do a quick visual scan of the diff for any string literal corruption.

### Testing Strategy

- **No new tests needed** — this is a mechanical rename with zero behavioral change
- `any` is a perfect alias for `interface{}` at the language level; the compiler produces identical bytecode
- All existing unit tests (`make test`) and integration tests (`make integration`) exercise the affected code paths
- `make integration` takes ~576-579s
- Kind cluster can degrade — use a fresh cluster if tests fail unexpectedly

### Previous Story Intelligence

**From R1.5 (immediately preceding):**
- Reconciler struct deduplication extracted shared behavior from 4 reconciler types
- `make integration` takes ~576-579s; Kind cluster can degrade with stale namespaces
- `make manifests generate` should be checked even for non-type structural changes
- All code in `controllers/vaultresourcecontroller/` was recently refactored — be aware of any new function signatures that may have been added

**From R1.2c (lint gate):**
- All 21 golangci-lint findings were resolved — codebase is lint-clean
- Any new code must maintain lint compliance

**From R1.1 (context keys, ToString fix):**
- `utils.ToString(name interface{})` was changed to use `fmt.Sprint` for safe conversion — the `interface{}` parameter type in this function will be renamed to `any`
- Context key types and accessors were introduced in `api/v1alpha1/utils/` — these may use `interface{}` in context value methods

### Project Structure Notes

- All changes confined to `api/v1alpha1/` (including `utils/` subdirectory) and `controllers/` (including `vaultresourcecontroller/` and `controllertestutils/` subdirectories)
- No new files created
- No new dependencies
- No external API changes — `any` is source-compatible with `interface{}` everywhere

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.6] — acceptance criteria, task list
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, story ordering (R1.6 is last)
- [Source: _bmad-output/project-context.md#Technology Stack] — Go 1.22 (any alias available since Go 1.18)
- [Source: _bmad-output/project-context.md#Imperative-to-Declarative Bridge] — `toMap()`, `GetPayload()`, `IsEquivalentToDesiredState()` — key `map[string]interface{}` usage
- [Source: api/v1alpha1/utils/vaultobject.go:30-32] — `VaultObject` interface with `GetPayload() map[string]interface{}` and `IsEquivalentToDesiredState(payload map[string]interface{}) bool`
- [Source: api/v1alpha1/payload_filter.go:7-8] — `filterPayloadToDesiredKeys(desiredState, payload map[string]interface{})` function
- [Source: api/v1alpha1/utils/vaultutils.go:26-45] — `write()`, `writeWithResponse()`, `read()` functions with `map[string]interface{}`
- [Source: api/v1alpha1/utils/commons.go:354] — `ToString(name interface{})` function
- [Source: controllers/vaultresourcecontroller/advanced-funcmap.go] — 31 occurrences in template function maps
- [Source: controllers/vaultsecret_controller.go:168] — `formatK8sSecret` with `data interface{}` param
- [Source: _bmad-output/implementation-artifacts/R1-5-reconciler-struct-deduplication.md] — previous story context

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### Change Log

### File List
