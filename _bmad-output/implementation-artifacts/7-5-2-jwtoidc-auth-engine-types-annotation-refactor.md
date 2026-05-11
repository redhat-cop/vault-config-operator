# Story 7.5.2: JWT/OIDC Auth Engine Types — Annotation Refactor

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the JWTOIDCAuthEngineConfig and JWTOIDCAuthEngineRole field annotations to follow the CRD Field Default & Validation Rules,
So that the 30+ affected fields have correct defaulting and validation semantics.

## Acceptance Criteria

1. **Given** all `+kubebuilder:default=""` / `false` / `0` / `{}` fields **When** markers are removed and `omitempty` ensured **Then** `make manifests generate test` passes
2. **Given** `NamespaceInState` (default `true`, `omitempty`) **When** `omitempty` is removed **Then** field always serialized
3. **Given** `BoundClaimsType` (default `"string"`, `omitempty`) **When** `omitempty` is removed **Then** field always serialized
4. **Given** `OIDCResponseMode` has documented values (`query`, `form_post`) **When** `+kubebuilder:validation:Enum` is added **Then** invalid values rejected
5. **Given** all changes **When** `make manifests generate fmt vet test` passes **Then** no regressions

## Tasks / Subtasks

- [ ] Task 1: Remove `+kubebuilder:default` from zero-value fields in `JWTOIDCConfig` (AC: 1)
  - [ ] 1.1: Remove `+kubebuilder:default=""` from 8 string fields; add `omitempty` to `OIDCDiscoveryURL` JSON tag (the only one missing it)
  - [ ] 1.2: Remove `+kubebuilder:default={}` from `ProviderConfig`
- [ ] Task 2: Remove `omitempty` from `NamespaceInState` JSON tag (AC: 2)
  - [ ] 2.1: Change `json:"namespaceInState,omitempty"` → `json:"namespaceInState"`
- [ ] Task 3: Remove `+kubebuilder:default` from zero-value fields in `JWTOIDCRole` (AC: 1)
  - [ ] 3.1: Remove `+kubebuilder:default=""` from 7 string fields (`RoleType`, `BoundSubject`, `GroupsClaim`, `TokenTTL`, `TokenMaxTTL`, `TokenExplicitMaxTTL`, `TokenType`)
  - [ ] 3.2: Remove `+kubebuilder:default=false` from 3 bool fields (`UserClaimJSONPointer`, `VerboseOIDCLogging`, `TokenNoDefaultPolicy`); add `omitempty` to their JSON tags
  - [ ] 3.3: Remove `+kubebuilder:default=0` from 5 int64 fields (`ClockSkewLeeway`, `ExpirationLeeway`, `NotBeforeLeeway`, `MaxAge`, `TokenNumUses`, `TokenPeriod`); add `omitempty` to their JSON tags
  - [ ] 3.4: Remove `+kubebuilder:default={}` from `BoundClaims` and `ClaimMappings`
- [ ] Task 4: Remove `omitempty` from `BoundClaimsType` JSON tag (AC: 3)
  - [ ] 4.1: Change `json:"boundClaimsType,omitempty"` → `json:"boundClaimsType"`
- [ ] Task 5: Add `+kubebuilder:validation:Enum` markers (AC: 4)
  - [ ] 5.1: Add `// +kubebuilder:validation:Enum={"query","form_post"}` to `OIDCResponseMode`
  - [ ] 5.2: Add `// +kubebuilder:validation:Enum={"string","glob"}` to `BoundClaimsType`
  - [ ] 5.3: Add `// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}` to `TokenType`
- [ ] Task 6: Run `make manifests generate fmt vet test` (AC: 1, 2, 3, 4, 5)
- [ ] Task 7: Run `make integration` — JWT/OIDC tests must pass (AC: 5)

## Dev Notes

### Scope: 2 Files, ~29 Field Changes

| File | R1 Removals | R2 Fixes | Enum Additions |
|------|-------------|----------|----------------|
| `api/v1alpha1/jwtoidcauthengineconfig_types.go` | 9 fields | 1 field | 1 field |
| `api/v1alpha1/jwtoidcauthenginerole_types.go` | 18 fields | 1 field | 2 fields |

The `JWTOIDCConfig` struct (lines 86-168 of `jwtoidcauthengineconfig_types.go`) contains all config-level fields. The `JWTOIDCRole` struct (lines 49-209 of `jwtoidcauthenginerole_types.go`) contains all role-level fields.

### Detailed Field Change Table — `JWTOIDCConfig` struct

**Rule 1 — Remove redundant zero-value `kubebuilder:default`, ensure `omitempty`:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `OIDCDiscoveryURL` | string | `""` | `json:"OIDCDiscoveryURL"` | Remove default, **add `omitempty`** |
| `OIDCDiscoveryCAPEM` | string | `""` | `json:"OIDCDiscoveryCAPEM,omitempty"` | Remove default only |
| `OIDCClientID` | string | `""` | `json:"OIDCClientID,omitempty"` | Remove default only |
| `OIDCResponseMode` | string | `""` | `json:"OIDCResponseMode,omitempty"` | Remove default only |
| `JWKSURL` | string | `""` | `json:"JWKSURL,omitempty"` | Remove default only |
| `JWKSCAPEM` | string | `""` | `json:"JWKSCAPEM,omitempty"` | Remove default only |
| `BoundIssuer` | string | `""` | `json:"boundIssuer,omitempty"` | Remove default only |
| `DefaultRole` | string | `""` | `json:"defaultRole,omitempty"` | Remove default only |
| `ProviderConfig` | *JSON | `{}` | `json:"providerConfig,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults (field must always serialize):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `NamespaceInState` | bool | `true` | `json:"namespaceInState,omitempty"` | Remove `omitempty` |

### Detailed Field Change Table — `JWTOIDCRole` struct

**Rule 1 — Remove redundant zero-value `kubebuilder:default`, ensure `omitempty`:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `RoleType` | string | `""` | `json:"roleType,omitempty"` | Remove default only |
| `UserClaimJSONPointer` | bool | `false` | `json:"userClaimJSONPointer"` | Remove default, add `omitempty` |
| `ClockSkewLeeway` | int64 | `0` | `json:"clockSkewLeeway"` | Remove default, add `omitempty` |
| `ExpirationLeeway` | int64 | `0` | `json:"expirationLeeway"` | Remove default, add `omitempty` |
| `NotBeforeLeeway` | int64 | `0` | `json:"notBeforeLeeway"` | Remove default, add `omitempty` |
| `BoundSubject` | string | `""` | `json:"boundSubject,omitempty"` | Remove default only |
| `BoundClaims` | *JSON | `{}` | `json:"boundClaims,omitempty"` | Remove default only |
| `GroupsClaim` | string | `""` | `json:"groupsClaim,omitempty"` | Remove default only |
| `ClaimMappings` | map | `{}` | `json:"claimMappings,omitempty"` | Remove default only |
| `VerboseOIDCLogging` | bool | `false` | `json:"verboseOIDCLogging"` | Remove default, add `omitempty` |
| `MaxAge` | int64 | `0` | `json:"maxage"` | Remove default, add `omitempty` |
| `TokenTTL` | string | `""` | `json:"tokenTTL,omitempty"` | Remove default only |
| `TokenMaxTTL` | string | `""` | `json:"tokenMaxTTL,omitempty"` | Remove default only |
| `TokenExplicitMaxTTL` | string | `""` | `json:"tokenExplicitMaxTTL,omitempty"` | Remove default only |
| `TokenNoDefaultPolicy` | bool | `false` | `json:"tokenNoDefaultPolicy"` | Remove default, add `omitempty` |
| `TokenNumUses` | int64 | `0` | `json:"tokenNumUses"` | Remove default, add `omitempty` |
| `TokenPeriod` | int64 | `0` | `json:"tokenPeriod"` | Remove default, add `omitempty` |
| `TokenType` | string | `""` | `json:"tokenType,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults (field must always serialize):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `BoundClaimsType` | string | `"string"` | `json:"boundClaimsType,omitempty"` | Remove `omitempty` |

### Enum Marker Details

**`OIDCResponseMode`:** Vault docs define allowed values: `query`, `form_post`. Add:
```
// +kubebuilder:validation:Enum={"query","form_post"}
```

**`BoundClaimsType`:** Vault docs define allowed values: `string`, `glob`. Add:
```
// +kubebuilder:validation:Enum={"string","glob"}
```

**`TokenType`:** Vault auth token type field accepts: `service`, `batch`, `default`, `default-service`, `default-batch`. Add (consistent with Story 7.5.1 LDAP):
```
// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
```

### Impact on `toMap()` and `IsEquivalentToDesiredState()`

These changes are **annotation-only** — they modify kubebuilder markers and JSON struct tags. They do NOT change:
- `JWTOIDCConfig.toMap()` (lines 303-321) — unchanged
- `JWTOIDCRole.toMap()` (lines 297-327) — unchanged
- `JWTOIDCAuthEngineConfig.IsEquivalentToDesiredState()` (lines 202-205) — unchanged
- `JWTOIDCAuthEngineRole.IsEquivalentToDesiredState()` (lines 276-279) — unchanged
- Any Go code logic

The CRD OpenAPI schema **will change** after `make manifests`. Fields that had explicit `default` values in the CRD YAML will lose them (R1 removals), and enum constraints will be added. This is a schema-level change with no Go code change.

### Impact on Existing Tests

**Unit tests (`api/v1alpha1/jwtoidcauthengineconfig_test.go`):**
- `TestJWTOIDCConfigToMap` — constructs `JWTOIDCConfig` with explicit field values → **unaffected**
- `TestJWTOIDCConfigToMapUnexportedCredentials` — tests unexported credential fields → **unaffected**
- `TestJWTOIDCConfigToMapProviderConfigJSON` — tests ProviderConfig JSON handling → **unaffected**
- `TestJWTOIDCAuthEngineConfigIsEquivalent*` tests — **unaffected** (annotation changes don't change Go runtime behavior)
- `TestJWTOIDCAuthEngineConfigIsDeletable` — **unaffected**
- `TestJWTOIDCAuthEngineConfigConditions` — **unaffected**

**Unit tests (`api/v1alpha1/jwtoidcauthenginerole_test.go`):**
- `TestJWTOIDCRoleToMap` — constructs `JWTOIDCRole` with explicit field values → **unaffected**
- `TestJWTOIDCRoleToMapBoundClaimsJSON` — tests BoundClaims JSON handling → **unaffected**
- `TestJWTOIDCRoleToMapClaimMappings` — tests ClaimMappings map handling → **unaffected**
- `TestJWTOIDCAuthEngineRoleIsEquivalent*` tests — **unaffected**
- `TestJWTOIDCAuthEngineRoleIsDeletable` — **unaffected**
- `TestJWTOIDCAuthEngineRoleConditions` — **unaffected**

**Integration tests (`controllers/jwtoidcauthengine_controller_test.go`):**
- Test fixtures: `test/jwtoidcauthengine/test-jwtoidc-auth-config.yaml` and `test-jwtoidc-auth-role.yaml`
- Config fixture sets `OIDCDiscoveryURL`, `OIDCClientID`, `OIDCCredentials` explicitly — does NOT rely on any CRD defaults being removed
- Role fixture sets `roleType: "oidc"`, `userClaim: "email"`, `allowedRedirectURIs`, `tokenPolicies: ["default"]`, `OIDCScopes: ["openid", "email"]` — all explicit values
- `BoundClaimsType` not set in fixture → API server will apply `+kubebuilder:default="string"` (retained, R2 just fixes omitempty) → `"string"` is a valid Enum value
- `OIDCResponseMode` not set in fixture → field absent (omitempty) → Enum not triggered, Vault uses its own default
- `TokenType` not set in fixture → field absent (omitempty) → Enum not triggered
- **No fixture relies on server-side defaulting for any R1 field being modified**

### Critical Warnings

1. **Do NOT modify `toMap()` or any Go logic.** This is purely an annotation + JSON tag refactor.
2. **Run `make manifests generate` after changes.** This regenerates CRDs in `config/crd/bases/`. The diff will show removed `default:` entries and added `enum:` entries in the OpenAPI schema.
3. **Run `make generate`** to regenerate `zz_generated.deepcopy.go` (struct tag changes may affect generated code).
4. **`OIDCResponseMode` Enum and empty values:** The field is Optional with omitempty. Users who don't set it will have it absent from the request, so the Enum is not triggered. Users who DO set it must provide `"query"` or `"form_post"`. This matches Vault's documented behavior.
5. **`BoundClaimsType` keeps its non-zero default.** After R2 fix (removing omitempty), the field always serializes with its default `"string"`. The Enum={"string","glob"} validates user input. Safe because the default value is a valid enum member.
6. **`NamespaceInState` keeps its non-zero default `true`.** After R2 fix (removing omitempty), the field always serializes. No Enum needed (boolean field).
7. **Test fixture review:** After changes, verify that YAML fixtures in `test/jwtoidcauthengine/` still apply without validation errors. All fixtures use explicit values that comply with new Enum constraints.

### Pattern for Bool Fields (R1)

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=false
UserClaimJSONPointer bool `json:"userClaimJSONPointer"`
```

After:
```go
// +kubebuilder:validation:Optional
UserClaimJSONPointer bool `json:"userClaimJSONPointer,omitempty"`
```

### Pattern for String Fields with `""` Default (R1)

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=""
GroupsClaim string `json:"groupsClaim,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
GroupsClaim string `json:"groupsClaim,omitempty"`
```

### Pattern for Int64 Fields with `0` Default (R1)

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=0
ClockSkewLeeway int64 `json:"clockSkewLeeway"`
```

After:
```go
// +kubebuilder:validation:Optional
ClockSkewLeeway int64 `json:"clockSkewLeeway,omitempty"`
```

### Pattern for Non-Zero Default Fields (R2) + Enum

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="string"
BoundClaimsType string `json:"boundClaimsType,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="string"
// +kubebuilder:validation:Enum={"string","glob"}
BoundClaimsType string `json:"boundClaimsType"`
```

### Pattern for `OIDCDiscoveryURL` (R1 — missing omitempty)

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=""
OIDCDiscoveryURL string `json:"OIDCDiscoveryURL"`
```

After:
```go
// +kubebuilder:validation:Optional
OIDCDiscoveryURL string `json:"OIDCDiscoveryURL,omitempty"`
```

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/jwtoidcauthengineconfig_types.go` | Modified | Remove 9 R1 markers, fix 1 R2 JSON tag, add 1 Enum marker |
| 2 | `api/v1alpha1/jwtoidcauthenginerole_types.go` | Modified | Remove 18 R1 markers, fix 1 R2 JSON tag, add 2 Enum markers |
| 3 | `config/crd/bases/redhatcop.redhat.io_jwtoidcauthengineconfigs.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 4 | `config/crd/bases/redhatcop.redhat.io_jwtoidcauthengineroles.yaml` | Regenerated | CRD schema updated by `make manifests` |

### Project Structure Notes

- CRD types live in `api/v1alpha1/` — both files are in this directory
- CRD schemas are regenerated into `config/crd/bases/` by `make manifests`
- DeepCopy is regenerated into `api/v1alpha1/zz_generated.deepcopy.go` by `make generate`
- Test fixtures in `test/jwtoidcauthengine/` — verify they pass new Enum constraints
- Integration test file: `controllers/jwtoidcauthengine_controller_test.go`
- Unit test files: `api/v1alpha1/jwtoidcauthengineconfig_test.go`, `api/v1alpha1/jwtoidcauthenginerole_test.go`

### References

- [Source: _bmad-output/project-context.md#CRD Field Default & Validation Rules] — Rules 1-6 governing annotation behavior
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.5.2] — Epic story definition and acceptance criteria
- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go:86-168] — `JWTOIDCConfig` struct with all field annotations
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go:49-209] — `JWTOIDCRole` struct with all field annotations
- [Source: api/v1alpha1/jwtoidcauthengineconfig_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/jwtoidcauthenginerole_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: controllers/jwtoidcauthengine_controller_test.go] — JWT/OIDC integration tests (must pass post-change)
- [Source: test/jwtoidcauthengine/] — Test fixtures (verify against new Enum constraints)

### Previous Story Intelligence

**From Story 7.5.1 (LDAP Auth Engine Types — Annotation Refactor):**
- Story 7.5.1 is ready-for-dev (not yet implemented) — same pattern applies identically
- Established patterns: R1 (remove zero-value defaults + ensure omitempty), R2 (remove omitempty from non-zero defaults), Enum additions
- `TokenType` Enum defined as `{"service","batch","default","default-service","default-batch"}` — reuse same values here
- `TLSMinVersion`/`TLSMaxVersion` Enum pattern from LDAP not applicable here (no TLS version fields in JWT/OIDC types)
- Bool fields that lose `+kubebuilder:default=false` gain `omitempty` on JSON tag — same pattern
- Int64 fields that lose `+kubebuilder:default=0` gain `omitempty` on JSON tag — same pattern

**From Epic 7 Retrospective:**
- Codebase is stable after hardening epic
- Coverage is strong for JWT/OIDC types (unit tests + integration tests both exist)
- Story 4.3 specifically tested JWT/OIDC auth engine config and role — those tests serve as regression safety net

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
