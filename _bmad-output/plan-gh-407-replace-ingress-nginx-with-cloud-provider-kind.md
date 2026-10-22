---
title: 'Replace ingress-nginx with cloud-provider-kind'
type: 'chore'
ticket: ''
created: '2026-10-22'
status: 'done'
baseline_revision: '10b4545adefc7bdd265461783bcf52157a6aba6f'
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

**Problem:** The `make integration` and `make helmchart-test` targets use ingress-nginx with Kind `extraPortMappings` to expose Vault. This causes CI failures due to IPv6/IPv4 mismatches, Podman rootless port-forwarding races, and a 675-line vendored manifest. A recent port-forward workaround was added to `make integration` to bypass these issues, but that itself adds complexity.

**Approach:** Replace ingress-nginx with `cloud-provider-kind` (v0.12.0) from `kubernetes-sigs`. Change the Vault Helm service to `type: LoadBalancer`, let cloud-provider-kind assign a routable external IP, and use that IP as VAULT_ADDR for integration tests. Remove the port-forward workaround from `make integration`. Remove all ingress-nginx artifacts. Clean up Kind cluster config (no more extraPortMappings or ingress-ready label).

## Boundaries & Constraints

**Always:**
- Use `--enable-lb-tunnel` when starting cloud-provider-kind (handles Podman rootless).
- Keep the health-check retry loop in `suite_integration_test.go` (still useful for LB readiness).
- `helmchart-test` must remain functional — it connects via in-cluster DNS (`vault.vault.svc:8200`), unaffected by service type change.
- `start-cloud-provider` is only a prerequisite for `integration`, not `helmchart-test`.

**Never:**
- Do not introduce Gateway API resources — use plain `Service type: LoadBalancer`.
- Do not change how `helmchart-test` connects to Vault (it uses in-cluster DNS).

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Integration happy path | Kind cluster up, cloud-provider-kind running, Vault deployed as LoadBalancer | LB external IP assigned, tests run against `http://<LB_IP>:8200` | N/A |
| LB IP not assigned | cloud-provider-kind slow or failed | Wait loop retries up to 60×5s | Makefile exits with error after timeout |
| cloud-provider-kind already running | Second `make integration` invoked | `pgrep` detects process, skips start | N/A |
| helmchart-test without cloud-provider | LB service stays Pending | In-cluster DNS `vault.vault.svc:8200` still works | No error — Pending LB is cosmetic |
| Podman rootless | Podman manages container networking | `--enable-lb-tunnel` creates tunnels from host to LB IPs | Handled by cloud-provider-kind |

</frozen-after-approval>

## Code Map

- `Makefile` (lines 1–671) — Main build file. Contains `kind-setup` (L220–241), `deploy-ingress` (L204–209), `integration` (L171–201), `helmchart-test` (L569–601), port variables (L11–12, L19–25), target-specific overrides (L171–173, L569–571), tool download targets. The `integration` recipe (L176–201) has the port-forward workaround at L190–201.
- `integration/cluster-kind.yaml.tpl` — Kind cluster template with `ingress-ready=true` label and `extraPortMappings` for ports 80/443. Template variables `VAULT_HOST_PORT_PLACEHOLDER` and `HTTPS_HOST_PORT_PLACEHOLDER` are substituted by `kind-setup`.
- `integration/cluster-kind.yaml` — Generated from the template (gitignored or committed). Will be replaced by a static file since no port substitution is needed.
- `integration/vault-values.yaml` — Vault Helm values. Currently has no `service` section (defaults to `ClusterIP`). Needs `server.service.type: LoadBalancer`.
- `internal/controller/suite_integration_test.go` (lines 83–107) — Vault health check with retry loop. Comment references "Podman rootless port-forwarding" and "ingress" which need updating. The logic itself (retry with 5s timeout) is still appropriate.
- `integration/helm/ingress-nginx/` — 4 files: `Chart.yaml`, `.helmignore`, `templates/_helpers.tpl`, `templates/ingress-vault.yaml`. All to be deleted.
- `integration/manifests/ingress-nginx-kind-deploy.yaml` — 675-line vendored ingress-nginx controller manifest. To be deleted.

## Tasks & Acceptance

**Execution:**
- [ ] `integration/cluster-kind.yaml.tpl` — Remove `ingress-ready=true` kubelet label and `extraPortMappings` section. Reduce to a minimal Kind cluster config with just the control-plane node role. Since no template variables remain, rename to `integration/cluster-kind.yaml` (static file) and remove the `.tpl` variant.
- [ ] `integration/vault-values.yaml` — Add `server.service.type: LoadBalancer` under the existing `server:` block so cloud-provider-kind can assign an external IP to the Vault service.
- [ ] `Makefile` — (a) Add `CLOUD_PROVIDER_KIND_VERSION ?= v0.12.0` variable. (b) Add `cloud-provider-kind` download target using `go-install-tool`. (c) Add `start-cloud-provider` target: checks `pgrep -f cloud-provider-kind`, starts if not running with `--enable-lb-tunnel`, logs to `/tmp/cloud-provider-kind.log`. (d) Remove `VAULT_HOST_PORT`, `HTTPS_HOST_PORT`, `INT_VAULT_PORT`, `INT_HTTPS_PORT`, `INT_VAULT_PF_PORT`, `HELM_VAULT_PORT`, `HELM_HTTPS_PORT` variables. (e) Remove target-specific port overrides from `integration` and `helmchart-test`. (f) Simplify `kind-setup`: remove `sed` template substitution (use static `cluster-kind.yaml`), simplify cluster-exists check to only verify image (no port inspection), remove `CURRENT_PORT` / `CURRENT_HTTPS_PORT` checks. (g) Replace `deploy-ingress` in `integration` prerequisites with `start-cloud-provider`. (h) Rewrite `integration` recipe: remove port-forward workaround, add LB IP wait loop (up to 60 attempts × 5s), set VAULT_ADDR to `http://<LB_IP>:8200`. (i) Delete `deploy-ingress` target entirely. (j) Add `cloud-provider-kind` process cleanup to `kind-teardown`.
- [ ] `internal/controller/suite_integration_test.go` — Update comment on health-check retry loop to reference cloud-provider-kind LB readiness instead of ingress and port-forwarding. Keep the retry logic unchanged.
- [ ] Delete `integration/helm/ingress-nginx/` (entire directory) and `integration/manifests/ingress-nginx-kind-deploy.yaml`.
- [ ] Delete `integration/cluster-kind.yaml.tpl` after creating static `integration/cluster-kind.yaml`.

**Acceptance Criteria:**
- Given a clean environment with no Kind cluster, when `make integration` is run, then it creates a Kind cluster, starts cloud-provider-kind, deploys Vault with a LoadBalancer service, waits for the external IP, and runs integration tests successfully against that IP.
- Given `make helmchart-test` is run, then it creates a separate Kind cluster (without cloud-provider-kind), deploys Vault, and the operator connects to Vault via in-cluster DNS — tests pass.
- Given ingress-nginx files have been deleted, when `grep -r ingress-nginx integration/` is run, then no matches are found.
- Given the port-forward workaround has been removed, when `grep -r port-forward Makefile` is run, then no matches are found in the integration target.

## Implementation Notes

- `--enable-lb-tunnel` does NOT work with podman rootless (requires sudo). Switched to `--enable-lb-port-mapping` which creates random host port mappings for the LB container. The integration recipe discovers the mapped port via `$(CONTAINER_RUNTIME) port <container> 8200/tcp`.
- `pkill -f cloud-provider-kind` in kind-teardown was self-destructing the make process because `-f` matches the full command line (including the make recipe text). Switched to PID-file approach (`/tmp/cloud-provider-kind.pid`).
- Added startup health check to `start-cloud-provider`: verifies PID is still alive after 3s sleep, and dumps log on failure.
- The "Handling connection for 9200" messages from port-forward are now gone from the test output.


## Plan Change Log


## Review Triage Log

**Pass 1 (quick):** 1 low, 4 false

| # | Verdict | Route | Evidence |
|---|---------|-------|----------|
| 1 | low | patch | `kind-teardown` kills cloud-provider-kind via PID file even when called from `helmchart-teardown`. Move cleanup to `integration-teardown` only. |
| 2 | false | — | `--enable-lb-tunnel` requires sudo with podman rootless (empirically proven). `--enable-lb-port-mapping` is the correct adaptation. Frozen text was based on the issue's unverified claim. |
| 3 | false | — | Same root cause as #2. Port-mapped connection is the necessary consequence. |
| 4 | false | — | `pgrep -f` matched the make recipe's own command line, causing self-termination. PID file is functionally equivalent and avoids the bug. |
| 5 | false | — | Plan metadata is updated continuously during workflow. Staged diff reflects point-in-time snapshot; resolved at commit. |


## Design Notes

**Why cloud-provider-kind over other options:**
cloud-provider-kind is the official Kind-native cloud provider from `kubernetes-sigs`. It eliminates the proxy layer entirely: `Host → LB external IP → Vault pod`. The `--enable-lb-tunnel` flag handles Podman rootless networking. It adds a Go binary dependency (installed via `go install`, consistent with existing tooling) and a background process, but removes ~750 lines of ingress-nginx artifacts and all the port-mapping/port-forwarding complexity.

**Why keep the health-check retry loop:**
Even with a direct LoadBalancer IP, there is a brief window where the IP is assigned but Vault is not yet serving. The 12-attempt retry with 5s timeout remains a robust guard.

**Static vs template Kind config:**
With no port mappings to substitute, the Kind cluster config becomes static. Removing the `.tpl` → `.yaml` generation step simplifies `kind-setup` and eliminates a generated file from version control.

## Verification

**Commands:**
- `make integration` — expected: all integration tests pass using cloud-provider-kind LoadBalancer IP
- `make helmchart-test` — expected: all helmchart tests pass using in-cluster DNS
- `grep -r 'ingress-nginx' integration/ Makefile` — expected: no matches
- `grep -r 'port-forward' Makefile` — expected: no matches in integration target
- `ls integration/helm/ingress-nginx/ integration/manifests/ingress-nginx-kind-deploy.yaml 2>&1` — expected: "No such file or directory"
