# Story 3.2: Integration Tests for PasswordPolicy Type

Status: ready-for-dev

## Story

As an operator developer,
I want integration tests for the PasswordPolicy type covering create, reconcile success, Vault state verification, password generation, and delete with Vault cleanup,
So that the password policy lifecycle is verified end-to-end.

## Acceptance Criteria

1. **Given** a PasswordPolicy CR is created in the test namespace with a valid HCL password policy **When** the reconciler processes it **Then** the password policy exists in Vault at `sys/policies/password/<name>` and ReconcileSuccessful=True

2. **Given** a successfully reconciled PasswordPolicy **When** the Vault password generate endpoint is called **Then** a password is returned that matches the policy constraints (e.g., 20 lowercase characters)

3. **Given** a PasswordPolicy CR using `spec.name` (overriding `metadata.name`) **When** the reconciler processes it **Then** the policy exists in Vault at `sys/policies/password/<spec.name>` (not `<metadata.name>`)

4. **Given** a successfully reconciled PasswordPolicy CR is deleted **When** the reconciler processes the deletion **Then** the policy is removed from Vault and the finalizer is cleared

## Tasks / Subtasks

- [ ] Task 1: Create test fixtures (AC: 1, 3)
  - [ ] 1.1: Create `test/passwordpolicy/simple-password-policy.yaml` — a PasswordPolicy CR with `metadata.name: test-simple-password-policy`, using `authentication.role: policy-admin`, with HCL generating 20-char lowercase passwords (same policy pattern used across the codebase)
  - [ ] 1.2: Create `test/passwordpolicy/named-password-policy.yaml` — a PasswordPolicy CR with `metadata.name: test-named-pp-metadata` and `spec.name: test-named-password-policy`, using `authentication.role: policy-admin`, with HCL generating 10-char uppercase+digit passwords (different policy to distinguish from fixture 1)

- [ ] Task 2: Create integration test file (AC: 1, 2, 3, 4)
  - [ ] 2.1: Create `controllers/passwordpolicy_controller_test.go` with `//go:build integration` tag, package `controllers`, standard Ginkgo imports
  - [ ] 2.2: Add `Describe("PasswordPolicy controller")` with `timeout := 120 * time.Second`, `interval := 2 * time.Second`
  - [ ] 2.3: Add `Context("When creating a simple PasswordPolicy")` — load `simple-password-policy.yaml` via `decoder.GetPasswordPolicyInstance`, set namespace to `vaultAdminNamespaceName`, create it, poll for `ReconcileSuccessful=True`
  - [ ] 2.4: After reconcile success, read the policy from Vault via `vaultClient.Logical().Read("sys/policies/password/test-simple-password-policy")`, verify `secret != nil` and `secret.Data["policy"]` contains the HCL text (use `ContainSubstring` for the charset rule)
  - [ ] 2.5: Verify password generation by calling `vaultClient.Logical().Read("sys/policies/password/test-simple-password-policy/generate")`, assert `secret.Data["password"]` matches `^[a-z]{20}$` regex pattern
  - [ ] 2.6: Add `Context("When creating a PasswordPolicy with spec.name override")` — load `named-password-policy.yaml`, set namespace, create, poll for `ReconcileSuccessful=True`
  - [ ] 2.7: After reconcile success, read from Vault at `sys/policies/password/test-named-password-policy` (the `spec.name` path, NOT the `metadata.name`), verify `secret != nil` and `secret.Data["policy"]` contains the HCL text
  - [ ] 2.8: Verify password generation by calling `vaultClient.Logical().Read("sys/policies/password/test-named-password-policy/generate")`, assert the password matches the fixture's policy constraints
  - [ ] 2.9: Add `Context("When deleting PasswordPolicies")` — delete both PasswordPolicy CRs, use `Eventually` to poll for K8s deletion (NotFound error), then verify Vault read returns nil for both paths
  - [ ] 2.10: Verify the finalizer was cleared by confirming deletion completes (the `Eventually` for NotFound confirms this — if the finalizer was stuck, the object would remain)

- [ ] Task 3: End-to-end verification (AC: 1, 2, 3, 4)
  - [ ] 3.1: Run `make integration` and verify the new PasswordPolicy tests pass alongside all existing tests
  - [ ] 3.2: Verify no regressions in other tests that use PasswordPolicy as a dependency (VaultSecret, RandomSecret, DatabaseSecretEngineStaticRole tests all create PasswordPolicy CRs)

## Dev Notes

### PasswordPolicy Is a Standard VaultResource — Simplest Reconcile Flow

PasswordPolicy uses the standard `VaultResource` reconciler with the simplest possible implementation:
1. `prepareContext()` enriches context
2. `NewVaultResource(&r.ReconcilerBase, instance)` creates the standard reconciler
3. `VaultResource.Reconcile()` → `CreateOrUpdate()` → reads from Vault, calls `IsEquivalentToDesiredState`, writes only if different
4. `ManageOutcome()` sets `ReconcileSuccessful` condition with `ObservedGeneration`
5. On delete: `manageCleanUpLogic()` removes the policy from Vault (`IsDeletable()` returns `true`)

Key simplifications vs. Policy type:
- **No `spec.type` field** — always uses `sys/policies/password/<name>` path
- **No `PrepareInternalValues`** — returns nil (no-op, no accessor resolution)
- **Simple `IsEquivalentToDesiredState`** — `reflect.DeepEqual(GetPayload(), payload)` with no field remapping
- **No webhook validation** — `ValidateUpdate` returns nil (no immutable field protection needed, no `spec.path`)

[Source: api/v1alpha1/passwordpolicy_types.go]
[Source: controllers/passwordpolicy_controller.go]

### Vault Path and Payload

- **Path:** `sys/policies/password/<name>` where `<name>` is `spec.name` (if set) or `metadata.name`
- **Write payload:** `{"policy": "<hcl>"}`
- **Read response:** `{"data": {"policy": "<hcl>"}}`
- **Generate endpoint:** `GET sys/policies/password/<name>/generate` → `{"data": {"password": "<generated>"}}`

The generate endpoint is unique to PasswordPolicy among all types in the operator. It provides functional verification that the policy works, not just that it was written.

[Source: api/v1alpha1/passwordpolicy_types.go#L38-L47 — GetPath, GetPayload]

### IsEquivalentToDesiredState — Known Extra-Field Behavior

`IsEquivalentToDesiredState` uses raw `reflect.DeepEqual(GetPayload(), payload)`. If Vault returns extra fields in the read response beyond `{"policy": "<hcl>"}`, the comparison returns `false`, causing a write on every reconcile. This is documented tech debt tracked in Story 7-4. It does not affect test correctness.

[Source: api/v1alpha1/passwordpolicy_types.go#L49-L51]
[Source: api/v1alpha1/passwordpolicy_test.go#L98-L116 — TestPasswordPolicyIsEquivalentExtraFieldsReturnsFalse]

### Vault Auth Setup — policy-admin Role

The `policy-admin` Kubernetes auth role is configured by Vault's init sidecar container:
```
vault write auth/kubernetes/role/policy-admin \
  bound_service_account_names=default \
  bound_service_account_namespaces=vault-admin \
  policies=vault-admin ttl=1h
```

The `vault-admin` policy grants full `/*` access. All PasswordPolicy fixtures use `authentication.role: policy-admin` and must be created in the `vault-admin` namespace.

[Source: integration/vault-values.yaml]

### Controller Registration — Already Done

The PasswordPolicy controller is already registered in `suite_integration_test.go`:
```go
err = (&PasswordPolicyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "PasswordPolicy")}).SetupWithManager(mgr)
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L169-L170]

### Decoder — GetPasswordPolicyInstance Already Exists

`GetPasswordPolicyInstance` exists in `controllers/controllertestutils/decoder.go` at lines 39-52. No decoder changes needed.

[Source: controllers/controllertestutils/decoder.go#L39-L52]

### PasswordPolicy Is Already Used as a Dependency in Other Tests

PasswordPolicy CRs are created as prerequisites in:
- `vaultsecret_controller_test.go` — uses `test/password-policy.yaml` (`simple-password-policy`)
- `vaultsecret_controller_v2_test.go` — uses `test/randomsecret/v2/00-passwordpolicy-simple-password-policy-v2.yaml`
- `randomsecret_controller_test.go` — uses same v2 fixture (`simple-password-policy-v2`)
- `databasesecretenginestaticrole_controller_test.go` — uses `test/databasesecretengine/password-policy.yaml` (`postgresql-password-policy`)

These tests prove PasswordPolicy CRs can be created and reconciled. The dedicated PasswordPolicy test adds:
1. **Direct Vault state verification** (reading the policy from Vault and checking content)
2. **Functional verification via generate endpoint** (proving the policy produces valid passwords)
3. **`spec.name` override verification** (ensuring the Vault path uses `spec.name` when set)
4. **Delete cleanup verification** (confirming Vault policy removal)

### Test Fixture Design

**Fixture 1: `test/passwordpolicy/simple-password-policy.yaml`**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PasswordPolicy
metadata:
  name: test-simple-password-policy
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  passwordPolicy: |
    length = 20
    rule "charset" {
      charset = "abcdefghijklmnopqrstuvwxyz"
    }
```

Uses `metadata.name` for the Vault path → `sys/policies/password/test-simple-password-policy`. Generates 20-char lowercase passwords. Uses unique name prefix `test-` to avoid collisions with existing fixtures (`simple-password-policy`, `simple-password-policy-v2`).

**Fixture 2: `test/passwordpolicy/named-password-policy.yaml`**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PasswordPolicy
metadata:
  name: test-named-pp-metadata
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: test-named-password-policy
  passwordPolicy: |
    length = 10
    rule "charset" {
      charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    }
```

Uses `spec.name` override → Vault path is `sys/policies/password/test-named-password-policy` (not `test-named-pp-metadata`). Different policy constraints (10-char uppercase+digit) to distinguish from fixture 1.

### Verifying Vault State

**Simple policy verification:**
```go
secret, err := vaultClient.Logical().Read("sys/policies/password/test-simple-password-policy")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data["policy"]).To(ContainSubstring("abcdefghijklmnopqrstuvwxyz"))
```

**Password generation verification:**
```go
genSecret, err := vaultClient.Logical().Read("sys/policies/password/test-simple-password-policy/generate")
Expect(err).To(BeNil())
Expect(genSecret).NotTo(BeNil())
password := genSecret.Data["password"].(string)
Expect(password).To(MatchRegexp("^[a-z]{20}$"))
```

**Named policy verification:**
```go
secret, err := vaultClient.Logical().Read("sys/policies/password/test-named-password-policy")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data["policy"]).To(ContainSubstring("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
```

**Delete verification:**
```go
Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("sys/policies/password/test-simple-password-policy")
    return secret == nil
}, timeout, interval).Should(BeTrue())
```

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `test/passwordpolicy/simple-password-policy.yaml` | New | Simple HCL password policy fixture (20-char lowercase) |
| 2 | `test/passwordpolicy/named-password-policy.yaml` | New | Password policy with `spec.name` override (10-char uppercase+digit) |
| 3 | `controllers/passwordpolicy_controller_test.go` | New | Integration test file — create, verify Vault state, generate password, spec.name override, delete |

No changes to decoder, suite setup, controllers, or types.

### No `make manifests generate` Needed

This story only adds integration test files and YAML fixtures. No CRD types, controllers, or webhooks are changed.

### Import Requirements for passwordpolicy_controller_test.go

```go
import (
    "regexp"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)
```

All are already indirect dependencies — no `go get` needed.

### Test Structure

```
Describe("PasswordPolicy controller")
  Context("When creating a simple PasswordPolicy")
    It("Should create the password policy in Vault and generate valid passwords")
      - Load fixture, set namespace, create
      - Eventually poll for ReconcileSuccessful=True
      - Read from Vault at sys/policies/password/<name>, verify "policy" key contains HCL
      - Call generate endpoint, verify password matches ^[a-z]{20}$ pattern
  Context("When creating a PasswordPolicy with spec.name override")
    It("Should create the policy at the spec.name path in Vault")
      - Load fixture, set namespace, create
      - Eventually poll for ReconcileSuccessful=True
      - Read from Vault at sys/policies/password/<spec.name>, verify "policy" key
      - Call generate endpoint, verify password matches expected pattern
  Context("When deleting PasswordPolicies")
    It("Should remove password policies from Vault")
      - Delete both PasswordPolicy CRs
      - Eventually poll for K8s NotFound
      - Verify Vault read returns nil for both paths
```

### Name Collision Prevention

Fixture names use the `test-` prefix to avoid collisions with PasswordPolicy CRs created by other integration tests:
- `simple-password-policy` — used by `vaultsecret_controller_test.go`
- `simple-password-policy-v2` — used by `randomsecret_controller_test.go` and `vaultsecret_controller_v2_test.go`
- `postgresql-password-policy` — used by `databasesecretenginestaticrole_controller_test.go`
- `test-simple-password-policy` — this story (fixture 1)
- `test-named-password-policy` — this story (fixture 2, via spec.name)

### Risk Considerations

- **Vault response extra fields:** Vault's read response for password policies may include fields beyond `"policy"`. Use `ContainSubstring` on `secret.Data["policy"]` rather than exact match to be resilient to whitespace normalization. The `secret.Data["policy"]` access should work cleanly because that's the documented response key.
- **Password generation determinism:** The generate endpoint returns a random password each call. Verify the pattern (length + character set) not the exact value.
- **Test ordering:** The create Contexts must run before the delete Context. Ginkgo v2 runs Contexts sequentially within a Describe by default.
- **`spec.name` validation pattern:** The `spec.name` field has kubebuilder pattern validation `[a-z0-9]([-a-z0-9]*[a-z0-9])?`. Both fixture names comply with this pattern.

### Previous Story Intelligence

**From Story 3.1 (Policy integration tests):**
- Established the Epic 3 pattern: create fixture → create CR → poll ReconcileSuccessful → verify Vault state → delete → verify Vault cleanup
- Used `ContainSubstring` for policy text verification (resilient to whitespace normalization)
- Tested both Vault API path variants (legacy `sys/policy/<name>` and typed `sys/policies/acl/<name>`)
- Verified accessor placeholder resolution via `PrepareInternalValues` — not applicable to PasswordPolicy (no-op)
- Two fixtures: one simple, one with specific behavior to test
- Entity test (`entity_controller_test.go`) remains the simplest standalone test pattern reference

**From Epic 2 Retrospective:**
- "Pattern-first investment paid off again" — follow the established create/verify/delete pattern
- "Epic 3 scope is simpler than Epic 2: create/delete lifecycle tests for 4 foundation types"
- Model quality note: prefer Opus-class models for implementation

### Git Intelligence (Recent Commits)

```
a1ccf6e Complete Epic 2 retrospective
8209c81 Add update scenarios to PKI and DatabaseSecretEngineStaticRole integration tests (Story 2.4)
8560dc8 Add update scenarios to Entity and EntityAlias integration tests (Story 2.3)
ae098e2 Add update scenarios to RandomSecret integration tests (Story 2.2)
a24cb2f Add update scenarios to VaultSecret integration tests (Story 2.1)
452285a Stabilize integration test infrastructure (Story 2.0)
```

Codebase is clean post-Epic 2. No pending changes affect this story.

### Project Structure Notes

- Test file goes in `controllers/passwordpolicy_controller_test.go` (standard controller test location)
- Test fixtures go in `test/passwordpolicy/` directory (follows `test/<feature>/` pattern)
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/passwordpolicy_types.go] — PasswordPolicy VaultObject implementation, GetPath, GetPayload, IsEquivalentToDesiredState
- [Source: api/v1alpha1/passwordpolicy_types.go#L73-L94] — PasswordPolicySpec fields
- [Source: api/v1alpha1/passwordpolicy_webhook.go] — Webhook (no validation rules)
- [Source: api/v1alpha1/passwordpolicy_test.go] — Unit tests (path, payload, equivalence, extra fields)
- [Source: controllers/passwordpolicy_controller.go] — Controller (standard VaultResource)
- [Source: controllers/suite_integration_test.go#L169-L170] — PasswordPolicy controller registration
- [Source: controllers/controllertestutils/decoder.go#L39-L52] — GetPasswordPolicyInstance decoder method
- [Source: controllers/entity_controller_test.go] — Simplest standalone integration test pattern reference
- [Source: controllers/vaultsecret_controller_test.go#L31-L49] — PasswordPolicy used as dependency (create + poll pattern)
- [Source: controllers/randomsecret_controller_test.go#L31-L49] — PasswordPolicy used as dependency
- [Source: api/v1alpha1/randomsecret_types.go#L241] — Password generate endpoint usage
- [Source: test/password-policy.yaml] — Existing fixture (different name, used by VaultSecret test)
- [Source: _bmad-output/planning-artifacts/epics.md#L373-L387] — Story 3.2 epic definition
- [Source: _bmad-output/implementation-artifacts/3-1-integration-tests-for-policy-type.md] — Previous story (pattern reference)
- [Source: _bmad-output/implementation-artifacts/epic-2-retro-2026-04-17.md#L126-L143] — Epic 3 readiness assessment
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy
- [Source: _bmad-output/project-context.md#L143-L149] — Integration test pattern

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### Change Log

### File List
