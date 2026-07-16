# Story 8.3: Upgrade Makefile K8s-coupled tool versions

Status: ready-for-dev

## Story

As an operator developer,
I want to upgrade the Makefile tool versions that must track the K8s stack (controller-gen, envtest, Kind, kubectl),
So that tooling is compatible with the new controller-runtime v0.24 / K8s v0.36 versions from Story 8.2 and CRD/RBAC generation, envtest binaries, and integration test clusters all work correctly.

## Acceptance Criteria

1. **Given** CONTROLLER_TOOLS_VERSION is v0.14.0, **When** it is updated to v0.21.0, **Then** `make manifests` and `make generate` produce valid CRDs and deepcopy code that compile and pass validation.

2. **Given** ENVTEST_VERSION is release-0.19 and ENVTEST_K8S_VERSION is 1.29.0, **When** they are updated to release-0.24 and 1.36.0 respectively, **Then** `make test` downloads the correct K8s 1.36 envtest binaries and all unit tests pass.

3. **Given** KIND_VERSION is v0.27.0, **When** it is updated to v0.32.0, **Then** `make kind-setup` creates a Kind cluster using the v1.36.1 node image and the cluster boots correctly.

4. **Given** KUBECTL_VERSION is v1.29.0, **When** it is updated to v1.36.1, **Then** the downloaded kubectl binary matches the Kind cluster K8s version and all `kubectl` operations in the Makefile work.

5. **Given** Kind v0.32.0 uses kubeadm v1beta4 for K8s 1.36+ and auto-translates unversioned patches, **When** `integration/cluster-kind.yaml.tpl` is reviewed, **Then** the existing unversioned kubeadmConfigPatches work correctly without changes (or are updated if necessary).

6. **Given** controller-gen v0.21.0 requires Go 1.26 (matching the project's Go version from Story 8.1), **When** the `controller-gen` Makefile target is updated, **Then** the `go-install-tool-compat` workaround is removed in favor of the standard `go-install-tool`.

7. **Given** controller-gen v0.21.0 may produce different CRD output (new markers, schema changes), **When** `make manifests generate` is run, **Then** all generated files are committed and the operator compiles and passes tests.

## Tasks / Subtasks

- [ ] Task 1: Update Makefile version variables (AC: #1, #2, #3, #4)
  - [ ] 1.1 Change `CONTROLLER_TOOLS_VERSION ?= v0.14.0` to `CONTROLLER_TOOLS_VERSION ?= v0.21.0`
  - [ ] 1.2 Change `ENVTEST_VERSION ?= release-0.19` to `ENVTEST_VERSION ?= release-0.24`
  - [ ] 1.3 Change `ENVTEST_K8S_VERSION ?= 1.29.0` to `ENVTEST_K8S_VERSION ?= 1.36.0`
  - [ ] 1.4 Change `KIND_VERSION ?= v0.27.0` to `KIND_VERSION ?= v0.32.0`
  - [ ] 1.5 Change `KUBECTL_VERSION ?= v1.29.0` to `KUBECTL_VERSION ?= v1.36.1`

- [ ] Task 2: Remove `go-install-tool-compat` workaround for controller-gen (AC: #6)
  - [ ] 2.1 Change the `controller-gen` install target from `$(call go-install-tool-compat,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION),go1.22.12)` to `$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))`
  - [ ] 2.2 Remove the `go-install-tool-compat` `define` block entirely (lines 326–339) if no other targets use it — currently only controller-gen uses it
  - [ ] 2.3 Remove the comment above `go-install-tool-compat` that references the workaround

- [ ] Task 3: Delete stale tool binaries from `bin/` (AC: #1, #2, #3)
  - [ ] 3.1 Run `rm -f bin/controller-gen* bin/setup-envtest* bin/kind*` to remove old versioned binaries and symlinks, forcing re-download at new versions
  - [ ] 3.2 Do NOT remove `bin/kubectl` (it's downloaded via curl, not `go install`, so the Makefile's `ifeq (,$(wildcard $(KUBECTL)))` guard handles re-download only when the binary is absent) — remove it with `rm -f bin/kubectl` to force re-download at the new version

- [ ] Task 4: Regenerate CRDs and deepcopy with controller-gen v0.21.0 (AC: #1, #7)
  - [ ] 4.1 Run `make manifests` — generates CRDs and RBAC using controller-gen v0.21.0
  - [ ] 4.2 Run `make generate` — regenerates `zz_generated.deepcopy.go`
  - [ ] 4.3 Diff the generated files against the previous version to review changes:
    - CRD files in `config/crd/bases/` — expect possible schema differences (new markers like `k8s:immutable`, `k8s:enum` support, updated OpenAPI descriptions)
    - RBAC ClusterRole in `config/rbac/role.yaml` — check for any changes
    - `api/v1alpha1/zz_generated.deepcopy.go` — expect possible code-generation style changes
  - [ ] 4.4 Review the `CRD_OPTIONS` variable: the current value `"crd:trivialVersions=true,preserveUnknownFields=false"` uses deprecated controller-gen flags — `trivialVersions` was removed in controller-gen v0.15+ (all CRDs are single-version by default) and `preserveUnknownFields` was removed too. If `make manifests` fails or warns about these, update the `manifests` target to remove `CRD_OPTIONS` usage and use the bare `crd` generator
  - [ ] 4.5 If the `CRD_OPTIONS` variable is not referenced in the `manifests` target (check: it currently is NOT used — the target uses inline `crd` in the controller-gen command), then remove the `CRD_OPTIONS` variable definition (line 80) as dead code

- [ ] Task 5: Verify Kind cluster configuration for v0.32.0 (AC: #5)
  - [ ] 5.1 Review `integration/cluster-kind.yaml.tpl` — the config uses `apiVersion: kind.x-k8s.io/v1alpha4` and unversioned `kubeadmConfigPatches` with map-format `kubeletExtraArgs`. Kind v0.32.0 auto-translates map-format `kubeletExtraArgs` to v1beta4 list format. Verify no explicit `apiVersion` is set in the patches (currently none is set — confirmed safe)
  - [ ] 5.2 Verify the `kind-setup` target in Makefile — it uses `--image docker.io/kindest/node:$(KUBECTL_VERSION)` to set the node image. With KUBECTL_VERSION=v1.36.1, this becomes `kindest/node:v1.36.1` which is the default node image for Kind v0.32.0
  - [ ] 5.3 The Envoy LB change (HAProxy→Envoy) only affects multi-control-plane clusters — this project uses a single control-plane node, so NO IMPACT

- [ ] Task 6: Compilation and test verification (AC: #1, #2, #7)
  - [ ] 6.1 Run `go build ./...` to verify compilation (CRD/deepcopy changes don't break anything)
  - [ ] 6.2 Run `go vet ./...` to verify static analysis
  - [ ] 6.3 Run `make test` to verify unit tests pass with new envtest K8s 1.36 binaries
  - [ ] 6.4 If `make test` fails due to envtest binary download issues, verify `$(ENVTEST) use 1.36.0 -p path` works — the envtest-v1.36.0 and envtest-v1.36.2 releases are available from controller-tools

- [ ] Task 7: Update project-context.md (AC: #1, #6)
  - [ ] 7.1 Update `controller-gen v0.14.0` → `controller-gen v0.21.0` in the Build & Dev Tooling section
  - [ ] 7.2 Update `Kind v0.27.0, kubectl v1.29.0` → `Kind v0.32.0, kubectl v1.36.1`
  - [ ] 7.3 Update `Container: golang:1.22 builder` → `Container: golang:1.26 builder` if not already done by Story 8.1
  - [ ] 7.4 Remove the `CRD_OPTIONS` reference if the variable was cleaned up in Task 4
  - [ ] 7.5 Update the webhook pattern documentation if controller-gen v0.21.0 changed anything about webhook markers (unlikely but verify)
  - [ ] 7.6 Update the webhook interface documentation from `webhook.Defaulter`/`webhook.Validator` to `webhook.CustomDefaulter[T]`/`webhook.CustomValidator[T]` and the `SetupWebhookWithManager` pattern to the new `ctrl.NewWebhookManagedBy(mgr, r).WithDefaulter(r).WithValidator(r).Complete()` form — if not already done by Story 8.2

## Dev Notes

### Scope Boundary

This story covers ONLY the Makefile tool version variables and their immediate consequences (CRD regeneration, envtest binary version, Kind/kubectl versions, controller-gen install target cleanup). It does NOT update:
- Go version (Story 8.1 — prerequisite, must be done first)
- controller-runtime or K8s client libs in go.mod (Story 8.2 — prerequisite, must be done first)
- CI workflow files (.github/workflows/*) (Story 8.4)
- Dockerfiles (Story 8.4)
- Operator SDK version (Epic 10, Story 10.1)
- Helm version (Epic 10, Story 10.2)
- golangci-lint version (Epic 10, Story 10.3)

**Prerequisites:** Stories 8.1 (Go 1.26) and 8.2 (controller-runtime v0.24 / K8s v0.36) MUST be completed first. controller-gen v0.21.0 requires Go 1.26. The ENVTEST_VERSION tracks the controller-runtime release branch. The Kind node image version should match the K8s libs version.

### Version Compatibility Matrix

| Variable | Current | Target | Rationale |
|----------|---------|--------|-----------|
| CONTROLLER_TOOLS_VERSION | v0.14.0 | v0.21.0 | CT v0.21 requires k8s.io/* v0.36 + Go 1.26 |
| ENVTEST_VERSION | release-0.19 | release-0.24 | Tracks controller-runtime release branch |
| ENVTEST_K8S_VERSION | 1.29.0 | 1.36.0 | Matches K8s libs v0.36 → K8s 1.36 |
| KIND_VERSION | v0.27.0 | v0.32.0 | Defaults to K8s 1.36.1 node image |
| KUBECTL_VERSION | v1.29.0 | v1.36.1 | Matches Kind node K8s version |

### controller-gen v0.14.0 → v0.21.0 Breaking Changes

Seven minor versions are spanned (v0.15 through v0.21). Key changes:

**v0.15 (K8s v0.30):**
- Removed `trivialVersions` and `preserveUnknownFields` CRD options — these were deprecated since v0.12. The current `CRD_OPTIONS` variable uses both, but the `manifests` target does NOT reference `CRD_OPTIONS` (it uses inline `crd` in the controller-gen command). So the variable is dead code — remove it.
- Added `k8s:validation:cel` marker support for CEL validation rules.

**v0.16 (K8s v0.31):**
- Added `+kubebuilder:validation:XIntOrString` marker.
- Improved CRD description generation — descriptions may change slightly in generated CRDs.

**v0.17 (K8s v0.32):**
- No major breaking changes for CRD generation.

**v0.18 (K8s v0.33):**
- Added `k8s:enum` marker support (auto-infers enum values from Go `type X string` with iota-like constants).
- Schema generation improvements.

**v0.19 (K8s v0.34):**
- Required Go 1.24 minimum.
- Deepcopy generation improvements.

**v0.20 (K8s v0.35):**
- Required Go 1.25 minimum.
- Added `k8s:immutable` marker support.

**v0.21 (K8s v0.36):**
- Required Go 1.26 minimum.
- Added `kubebuilder:externalDoc` marker.
- Added `Name` parameter support for webhook markers.
- This is the target version, matching the k8s.io/* v0.36 dependency stack.

**Expected CRD output changes:** The generated CRDs in `config/crd/bases/` will likely have differences in OpenAPI schema formatting, description text, and possibly new schema properties if any Go types match the auto-inference patterns (e.g., `k8s:enum` for string-typed constants). Review diffs carefully but expect them to be cosmetic/additive.

### `CRD_OPTIONS` Variable Cleanup

Line 80 of the Makefile defines:
```makefile
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"
```

This variable is **dead code** — it is not referenced by the `manifests` target (line 118–119):
```makefile
manifests: controller-gen
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
```

The `manifests` target uses bare `crd` (no options), not `$(CRD_OPTIONS)`. Moreover, `trivialVersions` and `preserveUnknownFields` were removed from controller-gen in v0.15. Remove the `CRD_OPTIONS` variable definition to prevent confusion.

### `go-install-tool-compat` Removal

The `go-install-tool-compat` function was introduced to work around controller-gen v0.14.0 being incompatible with Go 1.25+ (v0.14 required Go ≤1.22). It pins `GOTOOLCHAIN=go1.22.12` during `go install`.

With controller-gen upgraded to v0.21.0 (which requires Go 1.26, matching the project's Go version), this workaround is no longer needed. Switch the `controller-gen` target back to the standard `go-install-tool` function and remove the `go-install-tool-compat` definition entirely.

**Current (Makefile lines 296–298):**
```makefile
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool-compat,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION),go1.22.12)
```

**Target:**
```makefile
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))
```

### Kind v0.32.0 Breaking Changes Impact

**HAProxy → Envoy for LB:** Only affects multi-control-plane clusters. This project uses a single control-plane node in `integration/cluster-kind.yaml.tpl`. **NO IMPACT.**

**kubeadm v1beta4 for K8s 1.36+:** The existing `cluster-kind.yaml.tpl` uses unversioned kubeadmConfigPatches (no explicit `apiVersion` in the patch). Kind v0.32.0 auto-translates unversioned patches:
- Map-format `kubeletExtraArgs` → auto-translated to v1beta4 list format.
- The current patch sets `kubeletExtraArgs: node-labels: "ingress-ready=true"` in map format — this will be auto-converted. **NO CODE CHANGE NEEDED.**

**Node image:** Kind v0.32.0's default node image is `kindest/node:v1.36.1`. The Makefile's `kind-setup` target uses `--image docker.io/kindest/node:$(KUBECTL_VERSION)`. With KUBECTL_VERSION=v1.36.1, this resolves to the correct image. The containerd upgrade in Kind v0.32.0 requires using Kind v0.32.0 for `kind load` — the version variable update handles this.

### Envtest Binary Download

The `make test` target uses `$(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path` to download and locate envtest binaries. With ENVTEST_K8S_VERSION=1.36.0:
- The `setup-envtest` tool (installed from `release-0.24` branch) will download K8s 1.36.0 binaries (etcd + kube-apiserver) from the controller-tools envtest releases index.
- The envtest-v1.36.0 release is available on the controller-tools GitHub releases page.
- First run will take a few seconds to download; subsequent runs use cached binaries.

### Files to Modify

| File | Change |
|------|--------|
| `Makefile` | Version variables (5 vars), controller-gen install target, remove `go-install-tool-compat`, remove `CRD_OPTIONS` |
| `config/crd/bases/*.yaml` | Regenerated by `make manifests` with controller-gen v0.21.0 |
| `config/rbac/role.yaml` | Regenerated by `make manifests` |
| `api/v1alpha1/zz_generated.deepcopy.go` | Regenerated by `make generate` |
| `_bmad-output/project-context.md` | Tool version references |

**Files NOT modified in this story:**
- `integration/cluster-kind.yaml.tpl` — unversioned patches work with Kind v0.32.0 auto-translation
- `.github/workflows/*` — CI changes are Story 8.4
- `Dockerfile`, `ci.Dockerfile` — Docker changes are Story 8.4
- `go.mod`, `go.sum` — dependency changes are Stories 8.1 and 8.2
- Source code files (`*.go`) — no source changes expected (only generated files change)

### Previous Story Intelligence

**Story 8.1** (Go 1.22 → 1.26):
- Go 1.26 is a drop-in version bump with no source code changes
- `go.mod` uses `go 1.26` with `toolchain go1.26.4`
- The `go-install-tool-compat` workaround was introduced specifically because controller-gen v0.14.0 couldn't build with Go 1.25+. Now that controller-gen v0.21.0 requires Go 1.26, the workaround is obsolete.

**Story 8.2** (controller-runtime v0.17 → v0.24):
- All 44 webhook files migrated from `webhook.Defaulter`/`webhook.Validator` to `webhook.CustomDefaulter[T]`/`webhook.CustomValidator[T]`
- controller-runtime v0.24 uses k8s.io/* v0.36 — matching controller-gen v0.21.0's dependency alignment
- The `SetupWebhookWithManager` pattern changed to `ctrl.NewWebhookManagedBy(mgr, r).WithDefaulter(r).WithValidator(r).Complete()`
- ENVTEST_VERSION `release-0.19` was left unchanged in Story 8.2 (deferred to this story)

### Recommended Execution Order

1. **Update Makefile version variables** (Task 1) — all five at once
2. **Fix controller-gen install target** (Task 2) — remove compat workaround
3. **Clean stale binaries** (Task 3) — force re-download
4. **Run `make manifests generate`** (Task 4) — regenerate with new controller-gen
5. **Review and commit generated file diffs** (Task 4.3)
6. **Run `make test`** (Task 6) — verify everything works end-to-end
7. **Update project-context.md** (Task 7)

### Testing Standards

- Run `make manifests generate` to regenerate all CRDs, RBAC, and deepcopy
- Run `go build ./...` to verify compilation
- Run `go vet ./...` for static analysis
- Run `make test` (envtest-based unit tests) — this validates both envtest binary download and test execution
- Integration tests (`make integration`) require a Kind cluster — verify locally if feasible (Kind v0.32.0 + K8s 1.36.1 node). If the Kind cluster was created with the old Kind version, delete it first (`kind delete cluster --name vault-config-operator`) to force recreation with the new version and node image
- `make kind-setup` has logic to detect image/port mismatches and recreate automatically

### Project Structure Notes

- No new source files created; only Makefile and generated files modified
- The CRD files in `config/crd/bases/` may have significant diffs due to controller-gen version jump — review for correctness but expect them to be valid
- The `api/v1alpha1/zz_generated.deepcopy.go` generated file should have minimal changes (deepcopy generation is stable)

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 8, Story 8.3]
- [Source: _bmad-output/project-context.md — Build & Dev Tooling versions]
- [Source: Makefile — current version variables and tool install targets]
- [Source: integration/cluster-kind.yaml.tpl — Kind cluster configuration]
- [Source: controller-tools v0.21.0 release — https://github.com/kubernetes-sigs/controller-tools/releases/tag/v0.21.0]
- [Source: controller-tools compatibility matrix — CT v0.21 requires k8s.io/* v0.36, Go 1.26]
- [Source: Kind v0.32.0 release — https://github.com/kubernetes-sigs/kind/releases/tag/v0.32.0]
- [Source: controller-runtime v0.24.0 release — https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.24.0]
- [Source: setup-envtest README — https://github.com/kubernetes-sigs/controller-runtime/blob/main/tools/setup-envtest/README.md]
- [Source: Kubebuilder envtest docs — https://book.kubebuilder.io/reference/envtest.html]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
