# Story 3.1: Integration Tests for Policy Type

Status: done

## Story

As an operator developer,
I want integration tests for the Policy type covering create, reconcile success, and delete with Vault cleanup,
So that the most fundamental Vault resource type has end-to-end test coverage.

## Acceptance Criteria

1. **Given** a Policy CR is created in the test namespace with a valid HCL policy **When** the reconciler processes it **Then** the policy exists in Vault at the correct path and ReconcileSuccessful=True

2. **Given** a Policy CR with `${auth/kubernetes/@accessor}` placeholders is created **When** the reconciler processes it **Then** `PrepareInternalValues` resolves the accessor placeholders and the resolved policy exists in Vault

3. **Given** a successfully reconciled Policy CR is deleted **When** the reconciler processes the deletion **Then** the policy is removed from Vault and the finalizer is cleared

## Tasks / Subtasks

- [x] Task 1: Create test fixtures (AC: 1, 2)
  - [x] 1.1: Create `test/policy/simple-policy.yaml` — a minimal Policy CR with a simple HCL policy (no accessor placeholders), using `authentication.role: policy-admin`, no `spec.type` field (legacy `sys/policy/<name>` path)
  - [x] 1.2: Create `test/policy/acl-policy-with-accessor.yaml` — a Policy CR with `spec.type: acl` and HCL containing `${auth/kubernetes/@accessor}` placeholders, using `authentication.role: policy-admin` (tests both typed path and `PrepareInternalValues`)

- [x] Task 2: Create integration test file (AC: 1, 2, 3)
  - [x] 2.1: Create `controllers/policy_controller_test.go` with `//go:build integration` tag, package `controllers`, standard Ginkgo imports
  - [x] 2.2: Add `Describe("Policy controller")` with `timeout := 120s`, `interval := 2s`
  - [x] 2.3: Add `Context("When creating a simple Policy")` — loads `simple-policy.yaml` via `decoder.GetPolicyInstance`, sets namespace to `vaultAdminNamespaceName`, creates it, polls for `ReconcileSuccessful=True`
  - [x] 2.4: After reconcile success, read the policy from Vault via `vaultClient.Logical().Read("sys/policy/<name>")`, verify `secret != nil` and `secret.Data["rules"]` matches the HCL text from the fixture
  - [x] 2.5: Add `Context("When creating a Policy with accessor placeholder")` — loads `acl-policy-with-accessor.yaml`, sets namespace to `vaultAdminNamespaceName`, creates it, polls for `ReconcileSuccessful=True`
  - [x] 2.6: After reconcile success, read the policy from Vault via `vaultClient.Logical().Read("sys/policies/acl/<name>")`, verify the policy text does NOT contain `${auth/kubernetes/@accessor}` (placeholder was resolved) and DOES contain an actual accessor string (e.g., `auth_kubernetes_`)
  - [x] 2.7: Add `Context("When deleting Policies")` — delete both Policy CRs, use `Eventually` to poll for deletion from K8s (NotFound error), then verify the policies no longer exist in Vault (read returns nil or 404)
  - [x] 2.8: Verify the finalizer was cleared by checking the deletion completes (the `Eventually` for NotFound implicitly confirms this — if the finalizer was stuck, the object would remain)

- [x] Task 3: End-to-end verification (AC: 1, 2, 3)
  - [x] 3.1: Run `make integration` and verify the new Policy tests pass alongside all existing tests
  - [x] 3.2: Verify no regressions in other tests that use Policy as a dependency (VaultSecret, RandomSecret, PKI, Database tests all create Policy CRs)

### Review Findings

- [x] [Review][Patch] Ensure Policy resources are cleaned up even when an earlier `Ordered` spec fails [controllers/policy_controller_test.go:27-34] — added `AfterAll` cleanup guard
- [x] [Review][Patch] Strengthen Vault policy assertions so malformed HCL cannot pass on a single substring match [controllers/policy_controller_test.go:67-68,105-106] — added capabilities assertion and safe type assertion

## Dev Notes

### Policy Is a Standard VaultResource Type — Simple Reconcile Flow

Policy uses the standard `VaultResource` reconciler:
1. `prepareContext()` enriches context with kubeClient, vaultClient, etc.
2. `NewVaultResource(&r.ReconcilerBase, instance)` creates the standard reconciler
3. `VaultResource.Reconcile()` → `CreateOrUpdate()` → reads from Vault, calls `IsEquivalentToDesiredState()`, writes only if different
4. `ManageOutcome()` sets `ReconcileSuccessful` condition with `ObservedGeneration`
5. On delete: `manageCleanUpLogic()` removes the policy from Vault (because `IsDeletable()` returns `true`)

[Source: controllers/policy_controller.go — standard VaultResource pattern]

### Policy Has NO `spec.path` Field

Unlike most types, `Policy` does not have a `spec.path` field. The Vault path is computed in `GetPath()`:
- Without `spec.type`: `sys/policy/<name>` (legacy Vault path)
- With `spec.type: acl`: `sys/policies/acl/<name>` (modern Vault path)

Where `<name>` is `spec.name` (if set) or `metadata.name`.

The webhook has no immutability checks — `ValidateUpdate` returns nil. This is correct because there's no `spec.path` to protect.

[Source: api/v1alpha1/policy_types.go#L43-L53 — GetPath]
[Source: api/v1alpha1/policy_webhook.go#L61-L65 — ValidateUpdate is a no-op]

### IsEquivalentToDesiredState Has Custom Logic

Policy's `IsEquivalentToDesiredState` is not a simple `reflect.DeepEqual(toMap(), payload)`. It has conditional field remapping:

```go
func (d *Policy) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.GetPayload()                                    // {"policy": "<hcl>"}
    desiredState["name"] = map[bool]string{true: d.Spec.Name, false: d.Name}[d.Spec.Name != ""]
    if d.Spec.Type == "" {
        desiredState["rules"] = desiredState["policy"]   // legacy: rename "policy" → "rules"
        delete(desiredState, "policy")
    }
    return reflect.DeepEqual(desiredState, payload)
}
```

- **Legacy path** (`spec.type` empty): Vault returns `{"name": "...", "rules": "..."}` — so the desired state uses `rules` key
- **Typed ACL path** (`spec.type: acl`): Vault returns `{"name": "...", "policy": "..."}` — so the desired state keeps `policy` key

This means the Vault read response shape differs based on which API path is used. The test should verify both paths.

[Source: api/v1alpha1/policy_types.go#L60-L67]

### PrepareInternalValues — Accessor Placeholder Resolution

When the HCL policy text contains `${auth/<engine-path>/@accessor}` placeholders, `PrepareInternalValues` resolves them:

1. Regex check for `\${[^}]+}` — fast-path returns nil if no placeholders
2. Reads `sys/auth` from Vault to get all auth engine mounts with their accessors
3. For each mount, replaces `${auth/<path>/@accessor}` with the actual accessor string (e.g., `auth_kubernetes_abc123`)
4. On read error: logs warning and returns nil (does NOT fail reconcile — placeholder remains unresolved)
5. On nil secret: returns an error (fails reconcile)

**Critical:** The accessor resolution modifies `d.Spec.Policy` in-place on the Go struct. The resolved text is what gets written to Vault. The original CR spec in Kubernetes still contains the placeholder. Subsequent reconciles re-resolve the placeholder each time.

[Source: api/v1alpha1/policy_types.go#L78-L108]

### Vault API Response Shapes

**Legacy path** `GET sys/policy/<name>`:
```json
{
  "data": {
    "name": "my-policy",
    "rules": "path \"secret/*\" {\n  capabilities = [\"read\"]\n}"
  }
}
```

**Typed ACL path** `GET sys/policies/acl/<name>`:
```json
{
  "data": {
    "name": "my-policy",
    "policy": "path \"secret/*\" {\n  capabilities = [\"read\"]\n}"
  }
}
```

Access the data via `secret.Data["rules"]` or `secret.Data["policy"]` respectively.

### Vault Auth Setup — policy-admin Role

The `policy-admin` Kubernetes auth role is configured by Vault's init sidecar container (in `integration/vault-values.yaml`):
```bash
vault write auth/kubernetes/role/policy-admin \
  bound_service_account_names=default \
  bound_service_account_namespaces=vault-admin \
  policies=vault-admin ttl=1h
```

The `vault-admin` policy grants full `/*` access. All Policy fixtures use `authentication.role: policy-admin` and must be created in the `vault-admin` namespace to authenticate.

[Source: integration/vault-values.yaml#L164-L167]

### Controller Registration — Already Done

The Policy controller is already registered in `suite_integration_test.go`:
```go
err = (&PolicyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "Policy")}).SetupWithManager(mgr)
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L133-L134]

### Decoder — GetPolicyInstance Already Exists

`GetPolicyInstance` exists in `controllers/controllertestutils/decoder.go` at lines 70-83. No decoder changes needed.

[Source: controllers/controllertestutils/decoder.go#L70-L83]

### Policy Is Already Used as a Dependency in Other Tests

Policy CRs are created as prerequisites in:
- `vaultsecret_controller_test.go` — creates `kv-engine-admin`, `secret-writer`, `secret-reader` policies
- `randomsecret_controller_test.go` — creates v2 variants
- `pkisecretengine_controller_test.go` — creates `pki-engine-admin`
- `databasesecretenginestaticrole_controller_test.go` — creates `database-engine-admin`

These tests already prove Policy CRs can be created and reconciled. The dedicated Policy test adds:
1. **Direct Vault state verification** (reading the policy back from Vault and checking content)
2. **Delete cleanup verification** (confirming Vault policy removal)
3. **Accessor placeholder verification** (confirming `PrepareInternalValues` works)
4. **Both API path variants** (legacy `sys/policy/` and typed `sys/policies/acl/`)

### Test Fixture Design

**Fixture 1: `test/policy/simple-policy.yaml`** — Tests the legacy `sys/policy/<name>` path:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Policy
metadata:
  name: test-simple-policy
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  policy: |
    path "secret/data/test/*" {
      capabilities = ["create", "read", "update", "delete", "list"]
    }
```

No `spec.type` → uses `sys/policy/test-simple-policy`. No accessor placeholders → `PrepareInternalValues` is a no-op.

**Fixture 2: `test/policy/acl-policy-with-accessor.yaml`** — Tests the typed `sys/policies/acl/<name>` path AND accessor resolution:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Policy
metadata:
  name: test-acl-policy
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: acl
  policy: |
    path "{{identity.entity.aliases.${auth/kubernetes/@accessor}.metadata.service_account_namespace}}/kv/*" {
      capabilities = ["read", "list"]
    }
```

`spec.type: acl` → uses `sys/policies/acl/test-acl-policy`. Contains `${auth/kubernetes/@accessor}` → `PrepareInternalValues` resolves it.

### Verifying Vault State

**Simple policy verification:**
```go
secret, err := vaultClient.Logical().Read("sys/policy/test-simple-policy")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data["rules"]).To(ContainSubstring("secret/data/test/*"))
```

**ACL policy with accessor verification:**
```go
secret, err := vaultClient.Logical().Read("sys/policies/acl/test-acl-policy")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
policyText := secret.Data["policy"].(string)
Expect(policyText).NotTo(ContainSubstring("${auth/kubernetes/@accessor}"))
Expect(policyText).To(ContainSubstring("auth_kubernetes_"))
```

The accessor value follows the pattern `auth_kubernetes_<hash>`. Checking for `auth_kubernetes_` prefix confirms resolution without depending on the exact hash.

**Delete verification:**
```go
Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("sys/policy/test-simple-policy")
    return secret == nil
}, timeout, interval).Should(BeTrue())
```

### Copy-Paste Bug in policy_types.go (Non-Blocking)

Line 37 has a copy-paste error:
```go
var _ vaultutils.ConditionsAware = &PKISecretEngineRole{}  // should be &Policy{}
```

This doesn't affect functionality (Policy still implements ConditionsAware via `GetConditions`/`SetConditions`) but the compile-time check asserts the wrong type. **Do NOT fix this in this story** — it's unrelated tech debt. Document for a future cleanup.

[Source: api/v1alpha1/policy_types.go#L37]

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `test/policy/simple-policy.yaml` | New | Simple HCL policy fixture (no accessor, no type) |
| 2 | `test/policy/acl-policy-with-accessor.yaml` | New | Typed ACL policy fixture with accessor placeholder |
| 3 | `controllers/policy_controller_test.go` | New | Integration test file — create, verify Vault state, delete |

No changes to decoder, suite setup, controllers, or types.

### No `make manifests generate` Needed

This story only adds integration test files and YAML fixtures. No CRD types, controllers, or webhooks are changed.

### Import Requirements for policy_controller_test.go

```go
import (
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
Describe("Policy controller")
  Context("When creating a simple Policy")
    It("Should create the policy in Vault at sys/policy/<name>")
      - Load fixture, set namespace, create
      - Eventually poll for ReconcileSuccessful=True
      - Read from Vault via sys/policy/<name>, verify "rules" key contains HCL
  Context("When creating a Policy with accessor placeholder and type: acl")
    It("Should resolve accessor and create at sys/policies/acl/<name>")
      - Load fixture, set namespace, create
      - Eventually poll for ReconcileSuccessful=True
      - Read from Vault via sys/policies/acl/<name>, verify "policy" key
      - Verify accessor placeholder was resolved
  Context("When deleting Policies")
    It("Should remove policies from Vault")
      - Delete both Policy CRs
      - Eventually poll for K8s NotFound
      - Verify Vault read returns nil for both paths
```

### Risk Considerations

- **Accessor placeholder resolution depends on sys/auth read:** If the Vault auth engine mount is not yet initialized when the test runs, `PrepareInternalValues` will fail. The `policy-admin` role requires Kubernetes auth to be enabled, which happens in Vault init. The integration test infrastructure should have this ready before tests run.
- **Policy text whitespace:** Vault may normalize whitespace in HCL policies. Use `ContainSubstring` for content verification, not exact string matching.
- **Test ordering:** The two create Contexts must run before the delete Context. Ginkgo v2 runs Contexts sequentially within a Describe by default, so this is safe.
- **Name collisions:** The test fixtures use unique names (`test-simple-policy`, `test-acl-policy`) that don't conflict with policies created by other integration tests (`kv-engine-admin`, `secret-writer`, etc.).

### Previous Story Intelligence

**From Epic 2 Retrospective:**
- "Pattern-first investment paid off again" — use established patterns from Epic 2 (create → poll → verify → delete → verify Vault cleanup)
- "Update tests are the highest-value integration tests" — this story focuses on create/delete per the epic scope, but update testing could be added in a future enhancement
- "Epic 3 scope is simpler than Epic 2: create/delete lifecycle tests for 4 foundation types"
- "Story 3.4 (AuthEngineMount) may need a GetAuthEngineMountInstance decoder method and new test fixture — verify at story creation time"

**From Story 2.4 (last completed story):**
- Established pattern for verifying Vault state via `vaultClient.Logical().Read()`
- Used `ContainSubstring` for flexible assertion on Vault response data
- The Entity test (`entity_controller_test.go`) is the simplest standalone pattern reference — create, poll, delete in ~60 lines

**From Story 2.0:**
- Integration test infrastructure is stable: idempotent Kind setup, namespace create-or-get, vendored manifests
- The `vaultClient` variable is available in all integration tests (set up in BeforeSuite)

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

- Test file goes in `controllers/policy_controller_test.go` (standard controller test location)
- Test fixtures go in `test/policy/` directory (follows `test/<feature>/` pattern used by `test/pkisecretengine/`, `test/policy/`, etc.)
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/policy_types.go#L36-L108] — Policy VaultObject implementation, GetPath, GetPayload, IsEquivalentToDesiredState, PrepareInternalValues
- [Source: api/v1alpha1/policy_types.go#L119-L145] — PolicySpec fields
- [Source: api/v1alpha1/policy_webhook.go] — Webhook (no validation rules)
- [Source: controllers/policy_controller.go] — Controller (standard VaultResource)
- [Source: controllers/suite_integration_test.go#L133-L134] — Policy controller registration
- [Source: controllers/controllertestutils/decoder.go#L70-L83] — GetPolicyInstance decoder method
- [Source: controllers/entity_controller_test.go] — Simplest standalone integration test pattern reference
- [Source: controllers/vaultsecret_controller_test.go#L56-L102] — Policy used as dependency (create + poll pattern)
- [Source: integration/vault-values.yaml#L164-L167] — policy-admin Vault auth role setup
- [Source: _bmad-output/planning-artifacts/epics.md#L355-L371] — Story 3.1 epic definition
- [Source: _bmad-output/implementation-artifacts/epic-2-retro-2026-04-17.md#L126-L143] — Epic 3 readiness assessment
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy
- [Source: _bmad-output/project-context.md#L143-L149] — Integration test pattern
- [Source: _bmad-output/implementation-artifacts/2-4-add-update-scenarios-to-pki-and-databasesecretenginestaticrole-integration-tests.md] — Previous story (pattern reference)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (Cursor)

### Debug Log References

None — clean implementation with no failures.

### Completion Notes List

- Created two test fixtures covering both Policy API paths: legacy `sys/policy/<name>` (no type) and modern `sys/policies/acl/<name>` (type: acl)
- ACL fixture includes `${auth/kubernetes/@accessor}` placeholder to test `PrepareInternalValues` accessor resolution
- Integration test verifies all 3 acceptance criteria: create with Vault state verification, accessor placeholder resolution, and delete with Vault cleanup
- Test structure follows established patterns from Entity test (simplest standalone reference) with Describe/Context/It hierarchy
- Both Vault API response shapes verified: `rules` key for legacy path, `policy` key for typed ACL path
- Delete context confirms finalizer cleanup by polling for K8s NotFound, then verifying Vault state is clean
- All existing integration tests continue to pass with zero regressions

### Change Log

- 2026-04-18: Implemented Story 3.1 — created test fixtures and integration test for Policy type

### File List

- `test/policy/simple-policy.yaml` (new) — Simple HCL policy fixture, no accessor, no type
- `test/policy/acl-policy-with-accessor.yaml` (new) — Typed ACL policy fixture with accessor placeholder
- `controllers/policy_controller_test.go` (new) — Integration test: create, verify Vault state, delete
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified) — Story status updates
- `_bmad-output/implementation-artifacts/3-1-integration-tests-for-policy-type.md` (modified) — Story file updates
