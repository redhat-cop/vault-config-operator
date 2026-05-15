# Story R1.4: Test Decoder Generics Rewrite

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the 30+ identical `Get<Type>Instance` methods in `controllertestutils/decoder.go` replaced with a single generic function,
So that adding a new CRD type no longer requires copying another decode method.

## Acceptance Criteria

1. **Given** a generic decode function `DecodeInstance[T runtime.Object]` **When** called with any CRD type and a valid YAML file **Then** it returns the typed instance
2. **Given** a YAML file with a mismatched kind **When** `DecodeInstance` is called **Then** it returns `errDecode`
3. **Given** all test call sites migrated from `d.GetPolicyInstance(f)` to `DecodeInstance[*Policy](d, f)` (or equivalent) **When** `make test` is run **Then** all existing tests pass
4. **Given** the rewrite **When** line count is compared **Then** ~500 lines of duplicate code are eliminated

## Tasks / Subtasks

- [ ] Task 1: Refactor `decodeFile` and implement generic `DecodeInstance[T]` in `decoder.go` (AC: 1, 2)
  - [ ] 1.1: Convert `decodeFile` from a method `(d *decoder) decodeFile(filename)` to a standalone function `decodeFile(filename)` — it only uses the package-level `runtimeDecoder` variable, not any `decoder` fields
  - [ ] 1.2: Fix the `ioutil.ReadFile` → `os.ReadFile` in `decodeFile` if R1.3 has not yet been merged (line 60); remove `"io/ioutil"` import
  - [ ] 1.3: Add a top-level generic function `func DecodeInstance[T runtime.Object](filename string) (T, error)` that calls `decodeFile`, checks `gvk.Kind` against `reflect.TypeOf(*new(T)).Elem().Name()`, and type-asserts the result
  - [ ] 1.4: Ensure the type constraint uses `runtime.Object` (from `k8s.io/apimachinery/pkg/runtime`) — all CRD types implement this via generated `DeepCopyObject()`
- [ ] Task 2: Remove all 34 `Get<Type>Instance` methods (AC: 4)
  - [ ] 2.1: Delete all methods from line 67 through line 577
  - [ ] 2.2: Verify the file is reduced to ~55 lines (init, NewDecoder, CreateFromYAML, decodeFile, DecodeInstance)
- [ ] Task 3: Migrate all test call sites and add imports (AC: 3)
  - [ ] 3.1: Add `"github.com/redhat-cop/vault-config-operator/controllers/controllertestutils"` import to `databasesecretenginestaticrole_controller_test.go`, `pkisecretengine_controller_test.go`, and `driftdetection_controller_test.go`
  - [ ] 3.2: Migrate 8 calls in `databasesecretenginestaticrole_controller_test.go`
  - [ ] 3.3: Migrate 5 calls in `pkisecretengine_controller_test.go`
  - [ ] 3.4: Migrate 1 call in `driftdetection_controller_test.go`
- [ ] Task 4: Run `make test` — all unit and envtest tests pass (AC: 3)
- [ ] Task 5: Run `make integration` — all integration tests pass (AC: 3)

## Dev Notes

### Prerequisites

Stories R1.1, R1.2a, R1.2b, R1.3, R1.2c, R1.7, R1.8, and R1.9 must be completed and merged first. Per the epic ordering: R1.1 → R1.2a → R1.2b → R1.3 → R1.2c → R1.7 → R1.8 → R1.9 → **R1.4** → R1.5 → R1.6.

R1.3 modifies `decoder.go` (replaces `ioutil.ReadFile` with `os.ReadFile` on line 60 and adds `panic(err)` for `AddToScheme` from R1.2a). Verify the `decodeFile` function uses `os.ReadFile` before starting. If R1.3 has not been merged yet, include that fix in this story.

### Go Generics Availability

Go 1.22 (the project's version per `go.mod`) fully supports generics (available since Go 1.18). No language version change needed.

### Generic Function Design

The generic function replaces all 34 identical `Get<Type>Instance` methods. Since `decoder` is an empty struct (no fields) and `decodeFile` only uses the package-level `runtimeDecoder`, the generic function does NOT need a `decoder` receiver or parameter. Here is the exact implementation:

```go
func DecodeInstance[T runtime.Object](filename string) (T, error) {
	obj, gvk, err := decodeFile(filename)
	if err != nil {
		var zero T
		return zero, err
	}

	expected := reflect.TypeOf(*new(T)).Elem().Name()
	if gvk.Kind != expected {
		var zero T
		return zero, errDecode
	}

	typed, ok := obj.(T)
	if !ok {
		var zero T
		return zero, errDecode
	}
	return typed, nil
}
```

As part of this change, `decodeFile` should be converted from a method to a standalone function since it doesn't use any `decoder` fields:

```go
func decodeFile(filename string) (runtime.Object, *schema.GroupVersionKind, error) {
	stream, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	return runtimeDecoder.Decode(stream, nil, nil)
}
```

**Key design decisions:**

1. **Top-level function, not a method** — Go generics do not support type parameters on methods (only on functions and types). `DecodeInstance` must be a package-level function. Since `decoder` is an empty struct with no state, the function takes only the filename parameter.

2. **`decodeFile` becomes a standalone function** — it only uses the package-level `runtimeDecoder` variable, not any `decoder` struct fields. Converting it simplifies the code and makes it callable from `DecodeInstance` without needing a `decoder` instance. `CreateFromYAML` remains a method on `decoder` since it serves a different purpose (unstructured creation via API server).

3. **Type constraint is `runtime.Object`** — this is the interface all CRD types implement (via generated `DeepCopyObject()` in `zz_generated.deepcopy.go`). The constraint is from `k8s.io/apimachinery/pkg/runtime`.

4. **`var zero T` for error returns** — Go generics require explicitly declaring a zero-value variable for typed nil returns. `return nil, err` does not compile when `T` is a type parameter.

5. **Kind name extraction via `reflect.TypeOf(*new(T)).Elem().Name()`** — this gets the struct name from the pointer type parameter. For `T = *redhatcopv1alpha1.Policy`, `*new(T)` is `*Policy` (nil pointer), `.Elem()` dereferences to `Policy`, `.Name()` returns `"Policy"`. This matches the existing pattern: `reflect.TypeOf(redhatcopv1alpha1.Policy{}).Name()`.

6. **The `decoder` struct stays unexported** — it is still used by `CreateFromYAML` and `NewDecoder()`. The `DecodeInstance` function does not reference it.

### Call Site Migration Pattern

Every existing call site follows the same transformation:

**Before (method call):**
```go
instance, err := decoder.GetPolicyInstance("../test/some-policy.yaml")
```

**After (generic function call):**
```go
instance, err := controllertestutils.DecodeInstance[*redhatcopv1alpha1.Policy]("../test/some-policy.yaml")
```

**Import requirement:** All 3 affected test files are in `package controllers`. They currently access `decoder` via a shared package-level variable in `suite_integration_test.go` and do NOT import `controllertestutils` themselves. Each must add:

```go
"github.com/redhat-cop/vault-config-operator/controllers/controllertestutils"
```

The `redhatcopv1alpha1` import is already present in all 3 files.

### Complete Call Site Inventory (14 calls across 3 files)

All remaining `Get<Type>Instance` calls are for **delete operations** or **mutation-before-create** scenarios where `CreateFromYAML` is not suitable (the newer `CreateFromYAML` pattern uses unstructured objects).

**`controllers/databasesecretenginestaticrole_controller_test.go` (8 calls) — add `controllertestutils` import:**

| Line | Old Call | New Call |
|------|----------|----------|
| 322 | `decoder.GetDatabaseSecretEngineStaticRoleInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.DatabaseSecretEngineStaticRole](...)` |
| 342 | `decoder.GetDatabaseSecretEngineConfigInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.DatabaseSecretEngineConfig](...)` |
| 361 | `decoder.GetRandomSecretInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.RandomSecret](...)` |
| 381 | `decoder.GetSecretEngineMountInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.SecretEngineMount](...)` |
| 400 | `decoder.GetPasswordPolicyInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.PasswordPolicy](...)` |
| 420 | `decoder.GetSecretEngineMountInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.SecretEngineMount](...)` |
| 440 | `decoder.GetKubernetesAuthEngineRoleInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.KubernetesAuthEngineRole](...)` |
| 458 | `decoder.GetPolicyInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.Policy](...)` |

**`controllers/pkisecretengine_controller_test.go` (5 calls) — add `controllertestutils` import:**

| Line | Old Call | New Call |
|------|----------|----------|
| 312 | `decoder.GetPKISecretEngineRoleInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.PKISecretEngineRole](...)` |
| 331 | `decoder.GetPKISecretEngineConfigInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.PKISecretEngineConfig](...)` |
| 351 | `decoder.GetSecretEngineMountInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.SecretEngineMount](...)` |
| 371 | `decoder.GetKubernetesAuthEngineRoleInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.KubernetesAuthEngineRole](...)` |
| 390 | `decoder.GetPolicyInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.Policy](...)` |

**`controllers/driftdetection_controller_test.go` (1 call) — add `controllertestutils` import:**

| Line | Old Call | New Call |
|------|----------|----------|
| 286 | `decoder.GetPolicyInstance(...)` | `controllertestutils.DecodeInstance[*redhatcopv1alpha1.Policy](...)` |

This last call is the **mutation-before-create** exception case: the drift-detection "disabled" test decodes the fixture, overrides `Name` to `"test-drift-policy-disabled"`, sets namespace, and creates via typed `k8sIntegrationClient.Create`. This pattern is explicitly preserved by the project context docs.

### Why Not All Tests Use `Get<Type>Instance`

Most integration tests (24 files, ~99 calls total) use the newer `CreateFromYAML` pattern which creates via unstructured objects and lets server-side defaulting apply. The `Get<Type>Instance` methods are only used for:

1. **Delete operations** — need a typed object for `k8sIntegrationClient.Delete()` (requires the GVK to be set, which the typed object provides)
2. **Mutation-before-create** — the drift-detection "disabled" test overrides `metadata.name` before creating

After this rewrite, both use cases are served by `DecodeInstance[T]`.

### Unused Methods After Migration

After migrating the 14 call sites above, the remaining 24 `Get<Type>Instance` methods (for types like `VaultSecret`, `Entity`, `EntityAlias`, `Group`, `GroupAlias`, `LDAPAuthEngineConfig`, `LDAPAuthEngineGroup`, `JWTOIDCAuthEngineConfig`, `JWTOIDCAuthEngineRole`, `AuthEngineMount`, `RabbitMQSecretEngineConfig`, `RabbitMQSecretEngineRole`, `KubernetesSecretEngineConfig`, `KubernetesSecretEngineRole`, `IdentityOIDCScope`, `IdentityOIDCAssignment`, `IdentityOIDCClient`, `IdentityOIDCProvider`, `IdentityTokenConfig`, `IdentityTokenKey`, `IdentityTokenRole`, `Audit`, `AuditRequestHeader`) have **zero callers** — their tests were migrated to `CreateFromYAML` in earlier epics. Delete them all.

### Import Changes in Test Files

The 3 affected test files (`databasesecretenginestaticrole_controller_test.go`, `pkisecretengine_controller_test.go`, `driftdetection_controller_test.go`) are all in `package controllers`. They currently access `decoder` through a shared package-level variable declared in `suite_integration_test.go`:

```go
var decoder = controllertestutils.NewDecoder()
```

The affected test files do **NOT** currently import `controllertestutils` — only `suite_integration_test.go` does. Each of the 3 files must add this import:

```go
"github.com/redhat-cop/vault-config-operator/controllers/controllertestutils"
```

The `redhatcopv1alpha1` import is already present in all 3 files.

After migration, the `decoder` variable in `suite_integration_test.go` is only used for `CreateFromYAML` calls (which remain as `decoder.CreateFromYAML(...)` method calls). The `suite_integration_test.go` file itself does NOT change.

### What NOT to Touch

- Do NOT modify `CreateFromYAML` — it uses unstructured objects and is a different code path
- Do NOT modify `init()` or `NewDecoder()` — they are unchanged
- Do NOT modify any `*_types.go` files — no CRD changes
- Do NOT modify any controller files — no reconciler changes
- Do NOT modify `main.go`
- Do NOT create a `.golangci.yml` config file
- Do NOT change the `decodeFile` method (except the `ioutil` → `os` fix if R1.3 is not merged)
- Do NOT modify test fixtures (YAML files in `test/`)
- Do NOT change test logic — only the decoder call syntax changes

### Expected Final `decoder.go` (~55 lines)

After the rewrite, the file should contain only:
1. Package declaration and imports (~15 lines)
2. `decoder` struct type, `runtimeDecoder` var, and `errDecode` sentinel (~5 lines)
3. `init()` function (~5 lines)
4. `NewDecoder()` function (~3 lines)
5. `CreateFromYAML()` method on `decoder` (~17 lines)
6. `decodeFile()` standalone function (~6 lines)
7. `DecodeInstance[T]()` generic function (~15 lines)

Total: ~55 lines, down from 577. **~520 lines of duplicate code eliminated.**

Note: `CreateFromYAML` remains a method on `decoder` (it uses `client.Client` and `context.Context` — a different code path from the typed decode). The `decoder` struct is still needed for `NewDecoder()` and `CreateFromYAML`.

### Testing Strategy

- No new tests needed — this is a structural refactoring story
- Existing unit tests (`make test`) exercise the decoder indirectly via envtest suite setup
- Existing integration tests (`make integration`) exercise all `Get<Type>Instance` call sites being migrated
- The generic function has identical behavior to the typed methods — same `decodeFile` call, same `reflect.TypeOf` kind check, same type assertion, same `errDecode` return
- `make manifests generate` should produce zero diffs — no `*_types.go` Spec/Status changes or RBAC marker changes

### Kind Cluster Considerations

- `make integration` takes ~576-579s (from R1.2a/R1.2b/R1.3 experience)
- Kind cluster can degrade (terminating namespaces) — if tests fail unexpectedly, try a fresh cluster with `make integration` (which recreates the cluster)
- Run `make manifests generate` even for non-type changes — catches unexpected diffs

### Previous Story Intelligence

**From R1.3 (immediately preceding in the structural refactoring sequence):**
- `decoder.go` was modified: `ioutil.ReadFile` → `os.ReadFile` on line 60, `AddToScheme` error check from R1.2a
- `make integration` takes ~576-579s
- Kind cluster can degrade — fresh cluster if tests fail unexpectedly
- `go mod tidy` should be run after changes to ensure clean dependency graph

**From R1.2c (lint gate, preceding R1.7/R1.8/R1.9):**
- All 21 golangci-lint findings were resolved — the codebase should be lint-clean
- `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...` exits 0

**From R1.1:**
- Context key migration across 18+ files established the pattern of large mechanical refactors
- Verify compilation after each task before moving to the next

### Project Structure Notes

- Only `controllers/controllertestutils/decoder.go` and 3 test files are modified
- No new files created
- No new dependencies added
- All changes are within existing code organization boundaries
- The `DecodeInstance` function stays in `controllers/controllertestutils/decoder.go` where all decoder functions live

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.4] — acceptance criteria, task list, scope
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, story ordering (R1.4 is 9th: after R1.2c, R1.7, R1.8, R1.9)
- [Source: _bmad-output/project-context.md#Technology Stack & Versions] — Go 1.22.0 (generics available since 1.18)
- [Source: _bmad-output/project-context.md#Testing Rules] — decoder usage patterns, `CreateFromYAML` vs typed decode
- [Source: _bmad-output/project-context.md#Integration Test Pattern] — fixture creation, typed references after creation
- [Source: controllers/controllertestutils/decoder.go] — current 577-line file with 34 identical methods
- [Source: controllers/databasesecretenginestaticrole_controller_test.go:322-458] — 8 typed decode calls for delete operations
- [Source: controllers/pkisecretengine_controller_test.go:312-390] — 5 typed decode calls for delete operations
- [Source: controllers/driftdetection_controller_test.go:286] — 1 typed decode call for mutation-before-create
- [Source: _bmad-output/implementation-artifacts/R1-3-dependency-modernization-drop-deprecated-replace-handrolled.md] — previous story context
- [Source: _bmad-output/implementation-artifacts/R1-2c-lint-green-gate-verify-full-compliance.md] — lint gate story

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### Change Log

### File List
