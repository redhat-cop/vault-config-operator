# Story 2.3: Add Update Scenarios to Entity and EntityAlias Integration Tests

Status: ready-for-dev

## Story

As an operator developer,
I want integration tests that modify Entity/EntityAlias specs and verify reconciliation,
So that identity update paths are validated.

## Acceptance Criteria

1. **Given** an Entity that has been successfully reconciled **When** I update the Entity spec (e.g., change metadata or policies) **Then** the reconciler detects the change, `IsEquivalentToDesiredState` returns false, and the Entity is updated in Vault **And** the ReconcileSuccessful condition is updated with a new ObservedGeneration

2. **Given** an EntityAlias that has been successfully reconciled **When** I update the EntityAlias spec (e.g., change customMetadata) **Then** the reconciler updates the alias in Vault **And** the ReconcileSuccessful condition is updated with a new ObservedGeneration

## Tasks / Subtasks

- [ ] Task 1: Add update scenario to Entity integration test (AC: 1)
  - [ ] 1.1: In `controllers/entity_controller_test.go`, add a new `Context("When updating an Entity")` block inside the existing `Describe("Entity controller")`
  - [ ] 1.2: Create the Entity from `../test/identity/01-entity-sample.yaml`, set namespace to `vaultAdminNamespaceName`, wait for `ReconcileSuccessful=True`
  - [ ] 1.3: Read the Entity from Vault via `vaultClient.Logical().Read("identity/entity/name/test-entity")` and verify the initial state: `metadata` contains `{"team":"engineering","environment":"test"}`, `policies` contains `["default"]`, `disabled` is `false`
  - [ ] 1.4: Record the initial `ObservedGeneration` from the `ReconcileSuccessful` condition
  - [ ] 1.5: `Get()` the latest Entity from the API (fresh ResourceVersion), update `Spec.Policies` to `["default", "kv-reader"]` and add a new metadata key `Spec.Metadata["owner"] = "integration-test"`, then call `k8sIntegrationClient.Update(ctx, instance)`
  - [ ] 1.6: Use `Eventually` (timeout 120s, interval 2s) to poll Vault at `identity/entity/name/test-entity` until the `policies` field contains both `default` and `kv-reader`
  - [ ] 1.7: Verify the Vault entity also has `metadata["owner"] == "integration-test"` and the original metadata keys are preserved
  - [ ] 1.8: Verify the `ReconcileSuccessful` condition's `ObservedGeneration` is greater than the initial value

- [ ] Task 2: Verify Entity cleanup (AC: 1)
  - [ ] 2.1: Delete the Entity, wait for the object to be removed from K8s
  - [ ] 2.2: Verify the entity is deleted from Vault by reading `identity/entity/name/test-entity` and confirming 404

- [ ] Task 3: Add update scenario to EntityAlias integration test (AC: 2)
  - [ ] 3.1: In `controllers/entityalias_controller_test.go`, add a new `Context("When updating an EntityAlias")` block inside the existing `Describe("EntityAlias controller")`
  - [ ] 3.2: Create the Entity from `../test/identity/01-entity-sample.yaml`, wait for `ReconcileSuccessful=True` (EntityAlias depends on Entity)
  - [ ] 3.3: Create the EntityAlias from `../test/identity/02-entityalias-sample.yaml`, wait for `ReconcileSuccessful=True`
  - [ ] 3.4: Verify `Status.ID` is not empty
  - [ ] 3.5: Read the EntityAlias from Vault via `vaultClient.Logical().Read("identity/entity-alias/id/" + created.Status.ID)` and verify the initial `custom_metadata` contains `{"contact":"admin@example.com"}`
  - [ ] 3.6: Record the initial `ObservedGeneration` from the `ReconcileSuccessful` condition
  - [ ] 3.7: `Get()` the latest EntityAlias from the API, update `Spec.CustomMetadata` to `{"contact":"admin@example.com", "purpose":"integration-test"}`, call `k8sIntegrationClient.Update(ctx, instance)`
  - [ ] 3.8: Use `Eventually` (timeout 120s, interval 2s) to poll Vault at `identity/entity-alias/id/{Status.ID}` until `custom_metadata["purpose"]` equals `"integration-test"`
  - [ ] 3.9: Verify the `ReconcileSuccessful` condition's `ObservedGeneration` is greater than the initial value

- [ ] Task 4: Verify EntityAlias and Entity cleanup (AC: 2)
  - [ ] 4.1: Delete the EntityAlias first, wait for removal from K8s
  - [ ] 4.2: Delete the Entity, wait for removal from K8s
  - [ ] 4.3: Verify both are cleaned up from Vault (EntityAlias by ID, Entity by name)

## Dev Notes

### Entity and EntityAlias Use the Standard VaultResource Reconcile Flow

Both Entity and EntityAlias use the standard `VaultResource` reconciler (not `VaultEngineResource` or `VaultAuditResource`). The reconcile flow is:

1. `prepareContext()` — enriches context with kubeClient, restConfig, vaultConnection, vaultClient
2. `vaultResource.Reconcile(ctx, instance)` — delegates to `VaultEndpoint.CreateOrUpdate()`
3. `CreateOrUpdate()` reads current Vault state via `read(ctx, path)`, calls `IsEquivalentToDesiredState(currentPayload)`, and writes only if `false`
4. `ManageOutcome()` sets `ReconcileSuccessful` condition with `ObservedGeneration`

When the spec is updated via the K8s API, the generation bumps, the `PeriodicReconcilePredicate` passes (generation changed), reconcile runs, reads Vault, `IsEquivalentToDesiredState` returns `false` (policies/metadata differ), and writes the new state.

[Source: api/v1alpha1/utils/vaultobject.go#L144-L159]
[Source: controllers/entity_controller.go]
[Source: controllers/entityalias_controller.go]

### Entity's IsEquivalentToDesiredState — Custom Extra-Field Deletion

Entity explicitly `delete()`s 11 Vault-added keys before `reflect.DeepEqual`:
- `name`, `id`, `aliases`, `creation_time`, `last_update_time`, `merged_entity_ids`, `direct_group_ids`, `group_ids`, `inherited_group_ids`, `namespace_id`, `bucket_key_hash`

The remaining comparison is between `toMap()` output (`metadata`, `policies`, `disabled`) and the filtered Vault payload. When we change `policies` from `["default"]` to `["default", "kv-reader"]`, the Vault read will still show the old value, `IsEquivalentToDesiredState` returns `false`, and the write proceeds.

**Vault API behavior for Entity:** Writing to `identity/entity/name/{name}` with a payload updates the existing entity (merge semantics — new fields added, existing fields updated).

[Source: api/v1alpha1/entity_types.go#L160-L174]

### EntityAlias's IsEquivalentToDesiredState and PrepareInternalValues — Complex Internal State

EntityAlias has a fundamentally different lifecycle from Entity:

1. **`PrepareInternalValues`** does significant work on every reconcile:
   - Reads `sys/auth/{authEngineMountPath}` to get the mount accessor
   - Reads `/identity/entity/name/{entityName}` to get the entity's canonical ID
   - On **first creation** (`Status.ID == ""`): creates the alias via `POST /identity/entity-alias`, stores the ID in `Status.ID`, and updates the K8s status
   - Sets `retrievedName`, `retrievedAliasID`, `retrievedMountAccessor`, `retrievedCanonicalID` internal fields

2. **`toMap()`** uses the `retrieved*` internal fields (not spec fields directly):
   - `name` = `retrievedName` (from spec.Name or metadata.Name)
   - `id` = `retrievedAliasID` (from Status.ID)
   - `mount_accessor` = `retrievedMountAccessor` (looked up each reconcile)
   - `canonical_id` = `retrievedCanonicalID` (looked up each reconcile)
   - `custom_metadata` = `CustomMetadata` (only if non-empty)

3. **`IsEquivalentToDesiredState`** deletes 8 Vault-added keys:
   - `creation_time`, `last_update_time`, `merged_from_canonical_ids`, `metadata`, `mount_path`, `mount_type`, `local`, `namespace_id`

4. **`GetPath()`** returns `/identity/entity-alias/id/{Status.ID}` — uses the Status ID, not a name-based path.

**Critical for update test:** The only user-visible spec field that `toMap()` exposes is `customMetadata`. Changing `authEngineMountPath` or `entityName` would change the resolved accessor/canonical_id, but these are structural changes that would break the alias semantics. **The safe update for testing is changing `customMetadata`** — this flows directly through `toMap()` into the comparison.

[Source: api/v1alpha1/entityalias_types.go#L139-L149]
[Source: api/v1alpha1/entityalias_types.go#L155-L214]

### No Webhooks Exist for Entity or EntityAlias

Unlike most CRD types, Entity and EntityAlias have **no webhook files** (`entity_webhook.go` / `entityalias_webhook.go` do not exist). This means:
- No `spec.path` immutability rule (these types don't have a `spec.path` field anyway — Entity uses a name-based Vault path, EntityAlias uses Status.ID-based path)
- No validation on update (any spec change is accepted)
- No defaulting

This simplifies the update test — no webhook will reject the update.

### Vault API Read Paths and Response Shapes

**Entity read:** `GET identity/entity/name/{name}` returns:
```json
{
  "data": {
    "id": "entity-uuid",
    "name": "test-entity",
    "metadata": {"team": "engineering", "environment": "test"},
    "policies": ["default"],
    "disabled": false,
    "aliases": [...],
    "creation_time": "...",
    "last_update_time": "...",
    "merged_entity_ids": null,
    "direct_group_ids": [...],
    "group_ids": [...],
    "inherited_group_ids": null,
    "namespace_id": "root",
    "bucket_key_hash": "..."
  }
}
```

**EntityAlias read:** `GET identity/entity-alias/id/{id}` returns:
```json
{
  "data": {
    "id": "alias-uuid",
    "name": "test-entity-alias",
    "canonical_id": "entity-uuid",
    "mount_accessor": "auth_kubernetes_...",
    "mount_path": "auth/kubernetes/",
    "mount_type": "kubernetes",
    "custom_metadata": {"contact": "admin@example.com"},
    "creation_time": "...",
    "last_update_time": "...",
    "local": false,
    "namespace_id": "root"
  }
}
```

Use `secret.Data["policies"]` etc. to extract fields from the Vault read response. The `Logical().Read()` returns `*vault.Secret` where `Data` is `map[string]interface{}`.

### Accessing the Vault Client in Integration Tests

The integration test suite provides the `vaultClient` variable (type `*vault.Client`) configured in `BeforeSuite`. This is the same client pattern used by other integration tests that read from Vault directly.

To read Entity state from Vault:
```go
secret, err := vaultClient.Logical().Read("identity/entity/name/test-entity")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
policies := secret.Data["policies"]
```

**Important:** Vault returns `[]interface{}` for list fields, not `[]string`. Use type assertion or `gomega` matchers that handle interface slices:
```go
policies := secret.Data["policies"].([]interface{})
Expect(policies).To(ContainElement("kv-reader"))
```

For metadata (Vault returns `map[string]interface{}`):
```go
metadata := secret.Data["metadata"].(map[string]interface{})
Expect(metadata["owner"]).To(Equal("integration-test"))
```

### Get Before Update — Critical Pattern

Always call `Get()` immediately before `Update()` to get the latest ResourceVersion. A reconcile that modifies status between your Get and Update will cause a conflict error from the API server.

Pattern from Story 2.1/2.2:
```go
By("Getting the latest Entity")
Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())

By("Updating the Entity spec")
created.Spec.Policies = []string{"default", "kv-reader"}
created.Spec.Metadata["owner"] = "integration-test"
Expect(k8sIntegrationClient.Update(ctx, created)).Should(Succeed())
```

### Verifying Vault State After Update via Polling

After calling `Update()`, the reconciler doesn't fire synchronously. Use `Eventually` to poll Vault until the state changes:

```go
Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("identity/entity/name/test-entity")
    if err != nil || secret == nil {
        return false
    }
    policies, ok := secret.Data["policies"].([]interface{})
    if !ok {
        return false
    }
    for _, p := range policies {
        if p == "kv-reader" {
            return true
        }
    }
    return false
}, timeout, interval).Should(BeTrue())
```

### EntityAlias Update — customMetadata Is the Only Safe Spec Field to Change

EntityAlias `toMap()` produces: `name`, `id`, `mount_accessor`, `canonical_id`, and optionally `custom_metadata`. The first four are resolved internally each reconcile:
- `name` = metadata.name (or spec.Name override) — immutable
- `id` = Status.ID — set once on creation, immutable
- `mount_accessor` = resolved from `authEngineMountPath` each reconcile
- `canonical_id` = resolved from `entityName` each reconcile

Changing `authEngineMountPath` or `entityName` would make the alias point to a different entity/mount, which is a destructive operation. The only safe user-facing field change is `customMetadata`.

**Test approach:** Add a new key to `CustomMetadata` and verify it appears in Vault.

### Test Fixtures — Reuse Existing

Both tests reuse existing fixtures:
- `test/identity/01-entity-sample.yaml` — Entity with `metadata: {team: engineering, environment: test}`, `policies: [default]`, `disabled: false`
- `test/identity/02-entityalias-sample.yaml` — EntityAlias with `authEngineMountPath: kubernetes`, `entityName: test-entity`, `customMetadata: {contact: admin@example.com}`

No new fixtures needed. Updates are performed in-code by modifying the spec after creation.

### Test Structure — New Context Blocks

Add update tests as **new `Context` blocks** after the existing create-delete Contexts. Do NOT modify the existing test assertions or flow.

**Entity test** (`entity_controller_test.go`): Add `Context("When updating an Entity")` after the existing `Context("When creating an Entity")` (line 59).

**EntityAlias test** (`entityalias_controller_test.go`): Add `Context("When updating an EntityAlias")` after the existing `Context("When creating an EntityAlias")` (line 95).

Each new Context is independent — creates its own resources, performs the update, verifies, then cleans up.

### EntityAlias Deletion Verification — Use Status.ID

When verifying EntityAlias cleanup from Vault, read by the Status.ID that was captured during the test:
```go
secret, err := vaultClient.Logical().Read("identity/entity-alias/id/" + capturedAliasID)
```
After deletion, this should return nil or a 404.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/entity_controller_test.go` | Modify | Add `Context("When updating an Entity")` with update + Vault verification |
| 2 | `controllers/entityalias_controller_test.go` | Modify | Add `Context("When updating an EntityAlias")` with update + Vault verification |

No new fixtures needed. No decoder changes needed — `GetEntityInstance` and `GetEntityAliasInstance` already exist. No controller or type changes needed.

### No `make manifests generate` Needed

This story only modifies integration test files. No CRD types, controllers, or webhooks are changed.

### Import Requirements for Updated Test Files

The Entity test will need additional imports for Vault client access:
```go
import (
    "time"
    vault "github.com/hashicorp/vault/api"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)
```

Check whether `vaultClient` is already accessible from the test file scope (declared in `suite_integration_test.go`). If not, the test needs to construct one from the `VAULT_ADDR` and `VAULT_TOKEN` environment variables. Look at how other integration tests (e.g., `randomsecret_controller_test.go`) access the Vault client.

### Previous Story Intelligence

**From Story 2.2 (ready-for-dev):**
- Established the update test pattern: Create → Wait → Get-before-Update → Modify spec → Update → Poll Vault → Verify ObservedGeneration → Cleanup
- RandomSecret has a unique update mechanism (refreshPeriod guard), but Entity/EntityAlias follow the standard VaultResource flow — much simpler

**From Story 2.1 (ready-for-dev):**
- First update scenario test in the codebase. Established the `Get` → modify → `Update` → `Eventually` poll pattern
- VaultSecret is also non-standard (no VaultObject); Entity is a standard VaultObject so the update path is the straightforward `CreateOrUpdate` → `IsEquivalentToDesiredState` check

**From Story 2.0 (ready-for-dev):**
- Story 2.0 stabilizes integration test infrastructure (idempotent Kind, namespace handling)
- Story 2.0 MUST complete before this story
- Namespace create-or-get pattern prevents re-run failures

**From Epic 1 Retrospective:**
- "Pattern-first investment pays dividends" — this story can directly reuse the update pattern established by Stories 2.1 and 2.2
- Entity and EntityAlias unit tests (Story 1.6) verified `IsEquivalentToDesiredState` strips correct Vault-only keys and correctly detects changes
- Story 7-4 (extra-field hardening) is not a blocker — Entity already handles extra fields via explicit `delete()` calls

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

- **Vault policies list type:** Vault returns `[]interface{}` for `policies`, not `[]string`. Use `ContainElement("kv-reader")` matcher or iterate with type assertions. Entity's `toMap()` sends `[]string` but `IsEquivalentToDesiredState` compares via `reflect.DeepEqual` — Vault's `[]interface{}` vs Go's `[]string` could cause a false negative. However, this is exactly the existing behavior that the unit tests (Story 1.6) documented. If the integration test passes the create phase, the update phase will work the same way. If Vault returns `[]interface{}{"default"}` and `toMap()` returns `[]string{"default"}`, `reflect.DeepEqual` returns `false`, causing an unnecessary write every reconcile. This is the known AC#4 issue deferred to Story 7-4. For this test, it means the reconciler always writes (even when nothing changed), but the update test still verifies the correct flow: spec change → reconcile → Vault reflects new state.
- **EntityAlias custom_metadata type:** Verify whether Vault returns `custom_metadata` as `map[string]interface{}` or `map[string]string`. The `toMap()` sends `map[string]string`. If Vault returns `map[string]interface{}`, `reflect.DeepEqual` would return `false` every time (same AC#4 issue). The test should verify the end-to-end behavior regardless.
- **Entity metadata map type:** Similarly, `metadata` in `toMap()` is `map[string]string` but Vault may return `map[string]interface{}`. Same consideration.
- **Resource conflicts on Update:** A reconcile between Get and Update can cause a conflict. Do Get→Update without delay. If flaky, the standard pattern is to retry.
- **EntityAlias requires Entity to exist first:** The update test must create the Entity before the EntityAlias, and delete EntityAlias before Entity. Respect dependency ordering.

### References

- [Source: controllers/entity_controller_test.go] — Existing Entity integration test (60 lines, single create-delete Context)
- [Source: controllers/entityalias_controller_test.go] — Existing EntityAlias integration test (95 lines, single create-delete Context)
- [Source: api/v1alpha1/entity_types.go#L132-L138] — Entity `toMap()` (metadata, policies, disabled)
- [Source: api/v1alpha1/entity_types.go#L160-L174] — Entity `IsEquivalentToDesiredState` (11 Vault-only key deletions)
- [Source: api/v1alpha1/entityalias_types.go#L139-L149] — EntityAlias `toMap()` (retrieved* fields + customMetadata)
- [Source: api/v1alpha1/entityalias_types.go#L155-L214] — EntityAlias `PrepareInternalValues` (accessor lookup, entity ID lookup, alias creation)
- [Source: api/v1alpha1/entityalias_types.go#L229-L240] — EntityAlias `IsEquivalentToDesiredState` (8 Vault-only key deletions)
- [Source: api/v1alpha1/utils/vaultobject.go#L144-L159] — `VaultEndpoint.CreateOrUpdate()` (read → IsEquivalent → conditional write)
- [Source: test/identity/01-entity-sample.yaml] — Entity test fixture
- [Source: test/identity/02-entityalias-sample.yaml] — EntityAlias test fixture
- [Source: api/v1alpha1/entity_test.go] — Entity unit tests (Story 1.6)
- [Source: api/v1alpha1/entityalias_test.go] — EntityAlias unit tests (Story 1.6)
- [Source: _bmad-output/implementation-artifacts/2-2-add-update-scenarios-to-randomsecret-integration-tests.md] — Story 2.2 (update pattern reference)
- [Source: _bmad-output/implementation-artifacts/2-1-add-update-scenarios-to-vaultsecret-integration-tests.md] — Story 2.1 (update pattern reference)
- [Source: _bmad-output/implementation-artifacts/2-0-stabilize-integration-test-infrastructure.md] — Story 2.0 (prerequisite)
- [Source: _bmad-output/planning-artifacts/epics.md#L317-L333] — Story 2.3 epic definition
- [Source: _bmad-output/implementation-artifacts/epic-1-retro-2026-04-15.md] — Epic 1 retrospective

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### Change Log

### File List
