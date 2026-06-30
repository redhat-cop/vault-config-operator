# Story D2.2: Standardize Kubernetes Auth Engine Docs

Status: ready-for-dev

## Story

As a user configuring Kubernetes authentication,
I want comprehensive, well-structured documentation for KubernetesAuthEngineConfig and KubernetesAuthEngineRole,
So that I can correctly configure the most common auth method.

## Acceptance Criteria

1. **Given** the existing KubernetesAuthEngine content in `auth-engines.md` (lines 39-113) **When** it is extracted to `docs/auth-engines/kubernetes.md` and standardized per the template **Then** it contains:
   - Overview linking to Vault Kubernetes auth docs
   - KubernetesAuthEngineConfig: complete YAML example, field descriptions (camelCase), `kubernetesCACert` behavior table, Vault CLI equivalent
   - KubernetesAuthEngineRole: complete YAML example, field descriptions, `targetNamespaceSelector` explanation, Vault CLI equivalent

2. **Given** the new `kubernetes.md` file **When** validated against the template structure **Then** it follows the same structure as `docs/auth-engines/cert.md` (Overview → Config CRD → Role CRD → See Also)

3. **Given** `docs/auth-engines.md` (the redirect pointer, post-D2.1) **When** the Kubernetes content is moved **Then** no Kubernetes-specific content remains in `auth-engines.md` (it should already be a redirect after D2.1)

## Tasks / Subtasks

- [ ] Task 1: Create `docs/auth-engines/kubernetes.md` (AC: 1, 2)
  - [ ] 1.1: Write Overview section — 2-3 sentences explaining Kubernetes auth, link to Vault docs, list the two CRDs
  - [ ] 1.2: Write KubernetesAuthEngineConfig section with Example YAML, Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.3: Include the `kubernetesCACert` behavior table (from existing `auth-engines.md` lines 65-70) — this is a key differentiator for this engine
  - [ ] 1.4: Write KubernetesAuthEngineRole section with Example YAML, Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.5: Include `targetNamespaces` explanation (selector vs static list, mutual exclusivity)
  - [ ] 1.6: Add "See Also" section with links to auth-section.md, contributing-vault-apis.md, and Vault docs

- [ ] Task 2: Audit field names for camelCase consistency (AC: 1)
  - [ ] 2.1: Cross-reference all field names in the new doc against the Go CRD types (`kubernetesauthengineconfig_types.go`, `kubernetesauthenginerole_types.go`) — field names in the doc MUST match the `json:` tag values exactly
  - [ ] 2.2: Fix any residual snake_case field names from the original `auth-engines.md` source (D1.3 did NOT audit Kubernetes section — confirmed in D1 retro)

- [ ] Task 3: Verify links and structure (AC: 2)
  - [ ] 3.1: Verify relative links resolve correctly from `docs/auth-engines/kubernetes.md` (`../auth-section.md`, `../contributing-vault-apis.md`)
  - [ ] 3.2: Verify structure matches `cert.md` pattern: heading hierarchy, section ordering, table format

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 1 new file: `docs/auth-engines/kubernetes.md`

### Dependency on D2.1

This story assumes D2.1 has been completed (creating `docs/auth-engines/index.md` and the redirect pointer in `docs/auth-engines.md`). If D2.1 is NOT yet done, this story can still proceed — the `kubernetes.md` file can be created independently. The index will reference it via `[kubernetes.md](kubernetes.md)`.

### Source Content Location

The content to extract and standardize is in `docs/auth-engines.md` lines 39-113:
- `## KubernetesAuthEngineConfig` (lines 39-73)
- `## KubernetesAuthEngineRole` (lines 75-113)

### Template to Follow

Use `docs/engine-doc-template.md` as the structural pattern. Use `docs/auth-engines/cert.md` as the concrete reference implementation (it is the only completed per-engine doc).

Key structural requirements from the template:
1. Title: `# Kubernetes Auth Engine`
2. Link to Vault docs immediately below title
3. `## Overview` — 2-3 sentences + CRD list
4. `## KubernetesAuthEngineConfig` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
5. `## KubernetesAuthEngineRole` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
6. `## See Also` (no Credential Resolution section — this engine doesn't use external credentials)

### KubernetesAuthEngineConfig — Complete Field Reference

From `api/v1alpha1/kubernetesauthengineconfig_types.go`, the `KAECConfig` struct has these fields:

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| kubernetesHost | `json:"kubernetesHost"` | `kubernetes_host` | string | Yes | `https://kubernetes.default.svc:443` |
| kubernetesCACert | `json:"kubernetesCACert,omitempty"` | `kubernetes_ca_cert` | string | No | see behavior table |
| PEMKeys | `json:"PEMKeys,omitempty"` | `pem_keys` | []string | No | — |
| issuer | `json:"issuer,omitempty"` | `issuer` | string | No | `kubernetes/serviceaccount` |
| disableISSValidation | `json:"disableISSValidation,omitempty"` | `disable_iss_validation` | bool | No | false |
| disableLocalCAJWT | `json:"disableLocalCAJWT,omitempty"` | `disable_local_ca_jwt` | bool | No | false |
| useOperatorPodCA | `json:"useOperatorPodCA"` | — (not sent to Vault) | bool | No | true |
| useAnnotationsAsAliasMetadata | `json:"useAnnotationsAsAliasMetadata,omitempty"` | `use_annotations_as_alias_metadata` | bool | No | false |

Additional top-level spec fields:
- `path` (Required) — mount path for the Kubernetes auth engine
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `tokenReviewerServiceAccount` (Optional) — service account for token review
- `name` (Optional) — override Vault object name

**Important:** `useOperatorPodCA` is an operator-side field (controls whether the operator pod's CA cert is injected). It is NOT sent to the Vault API. It only takes effect when `kubernetesCACert` is unset AND `disableLocalCAJWT` is true.

### KubernetesAuthEngineRole — Complete Field Reference

From `api/v1alpha1/kubernetesauthenginerole_types.go`, the `VRole` struct has these fields:

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| targetServiceAccounts | `json:"targetServiceAccounts"` | `bound_service_account_names` | []string | Yes (min 1) | `["default"]` |
| audience | `json:"audience,omitempty"` | `audience` | *string | No | — |
| aliasNameSource | `json:"aliasNameSource"` | `alias_name_source` | string | No | `serviceaccount_uid` |
| tokenTTL | `json:"tokenTTL,omitempty"` | `token_ttl` | int | No | 0 |
| policies | `json:"policies"` | `token_policies` | []string | Yes (min 1) | — |
| tokenMaxTTL | `json:"tokenMaxTTL,omitempty"` | `token_max_ttl` | int | No | 0 |
| tokenBoundCIDRs | `json:"tokenBoundCIDRs,omitempty"` | `token_bound_cidrs` | []string | No | — |
| tokenExplicitMaxTTL | `json:"tokenExplicitMaxTTL,omitempty"` | `token_explicit_max_ttl` | int | No | 0 |
| tokenNoDefaultPolicy | `json:"tokenNoDefaultPolicy,omitempty"` | `token_no_default_policy` | bool | No | false |
| tokenNumUses | `json:"tokenNumUses,omitempty"` | `token_num_uses` | int | No | 0 |
| tokenPeriod | `json:"tokenPeriod,omitempty"` | `token_period` | int | No | 0 |
| tokenType | `json:"tokenType"` | `token_type` | string | No | `default` |

Additional top-level spec fields:
- `path` (Required) — mount path of the Kubernetes auth engine
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `targetNamespaces` (Required) — namespace binding config (see below)
- `name` (Optional) — override Vault object name

**TargetNamespaces — Mutual Exclusivity Rule:**
The `targetNamespaces` field contains exactly ONE of:
- `targetNamespaceSelector` — a Kubernetes label selector; namespaces are resolved dynamically at each reconcile
- `targetNamespaces` — a static list of namespace names

The webhook validates that exactly one is specified. If the selector matches zero namespaces, the role is written with `bound_service_account_namespaces=["__no_namespace__"]` as a workaround to avoid Vault API errors.

### kubernetesCACert Behavior Table

This table MUST be included in the doc — it is critical for users to understand the CA resolution logic:

| `kubernetesCACert` | `disableLocalCAJWT` | `useOperatorPodCA` | Behaviour |
|---|---|---|---|
| set | ignored | ignored | The provided CA cert is used |
| unset | false | ignored | Vault pod's `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` is used |
| unset | true | false | The default OS CA where Vault runs is used |
| unset | true | true | The operator pod's CA cert is injected and used |

### No Credential Resolution Section

Unlike LDAP, JWT/OIDC, GCP, or Azure auth engines, the Kubernetes auth engine does NOT use external credentials (no `RootCredentialConfig`). The `tokenReviewerServiceAccount` is resolved internally by the operator (it creates a short-lived JWT token from the specified service account). Therefore, do NOT include a "Credential Resolution" section.

### Known Issues in Source Content

From D1 retrospective (section "Potential Friction Points"):
> Kubernetes and LDAP sections were NOT explicitly audited for snake_case field names in D1.3 — D2.2 and D2.3 will handle during extraction

Action: When extracting from `auth-engines.md`, carefully verify ALL field names use camelCase (matching JSON tags). The original source may still have `tokenReviewerServiceAccount.name` (correct) but potentially describe field behavior using snake_case Vault API names. In the field descriptions table, always use camelCase. Vault API names belong only in the "Vault CLI Equivalent" section.

### Relative Link Conventions

From `docs/auth-engines/kubernetes.md`:
- To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`
- To other engine files: `cert.md`, `ldap.md` (same directory)
- External: full URLs to Vault documentation

### Previous Story Intelligence

**From D2.1 (Directory Structure & Index Page):**
- Created `docs/auth-engines/index.md` with engine table linking to `kubernetes.md`
- Replaced `docs/auth-engines.md` with redirect pointer
- AuthEngineMount section is in the index page (not per-engine files)
- Cross-references from other docs to `auth-engines.md#kubernetesauthengineconfig` were documented for updating in D2.2-D2.5
- No cross-references to Kubernetes anchors were found in other doc files (only in `namespace-config.yaml` example which uses the CRD kind directly)

**From D1.1 (Template Creation):**
- Template was patched 4 times in review — always use current version at `docs/engine-doc-template.md`
- DNFR1-DNFR5 requirements define documentation standards

**From D1.2 (CertAuth Documentation):**
- First per-engine file at `docs/auth-engines/cert.md` — use as reference implementation
- Validates the template pattern works; established relative link patterns

**From D1.3 (Link and Naming Fixes):**
- Fixed snake_case→camelCase in GCP and Azure sections
- Kubernetes and LDAP were NOT in scope — residual snake_case possible
- Fixed leading-space code fences and broken cross-references

**From D1 Retrospective:**
- Documentation stories expect 3+ review findings — this is normal
- Opus 4.6 recommended for all stories
- D2 assessed as ready — no preparation needed

### Project Structure Notes

```
docs/
├── auth-engines/
│   ├── index.md          ← D2.1
│   ├── cert.md           ← EXISTS (D1.2) — reference implementation
│   └── kubernetes.md     ← NEW (this story)
├── auth-engines.md       ← redirect pointer (D2.1)
├── auth-section.md       ← shared auth config docs (unchanged)
├── engine-doc-template.md ← template (D1.1, review-patched)
└── ...
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D2.2] — Story requirements and acceptance criteria
- [Source: docs/auth-engines.md:39-113] — Kubernetes auth content to extract and standardize
- [Source: docs/auth-engines/cert.md] — Reference implementation for template pattern
- [Source: docs/engine-doc-template.md] — Template structure (D1.1, review-patched 4 times)
- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go] — CRD field definitions for Config
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go] — CRD field definitions for Role
- [Source: api/v1alpha1/utils/commons.go:420-430] — TargetNamespaceConfig struct definition
- [Source: _bmad-output/implementation-artifacts/d2-1-create-auth-engines-directory-structure-and-index-page.md] — Previous story context
- [Source: _bmad-output/implementation-artifacts/epic-d1-retro-2026-06-28.md] — D1 retro: known friction points for D2.2
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
