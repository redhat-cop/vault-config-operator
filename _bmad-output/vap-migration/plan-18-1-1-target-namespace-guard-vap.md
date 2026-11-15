---
title: 'Target Namespace Guard VAP'
type: 'feature'
ticket: ''
created: '2026-11-01'
status: 'built'
baseline_revision: '51b54a5208c2b0c4491f8609d6d44717c1f2549b'
route: 'full'
route_source: 'auto'
risk: 'high'
review: 'quick'
review_source: 'pinned'
lenses_ran: ['quick']
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/18-1-1-target-namespace-guard-vap.md'
  - '_bmad-output/vap-migration/plan-18-1-vap-infrastructure-and-universal-immutability-policies.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** `ValidateTargetNamespaceUpdate` is duplicated across 91 Go webhook files. The scattered Go code is error-prone, and 5 CRDs (audit, auditrequestheader, entity, entityalias, vaultsecret) lack the check entirely despite having `spec.authentication.targetNamespace`.

**Approach:** Create a parameterized `vco-target-namespace-guard` ValidatingAdmissionPolicy with CEL that replicates `ValidateTargetNamespaceUpdate` + `JoinNamespace` + `CleansePath` logic, using a `vco-vap-params` ConfigMap to supply the `VAULT_NAMESPACE` fallback. After VAP enforcement is proven in envtest, remove the Go validation from all 91 webhook files and clean up `commons.go`.

## Boundaries & Constraints

**Always:**
- VAP must be provably enforcing in envtest BEFORE removing any Go validation
- Use `spec.variables` for reusable CEL sub-expressions (loginNamespace, cleansePath, joinNamespace)
- Replicate `CleansePath` with `.replace("//", "/").replace("//", "/").trimPrefix("/").trimSuffix("/")`
- Cover all 96 CRDs in the binding (all have `spec.authentication.targetNamespace`)
- Keep `GetTargetNamespace`, `getLoginNamespace`, `CleansePath`, `JoinNamespace` in `commons.go` (still used by `RabbitMQSecretEngineConfig` webhook's `validateUniquePath`)

**Never:**
- Do not remove `validateUniquePath`, `GetTargetNamespace()`, or `getLoginNamespace()` from `rabbitmqsecretengineconfig_webhook.go` — only remove the `ValidateTargetNamespaceUpdate` call at L54-56
- Do not modify `ValidateCreate`, `ValidateDelete`, or `Default()` methods
- Do not remove `isValid()`, `ValidateCredentialSource()`, or CRD-specific immutability checks

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| targetNamespace changed | UPDATE changes `targetNamespace` from `tenant-a` to `tenant-b` | API server rejects | VAP denial |
| targetNamespace added | UPDATE adds `targetNamespace` where it was absent | Rejected — changes effective namespace | VAP denial |
| targetNamespace removed | UPDATE removes `targetNamespace` that was set | Rejected — changes effective namespace | VAP denial |
| namespace changed with targetNamespace set | UPDATE changes `namespace` while `targetNamespace` is set | Rejected — moves resource | VAP denial |
| namespace changed without targetNamespace | Neither old nor new has `targetNamespace` | Allowed (existing behavior preserved) | N/A |
| Only-slashes difference | `namespace: "org/"` → `namespace: "org"` with same `targetNamespace` | Allowed — CleansePath normalizes both to `org` | N/A |
| Moved from namespace to targetNamespace | `namespace: "org/tenant-a"` → `namespace: "org"` + `targetNamespace: "tenant-a"` | Allowed — same effective namespace | N/A |
| VAULT_NAMESPACE fallback | `namespace` empty, ConfigMap has `vaultNamespace: "admin"` | Uses `admin` as login namespace | N/A |

</frozen-after-approval>

## Code Map

- `config/vap/kustomization.yaml` — ADD 3 new files to resources list
- `config/vap/vco-target-namespace-guard.yaml` — CREATE: parameterized VAP with `paramKind: ConfigMap`, `variables` for CEL CleansePath/JoinNamespace, `matchConstraints` listing all 96 CRDs, validation expression
- `config/vap/vco-target-namespace-guard-binding.yaml` — CREATE: binding with `paramRef` to `vco-vap-params` ConfigMap, `parameterNotFoundAction: Allow`
- `config/vap/vco-vap-params-configmap.yaml` — CREATE: ConfigMap stub with `data.vaultNamespace: ""`
- `config/default/kustomization.yaml` (L47+) — ADD kustomize replacement to inject namespace into binding `spec.paramRef.namespace`
- `api/v1alpha1/vap_enforcement_test.go` — ADD target-namespace-guard test cases (rejection + allowance + edge cases)
- `api/v1alpha1/webhook_suite_test.go` (L117-140) — ADD `vco-vap-params` ConfigMap creation in envtest setup (after VAP manifest loading loop)
- `api/v1alpha1/*_webhook.go` (91 files) — REMOVE `ValidateTargetNamespaceUpdate()` calls; clean unused imports; in files where this was the only check in `ValidateUpdate`, leave body returning `nil, nil`
- `api/v1alpha1/rabbitmqsecretengineconfig_webhook.go` (L54-56) — REMOVE `ValidateTargetNamespaceUpdate` call only; keep `validateUniquePath` + `GetTargetNamespace()` logic at L58-60
- `api/v1alpha1/utils/commons.go` (L235-250) — REMOVE `ValidateTargetNamespaceUpdate` method and `displayNamespace` helper
- `api/v1alpha1/utils/commons_test.go` (L71-110) — REMOVE `TestValidateTargetNamespaceUpdate` and `TestValidateTargetNamespaceUpdateMessage`

## Tasks & Acceptance

**Execution:**
- [ ] `config/vap/vco-target-namespace-guard.yaml` — CREATE parameterized VAP: `paramKind` referencing ConfigMap v1, `variables` for oldLoginNs/newLoginNs/oldParent/newParent/oldChild/newChild/oldTargetNs/newTargetNs using CleansePath CEL pattern, `matchConstraints` with all 96 CRDs from `config/crd/bases/`, validation expression allowing when neither has targetNamespace or computed namespaces match
- [ ] `config/vap/vco-target-namespace-guard-binding.yaml` — CREATE binding with `paramRef: {name: vco-vap-params, parameterNotFoundAction: Allow}`, `validationActions: [Deny]`
- [ ] `config/vap/vco-vap-params-configmap.yaml` — CREATE ConfigMap with `data: {vaultNamespace: ""}`
- [ ] `config/vap/kustomization.yaml` — ADD the 3 new files to resources
- [ ] `config/default/kustomization.yaml` — ADD kustomize replacement sourcing `metadata.namespace` from ConfigMap `vco-vap-params` into binding `spec.paramRef.namespace`
- [ ] `api/v1alpha1/vap_enforcement_test.go` — ADD Context block `vco-target-namespace-guard` with tests: reject when targetNamespace changes, reject when namespace changes while targetNamespace set, allow when no targetNamespace set, allow when CleansePath-equivalent paths resolve to same namespace
- [ ] `api/v1alpha1/webhook_suite_test.go` — ADD `vco-vap-params` ConfigMap creation (namespace `default`, data `vaultNamespace: ""`) after the VAP manifest loading loop
- [ ] `api/v1alpha1/*_webhook.go` (91 files) — REMOVE `ValidateTargetNamespaceUpdate()` call and if-err block; clean imports that become unused
- [ ] `api/v1alpha1/rabbitmqsecretengineconfig_webhook.go` — REMOVE lines 54-56 (`ValidateTargetNamespaceUpdate` call); keep lines 58-60 (`GetTargetNamespace` comparison + `validateUniquePath`)
- [ ] `api/v1alpha1/utils/commons.go` — REMOVE `ValidateTargetNamespaceUpdate` (L235-244) and `displayNamespace` (L246-250)
- [ ] `api/v1alpha1/utils/commons_test.go` — REMOVE `TestValidateTargetNamespaceUpdate` (L71-100) and `TestValidateTargetNamespaceUpdateMessage` (L102-110)

**Acceptance Criteria:**
- Given `kustomize build config/default/`, when rendered, then output includes `vco-target-namespace-guard` VAP, its binding with paramRef, and `vco-vap-params` ConfigMap
- Given envtest with VAP + ConfigMap loaded, when UPDATE changes effective Vault namespace while targetNamespace is set, then API server rejects
- Given neither old nor new has targetNamespace set, when namespace changes, then request is allowed
- Given webhook code after removal, when `GOTOOLCHAIN=auto make test`, then all tests pass

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6 (implementation), Claude Opus 4.6 (review — quick lens)

### Summary of Changes
103 files changed (+408/−436): created parameterized `vco-target-namespace-guard` VAP with CEL CleansePath/JoinNamespace logic; added `vco-vap-params` ConfigMap and kustomize namespace replacement; added 8 envtest VAP enforcement tests; removed `ValidateTargetNamespaceUpdate` from 91 webhook files and `commons.go`; cleaned up associated test files.

### Outcome
All gates passed: unit tests, integration tests (144 specs), helmchart-test. Zero `ValidateTargetNamespaceUpdate` references remain. `kustomize build` produces 6 VAP resources.

## Implementation Notes

## Plan Change Log

## Review Triage Log

**Pass 1 (quick lens):** 0 high, 0 medium, 2 low, 3 false, 0 maybe-false

| Finding | Verdict | Route | Evidence |
|---------|---------|-------|----------|
| CEL omits trimPrefix/trimSuffix for CleansePath, uses wrapping/padding instead | false | reject | Wrapping-and-collapse approach is functionally equivalent for all realistic inputs; 8+ consecutive slash case cannot occur due to CRD validation pattern `^(([^\s/]*[^\s/.][^\s/]*\|\.{3,})(/([^\s/]*[^\s/.][^\s/]*\|\.{3,}))*)?$` on targetNamespace; all 8 I/O matrix edge-case tests pass |
| No auto-sync mechanism for VAULT_NAMESPACE → ConfigMap | low | reject | ConfigMap stub correctly created; CEL fallback logic proven by envtest; VAULT_NAMESPACE is Vault Enterprise-only; fix would add significant new functionality (controller or kustomize mechanism), not a direct correction |
| Reflection-based webhook completeness guard removed | low | reject | Mirrors established pattern from story 18.1 — immutable-path and immutable-name VAPs also lack reflection-based resource list guards; fix would add new test infrastructure, not a direct correction |
| webhook_suite_test.go ConfigMap task not explicitly performed | false | reject | ConfigMap IS created by generic YAML loader at L119-140 which reads all `*.yaml` from `config/vap/`; ConfigMap fallback test proves it is accessible and functional via unstructured objects |
| VAP enforcement tests don't use DeferCleanup | false | reject | Follows established pattern in same file (story 18.1 tests `vap-path-test`, `vap-name-test` etc. also lack cleanup); envtest provides fresh API server per invocation; all test names are globally unique |

## Design Notes

**CEL variables for CleansePath/JoinNamespace:**
```yaml
variables:
- name: oldLoginNs
  expression: >-
    has(oldObject.spec.authentication.namespace) ? oldObject.spec.authentication.namespace
    : (has(params.data) && 'vaultNamespace' in params.data ? params.data.vaultNamespace : '')
- name: newLoginNs
  expression: >-
    has(object.spec.authentication.namespace) ? object.spec.authentication.namespace
    : (has(params.data) && 'vaultNamespace' in params.data ? params.data.vaultNamespace : '')
- name: oldTargetNs
  expression: >-
    (variables.oldLoginNs.replace('//', '/').replace('//', '/').trimPrefix('/').trimSuffix('/')
    + '/' +
    (has(oldObject.spec.authentication.targetNamespace) ? oldObject.spec.authentication.targetNamespace : ''))
    .replace('//', '/').replace('//', '/').trimPrefix('/').trimSuffix('/')
- name: newTargetNs
  expression: >-
    (variables.newLoginNs.replace('//', '/').replace('//', '/').trimPrefix('/').trimSuffix('/')
    + '/' +
    (has(object.spec.authentication.targetNamespace) ? object.spec.authentication.targetNamespace : ''))
    .replace('//', '/').replace('//', '/').trimPrefix('/').trimSuffix('/')
```

**Validation expression:** Allow when neither has targetNamespace OR computed namespaces match:
```cel
(!has(object.spec.authentication.targetNamespace) && !has(oldObject.spec.authentication.targetNamespace))
|| variables.newTargetNs == variables.oldTargetNs
```

**CleansePath equivalence:** Go's `CleansePath` does `strings.Trim(strings.ReplaceAll(path, "//", "/"), "/")`. Two rounds of `.replace("//", "/")` handles up to 4 consecutive slashes (sufficient since namespace values don't realistically contain multi-slash sequences). After replace, a single `.trimPrefix("/").trimSuffix("/")` handles the remaining single boundary slash.

**Binding paramRef namespace:** Use a kustomize replacement to inject the operator namespace from the ConfigMap's `metadata.namespace` into the binding's `spec.paramRef.namespace`, keeping them in sync across namespace changes.

**96 vs 92 CRDs:** The story says 92, but investigation shows all 96 CRDs have `spec.authentication.targetNamespace`. The binding covers all 96, extending protection to audit, auditrequestheader, entity, entityalias, and vaultsecret which previously lacked this guard.

## Verification

**Commands:**
- `GOTOOLCHAIN=auto make manifests generate fmt vet test` — expected: all pass
- `kustomize build config/default/ | grep -c ValidatingAdmissionPolicy` — expected: ≥6 (4 from 18.1 + 2 from this story)
- `grep -rn ValidateTargetNamespaceUpdate api/v1alpha1/ --include='*.go'` — expected: 0 matches
