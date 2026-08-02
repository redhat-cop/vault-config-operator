# Story 10.3: Upgrade golangci-lint from v1.64 to v2.x

Status: ready-for-dev

## Story

As an operator developer,
I want to upgrade golangci-lint from v1.64.8 to v2.12.2,
So that we benefit from new linters, performance improvements, improved configuration structure, and remain on the supported major version (v1 is EOL).

## Acceptance Criteria

1. **Given** `GOLANGCI_LINT_VERSION` is `v1.64.8` in the Makefile
   **When** it is updated to `v2.12.2`
   **Then** `make golangci-lint` downloads the new binary and `$(LOCALBIN)/golangci-lint --version` reports `v2.12.2`

2. **Given** golangci-lint v2 uses Go module path `github.com/golangci/golangci-lint/v2/cmd/golangci-lint`
   **When** the Makefile `go-install-tool` call is updated to the v2 module path
   **Then** `go install` resolves the correct package and the binary is functional

3. **Given** golangci-lint v2 requires a config file with `version: "2"` or no config file
   **When** a minimal `.golangci.yml` is created with v2 format
   **Then** `golangci-lint run ./...` uses the committed configuration

4. **Given** the project currently passes lint with v1.64.8 (zero findings from R1.2c verification)
   **When** `golangci-lint run ./...` is run with v2.12.2
   **Then** either exit code 0, or new findings are triaged and documented (not necessarily all fixed in this story)

5. **Given** a `lint` make target does not currently exist
   **When** one is added that invokes `$(GOLANGCI_LINT) run ./...`
   **Then** `make lint` runs the linter and can be integrated into CI workflows later

6. **Given** all changes are applied
   **When** `make fmt vet test` is run
   **Then** all pass without errors (lint upgrade has no effect on compilation or tests)

## Tasks / Subtasks

- [ ] Task 1: Update `GOLANGCI_LINT_VERSION` and module path in Makefile (AC: #1, #2)
  - [ ] Change `GOLANGCI_LINT_VERSION ?= v1.64.8` → `GOLANGCI_LINT_VERSION ?= v2.12.2`
  - [ ] Change install path from `github.com/golangci/golangci-lint/cmd/golangci-lint` → `github.com/golangci/golangci-lint/v2/cmd/golangci-lint`
  - [ ] Delete stale binary: `rm -f bin/golangci-lint bin/golangci-lint-*`
  - [ ] Run `make golangci-lint` and verify `bin/golangci-lint version` reports `v2.12.2`

- [ ] Task 2: Create `.golangci.yml` with v2 format (AC: #3)
  - [ ] Create `.golangci.yml` at project root with `version: "2"` and minimal sensible configuration
  - [ ] Enable default linters plus any that were implicitly used in v1 default set
  - [ ] Verify `golangci-lint run ./...` uses the config (check `golangci-lint config` output)

- [ ] Task 3: Add `lint` make target (AC: #5)
  - [ ] Add `.PHONY: lint` and `lint: golangci-lint` target that runs `$(GOLANGCI_LINT) run ./...`
  - [ ] Verify `make lint` works

- [ ] Task 4: Run lint and triage findings (AC: #4)
  - [ ] Run `golangci-lint run ./...` — capture output
  - [ ] If zero findings → done
  - [ ] If new findings exist, categorize:
    - **Must-fix**: Issues that indicate real bugs or correctness problems → fix in this story
    - **Should-fix**: Style/quality issues that are straightforward → fix if trivial
    - **Defer**: Complex issues or false positives → add `//nolint` with justification or adjust `.golangci.yml` exclusions
  - [ ] Document any deferred findings in Completion Notes

- [ ] Task 5: Verify build and tests (AC: #6)
  - [ ] Run `make fmt vet`
  - [ ] Run `make test`
  - [ ] Run `go build ./...` (or `go build ./cmd/` if Story 10.0 is done)

## Dev Notes

### Scope and Dependencies

This story upgrades the golangci-lint tool version and creates a committed lint configuration. It does NOT cover:

| Change | Story |
|--------|-------|
| go/v3 → go/v4 layout migration | Story 10.0 |
| Kustomize v3 → v5 syntax migration | Story 10.0a |
| Operator SDK v1.31 → v1.42 | Story 10.1 |
| Helm v3 → v4 | Story 10.2 |
| OPM + kustomize tool versions | Story 10.4 |
| kube-rbac-proxy removal | Story 10.5 |

**Prerequisites:** None strictly required. golangci-lint is functionally independent of the other Epic 10 stories. However, if Story 10.0 (go/v4 layout) is completed first, the lint target paths will reference `./cmd/` and `./internal/controller/` instead of `./` — the linter handles both cases since it lints `./...` recursively.

**Ordering note:** If Story 10.0 is done first, source files will be under `internal/controller/` instead of `controllers/`. This doesn't affect the lint upgrade itself — `./...` covers all packages regardless of directory structure. If Story 10.0 is NOT yet done, paths remain unchanged.

### Version Correction

The epics file references upgrading from "v1.59.1". The actual current version in the Makefile is **v1.64.8** — it was upgraded during Story R1.2c (lint green gate verification, June 2026). This story upgrades from v1.64.8 to v2.12.2.

### golangci-lint v2 Key Changes

| v2 Change | Project Impact | Action |
|-----------|---------------|--------|
| Go module path includes `/v2/` | Makefile install line breaks | Update module path |
| Config requires `version: "2"` | No config exists — could run without, but best to create one | Create `.golangci.yml` |
| v1 config files rejected outright | No impact (no existing config) | N/A |
| Default linters changed | May gain/lose enabled linters vs v1 defaults | Review and configure |
| No timeout by default (was 1min in v1) | Long lint runs won't be killed | Good for this project's 75+ controller files |
| `issues.show-stats` enabled by default | More verbose output | No action needed |
| `issues.exclude-generated` default changed to `strict` | Generated files (`zz_generated.deepcopy.go`) excluded more aggressively | Good — fewer false positives |
| Some linters renamed | `exportloopref` → removed (fixed in Go 1.22+), `gomnd` → `mnd`, etc. | N/A (no config references them) |
| `golangci-lint migrate` command | Not needed — no v1 config to migrate | Skip |

### golangci-lint v2 Default Linters

As of v2.12.2, the default enabled linters are:
- `copyloopvar` (replaced `exportloopref`, Go 1.22+ loop var semantics)
- `errcheck`
- `gosimple`
- `govet`
- `ineffassign`
- `staticcheck`
- `typecheck`
- `unused`

These are the same core linters that were default in v1 (minus `exportloopref` which is no longer needed). The project already passes all of these from the R1.2c green gate.

### No Existing `.golangci.yml` — What This Means

The project has **never** had a committed golangci-lint config file. In v1, this meant running with all defaults. In v2, this still works — golangci-lint v2 runs with defaults if no config is found. However, creating a minimal config is recommended because:
1. It documents the project's lint expectations
2. It allows exclusion patterns for known acceptable patterns
3. It enables adding stricter linters incrementally
4. It prevents surprise behavior from future golangci-lint updates

### Recommended Minimal `.golangci.yml`

```yaml
version: "2"

linters:
  default: standard
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - gosimple
    - ineffassign

issues:
  exclude-generated: strict

formatters:
  enable:
    - gofmt
```

The `default: standard` setting enables the standard set of linters. Additional linters can be enabled incrementally in future stories.

### Makefile Changes (Complete)

| Location | Current | New |
|----------|---------|-----|
| Line 27 | `GOLANGCI_LINT_VERSION ?= v1.64.8` | `GOLANGCI_LINT_VERSION ?= v2.12.2` |
| Line 305 | `$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))` | `$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))` |
| New target | (doesn't exist) | `.PHONY: lint`<br>`lint: golangci-lint`<br>`\t$(GOLANGCI_LINT) run ./...` |

### Files to Change (Complete List)

| File | Change |
|------|--------|
| `Makefile` | Version bump v1.64.8→v2.12.2, module path `/v2/`, add `lint` target |
| `.golangci.yml` (NEW) | Create with v2 format — minimal config |

### Files NOT to Change

- `go.mod` / `go.sum` — golangci-lint is not a Go dependency, it's installed as a standalone binary
- `.github/workflows/*.yaml` — golangci-lint is not currently in CI (project context confirms this); wiring it into CI is out of scope
- Any Go source files — unless lint findings in Task 4 require fixes
- `_bmad-output/project-context.md` — will be updated post-implementation by dev agent

### New Findings Triage Strategy

When golangci-lint v2 finds new issues (likely due to updated linter rules), apply this decision tree:

1. **Real bugs** (nil dereference, race conditions, resource leaks) → Fix immediately
2. **Correctness improvements** (unused variables, dead code, inefficient patterns) → Fix if < 5 minutes each
3. **Style issues** (naming, comment format, import ordering) → Fix only if project-wide consistency improves
4. **False positives on intentional patterns** → Add `//nolint:<linter> // <reason>` or exclude in `.golangci.yml`
5. **Findings in generated files** → Ensure `exclude-generated: strict` handles them; if not, add path exclusion

### Potential New Findings to Expect

Based on v2 linter changes and this codebase's patterns:
- `errcheck` — may find new unchecked returns in test files (these are acceptable in test setup code)
- `staticcheck` — may flag new deprecated patterns from newer Go/K8s library versions
- `unused` — may flag fields used only via reflection (CRD types use JSON tags extensively)
- `govet` — may have new checks for struct alignment or copylocks

### Anti-Patterns to Avoid

- **DO NOT** add golangci-lint to CI in this story — that's a separate decision requiring team consensus
- **DO NOT** suppress all findings with blanket `//nolint` annotations — each needs justification
- **DO NOT** enable many additional linters in this story — start with defaults, add incrementally
- **DO NOT** fix lint findings in test files that are merely stylistic — integration tests have different quality tradeoffs
- **DO NOT** modify `go.mod` — golangci-lint is a binary tool, not a library dependency
- **DO NOT** use the `curl | sh` installer — the project uses `go install` pattern consistently for all tools

### Verification Commands (Run After Completion)

```bash
rm -f bin/golangci-lint bin/golangci-lint-*
make golangci-lint
bin/golangci-lint version
make lint
make fmt vet test
```

### Rollback Strategy

If golangci-lint v2 causes unexpected issues:
1. Revert `GOLANGCI_LINT_VERSION` to `v1.64.8`
2. Revert module path to remove `/v2/`
3. Delete `.golangci.yml` (v1 works without it)
4. Remove `lint` target if it was added

### Project Structure Notes

- Alignment with unified project structure: This story adds `.golangci.yml` at project root (standard location per golangci-lint convention)
- Detected conflicts or variances: The project has never had committed lint config — this story establishes the baseline

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.3]
- [Source: _bmad-output/project-context.md#Build & Dev Tooling]
- [Source: _bmad-output/project-context.md#Code Quality Gates]
- [Source: _bmad-output/implementation-artifacts/R1-2c-lint-green-gate-verify-full-compliance.md]
- [golangci-lint v2 migration guide](https://golangci-lint.run/docs/product/migration-guide/)
- [golangci-lint v2.12.2 release](https://github.com/golangci/golangci-lint/releases/tag/v2.12.2)
- [golangci-lint v2 announcement](https://ldez.github.io/blog/2025/03/23/golangci-lint-v2/)
- [golangci-lint local installation](https://golangci-lint.run/docs/welcome/install/local/)

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
