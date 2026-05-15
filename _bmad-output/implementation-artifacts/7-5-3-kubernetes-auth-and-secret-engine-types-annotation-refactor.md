# Story 7.5.3: Kubernetes Auth & Secret Engine Types — Annotation Refactor

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the Kubernetes auth engine and secret engine types to follow the CRD Field Default & Validation Rules,
So that defaulting semantics are consistent with the rest of the codebase.

## Acceptance Criteria

1. **Given** zero-value defaults on `DisableISSValidation`, `DisableLocalCAJWT` (auth config), token int fields, `TokenNoDefaultPolicy` (auth role), `DisableLocalCAJWT` (secret config), `DefaultTTL`/`MaxTTL` (secret role) **When** `+kubebuilder:default` markers are removed **Then** `make manifests generate test` passes
2. **Given** `KubernetesHost` and `UseOperatorPodCA` (auth config) have non-zero defaults with `omitempty` **When** `omitempty` is removed from their JSON tags **Then** the fields are always present in serialized JSON
3. **Given** `AliasNameSource` and `TokenType` (auth role) have non-zero defaults with `omitempty` **When** `omitempty` is removed from their JSON tags **Then** the fields are always present in serialized JSON
4. **Given** `KubernetesRoleType` (secret role) has non-zero default `"Role"` with `omitempty` **When** `omitempty` is removed **Then** the field is always present in serialized JSON
5. **Given** all changes **When** `make integration` is run **Then** Kubernetes auth tests (Story 4.1) and secret engine tests (Story 5.3) pass

## Tasks / Subtasks

- [x] Task 1: Remove `+kubebuilder:default` from zero-value fields in `KAECConfig` (AC: 1)
  - [x] 1.1: Remove `+kubebuilder:default=false` from `DisableISSValidation` (line 143); JSON tag already has `omitempty`
  - [x] 1.2: Remove `+kubebuilder:default=false` from `DisableLocalCAJWT` (line 148); JSON tag already has `omitempty`
- [x] Task 2: Remove `omitempty` from non-zero default fields in `KAECConfig` (AC: 2)
  - [x] 2.1: Change `json:"kubernetesHost,omitempty"` → `json:"kubernetesHost"` (line 126)
  - [x] 2.2: Change `json:"useOperatorPodCA,omitempty"` → `json:"useOperatorPodCA"` (line 156)
- [x] Task 3: Remove `+kubebuilder:default` from zero-value fields in `VRole` (AC: 1)
  - [x] 3.1: Remove `+kubebuilder:default:=0` from `TokenTTL` (line 165); JSON tag already has `omitempty`
  - [x] 3.2: Remove `+kubebuilder:default:=0` from `TokenMaxTTL` (line 177); JSON tag already has `omitempty`
  - [x] 3.3: Remove `+kubebuilder:default:=0` from `TokenExplicitMaxTTL` (line 188); JSON tag already has `omitempty`
  - [x] 3.4: Remove `+kubebuilder:default:=false` from `TokenNoDefaultPolicy` (line 192); JSON tag already has `omitempty`
  - [x] 3.5: Remove `+kubebuilder:default:=0` from `TokenNumUses` (line 198); JSON tag already has `omitempty`
  - [x] 3.6: Remove `+kubebuilder:default:=0` from `TokenPeriod` (line 203); JSON tag already has `omitempty`
- [x] Task 4: Remove `omitempty` from non-zero default fields in `VRole` (AC: 3)
  - [x] 4.1: Change `json:"aliasNameSource,omitempty"` → `json:"aliasNameSource"` (line 161)
  - [x] 4.2: Change `json:"tokenType,omitempty"` → `json:"tokenType"` (line 209)
- [x] Task 5: Remove `+kubebuilder:default` from zero-value fields in `KubeSEConfig` (AC: 1)
  - [x] 5.1: Remove `+kubebuilder:default=false` from `DisableLocalCAJWT` (line 191); JSON tag already has `omitempty`
- [x] Task 6: Remove `+kubebuilder:default` from zero-value fields in `KubeSERole` (AC: 1)
  - [x] 6.1: Remove `+kubebuilder:default="0s"` from `DefaultTTL` (line 114); JSON tag already has `omitempty`
  - [x] 6.2: Remove `+kubebuilder:default="0s"` from `MaxTTL` (line 119); JSON tag already has `omitempty`
- [x] Task 7: Remove `omitempty` from non-zero default fields in `KubeSERole` (AC: 4)
  - [x] 7.1: Change `json:"kubernetesRoleType,omitempty"` → `json:"kubernetesRoleType"` (line 139)
- [x] Task 8: Run `make manifests generate fmt vet test` (AC: 1, 2, 3, 4)
- [x] Task 9: Run `make integration` — Kubernetes auth and secret engine tests must pass (AC: 5)

## Dev Notes

### Scope: 4 Files, 16 Field Changes

| File | R1 Removals | R2 Fixes | Enum Additions |
|------|-------------|----------|----------------|
| `api/v1alpha1/kubernetesauthengineconfig_types.go` | 2 fields | 2 fields | 0 |
| `api/v1alpha1/kubernetesauthenginerole_types.go` | 6 fields | 2 fields | 0 |
| `api/v1alpha1/kubernetessecretengineconfig_types.go` | 1 field | 0 | 0 |
| `api/v1alpha1/kubernetessecretenginerole_types.go` | 2 fields | 1 field | 0 |

No new Enum markers in this story — all relevant enums already exist:
- `AliasNameSource` already has `+kubebuilder:validation:Enum:={"serviceaccount_uid", "serviceaccount_name"}`
- `TokenType` already has `+kubebuilder:validation:Enum:={"service","batch","default","default-service","default-batch"}`
- `KubernetesRoleType` already has `+kubebuilder:validation:Enum={"Role","ClusterRole"}`

### Detailed Field Change Table — `KAECConfig` struct (lines 121-163)

**Rule 1 — Remove redundant zero-value `kubebuilder:default` (already have `omitempty`):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `DisableISSValidation` | bool | `false` | `json:"disableISSValidation,omitempty"` | Remove default only |
| `DisableLocalCAJWT` | bool | `false` | `json:"disableLocalCAJWT,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults (field must always serialize):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `KubernetesHost` | string | `"https://kubernetes.default.svc:443"` | `json:"kubernetesHost,omitempty"` | Remove `omitempty` |
| `UseOperatorPodCA` | bool | `true` | `json:"useOperatorPodCA,omitempty"` | Remove `omitempty` |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `KubernetesCACert` | string | none | `json:"kubernetesCACert,omitempty"` | Optional, no default |
| `PEMKeys` | []string | none | `json:"PEMKeys,omitempty"` | Optional, no default |
| `Issuer` | string | none | `json:"issuer,omitempty"` | Optional, no default |
| `UseAnnotationsAsAliasMetadata` | bool | none | `json:"useAnnotationsAsAliasMetadata,omitempty"` | Optional, no default |

### Detailed Field Change Table — `VRole` struct (lines 145-213)

**Rule 1 — Remove redundant zero-value `kubebuilder:default` (already have `omitempty`):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `TokenTTL` | int | `0` | `json:"tokenTTL,omitempty"` | Remove default only |
| `TokenMaxTTL` | int | `0` | `json:"tokenMaxTTL,omitempty"` | Remove default only |
| `TokenExplicitMaxTTL` | int | `0` | `json:"tokenExplicitMaxTTL,omitempty"` | Remove default only |
| `TokenNoDefaultPolicy` | bool | `false` | `json:"tokenNoDefaultPolicy,omitempty"` | Remove default only |
| `TokenNumUses` | int | `0` | `json:"tokenNumUses,omitempty"` | Remove default only |
| `TokenPeriod` | int | `0` | `json:"tokenPeriod,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults (field must always serialize):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `AliasNameSource` | string | `"serviceaccount_uid"` | `json:"aliasNameSource,omitempty"` | Remove `omitempty` |
| `TokenType` | string | `"default"` | `json:"tokenType,omitempty"` | Remove `omitempty` |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `TargetServiceAccounts` | []string | `{"default"}` | `json:"targetServiceAccounts"` | Non-zero default, no `omitempty` |
| `Audience` | *string | none | `json:"audience,omitempty"` | Optional pointer, no default |
| `Policies` | []string | Required | `json:"policies"` | Required, no default |
| `TokenBoundCIDRs` | []string | none | `json:"tokenBoundCIDRs,omitempty"` | Optional, no default |

### Detailed Field Change Table — `KubeSEConfig` struct (lines 180-195)

**Rule 1 — Remove redundant zero-value `kubebuilder:default` (already has `omitempty`):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `DisableLocalCAJWT` | bool | `false` | `json:"disableLocalCAJWT,omitempty"` | Remove default only |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `KubernetesHost` | string | none (Required) | `json:"kubernetesHost,omitempty"` | Required, no default |
| `KubernetesCACert` | string | none | `json:"kubernetesCACert,omitempty"` | Optional, no default |

### Detailed Field Change Table — `KubeSERole` struct (lines 98-156)

**Rule 1 — Remove redundant zero-value `kubebuilder:default`:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `DefaultTTL` | metav1.Duration | `"0s"` | `json:"defaultTTL,omitempty"` | Remove default only |
| `MaxTTL` | metav1.Duration | `"0s"` | `json:"maxTTL,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `KubernetesRoleType` | string | `"Role"` | `json:"kubernetesRoleType,omitempty"` | Remove `omitempty` |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `AllowedKubernetesNamespaces` | []string | none | `json:"allowedKubernetesNamespaces,omitempty"` | Optional, no default |
| `AllowedKubernetesNamespaceSelector` | string | none | `json:"allowedKubernetesNamespaceSelector,omitempty"` | Optional, no default |
| `DefaultAudiences` | string | none | `json:"defaultAudiences,omitempty"` | Optional, no default |
| `ServiceAccountName` | string | none | `json:"serviceAccountName,omitempty"` | Optional, no default |
| `KubernetesRoleName` | string | none | `json:"kubernetesRoleName,omitempty"` | Optional, no default |
| `GenerateRoleRules` | string | none | `json:"generateRoleRules,omitempty"` | Optional, no default |
| `NameTemplate` | string | none | `json:"nameTemplate,omitempty"` | Optional, no default |
| `ExtraAnnotations` | map | none | `json:"extraAnnotations,omitempty"` | Optional, no default |
| `ExtraLabels` | map | none | `json:"extraLabels,omitempty"` | Optional, no default |

### Impact on `toMap()` and `IsEquivalentToDesiredState()`

These changes are **annotation-only** — they modify kubebuilder markers and JSON struct tags. They do NOT change:
- `KAECConfig.toMap()` (lines 208-220) — unchanged
- `VRole.toMap()` (lines 262-280) — unchanged
- `KubeSEConfig.toMap()` (lines 197-204) — unchanged
- `KubeSERole.toMap()` (lines 158-173) — unchanged
- `KubernetesAuthEngineConfig.IsEquivalentToDesiredState()` (lines 76-79) — unchanged
- `KubernetesAuthEngineRole.IsEquivalentToDesiredState()` (lines 83-86) — unchanged
- `KubernetesSecretEngineConfig.IsEquivalentToDesiredState()` (lines 119-123) — unchanged (includes `delete(desiredState, "service_account_jwt")`)
- `KubernetesSecretEngineRole.IsEquivalentToDesiredState()` (lines 77-80) — unchanged
- Any Go code logic

The CRD OpenAPI schema **will change** after `make manifests`. Fields that had explicit `default` values in the CRD YAML will lose them (R1 removals). No new Enum markers are added in this story. This is a schema-level change with no Go code change.

### Impact on Existing Tests

**Unit tests (`api/v1alpha1/kubernetesauthengineconfig_test.go`):**
- `TestKAECConfigToMap` — constructs `KAECConfig` with explicit field values → **unaffected**
- `TestKAECConfigToMapUnexportedTokenReviewerJWT` — tests unexported field → **unaffected**
- `TestKubernetesAuthEngineConfigIsEquivalent*` tests — **unaffected** (annotation changes don't change Go runtime behavior)
- `TestKubernetesAuthEngineConfigIsDeletable` — **unaffected**
- `TestKubernetesAuthEngineConfigConditions` — **unaffected**
- `TestKubernetesAuthEngineConfig_PrepareInternalValues_*` — **unaffected**

**Unit tests (`api/v1alpha1/kubernetesauthenginerole_test.go`):**
- `TestVRoleToMap` — constructs `VRole` with explicit field values → **unaffected**
- `TestVRoleToMapAudienceNil` / `TestVRoleToMapAudienceSet` — **unaffected**
- `TestVRoleToMapUnexportedNamespaces` — **unaffected**
- `TestKubernetesAuthEngineRoleIsEquivalent*` tests — **unaffected**
- `TestKubernetesAuthEngineRoleIsDeletable` — **unaffected**
- `TestKubernetesAuthEngineRoleConditions` — **unaffected**
- `TestKubernetesAuthEngineRole_PrepareInternalValues_*` — **unaffected**

**Unit tests (`api/v1alpha1/kubernetessecretengineconfig_test.go`):**
- `TestKubeSEConfigToMap` — constructs `KubeSEConfig` with explicit field values → **unaffected**
- `TestKubernetesSecretEngineConfigIsEquivalent*` tests — **unaffected**
- `TestKubernetesSecretEngineConfigIsDeletable` — **unaffected**
- `TestKubernetesSecretEngineConfigConditions` — **unaffected**
- `TestKubernetesSecretEngineConfig_PrepareInternalValues_*` — **unaffected**

**Unit tests (`api/v1alpha1/kubernetessecretenginerole_test.go`):**
- `TestKubeSERoleToMap` — constructs `KubeSERole` with explicit field values → **unaffected**
- `TestKubernetesSecretEngineRoleIsEquivalent*` tests — **unaffected**
- `TestKubernetesSecretEngineRoleIsDeletable` — **unaffected**
- `TestKubernetesSecretEngineRoleConditions` — **unaffected**

**Integration tests (`controllers/kubernetesauthengine_controller_test.go`):**
- Config fixture (`test/kubernetesauthengine/test-kube-auth-config.yaml`) sets `kubernetesHost`, `disableLocalCAJWT`, `useOperatorPodCA` explicitly — does NOT rely on any CRD defaults being removed
- Role fixture (`test/kubernetesauthengine/test-kube-auth-role.yaml`) does NOT set `tokenTTL`, `tokenMaxTTL`, `tokenType`, `aliasNameSource` — after R1, zero-value defaults simply won't be server-applied (Go zero value is identical); after R2, `aliasNameSource` defaults to `"serviceaccount_uid"` and `tokenType` defaults to `"default"` (both valid Enum values, always serialized)
- Role-selector fixture (`test/kubernetesauthengine/test-kube-auth-role-selector.yaml`) — same pattern, no issues
- **No fixture relies on server-side defaulting for any R1 field being modified**

**Integration tests (`controllers/kubernetessecretengine_controller_test.go`):**
- Config fixture (`test/kubernetessecretengine/test-kubese-config.yaml`) sets `kubernetesHost` explicitly, does NOT set `disableLocalCAJWT` — after R1, field defaults to Go zero (`false`), same as before
- Role fixture (`test/kubernetessecretengine/test-kubese-role.yaml`) sets `kubernetesRoleType: "ClusterRole"` explicitly — valid Enum value, no issue. Does NOT set `defaultTTL`/`maxTTL` — after R1, zero-duration default is removed but Go zero is equivalent
- **No fixture relies on server-side defaulting for any R1 field being modified**

### Critical Warnings

1. **Do NOT modify `toMap()` or any Go logic.** This is purely an annotation + JSON tag refactor.
2. **Run `make manifests generate` after changes.** This regenerates CRDs in `config/crd/bases/`. The diff will show removed `default:` entries in the OpenAPI schema.
3. **Run `make generate`** to regenerate `zz_generated.deepcopy.go` (struct tag changes may affect generated code).
4. **Note the colon syntax variant:** Several markers in `kubernetesauthenginerole_types.go` use `+kubebuilder:default:=0` (with colon-equals) instead of `+kubebuilder:default=0`. Both are valid kubebuilder marker syntax. Remove the entire line regardless of syntax variant.
5. **`UseOperatorPodCA` keeps its non-zero default `true`.** After R2 fix (removing omitempty), the field always serializes. Test fixture explicitly sets `false` — this overrides the default and is stored correctly.
6. **`DefaultTTL`/`MaxTTL` are `metav1.Duration` type, not `int64`.** The `+kubebuilder:default="0s"` is a zero-duration default — remove it per R1. The Go zero value `metav1.Duration{}` is equivalent to `0s`.
7. **Token fields in `VRole` are `int`, not `int64`.** Unlike LDAP which uses `int64`, the Kubernetes auth role uses `int`. This is the existing type — do not change it.
8. **Test fixture review:** After changes, verify that YAML fixtures in `test/kubernetesauthengine/` and `test/kubernetessecretengine/` still apply without validation errors.

### Pattern for Bool Fields (R1) — All 4 Already Have `omitempty`

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=false
DisableISSValidation bool `json:"disableISSValidation,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
DisableISSValidation bool `json:"disableISSValidation,omitempty"`
```

### Pattern for Int Fields with `0` Default (R1) — All Already Have `omitempty`

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default:=0
TokenTTL int `json:"tokenTTL,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
TokenTTL int `json:"tokenTTL,omitempty"`
```

### Pattern for Duration Fields with `"0s"` Default (R1)

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

### Pattern for Non-Zero Default Fields (R2) — Remove `omitempty`

Before:
```go
// +kubebuilder:validation:Required
// +kubebuilder:default="https://kubernetes.default.svc:443"
KubernetesHost string `json:"kubernetesHost,omitempty"`
```

After:
```go
// +kubebuilder:validation:Required
// +kubebuilder:default="https://kubernetes.default.svc:443"
KubernetesHost string `json:"kubernetesHost"`
```

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/kubernetesauthengineconfig_types.go` | Modified | Remove 2 R1 markers, fix 2 R2 JSON tags |
| 2 | `api/v1alpha1/kubernetesauthenginerole_types.go` | Modified | Remove 6 R1 markers, fix 2 R2 JSON tags |
| 3 | `api/v1alpha1/kubernetessecretengineconfig_types.go` | Modified | Remove 1 R1 marker |
| 4 | `api/v1alpha1/kubernetessecretenginerole_types.go` | Modified | Remove 2 R1 markers, fix 1 R2 JSON tag |
| 5 | `config/crd/bases/redhatcop.redhat.io_kubernetesauthengineconfigs.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 6 | `config/crd/bases/redhatcop.redhat.io_kubernetesauthengineroles.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 7 | `config/crd/bases/redhatcop.redhat.io_kubernetessecretengineconfigs.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 8 | `config/crd/bases/redhatcop.redhat.io_kubernetessecretengineroles.yaml` | Regenerated | CRD schema updated by `make manifests` |

### Project Structure Notes

- CRD types live in `api/v1alpha1/` — all 4 files are in this directory
- CRD schemas are regenerated into `config/crd/bases/` by `make manifests`
- DeepCopy is regenerated into `api/v1alpha1/zz_generated.deepcopy.go` by `make generate`
- Test fixtures in `test/kubernetesauthengine/` and `test/kubernetessecretengine/` — verify they pass post-change
- Integration test files: `controllers/kubernetesauthengine_controller_test.go`, `controllers/kubernetessecretengine_controller_test.go`
- Unit test files: `api/v1alpha1/kubernetesauthengineconfig_test.go`, `api/v1alpha1/kubernetesauthenginerole_test.go`, `api/v1alpha1/kubernetessecretengineconfig_test.go`, `api/v1alpha1/kubernetessecretenginerole_test.go`

### References

- [Source: _bmad-output/project-context.md#CRD Field Default & Validation Rules] — Rules 1-6 governing annotation behavior
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.5.3] — Epic story definition and acceptance criteria
- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go:121-163] — `KAECConfig` struct with all field annotations
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go:145-213] — `VRole` struct with all field annotations
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go:180-195] — `KubeSEConfig` struct with all field annotations
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go:98-156] — `KubeSERole` struct with all field annotations
- [Source: api/v1alpha1/kubernetesauthengineconfig_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/kubernetesauthenginerole_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/kubernetessecretengineconfig_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/kubernetessecretenginerole_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: controllers/kubernetesauthengine_controller_test.go] — Kubernetes auth integration tests (must pass post-change)
- [Source: controllers/kubernetessecretengine_controller_test.go] — Kubernetes secret engine integration tests (must pass post-change)
- [Source: test/kubernetesauthengine/] — Test fixtures (verify post-change)
- [Source: test/kubernetessecretengine/] — Test fixtures (verify post-change)

### Previous Story Intelligence

**From Story 7.5.1 (LDAP Auth Engine Types — Annotation Refactor):**
- Story 7.5.1 is ready-for-dev (not yet implemented) — same pattern applies
- Established patterns: R1 (remove zero-value defaults), R2 (remove omitempty from non-zero defaults)
- `TokenType` Enum defined as `{"service","batch","default","default-service","default-batch"}` — already present on `VRole.TokenType`, no action needed
- Bool fields that lose `+kubebuilder:default=false` keep existing `omitempty` — same pattern here (all R1 bool fields already have `omitempty`)

**From Story 7.5.2 (JWT/OIDC Auth Engine Types — Annotation Refactor):**
- Story 7.5.2 is ready-for-dev — same pattern applies
- Confirms: R1 fields with existing `omitempty` only need default removal (no JSON tag change)
- Confirms: R2 fields with non-zero defaults need `omitempty` removal from JSON tag only

**Key difference from Stories 7.5.1 and 7.5.2:**
- This story has NO new Enum markers to add — all enums already exist on the Kubernetes types
- Several R1 fields already have `omitempty` in their JSON tags (unlike LDAP/JWT where some bool fields lacked `omitempty`)
- Uses `int` type for token fields (not `int64` like LDAP) — existing type, do not change
- Uses `metav1.Duration` for TTL fields in `KubeSERole` — a type not seen in Stories 7.5.1/7.5.2
- Smallest scope of the first 3 stories: 16 total field changes (vs ~27 LDAP, ~29 JWT/OIDC)

**From Epic 7 Retrospective:**
- Codebase is stable after hardening epic
- Coverage is strong for Kubernetes types (unit tests + integration tests both exist)
- Story 4.1 specifically tested Kubernetes auth engine config and role — those tests serve as regression safety net
- Story 5.3 tested Kubernetes secret engine types — additional regression coverage

### Git Intelligence

- Latest commit: `44cad20` (Bmad epic 7 squash merge)
- Branch is clean on main
- No pending changes that could conflict with this annotation refactor
- All CI checks passing on main

## Dev Agent Record

### Agent Model Used

Claude Opus 4 (claude-sonnet-4-20250514)

### Debug Log References

- Integration test run 1 failed: R2 `omitempty` removal on `aliasNameSource`/`tokenType` caused empty strings to bypass CRD defaulting and trigger Enum validation rejection across all KubernetesAuthEngineRole test fixtures
- Integration test run 2 failed: infrastructure flakiness (namespace vault-admin stuck terminating from run 1)
- Integration test run 3: all 90 tests passed (575.815s, 54.0% coverage)

### Completion Notes List

- R1: Removed 11 redundant zero-value `+kubebuilder:default` markers across 4 type files (3 bool false, 6 int 0, 2 Duration "0s")
- R2: Removed `omitempty` from 5 non-zero default fields (kubernetesHost, useOperatorPodCA, aliasNameSource, tokenType, kubernetesRoleType) so they always serialize
- CRD schemas regenerated: `default:` entries removed for R1 fields; `kubernetesHost` now in CRD `required` list
- Updated 13 YAML test fixtures to explicitly set `aliasNameSource: "serviceaccount_uid"` and `tokenType: "default"` — required because removing `omitempty` causes Go marshalling to send empty strings which bypass CRD server-side defaulting
- Updated config sample for consistency
- All unit tests pass (make test), all 90 integration tests pass (make integration)
- No Go logic changes — purely annotation + JSON tag + test fixture refactor

### File List

- api/v1alpha1/kubernetesauthengineconfig_types.go (modified — R1: 2 defaults removed, R2: 2 omitempty removed)
- api/v1alpha1/kubernetesauthenginerole_types.go (modified — R1: 6 defaults removed, R2: 2 omitempty removed)
- api/v1alpha1/kubernetessecretengineconfig_types.go (modified — R1: 1 default removed)
- api/v1alpha1/kubernetessecretenginerole_types.go (modified — R1: 2 defaults removed, R2: 1 omitempty removed)
- config/crd/bases/redhatcop.redhat.io_kubernetesauthengineconfigs.yaml (regenerated)
- config/crd/bases/redhatcop.redhat.io_kubernetesauthengineroles.yaml (regenerated)
- config/crd/bases/redhatcop.redhat.io_kubernetessecretengineconfigs.yaml (regenerated)
- config/crd/bases/redhatcop.redhat.io_kubernetessecretengineroles.yaml (regenerated)
- test/kubernetesauthengine/test-kube-auth-role.yaml (modified — added aliasNameSource, tokenType)
- test/kubernetesauthengine/test-kube-auth-role-selector.yaml (modified — added aliasNameSource, tokenType)
- test/database-engine-admin-role.yaml (modified — added aliasNameSource, tokenType)
- test/kv-engine-admin-role.yaml (modified — added aliasNameSource, tokenType)
- test/kube-auth-engine-role.yaml (modified — added aliasNameSource, tokenType)
- test/secret-writer-role.yaml (modified — added aliasNameSource, tokenType)
- test/rabbitmq-engine-admin-role.yaml (modified — added aliasNameSource, tokenType)
- test/databasesecretengine/database-secret-engine-auth-role.yaml (modified — added aliasNameSource, tokenType)
- test/pkisecretengine/pki-secret-engine-kube-auth-role.yaml (modified — added aliasNameSource, tokenType)
- test/vaultsecret/kubernetesauthenginerole-secret-reader.yaml (modified — added aliasNameSource, tokenType)
- test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml (modified — added aliasNameSource, tokenType)
- test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml (modified — added aliasNameSource, tokenType)
- test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml (modified — added aliasNameSource, tokenType)
- config/samples/redhatcop_v1alpha1_kubernetesauthenginerole.yaml (modified — added aliasNameSource, tokenType)

### Review Findings

- [x] [Review][Decision] Preserve server-default semantics for typed clients after removing `omitempty` from non-zero default fields — dismissed by user as an intentional contract change for programmatic clients. Removing `omitempty` from `kubernetesHost`, `useOperatorPodCA`, `aliasNameSource`, `tokenType`, and `kubernetesRoleType` now causes zero-valued typed Go objects to serialize explicit empty/false values instead of omitting the keys. In this repo, integration fixtures are loaded through the typed decoder in `controllers/controllertestutils/decoder.go` and then created with `k8sIntegrationClient.Create(...)`, which is why the role fixtures had to be patched to set `aliasNameSource` and `tokenType` explicitly. The role mutating webhooks do not backfill defaults, and the auth config mutating webhook only reacts to `UseOperatorPodCA` after that value has already been decoded, so programmatic clients that previously relied on CRD defaulting now either fail validation for enum-backed role fields or silently change auth-config behavior.
