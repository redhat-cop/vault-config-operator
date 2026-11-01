# Story 18.4: VAP Test Suite — envtest + Integration Coverage

Status: ready-for-dev

## Story

As an operator developer,
I want comprehensive test coverage for all VAP rules in both envtest (unit tier) and Kind (integration tier),
So that VAP CEL expressions are regression-tested and validated against the real Kubernetes API server.

## Acceptance Criteria

1. **AC-1 — envtest VAP admission plugin enabled:** `api/v1alpha1/webhook_suite_test.go` configures `testEnv` with `KubeAPIServerFlags` including `--enable-admission-plugins=ValidatingAdmissionPolicy` and loads all VAP manifests from `config/vap/` after the environment starts.
2. **AC-2 — envtest immutable-path rejection:** A table-driven test creates a CRD resource, then attempts an UPDATE changing `spec.path` → the API server rejects the update. Covers a representative sample of the 72 path-bound CRDs, including all 4 previously-missing gap CRDs (AzureAuthEngineRole, GCPAuthEngineRole, JWTOIDCAuthEngineRole, LDAPAuthEngineGroup).
3. **AC-3 — envtest immutable-name rejection:** A table-driven test creates a CRD resource, then attempts an UPDATE changing `spec.name` → the API server rejects the update. Covers a representative sample of the 65 name-bound CRDs, including a subset of the 31 previously-missing gap CRDs.
4. **AC-4 — envtest target-namespace-guard rejection:** A table-driven test creates a CRD resource with `spec.authentication.targetNamespace` set, then attempts an UPDATE changing the effective Vault namespace → the API server rejects the update. Also verifies that changes are allowed when `targetNamespace` is unset on both old and new.
5. **AC-5 — envtest credential source VAP rejection:** Tests verify that creating or updating a resource with zero or >1 credential sources set is rejected (rootCredentials, bindCredentials, and single-CRD credential groups).
6. **AC-6 — envtest CRD-specific VAP rejection:** For each CRD-specific VAP from Story 18.2 (TransitSecretEngineKey, AWSSecretEngineRole, RandomSecret, mounts, etc.), at least one invalid CREATE and one invalid UPDATE are tested and rejected.
7. **AC-7 — envtest VAP allowance (positive tests):** For each shared VAP, at least one test verifies that valid mutations ARE allowed (same path, same name, no namespace change). For credential source VAPs, verify that exactly-one-source passes.
8. **AC-8 — gap fix coverage:** All 4 previously-missing path-immutability CRDs and a representative subset (≥10) of the 31 previously-missing name-immutability CRDs are explicitly tested.
9. **AC-9 — Kind integration VAP enforcement:** `controllers/vap_integration_test.go` verifies at least one shared VAP (immutable-path) and one CRD-specific VAP are enforced in the Kind cluster during `make integration`.
10. **AC-10 — existing webhook tests adapted:** `api/v1alpha1/webhook_validate_update_test.go` is deleted or stripped to only retain test cases for residual webhook logic (RabbitMQ cross-resource, CRD-specific immutability checks retained in Go). The webhook suite file (`webhook_suite_test.go`) is kept.
11. **AC-11 — build passes cleanly:** `make manifests generate fmt vet test lint` passes with zero errors.

## Tasks / Subtasks

- [ ] Task 1: Configure envtest for VAP support (AC: #1)
  - [ ] 1.1 — In `api/v1alpha1/webhook_suite_test.go`, add `KubeAPIServerFlags` to `testEnv` to enable the `ValidatingAdmissionPolicy` admission plugin (append to `envtest.DefaultKubeAPIServerFlags`)
  - [ ] 1.2 — After `testEnv.Start()`, load all VAP manifests from `config/vap/` by reading YAML files and applying them via `k8sClient.Create()` (VAPs are not CRDs, so `CRDDirectoryPaths` won't work)
  - [ ] 1.3 — Create the `vco-vap-params` ConfigMap in the test namespace before VAP tests run (required by the parameterized target-namespace-guard VAP)
  - [ ] 1.4 — Add `Eventually()` delay after VAP resource creation to ensure the API server has registered and is enforcing the policies before test assertions
- [ ] Task 2: Create envtest VAP test file for shared VAPs (AC: #2, #3, #4, #7, #8)
  - [ ] 2.1 — Create `api/v1alpha1/vap_test.go` (build tag: `//go:build !integration`)
  - [ ] 2.2 — Table-driven `TestVAPImmutablePath` — for each CRD in the sample set: create a valid resource, attempt to update `spec.path`, assert rejection with "spec.path cannot be updated". Then test that updating a non-path field is allowed.
  - [ ] 2.3 — Table-driven `TestVAPImmutableName` — for each CRD in the sample set: create a valid resource, attempt to update `spec.name`, assert rejection. Test that updates without name change are allowed. Test the `has()` edge case: resource created without `spec.name`, then update adding `spec.name` → rejected.
  - [ ] 2.4 — Table-driven `TestVAPTargetNamespaceGuard` — create resource with `targetNamespace` set + `namespace` set, then change `namespace` on update → rejected. Verify that change is allowed when `targetNamespace` is empty on both old and new.
  - [ ] 2.5 — Include ALL 4 gap CRDs in path tests: AzureAuthEngineRole, GCPAuthEngineRole, JWTOIDCAuthEngineRole, LDAPAuthEngineGroup
  - [ ] 2.6 — Include ≥10 previously-missing name-immutability gap CRDs in name tests
- [ ] Task 3: Create envtest VAP test file for credential source + CRD-specific VAPs (AC: #5, #6, #7)
  - [ ] 3.1 — Add credential source tests to `vap_test.go` (or a separate `vap_credential_test.go`): test rootCredentials with 0 sources, 2 sources, and exactly 1 source
  - [ ] 3.2 — Add CRD-specific VAP tests for TransitSecretEngineKey: convergent without derived → rejected, keySize on non-hmac → rejected, autoRotatePeriod < 1h → rejected, type change on update → rejected, exportable true→false → rejected, valid create → allowed
  - [ ] 3.3 — Add CRD-specific VAP tests for mount types: AuthEngineMount update changing `type` → rejected, update changing only `config` → allowed
  - [ ] 3.4 — Add CRD-specific VAP tests for at least: AWSSecretEngineRole (credentialType restrictions), RandomSecret (secretKey immutability), DatabaseSecretEngineStaticRole (rotation mutual exclusivity), PKISecretEngineConfig (ACME/cluster), KubernetesAuthEngineRole (targetNamespaceSelector vs targetNamespaces)
- [ ] Task 4: Create Kind integration test for VAP enforcement (AC: #9)
  - [ ] 4.1 — Create `internal/controller/vap_integration_test.go` (build tag: `//go:build integration`)
  - [ ] 4.2 — Test immutable-path VAP: create a Policy resource, attempt path update → rejected by Kind cluster's API server
  - [ ] 4.3 — Test at least one CRD-specific VAP: create a TransitSecretEngineKey or mount with invalid fields → rejected
  - [ ] 4.4 — Verify VAPs are deployed: list `ValidatingAdmissionPolicy` resources in the Kind cluster, assert expected count
- [ ] Task 5: Adapt existing webhook test files (AC: #10)
  - [ ] 5.1 — Delete `api/v1alpha1/webhook_validate_update_test.go` — this file tests Go-level webhook validation for path/name/target-ns that is now replaced by VAPs. VAP tests in Task 2 provide equivalent coverage.
  - [ ] 5.2 — Keep `api/v1alpha1/webhook_suite_test.go` — still needed for residual webhook tests and as the envtest harness for VAP tests
  - [ ] 5.3 — Keep `api/v1alpha1/transitsecretenginekey_webhook_test.go` — tests CRD-specific Go validation that MAY still be in the webhook (if any residual validation remains after 18.2/18.3 VAP migration). If Transit is fully VAP-covered, adapt to test via API server.
  - [ ] 5.4 — Keep `api/v1alpha1/rabbitmqsecretengineconfig_webhook_test.go` — tests cross-resource `validateUniquePath` which remains as a Go webhook
  - [ ] 5.5 — Keep `api/v1alpha1/namespace_webhook_test.go` and `targetnamespace_webhook_test.go` — evaluate if these test VAP-migrated logic. If so, migrate test cases to `vap_test.go` and delete originals.
- [ ] Task 6: Build validation (AC: #11)
  - [ ] 6.1 — Run `make manifests generate fmt vet`
  - [ ] 6.2 — Run `make test` (envtest unit tests including new VAP tests)
  - [ ] 6.3 — Run `make lint`
  - [ ] 6.4 — Verify `make integration` passes (if Kind cluster available)

## Dev Notes

### Dependency: Stories 18.1, 18.2, and 18.3 MUST be done first

This story tests the VAP manifests created in 18.1 and 18.2, and assumes the webhook code cleanup from 18.3 is already complete. The `config/vap/` directory must exist and contain all VAP YAML files. If it does not exist, **stop and report** that prerequisite stories must be completed first.

Verify before starting:
- `config/vap/` exists with `kustomization.yaml` and all VAP/binding YAMLs
- `webhook_validate_update_test.go` should still exist (this story is responsible for deleting it)
- `make test` passes in the current state

### Novelty Risk: MEDIUM

envtest VAP testing is a newer pattern. Key risks:
1. **VAP propagation delay:** After creating `ValidatingAdmissionPolicy` + binding via `k8sClient.Create()`, there may be a delay before the API server enforces. Use `Eventually()` with retries.
2. **envtest K8s version:** Makefile sets `ENVTEST_K8S_VERSION = 1.37.0` — VAPs are GA since K8s 1.30, so no version concerns.
3. **Parameterized VAP in envtest:** The `vco-target-namespace-guard` uses a ConfigMap param. The ConfigMap must be created before the VAP binding can take effect.
4. **YAML loading in envtest:** VAP manifests are Kubernetes resources (not CRDs), so they can't be loaded via `CRDDirectoryPaths`. Must be loaded post-startup via the k8s client.

### envtest VAP Setup — Detailed Implementation

**In `webhook_suite_test.go` `BeforeSuite`:**

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

**After `testEnv.Start()` and `k8sClient` initialization, load VAP manifests:**

```go
// Load and apply VAP manifests
vapDir := filepath.Join("..", "..", "config", "vap")
vapFiles, err := filepath.Glob(filepath.Join(vapDir, "*.yaml"))
Expect(err).NotTo(HaveOccurred())
Expect(vapFiles).NotTo(BeEmpty(), "no VAP manifests found in config/vap/")

for _, vapFile := range vapFiles {
    data, err := os.ReadFile(vapFile)
    Expect(err).NotTo(HaveOccurred())
    // Split multi-document YAML and decode+create each resource
    // Use apimachinery unstructured.Unstructured to handle
    // both ValidatingAdmissionPolicy and ValidatingAdmissionPolicyBinding
}
```

**⚠️ YAML decoding approach:** Use `k8s.io/apimachinery/pkg/apis/meta/v1/unstructured` and `k8s.io/apimachinery/pkg/util/yaml` to decode VAP manifests into `unstructured.Unstructured` objects, then create them via `k8sClient.Create()`. This avoids importing the entire `admissionregistration/v1` typed API (though importing it is also fine — it's already a transitive dependency). The scheme must have `admissionregistration.k8s.io/v1` registered for typed decoding; alternatively, use dynamic client.

**ConfigMap for parameterized VAP:**

```go
// Create the vco-vap-params ConfigMap needed by vco-target-namespace-guard
vapParamsConfigMap := &corev1.ConfigMap{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "vco-vap-params",
        Namespace: "default",  // envtest uses "default" unless namespaced otherwise
    },
    Data: map[string]string{
        "vaultNamespace": "",  // empty fallback for tests
    },
}
err = k8sClient.Create(ctx, vapParamsConfigMap)
Expect(err).NotTo(HaveOccurred())
```

**VAP readiness wait:**

```go
// Wait for VAPs to be enforced — the API server needs time to compile
// and activate CEL expressions
time.Sleep(2 * time.Second) // or use Eventually with a probe
```

Prefer an active probe over sleep: attempt a known-invalid operation and wait for rejection:

```go
Eventually(func() bool {
    // Try to create+update a Policy with changed path
    // If the update is rejected, VAPs are active
    // If it's allowed, they're not ready yet
    ...
}, "30s", "1s").Should(BeTrue(), "VAPs not enforced after 30s")
```

### Test File Structure

**New files:**
| File | Build Tag | Purpose |
|------|-----------|---------|
| `api/v1alpha1/vap_test.go` | `//go:build !integration` | Shared VAP envtest tests (path, name, target-ns, credential source, CRD-specific) |
| `internal/controller/vap_integration_test.go` | `//go:build integration` | Kind cluster VAP enforcement tests |

**Deleted files:**
| File | Reason |
|------|--------|
| `api/v1alpha1/webhook_validate_update_test.go` | Tests Go-level webhook validation now replaced by VAPs. VAP tests provide equivalent coverage. |

**Retained files:**
| File | Reason |
|------|--------|
| `api/v1alpha1/webhook_suite_test.go` | envtest harness — updated with VAP support |
| `api/v1alpha1/transitsecretenginekey_webhook_test.go` | Tests CRD-specific webhook logic if any remains |
| `api/v1alpha1/rabbitmqsecretengineconfig_webhook_test.go` | Tests residual cross-resource webhook |
| `api/v1alpha1/namespace_webhook_test.go` | Evaluate — may test VAP-migrated logic |
| `api/v1alpha1/targetnamespace_webhook_test.go` | Evaluate — may test VAP-migrated logic |

### envtest Test Patterns — `vap_test.go`

**Table-driven pattern for shared VAPs:**

```go
var _ = Describe("VAP Immutable Path", func() {
    type pathTestCase struct {
        name     string
        obj      client.Object
        mutate   func(obj client.Object) // change spec.path
    }

    entries := []pathTestCase{
        {
            name: "Policy",
            obj: &Policy{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-policy",
                    Namespace: "default",
                },
                Spec: PolicySpec{
                    Path: "sys/policy/test",
                    Authentication: vaultutils.KubeAuthConfiguration{...},
                },
            },
            mutate: func(obj client.Object) {
                obj.(*Policy).Spec.Path = "sys/policy/changed"
            },
        },
        // ... more CRDs ...
    }

    for _, tc := range entries {
        tc := tc
        It(fmt.Sprintf("should reject path change for %s", tc.name), func() {
            // Create the resource
            Expect(k8sClient.Create(ctx, tc.obj)).To(Succeed())
            // Fetch the latest version (resourceVersion needed for update)
            Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(tc.obj), tc.obj)).To(Succeed())
            // Mutate path
            tc.mutate(tc.obj)
            // Attempt update — expect rejection
            err := k8sClient.Update(ctx, tc.obj)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(ContainSubstring("spec.path cannot be updated"))
        })
    }
})
```

**Important test resource considerations:**
- Each test resource needs a unique `metadata.name` to avoid conflicts between test cases
- Resources are namespace-scoped — use `"default"` namespace in envtest
- Resources need a minimal valid `spec.authentication` for CRDs that require it (all CRDs in this project have `KubeAuthConfiguration`)
- Clean up created resources in `AfterEach` or use unique names with `GenerateName`
- The API server validates required fields — provide minimum required fields for each CRD type

### Representative CRD Sample Sets

**Path immutability test sample (≥15 CRDs including all 4 gaps):**
1. Policy (simplest, good smoke test)
2. PasswordPolicy
3. AzureAuthEngineRole (**gap fix**)
4. GCPAuthEngineRole (**gap fix**)
5. JWTOIDCAuthEngineRole (**gap fix**)
6. LDAPAuthEngineGroup (**gap fix**)
7. AppRoleAuthEngineRole
8. AWSSecretEngineRole
9. DatabaseSecretEngineConfig
10. TransitSecretEngineKey
11. KubernetesAuthEngineRole
12. SSHSecretEngineConfig
13. ConsulSecretEngineRole
14. NomadSecretEngineRole
15. PKISecretEngineRole

**Name immutability test sample (≥12 CRDs including ≥10 from gaps):**
1. QuotaConfig (gap fix — new in Epic 17)
2. RateLimitQuota (gap fix — new in Epic 17)
3. LeaseCountQuota (gap fix — new in Epic 17)
4. CORSConfig (gap fix — new in Epic 17)
5. KeyRotationConfig (gap fix — new in Epic 17)
6. RotationPolicy (gap fix — new in Epic 17)
7. UICustomHeader (gap fix — new in Epic 17)
8. LoggerConfig (gap fix — new in Epic 17)
9. AWSSecretEngineConfig (gap fix)
10. ConsulSecretEngineConfig (gap fix)
11. AppRoleAuthEngineRole (already had name check — regression test)
12. KubernetesAuthEngineRole (already had name check — regression test)

### Kind Integration Tests — `vap_integration_test.go`

The integration test suite uses `USE_EXISTING_CLUSTER=true` with a Kind cluster. VAPs deployed via kustomize are available in the cluster.

```go
//go:build integration

package controller

var _ = Describe("VAP Integration", func() {
    Context("immutable-path enforcement in Kind", func() {
        It("should reject spec.path change for Policy", func() {
            policy := &redhatcopv1alpha1.Policy{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "vap-int-test-policy",
                    Namespace: vaultTestNamespaceName,
                },
                Spec: redhatcopv1alpha1.PolicySpec{
                    Path: "sys/policy/vaptest",
                    Authentication: validKubeAuth(),
                    PolicyDocument: "path \"secret/*\" { capabilities = [\"read\"] }",
                },
            }
            Expect(k8sIntegrationClient.Create(ctx, policy)).To(Succeed())
            // ... fetch, mutate path, attempt update, expect rejection ...
        })
    })

    Context("VAP manifests deployed", func() {
        It("should have expected ValidatingAdmissionPolicy count", func() {
            // List all VAPs in cluster
            // Assert count matches expected (3 shared + credential + CRD-specific)
        })
    })
})
```

**Integration test deployment prerequisite:** VAP manifests must be deployed to the Kind cluster before integration tests run. This happens via `make deploy` which uses the kustomize overlay that includes `config/vap/`. Verify the `make integration` target's deploy step includes VAP manifests.

### Handling webhook_validate_update_test.go Deletion

This file (~1,315 lines) tests Go-level `ValidateUpdate` for path/name immutability and target-namespace-update across many CRDs. After Stories 18.1–18.3:
- The Go `ValidateUpdate` methods no longer check path/name immutability (removed in 18.1) or `isValid()` (removed in 18.2/18.3)
- These tests would FAIL because the Go methods now return `nil` for previously-rejected mutations
- The new `vap_test.go` tests cover the same scenarios via the API server's VAP enforcement

**Decision: DELETE the file.** The test coverage is replaced by `vap_test.go`.

Also evaluate `namespace_webhook_test.go` and `targetnamespace_webhook_test.go`:
- If they test `ValidateTargetNamespaceUpdate()` (which is VAP-migrated), their test cases should be moved to `vap_test.go` and the original files deleted
- If they test other namespace-related logic that remains in Go, keep them

### Critical Test Authoring Patterns

1. **Use `GenerateName` or unique suffixes** to avoid test name collisions:
   ```go
   ObjectMeta: metav1.ObjectMeta{
       GenerateName: "vap-path-test-",
       Namespace:    "default",
   }
   ```

2. **Minimal valid resource spec** — many CRDs require `spec.authentication` with `path` and `serviceAccount`:
   ```go
   Authentication: vaultutils.KubeAuthConfiguration{
       Path:              "kubernetes",
       Role:              "test-role",
       ServiceAccountRef: &corev1.LocalObjectReference{Name: "default"},
   }
   ```

3. **Expect StatusError with VAP message** — VAP rejections come as `apierrors.StatusError` with the `message` from the CEL validation:
   ```go
   var statusErr *apierrors.StatusError
   Expect(errors.As(err, &statusErr)).To(BeTrue())
   Expect(statusErr.Status().Message).To(ContainSubstring("spec.path cannot be updated"))
   ```

4. **Resource cleanup** — use `DeferCleanup` in Ginkgo:
   ```go
   DeferCleanup(func() {
       _ = k8sClient.Delete(ctx, obj)
   })
   ```

5. **Wait for VAP compilation** — the first test in the suite should use `Eventually()` to probe VAP readiness before asserting:
   ```go
   Eventually(func() error {
       // Attempt a known-invalid update
       return k8sClient.Update(ctx, mutatedObj)
   }, "30s", "1s").Should(HaveOccurred())
   ```

### Architecture Compliance

| Requirement | Compliance |
|-------------|------------|
| envtest K8s version supports VAPs | ✅ `ENVTEST_K8S_VERSION = 1.37.0` (VAPs GA since 1.30) |
| Build tag separation | ✅ envtest tests: `!integration`, Kind tests: `integration` |
| Ginkgo/Gomega BDD framework | ✅ Consistent with all existing tests |
| Table-driven test pattern | ✅ Consistent with `webhook_validate_update_test.go` and controller integration tests |
| Shared framework protection | ✅ No framework code changes — only test files |
| Test file naming convention | ✅ `vap_test.go` / `vap_integration_test.go` follows existing patterns |
| No new dependencies | ✅ Uses existing k8s.io/apimachinery, controller-runtime, Ginkgo/Gomega |

### File Structure Requirements

**New files:**
- `api/v1alpha1/vap_test.go` — envtest VAP tests
- `internal/controller/vap_integration_test.go` — Kind integration VAP tests

**Modified files:**
- `api/v1alpha1/webhook_suite_test.go` — add VAP admission plugin, load VAP manifests, create param ConfigMap

**Deleted files:**
- `api/v1alpha1/webhook_validate_update_test.go` — replaced by VAP tests

### Testing Strategy

- **Unit tier (envtest):** Full coverage of all shared VAPs (path, name, target-ns-guard) and all credential source + CRD-specific VAPs. Table-driven tests with both rejection (invalid) and allowance (valid) cases.
- **Integration tier (Kind):** Smoke-level tests proving VAPs are deployed and enforced in a real Kind cluster. Not exhaustive — the envtest suite provides thorough coverage.
- **Regression:** `make test` and `make integration` must both pass. All existing tests that remain (RabbitMQ webhook, Transit webhook, etc.) must continue to work.

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 18.4] — Full story requirements
- [Source: _bmad-output/planning-artifacts/epics.md#Epic 18 header] — VAP architecture, gap analysis, test infrastructure notes
- [Source: api/v1alpha1/webhook_suite_test.go] — Current envtest setup (lines 62–272)
- [Source: api/v1alpha1/webhook_validate_update_test.go] — Existing path/name tests to be replaced (~1,315 lines)
- [Source: internal/controller/suite_integration_test.go] — Integration test suite setup (USE_EXISTING_CLUSTER pattern)
- [Source: Makefile:31-32] — `ENVTEST_K8S_VERSION = 1.37.0`
- [Source: Makefile:157-158] — `make test` target
- [Source: Makefile:163-212] — `make integration` target
- [Source: _bmad-output/implementation-artifacts/18-1-vap-infrastructure-and-universal-immutability-policies.md] — VAP infrastructure and envtest config details
- [Source: _bmad-output/implementation-artifacts/18-2-credential-source-and-crd-specific-vaps.md] — Credential source and CRD-specific VAP manifests
- [Source: _bmad-output/implementation-artifacts/18-3-webhook-code-removal-and-residual-webhook-cleanup-and-csv-version-bump.md] — Webhook cleanup and residual webhook classification
- Kubernetes VAP documentation: https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/
- envtest documentation: https://book.kubebuilder.io/reference/envtest

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

### Review Model Used

{{review_model_name_version}}

### Review Findings

### Decisions Needed / Decisions Taken

### Fixes Applied
