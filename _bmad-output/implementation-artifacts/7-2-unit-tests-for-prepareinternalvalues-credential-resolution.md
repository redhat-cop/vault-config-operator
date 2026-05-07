# Story 7.2: Unit Tests for `PrepareInternalValues` Credential Resolution

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want unit tests for the 15 types with non-trivial `PrepareInternalValues`,
So that credential resolution from Kubernetes Secrets, RandomSecrets, and VaultSecrets is verified without a live Vault.

## Acceptance Criteria

1. **Given** each of the 15 Epic 7.2 types with non-trivial `PrepareInternalValues` **When** the unit test suite runs **Then** each type has focused coverage for its credential-resolution or lookup branch using fake collaborators only (`kubeClient`, `vaultClient`, and `restConfig` as needed), with no live Vault and no integration harness

2. **Given** a type whose credentials are resolved from `Secret`, `VaultSecret`, or `RandomSecret` references **When** `PrepareInternalValues` is called with a context containing a fake `kubeClient` and/or fake `vaultClient` **Then** the internal retrieved fields are populated exactly as the corresponding payload-building logic expects

3. **Given** a `Policy` containing `${auth/<mount>/@accessor}` placeholders **When** `PrepareInternalValues` runs against a fake Vault `sys/auth` response **Then** the placeholders are replaced with returned accessor values, while the existing fast-path and nil-response behavior are preserved

4. **Given** the lookup-driven types (`KubernetesAuthEngineRole`, `EntityAlias`, `GroupAlias`, `RandomSecret`, and `KubernetesAuthEngineConfig`) **When** `PrepareInternalValues` is called with the minimal fake collaborators needed for their non-trivial branches **Then** resolved namespaces, IDs, generated secrets, and token reviewer JWT behavior are verified without changing production code solely for testability

5. **Given** the new tests are added **When** `make test` runs **Then** the `api/v1alpha1` unit suite passes with zero regressions

## Tasks / Subtasks

- [ ] Task 1: Add shared `PrepareInternalValues` unit-test helpers under `api/v1alpha1/` (AC: 1, 2, 3, 4)
  - [ ] 1.1: Create a test-only helper file such as `api/v1alpha1/prepareinternalvalues_test_helpers_test.go` in package `v1alpha1`
  - [ ] 1.2: Add a helper that builds a same-package context carrying `"kubeClient"`, `"vaultClient"`, and optional `"restConfig"` entries exactly like `controllers/commons.go`
  - [ ] 1.3: Reuse the existing `httptest` + Vault API client pattern from `api/v1alpha1/utils/vaultobject_test.go` rather than introducing production seams or new dependencies
  - [ ] 1.4: Register the CRD scheme plus `corev1` in fake-client helpers so Kubernetes `Secret`, `Namespace`, `RandomSecret`, `Entity`, and `Group` objects can be supplied directly in unit tests

- [ ] Task 2: Cover the shared root-credential resolution pattern for the config types that use `RootCredentialConfig` (AC: 1, 2)
  - [ ] 2.1: Extend `api/v1alpha1/databasesecretengineconfig_test.go` with a table-driven suite covering `Secret`, `VaultSecret`, and `RandomSecret` branches, including KV v1 vs KV v2 nested `data` handling and username precedence when `spec.username` is preset
  - [ ] 2.2: Extend `api/v1alpha1/rabbitmqsecretengineconfig_test.go`, `ldapauthengineconfig_test.go`, `azureauthengineconfig_test.go`, `azuresecretengineconfig_test.go`, `gcpauthengineconfig_test.go`, and `quaysecretengineconfig_test.go` with focused `PrepareInternalValues` tests that prove the correct retrieved credential fields are set for the branch each type actually uses
  - [ ] 2.3: For Azure/GCP config types, add one no-op/default-credential test and one non-default credential-resolution test so the `reflect.DeepEqual(..., RootCredentialConfig{...})` short-circuit is locked down
  - [ ] 2.4: Assert internal fields directly in same-package tests where that is the clearest signal (`retrievedUsername`, `retrievedPassword`, `retrievedClientID`, `retrievedClientPassword`, `retrievedToken`) instead of only asserting downstream payloads

- [ ] Task 3: Cover the non-`RootCredentialConfig` secret-resolution types (AC: 1, 2)
  - [ ] 3.1: Extend `api/v1alpha1/githubsecretengineconfig_test.go` with `PrepareInternalValues` coverage for Kubernetes SSH-auth secret resolution, Vault-secret key resolution, and the wrong-secret-type error path
  - [ ] 3.2: Extend `api/v1alpha1/kubernetessecretengineconfig_test.go` with `PrepareInternalValues` coverage for Kubernetes service-account-token secret resolution, Vault-secret key resolution, and the wrong-secret-type error path
  - [ ] 3.3: Keep these tests pure unit tests in `api/v1alpha1/`; do not move them into controllers or envtest integration flows

- [ ] Task 4: Cover the lookup- and placeholder-driven types (AC: 1, 3, 4)
  - [ ] 4.1: Extend `api/v1alpha1/policy_test.go` with tests for the no-placeholder fast path, successful `${auth/<mount>/@accessor}` replacement using a fake `sys/auth` response, and the existing `secret == nil` error behavior
  - [ ] 4.2: Extend `api/v1alpha1/kubernetesauthenginerole_test.go` with tests for explicit namespaces, label-selector namespace discovery, and the zero-match `__no_namespace__` fallback
  - [ ] 4.3: Extend `api/v1alpha1/entityalias_test.go` and `groupalias_test.go` with tests that pre-populate `Status.ID` to avoid create-side effects, then verify mount accessor lookup, canonical ID lookup, retrieved name/id fields, and payload-ready internal state
  - [ ] 4.4: If alias creation itself is tested, fully fake both Vault write and `kubeClient.Status().Update`; otherwise keep the story scoped to lookup/population behavior only

- [ ] Task 5: Cover the remaining specialized `PrepareInternalValues` types (AC: 1, 4)
  - [ ] 5.1: Extend `api/v1alpha1/randomsecret_test.go` with a deterministic inline-password-policy test and a password-policy-name test backed by a fake Vault `/sys/policies/password/<name>/generate` response
  - [ ] 5.2: Extend `api/v1alpha1/kubernetesauthengineconfig_test.go` with a no-op test for `TokenReviewerServiceAccount == nil` and, if feasible without production refactor, a fake TokenRequest API-server test that drives `GetJWTTokenWithDuration` via `rest.Config`
  - [ ] 5.3: If the TokenRequest HTTP branch proves too invasive for this story, lock down current nil/no-op and error-propagation behavior explicitly and document the remaining gap in the completion notes rather than changing production code for tests only

- [ ] Task 6: Regression verification (AC: 5)
  - [ ] 6.1: Run `go test ./api/v1alpha1/... -count=1` during development for fast iteration
  - [ ] 6.2: Run `make test` before closing the story
  - [ ] 6.3: Run `make fmt && make vet` if any new helper file or test formatting drift appears

## Dev Notes

### Artifact Availability

- No dedicated PRD, architecture, or UX files were found under `_bmad-output/planning-artifacts`
- This story is grounded in `epics.md`, `_bmad-output/project-context.md`, prior implementation artifacts, and the current `api/v1alpha1` code

### Reconcile Context Contract

`PrepareInternalValues` is invoked by the standard reconcile flow before `PrepareTLSConfig` and before Vault write/update logic. The unit tests should mirror the real context contract, not invent a new test-only abstraction:

- `"kubeClient"`: controller-runtime `client.Client`
- `"vaultClient"`: `*vault.Client`
- `"restConfig"`: `*rest.Config` for the token-reviewer JWT path
- `"vaultConnection"` is part of the production context, but it is not needed for Story 7.2 test coverage unless a test explicitly touches code that reads it

The helper should use the same string context keys as production code to catch mismatches.

### Type Coverage Plan

| Type | Primary branch to test | Target file | Assertion focus |
|---|---|---|---|
| `DatabaseSecretEngineConfig` | `Secret`, `VaultSecret`, `RandomSecret` (KV v1 + KV v2) | `api/v1alpha1/databasesecretengineconfig_test.go` | `retrievedUsername` / `retrievedPassword`, username precedence, no live Vault |
| `RabbitMQSecretEngineConfig` | Shared root-credential helper path | `api/v1alpha1/rabbitmqsecretengineconfig_test.go` | `retrievedUsername` / `retrievedPassword` |
| `LDAPAuthEngineConfig` | Bind credential resolution only | `api/v1alpha1/ldapauthengineconfig_test.go` | Bind DN / password retrieved fields; do not mix in TLS tests |
| `AzureAuthEngineConfig` | Default short-circuit and resolved credentials | `api/v1alpha1/azureauthengineconfig_test.go` | `retrievedClientID` / `retrievedClientPassword` |
| `AzureSecretEngineConfig` | Default short-circuit and resolved credentials | `api/v1alpha1/azuresecretengineconfig_test.go` | `retrievedClientID` / `retrievedClientPassword` |
| `GCPAuthEngineConfig` | Default short-circuit and resolved credentials | `api/v1alpha1/gcpauthengineconfig_test.go` | Resolved credential fields used by payload generation |
| `GitHubSecretEngineConfig` | SSH secret branch and Vault-secret branch | `api/v1alpha1/githubsecretengineconfig_test.go` | `retrievedSSHKey`, wrong secret type error |
| `KubernetesSecretEngineConfig` | Service-account-token secret branch and Vault-secret branch | `api/v1alpha1/kubernetessecretengineconfig_test.go` | `retrievedServiceAccountJWT`, wrong secret type error |
| `KubernetesAuthEngineConfig` | Nil token-reviewer path and, if feasible, token request branch | `api/v1alpha1/kubernetesauthengineconfig_test.go` | `retrievedTokenReviewerJWT`, no-op/error semantics |
| `KubernetesAuthEngineRole` | Explicit namespaces, selector lookup, empty selector fallback | `api/v1alpha1/kubernetesauthenginerole_test.go` | `namespaces` internal field and `__no_namespace__` behavior |
| `QuaySecretEngineConfig` | Shared root-credential helper path | `api/v1alpha1/quaysecretengineconfig_test.go` | `retrievedToken` |
| `Policy` | Placeholder fast path and `sys/auth` accessor replacement | `api/v1alpha1/policy_test.go` | Resolved `Spec.Policy` text, nil-secret behavior |
| `EntityAlias` | Accessor lookup + canonical ID lookup with prefilled `Status.ID` | `api/v1alpha1/entityalias_test.go` | `retrievedMountAccessor`, `retrievedCanonicalID`, `retrievedAliasID`, `retrievedName` |
| `GroupAlias` | Accessor lookup + canonical ID lookup with prefilled `Status.ID` | `api/v1alpha1/groupalias_test.go` | `retrievedMountAccessor`, `retrievedCanonicalID`, `retrievedAliasID`, `retrievedName` |
| `RandomSecret` | Inline policy generation and password-policy-name Vault generation | `api/v1alpha1/randomsecret_test.go` | `calculatedSecret` / generated password behavior |

### Shared Test Patterns to Reuse

- Stay in package `v1alpha1` so tests can directly inspect unexported internal fields populated by `PrepareInternalValues`
- Use standard Go `testing` tests, matching the existing `api/v1alpha1/*_test.go` style
- Use `fake.NewClientBuilder()` for Kubernetes objects instead of envtest or controller tests
- Reuse the tiny `httptest` Vault-client pattern already present in `api/v1alpha1/utils/vaultobject_test.go`; duplicating a small test helper in `api/v1alpha1/` is preferable to exporting test-only code from production packages
- For assertions, prefer direct internal-field checks or independently constructed expectations; do not build expectations by re-running the same implementation logic under test

### Type-Specific Guardrails

- `DatabaseSecretEngineConfig` is the best representative table-driven suite for the shared `RootCredentialConfig` pattern because it exercises:
  - Kubernetes `Secret`
  - Vault secret reads
  - `RandomSecret` dereference
  - KV v1 vs KV v2 payload shape
  - username override vs auto-populated username
- `RabbitMQSecretEngineConfig`, `LDAPAuthEngineConfig`, `AzureAuthEngineConfig`, `AzureSecretEngineConfig`, `GCPAuthEngineConfig`, and `QuaySecretEngineConfig` should each get focused tests that prove their branch-specific internal field is actually populated, but they do not each need to re-implement the full `DatabaseSecretEngineConfig` matrix
- `GitHubSecretEngineConfig` requires `corev1.SecretTypeSSHAuth` and `corev1.SSHAuthPrivateKey`; use the real secret type constants instead of hand-rolled strings
- `KubernetesSecretEngineConfig` requires `corev1.SecretTypeServiceAccountToken` and `corev1.ServiceAccountTokenKey`; include the wrong-secret-type branch because the code validates secret type before reading data
- `KubernetesAuthEngineConfig` resolves JWTs through `GetJWTTokenWithDuration`, which uses `restConfig` and the Kubernetes TokenRequest API; avoid envtest for this story
- `Policy` intentionally returns `nil` on Vault read error but returns an error when `sys/auth` unexpectedly returns `nil`; lock down current behavior rather than "fixing" it inside this story
- `EntityAlias` and `GroupAlias` both have write-side effects when `Status.ID == ""`; pre-populate `Status.ID` unless the test fully mocks alias creation and status update
- `RandomSecret` has both a no-Vault inline-generation path and a Vault-backed password-policy path; cover both because they exercise different code paths and different failure modes

### What Not to Do

- Do not introduce controller tests, envtest suites, or integration fixtures for this story
- Do not change production code only to make these unit tests easier unless the tests expose a real defect
- Do not widen scope into `PrepareTLSConfig`; LDAP TLS and other TLS-related logic belong elsewhere
- Do not couple expectations to `GetPayload()` or other helpers that merely repackage the same internal fields; assert the resolved internal state or independently constructed outputs instead

### Project Structure Notes

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `api/v1alpha1/prepareinternalvalues_test_helpers_test.go` | New | Shared fake context / fake Vault / fake client helpers for `PrepareInternalValues` tests |
| 2 | `api/v1alpha1/databasesecretengineconfig_test.go` | Modified | Representative table-driven `RootCredentialConfig` credential-resolution coverage |
| 3 | `api/v1alpha1/rabbitmqsecretengineconfig_test.go` | Modified | Shared root-credential branch coverage |
| 4 | `api/v1alpha1/ldapauthengineconfig_test.go` | Modified | Bind-credential resolution coverage |
| 5 | `api/v1alpha1/azureauthengineconfig_test.go` | Modified | Default/no-op + resolved-credential coverage |
| 6 | `api/v1alpha1/azuresecretengineconfig_test.go` | Modified | Default/no-op + resolved-credential coverage |
| 7 | `api/v1alpha1/gcpauthengineconfig_test.go` | Modified | Default/no-op + resolved-credential coverage |
| 8 | `api/v1alpha1/githubsecretengineconfig_test.go` | Modified | SSH secret + Vault-secret credential resolution |
| 9 | `api/v1alpha1/kubernetessecretengineconfig_test.go` | Modified | ServiceAccount token + Vault-secret credential resolution |
| 10 | `api/v1alpha1/kubernetesauthengineconfig_test.go` | Modified | Token-reviewer JWT path / no-op behavior |
| 11 | `api/v1alpha1/kubernetesauthenginerole_test.go` | Modified | Namespace selection / fallback coverage |
| 12 | `api/v1alpha1/quaysecretengineconfig_test.go` | Modified | Token resolution coverage |
| 13 | `api/v1alpha1/policy_test.go` | Modified | Accessor placeholder resolution |
| 14 | `api/v1alpha1/entityalias_test.go` | Modified | Accessor + canonical ID lookup coverage |
| 15 | `api/v1alpha1/groupalias_test.go` | Modified | Accessor + canonical ID lookup coverage |
| 16 | `api/v1alpha1/randomsecret_test.go` | Modified | Inline and Vault-backed password generation coverage |

### Previous Story Intelligence

- Story 7.1 is already `ready-for-dev` but not yet implemented, so there are no fresh implementation learnings to inherit from it
- Story 7.0 is also still `ready-for-dev`; it targets `controllers/` integration helpers and does not overlap with this `api/v1alpha1` unit-test work
- Story 1.5 is the most useful historical precedent for this story's test style:
  - same-package `api/v1alpha1` tests
  - direct inspection of unexported retrieved fields
  - table-driven `testing`-package structure
  - explicit coverage of custom behavior rather than only "happy path" payload checks
- Story 6.1 is the most useful precedent for the alias types because it already documented the asymmetric `PrepareInternalValues` lookup pattern for `GroupAlias`

### Git Intelligence (Recent Commits)

```text
9fc8b3c Bmad epic 6 (#321)
7ce3e42 Merge pull request #320 from raffaelespazzoli/bmad-epic-5
d64b2b1 Complete Epic 5 retrospective and close epic
e5e982c Add integration tests for KubernetesSecretEngineConfig and KubernetesSecretEngineRole (Story 5.3)
168e7e0 Fix RabbitMQ role vhosts assertion type mismatch
```

The recent history reinforces two useful patterns for Story 7.2:

- the `api/v1alpha1` unit-test suite is already the place where type-level behavior is locked down
- recent work on Kubernetes secret engine and RabbitMQ types means the relevant test files already exist and should be extended rather than replaced

### References

- [Source: _bmad-output/planning-artifacts/epics.md] — Epic 7 and Story 7.2 scope, acceptance criteria, and neighboring-story boundaries
- [Source: _bmad-output/project-context.md] — pinned versions, testing rules, context-value contract, and unit-test guidance
- [Source: controllers/commons.go] — `prepareContext()` keys (`kubeClient`, `restConfig`, `vaultConnection`, `vaultClient`)
- [Source: controllers/vaultresourcecontroller/vaultresourcereconciler.go] — `PrepareInternalValues` call order in reconcile flow
- [Source: api/v1alpha1/utils/vaultobject.go] — `VaultObject` interface contract
- [Source: api/v1alpha1/utils/vaultutils.go] — `ReadSecret` helper and Vault 404 handling
- [Source: api/v1alpha1/utils/vaultobject_test.go] — existing `httptest` Vault client pattern
- [Source: api/v1alpha1/databasesecretengineconfig_types.go] — representative `RootCredentialConfig` credential resolution, KV v2 branch, username precedence
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go] — shared root-credential resolution plus retrieved username/password fields
- [Source: api/v1alpha1/ldapauthengineconfig_types.go] — bind credential resolution and retrieved bind fields
- [Source: api/v1alpha1/azureauthengineconfig_types.go] — resolved Azure auth credentials and default short-circuit
- [Source: api/v1alpha1/azuresecretengineconfig_types.go] — resolved Azure secret-engine credentials and default short-circuit
- [Source: api/v1alpha1/gcpauthengineconfig_types.go] — resolved GCP auth credentials and default short-circuit
- [Source: api/v1alpha1/githubsecretengineconfig_types.go] — SSH secret / Vault-secret credential resolution
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go] — service-account-token secret / Vault-secret credential resolution
- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go] — token-reviewer JWT retrieval via `GetJWTTokenWithDuration`
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go] — namespace selection and `__no_namespace__` fallback
- [Source: api/v1alpha1/quaysecretengineconfig_types.go] — token retrieval and `retrievedToken`
- [Source: api/v1alpha1/policy_types.go] — placeholder detection and `sys/auth` accessor replacement
- [Source: api/v1alpha1/entityalias_types.go] — accessor / canonical-ID lookup and alias side effects
- [Source: api/v1alpha1/groupalias_types.go] — accessor / canonical-ID lookup and alias side effects
- [Source: api/v1alpha1/randomsecret_types.go] — inline password generation and password-policy-name generation path
- [Source: _bmad-output/implementation-artifacts/1-5-unit-tests-tomap-and-isequivalenttodesiredstate-secret-engine-config-role-types.md] — same-package unit-test conventions for existing config-type coverage
- [Source: _bmad-output/implementation-artifacts/6-1-integration-tests-for-group-and-groupalias-types.md] — alias lookup behavior and `Status.ID` guidance
- [Source: _bmad-output/implementation-artifacts/7-1-webhook-validation-tests-for-immutable-spec-path-rule.md] — current story-document structure and repo-local conventions

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
