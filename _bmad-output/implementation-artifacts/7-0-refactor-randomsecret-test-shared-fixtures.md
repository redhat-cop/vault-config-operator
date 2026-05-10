# Story 7.0: Refactor RandomSecret Test Shared Fixtures

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the duplicated KV v2 bootstrap/teardown code in RandomSecret and VaultSecret integration tests extracted into shared helper functions,
So that future test modifications only change one place and new KV v2 test scenarios can be added with minimal boilerplate.

## Acceptance Criteria

1. **Given** the KV v2 bootstrap stack (PasswordPolicy, 2 Policies, 2-3 KubernetesAuthEngineRoles, SecretEngineMount) is needed by a test Context **When** the test sets up its infrastructure **Then** it calls a single shared helper function instead of ~190 lines of copy-pasted code

2. **Given** the KV v2 bootstrap stack is no longer needed **When** the test tears down its infrastructure **Then** it calls a single shared teardown helper instead of ~100 lines of repeated Eventually+Delete+JSON marshal blocks

3. **Given** the shared helpers are introduced **When** `make integration` runs **Then** all existing tests pass with zero regressions — no behavioral change, same assertions, same ordering

4. **Given** `test/randomsecret/v2/08-passwordpolicy-simple-password-policy-v2.yaml` exists as a dead duplicate **When** the refactoring is complete **Then** the file is deleted (it duplicates `00-passwordpolicy-simple-password-policy-v2.yaml` and is never referenced)

5. **Given** the repeated `Eventually(func() bool { ... ReconcileSuccessful ... })` pattern appears 30+ times across these test files **When** the refactoring is complete **Then** a shared `waitForReconcileSuccess` helper replaces the pattern (or is used by the bootstrap helper)

## Tasks / Subtasks

- [x] Task 1: Create shared helper file (AC: 1, 2, 5)
  - [x] 1.1: Create `controllers/integration_test_helpers_test.go` with `//go:build integration` tag
  - [x] 1.2: Implement `waitForReconcileSuccess` generic helper
  - [x] 1.3: Implement `SetupKVv2Stack` helper (creates PasswordPolicy, Policies, KubernetesAuthEngineRoles, SecretEngineMount; returns a struct with all created instances)
  - [x] 1.4: Implement `TeardownKVv2Stack` helper (deletes all resources + waits for Vault cleanup)
  - [x] 1.5: Implement `SetupKVv2StackWithReader` variant that also creates the secret-reader policy and role (needed by retain and VaultSecret v2 tests)

- [x] Task 2: Refactor `randomsecret_controller_test.go` (AC: 1, 2, 3, 5)
  - [x] 2.1: Replace bootstrap block in "retain policy" Context with `SetupKVv2StackWithReader` call
  - [x] 2.2: Replace bootstrap block in "multi-key" Context with `SetupKVv2Stack` call
  - [x] 2.3: Replace bootstrap block in "multi-key recreate" Context with `SetupKVv2Stack` call
  - [x] 2.4: Replace bootstrap block in "refreshPeriod" Context with `SetupKVv2Stack` call
  - [x] 2.5: Replace teardown blocks in all 4 Contexts with `TeardownKVv2Stack` calls
  - [x] 2.6: Replace inline `Eventually ReconcileSuccessful` checks for RandomSecret/VaultSecret resources with `waitForReconcileSuccess`

- [x] Task 3: Refactor `vaultsecret_controller_v2_test.go` (AC: 1, 2, 3, 5)
  - [x] 3.1: Replace bootstrap block with `SetupKVv2StackWithReader` call
  - [x] 3.2: Replace teardown block with `TeardownKVv2Stack` call
  - [x] 3.3: Replace inline `Eventually ReconcileSuccessful` checks with `waitForReconcileSuccess`

- [x] Task 4: Remove dead fixture (AC: 4)
  - [x] 4.1: Delete `test/randomsecret/v2/08-passwordpolicy-simple-password-policy-v2.yaml`

- [x] Task 5: Verify no regressions (AC: 3)
  - [x] 5.1: Run `make test` — unit tests pass
  - [x] 5.2: Run `make integration` — all 83+ specs pass with zero regressions

### Review Findings

- [x] [Review][Patch] Replace the remaining inline `Eventually(... ReconcileSuccessful ...)` block for `VaultSecret` creation with `waitForReconcileSuccess` to fully satisfy AC 5 and Task 3.3 [`controllers/vaultsecret_controller_v2_test.go:68`]

## Dev Notes

### The Duplication Problem

The KV v2 bootstrap stack is **copy-pasted 4 times** in `randomsecret_controller_test.go` (one per Context: retain, multi-key, multi-key recreate, refreshPeriod) and **1 time** in `vaultsecret_controller_v2_test.go`. Each copy is ~190 lines of boilerplate that:

1. Creates PasswordPolicy from `test/randomsecret/v2/00-passwordpolicy-simple-password-policy-v2.yaml` in `vaultAdminNamespaceName`
2. Waits for ReconcileSuccessful
3. Creates Policy from `01-policy-kv-engine-admin-v2.yaml` in `vaultAdminNamespaceName`
4. Waits for ReconcileSuccessful
5. Creates Policy from `04-policy-secret-writer-v2.yaml` in `vaultAdminNamespaceName`
6. Waits for ReconcileSuccessful
7. (Optionally) Creates Policy from `test/vaultsecret/v2/00-policy-secret-reader-v2.yaml` in `vaultAdminNamespaceName`
8. Creates KubernetesAuthEngineRole from `02-kubernetesauthenginerole-kv-engine-admin-v2.yaml` in `vaultAdminNamespaceName`
9. Waits for ReconcileSuccessful
10. Creates KubernetesAuthEngineRole from `05-kubernetesauthenginerole-secret-writer-v2.yaml` in `vaultAdminNamespaceName`
11. Waits for ReconcileSuccessful
12. (Optionally) Creates KubernetesAuthEngineRole from `test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml` in `vaultAdminNamespaceName`
13. Creates SecretEngineMount from `03-secretenginemount-kv-v2.yaml` in `vaultTestNamespaceName`
14. Waits for ReconcileSuccessful

The teardown is equally duplicated: delete each resource in reverse dependency order, each with an `Eventually` block that reads from Vault, marshals to JSON, and asserts nil.

### Two Variants — With and Without Reader Resources

| Variant | Used By | Extra Resources |
|---------|---------|-----------------|
| **Base stack** | multi-key, multi-key recreate, refreshPeriod Contexts | 6 resources: PasswordPolicy, 2 Policies, 2 KubernetesAuthEngineRoles, SecretEngineMount |
| **With reader** | retain Context, VaultSecret v2 test | Base + 2 more: secret-reader Policy + secret-reader KubernetesAuthEngineRole (from `test/vaultsecret/v2/`) |

The helper must support both variants. Suggested approach: `SetupKVv2Stack` for base, `SetupKVv2StackWithReader` adds the reader resources.

### Shared Helper Design

**File:** `controllers/integration_test_helpers_test.go`

This file uses `//go:build integration` and is in the `controllers` package (same as all other integration test files). It has access to the package-level `decoder`, `k8sIntegrationClient`, `ctx`, `vaultClient`, `vaultAdminNamespaceName`, `vaultTestNamespaceName` variables defined in `suite_integration_test.go`.

**Return struct pattern:**

```go
type KVv2Stack struct {
    PasswordPolicy       *redhatcopv1alpha1.PasswordPolicy
    PolicyKVEngineAdmin  *redhatcopv1alpha1.Policy
    PolicySecretWriter   *redhatcopv1alpha1.Policy
    PolicySecretReader   *redhatcopv1alpha1.Policy           // nil if not using reader variant
    RoleKVEngineAdmin    *redhatcopv1alpha1.KubernetesAuthEngineRole
    RoleSecretWriter     *redhatcopv1alpha1.KubernetesAuthEngineRole
    RoleSecretReader     *redhatcopv1alpha1.KubernetesAuthEngineRole // nil if not using reader variant
    SecretEngineMount    *redhatcopv1alpha1.SecretEngineMount
}
```

The teardown function takes this struct and deletes everything in reverse order.

**`waitForReconcileSuccess` helper:**

```go
func waitForReconcileSuccess[T client.Object](ctx context.Context, obj T, key types.NamespacedName, timeout, interval time.Duration) {
    Eventually(func() bool {
        err := k8sIntegrationClient.Get(ctx, key, obj)
        if err != nil {
            return false
        }
        conditionsAware, ok := any(obj).(redhatcopv1alpha1.ConditionsAware)
        if !ok {
            return false
        }
        for _, condition := range conditionsAware.GetConditions() {
            if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
                return true
            }
        }
        return false
    }, timeout, interval).Should(BeTrue())
}
```

**IMPORTANT:** Go 1.22 supports type parameters. All CRD types implement `ConditionsAware` interface. However, if generics cause complexity, a simpler approach is to use `func waitForReconcileSuccess(ctx context.Context, key types.NamespacedName, obj client.Object, getConditions func() []metav1.Condition, ...)` or just keep a non-generic version that accepts `client.Object` and does type assertion.

**Alternative simpler approach** (non-generic, avoids generics complexities with controller-runtime's `client.Object`):

```go
func waitForReconcileSuccessCondition(ctx context.Context, key types.NamespacedName, obj client.Object, timeout, interval time.Duration) {
    Eventually(func() bool {
        if err := k8sIntegrationClient.Get(ctx, key, obj); err != nil {
            return false
        }
        ca := obj.(redhatcopv1alpha1.ConditionsAware)
        for _, c := range ca.GetConditions() {
            if c.Type == vaultresourcecontroller.ReconcileSuccessful && c.Status == metav1.ConditionTrue {
                return true
            }
        }
        return false
    }, timeout, interval).Should(BeTrue())
}
```

This works because all CRD types implement `ConditionsAware` (they all have `GetConditions()` and `SetConditions()`).

[Source: api/v1alpha1/ — all types implement ConditionsAware interface]

### `waitForVaultCleanup` helper for teardown

The repeated teardown pattern can also be extracted:

```go
func waitForVaultCleanup(vaultPath string, timeout, interval time.Duration) {
    Eventually(func() error {
        secret, err := vaultClient.Logical().Read(vaultPath)
        if secret == nil {
            return nil
        }
        out, _ := json.Marshal(secret)
        return fmt.Errorf("secret is not nil %s", string(out))
    }, timeout, interval).Should(Succeed())
}
```

### Fixture Paths (Constants)

Define constants to avoid typos and make path changes one-line:

```go
const (
    fixturePasswordPolicyV2      = "../test/randomsecret/v2/00-passwordpolicy-simple-password-policy-v2.yaml"
    fixturePolicyKVEngineAdminV2  = "../test/randomsecret/v2/01-policy-kv-engine-admin-v2.yaml"
    fixtureRoleKVEngineAdminV2   = "../test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml"
    fixtureSEMKVv2               = "../test/randomsecret/v2/03-secretenginemount-kv-v2.yaml"
    fixturePolicySecretWriterV2  = "../test/randomsecret/v2/04-policy-secret-writer-v2.yaml"
    fixtureRoleSecretWriterV2    = "../test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml"
    fixturePolicySecretReaderV2  = "../test/vaultsecret/v2/00-policy-secret-reader-v2.yaml"
    fixtureRoleSecretReaderV2    = "../test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml"
)
```

### Dead Fixture to Remove

`test/randomsecret/v2/08-passwordpolicy-simple-password-policy-v2.yaml` is a byte-for-byte copy of `00-passwordpolicy-simple-password-policy-v2.yaml`. No code in the repo references `08-passwordpolicy`. Delete it.

### File Scope — Lines Before and After (Estimated)

| File | Before (lines) | After (est.) | Reduction |
|------|----------------|--------------|-----------|
| `randomsecret_controller_test.go` | 1718 | ~700 | ~60% |
| `vaultsecret_controller_v2_test.go` | 491 | ~200 | ~60% |
| `integration_test_helpers_test.go` | 0 (new) | ~150 | N/A |
| **Net** | **2209** | **~1050** | **~52%** |

### What NOT to Refactor

- **`vaultsecret_controller_test.go` (v1 tests)**: Uses different fixtures (`test/random-secret.yaml`, `test/vaultsecret/randomsecret-another-password.yaml`) with a different auth setup (v1 KV engine). The duplication pattern is less severe (single Context). Leave it as-is.
- **`databasesecretenginestaticrole_controller_test.go`**: Uses `test/databasesecretengine/database-random-secret.yaml` with a completely different stack (database engine, not KV v2). Not part of this refactoring.
- **Test logic within each Context**: Only extract the bootstrap/teardown boilerplate. The actual test assertions (create RandomSecret, verify Vault, delete, verify retain, etc.) stay inline — they are unique per Context and should remain readable.

### Behavior Preservation Constraints

- **Creation order must be preserved**: PasswordPolicy → Policies → Roles → SecretEngineMount. Roles depend on policies being in Vault.
- **Namespace assignment must be preserved**: Admin resources (PasswordPolicy, Policies, Roles) go to `vaultAdminNamespaceName`; test resources (SecretEngineMount) go to `vaultTestNamespaceName`.
- **Deletion order must be preserved**: SecretEngineMount → Roles → Policies → PasswordPolicy (reverse of creation, respecting dependencies).
- **Each Context must get fresh instances**: The helpers must decode fresh instances from YAML each time (not reuse Go objects across Contexts). This is how the current code works — each Context calls `decoder.Get*Instance()` independently.

### No `make manifests generate` Needed

This story modifies only integration test files and deletes a dead YAML fixture. No CRD types, controllers, or webhooks are changed.

### Project Structure Notes

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/integration_test_helpers_test.go` | New | Shared KV v2 bootstrap/teardown helpers + `waitForReconcileSuccess` + fixture path constants |
| 2 | `controllers/randomsecret_controller_test.go` | Modified | Replace 4x bootstrap + 4x teardown blocks with helper calls (~1000 lines removed) |
| 3 | `controllers/vaultsecret_controller_v2_test.go` | Modified | Replace 1x bootstrap + 1x teardown blocks with helper calls (~300 lines removed) |
| 4 | `test/randomsecret/v2/08-passwordpolicy-simple-password-policy-v2.yaml` | Deleted | Dead duplicate fixture (never referenced) |

### Previous Story Intelligence

**From Story 6.4 (most recent predecessor — Audit integration tests):**
- Last story before Epic 7 — codebase is clean with 83+ integration specs passing
- Used Ordered Describe block with shared state; this story does NOT change test structure, only extracts helpers
- All 53.7% coverage maintained

**From Epic 6 Retrospective:**
- "Continue detailed dev notes in story specs" — applied
- All retro action items are process-level, no technical changes needed before this story
- No dependencies on Epic 6 artifacts

**From Epic 2 Story 2.2 (RandomSecret update tests — original creator of these tests):**
- The RandomSecret integration tests were added in Epic 2 with the retain, multi-key, and recreate scenarios
- The refresh scenario was added later
- The duplication was inherited from the original pre-existing test structure and accumulated over time

### Git Intelligence (Recent Commits)

```
9fc8b3c Bmad epic 6 (#321)
7ce3e42 Merge pull request #320 from raffaelespazzoli/bmad-epic-5
d64b2b1 Complete Epic 5 retrospective and close epic
```

Codebase clean on main. All integration tests passing post-Epic 6 merge.

### References

- [Source: controllers/randomsecret_controller_test.go] — 1718-line file with 4 Contexts sharing identical KV v2 bootstrap (~190 lines each) and teardown (~100 lines each)
- [Source: controllers/vaultsecret_controller_v2_test.go] — 491-line file with 1 Context using the same KV v2 bootstrap + reader variant
- [Source: controllers/vaultsecret_controller_test.go] — V1 tests using different fixtures; NOT in scope
- [Source: controllers/suite_integration_test.go] — Package-level variables (decoder, k8sIntegrationClient, ctx, vaultClient, namespace names) available to helper file
- [Source: controllers/controllertestutils/decoder.go] — Decoder methods for all fixture types
- [Source: test/randomsecret/v2/] — KV v2 fixture YAML files (00-11)
- [Source: test/randomsecret/v2/08-passwordpolicy-simple-password-policy-v2.yaml] — Dead duplicate to delete
- [Source: test/vaultsecret/v2/00-policy-secret-reader-v2.yaml] — Reader policy used by retain/VaultSecret v2 tests
- [Source: test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml] — Reader role used by retain/VaultSecret v2 tests
- [Source: _bmad-output/project-context.md] — Integration test patterns, ConditionsAware interface
- [Source: _bmad-output/implementation-artifacts/epic-6-retro-2026-05-02.md] — Latest retrospective, no blockers

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (Cursor)

### Debug Log References

- Initial `waitForVaultCleanup` helper incorrectly checked the `err` return from `vaultClient.Logical().Read()`. When a SecretEngineMount is deleted, reading from its former path returns HTTP 400 with nil secret. The original code intentionally ignores the error and only checks if `secret == nil`. Fixed by using `_` for the error return.

### Completion Notes List

- Created `controllers/integration_test_helpers_test.go` (~170 lines) with shared helpers: `waitForReconcileSuccess`, `waitForVaultCleanup`, `SetupKVv2Stack`, `SetupKVv2StackWithReader`, `TeardownKVv2Stack`, `KVv2Stack` struct, and fixture path constants.
- Refactored `randomsecret_controller_test.go` from 1718 lines to ~380 lines (~78% reduction) by replacing 4x bootstrap blocks and 4x teardown blocks with helper calls, plus replacing inline reconcile-success checks with `waitForReconcileSuccess`.
- Refactored `vaultsecret_controller_v2_test.go` from 491 lines to ~150 lines (~69% reduction) using the same shared helpers.
- Deleted dead fixture `test/randomsecret/v2/08-passwordpolicy-simple-password-policy-v2.yaml` (byte-for-byte duplicate, never referenced).
- All 83 integration specs pass with zero regressions. Coverage maintained at 53.7%.
- Unit tests pass clean.

### File List

- `controllers/integration_test_helpers_test.go` — NEW — Shared KV v2 bootstrap/teardown helpers + waitForReconcileSuccess + fixture path constants
- `controllers/randomsecret_controller_test.go` — MODIFIED — Replaced 4x bootstrap + 4x teardown blocks with helper calls (~1340 lines removed)
- `controllers/vaultsecret_controller_v2_test.go` — MODIFIED — Replaced 1x bootstrap + 1x teardown blocks with helper calls (~340 lines removed)
- `test/randomsecret/v2/08-passwordpolicy-simple-password-policy-v2.yaml` — DELETED — Dead duplicate fixture
