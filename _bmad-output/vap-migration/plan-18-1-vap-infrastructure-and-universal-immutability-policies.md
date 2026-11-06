---
title: 'VAP Infrastructure + Path/Name Immutability Policies'
type: 'feature'
ticket: ''
created: '2026-11-01'
status: 'built'
baseline_revision: '6fe676c7cd08ab8a6d458e4b9750114ba4bcd909'
route: 'full'
route_source: 'auto'
risk: 'high'
review: 'quick'
review_source: 'pinned'
lenses_ran: ['quick']
review_loop_iteration: 0
context:
  - '_bmad-output/implementation-artifacts/18-1-vap-infrastructure-and-universal-immutability-policies.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Immutability validation for `spec.path` and `spec.name` is duplicated across 92 Go webhook files, with 4 CRDs missing path checks and 31 missing name checks. The scattered Go code is error-prone and hard to audit.

**Approach:** Create `config/vap/` with two shared ValidatingAdmissionPolicies (vco-immutable-path, vco-immutable-name) using CEL, wire them into kustomize, configure envtest for VAP support, then remove the migrated path/name Go validation from webhook `ValidateUpdate` methods. The target-namespace-guard VAP is deferred to a follow-up plan.

## Boundaries & Constraints

**Always:**
- VAPs must be provably enforcing in envtest BEFORE removing any Go validation
- Use `admissionregistration.k8s.io/v1` (GA since K8s 1.30; envtest uses K8s 1.37 assets)
- Keep all `ValidateTargetNamespaceUpdate` calls, `isValid()` calls, `ValidateCredentialSource()` calls, CRD-specific immutability checks, `ValidateCreate`/`ValidateDelete` bodies, and `Default()` methods untouched
- RabbitMQSecretEngineConfig's `Handle` method: only remove path check (L54-56); keep everything else

**Never:**
- Do not remove `ValidateTargetNamespaceUpdate` calls (deferred to follow-up plan)
- Do not remove credential validation or CRD-specific checks (Story 18.2 scope)
- Do not modify `ValidateCreate`, `ValidateDelete`, or `Default()` methods

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Path changed | UPDATE changing `spec.path` on any of 72 CRDs | API server rejects: "spec.path cannot be updated" | VAP denial |
| Path unchanged | UPDATE not changing `spec.path` | Allowed | N/A |
| Name changed | UPDATE changing `spec.name` on any of 65 CRDs | Rejected: "spec.name cannot be updated" | VAP denial |
| Name never set | `spec.name` absent on both old and new | Allowed | N/A |
| Name added after creation | Adding `spec.name` (was absent) | Rejected — changes effective Vault path | VAP denial |
| CRD-specific check | UPDATE changing `localSecretIDs` on AppRoleAuthEngineRole | Still rejected by Go webhook | Webhook error |

</frozen-after-approval>

## Code Map

- `config/default/kustomization.yaml` (L21) — add `- ../vap` after `- ../webhook`
- `config/vap/` — NEW: kustomization.yaml, vco-immutable-path.yaml, vco-immutable-path-binding.yaml, vco-immutable-name.yaml, vco-immutable-name-binding.yaml
- `api/v1alpha1/*_webhook.go` (92 files) — remove path checks from 68 files (pattern: `Spec.Path != oldObj.Spec.Path` → `errors.New("spec.path cannot be updated")`); remove name checks from 34 files (pattern: `Spec.Name != oldObj.Spec.Name`); `rabbitmqsecretengineconfig_webhook.go` uses `Handle()` not `ValidateUpdate()` — remove path check at L54-56 only
- `api/v1alpha1/webhook_suite_test.go` (L73-85) — add `KubeAPIServerFlags` for VAP plugin; load VAP manifests post-start via k8sClient
- `api/v1alpha1/webhook_validate_update_test.go` (1314 lines) — remove path/name rejection tests; keep CRD-specific tests

## Tasks & Acceptance

**Execution:**
- [ ] `config/vap/kustomization.yaml` — CREATE listing VAP/binding YAMLs as resources
- [ ] `config/vap/vco-immutable-path.yaml` — CREATE VAP: CEL `object.spec.path == oldObject.spec.path`
- [ ] `config/vap/vco-immutable-path-binding.yaml` — CREATE binding: 72 CRDs in `redhatcop.redhat.io`, UPDATE ops
- [ ] `config/vap/vco-immutable-name.yaml` — CREATE VAP: CEL handling `has()`/absent semantics
- [ ] `config/vap/vco-immutable-name-binding.yaml` — CREATE binding: 65 CRDs
- [ ] `config/default/kustomization.yaml` — ADD `- ../vap` to resources
- [ ] `api/v1alpha1/webhook_suite_test.go` — ADD VAP plugin flag + manifest loading + smoke test
- [ ] `api/v1alpha1/*_webhook.go` — REMOVE path checks (68 files) and name checks (34 files); clean unused imports
- [ ] `api/v1alpha1/rabbitmqsecretengineconfig_webhook.go` — REMOVE path check in Handle (L54-56) only
- [ ] `api/v1alpha1/webhook_validate_update_test.go` — REMOVE path/name rejection tests; keep CRD-specific

**Acceptance Criteria:**
- Given `kustomize build config/default/`, when rendered, then output includes path and name VAP manifests
- Given envtest with VAP plugin, when UPDATE changes `spec.path`, then API server rejects
- Given envtest, when UPDATE changes `spec.name`, then API server rejects
- Given webhook code after removal, when CRD-specific check triggers, then webhook still rejects
- Given `GOTOOLCHAIN=auto make test`, when run, then all tests pass

## Implementation Notes

## Plan Change Log

## Review Triage Log

**Pass 1 (quick lens):** 0 high, 0 medium, 2 low, 2 false, 0 maybe-false

| Finding | Verdict | Route | Evidence |
|---------|---------|-------|----------|
| VAP smoke tests lack `Eventually()` | false | reject | Tests passed; K8s 1.37 envtest enforces VAPs immediately |
| Comment removed from ValidateCreate in kubernetessecretenginerole | false | reject | Misplaced comment about path immutability in wrong method; no behavior change |
| LeaseCountQuota/RateLimitQuota lose detailed error format | low | reject | Deliberate consistency trade-off; VAP message still clear; unlikely to affect users |
| Duplicate log line in pkisecretenginerole_webhook.go | low | patch | Pre-existing duplicate exposed by removal; removed second instance |

## Design Notes

**CEL for vco-immutable-name:**
```cel
!has(oldObject.spec.name) && !has(object.spec.name) ||
has(oldObject.spec.name) && has(object.spec.name) && object.spec.name == oldObject.spec.name
```
With `omitempty` JSON tags, `has()` returns false for absent and empty-string `spec.name`. This is correct: empty = "not set" = uses `metadata.name` fallback.

**envtest VAP loading:** VAPs aren't CRDs — `CRDDirectoryPaths` won't load them. After `testEnv.Start()`, read YAMLs from `config/vap/` and apply via `k8sClient.Create()`. Use `Eventually()` to wait for enforcement before rejection tests.

**Execution order:** Create VAP manifests → configure envtest → verify enforcement → remove Go validation → update tests.

**Deriving CRD lists:** The 72 path CRDs = all types whose Go spec struct has a `Path` field used in `GetPath()`. The 65 name CRDs = types with `Name` field used in `GetPath()` via `ResourceName()`. Derive exact plural names from `config/crd/bases/` filenames.

## Verification

**Commands:**
- `GOTOOLCHAIN=auto make manifests generate fmt vet test` — expected: all pass
- `kustomize build config/default/ | grep -c ValidatingAdmissionPolicy` — expected: ≥2
