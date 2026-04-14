---
project_name: 'vault-config-operator'
user_name: 'Raffa'
date: '2026-04-11'
sections_completed: ['technology_stack', 'language_rules', 'framework_rules', 'testing_rules', 'code_quality', 'workflow_rules', 'critical_rules']
status: 'complete'
rule_count: 47
optimized_for_llm: true
---

# Project Context for AI Agents

_This file contains critical rules and patterns that AI agents must follow when implementing code in this project. Focus on unobvious details that agents might otherwise miss._

---

## Technology Stack & Versions

### Core
- **Language:** Go 1.22.0
- **K8s Framework:** controller-runtime v0.17.3, Kubebuilder v3 layout
- **OLM/SDK:** Operator SDK v1.31.0
- **K8s API libs:** k8s.io/api, apimachinery, client-go v0.29.2
- **Vault Client:** hashicorp/vault/api v1.14.0

### Key Dependencies
- Masterminds/sprig/v3 v3.2.3 (template functions for VaultSecret)
- hashicorp/hcl/v2 v2.21.0, BurntSushi/toml v1.4.0 (config parsing)
- go-logr/logr v1.4.2 (structured logging via controller-runtime/zap)
- onsi/ginkgo/v2 v2.19.0 + onsi/gomega v1.33.1 (BDD testing)
- scylladb/go-set v1.0.2 (set data structures)

### Build & Dev Tooling
- controller-gen v0.14.0 (CRD/RBAC generation)
- kustomize v5.4.3, Helm v3.11.0
- golangci-lint v1.59.1 (no committed config — uses defaults or shared workflow config)
- Kind v0.27.0, kubectl v1.29.0, Vault 1.19.0 (integration testing)
- Container: golang:1.22 builder → registry.access.redhat.com/ubi9/ubi-minimal runtime
- CI: GitHub Actions via reusable workflows from redhat-cop/github-workflows-operators

## Critical Implementation Rules

### Go Language Rules

#### Admission Webhooks (Required for Every Type)
- Every CRD type **must** have a corresponding `*_webhook.go` file implementing both `webhook.Defaulter` and `webhook.Validator` interfaces.
- Compile-time checks: `var _ webhook.Defaulter = &MyType{}` and `var _ webhook.Validator = &MyType{}`
- Webhook file must declare a package-level logger: `var mytypelog = logf.Log.WithName("mytype-resource")`
- `SetupWebhookWithManager` follows a fixed pattern: `ctrl.NewWebhookManagedBy(mgr).For(r).Complete()`
- Kubebuilder marker comments are required for both mutating and validating paths:
  - `//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-<lowercase>,mutating=true,...`
  - `//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-<lowercase>,mutating=false,...`
- **Critical rule:** `ValidateUpdate` must **always** reject changes to `spec.path` — immutable after creation: `errors.New("spec.path cannot be updated")`
- Types with mount semantics (engines) may additionally restrict updates to only the config sub-struct.
- Webhook must be registered in `main.go` inside the `ENABLE_WEBHOOKS` env-var guard block.

#### ConditionsAware Implementation
- Status struct must contain: `Conditions []metav1.Condition` with exact struct tags: `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
- Must include these kubebuilder markers on the Conditions field:
  ```
  // +patchMergeKey=type
  // +patchStrategy=merge
  // +listType=map
  // +listMapKey=type
  ```
- Implement `GetConditions()` and `SetConditions()` — simple getter/setter on `m.Status.Conditions`.
- The framework (`ManageOutcome`) handles setting `ReconcileSuccessful`/`ReconcileFailed` conditions automatically; controllers must **not** set conditions directly.

#### Imperative-to-Declarative Bridge: `toMap()` and `IsEquivalentToDesiredState()`
- `toMap()` is defined on the inline Vault-specific config struct (e.g., `DBSEConfig`, `AuthMount`) and converts CRD spec fields to a `map[string]interface{}` matching the Vault API's JSON field names (snake_case).
- `GetPayload()` calls `d.Spec.toMap()` to produce the map sent to Vault on write.
- `IsEquivalentToDesiredState(payload)` compares the current Vault state (read response) against the desired state (from `toMap()`) using `reflect.DeepEqual`. This is the core declarative reconciliation check — if equivalent, no write is issued.
- **Critical nuance:** Vault often returns extra fields not sent by the operator, or restructures fields (e.g., DB config moves `connection_url`/`username` into a `connection_details` sub-map). `IsEquivalentToDesiredState` must account for these transformations — filter payload to only managed keys, remap fields as Vault returns them.
- For engine mounts (`VaultEngineObject`), `IsEquivalentToDesiredState` compares the **tune config** (`Config.toMap()`) not the full mount spec, because Vault's read response for mounts returns only tune-level fields.

#### Context-Carried Values
- `prepareContext()` in `controllers/commons.go` enriches the context with 4 values: `"kubeClient"`, `"restConfig"`, `"vaultConnection"`, `"vaultClient"`.
- All downstream code retrieves these via `context.Value("key").(Type)` — type assertions with no safety check (will panic if missing).
- The `vaultClient` is obtained per-reconcile via `GetKubeAuthConfiguration().GetVaultClient(ctx, namespace)`.

#### Error Management Pattern
- Controllers: `apierrors.IsNotFound(err)` → return `reconcile.Result{}, nil` (don't requeue on deleted objects).
- All other errors from Get/context-preparation → `ManageOutcome(ctx, r.ReconcilerBase, instance, err)` which sets a `ReconcileFailed` condition on the CR status and returns the error for requeue.
- In `_types.go` methods: log the error with `log.Error(err, "descriptive message", "key", value)` then return the error up the chain. Never swallow errors silently.
- Vault 404s: type-assert to `*vault.ResponseError`, check `StatusCode == 404`, return nil (treat as non-fatal).

#### Logging Conventions
- **Controllers:** `log := log.FromContext(ctx)` at top of Reconcile. Use `log.V(1).Info(...)` for debug/verbose messages.
- **Controllers also use** `r.Log.Error(...)` (the `ReconcilerBase.Log` field) for errors in `prepareContext` failures and map-func handlers.
- **Webhooks:** Use the package-level `var mytypelog = logf.Log.WithName("mytype-resource")` logger. Every handler logs its operation: `mytypelog.Info("validate update", "name", r.Name)`.
- **Types (`_types.go`):** Use `log.FromContext(context)` when a context is available (e.g., in `PrepareInternalValues`).
- Structured key-value pairs: always `"key", value` format, never `fmt.Sprintf`.

### Framework-Specific Rules (Kubernetes Operator)

#### Controller Structure
- Every controller embeds `vaultresourcecontroller.ReconcilerBase` (not a pointer).
- Controller registration in `main.go` always uses: `&controllers.MyTypeReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "MyType")}`.
- `SetupWithManager` uses `builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())` on the `For()` call to enable generation-based + optional drift detection filtering.

#### Reconcile Flow (Standard Types)
1. Fetch instance via `r.GetClient().Get(ctx, req.NamespacedName, instance)`.
2. If `apierrors.IsNotFound` → return `reconcile.Result{}, nil`.
3. Call `prepareContext(ctx, r.ReconcilerBase, instance)` to build enriched context.
4. Create `vaultresourcecontroller.NewVaultResource(&r.ReconcilerBase, instance)` (or `NewVaultEngineResource` for mount types).
5. Delegate to `vaultResource.Reconcile(ctx1, instance)`.

#### Three Reconciler Variants
- **`VaultResource`** — standard types (policies, roles, configs). Uses `CreateOrUpdate` which reads from Vault, calls `IsEquivalentToDesiredState`, writes only if different.
- **`VaultEngineResource`** — mount types (`AuthEngineMount`, `SecretEngineMount`). Creates/enables the engine, then tunes it separately via `GetEngineTunePath()`/`GetTunePayload()`. Implements `VaultEngineObject` interface.
- **`VaultAuditResource`** — audit types. Separate reconciler for audit device lifecycle.

#### Finalizer Management
- Finalizers are managed automatically by `ManageOutcome` — added on first successful reconcile if `IsDeletable()` returns true.
- Finalizer name is derived from the object via `vaultutils.GetFinalizer(obj)`.
- On deletion: the framework checks for `ReconcileSuccessful` condition before deleting from Vault (only clean up if it was ever successfully created).

#### Kubebuilder RBAC Markers
- Every controller must declare RBAC markers for its resource (get/list/watch/create/update/patch/delete), status subresource, and finalizers.
- Common shared RBAC: `secrets` (get/list/watch), `serviceaccounts/token` (get/list/watch or create), `events` (get/list/watch/create/patch).

#### CRD Type File Structure (`*_types.go`)
- Standard Spec fields in order: `Connection`, `Authentication`, inline config struct, `Path` (if applicable), `Name` (optional override for Vault object name).
- The `Name` field allows `spec.name` to override `metadata.name` for the Vault object, with pattern validation: `[a-z0-9]([-a-z0-9]*[a-z0-9])?`.
- `GetPath()` must check `d.Spec.Name != ""` first, falling back to `d.Name` (metadata name).
- `init()` must call `SchemeBuilder.Register(&MyType{}, &MyTypeList{})`.

### Testing Rules

#### Two Test Tiers (Build Tag Separation)
- **Unit tests** (`suite_test.go`): build tag `//go:build !integration` — runs with `go test ./...` (default `make test`). Uses controller-runtime `envtest` (local etcd + API server), CRDs loaded from `config/crd/bases`. No real Vault.
- **Integration tests** (`suite_integration_test.go` + `*_controller_test.go`): build tag `//go:build integration` — requires a live Kind cluster with Vault. Sets `USE_EXISTING_CLUSTER=true`, expects `VAULT_ADDR` and `VAULT_TOKEN` env vars. Registers all controllers against a real manager, creates test namespaces (`vault-admin`, `test-vault-config-operator`).

#### Integration Test Infrastructure Philosophy
Vault typically acts as an intermediary between vault-config-operator and a secured service (e.g., PostgreSQL, RabbitMQ) or auth provider (e.g., LDAP, Kubernetes). The following rule determines how each external dependency is handled in integration tests:

1. **Install in Kind** — If the service can be installed in the Kind cluster and configured to work with Vault, the test **must** deploy it as a real service (e.g., PostgreSQL via Helm, RabbitMQ via Helm, OpenLDAP via manifests).
2. **Mock it** — If the service cannot be installed but can be simply mocked (e.g., a static JWKS endpoint for JWT/OIDC auth), the test **should** use a lightweight mock.
3. **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test **must not** be run. These types rely on unit test coverage only.

This rule applies to all current and future integration tests. When adding a new type, classify its external dependency into one of these three categories and document the decision in the story file.

#### Integration Test Pattern
- Uses Ginkgo v2 `Describe`/`Context`/`It` BDD blocks with dot-imported `gomega` matchers.
- Test fixtures are YAML files in `test/` directory, loaded via `controllertestutils.decoder.Get<TypeName>Instance("../test/<path>.yaml")`.
- The decoder (`controllertestutils/decoder.go`) deserializes YAML files into typed CRD objects. Each type needs a `Get<TypeName>Instance` method added to the decoder.
- Reconcile success is verified by polling the CR status for `ReconcileSuccessful` condition with `Eventually(func() bool {...}, timeout, interval).Should(BeTrue())`.
- Standard timeout: `120s`, poll interval: `2s`.
- Tests create the resource, wait for successful reconcile, then delete and wait for deletion.

#### Unit Tests (non-Ginkgo)
- Pure Go `testing.Test*` functions used for utility packages (e.g., `vaultsecretutils/hash_test.go`).
- Table-driven tests with explicit expected values.

#### Adding Tests for New Types
- Add a `Get<NewType>Instance` method to `controllers/controllertestutils/decoder.go`.
- Create YAML fixtures in `test/<feature>/` directory.
- Write integration test in `controllers/<newtype>_controller_test.go` with `//go:build integration` tag.

### Code Quality & Style Rules

#### File Naming Conventions
- CRD types: `api/v1alpha1/<lowercasetype>_types.go` (e.g., `authenginemount_types.go`)
- Webhooks: `api/v1alpha1/<lowercasetype>_webhook.go`
- Controllers: `controllers/<lowercasetype>_controller.go`
- Controller tests: `controllers/<lowercasetype>_controller_test.go`
- Generated deepcopy: `api/v1alpha1/zz_generated.deepcopy.go` (never edit manually)

#### Code Organization
- All CRD types, interface implementations, and `toMap()` live together in `*_types.go` — not split across files.
- Shared Vault utilities: `api/v1alpha1/utils/` — `VaultObject`, `VaultEndpoint`, `VaultConnection`, `KubeAuthConfiguration`, helper functions.
- Shared controller logic: `controllers/vaultresourcecontroller/` — `ReconcilerBase`, `ManageOutcome`, predicates, reconciler variants.
- Common controller helpers: `controllers/commons.go` — `prepareContext()`, `VaultAuthenticableResource` interface.

#### Kubebuilder Markers
- All exported Spec fields must have `// +kubebuilder:validation:Required` or `// +kubebuilder:validation:Optional`.
- List fields use `// +listType=set` for unique items, `// +listType=map` with `// +listMapKey=` for keyed lists.
- Map fields use `// +mapType=granular`.
- The root type must have `//+kubebuilder:object:root=true` and `//+kubebuilder:subresource:status`.
- RBAC markers on reconciler functions — one set per controller file.

#### JSON Tag Conventions
- CRD fields use camelCase json tags: `json:"fieldName,omitempty"`.
- `toMap()` converts to snake_case keys matching Vault API expectations.
- Unexported internal fields use `json:"-"` to exclude from serialization.
- `Path` is a custom type `vaultutils.Path` (string alias), serialized as `json:"path,omitempty"`.

#### Code Quality Gates
- CI runs `go fmt`, `go vet`, and `make test` (envtest-based unit tests). No golangci-lint in CI.
- golangci-lint v1.59.1 is available locally via `make golangci-lint` but has no committed config (`.golangci.yml`) and is not wired into `make test` or the CI workflow.
- Effective quality enforcement: `go fmt` formatting + `go vet` static analysis only.

### Development Workflow Rules

#### CI Pipeline
- GitHub Actions with reusable workflows from `redhat-cop/github-workflows-operators` (pinned by commit SHA).
- PR workflow (`pr.yaml`): runs unit tests, integration tests, and helmchart tests.
- Push workflow (`push.yaml`): same tests + image push + OLM community operator PR (on tags).
- Tag-based releases: pushing a `v*` tag triggers full release pipeline.

#### Build Targets (Makefile)
- `make manifests` — regenerate CRDs and RBAC via controller-gen. **Run after any type or RBAC marker change.**
- `make generate` — regenerate `zz_generated.deepcopy.go`. **Run after any `*_types.go` change.**
- `make fmt && make vet` — format and static analysis.
- `make test` — unit tests with envtest (excludes `//go:build integration`).
- `make integration` — full integration suite: Kind cluster, Vault, cert-manager, ingress, then runs integration-tagged tests.
- `make docker-build` — builds container image (runs `test` first).
- `make bundle` — regenerates OLM bundle manifests.
- `make helmchart` — generates Helm chart from kustomize overlays.

#### Local Development
- **Tiltfile** for live-reload development: uses `podman` build with `ci.Dockerfile`, deploys via kustomize `config/local-development/tilt`.
- `ENABLE_WEBHOOKS=false` env var disables webhook registration for local testing without cert-manager.
- `SYNC_PERIOD_SECONDS` env var controls cache sync interval (default: 36000s / 10hrs).
- `ENABLE_DRIFT_DETECTION=true` env var enables periodic reconciliation for drift detection.

#### Adding a New Vault API Type (Full Checklist)
1. `operator-sdk create api --group redhatcop --version v1alpha1 --kind MyType --resource --controller`
2. `operator-sdk create webhook --group redhatcop --version v1alpha1 --kind MyType --defaulting --programmatic-validation`
3. Define `*_types.go`: Spec with `Connection`, `Authentication`, inline config struct, `Path`, `Name`; implement `VaultObject` + `ConditionsAware`; add `toMap()` and `IsEquivalentToDesiredState()`
4. Implement webhook in `*_webhook.go`: `Defaulter`, `Validator`, immutable `spec.path` rule
5. Implement controller in `*_controller.go`: embed `ReconcilerBase`, standard reconcile flow
6. Register controller + webhook in `main.go`
7. Add decoder method in `controllertestutils/decoder.go`
8. Create test YAML fixtures in `test/`
9. Write integration test with `//go:build integration` tag
10. Run `make manifests generate fmt vet test`

### Critical Don't-Miss Rules

#### Anti-Patterns to Avoid
- **Never write directly to Vault without the `IsEquivalentToDesiredState` check.** The `CreateOrUpdate` flow reads first and only writes if state diverges. Writing unconditionally causes unnecessary Vault audit log noise and potential rate limiting.
- **Never set conditions on the CR status directly in controllers.** Always delegate to `ManageOutcome()` / `ManageOutcomeWithRequeue()` which handles both success and failure conditions consistently.
- **Never add finalizers manually in controllers.** `ManageOutcome` adds the finalizer automatically after first successful reconcile when `IsDeletable()` returns true.
- **Never create a new logger in controllers or types.** Use `log.FromContext(ctx)` or `r.Log` — never `logr.New()` or `ctrl.Log`.

#### Vault API Gotchas
- Vault's read response often restructures fields compared to the write payload (e.g., `connection_url` becomes nested in `connection_details`). `IsEquivalentToDesiredState` must transform the desired state to match Vault's read format before comparison.
- Vault returns extra fields not managed by the operator. Filter the read payload to only the keys the operator manages before calling `reflect.DeepEqual`.
- Vault returns `[]interface{}` for lists, not `[]string`. The `toInterfaceArray()` helper converts `[]string` to `[]interface{}` for accurate comparison.
- Engine mounts (auth/secret) compare only the **tune config**, not the full mount spec, because Vault's tune endpoint returns a different payload structure than the enable endpoint.

#### Code Generation Pitfalls
- After modifying `*_types.go`, **always** run `make manifests generate`. Forgetting this leaves CRDs, RBAC, and deepcopy out of sync.
- The `PROJECT` file lists Kubebuilder-managed resources but may drift from `main.go` registrations (e.g., Entity/EntityAlias exist in code but may not be in `PROJECT`). `main.go` is the source of truth for what's actually deployed.
- `zz_generated.deepcopy.go` is auto-generated — never edit it manually.

#### Context Value Contract
- The enriched context from `prepareContext()` carries `"kubeClient"`, `"restConfig"`, `"vaultConnection"`, `"vaultClient"` — all retrieved via unsafe type assertions. If any of these are missing, the operator will panic. Any new context value must follow the same pattern.

#### Deletion Safety
- `manageCleanUpLogic` only deletes from Vault if the CR previously had a `ReconcileSuccessful=True` condition. This prevents deleting Vault resources that were never successfully created.
- `IsDeletable()` controls whether the resource gets a finalizer and whether Vault cleanup happens. Types returning `false` will not be cleaned up from Vault on CR deletion.

---

## Usage Guidelines

**For AI Agents:**
- Read this file before implementing any code
- Follow ALL rules exactly as documented
- When in doubt, prefer the more restrictive option
- Update this file if new patterns emerge

**For Humans:**
- Keep this file lean and focused on agent needs
- Update when technology stack changes
- Review quarterly for outdated rules
- Remove rules that become obvious over time

Last Updated: 2026-04-14
