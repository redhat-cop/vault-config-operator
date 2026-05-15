# Story 7.5.4: Azure & GCP Auth/Secret Engine Types — Annotation Refactor

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the Azure and GCP engine types to follow the CRD Field Default & Validation Rules,
So that these cloud provider types have consistent annotation patterns.

## Acceptance Criteria

1. **Given** zero-value defaults across Azure/GCP role and config types **When** `+kubebuilder:default` markers removed and `omitempty` ensured **Then** `make manifests generate test` passes
2. **Given** `Environment` fields (both Azure configs) with non-zero default `"AzurePublicCloud"` and `omitempty` **When** `omitempty` removed and `+kubebuilder:validation:Enum` added **Then** validated at admission and always serialized
3. **Given** GCP `IAMalias`, `IAMmetadata`, `GCEalias`, `GCEmetadata` fields with non-zero defaults and `omitempty` **When** `omitempty` removed **Then** fields always present in serialized JSON
4. **Given** `GCEalias` accepts `instance_id` or `role_id` **When** `+kubebuilder:validation:Enum` added **Then** invalid values rejected at admission
5. **Given** all changes **When** `make manifests generate fmt vet test` passes **Then** no regressions

## Tasks / Subtasks

- [x] Task 1: Refactor `azureauthengineconfig_types.go` — `AzureConfig` struct (AC: 1, 2)
  - [x] 1.1: Remove `+kubebuilder:default=""` from `ClientID` (line 106); JSON tag already has `omitempty`
  - [x] 1.2: Remove `omitempty` from `Environment` JSON tag (line 100): `json:"environment,omitempty"` → `json:"environment"`
  - [x] 1.3: Add `// +kubebuilder:validation:Enum={"AzurePublicCloud","AzureUSGovernmentCloud","AzureChinaCloud","AzureGermanCloud"}` to `Environment`
- [x] Task 2: Refactor `azureauthenginerole_types.go` — `AzureRole` struct (AC: 1)
  - [x] 2.1: Remove `+kubebuilder:default=""` from `TokenTTL` (line 122); JSON tag already has `omitempty`
  - [x] 2.2: Remove `+kubebuilder:default=""` from `TokenMaxTTL` (line 128); JSON tag already has `omitempty`
  - [x] 2.3: Remove `+kubebuilder:default=""` from `TokenExplicitMaxTTL` (line 156); JSON tag already has `omitempty`
  - [x] 2.4: Remove `+kubebuilder:default=false` from `TokenNoDefaultPolicy` (line 161); add `omitempty` to JSON tag: `json:"tokenNoDefaultPolicy"` → `json:"tokenNoDefaultPolicy,omitempty"`
  - [x] 2.5: Remove `+kubebuilder:default=0` from `TokenNumUses` (line 167); add `omitempty` to JSON tag: `json:"tokenNumUses"` → `json:"tokenNumUses,omitempty"`
  - [x] 2.6: Remove `+kubebuilder:default=0` from `TokenPeriod` (line 172); add `omitempty` to JSON tag: `json:"tokenPeriod"` → `json:"tokenPeriod,omitempty"`
  - [x] 2.7: Remove `+kubebuilder:default=""` from `TokenType` (line 180); JSON tag already has `omitempty`
- [x] Task 3: Refactor `azuresecretengineconfig_types.go` — `AzureSEConfig` struct (AC: 1, 2)
  - [x] 3.1: Remove `+kubebuilder:default=""` from `ClientID` (line 100); JSON tag already has `omitempty`
  - [x] 3.2: Remove `+kubebuilder:default=""` from `PasswordPolicy` (line 111); JSON tag already has `omitempty`
  - [x] 3.3: Remove `omitempty` from `Environment` JSON tag (line 107): `json:"environment,omitempty"` → `json:"environment"`
  - [x] 3.4: Remove `omitempty` from `RootPasswordTTL` JSON tag (line 117): `json:"rootPasswordTTL,omitempty"` → `json:"rootPasswordTTL"`
  - [x] 3.5: Add `// +kubebuilder:validation:Enum={"AzurePublicCloud","AzureUSGovernmentCloud","AzureChinaCloud","AzureGermanCloud"}` to `Environment`
- [x] Task 4: Refactor `azuresecretenginerole_types.go` — `AzureSERole` struct (AC: 1)
  - [x] 4.1: Remove `+kubebuilder:default=""` from `AzureRoles` (line 86); JSON tag already has `omitempty`
  - [x] 4.2: Remove `+kubebuilder:default=""` from `AzureGroups` (line 92); JSON tag already has `omitempty`
  - [x] 4.3: Remove `+kubebuilder:default=""` from `ApplicationObjectID` (line 98); JSON tag already has `omitempty`
  - [x] 4.4: Remove `+kubebuilder:default=false` from `PersistApp` (line 104); add `omitempty` to JSON tag: `json:"persistApp"` → `json:"persistApp,omitempty"`
  - [x] 4.5: Remove `+kubebuilder:default=""` from `TTL` (line 110); JSON tag already has `omitempty`
  - [x] 4.6: Remove `+kubebuilder:default=""` from `MaxTTL` (line 116); JSON tag already has `omitempty`
  - [x] 4.7: Remove `+kubebuilder:default=""` from `PermanentlyDelete` (line 122); JSON tag already has `omitempty`
  - [x] 4.8: Remove `+kubebuilder:default=""` from `SignInAudience` (line 128); JSON tag already has `omitempty`
  - [x] 4.9: Remove `+kubebuilder:default=""` from `Tags` (line 133); JSON tag already has `omitempty`
- [x] Task 5: Refactor `gcpauthengineconfig_types.go` — `GCPConfig` struct (AC: 1, 3, 4)
  - [x] 5.1: Remove `+kubebuilder:default=""` from `ServiceAccount` (line 92); JSON tag already has `omitempty`
  - [x] 5.2: Remove `omitempty` from `IAMalias` JSON tag (line 100): `json:"IAMalias,omitempty"` → `json:"IAMalias"`
  - [x] 5.3: Remove `omitempty` from `IAMmetadata` JSON tag (line 110): `json:"IAMmetadata,omitempty"` → `json:"IAMmetadata"`
  - [x] 5.4: Remove `omitempty` from `GCEalias` JSON tag (line 116): `json:"GCEalias,omitempty"` → `json:"GCEalias"`
  - [x] 5.5: Remove `omitempty` from `GCEmetadata` JSON tag (line 125): `json:"GCEmetadata,omitempty"` → `json:"GCEmetadata"`
  - [x] 5.6: Add `// +kubebuilder:validation:Enum={"instance_id","role_id"}` to `GCEalias`
- [x] Task 6: Refactor `gcpauthenginerole_types.go` — `GCPRole` struct (AC: 1)
  - [x] 6.1: Remove `+kubebuilder:default=false` from `AddGroupAliases` (line 100); add `omitempty` to JSON tag: `json:"addGroupAliases"` → `json:"addGroupAliases,omitempty"`
  - [x] 6.2: Remove `+kubebuilder:default=""` from `TokenTTL` (line 105); JSON tag already has `omitempty`
  - [x] 6.3: Remove `+kubebuilder:default=""` from `TokenMaxTTL` (line 110); JSON tag already has `omitempty`
  - [x] 6.4: Remove `+kubebuilder:default=""` from `TokenExplicitMaxTTL` (line 137); JSON tag already has `omitempty`
  - [x] 6.5: Remove `+kubebuilder:default=false` from `TokenNoDefaultPolicy` (line 142); add `omitempty` to JSON tag: `json:"tokenNoDefaultPolicy"` → `json:"tokenNoDefaultPolicy,omitempty"`
  - [x] 6.6: Remove `+kubebuilder:default=0` from `TokenNumUses` (line 148); add `omitempty` to JSON tag: `json:"tokenNumUses"` → `json:"tokenNumUses,omitempty"`
  - [x] 6.7: Remove `+kubebuilder:default=0` from `TokenPeriod` (line 153); add `omitempty` to JSON tag: `json:"tokenPeriod"` → `json:"tokenPeriod,omitempty"`
  - [x] 6.8: Remove `+kubebuilder:default=""` from `TokenType` (line 161); JSON tag already has `omitempty`
  - [x] 6.9: Remove `+kubebuilder:default=""` from `MaxJWTExp` (line 171); JSON tag already has `omitempty`
  - [x] 6.10: Remove `+kubebuilder:default=false` from `AllowGCEInference` (line 176); add `omitempty` to JSON tag: `json:"allowGCEInference"` → `json:"allowGCEInference,omitempty"`
- [x] Task 7: Run `make manifests generate fmt vet test` (AC: 1, 2, 3, 4, 5)

## Dev Notes

### Scope: 6 Files, ~36 Field Changes

| File | R1 Removals | R2 Fixes | Enum Additions |
|------|-------------|----------|----------------|
| `api/v1alpha1/azureauthengineconfig_types.go` | 1 field | 1 field | 1 field |
| `api/v1alpha1/azureauthenginerole_types.go` | 7 fields | 0 | 0 |
| `api/v1alpha1/azuresecretengineconfig_types.go` | 2 fields | 2 fields | 1 field |
| `api/v1alpha1/azuresecretenginerole_types.go` | 9 fields | 0 | 0 |
| `api/v1alpha1/gcpauthengineconfig_types.go` | 1 field | 4 fields | 1 field |
| `api/v1alpha1/gcpauthenginerole_types.go` | 10 fields | 0 | 0 |

### No Integration Tests Exist for These Types

Azure and GCP are **cloud-provider types** that fall under the "Skip it" category per the integration test infrastructure philosophy in `project-context.md`:

> "For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only."

There are NO test fixtures in `test/` and NO integration test files for these types. **Only unit tests exist** (in `api/v1alpha1/`). Task 7 runs `make test` (unit tests) — no `make integration` is needed for this story.

### Detailed Field Change Table — `AzureConfig` struct (lines 84-128)

**Rule 1 — Remove redundant zero-value `kubebuilder:default` (already has `omitempty`):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `ClientID` | string | `""` | `json:"clientID,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults + add Enum:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `Environment` | string | `"AzurePublicCloud"` | `json:"environment,omitempty"` | Remove `omitempty`, add Enum |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `TenantID` | string | none | `json:"tenantID"` | Required, no default |
| `Resource` | string | none | `json:"resource"` | Required, no default |
| `MaxRetries` | int64 | `3` | `json:"maxRetries"` | Non-zero default, no `omitempty` |
| `MaxRetryDelay` | int64 | `60` | `json:"maxRetryDelay"` | Non-zero default, no `omitempty` |
| `RetryDelay` | int64 | `4` | `json:"retryDelay"` | Non-zero default, no `omitempty` |

### Detailed Field Change Table — `AzureRole` struct (lines 77-182)

**Rule 1 — Remove redundant zero-value `kubebuilder:default`; ensure `omitempty`:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `TokenTTL` | string | `""` | `json:"tokenTTL,omitempty"` | Remove default only |
| `TokenMaxTTL` | string | `""` | `json:"tokenMaxTTL,omitempty"` | Remove default only |
| `TokenExplicitMaxTTL` | string | `""` | `json:"tokenExplicitMaxTTL,omitempty"` | Remove default only |
| `TokenNoDefaultPolicy` | bool | `false` | `json:"tokenNoDefaultPolicy"` | Remove default, add `omitempty` |
| `TokenNumUses` | int64 | `0` | `json:"tokenNumUses"` | Remove default, add `omitempty` |
| `TokenPeriod` | int64 | `0` | `json:"tokenPeriod"` | Remove default, add `omitempty` |
| `TokenType` | string | `""` | `json:"tokenType,omitempty"` | Remove default only |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `Name` | string | none | `json:"name"` | Required, no default |
| `BoundServicePrincipalIDs` | []string | none | `json:"boundServicePrincipalIDs,omitempty"` | Optional, no default |
| `BoundGroupIDs` | []string | none | `json:"boundGroupIDs,omitempty"` | Optional, no default |
| `BoundLocations` | []string | none | `json:"boundLocations,omitempty"` | Optional, no default |
| `BoundSubscriptionIDs` | []string | none | `json:"boundSubscriptionIDs,omitempty"` | Optional, no default |
| `BoundResourceGroups` | []string | none | `json:"boundResourceGroups,omitempty"` | Optional, no default |
| `BoundScaleSets` | []string | none | `json:"boundScaleSets,omitempty"` | Optional, no default |
| `TokenPolicies` | []string | none | `json:"tokenPolicies,omitempty"` | Optional, no default |
| `Policies` | []string | none | `json:"policies,omitempty"` | Optional (deprecated), no default |
| `TokenBoundCIDRs` | []string | none | `json:"tokenBoundCIDRs,omitempty"` | Optional, no default |

### Detailed Field Change Table — `AzureSEConfig` struct (lines 86-122)

**Rule 1 — Remove redundant zero-value `kubebuilder:default` (already have `omitempty`):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `ClientID` | string | `""` | `json:"clientID,omitempty"` | Remove default only |
| `PasswordPolicy` | string | `""` | `json:"passwordPolicy,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults (+ Enum for Environment):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `Environment` | string | `"AzurePublicCloud"` | `json:"environment,omitempty"` | Remove `omitempty`, add Enum |
| `RootPasswordTTL` | string | `"182d"` | `json:"rootPasswordTTL,omitempty"` | Remove `omitempty` |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `SubscriptionID` | string | none | `json:"subscriptionID"` | Required, no default |
| `TenantID` | string | none | `json:"tenantID"` | Required, no default |

### Detailed Field Change Table — `AzureSERole` struct (lines 82-135)

**Rule 1 — Remove redundant zero-value `kubebuilder:default`; ensure `omitempty`:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `AzureRoles` | string | `""` | `json:"azureRoles,omitempty"` | Remove default only |
| `AzureGroups` | string | `""` | `json:"azureGroups,omitempty"` | Remove default only |
| `ApplicationObjectID` | string | `""` | `json:"applicationObjectID,omitempty"` | Remove default only |
| `PersistApp` | bool | `false` | `json:"persistApp"` | Remove default, add `omitempty` |
| `TTL` | string | `""` | `json:"TTL,omitempty"` | Remove default only |
| `MaxTTL` | string | `""` | `json:"maxTTL,omitempty"` | Remove default only |
| `PermanentlyDelete` | string | `""` | `json:"permanentlyDelete,omitempty"` | Remove default only |
| `SignInAudience` | string | `""` | `json:"signInAudience,omitempty"` | Remove default only |
| `Tags` | string | `""` | `json:"tags,omitempty"` | Remove default only |

### Detailed Field Change Table — `GCPConfig` struct (lines 85-141)

**Rule 1 — Remove redundant zero-value `kubebuilder:default` (already has `omitempty`):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `ServiceAccount` | string | `""` | `json:"serviceAccount,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults (+ Enum for GCEalias):**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `IAMalias` | string | `"default"` | `json:"IAMalias,omitempty"` | Remove `omitempty` |
| `IAMmetadata` | string | `"default"` | `json:"IAMmetadata,omitempty"` | Remove `omitempty` |
| `GCEalias` | string | `"role_id"` | `json:"GCEalias,omitempty"` | Remove `omitempty`, add Enum |
| `GCEmetadata` | string | `"default"` | `json:"GCEmetadata,omitempty"` | Remove `omitempty` |

**Intentionally excluded from scope:**

| Field | Type | Default | JSON Tag | Why Excluded |
|-------|------|---------|----------|-------------|
| `CustomEndpoint` | *apiextensionsv1.JSON | `{}` | `json:"customEndpoint,omitempty"` | Pointer-to-JSON with empty-object default; `omitempty` checks nil (pointer zero) while `{}` is non-nil. Current behavior is correct: default prevents nil, omitempty only omits explicit nil. Edge case excluded per epic scope. |

### Detailed Field Change Table — `GCPRole` struct (lines 77-209)

**Rule 1 — Remove redundant zero-value `kubebuilder:default`; ensure `omitempty`:**

| Field | Type | Current Default | Current JSON Tag | Change Required |
|-------|------|-----------------|-----------------|-----------------|
| `AddGroupAliases` | bool | `false` | `json:"addGroupAliases"` | Remove default, add `omitempty` |
| `TokenTTL` | string | `""` | `json:"tokenTTL,omitempty"` | Remove default only |
| `TokenMaxTTL` | string | `""` | `json:"tokenMaxTTL,omitempty"` | Remove default only |
| `TokenExplicitMaxTTL` | string | `""` | `json:"tokenExplicitMaxTTL,omitempty"` | Remove default only |
| `TokenNoDefaultPolicy` | bool | `false` | `json:"tokenNoDefaultPolicy"` | Remove default, add `omitempty` |
| `TokenNumUses` | int64 | `0` | `json:"tokenNumUses"` | Remove default, add `omitempty` |
| `TokenPeriod` | int64 | `0` | `json:"tokenPeriod"` | Remove default, add `omitempty` |
| `TokenType` | string | `""` | `json:"tokenType,omitempty"` | Remove default only |
| `MaxJWTExp` | string | `""` | `json:"maxJWTExp,omitempty"` | Remove default only |
| `AllowGCEInference` | bool | `false` | `json:"allowGCEInference"` | Remove default, add `omitempty` |

**Intentionally excluded from scope:**

| Field | Type | Default | JSON Tag | Why Excluded |
|-------|------|---------|----------|-------------|
| `BoundServiceAccounts` | []string | `{}` | `json:"boundServiceAccounts,omitempty"` | Slice with empty-array default; Go zero for slices is nil, not `[]string{}`. Removing default would change nil vs empty-slice semantics in `toMap()`. Excluded per epic scope. |
| `BoundProjects` | []string | `{}` | `json:"boundProjects,omitempty"` | Same rationale as BoundServiceAccounts. |

**Already compliant (no change needed):**

| Field | Type | Default | JSON Tag | Why Compliant |
|-------|------|---------|----------|--------------|
| `Name` | string | none | `json:"name"` | Required, no default |
| `Type` | string | none | `json:"type"` | Required, no default |
| `TokenPolicies` | []string | none | `json:"tokenPolicies,omitempty"` | Optional, no default |
| `Policies` | []string | none | `json:"policies,omitempty"` | Optional (deprecated), no default |
| `TokenBoundCIDRs` | []string | none | `json:"tokenBoundCIDRs,omitempty"` | Optional, no default |
| `BoundZones` | []string | none | `json:"boundZones,omitempty"` | Optional, no default |
| `BoundRegions` | []string | none | `json:"boundRegions,omitempty"` | Optional, no default |
| `BoundInstanceGroups` | []string | none | `json:"boundInstanceGroups,omitempty"` | Optional, no default |
| `BoundLabels` | []string | none | `json:"boundLabels,omitempty"` | Optional, no default |

### Enum Marker Details

**`Environment` (both `AzureConfig` and `AzureSEConfig`):** Vault Azure docs define 4 accepted cloud environments. Add to both files:
```
// +kubebuilder:validation:Enum={"AzurePublicCloud","AzureUSGovernmentCloud","AzureChinaCloud","AzureGermanCloud"}
```

**`GCEalias`:** Vault GCP auth config docs accept exactly 2 values. Add:
```
// +kubebuilder:validation:Enum={"instance_id","role_id"}
```

**Note on `IAMalias`:** The field comment says "Must be either unique_id or role_id" but the current default is `"default"`. This is a pre-existing discrepancy with Vault API behavior (Vault may interpret `"default"` as `"role_id"`). Do NOT add an Enum marker to `IAMalias` in this story — it would break the existing default value. This is out of scope for annotation refactor; a future story can investigate and correct the default value.

**Note on `IAMmetadata` and `GCEmetadata`:** These fields accept comma-separated field names (e.g., `"project_id,role,service_account_id"`) plus the special value `"default"` and empty string `""`. They are NOT good Enum candidates because the value set is combinatorial. Do NOT add Enum markers.

### Impact on `toMap()` and `IsEquivalentToDesiredState()`

These changes are **annotation-only** — they modify kubebuilder markers and JSON struct tags. They do NOT change:
- `AzureConfig.toMap()` (lines 260-272) — unchanged
- `AzureRole.toMap()` (lines 240-261) — unchanged
- `AzureSEConfig.toMap()` (lines 259-270) — unchanged
- `AzureSERole.toMap()` (lines 196-209) — unchanged
- `GCPConfig.toMap()` (lines 273-283) — unchanged
- `GCPRole.toMap()` (lines 267-292) — unchanged
- `AzureAuthEngineConfig.IsEquivalentToDesiredState()` (lines 162-165) — unchanged
- `AzureAuthEngineRole.IsEquivalentToDesiredState()` (lines 219-222) — unchanged
- `AzureSecretEngineConfig.IsEquivalentToDesiredState()` (lines 159-162) — unchanged
- `AzureSecretEngineRole.IsEquivalentToDesiredState()` (lines 167-170) — unchanged
- `GCPAuthEngineConfig.IsEquivalentToDesiredState()` (lines 175-178) — unchanged
- `GCPAuthEngineRole.IsEquivalentToDesiredState()` (lines 246-249) — unchanged
- Any Go code logic

The CRD OpenAPI schema **will change** after `make manifests`. Fields that had explicit `default` values in the CRD YAML will lose them (R1 removals), and `enum` constraints will be added for `Environment` and `GCEalias`. This is a schema-level change with no Go code change.

### Impact on Existing Tests

**Unit tests (`api/v1alpha1/azureauthengineconfig_test.go`):**
- `TestAzureConfigToMap` — constructs `AzureConfig` with explicit field values → **unaffected**
- `TestAzureAuthEngineConfigIsEquivalent*` tests — **unaffected** (annotation changes don't change Go runtime behavior)
- `TestAzureAuthEngineConfigIsDeletable` — **unaffected**
- `TestAzureAuthEngineConfigConditions` — **unaffected**
- `TestAzureAuthEngineConfig_PrepareInternalValues_*` — **unaffected**

**Unit tests (`api/v1alpha1/azureauthenginerole_test.go`):**
- `TestAzureRoleToMap` — constructs `AzureRole` with explicit field values → **unaffected**
- `TestAzureAuthEngineRoleIsEquivalent*` tests — **unaffected**
- `TestAzureAuthEngineRoleIsDeletable` — **unaffected**
- `TestAzureAuthEngineRoleConditions` — **unaffected**

**Unit tests (`api/v1alpha1/azuresecretengineconfig_test.go`):**
- `TestAzureSEConfigToMap` — constructs `AzureSEConfig` with explicit field values → **unaffected**
- `TestAzureSecretEngineConfigIsEquivalent*` tests — **unaffected**
- `TestAzureSecretEngineConfigIsDeletable` — **unaffected**
- `TestAzureSecretEngineConfigConditions` — **unaffected**
- `TestAzureSecretEngineConfig_PrepareInternalValues_*` — **unaffected**

**Unit tests (`api/v1alpha1/azuresecretenginerole_test.go`):**
- `TestAzureSERoleToMap` — constructs `AzureSERole` with explicit field values → **unaffected**
- `TestAzureSecretEngineRoleIsEquivalent*` tests — **unaffected**
- `TestAzureSecretEngineRoleIsDeletable` — **unaffected**
- `TestAzureSecretEngineRoleConditions` — **unaffected**

**Unit tests (`api/v1alpha1/gcpauthengineconfig_test.go`):**
- `TestGCPConfigToMap` — constructs `GCPConfig` with explicit field values → **unaffected**
- `TestGCPAuthEngineConfigIsEquivalent*` tests — **unaffected**
- `TestGCPAuthEngineConfigIsDeletable` — **unaffected**
- `TestGCPAuthEngineConfigConditions` — **unaffected**
- `TestGCPAuthEngineConfig_PrepareInternalValues_*` — **unaffected**

**Unit tests (`api/v1alpha1/gcpauthenginerole_test.go`):**
- `TestGCPRoleToMap` — constructs `GCPRole` with explicit field values → **unaffected**
- `TestGCPAuthEngineRoleIsEquivalent*` tests — **unaffected**
- `TestGCPAuthEngineRoleIsDeletable` — **unaffected**
- `TestGCPAuthEngineRoleConditions` — **unaffected**

**No integration tests exist** for any Azure or GCP types (cloud provider — cannot be deployed in Kind).

### Critical Warnings

1. **Do NOT modify `toMap()` or any Go logic.** This is purely an annotation + JSON tag refactor.
2. **Run `make manifests generate` after changes.** This regenerates CRDs in `config/crd/bases/`. The diff will show removed `default:` entries and added `enum:` entries in the OpenAPI schema.
3. **Run `make generate`** to regenerate `zz_generated.deepcopy.go` (struct tag changes may affect generated code).
4. **Do NOT add `+kubebuilder:validation:Enum` to `IAMalias`.** The current default `"default"` doesn't match the comment's listed values (`unique_id`, `role_id`). Adding an Enum that excludes `"default"` would break the existing CRD. This is a pre-existing data issue, not an annotation refactor concern.
5. **Do NOT add `+kubebuilder:validation:Enum` to `IAMmetadata` or `GCEmetadata`.** These fields accept free-form comma-separated lists, not discrete enum values.
6. **Do NOT modify `CustomEndpoint`, `BoundServiceAccounts`, or `BoundProjects` defaults.** These have `+kubebuilder:default={}` on pointer/slice types where removing the default would change nil-vs-empty semantics. They are intentionally excluded per epic scope.
7. **Token fields are `int64`, not `int`.** Unlike Kubernetes auth roles (which use `int`), Azure and GCP auth roles use `int64` for `TokenNumUses` and `TokenPeriod`. This is the existing type — do not change it.
8. **`AzureConfig.MaxRetries`, `MaxRetryDelay`, `RetryDelay` are already compliant.** They have non-zero defaults (3, 60, 4) without `omitempty` — no change needed.

### Pattern for String Fields with `""` Default (R1) — Already Have `omitempty`

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=""
TokenTTL string `json:"tokenTTL,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
TokenTTL string `json:"tokenTTL,omitempty"`
```

### Pattern for Bool Fields (R1) — Need `omitempty` Added

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default=false
TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy"`
```

After:
```go
// +kubebuilder:validation:Optional
TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`
```

### Pattern for Int64 Fields with `0` Default (R1) — Need `omitempty` Added

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

### Pattern for Non-Zero Default Fields (R2) — Remove `omitempty`

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="AzurePublicCloud"
Environment string `json:"environment,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="AzurePublicCloud"
// +kubebuilder:validation:Enum={"AzurePublicCloud","AzureUSGovernmentCloud","AzureChinaCloud","AzureGermanCloud"}
Environment string `json:"environment"`
```

### Pattern for Non-Zero Default Fields (R2) — No Enum

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="role_id"
GCEalias string `json:"GCEalias,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="role_id"
// +kubebuilder:validation:Enum={"instance_id","role_id"}
GCEalias string `json:"GCEalias"`
```

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/azureauthengineconfig_types.go` | Modified | Remove 1 R1 marker, fix 1 R2 JSON tag, add 1 Enum |
| 2 | `api/v1alpha1/azureauthenginerole_types.go` | Modified | Remove 7 R1 markers (3 need omitempty added) |
| 3 | `api/v1alpha1/azuresecretengineconfig_types.go` | Modified | Remove 2 R1 markers, fix 2 R2 JSON tags, add 1 Enum |
| 4 | `api/v1alpha1/azuresecretenginerole_types.go` | Modified | Remove 9 R1 markers (1 needs omitempty added) |
| 5 | `api/v1alpha1/gcpauthengineconfig_types.go` | Modified | Remove 1 R1 marker, fix 4 R2 JSON tags, add 1 Enum |
| 6 | `api/v1alpha1/gcpauthenginerole_types.go` | Modified | Remove 10 R1 markers (3 need omitempty added) |
| 7 | `config/crd/bases/redhatcop.redhat.io_azureauthengineconfigs.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 8 | `config/crd/bases/redhatcop.redhat.io_azureauthengineroles.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 9 | `config/crd/bases/redhatcop.redhat.io_azuresecretengineconfigs.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 10 | `config/crd/bases/redhatcop.redhat.io_azuresecretengineroles.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 11 | `config/crd/bases/redhatcop.redhat.io_gcpauthengineconfigs.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 12 | `config/crd/bases/redhatcop.redhat.io_gcpauthengineroles.yaml` | Regenerated | CRD schema updated by `make manifests` |

### Project Structure Notes

- CRD types live in `api/v1alpha1/` — all 6 files are in this directory
- CRD schemas are regenerated into `config/crd/bases/` by `make manifests`
- DeepCopy is regenerated into `api/v1alpha1/zz_generated.deepcopy.go` by `make generate`
- No test fixtures exist in `test/` for Azure or GCP types (cloud provider skip)
- No integration test files exist for these types
- Unit test files: `api/v1alpha1/azureauthengineconfig_test.go`, `api/v1alpha1/azureauthenginerole_test.go`, `api/v1alpha1/azuresecretengineconfig_test.go`, `api/v1alpha1/azuresecretenginerole_test.go`, `api/v1alpha1/gcpauthengineconfig_test.go`, `api/v1alpha1/gcpauthenginerole_test.go`

### References

- [Source: _bmad-output/project-context.md#CRD Field Default & Validation Rules] — Rules 1-6 governing annotation behavior
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.5.4] — Epic story definition and acceptance criteria
- [Source: api/v1alpha1/azureauthengineconfig_types.go:84-128] — `AzureConfig` struct with all field annotations
- [Source: api/v1alpha1/azureauthenginerole_types.go:77-182] — `AzureRole` struct with all field annotations
- [Source: api/v1alpha1/azuresecretengineconfig_types.go:86-122] — `AzureSEConfig` struct with all field annotations
- [Source: api/v1alpha1/azuresecretenginerole_types.go:82-135] — `AzureSERole` struct with all field annotations
- [Source: api/v1alpha1/gcpauthengineconfig_types.go:85-141] — `GCPConfig` struct with all field annotations
- [Source: api/v1alpha1/gcpauthenginerole_types.go:77-209] — `GCPRole` struct with all field annotations
- [Source: api/v1alpha1/azureauthengineconfig_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/azureauthenginerole_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/azuresecretengineconfig_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/azuresecretenginerole_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/gcpauthengineconfig_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/gcpauthenginerole_test.go] — Existing unit tests (unaffected by annotation changes)

### Previous Story Intelligence

**From Story 7.5.1 (LDAP Auth Engine Types — Annotation Refactor):**
- Established patterns: R1 (remove zero-value defaults), R2 (remove omitempty from non-zero defaults)
- Bool fields that lose `+kubebuilder:default=false` and had no `omitempty` → add `omitempty` (same pattern applies here for `TokenNoDefaultPolicy`, `PersistApp`, `AddGroupAliases`, `AllowGCEInference`)
- String fields with `+kubebuilder:default=""` that already have `omitempty` → remove default only (same pattern)
- Int64 fields with `+kubebuilder:default=0` and no `omitempty` → remove default, add `omitempty` (same pattern for `TokenNumUses`, `TokenPeriod`)

**From Story 7.5.2 (JWT/OIDC Auth Engine Types — Annotation Refactor):**
- Confirms: R1 fields with existing `omitempty` only need default removal (no JSON tag change)
- Confirms: R2 fields with non-zero defaults need `omitempty` removal from JSON tag only

**From Story 7.5.3 (Kubernetes Auth & Secret Engine Types — Annotation Refactor):**
- Confirms: No Go code changes needed — annotation-only refactor
- Confirms: `make manifests generate` regenerates CRDs; `make test` validates
- Confirms: existing unit tests are unaffected by annotation changes (tests use explicit values)

**Key differences from Stories 7.5.1-7.5.3:**
- This story spans **6 files** (most in the epic so far) but each file is smaller
- No integration tests exist for Azure/GCP → only `make test` needed (no `make integration`)
- Enum values are well-defined for `Environment` (Azure cloud names) and `GCEalias` (instance_id/role_id)
- `IAMalias` has a pre-existing default/comment discrepancy — do NOT add Enum
- `CustomEndpoint`, `BoundServiceAccounts`, `BoundProjects` have `+kubebuilder:default={}` on pointer/slice types — intentionally excluded

**From Epic 7 Retrospective:**
- Codebase is stable after hardening epic
- Coverage exists for all Azure/GCP types via unit tests in `api/v1alpha1/`
- Stories 1.4 and 1.5 added `toMap()` and `IsEquivalentToDesiredState()` unit tests for these types

### Git Intelligence

- Latest commit: `44cad20` (Bmad epic 7 squash merge)
- Branch is clean on main
- No pending changes that could conflict with this annotation refactor
- All CI checks passing on main

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (Cursor)

### Debug Log References

None — clean execution, no errors encountered.

### Completion Notes List

- All 6 type files refactored: removed 30 redundant `+kubebuilder:default` markers (R1), fixed 7 `omitempty` tags on non-zero default fields (R2), added 3 `+kubebuilder:validation:Enum` markers
- `Environment` fields (both Azure configs): added `Enum={"AzurePublicCloud","AzureUSGovernmentCloud","AzureChinaCloud","AzureGermanCloud"}`, removed `omitempty`
- `GCEalias` field: added `Enum={"instance_id","role_id"}`, removed `omitempty`
- `IAMalias`, `IAMmetadata`, `GCEmetadata`: removed `omitempty` only (non-zero defaults), no Enum added per story spec
- Bool fields (`TokenNoDefaultPolicy`, `PersistApp`, `AddGroupAliases`, `AllowGCEInference`): removed `+kubebuilder:default=false`, added `omitempty`
- Int64 fields (`TokenNumUses`, `TokenPeriod`): removed `+kubebuilder:default=0`, added `omitempty`
- No Go logic changes — annotation-only refactor
- `make manifests generate fmt vet test` passes (unit tests 25.4% coverage, all green)
- `make integration` passes (577s, 54% coverage, all green)
- No `toMap()`, `IsEquivalentToDesiredState()`, or any Go code logic modified
- Excluded fields per story spec: `CustomEndpoint` (pointer-to-JSON), `BoundServiceAccounts`/`BoundProjects` (slice defaults)

### File List

| # | File | Status | Description |
|---|------|--------|-------------|
| 1 | `api/v1alpha1/azureauthengineconfig_types.go` | Modified | R1: removed default from ClientID; R2: removed omitempty from Environment; added Enum to Environment |
| 2 | `api/v1alpha1/azureauthenginerole_types.go` | Modified | R1: removed defaults from 7 fields (TokenTTL, TokenMaxTTL, TokenExplicitMaxTTL, TokenNoDefaultPolicy, TokenNumUses, TokenPeriod, TokenType); added omitempty to bool/int64 fields |
| 3 | `api/v1alpha1/azuresecretengineconfig_types.go` | Modified | R1: removed defaults from ClientID, PasswordPolicy; R2: removed omitempty from Environment, RootPasswordTTL; added Enum to Environment |
| 4 | `api/v1alpha1/azuresecretenginerole_types.go` | Modified | R1: removed defaults from 9 fields; added omitempty to PersistApp |
| 5 | `api/v1alpha1/gcpauthengineconfig_types.go` | Modified | R1: removed default from ServiceAccount; R2: removed omitempty from IAMalias, IAMmetadata, GCEalias, GCEmetadata; added Enum to GCEalias |
| 6 | `api/v1alpha1/gcpauthenginerole_types.go` | Modified | R1: removed defaults from 10 fields; added omitempty to AddGroupAliases, TokenNoDefaultPolicy, TokenNumUses, TokenPeriod, AllowGCEInference |
| 7 | `config/crd/bases/redhatcop.redhat.io_azureauthengineconfigs.yaml` | Regenerated | CRD schema: removed default for clientID, added enum for environment |
| 8 | `config/crd/bases/redhatcop.redhat.io_azureauthengineroles.yaml` | Regenerated | CRD schema: removed defaults for 7 token fields |
| 9 | `config/crd/bases/redhatcop.redhat.io_azuresecretengineconfigs.yaml` | Regenerated | CRD schema: removed defaults for clientID/passwordPolicy, added enum for environment |
| 10 | `config/crd/bases/redhatcop.redhat.io_azuresecretengineroles.yaml` | Regenerated | CRD schema: removed defaults for 9 fields |
| 11 | `config/crd/bases/redhatcop.redhat.io_gcpauthengineconfigs.yaml` | Regenerated | CRD schema: removed default for serviceAccount, added enum for GCEalias |
| 12 | `config/crd/bases/redhatcop.redhat.io_gcpauthengineroles.yaml` | Regenerated | CRD schema: removed defaults for 10 fields |
| 13 | `_bmad-output/implementation-artifacts/sprint-status.yaml` | Modified | Story status updated to review |
| 14 | `_bmad-output/implementation-artifacts/7-5-4-azure-and-gcp-auth-secret-engine-types-annotation-refactor.md` | Modified | Story file updated with completion |
