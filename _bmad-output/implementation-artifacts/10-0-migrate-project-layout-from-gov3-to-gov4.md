# Story 10.0: Migrate Project Layout from go/v3 to go/v4

Status: ready-for-dev

## Story

As an operator developer,
I want to migrate the project directory layout from Kubebuilder go/v3 to go/v4,
So that the project follows the current scaffolding standard and is compatible with newer Operator SDK versions.

## Acceptance Criteria

1. **Given** the project uses go/v3 layout with `main.go` at root and `controllers/` directory
   **When** `main.go` is moved to `cmd/main.go` and `controllers/` is moved to `internal/controller/`
   **Then** all Go import paths are updated and `go build ./cmd/main.go` succeeds

2. **Given** the Makefile references `main.go` and `controllers/`
   **When** the targets are updated to `cmd/main.go` and `internal/controller/`
   **Then** `make build`, `make run`, and `make manifests generate` all succeed

3. **Given** the Dockerfile copies `main.go` and `controllers/`
   **When** the paths are updated to `cmd/main.go` and `internal/controller/`
   **Then** `make docker-build` succeeds

4. **Given** controller test suites reference `CRDDirectoryPaths` with relative paths
   **When** the paths are updated to account for the new `internal/controller/` depth
   **Then** `make test` passes (all unit and controller tests)

5. **Given** the PROJECT file declares `go.kubebuilder.io/v3`
   **When** it is updated to `go.kubebuilder.io/v4`
   **Then** `operator-sdk` commands recognize the project as v4 layout

## Tasks / Subtasks

- [ ] Task 1: Move `main.go` to `cmd/main.go` (AC: #1)
  - [ ] Create `cmd/` directory
  - [ ] Move `main.go` → `cmd/main.go`
  - [ ] Update import paths within `cmd/main.go`: `controllers` → `controller` (package alias change)

- [ ] Task 2: Move `controllers/` to `internal/controller/` (AC: #1)
  - [ ] Create `internal/controller/` directory
  - [ ] Move all files from `controllers/` → `internal/controller/`
  - [ ] Subdirectories move as-is: `controllertestutils/`, `vaultresourcecontroller/`, `vaultsecretutils/`
  - [ ] Update `package controllers` → `package controller` in all moved files
  - [ ] Update all 74 files that import `github.com/redhat-cop/vault-config-operator/controllers` to use `github.com/redhat-cop/vault-config-operator/internal/controller`
  - [ ] Update sub-package imports: `controllers/vaultresourcecontroller` → `internal/controller/vaultresourcecontroller`, `controllers/controllertestutils` → `internal/controller/controllertestutils`, `controllers/vaultsecretutils` → `internal/controller/vaultsecretutils`

- [ ] Task 3: Update `cmd/main.go` references (AC: #1)
  - [ ] Change import from `"github.com/redhat-cop/vault-config-operator/controllers"` to `"github.com/redhat-cop/vault-config-operator/internal/controller"`
  - [ ] Change import from `"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"` to `"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"`
  - [ ] Update all `&controllers.XxxReconciler{...}` references to `&controller.XxxReconciler{...}` (package name changes from `controllers` plural to `controller` singular)

- [ ] Task 4: Update CRDDirectoryPaths in test suites (AC: #4)
  - [ ] `internal/controller/suite_test.go`: change `filepath.Join("..", "config", "crd", "bases")` → `filepath.Join("..", "..", "config", "crd", "bases")`
  - [ ] `internal/controller/suite_integration_test.go`: check and update any relative paths similarly

- [ ] Task 5: Update Dockerfile (AC: #3)
  - [ ] Change `COPY main.go main.go` → `COPY cmd/ cmd/`
  - [ ] Change `COPY controllers/ controllers/` → `COPY internal/ internal/`
  - [ ] Change `go build -a -o manager main.go` → `go build -a -o manager ./cmd/`

- [ ] Task 6: Update Makefile (AC: #2)
  - [ ] Change `go build -o bin/manager main.go` → `go build -o bin/manager ./cmd/`
  - [ ] Change `go run ./main.go` → `go run ./cmd/`

- [ ] Task 7: Update Tiltfile (AC: #2)
  - [ ] Change `compile_cmd` from `go build -o bin/manager main.go` → `go build -o bin/manager ./cmd/`
  - [ ] Change `deps=['./main.go','./api','./controllers']` → `deps=['./cmd','./api','./internal']`

- [ ] Task 8: Update PROJECT file (AC: #5)
  - [ ] Change `layout: go.kubebuilder.io/v3` → `layout: go.kubebuilder.io/v4`

- [ ] Task 9: Update bundle.Dockerfile labels (AC: #5)
  - [ ] Change `operators.operatorframework.io.metrics.project_layout=go.kubebuilder.io/v3` → `go.kubebuilder.io/v4`

- [ ] Task 10: Verify all builds and tests pass (AC: #1-5)
  - [ ] Run `go build ./cmd/` (confirms compilation)
  - [ ] Run `make manifests generate` (confirms code-gen still works)
  - [ ] Run `make fmt vet` (confirms formatting)
  - [ ] Run `make test` (confirms unit tests with envtest pass)

## Dev Notes

### Critical Implementation Order

Execute tasks in this exact order to catch breakage incrementally:
1. Move directories/files first (Tasks 1-2)
2. Fix all import paths (Tasks 2-3) — use `sed` or IDE bulk-rename for the 74+ file updates
3. Update `CRDDirectoryPaths` in test suites (Task 4)
4. Update build files (Tasks 5-8)
5. Run `go build ./cmd/` after each major step to catch errors early

### Import Path Mapping (Critical)

| Old Path | New Path |
|----------|----------|
| `github.com/redhat-cop/vault-config-operator/controllers` | `github.com/redhat-cop/vault-config-operator/internal/controller` |
| `github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller` | `github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller` |
| `github.com/redhat-cop/vault-config-operator/controllers/controllertestutils` | `github.com/redhat-cop/vault-config-operator/internal/controller/controllertestutils` |
| `github.com/redhat-cop/vault-config-operator/controllers/vaultsecretutils` | `github.com/redhat-cop/vault-config-operator/internal/controller/vaultsecretutils` |

### Package Name Change

The package declaration in all files under the moved directory changes:
- `package controllers` → `package controller` (singular, matching go/v4 convention)
- Sub-packages (`controllertestutils`, `vaultresourcecontroller`, `vaultsecretutils`) keep their existing package names

This means `cmd/main.go` references change from `controllers.XxxReconciler` to `controller.XxxReconciler`.

### Scope Assessment

- **75 Go files** in `controllers/` directory (including test files)
- **74 files** across the project reference the `controllers` import path
- **3 sub-packages**: `controllertestutils`, `vaultresourcecontroller`, `vaultsecretutils`
- `main.go` has **~50 controller registrations** referencing the `controllers` package
- Recommended approach: `git mv` for directory moves + `sed -i` for bulk import path replacement

### Files Requiring Changes

| File | Change |
|------|--------|
| `PROJECT` | Layout declaration v3→v4 |
| `cmd/main.go` (moved from root) | Import paths + package references |
| `internal/controller/*.go` (all 75 files) | Package declaration `controllers`→`controller` |
| `internal/controller/*.go` (files with cross-imports) | Internal import paths |
| `Dockerfile` | COPY paths + build command |
| `ci.Dockerfile` | No change needed (copies `bin/manager` — path unchanged) |
| `Makefile` | `build` and `run` targets |
| `Tiltfile` | Compile command + deps list |
| `bundle.Dockerfile` | Layout label |
| `internal/controller/suite_test.go` | CRDDirectoryPaths relative path (add one `..`) |
| `internal/controller/suite_integration_test.go` | Check for relative paths |

### Files NOT Requiring Changes

- `ci.Dockerfile` — copies pre-built `bin/manager`, no source paths
- `.github/workflows/*.yaml` — uses reusable workflows (no source path references)
- `api/` directory — stays in place (unchanged by go/v4 migration)
- `config/` directory — no source path references to controllers

### CRDDirectoryPaths Depth Explanation

Current: `controllers/suite_test.go` → `filepath.Join("..", "config", "crd", "bases")` resolves from `controllers/` up to project root, then down to `config/crd/bases/`.

After move: `internal/controller/suite_test.go` → need `filepath.Join("..", "..", "config", "crd", "bases")` because now two levels deep (`internal/controller/`).

### Reference

- Kubebuilder migration guide: https://book.kubebuilder.io/migration/reorganize-layout.html
- Manual migration: https://book-v3.book.kubebuilder.io/migration/manually_migration_guide_gov3_to_gov4

### Anti-Patterns to Avoid

- **DO NOT** rename `api/` directory — it stays at project root (go/v4 keeps `api/` in place)
- **DO NOT** change the `api/v1alpha1` package name — only `controllers` package changes
- **DO NOT** update the `OPERATOR_SDK_VERSION` in this story — that's Story 10.1
- **DO NOT** touch kustomize manifests syntax — that's Story 10.0a
- **DO NOT** reorganize the sub-packages (`vaultresourcecontroller`, `controllertestutils`, `vaultsecretutils`) — they move together under `internal/controller/` maintaining their structure
- **DO NOT** create an `internal/webhook/` directory — webhooks in this project are co-located in `api/v1alpha1/` (not in controllers) and don't move

### Project Structure Notes

Before:
```
vault-config-operator/
├── main.go
├── api/v1alpha1/
├── controllers/
│   ├── *.go (75 files)
│   ├── controllertestutils/
│   ├── vaultresourcecontroller/
│   └── vaultsecretutils/
├── config/
├── Dockerfile
└── Makefile
```

After:
```
vault-config-operator/
├── cmd/
│   └── main.go
├── api/v1alpha1/
├── internal/
│   └── controller/
│       ├── *.go (75 files)
│       ├── controllertestutils/
│       ├── vaultresourcecontroller/
│       └── vaultsecretutils/
├── config/
├── Dockerfile
└── Makefile
```

### Verification Commands (Run After Completion)

```bash
go build ./cmd/
make manifests generate
make fmt vet
make test
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.0]
- [Source: Kubebuilder v3→v4 migration guide](https://book.kubebuilder.io/migration/reorganize-layout.html)
- [Source: _bmad-output/project-context.md#Technology Stack & Versions]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
