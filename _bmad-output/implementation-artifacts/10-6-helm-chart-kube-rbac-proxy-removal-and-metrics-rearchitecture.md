# Story 10.6: Helm Chart kube-rbac-proxy Removal and Metrics Rearchitecture

Status: ready-for-dev

## Story

As an operator developer,
I want to update the Helm chart to remove the kube-rbac-proxy sidecar and align the metrics architecture with the controller-runtime authn/authz approach,
So that the Helm-based deployment matches the kustomize-based deployment and no longer references kube-rbac-proxy.

## Acceptance Criteria

1. **Given** `values.yaml.tpl` contains a `kube_rbac_proxy` configuration section
   **When** the section is removed and replaced with `metricsSecure: true`
   **Then** `helm template` no longer references kube-rbac-proxy image

2. **Given** `templates/manager.yaml` injects a kube-rbac-proxy sidecar when `enableMonitoring` is true
   **When** the sidecar block is removed and the manager container is updated to serve metrics directly on `:8443`
   **Then** `helm template` produces a Deployment with only the manager container
   **And** the manager container includes `--metrics-secure` and `--metrics-bind-address=:8443` args

3. **Given** the `enableMonitoring` flag controls metrics exposure
   **When** `enableMonitoring` is true
   **Then** the metrics port (8443) is exposed on the manager container directly (no proxy)

4. **Given** all Helm chart changes are applied
   **When** `helm template` and `helm lint` are run
   **Then** both succeed without errors or warnings

## Tasks / Subtasks

- [ ] Task 1: Remove `kube_rbac_proxy` section from `values.yaml.tpl` (AC: #1)
  - [ ] Delete the entire `kube_rbac_proxy:` block (lines 42-58 in current file)
  - [ ] Add `metricsSecure: true` below `enableMonitoring: true`

- [ ] Task 2: Remove kube-rbac-proxy sidecar from `templates/manager.yaml` (AC: #2, #3)
  - [ ] Delete the conditional kube-rbac-proxy sidecar container block (lines 31-52 in current file)
  - [ ] Add conditional metrics args to the manager container when `enableMonitoring` is true:
    - `--metrics-bind-address=:8443`
    - `--metrics-secure` (conditional on `.Values.metricsSecure`)
  - [ ] Add conditional metrics port (`containerPort: 8443, name: https`) to manager container when `enableMonitoring` is true

- [ ] Task 3: Update volume section in `templates/manager.yaml` (AC: #2)
  - [ ] Remove the `metrics-service-cert` volume block (lines 103-112 in current file) — manager uses controller-runtime auto-generated self-signed certs; no external cert mount needed

- [ ] Task 4: Verify with helm template and lint (AC: #4)
  - [ ] Run `helm template` on the chart directory and inspect output:
    - Only one container (manager) in the Deployment
    - Manager has `--metrics-bind-address=:8443` and `--metrics-secure` args
    - Port 8443 exposed on manager container
    - No references to `kube-rbac-proxy` or `kube_rbac_proxy`
  - [ ] Run `helm lint` on the chart directory — no errors or warnings
  - [ ] Run `make helmchart` to verify full chart generation pipeline

## Dev Notes

### Prerequisite: Story 10.5 MUST Be Completed First

Story 10.5 modifies `main.go` to:
- Bind the `--metrics-secure` flag to `secureMetrics` (default: `true`)
- Change the default `--metrics-bind-address` to `:8443`
- Add `FilterProvider: filters.WithAuthenticationAndAuthorization` for metrics authn/authz
- Remove kube-rbac-proxy from the kustomize-based deployment

**Critical timing:** After Story 10.5, the manager binary's default metrics bind address changes from `:8080` to `:8443`. The Helm chart's `manager.yaml` template currently does NOT pass `--metrics-bind-address` to the manager, so the manager uses the binary's compiled default. This means:

- **Before 10.5:** Manager defaults to `:8080`, kube-rbac-proxy listens on `:8443` → no port conflict
- **After 10.5 without 10.6:** Manager defaults to `:8443` (new binary default), kube-rbac-proxy ALSO on `:8443` → **PORT CONFLICT, pod crashes**

**Story 10.6 MUST be applied at the same time as Story 10.5** to prevent Helm chart deployments from breaking. If there is any delay between the two, add `--metrics-bind-address=127.0.0.1:8080` to the Helm chart manager args as a temporary workaround.

### Scope and Dependencies

This story modifies Helm chart templates and values only. It does NOT cover:

| Change | Story |
|--------|-------|
| go/v3 → go/v4 layout migration | Story 10.0 |
| Kustomize v3 → v5 syntax migration | Story 10.0a |
| Operator SDK v1.31 → v1.42 | Story 10.1 |
| Helm CLI v3 → v4 | Story 10.2 |
| golangci-lint v1 → v2 | Story 10.3 |
| OPM + kustomize tool versions | Story 10.4 |
| kube-rbac-proxy removal from kustomize + main.go changes | Story 10.5 |
| CI pipeline end-to-end validation | Story 10.7 |

### Current State of Files

**`config/helmchart/values.yaml.tpl`** (62 lines):
- Lines 42-58: `kube_rbac_proxy:` section with image (`quay.io/redhat-cop/kube-rbac-proxy:v0.11.0`), resources, securityContext
- Line 60: `enableMonitoring: true`
- Line 61: `enableCertManager: false`

**`config/helmchart/templates/manager.yaml`** (124 lines):
- Lines 31-52: Conditional kube-rbac-proxy sidecar container (`{{- if .Values.enableMonitoring }}`)
  - Listens on `0.0.0.0:8443`, forwards to `127.0.0.1:8080`
  - Uses `{{ .Values.kube_rbac_proxy.image.repository }}:{{ .Values.kube_rbac_proxy.image.tag }}`
  - Mounts `metrics-service-cert` volume at `/etc/certs/tls`
- Lines 53-89: Manager container
  - Command: `/manager` with `--leader-elect` + user-provided `.Values.args`
  - No `--metrics-bind-address` or `--metrics-secure` args
  - No metrics port exposed
  - Volume mounts: `webhook-service-cert` + user-provided `.Values.volumeMounts`
- Lines 102-112: Conditional `metrics-service-cert` volume (`{{- if .Values.enableMonitoring }}`)
  - When `enableCertManager: true` → secret `vault-config-operator-metrics-service-cert`
  - When `enableCertManager: false` → secret `vault-config-operator-certs`

**`config/helmchart/templates/certificate.yaml`** (34 lines):
- Conditional on `enableCertManager` — creates selfsigned Issuer + two Certificates (webhook + metrics)
- The metrics Certificate (lines 22-33) creates secret `vault-config-operator-metrics-service-cert`
- **No changes needed** — the metrics Certificate remains for future CertDir support

### Files to Change (Complete List)

| File | Change |
|------|--------|
| `config/helmchart/values.yaml.tpl` | Remove `kube_rbac_proxy:` section (17 lines), add `metricsSecure: true` |
| `config/helmchart/templates/manager.yaml` | Remove sidecar block, add metrics args/port to manager, remove `metrics-service-cert` volume |

### Files NOT to Change

- `config/helmchart/templates/certificate.yaml` — metrics Certificate stays for future CertDir support
- `config/helmchart/templates/_helpers.tpl` — no metrics-related content
- `config/helmchart/Chart.yaml.tpl` — no changes needed (chart version bumped by release process)
- `config/helmchart/kustomization.yaml` — no changes needed (kustomize overlay, not Helm templates)
- `config/helmchart/.helmignore` — no changes needed
- `main.go` or any Go source — those changes are Story 10.5
- `config/default/` or `config/rbac/` — kustomize changes are Story 10.5
- `Makefile` — no Helm chart generation changes needed

### Target State: `values.yaml.tpl`

```yaml
# ... (lines 1-41 unchanged) ...

enableMonitoring: true
metricsSecure: true
enableCertManager: false
```

The `kube_rbac_proxy:` section (lines 42-58) is removed entirely.

### Target State: `templates/manager.yaml` Manager Container Args

```yaml
      - command:
        - /manager
        args:
        - --leader-elect
        {{- if .Values.enableMonitoring }}
        - --metrics-bind-address=:8443
        {{- if .Values.metricsSecure }}
        - --metrics-secure
        {{- end }}
        {{- end }}
        {{- with .Values.args }}
          {{- toYaml . | nindent 8 }}
        {{- end }}
```

### Target State: `templates/manager.yaml` Manager Container Ports

Add a `ports:` block to the manager container (after the `name:` field, before `securityContext:`):

```yaml
        {{- if .Values.enableMonitoring }}
        ports:
        - containerPort: 8443
          name: https
          protocol: TCP
        {{- end }}
```

### Target State: `templates/manager.yaml` Volumes Section

```yaml
      volumes:
      - name: webhook-service-cert
        secret:
          {{- if .Values.enableCertManager }}
          secretName: vault-config-operator-webhook-service-cert
          {{- else }}
          secretName: webhook-server-cert
          {{- end }}
          defaultMode: 420
      {{- with .Values.volumes }}
        {{- toYaml . | nindent 6 }}
      {{- end }}
```

The `metrics-service-cert` volume block is removed. After Story 10.5, the manager uses controller-runtime's auto-generated self-signed certificates for the metrics endpoint — no external cert mount is needed.

### `enableMonitoring` Flag Semantics Change

| Aspect | Before (kube-rbac-proxy) | After (controller-runtime) |
|--------|--------------------------|---------------------------|
| What it controls | Whether to inject kube-rbac-proxy sidecar | Whether to add metrics args and port to manager |
| When `true` | Sidecar on :8443 forwards to manager on 127.0.0.1:8080 | Manager serves metrics directly on :8443 with TLS |
| When `false` | No sidecar, manager on :8080 (no TLS, not externally reachable) | No explicit metrics args — manager uses binary defaults (:8443 with TLS after 10.5) |
| Container count | 2 (manager + proxy) | 1 (manager only) |
| TLS termination | kube-rbac-proxy handles TLS | controller-runtime auto-generates self-signed certs |
| Authn/authz | kube-rbac-proxy performs SubjectAccessReview | controller-runtime `WithAuthenticationAndAuthorization` filter |

### TLS Certificate Strategy

After Story 10.5, the manager's metrics endpoint uses controller-runtime's auto-generated self-signed certificates. This means:

1. **No external cert volume needed** — controller-runtime generates certs at startup
2. **Prometheus ServiceMonitor** may need `insecureSkipVerify: true` in tlsConfig to scrape self-signed endpoints (this is handled in the kustomize config, not the Helm chart)
3. **CertManager metrics Certificate** (`certificate.yaml`) is preserved but currently unused — it creates the secret `vault-config-operator-metrics-service-cert` when `enableCertManager: true`, providing a foundation for future CertDir integration
4. Future enhancement: when `main.go` adds `--metrics-cert-dir` support, the Helm chart can re-add the volume mount and wire up CertManager-provided certs

### Breaking Changes for Existing Helm Installations

This is a **breaking change** for users upgrading from a previous chart version:

1. **Removed value:** `kube_rbac_proxy` — any custom overrides to `kube_rbac_proxy.image`, `kube_rbac_proxy.resources`, or `kube_rbac_proxy.securityContext` will be ignored. Helm `upgrade` will warn about unknown values if `--strict` is used.

2. **New value:** `metricsSecure: true` — controls whether `--metrics-secure` arg is passed. Users who want insecure metrics (not recommended) must set `metricsSecure: false`.

3. **Removed volume:** `metrics-service-cert` — deployments that relied on OpenShift service CA annotation (`vault-config-operator-certs` secret) for metrics TLS now use auto-generated self-signed certs. The secret may still exist but is no longer mounted.

4. **Port ownership change:** Port 8443 was previously owned by the kube-rbac-proxy sidecar. Now it's owned by the manager container directly. Existing NetworkPolicies targeting the proxy container by name will need updating.

**Document these in chart release notes.**

### Helm Template Output Verification

After implementation, `helm template test-release config/helmchart/` should produce a Deployment where:

1. **Only one container** named `vault-config-operator` (no `kube-rbac-proxy`)
2. Manager container args include `--metrics-bind-address=:8443` and `--metrics-secure`
3. Manager container has `ports:` section with `containerPort: 8443`
4. **No volume** named `metrics-service-cert`
5. **No references** to `kube_rbac_proxy`, `kube-rbac-proxy`, or `quay.io/redhat-cop/kube-rbac-proxy`
6. Webhook service cert volume and mount still present

### Anti-Patterns to Avoid

- **DO NOT** modify `main.go` or any Go source — that's Story 10.5
- **DO NOT** modify kustomize manifests (`config/default/`, `config/rbac/`, `config/prometheus/`) — that's Story 10.5
- **DO NOT** delete `config/helmchart/templates/certificate.yaml` or the metrics Certificate — it's useful for future CertDir support
- **DO NOT** change the Helm CLI version or Makefile targets — that's Story 10.2
- **DO NOT** change `config/helmchart/kustomization.yaml` — it's a kustomize overlay, not a Helm template
- **DO NOT** hardcode the kube-rbac-proxy image anywhere as a "fallback" — the sidecar approach is fully replaced
- **DO NOT** add a `--metrics-cert-dir` flag to the manager args — the binary doesn't support it until future main.go changes
- **DO NOT** keep the `metrics-service-cert` volume "just in case" — it would require a secret to exist, causing pod startup failures on vanilla Kubernetes clusters without cert provisioning
- **DO NOT** modify the manager's `--leader-elect` arg or the webhook cert volume mount — those are unrelated to this story

### Verification Commands (Run After Completion)

```bash
# Generate the Helm chart from kustomize
make helmchart

# Template the chart (requires Helm binary — make helm first)
helm template test-release config/helmchart/

# Lint the chart
helm lint config/helmchart/

# Verify no kube-rbac-proxy references remain
helm template test-release config/helmchart/ | grep -i "kube.rbac.proxy" && echo "FAIL: kube-rbac-proxy still referenced" || echo "PASS: no kube-rbac-proxy references"

# Verify manager has metrics args
helm template test-release config/helmchart/ | grep "metrics-bind-address" && echo "PASS: metrics-bind-address present" || echo "FAIL: missing metrics-bind-address"

# Verify single container (no sidecar)
helm template test-release config/helmchart/ | grep "name: kube-rbac-proxy" && echo "FAIL: sidecar still present" || echo "PASS: no sidecar"
```

### Rollback Strategy

If this story needs to be reverted:
1. Restore `kube_rbac_proxy:` section in `values.yaml.tpl`
2. Restore kube-rbac-proxy sidecar block in `templates/manager.yaml`
3. Restore `metrics-service-cert` volume block
4. Remove `metricsSecure` from `values.yaml.tpl`
5. Remove metrics args and port from manager container

**Important:** If Story 10.5 (main.go changes) is NOT reverted alongside this story, the Helm chart will have a port conflict (both kube-rbac-proxy and manager on `:8443`). Always revert 10.5 and 10.6 together.

### Previous Story Intelligence

From Story 10.2 (Helm upgrade):
- Anti-pattern: "DO NOT touch the Helm chart templates (`config/helmchart/`) — chart content changes are Story 10.6" — confirms this story's scope
- The Helm chart uses `apiVersion: v2` format which works with both Helm 3 and 4
- cert-manager was upgraded from v1.7.1 to v1.21.1 with `crds.enabled=true` instead of `installCRDs=true`

From Story 10.0a (kustomize migration):
- Anti-pattern: "DO NOT modify Helm chart templates or values — only the kustomize source files" — confirms clean separation
- The `config/helmchart/kustomization.yaml` uses `bases:` (deprecated) — this will be migrated to `resources:` in Story 10.0a if not already done

From Story 10.0 (go/v4 layout):
- After this story, the import path for controllers changes to `internal/controller/` — irrelevant to Helm chart templates

### Project Structure Notes

- All changes confined to `config/helmchart/` directory — 2 files modified, 0 files created, 0 files deleted
- The Helm chart generation pipeline (`make helmchart`) remains unchanged — it uses `kustomize build config/helmchart` + Helm packaging
- The `config/helmchart/kustomization.yaml` overlay chains through `../local-development/tilt` → `../local-development` → `../default` — changes to the kustomize base (Story 10.5) will flow through automatically

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.6]
- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.5]
- [Source: _bmad-output/project-context.md#Build & Dev Tooling]
- [Source: _bmad-output/implementation-artifacts/10-2-upgrade-helm-from-v3-11-to-v4.md]
- [Source: _bmad-output/implementation-artifacts/10-0a-kustomize-v3-to-v5-syntax-migration.md]
- [Source: config/helmchart/values.yaml.tpl — current kube_rbac_proxy config]
- [Source: config/helmchart/templates/manager.yaml — current sidecar injection]
- [Source: config/helmchart/templates/certificate.yaml — metrics cert resource]
- [Source: main.go — current metrics server options (lines 99-113)]
- [Kubebuilder metrics reference](https://book.kubebuilder.io/reference/metrics.html)

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
