# Deferred Work

## Deferred from: code review of story 1-1 (2026-04-12)

- **AC #4 contradiction**: Story 1.1 AC #4 states that extra unmanaged Vault fields should be ignored (return `true`), but 7 of 9 types use `reflect.DeepEqual` which returns `false` when extra keys are present. The tests document this existing behavior. Resolution requires a product decision: amend AC #4 to match reality, or change the `IsEquivalentToDesiredState` implementations to filter extra keys. This is tracked in Epic 7, Story 7-4 ("audit Vault API responses and harden IsEquivalentToDesiredState extra-field handling"). Additionally, the `VaultEndpoint.CreateOrUpdate()` path does **not** pre-filter the Vault read response before calling `IsEquivalentToDesiredState` — the pre-filtering claim in the story Dev Notes is incorrect.

## Deferred from: code review of story 1-6 (2026-04-13)

- **GroupAlias debug prints leak into test and runtime output**: `api/v1alpha1/groupalias_types.go` still contains `fmt.Print("desired state", ...)` and `fmt.Print("actual state", ...)` inside `GroupAlias.IsEquivalentToDesiredState()`. Story 1.6 correctly documented and tolerated this pre-existing production behavior rather than changing it, but it should be cleaned up when that code path is taken on directly.

## Deferred from: code review of 2-3-add-update-scenarios-to-entity-and-entityalias-integration-tests (2026-04-17)

- **ObservedGeneration assertion could be stronger**: Tests assert `ObservedGeneration > initial` (satisfies spec), but a stronger assertion `ObservedGeneration == metadata.generation` would more precisely verify the controller contract. Beyond spec requirement; defer to a test hardening pass.
- **Redundant Get before update**: Both update tests call `Get()` twice before `Update()` (once to record ObservedGeneration, once labeled "before update"). Minor clean-up; not a bug.
- **Post-update `disabled` field not re-verified in Vault**: After the Entity update, only policies and metadata are checked; `disabled` is not re-asserted. Spec doesn't require it; low risk.
- **Vault paths hardcoded as `"identity/entity/name/test-entity"`**: Should ideally derive from the CR `entityInstance.Name`. Not a bug (name is fixture-driven), but coupling concern for when fixtures change.
- **Duplicated `wait for ReconcileSuccessful` loop**: The Eventually polling pattern appears 3+ times per file. Pre-existing codebase pattern; extract to a helper when refactoring integration tests.
- **No conflict/retry on `k8sIntegrationClient.Update()`**: Single-shot update can produce 409 Conflict flakes under concurrent reconciliation. Pre-existing pattern across all Epic 2 stories; address in a broader test hardening story.
- **EntityAlias Vault deletion verified after both K8s deletes**: Timing dependency between K8s finalizer processing and Vault deletion is implicit. Low risk given sequential Ginkgo execution; document or add a sequential Vault delete check if flakiness observed.
