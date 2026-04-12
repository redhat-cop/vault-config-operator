# Story 1.6: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Remaining Types

Status: ready-for-dev

## Story

As an operator developer,
I want unit tests for Policy, RandomSecret, Audit, Group, GroupAlias, Entity, and EntityAlias,
So that the full type portfolio has declarative logic coverage.

**Note:** The epic user story text mentions VaultSecret, but VaultSecret does NOT implement `VaultObject` — it has no `toMap()`, `GetPayload()`, `GetPath()`, or `IsEquivalentToDesiredState()`. VaultSecret uses a completely different reconciliation model (reads from Vault, writes to Kubernetes Secret). It is correctly excluded from testing scope.

## Acceptance Criteria

1. **Given** a Policy instance with `Spec.Type == ""` and a Vault read response where the policy text is under `rules` (not `policy`) and a `name` key is present **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true` (the method renames `policy` → `rules` and adds `name` to match Vault's response)

2. **Given** a Policy instance with `Spec.Type == "acl"` and a Vault read response where the policy text is under `policy` and a `name` key is present **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true` (the method keeps `policy` key and adds `name`)

3. **Given** a RandomSecret instance **When** `IsEquivalentToDesiredState(payload)` is called with any payload **Then** it returns `false` (hardcoded — RandomSecret always writes unconditionally)

4. **Given** an Audit instance and a matching Vault read response with `type`, `description`, `local`, and `options` (as `map[string]string`) **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true`

5. **Given** an Audit instance and a Vault read response where `options` is `map[string]interface{}` instead of `map[string]string` **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `false` (the type assertion to `map[string]string` fails)

6. **Given** a Group instance with `Type == "internal"` **When** `toMap()` is called **Then** the returned map contains 5 keys: `type`, `metadata`, `policies`, `member_group_ids`, `member_entity_ids`

7. **Given** a Group instance with `Type == "external"` **When** `toMap()` is called **Then** the returned map contains only 3 keys: `type`, `metadata`, `policies` (member keys are excluded)

8. **Given** a Group instance and a Vault payload with `name` plus the managed keys **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true` (`name` is deleted from the payload before comparison)

9. **Given** a GroupAlias instance with `retrieved*` unexported fields set directly **When** `toMap()` and `IsEquivalentToDesiredState` are called **Then** correct behavior is verified — extra Vault keys (`creation_time`, `last_update_time`, `merged_from_canonical_ids`, `metadata`, `mount_path`, `mount_type`) are deleted from payload before comparison

10. **Given** an Entity instance with metadata, policies, and disabled fields **When** `toMap()` and `IsEquivalentToDesiredState` are called **Then** correct behavior is verified — 10 Vault-only keys are deleted from payload before comparison

11. **Given** an EntityAlias instance with `retrieved*` unexported fields set and `CustomMetadata` populated **When** `toMap()` is called **Then** the returned map includes `custom_metadata`

12. **Given** an EntityAlias instance with empty `CustomMetadata` **When** `toMap()` is called **Then** the returned map does NOT include `custom_metadata`

13. **Given** each type in this story **When** `GetPath()` is called with and without `Spec.Name` set **Then** the correct Vault path is returned

14. **Given** each type in this story **When** `IsDeletable()` is called **Then** all return `true`

## Types Covered

| # | Type | File | Map Method | `IsEquivalentToDesiredState` | `IsDeletable` | Keys | Existing Tests |
|---|------|------|------------|------------------------------|---------------|------|----------------|
| 1 | Policy | `api/v1alpha1/policy_types.go` | `GetPayload()` (1 key) | **Custom: adds `name`, conditionally renames `policy` → `rules`** | `true` | 1 (+name in equiv) | None |
| 2 | RandomSecret | `api/v1alpha1/randomsecret_types.go` | `getV1Payload()` / `GetPayload()` (1-2 keys, dynamic key name) | **Always returns `false`** | `true` | 1-2 | None |
| 3 | Audit | `api/v1alpha1/audit_types.go` | `GetPayload()` (4 keys) | **Custom: field-by-field comparison with type assertion on `options`** | `true` | 4 | None |
| 4 | Group | `api/v1alpha1/group_types.go` | `GroupSpec.toMap()` (3-5 keys, conditional) | **Custom: deletes `name` from payload** | `true` | 3-5 | None |
| 5 | GroupAlias | `api/v1alpha1/groupalias_types.go` | `GroupAliasSpec.toMap()` (4 keys, uses retrieved* fields) | **Custom: deletes 6 keys from payload; has `fmt.Print` debug** | `true` | 4 | None |
| 6 | Entity | `api/v1alpha1/entity_types.go` | `EntitySpec.toMap()` (3 keys) | **Custom: deletes 10 keys from payload** | `true` | 3 | None |
| 7 | EntityAlias | `api/v1alpha1/entityalias_types.go` | `EntityAliasSpec.toMap()` (4-5 keys, uses retrieved* fields, conditional `custom_metadata`) | **Custom: deletes 8 keys from payload** | `true` | 4-5 | None |

## Tasks / Subtasks

- [ ] Task 1: Add Policy unit tests (AC: 1, 2, 13, 14)
  - [ ] 1.1: Create `api/v1alpha1/policy_test.go`
  - [ ] 1.2: Add `TestPolicyGetPath` — 4 variants: with/without `Spec.Name`, with/without `Spec.Type`; verify `CleansePath("sys/policies/<type>/<name>")` when Type set, `CleansePath("sys/policy/<name>")` when Type empty
  - [ ] 1.3: Add `TestPolicyGetPayload` — verify single key `"policy"` mapped from `Spec.Policy`
  - [ ] 1.4: Add `TestPolicyIsEquivalentNoType` — **critical test**: when `Spec.Type == ""`, verify `policy` key is renamed to `rules` and `name` is added; a payload `{name: X, rules: Y}` → `true`
  - [ ] 1.5: Add `TestPolicyIsEquivalentWithType` — when `Spec.Type == "acl"`, verify `policy` key stays, `name` is added; a payload `{name: X, policy: Y}` → `true`
  - [ ] 1.6: Add `TestPolicyIsEquivalentNameFromSpec` — verify `name` uses `Spec.Name` when set (not metadata name)
  - [ ] 1.7: Add `TestPolicyIsEquivalentNameFromMetadata` — verify `name` falls back to `d.Name` (metadata) when `Spec.Name` empty
  - [ ] 1.8: Add `TestPolicyIsEquivalentNonMatching` — change the policy text → `false`
  - [ ] 1.9: Add `TestPolicyIsEquivalentExtraFields` — extra keys in payload → `false` (bare DeepEqual after mutations)
  - [ ] 1.10: Add `TestPolicyIsDeletable` — returns `true`
  - [ ] 1.11: Add `TestPolicyConditions` — GetConditions/SetConditions round-trip
- [ ] Task 2: Add RandomSecret unit tests (AC: 3, 13, 14)
  - [ ] 2.1: Create `api/v1alpha1/randomsecret_test.go`
  - [ ] 2.2: Add `TestRandomSecretGetPath` — with/without `Spec.Name`; verify `CleansePath(path + "/" + name)`
  - [ ] 2.3: Add `TestRandomSecretGetV1Payload` — set `calculatedSecret` directly (unexported field, same package); verify dynamic key `Spec.SecretKey` → `calculatedSecret` value
  - [ ] 2.4: Add `TestRandomSecretGetV1PayloadWithRefreshPeriod` — set `RefreshPeriod` with non-zero Duration; verify `ttl` key is present with duration string
  - [ ] 2.5: Add `TestRandomSecretGetV1PayloadNoRefreshPeriod` — nil `RefreshPeriod`; verify no `ttl` key
  - [ ] 2.6: Add `TestRandomSecretGetPayloadKVv2` — verify KV v2 wraps inner payload under `"data"` key; uses `IsKVSecretsEngineV2()` (path contains `/data/`)
  - [ ] 2.7: Add `TestRandomSecretGetPayloadKVv1` — verify KV v1 returns inner payload directly
  - [ ] 2.8: Add `TestRandomSecretIsEquivalentAlwaysFalse` — verify always returns `false` regardless of payload content
  - [ ] 2.9: Add `TestRandomSecretIsDeletable` — returns `true`
  - [ ] 2.10: Add `TestRandomSecretConditions` — GetConditions/SetConditions round-trip
- [ ] Task 3: Add Audit unit tests (AC: 4, 5, 13, 14)
  - [ ] 3.1: Create `api/v1alpha1/audit_test.go`
  - [ ] 3.2: Add `TestAuditGetPath` — verify returns `CleansePath("sys/audit/" + Spec.Path)`
  - [ ] 3.3: Add `TestAuditGetPayload` — verify 4 keys: `type`, `description`, `local`, `options`; verify `options` is `map[string]string` type
  - [ ] 3.4: Add `TestAuditIsEquivalentMatching` — matching payload with all 4 fields including `options` as `map[string]string` → `true`
  - [ ] 3.5: Add `TestAuditIsEquivalentTypeMismatch` — different `type` → `false`
  - [ ] 3.6: Add `TestAuditIsEquivalentDescriptionMismatch` — different `description` → `false`
  - [ ] 3.7: Add `TestAuditIsEquivalentLocalMismatch` — different `local` → `false`
  - [ ] 3.8: Add `TestAuditIsEquivalentOptionsMismatch` — different option value → `false`
  - [ ] 3.9: Add `TestAuditIsEquivalentOptionsLengthMismatch` — extra option key → `false`
  - [ ] 3.10: Add `TestAuditIsEquivalentOptionsWrongType` — **critical**: `options` as `map[string]interface{}` (not `map[string]string`) → `false` (type assertion fails)
  - [ ] 3.11: Add `TestAuditIsEquivalentExtraFields` — extra top-level keys in payload → `true` (Audit only checks its 4 fields, ignores extras)
  - [ ] 3.12: Add `TestAuditIsDeletable` — returns `true`
  - [ ] 3.13: Add `TestAuditConditions` — GetConditions/SetConditions round-trip
- [ ] Task 4: Add Group unit tests (AC: 6, 7, 8, 13, 14)
  - [ ] 4.1: Create `api/v1alpha1/group_test.go`
  - [ ] 4.2: Add `TestGroupGetPath` — with/without `Spec.Name`; verify path format `CleansePath("/identity/group/name/" + name)`
  - [ ] 4.3: Add `TestGroupToMapInternal` — `Type == "internal"`; verify 5 keys: `type`, `metadata`, `policies`, `member_group_ids`, `member_entity_ids`
  - [ ] 4.4: Add `TestGroupToMapExternal` — `Type == "external"`; verify only 3 keys: `type`, `metadata`, `policies`; verify `member_group_ids` and `member_entity_ids` are absent
  - [ ] 4.5: Add `TestGroupIsEquivalentMatching` — matching payload with extra `name` key → `true` (`name` is deleted from payload)
  - [ ] 4.6: Add `TestGroupIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 4.7: Add `TestGroupIsEquivalentExtraFields` — extra keys beyond `name` in payload → `false` (only `name` is deleted; DeepEqual catches others)
  - [ ] 4.8: Add `TestGroupGetPayload` — verify delegates to `Spec.toMap()`
  - [ ] 4.9: Add `TestGroupIsDeletable` — returns `true`
  - [ ] 4.10: Add `TestGroupConditions` — GetConditions/SetConditions round-trip
- [ ] Task 5: Add GroupAlias unit tests (AC: 9, 13, 14)
  - [ ] 5.1: Create `api/v1alpha1/groupalias_test.go`
  - [ ] 5.2: Add `TestGroupAliasGetPath` — verify returns `CleansePath("/identity/group-alias/id/" + Status.ID)` — **note: path depends on Status.ID, not Spec fields**
  - [ ] 5.3: Add `TestGroupAliasToMap` — set all 4 `retrieved*` unexported fields directly (same package); verify 4 keys: `name`, `id`, `mount_accessor`, `canonical_id`
  - [ ] 5.4: Add `TestGroupAliasIsEquivalentMatching` — matching payload with extra Vault keys (`creation_time`, `last_update_time`, `merged_from_canonical_ids`, `metadata`, `mount_path`, `mount_type`) → `true` (all 6 are deleted before comparison)
  - [ ] 5.5: Add `TestGroupAliasIsEquivalentNonMatching` — one managed field changed → `false`
  - [ ] 5.6: Add `TestGroupAliasIsEquivalentExtraFields` — extra keys beyond the 6 known Vault keys → `false` (only the 6 specific keys are deleted)
  - [ ] 5.7: Add `TestGroupAliasGetPayload` — verify delegates to `Spec.toMap()`
  - [ ] 5.8: Add `TestGroupAliasIsDeletable` — returns `true`
  - [ ] 5.9: Add `TestGroupAliasConditions` — GetConditions/SetConditions round-trip
- [ ] Task 6: Add Entity unit tests (AC: 10, 13, 14)
  - [ ] 6.1: Create `api/v1alpha1/entity_test.go`
  - [ ] 6.2: Add `TestEntityGetPath` — with/without `Spec.Name`; verify path format `CleansePath("/identity/entity/name/" + name)`
  - [ ] 6.3: Add `TestEntityToMap` — verify 3 keys: `metadata`, `policies`, `disabled`
  - [ ] 6.4: Add `TestEntityIsEquivalentMatching` — matching payload with all 10 extra Vault keys (`name`, `id`, `aliases`, `creation_time`, `last_update_time`, `merged_entity_ids`, `direct_group_ids`, `group_ids`, `inherited_group_ids`, `namespace_id`, `bucket_key_hash`) → `true` (all deleted before comparison)
  - [ ] 6.5: Add `TestEntityIsEquivalentNonMatching` — one managed field changed → `false`
  - [ ] 6.6: Add `TestEntityIsEquivalentExtraFields` — extra keys beyond the 10 known Vault keys → `false`
  - [ ] 6.7: Add `TestEntityGetPayload` — verify delegates to `Spec.toMap()`
  - [ ] 6.8: Add `TestEntityIsDeletable` — returns `true`
  - [ ] 6.9: Add `TestEntityConditions` — GetConditions/SetConditions round-trip
- [ ] Task 7: Add EntityAlias unit tests (AC: 11, 12, 13, 14)
  - [ ] 7.1: Create `api/v1alpha1/entityalias_test.go`
  - [ ] 7.2: Add `TestEntityAliasGetPath` — verify returns `CleansePath("/identity/entity-alias/id/" + Status.ID)` — **note: path depends on Status.ID**
  - [ ] 7.3: Add `TestEntityAliasToMapWithCustomMetadata` — set all 4 `retrieved*` fields + non-empty `CustomMetadata`; verify 5 keys: `name`, `id`, `mount_accessor`, `canonical_id`, `custom_metadata`
  - [ ] 7.4: Add `TestEntityAliasToMapWithoutCustomMetadata` — set `retrieved*` fields, leave `CustomMetadata` empty; verify only 4 keys, `custom_metadata` absent
  - [ ] 7.5: Add `TestEntityAliasIsEquivalentMatching` — matching payload with extra Vault keys (`creation_time`, `last_update_time`, `merged_from_canonical_ids`, `metadata`, `mount_path`, `mount_type`, `local`, `namespace_id`) → `true` (all 8 deleted before comparison)
  - [ ] 7.6: Add `TestEntityAliasIsEquivalentNonMatching` — one managed field changed → `false`
  - [ ] 7.7: Add `TestEntityAliasIsEquivalentExtraFields` — extra keys beyond the 8 known → `false`
  - [ ] 7.8: Add `TestEntityAliasGetPayload` — verify delegates to `Spec.toMap()`
  - [ ] 7.9: Add `TestEntityAliasIsDeletable` — returns `true`
  - [ ] 7.10: Add `TestEntityAliasConditions` — GetConditions/SetConditions round-trip
- [ ] Task 8: Verify all tests pass (AC: all)
  - [ ] 8.1: Run `go test ./api/v1alpha1/ -v -count=1` to confirm all new and existing tests pass
  - [ ] 8.2: Run `make test` to verify no regressions in full unit test suite

## Dev Notes

### Critical: Policy Has NO `toMap()` — Uses `GetPayload()` Directly

Unlike most types, Policy builds its map in `GetPayload()` returning a single key: `"policy"` → `d.Spec.Policy`. The `IsEquivalentToDesiredState` method then MUTATES this map by adding a `name` key and conditionally renaming `policy` → `rules`.

```go
func (d *Policy) GetPayload() map[string]interface{} {
    return map[string]interface{}{
        "policy": d.Spec.Policy,
    }
}
```

[Source: api/v1alpha1/policy_types.go#L55-L59]

### Critical: Policy `IsEquivalentToDesiredState` Has Conditional Field Remapping

The most complex `IsEquivalentToDesiredState` in this story. It mutates the desired state map in two ways:

1. **Adds `name`**: Uses `Spec.Name` if non-empty, else metadata `d.Name`
2. **Conditional rename**: If `Spec.Type == ""` (no type specified), renames `policy` to `rules` and deletes `policy`. This matches the legacy `sys/policy` endpoint behavior where Vault returns `rules` instead of `policy`.

```go
func (d *Policy) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.GetPayload()
    desiredState["name"] = map[bool]string{true: d.Spec.Name, false: d.Name}[d.Spec.Name != ""]
    if d.Spec.Type == "" {
        desiredState["rules"] = desiredState["policy"]
        delete(desiredState, "policy")
    }
    return reflect.DeepEqual(desiredState, payload)
}
```

Test matrix:
- `Type == ""`, `Spec.Name == ""` → desired: `{name: metadataName, rules: policyText}`, matching payload → `true`
- `Type == ""`, `Spec.Name == "foo"` → desired: `{name: "foo", rules: policyText}`, matching payload → `true`
- `Type == "acl"`, `Spec.Name == ""` → desired: `{name: metadataName, policy: policyText}`, matching payload → `true`
- `Type == "acl"`, `Spec.Name == "foo"` → desired: `{name: "foo", policy: policyText}`, matching payload → `true`

[Source: api/v1alpha1/policy_types.go#L60-L68]

### Critical: Policy `GetPath()` Has 4 Variants

`GetPath()` depends on two booleans (`Spec.Name != ""` and `Spec.Type != ""`):

| Spec.Name set | Spec.Type set | Path |
|---------------|---------------|------|
| Yes | Yes | `CleansePath("sys/policies/" + Type + "/" + Spec.Name)` |
| Yes | No | `CleansePath("sys/policy/" + Spec.Name)` |
| No | Yes | `CleansePath("sys/policies/" + Type + "/" + d.Name)` |
| No | No | `CleansePath("sys/policy/" + d.Name)` |

[Source: api/v1alpha1/policy_types.go#L43-L54]

### Critical: RandomSecret Has NO `toMap()` and `IsEquivalentToDesiredState` Always Returns `false`

RandomSecret is fundamentally different from every other type — its `IsEquivalentToDesiredState` is hardcoded to `false`, meaning every reconcile writes unconditionally. This is intentional because the generated password must be written regardless of Vault state.

```go
func (d *RandomSecret) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    return false
}
```

[Source: api/v1alpha1/randomsecret_types.go#L140-L142]

### Critical: RandomSecret `GetPayload()` Uses Dynamic Key Name and Unexported Field

The V1 payload uses `Spec.SecretKey` as the map key (not a fixed string) and `Spec.calculatedSecret` (unexported) as the value. Tests can set `calculatedSecret` directly since they're in the same package.

```go
func (d *RandomSecret) getV1Payload() map[string]interface{} {
    payload := map[string]interface{}{
        d.Spec.SecretKey: d.Spec.calculatedSecret,
    }
    if d.Spec.RefreshPeriod != nil && d.Spec.RefreshPeriod.Duration > 0 {
        payload[ttlKey] = d.Spec.RefreshPeriod.Duration.String()
    }
    return payload
}
```

For KV v2, `GetPayload()` wraps this under a `"data"` key. KV v2 is detected by `IsKVSecretsEngineV2()` which checks if the path contains `/data/`.

[Source: api/v1alpha1/randomsecret_types.go#L114-L124, L131-L138]

### Critical: Audit Has Unique Field-by-Field `IsEquivalentToDesiredState` (Not DeepEqual)

Audit is the ONLY type in this story (and one of very few in the entire operator) that does NOT use `reflect.DeepEqual`. Instead, it checks each field individually:

1. Compares `type`, `description`, `local` via `!=` operator
2. Type-asserts `payload["options"]` to `map[string]string` — fails if Vault returns `map[string]interface{}`
3. Compares options by length + per-key value check

**This means Audit IGNORES extra top-level keys** — any keys beyond `type`, `description`, `local`, `options` are simply not checked. This is the OPPOSITE behavior of most types where extra keys cause `false`.

```go
func (d *Audit) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredPayload := d.GetPayload()
    if payload["type"] != desiredPayload["type"] { return false }
    if payload["description"] != desiredPayload["description"] { return false }
    if payload["local"] != desiredPayload["local"] { return false }
    if options, ok := payload["options"].(map[string]string); !ok {
        return false
    } else {
        desiredOptions := desiredPayload["options"].(map[string]string)
        if len(options) != len(desiredOptions) { return false }
        for k, v := range options {
            if desiredOptions[k] != v { return false }
        }
    }
    return true
}
```

[Source: api/v1alpha1/audit_types.go#L130-L155]

### Audit Uses `VaultAuditResource` Reconciler (Not Standard `VaultResource`)

Audit implements `vaultutils.VaultAuditObject` in addition to `VaultObject`. The controller creates `NewVaultAuditResource` instead of `NewVaultResource`. This is an architectural note — it doesn't affect unit test writing but explains the type.

[Source: api/v1alpha1/audit_types.go#L104-L106]

### Critical: Group `toMap()` Has Conditional Keys Based on Type

`member_group_ids` and `member_entity_ids` are ONLY included when `Type == "internal"`. External groups get only 3 keys. Test both variants.

```go
func (i *GroupSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["type"] = i.Type
    payload["metadata"] = i.Metadata
    payload["policies"] = i.Policies
    if i.Type == "internal" {
        payload["member_group_ids"] = i.MemberGroupIDs
        payload["member_entity_ids"] = i.MemberEntityIDs
    }
    return payload
}
```

[Source: api/v1alpha1/group_types.go#L145-L155]

### Critical: Group `IsEquivalentToDesiredState` Mutates the Incoming Payload

Group deletes `"name"` from the **incoming `payload`** (not from desired state). This is the Entity/EntityAlias pattern of stripping Vault-added read-only keys.

```go
func (d *Group) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.toMap()
    delete(payload, "name")
    return reflect.DeepEqual(desiredState, payload)
}
```

[Source: api/v1alpha1/group_types.go#L177-L181]

### Critical: GroupAlias and EntityAlias Use Unexported `retrieved*` Fields in `toMap()`

Both GroupAlias and EntityAlias have `toMap()` methods that read from **unexported** fields (`retrievedName`, `retrievedAliasID`, `retrievedMountAccessor`, `retrievedCanonicalID`) that are normally populated by `PrepareInternalValues`. In unit tests, set these directly since tests are in `package v1alpha1`.

**GroupAlias `toMap()`** — 4 keys:

```go
func (i *GroupAliasSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["name"] = i.retrievedName
    payload["id"] = i.retrievedAliasID
    payload["mount_accessor"] = i.retrievedMountAccessor
    payload["canonical_id"] = i.retrievedCanonicalID
    return payload
}
```

[Source: api/v1alpha1/groupalias_types.go#L130-L137]

**EntityAlias `toMap()`** — 4-5 keys (conditional `custom_metadata`):

```go
func (i *EntityAliasSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["name"] = i.retrievedName
    payload["id"] = i.retrievedAliasID
    payload["mount_accessor"] = i.retrievedMountAccessor
    payload["canonical_id"] = i.retrievedCanonicalID
    if len(i.CustomMetadata) > 0 {
        payload["custom_metadata"] = i.CustomMetadata
    }
    return payload
}
```

[Source: api/v1alpha1/entityalias_types.go#L139-L148]

### Critical: GroupAlias `IsEquivalentToDesiredState` Contains `fmt.Print` Debug Statements

GroupAlias has `fmt.Print("desired state", desiredState)` and `fmt.Print("actual state", payload)` calls inside `IsEquivalentToDesiredState`. These are likely unintentional debug prints left in production code. Document this behavior in tests — do NOT remove the prints (this is production code, not test scope).

```go
func (d *GroupAlias) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.toMap()
    delete(payload, "creation_time")
    delete(payload, "last_update_time")
    delete(payload, "merged_from_canonical_ids")
    delete(payload, "metadata")
    delete(payload, "mount_path")
    delete(payload, "mount_type")
    fmt.Print("desired state", desiredState)
    fmt.Print("actual state", payload)
    return reflect.DeepEqual(desiredState, payload)
}
```

[Source: api/v1alpha1/groupalias_types.go#L218-L229]

### Critical: Entity Deletes 10 Keys from Payload

Entity has the most extensive payload cleanup — 10 Vault-only keys deleted: `name`, `id`, `aliases`, `creation_time`, `last_update_time`, `merged_entity_ids`, `direct_group_ids`, `group_ids`, `inherited_group_ids`, `namespace_id`, `bucket_key_hash`.

Note: That's actually **11** keys (count: name, id, aliases, creation_time, last_update_time, merged_entity_ids, direct_group_ids, group_ids, inherited_group_ids, namespace_id, bucket_key_hash). Test with all 11 present and verify they're ignored.

[Source: api/v1alpha1/entity_types.go#L160-L174]

### EntityAlias Deletes 8 Keys from Payload

`creation_time`, `last_update_time`, `merged_from_canonical_ids`, `metadata`, `mount_path`, `mount_type`, `local`, `namespace_id` — 8 keys total.

[Source: api/v1alpha1/entityalias_types.go#L229-L240]

### GroupAlias and EntityAlias `GetPath()` Depend on `Status.ID` (Not Spec Fields)

Both types use `Status.ID` (set during `PrepareInternalValues` after the first creation) in their path:
- GroupAlias: `CleansePath("/identity/group-alias/id/" + Status.ID)`
- EntityAlias: `CleansePath("/identity/entity-alias/id/" + Status.ID)`

In tests, set `Status.ID` directly on the CRD instance to test `GetPath()`.

[Source: api/v1alpha1/groupalias_types.go#L122-L124, api/v1alpha1/entityalias_types.go#L131-L133]

### Audit `GetPath()` Uses `Spec.Path` (Different from Most Types)

Audit uses `Spec.Path` (a plain `string`, not `vaultutils.Path`), prefixed with `sys/audit/`:

```go
func (d *Audit) GetPath() string {
    return vaultutils.CleansePath("sys/audit/" + d.Spec.Path)
}
```

There is no `Spec.Name` override — the path is always `sys/audit/{Spec.Path}`.

[Source: api/v1alpha1/audit_types.go#L116-L118]

### Slice Fields Stored as Go Types, Not `[]interface{}`

Like previous stories, all types store slice fields as their Go types (`[]string`, `map[string]string`, etc.). If Vault returns `[]interface{}` or `map[string]interface{}`, `reflect.DeepEqual` returns `false`. Document this behavior in tests.

Specifically:
- `Group.Policies` → `[]string`, `MemberGroupIDs` → `[]string`, `MemberEntityIDs` → `[]string`
- `Entity.Policies` → `[]string`
- `Audit.Options` → `map[string]string` (Audit's custom comparison handles this explicitly)

### All Types in This Story Return `IsDeletable() = true`

Unlike story 1.5 where some config types returned `false`, every type in story 1.6 returns `true` for `IsDeletable()`.

### Implementation Pattern — Standard Go `testing` Package

All tests in `api/v1alpha1/` use standard Go `testing` package (NOT Ginkgo). Follow the exact pattern from existing tests in stories 1.1–1.5.

**Build tag**: No build tag needed — files in `api/v1alpha1/` run with default `go test`.

**Import pattern**:
```go
package v1alpha1

import (
    "reflect"
    "testing"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
```

### Previous Story Intelligence (Stories 1.1–1.5)

Established patterns to follow:
- Table-driven tests with `t.Run` subtests
- `reflect.DeepEqual` for map comparisons
- Testing both positive (matching) and negative (non-matching) cases
- Extra-fields behavior: documented as-is (tests proving behavior without changing production code)
- Tests for `GetPath` with and without `spec.name` override (where applicable)
- Tests for `IsDeletable` and `GetConditions`/`SetConditions`
- Unexported fields accessed directly within same package
- No build tags needed for `api/v1alpha1/` test files

Key insight from previous stories — custom `IsEquivalentToDesiredState` patterns vary per type:
- Story 1.3: `DatabaseSecretEngineConfig` remaps fields to match Vault's restructured read response
- Story 1.4: `LDAPAuthEngineConfig` deletes `bindpass` from desiredState
- Story 1.5: 4 types with custom logic (GitHub deletes `prv_key`, Quay deletes `password`, K8s deletes `service_account_jwt`, RabbitMQ uses different map method)
- This story: Policy mutates desired state (adds name, renames key); Audit uses field-by-field comparison; Entity/EntityAlias/Group/GroupAlias delete multiple Vault-only keys from payload; RandomSecret hardcoded false

### Project Structure Notes

Create 7 new test files:
- `api/v1alpha1/policy_test.go`
- `api/v1alpha1/randomsecret_test.go`
- `api/v1alpha1/audit_test.go`
- `api/v1alpha1/group_test.go`
- `api/v1alpha1/groupalias_test.go`
- `api/v1alpha1/entity_test.go`
- `api/v1alpha1/entityalias_test.go`

No changes to `controllers/` directory. No decoder methods needed (unit tests only).

### References

- [Source: api/v1alpha1/policy_types.go#L43-L54] — GetPath (4 variants)
- [Source: api/v1alpha1/policy_types.go#L55-L59] — GetPayload (1 key: "policy")
- [Source: api/v1alpha1/policy_types.go#L60-L68] — IsEquivalentToDesiredState (adds name, conditional rules rename)
- [Source: api/v1alpha1/policy_types.go#L74-L76] — IsDeletable (true)
- [Source: api/v1alpha1/policy_types.go#L78-L109] — PrepareInternalValues (auth accessor resolution)
- [Source: api/v1alpha1/policy_types.go#L119-L145] — PolicySpec struct
- [Source: api/v1alpha1/randomsecret_types.go#L103-L108] — GetPath
- [Source: api/v1alpha1/randomsecret_types.go#L114-L124] — getV1Payload (dynamic key, calculatedSecret)
- [Source: api/v1alpha1/randomsecret_types.go#L131-L138] — GetPayload (KV v1/v2 wrapper)
- [Source: api/v1alpha1/randomsecret_types.go#L140-L142] — IsEquivalentToDesiredState (always false)
- [Source: api/v1alpha1/randomsecret_types.go#L72] — calculatedSecret unexported field
- [Source: api/v1alpha1/audit_types.go#L116-L118] — GetPath (sys/audit/ prefix)
- [Source: api/v1alpha1/audit_types.go#L120-L127] — GetPayload (4 keys)
- [Source: api/v1alpha1/audit_types.go#L130-L155] — IsEquivalentToDesiredState (field-by-field, ignores extras)
- [Source: api/v1alpha1/audit_types.go#L104-L106] — VaultAuditObject interface
- [Source: api/v1alpha1/audit_types.go#L30-L68] — AuditSpec struct
- [Source: api/v1alpha1/group_types.go#L134-L139] — GetPath
- [Source: api/v1alpha1/group_types.go#L141-L143] — GetPayload
- [Source: api/v1alpha1/group_types.go#L145-L155] — toMap (3-5 keys, conditional)
- [Source: api/v1alpha1/group_types.go#L177-L181] — IsEquivalentToDesiredState (deletes name)
- [Source: api/v1alpha1/group_types.go#L49-L79] — GroupConfig struct
- [Source: api/v1alpha1/groupalias_types.go#L122-L124] — GetPath (uses Status.ID)
- [Source: api/v1alpha1/groupalias_types.go#L126-L128] — GetPayload
- [Source: api/v1alpha1/groupalias_types.go#L130-L137] — toMap (4 keys, retrieved* fields)
- [Source: api/v1alpha1/groupalias_types.go#L218-L229] — IsEquivalentToDesiredState (deletes 6 keys, has fmt.Print)
- [Source: api/v1alpha1/groupalias_types.go#L45-L59] — GroupAliasSpec with retrieved* fields
- [Source: api/v1alpha1/groupalias_types.go#L143-L204] — PrepareInternalValues
- [Source: api/v1alpha1/entity_types.go#L121-L126] — GetPath
- [Source: api/v1alpha1/entity_types.go#L128-L130] — GetPayload
- [Source: api/v1alpha1/entity_types.go#L132-L137] — toMap (3 keys)
- [Source: api/v1alpha1/entity_types.go#L160-L174] — IsEquivalentToDesiredState (deletes 11 keys)
- [Source: api/v1alpha1/entity_types.go#L49-L66] — EntityConfig struct
- [Source: api/v1alpha1/entityalias_types.go#L131-L133] — GetPath (uses Status.ID)
- [Source: api/v1alpha1/entityalias_types.go#L135-L137] — GetPayload
- [Source: api/v1alpha1/entityalias_types.go#L139-L148] — toMap (4-5 keys, conditional custom_metadata)
- [Source: api/v1alpha1/entityalias_types.go#L229-L240] — IsEquivalentToDesiredState (deletes 8 keys)
- [Source: api/v1alpha1/entityalias_types.go#L34-L58] — EntityAliasSpec with retrieved* fields
- [Source: api/v1alpha1/entityalias_types.go#L60-L73] — EntityAliasConfig struct
- [Source: api/v1alpha1/entityalias_types.go#L155-L214] — PrepareInternalValues
- [Source: _bmad-output/implementation-artifacts/1-5-unit-tests-tomap-and-isequivalenttodesiredstate-secret-engine-config-role-types.md] — Previous story patterns
- [Source: _bmad-output/project-context.md] — Testing rules and conventions

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
