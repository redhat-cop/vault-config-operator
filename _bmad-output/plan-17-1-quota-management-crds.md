---
title: 'Quota Management CRDs — QuotaConfig + RateLimitQuota + LeaseCountQuota'
type: 'feature'
ticket: ''
created: '2026-10-20'
status: 'built'
baseline_revision: '3e01e1d8710c4bf160c970d35305b5e8f7ab6d68'
route: 'full'
route_source: 'auto'
risk: 'low'
review: 'quick'
review_source: 'pinned'
lenses_ran: ['quick']
review_loop_iteration: 0
context: []
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Vault's resource quota system (rate-limit quotas, lease-count quotas, and global quota config) cannot be managed declaratively via Kubernetes CRDs. Users must configure quotas manually through the Vault CLI/API. GitHub Issue #249 requests this capability.

**Approach:** Add three new CRD types — `QuotaConfig` (singleton global config at `sys/quotas/config`), `RateLimitQuota` (named resource at `sys/quotas/rate-limit/{name}`), and `LeaseCountQuota` (named resource at `sys/quotas/lease-count/{name}`, Enterprise-only) — following the `PasswordPolicy` controller pattern exactly, with standard `VaultObject` + `VaultResource` generic reconciler.

## Boundaries & Constraints

**Always:**
- Follow the `PasswordPolicy` pattern for controllers, webhooks, and type structure.
- `QuotaConfig.IsDeletable()` returns `false` — global config is always-present in Vault.
- Use `durationToSeconds()` for interval/blockInterval fields in `toMap()` — Vault returns integer seconds.
- Use `json.Number` for rate (float) and maxLeases (int) in `toMap()` — Vault returns `json.Number` via `UseNumber()`.
- Use `quotaPath` as the Go/JSON field name to avoid collision with `vaultutils.Path`; map to `"path"` in `toMap()`.
- Sort string slices (`rateLimitExemptPaths`) before comparison in `IsEquivalentToDesiredState`.
- `LeaseCountQuota` integration test is skipped — Enterprise-only API, not testable in OSS Kind cluster.

**Never:**
- Do not modify any shared framework files (`payload_filter.go`, `commons.go`, etc.).
- Do not implement webhook-based singleton uniqueness checks for QuotaConfig.
- Do not emit raw duration strings in `toMap()` — always convert to seconds.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| QuotaConfig create | CR with `enableRateLimitAuditLogging: true` | Vault `sys/quotas/config` updated, ReconcileSuccessful | Standard condition error |
| QuotaConfig delete | CR deleted | Vault config persists (IsDeletable=false) | No error — non-deletable |
| RateLimitQuota create | CR with `rate: 100`, `interval: "1m"` | Vault `sys/quotas/rate-limit/{name}` created | Standard condition error |
| RateLimitQuota update | `rate` changed 100→200 | Vault quota updated to match | Standard condition error |
| RateLimitQuota delete | CR deleted | Vault quota removed (404 on read) | Standard condition error |
| LeaseCountQuota on OSS | CR created against OSS Vault | ReconcileFailed — Enterprise-only API | Error surfaced as condition |
| Duration conversion | `interval: "2m"` | `toMap()` emits `json.Number("120")` | Invalid duration → passthrough |
| Float rate | `rate: 897.3` | `toMap()` emits `json.Number("897.3")` | N/A |
| Integer rate | `rate: 100` | `toMap()` emits `json.Number("100")` not `"100.0"` | N/A |

</frozen-after-approval>

## Code Map

- `api/v1alpha1/passwordpolicy_types.go` — Reference pattern: simple sys/ CRD with VaultObject interface, GetPath(), GetPayload(), IsEquivalentToDesiredState(), IsDeletable()
- `api/v1alpha1/passwordpolicy_webhook.go` — Reference pattern: Defaulter + Validator with ValidateTargetNamespaceUpdate
- `internal/controller/passwordpolicy_controller.go` — Reference pattern: ReconcilerBase + NewVaultResource + NewDefaultPeriodicReconcilePredicate
- `api/v1alpha1/sshsecretenginerole_types.go:219-228` — `durationToSeconds()` helper, package-level accessible
- `api/v1alpha1/databasesecretengineconfig_types.go:161-167` — `toInterfaceArray()` helper for []string→[]any
- `api/v1alpha1/awssecretenginerole_types.go:156-170` — `sortAnyStringSlice()` helper for slice sorting
- `api/v1alpha1/payload_filter.go` — `filterPayloadToDesiredKeys()`, `removeUnsetFields()` — do not modify
- `cmd/main.go:554-557` — Last controller registration (KerberosAuthEngineGroup), insert new controllers after line 557
- `cmd/main.go:932-936` — Last webhook registration (KerberosAuthEngineGroup), insert new webhooks before closing `}`
- `config/crd/kustomization.yaml` — Add 3 new CRD base files to resources list

## Tasks & Acceptance

**Execution:**
- [ ] `api/v1alpha1/quotaconfig_types.go` — CREATE — QuotaConfig CRD type: Spec with QuotaConfigData inline, VaultObject interface (GetPath→`sys/quotas/config`, IsDeletable→false), toMap() with toInterfaceArray + sortAnyStringSlice, IsEquivalentToDesiredState with filterPayloadToDesiredKeys
- [ ] `api/v1alpha1/quotaconfig_webhook.go` — CREATE — Defaulter (no-op), Validator (ValidateTargetNamespaceUpdate on update)
- [ ] `api/v1alpha1/ratelimitquota_types.go` — CREATE — RateLimitQuota CRD type: Spec with RateLimitQuotaConfig inline + Name field, VaultObject interface (GetPath with Name override, IsDeletable→true), toMap() with json.Number for rate, durationToSeconds for interval/blockInterval, quotaPath→"path" mapping
- [ ] `api/v1alpha1/ratelimitquota_webhook.go` — CREATE — Validator: rate>0, immutable spec.name, ValidateTargetNamespaceUpdate
- [ ] `api/v1alpha1/leasecountquota_types.go` — CREATE — LeaseCountQuota CRD type: same pattern as RateLimitQuota but simpler (maxLeases int, no duration fields), json.Number for maxLeases via strconv.Itoa
- [ ] `api/v1alpha1/leasecountquota_webhook.go` — CREATE — Validator: maxLeases≥1, immutable spec.name, ValidateTargetNamespaceUpdate
- [ ] `internal/controller/quotaconfig_controller.go` — CREATE — Standard VaultResource reconciler following PasswordPolicy pattern
- [ ] `internal/controller/ratelimitquota_controller.go` — CREATE — Standard VaultResource reconciler
- [ ] `internal/controller/leasecountquota_controller.go` — CREATE — Standard VaultResource reconciler
- [ ] `api/v1alpha1/quotaconfig_test.go` — CREATE — Table-driven tests: toMap() output, IsEquivalentToDesiredState positive/negative/extra-fields, string slice sorting
- [ ] `api/v1alpha1/ratelimitquota_test.go` — CREATE — Table-driven tests: toMap() with duration conversion + json.Number, IsEquivalentToDesiredState positive/negative/extra-fields
- [ ] `api/v1alpha1/leasecountquota_test.go` — CREATE — Table-driven tests: toMap() with json.Number, IsEquivalentToDesiredState positive/negative/extra-fields
- [ ] `internal/controller/quotaconfig_controller_test.go` — CREATE — Integration test: create, reconcile, update, delete-no-op (IsDeletable=false persistence)
- [ ] `internal/controller/ratelimitquota_controller_test.go` — CREATE — Integration test: create, reconcile, update, delete
- [ ] `test/quotaconfig/quota-config.yaml` — CREATE — QuotaConfig test fixture
- [ ] `test/quotaconfig/rate-limit-quota.yaml` — CREATE — RateLimitQuota test fixture
- [ ] `cmd/main.go` — MODIFY — Register 3 controllers after KerberosAuthEngineGroup (line 557) + 3 webhooks before closing ENABLE_WEBHOOKS block (line 935)
- [ ] `config/crd/kustomization.yaml` — MODIFY — Add 3 CRD base files after `make manifests`
- [ ] `docs/system-backend/quotas.md` — CREATE — Documentation for all 3 quota CRDs with Enterprise notice for LeaseCountQuota

**Acceptance Criteria:**
- Given a QuotaConfig CR is applied, when reconciled, then `sys/quotas/config` in Vault matches the CR spec and ReconcileSuccessful=True
- Given a RateLimitQuota CR is applied with `rate: 100` and `interval: "1m"`, when reconciled, then `sys/quotas/rate-limit/{name}` exists in Vault with `rate=100` and `interval=60`
- Given a LeaseCountQuota CR is applied with `maxLeases: 1000`, when reconciled, then `sys/quotas/lease-count/{name}` exists in Vault
- Given a QuotaConfig CR is deleted, when the reconciler processes it, then the global quota config persists in Vault
- Given `make manifests generate fmt vet test` is run, then it passes with no errors
- Given `make lint` is run, then it passes with no errors

## Implementation Notes

## Plan Change Log

## Review Triage Log

**Pass 1** (GPT 5.4 adversarial, quick lens) — verdicts: 2 high/medium, 1 low, 2 false

| # | Finding | Verdict | Route | Evidence |
|---|---------|---------|-------|----------|
| 1 | Rate field as string instead of float64 — docs show `rate: 100` numeric | `false` | rejected | Kubernetes auto-coerces YAML numbers to strings for string-typed CRD fields; users writing `rate: 100` get `"100"`. `controller-gen` constraint justifies the type change. |
| 2 | Omitting interval sends `0` instead of Vault default `1s` | `low` | rejected | Follows the project's majority pattern (radius, ssh, alicloud, approle, etc. all unconditionally emit durationToSeconds). Speculative harm — Vault likely treats `0` as "use default" for quota intervals like it does for TTL fields. |
| 3 | RateLimitQuota spec.name guard `oldObj.Spec.Name != ""` allows empty→non-empty mutation, orphaning Vault object | `medium` | `patch` | All 30+ existing webhooks use `newObj.Spec.Name != oldObj.Spec.Name` without the extra guard. Fixed: removed the guard. |
| 4 | LeaseCountQuota spec.name — same immutability hole as finding 3 | `medium` | `patch` | Same root cause and fix as finding 3. |
| 5 | QuotaConfig IsEquivalentToDesiredState doesn't sort Vault payload's string slices — false drift if Vault returns paths in different order | `medium` | `patch` | Existing pattern (awssecretenginerole) sorts both desiredState and payload. Fixed: added `sortAnyStringSlice(payload, ...)` calls. |

## Design Notes

**Singleton pattern (QuotaConfig):** QuotaConfig has no `spec.name` and no `spec.path`. `GetPath()` returns a fixed `"sys/quotas/config"`. Only one CR should exist per Vault instance; if multiple CRs target the same Vault, last-reconciled wins. No webhook uniqueness enforcement — the convention is documented.

**`quotaPath` field naming:** The Vault API field `path` collides with `vaultutils.Path` used for engine mount routing in `GetPath()`. The quota's `path` is a payload data field, so it's named `quotaPath` in Go/JSON and mapped to `"path"` in `toMap()`.

**`absolute_rate_limit_exempt_paths` uncertainty:** The Vault GET response sample doesn't show this field. If Vault doesn't return it on read, exclude it from IsEquivalentToDesiredState comparison by deleting it from desiredState before compare when the payload doesn't contain it. This is handled by `removeUnsetFields()`.

## Verification

**Commands:**
- `make manifests generate fmt vet` -- expected: clean exit, 3 new CRD YAMLs generated in `config/crd/bases/`
- `go test ./api/v1alpha1/... -run "QuotaConfig|RateLimitQuota|LeaseCountQuota" -v` -- expected: all unit tests pass
- `make lint` -- expected: no new lint errors
- `make test` -- expected: all existing + new unit tests pass
