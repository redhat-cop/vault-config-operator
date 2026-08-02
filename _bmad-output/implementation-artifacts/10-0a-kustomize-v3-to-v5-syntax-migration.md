# Story 10.0a: Kustomize v3 → v5 Syntax Migration

Status: ready-for-dev

## Story

As an operator developer,
I want to migrate the kustomize manifests from deprecated v3 syntax to v5 syntax,
So that the project is compatible with Kustomize v5+ and follows current Kubebuilder go/v4 scaffold conventions.

## Acceptance Criteria

1. **Given** `config/default/kustomization.yaml` uses `bases:` to reference sub-directories
   **When** `bases:` is replaced with `resources:`
   **Then** `kustomize build config/default` produces the same output

2. **Given** `config/default/kustomization.yaml` uses `patchesStrategicMerge:` to reference patch files
   **When** `patchesStrategicMerge:` is replaced with `patches:` using `- path:` entries
   **Then** `kustomize build config/default` produces the same output with patches applied

3. **Given** `config/default/kustomization.yaml` uses `vars:` with `objref:` and `fieldref:` for variable substitution
   **When** `vars:` is replaced with `replacements:` using source/target field path selectors
   **Then** `kustomize build config/default` correctly substitutes values in target manifests
   **And** the output is semantically identical to the previous `vars:` approach

4. **Given** all kustomize syntax has been migrated across all `config/` kustomization files
   **When** `make manifests` and `make deploy` are run
   **Then** both succeed without warnings about deprecated kustomize fields

## Tasks / Subtasks

- [ ] Task 1: Capture baseline output for regression comparison (AC: #1-3)
  - [ ] Run `kustomize build config/default > /tmp/kustomize-baseline.yaml` and save output
  - [ ] Run `kustomize build config/helmchart > /tmp/kustomize-helmchart-baseline.yaml`
  - [ ] Run `kustomize build config/local-development > /tmp/kustomize-localdev-baseline.yaml`
  - [ ] Run `kustomize build config/local-development/tilt > /tmp/kustomize-tilt-baseline.yaml`

- [ ] Task 2: Migrate `config/default/kustomization.yaml` — `bases:` → `resources:` (AC: #1)
  - [ ] Replace `bases:` keyword with `resources:`
  - [ ] Keep all entries unchanged (relative paths to sub-directories)

- [ ] Task 3: Migrate `config/default/kustomization.yaml` — `patchesStrategicMerge:` → `patches:` (AC: #2)
  - [ ] Replace `patchesStrategicMerge:` with `patches:`
  - [ ] Convert each bare filename entry to `- path: <filename>` format

- [ ] Task 4: Migrate `config/default/kustomization.yaml` — `vars:` → `replacements:` (AC: #3)
  - [ ] Remove the entire `vars:` block (3 active entries + commented cert-manager entries)
  - [ ] Update target files to use placeholder values instead of `$(VAR_NAME)` notation
  - [ ] Add `replacements:` block with source/target field path selectors
  - [ ] Verify `kustomize build config/default` output matches baseline for substituted values

- [ ] Task 5: Migrate other kustomization files with deprecated syntax (AC: #4)
  - [ ] `config/helmchart/kustomization.yaml`: `bases:` → `resources:`
  - [ ] `config/local-development/kustomization.yaml`: `bases:` → `resources:`
  - [ ] `config/local-development/tilt/kustomization.yaml`: `patchesStrategicMerge:` → `patches:`
  - [ ] `config/crd/kustomization.yaml`: `patchesStrategicMerge: []` → `patches: []`

- [ ] Task 6: Verify all builds produce identical output (AC: #1-4)
  - [ ] Compare `kustomize build config/default` against baseline
  - [ ] Compare `kustomize build config/helmchart` against baseline
  - [ ] Compare `kustomize build config/local-development` against baseline
  - [ ] Compare `kustomize build config/local-development/tilt` against baseline
  - [ ] Run `make manifests`
  - [ ] Run `make deploy` (dry-run or against test cluster)
  - [ ] Run `make helmchart` to verify helm chart generation still works

## Dev Notes

### Critical Implementation Order

1. Capture baseline output FIRST (Task 1)
2. Migrate simple syntax changes: `bases:` → `resources:`, `patchesStrategicMerge:` → `patches:` (Tasks 2, 3, 5)
3. Migrate `vars:` → `replacements:` LAST — this is the complex part (Task 4)
4. Verify after each step with `kustomize build` diff

### Files Requiring Changes

| File | Change | Complexity |
|------|--------|-----------|
| `config/default/kustomization.yaml` | All three migrations: `bases:`, `patchesStrategicMerge:`, `vars:` | High |
| `config/helmchart/kustomization.yaml` | `bases:` → `resources:` | Trivial |
| `config/local-development/kustomization.yaml` | `bases:` → `resources:` | Trivial |
| `config/local-development/tilt/kustomization.yaml` | `patchesStrategicMerge:` → `patches:` | Low |
| `config/crd/kustomization.yaml` | `patchesStrategicMerge: []` → `patches: []` | Trivial |
| `config/prometheus/monitor.yaml` | Replace `$(VAR)` placeholders with template values for replacements | Medium |
| `config/prometheus/rolebinding.yaml` | Replace `$(ROLE_NAME)` with template value | Low |
| `config/certmanager/certificate.yaml` | Replace `$(VAR)` placeholders with template values (commented section) | Low |

### `bases:` → `resources:` Migration (Trivial)

Direct keyword replacement. No other changes needed.

```yaml
# Before
bases:
- ../crd
- ../rbac

# After
resources:
- ../crd
- ../rbac
```

### `patchesStrategicMerge:` → `patches:` Migration (Low)

Each bare path entry becomes a `- path:` entry.

```yaml
# Before (config/default/kustomization.yaml)
patchesStrategicMerge:
- manager_auth_proxy_patch.yaml
- manager_webhook_patch.yaml

# After
patches:
- path: manager_auth_proxy_patch.yaml
- path: manager_webhook_patch.yaml
```

For `config/local-development/tilt/kustomization.yaml`:
```yaml
# Before
patchesStrategicMerge:
- ./remove-namespace.yaml
- ./replace-image.yaml

# After
patches:
- path: ./remove-namespace.yaml
- path: ./replace-image.yaml
```

For `config/crd/kustomization.yaml`:
```yaml
# Before
patchesStrategicMerge: []

# After
patches: []
```

### `vars:` → `replacements:` Migration (Complex — Key Guidance)

This is the most complex migration. The three active `vars:` entries are:

| Var Name | Source Object | Source Field | Used In |
|----------|--------------|-------------|---------|
| `METRICS_SERVICE_NAME` | `Service/controller-manager-metrics-service` (v1) | `metadata.name` | `config/prometheus/monitor.yaml`, `config/certmanager/certificate.yaml` |
| `METRICS_SERVICE_NAMESPACE` | `Service/controller-manager-metrics-service` (v1) | `metadata.namespace` | `config/prometheus/monitor.yaml`, `config/certmanager/certificate.yaml` |
| `ROLE_NAME` | `Role/prometheus-k8s` (rbac.authorization.k8s.io/v1) | `metadata.name` | `config/prometheus/rolebinding.yaml` |

#### Target File: `config/prometheus/monitor.yaml`

Current value: `serverName: $(METRICS_SERVICE_NAME).$(METRICS_SERVICE_NAMESPACE).svc`

After vars removal, the target file should use a placeholder pattern that `replacements:` with `delimiter`/`index` options can operate on:

```yaml
# Change monitor.yaml from:
serverName: $(METRICS_SERVICE_NAME).$(METRICS_SERVICE_NAMESPACE).svc

# To a placeholder for delimiter-based replacement:
serverName: METRICS_SERVICE_NAME.METRICS_SERVICE_NAMESPACE.svc
```

Then the `replacements:` block in `config/default/kustomization.yaml`:

```yaml
replacements:
- source:
    kind: Service
    version: v1
    name: controller-manager-metrics-service
    fieldPath: metadata.name
  targets:
  - select:
      kind: ServiceMonitor
      name: controller-manager-metrics-monitor
    fieldPaths:
    - spec.endpoints.0.tlsConfig.serverName
    options:
      delimiter: "."
      index: 0
- source:
    kind: Service
    version: v1
    name: controller-manager-metrics-service
    fieldPath: metadata.namespace
  targets:
  - select:
      kind: ServiceMonitor
      name: controller-manager-metrics-monitor
    fieldPaths:
    - spec.endpoints.0.tlsConfig.serverName
    options:
      delimiter: "."
      index: 1
```

#### Target File: `config/prometheus/rolebinding.yaml`

Current value: `name: $(ROLE_NAME)` (in `roleRef.name`)

```yaml
# Change rolebinding.yaml from:
roleRef:
  name: $(ROLE_NAME)

# To:
roleRef:
  name: prometheus-k8s
```

Then the `replacements:` block entry:

```yaml
- source:
    kind: Role
    version: v1
    group: rbac.authorization.k8s.io
    name: prometheus-k8s
    fieldPath: metadata.name
  targets:
  - select:
      kind: RoleBinding
      name: prometheus-k8s
    fieldPaths:
    - roleRef.name
```

#### Target File: `config/certmanager/certificate.yaml`

Current values in `dnsNames` list:
- `$(METRICS_SERVICE_NAME).$(METRICS_SERVICE_NAMESPACE).svc`
- `$(METRICS_SERVICE_NAME).$(METRICS_SERVICE_NAMESPACE).svc.cluster.local`

```yaml
# Change certificate.yaml from:
dnsNames:
- $(METRICS_SERVICE_NAME).$(METRICS_SERVICE_NAMESPACE).svc
- $(METRICS_SERVICE_NAME).$(METRICS_SERVICE_NAMESPACE).svc.cluster.local

# To placeholder values:
dnsNames:
- METRICS_SERVICE_NAME.METRICS_SERVICE_NAMESPACE.svc
- METRICS_SERVICE_NAME.METRICS_SERVICE_NAMESPACE.svc.cluster.local
```

Then add replacement entries targeting the Certificate `metrics-cert`:

```yaml
- source:
    kind: Service
    version: v1
    name: controller-manager-metrics-service
    fieldPath: metadata.name
  targets:
  - select:
      kind: Certificate
      group: cert-manager.io
      version: v1
      name: metrics-cert
    fieldPaths:
    - spec.dnsNames.0
    - spec.dnsNames.1
    options:
      delimiter: "."
      index: 0
- source:
    kind: Service
    version: v1
    name: controller-manager-metrics-service
    fieldPath: metadata.namespace
  targets:
  - select:
      kind: Certificate
      group: cert-manager.io
      version: v1
      name: metrics-cert
    fieldPaths:
    - spec.dnsNames.0
    - spec.dnsNames.1
    options:
      delimiter: "."
      index: 1
```

**Note:** The certmanager section is currently commented out in `config/default/kustomization.yaml`. The `replacements:` entries for cert-manager should also be commented out, matching the same pattern. Only uncomment when cert-manager integration is enabled.

### Handling the Commented CERTMANAGER Vars

The commented-out cert-manager vars (`CERTIFICATE_NAMESPACE`, `CERTIFICATE_NAME`, `SERVICE_NAMESPACE`, `SERVICE_NAME`) should be migrated to commented-out `replacements:` entries for future use. These are used in `config/certmanager/certificate.yaml` for the webhook serving cert (the `serving-cert` Certificate, as opposed to `metrics-cert`).

### Verification Strategy

After migration, run these commands and diff against baselines:

```bash
# Verify default overlay
kustomize build config/default > /tmp/kustomize-after.yaml
diff /tmp/kustomize-baseline.yaml /tmp/kustomize-after.yaml

# Verify helmchart overlay
kustomize build config/helmchart > /tmp/kustomize-helmchart-after.yaml
diff /tmp/kustomize-helmchart-baseline.yaml /tmp/kustomize-helmchart-after.yaml

# Verify local-development overlay
kustomize build config/local-development > /tmp/kustomize-localdev-after.yaml
diff /tmp/kustomize-localdev-baseline.yaml /tmp/kustomize-localdev-after.yaml

# Verify tilt overlay
kustomize build config/local-development/tilt > /tmp/kustomize-tilt-after.yaml
diff /tmp/kustomize-tilt-baseline.yaml /tmp/kustomize-tilt-after.yaml

# Final verification via Makefile
make manifests
make helmchart
```

The output should be **semantically identical** — the only acceptable differences are YAML formatting (key ordering, whitespace) that don't affect semantics.

### Anti-Patterns to Avoid

- **DO NOT** upgrade `KUSTOMIZE_VERSION` in Makefile — that's Story 10.4
- **DO NOT** remove kube-rbac-proxy related files or patches — that's Story 10.5
- **DO NOT** modify Helm chart templates or values — only the kustomize source files
- **DO NOT** remove the `patchesJson6902:` entries in `config/helmchart/kustomization.yaml` or `config/local-development/tilt/kustomization.yaml` — `patchesJson6902:` is NOT deprecated (only `patchesStrategicMerge:` is)
- **DO NOT** change the `configurations:` field in `config/crd/kustomization.yaml` — it is not deprecated
- **DO NOT** forget to update commented-out entries — they serve as documentation for cert-manager enablement

### Kustomize Version Compatibility Note

The current `KUSTOMIZE_VERSION` is `v5.4.3` (in Makefile). All kustomize v5.x versions support `replacements:`. The `vars:` keyword still works in v5 with deprecation warnings but may be removed in future versions. This migration ensures forward compatibility.

### Dependency on Story 10.0

Story 10.0 moves `controllers/` → `internal/controller/`. This story (10.0a) operates exclusively on `config/` directory files and has NO Go code changes. It can technically be implemented in parallel with 10.0, but the sprint status shows 10.0 as ready-for-dev first. Either ordering works.

### Previous Story Intelligence

Story 10.0 (the predecessor in this epic) establishes the go/v4 directory layout. Its anti-patterns explicitly state: "DO NOT touch kustomize manifests syntax — that's Story 10.0a". This confirms clean separation of concerns between the two stories.

### Project Structure Notes

All changes are confined to `config/` directory kustomization YAML files and the target manifests they reference. No Go code, no tests, no CI changes.

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 10.0a]
- [Kustomize replacements reference](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/replacements/)
- [Kubebuilder v3→v4 migration guide (kustomize section)](https://book-v3.book.kubebuilder.io/migration/manually_migration_guide_gov3_to_gov4)
- [Source: _bmad-output/project-context.md#Build & Dev Tooling]
- [Source: config/default/kustomization.yaml — current deprecated syntax]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Previous Review Notes (First Attempt — GPT 5.4)

> These findings are from a code review of the first implementation attempt. Commit references are no longer valid but the architectural guidance remains relevant.

- **[Decision]** AC4 verification (`make deploy`) was skipped because it requires a live cluster. Decide whether to accept alternate evidence or require actual `make deploy` success.
- **[Patch]** The tilt overlay `config/local-development/tilt/kustomization.yaml` depends on an ignored/generated `replace-image.yaml` — `kustomize build` is not reproducible from a clean checkout unless that file already exists.
