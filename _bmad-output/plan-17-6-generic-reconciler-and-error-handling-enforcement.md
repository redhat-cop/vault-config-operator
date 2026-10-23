---
title: 'Generic Reconciler + Error Handling Enforcement'
type: 'refactor'
ticket: ''
created: '2026-10-22'
status: 'built'
baseline_revision: '1d2652578784dbb09c181f42ae534ee3d1a85c36'
route: 'full'
route_source: 'auto'
risk: 'medium'
review: 'quick'
review_source: 'pinned'
lenses_ran: ['quick']
review_loop_iteration: 0
followup_review_recommended: false
context:
  - '_bmad-output/implementation-artifacts/17-6-generic-reconciler-and-error-handling-enforcement.md'
  - '_bmad-output/project-context.md'
warnings: ['oversized']
deferred: []
---

<intent-contract>

## Intent

**Problem:** ~65 controller files duplicate an identical 30-line `Reconcile()` body (Get → NotFound → prepareContext → NewVaultResource → Reconcile). Adding a new CRD requires copying this boilerplate. Additionally, `errcheck` compliance needs verification for code added since the R1.2a pass.

**Approach:** Create a Go-generics-based `GenericVaultReconciler[T, PT]` in `vaultresourcecontroller` with a `VaultResourceFactory` parameter. Move `prepareContext` + `VaultAuthenticableResource` to `vaultresourcecontroller`. Migrate ~65 standard controllers to a slim `SetupXxxController(mgr)` function (RBAC markers + registration call). Keep ~12 custom controllers untouched except for the `prepareContext` import path. Fix any errcheck violations found by `make lint`. Update contributor docs.

## Boundaries & Constraints

**Always:**
- RBAC kubebuilder markers MUST remain in per-type controller files (parsed by controller-gen)
- All existing unit and integration tests must pass with identical behavior
- `config/rbac/role.yaml` and `config/crd/bases/` must be identical before/after migration
- Shared framework files: do NOT modify existing behavior in `reconcile_skeleton.go`, `vaultresourcereconciler.go`, `vaultengineresourcereconciler.go`, `vaultauditresourcereconciler.go`, `vaultpkiengineresourcereconciler.go`, `vaultobject.go`, `commons.go` (existing behavior) — only ADD new exports
- Every errcheck fix must log the error with context before returning/handling

**Never:**
- Do not migrate custom controllers (Category 5: VaultSecret, RandomSecret, DatabaseSecretEngineConfig, TransitSecretEngineKey, RabbitMQSecretEngineConfig, AWSSecretEngineConfig, ConsulSecretEngineConfig, GCPSecretEngineConfig, LDAPSecretEngineConfig, MongoDBAtlasSecretEngineConfig, NomadSecretEngineConfig, TerraformCloudSecretEngineConfig, TOTPSecretEngineKey)
- Do not change webhook registration or CRD types
- Do not add new dependencies

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Standard controller reconcile | CR created for a generic-migrated type (e.g. Policy) | Identical reconcile flow: Get → prepareContext → NewVaultResource → Reconcile | Same NotFound/error handling as current |
| Engine mount reconcile | AuthEngineMount CR created | Generic reconciler uses VaultEngineResourceFactory | Same error handling |
| Custom controller reconcile | DatabaseSecretEngineConfig CR created | Hand-written Reconcile unchanged, only prepareContext import updated | No change |
| New CRD type added | Developer follows updated guide | Creates slim controller with `SetupGenericController` call | N/A |

</intent-contract>

## Code Map

- `internal/controller/vaultresourcecontroller/generic_reconciler.go` -- **CREATE**: `Reconcilable` interface, `VaultResourceFactory` type, factory wrappers (`StandardVaultResourceFactory`, `VaultEngineResourceFactory`, `VaultAuditResourceFactory`, `VaultPKIEngineResourceFactory`), `GenericVaultReconciler[T, PT]` struct + `Reconcile` method, `SetupGenericController` + `SetupGenericControllerWithWatches` helpers, compile-time interface checks
- `internal/controller/vaultresourcecontroller/utils.go` -- **MODIFY**: Add `PrepareContext` (exported) + `VaultAuthenticableResource` interface (moved from commons.go). Keep all existing functions unchanged.
- `internal/controller/commons.go` -- **MODIFY**: Remove `prepareContext` + `VaultAuthenticableResource` (moved to vaultresourcecontroller). Replace with thin wrapper calling the new location for backward compat, or update all callers directly.
- `internal/controller/policy_controller.go` (and ~46 other pure standard controllers) -- **MODIFY**: Remove `XxxReconciler` struct + `Reconcile` method + `SetupWithManager`. Replace with `SetupXxxController(mgr)` calling `SetupGenericController`. Keep RBAC markers.
- `internal/controller/ldapauthengineconfig_controller.go` (and ~14 other custom-SetupWithManager controllers) -- **MODIFY**: Remove `Reconcile` method. Replace with `SetupXxxController(mgr)` that creates `GenericVaultReconciler` manually and builds controller with custom `Watches`. Keep RBAC markers, keep custom `SetupWithManager` watch logic, keep `findApplicable*` helper methods.
- `internal/controller/authenginemount_controller.go`, `secretenginemount_controller.go` -- **MODIFY**: Migrate to generic with `VaultEngineResourceFactory`.
- `internal/controller/audit_controller.go` -- **MODIFY**: Migrate to generic with `VaultAuditResourceFactory`.
- `internal/controller/pkisecretengineconfig_controller.go` -- **MODIFY**: Migrate to generic with `VaultPKIEngineResourceFactory`.
- `internal/controller/*_controller.go` (~12 custom controllers) -- **MODIFY**: Update `prepareContext` call to use `vaultresourcecontroller.PrepareContext` import.
- `cmd/main.go` -- **MODIFY**: Replace ~65+ registration blocks with `controller.SetupXxxController(mgr)` calls. Keep custom controller registrations + webhook registrations.
- `internal/controller/vaultresourcecontroller/vaultresourcereconciler.go` -- READ ONLY: `NewVaultResource` — verify Reconcile signature matches `Reconcilable`
- `internal/controller/vaultresourcecontroller/vaultengineresourcereconciler.go` -- READ ONLY: `NewVaultEngineResource` — verify Reconcile signature
- `internal/controller/vaultresourcecontroller/vaultauditresourcereconciler.go` -- READ ONLY: `NewVaultAuditResource` — verify Reconcile signature
- `internal/controller/vaultresourcecontroller/vaultpkiengineresourcereconciler.go` -- READ ONLY: `NewVaultPKIEngineResource` — verify Reconcile signature
- `internal/controller/vaultresourcecontroller/reconcile_skeleton.go` -- READ ONLY: `ReconcileWithFunctions` — do not modify
- `docs/contributing-vault-apis.md` -- **MODIFY**: Update steps 6-7 for generic reconciler pattern
- `_bmad-output/project-context.md` -- **MODIFY**: Update "Adding a New Vault API Type" checklist steps 5-7
- `.golangci.yml` -- READ ONLY: Confirm errcheck enabled via `default: standard`

## Tasks & Acceptance

**Execution:**
- [ ] `internal/controller/vaultresourcecontroller/generic_reconciler.go` -- Create: define `Reconcilable` interface, `VaultResourceFactory`, four factory wrappers, `GenericVaultReconciler[T, PT]` with Reconcile method, `SetupGenericController` and `SetupGenericControllerWithWatches` helpers, compile-time `var _ Reconcilable = ...` checks for all 4 reconciler types
- [ ] `internal/controller/vaultresourcecontroller/utils.go` -- Add exported `PrepareContext` function and `VaultAuthenticableResource` interface (moved from commons.go); keep all existing code
- [ ] `internal/controller/commons.go` -- Remove `prepareContext` and `VaultAuthenticableResource`; update callers or leave thin delegation stub
- [ ] `internal/controller/policy_controller.go` (proof-of-concept) -- Migrate to slim `SetupPolicyController(mgr)` with RBAC markers preserved; run `make manifests generate fmt vet test` and diff `config/rbac/role.yaml`
- [ ] Remaining ~46 pure standard controllers -- Migrate each to `SetupXxxController` pattern; batch by alphabet, test after each batch
- [ ] ~15 custom-SetupWithManager controllers -- Replace only Reconcile body with `GenericVaultReconciler` instantiation; keep custom `SetupWithManager` watch logic and `findApplicable*` helpers
- [ ] Engine mount controllers (AuthEngineMount, SecretEngineMount) -- Migrate with `VaultEngineResourceFactory`
- [ ] `internal/controller/audit_controller.go` -- Migrate with `VaultAuditResourceFactory`
- [ ] `internal/controller/pkisecretengineconfig_controller.go` -- Migrate with `VaultPKIEngineResourceFactory`
- [ ] ~12 custom controllers (Category 5) -- Update only the `prepareContext` call to `vaultresourcecontroller.PrepareContext`
- [ ] `cmd/main.go` -- Replace ~65+ registration blocks with `controller.SetupXxxController(mgr)` calls; keep custom + webhook registrations
- [ ] Run `make lint` -- Capture any errcheck violations; fix all with error logging context
- [ ] `docs/contributing-vault-apis.md` -- Update steps 6 (controller) and 7 (main.go) for generic pattern; keep old pattern as "custom controller" subsection
- [ ] `_bmad-output/project-context.md` -- Update steps 5-7 in "Adding a New Vault API Type" checklist; remove step 7 (decoder, already generic)
- [ ] Final validation -- `make manifests generate fmt vet test lint` all green; diff `config/rbac/role.yaml` and `config/crd/bases/` identical to baseline

**Acceptance Criteria:**
- Given a standard CRD type (e.g. Policy), when reconciled after migration, then behavior is identical to pre-migration (all unit tests pass)
- Given `make manifests` is run after migration, then `config/rbac/role.yaml` is byte-identical to baseline
- Given `make manifests` is run after migration, then `config/crd/bases/` is identical to baseline
- Given `make lint` is run, then zero errcheck violations
- Given a new developer follows `docs/contributing-vault-apis.md`, then the guide shows the `SetupGenericController` pattern as default

## Implementation Notes


## Plan Change Log


## Review Triage Log

### 2026-10-22 — Review pass
- verdicts: 3 findings — high 0, medium 0, low 1, false 2, maybe-false 0
- findings:
  - `[false]` `[reject]` GenericVaultReconciler.Reconcile() returns raw Get() error without ManageOutcome — Refuted: the generic reconciler faithfully replicates the existing standard controller pattern. All 96 pre-migration controllers returned the raw Get error identically. ManageOutcome requires a valid instance for status update, which is unavailable on Get failure. The project-context rule is aspirational; no controller implemented it.
  - `[false]` `[reject]` SetControllerReference errcheck fix continues on error, may create orphaned resources — Refuted: the original code used `_ =` (explicit discard). The story mandates errcheck compliance (log the error), not control flow changes. Returning an error would alter CreateOrUpdateResource's behavior, violating the shared framework protection rule. The pre-existing behavior of continuing is intentional.
  - `[low]` `[patch]` project-context.md still references commons.go as canonical home of prepareContext in 4 sections — Applied: updated §Context-Carried Values, §Reconcile Flow, §Code Organization, and §Context Value Contract to reference vaultresourcecontroller.PrepareContext as canonical location with backward-compat wrapper note.


## Design Notes

**Generic reconciler pattern:** The `GenericVaultReconciler[T, PT]` uses Go 1.18+ type constraints: `T` is the value type, `PT` is `*T` constrained to `client.Object + VaultObject + ConditionsAware`. The factory parameter (`VaultResourceFactory`) selects which reconciler variant to use (standard, engine, audit, PKI). This is proven in the codebase — R1.4 already uses `DecodeInstance[T]` generics successfully.

**prepareContext move:** `prepareContext` operates on `ReconcilerBase` and `VaultObject`, both already in `vaultresourcecontroller`. Moving it there is logically correct. Export it as `PrepareContext` so the generic reconciler (same package) and custom controllers (different package) can both call it.

**Controller file after migration:** ~20 lines — RBAC markers + `SetupXxxController(mgr)` function. Custom-SetupWithManager controllers keep their watch predicates and `findApplicable*` helpers alongside the `Setup` function.

**Factory wrappers bridge the interface gap:** Existing `NewVaultResource` returns `*VaultResource`, not `Reconcilable`. The factory wrappers (`StandardVaultResourceFactory` etc.) return `Reconcilable`, hiding the concrete type.

## Verification

**Commands:**
- `make manifests generate fmt vet` -- expected: zero errors, identical generated output
- `make test` -- expected: all unit tests pass
- `make lint` -- expected: zero violations
- `diff <(git show HEAD:config/rbac/role.yaml) config/rbac/role.yaml` -- expected: no diff
- `diff -r <(git archive HEAD config/crd/bases/ | tar -xO) config/crd/bases/` -- expected: no diff

## Auto Run Result

**Summary:** Implemented Story 17.6 — Generic Reconciler + Error Handling Enforcement. Created `GenericVaultReconciler[T, PT]` with 4 factory wrappers and migrated 83 controllers (62 pure standard, 17 custom-watch, 2 engine mount, 1 audit, 1 PKI) from ~80-line boilerplate to slim `SetupXxxController()` functions. 13 custom controllers untouched. `main.go` refactored from ~440 registration lines to compact loop + 13 custom blocks. Net: −4,328 lines. Errcheck fix applied. Docs updated.

**Files changed (90 files):**
- `internal/controller/vaultresourcecontroller/generic_reconciler.go` — NEW: Reconcilable interface, VaultResourceFactory, 4 factory wrappers, GenericVaultReconciler[T,PT], SetupGenericController, SetupGenericControllerWithBuilder
- `internal/controller/vaultresourcecontroller/utils.go` — Added exported PrepareContext + VaultAuthenticableResource; fixed errcheck on SetControllerReference
- `internal/controller/commons.go` — Replaced with thin delegation wrapper to vaultresourcecontroller.PrepareContext
- `internal/controller/*_controller.go` (83 files) — Migrated to SetupXxxController pattern with preserved RBAC markers
- `cmd/main.go` — Registration refactored to compact loop + 13 custom blocks
- `docs/contributing-vault-apis.md` — Updated steps 6-7 for generic reconciler pattern
- `_bmad-output/project-context.md` — Updated checklist steps 5-7, fixed prepareContext location references, added generic_reconciler.go to framework protection list

**Review findings breakdown:**
- 3 findings from GPT 5.4 quick review
- 2 rejected as false (generic reconciler Get error handling matches existing pattern; errcheck fix preserves intended behavior per framework protection rule)
- 1 low patched (project-context.md doc references updated)
- 0 deferred
- Patched counts by verdict: low 1

**Follow-up review recommendation:** false (no high patches, only 1 low patch)

**Verification performed:**
- `make manifests generate fmt vet` — zero errors
- `make test` — all unit tests pass (7 packages, 25.5% coverage api/v1alpha1)
- `make lint` — 0 issues (errcheck compliant)
- `config/rbac/role.yaml` — no diff vs baseline (byte-identical)
- `config/crd/bases/` — no diff vs baseline (identical)

**Residual risks:**
- Integration tests not run (require live Kind cluster with Vault). Migration is purely structural — identical runtime behavior — but integration test verification recommended before merge.
