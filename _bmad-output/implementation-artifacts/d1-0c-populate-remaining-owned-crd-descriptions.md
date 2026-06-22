# Story D1.0c: Populate Remaining Owned CRD Descriptions

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator maintainer,
I want the CSV base owned CRD list to contain all CRD kinds in correct alphabetical order with verified descriptions,
So that `operator-sdk bundle validate` produces zero empty-description warnings and the bundle metadata is publication-ready.

## Acceptance Criteria

1. **Given** the CSV base owned list currently has 47 entries with descriptions auto-populated during R1-8's `make bundle` run **When** the list is audited **Then** all 47 entries have non-empty descriptions matching the Go doc comment on each root type struct (`<Kind> is the Schema for the <lowercase-plural> API`)
2. **Given** the CSV base owned list is currently NOT alphabetically sorted (R1-8 and auto-discovered entries were prepended at the top) **When** the list is re-sorted **Then** all entries are in strict alphabetical order by `kind` field
3. **Given** the sorted list with verified descriptions **When** `kustomize build config/manifests | operator-sdk generate bundle` is run (after sorting) **Then** the generated CSV at `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` contains all 47 owned CRDs with non-empty descriptions (note: `make bundle` includes `operator-sdk generate kustomize manifests` which rewrites the CSV base ordering — the source-of-truth sort is in the CSV base, and bundle generation is run separately after sorting to validate descriptions propagate correctly)
4. **Given** the regenerated bundle **When** `operator-sdk bundle validate ./bundle` is run **Then** zero empty-description warnings are emitted for any owned CRD
5. **Given** this is a metadata-only story **When** all changes are complete **Then** zero Go source files are modified

## Tasks / Subtasks

- [x] Task 1: Re-sort the owned CRD list alphabetically by `kind` (AC: 2)
  - [x] 1.1: In `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`, re-order the `spec.customresourcedefinitions.owned` list so entries are sorted alphabetically by `kind` field
  - [x] 1.2: Verify the corrected order is:
    1. Audit
    2. AuditRequestHeader
    3. AuthEngineMount
    4. AzureAuthEngineConfig
    5. AzureAuthEngineRole
    6. AzureSecretEngineConfig
    7. AzureSecretEngineRole
    8. CertAuthEngineConfig
    9. CertAuthEngineRole
    10. DatabaseSecretEngineConfig
    11. DatabaseSecretEngineRole
    12. DatabaseSecretEngineStaticRole
    13. Entity
    14. EntityAlias
    15. GCPAuthEngineConfig
    16. GCPAuthEngineRole
    17. GitHubSecretEngineConfig
    18. GitHubSecretEngineRole
    19. Group
    20. GroupAlias
    21. IdentityOIDCAssignment
    22. IdentityOIDCClient
    23. IdentityOIDCProvider
    24. IdentityOIDCScope
    25. IdentityTokenConfig
    26. IdentityTokenKey
    27. IdentityTokenRole
    28. JWTOIDCAuthEngineConfig
    29. JWTOIDCAuthEngineRole
    30. KubernetesAuthEngineConfig
    31. KubernetesAuthEngineRole
    32. KubernetesSecretEngineConfig
    33. KubernetesSecretEngineRole
    34. LDAPAuthEngineConfig
    35. LDAPAuthEngineGroup
    36. PasswordPolicy
    37. PKISecretEngineConfig
    38. PKISecretEngineRole
    39. Policy
    40. QuaySecretEngineConfig
    41. QuaySecretEngineRole
    42. QuaySecretEngineStaticRole
    43. RabbitMQSecretEngineConfig
    44. RabbitMQSecretEngineRole
    45. RandomSecret
    46. SecretEngineMount
    47. VaultSecret
  - [x] 1.3: Preserve the existing description, displayName, name, and version for each entry — do not modify field values, only reorder entries
- [x] Task 2: Verify all descriptions match Go doc comments (AC: 1)
  - [x] 2.1: For each of the 47 entries, confirm the `description` value matches the Go doc comment on the root type struct in `api/v1alpha1/<lowercase>_types.go` (the `// <Kind> is the Schema for the <lowercase-plural> API` pattern)
  - [x] 2.2: Note the one known exception: `KubernetesAuthEngineRole` has a custom description (`can be used to define a KubernetesAuthEngineRole for the kube-auth authentication method`) — this is intentional and should be preserved
- [x] Task 3: Regenerate bundle and validate (AC: 3, 4)
  - [x] 3.1: Run `make bundle`
  - [x] 3.2: Inspect `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` — confirm all 47 owned CRDs have non-empty descriptions
  - [x] 3.3: Run `operator-sdk bundle validate ./bundle` — zero empty-description warnings

## Dev Notes

### Background: How D1.0c's Scope Evolved

The R1 retrospective (2026-06-21) identified that after R1-8 added 3 owned CRD descriptions (AzureSecretEngineConfig, Entity, EntityAlias), 15 CRDs remained missing from the CSV base. Action item #5 created this story to populate those remaining descriptions.

However, when R1-8 ran `make bundle`, the `operator-sdk generate kustomize manifests` command auto-discovered ALL CRDs from `config/crd/bases/` and auto-added the missing ones to the CSV base file — with descriptions auto-populated from the Go doc comments on each root type struct. This brought the total from 33 → 47 entries, all with non-empty descriptions.

**The descriptions are already populated.** The remaining work is:
1. **Fix alphabetical ordering** — the R1-8 entries and auto-discovered entries were prepended at the top of the list instead of being inserted at their correct alphabetical positions
2. **Verify accuracy** — confirm all auto-populated descriptions match the Go doc comments
3. **Validate** — run `make bundle` + `operator-sdk bundle validate` to confirm zero warnings

### Current Ordering Problems

The owned list has three groups of entries that are NOT alphabetically interleaved:

**Group 1 (lines 32-57):** R1-8 manually-added entries + auto-discovered entries, prepended at top:
- AzureSecretEngineConfig, Entity, EntityAlias (R1-8)
- AuditRequestHeader, Audit (auto-discovered)

**Group 2 (lines 58-295):** Original 33 entries + remaining auto-discovered entries, in their original positions:
- AuthEngineMount through VaultSecret

**Group 3 (ordering issues within original entries):**
- GroupAlias appears before Group (should be reversed)

After re-sorting, all 47 entries must be in strict alphabetical order by `kind`.

### This Is a Metadata-Only Story

Per the epic rule: "Every story must pass `make manifests generate fmt vet test` and `make integration` unless the story is metadata-only, in which case `make bundle` is the minimum required verification."

This story changes **zero Go source files**. The only changes are:
1. `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` — re-sort owned CRD entries

No `make test` or `make integration` run is required, but `make bundle` (which includes `make manifests`) is the minimum gate.

### How Owned CRD Descriptions Flow Into the Bundle

```
api/v1alpha1/*_types.go           ← Go doc comments on root struct
         ↓
controller-gen (make manifests)   ← Writes openAPIV3Schema.description in config/crd/bases/*.yaml
         ↓
config/manifests/bases/vault-config-operator.clusterserviceversion.yaml
  spec.customresourcedefinitions.owned[].description  ← The CSV base (source-of-truth)
         ↓
operator-sdk generate kustomize manifests  ← Auto-adds missing CRDs from config/crd/bases/
         ↓
kustomize build config/manifests | operator-sdk generate bundle
         ↓
bundle/manifests/vault-config-operator.clusterserviceversion.yaml  ← Generated CSV
         ↓
operator-sdk bundle validate ./bundle  ← Warns if owned CRD description is empty
```

### Owned Entry Format

All entries follow the standard 5-field format:
```yaml
- description: <Kind> is the Schema for the <lowercase-plural> API
  displayName: <Space Separated Kind>
  kind: <Kind>
  name: <lowercase-plural>.redhatcop.redhat.io
  version: v1alpha1
```

Exception: `KubernetesAuthEngineRole` has a custom description — preserve it as-is.

### What NOT to Do

- Do NOT modify any Go source files — this is metadata-only
- Do NOT change Go doc comments on the root type structs
- Do NOT modify `alm-examples` or `spec.minKubeVersion` — those are separate stories
- Do NOT create a `.golangci.yml` config file
- Do NOT run `make test` or `make integration` — only `make bundle` is the verification gate
- Do NOT modify description text — only re-sort entries
- Do NOT change any fields (description, displayName, name, version) — only move entries to correct alphabetical positions

### Verification Checklist

After `make bundle`:
1. `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` should exist
2. All 47 owned CRDs should have non-empty descriptions
3. `operator-sdk bundle validate ./bundle` should not emit empty-description warnings
4. The owned list in the CSV base should be alphabetically sorted by `kind`

### Previous Story Intelligence

**From R1-8 (direct predecessor — populated 3 CRD descriptions):**
- Added AzureSecretEngineConfig, Entity, EntityAlias to CSV base
- Discovered that `operator-sdk generate kustomize manifests` auto-adds missing CRDs with descriptions from Go doc comments
- Review finding: "Owned CRD entries were inserted at the top of the CSV base list instead of at the required alphabetical insertion points" — this finding was carried forward and is resolved by this story
- The `bundle/` directory is gitignored — only the CSV base file needs committing
- `make bundle` is the verification gate

**From R1 Retrospective (2026-06-21):**
- Action item #5: "Populate remaining 15 owned CRD descriptions in bundle CSV base"
- The descriptions were auto-populated during R1-8's `make bundle` run, so the remaining work is ordering and verification
- [Source: `_bmad-output/implementation-artifacts/epic-R1-retro-2026-06-21.md#Action Items`]

**From R1-8 Completion Notes:**
- The auto-discovery by `operator-sdk generate kustomize manifests` brought the total from 33 → 47 entries
- All descriptions match the `<Kind> is the Schema for the <lowercase-plural> API` Go doc comment pattern
- [Source: `_bmad-output/implementation-artifacts/R1-8-populate-owned-crd-descriptions-for-community-operators-bundle.md`]

### golangci-lint Note

Not applicable — this story modifies zero Go source files. No lint check needed.

### Files Modified

Only 1 file: `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`

### Project Structure Notes

- Changes confined to `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` (re-sort existing entries)
- No new files created
- No Go source changes — no `make generate` or `make manifests` needed beyond what `make bundle` already runs
- The `bundle/` directory is gitignored generated output — only the CSV base file is committed

### References

- [Source: config/manifests/bases/vault-config-operator.clusterserviceversion.yaml:30-295] — CSV base owned list (47 entries, needs alphabetical sort)
- [Source: _bmad-output/implementation-artifacts/R1-8-populate-owned-crd-descriptions-for-community-operators-bundle.md] — predecessor story, established CSV base as source-of-truth
- [Source: _bmad-output/implementation-artifacts/epic-R1-retro-2026-06-21.md#Action Items] — action item #5 that created this story
- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.8] — original acceptance criteria for owned CRD descriptions
- [Source: Makefile:350-355] — `make bundle` target definition
- [Source: config/manifests/kustomization.yaml] — manifests assembly
- [Source: _bmad-output/project-context.md] — project conventions and tooling

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- `make bundle` rewrites CSV base via `operator-sdk generate kustomize manifests`, undoing any prior sort. Solution: run `make bundle` first, then re-sort, then re-run only `kustomize build | operator-sdk generate bundle` + validate.
- Case-sensitive Python `sorted()` placed PKI before Password; fixed with `key=str.lower` for case-insensitive alphabetical sort.
- The generated bundle CSV (`bundle/manifests/`) applies its own ordering via kustomize/operator-sdk tooling — this is expected and doesn't affect the source-of-truth CSV base.

### Completion Notes List

- Re-sorted all 47 owned CRD entries in `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` to strict case-insensitive alphabetical order by `kind` field
- Verified all 47 descriptions match Go doc comments on root type structs (automated cross-reference of all `api/v1alpha1/*_types.go` files)
- Confirmed KubernetesAuthEngineRole custom description preserved as-is
- `make bundle` succeeded; `operator-sdk bundle validate ./bundle` passed with zero warnings
- Zero Go source files modified (metadata-only story)
- Changes: 25 insertions / 25 deletions (pure reorder, no field value changes)

### Change Log

- 2026-06-22: Re-sorted owned CRD list in CSV base to alphabetical order by kind; verified all 47 descriptions; regenerated and validated bundle

### File List

- `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` (modified — re-sorted owned CRD entries)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified — status tracking)
- `_bmad-output/implementation-artifacts/d1-0c-populate-remaining-owned-crd-descriptions.md` (modified — task checkboxes, dev record)

### Review Findings

- [x] [Review][Decision] Required `make bundle` verification flow does not currently preserve or prove the accepted final state — **Resolved:** AC3 revised to reflect actual verification path. The CSV base is the source-of-truth for ordering; `operator-sdk generate kustomize manifests` rewrites ordering as a known behavior. Bundle validation is run separately after sorting to confirm descriptions propagate. Future `make bundle` runs may reorder the CSV base, but descriptions remain intact and the source-of-truth sort is always recoverable from the committed CSV base.
