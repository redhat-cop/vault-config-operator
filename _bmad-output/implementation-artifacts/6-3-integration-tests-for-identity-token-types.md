# Story 6.3: Integration Tests for Identity Token Types

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration tests for IdentityTokenConfig, IdentityTokenKey, and IdentityTokenRole covering create, reconcile success, Vault state verification, update, and delete,
So that the identity token configuration lifecycle — including the 3-type dependency chain (Config → Key → Role) — is verified end-to-end.

## Scope Decision: Tier Classification

Per the project's three-tier integration test rule:

| Type | Dependency | Classification | Action |
|------|-----------|---------------|--------|
| IdentityTokenConfig | Vault identity OIDC API (internal) | **Tier 1: Already available** | **Test** — no external service |
| IdentityTokenKey | Vault identity OIDC API (internal) | **Tier 1: Already available** | **Test** — no external service |
| IdentityTokenRole | Vault identity OIDC API (internal) | **Tier 1: Already available** | **Test** — requires key to exist first |

No new infrastructure needed. All types interact with Vault's internal identity OIDC/token subsystem.

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

## Acceptance Criteria

1. **Given** an IdentityTokenConfig CR is created with an empty issuer **When** the reconciler processes it **Then** the config exists in Vault at `identity/oidc/config` and ReconcileSuccessful=True

2. **Given** an IdentityTokenKey CR is created with rotationPeriod, verificationTTL, allowedClientIDs, and algorithm **When** the reconciler processes it **Then** the key exists in Vault at `identity/oidc/key/{name}` with correct settings and ReconcileSuccessful=True

3. **Given** an IdentityTokenRole CR is created referencing the test key **When** the reconciler processes it **Then** the role exists in Vault at `identity/oidc/role/{name}` with the correct key and ttl, and ReconcileSuccessful=True

4. **Given** the IdentityTokenKey CR spec is updated (e.g., algorithm changed from RS256 to ES256) **When** the reconciler processes the update **Then** the Vault key reflects the change and ObservedGeneration increases

5. **Given** the IdentityTokenRole CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the role is removed from Vault and the CR is removed from K8s

6. **Given** the IdentityTokenKey CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the key is removed from Vault and the CR is removed from K8s

7. **Given** the IdentityTokenConfig CR is deleted (IsDeletable=false) **When** the reconciler processes the deletion **Then** the CR is removed from K8s BUT the Vault config persists (no finalizer cleanup)

## Tasks / Subtasks

- [ ] Task 1: Register controllers in suite_integration_test.go (AC: 1, 2, 3)
  - [ ] 1.1: Add `IdentityTokenConfigReconciler` registration
  - [ ] 1.2: Add `IdentityTokenKeyReconciler` registration
  - [ ] 1.3: Add `IdentityTokenRoleReconciler` registration

- [ ] Task 2: Add decoder methods (AC: 1, 2, 3)
  - [ ] 2.1: Add `GetIdentityTokenConfigInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 2.2: Add `GetIdentityTokenKeyInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 2.3: Add `GetIdentityTokenRoleInstance` to `controllers/controllertestutils/decoder.go`

- [ ] Task 3: Create integration test file (AC: 1, 2, 3, 4, 5, 6, 7)
  - [ ] 3.1: Create `controllers/identitytoken_controller_test.go` with `//go:build integration` tag
  - [ ] 3.2: Add context for IdentityTokenConfig creation — create, poll for ReconcileSuccessful=True, verify Vault state
  - [ ] 3.3: Add context for IdentityTokenKey creation — create, poll for ReconcileSuccessful=True, verify Vault state
  - [ ] 3.4: Add context for IdentityTokenRole creation — create, poll for ReconcileSuccessful=True, verify Vault state
  - [ ] 3.5: Add context for IdentityTokenKey update — change algorithm, verify Vault reflects change, verify ObservedGeneration increased
  - [ ] 3.6: Add deletion context — delete Role (IsDeletable=true, verify Vault cleanup), delete Key (IsDeletable=true, verify Vault cleanup), delete Config (IsDeletable=false, verify Vault persists)

- [ ] Task 4: End-to-end verification (AC: 1, 2, 3, 4, 5, 6, 7)
  - [ ] 4.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [ ] 4.2: Verify no regressions — existing tests unaffected

## Dev Notes

### All 3 Types — Standard VaultResource Reconciler

All three identity token types use `NewVaultResource` — standard reconcile flow (read → compare → write if different). No PrepareInternalValues logic (all return nil). All controllers are identical in structure.

[Source: controllers/identitytokenconfig_controller.go, identitytokenkey_controller.go, identitytokenrole_controller.go]

### Deletability — MIXED

| Type | IsDeletable | Behavior |
|------|-------------|----------|
| IdentityTokenConfig | **false** | No finalizer. CR deletion removes K8s object only. Vault config persists. |
| IdentityTokenKey | **true** | Finalizer added. CR deletion removes key from Vault. |
| IdentityTokenRole | **true** | Finalizer added. CR deletion removes role from Vault. |

**CRITICAL:** IdentityTokenConfig is `IsDeletable() == false`. The delete test for Config must verify that Vault state PERSISTS after K8s CR deletion. This follows the same pattern as KubernetesAuthEngineConfig, LDAPAuthEngineConfig, JWTOIDCAuthEngineConfig.

[Source: api/v1alpha1/identitytokenconfig_types.go#L92-L94 — IsDeletable=false]
[Source: api/v1alpha1/identitytokenkey_types.go#L116-L118 — IsDeletable=true]
[Source: api/v1alpha1/identitytokenrole_types.go#L110-L112 — IsDeletable=true]

### Vault API Paths

| Type | GetPath() | Vault API Path |
|------|-----------|---------------|
| IdentityTokenConfig | `identity/oidc/config` | Fixed path — singleton config, no name parameter |
| IdentityTokenKey | `identity/oidc/key/{spec.name or metadata.name}` | Named key |
| IdentityTokenRole | `identity/oidc/role/{spec.name or metadata.name}` | Named role |

**IMPORTANT:** IdentityTokenConfig has a **fixed path** (`identity/oidc/config`) — it is a singleton resource. There is no name component in the path. Both Key and Role use `spec.Name` if set, falling back to `metadata.name`.

[Source: api/v1alpha1/identitytokenconfig_types.go#L104-L106]
[Source: api/v1alpha1/identitytokenkey_types.go#L128-L133]
[Source: api/v1alpha1/identitytokenrole_types.go#L122-L127]

### Dependency Chain — Creation Order Matters

```
IdentityTokenConfig                     ← singleton, no dependencies
    └── IdentityTokenKey (test-key)     ← standalone (config must exist for OIDC subsystem)
        └── IdentityTokenRole (test-role) ← references test-key via `key` field
```

Vault validates:
- **Role** references `key: test-key` → key must exist first
- **Config** is the global OIDC issuer config — should be set before keys/roles but Vault doesn't strictly enforce this

**Create order:** Config → Key → Role
**Delete order (reverse):** Role → Key → Config

### toMap — Per-Type Field Mapping

**IdentityTokenConfig toMap:**
```go
func (i *IdentityTokenConfigSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["issuer"] = i.Issuer
    return payload
}
```
Always includes `issuer` (even if empty string). The test fixture sets `issuer: ""` so the payload is `{"issuer": ""}`.

[Source: api/v1alpha1/identitytokenconfig_types.go#L112-L116]

**IdentityTokenKey toMap:**
```go
func (i *IdentityTokenKeySpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["rotation_period"] = i.RotationPeriod
    payload["verification_ttl"] = i.VerificationTTL
    payload["allowed_client_ids"] = i.AllowedClientIDs
    payload["algorithm"] = i.Algorithm
    return payload
}
```
All 4 fields always included.

[Source: api/v1alpha1/identitytokenkey_types.go#L139-L146]

**IdentityTokenRole toMap:**
```go
func (i *IdentityTokenRoleSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["key"] = i.Key
    if i.Template != "" {
        payload["template"] = i.Template
    }
    if i.ClientID != "" {
        payload["client_id"] = i.ClientID
    }
    payload["ttl"] = i.TTL
    return payload
}
```
`key` and `ttl` always included; `template` and `client_id` conditionally included (only when non-empty). The test fixture has no template or clientID set, so only `key` and `ttl` will be in the payload.

[Source: api/v1alpha1/identitytokenrole_types.go#L133-L144]

### IsEquivalentToDesiredState — All Use Simple DeepEqual

All three types use `reflect.DeepEqual(desiredState, payload)` with no field filtering.

**IMPORTANT for IdentityTokenKey:** Vault may return duration fields (`rotation_period`, `verification_ttl`) as integers (seconds) rather than duration strings. Since `IsEquivalentToDesiredState` uses simple DeepEqual, this causes false negatives — the reconciler writes on every cycle. This does NOT affect ReconcileSuccessful=True or test correctness. Do NOT assert on TTL field values from Vault read; assert on `algorithm` and `allowed_client_ids` instead.

**IMPORTANT for IdentityTokenRole:** Vault returns `ttl` as an integer (seconds, e.g., `43200` for "12h") and may include server-generated `client_id`. Do NOT assert on `ttl` from Vault read. Assert on `key` field instead.

### Webhooks — All Scaffold-Only

All three webhooks (IdentityTokenConfig, IdentityTokenKey, IdentityTokenRole) have scaffold-only validation (return `nil, nil`). No immutable field rules. Updates can change any field.

[Source: api/v1alpha1/identitytokenconfig_webhook.go, identitytokenkey_webhook.go, identitytokenrole_webhook.go]

### Kubebuilder Defaults (IdentityTokenKey)

```go
RotationPeriod  string   `json:"rotationPeriod,omitempty"` // +kubebuilder:default:="24h"
VerificationTTL string   `json:"verificationTTL,omitempty"` // +kubebuilder:default:="24h"
Algorithm       string   `json:"algorithm,omitempty"` // +kubebuilder:default:="RS256"
```

The test fixture explicitly sets these values, so defaulting is a no-op.

### Kubebuilder Defaults (IdentityTokenRole)

```go
TTL string `json:"ttl,omitempty"` // +kubebuilder:default:="24h"
```

The test fixture overrides this with `ttl: "12h"`.

### Controller Registration — MUST Be Added to suite_integration_test.go

**CRITICAL:** None of the three Identity Token controllers are registered in `controllers/suite_integration_test.go`. All three must be added for the integration tests to work.

Add after the existing `EntityAlias` registration (line 218):

```go
err = (&IdentityTokenConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "IdentityTokenConfig")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&IdentityTokenKeyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "IdentityTokenKey")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&IdentityTokenRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "IdentityTokenRole")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())
```

[Source: controllers/suite_integration_test.go — missing Token controllers; cf. main.go for reference]

### Existing Test Fixtures — Use As-Is

All 3 fixtures already exist in `test/identitytoken/`:

| Fixture | Name | Key Spec Fields |
|---------|------|----------------|
| `test/identitytoken/identitytokenconfig.yaml` | test-token-config | issuer: "" |
| `test/identitytoken/identitytokenkey.yaml` | test-key | rotationPeriod: 24h, verificationTTL: 24h, allowedClientIDs: ["*"], algorithm: RS256 |
| `test/identitytoken/identitytokenrole.yaml` | test-role | key: test-key, ttl: 12h |

No new fixtures needed. Use existing fixtures via decoder methods.

### Vault API Response Shapes

**GET `identity/oidc/config`:**
```json
{
  "data": {
    "issuer": "https://vault.example.com:8200/v1/identity/oidc"
  }
}
```
**NOTE:** Even when issuer is written as `""`, Vault returns a default-populated issuer URL (based on `api_addr`). Do NOT assert `issuer == ""`. Instead, assert the response is non-nil (config exists).

**GET `identity/oidc/key/test-key`:**
```json
{
  "data": {
    "rotation_period": 86400,
    "verification_ttl": 86400,
    "allowed_client_ids": ["*"],
    "algorithm": "RS256"
  }
}
```
**NOTE:** Vault returns `rotation_period` and `verification_ttl` as integers (seconds), not duration strings. Assert on `algorithm` (string) and `allowed_client_ids` ([]interface{}) for reliable verification.

**GET `identity/oidc/role/test-role`:**
```json
{
  "data": {
    "key": "test-key",
    "template": "",
    "client_id": "...(server-generated UUID)...",
    "ttl": 43200
  }
}
```
**NOTE:** Vault returns `ttl` as integer (seconds) and may include an empty `template` and server-generated `client_id`. Assert on `key` field (string) for reliable verification.

### Verifying Vault State

**Config verification:**
```go
secret, err := vaultClient.Logical().Read("identity/oidc/config")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data).NotTo(BeNil())
```
Config is a singleton — just verify it exists and is readable. Do NOT assert on `issuer` value (Vault populates a default).

**Key verification:**
```go
secret, err := vaultClient.Logical().Read("identity/oidc/key/test-key")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

algorithm, ok := secret.Data["algorithm"].(string)
Expect(ok).To(BeTrue(), "expected algorithm to be a string")
Expect(algorithm).To(Equal("RS256"))

allowedClientIDs, ok := secret.Data["allowed_client_ids"].([]interface{})
Expect(ok).To(BeTrue(), "expected allowed_client_ids to be []interface{}")
Expect(allowedClientIDs).To(ContainElement("*"))
```

**Role verification:**
```go
secret, err := vaultClient.Logical().Read("identity/oidc/role/test-role")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

key, ok := secret.Data["key"].(string)
Expect(ok).To(BeTrue(), "expected key to be a string")
Expect(key).To(Equal("test-key"))
```

**Delete verification for IsDeletable=true (Key and Role):**
```go
Expect(k8sIntegrationClient.Delete(ctx, instance)).Should(Succeed())
lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.IdentityTokenKey{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("identity/oidc/key/test-key")
    if err != nil {
        return false
    }
    return secret == nil
}, timeout, interval).Should(BeTrue())
```

**Delete verification for IsDeletable=false (Config):**
```go
Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, lookupKey, &redhatcopv1alpha1.IdentityTokenConfig{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

// Vault config MUST still exist — IsDeletable=false means no Vault cleanup
secret, err := vaultClient.Logical().Read("identity/oidc/config")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data).NotTo(BeNil())
```

### Test Structure

```
Describe("Identity Token controllers", Ordered)
  var configInstance *redhatcopv1alpha1.IdentityTokenConfig
  var keyInstance *redhatcopv1alpha1.IdentityTokenKey
  var roleInstance *redhatcopv1alpha1.IdentityTokenRole

  AfterAll: best-effort delete all instances (reverse order):
    role → key → config

  Context("When creating an IdentityTokenConfig")
    It("Should configure the OIDC issuer in Vault")
      - Load identitytokenconfig.yaml via decoder.GetIdentityTokenConfigInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read identity/oidc/config from Vault
      - Verify response is non-nil (config exists)

  Context("When creating an IdentityTokenKey")
    It("Should create the key in Vault with correct settings")
      - Load identitytokenkey.yaml via decoder.GetIdentityTokenKeyInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read identity/oidc/key/test-key from Vault
      - Verify algorithm = "RS256"
      - Verify allowed_client_ids contains "*"

  Context("When creating an IdentityTokenRole")
    It("Should create the role in Vault referencing the key")
      - Load identitytokenrole.yaml via decoder.GetIdentityTokenRoleInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read identity/oidc/role/test-role from Vault
      - Verify key = "test-key"

  Context("When updating an IdentityTokenKey")
    It("Should update the key in Vault and reflect updated ObservedGeneration")
      - Record initial ObservedGeneration
      - Get latest IdentityTokenKey CR, change algorithm to "ES256"
      - Update the CR
      - Eventually verify Vault key reflects algorithm = "ES256"
      - Verify ObservedGeneration increased (strictly greater than baseline)

  Context("When deleting Identity Token resources")
    It("Should clean up deletable resources from Vault in reverse dependency order")
      - Delete Role (IsDeletable=true → Vault cleanup)
        - Eventually verify K8s deletion (NotFound)
        - Eventually verify role removed from Vault (Read returns nil)
      - Delete Key (IsDeletable=true → Vault cleanup)
        - Eventually verify K8s deletion (NotFound)
        - Eventually verify key removed from Vault (Read returns nil)
      - Delete Config (IsDeletable=false → Vault persists)
        - Eventually verify K8s deletion (NotFound)
        - Read identity/oidc/config from Vault — MUST still exist
```

### Decoder Methods — ALL 3 Must Be Added

None of the three Identity Token decoder methods exist. Add them following the established pattern:

```go
func (d *decoder) GetIdentityTokenConfigInstance(filename string) (*redhatcopv1alpha1.IdentityTokenConfig, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.IdentityTokenConfig{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.IdentityTokenConfig)
        return o, nil
    }
    return nil, errDecode
}

func (d *decoder) GetIdentityTokenKeyInstance(filename string) (*redhatcopv1alpha1.IdentityTokenKey, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.IdentityTokenKey{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.IdentityTokenKey)
        return o, nil
    }
    return nil, errDecode
}

func (d *decoder) GetIdentityTokenRoleInstance(filename string) (*redhatcopv1alpha1.IdentityTokenRole, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.IdentityTokenRole{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.IdentityTokenRole)
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

Fixture names use `test-token-config`, `test-key`, `test-role`:
- `test-token-config` — unique, no collision with any existing fixtures
- `test-key` — **POTENTIAL COLLISION WARNING:** The OIDC key named "default" is used by IdentityOIDCClient (from Story 6.2). However, this fixture uses `test-key` which is distinct. No collision.
- `test-role` — **POTENTIAL COLLISION WARNING:** Check that no other test uses `identity/oidc/role/test-role`. The IdentityOIDCProvider tests use `test-provider` and IdentityOIDCClient tests use `test-client`. No collision.

The `test/identitytoken/` directory is separate from `test/identityoidc/` used by Story 6.2.

### ObservedGeneration Baseline Assertion

Per Epic 5 retrospective action item: When testing updates, record initial ObservedGeneration BEFORE the update, then assert the post-update value is strictly greater than the recorded baseline.

### Checked Type Assertions

Per project convention: always use two-value form `val, ok := x.(string)` with `Expect(ok).To(BeTrue())` for all Vault response field assertions.

### No `make manifests generate` Needed

This story adds an integration test file, decoder methods, and controller registrations in the test suite. No CRD types, controllers, or webhooks are changed. No Makefile changes needed.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/suite_integration_test.go` | Modified | Register 3 Identity Token controllers for integration tests |
| 2 | `controllers/controllertestutils/decoder.go` | Modified | Add 3 decoder methods (Config, Key, Role) |
| 3 | `controllers/identitytoken_controller_test.go` | New | Integration test — Token lifecycle: create 3 types in dependency order, update key, delete in reverse order with mixed IsDeletable behavior |

No new fixtures needed (existing `test/identitytoken/*.yaml` fixtures are used).
No changes to controllers, webhooks, types, Makefile, or other infrastructure.

### Previous Story Intelligence

**From Story 6.2 (Identity OIDC integration tests — direct predecessor):**
- Established dependency chain pattern for Vault identity OIDC types
- Demonstrated Ordered Describe block with shared state across Contexts
- Demonstrated IsDeletable=true Vault cleanup verification
- Used `allowedClientIDs: ["*"]` wildcard pattern — same concept applies here
- 4 decoder methods added to decoder.go — same file needs 3 more methods

**From Story 6.1 (Group and GroupAlias integration tests):**
- Demonstrated IsDeletable=true Vault cleanup verification for identity types
- Demonstrated dependency-ordered creation (Group before GroupAlias) and reverse deletion
- GroupAlias uses non-trivial PrepareInternalValues — Token types do NOT (all return nil)

**From Epic 4 (JWTOIDCAuthEngineConfig — IsDeletable=false reference):**
- JWTOIDCAuthEngineConfig tests demonstrate the IsDeletable=false delete verification pattern:
  - Delete CR from K8s, verify NotFound
  - Read Vault path — assert config data still present
- Same pattern must be used for IdentityTokenConfig

**From Epic 5 Retrospective:**
- ObservedGeneration baseline assertion guidance: record before update, assert strictly greater after
- Continue using Opus-class models
- No infrastructure needed for any Epic 6 story

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

- Controller registration additions in `controllers/suite_integration_test.go` (3 new registrations after line 218)
- Decoder changes in `controllers/controllertestutils/decoder.go` (add 3 methods at end of file)
- Test file goes in `controllers/identitytoken_controller_test.go` (combines all 3 Token types in one test file since they form a dependency chain)
- Existing fixtures in `test/identitytoken/` are used — no new fixture files
- No Makefile changes needed
- No new infrastructure directories
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/identitytokenconfig_types.go] — VaultObject implementation, GetPath (identity/oidc/config — singleton), toMap (issuer always included), IsDeletable=false, IsEquivalentToDesiredState (simple DeepEqual)
- [Source: api/v1alpha1/identitytokenkey_types.go] — VaultObject implementation, GetPath (identity/oidc/key/{name}), toMap (rotation_period, verification_ttl, allowed_client_ids, algorithm), IsDeletable=true, kubebuilder defaults (rotationPeriod=24h, verificationTTL=24h, algorithm=RS256)
- [Source: api/v1alpha1/identitytokenrole_types.go] — VaultObject implementation, GetPath (identity/oidc/role/{name}), toMap (key always, template/client_id conditional, ttl always), IsDeletable=true, kubebuilder default (ttl=24h)
- [Source: api/v1alpha1/identitytokenconfig_webhook.go] — Scaffold-only webhook, no validation logic
- [Source: api/v1alpha1/identitytokenkey_webhook.go] — Scaffold-only webhook, no validation logic
- [Source: api/v1alpha1/identitytokenrole_webhook.go] — Scaffold-only webhook, no validation logic
- [Source: controllers/identitytokenconfig_controller.go] — Standard VaultResource reconciler
- [Source: controllers/identitytokenkey_controller.go] — Standard VaultResource reconciler
- [Source: controllers/identitytokenrole_controller.go] — Standard VaultResource reconciler
- [Source: controllers/suite_integration_test.go] — Token controllers NOT registered (must add after line 218)
- [Source: controllers/controllertestutils/decoder.go] — No Token decoder methods exist (add 3)
- [Source: test/identitytoken/identitytokenconfig.yaml] — Existing fixture (test-token-config, issuer: "")
- [Source: test/identitytoken/identitytokenkey.yaml] — Existing fixture (test-key, RS256, 24h rotation/verification, allowedClientIDs: ["*"])
- [Source: test/identitytoken/identitytokenrole.yaml] — Existing fixture (test-role, key: test-key, ttl: 12h)
- [Source: _bmad-output/implementation-artifacts/6-2-integration-tests-for-identity-oidc-types.md] — Predecessor story pattern
- [Source: _bmad-output/implementation-artifacts/6-1-integration-tests-for-group-and-groupalias-types.md] — Epic 6 pattern reference
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy
- [Source: _bmad-output/project-context.md#L143-L147] — Non-deletable type delete verification rule

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
