# Deferred Work

## Deferred from: code review of story 1-1 (2026-04-12)

- **AC #4 contradiction**: Story 1.1 AC #4 states that extra unmanaged Vault fields should be ignored (return `true`), but 7 of 9 types use `reflect.DeepEqual` which returns `false` when extra keys are present. The tests document this existing behavior. Resolution requires a product decision: amend AC #4 to match reality, or change the `IsEquivalentToDesiredState` implementations to filter extra keys. This is tracked in Epic 7, Story 7-4 ("audit Vault API responses and harden IsEquivalentToDesiredState extra-field handling"). Additionally, the `VaultEndpoint.CreateOrUpdate()` path does **not** pre-filter the Vault read response before calling `IsEquivalentToDesiredState` — the pre-filtering claim in the story Dev Notes is incorrect.

## Deferred from: code review of story 1-6 (2026-04-13)

- **GroupAlias debug prints leak into test and runtime output**: `api/v1alpha1/groupalias_types.go` still contains `fmt.Print("desired state", ...)` and `fmt.Print("actual state", ...)` inside `GroupAlias.IsEquivalentToDesiredState()`. Story 1.6 correctly documented and tolerated this pre-existing production behavior rather than changing it, but it should be cleaned up when that code path is taken on directly.
