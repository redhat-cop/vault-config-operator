# Story D1.3: Fix Broken Links and Field Naming Inconsistencies

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a documentation reader,
I want all links to work and field names to consistently use camelCase (matching the CRD spec),
So that the docs are accurate and navigable.

## Acceptance Criteria

1. **Given** `secret-engines.md` lines 19-20 have broken markdown links with spaces before parentheses (`[AzureSecretEngineConfig] (#azuresecretengineconfig)`) **When** the links are fixed **Then** all TOC links render correctly
2. **Given** `auth-engines.md` GCPAuthEngineRole section uses mixed snake_case and camelCase field names **When** all field references are updated to camelCase (matching the CRD Go struct json tags) **Then** field naming is consistent across all engine docs
3. **Given** cross-references exist between doc files (e.g., `secret-management.md#RandomSecret`) **When** all internal links are audited **Then** every link resolves correctly

## Tasks / Subtasks

- [x] Task 1: Fix broken TOC links in `docs/secret-engines.md` (AC: 1)
  - [x] 1.1: Line 19: `[AzureSecretEngineConfig] (#azuresecretengineconfig)` → `[AzureSecretEngineConfig](#azuresecretengineconfig)` (remove space before paren, remove trailing whitespace)
  - [x] 1.2: Line 20: `[AzureSecretEngineRole] (#azuresecretenginerole)` → `[AzureSecretEngineRole](#azuresecretenginerole)` (remove space before paren)
- [x] Task 2: Fix broken `#RandomSecret` cross-file references in `docs/secret-engines.md` (AC: 3)
  - [x] 2.1: Line 97: `[RandomSecret](#RandomSecret)` → `[RandomSecret](secret-management.md#randomsecret)` (wrong file + wrong case)
  - [x] 2.2: Line 286: `[RandomSecret](#RandomSecret)` → `[RandomSecret](secret-management.md#randomsecret)` (wrong file + wrong case)
  - [x] 2.3: Line 416: `[RandomSecret](#RandomSecret)` → `[RandomSecret](secret-management.md#randomsecret)` (wrong file + wrong case)
  - [x] 2.4: Line 681: `[RandomSecret](secret-management.md#RandomSecret)` → `[RandomSecret](secret-management.md#randomsecret)` (correct file, wrong anchor case)
- [x] Task 3: Fix broken `#RandomSecret` cross-file references in `docs/auth-engines.md` (AC: 3)
  - [x] 3.1: Line 159: `[RandomSecret](#RandomSecret)` → `[RandomSecret](secret-management.md#randomsecret)` (wrong file + wrong case)
  - [x] 3.2: Line 314: `[RandomSecret](secret-management.md#RandomSecret)` → `[RandomSecret](secret-management.md#randomsecret)` (correct file, wrong anchor case)
  - [x] 3.3: Line 671: `[RandomSecret](secret-management.md#RandomSecret)` → `[RandomSecret](secret-management.md#randomsecret)` (correct file, wrong anchor case)
- [x] Task 4: Fix broken reference link in `docs/secret-management.md` (AC: 3)
  - [x] 4.1: Line 10: `[Password Policy]` is a reference-style link with no definition → change to `[Password Policy](https://www.vaultproject.io/docs/concepts/password-policies)` or add a reference definition at the bottom of the file
  - [x] 4.2: Line 123: `[authentication](#the-authentication-section)` links to an anchor in the wrong file → change to `[authentication](auth-section.md#the-authentication-section)` (the heading exists in `auth-section.md`, not `secret-management.md`)
- [x] Task 5: Fix double-hash anchor in `docs/end-to-end-example.md` (AC: 3)
  - [x] 5.1: Line 12: `[here](./../readme.md##deploying-the-operator)` → `[here](./../readme.md#deploying-the-operator)` (remove duplicate `#`)
- [x] Task 6: Fix snake_case field names in GCPAuthEngineRole section of `docs/auth-engines.md` (AC: 2)
  - [x] 6.1: Line 546: `bound_service_accounts` → `boundServiceAccounts`
  - [x] 6.2: Line 548: `bound_projects` → `boundProjects`
  - [x] 6.3: Line 550: `add_group_aliases` → `addGroupAliases`
  - [x] 6.4: Line 552: `token_ttl` → `tokenTTL`
  - [x] 6.5: Line 554: `token_max_ttl` → `tokenMaxTTL`
  - [x] 6.6: Line 556: `token_policies` → `tokenPolicies`
  - [x] 6.7: Line 560: `token_bound_cidrs` → `tokenBoundCIDRs`
  - [x] 6.8: Line 562: `token_explicit_max_ttl` → `tokenExplicitMaxTTL`
  - [x] 6.9: Line 564: `token_no_default_policy` → `tokenNoDefaultPolicy`
  - [x] 6.10: Line 566: `token_num_uses` → `tokenNumUses`
  - [x] 6.11: Line 568: `token_period` → `tokenPeriod`
  - [x] 6.12: Line 570: `token_type` → `tokenType`
  - [x] 6.13: Line 574: `max_jwt_exp` → `maxJWTExp`
  - [x] 6.14: Line 576: `allow_gce_inference` → `allowGCEInference`
  - [x] 6.15: Line 580: `bound_zones` → `boundZones`
  - [x] 6.16: Line 582: `bound_regions` → `boundRegions`
  - [x] 6.17: Line 584: `bound_instance_groups` → `boundInstanceGroups`
  - [x] 6.18: Line 586: `bound_labels` → `boundLabels`
- [x] Task 7: Fix snake_case field names in AzureAuthEngineConfig section of `docs/auth-engines.md` (AC: 2)
  - [x] 7.1: Line 626: `tenant_id` → `tenantID`
  - [x] 7.2: Line 634: `client_id` → `clientID`
  - [x] 7.3: Line 636: `client_secret` → `azureCredentials` (the CRD uses `azureCredentials` of type `RootCredentialConfig`, not a bare `client_secret` string)
  - [x] 7.4: Line 638: `max_retries` → `maxRetries`
  - [x] 7.5: Line 640: `max_retry_delay` → `maxRetryDelay`
  - [x] 7.6: Line 642: `retry_delay` → `retryDelay`
- [x] Task 8: Fix snake_case field names in AzureAuthEngineRole section of `docs/auth-engines.md` (AC: 2)
  - [x] 8.1: Line 717: YAML example uses PascalCase `BoundResourceGroups:` → `boundResourceGroups:` (this is invalid YAML for the CRD)
  - [x] 8.2: Line 743: `bound_service_principal_ids` → `boundServicePrincipalIDs`
  - [x] 8.3: Line 745: `bound_group_ids` → `boundGroupIDs`
  - [x] 8.4: Line 747: `bound_locations` → `boundLocations`
  - [x] 8.5: Line 749: `bound_subscription_ids` → `boundSubscriptionIDs`
  - [x] 8.6: Line 751: `bound_resource_groups` → `boundResourceGroups`
  - [x] 8.7: Line 753: `bound_scale_sets` → `boundScaleSets`
  - [x] 8.8: Line 755: `token_ttl` → `tokenTTL`
  - [x] 8.9: Line 757: `token_max_ttl` → `tokenMaxTTL`
  - [x] 8.10: Line 759: `token_policies` → `tokenPolicies`
  - [x] 8.11: Line 763: `token_bound_cidrs` → `tokenBoundCIDRs`
  - [x] 8.12: Line 765: `token_explicit_max_ttl` → `tokenExplicitMaxTTL`
  - [x] 8.13: Line 767: `token_no_default_policy` → `tokenNoDefaultPolicy`
  - [x] 8.14: Line 769: `token_num_uses` → `tokenNumUses`
  - [x] 8.15: Line 771: `token_period` → `tokenPeriod`
  - [x] 8.16: Line 773: `token_type` → `tokenType`
- [x] Task 9: Fix leading-space code fences in `docs/auth-engines.md` (AC: 3)
  - [x] 9.1: Line 499: ` ```yaml` → ````yaml` (remove leading space)
  - [x] 9.2: Line 540: ` ``` ` → ` ``` ` (remove leading space and trailing space — closing fence in GCPAuthEngineRole YAML block)
  - [x] 9.3: Line 684: ` ```yaml` → ````yaml` (remove leading space — opening fence in AzureAuthEngineRole YAML block)
  - [x] 9.4: Line 739: ` ``` ` → ` ``` ` (remove leading space and trailing space — closing fence in AzureAuthEngineRole YAML block)
- [x] Task 10: Final audit pass (AC: 1, 2, 3)
  - [x] 10.1: Grep all doc files for `] (` pattern (space before paren in markdown links) — confirm zero remaining
  - [x] 10.2: Grep all doc files for `#RandomSecret` (mixed case) — confirm zero remaining
  - [x] 10.3: Grep all doc files for `##[a-z]` (double-hash anchors) — confirm zero remaining
  - [x] 10.4: Grep `auth-engines.md` for `_` in field description lines of GCP and Azure sections — confirm zero remaining snake_case field names
  - [x] 10.5: Spot-check 3-5 other engine doc sections for snake_case field descriptions (identify any NOT caught above)

### Review Findings

- [x] [Review][Patch] Correct stale OIDC-specific prose under `azureCredentials` so the Azure auth config docs match the CRD’s `RootCredentialConfig` semantics. [docs/auth-engines.md:646] — fixed
- [x] [Review][Patch] Finish the camelCase sweep in `auth-engines.md`; several JWT/OIDC and GCP descriptions still reference snake_case names like `token_ttl`, `jwks_url`, `bound_claims_type`, `iam_alias`, and `gce_alias`, which conflicts with the story’s “all field references” requirement. [docs/auth-engines.md:203] — fixed

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are edits to existing markdown files in `docs/`.

### Files Modified

4 files total:

| File | Changes |
|------|---------|
| `docs/secret-engines.md` | Fix 2 broken TOC links (lines 19-20), fix 4 broken `#RandomSecret` anchors |
| `docs/auth-engines.md` | Fix 18 snake_case→camelCase in GCPAuthEngineRole, 6 in AzureAuthEngineConfig, 16 in AzureAuthEngineRole, 1 YAML PascalCase fix, fix 3 broken `#RandomSecret` anchors, fix 4 leading-space code fences |
| `docs/secret-management.md` | Fix 1 broken reference link (`[Password Policy]`), fix 1 cross-file anchor (`#the-authentication-section`) |
| `docs/end-to-end-example.md` | Fix 1 double-hash anchor (`##deploying-the-operator`) |

### Verified CRD Field Name Mappings

These mappings were verified against the actual Go struct json tags in `api/v1alpha1/`:

**GCPAuthEngineRole** (from `gcpauthenginerole_types.go`):

| snake_case (Vault API) | camelCase (CRD json tag) |
|------------------------|--------------------------|
| `bound_service_accounts` | `boundServiceAccounts` |
| `bound_projects` | `boundProjects` |
| `add_group_aliases` | `addGroupAliases` |
| `token_ttl` | `tokenTTL` |
| `token_max_ttl` | `tokenMaxTTL` |
| `token_policies` | `tokenPolicies` |
| `token_bound_cidrs` | `tokenBoundCIDRs` |
| `token_explicit_max_ttl` | `tokenExplicitMaxTTL` |
| `token_no_default_policy` | `tokenNoDefaultPolicy` |
| `token_num_uses` | `tokenNumUses` |
| `token_period` | `tokenPeriod` |
| `token_type` | `tokenType` |
| `max_jwt_exp` | `maxJWTExp` |
| `allow_gce_inference` | `allowGCEInference` |
| `bound_zones` | `boundZones` |
| `bound_regions` | `boundRegions` |
| `bound_instance_groups` | `boundInstanceGroups` |
| `bound_labels` | `boundLabels` |

**AzureAuthEngineConfig** (from `azureauthengineconfig_types.go`):

| snake_case (Vault API) | camelCase (CRD json tag) |
|------------------------|--------------------------|
| `tenant_id` | `tenantID` |
| `client_id` | `clientID` |
| `client_secret` | `azureCredentials` (type `RootCredentialConfig`) |
| `max_retries` | `maxRetries` |
| `max_retry_delay` | `maxRetryDelay` |
| `retry_delay` | `retryDelay` |

**AzureAuthEngineRole** (from `azureauthenginerole_types.go`):

| snake_case (Vault API) | camelCase (CRD json tag) |
|------------------------|--------------------------|
| `bound_service_principal_ids` | `boundServicePrincipalIDs` |
| `bound_group_ids` | `boundGroupIDs` |
| `bound_locations` | `boundLocations` |
| `bound_subscription_ids` | `boundSubscriptionIDs` |
| `bound_resource_groups` | `boundResourceGroups` |
| `bound_scale_sets` | `boundScaleSets` |
| `token_ttl` | `tokenTTL` |
| `token_max_ttl` | `tokenMaxTTL` |
| `token_policies` | `tokenPolicies` |
| `token_bound_cidrs` | `tokenBoundCIDRs` |
| `token_explicit_max_ttl` | `tokenExplicitMaxTTL` |
| `token_no_default_policy` | `tokenNoDefaultPolicy` |
| `token_num_uses` | `tokenNumUses` |
| `token_period` | `tokenPeriod` |
| `token_type` | `tokenType` |

### Important: Capitalization Nuances

Some CRD field names have unusual capitalization that matches Go naming conventions, NOT simple camelCase. Pay special attention to:
- `tokenTTL` (not `tokenTtl`) — TTL is an abbreviation kept uppercase
- `tokenMaxTTL` (not `tokenMaxTtl`)
- `tokenBoundCIDRs` (not `tokenBoundCidrs`) — CIDRs is an abbreviation
- `tokenExplicitMaxTTL` (not `tokenExplicitMaxTtl`)
- `maxJWTExp` (not `maxJwtExp`) — JWT is an abbreviation
- `allowGCEInference` (not `allowGceInference`) — GCE is an abbreviation
- `tenantID` (not `tenantId`) — ID is an abbreviation
- `clientID` (not `clientId`)
- `boundServicePrincipalIDs` (not `boundServicePrincipalIds`)
- `boundGroupIDs` (not `boundGroupIds`)
- `boundSubscriptionIDs` (not `boundSubscriptionIds`)

These MUST match the Go json tags exactly. Cross-reference with the source files listed above.

### `client_secret` → `azureCredentials` Clarification

The `client_secret` field description in `AzureAuthEngineConfig` does NOT map to a simple `clientSecret` CRD field. The CRD uses `azureCredentials` (json tag: `"azureCredentials,omitempty"`), which is of type `vaultutils.RootCredentialConfig` — a struct with `secret`, `vaultSecret`, and `randomSecret` sub-fields for the three credential resolution methods. When updating this field description, change the name to `azureCredentials` and note that it supports the standard three credential resolution methods (Kubernetes Secret, Vault Secret, RandomSecret).

### Scope Boundaries

- Do NOT restructure or reorganize any doc files — D2/D3 epics handle the per-engine split
- Do NOT add new sections, field description tables, or Vault CLI equivalents — that is D2/D3 standardization scope
- Do NOT modify YAML example content beyond the single PascalCase fix (Task 8.1)
- Do NOT modify any Go code, CRD types, or controllers
- Do NOT run `make manifests generate` or `make test`
- Fix ONLY broken links, field naming (snake_case→camelCase), and code fence formatting issues

### Discovery Approach for Task 10

The final audit (Task 10) exists to catch issues the initial analysis may have missed. Useful grep patterns:

```bash
# Find space-before-paren in markdown links
rg '\] \(' docs/

# Find mixed-case #RandomSecret anchors
rg '#RandomSecret' docs/

# Find double-hash anchors
rg '##[a-z]' docs/

# Find snake_case in field description lines (broader check)
rg '_[a-z]+' docs/auth-engines.md | grep -v '```' | grep -v 'yaml' | grep -v 'http'
```

If Task 10 discovers additional snake_case issues in other engine sections (e.g., Kubernetes, LDAP, JWT/OIDC sections of `auth-engines.md` or any section of `secret-engines.md`), fix them as well. The AC says "all field references are updated to camelCase" — this is a sweep, not limited to GCP/Azure.

### Previous Story Intelligence

**From D1.1 (Create Documentation Template — sibling story):**
- D1.1 identified the same field naming problems: "Mixed camelCase/snake_case field names in GCPAuthEngineRole section"
- D1.1 notes `secret-engines.md` "Has broken markdown links (lines 19-20: space before parentheses)"
- D1.1 establishes DNFR3: "Field descriptions must use camelCase (CRD field names), not snake_case (Vault API names)"
- D1.1 lists exactly which doc files have which issues — use as a cross-reference

**From D1.0a, D1.0b, D1.0c (sibling code/metadata stories):**
- These are runtime fix and metadata stories — no documentation changes. No learnings applicable to D1.3.

### Phase 1.5 Non-Functional Requirements to Follow

- **DNFR3:** Field descriptions must use camelCase (CRD field names), not snake_case (Vault API names) — this is the core requirement for Tasks 6-8
- **DNFR4:** All internal cross-references between doc files must work — this is the core requirement for Tasks 1-5

### How This Story Relates to D2/D3

D1.3 fixes quality issues in the CURRENT monolith doc files. When D2 splits `auth-engines.md` into per-engine files and D3 splits `secret-engines.md`, those stories will extract content from the already-fixed files. Fixing now means D2/D3 start from a clean baseline rather than inheriting broken links and naming issues.

### Project Structure Notes

- All doc files are in `docs/` (flat structure, 9 markdown files)
- No new files created
- No files deleted
- No directory structure changes
- Future D2/D3 will create `docs/auth-engines/` and `docs/secret-engines/` subdirectories

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D1.3] — Story requirements and acceptance criteria
- [Source: _bmad-output/planning-artifacts/epics.md#Phase 1.5 Requirements] — DNFR3 (camelCase), DNFR4 (cross-references)
- [Source: api/v1alpha1/gcpauthenginerole_types.go] — GCP auth engine role CRD field json tags (authoritative camelCase names)
- [Source: api/v1alpha1/azureauthengineconfig_types.go] — Azure auth engine config CRD field json tags
- [Source: api/v1alpha1/azureauthenginerole_types.go] — Azure auth engine role CRD field json tags
- [Source: docs/secret-engines.md:19-20] — Broken TOC links (space before paren)
- [Source: docs/secret-engines.md:97,286,416,681] — Broken #RandomSecret anchors
- [Source: docs/auth-engines.md:546-586] — GCPAuthEngineRole snake_case field names
- [Source: docs/auth-engines.md:626-642] — AzureAuthEngineConfig snake_case field names
- [Source: docs/auth-engines.md:717,743-773] — AzureAuthEngineRole PascalCase YAML + snake_case field names
- [Source: docs/auth-engines.md:159,314,671] — Broken #RandomSecret anchors
- [Source: docs/auth-engines.md:499,540,684,739] — Leading-space code fences
- [Source: docs/secret-management.md:10] — Broken [Password Policy] reference link
- [Source: docs/secret-management.md:123] — Broken cross-file anchor (#the-authentication-section)
- [Source: docs/end-to-end-example.md:12] — Double-hash anchor (##deploying-the-operator)
- [Source: _bmad-output/implementation-artifacts/d1-1-create-documentation-template-and-pattern-guide.md] — Sibling story identifying same issues
- [Source: _bmad-output/project-context.md#JSON Tag Conventions] — camelCase json tag convention

## Dev Agent Record

### Agent Model Used

Claude Opus 4 (claude-sonnet-4-20250514)

### Debug Log References

### Completion Notes List

- Fixed 2 broken TOC links in `docs/secret-engines.md` (space before parenthesis)
- Fixed 4 broken `#RandomSecret` cross-file references in `docs/secret-engines.md` (wrong file and/or wrong anchor case)
- Fixed 3 broken `#RandomSecret` references in `docs/auth-engines.md` (wrong file and/or wrong anchor case)
- Fixed 1 broken `#RandomSecret` reference in `docs/engine-doc-template.md` (wrong anchor case, discovered in audit)
- Fixed broken `[Password Policy]` reference-style link in `docs/secret-management.md` (converted to inline link)
- Fixed broken `[authentication](#the-authentication-section)` cross-file link in `docs/secret-management.md` (pointed to correct file `auth-section.md`)
- Fixed double-hash anchor `##deploying-the-operator` in `docs/end-to-end-example.md`
- Fixed 18 snake_case→camelCase field names in GCPAuthEngineRole section of `docs/auth-engines.md`
- Fixed 6 snake_case→camelCase field names in AzureAuthEngineConfig section (including `client_secret`→`azureCredentials`)
- Fixed PascalCase `BoundResourceGroups:` → `boundResourceGroups:` in AzureAuthEngineRole YAML example
- Fixed 16 snake_case→camelCase field names in AzureAuthEngineRole section of `docs/auth-engines.md`
- Fixed 4 leading-space code fences in `docs/auth-engines.md` (GCPAuthEngineRole and AzureAuthEngineRole blocks)
- Final audit confirmed zero remaining broken patterns across all doc files

### Change Log

- 2026-06-23: Implemented story D1.3 — fixed all broken links, field naming inconsistencies, and code fence formatting issues across documentation files

### File List

- docs/secret-engines.md (modified)
- docs/auth-engines.md (modified)
- docs/secret-management.md (modified)
- docs/end-to-end-example.md (modified)
- docs/engine-doc-template.md (modified)
