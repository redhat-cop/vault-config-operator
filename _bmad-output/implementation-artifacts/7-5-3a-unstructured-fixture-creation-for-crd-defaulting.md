# Story 7.5.3a: Unstructured Fixture Creation for CRD Defaulting

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration test fixtures to be created using unstructured objects that preserve YAML field semantics,
So that CRD server-side defaulting applies correctly and test fixtures don't need explicit values for defaulted fields.

## Acceptance Criteria

1. **Given** a YAML fixture that omits fields with `+kubebuilder:default` markers **When** the fixture is created via the new `CreateFromYAML` method **Then** the API server applies CRD defaults and the resource passes validation
2. **Given** the new `CreateFromYAML` method **When** all integration tests use it for fixture creation **Then** `make integration` passes with no regressions
3. **Given** test fixtures that had explicit default values added in Stories 7.5.1, 7.5.2, and 7.5.3 **When** those explicit values are reverted **Then** the fixtures still work because CRD defaulting supplies the values
4. **Given** the typed decoder methods (e.g., `GetKubernetesAuthEngineRoleInstance`) **When** tests need typed object references for later operations (Delete, GetPath, assertions) **Then** typed objects are obtained via a subsequent `Get` call after unstructured creation

## Tasks / Subtasks

- [ ] Task 1: Add `CreateFromYAML` method to the decoder (AC: 1)
  - [ ] 1.1: Implement `CreateFromYAML(ctx, client, filename, namespace) (string, error)` in `controllers/controllertestutils/decoder.go` — reads YAML, decodes to `unstructured.Unstructured`, sets namespace, creates via client, returns the object name
  - [ ] 1.2: Write a unit test for `CreateFromYAML` that verifies unstructured creation preserves only YAML-present fields
- [ ] Task 2: Refactor `KubernetesAuthEngine` integration tests to use `CreateFromYAML` (AC: 2)
  - [ ] 2.1: Refactor `kubernetesauthengine_controller_test.go` — replace `decoder.GetXxx` + `Create` with `CreateFromYAML` + typed `Get` for config and role fixtures
  - [ ] 2.2: Verify Kubernetes auth engine tests pass with `make integration` (run only the relevant test if possible, otherwise full suite)
- [ ] Task 3: Refactor shared test helpers to use `CreateFromYAML` (AC: 2)
  - [ ] 3.1: Refactor `SetupKVv2Stack` and `SetupKVv2StackWithReader` in `integration_test_helpers_test.go` — replace `decoder.GetXxx` + `Create` with `CreateFromYAML` + typed `Get` for all `KubernetesAuthEngineRole` fixtures
  - [ ] 3.2: Apply same pattern to `PasswordPolicy`, `Policy`, `SecretEngineMount` fixtures in these helpers (consistency)
- [ ] Task 4: Refactor remaining integration tests to use `CreateFromYAML` (AC: 2)
  - [ ] 4.1: Refactor `databasesecretenginestaticrole_controller_test.go`
  - [ ] 4.2: Refactor `pkisecretengine_controller_test.go`
  - [ ] 4.3: Refactor `vaultsecret_controller_test.go`
  - [ ] 4.4: Refactor `kubernetessecretengine_controller_test.go`
  - [ ] 4.5: Refactor all other controller test files that use typed decoder + Create pattern
- [ ] Task 5: Revert explicit default values from test fixtures (AC: 3)
  - [ ] 5.1: Revert `aliasNameSource`/`tokenType` additions from 13 KubernetesAuthEngineRole fixtures (Story 7.5.3)
  - [ ] 5.2: Revert `boundClaimsType: "string"` from JWT/OIDC role fixture (Story 7.5.2)
  - [ ] 5.3: Revert `tlsMinVersion`/`tlsMaxVersion` from LDAP config fixture (Story 7.5.1) — verify these are revertible (field must be absent from YAML for CRD default to apply)
- [ ] Task 6: Run `make manifests generate fmt vet test` (AC: 2)
- [ ] Task 7: Run `make integration` — full suite must pass (AC: 1, 2, 3)

## Dev Notes

### Problem Statement

The current test infrastructure uses typed Go struct serialization (`decoder.GetXxxInstance` → `client.Create`) which always serializes fields without `omitempty` in their JSON tags, even when those fields were absent in the original YAML fixture. This prevents CRD server-side defaulting from applying — the API server sees the field as "present with zero value" rather than "absent", so it skips defaulting and runs validation against the zero value.

This was discovered during Epic 7.5's R2 changes (removing `omitempty` from non-zero default fields). Each story (7.5.1, 7.5.2, 7.5.3) required adding explicit default values to test fixtures as a workaround. Stories 7.5.4–7.5.6 will hit the same problem unless the infrastructure is fixed first.

### Root Cause Chain

1. YAML fixture omits field (e.g., no `tokenType` in role YAML)
2. `serializer.UniversalDeserializer` decodes YAML into typed Go struct → field gets Go zero value (`""`)
3. `k8sIntegrationClient.Create()` marshals Go struct to JSON → `"tokenType": ""` is serialized (no `omitempty`)
4. API server receives `tokenType: ""` → field is PRESENT → CRD defaulting is SKIPPED
5. Enum validation runs on `""` → rejects with "Unsupported value"

### Solution: `CreateFromYAML` Using Unstructured Objects

Add a method to the decoder that reads YAML directly into an `unstructured.Unstructured` object (which only contains fields present in the YAML), then creates via the client. The API server correctly identifies absent fields and applies CRD defaults.

```go
import (
    "context"
    "os"

    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/util/yaml"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

func (d *decoder) CreateFromYAML(ctx context.Context, c client.Client, filename string, namespace string) (string, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return "", err
    }

    obj := &unstructured.Unstructured{}
    if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096).Decode(obj); err != nil {
        return "", err
    }

    obj.SetNamespace(namespace)
    if err := c.Create(ctx, obj); err != nil {
        return "", err
    }

    return obj.GetName(), nil
}
```

### Test Migration Pattern

**Before (current):**
```go
roleInstance, err = decoder.GetKubernetesAuthEngineRoleInstance("../test/.../role.yaml")
Expect(err).To(BeNil())
roleInstance.Namespace = vaultAdminNamespaceName
Expect(k8sIntegrationClient.Create(ctx, roleInstance)).Should(Succeed())
```

**After (new):**
```go
name, err := decoder.CreateFromYAML(ctx, k8sIntegrationClient, "../test/.../role.yaml", vaultAdminNamespaceName)
Expect(err).To(BeNil())
roleInstance = &redhatcopv1alpha1.KubernetesAuthEngineRole{}
Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: name, Namespace: vaultAdminNamespaceName}, roleInstance)).Should(Succeed())
```

The typed `Get` retrieves the object WITH server-applied defaults, so `roleInstance.Spec.TokenType` will be `"default"` even though the YAML didn't specify it.

### Impact on Existing Test Patterns

- **Create pattern**: Changes from typed Create to unstructured Create + typed Get (2 calls instead of 1)
- **Delete pattern**: Unchanged — still uses typed object reference
- **Assertion pattern**: Unchanged — typed object has all fields populated (including defaults) after Get
- **waitForReconcileSuccess**: Unchanged — still uses typed object
- **Stack helpers**: `SetupKVv2Stack` and `SetupKVv2StackWithReader` need refactoring to new pattern
- **Cleanup (AfterAll/TeardownKVv2Stack)**: Unchanged — still uses typed object pointers

### Fixtures to Revert (Story 7.5.3)

| Fixture File | Fields to Remove |
|---|---|
| `test/kubernetesauthengine/test-kube-auth-role.yaml` | `aliasNameSource`, `tokenType` |
| `test/kubernetesauthengine/test-kube-auth-role-selector.yaml` | `aliasNameSource`, `tokenType` |
| `test/database-engine-admin-role.yaml` | `aliasNameSource`, `tokenType` |
| `test/kv-engine-admin-role.yaml` | `aliasNameSource`, `tokenType` |
| `test/kube-auth-engine-role.yaml` | `aliasNameSource`, `tokenType` |
| `test/secret-writer-role.yaml` | `aliasNameSource`, `tokenType` |
| `test/rabbitmq-engine-admin-role.yaml` | `aliasNameSource`, `tokenType` |
| `test/databasesecretengine/database-secret-engine-auth-role.yaml` | `aliasNameSource`, `tokenType` |
| `test/pkisecretengine/pki-secret-engine-kube-auth-role.yaml` | `aliasNameSource`, `tokenType` |
| `test/vaultsecret/kubernetesauthenginerole-secret-reader.yaml` | `aliasNameSource`, `tokenType` |
| `test/vaultsecret/v2/00-kubernetesauthenginerole-secret-reader-v2.yaml` | `aliasNameSource`, `tokenType` |
| `test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml` | `aliasNameSource`, `tokenType` |
| `test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml` | `aliasNameSource`, `tokenType` |

### Fixtures to Revert (Story 7.5.2)

| Fixture File | Fields to Remove |
|---|---|
| `test/jwtoidcauthengine/test-jwtoidc-auth-role.yaml` | `boundClaimsType` |

### Fixtures to Revert (Story 7.5.1)

| Fixture File | Fields to Remove |
|---|---|
| `test/ldapauthengine/test-ldap-auth-config.yaml` | `tlsMinVersion`, `tlsMaxVersion` (verify these were added for this reason) |

### Critical Warnings

1. **Do NOT remove the typed decoder methods** — they are still needed for tests that read typed objects without creating them, and for the Get step after unstructured creation
2. **The `os.ReadFile` import** replaces the deprecated `ioutil.ReadFile` already used in decoder.go — use `os.ReadFile` in the new method
3. **Namespace must be set after YAML parse** — the unstructured object may or may not have a namespace in the YAML; always override with the test namespace
4. **The `unstructured.Unstructured` package** is at `k8s.io/apimachinery/pkg/apis/meta/v1/unstructured` — this dependency is already in the project
5. **Story 7.5.1 fixture revert needs verification** — check whether `tlsMinVersion`/`tlsMaxVersion` were added specifically for the omitempty issue or for a different reason before reverting
6. **Config sample file** `config/samples/redhatcop_v1alpha1_kubernetesauthenginerole.yaml` had `aliasNameSource`/`tokenType` added in 7.5.3 — this is a user-facing example, keep the explicit values (they serve as documentation)

### Affected Files Summary

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/controllertestutils/decoder.go` | Modified | Add `CreateFromYAML` method |
| 2 | `controllers/kubernetesauthengine_controller_test.go` | Modified | Use `CreateFromYAML` + typed Get |
| 3 | `controllers/kubernetessecretengine_controller_test.go` | Modified | Use `CreateFromYAML` + typed Get |
| 4 | `controllers/integration_test_helpers_test.go` | Modified | Refactor stack helpers |
| 5 | `controllers/databasesecretenginestaticrole_controller_test.go` | Modified | Use `CreateFromYAML` + typed Get |
| 6 | `controllers/pkisecretengine_controller_test.go` | Modified | Use `CreateFromYAML` + typed Get |
| 7 | `controllers/vaultsecret_controller_test.go` | Modified | Use `CreateFromYAML` + typed Get |
| 8+ | Other controller test files | Modified | Use `CreateFromYAML` + typed Get |
| 14+ | 13+ YAML test fixtures | Reverted | Remove explicit default values |

### Project Structure Notes

- Decoder lives in `controllers/controllertestutils/decoder.go`
- All integration tests are in `controllers/` with `//go:build integration` tag
- Test fixtures are in `test/` subdirectories organized by engine type
- The `unstructured` package is already an indirect dependency via controller-runtime

### References

- [Source: controllers/controllertestutils/decoder.go] — Current decoder implementation using typed serialization
- [Source: controllers/integration_test_helpers_test.go] — Shared KVv2Stack helpers using typed decoder
- [Source: controllers/kubernetesauthengine_controller_test.go] — Kubernetes auth engine test patterns
- [Source: _bmad-output/implementation-artifacts/7-5-3-kubernetes-auth-and-secret-engine-types-annotation-refactor.md#Debug Log References] — Documents the R2 omitempty + CRD defaulting interaction
- [Source: _bmad-output/implementation-artifacts/7-5-1-ldap-auth-engine-types-annotation-refactor.md#File List] — Story 7.5.1 fixture modifications
- [Source: _bmad-output/implementation-artifacts/7-5-2-jwtoidc-auth-engine-types-annotation-refactor.md#File List] — Story 7.5.2 fixture modifications
- [Source: _bmad-output/project-context.md] — Project conventions and testing standards

### Previous Story Intelligence

**From Story 7.5.3 (Kubernetes Auth & Secret Engine Types — Annotation Refactor):**
- Discovered: removing `omitempty` from JSON tags (R2) causes Go struct serialization to send empty strings for absent YAML fields
- Root cause: `serializer.UniversalDeserializer` sets Go zero values for absent YAML fields, then `client.Create` serializes all non-omitempty fields
- 13 KubernetesAuthEngineRole fixtures required `aliasNameSource`/`tokenType` additions as workaround
- Integration tests failed with Enum validation rejecting `""` for `tokenType` and `aliasNameSource`

**From Story 7.5.2 (JWT/OIDC Auth Engine Types — Annotation Refactor):**
- Same R2 pattern: `boundClaimsType` added to JWT/OIDC role fixture
- 1 fixture modified

**From Story 7.5.1 (LDAP Auth Engine Types — Annotation Refactor):**
- `tlsMinVersion`/`tlsMaxVersion` added to LDAP config fixture
- 1 fixture modified — verify this was the same R2 issue

### Git Intelligence

- Latest commit: `2dfb1b3` (Story 7.5.2)
- Story 7.5.3 changes are uncommitted (in working tree)
- Branch is on main

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
