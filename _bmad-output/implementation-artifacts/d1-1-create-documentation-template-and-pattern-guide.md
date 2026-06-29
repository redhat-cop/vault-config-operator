# Story D1.1: Create Documentation Template and Pattern Guide

Status: done

## Story

As a documentation contributor,
I want a clear template that defines the structure every engine doc file must follow,
So that all engine docs are consistent and contributors know exactly what to write.

## Acceptance Criteria

1. **Given** the need for consistent engine documentation **When** a template file is created at `docs/engine-doc-template.md` **Then** it contains the following sections in order:
   1. Title with link to Vault documentation
   2. Overview paragraph describing the engine
   3. Config CRD section: description, full YAML example with all common fields, field descriptions table (camelCase names), Vault CLI equivalent command
   4. Role/Group CRD section: same structure as config
   5. Credential resolution section (if applicable): examples for Kubernetes Secret, Vault Secret, and RandomSecret methods
   6. Links to related docs (`auth-section.md`, `contributing-vault-apis.md`)

2. **Given** the template is created **When** reviewed against `identities.md` and `audit-management.md` (the current gold standard docs) **Then** the template matches or improves upon their quality

## Tasks / Subtasks

- [x] Task 1: Analyze existing gold-standard docs for structural patterns (AC: 2)
  - [x] 1.1: Review `docs/identities.md` structure — per-CRD pattern: H2 heading, description+link, YAML example, field descriptions as list
  - [x] 1.2: Review `docs/audit-management.md` structure — per-CRD pattern: H2 heading, description+link, YAML example, `### Field Description` header, field list
  - [x] 1.3: Review existing auth-engine and secret-engine docs for Vault CLI equivalents and credential resolution sections already in use
  - [x] 1.4: Identify inconsistencies to fix in the template (mixed heading styles, missing field tables, no Vault CLI equivalents in some docs)
- [x] Task 2: Create `docs/engine-doc-template.md` (AC: 1)
  - [x] 2.1: Write the template with placeholder sections in the exact order specified in AC1
  - [x] 2.2: Use `{{placeholder}}` syntax for variable content (engine name, CRD kind, Vault doc URL, etc.)
  - [x] 2.3: Include inline comments/instructions for contributors explaining what each section must contain
  - [x] 2.4: Provide a filled-in example section alongside each template section so contributors see the expected output
  - [x] 2.5: Field descriptions must use a markdown table format (`Field | Type | Required | Description`) with camelCase field names (DNFR3)
  - [x] 2.6: YAML examples must use `apiVersion: redhatcop.redhat.io/v1alpha1` (DNFR2)
  - [x] 2.7: Include the Vault CLI equivalent command after each CRD's YAML example
  - [x] 2.8: Include credential resolution section template with three subsections: Kubernetes Secret, Vault Secret, RandomSecret — each with a YAML snippet
  - [x] 2.9: Include links section template pointing to `auth-section.md` and `contributing-vault-apis.md`
- [x] Task 3: Validate template quality against gold standards (AC: 2)
  - [x] 3.1: Verify every structural element in `identities.md` and `audit-management.md` has a corresponding template section
  - [x] 3.2: Verify the template adds improvements not present in either: field description tables (vs plain lists), Vault CLI equivalents, credential resolution section
  - [x] 3.3: Verify DNFR1-DNFR3 compliance: all camelCase field names, valid YAML apiVersion, consistent section ordering

### Review Findings

- [x] [Review][Patch] Template hardcodes `Role` and does not support `Group` variants [`docs/engine-doc-template.md:24`]
- [x] [Review][Patch] Secret-engine examples still use auth-style paths and commands [`docs/engine-doc-template.md:84`]
- [x] [Review][Patch] Role/Group YAML example omits the optional `connection` block pattern [`docs/engine-doc-template.md:141`]
- [x] [Review][Patch] Credential-resolution placeholders do not support nested credential objects like `OIDCCredentials` [`docs/engine-doc-template.md:221`]

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverable is a single markdown file: `docs/engine-doc-template.md`.

### Existing Doc Patterns to Standardize

The current docs use inconsistent patterns that the template must unify:

**`identities.md`** (11 CRDs documented):
- Title: `# Identities` with link to Vault concepts page
- TOC: bullet list of anchor links
- Per-CRD: `## CRDName`, description paragraph with Vault API doc link, YAML code block, field descriptions as a plain bullet list
- Inconsistent field section headings: `### Entity Fields` vs bare `The following fields are available:`
- No Vault CLI equivalents
- No credential resolution section (not applicable for identity types)

**`audit-management.md`** (2 CRDs):
- Title: `# Audit Management` with link to Vault audit docs
- TOC: bullet list of anchor links
- Per-CRD: `## CRDName`, description with Vault doc link, YAML code block, `### Field Description` heading, field list
- Consistent structure but uses plain bullet lists for fields, not tables

**`auth-engines.md`** (10 CRDs, 775 lines):
- Monolith file — D2 will split into per-engine files
- Has Vault CLI equivalents for some engines (e.g., `SecretEngineMount`)
- Has credential resolution examples for LDAP, JWT/OIDC (three methods: K8s Secret, Vault Secret, RandomSecret)
- Inconsistent: some engines have detailed field descriptions, others have almost none
- Mixed camelCase/snake_case field names in GCPAuthEngineRole section

**`secret-engines.md`** (16 CRDs, 741 lines):
- Monolith file — D3 will split into per-engine files
- Has broken markdown links (lines 19-20: space before parentheses)
- Has Vault CLI equivalents for some engines
- Has credential resolution examples for some engines

### Template Structure (Exact Sections)

The template must define these sections in this order:

```
# {EngineName} {Auth|Secret} Engine
  → Title with link to Vault documentation

## Overview
  → 2-3 sentence description of what the engine does

## {EngineName}{Auth|Secret}EngineConfig
  → Config CRD section:
    ### Description (paragraph with Vault API doc link)
    ### Example (full YAML with apiVersion: redhatcop.redhat.io/v1alpha1)
    ### Vault CLI Equivalent (shell code block)
    ### Field Descriptions (markdown table: Field | Type | Required | Description)

## {EngineName}{Auth|Secret}EngineRole (or Group)
  → Role/Group CRD section (same sub-structure as Config)

## Credential Resolution (if applicable)
  → Three subsections:
    ### Using a Kubernetes Secret
    ### Using a Vault Secret
    ### Using a RandomSecret

## See Also
  → Links to auth-section.md, contributing-vault-apis.md, related docs
```

### Improvements Over Gold Standards

The template adds these improvements not consistently present in existing docs:
1. **Field description tables** instead of plain bullet lists — structured format is scannable and consistent
2. **Vault CLI equivalent** for every CRD — helps users understand what the CR does in Vault terms
3. **Credential resolution** section with all three methods — currently scattered and inconsistent
4. **Consistent heading hierarchy** — H1 for page title, H2 for each CRD, H3 for subsections
5. **camelCase field names** enforced (DNFR3) — some existing docs mix snake_case

### Phase 1.5 Non-Functional Requirements to Follow

- **DNFR1:** Every engine doc file must follow this template
- **DNFR2:** All YAML examples must use `apiVersion: redhatcop.redhat.io/v1alpha1`
- **DNFR3:** Field descriptions must use camelCase (CRD field names), not snake_case (Vault API names)
- **DNFR4:** All internal cross-references between doc files must work after the doc split (D2/D3 scope, but template should use relative links that work in the target directory structure)
- **DNFR5:** Phase 2 engine implementation epics must include documentation as AC

### How Subsequent Stories Use This Template

- **D1.2**: Uses the template to document CertAuthEngineConfig and CertAuthEngineRole (first real usage)
- **D1.3**: Fixes broken links and field naming inconsistencies in existing docs
- **D2.1-D2.5**: Splits `auth-engines.md` into per-engine files, standardizes each to this template
- **D3.1-D3.4**: Splits `secret-engines.md` into per-engine files, standardizes each to this template
- **Phase 2 (Epics 11-16)**: Every new engine type must include documentation following this template

### Authentication Section Pattern

All engine docs reference the common `authentication` spec field. The template should NOT duplicate the full authentication explanation — instead, link to `docs/auth-section.md` which documents:
- `authentication.path`: Kubernetes auth engine mount path
- `authentication.role`: Vault role to authenticate as
- `authentication.namespace`: Vault namespace (optional)
- `authentication.serviceAccount.name`: Service account for JWT token

YAML examples in the template should include the `authentication` block with typical values (`path: kubernetes`, `role: policy-admin`) but field descriptions should just say "See [Authentication](auth-section.md)".

### Connection Section Pattern

Most CRDs also have an optional `connection` spec field for connecting to a non-default Vault instance. The template should include `connection` in YAML examples with a comment like `# Optional: override Vault connection settings` and link to the relevant docs.

### Files Created

Only 1 file: `docs/engine-doc-template.md`

### What NOT to Do

- Do NOT modify any existing doc files — D1.3 handles fixes, D2/D3 handle standardization
- Do NOT create per-engine files — D2/D3 scope
- Do NOT create the `docs/auth-engines/` or `docs/secret-engines/` directories — D2.1/D3.1 scope
- Do NOT modify any Go code, CRD types, or controllers
- Do NOT run `make manifests generate` or `make test`

### Project Structure Notes

- Template goes in `docs/` alongside existing doc files
- The `docs/` directory currently has 9 markdown files (flat structure)
- Future stories (D2, D3) will create `docs/auth-engines/` and `docs/secret-engines/` subdirectories
- The template file name `engine-doc-template.md` is specified in the epic AC

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D1.1] — Story requirements and acceptance criteria
- [Source: _bmad-output/planning-artifacts/epics.md#Phase 1.5 Requirements] — DOC1-DOC10, DNFR1-DNFR5
- [Source: docs/identities.md] — Gold standard doc (11 CRDs, best current quality)
- [Source: docs/audit-management.md] — Gold standard doc (2 CRDs, consistent structure)
- [Source: docs/auth-engines.md] — Current monolith (775 lines, has credential resolution and CLI examples to extract patterns from)
- [Source: docs/secret-engines.md] — Current monolith (has broken links on lines 19-20, CLI equivalents)
- [Source: docs/auth-section.md] — Common authentication section documentation (link target for template)
- [Source: docs/contributing-vault-apis.md] — Developer guide for adding new types (link target for template)

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (via Cursor)

### Debug Log References

No debug issues — documentation-only story, no code compilation or test execution required.

### Completion Notes List

- Analyzed all 4 existing doc files (identities.md, audit-management.md, auth-engines.md, secret-engines.md) for structural patterns
- Identified 5 key inconsistencies: mixed heading styles, missing field tables, no Vault CLI equivalents in gold standards, scattered credential resolution, mixed camelCase/snake_case
- Created `docs/engine-doc-template.md` with all required sections in exact order per AC1
- Template uses `{{placeholder}}` syntax, includes inline HTML comment instructions, and provides filled-in examples alongside each section
- Field description tables use markdown format (Field | Type | Required | Description) with camelCase names (DNFR3)
- All YAML examples use `apiVersion: redhatcop.redhat.io/v1alpha1` (DNFR2)
- Vault CLI equivalent included after each CRD YAML example
- Credential resolution section provides three methods (Kubernetes Secret, Vault Secret, RandomSecret) with examples
- Links section references auth-section.md and contributing-vault-apis.md
- Validated template covers all structural elements from gold standards plus adds three improvements (tables, CLI equivalents, credential resolution)
- Confirmed DNFR1-DNFR3 compliance

### File List

- docs/engine-doc-template.md (new)
