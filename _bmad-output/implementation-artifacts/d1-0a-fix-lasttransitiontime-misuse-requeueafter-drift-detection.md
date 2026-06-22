# Story D1.0a: Fix LastTransitionTime Misuse — Migrate Drift Detection to RequeueAfter

Status: ready-for-dev

## Story

As an operator developer,
I want the `LastTransitionTime` force-override removed and drift detection timing moved to `RequeueAfter`,
So that the operator follows standard Kubernetes condition API conventions and reduces unnecessary etcd write pressure.

## Acceptance Criteria

1. **Given** `ManageOutcomeWithRequeue` currently force-overrides `LastTransitionTime` after `apimeta.SetStatusCondition` (lines 157-164 of `utils.go`) **When** the force-override is removed **Then** `apimeta.SetStatusCondition` operates with standard K8s semantics — `LastTransitionTime` only updates when `Status` actually changes
2. **Given** `ManageOutcome` currently passes `requeueAfter: 0` **When** drift detection is enabled (`ENABLE_DRIFT_DETECTION=true`) and the reconcile succeeds (`issue == nil`) **Then** `ManageOutcome` passes `requeueAfter: SyncPeriod` to `ManageOutcomeWithRequeue`, causing the controller-runtime work queue to schedule the next drift-detection reconcile automatically
3. **Given** `PeriodicReconcilePredicate.Update()` currently contains timestamp-based periodic reconciliation logic **When** simplified to generation-only filtering **Then** the predicate only reconciles on `generation` change; drift detection is handled entirely by `RequeueAfter` via the work queue
4. **Given** all changes **When** `make manifests generate fmt vet test` is run **Then** zero diffs from `manifests generate`, all unit tests pass
5. **Given** all changes **When** `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...` is run with v1.64.8 **Then** zero findings (lint-clean baseline maintained)

## Tasks / Subtasks

- [ ] Task 1: Remove `LastTransitionTime` force-override in `ManageOutcomeWithRequeue` (AC: 1)
  - [ ] 1.1: In `controllers/vaultresourcecontroller/utils.go`, delete lines 157-164 (the `// apimeta.SetStatusCondition only updates...` comment + `for i := range conditions` loop)
  - [ ] 1.2: Verify compilation: `go build ./...`
- [ ] Task 2: Add `RequeueAfter` for drift detection in `ManageOutcome` (AC: 2)
  - [ ] 2.1: Modify `ManageOutcome` to calculate `requeueAfter` conditionally: `SyncPeriod` when `issue == nil && IsDriftDetectionEnabled()`, else `0`
  - [ ] 2.2: Verify no other code calls `ManageOutcomeWithRequeue` directly (search for all call sites — expected: only `ManageOutcome`)
  - [ ] 2.3: Verify compilation: `go build ./...`
- [ ] Task 3: Simplify `PeriodicReconcilePredicate.Update()` to generation-only filtering (AC: 3)
  - [ ] 3.1: Replace the `Update` method body: keep only the nil-object check and generation-change check; remove the `IsDriftDetectionEnabled()` check, the `ConditionsAware` type assertion, the `LastTransitionTime` comparison, and the `SyncPeriod` elapsed check
  - [ ] 3.2: Update doc comments on `PeriodicReconcilePredicate` struct and `Update` method to reflect that it is now generation-only (drift detection uses `RequeueAfter`)
  - [ ] 3.3: Keep struct fields and constructors unchanged (exported API — backward compatible)
  - [ ] 3.4: Verify compilation: `go build ./...`
- [ ] Task 4: Update unit tests in `utils_test.go` (AC: 4)
  - [ ] 4.1: Update `TestPeriodicReconcilePredicate_Update` — predicate no longer does time-based checks; tests that expected `true` for elapsed-interval-with-drift-enabled should now expect `false`; remove drift-detection-specific test cases or convert them to assert `false`
  - [ ] 4.2: Keep `TestIsDriftDetectionEnabled` unchanged — the function is still used by `ManageOutcome`
  - [ ] 4.3: Run `make test` — all unit tests pass
- [ ] Task 5: Verify no regressions (AC: 4, 5)
  - [ ] 5.1: Run `make manifests generate fmt vet test` — zero diffs, all tests pass
  - [ ] 5.2: Run `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...` — exit 0, zero findings

## Dev Notes

### Background: Why This Fix Exists

In Story R1.3 (dependency modernization), `AddOrReplaceCondition` was migrated to `apimeta.SetStatusCondition`. The stdlib function has smarter `LastTransitionTime` handling — it only updates the timestamp when `Status` actually changes. The `PeriodicReconcilePredicate` was reading `ReconcileSuccessful.LastTransitionTime` as a "when was the last reconcile" heartbeat. The R1.2c lint gate caught the regression and applied a band-aid fix: forcefully overriding `LastTransitionTime` after `apimeta.SetStatusCondition` (lines 157-164 of `utils.go`). The R1 retrospective identified this as a K8s condition API convention violation and designed the proper fix using `RequeueAfter`.

### Exact Code Changes

#### Task 1: Remove Force-Override (utils.go lines 155-165)

**Current code** (`controllers/vaultresourcecontroller/utils.go:155-165`):

```go
	conditions := conditionsAware.GetConditions()
	apimeta.SetStatusCondition(&conditions, condition)
	// apimeta.SetStatusCondition only updates LastTransitionTime when Status changes.
	// We always stamp it so observers can detect that reconciliation occurred.
	for i := range conditions {
		if conditions[i].Type == condition.Type {
			conditions[i].LastTransitionTime = condition.LastTransitionTime
			break
		}
	}
	conditionsAware.SetConditions(conditions)
```

**After change:**

```go
	conditions := conditionsAware.GetConditions()
	apimeta.SetStatusCondition(&conditions, condition)
	conditionsAware.SetConditions(conditions)
```

Delete the comment (line 157-158) and the `for i := range` loop (lines 159-164). Keep the 3 surrounding lines unchanged.

#### Task 2: Modify ManageOutcome (utils.go line 198-200)

**Current code** (`controllers/vaultresourcecontroller/utils.go:198-200`):

```go
func ManageOutcome(context context.Context, r ReconcilerBase, obj client.Object, issue error) (reconcile.Result, error) {
	return ManageOutcomeWithRequeue(context, r, obj, issue, 0)
}
```

**After change:**

```go
func ManageOutcome(context context.Context, r ReconcilerBase, obj client.Object, issue error) (reconcile.Result, error) {
	requeueAfter := time.Duration(0)
	if issue == nil && IsDriftDetectionEnabled() {
		requeueAfter = SyncPeriod
	}
	return ManageOutcomeWithRequeue(context, r, obj, issue, requeueAfter)
}
```

**Why only on success:** When `issue != nil`, controller-runtime requeues with exponential backoff regardless of `RequeueAfter`. Setting it on the error path is unnecessary and misleading.

**Why this works end-to-end:** `ReconcileWithFunctions` (in `reconcile_skeleton.go`) calls `ManageOutcome` on both the success path (line 82) and error paths (lines 67, 73, 80). After a successful reconcile, the returned `Result{RequeueAfter: SyncPeriod}` tells the controller-runtime work queue to schedule the next reconciliation in `SyncPeriod` seconds. The work queue triggers `Reconcile` directly — it does NOT generate an Update event, so the predicate is not involved.

**Controllers that bypass `ManageOutcome`:** `VaultSecretReconciler` and `RandomSecretReconciler` call `ManageOutcomeWithRequeue` directly with custom durations. `DatabaseSecretEngineConfigReconciler` returns raw `reconcile.Result` in some paths. These controllers already manage their own requeue timing and are NOT affected by this change — they do not use `ReconcileWithFunctions` or `ManageOutcome`. No action needed for them.

#### Task 3: Simplify PeriodicReconcilePredicate.Update() (utils.go lines 235-268)

**Current code** (`controllers/vaultresourcecontroller/utils.go:235-268`):

```go
func (p PeriodicReconcilePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}
	if e.ObjectNew.GetGeneration() != e.ObjectOld.GetGeneration() {
		return true
	}
	if !IsDriftDetectionEnabled() {
		return false
	}
	if conditionsAware, ok := e.ObjectNew.(vaultutils.ConditionsAware); ok {
		conditions := conditionsAware.GetConditions()
		for _, condition := range conditions {
			if condition.Type == ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
				timeSinceLastReconcile := time.Since(condition.LastTransitionTime.Time)
				if timeSinceLastReconcile >= SyncPeriod {
					return true
				}
				break
			}
		}
	}
	return false
}
```

**After change:**

```go
func (p PeriodicReconcilePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}
	return e.ObjectNew.GetGeneration() != e.ObjectOld.GetGeneration()
}
```

The predicate becomes a pure generation-based filter. Drift detection is handled entirely by `RequeueAfter` via the controller-runtime work queue — no predicate involvement.

**What can be cleaned up:** After this change, the following become dead code within the predicate:
- The `ReconcileInterval` struct field — no longer read by `Update()`
- The imports used only by the old predicate logic (verify if `vaultutils`, `metav1`, `time` are still needed elsewhere in `utils.go` before removing)

Keep the struct field and constructors for backward compatibility — they are exported API. The `ReconcileInterval` field is simply unused.

**Import cleanup:** After simplifying `Update()`, check if any imports in `utils.go` become unused. The `event` import is still needed. `vaultutils` is used elsewhere. `metav1` is used in `ManageOutcomeWithRequeue`. `time` is used for `SyncPeriod` and `ManageOutcomeWithRequeue`. No imports should need removal.

#### Task 4: Update Unit Tests (utils_test.go)

**Current test cases in `TestPeriodicReconcilePredicate_Update`:**

| Test Case | Old Expected | New Expected | Reason |
|-----------|-------------|-------------|--------|
| "generation changes" | `true` | `true` | Generation change still triggers reconcile |
| "interval elapsed + drift enabled" | `true` | **`false`** | Predicate no longer does time-based checks — `RequeueAfter` handles this |
| "interval elapsed + drift disabled" | `false` | `false` | Same — no reconcile on same generation |
| "interval not elapsed" | `false` | `false` | Same — no reconcile on same generation |
| "last reconcile failed" | `false` | `false` | Same — no reconcile on same generation |
| "no conditions + drift enabled" | `false` | `false` | Same — no reconcile on same generation |

**What to change:**
1. The "interval elapsed + drift enabled" test case: change `expectedResult` from `true` to `false`
2. Remove the `driftDetectionEnabled` field and env var setup/teardown from all non-generation-change test cases — the predicate no longer checks drift detection
3. Remove mock `GetConditions` setup — the predicate no longer type-asserts to `ConditionsAware`
4. Simplify remaining tests: only 2 meaningful cases — "generation changed → true" and "generation unchanged → false"

Alternatively, keep the extra test cases for documentation but update their expected values and remove the drift-detection-specific setup. The important thing is that the test proves the predicate is generation-only.

### Impact Analysis

**Behavioral changes:**
- `ReconcileSuccessful.LastTransitionTime` now follows standard K8s semantics (updates only on status transition)
- Drift detection reconciliation is triggered by the work queue timer (`RequeueAfter`) instead of by predicate timestamp checks
- No need for annotation triggers (`triggerNonSpecUpdate`) — reconciliation happens automatically

**What stays the same:**
- Drift detection feature behavior (Vault state is checked periodically and corrected)
- `ENABLE_DRIFT_DETECTION` and `SYNC_PERIOD_SECONDS` env vars
- All CRD schemas, RBAC markers, webhook behavior
- Controller registration in `main.go`
- `ReconcileWithFunctions` skeleton — it calls `ManageOutcome` which now returns `RequeueAfter`

**Reduced etcd write pressure:** Status updates only occur when the condition status actually changes (e.g., from `ReconcileFailed` to `ReconcileSuccessful`). Previously, `LastTransitionTime` was updated on EVERY reconcile, causing a status write even when nothing changed.

### What NOT to Touch

- Do NOT modify `controllers/driftdetection_controller_test.go` — that is D1.0b scope. The integration tests WILL FAIL after this change because they rely on `LastTransitionTime` advancing and `triggerNonSpecUpdate` + predicate timestamp checks. D1.0b redesigns them for `RequeueAfter`.
- Do NOT modify `controllers/vaultresourcecontroller/reconcile_skeleton.go` — it calls `ManageOutcome` which handles the `RequeueAfter` internally
- Do NOT modify any `*_controller.go` files — they use `NewDefaultPeriodicReconcilePredicate()` which still works (returns a generation-based predicate)
- Do NOT modify `main.go` — `SetSyncPeriod` and `SYNC_PERIOD_SECONDS` wiring remain unchanged
- Do NOT modify any `*_types.go` files — no CRD schema changes
- Do NOT create a `.golangci.yml` config file
- Do NOT run `make integration` — the drift detection integration tests will fail (expected; D1.0b fixes them)

### Verification Strategy

1. **`make manifests generate`** — zero diffs expected (no type or RBAC changes)
2. **`make fmt vet test`** — all unit tests pass (including updated predicate tests)
3. **`golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...`** (v1.64.8) — zero findings
4. Do NOT run `make integration` — drift detection tests are expected to fail (D1.0b scope)

### Known Integration Test Breakage (D1.0b scope)

The following integration tests in `controllers/driftdetection_controller_test.go` will fail after this change:

1. **"Should correct drift when Vault policy is manually modified"** — relies on `triggerNonSpecUpdate` + predicate timestamp check to trigger reconciliation. With `RequeueAfter`, reconciliation happens automatically — no annotation trigger needed.
2. **"Should not write when no drift exists (no false positive)"** — checks `getReconcileSuccessfulTime(updated).After(initialTime.Time)`. After this change, `LastTransitionTime` doesn't advance on same-status reconciles, so this check will always fail. D1.0b needs to redesign the "no false positive" assertion.
3. **"Should correct drift when tune config is manually modified"** — same annotation-trigger pattern.
4. **"Drift detection disabled — drift persists"** — this test may actually PASS because it checks that drift is NOT corrected when drift detection is disabled. However, the `LastTransitionTime` assertion at line 356 will need updating.

### Previous Story Intelligence

**From R1.2c (lint green gate — applied the band-aid):**
- The force-override fix was a 6-line addition to `ManageOutcomeWithRequeue` after discovering that `apimeta.SetStatusCondition` broke drift detection
- The review explicitly flagged this as temporary: "dedicated story/epic will redesign `PeriodicReconcilePredicate` to use a different signal so `apimeta.SetStatusCondition` can operate unmodified"
- [Source: `_bmad-output/implementation-artifacts/R1-2c-lint-green-gate-verify-full-compliance.md#R1.3 Regression Fix`]

**From R1.3 (dependency modernization — introduced SetStatusCondition):**
- The migration from `AddOrReplaceCondition` to `apimeta.SetStatusCondition` was Task 4
- Dev notes explicitly warned about `LastTransitionTime` semantic difference
- [Source: `_bmad-output/implementation-artifacts/R1-3-dependency-modernization-drop-deprecated-replace-handrolled.md#Task 4`]

**From R1.5 (reconciler struct deduplication):**
- All 4 reconciler types now delegate to `ReconcileWithFunctions` → `ManageOutcome`
- `ManageOutcome` is the single entry point for condition management
- The shared skeleton is in `reconcile_skeleton.go`
- [Source: `_bmad-output/implementation-artifacts/R1-5-reconciler-struct-deduplication.md`]

**From R1 Retrospective:**
- "The proper fix is to use `RequeueAfter` for drift detection timing — the idiomatic controller-runtime pattern for periodic reconciliation"
- Solution has 3 parts: remove force-override, return `RequeueAfter`, simplify predicate
- [Source: `_bmad-output/implementation-artifacts/epic-R1-retro-2026-06-21.md#Proposed Solution for LastTransitionTime`]

### golangci-lint Version

Use **v1.64.8** (the verified baseline from R1.2c). The Makefile declares `v1.59.1` — override manually:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

### Kind Cluster Note

Do NOT spin up a Kind cluster or run `make integration` for this story. This story is unit-test-only. The integration test updates are D1.0b.

### Project Structure Notes

- Only 2 files modified: `controllers/vaultresourcecontroller/utils.go`, `controllers/vaultresourcecontroller/utils_test.go`
- No new files created
- No CRD schema changes
- No new dependencies
- All changes confined to `controllers/vaultresourcecontroller/` package

### References

- [Source: controllers/vaultresourcecontroller/utils.go:131-200] — `ManageOutcomeWithRequeue`, `ManageOutcome` (code to modify)
- [Source: controllers/vaultresourcecontroller/utils.go:202-268] — `PeriodicReconcilePredicate` (code to simplify)
- [Source: controllers/vaultresourcecontroller/utils.go:49-64] — `SyncPeriod`, `SetSyncPeriod`, `IsDriftDetectionEnabled` (used by new `ManageOutcome` logic)
- [Source: controllers/vaultresourcecontroller/reconcile_skeleton.go:56-83] — `ReconcileWithFunctions` (calls `ManageOutcome` — no changes needed)
- [Source: controllers/vaultresourcecontroller/utils_test.go:76-264] — unit tests (to update)
- [Source: controllers/driftdetection_controller_test.go] — integration tests (D1.0b scope — DO NOT MODIFY)
- [Source: main.go:114-141] — `SYNC_PERIOD_SECONDS` parsing and `SetSyncPeriod` call
- [Source: _bmad-output/implementation-artifacts/epic-R1-retro-2026-06-21.md#Proposed Solution for LastTransitionTime] — design rationale
- [Source: _bmad-output/implementation-artifacts/R1-2c-lint-green-gate-verify-full-compliance.md#R1.3 Regression Fix] — band-aid fix being replaced
- [Source: _bmad-output/implementation-artifacts/R1-3-dependency-modernization-drop-deprecated-replace-handrolled.md#Task 4] — SetStatusCondition migration context
- [Source: _bmad-output/implementation-artifacts/R1-5-reconciler-struct-deduplication.md] — ReconcileWithFunctions shared skeleton
- [Source: _bmad-output/project-context.md#Framework-Specific Rules] — controller-runtime reconcile flow

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
