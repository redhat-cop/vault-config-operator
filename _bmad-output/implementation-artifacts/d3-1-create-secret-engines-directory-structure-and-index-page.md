# Story D3.1: Create Secret-Engines Directory Structure and Index Page

Status: done

## Story

As a documentation reader,
I want an organized directory with an index page listing all secret engines,
So that I can quickly navigate to the engine I need.

## Acceptance Criteria

1. **Given** the decision to split `secret-engines.md` into per-engine files **When** the directory `docs/secret-engines/` is set up with an `index.md` **Then** the index page contains:
   - Overview of secret engine support in the operator
   - SecretEngineMount section (the generic mount type — kept in the index since it is not engine-specific)
   - Table of all supported secret engines with links to individual per-engine files
   - Reference to `auth-section.md` for authentication configuration

2. **Given** the original `secret-engines.md` **When** the split is complete **Then** `secret-engines.md` is replaced with a short redirect/pointer to `docs/secret-engines/index.md`

## Tasks / Subtasks

- [x] Task 1: Create `docs/secret-engines/index.md` (AC: 1)
  - [x] 1.1: Create the `docs/secret-engines/` directory and write the page title and overview paragraph describing secret engine support in the operator (2-3 sentences, explain that the operator supports dynamic secrets via Config CRDs and Role CRDs for each engine)
  - [x] 1.2: Include the SecretEngineMount section — copy verbatim from `docs/secret-engines.md` lines 23-50 (the `## SecretEngineMount` heading, YAML example, `type` and `path` field descriptions, Vault CLI equivalent). This is the generic mount and belongs in the index, not in a per-engine file
  - [x] 1.3: Create a markdown table listing all supported secret engines with columns: Engine | Config CRD | Role CRD(s) | File. Link each file to its future per-engine file path (relative links). Use these entries:

    | Engine | Config CRD | Role CRD(s) | File |
    |--------|-----------|-------------|------|
    | Database | DatabaseSecretEngineConfig | DatabaseSecretEngineRole, DatabaseSecretEngineStaticRole | [database.md](database.md) |
    | PKI | PKISecretEngineConfig | PKISecretEngineRole | [pki.md](pki.md) |
    | RabbitMQ | RabbitMQSecretEngineConfig | RabbitMQSecretEngineRole | [rabbitmq.md](rabbitmq.md) |
    | GitHub | GitHubSecretEngineConfig | GitHubSecretEngineRole | [github.md](github.md) |
    | Quay | QuaySecretEngineConfig | QuaySecretEngineRole, QuaySecretEngineStaticRole | [quay.md](quay.md) |
    | Kubernetes | KubernetesSecretEngineConfig | KubernetesSecretEngineRole | [kubernetes.md](kubernetes.md) |
    | Azure | AzureSecretEngineConfig | AzureSecretEngineRole | [azure.md](azure.md) |

  - [x] 1.4: Add a "Common Configuration" section with a link to `../auth-section.md` for the shared `authentication` block and a link to `../contributing-vault-apis.md` for the shared `connection` block
  - [x] 1.5: Add a "See Also" section with links to the main README, contributing guide, and auth engines index (`../auth-engines/index.md`)

- [x] Task 2: Replace `docs/secret-engines.md` with a redirect pointer (AC: 2)
  - [x] 2.1: Replace the contents of `docs/secret-engines.md` with a short document that:
    - States the secret engine documentation has moved to `docs/secret-engines/`
    - Links to `secret-engines/index.md` as the new location
    - Preserves the original page title for any external bookmarks
  - [x] 2.2: Keep the file small (under 15 lines) — it is just a pointer

- [x] Task 3: Update `docs/auth-engines/index.md` See Also section (AC: 1)
  - [x] 3.1: Replace the HTML comment placeholder `<!-- Secret engines index will be created in D3: ../secret-engines/index.md -->` with an actual link to `../secret-engines/index.md` in the See Also section

- [x] Task 4: Verify cross-references (AC: 1, 2)
  - [x] 4.1: Check if any other doc files link to `secret-engines.md` sections — document all found cross-references in the Dev Agent Record for downstream stories (D3.2-D3.4 will update them as content moves)
  - [x] 4.2: The known cross-references in `readme.md` lines 84-95 link to `docs/secret-engines.md` with anchors — document these for downstream stories

### Review Findings

- [x] [Review][Patch] Preserve the original redirect page title in `docs/secret-engines.md` [`docs/secret-engines.md:1`] The story explicitly requires keeping the original `# Secret Engines` title for external bookmarks, but the redirect page was changed to `# Secret Engines APIs`. — Fixed: restored to `# Secret Engines`.

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 1 new directory: `docs/secret-engines/`
- 1 new file: `docs/secret-engines/index.md`
- 1 modified file: `docs/secret-engines.md` (replaced with redirect pointer)
- 1 modified file: `docs/auth-engines/index.md` (replace D3 placeholder comment with actual link)

### Current State of `docs/secret-engines/`

The directory does NOT exist yet (unlike `docs/auth-engines/` which was pre-created by D1.2). You must create it as part of this story.

### SecretEngineMount Belongs in the Index

The `SecretEngineMount` CRD is a generic mount type — it enables any secret engine at a path. It is not engine-specific, so it should live in the index page alongside the engine table, NOT in its own per-engine file. Copy the existing content from `docs/secret-engines.md` lines 23-50 (heading, YAML example, field descriptions, and Vault CLI equivalent).

### Do NOT Extract Per-Engine Content Yet

Stories D3.2-D3.4 handle extracting and standardizing the per-engine documentation. This story ONLY creates the index and redirect. The per-engine files linked in the table (database.md, pki.md, etc.) will be dead links until those stories are completed — this is expected.

### Template Reference

The documentation template is at `docs/engine-doc-template.md` (created in D1.1, review-patched 4 times). Per-engine files created in D3.2-D3.4 must follow this template. The index page does NOT follow the per-engine template — it is a navigation/overview page.

### Cross-Reference Impact

Replacing `secret-engines.md` with a redirect will break any direct anchor links (e.g., `secret-engines.md#databasesecretengineconfig`). This is acceptable because:
1. The redirect file will point readers to the new location
2. D3.2-D3.4 will create the actual per-engine files where the content lives
3. Any internal doc cross-references will be updated as each engine's content is moved

### Known Cross-References to `docs/secret-engines.md`

From `readme.md` lines 84-95:

| Line | Link Target | Downstream Story |
|------|-------------|-----------------|
| 84 | `secret-engines.md#SecretEngineMount` | Stays in index — consider updating link to `secret-engines/index.md#secretenginemount` in any D3.x story |
| 85 | `secret-engines.md#DatabaseSecretEngineConfig` | D3.2 |
| 86 | `secret-engines.md#DatabaseSecretEngineRole` | D3.2 |
| 87 | `secret-engines.md#GitHubSecretEngineConfig` | D3.4 |
| 88 | `secret-engines.md#GitHubSecretEngineRole` | D3.4 |
| 89 | `secret-engines.md#pkisecretengineconfig` | D3.3 |
| 90 | `secret-engines.md#pkisecretenginerole` | D3.3 |
| 91 | `secret-engines.md#QuaySecretEngineConfig` | D3.4 |
| 92 | `secret-engines.md#QuaySecretEngineRole` | D3.4 |
| 93 | `secret-engines.md#QuaySecretEngineStaticRole` | D3.4 |
| 94 | `secret-engines.md#rabbitmqsecretengineconfig` | D3.3 |
| 95 | `secret-engines.md#rabbitmqsecretenginerole` | D3.3 |

Note: `readme.md` does NOT reference `KubernetesSecretEngine*` or `AzureSecretEngine*` — these were added later and never got readme entries. D3.4 may optionally add them.

### Auth-Engines Index Placeholder

`docs/auth-engines/index.md` line 46 contains:
```
<!-- Secret engines index will be created in D3: ../secret-engines/index.md -->
```

Replace this with an actual link in the See Also section. Follow the same style as the existing entries.

### Redirect Pointer Pattern (from D2.1)

Follow the exact same pattern used for `docs/auth-engines.md`:
```markdown
# Secret Engines APIs

This documentation has moved to [docs/secret-engines/](secret-engines/index.md).

The secret engine documentation has been split into per-engine files for easier navigation. See the [secret engines index](secret-engines/index.md) for the full list of supported engines and links to each engine's documentation.
```

### Relative Link Conventions

- From `docs/secret-engines/index.md`:
  - To per-engine files: `database.md`, `pki.md`, etc. (same directory)
  - To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`
  - To auth engines index: `../auth-engines/index.md`
  - To root README: `../../readme.md`
- From `docs/secret-engines.md` (redirect):
  - To new index: `secret-engines/index.md`

### Previous Story Intelligence

**From D2.1 (Auth-Engines Directory Structure & Index — the exact precedent):**
- Created `docs/auth-engines/index.md` with overview, engine table, common config section, see also
- Replaced `docs/auth-engines.md` with 5-line redirect pointer
- Documented cross-references in readme.md for downstream stories
- Verified relative links from subdirectory resolve correctly
- AuthEngineMount section placed in index (not per-engine file) — follow same pattern for SecretEngineMount
- The secret engines index comment placeholder was added in See Also section (now needs activation)
- Agent model: Opus 4.6

**From D2.5 (Last story in Epic D2):**
- Zero review findings — the team fully internalized the template pattern by end of D2
- Credential resolution documentation patterns established (Pattern A flat prefix, Pattern B nested object)
- Both GCP and Azure patterns documented correctly — applicable to secret engine credential docs in D3.2-D3.4

**From D2 Retrospective:**
- D3 readiness: No preparation work needed, all preconditions satisfied
- Template proven across 6 auth engine docs — applies directly to secret engines
- Secret engines have more CRD variety than auth engines (some have Config + Role + StaticRole) — note the Role CRD(s) column in the table accommodates this
- Potential friction: more credential resolution patterns across secret engines — manageable within D3.2-D3.4 stories
- Recommendation: Continue using Opus 4.6 for all stories

**From D1.3 (Link and Naming Fixes):**
- Fixed 2 broken TOC links in `secret-engines.md` (space before parenthesis on lines 19-20)
- Fixed 4 broken `#RandomSecret` cross-file references in `secret-engines.md`
- Current `secret-engines.md` content is clean (quality baseline established)

### Project Structure Notes

```
docs/
├── secret-engines/
│   ├── index.md          ← NEW (this story)
│   ├── database.md       ← D3.2 (future)
│   ├── pki.md            ← D3.3 (future)
│   ├── rabbitmq.md       ← D3.3 (future)
│   ├── github.md         ← D3.4 (future)
│   ├── quay.md           ← D3.4 (future)
│   ├── kubernetes.md     ← D3.4 (future)
│   └── azure.md          ← D3.4 (future)
├── secret-engines.md     ← MODIFY → redirect pointer (this story)
├── auth-engines/
│   └── index.md          ← MODIFY → replace D3 comment with link (this story)
├── auth-section.md       ← unchanged
├── engine-doc-template.md ← unchanged (reference for D3.2-D3.4)
└── ...
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D3.1] — Story requirements and acceptance criteria
- [Source: docs/secret-engines.md:23-50] — SecretEngineMount section to copy into index
- [Source: docs/secret-engines.md:1-21] — Current TOC showing all secret engine types
- [Source: docs/auth-engines/index.md] — Reference implementation for index page structure (D2.1)
- [Source: docs/auth-engines.md] — Reference implementation for redirect pointer pattern (D2.1)
- [Source: docs/engine-doc-template.md] — Template for per-engine files (D1.1, review-patched)
- [Source: _bmad-output/implementation-artifacts/d2-1-create-auth-engines-directory-structure-and-index-page.md] — Exact precedent story — follow same pattern
- [Source: _bmad-output/implementation-artifacts/epic-d2-retro-2026-07-02.md#Epic D3 Preparation] — Readiness assessment
- [Source: readme.md:84-95] — Cross-references to secret-engines.md that will need updating in D3.2-D3.4
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

Opus 4.6

### Debug Log References

### Completion Notes List

- Created `docs/secret-engines/` directory and `docs/secret-engines/index.md` with overview paragraph, SecretEngineMount section (copied verbatim from original `secret-engines.md` lines 23-50), supported secret engines table with 7 engine entries and relative links to future per-engine files, Common Configuration section linking to `auth-section.md` and `contributing-vault-apis.md`, and See Also section linking to README, contributing guide, and auth engines index
- Replaced `docs/secret-engines.md` (741 lines) with 5-line redirect pointer following the exact pattern from `docs/auth-engines.md` (D2.1 precedent)
- Updated `docs/auth-engines/index.md` See Also section: replaced HTML comment placeholder with actual link to `../secret-engines/index.md`
- Cross-reference verification: confirmed only `readme.md` lines 84-95 reference `secret-engines.md` with anchors (12 links total). No other doc files within `docs/` contain references. All 12 links are documented in Dev Notes and assigned to downstream stories (D3.2, D3.3, D3.4) for updating as per-engine content is extracted

### Cross-References Found (for downstream stories)

From `readme.md`:

| Line | Current Link | Downstream Story |
|------|-------------|-----------------|
| 84 | `./docs/secret-engines.md#SecretEngineMount` | Stays in index — update to `./docs/secret-engines/index.md#secretenginemount` in any D3.x story |
| 85 | `./docs/secret-engines.md#DatabaseSecretEngineConfig` | D3.2 |
| 86 | `./docs/secret-engines.md#DatabaseSecretEngineRole` | D3.2 |
| 87 | `./docs/secret-engines.md#GitHubSecretEngineConfig` | D3.4 |
| 88 | `./docs/secret-engines.md#GitHubSecretEngineRole` | D3.4 |
| 89 | `./docs/secret-engines.md#pkisecretengineconfig` | D3.3 |
| 90 | `./docs/secret-engines.md#pkisecretenginerole` | D3.3 |
| 91 | `./docs/secret-engines.md#QuaySecretEngineConfig` | D3.4 |
| 92 | `./docs/secret-engines.md#QuaySecretEngineRole` | D3.4 |
| 93 | `./docs/secret-engines.md#QuaySecretEngineStaticRole` | D3.4 |
| 94 | `./docs/secret-engines.md#rabbitmqsecretengineconfig` | D3.3 |
| 95 | `./docs/secret-engines.md#rabbitmqsecretenginerole` | D3.3 |

No references found in other doc files (`docs/secret-management.md`, `docs/end-to-end-example.md`, `docs/contributing-vault-apis.md`, `docs/auth-section.md`, etc.).

### File List

- docs/secret-engines/index.md (new)
- docs/secret-engines.md (modified — replaced with redirect pointer)
- docs/auth-engines/index.md (modified — replaced D3 comment placeholder with actual link)

### Change Log

- 2026-07-03: Created secret-engines directory structure, index page, redirect pointer, and activated auth-engines cross-link (D3.1 implementation complete)
