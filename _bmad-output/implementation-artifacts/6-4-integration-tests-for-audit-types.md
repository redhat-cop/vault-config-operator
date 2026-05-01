# Story 6.4: Integration Tests for Audit Types

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration tests for Audit and AuditRequestHeader covering create, reconcile success, Vault state verification, update, and delete,
So that audit device management and request header configuration are verified end-to-end.

## Scope Decision: Tier Classification

Per the project's three-tier integration test rule:

| Type | Dependency | Classification | Action |
|------|-----------|---------------|--------|
| Audit | Vault sys/audit API (internal) | **Tier 1: Already available** | **Test** — no external service |
| AuditRequestHeader | Vault sys/config/auditing API (internal) | **Tier 1: Already available** | **Test** — no external service |

No new infrastructure needed. Both types interact with Vault's internal system APIs.

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

## Acceptance Criteria

1. **Given** an Audit CR for a file audit device is created **When** the reconciler processes it **Then** the audit device is enabled in Vault (visible in `Sys().ListAudit()`) and ReconcileSuccessful=True

2. **Given** an AuditRequestHeader CR is created with `hmac: true` **When** the reconciler processes it **Then** the header is configured in Vault at `sys/config/auditing/request-headers/{name}` and ReconcileSuccessful=True

3. **Given** the Audit CR spec is updated (e.g., description changed) **When** the reconciler processes the update **Then** the Vault audit device reflects the change (disable + re-enable) and ObservedGeneration increases

4. **Given** the AuditRequestHeader CR spec is updated (e.g., hmac changed from true to false) **When** the reconciler processes the update **Then** the Vault request header config reflects the change and ObservedGeneration increases

5. **Given** the AuditRequestHeader CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the header is removed from Vault and the CR is removed from K8s

6. **Given** the Audit CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the audit device is disabled in Vault and the CR is removed from K8s

## Tasks / Subtasks

- [ ] Task 1: Create test fixtures (AC: 1, 2)
  - [ ] 1.1: Create `test/audit/audit.yaml` — file audit device writing to stdout
  - [ ] 1.2: Create `test/audit/auditrequestheader.yaml` — request header with hmac=true

- [ ] Task 2: Register controllers in suite_integration_test.go (AC: 1, 2)
  - [ ] 2.1: Add `AuditReconciler` registration
  - [ ] 2.2: Add `AuditRequestHeaderReconciler` registration

- [ ] Task 3: Add decoder methods (AC: 1, 2)
  - [ ] 3.1: Add `GetAuditInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 3.2: Add `GetAuditRequestHeaderInstance` to `controllers/controllertestutils/decoder.go`

- [ ] Task 4: Create integration test file (AC: 1, 2, 3, 4, 5, 6)
  - [ ] 4.1: Create `controllers/audit_controller_test.go` with `//go:build integration` tag
  - [ ] 4.2: Add context for Audit creation — create, poll ReconcileSuccessful=True, verify Vault state via Sys().ListAudit()
  - [ ] 4.3: Add context for AuditRequestHeader creation — create, poll ReconcileSuccessful=True, verify Vault state via Logical().Read()
  - [ ] 4.4: Add context for Audit update — change description, verify Vault reflects change, verify ObservedGeneration increased
  - [ ] 4.5: Add context for AuditRequestHeader update — change hmac, verify Vault reflects change, verify ObservedGeneration increased
  - [ ] 4.6: Add deletion context — delete AuditRequestHeader (verify Vault cleanup), delete Audit (verify device disabled)

- [ ] Task 5: End-to-end verification (AC: 1, 2, 3, 4, 5, 6)
  - [ ] 5.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [ ] 5.2: Verify no regressions — existing tests unaffected

## Dev Notes

### Two Different Reconciler Variants — CRITICAL

These two types use **different** reconciler variants:

| Type | Reconciler | Vault API | Create Method | Delete Method |
|------|-----------|-----------|---------------|---------------|
| **Audit** | `VaultAuditResource` (`NewVaultAuditResource`) | `Sys()` API — EnableAuditWithOptions / ListAudit / DisableAudit | Device enable | Device disable |
| **AuditRequestHeader** | `VaultResource` (`NewVaultResource`) | `Logical()` API — standard Read/Write/Delete on path | Logical Write | Logical Delete |

**Audit** implements **both** `VaultObject` and `VaultAuditObject` interfaces.
**AuditRequestHeader** implements only `VaultObject` (standard VaultResource flow).

[Source: controllers/audit_controller.go — NewVaultAuditResource]
[Source: controllers/auditrequestheader_controller.go — NewVaultResource]

### Deletability — Both IsDeletable=true

| Type | IsDeletable | Behavior |
|------|-------------|----------|
| Audit | **true** | Finalizer added. CR deletion disables audit device in Vault via `Sys().DisableAudit()`. |
| AuditRequestHeader | **true** | Finalizer added. CR deletion removes header config from Vault via `Logical().Delete()`. |

Both types require Vault cleanup verification after K8s deletion.

[Source: api/v1alpha1/audit_types.go#L166-L168 — IsDeletable=true]
[Source: api/v1alpha1/auditrequestheader_types.go#L126-L128 — IsDeletable=true]

### No Dependency Between Types

Audit (audit device) and AuditRequestHeader (header config) are independent Vault system APIs. No creation ordering required. However, for logical grouping, test Audit first then AuditRequestHeader.

### Vault API Paths

| Type | GetPath() | Vault API Path |
|------|-----------|---------------|
| Audit | `sys/audit/{spec.path}` | Named audit device path — `spec.path` is the device name |
| AuditRequestHeader | `sys/config/auditing/request-headers/{spec.name}` | Named header — uses `spec.name`, NOT `metadata.name` |

**IMPORTANT:** AuditRequestHeader uses `d.Spec.Name` in GetPath(), not `metadata.name`. The test fixture must set `spec.name` explicitly.

[Source: api/v1alpha1/audit_types.go#L117-L119 — GetPath]
[Source: api/v1alpha1/auditrequestheader_types.go#L105-L107 — GetPath]

### Audit — GetPayload and IsEquivalentToDesiredState

**GetPayload:**
```go
func (d *Audit) GetPayload() map[string]interface{} {
    payload := map[string]interface{}{
        "type":        d.Spec.Type,
        "description": d.Spec.Description,
        "local":       d.Spec.Local,
        "options":     d.Spec.Options,
    }
    return payload
}
```

**IsEquivalentToDesiredState:** Custom field-by-field comparison (NOT simple DeepEqual). Compares `type`, `description`, `local` directly, then compares `options` map key-by-key with length check.

[Source: api/v1alpha1/audit_types.go#L121-L148]

### Audit Update Mechanism — Disable + Re-Enable

Vault audit devices **cannot be updated in place**. The `VaultAuditEndpoint.CreateOrUpdate` flow:
1. Checks if device exists via `Sys().ListAudit()`
2. If exists, calls `IsEquivalentToDesired()` which reads current config from ListAudit and calls `IsEquivalentToDesiredState()`
3. If not equivalent: disables the device, then re-enables with new configuration

This means the update test must verify the new configuration appears after the disable/re-enable cycle.

[Source: api/v1alpha1/utils/vaultauditobject.go#L148-L183 — CreateOrUpdate]

### AuditRequestHeader — GetPayload and IsEquivalentToDesiredState

**GetPayload:**
```go
func (d *AuditRequestHeader) GetPayload() map[string]interface{} {
    return map[string]interface{}{
        "hmac": d.Spec.HMAC,
    }
}
```

**IsEquivalentToDesiredState:** Checks only the `hmac` field by type assertion — inherently ignores extra fields.

```go
func (d *AuditRequestHeader) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    if hmac, ok := payload["hmac"].(bool); ok {
        return hmac == d.Spec.HMAC
    }
    return false
}
```

[Source: api/v1alpha1/auditrequestheader_types.go#L109-L120]

### Webhooks — NONE Exist

Neither Audit nor AuditRequestHeader have webhook files. No `audit_webhook.go` or `auditrequestheader_webhook.go` exist. No webhook registration in `main.go` for these types. This means:
- No immutable `spec.path` validation at webhook level
- No defaulting webhooks
- No admission validation
- Updates can change any field without webhook rejection

[Source: main.go ENABLE_WEBHOOKS block — Audit/AuditRequestHeader absent]

### Verifying Vault State

**Audit verification via Sys().ListAudit():**
```go
audits, err := vaultClient.Sys().ListAudit()
Expect(err).To(BeNil())

auditDevice, exists := audits["test-audit/"]
Expect(exists).To(BeTrue(), "expected audit device 'test-audit' to exist")
Expect(auditDevice.Type).To(Equal("file"))
Expect(auditDevice.Options["file_path"]).To(Equal("stdout"))
```

**CRITICAL:** The `ListAudit()` map uses the audit path as key **with a trailing slash**. The test must append `/` to the path when looking up the device. The returned value is a `*vault.Audit` struct with `Type` (string), `Description` (string), `Local` (bool), `Options` (map[string]string), `Path` (string).

**AuditRequestHeader verification via Logical().Read():**
```go
secret, err := vaultClient.Logical().Read("sys/config/auditing/request-headers/X-Custom-Test-Header")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data).NotTo(BeNil())
```

**IMPORTANT — Nested Response Format:** Vault returns the request header config in a nested format: `{"header-name": {"hmac": true}}`. The `secret.Data` map key is the header name itself. To verify HMAC:
```go
headerData, ok := secret.Data["X-Custom-Test-Header"].(map[string]interface{})
Expect(ok).To(BeTrue(), "expected header data to be a map")
hmac, ok := headerData["hmac"].(bool)
Expect(ok).To(BeTrue(), "expected hmac to be a bool")
Expect(hmac).To(BeTrue())
```

**Audit delete verification (IsDeletable=true):**
```go
Expect(k8sIntegrationClient.Delete(ctx, auditInstance)).Should(Succeed())
lookupKey := types.NamespacedName{Name: auditInstance.Name, Namespace: auditInstance.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.Audit{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    audits, err := vaultClient.Sys().ListAudit()
    if err != nil {
        return false
    }
    _, exists := audits["test-audit/"]
    return !exists
}, timeout, interval).Should(BeTrue())
```

**AuditRequestHeader delete verification (IsDeletable=true):**
```go
Expect(k8sIntegrationClient.Delete(ctx, headerInstance)).Should(Succeed())
lookupKey := types.NamespacedName{Name: headerInstance.Name, Namespace: headerInstance.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.AuditRequestHeader{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("sys/config/auditing/request-headers/X-Custom-Test-Header")
    if err != nil {
        return false
    }
    return secret == nil
}, timeout, interval).Should(BeTrue())
```

### Test Structure

```
Describe("Audit controllers", Ordered)
  var auditInstance *redhatcopv1alpha1.Audit
  var headerInstance *redhatcopv1alpha1.AuditRequestHeader

  AfterAll: best-effort delete both instances:
    headerInstance → auditInstance

  Context("When creating an Audit device")
    It("Should enable a file audit device in Vault")
      - Load audit.yaml via decoder.GetAuditInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Sys().ListAudit() — verify device exists with trailing-slash key
      - Verify Type, Options["file_path"]

  Context("When creating an AuditRequestHeader")
    It("Should configure the request header in Vault")
      - Load auditrequestheader.yaml via decoder.GetAuditRequestHeaderInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Logical().Read sys/config/auditing/request-headers/X-Custom-Test-Header
      - Verify hmac=true in nested response format

  Context("When updating an Audit device")
    It("Should update the audit device via disable/re-enable and reflect updated ObservedGeneration")
      - Record initial ObservedGeneration
      - Get latest Audit CR, change description to "updated-description"
      - Update the CR
      - Eventually verify Sys().ListAudit() reflects new description
      - Verify ObservedGeneration increased (strictly greater than baseline)

  Context("When updating an AuditRequestHeader")
    It("Should update the header config in Vault and reflect updated ObservedGeneration")
      - Record initial ObservedGeneration
      - Get latest AuditRequestHeader CR, change hmac from true to false
      - Update the CR
      - Eventually verify Logical().Read reflects hmac=false
      - Verify ObservedGeneration increased (strictly greater than baseline)

  Context("When deleting Audit resources")
    It("Should clean up both resources from Vault")
      - Delete AuditRequestHeader (IsDeletable=true)
        - Eventually verify K8s deletion (NotFound)
        - Eventually verify header removed from Vault (Read returns nil)
      - Delete Audit (IsDeletable=true)
        - Eventually verify K8s deletion (NotFound)
        - Eventually verify device removed from Vault (not in ListAudit)
```

### Controller Registration — MUST Be Added to suite_integration_test.go

**CRITICAL:** Neither Audit nor AuditRequestHeader controllers are registered in `controllers/suite_integration_test.go`. Both must be added for integration tests to work.

Add after the existing `EntityAlias` registration (line 218):

```go
err = (&AuditReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "Audit")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&AuditRequestHeaderReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AuditRequestHeader")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())
```

[Source: controllers/suite_integration_test.go — Audit controllers absent; cf. main.go lines 321-329 for reference]

### Test Fixtures — MUST Be Created

No `test/audit/` directory exists. Create fixtures based on `config/samples/` but simplified for testing:

**`test/audit/audit.yaml`:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Audit
metadata:
  name: test-audit
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-audit
  type: file
  description: "test audit device"
  options:
    file_path: stdout
```

**`test/audit/auditrequestheader.yaml`:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuditRequestHeader
metadata:
  name: test-audit-request-header
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: X-Custom-Test-Header
  hmac: true
```

Key fixture decisions:
- Audit uses `file` type with `file_path: stdout` — safe for Kind cluster testing, no file system requirements
- Audit `path` is `test-audit` — unique, no collision with existing tests
- AuditRequestHeader `spec.name` is `X-Custom-Test-Header` — unique custom header name, no collision
- AuditRequestHeader `metadata.name` is `test-audit-request-header` — distinct from `spec.name` to verify path uses spec.name

### Decoder Methods — BOTH Must Be Added

Neither decoder method exists. Add them following the established pattern:

```go
func (d *decoder) GetAuditInstance(filename string) (*redhatcopv1alpha1.Audit, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.Audit{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.Audit)
        return o, nil
    }
    return nil, errDecode
}

func (d *decoder) GetAuditRequestHeaderInstance(filename string) (*redhatcopv1alpha1.AuditRequestHeader, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.AuditRequestHeader{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.AuditRequestHeader)
        return o, nil
    }
    return nil, errDecode
}
```

[Source: controllers/controllertestutils/decoder.go — existing pattern, e.g. GetPasswordPolicyInstance]

### Import Requirements

```go
import (
    "time"

    vault "github.com/hashicorp/vault/api"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)
```

Note: The `vault` import is needed for `Sys().ListAudit()` return type `map[string]*vault.Audit`. The `vault` package is already imported in `suite_integration_test.go` so no new module dependency is required.

### Checked Type Assertions

Per project convention: always use two-value form for all Vault response field assertions.

For Audit `ListAudit()` response: the returned `*vault.Audit` struct has typed fields (string, bool, map), so type assertions are on the map lookup (`audits["test-audit/"]`) not individual fields.

For AuditRequestHeader `Logical().Read()` response: the nested map structure requires two levels of type assertion:
```go
headerData, ok := secret.Data["X-Custom-Test-Header"].(map[string]interface{})
Expect(ok).To(BeTrue())
hmac, ok := headerData["hmac"].(bool)
Expect(ok).To(BeTrue())
```

### ObservedGeneration Baseline Assertion

Per Epic 5 retrospective action item: When testing updates, record initial ObservedGeneration BEFORE the update, then assert the post-update value is strictly greater than the recorded baseline.

### No `make manifests generate` Needed

This story adds an integration test file, decoder methods, controller registrations in the test suite, and test fixtures. No CRD types, controllers, or webhooks are changed. No Makefile changes needed.

### Name Collision Prevention

- `test-audit` — unique path name, no collision with any existing fixtures
- `X-Custom-Test-Header` — unique header name, no collision with any existing tests
- `test/audit/` directory is new and separate from all other test fixture directories

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `test/audit/audit.yaml` | New | Test fixture — file audit device writing to stdout |
| 2 | `test/audit/auditrequestheader.yaml` | New | Test fixture — request header with hmac=true |
| 3 | `controllers/suite_integration_test.go` | Modified | Register 2 Audit controllers for integration tests (after line 218) |
| 4 | `controllers/controllertestutils/decoder.go` | Modified | Add 2 decoder methods (Audit, AuditRequestHeader) |
| 5 | `controllers/audit_controller_test.go` | New | Integration test — Audit device lifecycle + AuditRequestHeader lifecycle: create, verify Vault, update, delete with Vault cleanup |

No changes to controllers, webhooks, types, Makefile, or other infrastructure.

### Previous Story Intelligence

**From Story 6.3 (Identity Token integration tests — direct predecessor):**
- Established Ordered Describe block with shared state across Contexts
- Demonstrated mixed IsDeletable behavior testing (true/false) — Story 6.4 has only IsDeletable=true for both types
- 3 decoder methods added to decoder.go — same file needs 2 more methods
- 3 controller registrations added to suite_integration_test.go — same file needs 2 more registrations

**From Story 6.1 (Group and GroupAlias integration tests):**
- Demonstrated IsDeletable=true Vault cleanup verification for identity types
- Established pattern for tests with no dependency chain (Group/GroupAlias have a dependency, but within this story Audit/AuditRequestHeader do NOT)

**From Story 3.1 (Policy integration tests — structure reference):**
- Clean example of independent types tested in a single file
- Simple create → verify → delete flow with Vault state assertions
- Reference for `Sys().ListAudit()` pattern (similar to how Policy uses `Logical().Read("sys/policy/...")`)

**From Epic 5 Retrospective:**
- ObservedGeneration baseline assertion guidance: record before update, assert strictly greater after
- Continue using Opus-class models

[Source: _bmad-output/implementation-artifacts/6-3-integration-tests-for-identity-token-types.md]
[Source: _bmad-output/implementation-artifacts/epic-5-retro-2026-04-29.md]

### Git Intelligence (Recent Commits)

```
7ce3e42 Merge pull request #320 from raffaelespazzoli/bmad-epic-5
d64b2b1 Complete Epic 5 retrospective and close epic
e5e982c Add integration tests for KubernetesSecretEngineConfig and KubernetesSecretEngineRole (Story 5.3)
168e7e0 Fix RabbitMQ role vhosts assertion type mismatch
c13227f Add integration tests for RabbitMQSecretEngineConfig and RabbitMQSecretEngineRole (Story 5.2)
```

Codebase is clean post-Epic 5 merge to main. All integration tests passing.

### Project Structure Notes

- Controller registration additions in `controllers/suite_integration_test.go` (2 new registrations after line 218)
- Decoder changes in `controllers/controllertestutils/decoder.go` (add 2 methods at end of file)
- Test file goes in `controllers/audit_controller_test.go` (combines both Audit types in one test file since they're logically grouped)
- New fixtures in `test/audit/` directory (2 YAML files)
- No Makefile changes needed
- No new infrastructure directories beyond `test/audit/`
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/audit_types.go] — VaultObject + VaultAuditObject implementation, GetPath (sys/audit/{path}), GetPayload (type/description/local/options), IsEquivalentToDesiredState (custom field-by-field), IsDeletable=true, PrepareInternalValues=nil
- [Source: api/v1alpha1/auditrequestheader_types.go] — VaultObject implementation, GetPath (sys/config/auditing/request-headers/{spec.name}), GetPayload (hmac only), IsEquivalentToDesiredState (hmac-only check), IsDeletable=true, kubebuilder default (hmac=false)
- [Source: controllers/audit_controller.go] — VaultAuditResource reconciler (NewVaultAuditResource)
- [Source: controllers/auditrequestheader_controller.go] — Standard VaultResource reconciler (NewVaultResource)
- [Source: controllers/vaultresourcecontroller/vaultauditresourcereconciler.go] — VaultAuditResource reconcile flow, manageCleanUpLogic (DisableAudit only if ReconcileSuccessful=True)
- [Source: api/v1alpha1/utils/vaultauditobject.go] — VaultAuditEndpoint: Exists (ListAudit with trailing-slash key), Enable (EnableAuditWithOptions), Disable (DisableAudit), CreateOrUpdate (disable+re-enable if not equivalent), DeleteIfExists (disable if exists)
- [Source: controllers/suite_integration_test.go] — Audit controllers NOT registered (must add after line 218)
- [Source: controllers/controllertestutils/decoder.go] — No Audit decoder methods exist (add 2)
- [Source: main.go#L321-L329] — Controller registration pattern for Audit and AuditRequestHeader
- [Source: config/samples/redhatcop_v1alpha1_audit.yaml] — Sample fixture reference (file type, stdout)
- [Source: config/samples/redhatcop_v1alpha1_auditrequestheader.yaml] — Sample fixture reference (custom-header, hmac=true)
- [Source: _bmad-output/implementation-artifacts/6-3-integration-tests-for-identity-token-types.md] — Predecessor story pattern
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
