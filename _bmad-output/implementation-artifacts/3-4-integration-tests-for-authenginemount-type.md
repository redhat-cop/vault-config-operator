# Story 3.4: Integration Tests for AuthEngineMount Type

Status: review

## Story

As an operator developer,
I want integration tests for the AuthEngineMount type covering create, tune verification, accessor status, spec.name override, and delete with Vault cleanup,
So that the foundation for all auth engine types is verified end-to-end.

## Acceptance Criteria

1. **Given** an AuthEngineMount CR is created in the test namespace with type=approle **When** the reconciler processes it **Then** the auth method is enabled in Vault at `sys/auth/{path}/{name}`, the accessor is populated in `status.accessor`, and ReconcileSuccessful=True

2. **Given** an AuthEngineMount CR is created with non-default tune config (maxLeaseTTL, listingVisibility) **When** the reconciler processes it **Then** the tune config in Vault at `{path}/tune` matches the specified values

3. **Given** an AuthEngineMount CR using `spec.name` (overriding `metadata.name`) **When** the reconciler processes it **Then** the auth method is mounted at a path derived from `spec.name` (not `metadata.name`)

4. **Given** a successfully mounted AuthEngineMount is deleted **When** the reconciler processes the deletion **Then** the auth method is disabled from Vault and the finalizer is cleared

## Tasks / Subtasks

- [x] Task 1: Add decoder method (AC: 1, 2, 3)
  - [x] 1.1: Add `GetAuthEngineMountInstance` method to `controllers/controllertestutils/decoder.go` following the existing pattern (decode YAML → type-assert to `*redhatcopv1alpha1.AuthEngineMount`)

- [x] Task 2: Create test fixtures (AC: 1, 2, 3)
  - [x] 2.1: Create `test/authenginemount/simple-approle-mount.yaml` — a minimal AuthEngineMount CR with `metadata.name: test-aem-simple`, `type: approle`, `path: test-auth-mount`, using `authentication.role: policy-admin`
  - [x] 2.2: Create `test/authenginemount/tuned-approle-mount.yaml` — an AuthEngineMount CR with `metadata.name: test-aem-tuned`, `type: approle`, `path: test-auth-mount`, `config.maxLeaseTTL: "8760h"`, `config.listingVisibility: "unauth"`, using `authentication.role: policy-admin`
  - [x] 2.3: Create `test/authenginemount/named-approle-mount.yaml` — an AuthEngineMount CR with `metadata.name: test-aem-metadata`, `spec.name: test-aem-named`, `type: approle`, `path: test-auth-mount`, using `authentication.role: policy-admin`

- [x] Task 3: Create integration test file (AC: 1, 2, 3, 4)
  - [x] 3.1: Create `controllers/authenginemount_controller_test.go` with `//go:build integration` tag, package `controllers`, standard Ginkgo imports
  - [x] 3.2: Add `Describe("AuthEngineMount controller")` with `timeout := 120 * time.Second`, `interval := 2 * time.Second`
  - [x] 3.3: Add `Context("When creating a simple AuthEngineMount")` — load `simple-approle-mount.yaml` via `decoder.GetAuthEngineMountInstance`, set namespace to `vaultAdminNamespaceName`, create it, poll for `ReconcileSuccessful=True`
  - [x] 3.4: After reconcile success, verify the auth mount exists in Vault by reading `sys/auth` and checking for the key `test-auth-mount/test-aem-simple/` (trailing slash)
  - [x] 3.5: Verify `status.accessor` is populated (non-empty string) on the CR after reconcile
  - [x] 3.6: Add `Context("When creating an AuthEngineMount with tune config")` — load `tuned-approle-mount.yaml`, set namespace, create, poll for `ReconcileSuccessful=True`
  - [x] 3.7: After reconcile success, read the tune config from Vault via `vaultClient.Logical().Read("sys/auth/test-auth-mount/test-aem-tuned/tune")`, verify `secret.Data["max_lease_ttl"]` equals 31536000 (integer seconds for "8760h") and `secret.Data["listing_visibility"]` equals `"unauth"`
  - [x] 3.8: Add `Context("When creating an AuthEngineMount with spec.name override")` — load `named-approle-mount.yaml`, set namespace, create, poll for `ReconcileSuccessful=True`
  - [x] 3.9: After reconcile success, verify the mount exists at key `test-auth-mount/test-aem-named/` in `sys/auth` (not `test-auth-mount/test-aem-metadata/`)
  - [x] 3.10: Add `Context("When deleting AuthEngineMounts")` — delete all three AuthEngineMount CRs, use `Eventually` to poll for K8s deletion (NotFound error), then verify the mounts no longer exist in `sys/auth` output
  - [x] 3.11: Verify the finalizer was cleared by confirming deletion completes (the `Eventually` for NotFound confirms this)

- [x] Task 4: End-to-end verification (AC: 1, 2, 3, 4)
  - [x] 4.1: Run `make integration` and verify the new AuthEngineMount tests pass alongside all existing tests
  - [x] 4.2: Verify no regressions — the existing `kubernetes` auth mount at `auth/kubernetes` is unaffected by the new test mounts at `auth/test-auth-mount/*`

### Review Findings

- [x] [Review][Patch] Assert auth mount tune TTL as a numeric value instead of string formatting [`controllers/authenginemount_controller_test.go:120`]

## Dev Notes

### AuthEngineMount Uses the VaultEngineResource Reconciler — NOT VaultResource

This is the same reconcile variant as SecretEngineMount (Story 3.3). AuthEngineMount uses `NewVaultEngineResource` (not `NewVaultResource`), which has a fundamentally different reconcile flow:

1. `prepareContext()` enriches context with kubeClient, vaultClient, etc.
2. `NewVaultEngineResource(&r.ReconcilerBase, instance)` creates the engine reconciler
3. `VaultEngineResource.Reconcile()` → `manageReconcileLogic()`:
   - `PrepareInternalValues()` — no-op for AuthEngineMount
   - `PrepareTLSConfig()` — no-op for AuthEngineMount
   - `Exists()` — reads `sys/auth` to check if the auth method is already mounted
   - If NOT found → `Create()` — enables the auth method at the computed path
   - If found → `CreateOrUpdateTuneConfig()` — reads current tune config, calls `IsEquivalentToDesiredState()`, writes tune config only if different
   - `GetAccessor()` — reads `sys/auth` to extract the accessor string for the mount
   - `SetAccessor(accessor)` — stores the accessor in `status.accessor`
4. `ManageOutcome()` sets `ReconcileSuccessful` condition

[Source: controllers/authenginemount_controller.go#L57-L78 — Reconcile flow]
[Source: controllers/vaultresourcecontroller/vaultengineresourcereconciler.go#L94-L134 — manageReconcileLogic]
[Source: api/v1alpha1/utils/vaultengineobject.go#L79-L97 — Exists, CreateOrUpdateTuneConfig]

### Key Difference from SecretEngineMount: `IsEquivalentToDesiredState`

AuthEngineMount's `IsEquivalentToDesiredState` does NOT delete any keys from the config map:

```go
func (d *AuthEngineMount) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    configMap := d.Spec.Config.toMap()
    return reflect.DeepEqual(configMap, payload)
}
```

Contrast with SecretEngineMount which deletes `options` and `description` before comparison. The reason: `AuthMountConfig.toMap()` already includes `options`, `description`, and `token_type` in its output (10 fields total), and all of these are tune-level fields for auth mounts.

[Source: api/v1alpha1/authenginemount_types.go#L183-L186 — IsEquivalentToDesiredState]

### AuthMountConfig.toMap() — Full Field Set (10 Fields)

```go
func (mc *AuthMountConfig) toMap() map[string]interface{} {
    return map[string]interface{}{
        "default_lease_ttl":            mc.DefaultLeaseTTL,
        "max_lease_ttl":                mc.MaxLeaseTTL,
        "audit_non_hmac_request_keys":  mc.AuditNonHMACRequestKeys,
        "audit_non_hmac_response_keys": mc.AuditNonHMACResponseKeys,
        "listing_visibility":           mc.ListingVisibility,
        "passthrough_request_headers":  mc.PassthroughRequestHeaders,
        "allowed_response_headers":     mc.AllowedResponseHeaders,
        "token_type":                   mc.TokenType,
        "description":                  mc.Description,
        "options":                      mc.Options,
    }
}
```

Note three extra fields vs SecretEngineMount's `MountConfig.toMap()`: `token_type`, `description` (as `*string`), and `options` (as `map[string]string`).

[Source: api/v1alpha1/authenginemount_types.go#L147-L160]

### GetPath() — Auth Mount Path Composition

```go
func (d *AuthEngineMount) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath(d.GetEngineListPath() + "/" + string(d.Spec.Path) + "/" + d.Spec.Name)
    }
    return vaultutils.CleansePath(d.GetEngineListPath() + "/" + string(d.Spec.Path) + "/" + d.Name)
}
```

`GetEngineListPath()` returns `"sys/auth"`. `spec.path` is Required for AuthEngineMount.

For test fixtures:
- `test-aem-simple` with `path: test-auth-mount` → `sys/auth/test-auth-mount/test-aem-simple`
- `test-aem-tuned` with `path: test-auth-mount` → `sys/auth/test-auth-mount/test-aem-tuned`
- `test-aem-metadata` with `spec.name: test-aem-named`, `path: test-auth-mount` → `sys/auth/test-auth-mount/test-aem-named`

[Source: api/v1alpha1/authenginemount_types.go#L60-L65]

### Webhook Validation — Immutable Path AND Mount Fields

AuthEngineMount's webhook has **two validation rules** in `ValidateUpdate`:

1. `spec.path` cannot be changed (standard immutable path rule)
2. **Only `spec.config` can be modified** — all other `AuthMount` fields (`type`, `description`, `local`, `sealWrap`) are immutable after creation

This validation is NOT tested in this story (webhook tests are Epic 7, Story 7-1), but it's important context for why the tune-only comparison makes sense.

[Source: api/v1alpha1/authenginemount_webhook.go#L63-L78 — ValidateUpdate]

### Vault API Response Shapes

**`GET sys/auth`** — Returns all auth mounts as a flat map. Keys have a trailing slash:
```json
{
  "data": {
    "kubernetes/": {
      "type": "kubernetes",
      "accessor": "auth_kubernetes_abc123",
      "config": { "default_lease_ttl": 0, "max_lease_ttl": 0, ... },
      ...
    },
    "token/": { ... },
    "test-auth-mount/test-aem-simple/": {
      "type": "approle",
      "accessor": "auth_approle_xyz789",
      "config": { ... },
      ...
    }
  }
}
```

The `retrieveAccessor()` method iterates over keys, trimming trailing `/`, and compares against the path component after `sys/auth/` (also trimmed). For fixture `test-aem-simple` with `path: test-auth-mount`:
- GetPath() = `sys/auth/test-auth-mount/test-aem-simple`
- TrimPrefix with `sys/auth` → `/test-auth-mount/test-aem-simple`
- Trim `/` → `test-auth-mount/test-aem-simple`
- Key in response: `test-auth-mount/test-aem-simple/`
- Trim `/` → `test-auth-mount/test-aem-simple`
- Match!

**`GET sys/auth/{path}/tune`** — Returns tune-level config:
```json
{
  "data": {
    "default_lease_ttl": 0,
    "max_lease_ttl": 31536000,
    "force_no_cache": false,
    "listing_visibility": "unauth",
    "token_type": "default-service",
    ...
  }
}
```

TTL values are returned as **integer seconds**. The tune path for auth mounts is `sys/auth/{mount-path}/tune`, e.g. `sys/auth/test-auth-mount/test-aem-tuned/tune`.

### Why `approle` Type for Test Fixtures

The `approle` auth method is chosen because:
1. **Built-in** — available in all Vault versions, no external service dependencies
2. **No configuration required** — unlike `kubernetes` (needs API server config) or `ldap` (needs LDAP server)
3. **Does not conflict** with the existing `kubernetes` auth mount at `auth/kubernetes` used by the test infrastructure
4. **Simple lifecycle** — enable, tune, disable — exactly what we're testing

Per the project's integration test philosophy: approle needs no external service → it fits the "no dependency" category.

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

### Vault Auth Setup — policy-admin Role

The `policy-admin` Kubernetes auth role in `vault-admin` namespace has the `vault-admin` policy with full `/*` access:
```
path "/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
```

This grants access to `sys/auth/*` needed for enabling, tuning, and disabling auth methods. All AuthEngineMount fixtures use `authentication.role: policy-admin` and must be created in the `vault-admin` namespace.

[Source: integration/vault-values.yaml#L167-L171]

### Controller Registration — Already Done

The AuthEngineMount controller is already registered in `suite_integration_test.go`:
```go
err = (&AuthEngineMountReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AuthEngineMount")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L148-L149]

### Decoder — GetAuthEngineMountInstance MUST BE ADDED

`GetAuthEngineMountInstance` does **not** exist in `controllers/controllertestutils/decoder.go`. It must be added following the established pattern:

```go
func (d *decoder) GetAuthEngineMountInstance(filename string) (*redhatcopv1alpha1.AuthEngineMount, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }

    kind := reflect.TypeOf(redhatcopv1alpha1.AuthEngineMount{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.AuthEngineMount)
        return o, nil
    }

    return nil, errDecode
}
```

[Source: controllers/controllertestutils/decoder.go — existing pattern at lines 100-112]

### Test Fixture Design

**Fixture 1: `test/authenginemount/simple-approle-mount.yaml`** — Minimal mount, tests basic enable flow:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: test-aem-simple
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: approle
  path: test-auth-mount
```

Mounts at `sys/auth/test-auth-mount/test-aem-simple`. No tune config overrides → uses Vault defaults.

**Fixture 2: `test/authenginemount/tuned-approle-mount.yaml`** — Tests tune config:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: test-aem-tuned
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: approle
  path: test-auth-mount
  config:
    maxLeaseTTL: "8760h"
    listingVisibility: "unauth"
```

Custom `maxLeaseTTL` and `listingVisibility` → verify tune endpoint reflects these values.

**Fixture 3: `test/authenginemount/named-approle-mount.yaml`** — Tests spec.name override:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: test-aem-metadata
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: test-aem-named
  type: approle
  path: test-auth-mount
```

`spec.name: test-aem-named` → mounts at `sys/auth/test-auth-mount/test-aem-named`, NOT `sys/auth/test-auth-mount/test-aem-metadata`.

### Verifying Vault State

**Auth mount existence verification:**
```go
secret, err := vaultClient.Logical().Read("sys/auth")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
_, exists := secret.Data["test-auth-mount/test-aem-simple/"]
Expect(exists).To(BeTrue(), "expected mount 'test-auth-mount/test-aem-simple/' in sys/auth")
```

Note the trailing `/` on the mount key — Vault always appends `/` to mount keys.

**Accessor verification:**
```go
Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())
Expect(created.Status.Accessor).NotTo(BeEmpty())
```

The accessor is a string like `auth_approle_abc123` set by the reconciler after reading `sys/auth`.

**Tune config verification:**
```go
tuneSecret, err := vaultClient.Logical().Read("sys/auth/test-auth-mount/test-aem-tuned/tune")
Expect(err).To(BeNil())
Expect(tuneSecret).NotTo(BeNil())
```

For TTL values: Vault returns integer seconds in the tune response. `"8760h"` = 8760 * 3600 = 31536000 seconds. The Vault Go client returns `json.Number` for numeric values.

```go
maxLeaseTTL, ok := tuneSecret.Data["max_lease_ttl"].(json.Number)
Expect(ok).To(BeTrue())
maxLeaseTTLInt, err := maxLeaseTTL.Int64()
Expect(err).To(BeNil())
Expect(maxLeaseTTLInt).To(Equal(int64(31536000)))
```

For `listing_visibility`:
```go
Expect(tuneSecret.Data["listing_visibility"]).To(Equal("unauth"))
```

**Delete verification:**
```go
Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("sys/auth")
    if err != nil || secret == nil {
        return false
    }
    _, exists := secret.Data["test-auth-mount/test-aem-simple/"]
    return !exists
}, timeout, interval).Should(BeTrue())
```

### Name Collision Prevention

Fixture names use the `test-aem-` prefix and unique names that don't collide with existing auth mounts:
- `kubernetes/` — the default Kubernetes auth mount (used by the integration test infrastructure)
- `token/` — the default token auth mount (always present)
- `kube-authengine-mount-sample/authenginemount-sample` — existing test fixture (test/kube-auth-engine-mount.yaml) but not created in current integration tests
- `test-auth-mount/test-aem-simple` — this story (fixture 1)
- `test-auth-mount/test-aem-tuned` — this story (fixture 2)
- `test-auth-mount/test-aem-named` — this story (fixture 3, via spec.name)

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/controllertestutils/decoder.go` | Modified | Add `GetAuthEngineMountInstance` decoder method |
| 2 | `test/authenginemount/simple-approle-mount.yaml` | New | Minimal approle mount fixture (no tune) |
| 3 | `test/authenginemount/tuned-approle-mount.yaml` | New | Approle mount with maxLeaseTTL and listingVisibility |
| 4 | `test/authenginemount/named-approle-mount.yaml` | New | Approle mount with spec.name override |
| 5 | `controllers/authenginemount_controller_test.go` | New | Integration test — create, verify mount + accessor + tune, spec.name, delete |

No changes to suite setup, controllers, or types.

### No `make manifests generate` Needed

This story only adds an integration test file, YAML fixtures, and a decoder method. No CRD types, controllers, or webhooks are changed.

### Import Requirements for authenginemount_controller_test.go

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
Describe("AuthEngineMount controller")
  Context("When creating a simple AuthEngineMount")
    It("Should enable the auth method in Vault and populate the accessor")
      - Load fixture, set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read sys/auth, verify "test-auth-mount/test-aem-simple/" key exists
      - Verify mount data has type "approle"
      - Verify status.accessor is non-empty on the CR
  Context("When creating an AuthEngineMount with tune config")
    It("Should apply the tune config in Vault")
      - Load fixture, set namespace, create
      - Eventually poll for ReconcileSuccessful=True
      - Read sys/auth/test-auth-mount/test-aem-tuned/tune
      - Verify max_lease_ttl = 31536000 (8760h in seconds)
      - Verify listing_visibility = "unauth"
  Context("When creating an AuthEngineMount with spec.name override")
    It("Should mount at the spec.name path")
      - Load fixture, set namespace, create
      - Eventually poll for ReconcileSuccessful=True
      - Read sys/auth, verify "test-auth-mount/test-aem-named/" exists
      - Verify "test-auth-mount/test-aem-metadata/" does NOT exist
  Context("When deleting AuthEngineMounts")
    It("Should disable the auth methods in Vault")
      - Delete all three AuthEngineMount CRs
      - Eventually poll for K8s NotFound on each
      - Verify all three mounts are gone from sys/auth
```

### Risk Considerations

- **TTL type assertion:** Vault returns TTL values as `json.Number` from the Go client. Use `json.Number` type assertion, not `float64` or `int`. If the Vault client version handles this differently, fall back to `fmt.Sprintf` comparison.
- **Auth mount key format:** Vault auth mount keys always have a trailing `/`. When checking `sys/auth` response, look for `"test-auth-mount/test-aem-simple/"` not `"test-auth-mount/test-aem-simple"`. The path separator IS a `/` within the key (e.g., `test-auth-mount/test-aem-simple/`).
- **Namespace isolation:** All fixtures use `vault-admin` namespace with `policy-admin` role. The existing `kubernetes/` auth mount is unaffected since test fixtures use path `test-auth-mount` which creates mounts under `auth/test-auth-mount/`.
- **Test ordering:** Ginkgo v2 runs Contexts sequentially within a Describe by default. The create Contexts must complete before the delete Context runs.
- **Existing mount collision:** Fixture mount paths (`test-auth-mount/test-aem-*`) are verified unique against existing auth mounts (kubernetes/, token/).
- **Auth engine tune endpoint path:** For auth mounts with nested paths (e.g., `test-auth-mount/test-aem-tuned`), the tune endpoint is `sys/auth/test-auth-mount/test-aem-tuned/tune`. Verify the Vault client correctly handles this path.
- **`IsEquivalentToDesiredState` potential false drift:** AuthEngineMount's `IsEquivalentToDesiredState` compares ALL fields from `AuthMountConfig.toMap()` (including `token_type`, `description`, `options`) against the tune response. If Vault's tune response includes extra fields or different types, this may cause a tune write on every reconcile. This is tracked tech debt (Story 7-4) and does not affect test correctness — the test verifies the CR reconciles and Vault state is correct, not idempotency.

### Previous Story Intelligence

**From Story 3.3 (SecretEngineMount integration tests):**
- Established the VaultEngineResource integration test pattern: create fixture → create CR → poll ReconcileSuccessful → verify mount exists in sys/mounts → verify accessor → verify tune config → delete → verify cleanup
- Demonstrated `json.Number` type assertion for TTL verification
- Tested `spec.name` override → Vault path uses `spec.name`
- Verified trailing `/` on mount keys

**From Story 3.2 (PasswordPolicy integration tests):**
- Established the generate endpoint verification pattern
- Tested `spec.name` override → Vault path uses `spec.name`

**From Story 3.1 (Policy integration tests):**
- Established the Epic 3 pattern: create fixture → create CR → poll ReconcileSuccessful → verify Vault state → delete → verify Vault cleanup

**From Epic 2 Retrospective:**
- "Pattern-first investment paid off" — follow established create/verify/delete pattern
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

- Decoder change in `controllers/controllertestutils/decoder.go` (add one method)
- Test file goes in `controllers/authenginemount_controller_test.go` (standard controller test location)
- Test fixtures go in `test/authenginemount/` directory (follows `test/<feature>/` pattern)
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/authenginemount_types.go] — AuthEngineMount VaultObject + VaultEngineObject implementation, GetPath, GetPayload, IsEquivalentToDesiredState, AuthMountConfig.toMap, AuthMount.toMap
- [Source: api/v1alpha1/authenginemount_types.go#L32-L54] — AuthEngineMountSpec fields
- [Source: api/v1alpha1/authenginemount_types.go#L60-L65] — GetPath computation
- [Source: api/v1alpha1/authenginemount_types.go#L67-L88] — AuthMount struct fields
- [Source: api/v1alpha1/authenginemount_types.go#L90-L141] — AuthMountConfig struct fields
- [Source: api/v1alpha1/authenginemount_types.go#L147-L160] — AuthMountConfig.toMap (10 fields)
- [Source: api/v1alpha1/authenginemount_types.go#L183-L186] — IsEquivalentToDesiredState (no field deletion)
- [Source: api/v1alpha1/authenginemount_types.go#L204-L212] — GetEngineListPath, GetEngineTunePath, GetTunePayload
- [Source: api/v1alpha1/authenginemount_types.go#L218-L229] — Status with Accessor field
- [Source: api/v1alpha1/authenginemount_webhook.go#L63-L78] — Webhook: immutable path + only config modifiable
- [Source: controllers/authenginemount_controller.go] — Controller (VaultEngineResource)
- [Source: controllers/vaultresourcecontroller/vaultengineresourcereconciler.go#L94-L134] — VaultEngineResource.manageReconcileLogic
- [Source: api/v1alpha1/utils/vaultengineobject.go] — VaultEngineEndpoint: Exists, Create, CreateOrUpdateTuneConfig, GetAccessor, retrieveAccessor
- [Source: api/v1alpha1/utils/vaultengineobject.go#L60-L73] — retrieveAccessor key matching logic (trim trailing /, compare)
- [Source: controllers/suite_integration_test.go#L148-L149] — AuthEngineMount controller registration
- [Source: controllers/controllertestutils/decoder.go] — Decoder (GetAuthEngineMountInstance MUST BE ADDED)
- [Source: controllers/entity_controller_test.go] — Simplest standalone integration test pattern reference
- [Source: controllers/secretenginemount_controller_test.go] — Closest analogous test (Story 3.3, VaultEngineResource)
- [Source: integration/vault-values.yaml#L161-L171] — vault-admin-initializer: auth/kubernetes enabled, policy-admin role
- [Source: _bmad-output/planning-artifacts/epics.md#L407-L423] — Story 3.4 epic definition
- [Source: _bmad-output/implementation-artifacts/3-3-integration-tests-for-secretenginemount-type.md] — Previous story (closest pattern reference)
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy
- [Source: _bmad-output/project-context.md#L143-L149] — Integration test pattern
- [Source: _bmad-output/project-context.md#L69-L74] — Engine mount tune-only comparison rule

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (Cursor)

### Debug Log References

None — clean implementation with no failures.

### Completion Notes List

- Added `GetAuthEngineMountInstance` decoder method following the established pattern in `decoder.go`
- Created three test fixtures covering all AC scenarios: simple mount (AC1), tuned mount with maxLeaseTTL and listingVisibility (AC2), and spec.name override mount (AC3)
- All fixtures use `type: approle` (built-in, no external dependency, no configuration required)
- Integration test follows the established VaultEngineResource pattern from Story 3.3 (SecretEngineMount): create → poll ReconcileSuccessful → verify Vault state → delete → verify cleanup
- Verified auth mount existence via `sys/auth` key lookup with trailing `/` (Vault convention)
- Verified `status.accessor` is populated and matches the Vault-side accessor value
- Tune config verified via `sys/auth/{mount-path}/tune` endpoint: `max_lease_ttl=31536000` (8760h in seconds), `listing_visibility="unauth"`
- `spec.name` override verified: mount exists at `test-auth-mount/test-aem-named/` (spec.name), NOT at `test-auth-mount/test-aem-metadata/` (metadata.name)
- Delete context confirms all three mounts removed from Vault and finalizers cleared (K8s NotFound)
- `AfterAll` cleanup guard prevents cascading nil-pointer panics if earlier contexts fail
- All existing integration tests continue to pass with zero regressions; coverage increased from 38.0% to 38.8%

### Change Log

- 2026-04-19: Implemented Story 3.4 — created decoder method, test fixtures, and integration test for AuthEngineMount type

### File List

- `controllers/controllertestutils/decoder.go` (modified) — Added `GetAuthEngineMountInstance` decoder method
- `test/authenginemount/simple-approle-mount.yaml` (new) — Minimal approle mount fixture (no tune config)
- `test/authenginemount/tuned-approle-mount.yaml` (new) — Approle mount with maxLeaseTTL and listingVisibility tune config
- `test/authenginemount/named-approle-mount.yaml` (new) — Approle mount with spec.name override
- `controllers/authenginemount_controller_test.go` (new) — Integration test: create, verify mount + accessor + tune, spec.name override, delete
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified) — Story status updates
- `_bmad-output/implementation-artifacts/3-4-integration-tests-for-authenginemount-type.md` (modified) — Story file updates
