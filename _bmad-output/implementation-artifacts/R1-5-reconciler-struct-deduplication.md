# Story R1.5: Reconciler Struct Deduplication

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the duplicated `Reconcile()` and `manageCleanUpLogic()` methods across the 4 reconciler types extracted into a shared skeleton,
So that the finalizer/deletion/outcome flow is defined once and bug fixes apply everywhere.

## Acceptance Criteria

1. **Given** the deletion flow (check timestamp → check finalizer → cleanup → remove finalizer → update) is identical across all 4 types **When** extracted into a shared helper accepting cleanup and reconcile functions **Then** the flow is defined once
2. **Given** `manageCleanUpLogic` is identical except for the endpoint's `DeleteIfExists` call **When** parameterized on a `deleteFunc` **Then** cleanup logic is defined once
3. **Given** all 4 reconciler types use the shared skeleton **When** each type's `Reconcile()` method is compared **Then** only the type-specific `manageReconcileLogic` and endpoint constructor differ
4. **Given** all changes **When** `make test` and `make integration` pass **Then** no regressions

## Tasks / Subtasks

- [x] Task 1: Define a `deleteFunc` type and extract `manageCleanUpLogic` into a shared function (AC: 2)
  - [x] 1.1: Define `type deleteFunc func(ctx context.Context) error` in a new or existing file in `controllers/vaultresourcecontroller/`
  - [x] 1.2: Create a shared `manageCleanUpLogic(ctx context.Context, instance client.Object, deleteFn deleteFunc) error` package-level function that encapsulates the IsDeletable guard + ReconcileSuccessful condition check + deleteFn call
  - [x] 1.3: Each type's `manageCleanUpLogic` becomes a one-liner delegating to the shared function with its endpoint's `DeleteIfExists` method reference
- [x] Task 2: Extract the `Reconcile()` deletion/finalizer/outcome skeleton into a shared function (AC: 1, 3)
  - [x] 2.1: Define `type reconcileFunc func(ctx context.Context, instance client.Object) error` for the type-specific reconcile logic
  - [x] 2.2: Create `ReconcileWithFunctions(ctx context.Context, reconcilerBase *ReconcilerBase, instance client.Object, cleanupFn deleteFunc, reconcileFn reconcileFunc) (ctrl.Result, error)` — the shared skeleton that handles: deletion timestamp check → finalizer guard → cleanup → remove finalizer → update → OR → reconcileFn → ManageOutcome
  - [x] 2.3: Ensure the shared skeleton uses the same log messages as the current `VaultResource.Reconcile()` (the canonical version without debug extras)
- [x] Task 3: Refactor all 4 types to use the shared skeleton (AC: 1, 3)
  - [x] 3.1: Refactor `VaultResource.Reconcile()` to call `ReconcileWithFunctions`
  - [x] 3.2: Refactor `VaultEngineResource.Reconcile()` to call `ReconcileWithFunctions`
  - [x] 3.3: Refactor `VaultAuditResource.Reconcile()` to call `ReconcileWithFunctions`
  - [x] 3.4: Refactor `VaultPKIEngineResource.Reconcile()` to call `ReconcileWithFunctions`
- [x] Task 4: Remove dead code and verify compilation (AC: 3)
  - [x] 4.1: Remove `VaultPKIEngineResource`'s extra debug log lines that differ from the canonical pattern (the `"Delete"`, `"Finaliter?"`, `"RemoveFinalizer"`, `"DeleteIfExists"` Info logs)
  - [x] 4.2: Normalize `VaultAuditResource`'s log message from `"starting audit reconcile cycle"` to `"starting reconcile cycle"` (now handled by shared skeleton)
  - [x] 4.3: Run `go build ./...` to verify compilation
- [x] Task 5: Run `make test` and `make integration` (AC: 4)
  - [x] 5.1: Run `make manifests generate fmt vet test`
  - [x] 5.2: Run `make integration`

## Dev Notes

### Prerequisites

Stories R1.1 through R1.4 must be completed and merged first. Per the epic ordering: R1.1 → R1.2a → R1.2b → R1.3 → R1.2c → R1.7 → R1.8 → R1.9 → R1.4 → **R1.5** → R1.6.

### Scope

**Files to modify:** `controllers/vaultresourcecontroller/` — 4 reconciler files:
- `vaultresourcereconciler.go` (117 lines)
- `vaultengineresourcereconciler.go` (135 lines)
- `vaultauditresourcereconciler.go` (121 lines)
- `vaultpkiengineresourcereconciler.go` (152 lines)

A new file MAY be created (e.g., `reconcile_skeleton.go`) or the shared functions can be added to an existing file. Choose whichever is cleaner — placing in a new file is preferred for reviewability.

**Files NOT to touch:**
- `utils.go` — `ManageOutcome`, `ReconcilerBase`, constants stay as-is
- Any controller files (`*_controller.go`) — they call `NewVaultResource(...)` / `vaultResource.Reconcile(ctx1, instance)` which is unchanged
- Any `*_types.go` files
- Any test files — no test changes needed
- `main.go`

### Current Duplication Analysis

All 4 types share this identical skeleton in `Reconcile()`:

```go
func (r *Type) Reconcile(ctx context.Context, instance client.Object) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    log.Info("starting reconcile cycle")
    log.V(1).Info("reconcile", "instance", instance)
    if !instance.GetDeletionTimestamp().IsZero() {
        if !controllerutil.ContainsFinalizer(instance, vaultutils.GetFinalizer(instance)) {
            return reconcile.Result{}, nil
        }
        err := r.manageCleanUpLogic(ctx, instance)
        if err != nil {
            log.Error(err, "unable to delete instance", "instance", instance)
            return ManageOutcome(ctx, *r.reconcilerBase, instance, err)
        }
        controllerutil.RemoveFinalizer(instance, vaultutils.GetFinalizer(instance))
        err = r.reconcilerBase.GetClient().Update(ctx, instance)
        if err != nil {
            log.Error(err, "unable to update instance", "instance", instance)
            return ManageOutcome(ctx, *r.reconcilerBase, instance, err)
        }
        return reconcile.Result{}, nil
    }
    err := r.manageReconcileLogic(ctx, instance)
    if err != nil {
        log.Error(err, "unable to complete reconcile logic", "instance", instance)
        return ManageOutcome(ctx, *r.reconcilerBase, instance, err)
    }
    return ManageOutcome(ctx, *r.reconcilerBase, instance, err)
}
```

And all 4 share this cleanup pattern in `manageCleanUpLogic()`:

```go
func (r *Type) manageCleanUpLogic(context context.Context, instance client.Object) error {
    if vaultObject, ok := instance.(vaultutils.VaultObject); ok {
        if !vaultObject.IsDeletable() {
            return nil
        }
    }
    if conditionAware, ok := instance.(vaultutils.ConditionsAware); ok {
        for _, condition := range conditionAware.GetConditions() {
            if condition.Status == metav1.ConditionTrue && condition.Type == ReconcileSuccessful {
                err := r.<endpoint>.DeleteIfExists(context)
                if err != nil {
                    log.Error(err, "unable to delete vault resource", "instance", instance)
                    return err
                }
            }
        }
    }
    return nil
}
```

The ONLY differences across the 4 types are:
1. **`manageCleanUpLogic`**: which endpoint's `DeleteIfExists` is called (`r.vaultEndpoint`, `r.vaultEngineEndpoint`, `r.vaultAuditEndpoint`, `r.vaultPKIEngineEndpoint`)
2. **`manageReconcileLogic`**: completely type-specific business logic (standard CreateOrUpdate, engine Create-then-Tune, audit enable/update, PKI generate/sign pipeline)
3. **Minor log message variations**: `VaultAuditResource` says `"starting audit reconcile cycle"` instead of `"starting reconcile cycle"`, `VaultPKIEngineResource` has extra debug `.Info()` calls

### Recommended Design

**Option A (Function-based — recommended):**

```go
// deleteFunc is the type-specific vault cleanup function
type deleteFunc func(ctx context.Context) error

// reconcileFunc is the type-specific reconcile logic
type reconcileFunc func(ctx context.Context, instance client.Object) error

// manageCleanUpLogic is the shared cleanup skeleton
func manageCleanUpLogic(ctx context.Context, instance client.Object, deleteFn deleteFunc) error {
    log := log.FromContext(ctx)
    if vaultObject, ok := instance.(vaultutils.VaultObject); ok {
        if !vaultObject.IsDeletable() {
            return nil
        }
    }
    if conditionAware, ok := instance.(vaultutils.ConditionsAware); ok {
        for _, condition := range conditionAware.GetConditions() {
            if condition.Status == metav1.ConditionTrue && condition.Type == ReconcileSuccessful {
                err := deleteFn(ctx)
                if err != nil {
                    log.Error(err, "unable to delete vault resource", "instance", instance)
                    return err
                }
            }
        }
    }
    return nil
}

// ReconcileWithFunctions is the shared reconcile skeleton
func ReconcileWithFunctions(ctx context.Context, reconcilerBase *ReconcilerBase, instance client.Object, cleanupFn deleteFunc, reconcileFn reconcileFunc) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    log.Info("starting reconcile cycle")
    log.V(1).Info("reconcile", "instance", instance)
    if !instance.GetDeletionTimestamp().IsZero() {
        if !controllerutil.ContainsFinalizer(instance, vaultutils.GetFinalizer(instance)) {
            return reconcile.Result{}, nil
        }
        err := manageCleanUpLogic(ctx, instance, cleanupFn)
        if err != nil {
            log.Error(err, "unable to delete instance", "instance", instance)
            return ManageOutcome(ctx, *reconcilerBase, instance, err)
        }
        controllerutil.RemoveFinalizer(instance, vaultutils.GetFinalizer(instance))
        err = reconcilerBase.GetClient().Update(ctx, instance)
        if err != nil {
            log.Error(err, "unable to update instance", "instance", instance)
            return ManageOutcome(ctx, *reconcilerBase, instance, err)
        }
        return reconcile.Result{}, nil
    }
    err := reconcileFn(ctx, instance)
    if err != nil {
        log.Error(err, "unable to complete reconcile logic", "instance", instance)
        return ManageOutcome(ctx, *reconcilerBase, instance, err)
    }
    return ManageOutcome(ctx, *reconcilerBase, instance, err)
}
```

**Each type's `Reconcile()` then becomes a thin wrapper:**

```go
func (r *VaultResource) Reconcile(ctx context.Context, instance client.Object) (ctrl.Result, error) {
    return ReconcileWithFunctions(ctx, r.reconcilerBase, instance,
        r.vaultEndpoint.DeleteIfExists,
        r.manageReconcileLogic,
    )
}
```

**Why NOT an interface-based approach:** Go generics/interfaces would require the 4 endpoint types to implement a common interface (`DeleteIfExists(ctx) error`). They already do this implicitly, but they are struct values from different packages. The function-based approach is simpler, requires zero interface additions, and the method-reference syntax (`r.vaultEndpoint.DeleteIfExists`) works naturally since all 4 endpoint types have `DeleteIfExists(context.Context) error`.

### Critical: PKI Extra Logs and Audit Log Difference

- `VaultPKIEngineResource.Reconcile()` currently has extra `log.Info("Delete", ...)`, `log.Info("Finaliter?", ...)`, and `log.Info("RemoveFinalizer", ...)` calls that the other 3 types do NOT have. These appear to be leftover debug logging (note the typo "Finaliter"). **Remove them** — the shared skeleton logs the same way for all types.
- `VaultPKIEngineResource.manageCleanUpLogic()` has an extra `log.Info("DeleteIfExists", ...)` before the delete call. **Remove it** — not present in the other 3 types.
- `VaultAuditResource.Reconcile()` logs `"starting audit reconcile cycle"` instead of `"starting reconcile cycle"`. **Normalize to the standard message** — the shared skeleton uses one log line.
- `VaultAuditResource.manageCleanUpLogic()` uses error message `"unable to disable audit device"` instead of `"unable to delete vault resource"`. **Normalize** — the shared `manageCleanUpLogic` function uses the standard message.

These are cosmetic differences from copy-paste drift. The shared skeleton normalizes behavior.

### Endpoint DeleteIfExists Signatures

All 4 endpoint types have the same `DeleteIfExists` signature:

| Endpoint Type | Method Signature |
|---------------|-----------------|
| `*vaultutils.VaultEndpoint` | `func (ve *VaultEndpoint) DeleteIfExists(context context.Context) error` |
| `*vaultutils.VaultEngineEndpoint` | `func (ve *VaultEngineEndpoint) DeleteIfExists(context context.Context) error` |
| `*vaultutils.VaultAuditEndpoint` | `func (ve *VaultAuditEndpoint) DeleteIfExists(context context.Context) error` |
| `*vaultutils.VaultPKIEngineEndpoint` | `func (ve *VaultPKIEngineEndpoint) DeleteIfExists(context context.Context) error` |

All match `func(context.Context) error` — they can all be passed directly as `deleteFunc` via method reference (e.g., `r.vaultEndpoint.DeleteIfExists`).

### Callers of Each Type (DO NOT CHANGE)

The controller files that call these types remain unchanged:
- `NewVaultResource`: ~25 controller files (policy, groups, roles, configs, identity, etc.)
- `NewVaultEngineResource`: `secretenginemount_controller.go`, `authenginemount_controller.go`
- `NewVaultAuditResource`: `audit_controller.go`
- `NewVaultPKIEngineResource`: `pkisecretengineconfig_controller.go`

All callers use the same pattern: `vaultResource.Reconcile(ctx1, instance)`. The `Reconcile()` method signature stays identical — callers never notice the internal refactoring.

### What NOT to Touch

- Do NOT modify `utils.go` — `ManageOutcome`, `ManageOutcomeWithRequeue`, `ReconcilerBase`, constants, predicates stay as-is
- Do NOT modify any `*_controller.go` files — the external API (`NewVaultXxx` + `.Reconcile()`) is unchanged
- Do NOT modify `manageReconcileLogic` business logic — only change how it's called (via function reference)
- Do NOT introduce new interfaces in `api/v1alpha1/utils/` — use function types
- Do NOT modify any `*_types.go` files
- Do NOT modify any test files
- Do NOT create a `.golangci.yml` config file
- Do NOT modify `main.go`
- Do NOT touch `dynamicclientutils.go` or `advanced-funcmap.go`

### Expected Outcome

After refactoring:
- The shared `ReconcileWithFunctions` + `manageCleanUpLogic` live in one place (new file or section)
- Each of the 4 type files retains: struct definition, constructor, `Reconcile()` one-liner, and `manageReconcileLogic()`
- Total code reduction: ~60-80 lines eliminated across the 4 files (the duplicated skeleton is ~25 lines × 4 = 100 lines, replaced by 4 one-liners + 1 shared implementation of ~30 lines)
- No behavioral change — all existing tests pass without modification

### Testing Strategy

- **No new tests needed** — this is a structural refactoring with zero behavioral change
- The existing integration tests exercise all 4 reconciler paths (create, update, delete lifecycle)
- `make test` exercises the envtest suite which tests reconciliation flow
- `make integration` exercises the full Kind+Vault path for all CRD types
- `make manifests generate` should produce zero diffs — no `*_types.go` or RBAC marker changes

### Kind Cluster Considerations

- `make integration` takes ~576-579s
- Kind cluster can degrade (terminating namespaces) — if tests fail unexpectedly, try a fresh cluster with `make integration` (which recreates the cluster)
- Run `make manifests generate` even for non-type changes — catches unexpected diffs

### Previous Story Intelligence

**From R1.4 (immediately preceding):**
- Decoder generics rewrite established the pattern of extracting shared behavior from duplicated code
- Go 1.22 generics are available, but this story uses function types (simpler, no generic needed)
- `make integration` takes ~576-579s
- Kind cluster can degrade — fresh cluster if tests fail unexpectedly

**From R1.2c (lint gate):**
- All 21 golangci-lint findings were resolved — the codebase should be lint-clean
- `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...` exits 0
- Any new code must maintain lint compliance

**From R1.1 (context key migration):**
- Large mechanical refactors across multiple files should verify compilation after each step
- Verify no regressions between tasks

### Project Structure Notes

- All changes confined to `controllers/vaultresourcecontroller/` package
- One new file likely added (`reconcile_skeleton.go` or similar)
- No new dependencies added
- No external API changes — callers see the same `Reconcile(ctx, instance)` method
- Existing struct types and constructors preserved (backward compatible)

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.5] — acceptance criteria, task list
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, story ordering (R1.5 follows R1.4)
- [Source: _bmad-output/project-context.md#Three Reconciler Variants] — reconciler type descriptions
- [Source: _bmad-output/project-context.md#Finalizer Management] — finalizer lifecycle
- [Source: controllers/vaultresourcecontroller/vaultresourcereconciler.go] — canonical VaultResource implementation (117 lines)
- [Source: controllers/vaultresourcecontroller/vaultengineresourcereconciler.go] — VaultEngineResource (135 lines)
- [Source: controllers/vaultresourcecontroller/vaultauditresourcereconciler.go] — VaultAuditResource (121 lines)
- [Source: controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go] — VaultPKIEngineResource (152 lines)
- [Source: controllers/vaultresourcecontroller/utils.go:43-189] — ManageOutcome, ReconcilerBase, constants
- [Source: api/v1alpha1/utils/vaultobject.go:74] — VaultEndpoint.DeleteIfExists signature
- [Source: api/v1alpha1/utils/vaultengineobject.go:36] — VaultEngineEndpoint struct (embeds VaultEndpoint)
- [Source: api/v1alpha1/utils/vaultauditobject.go:186] — VaultAuditEndpoint.DeleteIfExists
- [Source: api/v1alpha1/utils/vautlpkiengineobject.go:63] — VaultPKIEngineEndpoint.DeleteIfExists
- [Source: _bmad-output/implementation-artifacts/R1-4-test-decoder-generics-rewrite.md] — previous story context

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Initial integration test failure was infrastructure-related (Kind cluster degraded, Vault Helm chart timed out). Resolved by deleting and recreating the Kind cluster.

### Completion Notes List

- Created `reconcile_skeleton.go` with shared `deleteFunc` type, `reconcileFunc` type, `manageCleanUpLogic` function, and `ReconcileWithFunctions` function
- Refactored all 4 reconciler types (`VaultResource`, `VaultEngineResource`, `VaultAuditResource`, `VaultPKIEngineResource`) to delegate to `ReconcileWithFunctions`
- Removed duplicated `manageCleanUpLogic` methods from all 4 types (each was ~20 lines, now eliminated)
- Removed duplicated `Reconcile()` skeleton from all 4 types (each was ~25 lines, now a 3-line wrapper)
- Removed PKI-specific debug log lines (`"processing deletion"`, `"no finalizer found, skipping cleanup"`, `"removing finalizer"`, `"deleting vault resource if exists"`)
- Normalized VaultAuditResource log message from `"starting audit reconcile cycle"` to standard `"starting reconcile cycle"` (via shared skeleton)
- Normalized VaultAuditResource error message from `"unable to disable audit device"` to standard `"unable to delete vault resource"` (via shared `manageCleanUpLogic`)
- `make manifests generate fmt vet test` passes with zero diffs and all tests green
- `make integration` passes (576s, consistent with baseline)

### Change Log

- 2026-06-20: Story R1.5 implemented — reconciler struct deduplication via shared skeleton functions

### File List

- controllers/vaultresourcecontroller/reconcile_skeleton.go (new)
- controllers/vaultresourcecontroller/vaultresourcereconciler.go (modified)
- controllers/vaultresourcecontroller/vaultengineresourcereconciler.go (modified)
- controllers/vaultresourcecontroller/vaultauditresourcereconciler.go (modified)
- controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go (modified)
