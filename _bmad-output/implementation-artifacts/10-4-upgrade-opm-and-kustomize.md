# Story 10.4: Upgrade OPM and Kustomize

Status: ready-for-dev

## Story

As an operator developer,
I want to upgrade OPM from v1.23.0 to v1.73.0 and Kustomize from v5.4.3 to v5.8.1,
So that OLM catalog building and kustomize rendering use current tooling.

## Acceptance Criteria

1. **Given** OPM version is hardcoded to `v1.23.0` in the Makefile `opm` target URL
   **When** an `OPM_VERSION` variable is introduced and set to `v1.73.0`
   **Then** `make opm` downloads the correct binary and `$(LOCALBIN)/opm version` reports `v1.73.0`

2. **Given** `make catalog-build` uses `$(OPM) index add` with `--container-tool`, `--mode semver`, `--tag`, and `--bundles` flags
   **When** OPM is upgraded to v1.73.0
   **Then** `make catalog-build` succeeds (the `opm index add` command is deprecated but still functional in v1.73.0)

3. **Given** `KUSTOMIZE_VERSION` is `v5.4.3`
   **When** it is updated to `v5.8.1`
   **Then** `make kustomize` downloads the new version and `$(LOCALBIN)/kustomize version` reports `v5.8.1`

4. **Given** kustomize v5.8.1 includes Helm v4 interop fixes and namespace propagation improvements
   **When** `make manifests` is run
   **Then** CRD and RBAC generation succeeds without errors

5. **Given** kustomize v5.8.1 is installed
   **When** `make deploy` is run (dry-run to verify kustomize build)
   **Then** `kustomize build config/default` produces valid output

6. **Given** kustomize v5.8.1 is installed
   **When** `make helmchart` is run
   **Then** chart generation and `helm lint` pass

7. **Given** all changes are applied
   **When** `make build` and `make test` are run
   **Then** both pass without errors (tool version changes have no effect on Go compilation or tests)

## Tasks / Subtasks

- [ ] Task 1: Introduce `OPM_VERSION` variable and update `opm` target (AC: #1)
  - [ ] Add `OPM_VERSION ?= v1.73.0` near other tool version variables (around line 21, after `OPERATOR_SDK_VERSION`)
  - [ ] Update the hardcoded `v1.23.0` in the `opm` target's `curl` URL to use `$(OPM_VERSION)`
  - [ ] Delete stale binary: `rm -f bin/opm`
  - [ ] Run `make opm` and verify `bin/opm version` reports v1.73.0

- [ ] Task 2: Update `KUSTOMIZE_VERSION` in Makefile (AC: #3)
  - [ ] Change `KUSTOMIZE_VERSION ?= v5.4.3` → `KUSTOMIZE_VERSION ?= v5.8.1`
  - [ ] Delete stale binary: `rm -f bin/kustomize bin/kustomize-*`
  - [ ] Run `make kustomize` and verify `bin/kustomize version` reports v5.8.1

- [ ] Task 3: Verify kustomize-dependent targets (AC: #4, #5, #6)
  - [ ] Run `make manifests` — confirms CRD/RBAC generation works with kustomize v5.8.1
  - [ ] Run `kustomize build config/default` — confirms kustomize overlay rendering
  - [ ] Run `make helmchart` — confirms chart generation and helm lint pass

- [ ] Task 4: Verify catalog-build (AC: #2)
  - [ ] Run `make catalog-build` (requires a pushed bundle image, so verify at minimum that `make opm` succeeds and `opm version` is correct)
  - [ ] If a bundle image is available, run `make catalog-build` and confirm success
  - [ ] Note: `opm index add` emits a deprecation warning — this is expected and acceptable

- [ ] Task 5: Verify build and tests (AC: #7)
  - [ ] Run `make fmt vet`
  - [ ] Run `make test`
  - [ ] Run `go build ./...` (or `go build ./cmd/` if Story 10.0 is done)

## Dev Notes

### Scope and Dependencies

This story upgrades two CLI tool versions: OPM and Kustomize. It does NOT cover:

| Change | Story |
|--------|-------|
| go/v3 → go/v4 layout migration | Story 10.0 |
| Kustomize v3 → v5 syntax migration | Story 10.0a |
| Operator SDK v1.31 → v1.42 | Story 10.1 |
| Helm v3 → v4 | Story 10.2 |
| golangci-lint v1 → v2 | Story 10.3 |
| kube-rbac-proxy removal | Story 10.5 |
| Helm chart metrics rearchitecture | Story 10.6 |
| Migrating from `opm index add` to File-Based Catalogs (FBC) | Future story (separate scope) |

**Prerequisites:** Stories 10.0a (kustomize v3→v5 syntax migration) and 10.1 (Operator SDK upgrade) should ideally be completed first. The kustomize syntax migration ensures that config manifests use `resources:`, `patches:`, and `replacements:` instead of the deprecated `bases:`, `patchesStrategicMerge:`, and `vars:` syntax. Kustomize v5.8.1 still supports the old syntax, so this story can proceed independently if needed, but the output may differ if old syntax is still in place.

**Ordering note:** If Story 10.0 (go/v4 layout) is done first, source files will be under `internal/controller/` instead of `controllers/` and `cmd/main.go` instead of `main.go`. This does not affect OPM or kustomize — these tools operate on `config/` manifests, not Go source files.

### Version Corrections from Epic

The epics file specifies OPM target as "v1.65.x" and Kustomize as "v5.8.1". The latest stable OPM release (as of August 2026) is **v1.73.0**, so this story targets that version instead of v1.65.x to minimize the upgrade gap.

### OPM Upgrade: v1.23.0 → v1.73.0

#### Current State

The OPM version is **hardcoded** directly in the download URL (Makefile line 366):

```makefile
curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.23.0/$${OS}-$${ARCH}-opm
```

Unlike other tools (kustomize, helm, controller-gen) which use version variables, OPM has no `OPM_VERSION` variable. This story introduces one for consistency.

#### Download URL Pattern

The URL pattern is unchanged between v1.23.0 and v1.73.0:

```
https://github.com/operator-framework/operator-registry/releases/download/v<VERSION>/<OS>-<ARCH>-opm
```

Asset naming (`linux-amd64-opm`, `darwin-arm64-opm`, etc.) is identical across versions. No Makefile target restructuring is needed beyond the version reference.

#### `opm index add` Deprecation

The current `catalog-build` target uses:

```makefile
$(OPM) index add --container-tool $(CONTAINER_RUNTIME) --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)
```

**Status in v1.73.0:** The `opm index add` command is **deprecated** but still functional. It emits a deprecation warning directing users to File-Based Catalogs (FBC). The command, flags (`--container-tool`, `--mode`, `--tag`, `--bundles`, `--from-index`), and behavior are unchanged.

**Future removal:** A pending PR (#1986) will remove `opm index`, `opm registry`, and all SQLite-based catalog support in a future release. Migrating to FBC format is a separate, larger scope change that should be tracked as its own story.

**For this story:** Accept the deprecation warning. Do NOT migrate to FBC — that is out of scope.

#### OPM Breaking Changes Impact

| OPM Change (v1.23 → v1.73) | Project Impact | Action |
|------------------------------|---------------|--------|
| `opm index add` deprecated | Warning printed during `make catalog-build` | Accept warning — FBC migration is separate |
| Go version requirement bumped | No impact — OPM is a standalone binary | None |
| New `opm render` command available | Not used in current Makefile | None (future FBC migration) |
| `opm alpha generate dockerfile` available | Not used | None |
| Binary size increased (~69MB vs ~40MB) | Larger download, no functional impact | None |

### Kustomize Upgrade: v5.4.3 → v5.8.1

#### Module Path

The Go module path stays the same: `sigs.k8s.io/kustomize/kustomize/v5`. No Makefile `go-install-tool` path change is needed.

#### Kustomize v5.8.1 Key Changes

| Change | Project Impact | Action |
|--------|---------------|--------|
| Helm v4 interop fixes (#6016) | Critical if Story 10.2 (Helm v4) is done — ensures `kustomize build` works with Helm v4 | Positive — enables Helm v4 compatibility |
| Namespace propagation fix (#5940, #6044) | May affect `kustomize build config/default` output if namespace transformer is used | Verify output is identical or improved |
| Allow empty patch files (#5990) | No impact unless project has empty patches | None |
| `pkg/errors` dependency removed (#6057) | No impact — kustomize is a standalone binary | None |
| Regex support for replacement selectors (#5863) | Not used in current manifests | None |
| PatchArgs API type (#5930) | Not used | None |

#### Namespace Propagation Behavioral Change

Kustomize v5.8.1 changed how namespaces propagate to Helm charts: instead of naively replacing `.metadata.namespace` on all resources, it now passes the namespace parameter to Helm directly. This project's `helmchart` target uses `kustomize build ./config/helmchart` which doesn't involve Helm charts within kustomize (it's a pure kustomize overlay). So this change has **no impact** on the helmchart generation.

However, `kustomize build config/default` does use namespace transformer settings. Verify the output is semantically equivalent after the upgrade.

### Makefile Changes (Complete)

| Location | Current | New |
|----------|---------|-----|
| (NEW — near line 21) | (doesn't exist) | `OPM_VERSION ?= v1.73.0` |
| Line 9 | `KUSTOMIZE_VERSION ?= v5.4.3` | `KUSTOMIZE_VERSION ?= v5.8.1` |
| Line 366 (opm target) | `...releases/download/v1.23.0/$${OS}-$${ARCH}-opm` | `...releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm` |

### Complete Makefile Diff

#### 1. Add OPM_VERSION variable

After `OPERATOR_SDK_VERSION ?= v1.31.0` (line 21), add:

```makefile
OPM_VERSION ?= v1.73.0
```

#### 2. Update KUSTOMIZE_VERSION

```makefile
# Before:
KUSTOMIZE_VERSION ?= v5.4.3
# After:
KUSTOMIZE_VERSION ?= v5.8.1
```

#### 3. Update opm target URL

```makefile
# Before (line 366):
curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.23.0/$${OS}-$${ARCH}-opm ;\
# After:
curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/$(OPM_VERSION)/$${OS}-$${ARCH}-opm ;\
```

### Files to Change (Complete List)

| File | Change |
|------|--------|
| `Makefile` line 9 | `KUSTOMIZE_VERSION ?= v5.4.3` → `v5.8.1` |
| `Makefile` (new, ~line 22) | Add `OPM_VERSION ?= v1.73.0` |
| `Makefile` line 366 | Replace hardcoded `v1.23.0` in `curl` URL with `$(OPM_VERSION)` |

### Files NOT to Change

- `go.mod` / `go.sum` — OPM and kustomize are standalone binaries, not Go library dependencies
- `.github/workflows/*.yaml` — CI does not reference OPM_VERSION or KUSTOMIZE_VERSION directly; it uses the Makefile targets
- `config/` kustomize manifests — manifest content changes are Story 10.0a (kustomize syntax migration)
- `config/helmchart/` — chart template content is unchanged
- Any Go source files — these tool upgrades don't affect Go code
- `bundle.Dockerfile` — no OPM or kustomize references
- `Dockerfile` / `ci.Dockerfile` — no build tool references

### Anti-Patterns to Avoid

- **DO NOT** migrate from `opm index add` to File-Based Catalog (FBC) format — that is a separate, larger scope change
- **DO NOT** suppress the `opm index add` deprecation warning — it's informational and expected
- **DO NOT** change kustomize manifest syntax (`bases:` → `resources:`, etc.) — that's Story 10.0a
- **DO NOT** update Helm version — that's Story 10.2
- **DO NOT** update controller-gen, envtest, or other tool versions — those are separate stories or already done
- **DO NOT** modify the `go-install-tool` function — the kustomize module path is unchanged (`v5`)
- **DO NOT** add the OPM version to CI workflow files — CI uses the Makefile target, not a separate version variable

### Verification Commands (Run After Completion)

```bash
rm -f bin/opm bin/kustomize bin/kustomize-*
make opm
bin/opm version
make kustomize
bin/kustomize version
make manifests
make helmchart
make fmt vet test
```

### Rollback Strategy

If either tool upgrade causes unexpected issues:

**OPM rollback:**
1. Change `OPM_VERSION ?= v1.73.0` back to `OPM_VERSION ?= v1.23.0` (or revert to hardcoded URL)
2. Delete `bin/opm` and re-download

**Kustomize rollback:**
1. Revert `KUSTOMIZE_VERSION` to `v5.4.3`
2. Delete `bin/kustomize bin/kustomize-*` and re-download

### Previous Story Intelligence

From Story 10.3 (golangci-lint upgrade) and Story 10.2 (Helm upgrade):
- The `go-install-tool` function uses a version-stamped binary pattern (`$(1)-$(3)`) with a symlink. Kustomize follows this pattern. OPM does NOT — it uses the `ifeq/wildcard` pattern like operator-sdk and helm. Keep OPM's existing download pattern (curl-based, not go-install-tool).
- CI workflows reference tool versions via Makefile targets, not separate variables. Introducing `OPM_VERSION` as a Makefile variable is sufficient — no CI file changes needed.
- Story 10.2 noted that Helm 4 download URL pattern is unchanged from v3 — similarly, OPM's download URL pattern is unchanged from v1.23 to v1.73.

### Project Structure Notes

- Alignment with unified project structure: No structural changes — this story only bumps tool versions and introduces a version variable
- Detected conflicts or variances: The OPM version was the only tool with a hardcoded version in the download URL instead of using a variable — this story corrects that inconsistency

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.4]
- [Source: _bmad-output/project-context.md#Build & Dev Tooling]
- [Source: _bmad-output/implementation-artifacts/10-2-upgrade-helm-from-v3-11-to-v4.md]
- [Source: _bmad-output/implementation-artifacts/10-3-upgrade-golangci-lint-from-v1-59-to-v2.md]
- [OPM v1.73.0 release](https://github.com/operator-framework/operator-registry/releases/tag/v1.73.0)
- [Kustomize v5.8.1 release](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv5.8.1)
- [OPM File-Based Catalog documentation](https://olm.operatorframework.io/docs/tasks/creating-a-catalog/)
- [Kustomize Helm v4 interop fix](https://github.com/kubernetes-sigs/kustomize/pull/6016)

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Previous Review Notes (First Attempt — GPT 5.4)

> These findings are from a code review of the first implementation attempt. Commit references are no longer valid but the architectural guidance remains relevant.

- **[Decision]** AC2 requires `make catalog-build` to succeed, but this command requires a pushed bundle image. The first attempt only confirmed `make opm` works. Decide whether to accept alternate evidence or require a real `make catalog-build` run.
- **[Patch]** After adopting the version-suffixed opm binary + symlink pattern, the story's verification commands and rollback instructions should reference `bin/opm-$(OPM_VERSION)` not just `bin/opm`.
