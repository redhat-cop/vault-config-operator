# Story 18.1.1: Target Namespace Guard VAP

Status: ready-for-dev

## Story

As an operator developer,
I want a shared parameterized ValidatingAdmissionPolicy that prevents Vault namespace changes when `targetNamespace` is set,
So that the `ValidateTargetNamespaceUpdate` Go code can be removed from all 91 webhook files and replaced by a single CEL-based admission policy — completing the universal immutability VAP set started in Story 18.1.

## Acceptance Criteria

1. **Given** `config/vap/` **When** `kustomize build config/default/` is run **Then** the output includes `vco-target-namespace-guard` VAP, its binding (92 CRDs), and the `vco-vap-params` ConfigMap
2. **Given** any of the 92 CRDs with `spec.authentication` **When** a user submits an UPDATE that changes the effective Vault namespace while `targetNamespace` is set on either old or new **Then** the API server rejects the request
3. **Given** neither old nor new resource has `targetNamespace` set **When** `spec.authentication.namespace` changes **Then** the request is allowed (existing behavior preserved)
4. **Given** a ConfigMap `vco-vap-params` in the operator namespace **When** the operator's `VAULT_NAMESPACE` env var is set **Then** the ConfigMap contains the env var value, and the VAP uses it as the fallback for empty `spec.authentication.namespace`
5. **Given** the CEL expression **When** tested with edge cases (empty namespace, namespaces with leading/trailing slashes, nested namespaces like `org/team/tenant`) **Then** it produces the same result as the Go `JoinNamespace(CleansePath(parent) + "/" + CleansePath(child))` logic
6. **Given** the Go webhook code after `ValidateTargetNamespaceUpdate` removal **When** `GOTOOLCHAIN=auto make test` is run **Then** all tests pass
7. **Given** `api/v1alpha1/utils/commons.go` **When** `ValidateTargetNamespaceUpdate` is removed **Then** `displayNamespace` is also removed (only caller); `GetTargetNamespace` and `getLoginNamespace` are kept (still used by `RabbitMQSecretEngineConfig` webhook's `validateUniquePath`)

## Tasks / Subtasks

- [ ] Task 1 — Create `vco-target-namespace-guard` parameterized VAP (AC: #1, #2, #3, #4, #5)
  - [ ] 1.1 Create `config/vap/vco-target-namespace-guard.yaml` — ValidatingAdmissionPolicy with `spec.paramKind` referencing ConfigMap, CEL replicating `ValidateTargetNamespaceUpdate` + `JoinNamespace` + `CleansePath` logic
  - [ ] 1.2 Create `config/vap/vco-target-namespace-guard-binding.yaml` — ValidatingAdmissionPolicyBinding listing all 92 CRDs, paramRef to `vco-vap-params` ConfigMap
  - [ ] 1.3 Create `config/vap/vco-vap-params-configmap.yaml` — ConfigMap stub with `data.vaultNamespace` key
  - [ ] 1.4 Update `config/vap/kustomization.yaml` — add the 3 new files to resources
- [ ] Task 2 — Add envtest VAP enforcement smoke tests (AC: #2, #3, #5)
  - [ ] 2.1 Add target-namespace-guard test cases to `api/v1alpha1/vap_enforcement_test.go` — test rejection when targetNamespace is set and effective namespace changes
  - [ ] 2.2 Add allowance test — namespace change allowed when no targetNamespace is set
  - [ ] 2.3 Add edge-case tests for CleansePath-equivalent behavior (trailing slashes, nested namespaces)
  - [ ] 2.4 Create `vco-vap-params` ConfigMap in envtest setup (`webhook_suite_test.go`)
- [ ] Task 3 — Remove `ValidateTargetNamespaceUpdate` from webhook files (AC: #6)
  - [ ] 3.1 Remove `ValidateTargetNamespaceUpdate()` calls from all 91 webhook `ValidateUpdate` methods (all except `vaultsecret_webhook.go` which doesn't have it)
  - [ ] 3.2 Remove the `ValidateTargetNamespaceUpdate` call from `rabbitmqsecretengineconfig_webhook.go` Handle method (L54-56 after 18.1 removals); keep `validateUniquePath` + `GetTargetNamespace()` intact
  - [ ] 3.3 Clean up unused imports after removal
- [ ] Task 4 — Clean up commons.go (AC: #7)
  - [ ] 4.1 Remove `ValidateTargetNamespaceUpdate` method from `api/v1alpha1/utils/commons.go`
  - [ ] 4.2 Remove `displayNamespace` helper (only used by `ValidateTargetNamespaceUpdate`)
  - [ ] 4.3 Keep `GetTargetNamespace`, `getLoginNamespace`, `CleansePath`, `JoinNamespace` (still referenced)
  - [ ] 4.4 Remove `ValidateTargetNamespaceUpdate` test cases from `api/v1alpha1/utils/commons_test.go`
- [ ] Task 5 — Remove target-namespace-guard test cases from webhook tests (AC: #6)
  - [ ] 5.1 Remove any remaining `ValidateTargetNamespaceUpdate`-related test cases from `api/v1alpha1/webhook_validate_update_test.go`
- [ ] Task 6 — Code generation and validation (AC: #6)
  - [ ] 6.1 `GOTOOLCHAIN=auto make manifests generate fmt vet test` passes
  - [ ] 6.2 `kustomize build config/default/ | grep -c ValidatingAdmissionPolicy` returns ≥6 (2 from 18.1 + 2 from this + bindings)

## Dev Notes

### Overview

This story completes the universal immutability VAP set started in Story 18.1. It creates the third and most complex shared VAP — a **parameterized** policy that uses a ConfigMap to provide the `VAULT_NAMESPACE` fallback value.

**Prerequisite:** Story 18.1 (done) — `config/vap/` infrastructure exists, envtest is configured for VAPs.
**Must complete before:** Story 18.3 — which removes all remaining webhook code and expects `ValidateTargetNamespaceUpdate` to already be VAP-covered.

### CEL Expression Design

The Go logic to replicate:
```go
func (kc *KubeAuthConfiguration) ValidateTargetNamespaceUpdate(old *KubeAuthConfiguration) error {
    oldNamespace, newNamespace := old.GetTargetNamespace(), kc.GetTargetNamespace()
    if (kc.TargetNamespace != "" || old.TargetNamespace != "") && newNamespace != oldNamespace {
        return fmt.Errorf("spec.authentication cannot be updated to a different Vault namespace...")
    }
    return nil
}
```

Where:
- `GetTargetNamespace()` = `JoinNamespace(getLoginNamespace(), TargetNamespace)`
- `getLoginNamespace()` = `Namespace || os.Getenv("VAULT_NAMESPACE")`
- `JoinNamespace(parent, child)` = `CleansePath(CleansePath(parent) + "/" + CleansePath(child))`
- `CleansePath(path)` = `strings.Trim(strings.ReplaceAll(path, "//", "/"), "/")`

**CEL approach:** CEL lacks `strings.Trim()` but supports `trimPrefix("/")` and `trimSuffix("/")`. Since `CleansePath` only trims `/`, chaining these is equivalent. The `params.data.vaultNamespace` provides the `VAULT_NAMESPACE` fallback.

### Parameterized VAP Setup

- VAP `spec.paramKind`: `apiVersion: v1`, `kind: ConfigMap`
- Binding `spec.paramRef`: `name: "vco-vap-params"`, `namespace: "<operator-namespace>"`, `parameterNotFoundAction: Allow`
- ConfigMap: `data.vaultNamespace` holds the `VAULT_NAMESPACE` env var value
- The operator deployment creates/updates this ConfigMap on startup (or via kustomize manifest)

### 92 CRDs with `spec.authentication`

All 92 webhook types (including VaultSecret) have `spec.authentication`. All are listed in the binding. Note: VaultSecret's webhook does NOT currently call `ValidateTargetNamespaceUpdate` — the VAP adds this protection automatically.

### References

- [Source: api/v1alpha1/utils/commons.go#ValidateTargetNamespaceUpdate — lines 237-244]
- [Source: api/v1alpha1/utils/commons.go#GetTargetNamespace — lines 231-233]
- [Source: api/v1alpha1/utils/commons.go#CleansePath — line 428-430]
- [Source: api/v1alpha1/utils/commons.go#JoinNamespace — lines 262-264]
- [Source: _bmad-output/implementation-artifacts/18-1-vap-infrastructure-and-universal-immutability-policies.md#Part D]
- [Source: _bmad-output/vap-migration/deferred-work.md]

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
