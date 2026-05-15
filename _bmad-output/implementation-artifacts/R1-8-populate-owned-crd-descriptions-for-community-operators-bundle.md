# Story R1.8: Populate Owned CRD Descriptions for Community Operators Bundle

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator maintainer,
I want the generated CSV to contain non-empty owned CRD descriptions for the flagged APIs,
So that the Community Operators bundle presents complete metadata instead of warnings.

## Acceptance Criteria

1. **Given** Community Operators validation warns that owned CRDs `azuresecretengineconfigs.redhatcop.redhat.io`, `entities.redhatcop.redhat.io`, and `entityaliases.redhatcop.redhat.io` have empty descriptions **When** the description source-of-truth is fixed and the bundle is regenerated **Then** each owned CRD entry in the CSV contains a non-empty, human-readable description
2. **Given** CRD descriptions in the bundle CSV originate from the CSV base file at `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` under `spec.customresourcedefinitions.owned` **When** the fix is implemented **Then** the project uses the CSV base as the stable source-of-truth (matching the pattern of all 33 existing owned entries)
3. **Given** `make bundle` is run after the metadata fix **When** bundle validation completes **Then** the empty-description warnings for AzureSecretEngineConfig, Entity, and EntityAlias are no longer emitted
4. **Given** the affected APIs already have CRD material in the repo **When** descriptions are updated **Then** the wording matches the Go root type doc comment pattern used by every other CRD in the project

## Tasks / Subtasks

- [ ] Task 1: Verify the description source-of-truth pipeline (AC: 2)
  - [ ] 1.1: Confirm that owned CRD descriptions flow from `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` → `kustomize build config/manifests` → `operator-sdk generate bundle` → `bundle/manifests/vault-config-operator.clusterserviceversion.yaml`
  - [ ] 1.2: Confirm that `operator-sdk generate kustomize manifests --interactive=false` auto-discovers CRDs from `config/crd/bases/` and adds any missing ones to the CSV with empty descriptions — this is the root cause of the warnings
  - [ ] 1.3: Verify the existing pattern: all 33 current entries in the CSV base `owned` list have the format `description: <Kind> is the Schema for the <lowercase-plural> API`
- [ ] Task 2: Add `AzureSecretEngineConfig` to the CSV base owned list (AC: 1, 4)
  - [ ] 2.1: Add entry after line 48 (after AzureAuthEngineRole, before AzureSecretEngineRole) to maintain alphabetical order:
    ```yaml
    - description: AzureSecretEngineConfig is the Schema for the azuresecretengineconfigs
        API
      displayName: Azure Secret Engine Config
      kind: AzureSecretEngineConfig
      name: azuresecretengineconfigs.redhatcop.redhat.io
      version: v1alpha1
    ```
  - [ ] 2.2: Verify the description matches the Go comment in `api/v1alpha1/azuresecretengineconfig_types.go:68` (`// AzureSecretEngineConfig is the Schema for the azuresecretengineconfigs API`)
- [ ] Task 3: Add `Entity` to the CSV base owned list (AC: 1, 4)
  - [ ] 3.1: Add entry after GitHubSecretEngineRole (alphabetically between GitHub* and Group*):
    ```yaml
    - description: Entity is the Schema for the entities API
      displayName: Entity
      kind: Entity
      name: entities.redhatcop.redhat.io
      version: v1alpha1
    ```
  - [ ] 3.2: Verify the description matches the Go comment in `api/v1alpha1/entity_types.go:79` (`// Entity is the Schema for the entities API`)
- [ ] Task 4: Add `EntityAlias` to the CSV base owned list (AC: 1, 4)
  - [ ] 4.1: Add entry immediately after the Entity entry (alphabetically before Group*):
    ```yaml
    - description: EntityAlias is the Schema for the entityaliases API
      displayName: Entity Alias
      kind: EntityAlias
      name: entityaliases.redhatcop.redhat.io
      version: v1alpha1
    ```
  - [ ] 4.2: Verify the description matches the Go comment in `api/v1alpha1/entityalias_types.go:90` (`// EntityAlias is the Schema for the entityaliases API`)
- [ ] Task 5: Regenerate bundle and validate (AC: 1, 3)
  - [ ] 5.1: Run `make bundle`
  - [ ] 5.2: Inspect `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` — confirm the three CRDs now have non-empty descriptions in `spec.customresourcedefinitions.owned`
  - [ ] 5.3: Run `operator-sdk bundle validate ./bundle` and verify the empty-description warnings for AzureSecretEngineConfig, Entity, and EntityAlias are gone
- [ ] Task 6: Commit (AC: 1, 3)
  - [ ] 6.1: Commit the CSV base change (`config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`)
  - [ ] 6.2: Commit the regenerated bundle output (`bundle/` directory is tracked in git)

## Dev Notes

### Root Cause

The CSV base file `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` contains `spec.customresourcedefinitions.owned` with **33 entries** — but the project has **51 CRD kinds**. When `operator-sdk generate kustomize manifests` runs (as part of `make bundle`), it auto-discovers all CRDs from `config/crd/bases/` and adds any missing ones to the generated CSV with **empty descriptions**. The three CRDs flagged by Community Operators validation (`AzureSecretEngineConfig`, `Entity`, `EntityAlias`) are simply missing from the CSV base.

The other 15 missing CRDs (Audit, AuditRequestHeader, CertAuth*, Identity*, etc.) also have empty descriptions in the generated bundle, but they are **not in scope** for this story — R1.8 only addresses the three explicitly flagged by Community Operators PR `#9655`.

### This Is a Metadata-Only Story

Per the epic rule: "Every story must pass `make manifests generate fmt vet test` and `make integration` unless the story is metadata-only, in which case `make bundle` is the minimum required verification."

This story changes **zero Go source files**. The only changes are:
1. `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` — add 3 owned CRD entries (~15 lines)
2. Regenerated `bundle/` output from `make bundle`

No `make test` or `make integration` run is required, but `make bundle` (which includes `make manifests`) is the minimum gate.

### How Owned CRD Descriptions Flow Into the Bundle

```
api/v1alpha1/*_types.go           ← Go doc comments on root struct (e.g., "// Entity is the Schema for the entities API")
         ↓
controller-gen (make manifests)   ← Writes openAPIV3Schema.description in config/crd/bases/*.yaml
         ↓
config/manifests/bases/vault-config-operator.clusterserviceversion.yaml
  spec.customresourcedefinitions.owned[].description  ← Manual entries (the CSV base)
         ↓
operator-sdk generate kustomize manifests  ← Merges CRD discovery with CSV base
         ↓
kustomize build config/manifests | operator-sdk generate bundle
         ↓
bundle/manifests/vault-config-operator.clusterserviceversion.yaml  ← Generated CSV
         ↓
operator-sdk bundle validate ./bundle  ← Warns if owned CRD description is empty
```

**Key point:** `operator-sdk generate kustomize manifests` does NOT copy the `openAPIV3Schema.description` from the CRD YAML into the `owned` description. The CSV base `owned` list is the **sole source-of-truth** for owned CRD descriptions in the bundle. A CRD missing from the base gets an empty description in the generated CSV.

### Existing Owned List Pattern

All 33 existing entries follow an identical 5-field format:

```yaml
- description: <Kind> is the Schema for the <lowercase-plural> API
  displayName: <Space Separated Kind>
  kind: <Kind>
  name: <lowercase-plural>.redhatcop.redhat.io
  version: v1alpha1
```

The only exception is `KubernetesAuthEngineRole` which has a custom description (`can be used to define a KubernetesAuthEngineRole for the kube-auth authentication method`). Use the standard `Schema for…` pattern for the three new entries.

### Alphabetical Insertion Points

The CSV base owned list is sorted alphabetically by `kind`. Insert locations:

| New Entry | Insert After | Insert Before |
|-----------|-------------|---------------|
| `AzureSecretEngineConfig` | `AzureAuthEngineRole` (line 47) | `AzureSecretEngineRole` (line 49) |
| `Entity` | `GitHubSecretEngineRole` (line 94) | `GroupAlias` (line 96) |
| `EntityAlias` | `Entity` (new entry above) | `GroupAlias` (line 96) |

### What NOT to Do

- Do NOT modify any Go source files — this is metadata-only
- Do NOT change Go doc comments on the root type structs — they are already correct
- Do NOT add entries for the other 15 missing CRDs (Audit, AuditRequestHeader, CertAuth*, Identity*, etc.) — those are out of scope
- Do NOT modify `alm-examples` — that was R1.7's scope
- Do NOT add `spec.minKubeVersion` — that is R1.9's scope
- Do NOT create a `.golangci.yml` config file
- Do NOT run `make test` or `make integration` — this story only requires `make bundle` as the verification gate
- Do NOT reorder existing entries in the CSV base — only insert new entries in alphabetical position

### Prerequisites

Per the epic ordering: R1.1 → R1.2a → R1.2b → R1.3 → R1.2c → R1.7 → **R1.8** → R1.9 → R1.4 → R1.5 → R1.6. This story follows R1.7 (bundle example annotations). However, R1.8 has no code dependencies on R1.1–R1.7 — it touches only metadata files. It can technically be done in any order, but the epic ordering places it here.

### Verification Checklist

After `make bundle`:
1. `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` should exist
2. The `spec.customresourcedefinitions.owned` section should contain non-empty descriptions for `AzureSecretEngineConfig`, `Entity`, and `EntityAlias`
3. `operator-sdk bundle validate ./bundle` should not emit empty-description warnings for those three CRDs
4. The remaining warnings (R1.9: missing `spec.minKubeVersion`) are expected and are a separate story

### Previous Story Intelligence

**From R1.7 (immediately preceding in ordering):**
- Also a metadata-only story — added Entity/EntityAlias sample annotations to bundle
- Changed `config/samples/kustomization.yaml` and regenerated `bundle/`
- `make bundle` is the verification gate (same as this story)
- The R1.7 story confirmed that `bundle/` directory is tracked in git and must be committed
- R1.7 explicitly noted: "Do NOT add owned CRD descriptions for Entity/EntityAlias — that is R1.8's scope"

**From Epic R1 preamble:**
- Community Operators PR `#9655` surfaced these bundle metadata warnings
- Stories R1.7-R1.9 address the three addressable warning categories
- The deprecated `operatorhub` validator warning and the FBC migration recommendation are explicitly out of scope

### Project Structure Notes

- Changes confined to `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` (add 3 entries, ~15 lines) and regenerated `bundle/` output
- No new files created
- No Go source changes — no `make generate` or `make manifests` needed beyond what `make bundle` already runs
- The `bundle/` directory is generated output that is tracked in git

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.8] — acceptance criteria, task list
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, story ordering, Community Operators PR #9655 context
- [Source: config/manifests/bases/vault-config-operator.clusterserviceversion.yaml:30-220] — CSV base owned list (33 entries, missing AzureSecretEngineConfig/Entity/EntityAlias)
- [Source: api/v1alpha1/azuresecretengineconfig_types.go:68] — Go doc comment for AzureSecretEngineConfig
- [Source: api/v1alpha1/entity_types.go:79] — Go doc comment for Entity
- [Source: api/v1alpha1/entityalias_types.go:90] — Go doc comment for EntityAlias
- [Source: Makefile:350-355] — `make bundle` target definition
- [Source: config/manifests/kustomization.yaml] — manifests assembly (includes bases + ../default + ../samples)
- [Source: _bmad-output/implementation-artifacts/R1-7-bundle-example-annotations-for-entity-and-entityalias.md] — previous story context

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### Change Log

### File List
