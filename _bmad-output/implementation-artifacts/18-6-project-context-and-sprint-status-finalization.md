# Story 18.6: Project Context + Sprint Status Finalization

Status: ready-for-dev

## Story

As a project maintainer,
I want the sprint-status.yaml, project-context.md, and epics.md fully updated for Epic 18,
So that project tracking artifacts reflect the VAP migration and all downstream workflows pick up the new rules.

## Acceptance Criteria

1. **AC-1 — Sprint status finalized:** `sprint-status.yaml` shows all stories 18.1–18.5 as `done`, epic-18 as `done`, and an `epic-18-retrospective` entry.
2. **AC-2 — project-context.md admission control rules updated:** The "Admission Webhooks" section in `_bmad-output/project-context.md` is rewritten to reflect the VAP-first approach:
   - `ValidateUpdate` immutability is now enforced by VAPs, not Go webhook code
   - New types must be added to shared VAP binding resource lists
   - Residual mutating webhooks documented (which types and why)
   - envtest VAP configuration documented (`--enable-admission-plugins=ValidatingAdmissionPolicy`)
   - The three-tier VAP architecture is referenced (Tier 1 universal, Tier 2 credential-source, Tier 3 CRD-specific)
3. **AC-3 — New-type checklist updated:** The "Adding a New Vault API Type (Full Checklist)" section in `project-context.md` includes VAP binding steps:
   - Step 4 updated from "Implement webhook" to "Add to shared VAP bindings + implement webhook only if defaulting logic needed"
   - New step for creating CRD-specific VAP if the type has unique validation rules
   - New step for adding the CRD to `config/vap/` binding manifests (path, name, target-namespace-guard)
4. **AC-4 — project-context.md technology references updated:** Technology stack sections reflect:
   - Minimum Kubernetes 1.30 / OpenShift 4.17 (VAPs GA)
   - `minKubeVersion: 1.30.0` in CSV
   - `config/vap/` directory and kustomize overlay
5. **AC-5 — epics.md complete:** Epic 18 is fully present in `_bmad-output/planning-artifacts/epics.md` with all six story descriptions and final status annotation.
6. **AC-6 — Retrospective action items tracked:** Any retro action items (e.g., "add VAP binding step to new-type checklist") are added to the `action_items:` section of `sprint-status.yaml`.
7. **AC-7 — Anti-patterns updated:** The "Anti-Patterns to Avoid" and "Critical Don't-Miss Rules" sections in `project-context.md` include VAP-era rules:
   - "Never add validating logic in Go webhooks for new types — use VAPs (CEL) instead"
   - "When adding a new CRD, always add it to the shared VAP bindings in `config/vap/`"
8. **AC-8 — Documentation gate verified:** Every new user-facing capability from Epic 18 has corresponding documentation (Story 18.5's `docs/admission-control.md` exists and is linked).

## Tasks / Subtasks

- [ ] Task 1: Verify all prerequisite stories are done (AC: #1)
  - [ ] 1.1 — Confirm stories 18.1 through 18.5 are all in `done` status in `sprint-status.yaml`
  - [ ] 1.2 — If any story is not `done`, STOP and report which stories are incomplete
- [ ] Task 2: Update project-context.md — Admission Control section (AC: #2, #7)
  - [ ] 2.1 — Rewrite the "Admission Webhooks (Required for Every Type)" subsection under "Go Language Rules":
    - Rename to "Admission Control (VAPs + Residual Webhooks)"
    - Document that validating logic is now expressed as VAPs in `config/vap/` using CEL expressions
    - Document the three-tier VAP architecture: Tier 1 (universal shared — immutable-path, immutable-name, target-namespace-guard), Tier 2 (semi-shared credential source), Tier 3 (CRD-specific)
    - Document residual mutating webhooks: AWSAuthEngineClientConfig, CFAuthEngineConfig, KubernetesAuthEngineConfig (complex mutation), RADIUSAuthEngineConfig, GCPSecretEngineConfig, OktaAuthEngineConfig, SSHSecretEngineConfig, TerraformCloudSecretEngineConfig, UserpassAuthEngineUser (simple credential key defaulters)
    - Document the one retained validating webhook: RabbitMQSecretEngineConfig (cross-resource `validateUniquePath` — requires `client.List()`)
    - Document that new types NO LONGER need `admission.Validator[*T]` implementations unless they have cross-resource validation needing API calls
    - Keep `admission.Defaulter[*T]` documentation for types that still need runtime defaulting
    - Document envtest VAP configuration: `KubeAPIServerFlags: []string{"--enable-admission-plugins=ValidatingAdmissionPolicy"}`
  - [ ] 2.2 — Update "Anti-Patterns to Avoid" with VAP-era rule: "Never add validating logic in Go webhooks for new types — express validation rules as VAPs in `config/vap/` using CEL"
  - [ ] 2.3 — Update "Critical Don't-Miss Rules" or add a new rule: "When adding a new CRD, add it to the shared VAP binding resource lists in `config/vap/vco-immutable-path-binding.yaml`, `vco-immutable-name-binding.yaml`, and `vco-target-namespace-guard-binding.yaml`"
- [ ] Task 3: Update project-context.md — New Type Checklist (AC: #3)
  - [ ] 3.1 — Revise step 4 ("Implement webhook") to: "Add to shared VAP bindings (`config/vap/vco-immutable-path-binding.yaml`, `vco-immutable-name-binding.yaml`, `vco-target-namespace-guard-binding.yaml`) + implement webhook in `*_webhook.go` ONLY if the type needs runtime defaulting logic (most types do NOT)"
  - [ ] 3.2 — Add new step: "If the type has unique validation logic (field constraints, conditional requirements, mutual exclusivity), create a CRD-specific VAP in `config/vap/vco-validate-<lowercase-kind>.yaml` with its binding"
  - [ ] 3.3 — Add new step: "If the type has credential source fields (rootCredentials, bindCredentials, etc.), add it to the appropriate shared credential-source VAP binding"
  - [ ] 3.4 — Update step 6 (main.go registration) to note that types without defaulting logic do NOT register webhooks at all
- [ ] Task 4: Update project-context.md — Technology references (AC: #4)
  - [ ] 4.1 — Update the "Core" section under Technology Stack: note that K8s 1.30+ is required (VAPs GA), or at minimum note the CSV minKubeVersion is 1.30.0
  - [ ] 4.2 — Add `config/vap/` to the "Code Organization" section as the location for VAP manifests
  - [ ] 4.3 — Update "Build Targets" if needed: note that `make manifests` now also triggers kustomize inclusion of VAP manifests via the `config/default` overlay
- [ ] Task 5: Update project-context.md — Webhook-specific rules cleanup (AC: #2)
  - [ ] 5.1 — Remove or annotate rules that are now VAP-handled:
    - The rule "ValidateUpdate must always reject changes to spec.path" → note this is now enforced by VAP `vco-immutable-path`
    - The rule "ValidateUpdate must also reject changes to spec.name" → note this is now enforced by VAP `vco-immutable-name`
    - The rule about `ValidateTargetNamespaceUpdate` → note this is now enforced by VAP `vco-target-namespace-guard`
  - [ ] 5.2 — Keep the "Strict webhook validation philosophy" rule but update it to reference VAPs: "Express validation rules as CEL in VAPs. Reject invalid field combinations at admission time."
  - [ ] 5.3 — Keep the "Credential key defaulting" rule as-is — this remains in mutating webhooks
  - [ ] 5.4 — Keep the "Credential validation" rule about `ValidateCredentialSource()` but note it's now enforced by credential-source VAPs, not Go code
- [ ] Task 6: Verify epics.md (AC: #5)
  - [ ] 6.1 — Confirm Epic 18 is fully present in `_bmad-output/planning-artifacts/epics.md` with all 6 stories
  - [ ] 6.2 — If any story description is missing, add it
- [ ] Task 7: Finalize sprint-status.yaml (AC: #1, #6)
  - [ ] 7.1 — Update all 18.x story statuses to `done` (assuming prerequisites are met — see Task 1)
  - [ ] 7.2 — Update `epic-18` status to `done`
  - [ ] 7.3 — Add `epic-18-retrospective: done` entry (or `optional` if retrospective is deferred)
  - [ ] 7.4 — Update `last_updated` timestamp
  - [ ] 7.5 — Add change log comment at top: `# Epic 18 done YYYY-MM-DD`
  - [ ] 7.6 — Add action items from retrospective to `action_items:` section
- [ ] Task 8: Documentation gate verification (AC: #8)
  - [ ] 8.1 — Verify `docs/admission-control.md` exists (created by Story 18.5)
  - [ ] 8.2 — Verify it covers: three-tier VAP architecture, shared VAP and binding workflow, target-namespace-guard parameterized ConfigMap, residual webhooks, K8s version requirements
  - [ ] 8.3 — If documentation is incomplete, document the gap as a retro action item
- [ ] Task 9: Run retrospective (AC: #6)
  - [ ] 9.1 — Review all 6 stories for lessons learned
  - [ ] 9.2 — Identify action items and add to `sprint-status.yaml`
  - [ ] 9.3 — Write retrospective entry in `_bmad-output/implementation-artifacts/epic-18-retro-YYYY-MM-DD.md`

## Dev Notes

### Dependency: ALL Stories 18.1–18.5 Must Be Done First

This is the final "close the loop" story for Epic 18. **Do not start this story until all preceding stories (18.1–18.5) are verified as `done` in sprint-status.yaml.** If any preceding story is not done, halt and report.

### What Changed in Epic 18 (Summary for project-context.md Updates)

Epic 18 migrated validating admission webhooks to Kubernetes ValidatingAdmissionPolicies (VAPs). Key changes:

1. **New `config/vap/` directory** containing ~20 VAP manifests organized in three tiers:
   - **Tier 1 — Universal shared (3 VAPs, ~228 bindings):** `vco-immutable-path` (72 CRDs), `vco-immutable-name` (65 CRDs), `vco-target-namespace-guard` (92 CRDs, parameterized with ConfigMap)
   - **Tier 2 — Semi-shared credential source (2-3 VAPs):** `vco-exactly-one-root-credential` (7 CRDs), `vco-exactly-one-bind-credential`, etc.
   - **Tier 3 — CRD-specific (~15 VAPs):** `vco-validate-transitsecretenginekey`, `vco-validate-awssecretenginerole`, etc.

2. **~82 webhook files deleted** (had only validation logic, no defaulting). ~10 retained for defaulting + 1 for cross-resource validation (RabbitMQ).

3. **35 validation gaps fixed:** 4 missing `spec.path` immutability checks + 31 missing `spec.name` immutability checks — automatically covered by shared VAP bindings.

4. **CSV version bumps:** `minKubeVersion: 1.29.0` → `1.30.0`, `com.redhat.openshift.versions: v4.16` → `v4.17` (VAPs are GA in K8s 1.30 / OCP 4.17).

5. **envtest configured for VAPs:** `KubeAPIServerFlags: []string{"--enable-admission-plugins=ValidatingAdmissionPolicy"}` added to envtest setup.

6. **Documentation:** `docs/admission-control.md` created explaining the three-tier approach.

### Exact project-context.md Sections to Modify

The file is at `_bmad-output/project-context.md`. The following sections need changes:

| Section | Line Range (approx) | Change Type |
|---------|---------------------|-------------|
| `#### Admission Webhooks (Required for Every Type)` | ~55–70 | **Major rewrite** — rename, document VAP-first approach, keep residual webhook docs |
| `#### Adding a New Vault API Type (Full Checklist)` | ~300–310 | **Update steps** — add VAP binding steps, update webhook step |
| `### Anti-Patterns to Avoid` | ~320–330 | **Add** — VAP-era anti-patterns |
| `### Code Organization` | ~225–230 | **Add** — `config/vap/` directory |
| `## Technology Stack & Versions` | ~20–50 | **Update** — note K8s 1.30+ requirement |
| `#### ConditionsAware Implementation` | ~72 | No change |
| `### Testing Rules` | ~155–210 | **Add** — envtest VAP configuration note |

### project-context.md "Admission Webhooks" Section — Before/After Guidance

**Current section name:** "Admission Webhooks (Required for Every Type)"

**New section name:** "Admission Control (VAPs + Residual Webhooks)"

**New content structure:**
1. **VAP-first approach:** All validating logic is expressed as ValidatingAdmissionPolicies (CEL) in `config/vap/`. Go webhook validators are NOT used for new types.
2. **Three-tier architecture:** Reference the tiers (Tier 1 universal, Tier 2 credential-source, Tier 3 CRD-specific).
3. **Shared VAP naming:** `vco-*` prefix for all policies.
4. **New CRD requirements:** Must be added to shared VAP bindings (path, name, target-namespace-guard). Must have a CRD-specific VAP if unique validation exists.
5. **Residual webhook contracts:**
   - Mutating webhooks (`admission.Defaulter[*T]`) remain for types needing runtime defaulting.
   - The RabbitMQ cross-resource validating webhook remains (requires `client.List()`).
   - Types without defaulting logic have NO webhook registration at all.
6. **Webhook file requirements (for types that still have webhooks):** Keep compile-time check, kubebuilder markers, logger, `SetupWebhookWithManager` pattern. But only for `admission.Defaulter[*T]` — no `admission.Validator[*T]` needed.
7. **Target-namespace-guard parameterization:** ConfigMap `vco-vap-params` holds `VAULT_NAMESPACE` env var value.

### sprint-status.yaml Update Pattern

Add to the change log comments at the top (near line 38):
```yaml
# Epic 18 done YYYY-MM-DD
# Epic 18 retrospective YYYY-MM-DD
# Story 18.6 done YYYY-MM-DD
# Story 18.5 done YYYY-MM-DD
# Story 18.4 done YYYY-MM-DD
```

Update `development_status` entries:
```yaml
18-1-vap-infrastructure-and-universal-immutability-policies: done
18-2-credential-source-and-crd-specific-vaps: done
18-3-webhook-code-removal-and-residual-webhook-cleanup-and-csv-version-bump: done
18-4-vap-test-suite-envtest-and-integration-coverage: done
18-5-documentation-vap-approach-and-admission-control-guide: done
18-6-project-context-and-sprint-status-finalization: done
epic-18: done
epic-18-retrospective: done  # or optional
```

### Retrospective Guidance

The Epic 18 retrospective should evaluate:

1. **VAP approach effectiveness:** Did the shared-policy strategy work? Were CEL expressions maintainable? Was the three-tier organization clear?
2. **Gap coverage:** Did the 35 fixed gaps (4 path + 31 name) create any regressions? Were there additional gaps discovered?
3. **Test infrastructure:** Did envtest VAP testing work reliably? Any flakiness with VAP propagation delay?
4. **Webhook removal scope:** Was the ~82 webhook file deletion clean? Any unexpected issues from removing webhook registrations?
5. **Documentation completeness:** Is `docs/admission-control.md` sufficient for future contributors?
6. **Action items to propose:**
   - "Verify VAP bindings are updated when future epics add new CRD types" (ongoing)
   - "Monitor CEL expression complexity — consider splitting overly complex CRD-specific VAPs" (ongoing)
   - If any issues surfaced: track remediation

### Novelty Risk: LOW

This is a documentation/finalization story with no code changes beyond project tracking files. The risk is limited to:
- **Incomplete project-context.md updates** — mitigated by the detailed section-by-section task breakdown above
- **Missing retrospective insights** — mitigated by the structured retrospective guidance
- **Stale sprint-status.yaml** — mitigated by verifying all prerequisites first (Task 1)

### Project Structure Notes

**Files to modify:**
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — story/epic status updates
- `_bmad-output/project-context.md` — admission control rules, new-type checklist, technology stack
- `_bmad-output/planning-artifacts/epics.md` — verify Epic 18 completeness

**Files to create:**
- `_bmad-output/implementation-artifacts/epic-18-retro-YYYY-MM-DD.md` — retrospective entry

**No Go code, YAML manifests, or test files are modified in this story.**

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 18.6 — lines 3596–3622] — Story requirements and acceptance criteria
- [Source: _bmad-output/planning-artifacts/epics.md#Epic 18 header — lines 3278–3310] — Epic scope, architecture decisions, shared-policy strategy
- [Source: _bmad-output/project-context.md#Admission Webhooks — lines 55–70] — Current webhook rules to rewrite
- [Source: _bmad-output/project-context.md#Adding a New Vault API Type — lines 300–310] — Current checklist to update
- [Source: _bmad-output/project-context.md#Anti-Patterns to Avoid — lines 320–330] — Anti-patterns to extend
- [Source: _bmad-output/implementation-artifacts/18-1-vap-infrastructure-and-universal-immutability-policies.md] — VAP infrastructure details (Tier 1 VAPs, envtest config, webhook removal pattern)
- [Source: _bmad-output/implementation-artifacts/18-2-credential-source-and-crd-specific-vaps.md] — Tier 2 + Tier 3 VAP details, CEL patterns
- [Source: _bmad-output/implementation-artifacts/18-3-webhook-code-removal-and-residual-webhook-cleanup-and-csv-version-bump.md] — Webhook classification (retain vs. remove), CSV version bumps, residual webhook inventory

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
