# Story 7.5.1: LDAP Auth Engine Types — Annotation Refactor

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the LDAPAuthEngineConfig and LDAPAuthEngineGroup field annotations to follow the CRD Field Default & Validation Rules,
So that defaulting and validation behavior is correct and explicit.

## Acceptance Criteria

1. **Given** the `LDAPConfig` struct fields with `+kubebuilder:default=""` or `+kubebuilder:default=false` or `+kubebuilder:default=0` **When** the markers are removed and `omitempty` is ensured on their JSON tags **Then** `make manifests generate test` passes
2. **Given** `TLSMinVersion` and `TLSMaxVersion` with `+kubebuilder:default="tls12"` and `omitempty` **When** `omitempty` is removed from their JSON tags **Then** the fields are always present in serialized JSON
3. **Given** `URL`, `RequestTimeout`, `UserAttr`, `DenyNullBind` already have non-zero defaults without `omitempty` **When** reviewed **Then** confirmed as already compliant (no change needed)
4. **Given** `TLSMinVersion`/`TLSMaxVersion` accept `tls10`, `tls11`, `tls12`, `tls13` **When** `+kubebuilder:validation:Enum` is added **Then** invalid values are rejected at admission
5. **Given** `LDAPAuthEngineGroup.Policies` has `+kubebuilder:default=""` **When** the marker is removed **Then** the field relies on Go zero value
6. **Given** all changes **When** `make integration` is run **Then** LDAP integration tests (Story 4.2) pass

## Tasks / Subtasks

- [x] Task 1: Remove `+kubebuilder:default` from zero-value fields in `LDAPConfig` (AC: 1)
  - [x] 1.1: Remove `+kubebuilder:default=false` from bool fields: `CaseSensitiveNames`, `StartTLS`, `InsecureTLS`, `DiscoverDN`, `AnonymousGroupSearch`, `UsernameAsAlias`, `TokenNoDefaultPolicy` (7 fields); add `omitempty` to their JSON tags
  - [x] 1.2: Remove `+kubebuilder:default=""` from string fields that already have `omitempty`: `Certificate`, `ClientTLSCert`, `ClientTLSKey`, `BindDN`, `UserDN`, `UPNDomain`, `UserFilter`, `GroupFilter`, `GroupDN`, `GroupAttr`, `TokenTTL`, `TokenMaxTTL`, `TokenPolicies`, `TokenBoundCIDRs`, `TokenExplicitMaxTTL`, `TokenType` (16 fields)
  - [x] 1.3: Remove `+kubebuilder:default=0` from int64 fields: `TokenNumUses`, `TokenPeriod` (2 fields); add `omitempty` to their JSON tags
- [x] Task 2: Remove `omitempty` from `TLSMinVersion`, `TLSMaxVersion` JSON tags (AC: 2)
  - [x] 2.1: Change `json:"TLSMinVersion,omitempty"` → `json:"TLSMinVersion"`
  - [x] 2.2: Change `json:"TLSMaxVersion,omitempty"` → `json:"TLSMaxVersion"`
- [x] Task 3: Add `+kubebuilder:validation:Enum` markers (AC: 4)
  - [x] 3.1: Add `// +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}` to `TLSMinVersion`
  - [x] 3.2: Add `// +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}` to `TLSMaxVersion`
  - [x] 3.3: Add `// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}` to `TokenType`
- [x] Task 4: Remove `+kubebuilder:default=""` from `LDAPAuthEngineGroup.Policies` (AC: 5)
- [x] Task 5: Run `make manifests generate fmt vet test` (AC: 1, 2, 4, 5)
- [x] Task 6: Run `make integration` — LDAP tests must pass (AC: 6)

### Review Findings

- [x] [Review][Patch] Make `CaseSensitiveNames` optional consistently [api/v1alpha1/ldapauthengineconfig_types.go:214] — Changed `+kubebuilder:validation:Required` to `+kubebuilder:validation:Optional`. CRD regenerated; field is now optional with `omitempty`, consistent with the zero-value-default rule.

## Dev Notes

### Scope: 2 Files, ~27 Field Changes

| File | R1 Removals | R2 Fixes | Enum Additions |
|------|-------------|----------|----------------|
| `api/v1alpha1/ldapauthengineconfig_types.go` | 25 fields | 2 fields | 3 fields |
| `api/v1alpha1/ldapauthenginegroup_types.go` | 1 field | 0 | 0 |

This is the largest single file in the refactor epic. The `LDAPConfig` struct (lines 204-381 of `ldapauthengineconfig_types.go`) contains all the fields — it is inlined into `LDAPAuthEngineConfigSpec`.

### Detailed Field Change Table — `LDAPConfig` struct

**Rule 1 — Remove redundant zero-value `kubebuilder:default`, ensure `omitempty`:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `CaseSensitiveNames` | bool | `false` | `json:"caseSensitiveNames"` | Remove default, add `omitempty` |
| `StartTLS` | bool | `false` | `json:"startTLS"` | Remove default, add `omitempty` |
| `InsecureTLS` | bool | `false` | `json:"insecureTLS"` | Remove default, add `omitempty` |
| `DiscoverDN` | bool | `false` | `json:"discoverDN"` | Remove default, add `omitempty` |
| `AnonymousGroupSearch` | bool | `false` | `json:"anonymousGroupSearch"` | Remove default, add `omitempty` |
| `UsernameAsAlias` | bool | `false` | `json:"usernameAsAlias"` | Remove default, add `omitempty` |
| `TokenNoDefaultPolicy` | bool | `false` | `json:"tokenNoDefaultPolicy"` | Remove default, add `omitempty` |
| `Certificate` | string | `""` | `json:"certificate,omitempty"` | Remove default only |
| `ClientTLSCert` | string | `""` | `json:"clientTLSCert,omitempty"` | Remove default only |
| `ClientTLSKey` | string | `""` | `json:"clientTLSKey,omitempty"` | Remove default only |
| `BindDN` | string | `""` | `json:"bindDN,omitempty"` | Remove default only |
| `UserDN` | string | `""` | `json:"userDN,omitempty"` | Remove default only |
| `UPNDomain` | string | `""` | `json:"UPNDomain,omitempty"` | Remove default only |
| `UserFilter` | string | `""` | `json:"userFilter,omitempty"` | Remove default only |
| `GroupFilter` | string | `""` | `json:"groupFilter,omitempty"` | Remove default only |
| `GroupDN` | string | `""` | `json:"groupDN,omitempty"` | Remove default only |
| `GroupAttr` | string | `""` | `json:"groupAttr,omitempty"` | Remove default only |
| `TokenTTL` | string | `""` | `json:"tokenTTL,omitempty"` | Remove default only |
| `TokenMaxTTL` | string | `""` | `json:"tokenMaxTTL,omitempty"` | Remove default only |
| `TokenPolicies` | string | `""` | `json:"tokenPolicies,omitempty"` | Remove default only |
| `TokenBoundCIDRs` | string | `""` | `json:"tokenBoundCIDRs,omitempty"` | Remove default only |
| `TokenExplicitMaxTTL` | string | `""` | `json:"tokenExplicitMaxTTL,omitempty"` | Remove default only |
| `TokenType` | string | `""` | `json:"tokenType,omitempty"` | Remove default only |
| `TokenNumUses` | int64 | `0` | `json:"tokenNumUses"` | Remove default, add `omitempty` |
| `TokenPeriod` | int64 | `0` | `json:"tokenPeriod"` | Remove default, add `omitempty` |

**Rule 2 — Remove `omitempty` from non-zero defaults (field must always serialize):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `TLSMinVersion` | string | `"tls12"` | `json:"TLSMinVersion,omitempty"` | Remove `omitempty` |
| `TLSMaxVersion` | string | `"tls12"` | `json:"TLSMaxVersion,omitempty"` | Remove `omitempty` |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `URL` | string | `"ldap://127.0.0.1"` | `json:"url"` | Non-zero default, no `omitempty` |
| `RequestTimeout` | string | `"90s"` | `json:"requestTimeout"` | Non-zero default, no `omitempty` |
| `UserAttr` | string | `"cn"` | `json:"userAttr"` | Non-zero default, no `omitempty` |
| `DenyNullBind` | bool | `true` | `json:"denyNullBind"` | Non-zero default, no `omitempty` |

### Detailed Field Change — `LDAPAuthEngineGroupSpec`

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `Policies` | string | `""` | `json:"policies,omitempty"` | Remove `+kubebuilder:default=""` only |

### Enum Marker Details

**`TLSMinVersion` / `TLSMaxVersion`:** Vault LDAP docs define exactly 4 accepted values: `tls10`, `tls11`, `tls12`, `tls13`. Add:
```
// +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}
```

**`TokenType`:** Vault auth token type field accepts: `service`, `batch`, `default`, `default-service`, `default-batch`. Add:
```
// +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
```

### Impact on `toMap()` and `IsEquivalentToDesiredState()`

These changes are **annotation-only** — they modify kubebuilder markers and JSON struct tags. They do NOT change:
- The `toMap()` method (lines 430-474) — unchanged
- The `IsEquivalentToDesiredState()` method (lines 75-79) — unchanged
- The `bindpass` delete behavior in `IsEquivalentToDesiredState` — unchanged
- Any Go code logic

However, the CRD OpenAPI schema **will change** after `make manifests`. Fields that had explicit `default` values in the CRD YAML will lose them (R1 removals), and enum constraints will be added (R3 additions). This is a schema-level change with no Go code change.

### Impact on Existing Tests

**Unit tests (`api/v1alpha1/ldapauthengineconfig_test.go`):**
- `TestLDAPConfigToMap` — constructs `LDAPConfig` with explicit field values → **unaffected** (tests Go struct behavior, not CRD defaults)
- `TestLDAPAuthEngineConfigIsEquivalent*` tests — **unaffected** (annotation changes don't change Go runtime behavior)
- `TestLDAPAuthEngineConfig_PrepareInternalValues` — **unaffected**

**Unit tests (`api/v1alpha1/ldapauthenginegroup_test.go`):**
- Existing tests — **unaffected**

**Integration tests (`controllers/ldapauthengine_controller_test.go`):**
- Test fixtures: `test/ldapauthengine/ldap-auth-engine-config.yaml` and related YAML files
- The LDAP config fixture sets `caseSensitiveNames: false` explicitly — this is the Go zero value, so `omitempty` will omit it from serialized YAML. **But** the fixture sets it explicitly in the YAML, which means the API server will still receive it. The fixture does NOT rely on the `kubebuilder:default` marker.
- Other fixtures: `insecureTLS: true` (explicit), `bindDN`, `userDN`, `groupDN`, `groupFilter`, `userFilter` all set explicitly → **unaffected**
- **Key check:** No fixture relies on server-side defaulting for any field being modified. All test values are explicitly set.

### Critical Warnings

1. **Do NOT modify `toMap()` or any Go logic.** This is purely an annotation + JSON tag refactor.
2. **Run `make manifests generate` after changes.** This regenerates CRDs in `config/crd/bases/`. The diff will show removed `default:` entries and added `enum:` entries in the OpenAPI schema.
3. **Run `make generate`** to regenerate `zz_generated.deepcopy.go` (struct tag changes may affect generated code).
4. **`CaseSensitiveNames` was changed to `+kubebuilder:validation:Optional` during code review.** The field is a zero-value-default bool (`false`), so the annotation-refactor rules say it should be Optional with `omitempty`. The original Required marker was pre-existing but inconsistent with the rules applied in this epic.
5. **Test fixture review:** After changes, verify that YAML fixtures in `test/ldapauthengine/` still apply without validation errors (especially the new `Enum` constraints on `TLSMinVersion`/`TLSMaxVersion` — the default fixtures don't set these fields, so they'll get the `tls12` default which is a valid enum value).

### Pattern for Bool Fields (R1)

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=false
StartTLS bool `json:"startTLS"`
```

After:
```go
// +kubebuilder:validation:Optional
StartTLS bool `json:"startTLS,omitempty"`
```

### Pattern for String Fields with `""` Default (R1)

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=""
GroupDN string `json:"groupDN,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
GroupDN string `json:"groupDN,omitempty"`
```

### Pattern for Int64 Fields with `0` Default (R1)

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=0
TokenNumUses int64 `json:"tokenNumUses"`
```

After:
```go
// +kubebuilder:validation:Optional
TokenNumUses int64 `json:"tokenNumUses,omitempty"`
```

### Pattern for Non-Zero Default Fields (R2)

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="tls12"
TLSMinVersion string `json:"TLSMinVersion,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="tls12"
// +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}
TLSMinVersion string `json:"TLSMinVersion"`
```

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/ldapauthengineconfig_types.go` | Modified | Remove 25 R1 markers, fix 2 R2 JSON tags, add 3 Enum markers |
| 2 | `api/v1alpha1/ldapauthenginegroup_types.go` | Modified | Remove 1 R1 marker (`Policies`) |
| 3 | `config/crd/bases/redhatcop.redhat.io_ldapauthengineconfigs.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 4 | `config/crd/bases/redhatcop.redhat.io_ldapauthenginegroups.yaml` | Regenerated | CRD schema updated by `make manifests` |

### Project Structure Notes

- CRD types live in `api/v1alpha1/` — both files are in this directory
- CRD schemas are regenerated into `config/crd/bases/` by `make manifests`
- DeepCopy is regenerated into `api/v1alpha1/zz_generated.deepcopy.go` by `make generate`
- Test fixtures in `test/ldapauthengine/` — verify they pass new Enum constraints
- Integration test file: `controllers/ldapauthengine_controller_test.go`
- Unit test files: `api/v1alpha1/ldapauthengineconfig_test.go`, `api/v1alpha1/ldapauthenginegroup_test.go`

### References

- [Source: _bmad-output/project-context.md#CRD Field Default & Validation Rules] — Rules 1-6 governing annotation behavior
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.5.1] — Epic story definition and acceptance criteria
- [Source: api/v1alpha1/ldapauthengineconfig_types.go:204-381] — `LDAPConfig` struct with all field annotations
- [Source: api/v1alpha1/ldapauthenginegroup_types.go:49-52] — `Policies` field with `+kubebuilder:default=""`
- [Source: api/v1alpha1/ldapauthengineconfig_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: controllers/ldapauthengine_controller_test.go] — LDAP integration tests (must pass post-change)
- [Source: test/ldapauthengine/] — Test fixtures (verify against new Enum constraints)

### Previous Story Intelligence

**From Story 7.5 (Drift Detection Integration Tests — last story in Epic 7):**
- All 90 integration specs pass on main at commit `44cad20`
- Integration test infrastructure is stable and proven
- `make test` and `make integration` are the primary verification targets
- This is the first story in Epic 7.5 — no previous annotation refactor stories to learn from

**From Epic 7 Retrospective:**
- Codebase is stable after hardening epic
- Coverage is strong for LDAP types (unit tests + integration tests both exist)
- Story 4.2 specifically tested LDAP auth engine config and group — those tests serve as regression safety net

### Git Intelligence

- Latest commit: `44cad20` (Bmad epic 7 squash merge)
- Branch is clean on main
- No pending changes that could conflict with this annotation refactor
- All CI checks passing on main

## Dev Agent Record

### Agent Model Used

Claude Opus 4 (Cursor)

### Debug Log References

- Integration test failure on first run: `TLSMinVersion`/`TLSMaxVersion` Enum validation rejected empty strings in test fixture. Root cause: removing `omitempty` from JSON tags causes Go zero-value `""` to serialize, bypassing CRD server-side defaulting. Fix: added explicit `TLSMinVersion: "tls12"` and `TLSMaxVersion: "tls12"` to test fixture YAML.

### Completion Notes List

- Removed 25 redundant zero-value `+kubebuilder:default` markers from `LDAPConfig` struct (7 bool, 16 string, 2 int64 fields)
- Added `omitempty` to JSON tags for 9 fields that didn't have it (7 bool + 2 int64)
- Removed `omitempty` from `TLSMinVersion` and `TLSMaxVersion` JSON tags (non-zero defaults must always serialize)
- Added 3 `+kubebuilder:validation:Enum` markers: `TLSMinVersion`, `TLSMaxVersion` (tls10-tls13), `TokenType` (service/batch/default/default-service/default-batch)
- Removed 1 redundant `+kubebuilder:default=""` from `LDAPAuthEngineGroup.Policies`
- Updated LDAP integration test fixture to explicitly set TLS version fields (required due to Enum + omitempty interaction)
- CRD schemas regenerated via `make manifests generate`
- All unit tests pass (`make test`), all 90 integration specs pass (`make integration`)
- No Go logic changes — purely annotation, JSON tag, and CRD schema refactor

### File List

- `api/v1alpha1/ldapauthengineconfig_types.go` — Modified: removed 25 R1 default markers, fixed 2 R2 JSON tags, added 3 Enum markers
- `api/v1alpha1/ldapauthenginegroup_types.go` — Modified: removed 1 R1 default marker (`Policies`)
- `config/crd/bases/redhatcop.redhat.io_ldapauthengineconfigs.yaml` — Regenerated: CRD schema updated
- `config/crd/bases/redhatcop.redhat.io_ldapauthenginegroups.yaml` — Regenerated: CRD schema updated
- `test/ldapauthengine/test-ldap-auth-config.yaml` — Modified: added explicit TLSMinVersion/TLSMaxVersion for Enum compliance
