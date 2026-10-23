# Story 17.0: Vault API Observability — HTTP Metrics + Reconcile Trace Logging

Status: done

## Story

As an operator administrator,
I want Prometheus metrics tracking every Vault API call the operator makes, and high-verbosity trace logging for reconciliation decisions,
So that I can monitor operator-to-Vault traffic volume, detect runaway reconciliation loops, and troubleshoot reconciliation behavior without modifying the operator.

**Addresses:** [GitHub Issue #193](https://github.com/redhat-cop/vault-config-operator/issues/193), [Related PR #194](https://github.com/redhat-cop/vault-config-operator/pull/194) (stale, never merged)

## Acceptance Criteria

### Part A — Vault API HTTP Metrics

1. **Given** the operator is running and making Vault API calls **When** a Prometheus scrape hits the `/metrics` endpoint **Then** a `vault_api_requests_total` counter is present with labels `vault_address`, `method`, `path`, `status_code`
2. **Given** the operator is making Vault API calls **When** a Prometheus scrape hits the `/metrics` endpoint **Then** a `vault_api_request_duration_seconds` histogram is present with labels `vault_address`, `method`, `path`
3. **Given** the env var `VAULT_METRICS_NORMALIZE_PATHS` is unset or `false` **When** metrics are recorded **Then** the `path` label contains the raw Vault API path with only the `v1/` prefix stripped (e.g., `sys/policies/acl/my-policy`, `auth/kubernetes/role/my-role`)
4. **Given** the env var `VAULT_METRICS_NORMALIZE_PATHS` is `true` **When** metrics are recorded **Then** the `path` label contains a normalized path with dynamic segments collapsed to `*` (e.g., `sys/policies/acl/*`, `auth/*/role/*`)
5. **Given** CRs targeting different Vault instances (e.g., `https://vault-a:8200`, `https://vault-b:8200`) **When** metrics are recorded **Then** each Vault instance's traffic is distinguishable via the `vault_address` label (value is `host:port` from the request URL)

### Part B — High-Verbosity Reconcile Trace Logging

6. **Given** the operator runs with `-v=6` (or `--zap-log-level=6`) **When** a reconciliation cycle executes **Then** the following are logged: read result (found/not-found), drift decision (in-sync/out-of-sync), action taken (create/update/skip), and the Vault path involved
7. **Given** the operator runs at default verbosity **When** a reconciliation cycle executes **Then** no additional log output is produced (zero noise)

## Tasks / Subtasks

- [ ] Task 1 — Create `instrumentedRoundTripper` and metrics registration (AC: #1, #2, #3, #4, #5)
  - [ ] 1.1 Create `api/v1alpha1/utils/metrics.go` with `instrumentedRoundTripper` struct implementing `http.RoundTripper`
  - [ ] 1.2 Define `vault_api_requests_total` CounterVec with labels `vault_address`, `method`, `path`, `status_code`
  - [ ] 1.3 Define `vault_api_request_duration_seconds` HistogramVec with labels `vault_address`, `method`, `path`
  - [ ] 1.4 Register both metrics with `sigs.k8s.io/controller-runtime/pkg/metrics` → `metrics.Registry`
  - [ ] 1.5 Implement `normalizePath()` function to collapse dynamic path segments
  - [ ] 1.6 Implement `cleanPath()` function that strips `v1/` prefix and normalizes slashes (used when normalization is off)
  - [ ] 1.7 Read `VAULT_METRICS_NORMALIZE_PATHS` env var at init time to select path transformation
  - [ ] 1.8 Promote `prometheus/client_golang` from indirect to direct dependency in `go.mod`
- [ ] Task 2 — Inject instrumented transport in Vault client creation (AC: #1, #2)
  - [ ] 2.1 Wrap `config.HttpClient.Transport` in `getConnectionConfig()` (`api/v1alpha1/utils/commons.go`, after TLS config, before return)
  - [ ] 2.2 Wrap `config.HttpClient.Transport` in `createVaultClient()` when using `vault.DefaultConfig()` fallback path
- [ ] Task 3 — Add V(6) trace logging to reconcile endpoints (AC: #4, #5)
  - [ ] 3.1 Add V(6) log lines in `VaultEndpoint.CreateOrUpdate()` (`api/v1alpha1/utils/vaultobject.go`)
  - [ ] 3.2 Add V(6) log lines in `VaultEndpoint.DeleteIfExists()` (`api/v1alpha1/utils/vaultobject.go`)
  - [ ] 3.3 Add V(6) log lines in `VaultAuditEndpoint.CreateOrUpdate()` (`api/v1alpha1/utils/vaultauditobject.go`)
  - [ ] 3.4 Add V(6) log lines in `VaultEngineEndpoint.CreateOrUpdateTuneConfig()` (`api/v1alpha1/utils/vaultengineobject.go`)
  - [ ] 3.5 Add V(6) log lines in `ReconcileWithFunctions()` (`internal/controller/vaultresourcecontroller/reconcile_skeleton.go`)
- [ ] Task 4 — Unit tests (AC: #1, #2, #3, #4)
  - [ ] 4.1 Table-driven tests for `normalizePath()` covering all documented path patterns
  - [ ] 4.2 Table-driven tests for `cleanPath()` (v1/ stripping, slash normalization)
  - [ ] 4.3 Unit test for `instrumentedRoundTripper` verifying metrics are recorded on a mocked round trip (both normalized and raw modes)
- [ ] Task 5 — Validation and regression
  - [ ] 5.1 `make manifests generate fmt vet test` passes
  - [ ] 5.2 `make lint` passes (errcheck on new code)
  - [ ] 5.3 Verify no existing tests regress

## Dev Notes

### Part A — Vault API HTTP Metrics

#### Architecture

Create a single new file `api/v1alpha1/utils/metrics.go`. This file contains:
- The `instrumentedRoundTripper` struct wrapping `http.RoundTripper`
- The two Prometheus metric collectors (CounterVec, HistogramVec)
- The `normalizePath()` function
- An `init()` or package-level `var` block that registers metrics

The instrumented transport intercepts **every** HTTP call from the Vault client — reconciler reads/writes/deletes, auth logins, token renewals, sys calls — with zero changes to any controller or type file.

#### `instrumentedRoundTripper` Implementation

```go
package utils

import (
    "net/http"
    "strconv"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
    vaultAPIRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "vault_api_requests_total",
            Help: "Total number of Vault API requests by Vault instance, method, normalized path, and HTTP status code.",
        },
        []string{"vault_address", "method", "path", "status_code"},
    )
    vaultAPIRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "vault_api_request_duration_seconds",
            Help:    "Duration of Vault API requests in seconds by Vault instance, method, and normalized path.",
            Buckets: prometheus.DefBuckets,
        },
        []string{"vault_address", "method", "path"},
    )
)

func init() {
    crmetrics.Registry.MustRegister(vaultAPIRequestsTotal, vaultAPIRequestDuration)
}

type instrumentedRoundTripper struct {
    inner http.RoundTripper
}

func newInstrumentedRoundTripper(rt http.RoundTripper) http.RoundTripper {
    if rt == nil {
        rt = http.DefaultTransport
    }
    return &instrumentedRoundTripper{inner: rt}
}

// normalizePaths is read once at init from VAULT_METRICS_NORMALIZE_PATHS.
// Default: false (raw paths — better debuggability for bounded CR counts).
var normalizePaths bool

func init() {
    if v, ok := os.LookupEnv("VAULT_METRICS_NORMALIZE_PATHS"); ok && v == "true" {
        normalizePaths = true
    }
}

// metricsPath returns the path label value — either raw (v1/ stripped) or normalized.
func metricsPath(rawPath string) string {
    if normalizePaths {
        return normalizePath(rawPath)
    }
    return cleanPath(rawPath)
}

func (irt *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    vaultAddr := req.URL.Host
    path := metricsPath(req.URL.Path)
    start := time.Now()

    resp, err := irt.inner.RoundTrip(req)

    duration := time.Since(start).Seconds()
    statusCode := 0
    if resp != nil {
        statusCode = resp.StatusCode
    }
    vaultAPIRequestDuration.WithLabelValues(vaultAddr, req.Method, path).Observe(duration)
    vaultAPIRequestsTotal.WithLabelValues(vaultAddr, req.Method, path, strconv.Itoa(statusCode)).Inc()

    return resp, err
}
```

#### `vault_address` Label — Multi-Instance Discrimination

The operator can manage CRs targeting different Vault servers — each CR's `spec.connection.address` may point to a different Vault instance. The `vault_address` label uses `req.URL.Host` (which is `host:port`, e.g., `vault.example.com:8200`) extracted directly from the outgoing HTTP request. This:
- Requires zero state threading — the round tripper reads it from the request itself
- Naturally groups all traffic per Vault instance regardless of which controller or CRD triggered the call
- Keeps cardinality low — typically 1–3 Vault instances per operator deployment

**Cardinality note:** With `vault_address` × `method` (2–3 values) × `path` (~30 normalized) × `status_code` (~5), the expected cardinality is ~450–1350 series per Vault instance. Well within safe Prometheus bounds.

#### Critical Detail — Double-Wrap Prevention

`getConnectionConfig()` and `createVaultClient()` may both be called for the same Vault client creation flow. The transport must be wrapped **exactly once**. Two safe approaches:
1. **Type-guard:** Before wrapping, check `if _, already := transport.(*instrumentedRoundTripper); !already { ... }`.
2. **Single injection point:** Wrap only in `getConnectionConfig()` and ensure the `vault.DefaultConfig()` fallback in `createVaultClient()` also wraps. Since `createVaultClient()` calls `getConnectionConfig()` when `vaultConnection != nil`, the `DefaultConfig()` fallback is the only additional injection point.

**Recommended approach: Type-guard.** It's defensive and idempotent — safe even if the call graph changes in the future.

#### Critical Detail — TLS Configuration Ordering

In `getConnectionConfig()`, the TLS configuration installs a custom client certificate callback on `config.HttpClient.Transport.(*http.Transport).TLSClientConfig`. The instrumented round tripper must wrap **after** all TLS configuration is complete so it delegates to the fully-configured transport. Insert the wrapping as the **last** operation before `return config, nil` on line 166:

```go
// Wrap transport for Vault API metrics — must be after TLS configuration
config.HttpClient.Transport = newInstrumentedRoundTripper(config.HttpClient.Transport)
return config, nil
```

In `createVaultClient()`, for the `vault.DefaultConfig()` fallback path (line 367), wrap immediately after `config` is created:

```go
config = vault.DefaultConfig()
config.HttpClient.Transport = newInstrumentedRoundTripper(config.HttpClient.Transport)
```

#### Path Transformation — Two Modes

**`cleanPath(rawPath string) string`** (default, `VAULT_METRICS_NORMALIZE_PATHS` unset or `false`):
- Strip leading `v1/` prefix
- Normalize double slashes, trim leading/trailing slashes
- Result: raw Vault API path, e.g., `sys/policies/acl/my-policy`, `auth/kubernetes/role/my-role`
- Best for: typical deployments with bounded CR counts where per-resource visibility matters

**`normalizePath(rawPath string) string`** (when `VAULT_METRICS_NORMALIZE_PATHS=true`):
- Calls `cleanPath()` first, then collapses dynamic segments to `*`
- Target ~30 distinct normalized path patterns covering 95%+ of traffic
- Best for: very large deployments (1000+ CRs) where cardinality could stress Prometheus

Key normalized patterns:

| Vault API Pattern | Normalized Form |
|---|---|
| `v1/sys/policies/acl/my-policy` | `sys/policies/acl/*` |
| `v1/sys/quotas/rate-limit/my-quota` | `sys/quotas/rate-limit/*` |
| `v1/auth/kubernetes/login` | `auth/*/login` |
| `v1/auth/kubernetes/role/my-role` | `auth/*/role/*` |
| `v1/auth/kubernetes/config` | `auth/*/config` |
| `v1/secret/data/my-secret` | `*/data/*` |
| `v1/database/config/my-db` | `*/config/*` |
| `v1/database/roles/my-role` | `*/roles/*` |
| `v1/sys/mounts/secret/tune` | `sys/mounts/*/tune` |
| `v1/sys/auth` | `sys/auth` |
| `v1/sys/audit/my-device` | `sys/audit/*` |

Implementation strategy:
1. Strip the leading `v1/` prefix (Vault API version prefix).
2. Strip any leading or trailing slashes, normalize `//` to `/`.
3. Split on `/` and apply pattern-specific collapsing rules.
4. For `sys/` prefix paths: collapse named resource segments (the segment after the endpoint type).
5. For `auth/` prefix paths: collapse the mount name (first segment after `auth/`), keep the operation (`login`, `role/*`, `config`).
6. For all other paths (secret engine paths): collapse the mount name (first segment), keep the operation type, collapse named resource identifiers.

**Critical constraint:** The function operates on `req.URL.Path` which includes the `v1/` prefix from Vault's HTTP API. Always strip it first.

#### Dependency Change

`prometheus/client_golang` is currently an indirect dependency at v1.24.0 (transitive via controller-runtime). Promote to direct:

```bash
go get github.com/prometheus/client_golang@v1.24.0
```

This only moves the line in `go.mod` from `// indirect` to direct — no version change needed. The `prometheus/client_model` and related packages remain indirect.

### Part B — Reconcile Trace Logging

#### Logging Pattern

All trace logging uses `log.V(6).Info(...)` — visible only with `-v=6` or higher. **Never log payload content** — payloads may contain secrets. Log only:
- The Vault path
- The operation (read/write/delete/skip)
- The decision (found/not-found, in-sync/out-of-sync)

#### Specific Injection Points

**`VaultEndpoint.CreateOrUpdate()`** — `api/v1alpha1/utils/vaultobject.go` starting at line 175:
```go
func (ve *VaultEndpoint) CreateOrUpdate(context context.Context) error {
    log := log.FromContext(context)
    currentPayload, found, err := read(context, ve.vaultObject.GetPath())
    if err != nil {
        log.Error(err, "unable to read object at", "path", ve.vaultObject.GetPath())
        return err
    }
    if !found {
        log.V(6).Info("vault object not found, creating", "path", ve.vaultObject.GetPath())
        return write(context, ve.vaultObject.GetPath(), ve.vaultObject.GetPayload())
    } else {
        if !ve.vaultObject.IsEquivalentToDesiredState(currentPayload) {
            log.V(6).Info("vault object out of sync, updating", "path", ve.vaultObject.GetPath())
            // ... existing write logic ...
        } else {
            log.V(6).Info("vault object in sync, skipping write", "path", ve.vaultObject.GetPath())
        }
    }
    return nil
}
```

**`VaultEndpoint.DeleteIfExists()`** — `api/v1alpha1/utils/vaultobject.go` around line 90:
- Log before the delete call: `log.V(6).Info("deleting vault object", "path", ve.vaultObject.GetPath())`
- Log 404 as no-op: `log.V(6).Info("vault object not found on delete, no-op", "path", ve.vaultObject.GetPath())`

**`VaultAuditEndpoint.CreateOrUpdate()`** — `api/v1alpha1/utils/vaultauditobject.go` around line 149:
- Log exists/not-exists check result
- Log equivalent/not-equivalent decision
- Log enable/disable+re-enable action

**`VaultEngineEndpoint.CreateOrUpdateTuneConfig()`** — `api/v1alpha1/utils/vaultengineobject.go` around line 83:
- Log tune read result
- Log in-sync/out-of-sync decision

**`ReconcileWithFunctions()`** — `internal/controller/vaultresourcecontroller/reconcile_skeleton.go` around line 57:
- Log deletion path entry: `log.V(6).Info("reconcile: processing deletion", "name", instance.GetName())`
- Log reconcile path entry: `log.V(6).Info("reconcile: processing reconciliation", "name", instance.GetName())`

### Shared Framework File Safety

This story **modifies behavior** in shared framework files. The changes are additive and non-breaking:
- **`vaultobject.go`**: Adding `log.V(6)` calls — no behavioral change at default verbosity
- **`vaultengineobject.go`**: Adding `log.V(6)` calls — same
- **`vaultauditobject.go`**: Adding `log.V(6)` calls — same
- **`reconcile_skeleton.go`**: Adding `log.V(6)` calls — same
- **`commons.go`**: Wrapping transport — transparent; the inner transport is unchanged, only wrapped with metrics recording

None of these changes alter existing return values, error handling, or control flow. The `instrumentedRoundTripper` is a transparent proxy — it passes requests through unchanged and only observes the response status code and timing.

### File Structure

| File | Action | Purpose |
|---|---|---|
| `api/v1alpha1/utils/metrics.go` | **CREATE** | `instrumentedRoundTripper`, metrics vars, `normalizePath()`, metric registration |
| `api/v1alpha1/utils/metrics_test.go` | **CREATE** | Unit tests for `normalizePath()`, `instrumentedRoundTripper` |
| `api/v1alpha1/utils/commons.go` | **MODIFY** | Inject transport wrapping in `getConnectionConfig()` and `createVaultClient()` |
| `api/v1alpha1/utils/vaultobject.go` | **MODIFY** | Add V(6) trace logging to `CreateOrUpdate()` and `DeleteIfExists()` |
| `api/v1alpha1/utils/vaultengineobject.go` | **MODIFY** | Add V(6) trace logging to `CreateOrUpdateTuneConfig()` |
| `api/v1alpha1/utils/vaultauditobject.go` | **MODIFY** | Add V(6) trace logging to `CreateOrUpdate()` and `DeleteIfExists()` |
| `internal/controller/vaultresourcecontroller/reconcile_skeleton.go` | **MODIFY** | Add V(6) trace logging to `ReconcileWithFunctions()` |
| `go.mod` | **MODIFY** | Promote `prometheus/client_golang` to direct dependency |

### Testing Strategy

**Unit tests** (`api/v1alpha1/utils/metrics_test.go`):
1. `TestNormalizePath` — table-driven with ~20 cases covering all documented patterns plus edge cases (empty path, single segment, trailing slash, `v1/` prefix, double slash)
2. `TestInstrumentedRoundTripper` — create a mock `http.RoundTripper`, wrap it, make a request, assert the counter and histogram are recorded. Use `prometheus.NewRegistry()` for test isolation (do NOT test against the global registry to avoid cross-test pollution).

**No integration test required for Part A** — the metrics endpoint is already exposed by controller-runtime at `:8443/metrics`. The instrumented transport captures everything automatically once injected. Verification in a real cluster is trivial: `curl -k https://localhost:8443/metrics | grep vault_api_`.

**No dedicated test for Part B** — V(6) logging is observational only. Verify manually by running with `-v=6`.

### Import Additions

For `metrics.go`:
```go
import (
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)
```

No new imports needed for trace logging files — they already import `log "sigs.k8s.io/controller-runtime/pkg/log"`.

- **Novelty Risk:** MEDIUM — touches shared framework (`commons.go`, `vaultobject.go`, `reconcile_skeleton.go`), but changes are additive (transport wrapping + V(6) log calls). No control flow or return value changes.
- Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings. _(Not directly applicable to this story — no new CRD types.)_

### Project Structure Notes

- All metrics infrastructure lives in `api/v1alpha1/utils/` alongside the Vault client code it instruments — consistent with the project's co-location pattern.
- Test file `metrics_test.go` in the same package enables white-box testing of `normalizePath()` (unexported function).
- No new controller files, webhook files, or CRD types — this story is purely infrastructure.

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 17, Story 17.0, lines 2823–2874]
- [Source: api/v1alpha1/utils/commons.go — `getConnectionConfig()` lines 134–166, `createVaultClient()` lines 354–394]
- [Source: api/v1alpha1/utils/vaultobject.go — `CreateOrUpdate()` lines 175–193, `DeleteIfExists()` lines 90–103]
- [Source: api/v1alpha1/utils/vaultengineobject.go — `CreateOrUpdateTuneConfig()` lines 83–95]
- [Source: api/v1alpha1/utils/vaultauditobject.go — `CreateOrUpdate()` lines 149–182, `DeleteIfExists()` lines 186–197]
- [Source: internal/controller/vaultresourcecontroller/reconcile_skeleton.go — `ReconcileWithFunctions()` lines 57–83]
- [Source: _bmad-output/project-context.md — Shared framework protection rule]
- [Source: go.mod — `prometheus/client_golang v1.24.0 // indirect`]
- [Source: cmd/main.go — metrics server options lines 106–113]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

### Review Model Used

- Quick lens: GPT 5.4

### Review Findings

Pass 1 (quick) — 2 findings: 0 high, 0 medium, 2 low, 0 false
- `[low]` `[reject]` normalizePath doesn't collapse 3+ segment subpath names for identity/oidc paths — real but rare; plan targets ~95% coverage
- `[low]` `[reject]` sysNamedRe doesn't cover sys/config/auditing/request-headers/<name> — real but rare; fix adds regex maintenance surface

### Decisions Needed / Decisions Taken

None — both low findings rejected as acceptable coverage gaps within plan's ~95% target.

### Fixes Applied

None — no patches required.
