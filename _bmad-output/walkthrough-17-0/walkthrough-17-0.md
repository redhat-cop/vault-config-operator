# Walkthrough: Story 17.0 — Vault API Observability

_Target: commit `03439fe` on `epic-17`_

- [ ] **Block 1: Intent** _(current)_
- [ ] **Block 2: Broad Strokes**
- [ ] **Block 3: Metrics Infrastructure** — `instrumentedRoundTripper`, metric definitions, path normalization
- [ ] **Block 4: Transport Injection** — wrapping the Vault HTTP clients in `commons.go`
- [ ] **Block 5: Reconcile Trace Logging** — V(6) log lines across four endpoint files + skeleton
- [ ] **Block 6: Tests** — unit tests for normalization and round tripper
- [ ] **Block 7: Periphery** — `go.mod` dependency promotion, plan file

---

## Block 1: Intent

From the [plan](../plan-17-0-vault-api-observability-metrics-and-reconcile-trace-logging.md) (`<frozen-after-approval>` section):

> **Problem:** The operator makes Vault API calls with no Prometheus metrics and no high-verbosity trace logging for reconciliation decisions, making it impossible to monitor operator-to-Vault traffic or troubleshoot reconciliation behavior without modifying the operator.
>
> **Approach:** Create an `instrumentedRoundTripper` that wraps every Vault HTTP client transport to emit `vault_api_requests_total` and `vault_api_request_duration_seconds` Prometheus metrics, and add `log.V(6)` trace lines in all reconcile decision points so they appear only at high verbosity.

This addresses [GitHub Issue #193](https://github.com/redhat-cop/vault-config-operator/issues/193). The change touches 8 source files (2 new, 6 modified) with +617 −45 lines, most of which are the new `metrics.go` and its tests.

---

## Block 2: Broad Strokes

The change has two parts wired through five entry points:

1. **[`metrics.go`](../../../api/v1alpha1/utils/metrics.go)** — the new `instrumentedRoundTripper` and Prometheus metric definitions. The single file that owns all metrics infrastructure.
2. **[`commons.go:165`](../../../api/v1alpha1/utils/commons.go)** — the injection point in `getConnectionConfig()`, wrapping after TLS setup.
3. **[`commons.go:369`](../../../api/v1alpha1/utils/commons.go)** — the second injection in `createVaultClient()` for the `DefaultConfig()` fallback.
4. **[`vaultobject.go:184`](../../../api/v1alpha1/utils/vaultobject.go)** — representative V(6) trace logging in `CreateOrUpdate()`, the most used reconcile path.
5. **[`reconcile_skeleton.go:61`](../../../internal/controller/vaultresourcecontroller/reconcile_skeleton.go)** — V(6) at the top-level reconcile dispatcher, distinguishing deletion vs. reconcile paths.

---

## Block 3: Metrics Infrastructure

The entire metrics subsystem lives in one new file. It defines an HTTP round tripper that wraps any `http.RoundTripper`, records Prometheus counter and histogram on every call, and provides two path-cleaning modes (raw with `v1/` stripped, or normalized with dynamic segments collapsed to `*`).

Key spots:

- Counter and histogram vars: [`metrics.go:33–49`](../../../api/v1alpha1/utils/metrics.go)
- `init()` — registers with controller-runtime and reads `VAULT_METRICS_NORMALIZE_PATHS`: [`metrics.go:52–55`](../../../api/v1alpha1/utils/metrics.go)
- `RoundTrip()` — the interception: [`metrics.go:95–111`](../../../api/v1alpha1/utils/metrics.go)
- Type-guard double-wrap prevention: [`metrics.go:69–71`](../../../api/v1alpha1/utils/metrics.go)
- `cleanPath()`: [`metrics.go:123–132`](../../../api/v1alpha1/utils/metrics.go)
- `normalizePath()` with regex patterns: [`metrics.go:148–190`](../../../api/v1alpha1/utils/metrics.go)
- `wrapConfigTransport()` helper: [`metrics.go:195–201`](../../../api/v1alpha1/utils/metrics.go)

---

## Block 4: Transport Injection

Two lines, inserted at the two places a Vault `*api.Config` is created:

- After TLS in `getConnectionConfig()`: [`commons.go:165`](../../../api/v1alpha1/utils/commons.go)
- In the `vault.DefaultConfig()` fallback of `createVaultClient()`: [`commons.go:369`](../../../api/v1alpha1/utils/commons.go)

Both call `wrapConfigTransport(config.HttpClient, config.Address)`. The type-guard in `newInstrumentedRoundTripper` makes this idempotent — if both paths fire for the same client, the second wrap is a no-op.

---

## Block 5: Reconcile Trace Logging

V(6) log lines added at every reconcile decision point. All follow the same pattern: `log.V(6).Info("<decision>", "path", ...)`. Zero output at default verbosity.

- **`vaultobject.go`** — [`CreateOrUpdate()`](../../../api/v1alpha1/utils/vaultobject.go): not-found/create (line 184), out-of-sync/update (line 187), in-sync/skip (line 195). [`DeleteIfExists()`](../../../api/v1alpha1/utils/vaultobject.go): before delete (line 92), 404 no-op (line 96).
- **`vaultengineobject.go`** — [`CreateOrUpdateTuneConfig()`](../../../api/v1alpha1/utils/vaultengineobject.go): out-of-sync (line 92), in-sync (line 96).
- **`vaultauditobject.go`** — [`CreateOrUpdate()`](../../../api/v1alpha1/utils/vaultauditobject.go): not found/enabling (line 159), exists/checking (line 163), out-of-sync/re-enabling (line 174), in-sync/skip (line 184). [`DeleteIfExists()`](../../../api/v1alpha1/utils/vaultauditobject.go): exists/disabling (line 193), not found/no-op (line 197).
- **`reconcile_skeleton.go`** — [`ReconcileWithFunctions()`](../../../internal/controller/vaultresourcecontroller/reconcile_skeleton.go): deletion path (line 61), reconcile path (line 78).

---

## Block 6: Tests

All tests in [`metrics_test.go`](../../../api/v1alpha1/utils/metrics_test.go), 44 cases total:

- `TestCleanPath` (10 cases) — v1/ stripping, slash normalization, edge cases: [lines 17–42](../../../api/v1alpha1/utils/metrics_test.go)
- `TestNormalizePath` (29 cases) — sys/, auth/, secret engine paths, passthrough cases: [lines 44–94](../../../api/v1alpha1/utils/metrics_test.go)
- `TestInstrumentedRoundTripper` — real HTTP call, verifies counter + histogram with correct labels: [lines 96–80](../../../api/v1alpha1/utils/metrics_test.go)
- `TestInstrumentedRoundTripperTransportError` — transport error records `status_code=0`: [lines 82–133](../../../api/v1alpha1/utils/metrics_test.go)
- `TestInstrumentedRoundTripperDoubleWrap` — type-guard idempotency: [lines 135–141](../../../api/v1alpha1/utils/metrics_test.go)

Tests use isolated `prometheus.NewRegistry()` — no global registry pollution.

---

## Block 7: Periphery

- **`go.mod`** — `prometheus/client_golang` promoted from indirect v1.24.0 to direct v1.25.0; transitive deps bumped (crypto, net, sys, text, tools, protobuf — all patch-level).
- **`go.sum`** — corresponding checksum updates.
- **Plan file** — [`plan-17-0-vault-api-observability-metrics-and-reconcile-trace-logging.md`](../plan-17-0-vault-api-observability-metrics-and-reconcile-trace-logging.md) added to `_bmad-output/`.
