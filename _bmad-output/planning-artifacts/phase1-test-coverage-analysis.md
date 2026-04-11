# Phase 1: Test Coverage Analysis — Stabilization Report

**Project:** vault-config-operator
**Date:** 2026-04-11
**Status:** Analysis complete — prioritized test backlog produced

---

## Executive Summary

The vault-config-operator has **47 CRD types** with controllers, but test coverage is critically low:

- **Overall unit test code coverage: 2.2%**
- **Integration tests exist for only 7 of 47 types (15%)**
- **0 of 47 types have Update scenario tests**
- **0 integration tests exercise `IsEquivalentToDesiredState`**
- **0 integration tests cover error/failure paths**

---

## 1. Unit Test Coverage (per-package)

| Package | Coverage | Notes |
|---------|----------|-------|
| `controllers/vaultsecretutils` | **100.0%** | `HashData`, `HashMeta`, `GetResourceVersion` — fully tested |
| `controllers/vaultresourcecontroller` | **3.9%** | Only `PeriodicReconcilePredicate.Update()` at 93.3%; all other functions 0% |
| `api/v1alpha1` | **2.7%** | Only `init()` functions (scheme registration) hit; all VaultObject methods 0% |
| `api/v1alpha1/utils` | **0.0%** | `VaultEndpoint`, `VaultConnection`, all helpers — zero coverage |
| `controllers` | **0.0%** | All 47 reconcilers — zero unit test coverage |
| `controllers/controllertestutils` | **0.0%** | Decoder utility — only used by integration tests |
| Root (`main.go`) | **0.0%** | Not unit-testable in current form |

**Function-level stats:** 105 of 1,914 functions have any coverage (5.5%). 1,809 functions at 0%.

### API Types Unit Tests (existing)

| Test File | Types Covered | What's Tested |
|-----------|--------------|---------------|
| `identitytoken_test.go` | IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole | `IsEquivalentToDesiredState`, `toMap` |
| `identityoidc_test.go` | IdentityOIDCScope, IdentityOIDCProvider, IdentityOIDCClient, IdentityOIDCAssignment | `IsEquivalentToDesiredState`, `toMap` |
| `secretenginemount_test.go` | SecretEngineMount | `GetPath` only |

**That means 40 of 47 types have zero unit tests for their VaultObject methods.**

---

## 2. Integration Test Coverage (type matrix)

### Types WITH integration tests (7 of 47)

| Type | Test File | Create | Update | Delete | Error Paths | `IsEquivalentToDesiredState` |
|------|-----------|--------|--------|--------|-------------|------------------------------|
| VaultSecret | `vaultsecret_controller_test.go` | Yes | No | Yes | No | No |
| VaultSecret (v2) | `vaultsecret_controller_v2_test.go` | Yes | No | Yes | No | No |
| RandomSecret | `randomsecret_controller_test.go` | Yes | No | Yes | No | No |
| Entity | `entity_controller_test.go` | Yes | No | Yes | No | No |
| EntityAlias | `entityalias_controller_test.go` | Yes | No | Yes | No | No |
| PKISecretEngineConfig + Role | `pkisecretengine_controller_test.go` | Yes | No | Yes | No | No |
| DatabaseSecretEngineStaticRole* | `databasesecretenginestaticrole_controller_test.go` | Yes | No | Yes | No | No |

\* Note: the StaticRole test file actually tests `DatabaseSecretEngineConfig` setup flow, not the static role reconciliation directly.

### Types WITHOUT any integration tests (40 of 47)

**Auth Engines (12 types — 0 tested):**
- AuthEngineMount, KubernetesAuthEngineConfig, KubernetesAuthEngineRole, LDAPAuthEngineConfig, LDAPAuthEngineGroup, JWTOIDCAuthEngineConfig, JWTOIDCAuthEngineRole, AzureAuthEngineConfig, AzureAuthEngineRole, GCPAuthEngineConfig, GCPAuthEngineRole, CertAuthEngineConfig, CertAuthEngineRole

**Secret Engines (14 types — 0 tested):**
- SecretEngineMount, DatabaseSecretEngineConfig, DatabaseSecretEngineRole, RabbitMQSecretEngineConfig, RabbitMQSecretEngineRole, GitHubSecretEngineConfig, GitHubSecretEngineRole, AzureSecretEngineConfig, AzureSecretEngineRole, QuaySecretEngineConfig, QuaySecretEngineRole, QuaySecretEngineStaticRole, KubernetesSecretEngineConfig, KubernetesSecretEngineRole

**Policies & Passwords (2 types — 0 tested):**
- Policy, PasswordPolicy

**Identity (11 types — 0 tested):**
- Group, GroupAlias, IdentityOIDCProvider, IdentityOIDCScope, IdentityOIDCClient, IdentityOIDCAssignment, IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole

**Audit (2 types — 0 tested):**
- Audit, AuditRequestHeader

---

## 3. Scenario Coverage Gaps (across ALL integration tests)

| Scenario | Covered? | Risk |
|----------|----------|------|
| Create resource → reconcile success | Yes (7 types) | Low for tested types |
| Delete resource → Vault cleanup | Yes (7 types) | Low for tested types |
| Update resource → `IsEquivalentToDesiredState` → conditional write | **NO** | **CRITICAL** — the core declarative reconciliation logic is untested end-to-end |
| Error paths (bad auth, Vault unreachable, invalid config) | **NO** | **HIGH** — no negative test scenarios exist |
| Webhook validation (immutable path, invalid fields) | **NO** | **HIGH** — 43 webhooks, zero validation tests |
| `PrepareInternalValues` with credentials (K8s secrets, RandomSecrets, VaultSecrets) | **NO** | **HIGH** — credential resolution logic untested |
| Drift detection (`ENABLE_DRIFT_DETECTION=true`) | **NO** | **MEDIUM** — newer feature, not yet tested |

---

## 4. Webhook Coverage

- **43 types** have webhooks; **4 types** lack them (Audit, AuditRequestHeader, Entity, EntityAlias)
- **0 webhook-specific test files** exist (no `*_webhook_test.go`)
- The `webhook_suite_test.go` is bootstrap only (no actual test cases)
- Critical webhook rules (immutable `spec.path`) are completely untested

---

## 5. `IsEquivalentToDesiredState` Coverage

This is the most critical function per type — the bridge from Vault's imperative API to the operator's declarative model.

| Coverage Level | Types | Count |
|----------------|-------|-------|
| **Unit tested** | IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole, IdentityOIDCScope, IdentityOIDCProvider, IdentityOIDCClient, IdentityOIDCAssignment | **7** |
| **Untested** (simple `reflect.DeepEqual`) | Most standard types (roles, simple configs) | **~25** |
| **Untested** (complex — field remapping/filtering) | DatabaseSecretEngineConfig, AuthEngineMount, SecretEngineMount, Policy, RabbitMQSecretEngineConfig, PKISecretEngineConfig | **~6** |
| **Not applicable** | VaultSecret (uses different sync model) | **1** |

**Highest risk:** The 6 types with complex `IsEquivalentToDesiredState` logic (field remapping, filtering) have zero test coverage.

---

## 6. Types with Non-Trivial `PrepareInternalValues` (credential resolution)

These types resolve credentials from Kubernetes Secrets, RandomSecrets, or Vault at reconcile time — all untested:

| Type | What `PrepareInternalValues` does |
|------|-----------------------------------|
| DatabaseSecretEngineConfig | Resolves root credentials from K8s Secret / RandomSecret / VaultSecret |
| RabbitMQSecretEngineConfig | Same credential resolution pattern |
| LDAPAuthEngineConfig | Credential resolution + TLS config |
| AzureAuthEngineConfig | Optional credential resolution |
| AzureSecretEngineConfig | Optional credential resolution |
| GCPAuthEngineConfig | Optional credential resolution |
| GitHubSecretEngineConfig | Credential resolution + TLS |
| KubernetesSecretEngineConfig | Credential resolution |
| KubernetesAuthEngineConfig | JWT token generation from ServiceAccount |
| KubernetesAuthEngineRole | Namespace selector resolution |
| QuaySecretEngineConfig | Credential resolution |
| Policy | Auth engine accessor placeholder expansion (`${auth/.../@accessor}`) |
| EntityAlias | Auth accessor lookup via `sys/auth` |
| GroupAlias | Auth accessor lookup via `sys/auth` |
| RandomSecret | Password generation |

---

## 7. Test Fixtures Inventory

**93 YAML fixtures** under `test/` covering many types. However, most are only used by the existing 7 integration tests or are samples for documentation. Many types have fixtures available but no test that uses them.

**Types with fixtures but NO tests:**
- Database engine (full set), RabbitMQ, GitHub, JWT/OIDC, LDAP, Kubernetes secret engine, Groups, Identity OIDC, Identity Token, Kubernetes auth, Policies, Password policies

---

## 8. Prioritized Test Backlog

### Priority 1 — CRITICAL (stabilization foundation)

| # | Action | Reason |
|---|--------|--------|
| 1.1 | **Unit tests for `IsEquivalentToDesiredState` on all 46 types** | Core declarative logic — if this is wrong, the operator either writes unnecessarily or misses drift. Start with the 6 complex types (DB config, engine mounts, Policy, RabbitMQ, PKI). |
| 1.2 | **Unit tests for `toMap()` on all types** | Ensures CRD fields map correctly to Vault API snake_case keys. |
| 1.3 | **Integration tests: Update scenario** for at least the 7 currently-tested types | Currently zero Update tests anywhere. Add: change a field → verify Vault state changes. |

### Priority 2 — HIGH (broad type coverage)

| # | Action | Reason |
|---|--------|--------|
| 2.1 | **Integration tests for Policy and PasswordPolicy** | Core building blocks used by many other types. |
| 2.2 | **Integration tests for SecretEngineMount and AuthEngineMount** | Engine mount types are prerequisites for all engine-specific types. |
| 2.3 | **Integration tests for DatabaseSecretEngineConfig and DatabaseSecretEngineRole** | Most complex type with credential resolution, root password rotation. |
| 2.4 | **Integration tests for KubernetesAuthEngineConfig and KubernetesAuthEngineRole** | Most commonly used auth method. Fixtures exist. |

### Priority 3 — MEDIUM (remaining engine types)

| # | Action | Reason |
|---|--------|--------|
| 3.1 | Integration tests for LDAP auth (Config + Group) | Fixtures exist. Non-trivial `PrepareInternalValues`. |
| 3.2 | Integration tests for JWT/OIDC auth (Config + Role) | Fixtures exist. |
| 3.3 | Integration tests for RabbitMQ (Config + Role) | Complex `IsEquivalentToDesiredState` + credential resolution. |
| 3.4 | Integration tests for remaining secret engines (GitHub, Azure, GCP, Quay, Kubernetes) | Expand coverage to full engine portfolio. |
| 3.5 | Integration tests for Identity types (Group, GroupAlias, OIDC, Token) | Fixtures exist for all. |
| 3.6 | Integration tests for Audit types | Simple types but currently zero coverage. |

### Priority 4 — LOWER (hardening)

| # | Action | Reason |
|---|--------|--------|
| 4.1 | **Webhook validation tests** | 43 webhooks with zero test coverage. Especially immutable `spec.path` rule. |
| 4.2 | **Error path tests** | Vault unreachable, bad credentials, invalid config — none tested today. |
| 4.3 | **`PrepareInternalValues` unit tests** for credential resolution types | 15 types with non-trivial logic, all untested. |
| 4.4 | **Drift detection integration tests** (`ENABLE_DRIFT_DETECTION=true`) | Newer feature with zero coverage. |

---

## 9. Decoder Gap

The `controllertestutils/decoder.go` currently supports only: PasswordPolicy, VaultSecret, Policy, KubernetesAuthEngineRole, SecretEngineMount, RandomSecret, PKISecretEngineConfig, PKISecretEngineRole, DatabaseSecretEngineConfig, DatabaseSecretEngineStaticRole, Entity, EntityAlias.

**Missing decoder methods needed for Priority 2+:** AuthEngineMount, DatabaseSecretEngineRole, LDAPAuthEngineConfig, LDAPAuthEngineGroup, JWTOIDCAuthEngineConfig, JWTOIDCAuthEngineRole, KubernetesAuthEngineConfig, RabbitMQSecretEngineConfig, RabbitMQSecretEngineRole, GitHubSecretEngineConfig, GitHubSecretEngineRole, AzureAuthEngineConfig, AzureAuthEngineRole, AzureSecretEngineConfig, AzureSecretEngineRole, GCPAuthEngineConfig, GCPAuthEngineRole, CertAuthEngineConfig, CertAuthEngineRole, QuaySecretEngineConfig, QuaySecretEngineRole, QuaySecretEngineStaticRole, KubernetesSecretEngineConfig, KubernetesSecretEngineRole, Group, GroupAlias, Audit, AuditRequestHeader, IdentityOIDCProvider, IdentityOIDCScope, IdentityOIDCClient, IdentityOIDCAssignment, IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole.
