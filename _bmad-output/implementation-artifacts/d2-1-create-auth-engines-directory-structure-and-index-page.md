# Story D2.1: Create Auth-Engines Directory Structure and Index Page

Status: ready-for-dev

## Story

As a documentation reader,
I want an organized directory with an index page listing all auth engines,
So that I can quickly navigate to the engine I need.

## Acceptance Criteria

1. **Given** the decision to split `auth-engines.md` into per-engine files **When** the directory `docs/auth-engines/` is set up with an `index.md` **Then** the index page contains:
   - Overview of auth engine support in the operator
   - AuthEngineMount section (the generic mount type — kept in the index since it is not engine-specific)
   - Table of all supported auth engines with links to individual per-engine files
   - Reference to `auth-section.md` for authentication configuration

2. **Given** the original `auth-engines.md` **When** the split is complete **Then** `auth-engines.md` is replaced with a short redirect/pointer to `docs/auth-engines/index.md`

## Tasks / Subtasks

- [ ] Task 1: Create `docs/auth-engines/index.md` (AC: 1)
  - [ ] 1.1: Write the page title and overview paragraph describing auth engine support in the operator (2-3 sentences, explain that each engine has a Config CRD and a Role/Group CRD)
  - [ ] 1.2: Include the AuthEngineMount section — copy verbatim from `auth-engines.md` lines 18-37 (the `## AuthEngineMount` heading, YAML example, `type` and `path` field descriptions). This is the generic mount and belongs in the index, not in a per-engine file
  - [ ] 1.3: Create a markdown table listing all supported auth engines with columns: Engine | Config CRD | Role/Group CRD | File. Link each file to its future per-engine file path (relative links). Use these entries:

    | Engine | Config CRD | Role/Group CRD | File |
    |--------|-----------|----------------|------|
    | Kubernetes | KubernetesAuthEngineConfig | KubernetesAuthEngineRole | [kubernetes.md](kubernetes.md) |
    | LDAP | LDAPAuthEngineConfig | LDAPAuthEngineGroup | [ldap.md](ldap.md) |
    | JWT/OIDC | JWTOIDCAuthEngineConfig | JWTOIDCAuthEngineRole | [jwt-oidc.md](jwt-oidc.md) |
    | GCP | GCPAuthEngineConfig | GCPAuthEngineRole | [gcp.md](gcp.md) |
    | Azure | AzureAuthEngineConfig | AzureAuthEngineRole | [azure.md](azure.md) |
    | TLS Certificate | CertAuthEngineConfig | CertAuthEngineRole | [cert.md](cert.md) |

  - [ ] 1.4: Add a "Common Configuration" section with a link to `../auth-section.md` for the shared `authentication` block and a link to `../contributing-vault-apis.md` for the shared `connection` block
  - [ ] 1.5: Add a "See Also" section with links to the main README, contributing guide, and secret engines index (future `../secret-engines/index.md` — use a comment noting this will be created in D3)

- [ ] Task 2: Replace `docs/auth-engines.md` with a redirect pointer (AC: 2)
  - [ ] 2.1: Replace the contents of `docs/auth-engines.md` with a short document that:
    - States the auth engine documentation has moved to `docs/auth-engines/`
    - Links to `auth-engines/index.md` as the new location
    - Preserves the original page title for any external bookmarks
  - [ ] 2.2: Keep the file small (under 15 lines) — it is just a pointer

- [ ] Task 3: Update `docs/auth-engines/cert.md` relative links (AC: 1)
  - [ ] 3.1: Verify that `cert.md`'s relative links (`../auth-section.md`, `../contributing-vault-apis.md`) still resolve correctly from `docs/auth-engines/` — they should, since cert.md was created in D1.2 under this directory already
  - [ ] 3.2: If any links are broken, fix them

- [ ] Task 4: Verify cross-references (AC: 1, 2)
  - [ ] 4.1: Check if any other doc files (`secret-management.md`, `end-to-end-example.md`, `secret-engines.md`, `contributing-vault-apis.md`) link to `auth-engines.md` sections (e.g., `auth-engines.md#kubernetesauthengineconfig`) — if so, they will need updating in D2.2-D2.5 when the content actually moves
  - [ ] 4.2: Document any found cross-references in the Dev Agent Record for downstream stories

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 1 new file: `docs/auth-engines/index.md`
- 1 modified file: `docs/auth-engines.md` (replaced with redirect pointer)
- 0-1 modified file: `docs/auth-engines/cert.md` (only if link fixes needed)

### Current State of `docs/auth-engines/`

The directory already exists (created in D1.2). It currently contains only `cert.md` (the TLS Certificate auth engine doc, created and validated in D1.2). No `index.md` exists yet.

### AuthEngineMount Belongs in the Index

The `AuthEngineMount` CRD is a generic mount type — it enables any auth engine at a path. It is not engine-specific, so it should live in the index page alongside the engine table, NOT in its own per-engine file. Copy the existing content from `auth-engines.md` lines 18-37.

### Do NOT Extract Per-Engine Content Yet

Stories D2.2-D2.5 handle extracting and standardizing the per-engine documentation. This story ONLY creates the index and redirect. The per-engine files linked in the table (kubernetes.md, ldap.md, etc.) will be dead links until those stories are completed — this is expected.

### Template Reference

The documentation template is at `docs/engine-doc-template.md` (created in D1.1, review-patched 4 times). Per-engine files created in D2.2-D2.5 must follow this template. The index page does NOT follow the per-engine template — it is a navigation/overview page.

### Cross-Reference Impact

Replacing `auth-engines.md` with a redirect will break any direct anchor links (e.g., `auth-engines.md#kubernetesauthengineconfig`). This is acceptable because:
1. The redirect file will point readers to the new location
2. D2.2-D2.5 will create the actual per-engine files where the content lives
3. Any internal doc cross-references will be updated as each engine's content is moved

### Known Residual Issues in Source Content

From the D1 retrospective:
- Kubernetes and LDAP sections in `auth-engines.md` were NOT explicitly audited for snake_case field names in D1.3 — D2.2 and D2.3 will handle during extraction
- The D1.1 template was patched 4 times during review — always use the current version of `docs/engine-doc-template.md`, not the pre-review version

### Relative Link Conventions

- From `docs/auth-engines/index.md`:
  - To per-engine files: `kubernetes.md`, `ldap.md`, etc. (same directory)
  - To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`
  - To root README: `../../readme.md`
- From `docs/auth-engines.md` (redirect):
  - To new index: `auth-engines/index.md`

### Previous Story Intelligence

**From D1.1 (Template Creation):**
- Established the documentation template all per-engine files must follow
- Defined DNFR1-DNFR5 requirements
- Template improvements: field description tables, Vault CLI equivalents, credential resolution sections

**From D1.2 (CertAuth Documentation):**
- First per-engine file created at `docs/auth-engines/cert.md`
- Validated the template pattern works in practice
- Established relative link patterns from `docs/auth-engines/` subdirectory

**From D1.3 (Link and Naming Fixes):**
- Fixed snake_case→camelCase in GCP and Azure sections
- Fixed broken RandomSecret cross-references
- Fixed leading-space code fences
- Scope explicitly excluded restructuring (deferred to D2)

**From D1 Retrospective:**
- All D1 retro action items resolved (6/6)
- D2 readiness assessed as complete — no preparation needed
- Template ready, directory exists, source quality cleaned
- Documentation stories have higher review finding rate — expect 3+ findings
- Opus 4.6 recommended for all stories

### Project Structure Notes

```
docs/
├── auth-engines/
│   ├── index.md          ← NEW (this story)
│   ├── cert.md           ← EXISTS (D1.2)
│   ├── kubernetes.md     ← D2.2 (future)
│   ├── ldap.md           ← D2.3 (future)
│   ├── jwt-oidc.md       ← D2.4 (future)
│   ├── gcp.md            ← D2.5 (future)
│   └── azure.md          ← D2.5 (future)
├── auth-engines.md       ← MODIFY → redirect pointer (this story)
├── auth-section.md       ← unchanged
├── engine-doc-template.md ← unchanged (reference for D2.2-D2.5)
└── ...
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D2.1] — Story requirements and acceptance criteria
- [Source: _bmad-output/planning-artifacts/epics.md#Phase 1.5 Requirements] — DOC4, DOC6, DNFR1-DNFR4
- [Source: _bmad-output/implementation-artifacts/epic-d1-retro-2026-06-28.md#Epic D2 Preparation] — Readiness assessment and known friction points
- [Source: docs/auth-engines.md:18-37] — AuthEngineMount section to copy into index
- [Source: docs/auth-engines.md:1-17] — Current TOC showing all auth engine types
- [Source: docs/auth-engines/cert.md] — Existing per-engine file (D1.2) — validates relative link pattern
- [Source: docs/engine-doc-template.md] — Template for per-engine files (D1.1, review-patched)
- [Source: _bmad-output/implementation-artifacts/d1-1-create-documentation-template-and-pattern-guide.md] — Template creation story with DNFR definitions
- [Source: _bmad-output/implementation-artifacts/d1-3-fix-broken-links-and-field-naming-inconsistencies.md] — Quality fixes applied to auth-engines.md before split
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
