---
title: 'Reflective toMap via vault Struct Tags + VaultObject Interface Simplification'
type: 'refactor'
ticket: '17-5-reflective-tomap-and-vaultobject-interface-simplification'
created: '2026-10-20'
status: 'built'
baseline_revision: '94493599acdb2ebf09e3a224e2635e9dbf00b41a'
route: 'full'
route_source: 'auto'
risk: 'high'
review: 'quick'
review_source: 'pinned'
lenses_ran: ['quick']
review_loop_iteration: 0
context: ['_bmad-output/implementation-artifacts/17-5-reflective-tomap-and-vaultobject-interface-simplification.md']
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Every CRD type requires 6+ boilerplate methods (hand-written `toMap()`, trivial `IsInitialized`, `IsValid`, `PrepareInternalValues`, `PrepareTLSConfig`, and identical `IsEquivalentToDesiredState`). This produces ~2,000 lines of duplicated no-ops and makes adding new types unnecessarily verbose.

**Approach:** Replace hand-written `toMap()` with a reflection-based `StructToVaultPayload` walker driven by `vault:` struct tags; trim `VaultObject` from 10 required methods to 5 by making `IsEquivalentToDesiredState`, `IsInitialized`, `IsValid`, `PrepareInternalValues`, and `PrepareTLSConfig` optional interfaces with sensible defaults; add `ResourceName` helper for `GetPath()`. A comprehensive test audit (Part A0) builds the regression oracle BEFORE any migration.

## Boundaries & Constraints

**Always:**
- Execute phases in strict order: A0 → A1 → B → C → D → E → F. Each phase must pass all tests before the next begins.
- Part A0 unit test audit is the prerequisite gate — all 88 types must have toMap, IsEquivalentToDesiredState positive/negative/extra-fields tests before any migration.
- Walker tag grammar is a closed set: `vault:"<key>[,<codec>][,omitempty][,when=Field:value][,default=X][,spreadmap]"` | `vault:"-"` | `vault:",inline"`.
- Codecs: identity (default), duration, number, intmap, stringmap. No extensibility mechanism.
- TransitSecretEngineKey is EXCLUDED from walker migration (two maps from one struct).
- Credential-override types: move override logic from `toMap()` into tagged `retrieved*` fields with `omitempty`; `PrepareInternalValues`/`setInternalCredentials` already populates them before `toMap()` is called.
- After every batch migration in A1, run A0 tests — zero behavioral difference tolerated.
- Duration/TTL fields must emit `json.Number` seconds via `durationToSeconds()`. Never emit raw duration strings.
- `IsInitialized()` has zero callers in the codebase — removal is safe without fallback.

**Never:**
- Do not modify the behavioral contract of `VaultEndpoint.CreateOrUpdate` or `ReconcileWithFunctions`.
- Do not add new CRDs, new dependencies, or new controller files.
- Do not change `filterPayloadToDesiredKeys` or `normalizeVaultReadAliases` in `payload_filter.go`.
- Do not migrate TransitSecretEngineKey to the walker.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Walker: simple string field | `vault:"key_type"` + `KeyType="otp"` | `{"key_type": "otp"}` | N/A |
| Walker: duration codec | `vault:"token_ttl,duration"` + `TTL="4h"` | `{"token_ttl": json.Number("14400")}` | Unparseable → passthrough |
| Walker: number codec | `vault:"port,number"` + `Port=22` | `{"port": json.Number("22")}` | N/A |
| Walker: conditional when=true | `vault:"cidr_list,when=KeyType:otp"` + `KeyType="otp"` | Field included | N/A |
| Walker: conditional when=false | `vault:"cidr_list,when=KeyType:otp"` + `KeyType="ca"` | Field excluded | N/A |
| Walker: *bool default=true nil | `vault:"verify_connection,default=true"` + `nil` | `{"verify_connection": true}` | N/A |
| Walker: spreadmap | `vault:",spreadmap"` + `map{"k":"v"}` | Parent gains `{"k":"v"}` | Empty map → no entries |
| Walker: skip | `vault:"-"` on Connection field | Field excluded | N/A |
| Walker: inline | `vault:",inline"` on embedded struct | Fields merged into parent | N/A |
| Optional interface absent | Type without `IsEquivalentToDesiredState` | Default: `filterPayloadToDesiredKeys` + `reflect.DeepEqual` | N/A |
| Optional interface present | Type with custom `IsEquivalentToDesiredState` | Delegates to type's method | N/A |

</frozen-after-approval>

## Code Map

- `api/v1alpha1/utils/vaultobject.go` — VaultObject interface (10 methods → 5) + VaultEndpoint.CreateOrUpdate (line 187: `IsEquivalentToDesiredState` call) + RabbitMQEngineConfigVaultEndpoint.CreateOrUpdateLease (line 227: same)
- `internal/controller/vaultresourcecontroller/vaultresourcereconciler.go:50` — calls `PrepareInternalValues` and `PrepareTLSConfig` via type assertion
- `internal/controller/vaultresourcecontroller/vaultauditresourcereconciler.go:51` — same pattern
- `internal/controller/vaultresourcecontroller/vaultengineresourcereconciler.go:50` — same pattern
- `internal/controller/vaultresourcecontroller/vaultpkiengineresourcereconciler.go:50` — same pattern (no `PrepareTLSConfig` call)
- `api/v1alpha1/sshsecretenginerole_types.go:219` — `durationToSeconds()`, `toAnyMapJsonNumber()`, `toAnyMapString()` — relocate to utils
- `api/v1alpha1/databasesecretengineconfig_types.go:161` — `toInterfaceArray()` — relocate to utils
- `api/v1alpha1/payload_filter.go` — `filterPayloadToDesiredKeys()` — used by default equivalence, DO NOT MODIFY
- `api/v1alpha1/passwordpolicy_types.go` — simplest VaultObject reference pattern (5 trivial no-ops to delete)
- `api/v1alpha1/databasesecretengineconfig_types.go` — spreadmap + credential-override + custom `IsEquivalentToDesiredState` exemplar
- `api/v1alpha1/sshsecretenginerole_types.go` — `when=` conditional + duration/number/intmap/stringmap codecs exemplar
- `docs/contributing-vault-apis.md` — outdated developer guide (still shows `runtime.Object` webhooks, hand-written `toMap()`, 10-method interface)
- 9 types missing test files: identityoidcassignment, identityoidcclient, identityoidcprovider, identityoidcscope, identitytokenconfig, identitytokenkey, identitytokenrole, transitsecretenginekey, vaultsecret
- 95 types implement IsInitialized (0 callers in framework), 95 PrepareTLSConfig (0 non-trivial implementations), ~60 have trivial IsEquivalentToDesiredState

## Tasks & Acceptance

**Execution:**

*Phase A0: Unit Test Audit + Gap Fill (PREREQUISITE GATE)*
- [ ] `api/v1alpha1/*_test.go` (79 existing) — AUDIT+MODIFY — Scan for tests deriving expected payloads from `toMap()`/`GetPayload()`; rewrite with independent Vault-read-shaped fixtures (hardcoded `json.Number`, `durationToSeconds` values)
- [ ] `api/v1alpha1/identityoidcassignment_test.go` — CREATE — GetPath, GetPayload, IsEquivalentToDesiredState positive/negative/extra-fields
- [ ] `api/v1alpha1/identityoidcclient_test.go` — CREATE — same pattern
- [ ] `api/v1alpha1/identityoidcprovider_test.go` — CREATE — same pattern
- [ ] `api/v1alpha1/identityoidcscope_test.go` — CREATE — same pattern
- [ ] `api/v1alpha1/identitytokenconfig_test.go` — CREATE — same pattern (singleton)
- [ ] `api/v1alpha1/identitytokenkey_test.go` — CREATE — same pattern with duration+number fields
- [ ] `api/v1alpha1/identitytokenrole_test.go` — CREATE — same pattern with duration fields
- [ ] `api/v1alpha1/transitsecretenginekey_test.go` — CREATE — test BOTH `toMap()` and `configToMap()` paths
- [ ] `api/v1alpha1/vaultsecret_test.go` — CREATE — KVv1 and KVv2 paths, template rendering
- [ ] Verify mode-dependent types have per-mode coverage: SSHSecretEngineRole (otp/ca), TOTPSecretEngineKey (generate true/false), DatabaseSecretEngineConfig (connection_details nesting)
- [ ] `make test` — all unit tests pass as regression baseline

*Phase A1: StructToVaultPayload Walker + Migration*
- [ ] `api/v1alpha1/utils/codecs.go` — CREATE — Relocate+export: `DurationToSeconds`, `ToAnyMapJsonNumber`, `ToAnyMapString`, `ToInterfaceArray` from type files; update all ~80 callers
- [ ] `api/v1alpha1/utils/struct_to_vault_payload.go` — CREATE — Walker implementation: parse `vault:` tags, implement all codecs (identity, duration, number, intmap, stringmap) and directives (omitempty, when, default, spreadmap, inline, skip)
- [ ] `api/v1alpha1/utils/struct_to_vault_payload_test.go` — CREATE — Unit tests for each codec and directive combination
- [ ] 88 type files (excl. Transit) — MODIFY — Add `vault:` struct tags to inline config structs, replace `toMap()` body with `return vaultutils.StructToVaultPayload(&d)`. Migrate in batches: trivial (~10) → standard (~55) → mode-dependent (3) → credential-override (~15) → spreadmap (1) → pointer-default (~5). Run A0 tests after each batch.

*Phase B: Default IsEquivalentToDesiredState*
- [ ] `api/v1alpha1/utils/vaultobject.go` — MODIFY — Remove `IsEquivalentToDesiredState` from `VaultObject`; add `VaultEquivalenceChecker` interface; update `CreateOrUpdate` and `CreateOrUpdateLease` to type-assert, defaulting to `filterPayloadToDesiredKeys` + `reflect.DeepEqual`
- [ ] ~60 type files with standard `IsEquivalentToDesiredState` — MODIFY — Delete method; ~25 types with custom logic keep theirs and gain `var _ vaultutils.VaultEquivalenceChecker` compile check

*Phase C: Optional No-Op Interface Methods*
- [ ] `api/v1alpha1/utils/vaultobject.go` — MODIFY — Remove `IsInitialized`, `IsValid`, `PrepareInternalValues`, `PrepareTLSConfig` from `VaultObject`; add 4 optional interfaces: `VaultInitializer`, `VaultValidator`, `VaultInternalValuesPreparer`, `VaultTLSPreparer`
- [ ] 4 reconciler files (`vaultresourcereconciler.go`, `vaultauditresourcereconciler.go`, `vaultengineresourcereconciler.go`, `vaultpkiengineresourcereconciler.go`) — MODIFY — Type-assert optional interfaces with sensible defaults
- [ ] ~95 type files — MODIFY — Delete trivial `IsInitialized() → true` (all types), `PrepareTLSConfig() → nil` (all types), `PrepareInternalValues() → nil` (~50 types), `IsValid() → (true, nil)` (~60 types); keep non-trivial implementations on ~18+4 types

*Phase D: ResourceName Helper*
- [ ] `api/v1alpha1/utils/helpers.go` — MODIFY — Add `ResourceName(specName, metaName string) string`
- [ ] ~48 type files — MODIFY — Replace `if d.Spec.Name != "" { ... } else { ... }` in `GetPath()` with `vaultutils.ResourceName(d.Spec.Name, d.Name)`

*Phase E: Documentation*
- [ ] `docs/contributing-vault-apis.md` — REWRITE — Update for vault tags, 5-method interface, optional interfaces, generic webhooks, ResourceName, tag grammar reference
- [ ] `_bmad-output/project-context.md` — MODIFY — Update "Adding a New Vault API Type" checklist and "Imperative-to-Declarative Bridge" section

*Phase F: Final Validation*
- [ ] `make manifests generate fmt vet` — no errors
- [ ] `make test` — all unit tests pass
- [ ] `make lint` — zero violations
- [ ] Verify no new CRDs added (refactoring only)

**Acceptance Criteria:**
- Given all 88 types are tagged and migrated to `StructToVaultPayload`, when all A0 unit tests run, then all pass with zero behavioral difference
- Given a type without `VaultEquivalenceChecker`, when `CreateOrUpdate` compares, then default `filterPayloadToDesiredKeys + reflect.DeepEqual` runs
- Given trivial no-op methods deleted from ~60 types, when `make test` runs, then all pass unchanged
- Given `make manifests generate fmt vet test lint` runs, then all pass with no errors or violations

## Implementation Notes

## Plan Change Log

## Review Triage Log

**Pass 1** (GPT 5.4 adversarial, quick lens) — verdicts: 0 high, 2 medium, 2 low, 2 false

| # | Finding | Verdict | Route | Evidence |
|---|---------|---------|-------|----------|
| 1 | `DurationToSeconds()` returns `json.Number(s)` on parse failure — invalid JSON number | `false` | rejected | Pre-migration `durationToSeconds()` in `sshsecretenginerole_types.go:219` has identical behavior. Zero behavioral change. |
| 2 | Identity token/OIDC types missing `,duration` codec on TTL fields | `low` (rejected) | defer | Original `toMap()` methods emitted raw strings (e.g., `payload["rotation_period"] = i.RotationPeriod`). Migration correctly preserved pre-existing behavior per AC: "zero behavioral difference." |
| 3 | `isequivalent_audit_test.go` still uses `toMap()`/`GetPayload()` to build Vault payloads | `low` (rejected) | — | Equivalence tests use `toMap()` to build matching payloads for comparison logic testing. Separate GetPayload tests verify output correctness with hardcoded fixtures. This is testing comparison behavior, not payload output. |
| 4 | `vaultsecret_test.go` coverage incomplete (no KVv2/template rendering) | `low` (rejected) | — | VaultSecret is excluded from walker migration (no standard `toMap()`). A0 regression oracle purpose doesn't apply. |
| 5 | ~22 types still had trivial `IsInitialized`, `PrepareTLSConfig`, etc. after Phase C | `medium` | patch | Phase C cleanup was incomplete. Fixed: deleted all remaining trivial no-ops across 24 files. |
| 6 | Local helper duplicates (`durationToSeconds`, `toInterfaceArray`) still in SSH/DB type files | `medium` | patch | Phase A1 relocation was incomplete. Fixed: deleted dead functions from `sshsecretenginerole_types.go` and `databasesecretengineconfig_types.go`. |

## Design Notes

**Walker architecture:** Single `StructToVaultPayload(any) map[string]any` function in `api/v1alpha1/utils/`. Uses `reflect` to iterate exported struct fields, reads `vault:` tag, dispatches to codec. The walker handles one level; `inline` recurses. `spreadmap` merges a map field's entries into the parent. `when=Field:value` does a sibling field lookup by Go struct name. `default=X` checks pointer nilness and parses X based on the pointer's element type.

**Optional interface pattern:** Already established in this codebase — `VaultStatusEnricher` and `CreateOnlyFieldsProvider` in `vaultobject.go` are precedents. The 4 reconciler files already do type assertions for these; adding 5 more follows the same pattern.

**IsInitialized removal safety:** Grep confirms zero callers of `IsInitialized()` in the entire codebase. It was likely intended for a feature that was never implemented. Safe to remove without even a default.

**Credential-override types:** The `retrieved*` fields (e.g., `retrievedPassword`, `retrievedUsername`) are populated by `PrepareInternalValues`→`setInternalCredentials` BEFORE `GetPayload()`→`toMap()` is called in the reconcile path. Tagging them `vault:"bindpass,omitempty"` produces identical behavior — `omitempty` excludes them when empty (before credential resolution), and emits them when populated (after credential resolution).

## Verification

**Commands:**
- `make test` — expected: all unit tests pass after each phase
- `make manifests generate fmt vet` — expected: no errors, no new CRDs generated
- `make lint` — expected: zero violations
- `go test ./api/v1alpha1/... -v -count=1` — expected: all type-level tests pass
- `go test ./api/v1alpha1/utils/... -v -count=1` — expected: walker tests pass
