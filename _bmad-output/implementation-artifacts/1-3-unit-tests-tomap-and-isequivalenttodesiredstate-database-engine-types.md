# Story 1.3: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Database Engine Types (Complex)

Status: ready-for-dev

## Story

As an operator developer,
I want unit tests for DatabaseSecretEngineConfig where Vault restructures fields in its read response (moving fields into `connection_details`),
So that the most complex `IsEquivalentToDesiredState` implementation is verified.

## Acceptance Criteria

1. **Given** a DatabaseSecretEngineConfig instance with `connectionURL`, `username`, `disableEscaping` populated **When** `IsEquivalentToDesiredState(payload)` is called with a Vault read response where those fields are nested under `connection_details` **Then** it returns `true` (correctly remaps fields to match Vault's structure)

2. **Given** a DatabaseSecretEngineConfig with `AllowedRoles` as `[]string` **When** compared against a Vault response with `allowed_roles` as `[]interface{}` **Then** it returns `true` (handles Go type differences via `toInterfaceArray` conversion)

3. **Given** a DatabaseSecretEngineConfig where `RootPasswordRotation.Enable` is true and `Status.LastRootPasswordRotation` is zero **When** `IsEquivalentToDesiredState` is called **Then** it returns `false` (forces initial rotation)

4. **Given** a DatabaseSecretEngineRole and DatabaseSecretEngineStaticRole with all fields populated **When** `toMap()` is called **Then** the returned maps contain correctly-named snake_case keys matching Vault API field names

5. **Given** each type and a matching Vault read response **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true`

6. **Given** each type and a Vault response with one managed field changed **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `false`

## Types Covered

| # | Type | File | Config Struct | Has Existing Tests | Test File |
|---|------|------|---------------|--------------------|-----------|
| 1 | DatabaseSecretEngineConfig | `api/v1alpha1/databasesecretengineconfig_types.go` | `DBSEConfig` (12+ keys, dynamic keys from `DatabaseSpecificConfig`) | **No** toMap/IsEquivalent tests | **None** — create new |
| 2 | DatabaseSecretEngineRole | `api/v1alpha1/databasesecretenginerole_types.go` | `DBSERole` (7 keys) | **No** tests | **None** — create new |
| 3 | DatabaseSecretEngineStaticRole | `api/v1alpha1/databasesecretenginestaticrole_types.go` | `DBSEStaticRole` (5-6 keys + conditional `credential_config`) | **No** tests | **None** — create new |

## Tasks / Subtasks

- [ ] Task 1: Add DatabaseSecretEngineConfig unit tests (AC: 1, 2, 3)
  - [ ] 1.1: Create `api/v1alpha1/databasesecretengineconfig_test.go`
  - [ ] 1.2: Add `TestDatabaseSecretEngineConfigGetPath` — with `spec.name`, without (fallback to `metadata.name`); verify path format `{path}/config/{name}`
  - [ ] 1.3: Add `TestDBSEConfigToMap` — verify always-set keys: `plugin_name`, `plugin_version`, `verify_connection`, `allowed_roles`, `root_credentials_rotate_statements`, `password_policy`, `connection_url`, `disable_escaping`
  - [ ] 1.4: Add `TestDBSEConfigToMapDatabaseSpecificConfig` — verify dynamic keys from `DatabaseSpecificConfig` map are merged into top-level payload
  - [ ] 1.5: Add `TestDBSEConfigToMapUsernameField` — test 3 branches: (a) `Username` set → uses it, (b) `Username` empty + `retrievedUsername` set → uses retrieved, (c) both empty → no `username` key
  - [ ] 1.6: Add `TestDBSEConfigToMapPasswordField` — test: (a) `retrievedPassword` set → `password` key present, (b) empty → no `password` key
  - [ ] 1.7: Add `TestDBSEConfigToMapAllowedRolesTypeConversion` — verify `AllowedRoles` produces `[]interface{}` (via `toInterfaceArray`), not `[]string`
  - [ ] 1.8: Add `TestDatabaseSecretEngineConfigIsEquivalentRootPasswordRotationGate` — `RootPasswordRotation.Enable=true` + `LastRootPasswordRotation.IsZero()` → `false`
  - [ ] 1.9: Add `TestDatabaseSecretEngineConfigIsEquivalentRootPasswordRotationWithTimestamp` — `RootPasswordRotation.Enable=true` + non-zero timestamp → proceeds to normal comparison
  - [ ] 1.10: Add `TestDatabaseSecretEngineConfigIsEquivalentConnectionDetailsRemapping` — verify `connection_url`, `disable_escaping`, `root_credentials_rotate_statements`, `username` are moved into `connection_details` sub-map; `password`, `connection_url`, `username`, `disable_escaping` are deleted from top-level
  - [ ] 1.11: Add `TestDatabaseSecretEngineConfigIsEquivalentMatching` — full matching payload with `connection_details` structure → `true`
  - [ ] 1.12: Add `TestDatabaseSecretEngineConfigIsEquivalentNonMatching` — one managed field changed → `false`
  - [ ] 1.13: Add `TestDatabaseSecretEngineConfigIsEquivalentExtraFieldsFiltered` — extra keys in payload are filtered out before comparison → `true` (this type DOES filter, unlike simpler types)
  - [ ] 1.14: Add `TestDatabaseSecretEngineConfigIsEquivalentRootRotateStatementsRemainAtTopLevel` — verify `root_credentials_rotate_statements` exists BOTH at top-level AND inside `connection_details` in desiredState
  - [ ] 1.15: Add `TestDatabaseSecretEngineConfigIsDeletable` — returns `true`
  - [ ] 1.16: Add `TestDatabaseSecretEngineConfigConditions` — GetConditions/SetConditions round-trip
  - [ ] 1.17: Add `TestDatabaseSecretEngineConfigGetRootPasswordRotationPath` — verify path format `{path}/rotate-root/{name}`
- [ ] Task 2: Add DatabaseSecretEngineRole unit tests (AC: 4, 5, 6)
  - [ ] 2.1: Create `api/v1alpha1/databasesecretenginerole_test.go`
  - [ ] 2.2: Add `TestDatabaseSecretEngineRoleGetPath` — with `spec.name`, without; verify format `{path}/roles/{name}`
  - [ ] 2.3: Add `TestDBSERoleToMap` — verify all 7 keys: `db_name`, `default_ttl`, `max_ttl`, `creation_statements`, `revocation_statements`, `rollback_statements`, `renew_statements`
  - [ ] 2.4: Add `TestDBSERoleToMapDurationTypes` — verify `default_ttl` and `max_ttl` are stored as `metav1.Duration` values (not strings or ints)
  - [ ] 2.5: Add `TestDatabaseSecretEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 2.6: Add `TestDatabaseSecretEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 2.7: Add `TestDatabaseSecretEngineRoleIsEquivalentExtraFields` — extra keys in payload → `false` (bare `reflect.DeepEqual`, no filtering)
  - [ ] 2.8: Add `TestDatabaseSecretEngineRoleIsDeletable` — returns `true`
  - [ ] 2.9: Add `TestDatabaseSecretEngineRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 3: Add DatabaseSecretEngineStaticRole unit tests (AC: 4, 5, 6)
  - [ ] 3.1: Create `api/v1alpha1/databasesecretenginestaticrole_test.go`
  - [ ] 3.2: Add `TestDatabaseSecretEngineStaticRoleGetPath` — with `spec.name`, without; verify format `{path}/static-roles/{name}`
  - [ ] 3.3: Add `TestDBSEStaticRoleToMap` — verify always-set keys: `db_name`, `username`, `rotation_period`, `rotation_statements`, `credential_type`
  - [ ] 3.4: Add `TestDBSEStaticRoleToMapPasswordCredentialConfig` — when `PasswordCredentialConfig` is non-nil, verify `credential_config` key contains `map[string]string{"password_policy": ...}`
  - [ ] 3.5: Add `TestDBSEStaticRoleToMapRSACredentialConfig` — when `RSAPrivateKeyCredentialConfig` is non-nil, verify `credential_config` key contains `map[string]string{"key_bits": "<string>", "format": ...}`
  - [ ] 3.6: Add `TestDBSEStaticRoleToMapNoCredentialConfig` — when both configs are nil, verify no `credential_config` key in map
  - [ ] 3.7: Add `TestDatabaseSecretEngineStaticRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 3.8: Add `TestDatabaseSecretEngineStaticRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 3.9: Add `TestDatabaseSecretEngineStaticRoleIsEquivalentExtraFields` — extra keys in payload → `false` (bare `reflect.DeepEqual`, no filtering)
  - [ ] 3.10: Add `TestDatabaseSecretEngineStaticRoleIsEquivalentCredentialConfigTypeMismatch` — `credential_config` as `map[string]string` vs `map[string]interface{}` → `false` (documents type coercion issue)
  - [ ] 3.11: Add `TestDatabaseSecretEngineStaticRoleIsDeletable` — returns `true`
  - [ ] 3.12: Add `TestDatabaseSecretEngineStaticRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 4: Add helper function tests (AC: 2)
  - [ ] 4.1: Add `TestToInterfaceArray` in the config test file — verify `[]string{"a","b"}` → `[]interface{}{"a","b"}`, empty slice → empty `[]interface{}{}`, nil slice → empty `[]interface{}{}`
  - [ ] 4.2: Add `TestPasswordCredentialConfigToMap` — verify produces `map[string]string{"password_policy": value}`
  - [ ] 4.3: Add `TestRSAPrivateKeyCredentialConfigToMap` — verify `key_bits` is string (`strconv.Itoa`), `format` is string
- [ ] Task 5: Verify all tests pass (AC: all)
  - [ ] 5.1: Run `go test ./api/v1alpha1/ -v -count=1` to confirm all new and existing tests pass
  - [ ] 5.2: Run `make test` to verify no regressions in full unit test suite

## Dev Notes

### Critical: DatabaseSecretEngineConfig Has the Most Complex `IsEquivalentToDesiredState` in the Operator

This is the only type that performs **field remapping** before comparison. The Vault `/database/config/{name}` read response restructures fields:

**Write payload (from `toMap()`):**
```
{plugin_name, plugin_version, verify_connection, allowed_roles, root_credentials_rotate_statements,
 password_policy, connection_url, username, password, disable_escaping, ...dynamic_keys}
```

**Vault read response restructures to:**
```
{plugin_name, plugin_version, verify_connection, allowed_roles, root_credentials_rotate_statements,
 password_policy, connection_details: {connection_url, disable_escaping, root_credentials_rotate_statements, username}, ...dynamic_keys}
```

`IsEquivalentToDesiredState` transforms `desiredState` to match Vault's structure:
1. Copies `connection_url`, `disable_escaping`, `root_credentials_rotate_statements`, `username` into a new `connection_details` sub-map
2. Deletes `password`, `connection_url`, `username`, `disable_escaping` from top-level
3. **Does NOT delete `root_credentials_rotate_statements` from top-level** — it remains duplicated in both top-level AND inside `connection_details`
4. Filters `payload` to only keys that exist in `desiredState` (or equal `"connection_details"`)
5. Compares with `reflect.DeepEqual`

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L93-L119]

### Critical: `root_credentials_rotate_statements` Duplication

After remapping, `desiredState` contains `root_credentials_rotate_statements` in **two places**:
- Top-level: `desiredState["root_credentials_rotate_statements"]`
- Nested: `desiredState["connection_details"]["root_credentials_rotate_statements"]`

This means the Vault read response **must also** have it in both places for `IsEquivalentToDesiredState` to return `true`. Test this explicitly.

### Critical: DatabaseSecretEngineConfig DOES Filter Extra Fields (Unlike Role/StaticRole)

`DatabaseSecretEngineConfig.IsEquivalentToDesiredState` builds a `filteredPayload` that only includes keys from the Vault response that exist in `desiredState`. This means extra Vault fields are correctly ignored — a **different** behavior from `DatabaseSecretEngineRole` and `DatabaseSecretEngineStaticRole` which use bare `reflect.DeepEqual`.

Test both behaviors:
- Config: extra fields → `true` (filtered)
- Role/StaticRole: extra fields → `false` (not filtered)

### Critical: Root Password Rotation Gate

`IsEquivalentToDesiredState` returns `false` immediately (without comparing state) when ALL of:
- `RootPasswordRotation != nil`
- `RootPasswordRotation.Enable == true`
- `Status.LastRootPasswordRotation.IsZero() == true`

This forces the reconciler to write to Vault (and trigger root password rotation) until a rotation timestamp is recorded. Test this gate with both zero and non-zero timestamps.

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L94-L96]

### `toMap()` Username/Password Branches

`DBSEConfig.toMap()` has conditional logic for `username` and `password`:

**Username (3 branches):**
1. `i.Username != ""` → `payload["username"] = i.Username`
2. `i.Username == ""` AND `i.retrievedUsername != ""` → `payload["username"] = i.retrievedUsername`
3. Both empty → no `username` key in payload

**Password (2 branches):**
1. `i.retrievedPassword != ""` → `payload["password"] = i.retrievedPassword`
2. `i.retrievedPassword == ""` → no `password` key

**Testing note:** `retrievedUsername` and `retrievedPassword` are **unexported** fields (json `"-"`). In unit tests within `package v1alpha1`, you CAN access them directly since the tests are in the same package.

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L364-L388]

### `toMap()` AllowedRoles Uses `toInterfaceArray` — Type Matters for DeepEqual

`toInterfaceArray([]string) []interface{}` converts `[]string` to `[]interface{}`. This is critical because Vault returns `allowed_roles` as `[]interface{}`, not `[]string`. The `toMap()` output already uses `[]interface{}`, so `DeepEqual` will match correctly.

Test that:
- `toInterfaceArray([]string{"role1", "role2"})` returns `[]interface{}{"role1", "role2"}`
- `toInterfaceArray(nil)` returns `[]interface{}{}` (empty, not nil — because `append` on nil creates a new slice)
- `toInterfaceArray([]string{})` returns `[]interface{}{}`

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L121-L127]

### DatabaseSecretEngineRole: `metav1.Duration` Values in toMap

`DBSERole.toMap()` stores `default_ttl` and `max_ttl` as `metav1.Duration` values directly in the map. The Vault API will return these as different types (likely numeric seconds or string). For unit testing, test with the same `metav1.Duration` type in the expected payload to match the `toMap()` output.

`creation_statements`, `revocation_statements`, `rollback_statements`, `renew_statements` are `[]string` — NOT converted via `toInterfaceArray`. This means if Vault returns them as `[]interface{}`, `reflect.DeepEqual` will return `false`. Document this in the test.

[Source: api/v1alpha1/databasesecretenginerole_types.go#L183-L193]

### DatabaseSecretEngineStaticRole: Credential Config Nested Map Type

`PasswordCredentialConfig.toMap()` and `RSAPrivateKeyCredentialConfig.toMap()` both return `map[string]string` (not `map[string]interface{}`). This nested map is stored as `payload["credential_config"]`. If Vault returns `credential_config` as `map[string]interface{}`, `reflect.DeepEqual` will return `false` due to map type mismatch.

Test this explicitly to document the behavior:
- `map[string]string{"password_policy": "foo"}` ≠ `map[string]interface{}{"password_policy": "foo"}`

[Source: api/v1alpha1/databasesecretenginestaticrole_types.go#L99-L121]

### RSA Key Bits: String Conversion

`RSAPrivateKeyCredentialConfig.toMap()` converts `KeyBits` (int) to string via `strconv.Itoa`. The map contains `"key_bits": "2048"` (string), not `"key_bits": 2048` (int). Test that the string conversion is correct.

### `DatabaseSpecificConfig` Dynamic Keys

`DBSEConfig.toMap()` iterates over `DatabaseSpecificConfig map[string]string` and merges each key/value into the top-level payload. These dynamic keys are NOT known at compile time — they depend on the database plugin (e.g., `tls_ca`, `tls_certificate_key` for MongoDB, `connect_timeout` for MySQL).

Test:
- Non-nil map with entries → keys merged into payload at top level
- Empty map → no extra keys
- Nil map → no extra keys (no panic)

### Implementation Pattern — Standard Go `testing` Package

All tests in `api/v1alpha1/` use the standard Go `testing` package (NOT Ginkgo). Follow the exact pattern from existing tests in `secretenginemount_test.go` and `identityoidc_test.go`.

**Build tag**: These files do NOT need a build tag — they are in `api/v1alpha1/` which runs with default `go test`.

**Import pattern**:
```go
package v1alpha1

import (
    "reflect"
    "testing"
    "time"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
```

### Previous Story Intelligence (Stories 1.1 and 1.2)

Stories 1.1 and 1.2 established these patterns:
- Table-driven tests with `t.Run` subtests
- `reflect.DeepEqual` for map comparisons
- Testing both positive (matching) and negative (non-matching) cases
- Extra-fields behavior: documented as-is (tests proving behavior without changing production code)
- Tests for `GetPath` with and without `spec.name` override
- Tests for `IsDeletable` and `GetConditions`/`SetConditions`
- No build tags needed for `api/v1alpha1/` test files
- Engine mount types (story 1.2) demonstrated tune-only comparison semantics

Key difference from stories 1.1/1.2: **DatabaseSecretEngineConfig has the only `IsEquivalentToDesiredState` that filters extra fields from the payload.** This is more robust than the bare `DeepEqual` used by the simpler types and the role/static-role types.

### Project Structure Notes

- Create `api/v1alpha1/databasesecretengineconfig_test.go` (new file)
- Create `api/v1alpha1/databasesecretenginerole_test.go` (new file)
- Create `api/v1alpha1/databasesecretenginestaticrole_test.go` (new file)
- No changes to `controllers/` directory
- No decoder methods needed (unit tests only)
- Existing controller test `controllers/databasesecretenginestaticrole_controller_test.go` is an integration test — unrelated to these unit tests

### References

- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L78-L83] — GetPath, GetRootPasswordRotationPath
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L90-L92] — GetPayload
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L93-L119] — IsEquivalentToDesiredState (complex field remapping + filtering)
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L121-L127] — toInterfaceArray helper
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L364-L388] — DBSEConfig.toMap()
- [Source: api/v1alpha1/databasesecretenginerole_types.go#L70-L75] — GetPath
- [Source: api/v1alpha1/databasesecretenginerole_types.go#L76-L82] — GetPayload, IsEquivalentToDesiredState
- [Source: api/v1alpha1/databasesecretenginerole_types.go#L183-L193] — DBSERole.toMap()
- [Source: api/v1alpha1/databasesecretenginestaticrole_types.go#L99-L103] — PasswordCredentialConfig.toMap()
- [Source: api/v1alpha1/databasesecretenginestaticrole_types.go#L116-L121] — RSAPrivateKeyCredentialConfig.toMap()
- [Source: api/v1alpha1/databasesecretenginestaticrole_types.go#L123-L137] — DBSEStaticRole.toMap()
- [Source: api/v1alpha1/databasesecretenginestaticrole_types.go#L151-L163] — GetPath, GetPayload, IsEquivalentToDesiredState
- [Source: _bmad-output/implementation-artifacts/1-1-unit-tests-tomap-and-isequivalenttodesiredstate-simple-standard-types.md] — Previous story patterns (simple types)
- [Source: _bmad-output/implementation-artifacts/1-2-unit-tests-tomap-and-isequivalenttodesiredstate-engine-mount-types.md] — Previous story patterns (engine mount types)
- [Source: _bmad-output/project-context.md] — Testing rules and conventions

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
