# Story R1.7: Bundle Example Annotations for `Entity` and `EntityAlias`

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator maintainer,
I want the bundle metadata to include valid examples for the `Entity` and `EntityAlias` APIs,
So that Community Operators validation no longer warns that those provided APIs lack example annotations.

## Acceptance Criteria

1. **Given** Community Operators validation warns that `redhatcop.redhat.io/v1alpha1, Kind=Entity` and `Kind=EntityAlias` do not have example annotations **When** the example source-of-truth is updated and the bundle is regenerated **Then** the generated CSV contains valid examples for both APIs
2. **Given** the repository already contains `config/samples/redhatcop_v1alpha1_entity.yaml` and `config/samples/redhatcop_v1alpha1_entityalias.yaml` **When** those examples are normalized to current schema expectations **Then** they are suitable for reuse in bundle metadata
3. **Given** `make bundle` runs successfully **When** `operator-sdk bundle validate ./bundle` is executed by the target **Then** the "provided API should have an example annotation" warnings for `Entity` and `EntityAlias` are no longer emitted
4. **Given** the examples are added to bundle metadata **When** future CRD schema changes happen **Then** the source file for those examples is obvious and documented in the story notes or implementation comments

## Tasks / Subtasks

- [x] Task 1: Understand how bundle examples are sourced (AC: 4)
  - [x] 1.1: Confirm the pipeline: `config/samples/kustomization.yaml` → `config/manifests/kustomization.yaml` (includes `../samples`) → `make bundle` → `operator-sdk generate bundle` → `alm-examples` annotation in generated CSV
  - [x] 1.2: Verify that the **only** change needed is adding the two sample filenames to `config/samples/kustomization.yaml` — no Go code, markers, or type-level annotations are involved
- [x] Task 2: Normalize `config/samples/redhatcop_v1alpha1_entity.yaml` (AC: 2)
  - [x] 2.1: Review current file against `EntitySpec` schema (`entity_types.go`): fields are `authentication` (required), inline `EntityConfig` with `metadata` (map[string]string), `policies` ([]string), `disabled` (bool), plus optional `connection` and `name`
  - [x] 2.2: Verify the sample is a valid minimal-but-realistic example — current content already has `authentication`, `metadata`, `policies`, `disabled` which covers all user-facing fields
  - [x] 2.3: If any fields are invalid or missing required values, update them; if the sample is already valid, no changes needed
- [x] Task 3: Normalize `config/samples/redhatcop_v1alpha1_entityalias.yaml` (AC: 2)
  - [x] 3.1: Review current file against `EntityAliasSpec` schema (`entityalias_types.go`): fields are `authentication` (required), inline `EntityAliasConfig` with `authEngineMountPath` (required), `entityName` (required), `customMetadata` (map[string]string, optional), plus optional `connection` and `name`
  - [x] 3.2: Verify the sample is a valid minimal-but-realistic example — current content already has `authentication`, `authEngineMountPath`, `entityName`, `customMetadata` which covers all user-facing fields
  - [x] 3.3: If any fields are invalid or missing required values, update them; if the sample is already valid, no changes needed
- [x] Task 4: Add samples to kustomization resources (AC: 1)
  - [x] 4.1: Edit `config/samples/kustomization.yaml` — add `- redhatcop_v1alpha1_entity.yaml` and `- redhatcop_v1alpha1_entityalias.yaml` to the `resources:` list (before the `#+kubebuilder:scaffold:manifestskustomizesamples` marker)
- [x] Task 5: Regenerate bundle and validate (AC: 1, 3)
  - [x] 5.1: Run `make bundle`
  - [x] 5.2: Inspect the generated CSV in `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` — confirm the `alm-examples` annotation now contains JSON entries for `Entity` and `EntityAlias`
  - [x] 5.3: Verify `operator-sdk bundle validate ./bundle` output no longer includes warnings for `Entity` or `EntityAlias` missing examples
- [x] Task 6: Commit and verify (AC: 1, 3, 4)
  - [x] 6.1: Commit the kustomization change and any sample normalization
  - [x] 6.2: Commit the regenerated bundle output (the `bundle/` directory is tracked in git)

## Dev Notes

### Root Cause

The `config/samples/kustomization.yaml` file lists 47 sample YAML files that are included in bundle generation. The Entity and EntityAlias sample files (`redhatcop_v1alpha1_entity.yaml`, `redhatcop_v1alpha1_entityalias.yaml`) **already exist** in `config/samples/` but are **not listed** in the kustomization resources. This is why `operator-sdk bundle validate` warns that these two provided APIs have no example annotation — the samples exist but are never fed into the bundle generator.

### This Is a Metadata-Only Story

Per the epic rule: "Every story must pass `make manifests generate fmt vet test` and `make integration` unless the story is metadata-only, in which case `make bundle` is the minimum required verification."

This story changes **zero Go source files**. The only changes are:
1. `config/samples/kustomization.yaml` — add 2 lines
2. Possibly normalize the 2 sample YAML files (likely no changes needed — they already match the CRD schema)
3. Regenerated `bundle/` output from `make bundle`

No `make test` or `make integration` run is required, but `make bundle` (which includes `make manifests`) is the minimum gate.

### How Bundle Example Generation Works

```
config/samples/*.yaml           ← Individual sample YAML files (one per CRD kind)
         ↓
config/samples/kustomization.yaml  ← Lists which samples to include as resources
         ↓
config/manifests/kustomization.yaml  ← Includes ../samples as a resource
         ↓
kustomize build config/manifests   ← Assembles all manifests + samples
         ↓
operator-sdk generate bundle       ← Reads samples → writes alm-examples annotation in CSV
         ↓
bundle/manifests/vault-config-operator.clusterserviceversion.yaml  ← Generated CSV with alm-examples
         ↓
operator-sdk bundle validate ./bundle  ← Validates that every owned CRD kind has an example
```

The CSV base file (`config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`) has `alm-examples: '[]'` which is a placeholder — the real content is generated by `operator-sdk generate bundle` from the kustomized sample resources.

### Current Sample File Status

**Entity sample** (`config/samples/redhatcop_v1alpha1_entity.yaml`):
- Already has: `apiVersion`, `kind`, `metadata` (with labels), `spec.authentication`, `spec.metadata`, `spec.policies`, `spec.disabled`
- Matches `EntitySpec` / `EntityConfig` schema in `entity_types.go`
- All required fields present, all optional fields have realistic values
- Likely no normalization needed

**EntityAlias sample** (`config/samples/redhatcop_v1alpha1_entityalias.yaml`):
- Already has: `apiVersion`, `kind`, `metadata` (with labels), `spec.authentication`, `spec.authEngineMountPath`, `spec.entityName`, `spec.customMetadata`
- Matches `EntityAliasSpec` / `EntityAliasConfig` schema in `entityalias_types.go`
- All required fields present, optional `customMetadata` has realistic values
- Likely no normalization needed

### Sample Consistency With Test Fixtures

The integration test fixtures at `test/identity/01-entity-sample.yaml` and `test/identity/02-entityalias-sample.yaml` use a similar structure but with simpler metadata (no kustomize labels). The `config/samples/` versions are the bundle source-of-truth — test fixtures are separate.

### What NOT to Do

- Do NOT modify any Go source files — this is metadata-only
- Do NOT modify `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` — the `alm-examples` annotation is auto-generated
- Do NOT add owned CRD descriptions for Entity/EntityAlias — that is R1.8's scope
- Do NOT add `spec.minKubeVersion` — that is R1.9's scope
- Do NOT create a `.golangci.yml` config file
- Do NOT run `make test` or `make integration` — this story only requires `make bundle` as the verification gate
- Do NOT reorder existing entries in `config/samples/kustomization.yaml` — add new entries at the end (before the scaffold marker)

### Prerequisites

Per the epic ordering: R1.1 → R1.2a → R1.2b → R1.3 → R1.2c → **R1.7** → R1.8 → R1.9 → R1.4 → R1.5 → R1.6. This story follows R1.2c (lint green gate). However, R1.7 has no code dependencies on R1.1–R1.2c — it touches only metadata files. It can technically be done in any order, but the epic ordering places it here.

### Verification Checklist

After `make bundle`:
1. `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` should exist
2. The `alm-examples` annotation should contain a JSON array with entries for `Entity` and `EntityAlias`
3. `operator-sdk bundle validate ./bundle` should not emit warnings about `Entity` or `EntityAlias` missing examples
4. The remaining warnings (R1.8: empty owned CRD descriptions for `AzureSecretEngineConfig`/`Entity`/`EntityAlias`; R1.9: missing `spec.minKubeVersion`) are expected and are separate stories

### Previous Story Intelligence

**From R1.2c (immediately preceding in ordering):**
- Verification-only story confirming lint compliance
- No source code changes — this is also a metadata-only story
- `make integration` takes ~576-579s — but NOT required for this story

**From Epic R1 preamble:**
- Community Operators PR `#9655` surfaced these bundle metadata warnings
- Stories R1.7-R1.9 address the three addressable warning categories
- The `operatorhub` validator deprecation warning and FBC migration recommendation are explicitly out of scope

### Project Structure Notes

- Changes confined to `config/samples/kustomization.yaml` (add 2 lines) and regenerated `bundle/` output
- No new files created — sample files already exist
- No Go source changes — no `make generate` or `make manifests` needed beyond what `make bundle` already runs
- The `bundle/` directory is generated output that is tracked in git

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.7] — acceptance criteria, task list
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, story ordering, Community Operators PR #9655 context
- [Source: config/samples/kustomization.yaml] — current resources list (47 entries, missing Entity/EntityAlias)
- [Source: config/samples/redhatcop_v1alpha1_entity.yaml] — existing Entity sample
- [Source: config/samples/redhatcop_v1alpha1_entityalias.yaml] — existing EntityAlias sample
- [Source: config/manifests/kustomization.yaml] — manifests assembly (includes ../samples)
- [Source: config/manifests/bases/vault-config-operator.clusterserviceversion.yaml:5] — `alm-examples: '[]'` placeholder
- [Source: Makefile:350-355] — `make bundle` target definition
- [Source: api/v1alpha1/entity_types.go] — Entity CRD schema (EntitySpec, EntityConfig)
- [Source: api/v1alpha1/entityalias_types.go] — EntityAlias CRD schema (EntityAliasSpec, EntityAliasConfig)
- [Source: _bmad-output/implementation-artifacts/R1-2c-lint-green-gate-verify-full-compliance.md] — previous story context

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues encountered — straightforward metadata-only story.

### Completion Notes List

- Task 1: Confirmed the bundle pipeline — samples in kustomization.yaml → manifests assembly → `operator-sdk generate bundle` → `alm-examples` annotation in CSV. Only change needed was adding 2 entries to `config/samples/kustomization.yaml`.
- Task 2: Entity sample already valid against EntitySpec schema — all required fields present, no normalization needed.
- Task 3: EntityAlias sample already valid against EntityAliasSpec schema — all required fields present, no normalization needed.
- Task 4: Added `redhatcop_v1alpha1_entity.yaml` and `redhatcop_v1alpha1_entityalias.yaml` to `config/samples/kustomization.yaml` before the scaffold marker.
- Task 5: `make bundle` regenerated successfully. CSV at `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` now contains Entity (line 392) and EntityAlias (line 421) in the `alm-examples` annotation. `operator-sdk bundle validate ./bundle` passed with "All validation tests have completed successfully" — zero warnings.
- Task 6: Changes committed as single commit covering kustomization change and regenerated bundle output.

### Change Log

- 2026-06-20: Added Entity and EntityAlias sample references to config/samples/kustomization.yaml; regenerated bundle manifests; bundle validation passes with no example-annotation warnings.

### File List

- config/samples/kustomization.yaml (modified — added 2 resource entries for entity and entityalias)
- _bmad-output/implementation-artifacts/R1-7-bundle-example-annotations-for-entity-and-entityalias.md (modified — story tracking)
- _bmad-output/implementation-artifacts/sprint-status.yaml (modified — status tracking)
