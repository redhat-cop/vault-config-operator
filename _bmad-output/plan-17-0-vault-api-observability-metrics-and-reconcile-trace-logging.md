---
title: 'Vault API Observability — HTTP Metrics + Reconcile Trace Logging'
type: 'feature'
ticket: ''
created: '2026-10-20'
status: 'built'
baseline_revision: '8460dde87ff8c5269332fc1139cf0c902d43c4e3'
route: 'full'
route_source: 'auto'
risk: 'medium'
review: 'quick'
review_source: 'pinned'
lenses_ran: ['quick']
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** The operator makes Vault API calls with no Prometheus metrics and no high-verbosity trace logging for reconciliation decisions, making it impossible to monitor operator-to-Vault traffic or troubleshoot reconciliation behavior without modifying the operator.

**Approach:** Create an `instrumentedRoundTripper` that wraps every Vault HTTP client transport to emit `vault_api_requests_total` and `vault_api_request_duration_seconds` Prometheus metrics, and add `log.V(6)` trace lines in all reconcile decision points so they appear only at high verbosity.

## Boundaries & Constraints

**Always:** Register metrics with the controller-runtime registry (`sigs.k8s.io/controller-runtime/pkg/metrics`). Wrap transport after TLS configuration is complete. Use a type-guard to prevent double-wrapping. Never log payload content (may contain secrets) — log only paths, operations, and decisions. V(6) trace logging must produce zero output at default verbosity.

**Never:** Do not alter existing return values, error handling, or control flow in any modified file. Do not create new CRD types or controller files. Do not use a global Prometheus registry for tests.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Normal Vault call | Any Vault HTTP request | Counter incremented, histogram observed with `vault_address`, `method`, `path`, `status_code` labels | On transport error, metrics still recorded with `status_code=0` |
| Path normalization off (default) | `v1/sys/policies/acl/my-policy` | `path` label = `sys/policies/acl/my-policy` (v1/ stripped only) | N/A |
| Path normalization on | `v1/auth/kubernetes/role/my-role` with `VAULT_METRICS_NORMALIZE_PATHS=true` | `path` label = `auth/*/role/*` | N/A |
| Double-wrap prevention | Transport already wrapped | Same `instrumentedRoundTripper` returned, no nested wrapping | N/A |
| DefaultConfig fallback | `vaultConnection` is nil in `createVaultClient()` | `vault.DefaultConfig()` transport wrapped | N/A |
| V(6) trace — create | Object not found in Vault | Logs "vault object not found, creating" with path | N/A |
| V(6) trace — update | Object out of sync | Logs "vault object out of sync, updating" with path | N/A |
| V(6) trace — skip | Object in sync | Logs "vault object in sync, skipping write" with path | N/A |
| V(6) at default verbosity | Normal reconciliation at default log level | Zero additional log output | N/A |

</frozen-after-approval>

## Code Map

- `api/v1alpha1/utils/commons.go` -- Vault client creation; `getConnectionConfig()` (line 134) returns config after TLS at line 165; `createVaultClient()` (line 354) has `vault.DefaultConfig()` fallback at line 367
- `api/v1alpha1/utils/vaultobject.go` -- `CreateOrUpdate()` (line 174): read→found/not-found→equivalent check→write/skip; `DeleteIfExists()` (line 89): delete with 404 handling
- `api/v1alpha1/utils/vaultengineobject.go` -- `CreateOrUpdateTuneConfig()` (line 83): read tune→equivalent check→write
- `api/v1alpha1/utils/vaultauditobject.go` -- `CreateOrUpdate()` (line 149): exists check→enable or equivalent check→disable+re-enable; `DeleteIfExists()` (line 186): exists→disable
- `internal/controller/vaultresourcecontroller/reconcile_skeleton.go` -- `ReconcileWithFunctions()` (line 57): deletion vs reconcile path branching
- `go.mod` line 91 -- `prometheus/client_golang v1.24.0 // indirect` to promote to direct

## Tasks & Acceptance

**Execution:**
- [ ] `api/v1alpha1/utils/metrics.go` -- CREATE -- `instrumentedRoundTripper` struct, `vault_api_requests_total` CounterVec, `vault_api_request_duration_seconds` HistogramVec, `normalizePath()`, `cleanPath()`, `metricsPath()`, `newInstrumentedRoundTripper()` with type-guard, `init()` for metric registration and `VAULT_METRICS_NORMALIZE_PATHS` env read
- [ ] `api/v1alpha1/utils/metrics_test.go` -- CREATE -- Table-driven tests for `normalizePath()` (~20 cases), `cleanPath()`, and `instrumentedRoundTripper` using isolated `prometheus.NewRegistry()`
- [ ] `api/v1alpha1/utils/commons.go` -- MODIFY -- Wrap transport in `getConnectionConfig()` as last operation before return (line 165), and in `createVaultClient()` after `vault.DefaultConfig()` fallback (line 367). Add `metrics.go` is same package so no import needed.
- [ ] `api/v1alpha1/utils/vaultobject.go` -- MODIFY -- Add `log.V(6).Info()` in `CreateOrUpdate()` for not-found/create, out-of-sync/update, in-sync/skip; and in `DeleteIfExists()` before delete and on 404 no-op
- [ ] `api/v1alpha1/utils/vaultengineobject.go` -- MODIFY -- Add `log.V(6).Info()` in `CreateOrUpdateTuneConfig()` for in-sync/out-of-sync decision
- [ ] `api/v1alpha1/utils/vaultauditobject.go` -- MODIFY -- Add `log.V(6).Info()` in `CreateOrUpdate()` for exists/not-exists, equivalent/not-equivalent, enable/re-enable; and in `DeleteIfExists()` for exists/not-exists
- [ ] `internal/controller/vaultresourcecontroller/reconcile_skeleton.go` -- MODIFY -- Add `log.V(6).Info()` in `ReconcileWithFunctions()` for deletion path and reconcile path
- [ ] `go.mod` -- MODIFY -- Promote `prometheus/client_golang` from indirect to direct via `go get`

**Acceptance Criteria:**
- Given the operator is running and making Vault API calls, when Prometheus scrapes `/metrics`, then `vault_api_requests_total` and `vault_api_request_duration_seconds` are present with labels `vault_address`, `method`, `path`, `status_code` (counter) / without `status_code` (histogram)
- Given `VAULT_METRICS_NORMALIZE_PATHS` is unset, when metrics are recorded, then `path` has only `v1/` prefix stripped
- Given `VAULT_METRICS_NORMALIZE_PATHS=true`, when metrics are recorded, then `path` has dynamic segments collapsed to `*`
- Given the operator runs at default verbosity, when reconciliation executes, then no additional log output is produced
- Given the operator runs with `-v=6`, when reconciliation executes, then read result, drift decision, and action taken are logged with the Vault path
- Given `make manifests generate fmt vet test` is run, then all commands pass with no regressions
- Given `make lint` is run, then no new lint violations are introduced

## Implementation Notes


## Plan Change Log


## Review Triage Log

Pass 1 (quick, GPT 5.4): 0 high, 0 medium, 2 low, 0 false, 0 maybe-false

| # | Verdict | Route | Evidence |
|---|---------|-------|----------|
| 1 | low (rejected) | — | `normalizePath` secret engine branch leaves 3+ segment subpath names uncollapsed (e.g. `identity/oidc/scope/<name>` → `*/oidc/*/name`). Real but rare path; fix requires restructuring heuristics. Plan targets ~95% coverage. |
| 2 | low (rejected) | — | `sysNamedRe` doesn't cover `sys/config/auditing/request-headers/<name>`. Real but rare path; fix adds regex maintenance surface. Plan targets ~95% coverage. |


## Design Notes

**Double-wrap type-guard:** `newInstrumentedRoundTripper` checks `if _, already := rt.(*instrumentedRoundTripper); already { return rt }` before wrapping — idempotent regardless of call graph changes.

**TLS ordering:** In `getConnectionConfig()`, the wrap must come after `applyConnectionTLS()` so the inner transport has its `TLSClientConfig.GetClientCertificate` callback installed.

**Path normalization strategy:** `cleanPath` strips `v1/` and normalizes slashes. `normalizePath` additionally collapses dynamic segments: `sys/` paths collapse named resources, `auth/` paths collapse mount names and resource names, all others collapse mount name and resource identifiers. Target ~30 distinct patterns.

**Test isolation:** Use `prometheus.NewRegistry()` per test for `instrumentedRoundTripper` tests — never the global registry.

## Verification

**Commands:**
- `make manifests generate fmt vet test` -- expected: all pass, no regressions
- `make lint` -- expected: clean, no new violations
- `go test ./api/v1alpha1/utils/ -run TestNormalize -v` -- expected: all path normalization cases pass
- `go test ./api/v1alpha1/utils/ -run TestInstrumented -v` -- expected: round tripper metrics recorded correctly
