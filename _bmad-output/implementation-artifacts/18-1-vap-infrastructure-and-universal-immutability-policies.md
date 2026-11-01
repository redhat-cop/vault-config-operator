# Story 18.1: VAP Infrastructure + Universal Immutability Policies

Status: ready-for-dev

## Story

As an operator developer,
I want shared ValidatingAdmissionPolicy manifests for the three universal immutability rules (path, name, target-namespace-guard),
So that all CRDs are consistently protected by admission-time validation — including the 35 CRDs currently missing checks.

## Acceptance Criteria

1. **Given** the operator source tree **When** a developer runs `make manifests` **Then** the `config/vap/` directory contains all VAP and binding manifests, included via kustomize overlay
2. **Given** the envtest suite **When** tests start **Then** the API server is configured with `--enable-admission-plugins=ValidatingAdmissionPolicy` and VAP manifests are loaded
3. **Given** any of the 72 CRDs with `spec.path` (existing 68 + AzureAuthEngineRole, GCPAuthEngineRole, JWTOIDCAuthEngineRole, LDAPAuthEngineGroup) **When** a user submits an UPDATE that changes `spec.path` **Then** the API server rejects the request with message "spec.path cannot be updated"
4. **Given** any of the 72 CRDs **When** a user submits an UPDATE that does NOT change `spec.path` **Then** the request is allowed
5. **Given** any of the 65 CRDs with `spec.name` used in `GetPath()` (existing 34 + 31 missing) **When** a user submits an UPDATE that changes `spec.name` **Then** the API server rejects the request with message "spec.name cannot be updated"
6. **Given** any of the 65 CRDs **When** a user submits an UPDATE that does NOT change `spec.name` **Then** the request is allowed
7. **Given** a CRD using `ResourceName(spec.Name, metadata.Name)` where `spec.name` was originally empty **When** a user submits an UPDATE adding `spec.name` (previously relying on `metadata.name`) **Then** the request is rejected — adding `spec.name` after creation would change the effective Vault path
8. **Given** any of the 92 CRDs with `spec.authentication` (all types including VaultSecret) **When** a user submits an UPDATE that changes the effective Vault namespace while `targetNamespace` is set on either old or new **Then** the API server rejects the request
9. **Given** a ConfigMap `vco-vap-params` in the operator namespace **When** the operator's `VAULT_NAMESPACE` env var is set **Then** the ConfigMap contains the env var value, and the VAP uses it as the fallback for empty `spec.authentication.namespace`
10. **Given** neither old nor new resource has `targetNamespace` set **When** `spec.authentication.namespace` changes **Then** the request is allowed (existing behavior preserved)
11. **Given** the migrated validating logic passes via VAPs **When** the corresponding Go validation code is removed from webhook `ValidateUpdate` methods **Then** `make test` still passes

## Tasks / Subtasks

- [ ] Task 1 — Create `config/vap/` kustomize overlay (AC: #1)
  - [ ] 1.1 Create `config/vap/kustomization.yaml` listing all VAP + binding YAML files as resources
  - [ ] 1.2 Add `- ../vap` to `config/default/kustomization.yaml` under `resources:`
  - [ ] 1.3 Verify `kustomize build config/default/` includes the VAP manifests in rendered output
- [ ] Task 2 — Create `vco-immutable-path` shared VAP (AC: #3, #4)
  - [ ] 2.1 Create `config/vap/vco-immutable-path.yaml` — ValidatingAdmissionPolicy with CEL: `object.spec.path == oldObject.spec.path`
  - [ ] 2.2 Create `config/vap/vco-immutable-path-binding.yaml` — ValidatingAdmissionPolicyBinding listing all 72 CRD resources
  - [ ] 2.3 Verify the binding `matchResources` covers all 72 CRDs in `redhatcop.redhat.io` group
- [ ] Task 3 — Create `vco-immutable-name` shared VAP (AC: #5, #6, #7)
  - [ ] 3.1 Create `config/vap/vco-immutable-name.yaml` — ValidatingAdmissionPolicy with CEL: `!has(oldObject.spec.name) && !has(object.spec.name) || has(oldObject.spec.name) && has(object.spec.name) && object.spec.name == oldObject.spec.name`
  - [ ] 3.2 Create `config/vap/vco-immutable-name-binding.yaml` — ValidatingAdmissionPolicyBinding listing all 65 CRD resources
- [ ] Task 4 — Create `vco-target-namespace-guard` parameterized VAP (AC: #8, #9, #10)
  - [ ] 4.1 Create `config/vap/vco-target-namespace-guard.yaml` — ValidatingAdmissionPolicy with `spec.paramKind` referencing ConfigMap, CEL replicating `ValidateTargetNamespaceUpdate` logic
  - [ ] 4.2 Create `config/vap/vco-target-namespace-guard-binding.yaml` — ValidatingAdmissionPolicyBinding listing all 92 CRDs
  - [ ] 4.3 Create `config/vap/vco-vap-params-configmap.yaml` — ConfigMap stub for `VAULT_NAMESPACE` value
- [ ] Task 5 — Configure envtest for VAP support (AC: #2)
  - [ ] 5.1 Update `api/v1alpha1/webhook_suite_test.go` — add `KubeAPIServerFlags: []string{"--enable-admission-plugins=ValidatingAdmissionPolicy"}` to `testEnv`
  - [ ] 5.2 Add VAP manifest directory path to envtest: load `config/vap/` alongside CRDs
  - [ ] 5.3 Verify envtest starts with VAP admission plugin enabled
- [ ] Task 6 — Remove migrated Go validation from webhook files (AC: #11)
  - [ ] 6.1 Remove `spec.path` immutability checks from all 72 webhook `ValidateUpdate` methods (the `if newObj.Spec.Path != oldObj.Spec.Path` blocks)
  - [ ] 6.2 Remove `spec.name` immutability checks from all 34 webhook `ValidateUpdate` methods that currently have them
  - [ ] 6.3 Remove `ValidateTargetNamespaceUpdate()` calls from all 91 webhook `ValidateUpdate` methods (VaultSecret has it too)
  - [ ] 6.4 Remove the `ValidateTargetNamespaceUpdate` method from `api/v1alpha1/utils/commons.go` (and `GetTargetNamespace`, `getLoginNamespace`, `displayNamespace` if no longer referenced by non-webhook code)
  - [ ] 6.5 Clean up unused imports (`errors`, `fmt`) from webhook files after removal
  - [ ] 6.6 **CAUTION:** Do NOT remove any `isValid()` calls, credential validation calls, or CRD-specific validation logic — those are Story 18.2 scope
  - [ ] 6.7 **CAUTION:** Do NOT remove `ValidateCreate`/`ValidateDelete` method bodies — only strip the migrated checks from `ValidateUpdate`
  - [ ] 6.8 Keep all `admission.Defaulter[*T]` implementations untouched
- [ ] Task 7 — Update existing webhook tests (AC: #11)
  - [ ] 7.1 Update `api/v1alpha1/webhook_validate_update_test.go` — path/name immutability rejection tests should now expect VAP-level rejection (or be relocated to a new VAP-specific test file)
  - [ ] 7.2 Verify non-migrated webhook test cases still pass (CRD-specific validations in `transitsecretenginekey_webhook_test.go` etc.)
- [ ] Task 8 — Code generation and validation
  - [ ] 8.1 `make manifests generate fmt vet test` passes
  - [ ] 8.2 `make lint` passes (errcheck, staticcheck, all standard linters)
  - [ ] 8.3 Verify no existing tests regress

## Dev Notes

### Overview

This is the **first VAP in the project** — it introduces a new admission control layer using Kubernetes `ValidatingAdmissionPolicy` (GA since K8s 1.30). The story creates three shared VAPs that replace the repetitive Go validation code scattered across 92 webhook files. It also fixes 35 validation gaps discovered during analysis.

**Novelty Risk:** HIGH — First VAP in the project, new CEL patterns, parameterized VAP setup, envtest VAP configuration.

### Part A — VAP Infrastructure (`config/vap/`)

**Directory structure:**

```
config/vap/
├── kustomization.yaml
├── vco-immutable-path.yaml              # ValidatingAdmissionPolicy
├── vco-immutable-path-binding.yaml      # ValidatingAdmissionPolicyBinding
├── vco-immutable-name.yaml              # ValidatingAdmissionPolicy
├── vco-immutable-name-binding.yaml      # ValidatingAdmissionPolicyBinding
├── vco-target-namespace-guard.yaml      # ValidatingAdmissionPolicy (parameterized)
├── vco-target-namespace-guard-binding.yaml  # ValidatingAdmissionPolicyBinding
└── vco-vap-params-configmap.yaml        # ConfigMap for VAULT_NAMESPACE fallback
```

**Kustomize integration:**
- `config/vap/kustomization.yaml` lists all YAML files above as `resources:`
- `config/default/kustomization.yaml` adds `- ../vap` to its `resources:` list (add after `- ../webhook`)
- The Helm chart build (`make helmchart`) should pick this up via the kustomize overlay

**VAP API version:** Use `apiVersion: admissionregistration.k8s.io/v1` (GA since K8s 1.30, our envtest K8s assets are 1.36.x).

### Part B — `vco-immutable-path` VAP

**CEL expression:**
```cel
object.spec.path == oldObject.spec.path
```

**Failure message:** `"spec.path cannot be updated"`

**matchConstraints:** `operations: ["UPDATE"]`, `resourceRules` listing all 72 CRDs in `redhatcop.redhat.io` API group.

**72 CRDs with `spec.path`** (verified from grep of `spec.path cannot be updated` in webhook files — the 68 existing + 4 gaps):

The binding's `matchResources.resourceRules` should list all 72 resource names (plural form as used in CRD definitions) under `apiGroups: ["redhatcop.redhat.io"]`, `apiVersions: ["v1alpha1"]`.

**4 CRDs currently MISSING path immutability checks** (confirmed by inspecting their webhook files — they have `ValidateTargetNamespaceUpdate` but NO `spec.path` check):
1. `AzureAuthEngineRole` — `azureauthenginerole_webhook.go` line 61-68: only `ValidateTargetNamespaceUpdate`, no path check
2. `GCPAuthEngineRole` — `gcpauthenginerole_webhook.go` line 59-66: only `ValidateTargetNamespaceUpdate`, no path check
3. `JWTOIDCAuthEngineRole` — `jwtoidcauthenginerole_webhook.go` line 64-73: has `ValidateTargetNamespaceUpdate` + `validateAllowedRedirectURIs`, no path check
4. `LDAPAuthEngineGroup` — `ldapauthenginegroup_webhook.go` line 59-66: only `ValidateTargetNamespaceUpdate`, no path check

These 4 gaps are **automatically fixed** by the shared VAP — they are included in the 72-CRD binding.

### Part C — `vco-immutable-name` VAP

**CEL expression:**
```cel
!has(oldObject.spec.name) && !has(object.spec.name) || has(oldObject.spec.name) && has(object.spec.name) && object.spec.name == oldObject.spec.name
```

This handles three cases:
1. `spec.name` was never set on old, never set on new → ALLOW (both use `metadata.name` fallback)
2. `spec.name` was set on old, same value on new → ALLOW
3. `spec.name` was set on old but changed, OR added/removed → REJECT

**Failure message:** `"spec.name cannot be updated"`

**matchConstraints:** `operations: ["UPDATE"]`, `resourceRules` listing all 65 CRDs.

**Critical nuance about `has()` vs empty string:**
- In CRD schema, `spec.name` is a `string` field with `omitempty` JSON tag. When not set by the user, it won't appear in the stored object (server-side). `has(object.spec.name)` returns `false` when the field is absent from the stored JSON.
- When the user sets `spec.name: ""` explicitly, the omitempty tag means it's serialized as absent — so `has()` returns `false` for empty string too. This is the correct behavior: empty string is equivalent to "not set" (uses metadata.name fallback via `ResourceName()`).

**31 CRDs with `spec.name` MISSING immutability checks** (from epic analysis):
- These are the types that have `Name` in their spec and use it in `GetPath()` via `ResourceName(spec.Name, metadata.Name)` or direct `spec.Name` reference, but their webhook `ValidateUpdate` does NOT check `spec.name` changes.
- All 31 are automatically fixed by inclusion in the 65-CRD binding.

### Part D — `vco-target-namespace-guard` VAP (Parameterized)

**Logic to replicate** (from `api/v1alpha1/utils/commons.go` lines 237-244):
```go
func (kc *KubeAuthConfiguration) ValidateTargetNamespaceUpdate(old *KubeAuthConfiguration) error {
    oldNamespace, newNamespace := old.GetTargetNamespace(), kc.GetTargetNamespace()
    if (kc.TargetNamespace != "" || old.TargetNamespace != "") && newNamespace != oldNamespace {
        return fmt.Errorf("spec.authentication cannot be updated to a different Vault namespace ...")
    }
    return nil
}
```

Where `GetTargetNamespace()` = `JoinNamespace(getLoginNamespace(), TargetNamespace)` and `getLoginNamespace()` = `Namespace || VAULT_NAMESPACE env var`.

**CEL expression** (must replicate `JoinNamespace` + `CleansePath` logic):

```cel
// Compute effective old and new target namespaces
// getLoginNamespace: use spec.authentication.namespace, fallback to params.data.vaultNamespace
// JoinNamespace: CleansePath(CleansePath(parent) + "/" + CleansePath(child))
// CleansePath: trim leading/trailing slashes, collapse double slashes

// Guard condition: only enforce when targetNamespace is set on either old or new
!(oldObject.spec.authentication.targetNamespace != "" || object.spec.authentication.targetNamespace != "") ||
// If guard is active, compare effective namespaces
(
  // Helper: build effective namespace for old
  // oldLogin = old.namespace or params fallback
  // oldTarget = JoinNamespace(oldLogin, old.targetNamespace)
  // Same for new
  // Compare: oldTarget == newTarget
)
```

**⚠️ CEL Complexity Warning:** The `JoinNamespace` → `CleansePath` chain involves string manipulation (trim slashes, collapse `//` to `/`). CEL has `strings.replace()`, `trim()` (via `trimPrefix`/`trimSuffix`), and concatenation. The dev must test this carefully with edge cases:
- Empty `namespace` with `VAULT_NAMESPACE` fallback
- Namespaces with leading/trailing slashes
- Nested namespaces like `org/team/tenant`

**Parameterized VAP setup:**
- `spec.paramKind` on the VAP references `apiVersion: v1`, `kind: ConfigMap`
- The binding references the ConfigMap via `spec.paramRef: { name: "vco-vap-params", namespace: "<operator-namespace>" }`
- The ConfigMap has `data.vaultNamespace` key holding the `VAULT_NAMESPACE` env var value
- The operator deployment must create/update this ConfigMap on startup (or via kustomize)

**92 CRDs with `spec.authentication`** — this is every CRD type, including VaultSecret. All are listed in the binding.

### Part E — envtest Configuration for VAP

**Current envtest setup** (`api/v1alpha1/webhook_suite_test.go` lines 73-85):
```go
testEnv = &envtest.Environment{
    CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
    ErrorIfCRDPathMissing: false,
    BinaryAssetsDirectory: binaryAssetsDir,
    WebhookInstallOptions: envtest.WebhookInstallOptions{
        Paths: []string{filepath.Join("..", "..", "config", "webhook")},
    },
}
```

**Required changes:**
1. Add `KubeAPIServerFlags` to enable the VAP admission plugin:
   ```go
   testEnv = &envtest.Environment{
       CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
       ErrorIfCRDPathMissing: false,
       BinaryAssetsDirectory: binaryAssetsDir,
       KubeAPIServerFlags: append(
           envtest.DefaultKubeAPIServerFlags,
           "--enable-admission-plugins=ValidatingAdmissionPolicy",
       ),
       WebhookInstallOptions: envtest.WebhookInstallOptions{
           Paths: []string{filepath.Join("..", "..", "config", "webhook")},
       },
   }
   ```
2. After `testEnv.Start()`, load VAP manifests by applying them via the k8s client (they're not CRDs, so `CRDDirectoryPaths` won't work). Read YAML files from `config/vap/` and apply them using `k8sClient.Create()`.
3. Create the `vco-vap-params` ConfigMap before VAP tests run.

**⚠️ envtest VAP readiness:** envtest in controller-runtime v0.24.x with K8s 1.36 assets should support VAPs natively (GA since 1.30). However, there may be a propagation delay between creating the VAP resources and enforcement. Use `Eventually()` to wait for VAP enforcement to be active before running rejection tests.

### Part F — Webhook Code Removal

**Scope:** Remove ONLY the three migrated validation checks from `ValidateUpdate` methods:
1. `spec.path` immutability check (`if newObj.Spec.Path != oldObj.Spec.Path`)
2. `spec.name` immutability check (`if newObj.Spec.Name != oldObj.Spec.Name`)
3. `ValidateTargetNamespaceUpdate()` call

**DO NOT remove:**
- `isValid()` calls (Story 18.2 scope)
- Credential validation calls (`credentials.ValidateCredentialSource()`)
- CRD-specific immutability checks (e.g., AppRoleAuthEngineRole's `localSecretIDs`, TransitSecretEngineKey's 5+ immutable fields)
- Any `ValidateCreate` logic
- Any `Default()` method logic

**Example transformation** for a typical webhook (e.g., `approleauthenginerole_webhook.go`):

Before:
```go
func (r *AppRoleAuthEngineRole) ValidateUpdate(ctx context.Context, oldObj, newObj *AppRoleAuthEngineRole) (admission.Warnings, error) {
    approleauthenginerolelog.Info("validate update", "name", newObj.Name)
    if err := newObj.Spec.Authentication.ValidateTargetNamespaceUpdate(&oldObj.Spec.Authentication); err != nil {
        return nil, err
    }
    if newObj.Spec.Path != oldObj.Spec.Path {
        return nil, errors.New("spec.path cannot be updated")
    }
    if newObj.Spec.Name != oldObj.Spec.Name {
        return nil, errors.New("spec.name cannot be updated")
    }
    if newObj.Spec.LocalSecretIDs != oldObj.Spec.LocalSecretIDs {
        return nil, errors.New("spec.localSecretIDs cannot be updated")
    }
    return nil, nil
}
```

After:
```go
func (r *AppRoleAuthEngineRole) ValidateUpdate(ctx context.Context, oldObj, newObj *AppRoleAuthEngineRole) (admission.Warnings, error) {
    approleauthenginerolelog.Info("validate update", "name", newObj.Name)
    if newObj.Spec.LocalSecretIDs != oldObj.Spec.LocalSecretIDs {
        return nil, errors.New("spec.localSecretIDs cannot be updated")
    }
    return nil, nil
}
```

**Webhooks that become no-op after removal** (only had path/name/target-ns checks): these `ValidateUpdate` methods will become just the log line + `return nil, nil`. This is expected and will be cleaned up in Story 18.3.

**Utility function cleanup:**
- `ValidateTargetNamespaceUpdate` in `api/v1alpha1/utils/commons.go` — check if any non-webhook code references it. If not, remove.
- `GetTargetNamespace` — used by `ValidateTargetNamespaceUpdate` and possibly by other code (reconciler). **Check usages before removing.** If used elsewhere, keep it.
- `getLoginNamespace`, `displayNamespace` — internal helpers, safe to remove if `ValidateTargetNamespaceUpdate` is removed.

### Part G — Existing Test Impact

**`api/v1alpha1/webhook_validate_update_test.go`** (~1300 lines):
- Contains table-driven tests for path/name immutability rejection across many CRD types
- These tests call `ValidateUpdate` directly on Go structs — after VAP migration, the Go `ValidateUpdate` no longer rejects path/name changes
- **Decision needed:** Either:
  - (A) Move these test cases to a new VAP test file that tests via envtest API server (preferred — proves VAPs work)
  - (B) Delete these test cases and rely on Story 18.4's comprehensive VAP test suite
  - **Recommendation:** Option B — delete from webhook test, let Story 18.4 add comprehensive VAP tests. This story should just ensure existing non-migrated tests still pass.

### Architecture Compliance

| Requirement | Compliance |
|-------------|------------|
| VAP manifests in `config/vap/` | ✅ New directory under existing kustomize structure |
| Kustomize overlay integration | ✅ Added to `config/default/kustomization.yaml` |
| envtest VAP support | ✅ K8s 1.36 assets, `--enable-admission-plugins=ValidatingAdmissionPolicy` |
| Shared framework protection | ✅ No behavioral changes to shared framework files |
| Error management pattern | ✅ N/A — removing Go validation, not adding |
| CRD registration checklist | ✅ No new CRDs in this story |

### File Structure Requirements

**New files:**
- `config/vap/kustomization.yaml`
- `config/vap/vco-immutable-path.yaml`
- `config/vap/vco-immutable-path-binding.yaml`
- `config/vap/vco-immutable-name.yaml`
- `config/vap/vco-immutable-name-binding.yaml`
- `config/vap/vco-target-namespace-guard.yaml`
- `config/vap/vco-target-namespace-guard-binding.yaml`
- `config/vap/vco-vap-params-configmap.yaml`

**Modified files (high-level — 92 webhook files + infra):**
- `config/default/kustomization.yaml` — add `- ../vap` resource
- `api/v1alpha1/webhook_suite_test.go` — add `KubeAPIServerFlags` for VAP
- `api/v1alpha1/webhook_validate_update_test.go` — remove/update path/name/target-ns test cases
- `api/v1alpha1/*_webhook.go` (up to 92 files) — remove migrated validation checks from `ValidateUpdate`
- `api/v1alpha1/utils/commons.go` — potentially remove `ValidateTargetNamespaceUpdate` and helpers

### Testing Strategy

- **Unit-tier (envtest):** Verify VAPs are enforced by the API server in `webhook_suite_test.go` or a new test file alongside it. Create a CRD instance, attempt an update changing `spec.path` → expect rejection.
- **This story does NOT require full VAP test coverage** — that is Story 18.4's scope. A minimal smoke test proving the VAP admission plugin is active and enforcing is sufficient.
- **Regression:** `make test` must pass. All non-migrated webhook tests must continue to work.

### Critical Implementation Order

1. **First:** Create all VAP manifests and kustomize overlay (Tasks 1-4)
2. **Second:** Configure envtest and verify VAPs are enforced (Task 5)
3. **Third:** Verify VAP enforcement works with a smoke test
4. **Fourth:** Remove Go validation code from webhook files (Task 6)
5. **Fifth:** Update webhook tests (Task 7)
6. **Last:** Run `make manifests generate fmt vet test lint` (Task 8)

**Rationale:** VAPs must be provably working BEFORE removing the Go validation code. If VAPs fail to enforce for any reason, the Go code is still the safety net.

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic 18 — Story 18.1]
- [Source: api/v1alpha1/utils/commons.go#ValidateTargetNamespaceUpdate — lines 237-244]
- [Source: api/v1alpha1/utils/commons.go#KubeAuthConfiguration — lines 52-75]
- [Source: api/v1alpha1/webhook_suite_test.go — envtest setup lines 73-85]
- [Source: api/v1alpha1/approleauthenginerole_webhook.go — typical webhook with all 3 checks]
- [Source: api/v1alpha1/azureauthenginerole_webhook.go — webhook MISSING path immutability]
- [Source: api/v1alpha1/gcpauthenginerole_webhook.go — webhook MISSING path immutability]
- [Source: api/v1alpha1/jwtoidcauthenginerole_webhook.go — webhook MISSING path immutability]
- [Source: api/v1alpha1/ldapauthenginegroup_webhook.go — webhook MISSING path immutability]
- [Source: config/default/kustomization.yaml — kustomize overlay structure]
- [Kubernetes docs: ValidatingAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
- [CEL spec: Kubernetes CEL expressions](https://kubernetes.io/docs/reference/using-api/cel/)

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
