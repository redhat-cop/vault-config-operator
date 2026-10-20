# Story 17.4: UI and Observability CRDs (UICustomHeader + LoggerConfig)

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for UICustomHeader and LoggerConfig,
So that Vault's custom UI response headers and runtime log levels can be managed declaratively.

## Acceptance Criteria

### UICustomHeader

1. **Given** a UICustomHeader CR is created with header `values` **When** the reconciler processes it **Then** the header is configured in Vault at `sys/config/ui/headers/{name}` and `ReconcileSuccessful=True`
2. **Given** a UICustomHeader CR is updated (e.g., values list changed) **When** the reconciler processes the update **Then** the Vault header configuration is updated to match
3. **Given** a UICustomHeader CR is deleted **When** the reconciler processes deletion **Then** the header is removed from Vault
4. **Given** the webhook validates updates **When** a user attempts to change `spec.name` **Then** the update is rejected with `"spec.name cannot be updated"`
5. **Given** the webhook validates creates/updates **When** `values` is empty **Then** the request is rejected

### LoggerConfig

6. **Given** a LoggerConfig CR is created with a logger name and `level` (trace, debug, info, warn, error) **When** the reconciler processes it **Then** the logger level is set in Vault at `sys/loggers/{name}` and `ReconcileSuccessful=True`
7. **Given** a LoggerConfig CR is updated (e.g., level changed) **When** the reconciler processes the update **Then** the Vault logger level is updated to match
8. **Given** a LoggerConfig CR is deleted **When** the reconciler processes deletion **Then** the logger setting is reverted to Vault's configured default
9. **Given** the webhook validates creates/updates **When** a user provides an invalid `level` **Then** the request is rejected via kubebuilder Enum validation
10. **Given** the webhook validates updates **When** a user attempts to change `spec.name` **Then** the update is rejected with `"spec.name cannot be updated"`

### Cross-Cutting

11. **Given** the CRD types are implemented **When** the story is marked done **Then** documentation exists at `docs/system-backend/ui-headers.md` and `docs/system-backend/loggers.md`
12. **Given** the CRD types are implemented **When** `make manifests generate fmt vet test` and `make lint` are run **Then** all pass with zero errors

## Tasks / Subtasks

- [ ] Task 1 — UICustomHeader CRD type (AC: #1, #2)
  - [ ] 1.1 Create `api/v1alpha1/uicustomheader_types.go`: UICustomHeaderSpec with `Connection`, `Authentication`, inline `UICustomHeaderConfig`, `Name`; `GetPath()` → `"sys/config/ui/headers/{name}"`; `IsDeletable()` → `true`; `toMap()` emitting `values` list
  - [ ] 1.2 Implement custom `IsEquivalentToDesiredState()` that handles Vault's read response shape mismatch (see Dev Notes — Critical API Quirk)
  - [ ] 1.3 Register `&UICustomHeader{}` and `&UICustomHeaderList{}` in `init()`
- [ ] Task 2 — UICustomHeader webhook (AC: #4, #5)
  - [ ] 2.1 Create `api/v1alpha1/uicustomheader_webhook.go`: `admission.Defaulter[*UICustomHeader]`, `admission.Validator[*UICustomHeader]`
  - [ ] 2.2 `ValidateCreate`: reject if `Values` is empty (len == 0)
  - [ ] 2.3 `ValidateUpdate`: call `Authentication.ValidateTargetNamespaceUpdate()`, reject `spec.name` changes (immutable), reject if `Values` is empty
  - [ ] 2.4 Add kubebuilder webhook markers for mutating and validating paths
- [ ] Task 3 — UICustomHeader controller (AC: #1, #2, #3)
  - [ ] 3.1 Create `internal/controller/uicustomheader_controller.go` following PasswordPolicy controller pattern
  - [ ] 3.2 Add RBAC markers for `uicustomheaders` resource
- [ ] Task 4 — LoggerConfig CRD type (AC: #6, #7, #8)
  - [ ] 4.1 Create `api/v1alpha1/loggerconfig_types.go`: LoggerConfigSpec with `Connection`, `Authentication`, inline `LoggerConfigConfig`, `Name`; `GetPath()` → `"sys/loggers/{name}"`; `IsDeletable()` → `true`
  - [ ] 4.2 Implement `toMap()` emitting `{"level": "..."}` (write payload shape)
  - [ ] 4.3 Implement custom `IsEquivalentToDesiredState()` that handles Vault read response shape mismatch (see Dev Notes — Critical API Quirk)
  - [ ] 4.4 Register `&LoggerConfig{}` and `&LoggerConfigList{}` in `init()`
- [ ] Task 5 — LoggerConfig webhook (AC: #9, #10)
  - [ ] 5.1 Create `api/v1alpha1/loggerconfig_webhook.go`: `admission.Defaulter[*LoggerConfig]`, `admission.Validator[*LoggerConfig]`
  - [ ] 5.2 `ValidateCreate`/`ValidateUpdate`: `level` validated via `+kubebuilder:validation:Enum` (no runtime check needed)
  - [ ] 5.3 `ValidateUpdate`: call `Authentication.ValidateTargetNamespaceUpdate()`, reject `spec.name` changes (immutable)
  - [ ] 5.4 Add kubebuilder webhook markers
- [ ] Task 6 — Registration (AC: #1, #6)
  - [ ] 6.1 Register both controllers in `cmd/main.go`
  - [ ] 6.2 Register both webhooks in `cmd/main.go` inside `ENABLE_WEBHOOKS` guard
- [ ] Task 7 — Unit tests (AC: #1, #2, #4, #5, #6, #7, #9, #10)
  - [ ] 7.1 Create `api/v1alpha1/uicustomheader_test.go`: `GetPath` (with/without spec.name), `GetPayload`/toMap, `IsEquivalentToDesiredState` positive (both `value`/`values` read shapes) + negative + extra fields, `IsDeletable` true, `ConditionsAware`
  - [ ] 7.2 Create `api/v1alpha1/loggerconfig_test.go`: `GetPath` (with/without spec.name), `GetPayload`/toMap, `IsEquivalentToDesiredState` positive (logger-name-keyed read shape) + negative + extra fields, `IsDeletable` true, `ConditionsAware`
- [ ] Task 8 — Integration tests (AC: #1, #2, #3, #6, #7, #8)
  - [ ] 8.1 Create test fixtures: `test/uicustomheader/simple-ui-custom-header.yaml`, `test/uicustomheader/named-ui-custom-header.yaml`
  - [ ] 8.2 Create test fixtures: `test/loggerconfig/simple-logger-config.yaml`, `test/loggerconfig/named-logger-config.yaml`
  - [ ] 8.3 Create `internal/controller/uicustomheader_controller_test.go` with `//go:build integration`: create → reconcile success → verify Vault state → delete → verify removal
  - [ ] 8.4 Create `internal/controller/loggerconfig_controller_test.go` with `//go:build integration`: create → reconcile success → verify Vault state → delete → verify logger reverts
- [ ] Task 9 — Code generation + CRD registration (AC: #12)
  - [ ] 9.1 Run `make manifests generate fmt vet`
  - [ ] 9.2 Add new CRD YAML files to `config/crd/kustomization.yaml` resources list
  - [ ] 9.3 Run `make test` — all unit tests pass
  - [ ] 9.4 Run `make lint` — zero violations
- [ ] Task 10 — Documentation (AC: #11)
  - [ ] 10.1 Create `docs/system-backend/ui-headers.md` following `docs/engine-doc-template.md` adapted for system backend
  - [ ] 10.2 Create `docs/system-backend/loggers.md` following `docs/engine-doc-template.md` adapted for system backend
  - [ ] 10.3 Create `docs/system-backend/index.md` if not already present from Story 17.2 or 17.3

## Dev Notes

### Pattern: Named Resources with `spec.name` Override

Both CRDs are standard **named resources** following the PasswordPolicy pattern. Each has a `spec.name` field that overrides `metadata.name` for the Vault object path. Key correspondences:

| Aspect | PasswordPolicy | UICustomHeader | LoggerConfig |
|---|---|---|---|
| Path | `sys/policies/password/{name}` | `sys/config/ui/headers/{name}` | `sys/loggers/{name}` |
| Reconciler | `VaultResource` (standard) | `VaultResource` (standard) | `VaultResource` (standard) |
| `IsDeletable` | `true` | `true` | `true` |
| Has `spec.path` | No (hardcoded prefix) | No (hardcoded prefix) | No (hardcoded prefix) |
| Has `spec.name` | Yes | Yes | Yes |
| Custom `IsEquivalentToDesiredState` | No | **Yes** (see API quirk below) | **Yes** (see API quirk below) |

### ⚠️ Critical API Quirk — Vault Read/Write Shape Mismatch

**Both types have Vault APIs where the read response shape differs from the write payload shape.** This requires custom `IsEquivalentToDesiredState` implementations. Without them, the standard `filterPayloadToDesiredKeys` + `reflect.DeepEqual` comparison would fail every reconcile, causing unnecessary writes.

#### UICustomHeader Read/Write Mismatch

**Write payload:** `POST sys/config/ui/headers/{name}` with `{"values": ["val1", "val2"]}`

**Read response:** `GET sys/config/ui/headers/{name}` returns either:
- Single value: `{"value": "val1"}` (without `?multivalue=true` query param)
- Multiple values: `{"values": ["val1", "val2"]}` (with `?multivalue=true` query param)

The framework's `read()` function in `api/v1alpha1/utils/vaultutils.go` calls `vaultClient.Logical().Read(path)` which does a plain GET without query parameters. There is no `ReadWithData` support in the VaultEndpoint framework. Therefore the standard read returns the **single-value format** (`value` key, string type).

**Custom `IsEquivalentToDesiredState` implementation:**

```go
func (d *UICustomHeader) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredValues := d.Spec.Values

    // Case 1: Vault returned multivalue format (if API evolves or proxy adds param)
    if vaultValues, ok := payload["values"]; ok {
        return reflect.DeepEqual(toInterfaceArray(desiredValues), vaultValues)
    }

    // Case 2: Vault returned single-value format (standard GET behavior)
    if vaultValue, ok := payload["value"]; ok {
        if len(desiredValues) == 1 {
            return desiredValues[0] == vaultValue
        }
        // Multi-value desired but only single-value readable — always re-write.
        // Vault POST is idempotent so this is safe. Document as known limitation.
        return false
    }

    // No value key at all — empty response means not configured
    return len(desiredValues) == 0
}
```

**Known limitation:** For UICustomHeader CRs with multiple values, the operator will re-write on every reconcile because the standard read can only retrieve the first value. This is acceptable: the POST is idempotent and reconcile frequency is controlled by `SYNC_PERIOD_SECONDS` (default 10hrs). Document this in the CRD docs.

#### LoggerConfig Read/Write Mismatch

**Write payload:** `POST sys/loggers/{name}` with `{"level": "debug"}`

**Read response:** `GET sys/loggers/{name}` returns `{"core": "info"}` — the key is the **logger name**, not `"level"`.

The desired state from `toMap()` has shape `{"level": "info"}` but the Vault read response has shape `{"core": "info"}`. `filterPayloadToDesiredKeys` would find no matching keys and return `{}`, causing `reflect.DeepEqual({"level": "info"}, {})` → `false` → unnecessary write every reconcile.

**Custom `IsEquivalentToDesiredState` implementation:**

```go
func (d *LoggerConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
    loggerName := d.Spec.Name
    if loggerName == "" {
        loggerName = d.Name
    }

    // Vault read format: {loggerName: level}
    if vaultLevel, ok := payload[loggerName]; ok {
        return d.Spec.Level == vaultLevel
    }

    // Fallback: check write format (defensive)
    if level, ok := payload["level"]; ok {
        return d.Spec.Level == level
    }

    return false
}
```

### LoggerConfig — Transient Nature (Document Prominently)

Per the Vault API docs: *"Changes made to the log level using this endpoint are not persisted and will be restored to either the default log level (info) or the level specified using `log_level` in vault.hcl or the `VAULT_LOG_LEVEL` environment variable once the Vault service is reloaded or restarted."*

This means:
- After Vault restart, logger overrides are lost
- The operator's periodic reconcile (drift detection via `RequeueAfter`) re-applies the desired level
- This is **correct declarative behavior** — the operator continuously enforces the desired state
- Document this prominently in the CRD documentation

### UICustomHeader — `sudo` Capability Required

Per the Vault API docs: *"All of these endpoints require the `sudo` capability in addition to any path-specific capabilities."*

The Vault token used by the operator must have `sudo` on `sys/config/ui/headers/*`. Document this in the CRD docs. The test infrastructure uses a root token which has all capabilities including `sudo`.

### UICustomHeader — Type File Structure (`uicustomheader_types.go`)

```go
package v1alpha1

import (
    "context"
    "reflect"

    vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ vaultutils.VaultObject = &UICustomHeader{}
var _ vaultutils.ConditionsAware = &UICustomHeader{}

// UICustomHeaderSpec defines the desired state of UICustomHeader
type UICustomHeaderSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Optional
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    UICustomHeaderConfig `json:",inline"`

    // The name of the custom header in Vault. If specified, takes precedence over {metadata.name}
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?`
    Name string `json:"name,omitempty"`
}

// UICustomHeaderConfig contains the Vault UI custom header configuration fields.
type UICustomHeaderConfig struct {
    // Values is the list of values to return for this custom header.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinItems=1
    Values []string `json:"values"`
}
```

**Important — `spec.name` pattern validation:** HTTP header names can contain uppercase, lowercase, digits, and hyphens (e.g., `X-Custom-Header`). The pattern must allow uppercase letters, unlike most CRD Name fields. Use `[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?` instead of the standard lowercase pattern.

### `UICustomHeader.toMap()` Implementation

```go
func (c *UICustomHeaderConfig) toMap() map[string]any {
    return map[string]any{
        "values": toInterfaceArray(c.Values),
    }
}
```

The `toInterfaceArray` helper is already defined in `databasesecretengineconfig_types.go` (package-level function in `api/v1alpha1`). It converts `[]string` to `[]any` for Vault API compatibility.

### `UICustomHeader.GetPath()` Implementation

```go
func (d *UICustomHeader) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath("sys/config/ui/headers/" + d.Spec.Name)
    }
    return vaultutils.CleansePath("sys/config/ui/headers/" + d.Name)
}
```

### LoggerConfig — Type File Structure (`loggerconfig_types.go`)

```go
package v1alpha1

import (
    "context"

    vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ vaultutils.VaultObject = &LoggerConfig{}
var _ vaultutils.ConditionsAware = &LoggerConfig{}

// LoggerConfigSpec defines the desired state of LoggerConfig
type LoggerConfigSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Optional
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    LoggerConfigConfig `json:",inline"`

    // The name of the Vault logger to configure (e.g., "audit", "core", "expiration").
    // If specified, takes precedence over {metadata.name}
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name,omitempty"`
}

// LoggerConfigConfig contains the Vault logger configuration fields.
type LoggerConfigConfig struct {
    // Level is the log verbosity level. Supported values in order of detail:
    // trace, debug, info, warn, error.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Enum={"trace","debug","info","warn","error"}
    Level string `json:"level"`
}
```

**Key design decisions:**
- `Level` uses `+kubebuilder:validation:Required` — a LoggerConfig without a level is meaningless
- `Level` uses `+kubebuilder:validation:Enum` — validates at admission, no runtime check needed
- `Level` has no `omitempty` — it's required and always present in JSON
- No `+kubebuilder:default` on `Level` — user must explicitly choose a log level

### `LoggerConfig.toMap()` Implementation

```go
func (c *LoggerConfigConfig) toMap() map[string]any {
    return map[string]any{
        "level": c.Level,
    }
}
```

Simple string field — no `json.Number` or `durationToSeconds()` needed. Vault returns a string level, and we write a string level.

### `LoggerConfig.GetPath()` Implementation

```go
func (d *LoggerConfig) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath("sys/loggers/" + d.Spec.Name)
    }
    return vaultutils.CleansePath("sys/loggers/" + d.Name)
}
```

### Webhook Implementations

**UICustomHeader webhook (`uicustomheader_webhook.go`):**

```go
var uicustomheaderlog = logf.Log.WithName("uicustomheader-resource")

// ValidateCreate: reject if Values is empty
func (r *UICustomHeader) ValidateCreate(ctx context.Context, obj *UICustomHeader) (admission.Warnings, error) {
    uicustomheaderlog.Info("validate create", "name", obj.Name)
    if len(obj.Spec.Values) == 0 {
        return nil, errors.New("spec.values must contain at least one value")
    }
    return nil, nil
}

// ValidateUpdate: target namespace, spec.name immutability, Values not empty
func (r *UICustomHeader) ValidateUpdate(ctx context.Context, oldObj, newObj *UICustomHeader) (admission.Warnings, error) {
    uicustomheaderlog.Info("validate update", "name", newObj.Name)
    if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
        return nil, err
    }
    if newObj.Spec.Name != oldObj.Spec.Name {
        return nil, errors.New("spec.name cannot be updated")
    }
    if len(newObj.Spec.Values) == 0 {
        return nil, errors.New("spec.values must contain at least one value")
    }
    return nil, nil
}
```

**Note:** The `Values` empty check in the webhook is defense-in-depth. The `+kubebuilder:validation:MinItems=1` marker also enforces this at the CRD schema level.

**LoggerConfig webhook (`loggerconfig_webhook.go`):**

```go
var loggerconfiglog = logf.Log.WithName("loggerconfig-resource")

// ValidateCreate: level validated by kubebuilder Enum — no runtime check needed
func (r *LoggerConfig) ValidateCreate(ctx context.Context, obj *LoggerConfig) (admission.Warnings, error) {
    loggerconfiglog.Info("validate create", "name", obj.Name)
    return nil, nil
}

// ValidateUpdate: target namespace + spec.name immutability
func (r *LoggerConfig) ValidateUpdate(ctx context.Context, oldObj, newObj *LoggerConfig) (admission.Warnings, error) {
    loggerconfiglog.Info("validate update", "name", newObj.Name)
    if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
        return nil, err
    }
    if newObj.Spec.Name != oldObj.Spec.Name {
        return nil, errors.New("spec.name cannot be updated")
    }
    return nil, nil
}
```

**Kubebuilder markers for both webhooks:**
```
//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-uicustomheader,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=uicustomheaders,verbs=create,versions=v1alpha1,name=muicustomheader.kb.io,admissionReviewVersions={v1,v1beta1}
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-uicustomheader,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=uicustomheaders,verbs=create;update,versions=v1alpha1,name=vuicustomheader.kb.io,admissionReviewVersions={v1,v1beta1}

//+kubebuilder:webhook:path=/mutate-redhatcop-redhat-io-v1alpha1-loggerconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=loggerconfigs,verbs=create,versions=v1alpha1,name=mloggerconfig.kb.io,admissionReviewVersions={v1,v1beta1}
//+kubebuilder:webhook:path=/validate-redhatcop-redhat-io-v1alpha1-loggerconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=redhatcop.redhat.io,resources=loggerconfigs,verbs=create;update,versions=v1alpha1,name=vloggerconfig.kb.io,admissionReviewVersions={v1,v1beta1}
```

### Controller Implementations

Both controllers are carbon copies of `internal/controller/passwordpolicy_controller.go` with type substitutions:

```go
// UICustomHeader
type UICustomHeaderReconciler struct {
    vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=uicustomheaders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=uicustomheaders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=uicustomheaders/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch

// LoggerConfig
type LoggerConfigReconciler struct {
    vaultresourcecontroller.ReconcilerBase
}

//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=loggerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=loggerconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=loggerconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### `cmd/main.go` Registration

**Controllers** — insert alphabetically near other Epic 17 controllers:
```go
if err = (&controller.LoggerConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LoggerConfig")}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "LoggerConfig")
    os.Exit(1)
}
// ...
if err = (&controller.UICustomHeaderReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "UICustomHeader")}).SetupWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create controller", "controller", "UICustomHeader")
    os.Exit(1)
}
```

**Webhooks** — insert inside `ENABLE_WEBHOOKS` guard:
```go
if err = (&redhatcopv1alpha1.LoggerConfig{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create webhook", "webhook", "LoggerConfig")
    os.Exit(1)
}
if err = (&redhatcopv1alpha1.UICustomHeader{}).SetupWebhookWithManager(mgr); err != nil {
    setupLog.Error(err, "unable to create webhook", "webhook", "UICustomHeader")
    os.Exit(1)
}
```

### Unit Tests

**`uicustomheader_test.go`** — Required test functions:

1. **`TestUICustomHeaderGetPath`** — table-driven: with `spec.name` (e.g., `"X-Custom-Header"`), without `spec.name` (metadata name fallback)
2. **`TestUICustomHeaderGetPayload`** — verify `toMap()` output: `{"values": []any{"val1", "val2"}}`
3. **`TestUICustomHeaderIsEquivalentToDesiredState`** — positive with single-value read shape: `{"value": "val1"}` when desired is `["val1"]`
4. **`TestUICustomHeaderIsEquivalentMultiValue`** — positive with multi-value read shape: `{"values": []any{"val1", "val2"}}` when desired matches
5. **`TestUICustomHeaderIsEquivalentNegative`** — mismatched single value returns `false`
6. **`TestUICustomHeaderIsEquivalentExtraFields`** — extra Vault-returned keys are handled correctly
7. **`TestUICustomHeaderIsDeletable`** — returns `true`
8. **`TestUICustomHeaderConditions`** — set/get roundtrip

**Critical test construction:** Construct Vault read-shaped maps independently, not from `toMap()`:
```go
// Simulated Vault single-value read response
singleValuePayload := map[string]any{
    "value": "my-header-value",
}

// Simulated Vault multi-value read response
multiValuePayload := map[string]any{
    "values": []any{"val1", "val2"},
}
```

**`loggerconfig_test.go`** — Required test functions:

1. **`TestLoggerConfigGetPath`** — table-driven: with `spec.name` ("core"), without `spec.name` (metadata name)
2. **`TestLoggerConfigGetPayload`** — verify `toMap()` output: `{"level": "debug"}`
3. **`TestLoggerConfigIsEquivalentToDesiredState`** — positive with Vault read shape: `{"core": "debug"}` when desired level is `"debug"` and logger name is `"core"`
4. **`TestLoggerConfigIsEquivalentNegative`** — mismatched level: Vault returns `{"core": "info"}` when desired is `"debug"` → `false`
5. **`TestLoggerConfigIsEquivalentExtraFields`** — extra Vault-returned keys ignored
6. **`TestLoggerConfigIsEquivalentWithSpecName`** — logger name from `spec.name` used as payload key
7. **`TestLoggerConfigIsDeletable`** — returns `true`
8. **`TestLoggerConfigConditions`** — set/get roundtrip

**Critical test construction for LoggerConfig:**
```go
// Simulated Vault read response — key is logger name, value is level
vaultPayload := map[string]any{
    "core": "debug",
}
// Test with spec.name = "core" and level = "debug" → should return true
```

### Integration Tests

**Classification:** Both endpoints are standard `/sys/` endpoints on OSS Vault — **Install in Kind** category. The test Vault root token has all capabilities including `sudo`.

**UICustomHeader integration test (`uicustomheader_controller_test.go`):**
- `Ordered` Describe decorator (cross-context lifecycle state)
- Create simple header → assert `ReconcileSuccessful=True` → verify header in Vault via `GET sys/config/ui/headers/{name}`
- Create named header (spec.name override) → verify at correct Vault path
- Delete both → verify headers removed from Vault

**LoggerConfig integration test (`loggerconfig_controller_test.go`):**
- `Ordered` Describe decorator
- Create logger config (e.g., `core` at `debug`) → assert `ReconcileSuccessful=True` → verify via `GET sys/loggers/core`
- Create named logger config → verify at correct Vault path
- Delete both → verify loggers revert (GET may return the default level or the entry disappears)

**Integration test verification for LoggerConfig read response:** The Vault read for `sys/loggers/{name}` returns `{loggerName: level}`. Read the response and verify using the logger name as key:
```go
secret, err := vaultClient.Logical().Read("sys/loggers/core")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data["core"]).To(Equal("debug"))
```

### Test Fixtures

**`test/uicustomheader/simple-ui-custom-header.yaml`:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: UICustomHeader
metadata:
  name: x-custom-test-header
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  values:
    - "test-value-1"
```

**`test/uicustomheader/named-ui-custom-header.yaml`:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: UICustomHeader
metadata:
  name: test-named-header-metadata
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: x-named-test-header
  values:
    - "named-value-1"
    - "named-value-2"
```

**`test/loggerconfig/simple-logger-config.yaml`:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LoggerConfig
metadata:
  name: core
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  level: debug
```

**`test/loggerconfig/named-logger-config.yaml`:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LoggerConfig
metadata:
  name: test-named-logger-metadata
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: audit
  level: trace
```

### CRD Registration Checklist

After `make manifests`:
1. Verify `config/crd/bases/redhatcop.redhat.io_uicustomheaders.yaml` was generated
2. Verify `config/crd/bases/redhatcop.redhat.io_loggerconfigs.yaml` was generated
3. Add both to `config/crd/kustomization.yaml` resources list
4. Add commented-out webhook and cainjection patch lines (follow PasswordPolicy's pattern)

### File Structure

| File | Action | Purpose |
|---|---|---|
| `api/v1alpha1/uicustomheader_types.go` | **CREATE** | CRD type, VaultObject impl, toMap, custom IsEquivalentToDesiredState |
| `api/v1alpha1/uicustomheader_webhook.go` | **CREATE** | Defaulter + Validator, spec.name immutability, values-not-empty |
| `api/v1alpha1/uicustomheader_test.go` | **CREATE** | Unit tests |
| `api/v1alpha1/loggerconfig_types.go` | **CREATE** | CRD type, VaultObject impl, toMap, custom IsEquivalentToDesiredState |
| `api/v1alpha1/loggerconfig_webhook.go` | **CREATE** | Defaulter + Validator, spec.name immutability |
| `api/v1alpha1/loggerconfig_test.go` | **CREATE** | Unit tests |
| `internal/controller/uicustomheader_controller.go` | **CREATE** | Standard VaultResource controller |
| `internal/controller/loggerconfig_controller.go` | **CREATE** | Standard VaultResource controller |
| `internal/controller/uicustomheader_controller_test.go` | **CREATE** | Integration test (//go:build integration) |
| `internal/controller/loggerconfig_controller_test.go` | **CREATE** | Integration test (//go:build integration) |
| `test/uicustomheader/simple-ui-custom-header.yaml` | **CREATE** | Simple test fixture |
| `test/uicustomheader/named-ui-custom-header.yaml` | **CREATE** | Named (spec.name override) test fixture |
| `test/loggerconfig/simple-logger-config.yaml` | **CREATE** | Simple test fixture |
| `test/loggerconfig/named-logger-config.yaml` | **CREATE** | Named (spec.name override) test fixture |
| `docs/system-backend/ui-headers.md` | **CREATE** | UICustomHeader documentation |
| `docs/system-backend/loggers.md` | **CREATE** | LoggerConfig documentation |
| `docs/system-backend/index.md` | **CREATE/MODIFY** | System backend docs index (if not already from 17.2/17.3) |
| `cmd/main.go` | **MODIFY** | Register controllers + webhooks |
| `config/crd/kustomization.yaml` | **MODIFY** | Add new CRD YAML resources |
| `go.mod` / `go.sum` | **NO CHANGE** | No new dependencies |
| `zz_generated.deepcopy.go` | **AUTO-GEN** | Via `make generate` |

- **Novelty Risk:** LOW — standard named-resource pattern. The only non-trivial element is the custom `IsEquivalentToDesiredState` for both types due to Vault read/write shape mismatches.
- Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings. _(Not applicable — neither type has duration or numeric fields. UICustomHeader has a string list, LoggerConfig has a string enum.)_

### Project Structure Notes

- Type files follow established pattern in `api/v1alpha1/` — co-located with all other CRD types
- Controllers follow `internal/controller/` convention
- Test fixtures in `test/<typename>/` — consistent with `test/passwordpolicy/`, `test/policy/`, etc.
- Documentation in `docs/system-backend/` — should already exist from Story 17.2 or 17.3 (first system-backend docs in Epic 17)
- No new dependencies, no shared framework changes, no reconciler variant changes

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 17, Story 17.4, lines 3023–3063]
- [Source: api/v1alpha1/passwordpolicy_types.go — named-resource pattern reference]
- [Source: api/v1alpha1/passwordpolicy_webhook.go — webhook pattern reference]
- [Source: internal/controller/passwordpolicy_controller.go — standard VaultResource controller pattern]
- [Source: api/v1alpha1/passwordpolicy_test.go — unit test pattern reference]
- [Source: internal/controller/passwordpolicy_controller_test.go — integration test pattern reference]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys, removeUnsetFields]
- [Source: api/v1alpha1/databasesecretengineconfig_types.go — toInterfaceArray helper (line 161)]
- [Source: api/v1alpha1/utils/vaultobject.go — VaultEndpoint.CreateOrUpdate and read() function]
- [Source: api/v1alpha1/utils/vaultutils.go — read() calls Logical().Read() without query params (line 48)]
- [Source: _bmad-output/project-context.md — All critical implementation rules]
- [Source: _bmad-output/implementation-artifacts/17-3-rotation-policy-crd.md — previous story pattern]
- [Vault API: /sys/config/ui/headers](https://developer.hashicorp.com/vault/api-docs/system/config-ui-headers)
- [Vault API: /sys/loggers](https://developer.hashicorp.com/vault/api-docs/system/loggers)

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

### Review Model Used

### Review Findings

### Decisions Needed / Decisions Taken

### Fixes Applied
