# Story D1.0b: Update Drift Detection Tests for RequeueAfter

Status: ready-for-dev

## Story

As an operator developer,
I want the drift detection integration tests updated to rely on `RequeueAfter` for periodic reconciliation,
So that the test suite passes after D1.0a removed `LastTransitionTime` heartbeating and simplified the predicate to generation-only filtering.

## Acceptance Criteria

1. **Given** D1.0a removed the `triggerNonSpecUpdate` → predicate timestamp-check reconcile trigger **When** the "Should correct drift when Vault policy is manually modified" test is updated **Then** it relies on `RequeueAfter` natural requeue (wait for `SyncPeriod` to elapse) instead of annotation-triggered reconciliation
2. **Given** D1.0a made `LastTransitionTime` follow standard K8s semantics (only updates on status transition) **When** the "Should not write when no drift exists (no false positive)" test is updated **Then** it uses generation-based or Vault state assertions instead of `LastTransitionTime` advancing to prove reconciliation occurred
3. **Given** D1.0a simplified `PeriodicReconcilePredicate` to generation-only filtering **When** the "Should correct drift when tune config is manually modified" test is updated **Then** it relies on `RequeueAfter` natural requeue
4. **Given** D1.0a made `ManageOutcome` return `RequeueAfter: SyncPeriod` only when drift detection is enabled and reconcile succeeds **When** the "Drift detection disabled — drift persists" test is updated **Then** it validates that without `RequeueAfter`, no periodic reconciliation occurs (drift persists)
5. **Given** all test changes **When** `make integration` is run (with Kind cluster and Vault) **Then** all drift detection integration tests pass
6. **Given** all test changes **When** `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...` is run with v1.64.8 **Then** zero findings

## Tasks / Subtasks

- [ ] Task 1: Remove `triggerNonSpecUpdate` helper function (AC: 1, 3)
  - [ ] 1.1: Delete the `triggerNonSpecUpdate` function (lines 39-47) — it is no longer needed because `RequeueAfter` handles periodic reconciliation via the controller-runtime work queue
  - [ ] 1.2: Remove any remaining callers in the file
- [ ] Task 2: Remove `getReconcileSuccessfulTime` helper function (AC: 2)
  - [ ] 2.1: Delete the `getReconcileSuccessfulTime` function (lines 49-56) — `LastTransitionTime` no longer advances on same-status reconciles, so it cannot be used to detect "last reconcile"
  - [ ] 2.2: Remove any remaining callers in the file
- [ ] Task 3: Rewrite "Should correct drift when Vault policy is manually modified" (AC: 1)
  - [ ] 3.1: Remove the `triggerNonSpecUpdate` call and the "Waiting for the SyncPeriod to elapse" + manual `time.Sleep`
  - [ ] 3.2: Replace with: after introducing drift in Vault, simply poll with `Eventually` for the operator to correct it — `RequeueAfter` will fire automatically after SyncPeriod (set to 5s in `enableDriftDetection`)
  - [ ] 3.3: The `Eventually` timeout should be generous enough to account for SyncPeriod (5s) + reconcile processing time — use 30s timeout with 2s interval
- [ ] Task 4: Rewrite "Should not write when no drift exists (no false positive)" (AC: 2)
  - [ ] 4.1: Remove the `getReconcileSuccessfulTime` / `LastTransitionTime` assertion pattern
  - [ ] 4.2: New strategy: after initial reconcile succeeds, wait for at least one `RequeueAfter` cycle to fire (sleep > SyncPeriod), then assert Vault policy still has exact original rules. Use `Consistently` to verify no unnecessary writes occur over a window
  - [ ] 4.3: The key assertion: Vault state matches the CR spec (original rules) throughout the observation window — proving no false positive writes
- [ ] Task 5: Rewrite "Should correct drift when tune config is manually modified" (AC: 3)
  - [ ] 5.1: Remove the `triggerNonSpecUpdate` call and `time.Sleep`
  - [ ] 5.2: After introducing tune config drift in Vault, poll with `Eventually` for automatic correction via `RequeueAfter`
  - [ ] 5.3: Use 30s timeout with 2s interval (same pattern as Task 3)
- [ ] Task 6: Rewrite "Drift detection disabled — drift persists" (AC: 4)
  - [ ] 6.1: Remove the `triggerNonSpecUpdate` call
  - [ ] 6.2: Remove the `LastTransitionTime` assertion (lines 352-357)
  - [ ] 6.3: New strategy: when drift detection is disabled, `ManageOutcome` returns `RequeueAfter: 0` — no periodic reconciliation. Introduce drift in Vault, use `Consistently` over an interval longer than what SyncPeriod would have been to prove drift persists
  - [ ] 6.4: Set `SyncPeriod` to a large value (36000s as currently done) AND unset `ENABLE_DRIFT_DETECTION` — since `ManageOutcome` checks `IsDriftDetectionEnabled()`, `RequeueAfter` will be 0 regardless of SyncPeriod value
- [ ] Task 7: Clean up unused imports (AC: 6)
  - [ ] 7.1: After removing `triggerNonSpecUpdate` and `getReconcileSuccessfulTime`, check if these imports are still needed: `"context"` (may still be used by test helpers), `metav1` (may be unused if no more `LastTransitionTime` references)
  - [ ] 7.2: Verify `context` is still needed — it's used by `k8sIntegrationClient.Get(ctx, ...)` so it stays
  - [ ] 7.3: Check if `metav1` is still needed — used in BeforeAll for condition checks; keep if still referenced
- [ ] Task 8: Verify (AC: 5, 6)
  - [ ] 8.1: Run `make integration` — all integration tests pass (requires Kind cluster + Vault)
  - [ ] 8.2: Run `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...` — zero findings
  - [ ] 8.3: Run `make test` — all unit tests pass (sanity check, should be unaffected)

## Dev Notes

### Background: Why These Tests Break After D1.0a

D1.0a made 3 changes to `controllers/vaultresourcecontroller/utils.go`:
1. Removed the `LastTransitionTime` force-override loop (lines 157-164) — `apimeta.SetStatusCondition` now operates with standard K8s semantics
2. Modified `ManageOutcome` to return `RequeueAfter: SyncPeriod` when `issue == nil && IsDriftDetectionEnabled()`
3. Simplified `PeriodicReconcilePredicate.Update()` to generation-only filtering

The integration tests in `controllers/driftdetection_controller_test.go` relied on the OLD mechanism:
- **Old flow:** wait for SyncPeriod → call `triggerNonSpecUpdate` (sets annotation) → annotation change triggers metadata-only Update event → predicate checks `LastTransitionTime` elapsed → allows reconcile → drift corrected
- **New flow:** initial reconcile succeeds → `ManageOutcome` returns `Result{RequeueAfter: SyncPeriod}` → controller-runtime work queue schedules next reconcile after SyncPeriod → reconcile fires automatically → drift corrected

The new flow is simpler: no annotation trigger needed, no predicate involvement for drift detection. The work queue timer handles everything.

### Key Design Principle: Work Queue Requeue vs Predicate-Based

After D1.0a:
- **Generation changes** → handled by the predicate (immediate reconcile on spec change)
- **Drift detection** → handled by `RequeueAfter` (controller-runtime work queue fires Reconcile directly — does NOT generate an Update event, so the predicate is NOT involved)

This means the tests should NOT try to trigger reconciliation via annotation updates or metadata changes. They should simply wait for the automatic `RequeueAfter` to fire.

### Test Timing Considerations

`enableDriftDetection(5 * time.Second)` sets SyncPeriod to 5s. After D1.0a:
- A successful reconcile returns `Result{RequeueAfter: 5s}`
- After 5s, controller-runtime calls `Reconcile` again
- If drift exists → corrected → returns `Result{RequeueAfter: 5s}` again
- If no drift → no Vault write → returns `Result{RequeueAfter: 5s}` again

The tests should use `Eventually` with a timeout that accommodates:
- Initial reconcile completing (~2-5s)
- One SyncPeriod (5s)
- Reconcile processing time (~1-2s)
- Total: ~12-15s minimum, use 30s timeout for safety

### Exact Test Rewrites

#### Test 1: "Should correct drift when Vault policy is manually modified"

**Current flow (BROKEN after D1.0a):**
```
1. Overwrite policy in Vault
2. time.Sleep(6s) ← wait for SyncPeriod
3. triggerNonSpecUpdate ← trigger annotation change
4. Eventually: check Vault has original rules
```

**New flow:**
```
1. Overwrite policy in Vault
2. Eventually: check Vault has original rules (30s timeout, 2s interval)
```

The `RequeueAfter` from the BeforeAll's initial successful reconcile will fire after 5s, triggering a new reconcile that detects and corrects the drift. No annotation trigger needed.

#### Test 2: "Should not write when no drift exists (no false positive)"

**Current flow (BROKEN after D1.0a):**
```
1. Get initial LastTransitionTime
2. time.Sleep(6s)
3. triggerNonSpecUpdate
4. Eventually: LastTransitionTime advanced (proves reconcile ran)
5. Check Vault still has original rules
```

**New flow:**
```
1. Verify Vault currently has correct rules (baseline)
2. Wait for RequeueAfter cycle to fire (sleep > SyncPeriod, e.g., 7s)
3. Consistently (10s window, 2s interval): Vault policy still equals original rules
```

We don't need to prove "reconcile ran" — we only need to prove "no unnecessary Vault write happened." The `Consistently` assertion over a window that spans at least one RequeueAfter cycle proves this.

#### Test 3: "Should correct drift when tune config is manually modified"

**Current flow (BROKEN after D1.0a):**
```
1. Modify tune config in Vault
2. time.Sleep(6s)
3. triggerNonSpecUpdate
4. Eventually: check tune config corrected
```

**New flow:**
```
1. Modify tune config in Vault
2. Eventually: check tune config corrected (30s timeout, 2s interval)
```

Same pattern as Test 1.

#### Test 4: "Drift detection disabled — drift persists"

**Current flow (PARTIALLY BROKEN after D1.0a):**
```
1. Record baseline LastTransitionTime
2. Overwrite policy in Vault
3. triggerNonSpecUpdate
4. Consistently: Vault still has drifted rules (drift persists)
5. Assert LastTransitionTime did NOT advance
```

**New flow:**
```
1. Overwrite policy in Vault
2. Consistently (15s window, 2s interval): Vault still has drifted rules
```

With drift detection disabled, `ManageOutcome` returns `RequeueAfter: 0` — no periodic requeue happens. The annotation trigger is irrelevant because the predicate is generation-only. The drift persists because there's no mechanism to trigger another reconcile. We just need to prove drift persists over a reasonable window.

### What NOT to Touch

- Do NOT modify `controllers/vaultresourcecontroller/utils.go` — that was D1.0a's scope
- Do NOT modify `controllers/vaultresourcecontroller/utils_test.go` — those unit tests were updated in D1.0a
- Do NOT modify any `*_controller.go` files
- Do NOT modify test fixtures in `test/drift-detection/` — the YAML fixtures are still valid
- Do NOT modify `controllers/suite_integration_test.go` — controller registration is unchanged
- Do NOT create a `.golangci.yml` config file

### Files Modified

Only 1 file: `controllers/driftdetection_controller_test.go`

### Imports After Changes

Expected remaining imports (verify after edits):
```go
import (
    "os"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/controllers/controllertestutils"
    "github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    "sigs.k8s.io/controller-runtime/pkg/client"
)
```

Notes:
- `"context"` is removed — `triggerNonSpecUpdate` was the only user; the global `ctx` variable used in test bodies is from the suite, not this import
- `metav1` stays — used in BeforeAll condition checks (`metav1.ConditionTrue`)
- `apierrors` stays — used in AfterAll cleanup (`apierrors.IsNotFound`)
- `client` stays — used in AfterAll cleanup (`client.Object` for Delete)
- `controllertestutils` stays — used in "disabled" test for `DecodeInstance`

### enableDriftDetection Helper: Keep Unchanged

The `enableDriftDetection` function (lines 22-37) is still correct and needed:
- Sets `ENABLE_DRIFT_DETECTION=true`
- Sets `SyncPeriod` to the test value (5s)
- Returns cleanup function
- `ManageOutcome` reads `IsDriftDetectionEnabled()` and `SyncPeriod` — these still need to be set for the integration tests to work

### Verification Strategy

1. **Set up Kind cluster + Vault:** `make kind-setup` (or equivalent — see Makefile)
2. **Run integration tests:** `make integration` or `go test -tags=integration ./controllers/ -v -timeout 300s`
3. **Lint check:** `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...` (v1.64.8)
4. **Unit test sanity:** `make test`

### golangci-lint Version

Use **v1.64.8** (the verified baseline from R1.2c). The Makefile declares `v1.59.1` — override manually:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
```

### Previous Story Intelligence

**From D1.0a (the prerequisite story):**
- Removed the `LastTransitionTime` force-override in `ManageOutcomeWithRequeue`
- Modified `ManageOutcome` to return `RequeueAfter: SyncPeriod` when `issue == nil && IsDriftDetectionEnabled()`
- Simplified `PeriodicReconcilePredicate.Update()` to generation-only filtering
- D1.0a dev notes explicitly state: "Do NOT modify `controllers/driftdetection_controller_test.go` — that is D1.0b scope"
- D1.0a warns: "The integration tests WILL FAIL after this change because they rely on `LastTransitionTime` advancing and `triggerNonSpecUpdate` + predicate timestamp checks"

**From R1.2c (lint green gate — applied the band-aid):**
- The force-override fix was temporary: "dedicated story/epic will redesign `PeriodicReconcilePredicate` to use a different signal"
- [Source: `_bmad-output/implementation-artifacts/R1-2c-lint-green-gate-verify-full-compliance.md#R1.3 Regression Fix`]

**From R1 Retrospective:**
- "The proper fix is to use `RequeueAfter` for drift detection timing — the idiomatic controller-runtime pattern for periodic reconciliation"
- D1.0b was explicitly called out as the test-update companion to D1.0a
- [Source: `_bmad-output/implementation-artifacts/epic-R1-retro-2026-06-21.md#Proposed Solution for LastTransitionTime`]

### Project Structure Notes

- Only 1 file modified: `controllers/driftdetection_controller_test.go`
- No new files created
- No CRD schema changes
- No new dependencies
- Test YAML fixtures in `test/drift-detection/` remain unchanged

### References

- [Source: controllers/driftdetection_controller_test.go] — the file being rewritten
- [Source: controllers/vaultresourcecontroller/utils.go:198-200] — `ManageOutcome` (returns `RequeueAfter` after D1.0a)
- [Source: controllers/vaultresourcecontroller/utils.go:235-268] — `PeriodicReconcilePredicate.Update()` (generation-only after D1.0a)
- [Source: controllers/vaultresourcecontroller/utils.go:49-64] — `SyncPeriod`, `IsDriftDetectionEnabled` (read by `ManageOutcome`)
- [Source: controllers/vaultresourcecontroller/reconcile_skeleton.go:56-83] — `ReconcileWithFunctions` (calls `ManageOutcome`)
- [Source: controllers/suite_integration_test.go] — test infrastructure (k8sIntegrationClient, vaultClient, namespaces)
- [Source: test/drift-detection/policy-drift-test.yaml] — Policy fixture
- [Source: test/drift-detection/secretenginemount-drift-test.yaml] — SecretEngineMount fixture
- [Source: _bmad-output/implementation-artifacts/d1-0a-fix-lasttransitiontime-misuse-requeueafter-drift-detection.md] — prerequisite story
- [Source: _bmad-output/implementation-artifacts/epic-R1-retro-2026-06-21.md#Proposed Solution for LastTransitionTime] — design rationale
- [Source: _bmad-output/project-context.md#Framework-Specific Rules] — controller-runtime reconcile flow

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
