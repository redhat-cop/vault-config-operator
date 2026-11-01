# Story 18.3: Webhook Code Removal + Residual Webhook Cleanup + CSV Version Bump

Status: ready-for-dev

## Story

As an operator developer,
I want the migrated validating webhook code removed and the remaining webhooks clearly documented,
So that the codebase has a single source of truth for admission validation (VAPs for validation, minimal webhooks for mutation and cross-resource checks).

## Acceptance Criteria

1. **AC-1 — Migrated validation code removed:** All Go webhook validation logic migrated to VAPs in Stories 18.1 and 18.2 is deleted. `make test` and `make integration` still pass.
2. **AC-2 — Residual validating webhook retained:** `RabbitMQSecretEngineConfig` keeps its cross-resource `validateUniquePath` webhook (requires `client.List()`).
3. **AC-3 — Residual mutating webhooks retained:** AWSAuthEngineClientConfig, CFAuthEngineConfig (raw admission request inspection), GCPSecretEngineConfig, OktaAuthEngineConfig, RADIUSAuthEngineConfig, SSHSecretEngineConfig, TerraformCloudSecretEngineConfig, UserpassAuthEngineUser (simple credential key defaulters), KubernetesAuthEngineConfig (filesystem CA cert read).
4. **AC-4 — Defaulter-only webhook files:** Webhook files that previously contained both defaulting and validation logic retain `admission.Defaulter[*T]` but the `admission.Validator[*T]` methods return `nil, nil` (or validator registration is removed entirely if no validation remains).
5. **AC-5 — Kustomize webhook config cleaned:** `config/webhook/manifests.yaml` only contains residual webhooks; removed validating webhook paths are absent after `make manifests`.
6. **AC-6 — main.go registration cleaned:** Only residual webhooks are registered. Types fully covered by VAPs no longer register webhooks (or register only the defaulter).
7. **AC-7 — CSV minKubeVersion bumped:** `spec.minKubeVersion` changed from `1.29.0` → `1.30.0` in `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`.
8. **AC-8 — OCP version annotation bumped:** `com.redhat.openshift.versions` changed from `v4.16` → `v4.17` in `config/manifests/annotations.yaml`.
9. **AC-9 — Simple defaulter +kubebuilder:default evaluation:** Each of the 6 simple defaulters is evaluated for migration to `+kubebuilder:default` CRD markers. Those that can migrate do; those that cannot are documented with rationale.
10. **AC-10 — Build passes cleanly:** `make manifests generate fmt vet test lint` passes with zero errors.

## Tasks / Subtasks

- [ ] Task 1: Phase A — Remove VAP-migrated immutability checks from webhooks (AC: #1, #4)
  - [ ] 1.1 — Remove `spec.path` immutability checks from all ~68 webhook `ValidateUpdate` methods (covered by `vco-immutable-path` VAP from 18.1)
  - [ ] 1.2 — Remove `spec.name` immutability checks from all ~65 webhook `ValidateUpdate` methods (covered by `vco-immutable-name` VAP from 18.1)
  - [ ] 1.3 — Remove `ValidateTargetNamespaceUpdate()` calls from all ~91 webhook `ValidateUpdate` methods (covered by `vco-target-namespace-guard` VAP from 18.1)
  - [ ] 1.4 — Run `make test` to verify no regressions
- [ ] Task 2: Phase B — Remove VAP-migrated isValid() / credential checks from webhooks (AC: #1, #4)
  - [ ] 2.1 — Remove `isValid()` / `IsValid()` calls from `ValidateCreate` and `ValidateUpdate` in ~29 webhook files (covered by 18.2 CRD-specific + credential-source VAPs)
  - [ ] 2.2 — Remove any `ValidateCredentialSource()` calls from webhooks (covered by 18.2 credential mutual-exclusivity VAPs)
  - [ ] 2.3 — For each webhook file, determine final disposition: fully remove (no defaulter or validator logic remains) vs. defaulter-only (keep `Defaulter[*T]`, gut validators to `return nil, nil`)
  - [ ] 2.4 — Run `make test` to verify no regressions
- [ ] Task 3: Gut or delete fully-migrated webhook files (AC: #4, #6)
  - [ ] 3.1 — For ~70+ webhook files with no remaining defaulter logic and no remaining validator logic: remove `WithValidator(r)` from `SetupWebhookWithManager` and remove the validating kubebuilder marker; keep only the mutating scaffold returning `nil` from `Default` — OR — delete the webhook file entirely and remove from main.go (see Dev Notes for which approach)
  - [ ] 3.2 — For defaulter-only webhooks: remove `WithValidator(r)` chain call, remove validating kubebuilder marker, remove `Validator[*T]` compile-time check, delete `ValidateCreate`/`ValidateUpdate`/`ValidateDelete` methods; keep only `Defaulter[*T]` + `Default` + mutating kubebuilder marker
  - [ ] 3.3 — Verify the RabbitMQ `RabbitMQSecretEngineConfigValidation` handler struct and its `validateUniquePath` logic is untouched but remove the immutability/target-namespace checks from within its `Handle` method (those are now VAP-covered)
- [ ] Task 4: Clean up main.go webhook registrations (AC: #6)
  - [ ] 4.1 — Remove `SetupWebhookWithManager` calls for types that no longer have any webhook (fully VAP-covered, no defaulter)
  - [ ] 4.2 — Keep `SetupWebhookWithManager` calls for types that retain a defaulter
  - [ ] 4.3 — Keep the RabbitMQ raw handler registration: `mgr.GetWebhookServer().Register("/validate-...", &webhook.Admission{Handler: &redhatcopv1alpha1.RabbitMQSecretEngineConfigValidation{Client: mgr.GetClient()}})`
  - [ ] 4.4 — Verify total webhook registrations match expected residual count
- [ ] Task 5: Evaluate simple defaulters for +kubebuilder:default migration (AC: #9)
  - [ ] 5.1 — GCPSecretEngineConfig: defaults `passwordKey` to `"credentials"` — NOT migratable (override of inherited `"password"`, not a static default for the field)
  - [ ] 5.2 — OktaAuthEngineConfig: defaults `passwordKey` to `"api_token"` — NOT migratable (override of inherited `"password"`)
  - [ ] 5.3 — SSHSecretEngineConfig: defaults `passwordKey` to `"private_key"`, `usernameKey` to `"public_key"` — NOT migratable (conditional on `CAKeyReference != nil`, and override of inherited defaults)
  - [ ] 5.4 — TerraformCloudSecretEngineConfig: defaults `passwordKey` to `"token"` — NOT migratable (override of inherited `"password"`)
  - [ ] 5.5 — UserpassAuthEngineUser: defaults `passwordKey` to `"password"` — already the kubebuilder default from `RootCredentialConfig`; defaulter is redundant. Evaluate removing the defaulter entirely after verifying the CRD schema default handles it.
  - [ ] 5.6 — RADIUSAuthEngineConfig: defaults `passwordKey` to `"secret"` — NOT migratable (remaps from inherited `"password"` to `"secret"`, uses raw admission request inspection)
  - [ ] 5.7 — Document each decision in this story file's Completion Notes
- [ ] Task 6: CSV and platform version bumps (AC: #7, #8)
  - [ ] 6.1 — In `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`: change `minKubeVersion: 1.29.0` → `minKubeVersion: 1.30.0`
  - [ ] 6.2 — In `config/manifests/annotations.yaml`: change `com.redhat.openshift.versions: v4.16` → `com.redhat.openshift.versions: v4.17`
- [ ] Task 7: Regenerate manifests and validate (AC: #5, #10)
  - [ ] 7.1 — Run `make manifests generate fmt vet`
  - [ ] 7.2 — Verify `config/webhook/manifests.yaml` only contains residual webhook entries
  - [ ] 7.3 — Run `make test` (envtest unit tests)
  - [ ] 7.4 — Run `make lint`
  - [ ] 7.5 — Verify `make integration` passes (if Kind cluster available)

## Dev Notes

### Dependency: Stories 18.1 and 18.2 MUST be done first

This story assumes all VAPs from 18.1 (universal immutability) and 18.2 (credential source + CRD-specific) are in place and tested. Do NOT remove any Go validation code until the corresponding VAP tests pass.

### Webhook File Classification — What Stays vs. What Goes

**Category 1 — RETAIN with both defaulter + validator (1 type):**
- `RabbitMQSecretEngineConfig` — custom `RabbitMQSecretEngineConfigValidation` handler uses `client.List()` for cross-resource uniqueness. Remove the `spec.path` immutability check and `ValidateTargetNamespaceUpdate()` call from its `Handle` method (lines 54-59 in current code), but keep `validateUniquePath`. This is the ONLY retained validating webhook.

**Category 2 — RETAIN with defaulter-only, complex mutation (3 types):**
- `AWSAuthEngineClientConfig` — raw admission request inspection via `awsCredentialKeyOmitted()` to remap `usernameKey`/`passwordKey`. Cannot be expressed as `+kubebuilder:default`.
- `CFAuthEngineConfig` — raw admission request inspection via `cfCredentialKeyOmitted()` to remap `usernameKey`/`passwordKey`. Cannot be expressed as `+kubebuilder:default`.
- `KubernetesAuthEngineConfig` — reads `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` from the filesystem. Cannot be a CRD marker.

**Category 3 — RETAIN with defaulter-only, simple mutation (6 types):**
- `RADIUSAuthEngineConfig` — remaps `passwordKey` from `"password"` to `"secret"` via `radiusCredentialKeyOmitted()`. Cannot be a CRD marker.
- `GCPSecretEngineConfig` — remaps `passwordKey` from `"password"` to `"credentials"`. Cannot be a CRD marker.
- `OktaAuthEngineConfig` — defaults `passwordKey` to `"api_token"` from empty. Cannot be a CRD marker.
- `SSHSecretEngineConfig` — conditional remap of `passwordKey`/`usernameKey` when `CAKeyReference != nil`. Cannot be a CRD marker.
- `TerraformCloudSecretEngineConfig` — remaps `passwordKey` from `"password"` to `"token"`. Cannot be a CRD marker.
- `UserpassAuthEngineUser` — defaults `passwordKey` to `"password"` (redundant with kubebuilder default). Evaluate removal; if `RootCredentialConfig` CRD schema already defaults `passwordKey` to `"password"`, this defaulter is a no-op and the webhook can be fully removed.

**Category 4 — FULLY REMOVE (everything else, ~82 types):**
All other webhooks had ONLY validation logic (path/name immutability, target-namespace guard, isValid checks). After VAP migration, these files have no purpose. For each:
1. Delete the `*_webhook.go` file entirely
2. Remove the `SetupWebhookWithManager` call from `main.go`
3. The kubebuilder markers in the deleted file will cause `make manifests` to automatically regenerate `config/webhook/manifests.yaml` without those webhook paths

### Approach: Delete vs. Gut

**Recommended: Delete webhook files for Category 4 types.** When a webhook file has no defaulter logic (the `Default` method just logs and returns nil), there is no functional reason to keep it. Deleting the file:
- Removes ~80 lines of boilerplate per type × ~82 types ≈ **6,500 lines** of dead code
- Simplifies `main.go` by removing ~82 registration calls
- Makes `config/webhook/manifests.yaml` dramatically smaller
- Eliminates unnecessary webhook TLS handshake overhead for types that don't need it

**Important:** The `admission.Defaulter[*T]` and `admission.Validator[*T]` interfaces are NOT required by the operator framework. They are only needed if the type has actual defaulting or validation logic. Types without webhooks will still work — the API server simply doesn't call any webhook for them.

### RabbitMQ Webhook Cleanup Detail

The `RabbitMQSecretEngineConfigValidation.Handle` method currently contains:
1. `spec.path` immutability check (line 54-56) → **REMOVE** (VAP handles this)
2. `ValidateTargetNamespaceUpdate()` call (line 57-59) → **REMOVE** (VAP handles this)
3. `validateUniquePath()` call → **KEEP** (requires `client.List()`, cannot be expressed in CEL)

After cleanup, the UPDATE case in `Handle` should only call `validateUniquePath` when `targetNamespace` changed, plus the path uniqueness check on CREATE. The `spec.path` comparison that triggers `validateUniquePath` on namespace change is separate from the immutability check and must be preserved.

### main.go Cleanup Pattern

Current `main.go` has 91 `SetupWebhookWithManager` calls plus 1 raw `Register` for RabbitMQ. After cleanup:
- **~10 `SetupWebhookWithManager` calls remain** (Category 2 + Category 3 types)
- **1 raw `Register` call remains** (RabbitMQ cross-resource validator)
- **~81 registration calls removed**

### CSV Version Bump Rationale

VAPs are GA in Kubernetes 1.30 (KEP-3488). The operator now REQUIRES VAP support:
- `minKubeVersion: 1.30.0` — Kubernetes 1.30+
- `com.redhat.openshift.versions: v4.17` — OpenShift 4.17 ships Kubernetes 1.30

These are exact file locations:
- `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` line 348: `minKubeVersion: 1.29.0` → `1.30.0`
- `config/manifests/annotations.yaml` line 7: `com.redhat.openshift.versions: v4.16` → `v4.17`

### Existing Test Files to Update

- `api/v1alpha1/webhook_validate_update_test.go` — 1,315-line test file with table-driven tests for path/name immutability and target-namespace-update checks. After removing Go validation code, these tests will fail. **Options:**
  1. **Delete the test file** — The tests validated Go webhook logic that is now replaced by VAPs. VAP tests (Story 18.4) cover the same behavior via CEL.
  2. **Keep and adapt** — If some tested webhooks still have Go validation (RabbitMQ), keep only those test cases.
  - **Recommended:** Delete the file. The test validates Go-level webhook behavior that will no longer exist. Story 18.4 provides equivalent coverage via VAP/envtest tests.

- `api/v1alpha1/webhook_suite_test.go` — Test suite setup. Keep if any webhook tests remain; delete if all webhook test files are removed.

### File Inventory Summary

| Change | Count | Action |
|--------|-------|--------|
| Webhook files to delete | ~82 | Remove entirely (Category 4) |
| Webhook files to gut to defaulter-only | ~9 | Keep Defaulter, remove Validator (Category 2+3) |
| Webhook files to keep as-is | 1 | RabbitMQ (Category 1) — minus immutability/target-ns checks |
| main.go registrations to remove | ~81 | Delete SetupWebhookWithManager calls |
| main.go registrations to keep | ~10 + 1 | Defaulter types + RabbitMQ raw handler |
| Test files to delete/adapt | 1-2 | webhook_validate_update_test.go + possibly suite |

### Novelty Risk: LOW

This is a mechanical deletion/cleanup story. The risk is in removing too much or too little:
- **Too much:** Accidentally removing a defaulter that's still needed → operator creates resources with wrong default values
- **Too little:** Leaving dead webhook code that creates unnecessary overhead and confusion
- **Mitigation:** The build pipeline (`make test`, `make lint`, `make integration`) catches regressions. The Category 2+3 classification above explicitly lists every type that MUST retain a webhook.

### Project Structure Notes

- All webhook files live in `api/v1alpha1/*_webhook.go`
- Webhook manifests generated by `make manifests` from kubebuilder markers → `config/webhook/manifests.yaml`
- Webhook registrations in `cmd/main.go` inside `ENABLE_WEBHOOKS` guard
- Kustomize webhook config: `config/webhook/kustomization.yaml` (no changes needed — it includes `manifests.yaml` which is auto-generated)
- VAP manifests (from 18.1/18.2): `config/vap/` directory
- CSV: `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml`
- Annotations: `config/manifests/annotations.yaml`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 18.3] — Full story requirements and acceptance criteria
- [Source: _bmad-output/planning-artifacts/epics.md#Epic 18 header] — VAP architecture decisions, shared-policy strategy
- [Source: _bmad-output/project-context.md#Admission Webhooks] — Current webhook rules and patterns
- [Source: config/manifests/bases/vault-config-operator.clusterserviceversion.yaml:348] — Current minKubeVersion
- [Source: config/manifests/annotations.yaml:7] — Current OCP version annotation
- [Source: cmd/main.go:295-620] — Current webhook registration block
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_webhook.go] — Cross-resource webhook that MUST be retained
- [Source: api/v1alpha1/awsauthengineconfig_webhook.go] — Complex defaulter pattern (raw admission request)
- [Source: api/v1alpha1/kubernetesauthengineconfig_webhook.go] — Filesystem-reading defaulter

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
