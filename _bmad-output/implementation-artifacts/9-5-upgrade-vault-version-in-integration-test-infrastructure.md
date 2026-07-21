# Story 9.5: Upgrade Vault version in integration test infrastructure to 2.0.3

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want to upgrade the Vault version used in integration tests and local development from 1.19.x to 2.0.3,
So that we test against the current supported Vault release and verify compatibility with the Vault 2.0 major version.

## Acceptance Criteria

1. **Given** integration tests use Vault 1.19.0, **When** all Vault version references are updated to 2.0.3 and any breaking changes are addressed, **Then** `make integration` passes against the new Vault version.

2. **Given** the Vault Helm chart version must match the Vault version, **When** `VAULT_CHART_VERSION` is updated to 0.34.0 (ships Vault 2.0.3 by default), **Then** `make deploy-vault` deploys the correct Vault version.

3. **Given** Vault 2.0 is a major version change from 1.19, **When** the Vault 2.0 CHANGELOG and migration guide are reviewed, **Then** any operator code or test infrastructure affected by breaking changes is adapted.

4. **Given** local-development values reference `1.19.2-ubi` images, **When** they are updated to `2.0.3-ubi`, **Then** the local development environment deploys Vault 2.0.3.

5. **Given** all version changes are applied, **When** `go build ./...` and `make test` are run, **Then** both succeed without source code changes (the Vault CLI binary version only affects integration tests).

## Tasks / Subtasks

- [ ] Task 1: Review Vault 2.0 breaking changes for operator impact (AC: #3)
  - [ ] 1.1 Review path canonicalization change — Vault rejects `//`, `/../`, `/./` paths. Verify the operator never constructs non-canonical paths (it uses `spec.path` + `metadata.name` composition — should be safe).
  - [ ] 1.2 Review container runtime change — Vault 2.0 runs as non-root `vault` user by default. Verify sidecar containers in `vault-values.yaml` can still access shared volumes (emptyDir is world-writable, secrets are readable by all — should be safe).
  - [ ] 1.3 Review IPC_LOCK change — was required in 2.0.0, **removed** in 2.0.2. Since target is 2.0.3, no IPC_LOCK capability needed.
  - [ ] 1.4 Review plugin registration changes — verify `vault plugin register` CLI syntax is unchanged. The integration tests register `vault-plugin-secrets-github` v1.3.0.
  - [ ] 1.5 Review `vault operator init` / `vault operator unseal` CLI changes — used by auto-initializer and auto-unsealer sidecars.
  - [ ] 1.6 Review `vault auth enable kubernetes` + `vault write auth/kubernetes/config` — used by vault-admin-initializer sidecar.
  - [ ] 1.7 Review `vault policy write` / `vault policy list` — used by vault-admin-initializer.
  - [ ] 1.8 Document any required changes or confirm all CLI operations are backward-compatible.
- [ ] Task 2: Update Makefile version variables (AC: #1, #2)
  - [ ] 2.1 Change `VAULT_VERSION ?= 1.19.0` → `VAULT_VERSION ?= 2.0.3` (line 16)
  - [ ] 2.2 Change `VAULT_CHART_VERSION ?= 0.30.0` → `VAULT_CHART_VERSION ?= 0.34.0` (line 18)
- [ ] Task 3: Update integration/vault-values.yaml (AC: #1)
  - [ ] 3.1 Change `image: hashicorp/vault:1.19.0` → `image: hashicorp/vault:2.0.3` in `auto-initializer` sidecar (line 61)
  - [ ] 3.2 Change `image: hashicorp/vault:1.19.0` → `image: hashicorp/vault:2.0.3` in `auto-unsealer` sidecar (line 89)
  - [ ] 3.3 Change `image: hashicorp/vault:1.19.0` → `image: hashicorp/vault:2.0.3` in `github-module-loader` sidecar (line 118)
  - [ ] 3.4 Change `image: hashicorp/vault:1.19.0` → `image: hashicorp/vault:2.0.3` in `vault-admin-initializer` sidecar (line 151)
- [ ] Task 4: Update config/local-development/vault-values.yaml (AC: #4)
  - [ ] 4.1 Change `tag: "1.19.2-ubi"` → `tag: "2.0.3-ubi"` in `injector.agentImage` (line 16)
  - [ ] 4.2 Change `tag: "1.19.2-ubi"` → `tag: "2.0.3-ubi"` in `server.image` (line 26)
  - [ ] 4.3 Change `image: registry.connect.redhat.com/hashicorp/vault:1.19.2-ubi` → `image: registry.connect.redhat.com/hashicorp/vault:2.0.3-ubi` in `auto-initializer` sidecar (line 111)
  - [ ] 4.4 Change `image: registry.connect.redhat.com/hashicorp/vault:1.19.2-ubi` → `image: registry.connect.redhat.com/hashicorp/vault:2.0.3-ubi` in `auto-unsealer` sidecar (line 148)
  - [ ] 4.5 Change `image: registry.connect.redhat.com/hashicorp/vault:1.19.2-ubi` → `image: registry.connect.redhat.com/hashicorp/vault:2.0.3-ubi` in `github-module-loader` sidecar (line 181)
- [ ] Task 5: Verify Vault CLI binary download URL works (AC: #5)
  - [ ] 5.1 Confirm `https://releases.hashicorp.com/vault/2.0.3/vault_2.0.3_linux_amd64.zip` resolves (the Makefile `vault` target downloads this)
  - [ ] 5.2 If URL pattern changed for Vault 2.x, update the download URL in Makefile (line 486)
- [ ] Task 6: Run full test suite (AC: #1, #5)
  - [ ] 6.1 Run `go build ./...` — confirm compilation (no source changes expected)
  - [ ] 6.2 Run `make test` — confirm unit tests pass (no Vault interaction)
  - [ ] 6.3 Run `make integration` — confirm full integration test suite passes against Vault 2.0.3
  - [ ] 6.4 If any integration tests fail, diagnose whether it's a Vault 2.0 behavioral change and fix accordingly
- [ ] Task 7: Update project-context.md Vault version reference (AC: #1)
  - [ ] 7.1 Change `Vault 1.19.0 (integration testing)` → `Vault 2.0.3 (integration testing)` in the Build & Dev Tooling section

## Dev Notes

### Vault 2.0 Breaking Changes — Operator Impact Assessment

| Breaking Change | Introduced | Impact on This Project |
|-----------------|-----------|----------------------|
| Path canonicalization — rejects `//`, `/../`, `/./` | 2.0.0 | **None.** Operator composes paths as `spec.path` + `/` + `metadata.name`. No double-slashes or relative refs. |
| `sys/rekey`, `sys/generate-root` require auth token | 2.0.0 | **None.** Operator never calls these endpoints. Integration test sidecars don't use them either. |
| Azure auth config precedence over env vars | 2.0.0 | **None for tests.** Azure auth is not testable in Kind (cloud provider). Operator code just writes the config — Vault's internal precedence doesn't affect the write API. |
| Docker image runs as non-root `vault` user | 2.0.0 | **Low risk.** Sidecar containers run vault CLI commands. The emptyDir `plugins` volume is world-writable (default mode 0777). The secret `vault-root-token` volume is readable by all (default mode 0644). No permission issues expected. |
| IPC_LOCK capability required | 2.0.0 | **NOT APPLICABLE.** Removed in 2.0.2. Target is 2.0.3 — no IPC_LOCK needed. |
| RSA key cap at 8192 bits | 2.0.2 | **None for tests.** PKI integration tests use default key sizes. |
| Identity templates disallow wildcards/globs | 2.0.1 | **None.** Identity OIDC integration tests don't use wildcard templates. |
| UBI images now built on UBI 10 | 2.0.0 | **Cosmetic only.** Local-dev uses UBI images — UBI 10 is backward-compatible. |

### Helm Chart Compatibility

- **Chart v0.34.0** (released 2026-07-02): defaults to Vault 2.0.3, tested with Vault 2.0.3, 1.21-1.19 and Kubernetes v1.36-1.32.
- **Also ships:** vault-csi-provider v1.7.3, vault-k8s (injector) v1.7.5.
- The `integration/vault-values.yaml` does NOT set `server.image` (relies on chart defaults for the main server container). The 4 sidecar images are explicit overrides that must be updated.
- The `config/local-development/vault-values.yaml` explicitly sets `server.image.tag` — must be updated.

### Vault CLI Binary Download

The Makefile downloads the Vault CLI binary from `https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_${OS}_${ARCH}.zip`. HashiCorp's release infrastructure uses the same URL pattern for 2.x releases. Verify `vault_2.0.3_linux_amd64.zip` exists at that URL before proceeding.

### Integration Test Sidecar Containers

The `integration/vault-values.yaml` has 4 sidecar containers using the Vault image:

1. **auto-initializer** — runs `vault operator init`, `vault status`, creates K8s secret via kubectl
2. **auto-unsealer** — runs `vault operator unseal`, `vault status`
3. **github-module-loader** — runs `vault plugin register`, `vault plugin list`
4. **vault-admin-initializer** — runs `vault auth enable kubernetes`, `vault write auth/kubernetes/config`, `vault write auth/kubernetes/role/policy-admin`, `vault policy write`, `vault policy list`

All CLI commands used are core Vault operations unchanged in 2.0. The `vault plugin register` syntax is the same. The `vault auth enable` and `vault write` commands are standard stable API paths.

### Files to Modify

| File | Change | Lines Affected |
|------|--------|---------------|
| `Makefile` | `VAULT_VERSION` 1.19.0 → 2.0.3, `VAULT_CHART_VERSION` 0.30.0 → 0.34.0 | 16, 18 |
| `integration/vault-values.yaml` | 4 image tags: `hashicorp/vault:1.19.0` → `hashicorp/vault:2.0.3` | 61, 89, 118, 151 |
| `config/local-development/vault-values.yaml` | 2 `tag:` fields + 3 `image:` fields: `1.19.2-ubi` → `2.0.3-ubi` | 16, 26, 111, 148, 181 |
| `_bmad-output/project-context.md` | Update Vault version in Build & Dev Tooling section | ~37 |

**Files NOT modified:**
- No Go source code changes — Vault API compatibility is handled by vault/api library (upgraded in Story 9.1)
- No test code changes — integration tests use Vault through the operator, not direct CLI
- No CRD changes
- No Dockerfile changes
- No CI workflow changes (CI uses Makefile targets which pick up the version variables)

### Potential Failure Scenarios During `make integration`

If integration tests fail after the upgrade, investigate these areas:

1. **Plugin registration failure** — If `vault plugin register` fails, check if Vault 2.0 changed the plugin catalog API or requires additional parameters. The github secrets plugin v1.3.0 was built against Vault SDK <2.0 — it should still work as Vault maintains plugin backward compatibility.

2. **Kubernetes auth config failure** — If `vault write auth/kubernetes/config` fails, check if the `token_reviewer_jwt` parameter is still accepted (it has been deprecated in favor of Vault's own service account since Vault 1.21+, but should still work).

3. **Volume permission errors** — If sidecar containers can't write to `/usr/local/libexec/vault`, the non-root `vault` user may lack permissions. Fix: add `securityContext: {runAsUser: 100}` matching the vault user UID, or explicitly set emptyDir permissions.

4. **Init container kubectl/jq download failures** — Unrelated to Vault version, but if the test environment has network issues. These containers use UBI 8 images (unchanged).

### Anti-Pattern Prevention

- **DO NOT** upgrade the Vault Go client library (`hashicorp/vault/api`) in this story — that was Story 9.1.
- **DO NOT** change any operator source code (`*.go` files) unless a Vault 2.0 behavioral change breaks compilation or API contract. This story is infrastructure-only.
- **DO NOT** upgrade Helm chart values structure — chart v0.34.0 is backward-compatible with existing values files.
- **DO NOT** change the GitHub secrets plugin version (v1.3.0 in integration, v2.0.0 in local-dev) — plugin version management is out of scope.
- **DO NOT** change `KUBECTL_VERSION`, `KIND_VERSION`, or other tool versions — only Vault-related versions change.
- **DO NOT** add `securityContext` or capability settings unless integration tests actually fail due to permission issues.
- **DO NOT** change the `vault-k8s` injector version in local-dev values unless tests fail — the injector is disabled in integration tests and optional in local-dev.

### Scope Boundary

**IN scope:** Update all Vault version references (Makefile, Helm values files, project-context.md), verify Vault 2.0 CLI compatibility, run full test suite, fix any breakage caused by Vault 2.0 behavioral changes.

**OUT of scope:**
- Vault Go client library upgrade (Story 9.1 — already done)
- Test dependency upgrades (Story 9.2)
- Peripheral dependency upgrades (Story 9.3)
- pkg/errors migration (Story 9.4)
- Any operator source code refactoring
- Vault Helm chart values restructuring

### Previous Story Intelligence

**Story 9.4 (Evaluate pkg/errors migration, ready-for-dev):** Single-file code change removing `go-multierror`. No infrastructure impact. Confirmed all tests pass on current stack.

**Story 9.3 (Upgrade peripheral deps, ready-for-dev):** Pure `go.mod`/`go.sum` upgrade. Established pattern: `go get` + `go mod tidy` + `go build ./...` + `make test`.

**Story 9.1 (Upgrade vault/api, ready-for-dev):** Upgraded the Vault Go client library from v1.14 to v1.23. This ensures the operator's Vault API calls are compatible with Vault 2.0 server — the client library handles protocol negotiation.

**Epic 8 (Go + K8s Stack Upgrade, done 2026-07-19):** Story 8.3 updated Makefile tool versions (KIND, kubectl, etc.) following the same pattern. Story 8.4 adapted CI/Dockerfiles. Both established: change version constants, run tests, fix breakage.

### Project Structure Notes

- All changes are configuration/infrastructure files — no Go code paths affected
- Vault version appears in exactly 3 files: `Makefile`, `integration/vault-values.yaml`, `config/local-development/vault-values.yaml`
- The CI pipeline (`pr.yaml`, `push.yaml`) does NOT hardcode Vault versions — it uses `make integration` which reads from Makefile
- `project-context.md` documents `Vault 1.19.0` in the Build & Dev Tooling section — update for accuracy

### References

- [Source: Makefile:16 — VAULT_VERSION ?= 1.19.0]
- [Source: Makefile:18 — VAULT_CHART_VERSION ?= 0.30.0]
- [Source: Makefile:486 — Vault CLI download URL pattern]
- [Source: integration/vault-values.yaml:61,89,118,151 — 4 sidecar container images using hashicorp/vault:1.19.0]
- [Source: config/local-development/vault-values.yaml:16,26,111,148,181 — 5 vault image references using 1.19.2-ubi]
- [Source: _bmad-output/project-context.md:37 — "Vault 1.19.0 (integration testing)" in Build & Dev Tooling]
- [Source: _bmad-output/planning-artifacts/epics.md:1978-2008 — Story 9.5 epic definition]
- [Source: https://github.com/hashicorp/vault/releases/tag/v2.0.0 — Vault 2.0.0 release notes with breaking changes]
- [Source: https://docs.hashicorp.com/vault/docs/updates/important-changes — Vault breaking changes documentation]
- [Source: https://github.com/hashicorp/vault-helm/releases/tag/v0.34.0 — Helm chart 0.34.0 ships Vault 2.0.3, tested with K8s v1.36-1.32]
- [Source: https://docs.hashicorp.com/vault/docs/updates/release-notes — Vault 2.0.2 removed IPC_LOCK requirement]

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
