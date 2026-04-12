---
stepsCompleted: [1, 2]
inputDocuments:
  - _bmad-output/planning-artifacts/phase1-test-coverage-analysis.md
  - _bmad-output/planning-artifacts/phase2-expansion-analysis.md
  - _bmad-output/project-context.md
---

# vault-config-operator - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for the vault-config-operator test stabilization phase, decomposing the requirements from the Phase 1 Test Coverage Analysis into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: Unit tests for `IsEquivalentToDesiredState` on all 46 applicable CRD types, verifying correct declarative state comparison against Vault API responses
FR2: Unit tests for `toMap()` on all 46 applicable CRD types, verifying correct CRD field-to-Vault-API-key mapping (camelCase → snake_case)
FR3: Integration tests with Update scenarios (change spec field → reconcile → verify Vault state updated) for the 7 types that already have Create/Delete integration tests
FR4: Integration tests for Policy and PasswordPolicy types (create, reconcile success, delete with Vault cleanup)
FR5: Integration tests for SecretEngineMount and AuthEngineMount types (create, reconcile success, delete with Vault cleanup)
FR6: Integration tests for DatabaseSecretEngineConfig and DatabaseSecretEngineRole (create, reconcile success, delete, credential resolution)
FR7: Integration tests for KubernetesAuthEngineConfig and KubernetesAuthEngineRole (create, reconcile success, delete)
FR8: Integration tests for LDAP auth engine types — LDAPAuthEngineConfig and LDAPAuthEngineGroup
FR9: Integration tests for JWT/OIDC auth engine types — JWTOIDCAuthEngineConfig and JWTOIDCAuthEngineRole
FR10: Integration tests for RabbitMQ secret engine types — RabbitMQSecretEngineConfig and RabbitMQSecretEngineRole
FR11: Integration tests for remaining secret engines — GitHubSecretEngineConfig/Role, AzureSecretEngineConfig/Role, QuaySecretEngineConfig/Role/StaticRole, KubernetesSecretEngineConfig/Role
FR12: Integration tests for Identity types — Group, GroupAlias, IdentityOIDCProvider, IdentityOIDCScope, IdentityOIDCClient, IdentityOIDCAssignment, IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole
FR13: Integration tests for Audit types — Audit and AuditRequestHeader
FR14: Webhook validation tests covering the immutable `spec.path` rule, field validation, and defaulting behavior
FR15: Error path integration tests — bad authentication, Vault unreachable, invalid configuration
FR16: Unit tests for `PrepareInternalValues` on the 15 types with non-trivial credential resolution logic
FR17: Integration tests for drift detection (`ENABLE_DRIFT_DETECTION=true`) verifying periodic reconciliation corrects Vault drift

### Non-Functional Requirements

NFR1: All existing tests must continue to pass — zero regression
NFR2: Integration tests must follow existing patterns (Ginkgo v2, `Eventually` polling, controllertestutils decoder)
NFR3: New decoder methods must be added to `controllertestutils/decoder.go` for each newly tested type
NFR4: Test YAML fixtures in `test/` directory for each new integration test
NFR5: Unit tests must run without external dependencies (no Vault, no cluster)
NFR6: Build tags: unit tests `//go:build !integration`, integration tests `//go:build integration`

### Additional Requirements

- Integration test suite (`suite_integration_test.go`) must register new controllers for any newly tested types
- Test namespaces `vault-admin` and `test-vault-config-operator` are shared across all integration tests
- Vault integration tests require a running Vault instance (`VAULT_ADDR`, `VAULT_TOKEN` env vars)
- Types that require external services for integration testing (RabbitMQ, LDAP, databases, cloud providers) may need mock services or test infrastructure additions

### FR Coverage Map

FR1: Epic 1 — `IsEquivalentToDesiredState` unit tests for all 46 types
FR2: Epic 1 — `toMap()` unit tests for all 46 types
FR3: Epic 2 — Update scenario integration tests for 7 currently-tested types
FR4: Epic 3 — Policy + PasswordPolicy integration tests
FR5: Epic 3 — SecretEngineMount + AuthEngineMount integration tests
FR6: Epic 5 — DatabaseSecretEngineConfig + Role integration tests
FR7: Epic 4 — KubernetesAuthEngineConfig + Role integration tests
FR8: Epic 4 — LDAPAuthEngineConfig + Group integration tests
FR9: Epic 4 — JWTOIDCAuthEngineConfig + Role integration tests
FR10: Epic 5 — RabbitMQSecretEngineConfig + Role integration tests
FR11: Epic 5 — Remaining secret engine integration tests
FR12: Epic 6 — Identity type integration tests
FR13: Epic 6 — Audit type integration tests
FR14: Epic 7 — Webhook validation tests
FR15: Epic 7 — Error path integration tests
FR16: Epic 7 — PrepareInternalValues unit tests
FR17: Epic 7 — Drift detection integration tests

## Epic List

### Epic 1: Core Declarative Logic Unit Tests
Verify the correctness of the imperative-to-declarative bridge (`toMap` + `IsEquivalentToDesiredState`) for every CRD type — the most critical untested logic in the operator.
**FRs covered:** FR1, FR2

### Epic 2: Integration Test Update Scenarios
Add Update (spec change → reconcile → Vault write) scenarios to the 7 types that already have Create/Delete integration tests, closing the biggest scenario gap.
**FRs covered:** FR3

### Epic 3: Foundation Type Integration Tests
Integration tests for the building-block types that other types depend on — Policy, PasswordPolicy, SecretEngineMount, AuthEngineMount.
**FRs covered:** FR4, FR5

### Epic 4: Auth Engine Integration Tests
Full integration test coverage for all auth engine types — Kubernetes, LDAP, JWT/OIDC.
**FRs covered:** FR7, FR8, FR9

### Epic 5: Secret Engine Integration Tests
Full integration test coverage for all secret engine types — Database, RabbitMQ, GitHub, Azure, GCP, Quay, Kubernetes.
**FRs covered:** FR6, FR10, FR11

### Epic 6: Identity & Audit Integration Tests
Integration test coverage for identity management types (Group, GroupAlias, OIDC, Token) and audit types.
**FRs covered:** FR12, FR13

### Epic 7: Hardening — Webhooks, Error Paths & Credential Resolution
Webhook validation tests, error/failure scenario tests, `PrepareInternalValues` unit tests, and drift detection tests.
**FRs covered:** FR14, FR15, FR16, FR17

---

## Epic 1: Core Declarative Logic Unit Tests

Verify the correctness of the imperative-to-declarative bridge (`toMap` + `IsEquivalentToDesiredState`) for every CRD type — the most critical untested logic in the operator. These unit tests run without Vault or a cluster and catch field mapping bugs and comparison logic errors.

### Story 1.1: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Simple Standard Types

As an operator developer,
I want unit tests verifying `toMap()` produces correct snake_case Vault API payloads and `IsEquivalentToDesiredState` correctly compares desired vs actual state,
So that I can confidently modify CRD fields without breaking the declarative reconciliation logic.

**Types covered:** IdentityOIDCScope, IdentityOIDCProvider, IdentityOIDCClient, IdentityOIDCAssignment, IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole (extend existing tests), PasswordPolicy, AuditRequestHeader

**Acceptance Criteria:**

**Given** a CRD instance with all fields populated
**When** `toMap()` is called on the inline config struct
**Then** the returned map keys are snake_case and match the Vault API field names exactly

**Given** a CRD instance and a Vault read response payload that matches the desired state
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it returns `true`

**Given** a CRD instance and a Vault read response payload with a different value for one managed field
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it returns `false`

**Given** a CRD instance and a Vault read response payload containing extra fields not managed by the operator
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it returns `true` (extra fields are ignored)

### Story 1.2: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Engine Mount Types

As an operator developer,
I want unit tests for the engine mount types where `IsEquivalentToDesiredState` compares only the tune config (not the full mount spec),
So that the unique comparison semantics of mount types are verified.

**Types covered:** AuthEngineMount, SecretEngineMount

**Acceptance Criteria:**

**Given** an AuthEngineMount instance with Config fields populated
**When** `Config.toMap()` is called
**Then** the returned map contains `default_lease_ttl`, `max_lease_ttl`, `listing_visibility`, and all other tune fields

**Given** an AuthEngineMount instance and a Vault tune response payload
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it compares only `Config.toMap()` against the payload (not the full mount spec)

**Given** a SecretEngineMount instance with the same pattern
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** same tune-only comparison behavior is verified

### Story 1.3: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Database Engine Types (Complex)

As an operator developer,
I want unit tests for DatabaseSecretEngineConfig where Vault restructures fields in its read response (moving fields into `connection_details`),
So that the most complex `IsEquivalentToDesiredState` implementation is verified.

**Types covered:** DatabaseSecretEngineConfig, DatabaseSecretEngineRole, DatabaseSecretEngineStaticRole

**Acceptance Criteria:**

**Given** a DatabaseSecretEngineConfig instance with `connectionURL`, `username`, `disableEscaping` populated
**When** `IsEquivalentToDesiredState(payload)` is called with a Vault read response where those fields are nested under `connection_details`
**Then** it returns `true` (correctly remaps fields to match Vault's structure)

**Given** a DatabaseSecretEngineConfig with `AllowedRoles` as `[]string`
**When** compared against a Vault response with `allowed_roles` as `[]interface{}`
**Then** it returns `true` (handles Go type differences)

**Given** a DatabaseSecretEngineConfig where `RootPasswordRotation.Enable` is true and `Status.LastRootPasswordRotation` is zero
**When** `IsEquivalentToDesiredState` is called
**Then** it returns `false` (forces initial rotation)

### Story 1.4: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Auth Engine Config Types

As an operator developer,
I want unit tests for all auth engine configuration and role types,
So that the field mappings for Kubernetes, LDAP, JWT/OIDC, Azure, GCP, and Cert auth are verified.

**Types covered:** KubernetesAuthEngineConfig, KubernetesAuthEngineRole, LDAPAuthEngineConfig, LDAPAuthEngineGroup, JWTOIDCAuthEngineConfig, JWTOIDCAuthEngineRole, AzureAuthEngineConfig, AzureAuthEngineRole, GCPAuthEngineConfig, GCPAuthEngineRole, CertAuthEngineConfig, CertAuthEngineRole

**Acceptance Criteria:**

**Given** each auth engine type instance with representative field values
**When** `toMap()` is called
**Then** all fields map correctly to Vault API snake_case keys

**Given** each auth engine type and a matching Vault read response
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it returns `true`

**Given** each auth engine type and a Vault response with one managed field changed
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it returns `false`

### Story 1.5: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Secret Engine Config & Role Types

As an operator developer,
I want unit tests for all secret engine configuration and role types,
So that field mappings for RabbitMQ, GitHub, Azure, PKI, Quay, and Kubernetes secret engines are verified.

**Types covered:** RabbitMQSecretEngineConfig, RabbitMQSecretEngineRole, GitHubSecretEngineConfig, GitHubSecretEngineRole, AzureSecretEngineConfig, AzureSecretEngineRole, PKISecretEngineConfig, PKISecretEngineRole, QuaySecretEngineConfig, QuaySecretEngineRole, QuaySecretEngineStaticRole, KubernetesSecretEngineConfig, KubernetesSecretEngineRole

**Acceptance Criteria:**

**Given** each secret engine type instance with representative field values
**When** `toMap()` is called
**Then** all fields map correctly to Vault API snake_case keys

**Given** each type and a matching Vault read response
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it returns `true`

**Given** each type and a response with one managed field changed
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it returns `false`

### Story 1.6: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Remaining Types

As an operator developer,
I want unit tests for Policy, RandomSecret, VaultSecret, Audit, Group, GroupAlias, Entity, and EntityAlias,
So that the full type portfolio has declarative logic coverage.

**Types covered:** Policy, RandomSecret, Audit, Group, GroupAlias, Entity, EntityAlias

**Acceptance Criteria:**

**Given** a Policy instance with `${auth/kubernetes/@accessor}` placeholder in the policy text
**When** `toMap()` is called (before `PrepareInternalValues` resolution)
**Then** the policy text is included as-is in the payload

**Given** a Policy instance and a Vault read response where `name` and `rules`/`policy` keys differ based on `Spec.Type`
**When** `IsEquivalentToDesiredState(payload)` is called
**Then** it correctly handles the conditional field remapping

**Given** each remaining type with representative field values
**When** `toMap()` and `IsEquivalentToDesiredState` are exercised
**Then** correct behavior is verified

---

## Epic 2: Integration Test Update Scenarios

Add Update scenarios to the 7 types that already have Create/Delete integration tests. This closes the critical gap where the `IsEquivalentToDesiredState` → conditional write flow has never been exercised end-to-end.

### Story 2.1: Add Update scenarios to VaultSecret integration tests

As an operator developer,
I want integration tests that modify a VaultSecret spec field and verify the reconciler updates the Kubernetes Secret accordingly,
So that the VaultSecret update path is validated end-to-end.

**Acceptance Criteria:**

**Given** a VaultSecret that has been successfully reconciled (ReconcileSuccessful=True)
**When** I update a field in the VaultSecret spec (e.g., change a template or secret path)
**Then** the reconciler detects the change and updates the generated Kubernetes Secret
**And** the ReconcileSuccessful condition is updated with a new ObservedGeneration

### Story 2.2: Add Update scenarios to RandomSecret integration tests

As an operator developer,
I want integration tests that modify a RandomSecret spec and verify the reconciler updates the Vault secret,
So that the RandomSecret update path is validated.

**Acceptance Criteria:**

**Given** a RandomSecret that has been successfully reconciled
**When** I update the RandomSecret spec (e.g., change the password policy or secret key)
**Then** the reconciler detects the generation change and writes the updated secret to Vault

### Story 2.3: Add Update scenarios to Entity and EntityAlias integration tests

As an operator developer,
I want integration tests that modify Entity/EntityAlias specs and verify reconciliation,
So that identity update paths are validated.

**Acceptance Criteria:**

**Given** an Entity that has been successfully reconciled
**When** I update the Entity spec (e.g., change metadata or policies)
**Then** the reconciler detects the change, `IsEquivalentToDesiredState` returns false, and the Entity is updated in Vault

**Given** an EntityAlias that has been successfully reconciled
**When** I update the EntityAlias spec
**Then** the reconciler updates the alias in Vault

### Story 2.4: Add Update scenarios to PKI and DatabaseSecretEngineStaticRole integration tests

As an operator developer,
I want integration tests that modify PKI config/role and DB static role specs,
So that these update paths are validated.

**Acceptance Criteria:**

**Given** a PKISecretEngineConfig that has been successfully reconciled
**When** I update a config field (e.g., max_lease_ttl)
**Then** the reconciler updates the PKI config in Vault

**Given** a PKISecretEngineRole that has been successfully reconciled
**When** I update a role field (e.g., allowed_domains)
**Then** the reconciler updates the role in Vault

---

## Epic 3: Foundation Type Integration Tests

Integration tests for the building-block types that other types depend on. These must be stable before testing engine-specific types.

### Story 3.1: Integration tests for Policy type

As an operator developer,
I want integration tests for the Policy type covering create, reconcile success, and delete with Vault cleanup,
So that the most fundamental Vault resource type has end-to-end test coverage.

**Acceptance Criteria:**

**Given** a Policy CR is created in the test namespace with a valid HCL policy
**When** the reconciler processes it
**Then** the policy exists in Vault at the correct path and ReconcileSuccessful=True

**Given** a successfully reconciled Policy CR is deleted
**When** the reconciler processes the deletion
**Then** the policy is removed from Vault and the finalizer is cleared

**Implementation notes:** Add `GetPolicyInstance` decoder method (already exists). Create/use test fixtures from `test/kv-engine-admin-policy.yaml`.

### Story 3.2: Integration tests for PasswordPolicy type

As an operator developer,
I want integration tests for the PasswordPolicy type,
So that password policy lifecycle is verified end-to-end.

**Acceptance Criteria:**

**Given** a PasswordPolicy CR is created
**When** the reconciler processes it
**Then** the password policy exists in Vault and ReconcileSuccessful=True

**Given** a successfully reconciled PasswordPolicy is deleted
**When** the reconciler processes the deletion
**Then** the policy is removed from Vault

### Story 3.3: Integration tests for SecretEngineMount type

As an operator developer,
I want integration tests for SecretEngineMount covering create, tune verification, and delete,
So that the foundation for all secret engine types is verified.

**Acceptance Criteria:**

**Given** a SecretEngineMount CR is created (e.g., type=kv)
**When** the reconciler processes it
**Then** the engine is enabled in Vault at `sys/mounts/{path}` and ReconcileSuccessful=True

**Given** a successfully mounted SecretEngineMount is deleted
**When** the reconciler processes the deletion
**Then** the engine is disabled/unmounted in Vault

**Implementation notes:** Add `GetSecretEngineMountInstance` decoder method (already exists). Fixtures exist in `test/`.

### Story 3.4: Integration tests for AuthEngineMount type

As an operator developer,
I want integration tests for AuthEngineMount covering create, tune verification, and delete,
So that the foundation for all auth engine types is verified.

**Acceptance Criteria:**

**Given** an AuthEngineMount CR is created (e.g., type=kubernetes)
**When** the reconciler processes it
**Then** the auth method is enabled in Vault at `sys/auth/{path}` and ReconcileSuccessful=True

**Given** a successfully mounted AuthEngineMount is deleted
**When** the reconciler processes the deletion
**Then** the auth method is disabled in Vault

**Implementation notes:** Add `GetAuthEngineMountInstance` decoder method to `controllertestutils/decoder.go`. Create test fixture YAML.

---

## Epic 4: Auth Engine Integration Tests

Full integration test coverage for all auth engine types.

### Story 4.1: Integration tests for KubernetesAuthEngineConfig and KubernetesAuthEngineRole

As an operator developer,
I want integration tests for the Kubernetes auth engine configuration and role types,
So that the most commonly used auth method has end-to-end test coverage.

**Types:** KubernetesAuthEngineConfig, KubernetesAuthEngineRole
**Fixtures:** `test/kube-auth-engine-config.yaml`, `test/kube-auth-engine-role.yaml`

**Acceptance Criteria:**

**Given** a KubernetesAuthEngineConfig CR is created with a valid config
**When** the reconciler processes it
**Then** the config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

**Given** a KubernetesAuthEngineRole CR is created
**When** the reconciler processes it
**Then** the role exists in Vault and ReconcileSuccessful=True

**Given** both CRs are deleted
**When** the reconciler processes the deletions
**Then** both are cleaned up from Vault

**Implementation notes:** Add decoder methods for both types. These types have non-trivial `PrepareInternalValues` (JWT token from ServiceAccount).

### Story 4.2: Integration tests for LDAPAuthEngineConfig and LDAPAuthEngineGroup

As an operator developer,
I want integration tests for LDAP auth engine types,
So that LDAP authentication configuration is verified end-to-end.

**Types:** LDAPAuthEngineConfig, LDAPAuthEngineGroup
**Fixtures:** `test/ldapauthengine/ldap-auth-engine-config.yaml`, `test/ldapauthengine/ldap-auth-engine-group.yaml`

**Acceptance Criteria:**

**Given** a LDAPAuthEngineConfig CR is created
**When** the reconciler processes it
**Then** the LDAP config is written to Vault and ReconcileSuccessful=True

**Given** a LDAPAuthEngineGroup CR is created
**When** the reconciler processes it
**Then** the group mapping exists in Vault

**Implementation notes:** Requires LDAP test infrastructure (or mock). Add decoder methods. LDAPAuthEngineConfig has non-trivial `PrepareInternalValues` (credentials + TLS).

### Story 4.3: Integration tests for JWTOIDCAuthEngineConfig and JWTOIDCAuthEngineRole

As an operator developer,
I want integration tests for JWT/OIDC auth engine types,
So that JWT/OIDC authentication is verified end-to-end.

**Types:** JWTOIDCAuthEngineConfig, JWTOIDCAuthEngineRole
**Fixtures:** `test/jwtoidcauthengine/jwtoidc-auth-engine-config.yaml`, `test/jwtoidcauthengine/jwtoidc-auth-engine-role.yaml`

**Acceptance Criteria:**

**Given** a JWTOIDCAuthEngineConfig CR and JWTOIDCAuthEngineRole CR are created
**When** the reconcilers process them
**Then** the config and role exist in Vault with ReconcileSuccessful=True

**Given** both CRs are deleted
**When** the reconcilers process the deletions
**Then** both are cleaned up from Vault

**Implementation notes:** Add decoder methods.

---

## Epic 5: Secret Engine Integration Tests

Full integration test coverage for all secret engine types.

### Story 5.1: Integration tests for DatabaseSecretEngineConfig and DatabaseSecretEngineRole

As an operator developer,
I want integration tests for the Database secret engine config and role types,
So that the most complex secret engine (with credential resolution and root password rotation) is verified.

**Types:** DatabaseSecretEngineConfig, DatabaseSecretEngineRole

**Acceptance Criteria:**

**Given** a DatabaseSecretEngineConfig CR with root credentials from a Kubernetes Secret
**When** the reconciler processes it
**Then** the database connection is configured in Vault with `PrepareInternalValues` resolving credentials correctly

**Given** a DatabaseSecretEngineRole CR referencing the config
**When** the reconciler processes it
**Then** the role exists in Vault and can generate credentials

**Implementation notes:** Requires a test database (PostgreSQL in Kind). Decoder methods exist. Most complex integration test — exercises `PrepareInternalValues`, credential resolution, and `IsEquivalentToDesiredState` field remapping.

### Story 5.2: Integration tests for RabbitMQ secret engine types

As an operator developer,
I want integration tests for RabbitMQSecretEngineConfig and RabbitMQSecretEngineRole,
So that the RabbitMQ engine lifecycle is verified.

**Types:** RabbitMQSecretEngineConfig, RabbitMQSecretEngineRole
**Fixtures:** `test/rabbitmq-engine-config.yaml`, `test/rabbitmq-engine-owner-role.yaml`

**Acceptance Criteria:**

**Given** RabbitMQ engine config and role CRs are created
**When** the reconcilers process them
**Then** config and role exist in Vault with ReconcileSuccessful=True

**Implementation notes:** Requires RabbitMQ test instance or mock. RabbitMQSecretEngineConfig uses custom admission webhook handler (manually registered path).

### Story 5.3: Integration tests for remaining secret engine types

As an operator developer,
I want integration tests for GitHub, Azure, Quay, and Kubernetes secret engine types,
So that the full secret engine portfolio has test coverage.

**Types:** GitHubSecretEngineConfig/Role, AzureSecretEngineConfig/Role, QuaySecretEngineConfig/Role/StaticRole, KubernetesSecretEngineConfig/Role

**Acceptance Criteria:**

**Given** each secret engine config and role CR pair is created
**When** the reconcilers process them
**Then** configs and roles exist in Vault with ReconcileSuccessful=True

**Given** the CRs are deleted
**When** the reconcilers process deletions
**Then** Vault resources are cleaned up

**Implementation notes:** Add decoder methods for all types. Some types may need mock external services. Can be broken into sub-tasks per engine.

---

## Epic 6: Identity & Audit Integration Tests

Integration test coverage for identity management and audit types.

### Story 6.1: Integration tests for Group and GroupAlias types

As an operator developer,
I want integration tests for Group and GroupAlias,
So that identity group management is verified.

**Types:** Group, GroupAlias
**Fixtures:** `test/groups/group.yaml`, `test/groups/groupalias.yaml`

**Acceptance Criteria:**

**Given** a Group CR is created
**When** the reconciler processes it
**Then** the group exists in Vault with ReconcileSuccessful=True

**Given** a GroupAlias CR referencing the group and an auth mount
**When** the reconciler processes it
**Then** the alias exists in Vault (GroupAlias has non-trivial `PrepareInternalValues` for accessor lookup)

### Story 6.2: Integration tests for Identity OIDC types

As an operator developer,
I want integration tests for IdentityOIDCProvider, IdentityOIDCScope, IdentityOIDCClient, IdentityOIDCAssignment,
So that the OIDC identity provider configuration lifecycle is verified.

**Types:** IdentityOIDCProvider, IdentityOIDCScope, IdentityOIDCClient, IdentityOIDCAssignment
**Fixtures:** `test/identityoidc/*.yaml`

**Acceptance Criteria:**

**Given** OIDC scope, client, assignment, and provider CRs are created in dependency order
**When** the reconcilers process them
**Then** all OIDC resources exist in Vault with ReconcileSuccessful=True

### Story 6.3: Integration tests for Identity Token types

As an operator developer,
I want integration tests for IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole,
So that the identity token configuration lifecycle is verified.

**Types:** IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole
**Fixtures:** `test/identitytoken/*.yaml`

**Acceptance Criteria:**

**Given** token config, key, and role CRs are created
**When** the reconcilers process them
**Then** all token resources exist in Vault with ReconcileSuccessful=True

### Story 6.4: Integration tests for Audit types

As an operator developer,
I want integration tests for Audit and AuditRequestHeader,
So that audit device management is verified.

**Types:** Audit, AuditRequestHeader

**Acceptance Criteria:**

**Given** an Audit CR for a file audit device is created
**When** the reconciler processes it
**Then** the audit device is enabled in Vault with ReconcileSuccessful=True

**Given** an AuditRequestHeader CR is created
**When** the reconciler processes it
**Then** the audit request header is configured in Vault

**Implementation notes:** Audit types use `VaultAuditResource` reconciler variant. No webhooks exist for these types.

---

## Epic 7: Hardening — Webhooks, Error Paths & Credential Resolution

Webhook validation tests, error path tests, `PrepareInternalValues` unit tests, and drift detection.

### Story 7.1: Webhook validation tests for immutable `spec.path` rule

As an operator developer,
I want unit tests verifying that `ValidateUpdate` rejects changes to `spec.path` on all types that have this rule,
So that the most critical webhook validation is tested.

**Acceptance Criteria:**

**Given** an existing CR instance and a modified copy with a different `spec.path`
**When** `ValidateUpdate(old)` is called
**Then** it returns an error containing "spec.path cannot be updated"

**Given** an existing CR instance and a modified copy with only non-path fields changed
**When** `ValidateUpdate(old)` is called
**Then** it returns nil (update is allowed)

**Implementation notes:** Test all types that have `spec.path` in their Spec. Can be a table-driven test across types.

### Story 7.2: Unit tests for `PrepareInternalValues` credential resolution

As an operator developer,
I want unit tests for the 15 types with non-trivial `PrepareInternalValues`,
So that credential resolution from Kubernetes Secrets, RandomSecrets, and VaultSecrets is verified without a live Vault.

**Types:** DatabaseSecretEngineConfig, RabbitMQSecretEngineConfig, LDAPAuthEngineConfig, AzureAuthEngineConfig, AzureSecretEngineConfig, GCPAuthEngineConfig, GitHubSecretEngineConfig, KubernetesSecretEngineConfig, KubernetesAuthEngineConfig, KubernetesAuthEngineRole, QuaySecretEngineConfig, Policy, EntityAlias, GroupAlias, RandomSecret

**Acceptance Criteria:**

**Given** a type with `RootCredentials.Secret` referencing a Kubernetes Secret
**When** `PrepareInternalValues` is called with a context containing a mock kubeClient
**Then** the internal credential fields are populated correctly

**Given** a Policy with `${auth/kubernetes/@accessor}` placeholder
**When** `PrepareInternalValues` is called with a context containing a Vault client that returns auth engine data
**Then** the placeholder is resolved to the actual accessor string

### Story 7.3: Error path integration tests

As an operator developer,
I want integration tests that verify graceful error handling when Vault is unreachable or authentication fails,
So that the operator doesn't crash or enter infinite retry loops on expected failure conditions.

**Acceptance Criteria:**

**Given** a CR with invalid authentication configuration (non-existent ServiceAccount)
**When** the reconciler attempts to create the Vault client
**Then** `prepareContext` fails, `ManageOutcome` sets ReconcileFailed condition, and the error is logged

**Given** a CR referencing a non-existent Vault path
**When** the reconciler attempts to write
**Then** the error is handled gracefully with ReconcileFailed condition

### Story 7.4: Drift detection integration tests (MOVED — was 7.5, renumbered)

**Note:** Original 7.4 was renumbered to 7.5 to insert the new Story 7.4 below.

### Story 7.4: Audit Vault API responses and harden `IsEquivalentToDesiredState` extra-field handling

As an operator developer,
I want to audit every Vault API read response for all 46 CRD types and ensure `IsEquivalentToDesiredState` correctly ignores extra fields Vault returns,
So that the operator never enters an unnecessary write loop where it rewrites identical state on every reconcile cycle.

**Background:**

The `VaultEndpoint.CreateOrUpdate()` flow reads from Vault, then passes the raw `secret.Data` map directly to `IsEquivalentToDesiredState(payload)` with no pre-filtering. Types handle extra fields inconsistently:
- **Entity/EntityAlias:** Explicitly `delete()` known Vault-added keys before `reflect.DeepEqual` — correct
- **AuditRequestHeader:** Checks only the `hmac` field by type assertion — inherently ignores extras — correct
- **DatabaseSecretEngineConfig:** Custom field remapping logic — correct
- **All other types (~38):** Bare `reflect.DeepEqual(desiredState, payload)` — will return `false` if Vault adds any extra fields, causing an unnecessary write every reconcile

Additionally, some Vault endpoints may return duration values as integers (seconds) rather than the string format the operator sends (e.g., `"24h"` written, `86400` read back), which would also cause false drift detection.

**Acceptance Criteria:**

**Given** a running Vault instance at the project's target version (currently 1.19.0)
**When** each of the 46 CRD types is written to Vault and then read back
**Then** the exact set of extra fields (keys in read response not in write payload) is documented per type

**Given** the documented extra fields per type
**When** each type's `IsEquivalentToDesiredState` is called with a payload containing those extra fields
**Then** it returns `true` (no false drift detected)

**Given** any type where Vault returns values in a different format (e.g., duration as int vs string)
**When** `IsEquivalentToDesiredState` is called with the Vault-formatted values
**Then** it returns `true` (type coercion is handled)

**Implementation notes:**
- Phase 1 (Audit): Deploy Vault via `make deploy-vault`, write test fixtures for all 46 types, read back responses, diff write payload vs read response. Produce a report of extra fields per type and any type coercion issues.
- Phase 2 (Fix): For types with extra fields, adopt one of the proven patterns: (a) explicit `delete()` of known extras (Entity pattern), (b) field-level comparison (AuditRequestHeader pattern), or (c) filter payload to only desired keys before `DeepEqual`.
- Phase 3 (Test): Add unit tests verifying each type handles its specific extra fields correctly.
- Consider a shared helper function (e.g., `filterToDesiredKeys(desired, payload)`) to avoid repeating the pattern in every type.
- Types that require external services (LDAP, RabbitMQ, databases, cloud providers) may need mock responses for the audit step.

### Story 7.5: Drift detection integration tests

As an operator developer,
I want integration tests verifying that when `ENABLE_DRIFT_DETECTION=true`, the reconciler periodically re-checks Vault state and corrects drift,
So that the drift detection feature is validated.

**Dependency:** Story 7.4 (extra-field audit) should be completed first — drift detection relies on `IsEquivalentToDesiredState` returning correct results to detect actual drift vs false positives from extra fields.

**Acceptance Criteria:**

**Given** a resource is successfully reconciled with drift detection enabled
**When** the Vault state is manually modified (via Vault client directly)
**And** enough time passes for the periodic reconciliation to trigger
**Then** the reconciler detects the drift via `IsEquivalentToDesiredState` returning false and writes the correct state back to Vault

---
---

# Phase 2: Expansion — Dependency Upgrades + Engine Coverage

## Phase 2 Requirements Inventory

### Dependency Upgrade Requirements

DU1: Upgrade Go from 1.22 to 1.24+ (required by latest controller-runtime, ginkgo, gomega)
DU2: Upgrade controller-runtime from v0.17.3 to v0.23.x in lockstep with K8s client libraries (api, apimachinery, client-go, apiextensions-apiserver) from v0.29.2 to v0.35.x
DU3: Upgrade Operator SDK from v1.31.0 to v1.42.x (Makefile, Dockerfile, bundle, project layout)
DU4: Upgrade hashicorp/vault/api from v1.14.0 to v1.23.x
DU5: Upgrade test dependencies (ginkgo v2.19→v2.28, gomega v1.33→v1.39)
DU6: Upgrade peripheral dependencies (hcl/v2 v2.21→v2.24, sprig/v3 v3.2→v3.3, logr v1.4.2→v1.4.3)
DU7: Upgrade security-sensitive indirect dependencies (golang.org/x/crypto, golang.org/x/net)
DU8: Evaluate migration from archived `pkg/errors` to Go standard `fmt.Errorf` with `%w` wrapping
DU9: Upgrade Makefile K8s-coupled tools: controller-gen v0.14→v0.20, envtest release-0.17→release-0.23, ENVTEST_K8S_VERSION 1.29→1.35, kubectl v1.29→v1.35, Kind v0.27→v0.31
DU10: Upgrade Dockerfile builder image from golang:1.22 to golang:1.24+; add multi-arch support
DU11: Update CI workflow version references (GO_VERSION, OPERATOR_SDK_VERSION, reusable workflow pin)
DU12: Upgrade Helm from v3.11.0 to v4.x (major version change with breaking changes)
DU13: Upgrade golangci-lint from v1.59.1 to v2.x (major version change with config format changes)
DU14: Upgrade OPM from v1.23.0 to v1.65.x
DU15: Upgrade Kustomize from v5.4.3 to v5.8.x
DU16: Upgrade Vault integration test infrastructure (Vault 1.19→1.21, chart version, vault-values.yaml images)
DU17: Upgrade cert-manager in helmchart-test from v1.7.1 to current
DU18: Update bundle.Dockerfile labels for new Operator SDK version

### Secret Engine Requirements

SE1: AWS secret engine CRDs (config + role for IAM users, STS assumed roles, federation tokens)
SE2: Transit secret engine CRDs (config + key management + encryption/decryption operations)
SE3: SSH secret engine CRDs (config + role for signed keys and OTP)
SE4: Consul secret engine CRDs (config + role for dynamic ACL tokens)
SE5: GCP secret engine CRDs (config + role for service account keys and OAuth tokens)
SE6: LDAP/AD secret engine CRDs (config + static role + dynamic role + library set)
SE7: Nomad secret engine CRDs (config + role for dynamic Nomad tokens)
SE8: TOTP secret engine CRDs (key generation and management)
SE9: MongoDB Atlas secret engine CRDs (config + role for dynamic API keys)
SE10: Terraform Cloud secret engine CRDs (config + role for dynamic API tokens)

### Auth Engine Requirements

AE1: AppRole auth engine CRDs (config + role + secret-id management)
AE2: AWS auth engine CRDs (config + role for IAM and EC2 auth types)
AE3: Userpass auth engine CRDs (config + user management)
AE4: GitHub auth engine CRDs (config + team/org mapping)
AE5: Okta auth engine CRDs (config + group mapping)
AE6: RADIUS auth engine CRDs (config + user mapping)
AE7: AliCloud auth engine CRDs (config + role)
AE8: OCI auth engine CRDs (config + role)
AE9: Kerberos auth engine CRDs (config + SPNEGO setup)
AE10: Cloud Foundry auth engine CRDs (config + role)

### Phase 2 Non-Functional Requirements

PNFR1: All dependency upgrades must maintain backward compatibility with existing CRD APIs (no CRD schema changes)
PNFR2: CI pipeline must pass after each upgrade epic
PNFR3: New CRD types must follow established patterns from project-context.md (VaultObject, ConditionsAware, toMap, IsEquivalentToDesiredState, webhooks)
PNFR4: New types must include unit tests for toMap() and IsEquivalentToDesiredState() from day one
PNFR5: New types must include integration tests (create, update, delete scenarios) from day one
PNFR6: New types must have admission webhooks with immutable spec.path validation

### Phase 2 FR Coverage Map

DU1, DU10, DU11: Epic 8 — Go + Kubernetes stack upgrade (Go version, Dockerfile, CI)
DU2, DU9: Epic 8 — controller-runtime + K8s libs + coupled Makefile tools
DU4-DU8: Epic 9 — Vault API + peripheral + security upgrades, pkg/errors evaluation
DU16: Epic 9 — Vault integration test infrastructure upgrade
DU3, DU18: Epic 10 — Operator SDK upgrade + bundle.Dockerfile
DU12, DU17: Epic 10 — Helm v3→v4 upgrade + cert-manager update
DU13: Epic 10 — golangci-lint v1→v2 upgrade
DU14, DU15: Epic 10 — OPM + Kustomize upgrades
SE1-SE3: Epic 11 — High-priority missing secret engines (AWS, Transit, SSH)
SE4-SE6: Epic 12 — Medium-priority missing secret engines (Consul, GCP, LDAP/AD)
SE7-SE10: Epic 13 — Lower-priority missing secret engines (Nomad, TOTP, MongoDB Atlas, Terraform Cloud)
AE1-AE2: Epic 14 — High-priority missing auth methods (AppRole, AWS)
AE3-AE5: Epic 15 — Medium-priority missing auth methods (Userpass, GitHub, Okta)
AE6-AE10: Epic 16 — Lower-priority missing auth methods (RADIUS, AliCloud, OCI, Kerberos, CF)

## Phase 2 Epic List

### Epic 8: Go + Kubernetes Stack Upgrade
Upgrade Go 1.22→1.24, controller-runtime v0.17→v0.23, K8s libs v0.29→v0.35, and all coupled tools (controller-gen, envtest, kubectl, Kind). Update Dockerfile, CI workflows, and Makefile version variables.
**FRs covered:** DU1, DU2, DU9, DU10, DU11

### Epic 9: Vault API + Peripheral Dependency Upgrades
Upgrade vault/api v1.14→v1.23, test deps (ginkgo, gomega), peripheral deps (hcl, sprig, logr), security-sensitive indirect deps (x/crypto, x/net), Vault integration test infrastructure (1.19→1.21), and evaluate pkg/errors migration.
**FRs covered:** DU4, DU5, DU6, DU7, DU8, DU16

### Epic 10: Operator SDK + Build Tooling Upgrades
Upgrade Operator SDK v1.31→v1.42, Helm v3→v4, golangci-lint v1→v2, OPM v1.23→v1.65, Kustomize v5.4→v5.8, cert-manager in helmchart-test, and bundle.Dockerfile labels.
**FRs covered:** DU3, DU12, DU13, DU14, DU15, DU17, DU18

### Epic 11: High-Priority Missing Secret Engines (AWS, Transit, SSH)
Implement CRD types for AWS, Transit, and SSH secret engines following established operator patterns.
**FRs covered:** SE1, SE2, SE3

### Epic 12: Medium-Priority Missing Secret Engines (Consul, GCP, LDAP/AD)
Implement CRD types for Consul, GCP, and LDAP/AD secret engines.
**FRs covered:** SE4, SE5, SE6

### Epic 13: Lower-Priority Missing Secret Engines (Nomad, TOTP, MongoDB Atlas, Terraform Cloud)
Implement CRD types for remaining secret engines with lower community demand.
**FRs covered:** SE7, SE8, SE9, SE10

### Epic 14: High-Priority Missing Auth Methods (AppRole, AWS)
Implement CRD types for AppRole and AWS auth methods — the two most commonly used auth methods not yet supported.
**FRs covered:** AE1, AE2

### Epic 15: Medium-Priority Missing Auth Methods (Userpass, GitHub, Okta)
Implement CRD types for Userpass, GitHub, and Okta auth methods.
**FRs covered:** AE3, AE4, AE5

### Epic 16: Lower-Priority Missing Auth Methods (RADIUS, AliCloud, OCI, Kerberos, CF)
Implement CRD types for remaining auth methods with lower community demand.
**FRs covered:** AE6, AE7, AE8, AE9, AE10

---

## Epic 8: Go + Kubernetes Stack Upgrade

Upgrade the core Go + Kubernetes dependency stack. This is the highest-risk upgrade — controller-runtime, K8s client libs, and Go must move together. Requires adaptation of every controller, webhook, envtest suite, CI pipeline, and Dockerfile.

**Precondition:** Phase 1 test stabilization must be substantially complete — the test suite is the safety net for this upgrade.

### Story 8.1: Upgrade Go from 1.22 to 1.24

As an operator developer,
I want to upgrade the Go version from 1.22 to 1.24,
So that we can use the latest controller-runtime and benefit from Go language improvements and security fixes.

**Acceptance Criteria:**

**Given** the project uses Go 1.22
**When** go.mod is updated to `go 1.24` and all source files are adapted
**Then** `go build ./...` succeeds, `go vet ./...` passes, `go test ./...` passes

**Given** the Dockerfile builder stage references `golang:1.22`
**When** the base image is updated to `golang:1.24`
**Then** the container builds and operator binary runs correctly

**Given** CI workflows (pr.yaml, push.yaml) reference `GO_VERSION: ~1.22`
**When** both are updated to `GO_VERSION: ~1.24`
**Then** all CI jobs pass

**Given** the Dockerfile hardcodes `GOARCH=amd64`
**When** the build is updated to use `TARGETARCH` build arg for multi-arch support
**Then** the image can be built for both amd64 and arm64

### Story 8.2: Upgrade controller-runtime v0.17 → v0.23 and K8s libs v0.29 → v0.35

As an operator developer,
I want to upgrade controller-runtime and K8s client libraries to the latest versions,
So that the operator is compatible with current Kubernetes versions and benefits from upstream fixes.

**Acceptance Criteria:**

**Given** go.mod pins controller-runtime v0.17.3 and K8s libs v0.29.2
**When** all are updated to v0.23.x / v0.35.x
**Then** `go build ./...` succeeds after adapting to any API changes

**Given** controller-runtime may have changed webhook registration, manager options, or reconciler interfaces
**When** all 47 controllers and associated webhooks are adapted
**Then** the operator starts, registers all controllers and webhooks, and reconciles correctly

**Given** envtest behavior may differ between versions
**When** both unit and integration test suites are adapted
**Then** `make test` and `make integration` pass

**Implementation notes:** Review controller-runtime release notes for each minor version (v0.18 through v0.23) for breaking changes. Key areas: Manager.Start() signature, webhook server setup, envtest CRD loading, client.Reader vs client.Client, predicate APIs. This is the largest single story in the project.

### Story 8.3: Upgrade Makefile K8s-coupled tool versions

As an operator developer,
I want to upgrade the Makefile tool versions that must track the K8s stack,
So that tooling is compatible with the new controller-runtime and K8s lib versions.

**Versions to update:**

| Variable | Current | Target |
|----------|---------|--------|
| CONTROLLER_TOOLS_VERSION | v0.14.0 | v0.20.1 |
| ENVTEST_VERSION | release-0.17 | release-0.23 |
| ENVTEST_K8S_VERSION | 1.29.0 | 1.35.0 |
| KUBECTL_VERSION | v1.29.0 | v1.35.x |
| KIND_VERSION | v0.27.0 | v0.31.0 |

**Acceptance Criteria:**

**Given** controller-gen v0.14 is used
**When** CONTROLLER_TOOLS_VERSION is updated to v0.20.1
**Then** `make manifests` and `make generate` produce valid CRDs and deepcopy code

**Given** ENVTEST_VERSION and ENVTEST_K8S_VERSION are updated
**When** `make test` is run
**Then** envtest downloads and uses the correct K8s 1.35 binaries

**Given** KIND_VERSION is updated to v0.31.0 and KUBECTL_VERSION to v1.35.x
**When** `make kind-setup` and `make integration` are run
**Then** the Kind cluster and integration tests work correctly

### Story 8.4: Adapt CI pipeline and Dockerfiles for new versions

As an operator developer,
I want the CI workflows, Dockerfiles, and related configs to reflect the upgraded versions,
So that builds, tests, and releases work with the updated stack.

**Acceptance Criteria:**

**Given** pr.yaml and push.yaml reference `GO_VERSION: ~1.22` and `OPERATOR_SDK_VERSION: v1.31.0`
**When** all version references are updated
**Then** CI runs green on a PR with all changes

**Given** the Kind node image in `make kind-setup` uses `kindest/node:$(KUBECTL_VERSION)`
**When** KUBECTL_VERSION is updated
**Then** the Kind cluster boots with the correct K8s version

**Given** the `redhat-cop/github-workflows-operators` is pinned at v1.1.6
**When** we check for newer versions of the reusable workflow
**Then** the pin is updated if a newer compatible version exists

---

## Epic 9: Vault API + Peripheral Dependency Upgrades

Upgrade vault/api, test dependencies, peripheral libraries, and security-sensitive indirect dependencies. These are lower-risk than the K8s stack upgrade and can be done independently.

### Story 9.1: Upgrade hashicorp/vault/api from v1.14 to v1.23

As an operator developer,
I want to upgrade the Vault API client library to the latest version,
So that the operator can use new Vault API features and benefits from upstream fixes.

**Acceptance Criteria:**

**Given** go.mod pins vault/api v1.14.0
**When** it is updated to v1.23.0
**Then** `go build ./...` succeeds and all tests pass

**Given** vault/api v1.23 may expose new client options or deprecate old ones
**When** the codebase is reviewed for deprecated API usage
**Then** any deprecated patterns are migrated

**Implementation notes:** Review vault/api CHANGELOG for breaking changes between v1.14 and v1.23. Key areas: client configuration, TLS setup, retry behavior, response handling.

### Story 9.2: Upgrade test dependencies (ginkgo, gomega)

As an operator developer,
I want to upgrade Ginkgo to v2.28 and Gomega to v1.39,
So that we benefit from new test features (JSON reports, semantic version filtering) and remain on supported versions.

**Acceptance Criteria:**

**Given** go.mod pins ginkgo v2.19.0 and gomega v1.33.1
**When** both are updated to latest
**Then** `make test` and `make integration` pass without test changes (backward compatible)

### Story 9.3: Upgrade peripheral and security dependencies

As an operator developer,
I want to upgrade hcl/v2, sprig/v3, logr, x/crypto, and x/net to their latest versions,
So that we have current security patches and bug fixes.

**Acceptance Criteria:**

**Given** multiple peripheral deps are behind by 1-3 minor versions
**When** all are updated via `go get -u` for each
**Then** `go build ./...` succeeds and all tests pass

### Story 9.4: Evaluate and plan pkg/errors migration

As an operator developer,
I want to assess the effort of migrating from the archived `github.com/pkg/errors` to Go standard error wrapping (`fmt.Errorf` with `%w`),
So that we can decide whether to include this in the upgrade or defer it.

**Acceptance Criteria:**

**Given** `pkg/errors` is used throughout the codebase (Wrap, Wrapf, New, WithStack)
**When** all call sites are inventoried
**Then** a migration plan is produced with estimated effort

**Given** the migration plan
**When** reviewed by the team
**Then** a decision is made: migrate now, migrate incrementally, or defer

### Story 9.5: Upgrade Vault version in integration test infrastructure

As an operator developer,
I want to upgrade the Vault version used in integration tests and local development from 1.19.x to 1.21.x,
So that we test against a current Vault release.

**Files to update:**

| File | Current | Target |
|------|---------|--------|
| Makefile `VAULT_VERSION` | 1.19.0 | 1.21.4 |
| Makefile `VAULT_CHART_VERSION` | 0.30.0 | Matching chart for 1.21.x |
| integration/vault-values.yaml (4 image refs) | hashicorp/vault:1.19.0 | hashicorp/vault:1.21.4 |
| config/local-development/vault-values.yaml (3 image refs) | hashicorp/vault:1.19.2-ubi | hashicorp/vault:1.21.x-ubi |

**Acceptance Criteria:**

**Given** integration tests use Vault 1.19.0
**When** all Vault version references are updated to 1.21.4
**Then** `make integration` passes against the new Vault version

**Given** the Vault Helm chart version must match the Vault version
**When** VAULT_CHART_VERSION is updated to the chart that ships Vault 1.21.x
**Then** `make deploy-vault` deploys the correct Vault version

---

## Epic 10: Operator SDK + Build Tooling Upgrades

Upgrade Operator SDK, Helm, golangci-lint, OPM, and other build/CI tools that have major version gaps.

### Story 10.1: Upgrade Operator SDK from v1.31 to v1.42

As an operator developer,
I want to upgrade Operator SDK to the latest version,
So that we benefit from improved scaffolding, bundle generation, and compatibility with the latest OLM.

**Files to update:**
- Makefile: `OPERATOR_SDK_VERSION`
- pr.yaml, push.yaml: `OPERATOR_SDK_VERSION: v1.31.0`
- bundle.Dockerfile: `operator-sdk-v1.31.0` label and `go.kubebuilder.io/v3` layout label

**Acceptance Criteria:**

**Given** the project uses Operator SDK v1.31
**When** the SDK tooling is updated to v1.42 and the Makefile/Dockerfile are adapted per the upgrade guide
**Then** `make manifests generate` produces valid output
**And** `make bundle` generates a valid OLM bundle
**And** `make docker-build` succeeds

**Given** bundle.Dockerfile labels reference v1.31.0
**When** labels are updated to v1.42.x
**Then** the bundle image builds and validates correctly

**Implementation notes:** Follow the official [Operator SDK upgrade guide](https://sdk.operatorframework.io/docs/upgrading-sdk-version/) for each minor version from v1.31 to v1.42. Key areas: Makefile targets, kustomize version, OLM bundle format, scorecard config.

### Story 10.2: Upgrade Helm from v3.11 to v4.x

As an operator developer,
I want to upgrade Helm from v3.11.0 to v4.x,
So that we benefit from Helm 4 features and remain on a supported version.

**Acceptance Criteria:**

**Given** HELM_VERSION is v3.11.0
**When** it is updated to v4.1.x
**Then** `make helmchart` produces a valid chart
**And** `make helmchart-test` passes on the Kind cluster

**Given** Helm 4 has breaking changes from Helm 3
**When** all Helm commands in the Makefile are reviewed
**Then** any incompatible flags or behaviors are adapted

**Given** the helmchart-test target installs cert-manager v1.7.1 (long EOL)
**When** the cert-manager version is updated to a current release
**Then** the helmchart test passes with the updated cert-manager

**Implementation notes:** Helm 4 changes include: chart API version requirements, plugin validation, dependency handling. The kustomize v5.8+ release includes Helm 4 compat fixes. Also update cert-manager from v1.7.1 to current.

### Story 10.3: Upgrade golangci-lint from v1.59 to v2.x

As an operator developer,
I want to upgrade golangci-lint from v1.59.1 to v2.x,
So that we benefit from new linters, performance improvements, and remain on the supported major version.

**Acceptance Criteria:**

**Given** GOLANGCI_LINT_VERSION is v1.59.1
**When** it is updated to v2.11.x
**Then** `make golangci-lint` downloads the new version

**Given** golangci-lint v2 changed config file format and renamed some linters
**When** `.golangci.yml` is migrated to v2 format
**Then** `golangci-lint run` passes (or new findings are triaged)

**Implementation notes:** golangci-lint v2 is a major version with config format changes. Use `golangci-lint migrate` to convert the config file. Some linters were renamed or removed.

### Story 10.4: Upgrade OPM and Kustomize

As an operator developer,
I want to upgrade OPM from v1.23.0 to v1.65.x and Kustomize from v5.4.3 to v5.8.x,
So that OLM catalog building and kustomize rendering use current tooling.

**Acceptance Criteria:**

**Given** OPM version is hardcoded to v1.23.0 in the Makefile `opm` target
**When** it is updated to v1.65.x
**Then** `make catalog-build` succeeds

**Given** KUSTOMIZE_VERSION is v5.4.3
**When** it is updated to v5.8.1
**Then** `make manifests`, `make deploy`, `make helmchart` all succeed

---

## Epic 11: High-Priority Missing Secret Engines (AWS, Transit, SSH)

Implement the three most requested missing secret engines. Each follows the established pattern: types, controller, webhook, tests, test fixtures, decoder.

### Story 11.1: AWS Secret Engine — Config and Role CRDs

As an operator developer,
I want CRDs for AWSSecretEngineConfig (root credentials config) and AWSSecretEngineRole (IAM user, assumed role, federation token),
So that Vault's AWS secret engine can be managed declaratively.

**Vault API paths:** `{path}/config/root`, `{path}/config/lease`, `{path}/roles/{name}`
**Vault docs:** https://developer.hashicorp.com/vault/api-docs/secret/aws

**Acceptance Criteria:**

**Given** an AWSSecretEngineConfig CR is created with AWS access key credentials
**When** the reconciler processes it
**Then** the root config is written to Vault and ReconcileSuccessful=True

**Given** an AWSSecretEngineRole CR is created with credential_type (iam_user | assumed_role | federation_token)
**When** the reconciler processes it
**Then** the role exists in Vault and can generate dynamic AWS credentials

**Given** both CRs are deleted
**When** the reconcilers process deletions
**Then** Vault resources are cleaned up

**Implementation notes:** Implement VaultObject, ConditionsAware interfaces. AWSSecretEngineConfig has non-trivial PrepareInternalValues (AWS credentials from K8s Secret). Include unit tests for toMap/IsEquivalentToDesiredState, integration tests, webhook with immutable path. AWS roles have 3 credential types with different field sets — toMap must handle conditional fields.

### Story 11.2: Transit Secret Engine — Config and Key CRDs

As an operator developer,
I want CRDs for TransitSecretEngineKey (encryption key lifecycle),
So that Vault's Transit encryption-as-a-service can be managed declaratively.

**Vault API paths:** `{path}/keys/{name}`, `{path}/keys/{name}/config`
**Vault docs:** https://developer.hashicorp.com/vault/api-docs/secret/transit

**Acceptance Criteria:**

**Given** a TransitSecretEngineKey CR is created with key type and configuration
**When** the reconciler processes it
**Then** the key exists in Vault and ReconcileSuccessful=True

**Given** the key spec is updated (e.g., min_decryption_version, deletion_allowed)
**When** the reconciler processes the update
**Then** the key configuration is updated in Vault via the config endpoint

**Given** the CR is deleted
**When** the reconciler processes deletion
**Then** the key is deleted from Vault (if deletion_allowed=true)

**Implementation notes:** Transit keys have two API paths — create at `keys/{name}` and update config at `keys/{name}/config`. IsEquivalentToDesiredState must compare against the config-level fields only (Vault returns full key metadata on read). Consider a TransitSecretEngineCache CRD for cache configuration if needed.

### Story 11.3: SSH Secret Engine — Config and Role CRDs

As an operator developer,
I want CRDs for SSHSecretEngineConfig and SSHSecretEngineRole,
So that Vault's SSH secret engine (signed keys and OTP) can be managed declaratively.

**Vault API paths:** `{path}/config/ca`, `{path}/roles/{name}`
**Vault docs:** https://developer.hashicorp.com/vault/api-docs/secret/ssh

**Acceptance Criteria:**

**Given** an SSHSecretEngineConfig CR is created with CA key configuration
**When** the reconciler processes it
**Then** the SSH CA is configured in Vault and ReconcileSuccessful=True

**Given** an SSHSecretEngineRole CR is created with key_type (ca | otp)
**When** the reconciler processes it
**Then** the role exists in Vault

**Implementation notes:** SSH has two modes (signed certs vs OTP) with different role fields. toMap must handle conditional fields based on key_type. Config involves CA key pair (may need PrepareInternalValues for private key from K8s Secret).

---

## Epic 12: Medium-Priority Missing Secret Engines (Consul, GCP, LDAP/AD)

### Story 12.1: Consul Secret Engine — Config and Role CRDs

As an operator developer,
I want CRDs for ConsulSecretEngineConfig and ConsulSecretEngineRole,
So that Vault's Consul secret engine can be managed declaratively.

**Vault API paths:** `{path}/config/access`, `{path}/roles/{name}`

**Acceptance Criteria:**

**Given** a ConsulSecretEngineConfig CR is created with Consul address and token
**When** the reconciler processes it
**Then** the Consul config is written to Vault and ReconcileSuccessful=True

**Given** a ConsulSecretEngineRole CR is created with policies/service identities
**When** the reconciler processes it
**Then** the role exists in Vault and can generate dynamic Consul tokens

**Implementation notes:** ConsulSecretEngineConfig has PrepareInternalValues for Consul token from K8s Secret. Straightforward pattern following existing engine implementations.

### Story 12.2: GCP Secret Engine — Config and Role CRDs

As an operator developer,
I want CRDs for GCPSecretEngineConfig and GCPSecretEngineRole,
So that Vault's GCP secret engine can be managed declaratively (complementing the existing GCP auth engine CRDs).

**Vault API paths:** `{path}/config`, `{path}/roleset/{name}`, `{path}/static-account/{name}`

**Acceptance Criteria:**

**Given** a GCPSecretEngineConfig CR is created with GCP credentials
**When** the reconciler processes it
**Then** the config is written to Vault

**Given** a GCPSecretEngineRoleset CR is created
**When** the reconciler processes it
**Then** the roleset exists in Vault and can generate service account keys or OAuth tokens

**Implementation notes:** GCP secrets engine uses "rolesets" (not "roles") and "static accounts". PrepareInternalValues for GCP credentials JSON from K8s Secret.

### Story 12.3: LDAP/AD Secret Engine — Config and Role CRDs

As an operator developer,
I want CRDs for LDAPSecretEngineConfig, LDAPSecretEngineStaticRole, and LDAPSecretEngineDynamicRole,
So that Vault's LDAP/AD secret engine (password rotation, dynamic credentials) can be managed declaratively.

**Vault API paths:** `{path}/config`, `{path}/static-role/{name}`, `{path}/role/{name}`, `{path}/library/{name}`

**Acceptance Criteria:**

**Given** an LDAPSecretEngineConfig CR is created with LDAP bind credentials
**When** the reconciler processes it
**Then** the LDAP config is written to Vault

**Given** an LDAPSecretEngineStaticRole CR is created
**When** the reconciler processes it
**Then** the static role exists in Vault for managed password rotation

**Implementation notes:** LDAP secrets engine supports 3 schemas (openldap, ad, racf). PrepareInternalValues for bind credentials from K8s Secret. Consider library sets as a separate CRD (LDAPSecretEngineLibrary).

---

## Epic 13: Lower-Priority Missing Secret Engines (Nomad, TOTP, MongoDB Atlas, Terraform Cloud)

### Story 13.1: Nomad Secret Engine — Config and Role CRDs

As an operator developer,
I want CRDs for NomadSecretEngineConfig and NomadSecretEngineRole,
So that Vault's Nomad secret engine can be managed declaratively.

**Vault API paths:** `{path}/config/access`, `{path}/role/{name}`

**Acceptance Criteria:**

**Given** config and role CRs are created
**When** the reconcilers process them
**Then** both exist in Vault with ReconcileSuccessful=True and can generate dynamic Nomad tokens

### Story 13.2: TOTP Secret Engine — Key CRD

As an operator developer,
I want a CRD for TOTPSecretEngineKey,
So that TOTP key generation and management can be managed declaratively.

**Vault API paths:** `{path}/keys/{name}`

**Acceptance Criteria:**

**Given** a TOTPSecretEngineKey CR is created with issuer and account name
**When** the reconciler processes it
**Then** the TOTP key exists in Vault

### Story 13.3: MongoDB Atlas Secret Engine — Config and Role CRDs

As an operator developer,
I want CRDs for MongoDBAtlasSecretEngineConfig and MongoDBAtlasSecretEngineRole,
So that Vault's MongoDB Atlas secret engine can be managed declaratively.

**Vault API paths:** `{path}/config`, `{path}/roles/{name}`

**Acceptance Criteria:**

**Given** config and role CRs are created
**When** the reconcilers process them
**Then** both exist in Vault and can generate dynamic Atlas API keys

### Story 13.4: Terraform Cloud Secret Engine — Config and Role CRDs

As an operator developer,
I want CRDs for TerraformCloudSecretEngineConfig and TerraformCloudSecretEngineRole,
So that Vault's Terraform Cloud secret engine can be managed declaratively.

**Vault API paths:** `{path}/config`, `{path}/role/{name}`

**Acceptance Criteria:**

**Given** config and role CRs are created
**When** the reconcilers process them
**Then** both exist in Vault and can generate dynamic Terraform Cloud API tokens

---

## Epic 14: High-Priority Missing Auth Methods (AppRole, AWS)

### Story 14.1: AppRole Auth Engine — Config and Role CRDs

As an operator developer,
I want CRDs for AppRoleAuthEngineConfig and AppRoleAuthEngineRole,
So that Vault's AppRole auth method (the #1 machine-to-machine auth) can be managed declaratively.

**Vault API paths:** `auth/{path}/role/{name}`, `auth/{path}/role/{name}/secret-id`

**Acceptance Criteria:**

**Given** an AppRoleAuthEngineRole CR is created with policies and token settings
**When** the reconciler processes it
**Then** the role exists in Vault and ReconcileSuccessful=True

**Given** the role spec includes secret_id configuration (bound_cidr_list, secret_id_ttl)
**When** the reconciler processes it
**Then** the secret-id constraints are configured on the role

**Given** the CR is deleted
**When** the reconciler processes deletion
**Then** the role is removed from Vault

**Implementation notes:** AppRole has no separate "config" endpoint — the mount itself is the config. Roles are the primary resource. Secret-ID management is operational (generate, list, destroy) — the CRD manages the role definition, not individual secret-IDs. Consider whether a separate CRD for secret-id generation is needed.

### Story 14.2: AWS Auth Engine — Config and Role CRDs

As an operator developer,
I want CRDs for AWSAuthEngineConfig and AWSAuthEngineRole,
So that Vault's AWS auth method can be managed declaratively.

**Vault API paths:** `auth/{path}/config/client`, `auth/{path}/config/identity`, `auth/{path}/role/{name}`

**Acceptance Criteria:**

**Given** an AWSAuthEngineConfig CR is created with AWS credentials and STS endpoint
**When** the reconciler processes it
**Then** the client config is written to Vault

**Given** an AWSAuthEngineRole CR is created with auth_type (iam | ec2) and bound constraints
**When** the reconciler processes it
**Then** the role exists in Vault

**Implementation notes:** AWSAuthEngineConfig has PrepareInternalValues for AWS credentials from K8s Secret. Roles have two auth types (IAM, EC2) with different bound constraint fields — toMap must handle conditional fields.

---

## Epic 15: Medium-Priority Missing Auth Methods (Userpass, GitHub, Okta)

### Story 15.1: Userpass Auth Engine — User CRD

As an operator developer,
I want a CRD for UserpassAuthEngineUser,
So that Vault userpass accounts can be managed declaratively.

**Vault API paths:** `auth/{path}/users/{name}`

**Acceptance Criteria:**

**Given** a UserpassAuthEngineUser CR is created with policies and token settings
**When** the reconciler processes it
**Then** the user exists in Vault (password from K8s Secret via PrepareInternalValues)

**Implementation notes:** Userpass has no separate config endpoint. The user CRD manages user accounts. Password should come from K8s Secret reference, not inline in the CR spec.

### Story 15.2: GitHub Auth Engine — Config and Team/Org Mapping CRDs

As an operator developer,
I want CRDs for GitHubAuthEngineConfig and GitHubAuthEngineTeamMap/OrgMap,
So that Vault's GitHub auth method can be managed declaratively.

**Vault API paths:** `auth/{path}/config`, `auth/{path}/map/teams/{name}`, `auth/{path}/map/users/{name}`

**Acceptance Criteria:**

**Given** a GitHubAuthEngineConfig CR is created with organization
**When** the reconciler processes it
**Then** the config is written to Vault

**Given** team/user mapping CRs are created
**When** the reconciler processes them
**Then** the mappings exist in Vault

### Story 15.3: Okta Auth Engine — Config and Group CRDs

As an operator developer,
I want CRDs for OktaAuthEngineConfig and OktaAuthEngineGroup,
So that Vault's Okta auth method can be managed declaratively.

**Vault API paths:** `auth/{path}/config`, `auth/{path}/groups/{name}`, `auth/{path}/users/{name}`

**Acceptance Criteria:**

**Given** an OktaAuthEngineConfig CR is created with Okta org, API token, and base URL
**When** the reconciler processes it
**Then** the config is written to Vault

**Given** group mapping CRs are created
**When** the reconciler processes them
**Then** the group mappings exist in Vault

**Implementation notes:** OktaAuthEngineConfig has PrepareInternalValues for API token from K8s Secret.

---

## Epic 16: Lower-Priority Missing Auth Methods (RADIUS, AliCloud, OCI, Kerberos, CF)

### Story 16.1: RADIUS Auth Engine — Config and User CRDs

As an operator developer,
I want CRDs for RADIUSAuthEngineConfig and RADIUSAuthEngineUser,
So that Vault's RADIUS auth method can be managed declaratively.

**Vault API paths:** `auth/{path}/config`, `auth/{path}/users/{name}`

**Acceptance Criteria:**

**Given** config and user CRs are created
**When** the reconcilers process them
**Then** the RADIUS config and user policies exist in Vault

### Story 16.2: AliCloud Auth Engine — Config and Role CRDs

As an operator developer,
I want CRDs for AliCloudAuthEngineConfig and AliCloudAuthEngineRole,
So that Vault's AliCloud auth method can be managed declaratively.

**Vault API paths:** `auth/{path}/config`, `auth/{path}/role/{name}`

**Acceptance Criteria:**

**Given** config and role CRs are created
**When** the reconcilers process them
**Then** both exist in Vault with ReconcileSuccessful=True

### Story 16.3: OCI Auth Engine — Config and Role CRDs

As an operator developer,
I want CRDs for OCIAuthEngineConfig and OCIAuthEngineRole,
So that Vault's OCI auth method can be managed declaratively.

**Vault API paths:** `auth/{path}/config`, `auth/{path}/role/{name}`

**Acceptance Criteria:**

**Given** config and role CRs are created
**When** the reconcilers process them
**Then** both exist in Vault with ReconcileSuccessful=True

### Story 16.4: Kerberos Auth Engine — Config and LDAP Group CRDs

As an operator developer,
I want CRDs for KerberosAuthEngineConfig and KerberosAuthEngineLDAPGroup,
So that Vault's Kerberos auth method can be managed declaratively.

**Vault API paths:** `auth/{path}/config`, `auth/{path}/config/ldap`, `auth/{path}/groups/{name}`

**Acceptance Criteria:**

**Given** a KerberosAuthEngineConfig CR is created with keytab and service account
**When** the reconciler processes it
**Then** the Kerberos config is written to Vault

**Implementation notes:** Kerberos has a two-part config (Kerberos config + LDAP config). Keytab content via K8s Secret PrepareInternalValues.

### Story 16.5: Cloud Foundry Auth Engine — Config and Role CRDs

As an operator developer,
I want CRDs for CFAuthEngineConfig and CFAuthEngineRole,
So that Vault's Cloud Foundry auth method can be managed declaratively.

**Vault API paths:** `auth/{path}/config`, `auth/{path}/roles/{name}`

**Acceptance Criteria:**

**Given** config and role CRs are created
**When** the reconcilers process them
**Then** both exist in Vault with ReconcileSuccessful=True
