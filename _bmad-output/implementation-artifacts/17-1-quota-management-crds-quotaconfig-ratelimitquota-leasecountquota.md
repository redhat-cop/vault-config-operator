# Story 17.1: Quota Management CRDs (QuotaConfig + RateLimitQuota + LeaseCountQuota)

Status: done

## Story

As an operator developer,
I want CRDs for QuotaConfig, RateLimitQuota, and LeaseCountQuota,
So that Vault's resource quota system can be managed declaratively.

**Closes:** [GitHub Issue #249 — Feature Request: Rate Limiting with Resource Quotas](https://github.com/redhat-cop/vault-config-operator/issues/249)

## Acceptance Criteria

1. **Given** a QuotaConfig CR is created with `enableRateLimitAuditLogging` and `rateLimitExemptPaths` **When** the reconciler processes it **Then** the global quota config is written to Vault at `sys/quotas/config` and ReconcileSuccessful=True
2. **Given** a RateLimitQuota CR is created with `rate`, `path`, `interval`, and `blockInterval` **When** the reconciler processes it **Then** a named rate limit quota exists in Vault at `sys/quotas/rate-limit/{name}` and ReconcileSuccessful=True
3. **Given** a LeaseCountQuota CR is created with `maxLeases` and `path` **When** the reconciler processes it **Then** a named lease count quota exists in Vault at `sys/quotas/lease-count/{name}` and ReconcileSuccessful=True
4. **Given** quota CRs are updated (e.g., rate changed from 100 to 200) **When** the reconciler processes the update **Then** the Vault quota configuration is updated to match
5. **Given** RateLimitQuota or LeaseCountQuota CRs are deleted **When** the reconciler processes deletion **Then** the named quotas are removed from Vault
6. **Given** the QuotaConfig CR is deleted **When** the reconciler processes deletion **Then** the global config is NOT deleted from Vault (`IsDeletable()` returns false — always-present config)
7. **Given** the CRD types are implemented **When** the story is marked done **Then** documentation exists at `docs/system-backend/quotas.md`

## Tasks / Subtasks

- [ ] Task 1 — Create `QuotaConfig` CRD type (AC: #1, #6)
  - [ ] 1.1 Create `api/v1alpha1/quotaconfig_types.go` — Spec, Status, VaultObject interface, toMap(), IsEquivalentToDesiredState()
  - [ ] 1.2 Create `api/v1alpha1/quotaconfig_webhook.go` — Defaulter, Validator (immutability rules)
  - [ ] 1.3 Create `internal/controller/quotaconfig_controller.go` — standard VaultResource reconciler
  - [ ] 1.4 Register controller in `cmd/main.go` (controller + webhook)
- [ ] Task 2 — Create `RateLimitQuota` CRD type (AC: #2, #4, #5)
  - [ ] 2.1 Create `api/v1alpha1/ratelimitquota_types.go` — Spec, Status, VaultObject interface, toMap(), IsEquivalentToDesiredState()
  - [ ] 2.2 Create `api/v1alpha1/ratelimitquota_webhook.go` — Defaulter, Validator (immutability + rate validation)
  - [ ] 2.3 Create `internal/controller/ratelimitquota_controller.go` — standard VaultResource reconciler
  - [ ] 2.4 Register controller in `cmd/main.go` (controller + webhook)
- [ ] Task 3 — Create `LeaseCountQuota` CRD type (AC: #3, #4, #5)
  - [ ] 3.1 Create `api/v1alpha1/leasecountquota_types.go` — Spec, Status, VaultObject interface, toMap(), IsEquivalentToDesiredState()
  - [ ] 3.2 Create `api/v1alpha1/leasecountquota_webhook.go` — Defaulter, Validator (immutability + maxLeases validation)
  - [ ] 3.3 Create `internal/controller/leasecountquota_controller.go` — standard VaultResource reconciler
  - [ ] 3.4 Register controller in `cmd/main.go` (controller + webhook)
- [ ] Task 4 — Unit tests (AC: #1, #2, #3, #4)
  - [ ] 4.1 Create `api/v1alpha1/quotaconfig_test.go` — toMap() output, IsEquivalentToDesiredState (positive + negative + extra-fields)
  - [ ] 4.2 Create `api/v1alpha1/ratelimitquota_test.go` — toMap() output, IsEquivalentToDesiredState (positive + negative + extra-fields), duration conversion
  - [ ] 4.3 Create `api/v1alpha1/leasecountquota_test.go` — toMap() output, IsEquivalentToDesiredState (positive + negative + extra-fields)
- [ ] Task 5 — Integration tests (AC: #1, #2, #4, #5, #6)
  - [ ] 5.1 Create test fixtures: `test/quotaconfig/quota-config.yaml`, `test/quotaconfig/rate-limit-quota.yaml`
  - [ ] 5.2 Create `internal/controller/quotaconfig_controller_test.go` — create, reconcile, update, delete-no-op (IsDeletable=false persistence check)
  - [ ] 5.3 Create `internal/controller/ratelimitquota_controller_test.go` — create, reconcile, update, delete
  - [ ] 5.4 **Skip** LeaseCountQuota integration test — Enterprise-only API, not testable in OSS Kind cluster
- [ ] Task 6 — Documentation (AC: #7)
  - [ ] 6.1 Create `docs/system-backend/` directory and `docs/system-backend/quotas.md`
  - [ ] 6.2 Add link from main readme or docs index
- [ ] Task 7 — Code generation and validation
  - [ ] 7.1 `make manifests generate fmt vet test` passes
  - [ ] 7.2 `make lint` passes
  - [ ] 7.3 Add new CRD YAMLs to `config/crd/kustomization.yaml`
  - [ ] 7.4 Verify no existing tests regress

## Dev Notes

### Overview

This story introduces 3 new CRD types for Vault's quota management system. All three use the standard `VaultObject` + `VaultEndpoint` + `VaultResource` generic reconciler — no custom reconciler logic required. No credentials, no external dependencies.

`QuotaConfig` is the **first singleton CRD** in this project — it manages a single fixed Vault endpoint with no `{name}` segment. This is the only novel element.

### Vault API Reference

| CRD | Vault Path | Method | Vault Docs |
|-----|-----------|--------|------------|
| QuotaConfig | `sys/quotas/config` | POST/GET (no DELETE) | [quotas-config](https://developer.hashicorp.com/vault/api-docs/system/quotas-config) |
| RateLimitQuota | `sys/quotas/rate-limit/{name}` | POST/GET/DELETE/LIST | [rate-limit-quotas](https://developer.hashicorp.com/vault/api-docs/system/rate-limit-quotas) |
| LeaseCountQuota | `sys/quotas/lease-count/{name}` | POST/GET/DELETE/LIST | [lease-count-quotas](https://developer.hashicorp.com/vault/api-docs/system/lease-count-quotas) |

**⚠️ LeaseCountQuota is Enterprise-only.** The Vault API docs state "Appropriate Vault Enterprise license required" for `/sys/quotas/lease-count`. The CRD is still created for Enterprise users — OSS Vault will return an API error, surfaced as a `ReconcileFailed` condition on the CR. Document this clearly.

### CRD 1: QuotaConfig (Singleton)

**Pattern:** Follows `PasswordPolicy` structure, but simpler — no `Path`, no `Name`, fixed endpoint.

**Vault write params (POST `sys/quotas/config`):**
- `enable_rate_limit_audit_logging` (bool, default false)
- `enable_rate_limit_response_headers` (bool, default false)
- `rate_limit_exempt_paths` ([]string, default [])
- `absolute_rate_limit_exempt_paths` ([]string, default [])

**Vault read response (GET `sys/quotas/config`):**
```json
{
  "enable_rate_limit_audit_logging": false,
  "enable_rate_limit_response_headers": false,
  "rate_limit_exempt_paths": ["sys/internal/ui/mounts", "sys/health", ...]
}
```

**⚠️ `absolute_rate_limit_exempt_paths` is NOT shown in the Vault GET sample response.** It may not be returned on read. The dev must verify against a live Vault instance:
1. Write a config with `absolute_rate_limit_exempt_paths` set
2. Read it back and check if the field is present
3. If NOT returned: exclude from `IsEquivalentToDesiredState` comparison (emit in `toMap()` for writes but delete from `desiredState` before compare)
4. If returned: include normally

**Type structure:**

```go
// QuotaConfigSpec defines the desired state of QuotaConfig
type QuotaConfigSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication"`

    QuotaConfigData `json:",inline"`
}

type QuotaConfigData struct {
    // EnableRateLimitAuditLogging enables audit logging of requests rejected due to rate limit quota violations.
    // +kubebuilder:validation:Optional
    EnableRateLimitAuditLogging bool `json:"enableRateLimitAuditLogging,omitempty"`

    // EnableRateLimitResponseHeaders adds rate limit quota HTTP headers to responses (Retry-After, X-Ratelimit-*).
    // +kubebuilder:validation:Optional
    EnableRateLimitResponseHeaders bool `json:"enableRateLimitResponseHeaders,omitempty"`

    // RateLimitExemptPaths lists relative paths exempt from all rate limit quotas.
    // +kubebuilder:validation:Optional
    // +listType=set
    RateLimitExemptPaths []string `json:"rateLimitExemptPaths,omitempty"`

    // AbsoluteRateLimitExemptPaths lists absolute paths exempt from all rate limit quotas.
    // Must be called from the root namespace.
    // +kubebuilder:validation:Optional
    // +listType=set
    AbsoluteRateLimitExemptPaths []string `json:"absoluteRateLimitExemptPaths,omitempty"`
}
```

**Key implementation details:**

```go
func (d *QuotaConfig) GetPath() string {
    return "sys/quotas/config"  // Fixed path — no CleansePath needed, no name segment
}

func (d *QuotaConfig) IsDeletable() bool {
    return false  // Global config cannot be deleted from Vault
}

func (d *QuotaConfig) GetPayload() map[string]any {
    return d.Spec.QuotaConfigData.toMap()
}
```

**`toMap()` implementation:**

```go
func (c *QuotaConfigData) toMap() map[string]any {
    payload := map[string]any{
        "enable_rate_limit_audit_logging":    c.EnableRateLimitAuditLogging,
        "enable_rate_limit_response_headers": c.EnableRateLimitResponseHeaders,
    }
    if len(c.RateLimitExemptPaths) > 0 {
        payload["rate_limit_exempt_paths"] = toInterfaceArray(c.RateLimitExemptPaths)
    }
    if len(c.AbsoluteRateLimitExemptPaths) > 0 {
        payload["absolute_rate_limit_exempt_paths"] = toInterfaceArray(c.AbsoluteRateLimitExemptPaths)
    }
    return payload
}
```

**`IsEquivalentToDesiredState` notes:**
- Vault returns `rate_limit_exempt_paths` as `[]interface{}` — use `toInterfaceArray()` in `toMap()` for match.
- Sort string slices before comparison (like CFAuthEngineConfig does with `sortAnyStringSlice`).
- The bool fields compare directly — Vault returns native bools.
- If `absolute_rate_limit_exempt_paths` is not returned by Vault on read, delete it from desiredState before comparison.

**Webhook:**
- `ValidateCreate`: No special validation (all fields optional).
- `ValidateUpdate`: No immutable fields (singleton, no path/name). Call `newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication)`.
- `ValidateDelete`: No-op.
- `Default`: No-op.

**Singleton convention:** QuotaConfig has no `spec.name` and no `spec.path`. Only one QuotaConfig CR should exist per Vault instance. If multiple CRs exist targeting the same Vault, last-reconciled wins. Document this convention. Do NOT implement a webhook-based uniqueness check — that would require a kube client in the webhook, adding significant complexity for low value.

### CRD 2: RateLimitQuota (Named Resource)

**Pattern:** Follows `PasswordPolicy` closely — named resource at a fixed sys path, standard VaultResource reconciler, no credentials.

**Vault write params (POST `sys/quotas/rate-limit/{name}`):**
- `rate` (float, required, must be positive)
- `path` (string, optional — scopes quota to mount/namespace, empty = global)
- `interval` (string duration, optional, default "1s")
- `block_interval` (string duration, optional)
- `role` (string, optional — scopes to auth role)
- `inheritable` (bool, optional — Enterprise namespace feature)

**Vault read response (GET `sys/quotas/rate-limit/{name}`):**
```json
{
  "block_interval": 300,
  "group_by": "ip",
  "interval": 2,
  "name": "global-rate-limiter",
  "path": "",
  "rate": 897.3,
  "role": "",
  "secondary_rate": 0,
  "type": "rate-limit"
}
```

**⚠️ Critical Vault-read format observations:**
- `interval` and `block_interval` are returned as **integer seconds** (not duration strings). `toMap()` MUST use `durationToSeconds()`.
- `rate` is returned as a float number. With `json.Decoder.UseNumber()`, it arrives as `json.Number`. `toMap()` MUST emit `json.Number`.
- `name`, `type`, `group_by`, `secondary_rate` are auto-populated by Vault — NOT included in `toMap()`. `filterPayloadToDesiredKeys` excludes them automatically.
- `path` and `role` are returned even when empty (`""`). If the CRD field is empty, omit from `toMap()` — `filterPayloadToDesiredKeys` handles the mismatch.

**Type structure:**

```go
// RateLimitQuotaSpec defines the desired state of RateLimitQuota
type RateLimitQuotaSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication"`

    RateLimitQuotaConfig `json:",inline"`

    // The name of the quota created in Vault. If specified, takes precedence over metadata.name.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name,omitempty"`
}

type RateLimitQuotaConfig struct {
    // Rate is the maximum number of requests per interval. Must be positive.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=0
    // +kubebuilder:validation:ExclusiveMinimum=true
    Rate float64 `json:"rate"`

    // Path scopes this quota to a specific mount or namespace path.
    // An empty path creates a global quota.
    // +kubebuilder:validation:Optional
    QuotaPath string `json:"quotaPath,omitempty"`

    // Interval is the duration window for rate limiting (e.g., "1s", "5m"). Defaults to "1s".
    // +kubebuilder:validation:Optional
    Interval string `json:"interval,omitempty"`

    // BlockInterval is the duration a client is blocked after exceeding the rate limit (e.g., "30s", "5m").
    // +kubebuilder:validation:Optional
    BlockInterval string `json:"blockInterval,omitempty"`

    // Role scopes this quota to a specific auth role. Only valid when quotaPath points to an auth mount.
    // +kubebuilder:validation:Optional
    Role string `json:"role,omitempty"`

    // Inheritable makes this quota apply to all child namespaces. Enterprise only.
    // Only valid when quotaPath is set to a namespace.
    // +kubebuilder:validation:Optional
    Inheritable bool `json:"inheritable,omitempty"`
}
```

**⚠️ Critical naming decision — `quotaPath` vs `path`:** The Vault API field is `path`, but this project reserves `spec.path` (type `vaultutils.Path`) for engine mount routing in `GetPath()`. The quota's `path` is a payload data field that scopes the quota. Using `quotaPath` (JSON tag `json:"quotaPath"`) avoids confusion. The `toMap()` method maps `quotaPath` → `"path"` in the Vault payload.

**Key implementation:**

```go
func (d *RateLimitQuota) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath("sys/quotas/rate-limit/" + d.Spec.Name)
    }
    return vaultutils.CleansePath("sys/quotas/rate-limit/" + d.Name)
}

func (d *RateLimitQuota) IsDeletable() bool {
    return true
}
```

**`toMap()` implementation:**

```go
func (c *RateLimitQuotaConfig) toMap() map[string]any {
    payload := map[string]any{
        "rate": json.Number(strconv.FormatFloat(c.Rate, 'f', -1, 64)),
    }
    if c.QuotaPath != "" {
        payload["path"] = c.QuotaPath
    }
    if c.Interval != "" {
        payload["interval"] = durationToSeconds(c.Interval)
    }
    if c.BlockInterval != "" {
        payload["block_interval"] = durationToSeconds(c.BlockInterval)
    }
    if c.Role != "" {
        payload["role"] = c.Role
    }
    if c.Inheritable {
        payload["inheritable"] = c.Inheritable
    }
    return payload
}
```

**`IsEquivalentToDesiredState`:** Standard pattern — `filterPayloadToDesiredKeys` handles extra Vault fields (`name`, `type`, `group_by`, `secondary_rate`).

```go
func (d *RateLimitQuota) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.RateLimitQuotaConfig.toMap()
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

**Webhook:**
- `ValidateCreate`: Validate `rate > 0` (belt-and-suspenders with kubebuilder marker).
- `ValidateUpdate`: Reject changes to `spec.name` (immutable). Call `ValidateTargetNamespaceUpdate`. Validate `rate > 0`.
- No `spec.path` immutability (RateLimitQuota has no `vaultutils.Path` field).
- `Default`: No-op.

### CRD 3: LeaseCountQuota (Named Resource, Enterprise)

**Pattern:** Nearly identical to RateLimitQuota but simpler — fewer fields.

**Vault write params (POST `sys/quotas/lease-count/{name}`):**
- `max_leases` (int, required)
- `path` (string, optional — scopes to mount/namespace, empty = global)
- `role` (string, optional — scopes to auth role)
- `inheritable` (bool, optional — Enterprise namespace feature)

**Vault read response (GET `sys/quotas/lease-count/{name}`):**
```json
{
  "max_leases": 1000,
  "name": "global-lease-count-quota",
  "path": "",
  "role": "",
  "type": "lease-count"
}
```

**Type structure:**

```go
type LeaseCountQuotaSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication"`

    LeaseCountQuotaConfig `json:",inline"`

    // The name of the quota created in Vault. If specified, takes precedence over metadata.name.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name,omitempty"`
}

type LeaseCountQuotaConfig struct {
    // MaxLeases is the maximum number of leases allowed by this quota.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=1
    MaxLeases int `json:"maxLeases"`

    // Path scopes this quota to a specific mount or namespace path.
    // An empty path creates a global quota.
    // +kubebuilder:validation:Optional
    QuotaPath string `json:"quotaPath,omitempty"`

    // Role scopes this quota to a specific auth role. Only valid when quotaPath points to an auth mount.
    // +kubebuilder:validation:Optional
    Role string `json:"role,omitempty"`

    // Inheritable makes this quota apply to all child namespaces. Enterprise only.
    // Only valid when quotaPath is set to a namespace.
    // +kubebuilder:validation:Optional
    Inheritable bool `json:"inheritable,omitempty"`
}
```

**`toMap()` implementation:**

```go
func (c *LeaseCountQuotaConfig) toMap() map[string]any {
    payload := map[string]any{
        "max_leases": json.Number(strconv.Itoa(c.MaxLeases)),
    }
    if c.QuotaPath != "" {
        payload["path"] = c.QuotaPath
    }
    if c.Role != "" {
        payload["role"] = c.Role
    }
    if c.Inheritable {
        payload["inheritable"] = c.Inheritable
    }
    return payload
}
```

**Key implementation:**

```go
func (d *LeaseCountQuota) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath("sys/quotas/lease-count/" + d.Spec.Name)
    }
    return vaultutils.CleansePath("sys/quotas/lease-count/" + d.Name)
}

func (d *LeaseCountQuota) IsDeletable() bool {
    return true
}
```

**Webhook:**
- `ValidateCreate`: Validate `maxLeases >= 1`.
- `ValidateUpdate`: Reject changes to `spec.name` (immutable). Call `ValidateTargetNamespaceUpdate`. Validate `maxLeases >= 1`.
- `Default`: No-op.

### Controller Pattern (All Three Types)

All three controllers follow the `PasswordPolicy` controller pattern exactly — simplest possible:

```go
type QuotaConfigReconciler struct {
    vaultresourcecontroller.ReconcilerBase
}

func (r *QuotaConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    instance := &redhatcopv1alpha1.QuotaConfig{}
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
        return vaultresourcecontroller.ManageOutcome(ctx, r.ReconcilerBase, instance, err)
    }
    vaultResource := vaultresourcecontroller.NewVaultResource(&r.ReconcilerBase, instance)
    return vaultResource.Reconcile(ctx1, instance)
}

func (r *QuotaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&redhatcopv1alpha1.QuotaConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
        Complete(r)
}
```

No credential watching, no Secret/RandomSecret watches — just the basic `For()` + predicate pattern.

### Registration in `cmd/main.go`

**Controllers** — add after line 556 (after KerberosAuthEngineGroup), before the `if webhooks` block:

```go
if err = (&controller.QuotaConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "QuotaConfig")}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "QuotaConfig")
    os.Exit(1)
}
if err = (&controller.RateLimitQuotaReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RateLimitQuota")}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "RateLimitQuota")
    os.Exit(1)
}
if err = (&controller.LeaseCountQuotaReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LeaseCountQuota")}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "LeaseCountQuota")
    os.Exit(1)
}
```

**Webhooks** — add before the `}` closing the `ENABLE_WEBHOOKS` block (line 936), after KerberosAuthEngineGroup webhook:

```go
if err = (&redhatcopv1alpha1.QuotaConfig{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create webhook", "webhook", "QuotaConfig")
    os.Exit(1)
}
if err = (&redhatcopv1alpha1.RateLimitQuota{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create webhook", "webhook", "RateLimitQuota")
    os.Exit(1)
}
if err = (&redhatcopv1alpha1.LeaseCountQuota{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create webhook", "webhook", "LeaseCountQuota")
    os.Exit(1)
}
```

### RBAC Markers

Each controller needs these markers:

```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=quotaconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=quotaconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=quotaconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

(Replace `quotaconfigs` with `ratelimitquotas` / `leasecountquotas` for the other two.)

### Unit Test Strategy

Use `testing.Test*` table-driven tests following the project standard. **Never derive expected payloads from code under test.**

**For each type, test:**

1. **`toMap()` output** — construct an instance, call `toMap()`, compare against hardcoded expected map.
2. **`IsEquivalentToDesiredState` positive** — construct a Vault-read-shaped payload (with `json.Number` for numerics, extra fields like `name`/`type`) and assert `true`.
3. **`IsEquivalentToDesiredState` negative** — change a managed field value in the Vault payload and assert `false`.
4. **`IsEquivalentToDesiredState` extra-fields** — include Vault-added fields (`name`, `type`, `group_by`, etc.) and assert they don't cause false drift.

**RateLimitQuota-specific test cases:**
- `interval: "2m"` → `toMap()` emits `json.Number("120")` (via `durationToSeconds`)
- `blockInterval: "5m"` → `toMap()` emits `json.Number("300")`
- `rate: 897.3` → `toMap()` emits `json.Number("897.3")`
- `rate: 100` → `toMap()` emits `json.Number("100")` (not "100.0")
- Optional fields omitted when zero/empty — `filterPayloadToDesiredKeys` handles extra Vault fields

**QuotaConfig-specific test cases:**
- `rateLimitExemptPaths` sorting — test that ordering differences don't cause false drift
- Verify `absolute_rate_limit_exempt_paths` handling (see dev note above)

### Integration Test Strategy

**QuotaConfig** (`internal/controller/quotaconfig_controller_test.go`):
- `//go:build integration`
- `Describe("QuotaConfig controller", Ordered, func() {...})`
- Create QuotaConfig CR → wait for ReconcileSuccessful
- Read `sys/quotas/config` from Vault → verify field values
- Update CR (change `enableRateLimitAuditLogging`) → wait for reconcile → verify Vault state updated
- Delete CR → verify Vault config **persists** (IsDeletable=false non-deletable type delete verification per project-context rule)

**RateLimitQuota** (`internal/controller/ratelimitquota_controller_test.go`):
- `//go:build integration`
- `Describe("RateLimitQuota controller", Ordered, func() {...})`
- Create RateLimitQuota CR → wait for ReconcileSuccessful
- Read `sys/quotas/rate-limit/{name}` from Vault → verify fields
- Update CR (change rate) → wait for reconcile → verify updated
- Delete CR → wait for K8s NotFound → verify Vault quota deleted (read returns 404)

**LeaseCountQuota** — **SKIP integration test.** Enterprise-only API. The OSS Vault deployed in the Kind cluster cannot serve `/sys/quotas/lease-count`. Unit tests provide coverage.

**Test fixtures:**

`test/quotaconfig/quota-config.yaml`:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: QuotaConfig
metadata:
  name: quota-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  enableRateLimitAuditLogging: true
  rateLimitExemptPaths:
    - "sys/health"
```

`test/quotaconfig/rate-limit-quota.yaml`:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RateLimitQuota
metadata:
  name: test-rate-limiter
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  rate: 100
  interval: "1m"
```

### File Structure

| File | Action | Purpose |
|---|---|---|
| `api/v1alpha1/quotaconfig_types.go` | **CREATE** | QuotaConfig CRD type, VaultObject, toMap(), IsEquivalentToDesiredState() |
| `api/v1alpha1/quotaconfig_webhook.go` | **CREATE** | QuotaConfig admission webhooks |
| `api/v1alpha1/quotaconfig_test.go` | **CREATE** | Unit tests for QuotaConfig toMap/IsEquivalent |
| `api/v1alpha1/ratelimitquota_types.go` | **CREATE** | RateLimitQuota CRD type, VaultObject, toMap(), IsEquivalentToDesiredState() |
| `api/v1alpha1/ratelimitquota_webhook.go` | **CREATE** | RateLimitQuota admission webhooks |
| `api/v1alpha1/ratelimitquota_test.go` | **CREATE** | Unit tests for RateLimitQuota toMap/IsEquivalent |
| `api/v1alpha1/leasecountquota_types.go` | **CREATE** | LeaseCountQuota CRD type, VaultObject, toMap(), IsEquivalentToDesiredState() |
| `api/v1alpha1/leasecountquota_webhook.go` | **CREATE** | LeaseCountQuota admission webhooks |
| `api/v1alpha1/leasecountquota_test.go` | **CREATE** | Unit tests for LeaseCountQuota toMap/IsEquivalent |
| `internal/controller/quotaconfig_controller.go` | **CREATE** | QuotaConfig reconciler |
| `internal/controller/ratelimitquota_controller.go` | **CREATE** | RateLimitQuota reconciler |
| `internal/controller/leasecountquota_controller.go` | **CREATE** | LeaseCountQuota reconciler |
| `internal/controller/quotaconfig_controller_test.go` | **CREATE** | Integration test for QuotaConfig |
| `internal/controller/ratelimitquota_controller_test.go` | **CREATE** | Integration test for RateLimitQuota |
| `test/quotaconfig/quota-config.yaml` | **CREATE** | QuotaConfig test fixture |
| `test/quotaconfig/rate-limit-quota.yaml` | **CREATE** | RateLimitQuota test fixture |
| `docs/system-backend/quotas.md` | **CREATE** | Documentation for quota CRDs |
| `cmd/main.go` | **MODIFY** | Register 3 controllers + 3 webhooks |
| `config/crd/kustomization.yaml` | **MODIFY** | Add 3 new CRD YAML files |

### Shared Framework Files — NO Changes

This story does **not** modify any shared framework files. All code is new CRD types, controllers, and webhooks.

### Existing Utilities to Reuse

| Utility | Location | Usage |
|---------|----------|-------|
| `durationToSeconds()` | `api/v1alpha1/sshsecretenginerole_types.go` | Convert interval/blockInterval to json.Number seconds |
| `filterPayloadToDesiredKeys()` | `api/v1alpha1/payload_filter.go` | Filter Vault read response to managed keys |
| `toInterfaceArray()` | `api/v1alpha1/databasesecretengineconfig_types.go` | Convert []string to []any for list comparison |
| `sortAnyStringSlice()` | `api/v1alpha1/cfauthengineconfig_types.go` | Sort string slices before comparison |
| `removeUnsetFields()` | `api/v1alpha1/payload_filter.go` | Remove zero-value fields absent from Vault response |
| `CleansePath()` | `api/v1alpha1/utils/commons.go` | Normalize Vault paths |
| `vaultresourcecontroller.NewVaultResource()` | `internal/controller/vaultresourcecontroller/` | Standard reconciler factory |
| `vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate()` | Same package | Generation-based predicate for drift detection |

**Note on `durationToSeconds()` location:** It's defined in `sshsecretenginerole_types.go` (package `v1alpha1`), accessible to all types in the same package. No import needed.

### CRD Registration Checklist

After `make manifests`:
1. Add `bases/redhatcop.redhat.io_quotaconfigs.yaml` to `config/crd/kustomization.yaml`
2. Add `bases/redhatcop.redhat.io_ratelimitquotas.yaml` to `config/crd/kustomization.yaml`
3. Add `bases/redhatcop.redhat.io_leasecountquotas.yaml` to `config/crd/kustomization.yaml`

### Import Lists

**`_types.go` files:**
```go
import (
    "context"
    "encoding/json"
    "reflect"
    "strconv"

    vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)
```
(`"strconv"` needed for `json.Number(strconv.Itoa(...))` and `strconv.FormatFloat`. `"encoding/json"` for `json.Number`. QuotaConfig may not need `strconv` if it has no numeric fields.)

**`_webhook.go` files:**
```go
import (
    "context"
    "errors"

    ctrl "sigs.k8s.io/controller-runtime"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)
```

**`_controller.go` files:**
```go
import (
    "context"

    apierrors "k8s.io/apimachinery/pkg/api/errors"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/builder"
    "sigs.k8s.io/controller-runtime/pkg/reconcile"

    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
)
```

- **Novelty Risk:** LOW — follows established CRD patterns (Policy, PasswordPolicy). The only novel element is the singleton QuotaConfig (no name segment in path, IsDeletable=false). RateLimitQuota and LeaseCountQuota are standard named resources.
- Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings.

### Project Structure Notes

- All new files follow the standard naming convention: lowercase type name as filename prefix.
- `docs/system-backend/` is a new directory — the first system backend docs. Create it with this story.
- Test fixtures in `test/quotaconfig/` — shared directory for all quota-related test YAML.
- No conflicts with existing files — all files are new.

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 17, Story 17.1, lines 2877–2933]
- [Source: _bmad-output/project-context.md — Adding a New Vault API Type checklist, lines 300-311]
- [Source: _bmad-output/project-context.md — CRD Field Default & Validation Rules, lines 240-258]
- [Source: _bmad-output/project-context.md — toMap() output normalization rule, lines 88-90]
- [Source: api/v1alpha1/passwordpolicy_types.go — Reference pattern for simple sys/ CRD]
- [Source: api/v1alpha1/passwordpolicy_webhook.go — Reference webhook pattern]
- [Source: internal/controller/passwordpolicy_controller.go — Reference controller pattern]
- [Source: api/v1alpha1/cfauthengineconfig_types.go — Reference for toMap() with json.Number, sortAnyStringSlice]
- [Source: api/v1alpha1/sshsecretenginerole_types.go:216-228 — durationToSeconds() helper]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys, removeUnsetFields]
- [Source: Vault API — sys/quotas/config](https://developer.hashicorp.com/vault/api-docs/system/quotas-config)
- [Source: Vault API — sys/quotas/rate-limit](https://developer.hashicorp.com/vault/api-docs/system/rate-limit-quotas)
- [Source: Vault API — sys/quotas/lease-count](https://developer.hashicorp.com/vault/api-docs/system/lease-count-quotas)
- [Source: GitHub Issue #249](https://github.com/redhat-cop/vault-config-operator/issues/249)

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

### Review Model Used

- Quick lens: GPT 5.4 Medium (adversarial review)

### Review Findings

Pass 1 — 5 findings: 0 high, 3 medium, 1 low, 2 false
- `[false]` `[reject]` Rate field as string instead of float64 — Kubernetes auto-coerces YAML numbers to strings for string-typed CRD fields
- `[low]` `[reject]` Omitting interval sends 0 instead of Vault default 1s — follows majority project pattern; Vault likely treats 0 as "use default"
- `[medium]` `[patch]` RateLimitQuota spec.name guard allows empty-to-non-empty mutation — removed extra guard to match existing webhook pattern
- `[medium]` `[patch]` LeaseCountQuota spec.name — same immutability hole — same fix as above
- `[medium]` `[patch]` QuotaConfig IsEquivalentToDesiredState doesn't sort string slices — added sortAnyStringSlice calls

### Decisions Needed / Decisions Taken

None — all findings resolved during review pass.

### Fixes Applied

- Removed extra spec.name immutability guard in RateLimitQuota and LeaseCountQuota webhooks
- Added sortAnyStringSlice calls to QuotaConfig IsEquivalentToDesiredState
