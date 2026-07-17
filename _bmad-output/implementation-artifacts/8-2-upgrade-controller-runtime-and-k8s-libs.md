# Story 8.2: Upgrade controller-runtime v0.17 → v0.24 and K8s libs v0.29 → v0.36

Status: done

## Story

As an operator developer,
I want to upgrade controller-runtime and K8s client libraries to the latest versions,
So that the operator is compatible with current Kubernetes versions (1.36) and benefits from upstream fixes, security patches, and the new generic webhook API.

## Acceptance Criteria

1. **Given** go.mod pins controller-runtime v0.17.3 and K8s libs v0.29.2, **When** all are updated to v0.24.x / v0.36.x, **Then** `go build ./...` succeeds after adapting to all API changes.

2. **Given** controller-runtime v0.20 removed `webhook.Defaulter` and `webhook.Validator` and v0.23 introduced generic typed webhooks, **When** all 44 webhook files are migrated to `webhook.CustomDefaulter[T]` / `webhook.CustomValidator[T]` with typed method signatures, **Then** all webhooks compile, register correctly, and pass admission requests.

3. **Given** the webhook builder API changed from `ctrl.NewWebhookManagedBy(mgr).For(r).Complete()` to `ctrl.NewWebhookManagedBy(mgr, &Type{}).WithDefaulter(&Type{}).WithValidator(&Type{}).Complete()`, **When** all 44 `SetupWebhookWithManager` methods are updated, **Then** webhook registration succeeds at startup.

4. **Given** envtest behavior may differ between controller-runtime versions, **When** both unit and integration test suites are adapted, **Then** `make test` passes.

5. **Given** several transitive dependencies are removed or restructured across CR v0.18–v0.24, **When** `go mod tidy` resolves the dependency graph, **Then** no import errors remain and the binary builds cleanly.

## Tasks / Subtasks

- [ ] Task 1: Update go.mod dependency versions (AC: #1, #5)
  - [ ] 1.1 Change `sigs.k8s.io/controller-runtime` from `v0.17.3` to `v0.24.1`
  - [ ] 1.2 Change `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go`, `k8s.io/apiextensions-apiserver` from `v0.29.2` to `v0.36.0`
  - [ ] 1.3 Change `k8s.io/component-base` from `v0.29.2` to `v0.36.0` (indirect)
  - [ ] 1.4 Run `go mod tidy` — expect major go.sum changes and removal of deprecated transitive deps (e.g., `imdario/mergo`, `golang/groupcache`, `matttproud/golang_protobuf_extensions/v2`, `google.golang.org/appengine`)
  - [ ] 1.5 Verify `go build ./...` reports only webhook/API migration errors (not dependency resolution failures)

- [ ] Task 2: Migrate all 44 webhook files from `webhook.Defaulter`/`webhook.Validator` to `webhook.CustomDefaulter[T]`/`webhook.CustomValidator[T]` (AC: #2, #3)
  - [ ] 2.1 For each `*_webhook.go` file, apply the migration pattern below — this is the largest single task
  - [ ] 2.2 Update `SetupWebhookWithManager` from `ctrl.NewWebhookManagedBy(mgr).For(r).Complete()` to `ctrl.NewWebhookManagedBy(mgr, &Type{}).WithDefaulter(&Type{}).WithValidator(&Type{}).Complete()`
  - [ ] 2.3 Change compile-time checks from `var _ webhook.Defaulter = &Type{}` to `var _ webhook.CustomDefaulter[Type] = &Type{}`; same for Validator → CustomValidator
  - [ ] 2.4 Update `Default()` → `Default(ctx context.Context, obj runtime.Object) error` (CustomDefaulter) — cast obj to `*Type` or use the generic typed interface
  - [ ] 2.5 Update `ValidateCreate()` → `ValidateCreate(ctx context.Context, obj *Type) (admission.Warnings, error)`
  - [ ] 2.6 Update `ValidateUpdate(old runtime.Object)` → `ValidateUpdate(ctx context.Context, oldObj, newObj *Type) (admission.Warnings, error)` — remove manual type assertions on `old` parameter (e.g., `old.(*DatabaseSecretEngineConfig)` becomes the typed `oldObj` parameter directly)
  - [ ] 2.7 Update `ValidateDelete()` → `ValidateDelete(ctx context.Context, obj *Type) (admission.Warnings, error)`
  - [ ] 2.8 Update the RabbitMQSecretEngineConfig special webhook registration in `main.go` (line 515: `mgr.GetWebhookServer().Register(...)`) — this manually-registered `webhook.Admission{Handler: ...}` handler may need adaptation

- [ ] Task 3: Update main.go webhook registration calls (AC: #3)
  - [ ] 3.1 All `SetupWebhookWithManager(mgr)` calls in `main.go` remain structurally the same (they call the method on the type instance), but verify they compile after the method signature changes in Task 2
  - [ ] 3.2 If the `RabbitMQSecretEngineConfigValidation` custom handler's interface has changed, update accordingly

- [ ] Task 4: Verify ReconcilerBase and controller compatibility (AC: #1)
  - [ ] 4.1 Check `mgr.GetConfig()` in `controllers/vaultresourcecontroller/utils.go` `NewFromManager()` — this method was deprecated in CR v0.19. Verify it still compiles (it may still exist as deprecated but functional). If removed, replace with `mgr.GetHTTPClient()` and obtain rest.Config via the manager's REST mapper or pass it explicitly
  - [ ] 4.2 Check the 10 controllers that use `Watches(&corev1.Secret{TypeMeta: ...}, handler.EnqueueRequestsFromMapFunc(...))` — verify the `Watches()` API still accepts object instances. Controllers: `kubernetesauthenginerole`, `kubernetessecretengineconfig`, `databasesecretengineconfig`, `githubsecretengineconfig`, `ldapauthengineconfig`, `gcpauthengineconfig`, `jwtoidcauthengineconfig`, `azureauthengineconfig`, `azuresecretengineconfig`, `quaysecretengineconfig`
  - [ ] 4.3 Verify `record.EventRecorder` from `mgr.GetEventRecorderFor()` still works — v0.23 migrated to `events.k8s.io` API group

- [ ] Task 5: Fix import path changes and removed packages (AC: #5)
  - [ ] 5.1 Replace `imdario/mergo` imports with `dario.cat/mergo` if used directly (check — likely only indirect)
  - [ ] 5.2 Remove `golang/groupcache` references if present (removed in v0.20)
  - [ ] 5.3 Remove `matttproud/golang_protobuf_extensions/v2` (removed in v0.18)
  - [ ] 5.4 Remove `google.golang.org/appengine` (removed in v0.21)
  - [ ] 5.5 Update `sigs.k8s.io/yaml` — it was downgraded in v0.18 (v1.4.0 → v1.3.0) then upgraded back (v1.4.0 → v1.6.0 in v0.22) — `go mod tidy` should resolve, but verify
  - [ ] 5.6 Verify `gopkg.in/yaml.v2` is no longer needed (removed from CR deps in v0.23) — if our code imports it directly, keep it

- [ ] Task 6: Address envtest and test infrastructure changes (AC: #4)
  - [ ] 6.1 `controllers/suite_test.go` — envtest.Environment setup should remain compatible; verify CRD loading still works
  - [ ] 6.2 `controllers/suite_integration_test.go` — verify `ctrl.NewManager(cfg, ctrl.Options{...})` still compiles (v0.20 removed deprecated `SyncPeriod` from cluster options; our code uses `cache.Options.SyncPeriod` which is fine)
  - [ ] 6.3 Verify envtest stops correctly — v0.24 changed envtest to stop the whole process group
  - [ ] 6.4 Check if any test uses deprecated `admission.Decoder` as a struct (v0.18 made it an interface)

- [ ] Task 7: Compilation verification and fixes (AC: #1)
  - [ ] 7.1 Run `go build ./...` and fix all remaining compilation errors
  - [ ] 7.2 Run `go vet ./...`
  - [ ] 7.3 Run `make test` (envtest-based unit tests)

- [ ] Task 8: Update project-context.md (AC: #1)
  - [ ] 8.1 Update controller-runtime version from v0.17.3 to v0.24.x
  - [ ] 8.2 Update K8s API libs version from v0.29.2 to v0.36.x
  - [ ] 8.3 Update the webhook implementation pattern documentation to reflect CustomDefaulter[T]/CustomValidator[T]
  - [ ] 8.4 Update the `SetupWebhookWithManager` pattern in the "Adding a New Vault API Type" checklist

## Dev Notes

### Scope Boundary

This story covers ONLY the controller-runtime and K8s client library upgrade plus all necessary source code migrations. It does NOT update:
- controller-gen version (Story 8.3: CONTROLLER_TOOLS_VERSION)
- ENVTEST_VERSION or ENVTEST_K8S_VERSION (Story 8.3)
- Kind or kubectl versions (Story 8.3)
- CI workflow pins or Dockerfiles (Story 8.4)
- Go version (Story 8.1 — prerequisite, must be done first)

**Prerequisite:** Story 8.1 (Go 1.26 upgrade) MUST be completed first. CR v0.24 requires Go 1.26 minimum.

### Version Compatibility Matrix

| Component | Current | Target | CR Version |
|-----------|---------|--------|------------|
| controller-runtime | v0.17.3 | v0.24.1 | — |
| k8s.io/api | v0.29.2 | v0.36.0 | — |
| k8s.io/apimachinery | v0.29.2 | v0.36.0 | — |
| k8s.io/client-go | v0.29.2 | v0.36.0 | — |
| k8s.io/apiextensions-apiserver | v0.29.2 | v0.36.0 | — |
| k8s.io/component-base | v0.29.2 | v0.36.0 | — |
| k8s.io/klog/v2 | v2.110.1 | v2.140.0 | — |

### Breaking Changes Per Version (Comprehensive)

#### v0.18 (K8s v1.30)
- `admission.Decoder` became an interface (was a struct) — if any code instantiates a Decoder directly, it will break. This project doesn't create Decoders directly, so LOW RISK.
- Removed `v1alpha1.ControllerManagerConfiguration` — this project does not use it. NO IMPACT.
- Source, Event, Predicate, Handler gained generics — backward compatible for existing non-generic usage.
- `client.SubResourceCreateOptions` signature fixed — verify any subresource client usage.

#### v0.19 (K8s v1.31)
- `admission.Defaulter`/`admission.Validator` deprecated (removal announced for v0.20) — this project uses `webhook.Defaulter`/`webhook.Validator` which are the same interfaces. CRITICAL — must plan migration.
- `client.options.WarningHandler` removed — this project does not use it directly. NO IMPACT.
- Controller name validation added and enforced — all controller names in this project are already valid Go identifiers. LOW RISK.
- Generic `TypedReconciler` added — optional; existing reconcilers still work.

#### v0.20 (K8s v1.32) — HIGHEST IMPACT
- **`webhook.Defaulter` and `webhook.Validator` REMOVED.** All 44 webhook files MUST be migrated to `webhook.CustomDefaulter` and `webhook.CustomValidator`. This is the single largest code change in this story.
- **`webhook.CustomDefaulter` interface:** `Default(ctx context.Context, obj runtime.Object) error` — note the addition of `ctx` and `obj` parameters and the `error` return type.
- **`webhook.CustomValidator` interface:** methods now take `(ctx context.Context, obj runtime.Object) (admission.Warnings, error)` — note the `ctx` parameter. `ValidateUpdate` takes `(ctx, oldObj, newObj runtime.Object)`.
- Deprecated `SyncPeriod` option removed from cluster — this project uses `cache.Options.SyncPeriod` which is the correct replacement. NO IMPACT.
- `webhook.CustomDefaulter` stops deleting unknown fields — behavior change for defaulting webhooks that relied on field stripping. LOW RISK for this project since our defaulters are mostly no-ops.

#### v0.21 (K8s v1.33)
- Client-side rate limiter disabled by default — REST client no longer has a built-in rate limiter. The operator's API calls to the K8s API server may increase throughput. If the cluster enforces API priority and fairness (APF), this is fine. LOW RISK.
- `reconcile.Result.Requeue` deprecated — this project uses `Result{RequeueAfter: SyncPeriod}` via `ManageOutcome`, not bare `Requeue: true`. NO IMPACT.

#### v0.22 (K8s v1.34)
- Native SSA support added — new feature, not breaking for existing code.
- `MatchingLabelsSelector` and `MatchingFieldsSelector` default to `Nothing` if nil — verify no list calls rely on nil selector meaning "match all".
- FakeClient: removed support for objects with pointer ObjectMeta — only affects tests using the fake client with unusual object structures. LOW RISK.

#### v0.23 (K8s v1.35) — SECOND HIGHEST IMPACT
- **Generic Validator and Defaulter:** The `CustomDefaulter`/`CustomValidator` interfaces gained generics. Method signatures change from `runtime.Object` to the concrete type (e.g., `*Policy`). This is the FINAL migration target.
- **Webhook builder API changed:** `builder.WebhookManagedBy(mgr).For(&Type{})` → `builder.WebhookManagedBy(mgr, &Type{})`. The `For()` method is removed from the webhook builder chain. Must use `.WithDefaulter(&Type{}).WithValidator(&Type{}).Complete()`.
- **Events API migration:** `GetEventRecorderFor` now uses `events.k8s.io` API group instead of core. RBAC markers for events must change from `""` (core) apiGroup to `events.k8s.io`. CHECK: does this project use event recording? The controllers use `ManageOutcome` which may record events internally.
- Priority queue enabled by default — no code change needed, but behavioral change: reconcile order may differ.

#### v0.24 (K8s v1.36)
- Removed deprecated custom path function for webhooks — this project does not use custom webhook paths via the deprecated API. NO IMPACT.
- Scheme builder deprecated — this project uses `runtime.NewScheme()` + `AddToScheme`. NO IMPACT (deprecation, not removal).
- envtest: ensures the whole process group is stopped — behavioral improvement, no code change needed.

### Webhook Migration Pattern (Critical)

Every webhook file must be transformed from the OLD pattern to the NEW pattern. There are 44 files to migrate.

**OLD pattern (v0.17):**
```go
func (r *Policy) SetupWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManagedBy(mgr).
        For(r).
        Complete()
}

var _ webhook.Defaulter = &Policy{}

func (r *Policy) Default() {
    policylog.Info("default", "name", r.Name)
}

var _ webhook.Validator = &Policy{}

func (r *Policy) ValidateCreate() (admission.Warnings, error) {
    return nil, nil
}

func (r *Policy) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
    if r.Spec.Path != old.(*Policy).Spec.Path {
        return nil, errors.New("spec.path cannot be updated")
    }
    return nil, nil
}

func (r *Policy) ValidateDelete() (admission.Warnings, error) {
    return nil, nil
}
```

**NEW pattern (v0.24):**
```go
func (r *Policy) SetupWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManagedBy(mgr, r).
        WithDefaulter(r).
        WithValidator(r).
        Complete()
}

var _ webhook.CustomDefaulter[Policy] = &Policy{}

func (r *Policy) Default(ctx context.Context, obj *Policy) error {
    policylog.Info("default", "name", obj.Name)
    return nil
}

var _ webhook.CustomValidator[Policy] = &Policy{}

func (r *Policy) ValidateCreate(ctx context.Context, obj *Policy) (admission.Warnings, error) {
    return nil, nil
}

func (r *Policy) ValidateUpdate(ctx context.Context, oldObj, newObj *Policy) (admission.Warnings, error) {
    if newObj.Spec.Path != oldObj.Spec.Path {
        return nil, errors.New("spec.path cannot be updated")
    }
    return nil, nil
}

func (r *Policy) ValidateDelete(ctx context.Context, obj *Policy) (admission.Warnings, error) {
    return nil, nil
}
```

**Key migration rules for each file:**
1. `SetupWebhookWithManager`: Change `.For(r)` to pass the type as second arg to `NewWebhookManagedBy`, then chain `.WithDefaulter(r).WithValidator(r)`
2. Compile-time checks: `webhook.Defaulter` → `webhook.CustomDefaulter[Type]`, `webhook.Validator` → `webhook.CustomValidator[Type]`
3. `Default()` → `Default(ctx context.Context, obj *Type) error` — the receiver `r` is no longer the object being defaulted; use `obj` instead. Return `nil` for no-op defaults
4. `ValidateCreate()` → `ValidateCreate(ctx context.Context, obj *Type)` — use `obj` instead of `r` for accessing the object
5. `ValidateUpdate(old runtime.Object)` → `ValidateUpdate(ctx context.Context, oldObj, newObj *Type)` — remove ALL manual type assertions like `old.(*Type)` since `oldObj` is already typed. Use `newObj` where the old code used `r`
6. `ValidateDelete()` → `ValidateDelete(ctx context.Context, obj *Type)` — use `obj` instead of `r`
7. Add `"context"` to imports, remove `"k8s.io/apimachinery/pkg/runtime"` import if no longer used

### Files to Modify (Comprehensive)

**Webhook files (44 files in `api/v1alpha1/`):**
All `*_webhook.go` files. Complete list:
- `authenginemount_webhook.go`
- `azureauthengineconfig_webhook.go`, `azureauthenginerole_webhook.go`
- `azuresecretengineconfig_webhook.go`, `azuresecretenginerole_webhook.go`
- `certauthengineconfig_webhook.go`, `certauthenginerole_webhook.go`
- `databasesecretengineconfig_webhook.go`, `databasesecretenginerole_webhook.go`, `databasesecretenginestaticrole_webhook.go`
- `gcpauthengineconfig_webhook.go`, `gcpauthenginerole_webhook.go`
- `githubsecretengineconfig_webhook.go`, `githubsecretenginerole_webhook.go`
- `group_webhook.go`, `groupalias_webhook.go`
- `identityoidcassignment_webhook.go`, `identityoidcclient_webhook.go`, `identityoidcprovider_webhook.go`, `identityoidcscope_webhook.go`
- `identitytokenconfig_webhook.go`, `identitytokenkey_webhook.go`, `identitytokenrole_webhook.go`
- `jwtoidcauthengineconfig_webhook.go`, `jwtoidcauthenginerole_webhook.go`
- `kubernetesauthengineconfig_webhook.go`, `kubernetesauthenginerole_webhook.go`
- `kubernetessecretengineconfig_webhook.go`, `kubernetessecretenginerole_webhook.go`
- `ldapauthengineconfig_webhook.go`, `ldapauthenginegroup_webhook.go`
- `namespace_webhook.go`
- `passwordpolicy_webhook.go`
- `pkisecretengineconfig_webhook.go`, `pkisecretenginerole_webhook.go`
- `policy_webhook.go`
- `quaysecretengineconfig_webhook.go`, `quaysecretenginerole_webhook.go`, `quaysecretenginestaticrole_webhook.go`
- `rabbitmqsecretenginerole_webhook.go`
- `randomsecret_webhook.go`
- `secretenginemount_webhook.go`
- `vaultsecret_webhook.go`

**Webhook test file:**
- `api/v1alpha1/webhook_validate_update_test.go` — verify test assertions still compile after method signature changes

**Main file:**
- `main.go` — line 515 RabbitMQSecretEngineConfig manual webhook registration; rest of webhook registrations call `SetupWebhookWithManager` which changes internally

**Dependency files:**
- `go.mod` — major version bumps
- `go.sum` — fully regenerated

**Project documentation:**
- `_bmad-output/project-context.md` — version references and webhook pattern documentation

**Files NOT modified in this story:**
- `Makefile` — CONTROLLER_TOOLS_VERSION, ENVTEST_VERSION etc. are Story 8.3
- `.github/workflows/*` — CI changes are Story 8.4
- `Dockerfile`, `ci.Dockerfile` — Docker changes are Story 8.4
- Controller files (`controllers/*_controller.go`) — the reconciler interface didn't change (standard `Reconcile(ctx, req) (Result, error)` is still valid)
- `controllers/vaultresourcecontroller/` — ReconcilerBase, ManageOutcome, predicates should compile without changes
- `api/v1alpha1/*_types.go` — type definitions unchanged

### ReconcilerBase `mgr.GetConfig()` Deprecation

`controllers/vaultresourcecontroller/utils.go` contains `NewFromManager()` which calls `mgr.GetConfig()` to obtain the `rest.Config`. This method was deprecated in CR v0.19. As of v0.24, it still exists and compiles but emits a deprecation warning. The replacement depends on usage:
- If the rest.Config is needed for creating additional K8s clients: use `mgr.GetHTTPClient()` for the HTTP transport, or pass the `rest.Config` explicitly from `main.go`'s `ctrl.GetConfigOrDie()`
- Since `NewFromManager` is called from `main.go` where the config is already available, the cleanest fix is to pass `cfg` (from `ctrl.GetConfigOrDie()`) as an additional parameter

**Recommended approach:** Verify `mgr.GetConfig()` still compiles in v0.24. If it does (as deprecated), leave it for now and address in a future cleanup. If it was removed, refactor `NewFromManager` to accept `*rest.Config` as a parameter.

### Controllers with `Watches()` Calls

10 controllers use `Watches()` to watch `corev1.Secret` objects and trigger reconciliation when referenced secrets change. They pass an object instance with `TypeMeta`:

```go
Watches(&corev1.Secret{
    TypeMeta: metav1.TypeMeta{Kind: "Secret"},
}, handler.EnqueueRequestsFromMapFunc(...))
```

The `TypeMeta` field was always unnecessary — it's the type parameter itself that matters. In newer CR versions, verify that `Watches()` still accepts `client.Object` as the first argument. The `handler.EnqueueRequestsFromMapFunc` signature should be stable.

**Controllers affected:** `kubernetesauthenginerole`, `kubernetessecretengineconfig`, `databasesecretengineconfig`, `githubsecretengineconfig`, `ldapauthengineconfig`, `gcpauthengineconfig`, `jwtoidcauthengineconfig`, `azureauthengineconfig`, `azuresecretengineconfig`, `quaysecretengineconfig`

### Webhooks with Complex Validation Logic

Most webhooks are simple (no-op defaulters, path-immutability validators). The following have non-trivial logic that needs careful migration:

**Webhooks with `old.(*Type)` type assertions (must use `oldObj` parameter directly):**
- `authenginemount_webhook.go` — checks `spec.path` immutability
- `azureauthengineconfig_webhook.go` — checks `spec.path` immutability + credential validation
- `azuresecretengineconfig_webhook.go` — checks `spec.path` immutability + credential validation
- `certauthengineconfig_webhook.go` — checks `spec.path` immutability
- `databasesecretengineconfig_webhook.go` — checks `spec.path` immutability
- `databasesecretenginerole_webhook.go` — checks `spec.path` immutability
- `githubsecretengineconfig_webhook.go` — checks `spec.path` immutability
- `kubernetesauthengineconfig_webhook.go` — checks `spec.path` immutability + credential validation
- `kubernetessecretengineconfig_webhook.go` — checks `spec.path` immutability + credential validation
- `ldapauthengineconfig_webhook.go` — checks `spec.path` immutability + credential validation
- `quaysecretengineconfig_webhook.go` — checks `spec.path` immutability + credential validation
- `secretenginemount_webhook.go` — checks `spec.path` immutability

**Webhooks with credential validation (`credentials.ValidateCredentialSource`):**
These call `r.isValid()` or `credentials.ValidateCredentialSource()` — the validation logic itself doesn't change, but the caller must use `obj`/`newObj` instead of `r`.

**Special webhook (not using SetupWebhookWithManager):**
- `RabbitMQSecretEngineConfigValidation` in `main.go` line 515 — manually registered via `mgr.GetWebhookServer().Register(...)`. This uses `webhook.Admission{Handler: ...}` which may still work, but verify the `admission.Handler` interface hasn't changed. The `admission.Decoder` is now an interface (since v0.18).

### Recommended Migration Strategy

Due to the large number of files (44 webhooks), use a systematic approach:

1. **Start with go.mod**: Update versions, run `go mod tidy`, get the dependency graph correct
2. **Fix compilation errors iteratively**: Start with a single webhook (e.g., `policy_webhook.go` as the simplest), get it compiling, then apply the same pattern to all others
3. **Use search-and-replace patterns** for the repetitive parts:
   - `webhook.Defaulter` → `webhook.CustomDefaulter[Type]`
   - `webhook.Validator` → `webhook.CustomValidator[Type]`
   - `For(r).\n\t\tComplete()` → `, r).\n\t\tWithDefaulter(r).\n\t\tWithValidator(r).\n\t\tComplete()`
4. **Manually handle**: ValidateUpdate methods that have `old.(*Type)` assertions — each needs the assertion removed and parameter renamed
5. **Test after each batch**: Run `go build ./...` after every few files

### Events API Change (v0.23)

controller-runtime v0.23 migrated to the new `events.k8s.io` API group for event recording. The `GetEventRecorderFor` function now uses this API group. If the operator's controllers or `ManageOutcome` record events via the manager's event recorder, the RBAC markers for events may need updating:

**Check needed:** Search for `Eventf`, `Event`, `GetEventRecorderFor`, `record.EventRecorder` in the codebase. If events are used:
- RBAC markers must change from `//+kubebuilder:rbac:groups="",resources=events,...` to `//+kubebuilder:rbac:groups=events.k8s.io,resources=events,...`
- Note: CRD-level RBAC markers don't change — only event-related RBAC

**However:** controller-gen version update is Story 8.3. The RBAC regeneration from markers happens in that story. In this story, only update the marker comments if events RBAC exists.

### Previous Story Intelligence

Story 8.1 (Go 1.22 → 1.26 upgrade) established:
- Go 1.26 is a drop-in version bump with no source code changes required
- `go.mod` uses `go 1.26` with `toolchain go1.26.4`
- Dockerfile uses multi-arch pattern with `$BUILDPLATFORM`, `$TARGETOS`, `$TARGETARCH`
- CI workflows use `GO_VERSION: ~1.26`
- The `go.sum` changes from a Go version bump are large but expected

For this story, the `go.sum` changes will be even larger since we're upgrading the entire K8s dependency stack.

### Testing Standards

- Run `go build ./...` to verify compilation
- Run `go vet ./...` for static analysis
- Run `make test` (envtest-based unit tests with build tag `!integration`)
- Integration tests (`make integration`) require a Kind cluster — verify locally if feasible but depend on Story 8.3 for ENVTEST_K8S_VERSION update. Unit tests should work with current envtest binaries since envtest is backward-compatible
- The webhook validation test file (`api/v1alpha1/webhook_validate_update_test.go`) must be updated if it calls the old-signature methods

### Project Structure Notes

- No new files created; only existing files modified
- All 44 webhook files follow the same structural pattern, enabling batch migration
- The webhook pattern change is the largest single refactor in the project's history
- After this story, all webhooks will use the modern type-safe generic pattern

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 8, Story 8.2]
- [Source: _bmad-output/project-context.md — Technology Stack & Versions]
- [Source: go.mod — current dependency versions]
- [Source: api/v1alpha1/policy_webhook.go — representative webhook pattern]
- [Source: api/v1alpha1/databasesecretengineconfig_webhook.go — webhook with ValidateUpdate type assertion]
- [Source: main.go — line 515 RabbitMQSecretEngineConfig manual webhook registration]
- [Source: controllers/suite_test.go — envtest unit test setup]
- [Source: controllers/suite_integration_test.go — integration test setup with ctrl.NewManager]
- [Source: controller-runtime v0.18.0 release notes — https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.18.0]
- [Source: controller-runtime v0.19.0 release notes — https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.19.0]
- [Source: controller-runtime v0.20.0 release notes — https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.20.0]
- [Source: controller-runtime v0.21.0 release notes — https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.21.0]
- [Source: controller-runtime v0.22.0 release notes — https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.22.0]
- [Source: controller-runtime v0.23.0 release notes — https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.23.0]
- [Source: controller-runtime v0.24.0 release notes — https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.24.0]
- [Source: controller-runtime compatibility matrix — CR v0.24 requires k8s.io/* v0.36 and Go 1.26]

### Review Findings

- [x] [Review][Patch] Generated CRDs contain unresolved `LocalObjectReference` `$ref`s — reverted CRDs to HEAD; regeneration deferred to Story 8.3 (controller-gen upgrade)
- [x] [Review][Patch] Events RBAC updated from core API group to `events.k8s.io` in all 33 controller markers and `config/rbac/role.yaml`
- [x] [Review][Defer] Broken validating webhook markers are pre-existing in `Policy` and `PasswordPolicy` [api/v1alpha1/policy_webhook.go:50, api/v1alpha1/passwordpolicy_webhook.go:50] - deferred, pre-existing
- [x] [Review][Defer] Copy-pasted `authenginemountlog` usage remains in several migrated defaulters [api/v1alpha1/databasesecretenginerole_webhook.go:44, api/v1alpha1/kubernetesauthenginerole_webhook.go:44, api/v1alpha1/randomsecret_webhook.go:44] - deferred, pre-existing

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
