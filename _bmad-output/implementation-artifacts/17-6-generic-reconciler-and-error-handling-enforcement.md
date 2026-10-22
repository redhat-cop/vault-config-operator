# Story 17.6: Generic Reconciler + Error Handling Enforcement

Status: done
## Story

As an operator developer,
I want a generic reconciler that eliminates the duplicated 70-line `Reconcile()` body from ~76 controller files, and comprehensive error checking enforced by linter,
So that new CRDs need only RBAC markers plus a registration call instead of a full controller file, and no error is silently swallowed.

## Acceptance Criteria

### Part A — Generic VaultResource Reconciler

1. **Given** a new CRD type implementing `VaultObject` + `ConditionsAware` **When** a developer registers it with `SetupGenericController[MyType](mgr, "MyType")` **Then** the generic reconciler handles Get, prepareContext, NewVaultResource, Reconcile — no hand-written Reconcile body needed
2. **Given** the generic reconciler is instantiated for all standard-body controllers (~60+ types) **When** all existing unit and integration tests run **Then** all tests pass with identical behavior
3. **Given** a controller with custom Reconcile logic (see exclusion list below) **When** the generic migration runs **Then** these controllers keep their hand-written `Reconcile()` method untouched
4. **Given** the generic reconciler replaces per-type Reconcile bodies **When** `make manifests generate fmt vet test` runs **Then** all unit tests pass and `make lint` has zero violations
5. **Given** engine mount controllers (AuthEngineMount, SecretEngineMount) use `NewVaultEngineResource` **When** the generic reconciler is parameterized **Then** a `VaultResourceFactory` function parameter selects the correct reconciler variant
6. **Given** each standard controller retains its RBAC marker comments **When** `make manifests` generates RBAC **Then** all existing RBAC rules are preserved identically

### Part B — Error Handling Enforcement

7. **Given** the `errcheck` linter is in the `default: standard` preset of `.golangci.yml` **When** `make lint` runs **Then** every function call returning an error has the error checked — zero violations
8. **Given** an error is returned from a Vault API call, K8s client call, or internal helper **When** the caller receives it **Then** the error is logged with context before being returned or handled
9. **Given** the codebase passes `errcheck` **When** CI runs **Then** `errcheck` violations fail the build (already enforced via `make lint` in PR workflow)

## Tasks / Subtasks

### Part A — Generic Reconciler

- [ ] Task 1 — Analyze and classify all 88 controllers (AC: #3)
  - [ ] 1.1 Produce a definitive list of standard-body controllers eligible for generic migration (~60-65 types)
  - [ ] 1.2 Produce a definitive exclusion list of custom-logic controllers that keep hand-written Reconcile (~12 types)
  - [ ] 1.3 Produce a list of controllers with standard Reconcile but custom SetupWithManager (~15 types with credential watches)
- [ ] Task 2 — Create generic reconciler infrastructure (AC: #1, #5)
  - [ ] 2.1 Create `internal/controller/vaultresourcecontroller/generic_reconciler.go` with `GenericVaultReconciler[T, PT]`
  - [ ] 2.2 Implement `VaultResourceFactory` type alias for the factory function parameter
  - [ ] 2.3 Implement generic `Reconcile(ctx, req)` method with identical logic to current standard body
  - [ ] 2.4 Implement `SetupGenericController[T, PT]` registration helper
  - [ ] 2.5 Implement `SetupGenericControllerWithWatches[T, PT]` variant for controllers needing custom watches
- [ ] Task 3 — Migrate standard-body controllers (AC: #2, #6)
  - [ ] 3.1 Migrate pure standard controllers (~60 types): replace Reconcile body with generic, keep RBAC markers + SetupWithManager
  - [ ] 3.2 Migrate standard-Reconcile + custom-SetupWithManager controllers (~15 types): replace only Reconcile body, keep custom SetupWithManager
  - [ ] 3.3 Migrate engine mount controllers (AuthEngineMount, SecretEngineMount) using `VaultEngineResourceFactory`
  - [ ] 3.4 Migrate Audit controller using `VaultAuditResourceFactory`
  - [ ] 3.5 Migrate PKISecretEngineConfig controller using `VaultPKIEngineResourceFactory`
  - [ ] 3.6 Verify RBAC marker comments are preserved in every migrated file
- [ ] Task 4 — Update main.go registration (AC: #1, #2)
  - [ ] 4.1 Replace ~60+ controller registration blocks with generic registration calls
  - [ ] 4.2 Keep hand-written registration for custom-logic controllers
  - [ ] 4.3 Verify webhook registration is unaffected
- [ ] Task 5 — Validate migration (AC: #2, #4)
  - [ ] 5.1 Run `make manifests generate fmt vet` — zero errors
  - [ ] 5.2 Run `make test` — all unit tests pass
  - [ ] 5.3 Run `make lint` — zero violations
  - [ ] 5.4 Diff generated RBAC yaml before/after migration — must be identical
  - [ ] 5.5 Diff generated CRD yaml before/after migration — must be identical

### Part B — Error Handling Enforcement

- [ ] Task 6 — Audit errcheck status (AC: #7)
  - [ ] 6.1 Run `make lint` and capture any `errcheck` violations
  - [ ] 6.2 Cross-reference with R1.2a story (completed 2026-05-15) to identify what was fixed then
  - [ ] 6.3 Identify any new violations introduced since R1.2a (Epic 11-16 code)
- [ ] Task 7 — Fix errcheck violations (AC: #7, #8)
  - [ ] 7.1 Fix all violations in `api/v1alpha1/utils/` (Vault client layer)
  - [ ] 7.2 Fix all violations in `internal/controller/` (reconcilers)
  - [ ] 7.3 Fix all violations in `api/v1alpha1/` (type files)
  - [ ] 7.4 Fix all violations in `cmd/main.go`
  - [ ] 7.5 For each fix: add error logging with context before return/handle
- [ ] Task 8 — Validate errcheck enforcement (AC: #7, #9)
  - [ ] 8.1 Run `make lint` — zero errcheck violations
  - [ ] 8.2 Verify `make lint` is in CI PR workflow (`pr.yaml`) as blocking gate
  - [ ] 8.3 Run `make test` — all tests still pass after fixes
- [ ] Task 9 — Update "Adding a New Vault API" documentation (AC: #1, #2)
  - [ ] 9.1 Update `docs/contributing-vault-apis.md` step 6 (controller): replace hand-written Reconcile body with generic `SetupGenericController` pattern; keep old pattern as "custom controller" alternative
  - [ ] 9.2 Update `docs/contributing-vault-apis.md` step 7 (main.go registration): replace manual `ReconcilerBase` init with `SetupXxxController(mgr)` call
  - [ ] 9.3 Update `_bmad-output/project-context.md` section "Adding a New Vault API Type (Full Checklist)" steps 5–6: reflect generic reconciler registration instead of manual controller + main.go wiring
  - [ ] 9.4 Remove step 7 ("Add decoder method in `controllertestutils/decoder.go`") from project-context checklist — this was already eliminated by R1.4 generic decoder
- [ ] Task 10 — Final validation
  - [ ] 10.1 Run `make manifests generate fmt vet test lint` — all green
  - [ ] 10.2 Verify no behavioral changes in any existing controller

## Dev Notes

### ⚠️ Shared Framework Protection Rule

**This story modifies the reconciler framework layer.** Per project rules, story-scoped work must not **modify existing behavior** in shared framework files. The generic reconciler ADDS new exported types/functions — it does NOT change existing `ReconcileWithFunctions`, `ManageOutcome`, `VaultResource.Reconcile`, or any other existing function behavior. If any behavioral change is needed, escalate to the user before proceeding.

**Shared framework files (do NOT modify existing behavior):**
- `internal/controller/vaultresourcecontroller/reconcile_skeleton.go`
- `internal/controller/vaultresourcecontroller/vaultresourcereconciler.go`
- `internal/controller/vaultresourcecontroller/vaultengineresourcereconciler.go`
- `internal/controller/vaultresourcecontroller/vaultauditresourcereconciler.go`
- `internal/controller/vaultresourcecontroller/vaultpkiengineresourcereconciler.go`
- `internal/controller/vaultresourcecontroller/utils.go`
- `api/v1alpha1/utils/vaultobject.go`
- `internal/controller/commons.go`

### Controller Classification (Complete Inventory)

**88 total controller files** in `internal/controller/`.

#### Category 1: Pure Standard Controllers (~47 types) — Full Generic Migration

These have an identical Reconcile body (Get → NotFound → prepareContext → `NewVaultResource` → Reconcile) and a standard `SetupWithManager` (no custom watches). **Replace both Reconcile body and registration with generic.**

Current standard body pattern (every one of these is identical):
```go
func (r *XxxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    instance := &redhatcopv1alpha1.Xxx{}
    err := r.GetClient().Get(ctx, req.NamespacedName, instance)
    if err != nil {
        if apierrors.IsNotFound(err) { return reconcile.Result{}, nil }
        return reconcile.Result{}, err
    }
    ctx1, err := prepareContext(ctx, r.ReconcilerBase, instance)
    if err != nil {
        r.Log.Error(err, "unable to prepare context", "instance", instance)
        return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
    }
    vaultResource := vaultresourcecontroller.NewVaultResource(&r.ReconcilerBase, instance)
    return vaultResource.Reconcile(ctx1, instance)
}
```

Types in this category (exhaustive list — verify during Task 1):
`Policy`, `PasswordPolicy`, `IdentityOIDCAssignment`, `IdentityOIDCClient`, `IdentityOIDCProvider`, `IdentityOIDCScope`, `IdentityTokenConfig`, `IdentityTokenKey`, `IdentityTokenRole`, `Entity`, `EntityAlias`, `Group`, `GroupAlias`, `Namespace`, `AuditRequestHeader`, `SSHSecretEngineRole`, `AliCloudAuthEngineRole`, `AppRoleAuthEngineRole`, `AWSAuthEngineIdentityConfig`, `AWSAuthEngineRole`, `AWSSecretEngineRole`, `CFAuthEngineRole`, `ConsulSecretEngineRole`, `GCPSecretEngineRoleset`, `GCPSecretEngineStaticAccount`, `GitHubAuthEngineConfig`, `GitHubAuthEngineRUserMap`, `GitHubAuthEngineTeamMap`, `KerberosAuthEngineGroup`, `LDAPSecretEngineDynamicRole`, `LDAPSecretEngineStaticRole`, `MongoDBAtlasSecretEngineRole`, `NomadSecretEngineRole`, `OCIAuthEngineConfig`, `OCIAuthEngineRole`, `OktaAuthEngineGroup`, `RadiusAuthEngineUser`, `TerraformCloudSecretEngineRole`, `AzureSecretEngineRole`, `AzureAuthEngineRole`, `GCPAuthEngineRole`, `DatabaseSecretEngineStaticRole`, `QuaySecretEngineStaticRole`, `CertAuthEngineConfig`, `CertAuthEngineRole`, `QuaySecretEngineRole`, `GitHubSecretEngineRole`, `DatabaseSecretEngineRole`, `RabbitMQSecretEngineRole`, `PKISecretEngineRole`, `JWTOIDCAuthEngineRole`, `LDAPAuthEngineGroup`, `KubernetesSecretEngineRole`

#### Category 2: Standard Reconcile + Custom SetupWithManager (~15 types)

These have the **identical standard Reconcile body** but a custom `SetupWithManager` with `Watches` for credential Secrets/RandomSecrets. **Replace only the Reconcile body; keep the custom SetupWithManager.**

Types (verify during Task 1):
`KubernetesAuthEngineRole` (watches ServiceAccounts), `KubernetesAuthEngineConfig`, `KerberosAuthEngineConfig`, `KerberosAuthEngineConfigLDAP`, `SSHSecretEngineConfig`, `KubernetesSecretEngineConfig`, `GitHubSecretEngineConfig`, `CFAuthEngineConfig`, `RadiusAuthEngineConfig`, `UserpassAuthEngineUser`, `AzureAuthEngineConfig`, `GCPAuthEngineConfig`, `OktaAuthEngineConfig`, `QuaySecretEngineConfig`, `JWTOIDCAuthEngineConfig`, `AzureSecretEngineConfig`, `LDAPAuthEngineConfig`

These controllers have >100 lines because SetupWithManager includes `Watches()` for `corev1.Secret` or `RandomSecret` predicates. The Reconcile body itself is the standard template.

#### Category 3: Engine Mount Controllers (3 types) — Generic with Factory Parameter

Use `NewVaultEngineResource` / `NewVaultPKIEngineResource`:
- `AuthEngineMount` → `VaultEngineResourceFactory`
- `SecretEngineMount` → `VaultEngineResourceFactory`
- `PKISecretEngineConfig` → `VaultPKIEngineResourceFactory`

#### Category 4: Audit Controller (1 type) — Generic with Factory Parameter

- `Audit` → `VaultAuditResourceFactory`

#### Category 5: Fully Custom Controllers (~12 types) — NO MIGRATION

These have custom Reconcile logic that cannot be replaced:

| Controller | Lines | Reason |
|---|---|---|
| `VaultSecret` | 487 | Template rendering, K8s Secret creation, KV merge |
| `RandomSecret` | ~210 | Random secret generation, custom reconcile |
| `DatabaseSecretEngineConfig` | 323 | Root password rotation, credential watches |
| `TransitSecretEngineKey` | 83 | Custom `VaultTransitKeyEndpoint` |
| `RabbitMQSecretEngineConfig` | 131 | Custom lease management, always-write pattern |
| `AWSSecretEngineConfig` | 218 | Always-write pattern (write-only credentials), custom SetupWithManager |
| `ConsulSecretEngineConfig` | ~similar | Same always-write credential pattern |
| `GCPSecretEngineConfig` | ~similar | Same always-write credential pattern |
| `LDAPSecretEngineConfig` | 298 | Custom reconcile with credential handling |
| `MongoDBAtlasSecretEngineConfig` | ~similar | Same always-write credential pattern |
| `NomadSecretEngineConfig` | ~similar | Same always-write credential pattern |
| `TerraformCloudSecretEngineConfig` | ~similar | Same always-write credential pattern |
| `TOTPSecretEngineKey` | ~similar | Custom endpoint |

**These controllers call `ReconcileWithFunctions` directly or use custom reconcile logic (always-write, credential rotation, custom endpoints).** They keep their hand-written Reconcile method unchanged.

### Generic Reconciler Design

#### Core Type (`generic_reconciler.go`)

```go
package vaultresourcecontroller

import (
    "context"

    vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconcilable is implemented by VaultResource, VaultEngineResource, VaultAuditResource, etc.
type Reconcilable interface {
    Reconcile(ctx context.Context, instance client.Object) (ctrl.Result, error)
}

// VaultResourceFactory creates the appropriate reconcilable for a given instance.
type VaultResourceFactory func(base *ReconcilerBase, obj client.Object) Reconcilable

// Standard factories — wrap existing constructors
func StandardVaultResourceFactory(base *ReconcilerBase, obj client.Object) Reconcilable {
    return NewVaultResource(base, obj)
}

func VaultEngineResourceFactory(base *ReconcilerBase, obj client.Object) Reconcilable {
    return NewVaultEngineResource(base, obj)
}

func VaultAuditResourceFactory(base *ReconcilerBase, obj client.Object) Reconcilable {
    return NewVaultAuditResource(base, obj)
}

func VaultPKIEngineResourceFactory(base *ReconcilerBase, obj client.Object) Reconcilable {
    return NewVaultPKIEngineResource(base, obj)
}

// GenericVaultReconciler is a generic reconciler for standard VaultObject types.
// T is the value type (e.g., Policy), PT is the pointer type constraint.
type GenericVaultReconciler[T any, PT interface {
    *T
    client.Object
    vaultutils.VaultObject
    vaultutils.ConditionsAware
}] struct {
    ReconcilerBase
    Factory VaultResourceFactory
}

func (r *GenericVaultReconciler[T, PT]) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    instance := PT(new(T))
    err := r.GetClient().Get(ctx, req.NamespacedName, instance)
    if err != nil {
        if apierrors.IsNotFound(err) {
            return reconcile.Result{}, nil
        }
        return reconcile.Result{}, err
    }

    ctx1, err := prepareContext(ctx, r.ReconcilerBase, instance)
    if err != nil {
        r.Log.Error(err, "unable to prepare context", "instance", instance)
        return ManageOutcome(ctx, r.ReconcilerBase, instance, err)
    }

    resource := r.Factory(&r.ReconcilerBase, instance)
    return resource.Reconcile(ctx1, instance)
}
```

**Critical design constraint:** The `prepareContext` function is in `internal/controller/commons.go` (package `controller`), not in `vaultresourcecontroller`. The generic reconciler lives in `vaultresourcecontroller`. This means either:

**Option A (recommended):** Move `prepareContext` to `vaultresourcecontroller` package (it only depends on `ReconcilerBase` and `VaultObject` — both already in that package). This is a minor refactor that doesn't change behavior.

**Option B:** Keep `prepareContext` in `controller` package and have the generic reconciler accept a `ContextPreparer` function parameter.

**Option C:** Put the generic reconciler in `controller` package alongside `prepareContext`.

**Recommendation:** Option A — move `prepareContext` to `vaultresourcecontroller`. It logically belongs there (it operates on `ReconcilerBase`). This is an additive refactor (the function moves but signature and behavior are identical). Update the ~12 custom controllers that call `prepareContext` directly to use the new package path. Verify with `make test`.

#### Per-Type Controller File (After Migration)

A standard controller file shrinks from ~82 lines to ~20 lines:

```go
package controller

import (
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/builder"
)

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=policies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=policies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=policies/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

func SetupPolicyController(mgr ctrl.Manager) error {
    return vaultresourcecontroller.SetupGenericController[
        redhatcopv1alpha1.Policy,
    ](mgr, "Policy", vaultresourcecontroller.StandardVaultResourceFactory)
}
```

**RBAC marker preservation is critical.** The kubebuilder RBAC markers MUST remain in per-type controller files — they are parsed by `controller-gen` to generate ClusterRole rules. Do NOT move them to a generic file.

#### Registration in `main.go`

Before:
```go
if err = (&controller.PolicyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "Policy")}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "Policy")
    os.Exit(1)
}
```

After:
```go
if err = controller.SetupPolicyController(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "Policy")
    os.Exit(1)
}
```

#### `SetupGenericController` Helper

```go
func SetupGenericController[T any, PT interface {
    *T
    client.Object
    vaultutils.VaultObject
    vaultutils.ConditionsAware
}](mgr manager.Manager, name string, factory VaultResourceFactory) error {
    reconciler := &GenericVaultReconciler[T, PT]{
        ReconcilerBase: NewFromManager(mgr, name),
        Factory:        factory,
    }
    return ctrl.NewControllerManagedBy(mgr).
        For(PT(new(T)), builder.WithPredicates(NewDefaultPeriodicReconcilePredicate())).
        Complete(reconciler)
}
```

#### Controllers with Custom SetupWithManager

These keep a hand-written `SetupXxxController(mgr)` function that creates the generic reconciler but builds the controller manually with custom `Watches`:

```go
func SetupLDAPAuthEngineConfigController(mgr ctrl.Manager) error {
    reconciler := &vaultresourcecontroller.GenericVaultReconciler[
        redhatcopv1alpha1.LDAPAuthEngineConfig,
    ]{
        ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPAuthEngineConfig"),
        Factory:        vaultresourcecontroller.StandardVaultResourceFactory,
    }
    return ctrl.NewControllerManagedBy(mgr).
        For(&redhatcopv1alpha1.LDAPAuthEngineConfig{}, builder.WithPredicates(
            vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
        Watches(&corev1.Secret{...}, handler.EnqueueRequestsFromMapFunc(...), ...).
        Complete(reconciler)
}
```

### prepareContext Refactoring Detail

`prepareContext` currently lives in `internal/controller/commons.go`:

```go
func prepareContext(ctx context.Context, r vaultresourcecontroller.ReconcilerBase, VAR VaultAuthenticableResource) (context.Context, error) {
    // ... enriches context with kubeClient, restConfig, vaultConnection, vaultClient
}
```

The `VaultAuthenticableResource` interface is also in `commons.go`:
```go
type VaultAuthenticableResource interface {
    client.Object
    vaultutils.VaultObject
}
```

**Move both to `vaultresourcecontroller` package** since they operate on `ReconcilerBase` and `VaultObject` — both already there. Update all 12 custom controllers that call `prepareContext` directly. The function signature stays identical.

### Reconcilable Interface Requirement

The existing reconciler types (`VaultResource`, `VaultEngineResource`, `VaultAuditResource`, `VaultPKIEngineResource`) all have a `Reconcile(ctx, instance) (ctrl.Result, error)` method. The `Reconcilable` interface formalizes what already exists. Verify that all four types satisfy the interface with compile-time checks:

```go
var _ Reconcilable = &VaultResource{}
var _ Reconcilable = &VaultEngineResource{}
var _ Reconcilable = &VaultAuditResource{}
var _ Reconcilable = &VaultPKIEngineResource{}
```

### Part B — errcheck Status Assessment

**Current state:** `.golangci.yml` uses `default: standard` (golangci-lint v2), which includes `errcheck`. Story R1.2a (completed 2026-05-15) performed an initial errcheck pass. The project-context.md documents errcheck as mandatory.

**What to verify:**
1. Run `make lint` — count any errcheck violations
2. Scan for `_ =` assignments to error-returning functions (explicit suppression)
3. Scan for `controllerutil.SetControllerReference(owner, obj, r.GetScheme())` with `_ =` (known existing pattern in `utils.go` line 240)
4. Scan for `vaultClient.Logical().Delete(path)` or `.Write(path, data)` without error checks
5. Focus on code added in Epics 11-16 (post-R1.2a)

**Known suppression:** `utils.go:240` has `_ = controllerutil.SetControllerReference(...)`. This should be fixed to check the error and log it.

**Fix pattern:** For every unchecked error:
```go
// Before (violation)
_ = someFunc()

// After (compliant)
if err := someFunc(); err != nil {
    log.Error(err, "descriptive message", "key", value)
    // return err  — or handle appropriately
}
```

### Migration Strategy (Execution Order)

**Phase 1: Infrastructure (Tasks 2, moving prepareContext)**
1. Create `generic_reconciler.go` with types and factory wrappers
2. Move `prepareContext` + `VaultAuthenticableResource` to `vaultresourcecontroller`
3. Update all custom controllers importing `prepareContext`
4. Run `make test` — verify no behavioral change

**Phase 2: Migrate standard controllers (Task 3.1)**
1. Start with a single controller (e.g., `Policy`) as a proof-of-concept
2. Run `make manifests generate fmt vet test` — verify identical behavior
3. Diff RBAC/CRD yamls — verify identical
4. Migrate remaining ~46 pure standard controllers in alphabetical batches
5. Run full test suite after each batch

**Phase 3: Migrate custom-SetupWithManager controllers (Task 3.2)**
1. Migrate one (e.g., `KubernetesAuthEngineConfig`) as proof-of-concept
2. Verify custom watches still work
3. Migrate remaining ~14 types

**Phase 4: Migrate engine/audit controllers (Tasks 3.3-3.5)**
1. Migrate AuthEngineMount, SecretEngineMount, PKISecretEngineConfig, Audit

**Phase 5: Update main.go (Task 4)**
1. Replace registration blocks for all migrated controllers

**Phase 6: errcheck audit and fixes (Tasks 6-8)**
1. Run `make lint`, capture violations
2. Fix all violations
3. Final validation

### File Structure Summary

| File | Action | Purpose |
|---|---|---|
| `internal/controller/vaultresourcecontroller/generic_reconciler.go` | **CREATE** | GenericVaultReconciler[T,PT], VaultResourceFactory, SetupGenericController, factory wrappers |
| `internal/controller/vaultresourcecontroller/utils.go` | **MODIFY** | Add prepareContext + VaultAuthenticableResource (moved from commons.go) |
| `internal/controller/commons.go` | **MODIFY** | Remove prepareContext + VaultAuthenticableResource (moved to vaultresourcecontroller) |
| `internal/controller/*_controller.go` (~65 files) | **MODIFY** | Replace Reconcile body with generic Setup function, keep RBAC markers |
| `internal/controller/*_controller.go` (~12 custom files) | **MODIFY** | Update prepareContext import path only |
| `cmd/main.go` | **MODIFY** | Replace ~65 registration blocks with Setup function calls |
| `.golangci.yml` | **NO CHANGE** | errcheck already enabled via `default: standard` |
| `go.mod` / `go.sum` | **NO CHANGE** | No new dependencies |
| `docs/contributing-vault-apis.md` | **MODIFY** | Update steps 6-7 for generic reconciler pattern |
| `_bmad-output/project-context.md` | **MODIFY** | Update "Adding a New Vault API Type" checklist |
| Various source files | **MODIFY** | Fix errcheck violations (Part B) |

- **Novelty Risk:** MEDIUM — Go generics are well-understood in Go 1.27; the generic body is a mechanical extraction of existing code. The risk is in the volume of files touched (~65 controllers + main.go) and ensuring RBAC markers are preserved.
- Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings. _(Not applicable to this story — no new CRD types.)_

### Project Structure Notes

- Generic reconciler follows established pattern in `internal/controller/vaultresourcecontroller/` — co-located with existing reconciler variants
- Per-type controller files remain in `internal/controller/` — preserves file naming convention and RBAC marker locations
- No new CRD types, no new dependencies, no webhook changes
- The migration is purely structural — all runtime behavior must be identical

### Documentation Updates Required

Two documents describe how to add a new Vault API type. Both must be updated to reflect the generic reconciler.

#### `docs/contributing-vault-apis.md` (public contributor guide)

**Step 6 (Controller)** currently shows the full 30-line hand-written `Reconcile()` + `SetupWithManager()`. Replace with:

```go
// Standard types — use generic reconciler (no hand-written Reconcile body needed):

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=myvaulttypes,...
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=myvaulttypes/status,...
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=myvaulttypes/finalizers,...
//+kubebuilder:rbac:groups="",resources=secrets,...
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,...
//+kubebuilder:rbac:groups="",resources=events,...

func SetupMyVaultTypeController(mgr ctrl.Manager) error {
    return vaultresourcecontroller.SetupGenericController[
        redhatcopv1alpha1.MyVaultType,
    ](mgr, "MyVaultType", vaultresourcecontroller.StandardVaultResourceFactory)
}
```

Keep the old hand-written pattern as a "Custom controller" subsection for types needing custom Reconcile logic (credential watches, rotation, etc.).

**Step 7 (main.go)** currently shows manual `ReconcilerBase` init. Replace with:

```go
if err = controller.SetupMyVaultTypeController(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "MyVaultType")
    os.Exit(1)
}
```

#### `_bmad-output/project-context.md` (AI agent rules)

**"Adding a New Vault API Type (Full Checklist)"** — current 11 steps need updating:

| Step | Current | After |
|---|---|---|
| 5 | Implement controller in `*_controller.go`: embed `ReconcilerBase`, standard reconcile flow | Create `*_controller.go` with RBAC markers + `SetupXxxController` calling `SetupGenericController` |
| 6 | Register controller + webhook in `main.go` | Register via `controller.SetupXxxController(mgr)` + webhook in `main.go` |
| 7 | Add decoder method in `controllertestutils/decoder.go` | **DELETE** — generic decoder (`DecodeInstance[T]`) was added in R1.4, no per-type method needed |

Steps 1-4 and 8-11 remain unchanged.

### Testing Strategy

**Unit tests:** Existing `make test` (envtest) exercises all controller registration. If any controller fails to register or reconcile, tests fail. No new unit tests needed for the generic reconciler itself (it delegates to existing tested code).

**RBAC verification:** `make manifests` must produce identical RBAC output. Diff `config/rbac/role.yaml` before and after migration.

**CRD verification:** `make manifests` must produce identical CRD output. Diff `config/crd/bases/` before and after.

**Integration tests:** All existing integration tests must pass unchanged — the generic reconciler produces identical runtime behavior.

### Previous Story Intelligence

**Story 17.3 (RotationPolicy CRD)** established the latest controller pattern. Key learnings:
- Controller file is a direct copy of PasswordPolicy with type substitutions
- This confirms the extreme duplication the generic reconciler eliminates
- The RotationPolicyConfig inline struct naming pattern is the latest convention

**Story R1.2a (errcheck fixes)** completed 2026-05-15. Key learnings:
- errcheck was run across the entire codebase
- All violations at that time were fixed
- Since then, Epics 11-16 added ~30 new type files + controllers — these need verification

**Story R1.4 (generic decoder)** successfully used Go generics (`DecodeInstance[T]`) to replace 34 per-type methods. This proves the generic pattern works in this codebase and the team is familiar with it.

### Git Intelligence

Recent commits show:
- `11633fd` (latest): Added stories 17.5 and 17.6 to sprint status
- `3ce6784`: Upgraded to Go 1.27, Kubernetes 1.37 — confirms generics are fully supported
- `8e6098e`: Added rotation_schedule support to DatabaseSecretEngineConfig — confirms custom controller complexity

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 17, Story 17.6, lines 3215–3274]
- [Source: internal/controller/policy_controller.go — canonical standard controller pattern]
- [Source: internal/controller/passwordpolicy_controller.go — identical standard pattern]
- [Source: internal/controller/vaultresourcecontroller/vaultresourcereconciler.go — VaultResource type]
- [Source: internal/controller/vaultresourcecontroller/reconcile_skeleton.go — ReconcileWithFunctions]
- [Source: internal/controller/vaultresourcecontroller/utils.go — ReconcilerBase, ManageOutcome]
- [Source: internal/controller/commons.go — prepareContext, VaultAuthenticableResource]
- [Source: api/v1alpha1/utils/vaultobject.go — VaultObject interface]
- [Source: api/v1alpha1/utils/commons.go — ConditionsAware interface]
- [Source: internal/controller/databasesecretengineconfig_controller.go — custom controller example]
- [Source: internal/controller/transitsecretenginekey_controller.go — custom endpoint example]
- [Source: internal/controller/rabbitmqsecretengineconfig_controller.go — custom reconcile example]
- [Source: internal/controller/awssecretengineconfig_controller.go — always-write credential example]
- [Source: _bmad-output/project-context.md — all critical implementation rules]
- [Source: .golangci.yml — linter configuration, errcheck via default:standard]
- [Source: cmd/main.go — controller/webhook registration (~954 lines)]
- [Source: _bmad-output/implementation-artifacts/17-3-rotation-policy-crd.md — latest controller pattern]
- [Source: go.mod — Go 1.27.0 confirmed]
- [Source: docs/contributing-vault-apis.md — public contributor guide, steps 6-7 need generic reconciler update]
- [Source: _bmad-output/project-context.md#Adding a New Vault API Type — AI agent checklist, steps 5-7 need update]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

### Review Model Used

- Quick lens: GPT 5.4 Medium

### Review Findings

Pass 1 — 3 findings: 0 high, 0 medium, 1 low, 2 false
- `[false]` `[reject]` GenericVaultReconciler.Reconcile() returns raw Get() error without ManageOutcome — all 96 pre-migration controllers had identical behavior; ManageOutcome requires a valid instance unavailable on Get failure
- `[false]` `[reject]` SetControllerReference errcheck fix continues on error, may create orphaned resources — original code used explicit discard (`_ =`); story mandates errcheck compliance, not control flow changes
- `[low]` `[patch]` project-context.md still references commons.go as canonical home of prepareContext — fixed: updated 4 sections to reference vaultresourcecontroller.PrepareContext

### Decisions Needed / Decisions Taken

None — all findings resolved during review pass.

### Fixes Applied

- Updated project-context.md sections (Context-Carried Values, Reconcile Flow, Code Organization, Context Value Contract) to reference vaultresourcecontroller.PrepareContext as canonical location
