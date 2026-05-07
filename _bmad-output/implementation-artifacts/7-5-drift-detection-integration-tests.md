# Story 7.5: Drift Detection Integration Tests

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration tests verifying that when `ENABLE_DRIFT_DETECTION=true`, the reconciler periodically re-checks Vault state and corrects drift,
So that the drift detection feature is validated end-to-end with real Vault interactions.

## Acceptance Criteria

1. **Given** a Policy resource is successfully reconciled with drift detection enabled **When** the Vault policy is manually modified via the Vault client **And** a non-spec annotation update triggers the periodic reconcile predicate **Then** the reconciler detects the drift via `IsEquivalentToDesiredState` returning false and writes the correct policy back to Vault

2. **Given** a SecretEngineMount resource is successfully reconciled with drift detection enabled **When** the Vault mount's tune configuration is manually modified via the Vault client **And** a non-spec annotation update triggers the periodic reconcile predicate **Then** the reconciler detects the tune config drift and writes the correct tune back to Vault

3. **Given** a resource is successfully reconciled with drift detection enabled **When** the Vault state has NOT been modified externally **And** a non-spec annotation update triggers the periodic reconcile predicate **Then** `IsEquivalentToDesiredState` returns true and no write occurs (no false positive drift)

4. **Given** drift detection is disabled (`ENABLE_DRIFT_DETECTION` unset) **When** the Vault state is manually modified **And** a non-spec annotation update is applied to the CR **Then** the predicate rejects the Update event (no reconcile triggered) and the drift persists uncorrected

5. **Given** the drift detection tests are added **When** `make integration` runs **Then** all existing specs pass with zero regressions

## Tasks / Subtasks

- [ ] Task 1: Create drift detection integration test file (AC: 1, 2, 3, 4, 5)
  - [ ] 1.1: Create `controllers/driftdetection_controller_test.go` with `//go:build integration` tag
  - [ ] 1.2: Use `Describe("Drift detection", Ordered, ...)` — Ordered is required because tests share state across Contexts (create → drift → verify → cleanup lifecycle)
  - [ ] 1.3: Import the `vaultresourcecontroller` package for `SetSyncPeriod`, `SyncPeriod`, and `ReconcileSuccessful` constants

- [ ] Task 2: Implement drift detection test helper functions (AC: 1, 2, 3, 4)
  - [ ] 2.1: Create `enableDriftDetection(syncPeriod time.Duration)` helper that sets `ENABLE_DRIFT_DETECTION=true` env var and calls `vaultresourcecontroller.SetSyncPeriod(syncPeriod)`, returns a cleanup func that restores original values
  - [ ] 2.2: Create `triggerNonSpecUpdate(ctx, client, obj)` helper that adds/updates a `drift-detection-trigger` annotation on the CR to generate an Update event without changing the generation (this triggers the periodic reconcile predicate)
  - [ ] 2.3: Create `waitForReconcileSuccess(ctx, client, lookupKey, obj, timeout, interval)` helper (reuse from 7.0 shared helpers if available, otherwise inline)

- [ ] Task 3: Implement Policy drift detection test (AC: 1, 3)
  - [ ] 3.1: Context "Policy drift detection with drift detection enabled"
  - [ ] 3.2: `BeforeAll`: call `enableDriftDetection(5 * time.Second)`, load `../test/policy/simple-policy.yaml` fixture, set namespace to `vaultAdminNamespaceName`, create the CR, wait for `ReconcileSuccessful=True`
  - [ ] 3.3: It "Should correct drift when Vault policy is manually modified" (AC: 1):
    - Record the original policy rules from Vault via `vaultClient.Logical().Read("sys/policy/test-simple-policy")`
    - Overwrite the policy in Vault directly via `vaultClient.Logical().Write("sys/policy/test-simple-policy", map[string]interface{}{"policy": "# drifted policy\npath \"drifted/*\" {\n  capabilities = [\"read\"]\n}"})` — this bypasses the operator
    - Verify the drift is present: read back from Vault, confirm rules changed
    - Wait >5 seconds for the SyncPeriod to elapse
    - Trigger non-spec update (annotation change) on the CR
    - `Eventually`: read the Vault policy and verify it matches the original CR-defined rules (the operator corrected the drift)
  - [ ] 3.4: It "Should not write when no drift exists (no false positive)" (AC: 3):
    - Trigger another non-spec annotation update (after the previous reconcile)
    - Wait for the reconcile to complete (poll CR for updated `ReconcileSuccessful` condition `LastTransitionTime` advancing)
    - Verify the Vault policy still has the correct rules (unchanged — no unnecessary write)
  - [ ] 3.5: `AfterAll`: delete the Policy CR, wait for removal, call the cleanup func from `enableDriftDetection`

- [ ] Task 4: Implement SecretEngineMount drift detection test (AC: 2)
  - [ ] 4.1: Context "SecretEngineMount tune config drift detection"
  - [ ] 4.2: `BeforeAll`: enable drift detection (if not already enabled by shared setup), load a SecretEngineMount fixture (e.g., a KV v2 mount at a unique path like `drift-test-kv`), set namespace to `vaultAdminNamespaceName`, create the CR, wait for `ReconcileSuccessful=True`
  - [ ] 4.3: It "Should correct drift when tune config is manually modified" (AC: 2):
    - Read current tune config from Vault via `vaultClient.Sys().MountConfig("drift-test-kv")` or `vaultClient.Logical().Read("sys/mounts/drift-test-kv/tune")`
    - Modify the tune config directly via `vaultClient.Sys().TuneMount("drift-test-kv", vault.MountConfigInput{Description: "drifted description"})` or equivalent `Logical().Write` to the tune endpoint — change a field the operator manages (e.g., `default_lease_ttl` or `description`)
    - Verify the drift is present
    - Wait >5s for SyncPeriod, trigger annotation update
    - `Eventually`: read tune config from Vault and verify it matches the CR spec (drift corrected)
  - [ ] 4.4: `AfterAll`: delete the SecretEngineMount CR, wait for removal

- [ ] Task 5: Implement drift detection disabled test (AC: 4)
  - [ ] 5.1: Context "Drift detection disabled — drift persists"
  - [ ] 5.2: `BeforeAll`: ensure `ENABLE_DRIFT_DETECTION` is unset, restore default `SyncPeriod`, create a simple Policy CR, wait for `ReconcileSuccessful=True`
  - [ ] 5.3: It "Should NOT correct drift when drift detection is disabled" (AC: 4):
    - Overwrite the policy in Vault directly (same technique as Task 3.3)
    - Trigger annotation update on the CR
    - `Consistently` for 10 seconds: read Vault policy and verify it still has the drifted content (operator did NOT reconcile because predicate rejected the Update event)
  - [ ] 5.4: `AfterAll`: delete the Policy CR, wait for removal

- [ ] Task 6: Create test fixtures (if needed) (AC: 1, 2)
  - [ ] 6.1: Reuse existing `test/policy/simple-policy.yaml` for Policy tests — already used by `policy_controller_test.go`; if name conflicts would cause issues in parallel, create `test/drift-detection/policy-drift-test.yaml` with a unique name (e.g., `test-drift-policy`)
  - [ ] 6.2: For SecretEngineMount: create `test/drift-detection/secretenginemount-drift-test.yaml` — a KV v2 mount at path `drift-test-kv` with some non-default tune config (e.g., `description: "drift test mount"`)

- [ ] Task 7: Verify no regressions (AC: 5)
  - [ ] 7.1: Run `make test` — unit tests pass
  - [ ] 7.2: Run `make integration` — all specs pass including new drift detection tests

## Dev Notes

### How Drift Detection Works (End-to-End Flow)

The drift detection mechanism has three independent pieces that must all work:

1. **Manager cache resync** — `Cache.SyncPeriod` in `ctrl.Options` controls how often the informer resyncs all objects. In production, `SYNC_PERIOD_SECONDS` sets this (default 10h). In the integration test suite, `Cache.SyncPeriod` is NOT set (nil), so there are **no automatic periodic resyncs**.

2. **`PeriodicReconcilePredicate`** — attached to every controller via `builder.WithPredicates(NewDefaultPeriodicReconcilePredicate())`. On an Update event with unchanged generation, it checks:
   - `IsDriftDetectionEnabled()` → reads `ENABLE_DRIFT_DETECTION` env var at runtime (must be exactly `"true"`)
   - `time.Since(ReconcileSuccessful.LastTransitionTime) >= ReconcileInterval` → uses package-level `SyncPeriod` set via `SetSyncPeriod()`

3. **`CreateOrUpdate`** — in `api/v1alpha1/utils/vaultobject.go:159-174`. When reconcile runs: `read()` → `IsEquivalentToDesiredState(payload)` → `write()` only if not equivalent.

### Test Triggering Strategy (Critical Architecture Decision)

Since the integration test suite creates the manager **without** `Cache.SyncPeriod`, the cache does NOT automatically resync objects. We cannot rely on automatic Update events. Instead, we **manually trigger** the predicate by adding an annotation to the CR:

```go
func triggerNonSpecUpdate(ctx context.Context, c client.Client, obj client.Object) error {
    annotations := obj.GetAnnotations()
    if annotations == nil {
        annotations = make(map[string]string)
    }
    annotations["drift-detection-trigger"] = time.Now().Format(time.RFC3339Nano)
    obj.SetAnnotations(annotations)
    return c.Update(ctx, obj)
}
```

Updating annotations **does NOT change `metadata.generation`** (only spec changes bump generation). This produces an Update event where `ObjectOld.Generation == ObjectNew.Generation`, which falls into the periodic reconcile branch of the predicate. Combined with `ENABLE_DRIFT_DETECTION=true` and a short `SyncPeriod` (5s), the predicate allows the reconcile.

**Why this is correct:** In production, the cache resync generates the same type of Update event (generation unchanged). The annotation approach simulates exactly that event in a controlled, deterministic way. The predicate logic is identical regardless of what caused the Update event.

### Environment Variable / Package State Management

The tests modify two pieces of global state. Always save and restore:

```go
func enableDriftDetection(syncPeriod time.Duration) func() {
    origSyncPeriod := vaultresourcecontroller.SyncPeriod
    origDriftEnv, origDriftSet := os.LookupEnv("ENABLE_DRIFT_DETECTION")

    os.Setenv("ENABLE_DRIFT_DETECTION", "true")
    vaultresourcecontroller.SetSyncPeriod(syncPeriod)

    return func() {
        vaultresourcecontroller.SetSyncPeriod(origSyncPeriod)
        if origDriftSet {
            os.Setenv("ENABLE_DRIFT_DETECTION", origDriftEnv)
        } else {
            os.Unsetenv("ENABLE_DRIFT_DETECTION")
        }
    }
}
```

Use `BeforeAll`/`AfterAll` in each `Ordered` Describe block to scope the state change.

### Direct Vault Modification (How to Write to Vault Bypassing the Operator)

The integration test suite already has a package-level `vaultClient *vault.Client` (root token, full Vault access). Use it to modify Vault state directly:

**Policy:**
```go
// Write a drifted policy directly to Vault
_, err := vaultClient.Logical().Write("sys/policy/test-drift-policy", map[string]interface{}{
    "policy": `path "drifted/*" { capabilities = ["read"] }`,
})
```

**SecretEngineMount tune:**
```go
// Modify tune config directly — change description
err := vaultClient.Sys().TuneMount("drift-test-kv", vault.MountConfigInput{
    Description: &driftedDescription,
})
```

Alternative for tune (if `vault.MountConfigInput` doesn't expose the field you need):
```go
_, err := vaultClient.Logical().Write("sys/mounts/drift-test-kv/tune", map[string]interface{}{
    "description": "drifted description",
})
```

### Why Policy and SecretEngineMount Were Chosen

- **Policy** — Simplest type. No external dependencies. Clean Vault API (`sys/policy/<name>`). Easy to read and verify. IsDeletable=true, so cleanup is straightforward. Represents the standard `VaultResource` reconciler path.

- **SecretEngineMount** — Represents the `VaultEngineResource` reconciler path. Tune-only comparison via `IsEquivalentToDesiredState` is architecturally different from standard types. Verifying drift correction for tune config covers a distinct code path.

Together, these two types cover both reconciler variants (`VaultResource` and `VaultEngineResource`) without requiring external services (databases, LDAP, etc.).

### Test Fixture Specifications

**`test/drift-detection/policy-drift-test.yaml`:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Policy
metadata:
  name: test-drift-policy
spec:
  authentication:
    path: kubernetes
    role: policy-admin
    serviceAccount:
      name: default
  policy: |
    path "secret/data/drift-test/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }
```

**`test/drift-detection/secretenginemount-drift-test.yaml`:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: drift-test-kv
spec:
  authentication:
    path: kubernetes
    role: policy-admin
    serviceAccount:
      name: default
  type: kv
  path: drift-test-kv
  config:
    options:
      version: "2"
    description: "drift test mount"
    defaultLeaseTTL: "1h"
```

Use unique names (`test-drift-policy`, `drift-test-kv`) to avoid collisions with other integration tests running in the same cluster.

### Verifying "No False Positive" (AC: 3)

After drift is corrected and the operator reconciles successfully, trigger another non-spec update. The reconciler should:
1. Read Vault state
2. Call `IsEquivalentToDesiredState(payload)` → returns `true` (no drift)
3. Skip `write()` — no unnecessary Vault write

To verify no write occurred, check that the `ReconcileSuccessful` condition's `LastTransitionTime` is updated (reconcile ran) but the Vault state hasn't changed. The condition timestamp advancing proves the reconcile ran; the unchanged Vault state proves no write occurred.

**Important:** After Story 7.4 fixes `IsEquivalentToDesiredState` to ignore extra Vault fields, this "no false positive" test validates that the fix works in a real Vault integration scenario — extra fields returned by Vault do NOT trigger unnecessary writes.

### `Consistently` for Disabled Drift Detection (AC: 4)

Use Gomega's `Consistently` matcher (not `Eventually`) to prove the drift persists:

```go
driftedRules := `path "drifted/*" { capabilities = ["read"] }`
Consistently(func() string {
    secret, err := vaultClient.Logical().Read("sys/policy/test-drift-policy")
    if err != nil || secret == nil {
        return ""
    }
    return secret.Data["rules"].(string)
}, 10*time.Second, 2*time.Second).Should(ContainSubstring("drifted"))
```

`Consistently` asserts the condition holds true for the entire duration (10s), proving the operator did NOT correct the drift.

### Dependency on Story 7.4

Story 7.4 (extra-field audit + `filterPayloadToDesiredKeys`) must be completed before this story. Without 7.4's fixes:
- Types using bare `reflect.DeepEqual` will see extra Vault fields as drift, causing false positives
- The "no false positive" test (AC: 3) would fail because `IsEquivalentToDesiredState` would return `false` due to extra Vault fields even when no real drift exists

If 7.4 is not yet complete, this story's tests will still be structurally correct but AC: 3 may fail for some types. The developer should verify 7.4 is merged before running `make integration`.

### Timing Considerations

- Use `SyncPeriod` of `5 * time.Second` for tests — short enough for fast tests, long enough to be deterministic
- After modifying Vault and before triggering the annotation update, wait at least 6 seconds (`time.Sleep(6 * time.Second)`) to ensure the `ReconcileSuccessful.LastTransitionTime` is old enough for the predicate to allow reconcile
- Use `Eventually` with `timeout: 60s` and `interval: 2s` for drift correction verification (the reconciler may need a few seconds to process)

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/driftdetection_controller_test.go` | New | Drift detection integration tests |
| 2 | `test/drift-detection/policy-drift-test.yaml` | New | Policy fixture with unique name |
| 3 | `test/drift-detection/secretenginemount-drift-test.yaml` | New | SecretEngineMount fixture with unique name |

### Previous Story Intelligence

**From Story 7.4 (Audit & Harden IsEquivalentToDesiredState):**
- `filterPayloadToDesiredKeys` helper filters Vault read responses to only managed keys
- 31 bare-DeepEqual types updated to use the filter pattern
- Existing "payloadWithExtra" tests updated from expecting `false` to `true`
- No CRD schema changes, no controller changes, no webhook changes

**From Story 7.3 (Error Path Integration Tests):**
- Integration test infrastructure is stable
- Standard pattern: create CR → poll for condition → verify → delete
- `vault-admin` namespace has working auth roles: `policy-admin`, `database-engine-admin`, `kv-engine-admin`, `secret-writer`
- `AfterAll` cleanup blocks used for defensive cleanup

**From Epic 6 Retrospective:**
- Codebase stable on main at commit `9fc8b3c`
- Coverage at 53.7%
- "Continue detailed dev notes in story specs" — applied

### Key Source References

- `controllers/vaultresourcecontroller/utils.go:191-255` — `PeriodicReconcilePredicate` implementation
- `controllers/vaultresourcecontroller/utils.go:48-63` — `SyncPeriod`, `SetSyncPeriod`, `IsDriftDetectionEnabled`
- `controllers/vaultresourcecontroller/utils_test.go:76-222` — Unit tests for drift detection predicate
- `api/v1alpha1/utils/vaultobject.go:159-174` — `CreateOrUpdate` flow
- `controllers/suite_integration_test.go:54-66,81-104` — Integration test Vault client setup
- `controllers/policy_controller_test.go` — Reference integration test pattern
- `controllers/policy_controller.go:77-81` — `SetupWithManager` with `NewDefaultPeriodicReconcilePredicate()`
- `main.go:114-141` — Production `SYNC_PERIOD_SECONDS` + `SetSyncPeriod` wiring
- `_bmad-output/project-context.md#Vault API Gotchas` — Filter read payload to managed keys

### Project Structure Notes

- Integration tests live in `controllers/` with `//go:build integration` tag — same package as the suite
- Test fixtures in `test/` directory, organized by feature subdirectory
- All integration tests share the same manager, `k8sIntegrationClient`, `vaultClient`, `ctx`, and test namespaces from `suite_integration_test.go`
- Global state (`SyncPeriod`, `ENABLE_DRIFT_DETECTION` env) must be saved/restored to avoid affecting other tests

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
