# Story R1.2c: Lint Green Gate — Verify Full Lint Compliance

Status: ready-for-dev

## Story

As an operator developer,
I want confirmation that ALL lint findings are resolved after Stories R1.1, R1.2a, R1.2b, and R1.3 are complete,
So that we have a verified clean baseline before structural refactoring stories (R1.4–R1.6).

## Acceptance Criteria

1. **Given** Stories R1.1, R1.2a, R1.2b, and R1.3 are all complete **When** `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...` is run **Then** exit code 0, zero findings
2. **Given** `go vet ./...` **When** run **Then** exit code 0
3. **Given** `gofmt -l ./...` **When** run on non-generated files **Then** no output (all formatted)
4. **Given** the full lint and vet pass **When** `make test` and `make integration` are run **Then** all tests pass

## Prerequisites — Hard Dependencies

This story is a verification gate. It MUST NOT be started until ALL four prerequisite stories are complete and merged:

| Story | What it resolves | Finding count |
|-------|-----------------|---------------|
| R1.1 | SA1029 string context keys, PKI bug, webhook loggers, unsafe ToString | 13 |
| R1.2a | errcheck unchecked error returns (`ConfigureTLS`, `json.Encode`, `AddToScheme`) | 6 |
| R1.2b | SA1019 deprecated `rand.Seed` | 1 |
| R1.3 | SA1019 deprecated `ioutil`, dependency modernization, file rename | 1 |
| **Total** | | **21** |

If any prerequisite is not yet merged, **STOP** and report which stories are still pending.

## Tasks / Subtasks

- [ ] Task 0: Verify prerequisites (Gate check)
  - [ ] 0.1: Confirm R1.1 changes are merged (typed context keys in `api/v1alpha1/utils/`, PKI fix in `vautlpkiengineobject.go`, webhook logger fixes, safe `ToString`)
  - [ ] 0.2: Confirm R1.2a changes are merged (`ConfigureTLS` error handling in `commons.go`, `_ =` prefixes on `json.Encode` in test helpers, `AddToScheme` panic in `decoder.go`)
  - [ ] 0.3: Confirm R1.2b changes are merged (deleted `rand.Seed` init function and `"time"` import in `randomsecret_types.go`)
  - [ ] 0.4: Confirm R1.3 changes are merged (`ioutil.ReadFile` → `os.ReadFile` in `decoder.go`, `pkg/errors` removal, `vautlpkiengineobject.go` → `vaultpkiengineobject.go` rename, etc.)
- [ ] Task 1: Install correct golangci-lint version (AC: 1)
  - [ ] 1.1: The epic's lint baseline was captured with **golangci-lint v1.64.8**. The Makefile `GOLANGCI_LINT_VERSION` is `v1.59.1` (older). Install v1.64.8: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8` or override: `make golangci-lint GOLANGCI_LINT_VERSION=v1.64.8`
  - [ ] 1.2: Verify version: `golangci-lint --version` must show v1.64.8
- [ ] Task 2: Run full lint suite and confirm zero findings (AC: 1)
  - [ ] 2.1: Run `golangci-lint run --max-issues-per-linter=100 --max-same-issues=100 ./...`
  - [ ] 2.2: Confirm exit code 0 and zero findings
  - [ ] 2.3: If any findings remain, identify which prerequisite story missed them and document in Completion Notes — do NOT fix them in this story (this is a verification gate, not a fix story)
- [ ] Task 3: Run `go vet` and `gofmt` (AC: 2, 3)
  - [ ] 3.1: Run `go vet ./...` — confirm exit code 0
  - [ ] 3.2: Run `gofmt -l $(find . -name '*.go' ! -name 'zz_generated*' ! -path './vendor/*')` — confirm no output
- [ ] Task 4: Run full test suite (AC: 4)
  - [ ] 4.1: Run `make manifests generate fmt vet test` — confirm all pass with zero diffs from `manifests generate`
  - [ ] 4.2: Run `make integration` — confirm all pass (~576-579s expected runtime)
- [ ] Task 5: Document lint gate for future stories
  - [ ] 5.1: Record the exact lint command and version in the Completion Notes as the ongoing quality gate command for all future R1 stories

## Dev Notes

### This Is a Verification-Only Story

No new source code changes are expected. If all four prerequisite stories are correctly implemented, every task in this story should pass without modifications. The value of this story is the **verified confirmation** that the lint baseline is clean.

If lint findings DO remain, the correct response is:
1. Document which findings remain and which prerequisite story should have resolved them
2. Mark this story as blocked
3. Create a follow-up fix in the responsible prerequisite story
4. Do NOT fix lint issues in this story — it is strictly a gate

### golangci-lint Version Mismatch

The Makefile declares `GOLANGCI_LINT_VERSION ?= v1.59.1` but the epic's lint baseline was captured with **v1.64.8**. The baseline findings (21 total) were identified using v1.64.8. Using v1.59.1 might miss some checks or report different findings. **You must use v1.64.8 for this gate.**

There is no committed `.golangci.yml` configuration file in the repository. The lint runs with default settings, which means the default set of linters is enabled (including `staticcheck`, `errcheck`, `govet`, etc.).

### Expected Lint Finding Resolution Map

The following 21 findings were in the original baseline. After R1.1 + R1.2a + R1.2b + R1.3, all should be resolved:

| # | Linter | Finding | File | Resolved by |
|---|--------|---------|------|-------------|
| 1-13 | staticcheck SA1029 | `context.WithValue` / `.Value()` with string key | 7+ files in `controllers/` and `api/v1alpha1/` | R1.1 |
| 14 | errcheck | `config.ConfigureTLS(&tlsConfig)` unchecked | `api/v1alpha1/utils/commons.go` | R1.2a |
| 15 | errcheck | `json.NewEncoder(w).Encode(resp)` unchecked | `api/v1alpha1/utils/vaultobject_test.go` | R1.2a |
| 16-17 | errcheck | `json.NewEncoder(w).Encode(resp)` unchecked (2x) | `api/v1alpha1/prepareinternalvalues_test_helpers_test.go` | R1.2a |
| 18 | errcheck | `AddToScheme(scheme)` unchecked | `controllers/controllertestutils/decoder.go` | R1.2a |
| 19 | errcheck | `json.NewEncoder(w).Encode(resp)` unchecked | possible additional site | R1.2a |
| 20 | staticcheck SA1019 | `rand.Seed(time.Now().UnixNano())` deprecated | `api/v1alpha1/randomsecret_types.go` | R1.2b |
| 21 | staticcheck SA1019 | `ioutil.ReadFile` deprecated | `controllers/controllertestutils/decoder.go` | R1.3 |

### What NOT to Do

- Do NOT fix any lint findings — this is a gate, not a fix story
- Do NOT update the Makefile's `GOLANGCI_LINT_VERSION` — that may be a separate story or part of Epic 10 (golangci-lint upgrade from v1.59 to v2)
- Do NOT create a `.golangci.yml` config file — the project intentionally runs with defaults
- Do NOT modify any source files

### Kind Cluster Considerations

- `make integration` requires a live Kind cluster with Vault
- Runtime: ~576-579s (from R1.1 and R1.2a experience)
- Kind cluster can degrade (terminating namespaces) — if integration tests fail unexpectedly, try a fresh cluster with `make integration` (which recreates the cluster)

### Project Structure Notes

- No files created or modified in this story
- This is a pure verification gate
- The `make manifests generate` step should produce zero diffs (no type changes in this story)

### Previous Story Intelligence

**From R1.2b (immediately preceding):**
- `make integration` takes ~576-579s — budget accordingly
- Kind cluster can degrade — fresh cluster if tests fail unexpectedly
- Run `make manifests generate` even for non-type changes — catches unexpected diffs

**From R1.2a:**
- errcheck violations were resolved across 4 files (5 changes)
- The `ConfigureTLS` fix in `commons.go` was the only production code change
- Test helper changes (`_ =` prefix, `panic(err)` in init) are zero-risk

**From R1.1:**
- Largest story in the sequence — 57 context key migration sites across 18+ files
- PKI `CreateOrUpdateConfig` bug fixed (write path parameter)
- 3 webhook logger copy-paste errors fixed
- `ToString` made safe against non-string panics

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.2c] — acceptance criteria, lint baseline table, task list
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic preamble, lint baseline (21 findings), story ordering (R1.1 → R1.2a → R1.2b → R1.3 → R1.2c)
- [Source: _bmad-output/project-context.md#Code Quality Gates] — golangci-lint availability, no committed config
- [Source: _bmad-output/implementation-artifacts/R1-1-correctness-fixes-context-keys-pki-bug-webhook-loggers-tostring.md] — R1.1 story context
- [Source: _bmad-output/implementation-artifacts/R1-2a-fix-unchecked-error-returns-errcheck.md] — R1.2a story context
- [Source: _bmad-output/implementation-artifacts/R1-2b-remove-deprecated-rand-seed.md] — R1.2b story context
- [Source: Makefile:23] — `GOLANGCI_LINT_VERSION ?= v1.59.1` (older than epic baseline v1.64.8)
- [Source: Makefile:297-300] — `make golangci-lint` target for installation

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### Change Log

### File List
