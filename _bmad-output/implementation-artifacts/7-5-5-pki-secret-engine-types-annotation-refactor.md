# Story 7.5.5: PKI Secret Engine Types — Annotation Refactor

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want the PKI secret engine config and role types to follow the CRD Field Default & Validation Rules,
So that the many non-zero defaults in PKI types are correctly annotated.

## Acceptance Criteria

1. **Given** non-zero defaults (`Type`, `PrivateKeyType`, `Format`, `KeyType`, `KeyBits`, `MaxPathLength`, `CRLExpiry`, `CertificateKey`) in config with `omitempty` **When** `omitempty` removed **Then** fields always serialized
2. **Given** role fields (`UseCSRCommonName`, `UseCSRSans`, `NotBeforeDuration`, `KeyType`, `KeyBits`) with non-zero defaults and `omitempty` **When** `omitempty` removed **Then** fields always serialized
3. **Given** `TTL`, `MaxTTL` in role with `+kubebuilder:default="0s"` **When** markers removed **Then** Go zero Duration used
4. **Given** all changes **When** `make manifests generate fmt vet test` passes **Then** no regressions
5. **Given** all changes **When** `make integration` is run **Then** PKI integration tests pass

## Tasks / Subtasks

- [ ] Task 1: Refactor `pkisecretengineconfig_types.go` — remove `omitempty` from R2 fields (AC: 1)
  - [ ] 1.1: `Type` (line 68): `json:"type,omitempty"` → `json:"type"`
  - [ ] 1.2: `PrivateKeyType` (line 74): `json:"privateKeyType,omitempty"` → `json:"privateKeyType"`
  - [ ] 1.3: `Format` (line 107): `json:"format,omitempty"` → `json:"format"`
  - [ ] 1.4: `KeyType` (line 117): `json:"keyType,omitempty"` → `json:"keyType"`
  - [ ] 1.5: `KeyBits` (line 122): `json:"keyBits,omitempty"` → `json:"keyBits"`
  - [ ] 1.6: `MaxPathLength` (line 127): `json:"maxPathLength,omitempty"` → `json:"maxPathLength"`
  - [ ] 1.7: `CRLExpiry` (line 203): `json:"CRLExpiry,omitempty"` → `json:"CRLExpiry"`
  - [ ] 1.8: `CertificateKey` (line 218): `json:"certificateKey,omitempty"` → `json:"certificateKey"`
- [ ] Task 2: Refactor `pkisecretenginerole_types.go` — R1 + R2 fields (AC: 2, 3)
  - [ ] 2.1: `TTL` (line 103): remove `+kubebuilder:default="0s"`; keep `json:"TTL,omitempty"` unchanged
  - [ ] 2.2: `MaxTTL` (line 108): remove `+kubebuilder:default="0s"`; keep `json:"maxTTL,omitempty"` unchanged
  - [ ] 2.3: `KeyType` (line 178): `json:"keyType,omitempty"` → `json:"keyType"`
  - [ ] 2.4: `KeyBits` (line 183): `json:"keyBits,omitempty"` → `json:"keyBits"`
  - [ ] 2.5: `UseCSRCommonName` (line 206): `json:"useCSRCommonName,omitempty"` → `json:"useCSRCommonName"`
  - [ ] 2.6: `UseCSRSans` (line 211): `json:"useCSRSans,omitempty"` → `json:"useCSRSans"`
  - [ ] 2.7: `NotBeforeDuration` (line 270): `json:"notBeforeDuration,omitempty"` → `json:"notBeforeDuration"`
- [ ] Task 3: Run `make manifests generate fmt vet test` (AC: 4)
- [ ] Task 4: Run `make integration` — PKI tests must pass (AC: 5)

## Dev Notes

### Scope: 2 Files, 15 Field Changes

| File | R1 Removals | R2 Fixes | Enum Additions |
|------|-------------|----------|----------------|
| `api/v1alpha1/pkisecretengineconfig_types.go` | 0 fields | 8 fields | 0 (all already present) |
| `api/v1alpha1/pkisecretenginerole_types.go` | 2 fields | 5 fields | 0 (already present) |

This story is **heavily R2** (remove `omitempty` from non-zero defaults). Unlike previous stories, there are no bool/int zero-value defaults needing `omitempty` added — the only R1 fields are `metav1.Duration` types with `"0s"` defaults.

### Struct Layout — Multiple Inlined Structs in Config

`PKISecretEngineConfigSpec` uses 4 inlined structs:
- `PKIType` (lines 63-75) — `Type`, `PrivateKeyType`
- `PKICommon` (lines 77-170) — `CommonName`, `Format`, `KeyType`, `KeyBits`, `MaxPathLength`, etc.
- `PKIConfig` (lines 172-177) — wraps `PKIConfigUrls` and `PKIConfigCRL`
  - `PKIConfigCRL` (lines 199-208) — `CRLExpiry`, `CRLDisable`
- `PKIIntermediate` (lines 210-227) — `ExternalSignSecret`, `CertificateKey`, `InternalSign`

The R2 fields span 4 of these structs. Ensure you locate each field in the correct struct.

### Detailed Field Change Table — Config (R2 only — remove `omitempty`)

| Field | Struct | Type | Current Default | Current JSON Tag | Change Required |
|-------|--------|------|-----------------|-----------------|-----------------|
| `Type` | `PKIType` | string | `"root"` | `json:"type,omitempty"` | Remove `omitempty` |
| `PrivateKeyType` | `PKIType` | string | `"internal"` | `json:"privateKeyType,omitempty"` | Remove `omitempty` |
| `Format` | `PKICommon` | string | `"pem"` | `json:"format,omitempty"` | Remove `omitempty` |
| `KeyType` | `PKICommon` | string | `"rsa"` | `json:"keyType,omitempty"` | Remove `omitempty` |
| `KeyBits` | `PKICommon` | int | `2048` | `json:"keyBits,omitempty"` | Remove `omitempty` |
| `MaxPathLength` | `PKICommon` | int | `-1` | `json:"maxPathLength,omitempty"` | Remove `omitempty` |
| `CRLExpiry` | `PKIConfigCRL` | metav1.Duration | `"72h"` | `json:"CRLExpiry,omitempty"` | Remove `omitempty` |
| `CertificateKey` | `PKIIntermediate` | string | `"tls.crt"` | `json:"certificateKey,omitempty"` | Remove `omitempty` |

### Detailed Field Change Table — Role

**Rule 1 — Remove redundant zero-value `kubebuilder:default` (keep `omitempty`):**

| Field | Struct | Type | Current Default | Current JSON Tag | Change Required |
|-------|--------|------|-----------------|-----------------|-----------------|
| `TTL` | `PKIRole` | metav1.Duration | `"0s"` | `json:"TTL,omitempty"` | Remove default only |
| `MaxTTL` | `PKIRole` | metav1.Duration | `"0s"` | `json:"maxTTL,omitempty"` | Remove default only |

**Rule 2 — Remove `omitempty` from non-zero defaults:**

| Field | Struct | Type | Current Default | Current JSON Tag | Change Required |
|-------|--------|------|-----------------|-----------------|-----------------|
| `KeyType` | `PKIRole` | string | `"rsa"` | `json:"keyType,omitempty"` | Remove `omitempty` |
| `KeyBits` | `PKIRole` | int | `2048` | `json:"keyBits,omitempty"` | Remove `omitempty` |
| `UseCSRCommonName` | `PKIRole` | bool | `true` | `json:"useCSRCommonName,omitempty"` | Remove `omitempty` |
| `UseCSRSans` | `PKIRole` | bool | `true` | `json:"useCSRSans,omitempty"` | Remove `omitempty` |
| `NotBeforeDuration` | `PKIRole` | metav1.Duration | `"30s"` | `json:"notBeforeDuration,omitempty"` | Remove `omitempty` |

### Already Compliant — Config (no change needed)

| Field | Struct | Type | Default | JSON Tag | Why Compliant |
|-------|--------|------|---------|----------|--------------|
| `CommonName` | `PKICommon` | string | none | `json:"commonName,omitempty"` | Required, no default |
| `AltNames` | `PKICommon` | string | none | `json:"altNames,omitempty"` | Optional, no default |
| `IPSans` | `PKICommon` | string | none | `json:"IPSans,omitempty"` | Optional, no default |
| `URISans` | `PKICommon` | string | none | `json:"URISans,omitempty"` | Optional, no default |
| `OtherSans` | `PKICommon` | string | none | `json:"otherSans,omitempty"` | Optional, no default |
| `TTL` | `PKICommon` | metav1.Duration | none | `json:"TTL,omitempty"` | Optional, no default |
| `PrivateKeyFormat` | `PKICommon` | string | none | `json:"privateKeyFormat,omitempty"` | Optional, no default |
| `ExcludeCnFromSans` | `PKICommon` | bool | none | `json:"excludeCnFromSans,omitempty"` | Zero-value default, `omitempty` present |
| `PermittedDnsDomains` | `PKICommon` | []string | none | `json:"permittedDnsDomains,omitempty"` | Optional, no default |
| `OU` thru `SerialNumber` | `PKICommon` | string | none | various | Optional, no default |
| `IssuingCertificates` | `PKIConfigUrls` | []string | none | `json:"issuingCertificates,omitempty"` | Optional, no default |
| `CRLDistributionPoints` | `PKIConfigUrls` | []string | none | `json:"CRLDistributionPoints,omitempty"` | Optional, no default |
| `OcspServers` | `PKIConfigUrls` | []string | none | `json:"ocspServers,omitempty"` | Optional, no default |
| `CRLDisable` | `PKIConfigCRL` | bool | none | `json:"CRLDisable,omitempty"` | Zero-value default, `omitempty` present |
| `ExternalSignSecret` | `PKIIntermediate` | *corev1.LocalObjectReference | none | `json:"externalSignSecret,omitempty"` | Optional pointer, no default |
| `InternalSign` | `PKIIntermediate` | *corev1.LocalObjectReference | none | `json:"internalSign,omitempty"` | Optional pointer, no default |

### Already Compliant — Role (no change needed)

All bool fields without defaults (AllowLocalhost, AllowedDomainsTemplate, AllowBareDomains, AllowSubdomains, AllowGlobDomains, AllowAnyName, EnforceHostnames, AllowIPSans, ServerFlag, ClientFlag, CodeSigningFlag, EmailProtectionFlag, GenerateLease, NoStore, RequireCn, BasicConstraintsValidForNonCa) are already compliant: zero-value default, `omitempty` present.

All slice fields without defaults (AllowedDomains, AllowedURISans, ExtKeyUsageOids, PolicyIdentifiers, KeyUsage, ExtKeyUsage) are already compliant.

All string fields without defaults (AllowedOtherSans, OU, Organization, Country, Locality, Province, StreetAddress, PostalCode, SerialNumber) are already compliant.

### Enum Markers — All Already Present

Unlike previous stories, PKI types already have all appropriate Enum markers:
- `Type`: `+kubebuilder:validation:Enum:={"root","intermediate"}`
- `PrivateKeyType`: `+kubebuilder:validation:Enum:={"internal","exported"}`
- `Format`: `+kubebuilder:validation:Enum:={"pem","pem_bundle","der"}`
- `KeyType` (both PKICommon and PKIRole): `+kubebuilder:validation:Enum:={"rsa","ec"}`
- `KeyUsage` type alias: `+kubebuilder:validation:Enum:=DigitalSignature;KeyAgreement;...`
- `ExtKeyUsage` type alias: `+kubebuilder:validation:Enum:=ServerAuth;ClientAuth;...`

No new Enum markers are needed.

### R1 Duration Field Rationale — TTL / MaxTTL

`metav1.Duration` wraps `time.Duration`. The zero value is `Duration{Duration: 0}` which serializes as `"0s"`. The `+kubebuilder:default="0s"` is redundant because:
- Go initializes `metav1.Duration` to zero value (equivalent to `0s`)
- `omitempty` remains on the JSON tag, keeping serialized YAML clean
- The `toMap()` method always produces `payload["ttl"] = metav1.Duration{Duration: 0}` regardless of whether the API server injected `"0s"` or Go used zero value

### Impact on `toMap()` and `IsEquivalentToDesiredState()`

These changes are **annotation-only** — they modify kubebuilder markers and JSON struct tags. They do NOT change:
- `PKICommon.toMap()` (lines 523-550) — unchanged
- `PKIConfigUrls.toMap()` (lines 552-560) — unchanged
- `PKIConfigCRL.toMap()` (lines 562-569) — unchanged
- `PKIRole.toMap()` (lines 325-366) — unchanged
- `PKISecretEngineConfig.IsEquivalentToDesiredState()` (lines 439-442) — unchanged
- `PKISecretEngineRole.IsEquivalentToDesiredState()` (lines 74-77) — unchanged
- Any Go code logic

The CRD OpenAPI schema **will change** after `make manifests`. R2 fields that had `omitempty` will now always be present in serialized objects. R1 fields will lose their `default:` entries in the CRD YAML. No Go code change.

### Impact on Existing Tests

**Unit tests (`api/v1alpha1/pkisecretengineconfig_test.go`):**
- `TestPKISecretEngineConfigGetPath` — path logic → **unaffected**
- `TestPKICommonToMap` — constructs `PKICommon` with explicit field values → **unaffected**
- `TestPKIConfigUrlsToMap` — constructs `PKIConfigUrls` with explicit values → **unaffected**
- `TestPKIConfigCRLToMap` — constructs `PKIConfigCRL` with explicit values → **unaffected**
- `TestPKISecretEngineConfigIsEquivalentMatching` — **unaffected**
- `TestPKISecretEngineConfigIsEquivalentNonMatching` — **unaffected**
- `TestPKISecretEngineConfigIsEquivalentExtraFields` — **unaffected**
- `TestPKISecretEngineConfigIsDeletable` — **unaffected**
- `TestPKISecretEngineConfigConditions` — **unaffected**

**Unit tests (`api/v1alpha1/pkisecretenginerole_test.go`):**
- `TestPKISecretEngineRoleGetPath` — path logic → **unaffected**
- `TestPKIRoleToMap` — constructs `PKIRole` with explicit field values → **unaffected**
- `TestPKISecretEngineRoleIsEquivalentMatching` — **unaffected**
- `TestPKISecretEngineRoleIsEquivalentNonMatching` — **unaffected**
- `TestPKISecretEngineRoleIsEquivalentExtraFields` — **unaffected**
- `TestPKISecretEngineRoleIsDeletable` — **unaffected**
- `TestPKISecretEngineRoleConditions` — **unaffected**

**Integration tests (`controllers/pkisecretengine_controller_test.go`):**
- Test fixtures set `type`, `privateKeyType`, `commonName`, `TTL` explicitly — NOT relying on kubebuilder defaults
- Fixtures do NOT set `format`, `keyType`, `keyBits`, `maxPathLength`, `CRLExpiry`, `certificateKey`, `useCSRCommonName`, `useCSRSans`, `notBeforeDuration` — but these are R2 fields whose kubebuilder defaults REMAIN; only `omitempty` is removed
- R1 fields (`TTL`, `MaxTTL` on role): `maxTTL` is set explicitly (`"8760h"`); `TTL` is not set. Without `+kubebuilder:default="0s"`, `TTL` won't be API-server-defaulted, but Go's zero value `Duration{Duration: 0}` produces identical runtime behavior
- **Key check:** No fixture relies on server-side defaulting for any field being modified in a way that would change behavior
- All integration tests should pass without modification

### Critical Warnings

1. **Do NOT modify `toMap()` or any Go logic.** This is purely an annotation + JSON tag refactor.
2. **Run `make manifests generate` after changes.** This regenerates CRDs in `config/crd/bases/`. The diff will show removed `default:` entries (R1 fields) in the OpenAPI schema.
3. **Run `make generate`** to regenerate `zz_generated.deepcopy.go` (struct tag changes may affect generated code).
4. **Do NOT add new Enum markers.** All appropriate Enum constraints are already present on PKI types. `KeyBits` is intentionally NOT an Enum candidate (valid values depend on `KeyType`).
5. **Do NOT change `MaxPathLength`'s default value of `-1`.** This is a valid non-zero default meaning "no limit". We only remove `omitempty` from the JSON tag.
6. **R2 fields retain their `+kubebuilder:default` markers.** Only `omitempty` is removed from the JSON tag. The default value annotation stays.
7. **R1 fields (`TTL`, `MaxTTL`) retain `omitempty` on JSON tag.** Only the `+kubebuilder:default="0s"` marker line is removed.
8. **PKICommon `TTL` (config) vs PKIRole `TTL` (role) are different fields.** Config's `TTL` (line 101) has NO default and is already compliant. Role's `TTL` (line 103-104) has `+kubebuilder:default="0s"` which is the R1 target.
9. **`KeyType` appears in BOTH `PKICommon` (config, line 117) and `PKIRole` (role, line 178).** Both need `omitempty` removed. Do NOT confuse the two.
10. **`KeyBits` appears in BOTH `PKICommon` (config, line 122) and `PKIRole` (role, line 183).** Both need `omitempty` removed.

### Pattern for Non-Zero Default Fields (R2) — Remove `omitempty`

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:validation:Enum:={"rsa","ec"}
// +kubebuilder:default="rsa"
KeyType string `json:"keyType,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:validation:Enum:={"rsa","ec"}
// +kubebuilder:default="rsa"
KeyType string `json:"keyType"`
```

### Pattern for Zero-Value Duration Fields (R1) — Remove Default Only

Before:
```go
// +kubebuilder:validation:Optional
// +kubebuilder:default="0s"
TTL metav1.Duration `json:"TTL,omitempty"`
```

After:
```go
// +kubebuilder:validation:Optional
TTL metav1.Duration `json:"TTL,omitempty"`
```

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/pkisecretengineconfig_types.go` | Modified | Remove `omitempty` from 8 R2 fields |
| 2 | `api/v1alpha1/pkisecretenginerole_types.go` | Modified | Remove 2 R1 default markers, remove `omitempty` from 5 R2 fields |
| 3 | `config/crd/bases/redhatcop.redhat.io_pkisecretengineconfigs.yaml` | Regenerated | CRD schema updated by `make manifests` |
| 4 | `config/crd/bases/redhatcop.redhat.io_pkisecretengineroles.yaml` | Regenerated | CRD schema updated by `make manifests` |

### Project Structure Notes

- CRD types live in `api/v1alpha1/` — both files are in this directory
- CRD schemas are regenerated into `config/crd/bases/` by `make manifests`
- DeepCopy is regenerated into `api/v1alpha1/zz_generated.deepcopy.go` by `make generate`
- Test fixtures in `test/pkisecretengine/` — verify they pass after changes (fixtures set values explicitly, not relying on defaults being modified)
- Integration test file: `controllers/pkisecretengine_controller_test.go` (has `//go:build integration` tag)
- Unit test files: `api/v1alpha1/pkisecretengineconfig_test.go`, `api/v1alpha1/pkisecretenginerole_test.go`

### References

- [Source: _bmad-output/project-context.md#CRD Field Default & Validation Rules] — Rules 1-6 governing annotation behavior
- [Source: _bmad-output/planning-artifacts/epics.md#Story 7.5.5] — Epic story definition and acceptance criteria
- [Source: api/v1alpha1/pkisecretengineconfig_types.go:63-75] — `PKIType` struct (Type, PrivateKeyType)
- [Source: api/v1alpha1/pkisecretengineconfig_types.go:77-170] — `PKICommon` struct (Format, KeyType, KeyBits, MaxPathLength)
- [Source: api/v1alpha1/pkisecretengineconfig_types.go:199-208] — `PKIConfigCRL` struct (CRLExpiry)
- [Source: api/v1alpha1/pkisecretengineconfig_types.go:210-227] — `PKIIntermediate` struct (CertificateKey)
- [Source: api/v1alpha1/pkisecretenginerole_types.go:99-271] — `PKIRole` struct with all field annotations
- [Source: api/v1alpha1/pkisecretengineconfig_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: api/v1alpha1/pkisecretenginerole_test.go] — Existing unit tests (unaffected by annotation changes)
- [Source: controllers/pkisecretengine_controller_test.go] — PKI integration tests (must pass post-change)
- [Source: test/pkisecretengine/] — Test fixtures (verify against changes)

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

**Key differences from Stories 7.5.1-7.5.4:**
- This story is **R2-heavy** (13 out of 15 changes are R2). Only 2 R1 fields exist.
- No bool/int zero-value defaults needing `omitempty` added (unlike previous stories)
- The R1 fields are `metav1.Duration` type with `"0s"` default — a pattern not seen in prior stories
- All Enum markers are already present — no new Enums to add
- PKI integration tests exist and must be run (unlike Story 7.5.4 which had no integration tests)
- Config type uses 4 inlined structs — fields are spread across `PKIType`, `PKICommon`, `PKIConfigCRL`, `PKIIntermediate`
- `KeyType` and `KeyBits` appear in BOTH config and role files — must be changed in both

**From Epic 7 Retrospective:**
- Codebase is stable after hardening epic
- Coverage exists for PKI types via both unit tests and integration tests
- Story 5.3 tested PKI integration (remaining secret engine types)

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
