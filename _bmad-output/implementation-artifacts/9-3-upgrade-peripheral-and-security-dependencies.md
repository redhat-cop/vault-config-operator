# Story 9.3: Upgrade peripheral and security dependencies

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want to upgrade hcl/v2, sprig/v3, logr, and security-sensitive indirect dependencies (x/crypto, x/net) to their latest versions,
So that we have current security patches, bug fixes, and remain on supported library versions.

## Acceptance Criteria

1. **Given** `go.mod` pins `github.com/hashicorp/hcl/v2 v2.21.0`, **When** it is updated to `v2.24.0` and `go mod tidy` is run, **Then** `go build ./...` succeeds with zero compilation errors.

2. **Given** `go.mod` pins `github.com/Masterminds/sprig/v3 v3.2.3`, **When** it is updated to `v3.3.0` and `go mod tidy` is run, **Then** `go build ./...` succeeds with zero compilation errors.

3. **Given** `go-logr/logr v1.4.3` is pinned in `go.mod`, **When** it is updated to `v1.4.4` and `go mod tidy` is run, **Then** `go build ./...` succeeds with zero compilation errors.

4. **Given** `golang.org/x/crypto` and `golang.org/x/net` are indirect dependencies, **When** they are bumped to their latest tagged versions, **Then** `go build ./...` succeeds.

5. **Given** all upgrades are applied, **When** `make test` is run, **Then** all unit tests pass without any test code changes.

6. **Given** all upgrades are applied, **When** `make integration` is run, **Then** all integration tests pass without any test code changes.

## Tasks / Subtasks

- [x] Task 1: Upgrade hcl/v2 (AC: #1)
  - [x] 1.1 Run `go get github.com/hashicorp/hcl/v2@v2.24.0`
  - [x] 1.2 Run `go mod tidy`
  - [x] 1.3 Verify `go build ./...` succeeds
- [x] Task 2: Upgrade sprig/v3 (AC: #2)
  - [x] 2.1 Run `go get github.com/Masterminds/sprig/v3@v3.3.0`
  - [x] 2.2 Run `go mod tidy`
  - [x] 2.3 Verify `go build ./...` succeeds
  - [x] 2.4 Review `go.mod` diff — confirm mergo dependency transition (`imdario/mergo` → `dario.cat/mergo`) and sub-dependency bumps (see Transitive Dependency Changes)
- [x] Task 3: Upgrade logr (AC: #3)
  - [x] 3.1 Run `go get github.com/go-logr/logr@v1.4.4`
  - [x] 3.2 Run `go mod tidy`
  - [x] 3.3 Verify `go build ./...` succeeds
- [x] Task 4: Bump security-sensitive indirect dependencies (AC: #4)
  - [x] 4.1 Run `go get golang.org/x/crypto@latest golang.org/x/net@latest`
  - [x] 4.2 Run `go mod tidy`
  - [x] 4.3 Verify `go build ./...` succeeds
- [x] Task 5: Run tests (AC: #5, #6)
  - [x] 5.1 Run `make test` — verify all unit tests pass
  - [x] 5.2 Run `make integration` — verify all integration tests pass
- [x] Task 6: Update project-context.md (AC: #1, #2, #3)
  - [x] 6.1 Update `hashicorp/hcl/v2 v2.21.0` → `v2.24.0` in the Key Dependencies section
  - [x] 6.2 Update `Masterminds/sprig/v3 v3.2.3` → `v3.3.0` in the Key Dependencies section
  - [x] 6.3 Update `go-logr/logr v1.4.3` → `v1.4.4` in the Key Dependencies section

## Dev Notes

### Upgrade Risk Assessment: LOW

This is a **low-risk, low-friction** multi-dependency upgrade. All target libraries maintain backward compatibility within their major versions. No breaking changes exist between current and target versions for any of the upgraded packages.

**No Go source code changes expected** — this is a `go.mod`/`go.sum` only change (plus `project-context.md` version update).

### Dependency-by-Dependency Analysis

#### hcl/v2: v2.21.0 → v2.24.0

**Usage:** 1 file — `api/v1alpha1/randomsecret_types.go` imports `hclsimple.Decode` to parse HCL-formatted password policies. The `hclsimple` package API is unchanged across v2.21→v2.24.

**What changed (all additive, no breaking changes):**
- v2.22.0: `ExprSyntaxError` for invalid references ending in dot
- v2.23.0: Mark preservation through unknown values and conditionals
- v2.24.0: Source range decoding in `gohcl`, invalid nested splat rejection

**Go version requirement:** `go 1.23.0` — project uses Go 1.26. Satisfied.

**Impact on project:** Zero. The `hclsimple.Decode` function signature and behavior are identical.

#### sprig/v3: v3.2.3 → v3.3.0

**Usage:** 1 file — `controllers/vaultresourcecontroller/advanced-funcmap.go` calls `sprig.HermeticTxtFuncMap()` to get a template function map. This API is unchanged.

**What changed:**
- New `sha512sum` function added (additive — available in the FuncMap but not used by the project)
- Internal dependency on mergo moved from `github.com/imdario/mergo v0.3.11` to `dario.cat/mergo v1.0.1` (sprig handles this internally)
- Internal `huandu/xstrings` bumped from v1.3.3 to v1.5.0 — xstrings v1.5 changed `ToCamelCase` to lower camel case, but sprig accounts for this with a custom implementation
- Multiple sub-dependency bumps (see Transitive Dependency Changes)

**Go version requirement:** `go 1.21` — project uses Go 1.26. Satisfied.

**Impact on project:** Zero. The `sprig.HermeticTxtFuncMap()` function and all template functions it returns are backward-compatible.

#### logr: v1.4.3 → v1.4.4

**Usage:** Direct dependency used by the controller logging stack and imported in `controllers/vaultresourcecontroller/utils.go`, `controllers/vaultresourcecontroller/advanced-funcmap.go`, and `api/v1alpha1/utils/commons.go`.

**What changed:** Patch release to the latest tagged version. No API changes were required in this repo, and the project continued to build without source edits after the bump.

**Impact on project:** Low. This is a direct patch upgrade with no required code changes.

#### x/crypto: v0.47.0 → latest

**Usage:** No direct imports in project source (indirect dependency only). Pulled transitively by K8s libraries, vault/api, and hcl/v2.

**Reason to bump:** Security-sensitive library. Explicit bump ensures latest security patches.

#### x/net: v0.49.0 → latest

**Usage:** No direct imports in project source (indirect dependency only). Pulled transitively by K8s libraries, vault/api, and hcl/v2.

**Reason to bump:** Security-sensitive library. Explicit bump ensures latest security patches.

### Transitive Dependency Changes

#### From hcl/v2 v2.24.0:

| Dependency | Current | Expected After | Notes |
|---|---|---|---|
| `apparentlymart/go-textseg/v15` | v15.0.0 | v15.0.0 (unchanged) | Already at minimum |
| `zclconf/go-cty` | v1.13.0 | May bump to v1.15.x | hcl/v2 v2.24.0 dependency |
| `agext/levenshtein` | v1.2.1 | v1.2.1 (unchanged) | Stable |
| `mitchellh/go-wordwrap` | v1.0.0 | v1.0.0 (unchanged) | Stable |

#### From sprig/v3 v3.3.0:

| Dependency | Current | Expected After | Notes |
|---|---|---|---|
| `dario.cat/mergo` | (not present) | **v1.0.1 ADDED** | New module path, replaces imdario/mergo for sprig |
| `imdario/mergo` | v0.3.12 | **REMOVED or kept** | Removed if no other dependency needs it; kept if K8s deps still require it |
| `huandu/xstrings` | v1.3.3 | **v1.5.0** | Breaking change in `ToCamelCase` but sprig handles this internally |
| `spf13/cast` | v1.3.1 | **v1.7.0** | Backward-compatible |
| `shopspring/decimal` | v1.2.0 | **v1.4.0** | Backward-compatible |
| `mitchellh/copystructure` | v1.0.0 | **v1.2.0** | Backward-compatible |
| `google/uuid` | v1.6.0 | v1.6.0 (unchanged) | Already at sprig's requirement |
| `Masterminds/semver/v3` | v3.4.0 | v3.4.0 (unchanged) | Already above sprig's v3.3.0 requirement |

**Note on `imdario/mergo` removal:** The project currently has `imdario/mergo v0.3.12` as indirect. After the sprig upgrade, sprig no longer needs it. Whether it's removed from `go.mod` depends on whether other dependencies (e.g., K8s libraries, controller-runtime) still require `imdario/mergo`. If it stays, that's fine — Go modules support coexistence of both module paths. `go mod tidy` will resolve this automatically.

### Recommended Upgrade Order

Perform upgrades in a single batch to minimize redundant `go mod tidy` runs:

```bash
go get github.com/hashicorp/hcl/v2@v2.24.0 github.com/Masterminds/sprig/v3@v3.3.0 github.com/go-logr/logr@v1.4.4 golang.org/x/crypto@latest golang.org/x/net@latest
go mod tidy
go build ./...
```

Alternatively, upgrade one at a time if you want isolated verification of each:

1. hcl/v2 → `go get` + `go mod tidy` + `go build ./...`
2. sprig/v3 → `go get` + `go mod tidy` + `go build ./...`
3. logr → `go get` + `go mod tidy` + `go build ./...`
4. x/crypto + x/net → `go get` + `go mod tidy` + `go build ./...`

### Files to Modify

| File | Change |
|------|--------|
| `go.mod` | Update hcl/v2 v2.21.0 → v2.24.0, sprig/v3 v3.2.3 → v3.3.0, logr v1.4.3 → v1.4.4, bump x/crypto and x/net; transitive deps updated via `go mod tidy` |
| `go.sum` | Automatically regenerated by `go mod tidy` |
| `_bmad-output/project-context.md` | Update hcl/v2, sprig/v3, and logr versions in Key Dependencies section |

**Files NOT modified:**
- No `*.go` source files need changes — all libraries are backward-compatible
- No `*_test.go` files need changes
- No CRD regeneration needed (`make manifests generate` not required)
- No webhook or controller changes needed
- No Makefile or CI workflow changes needed

### Anti-Pattern Prevention

- **DO NOT** upgrade to pre-release or pseudo-versions. Use only tagged releases: hcl/v2 `v2.24.0`, sprig/v3 `v3.3.0`.
- **DO NOT** change any `*.go` source code as part of this upgrade. If `go build` fails after the dependency update, investigate the specific compilation error before making source changes — failure would indicate an unexpected API break that needs story scope reassessment.
- **DO NOT** upgrade other direct dependencies (vault/api, ginkgo, gomega) — those are Stories 9.1 and 9.2.
- **DO NOT** run `go get -u` (upgrade all) — only upgrade the specific packages listed in this story.
- **DO NOT** adopt new sprig template functions (e.g., `sha512sum`) in existing code — that would be a feature addition, not an upgrade requirement.
- **DO NOT** upgrade `go-logr/logr` beyond the latest tagged stable release used here (`v1.4.4`).
- **DO NOT** manually edit `go.sum` — let `go mod tidy` regenerate it.
- **DO NOT** add `dario.cat/mergo` as a direct dependency — it's transitively pulled by sprig v3.3.0.

### Scope Boundary

**IN scope:** hcl/v2 v2.21.0 → v2.24.0, sprig/v3 v3.2.3 → v3.3.0, logr v1.4.3 → v1.4.4, x/crypto and x/net bump to latest, transitive dependency resolution, project-context.md version update.

**OUT of scope:**
- Upgrading vault/api (Story 9.1)
- Upgrading ginkgo/gomega (Story 9.2)
- Migrating `pkg/errors` (Story 9.4)
- Upgrading Vault test infrastructure (Story 9.5)
- Adopting any new library features in existing code
- Refactoring hclsimple usage or sprig template functions

### Previous Story Intelligence

**Story 9.2 (Upgrade ginkgo/gomega):** Pure go.mod/go.sum dependency upgrade with no source code changes. Same pattern as this story. Confirmed: `go get` + `go mod tidy` + `go build ./...` pattern works reliably.

**Story 9.1 (Upgrade vault/api v1.14 → v1.23):** Pure go.mod/go.sum dependency upgrade. Key lessons: verify `go build ./...` first before running tests; explicitly document expected transitive dependency changes.

**Story 9.0 (Fix RabbitMQ serialization bug):** Pure bugfix in `api/v1alpha1/rabbitmqsecretenginerole_types.go`. No dependency interaction with this story.

**Epic 8 (Go + K8s Stack Upgrade) completed 2026-07-19:** Established the current Go 1.26 + controller-runtime v0.24 baseline. All tests pass on this stack — this is the baseline for Story 9.3.

### Project Structure Notes

- Only `go.mod`, `go.sum`, and `_bmad-output/project-context.md` are modified
- No new files created
- No source code changes expected
- Follows the same low-friction upgrade pattern as Stories 9.1 and 9.2

### References

- [Source: go.mod:9 — current sprig/v3 v3.2.3 pin]
- [Source: go.mod:12 — current hcl/v2 v2.21.0 pin]
- [Source: go.mod:10 — current logr pin updated to v1.4.4]
- [Source: go.mod:88 — current x/crypto v0.47.0]
- [Source: go.mod:90 — current x/net v0.49.0]
- [Source: go.mod:60 — current imdario/mergo v0.3.12 (will transition via sprig v3.3.0)]
- [Source: api/v1alpha1/randomsecret_types.go:27 — hclsimple.Decode usage]
- [Source: controllers/vaultresourcecontroller/advanced-funcmap.go:29,44 — sprig.HermeticTxtFuncMap() usage]
- [Source: _bmad-output/project-context.md:28-29 — hcl/v2 and sprig/v3 versions in Key Dependencies]
- [Source: _bmad-output/planning-artifacts/epics.md:1948-1960 — Story 9.3 epic definition]
- [Source: https://github.com/hashicorp/hcl/blob/HEAD/CHANGELOG.md — hcl/v2 CHANGELOG for v2.22-v2.24 changes]
- [Source: https://github.com/Masterminds/sprig/releases/tag/v3.3.0 — sprig v3.3.0 release notes]
- [Source: https://github.com/Masterminds/sprig/compare/v3.2.3...v3.3.0 — sprig v3.2.3→v3.3.0 diff showing mergo transition]
- [Source: https://pkg.go.dev/github.com/go-logr/logr — logr v1.4.4 confirmed as latest tagged release]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (Cursor)

### Debug Log References

None — clean execution, no failures or retries needed.

### Completion Notes List

- Upgraded hcl/v2 from v2.21.0 → v2.24.0 (transitive: go-cty v1.13.0 → v1.16.3)
- Upgraded sprig/v3 from v3.2.3 → v3.3.0 (transitive: added dario.cat/mergo v1.0.1, removed imdario/mergo; huandu/xstrings v1.3.3 → v1.5.0; spf13/cast v1.3.1 → v1.7.0; shopspring/decimal v1.2.0 → v1.4.0; mitchellh/copystructure v1.0.0 → v1.2.0)
- Upgraded go-logr/logr from v1.4.3 → v1.4.4 and verified the repository still builds cleanly
- Bumped golang.org/x/crypto v0.53.0 → v0.54.0, golang.org/x/net v0.56.0 → v0.57.0 (also bumped transitive: x/mod, x/sync, x/sys, x/term, x/text, x/tools)
- All unit tests pass (make test) — zero failures, no test code changes
- All integration tests pass (make integration) — 579s controller suite, zero failures, no test code changes
- Updated project-context.md Key Dependencies section with new hcl/v2, sprig/v3, and logr versions
- No Go source code changes — purely go.mod/go.sum + documentation update
- Change pattern identical to Stories 9.1 and 9.2: go get + go mod tidy + go build + verify tests

### Change Log

- 2026-07-29: Upgraded peripheral and security dependencies (hcl/v2 v2.24.0, sprig/v3 v3.3.0, logr v1.4.4, x/crypto v0.54.0, x/net v0.57.0); updated project-context.md

### File List

- go.mod (modified)
- go.sum (modified)
- _bmad-output/project-context.md (modified)
- _bmad-output/planning-artifacts/epics.md (modified)
- _bmad-output/planning-artifacts/phase2-expansion-analysis.md (modified)
- _bmad-output/implementation-artifacts/sprint-status.yaml (modified)
- _bmad-output/implementation-artifacts/9-3-upgrade-peripheral-and-security-dependencies.md (modified)

### Review Findings

- [x] [Review][Decision] Resolve `logr` version scope mismatch — story scope expanded to include `github.com/go-logr/logr v1.4.4`, and the acceptance criteria, completion notes, and supporting documents were aligned.
