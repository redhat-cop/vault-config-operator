# Story 10.1: Upgrade Operator SDK from v1.31 to v1.42

Status: ready-for-dev

## Story

As an operator developer,
I want to upgrade Operator SDK from v1.31.0 to v1.42.3,
So that we benefit from improved bundle generation, scorecard validation, and compatibility with the latest OLM ecosystem.

## Acceptance Criteria

1. **Given** `OPERATOR_SDK_VERSION` is `v1.31.0` in the Makefile
   **When** it is updated to `v1.42.3`
   **Then** `make operator-sdk` downloads the correct binary and `$(LOCALBIN)/operator-sdk --version` reports `v1.42.3`

2. **Given** CI workflows pass `OPERATOR_SDK_VERSION: v1.31.0` to reusable workflows
   **When** both `pr.yaml` and `push.yaml` are updated to `v1.42.3`
   **Then** CI uses the correct SDK version for bundle generation and validation

3. **Given** `bundle.Dockerfile` contains the label `operator-sdk-v1.31.0`
   **When** it is updated to `operator-sdk-v1.42.3`
   **Then** the bundle image metadata reflects the correct SDK version

4. **Given** scorecard test images reference `quay.io/operator-framework/scorecard-test:v1.13.0`
   **When** they are updated to `v1.42.3`
   **Then** scorecard tests use the matching SDK version images

5. **Given** all version references are updated
   **When** `make bundle` is run
   **Then** the OLM bundle regenerates cleanly and `operator-sdk bundle validate ./bundle` passes

6. **Given** all version references are updated
   **When** `make manifests generate` and `make build` are run
   **Then** both succeed without errors

## Tasks / Subtasks

- [ ] Task 1: Update `OPERATOR_SDK_VERSION` in Makefile (AC: #1)
  - [ ] Change `OPERATOR_SDK_VERSION ?= v1.31.0` → `OPERATOR_SDK_VERSION ?= v1.42.3`
  - [ ] Delete stale binary: `rm -f bin/operator-sdk` (force re-download)
  - [ ] Run `make operator-sdk` and verify `bin/operator-sdk version` reports v1.42.3

- [ ] Task 2: Update CI workflow files (AC: #2)
  - [ ] `.github/workflows/pr.yaml`: change `OPERATOR_SDK_VERSION: v1.31.0` → `OPERATOR_SDK_VERSION: v1.42.3`
  - [ ] `.github/workflows/push.yaml`: change `OPERATOR_SDK_VERSION: v1.31.0` → `OPERATOR_SDK_VERSION: v1.42.3`

- [ ] Task 3: Update `bundle.Dockerfile` SDK version label (AC: #3)
  - [ ] Change `operators.operatorframework.io.metrics.builder=operator-sdk-v1.31.0` → `operator-sdk-v1.42.3`
  - [ ] Note: The `project_layout` label (`go.kubebuilder.io/v3` → `go.kubebuilder.io/v4`) should already be updated by Story 10.0; verify it is correct, fix if not

- [ ] Task 4: Update scorecard test images (AC: #4)
  - [ ] `config/scorecard/patches/olm.config.yaml`: change all `scorecard-test:v1.13.0` → `scorecard-test:v1.42.3` (5 image references)
  - [ ] `config/scorecard/patches/basic.config.yaml`: change `scorecard-test:v1.13.0` → `scorecard-test:v1.42.3` (1 image reference)

- [ ] Task 5: Regenerate OLM bundle and validate (AC: #5)
  - [ ] Run `make bundle` — this regenerates bundle manifests with the new SDK
  - [ ] Review generated diff in `bundle/` for unexpected changes
  - [ ] Verify `operator-sdk bundle validate ./bundle` passes
  - [ ] Check that `bundle/metadata/annotations.yaml` is correct (SDK version, layout, channels)

- [ ] Task 6: Verify build and code generation (AC: #6)
  - [ ] Run `make manifests generate` — confirms code-gen still works
  - [ ] Run `make build` — confirms compilation
  - [ ] Run `make fmt vet` — confirms formatting/linting
  - [ ] Run `make test` — confirms unit tests pass

## Dev Notes

### Scope and Dependencies

This story upgrades the Operator SDK **binary version only**. It does NOT cover the scaffolding-level changes introduced in SDK v1.38+ (kube-rbac-proxy removal, metrics authn/authz, network policies, cert-manager TLS). Those changes are handled by:

| Change | Story |
|--------|-------|
| go/v3 → go/v4 layout migration | Story 10.0 |
| Kustomize v3 → v5 syntax migration | Story 10.0a |
| kube-rbac-proxy removal + authn/authz | Story 10.5 |
| Helm chart kube-rbac-proxy removal | Story 10.6 |
| golangci-lint v1 → v2 | Story 10.3 |
| OPM + kustomize tool version upgrades | Story 10.4 |

**Prerequisites:** Stories 10.0 and 10.0a must be completed first. The SDK v1.42 expects go/v4 layout and kustomize v5 syntax.

### Operator SDK Upgrade Path (v1.31 → v1.42): What's Relevant

The Operator SDK is backwards-compatible across minor versions post-1.0. No mandatory code migrations exist for the SDK binary itself between v1.31 and v1.42. Key SDK releases in this range:

- **v1.32–v1.37**: No Go operator migrations
- **v1.38**: Introduced scaffolding for kube-rbac-proxy removal and metrics authn/authz (handled by Story 10.5)
- **v1.39**: Added network policy scaffolding (optional, not required for existing projects)
- **v1.40**: Bumped OPM to v1.55, added certwatcher TLS support, envtest version automation (OPM handled by Story 10.4, certwatcher is optional)
- **v1.41**: Bumped golangci-lint to v2, controller-gen to v0.18, controller-runtime to v0.21 (handled by Story 10.3 and already done in Epic 8)
- **v1.42.0–v1.42.3**: No migrations required

**Bottom line:** This story is a straightforward version bump of the SDK binary, scorecard images, and CI references. The heavy lifting (layout, kustomize, metrics) is in sibling stories.

### Latest Stable Version

The latest stable Operator SDK release is **v1.42.3** (released June 26, 2026). Use this specific patch version (not just v1.42) to pick up all bug fixes.

### Files to Change (Complete List)

| File | Change |
|------|--------|
| `Makefile` | `OPERATOR_SDK_VERSION ?= v1.31.0` → `v1.42.3` |
| `.github/workflows/pr.yaml` | `OPERATOR_SDK_VERSION: v1.31.0` → `v1.42.3` |
| `.github/workflows/push.yaml` | `OPERATOR_SDK_VERSION: v1.31.0` → `v1.42.3` |
| `bundle.Dockerfile` | `operator-sdk-v1.31.0` → `operator-sdk-v1.42.3` (line 9) |
| `config/scorecard/patches/olm.config.yaml` | `scorecard-test:v1.13.0` → `v1.42.3` (5 occurrences) |
| `config/scorecard/patches/basic.config.yaml` | `scorecard-test:v1.13.0` → `v1.42.3` (1 occurrence) |

### Files NOT to Change

- `cmd/main.go` — No SDK-driven `main.go` changes in this story (metrics/TLS changes are Story 10.5)
- `config/default/kustomization.yaml` — kustomize syntax changes are Story 10.0a; kube-rbac-proxy patches are Story 10.5
- `config/rbac/` — auth_proxy removal is Story 10.5
- `go.mod` — No Go dependency changes driven by the SDK binary upgrade itself (dependencies are managed by Epics 8 and 9)
- `Dockerfile` / `ci.Dockerfile` — Layout path changes are Story 10.0; no SDK-driven Dockerfile changes
- `.golangci.yml` — golangci-lint v2 migration is Story 10.3

### Scorecard Image Versioning

The scorecard test images (`quay.io/operator-framework/scorecard-test`) should match the SDK version. The current images are at `v1.13.0` (extremely old — predates the operator by years). Updating to `v1.42.3` ensures compatibility with the latest OLM bundle validation.

### Bundle Regeneration

Running `make bundle` with the new SDK may produce minor diff in generated files:
- `bundle/metadata/annotations.yaml` may update format
- `bundle/manifests/` ClusterServiceVersion may have minor cosmetic changes
- The custom `com.redhat.openshift.versions: v4.16` annotation is appended after bundle generation (Makefile line 346)

Review any diff carefully — the bundle content should be semantically identical. If the CSV changes structurally, investigate whether it's a new SDK requirement.

### CI Integration

The CI workflows use reusable workflows from `redhat-cop/github-workflows-operators`. The `OPERATOR_SDK_VERSION` input is passed through to those workflows, which download the SDK binary at that version. Ensure the reusable workflow supports v1.42.x (it should, as the download URL pattern is version-agnostic).

### Anti-Patterns to Avoid

- **DO NOT** apply kube-rbac-proxy removal changes — that's Story 10.5
- **DO NOT** modify `cmd/main.go` for metrics/TLS — that's Story 10.5
- **DO NOT** change Go dependencies or controller-runtime version — those are already up-to-date from Epics 8/9
- **DO NOT** update golangci-lint version — that's Story 10.3
- **DO NOT** update OPM or kustomize tool versions — that's Story 10.4
- **DO NOT** change kustomize manifest syntax — that's Story 10.0a
- **DO NOT** move directories or change import paths — that's Story 10.0

### Verification Commands (Run After Completion)

```bash
rm -f bin/operator-sdk
make operator-sdk
bin/operator-sdk version
make manifests generate
make bundle
make fmt vet
make test
```

### Project Structure Notes

- Alignment with unified project structure: No structural changes — this story only bumps version numbers
- Detected conflicts or variances: The scorecard images (`v1.13.0`) are significantly behind the SDK version — this story corrects that mismatch

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.1]
- [Source: _bmad-output/project-context.md#Technology Stack & Versions]
- [Source: _bmad-output/implementation-artifacts/10-0-migrate-project-layout-from-gov3-to-gov4.md]
- [Operator SDK v1.42.3 release](https://github.com/operator-framework/operator-sdk/releases/tag/v1.42.3)
- [Operator SDK upgrade guide](https://v1-42-x.sdk.operatorframework.io/docs/upgrading-sdk-version/)

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Previous Review Notes (First Attempt — GPT 5.4)

> These findings are from a code review of the first implementation attempt. Commit references are no longer valid but the architectural guidance remains relevant.

- **[Patch]** The Makefile `operator-sdk` target only downloads if `bin/operator-sdk` is missing. It does not check version — a stale cached binary at the old version will silently be reused. Add version-suffix caching (e.g., `bin/operator-sdk-$(OPERATOR_SDK_VERSION)` + symlink) consistent with other tool targets.
- **[Patch]** AC5 and AC6 require evidence that `make bundle`, `operator-sdk bundle validate`, `make manifests generate`, `make build`, `make fmt vet`, and `make test` were actually run. Record this evidence in the Dev Agent Record.
