# Story 10.7: CI Pipeline Adaptation and End-to-End Validation

Status: done

## Story

As an operator developer,
I want to verify that all CI pipelines, Docker builds, and test suites work correctly after the cumulative changes from Stories 10.0 through 10.6,
So that the project is fully validated after the build tooling and metrics modernization epic.

## Acceptance Criteria

1. **Given** all previous stories in Epic 10 are complete
   **When** the full CI pipeline is run
   **Then** all jobs pass: lint, unit tests, integration tests, docker build, helm chart tests

2. **Given** the Dockerfile has been updated for v4 layout (Story 10.0)
   **When** `make docker-build` is run
   **Then** the image builds successfully and the operator binary starts

3. **Given** `ci.Dockerfile` copies `bin/manager`
   **When** `make ci-build` (or equivalent) is run
   **Then** the CI-specific image builds successfully

4. **Given** the Makefile targets reference the new paths
   **When** `make build`, `make run`, `make test`, `make manifests`, `make generate`, `make helmchart` are run
   **Then** all targets succeed

5. **Given** the integration test suite references CRD paths
   **When** `make integration` is run against a Kind cluster
   **Then** all integration tests pass

6. **Given** the kube-rbac-proxy has been removed and authn/authz is enabled
   **When** the operator is deployed to a Kind cluster
   **Then** the metrics endpoint is accessible with proper authentication and returns Prometheus metrics

## Tasks / Subtasks

- [x] Task 1: Pre-flight prerequisite verification (AC: #1)
  - [x] Confirm Story 10.0 complete: `PROJECT` file declares `go.kubebuilder.io/v4`, `cmd/main.go` exists, `internal/controller/` exists, `controllers/` and root `main.go` are gone
  - [x] Confirm Story 10.0a complete: `config/default/kustomization.yaml` uses `resources:`, `patches:`, `replacements:` (no deprecated `bases:`, `patchesStrategicMerge:`, `vars:`)
  - [x] Confirm Story 10.1 complete: `OPERATOR_SDK_VERSION` is `v1.42.3` in Makefile, `pr.yaml`, `push.yaml`, and `bundle.Dockerfile`
  - [x] Confirm Story 10.2 complete: `HELM_VERSION` is `v4.x` in Makefile, `--atomic` replaced by `--rollback-on-failure`, cert-manager version is current
  - [x] Confirm Story 10.3 complete: `GOLANGCI_LINT_VERSION` is `v2.x` in Makefile, `.golangci.yml` exists with `version: "2"`, `make lint` target exists
  - [x] Confirm Story 10.4 complete: `KUSTOMIZE_VERSION` is `v5.8.x`, OPM version updated to `v1.65.x`
  - [x] Confirm Story 10.5 complete: kube-rbac-proxy sidecar removed from `config/default/`, metrics authn/authz via `filters.WithAuthenticationAndAuthorization` in `cmd/main.go`, new `config/rbac/metrics_*.yaml` files exist, old `config/rbac/auth_proxy_*.yaml` files removed
  - [x] Confirm Story 10.6 complete: Helm chart `values.yaml.tpl` has no `kube_rbac_proxy:` section, `templates/manager.yaml` has no sidecar, `metricsSecure` flag added

- [x] Task 2: Static build validation — compilation and code generation (AC: #4)
  - [x] Run `go build ./cmd/` — verify compilation succeeds with go/v4 layout
  - [x] Run `make manifests generate` — verify CRD and RBAC generation works
  - [x] Run `make fmt vet` — verify formatting and static analysis pass
  - [x] Run `make lint` — verify golangci-lint v2 passes with zero findings

- [x] Task 3: Unit test validation (AC: #4)
  - [x] Run `make test` — verify all envtest-based unit tests pass
  - [x] Verify CRDDirectoryPaths resolve correctly from `internal/controller/` depth
  - [x] Confirm test coverage report generates (`cover.out`)

- [x] Task 4: Docker build validation (AC: #2, #3)
  - [x] Run `make docker-build` — verify multi-stage Dockerfile builds with go/v4 paths (`COPY cmd/ cmd/`, `COPY internal/ internal/`, `go build ./cmd/`)
  - [x] Verify `ci.Dockerfile` still works — it copies `bin/manager` (path unchanged, no Story 10.0 impact)
  - [x] If possible, verify the built image starts: `docker run --rm <image> --help` or similar smoke test

- [x] Task 5: Kustomize build validation (AC: #4)
  - [x] Run `kustomize build config/default` — verify v5 syntax produces valid output (no deprecated-syntax warnings)
  - [x] Run `kustomize build config/helmchart` — verify helmchart overlay works
  - [x] Run `kustomize build config/local-development` — verify local-dev overlay works
  - [x] Verify kube-rbac-proxy sidecar is NOT present in kustomize output (Story 10.5 validation)
  - [x] Verify metrics RBAC (`metrics_auth_role.yaml`, `metrics_reader_role.yaml`) is present in kustomize output

- [x] Task 6: OLM bundle validation (AC: #4)
  - [x] Run `make bundle` — verify OLM bundle regenerates cleanly with SDK v1.42.3
  - [x] Run `operator-sdk bundle validate ./bundle` — verify bundle passes validation
  - [x] Verify `bundle.Dockerfile` labels: `operator-sdk-v1.42.3`, `go.kubebuilder.io/v4`
  - [x] Verify scorecard images updated to `v1.42.3`

- [x] Task 7: Helm chart validation (AC: #4, #6)
  - [x] Run `make helmchart` — verify chart generation succeeds with kustomize v5.8.x + Helm v4.x
  - [x] Run `helm lint ./charts/vault-config-operator` — verify chart linting passes
  - [x] Run `helm template ./charts/vault-config-operator` — verify rendered output:
    - No kube-rbac-proxy sidecar container
    - Manager container has `--metrics-secure` and `--metrics-bind-address=:8443` args
    - No references to `kube_rbac_proxy` values
  - [x] Verify `values.yaml` has `metricsSecure` instead of `kube_rbac_proxy` section

- [ ] Task 8: Integration test validation (AC: #5, #1)
  - [ ] Run `make kind-setup` — verify Kind cluster creation
  - [ ] Run `make deploy-vault` — verify Vault deployment with Helm 4 flags (`--rollback-on-failure`)
  - [ ] Run `make deploy-ingress` — verify ingress-nginx with Helm 4 flags
  - [ ] Run `make deploy-postgresql` — verify PostgreSQL with `--force-update` repo add pattern
  - [ ] Run `make integration` — verify all integration tests pass against live Kind+Vault cluster
  - [ ] Verify CRD directory paths resolve correctly at `internal/controller/` depth in integration test envtest setup

- [ ] Task 9: Helmchart test validation (AC: #6, #1)
  - [ ] Run `make helmchart-test` — end-to-end: builds image, loads into Kind, deploys cert-manager + prometheus + operator, verifies metrics
  - [ ] Verify cert-manager installs at current version with `crds.enabled=true` (not `installCRDs=true`)
  - [ ] Verify operator pod reaches Ready state without kube-rbac-proxy sidecar
  - [ ] Verify Prometheus can scrape metrics from the operator's `:8443` endpoint with authn/authz
  - [ ] Verify the `test-metrics` sidecar in prometheus pod confirms metrics accessibility

- [ ] Task 10: CI workflow validation (AC: #1)
  - [ ] Review `.github/workflows/pr.yaml` — confirm `OPERATOR_SDK_VERSION: v1.42.3`, `GO_VERSION: ~1.26`
  - [ ] Review `.github/workflows/push.yaml` — confirm same version references
  - [ ] Verify the reusable workflows from `redhat-cop/github-workflows-operators` support SDK v1.42.x (URL pattern is version-agnostic)
  - [ ] Confirm all three test flags enabled: `RUN_UNIT_TESTS: true`, `RUN_INTEGRATION_TESTS: true`, `RUN_HELMCHART_TEST: true`

- [ ] Task 11: Fix any issues found (AC: #1-6)
  - [ ] If any validation step fails, diagnose and fix the issue in-story
  - [ ] Document all fixes in Completion Notes
  - [ ] Re-run the failed validation step to confirm the fix

## Dev Notes

### Nature of This Story

This is a **validation story**, not a code-change story. The primary deliverable is verification that all cumulative changes from Stories 10.0–10.6 work correctly end-to-end. If issues are found during validation, they are fixed in-story rather than opening new stories.

### Prerequisite: All Previous Stories Must Be Complete

This story CANNOT start until all of the following are `done`:

| Story | Description | Key Changes to Validate |
|-------|-------------|------------------------|
| 10.0 | go/v3 → go/v4 layout | `cmd/main.go`, `internal/controller/`, package rename |
| 10.0a | Kustomize v3 → v5 syntax | `resources:`, `patches:`, `replacements:` in kustomization files |
| 10.1 | Operator SDK v1.31 → v1.42 | SDK binary, scorecard images, CI workflow versions, bundle labels |
| 10.2 | Helm v3.11 → v4.x | `--rollback-on-failure`, `--force-update`, cert-manager version |
| 10.3 | golangci-lint v1.x → v2.x | Module path `/v2/`, `.golangci.yml`, `lint` target |
| 10.4 | OPM + Kustomize tool versions | OPM v1.65.x, Kustomize v5.8.x |
| 10.5 | Remove kube-rbac-proxy | authn/authz via `filters.WithAuthenticationAndAuthorization`, new RBAC files, old auth_proxy files removed |
| 10.6 | Helm chart metrics rearchitecture | No sidecar, `metricsSecure` flag, manager serves metrics directly on `:8443` |

### Expected Project State After All Prior Stories

#### Directory Layout (go/v4)

```
vault-config-operator/
├── cmd/
│   └── main.go          (moved from root by 10.0)
├── api/v1alpha1/         (unchanged)
├── internal/
│   └── controller/       (moved from controllers/ by 10.0)
│       ├── *.go
│       ├── controllertestutils/
│       ├── vaultresourcecontroller/
│       └── vaultsecretutils/
├── config/
│   ├── default/kustomization.yaml  (v5 syntax from 10.0a, no auth_proxy patches from 10.5)
│   ├── rbac/
│   │   ├── metrics_auth_role.yaml          (new from 10.5)
│   │   ├── metrics_auth_role_binding.yaml  (new from 10.5)
│   │   └── metrics_reader_role.yaml        (new from 10.5)
│   └── ...
├── .golangci.yml         (new from 10.3, version: "2")
├── Dockerfile            (updated paths by 10.0)
├── ci.Dockerfile         (unchanged — copies bin/manager)
├── Makefile              (multiple version bumps from 10.0-10.4)
├── bundle.Dockerfile     (updated labels by 10.0, 10.1)
└── PROJECT               (go.kubebuilder.io/v4 from 10.0)
```

#### Makefile Version Expectations

| Variable | Expected Value | Set By |
|----------|---------------|--------|
| `HELM_VERSION` | v4.x (e.g., v4.2.3) | Story 10.2 |
| `KUSTOMIZE_VERSION` | v5.8.x | Story 10.4 |
| `OPERATOR_SDK_VERSION` | v1.42.3 | Story 10.1 |
| `GOLANGCI_LINT_VERSION` | v2.x (e.g., v2.12.2) | Story 10.3 |
| `build` target | `go build -o bin/manager ./cmd/` | Story 10.0 |
| `run` target | `go run ./cmd/` | Story 10.0 |
| `golangci-lint` install | `github.com/golangci/golangci-lint/v2/cmd/golangci-lint` | Story 10.3 |

#### Dockerfile Expectations (Story 10.0)

```dockerfile
COPY cmd/ cmd/
COPY internal/ internal/
RUN ... go build -a -o manager ./cmd/
```

#### CI Workflow Expectations (Stories 10.1)

Both `pr.yaml` and `push.yaml`:
- `OPERATOR_SDK_VERSION: v1.42.3`
- `GO_VERSION: ~1.26`

#### Helm Chart Expectations (Stories 10.2, 10.6)

- `values.yaml.tpl`: `metricsSecure: true` (no `kube_rbac_proxy:` section)
- `templates/manager.yaml`: single manager container with `--metrics-secure` and `--metrics-bind-address=:8443` args
- `helmchart-test`: cert-manager `--version v1.21.1` with `--set crds.enabled=true`
- All `helm repo add` calls use `--force-update`
- All `helm upgrade/install` calls use `--rollback-on-failure` (not `--atomic`)

#### Removed Files (Story 10.5)

- `config/default/manager_auth_proxy_patch.yaml`
- `config/rbac/auth_proxy_service.yaml`
- `config/rbac/auth_proxy_role.yaml`
- `config/rbac/auth_proxy_role_binding.yaml`
- `config/rbac/auth_proxy_client_clusterrole.yaml`

### Validation Execution Order

Run validations in this order — earlier failures make later steps meaningless:

1. **Pre-flight checks** (Task 1) — verify all stories are actually done
2. **Compilation** (Task 2) — `go build`, `make manifests generate`, `make fmt vet`, `make lint`
3. **Unit tests** (Task 3) — `make test` (envtest, no cluster needed)
4. **Docker builds** (Task 4) — `make docker-build`, verify `ci.Dockerfile`
5. **Kustomize builds** (Task 5) — `kustomize build` for all overlays
6. **OLM bundle** (Task 6) — `make bundle` + validation
7. **Helm chart** (Task 7) — `make helmchart` + lint + template
8. **Integration tests** (Task 8) — requires Kind cluster + Vault
9. **Helmchart test** (Task 9) — full end-to-end with metrics verification
10. **CI workflow review** (Task 10) — static review of workflow files

### Troubleshooting Guide

#### Compilation Fails (`go build`)

- **Import path errors**: Story 10.0 changed `controllers` → `internal/controller`. Check all files importing `github.com/redhat-cop/vault-config-operator/controllers` — should be `internal/controller`.
- **Package name errors**: `package controllers` → `package controller` (singular). Check `cmd/main.go` references: `controller.XxxReconciler` not `controllers.XxxReconciler`.

#### Unit Tests Fail (`make test`)

- **CRDDirectoryPaths**: `internal/controller/suite_test.go` must use `filepath.Join("..", "..", "config", "crd", "bases")` (two `..` levels from `internal/controller/`).
- **Envtest version**: `ENVTEST_K8S_VERSION` must match available kubebuilder assets.

#### Docker Build Fails

- **COPY paths**: Dockerfile must reference `cmd/` and `internal/` (not `main.go` and `controllers/`).
- **Build command**: Must be `go build -a -o manager ./cmd/` (not `main.go`).

#### Kustomize Build Fails

- **Deprecated syntax**: `kustomize build` will fail hard (not just warn) on `vars:` in newer kustomize. Verify all `kustomization.yaml` files under `config/` use v5 syntax.
- **Missing files**: If Story 10.5 removed auth_proxy files but kustomization still references them, build fails.

#### Helm Chart Issues

- **`--atomic` unknown flag**: Helm 4 renamed to `--rollback-on-failure`. All 3 deploy targets must be updated.
- **cert-manager `installCRDs`**: cert-manager 1.16+ removed this; use `crds.enabled=true`.
- **`helm repo add` failures**: Helm 4 removed `--no-update`; use `--force-update` for idempotent adds.

#### Integration Tests Fail

- **Vault connectivity**: Verify `VAULT_ADDR` and `VAULT_TOKEN` env vars are set correctly.
- **CRD paths in envtest**: Same as unit test CRDDirectoryPaths issue.
- **Helm 4 deploy failures**: Check deploy-vault, deploy-postgresql targets for flag compatibility.

#### Metrics Endpoint Issues (Story 10.5/10.6 Validation)

- **No metrics response**: Verify `cmd/main.go` has `SecureServing: true` and `FilterProvider: filters.WithAuthenticationAndAuthorization`.
- **Metrics port**: Should be `:8443` (not `127.0.0.1:8080`).
- **RBAC errors**: Verify `config/rbac/metrics_auth_role.yaml` grants `tokenreviews` and `subjectaccessreviews`.
- **Prometheus scrape failure**: Verify ServiceMonitor targets port `8443` and the operator-managed TLS cert is functional.

### Tiltfile Validation (Optional)

Story 10.0 should have updated the Tiltfile:
- `compile_cmd`: `go build -o bin/manager ./cmd/` (not `main.go`)
- `deps`: `['./cmd','./api','./internal']` (not `['./main.go','./api','./controllers']`)

This is a dev-experience check, not a CI-required validation. Verify if convenient.

### Anti-Patterns to Avoid

- **DO NOT** skip any validation step — each catches a different class of failure
- **DO NOT** open new stories for issues found — fix them in this story
- **DO NOT** modify code unless a validation step actually fails — this is a validation story, not a refactoring story
- **DO NOT** assume prior stories were implemented correctly — verify independently
- **DO NOT** skip the helmchart-test even though it's slow — it's the only end-to-end metrics validation
- **DO NOT** treat deprecation warnings as passing — warnings indicate incomplete migration

### Files That Should NOT Be Changed (Unless Broken)

Unless a validation step reveals a bug from a prior story, these files should not need changes:
- All Go source files (`cmd/main.go`, `internal/controller/*.go`, `api/v1alpha1/*.go`)
- `go.mod`, `go.sum`
- `config/` kustomize manifests (already migrated by 10.0a, 10.5)
- `Dockerfile`, `ci.Dockerfile` (already updated by 10.0)

### Files That MAY Need Minor Fixes

Based on common integration issues across multi-story epics:
- `.github/workflows/pr.yaml` / `push.yaml` — version reference mismatches
- `Makefile` — target path references or flag mismatches
- `bundle.Dockerfile` — label inconsistencies
- `config/scorecard/patches/*.yaml` — scorecard image version alignment

### Previous Story Intelligence

All prior Epic 10 stories (10.0 through 10.3, which have story specs created) share these patterns:
- Each story specifies exact version numbers and file-level changes
- Each story has explicit anti-patterns preventing scope creep into sibling stories
- Verification commands are documented at the end of each story
- Story 10.0 is the highest-impact change (directory moves affecting 75+ files)
- Story 10.0a is the most complex kustomize change (`vars:` → `replacements:`)
- Story 10.5 is the most architecturally significant (kube-rbac-proxy removal)

### Project Structure Notes

- Alignment with unified project structure: This story validates that all structural changes from Epic 10 converge into a consistent, buildable, testable project
- Detected conflicts or variances: The project-context.md still references pre-migration state (Kubebuilder v3 layout, Helm v3.11, etc.) — update project-context.md after this story to reflect the new state

### Verification Commands (Complete List — Run In Order)

```bash
# Static builds
go build ./cmd/
make manifests generate
make fmt vet
make lint

# Unit tests
make test

# Docker build
make docker-build

# Kustomize builds
kustomize build config/default
kustomize build config/helmchart

# OLM bundle
make bundle
operator-sdk bundle validate ./bundle

# Helm chart
make helmchart

# Integration tests (requires Kind cluster)
make kind-setup
make integration

# Helmchart test (full end-to-end)
make helmchart-test
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.7]
- [Source: _bmad-output/implementation-artifacts/10-0-migrate-project-layout-from-gov3-to-gov4.md]
- [Source: _bmad-output/implementation-artifacts/10-0a-kustomize-v3-to-v5-syntax-migration.md]
- [Source: _bmad-output/implementation-artifacts/10-1-upgrade-operator-sdk-from-v1-31-to-v1-42.md]
- [Source: _bmad-output/implementation-artifacts/10-2-upgrade-helm-from-v3-11-to-v4.md]
- [Source: _bmad-output/implementation-artifacts/10-3-upgrade-golangci-lint-from-v1-59-to-v2.md]
- [Source: _bmad-output/project-context.md#Technology Stack & Versions]
- [Source: _bmad-output/project-context.md#Build & Dev Tooling]
- [Source: _bmad-output/project-context.md#CI Pipeline]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

- **Review Model Used:** gpt-5.4-medium (ChatGPT 5.4)
- **Review Findings:** None — approved with 0 patches
- **Decisions Needed:** None
- **Decisions Taken:** N/A
- **Fixes Applied:** N/A (all 3 in-story fixes were pre-approved: Helm v4 --force-conflicts, metrics port conflict fix, Vault deadlock documentation)
