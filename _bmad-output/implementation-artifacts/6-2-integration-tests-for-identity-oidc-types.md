# Story 6.2: Integration Tests for Identity OIDC Types

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration tests for IdentityOIDCScope, IdentityOIDCAssignment, IdentityOIDCClient, and IdentityOIDCProvider covering create, reconcile success, Vault state verification, update, and delete,
So that the full OIDC identity provider configuration lifecycle — including the 4-type dependency chain — is verified end-to-end.

## Scope Decision: Tier Classification

Per the project's three-tier integration test rule:

| Type | Dependency | Classification | Action |
|------|-----------|---------------|--------|
| IdentityOIDCScope | Vault identity OIDC API (internal) | **Tier 1: Already available** | **Test** — no external service |
| IdentityOIDCAssignment | Vault identity OIDC API (internal) | **Tier 1: Already available** | **Test** — no external service |
| IdentityOIDCClient | Vault identity OIDC API (internal) | **Tier 1: Already available** | **Test** — no external service |
| IdentityOIDCProvider | Vault identity OIDC API (internal) | **Tier 1: Already available** | **Test** — no external service |

No new infrastructure needed. All types interact with Vault's internal identity OIDC subsystem.

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

## Acceptance Criteria

1. **Given** an IdentityOIDCScope CR is created with a template and description **When** the reconciler processes it **Then** the scope exists in Vault at `identity/oidc/scope/{name}` with the correct template and description, and ReconcileSuccessful=True

2. **Given** an IdentityOIDCAssignment CR is created with entity_ids and group_ids **When** the reconciler processes it **Then** the assignment exists in Vault at `identity/oidc/assignment/{name}` and ReconcileSuccessful=True

3. **Given** an IdentityOIDCClient CR is created referencing the assignment **When** the reconciler processes it **Then** the client exists in Vault at `identity/oidc/client/{name}` with correct key, client_type, redirect_uris, assignments, and ReconcileSuccessful=True

4. **Given** an IdentityOIDCProvider CR is created referencing the scope **When** the reconciler processes it **Then** the provider exists in Vault at `identity/oidc/provider/{name}` with correct scopes_supported, allowed_client_ids, and ReconcileSuccessful=True

5. **Given** the IdentityOIDCScope CR spec is updated (e.g., description changed) **When** the reconciler processes the update **Then** the Vault scope reflects the change and ObservedGeneration increases

6. **Given** all OIDC CRs are deleted in reverse dependency order (Provider → Client → Assignment → Scope) **When** the reconcilers process the deletions **Then** all resources are removed from Vault (all IsDeletable=true) and all CRs are removed from K8s

## Tasks / Subtasks

- [ ] Task 1: Register controllers in suite_integration_test.go (AC: 1, 2, 3, 4)
  - [ ] 1.1: Add `IdentityOIDCScopeReconciler` registration
  - [ ] 1.2: Add `IdentityOIDCAssignmentReconciler` registration
  - [ ] 1.3: Add `IdentityOIDCClientReconciler` registration
  - [ ] 1.4: Add `IdentityOIDCProviderReconciler` registration

- [ ] Task 2: Add decoder methods (AC: 1, 2, 3, 4)
  - [ ] 2.1: Add `GetIdentityOIDCScopeInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 2.2: Add `GetIdentityOIDCAssignmentInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 2.3: Add `GetIdentityOIDCClientInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 2.4: Add `GetIdentityOIDCProviderInstance` to `controllers/controllertestutils/decoder.go`

- [ ] Task 3: Create integration test file (AC: 1, 2, 3, 4, 5, 6)
  - [ ] 3.1: Create `controllers/identityoidc_controller_test.go` with `//go:build integration` tag
  - [ ] 3.2: Add context for IdentityOIDCScope creation — create, poll for ReconcileSuccessful=True, verify Vault state
  - [ ] 3.3: Add context for IdentityOIDCAssignment creation — create, poll for ReconcileSuccessful=True, verify Vault state
  - [ ] 3.4: Add context for IdentityOIDCClient creation — create, poll for ReconcileSuccessful=True, verify Vault state
  - [ ] 3.5: Add context for IdentityOIDCProvider creation — create, poll for ReconcileSuccessful=True, verify Vault state
  - [ ] 3.6: Add context for IdentityOIDCScope update — update description, verify Vault reflects change, verify ObservedGeneration increased
  - [ ] 3.7: Add deletion context — delete all in reverse order (Provider → Client → Assignment → Scope), verify K8s deletion and Vault cleanup for each

- [ ] Task 4: End-to-end verification (AC: 1, 2, 3, 4, 5, 6)
  - [ ] 4.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [ ] 4.2: Verify no regressions — existing tests unaffected

## Dev Notes

### All 4 Types — Standard VaultResource Reconciler

All four OIDC types use `NewVaultResource` — standard reconcile flow (read → compare → write if different). No PrepareInternalValues logic (all return nil). All controllers are identical in structure.

[Source: controllers/identityoidcscope_controller.go, identityoidcprovider_controller.go, identityoidcclient_controller.go, identityoidcassignment_controller.go]

### All 4 Types — IsDeletable = true

All OIDC types return `IsDeletable() == true`. After deletion:
- Vault resource is deleted via finalizer
- Delete test must verify BOTH K8s NotFound AND Vault Read returns nil

[Source: api/v1alpha1/identityoidcscope_types.go#L99, identityoidcprovider_types.go#L109, identityoidcclient_types.go#L130, identityoidcassignment_types.go#L101]

### Vault API Paths

| Type | GetPath() | Vault API Path |
|------|-----------|---------------|
| IdentityOIDCScope | `identity/oidc/scope/{spec.name or metadata.name}` | Same for read/write |
| IdentityOIDCAssignment | `identity/oidc/assignment/{spec.name or metadata.name}` | Same for read/write |
| IdentityOIDCClient | `identity/oidc/client/{spec.name or metadata.name}` | Same for read/write |
| IdentityOIDCProvider | `identity/oidc/provider/{spec.name or metadata.name}` | Same for read/write |

All use `spec.Name` if set, falling back to `metadata.name`.

### Dependency Chain — Creation Order Matters

```
IdentityOIDCScope (test-scope)          ← standalone
IdentityOIDCAssignment (test-assignment) ← standalone
    └── IdentityOIDCClient (test-client)  ← references test-assignment
IdentityOIDCProvider (test-provider)     ← references test-scope via scopesSupported
```

Vault validates referenced resources exist:
- **Client** references `assignments: [test-assignment]` → assignment must exist first
- **Provider** references `scopesSupported: [test-scope]` → scope must exist first
- **Provider** uses `allowedClientIDs: ["*"]` → wildcard, no specific client dependency

**Create order:** Scope → Assignment → Client → Provider
**Delete order (reverse):** Provider → Client → Assignment → Scope

### toMap — Per-Type Field Mapping

**IdentityOIDCScope toMap:**
```go
func (i *IdentityOIDCScopeSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    if i.Template != "" {
        payload["template"] = i.Template
    }
    if i.Description != "" {
        payload["description"] = i.Description
    }
    return payload
}
```
Conditional inclusion — only non-empty fields are sent. Vault read response will include both fields.

[Source: api/v1alpha1/identityoidcscope_types.go#L122-L131]

**IdentityOIDCAssignment toMap:**
```go
func (i *IdentityOIDCAssignmentSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["entity_ids"] = i.EntityIDs
    payload["group_ids"] = i.GroupIDs
    return payload
}
```
Always includes both fields (even if empty slices).

[Source: api/v1alpha1/identityoidcassignment_types.go#L124-L129]

**IdentityOIDCClient toMap:**
```go
func (i *IdentityOIDCClientSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["key"] = i.Key
    payload["redirect_uris"] = i.RedirectURIs
    payload["assignments"] = i.Assignments
    payload["client_type"] = i.ClientType
    payload["id_token_ttl"] = i.IDTokenTTL
    payload["access_token_ttl"] = i.AccessTokenTTL
    return payload
}
```
All 6 fields always sent. **CRITICAL:** Vault returns additional server-generated fields (`client_id`, `client_secret`) and may return TTL values as integers (seconds) instead of duration strings. Since `IsEquivalentToDesiredState` uses simple `reflect.DeepEqual` without filtering, it will always return false — the reconciler writes on every cycle but still achieves ReconcileSuccessful=True. This is known tech debt (Story 7-4).

[Source: api/v1alpha1/identityoidcclient_types.go#L153-L162]

**IdentityOIDCProvider toMap:**
```go
func (i *IdentityOIDCProviderSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    if i.Issuer != "" {
        payload["issuer"] = i.Issuer
    }
    payload["allowed_client_ids"] = i.AllowedClientIDs
    payload["scopes_supported"] = i.ScopesSupported
    return payload
}
```
`issuer` is conditionally included; `allowed_client_ids` and `scopes_supported` always included.

[Source: api/v1alpha1/identityoidcprovider_types.go#L132-L140]

### IsEquivalentToDesiredState — All Use Simple DeepEqual

All four types use `reflect.DeepEqual(desiredState, payload)` with no field filtering. Vault may return extra fields not managed by the operator, causing false negatives. This does NOT affect ReconcileSuccessful=True or test correctness.

### IdentityOIDCClient — Kubebuilder Defaults

```go
Key        string `json:"key,omitempty"` // +kubebuilder:default:="default"
ClientType string `json:"clientType,omitempty"` // +kubebuilder:default:="confidential"
IDTokenTTL string `json:"idTokenTTL,omitempty"` // +kubebuilder:default:="24h"
AccessTokenTTL string `json:"accessTokenTTL,omitempty"` // +kubebuilder:default:="24h"
```

These defaults are applied by the API server at admission time. The test fixture explicitly sets these values, so defaulting is a no-op.

[Source: api/v1alpha1/identityoidcclient_types.go#L47-L87]

### IdentityOIDCClient Webhook — Immutable key and clientType

```go
func (r *IdentityOIDCClient) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
    oldClient := old.(*IdentityOIDCClient)
    if r.Spec.Key != oldClient.Spec.Key {
        return nil, errors.New("spec.key cannot be updated")
    }
    if r.Spec.ClientType != oldClient.Spec.ClientType {
        return nil, errors.New("spec.clientType cannot be updated")
    }
    return nil, nil
}
```

**IMPORTANT:** When testing Client update, do NOT change `key` or `clientType` — the webhook will reject the update. Update `redirectURIs`, `assignments`, `idTokenTTL`, or `accessTokenTTL` instead.

[Source: api/v1alpha1/identityoidcclient_webhook.go#L58-L69]

### Other Webhooks — Scaffold-Only

IdentityOIDCScope, IdentityOIDCProvider, and IdentityOIDCAssignment webhooks have scaffold-only validation (return `nil, nil`). No immutable field rules.

[Source: api/v1alpha1/identityoidcscope_webhook.go, identityoidcprovider_webhook.go, identityoidcassignment_webhook.go]

### Controller Registration — MUST Be Added to suite_integration_test.go

**CRITICAL:** None of the four OIDC controllers are registered in `controllers/suite_integration_test.go`. All four must be added for the integration tests to work.

Add after the existing `GroupAlias` registration (around line 212):

```go
err = (&IdentityOIDCScopeReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "IdentityOIDCScope")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&IdentityOIDCAssignmentReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "IdentityOIDCAssignment")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&IdentityOIDCClientReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "IdentityOIDCClient")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&IdentityOIDCProviderReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "IdentityOIDCProvider")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())
```

[Source: controllers/suite_integration_test.go — missing OIDC controllers; cf. main.go#L331-L348 for reference]

### Existing Test Fixtures — Use As-Is

All 4 fixtures already exist in `test/identityoidc/`:

| Fixture | Name | Key Spec Fields |
|---------|------|----------------|
| `test/identityoidc/identityoidcscope.yaml` | test-scope | template (groups claim), description |
| `test/identityoidc/identityoidcassignment.yaml` | test-assignment | entityIDs: [], groupIDs: [] |
| `test/identityoidc/identityoidcclient.yaml` | test-client | key: default, clientType: confidential, redirectURIs, assignments: [test-assignment] |
| `test/identityoidc/identityoidcprovider.yaml` | test-provider | allowedClientIDs: [*], scopesSupported: [test-scope] |

No new fixtures needed. Use existing fixtures via decoder methods.

### Vault API Response Shapes

**GET `identity/oidc/scope/test-scope`:**
```json
{
  "data": {
    "template": "{ \"groups\": {{identity.entity.groups.names}} }",
    "description": "A test scope for groups claim."
  }
}
```

**GET `identity/oidc/assignment/test-assignment`:**
```json
{
  "data": {
    "entity_ids": [],
    "group_ids": []
  }
}
```

**GET `identity/oidc/client/test-client`:**
```json
{
  "data": {
    "key": "default",
    "redirect_uris": ["https://example.com/callback"],
    "assignments": ["test-assignment"],
    "client_type": "confidential",
    "id_token_ttl": 86400,
    "access_token_ttl": 86400,
    "client_id": "...(server-generated UUID)...",
    "client_secret": "...(server-generated secret, only for confidential clients)..."
  }
}
```
**CRITICAL:** Vault returns `id_token_ttl` and `access_token_ttl` as **integers (seconds)** not duration strings. Vault also returns server-generated `client_id` and `client_secret`. Do NOT assert on TTL values returned from Vault read (type mismatch with the string in toMap). Assert on `key`, `client_type`, `redirect_uris`, `assignments` instead.

**GET `identity/oidc/provider/test-provider`:**
```json
{
  "data": {
    "issuer": "...(Vault-generated issuer URL)...",
    "allowed_client_ids": ["*"],
    "scopes_supported": ["test-scope"]
  }
}
```
**NOTE:** Vault may populate `issuer` with a default value even if not set in the write payload. Do NOT assert on `issuer` unless explicitly set.

### Verifying Vault State

**Scope verification:**
```go
secret, err := vaultClient.Logical().Read("identity/oidc/scope/test-scope")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

description, ok := secret.Data["description"].(string)
Expect(ok).To(BeTrue(), "expected description to be a string")
Expect(description).To(Equal("A test scope for groups claim."))

template, ok := secret.Data["template"].(string)
Expect(ok).To(BeTrue(), "expected template to be a string")
Expect(template).To(ContainSubstring("identity.entity.groups.names"))
```

**Assignment verification:**
```go
secret, err := vaultClient.Logical().Read("identity/oidc/assignment/test-assignment")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
```
Assignment with empty entity_ids/group_ids — verify the resource exists. Vault returns these as `[]interface{}` or `nil`.

**Client verification:**
```go
secret, err := vaultClient.Logical().Read("identity/oidc/client/test-client")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

clientType, ok := secret.Data["client_type"].(string)
Expect(ok).To(BeTrue(), "expected client_type to be a string")
Expect(clientType).To(Equal("confidential"))

key, ok := secret.Data["key"].(string)
Expect(ok).To(BeTrue(), "expected key to be a string")
Expect(key).To(Equal("default"))

clientID, ok := secret.Data["client_id"].(string)
Expect(ok).To(BeTrue(), "expected client_id to be a string")
Expect(clientID).NotTo(BeEmpty())
```

**Provider verification:**
```go
secret, err := vaultClient.Logical().Read("identity/oidc/provider/test-provider")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

scopesSupported, ok := secret.Data["scopes_supported"].([]interface{})
Expect(ok).To(BeTrue(), "expected scopes_supported to be []interface{}")
Expect(scopesSupported).To(ContainElement("test-scope"))
```

**Delete verification (all IsDeletable=true):**
```go
Expect(k8sIntegrationClient.Delete(ctx, instance)).Should(Succeed())
lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.IdentityOIDCScope{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("identity/oidc/scope/test-scope")
    if err != nil {
        return false
    }
    return secret == nil
}, timeout, interval).Should(BeTrue())
```

### Test Structure

```
Describe("Identity OIDC controllers", Ordered)
  var scopeInstance *redhatcopv1alpha1.IdentityOIDCScope
  var assignmentInstance *redhatcopv1alpha1.IdentityOIDCAssignment
  var clientInstance *redhatcopv1alpha1.IdentityOIDCClient
  var providerInstance *redhatcopv1alpha1.IdentityOIDCProvider

  AfterAll: best-effort delete all instances (reverse order):
    provider → client → assignment → scope

  Context("When creating an IdentityOIDCScope")
    It("Should create the scope in Vault with correct settings")
      - Load identityoidcscope.yaml via decoder.GetIdentityOIDCScopeInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read identity/oidc/scope/test-scope from Vault
      - Verify description = "A test scope for groups claim."
      - Verify template contains "identity.entity.groups.names"

  Context("When creating an IdentityOIDCAssignment")
    It("Should create the assignment in Vault")
      - Load identityoidcassignment.yaml via decoder.GetIdentityOIDCAssignmentInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read identity/oidc/assignment/test-assignment from Vault
      - Verify resource exists (non-nil response)

  Context("When creating an IdentityOIDCClient")
    It("Should create the client in Vault with correct settings")
      - Load identityoidcclient.yaml via decoder.GetIdentityOIDCClientInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read identity/oidc/client/test-client from Vault
      - Verify client_type = "confidential"
      - Verify key = "default"
      - Verify client_id is non-empty (server-generated)

  Context("When creating an IdentityOIDCProvider")
    It("Should create the provider in Vault with correct settings")
      - Load identityoidcprovider.yaml via decoder.GetIdentityOIDCProviderInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read identity/oidc/provider/test-provider from Vault
      - Verify scopes_supported contains "test-scope"

  Context("When updating an IdentityOIDCScope")
    It("Should update the scope in Vault and reflect updated ObservedGeneration")
      - Record initial ObservedGeneration
      - Get latest IdentityOIDCScope CR, change description to "Updated scope description."
      - Update the CR
      - Eventually verify Vault scope reflects updated description
      - Verify ObservedGeneration increased (strictly greater than baseline)

  Context("When deleting all OIDC resources")
    It("Should clean up all resources from Vault in reverse dependency order")
      - Delete Provider (IsDeletable=true → Vault cleanup)
        - Eventually verify K8s deletion (NotFound)
        - Eventually verify provider removed from Vault (Read returns nil)
      - Delete Client (IsDeletable=true → Vault cleanup)
        - Eventually verify K8s deletion (NotFound)
        - Eventually verify client removed from Vault (Read returns nil)
      - Delete Assignment (IsDeletable=true → Vault cleanup)
        - Eventually verify K8s deletion (NotFound)
        - Eventually verify assignment removed from Vault (Read returns nil)
      - Delete Scope (IsDeletable=true → Vault cleanup)
        - Eventually verify K8s deletion (NotFound)
        - Eventually verify scope removed from Vault (Read returns nil)
```

### Decoder Methods — ALL 4 Must Be Added

None of the four OIDC decoder methods exist. Add them following the established pattern:

```go
func (d *decoder) GetIdentityOIDCScopeInstance(filename string) (*redhatcopv1alpha1.IdentityOIDCScope, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.IdentityOIDCScope{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.IdentityOIDCScope)
        return o, nil
    }
    return nil, errDecode
}
```

Repeat pattern for `GetIdentityOIDCAssignmentInstance`, `GetIdentityOIDCClientInstance`, `GetIdentityOIDCProviderInstance`.

[Source: controllers/controllertestutils/decoder.go — existing pattern, e.g. GetPasswordPolicyInstance]

### Import Requirements

```go
import (
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)
```

No additional imports needed beyond the standard integration test set.

### Name Collision Prevention

Fixture names use `test-scope`, `test-assignment`, `test-client`, `test-provider`:
- These don't collide with any existing test fixtures
- The `test/identityoidc/` directory is separate from other test directories
- No overlap with Group/GroupAlias tests (`test-group`, `test-groupalias`) or Entity/EntityAlias tests

### ObservedGeneration Baseline Assertion

Per Epic 5 retrospective action item: When testing updates, record initial ObservedGeneration BEFORE the update, then assert the post-update value is strictly greater than the recorded baseline.

### Checked Type Assertions

Per project convention: always use two-value form `val, ok := x.(string)` with `Expect(ok).To(BeTrue())` for all Vault response field assertions.

### No `make manifests generate` Needed

This story adds an integration test file, decoder methods, and controller registrations in the test suite. No CRD types, controllers, or webhooks are changed. No Makefile changes needed.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/suite_integration_test.go` | Modified | Register 4 OIDC controllers for integration tests |
| 2 | `controllers/controllertestutils/decoder.go` | Modified | Add 4 decoder methods (Scope, Assignment, Client, Provider) |
| 3 | `controllers/identityoidc_controller_test.go` | New | Integration test — OIDC lifecycle: create 4 types in dependency order, update scope, delete in reverse order |

No new fixtures needed (existing `test/identityoidc/*.yaml` fixtures are used).
No changes to controllers, webhooks, types, Makefile, or other infrastructure.

### Previous Story Intelligence

**From Story 6.1 (Group and GroupAlias integration tests — direct predecessor):**
- Established Epic 6 Tier 1 "no new infrastructure" pattern — same applies here
- Demonstrated IsDeletable=true Vault cleanup verification for identity types
- Demonstrated Ordered Describe block with shared state across Contexts
- Demonstrated dependency-ordered creation (Group before GroupAlias) and reverse deletion

**From PasswordPolicy integration tests (closest analogy for simple types):**
- Demonstrates straightforward create → verify Vault → delete → verify Vault cleanup pattern
- Good reference for types with no PrepareInternalValues and simple toMap logic

**From Epic 5 Retrospective:**
- Story 6.2 noted as "4-type dependency chain"
- No infrastructure needed for any Epic 6 story
- ObservedGeneration baseline assertion guidance: record before update, assert strictly greater after
- Continue using Opus-class models

[Source: _bmad-output/implementation-artifacts/epic-5-retro-2026-04-29.md]

### Git Intelligence (Recent Commits)

```
7ce3e42 Merge pull request #320 from raffaelespazzoli/bmad-epic-5
d64b2b1 Complete Epic 5 retrospective and close epic
e5e982c Add integration tests for KubernetesSecretEngineConfig and KubernetesSecretEngineRole (Story 5.3)
168e7e0 Fix RabbitMQ role vhosts assertion type mismatch
c13227f Add integration tests for RabbitMQSecretEngineConfig and RabbitMQSecretEngineRole (Story 5.2)
```

Codebase is clean post-Epic 5 merge to main. All 63 integration tests passing with 46.0% coverage.

### Project Structure Notes

- Controller registration additions in `controllers/suite_integration_test.go` (4 new registrations after line 212)
- Decoder changes in `controllers/controllertestutils/decoder.go` (add 4 methods)
- Test file goes in `controllers/identityoidc_controller_test.go` (combines all 4 OIDC types in one test file since they form a dependency chain)
- Existing fixtures in `test/identityoidc/` are used — no new fixture files
- No Makefile changes needed
- No new infrastructure directories
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/identityoidcscope_types.go] — VaultObject implementation, GetPath (identity/oidc/scope/{name}), toMap (conditional template/description), IsDeletable=true, IsEquivalentToDesiredState (simple DeepEqual)
- [Source: api/v1alpha1/identityoidcassignment_types.go] — VaultObject implementation, GetPath (identity/oidc/assignment/{name}), toMap (entity_ids, group_ids), IsDeletable=true
- [Source: api/v1alpha1/identityoidcclient_types.go] — VaultObject implementation, GetPath (identity/oidc/client/{name}), toMap (6 fields including TTLs), IsDeletable=true, kubebuilder defaults (key=default, clientType=confidential, idTokenTTL=24h, accessTokenTTL=24h)
- [Source: api/v1alpha1/identityoidcprovider_types.go] — VaultObject implementation, GetPath (identity/oidc/provider/{name}), toMap (conditional issuer, allowed_client_ids, scopes_supported), IsDeletable=true
- [Source: api/v1alpha1/identityoidcscope_webhook.go] — Scaffold-only webhook, no validation logic
- [Source: api/v1alpha1/identityoidcprovider_webhook.go] — Scaffold-only webhook, no validation logic
- [Source: api/v1alpha1/identityoidcclient_webhook.go] — ValidateUpdate rejects changes to key and clientType
- [Source: api/v1alpha1/identityoidcassignment_webhook.go] — Scaffold-only webhook, no validation logic
- [Source: controllers/identityoidcscope_controller.go] — Standard VaultResource reconciler
- [Source: controllers/identityoidcprovider_controller.go] — Standard VaultResource reconciler
- [Source: controllers/identityoidcclient_controller.go] — Standard VaultResource reconciler
- [Source: controllers/identityoidcassignment_controller.go] — Standard VaultResource reconciler
- [Source: controllers/suite_integration_test.go] — OIDC controllers NOT registered (must add)
- [Source: main.go#L331-L348] — All 4 OIDC controllers registered in production
- [Source: controllers/controllertestutils/decoder.go] — No OIDC decoder methods exist
- [Source: test/identityoidc/identityoidcscope.yaml] — Existing fixture (test-scope)
- [Source: test/identityoidc/identityoidcassignment.yaml] — Existing fixture (test-assignment)
- [Source: test/identityoidc/identityoidcclient.yaml] — Existing fixture (test-client)
- [Source: test/identityoidc/identityoidcprovider.yaml] — Existing fixture (test-provider)
- [Source: controllers/passwordpolicy_controller_test.go] — Reference pattern for simple type integration tests
- [Source: _bmad-output/implementation-artifacts/epic-5-retro-2026-04-29.md] — Epic 5 retrospective, Epic 6 readiness
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
