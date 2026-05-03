# Story 7.0: Refactor RandomSecret Test to Use Shared Fixtures

Status: backlog

## Story

As an operator developer,
I want the RandomSecret integration tests to share prerequisite resources across Context blocks using Ginkgo BeforeAll/AfterAll,
So that redundant setup/teardown is eliminated and the integration test suite runs significantly faster.

## Background / Motivation

Performance profiling of the integration test suite (Epic 6 retrospective, 2026-05-02) identified RandomSecret and VaultSecret tests as the biggest time outliers — consuming ~31% of total test time despite not integrating with any external tools (no PostgreSQL, RabbitMQ, LDAP, etc.).

Root cause analysis revealed that the slowness is **structural**: each of the 4 `Context` blocks in `randomsecret_controller_test.go` independently creates and tears down the same 6 prerequisite resources (PasswordPolicy, 2 Policies, 2 KubernetesAuthEngineRoles, SecretEngineMount), each requiring its own `Eventually` polling loop (2-second interval, typically 2-6 seconds per poll). This results in ~12 redundant `Eventually` round-trips × 3 extra copies = ~100-150 seconds of wasted time.

### Current test structure (problematic)

```
Describe("Random Secret controller for v2 secrets")
  Context("retain policy")           → creates 6-8 prereqs, tests, tears down all
  Context("multi-key")               → creates 6 prereqs (same), tests, tears down all
  Context("multi-key-recreate")      → creates 6 prereqs (same), tests, tears down all
  Context("refreshPeriod")           → creates 6 prereqs (same), tests, tears down all
```

### Target test structure

```
Describe("Random Secret controller for v2 secrets")
  BeforeAll → creates shared prereqs once
  AfterAll  → tears down shared prereqs once
  Context("retain policy")           → creates only test-specific resources
  Context("multi-key")               → creates only test-specific resources
  Context("multi-key-recreate")      → creates only test-specific resources
  Context("refreshPeriod")           → creates only test-specific resources
```

## Scope

| File | Change | Impact |
|------|--------|--------|
| `controllers/randomsecret_controller_test.go` (1717 lines) | Major refactor — lift shared prereqs to BeforeAll/AfterAll | Primary target |
| `controllers/vaultsecret_controller_test.go` (679 lines) | No change — single Context, nothing to deduplicate | Out of scope |
| `controllers/vaultsecret_controller_v2_test.go` (491 lines) | No change — single Context, nothing to deduplicate | Out of scope |

Cross-file fixture sharing (between RandomSecret and VaultSecret v2 tests) is explicitly out of scope — it would require deeper structural changes for marginal gain.

## Acceptance Criteria

1. **Given** the refactored `randomsecret_controller_test.go` **When** the full integration test suite runs **Then** all existing RandomSecret tests pass with identical assertions and behavior.

2. **Given** shared prerequisites (PasswordPolicy, Policies, KubernetesAuthEngineRoles, SecretEngineMount) **When** the Describe block starts **Then** they are created exactly once in BeforeAll and torn down exactly once in AfterAll.

3. **Given** Context 1 (retain) needs extra resources (secret-reader policy + role for VaultSecret reading) **When** that Context runs **Then** the extra resources are created/destroyed within that Context's BeforeAll/AfterAll or inline, not in the shared set.

4. **Given** all 4 Context blocks **When** they execute sequentially **Then** there is no cross-Context state leakage — each Context's test-specific resources (RandomSecret, VaultSecret) are fully self-contained.

5. **Given** the refactored suite **When** compared to the pre-refactor run **Then** the RandomSecret test group duration decreases by at least 40%.

## Tasks / Subtasks

- [ ] Task 1: Identify shared prerequisites
  - [ ] 1.1: Catalog which YAML fixtures are identical across all 4 Context blocks
  - [ ] 1.2: Identify Context-1-only extras (secret-reader policy + KubernetesAuthEngineRole)

- [ ] Task 2: Refactor to BeforeAll/AfterAll
  - [ ] 2.1: Add `BeforeAll` at `Describe` level — create shared PasswordPolicy, Policies, KubernetesAuthEngineRoles, SecretEngineMount with their `Eventually` blocks
  - [ ] 2.2: Add `AfterAll` at `Describe` level — delete shared resources with their `Eventually` blocks
  - [ ] 2.3: Remove duplicated setup/teardown from each of the 4 `It` blocks
  - [ ] 2.4: For Context 1, keep the extra secret-reader resources scoped to that Context

- [ ] Task 3: Verify isolation
  - [ ] 3.1: Ensure each Context's test-specific resources (RandomSecret instances, VaultSecret instances) use unique names and don't collide
  - [ ] 3.2: Verify the SecretEngineMount and its KV path remain available across all Contexts

- [ ] Task 4: End-to-end verification
  - [ ] 4.1: Run `make integration` and verify all tests pass (RandomSecret + all others)
  - [ ] 4.2: Compare before/after timing from Ginkgo JSON report to confirm improvement

## Dev Notes

### Ginkgo BeforeAll/AfterAll semantics

Ginkgo v2 `BeforeAll` runs once before the first spec in a container (Describe/Context). If it fails, all specs in that container are skipped. This is the desired behavior — if shared infrastructure can't be created, there's no point running any of the tests.

Variables created in `BeforeAll` must be declared at the container scope (before the `BeforeAll` block) so they're accessible in the `It` blocks.

### The refreshPeriod test's ~20-second wait is inherent

The `refreshPeriod` Context (Context 4) uses `11-randomsecret-refresh-v2.yaml` with `refreshPeriod: 15s`, plus a 5-second `Consistently` check. This ~20-second wait is testing real timer behavior and cannot be optimized away. The savings from this refactoring come from eliminating the ~4× duplicated infrastructure setup, not from changing test logic.

### Shared fixture catalog

| Resource | YAML path | Used by |
|----------|-----------|---------|
| PasswordPolicy | `test/randomsecret/v2/00-passwordpolicy-simple-password-policy-v2.yaml` | All 4 Contexts |
| Policy (kv-engine-admin) | `test/randomsecret/v2/01-policy-kv-engine-admin-v2.yaml` | All 4 Contexts |
| Policy (secret-writer) | `test/randomsecret/v2/04-policy-secret-writer-v2.yaml` | All 4 Contexts |
| Policy (secret-reader) | `test/vaultsecret/v2/00-policy-secret-reader-v2.yaml` | Context 1 only |
| KubernetesAuthEngineRole (kv-engine-admin) | `test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml` | All 4 Contexts |
| KubernetesAuthEngineRole (secret-writer) | `test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml` | All 4 Contexts |
| KubernetesAuthEngineRole (secret-reader) | `test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml` | Context 1 only |
| SecretEngineMount | `test/randomsecret/v2/03-secretenginemount-kv-v2.yaml` | All 4 Contexts |

### Estimated time savings

- Current: ~4 × (6 prereq creates + 6 prereq deletes) × ~3-4s each = ~150-200s in prereq overhead
- After: 1 × (6 prereq creates + 6 prereq deletes) × ~3-4s each = ~40-50s in prereq overhead
- **Net savings: ~100-150 seconds** (~40-60% of total RandomSecret test time)
