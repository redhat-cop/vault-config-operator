# Story 18.5: Documentation — VAP Approach + Admission Control Guide

Status: done
baseline_revision: bfb1547e824e02734aa05463e39ee6863a3151df
review: 'quick'
review_source: 'pinned'
lenses_ran: quick

## Story

As an operator user and contributor,
I want documentation explaining the VAP-based admission control approach,
So that I understand how validation rules are enforced and how to add rules for new CRDs.

## Acceptance Criteria

1. **AC-1 — `docs/admission-control.md` exists:** A new documentation file explains the three-tier VAP architecture (universal shared, semi-shared credential, CRD-specific), how shared VAPs and bindings work, the target-namespace-guard parameterized VAP and ConfigMap setup, which webhooks remain and why (mutation, cross-resource checks), and Kubernetes version requirements (VAPs GA in 1.30+).
2. **AC-2 — Developer guide updated:** `docs/contributing-vault-apis.md` includes updated steps for adding a new CRD type reflecting the VAP-first approach: adding the new CRD to shared VAP bindings (path, name, target-ns), creating a CRD-specific VAP if the type has unique validation rules, testing the VAP in envtest.
3. **AC-3 — `project-context.md` admission rules updated:** The "Admission Webhooks" section in `_bmad-output/project-context.md` is rewritten to reflect the VAP approach: ValidateUpdate immutability is enforced by VAPs (not Go webhook code), new types must be added to shared VAP binding resource lists, residual mutating webhooks are documented, envtest VAP configuration is documented.
4. **AC-4 — New-type checklist updated in `project-context.md`:** The "Adding a New Vault API Type (Full Checklist)" section replaces webhook-centric step 4 with VAP binding + optional webhook step, and adds VAP binding as a distinct step.
5. **AC-5 — All doc content is accurate:** Documentation matches the actual implementation from Stories 18.1–18.4 (file paths, policy names, CEL patterns, envtest flags, residual webhook list).

## Tasks / Subtasks

- [ ] Task 1: Create `docs/admission-control.md` (AC: #1, #5)
  - [ ] 1.1 — Write overview section: purpose of admission control, transition from Go webhooks to VAPs
  - [ ] 1.2 — Write Tier 1 section: universal shared VAPs (`vco-immutable-path`, `vco-immutable-name`, `vco-target-namespace-guard`)
  - [ ] 1.3 — Write Tier 2 section: credential source shared VAPs (`vco-exactly-one-root-credential`, `vco-exactly-one-bind-credential`)
  - [ ] 1.4 — Write Tier 3 section: CRD-specific VAPs (list all ~15 `vco-validate-*` policies)
  - [ ] 1.5 — Write parameterized VAP section: `vco-target-namespace-guard` ConfigMap setup, `vco-vap-params` data, operator namespace
  - [ ] 1.6 — Write residual webhooks section: which webhooks remain (RabbitMQ cross-resource validator, AWS/CF/GCP/Okta/RADIUS/SSH/TFC/Userpass/KubernetesAuth defaulters) and why each cannot be a VAP
  - [ ] 1.7 — Write Kubernetes version requirements section: VAPs GA in K8s 1.30 (OCP 4.17), `minKubeVersion: 1.30.0`
  - [ ] 1.8 — Write manifest location section: all VAP manifests under `config/vap/`, naming convention `vco-*`, kustomize overlay integration
- [ ] Task 2: Update `docs/contributing-vault-apis.md` (AC: #2, #5)
  - [ ] 2.1 — Replace step 5 webhook section: remove Go `ValidateUpdate` immutability code examples; instead explain that immutability is enforced by shared VAPs
  - [ ] 2.2 — Add new step after webhook: "Add to VAP bindings" — add new CRD to `vco-immutable-path-binding.yaml` (if type has `spec.path`), `vco-immutable-name-binding.yaml` (if type has `spec.name`), and `vco-target-namespace-guard-binding.yaml` (all types)
  - [ ] 2.3 — Add CRD-specific VAP guidance: if the type has unique validation rules (credential mutual exclusivity, conditional fields, immutable sub-fields), create a dedicated `config/vap/vco-validate-<kind>.yaml` + binding
  - [ ] 2.4 — Add VAP testing guidance: explain envtest VAP test pattern for new CRD types
  - [ ] 2.5 — Keep webhook step for types that need defaulting logic only (credential key remapping, filesystem reads); remove webhook step for types with no defaulting
- [ ] Task 3: Update `_bmad-output/project-context.md` (AC: #3, #4, #5)
  - [ ] 3.1 — Rewrite "Admission Webhooks (Required for Every Type)" section header to "Admission Control (VAPs + Residual Webhooks)"
  - [ ] 3.2 — Replace webhook-first rules with VAP-first rules: immutability enforced by VAPs, credential validation by VAPs, CRD-specific validation by VAPs
  - [ ] 3.3 — Document residual webhook rules: when a webhook is still required (defaulting, cross-resource validation), credential key defaulting pattern remains in webhooks
  - [ ] 3.4 — Update "Adding a New Vault API Type (Full Checklist)" step 2 (currently says `operator-sdk create webhook`): clarify webhook scaffold is only needed if the type has defaulting logic
  - [ ] 3.5 — Update step 4 (currently says "Implement webhook"): split into "Add to shared VAP bindings" + "Implement webhook only if defaulting needed"
  - [ ] 3.6 — Add new VAP section: manifest location (`config/vap/`), shared policy naming convention (`vco-*`), parameterized ConfigMap, envtest configuration (`--enable-admission-plugins=ValidatingAdmissionPolicy`)
  - [ ] 3.7 — Update the "Strict webhook validation philosophy" bullet: reframe as "Strict admission validation philosophy" applying to both VAPs and residual webhooks
  - [ ] 3.8 — Remove or annotate now-outdated rules: `ValidateUpdate must always reject changes to spec.path` → "Enforced by `vco-immutable-path` VAP; no Go code needed"
  - [ ] 3.9 — Keep the credential key defaulting rules (these remain as Go webhooks)
- [ ] Task 4: Verify documentation accuracy (AC: #5)
  - [ ] 4.1 — Cross-check all VAP policy names against actual files in `config/vap/`
  - [ ] 4.2 — Cross-check residual webhook list against actual `*_webhook.go` files remaining after Story 18.3
  - [ ] 4.3 — Verify envtest flags match actual `webhook_suite_test.go` configuration from Story 18.1
  - [ ] 4.4 — Verify ConfigMap name and data key match actual `vco-vap-params-configmap.yaml`

## Dev Notes

### Dependency: Stories 18.1–18.4 should be done first

This documentation story describes the completed VAP implementation. The dev agent should verify the actual VAP manifests, webhook files, and test configuration exist before documenting them. If the previous stories are not yet implemented, the documentation should be written based on the epic specifications and story files (18-1, 18-2, 18-3) and adjusted if filenames or patterns differ.

### File Inventory — What to Create and Modify

| File | Action | Description |
|------|--------|-------------|
| `docs/admission-control.md` | **CREATE** | New admission control architecture guide |
| `docs/contributing-vault-apis.md` | **MODIFY** | Update steps 5 and add VAP binding step |
| `_bmad-output/project-context.md` | **MODIFY** | Rewrite admission control section, update new-type checklist |

### `docs/admission-control.md` — Content Structure

Follow the pattern of existing docs (e.g., `docs/policy-management.md`, `docs/secret-management.md`) for tone and formatting. The document should be user-facing (for operator users and contributors), not just internal developer notes.

**Recommended outline:**

```
# Admission Control

## Overview
- Brief explanation of admission control in Kubernetes
- vault-config-operator uses ValidatingAdmissionPolicies (VAPs) for declarative validation
- Residual Go webhooks for mutation and cross-resource checks

## Architecture — Three-Tier VAP Strategy

### Tier 1: Universal Shared Policies
- vco-immutable-path — prevents spec.path changes on UPDATE (72 CRDs)
- vco-immutable-name — prevents spec.name changes on UPDATE (65 CRDs)
- vco-target-namespace-guard — prevents Vault namespace changes when targetNamespace is set (92 CRDs, parameterized)

### Tier 2: Credential Source Shared Policies
- vco-exactly-one-root-credential — mutual exclusivity for rootCredentials (7 CRDs)
- vco-exactly-one-bind-credential — mutual exclusivity for bindCredentials/ldapCredentials

### Tier 3: CRD-Specific Policies
- List each vco-validate-* policy with a one-line description

## Parameterized VAP — Target Namespace Guard
- How the ConfigMap vco-vap-params works
- VAULT_NAMESPACE env var fallback
- Operator namespace setup

## Residual Webhooks
- RabbitMQ cross-resource validator (requires client.List())
- Credential key defaulters (AWS, CF, GCP, Okta, RADIUS, SSH, TFC, Userpass)
- KubernetesAuthEngineConfig (filesystem CA cert read)
- Why each cannot be expressed as a VAP

## Kubernetes Version Requirements
- VAPs are GA in Kubernetes 1.30 (KEP-3488)
- minKubeVersion: 1.30.0
- OpenShift 4.17+ (ships Kubernetes 1.30)

## Manifest Location
- config/vap/ directory structure
- Naming convention: vco-* prefix
- Kustomize overlay integration via config/default/kustomization.yaml
```

### `docs/contributing-vault-apis.md` — Changes Required

**Current step 5** (lines 129–157) shows a full webhook implementation with `ValidateUpdate` containing:
- `ValidateTargetNamespaceUpdate` call
- `spec.path` immutability check

**After update:** Step 5 should:
1. Explain that validation is handled by VAPs — no Go `ValidateUpdate` immutability code needed
2. Show a minimal webhook scaffold for types that need defaulting only
3. For types with NO defaulting: skip the webhook scaffold entirely
4. Reference `docs/admission-control.md` for full VAP details

**New step (after step 5):** "Add to VAP bindings"
```
6. Add the new CRD to shared VAP binding YAML files:
   - config/vap/vco-immutable-path-binding.yaml (if type has spec.path)
   - config/vap/vco-immutable-name-binding.yaml (if type has spec.name)
   - config/vap/vco-target-namespace-guard-binding.yaml (all types)
   If the type has unique validation rules, create a dedicated VAP:
   - config/vap/vco-validate-<lowercase-kind>.yaml
   - config/vap/vco-validate-<lowercase-kind>-binding.yaml
```

**Step numbers will shift** — current steps 6–9 become 7–10. Update cross-references accordingly.

### `_bmad-output/project-context.md` — Changes Required

**Section: "Admission Webhooks (Required for Every Type)"** (lines 56–70)

This entire section must be rewritten. The current text describes a webhook-first approach where every type needs a `*_webhook.go` file with `admission.Validator[*T]`. After Epic 18:
- **Validation is VAP-first** — shared VAPs handle path/name immutability and target-namespace checks for all types
- **Webhooks are only needed for mutation** (defaulting) and the single RabbitMQ cross-resource validator
- **Not every type needs a webhook file** — types without defaulting logic have no webhook at all (~82 types had their webhook files deleted in Story 18.3)

**Replacement content outline:**

```markdown
#### Admission Control (VAPs + Residual Webhooks)
- **ValidatingAdmissionPolicies (VAPs)** enforce all validation rules declaratively via CEL expressions
  - Shared VAPs in `config/vap/`: `vco-immutable-path` (72 CRDs), `vco-immutable-name` (65 CRDs), `vco-target-namespace-guard` (92 CRDs, parameterized)
  - Credential source VAPs: `vco-exactly-one-root-credential`, `vco-exactly-one-bind-credential`, plus per-CRD credential VAPs
  - CRD-specific VAPs: ~15 `vco-validate-*` policies for unique validation logic
- **New types must be added to shared VAP binding resource lists** in `config/vap/vco-immutable-path-binding.yaml`, `vco-immutable-name-binding.yaml`, `vco-target-namespace-guard-binding.yaml`
- **Strict admission validation philosophy:** Reject invalid field combinations at admission via CEL. Prefer VAP over Go webhook code for validation.
- **Residual Go webhooks** exist only for:
  - **Mutating defaulters:** AWSAuthEngineClientConfig, CFAuthEngineConfig (raw admission request inspection), GCPSecretEngineConfig, OktaAuthEngineConfig, RADIUSAuthEngineConfig, SSHSecretEngineConfig, TerraformCloudSecretEngineConfig, UserpassAuthEngineUser, KubernetesAuthEngineConfig (filesystem CA cert read)
  - **Cross-resource validator:** RabbitMQSecretEngineConfig (`validateUniquePath` requires `client.List()`)
- **Webhook files are NOT required for types without defaulting logic.** Most types (~82) have no webhook file at all.
- **Credential key defaulting** rules remain in Go webhooks (AWS `access_key`/`secret_key`, Okta `api_token`, etc.)
- **envtest VAP configuration:** The envtest suite enables VAPs via `KubeAPIServerFlags: []string{"--enable-admission-plugins=ValidatingAdmissionPolicy"}`
```

**Section: "Adding a New Vault API Type (Full Checklist)"** (lines 300–310)

Current steps:
1. `operator-sdk create api` — KEEP
2. `operator-sdk create webhook` — UPDATE: "Only if type needs defaulting logic"
3. Define `*_types.go` — KEEP
4. Implement webhook — UPDATE: "Implement webhook only if defaulting needed; immutability validation handled by VAPs"
5. Create `*_controller.go` — KEEP
6. Register in `main.go` — UPDATE: "Register controller. Register webhook only if defaulting webhook exists."
7. Create test YAML fixtures — KEEP
8. Write integration test — KEEP
9. `make manifests generate fmt vet test` — KEEP
10. Create documentation — KEEP
**NEW step (insert after step 3 or 4):** "Add to shared VAP binding YAML files (`config/vap/vco-immutable-path-binding.yaml`, `vco-immutable-name-binding.yaml`, `vco-target-namespace-guard-binding.yaml`). Create CRD-specific VAP if the type has unique validation rules."

### Existing Documentation Patterns — Follow These

- **Tone:** Concise, reference-oriented — not tutorial-style. See `docs/policy-management.md` and `docs/namespaces.md` as examples.
- **Code examples:** Use fenced code blocks with language tags (`yaml`, `cel`, `go`).
- **Links:** Use relative links to other docs files (e.g., `[Contributing a new Vault API](contributing-vault-apis.md)`).
- **Headers:** Use `#` for title, `##` for major sections, `###` for subsections.

### VAP Manifest Reference (for accurate documentation)

All VAP manifests live under `config/vap/`. The naming convention:
- Shared policies: `vco-immutable-path`, `vco-immutable-name`, `vco-target-namespace-guard`
- Credential source: `vco-exactly-one-root-credential`, `vco-exactly-one-bind-credential`
- CRD-specific: `vco-validate-<lowercase-kind>` (e.g., `vco-validate-transitsecretenginekey`)
- Each policy has a corresponding `-binding.yaml`
- The target-namespace-guard has a `vco-vap-params-configmap.yaml`

### Residual Webhook Reference (for accurate documentation)

After Story 18.3 cleanup, the following webhooks remain:

| Type | Webhook Purpose | Why Not VAP |
|------|----------------|-------------|
| RabbitMQSecretEngineConfig | Cross-resource `validateUniquePath` | Requires `client.List()` API call |
| AWSAuthEngineClientConfig | Credential key defaulting (raw admission request) | Mutation, not validation |
| CFAuthEngineConfig | Credential key defaulting (raw admission request) | Mutation, not validation |
| KubernetesAuthEngineConfig | CA cert read from filesystem | Mutation, not validation |
| RADIUSAuthEngineConfig | `passwordKey` remap to `"secret"` | Mutation, not validation |
| GCPSecretEngineConfig | `passwordKey` remap to `"credentials"` | Mutation, not validation |
| OktaAuthEngineConfig | `passwordKey` default to `"api_token"` | Mutation, not validation |
| SSHSecretEngineConfig | Conditional `passwordKey`/`usernameKey` remap | Mutation, not validation |
| TerraformCloudSecretEngineConfig | `passwordKey` remap to `"token"` | Mutation, not validation |
| UserpassAuthEngineUser | `passwordKey` default (evaluate removal — may be redundant with CRD default) | Mutation, not validation |

### Novelty Risk: LOW

This is a pure documentation story. No code changes, no risk of regressions. The only risk is inaccuracy — mitigated by cross-checking against actual files from Stories 18.1–18.4.

### Project Structure Notes

- New file `docs/admission-control.md` follows the existing docs directory structure
- No new directories created
- Modifications to `docs/contributing-vault-apis.md` and `_bmad-output/project-context.md` preserve existing file structure and formatting

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 18.5] — Full story requirements and acceptance criteria
- [Source: _bmad-output/planning-artifacts/epics.md#Epic 18 header] — VAP architecture decisions, three-tier strategy, residual webhook classification
- [Source: _bmad-output/project-context.md#Admission Webhooks] — Current webhook rules to rewrite (lines 56–70)
- [Source: _bmad-output/project-context.md#Adding a New Vault API Type] — Current checklist to update (lines 300–310)
- [Source: docs/contributing-vault-apis.md#Steps] — Current developer guide webhook step (lines 129–157)
- [Source: 18-1-vap-infrastructure-and-universal-immutability-policies.md] — VAP infrastructure, directory structure, envtest config, CEL expressions
- [Source: 18-2-credential-source-and-crd-specific-vaps.md] — Credential source and CRD-specific VAP manifests, CEL patterns
- [Source: 18-3-webhook-code-removal-and-residual-webhook-cleanup-and-csv-version-bump.md] — Webhook file classification (retain vs. delete), residual webhook inventory

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None

### Completion Notes List

- Created `docs/admission-control.md` — comprehensive guide covering three-tier VAP architecture, parameterized VAP setup, residual webhooks, K8s version requirements, and manifest location
- Updated `docs/contributing-vault-apis.md` — replaced webhook-centric step 5 with VAP-first approach; added new step 6 for VAP binding updates; renumbered steps 7–10; updated step 1 to note webhook scaffold is optional
- Updated `_bmad-output/project-context.md` — rewrote "Admission Webhooks" section to "Admission Control (VAPs + Residual Webhooks)"; updated new-type checklist steps 2, 4, 5, 7 to reflect VAP-first approach; added step 4 as VAP binding step
- All CRD counts verified against actual VAP manifests: 72 path, 65 name, 96 target-ns, 33 total policies, 33 bindings
- ConfigMap name (`vco-vap-params`) and data key (`vaultNamespace`) verified against actual manifest
- envtest configuration verified: `ControlPlane.GetAPIServer().Configure().Append("enable-admission-plugins", "ValidatingAdmissionPolicy")`
- Residual webhook list verified against actual remaining `*_webhook.go` files (24 files: 7 defaulter-only, 1 cross-resource validator, 16 retained validators)
- UserpassAuthEngineUser removed from residual webhook table (file was deleted in Story 18.3)
- Fixed "~82 types have no webhook" to "72 of 96 types have no webhook file" to match actual counts

### File List

New: docs/admission-control.md
Modified: docs/contributing-vault-apis.md, _bmad-output/project-context.md

## Code Review Record

### Review Model Used

Grok 4.7 xhigh (quick lens)

### Review Findings

## Review Triage Log

| # | Source | Location | Claim | Verdict | Evidence | Route |
|---|--------|----------|-------|---------|----------|-------|
| 1 | quick | `docs/admission-control.md` Overview + Retained table | Overview says residual webhooks are only mutation/cross-resource; table labels 8 validators as "No-op (logging only)" when all 16 have real validation | high | All 16 retained validators verified to have active validation logic (IsValid, immutability checks, field constraints) | patch |
| 2 | quick | `docs/contributing-vault-apis.md` steps 1+5+6 | Step 1 scaffolds `--programmatic-validation`, step 5 says `--defaulting` only; step 6 references "binding YAML files" but resource lists are on the policy objects | medium | Confirmed: binding files have no resource lists; policy files have `matchConstraints.resourceRules.resources` | patch |
| 3 | quick | `docs/admission-control.md` Parameterized VAP section | References `ValidateTargetNamespaceUpdate` method that no longer exists | low | `grep -r ValidateTargetNamespaceUpdate api/` returns nothing | patch |
| 4 | quick | `docs/admission-control.md` ConfigMap Setup | Kustomize wiring description slightly imprecise | low | Description is functionally correct: namespace transformer → ConfigMap → replacement → binding paramRef.namespace | rejected — description is roughly correct, fix adds complexity |
| 5 | quick | `docs/admission-control.md` Tier 3 table | Field name simplifications vs exact JSON tags | low | Summary descriptions, not CEL references; actual expressions in YAML | rejected — acceptable simplifications for summary table |
| 6 | quick | `_bmad-output/project-context.md` credential defaulting bullet | Removed nuance about empty-string fallback and inherited values | medium | Original text included "Empty-string fallback is acceptable when no admission request is present" — now missing | patch |

### Decisions Needed / Decisions Taken

- Rejected Finding 4 (kustomize wiring): description is functionally correct, namePrefix concern is operational not documentation
- Rejected Finding 5 (field name simplifications): acceptable for summary table, users read actual YAML for CEL

### Fixes Applied

1. **Finding 1+2 (retained validators):** Fixed overview to mention type-specific validation in residual webhooks. Updated all 16 rows in Retained Webhook Validators table with accurate validation descriptions (removed incorrect "No-op validators (logging only)" labels).
2. **Finding 3 (step contradictions):** Removed `--programmatic-validation` from step 1 webhook command in `contributing-vault-apis.md` and project-context.md checklist. Changed "binding YAML files" to "policy YAML files" and clarified resource lists are on `ValidatingAdmissionPolicy` objects (`spec.matchConstraints.resourceRules[].resources`).
3. **Finding 4 (historical method):** Updated parameterized VAP description to use past tense for `ValidateTargetNamespaceUpdate` ("previously implemented in... now removed").
4. **Finding 7 (credential defaulting):** Restored nuance to credential key defaulting bullet in project-context.md — distinguished between request-inspection defaulters (AWS, CF, RADIUS) and value-check defaulters (GCP, Okta, SSH, TFC), restored empty-string fallback and inherited-default rules.
