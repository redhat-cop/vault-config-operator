# Story 9.1: Upgrade hashicorp/vault/api from v1.14.0 to v1.23.0

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want to upgrade the Vault API client library from v1.14.0 to v1.23.0,
So that the operator benefits from upstream security patches, dependency modernization (go-jose v4, backoff v4), and improved retry behavior (Retry-After header support).

## Acceptance Criteria

1. **Given** `go.mod` pins `github.com/hashicorp/vault/api v1.14.0`, **When** it is updated to `v1.23.0` and `go mod tidy` is run, **Then** `go build ./...` succeeds with zero compilation errors.

2. **Given** vault/api v1.23.0 has updated transitive dependencies (`cenkalti/backoff/v3` → `v4`, `go-retryablehttp` v0.7.6 → v0.7.8, `go-secure-stdlib/parseutil` v0.1.6 → v0.2.0, `natefinch/atomic` v1.0.1 added), **When** `go mod tidy` resolves the dependency graph, **Then** all direct and transitive dependencies are compatible and no conflicting version requirements exist.

3. **Given** the project uses `*vault.ResponseError` type assertions in error handling (14 files), **When** the upgrade is applied, **Then** all existing error handling patterns compile and work correctly (vault/api v1.23.0 does not change the `ResponseError` type signature).

4. **Given** the upgrade is applied, **When** `make test` is run, **Then** all unit tests pass.

5. **Given** the upgrade is applied, **When** `make integration` is run, **Then** all integration tests pass against the current Vault test infrastructure (Vault 1.19.0).

## Tasks / Subtasks

- [ ] Task 1: Upgrade vault/api dependency (AC: #1, #2)
  - [ ] 1.1 Run `go get github.com/hashicorp/vault/api@v1.23.0`
  - [ ] 1.2 Run `go mod tidy` to resolve transitive dependency changes
  - [ ] 1.3 Verify `go build ./...` succeeds
  - [ ] 1.4 Review `go.mod` diff — confirm expected dependency version changes (see Transitive Dependency Changes section below)
- [ ] Task 2: Verify no deprecated API usage requires migration (AC: #3)
  - [ ] 2.1 Confirm all 14 files importing `vault/api` compile without errors or warnings
  - [ ] 2.2 Verify `*vault.ResponseError` type assertions still work (type is unchanged in v1.23.0)
  - [ ] 2.3 Verify `vault.DefaultConfig()`, `vault.NewClient()`, `vault.TLSConfig`, `vault.Secret`, `vault.EnableAuditOptions`, `vault.LifetimeWatcherInput` — all used by this project — remain unchanged
- [ ] Task 3: Run tests (AC: #4, #5)
  - [ ] 3.1 Run `make test` — verify all unit tests pass
  - [ ] 3.2 Run `make integration` — verify all integration tests pass
- [ ] Task 4: Update project-context.md (AC: #1)
  - [ ] 4.1 Update `Vault Client: hashicorp/vault/api v1.14.0` → `v1.23.0` in the Technology Stack section

## Dev Notes

### Upgrade Risk Assessment: LOW

This is a **low-risk, low-friction** dependency upgrade. The vault/api module follows independent versioning from the Vault server (v1.14.0 → v1.23.0 spans ~9 API module releases, not Vault server versions). The API surface used by this project is stable and unchanged between these versions.

**No code changes expected** — this is a `go.mod`/`go.sum` only change (plus `project-context.md` version update).

### vault/api Module Versioning

The `github.com/hashicorp/vault/api` module version numbers are **independent** from Vault server version numbers. v1.23.0 is the latest tagged API release (March 23, 2026). There is no vault/api v2.0.0 — the Vault 2.0 server release did not create a v2 API module. The CVE referenced in Snyk (SNYK-GOLANG-GITHUBCOMHASHICORPVAULTAPI-16109648) that suggests upgrading to v2.0.0 is an incorrect advisory.

A CVE fix (CVE-2026-34986 for go-jose) was committed April 7, 2026 but not yet included in any tagged release. Monitor hashicorp/vault#31938 for a v1.23.1 or v1.24.0 release. The project's existing `go-jose/go-jose/v4 v4.1.3` (pulled in via K8s dependencies) is already at a version that includes the fix.

### API Surface Used by This Project (All Unchanged in v1.23.0)

| API Element | Usage Location(s) |
|---|---|
| `vault.DefaultConfig()` | `utils/commons.go`, test files |
| `vault.NewClient(config)` | `utils/commons.go`, test files |
| `vault.Config` (Address, Timeout, MaxRetries, ConfigureTLS) | `utils/commons.go` |
| `vault.TLSConfig` (CACertBytes, ClientKey, ClientCert, CACert, TLSServerName, Insecure) | `utils/commons.go` |
| `*vault.Client` methods: `Logical().Read/Write/Delete`, `Sys().ListAudit/EnableAuditWithOptions/DisableAudit`, `Auth().Token().LookupSelf()`, `SetToken`, `SetNamespace`, `WithNamespace`, `NewLifetimeWatcher` | 14 files across `api/v1alpha1/utils/`, `api/v1alpha1/`, `controllers/` |
| `*vault.Secret` (.Data, .Auth) | Response handling across utils/ |
| `*vault.ResponseError` (.StatusCode) | Error handling in `vaultutils.go`, `vaultobject.go`, `vaultpkiengineobject.go`, `vaultauditobject.go`, `audit_controller_test.go` |
| `vault.EnableAuditOptions` | `vaultauditobject.go` |
| `vault.LifetimeWatcherInput` | `commons.go` |

### Behavioral Changes (Transparent, No Code Required)

1. **Retry-After header support (v1.20.0):** The default API client now checks for the `Retry-After` header on responses and waits the specified duration before retrying. This is transparent — no code changes needed, behavior improves automatically.

2. **go-jose v3 → v4 transition (v1.17.0):** vault/api v1.23.0 depends on `go-jose/go-jose/v4`. This project's `go.mod` already has `go-jose/go-jose/v4 v4.1.3` (via K8s dependencies), so no new module path is introduced.

### Transitive Dependency Changes

Expected `go.mod` changes after running `go get` + `go mod tidy`:

| Dependency | Before (v1.14.0) | After (v1.23.0) | Notes |
|---|---|---|---|
| `cenkalti/backoff/v3` | v3.0.0 | **REMOVED** | v3 replaced by v4 in vault/api v1.23.0 |
| `cenkalti/backoff/v4` | (not present) | **v4.3.0 ADDED** | New module path, coexists if anything else needs v3 |
| `go-retryablehttp` | v0.7.6 | **v0.7.8** | Minor bump |
| `go-secure-stdlib/parseutil` | v0.1.6 | **v0.2.0** | Minor API surface change — this project does not import it directly |
| `go-sockaddr` | v1.0.2 | **v1.0.7** | Patch bump |
| `hashicorp/hcl` (v1) | v1.0.0 | **v1.0.1-vault-7** | Vault-specific fork tag |
| `natefinch/atomic` | (not present) | **v1.0.1 ADDED** | New indirect dependency |
| `golang.org/x/net` | v0.49.0 | v0.49.0 (unchanged) | Project is already above vault/api's v0.47.0 minimum |
| `golang.org/x/time` | v0.14.0 | v0.14.0 (unchanged) | Project is already above vault/api's v0.12.0 minimum |
| `go-hclog` | (not in direct) | **v1.6.3** (new indirect) | vault/api v1.23.0 dependency |
| `fatih/color` | (not present) | **v0.18.0** (new indirect) | Transitive via go-hclog |

**Note:** `go mod tidy` may also adjust other transitive versions. The key point is that no **direct** dependency in go.mod changes except `vault/api` itself.

**Important:** `cenkalti/backoff/v3` removal depends on whether any other module in the dependency tree still requires it. If another dependency (e.g., controller-runtime's transitive tree) still needs v3, both v3 and v4 will coexist in go.mod. This is fine — Go modules support major version coexistence.

### Files to Modify

| File | Change |
|------|--------|
| `go.mod` | Update `github.com/hashicorp/vault/api` from `v1.14.0` to `v1.23.0`; transitive deps updated via `go mod tidy` |
| `go.sum` | Automatically regenerated by `go mod tidy` |
| `_bmad-output/project-context.md` | Update `Vault Client: hashicorp/vault/api v1.14.0` → `v1.23.0` in Technology Stack section |

**Files NOT modified:**
- No `.go` source files need changes — the vault/api v1.23.0 API surface is backward-compatible with v1.14.0 for all types and methods this project uses
- No CRD regeneration needed (`make manifests generate` not required)
- No webhook changes needed
- No controller changes needed
- No test changes needed

### Anti-Pattern Prevention

- **DO NOT** upgrade to a vault/api pre-release or pseudo-version (v1.23.1-0.2026...). Use only the tagged release `v1.23.0`.
- **DO NOT** attempt to upgrade to `vault/api v2.0.0` — it does not exist. The Snyk CVE advisory suggesting this is incorrect.
- **DO NOT** change any `.go` source code as part of this upgrade. If `go build` fails after the dependency update, investigate the specific compilation error before making source changes — failure would indicate an unexpected API break that needs story scope reassessment.
- **DO NOT** upgrade Vault server version in test infrastructure — that is Story 9.5's scope.
- **DO NOT** upgrade other direct dependencies (ginkgo, gomega, hcl/v2, sprig/v3) — those are Stories 9.2 and 9.3.
- **DO NOT** run `go get -u` (upgrade all) — only upgrade vault/api specifically with `go get github.com/hashicorp/vault/api@v1.23.0`.

### Scope Boundary

**IN scope:** vault/api v1.14.0 → v1.23.0, its transitive dependency resolution, project-context.md version update.

**OUT of scope:**
- Migrating `err.(*vault.ResponseError)` type assertions to `errors.As` pattern — this is a code modernization improvement, not required for the upgrade (both patterns work with v1.23.0). Could be a future R2 story.
- Migrating `Logical().Read/Write` to `Logical().ReadWithContext/WriteWithContext` — not required, both patterns work. The non-context variants are not deprecated.
- Adopting `vault/api/auth/kubernetes` helper package — this would simplify kube auth login but is a refactoring, not an upgrade requirement.
- Fixing the `namespace_controller.go` raw string context key (`ctx1.Value("vaultClient")`) — pre-existing issue, not related to this upgrade.

### Previous Story Intelligence

**Story 9.0 (Fix RabbitMQ serialization bug):** Pure bugfix in `api/v1alpha1/rabbitmqsecretenginerole_types.go` — serialization logic only. No dependency changes. No interaction with this story.

**Epic 8 (Go + K8s Stack Upgrade) completed 2026-07-19:**
- Go upgraded from 1.22 → 1.26, controller-runtime v0.17 → v0.24, K8s libs v0.29 → v0.36
- 44 webhook files migrated to new generic admission interfaces
- All tests pass on the updated stack — this is the baseline for Story 9.1
- Key learnings: just-in-time version audit prevents stale targets; sequential dependency chains need explicit scope boundary documentation

**Epic 8 Retrospective (2026-07-20) confirmed:**
- vault/api v1.23.0 is confirmed still current — no newer tagged release exists
- CVE-2026-34986 (go-jose) fix committed but not in any tagged API release; project's existing go-jose v4.1.3 already includes the fix
- 3 immediate remediation items (webhook markers, loggers, mount path docs) were identified — these are independent of Story 9.1 and should be done before or alongside this story

### Project Structure Notes

- Only `go.mod`, `go.sum`, and `_bmad-output/project-context.md` are modified
- No new files created
- No source code changes expected
- Follows the same low-friction upgrade pattern as Epic 8 stories

### References

- [Source: go.mod:13 — current vault/api v1.14.0 pin]
- [Source: _bmad-output/project-context.md:24 — Vault Client version in Technology Stack]
- [Source: _bmad-output/planning-artifacts/epics.md:1916-1932 — Story 9.1 epic definition]
- [Source: _bmad-output/implementation-artifacts/epic-8-retro-2026-07-20.md:107-124 — Epic 9 version target audit]
- [Source: _bmad-output/implementation-artifacts/9-0-fix-rabbitmq-vhost-vhosttopic-serialization-bug.md — Previous story]
- [Source: https://pkg.go.dev/github.com/hashicorp/vault/api@v1.23.0 — vault/api v1.23.0 module page with dependency list]
- [Source: https://github.com/hashicorp/vault/blob/main/CHANGELOG.md — Vault CHANGELOG for API changes]
- [Source: https://github.com/rook/rook/issues/17413 — Confirms vault/api v2.0.0 does not exist]

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
