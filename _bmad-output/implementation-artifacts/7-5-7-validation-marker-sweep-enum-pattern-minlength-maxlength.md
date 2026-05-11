# Story 7.5.7: Validation Marker Sweep — Enum, Pattern, MinLength/MaxLength

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want validation markers (`Enum`, `Pattern`, `Minimum`/`Maximum`) added to fields across all CRD types where the accepted values are clearly constrained,
So that invalid values are rejected at admission time rather than failing at Vault API call time.

## Acceptance Criteria

1. **Given** `TokenType` fields across all auth role types and mount tune **When** `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}` added **Then** invalid token types rejected at admission
2. **Given** `GCPRole.Type` (iam/gce), `JWTOIDCRole.RoleType` (oidc/jwt), `PKICommon.PrivateKeyFormat` (der/pkcs8), `AzureSERole.SignInAudience` **When** `Enum` markers added with documented values **Then** invalid values rejected at admission
3. **Given** numeric fields with documented bounds (`OCSPCacheSize`, `RoleCacheSize`, `OCSPMaxRetries`, `MaxPathLength`, `RefreshThreshold`, `ApplicationID`, token tuning) **When** `Minimum`/`Maximum` markers added **Then** out-of-range values rejected at admission
4. **Given** all existing `Enum` markers on PKI, identity, database, Quay, audit, policy, Kubernetes, group, mount types **When** audited **Then** confirmed complete with no missing values
5. **Given** all changes **When** `make manifests generate fmt vet test` passes **Then** no regressions

## Tasks / Subtasks

- [ ] Task 1: Add `TokenType` Enum to remaining auth role types and mount tune (AC: 1)
  - [ ] 1.1: `certauthenginerole_types.go` — Add `// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}` to `TokenType`
  - [ ] 1.2: `azureauthenginerole_types.go` — Add same Enum to `TokenType`
  - [ ] 1.3: `gcpauthenginerole_types.go` — Add same Enum to `TokenType`
  - [ ] 1.4: `authenginemount_types.go` — Add same Enum to `AuthMountConfig.TokenType`
- [ ] Task 2: Add other Enum markers for closed-value-set fields (AC: 2)
  - [ ] 2.1: `gcpauthenginerole_types.go` — Add `// +kubebuilder:validation:Enum={"iam","gce"}` to `GCPRole.Type`
  - [ ] 2.2: `jwtoidcauthenginerole_types.go` — Add `// +kubebuilder:validation:Enum={"oidc","jwt"}` to `JWTOIDCRole.RoleType`
  - [ ] 2.3: `pkisecretengineconfig_types.go` — Add `// +kubebuilder:validation:Enum={"der","pkcs8"}` to `PKICommon.PrivateKeyFormat`
  - [ ] 2.4: `azuresecretenginerole_types.go` — Add `// +kubebuilder:validation:Enum={"AzureADMyOrg","AzureADMultipleOrgs","AzureADandPersonalMicrosoftAccount","PersonalMicrosoftAccount"}` to `AzureSERole.SignInAudience`
- [ ] Task 3: Add `Minimum`/`Maximum` markers for numeric fields (AC: 3)
  - [ ] 3.1: `certauthengineconfig_types.go` — `OCSPCacheSize`: `Minimum=0`; `RoleCacheSize`: `Minimum=-1`
  - [ ] 3.2: `certauthenginerole_types.go` — `OCSPMaxRetries`: `Minimum=0`; `TokenNumUses`: `Minimum=0`
  - [ ] 3.3: `vaultsecret_types.go` — `RefreshThreshold`: `Minimum=1`, `Maximum=100`
  - [ ] 3.4: `githubsecretengineconfig_types.go` — `ApplicationID`: `Minimum=1`
  - [ ] 3.5: `pkisecretengineconfig_types.go` — `MaxPathLength`: `Minimum=-1`
  - [ ] 3.6: `kubernetesauthenginerole_types.go` — `Minimum=0` on `TokenTTL`, `TokenMaxTTL`, `TokenExplicitMaxTTL`, `TokenPeriod`, `TokenNumUses`
  - [ ] 3.7: `gcpauthenginerole_types.go` — `Minimum=0` on `TokenNumUses`, `TokenPeriod`
  - [ ] 3.8: `azureauthenginerole_types.go` — `Minimum=0` on `TokenNumUses`, `TokenPeriod`
  - [ ] 3.9: `ldapauthengineconfig_types.go` — `Minimum=0` on `TokenNumUses`, `TokenPeriod`
  - [ ] 3.10: `jwtoidcauthenginerole_types.go` — `Minimum=0` on `TokenNumUses`, `TokenPeriod`, `MaxAge`
- [ ] Task 4: Audit existing Enum markers for completeness (AC: 4)
  - [ ] 4.1: Confirm `PKIRole.KeyType` Enum `{"rsa","ec"}` — decide whether to add `"any"` for CSR signing use case
  - [ ] 4.2: Confirm all other existing Enums are complete (no-op if complete)
- [ ] Task 5: Pattern marker audit (AC: 5)
  - [ ] 5.1: Confirm no safe Pattern additions needed beyond existing `Name` DNS-label and `ConnectionURI` patterns
- [ ] Task 6: Run `make manifests generate fmt vet test` (AC: 5)
- [ ] Task 7: Run `make integration` — all integration tests must pass

## Dev Notes

### Ordering Dependency

This story MUST be implemented AFTER Stories 7.5.1–7.5.6 are complete. All 12 files modified by this story are also modified by earlier stories (R1/R2 annotation changes). The changes here are purely additive (new marker lines) and non-conflicting, but the file state must reflect the completed R1/R2 work.

### Scope: 12 Files, ~30 Marker Additions (8 Enum + ~22 Minimum/Maximum)

| File | Enum Additions | Min/Max Additions | Total |
|------|----------------|-------------------|-------|
| `api/v1alpha1/certauthenginerole_types.go` | 1 (TokenType) | 2 (OCSPMaxRetries, TokenNumUses) | 3 |
| `api/v1alpha1/certauthengineconfig_types.go` | 0 | 2 (OCSPCacheSize, RoleCacheSize) | 2 |
| `api/v1alpha1/azureauthenginerole_types.go` | 1 (TokenType) | 2 (TokenNumUses, TokenPeriod) | 3 |
| `api/v1alpha1/gcpauthenginerole_types.go` | 2 (Type, TokenType) | 2 (TokenNumUses, TokenPeriod) | 4 |
| `api/v1alpha1/authenginemount_types.go` | 1 (TokenType) | 0 | 1 |
| `api/v1alpha1/jwtoidcauthenginerole_types.go` | 1 (RoleType) | 3 (TokenNumUses, TokenPeriod, MaxAge) | 4 |
| `api/v1alpha1/pkisecretengineconfig_types.go` | 1 (PrivateKeyFormat) | 1 (MaxPathLength) | 2 |
| `api/v1alpha1/azuresecretenginerole_types.go` | 1 (SignInAudience) | 0 | 1 |
| `api/v1alpha1/vaultsecret_types.go` | 0 | 2 (RefreshThreshold Min+Max) | 2 |
| `api/v1alpha1/githubsecretengineconfig_types.go` | 0 | 1 (ApplicationID) | 1 |
| `api/v1alpha1/kubernetesauthenginerole_types.go` | 0 | 5 (token tuning fields) | 5 |
| `api/v1alpha1/ldapauthengineconfig_types.go` | 0 | 2 (TokenNumUses, TokenPeriod) | 2 |
| **TOTAL** | **8** | **22** | **30** |

### Detailed Field Change Tables

#### Task 1 — TokenType Enum (4 fields, same value set)

All `TokenType` fields accept the same Vault token-type tuning values. `VRole.TokenType` in `kubernetesauthenginerole_types.go` already has this Enum; `LDAPConfig.TokenType` and `JWTOIDCRole.TokenType` are being added by Stories 7.5.1 and 7.5.2 respectively. The 4 below are the remaining gaps.

| File | Struct | Field | Go Type | Current Markers | Enum to Add |
|------|--------|-------|---------|-----------------|-------------|
| `certauthenginerole_types.go` | `CertAuthEngineRoleInternal` | `TokenType` | `string` | Optional | `Enum={"service","batch","default","default-service","default-batch"}` |
| `azureauthenginerole_types.go` | `AzureRole` | `TokenType` | `string` | Optional | same |
| `gcpauthenginerole_types.go` | `GCPRole` | `TokenType` | `string` | Optional | same |
| `authenginemount_types.go` | `AuthMountConfig` | `TokenType` | `string` | Optional | same |

All four fields have `omitempty` on their JSON tags. When omitted by the user, the field is absent from the request and the Enum is not triggered. When explicitly set, only valid values are accepted.

#### Task 2 — Other Enum Markers (4 fields, distinct value sets)

**`GCPRole.Type` (gcpauthenginerole_types.go):**

| Field | Struct | Go Type | Current Markers | JSON Tag | Enum to Add |
|-------|--------|---------|-----------------|----------|-------------|
| `Type` | `GCPRole` | `string` | Required | `json:"type"` | `Enum={"iam","gce"}` |

The field is Required with no default. GCP auth roles are strictly either IAM or GCE type. The Vault API rejects other values.

**`JWTOIDCRole.RoleType` (jwtoidcauthenginerole_types.go):**

| Field | Struct | Go Type | Current Markers | JSON Tag | Enum to Add |
|-------|--------|---------|-----------------|----------|-------------|
| `RoleType` | `JWTOIDCRole` | `string` | Optional | `json:"roleType,omitempty"` | `Enum={"oidc","jwt"}` |

Comment: "Type of role, either 'oidc' (default) or 'jwt'". Optional with omitempty — when absent, Vault defaults to "oidc".

**`PKICommon.PrivateKeyFormat` (pkisecretengineconfig_types.go):**

| Field | Struct | Go Type | Current Markers | JSON Tag | Enum to Add |
|-------|--------|---------|-----------------|----------|-------------|
| `PrivateKeyFormat` | `PKICommon` | `string` | Optional | `json:"privateKeyFormat,omitempty"` | `Enum={"der","pkcs8"}` |

Comment: "The other option is pkcs8". Only two formats exist. When absent, Vault uses "der".

**`AzureSERole.SignInAudience` (azuresecretenginerole_types.go):**

| Field | Struct | Go Type | Current Markers | JSON Tag | Enum to Add |
|-------|--------|---------|-----------------|----------|-------------|
| `SignInAudience` | `AzureSERole` | `string` | Optional | `json:"signInAudience,omitempty"` | `Enum={"AzureADMyOrg","AzureADMultipleOrgs","AzureADandPersonalMicrosoftAccount","PersonalMicrosoftAccount"}` |

Comment lists exactly these 4 valid values. After Story 7.5.4 removes the `+kubebuilder:default=""`, field is Optional with omitempty — absent when not set, Enum validates only explicit values.

#### Task 3 — Minimum/Maximum Markers

**Cert Auth — Cache and Retry Bounds:**

| File | Struct | Field | Go Type | Default | Marker to Add | Rationale |
|------|--------|-------|---------|---------|---------------|-----------|
| `certauthengineconfig_types.go` | `CertAuthEngineConfigInternal` | `OCSPCacheSize` | `int` | `100` | `Minimum=0` | LRU cache size cannot be negative |
| `certauthengineconfig_types.go` | `CertAuthEngineConfigInternal` | `RoleCacheSize` | `int` | `200` | `Minimum=-1` | `-1` disables caching per Vault docs |
| `certauthenginerole_types.go` | `CertAuthEngineRoleInternal` | `OCSPMaxRetries` | `int64` | `4` | `Minimum=0` | `0` disables retries per Vault docs |
| `certauthenginerole_types.go` | `CertAuthEngineRoleInternal` | `TokenNumUses` | `int64` | `0` | `Minimum=0` | `0` means unlimited; negative invalid |

**PKI — Path Length:**

| File | Struct | Field | Go Type | Default | Marker to Add | Rationale |
|------|--------|-------|---------|---------|---------------|-----------|
| `pkisecretengineconfig_types.go` | `PKICommon` | `MaxPathLength` | `int` | `-1` | `Minimum=-1` | `-1` means no limit per Vault docs |

**VaultSecret — Percentage:**

| File | Struct | Field | Go Type | Default | Markers to Add | Rationale |
|------|--------|-------|---------|---------|----------------|-----------|
| `vaultsecret_types.go` | `VaultSecretSpec` | `RefreshThreshold` | `int` | `90` | `Minimum=1`, `Maximum=100` | Percentage of lease lifetime |

**GitHub — App ID:**

| File | Struct | Field | Go Type | Default | Marker to Add | Rationale |
|------|--------|-------|---------|---------|---------------|-----------|
| `githubsecretengineconfig_types.go` | `GHConfig` | `ApplicationID` | `int64` | (none, Required) | `Minimum=1` | GitHub App IDs are positive integers |

**Token Tuning — Non-negative bounds across auth types:**

All Vault token tuning numeric fields are non-negative (`0` = unlimited/unset for counts and periods).

| File | Struct | Fields | Go Type | Marker | Count |
|------|--------|--------|---------|--------|-------|
| `kubernetesauthenginerole_types.go` | `VRole` | `TokenTTL`, `TokenMaxTTL`, `TokenExplicitMaxTTL`, `TokenPeriod`, `TokenNumUses` | `int` | `Minimum=0` each | 5 |
| `gcpauthenginerole_types.go` | `GCPRole` | `TokenNumUses`, `TokenPeriod` | `int64` | `Minimum=0` each | 2 |
| `azureauthenginerole_types.go` | `AzureRole` | `TokenNumUses`, `TokenPeriod` | `int64` | `Minimum=0` each | 2 |
| `ldapauthengineconfig_types.go` | `LDAPConfig` | `TokenNumUses`, `TokenPeriod` | `int64` | `Minimum=0` each | 2 |
| `jwtoidcauthenginerole_types.go` | `JWTOIDCRole` | `TokenNumUses`, `TokenPeriod`, `MaxAge` | `int64` | `Minimum=0` each | 3 |

**NOT in scope:** `ClockSkewLeeway`, `ExpirationLeeway`, `NotBeforeLeeway` in `JWTOIDCRole` — Vault allows `-1` to disable leeway.

#### Task 4 — Existing Enum Audit

Pre-existing Enum markers confirmed complete (no changes needed):

| File | Field | Enum Values | Status |
|------|-------|-------------|--------|
| `group_types.go` | `Type` | `internal`, `external` | Complete |
| `authenginemount_types.go` | `ListingVisibility` | `unauth`, `hidden` | Complete |
| `secretenginemount_types.go` | `ListingVisibility` | `unauth`, `hidden` | Complete |
| `randomsecret_types.go` | `KvSecretRetainPolicy` | `Delete`, `Retain` | Complete |
| `identityoidcclient_types.go` | `ClientType` | `confidential`, `public` | Complete |
| `kubernetesauthenginerole_types.go` | `AliasNameSource` | `serviceaccount_uid`, `serviceaccount_name` | Complete |
| `kubernetesauthenginerole_types.go` | `TokenType` | `service`, `batch`, `default`, `default-service`, `default-batch` | Complete |
| `kubernetessecretenginerole_types.go` | `KubernetesRoleType` | `Role`, `ClusterRole` | Complete |
| `quaysecretenginerole_types.go` | `Permission` | `admin`, `read`, `write` | Complete |
| `quaysecretenginerole_types.go` | `TeamRole` | `admin`, `creator`, `member` | Complete |
| `quaysecretenginerole_types.go` | `NamespaceType` | `organization`, `user` | Complete |
| `databasesecretengineconfig_types.go` | `PasswordAuthentication` | `password`, `scram-sha-256` | Complete |
| `databasesecretenginestaticrole_types.go` | `CredentialType` | `password`, `rsa_private_key` | Complete |
| `databasesecretenginestaticrole_types.go` | `KeyBits` | `2048`, `3072`, `4096` | Complete |
| `databasesecretenginestaticrole_types.go` | `Format` | `pkcs8` | Complete |
| `pkisecretengineconfig_types.go` | `Type` | `root`, `intermediate` | Complete |
| `pkisecretengineconfig_types.go` | `PrivateKeyType` | `internal`, `exported` | Complete |
| `pkisecretengineconfig_types.go` | `Format` | `pem`, `pem_bundle`, `der` | Complete |
| `pkisecretengineconfig_types.go` | `KeyType` | `rsa`, `ec` | Complete |
| `pkisecretenginerole_types.go` | `KeyType` | `rsa`, `ec` | Review: consider adding `any` |
| `pkisecretenginerole_types.go` | `KeyUsage` | 9 values (DigitalSignature...) | Complete |
| `pkisecretenginerole_types.go` | `ExtKeyUsage` | 13 values (ServerAuth...) | Complete |
| `identitytokenkey_types.go` | `Algorithm` | `RS256`..`EdDSA` (7 values) | Complete |
| `vaultsecret_types.go` | `RequestType` | `GET`, `POST` | Complete |
| `audit_types.go` | `Type` | `file`, `socket`, `syslog` | Complete |
| `policy_types.go` | `Type` | `acl` | Complete |

**Decision point — `PKIRole.KeyType`:** The existing Enum `{"rsa","ec"}` matches the comment "currently, rsa and ec are supported", but the comment also says "or when signing CSRs any can be specified." If `"any"` is a valid value for this field, the Enum should be extended to `{"rsa","ec","any"}`. Verify against Vault PKI docs. If `"any"` is valid, add it; otherwise leave as-is.

#### Task 5 — Pattern Marker Audit Results

Existing Pattern markers are comprehensive:
- 30 CRD types have `Name` field with DNS-label pattern `[a-z0-9]([-a-z0-9]*[a-z0-9])?`
- `rabbitmqsecretengineconfig_types.go` has `ConnectionURI` with `^(http|https):\/\/.+$`
- `audit_types.go` has `Path` and `Type` with `^[a-zA-Z0-9/_-]+$`

No safe Pattern additions identified. Candidates considered and rejected:
- **Duration strings** (e.g., `TokenTTL` string fields): Vault accepts multiple formats (Go duration, integer seconds); a regex would be overly restrictive
- **URL fields** (`OIDCDiscoveryURL`, `JWKSURL`, `Issuer`): formats vary (IPv6, ports, paths); brittle regex would break valid deployments
- **`KubernetesHost`**: while typically `https://`, lab clusters may use HTTP; too restrictive

### Deliberately Out of Scope

1. **`IAMalias` in `gcpauthengineconfig_types.go`** — Comment says "must be unique_id or role_id" but default is `"default"`. Pre-existing discrepancy; adding Enum would break the existing default. Story 7.5.4 explicitly documents this as out of scope.
2. **`IAMmetadata` / `GCEmetadata`** — Free-form comma-separated field lists plus `"default"` sentinel. Not Enum candidates.
3. **`OIDCResponseTypes` (`[]string`) and `JWTSupportedAlgs` (`[]string`)** — Per-item validation on string slices requires different approach (item-level Enum via OpenAPI). Low value for the complexity.
4. **`PasswordPolicyRule.RuleType`** — Not a standard JSON field (uses `hcl:"type,label"` tag). Not a CRD admission validation target.
5. **`AuthMount.Type` / `Mount.Type`** — Open set of Vault plugin/mount types. Not Enum candidates.
6. **`TemplatizedK8sSecret.Type`** — Kubernetes Secret types include custom types; Enum would reject valid custom values.
7. **`KeyBits` in PKI** — Valid values depend on `KeyType` (rsa vs ec); a single Minimum/Maximum would reject valid combinations. Requires CEL or per-key-type validation.
8. **`ClockSkewLeeway`, `ExpirationLeeway`, `NotBeforeLeeway`** in JWT/OIDC — Vault allows `-1` to disable leeway. `Minimum=0` would be wrong.

### Impact on `toMap()` and `IsEquivalentToDesiredState()`

All changes are **marker-only** — they add kubebuilder validation annotations. No Go code logic, JSON tags, or struct field changes.

### Impact on Existing Tests

**Unit tests (all unaffected — use explicit valid field values, not relying on validation):**

Unit tests construct struct values directly in Go code. Kubebuilder validation markers are enforced by the API server at admission time, not at Go runtime. All existing unit tests will continue to pass.

**Integration tests (all should pass — fixtures use valid values):**

Integration test YAML fixtures set field values explicitly. None use values that would violate the new Enum or Minimum/Maximum constraints. Specific checks:

- **TokenType:** No integration fixture sets `tokenType` to an invalid value. Fixtures that omit it rely on omitempty (absent = no validation).
- **RefreshThreshold:** VaultSecret fixtures use `90` (default) or explicit valid percentages.
- **ApplicationID:** GitHub fixtures set valid positive app IDs.
- **Token tuning int fields:** No fixture uses negative values.

**No integration tests (cloud/external dependencies — unit tests only):**

- CertAuthEngine (no infrastructure mock in Kind)
- AzureAuthEngine, AzureSecretEngine (Azure cloud)
- GCPAuthEngine (GCP cloud)
- GitHubSecretEngine (GitHub App)

### Pattern for Adding Enum Marker

Add the Enum line directly before the field, after any existing Optional/default markers:

```go
// For machine based authentication cases, you should use batch type tokens.
// +kubebuilder:validation:Optional
// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
TokenType string `json:"tokenType,omitempty"`
```

### Pattern for Adding Minimum/Maximum Marker

Add the Minimum (and optionally Maximum) line before the field:

```go
// RefreshThreshold ...
// +kubebuilder:validation:Required
// +kubebuilder:default=90
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
RefreshThreshold int `json:"refreshThreshold,omitempty"`
```

For non-negative fields:

```go
// The maximum number of times a generated token may be used ...
// +kubebuilder:validation:Optional
// +kubebuilder:validation:Minimum=0
TokenNumUses int64 `json:"tokenNumUses,omitempty"`
```

### Critical Warnings

1. **Do NOT modify `toMap()`, `IsEquivalentToDesiredState()`, or any Go code logic.** This is purely a marker addition.
2. **Run `make manifests generate` after changes.** CRD schemas will gain `enum:` and `minimum:`/`maximum:` entries in the OpenAPI spec.
3. **Run `make generate`** for completeness (struct tag changes are not expected, but verify).
4. **This story depends on Stories 7.5.1–7.5.6 being complete.** Those stories modify the same files (R1/R2 changes). Implement this story only after all prior stories in Epic 7.5 are done.
5. **Do NOT add Enum to `IAMalias` in `gcpauthengineconfig_types.go`.** The `"default"` default doesn't match the comment's listed values. See Story 7.5.4 notes.
6. **Do NOT add `Minimum=0` to JWT/OIDC leeway fields** (`ClockSkewLeeway`, `ExpirationLeeway`, `NotBeforeLeeway`). Vault uses `-1` to disable leeway.
7. **All `TokenType` fields have `omitempty`.** The Enum only validates when a user explicitly provides a value. Absent fields are not validated.
8. **`SignInAudience` Enum uses exact Azure AD audience names** with mixed case. These are case-sensitive in the Vault API.
9. **`GCPRole.Type` is Required with no default.** The Enum `{"iam","gce"}` catches invalid values at admission instead of at Vault API call time.
10. **`PKIRole.KeyType` existing Enum may need `"any"` added.** Check Vault PKI docs — if "any" is a valid value for CSR signing, extend the Enum from `{"rsa","ec"}` to `{"rsa","ec","any"}`. If unsure, leave as-is (conservative approach).

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/certauthenginerole_types.go` | Modified | Add Enum to TokenType, Minimum=0 to OCSPMaxRetries + TokenNumUses |
| 2 | `api/v1alpha1/certauthengineconfig_types.go` | Modified | Add Minimum=0 to OCSPCacheSize, Minimum=-1 to RoleCacheSize |
| 3 | `api/v1alpha1/azureauthenginerole_types.go` | Modified | Add Enum to TokenType, Minimum=0 to TokenNumUses + TokenPeriod |
| 4 | `api/v1alpha1/gcpauthenginerole_types.go` | Modified | Add Enum to Type + TokenType, Minimum=0 to TokenNumUses + TokenPeriod |
| 5 | `api/v1alpha1/authenginemount_types.go` | Modified | Add Enum to TokenType |
| 6 | `api/v1alpha1/jwtoidcauthenginerole_types.go` | Modified | Add Enum to RoleType, Minimum=0 to TokenNumUses + TokenPeriod + MaxAge |
| 7 | `api/v1alpha1/pkisecretengineconfig_types.go` | Modified | Add Enum to PrivateKeyFormat, Minimum=-1 to MaxPathLength |
| 8 | `api/v1alpha1/azuresecretenginerole_types.go` | Modified | Add Enum to SignInAudience |
| 9 | `api/v1alpha1/vaultsecret_types.go` | Modified | Add Minimum=1 + Maximum=100 to RefreshThreshold |
| 10 | `api/v1alpha1/githubsecretengineconfig_types.go` | Modified | Add Minimum=1 to ApplicationID |
| 11 | `api/v1alpha1/kubernetesauthenginerole_types.go` | Modified | Add Minimum=0 to 5 token tuning int fields |
| 12 | `api/v1alpha1/ldapauthengineconfig_types.go` | Modified | Add Minimum=0 to TokenNumUses + TokenPeriod |
| 13+ | `config/crd/bases/*.yaml` | Regenerated | CRD schemas updated by `make manifests` |

### Project Structure Notes

- All CRD types live in `api/v1alpha1/`
- CRD schemas are regenerated into `config/crd/bases/` by `make manifests`
- DeepCopy is regenerated into `api/v1alpha1/zz_generated.deepcopy.go` by `make generate`
- Unit test files: `api/v1alpha1/*_test.go` — verify they pass after changes
- Integration test files: `controllers/*_controller_test.go` — verify with `make integration`

### References

- [Source: _bmad-output/project-context.md#CRD Field Default & Validation Rules] — Rules 3-5 governing Enum, Minimum/Maximum, Pattern markers
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.5.7] — Epic story definition and acceptance criteria
- [Source: api/v1alpha1/certauthenginerole_types.go:210-215] — TokenType field needing Enum
- [Source: api/v1alpha1/azureauthenginerole_types.go:176-181] — TokenType field needing Enum
- [Source: api/v1alpha1/gcpauthenginerole_types.go:82-84] — Type field (iam/gce) needing Enum
- [Source: api/v1alpha1/gcpauthenginerole_types.go:158-162] — TokenType field needing Enum
- [Source: api/v1alpha1/authenginemount_types.go:134-136] — TokenType field needing Enum
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go:55-58] — RoleType field needing Enum
- [Source: api/v1alpha1/pkisecretengineconfig_types.go:109-111] — PrivateKeyFormat field needing Enum
- [Source: api/v1alpha1/azuresecretenginerole_types.go:125-129] — SignInAudience field needing Enum
- [Source: api/v1alpha1/vaultsecret_types.go:35-40] — RefreshThreshold needing Minimum+Maximum
- [Source: api/v1alpha1/githubsecretengineconfig_types.go:63-65] — ApplicationID needing Minimum
- [Source: api/v1alpha1/certauthengineconfig_types.go:95-103] — OCSPCacheSize and RoleCacheSize needing Minimum
- [Source: api/v1alpha1/certauthenginerole_types.go:158-161] — OCSPMaxRetries needing Minimum
- [Source: api/v1alpha1/pkisecretengineconfig_types.go:124-127] — MaxPathLength needing Minimum

### Previous Story Intelligence

**From Story 7.5.1 (LDAP Auth Engine Types — Annotation Refactor):**
- Added Enum to TLSMinVersion/TLSMaxVersion: `{"tls10","tls11","tls12","tls13"}`
- Added Enum to LDAPConfig.TokenType: `{"service","batch","default","default-service","default-batch"}`
- Confirmed: Enum markers on Optional+omitempty fields validate only when user explicitly provides a value

**From Story 7.5.2 (JWT/OIDC Auth Engine Types — Annotation Refactor):**
- Added Enum to OIDCResponseMode: `{"query","form_post"}`
- Added Enum to BoundClaimsType: `{"string","glob"}`
- Added Enum to JWTOIDCRole.TokenType: same 5 values as LDAP
- Did NOT add Enum to RoleType (left for this story)

**From Story 7.5.3 (Kubernetes Auth & Secret Engine Types):**
- Confirmed VRole.TokenType, AliasNameSource, KubernetesRoleType Enums already present — no new Enums added

**From Story 7.5.4 (Azure & GCP Auth/Secret Engine Types):**
- Added Enum to Environment (2 Azure configs): `{"AzurePublicCloud","AzureUSGovernmentCloud","AzureChinaCloud","AzureGermanCloud"}`
- Added Enum to GCEalias: `{"instance_id","role_id"}`
- Explicitly documented IAMalias as out of scope for Enum (default/comment discrepancy)
- Did NOT add Enum to AzureRole.TokenType, GCPRole.TokenType, GCPRole.Type, or SignInAudience

**From Story 7.5.5 (PKI Secret Engine Types):**
- Confirmed all PKI Enum markers already present — no new Enums needed
- KeyBits intentionally not Enum'd (depends on KeyType)

**From Story 7.5.6 (Identity & Remaining Types):**
- No Enum markers added (deferred to this story)
- Confirmed: this story is the dedicated validation marker sweep

**From Epic 7 Retrospective:**
- Codebase is stable after hardening epic
- Comprehensive test coverage exists across all modified types

### Git Intelligence

- Latest commit: `44cad20` (Bmad epic 7 squash merge)
- Branch is clean on main
- No pending changes that could conflict
- All CI checks passing on main

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
