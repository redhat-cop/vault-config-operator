# Story 7.4: Audit Vault API Responses and Harden `IsEquivalentToDesiredState` Extra-Field Handling

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want to audit every Vault API read response for all 46 CRD types and ensure `IsEquivalentToDesiredState` correctly ignores extra fields Vault returns,
So that the operator never enters an unnecessary write loop where it rewrites identical state on every reconcile cycle.

## Acceptance Criteria

1. **Given** a running Vault instance at the project's target version (currently 1.19.0) **When** each of the 46 CRD types is written to Vault and then read back **Then** the exact set of extra fields (keys in read response not in write payload) is documented per type

2. **Given** the documented extra fields per type **When** each type's `IsEquivalentToDesiredState` is called with a payload containing those extra fields **Then** it returns `true` (no false drift detected)

3. **Given** any type where Vault returns values in a different format (e.g., duration as int vs string) **When** `IsEquivalentToDesiredState` is called with the Vault-formatted values **Then** it returns `true` (type coercion is handled)

## Tasks / Subtasks

### Phase 1: Audit (Document extra fields per type)

- [ ] Task 1: Create audit test infrastructure (AC: 1)
  - [ ] 1.1: Create `api/v1alpha1/isequivalent_audit_test.go` with build tag `//go:build !integration` containing a table-driven audit that documents for each type: what `toMap()` produces (desired keys) vs what Vault actually returns (all keys including extras)
  - [ ] 1.2: For types where Vault behavior is well-known (from existing integration tests, Vault docs, or code comments), document extra fields as hardcoded test fixtures
  - [ ] 1.3: For types requiring external services (cloud providers, LDAP, RabbitMQ, etc.), document expected extras from Vault API documentation

- [ ] Task 2: Document extra fields per type category (AC: 1)
  - [ ] 2.1: Document fields for bare-DeepEqual types (31 types) — these are the highest risk
  - [ ] 2.2: Document fields for desired-side-only-delete types (5 types: GitHubSecretEngineConfig, KubernetesSecretEngineConfig, LDAPAuthEngineConfig, QuaySecretEngineConfig, Policy)
  - [ ] 2.3: Verify custom-handling types (9 types) still handle all known extras correctly

### Phase 2: Fix (Harden IsEquivalentToDesiredState)

- [ ] Task 3: Implement shared helper function (AC: 2, 3)
  - [ ] 3.1: Create `filterPayloadToDesiredKeys(desiredState, payload map[string]interface{}) map[string]interface{}` in `api/v1alpha1/utils/` or as a package-level helper in `api/v1alpha1/`
  - [ ] 3.2: The helper filters `payload` to only contain keys present in `desiredState` (top-level), returning a new map safe for `reflect.DeepEqual`
  - [ ] 3.3: Optionally add duration normalization helper: `normalizeDurationValue(val interface{}) interface{}` that converts int seconds back to Go duration string if needed

- [ ] Task 4: Fix the 31 bare-DeepEqual types (AC: 2)
  - [ ] 4.1: Update each type's `IsEquivalentToDesiredState` to use the filter pattern: `filteredPayload := filterPayloadToDesiredKeys(desiredState, payload); return reflect.DeepEqual(desiredState, filteredPayload)`
  - [ ] 4.2: Affected types: AuthEngineMount, AzureAuthEngineConfig, AzureAuthEngineRole, AzureSecretEngineConfig, AzureSecretEngineRole, CertAuthEngineConfig, CertAuthEngineRole, DatabaseSecretEngineRole, DatabaseSecretEngineStaticRole, GCPAuthEngineConfig, GCPAuthEngineRole, GitHubSecretEngineRole, IdentityOIDCAssignment, IdentityOIDCClient, IdentityOIDCProvider, IdentityOIDCScope, IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole, JWTOIDCAuthEngineConfig, JWTOIDCAuthEngineRole, KubernetesAuthEngineConfig, KubernetesAuthEngineRole, KubernetesSecretEngineRole, LDAPAuthEngineGroup, PasswordPolicy, PKISecretEngineConfig, PKISecretEngineRole, QuaySecretEngineRole, QuaySecretEngineStaticRole, RabbitMQSecretEngineConfig, RabbitMQSecretEngineRole

- [ ] Task 5: Fix the 5 desired-side-only-delete types (AC: 2)
  - [ ] 5.1: GitHubSecretEngineConfig — add payload filtering after `delete(desiredState, "prv_key")`
  - [ ] 5.2: KubernetesSecretEngineConfig — add payload filtering after `delete(desiredState, "service_account_jwt")`
  - [ ] 5.3: LDAPAuthEngineConfig — add payload filtering after `delete(desiredState, "bindpass")`
  - [ ] 5.4: QuaySecretEngineConfig — add payload filtering after `delete(desiredState, "password")`
  - [ ] 5.5: Policy — add payload filtering after existing name/rules remapping logic

- [ ] Task 6: Review custom-handling types (AC: 2)
  - [ ] 6.1: DatabaseSecretEngineConfig — already filters payload; verify `connection_details` sub-map filtering is complete
  - [ ] 6.2: SecretEngineMount — currently compares tune config only; verify Vault tune read doesn't add extras
  - [ ] 6.3: Entity/EntityAlias/GroupAlias — `delete()` on payload approach; verify all Vault-added keys are listed
  - [ ] 6.4: Group — only deletes `"name"` from payload; likely needs more deletes or switch to filter pattern
  - [ ] 6.5: Audit/AuditRequestHeader — field-by-field approach inherently ignores extras; verify correctness

### Phase 3: Test (Unit tests per type)

- [ ] Task 7: Add/update unit tests for fixed types (AC: 2, 3)
  - [ ] 7.1: For each fixed type, add a test case with a payload containing documented extra fields — must return `true`
  - [ ] 7.2: Verify existing negative tests still pass (payload with wrong *managed* field values must still return `false`)
  - [ ] 7.3: Add duration-coercion test cases if any types have this issue (e.g., IdentityTokenKey `rotation_period`, `verification_ttl`)

- [ ] Task 8: Verify no regressions (AC: 2)
  - [ ] 8.1: Run `make test` — all unit tests pass
  - [ ] 8.2: Run `make integration` — all 83+ integration specs pass

## Dev Notes

### The Core Problem

`VaultEndpoint.CreateOrUpdate()` (in `api/v1alpha1/utils/vaultobject.go:159-174`) reads from Vault via `read()` which returns `secret.Data` — the raw Vault API response. This map is passed **directly** to `IsEquivalentToDesiredState(payload)`. The operator's desired state (from `toMap()`) typically contains only the fields the operator manages. But Vault's read response includes **additional** fields (timestamps, IDs, computed values, metadata). If `IsEquivalentToDesiredState` uses bare `reflect.DeepEqual(desiredState, payload)`, the extra keys make the maps unequal, causing an unnecessary `write()` every reconcile cycle.

### Current State: 3 Categories of Implementations

**Category A — Bare DeepEqual (31 types, BROKEN for extra fields):**
```go
func (d *KubernetesAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.KAECConfig.toMap()
    return reflect.DeepEqual(desiredState, payload)
}
```
These will return `false` whenever Vault adds *any* extra key to the read response.

**Category B — Desired-side secret deletion (5 types, PARTIALLY BROKEN):**
```go
func (d *GitHubSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.toMap()
    delete(desiredState, "prv_key")
    return reflect.DeepEqual(desiredState, payload)
}
```
These remove credentials from `desiredState` (so a missing credential in payload doesn't cause false drift) but still require `payload` to have **exactly** the remaining keys — any extra Vault field breaks it.

**Category C — Custom handling (9 types, mostly CORRECT):**
- `DatabaseSecretEngineConfig`: Filters payload to only managed keys — **correct pattern to follow**
- `Entity`/`EntityAlias`/`GroupAlias`: Explicit `delete()` on payload for known extras — correct if list is complete
- `Audit`/`AuditRequestHeader`: Field-by-field comparison, inherently ignores extras — correct
- `RandomSecret`: Always returns `false` — correct (intentional always-update)
- `Group`: Only deletes `"name"` from payload — likely incomplete
- `SecretEngineMount`: Compares tune config only — needs verification

### The Recommended Fix Pattern

Follow the `DatabaseSecretEngineConfig` approach — filter `payload` to only keys present in `desiredState`:

```go
func filterPayloadToDesiredKeys(desiredState, payload map[string]interface{}) map[string]interface{} {
    filtered := make(map[string]interface{}, len(desiredState))
    for key := range desiredState {
        if val, exists := payload[key]; exists {
            filtered[key] = val
        }
    }
    return filtered
}
```

Then each type becomes:
```go
func (d *KubernetesAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.KAECConfig.toMap()
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

### Helper Function Location

Place `filterPayloadToDesiredKeys` as an **exported** function in `api/v1alpha1/utils/vaultobject.go` (or a new file `api/v1alpha1/utils/payload_filter.go`) since `VaultObject` interface lives there. Alternatively, since types are in `api/v1alpha1/` and not `api/v1alpha1/utils/`, a package-level unexported helper in `api/v1alpha1/` avoids import cycles. Choose the location that avoids circular imports.

**Import structure check:** Types in `api/v1alpha1/` already import `api/v1alpha1/utils` (for `vaultutils.KubeAuthConfiguration`, etc.). So placing the helper in `api/v1alpha1/utils` and having types call `vaultutils.FilterPayloadToDesiredKeys(...)` works. OR define it unexported in `api/v1alpha1/` directly.

**Recommendation:** Define it as an unexported function `filterPayloadToDesiredKeys` in a new file `api/v1alpha1/payload_filter.go` in the same package as all the `*_types.go` files. This avoids any import changes and keeps it co-located with its consumers.

### Duration/Type Coercion Concern

Some Vault endpoints return duration values as integers (seconds) instead of the string format sent (e.g., `"24h"` written, `86400` or `"86400"` read back). Known risk areas:
- `IdentityTokenKey`: `rotation_period`, `verification_ttl`
- `IdentityTokenRole`: `token_ttl`
- `DatabaseSecretEngineRole`: `default_ttl`, `max_ttl`
- `DatabaseSecretEngineStaticRole`: `rotation_period`
- Auth engine roles with `token_ttl`, `token_max_ttl`, etc.

**If filtering alone solves the problem** (Vault returns the same value format it was given), no coercion is needed. The audit in Phase 1 will determine if coercion is required. If it is, add a `normalizeDuration` helper that:
1. If `payload[key]` is `json.Number` or `float64`, convert to duration string
2. If `payload[key]` is already a string, leave it

### Critical Architecture Constraints

1. **Never call `toMap()` in tests to build expected values** — construct independent Vault-read-shaped fixtures (project-context.md rule)
2. **No `make manifests generate` needed** — this story modifies only `*_types.go` method bodies and adds test files; no CRD schema changes
3. **Existing integration tests are the final regression gate** — all 83+ specs must still pass after changes
4. **Mutation of payload parameter**: The Entity/EntityAlias pattern `delete(payload, ...)` mutates the caller's map. The filter pattern creates a new map, which is safer. When migrating types that currently delete from payload, prefer the filter pattern to avoid side effects.

### AuthEngineMount/SecretEngineMount Special Case

These are `VaultEngineObject` types, not standard `VaultObject`. Their `IsEquivalentToDesiredState` compares **tune config** (from `d.Spec.Config.toMap()`) against the Vault tune read response. The tune endpoint (`/sys/mounts/<path>/tune` or `/sys/auth/<path>/tune`) may return extra fields like `force_no_cache`, `token_type`, etc. These types need the same filter treatment but on the Config.toMap() level.

From `authenginemount_types.go`:
```go
func (d *AuthEngineMount) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.Config.toMap()
    return reflect.DeepEqual(desiredState, payload)
}
```

### Types NOT to Modify

- **RandomSecret** — returns `false` always (intentional behavior)
- **Audit** — field-by-field comparison already ignores extras correctly
- **AuditRequestHeader** — checks only `hmac` field, ignores extras correctly

### Test Pattern for Extra-Field Tolerance

For each fixed type, add a test case like:

```go
func TestKubernetesAuthEngineConfigIsEquivalentIgnoresExtraFields(t *testing.T) {
    config := &KubernetesAuthEngineConfig{
        Spec: KubernetesAuthEngineConfigSpec{
            Path: "kubernetes",
            KAECConfig: KAECConfig{
                KubernetesHost:   "https://kubernetes.default.svc:443",
                KubernetesCACert: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
            },
        },
    }

    // Vault-read fixture with extra fields Vault typically adds
    payload := map[string]interface{}{
        "kubernetes_host":                   "https://kubernetes.default.svc:443",
        "kubernetes_ca_cert":                "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
        "token_reviewer_jwt":                "",
        "pem_keys":                          []string(nil),
        "issuer":                            "",
        "disable_iss_validation":            false,
        "disable_local_ca_jwt":              false,
        "use_annotations_as_alias_metadata": false,
        // Extra fields Vault adds but operator doesn't manage:
        "accessor":           "auth_kubernetes_abc123",
        "local":              false,
        "seal_wrap":          false,
    }

    if !config.IsEquivalentToDesiredState(payload) {
        t.Error("expected true: extra Vault fields should be ignored")
    }
}
```

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/payload_filter.go` | New | Shared `filterPayloadToDesiredKeys` helper |
| 2 | `api/v1alpha1/payload_filter_test.go` | New | Unit tests for the filter helper |
| 3 | `api/v1alpha1/authenginemount_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 4 | `api/v1alpha1/azureauthengineconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 5 | `api/v1alpha1/azureauthenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 6 | `api/v1alpha1/azuresecretengineconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 7 | `api/v1alpha1/azuresecretenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 8 | `api/v1alpha1/certauthengineconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 9 | `api/v1alpha1/certauthenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 10 | `api/v1alpha1/databasesecretenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 11 | `api/v1alpha1/databasesecretenginestaticrole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 12 | `api/v1alpha1/gcpauthengineconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 13 | `api/v1alpha1/gcpauthenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 14 | `api/v1alpha1/githubsecretengineconfig_types.go` | Modify | Fix + add payload filtering |
| 15 | `api/v1alpha1/githubsecretenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 16 | `api/v1alpha1/group_types.go` | Modify | Expand `delete()` list or switch to filter |
| 17 | `api/v1alpha1/identityoidcassignment_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 18 | `api/v1alpha1/identityoidcclient_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 19 | `api/v1alpha1/identityoidcprovider_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 20 | `api/v1alpha1/identityoidcscope_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 21 | `api/v1alpha1/identitytokenconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 22 | `api/v1alpha1/identitytokenkey_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 23 | `api/v1alpha1/identitytokenrole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 24 | `api/v1alpha1/jwtoidcauthengineconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 25 | `api/v1alpha1/jwtoidcauthenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 26 | `api/v1alpha1/kubernetesauthengineconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 27 | `api/v1alpha1/kubernetesauthenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 28 | `api/v1alpha1/kubernetessecretengineconfig_types.go` | Modify | Fix + add payload filtering |
| 29 | `api/v1alpha1/kubernetessecretenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 30 | `api/v1alpha1/ldapauthengineconfig_types.go` | Modify | Fix + add payload filtering |
| 31 | `api/v1alpha1/ldapauthenginegroup_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 32 | `api/v1alpha1/passwordpolicy_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 33 | `api/v1alpha1/pkisecretengineconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 34 | `api/v1alpha1/pkisecretenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 35 | `api/v1alpha1/policy_types.go` | Modify | Fix + add payload filtering |
| 36 | `api/v1alpha1/quaysecretengineconfig_types.go` | Modify | Fix + add payload filtering |
| 37 | `api/v1alpha1/quaysecretenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 38 | `api/v1alpha1/quaysecretenginestaticrole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 39 | `api/v1alpha1/rabbitmqsecretengineconfig_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 40 | `api/v1alpha1/rabbitmqsecretenginerole_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 41 | `api/v1alpha1/secretenginemount_types.go` | Modify | Fix `IsEquivalentToDesiredState` |
| 42 | `api/v1alpha1/*_test.go` (multiple) | Modify/New | Add extra-field tolerance tests per type |

### Previous Story Intelligence

**From Story 7.3 (Error path integration tests):**
- Integration test infrastructure is stable with 83+ specs passing
- `vault-admin` namespace has working auth roles: `policy-admin`, `database-engine-admin`, `kv-engine-admin`, `secret-writer`
- Standard pattern: create CR → poll for condition → verify → delete

**From Epic 6 Retrospective:**
- "Continue detailed dev notes in story specs" — applied
- Codebase stable on main at commit `9fc8b3c`
- Coverage at 53.7%

**From existing `IsEquivalentToDesiredState` tests (Story 1.x, Epic 6):**
- Tests use hardcoded Vault-read-shaped fixtures (not `toMap()` output)
- `[]interface{}` for list fields (Vault JSON deserialization)
- Table-driven tests for multi-scenario coverage
- Existing tests in `identityoidc_test.go`, `identitytoken_test.go`, `authenginemount_test.go` already test "payloadWithExtra" scenarios — these currently expect `false` and will need updating to expect `true` after the fix

### Git Intelligence (Recent Commits)

```
9fc8b3c Bmad epic 6 (#321)
7ce3e42 Merge pull request #320 from raffaelespazzoli/bmad-epic-5
d64b2b1 Complete Epic 5 retrospective and close epic
```

No recent changes to `IsEquivalentToDesiredState` implementations or the `CreateOrUpdate` flow.

### Existing Test Updates Required

Several existing tests explicitly verify that payloads with extra fields return `false`. After fixing the types, these tests must be updated to expect `true`:

- `api/v1alpha1/identityoidc_test.go` — "payloadWithExtra" test cases for IdentityOIDCAssignment, Client, Provider, Scope
- `api/v1alpha1/identitytoken_test.go` — "payloadWithExtra" test cases for IdentityTokenConfig, Key, Role
- `api/v1alpha1/authenginemount_test.go` — extra-field tests
- `api/v1alpha1/secretenginemount_test.go` — extra-field tests

These tests were written to *document the current (broken) behavior*. After the fix, they should be updated to assert the correct behavior (extra fields ignored → returns `true`).

### Scope Boundaries

**In scope:**
- Modifying `IsEquivalentToDesiredState` method bodies in `*_types.go`
- Adding shared helper function
- Adding/updating unit tests
- No CRD schema changes, no controller changes, no webhook changes

**Out of scope:**
- Changing `CreateOrUpdate` flow in `vaultobject.go`
- Modifying `read()` or `write()` in `vaultutils.go`
- Adding integration tests for drift detection (that's Story 7.5)
- Nested map filtering (only top-level key filtering; sub-map comparison is left to type-specific logic where already handled, e.g., DatabaseSecretEngineConfig)

### References

- [Source: api/v1alpha1/utils/vaultobject.go#L28-49] — `VaultObject` interface definition
- [Source: api/v1alpha1/utils/vaultobject.go#L159-174] — `CreateOrUpdate` flow calling `IsEquivalentToDesiredState`
- [Source: api/v1alpha1/utils/vaultutils.go#L45-61] — `read()` returning `secret.Data`
- [Source: api/v1alpha1/entity_types.go#L160-174] — Entity's `delete(payload, ...)` pattern (reference for custom handling)
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L93-119] — DatabaseSecretEngineConfig's filtered payload pattern (THE pattern to follow)
- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L76-78] — Bare DeepEqual pattern (broken)
- [Source: _bmad-output/project-context.md#Imperative-to-Declarative Bridge] — Architecture rules for IsEquivalentToDesiredState
- [Source: _bmad-output/project-context.md#Vault API Gotchas] — "Filter the read payload to only the keys the operator manages"
- [Source: _bmad-output/project-context.md#Unit Test Payload Construction] — Never derive expected payloads from code under test
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.4] — Epic requirements and acceptance criteria

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
