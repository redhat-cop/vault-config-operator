# Story 3.3: Integration Tests for SecretEngineMount Type

Status: ready-for-dev

## Story

As an operator developer,
I want integration tests for the SecretEngineMount type covering create, tune verification, accessor status, and delete with Vault cleanup,
So that the foundation for all secret engine types is verified end-to-end.

## Acceptance Criteria

1. **Given** a SecretEngineMount CR is created in the test namespace with type=kv **When** the reconciler processes it **Then** the engine is enabled in Vault at `sys/mounts/{path}`, the accessor is populated in `status.accessor`, and ReconcileSuccessful=True

2. **Given** a SecretEngineMount CR is created with non-default tune config (maxLeaseTTL, listingVisibility) **When** the reconciler processes it **Then** the tune config in Vault at `{path}/tune` matches the specified values

3. **Given** a SecretEngineMount CR using `spec.name` (overriding `metadata.name`) **When** the reconciler processes it **Then** the engine is mounted at a path derived from `spec.name` (not `metadata.name`)

4. **Given** a successfully mounted SecretEngineMount is deleted **When** the reconciler processes the deletion **Then** the engine is disabled/unmounted from Vault and the finalizer is cleared

## Tasks / Subtasks

- [ ] Task 1: Create test fixtures (AC: 1, 2, 3)
  - [ ] 1.1: Create `test/secretenginemount/simple-kv-mount.yaml` — a minimal SecretEngineMount CR with `metadata.name: test-kv-mount`, `type: kv`, `options: {version: "2"}`, using `authentication.role: policy-admin`, no `spec.path` (mounts at `sys/mounts/test-kv-mount`)
  - [ ] 1.2: Create `test/secretenginemount/tuned-kv-mount.yaml` — a SecretEngineMount CR with `metadata.name: test-tuned-kv-mount`, `type: kv`, `options: {version: "2"}`, `config.maxLeaseTTL: "8760h"`, `config.listingVisibility: "unauth"`, using `authentication.role: policy-admin`, no `spec.path`
  - [ ] 1.3: Create `test/secretenginemount/named-kv-mount.yaml` — a SecretEngineMount CR with `metadata.name: test-named-sem-metadata`, `spec.name: test-named-kv-mount`, `type: kv`, `options: {version: "2"}`, using `authentication.role: policy-admin`, no `spec.path` (mounts at `sys/mounts/test-named-kv-mount`)

- [ ] Task 2: Create integration test file (AC: 1, 2, 3, 4)
  - [ ] 2.1: Create `controllers/secretenginemount_controller_test.go` with `//go:build integration` tag, package `controllers`, standard Ginkgo imports
  - [ ] 2.2: Add `Describe("SecretEngineMount controller")` with `timeout := 120 * time.Second`, `interval := 2 * time.Second`
  - [ ] 2.3: Add `Context("When creating a simple SecretEngineMount")` — load `simple-kv-mount.yaml` via `decoder.GetSecretEngineMountInstance`, set namespace to `vaultAdminNamespaceName`, create it, poll for `ReconcileSuccessful=True`
  - [ ] 2.4: After reconcile success, verify the mount exists in Vault by reading `sys/mounts` and checking for the key matching the mount path (key format: `test-kv-mount/`)
  - [ ] 2.5: Verify `status.accessor` is populated (non-empty string) on the CR after reconcile
  - [ ] 2.6: Add `Context("When creating a SecretEngineMount with tune config")` — load `tuned-kv-mount.yaml`, set namespace, create, poll for `ReconcileSuccessful=True`
  - [ ] 2.7: After reconcile success, read the tune config from Vault via `vaultClient.Logical().Read("sys/mounts/test-tuned-kv-mount/tune")`, verify `secret.Data["max_lease_ttl"]` matches the configured value (Vault returns duration as integer seconds: 31536000 for "8760h") and `secret.Data["listing_visibility"]` equals `"unauth"`
  - [ ] 2.8: Add `Context("When creating a SecretEngineMount with spec.name override")` — load `named-kv-mount.yaml`, set namespace, create, poll for `ReconcileSuccessful=True`
  - [ ] 2.9: After reconcile success, verify the mount exists at `test-named-kv-mount/` in `sys/mounts` (not `test-named-sem-metadata/`)
  - [ ] 2.10: Add `Context("When deleting SecretEngineMounts")` — delete all three SecretEngineMount CRs, use `Eventually` to poll for K8s deletion (NotFound error), then verify the mounts no longer exist in `sys/mounts` output
  - [ ] 2.11: Verify the finalizer was cleared by confirming deletion completes (the `Eventually` for NotFound confirms this)

- [ ] Task 3: End-to-end verification (AC: 1, 2, 3, 4)
  - [ ] 3.1: Run `make integration` and verify the new SecretEngineMount tests pass alongside all existing tests
  - [ ] 3.2: Verify no regressions in other tests that use SecretEngineMount as a dependency (VaultSecret, RandomSecret, PKI, Database, VaultSecret v2 tests all create SecretEngineMount CRs)

## Dev Notes

### SecretEngineMount Uses the VaultEngineResource Reconciler — NOT VaultResource

This is the critical difference from Stories 3.1 (Policy) and 3.2 (PasswordPolicy). SecretEngineMount uses `NewVaultEngineResource` (not `NewVaultResource`), which has a fundamentally different reconcile flow:

1. `prepareContext()` enriches context with kubeClient, vaultClient, etc.
2. `NewVaultEngineResource(&r.ReconcilerBase, instance)` creates the engine reconciler
3. `VaultEngineResource.Reconcile()` → `manageReconcileLogic()`:
   - `PrepareInternalValues()` — no-op for SecretEngineMount
   - `PrepareTLSConfig()` — no-op for SecretEngineMount
   - `Exists()` — reads `sys/mounts` to check if the engine is already mounted
   - If NOT found → `Create()` — enables the engine at the computed path
   - If found → `CreateOrUpdateTuneConfig()` — reads current tune config, calls `IsEquivalentToDesiredState()`, writes tune config only if different
   - `GetAccessor()` — reads `sys/mounts` to extract the accessor string for the mount
   - `SetAccessor(accessor)` — stores the accessor in `status.accessor`
4. `ManageOutcome()` sets `ReconcileSuccessful` condition

**This means:** The engine is created once (enable) and then only the tune config is updated on subsequent reconciles. The `GetPayload()` (full mount spec) is used for the initial enable, but `IsEquivalentToDesiredState()` compares against the **tune config only** (`Config.toMap()` with `options` and `description` deleted).

[Source: controllers/secretenginemount_controller.go#L73-L75 — NewVaultEngineResource usage]
[Source: controllers/vaultresourcecontroller/vaultengineresourcereconciler.go#L94-L134 — manageReconcileLogic]
[Source: api/v1alpha1/utils/vaultengineobject.go#L79-L97 — Exists, CreateOrUpdateTuneConfig]

### IsEquivalentToDesiredState — Tune-Only Comparison

SecretEngineMount's `IsEquivalentToDesiredState` compares only the **tune config**, not the full mount spec:

```go
func (d *SecretEngineMount) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    configMap := d.Spec.Config.toMap()
    delete(configMap, "options")
    delete(configMap, "description")
    return reflect.DeepEqual(configMap, payload)
}
```

The `payload` parameter is the Vault tune endpoint read response, which contains fields like `default_lease_ttl`, `max_lease_ttl`, `force_no_cache`, etc. The `options` and `description` keys are deleted because they are not part of the tune response.

**Vault tune response returns TTLs as integer seconds**, not duration strings. For example, `maxLeaseTTL: "8760h"` in the spec becomes `max_lease_ttl: 31536000` in the Vault response. However, `IsEquivalentToDesiredState` compares against `Config.toMap()` which has the string value `"8760h"`. This means the comparison likely returns `false` on re-read (the string `"8760h"` != integer `31536000`), causing a tune write on every reconcile. This is tracked tech debt (Story 7-4).

**For test verification**, when reading the tune config from Vault, expect integer seconds for TTL fields, not string durations.

[Source: api/v1alpha1/secretenginemount_types.go#L63-L68 — IsEquivalentToDesiredState]

### GetPath() — Mount Path Computation

```go
func (d *SecretEngineMount) GetPath() string {
    // GetEngineListPath() returns "sys/mounts"
    // If path is empty:
    //   if spec.name set → sys/mounts/{spec.name}
    //   else → sys/mounts/{metadata.name}
    // If path is set:
    //   if spec.name set → sys/mounts/{path}/{spec.name}
    //   else → sys/mounts/{path}/{metadata.name}
}
```

For fixtures without `spec.path`:
- `test-kv-mount` → `sys/mounts/test-kv-mount`
- `test-tuned-kv-mount` → `sys/mounts/test-tuned-kv-mount`
- `test-named-sem-metadata` with `spec.name: test-named-kv-mount` → `sys/mounts/test-named-kv-mount`

[Source: api/v1alpha1/secretenginemount_types.go#L43-L59 — GetPath]

### Webhook Validation — Immutable Path AND Mount Fields

SecretEngineMount's webhook has **two validation rules** in `ValidateUpdate`:

1. `spec.path` cannot be changed (standard immutable path rule)
2. **Only `spec.config` can be modified** — all other `Mount` fields (`type`, `description`, `local`, `sealWrap`, `externalEntropyAccess`, `options`) are immutable after creation

This is enforced by zeroing out `Config` on both old and new `Mount` structs and comparing with `reflect.DeepEqual`. If any non-config Mount field differs, the webhook rejects the update.

This validation is NOT tested in this story (webhook tests are Epic 7, Story 7-1), but it's important context for why the tune-only comparison makes sense.

[Source: api/v1alpha1/secretenginemount_webhook.go#L64-L79 — ValidateUpdate]

### Vault API Response Shapes

**`GET sys/mounts`** — Returns all mounts as a flat map. Keys have a trailing slash:
```json
{
  "data": {
    "test-kv-mount/": {
      "type": "kv",
      "accessor": "kv_abc123",
      "config": { "default_lease_ttl": 0, "max_lease_ttl": 0, ... },
      "options": { "version": "2" },
      ...
    },
    "sys/": { ... },
    "secret/": { ... }
  }
}
```

The `Exists()` method iterates over keys, trimming trailing `/`, and compares against the mount path (also trimmed). The accessor is extracted from the matching mount data.

**`GET sys/mounts/{path}/tune`** — Returns tune-level config:
```json
{
  "data": {
    "default_lease_ttl": 0,
    "max_lease_ttl": 31536000,
    "force_no_cache": false,
    "listing_visibility": "unauth",
    ...
  }
}
```

TTL values are returned as **integer seconds** (e.g., `31536000` for "8760h", `0` for default/empty). String fields are returned as strings.

### Vault Auth Setup — policy-admin Role

The `policy-admin` Kubernetes auth role in `vault-admin` namespace has the `vault-admin` policy with full `/*` access:
```
path "/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
```

This grants access to `sys/mounts/*` needed for enabling, tuning, and disabling engines. All SecretEngineMount fixtures use `authentication.role: policy-admin` and must be created in the `vault-admin` namespace.

[Source: integration/vault-values.yaml#L167-L171]

### Controller Registration — Already Done

The SecretEngineMount controller is already registered in `suite_integration_test.go`:
```go
err = (&SecretEngineMountReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "SecretEngineMount")}).SetupWithManager(mgr)
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L142-L143]

### Decoder — GetSecretEngineMountInstance Already Exists

`GetSecretEngineMountInstance` exists in `controllers/controllertestutils/decoder.go` at lines 100-112. No decoder changes needed.

[Source: controllers/controllertestutils/decoder.go#L100-L112]

### SecretEngineMount Is Already Used as a Dependency in Other Tests

SecretEngineMount CRs are created as prerequisites in:
- `vaultsecret_controller_test.go` — `test/kv-secret-engine.yaml` (kv at `test-vault-config-operator/kv`)
- `vaultsecret_controller_v2_test.go` — `test/randomsecret/v2/03-secretenginemount-kv-v2.yaml` (kv at `test-vault-config-operator/kv-v2`)
- `randomsecret_controller_test.go` — same v2 fixture across 4 contexts
- `pkisecretengine_controller_test.go` — `test/pkisecretengine/pki-secret-engine.yaml` (pki at `test-vault-config-operator/pki`)
- `databasesecretenginestaticrole_controller_test.go` — `test/database-secret-engine.yaml` and `test/databasesecretengine/database-kv-engine-mount.yaml`

These tests prove SecretEngineMount CRs can be created and reconciled, but they never:
1. Verify the mount exists in Vault via `sys/mounts` read
2. Verify the tune config matches the spec
3. Verify the accessor is populated in status
4. Verify Vault mount is removed after CR deletion
5. Test the `spec.name` override behavior

### Test Fixture Design

**Fixture 1: `test/secretenginemount/simple-kv-mount.yaml`** — Minimal mount, tests basic enable flow:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: test-kv-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: kv
  options:
    version: "2"
```

No `spec.path` → mounts at `sys/mounts/test-kv-mount`. No tune config overrides → uses Vault defaults.

**Fixture 2: `test/secretenginemount/tuned-kv-mount.yaml`** — Tests tune config:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: test-tuned-kv-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: kv
  config:
    maxLeaseTTL: "8760h"
    listingVisibility: "unauth"
  options:
    version: "2"
```

Custom `maxLeaseTTL` and `listingVisibility` → verify tune endpoint reflects these values.

**Fixture 3: `test/secretenginemount/named-kv-mount.yaml`** — Tests spec.name override:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: test-named-sem-metadata
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: test-named-kv-mount
  type: kv
  options:
    version: "2"
```

`spec.name: test-named-kv-mount` → mounts at `sys/mounts/test-named-kv-mount`, NOT `sys/mounts/test-named-sem-metadata`.

### Verifying Vault State

**Mount existence verification:**
```go
secret, err := vaultClient.Logical().Read("sys/mounts")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
_, exists := secret.Data["test-kv-mount/"]
Expect(exists).To(BeTrue(), "expected mount 'test-kv-mount/' in sys/mounts")
```

Note the trailing `/` on the mount key — Vault always appends `/` to mount keys.

**Accessor verification:**
```go
Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())
Expect(created.Status.Accessor).NotTo(BeEmpty())
```

The accessor is a string like `kv_abc123` set by the reconciler after reading `sys/mounts`.

**Tune config verification:**
```go
tuneSecret, err := vaultClient.Logical().Read("sys/mounts/test-tuned-kv-mount/tune")
Expect(err).To(BeNil())
Expect(tuneSecret).NotTo(BeNil())
```

For TTL values: Vault returns integer seconds in the tune response. `"8760h"` = 8760 * 3600 = 31536000 seconds. Use `json.Number` or type-assert to `json.Number` then convert. The Vault Go client returns `json.Number` for numeric values when using `Logical().Read()`.

```go
maxLeaseTTL, ok := tuneSecret.Data["max_lease_ttl"].(json.Number)
Expect(ok).To(BeTrue())
maxLeaseTTLInt, err := maxLeaseTTL.Int64()
Expect(err).To(BeNil())
Expect(maxLeaseTTLInt).To(Equal(int64(31536000)))
```

Alternatively, the Vault client may return the value as a `json.Number` string that can be compared:
```go
Expect(tuneSecret.Data["max_lease_ttl"]).To(BeEquivalentTo(json.Number("31536000")))
```

For `listing_visibility`:
```go
Expect(tuneSecret.Data["listing_visibility"]).To(Equal("unauth"))
```

**Delete verification:**
```go
Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("sys/mounts")
    if err != nil || secret == nil {
        return false
    }
    _, exists := secret.Data["test-kv-mount/"]
    return !exists
}, timeout, interval).Should(BeTrue())
```

### Name Collision Prevention

Fixture names use the `test-` prefix and unique names that don't collide with existing SecretEngineMount CRs:
- `kv` — used by vaultsecret_controller_test.go
- `kv-v2` — used by randomsecret and vaultsecret v2 tests
- `kv-db` — used by databasesecretenginestaticrole tests
- `database` — used by database tests
- `pki` — used by pki tests
- `test-kv-mount` — this story (fixture 1)
- `test-tuned-kv-mount` — this story (fixture 2)
- `test-named-kv-mount` — this story (fixture 3, via spec.name)

### MountConfig.toMap() Field Mapping

```go
func (mc *MountConfig) toMap() map[string]interface{} {
    return map[string]interface{}{
        "default_lease_ttl":            mc.DefaultLeaseTTL,
        "max_lease_ttl":                mc.MaxLeaseTTL,
        "force_no_cache":               mc.ForceNoCache,
        "audit_non_hmac_request_keys":  mc.AuditNonHMACRequestKeys,
        "audit_non_hmac_response_keys": mc.AuditNonHMACResponseKeys,
        "listing_visibility":           mc.ListingVisibility,
        "passthrough_request_headers":  mc.PassthroughRequestHeaders,
        "allowed_response_headers":     mc.AllowedResponseHeaders,
    }
}
```

These are the tune-level fields the reconciler manages. All use snake_case matching the Vault API.

[Source: api/v1alpha1/secretenginemount_types.go#L255-L266]

### Status.Accessor — Unique to Engine Mount Types

Unlike Policy or PasswordPolicy, SecretEngineMount has `Status.Accessor` which stores the Vault engine accessor string. This is populated by the reconciler after the engine is created and confirmed. The accessor is used by other types (like Policy with `${auth/kubernetes/@accessor}` placeholders) to reference the engine.

The test should verify this field is populated after successful reconcile — it's a key contract of the engine mount types.

[Source: api/v1alpha1/secretenginemount_types.go#L96-L98 — SetAccessor]
[Source: api/v1alpha1/secretenginemount_types.go#L218-L219 — Status.Accessor field]

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `test/secretenginemount/simple-kv-mount.yaml` | New | Minimal KV v2 mount fixture (no path, no tune) |
| 2 | `test/secretenginemount/tuned-kv-mount.yaml` | New | KV v2 mount with maxLeaseTTL and listingVisibility |
| 3 | `test/secretenginemount/named-kv-mount.yaml` | New | KV v2 mount with spec.name override |
| 4 | `controllers/secretenginemount_controller_test.go` | New | Integration test — create, verify mount + accessor + tune, spec.name, delete |

No changes to decoder, suite setup, controllers, or types.

### No `make manifests generate` Needed

This story only adds integration test files and YAML fixtures. No CRD types, controllers, or webhooks are changed.

### Import Requirements for secretenginemount_controller_test.go

```go
import (
    "encoding/json"
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

The `encoding/json` import is needed for `json.Number` type assertions when verifying Vault tune config TTL values. All others are already indirect dependencies — no `go get` needed.

### Test Structure

```
Describe("SecretEngineMount controller")
  Context("When creating a simple SecretEngineMount")
    It("Should enable the engine in Vault and populate the accessor")
      - Load fixture, set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read sys/mounts, verify "test-kv-mount/" key exists
      - Verify mount data has type "kv" and options.version "2"
      - Verify status.accessor is non-empty on the CR
  Context("When creating a SecretEngineMount with tune config")
    It("Should apply the tune config in Vault")
      - Load fixture, set namespace, create
      - Eventually poll for ReconcileSuccessful=True
      - Read sys/mounts/test-tuned-kv-mount/tune
      - Verify max_lease_ttl = 31536000 (8760h in seconds)
      - Verify listing_visibility = "unauth"
  Context("When creating a SecretEngineMount with spec.name override")
    It("Should mount at the spec.name path")
      - Load fixture, set namespace, create
      - Eventually poll for ReconcileSuccessful=True
      - Read sys/mounts, verify "test-named-kv-mount/" exists
      - Verify "test-named-sem-metadata/" does NOT exist
  Context("When deleting SecretEngineMounts")
    It("Should disable the engines in Vault")
      - Delete all three SecretEngineMount CRs
      - Eventually poll for K8s NotFound on each
      - Verify all three mounts are gone from sys/mounts
```

### Risk Considerations

- **TTL type assertion:** Vault returns TTL values as `json.Number` from the Go client. Use `json.Number` type assertion, not `float64` or `int`. If the Vault client version handles this differently, fall back to `fmt.Sprintf` comparison.
- **Mount key trailing slash:** Vault mount keys always have a trailing `/`. When checking `sys/mounts` response, look for `"test-kv-mount/"` not `"test-kv-mount"`.
- **Namespace isolation:** All fixtures use `vault-admin` namespace with `policy-admin` role. Existing tests use `test-vault-config-operator` namespace with specific roles. No cross-contamination.
- **Test ordering:** Ginkgo v2 runs Contexts sequentially within a Describe by default. The create Contexts must complete before the delete Context runs.
- **Existing mount collision:** Fixture names are verified unique against all existing test mounts (kv, kv-v2, kv-db, database, pki, raf-backstage-demo, kubese, rabbitmq).
- **VaultEngineEndpoint.Exists() reads sys/mounts:** The `Exists()` method reads the full mounts list and iterates. If the test creates mounts in parallel with other tests that also create mounts, there's no conflict because each mount has a unique path.

### Previous Story Intelligence

**From Story 3.2 (PasswordPolicy integration tests):**
- Established the generate endpoint verification pattern (functional test beyond just data verification)
- Used `ContainSubstring` for policy text verification
- Tested `spec.name` override → Vault path uses `spec.name`
- Entity test (`entity_controller_test.go`) is the simplest standalone pattern reference

**From Story 3.1 (Policy integration tests):**
- Established the Epic 3 pattern: create fixture → create CR → poll ReconcileSuccessful → verify Vault state → delete → verify Vault cleanup
- Tested both Vault API path variants
- Verified accessor placeholder resolution (relevant because SecretEngineMount accessors are what Policy placeholders resolve to)

**From Epic 2 Retrospective:**
- "Pattern-first investment paid off" — follow established create/verify/delete pattern
- "Epic 3 scope is simpler than Epic 2: create/delete lifecycle tests for 4 foundation types"
- Prefer Opus-class models for implementation

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

- Test file goes in `controllers/secretenginemount_controller_test.go` (standard controller test location)
- Test fixtures go in `test/secretenginemount/` directory (follows `test/<feature>/` pattern)
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/secretenginemount_types.go] — SecretEngineMount VaultObject + VaultEngineObject implementation, GetPath, GetPayload, IsEquivalentToDesiredState, MountConfig.toMap, Mount.toMap
- [Source: api/v1alpha1/secretenginemount_types.go#L100-L125] — SecretEngineMountSpec fields
- [Source: api/v1alpha1/secretenginemount_types.go#L128-L160] — Mount struct fields
- [Source: api/v1alpha1/secretenginemount_types.go#L163-L207] — MountConfig struct fields
- [Source: api/v1alpha1/secretenginemount_types.go#L209-L219] — Status with Accessor field
- [Source: api/v1alpha1/secretenginemount_webhook.go] — Webhook: immutable path + only config modifiable
- [Source: controllers/secretenginemount_controller.go] — Controller (VaultEngineResource)
- [Source: controllers/vaultresourcecontroller/vaultengineresourcereconciler.go#L94-L134] — VaultEngineResource.manageReconcileLogic
- [Source: api/v1alpha1/utils/vaultengineobject.go] — VaultEngineEndpoint: Exists, Create, CreateOrUpdateTuneConfig, GetAccessor
- [Source: controllers/suite_integration_test.go#L142-L143] — SecretEngineMount controller registration
- [Source: controllers/controllertestutils/decoder.go#L100-L112] — GetSecretEngineMountInstance decoder method
- [Source: controllers/entity_controller_test.go] — Simplest standalone integration test pattern reference
- [Source: controllers/vaultsecret_controller_test.go#L198-L214] — SecretEngineMount used as dependency
- [Source: integration/vault-values.yaml#L167-L171] — policy-admin Vault auth role setup
- [Source: _bmad-output/planning-artifacts/epics.md#L389-L405] — Story 3.3 epic definition
- [Source: _bmad-output/implementation-artifacts/3-2-integration-tests-for-passwordpolicy-type.md] — Previous story (pattern reference)
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
