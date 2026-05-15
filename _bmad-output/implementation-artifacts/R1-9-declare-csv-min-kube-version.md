# Story R1.9: Declare CSV `minKubeVersion`

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator maintainer,
I want the generated CSV to declare an explicit minimum supported Kubernetes version,
So that release metadata reflects the tested support floor instead of implying support for every possible cluster version.

## Acceptance Criteria

1. **Given** Community Operators validation warns that `csv.Spec.minKubeVersion` is not informed **When** a support floor is selected from the project's real test and toolchain constraints **Then** the CSV declares an explicit `spec.minKubeVersion`
2. **Given** the project currently tests against a concrete envtest/Kubernetes toolchain baseline **When** the chosen minimum version is documented in the story implementation **Then** future maintainers can understand why that floor was selected
3. **Given** `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` is the CSV base **When** bundle metadata is regenerated **Then** the generated bundle preserves the declared `minKubeVersion`
4. **Given** `make bundle` runs after the update **When** bundle validation completes **Then** the missing `minKubeVersion` warning is no longer emitted

## Tasks / Subtasks

- [ ] Task 1: Determine the supported Kubernetes floor (AC: 1, 2)
  - [ ] 1.1: Confirm K8s dependency baseline from `go.mod` — `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go` are all pinned at `v0.29.2` (maps to Kubernetes 1.29)
  - [ ] 1.2: Confirm envtest version from `Makefile:19` — `ENVTEST_K8S_VERSION = 1.29.0`
  - [ ] 1.3: Confirm Kind integration node image — `kindest/node:v1.29.0` via `KUBECTL_VERSION` in `Makefile:8`
  - [ ] 1.4: Confirm controller-runtime `v0.17.3` targets K8s 1.29 (controller-runtime 0.17.x release notes)
  - [ ] 1.5: Select `1.29.0` as the `minKubeVersion` value — this is the exact version the project compiles against and tests on
- [ ] Task 2: Add `spec.minKubeVersion` to the CSV base (AC: 1, 3)
  - [ ] 2.1: Edit `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`
  - [ ] 2.2: Insert `minKubeVersion: "1.29.0"` at line 268, between `maturity: alpha` (line 267) and `provider:` (line 268) — alphabetically sorted alongside sibling spec fields
- [ ] Task 3: Regenerate bundle and validate (AC: 3, 4)
  - [ ] 3.1: Run `make bundle`
  - [ ] 3.2: Inspect `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` — confirm `minKubeVersion: "1.29.0"` appears in the `spec` section
  - [ ] 3.3: Run `operator-sdk bundle validate ./bundle` and verify the missing `minKubeVersion` warning is gone
- [ ] Task 4: Commit (AC: 1, 3)
  - [ ] 4.1: Commit the CSV base change (`config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`)
  - [ ] 4.2: Commit the regenerated bundle output (`bundle/` directory is tracked in git)

## Dev Notes

### This Is a Metadata-Only Story

Per the epic rule: "Every story must pass `make manifests generate fmt vet test` and `make integration` unless the story is metadata-only, in which case `make bundle` is the minimum required verification."

This story changes **zero Go source files**. The only changes are:
1. `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` — add 1 field (`minKubeVersion`)
2. Regenerated `bundle/` output from `make bundle`

No `make test` or `make integration` run is required, but `make bundle` (which includes `make manifests`) is the minimum gate.

### Why `1.29.0`

All three independent version signals converge on **Kubernetes 1.29**:

| Signal | Value | Source |
|--------|-------|--------|
| `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go` | `v0.29.2` | `go.mod` lines 16–19 |
| `ENVTEST_K8S_VERSION` | `1.29.0` | `Makefile` line 19 |
| Kind node image (`KUBECTL_VERSION`) | `v1.29.0` | `Makefile` line 8, used at lines 162/167/171 |
| controller-runtime | `v0.17.3` | `go.mod` line 20 — targets K8s 1.29 |

The CSV `minKubeVersion` field uses the format `"X.Y.Z"` (string, quoted). OLM semver-parses this value.

### How `minKubeVersion` Flows Into the Bundle

```
config/manifests/bases/vault-config-operator.clusterserviceversion.yaml
  spec.minKubeVersion: "1.29.0"       ← Manual entry (CSV base)
         ↓
kustomize build config/manifests | operator-sdk generate bundle
         ↓
bundle/manifests/vault-config-operator.clusterserviceversion.yaml
  spec.minKubeVersion: "1.29.0"       ← Preserved in generated CSV
         ↓
operator-sdk bundle validate ./bundle  ← No longer warns about missing minKubeVersion
```

The `kustomize build` + `operator-sdk generate bundle` pipeline preserves all `spec.*` fields from the base CSV. No special kustomize patch is needed.

### Exact Insertion Point in the CSV Base

The CSV base file ends with these spec-level fields (lines 267–270):

```yaml
  maturity: alpha
  provider:
    name: Red Hat Community of Practice
  version: 0.1.0
```

Insert `minKubeVersion: "1.29.0"` between `maturity` and `provider` to maintain alphabetical order of spec-level keys:

```yaml
  maturity: alpha
  minKubeVersion: "1.29.0"
  provider:
    name: Red Hat Community of Practice
  version: 0.1.0
```

The indentation is 2 spaces (matching all other spec-level fields in the file).

### What NOT to Do

- Do NOT modify any Go source files — this is metadata-only
- Do NOT change `ENVTEST_K8S_VERSION`, `KUBECTL_VERSION`, or any Makefile variables — the version floor is derived from them, not the other way around
- Do NOT use a bare `1.29` without the patch version — OLM expects semver format `X.Y.Z`
- Do NOT add owned CRD descriptions — that was R1.8's scope
- Do NOT modify `alm-examples` — that was R1.7's scope
- Do NOT create a `.golangci.yml` config file
- Do NOT run `make test` or `make integration` — this story only requires `make bundle` as the verification gate

### Prerequisites

Per the epic ordering: R1.1 → R1.2a → R1.2b → R1.3 → R1.2c → R1.7 → R1.8 → **R1.9** → R1.4 → R1.5 → R1.6. This story follows R1.8 (owned CRD descriptions). However, R1.9 has no code dependencies on R1.1–R1.8 — it touches only metadata files. It can technically be done in any order, but the epic ordering places it here.

### Verification Checklist

After `make bundle`:
1. `bundle/manifests/vault-config-operator.clusterserviceversion.yaml` should exist
2. The `spec` section should contain `minKubeVersion: "1.29.0"`
3. `operator-sdk bundle validate ./bundle` should not emit the missing `minKubeVersion` warning
4. All other existing bundle content should be unchanged (no unintended diffs)

### Previous Story Intelligence

**From R1.8 (immediately preceding in ordering):**
- Also a metadata-only story — added 3 owned CRD descriptions to the CSV base
- Changed `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` (same file this story edits)
- `make bundle` is the verification gate (same as this story)
- R1.8 confirmed that `bundle/` directory is tracked in git and must be committed
- R1.8 explicitly noted: "Do NOT add `spec.minKubeVersion` — that is R1.9's scope"

**From Epic R1 preamble:**
- Community Operators PR `#9655` surfaced bundle metadata warnings including missing `minKubeVersion`
- Stories R1.7-R1.9 address the three addressable warning categories
- The deprecated `operatorhub` validator warning and the FBC migration recommendation are explicitly out of scope

### Project Structure Notes

- Changes confined to `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` (add 1 line) and regenerated `bundle/` output
- No new files created
- No Go source changes — no `make generate` or `make manifests` needed beyond what `make bundle` already runs
- The `bundle/` directory is generated output that is tracked in git

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.9] — acceptance criteria, task list
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, story ordering, Community Operators PR #9655 context
- [Source: config/manifests/bases/vault-config-operator.clusterserviceversion.yaml:267-270] — CSV base spec tail (maturity/provider/version — insertion point)
- [Source: go.mod:16-20] — K8s client libs v0.29.2, controller-runtime v0.17.3
- [Source: Makefile:8] — KUBECTL_VERSION v1.29.0
- [Source: Makefile:19] — ENVTEST_K8S_VERSION 1.29.0
- [Source: Makefile:350-355] — `make bundle` target definition
- [Source: _bmad-output/implementation-artifacts/R1-8-populate-owned-crd-descriptions-for-community-operators-bundle.md] — previous story context

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### Change Log

### File List
