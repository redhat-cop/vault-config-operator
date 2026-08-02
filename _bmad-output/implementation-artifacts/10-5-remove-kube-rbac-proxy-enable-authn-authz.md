# Story 10.5: Remove kube-rbac-proxy and Enable controller-runtime authn/authz

Status: ready-for-dev

## Story

As an operator developer,
I want to remove the kube-rbac-proxy sidecar and replace it with controller-runtime's built-in `WithAuthenticationAndAuthorization` filter for metrics endpoint protection,
So that the operator no longer depends on an external proxy image for metrics security and follows the current Kubebuilder best practice.

## Acceptance Criteria

1. **Given** the kube-rbac-proxy sidecar is currently injected via `manager_auth_proxy_patch.yaml`
   **When** the patch and all auth_proxy RBAC files are removed
   **Then** `kustomize build config/default` produces a Deployment with only the manager container (no sidecar)

2. **Given** `cmd/main.go` binds metrics to `127.0.0.1:8080` with proxy forwarding
   **When** the metrics bind address is changed to `:8443` with `SecureServing: true` and `FilterProvider: filters.WithAuthenticationAndAuthorization`
   **Then** the metrics endpoint is directly served by the manager with authn/authz protection

3. **Given** the `--metrics-secure` flag is not currently bound
   **When** `flag.BoolVar(&secureMetrics, "metrics-secure", true, ...)` is added
   **Then** metrics security can be toggled via command-line flag (default: enabled)

4. **Given** new RBAC files are created for metrics authn/authz
   **When** `make manifests` is run
   **Then** the generated RBAC allows the controller-manager to perform token reviews and subject access reviews

5. **Given** all changes are applied
   **When** `make build`, `make test`, and `make manifests` are run
   **Then** all succeed without errors

## Tasks / Subtasks

- [ ] Task 1: Remove kube-rbac-proxy files (AC: #1)
  - [ ] Delete `config/default/manager_auth_proxy_patch.yaml`
  - [ ] Delete `config/rbac/auth_proxy_service.yaml`
  - [ ] Delete `config/rbac/auth_proxy_role.yaml`
  - [ ] Delete `config/rbac/auth_proxy_role_binding.yaml`
  - [ ] Delete `config/rbac/auth_proxy_client_clusterrole.yaml`

- [ ] Task 2: Create new metrics RBAC files (AC: #4)
  - [ ] Create `config/rbac/metrics_auth_role.yaml` — ClusterRole granting `tokenreviews` and `subjectaccessreviews` create
  - [ ] Create `config/rbac/metrics_auth_role_binding.yaml` — ClusterRoleBinding for the controller-manager ServiceAccount
  - [ ] Create `config/rbac/metrics_reader_role.yaml` — ClusterRole with GET access to `/metrics` nonResourceURL

- [ ] Task 3: Create metrics Service and Deployment patch (AC: #1)
  - [ ] Create `config/default/metrics_service.yaml` — Service on port 8443 targeting the manager container directly (replaces `auth_proxy_service.yaml` that was in `config/rbac/`)
  - [ ] Create `config/default/manager_metrics_patch.yaml` — Deployment patch adding `--metrics-bind-address=:8443`, `--metrics-secure`, `--health-probe-bind-address=:8081`, `--leader-elect` args, and containerPort 8443

- [ ] Task 4: Update kustomization files (AC: #1, #4)
  - [ ] `config/rbac/kustomization.yaml`: remove 4 `auth_proxy_*` entries, add 3 `metrics_*` entries
  - [ ] `config/default/kustomization.yaml`: remove `manager_auth_proxy_patch.yaml` from patches, add `metrics_service.yaml` to resources, add `manager_metrics_patch.yaml` to patches

- [ ] Task 5: Update `cmd/main.go` for authn/authz (AC: #2, #3)
  - [ ] Add import `"sigs.k8s.io/controller-runtime/pkg/metrics/filters"`
  - [ ] Add `flag.BoolVar(&secureMetrics, "metrics-secure", true, "If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")`
  - [ ] Change `metrics-bind-address` default from `":8080"` to `":8443"`
  - [ ] After `metricsServerOptions` struct initialization, add: `if secureMetrics { metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization }`

- [ ] Task 6: Verify builds and kustomize output (AC: #1-5)
  - [ ] Run `go build ./cmd/` (confirms compilation with new import)
  - [ ] Run `kustomize build config/default` and verify: only manager container (no sidecar), args include `--metrics-secure`, Service for metrics present
  - [ ] Run `make manifests generate`
  - [ ] Run `make fmt vet`
  - [ ] Run `make test`

## Dev Notes

### Critical Prerequisites

This story requires Stories 10.0 and 10.0a to be completed first:
- **Story 10.0**: `main.go` → `cmd/main.go`, `controllers/` → `internal/controller/` (go/v4 layout)
- **Story 10.0a**: kustomize syntax migrated to v5 (`resources:`, `patches:`, `replacements:`)

All file paths in this story assume the post-10.0/10.0a state.

### Implementation Order

1. Create new RBAC files (Task 2)
2. Create metrics Service and Deployment patch (Task 3)
3. Update kustomization files (Task 4) — do this BEFORE deleting old files so you can compare kustomize output
4. Delete old kube-rbac-proxy files (Task 1)
5. Update `cmd/main.go` (Task 5)
6. Verify (Task 6)

### controller-runtime Version Compatibility

The project uses controller-runtime **v0.24.1** (from Epic 8). The `WithAuthenticationAndAuthorization` filter was introduced in **v0.18.4**. No dependency changes needed.

Import path: `sigs.k8s.io/controller-runtime/pkg/metrics/filters`

### File Content Specifications

#### `config/rbac/metrics_auth_role.yaml`

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: metrics-auth-role
rules:
- apiGroups:
  - authentication.k8s.io
  resources:
  - tokenreviews
  verbs:
  - create
- apiGroups:
  - authorization.k8s.io
  resources:
  - subjectaccessreviews
  verbs:
  - create
```

#### `config/rbac/metrics_auth_role_binding.yaml`

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: metrics-auth-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: metrics-auth-role
subjects:
- kind: ServiceAccount
  name: controller-manager
  namespace: system
```

#### `config/rbac/metrics_reader_role.yaml`

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: metrics-reader
rules:
- nonResourceURLs:
  - "/metrics"
  verbs:
  - get
```

#### `config/default/metrics_service.yaml`

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    control-plane: vault-config-operator
  name: controller-manager-metrics-service
  namespace: system
spec:
  ports:
  - name: https
    port: 8443
    protocol: TCP
    targetPort: 8443
  selector:
    control-plane: vault-config-operator
```

**Key differences from the old `auth_proxy_service.yaml`:**
- Moved from `config/rbac/` to `config/default/` (Kubebuilder v4 convention)
- Removed OpenShift annotation `service.alpha.openshift.io/serving-cert-secret-name` — controller-runtime generates self-signed TLS certs internally; the OpenShift serving cert is no longer needed for the metrics endpoint
- `targetPort` changed from `https` (named port on kube-rbac-proxy) to `8443` (numeric, targeting manager directly)

#### `config/default/manager_metrics_patch.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  template:
    spec:
      containers:
      - name: manager
        args:
        - "--metrics-bind-address=:8443"
        - "--metrics-secure"
        - "--health-probe-bind-address=:8081"
        - "--leader-elect"
        ports:
        - containerPort: 8443
          name: https
          protocol: TCP
```

**What this replaces from `manager_auth_proxy_patch.yaml`:**
- The old patch added a kube-rbac-proxy sidecar container, set `--metrics-bind-address=127.0.0.1:8080` (localhost, behind proxy), and mounted TLS cert volumes
- The new patch sets `--metrics-bind-address=:8443` (all interfaces, no proxy), `--metrics-secure` (enables authn/authz), and exposes port 8443 directly on the manager container
- No volume mounts needed — controller-runtime generates self-signed certs automatically

### Kustomization File Changes

#### `config/rbac/kustomization.yaml` (after Story 10.0a)

Remove these 4 entries:
```yaml
- auth_proxy_service.yaml
- auth_proxy_role.yaml
- auth_proxy_role_binding.yaml
- auth_proxy_client_clusterrole.yaml
```

Add these 3 entries:
```yaml
- metrics_auth_role.yaml
- metrics_auth_role_binding.yaml
- metrics_reader_role.yaml
```

The comment block "Comment the following 4 lines if you want to disable the auth proxy" should be replaced with:
```yaml
# The following RBAC configurations are used to protect
# the metrics endpoint with authn/authz. These configurations
# ensure that only authorized users and service accounts
# can access the metrics endpoint.
- metrics_auth_role.yaml
- metrics_auth_role_binding.yaml
- metrics_reader_role.yaml
```

#### `config/default/kustomization.yaml` (after Story 10.0a)

In the `resources:` section, add `metrics_service.yaml`:
```yaml
resources:
- ../crd
- ../rbac
- ../manager
- ../webhook
- ../prometheus
- metrics_service.yaml
```

In the `patches:` section, replace `manager_auth_proxy_patch.yaml` with `manager_metrics_patch.yaml`:
```yaml
patches:
- path: manager_metrics_patch.yaml
- path: manager_webhook_patch.yaml
```

Update the comment on the patch from "Protect the /metrics endpoint by putting it behind auth" to something like:
```yaml
# [METRICS] The following patch enables the metrics endpoint using HTTPS and the port :8443.
# More info: https://book.kubebuilder.io/reference/metrics.html
- path: manager_metrics_patch.yaml
```

### `cmd/main.go` Changes (Exact Diff)

**1. Add import** — in the import block, add:
```go
"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
```

**2. Bind the `--metrics-secure` flag** — add this line after the existing `enableHTTP2` flag declaration area (before `flag.Parse()`):
```go
flag.BoolVar(&secureMetrics, "metrics-secure", true,
    "If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
```

**3. Change metrics-bind-address default** — change:
```go
flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
```
to:
```go
flag.StringVar(&metricsAddr, "metrics-bind-address", ":8443", "The address the metric endpoint binds to.")
```

**4. Add FilterProvider** — after the `metricsServerOptions` struct initialization block (after line `TLSOpts: tlsOpts,` and before the `}`), add the FilterProvider conditionally. Change the block to:
```go
metricsServerOptions := metricsserver.Options{
    BindAddress:   metricsAddr,
    SecureServing: secureMetrics,
    TLSOpts:       tlsOpts,
}

if secureMetrics {
    metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
}
```

Remove the long TODO comment about TLSOpts since controller-runtime's self-signed cert approach is now the standard pattern for non-CertManager deployments.

### Prometheus ServiceMonitor Impact

The existing `config/prometheus/monitor.yaml` ServiceMonitor references:
- `port: https` — still valid (the new Service keeps port name `https`)
- `bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token` — still valid (controller-runtime's authn validates bearer tokens via TokenReview)
- `tlsConfig.ca.secret.name: vault-config-operator-certs` — **may need updating** since controller-runtime uses self-signed certs (not the OpenShift-generated ones)

**This story does NOT modify the ServiceMonitor.** The TLS configuration adjustment is a known follow-up for Story 10.7 (CI pipeline adaptation and end-to-end validation). For development/testing, Prometheus will still scrape metrics successfully if `insecureSkipVerify: true` is set, or if CertManager is enabled separately.

### RBAC Name Mapping

| Old File | Old Resource Name | New File | New Resource Name |
|----------|-------------------|----------|-------------------|
| `auth_proxy_role.yaml` | `proxy-role` | `metrics_auth_role.yaml` | `metrics-auth-role` |
| `auth_proxy_role_binding.yaml` | `proxy-rolebinding` | `metrics_auth_role_binding.yaml` | `metrics-auth-rolebinding` |
| `auth_proxy_client_clusterrole.yaml` | `metrics-reader` | `metrics_reader_role.yaml` | `metrics-reader` (unchanged) |
| `auth_proxy_service.yaml` | `controller-manager-metrics-service` | `metrics_service.yaml` | `controller-manager-metrics-service` (unchanged) |

The Service name must remain `controller-manager-metrics-service` because the `replacements:` block in `config/default/kustomization.yaml` (from Story 10.0a) references it as a source by name.

### Anti-Patterns to Avoid

- **DO NOT** modify Helm chart templates or values — that's Story 10.6
- **DO NOT** modify the Prometheus ServiceMonitor (`config/prometheus/monitor.yaml`) — TLS config adjustments are Story 10.7
- **DO NOT** remove or modify the `manager_webhook_patch.yaml` — it's unrelated to metrics
- **DO NOT** change the `replacements:` block in `config/default/kustomization.yaml` — the Service name hasn't changed, so replacements still resolve correctly
- **DO NOT** add CertManager integration for metrics TLS — that's a separate optional enhancement
- **DO NOT** remove the `config/prometheus/` directory or its RBAC files — Prometheus integration is independent of the kube-rbac-proxy removal
- **DO NOT** change the metrics Service name from `controller-manager-metrics-service` — the `replacements:` in the default kustomization and the Prometheus config depend on this name

### Verification Strategy

After all changes, verify with these commands:

```bash
# 1. Compilation check
go build ./cmd/

# 2. Kustomize output inspection
kustomize build config/default | grep -A 50 "kind: Deployment" | head -80
# Verify: only one container (manager), no kube-rbac-proxy sidecar
# Verify: manager args include --metrics-bind-address=:8443, --metrics-secure
# Verify: manager has containerPort 8443

kustomize build config/default | grep -A 10 "kind: Service" | grep -A 10 "metrics"
# Verify: metrics Service exists with port 8443

kustomize build config/default | grep -A 5 "kind: ClusterRole" | grep -A 5 "metrics"
# Verify: metrics-auth-role and metrics-reader ClusterRoles exist

# 3. Full make targets
make manifests generate
make fmt vet
make test
```

### Architecture Context

The project uses the standard controller-runtime reconciler pattern with `ReconcilerBase` embedding. The metrics configuration is in `cmd/main.go` and is manager-wide — it does not affect individual controller reconcilers.

After this change:
- **Metrics endpoint**: served directly by the manager on `:8443` with HTTPS (self-signed cert)
- **Authentication**: incoming requests must present a valid Kubernetes bearer token (validated via TokenReview against the API server)
- **Authorization**: the bearer token's identity must have `get` permission on the `/metrics` non-resource URL (granted by the `metrics-reader` ClusterRole)
- **No sidecar**: the kube-rbac-proxy container is completely removed from the Deployment

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.5]
- [Kubebuilder Metrics Reference](https://book.kubebuilder.io/reference/metrics.html)
- [controller-runtime filters.WithAuthenticationAndAuthorization](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/metrics/filters#WithAuthenticationAndAuthorization)
- [Kubebuilder kube-rbac-proxy deprecation discussion](https://github.com/kubernetes-sigs/kubebuilder/discussions/3907)
- [Kubebuilder PR #4003 — authn/authz for metrics](https://github.com/kubernetes-sigs/kubebuilder/pull/4003)
- [Source: _bmad-output/project-context.md#Technology Stack & Versions]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Previous Review Notes (First Attempt — GPT 5.4)

> These findings are from a code review of the first implementation attempt. Commit references are no longer valid but the architectural guidance remains relevant.

- **[Decision]** After removing the kube-rbac-proxy sidecar, `config/prometheus/monitor.yaml` still references the old serving-cert trust anchor (`vault-config-operator-certs`), but this story moves metrics to controller-runtime self-signed HTTPS. This may temporarily break the default Prometheus scrape path. Decide whether Story 10.5 must maintain scrape continuity or if it's acceptable to fix in Story 10.6/10.7.
