# Story 2.2: Add Update Scenarios to RandomSecret Integration Tests

Status: done

## Story

As an operator developer,
I want integration tests that modify a RandomSecret spec and verify the reconciler updates the Vault secret,
So that the RandomSecret update path is validated.

## Acceptance Criteria

1. **Given** a RandomSecret with a `refreshPeriod` that has been successfully reconciled **When** I update the `SecretFormat` (e.g., switch from a password policy name to an inline password policy) **Then** on the next refresh cycle, the reconciler generates a new password using the updated format and writes it to Vault

2. **Given** a RandomSecret with a `refreshPeriod` that has been successfully reconciled **When** the refresh period elapses **Then** the reconciler regenerates the secret and the `ObservedGeneration` on the `ReconcileSuccessful` condition reflects the current generation

## Tasks / Subtasks

- [x] Task 1: Create a test fixture for a RandomSecret with `refreshPeriod` (AC: 1, 2)
  - [x] 1.1: Create `test/randomsecret/v2/11-randomsecret-refresh-v2.yaml` — a RandomSecret with `refreshPeriod: 15s`, `secretKey: password`, `passwordPolicyName: simple-password-policy-v2`, `isKVSecretsEngineV2: true`, using the `secret-writer-v2` role at path `test-vault-config-operator/kv-v2/data`
  - [x] 1.2: Verify the fixture passes webhook validation by checking the YAML is well-formed and meets the `validateEitherPasswordPolicyReferenceOrInline` constraint

- [x] Task 2: Add update scenario Context to the RandomSecret integration test (AC: 1, 2)
  - [x] 2.1: In `controllers/randomsecret_controller_test.go`, add a new `Context("When updating a RandomSecret with refreshPeriod")` block inside the existing `Describe("Random Secret controller for v2 secrets")`
  - [x] 2.2: Create the full dependency chain (PasswordPolicy, Policies, KubernetesAuthEngineRoles, SecretEngineMount) — reuse the same fixture paths as the existing "retain" Context
  - [x] 2.3: Create the RandomSecret from the new fixture (`11-randomsecret-refresh-v2.yaml`), wait for `ReconcileSuccessful=True`
  - [x] 2.4: Read the Vault secret via `vaultClient.Logical().Read()` at the RandomSecret's path, extract the `password` value from the `data` sub-map (KV v2 returns `{"data": {"password": "..."}}`), and record it as `initialPassword`
  - [x] 2.5: Verify `initialPassword` matches the `simple-password-policy-v2` pattern (20-char lowercase: `regexp.MustCompile("^[a-z]{20}$")`)
  - [x] 2.6: Record the initial `ObservedGeneration` from the `ReconcileSuccessful` condition

- [x] Task 3: Perform the spec update and verify Vault reflects the change (AC: 1)
  - [x] 3.1: `Get()` the latest RandomSecret from the API (required for fresh ResourceVersion)
  - [x] 3.2: Update the spec: clear `Spec.SecretFormat.PasswordPolicyName` and set `Spec.SecretFormat.InlinePasswordPolicy` to a distinguishable policy (10-char uppercase):
    ```
    length = 10
    rule "charset" {
      charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
      min-chars = 10
    }
    ```
  - [x] 3.3: Call `k8sIntegrationClient.Update(ctx, instance)` — this triggers the `needsCreation` predicate (SecretFormat changed) and bumps generation
  - [x] 3.4: Use `Eventually` (timeout 60s, interval 2s) to poll the Vault secret until the `password` key value differs from `initialPassword` — the refreshPeriod (15s) must elapse before the reconciler writes
  - [x] 3.5: Verify the new password matches the inline policy pattern (10-char uppercase: `regexp.MustCompile("^[A-Z]{10}$")`)

- [x] Task 4: Verify ObservedGeneration incremented (AC: 2)
  - [x] 4.1: Use `Eventually` to poll the RandomSecret CR until `ReconcileSuccessful` condition's `ObservedGeneration` is greater than the initial value recorded in Task 2.6
  - [x] 4.2: Verify the condition `Status` is `metav1.ConditionTrue`

- [x] Task 5: Clean up all resources (AC: 1, 2)
  - [x] 5.1: Delete RandomSecret, verify removal from K8s
  - [x] 5.2: Delete SecretEngineMount, verify engine disabled in Vault
  - [x] 5.3: Delete KubernetesAuthEngineRoles, verify removed from Vault
  - [x] 5.4: Delete Policies, verify removed from Vault
  - [x] 5.5: Delete PasswordPolicy, verify removed from Vault
  - [x] 5.6: Follow the exact cleanup pattern from the existing "retain" Context (delete in reverse dependency order, poll Vault to confirm deletion)

### Review Findings

- [x] [Review][Patch] Prove the update is refresh-gated rather than merely eventual [controllers/randomsecret_controller_test.go:1178]
- [x] [Review][Patch] Assert `ReconcileSuccessful.ObservedGeneration` matches the current object generation, not just a larger value [controllers/randomsecret_controller_test.go:1212]

## Dev Notes

### RandomSecret Has a Unique Update Mechanism — NOT Like Standard VaultObject Types

RandomSecret does NOT follow the standard `IsEquivalentToDesiredState` → conditional-write pattern. Its `IsEquivalentToDesiredState` **always returns `false`** (line 140 of `randomsecret_types.go`). Instead, the reconcile flow is:

1. **Guard check (controller line 101):** If `LastVaultSecretUpdate` is set AND (no `RefreshPeriod` OR refresh time hasn't elapsed), return immediately — **no write**.
2. **Reconcile logic:** Generate a new password via `PrepareInternalValues` → `CreateOrMergeKV` to Vault.
3. **Requeue:** If `RefreshPeriod > 0`, requeue for the next refresh cycle.

**Critical implication:** Without `RefreshPeriod`, a RandomSecret can NEVER be updated in-place — the guard at line 101 always returns early because `LastVaultSecretUpdate` is set and `RefreshPeriod` is nil. The only way to get a new secret is delete + recreate. **The test MUST use a `refreshPeriod`** to exercise the update path.

[Source: controllers/randomsecret_controller.go#L100-L103]
[Source: api/v1alpha1/randomsecret_types.go#L140-L142]

### Custom Predicate Limits What Spec Changes Trigger Reconcile

The RandomSecret controller uses a custom `needsCreation` predicate (AND-composed with `PeriodicReconcilePredicate`). The `UpdateFunc` returns `true` only for:
- Deletion timestamp changed
- `Spec.RefreshPeriod` changed
- `Spec.SecretFormat` changed (DeepEqual comparison)

Changing `Spec.SecretKey` alone does **NOT** trigger the predicate. However, with `RefreshPeriod` set, the requeue timer fires regardless, and the next reconcile uses the updated spec.

For this test, we change `SecretFormat` (switches password policy), which both triggers the predicate AND produces an observably different password format.

[Source: controllers/randomsecret_controller.go#L172-L195]

### Update Flow Timeline (What the Test Observes)

1. **T=0s:** RandomSecret created → reconcile fires → password generated → written to Vault → `LastVaultSecretUpdate` set → requeue after 15s
2. **T=5s (approx):** Spec updated (SecretFormat changed) → predicate fires → reconcile runs → **guard blocks** (only ~5s elapsed of 15s RefreshPeriod) → `return reconcile.Result{}, nil`
3. **T=15s:** Requeue fires → reconcile runs → guard passes → new password generated with **new SecretFormat** → written to Vault
4. **T=15s+:** Test polls Vault → sees new password → passes

The test should use a 60s `Eventually` timeout (4x the refresh period) to account for timing variance.

### KV v2 Read Response Structure

For KV v2, the Vault `Logical().Read()` response wraps data under a `data` sub-map:

```json
{
  "data": {
    "data": {
      "password": "abcdefghijklmnopqrst"
    },
    "metadata": { ... }
  }
}
```

To extract the password:
```go
secret, _ := vaultClient.Logical().Read(rsInstance.GetPath())
data := secret.Data["data"].(map[string]interface{})
passwordValue := data["password"].(string)
```

The existing tests (multikey Context, lines 763-771) demonstrate this exact pattern.

[Source: controllers/randomsecret_controller_test.go#L763-L771]

### CreateOrMergeKV Merges Keys (Does Not Replace)

When the RandomSecret writes to Vault, it uses `CreateOrMergeKV` which merges the new payload into the existing KV secret. This means:
- Initial write: `{password: "abc"}` → Vault has `{password: "abc"}`
- Update with same secretKey: `{password: "xyz"}` → Vault has `{password: "xyz"}` (value replaced)
- If secretKey changed to `newkey`: `{newkey: "xyz"}` → Vault has `{password: "abc", newkey: "xyz"}` (merged)

For this test, we keep `secretKey: password` and only change the password policy, so the `password` key is overwritten with the new format.

### Switching from PasswordPolicyName to InlinePasswordPolicy

The test updates `SecretFormat` from using a Vault-side `passwordPolicyName` to an `InlinePasswordPolicy`. This:
1. Triggers the predicate (SecretFormat changed)
2. Produces an observably different password (10-char uppercase vs 20-char lowercase)
3. Avoids needing a second PasswordPolicy CR (inline policy is self-contained)

The webhook's `validateEitherPasswordPolicyReferenceOrInline` requires exactly one of the two. When switching, **clear the old field before setting the new one**:

```go
created.Spec.SecretFormat = redhatcopv1alpha1.VaultPasswordPolicy{
    InlinePasswordPolicy: "length = 10\nrule \"charset\" {\n  charset = \"ABCDEFGHIJKLMNOPQRSTUVWXYZ\"\n  min-chars = 10\n}",
}
```

This assigns a new struct, implicitly clearing `PasswordPolicyName`.

[Source: api/v1alpha1/randomsecret_types.go#L329-L341]

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/randomsecret_controller_test.go` | Modify | Add new `Context("When updating a RandomSecret with refreshPeriod")` block with the update test |
| 2 | `test/randomsecret/v2/11-randomsecret-refresh-v2.yaml` | New | RandomSecret fixture with `refreshPeriod: 15s` |

No decoder changes needed — `GetRandomSecretInstance` already exists.

No new dependency fixtures needed — the test reuses existing `00-*` through `05-*` fixtures from `test/randomsecret/v2/`.

### Test Structure — New Context Block

Add the update test as a **third `Context` block** inside the existing `Describe("Random Secret controller for v2 secrets")`. This follows the established pattern where each Context has its own independent dependency chain. Do NOT modify the two existing Contexts.

The existing Contexts are:
1. `"When recreating a random secret with retain policy"` (lines 28–578)
2. `"When creating multiple RandomSecrets contributing to the same Vault path"` (lines 581–957)

The new Context:
3. `"When updating a RandomSecret with refreshPeriod"` — add after line 957

### Dependency Chain (Same as Existing Tests)

The update test reuses the same dependency fixtures as the existing "retain" Context:

1. `PasswordPolicy` — `../test/randomsecret/v2/00-passwordpolicy-simple-password-policy-v2.yaml`
2. `Policy` (kv-engine-admin) — `../test/randomsecret/v2/01-policy-kv-engine-admin-v2.yaml`
3. `Policy` (secret-writer) — `../test/randomsecret/v2/04-policy-secret-writer-v2.yaml`
4. `KubernetesAuthEngineRole` (kv-engine-admin) — `../test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml`
5. `KubernetesAuthEngineRole` (secret-writer) — `../test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml`
6. `SecretEngineMount` (kv-v2) — `../test/randomsecret/v2/03-secretenginemount-kv-v2.yaml`
7. `RandomSecret` (with refreshPeriod) — `../test/randomsecret/v2/11-randomsecret-refresh-v2.yaml`

Items 1–5 go in `vault-admin` namespace; items 6–7 in `test-vault-config-operator` namespace.

Note: no Policy/Role for secret-reader or VaultSecret needed — this test only writes to Vault, doesn't read back via VaultSecret.

### Get Before Update — Critical Pattern

When calling `k8sIntegrationClient.Update()`, the object MUST have the latest `ResourceVersion`. Always call `Get()` immediately before `Update()`. If a reconcile modifies the status between `Get` and `Update`, the API server returns a conflict error. The test should handle this by getting the latest object right before the update call.

Pattern established by Story 2.1:
```go
By("Getting the latest RandomSecret")
Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())

By("Updating the RandomSecret spec")
created.Spec.SecretFormat = redhatcopv1alpha1.VaultPasswordPolicy{...}
Expect(k8sIntegrationClient.Update(ctx, created)).Should(Succeed())
```

### Timing Considerations

- `refreshPeriod: 15s` means the first refresh happens ~15s after initial reconcile
- The spec update at ~5s won't take effect immediately (guard blocks)
- The requeue at ~15s picks up the new spec
- Use `Eventually(func() ..., 60*time.Second, 2*time.Second)` for the Vault check
- Standard 120s timeout for other reconcile checks

### No `make manifests generate` Needed

This story only modifies test files and adds a YAML fixture. No CRD types, controllers, or webhooks are changed.

### Previous Story Intelligence

**From Story 2.1 (ready-for-dev):**
- Story 2.1 established the update test pattern for VaultSecret: Get → modify spec → Update → poll for change → verify ObservedGeneration
- VaultSecret's update mechanism is different (shouldSync/manageSyncLogic), but the test structure (Get-before-Update, Eventually polling, ObservedGeneration verification) applies here
- Story 2.1 recommends inserting update steps into existing test flows; for RandomSecret, a new Context is cleaner due to the refreshPeriod requirement

**From Story 2.0 (ready-for-dev):**
- Story 2.0 stabilizes integration test infrastructure (idempotent Kind, namespace handling)
- Story 2.0 MUST complete before this story
- The namespace create-or-get pattern prevents re-run failures

**From Epic 1 Retrospective:**
- "Pattern-first investment pays dividends" — reuse the dependency chain setup pattern from existing Contexts
- Epic 2 update scenarios exercise the reconcile pipeline end-to-end
- Story 7-4 (extra-field hardening) is not a blocker for this story

### Git Intelligence (Recent Commits)

```
910acbd Complete Epic 1 retrospective and fix identified tech debt
f1e57e7 Update push.yaml with permissions for nested workflow
cd7e5b8 Pre-load busybox image into kind to avoid Docker Hub rate limits in CI
511af21 Fix helmchart-test hang: add wget timeout and fix sidecar script portability
9110587 Add integration test philosophy rule and Story 2.0 for infrastructure stabilization
```

Commit `910acbd` resolved GroupAlias debug prints and KubernetesSecretEngineRole field mapping. Codebase is clean for Epic 2.

### Risk Considerations

- **RefreshPeriod timing sensitivity:** The test depends on the refresh cycle firing within the Eventually timeout. Use a 15s refresh with 60s timeout (4x margin). If CI is slow, the refresh may take longer than expected.
- **Inline password policy parsing:** The `GenerateNewPassword` method uses `hclsimple.Decode` for inline policies. The HCL syntax must be exact. Test the inline policy string format carefully.
- **KV v2 merge behavior:** After the update, the Vault secret at the same path/key is overwritten (not merged at key level for same key). Verify by checking the value format changed, not by checking for key absence.
- **Resource conflicts on Update:** A reconcile that fires between Get and Update can cause a conflict. The test should do Get→Update without delay. If flaky, wrap in a retry loop.

### References

- [Source: controllers/randomsecret_controller_test.go] — Existing RandomSecret integration tests (958 lines, two Contexts)
- [Source: controllers/randomsecret_controller.go#L100-L103] — Reconcile guard (LastVaultSecretUpdate + RefreshPeriod)
- [Source: controllers/randomsecret_controller.go#L149-L168] — manageReconcileLogic (PrepareInternalValues + CreateOrMergeKV)
- [Source: controllers/randomsecret_controller.go#L172-L195] — Custom predicate (SecretFormat and RefreshPeriod changes only)
- [Source: api/v1alpha1/randomsecret_types.go#L40-L90] — RandomSecretSpec (all fields)
- [Source: api/v1alpha1/randomsecret_types.go#L131-L142] — GetPayload and IsEquivalentToDesiredState (always false)
- [Source: api/v1alpha1/randomsecret_types.go#L225-L257] — GenerateNewPassword (PasswordPolicyName vs InlinePasswordPolicy)
- [Source: api/v1alpha1/randomsecret_types.go#L315-L327] — isValid (validateEitherPasswordPolicyReferenceOrInline)
- [Source: test/randomsecret/v2/08-randomsecret-retain-v2.yaml] — Existing fixture (no refreshPeriod)
- [Source: test/randomsecret/v2/03-secretenginemount-kv-v2.yaml] — KV v2 engine mount fixture
- [Source: _bmad-output/implementation-artifacts/2-1-add-update-scenarios-to-vaultsecret-integration-tests.md] — Story 2.1 (update pattern reference)
- [Source: _bmad-output/implementation-artifacts/2-0-stabilize-integration-test-infrastructure.md] — Story 2.0 (prerequisite)
- [Source: _bmad-output/planning-artifacts/epics.md#L305-L316] — Story 2.2 epic definition
- [Source: _bmad-output/implementation-artifacts/epic-1-retro-2026-04-15.md] — Epic 1 retrospective

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (Cursor)

### Debug Log References

- Initial run: inline password policy failed with `Unsupported file format; Cannot read from password: unrecognized file format suffix ""` — `hclsimple.Decode` uses the filename parameter for format detection, and `d.Spec.SecretKey` ("password") has no `.hcl` extension.
- Fix: Changed `hclsimple.Decode` first argument from `d.Spec.SecretKey` to `"policy.hcl"` in `randomsecret_types.go:228`. This is a bug fix — the filename is only used for format detection, not file I/O.
- After fix: all 14 integration tests pass (287s).

### Completion Notes List

- Created YAML fixture `test/randomsecret/v2/11-randomsecret-refresh-v2.yaml` with `refreshPeriod: 15s`
- Added third Context block "When updating a RandomSecret with refreshPeriod" to the RandomSecret integration test
- Test creates full dependency chain, creates RandomSecret, reads initial password, updates SecretFormat from PasswordPolicyName to InlinePasswordPolicy, waits for refresh cycle, verifies new password format, verifies ObservedGeneration incremented, and cleans up all resources
- Fixed pre-existing bug in `GenerateNewPassword` where `hclsimple.Decode` was passed `d.Spec.SecretKey` as filename (no extension), causing inline password policy parsing to fail
- AC 1 verified: after updating SecretFormat and waiting for refresh, Vault secret reflects the new 10-char uppercase format
- AC 2 verified: ObservedGeneration on ReconcileSuccessful condition increments after the update

### Change Log

- 2026-04-16: Implemented Story 2.2 — added RandomSecret update scenario integration test and fixed inline password policy parsing bug

### File List

- `test/randomsecret/v2/11-randomsecret-refresh-v2.yaml` (new) — RandomSecret fixture with refreshPeriod
- `controllers/randomsecret_controller_test.go` (modified) — Added "When updating a RandomSecret with refreshPeriod" Context
- `api/v1alpha1/randomsecret_types.go` (modified) — Fixed hclsimple.Decode filename parameter for inline policy parsing
