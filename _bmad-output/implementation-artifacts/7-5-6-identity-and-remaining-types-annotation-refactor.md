# Story 7.5.6: Identity & Remaining Types — Annotation Refactor

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want all remaining CRD types to follow the CRD Field Default & Validation Rules,
So that the entire codebase is consistent.

## Acceptance Criteria

1. **Given** all remaining zero-value `+kubebuilder:default` markers (bool `false`, int `0`, Duration `"0s"`) **When** markers removed **Then** fields rely on Go zero values and `omitempty` keeps YAML clean
2. **Given** all remaining non-zero `+kubebuilder:default` fields with `omitempty` **When** `omitempty` removed from JSON tag **Then** fields always serialized in JSON output
3. **Given** all changes **When** `make manifests generate fmt vet test` passes **Then** no regressions
4. **Given** all changes **When** `make integration` is run **Then** all integration tests pass
5. **Given** all `*_types.go` files **When** audited post-refactor **Then** zero Rule 1 or Rule 2 violations remain across the entire codebase

## Tasks / Subtasks

- [ ] Task 1: Refactor identity types — 5 files, 10 fields (AC: 1, 2)
  - [ ] 1.1: `identityoidcclient_types.go` — R2: remove `omitempty` from `Key`, `ClientType`, `IDTokenTTL`, `AccessTokenTTL`
  - [ ] 1.2: `identitytokenkey_types.go` — R2: remove `omitempty` from `RotationPeriod`, `VerificationTTL`, `Algorithm`
  - [ ] 1.3: `identitytokenrole_types.go` — R2: remove `omitempty` from `TTL`
  - [ ] 1.4: `group_types.go` — R2: remove `omitempty` from `Type`
  - [ ] 1.5: `entity_types.go` — R1: remove `+kubebuilder:default:=false` from `Disabled`
- [ ] Task 2: Refactor cert auth engine types — 2 files, 10 fields (AC: 1, 2)
  - [ ] 2.1: `certauthengineconfig_types.go` — R1: remove `+kubebuilder:default:=false` from `DisableBinding`, `EnableIdentityAliasMetadata`; R2: remove `omitempty` from `OCSPCacheSize`, `RoleCacheSize`
  - [ ] 2.2: `certauthenginerole_types.go` — R1: remove `+kubebuilder:default:=false` from `OCSPEnabled`, `OCSPFailOpen`, `OCSPQueryAllServers`, `TokenNoDefaultPolicy`; remove `+kubebuilder:default:=0` from `TokenNumUses`; R2: remove `omitempty` from `OCSPMaxRetries`
- [ ] Task 3: Refactor database engine types — 2 files, 4 fields (AC: 1, 2)
  - [ ] 3.1: `databasesecretengineconfig_types.go` — R2: remove `omitempty` from `AllowedRoles`, `PasswordAuthentication`
  - [ ] 3.2: `databasesecretenginerole_types.go` — R1: remove `+kubebuilder:default="0s"` from `DefaultTTL`, `MaxTTL`
- [ ] Task 4: Refactor mount types — 2 files, 6 fields (AC: 1, 2)
  - [ ] 4.1: `secretenginemount_types.go` — R1: remove `+kubebuilder:default:=false` from `Local`, `SealWrap`, `ExternalEntropyAccess`; R1 special: `ForceNoCache` — remove `+kubebuilder:default:=false` AND add `omitempty` to JSON tag; R2: remove `omitempty` from `ListingVisibility`
  - [ ] 4.2: `authenginemount_types.go` — R2: remove `omitempty` from `ListingVisibility`
- [ ] Task 5: Refactor remaining types — 5 files, 10 fields (AC: 1, 2)
  - [ ] 5.1: `githubsecretengineconfig_types.go` — R2: remove `omitempty` from `GitHubAPIBaseURL`
  - [ ] 5.2: `quaysecretengineconfig_types.go` — R1: remove `+kubebuilder:default=false` from `DisableSslVerification`
  - [ ] 5.3: `quaysecretenginerole_types.go` — R2: remove `omitempty` from `NamespaceType`
  - [ ] 5.4: `randomsecret_types.go` — R1: remove `+kubebuilder:default=false` from `IsKVSecretsEngineV2`; R2: remove `omitempty` from `KvSecretRetainPolicy`
  - [ ] 5.5: `vaultsecret_types.go` — R1: remove `+kubebuilder:default=false` from `SyncOnResourceChange`; R2: remove `omitempty` from `RefreshThreshold`, `Path`, `RequestType`
- [ ] Task 6: Run `make manifests generate fmt vet test` (AC: 3)
- [ ] Task 7: Run `make integration` — all integration tests must pass (AC: 4)
- [ ] Task 8: Final audit — grep all `*_types.go` for remaining R1/R2 violations; confirm zero (AC: 5)

## Dev Notes

### Scope: 17 Files, 40 Field Changes (18 R1 + 22 R2)

This is the largest story in Epic 7.5, covering all files not addressed by Stories 7.5.1-7.5.5.

| File | R1 Removals | R2 Fixes | Total |
|------|-------------|----------|-------|
| `api/v1alpha1/identityoidcclient_types.go` | 0 | 4 | 4 |
| `api/v1alpha1/identitytokenkey_types.go` | 0 | 3 | 3 |
| `api/v1alpha1/identitytokenrole_types.go` | 0 | 1 | 1 |
| `api/v1alpha1/group_types.go` | 0 | 1 | 1 |
| `api/v1alpha1/entity_types.go` | 1 | 0 | 1 |
| `api/v1alpha1/certauthengineconfig_types.go` | 2 | 2 | 4 |
| `api/v1alpha1/certauthenginerole_types.go` | 5 | 1 | 6 |
| `api/v1alpha1/databasesecretengineconfig_types.go` | 0 | 2 | 2 |
| `api/v1alpha1/databasesecretenginerole_types.go` | 2 | 0 | 2 |
| `api/v1alpha1/secretenginemount_types.go` | 4 | 1 | 5 |
| `api/v1alpha1/authenginemount_types.go` | 0 | 1 | 1 |
| `api/v1alpha1/githubsecretengineconfig_types.go` | 0 | 1 | 1 |
| `api/v1alpha1/quaysecretengineconfig_types.go` | 1 | 0 | 1 |
| `api/v1alpha1/quaysecretenginerole_types.go` | 0 | 1 | 1 |
| `api/v1alpha1/randomsecret_types.go` | 1 | 1 | 2 |
| `api/v1alpha1/vaultsecret_types.go` | 1 | 3 | 4 |
| `api/v1alpha1/auditrequestheader_types.go` | 1 | 0 | 1 |
| **TOTAL** | **18** | **22** | **40** |

### Detailed Field Change Tables

#### Task 1 — Identity Types (R2 dominant)

**`identityoidcclient_types.go` (4 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `Key` | `IdentityOIDCClientConfig` | `string` | `"default"` | `json:"key,omitempty"` | Remove `omitempty` |
| `ClientType` | `IdentityOIDCClientConfig` | `string` | `"confidential"` | `json:"clientType,omitempty"` | Remove `omitempty` |
| `IDTokenTTL` | `IdentityOIDCClientConfig` | `string` | `"24h"` | `json:"idTokenTTL,omitempty"` | Remove `omitempty` |
| `AccessTokenTTL` | `IdentityOIDCClientConfig` | `string` | `"24h"` | `json:"accessTokenTTL,omitempty"` | Remove `omitempty` |

**`identitytokenkey_types.go` (3 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `RotationPeriod` | `IdentityTokenKeyConfig` | `string` | `"24h"` | `json:"rotationPeriod,omitempty"` | Remove `omitempty` |
| `VerificationTTL` | `IdentityTokenKeyConfig` | `string` | `"24h"` | `json:"verificationTTL,omitempty"` | Remove `omitempty` |
| `Algorithm` | `IdentityTokenKeyConfig` | `string` | `"RS256"` | `json:"algorithm,omitempty"` | Remove `omitempty` |

**`identitytokenrole_types.go` (1 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `TTL` | `IdentityTokenRoleConfig` | `string` | `"24h"` | `json:"ttl,omitempty"` | Remove `omitempty` |

**`group_types.go` (1 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `Type` | `GroupConfig` | `string` | `"internal"` | `json:"type,omitempty"` | Remove `omitempty` |

**`entity_types.go` (1 R1):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `Disabled` | `EntityConfig` | `bool` | `false` | `json:"disabled,omitempty"` | Remove `+kubebuilder:default:=false` marker only |

#### Task 2 — Cert Auth Engine Types (R1 heavy)

**`certauthengineconfig_types.go` (2 R1 + 2 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `DisableBinding` | `CertAuthEngineConfigInternal` | `bool` | `false` | `json:"disableBinding,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `EnableIdentityAliasMetadata` | `CertAuthEngineConfigInternal` | `bool` | `false` | `json:"enableIdentityAliasMetadata,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `OCSPCacheSize` | `CertAuthEngineConfigInternal` | `int` | `100` | `json:"ocspCacheSize,omitempty"` | R2: remove `omitempty` |
| `RoleCacheSize` | `CertAuthEngineConfigInternal` | `int` | `200` | `json:"roleCacheSize,omitempty"` | R2: remove `omitempty` |

**`certauthenginerole_types.go` (5 R1 + 1 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `OCSPEnabled` | `CertAuthEngineRoleInternal` | `bool` | `false` | `json:"ocspEnabled,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `OCSPFailOpen` | `CertAuthEngineRoleInternal` | `bool` | `false` | `json:"ocspFailOpen,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `OCSPMaxRetries` | `CertAuthEngineRoleInternal` | `int64` | `4` | `json:"ocspMaxRetries,omitempty"` | R2: remove `omitempty` |
| `OCSPQueryAllServers` | `CertAuthEngineRoleInternal` | `bool` | `false` | `json:"ocspQueryAllServers,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `TokenNoDefaultPolicy` | `CertAuthEngineRoleInternal` | `bool` | `false` | `json:"tokenNoDefaultPolicy,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `TokenNumUses` | `CertAuthEngineRoleInternal` | `int64` | `0` | `json:"tokenNumUses,omitempty"` | R1: remove `+kubebuilder:default:=0` |

#### Task 3 — Database Engine Types

**`databasesecretengineconfig_types.go` (2 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `AllowedRoles` | `DBSEConfig` | `[]string` | `{"*"}` | `json:"allowedRoles,omitempty"` | R2: remove `omitempty` |
| `PasswordAuthentication` | `DBSEConfig` | `string` | `"password"` | `json:"passwordAuthentication,omitempty"` | R2: remove `omitempty` |

**`databasesecretenginerole_types.go` (2 R1):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `DefaultTTL` | `DBSERole` | `metav1.Duration` | `"0s"` | `json:"defaultTTL,omitempty"` | R1: remove `+kubebuilder:default="0s"` |
| `MaxTTL` | `DBSERole` | `metav1.Duration` | `"0s"` | `json:"maxTTL,omitempty"` | R1: remove `+kubebuilder:default="0s"` |

#### Task 4 — Mount Types

**`secretenginemount_types.go` (4 R1 + 1 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `Local` | `Mount` | `bool` | `false` | `json:"local,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `SealWrap` | `Mount` | `bool` | `false` | `json:"sealWrap,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `ExternalEntropyAccess` | `Mount` | `bool` | `false` | `json:"externalEntropyAccess,omitempty"` | R1: remove `+kubebuilder:default:=false` |
| `ForceNoCache` | `MountConfig` | `bool` | `false` | `json:"forceNoCache"` | R1 special: remove `+kubebuilder:default:=false` AND add `omitempty` → `json:"forceNoCache,omitempty"` |
| `ListingVisibility` | `MountConfig` | `string` | `"hidden"` | `json:"listingVisibility,omitempty"` | R2: remove `omitempty` |

**`authenginemount_types.go` (1 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `ListingVisibility` | `AuthMountConfig` | `string` | `"hidden"` | `json:"listingVisibility,omitempty"` | R2: remove `omitempty` |

#### Task 5 — Remaining Types

**`githubsecretengineconfig_types.go` (1 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `GitHubAPIBaseURL` | `GHConfig` | `string` | `"https://api.github.com"` | `json:"gitHubAPIBaseURL,omitempty"` | R2: remove `omitempty` |

**`quaysecretengineconfig_types.go` (1 R1):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `DisableSslVerification` | `QuayConfig` | `bool` | `false` | `json:"disableSslVerification,omitempty"` | R1: remove `+kubebuilder:default=false` |

**`quaysecretenginerole_types.go` (1 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `NamespaceType` | `QuayBaseRole` | `NamespaceType` | `"organization"` | `json:"namespaceType,omitempty"` | R2: remove `omitempty` |

**`randomsecret_types.go` (1 R1 + 1 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `IsKVSecretsEngineV2` | `RandomSecretSpec` | `bool` | `false` | `json:"isKVSecretsEngineV2,omitempty"` | R1: remove `+kubebuilder:default=false` |
| `KvSecretRetainPolicy` | `RandomSecretSpec` | `string` | `"Delete"` | `json:"kvSecretRetainPolicy,omitempty"` | R2: remove `omitempty` |

**`vaultsecret_types.go` (1 R1 + 3 R2):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `RefreshThreshold` | `VaultSecretSpec` | `int` | `90` | `json:"refreshThreshold,omitempty"` | R2: remove `omitempty` |
| `SyncOnResourceChange` | `VaultSecretSpec` | `bool` | `false` | `json:"syncOnResourceChange,omitempty"` | R1: remove `+kubebuilder:default=false` |
| `Path` | `VaultSecretDefinition` | `vaultutils.Path` | `kubernetes` | `json:"path,omitempty"` | R2: remove `omitempty` |
| `RequestType` | `VaultSecretDefinition` | `string` | `GET` | `json:"requestType,omitempty"` | R2: remove `omitempty` |

**`auditrequestheader_types.go` (1 R1):**

| Field | Struct | Go Type | Default | Current JSON Tag | Change |
|-------|--------|---------|---------|------------------|--------|
| `HMAC` | `AuditRequestHeaderSpec` | `bool` | `false` | `json:"hmac,omitempty"` | R1: remove `+kubebuilder:default=false` |

### ForceNoCache Special Case

`ForceNoCache` in `secretenginemount_types.go` is the **only field** in this story that requires BOTH changes:
1. Remove `+kubebuilder:default:=false` (marker is redundant — Go zero is `false`)
2. Add `omitempty` to JSON tag: `json:"forceNoCache"` → `json:"forceNoCache,omitempty"`

This field currently has no `omitempty` AND has a redundant `false` default. Both are wrong per the rules. All other R1 fields already have `omitempty` — they just need the marker removed.

### Out of Scope: `CreateRepositories *bool` in quaysecretenginerole_types.go

`CreateRepositories *bool` has `+kubebuilder:default=false`. For pointer types, Go zero is `nil`, so `false` is technically non-zero. However, the pointer semantics (nil vs false) are intentional for distinguishing "unset" from "explicitly false". This field is **not in scope** for this story.

### Impact on `toMap()` and `IsEquivalentToDesiredState()`

All changes are **annotation-only** — they modify kubebuilder markers and JSON struct tags. They do NOT change any `toMap()`, `IsEquivalentToDesiredState()`, `GetPayload()`, or other Go code logic.

Files with these methods that are being modified:
- `identityoidcclient_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `identitytokenkey_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `identitytokenrole_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `group_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `entity_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `certauthengineconfig_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `certauthenginerole_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `databasesecretengineconfig_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `databasesecretenginerole_types.go`: `toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `secretenginemount_types.go`: `Mount.toMap()`, `MountConfig.toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `authenginemount_types.go`: `AuthMount.toMap()`, `AuthMountConfig.toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `githubsecretengineconfig_types.go`: `GHConfig.toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `quaysecretengineconfig_types.go`: `QuayConfig.toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `quaysecretenginerole_types.go`: `QuayRole.toMap()`, `IsEquivalentToDesiredState()` — unchanged
- `randomsecret_types.go`: `GetPayload()` (no `toMap`), `IsEquivalentToDesiredState()` (always returns `false`) — unchanged
- `vaultsecret_types.go`: no `toMap`/`IsEquivalentToDesiredState` — unchanged
- `auditrequestheader_types.go`: `GetPayload()`, `IsEquivalentToDesiredState()` — unchanged

### Impact on Existing Tests

**Unit tests (all unaffected — use explicit field values, not relying on defaults):**
- `api/v1alpha1/identityoidcclient_test.go` (if exists) or `identityoidc_test.go`
- `api/v1alpha1/identitytokenkey_test.go` / `identitytoken_test.go`
- `api/v1alpha1/identitytokenrole_test.go` / `identitytoken_test.go`
- `api/v1alpha1/group_test.go`
- `api/v1alpha1/entity_test.go`
- `api/v1alpha1/certauthengineconfig_test.go`
- `api/v1alpha1/certauthenginerole_test.go`
- `api/v1alpha1/databasesecretengineconfig_test.go`
- `api/v1alpha1/databasesecretenginerole_test.go`
- `api/v1alpha1/secretenginemount_test.go`
- `api/v1alpha1/authenginemount_test.go`
- `api/v1alpha1/githubsecretengineconfig_test.go`
- `api/v1alpha1/quaysecretengineconfig_test.go`
- `api/v1alpha1/quaysecretenginerole_test.go`
- `api/v1alpha1/randomsecret_test.go`
- `api/v1alpha1/auditrequestheader_test.go`
- `api/v1alpha1/isequivalent_audit_test.go` (broad equivalence tests)

**Integration tests (all should pass — fixtures use explicit values):**
- Entity, Group, GroupAlias: `controllers/entity_controller_test.go`, `controllers/group_controller_test.go`
- Identity OIDC: `controllers/identityoidc_controller_test.go`
- Identity Token: `controllers/identitytoken_controller_test.go`
- Database: `controllers/databasesecretengine_controller_test.go`
- SecretEngineMount/AuthEngineMount: tested via dependent controllers
- RandomSecret: `controllers/randomsecret_controller_test.go`
- VaultSecret: `controllers/vaultsecret_controller_test.go`, `controllers/vaultsecret_controller_v2_test.go`
- AuditRequestHeader: `controllers/audit_controller_test.go`

**No integration tests (cloud/external dependencies — unit tests only):**
- CertAuthEngine (no infrastructure mock in Kind)
- GitHubSecretEngine (GitHub App, cloud dependency)
- QuaySecretEngine (Quay registry, cloud dependency)

### Pattern for Non-Zero Default Fields (R2) — Remove `omitempty`

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default:="24h"
IDTokenTTL string `json:"idTokenTTL,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default:="24h"
IDTokenTTL string `json:"idTokenTTL"`
```

### Pattern for Zero-Value Bool Fields (R1) — Remove Default Only

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default:=false
DisableBinding bool `json:"disableBinding,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
DisableBinding bool `json:"disableBinding,omitempty"`
```

### Pattern for Zero-Value Duration Fields (R1) — Remove Default Only

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="0s"
DefaultTTL metav1.Duration `json:"defaultTTL,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
DefaultTTL metav1.Duration `json:"defaultTTL,omitempty"`
```

### Pattern for ForceNoCache Special Case (R1) — Remove Default AND Add omitempty

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default:=false
ForceNoCache bool `json:"forceNoCache"`
```

After:
```go
// +kubebuilder:validation:Optional
ForceNoCache bool `json:"forceNoCache,omitempty"`
```

### Final Audit (Task 8)

After all changes, grep the entire `api/v1alpha1/` directory for remaining violations:

1. **R1 check:** Find any `+kubebuilder:default` where value equals Go zero (`false`, `0`, `""`, `"0s"`). Exclude `*bool` pointer types (intentional non-zero semantics).
2. **R2 check:** Find any field with non-zero `+kubebuilder:default` AND `omitempty` in its JSON tag.

Both greps should return zero matches (excluding files covered by Stories 7.5.1-7.5.5 that were already fixed).

### Critical Warnings

1. **Do NOT modify `toMap()` or any Go logic.** This is purely an annotation + JSON tag refactor.
2. **Run `make manifests generate` after changes.** This regenerates CRDs in `config/crd/bases/`. The diff will show removed `default:` entries (R1 fields) and changed serialization behavior (R2 fields) in the OpenAPI schema.
3. **Run `make generate`** to regenerate `zz_generated.deepcopy.go` (struct tag changes may affect generated code).
4. **R2 fields retain their `+kubebuilder:default` markers.** Only `omitempty` is removed from the JSON tag. The default value annotation stays.
5. **R1 fields retain `omitempty` on JSON tag** (except `ForceNoCache` which gains it). Only the `+kubebuilder:default` marker line is removed.
6. **Do NOT change `CreateRepositories *bool` in `quaysecretenginerole_types.go`.** This pointer type has intentional nil vs false semantics.
7. **`VaultSecretDefinition.Path` has both `+kubebuilder:validation:Required` and `+kubebuilder:default=kubernetes`.** This is an existing pattern (default on required field). Only remove `omitempty` — do not change the Required/default markers.
8. **`AllowedRoles` in `databasesecretengineconfig_types.go` has a slice default `={"*"}`**. This is a valid non-zero default. Remove `omitempty` only.
9. **`certauthenginerole_types.go` has 6 fields** but spans multiple subsections of `CertAuthEngineRoleInternal` — verify you find all of them.
10. **`secretenginemount_types.go` has fields in TWO structs:** `Mount` (3 R1 fields) and `MountConfig` (1 R1 + 1 R2). Don't miss `MountConfig.ForceNoCache` and `MountConfig.ListingVisibility`.

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/identityoidcclient_types.go` | Modified | Remove `omitempty` from 4 R2 fields |
| 2 | `api/v1alpha1/identitytokenkey_types.go` | Modified | Remove `omitempty` from 3 R2 fields |
| 3 | `api/v1alpha1/identitytokenrole_types.go` | Modified | Remove `omitempty` from 1 R2 field |
| 4 | `api/v1alpha1/group_types.go` | Modified | Remove `omitempty` from 1 R2 field |
| 5 | `api/v1alpha1/entity_types.go` | Modified | Remove 1 R1 default marker |
| 6 | `api/v1alpha1/certauthengineconfig_types.go` | Modified | Remove 2 R1 markers, remove `omitempty` from 2 R2 fields |
| 7 | `api/v1alpha1/certauthenginerole_types.go` | Modified | Remove 5 R1 markers, remove `omitempty` from 1 R2 field |
| 8 | `api/v1alpha1/databasesecretengineconfig_types.go` | Modified | Remove `omitempty` from 2 R2 fields |
| 9 | `api/v1alpha1/databasesecretenginerole_types.go` | Modified | Remove 2 R1 duration default markers |
| 10 | `api/v1alpha1/secretenginemount_types.go` | Modified | Remove 4 R1 markers (1 adds omitempty), remove `omitempty` from 1 R2 field |
| 11 | `api/v1alpha1/authenginemount_types.go` | Modified | Remove `omitempty` from 1 R2 field |
| 12 | `api/v1alpha1/githubsecretengineconfig_types.go` | Modified | Remove `omitempty` from 1 R2 field |
| 13 | `api/v1alpha1/quaysecretengineconfig_types.go` | Modified | Remove 1 R1 default marker |
| 14 | `api/v1alpha1/quaysecretenginerole_types.go` | Modified | Remove `omitempty` from 1 R2 field |
| 15 | `api/v1alpha1/randomsecret_types.go` | Modified | Remove 1 R1 marker, remove `omitempty` from 1 R2 field |
| 16 | `api/v1alpha1/vaultsecret_types.go` | Modified | Remove 1 R1 marker, remove `omitempty` from 3 R2 fields |
| 17 | `api/v1alpha1/auditrequestheader_types.go` | Modified | Remove 1 R1 default marker |
| 18+ | `config/crd/bases/*.yaml` | Regenerated | CRD schemas updated by `make manifests` |

### Project Structure Notes

- All CRD types live in `api/v1alpha1/`
- CRD schemas are regenerated into `config/crd/bases/` by `make manifests`
- DeepCopy is regenerated into `api/v1alpha1/zz_generated.deepcopy.go` by `make generate`
- Unit test files: `api/v1alpha1/*_test.go` — verify they pass after changes
- Integration test files: `controllers/*_controller_test.go` — verify with `make integration`
- Test fixtures: `test/` directory — fixtures set values explicitly, not relying on defaults being modified

### References

- [Source: _bmad-output/project-context.md#CRD Field Default & Validation Rules] — Rules 1-6 governing annotation behavior
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.5.6] — Epic story definition and acceptance criteria
- [Source: api/v1alpha1/identityoidcclient_types.go] — Identity OIDC Client fields
- [Source: api/v1alpha1/identitytokenkey_types.go] — Identity Token Key fields
- [Source: api/v1alpha1/identitytokenrole_types.go] — Identity Token Role fields
- [Source: api/v1alpha1/group_types.go] — Group fields
- [Source: api/v1alpha1/entity_types.go] — Entity fields
- [Source: api/v1alpha1/certauthengineconfig_types.go] — Cert Auth Engine Config fields
- [Source: api/v1alpha1/certauthenginerole_types.go] — Cert Auth Engine Role fields
- [Source: api/v1alpha1/databasesecretengineconfig_types.go] — Database Secret Engine Config fields
- [Source: api/v1alpha1/databasesecretenginerole_types.go] — Database Secret Engine Role fields
- [Source: api/v1alpha1/secretenginemount_types.go] — Secret Engine Mount fields
- [Source: api/v1alpha1/authenginemount_types.go] — Auth Engine Mount fields
- [Source: api/v1alpha1/githubsecretengineconfig_types.go] — GitHub Secret Engine Config fields
- [Source: api/v1alpha1/quaysecretengineconfig_types.go] — Quay Secret Engine Config fields
- [Source: api/v1alpha1/quaysecretenginerole_types.go] — Quay Secret Engine Role fields
- [Source: api/v1alpha1/randomsecret_types.go] — RandomSecret fields
- [Source: api/v1alpha1/vaultsecret_types.go] — VaultSecret fields
- [Source: api/v1alpha1/auditrequestheader_types.go] — AuditRequestHeader fields

### Previous Story Intelligence

**From Story 7.5.1 (LDAP Auth Engine Types — Annotation Refactor):**
- Established patterns: R1 (remove zero-value defaults), R2 (remove omitempty from non-zero defaults)
- Confirmed: `make manifests generate` regenerates CRDs; `make test` validates
- Confirmed: existing unit tests are unaffected by annotation changes (tests use explicit values)

**From Story 7.5.2 (JWT/OIDC Auth Engine Types — Annotation Refactor):**
- Confirms: R1 fields with existing `omitempty` only need default removal (no JSON tag change)
- Confirms: R2 fields with non-zero defaults need `omitempty` removal from JSON tag only

**From Story 7.5.3 (Kubernetes Auth & Secret Engine Types — Annotation Refactor):**
- Confirms: No Go code changes needed — annotation-only refactor
- Confirms: integration tests pass after annotation changes when fixtures use explicit values

**From Story 7.5.4 (Azure & GCP Auth/Secret Engine Types — Annotation Refactor):**
- Confirmed multi-struct refactors work cleanly (6 files in that story)
- Confirmed Enum markers that already exist need no modification

**From Story 7.5.5 (PKI Secret Engine Types — Annotation Refactor):**
- Confirms: Duration fields with `"0s"` default only need marker removal (R1), keep `omitempty`
- Confirmed: Fields appearing in BOTH config and role files must be changed in both
- Confirmed: integration tests pass post-refactor

**Key differences from Stories 7.5.1-7.5.5:**
- This story is the **largest** in Epic 7.5: 17 files, 40 fields (previous largest was 7.5.4 with 6 files)
- Contains the **only ForceNoCache special case** where omitempty must be ADDED (not removed)
- Contains `int64` type R1 fields (`TokenNumUses`) — zero-value `0` for `int64` same as `int`
- Contains `[]string` R2 field (`AllowedRoles` with default `{"*"}`) — slice default pattern
- Includes `VaultSecretDefinition.Path` with `Required` + default — existing intentional pattern, don't modify
- This is the **sweep story** — Task 8 must confirm zero remaining violations across the entire codebase
- No Enum markers need to be added or modified (that's Story 7.5.7's scope)

**From Epic 7 Retrospective:**
- Codebase is stable after hardening epic
- Comprehensive test coverage exists across all modified types
- Integration tests cover identity, group, entity, database, mount, randomsecret, vaultsecret, and audit types

### Git Intelligence

- Latest commit: `44cad20` (Bmad epic 7 squash merge)
- Branch is clean on main
- No pending changes that could conflict with this annotation refactor
- All CI checks passing on main

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
