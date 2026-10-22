# Story 17.5: Reflective toMap via vault Struct Tags + VaultObject Interface Simplification

Status: done
## Story

As an operator developer,
I want the struct-to-Vault-payload mapping driven by struct tags with a single runtime reflection walker, and the VaultObject interface trimmed to only its essential methods,
So that adding a new CRD requires only annotated struct fields instead of writing 6+ boilerplate methods per type, and the ~2,000 lines of duplicated no-op methods are eliminated.

## Acceptance Criteria

### Part A0 — Unit Test Audit + Gap Fill (PREREQUISITE)
1. **Given** the 88 existing CRD types **When** a test audit is performed **Then** every type has at minimum: a `toMap()` output test, `IsEquivalentToDesiredState` positive/negative/extra-fields tests
2. **Given** the 9 types without test files (IdentityOIDCAssignment, IdentityOIDCClient, IdentityOIDCProvider, IdentityOIDCScope, IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole, TransitSecretEngineKey, VaultSecret) **When** Part A0 runs **Then** test files are created with full coverage
3. **Given** all existing test files **When** audited **Then** any test deriving expected payloads from `toMap()`/`GetPayload()` is rewritten with independent Vault-read-shaped fixtures

### Part A1 — StructToVaultPayload Walker
4. **Given** a struct with `vault:"key[,codec]"` tags **When** `StructToVaultPayload(&spec)` is called **Then** a `map[string]any` is returned with Vault API keys and correctly encoded values
5. **Given** a field tagged `vault:"-"` **When** processed **Then** the field is excluded
6. **Given** a field tagged `vault:"token_ttl,duration"` **When** encoded **Then** the value is converted via `durationToSeconds()` → `json.Number` seconds
7. **Given** a field tagged `vault:"token_num_uses,number"` **When** encoded **Then** the value is converted to `json.Number`
8. **Given** a field tagged `vault:",inline"` **When** processed **Then** the embedded struct's fields are merged into the parent payload
9. **Given** a field tagged `vault:"cidr_list,when=KeyType:otp"` and the struct has `KeyType == "ca"` **When** processed **Then** the field is excluded
10. **Given** a field tagged `vault:"cidr_list,when=KeyType:otp"` and `KeyType == "otp"` **When** processed **Then** the field is included
11. **Given** a `*bool` field tagged `vault:"verify_connection,default=true"` when `nil` **Then** the walker emits `true`; when non-nil, it emits the dereferenced value
12. **Given** a `*int` field tagged `vault:"max_retries,number,default=-1"` when `nil` **Then** the walker emits `json.Number("-1")`
13. **Given** a `map[string]string` field tagged `vault:",spreadmap"` **When** processed **Then** each key-value pair is merged into the parent payload
14. **Given** all existing types (except TransitSecretEngineKey) are migrated to vault struct tags **When** all Part A0 tests run **Then** all pass unchanged — zero behavioral difference

### Part B — Default IsEquivalentToDesiredState
15. **Given** a VaultObject that does NOT implement `VaultEquivalenceChecker` **When** compared **Then** `filterPayloadToDesiredKeys` + `reflect.DeepEqual` is the default
16. **Given** a VaultObject that DOES implement `VaultEquivalenceChecker` **When** compared **Then** it delegates to the type's custom method

### Part C — Optional No-Op Interface Methods
17. **Given** a VaultObject type with trivial `IsInitialized() → true`, `IsValid() → (true, nil)`, `PrepareInternalValues() → nil`, `PrepareTLSConfig() → nil` **When** those methods are removed **Then** the reconciler continues to function correctly

### Part D — ResourceName Helper
18. **Given** `vaultutils.ResourceName(specName, metaName)` **When** `specName` is non-empty **Then** it returns `specName`; otherwise `metaName`

## Tasks / Subtasks

### Phase A0: Unit Test Audit + Gap Fill (PREREQUISITE GATE — must complete before A1)
- [ ] Task 1 — Audit existing 79 test files for anti-pattern violations (AC: #3)
  - [ ] 1.1 Scan all `api/v1alpha1/*_test.go` for calls that derive expected payloads from `toMap()`/`GetPayload()`
  - [ ] 1.2 Rewrite violating tests with independent Vault-read-shaped fixtures (hardcoded `json.Number`, `durationToSeconds` values)
- [ ] Task 2 — Create test files for 9 untested types (AC: #1, #2)
  - [ ] 2.1 `identityoidcassignment_test.go` — GetPath, GetPayload, IsEquivalentToDesiredState ±, extra-fields
  - [ ] 2.2 `identityoidcclient_test.go` — same pattern
  - [ ] 2.3 `identityoidcprovider_test.go` — same pattern
  - [ ] 2.4 `identityoidcscope_test.go` — same pattern
  - [ ] 2.5 `identitytokenconfig_test.go` — same pattern
  - [ ] 2.6 `identitytokenkey_test.go` — same pattern
  - [ ] 2.7 `identitytokenrole_test.go` — same pattern
  - [ ] 2.8 `transitsecretenginekey_test.go` — both `toMap()` and `configToMap()` paths
  - [ ] 2.9 `vaultsecret_test.go` — GetPath, GetPayload, equivalence (KVv1 and KVv2 paths)
- [ ] Task 3 — Verify mode-dependent types have per-mode test coverage (AC: #1)
  - [ ] 3.1 `sshsecretenginerole_test.go` — verify tests for `key_type=otp` AND `key_type=ca` branches
  - [ ] 3.2 `totpsecretenginekey_test.go` — verify tests for `generate=true` AND `generate=false` branches
  - [ ] 3.3 `databasesecretengineconfig_test.go` — verify `connection_details` nesting test
- [ ] Task 4 — Run `make test` — all unit tests pass as regression baseline (AC: #1)

### Phase A1: StructToVaultPayload Walker
- [ ] Task 5 — Implement walker in `api/v1alpha1/utils/` (AC: #4–#13)
  - [ ] 5.1 Create `api/v1alpha1/utils/struct_to_vault_payload.go`
  - [ ] 5.2 Parse extended tag grammar: `vault:"<key>[,<codec>][,omitempty][,when=Field:value][,default=X][,spreadmap]"` | `vault:"-"` | `vault:",inline"`
  - [ ] 5.3 Implement identity codec (default — pass-through, with `[]string` → `[]any` conversion)
  - [ ] 5.4 Implement `duration` codec: call `durationToSeconds()` (must relocate from `sshsecretenginerole_types.go` to `utils/`)
  - [ ] 5.5 Implement `number` codec: `json.Number(strconv.Itoa(val))` for int, `json.Number(strconv.FormatFloat(...))` for float
  - [ ] 5.6 Implement `intmap` codec: `map[string]int` → `map[string]any{json.Number(...)}` (e.g., `AllowedUserKeyLengths`)
  - [ ] 5.7 Implement `stringmap` codec: `map[string]string` → `map[string]any` (e.g., `DefaultExtensions`)
  - [ ] 5.8 Implement `when=Field:value` conditional directive (resolve by Go struct field name)
  - [ ] 5.9 Implement `default=X` pointer-nil defaulting for `*bool` and `*int`
  - [ ] 5.10 Implement `spreadmap` directive: merge map entries into parent
  - [ ] 5.11 Implement `inline` directive: recurse into embedded struct
  - [ ] 5.12 Implement `omitempty` directive: skip zero-value fields
  - [ ] 5.13 Handle `vault:"-"` to exclude fields (Connection, Authentication, Path, credential structs)
- [ ] Task 6 — Unit tests for the walker itself (AC: #4–#13)
  - [ ] 6.1 Create `api/v1alpha1/utils/struct_to_vault_payload_test.go`
  - [ ] 6.2 Test each codec: identity, duration, number, intmap, stringmap
  - [ ] 6.3 Test each directive: omitempty, when, default, spreadmap, inline, skip(-)
  - [ ] 6.4 Test nested struct via inline
  - [ ] 6.5 Test nil pointer fields with and without default
  - [ ] 6.6 Test conditional inclusion/exclusion
- [ ] Task 7 — Relocate shared helpers to `api/v1alpha1/utils/` (AC: #4)
  - [ ] 7.1 Move `durationToSeconds()` from `sshsecretenginerole_types.go` → `utils/codecs.go` (exported)
  - [ ] 7.2 Move `toAnyMapJsonNumber()` → `utils/codecs.go` (exported)
  - [ ] 7.3 Move `toInterfaceArray()` → `utils/codecs.go` (exported)
  - [ ] 7.4 Update all 80+ callers to use the `utils` package versions
  - [ ] 7.5 Delete the per-file duplicate definitions
- [ ] Task 8 — Annotate all 88 type files with vault struct tags (AC: #14)
  - [ ] 8.1 Start with ~10 trivial types (PasswordPolicy, Policy, Namespace, etc.)
  - [ ] 8.2 Standard types (~55 types — roles, configs with simple field mapping)
  - [ ] 8.3 Mode-dependent types: SSHSecretEngineRole (`when=KeyType:otp`/`when=KeyType:ca`), TOTPSecretEngineKey (`when=Generate:true`/`when=Generate:false`)
  - [ ] 8.4 Credential-override types: move override logic from `toMap()` to `PrepareInternalValues()` for LDAPAuthEngineConfig, LDAPSecretEngineConfig, AWSSecretEngineConfig, DatabaseSecretEngineConfig, SSHSecretEngineConfig, OktaAuthEngineConfig, etc.
  - [ ] 8.5 Spread-map types: DatabaseSecretEngineConfig `DatabaseSpecificConfig`
  - [ ] 8.6 Pointer-default types: fields using `*bool` and `*int` (e.g., `verify_connection`, `max_retries`)
  - [ ] 8.7 **Exclude** TransitSecretEngineKey — keep hand-written `toMap()`/`configToMap()` (two maps from one struct)
  - [ ] 8.8 Replace `toMap()` body with `return vaultutils.StructToVaultPayload(&d)` for each migrated type
  - [ ] 8.9 Run Part A0 tests after EACH batch — all must pass unchanged

### Phase B: Default IsEquivalentToDesiredState
- [ ] Task 9 — Implement VaultEquivalenceChecker optional interface (AC: #15, #16)
  - [ ] 9.1 Remove `IsEquivalentToDesiredState` from `VaultObject` interface in `api/v1alpha1/utils/vaultobject.go`
  - [ ] 9.2 Define `VaultEquivalenceChecker` interface: `IsEquivalentToDesiredState(payload map[string]any) bool`
  - [ ] 9.3 Update `VaultEndpoint.CreateOrUpdate()` to type-assert `VaultEquivalenceChecker`; absent = default comparison (`filterPayloadToDesiredKeys` + `reflect.DeepEqual`)
  - [ ] 9.4 Update `RabbitMQEngineConfigVaultEndpoint.CreateOrUpdateLease()` similarly
- [ ] Task 10 — Remove standard IsEquivalentToDesiredState from ~60 types (AC: #15)
  - [ ] 10.1 Identify the ~60 types whose `IsEquivalentToDesiredState` is: `desiredState := d.GetPayload(); return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))`
  - [ ] 10.2 Delete those methods — the default in `CreateOrUpdate` now provides the same behavior
  - [ ] 10.3 Run all unit + integration tests — zero behavioral change
- [ ] Task 11 — Keep custom IsEquivalentToDesiredState on ~25 types (AC: #16)
  - [ ] 11.1 Types that use `removeUnsetFields` (auth role types with token aliases)
  - [ ] 11.2 Types that use `normalizeVaultReadAliases` (AppRole, AliCloud, CF, etc.)
  - [ ] 11.3 Types that delete write-only secrets (GitHub, Kubernetes, LDAP, AWS, DB, Okta, MongoDB, etc.)
  - [ ] 11.4 Types with field remapping (DatabaseSecretEngineConfig `connection_details`)
  - [ ] 11.5 Engine mount types (AuthEngineMount, SecretEngineMount — tune comparison)
  - [ ] 11.6 Types with `readVisibleMap()` (TOTPSecretEngineKey)
  - [ ] 11.7 Types with Nomad-style `type`→`token_type` aliasing

### Phase C: Optional No-Op Interface Methods
- [ ] Task 12 — Define optional interfaces in `vaultobject.go` (AC: #17)
  - [ ] 12.1 Remove `IsInitialized`, `IsValid`, `PrepareInternalValues`, `PrepareTLSConfig` from required `VaultObject` interface
  - [ ] 12.2 Define: `VaultInitializer { IsInitialized() bool }`
  - [ ] 12.3 Define: `VaultValidator { IsValid() (bool, error) }`
  - [ ] 12.4 Define: `VaultInternalValuesPreparer { PrepareInternalValues(context.Context, client.Object) error }`
  - [ ] 12.5 Define: `VaultTLSPreparer { PrepareTLSConfig(context.Context, client.Object) error }`
- [ ] Task 13 — Update framework callers to type-assert optional interfaces (AC: #17)
  - [ ] 13.1 Update `reconcile_skeleton.go` / `ReconcileWithFunctions` — type-assert `VaultInitializer`, default to `true`
  - [ ] 13.2 Update reconciler — type-assert `VaultValidator`, default to `(true, nil)`
  - [ ] 13.3 Update reconciler — type-assert `VaultInternalValuesPreparer`, default to `nil`
  - [ ] 13.4 Update reconciler — type-assert `VaultTLSPreparer`, default to `nil`
- [ ] Task 14 — Remove no-op methods from types (AC: #17)
  - [ ] 14.1 Delete `IsInitialized() → true` from ALL types (0 types have non-trivial implementations)
  - [ ] 14.2 Delete `IsValid() → (true, nil)` from ~60 types; keep on 4 types with actual validation: UserpassAuthEngineUser, TerraformCloudSecretEngineRole, TerraformCloudSecretEngineConfig, GCPSecretEngineConfig
  - [ ] 14.3 Delete `PrepareInternalValues() → nil` from ~50 types; keep on ~18 types with non-trivial logic (credential resolution, namespace population, accessor lookup, password generation, policy templates)
  - [ ] 14.4 Delete `PrepareTLSConfig() → nil` from ALL types (0 types have non-trivial implementations)
  - [ ] 14.5 Run all tests — zero behavioral change

### Phase D: ResourceName Helper
- [ ] Task 15 — Implement and apply ResourceName (AC: #18)
  - [ ] 15.1 Add `func ResourceName(specName, metaName string) string` in `api/v1alpha1/utils/helpers.go`
  - [ ] 15.2 Replace ~48 identical `if d.Spec.Name != "" { ... } else { ... }` branches in `GetPath()` methods
  - [ ] 15.3 Run all tests

### Phase E: Documentation Update
- [ ] Task 16 — Update "Contributing a New Vault API" developer guide (AC: #14, #15, #17)
  - [ ] 16.1 Rewrite `docs/contributing-vault-apis.md` to reflect the new workflow:
    - Step 2: Show `vault:"key,codec"` struct tags instead of hand-written `toMap()`
    - Step 3: `VaultObject` now has only 5 required methods (`GetPath`, `GetPayload`, `IsDeletable`, `GetKubeAuthConfiguration`, `GetVaultConnection`); show `GetPayload` delegating to `vaultutils.StructToVaultPayload`
    - Step 3: Explain optional interfaces (`VaultEquivalenceChecker`, `VaultValidator`, `VaultInternalValuesPreparer`) — only implement if non-trivial logic needed
    - Step 3: Remove boilerplate no-ops (`IsInitialized`, `PrepareTLSConfig`, `IsValid`, `PrepareInternalValues`) from the example — they're optional now
    - Step 5: Update webhook example to current `admission.Defaulter[*T]`/`admission.Validator[*T]` generic signatures (currently shows obsolete `runtime.Object`)
    - Step 6: Show `ResourceName` helper in `GetPath()` example
    - Add tag grammar reference table (codecs: duration, number, intmap, stringmap; directives: omitempty, when, default, spreadmap, inline, -)
    - Add "what NOT to do" section: never hand-write `toMap()` for new types (except Transit-style two-map cases)
  - [ ] 16.2 Update `_bmad-output/project-context.md` "Adding a New Vault API Type (Full Checklist)" section (~line 300) to match the new workflow
  - [ ] 16.3 Update `_bmad-output/project-context.md` "Imperative-to-Declarative Bridge" section (~line 84) to document `StructToVaultPayload` as the primary mechanism and `vault:` struct tags as the source of truth

### Phase F: Final Validation
- [ ] Task 17 — Full test suite + quality gates
  - [ ] 17.1 `make manifests generate fmt vet`
  - [ ] 17.2 `make test` — all unit tests pass
  - [ ] 17.3 `make lint` — zero violations
  - [ ] 17.4 Verify no new CRDs added (this story is refactoring only, no new types)

## Dev Notes

### ⚠️ Critical: This is a Shared Framework Refactor — Highest Risk Story in Epic 17

This story modifies the core `VaultObject` interface implemented by all 88 CRD types. Every change has blast radius across the entire operator. The Part A0 test audit is the safety net — it **MUST** be complete and passing before any migration begins.

**Mandatory ordering:** A0 → A1 → B → C → D. Each phase must pass all tests before the next begins. Never skip A0.

### Types Without Test Files (9 types — must be filled in Part A0)

| Type | Notes |
|---|---|
| `IdentityOIDCAssignment` | Standard pattern — `GetPath`, `GetPayload`, equivalence |
| `IdentityOIDCClient` | Has `client_type` enum field |
| `IdentityOIDCProvider` | Multiple list fields |
| `IdentityOIDCScope` | Simple single-field template |
| `IdentityTokenConfig` | Singleton pattern with `issuer` field |
| `IdentityTokenKey` | Duration and number fields |
| `IdentityTokenRole` | Duration fields, list fields |
| `TransitSecretEngineKey` | Two map methods: `toMap()` for create and `configToMap()` for tune — test BOTH |
| `VaultSecret` | KVv1 and KVv2 paths, template rendering — complex |

### Non-Trivial Implementations to KEEP (Do NOT Delete)

#### Non-Trivial `IsEquivalentToDesiredState` (~25 types)
These types have custom comparison logic and must retain their hand-written implementations, implementing the new `VaultEquivalenceChecker` interface:

| Category | Types |
|---|---|
| **Secret-stripping** | AWSSecretEngineConfig, GitHubSecretEngineConfig, KubernetesSecretEngineConfig, LDAPAuthEngineConfig, LDAPSecretEngineConfig, QuaySecretEngineConfig, ConsulSecretEngineConfig, MongoDBAtlasSecretEngineConfig, SSHSecretEngineConfig, OktaAuthEngineConfig, RADIUSAuthEngineConfig, NomadSecretEngineConfig, KerberosAuthEngineConfig, KerberosAuthEngineLDAPConfig, GCPSecretEngineConfig |
| **Vault-read alias normalization** | AppRoleAuthEngineRole, AliCloudAuthEngineRole, CFAuthEngineRole, OCIAuthEngineRole, AWSAuthEngineRole, AzureAuthEngineRole, GCPAuthEngineRole, JWTOIDCAuthEngineRole, KubernetesAuthEngineRole, KerberosAuthEngineGroup, LDAPAuthEngineGroup, OktaAuthEngineGroup, RADIUSAuthEngineUser, UserpassAuthEngineUser |
| **Field remapping** | DatabaseSecretEngineConfig (`connection_details` nesting) |
| **Tune comparison** | AuthEngineMount, SecretEngineMount (compare tune config not full spec) |
| **Type aliasing** | NomadSecretEngineRole (`type` → `token_type`) |
| **Read-visible subset** | TOTPSecretEngineKey (`readVisibleMap()`) |
| **Credential-change detection** | DatabaseSecretEngineConfig (`credentialsChanged` flag) |
| **GCP bindings normalization** | GCPSecretEngineRoleset, GCPSecretEngineStaticAccount |
| **RabbitMQ dual-map** | RabbitMQSecretEngineConfig (lease vs connection comparison) |

#### Non-Trivial `PrepareInternalValues` (~18 types)
| Category | Types |
|---|---|
| **Credential resolution** (`setInternalCredentials`) | LDAPAuthEngineConfig, LDAPSecretEngineConfig, AWSSecretEngineConfig, DatabaseSecretEngineConfig, SSHSecretEngineConfig, GitHubSecretEngineConfig, OktaAuthEngineConfig, KubernetesSecretEngineConfig, ConsulSecretEngineConfig, MongoDBAtlasSecretEngineConfig, TerraformCloudSecretEngineConfig, GCPSecretEngineConfig, NomadSecretEngineConfig, RADIUSAuthEngineConfig, KerberosAuthEngineLDAPConfig |
| **Accessor lookup** | EntityAlias, GroupAlias |
| **Namespace population** | KubernetesAuthEngineRole, Group |
| **Keytab from secret** | KerberosAuthEngineConfig |
| **Password generation** | RandomSecret |
| **Template resolution** | Policy |

#### Non-Trivial `IsValid` (4 types)
UserpassAuthEngineUser, TerraformCloudSecretEngineRole, TerraformCloudSecretEngineConfig, GCPSecretEngineConfig

#### TransitSecretEngineKey — EXCLUDED from Walker Migration
This type has two map methods (`toMap()` for initial creation, `configToMap()` for tune updates) producing different maps from the same struct. The walker produces a single map. A `group=` tag feature is not worth adding for one type. Keep hand-written methods.

### Walker Tag Grammar Reference

```
vault:"<vault_key>[,<codec>][,omitempty][,when=Field:value][,default=X][,spreadmap]"
vault:"-"          # Exclude this field entirely
vault:",inline"    # Merge embedded struct fields into parent
```

**Codecs (closed set — no extensibility needed):**
| Codec | Input Types | Output | Example |
|---|---|---|---|
| _(identity/default)_ | `string`, `bool`, `[]string` | pass-through; `[]string` → `[]any` | `vault:"allowed_roles"` |
| `duration` | `string` (Go duration) | `json.Number` (seconds) | `vault:"token_ttl,duration"` |
| `number` | `int`, `int64` | `json.Number` | `vault:"token_num_uses,number"` |
| `intmap` | `map[string]int` | `map[string]any{json.Number}` | `vault:"allowed_user_key_lengths,intmap"` |
| `stringmap` | `map[string]string` | `map[string]any` | `vault:"default_extensions,stringmap"` |

**Directives:**
| Directive | Description | Use Case |
|---|---|---|
| `omitempty` | Skip when zero value | Optional fields |
| `when=Field:value` | Include only when sibling Go field == value | SSH `key_type`, TOTP `generate` |
| `default=X` | Emit X when pointer is nil | `*bool verify_connection`, `*int max_retries` |
| `spreadmap` | Merge map entries into parent payload | DB `DatabaseSpecificConfig` |
| `inline` | Recurse into embedded struct | All inline config structs |
| `-` | Skip field entirely | Connection, Authentication, Path |

### Credential-Override Type Migration Strategy

Several types (LDAP, AWS, DB, SSH, Okta, Consul, MongoDB, etc.) have `toMap()` methods that read from unexported `retrieved*` fields populated by `PrepareInternalValues` → `setInternalCredentials`. Example pattern in `LDAPAuthEngineConfig`:

```go
func (config *LDAPAuthConfig) toMap() map[string]any {
    // ... normal fields ...
    if config.retrievedPassword != "" {
        result["bindpass"] = config.retrievedPassword
    }
    // ...
}
```

**Migration approach:** The `retrieved*` fields should have normal vault tags (e.g., `vault:"bindpass,omitempty"`). The `PrepareInternalValues` / `setInternalCredentials` already populates these fields before `toMap()` is called. The walker will emit them via tags exactly as the hand-written code does. The `omitempty` directive ensures they're excluded when empty (before credential resolution runs). No logic change needed — just tag annotation.

### DatabaseSecretEngineConfig `spreadmap` Pattern

The `DBSEConfig.toMap()` merges `DatabaseSpecificConfig` (a `map[string]string`) into the parent payload:

```go
for key, value := range d.DatabaseSpecificConfig {
    payload[key] = value
}
```

The walker handles this via `vault:",spreadmap"` on the `DatabaseSpecificConfig` field.

### Helper Relocation (Task 7 — Critical for Walker)

These functions are currently defined in individual type files and duplicated. They must move to `api/v1alpha1/utils/` to be shared by the walker:

| Function | Current Location | New Location |
|---|---|---|
| `durationToSeconds()` | `sshsecretenginerole_types.go` | `utils/codecs.go` (exported: `DurationToSeconds`) |
| `toAnyMapJsonNumber()` | `sshsecretenginerole_types.go` | `utils/codecs.go` (exported: `ToAnyMapJsonNumber`) |
| `toInterfaceArray()` | `databasesecretengineconfig_types.go` | `utils/codecs.go` (exported: `ToInterfaceArray`) |

All ~80 callers must be updated. These helpers are used by both `toMap()` methods and `IsEquivalentToDesiredState` methods, so they must be exported.

### Interface Change Summary

**Before (VaultObject — 10 required methods):**
```go
type VaultObject interface {
    GetPath() string
    GetPayload() map[string]any
    IsEquivalentToDesiredState(payload map[string]any) bool
    IsInitialized() bool
    IsValid() (bool, error)
    IsDeletable() bool
    PrepareInternalValues(context.Context, client.Object) error
    PrepareTLSConfig(context.Context, client.Object) error
    GetKubeAuthConfiguration() *KubeAuthConfiguration
    GetVaultConnection() *VaultConnection
}
```

**After (VaultObject — 5 required methods + 5 optional interfaces):**
```go
type VaultObject interface {
    GetPath() string
    GetPayload() map[string]any
    IsDeletable() bool
    GetKubeAuthConfiguration() *KubeAuthConfiguration
    GetVaultConnection() *VaultConnection
}

// Optional interfaces — absent means default/no-op behavior
type VaultEquivalenceChecker interface { IsEquivalentToDesiredState(map[string]any) bool }
type VaultInitializer interface { IsInitialized() bool }
type VaultValidator interface { IsValid() (bool, error) }
type VaultInternalValuesPreparer interface { PrepareInternalValues(context.Context, client.Object) error }
type VaultTLSPreparer interface { PrepareTLSConfig(context.Context, client.Object) error }
```

### Framework Files Modified (ESCALATION-LEVEL CHANGES)

| File | Change | Risk |
|---|---|---|
| `api/v1alpha1/utils/vaultobject.go` | Remove 5 methods from `VaultObject`, add 5 optional interfaces | **HIGH** — interface contract change |
| `api/v1alpha1/utils/vaultobject.go` (CreateOrUpdate) | Add default `IsEquivalentToDesiredState` logic | **HIGH** — affects all 88 types |
| `controllers/vaultresourcecontroller/reconcile_skeleton.go` | Type-assert optional interfaces | **HIGH** — affects all reconcilers |
| `api/v1alpha1/utils/struct_to_vault_payload.go` | **NEW FILE** — walker implementation | LOW — additive |
| `api/v1alpha1/utils/codecs.go` | **NEW FILE** — relocated helpers | LOW — additive (callers updated) |
| `api/v1alpha1/utils/helpers.go` | Add `ResourceName` helper | LOW — additive |
| `docs/contributing-vault-apis.md` | **REWRITE** — update for vault tags, reduced interface, new webhooks | LOW — documentation only |
| `_bmad-output/project-context.md` | Update "Adding a New Vault API" checklist + "Imperative-to-Declarative Bridge" | LOW — documentation only |

**Shared framework protection rule applies:** The reconcile skeleton and VaultEndpoint behavioral contract must not change. The only change is adding type-assertion fall-through paths. Existing types that implement the full old interface will still work — the optional interfaces are a superset.

### Compile Safety Net

All 88 types currently have `var _ vaultutils.VaultObject = &MyType{}` compile-time checks. After the interface change, these still compile because the 5 remaining required methods (`GetPath`, `GetPayload`, `IsDeletable`, `GetKubeAuthConfiguration`, `GetVaultConnection`) are still on every type. Types that keep custom `IsEquivalentToDesiredState` should add: `var _ vaultutils.VaultEquivalenceChecker = &MyType{}`.

### Migration Order for Part A1 Tag Annotation

1. **Trivial types** (~10): PasswordPolicy, Policy, Namespace, AuditRequestHeader, Audit, IdentityOIDCScope, IdentityOIDCAssignment, IdentityOIDCClient, IdentityOIDCProvider, IdentityTokenConfig
2. **Standard roles/configs** (~55): auth engine roles, secret engine roles, simple configs
3. **Mode-dependent** (3): SSHSecretEngineRole, TOTPSecretEngineKey, DatabaseSecretEngineStaticRole
4. **Credential-override** (~15): LDAP, AWS, DB, SSH, Okta, Consul, MongoDB, GCP, TF Cloud, Nomad, Kerberos, RADIUS, GitHub, Kubernetes, Quay
5. **Spread-map** (1): DatabaseSecretEngineConfig
6. **Pointer-default** (~5): types with `*bool`/`*int` fields
7. **Skip**: TransitSecretEngineKey

Run Part A0 tests after **each batch**. If any test fails, stop and diagnose before proceeding.

### No New CRDs, No New Dependencies

This story is pure refactoring. No new CRDs are added. No new Go dependencies are needed — `reflect` is stdlib, `encoding/json` is stdlib, `strconv` is stdlib. The walker uses only runtime reflection.

### Documentation — Developer Guide Must Be Updated

While there are no new user-facing CRD docs, the **developer-facing** guide for adding new types must be rewritten:

- **`docs/contributing-vault-apis.md`** — The current guide is significantly outdated. It still shows `runtime.Object` webhook signatures (replaced by generic `admission.Defaulter[*T]`/`admission.Validator[*T]` since controller-runtime v0.24), hand-written `toMap()` methods (now replaced by `vault:` struct tags + `StructToVaultPayload`), and the full 10-method `VaultObject` interface (now trimmed to 5 required + 5 optional). The guide must be rewritten end-to-end so that anyone contributing a new Vault API type follows the new pattern.
- **`_bmad-output/project-context.md`** — The "Adding a New Vault API Type" checklist (step 3: implement VaultObject) and the "Imperative-to-Declarative Bridge" section must reflect `StructToVaultPayload` as the primary payload mechanism and document the reduced interface contract.

This is critical: if the docs aren't updated, future CRD contributors (human or AI) will follow the old boilerplate pattern, defeating the purpose of this refactor.

- **Novelty Risk:** HIGH — shared framework refactor across all 88 types. Part A0 test audit is the safety net. Integration tests from Epics 2–7 provide end-to-end regression coverage.
- Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings.

### Project Structure Notes

- Walker lives in `api/v1alpha1/utils/` alongside existing VaultObject and VaultEndpoint — correct placement for shared type utilities
- Codec helpers relocate to `api/v1alpha1/utils/codecs.go` — co-located with the walker that uses them
- No new controller files, no new webhook files, no new CRD files
- Test files remain in `api/v1alpha1/` co-located with their type files

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 17, Story 17.5, lines 3067–3210]
- [Source: api/v1alpha1/utils/vaultobject.go — VaultObject interface, VaultEndpoint.CreateOrUpdate]
- [Source: api/v1alpha1/sshsecretenginerole_types.go — durationToSeconds, toAnyMapJsonNumber, conditional toMap]
- [Source: api/v1alpha1/databasesecretengineconfig_types.go — spreadmap pattern, credential-override, custom IsEquivalentToDesiredState]
- [Source: api/v1alpha1/passwordpolicy_types.go — simplest VaultObject pattern reference]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys, normalizeVaultReadAliases]
- [Source: controllers/vaultresourcecontroller/reconcile_skeleton.go — ReconcileWithFunctions callers of optional methods]
- [Source: docs/contributing-vault-apis.md — Developer guide for adding new CRD types (must be rewritten)]
- [Source: _bmad-output/project-context.md — all critical implementation rules, "Adding a New Vault API Type" checklist]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

### Review Model Used

- Quick lens: GPT 5.4 Medium (adversarial review)

### Review Findings

Pass 1 — 6 findings: 0 high, 2 medium, 2 low, 2 false
- `[false]` `[reject]` DurationToSeconds() returns json.Number(s) on parse failure — pre-migration had identical behavior
- `[low]` `[reject]` Identity token/OIDC types missing duration codec on TTL fields — migration correctly preserved pre-existing behavior
- `[low]` `[reject]` isequivalent_audit_test.go still uses toMap()/GetPayload() — tests comparison behavior, not payload output
- `[low]` `[reject]` vaultsecret_test.go coverage incomplete — VaultSecret excluded from walker migration
- `[medium]` `[patch]` ~22 types still had trivial IsInitialized, PrepareTLSConfig after Phase C — fixed: deleted remaining no-ops across 24 files
- `[medium]` `[patch]` Local helper duplicates (durationToSeconds, toInterfaceArray) still in SSH/DB type files — fixed: deleted dead functions

### Decisions Needed / Decisions Taken

None — all findings resolved during review pass.

### Fixes Applied

- Deleted remaining trivial no-op interface methods across 24 files (Phase C completion)
- Deleted dead local helper functions from sshsecretenginerole_types.go and databasesecretengineconfig_types.go
